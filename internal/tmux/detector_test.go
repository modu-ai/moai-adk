package tmux

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestSystemDetector_IsAvailable_WithRunner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		runFunc RunFunc
		want    bool
	}{
		{
			name: "tmux found",
			runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
				return "tmux 3.4", nil
			},
			want: true,
		},
		{
			name: "tmux not found",
			runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
				return "", fmt.Errorf("executable not found")
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := NewDetector(WithRunFunc(tt.runFunc))
			got := d.IsAvailable()
			if got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSystemDetector_Version(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		runFunc RunFunc
		want    string
		wantErr bool
	}{
		{
			name: "standard version output",
			runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
				return "tmux 3.4", nil
			},
			want: "3.4",
		},
		{
			name: "version with patch",
			runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
				return "tmux 3.3a", nil
			},
			want: "3.3a",
		},
		{
			name: "tmux not installed",
			runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
				return "", fmt.Errorf("not found")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := NewDetector(WithRunFunc(tt.runFunc))
			got, err := d.Version()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Version() = %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Version() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSystemDetector_Version_EmptyOutput(t *testing.T) {
	t.Parallel()

	d := NewDetector(WithRunFunc(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", nil
	}))

	_, err := d.Version()
	if err == nil {
		t.Error("expected error for empty version output")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("error should mention parsing, got: %v", err)
	}
}

func TestDefaultRun_KnownBinary(t *testing.T) {
	t.Parallel()

	// Test defaultRun with "echo", which should be available on all systems.
	output, err := defaultRun(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "hello") {
		t.Errorf("output = %q, want to contain 'hello'", output)
	}
}

func TestDefaultRun_NotFound(t *testing.T) {
	t.Parallel()

	_, err := defaultRun(context.Background(), "nonexistent_binary_xyz_12345")
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, should mention 'not found'", err)
	}
}

func TestDefaultRun_CommandFails(t *testing.T) {
	t.Parallel()

	// "false" is a standard Unix command that always exits with code 1.
	_, err := defaultRun(context.Background(), "false")
	if err == nil {
		t.Fatal("expected error for failing command")
	}
}

func TestNewDetector_DefaultRunFunc(t *testing.T) {
	t.Parallel()

	// Verify that NewDetector without options creates a valid detector.
	d := NewDetector()
	if d.run == nil {
		t.Fatal("run function should not be nil")
	}
}
