package models

import (
	"github.com/google/uuid"
)

func NewImage(id uuid.UUID, contentType string, content []byte) Image {
	return Image{
		ID:          id,
		ContentType: contentType,
		Content:     content,
	}
}

type Image struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	ContentType string    `json:"content_type" gorm:"type:varchar(255);not null"`
	Content     []byte    `json:"content" gorm:"type:bytea;not null"`
}