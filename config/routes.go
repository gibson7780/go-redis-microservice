package config

import (
	"fmt"

	// "github.com/gibson7780/go-project/common/utils/cache"
	"github.com/gibson7780/go-project/internal/stats"
	"github.com/gibson7780/go-project/internal/urls"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

/**
* Sets up API prefix route and all routers.
**/
func SetupRouter(db *sqlx.DB, redisClient redis.UniversalClient, urlsHandler *urls.Handler, statsHandler *stats.Handler) *gin.Engine {
	router := gin.Default()

	// NOTE: debugging middleware
	router.Use(func(c *gin.Context) {
		fmt.Println("Incoming request to:", c.Request.Method, c.Request.URL.Path, "from", c.Request.Host)
		c.Next()
	})

	// TODO: CORS for development, remove in PROD
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// base route
	api := router.Group("/api")

	// repo := stats.NewRepository(db)
	// statsService := stats.NewService(repo)
	// statsHandler := stats.NewHandler(statsService)

	statsRoutes := api.Group("/stats")
	statsRoutes.GET("/:code", statsHandler.GetStat)
	// statsRoutes.POST("", statsHandler.CreateStat)

	// urlRepo := urls.NewRepository(db)
	// urlService := urls.NewService(db, redisClient, urlRepo, statsService)
	// urlsHandler := urls.NewHandler(urlService)

	urlsRoutes := api.Group("/urls")
	router.GET("/:code", urlsHandler.GetUrl)
	urlsRoutes.POST("/", urlsHandler.CreateUrl)
	return router
}
