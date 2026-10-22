package homestate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/sqliteguard"
)

// An uncertain native open deliberately keeps its pins until process exit.
// The parent owns cleanup so Windows never removes a still-pinned fixture.
func withRetainedFactoryFixture(t *testing.T, body func(root string)) {
	t.Helper()
	if runtime.GOOS != "windows" {
		body("")
		return
	}
	const testEnv = "MOAI_HOMESTATE_RETAINED_TEST"
	const rootEnv = "MOAI_HOMESTATE_RETAINED_ROOT"
	name, root := os.Getenv(testEnv), os.Getenv(rootEnv)
	if name != "" || root != "" {
		if name != t.Name() || !filepath.IsAbs(root) || filepath.Clean(root) != root {
			t.Fatalf("invalid retained fixture child: test=%q root=%q", name, root)
		}
		info, err := os.Lstat(root)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("retained fixture root must be an existing directory: %v", err)
		}
		body(root)
		fmt.Fprintln(os.Stdout, "MOAI_RETAINED_FIXTURE_COMPLETE:"+name+":"+root)
		return
	}
	root = t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), executable, "-test.run=^"+regexp.QuoteMeta(t.Name())+"$", "-test.v")
	cmd.Env = append(os.Environ(), testEnv+"="+t.Name(), rootEnv+"="+root)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("retained fixture child failed: %v\n%s", err, output)
	}
	sentinel := "MOAI_RETAINED_FIXTURE_COMPLETE:" + t.Name() + ":" + root + "\n"
	if strings.Count(string(output), sentinel) != 1 {
		t.Fatalf("retained fixture child did not complete the test body exactly once:\n%s", output)
	}
	t.Logf("retained fixture child completed:\n%s", output)
}

// A managed child is created with an explicit owner policy. The Windows
// token's default owner need not equal its current user SID; the unrelated
// TempDir parent is deliberately left untouched.
func privateFactoryDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "private-factory")
	if err := sqliteguard.EnsureDirectory(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

// Raw historical-schema fixtures need the same before-bytes privacy
// precondition without using the production current-schema initializer.
func preparePrivateFactorySeed(t *testing.T, path string) {
	t.Helper()
	if err := sqliteguard.EnsureDirectory(filepath.Dir(path)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("expected new seed path, stat=%v", err)
	}
	if err := privatefile.WriteFile(path, nil); err != nil {
		t.Fatal(err)
	}
}
