package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()
	secretStr := os.Getenv("JWT_SECRET")
	if secretStr == "" {
		secretStr = "your_super_secret_jwt_key_here"
	}
	secret := []byte(secretStr)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 999,
		"role":    "printer",
		"exp":     time.Now().Add(time.Hour * 24 * 365 * 10).Unix(),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		fmt.Printf("Error signing token: %v\n", err)
		return
	}
	fmt.Println("--- PRINTER ACCESS TOKEN (10 YEARS) ---")
	fmt.Println(tokenString)
	fmt.Println("---------------------------------------")
}