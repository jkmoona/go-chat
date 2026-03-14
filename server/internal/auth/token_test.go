package auth

import (
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	_ = Setup("test-access-secret", "test-refresh-secret", false)
	m.Run()
}

func TestGenerateAndValidateAccessToken(t *testing.T) {
	token, err := GenerateAccessToken(42, "alice", 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}

	if claims.ID != "42" {
		t.Errorf("claims.ID = %q, want %q", claims.ID, "42")
	}
	if claims.Username != "alice" {
		t.Errorf("claims.Username = %q, want %q", claims.Username, "alice")
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken(7, "bob", 24*time.Hour)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	claims, err := ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error = %v", err)
	}

	if claims.ID != "7" {
		t.Errorf("claims.ID = %q, want %q", claims.ID, "7")
	}
	if claims.Username != "bob" {
		t.Errorf("claims.Username = %q, want %q", claims.Username, "bob")
	}
}

func TestExpiredAccessToken(t *testing.T) {
	token, err := GenerateAccessToken(1, "user", -time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	_, err = ValidateAccessToken(token)
	if err == nil {
		t.Error("ValidateAccessToken() should reject expired token")
	}
}

func TestExpiredRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken(1, "user", -time.Hour)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	_, err = ValidateRefreshToken(token)
	if err == nil {
		t.Error("ValidateRefreshToken() should reject expired token")
	}
}

func TestAccessTokenCantValidateAsRefresh(t *testing.T) {
	token, _ := GenerateAccessToken(1, "user", 15*time.Minute)

	_, err := ValidateRefreshToken(token)
	if err == nil {
		t.Error("access token should not validate as refresh token")
	}
}

func TestRefreshTokenCantValidateAsAccess(t *testing.T) {
	token, _ := GenerateRefreshToken(1, "user", 24*time.Hour)

	_, err := ValidateAccessToken(token)
	if err == nil {
		t.Error("refresh token should not validate as access token")
	}
}

func TestTamperedToken(t *testing.T) {
	token, _ := GenerateAccessToken(1, "user", 15*time.Minute)

	tampered := token[:len(token)-4] + "XXXX"
	_, err := ValidateAccessToken(tampered)
	if err == nil {
		t.Error("ValidateAccessToken() should reject tampered token")
	}
}

func TestSetupRejectsEmptySecrets(t *testing.T) {
	if err := Setup("", "refresh", false); err == nil {
		t.Error("Setup() should reject empty access secret")
	}
	if err := Setup("access", "", false); err == nil {
		t.Error("Setup() should reject empty refresh secret")
	}

	// restore valid secrets for other tests
	_ = Setup("test-access-secret", "test-refresh-secret", false)
}
