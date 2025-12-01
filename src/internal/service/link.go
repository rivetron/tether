package service

import (
	"context"
	"fmt"
	"strings"
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

func (s *LinkService) CreateLink(ctx context.Context, req *models.CreateLinkRequest, creatorIP string) (*models.CreateLinkResponse, error) {
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

	// TODO: Cache the link for faster access

	// * Generate the short URL
	shortURL := fmt.Sprintf("%s/%s", strings.TrimRight(s.config.Server.BaseURL, "/"), shortCode)

	response := &models.CreateLinkResponse{
		ID:          createdLink.ID.Hex(),
		OriginalURL: createdLink.OriginalURL,
		CreatedAt:   createdLink.CreatedAt,
		ExpiresAt:   createdLink.ExpiresAt,
		ShortCode:   createdLink.ShortCode,
		ShortURL:    shortURL,
	}

	return response, nil
}

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
