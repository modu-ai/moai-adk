package hook

// Stage decomposition of SessionStart Handle's synchronous cost (card t666).
//
// TestSessionStart_StageClock pins the measurement seam itself and always runs.
// TestSessionStart_HandleStageProbe is a measurement instrument, not a check:
// it is skipped unless MOAI_HANDLE_STAGE_PROBE_N is set, and it never fails on
// a latency value. It runs Handle in interleaved arms — observer installed and
// observer absent — so the instrumentation's own cost is measured against a
// control taken under the same machine load.

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestSessionStart_StageClock(t *testing.T) {
	var nilClock *stageClock
	nilClock.lap("ignored")
	nilClock.span("ignored")()

	orig := handleStageObserver
	t.Cleanup(func() { handleStageObserver = orig })

	handleStageObserver = nil
	if newStageClock() != nil {
		t.Fatal("newStageClock returned a clock with no observer installed")
	}

	var mu sync.Mutex
	got := map[string]time.Duration{}
	handleStageObserver = func(stage string, d time.Duration) {
		mu.Lock()
		defer mu.Unlock()
		got[stage] += d
	}
	c := newStageClock()
	time.Sleep(5 * time.Millisecond)
	c.lap("first")
	end := c.span("concurrent")
	time.Sleep(5 * time.Millisecond)
	end()
	c.lap("second")

	for _, stage := range []string{"first", "concurrent", "second"} {
		if got[stage] < 5*time.Millisecond {
			t.Errorf("stage %q reported %v, want >= 5ms", stage, got[stage])
		}
	}
}

func TestSessionStart_HandleStageProbe(t *testing.T) {
	raw := os.Getenv("MOAI_HANDLE_STAGE_PROBE_N")
	if raw == "" {
		t.Skip("measurement probe: set MOAI_HANDLE_STAGE_PROBE_N to run")
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		t.Fatalf("MOAI_HANDLE_STAGE_PROBE_N=%q: want a positive integer", raw)
	}
	// Optional: a repository root whose .moai/config is copied and whose
	// .moai/specs is linked into each project, with the real drift scan left in
	// place.
	populateFrom := os.Getenv("MOAI_HANDLE_STAGE_PROBE_POPULATE")
	// Optional: run one unrecorded session per project before the measured one.
	warm := os.Getenv("MOAI_HANDLE_STAGE_PROBE_WARM") != ""
	t.Setenv("ANTHROPIC_BASE_URL", "")

	origFn := driftCountFn
	origObs := handleStageObserver
	t.Cleanup(func() {
		driftCountFn = origFn
		handleStageObserver = origObs
	})
	if populateFrom == "" {
		driftCountFn = func(_ context.Context, _ string) (int, error) { return 0, nil }
	}

	var mu sync.Mutex
	stages := map[string][]time.Duration{}
	observer := func(stage string, d time.Duration) {
		mu.Lock()
		defer mu.Unlock()
		stages[stage] = append(stages[stage], d)
	}

	units := map[string][]time.Duration{}
	var withObs, withoutObs, lapSums []time.Duration
	for i := range n {
		for _, instrumented := range []bool{i%2 == 0, i%2 != 0} {
			completed := registerDeferredScanSeam(t)
			projectDir := t.TempDir()
			mkStateDir(t, projectDir)
			if populateFrom != "" {
				src := os.DirFS(filepath.Join(populateFrom, ".moai", "config"))
				if err := os.CopyFS(filepath.Join(projectDir, ".moai", "config"), src); err != nil {
					t.Fatalf("populate config: %v", err)
				}
				// The SPEC tree is only read by the scans; a symlink avoids
				// copying tens of megabytes per iteration.
				if err := os.Symlink(filepath.Join(populateFrom, ".moai", "specs"),
					filepath.Join(projectDir, ".moai", "specs")); err != nil {
					t.Fatalf("populate specs: %v", err)
				}
			}
			input := &HookInput{
				SessionID:     fmt.Sprintf("sess-stage-probe-%d-%t", i, instrumented),
				CWD:           projectDir,
				ProjectDir:    projectDir,
				HookEventName: "SessionStart",
			}
			if instrumented && populateFrom != "" && !warm {
				// The same four scans on a second, never-used project: the
				// first-session cost, one scan at a time.
				coldDir := t.TempDir()
				mkStateDir(t, coldDir)
				if err := os.CopyFS(filepath.Join(coldDir, ".moai", "config"),
					os.DirFS(filepath.Join(populateFrom, ".moai", "config"))); err != nil {
					t.Fatalf("populate cold config: %v", err)
				}
				if err := os.Symlink(filepath.Join(populateFrom, ".moai", "specs"),
					filepath.Join(coldDir, ".moai", "specs")); err != nil {
					t.Fatalf("populate cold specs: %v", err)
				}
				c0 := time.Now()
				_ = pruneTelemetry(coldDir)
				c1 := time.Now()
				_ = detectAndWrapStaleMemories(coldDir, c1)
				c2 := time.Now()
				_ = PresentPendingProposals(coldDir)
				c3 := time.Now()
				coldCtx, coldCancel := context.WithTimeout(context.Background(), sessionStartDriftTimeout)
				_, _ = driftCountFn(coldCtx, coldDir)
				coldCancel()
				c4 := time.Now()
				mu.Lock()
				units["cold.telemetry_prune"] = append(units["cold.telemetry_prune"], c1.Sub(c0))
				units["cold.stale_memory"] = append(units["cold.stale_memory"], c2.Sub(c1))
				units["cold.proposals"] = append(units["cold.proposals"], c3.Sub(c2))
				units["cold.drift"] = append(units["cold.drift"], c4.Sub(c3))
				mu.Unlock()
			}
			if warm {
				// One unrecorded session first, so the measured one sees the
				// state a project carries after its first session.
				handleStageObserver = nil
				warmCompleted := completed
				if _, err := NewSessionStartHandler(nil).Handle(context.Background(), input); err != nil {
					t.Fatalf("warm-up Handle: %v", err)
				}
				select {
				case <-warmCompleted:
				case <-time.After(10 * time.Second):
					t.Fatalf("iteration %d: warm-up deferred scan did not complete", i)
				}
				completed = registerDeferredScanSeam(t)
			}

			mu.Lock()
			before := len(stages["guard_liveness_advisory"])
			lapTotal := sumLaps(stages)
			mu.Unlock()

			if instrumented {
				handleStageObserver = observer
			} else {
				handleStageObserver = nil
			}
			h := NewSessionStartHandler(nil)
			start := time.Now()
			if _, err := h.Handle(context.Background(), input); err != nil {
				t.Fatalf("Handle: %v", err)
			}
			elapsed := time.Since(start)
			handleStageObserver = nil

			if instrumented {
				withObs = append(withObs, elapsed)
				mu.Lock()
				if len(stages["guard_liveness_advisory"]) != before+1 {
					t.Fatalf("iteration %d: the final lap was not reported", i)
				}
				lapSums = append(lapSums, sumLaps(stages)-lapTotal)
				mu.Unlock()
			} else {
				withoutObs = append(withoutObs, elapsed)
			}
			select {
			case <-completed:
				if instrumented {
					mu.Lock()
					units["deferred.completed_after_handle_start"] = append(
						units["deferred.completed_after_handle_start"], time.Since(start))
					mu.Unlock()
				}
			case <-time.After(10 * time.Second):
				t.Fatalf("iteration %d: deferred scan did not complete", i)
			}

			// Unit costs taken under the same load, outside Handle: one git
			// subprocess, and one project-root canonicalization.
			if instrumented {
				gitStart := time.Now()
				cmd := exec.Command("git", "rev-parse", "--git-dir")
				cmd.Dir = projectDir
				_ = cmd.Run()
				canonStart := time.Now()
				_ = homestate.CanonicalProjectRoot(projectDir)
				canonEnd := time.Now()
				// The four deferred advisory scans, run one at a time after the
				// Handle's own background run has completed.
				t0 := time.Now()
				_ = pruneTelemetry(projectDir)
				t1 := time.Now()
				_ = detectAndWrapStaleMemories(projectDir, t1)
				t2 := time.Now()
				_ = PresentPendingProposals(projectDir)
				t3 := time.Now()
				driftCtx, cancel := context.WithTimeout(context.Background(), sessionStartDriftTimeout)
				_, _ = driftCountFn(driftCtx, projectDir)
				cancel()
				t4 := time.Now()
				mu.Lock()
				units["unit.git_subprocess"] = append(units["unit.git_subprocess"], canonStart.Sub(gitStart))
				units["unit.canonical_project_root"] = append(units["unit.canonical_project_root"], canonEnd.Sub(canonStart))
				units["deferred.telemetry_prune"] = append(units["deferred.telemetry_prune"], t1.Sub(t0))
				units["deferred.stale_memory"] = append(units["deferred.stale_memory"], t2.Sub(t1))
				units["deferred.proposals"] = append(units["deferred.proposals"], t3.Sub(t2))
				units["deferred.drift"] = append(units["deferred.drift"], t4.Sub(t3))
				mu.Unlock()
			}
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "MEASUREMENT n=%d populated=%t warm=%t\n", n, populateFrom != "", warm)
	fmt.Fprintf(&b, "handle.instrumented   %s\n", dist(withObs))
	fmt.Fprintf(&b, "handle.control        %s\n", dist(withoutObs))
	fmt.Fprintf(&b, "laps.sum              %s\n", dist(lapSums))
	names := make([]string, 0, len(stages))
	for k := range stages {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		fmt.Fprintf(&b, "stage %-28s %s\n", k, dist(stages[k]))
	}
	for _, k := range []string{
		"unit.git_subprocess", "unit.canonical_project_root",
		"deferred.telemetry_prune", "deferred.stale_memory", "deferred.proposals", "deferred.drift",
		"deferred.completed_after_handle_start",
		"cold.telemetry_prune", "cold.stale_memory", "cold.proposals", "cold.drift",
	} {
		fmt.Fprintf(&b, "%-34s %s\n", k, dist(units[k]))
	}
	t.Log("\n" + b.String())
}

// sumLaps totals only sequential laps; concurrent task spans overlap the
// sync_group lap and would double-count.
func sumLaps(stages map[string][]time.Duration) time.Duration {
	var total time.Duration
	for k, ds := range stages {
		if strings.HasPrefix(k, "task.") {
			continue
		}
		for _, d := range ds {
			total += d
		}
	}
	return total
}

func dist(ds []time.Duration) string {
	if len(ds) == 0 {
		return "n=0"
	}
	s := slices.Clone(ds)
	slices.Sort(s)
	q := func(p float64) time.Duration { return s[int(p*float64(len(s)-1))] }
	return fmt.Sprintf("n=%d p50=%v p90=%v max=%v", len(s),
		q(0.5).Round(10*time.Microsecond), q(0.9).Round(10*time.Microsecond), s[len(s)-1].Round(10*time.Microsecond))
}
