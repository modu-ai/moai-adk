package homestate

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// EffectiveFactoryLease is the value that the factory status surface reports
// (SPEC-FACTORY-LEASE-RENEW-001, card t1663, REQ-FLR-009). It must report the values
// the factory opens with, without creating a database.
func TestEffectiveFactoryLeaseReportsTheOpenValues(t *testing.T) {
	root := factorySandbox(t)
	duration, renew, err := EffectiveFactoryLease(root)
	if err != nil {
		t.Fatalf("EffectiveFactoryLease without configuration: %v", err)
	}
	if duration != config.DefaultFactoryLeaseDuration || renew != config.DefaultFactoryLeaseRenewInterval {
		t.Fatalf("defaults = %s and %s, want %s and %s", duration, renew,
			config.DefaultFactoryLeaseDuration, config.DefaultFactoryLeaseRenewInterval)
	}

	writeFactoryWorkflowConfig(t, root, "workflow:\n  factory_lease:\n    duration: 20m\n    renew_interval: 4m\n")
	duration, renew, err = EffectiveFactoryLease(root)
	if err != nil {
		t.Fatalf("EffectiveFactoryLease with configuration: %v", err)
	}
	if duration != 20*time.Minute || renew != 4*time.Minute {
		t.Fatalf("configured = %s and %s, want 20m0s and 4m0s", duration, renew)
	}

	refused := factorySandbox(t)
	writeFactoryWorkflowConfig(t, refused, "workflow:\n  factory_lease:\n    duration: 0s\n")
	if _, _, err := EffectiveFactoryLease(refused); !errors.Is(err, errFactoryLeaseRefused) {
		t.Fatalf("EffectiveFactoryLease on a refused configuration: err = %v, want a factory lease refusal", err)
	}
	path, err := FactoryDBPath(refused)
	if err != nil {
		t.Fatalf("FactoryDBPath: %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("EffectiveFactoryLease left a database at %s (stat error %v)", path, statErr)
	}
}
