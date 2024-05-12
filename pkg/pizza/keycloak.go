package pizza

import (
	"context"
	"time"

	gocloak "github.com/Nerzal/gocloak/v13"
	jwt "github.com/golang-jwt/jwt/v5"
	zap "go.uber.org/zap"
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

type Keycloak struct {
	client *gocloak.GoCloak
	config OAuth2Config
	jwt    *gocloak.JWT
}

func NewKeycloak(config OAuth2Config) (*Keycloak, error) {
	k := &Keycloak{
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
		Log.Error("keycloak token retrospect failed", zap.Error(err))
	} else {
		if !*res.Active {
			Log.Error("keycloak token not active")
		} else {
			Log.Info("keycloak token", zap.Any("permissions", res.Permissions))
		}
	}

	return k, nil
}

func (k *Keycloak) GetToken(ctx context.Context, opt gocloak.TokenOptions) (*gocloak.JWT, error) {
	opt.ClientID = &k.config.ClientID
	opt.ClientSecret = &k.config.ClientSecret
	return k.client.GetToken(ctx, k.config.Realm, opt)
}

func (k *Keycloak) DecodeAccessToken(ctx context.Context, token string) (*jwt.Token, *jwt.MapClaims, error) {
	return k.client.DecodeAccessToken(ctx, token, k.config.Realm)
}
