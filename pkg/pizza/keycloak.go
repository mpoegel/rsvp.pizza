package pizza

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	gocloak "github.com/Nerzal/gocloak/v13"
	oidc "github.com/coreos/go-oidc"
	oauth2 "golang.org/x/oauth2"
)

const (
	UserCacheExpire = 1 * time.Minute
)

type User struct {
	ID         string
	Username   string
	Enabled    bool
	FirstName  string
	LastName   string
	Email      string
	Attributes map[string][]string
}

type Group struct {
}

type KeycloakAuthenticator struct {
	client     *gocloak.GoCloak
	oauth2Conf oauth2.Config
	realm      string
	verifier   *oidc.IDTokenVerifier
	jwt        *gocloak.JWT
	sessions   map[string]*TokenClaims
}

func NewKeycloak(ctx context.Context, config OAuth2Config) (*KeycloakAuthenticator, error) {
	provider, err := oidc.NewProvider(ctx, config.KeycloakURL+"/realms/"+config.Realm)
	if err != nil {
		return nil, err
	}
	k := &KeycloakAuthenticator{
		client: gocloak.NewClient(config.KeycloakURL),
		oauth2Conf: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL + "/login/callback",
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
		realm: config.Realm,
		verifier: provider.Verifier(&oidc.Config{
			ClientID: config.ClientID,
		}),
		sessions: make(map[string]*TokenClaims),
	}
	jwt, err := k.client.LoginClient(ctx, config.ClientID, config.ClientSecret, config.Realm)
	if err != nil {
		return nil, err
	}
	k.jwt = jwt

	res, err := k.client.RetrospectToken(ctx, jwt.AccessToken, config.ClientID, config.ClientSecret, config.Realm)
	if err != nil {
		slog.Error("keycloak token retrospect failed", "error", err)
	} else {
		if !*res.Active {
			slog.Error("keycloak token not active")
		} else {
			slog.Info("keycloak token", "permissions", res.Permissions)
		}
	}

	return k, nil
}

func (k *KeycloakAuthenticator) GetToken(ctx context.Context, opt AuthTokenOptions) (*JWT, error) {
	tokOpt := gocloak.TokenOptions{
		Username:     &opt.Username,
		Password:     &opt.Password,
		GrantType:    &opt.GrantType,
		RefreshToken: &opt.RefreshToken,
		ClientID:     &k.oauth2Conf.ClientID,
		ClientSecret: &k.oauth2Conf.ClientSecret,
	}
	jwt, err := k.client.GetToken(ctx, k.realm, tokOpt)
	if err != nil {
		return nil, err
	}
	return &JWT{
		AccessToken:      jwt.AccessToken,
		IDToken:          jwt.IDToken,
		ExpiresIn:        jwt.ExpiresIn,
		RefreshExpiresIn: jwt.RefreshExpiresIn,
		RefreshToken:     jwt.RefreshToken,
		TokenType:        jwt.TokenType,
		NotBeforePolicy:  jwt.NotBeforePolicy,
		SessionState:     jwt.SessionState,
		Scope:            jwt.Scope,
	}, nil
}

func (k *KeycloakAuthenticator) DecodeAccessToken(ctx context.Context, rawAccessToken string) (*AccessToken, error) {
	token, claims, err := k.client.DecodeAccessToken(ctx, rawAccessToken, k.realm)
	if err != nil {
		return nil, err
	}
	expTime, err := token.Claims.GetExpirationTime()
	if err != nil {
		return nil, err
	}

	jsonClaims, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}
	tokenClaims := &TokenClaims{}
	if err = json.Unmarshal(jsonClaims, tokenClaims); err != nil {
		return nil, err
	}

	return &AccessToken{
		Claims:    *tokenClaims,
		Signature: token.Signature,
		Valid:     token.Valid,
		ExpiresAt: expTime.Time,
	}, nil
}

func (k *KeycloakAuthenticator) GetAuthCodeURL(ctx context.Context, state string) string {
	k.sessions[state] = nil
	return k.oauth2Conf.AuthCodeURL(state)
}

func (k *KeycloakAuthenticator) ExchangeCodeForToken(ctx context.Context, state, code string) (*IDToken, error) {
	oauth2Token, err := k.oauth2Conf.Exchange(ctx, code)
	if err != nil {
		slog.Warn("failed to exchange code for token", "error", err)
		return nil, err
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		slog.Warn("no id_token field in oauth2 token")
		return nil, err
	}

	token, err := k.VerifyToken(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}
	k.sessions[state] = &token.Claims
	return token, nil
}

func (k *KeycloakAuthenticator) VerifyToken(ctx context.Context, rawToken string) (*IDToken, error) {
	idToken, err := k.verifier.Verify(ctx, rawToken)
	if err != nil {
		slog.Warn("failed to verify ID token", "error", err)
		return nil, err
	}

	var claims TokenClaims
	if err := idToken.Claims(&claims); err != nil {
		slog.Warn("failed to get claims", "error", err)
		return nil, err
	}

	return &IDToken{
		Claims:    claims,
		ExpiresAt: idToken.Expiry,
		Audience:  idToken.Audience,
		Issuer:    idToken.Issuer,
		Nonce:     idToken.Nonce,
		Subject:   idToken.Subject,
	}, nil
}

func (k *KeycloakAuthenticator) GetAuthURL() string {
	return k.oauth2Conf.Endpoint.AuthURL
}

func (k *KeycloakAuthenticator) IsValidSession(session string) (*TokenClaims, bool) {
	claims, ok := k.sessions[session]
	if ok && claims == nil {
		// pending completed login
		return nil, true
	}
	// either bad session or auth has expired
	if !ok || time.Now().After(time.Unix(claims.Exp, 0)) {
		delete(k.sessions, session)
		return nil, false
	}
	return claims, true
}
