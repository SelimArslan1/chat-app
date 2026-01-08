package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/SelimArslan1/chat-app/internal/db"
	"github.com/SelimArslan1/chat-app/internal/models"
)

func main() {
	_ = godotenv.Load()

	database := db.Connect()
	log.Println("DB connected:", database != nil)

	err := database.AutoMigrate(

		&models.User{},

		&models.Server{},

		&models.Channel{},

		&models.Message{},
	)

	if err != nil {

		log.Fatal("auto-migrate failed:", err)

	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
