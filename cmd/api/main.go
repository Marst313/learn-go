package main

import (
	"log"
	"os"

	"github.com/Marst/reminder-app/internal/database"
	"github.com/Marst/reminder-app/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// ! 1. LOAD ENV
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found using environtment variables")
	}

	// ! 2. GET DATABASE URL AND PORT
	databaseURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("🚀 Starting server...")
	log.Println("   Port:", port)
	log.Println("   DB_URL:", databaseURL)

	// ! 3. Connect to database
	err = database.Connect(databaseURL)
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// ! 4. CLOSE DATABASE IN LAST bcz defer
	defer database.Close()

	// ! 5. Migrations
	err = database.RunMigrations(databaseURL, "cmd/migrate/migrations")
	if err != nil {
		log.Fatal("Migration failed", err)
	}

	router := gin.Default()
	router.SetTrustedProxies([]string{"http://localhost:5173"})

	// ! 6. CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// ! 7. Router
	routes.RegisterRoutes(router)

	log.Printf("🚀 Server running on http://localhost:%s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}

}
