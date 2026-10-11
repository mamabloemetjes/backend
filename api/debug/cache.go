package debug

import (
	"mamabloemetjes_server/api/response"
	"net/http"

	"github.com/MonkyMars/gecho"
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
