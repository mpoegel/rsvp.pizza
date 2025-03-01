package pizza

import (
	"log/slog"
	"net/http"
	"path"
	"strings"
	"text/template"
	"time"

	uuid "github.com/google/uuid"
)

func (s *Server) authenticateRequest(r *http.Request) (*TokenClaims, bool) {
	var claims *TokenClaims
	for _, cookie := range r.Cookies() {
		if cookie.Name == "session" {
			var ok bool
			claims, ok = s.sessions[cookie.Value]
			// either bad session or auth has expired
			if !ok || time.Now().After(time.Unix(claims.Exp, 0)) {
				delete(s.sessions, cookie.Value)
				return nil, false
			}
		}
	}
	if claims == nil {
		return nil, false
	}

	return claims, true
}

func (s *Server) CheckAuthorization(r *http.Request) (*AccessToken, bool) {
	// check the authorization header for the access token
	rawAccessToken := r.Header.Get("Authorization")
	if rawAccessToken == "" {
		return nil, false
	}

	authParts := strings.Split(rawAccessToken, " ")
	if len(authParts) != 2 {
		return nil, false
	}

	// decode the access token
	accessToken, err := s.authenticator.DecodeAccessToken(r.Context(), authParts[1])
	if err != nil {
		return nil, false
	}

	// check token expiration
	if accessToken.ExpiresAt.Before(time.Now()) {
		return nil, false
	}

	return accessToken, true
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state := uuid.New()
	rawAccessToken := r.Header.Get("Authorization")
	if rawAccessToken == "" {
		s.sessions[state.String()] = nil
		http.Redirect(w, r, s.authenticator.GetAuthCodeURL(r.Context(), state.String()), http.StatusFound)
		return
	}

	authParts := strings.Split(rawAccessToken, " ")
	if len(authParts) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := s.authenticator.VerifyToken(r.Context(), authParts[1])
	if err != nil {
		s.sessions[state.String()] = nil
		http.Redirect(w, r, s.authenticator.GetAuthCodeURL(r.Context(), state.String()), http.StatusFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) HandleLoginCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if _, ok := s.sessions[state]; !ok {
		slog.Warn("state did not match")
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}

	idToken, err := s.authenticator.ExchangeCodeForToken(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		slog.Warn("failed to exchange code for token", "error", err)
		http.Error(w, "auth error", http.StatusInternalServerError)
		return
	}
	slog.Info("login success", "claims", idToken.Claims)
	cookie := &http.Cookie{
		Name:     "session",
		Value:    state,
		Path:     "/",
		Expires:  time.Now().AddDate(0, 0, 10),
		HttpOnly: true,
	}
	if strings.HasPrefix(s.config.OAuth2.RedirectURL, "https") {
		cookie.Secure = true
		cookie.SameSite = http.SameSiteNoneMode
	}
	if err := cookie.Valid(); err != nil {
		slog.Warn("bad cookie", "error", err)
	}
	http.SetCookie(w, cookie)
	r.AddCookie(cookie)

	s.sessions[state] = &idToken.Claims
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == "session" {
			delete(s.sessions, cookie.Value)
		}
	}

	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/logout.html"))
	if err != nil {
		slog.Error("template submit failure", "error", err)
		s.Handle500(w, r)
		return
	}

	if err = plate.Execute(w, nil); err != nil {
		slog.Error("template execution failure", "error", err)
		s.Handle500(w, r)
		return
	}
}
