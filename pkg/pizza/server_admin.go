package pizza

import (
	"net/http"
	"path"
	"text/template"
	"time"

	zap "go.uber.org/zap"
)

type AdminPageData struct {
}

func (s *Server) HandleAdmin(w http.ResponseWriter, r *http.Request) {
	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/admin.html"))
	if err != nil {
		Log.Error("template submit failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
	data := AdminPageData{}

	var claims *TokenClaims
	for _, cookie := range r.Cookies() {
		if cookie.Name == "session" {
			var ok bool
			claims, ok = s.sessions[cookie.Value]
			// either bad session or auth has expired
			if !ok || time.Now().After(time.Unix(claims.Exp, 0)) {
				delete(s.sessions, cookie.Value)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
		}
	}
	if claims == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if !claims.HasRole("pizza_host") {
		s.Handle4xx(w, r)
		return
	}

	if err = plate.Execute(w, data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
}
