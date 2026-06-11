package utils

import (
	"errors"
	"time"
	"wedding-system/config"
	"wedding-system/models"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(user *models.User) (string, error) {
	claims := config.CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Name:     user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(config.AppConfig.JWTExpireHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "wedding-system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

func ParseToken(tokenString string) (*config.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &config.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*config.CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
