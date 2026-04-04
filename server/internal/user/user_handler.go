package user

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jkmoona/go-chat/server/internal/auth"
)

func bindingError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		e := ve[0]
		field := strings.ToLower(e.Field())
		switch e.Tag() {
		case "required":
			return field + " is required"
		case "min":
			return field + " must be at least " + e.Param() + " characters"
		case "max":
			return field + " must be at most " + e.Param() + " characters"
		case "alphanum":
			return field + " must contain only letters and numbers"
		}
	}
	return "invalid request"
}

type Handler struct {
	Service
}

func NewHandler(s Service) *Handler {
	return &Handler{
		Service: s,
	}
}

func (h *Handler) CreateUser(c *gin.Context) {
	var u CreateUserReq
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindingError(err)})
		return
	}

	res, err := h.Service.CreateUser(c.Request.Context(), &u)
	if err != nil {
		switch {
		case errors.Is(err, ErrUsernameTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *Handler) Login(c *gin.Context) {
	var user LoginUserReq
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindingError(err)})
		return
	}

	u, err := h.Service.Login(c.Request.Context(), &user)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, ErrTokenGeneration):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", u.AccessToken, 900, "/", "", auth.SecureCookies(), true)
	c.SetCookie("refresh_token", u.RefreshToken, 86400, "/", "", auth.SecureCookies(), true)

	res := &LoginUserRes{
		Username: u.Username,
		ID:       u.ID,
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", "", -1, "/", "", auth.SecureCookies(), true)
	c.SetCookie("refresh_token", "", -1, "/", "", auth.SecureCookies(), true)
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (h *Handler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token missing"})
		return
	}

	claims, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	userID, _ := strconv.ParseInt(claims.ID, 10, 64)

	accessToken, err := auth.GenerateAccessToken(userID, claims.Username, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate access token"})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", accessToken, 900, "/", "", auth.SecureCookies(), true)
	c.JSON(http.StatusOK, gin.H{"success": "access token refreshed successfully"})

}
