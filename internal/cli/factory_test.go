package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// clearFactoryTestEnv isolates the factory signal variables from this test
// binary's ambient environment, on the same t.Setenv-restore contract as
// clearKanbanLauncherEnv (a developer running tests inside a factory session
// carries MOAI_FACTORY_* in the ambient env; the branches under test are
// unconditional on them).
func clearFactoryTestEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		config.EnvMoaiFactoryWorkers,
		config.EnvMoaiFactoryWorker,
		config.EnvMoaiKanban,
		config.EnvMoaiKanbanID,
		config.EnvMoaiKanbanLabel,
		config.EnvMoaiKanbanLeadAddr,
	} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
}

// TestParseFactoryFlag covers the four accepted forms, the required-count
// error paths, and the `--` discipline shared with parseKanbanFlag.
func TestParseFactoryFlag(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		args      []string
		wantN     int
		wantOn    bool
		wantRest  []string
		wantErr   bool
		errMarker string
	}{
		{name: "short with value", args: []string{"-f", "4"}, wantN: 4, wantOn: true},
		{name: "short equals", args: []string{"-f=3"}, wantN: 3, wantOn: true},
		{name: "long with value", args: []string{"--factory", "12"}, wantN: 12, wantOn: true},
		{name: "long equals", args: []string{"--factory=1"}, wantN: 1, wantOn: true},
		{name: "absent", args: []string{"-p", "work"}, wantOn: false, wantRest: []string{"-p", "work"}},
		{name: "name value survives", args: []string{"-f", "4", "--name", "worker-2"}, wantN: 4, wantOn: true, wantRest: []string{"--name", "worker-2"}},
		{name: "other flags preserved", args: []string{"-b", "--factory", "2", "-w"}, wantN: 2, wantOn: true, wantRest: []string{"-b", "-w"}},
		{name: "missing count", args: []string{"-f"}, wantErr: true, errMarker: "requires a worker count"},
		{name: "count is a flag", args: []string{"-f", "-b"}, wantErr: true, errMarker: "requires a worker count"},
		{name: "non-numeric count", args: []string{"--factory", "many"}, wantErr: true, errMarker: "requires a worker count"},
		{name: "zero count", args: []string{"-f", "0"}, wantErr: true, errMarker: "requires a worker count"},
		{name: "negative count", args: []string{"--factory=-2"}, wantErr: true, errMarker: "requires a worker count"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			gotN, gotOn, gotRest, err := parseFactoryFlag(c.args)
			if c.wantErr {
				if err == nil || !strings.Contains(err.Error(), c.errMarker) {
					t.Fatalf("parseFactoryFlag(%v) error = %v, want containing %q", c.args, err, c.errMarker)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFactoryFlag(%v) unexpected error: %v", c.args, err)
			}
			if gotN != c.wantN || gotOn != c.wantOn {
				t.Errorf("parseFactoryFlag(%v) = (%d, %v), want (%d, %v)", c.args, gotN, gotOn, c.wantN, c.wantOn)
			}
			wantRest := c.wantRest
			if wantRest == nil {
				wantRest = []string{}
			}
			if !slices.Equal(gotRest, wantRest) {
				t.Errorf("parseFactoryFlag(%v) rest = %v, want %v", c.args, gotRest, wantRest)
			}
		})
	}
}

// TestParseFactoryFlagStopsAtPassThroughMarker asserts the shared `--`
// discipline: nothing past the marker is read, and the marker plus everything
// after it is forwarded verbatim.
func TestParseFactoryFlagStopsAtPassThroughMarker(t *testing.T) {
	t.Parallel()

	args := []string{"--", "-f", "4"}
	n, on, rest, err := parseFactoryFlag(args)
	if err != nil || on || n != 0 {
		t.Fatalf("read past the pass-through marker: (%d, %v, %v)", n, on, err)
	}
	if !slices.Equal(rest, args) {
		t.Errorf("rest = %v, want %v verbatim", rest, args)
	}
}

// TestResolveFactoryBranch is the factory counterpart of the kanban §A.2
// truth table: -f N plus a worker-shape name selects the worker branch, -f N
// alone (or with a non-worker name) the lead branch, and no -f is a no-op
// regardless of the name.
func TestResolveFactoryBranch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		enabled  bool
		isWorker bool
		want     factoryBranch
	}{
		{true, false, factoryBranchLead},
		{true, true, factoryBranchWorker},
		{false, false, factoryBranchNone},
		{false, true, factoryBranchNone},
	}
	for _, c := range cases {
		if got := resolveFactoryBranch(c.enabled, c.isWorker); got != c.want {
			t.Errorf("resolveFactoryBranch(%v, %v) = %v, want %v", c.enabled, c.isWorker, got, c.want)
		}
	}
}

// TestParseFactoryWorkerLabelRecognizesWithoutConsuming is the load-bearing
// property shared with parseCompanionLabel: moai learns the label, and
// claude still receives the flag.
func TestParseFactoryWorkerLabelRecognizesWithoutConsuming(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"long form", []string{"--name", "worker-1"}, "worker-1"},
		{"short form", []string{"-n", "worker-12"}, "worker-12"},
		{"long equals", []string{"--name=worker-3"}, "worker-3"},
		{"short equals", []string{"-n=worker-2"}, "worker-2"},

		{"absent", []string{"-f", "4"}, ""},
		{"non-worker name", []string{"--name", "run-tjlgt1"}, ""},
		{"lead shape", []string{"--name", "lead-abc123"}, ""},
		{"unnumbered", []string{"--name", "worker"}, ""},
		{"suffix not a number", []string{"--name", "worker-a"}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			before := slices.Clone(c.args)

			got, ok := parseFactoryWorkerLabel(c.args)
			if got != c.want || ok != (c.want != "") {
				t.Errorf("parseFactoryWorkerLabel(%v) = (%q, %v), want (%q, %v)",
					before, got, ok, c.want, c.want != "")
			}
			if !slices.Equal(c.args, before) {
				t.Errorf("args mutated: %v -> %v", before, c.args)
			}
		})
	}
}

// TestEnterFactoryLeadModeEnv asserts the lead branch publishes the factory
// signal and reuses the kanban run-id / socket surfaces without seeding the
// kanban chain, and that the restore returns the prior absence.
func TestEnterFactoryLeadModeEnv(t *testing.T) {
	clearFactoryTestEnv(t)

	restore := enterFactoryLeadMode(4, "lead-abc123")
	defer restore()

	if got := os.Getenv(config.EnvMoaiFactoryWorkers); got != "4" {
		t.Errorf("MOAI_FACTORY_WORKERS = %q, want 4", got)
	}
	if got := os.Getenv(config.EnvMoaiKanbanID); got != "abc123" {
		t.Errorf("MOAI_KANBAN_ID = %q, want the adopted run id abc123", got)
	}
	if got := os.Getenv(config.EnvMoaiKanbanLeadAddr); got != "/tmp/moai-kanban-abc123" {
		t.Errorf("MOAI_KANBAN_LEAD_ADDR = %q, want /tmp/moai-kanban-abc123", got)
	}
	for _, key := range []string{config.EnvMoaiKanban, config.EnvMoaiKanbanLabel, config.EnvMoaiFactoryWorker} {
		if _, present := os.LookupEnv(key); present {
			t.Errorf("%s must stay unset on a factory lead (no kanban chain is seeded), got a value", key)
		}
	}

	restore()
	if _, present := os.LookupEnv(config.EnvMoaiFactoryWorkers); present {
		t.Error("restore left MOAI_FACTORY_WORKERS set; prior absence must be restored")
	}
}

// TestEnterFactoryLeadModeMintsRunID asserts a bare lead (no operator-supplied
// name) still publishes a well-shaped run id for the notice and the socket.
func TestEnterFactoryLeadModeMintsRunID(t *testing.T) {
	clearFactoryTestEnv(t)

	restore := enterFactoryLeadMode(2, "")
	defer restore()

	runID := os.Getenv(config.EnvMoaiKanbanID)
	if runID == "" {
		t.Fatal("MOAI_KANBAN_ID empty for a bare factory lead; expected a minted run id")
	}
	if _, ok := kanban.SplitLeadLabel(kanban.LeadLabel(runID)); !ok {
		t.Errorf("minted run id %q does not round-trip through the lead label shape", runID)
	}
}

// TestEnterFactoryWorkerModeEnv asserts the worker branch publishes its label
// and the run's count, and nothing that seeds a chain.
func TestEnterFactoryWorkerModeEnv(t *testing.T) {
	clearFactoryTestEnv(t)

	restore := enterFactoryWorkerMode("worker-3", 5)
	defer restore()

	if got := os.Getenv(config.EnvMoaiFactoryWorker); got != "worker-3" {
		t.Errorf("MOAI_FACTORY_WORKER = %q, want worker-3", got)
	}
	if got := os.Getenv(config.EnvMoaiFactoryWorkers); got != "5" {
		t.Errorf("MOAI_FACTORY_WORKERS = %q, want 5", got)
	}
	for _, key := range []string{config.EnvMoaiKanban, config.EnvMoaiKanbanLabel, config.EnvMoaiKanbanID} {
		if _, present := os.LookupEnv(key); present {
			t.Errorf("%s must stay unset on a factory worker, got a value", key)
		}
	}

	restore()
	if _, present := os.LookupEnv(config.EnvMoaiFactoryWorker); present {
		t.Error("restore left MOAI_FACTORY_WORKER set; prior absence must be restored")
	}
}

// TestResolveFactoryWorkerName covers the bump rule end to end through the
// registry: free names pass through, live claims bump to the next free
// number, dead claims are reused and pruned, and an unwritable registry
// degrades to the label as supplied.
func TestResolveFactoryWorkerName(t *testing.T) {
	t.Run("free name is kept and registered", func(t *testing.T) {
		root := t.TempDir()
		if got := resolveFactoryWorkerName(root, "worker-1", nil); got != "worker-1" {
			t.Fatalf("free name = %q, want worker-1", got)
		}
		reg := loadFactoryRegistry(factoryRegistryPath(root))
		if e, ok := reg["worker-1"]; !ok || e.PID != os.Getpid() {
			t.Errorf("worker-1 not registered to this pid: %+v", reg)
		}
	})

	t.Run("live claim bumps to the next free number", func(t *testing.T) {
		root := t.TempDir()
		// Simulate two live holders: worker-2 and worker-3.
		reg := map[string]factoryWorkerEntry{
			"worker-2": {PID: 11100},
			"worker-3": {PID: 11101},
		}
		if err := saveFactoryRegistry(factoryRegistryPath(root), reg); err != nil {
			t.Fatalf("seed registry: %v", err)
		}
		probe := factoryProcessAlive
		factoryProcessAlive = func(pid int) bool { return pid == 11100 || pid == 11101 }
		defer func() { factoryProcessAlive = probe }()

		var notes bytes.Buffer
		got := resolveFactoryWorkerName(root, "worker-2", &notes)
		if got != "worker-4" {
			t.Fatalf("bumped name = %q, want worker-4 (2 and 3 are live)", got)
		}
		if !strings.Contains(notes.String(), "worker-4") {
			t.Errorf("operator note missing the final name: %q", notes.String())
		}
	})

	t.Run("dead claim frees the name and is pruned", func(t *testing.T) {
		root := t.TempDir()
		if err := saveFactoryRegistry(factoryRegistryPath(root), map[string]factoryWorkerEntry{
			"worker-2": {PID: 11100},
		}); err != nil {
			t.Fatalf("seed registry: %v", err)
		}
		probe := factoryProcessAlive
		factoryProcessAlive = func(int) bool { return false }
		defer func() { factoryProcessAlive = probe }()

		got := resolveFactoryWorkerName(root, "worker-2", nil)
		if got != "worker-2" {
			t.Fatalf("dead claim should free the name, got %q", got)
		}
		reg := loadFactoryRegistry(factoryRegistryPath(root))
		if e, ok := reg["worker-2"]; !ok || e.PID != os.Getpid() {
			t.Errorf("worker-2 should be re-registered to this pid after pruning: %+v", reg)
		}
	})

	t.Run("unwritable registry degrades to the supplied label", func(t *testing.T) {
		// A FILE at the registry's directory path makes both read and write fail.
		blocker := filepath.Join(t.TempDir(), "blocker")
		if err := os.WriteFile(blocker, []byte("{}"), 0o600); err != nil {
			t.Fatalf("plant blocker: %v", err)
		}
		root := blocker // .moai/state/factory/ resolves under a file → fails

		if got := resolveFactoryWorkerName(root, "worker-7", nil); got != "worker-7" {
			t.Fatalf("fail-open name = %q, want worker-7 as supplied", got)
		}
	})
}

// TestReplaceNamedLabel covers the four claude-accepted name forms and the
// no-op paths. The bumped number must reach the backend argv.
func TestReplaceNamedLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		want []string
	}{
		{"long form", []string{"--name", "worker-2", "-b"}, []string{"--name", "worker-4", "-b"}},
		{"short form", []string{"-n", "worker-2"}, []string{"-n", "worker-4"}},
		{"long equals", []string{"--name=worker-2"}, []string{"--name=worker-4"}},
		{"short equals", []string{"-n=worker-2"}, []string{"-n=worker-4"}},
		{"different label untouched", []string{"--name", "other", "--name", "worker-2"}, []string{"--name", "other", "--name", "worker-4"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			before := slices.Clone(c.args)
			got := replaceNamedLabel(c.args, "worker-2", "worker-4")
			if !slices.Equal(got, c.want) {
				t.Errorf("replaceNamedLabel(%v) = %v, want %v", before, got, c.want)
			}
		})
	}

	t.Run("past the marker is not rewritten", func(t *testing.T) {
		t.Parallel()
		args := []string{"--", "--name", "worker-2"}
		if got := replaceNamedLabel(args, "worker-2", "worker-4"); !slices.Equal(got, args) {
			t.Errorf("rewrote beyond the pass-through marker: %v", got)
		}
	})

	t.Run("identical labels return the same slice", func(t *testing.T) {
		t.Parallel()
		args := []string{"--name", "worker-2"}
		if got := replaceNamedLabel(args, "worker-2", "worker-2"); !slices.Equal(got, args) {
			t.Errorf("no-op rewrite changed args: %v", got)
		}
	})
}

// TestRejectConflictingModes asserts -k and -f cannot both seed a session.
func TestRejectConflictingModes(t *testing.T) {
	t.Parallel()

	if err := rejectConflictingModes(true, true); err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("both tokens must be rejected as mutually exclusive, got %v", err)
	}
	if err := rejectConflictingModes(true, false); err != nil {
		t.Errorf("kanban alone must pass, got %v", err)
	}
	if err := rejectConflictingModes(false, true); err != nil {
		t.Errorf("factory alone must pass, got %v", err)
	}
	if err := rejectConflictingModes(false, false); err != nil {
		t.Errorf("neither token must pass, got %v", err)
	}
}

// TestRejectFactoryOnCG mirrors the kanban cg rejection: the sentinel error
// on a factory token, nil without one, and the parse error for a malformed
// count.
func TestRejectFactoryOnCG(t *testing.T) {
	t.Parallel()

	err := rejectFactoryOnCG([]string{"-f", "4"})
	if err == nil || !strings.Contains(err.Error(), factoryUnsupportedBackendSentinel) {
		t.Errorf("factory token on cg must carry the sentinel, got %v", err)
	}
	if err := rejectFactoryOnCG([]string{"-p", "work"}); err != nil {
		t.Errorf("no factory token must pass, got %v", err)
	}
	if err := rejectFactoryOnCG([]string{"-f"}); err == nil || !strings.Contains(err.Error(), "requires a worker count") {
		t.Errorf("malformed factory token must surface the parse error, got %v", err)
	}
}

// TestFactoryGenealogyInHelp is the binding genealogy AC: both launchers'
// help must state that the pre-3.1 factory flag was renamed to -k in #1513
// (7f61332ef) and that today's -f is a different feature. A user hunting
// "what happened to -f" reads this text first.
func TestFactoryGenealogyInHelp(t *testing.T) {
	t.Parallel()

	for _, cmd := range []string{ccCmd.Long, glmCmd.Long} {
		for _, marker := range []string{"--factory", "#1513", "7f61332ef", "RENAMED"} {
			if !strings.Contains(cmd, marker) {
				t.Errorf("help text missing genealogy marker %q", marker)
			}
		}
	}
}
