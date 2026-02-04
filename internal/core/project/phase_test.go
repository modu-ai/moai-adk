package project

import (
	"context"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/foundation"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

func TestPhaseExecutor_Execute_NonInteractive(t *testing.T) {
	root := t.TempDir()

	// Set up a Go project structure
	writeFile(t, root, "go.mod", "module test\ngo 1.22\n")
	writeFile(t, root, "main.go", "package main\nfunc main() {}\n")
	writeFile(t, root, "internal/app.go", "package internal\n")

	pe := newTestPhaseExecutor()

	opts := InitOptions{
		ProjectRoot:    root,
		NonInteractive: true,
	}

	result, err := pe.Execute(context.Background(), opts)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify defaults were applied
	if result.DevelopmentMode == "" {
		t.Error("expected DevelopmentMode to be set")
	}

	if len(result.CreatedFiles) == 0 {
		t.Error("expected at least one created file")
	}

	if len(result.CreatedDirs) == 0 {
		t.Error("expected at least one created directory")
	}
}

func TestPhaseExecutor_Execute_ExplicitDevelopmentMode(t *testing.T) {
	root := t.TempDir()

	pe := newTestPhaseExecutor()

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "my-app",
		Language:        "Go",
		Framework:       "none",
		UserName:        "testuser",
		ConvLang:        "en",
		DevelopmentMode: "tdd",
		NonInteractive:  true,
	}

	result, err := pe.Execute(context.Background(), opts)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.DevelopmentMode != "tdd" {
		t.Errorf("DevelopmentMode = %q, want %q", result.DevelopmentMode, "tdd")
	}
}

func TestPhaseExecutor_Execute_InvalidDevelopmentMode(t *testing.T) {
	root := t.TempDir()

	pe := newTestPhaseExecutor()

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "my-app",
		Language:        "Go",
		DevelopmentMode: "invalid-mode",
		NonInteractive:  true,
	}

	_, err := pe.Execute(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for invalid development mode")
	}
}

func TestPhaseExecutor_Execute_ExistingProjectWithoutForce(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, ".moai/config/sections")

	pe := newTestPhaseExecutor()

	opts := InitOptions{
		ProjectRoot:    root,
		NonInteractive: true,
	}

	_, err := pe.Execute(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for existing project without --force")
	}
}

func TestPhaseExecutor_Execute_ExistingProjectWithForce(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, ".moai/config/sections")
	writeFile(t, root, ".moai/config/sections/user.yaml", "user:\n  name: old\n")

	pe := newTestPhaseExecutor()

	opts := InitOptions{
		ProjectRoot:    root,
		ProjectName:    "new-project",
		Language:       "Go",
		UserName:       "testuser",
		ConvLang:       "en",
		Force:          true,
		NonInteractive: true,
	}

	result, err := pe.Execute(context.Background(), opts)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result.CreatedFiles) == 0 {
		t.Error("expected files to be created after force reinit")
	}

	// Verify new .moai/ was created
	assertFileExists(t, root+"/.moai/config/sections/user.yaml")
}

func TestPhaseExecutor_Execute_ContextCancellation(t *testing.T) {
	root := t.TempDir()
	pe := newTestPhaseExecutor()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := InitOptions{
		ProjectRoot:    root,
		NonInteractive: true,
	}

	_, err := pe.Execute(ctx, opts)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestPhaseExecutor_Execute_AutoDetectsMethodology(t *testing.T) {
	root := t.TempDir()

	// Create brownfield project with no tests -> should recommend DDD
	writeFile(t, root, "go.mod", "module test\ngo 1.22\n")
	for _, f := range []string{"a.go", "b.go", "c.go", "d.go", "e.go"} {
		writeFile(t, root, "pkg/"+f, "package pkg\n")
	}

	pe := newTestPhaseExecutor()

	opts := InitOptions{
		ProjectRoot:    root,
		NonInteractive: true,
		// DevelopmentMode intentionally left empty for auto-detection
	}

	result, err := pe.Execute(context.Background(), opts)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should auto-detect to "ddd" for brownfield with no tests
	if result.DevelopmentMode != "ddd" {
		t.Errorf("auto-detected DevelopmentMode = %q, want %q", result.DevelopmentMode, "ddd")
	}
}

func TestApplyDetectedDefaults(t *testing.T) {
	tests := []struct {
		name       string
		opts       InitOptions
		languages  []Language
		frameworks []Framework
		wantName   string
		wantLang   string
		wantFW     string
		wantConv   string
	}{
		{
			name:     "all defaults applied",
			opts:     InitOptions{ProjectRoot: "/tmp/my-project"},
			wantName: "my-project",
			wantLang: "Go", // fallback
			wantFW:   "none",
			wantConv: "en",
		},
		{
			name:      "detected language used",
			opts:      InitOptions{ProjectRoot: "/tmp/test"},
			languages: []Language{{Name: "Python", Confidence: 0.8}},
			wantName:  "test",
			wantLang:  "Python",
			wantFW:    "none",
			wantConv:  "en",
		},
		{
			name:       "detected framework used",
			opts:       InitOptions{ProjectRoot: "/tmp/test"},
			languages:  []Language{{Name: "Go", Confidence: 1.0}},
			frameworks: []Framework{{Name: "Gin"}},
			wantName:   "test",
			wantLang:   "Go",
			wantFW:     "Gin",
			wantConv:   "en",
		},
		{
			name:      "explicit values not overridden",
			opts:      InitOptions{ProjectRoot: "/tmp/test", ProjectName: "custom", Language: "Rust", Framework: "Axum", ConvLang: "ko"},
			languages: []Language{{Name: "Go", Confidence: 1.0}},
			wantName:  "custom",
			wantLang:  "Rust",
			wantFW:    "Axum",
			wantConv:  "ko",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyDetectedDefaults(tt.opts, tt.languages, tt.frameworks, "")

			if got.ProjectName != tt.wantName {
				t.Errorf("ProjectName = %q, want %q", got.ProjectName, tt.wantName)
			}
			if got.Language != tt.wantLang {
				t.Errorf("Language = %q, want %q", got.Language, tt.wantLang)
			}
			if got.Framework != tt.wantFW {
				t.Errorf("Framework = %q, want %q", got.Framework, tt.wantFW)
			}
			if got.ConvLang != tt.wantConv {
				t.Errorf("ConvLang = %q, want %q", got.ConvLang, tt.wantConv)
			}
		})
	}
}

func TestOsUserName(t *testing.T) {
	// Save original env vars
	origUser := os.Getenv("USER")
	origUsername := os.Getenv("USERNAME")

	t.Cleanup(func() {
		_ = os.Setenv("USER", origUser)
		_ = os.Setenv("USERNAME", origUsername)
	})

	t.Run("uses USER env var", func(t *testing.T) {
		_ = os.Setenv("USER", "testuser")
		_ = os.Setenv("USERNAME", "other")
		got := osUserName()
		if got != "testuser" {
			t.Errorf("osUserName() = %q, want %q", got, "testuser")
		}
	})

	t.Run("falls back to USERNAME", func(t *testing.T) {
		_ = os.Unsetenv("USER")
		_ = os.Setenv("USERNAME", "winuser")
		got := osUserName()
		if got != "winuser" {
			t.Errorf("osUserName() = %q, want %q", got, "winuser")
		}
	})

	t.Run("falls back to default", func(t *testing.T) {
		_ = os.Unsetenv("USER")
		_ = os.Unsetenv("USERNAME")
		got := osUserName()
		if got != "user" {
			t.Errorf("osUserName() = %q, want %q", got, "user")
		}
	})
}

func TestPhaseExecutor_Execute_MethodologyFallback(t *testing.T) {
	root := t.TempDir()

	// Empty root with no files should fallback to greenfield default
	pe := newTestPhaseExecutor()

	opts := InitOptions{
		ProjectRoot:    root,
		NonInteractive: true,
		// DevelopmentMode left empty - should be auto-detected
	}

	result, err := pe.Execute(context.Background(), opts)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should get a development mode assigned
	if result.DevelopmentMode == "" {
		t.Error("expected DevelopmentMode to be set via auto-detection")
	}
}

// newTestPhaseExecutor creates a PhaseExecutor with real implementations.
func newTestPhaseExecutor() *PhaseExecutor {
	registry := foundation.DefaultRegistry
	detector := NewDetector(registry, nil)
	methDetector := NewMethodologyDetector(nil)
	validator := NewValidator(nil)
	mgr := manifest.NewManager()
	initializer := NewInitializer(nil, mgr, nil)

	return NewPhaseExecutor(detector, methDetector, validator, initializer, nil)
}
