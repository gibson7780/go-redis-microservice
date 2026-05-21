package main

import (
	"fmt"
	"log"

	commonhelpers "github.com/gibson7780/go-project/common/utils"
	"github.com/gibson7780/go-project/common/utils/cache"
	"github.com/gibson7780/go-project/config"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

var (
	// grpc
	serviceName = "examples"
	httpAddr    = commonhelpers.GetEnvString("PORT", "7001")
)

func main() {
	// --- database setup ---
	db := config.InitDB()
	defer db.Close()

	// --- redis setup ---
	err := config.InitRedis(config.RedisConfig{
		Mode:         commonhelpers.GetEnvString("REDIS_MODE", "standalone"),
		Addrs:        []string{commonhelpers.GetEnvString("REDIS_ADDR", "localhost:6379")},
		Password:     commonhelpers.GetEnvString("REDIS_PASSWORD", ""),
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer config.CloseRedis()
	cacheService := cache.NewRedisCache(config.GetClient())

	// --- router setup ---
	router := config.SetupRouter(db, cacheService)

	// -- start server --
	if err := router.Run(fmt.Sprintf(":%s", httpAddr)); err != nil {
		log.Fatal("Failed to start server")
	}

}
