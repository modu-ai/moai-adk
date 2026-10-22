package homestate

import (
	"crypto/sha256"
	"fmt"
	"maps"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/sqliteguard"
)

func TestFactoryAndProfileLeaseNativePrivacy(t *testing.T) {
	snapshot := func() map[uint64]sqliteguard.Snapshot {
		state := make(map[uint64]sqliteguard.Snapshot)
		for _, owner := range sqliteguard.Status() {
			if _, exists := state[owner.Token]; exists {
				t.Fatalf("duplicate physical owner token: %+v", owner)
			}
			state[owner.Token] = owner
		}
		return state
	}
	requireState := func(want map[uint64]sqliteguard.Snapshot) {
		t.Helper()
		if got := snapshot(); !maps.Equal(got, want) {
			t.Fatalf("physical owner state = %+v, want %+v", got, want)
		}
	}
	requireOpened := func(before map[uint64]sqliteguard.Snapshot, path string) sqliteguard.Snapshot {
		t.Helper()
		got := snapshot()
		if len(got) != len(before)+1 {
			t.Fatalf("physical owners after opening %s = %+v, want exactly one addition to %+v", path, got, before)
		}
		digest := sha256.Sum256([]byte(strings.ToLower(filepath.Clean(path))))
		for token, owner := range got {
			if _, exists := before[token]; exists {
				continue
			}
			wantOwner := sqliteguard.Snapshot{Token: token, PathDigest: fmt.Sprintf("%x", digest), Stage: "live", Pins: 3}
			if token == 0 || owner != wantOwner {
				t.Fatalf("new physical owner = %+v, want %+v", owner, wantOwner)
			}
			want := maps.Clone(before)
			want[token] = owner
			requireState(want)
			return owner
		}
		t.Fatal("opening a database did not add a physical owner")
		return sqliteguard.Snapshot{}
	}
	home := filepath.Join(t.TempDir(), "managed-home")
	t.Setenv("MOAI_HOME", home)
	if err := EnsureHomeLayout(); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.CheckDirectory(home); err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	baseline := snapshot()
	factory, err := OpenFactory(project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := factory.Close(); err != nil {
			t.Error(err)
		}
	})
	factoryOwner := requireOpened(baseline, factory.Path)
	withFactory := maps.Clone(baseline)
	withFactory[factoryOwner.Token] = factoryOwner
	leases, err := OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := leases.Close(); err != nil {
			t.Error(err)
		}
	})
	leaseOwner := requireOpened(withFactory, leases.Path)
	if leaseOwner.PathDigest == factoryOwner.PathDigest {
		t.Fatal("factory and profile lease physical owners share a path digest")
	}
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
	withLease := maps.Clone(baseline)
	withLease[leaseOwner.Token] = leaseOwner
	requireState(withLease)
	if err := leases.DB.Close(); err != nil {
		t.Fatal(err)
	}
	requireState(baseline)
}
