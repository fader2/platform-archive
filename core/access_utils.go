package core

import (
	"net/http"

	request "github.com/dgrijalva/jwt-go/request"
)

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
