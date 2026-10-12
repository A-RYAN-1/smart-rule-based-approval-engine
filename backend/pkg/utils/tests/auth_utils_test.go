package tests

import (
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAuthUtils_Password(t *testing.T) {
	password := "my-secret-password"

	hash, err := utils.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Correct password
	err = utils.CheckPassword(password, hash)
	assert.NoError(t, err)

	// Incorrect password
	err = utils.CheckPassword("wrong-password", hash)
	assert.Error(t, err)
}

func TestAuthUtils_JWT(t *testing.T) {
	utils.SetJWTSecret("test-secret-that-is-at-least-32-bytes-long")
	userID := int64(123)
	role := "ADMIN"

	// Generate
	token, err := utils.GenerateToken(userID, role)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate
	claims, err := utils.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)

	// Invalid token
	_, err = utils.ValidateToken("invalid.token.string")
	assert.Error(t, err)

	// A token signed with another key is rejected
	utils.SetJWTSecret("a-different-secret-also-32-bytes-long!!")
	_, err = utils.ValidateToken(token)
	assert.Error(t, err)
}

func TestAuthUtils_JWTWithoutSecret(t *testing.T) {
	utils.SetJWTSecret("")
	_, err := utils.GenerateToken(1, "ADMIN")
	assert.ErrorIs(t, err, apperrors.ErrJWTSecretNotSet)
	_, err = utils.ValidateToken("any")
	assert.ErrorIs(t, err, apperrors.ErrJWTSecretNotSet)
}
