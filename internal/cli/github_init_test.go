// Package cli는 GitHub init 명령에 대한 테스트를 제공합니다.
// Package cli provides tests for GitHub init command.
package cli

import (
	"bytes"
	"context"
	"testing"
)

func TestNewInitCmd(t *testing.T) {
	cmd := newInitCmd()

	if cmd == nil {
		t.Fatal("newInitCmd returned nil")
	}

	if cmd.Use != "init" {
		t.Errorf("expected Use 'init', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestRunInit(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "default execution",
			args:    []string{},
			wantErr: false, // GREEN: 구현 완료
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newInitCmd()
			cmd.SetArgs(tt.args)

			out := &bytes.Buffer{}
			cmd.SetOut(out)

			err := cmd.RunE(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunE() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockInitHandler struct {
	detectRepoFunc    func() (string, error)
	promptLLMsFunc    func(ctx context.Context) ([]string, error)
	setupAuthFunc     func(ctx context.Context, llms []string) error
	installRunnerFunc func(ctx context.Context) error
}

func (m *mockInitHandler) DetectRepo() (string, error) {
	if m.detectRepoFunc != nil {
		return m.detectRepoFunc()
	}
	return "owner/repo", nil
}

func (m *mockInitHandler) PromptLLMs(ctx context.Context) ([]string, error) {
	if m.promptLLMsFunc != nil {
		return m.promptLLMsFunc(ctx)
	}
	return []string{"claude", "glm"}, nil
}

func (m *mockInitHandler) SetupAuth(ctx context.Context, llms []string) error {
	if m.setupAuthFunc != nil {
		return m.setupAuthFunc(ctx, llms)
	}
	return nil
}

func (m *mockInitHandler) InstallRunner(ctx context.Context) error {
	if m.installRunnerFunc != nil {
		return m.installRunnerFunc(ctx)
	}
	return nil
}

func TestInitHandlerIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("successful init flow", func(t *testing.T) {
		handler := &mockInitHandler{
			detectRepoFunc: func() (string, error) {
				return "test-owner/test-repo", nil
			},
			promptLLMsFunc: func(ctx context.Context) ([]string, error) {
				return []string{"claude", "glm", "gemini"}, nil
			},
			setupAuthFunc: func(ctx context.Context, llms []string) error {
				return nil
			},
			installRunnerFunc: func(ctx context.Context) error {
				return nil
			},
		}

		repo, err := handler.DetectRepo()
		if err != nil {
			t.Fatalf("DetectRepo failed: %v", err)
		}

		if repo != "test-owner/test-repo" {
			t.Errorf("expected repo 'test-owner/test-repo', got %q", repo)
		}

		llms, err := handler.PromptLLMs(ctx)
		if err != nil {
			t.Fatalf("PromptLLMs failed: %v", err)
		}

		if len(llms) != 3 {
			t.Errorf("expected 3 LLMs, got %d", len(llms))
		}

		if err := handler.SetupAuth(ctx, llms); err != nil {
			t.Fatalf("SetupAuth failed: %v", err)
		}

		if err := handler.InstallRunner(ctx); err != nil {
			t.Fatalf("InstallRunner failed: %v", err)
		}
	})
}
