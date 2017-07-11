package core

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	request "github.com/dgrijalva/jwt-go/request"
	"github.com/fader2/platform/config"
	"github.com/fader2/platform/consts"
	uuid "github.com/satori/go.uuid"
	lua "github.com/yuin/gopher-lua"
)

const (
	tokenDuration = 72
	expireOffset  = 3600

	luaAccessTokenTypeName = "accessToken"
)

var (
	PrivateKey            *rsa.PrivateKey
	DefJWTExpirationDelta = time.Hour * 24
)

func registerAccessTokenType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaAccessTokenTypeName)
	L.SetField(
		mt,
		"__index",
		L.SetFuncs(L.NewTable(), luaAccessTokenMethods),
	)
	L.SetGlobal("accessToken", mt)
	L.SetField(mt, "generate", L.NewFunction(func(L *lua.LState) int {
		id := uuid.FromStringOrNil(L.CheckString(1))
		if uuid.Equal(uuid.Nil, id) {
			L.RaiseError("invalid UUID")
			return 0
		}
		tokenString, err := luaAccessToken{}.generate(id)
		if err != nil {
			L.RaiseError("generate access_token from %q: %s", id.String(), err)
			return 0
		}
		L.Push(lua.LString(tokenString))
		return 1
	}))
	L.SetField(mt, "decode", L.NewFunction(func(L *lua.LState) int {
		id, err := luaAccessToken{}.decode(L.CheckString(1))
		if err != nil {
			L.RaiseError("error decode token: %s", err)
			return 0
		}
		L.Push(lua.LString(id))
		return 1
	}))
}

var luaAccessTokenMethods = map[string]lua.LGFunction{}

type luaAccessToken struct{}

func (m luaAccessToken) generate(id uuid.UUID) (string, error) {
	token := jwt.New(jwt.SigningMethodRS512)
	token.Claims = jwtClaims{
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

func (m luaAccessToken) decode(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&jwtClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			} else {
				return config.AppConfig.PublicKey, nil
			}
		},
	)
	if err != nil {
		return "", err
	}

	if err != nil || !token.Valid || token.Claims.Valid() != nil {
		return "", consts.ErrUnauthorized
	}

	claims := token.Claims.(*jwtClaims)
	return claims.UserID.String(), nil
}

type jwtClaims struct {
	UserID uuid.UUID

	jwt.StandardClaims
}

// extractor token from cookie

var _ request.Extractor = (*ExtractorTokenFromCookie)(nil)

type ExtractorTokenFromCookie struct {
	CookieName string
}

func (e ExtractorTokenFromCookie) ExtractToken(r *http.Request) (string, error) {
	c, err := r.Cookie(e.CookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}
