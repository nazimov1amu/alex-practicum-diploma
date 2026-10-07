package common

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/resources"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type claims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func Sign(userID uuid.UUID, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	})

	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func UserIDFromToken(tokenString, tokenType string) (uuid.UUID, error) {
	parsed := &claims{}
	_, err := jwt.ParseWithClaims(tokenString, parsed, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return uuid.Nil, err
	}
	if parsed.TokenType != tokenType {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	userID, err := uuid.Parse(parsed.Subject)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func UserIDFromContext(c *echo.Context) (uuid.UUID, error) {
	userID, _ := c.Get(resources.UserIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		return uuid.Nil, apperrors.Unauthorized(resources.InvalidAccessToken)
	}
	return userID, nil
}
