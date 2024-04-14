package auth

import (
	"backend-trainee-assignment-2024/internal/errs"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_GenerateToken(t *testing.T) {
	secretKey := "my_secret_key"
	authService := NewAuthService(secretKey)

	token, err := authService.GenerateToken(1, "admin")
	fmt.Println(token)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims := UserClaims{}
	_, err = jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "admin", claims.UserType)
}

func TestAuthService_ValidateToken(t *testing.T) {
	authService := NewAuthService("secret_key")

	// Проверка валидности токена
	token, err := authService.GenerateToken(1, "admin")
	assert.NoError(t, err)

	userType, err := authService.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "admin", userType)

	// Проверка ExpiresAt токена
	claims := UserClaims{
		UserID:   1,
		UserType: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	_, err = expiredToken.SignedString([]byte("secret_key"))
	assert.NoError(t, err)

	_, err = authService.ValidateToken(expiredToken.Raw)
	assert.ErrorIs(t, err, errs.ErrTokenExpired)

	// Test invalid token
	_, err = authService.ValidateToken("invalid_token")
	assert.ErrorIs(t, err, errs.ErrInvalidToken)
}
