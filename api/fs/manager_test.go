package fs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mamabloemetjes/backend/services"
	"github.com/mamabloemetjes/backend/structs"
)

func TestRegisterRoutesServesUploadByFilename(t *testing.T) {
	uploadDir := t.TempDir()
	filename := "image-800.webp"
	content := []byte("webp")
	if err := os.WriteFile(filepath.Join(uploadDir, filename), content, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &structs.Config{
		FileStorage: &structs.FileStorageConfig{UploadDir: uploadDir},
	}
	fileService := services.NewFileService(nil, cfg)
	router := chi.NewRouter()
	router.Get("/uploads/{filename}", fileService.Handler().ServeHTTP)

	req := httptest.NewRequest(http.MethodGet, "/uploads/"+filename, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Body.String() != string(content) {
		t.Fatalf("expected body %q, got %q", content, recorder.Body.String())
	}
}
