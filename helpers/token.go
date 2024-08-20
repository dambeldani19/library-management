package helpers

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var mySigningKey = []byte("myscreetkey")

type JWTClaims struct {
	ID   int    `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int, role string) (string, error) {
	claims := JWTClaims{
		ID:   userID,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(mySigningKey)

	return ss, err
}

func ValidateToken(tokenString string) (*int, *string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, nil, errors.New("invalid token signature")
		}
		return nil, nil, errors.New("your token was expired")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, nil, errors.New("your token was expired")
	}

	return &claims.ID, &claims.Role, nil

}
