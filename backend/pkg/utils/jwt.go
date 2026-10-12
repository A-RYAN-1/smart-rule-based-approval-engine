package utils

import (
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

// SetJWTSecret sets the signing key; main calls it once with JWT_SECRET
// from the config. Tokens cannot be created or verified before that.
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int64, role string) (string, error) {
	if len(jwtSecret) == 0 {
		return "", apperrors.ErrJWTSecretNotSet
	}
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(24 * time.Hour),
			),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*JWTClaims, error) {
	if len(jwtSecret) == 0 {
		return nil, apperrors.ErrJWTSecretNotSet
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Ensure token method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperrors.ErrUnexpectedSigningMethod
			}
			return jwtSecret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, apperrors.ErrInvalidToken
	}

	return claims, nil
}
