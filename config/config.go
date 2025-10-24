// Package config handles application configuration
package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config holds all configuration for the application
type Config struct {
	MongoURI          string
	DatabaseName      string
	JWTSecret         string
	ServerPort        string
	ExternalCourseAPI string
	LogDir            string
}

var AppConfig *Config

// LoadConfig loads configuration from environment variables
func LoadConfig() {
	// Load .env file if it exists
	_ = godotenv.Load()

	AppConfig = &Config{
		MongoURI:          getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:      getEnv("DATABASE_NAME", "quizmaster"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		ExternalCourseAPI: getEnv("EXTERNAL_COURSE_API", "http://localhost:9000/api/v1"),
		LogDir:            getEnv("LOG_DIR", "var/logs"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Database connection
var DB *mongo.Database

// ConnectDatabase establishes connection to MongoDB
func ConnectDatabase() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(AppConfig.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	DB = client.Database(AppConfig.DatabaseName)
	log.Println("Successfully connected to MongoDB")

	return nil
}

// GetCollection returns a MongoDB collection
func GetCollection(collectionName string) *mongo.Collection {
	return DB.Collection(collectionName)
}

// SetupLogger sets up logging to a daily rotating file
func SetupLogger() error {
	// Ensure log directory exists
	if err := os.MkdirAll(AppConfig.LogDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with current date
	logFileName := filepath.Join(AppConfig.LogDir, time.Now().Format("2006-01-02")+".log")
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Set log output to file and also to stdout
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return nil
}
