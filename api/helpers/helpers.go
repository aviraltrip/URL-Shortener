package helpers

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidURL   = errors.New("invalid URL")
	ErrSelfDomain   = errors.New("cannot shorten URL to own domain")
	ErrInvalidAlias = errors.New("alias must be 3-32 alphanumeric characters, hyphens, or underscores")
	ErrAliasKeyword = errors.New("alias cannot be a reserved system route")

	aliasRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)
	reserved   = map[string]bool{
		"api":     true,
		"health":  true,
		"ready":   true,
		"swagger": true,
		"metrics": true,
	}
)

func ValidateAndNormalizeURL(rawURL, appDomain string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", ErrInvalidURL
	}

	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Host == "" {
		return "", ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrInvalidURL
	}

	cleanDomain := strings.TrimPrefix(strings.TrimPrefix(appDomain, "http://"), "https://")
	cleanDomain = strings.Split(cleanDomain, "/")[0]

	if cleanDomain != "" && strings.EqualFold(parsed.Host, cleanDomain) {
		return "", ErrSelfDomain
	}

	return parsed.String(), nil
}

func ValidateCustomAlias(alias string) error {
	if alias == "" {
		return nil
	}
	if !aliasRegex.MatchString(alias) {
		return ErrInvalidAlias
	}
	if reserved[strings.ToLower(alias)] {
		return ErrAliasKeyword
	}
	return nil
}

func GenerateShortCode() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[:8]
}
