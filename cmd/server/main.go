package main

import (
	"log"

	"github.com/injunweb/backend-server/internal/api"
	"github.com/injunweb/backend-server/internal/config"
	"github.com/injunweb/backend-server/pkg/database"
	"github.com/injunweb/backend-server/pkg/kubernetes"
	"github.com/injunweb/backend-server/pkg/vault"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()

	err := kubernetes.Init()
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes: %v", err)
	}

	err = database.Init()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	err = vault.Init()
	if err != nil {
		log.Fatalf("Failed to initialize Vault: %v", err)
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://dashboard.injunweb.com",
			"http://localhost:5173",
			"http://localhost:5500",
		},
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	api.SetupRoutes(router)

	if err := router.Run(":" + config.AppConfig.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
