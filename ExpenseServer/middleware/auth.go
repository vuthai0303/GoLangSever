package middleware

import (
	"context"
	"net/http"
	"strings"

	"ExpenseServer/utils"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		claims, err := utils.VerifyToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		reqCtx := context.WithValue(c.Request.Context(), UserIDKey, claims.UserID)
		c.Request = c.Request.WithContext(reqCtx)
		c.Next()
	}
}
