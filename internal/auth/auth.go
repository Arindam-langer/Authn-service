// Package auth handles JWT token creation and validation
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

var (
	secretKey      = []byte("password")
	PhoneNamespace = uuid.MustParse("d5dfb738-9226-444b-9721-a3f169f45efc")
)

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

func IsValidToken(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("there was an error in parsing")
	}
	return secretKey, nil
}
