package security

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/kirillmashkov/gophermart/internal/app"
	"github.com/kirillmashkov/gophermart/internal/util"
)

type UserIDType string


func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		u := UserIDType("userID")
		if auth == "" {
			app.Log.Warn("No token")
			c := context.WithValue(r.Context(), u, "")
			next.ServeHTTP(w, r.WithContext(c))
			return
		}

		splits := strings.Split(auth, "Bearer ")
		token := splits[1]
		err, userID := getUserIDFromToken(token)

		if err != nil {
			c := context.WithValue(r.Context(), u, "")
			next.ServeHTTP(w, r.WithContext(c))
			return
		}

		c := context.WithValue(r.Context(), u, userID)
		next.ServeHTTP(w, r.WithContext(c))
	})
}

func getUserIDFromToken(tokenString string) (error, string) {

	claims := &util.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(app.SecretKey), nil
		})

	if err != nil {
		app.Log.Warn("Can't parse token")
		return err, ""
	}

	if !token.Valid {
		app.Log.Warn("Token is not valid")
		return nil, ""
	}

	if claims.UserID == "" {
		app.Log.Warn("Token doesn't contain UserID")
		return nil, ""
	}

	return nil, claims.UserID
}

