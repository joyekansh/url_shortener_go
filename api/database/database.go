package database

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joyekansh/url_shortener_go/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Ctx = context.Background()

// The function take paramater
// dbNo : Database number

func CreateClient(dbNo int) *redis.Client { // This function creates and returns a new redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("DB_ADDR"),
		Password: os.Getenv("DB_PASS"),
		DB:       dbNo,
	})

	return rdb // Gives us a redis client which can be connected to the database
}

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to PostgreSQL")
	}
	// Auto migrates the schema
	DB.AutoMigrate(&models.URL{})
}
