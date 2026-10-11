package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// SPEC-FACTORY-LEASE-RENEW-001 (card t1663), milestone M1, AC-FLR-009 (its M1 part):
// the factory status JSON reports the effective values of
// workflow.factory_lease.duration as lease_duration and of
// workflow.factory_lease.renew_interval as renew_interval (REQ-FLR-009).
func TestFactoryStatusReportsLeaseKeys(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		wantLease string
		wantRenew string
	}{
		{"defaults", "", "15m0s", "5m0s"},
		{"configured", "workflow:\n  factory_lease:\n    duration: 20m\n    renew_interval: 4m\n", "20m0s", "4m0s"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _ = fcFixture(t)
			root := factoryCardRoot()
			if c.body != "" {
				dir := filepath.Join(root, ".moai", "config", "sections")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("mkdir config sections: %v", err)
				}
				if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(c.body), 0o644); err != nil {
					t.Fatalf("write workflow.yaml: %v", err)
				}
			}
			out, _, err := runFactory(t, "status", "--json")
			if err != nil {
				t.Fatalf("factory status --json: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal([]byte(out), &got); err != nil {
				t.Fatalf("factory status --json output is not JSON: %v\n%s", err, out)
			}
			if got["lease_duration"] != c.wantLease {
				t.Fatalf("lease_duration = %v, want %q", got["lease_duration"], c.wantLease)
			}
			if got["renew_interval"] != c.wantRenew {
				t.Fatalf("renew_interval = %v, want %q", got["renew_interval"], c.wantRenew)
			}
		})
	}
}
