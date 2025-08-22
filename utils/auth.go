package utils

import (
	"errors"
	"finance-app/config"
	"log"
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

// Tambahkan di bagian atas file
const (
	DebugLevelNone = iota
	DebugLevelBasic
	DebugLevelVerbose
)

var DebugLevel = DebugLevelBasic // Atur level debug sesuai kebutuhan

func GetUserIDFromToken(c *gin.Context) (uint, error) {
	config := config.LoadConfig()

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		if DebugLevel >= DebugLevelBasic {
			log.Printf("DEBUG: Authorization header is missing")
		}
		return 0, errors.New("authorization header is required")
	}

	if DebugLevel >= DebugLevelVerbose {
		log.Printf("DEBUG: Raw Authorization header: %s", authHeader)
	}

	// kalau tidak ada prefix "Bearer", otomatis tambahin
	if !strings.HasPrefix(authHeader, "Bearer ") {
		authHeader = "Bearer " + authHeader
		if DebugLevel >= DebugLevelVerbose {
			log.Printf("DEBUG: Added Bearer prefix. New header: %s", authHeader)
		}
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[1] == "" {
		if DebugLevel >= DebugLevelBasic {
			log.Printf("DEBUG: Invalid authorization header format. Parts: %v", parts)
		}
		return 0, errors.New("invalid authorization header format")
	}
	tokenString := parts[1]

	if DebugLevel >= DebugLevelVerbose {
		log.Printf("DEBUG: Extracted token string: %s", tokenString)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			method := token.Method.Alg()
			if DebugLevel >= DebugLevelBasic {
				log.Printf("DEBUG: Unexpected signing method: %s", method)
			}
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		if DebugLevel >= DebugLevelBasic {
			log.Printf("DEBUG: JWT parsing error: %v", err)
		}
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if DebugLevel >= DebugLevelVerbose {
			log.Printf("DEBUG: All JWT claims: %+v", claims)
		}

		userIDFloat, ok := claims["user_id"]
		if !ok {
			if DebugLevel >= DebugLevelBasic {
				log.Printf("DEBUG: user_id claim not found in token")
			}
			return 0, errors.New("user_id claim not found in token")
		}

		userID, ok := userIDFloat.(float64)
		if !ok {
			if DebugLevel >= DebugLevelBasic {
				log.Printf("DEBUG: user_id is not a number. Type: %T, Value: %v", userIDFloat, userIDFloat)
			}
			return 0, errors.New("invalid user_id format in token")
		}

		if DebugLevel >= DebugLevelBasic {
			log.Printf("DEBUG: Successfully extracted user_id: %d", uint(userID))
		}
		return uint(userID), nil
	}

	if DebugLevel >= DebugLevelBasic {
		log.Printf("DEBUG: Token is invalid or claims cannot be extracted")
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
