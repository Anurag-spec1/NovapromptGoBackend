package models

import "time"

type Image struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	CloudinaryURL string    `json:"cloudinary_url"`
	CategoryID    *string   `json:"category_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Description struct {
	ID        string    `json:"id"`
	ImageID   string    `json:"image_id"`
	Body      string    `json:"body"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AdKey struct {
	Key       string          `json:"key"`
	Value     map[string]any  `json:"value"`
	UpdatedAt time.Time       `json:"updated_at"`
}