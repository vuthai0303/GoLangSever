package middleware

import (
	"context"
	"net/http"
	"strings"

	Constants "AuthServer/constants"
	Models "AuthServer/models"
	Utils "AuthServer/utils"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := Utils.VerifyToken(parts[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		userData := Models.Context{
			UserID: claims.UserID,
		}

		ctx := context.WithValue(r.Context(), Constants.USER_ID, claims.UserID)
		ctx = context.WithValue(ctx, Constants.USER_DATA, userData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
