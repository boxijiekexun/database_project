package middleware

import (
	"net/http"
	"strings"

	"project1/internal/auth"
	"project1/internal/response"
	"project1/internal/user"

	"github.com/gin-gonic/gin"
)

const claimsKey = "authClaims"

func AuthRequired(tokenManager *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, "missing bearer token")
			c.Abort()
			return
		}

		claims, err := tokenManager.Parse(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Set(claimsKey, claims)
		c.Next()
	}
}

func RequireAnyRole(required ...user.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}

		for _, role := range required {
			if claims.Role == role {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "forbidden")
		c.Abort()
	}
}

func RequireRole(required user.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}

		if claims.Role != required {
			response.Error(c, http.StatusForbidden, "forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}

func ClaimsFromContext(c *gin.Context) (auth.Claims, bool) {
	value, exists := c.Get(claimsKey)
	if !exists {
		return auth.Claims{}, false
	}

	claims, ok := value.(auth.Claims)
	return claims, ok
}
