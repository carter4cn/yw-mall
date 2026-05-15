// Package cryptox encrypts/decrypts sensitive PII fields (phone, real_name,
// id_card_no, totp_secret, ...) at rest using AES-256-GCM.
//
// Format on disk: "v1:" + base64url(nonce(12) || ciphertext || gcm_tag(16))
// The "v1:" prefix lets dual-write/dual-read code distinguish ciphertext from
// pre-migration plaintext rows during the Sprint 5 backfill.
//
// Key source: env MALL_FIELD_ENCRYPTION_KEY = 64 hex chars (32 bytes).
// Missing/invalid key → MustInit panics with a clear message so the process
// fails fast instead of silently writing garbage.
package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

const (
	// CiphertextPrefix tags ciphertext so we can distinguish it from legacy
	// plaintext during the dual-read/dual-write transition.
	CiphertextPrefix = "v1:"

	// EnvKey is the env var name holding the 32-byte hex-encoded AES key.
	EnvKey = "MALL_FIELD_ENCRYPTION_KEY"

	keyByteLen   = 32 // AES-256
	nonceByteLen = 12 // GCM standard nonce
)

var (
	once   sync.Once
	gcm    cipher.AEAD
	initErr error
)

// MustInit eagerly loads the key from env and panics on misconfiguration.
// Safe to call multiple times; subsequent calls are no-ops.
func MustInit() {
	if err := initOnce(); err != nil {
		panic(fmt.Sprintf("cryptox.MustInit: %v", err))
	}
}

func initOnce() error {
	once.Do(func() {
		raw := strings.TrimSpace(os.Getenv(EnvKey))
		if raw == "" {
			initErr = fmt.Errorf("env %s is empty; generate with: openssl rand -hex 32", EnvKey)
			return
		}
		key, err := hex.DecodeString(raw)
		if err != nil {
			initErr = fmt.Errorf("env %s is not valid hex: %w", EnvKey, err)
			return
		}
		if len(key) != keyByteLen {
			initErr = fmt.Errorf("env %s decodes to %d bytes, want %d", EnvKey, len(key), keyByteLen)
			return
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			initErr = fmt.Errorf("aes.NewCipher: %w", err)
			return
		}
		aead, err := cipher.NewGCM(block)
		if err != nil {
			initErr = fmt.Errorf("cipher.NewGCM: %w", err)
			return
		}
		gcm = aead
	})
	return initErr
}

// Encrypt returns "v1:" + base64url(nonce || ciphertext || tag).
// Safe to call before MustInit — it lazily initialises on first use.
func Encrypt(plaintext string) (string, error) {
	if err := initOnce(); err != nil {
		return "", err
	}
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, nonceByteLen)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	out := make([]byte, 0, nonceByteLen+len(sealed))
	out = append(out, nonce...)
	out = append(out, sealed...)
	return CiphertextPrefix + base64.RawURLEncoding.EncodeToString(out), nil
}

// Decrypt accepts the "v1:..." form produced by Encrypt and returns the
// original plaintext. Empty in → empty out (so callers can route empty
// columns through unchanged).
func Decrypt(b64 string) (string, error) {
	if err := initOnce(); err != nil {
		return "", err
	}
	if b64 == "" {
		return "", nil
	}
	if !strings.HasPrefix(b64, CiphertextPrefix) {
		return "", errors.New("cryptox.Decrypt: missing v1: prefix")
	}
	raw, err := base64.RawURLEncoding.DecodeString(b64[len(CiphertextPrefix):])
	if err != nil {
		return "", fmt.Errorf("cryptox.Decrypt: base64: %w", err)
	}
	if len(raw) < nonceByteLen+gcm.Overhead() {
		return "", errors.New("cryptox.Decrypt: ciphertext too short")
	}
	nonce, sealed := raw[:nonceByteLen], raw[nonceByteLen:]
	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", fmt.Errorf("cryptox.Decrypt: %w", err)
	}
	return string(plain), nil
}

// IsCiphertext returns true if s looks like a v1 ciphertext blob. Used by
// dual-read paths so we can decrypt new rows and pass legacy plaintext rows
// through unchanged until the Sprint 5 migration runs.
func IsCiphertext(s string) bool {
	return strings.HasPrefix(s, CiphertextPrefix) && len(s) > 20
}

// DecryptIfCiphertext is sugar for the read path: returns Decrypt(s) when s is
// a ciphertext blob, otherwise returns s as-is. Errors propagate so callers
// can distinguish "decrypt failed" from "was plaintext".
func DecryptIfCiphertext(s string) (string, error) {
	if !IsCiphertext(s) {
		return s, nil
	}
	return Decrypt(s)
}
