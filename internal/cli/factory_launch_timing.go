package cli

// factory_launch_timing.go — the codex lane launch's pre-exec timing
// collector (SPEC-CODEX-LANE-SLOTS-001 REQ-012). The launcher records one
// step per pre-exec phase — codex pre-exec init, join gate, active-run
// resolution, lane claim, exec handoff — and, when the phase exceeds the
// operator-configurable slow-launch threshold, prints one timing line per
// step naming the step and its wall-time. The instrumentation sits on the
// shared pre-exec sequence AHEAD of the POSIX/Windows exec split: the direct
// door replaces the process (syscall.Exec on POSIX), so the report is the
// launcher's last output before the seam, never a post-exec print.

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// The pre-exec step names the report carries (REQ-012: the report names each
// step). The codex lane launch records the lane claim in addition to the
// four required steps — it is the phase the original defect lived in.
const (
	factoryStepCodexPreInit = "codex pre-exec init"
	factoryStepJoinGate     = "join gate"
	factoryStepRunResolve   = "active-run resolution"
	factoryStepLaneClaim    = "lane claim"
	factoryStepExecHandoff  = "exec handoff"
)

// factoryLaunchStep is one measured pre-exec step. detail carries the
// debug-vocabulary suffix (SPEC-CODEX-DEBUG-MODE-001 — keys-only env
// content, outcomes); the threshold report never prints it, the debug dump
// does.
type factoryLaunchStep struct {
	name    string
	elapsed time.Duration
	detail  string
}

// factoryLaunchTiming collects the pre-exec steps of one launch. A nested
// step (active-run resolution inside join gate) is reported with its own
// wall-time and overlaps its parent's; the phase total is therefore the
// wall-clock from the first begin, never the sum of the step durations.
//
// Every method is nil-safe: a nil *factoryLaunchTiming measures nothing and
// prints nothing, so the shared join gate stays inert for the cc/glm twins,
// which pass nil.
type factoryLaunchTiming struct {
	start    time.Time
	started  bool
	reported bool
	// debug marks a collector created for a debug launch
	// (SPEC-CODEX-DEBUG-MODE-001): only then does beginDebug record the
	// debug-vocabulary steps, so the debug-off threshold report keeps
	// exactly its REQ-012 step set (the t1378 freeze).
	debug bool
	steps []factoryLaunchStep
}

// begin starts a step and returns the function that closes it. The returned
// closer is safe to call more than once and safe to defer.
func (t *factoryLaunchTiming) begin(name string) func() {
	if t == nil {
		return func() {}
	}
	if !t.started {
		t.started = true
		t.start = time.Now()
	}
	begin := time.Now()
	return func() {
		t.steps = append(t.steps, factoryLaunchStep{name: name, elapsed: time.Since(begin)})
	}
}

// phaseElapsed is the wall-clock the pre-exec phase has been running — the
// value the slow-launch threshold judges.
func (t *factoryLaunchTiming) phaseElapsed() time.Duration {
	if t == nil || !t.started {
		return 0
	}
	return time.Since(t.start)
}

// beginDebug records a debug-vocabulary step (SPEC-CODEX-DEBUG-MODE-001
// REQ-005) with an optional detail suffix. It measures nothing unless the
// collector was created for a debug launch — the debug-off threshold report
// keeps exactly its REQ-012 step set — and is nil-safe like begin.
// @MX:SPEC: SPEC-CODEX-DEBUG-MODE-001
func (t *factoryLaunchTiming) beginDebug(name, detail string) func() {
	if t == nil || !t.debug {
		return func() {}
	}
	end := t.begin(name)
	return func() {
		end()
		if detail != "" && len(t.steps) > 0 {
			t.steps[len(t.steps)-1].detail = detail
		}
	}
}

// annotateDetail attaches a detail to the most recently recorded step — the
// lane claim's claimed label, the anchor lock's outcome. No-op without debug
// or on an empty record; nil-safe. The launch path is single-threaded, so
// "most recent" is unambiguous.
// @MX:SPEC: SPEC-CODEX-DEBUG-MODE-001
func (t *factoryLaunchTiming) annotateDetail(detail string) {
	if t == nil || !t.debug || len(t.steps) == 0 {
		return
	}
	t.steps[len(t.steps)-1].detail = detail
}

// debugDump prints one line per recorded pre-exec step under the launcher
// debug prefix — unconditionally: the slow-launch threshold does not gate it
// (REQ-014), and the lines are this collector's recorded steps, the single
// recording site (no second instrumentation exists). Like reportSlow it must
// run BEFORE the platform exec seam (REQ-009): the direct door replaces the
// process, so a print after the seam never runs.
// @MX:SPEC: SPEC-CODEX-DEBUG-MODE-001
func (t *factoryLaunchTiming) debugDump(w io.Writer) {
	if t == nil || !t.started {
		return
	}
	for _, step := range t.steps {
		if step.detail != "" {
			_, _ = fmt.Fprintf(w, "%s %s took %s (%s)\n", launcherDebugTracePrefix, step.name, step.elapsed.Round(time.Microsecond), step.detail)
			continue
		}
		_, _ = fmt.Fprintf(w, "%s %s took %s\n", launcherDebugTracePrefix, step.name, step.elapsed.Round(time.Microsecond))
	}
}

// reportSlow prints one line per recorded step when the pre-exec phase
// exceeded the operator-configurable slow-launch threshold (REQ-012). The
// second call is a no-op: the direct exec door reports before the seam and
// the deferred error-path guard never duplicates the lines.
// @MX:NOTE: [AUTO] the report must print BEFORE the platform exec seam — the direct door replaces the process (syscall.Exec), so a print after the seam never runs; the deferred error-path guard stays inert once reported
// @MX:SPEC: SPEC-CODEX-LANE-SLOTS-001
func (t *factoryLaunchTiming) reportSlow(w io.Writer) {
	if t == nil || !t.started || t.reported {
		return
	}
	threshold := factorySlowLaunchThreshold()
	if t.phaseElapsed() < threshold {
		return
	}
	t.reported = true
	_, _ = fmt.Fprintf(w, "factory launch: pre-exec phase took %s (threshold %s) — per-step timings follow\n",
		t.phaseElapsed().Round(time.Millisecond), threshold)
	for _, step := range t.steps {
		_, _ = fmt.Fprintf(w, "factory launch: %s took %s\n", step.name, step.elapsed.Round(time.Millisecond))
	}
}

// factorySlowLaunchThreshold resolves the operator-configurable slow-launch
// threshold (REQ-012): MOAI_FACTORY_SLOW_LAUNCH_MS when it parses to a
// non-negative millisecond count, the config default otherwise.
func factorySlowLaunchThreshold() time.Duration {
	if raw := strings.TrimSpace(os.Getenv(config.EnvMoaiFactorySlowLaunchMS)); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms >= 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return config.DefaultFactorySlowLaunchThreshold
}
