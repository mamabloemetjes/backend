package fs

import (
	"mamabloemetjes_server/services"
	"mamabloemetjes_server/structs"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

type FileRoutesManager struct {
	logger      *gecho.Logger
	cfg         *structs.Config
	fileService *services.FileService
}

func NewFileRoutesManager(
	logger *gecho.Logger,
	fileService *services.FileService,
	cfg *structs.Config,
) *FileRoutesManager {
	return &FileRoutesManager{
		logger:      logger,
		fileService: fileService,
		cfg:         cfg,
	}
}

func (frm *FileRoutesManager) RegisterRoutes(r chi.Router) {
	r.Post("/upload", frm.UploadFile)
	r.Handle("/uploads/*", frm.fileService.Handler())
}
