package main

import (
	"fmt"
	"log"

	commonhelpers "github.com/gibson7780/go-project/common/utils"
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

	// --- router setup ---
	router := config.SetupRouter(db)

	// -- start server --
	if err := router.Run(fmt.Sprintf(":%s", httpAddr)); err != nil {
		log.Fatal("Failed to start server")
	}

}
