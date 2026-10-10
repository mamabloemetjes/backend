package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"io"
	"mamabloemetjes_server/structs"
	"mamabloemetjes_server/structs/tables"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/MonkyMars/gecho"
	"github.com/disintegration/imaging"
)

type FileService struct {
	logger *gecho.Logger
	cfg    *structs.Config
}

var sizes = []int{400, 800, 1600}

var (
	ErrTooLarge        = errors.New("file too large")
	ErrUnsupportedType = errors.New("unsupported file type")
	ErrBadImage        = errors.New("invalid image")
	allowedTypes       = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
	}
)

func NewFileService(logger *gecho.Logger, cfg *structs.Config) *FileService {
	return &FileService{
		logger: logger,
		cfg:    cfg,
	}
}

func (fss *FileService) Init() error {
	for _, d := range []string{fss.cfg.FileStorage.UploadDir, fss.cfg.FileStorage.OriginalsDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", d, err)
		}
	}

	if _, err := exec.LookPath("cwebp"); err != nil {
		return fmt.Errorf("cwebp not found in PATH: %w", err)
	}

	return nil
}

// Resizes and converts to webp.
func (fss *FileService) ProcessImage(imageData []byte, outDir, name string) ([]string, error) {
	img, err := imaging.Decode(bytes.NewReader(imageData), imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	var outputs []string
	cleanup := func() {
		for _, o := range outputs {
			os.Remove(o)
		}
	}

	for _, w := range sizes {
		resized := img
		// Only shrink, never enlarge
		if img.Bounds().Dx() > w {
			resized = imaging.Resize(img, w, 0, imaging.Lanczos) // 0 = keep aspect ratio
		}

		// PNG is a lossless intermediate, so quality is only lost once, in the WebP step
		var buf bytes.Buffer
		if err := imaging.Encode(&buf, resized, imaging.PNG); err != nil {
			cleanup()
			return nil, fmt.Errorf("encode %dpx: %w", w, err)
		}

		out := filepath.Join(outDir, fmt.Sprintf("%s-%d.webp", name, w))

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		cmd := exec.CommandContext(ctx, "cwebp", "-quiet", "-q", "80", "-o", out, "--", "-")
		cmd.Stdin = &buf
		var stderr strings.Builder
		cmd.Stderr = &stderr
		err := cmd.Run()
		cancel()
		if err != nil {
			os.Remove(out)
			cleanup()
			return nil, fmt.Errorf("cwebp %dpx: %w: %s", w, err, stderr.String())
		}
		outputs = append(outputs, out)
	}
	return outputs, nil
}

func (fss *FileService) SaveUpload(fh *multipart.FileHeader) (string, error) {
	if fh.Size > fss.cfg.FileStorage.MaxUploadBytes {
		return "", ErrTooLarge
	}
	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	data, err := io.ReadAll(io.LimitReader(src, fss.cfg.FileStorage.MaxUploadBytes+1))
	if err != nil {
		return "", err
	}
	return fss.SaveBytes(data)
}

// SaveBytes validates, stores the original and generates the WebP variants.
func (fss *FileService) SaveBytes(data []byte) (string, error) {
	if int64(len(data)) > fss.cfg.FileStorage.MaxUploadBytes {
		return "", ErrTooLarge
	}

	ext, ok := allowedTypes[http.DetectContentType(data)]
	if !ok {
		return "", ErrUnsupportedType
	}

	conf, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || conf.Width*conf.Height > 50_000_000 {
		return "", ErrBadImage
	}

	name, err := randomName()
	if err != nil {
		return "", err
	}

	origPath := filepath.Join(fss.cfg.FileStorage.OriginalsDir, name+ext)
	if err := os.WriteFile(origPath, data, 0o644); err != nil {
		return "", err
	}

	if _, err := fss.ProcessImage(data, fss.cfg.FileStorage.UploadDir, name); err != nil {
		os.Remove(origPath)
		return "", err
	}
	return name, nil
}

func randomName() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil // 32 hex chars
}

func (fss *FileService) URL(name string, width int) string {
	return fmt.Sprintf("%s/uploads/%s-%d.webp", fss.cfg.Server.ServerURL, name, width)
}

func (fss *FileService) SrcSet(name string) string {
	parts := make([]string, 0, len(sizes))
	for _, w := range sizes {
		parts = append(parts, fmt.Sprintf("%s %dw", fss.URL(name, w), w))
	}
	return strings.Join(parts, ", ")
}

type noDirFS struct{ fs http.FileSystem }

func (n noDirFS) Open(p string) (http.File, error) {
	f, err := n.fs.Open(p)
	if err != nil {
		return nil, err
	}
	s, err := f.Stat()
	if err != nil || s.IsDir() {
		f.Close()
		return nil, os.ErrNotExist
	}
	return f, nil
}

func (fss *FileService) Handler() http.Handler {
	fs := http.FileServer(noDirFS{http.Dir(fss.cfg.FileStorage.UploadDir)})
	return http.StripPrefix("/uploads/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		fs.ServeHTTP(w, r)
	}))
}

var validName = regexp.MustCompile(`^[a-f0-9]{32}$`)

func (fss *FileService) Delete(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid image name")
	}

	paths := []string{
		filepath.Join(fss.cfg.FileStorage.OriginalsDir, name+".jpg"),
		filepath.Join(fss.cfg.FileStorage.OriginalsDir, name+".png"),
	}
	for _, width := range sizes {
		paths = append(paths,
			filepath.Join(fss.cfg.FileStorage.UploadDir, fmt.Sprintf("%s-%d.webp", name, width)),
		)
	}

	var deleteErr error
	for _, path := range paths {
		err := os.Remove(path)
		if err == nil || errors.Is(err, os.ErrNotExist) {
			continue
		}
		deleteErr = errors.Join(deleteErr, fmt.Errorf("%s: %w", path, err))
	}
	if deleteErr != nil {
		return fmt.Errorf("delete image files: %w", deleteErr)
	}
	return nil
}

// Hydrate fills SrcSet for images that live on this server.
func (fss *FileService) HydrateImages(images []tables.ProductImage) {
	for i := range images {
		img := &images[i]
		if img.Name == "" {
			continue
		}
		img.SrcSet = fss.SrcSet(img.Name)
	}
}

func (fss *FileService) HydrateProduct(p *tables.Product) {
	fss.HydrateImages(p.Images)
}

func (fss *FileService) HydrateProducts(ps []tables.Product) {
	for i := range ps {
		fss.HydrateProduct(&ps[i])
	}
}
