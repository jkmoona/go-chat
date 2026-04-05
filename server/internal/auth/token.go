package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MyJWTClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var (
	accessSecret  string
	refreshSecret string
	secureCookies bool
)

func Setup(accessSec, refreshSec string, secure bool) error {
	if accessSec == "" {
		return fmt.Errorf("access token secret must not be empty")
	}
	if refreshSec == "" {
		return fmt.Errorf("refresh token secret must not be empty")
	}
	accessSecret = accessSec
	refreshSecret = refreshSec
	secureCookies = secure
	return nil
}

func SecureCookies() bool {
	return secureCookies
}

func GenerateAccessToken(userID int64, username string, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, MyJWTClaims{
		ID:       strconv.Itoa(int(userID)),
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    strconv.Itoa(int(userID)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	})

	signedToken, err := token.SignedString([]byte(accessSecret))
	return signedToken, err
}

func GenerateRefreshToken(userID int64, username string, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, MyJWTClaims{
		ID:       strconv.Itoa(int(userID)),
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    strconv.Itoa(int(userID)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	})

	signedToken, err := token.SignedString([]byte(refreshSecret))
	return signedToken, err
}

func ValidateAccessToken(tokenStr string) (*MyJWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &MyJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(accessSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*MyJWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

func ValidateRefreshToken(tokenStr string) (*MyJWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &MyJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(refreshSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*MyJWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
