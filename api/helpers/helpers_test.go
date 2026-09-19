package helpers

import (
	"testing"
)

func TestValidateAndNormalizeURL(t *testing.T) {
	tests := []struct {
		name      string
		rawURL    string
		domain    string
		wantURL   string
		expectErr bool
	}{
		{
			name:      "Valid HTTPS URL",
			rawURL:    "https://example.com/test",
			domain:    "short.io",
			wantURL:   "https://example.com/test",
			expectErr: false,
		},
		{
			name:      "Prepend HTTPS scheme if missing",
			rawURL:    "example.com/test",
			domain:    "short.io",
			wantURL:   "https://example.com/test",
			expectErr: false,
		},
		{
			name:      "Block own domain",
			rawURL:    "https://short.io/abc",
			domain:    "http://short.io",
			expectErr: true,
		},
		{
			name:      "Invalid empty URL",
			rawURL:    "",
			domain:    "short.io",
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndNormalizeURL(tt.rawURL, tt.domain)
			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error %v, got %v", tt.expectErr, err)
			}
			if !tt.expectErr && got != tt.wantURL {
				t.Errorf("expected %s, got %s", tt.wantURL, got)
			}
		})
	}
}
func TestValidateCustomAlias(t *testing.T) {
	if err := ValidateCustomAlias("valid-alias_123"); err != nil {
		t.Errorf("expected valid alias, got %v", err)
	}
	if err := ValidateCustomAlias("ab"); err == nil {
		t.Errorf("expected error for too short alias")
	}
	if err := ValidateCustomAlias("api"); err == nil {
		t.Errorf("expected error for reserved keyword 'api'")
	}
}
