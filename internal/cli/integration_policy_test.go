package cli

// integration_policy_test.go — REQ-MWQ-012's verb-layer tests (card t1479):
// the lane-role refusal leaves the policy record unchanged, and a hold
// without a reason is refused.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

func TestPolicyHoldRefusedForLaneRole(t *testing.T) {
	// REQ-MWQ-012: a session that declares the lane role cannot write the
	// policy, and the record is unchanged.
	root := t.TempDir()
	t.Setenv(config.EnvFactoryRole, "lane-2")
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	cmd := newIntegrationPolicyHoldCmd()
	cmd.SetArgs([]string{"--reason", "release-cut"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("a lane-role session must not write the policy")
	}
	policy, err := factory.ReadIntegrationWindowPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Policy != factory.PolicyOpen {
		t.Fatalf("a refused hold must leave the policy open, got %q", policy.Policy)
	}
}

func TestPolicyHoldRequiresReason(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	cmd := newIntegrationPolicyHoldCmd()
	if err := cmd.Execute(); err == nil {
		t.Fatalf("a hold without --reason must be refused")
	}
}

func TestPolicyOpenWritesOpenRecord(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	t.Setenv("MOAI_FACTORY_ROLE", "") // the lane env leaks into tests; a leader session makes no claim
	if err := factory.WriteIntegrationWindowPolicy(root, factory.IntegrationWindowPolicy{Policy: factory.PolicyHold, Reason: "release-cut"}); err != nil {
		t.Fatal(err)
	}
	cmd := newIntegrationPolicyOpenCmd()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("policy open must succeed outside a lane role: %v", err)
	}
	policy, err := factory.ReadIntegrationWindowPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Policy != factory.PolicyOpen {
		t.Fatalf("policy open must read open, got %q", policy.Policy)
	}
}

func TestPolicyRecordPathBesideWindowRecord(t *testing.T) {
	// The policy file sits in the same state directory as the window record
	// and does not collide with it.
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	if err := newIntegrationPolicyHoldCmd().Execute(); err == nil {
		t.Fatalf("hold without reason is refused")
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "state", factory.IntegrationWindowPolicyFileName)); !os.IsNotExist(err) {
		t.Fatalf("a refused hold must not leave a policy record: %v", err)
	}
}
