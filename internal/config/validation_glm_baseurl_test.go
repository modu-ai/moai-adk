package config

import (
	"errors"
	"testing"
)

// TestValidateLLM_RejectsUnsafeGLMBaseURL verifies AC-CGH-007 Scenario 7a/7b
// (REQ-CGH-007, SECURITY): a GLM base_url is validated as a well-formed https://
// URL whose host is the canonical api.z.ai family OR a legitimate user override.
// Malformed / non-https / hostless values are rejected with a clear error.
// DefaultGLMBaseURL always passes. Empty passes (falls back to default on load,
// EC-5).
func TestValidateLLM_RejectsUnsafeGLMBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		// Positive cases — must pass.
		{"canonical default", DefaultGLMBaseURL, false},
		{"z.ai apex host", "https://z.ai/api/anthropic", false},
		{"z.ai subdomain", "https://api.z.ai/api/anthropic", false},
		{"empty falls back to default", "", false},
		{"legitimate user override https self-host", "https://glm-proxy.mycompany.com/v1", false},

		// Negative cases — must reject.
		{"non-https http scheme", "http://api.z.ai/api/anthropic", true},
		{"ftp scheme", "ftp://api.z.ai/api/anthropic", true},
		{"no scheme bare host", "api.z.ai", true},
		{"missing host", "https://", true},
		{"garbage", "not a url at all", true},
		{"scheme-relative", "//api.z.ai/api/anthropic", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errs := validateGLMBaseURL(tt.baseURL)
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("expected error for base_url %q, got none", tt.baseURL)
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("expected no error for base_url %q, got: %v", tt.baseURL, errs)
			}
			if tt.wantErr && len(errs) > 0 {
				if !errors.Is(&errs[0], ErrInvalidConfig) {
					t.Errorf("expected ErrInvalidConfig, got: %v", errs[0].Wrapped)
				}
				if errs[0].Field != "llm.glm.base_url" {
					t.Errorf("expected field llm.glm.base_url, got: %q", errs[0].Field)
				}
			}
		})
	}
}

// TestValidate_RejectsUnsafeGLMBaseURLEndToEnd verifies the validation is wired
// into the top-level Validate() so a malformed base_url in a loaded config fails.
func TestValidate_RejectsUnsafeGLMBaseURLEndToEnd(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	cfg.LLM.GLM.BaseURL = "http://evil.example.com/steal-token"

	err := Validate(cfg, map[string]bool{})
	if err == nil {
		t.Fatal("expected Validate to reject an unsafe (non-https) GLM base_url")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got: %v", err)
	}
}

// TestValidate_AcceptsDefaultGLMBaseURLEndToEnd verifies the canonical default
// passes the end-to-end Validate() path (AC-CGH-007 Scenario 7b).
func TestValidate_AcceptsDefaultGLMBaseURLEndToEnd(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	if err := Validate(cfg, map[string]bool{}); err != nil {
		t.Errorf("default config (DefaultGLMBaseURL) must validate clean, got: %v", err)
	}
}
