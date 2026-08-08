package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// ContextUserID is the gin context key for the authenticated user id.
	ContextUserID = "userID"
	// ContextIsAdmin is the gin context key for admin flag.
	ContextIsAdmin = "isAdmin"
	// ContextEmail is the gin context key for email.
	ContextEmail = "email"
)

// Middleware validates Bearer JWTs.
func Middleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := ParseToken(secret, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextIsAdmin, claims.IsAdmin)
		c.Set(ContextEmail, claims.Email)
		c.Next()
	}
}

// RequireAdmin rejects non-admin users.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get(ContextIsAdmin)
		isAdmin, ok := raw.(bool)
		if !exists || !ok || !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin required"})
			return
		}
		c.Next()
	}
}

// UserIDFromContext returns the authenticated user id.
func UserIDFromContext(c *gin.Context) string {
	v, _ := c.Get(ContextUserID)
	s, _ := v.(string)
	return s
}

