package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/Emmanuella-codes/burnished-microservice/internal/api"
	"github.com/Emmanuella-codes/burnished-microservice/internal/config"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	server := api.NewServer(cfg)
	log.Printf("Starting server on port %s...", cfg.Port)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}