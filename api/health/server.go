package health

import (
	"mamabloemetjes_server/api/response"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (hrm *HealthRoutesManager) GetServerHealth(w http.ResponseWriter, r *http.Request) {
	healthStatus := hrm.healthService.GetServerHealthStatus()
	response.Send(gecho.Success(w,
		gecho.WithData(healthStatus),
	))
}

func (hrm *HealthRoutesManager) GetDatabaseHealth(w http.ResponseWriter, r *http.Request) {
	dbHealthStatus, err := hrm.healthService.GetDatabaseHealthStatus()
	if err != nil {
		response.Send(gecho.InternalServerError(w,
			gecho.WithMessage("error.database.healthCheckFailed"),
		))
		return
	}
	response.Send(gecho.Success(w,
		gecho.WithData(dbHealthStatus),
	))
}
