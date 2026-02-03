package hook

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestContractValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		workDir   string
		setupCtx  func() (context.Context, context.CancelFunc)
		wantErr   bool
		errTarget error
	}{
		{
			name:    "valid working directory",
			workDir: t.TempDir(),
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			wantErr: false,
		},
		{
			name:      "non-existent working directory",
			workDir:   "/nonexistent/path/that/does/not/exist",
			setupCtx:  func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) },
			wantErr:   true,
			errTarget: ErrHookContractFail,
		},
		{
			name:    "empty working directory string",
			workDir: "",
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			wantErr:   true,
			errTarget: ErrHookContractFail,
		},
		{
			name:    "cancelled context",
			workDir: t.TempDir(),
			setupCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // cancel immediately
				return ctx, cancel
			},
			wantErr:   true,
			errTarget: ErrHookContractFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			contract := NewContract(tt.workDir)
			ctx, cancel := tt.setupCtx()
			defer cancel()

			err := contract.Validate(ctx)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errTarget != nil && !errors.Is(err, tt.errTarget) {
					t.Errorf("error = %v, want errors.Is(%v)", err, tt.errTarget)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestContractGuarantees(t *testing.T) {
	t.Parallel()

	contract := NewContract("/tmp")
	guarantees := contract.Guarantees()

	if len(guarantees) == 0 {
		t.Fatal("Guarantees() returned empty list")
	}

	// Check that expected guarantees are present
	expectedKeywords := []string{
		"stdin",
		"exit code",
		"timeout",
		"config",
		"working directory",
	}

	for _, keyword := range expectedKeywords {
		found := false
		for _, g := range guarantees {
			if containsSubstring(g, keyword) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing guarantee containing %q", keyword)
		}
	}
}

func TestContractNonGuarantees(t *testing.T) {
	t.Parallel()

	contract := NewContract("/tmp")
	nonGuarantees := contract.NonGuarantees()

	if len(nonGuarantees) == 0 {
		t.Fatal("NonGuarantees() returned empty list")
	}

	// Check that expected non-guarantees are present
	expectedKeywords := []string{
		"PATH",
		"shell",
		"alias",
		"Python",
	}

	for _, keyword := range expectedKeywords {
		found := false
		for _, ng := range nonGuarantees {
			if containsSubstring(ng, keyword) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing non-guarantee containing %q", keyword)
		}
	}
}

// containsSubstring checks if s contains substr (case-sensitive).
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
