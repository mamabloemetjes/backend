package response

import (
	"log/slog"

	"github.com/MonkyMars/gecho"
)

// Send writes a gecho v0.7 response and centrally reports serialization or
// transport failures. Response constructors only build responses; Send is the
// operation that performs I/O.
func Send(resp *gecho.Response) {
	if err := resp.Send(); err != nil {
		slog.Error("failed to send HTTP response", "error", err)
	}
}
