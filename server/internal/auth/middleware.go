package auth

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const userIDKey = "userId"
const userIDIntKey = "userIdInt"
const usernameKey = "username"

func setUserContext(c *gin.Context, idStr, username string) {
	c.Set(userIDKey, idStr)
	c.Set(usernameKey, username)
	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		c.Set(userIDIntKey, id)
	}
}

func extractAuth(c *gin.Context) bool {
	tokenStr, err := c.Cookie("access_token")
	if err == nil {
		claims, err := ValidateAccessToken(tokenStr)
		if err == nil {
			setUserContext(c, claims.ID, claims.Username)
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
	setUserContext(c, claims.ID, claims.Username)
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

func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		extractAuth(c)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(userIDKey)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func GetUserIDInt(c *gin.Context) (int64, bool) {
	v, ok := c.Get(userIDIntKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

func GetUsername(c *gin.Context) (string, bool) {
	v, ok := c.Get(usernameKey)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
