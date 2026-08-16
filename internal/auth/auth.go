package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	passwordTime    = 3
	passwordMemory  = 64 * 1024
	passwordThreads = 4
	passwordKeyLen  = 32
)

func HashPassword(password string) (string, error) {
	if len(password) < 12 {
		return "", fmt.Errorf("password must be at least 12 characters")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, passwordTime, passwordMemory, passwordThreads, passwordKeyLen)
	return fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", passwordMemory, passwordTime, passwordThreads, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

func VerifyPassword(encoded, password string) bool {
	var memory uint32
	var time uint32
	var threads uint8
	var saltText string
	var keyText string
	if _, err := fmt.Sscanf(encoded, "argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", &memory, &time, &threads, &saltText, &keyText); err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(saltText)
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(keyText)
	if err != nil {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(want)))
	return subtleEqual(got, want)
}

func NewToken() (plain string, hash string, err error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", "", fmt.Errorf("generate token: %w", err)
	}
	plain = base64.RawURLEncoding.EncodeToString(data)
	return plain, HashToken(plain), nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func subtleEqual(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	var result byte
	for index := range left {
		result |= left[index] ^ right[index]
	}
	return result == 0
}
