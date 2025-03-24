package main

import (
	"context"
	"errors"
	"fmt"
	"golang-crud-clean-arch/m/db"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	/// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Menyajikan file statis di folder "views"
	e.Static("/static", "views")

	// Route untuk menampilkan index.html sebagai halaman utama
	e.GET("/", func(c echo.Context) error {
		return c.File("views/index.html")
	})

	// Tambahkan route untuk halaman repository
	e.GET("/repositories", func(c echo.Context) error {
		return c.File("views/repository.html")
	})

	// Routes
	e.POST("/create-user", CreateUser) // Pastikan ini ada di main.go
	e.GET("/users", GetUsers)
	e.PUT("/users/:id", UpdateUser)
	e.GET("/create-repo", CreateRepo)
	e.GET("/repo", GetRepos)
	e.DELETE("/users/:id", DeleteUser)

	// Start server
	if err := e.Start(":9000"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repository struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"` // Relasi ke User
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	AIEnabled bool      `json:"ai_enabled"` // Apakah terhubung ke AI Code Review
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Handler Create Users (POST request)
func CreateUser(c echo.Context) error {
	client := db.MongoConnect()
	defer client.Disconnect(context.TODO())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	table := db.MongoCollection("users", client)

	var user User
	if err := c.Bind(&user); err != nil {
		fmt.Println("Error binding data:", err)
		return c.String(http.StatusBadRequest, "Invalid request format")
	}

	// Cari ID terbesar saat ini
	var lastUser User
	opts := options.FindOne().SetSort(bson.M{"id": -1}) // Urutkan descending
	err := table.FindOne(ctx, bson.M{}, opts).Decode(&lastUser)

	if err != nil {
		// Jika tidak ada user, mulai dari ID 1
		user.ID = 1
	} else {
		// Jika ada user, tambahkan 1 ke ID terakhir
		user.ID = lastUser.ID + 1
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err = table.InsertOne(ctx, user)
	if err != nil {
		fmt.Println("Error inserting user:", err)
		return c.String(http.StatusInternalServerError, "Failed to insert user")
	}

	return c.JSON(http.StatusOK, user)
}

// Handler Update User
func UpdateUser(c echo.Context) error {
	client := db.MongoConnect()
	defer client.Disconnect(context.TODO())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	table := db.MongoCollection("users", client)

	// Ambil ID dari parameter URL dan konversi ke integer
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid user ID format")
	}

	// Ambil data dari request body
	var updatedUser User
	if err := c.Bind(&updatedUser); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request format")
	}

	// Perbarui data user
	update := bson.M{
		"$set": bson.M{
			"name":       updatedUser.Name,
			"email":      updatedUser.Email,
			"updated_at": time.Now(),
		},
	}

	result, err := table.UpdateOne(ctx, bson.M{"id": id}, update)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to update user")
	}

	if result.MatchedCount == 0 {
		return c.String(http.StatusNotFound, "User not found")
	}

	return c.JSON(http.StatusOK, bson.M{"message": "User updated successfully"})
}

// Handler Delete User
func DeleteUser(c echo.Context) error {
	client := db.MongoConnect()
	defer client.Disconnect(context.TODO()) // Pastikan koneksi ditutup setelah selesai

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	table := db.MongoCollection("users", client)

	// Ambil ID dari parameter URL dan ubah ke integer
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr) // Konversi string ke integer
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid user ID format")
	}

	// Hapus user berdasarkan ID
	result, err := table.DeleteOne(ctx, bson.M{"id": id}) // Cari dengan integer
	if err != nil {
		fmt.Println("Error deleting user:", err)
		return c.String(http.StatusInternalServerError, "Failed to delete user")
	}

	// Cek apakah ada dokumen yang terhapus
	if result.DeletedCount == 0 {
		return c.String(http.StatusNotFound, "User not found")
	}

	return c.String(http.StatusOK, "User deleted successfully")
}

// Handler Show Users
func GetUsers(c echo.Context) error {
	client := db.MongoConnect()
	defer client.Disconnect(context.TODO()) // Pastikan koneksi ditutup setelah selesai

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	table := db.MongoCollection("users", client)

	// Query semua data user
	cursor, err := table.Find(ctx, bson.M{})
	if err != nil {
		fmt.Println("Error fetching users:", err)
		return c.String(http.StatusInternalServerError, "Failed to fetch users")
	}
	defer cursor.Close(ctx)

	var users []User
	if err := cursor.All(ctx, &users); err != nil {
		fmt.Println("Error decoding users:", err)
		return c.String(http.StatusInternalServerError, "Failed to decode users")
	}

	return c.JSON(http.StatusOK, users)
}

// Handler Create Repository
func CreateRepo(c echo.Context) error {
	client := db.MongoConnect()
	defer client.Disconnect(context.TODO())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ambil data dari request body
	var repo Repository
	if err := c.Bind(&repo); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request format")
	}

	// Cek apakah user dengan UserID ada di database
	userTable := db.MongoCollection("users", client)
	var existingUser User
	err := userTable.FindOne(ctx, bson.M{"id": repo.UserID}).Decode(&existingUser)
	if err != nil {
		return c.String(http.StatusNotFound, "User not found")
	}

	// Simpan repository baru dengan referensi ke UserID
	repoTable := db.MongoCollection("repo", client)
	repo.CreatedAt = time.Now()
	repo.UpdatedAt = time.Now()

	_, err = repoTable.InsertOne(ctx, repo)
	if err != nil {
		fmt.Println("Error inserting repo:", err)
		return c.String(http.StatusInternalServerError, "Failed to insert repo")
	}

	return c.String(http.StatusOK, "Repository created successfully")
}

// Handler Show Repositories
func GetRepos(c echo.Context) error {
	client := db.MongoConnect()
	defer client.Disconnect(context.TODO()) // Tutup koneksi setelah selesai digunakan

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	table := db.MongoCollection("repo", client)

	// Query semua data repository
	cursor, err := table.Find(ctx, bson.M{})
	if err != nil {
		fmt.Println("Error fetching repositories:", err)
		return c.String(http.StatusInternalServerError, "Failed to fetch repositories")
	}
	defer cursor.Close(ctx)

	var repos []Repository
	if err := cursor.All(ctx, &repos); err != nil {
		fmt.Println("Error decoding repositories:", err)
		return c.String(http.StatusInternalServerError, "Failed to decode repositories")
	}

	return c.JSON(http.StatusOK, repos)
}
