package auth

import (
	"context"
	"errors"
	"testing"
)

func TestGeminiAuthHandler_Setup(t *testing.T) {
	t.Run("valid API key set success", func(t *testing.T) {
		ctx := context.Background()
		setSecretCalled := false
		validKey := "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3hbgL-Td123" // 39 chars, alphanumeric + -_

		mockSetter := &MockSecretSetter{
			SetSecretFunc: func(ctx context.Context, repo, name, value string) error {
				setSecretCalled = true
				if name != "GEMINI_API_KEY" {
					t.Errorf("secret name = %s, want GEMINI_API_KEY", name)
				}
				if value != validKey {
					t.Errorf("secret value mismatch")
				}
				return nil
			},
		}

		handler := NewGeminiAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/repo", validKey)

		if err != nil {
			t.Errorf("Setup() error = %v, want nil", err)
		}
		if !setSecretCalled {
			t.Error("SetSecret was not called")
		}
	})

	t.Run("invalid API key format returns error - REQ-CI-010.1", func(t *testing.T) {
		ctx := context.Background()
		mockSetter := &MockSecretSetter{}

		handler := NewGeminiAuthHandler(mockSetter)

		testCases := []struct {
			name  string
			key   string
			valid bool
		}{
			{"too short", "short", false},
			{"invalid characters", "AIza$Invalid@Chars#Here", false},
			{"contains spaces", "AIzaSyDaGmWKa4JsXZ HjGw", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := handler.Setup(ctx, "owner/repo", tc.key)
				if err == nil && !tc.valid {
					t.Error("Setup() error = nil, want error (invalid key format)")
				}
			})
		}
	})

	t.Run("secret set failure returns error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("secret set failed")

		mockSetter := &MockSecretSetter{
			SetSecretFunc: func(ctx context.Context, repo, name, value string) error {
				return expectedErr
			},
		}

		handler := NewGeminiAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/repo", "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3hbgL-Td123")

		if err == nil {
			t.Error("Setup() error = nil, want error")
		}
	})
}

func TestValidateGeminiAPIKey(t *testing.T) {
	t.Run("valid format", func(t *testing.T) {
		validKeys := []string{
			"AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3hbgL-Td123",
			"AIza0123456789",
		}

		for _, key := range validKeys {
			err := validateGeminiAPIKey(key)
			if err != nil {
				t.Errorf("validateGeminiAPIKey(%s) error = %v, want nil", key, err)
			}
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		invalidKeys := []string{
			"",
			"short",
			"AIza$Invalid",
			"AIza WithSpace",
		}

		for _, key := range invalidKeys {
			err := validateGeminiAPIKey(key)
			if err == nil {
				t.Errorf("validateGeminiAPIKey(%s) error = nil, want error", key)
			}
		}
	})
}

func TestMaskGeminiKey(t *testing.T) {
	t.Run("normal masking", func(t *testing.T) {
		key := "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3hbgL-Td123"
		masked := maskGeminiKey(key)

		expected := "A...d123"
		if masked != expected {
			t.Errorf("maskGeminiKey() = %s, want %s", masked, expected)
		}
	})

	t.Run("short key", func(t *testing.T) {
		key := "AIza"
		masked := maskGeminiKey(key)

		if masked != "***" {
			t.Errorf("maskGeminiKey() = %s, want ***", masked)
		}
	})
}
