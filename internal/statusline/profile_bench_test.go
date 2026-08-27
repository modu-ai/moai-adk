package statusline

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	gitpkg "github.com/modu-ai/moai-adk/internal/core/git"
)

// profile_bench_test.go — t215 warm-path decomposition assets.
//
// BenchmarkBuilder_Build (builder_test.go) pins the SLA with MOCK providers;
// the benchmarks here pin the REAL warm path with auto-detected providers
// against this checkout, plus a per-phase breakdown that attributes the wall
// clock observed by the external timing harness (bin/timeit_harness.py) to
// individual pipeline stages. Run:
//
//	go test ./internal/statusline/ -run '^$' -bench . -benchmem
//
// Distribution evidence (median/p95, not the mean go bench reports):
//
//	MOAI_PROFILE_PHASES=1 go test ./internal/statusline/ -run TestProfilePhaseDistributions -v

// profPhaseRounds controls how many timed samples the distribution reporter
// takes per phase. Enough for a stable p95, small enough to keep the gate-run
// under a second of wall per phase even for the subprocess-heavy stages.
const profPhaseRounds = 120

// profWarmSessionID mirrors a Claude Code session id in the warm fixture. A
// syntactically ordinary UUID; no state file for it ever exists, matching a
// cold first-render per phase while staying deterministic.
const profWarmSessionID = "8f3a2c1e-4b5d-6e7f-9a0b-1c2d3e4f5a6b"

// profFindMoaiRoot walks up from the working directory until it finds an
// ancestor holding a .moai directory — the project root a deployed render
// would discover through findProjectRootFn. Skips when run outside a moai
// project so the file stays usable in stripped CI trees.
func profFindMoaiRoot(tb testing.TB) string {
	tb.Helper()
	dir, err := os.Getwd()
	if err != nil {
		tb.Fatalf("getwd: %v", err)
	}
	for {
		if info, statErr := os.Stat(filepath.Join(dir, ".moai")); statErr == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			tb.Skip("no .moai ancestor above cwd; run inside a moai project checkout")
		}
		dir = parent
	}
}

// profTempStateScaffold creates the .moai/state subtree inside a scratch root
// so phase targets exercise real (existing-directory) stat/read paths without
// touching the checkout's runtime state.
func profTempStateScaffold(tb testing.TB) string {
	tb.Helper()
	root := tb.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
		tb.Fatalf("mkdir state scaffold: %v", err)
	}
	return root
}

// profWarmPayload marshals a Claude Code v2.1.246-shaped StdinData through the
// real type stack, so field names and nesting cannot drift from types.go. The
// project paths point at scaffold so the context-usage snapshot write lands in
// scratch space rather than the checkout.
func profWarmPayload(tb testing.TB, scaffold string) []byte {
	tb.Helper()
	used, remaining := 42.5, 57.5
	input := StdinData{
		HookEventName: "Status",
		SessionID:     profWarmSessionID,
		CWD:           scaffold,
		Model: &ModelInfo{
			ID:          "claude-opus-4-6",
			DisplayName: "Opus",
			Name:        "claude-opus-4-6",
		},
		Workspace: &WorkspaceInfo{CurrentDir: scaffold, ProjectDir: scaffold},
		Cost: &CostData{
			TotalUSD:        12.3456,
			TotalCostUSD:    12.3456,
			InputTokens:     1234567,
			OutputTokens:    45678,
			TotalDurationMS: 49_230_000,
		},
		ContextWindow: &ContextWindowInfo{
			UsedPercentage:      &used,
			RemainingPercentage: &remaining,
			ContextWindowSize:   1_000_000,
			TotalInputTokens:    4_500_000,
			TotalOutputTokens:   180_000,
			Used:                425_000,
			Total:               1_000_000,
			CurrentUsage: &CurrentUsageInfo{
				InputTokens:         12_500,
				CacheCreationTokens: 38_000,
				CacheReadTokens:     360_000,
				OutputTokens:        2_100,
			},
		},
		OutputStyle: &OutputStyleInfo{Name: "MoAI"},
		RateLimits: &RateLimitInfo{
			FiveHour: &RateLimitWindow{UsedPercentage: 31.4, ResetsAt: 1_787_000_000},
			SevenDay: &RateLimitWindow{UsedPercentage: 12.8, ResetsAt: 1_787_500_000},
		},
		Effort:   &EffortInfo{Level: "high"},
		Thinking: &ThinkingInfo{Enabled: true},
		Version:  "2.1.246",
	}
	raw, err := json.Marshal(input)
	if err != nil {
		tb.Fatalf("marshal warm payload: %v", err)
	}
	return raw
}

// profWarmOptions builds the CLI-equivalent option set: auto-detected git,
// version, and usage providers; the seeded theme; every segment enabled
// (SegmentConfig nil = all-on); color on as in an interactive deployment.
func profWarmOptions(projectRoot, scaffold string, todoEnabled bool) Options {
	return Options{
		Mode:          ModeDefault,
		NoColor:       false,
		RootDir:       projectRoot,
		ThemeName:     "catppuccin-mocha",
		SegmentConfig: nil,
		TodoEnabled:   &todoEnabled,
		HomeDir:       scaffold,
	}
}

// BenchmarkStatuslineWarmPath measures one complete deployed-shape render —
// fresh Builder (process-per-render in production makes builder construction
// per-render), Build(parse → collect → snapshot write → render). This is the
// in-process counterpart of the external `moai statusline < fixture.json`
// timing; subtract the process-startup floor measured externally.
func BenchmarkStatuslineWarmPath(b *testing.B) {
	projectRoot := profFindMoaiRoot(b)
	scaffold := profTempStateScaffold(b)
	payload := profWarmPayload(b, scaffold)
	opts := profWarmOptions(projectRoot, scaffold, true)

	builder := New(opts)

	b.ResetTimer()
	for range b.N {
		if _, err := builder.Build(context.Background(), bytes.NewReader(payload)); err != nil {
			b.Fatalf("Build: %v", err)
		}
	}
}

// BenchmarkPhaseBuilderInit measures statusline.New including the two git
// rev-parse spawns of repository auto-detection.
func BenchmarkPhaseBuilderInit(b *testing.B) {
	projectRoot := profFindMoaiRoot(b)
	scaffold := profTempStateScaffold(b)
	opts := profWarmOptions(projectRoot, scaffold, true)

	b.ResetTimer()
	for range b.N {
		_ = New(opts)
	}
}

// BenchmarkPhaseParseStdin measures JSON decode of the v2.1.246 payload.
func BenchmarkPhaseParseStdin(b *testing.B) {
	scaffold := profTempStateScaffold(b)
	payload := profWarmPayload(b, scaffold)
	builder := New(profWarmOptions(".", scaffold, true)).(*defaultBuilder)

	b.ResetTimer()
	for range b.N {
		if got := builder.parseStdin(bytes.NewReader(payload)); got == nil {
			b.Fatal("stdin parse returned nil")
		}
	}
}

// BenchmarkPhaseGitCollectStatus measures CollectGitStatus alone: symbolic-ref
// plus status plus upstream rev-list — three git subprocesses — against this
// real checkout.
func BenchmarkPhaseGitCollectStatus(b *testing.B) {
	projectRoot := profFindMoaiRoot(b)
	repo, err := gitpkg.NewRepository(projectRoot)
	if err != nil {
		b.Skipf("git repository unavailable at %s: %v", projectRoot, err)
	}
	collector := NewGitCollector(repo)
	ctx := context.Background()

	// Untimed priming round so the series reflects warm OS page caches.
	if _, warmErr := collector.CollectGitStatus(ctx); warmErr != nil {
		b.Skipf("git status collection failed during warm-up: %v", warmErr)
	}

	b.ResetTimer()
	for range b.N {
		if _, err := collector.CollectGitStatus(ctx); err != nil {
			b.Fatalf("CollectGitStatus: %v", err)
		}
	}
}

// BenchmarkPhaseVersionCheckUpdate measures the template-version config read
// and binary/template comparison. A FRESH collector per iteration mirrors
// production — a deployed render is a fresh process, so VersionCollector's
// in-process cache never serves a second render.
func BenchmarkPhaseVersionCheckUpdate(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for range b.N {
		collector := NewVersionCollector(versionForProfiling())
		if _, err := collector.CheckUpdate(ctx); err != nil {
			b.Fatalf("CheckUpdate: %v", err)
		}
	}
}

// BenchmarkPhaseInstantCollectors measures the no-subprocess collector group:
// memory (+ llm.yaml override probe), metrics (+ model-cache write), task,
// backlog, GitHub cache read, goal-armed probe — everything collectAll does
// before the parallel block, targeting the scratch scaffold.
func BenchmarkPhaseInstantCollectors(b *testing.B) {
	scaffold := profTempStateScaffold(b)
	payload := profWarmPayload(b, scaffold)
	builder := New(profWarmOptions(".", scaffold, true)).(*defaultBuilder)
	input := builder.parseStdin(bytes.NewReader(payload))
	if input == nil {
		b.Fatal("nil input")
	}

	b.ResetTimer()
	for range b.N {
		_ = CollectMemory(input)
		_ = CollectMetrics(input, scaffold)
		if task := CollectTask(); task != nil {
			_ = *task
		}
		boardRoot := resolveBoardRoot(input)
		_ = resolveBacklogCounts(boardRoot)
		_ = resolveGitHubCounts(boardRoot)
		_ = resolveGoalArmed(resolveProjectDir(input), input.SessionID)
	}
}

// BenchmarkPhaseRender measures renderer.Render on a fully collected
// StatusData — the widest-string, gradient-and-ANSI part of the pipeline.
func BenchmarkPhaseRender(b *testing.B) {
	projectRoot := profFindMoaiRoot(b)
	scaffold := profTempStateScaffold(b)
	payload := profWarmPayload(b, scaffold)
	builder := New(profWarmOptions(projectRoot, scaffold, true)).(*defaultBuilder)
	input := builder.parseStdin(bytes.NewReader(payload))
	if input == nil {
		b.Fatal("nil input")
	}
	data := builder.collectAll(context.Background(), input)
	mode := ModeDefault

	b.ResetTimer()
	for range b.N {
		_ = builder.renderer.Render(data, mode)
	}
}

// BenchmarkPhaseSnapshotWrite measures the context-usage telemetry record
// write (temp file + rename) that Build performs after collectAll.
func BenchmarkPhaseSnapshotWrite(b *testing.B) {
	scaffold := profTempStateScaffold(b)
	payload := profWarmPayload(b, scaffold)
	builder := New(profWarmOptions(".", scaffold, true)).(*defaultBuilder)
	input := builder.parseStdin(bytes.NewReader(payload))
	mem := CollectMemory(input)

	b.ResetTimer()
	for i := range b.N {
		writeContextUsage(resolveProjectDir(input), profWarmSessionID, 1000+i%2, *mem, handoffGuideStage(nil), "Opus", "high")
	}
}

// versionForProfiling supplies a non-empty binary version so the version
// comparison branch stays exercised exactly as in a release binary.
func versionForProfiling() string { return "3.1.3-rc.5" }

// TestProfilePhaseDistributions prints median/p95 wall per phase over
// profPhaseRounds rounds. Env-gated because CI needs neither its output nor
// its seconds. Run:
//
//	MOAI_PROFILE_PHASES=1 go test ./internal/statusline/ -run TestProfilePhaseDistributions -v
func TestProfilePhaseDistributions(t *testing.T) {
	if os.Getenv("MOAI_PROFILE_PHASES") != "1" {
		t.Skip("set MOAI_PROFILE_PHASES=1 to emit the per-phase distribution table")
	}
	projectRoot := profFindMoaiRoot(t)
	scaffold := profTempStateScaffold(t)
	payload := profWarmPayload(t, scaffold)
	options := profWarmOptions(projectRoot, scaffold, true)
	ctx := context.Background()

	builder := New(options).(*defaultBuilder)
	parsed := builder.parseStdin(bytes.NewReader(payload))
	if parsed == nil {
		t.Fatal("nil parsed input")
	}
	collected := builder.collectAll(ctx, parsed)

	repo, err := gitpkg.NewRepository(projectRoot)
	if err != nil {
		t.Skipf("git repository unavailable at %s: %v", projectRoot, err)
	}
	gitCollector := NewGitCollector(repo)

	type phase struct {
		name string
		run  func()
	}
	phases := []phase{
		{"build_end_to_end", func() {
			if _, buildErr := builder.Build(ctx, bytes.NewReader(payload)); buildErr != nil {
				t.Errorf("Build: %v", buildErr)
			}
		}},
		{"builder_init_new", func() { _ = New(options) }},
		{"stdin_parse", func() { _ = builder.parseStdin(bytes.NewReader(payload)) }},
		{"git_collect_status", func() {
			if _, gitErr := gitCollector.CollectGitStatus(ctx); gitErr != nil {
				t.Errorf("CollectGitStatus: %v", gitErr)
			}
		}},
		{"version_check_update", func() {
			collector := NewVersionCollector(versionForProfiling())
			if _, verErr := collector.CheckUpdate(ctx); verErr != nil {
				t.Errorf("CheckUpdate: %v", verErr)
			}
		}},
		{"instant_collectors", func() {
			_ = CollectMemory(parsed)
			_ = CollectMetrics(parsed, scaffold)
			_ = CollectTask()
			boardRoot := resolveBoardRoot(parsed)
			_ = resolveBacklogCounts(boardRoot)
			_ = resolveGitHubCounts(boardRoot)
			_ = resolveGoalArmed(resolveProjectDir(parsed), parsed.SessionID)
		}},
		{"snapshot_write", func() {
			writeContextUsage(resolveProjectDir(parsed), profWarmSessionID, 2000, MemoryData{}, handoffGuideStage(nil), "Opus", "high")
		}},
		{"render", func() { _ = builder.renderer.Render(collected, ModeDefault) }},
	}

	for _, ph := range phases {
		samples := make([]float64, 0, profPhaseRounds)
		for range 5 { // untimed priming rounds
			ph.run()
		}
		for range profPhaseRounds {
			start := time.Now()
			ph.run()
			samples = append(samples, float64(time.Since(start).Microseconds())/1000.0)
		}
		sort.Float64s(samples)
		median := samples[len(samples)/2]
		p95 := samples[int(float64(len(samples))*0.95)-1]
		t.Logf("%-22s N=%d median=%.3fms p95=%.3fms max=%.3fms", ph.name, len(samples), median, p95, samples[len(samples)-1])
	}
}
