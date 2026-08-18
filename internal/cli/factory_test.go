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
// unconditional on them). The per-lane agent cap joins the list for the same
// reason — its seed is fill-if-absent, so an ambient value would mask it.
func clearFactoryTestEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		config.EnvMoaiFactoryWorkers,
		config.EnvMoaiFactoryWorker,
		config.EnvMoaiKanban,
		config.EnvMoaiKanbanID,
		config.EnvMoaiKanbanLabel,
		config.EnvMoaiKanbanLeadAddr,
		config.EnvClaudeCodeMaxConcurrentSubagents,
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
		{name: "companion name stays kanban", args: []string{"-k", "--name", "planner"}, wantKanban: true, wantRest: []string{"--name", "planner"}},
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

// TestParseFactoryFlag is the t118 -f truth table: bare -f is enabled with
// no count, -f N carries the count, -f worker-<n> carries the worker number,
// the `=`-joined forms work, and a SUPPLIED value that is neither a positive
// integer nor a worker label errors (the factory has no SPEC shape to fall
// into). Lookalike tokens are not stolen.
func TestParseFactoryFlag(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		args          []string
		wantEnabled   bool
		wantWorkers   int
		wantWorkerNum int
		wantRest      []string
		wantErr       bool
		errMarker     string
	}{
		{name: "absent", args: []string{"-p", "work"}, wantRest: []string{"-p", "work"}},
		{name: "bare -f", args: []string{"-f"}, wantEnabled: true},
		{name: "long form bare", args: []string{"--factory", "-b"}, wantEnabled: true, wantRest: []string{"-b"}},
		{name: "-f N", args: []string{"-f", "4"}, wantEnabled: true, wantWorkers: 4},
		{name: "-f=N", args: []string{"-f=3"}, wantEnabled: true, wantWorkers: 3},
		{name: "--factory N", args: []string{"--factory", "12"}, wantEnabled: true, wantWorkers: 12},
		{name: "--factory=N", args: []string{"--factory=1"}, wantEnabled: true, wantWorkers: 1},
		{name: "-f worker-2", args: []string{"-f", "worker-2"}, wantEnabled: true, wantWorkerNum: 2},
		{name: "-f=worker-3", args: []string{"-f=worker-3"}, wantEnabled: true, wantWorkerNum: 3},
		{name: "--factory=worker-7", args: []string{"--factory=worker-7"}, wantEnabled: true, wantWorkerNum: 7},
		{name: "positional flag is not a value", args: []string{"-f", "-b"}, wantEnabled: true, wantRest: []string{"-b"}},
		{name: "zero count errors", args: []string{"-f", "0"}, wantErr: true, errMarker: "worker count of 1 or more"},
		{name: "negative joined count errors", args: []string{"-f=-2"}, wantErr: true, errMarker: "worker count of 1 or more"},
		{name: "non-numeric non-worker errors", args: []string{"-f", "SPEC-X-001"}, wantErr: true, errMarker: "worker label"},
		{name: "unnumbered worker errors", args: []string{"-f", "worker"}, wantErr: true, errMarker: "worker label"},
		{name: "worker zero errors", args: []string{"-f", "worker-0"}, wantErr: true, errMarker: "worker label"},
		{name: "lookalike long flag not stolen", args: []string{"--factory-reset"}, wantRest: []string{"--factory-reset"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			p, err := parseFactoryFlag(c.args)
			if c.wantErr {
				if err == nil || !strings.Contains(err.Error(), c.errMarker) {
					t.Fatalf("parseFactoryFlag(%v) error = %v, want containing %q", c.args, err, c.errMarker)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFactoryFlag(%v) unexpected error: %v", c.args, err)
			}
			if p.Enabled != c.wantEnabled || p.Workers != c.wantWorkers || p.WorkerNumber != c.wantWorkerNum {
				t.Errorf("parseFactoryFlag(%v) = (enabled %v, workers %d, workerNum %d), want (%v, %d, %d)",
					c.args, p.Enabled, p.Workers, p.WorkerNumber, c.wantEnabled, c.wantWorkers, c.wantWorkerNum)
			}
			wantRest := c.wantRest
			if wantRest == nil {
				wantRest = []string{}
			}
			if !slices.Equal(p.Rest, wantRest) {
				t.Errorf("parseFactoryFlag(%v) rest = %v, want %v", c.args, p.Rest, wantRest)
			}
		})
	}
}

// TestParseFactoryFlagPassThroughBoundary asserts the shared `--` discipline
// on the -f parse: nothing past the marker is read, and the marker plus
// everything after it is forwarded verbatim.
func TestParseFactoryFlagPassThroughBoundary(t *testing.T) {
	t.Parallel()

	args := []string{"--", "-f", "worker-1"}
	p, err := parseFactoryFlag(args)
	if err != nil || p.Enabled {
		t.Fatalf("read past the pass-through marker: (%v, %v)", err, p.Enabled)
	}
	if !slices.Equal(p.Rest, args) {
		t.Errorf("rest = %v, want %v verbatim", p.Rest, args)
	}
}

// TestParseLauncherEntryMerge is the t118 merge truth table: -f alone
// resolves the lead default, -f N carries N, -f worker-<n> desugars into the
// --name worker form with an unknown (0) count, the -k shapes are untouched
// when -f is absent, and -f plus -k (or plus an operator --name on the
// worker form) is a conflict error.
func TestParseLauncherEntryMerge(t *testing.T) {
	t.Parallel()

	t.Run("bare -f resolves the one-worker lead default", func(t *testing.T) {
		t.Parallel()
		p, err := parseLauncherEntry([]string{"-f"})
		if err != nil {
			t.Fatalf("parseLauncherEntry(-f): %v", err)
		}
		if !p.FactoryEnabled || p.FactoryWorkers != config.DefaultFactoryLeadWorkers {
			t.Errorf("bare -f = (factory %v, workers %d), want (true, %d)", p.FactoryEnabled, p.FactoryWorkers, config.DefaultFactoryLeadWorkers)
		}
		if len(p.Rest) != 0 {
			t.Errorf("bare -f rest = %v, want empty (the token must not reach the launcher)", p.Rest)
		}
	})

	t.Run("-f N carries the count", func(t *testing.T) {
		t.Parallel()
		p, err := parseLauncherEntry([]string{"-f", "4"})
		if err != nil {
			t.Fatalf("parseLauncherEntry(-f 4): %v", err)
		}
		if !p.FactoryEnabled || p.FactoryWorkers != 4 {
			t.Errorf("-f 4 = (factory %v, workers %d), want (true, 4)", p.FactoryEnabled, p.FactoryWorkers)
		}
		if len(p.Rest) != 0 {
			t.Errorf("-f 4 rest = %v, want empty (the token must not reach the launcher)", p.Rest)
		}
	})

	t.Run("-f worker-n desugars to the worker form with unknown count", func(t *testing.T) {
		t.Parallel()
		p, err := parseLauncherEntry([]string{"-f", "worker-2", "-b"})
		if err != nil {
			t.Fatalf("parseLauncherEntry(-f worker-2): %v", err)
		}
		if !p.FactoryEnabled || p.FactoryWorkers != 0 {
			t.Errorf("-f worker-2 = (factory %v, workers %d), want (true, 0 unknown)", p.FactoryEnabled, p.FactoryWorkers)
		}
		label, ok := parseFactoryWorkerLabel(p.Rest)
		if !ok || label != "worker-2" {
			t.Errorf("desugared rest %v carries label (%q, %v), want worker-2", p.Rest, label, ok)
		}
		if !slices.Equal(p.Rest, []string{"-b", "--name", "worker-2"}) {
			t.Errorf("desugared rest = %v, want [-b --name worker-2]", p.Rest)
		}
	})

	t.Run("-f N --name worker-i keeps N", func(t *testing.T) {
		t.Parallel()
		p, err := parseLauncherEntry([]string{"-f", "5", "--name", "worker-2"})
		if err != nil {
			t.Fatalf("parseLauncherEntry(-f 5 --name worker-2): %v", err)
		}
		if !p.FactoryEnabled || p.FactoryWorkers != 5 {
			t.Errorf("-f 5 --name worker-2 = (factory %v, workers %d), want (true, 5)", p.FactoryEnabled, p.FactoryWorkers)
		}
	})

	t.Run("-k shapes untouched without -f", func(t *testing.T) {
		t.Parallel()
		p, err := parseLauncherEntry([]string{"-k", "4", "--name", "worker-2"})
		if err != nil {
			t.Fatalf("parseLauncherEntry(-k 4 --name worker-2): %v", err)
		}
		if !p.FactoryEnabled || p.FactoryWorkers != 4 || !p.KanbanEnabled {
			t.Errorf("-k shape altered by the merge: %+v", p)
		}
	})

	t.Run("-f with -k is a conflict", func(t *testing.T) {
		t.Parallel()
		for _, args := range [][]string{
			{"-f", "4", "-k"},
			{"-f", "-k", "SPEC-X-001"},
			{"--factory=worker-2", "-k", "3"},
		} {
			if _, err := parseLauncherEntry(args); err == nil || !strings.Contains(err.Error(), "at most one") {
				t.Errorf("parseLauncherEntry(%v) = %v, want the one-entry-token conflict", args, err)
			}
		}
	})

	t.Run("-f worker-n plus an operator --name is a conflict", func(t *testing.T) {
		t.Parallel()
		if _, err := parseLauncherEntry([]string{"-f", "worker-2", "--name", "worker-3"}); err == nil ||
			!strings.Contains(err.Error(), "already names the worker") {
			t.Errorf("parseLauncherEntry(-f worker-2 --name worker-3) = %v, want the naming conflict", err)
		}
	})

	t.Run("pass-through marker protects a backend -f", func(t *testing.T) {
		t.Parallel()
		args := []string{"--", "-f"}
		p, err := parseLauncherEntry(args)
		if err != nil || p.FactoryEnabled || p.KanbanEnabled {
			t.Fatalf("read past the pass-through marker: (%v, %+v)", err, p)
		}
		if !slices.Equal(p.Rest, args) {
			t.Errorf("rest = %v, want %v verbatim", p.Rest, args)
		}
	})
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
		{"non-worker name", []string{"--name", "runner-tjlgt1"}, ""},
		{"lead shape", []string{"--name", "leader-abc123"}, ""},
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

	restore := enterFactoryLeadMode(4, "leader-abc123")
	defer restore()

	if got := os.Getenv(config.EnvMoaiFactoryWorkers); got != "4" {
		t.Errorf("MOAI_FACTORY_WORKERS = %q, want 4", got)
	}
	if got := os.Getenv(config.EnvMoaiKanbanID); got != "abc123" {
		t.Errorf("MOAI_KANBAN_ID = %q, want the adopted run id abc123", got)
	}
	if got := os.Getenv(config.EnvMoaiKanbanLeadAddr); got != "/tmp/moai-socket-factory/abc123" {
		t.Errorf("MOAI_KANBAN_LEAD_ADDR = %q, want /tmp/moai-socket-factory/abc123 (the factory socket directory)", got)
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

// TestEnterFactoryWorkerModeEnv asserts the worker branch publishes its label,
// the run's count, the per-lane agent cap, and nothing that seeds a chain.
// The count 0 (the incremental `-f worker-<n>` form) must publish as "0" —
// present for the presence-based readers (block-cap inject, model-override
// guard), value 0 for the notice's count-less degradation.
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
	// t118: the lane cap is seeded on the worker — it is where dispatched
	// cards are implemented and fanned out to subagents.
	if got := os.Getenv(config.EnvClaudeCodeMaxConcurrentSubagents); got != "10" {
		t.Errorf("%s = %q, want 10 (the per-lane cap)", config.EnvClaudeCodeMaxConcurrentSubagents, got)
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
	if _, present := os.LookupEnv(config.EnvClaudeCodeMaxConcurrentSubagents); present {
		t.Error("restore left the lane cap set; prior absence must be restored")
	}
}

// TestEnterFactoryWorkerModeUnknownCount pins the incremental form's count
// contract: workers=0 publishes "0" (present, value 0) rather than a
// fabricated fan-out size, and an operator-set cap survives the seed.
func TestEnterFactoryWorkerModeUnknownCount(t *testing.T) {
	clearFactoryTestEnv(t)
	t.Setenv(config.EnvClaudeCodeMaxConcurrentSubagents, "3")

	restore := enterFactoryWorkerMode("worker-5", 0)
	defer restore()

	if got := os.Getenv(config.EnvMoaiFactoryWorkers); got != "0" {
		t.Errorf("MOAI_FACTORY_WORKERS = %q, want 0 (count unknown on the incremental form)", got)
	}
	if _, present := os.LookupEnv(config.EnvMoaiFactoryWorkers); !present {
		t.Error("MOAI_FACTORY_WORKERS must be PRESENT (presence is the block-cap / override-guard signal)")
	}
	if got := os.Getenv(config.EnvClaudeCodeMaxConcurrentSubagents); got != "3" {
		t.Errorf("%s = %q, want the operator's 3 (an operator-set cap wins over the seed)", config.EnvClaudeCodeMaxConcurrentSubagents, got)
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

// TestRejectFactoryOnCG asserts the FACTORY forms of BOTH entry tokens — the
// v1.2.0 -k shapes and the t118 -f shapes — are rejected on cg with the
// factory sentinel, while the plain kanban forms fall through to the kanban
// rejection (rejectKanbanOnCG), and an invalid value or the -f+-k conflict
// surfaces the parse error.
func TestRejectFactoryOnCG(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		{"-k", "4"},
		{"-k", "--name", "worker-1"},
		{"-f"},
		{"-f", "4"},
		{"-f", "worker-2"},
		{"--factory=3"},
	} {
		if err := rejectFactoryOnCG(args); err == nil || !strings.Contains(err.Error(), factoryUnsupportedBackendSentinel) {
			t.Errorf("factory form %v on cg must carry the sentinel, got %v", args, err)
		}
	}
	if err := rejectFactoryOnCG([]string{"-p", "work"}); err != nil {
		t.Errorf("no entry token must pass, got %v", err)
	}
	if err := rejectFactoryOnCG([]string{"-k", "0"}); err == nil || !strings.Contains(err.Error(), "worker count of 1 or more") {
		t.Errorf("invalid factory count must surface the parse error, got %v", err)
	}
	if err := rejectFactoryOnCG([]string{"-f", "SPEC-X-001"}); err == nil || !strings.Contains(err.Error(), "worker label") {
		t.Errorf("invalid -f value must surface the parse error, got %v", err)
	}
	if err := rejectFactoryOnCG([]string{"-f", "4", "-k"}); err == nil || !strings.Contains(err.Error(), "at most one") {
		t.Errorf("-f plus -k on cg must surface the conflict, got %v", err)
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

// TestFactoryGenealogyInHelp is the binding genealogy AC (t118): both
// launchers' help must state the full flag history — renamed to -k in #1513
// (7f61332ef), retired v1.2.0, revived t118 — and must document the revived
// -f entry forms (count form and incremental worker form). A user hunting
// "what happened to -f" reads this text first.
func TestFactoryGenealogyInHelp(t *testing.T) {
	t.Parallel()

	for _, cmd := range []string{ccCmd.Long, glmCmd.Long} {
		for _, marker := range []string{
			"--factory", "#1513", "7f61332ef", "RENAMED", "RETIRED",
			"-f, --factory [N]", // the revived entry form is documented again
			"-f worker-<n>",     // the incremental single-worker form
			"-k <N>",            // the v1.2.0 unified shapes remain documented
			"t118",              // the revival names its own card
		} {
			if !strings.Contains(cmd, marker) {
				t.Errorf("help text missing genealogy/entry marker %q", marker)
			}
		}
	}
}

// TestFactoryDefaultWorkersConstant pins the operator-decided defaults: the
// legacy count-less -k worker-name form means 8 (t85), and the t118 bare -f
// means 1 (one worker, grown incrementally). The numbers are asserted where
// they live rather than re-derived at each call site.
func TestFactoryDefaultWorkersConstant(t *testing.T) {
	t.Parallel()

	if config.DefaultFactoryWorkers != 8 {
		t.Errorf("DefaultFactoryWorkers = %d, want the operator-decided 8 (legacy -k form)", config.DefaultFactoryWorkers)
	}
	if config.DefaultFactoryLeadWorkers != 1 {
		t.Errorf("DefaultFactoryLeadWorkers = %d, want 1 (t118 bare -f: one worker, grown incrementally)", config.DefaultFactoryLeadWorkers)
	}
}

// captureFactoryLaunch records the backend argv and the factory environment
// at the moment the launch happens (the deferred restores are still live
// there, which is the point — the signal REQ-FM-023 transports).
type factoryLaunchCapture struct {
	args    []string
	workers string
	worker  string
	addr    string
	cap     string
}

// installFactoryLaunchSeam swaps unifiedLaunchFunc, findProjectRootFn, and
// deps for a capture, and restores all three on test cleanup.
func installFactoryLaunchSeam(t *testing.T) *factoryLaunchCapture {
	t.Helper()
	c := &factoryLaunchCapture{}
	origLaunch := unifiedLaunchFunc
	unifiedLaunchFunc = func(_ string, _ string, args []string) error {
		c.args = args
		c.workers = os.Getenv(config.EnvMoaiFactoryWorkers)
		c.worker = os.Getenv(config.EnvMoaiFactoryWorker)
		c.addr = os.Getenv(config.EnvMoaiKanbanLeadAddr)
		c.cap = os.Getenv(config.EnvClaudeCodeMaxConcurrentSubagents)
		return nil
	}
	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	origDeps := deps
	deps = nil
	t.Cleanup(func() {
		unifiedLaunchFunc = origLaunch
		findProjectRootFn = origFn
		deps = origDeps
	})
	return c
}

// TestCC_FactoryEntryThroughRunCC drives the t118 -f surface through the real
// cc command: the token never reaches the launcher, the lead shapes publish
// the factory signal with the t118 socket scheme, and the incremental worker
// shape desugars into the worker branch with the per-lane cap live at launch.
func TestCC_FactoryEntryThroughRunCC(t *testing.T) {
	t.Run("bare -f is the one-worker factory lead", func(t *testing.T) {
		clearFactoryTestEnv(t)
		c := installFactoryLaunchSeam(t)

		buf := new(bytes.Buffer)
		ccCmd.SetOut(buf)
		ccCmd.SetErr(buf)
		if err := runCC(ccCmd, []string{"-f"}); err != nil {
			t.Fatalf("runCC(-f): %v", err)
		}
		for _, a := range c.args {
			if a == "-f" || a == "--factory" {
				t.Errorf("the -f token must not reach the launcher, got %v", c.args)
			}
		}
		if c.workers != "1" {
			t.Errorf("MOAI_FACTORY_WORKERS at launch = %q, want 1 (DefaultFactoryLeadWorkers)", c.workers)
		}
		if !strings.HasPrefix(c.addr, "/tmp/moai-socket-factory/") {
			t.Errorf("leader socket at launch = %q, want the /tmp/moai-socket-factory/ scheme", c.addr)
		}
	})

	t.Run("-f 4 carries the count to the lead", func(t *testing.T) {
		clearFactoryTestEnv(t)
		c := installFactoryLaunchSeam(t)

		buf := new(bytes.Buffer)
		ccCmd.SetOut(buf)
		ccCmd.SetErr(buf)
		if err := runCC(ccCmd, []string{"-f", "4"}); err != nil {
			t.Fatalf("runCC(-f 4): %v", err)
		}
		if c.workers != "4" {
			t.Errorf("MOAI_FACTORY_WORKERS at launch = %q, want 4", c.workers)
		}
		if c.worker != "" {
			t.Errorf("MOAI_FACTORY_WORKER at launch = %q, want unset on a lead", c.worker)
		}
	})

	t.Run("-f worker-2 desugars into the worker branch", func(t *testing.T) {
		clearFactoryTestEnv(t)
		c := installFactoryLaunchSeam(t)

		buf := new(bytes.Buffer)
		ccCmd.SetOut(buf)
		ccCmd.SetErr(buf)
		if err := runCC(ccCmd, []string{"-f", "worker-2"}); err != nil {
			t.Fatalf("runCC(-f worker-2): %v", err)
		}
		if c.worker != "worker-2" {
			t.Errorf("MOAI_FACTORY_WORKER at launch = %q, want worker-2", c.worker)
		}
		if c.workers != "0" {
			t.Errorf("MOAI_FACTORY_WORKERS at launch = %q, want 0 (count unknown on the incremental form)", c.workers)
		}
		if !slices.Contains(c.args, "worker-2") {
			t.Errorf("the desugared --name must reach the launcher, got %v", c.args)
		}
		if c.cap != "10" {
			t.Errorf("%s at launch = %q, want 10 (the per-lane cap)", config.EnvClaudeCodeMaxConcurrentSubagents, c.cap)
		}
	})

	t.Run("-f with -k errors before the launch", func(t *testing.T) {
		clearFactoryTestEnv(t)
		installFactoryLaunchSeam(t)

		buf := new(bytes.Buffer)
		ccCmd.SetOut(buf)
		ccCmd.SetErr(buf)
		if err := runCC(ccCmd, []string{"-f", "4", "-k"}); err == nil || !strings.Contains(err.Error(), "at most one") {
			t.Errorf("runCC(-f 4 -k) = %v, want the one-entry-token conflict", err)
		}
	})
}

// TestGLM_FactoryWorkerEntry mirrors the cc case for the glm command: the -f
// worker form selects the worker branch there too (same parse, same
// environment contract, GLM backend constant).
func TestGLM_FactoryWorkerEntry(t *testing.T) {
	clearFactoryTestEnv(t)
	c := installFactoryLaunchSeam(t)

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)
	if err := runGLM(glmCmd, []string{"-f", "worker-3"}); err != nil {
		t.Fatalf("runGLM(-f worker-3): %v", err)
	}
	if c.worker != "worker-3" {
		t.Errorf("MOAI_FACTORY_WORKER at launch = %q, want worker-3", c.worker)
	}
	if c.cap != "10" {
		t.Errorf("%s at launch = %q, want 10 (the per-lane cap)", config.EnvClaudeCodeMaxConcurrentSubagents, c.cap)
	}
}
