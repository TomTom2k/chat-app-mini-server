package mongodb

import (
	"context"
	"log"
	"time"

	"github.com/TomTom2k/chat-app/server/internal/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	Client     *mongo.Client
	Database   *mongo.Database
	initialized bool
)

// Initialize connects to MongoDB and initializes the database
func Initialize(cfg *config.Config) error {
	if initialized {
		return nil
	}

	clientOptions := options.Client().ApplyURI(cfg.MongoDBURI)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return err
	}

	// Ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	Client = client
	Database = client.Database(cfg.DatabaseName)
	initialized = true

	log.Println("[INFO]: MongoDB connected successfully")
	return nil
}

// OpenCollection returns a collection from the database
func OpenCollection(collectionName string) *mongo.Collection {
	if Database == nil {
		log.Fatal("[ERROR]: Database not initialized. Call mongodb.Initialize() first")
	}
	return Database.Collection(collectionName)
}

// Close closes the MongoDB connection
func Close() error {
	if Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return Client.Disconnect(ctx)
}
