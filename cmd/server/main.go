package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	commonhelpers "github.com/gibson7780/go-project/common/utils"
	"github.com/gibson7780/go-project/common/utils/cache"
	"github.com/gibson7780/go-project/config"
	"github.com/gibson7780/go-project/internal/jobs"
	"github.com/gibson7780/go-project/internal/stats"
	"github.com/gibson7780/go-project/internal/urls"
	"github.com/gibson7780/go-project/internal/worker"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

var (
	// grpc
	// serviceName = "go-project"
	httpAddr = commonhelpers.GetEnvString("PORT", "7001")
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
	redisClient := config.GetClient()
	_ = cache.NewRedisCache(redisClient)

	repo := stats.NewRepository(db)
	statsService := stats.NewService(repo)
	statsHandler := stats.NewHandler(statsService)

	urlRepo := urls.NewRepository(db)
	urlService := urls.NewService(db, redisClient, urlRepo, statsService)
	urlsHandler := urls.NewHandler(urlService)

	jobsRepository := jobs.NewRepository(db)
	jobsService := jobs.NewService(jobsRepository, redisClient)
	jobsHandler := jobs.NewHandler(jobsService)

	ctx, cancel := context.WithCancel(context.Background())

	w := worker.NewWorker(ctx, redisClient, urlService, statsService)
	defer cancel()
	// --- router setup ---
	router := config.SetupRouter(db, redisClient, urlsHandler, statsHandler, jobsHandler)

	// -- start server --
	// if err := router.Run(fmt.Sprintf(":%s", httpAddr)); err != nil {
	// 	log.Fatal("Failed to start server")
	// }

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpAddr),
		Handler: router, // 把 gin router 當作 handler 傳進去
	}
	log.Printf("🚀 Server started successfully, listening on http://localhost:%s", httpAddr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server")
		}
	}()

	// 等待關閉信號
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	w.Wait()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
