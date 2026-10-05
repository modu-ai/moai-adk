package quality

// Recursive language-marker detection tests (GH #1680, card t559): the heavy
// gate must find module roots below the project directory, bounded by
// config.DefaultGateMarkerScanDepth and skipping dependency/build/cache
// directories, with the detected root bound into step execution.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// writeNested creates a file at slash-relative path under dir, creating
// intermediate directories.
func writeNested(t *testing.T, dir, rel string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", p, err)
	}
	if err := os.WriteFile(p, []byte(""), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}

// TestQualityGate_detectToolchain_NestedModule verifies a Go module whose
// go.mod lives in a subdirectory is detected, with the detected root bound
// into step execution (stepDir) — the binding that keeps `go vet ./...`
// inside the module that owns it.
func TestQualityGate_detectToolchain_NestedModule(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeNested(t, dir, "apps/api/go.mod")
	writeNested(t, dir, "apps/api/main.go")

	g := NewQualityGate(&GateConfig{ProjectDir: dir})
	dt := g.detectToolchain()
	if dt == nil {
		t.Fatal("detectToolchain() should detect a go.mod in a subdirectory")
	}
	if dt.tc.markerFiles[0] != "go.mod" {
		t.Errorf("first marker = %q, want go.mod", dt.tc.markerFiles[0])
	}
	nested := filepath.Join(dir, "apps", "api")
	if dt.root != nested {
		t.Errorf("detected root = %q, want %q", dt.root, nested)
	}
	// The root binding must be live: steps execute at the module root.
	if got := g.stepDir("test"); got != nested {
		t.Errorf("stepDir after detection = %q, want %q", got, nested)
	}
}

// TestQualityGate_detectToolchain_NestedExclusions verifies the recursive
// scan never reports a module root that lives inside a dependency, build
// output, or VCS-internal directory (sourceScanSkipDirs).
func TestQualityGate_detectToolchain_NestedExclusions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		marker string // slash path of the only marker in the fixture
	}{
		{"node_modules", "node_modules/leftpad/go.mod"},
		{"vendor", "vendor/example.com/dep/go.mod"},
		{"git internals", ".git/hooks/go.mod"},
		{"dist build output", "dist/package.json"},
		{"python cache", "__pycache__/x/pyproject.toml"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			writeNested(t, dir, tc.marker)
			g := NewQualityGate(&GateConfig{ProjectDir: dir})
			if dt := g.detectToolchain(); dt != nil {
				t.Errorf("detectToolchain() = %v at %q, want nil (excluded directory)", dt.tc.markerFiles, dt.root)
			}
		})
	}
}

// TestQualityGate_detectToolchain_NestedDepthBound verifies the explicit
// depth bound: a marker exactly config.DefaultGateMarkerScanDepth levels
// below the project root is still found; one level deeper is not.
func TestQualityGate_detectToolchain_NestedDepthBound(t *testing.T) {
	t.Parallel()

	buildChain := func(t *testing.T, levels int) string {
		t.Helper()
		dir := t.TempDir()
		rel := ""
		for i := 0; i < levels; i++ {
			rel = filepath.Join(rel, "level")
		}
		writeNested(t, dir, filepath.Join(rel, "Gemfile"))
		return dir
	}

	t.Run("at the bound is detected", func(t *testing.T) {
		t.Parallel()
		dir := buildChain(t, config.DefaultGateMarkerScanDepth)
		g := NewQualityGate(&GateConfig{ProjectDir: dir})
		dt := g.detectToolchain()
		if dt == nil {
			t.Fatalf("marker at depth %d should be detected", config.DefaultGateMarkerScanDepth)
		}
		if dt.root != filepath.Join(dir, levelsPath(config.DefaultGateMarkerScanDepth)) {
			t.Errorf("detected root = %q, want the level-%d directory", dt.root, config.DefaultGateMarkerScanDepth)
		}
	})

	t.Run("beyond the bound is not", func(t *testing.T) {
		t.Parallel()
		dir := buildChain(t, config.DefaultGateMarkerScanDepth+1)
		g := NewQualityGate(&GateConfig{ProjectDir: dir})
		if dt := g.detectToolchain(); dt != nil {
			t.Errorf("marker at depth %d should not be detected, got %v at %q",
				config.DefaultGateMarkerScanDepth+1, dt.tc.markerFiles, dt.root)
		}
	})
}

// levelsPath builds a slash path of n "level" segments.
func levelsPath(n int) string {
	rel := ""
	for i := 0; i < n; i++ {
		rel = filepath.Join(rel, "level")
	}
	return rel
}

// TestQualityGate_detectToolchain_NestedTablePrecedence verifies the
// recursive scan ranks candidates by toolchains-table order, not by walk
// order: nested apps/api/go.mod (Go, earlier in the table) must win over
// nested apps/web/package.json (Node), even though WalkDir visits web after
// api alphabetically — the reverse pairing (web before api) proves the
// ranking is table order, not lexical luck.
func TestQualityGate_detectToolchain_NestedTablePrecedence(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeNested(t, dir, "apps/web/package.json")
	writeNested(t, dir, "apps/api/go.mod")

	g := NewQualityGate(&GateConfig{ProjectDir: dir})
	dt := g.detectToolchain()
	if dt == nil {
		t.Fatal("detectToolchain() should detect a nested toolchain")
	}
	if dt.tc.markerFiles[0] != "go.mod" {
		t.Errorf("table-order winner = %v, want go.mod (Go outranks Node)", dt.tc.markerFiles)
	}
	if want := filepath.Join(dir, "apps", "api"); dt.root != want {
		t.Errorf("detected root = %q, want %q", dt.root, want)
	}
}

// TestQualityGate_detectToolchain_RootLevelPrecedence verifies existing
// precedence is preserved: a root-level marker outranks any nested one, so
// today's single-project repositories behave exactly as before.
func TestQualityGate_detectToolchain_RootLevelPrecedence(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeNested(t, dir, "package.json")
	writeNested(t, dir, "apps/api/go.mod")

	g := NewQualityGate(&GateConfig{ProjectDir: dir})
	dt := g.detectToolchain()
	if dt == nil {
		t.Fatal("detectToolchain() should detect the root-level project")
	}
	if dt.tc.markerFiles[0] != "package.json" {
		t.Errorf("root-level winner = %v, want package.json (root outranks nested)", dt.tc.markerFiles)
	}
	if dt.root != dir {
		t.Errorf("detected root = %q, want the project root %q", dt.root, dir)
	}
}

// TestQualityGate_detectToolchain_NestedGlobMarker verifies glob-pattern
// markers (C#/.NET's *.csproj) are found by the recursive scan too.
func TestQualityGate_detectToolchain_NestedGlobMarker(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeNested(t, dir, "src/dotnet/MyApp.csproj")

	g := NewQualityGate(&GateConfig{ProjectDir: dir})
	dt := g.detectToolchain()
	if dt == nil {
		t.Fatal("detectToolchain() should detect a nested .csproj")
	}
	if dt.tc.markerFiles[0] != "*.csproj" {
		t.Errorf("first marker = %q, want *.csproj", dt.tc.markerFiles[0])
	}
}

// TestQualityGate_detectToolchain_NestedPythonRunner verifies the Python
// runner resolution scopes to the nested module root: the uv.lock beside the
// nested pyproject.toml — not the project top — decides the runner.
func TestQualityGate_detectToolchain_NestedPythonRunner(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeNested(t, dir, "services/worker/pyproject.toml")
	writeNested(t, dir, "services/worker/uv.lock")

	g := NewQualityGate(&GateConfig{ProjectDir: dir})
	dt := g.detectToolchain()
	if dt == nil {
		t.Fatal("detectToolchain() should detect the nested pyproject.toml")
	}
	if dt.tc.testStep == nil || dt.tc.testStep.binary != "uv" {
		t.Errorf("nested pytest runner = %v, want uv (resolved at the module root)", dt.tc.testStep)
	}
}

// TestQualityGate_detectToolchain_NestedConfigScope verifies step config-file
// detection scopes to the bound module root: the eslint config inside the
// nested Node module is found, and one absent there but present at the top is
// not.
func TestQualityGate_detectToolchain_NestedConfigScope(t *testing.T) {
	t.Parallel()

	t.Run("config inside the module root is found", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeNested(t, dir, "apps/web/package.json")
		writeNested(t, dir, "apps/web/.eslintrc.json")
		g := NewQualityGate(&GateConfig{ProjectDir: dir})
		if g.detectToolchain() == nil {
			t.Fatal("detectToolchain() should detect the nested package.json")
		}
		if !g.anyConfigFileExists([]string{".eslintrc.json"}) {
			t.Error("anyConfigFileExists should find .eslintrc.json at the module root")
		}
	})

	t.Run("config outside the module root is not", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeNested(t, dir, "apps/web/package.json")
		writeNested(t, dir, ".eslintrc.json") // top-level only
		g := NewQualityGate(&GateConfig{ProjectDir: dir})
		if g.detectToolchain() == nil {
			t.Fatal("detectToolchain() should detect the nested package.json")
		}
		if g.anyConfigFileExists([]string{".eslintrc.json"}) {
			t.Error("anyConfigFileExists should not see a top-level config from the module root scope")
		}
	})
}
