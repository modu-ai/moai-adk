package homestate

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// SPEC-FACTORY-LEASE-RENEW-001 (card t1663), milestone M2, ruling d-20261010T111725Z-3ef9
// (part b): a value of a factory_lease key that the bounds reject refuses the open, and the
// refusal names the key path, its allowed range, and the config file that holds the value.
// The file is matched by its place under the project root, so the test does not depend on
// how the root path is spelled.
func TestFactoryLeaseRefusalNamesTheConfigFile(t *testing.T) {
	root := factorySandbox(t)
	writeFactoryWorkflowConfig(t, root, "workflow:\n  factory_lease:\n    duration: 0s\n")
	_, err := OpenFactory(root)
	if !errors.Is(err, errFactoryLeaseRefused) {
		t.Fatalf("OpenFactory err = %v, want a factory lease refusal", err)
	}
	msg := err.Error()
	if !strings.Contains(msg, filepath.Join(".moai", "config", "sections", "workflow.yaml")) {
		t.Errorf("refusal does not name the config file: %v", err)
	}
	if !strings.Contains(msg, "workflow.factory_lease.duration") || !strings.Contains(msg, "greater than zero") {
		t.Errorf("refusal does not name the key and its allowed range: %v", err)
	}
}
