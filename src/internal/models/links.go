package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LinkMetadata contains additional information about the link
type LinkMetadata struct {
	Title       string            `bson:"title,omitempty" json:"title,omitempty"`
	Description string            `bson:"description,omitempty" json:"description,omitempty"`
	Tags        []string          `bson:"tags,omitempty" json:"tags,omitempty"`
	Category    string            `bson:"category,omitempty" json:"category,omitempty"`
	UserAgent   string            `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
	Custom      map[string]string `bson:"custom,omitempty" json:"custom,omitempty"`
}

// ShortLink represents a shortened URL in the system
type ShortLink struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OriginalURL string             `bson:"original_url" json:"original_url" validate:"required,url"`
	ShortCode   string             `bson:"short_code" json:"short_code"`
	CustomCode  *string            `bson:"custom_code,omitempty" json:"custom_code,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt   *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	ClickCount  int64              `bson:"click_count" json:"click_count"`
	CreatorIP   string             `bson:"creator_ip,omitempty" json:"creator_ip,omitempty"`
	Metadata    LinkMetadata       `bson:"metadata" json:"metadata"`
}
