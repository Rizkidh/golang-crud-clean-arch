package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoConnect() *mongo.Client {
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")

	if mongoURI == "" || dbName == "" {
		panic("❌ MongoDB configuration is missing in .env")
	}

	opts := options.Client().
		ApplyURI(mongoURI).
		SetMaxPoolSize(20).
		SetMinPoolSize(5).
		SetConnectTimeout(5 * time.Second)

	var client *mongo.Client
	var err error

	for i := 1; i <= 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, err = mongo.Connect(ctx, opts)
		if err == nil {
			// Coba ping
			if err = client.Database("admin").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err(); err == nil {
				fmt.Println("✅ MongoDB connection established and ping successful")
				return client
			}
		}

		fmt.Printf("⏳ Retry MongoDB connection (%d/5)...\n", i)
		time.Sleep(2 * time.Second)
	}

	panic(fmt.Sprintf("❌ MongoDB connection failed after retries: %v", err))
}

// PostgresConnect establishes a connection to PostgreSQL
func PostgresConnect() (*sql.DB, error) {
	host := os.Getenv("PG_HOST")
	port := os.Getenv("PG_PORT")
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	dbname := os.Getenv("PG_DB_NAME")
	sslmode := os.Getenv("PG_SSL_MODE")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to open PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("❌ Failed to ping PostgreSQL: %w", err)
	}

	fmt.Println("✅ Successfully connected to PostgreSQL!")
	return db, nil
}
