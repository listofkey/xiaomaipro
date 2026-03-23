package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// AESEncrypt encrypts plaintext with AES-GCM and returns base64(nonce+ciphertext).
func AESEncrypt(plaintext, key string) (string, error) {
	aesKey, err := normalizeKey(key)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt decrypts base64(nonce+ciphertext) produced by AESEncrypt.
func AESDecrypt(cipherB64, key string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", err
	}

	aesKey, err := normalizeKey(key)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// ValidateAESKey validates that key is usable for AES(16/24/32 bytes),
// either as raw text or base64-decoded bytes.
func ValidateAESKey(key string) error {
	_, err := normalizeKey(key)
	return err
}

// normalizeKey accepts:
// 1) raw text key with 16/24/32 bytes
// 2) base64 key whose decoded length is 16/24/32 bytes
func normalizeKey(key string) ([]byte, error) {
	k := strings.TrimSpace(key)
	if k == "" {
		return nil, errors.New("aes key is empty")
	}

	if decoded, err := base64.StdEncoding.DecodeString(k); err == nil {
		if isValidKeyLen(len(decoded)) {
			return decoded, nil
		}
	}

	raw := []byte(k)
	if isValidKeyLen(len(raw)) {
		return raw, nil
	}

	return nil, fmt.Errorf(
		"invalid aes key length: got %d bytes, need 16/24/32 bytes (raw or base64-decoded)",
		len(raw),
	)
}

func isValidKeyLen(n int) bool {
	return n == 16 || n == 24 || n == 32
}

// MaskIDCard masks ID card with first 3 and last 4 characters visible.
func MaskIDCard(idCard string) string {
	if len(idCard) < 8 {
		return "***"
	}
	runes := []rune(idCard)
	masked := make([]rune, len(runes))
	for i := range runes {
		if i < 3 || i >= len(runes)-4 {
			masked[i] = runes[i]
		} else {
			masked[i] = '*'
		}
	}
	return string(masked)
}

// MaskPhone masks phone with first 3 and last 4 characters visible.
func MaskPhone(phone string) string {
	if len(phone) < 8 {
		return "***"
	}
	runes := []rune(phone)
	return string(runes[:3]) + "****" + string(runes[len(runes)-4:])
}

// MaskRealName masks a real name except first character.
func MaskRealName(name string) string {
	runes := []rune(name)
	if len(runes) == 0 {
		return ""
	}
	if len(runes) == 1 {
		return string(runes[0])
	}
	masked := string(runes[0])
	for i := 1; i < len(runes); i++ {
		masked += "*"
	}
	return masked
}
