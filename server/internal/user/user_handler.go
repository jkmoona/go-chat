package user

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jkmoona/go-chat/server/internal/auth"
	"github.com/jkmoona/go-chat/server/internal/httputil"
)

type Handler struct {
	Service
}

func NewHandler(s Service) *Handler {
	return &Handler{
		Service: s,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var u CreateUserReq
	if err := httputil.DecodeAndValidate(r, &u); err != nil {
		httputil.ErrorJSON(w, http.StatusBadRequest, httputil.BindingError(err))
		return
	}

	res, err := h.Service.CreateUser(r.Context(), &u)
	if err != nil {
		switch {
		case errors.Is(err, ErrUsernameTaken):
			httputil.ErrorJSON(w, http.StatusConflict, err.Error())
		default:
			httputil.ErrorJSON(w, http.StatusInternalServerError, "Failed to create user")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, res)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var user LoginUserReq
	if err := httputil.DecodeAndValidate(r, &user); err != nil {
		httputil.ErrorJSON(w, http.StatusBadRequest, httputil.BindingError(err))
		return
	}

	u, err := h.Service.Login(r.Context(), &user)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			httputil.ErrorJSON(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, ErrTokenGeneration):
			httputil.ErrorJSON(w, http.StatusInternalServerError, "Authentication failed")
		default:
			httputil.ErrorJSON(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	httputil.SetCookie(w, "access_token", u.AccessToken, 900, auth.SecureCookies(), true)
	httputil.SetCookie(w, "refresh_token", u.RefreshToken, 86400, auth.SecureCookies(), true)

	res := &LoginUserRes{
		Username: u.Username,
		ID:       u.ID,
	}
	httputil.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	httputil.SetCookie(w, "access_token", "", -1, auth.SecureCookies(), true)
	httputil.SetCookie(w, "refresh_token", "", -1, auth.SecureCookies(), true)
	httputil.MsgJSON(w, http.StatusOK, "message", "logout successful")
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := httputil.ReadCookie(r, "refresh_token")
	if err != nil {
		httputil.ErrorJSON(w, http.StatusUnauthorized, "refresh token missing")
		return
	}

	claims, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		httputil.ErrorJSON(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	userID, _ := strconv.ParseInt(claims.ID, 10, 64)

	accessToken, err := auth.GenerateAccessToken(userID, claims.Username, 15*time.Minute)
	if err != nil {
		httputil.ErrorJSON(w, http.StatusInternalServerError, "could not generate access token")
		return
	}

	httputil.SetCookie(w, "access_token", accessToken, 900, auth.SecureCookies(), true)
	httputil.MsgJSON(w, http.StatusOK, "success", "access token refreshed successfully")
}
