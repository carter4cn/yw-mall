package cryptox

import (
	"os"
	"strings"
	"testing"
)

// devKey 跟 start.sh 注入的开发态默认一致 — 32 字节 hex
const devKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestHmac_EmptyReturnsEmpty(t *testing.T) {
	setupKey(t)
	if got := Hmac(""); got != "" {
		t.Errorf("Hmac('') = %q, want empty", got)
	}
}

func TestHmac_Deterministic(t *testing.T) {
	setupKey(t)
	a := Hmac("13800138001")
	b := Hmac("13800138001")
	if a != b {
		t.Errorf("Hmac not deterministic: %q vs %q", a, b)
	}
	if len(a) != 64 {
		t.Errorf("expected 64-char hex, got %d", len(a))
	}
}

func TestHmac_DifferentInputsDiffer(t *testing.T) {
	setupKey(t)
	a := Hmac("alice@example.com")
	b := Hmac("bob@example.com")
	if a == b {
		t.Errorf("different inputs produced same hash")
	}
}

func TestEncrypt_Roundtrip(t *testing.T) {
	setupKey(t)
	plain := "alice@example.com"
	enc, err := Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(enc, "v1:") {
		t.Errorf("expected v1: prefix, got %q", enc)
	}
	got, err := Decrypt(enc)
	if err != nil {
		t.Fatal(err)
	}
	if got != plain {
		t.Errorf("roundtrip: got %q, want %q", got, plain)
	}
}

func setupKey(t *testing.T) {
	t.Helper()
	os.Setenv(EnvKey, devKey)
	// once.Do 保护 init，多个 test 共享同一个 cryptox 实例没问题
}
