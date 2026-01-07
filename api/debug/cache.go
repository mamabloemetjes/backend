package debug

import (
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (drm *DebugRoutesManager) ClearCache(w http.ResponseWriter, r *http.Request) {
	err := drm.cacheService.ClearAll()
	if err != nil {
		gecho.InternalServerError(w,
			gecho.WithMessage("error.cache.clearFailed"),
			gecho.Send(),
		)
		return
	}

	gecho.Success(w,
		gecho.WithMessage("success.cache.cleared"),
		gecho.Send(),
	)
}
