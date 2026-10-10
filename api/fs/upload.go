package fs

import (
	"errors"
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
		gecho.BadRequest(w,
			gecho.WithMessage("error.file.parseFailed"),
			gecho.Send(),
		)
		return
	}
	// Remove any temp files the parser spilled to disk
	defer r.MultipartForm.RemoveAll()

	file, handler, err := r.FormFile("file")
	if err != nil {
		gecho.BadRequest(w,
			gecho.WithMessage("error.file.retrieveFailed"),
			gecho.Send(),
		)
		return
	}
	defer file.Close()

	// SaveUpload validates, stores the original, and generates the WebP variants
	name, err := frm.fileService.SaveUpload(handler)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrTooLarge):
			gecho.BadRequest(w, gecho.WithMessage("error.file.tooLarge"), gecho.Send())
		case errors.Is(err, services.ErrUnsupportedType):
			gecho.BadRequest(w, gecho.WithMessage("error.file.unsupportedType"), gecho.Send())
		case errors.Is(err, services.ErrBadImage):
			gecho.BadRequest(w, gecho.WithMessage("error.file.invalidImage"), gecho.Send())
		default:
			frm.logger.Error("Failed to save upload", gecho.Field("error", err))
			gecho.InternalServerError(w,
				gecho.WithMessage("error.file.processFailed"),
				gecho.Send(),
			)
		}
		return
	}

	gecho.Success(w,
		gecho.WithMessage("success.file.uploaded"),
		gecho.WithData(map[string]any{
			"name":   name,
			"url":    frm.fileService.URL(name, 800),
			"srcset": frm.fileService.SrcSet(name),
		}),
		gecho.Send(),
	)
}
