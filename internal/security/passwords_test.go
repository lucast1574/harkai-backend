package security

import (
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestProductionPasswordHash(t *testing.T) {
	hash, err := HashPassword("sample-test-password-123")
	if err != nil {
		t.Fatal(err)
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil || cost != 12 {
		t.Fatalf("wrong work factor: %d, %v", cost, err)
	}
	if !CheckPassword(hash, "sample-test-password-123") || CheckPassword(hash, "wrong-password") {
		t.Fatal("password verification failed")
	}
	if CheckPassword("", "any-password") {
		t.Fatal("passwordless Google account accepted password login")
	}
}
