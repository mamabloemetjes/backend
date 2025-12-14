package handling

import (
	"net/http"

	"github.com/MonkyMars/gecho"
)

func HandleError(err error, msg string, logger *gecho.Logger, w http.ResponseWriter) error {
	logger.Error("An error occurred", gecho.Field("error", err), gecho.Field("msg", msg), gecho.WithCallerSkip(3))

	return gecho.InternalServerError(w, gecho.Send())
}
