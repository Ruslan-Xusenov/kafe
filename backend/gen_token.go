package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env to get the real secret
	godotenv.Load()
	secretStr := os.Getenv("JWT_SECRET")
	if secretStr == "" {
		secretStr = "your_super_secret_jwt_key_here" // Fallback
	}
	secret := []byte(secretStr)
	
	// Create a token that lasts for 10 YEARS
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 999,      // System User ID for Printer
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
