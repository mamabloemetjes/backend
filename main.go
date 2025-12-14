package main

import (
	"fmt"
	"mamabloemetjes_server/api"
	"mamabloemetjes_server/config"
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/structs"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/MonkyMars/gecho"
	"github.com/joho/godotenv"
)

var logger *gecho.Logger
var cfg *structs.Config

func init() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found or error loading .env file, proceeding with system environment variables")
	}
	cfg = config.GetConfig()
	logger = config.InitializeLogger()
	err := database.Initialize()
	if err != nil {
		logger.Fatal("Failed to initialize database", gecho.Field("error", err))
	}
}

func main() {
	// Setup graceful shutdown BEFORE starting the server
	setupGracefulShutdown(logger)

	r := api.App()

	logger.Info(fmt.Sprintf("Starting server (%s) on %s", cfg.Server.AppName, cfg.Server.Port))

	// Start server
	if err := http.ListenAndServe(cfg.Server.Port, r); err != nil {
		logger.Error("Failed to start server", gecho.Field("error", err))
	}
}

// setupGracefulShutdown sets up signal handling for graceful application shutdown
func setupGracefulShutdown(logger *gecho.Logger) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Info("Graceful shutdown handler initialized")

	go func() {
		sig := <-c
		logger.Info("Received shutdown signal", "signal", sig)
		os.Exit(0)
	}()
}
