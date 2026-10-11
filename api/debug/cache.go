package debug

import (
	"net/http"

	"github.com/MonkyMars/gecho"

	"github.com/mamabloemetjes/backend/api/response"
)

func (drm *DebugRoutesManager) ClearCache(w http.ResponseWriter, r *http.Request) {
	err := drm.cacheService.ClearAll()
	if err != nil {
		response.Send(gecho.InternalServerError(w,
			gecho.WithMessage("error.cache.clearFailed"),
		))
		return
	}
	response.Send(gecho.Success(w,
		gecho.WithMessage("success.cache.cleared"),
	))
}
