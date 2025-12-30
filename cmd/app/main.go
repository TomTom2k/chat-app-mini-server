// @title           Chat App API
// @version         1.0
// @description     API documentation for Chat Application
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@chatapp.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"log"

	"github.com/TomTom2k/chat-app/server/docs"
	"github.com/TomTom2k/chat-app/server/internal/config"
	"github.com/TomTom2k/chat-app/server/internal/infrastructure/di"
	"github.com/TomTom2k/chat-app/server/internal/infrastructure/mongodb"
	"github.com/TomTom2k/chat-app/server/internal/infrastructure/server"
)

func main() {
	// Initialize Swagger docs (this calls init() in docs package)
	_ = docs.SwaggerInfo

	// Load configuration
	cfg := config.Load()

	// Initialize MongoDB connection
	if err := mongodb.Initialize(cfg); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongodb.Close()

	// Initialize dependency injection container
	container := di.NewContainer(cfg)

	// Initialize and start server
	srv := server.NewServer(cfg, container)

	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
