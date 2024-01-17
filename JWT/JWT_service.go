package jwt_service

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func IsTokenValid(tokenString string, secretKey []byte) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unsupported signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		fmt.Println("Error parsing token:", err)
		return false
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
		if expirationTime.Before(time.Now()) {
			fmt.Println("Token has expired")
			return false
		}
		return true
	} else {
		fmt.Println("Invalid token")
		return false
	}
}
func GenerateJWT(userID, username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	sampleSecretKey := []byte("test")
	claims := token.Claims.(jwt.MapClaims)
	claims["ID"] = userID
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()
	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func ParseJWT(tokenString string) (string, error) {
	sampleSecretKey := []byte("test")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unsupported signing method: %v", token.Header["alg"])
		}
		return sampleSecretKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["ID"].(string)
		return userID, nil
	} else {
		return "", fmt.Errorf("invalid token")
	}
}
