// Package internal is for internal functions or helpers functions nothing more
package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

var secretKey []byte

func init() {
	_ = godotenv.Load()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is required")
	}
	secretKey = []byte(secret)
}

func CreateToken(username, userID string) (string, error) {
	// generate a hash or something using password and username  make a uuid to send in jwt
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"id":       userID,
			"exp":      time.Now().Add(10 * time.Minute).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func IsValidToken(token *jwt.Token) (interface{}, error) {
	// how do you validate a token in go
	// you check its expiry -  cannot check it since there is not storing happening right now.
	// you check it parsed that is username or something in db or not
	// so right now we just parse it and return add a check here with if username or something hardcoded.
	{
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("there was an error in parsing")
		}
		return secretKey, nil
	}
}
