// Package auth handles JWT token creation and validation
package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var (
	secretKey      []byte
	PhoneNamespace uuid.UUID
)

func init() {
	_ = godotenv.Load()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is required")
	}
	secretKey = []byte(secret)

	nsStr := os.Getenv("PHONE_NAMESPACE")
	if nsStr == "" {
		panic("PHONE_NAMESPACE environment variable is required")
	}
	var err error
	PhoneNamespace, err = uuid.Parse(nsStr)
	if err != nil {
		panic(fmt.Sprintf("invalid PHONE_NAMESPACE: %v", err))
	}
}

func GeneratePhoneUUID(phoneNumber string) string {
	u := uuid.NewSHA1(PhoneNamespace, []byte(phoneNumber))
	return u.String()
}

func CreateToken(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id": userID,
			"exp":     time.Now().Add(10 * time.Minute).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func IsValidToken(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("there was an error in parsing")
	}
	return secretKey, nil
}

func VerifyToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, IsValidToken)
	if err != nil {
		return 0, err
	}
	if token == nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims type")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id not found in token claims")
	}

	return int(userIDFloat), nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
