package auth

import "testing"

func TestPasswordRoundTrip(t *testing.T) {
	t.Parallel()
	hash, err := HashPassword("a sufficiently long password")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(hash, "a sufficiently long password") {
		t.Fatal("expected the password to verify")
	}
	if VerifyPassword(hash, "a different long password") {
		t.Fatal("expected a different password to fail")
	}
}

func TestPasswordMinimumLength(t *testing.T) {
	t.Parallel()
	if _, err := HashPassword("short"); err == nil {
		t.Fatal("expected a short password to fail")
	}
}

func TestTokenHashIsStableAndDoesNotExposeToken(t *testing.T) {
	t.Parallel()
	plain, hash, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if hash != HashToken(plain) {
		t.Fatal("expected stable token hash")
	}
	if plain == hash {
		t.Fatal("token hash must not expose the token")
	}
}
