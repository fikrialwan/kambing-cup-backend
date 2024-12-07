package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string
		header := r.Header.Get("Authorization")
		cookie, err := r.Cookie("X-TRACECO-TOKEN")

		if header != "" {
			if !strings.Contains(header, "Bearer") {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return

			}

			token = strings.Replace(header, "Bearer ", "", -1)
		} else if err == nil {
			token = cookie.Value
		}

		if token == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return

		}

		jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, nil
			}

			return []byte(os.Getenv("SECRET")), nil
		})

		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		claims := jwtToken.Claims.(jwt.MapClaims)

		r.Header.Set("x-user-id", fmt.Sprintf("%f", claims["sub"].(float64)))
		r.Header.Set("x-user-exp", fmt.Sprintf("%f", claims["exp"].(float64)))
		r.Header.Set("x-user-role", claims["role"].(string))

		next.ServeHTTP(w, r)
	})
}

func AdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("x-user-role")
		if role != "SUPERADMIN" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
