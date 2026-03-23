package jwt

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Int64String int64

func (v Int64String) Int64() int64 {
	return int64(v)
}

func (v Int64String) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(v), 10))
}

func (v *Int64String) UnmarshalJSON(data []byte) error {
	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		parsed, err := strconv.ParseInt(stringValue, 10, 64)
		if err != nil {
			return err
		}

		*v = Int64String(parsed)
		return nil
	}

	var numberValue json.Number
	if err := json.Unmarshal(data, &numberValue); err != nil {
		return err
	}

	parsed, err := strconv.ParseInt(numberValue.String(), 10, 64)
	if err != nil {
		return err
	}

	*v = Int64String(parsed)
	return nil
}

type Claims struct {
	UserID Int64String `json:"user_id"`
	Type   string      `json:"type"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID int64, secret string, expireSeconds int64) (string, error) {
	claims := Claims{
		UserID: Int64String(userID),
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "xiaomaipro",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID int64, secret string, expireSeconds int64) (string, error) {
	claims := Claims{
		UserID: Int64String(userID),
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "xiaomaipro",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func ParseTokenAllowExpired(tokenStr, secret string) (*Claims, error) {
	parser := jwt.Parser{
		SkipClaimsValidation: true,
	}

	token, err := parser.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func RemainingTTL(claims *Claims) time.Duration {
	if claims == nil || claims.ExpiresAt == nil {
		return 0
	}

	return time.Until(claims.ExpiresAt.Time)
}
