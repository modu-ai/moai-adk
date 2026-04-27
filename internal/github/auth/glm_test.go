package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestGLMAuthHandler_Setup(t *testing.T) {
	t.Run("GLM 토큰 설정 성공", func(t *testing.T) {
		ctx := context.Background()
		setSecretCalls := []string{} // secret names tracking
		testToken := "test-glm-token-12345"

		mockSetter := &MockSecretSetter{
			SetSecretFunc: func(ctx context.Context, repo, name, value string) error {
				setSecretCalls = append(setSecretCalls, name)

				// GLM_API_KEY가 먼저 설정되어야 함
				if name == "GLM_API_KEY" && value != testToken {
					t.Errorf("GLM_API_KEY value mismatch: got %s, want %s", value, testToken)
				}
				return nil
			},
		}

		handler := NewGLMAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/repo", testToken)

		if err != nil {
			t.Errorf("Setup() error = %v, want nil", err)
		}

		// 5개 시크릿이 설정되어야 함 (GLM_API_KEY + 4개 env var)
		expectedCount := 5
		if len(setSecretCalls) != expectedCount {
			t.Errorf("SetSecret 호출 횟수 = %d, want %d", len(setSecretCalls), expectedCount)
		}

		// 첫 번째는 GLM_API_KEY여야 함
		if setSecretCalls[0] != "GLM_API_KEY" {
			t.Errorf("첫 번째 시크릿 = %s, want GLM_API_KEY", setSecretCalls[0])
		}

		// SPEC-GLM-001 env vars가 설정되어야 함
		expectedEnvVars := []string{
			"DISABLE_BETAS",
			"DISABLE_PROMPT_CACHING",
			"CLAUDE_CODE_USE_bedrock",
			"CLAUDE_CODE_USE_vertex",
		}

		for i, envVar := range expectedEnvVars {
			found := false
			for _, call := range setSecretCalls {
				if strings.Contains(call, envVar) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("SPEC-GLM-001 env var %s가 설정되지 않음", envVar)
			}
			_ = i // unused 변수 제거
		}
	})

	t.Run("빈 토큰은 에러", func(t *testing.T) {
		ctx := context.Background()
		mockSetter := &MockSecretSetter{}

		handler := NewGLMAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/repo", "")

		if err == nil {
			t.Error("Setup() error = nil, want error (empty token)")
		}
	})

	t.Run("secret 설정 실패 시 에러 반환", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("secret set failed")

		mockSetter := &MockSecretSetter{
			SetSecretFunc: func(ctx context.Context, repo, name, value string) error {
				return expectedErr
			},
		}

		handler := NewGLMAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/repo", "test-token")

		if err == nil {
			t.Error("Setup() error = nil, want error")
		}
	})
}

func TestValidateGLMToken(t *testing.T) {
	t.Run("유효한 토큰", func(t *testing.T) {
		validTokens := []string{
			"test-token-12345",
			"sk-valid-key",
			"a",
		}

		for _, token := range validTokens {
			err := validateGLMToken(token)
			if err != nil {
				t.Errorf("validateGLMToken(%s) error = %v, want nil", token, err)
			}
		}
	})

	t.Run("빈 토큰은 에러", func(t *testing.T) {
		err := validateGLMToken("")
		if err == nil {
			t.Error("validateGLMToken() error = nil, want error (empty token)")
		}
	})

	t.Run("공백만 있는 토큰은 에러", func(t *testing.T) {
		err := validateGLMToken("   ")
		if err == nil {
			t.Error("validateGLMToken() error = nil, want error (whitespace only)")
		}
	})
}
