package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	MongoDBURI    string
	DatabaseName  string
	JWTSecret     string
	Environment   string
}

func Load() *Config {
	// Load .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Println("[WARNING]: unable to load .env file, using environment variables")
	}

	config := &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		MongoDBURI:   getEnv("MONGODB_URI", ""),
		DatabaseName: getEnv("DATABASE_NAME", "chat_app"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Environment:  getEnv("ENVIRONMENT", "development"),
	}

	// Validate required configs
	if config.MongoDBURI == "" {
		log.Fatal("[ERROR]: MONGODB_URI is required")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}




