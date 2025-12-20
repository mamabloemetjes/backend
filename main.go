package main

import (
	"context"
	"errors"
	"fmt"
	"mamabloemetjes_server/api"
	"mamabloemetjes_server/api/admin"
	"mamabloemetjes_server/api/auth"
	"mamabloemetjes_server/api/health"
	"mamabloemetjes_server/api/middleware"
	"mamabloemetjes_server/api/products"
	"mamabloemetjes_server/config"
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/services"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MonkyMars/gecho"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		// Not fatal, just log a warning
		fmt.Println("Warning: could not load .env file")
	}

	// Load configuration
	cfg := config.GetConfig()

	// Initialize logger
	logLevel := gecho.ParseLogLevel(cfg.Server.LogLevel)
	logger := gecho.NewLogger(gecho.NewConfig(gecho.WithShowCaller(true), gecho.WithLogLevel(logLevel)))
	mwLogger := gecho.NewLogger(gecho.NewConfig(gecho.WithShowCaller(false), gecho.WithLogLevel(logLevel)))
	logger.Info("Logger initialized")

	// Initialize database
	db, err := database.Connect(logger)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Info("Database connected successfully")

	// Initialize services
	serviceManager := services.NewServiceManager(logger, cfg, db)
	if err := serviceManager.CacheService.Ping(); err != nil {
		logger.Error("Cache service ping failed", gecho.Field("error", err))
	} else {
		logger.Info("Cache service connected successfully")
	}

	// Initialize middleware
	mw := middleware.NewMiddleware(cfg, mwLogger, db)

	// Initialize route managers
	healthRoutes := health.NewHealthRoutesManager(serviceManager.HealthService)
	productRoutes := products.NewProductRoutesManager(logger, serviceManager.ProductService)
	authRoutes := auth.NewAuthRoutesManager(logger, serviceManager.AuthService, serviceManager.EmailService, serviceManager.CacheService, cfg, mw)
	adminRoutes := admin.NewAdminRoutesManager(logger, serviceManager.ProductService, mw)

	// Initialize main router manager
	routerManager := api.NewRouterManager(productRoutes, healthRoutes, authRoutes, adminRoutes)

	// Setup router
	r := api.App(routerManager, mw, cfg)

	// Setup server
	server := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: r,
	}

	// Graceful shutdown context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process interruption.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Fatal("server shutdown failed", gecho.Field("error", err))
		}
		serverStopCtx()
	}()

	logger.Info(fmt.Sprintf("Starting server (%s) on %s", cfg.Server.AppName, cfg.Server.Port))
	// Run the server
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()

	// Close resources
	var wg sync.WaitGroup
	var closeErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Closing database connection")
		if err := db.Close(); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("failed to close database: %w", err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Closing cache connection")
		if err := serviceManager.CacheService.Close(); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("failed to close cache: %w", err))
		}
	}()

	wg.Wait()

	logger.Info("Shutdown complete")
	return closeErr
}
