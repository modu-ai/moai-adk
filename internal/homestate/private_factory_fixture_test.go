package homestate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/sqliteguard"
)

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
