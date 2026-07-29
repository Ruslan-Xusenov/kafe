package middleware

import (
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/username/kafe-backend/internal/pkg/security"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		var tokenString string

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in Bearer token format"})
				c.Abort()
				return
			}
			tokenString = parts[1]
		} else if queryToken := c.Query("token"); queryToken != "" {
			// URL ?token= is ONLY allowed for WebSocket upgrade requests.
			// Regular HTTP requests must use the Authorization header to avoid
			// JWT tokens appearing in access logs and browser history.
			upgradeHeader := c.GetHeader("Upgrade")
			if !strings.EqualFold(upgradeHeader, "websocket") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header talab qilinadi"})
				c.Abort()
				return
			}
			tokenString = queryToken
		} else if wsToken := c.GetHeader("Sec-WebSocket-Protocol"); wsToken != "" {
			// Browser WebSocket clients cannot set Authorization headers. The
			// frontend sends the token as the auth.<jwt> subprotocol instead of
			// putting it in the URL query string.
			tokenString = strings.TrimSpace(wsToken)
			tokenString = strings.TrimPrefix(tokenString, "auth.")
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization talab qilinadi"})
			c.Abort()
			return
		}
		token, err := security.ValidateToken(tokenString)
		if err != nil || token == nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		var userID int
		switch value := claims["user_id"].(type) {
		case float64:
			if value <= 0 || math.Trunc(value) != value {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				c.Abort()
				return
			}
			userID = int(value)
		case int:
			userID = value
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		role, ok := claims["role"].(string)
		if !ok || strings.TrimSpace(role) == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Role not found in context"})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role context"})
			c.Abort()
			return
		}

		allowed := false
		for _, r := range roles {
			if r == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this resource"})
			c.Abort()
			return
		}

		c.Next()
	}
}
