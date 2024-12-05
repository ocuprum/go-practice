package pgsql

import (
	"context"
	"testapp/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func NewImageRepository(conn *gorm.DB) *ImageRepository {
	return &ImageRepository{conn: conn}
}

type ImageRepository struct {
	conn *gorm.DB
}

func (r *ImageRepository) Paginate(ctx context.Context, limit, offset int) (images []models.Image, err error) {
	err = r.conn.WithContext(ctx).Limit(limit).Offset(offset).Find(&images).Error
	if err != nil {
		return []models.Image{}, err
	}
	return images, nil
}

func (r *ImageRepository) Create(ctx context.Context, image models.Image) error {
	return r.conn.WithContext(ctx).Create(&image).Error
}

func (r *ImageRepository) Get(ctx context.Context, id uuid.UUID) (image models.Image, err error) {
	err = r.conn.WithContext(ctx).Where("id = ?", id).First(&image).Error
	if err != nil {
		return models.Image{}, err
	}
	return image, nil
}

func (r *ImageRepository) Delete(ctx context.Context, ids []uuid.UUID) error {
	return r.conn.WithContext(ctx).Where("id IN (?)", ids).Delete(&models.Image{}).Error
}
