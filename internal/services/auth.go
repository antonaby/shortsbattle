package services

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

var ErrorJWKKeySecretNotDefined = errors.New("JWK_KEY_SECRET not defined")

type KeyManager struct {
	set jwk.Set
}

func NewKeyManager() (*KeyManager, error) {
	keySecret := os.Getenv("JWK_KEY_SECRET")
	if len(keySecret) == 0 {
		return nil, ErrorJWKKeySecretNotDefined
	}

	set := jwk.NewSet()
	k1, err := jwk.Import([]byte(keySecret))
	if err != nil {
		return nil, err
	}

	err = k1.Set(jwk.KeyIDKey, "sb-1")
	if err != nil {
		return nil, err
	}

	err = k1.Set(jwk.AlgorithmKey, jwa.HS512())
	if err != nil {
		return nil, err
	}

	err = set.AddKey(k1)
	if err != nil {
		return nil, err
	}

	return &KeyManager{
		set: set,
	}, nil
}

func (km *KeyManager) GetDefaultKey() (jwk.Key, error) {
	key, ok := km.set.Key(0)
	if !ok {
		return nil, common.ServiceError{
			Code:    common.ErrorJWK,
			Message: "no default key",
		}
	}

	return key, nil
}

type AuthConfig struct {
	TokenExpTime time.Duration
	Issuer       string
	Audience     string
}

type AuthService struct {
	km     *KeyManager
	ps     *PlayersService
	config AuthConfig
}

func NewAuthService(km *KeyManager, ps *PlayersService, config AuthConfig) *AuthService {
	return &AuthService{
		km:     km,
		ps:     ps,
		config: config,
	}
}

// TODO: validate init data with bot token
func (as *AuthService) NewTokenFromTgInitData(initData string) ([]byte, error) {
	parsedData, err := initDataToMap(initData)
	if err != nil {
		return nil, err
	}

	tgId, err := common.ParseTgId(parsedData["userId"])
	if err != nil {
		return nil, common.ServiceError{
			Code:    common.ErrorTgInitData,
			Message: "failed to parse tg id",
			Cause:   err,
		}
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancelFunc()

	player, err := as.ps.CheckPlayerExistsOrCreate(ctx, qg.CreatePlayerParams{
		TgID: tgId,
		TgUsername: parsedData["username"],
		TgLanguageCode: parsedData["languageCode"],
	})

	if err != nil {
		return nil, err
	}

	currentTime := time.Now()

	token, err := jwt.NewBuilder().
		JwtID(uuid.NewString()).
		Subject(strconv.FormatInt(player.TgID, 10)). 
		Audience([]string{as.config.Audience}).
		Issuer(as.config.Issuer).
		IssuedAt(currentTime).
		Expiration(currentTime.Add(as.config.TokenExpTime)).
		Build()

	if err != nil {
		return nil, common.ServiceError{
			Code:    common.ErrorJWT,
			Message: "failed to create JWT",
			Cause:   err,
		}
	}

	defKey, err := as.km.GetDefaultKey()
	if err != nil {
		return nil, err
	}

	signed, err := jwt.Sign(token, jwt.WithKey(jwa.HS512(), defKey))
	if err != nil {
		return nil, common.ServiceError{
			Code:    common.ErrorJWK,
			Message: "failed to sign JWT",
			Cause:   err,
		}
	}

	return signed, nil
}

func (as *AuthService) ParseAndValidateJwt(rawToken []byte) (jwt.Token, error) {
	token, err := jwt.Parse(
		rawToken,
		jwt.WithVerify(true),
		jwt.WithValidate(true),
		jwt.WithKeySet(as.km.set),
		jwt.WithAudience(as.config.Audience),
		jwt.WithIssuer(as.config.Issuer),
	)

	if err != nil {
		return nil, common.ServiceError{
			Code:    common.ErrorJWT,
			Message: "failed to parse JWT",
			Cause:   err,
		}
	}

	return token, nil
}

func initDataToMap(initData string) (map[string]string, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, common.ServiceError{
			Code:    common.ErrorTgInitData,
			Message: "failed to parse init data",
			Cause:   err,
		}
	}

	m := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}

	return m, nil
}
