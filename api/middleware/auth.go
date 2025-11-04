package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func Aut(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		tokenStr := strings.TrimPrefix(h, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) { return []byte(jwtSecret), nil })
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}
		c.Set(string(userIDKey), claims["sub"])
		if em, ok := claims["email"]; ok {
			c.Set("email", em)
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(string(userIDKey))
	if !ok {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case int64:
		return t, true
	default:
		return 0, false
	}
}

func GetEmail(c *gin.Context) (string, bool) {
	v, ok := c.Get("email")
	if !ok {
		return "", false
	}
	if s, ok2 := v.(string); ok2 {
		return s, true
	}
	return "", false
}
