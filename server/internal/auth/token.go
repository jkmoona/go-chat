package auth

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MyJWTClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func getAccessSecret() string {
	secret := os.Getenv("ACCESS_TOKEN_SECRET")

	if secret == "" {
		panic("ACCESS_TOKEN_SECRET environment variable not set")
	}
	return secret
}
func getRefreshSecret() string {
	secret := os.Getenv("REFRESH_TOKEN_SECRET")
	if secret == "" {
		panic("REFRESH_TOKEN_SECRET environment variable not set")
	}
	return secret
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

	secret := getAccessSecret()
	signedToken, err := token.SignedString([]byte(secret))
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
	secret := getRefreshSecret()
	signedToken, err := token.SignedString([]byte(secret))
	return signedToken, err
}

func ValidateAccessToken(tokenStr string) (*MyJWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &MyJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(getAccessSecret()), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*MyJWTClaims)
	if !ok || !token.Valid {
		return nil, err
	}
	return claims, nil
}

func ValidateRefreshToken(tokenStr string) (*MyJWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &MyJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(getRefreshSecret()), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*MyJWTClaims)
	if !ok || !token.Valid {
		return nil, err
	}
	return claims, nil
}
