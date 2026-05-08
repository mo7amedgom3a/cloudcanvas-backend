package main

import (
	"cloudcanvas-backend/internal/handler"
	"cloudcanvas-backend/internal/repository"
	"cloudcanvas-backend/internal/service"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 1. Initialize Repository
	userRepo := repository.NewInMemoryUserRepository()

	// 2. Initialize Service
	userService := service.NewUserService(userRepo)

	// 3. Initialize Handler
	userHandler := handler.NewUserHandler(userService)

	// 4. Setup Routing
	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.HandleGetUser(w, r)
		case http.MethodPost:
			userHandler.HandleCreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// 5. Start Server
	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
