package ledger

import (
	"testing"
	"time"
)

// AC-AFL-018: the contract §6 defaults, pinned.
func TestLedgerDefaults_MatchContract(t *testing.T) {
	checks := []struct {
		name      string
		got, want any
	}{
		{"LeaseDuration", LeaseDuration, 15 * time.Minute},
		{"RenewEvery", RenewEvery, 5 * time.Minute},
		{"ReclaimCap", ReclaimCap, 3},
		{"SnapshotEvery", SnapshotEvery, 30 * time.Minute},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, contract says %v", c.name, c.got, c.want)
		}
	}
}
