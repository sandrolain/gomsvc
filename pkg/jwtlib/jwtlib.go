package jwtlib

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTParams[T any] struct {
	Subject   string
	Issuer    string
	Secret    []byte
	ExpiresAt time.Time
	Data      T
}

type Claims[T any] struct {
	jwt.RegisteredClaims
	Data T `json:"dat,omitempty"`
}

func CreateJWT[T any](params JWTParams[T]) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, Claims[T]{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(params.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    params.Issuer,
			Subject:   params.Subject,
		},
		params.Data,
	}).SignedString(params.Secret)
}

func ParseJWT[T any](jwtString string, params JWTParams[T]) (res *Claims[T], err error) {
	if jwtString == "" {
		err = errors.New("the jwt string is empty")
		return
	}
	token, err := jwt.ParseWithClaims(jwtString, &Claims[T]{}, func(token *jwt.Token) (interface{}, error) {
		return params.Secret, nil
	})
	if err != nil {
		return
	}
	if !token.Valid {
		err = errors.New("invalid JWT")
		return
	}
	res, ok := token.Claims.(*Claims[T])
	if !ok || res.IssuedAt.Unix() == 0 {
		err = errors.New("cannot obtain JWT claims")
		return
	}
	return
}

func ExtractInfoFromJWT[T any](jwtString string) (res *Claims[T], err error) {
	if jwtString == "" {
		err = errors.New("the jwt string is empty")
		return
	}
	token, _, err := new(jwt.Parser).ParseUnverified(jwtString, &Claims[T]{})
	if err != nil {
		return
	}
	res, ok := token.Claims.(*Claims[T])
	if !ok || res.IssuedAt.Unix() == 0 {
		err = errors.New("cannot obtain JWT Info")
		return
	}
	return
}
