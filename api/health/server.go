package health

import (
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (hrm *HealthRoutesManager) GetServerHealth(w http.ResponseWriter, r *http.Request) {
	healthStatus := hrm.healthService.GetServerHealthStatus()
	gecho.Success(w,
		gecho.WithData(healthStatus),
		gecho.Send(),
	)
}

func (hrm *HealthRoutesManager) GetDatabaseHealth(w http.ResponseWriter, r *http.Request) {
	dbHealthStatus, err := hrm.healthService.GetDatabaseHealthStatus()
	if err != nil {
		gecho.InternalServerError(w,
			gecho.WithMessage("Database health check failed"),
			gecho.Send(),
		)
		return
	}
	gecho.Success(w,
		gecho.WithData(dbHealthStatus),
		gecho.Send(),
	)
}
