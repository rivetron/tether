package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Kosha-Nirman/tether/src/internal/models"
	"github.com/Kosha-Nirman/tether/src/internal/repository"
	"github.com/Kosha-Nirman/tether/src/pkg/cache"
	"github.com/Kosha-Nirman/tether/src/pkg/config"
	"github.com/Kosha-Nirman/tether/src/pkg/utils"
)

type LinkService struct {
	linkRepo *repository.LinkRepository
	cache    *cache.Redis
	config   *config.Config
}

func NewLinkService(linkRepo *repository.LinkRepository, cache *cache.Redis, config *config.Config) *LinkService {
	return &LinkService{
		linkRepo,
		cache,
		config,
	}
}

func (s *LinkService) CreateLink(ctx context.Context, req *models.CreateLinkRequest, creatorIP string) (*models.ShortLink, error) {
	// ? Validate the original URL
	if !utils.IsValidURL(req.OriginalURL) {
		return nil, fmt.Errorf("invalid URL format")
	}

	// ? Normalize the URL
	normalizedURL := utils.NormalizeURL(req.OriginalURL)

	// ? Check if domain is blocked
	if utils.IsBlockedDomain(normalizedURL, s.config.Security.BlockedDomains) {
		return nil, fmt.Errorf("domain not allowed")
	}

	// * Generate a valid short-code
	var shortCode string
	var err error

	if req.CustomCode != nil && *req.CustomCode != "" {
		customCode := utils.SanitizeCustomCode(*req.CustomCode)
		if !utils.IsValidShortCode(customCode) {
			return nil, fmt.Errorf("invalid custom code")
		}

		// ? Check if custom code already exists
		exists, err := s.linkRepo.CheckShortCodeExists(ctx, customCode)
		if err != nil {
			return nil, fmt.Errorf("failed to check custom code availability: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("custom code already in use")
		}

		shortCode = customCode
	} else {
		// * Generate a unique short-code
		shortCode, err = s.generateUniqueShortCode(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short-code: %w", err)
		}
	}

	// * Calculate expiration time
	var expiresAt *time.Time
	if req.TTLHours != nil {
		if *req.TTLHours > 0 {
			expiry := time.Now().Add(time.Duration(*req.TTLHours) * time.Hour)
			expiresAt = &expiry
		}
	} else {
		// * Default TTL of 1 month (720 hours)
		expiry := time.Now().Add(720 * time.Hour)
		expiresAt = &expiry
	}

	// ? Validate TTL doesn't exceed maximum limit
	if expiresAt != nil && expiresAt.After(time.Now().Add(s.config.App.MaxTTL)) {
		return nil, fmt.Errorf("TTL exceeds maximum allowed duration")
	}

	// * Prepare metadata
	metadata := models.LinkMetadata{}
	if req.Metadata != nil {
		metadata = *req.Metadata
	}
	if req.Custom != nil {
		metadata.Custom = req.Custom
	}

	// * Create the link
	link := &models.ShortLink{
		OriginalURL: normalizedURL,
		ShortCode:   shortCode,
		CustomCode:  req.CustomCode,
		ExpiresAt:   expiresAt,
		CreatorIP:   creatorIP,
		Metadata:    metadata,
	}

	createdLink, err := s.linkRepo.Create(ctx, link)
	if err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	cacheKey := cache.LinkCacheKey(shortCode)
	cacheTTL := s.config.Cache.TTL
	if expiresAt != nil {
		remaining := time.Until(*expiresAt)
		if remaining < cacheTTL {
			cacheTTL = remaining
		}
	}

	if err := s.cache.Set(ctx, cacheKey, link, cacheTTL); err != nil {
		log.Printf("Failed to cache link: %v", err)
	}

	return createdLink, nil
}

func (s *LinkService) GetLink(ctx context.Context, shortCode string) (*models.ShortLink, error) {
	// * Try cache first
	cacheKey := cache.LinkCacheKey(shortCode)
	var link models.ShortLink

	err := s.cache.Get(ctx, cacheKey, &link)
	if err == nil {
		// ? Check if link is still valid
		if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
			// * Link has expired, remove from cache
			if err := s.cache.Delete(ctx, cacheKey); err != nil {
				log.Printf("Warning: Failed to delete expired link from cache: %v", err)
			}

			return nil, fmt.Errorf("link has expired")
		}

		return &link, nil
	}

	// ? Cache miss
	linkPtr, err := s.linkRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	// * Cache the result
	cacheTTL := s.config.Cache.TTL
	if linkPtr.ExpiresAt != nil {
		remaining := time.Until(*linkPtr.ExpiresAt)
		if remaining < cacheTTL {
			cacheTTL = remaining
		}
	}

	if err := s.cache.Set(ctx, cacheKey, link, cacheTTL); err != nil {
		log.Printf("Failed to cache link: %v", err)
	}

	return linkPtr, nil
}

func (s *LinkService) UpdateLink(
	ctx context.Context,
	shortCode string,
	mutate func(link *models.ShortLink) error,
) (*models.ShortLink, error) {
	// * Fetch existing link
	link, err := s.linkRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	// * Apply mutation and handle its error
	if err := mutate(link); err != nil {
		return nil, err
	}

	// * Central TTL validation: ensure ExpiresAt does not exceed MaxTTL
	if link.ExpiresAt != nil {
		if link.ExpiresAt.After(time.Now().Add(s.config.App.MaxTTL)) {
			return nil, fmt.Errorf("expires_at exceeds maximum allowed TTL")
		}
	}

	// * Persist DB update
	updatedLink, err := s.linkRepo.UpdateByShortCode(ctx, shortCode, link)
	if err != nil {
		return nil, fmt.Errorf("failed to update link: %w", err)
	}

	// * Cache handling
	cacheKey := cache.LinkCacheKey(shortCode)

	// ? If the link is expired (ExpiresAt set and in past) -> delete cache
	if updatedLink.ExpiresAt != nil && time.Until(*updatedLink.ExpiresAt) <= 0 {
		if err := s.cache.Delete(ctx, cacheKey); err != nil {
			log.Printf("Warning: failed to delete expired link from cache (%s): %v", shortCode, err)
		}
		return updatedLink, nil
	}

	// * Determine TTL for cache entry
	cacheTTL := s.cache.DefaultTTL
	if updatedLink.ExpiresAt != nil {
		remaining := time.Until(*updatedLink.ExpiresAt)
		if remaining < cacheTTL {
			cacheTTL = remaining
		}
	}

	if err := s.cache.Set(ctx, cacheKey, updatedLink, cacheTTL); err != nil {
		log.Printf("Warning: failed to update cache for link %s: %v", shortCode, err)
	}

	return updatedLink, nil
}

func (s *LinkService) DeleteLink(ctx context.Context, shortCode string) error {
	if err := s.linkRepo.DeleteByShortCode(ctx, shortCode); err != nil {
		return err
	}

	// * Remove from cache
	cacheKey := cache.LinkCacheKey(shortCode)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.Printf("Warning: Failed to delete link from cache: %v", err)
	}

	return nil
}

// <-------------------- Service Helper Functions -------------------->

func (s *LinkService) IncrementClickCount(ctx context.Context, shortCode string) (*models.ShortLink, error) {
	return s.UpdateLink(ctx, shortCode, func(link *models.ShortLink) error {
		link.ClickCount++
		return nil
	})
}

func (s *LinkService) ListLinks(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.ShortLink, int64, error) {
	links, err := s.linkRepo.List(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.linkRepo.Count(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	return links, total, nil
}

// <-------------------- Helper Functions -------------------->

func (s *LinkService) generateUniqueShortCode(ctx context.Context) (string, error) {
	maxAttempts := 10
	codeLength := s.config.App.ShortCodeLen

	for i := range maxAttempts {
		code, err := utils.GenerateShortCode(codeLength)
		if err != nil {
			return "", err
		}

		exists, err := s.linkRepo.CheckShortCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}

		if i > 5 {
			codeLength++
		}
	}

	return "", fmt.Errorf("failed to generate unique short-code after %d attempts", maxAttempts)
}
