package service

import (
	"context"
	"log"
	"time"
)

type CleanupService struct {
	linkService *LinkService
	ticker      *time.Ticker
	stopCh      chan struct{}
}

func NewCleanupService(linkService *LinkService) *CleanupService {
	return &CleanupService{
		linkService: linkService,
		stopCh:      make(chan struct{}),
	}
}

func (s *CleanupService) Start() {
	// ? Run Cleanup every 3 hours
	s.ticker = time.NewTicker(3 * time.Hour)

	// * Run initial cleanup
	go s.runCleanup()

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.runCleanup()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *CleanupService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}

	close(s.stopCh)
}

func (s *CleanupService) ForceCleanup(ctx context.Context) (int, error) {
	return s.linkService.CleanupExpiredLinks(ctx)
}

func (s *CleanupService) runCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	log.Println("Starting scheduled cleanup...")

	// * Clean up expired links
	expiredCount, err := s.linkService.CleanupExpiredLinks(ctx)
	if err != nil {
		log.Printf("Error cleaning up expired links: %v", err)
	} else {
		log.Printf("Cleaned up %d expired links", expiredCount)
	}

	log.Println("Scheduled cleanup completed.")
}
