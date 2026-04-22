package cli

import "testing"

// TestNormalizeStatuslineMode_Canonical canonical 값이 변경 없이 통과하는지 확인한다.
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

// TestNormalizeStatuslineMode_Deprecated deprecated alias가 v3 canonical 값으로 변환되는지 확인한다.
// minimal/compact → default, verbose → full (statusline.NormalizeMode 동작과 일치).
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

// TestNormalizeStatuslineMode_EmptyAndUnknown 빈 문자열과 알 수 없는 값이 "default"로 폴백하는지 확인한다.
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

// TestNormalizeStatuslineTheme 유효한 테마는 그대로 통과하고,
// 알 수 없거나 레거시 "default" 값은 "catppuccin-mocha"로 변환되는지 확인한다.
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

// TestNormalizeModel_Canonical canonical alias는 그대로 통과하는지 확인한다.
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

// TestNormalizeModel_Deprecated deprecated full-ID가 canonical alias로 변환되는지 확인한다.
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

// TestNormalizeModel_EmptyAndUnknown 빈 문자열은 빈 문자열을 반환하고,
// 알 수 없는 값은 ""(런타임 기본값)을 반환하는지 확인한다.
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

// TestValueOrDash 빈 문자열에서 "-"를 반환하는지 확인한다.
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

// TestValueOrDefault 빈 문자열에서 fallback을 반환하는지 확인한다.
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
