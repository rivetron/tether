package repository

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/Kosha-Nirman/tether/src/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LinkRepository struct {
	collection *mongo.Collection
}

func NewLinkRepository(db *mongo.Database) *LinkRepository {
	return &LinkRepository{
		collection: db.Collection("short_links"),
	}
}

// ? Creates a new short link
func (r *LinkRepository) Create(ctx context.Context, link *models.ShortLink) (*models.ShortLink, error) {
	link.ID = primitive.NewObjectID()
	link.CreatedAt = time.Now()
	link.IsActive = true
	link.ClickCount = 0

	result, err := r.collection.InsertOne(ctx, link)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("short code already exists")
		}

		return nil, fmt.Errorf("failed to create links: %w", err)
	}

	link.ID = result.InsertedID.(primitive.ObjectID)
	return link, nil
}

// ? Get a short link by its ID
func (r *LinkRepository) Get(ctx context.Context, id string) (*models.ShortLink, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	var link models.ShortLink

	filter := bson.M{"_id": objectID}

	err = r.collection.FindOne(ctx, filter).Decode(&link)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("link not found")
		}

		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	return &link, nil
}

// ? Get a short link by its short code
func (r *LinkRepository) GetByShortCode(ctx context.Context, shortCode string) (*models.ShortLink, error) {
	var link models.ShortLink

	filter := bson.M{
		"short_code": shortCode,
		"is_active":  true,
		"$or": []bson.M{
			{"expires_at": bson.M{"$gt": time.Now()}},
			{"expires_at": nil},
		},
	}

	err := r.collection.FindOne(ctx, filter).Decode(&link)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("link not found")
		}

		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	return &link, nil
}

// ? Updates a link by ID and returns the updated document
func (r *LinkRepository) Update(ctx context.Context, id primitive.ObjectID, link *models.ShortLink) (*models.ShortLink, error) {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": link}

	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedLink models.ShortLink
	err := r.collection.FindOneAndUpdate(ctx, filter, update, options).Decode(&updatedLink)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("link not found")
		}

		return nil, fmt.Errorf("failed to update link: %w", err)
	}

	return &updatedLink, nil
}

// ? Updates a link by short code and returns the updated document
func (r *LinkRepository) UpdateByShortCode(ctx context.Context, shortCode string, link *models.ShortLink) (*models.ShortLink, error) {
	filter := bson.M{"short_code": shortCode}
	update := bson.M{"$set": link}

	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedLink models.ShortLink
	err := r.collection.FindOneAndUpdate(ctx, filter, update, options).Decode(&updatedLink)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("link not found")
		}

		return nil, fmt.Errorf("failed to update link: %w", err)
	}

	return &updatedLink, nil
}

// ? Deletes a link by ID, setting is_active to false
func (r *LinkRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"is_active": false}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("link not found")
	}

	return nil
}

// ? Deletes a link by short code, setting is_active to false
func (r *LinkRepository) DeleteByShortCode(ctx context.Context, shortCode string) error {
	filter := bson.M{"short_code": shortCode}
	update := bson.M{"$set": bson.M{"is_active": false}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("link not found")
	}

	return nil
}

// <-------------------- Helper Functions -------------------->

// ? Increment Click count
func (r *LinkRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	filter := bson.M{"short_code": shortCode}
	update := bson.M{"$inc": bson.M{"click_count": 1}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}

	return nil
}

// ? Check if short link exists
func (r *LinkRepository) CheckShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	filter := bson.M{"short_code": shortCode}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check short code existence: %w", err)
	}

	return count > 0, nil
}

// ? Bulk deactivate links
func (r *LinkRepository) BulkDeactivate(ctx context.Context, shortCodes []string) error {
	filter := bson.M{"short_code": bson.M{"$in": shortCodes}}
	update := bson.M{"$set": bson.M{"is_active": false}}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to bulk deactivate links: %w", err)
	}

	return nil
}

// ? Count of links based on filter
func (r *LinkRepository) Count(ctx context.Context, filters map[string]any) (int64, error) {
	filter := bson.M{"is_active": true}
	maps.Copy(filter, filters)

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count links: %w", err)
	}

	return count, nil
}
