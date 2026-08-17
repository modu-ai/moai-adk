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

// TestParseKanbanFlagUnifiedEntry is the v1.2.0 truth table: ONE -k token
// selects either shape — bare/-k SPEC-ID is the kanban chain, a numeric
// positional (or a worker-shape --name with no positional) is the factory.
// The numeric discriminator is unambiguous: a SPEC identifier is never a bare
// integer, and an invalid SUPPLIED count errors rather than silently becoming
// a kanban SPEC identifier.
func TestParseKanbanFlagUnifiedEntry(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		args        []string
		wantSpec    string
		wantKanban  bool
		wantFactory bool
		wantWorkers int
		wantRest    []string
		wantErr     bool
		errMarker   string
	}{
		{name: "absent", args: []string{"-p", "work"}, wantRest: []string{"-p", "work"}},
		{name: "bare -k is kanban", args: []string{"-k"}, wantKanban: true},
		{name: "long form is kanban", args: []string{"--kanban", "-b"}, wantKanban: true, wantRest: []string{"-b"}},
		{name: "-k SPEC is kanban", args: []string{"-k", "SPEC-X-001", "--print"}, wantSpec: "SPEC-X-001", wantKanban: true, wantRest: []string{"--print"}},
		{name: "positional flag is not a value", args: []string{"-k", "-b"}, wantKanban: true, wantRest: []string{"-b"}},
		{name: "-k N is factory lead", args: []string{"-k", "4"}, wantKanban: true, wantFactory: true, wantWorkers: 4},
		{name: "-k=N", args: []string{"-k=3"}, wantKanban: true, wantFactory: true, wantWorkers: 3},
		{name: "--kanban N", args: []string{"--kanban", "12"}, wantKanban: true, wantFactory: true, wantWorkers: 12},
		{name: "--kanban=N", args: []string{"--kanban=1"}, wantKanban: true, wantFactory: true, wantWorkers: 1},
		{name: "-k N with worker name is factory worker", args: []string{"-k", "4", "--name", "worker-2"}, wantKanban: true, wantFactory: true, wantWorkers: 4, wantRest: []string{"--name", "worker-2"}},
		{
			// The count-less factory entry: the worker-shape NAME selects the
			// factory, so the operator default applies. A count-less FACTORY
			// LEAD does not exist — a bare -k is the kanban lead.
			name: "bare -k with worker name takes the default", args: []string{"-k", "--name", "worker-2"},
			wantKanban: true, wantFactory: true, wantWorkers: config.DefaultFactoryWorkers, wantRest: []string{"--name", "worker-2"},
		},
		{name: "companion name stays kanban", args: []string{"-k", "--name", "plan"}, wantKanban: true, wantRest: []string{"--name", "plan"}},
		{name: "zero count errors", args: []string{"-k", "0"}, wantErr: true, errMarker: "worker count of 1 or more"},
		{name: "negative joined count errors", args: []string{"-k=-2"}, wantErr: true, errMarker: "worker count of 1 or more"},
		{name: "joined non-numeric is not a spec form", args: []string{"-k=abc"}, wantErr: true, errMarker: "worker count of 1 or more"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			p, err := parseKanbanFlag(c.args)
			if c.wantErr {
				if err == nil || !strings.Contains(err.Error(), c.errMarker) {
					t.Fatalf("parseKanbanFlag(%v) error = %v, want containing %q", c.args, err, c.errMarker)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseKanbanFlag(%v) unexpected error: %v", c.args, err)
			}
			if p.Spec != c.wantSpec || p.KanbanEnabled != c.wantKanban ||
				p.FactoryEnabled != c.wantFactory || p.FactoryWorkers != c.wantWorkers {
				t.Errorf("parseKanbanFlag(%v) = (spec %q, kanban %v, factory %v, workers %d), want (%q, %v, %v, %d)",
					c.args, p.Spec, p.KanbanEnabled, p.FactoryEnabled, p.FactoryWorkers,
					c.wantSpec, c.wantKanban, c.wantFactory, c.wantWorkers)
			}
			wantRest := c.wantRest
			if wantRest == nil {
				wantRest = []string{}
			}
			if !slices.Equal(p.Rest, wantRest) {
				t.Errorf("parseKanbanFlag(%v) rest = %v, want %v", c.args, p.Rest, wantRest)
			}
		})
	}
}

// TestParseKanbanFlagPassThroughBoundary asserts the shared `--` discipline on
// the unified parse: nothing past the marker is read (a worker name there
// never selects the factory), and the marker plus everything after it is
// forwarded verbatim.
func TestParseKanbanFlagPassThroughBoundary(t *testing.T) {
	t.Parallel()

	args := []string{"--", "-k", "4", "--name", "worker-1"}
	p, err := parseKanbanFlag(args)
	if err != nil || p.KanbanEnabled || p.FactoryEnabled {
		t.Fatalf("read past the pass-through marker: (%v, %v, %v)", err, p.KanbanEnabled, p.FactoryEnabled)
	}
	if !slices.Equal(p.Rest, args) {
		t.Errorf("rest = %v, want %v verbatim", p.Rest, args)
	}
}

// TestRejectRetiredFactoryFlag asserts the retired entry token errors with
// the redirect to the unified -k surface — on every launcher that reads it —
// and that lookalike tokens (-foo, --factory-reset) are not stolen.
func TestRejectRetiredFactoryFlag(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		{"-f"},
		{"-f", "4"},
		{"--factory"},
		{"--factory=4"},
		{"-f=4"},
		{"cc", "-f"},
	} {
		if err := rejectRetiredFactoryFlag(args); err == nil || !strings.Contains(err.Error(), "retired") {
			t.Errorf("rejectRetiredFactoryFlag(%v) = %v, want the retirement error", args, err)
		}
	}
	for _, args := range [][]string{
		{"-p", "work"},
		{"-foo"},
		{"--factory-reset"},
		{"--", "-f"}, // past the marker belongs to the child process
	} {
		if err := rejectRetiredFactoryFlag(args); err != nil {
			t.Errorf("rejectRetiredFactoryFlag(%v) = %v, want nil (not a retired token)", args, err)
		}
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

// TestRejectFactoryOnCG asserts the FACTORY forms of the unified -k surface
// are rejected on cg with the factory sentinel, while the plain kanban forms
// fall through to the kanban rejection (rejectKanbanOnCG), and an invalid
// count surfaces the parse error.
func TestRejectFactoryOnCG(t *testing.T) {
	t.Parallel()

	err := rejectFactoryOnCG([]string{"-k", "4"})
	if err == nil || !strings.Contains(err.Error(), factoryUnsupportedBackendSentinel) {
		t.Errorf("factory count on cg must carry the sentinel, got %v", err)
	}
	if err := rejectFactoryOnCG([]string{"-k", "--name", "worker-1"}); err == nil || !strings.Contains(err.Error(), factoryUnsupportedBackendSentinel) {
		t.Errorf("worker-name factory form on cg must carry the sentinel, got %v", err)
	}
	if err := rejectFactoryOnCG([]string{"-p", "work"}); err != nil {
		t.Errorf("no -k token must pass, got %v", err)
	}
	if err := rejectFactoryOnCG([]string{"-k", "0"}); err == nil || !strings.Contains(err.Error(), "worker count of 1 or more") {
		t.Errorf("invalid factory count must surface the parse error, got %v", err)
	}
	// The plain kanban forms belong to the kanban rejection, not this one.
	if err := rejectFactoryOnCG([]string{"-k"}); err != nil {
		t.Errorf("bare -k is kanban's to reject, got %v", err)
	}
	if err := rejectFactoryOnCG([]string{"-k", "SPEC-X-001"}); err != nil {
		t.Errorf("-k SPEC is kanban's to reject, got %v", err)
	}
}

// TestRejectKanbanOnCGLeavesFactoryForms is the other half of the cg split:
// rejectKanbanOnCG fires for the kanban forms and deliberately passes the
// factory forms through to rejectFactoryOnCG.
func TestRejectKanbanOnCGLeavesFactoryForms(t *testing.T) {
	t.Parallel()

	if err := rejectKanbanOnCG([]string{"-k"}); err == nil || !strings.Contains(err.Error(), kanbanUnsupportedBackendSentinel) {
		t.Errorf("bare -k on cg must carry the kanban sentinel, got %v", err)
	}
	if err := rejectKanbanOnCG([]string{"-k", "SPEC-X-001"}); err == nil {
		t.Errorf("-k SPEC on cg must carry the kanban sentinel, got %v", err)
	}
	for _, args := range [][]string{
		{"-k", "4"},
		{"-k", "--name", "worker-2"},
	} {
		if err := rejectKanbanOnCG(args); err != nil {
			t.Errorf("factory form %v is the factory rejection's, not kanban's, got %v", args, err)
		}
	}
}

// TestFactoryGenealogyInHelp is the binding genealogy AC (v1.2.0): both
// launchers' help must state that the pre-3.1 factory flag was renamed to -k
// in #1513 (7f61332ef), that -f briefly returned and was RETIRED, and that
// the factory is entered on -k today. A user hunting "what happened to -f"
// reads this text first.
func TestFactoryGenealogyInHelp(t *testing.T) {
	t.Parallel()

	for _, cmd := range []string{ccCmd.Long, glmCmd.Long} {
		for _, marker := range []string{"--factory", "#1513", "7f61332ef", "RENAMED", "RETIRED", "-k <N>"} {
			if !strings.Contains(cmd, marker) {
				t.Errorf("help text missing genealogy marker %q", marker)
			}
		}
		// The retired flag must not survive as a documented entry form.
		if strings.Contains(cmd, "-f, --factory") || strings.Contains(cmd, "-f <N>") {
			t.Error("help text still documents the retired -f entry form")
		}
	}
}

// TestFactoryDefaultWorkersConstant pins the operator-decided default fan-out
// (t85): the count-less factory entry (a worker-shape --name with no N) means
// THIS count, so the number is asserted where it lives rather than re-derived
// at each call site.
func TestFactoryDefaultWorkersConstant(t *testing.T) {
	t.Parallel()

	if config.DefaultFactoryWorkers != 8 {
		t.Errorf("DefaultFactoryWorkers = %d, want the operator-decided 8", config.DefaultFactoryWorkers)
	}
}
