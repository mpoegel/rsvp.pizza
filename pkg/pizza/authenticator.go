package pizza

import (
	"context"
	"slices"
	"time"
)

type Authenticator interface {
	GetToken(ctx context.Context, opt AuthTokenOptions) (*JWT, error)
	DecodeAccessToken(ctx context.Context, rawAccessToken string) (*AccessToken, error)

	GetAuthCodeURL(ctx context.Context, state string) string
	ExchangeCodeForToken(ctx context.Context, state, code string) (*IDToken, error)
	VerifyToken(ctx context.Context, rawToken string) (*IDToken, error)

	GetAuthURL() string

	IsValidSession(session string) (*TokenClaims, bool)
}

type AuthTokenOptions struct {
	Username     string
	Password     string
	GrantType    string
	RefreshToken string
}

type JWT struct {
	AccessToken      string `json:"access_token"`
	IDToken          string `json:"id_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

type AccessToken struct {
	Claims    TokenClaims
	Signature []byte
	Valid     bool
	ExpiresAt time.Time
}

type TokenClaims struct {
	Exp               int64    `json:"exp"`
	Iat               int64    `json:"iat"`
	AuthTime          int64    `json:"auth_time"`
	Jti               string   `json:"jti"`
	Iss               string   `json:"iss"`
	Aud               string   `json:"aud"`
	Sub               string   `json:"sub"`
	Typ               string   `json:"typ"`
	Azp               string   `json:"azp"`
	SessionState      string   `json:"session_state"`
	At_hash           string   `json:"at_hash"`
	Acr               string   `json:"acr"`
	Sid               string   `json:"sid"`
	EmailVerified     bool     `json:"email_verified"`
	Name              string   `json:"name"`
	PreferredUsername string   `json:"preferred_username"`
	GivenName         string   `json:"given_name"`
	FamilyName        string   `json:"family_name"`
	Email             string   `json:"email"`
	Groups            []string `json:"groups"`
	Roles             []string `json:"roles"`
}

func (c *TokenClaims) HasRole(role string) bool {
	return slices.Contains(c.Roles, role)
}

func (c *TokenClaims) InGroup(group string) bool {
	return group == "" || slices.Contains(c.Groups, group)
}

type IDToken struct {
	Claims    TokenClaims
	ExpiresAt time.Time
	Audience  []string
	Issuer    string
	Nonce     string
	Subject   string
}
