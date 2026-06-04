package config

import (
	"fmt"

	commonhelpers "github.com/gibson7780/go-project/common/utils"

	// "github.com/gibson7780/go-project/common/utils/cache"
	"github.com/gibson7780/go-project/internal/auth"
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

	// auth (initialized here since no worker dependency)
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(
		authRepo,
		redisClient,
		commonhelpers.GetEnvString("JWT_SECRET", "change-me"),
	)
	authHandler := auth.NewHandler(authService)

	// requireAuth guards routes that need a valid Bearer access token
	requireAuth := auth.RequireAuth(authService)

	// repo := stats.NewRepository(db)
	// statsService := stats.NewService(repo)
	// statsHandler := stats.NewHandler(statsService)

	statsRoutes := api.Group("/stats")
	statsRoutes.Use(requireAuth)
	statsRoutes.GET("/:code", statsHandler.GetStat)
	// statsRoutes.POST("", statsHandler.CreateStat)

	// urlRepo := urls.NewRepository(db)
	// urlService := urls.NewService(db, redisClient, urlRepo, statsService)
	// urlsHandler := urls.NewHandler(urlService)

	router.GET("/:code", urlsHandler.GetUrl) // public short-link redirect

	urlsRoutes := api.Group("/urls")
	urlsRoutes.Use(requireAuth)
	urlsRoutes.POST("/", urlsHandler.CreateUrl)
	urlsRoutes.DELETE("/:code", urlsHandler.DeleteUrl)

	authRoutes := api.Group("/auth")
	authRoutes.POST("/signup", authHandler.Signup)
	authRoutes.POST("/signin", authHandler.Signin)
	authRoutes.POST("/signout", authHandler.Signout)

	return router
}
