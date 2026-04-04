package auth

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func extractAuth(c *gin.Context) bool {
	tokenStr, err := c.Cookie("access_token")
	if err == nil {
		claims, err := ValidateAccessToken(tokenStr)
		if err == nil {
			c.Set("userId", claims.ID)
			c.Set("username", claims.Username)
			return true
		}
	}

	refreshTokenStr, err := c.Cookie("refresh_token")
	if err != nil {
		return false
	}

	claims, err := ValidateRefreshToken(refreshTokenStr)
	if err != nil {
		return false
	}

	userID, _ := strconv.ParseInt(claims.ID, 10, 64)
	newAccessToken, err := GenerateAccessToken(userID, claims.Username, 15*time.Minute)
	if err != nil {
		return false
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", newAccessToken, 900, "/", "", SecureCookies(), true)
	c.Set("userId", claims.ID)
	c.Set("username", claims.Username)
	return true
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !extractAuth(c) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		c.Next()
	}
}
