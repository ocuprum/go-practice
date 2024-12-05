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
	ID          uuid.UUID `json:"id"`
	ContentType string    `json:"content_type"`
	Content     []byte    `json:"content"`
}