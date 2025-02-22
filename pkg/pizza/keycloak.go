package pizza

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	gocloak "github.com/Nerzal/gocloak/v13"
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
	client *gocloak.GoCloak
	config OAuth2Config
	jwt    *gocloak.JWT
}

func NewKeycloak(config OAuth2Config) (*KeycloakAuthenticator, error) {
	k := &KeycloakAuthenticator{
		client: gocloak.NewClient(config.KeycloakURL),
		config: config,
	}
	ctx := context.Background()
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
		ClientID:     &k.config.ClientID,
		ClientSecret: &k.config.ClientSecret,
	}
	jwt, err := k.client.GetToken(ctx, k.config.Realm, tokOpt)
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
	token, claims, err := k.client.DecodeAccessToken(ctx, rawAccessToken, k.config.Realm)
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
