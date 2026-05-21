package config

import (
	"fmt"

	"github.com/gibson7780/go-project/common/utils/cache"
	"github.com/gibson7780/go-project/internal/stats"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

/**
* Sets up API prefix route and all routers.
**/
func SetupRouter(db *sqlx.DB, cacheService cache.Cache) *gin.Engine {
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

	/***************
	* MICROSERVICES
	***************/

	// --- EXAMPLE MICROSERVICE ---

	repo := stats.NewRepository(db)
	service := stats.NewService(repo)
	statsHandler := stats.NewHandler(service)

	statsRoutes := api.Group("/stats")
	statsRoutes.GET("/:id", statsHandler.GetStats)
	statsRoutes.POST("", statsHandler.CreateStats)
	return router
}
