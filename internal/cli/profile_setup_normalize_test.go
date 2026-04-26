package cli

import "testing"

// TestNormalizeStatuslineMode_Canonical verifies that canonical values pass through unchanged.
func TestNormalizeStatuslineMode_Canonical(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"default", "default"},
		{"full", "full"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := normalizeStatuslineMode(tt.in); got != tt.want {
				t.Errorf("normalizeStatuslineMode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestNormalizeStatuslineMode_Deprecated verifies that deprecated aliases are converted to v3 canonical values.
// minimal/compact → default, verbose → full (consistent with statusline.NormalizeMode behavior).
func TestNormalizeStatuslineMode_Deprecated(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"minimal", "default"},
		{"compact", "default"},
		{"verbose", "full"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := normalizeStatuslineMode(tt.in); got != tt.want {
				t.Errorf("normalizeStatuslineMode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestNormalizeStatuslineMode_EmptyAndUnknown verifies that empty strings and unknown values fall back to "default".
func TestNormalizeStatuslineMode_EmptyAndUnknown(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "default"},
		{"something-else", "default"},
		{"extra", "default"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := normalizeStatuslineMode(tt.in); got != tt.want {
				t.Errorf("normalizeStatuslineMode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestNormalizeStatuslineTheme verifies that valid themes pass through unchanged
// and that unknown or legacy "default" values are converted to "catppuccin-mocha".
func TestNormalizeStatuslineTheme(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"catppuccin-mocha", "catppuccin-mocha"},
		{"catppuccin-latte", "catppuccin-latte"},
		{"default", "catppuccin-mocha"},
		{"", "catppuccin-mocha"},
		{"custom-theme", "catppuccin-mocha"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := normalizeStatuslineTheme(tt.in); got != tt.want {
				t.Errorf("normalizeStatuslineTheme(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestNormalizeModel_Canonical verifies that canonical aliases pass through unchanged.
func TestNormalizeModel_Canonical(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"opus", "opus"},
		{"opus[1m]", "opus[1m]"},
		{"sonnet", "sonnet"},
		{"sonnet[1m]", "sonnet[1m]"},
		{"haiku", "haiku"},
		{"opusplan", "opusplan"},
	}
	for _, tt := range tests {
		t.Run("canonical/"+tt.in, func(t *testing.T) {
			if got := normalizeModel(tt.in); got != tt.want {
				t.Errorf("normalizeModel(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestNormalizeModel_Deprecated verifies that deprecated full-IDs are converted to canonical aliases.
func TestNormalizeModel_Deprecated(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"claude-opus-4-7", "opus"},
		{"claude-opus-4-6", "opus"},
		{"claude-opus-4-7[1m]", "opus[1m]"},
		{"claude-opus-4-6[1m]", "opus[1m]"},
		{"claude-opus-4-6 1M", "opus[1m]"},
		{"claude-sonnet-4-6", "sonnet"},
		{"claude-sonnet-4-6[1m]", "sonnet[1m]"},
		{"claude-sonnet-4-6 1M", "sonnet[1m]"},
		{"claude-haiku-4-5", "haiku"},
	}
	for _, tt := range tests {
		t.Run("deprecated/"+tt.in, func(t *testing.T) {
			if got := normalizeModel(tt.in); got != tt.want {
				t.Errorf("normalizeModel(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestNormalizeModel_EmptyAndUnknown verifies that empty strings return empty strings
// and unknown values return "" (runtime default).
func TestNormalizeModel_EmptyAndUnknown(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"unknown-model", ""},
		{"claude-opus-99", ""},
		{"gpt-4", ""},
	}
	for _, tt := range tests {
		t.Run("unknown/"+tt.in, func(t *testing.T) {
			if got := normalizeModel(tt.in); got != tt.want {
				t.Errorf("normalizeModel(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestValueOrDash verifies that "-" is returned for empty strings.
func TestValueOrDash(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "-"},
		{"alice", "alice"},
		{"  ", "  "},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := valueOrDash(tt.in); got != tt.want {
				t.Errorf("valueOrDash(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestValueOrDefault verifies that the fallback is returned for empty strings.
func TestValueOrDefault(t *testing.T) {
	tests := []struct {
		in, fallback, want string
	}{
		{"", "fallback", "fallback"},
		{"value", "fallback", "value"},
		{"", "", ""},
		{"opus", "sonnet", "opus"},
	}
	for _, tt := range tests {
		t.Run(tt.in+"/"+tt.fallback, func(t *testing.T) {
			if got := valueOrDefault(tt.in, tt.fallback); got != tt.want {
				t.Errorf("valueOrDefault(%q, %q) = %q, want %q", tt.in, tt.fallback, got, tt.want)
			}
		})
	}
}
