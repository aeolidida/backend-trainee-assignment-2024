package auth

import (
	"fmt"
	"time"

	"backend-trainee-assignment-2024/internal/errs"

	"github.com/golang-jwt/jwt/v5"
)

// Структура для хранения пользовательской информации в токене
type UserClaims struct {
	UserID   int    `json:"user_id"`
	UserType string `json:"user_type"`
	jwt.RegisteredClaims
}

type AuthService struct {
	secretKey []byte
}

func NewAuthService(secretKey string) *AuthService {
	return &AuthService{secretKey: []byte(secretKey)}
}

// Функция для генерации токена для тестов
func (s *AuthService) GenerateToken(userID int, userType string) (string, error) {
	claims := UserClaims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// Функция для валидации и проверки типа пользователя
func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	if tokenString == "" {
		return "", errs.ErrRequiredToken
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		expirationTime, ok := claims["exp"].(float64)
		if !ok {
			return "", errs.ErrInvalidExpirationTime
		}

		if time.Unix(int64(expirationTime), 0).Before(time.Now()) {
			return "", errs.ErrTokenExpired
		}

		userType, ok := claims["user_type"].(string)
		if !ok {
			return "", errs.ErrInvalidUserType
		}

		return userType, nil
	}
	return "", errs.ErrInvalidToken
}
