package config

import (
	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	ServerPort     string
	JWTSecret      string
	JWTExpireHours int
	DBPath         string
}

var AppConfig = Config{
	ServerPort:     "8627",
	JWTSecret:      "wedding-system-secret-key-2024",
	JWTExpireHours: 24,
	DBPath:         "./data/wedding.db",
}

type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Name     string `json:"name"`
	jwt.RegisteredClaims
}
