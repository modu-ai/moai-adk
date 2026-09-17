package hook

// Card t666 R3 — does the SessionStart hook, run as the real short-lived
// process, ever populate the drift cache?
//
// A measurement instrument, not a check: skipped unless
// MOAI_DRIFT_CACHE_PROBE_BIN (a built moai binary) and
// MOAI_DRIFT_CACHE_PROBE_SRC (a git repository with a real SPEC tree) are set,
// and it never fails on the answer it measures.
//
// Each iteration clones the source (git clone --shared: history objects are
// borrowed, nothing is pushed back) and runs two arms in alternating order:
//
//	exit     `moai hook session-start` three times in a row. The process exits
//	         when Handle returns, exactly as under Claude Code.
//	control  `moai spec drift` first — a process that waits for the drift scan
//	         to finish and save — then `moai hook session-start` once.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSessionStart_DriftCacheProbe(t *testing.T) {
	bin := os.Getenv("MOAI_DRIFT_CACHE_PROBE_BIN")
	src := os.Getenv("MOAI_DRIFT_CACHE_PROBE_SRC")
	if bin == "" || src == "" {
		t.Skip("measurement probe: set MOAI_DRIFT_CACHE_PROBE_BIN and MOAI_DRIFT_CACHE_PROBE_SRC to run")
	}
	n := 2
	if raw := os.Getenv("MOAI_DRIFT_CACHE_PROBE_N"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			t.Fatalf("MOAI_DRIFT_CACHE_PROBE_N=%q: want a positive integer", raw)
		}
		n = v
	}
	moaiHome := t.TempDir()

	// Child environment: no ambient git location, no kanban/factory identity,
	// no Claude profile, a private MOAI_HOME.
	childEnv := func(projectDir string) []string {
		env := make([]string, 0, len(os.Environ())+2)
		for _, kv := range os.Environ() {
			name, _, _ := strings.Cut(kv, "=")
			switch {
			case strings.HasPrefix(name, "GIT_"),
				strings.HasPrefix(name, "MOAI_KANBAN"),
				strings.HasPrefix(name, "MOAI_FACTORY"),
				name == "MOAI_HOME", name == "CLAUDE_PROJECT_DIR", name == "CLAUDE_CONFIG_DIR":
				continue
			}
			env = append(env, kv)
		}
		return append(env, "MOAI_HOME="+moaiHome, "CLAUDE_PROJECT_DIR="+projectDir)
	}

	clone := func(label string) string {
		dst := filepath.Join(t.TempDir(), label)
		cmd := exec.Command("git", "clone", "--shared", "--quiet", src, dst)
		cmd.Env = childEnv(dst)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("clone %s: %v\n%s", label, err, out)
		}
		return dst
	}

	head := func(dir string) string {
		cmd := exec.Command("git", "rev-parse", "HEAD")
		cmd.Dir = dir
		cmd.Env = childEnv(dir)
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("rev-parse in %s: %v", dir, err)
		}
		return strings.TrimSpace(string(out))
	}

	// cacheState reads .moai/state/drift-cache.json and reports whether it
	// exists and is keyed on the clone's HEAD.
	cacheState := func(dir string) string {
		raw, err := os.ReadFile(filepath.Join(dir, ".moai", "state", driftCacheFileForProbe))
		if err != nil {
			return "absent"
		}
		var payload struct {
			HeadSHA string `json:"head_sha"`
		}
		if json.Unmarshal(raw, &payload) != nil {
			return "corrupt"
		}
		if payload.HeadSHA == head(dir) {
			return "present(head-match)"
		}
		return "present(stale)"
	}

	// runHook returns the process wall time and a git timeline: each git
	// subprocess the hook started, as milliseconds after the hook process was
	// launched. `git log` is the drift scan's cache-miss pass.
	runHook := func(dir string, k int) (time.Duration, string) {
		stdin, _ := json.Marshal(map[string]string{
			"session_id":      fmt.Sprintf("t666-r3-%d-%d", time.Now().UnixNano(), k),
			"cwd":             dir,
			"hook_event_name": "SessionStart",
			"source":          "startup",
		})
		tracePath := filepath.Join(t.TempDir(), "git-trace.log")
		cmd := exec.Command(bin, "hook", "session-start")
		cmd.Dir = dir
		cmd.Env = append(childEnv(dir), "GIT_TRACE="+tracePath)
		cmd.Stdin = bytes.NewReader(stdin)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		start := time.Now()
		if err := cmd.Run(); err != nil {
			t.Fatalf("hook session-start in %s: %v\n%s", dir, err, stderr.String())
		}
		elapsed := time.Since(start)

		raw, _ := os.ReadFile(tracePath)
		var timeline []string
		for line := range strings.SplitSeq(string(raw), "\n") {
			stamp, rest, ok := strings.Cut(line, " ")
			if !ok || !strings.Contains(rest, "trace: built-in: git ") {
				continue
			}
			at, err := time.ParseInLocation("15:04:05.000000", stamp, time.Local)
			if err != nil {
				continue
			}
			at = time.Date(start.Year(), start.Month(), start.Day(),
				at.Hour(), at.Minute(), at.Second(), at.Nanosecond(), time.Local)
			_, argv, _ := strings.Cut(rest, "trace: built-in: git ")
			timeline = append(timeline, fmt.Sprintf("%q@%dms", strings.TrimSpace(argv), at.Sub(start).Milliseconds()))
		}
		return elapsed, strings.Join(timeline, " ")
	}

	runDrift := func(dir string) time.Duration {
		cmd := exec.Command(bin, "spec", "drift")
		cmd.Dir = dir
		cmd.Env = childEnv(dir)
		start := time.Now()
		// A non-zero exit is how drift reports findings; only the process
		// finishing matters here.
		_ = cmd.Run()
		return time.Since(start)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "MEASUREMENT drift-cache probe n=%d bin=%s\n", n, bin)
	for i := range n {
		arms := []string{"exit", "control"}
		if i%2 == 1 {
			arms = []string{"control", "exit"}
		}
		for _, arm := range arms {
			dir := clone(fmt.Sprintf("%s-%d", arm, i))
			fmt.Fprintf(&b, "iter=%d arm=%s before=%s\n", i, arm, cacheState(dir))
			switch arm {
			case "exit":
				for k := 1; k <= 3; k++ {
					d, timeline := runHook(dir, k)
					fmt.Fprintf(&b, "iter=%d arm=exit hook#%d elapsed=%v cache=%s git=[%s]\n",
						i, k, d.Round(time.Millisecond), cacheState(dir), timeline)
				}
			case "control":
				d := runDrift(dir)
				fmt.Fprintf(&b, "iter=%d arm=control spec-drift elapsed=%v cache=%s\n",
					i, d.Round(time.Millisecond), cacheState(dir))
				h, timeline := runHook(dir, 1)
				fmt.Fprintf(&b, "iter=%d arm=control hook#1 elapsed=%v cache=%s git=[%s]\n",
					i, h.Round(time.Millisecond), cacheState(dir), timeline)
			}
		}
	}
	t.Log("\n" + b.String())
}

// driftCacheFileForProbe mirrors internal/spec's unexported driftCacheFilename.
const driftCacheFileForProbe = "drift-cache.json"
