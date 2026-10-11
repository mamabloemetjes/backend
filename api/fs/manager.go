package fs

import (
	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"

	"github.com/mamabloemetjes/backend/api/middleware"
	"github.com/mamabloemetjes/backend/services"
	"github.com/mamabloemetjes/backend/structs"
)

type FileRoutesManager struct {
	logger      *gecho.Logger
	cfg         *structs.Config
	fileService *services.FileService
	mw          *middleware.Middleware
}

func NewFileRoutesManager(
	logger *gecho.Logger,
	fileService *services.FileService,
	cfg *structs.Config,
	mw *middleware.Middleware,
) *FileRoutesManager {
	return &FileRoutesManager{
		logger:      logger,
		fileService: fileService,
		cfg:         cfg,
		mw:          mw,
	}
}

func (frm *FileRoutesManager) RegisterRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(frm.mw.UserAuthMiddleware)
		r.Use(frm.mw.AdminAuthMiddleware)
		r.Use(frm.mw.CSRFMiddleware())
		r.Post("/upload", frm.UploadFile)
	})
	r.Get("/uploads/{filename}", frm.fileService.Handler().ServeHTTP)
}
