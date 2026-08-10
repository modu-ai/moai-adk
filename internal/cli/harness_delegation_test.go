package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/delegationmap"
	"github.com/modu-ai/moai-adk/internal/harness/proposalgen"
	"github.com/modu-ai/moai-adk/internal/harness/routing"
)

// seedDelegationFixtures copies the analyzer's committed ledger fixture and map
// into a temp project root, so the CLI is exercised against the same synthetic
// data the package tests use rather than against live runtime state.
func seedDelegationFixtures(t *testing.T, root, ledgerFixture string) {
	t.Helper()
	base := filepath.Join("..", "harness", "delegationmap", "testdata")

	stateDir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	ledger, err := os.ReadFile(filepath.Join(base, ledgerFixture))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, routing.LedgerFileName), ledger, 0o644); err != nil {
		t.Fatal(err)
	}

	mapDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(mapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	mapBody, err := os.ReadFile(filepath.Join(base, "delegation_map.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(delegationmap.DefaultMapPath(root), mapBody, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestHarnessDelegationAnalyzeCmd is AC-HLA-015.
func TestHarnessDelegationAnalyzeCmd(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedDelegationFixtures(t, root, "two_kinds.jsonl")

	var out bytes.Buffer
	cmd := newHarnessRouterCmd()
	cmd.SetArgs([]string{"delegation", "analyze", "--project-root", root, "--json", "--dry-run"})
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("delegation analyze: %v", err)
	}

	var payload struct {
		Findings []struct {
			Kind       string `json:"kind"`
			Subcommand string `json:"subcommand"`
			Agent      string `json:"agent"`
		} `json:"findings"`
		Reason  string   `json:"reason"`
		DryRun  bool     `json:"dry_run"`
		Written []string `json:"written"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not parseable JSON: %v\n%s", err, out.String())
	}
	if len(payload.Findings) != 2 {
		t.Errorf("findings = %d, want 2: %s", len(payload.Findings), out.String())
	}
	if payload.Reason != "ok" {
		t.Errorf("reason = %q, want ok", payload.Reason)
	}
	if !payload.DryRun {
		t.Error("dry_run should be reported as true")
	}
	if len(payload.Written) != 0 {
		t.Errorf("a dry run must write nothing; wrote %v", payload.Written)
	}

	// The dry run must leave no directory under the proposals path at all.
	if _, err := os.Stat(proposalgen.ProposalDir(root)); !os.IsNotExist(err) {
		t.Errorf("a dry run created the proposals directory (stat err = %v)", err)
	}
}

// TestHarnessDelegationAnalyzeCmd_WritesWhenNotDryRun confirms the dry-run flag
// is what suppresses the write, not an absent write path — a dry-run assertion
// against a command that never writes would pass vacuously.
func TestHarnessDelegationAnalyzeCmd_WritesWhenNotDryRun(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedDelegationFixtures(t, root, "two_kinds.jsonl")

	var out bytes.Buffer
	cmd := newHarnessRouterCmd()
	cmd.SetArgs([]string{"delegation", "analyze", "--project-root", root, "--json"})
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("delegation analyze: %v", err)
	}

	ids, err := proposalgen.ListDraftIDs(proposalgen.ProposalDir(root))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("drafts written = %d, want 2", len(ids))
	}
	body, err := os.ReadFile(filepath.Join(proposalgen.ProposalDir(root), ids[0], "proposal.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "delegation_map:") {
		t.Errorf("draft %s does not carry a delegation_map pattern key: %s", ids[0], body)
	}
}

// TestHarnessDelegationAnalyzeCmd_GracefulNoOp confirms the CLI exits 0 with a
// machine-readable reason when there is no ledger to read.
func TestHarnessDelegationAnalyzeCmd_GracefulNoOp(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	var out bytes.Buffer
	cmd := newHarnessRouterCmd()
	cmd.SetArgs([]string{"delegation", "analyze", "--project-root", root, "--json"})
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("an absent ledger must not be an error: %v", err)
	}
	if !strings.Contains(out.String(), `"reason":"ledger-absent"`) {
		t.Errorf("expected a ledger-absent reason: %s", out.String())
	}
}
