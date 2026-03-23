package jwt

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	jwtv4 "github.com/golang-jwt/jwt/v4"
)

func TestGenerateAccessTokenStoresUserIDAsString(t *testing.T) {
	const userID int64 = 2035026620841992192

	token, err := GenerateAccessToken(userID, "secret", 3600)
	if err != nil {
		t.Fatalf("GenerateAccessToken error: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token format: %s", token)
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload error: %v", err)
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("unmarshal payload error: %v", err)
	}

	if claims["user_id"] != "2035026620841992192" {
		t.Fatalf("unexpected user_id claim: %#v", claims["user_id"])
	}
}

func TestParseTokenSupportsLegacyNumericUserID(t *testing.T) {
	token, err := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, jwtv4.MapClaims{
		"user_id": 2035026620841992192,
		"type":    "access",
		"iss":     "xiaomaipro",
	}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString error: %v", err)
	}

	claims, err := ParseTokenAllowExpired(token, "secret")
	if err != nil {
		t.Fatalf("ParseTokenAllowExpired error: %v", err)
	}

	if claims.UserID.Int64() != 2035026620841992192 {
		t.Fatalf("unexpected user id: %d", claims.UserID.Int64())
	}
}
