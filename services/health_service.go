package services

import (
	"context"
	"mamabloemetjes_server/database"
	"time"

	"github.com/MonkyMars/gecho"
)

var uptimeStart time.Time

func init() {
	uptimeStart = time.Now()
}

type serverHealthStatus struct {
	Uptime       float64   `json:"uptime"`        // in seconds
	CurrentTime  time.Time `json:"current_time"`  // server current time
	ServiceAlive bool      `json:"service_alive"` // always true if service is running
}

type databaseHealthStatus struct {
	Connected      bool      `json:"connected"`
	LastChecked    time.Time `json:"last_checked"`
	ResponseTimeMs int64     `json:"response_time_ms"`
}

type HealthService struct {
	logger *gecho.Logger
	db     *database.DB
	status serverHealthStatus
}

func NewHealthService(logger *gecho.Logger, db *database.DB) *HealthService {
	return &HealthService{
		logger: logger,
		db:     db,
		status: serverHealthStatus{
			Uptime:       0,
			CurrentTime:  time.Now(),
			ServiceAlive: true,
		},
	}
}

func (hs *HealthService) GetServerHealthStatus() serverHealthStatus {
	hs.status.Uptime = time.Since(uptimeStart).Seconds()
	hs.status.CurrentTime = time.Now()
	return hs.status
}

func (hs *HealthService) GetDatabaseHealthStatus() (databaseHealthStatus, error) {
	start := time.Now()
	err := hs.db.Ping(context.Background())
	elapsed := time.Since(start).Milliseconds()

	dbStatus := databaseHealthStatus{
		Connected:      err == nil,
		LastChecked:    time.Now(),
		ResponseTimeMs: elapsed,
	}

	if err != nil {
		hs.logger.Error("Database health check failed: ", err)
	}

	// You can return or log dbStatus as needed
	return dbStatus, err
}
