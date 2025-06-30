package util

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type SecurityUtil struct {
	tokenExp time.Duration
	secretKey string
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func NewSecurityUtil(tokenExp time.Duration, secretKey string) *SecurityUtil {
	return &SecurityUtil{tokenExp: tokenExp, secretKey: secretKey}
}

func (s *SecurityUtil) BuildJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
