package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness/routing"
)

// writeConfig writes .moai/config/sections/<name> under root with the given body.
func writeConfig(t *testing.T, root, name, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// seedTerminalPending records a pending routing row for sessionID with terminal
// evidence so a reached finalizer would close it.
func seedTerminalPending(t *testing.T, root, sessionID string) {
	t.Helper()
	store := routing.NewStore(filepath.Join(root, ".moai", "state"))
	if err := store.Record(routing.PendingRow{
		SessionID:         sessionID,
		MatchedSubcommand: "run",
		RequestDigest:     routing.RequestDigest("seed request"),
		RequestClass:      "feature",
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendEvidence(sessionID, routing.EvidenceRef{
		Kind: routing.KindGateExit, Value: "0", Terminal: true, Ref: "go test ./...",
	}); err != nil {
		t.Fatal(err)
	}
}

func ledgerRowCount(t *testing.T, root string) int {
	t.Helper()
	rows, _, err := routing.NewReader(filepath.Join(root, ".moai", "state", routing.LedgerFileName)).Read(routing.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	return len(rows)
}

func pendingFileExists(root, sessionID string) bool {
	_, err := os.Stat(filepath.Join(root, ".moai", "state", "routing-pending-"+sessionID+".json"))
	return err == nil
}

// TestHarnessObserveStop_RoutingLedgerGated is the AC-HEV-021 HOI dual-gate test.
// It exercises finalizeRoutingLedgerOnStop across BOTH gates with explicit
// fixtures. Every active-path fixture sets hook.opt_in.enabled: true explicitly —
// relying on the default would silently test the dormant path (spec.md §D.3).
func TestHarnessObserveStop_RoutingLedgerGated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		systemYAML   string // .moai/config/sections/system.yaml (HOI gate 0)
		harnessYAML  string // .moai/config/sections/harness.yaml (learning gate 1)
		wantFinalize bool
	}{
		{
			name:         "HOI off (opt_in false) -> Stop-path dormant, no finalize",
			systemYAML:   "hook:\n  opt_in:\n    enabled: false\n",
			harnessYAML:  "learning:\n  enabled: true\n",
			wantFinalize: false,
		},
		{
			name:         "HOI absent -> Stop-path dormant (fail-CLOSED default)",
			systemYAML:   "other: 1\n",
			harnessYAML:  "learning:\n  enabled: true\n",
			wantFinalize: false,
		},
		{
			name:         "HOI on + learning false -> gate 1 blocks",
			systemYAML:   "hook:\n  opt_in:\n    enabled: true\n",
			harnessYAML:  "learning:\n  enabled: false\n",
			wantFinalize: false,
		},
		{
			name:         "HOI on + learning true -> capture active, terminal evidence finalizes",
			systemYAML:   "hook:\n  opt_in:\n    enabled: true\n",
			harnessYAML:  "learning:\n  enabled: true\n",
			wantFinalize: true,
		},
		{
			name:         "HOI on + learning absent (default true) -> finalizes",
			systemYAML:   "hook:\n  opt_in:\n    enabled: true\n",
			harnessYAML:  "other: 1\n",
			wantFinalize: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeConfig(t, root, "system.yaml", tt.systemYAML)
			writeConfig(t, root, "harness.yaml", tt.harnessYAML)
			seedTerminalPending(t, root, "sess")

			var sink bytes.Buffer
			finalizeRoutingLedgerOnStop(root, "sess", &sink)

			got := ledgerRowCount(t, root)
			if tt.wantFinalize {
				if got != 1 {
					t.Fatalf("expected finalize (1 ledger row), got %d rows", got)
				}
				if pendingFileExists(root, "sess") {
					t.Fatal("finalized pending file should be removed")
				}
			} else {
				if got != 0 {
					t.Fatalf("expected NO finalize (0 rows), got %d rows", got)
				}
				if !pendingFileExists(root, "sess") {
					t.Fatal("gated-off path must leave the pending row intact")
				}
			}
		})
	}
}

// TestLedgerVerbs_NoOutcomeOnWriteSurfaces is the Go form of AC-HEV-011: the
// record and evidence write surfaces expose NO --outcome flag, while the list
// READ surface legitimately carries an --outcome row-selection filter.
func TestLedgerVerbs_NoOutcomeOnWriteSurfaces(t *testing.T) {
	t.Parallel()
	if newHarnessLedgerRecordCmd().Flags().Lookup("outcome") != nil {
		t.Error("`ledger record` MUST NOT expose an --outcome flag (un-fakeable outcome contract)")
	}
	if newHarnessLedgerEvidenceCmd().Flags().Lookup("outcome") != nil {
		t.Error("`ledger evidence` MUST NOT expose an --outcome flag (un-fakeable outcome contract)")
	}
	if newHarnessLedgerListCmd().Flags().Lookup("outcome") == nil {
		t.Error("`ledger list` SHOULD expose an --outcome READ filter")
	}
}

// TestLedgerVerbs_RecordEvidenceList exercises the record -> evidence flow via
// the real command tree (inherited --project-root), then a direct ledger write
// + list read-back.
func TestLedgerVerbs_RecordEvidenceList(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// record via the live harness tree (inherits --project-root).
	rec := newHarnessRouterCmd()
	rec.SetArgs([]string{"ledger", "record", "--project-root", root, "--subcommand", "run", "--session", "s1", "--tier", "M"})
	rec.SetIn(strings.NewReader("implement the routing ledger feature"))
	rec.SetOut(&bytes.Buffer{})
	rec.SetErr(&bytes.Buffer{})
	if err := rec.Execute(); err != nil {
		t.Fatalf("ledger record: %v", err)
	}
	if !pendingFileExists(root, "s1") {
		t.Fatal("record should create a pending row")
	}
	// The raw request text must NOT be persisted (privacy).
	pdata, _ := os.ReadFile(filepath.Join(root, ".moai", "state", "routing-pending-s1.json"))
	if bytes.Contains(pdata, []byte("implement the routing ledger feature")) {
		t.Fatal("record leaked verbatim request text into the pending file")
	}

	// evidence append via the live tree.
	ev := newHarnessRouterCmd()
	ev.SetArgs([]string{"ledger", "evidence", "--project-root", root, "--session", "s1", "--kind", "gate_exit", "--value", "0", "--terminal", "--ref", "go test"})
	ev.SetOut(&bytes.Buffer{})
	ev.SetErr(&bytes.Buffer{})
	if err := ev.Execute(); err != nil {
		t.Fatalf("ledger evidence: %v", err)
	}

	// invalid evidence kind is rejected (closed enum).
	bad := newHarnessRouterCmd()
	bad.SetArgs([]string{"ledger", "evidence", "--project-root", root, "--session", "s1", "--kind", "success"})
	bad.SetOut(&bytes.Buffer{})
	bad.SetErr(&bytes.Buffer{})
	if err := bad.Execute(); err == nil {
		t.Fatal("a free-text evidence kind must be rejected")
	}

	// Seed one finalized ledger row directly, then list it via the CLI.
	w := routing.NewWriter(filepath.Join(root, ".moai", "state", routing.LedgerFileName))
	if err := w.Append(routing.Row{
		SchemaVersion: routing.SchemaVersion, TS: time.Now().UTC().Format(time.RFC3339),
		SessionID: "s1", MatchedSubcommand: "run", Outcome: routing.OutcomeSuccess,
		RequestDigest: "sha256:0123456789ab",
	}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	ls := newHarnessRouterCmd()
	ls.SetArgs([]string{"ledger", "list", "--project-root", root, "--subcommand", "run", "--json"})
	ls.SetOut(&out)
	ls.SetErr(&bytes.Buffer{})
	if err := ls.Execute(); err != nil {
		t.Fatalf("ledger list: %v", err)
	}
	if !strings.Contains(out.String(), `"outcome":"success"`) || !strings.Contains(out.String(), `"matched_subcommand":"run"`) {
		t.Fatalf("list --json did not emit the finalized row: %s", out.String())
	}
}

// TestLedgerRecord_LearningDisabledNoOp: record is a silent no-op when learning
// is disabled (gate 1).
func TestLedgerRecord_LearningDisabledNoOp(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeConfig(t, root, "harness.yaml", "learning:\n  enabled: false\n")

	rec := newHarnessRouterCmd()
	rec.SetArgs([]string{"ledger", "record", "--project-root", root, "--subcommand", "run", "--session", "x"})
	rec.SetIn(strings.NewReader("anything"))
	rec.SetOut(&bytes.Buffer{})
	rec.SetErr(&bytes.Buffer{})
	if err := rec.Execute(); err != nil {
		t.Fatalf("record no-op should not error: %v", err)
	}
	if pendingFileExists(root, "x") {
		t.Fatal("learning-disabled record must not create a pending row")
	}
}
