package cli

// Nothing that was not read is promoted to a pass (the N62 invariant, applied to the identity marker itself): a marker
// file that holds no marker is a refusal and not an absence that a new marker could replace, for a ledger that has a
// marker on record and for one that has none.

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAFIdentity_AMarkerFileThatHoldsNoMarkerIsARefusalNotAnAbsence(t *testing.T) {
	afSkipWithoutSh(t)
	for name, forget := range map[string]bool{"a ledger with a marker": false, "a ledger without one": true} {
		t.Run(name, func(t *testing.T) {
			a := afN28Landed(t)
			if forget {
				afIDForget(t, a.db)
			}
			file := filepath.Join(afN38GitDir(t, a.repo), afIDMarkerFile)
			if err := os.WriteFile(file, []byte("not a marker\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			before := afSnapshot(t, a.st)
			for _, args := range [][]string{{"reap"}} {
				out, err := a.run(args...)
				if err == nil || !strings.Contains(err.Error(), "does not hold a repository marker") {
					t.Errorf("moai af %s: %v %.200s, want a refusal that says the marker file holds no marker", args[0], err, out)
				}
			}
			if after := afSnapshot(t, a.st); !reflect.DeepEqual(before, after) {
				t.Error("a refused verb changed the ledger")
			}
			if b, _ := os.ReadFile(file); string(b) != "not a marker\n" {
				t.Errorf("the damaged marker file was replaced by %q", b)
			}
			if got := afIDLedgerMarker(t, a.db); forget && got != "" {
				t.Errorf("the ledger recorded %q from a damaged file", got)
			}
		})
	}
}
