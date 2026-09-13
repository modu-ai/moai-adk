package homestate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestTempDiscriminantParity asserts the homestate temp guard and kanban's
// TempOriginReason agree on every anchor class: a base the queue resolver
// classifies temporary must get a project-local factory database, and a base
// it does not must get the home one. The guards live in sibling packages that
// cannot import each other (kanban imports homestate); without this test the
// two copies of the anchor set drift and one resolver silently substitutes
// the other, splitting a project's state across two trees.
func TestTempDiscriminantParity(t *testing.T) {
	t.Setenv("MOAI_HOME", "")

	// t.TempDir sits under os.TempDir(), the anchor both guards already
	// shared. The /tmp base exercises the anchor the homestate guard alone
	// used to miss on a machine whose TMPDIR points at a per-user directory
	// (macOS /var/folders), and /var/tmp the deliberate exclusion both sides
	// must keep.
	tmpBase := t.TempDir()

	varUnderTmp := ""
	if _, err := os.Stat("/tmp"); err == nil {
		dir, err := os.MkdirTemp("/tmp", "homestate-parity-")
		if err != nil {
			t.Fatalf("create /tmp base: %v", err)
		}
		varUnderTmp = dir
		t.Cleanup(func() { os.RemoveAll(dir) })
	} else {
		t.Log("/tmp unavailable; skipping the /tmp anchor case")
	}

	varUnderVarTmp := ""
	if _, err := os.Stat("/var/tmp"); err == nil {
		dir, err := os.MkdirTemp("/var/tmp", "homestate-parity-")
		if err != nil {
			t.Fatalf("create /var/tmp base: %v", err)
		}
		varUnderVarTmp = dir
		t.Cleanup(func() { os.RemoveAll(dir) })
	}

	cases := []struct {
		name    string
		base    string
		wantTmp bool
	}{
		{name: "under os.TempDir", base: tmpBase, wantTmp: true},
		{name: "non-temp working directory", base: ".", wantTmp: false},
	}
	if varUnderTmp != "" {
		cases = append(cases, struct {
			name    string
			base    string
			wantTmp bool
		}{"under /tmp", varUnderTmp, true})
	}
	if varUnderVarTmp != "" {
		cases = append(cases, struct {
			name    string
			base    string
			wantTmp bool
		}{"under /var/tmp (excluded)", varUnderVarTmp, false})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, kanbanTemp := kanban.TempOriginReason(tc.base)
			if kanbanTemp != tc.wantTmp {
				t.Fatalf("premise: kanban.TempOriginReason(%s) = %v, want %v", tc.base, kanbanTemp, tc.wantTmp)
			}
			fdb, err := homestate.FactoryDBPath(tc.base)
			if err != nil {
				t.Fatalf("FactoryDBPath(%s): %v", tc.base, err)
			}
			canonical := homestate.CanonicalProjectRoot(tc.base)
			local := strings.HasPrefix(fdb, canonical+string(filepath.Separator))
			if local != tc.wantTmp {
				t.Fatalf("factory db locality disagrees with the queue discriminant: base=%s fdb=%s canonical=%s local=%v wantTmp=%v", tc.base, fdb, canonical, local, tc.wantTmp)
			}
		})
	}
}
