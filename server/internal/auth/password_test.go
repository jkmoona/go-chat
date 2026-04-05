package auth

import (
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "securePass123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == password {
		t.Error("hash should not equal plaintext")
	}

	if err := CheckPassword(password, hash); err != nil {
		t.Errorf("CheckPassword() with correct password: error = %v", err)
	}
}

func TestCheckPasswordWrong(t *testing.T) {
	hash, _ := HashPassword("correct")

	if err := CheckPassword("wrong", hash); err == nil {
		t.Error("CheckPassword() with wrong password should return error")
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	hash1, _ := HashPassword("same")
	hash2, _ := HashPassword("same")

	if hash1 == hash2 {
		t.Error("bcrypt should produce different hashes for the same input")
	}
}
