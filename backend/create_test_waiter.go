package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/username/kafe-backend/internal/database"
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
	"github.com/username/kafe-backend/internal/service"
)

func main() {
	_ = godotenv.Load()
	if err := database.InitDB(); err != nil {
		log.Fatalf("DB Error: %v", err)
	}
	defer database.DB.Close()

	userRepo := repository.NewUserRepository(database.DB)
	authService := service.NewAuthService(userRepo)

	phone := "+998901234567"
	password := "123456"
	
	// Create waiter
	user, _, err := authService.Register("Test Ofitsant", phone, password, models.RoleWaiter)
	if err != nil {
		log.Fatalf("Error registering: %v", err)
	}
	
	fmt.Printf("\nSUCCESS! Test waiter created.\nLogin: %s\nPassword: %s\nRole: %s\n", user.Phone, password, user.Role)
}
