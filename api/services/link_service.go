package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aviraltrip/urlshortener/config"
	"github.com/aviraltrip/urlshortener/db/generated"
	"github.com/aviraltrip/urlshortener/helpers"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound      = errors.New("short URL not found or expired")
	ErrAliasConflict = errors.New("custom alias is already in use")
)

type LinkService struct {
	cfg        *config.Config
	db         *pgxpool.Pool
	queries    *generated.Queries
	redisCache *redis.Client
}

func NewLinkService(cfg *config.Config, db *pgxpool.Pool, redisCache *redis.Client) *LinkService {
	return &LinkService{
		cfg:        cfg,
		db:         db,
		queries:    generated.New(db),
		redisCache: redisCache,
	}
}

type ShortenResult struct {
	Link     generated.Link
	ShortURL string
}

func (s *LinkService) Shorten(ctx context.Context, rawURL, customShort string, expiryHours int) (*ShortenResult, error) {
	normURL, err := helpers.ValidateAndNormalizeURL(rawURL, s.cfg.Domain)
	if err != nil {
		return nil, err
	}
	if err := helpers.ValidateCustomAlias(customShort); err != nil {
		return nil, err
	}
	if expiryHours <= 0 {
		expiryHours = 24
	}
	shortCode := customShort
	if shortCode == "" {
		shortCode = helpers.GenerateShortCode()
	}
	expiresAt := time.Now().UTC().Add(time.Duration(expiryHours) * time.Hour)
	linkID := uuid.New()
	link, err := s.queries.CreateLink(ctx, generated.CreateLinkParams{
		ID:          linkID,
		ShortCode:   shortCode,
		OriginalUrl: normURL,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAliasConflict
		}
		return nil, fmt.Errorf("failed to create link: %w", err)
	}
	ttl := time.Until(link.ExpiresAt)
	if ttl > 0 && s.redisCache != nil {
		_ = s.redisCache.Set(ctx, link.ShortCode, link.OriginalUrl, ttl).Err()
	}
	domain := strings.TrimRight(s.cfg.Domain, "/")
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "http://" + domain
	}
	return &ShortenResult{
		Link:     link,
		ShortURL: fmt.Sprintf("%s/%s", domain, link.ShortCode),
	}, nil
}

func (s *LinkService) Resolve(ctx context.Context, shortCode string) (string, error) {
	if s.redisCache != nil {
		cachedURL, err := s.redisCache.Get(ctx, shortCode).Result()
		if err == nil && cachedURL != "" {
			go func(code string) {
				bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_ = s.queries.IncrementClickCount(bgCtx, code)
			}(shortCode)
			return cachedURL, nil
		}
	}
	link, err := s.queries.GetLinkByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to query link: %w", err)
	}
	ttl := time.Until(link.ExpiresAt)
	if ttl > 0 && s.redisCache != nil {
		_ = s.redisCache.Set(ctx, link.ShortCode, link.OriginalUrl, ttl).Err()
	}
	go func(code string) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = s.queries.IncrementClickCount(bgCtx, code)
	}(shortCode)

	return link.OriginalUrl, nil
}

func (s *LinkService) GetStats(ctx context.Context, shortCode string) (*generated.Link, error) {
	link, err := s.queries.GetLinkDetails(ctx, shortCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get link stats: %w", err)
	}
	return &link, nil
}

func (s *LinkService) Delete(ctx context.Context, shortCode string) error {
	if err := s.queries.DeleteLink(ctx, shortCode); err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}
	if s.redisCache != nil {
		_ = s.redisCache.Del(ctx, shortCode).Err()
	}
	return nil
}

func (s *LinkService) List(ctx context.Context, page, limit int32) ([]generated.Link, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	links, err := s.queries.ListLinks(ctx, generated.ListLinksParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list links: %w", err)
	}
	return links, nil
}
