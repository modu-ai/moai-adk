package harness

import (
	"os"
	"path/filepath"
	"testing"
)

// TestHasWebFrontend verifies the web project detection logic
func TestHasWebFrontend(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) (root string, cleanup func())
		expected bool
	}{
		{
			name: "React project detected via package.json",
			setup: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				pkgJSON := filepath.Join(tmpDir, "package.json")
				err := os.WriteFile(pkgJSON, []byte(`{
  "name": "test-app",
  "dependencies": {
    "react": "^18.0.0",
    "react-dom": "^18.0.0"
  }
}`), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return tmpDir, func() {}
			},
			expected: true,
		},
		{
			name: "Vue project detected via package.json",
			setup: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				pkgJSON := filepath.Join(tmpDir, "package.json")
				err := os.WriteFile(pkgJSON, []byte(`{
  "name": "test-app",
  "dependencies": {
    "vue": "^3.0.0"
  }
}`), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return tmpDir, func() {}
			},
			expected: true,
		},
		{
			name: "CLI tool without web frontend",
			setup: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				// No package.json, no index.html
				return tmpDir, func() {}
			},
			expected: false,
		},
		{
			name: "index.html detected in project root",
			setup: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				indexHTML := filepath.Join(tmpDir, "index.html")
				err := os.WriteFile(indexHTML, []byte(`<html><body>Test</body></html>`), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return tmpDir, func() {}
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, cleanup := tt.setup(t)
			defer cleanup()

			result := HasWebFrontend(root)
			if result != tt.expected {
				t.Errorf("HasWebFrontend() = %v, want %v", result, tt.expected)
			}
		})
	}
}
