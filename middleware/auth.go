package middleware

import (
	"fmt"
	"kambing-cup-backend/helper"
	"net/http"
	"os"
	"strconv"
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
				helper.WriteResponse(w, http.StatusUnauthorized, false, nil, helper.ErrUnauthorized, http.StatusText(http.StatusUnauthorized))
				return

			}

			token = strings.Replace(header, "Bearer ", "", -1)
		} else if err == nil {
			token = cookie.Value
		}

		if token == "" {
			helper.WriteResponse(w, http.StatusUnauthorized, false, nil, helper.ErrUnauthorized, http.StatusText(http.StatusUnauthorized))
			return

		}

		jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, nil
			}

			return []byte(os.Getenv("SECRET")), nil
		})

		if err != nil {
			helper.WriteResponse(w, http.StatusUnauthorized, false, nil, helper.ErrUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}

		claims := jwtToken.Claims.(jwt.MapClaims)

		r.Header.Set("x-user-id", strconv.Itoa(int(claims["sub"].(float64))))
		r.Header.Set("x-user-exp", fmt.Sprintf("%f", claims["exp"].(float64)))
		r.Header.Set("x-user-role", claims["role"].(string))

		next.ServeHTTP(w, r)
	})
}

func SuperAdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("x-user-role")
		if role != "SUPERADMIN" {
			helper.WriteResponse(w, http.StatusUnauthorized, false, nil, helper.ErrUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("x-user-role")
		if role != "ADMIN" {
			helper.WriteResponse(w, http.StatusUnauthorized, false, nil, helper.ErrUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AdminOrSuperAdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("x-user-role")
		if role != "ADMIN" && role != "SUPERADMIN" {
			helper.WriteResponse(w, http.StatusUnauthorized, false, nil, helper.ErrUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}

		next.ServeHTTP(w, r)
	})
}
