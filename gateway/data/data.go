package data

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var DB *sql.DB
var Cache *redis.Client

func Open() error {
	var database_url string = "host=/var/run/postgresql port=5432 dbname=kingdoms"
	var redis_address string = "localhost:6379"

	db, err := sql.Open("postgres", database_url)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return err
	}

	cache := redis.NewClient(&redis.Options{Addr: redis_address})

	if err := cache.Ping(context.Background()).Err(); err != nil {
		cache.Close()
		db.Close()
		return err
	}

	DB, Cache = db, cache
	return nil
}
func Close() {
	Cache.Close()
	DB.Close()
}
