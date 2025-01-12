package services

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"testapp/internal/models"
	"testapp/internal/repositories"
)

type ImageService struct {
	rep repositories.ImageRepository
}

func NewImageService(rep repositories.ImageRepository) *ImageService {
	return &ImageService{rep:rep}
}

func (s *ImageService) SaveFile(ctx context.Context, filename string, content io.Reader) error {
	if !DirectoryExists(filepath.Join(".", "assets", "uploads")) {
		return os.ErrNotExist
	}

	filename = filepath.Join(".", "assets", "uploads", filename)
	newFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newFile.Close()

	if _, err := io.Copy(newFile, content); err != nil {
		return err
	}

	return nil
}

func (s *ImageService) SaveFileToDB(ctx context.Context, contentType string, content io.Reader) error {
	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	image := models.NewImage(uuid.New(), contentType, contentBytes)

	if err = s.rep.Create(ctx, image); err != nil {
		return err
	}

	return nil
}

func (s *ImageService) Get(ctx context.Context, id uuid.UUID) (models.Image, error) {
	return s.rep.Get(ctx, id)
}

func (s *ImageService) ReadFile(ctx context.Context, filename string) ([]byte, error) {
	fileBytes, err := os.ReadFile(filepath.Join("assets", "tmp", filename))
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}