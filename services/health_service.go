package services

import (
	"context"
	"mamabloemetjes_server/database"
	"runtime"
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
	RamStats     *RamStats `json:"ram_stats"`
}

type RamStats struct {
	TotalMB     uint64 `json:"total_mb"`
	UsedMB      uint64 `json:"used_mb"`
	FreeMB      uint64 `json:"free_mb"`
	UsedPercent uint64 `json:"used_percent"`
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
			RamStats:     getRamStats(),
		},
	}
}

func getRamStats() *RamStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	totalMB := m.Sys / 1024 / 1024
	usedMB := m.Alloc / 1024 / 1024
	freeMB := totalMB - usedMB
	usedPercent := uint64(0)
	if totalMB > 0 {
		usedPercent = (usedMB * 100) / totalMB
	}

	return &RamStats{
		TotalMB:     totalMB,
		UsedMB:      usedMB,
		FreeMB:      freeMB,
		UsedPercent: usedPercent,
	}
}

func (hs *HealthService) GetServerHealthStatus() serverHealthStatus {
	hs.status.Uptime = time.Since(uptimeStart).Seconds()
	hs.status.CurrentTime = time.Now()
	hs.status.RamStats = getRamStats()
	return hs.status
}

func (hs *HealthService) GetDatabaseHealthStatus() (databaseHealthStatus, error) {
	start := time.Now()
	err := hs.db.PingContext(context.Background())
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
