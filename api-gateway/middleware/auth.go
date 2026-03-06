package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth() gin.HandlerFunc {
	secret := os.Getenv("JWT_SECRET")
	return func(c *gin.Context){
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header,"Bearer "){
			c.AbortWithStatusJSON(http.StatusUnauthorized,gin.H{"error":"missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(header,"Bearer ")
		token,err := jwt.Parse(tokenStr,func(t *jwt.Token)(interface{}, error){
			if _,ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		//an identify the authenticated user without parsing the JWT again.
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			if userID, exists := claims["sub"]; exists {
				c.Request.Header.Set("X-User-ID", userID.(string))
			}
		}

		c.Next()
	}
}