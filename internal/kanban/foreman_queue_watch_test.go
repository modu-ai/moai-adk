// foreman_queue_watch_test.go — SPEC-BACKLOG-JSON-DISCLOSURE-001 (card
// t395): the R1 red, made into a regression (AC-BJD-008..011).
//
// `.moai/reports/t395/r1-repro.md` OBSERVED the shipped foreman queue watch
// producing zero events across a real queue mutation on a migrated project:
// it polls `cksum` on a `backlog.json` that no longer exists, so its `cur`
// is `missing` on every iteration and the change branch is unreachable.
// The watch is dead, and dead silently.
//
// These tests take the watch script VERBATIM from the skill file — the
// local copy and the template mirror both — and run it. The only rebinding
// is the process working directory: the script's paths are relative to a
// project root, so pointing the process at a fixture root rebinds every
// target the script resolves without touching a byte of its text. That
// matters because AC-BJD-010 permits a repair whose watch covers more than
// one path; a rebinding that named a single variable would go undefined the
// moment that repair was chosen.
//
// TestForemanQueueWatch_ShippedJSONTargetIsSilent is the falsifiability
// condition: it pins the pre-repair form and asserts it stays silent on the
// same fixture, in the same window, across the same mutation. A check that
// passed against both targets would assert nothing.
//
// Every armed watch is bounded by a context deadline, so the `while true`
// loop is killed by the runtime rather than by a trailing kill the happy
// path might never reach.
package kanban

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// watchWindow bounds one armed watch. The script's poll interval is 5s and
// is deliberately not rebound (tightening it would be a second edit, and
// the skill body forbids it), so the window must span several polls.
const watchWindow = 16 * time.Second

// legacyForemanWatchScript is the PRE-REPAIR block, pinned verbatim from
// `.claude/skills/moai-kanban-foreman/SKILL.md` as it shipped at
// origin/develop@ad272be20. It exists so the falsifiability condition can
// be demonstrated rather than asserted.
const legacyForemanWatchScript = `f=.moai/state/todo/backlog.json
last=init
while true; do
  if [ -f "$f" ]; then cur=$(cksum "$f"); else cur=missing; fi
  if [ "$cur" != "$last" ]; then
    [ "$last" != init ] && echo "backlog changed"
    last=$cur
  fi
  sleep 5
done`

// foremanSkillPaths are the two copies of the skill that must agree
// (AC-BJD-011). Paths are relative to this package directory.
var foremanSkillPaths = map[string]string{
	"local":    "../../.claude/skills/moai-kanban-foreman/SKILL.md",
	"template": "../../internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md",
}

// extractForemanWatchScript returns the queue-watch shell block from a
// foreman SKILL.md: the first fenced `sh` block after the Queue watch step.
// Fence indentation is stripped so the block runs as written.
func extractForemanWatchScript(t *testing.T, skillPath string) string {
	t.Helper()
	raw, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read %s: %v", skillPath, err)
	}
	lines := strings.Split(string(raw), "\n")
	start := -1
	for i, ln := range lines {
		if strings.Contains(ln, "**Queue watch.**") {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("%s: no '**Queue watch.**' step found", skillPath)
	}
	open, indent := -1, ""
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "```sh" {
			open = i
			indent = lines[i][:strings.Index(lines[i], "```")]
			break
		}
	}
	if open < 0 {
		t.Fatalf("%s: no fenced sh block after the Queue watch step", skillPath)
	}
	var body []string
	for i := open + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "```" {
			return strings.Join(body, "\n")
		}
		body = append(body, strings.TrimPrefix(lines[i], indent))
	}
	t.Fatalf("%s: unterminated sh block after the Queue watch step", skillPath)
	return ""
}

// requirePOSIXWatchTools skips where the watch script's own tools are not
// the platform's. The foreman arms this loop through a POSIX shell; on a
// platform without `sh` and `cksum` there is nothing to observe, and a run
// that pretended otherwise would report a pass it never measured.
func requirePOSIXWatchTools(t *testing.T) {
	t.Helper()
	for _, tool := range []string{"sh", "cksum"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s unavailable on %s: the foreman queue watch is a POSIX shell loop", tool, runtime.GOOS)
		}
	}
}

// watchFixture builds an isolated project root whose queue lives at the
// canonical relative path the watch script resolves, and returns the root
// and a store over that queue. The live primary-checkout queue is never
// read, mutated, or measured by any of this.
func watchFixture(t *testing.T) (string, *BacklogStore) {
	t.Helper()
	root := t.TempDir()
	queueDir := filepath.Join(root, ".moai", "state", "todo")
	if err := os.MkdirAll(queueDir, 0o755); err != nil {
		t.Fatalf("mkdir queue dir: %v", err)
	}
	store := NewBacklogStore(filepath.Join(queueDir, "backlog.json"))
	if _, _, err := store.Add("fixture card one"); err != nil {
		t.Fatalf("seed queue: %v", err)
	}
	return root, store
}

// armWatch writes script to a file, runs it with root as its working
// directory under a watchWindow deadline, mutates the queue inside the
// window, and reports whether a change event was emitted. The deadline is
// what stops the loop: an armed watch never exits on its own.
func armWatch(t *testing.T, root, script string, mutate func()) bool {
	t.Helper()
	scriptPath := filepath.Join(t.TempDir(), "queue-watch.sh")
	if err := os.WriteFile(scriptPath, []byte(script+"\n"), 0o600); err != nil {
		t.Fatalf("write watch script: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), watchWindow)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", scriptPath)
	cmd.Dir = root
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start watch: %v", err)
	}

	// Let the first poll establish `last` before the queue moves.
	time.Sleep(2 * time.Second)
	mutate()

	fired := false
	sc := bufio.NewScanner(stdout)
	for sc.Scan() {
		if strings.Contains(sc.Text(), "backlog changed") {
			fired = true
		}
	}
	_ = cmd.Wait()
	return fired
}

// TestForemanQueueWatch_FiresOnMutation — AC-BJD-008. The migrated layout
// (backlog.db, no backlog.json) is the shape r1-repro.md observed dead.
func TestForemanQueueWatch_FiresOnMutation(t *testing.T) {
	requirePOSIXWatchTools(t)
	for _, name := range []string{"local", "template"} {
		t.Run(name, func(t *testing.T) {
			root, store := watchFixture(t)
			if _, err := os.Stat(filepath.Join(root, ".moai", "state", "todo", "backlog.json")); err == nil {
				t.Fatal("fixture carries a backlog.json — this case is the migrated layout")
			}
			script := extractForemanWatchScript(t, foremanSkillPaths[name])
			fired := armWatch(t, root, script, func() {
				if _, _, err := store.Add("fixture card two — queue mutation under watch"); err != nil {
					t.Errorf("mutate queue: %v", err)
				}
			})
			if !fired {
				t.Errorf("no change event within %s across a real queue mutation — the %s watch does not observe the authoritative store", watchWindow, name)
			}
		})
	}
}

// TestForemanQueueWatch_ShippedJSONTargetIsSilent — AC-BJD-008's
// falsifiability condition, demonstrated: the pre-repair target stays
// silent on the same fixture, same window, same mutation.
func TestForemanQueueWatch_ShippedJSONTargetIsSilent(t *testing.T) {
	requirePOSIXWatchTools(t)
	root, store := watchFixture(t)
	fired := armWatch(t, root, legacyForemanWatchScript, func() {
		if _, _, err := store.Add("fixture card two — queue mutation under watch"); err != nil {
			t.Errorf("mutate queue: %v", err)
		}
	})
	if fired {
		t.Error("the pre-repair backlog.json watch fired — then AC-BJD-008 is vacuous and passes against both targets")
	}
}

// TestForemanQueueWatch_FiresWithStaleJSONPresent — AC-BJD-009, Case B.
// r1-repro.md reasoned this case and did not observe it; this closes that
// recorded Gap. A backlog.json is written beside the database and never
// touched again, so a watch that still keys on it sees a frozen checksum.
func TestForemanQueueWatch_FiresWithStaleJSONPresent(t *testing.T) {
	requirePOSIXWatchTools(t)
	root, store := watchFixture(t)
	jsonPath := filepath.Join(root, ".moai", "state", "todo", "backlog.json")
	if err := os.WriteFile(jsonPath, []byte(legacyBacklogJSON), 0o600); err != nil {
		t.Fatalf("write stale backlog.json: %v", err)
	}
	script := extractForemanWatchScript(t, foremanSkillPaths["local"])
	fired := armWatch(t, root, script, func() {
		if _, _, err := store.Add("fixture card two — queue mutation under watch"); err != nil {
			t.Errorf("mutate queue: %v", err)
		}
	})
	if !fired {
		t.Errorf("no change event within %s with a stale backlog.json present (Case B)", watchWindow)
	}
}

// TestForemanQueueWatch_WatchTargetsAgree — AC-BJD-011: the two copies of
// the script are identical, and neither watches backlog.json.
func TestForemanQueueWatch_WatchTargetsAgree(t *testing.T) {
	local := extractForemanWatchScript(t, foremanSkillPaths["local"])
	template := extractForemanWatchScript(t, foremanSkillPaths["template"])
	if local != template {
		t.Errorf("the local and template queue-watch blocks differ:\n--- local ---\n%s\n--- template ---\n%s", local, template)
	}
	for name, script := range map[string]string{"local": local, "template": template} {
		if strings.Contains(script, "backlog.json") {
			t.Errorf("the %s watch still names backlog.json:\n%s", name, script)
		}
		if !strings.Contains(script, "backlog.db") {
			t.Errorf("the %s watch does not name backlog.db:\n%s", name, script)
		}
	}
}
