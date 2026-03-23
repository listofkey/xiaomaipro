package encrypt

import (
	"encoding/base64"
	"testing"
)

func TestAESEncryptDecryptWithRawKey(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	plain := "410522200008170839"

	cipher, err := AESEncrypt(plain, key)
	if err != nil {
		t.Fatalf("AESEncrypt failed: %v", err)
	}

	out, err := AESDecrypt(cipher, key)
	if err != nil {
		t.Fatalf("AESDecrypt failed: %v", err)
	}

	if out != plain {
		t.Fatalf("decrypt mismatch: got %q want %q", out, plain)
	}
}

func TestAESEncryptDecryptWithBase64Key(t *testing.T) {
	rawKey := []byte("0123456789abcdef0123456789abcdef")
	keyB64 := base64.StdEncoding.EncodeToString(rawKey)
	plain := "李华"

	cipher, err := AESEncrypt(plain, keyB64)
	if err != nil {
		t.Fatalf("AESEncrypt failed: %v", err)
	}

	out, err := AESDecrypt(cipher, keyB64)
	if err != nil {
		t.Fatalf("AESDecrypt failed: %v", err)
	}

	if out != plain {
		t.Fatalf("decrypt mismatch: got %q want %q", out, plain)
	}
}

func TestValidateAESKeyInvalidLength(t *testing.T) {
	if err := ValidateAESKey("your-aes-256-key-must-be-32byte"); err == nil {
		t.Fatal("expected invalid key length error, got nil")
	}
}
