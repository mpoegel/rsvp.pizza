package pizza

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"time"
)

type FileBasedAuth struct {
	filename string
}

func NewFileBasedAuth(filename string) *FileBasedAuth {
	a := &FileBasedAuth{
		filename: filename,
	}
	return a
}

func (a *FileBasedAuth) GetToken(ctx context.Context, opt AuthTokenOptions) (*JWT, error) {
	return &JWT{
		AccessToken:      "access123",
		IDToken:          "id123",
		ExpiresIn:        math.MaxInt,
		RefreshExpiresIn: math.MaxInt,
		RefreshToken:     "refresh123",
		TokenType:        "fileToken",
		NotBeforePolicy:  1,
		SessionState:     "session123",
		Scope:            "all",
	}, nil
}

func (a *FileBasedAuth) DecodeAccessToken(ctx context.Context, rawAccessToken string) (*AccessToken, error) {
	claims, err := a.loadClaims()
	if err != nil {
		return nil, err
	}
	return &AccessToken{
		Claims:    *claims,
		Signature: []byte("fakeSignature"),
		Valid:     true,
		ExpiresAt: time.Now().Add(100 * time.Hour),
	}, nil
}

func (a *FileBasedAuth) GetAuthCodeURL(ctx context.Context, state string) string {
	return "/login/callback?state=state123"
}

func (a *FileBasedAuth) ExchangeCodeForToken(ctx context.Context, state, code string) (*IDToken, error) {
	claims, err := a.loadClaims()
	if err != nil {
		return nil, err
	}
	return &IDToken{
		Claims:    *claims,
		ExpiresAt: time.Now().Add(100 * time.Hour),
		Audience:  []string{"all"},
		Issuer:    a.filename,
		Nonce:     "nonce123",
		Subject:   "subject123",
	}, nil
}

func (a *FileBasedAuth) VerifyToken(ctx context.Context, rawToken string) (*IDToken, error) {
	return a.ExchangeCodeForToken(ctx, "state123", "rawToken123")
}

func (a *FileBasedAuth) GetAuthURL() string {
	return "/"
}

func (a *FileBasedAuth) IsValidSession(session string) (*TokenClaims, bool) {
	if claims, err := a.loadClaims(); err != nil {
		return nil, false
	} else {
		return claims, true
	}
}

func (a *FileBasedAuth) loadClaims() (*TokenClaims, error) {
	var claims TokenClaims
	fp, err := os.Open(a.filename)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(fp)
	if err = decoder.Decode(&claims); err != nil {
		return nil, err
	}
	return &claims, nil
}
