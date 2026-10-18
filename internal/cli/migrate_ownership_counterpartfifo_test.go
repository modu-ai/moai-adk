//go:build !windows

package cli

// migrate_ownership_counterpartfifo_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001
// (card t1865, sync-audit finding F3): the verdict function reads the USER-root
// counterpart of a registered managed file (to learn whether the counterpart is
// installed, hash-equal and current), and it did so without asking what kind of
// file the counterpart was. A FIFO planted at that path made os.ReadFile wait for
// a writer, so `moai update` hung — and the card-added state recording even did
// it for a profile-mapped project, whose copy the profile gate keeps anyway.
//
// Expected behaviour, pinned here: a counterpart that is not a regular file is
// never opened and counts as unconfirmed. The project file stays; nothing is
// removed; the migration returns. The hang is detected with a bounded wait in a
// goroutine, exactly as in migrate_ownership_nonregular_test.go (the wait is not
// an assertion about speed), and the FIFO is poked until the blocked read ends so
// nothing outlives the temporary tree.

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// counterpartFifo replaces the user-root counterpart of rel (installed by the
// fixture) with a FIFO and returns its absolute path.
func counterpartFifo(t *testing.T, home, rel string) string {
	t.Helper()
	slug, rest, ok := userCounterpartLocation(rel)
	if !ok {
		t.Fatalf("%s has no user-root counterpart mapping", rel)
	}
	path := filepath.Join(userassets.RootBySlugDir(home, slug), filepath.FromSlash(rest))
	if err := os.Remove(path); err != nil {
		t.Fatalf("the fixture's counterpart %s is missing: %v", path, err)
	}
	mkfifo(t, path)
	return path
}

func TestCounterpartFifoIsNotOpened(t *testing.T) {
	// Verdict level: the profile-mapped and the ordinary project, both over a
	// registered managed file whose user counterpart is a FIFO.
	for _, profileMapped := range []bool{true, false} {
		name := "OrdinaryProject"
		if profileMapped {
			name = "ProfileMappedProject"
		}
		t.Run(name, func(t *testing.T) {
			root, home := ownershipFixtureBase(t)
			fifo := counterpartFifo(t, home, migIdenticalSkill)
			um, err := userassets.Load(userassets.ManifestPath(home))
			if err != nil {
				t.Fatal(err)
			}
			embedded, err := template.EmbeddedTemplates()
			if err != nil {
				t.Fatal(err)
			}
			cat, err := template.LoadEmbeddedCatalog()
			if err != nil {
				t.Fatal(err)
			}
			data := ownEmbeddedBytes(t, migIdenticalSkill)
			rootResolved, err := filepath.EvalSymlinks(root)
			if err != nil {
				t.Fatal(err)
			}
			env := ownershipEnv{
				entry: func(string) (*manifest.FileEntry, bool) {
					return &manifest.FileEntry{Provenance: manifest.TemplateManaged, TemplateHash: manifest.HashBytes(data)}, true
				},
				evidence: fakeProvenance{},
				counterpartConfirmed: func(rel string) bool {
					return counterpartConfirmedCaseAware(um, cat, embedded, home, rel)
				},
				counterpartStatus: func(rel string) counterpartStatus {
					return userCounterpartStatus(um, cat, embedded, home, rel)
				},
				notInCatalog:   func(rel string) bool { return skillNotInCatalog(cat, rel) },
				profileMapped:  profileMapped,
				removalEnabled: true,
				rootResolved:   rootResolved,
			}
			done := make(chan ownershipVerdict, 1)
			go func() {
				done <- classifyOwnership(env, migIdenticalSkill, filepath.Join(root, filepath.FromSlash(migIdenticalSkill)))
			}()
			select {
			case v := <-done:
				if v.removalEligible {
					t.Errorf("a FIFO counterpart confirmed a removal (class=%q keep=%q)", v.class, v.keep)
				}
				want := keepCounterpartUnconfirmed
				if profileMapped {
					want = keepProfileReach
				}
				if v.keep != want {
					t.Errorf("keep=%q, want %q", v.keep, want)
				}
				if v.counterpart != counterpartStatusUnconfirmed {
					t.Errorf("counterpart state %q, want %q", v.counterpart, counterpartStatusUnconfirmed)
				}
			case <-time.After(30 * time.Second):
				releaseBlockedFifoRead(done, fifo)
				t.Fatalf("the verdict did not return: it is blocked reading the user-root counterpart FIFO %s", fifo)
			}
		})
	}

	// Flow level: the real walk keeps the project copy and returns.
	t.Run("MigrationReturnsAndKeepsTheProjectCopy", func(t *testing.T) {
		root, home := ownershipFixtureBase(t)
		fifo := counterpartFifo(t, home, migIdenticalSkill)
		o := runNonRegularMigration(t, nonRegularWorld{root: root, home: home, fifos: []string{fifo}})
		if o.hung {
			t.Fatalf("the migration did not return: it is blocked reading the user-root counterpart FIFO %s", fifo)
		}
		if o.err != nil {
			t.Fatalf("migration: %v", o.err)
		}
		if len(o.removed) != 0 {
			t.Errorf("removed %v although the counterpart is a FIFO", o.removed)
		}
		r, ok := findRemaining(o.rc, migIdenticalSkill)
		switch {
		case !ok:
			t.Errorf("%s has no row in the remaining list", migIdenticalSkill)
		case r.Class != ownershipManaged || r.Reason != keepCounterpartUnconfirmed || r.Counterpart != counterpartStatusUnconfirmed:
			t.Errorf("%s: row class=%q reason=%q counterpart=%q, want managed/%q/%q", migIdenticalSkill, r.Class, r.Reason, r.Counterpart, keepCounterpartUnconfirmed, counterpartStatusUnconfirmed)
		}
		if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(migIdenticalSkill))); err != nil {
			t.Errorf("the project copy is gone: %v", err)
		}
		if info, err := os.Lstat(fifo); err != nil || info.Mode()&os.ModeNamedPipe == 0 {
			t.Errorf("the counterpart is no longer a FIFO (%v)", err)
		}
	})
}

// releaseBlockedFifoRead ends a read blocked on fifo and waits for the goroutine
// that held it, so nothing outlives the temporary tree.
func releaseBlockedFifoRead(done <-chan ownershipVerdict, fifos ...string) {
	deadline := time.After(30 * time.Second)
	for {
		pokeFifos(fifos)
		select {
		case <-done:
			return
		case <-deadline:
			return
		case <-time.After(20 * time.Millisecond):
		}
	}
}
