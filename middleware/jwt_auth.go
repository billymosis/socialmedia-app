package AppMiddleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/billymosis/socialmedia-app/handler/render"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

func ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		head := r.Header.Get("Authorization")
		if head == "" {
			render.Unauthorized(w, errors.New("No header found"))
			return
		}
		authHeader := strings.Split(head, "Bearer ")
		if len(authHeader) != 2 {
			render.Forbidden(w, errors.New("authorization not found in header"))
			return
		}
		jwtToken := authHeader[1]
		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			render.Forbidden(w, err)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), "userAuthCtx", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		http.Error(w, "err.Error()", http.StatusUnauthorized)

	},
	)
}
