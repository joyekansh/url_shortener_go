package database

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/redis/go-redis/v9"
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
