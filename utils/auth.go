package utils

import (
	"errors"
	"finance-app/config"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func GenerateToken(userID uint) (string, error) {
	config := config.LoadConfig()

	claims := jwt.MapClaims{}
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

func GetUserIDFromToken(c *gin.Context) (uint, error) {
	cfg := config.LoadConfig()

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, errors.New("authorization header is required")
	}

	// Pastikan format "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return 0, errors.New("authorization header format must be Bearer {token}")
	}

	tokenString := parts[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := uint(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, errors.New("invalid token")
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := GetUserIDFromToken(c)
		if err != nil {
			// Tambahin log
			println("JWT Error:", err.Error())

			c.JSON(401, gin.H{"error": "Unauthorized", "detail": err.Error()})
			c.Abort()
			return
		}
		c.Next()
	}
}
