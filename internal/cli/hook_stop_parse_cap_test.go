package cli

// SPEC-HOOK-STOP-PARSE-CAP-001 — behaviour tests for moai's own cap on
// consecutive stdin-parse-failure Stops under the Claude harness.
//
// Every expected value here is written from the SPEC, never read back from the
// implementation: N (8), the expiry (60 minutes), the record file-name rule
// (sha256 of "<kind>:<value>", lowercase hex, ".json"), and both harness
// reasons are literals. None of these tests runs in parallel — they set
// process environment and swap package seams.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/session"
)

const (
	// capN and capExpiry are the SPEC's values (REQ-SPC-004, REQ-SPC-007).
	capN      = 8
	capExpiry = 60 * time.Minute

	// capBroken is the representative malformed stdin of acceptance.md §0.
	capBroken = "not json{"

	// capStateRel is the state area of plan.md §C 4.
	capStateRel = ".moai/state/stop-parse-cap"

	// claudeFailClosedReasonLiteral is REQ-SPC-010, byte for byte.
	claudeFailClosedReasonLiteral = "fail-closed: hook stdin could not be parsed as JSON. Do not edit hook scripts or settings files to get past this; stop and tell a human operator (.moai/docs/hook-stdin-fail-closed.md)"
	// codexFailClosedReasonLiteral is the unchanged Codex reason (REQ-SPC-009).
	codexFailClosedReasonLiteral = "fail-closed: hook stdin could not be parsed as JSON (.moai/docs/hook-stdin-fail-closed.md)"

	// capReleasedMarker identifies the released-cap stderr line.
	capReleasedMarker = "stop-parse cap released"
	// capNotAppliedMarker identifies the could-not-apply stderr line.
	capNotAppliedMarker = "stop-parse cap not applied"
)

var capRecordNameRE = regexp.MustCompile(`^[0-9a-f]{64}\.json$`)

// --- key control -----------------------------------------------------------

// unsetEnvForTest removes key for the duration of the test and restores it.
func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
}

// useSessionKey makes the session id the counting key.
func useSessionKey(t *testing.T, id string) {
	t.Helper()
	t.Setenv(config.EnvClaudeCodeSessionID, id)
	unsetEnvForTest(t, config.EnvMoaiSessionPID)
}

// capTree is one node of a synthetic process tree.
type capTree map[int]struct {
	ppid int
	comm string
}

// viewOf builds a session.ProcessView over tree where every listed process is
// alive; the real ancestry walk runs over it.
func viewOf(self int, tree capTree) session.ProcessView {
	return session.ProcessView{
		Self: self,
		Info: func(pid int) (int, string, bool) {
			p, ok := tree[pid]
			if !ok {
				return 0, "", false
			}
			return p.ppid, p.comm, true
		},
		Alive: func(pid int) bool {
			_, ok := tree[pid]
			return ok
		},
	}
}

// useProcessView swaps the ancestry the fallback key walks.
func useProcessView(t *testing.T, v session.ProcessView) {
	t.Helper()
	orig := stopParseCapProcessView
	stopParseCapProcessView = func() session.ProcessView { return v }
	t.Cleanup(func() { stopParseCapProcessView = orig })
}

// useProcessKey clears the session id and walks tree from self.
func useProcessKey(t *testing.T, self int, tree capTree) {
	t.Helper()
	t.Setenv(config.EnvClaudeCodeSessionID, "")
	unsetEnvForTest(t, config.EnvMoaiSessionPID)
	useProcessView(t, viewOf(self, tree))
}

// hostTree is moai 400 → bash 300 → bash 200 → claude 100.
func hostTree() capTree {
	return capTree{
		400: {300, "moai"},
		300: {200, "bash"},
		200: {100, "bash"},
		100: {50, "claude"},
		50:  {1, "zsh"},
	}
}

// --- running ---------------------------------------------------------------

// newCapRoot points the hook at a fresh project root and returns it.
func newCapRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	return root
}

// runHookHere runs one hook subcommand like runHookWithStdin, but in the
// project root the environment already names, so several calls share one
// state area.
func runHookHere(t *testing.T, sub string, args []string, harness string, stdin []byte) hookRunResult {
	t.Helper()
	spy := &spyRegistry{}
	proto := &readCountingProtocol{Protocol: hook.NewProtocol()}
	origDeps := deps
	deps = &Dependencies{HookRegistry: spy, HookProtocol: proto}
	defer func() { deps = origDeps }()

	cmd := findHookSubcommand(t, sub)
	if err := cmd.ParseFlags([]string{"--harness=" + harness}); err != nil {
		t.Fatalf("parse --harness=%s: %v", harness, err)
	}
	defer func() { _ = cmd.Flags().Set("harness", "") }()
	cmd.SetContext(context.Background())

	path := filepath.Join(t.TempDir(), "stdin")
	if err := os.WriteFile(path, stdin, 0o600); err != nil {
		t.Fatalf("write stdin fixture: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open stdin fixture: %v", err)
	}
	origIn := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = origIn; _ = f.Close() }()

	var runErr error
	stdout, stderr := captureStdoutStderr(t, func() { runErr = cmd.RunE(cmd, args) })
	return hookRunResult{stdout: stdout, stderr: stderr, err: runErr, dispatched: len(spy.dispatched), reads: proto.reads, root: os.Getenv("CLAUDE_PROJECT_DIR")}
}

func runStop(t *testing.T, stdin string) hookRunResult {
	t.Helper()
	return runHookHere(t, "stop", nil, "", []byte(stdin))
}

func validPayload(ev hook.EventType, root string) string {
	switch ev {
	case hook.EventPreToolUse, hook.EventPostToolUse:
		return `{"hook_event_name":"` + string(ev) + `","session_id":"s","cwd":"` + root + `","tool_name":"Bash","tool_input":{"command":"true"}}`
	default:
		return `{"hook_event_name":"` + string(ev) + `","session_id":"s","cwd":"` + root + `"}`
	}
}

// --- expectations ----------------------------------------------------------

func claudeDenyBytes(t *testing.T, ev hook.EventType) string {
	t.Helper()
	row, ok := codexadapter.Lookup(codexadapter.HarnessClaude, ev, codexadapter.DecisionFatalError)
	if !ok {
		t.Fatalf("no Claude fatal_error row for %s", ev)
	}
	out, err := codexadapter.Render(ev, row.Outcome, claudeFailClosedReasonLiteral)
	if err != nil {
		t.Fatalf("render %s: %v", ev, err)
	}
	return string(out) + "\n"
}

func codexDenyBytes(t *testing.T, ev hook.EventType) string {
	t.Helper()
	out, _, err := codexadapter.TranslateCodex(ev, codexadapter.DecisionFatalError, codexFailClosedReasonLiteral)
	if err != nil {
		t.Fatalf("translate %s: %v", ev, err)
	}
	return string(out) + "\n"
}

const capReleasedOutput = "{}\n"

// --- state inspection ------------------------------------------------------

func capRecordName(kind, value string) string {
	sum := sha256.Sum256([]byte(kind + ":" + value))
	return hex.EncodeToString(sum[:]) + ".json"
}

func capRecordPath(root, kind, value string) string {
	return filepath.Join(root, capStateRel, capRecordName(kind, value))
}

type capRecordOnDisk struct {
	Count     int       `json:"count"`
	UpdatedAt time.Time `json:"updated_at"`
	KeyKind   string    `json:"key_kind"`
}

// capCount returns the consecutive count recorded for kind:value, or -1 when
// no parseable record exists.
func capCount(t *testing.T, root, kind, value string) int {
	t.Helper()
	b, err := os.ReadFile(capRecordPath(root, kind, value))
	if err != nil {
		return -1
	}
	var r capRecordOnDisk
	if err := json.Unmarshal(b, &r); err != nil {
		return -1
	}
	return r.Count
}

func writeCapRecord(t *testing.T, path string, count int, age time.Duration, kind string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	b, err := json.Marshal(capRecordOnDisk{Count: count, UpdatedAt: time.Now().Add(-age), KeyKind: kind})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write record: %v", err)
	}
}

// capRecordFiles lists the record-shaped entries of the state area.
func capRecordFiles(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, capStateRel))
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if capRecordNameRE.MatchString(e.Name()) {
			names = append(names, e.Name())
		}
	}
	return names
}

func linesContaining(s, marker string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.Contains(l, marker) {
			out = append(out, l)
		}
	}
	return out
}

// sinkRecords returns the codex-adapter sink records of root.
func sinkRecords(t *testing.T, root string) []codexadapter.Discard {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, codexadapter.DiagnosticSinkRel))
	if err != nil {
		return nil
	}
	var out []codexadapter.Discard
	for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if l == "" {
			continue
		}
		var d codexadapter.Discard
		if err := json.Unmarshal([]byte(l), &d); err != nil {
			t.Fatalf("sink line %q: %v", l, err)
		}
		out = append(out, d)
	}
	return out
}

// --- AC-SPC-001 / 002 -------------------------------------------------------

// TestStopParseCap_AC001_DenyUpToN — REQ-SPC-001/002: the first N consecutive
// parse-failure Stops keep the fail-closed deny, and the k-th call leaves the
// count at k. The four broken forms run twice.
func TestStopParseCap_AC001_DenyUpToN(t *testing.T) {
	root := newCapRoot(t)
	useSessionKey(t, "ac001")
	forms := brokenStdinForms(t, hook.EventStop)
	want := claudeDenyBytes(t, hook.EventStop)

	for k := 1; k <= capN; k++ {
		r := runHookHere(t, "stop", nil, "", forms[(k-1)%len(forms)].payload)
		if r.err != nil {
			t.Fatalf("call %d: RunE = %v, want nil", k, r.err)
		}
		if r.dispatched != 0 {
			t.Errorf("call %d: dispatched %d times, want 0", k, r.dispatched)
		}
		if r.stdout != want {
			t.Errorf("call %d: stdout = %q, want the Claude Stop deny %q", k, r.stdout, want)
		}
		if got := capCount(t, root, "session", "ac001"); got != k {
			t.Errorf("call %d: recorded count = %d, want %d", k, got, k)
		}
	}
}

// TestStopParseCap_AC002_ReleaseAboveN — REQ-SPC-003: calls 9 and 10 get the
// no-opinion answer, one stderr line with the count, N and key kind, and one
// record under a key distinct from the three existing ones.
func TestStopParseCap_AC002_ReleaseAboveN(t *testing.T) {
	root := newCapRoot(t)
	useSessionKey(t, "ac002")
	for k := 1; k <= capN; k++ {
		runStop(t, capBroken)
	}

	var keys []string
	for k := capN + 1; k <= capN+2; k++ {
		before := len(sinkRecords(t, root))
		r := runStop(t, capBroken)
		if r.stdout != capReleasedOutput {
			t.Errorf("call %d: stdout = %q, want %q", k, r.stdout, capReleasedOutput)
		}
		if r.dispatched != 0 {
			t.Errorf("call %d: dispatched %d times, want 0", k, r.dispatched)
		}
		lines := linesContaining(r.stderr, capReleasedMarker)
		if len(lines) != 1 {
			t.Fatalf("call %d: %d released lines in stderr, want exactly 1:\n%s", k, len(lines), r.stderr)
		}
		for _, tok := range []string{"on Stop", "harness claude", "consecutive parse failures " + strconv.Itoa(k), "N=8", "key: session"} {
			if !strings.Contains(lines[0], tok) {
				t.Errorf("call %d: released line %q lacks %q", k, lines[0], tok)
			}
		}
		recs := sinkRecords(t, root)
		if len(recs) != before+1 {
			t.Fatalf("call %d: %d new sink records, want 1", k, len(recs)-before)
		}
		key := recs[len(recs)-1].Key
		for _, taken := range []string{stdinParseFailClosedDiscardKey, stdinParseExemptDiscardKey, hookFaultDiscardKey} {
			if key == taken {
				t.Errorf("call %d: record key %q collides with an existing key", k, key)
			}
		}
		keys = append(keys, key)
		if got := capCount(t, root, "session", "ac002"); got != k {
			t.Errorf("call %d: recorded count = %d, want %d", k, got, k)
		}
	}
	if keys[0] != keys[1] {
		t.Errorf("record keys differ across released calls: %q vs %q", keys[0], keys[1])
	}
}

// --- AC-SPC-003 -------------------------------------------------------------

// TestStopParseCap_AC003_ParsedStopResets — REQ-SPC-005.
func TestStopParseCap_AC003_ParsedStopResets(t *testing.T) {
	t.Run("after eight", func(t *testing.T) {
		root := newCapRoot(t)
		useSessionKey(t, "ac003")
		for k := 1; k <= capN; k++ {
			runStop(t, capBroken)
		}
		r := runStop(t, validPayload(hook.EventStop, root))
		if r.dispatched != 1 {
			t.Errorf("(a) valid Stop dispatched %d times, want 1", r.dispatched)
		}
		if _, err := os.Lstat(capRecordPath(root, "session", "ac003")); !os.IsNotExist(err) {
			t.Errorf("(b) record still present after a parsed Stop (Lstat err = %v)", err)
		}
		r = runStop(t, capBroken)
		if r.stdout != claudeDenyBytes(t, hook.EventStop) {
			t.Errorf("(c) stdout = %q, want the deny", r.stdout)
		}
		if got := capCount(t, root, "session", "ac003"); got != 1 {
			t.Errorf("(c) recorded count = %d, want 1", got)
		}
	})

	t.Run("variant 1 released state", func(t *testing.T) {
		root := newCapRoot(t)
		useSessionKey(t, "ac003v1")
		for k := 1; k <= capN+2; k++ {
			runStop(t, capBroken)
		}
		runStop(t, validPayload(hook.EventStop, root))
		if r := runStop(t, capBroken); r.stdout != claudeDenyBytes(t, hook.EventStop) {
			t.Errorf("stdout = %q, want the deny after a parsed Stop reset the released state", r.stdout)
		}
	})

	t.Run("variant 2 no record creates nothing", func(t *testing.T) {
		root := newCapRoot(t)
		useSessionKey(t, "ac003v2")
		runStop(t, validPayload(hook.EventStop, root))
		if _, err := os.Lstat(filepath.Join(root, capStateRel)); !os.IsNotExist(err) {
			t.Errorf("(d) state area exists after a parsed Stop with no record (Lstat err = %v)", err)
		}
	})

	t.Run("variant 3 delete failure", func(t *testing.T) {
		root := newCapRoot(t)
		useSessionKey(t, "ac003v3")
		baseline := runStop(t, validPayload(hook.EventStop, root))
		for k := 1; k <= capN; k++ {
			runStop(t, capBroken)
		}
		orig := stopParseCapRemove
		stopParseCapRemove = func(string) error { return errors.New("injected remove failure") }
		t.Cleanup(func() { stopParseCapRemove = orig })

		r := runStop(t, validPayload(hook.EventStop, root))
		if r.dispatched != 1 || r.stdout != baseline.stdout {
			t.Errorf("(e) dispatched=%d stdout=%q, want 1 and %q", r.dispatched, r.stdout, baseline.stdout)
		}
		if lines := linesContaining(r.stderr, "injected remove failure"); len(lines) != 1 {
			t.Errorf("(e) %d stderr lines carry the delete failure, want exactly 1:\n%s", len(lines), r.stderr)
		}
	})
}

// --- AC-SPC-004 -------------------------------------------------------------

// TestStopParseCap_AC004_ToolUseDoesNotReset — REQ-SPC-006.
func TestStopParseCap_AC004_ToolUseDoesNotReset(t *testing.T) {
	root := newCapRoot(t)
	useSessionKey(t, "ac004")
	for k := 1; k <= 5; k++ {
		runStop(t, capBroken)
	}
	steps := []struct {
		ev    hook.EventType
		input string
	}{
		{hook.EventPreToolUse, validPayload(hook.EventPreToolUse, root)},
		{hook.EventPreToolUse, capBroken},
		{hook.EventPostToolUse, validPayload(hook.EventPostToolUse, root)},
		{hook.EventPostToolUse, capBroken},
	}
	for i, s := range steps {
		r := runHookHere(t, subcommandFor(t, s.ev), nil, "", []byte(s.input))
		if got := capCount(t, root, "session", "ac004"); got != 5 {
			t.Errorf("(a) after step %d (%s): count = %d, want 5", i+1, s.ev, got)
		}
		if s.ev == hook.EventPreToolUse && s.input == capBroken && r.stdout != claudeDenyBytes(t, hook.EventPreToolUse) {
			t.Errorf("(c) broken PreToolUse stdout = %q, want the fail-closed deny", r.stdout)
		}
	}
	for k := 6; k <= 9; k++ {
		r := runStop(t, capBroken)
		want := claudeDenyBytes(t, hook.EventStop)
		if k == 9 {
			want = capReleasedOutput
		}
		if r.stdout != want {
			t.Errorf("(b) cumulative call %d: stdout = %q, want %q", k, r.stdout, want)
		}
	}
}

// --- AC-SPC-005 -------------------------------------------------------------

// TestStopParseCap_AC005_IndependentOfHostCap — REQ-SPC-004.
func TestStopParseCap_AC005_IndependentOfHostCap(t *testing.T) {
	var runs [][]string
	for i, v := range []string{"<unset>", "8", "200"} {
		newCapRoot(t)
		useSessionKey(t, "ac005-"+strconv.Itoa(i))
		if v == "<unset>" {
			unsetEnvForTest(t, config.EnvClaudeCodeStopHookBlockCap)
		} else {
			t.Setenv(config.EnvClaudeCodeStopHookBlockCap, v)
		}
		var outs []string
		for k := 1; k <= capN+2; k++ {
			outs = append(outs, runStop(t, capBroken).stdout)
		}
		runs = append(runs, outs)
	}
	for i := 1; i < len(runs); i++ {
		for k := range runs[0] {
			if runs[i][k] != runs[0][k] {
				t.Errorf("environment %d call %d: stdout %q differs from the unset environment's %q", i, k+1, runs[i][k], runs[0][k])
			}
		}
	}
}

// TestStopParseCap_StaticChecks covers the source-level checks of AC-SPC-005
// and AC-SPC-007 (iii) over exactly the two named files. A missing second file
// is undeterminable, not a pass.
func TestStopParseCap_StaticChecks(t *testing.T) {
	files := []string{"hook_stdin_failclosed.go", "hook_stop_parse_cap.go"}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("%s unreadable — the static check is undeterminable: %v", f, err)
		}
		src := string(b)
		for _, forbidden := range []string{"EnvClaudeCodeStopHookBlockCap", "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP", "Getppid"} {
			if n := strings.Count(src, forbidden); n != 0 {
				t.Errorf("%s contains %q %d times, want 0", f, forbidden, n)
			}
		}
	}
}

// --- AC-SPC-006 -------------------------------------------------------------

// TestStopParseCap_AC006_Expiry — REQ-SPC-007.
func TestStopParseCap_AC006_Expiry(t *testing.T) {
	stale := capExpiry + time.Minute

	t.Run("expired own record and sweep", func(t *testing.T) {
		root := newCapRoot(t)
		useSessionKey(t, "ac006-K")
		writeCapRecord(t, capRecordPath(root, "session", "ac006-K"), capN, stale, "session")
		writeCapRecord(t, capRecordPath(root, "session", "ac006-K2"), 3, stale, "session")

		r := runStop(t, capBroken)
		if r.stdout != claudeDenyBytes(t, hook.EventStop) {
			t.Errorf("(a) stdout = %q, want the deny", r.stdout)
		}
		if got := capCount(t, root, "session", "ac006-K"); got != 1 {
			t.Errorf("(b) count = %d, want 1", got)
		}
		if _, err := os.Lstat(capRecordPath(root, "session", "ac006-K2")); !os.IsNotExist(err) {
			t.Errorf("(c) expired K2 record survived (Lstat err = %v)", err)
		}
	})

	t.Run("contrast released state is inherited", func(t *testing.T) {
		root := newCapRoot(t)
		useProcessKey(t, 400, hostTree())
		writeCapRecord(t, capRecordPath(root, "process", "100"), capN+1, capExpiry-time.Minute, "process")
		if r := runStop(t, capBroken); r.stdout != capReleasedOutput {
			t.Errorf("first call stdout = %q, want %q — a live released record is inherited", r.stdout, capReleasedOutput)
		}
		if got := capCount(t, root, "process", "100"); got != capN+2 {
			t.Errorf("count = %d, want %d", got, capN+2)
		}
	})

	t.Run("(d) sweep scope", func(t *testing.T) {
		root := newCapRoot(t)
		useSessionKey(t, "ac006d-K")
		dir := filepath.Join(root, capStateRel)
		one := capRecordPath(root, "session", "ac006d-K2")
		writeCapRecord(t, one, 2, stale, "session")
		two := []string{filepath.Join(dir, "123.json"), filepath.Join(dir, "notes.txt")}
		for _, p := range two {
			writeCapRecord(t, p, 2, stale, "session")
		}
		outside := filepath.Join(t.TempDir(), "outside.json")
		writeCapRecord(t, outside, 2, stale, "session")
		outsideBefore, _ := os.ReadFile(outside)
		link := capRecordPath(root, "session", "ac006d-link")
		if err := os.Symlink(outside, link); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		four := capRecordPath(root, "session", "ac006d-garbage")
		if err := os.WriteFile(four, []byte("{"), 0o644); err != nil {
			t.Fatal(err)
		}

		runStop(t, capBroken)

		if _, err := os.Lstat(one); !os.IsNotExist(err) {
			t.Errorf("① expired record survived (Lstat err = %v)", err)
		}
		for _, p := range append(two, four) {
			if _, err := os.Lstat(p); err != nil {
				t.Errorf("%s was removed: %v", filepath.Base(p), err)
			}
		}
		if fi, err := os.Lstat(link); err != nil || fi.Mode()&os.ModeSymlink == 0 {
			t.Errorf("③ link removed or replaced (err = %v)", err)
		}
		if after, err := os.ReadFile(outside); err != nil || string(after) != string(outsideBefore) {
			t.Errorf("③ link target changed or removed (err = %v)", err)
		}
	})

	t.Run("(e) symlinked state area", func(t *testing.T) {
		root := newCapRoot(t)
		useSessionKey(t, "ac006e")
		target := t.TempDir()
		victim := filepath.Join(target, capRecordName("session", "ac006e-victim"))
		writeCapRecord(t, victim, 2, stale, "session")
		before, _ := os.ReadFile(victim)
		if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(root, capStateRel)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}

		r := runStop(t, capBroken)
		if after, err := os.ReadFile(victim); err != nil || string(after) != string(before) {
			t.Errorf("record behind the link was removed or changed (err = %v)", err)
		}
		if r.stdout != claudeDenyBytes(t, hook.EventStop) {
			t.Errorf("stdout = %q, want the deny", r.stdout)
		}
		if len(linesContaining(r.stderr, capNotAppliedMarker)) != 1 {
			t.Errorf("want one %q line:\n%s", capNotAppliedMarker, r.stderr)
		}
		entries, _ := os.ReadDir(target)
		if len(entries) != 1 {
			t.Errorf("link target gained entries: %d, want 1", len(entries))
		}
	})
}

// --- AC-SPC-007 -------------------------------------------------------------

// TestStopParseCap_AC007_KeyResolution — REQ-SPC-014/015/008.
func TestStopParseCap_AC007_KeyResolution(t *testing.T) {
	t.Run("(i) session key first", func(t *testing.T) {
		root := newCapRoot(t)
		useSessionKey(t, "sess-A")
		var last hookRunResult
		for k := 1; k <= capN+1; k++ {
			last = runStop(t, capBroken)
		}
		if last.stdout != capReleasedOutput {
			t.Errorf("9th stdout = %q, want %q", last.stdout, capReleasedOutput)
		}
		if l := linesContaining(last.stderr, capReleasedMarker); len(l) != 1 || !strings.Contains(l[0], "key: session") {
			t.Errorf("released line should name the session key kind: %q", l)
		}
		useSessionKey(t, "sess-B")
		if r := runStop(t, capBroken); r.stdout != claudeDenyBytes(t, hook.EventStop) {
			t.Errorf("sess-B stdout = %q, want the deny", r.stdout)
		}
		if got := capCount(t, root, "session", "sess-B"); got != 1 {
			t.Errorf("sess-B count = %d, want 1", got)
		}

		useSessionKey(t, "sess-C")
		useProcessView(t, viewOf(400, hostTree()))
		runStop(t, capBroken)
		useProcessView(t, viewOf(401, capTree{401: {201, "moai"}, 201: {999, "bash"}, 999: {1, "claude"}}))
		runStop(t, capBroken)
		if got := capCount(t, root, "session", "sess-C"); got != 2 {
			t.Errorf("same session id under two hosts: count = %d, want 2", got)
		}
	})

	t.Run("(ii) fallback conditions", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			set  func(t *testing.T)
		}{
			{"absent", func(t *testing.T) { unsetEnvForTest(t, config.EnvClaudeCodeSessionID) }},
			{"empty", func(t *testing.T) { t.Setenv(config.EnvClaudeCodeSessionID, "") }},
			{"blank", func(t *testing.T) { t.Setenv(config.EnvClaudeCodeSessionID, "   ") }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				newCapRoot(t)
				unsetEnvForTest(t, config.EnvMoaiSessionPID)
				tc.set(t)
				useProcessView(t, viewOf(400, hostTree()))
				var last hookRunResult
				for k := 1; k <= capN+1; k++ {
					last = runStop(t, capBroken)
				}
				if last.stdout != capReleasedOutput {
					t.Errorf("9th stdout = %q, want %q", last.stdout, capReleasedOutput)
				}
				if l := linesContaining(last.stderr, capReleasedMarker); len(l) != 1 || !strings.Contains(l[0], "key: process") {
					t.Errorf("released line should name the process key kind: %q", l)
				}
			})
		}
	})

	t.Run("(iii) interposed shells are skipped", func(t *testing.T) {
		root := newCapRoot(t)
		t.Setenv(config.EnvClaudeCodeSessionID, "")
		unsetEnvForTest(t, config.EnvMoaiSessionPID)

		useProcessView(t, viewOf(400, hostTree()))
		runStop(t, capBroken)
		useProcessView(t, viewOf(401, capTree{401: {201, "moai"}, 201: {100, "bash"}, 100: {50, "claude"}, 50: {1, "zsh"}}))
		runStop(t, capBroken)

		if got := capCount(t, root, "process", "100"); got != 2 {
			t.Errorf("count for host 100 = %d, want 2 — both trees must resolve to the same host", got)
		}
		if n := len(capRecordFiles(t, root)); n != 1 {
			t.Errorf("%d records, want 1", n)
		}
	})

	t.Run("(iv) no key keeps blocking", func(t *testing.T) {
		root := newCapRoot(t)
		t.Setenv(config.EnvClaudeCodeSessionID, "")
		unsetEnvForTest(t, config.EnvMoaiSessionPID)
		useProcessView(t, session.ProcessView{
			Self:  400,
			Info:  func(int) (int, string, bool) { return 0, "", false },
			Alive: func(int) bool { return false },
		})
		for k := 1; k <= 12; k++ {
			r := runStop(t, capBroken)
			if r.stdout != claudeDenyBytes(t, hook.EventStop) {
				t.Errorf("call %d: stdout = %q, want the deny", k, r.stdout)
			}
			if len(linesContaining(r.stderr, capNotAppliedMarker)) != 1 {
				t.Errorf("call %d: want one %q line:\n%s", k, capNotAppliedMarker, r.stderr)
			}
		}
		if n := len(capRecordFiles(t, root)); n != 0 {
			t.Errorf("%d record files, want 0", n)
		}
	})

	t.Run("(v) file name confinement", func(t *testing.T) {
		for _, id := range []string{"../../escape", "a/b", `a\b`, "..", "x\x01y", "x\x7fy"} {
			root := newCapRoot(t)
			useSessionKey(t, id)
			r := runStop(t, capBroken)

			names := capRecordFiles(t, root)
			if len(names) != 1 || names[0] != capRecordName("session", id) {
				t.Errorf("id %q: records %v, want exactly [%s]", id, names, capRecordName("session", id))
			}
			allowed := map[string]bool{
				".moai": true, ".moai/state": true, capStateRel: true,
				capStateRel + "/" + capRecordName("session", id): true,
				".moai/logs": true, codexadapter.DiagnosticSinkRel: true,
			}
			_ = filepath.Walk(root, func(p string, _ os.FileInfo, err error) error {
				if err != nil || p == root {
					return nil
				}
				rel, _ := filepath.Rel(root, p)
				if !allowed[filepath.ToSlash(rel)] {
					t.Errorf("id %q: unexpected entry %s", id, rel)
				}
				return nil
			})
			rec, _ := os.ReadFile(capRecordPath(root, "session", id))
			if strings.Contains(string(rec), id) || strings.Contains(r.stderr, id) {
				t.Errorf("id %q: raw session id leaked into the record or stderr", id)
			}
		}

		nul := stopParseCapFileName("session", "a\x00b")
		if !capRecordNameRE.MatchString(nul) || nul != capRecordName("session", "a\x00b") {
			t.Errorf("NUL id: file name %q does not follow the rule", nul)
		}
		if stopParseCapFileName("session", "123") == stopParseCapFileName("process", "123") {
			t.Error(`session "123" and process 123 share a file name`)
		}
	})
}

// --- AC-SPC-008 -------------------------------------------------------------

// TestStopParseCap_AC008_UnusableRecord — REQ-SPC-008.
func TestStopParseCap_AC008_UnusableRecord(t *testing.T) {
	deny := func(t *testing.T) string { return claudeDenyBytes(t, hook.EventStop) }

	t.Run("(i) state area is a file", func(t *testing.T) {
		root := newCapRoot(t)
		useSessionKey(t, "ac008i")
		if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, capStateRel), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		for k := 1; k <= 10; k++ {
			r := runStop(t, capBroken)
			if r.stdout != deny(t) || len(linesContaining(r.stderr, capNotAppliedMarker)) != 1 {
				t.Errorf("call %d: stdout=%q, stderr=%q", k, r.stdout, r.stderr)
			}
		}
	})

	t.Run("(ii) unparseable own record", func(t *testing.T) {
		root := newCapRoot(t)
		useSessionKey(t, "ac008ii")
		p := capRecordPath(root, "session", "ac008ii")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("{"), 0o644); err != nil {
			t.Fatal(err)
		}
		for k := 1; k <= 9; k++ {
			r := runStop(t, capBroken)
			if k == 1 && capCount(t, root, "session", "ac008ii") != 1 {
				t.Errorf("record not replaced by a parseable count of 1")
			}
			want := deny(t)
			if k == 9 {
				want = capReleasedOutput
			}
			if r.stdout != want {
				t.Errorf("call %d: stdout = %q, want %q", k, r.stdout, want)
			}
		}
	})

	for _, shape := range []string{"symlink", "directory"} {
		t.Run("(iii) own record is a "+shape, func(t *testing.T) {
			root := newCapRoot(t)
			useSessionKey(t, "ac008iii")
			p := capRecordPath(root, "session", "ac008iii")
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(t.TempDir(), "target")
			if err := os.WriteFile(target, []byte("target bytes"), 0o644); err != nil {
				t.Fatal(err)
			}
			if shape == "symlink" {
				if err := os.Symlink(target, p); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			} else if err := os.Mkdir(p, 0o755); err != nil {
				t.Fatal(err)
			}
			for k := 1; k <= 10; k++ {
				r := runStop(t, capBroken)
				if r.stdout != deny(t) || len(linesContaining(r.stderr, capNotAppliedMarker)) != 1 {
					t.Errorf("call %d: stdout=%q, stderr=%q", k, r.stdout, r.stderr)
				}
			}
			if b, _ := os.ReadFile(target); string(b) != "target bytes" {
				t.Errorf("link target changed: %q", b)
			}
		})
	}
}

// --- AC-SPC-009 -------------------------------------------------------------

// TestStopParseCap_AC009_CodexUnchanged — REQ-SPC-009.
func TestStopParseCap_AC009_CodexUnchanged(t *testing.T) {
	root := newCapRoot(t)
	useSessionKey(t, "ac009")

	for k := 1; k <= 12; k++ {
		r := runHookHere(t, "stop", nil, "codex", []byte(capBroken))
		if r.stdout != capReleasedOutput {
			t.Errorf("(i) call %d: stdout = %q, want {}", k, r.stdout)
		}
		recs := sinkRecords(t, root)
		if len(recs) == 0 || recs[len(recs)-1].Key != stdinParseExemptDiscardKey {
			t.Errorf("(i) call %d: last record is not the exemption", k)
		}
	}
	for _, ev := range []hook.EventType{hook.EventPreToolUse, hook.EventPermissionRequest, hook.EventUserPromptSubmit} {
		r := runHookHere(t, subcommandFor(t, ev), nil, "codex", []byte(capBroken))
		if want := codexDenyBytes(t, ev); r.stdout != want {
			t.Errorf("(ii) %s: stdout = %q, want %q", ev, r.stdout, want)
		}
	}
	r := runHookHere(t, "agent", []string{"x-validation"}, "codex", []byte(capBroken))
	if want := codexDenyBytes(t, agentActionEvent("x-validation")); r.stdout != want {
		t.Errorf("(iii) agent x-validation: stdout = %q, want %q", r.stdout, want)
	}
	if n := len(capRecordFiles(t, root)); n != 0 {
		t.Errorf("codex calls created %d count records, want 0", n)
	}
}

// --- AC-SPC-010 -------------------------------------------------------------

// TestStopParseCap_AC010_ClaudeReason — REQ-SPC-010.
func TestStopParseCap_AC010_ClaudeReason(t *testing.T) {
	newCapRoot(t)
	useSessionKey(t, "ac010")
	type call struct {
		ev   hook.EventType
		sub  string
		args []string
	}
	calls := []call{
		{hook.EventPreToolUse, subcommandFor(t, hook.EventPreToolUse), nil},
		{hook.EventPermissionRequest, subcommandFor(t, hook.EventPermissionRequest), nil},
		{hook.EventStop, "stop", nil},
		{hook.EventUserPromptSubmit, subcommandFor(t, hook.EventUserPromptSubmit), nil},
		{agentActionEvent("x-validation"), "agent", []string{"x-validation"}},
	}
	for _, c := range calls {
		r := runHookHere(t, c.sub, c.args, "", []byte(capBroken))
		reason, ok := denyReason(c.ev, r.stdout)
		if !ok {
			t.Errorf("%s %v: no deny field in %q", c.sub, c.args, r.stdout)
			continue
		}
		if reason != claudeFailClosedReasonLiteral {
			t.Errorf("(a) %s %v: reason = %q, want %q", c.sub, c.args, reason, claudeFailClosedReasonLiteral)
		}
		if strings.Contains(reason, "disableAllHooks") || strings.Contains(reason, "moai update") {
			t.Errorf("(b) %s: reason carries recovery steps", c.sub)
		}
		if !strings.HasPrefix(reason, "fail-closed: ") || !strings.HasSuffix(reason, "(.moai/docs/hook-stdin-fail-closed.md)") {
			t.Errorf("(c) %s: reason frame broken: %q", c.sub, reason)
		}
	}
}

// --- AC-SPC-011 -------------------------------------------------------------

// TestStopParseCap_AC011_NoPayloadLeak — REQ-SPC-011.
func TestStopParseCap_AC011_NoPayloadLeak(t *testing.T) {
	const canary = "CANARY_7f3a"
	root := newCapRoot(t)
	useSessionKey(t, "ac011")
	payload := `{"tool_input":"` + canary + `"`
	for k := 1; k <= capN+2; k++ {
		r := runStop(t, payload)
		if k > capN && r.stdout != capReleasedOutput {
			t.Errorf("call %d: stdout = %q, want released", k, r.stdout)
		}
		if strings.Contains(r.stdout, canary) || strings.Contains(r.stderr, canary) {
			t.Errorf("call %d: canary leaked to stdout/stderr", k)
		}
	}
	for _, p := range []string{filepath.Join(root, codexadapter.DiagnosticSinkRel), capRecordPath(root, "session", "ac011")} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("%s unreadable: %v", p, err)
			continue
		}
		if strings.Contains(string(b), canary) {
			t.Errorf("canary leaked into %s", p)
		}
	}
}
