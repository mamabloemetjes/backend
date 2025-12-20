package database

import (
	"context"
	"database/sql"
	"fmt"
	"mamabloemetjes_server/config"
	"time"

	"github.com/MonkyMars/gecho"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// DB wraps the bun database connection with additional functionality
type DB struct {
	*bun.DB
	sqlDB  *sql.DB
	logger *gecho.Logger
}

// Connect establishes a connection to the database using centralized configuration
func Connect(logger *gecho.Logger) (*DB, error) {
	cfg := config.GetConfig()
	dbCfg := cfg.Database

	// Build DSN for pgdriver
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require",
		dbCfg.User,
		dbCfg.Password,
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.Name,
	)

	// Create pgdriver connector with connection pool settings
	connector := pgdriver.NewConnector(
		pgdriver.WithDSN(dsn),
		pgdriver.WithTimeout(30*time.Second),
		pgdriver.WithDialTimeout(10*time.Second),
		pgdriver.WithReadTimeout(dbCfg.ReadTimeout),
		pgdriver.WithWriteTimeout(dbCfg.WriteTimeout),
	)

	// Create SQL DB with connection pooling
	sqlDB := sql.OpenDB(connector)

	// Configure connection pool
	sqlDB.SetMaxOpenConns(dbCfg.MaxConns)
	sqlDB.SetMaxIdleConns(dbCfg.MinConns)
	sqlDB.SetConnMaxLifetime(dbCfg.MaxLifetime)
	sqlDB.SetConnMaxIdleTime(dbCfg.MaxIdleTime)

	// Create Bun DB with PostgreSQL dialect
	bunDB := bun.NewDB(sqlDB, pgdialect.New())

	// Add query hook for logging and monitoring
	bunDB.AddQueryHook(&queryHook{logger: logger})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := bunDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to database successfully",
		gecho.Field("host", dbCfg.Host),
		gecho.Field("database", dbCfg.Name),
		gecho.Field("max_conns", dbCfg.MaxConns),
	)

	return &DB{
		DB:     bunDB,
		sqlDB:  sqlDB,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if err := db.DB.Close(); err != nil {
		return err
	}
	return db.sqlDB.Close()
}

// Health checks the database connection health
func (db *DB) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}

// GetStats returns connection pool statistics for monitoring
func (db *DB) GetStats() sql.DBStats {
	return db.sqlDB.Stats()
}

// queryHook implements bun.QueryHook to monitor queries and handle errors
type queryHook struct {
	logger *gecho.Logger
}

func (h *queryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *queryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	duration := time.Since(event.StartTime)

	// Log slow queries (over 400ms)
	if duration > 400*time.Millisecond {
		h.logger.Warn("Slow database query detected",
			gecho.Field("query", event.Query),
			gecho.Field("duration", duration),
		)
	}

	// Log query errors with context
	if event.Err != nil {
		// Don't log "no rows" as an error (it's expected)
		if event.Err != sql.ErrNoRows {
			h.logger.Error("Database query error",
				gecho.Field("error", event.Err),
				gecho.Field("query", event.Query),
				gecho.Field("duration", duration),
			)
		}
	}
}
