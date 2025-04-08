package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func MongoConnect() *mongo.Client {
	if mongoClient != nil {
		return mongoClient // Gunakan koneksi yang sudah ada
	}

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: No .env file found. Using system environment variables.")
	}

	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")

	if mongoURI == "" || dbName == "" {
		panic("MongoDB configuration is missing in environment variables")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().
		ApplyURI(mongoURI).
		SetServerAPIOptions(serverAPI).
		SetMaxPoolSize(20).
		SetMinPoolSize(5).
		SetMaxConnecting(10).
		SetConnectTimeout(10 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		panic(err)
	}

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	mongoClient = client
	return client
}

func MongoCollection(coll string) *mongo.Collection {
	client := MongoConnect() // Gunakan koneksi yang sudah ada
	mongoDBName := os.Getenv("MONGO_DB_NAME")
	return client.Database(mongoDBName).Collection(coll)
}

func init() {
	_ = godotenv.Load()
}
