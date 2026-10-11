package fs

import (
	"errors"
	"mamabloemetjes_server/api/response"
	"net/http"

	"github.com/MonkyMars/gecho"

	"mamabloemetjes_server/services"
)

func (frm *FileRoutesManager) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Hard cap on the whole request body. ParseMultipartForm's argument only
	// limits memory use, not the total size, so without this a client can
	// stream unlimited data to disk. The extra 1MB covers multipart overhead.
	r.Body = http.MaxBytesReader(w, r.Body, frm.cfg.FileStorage.MaxUploadBytes+(1<<20))

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.Send(gecho.BadRequest(w,
			gecho.WithMessage("error.file.parseFailed"),
		))
		return
	}
	// Remove any temp files the parser spilled to disk
	defer r.MultipartForm.RemoveAll()

	file, handler, err := r.FormFile("file")
	if err != nil {
		response.Send(gecho.BadRequest(w,
			gecho.WithMessage("error.file.retrieveFailed"),
		))
		return
	}
	defer file.Close()

	// SaveUpload validates, stores the original, and generates the WebP variants
	name, err := frm.fileService.SaveUpload(handler)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrTooLarge):
			response.Send(gecho.BadRequest(w, gecho.WithMessage("error.file.tooLarge")))
		case errors.Is(err, services.ErrUnsupportedType):
			response.Send(gecho.BadRequest(w, gecho.WithMessage("error.file.unsupportedType")))
		case errors.Is(err, services.ErrBadImage):
			response.Send(gecho.BadRequest(w, gecho.WithMessage("error.file.invalidImage")))
		default:
			frm.logger.Error("Failed to save upload", gecho.Field("error", err))
			response.Send(gecho.InternalServerError(w,
				gecho.WithMessage("error.file.processFailed"),
			))
		}
		return
	}
	response.Send(gecho.Success(w,
		gecho.WithMessage("success.file.uploaded"),
		gecho.WithData(map[string]any{
			"name":   name,
			"srcset": frm.fileService.SrcSet(name),
		}),
	))
}
