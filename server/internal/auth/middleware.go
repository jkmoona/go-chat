package auth

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("access_token")
		if err != nil {
			handleTokenRefresh(c)
			return
		}

		claims, err := ValidateAccessToken(tokenStr)
		if err != nil {
			handleTokenRefresh(c)
			return
		}

		c.Set("userId", claims.ID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

func handleTokenRefresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	claims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	userID, _ := strconv.ParseInt(claims.ID, 10, 64)
	newAccessToken, err := GenerateAccessToken(userID, claims.Username, 15*time.Minute)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate new access token"})
		return
	}
	c.SetCookie("access_token", newAccessToken, 900, "/", "localhost", true, true)
	c.Set("userId", claims.ID)
	c.Set("username", claims.Username)
	c.Next()
}
