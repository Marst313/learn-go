package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Marst/reminder-app/internal/cron"
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

	// ! 6. SETUP CRON SCHEDULER
	scheduler := cron.NewScheduler()

	err = scheduler.RegisterJobs()
	if err != nil {
		log.Fatal("Failed to register cron jobs : ", err)
	}
	scheduler.Start()

	// ! 7. SETUP GIN ROUTER
	router := gin.Default()
	router.SetTrustedProxies([]string{"http://localhost:5173"})

	// ! 8. CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// ! 9. Router
	routes.RegisterRoutes(router)

	// ! 10. SETUP HTTP SERVER
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	go func() {
		log.Printf("🚀 Server running on http://localhost:%s", port)

		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// ! 11. GRACEFUL SHUTDOWN
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		log.Fatal("Server forced to shudown:", err)
	}
	log.Println("Server exited properly")

}
