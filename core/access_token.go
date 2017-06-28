package core

import (
	jwt "github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
)

type TokenJWTClaims struct {
	UserID uuid.UUID

	jwt.StandardClaims
}
