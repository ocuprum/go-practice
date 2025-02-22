package repositories

import (
	"context"
	"github.com/google/uuid"

	"testapp/internal/models"
)

type ImageRepository interface {
	Paginate(ctx context.Context, limit, offset int) ([]models.Image, error)
	Create(ctx context.Context, image models.Image) error
	Get(ctx context.Context, id uuid.UUID) (models.Image, error)
	Delete(ctx context.Context, ids []uuid.UUID) error
}