package core

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/fader2/platform/config"
	"github.com/fader2/platform/consts"
	"github.com/fader2/platform/objects"

	jwt "github.com/dgrijalva/jwt-go"

	uuid "github.com/satori/go.uuid"
)

const (
	tokenDuration = 72
	expireOffset  = 3600
)

var PrivateKey *rsa.PrivateKey
var DefJWTExpirationDelta = time.Hour * 24

func GenerateToken(id uuid.UUID) (string, error) {
	token := jwt.New(jwt.SigningMethodRS512)
	token.Claims = TokenJWTClaims{
		UserID: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(
				time.Hour * time.Duration(DefJWTExpirationDelta),
			).Unix(),
			IssuedAt: time.Now().Unix(),
		},
	}
	tokenString, err := token.SignedString(config.AppConfig.PrivateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func Authenticate(tokStr string, s objects.Storer) (*objects.User, error) {
	token, err := jwt.ParseWithClaims(
		tokStr,
		&TokenJWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			} else {
				return config.AppConfig.PublicKey, nil
			}
		},
	)
	if err != nil {
		return nil, err
	}

	if err != nil || !token.Valid || token.Claims.Valid() != nil {
		return nil, consts.ErrUnauthorized
	}

	claims := token.Claims.(*TokenJWTClaims)

	user, err := objects.GetUser(s, claims.UserID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
