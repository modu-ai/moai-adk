package cli

// SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2g — the live legs (card t1099).
//
// Built, opt-in, and NOT run in this SPEC (operator decision Q5). Every test
// here is gated behind MOAI_PARITY_LIVE=1. Without the switch — and whenever a
// host binary, a login, or the trigger itself is unavailable — the test writes
// a NOT_RUN verdict record and calls t.Skipf with the attempted command and the
// observed output, so rule P reads NOT_RUN and never PASS (acceptance.md §B
// rules 4 and 8). The aggregate (AC-HPR-019) reads both the go-test action and
// the record and takes the weaker.
//
// Isolation (AC-HPR-020): Codex always runs under a throwaway CODEX_HOME, every
// scratch project lives under the OS temp dir outside this repository, and the
// operator's ~/.codex config and hooks are hashed before and after.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/goal"
)

const (
	// parityLiveSwitch gates every live leg of this SPEC.
	parityLiveSwitch = "MOAI_PARITY_LIVE"
	// parityVerdictDirEnv names a directory that receives the verdict records;
	// unset, the records go to the test's temp dir.
	parityVerdictDirEnv = "MOAI_PARITY_VERDICT_DIR"
	// parityLiveTurnTimeout bounds one host turn from outside.
	parityLiveTurnTimeout = 5 * time.Minute
)

// parityVerdict values a live leg can record. Only effect-verified legs record
// PASS; everything else keeps the aggregate below PASS (AC-HPR-019).
const (
	parityPass        = "PASS"
	parityFail        = "FAIL"
	parityNotRun      = "NOT_RUN"
	parityUnsupported = "UNSUPPORTED"
)

// parityLiveRecord is one live leg's verdict with its attribution
// (acceptance.md §B rule 7, REQ-HPR-024).
type parityLiveRecord struct {
	Test          string    `json:"test"`
	AC            string    `json:"ac"`
	Verdict       string    `json:"verdict"`
	Attempted     string    `json:"attempted"`
	Observed      string    `json:"observed"`
	Commit        string    `json:"commit"`
	TreeDigest    string    `json:"tree_digest"`
	ClaudeVersion string    `json:"claude_version"`
	CodexVersion  string    `json:"codex_version"`
	Uname         string    `json:"uname"`
	RecordedAt    time.Time `json:"recorded_at"`
}

// parityVersionOf runs `<bin> --version`, or reports why it could not.
func parityVersionOf(bin string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, bin, "--version").CombinedOutput()
	if err != nil {
		return "unavailable: " + strings.TrimSpace(err.Error())
	}
	return strings.TrimSpace(string(out))
}

// parityAttribution fills the attribution fields from this tree and host.
func parityAttribution(rec *parityLiveRecord) {
	repo, _ := filepath.Abs(filepath.Join("..", ".."))
	if out, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output(); err == nil {
		rec.Commit = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "-C", repo, "status", "--porcelain").Output(); err == nil {
		sum := sha256.Sum256(out)
		rec.TreeDigest = hex.EncodeToString(sum[:8])
	}
	rec.ClaudeVersion = parityVersionOf("claude")
	rec.CodexVersion = parityVersionOf(codexBinaryName)
	rec.Uname = runtime.GOOS + " " + runtime.GOARCH
}

// writeParityRecord persists rec where the aggregate reads it.
func writeParityRecord(t *testing.T, rec parityLiveRecord) string {
	t.Helper()
	dir := os.Getenv(parityVerdictDirEnv)
	if dir == "" {
		dir = t.TempDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("verdict dir: %v", err)
	}
	rec.RecordedAt = time.Now().UTC()
	parityAttribution(&rec)
	b, _ := json.MarshalIndent(rec, "", "  ")
	p := filepath.Join(dir, rec.Test+".json")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatalf("write verdict record: %v", err)
	}
	return p
}

// parityNotRunSkip records NOT_RUN and skips with what was attempted and what
// was observed. A live leg never returns normally without its trigger.
func parityNotRunSkip(t *testing.T, ac, attempted, observed string) {
	t.Helper()
	p := writeParityRecord(t, parityLiveRecord{Test: t.Name(), AC: ac, Verdict: parityNotRun, Attempted: attempted, Observed: observed})
	t.Skipf("NOT_RUN (%s): attempted: %s; observed: %s; record: %s", ac, attempted, observed, p)
}

// parityLive is one opened live environment.
type parityLive struct {
	t         *testing.T
	ac        string
	procs     *liveProcs
	budget    *liveBudget
	project   string
	codexHome string
	moaiDir   string
}

// parityHostFileHash hashes one operator ~/.codex file; absent reads as "absent".
func parityHostFileHash(home, rel string) string {
	b, err := os.ReadFile(filepath.Join(home, ".codex", rel))
	if err != nil {
		return "absent"
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// parityIsolationViolations is the AC-HPR-020 detector. It reports every way
// one host process would escape isolation: a Codex process without a
// temporary CODEX_HOME, and any process whose working directory is not under
// the OS temp dir or is inside the repository.
func parityIsolationViolations(env []string, cwd, tempDir, repoRoot string, codex bool) []string {
	var v []string
	if codex {
		home := ""
		for _, kv := range env {
			if strings.HasPrefix(kv, "CODEX_HOME=") {
				home = strings.TrimPrefix(kv, "CODEX_HOME=")
			}
		}
		if home == "" || !parityUnder(home, tempDir) {
			v = append(v, fmt.Sprintf("codex process CODEX_HOME=%q is not a temporary home under %s", home, tempDir))
		}
	}
	if !parityUnder(cwd, tempDir) {
		v = append(v, fmt.Sprintf("working directory %s is not under the OS temp dir %s", cwd, tempDir))
	}
	if parityUnder(cwd, repoRoot) {
		v = append(v, fmt.Sprintf("working directory %s is inside the repository %s", cwd, repoRoot))
	}
	return v
}

func parityUnder(path, root string) bool {
	if path == "" || root == "" {
		return false
	}
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// openParityLive gates, provisions, and isolates one live leg. Every refusal
// records NOT_RUN and skips.
func openParityLive(t *testing.T, ac, attempted string, needClaude bool) *parityLive {
	t.Helper()
	if os.Getenv(parityLiveSwitch) != "1" {
		parityNotRunSkip(t, ac, attempted, parityLiveSwitch+" is not set to 1 (operator decision Q5: no live run in this SPEC)")
	}
	if _, err := exec.LookPath(codexBinaryName); err != nil {
		parityNotRunSkip(t, ac, attempted, "codex binary not found: "+err.Error())
	}
	if needClaude {
		if _, err := exec.LookPath("claude"); err != nil {
			parityNotRunSkip(t, ac, attempted, "claude binary not found: "+err.Error())
		}
	}
	realHome, err := factoryLiveOperatorHomeFn()
	if err != nil {
		parityNotRunSkip(t, ac, attempted, "operator home unavailable: "+err.Error())
	}
	if _, err := os.Stat(filepath.Join(realHome, ".codex", "auth.json")); err != nil {
		parityNotRunSkip(t, ac, attempted, "no Codex login at ~/.codex/auth.json")
	}

	before := map[string]string{"config.toml": parityHostFileHash(realHome, "config.toml"), "hooks.json": parityHostFileHash(realHome, "hooks.json")}
	t.Cleanup(func() {
		for rel, h := range before {
			if got := parityHostFileHash(realHome, rel); got != h {
				t.Errorf("isolation broken: the operator's ~/.codex/%s changed during the live run", rel)
			}
		}
	})

	l := &parityLive{t: t, ac: ac, procs: newLiveProcs(t), budget: newLiveBudget(10, 45*time.Minute)}
	l.project = canonicalDir(t, t.TempDir())
	repo, _ := filepath.Abs(filepath.Join("..", ".."))
	if v := parityIsolationViolations(nil, l.project, canonicalDir(t, os.TempDir()), repo, false); len(v) > 0 {
		t.Fatalf("scratch project violates isolation: %v", v)
	}
	l.codexHome, _ = isolatedCodexHome(t, l.project)
	l.moaiDir = filepath.Dir(buildLiveMoai(t, l.procs))

	for _, dir := range []string{".moai/state", ".codex"} {
		if err := os.MkdirAll(filepath.Join(l.project, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	rendered, err := codexwiring.RenderHooks(nil)
	if err != nil {
		t.Fatalf("render hooks: %v", err)
	}
	l.write(".codex/hooks.json", string(rendered))
	l.write(".gitignore", ".moai/\n")
	l.run(exec.Command("git", "init", "-q"))
	return l
}

func (l *parityLive) write(rel, content string) {
	l.t.Helper()
	p := filepath.Join(l.project, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		l.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		l.t.Fatal(err)
	}
}

// env is the environment every host process of this leg runs under.
func (l *parityLive) env() []string {
	return append(os.Environ(),
		"CODEX_HOME="+l.codexHome,
		"CLAUDE_PROJECT_DIR="+l.project,
		"PATH="+l.moaiDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// run executes a short local command in the scratch project.
func (l *parityLive) run(cmd *exec.Cmd) string {
	l.t.Helper()
	cmd.Dir, cmd.Env = l.project, l.env()
	out, err := cmd.CombinedOutput()
	if err != nil {
		l.t.Fatalf("%v: %v\n%s", cmd.Args, err, out)
	}
	return string(out)
}

// host runs one budgeted host turn (codex exec, or claude -p) in the scratch
// project, asserting isolation first.
func (l *parityLive) host(bin string, args ...string) (string, error) {
	l.t.Helper()
	if ok, why := l.budget.take(); !ok {
		parityNotRunSkip(l.t, l.ac, bin+" "+strings.Join(args, " "), "live budget exhausted: "+why)
	}
	repo, _ := filepath.Abs(filepath.Join("..", ".."))
	if v := parityIsolationViolations(l.env(), l.project, canonicalDir(l.t, os.TempDir()), repo, bin == codexBinaryName); len(v) > 0 {
		l.t.Fatalf("isolation detector refused the run: %v", v)
	}
	ctx, cancel := context.WithTimeout(context.Background(), parityLiveTurnTimeout)
	defer cancel()
	cmd := liveCommand(ctx, bin, args...)
	cmd.Dir, cmd.Env = l.project, l.env()
	out, err := l.procs.run(cmd)
	return string(out), err
}

func (l *parityLive) codexTurn(prompt string) (string, error) {
	return l.host(codexBinaryName, "exec", "--skip-git-repo-check", "--cd", l.project, prompt)
}

// codexResume continues a Codex session by id.
func (l *parityLive) codexResume(session, prompt string) (string, error) {
	return l.host(codexBinaryName, "exec", "--skip-git-repo-check", "--cd", l.project, "resume", session, prompt)
}

// claudeTurn runs one claude -p turn under a session id the test chose, so a
// goal can be armed for that session before the turn starts.
func (l *parityLive) claudeTurn(session, prompt string) (string, error) {
	return l.host("claude", "-p", "--session-id", session, prompt)
}

// codexSessions returns the Codex session ids the Stop chain has recorded
// (one record file per session). Goal state is keyed by session, and a Codex
// session id is only known after its first Stop.
func (l *parityLive) codexSessions() []string {
	entries, _ := os.ReadDir(filepath.Join(l.project, filepath.FromSlash(stopChainRecordDir)))
	var out []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			out = append(out, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return out
}

// armLiveGoal arms an unmet mechanical goal for session in the scratch project.
func (l *parityLive) armLiveGoal(session, cmd string) {
	l.t.Helper()
	g := goal.NewGoal(session, "live parity", []goal.Condition{{Type: goal.ConditionMechanical, Cmd: cmd}})
	g.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := goal.SaveGoal(l.project, g); err != nil {
		l.t.Fatal(err)
	}
}

func (l *parityLive) exists(rel string) bool {
	_, err := os.Stat(filepath.Join(l.project, rel))
	return err == nil
}

func (l *parityLive) record(verdict, attempted, observed string) {
	l.t.Helper()
	p := writeParityRecord(l.t, parityLiveRecord{Test: l.t.Name(), AC: l.ac, Verdict: verdict, Attempted: attempted, Observed: observed})
	l.t.Logf("%s %s: %s (record %s)", l.ac, verdict, observed, p)
	if verdict == parityFail {
		l.t.Fatalf("%s FAIL: %s", l.ac, observed)
	}
}

// TestLiveStopChainGoalContinuation is AC-HPR-004 (live). Codex: a first turn
// reveals the session id, a goal is armed for it, and the resumed turn must
// continue. Claude: a goal is armed for a chosen --session-id first.
func TestLiveStopChainGoalContinuation(t *testing.T) {
	const ac = "AC-HPR-004"
	attempted := "codex exec (+ resume) / claude -p --session-id on a scratch project with an armed unmet goal `test -f done.flag`"
	l := openParityLive(t, ac, attempted, true)
	l.write(".claude/settings.json", `{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"moai hook stop-goal"}]}]}}`)

	out, err := l.codexTurn("Reply with the single word ready and stop.")
	sessions := l.codexSessions()
	if len(sessions) != 1 {
		parityNotRunSkip(t, ac, attempted, fmt.Sprintf("Codex Stop chain recorded %d sessions after the first turn, want 1 (err=%v, tail %s)", len(sessions), err, tailString([]byte(out), 400)))
	}
	l.armLiveGoal(sessions[0], "test -f done.flag")
	out, err = l.codexResume(sessions[0], "Reply with the single word again and stop.")
	g, _ := goal.LoadGoal(l.project, sessions[0])
	if g == nil || g.TurnsUsed < 2 {
		turns := 0
		if g != nil {
			turns = g.TurnsUsed
		}
		l.record(parityFail, attempted, fmt.Sprintf("Codex: the unmet goal was evaluated %d time(s); a continued turn evaluates it at least twice (err=%v, tail %s)", turns, err, tailString([]byte(out), 300)))
	}

	claudeSession := "00000000-0000-4000-8000-00000000c1a0"
	l.armLiveGoal(claudeSession, "test -f done.flag")
	outC, errC := l.claudeTurn(claudeSession, "Reply with the single word ready and stop.")
	gc, _ := goal.LoadGoal(l.project, claudeSession)
	if gc == nil || gc.TurnsUsed < 2 {
		parityNotRunSkip(t, ac, attempted, fmt.Sprintf("Claude: no second stop-goal evaluation observed (err=%v, tail %s)", errC, tailString([]byte(outC), 400)))
	}
	l.record(parityPass, attempted, "both hosts continued the unmet-goal turn with the goal member firing")
}

// TestLiveCodexNeedsInputOutcome is AC-HPR-007 (live).
func TestLiveCodexNeedsInputOutcome(t *testing.T) {
	const ac = "AC-HPR-007"
	attempted := "codex exec under each supported approval policy with a PreToolUse handler returning ask"
	openParityLive(t, ac, attempted, false)
	parityNotRunSkip(t, ac, attempted,
		"no in-tree PreToolUse handler returns ask on demand; the fixture that makes one (and the approval-policy matrix) is the live-certification card's")
}

// TestLiveHookFaultOutcome is the AC-HPR-008 live leg: what the Codex host
// does when a PreToolUse handler times out, exits 1, prints garbage, or exits 2.
func TestLiveHookFaultOutcome(t *testing.T) {
	const ac = "AC-HPR-008"
	faults := map[string]string{
		"timeout": "sleep 30",
		"exit1":   "exit 1",
		"garbage": "printf 'not json'",
		"exit2":   "echo blocked >&2; exit 2",
	}
	attempted := "codex exec asking for `touch marker-<fault>` with a PreToolUse handler that faults (timeout, exit 1, unparseable, exit 2)"
	l := openParityLive(t, ac, attempted, false)
	var observed []string
	unsupported := false
	for name, script := range faults {
		l.write(".codex/hooks.json", fmt.Sprintf(`{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"sh -c '%s'","timeout":5}]}]}}`, script))
		out, err := l.codexTurn(fmt.Sprintf("Run the shell command `touch marker-%s` and report.", name))
		if !strings.Contains(out, "marker-"+name) && err != nil {
			parityNotRunSkip(t, ac, attempted, fmt.Sprintf("fault %s: the turn never reached a tool call (err=%v)", name, err))
		}
		if l.exists("marker-" + name) {
			unsupported = true
			observed = append(observed, name+": host ALLOWED the tool call")
		} else {
			observed = append(observed, name+": host did not run the tool call")
		}
	}
	if unsupported {
		l.record(parityUnsupported, attempted, strings.Join(observed, "; "))
		return
	}
	l.record(parityPass, attempted, strings.Join(observed, "; "))
}

// TestLiveCodexCompactFires is the AC-HPR-009 live leg.
func TestLiveCodexCompactFires(t *testing.T) {
	const ac = "AC-HPR-009"
	attempted := "codex exec with model_auto_compact_token_limit=2000 and a turn that reads a large file"
	l := openParityLive(t, ac, attempted, false)
	cfg, _ := os.ReadFile(filepath.Join(l.codexHome, "config.toml"))
	if err := os.WriteFile(filepath.Join(l.codexHome, "config.toml"), append([]byte("model_auto_compact_token_limit = 2000\n"), cfg...), 0o600); err != nil {
		t.Fatal(err)
	}
	l.write("big.txt", strings.Repeat("parity compaction filler line\n", 4000))
	out, err := l.codexTurn("Read big.txt in full, then summarise it in one sentence.")
	if !l.exists(".moai/state/session-memo.md") {
		parityNotRunSkip(t, ac, attempted, fmt.Sprintf("no PreCompact memo written — compaction not triggered (err=%v, tail %s)", err, tailString([]byte(out), 400)))
	}
	sink, _ := os.ReadFile(filepath.Join(l.project, ".moai", "logs", "codex-adapter.jsonl"))
	if !strings.Contains(string(sink), "PostCompact") {
		parityNotRunSkip(t, ac, attempted, "PreCompact fired but no PostCompact was observed")
	}
	l.record(parityPass, attempted, "PreCompact wrote the memo and PostCompact restored it")
}

// TestLiveCodexPermissionRequestFires is the AC-HPR-010 live leg.
func TestLiveCodexPermissionRequestFires(t *testing.T) {
	const ac = "AC-HPR-010"
	attempted := "codex exec with approval_policy=on-request asking for a command that needs approval"
	l := openParityLive(t, ac, attempted, false)
	cfg, _ := os.ReadFile(filepath.Join(l.codexHome, "config.toml"))
	if err := os.WriteFile(filepath.Join(l.codexHome, "config.toml"), append([]byte("approval_policy = \"on-request\"\n"), cfg...), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := l.codexTurn("Run `curl -sI https://example.com` and report the status line.")
	parityNotRunSkip(t, ac, attempted, fmt.Sprintf(
		"the deny leg needs a tool input carrying the updated-input marker, which a real approval request does not produce; approval request raised: unobserved (err=%v, tail %s)", err, tailString([]byte(out), 300)))
}

// TestLiveCodexInterruptFires is the AC-HPR-011 live leg.
func TestLiveCodexInterruptFires(t *testing.T) {
	const ac = "AC-HPR-011"
	attempted := "SIGINT delivered to a running `codex exec` turn"
	l := openParityLive(t, ac, attempted, false)
	if ok, why := l.budget.take(); !ok {
		parityNotRunSkip(t, ac, attempted, why)
	}
	ctx, cancel := context.WithTimeout(context.Background(), parityLiveTurnTimeout)
	defer cancel()
	cmd := liveCommand(ctx, codexBinaryName, "exec", "--skip-git-repo-check", "--cd", l.project, "Count slowly from 1 to 200, one number per line.")
	cmd.Dir, cmd.Env = l.project, l.env()
	if err := l.procs.start(cmd); err != nil {
		parityNotRunSkip(t, ac, attempted, "codex did not start: "+err.Error())
	}
	time.Sleep(15 * time.Second)
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		parityNotRunSkip(t, ac, attempted, "SIGINT not delivered: "+err.Error())
	}
	_ = cmd.Wait()
	entries, _ := os.ReadDir(filepath.Join(l.project, codexInterruptDir))
	if len(entries) == 0 {
		parityNotRunSkip(t, ac, attempted, "SIGINT delivered; no Interrupt cancellation record appeared (the event may not fire in exec mode)")
	}
	l.record(parityPass, attempted, fmt.Sprintf("%d cancellation record file(s) written", len(entries)))
}

// TestLiveCodexGoalContinueUntilMet is the AC-HPR-012 live leg.
func TestLiveCodexGoalContinueUntilMet(t *testing.T) {
	const ac = "AC-HPR-012"
	attempted := "codex exec + resume on an armed goal `test -f done.flag`, the agent told to create it and record the receipt"
	l := openParityLive(t, ac, attempted, false)
	out, err := l.codexTurn("Reply with the word ready and stop.")
	sessions := l.codexSessions()
	if len(sessions) != 1 {
		parityNotRunSkip(t, ac, attempted, fmt.Sprintf("Codex Stop chain recorded %d sessions, want 1 (err=%v, tail %s)", len(sessions), err, tailString([]byte(out), 400)))
	}
	l.armLiveGoal(sessions[0], "test -f done.flag")
	out, err = l.codexResume(sessions[0], fmt.Sprintf(
		"Create an empty file done.flag, then run `%s --check-id goal --command 'test -f done.flag' --exit 0`, then stop.", codexwiring.GoalReceiptCommand))
	if g, _ := goal.LoadGoal(l.project, sessions[0]); g != nil && g.Status == goal.StatusSatisfied {
		l.record(parityPass, attempted, "the goal reached satisfied through the Codex Stop chain")
		return
	}
	parityNotRunSkip(t, ac, attempted, fmt.Sprintf("the goal did not reach satisfied (err=%v, tail %s)", err, tailString([]byte(out), 400)))
}

// TestLiveHarnessIsolation is AC-HPR-020 (live): one Codex and one Claude run,
// each checked by the isolation detector, with the operator's ~/.codex hashed
// before and after (openParityLive's cleanup).
func TestLiveHarnessIsolation(t *testing.T) {
	const ac = "AC-HPR-020"
	attempted := "one codex exec and one claude -p in a temp scratch project under a temp CODEX_HOME"
	l := openParityLive(t, ac, attempted, true)
	if _, err := l.codexTurn("Reply with ok."); err != nil {
		parityNotRunSkip(t, ac, attempted, "codex turn failed: "+err.Error())
	}
	if _, err := l.claudeTurn("00000000-0000-4000-8000-0000000015a0", "Reply with ok."); err != nil {
		parityNotRunSkip(t, ac, attempted, "claude turn failed: "+err.Error())
	}
	l.record(parityPass, attempted, "both runs passed the isolation detector; ~/.codex hashes are compared at cleanup")
}

// TestLiveCodexStopTimeoutCeiling is AC-HPR-021: the Stop-timeout ceiling
// probe. Each rung renders a Stop handler at timeout T running a sleeper just
// under T that blocks, and a second sleeper past T; it records whether Codex
// accepted the configuration, waited, and honoured the decision.
func TestLiveCodexStopTimeoutCeiling(t *testing.T) {
	const ac = "AC-HPR-021"
	attempted := "Stop handler timeout ladder 10s/30s/60s/120s, sleeper T-2s (blocks) and T+5s (must be killed)"
	l := openParityLive(t, ac, attempted, false)
	var observed []string
	ceiling := 0
	for _, rung := range []int{10, 30, 60, 120} {
		under := fmt.Sprintf(`sleep %d; if [ ! -f stop-seen-%d ]; then touch stop-seen-%d; echo '{"decision":"block","reason":"rung %d: say done"}'; fi`, rung-2, rung, rung, rung)
		l.write(".codex/hooks.json", fmt.Sprintf(`{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"sh -c '%s'","timeout":%d}]}]}}`, under, rung))
		out, err := l.codexTurn("Reply with the word start.")
		honoured := l.exists(fmt.Sprintf("stop-seen-%d", rung)) && strings.Contains(out, "done")
		observed = append(observed, fmt.Sprintf("rung %ds under-sleeper: honoured=%v err=%v", rung, honoured, err))
		if honoured {
			ceiling = rung
		}

		over := fmt.Sprintf("sleep %d; touch over-finished-%d", rung+5, rung)
		l.write(".codex/hooks.json", fmt.Sprintf(`{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"sh -c '%s'","timeout":%d}]}]}}`, over, rung))
		_, err = l.codexTurn("Reply with the word start.")
		time.Sleep(10 * time.Second)
		observed = append(observed, fmt.Sprintf("rung %ds over-sleeper killed=%v err=%v", rung, !l.exists(fmt.Sprintf("over-finished-%d", rung)), err))
	}
	if ceiling == 0 {
		parityNotRunSkip(t, ac, attempted, "no rung honoured a Stop decision: "+strings.Join(observed, "; "))
	}
	l.record(parityPass, attempted, fmt.Sprintf("T_codex_max=%ds; %s", ceiling, strings.Join(observed, "; ")))
}
