
package token

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func getJWTSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET environment variable is not set")
	}
	return []byte(secret), nil
}

func GenerateToken(userID string) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

func GetUserID(tokenString string) (string, error) {
	jwtToken, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	userID, ok := claims["userID"].(string)
	if !ok {
		return "", errors.New("userID not found in token")
	}

	return userID, nil
}
