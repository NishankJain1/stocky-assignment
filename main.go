package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"stocky-assignment/api"
	"stocky-assignment/db"
)

var DB *sqlx.DB

func connectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ Error loading .env file")
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	DB, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}

	db.DB = DB // ✅ Share DB connection globally

	logrus.Infof("✅ Connected to PostgreSQL database: %s", os.Getenv("DB_NAME"))
}

func main() {
	connectDB()

	api.StartPriceUpdater()


	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Reward endpoint
	router.POST("/reward", api.AddReward)
	router.GET("/stats/:userId", api.GetStats)
	router.GET("/portfolio/:userId", api.GetPortfolio)
	router.GET("/historical-inr/:userId", api.GetHistoricalINR)
	router.GET("/today-stocks/:userId", api.GetTodayStocks)
	router.POST("/stock-adjustment", api.AddOrUpdateStockAdjustment)
	router.GET("/stock-adjustments", api.GetAllStockAdjustments)







	port := os.Getenv("PORT")
	logrus.Infof("🚀 Server running on port %s", port)
	router.Run(":" + port)
}
