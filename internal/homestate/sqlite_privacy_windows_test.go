package homestate

import (
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/sqliteguard"
)

func TestFactoryAndProfileLeaseNativePrivacy(t *testing.T) {
	home := filepath.Join(t.TempDir(), "managed-home")
	t.Setenv("MOAI_HOME", home)
	if err := EnsureHomeLayout(); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.CheckDirectory(home); err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	factory, err := OpenFactory(project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := factory.Close(); err != nil {
			t.Error(err)
		}
	})
	leases, err := OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := leases.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, path := range []string{factory.Path, leases.Path} {
		if err := privatefile.CheckDirectory(filepath.Dir(path)); err != nil {
			t.Errorf("database parent privacy: %v", err)
		}
		for _, suffix := range []string{"", "-wal", "-shm"} {
			if err := privatefile.Check(path + suffix); err != nil {
				t.Errorf("database/sidecar native privacy: %v", err)
			}
		}
	}
	// The public *sql.DB Close path must reach each physical owner too.
	if err := factory.DB.Close(); err != nil {
		t.Fatal(err)
	}
	if err := leases.DB.Close(); err != nil {
		t.Fatal(err)
	}
	for _, s := range sqliteguard.Status() {
		if s.Faulted {
			t.Errorf("native lifetime fault: %+v", s)
		}
	}
}
