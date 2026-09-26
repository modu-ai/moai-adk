package cli

// SPEC-HOOK-STDIN-FAILCLOSED-001 — behaviour tests. A hook invocation whose
// stdin cannot be parsed must answer a decision-bearing event with a
// fail-closed deny (except the (Codex, Stop) exemption) and must leave every
// observation event on its existing default output.
//
// Event sets are computed from codexadapter.DecisionBearingEvents, the
// exemption predicate, and the dispatcher's own subcommand table — never
// hand-listed here.

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// readCountingProtocol wraps the real protocol and counts ReadInput calls, the
// direct observation that the harness decision happened before stdin was read.
type readCountingProtocol struct {
	hook.Protocol
	reads int
}

func (p *readCountingProtocol) ReadInput(r io.Reader) (*hook.HookInput, error) {
	p.reads++
	return p.Protocol.ReadInput(r)
}

// brokenStdin is one of the four parse-failure forms of acceptance.md §0.
type brokenStdin struct {
	name    string
	canary  string
	payload []byte
}

func newStdinCanary(t *testing.T, form string) string {
	t.Helper()
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("canary: %v", err)
	}
	return "CANARY-" + strings.ToUpper(form) + "-" + hex.EncodeToString(b)
}

// depthPayload nests an array depth levels deep inside tool_input.
func depthPayload(ev hook.EventType, canary string, depth int) []byte {
	var b bytes.Buffer
	b.WriteString(`{"hook_event_name":"` + string(ev) + `","tool_input":{"note":"` + canary + `","x":`)
	b.WriteString(strings.Repeat("[", depth))
	b.WriteString(strings.Repeat("]", depth))
	b.WriteString(`}}`)
	return b.Bytes()
}

// brokenStdinForms builds the four forms for ev, each with its own canary.
func brokenStdinForms(t *testing.T, ev hook.EventType) []brokenStdin {
	t.Helper()
	malformed := newStdinCanary(t, "malformed")
	truncated := newStdinCanary(t, "truncated")
	oversize := newStdinCanary(t, "oversize")
	depth := newStdinCanary(t, "depth")

	var over bytes.Buffer
	over.WriteString(`{"hook_event_name":"` + string(ev) + `","tool_input":{"note":"` + oversize + `","content":"`)
	over.Write(bytes.Repeat([]byte("a"), 6<<20))
	over.WriteString(`"}}`)

	return []brokenStdin{
		{"malformed", malformed, []byte(`{"broken":"` + malformed)},
		{"truncated", truncated, []byte(`{"hook_event_name":"` + string(ev) + `","tool_name":"Bash","tool_input":{"note":"` + truncated + `",`)},
		{"oversize", oversize, over.Bytes()},
		{"depth", depth, depthPayload(ev, depth, 10001)},
	}
}

// swapStdinFile points os.Stdin at a file holding data. A file rather than a
// pipe: a pipe would block on the 6 MiB oversize form before anything reads it.
func swapStdinFile(t *testing.T, data []byte) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "stdin")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write stdin fixture: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open stdin fixture: %v", err)
	}
	orig := os.Stdin
	os.Stdin = f
	t.Cleanup(func() {
		os.Stdin = orig
		_ = f.Close()
	})
}

// captureStdoutStderr captures both process streams while fn runs.
func captureStdoutStderr(t *testing.T, fn func()) (string, string) {
	t.Helper()
	pipe := func() (*os.File, *os.File, chan string) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("pipe: %v", err)
		}
		done := make(chan string, 1)
		go func() {
			data, _ := io.ReadAll(r)
			done <- string(data)
		}()
		return r, w, done
	}
	_, wOut, outDone := pipe()
	_, wErr, errDone := pipe()
	origOut, origErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr
	func() {
		defer func() { os.Stdout, os.Stderr = origOut, origErr }()
		fn()
	}()
	_ = wOut.Close()
	_ = wErr.Close()
	return <-outDone, <-errDone
}

type hookRunResult struct {
	stdout     string
	stderr     string
	err        error
	dispatched int
	reads      int
	root       string
}

func findHookSubcommand(t *testing.T, name string) *cobra.Command {
	t.Helper()
	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	t.Fatalf("hook subcommand %q not found", name)
	return nil
}

// runHookWithStdin runs one hook subcommand (or `agent <action>`) with the
// real protocol, a spy registry, the given --harness value, and stdin, in an
// isolated project root.
func runHookWithStdin(t *testing.T, sub string, args []string, harness string, stdin []byte) hookRunResult {
	t.Helper()
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	spy := &spyRegistry{}
	proto := &readCountingProtocol{Protocol: hook.NewProtocol()}
	origDeps := deps
	deps = &Dependencies{HookRegistry: spy, HookProtocol: proto}
	t.Cleanup(func() { deps = origDeps })

	cmd := findHookSubcommand(t, sub)
	if err := cmd.ParseFlags([]string{"--harness=" + harness}); err != nil {
		t.Fatalf("parse --harness=%s: %v", harness, err)
	}
	t.Cleanup(func() { _ = cmd.Flags().Set("harness", "") })
	cmd.SetContext(context.Background())
	swapStdinFile(t, stdin)

	var runErr error
	stdout, stderr := captureStdoutStderr(t, func() { runErr = cmd.RunE(cmd, args) })
	return hookRunResult{stdout: stdout, stderr: stderr, err: runErr, dispatched: len(spy.dispatched), reads: proto.reads, root: root}
}

// subcommandFor returns the dispatcher subcommand that handles ev.
func subcommandFor(t *testing.T, ev hook.EventType) string {
	t.Helper()
	for _, s := range hookEventSubcommands {
		if s.event == ev {
			return s.use
		}
	}
	t.Fatalf("no dispatcher subcommand handles %s", ev)
	return ""
}

// expectedFailClosedReason assembles the reason from the two constants and
// the fixed frame of acceptance.md AC-HSF-001(e4). It deliberately does not
// call the implementation's own reason builder.
func expectedFailClosedReason() string {
	return "fail-closed: " + stdinParseFailureCause + " (" + stdinParseFailClosedDocID + ")"
}

func expectedClaudeFailClosed(t *testing.T, ev hook.EventType) string {
	t.Helper()
	row, ok := codexadapter.Lookup(codexadapter.HarnessClaude, ev, codexadapter.DecisionFatalError)
	if !ok {
		t.Fatalf("no Claude fatal_error row for %s", ev)
	}
	out, err := codexadapter.Render(ev, row.Outcome, expectedFailClosedReason())
	if err != nil {
		t.Fatalf("render %s: %v", ev, err)
	}
	return string(out)
}

func expectedCodexFailClosed(t *testing.T, ev hook.EventType) string {
	t.Helper()
	out, _, err := codexadapter.TranslateCodex(ev, codexadapter.DecisionFatalError, expectedFailClosedReason())
	if err != nil {
		t.Fatalf("translate %s: %v", ev, err)
	}
	return string(out)
}

// denyReason returns the deny field's reason when stdout carries the deny
// field of acceptance.md §0 for ev, and ok=false otherwise.
func denyReason(ev hook.EventType, stdout string) (string, bool) {
	var v map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &v); err != nil {
		return "", false
	}
	switch ev {
	case hook.EventPreToolUse:
		hso, _ := v["hookSpecificOutput"].(map[string]any)
		if hso["permissionDecision"] != "deny" {
			return "", false
		}
		r, _ := hso["permissionDecisionReason"].(string)
		return r, true
	case hook.EventPermissionRequest:
		hso, _ := v["hookSpecificOutput"].(map[string]any)
		dec, _ := hso["decision"].(map[string]any)
		if dec["behavior"] != "deny" {
			return "", false
		}
		r, _ := dec["message"].(string)
		return r, true
	default:
		if v["decision"] != "block" {
			return "", false
		}
		r, _ := v["reason"].(string)
		return r, true
	}
}

// hasAnyDenyField reports whether stdout carries any of the §0 deny fields.
func hasAnyDenyField(stdout string) bool {
	for _, ev := range codexadapter.DecisionBearingEvents() {
		if _, ok := denyReason(ev, stdout); ok {
			return true
		}
	}
	return false
}

func assertSingleJSONValue(t *testing.T, stdout string) {
	t.Helper()
	trimmed := strings.TrimSpace(stdout)
	if trimmed == "" || trimmed == "{}" {
		t.Fatalf("stdout = %q, want a fail-closed deny (neither empty nor {})", trimmed)
	}
	dec := json.NewDecoder(strings.NewReader(trimmed))
	var v any
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("stdout is not JSON: %q", trimmed)
	}
	if dec.More() {
		t.Fatalf("stdout carries more than one JSON value: %q", trimmed)
	}
}

// readSink returns the adapter sink records under root.
func readSink(t *testing.T, root string) []codexadapter.Discard {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, codexadapter.DiagnosticSinkRel))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("read sink: %v", err)
	}
	var out []codexadapter.Discard
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		var d codexadapter.Discard
		if err := json.Unmarshal(sc.Bytes(), &d); err != nil {
			t.Fatalf("sink line is not a Discard: %q", sc.Text())
		}
		out = append(out, d)
	}
	return out
}

// expectedStdinBytes is what the dispatcher observes: the input length, capped
// at the protocol's 5 MiB read limit.
func expectedStdinBytes(payload []byte) int {
	if len(payload) > 5<<20 {
		return 5 << 20
	}
	return len(payload)
}

// assertFailClosedObservability checks AC-HSF-007 for one fail-closed run: the
// stderr line, exactly one parse-failure record carrying the stdin byte count,
// and no canary on any surface.
func assertFailClosedObservability(t *testing.T, ev hook.EventType, harness string, form brokenStdin, r hookRunResult) {
	t.Helper()
	mode := "claude"
	if harness == "codex" {
		mode = "codex"
	}
	var line string
	for _, l := range strings.Split(r.stderr, "\n") {
		if strings.Contains(l, "fail-closed") && strings.Contains(l, string(ev)) {
			line = l
			break
		}
	}
	if line == "" {
		t.Fatalf("no fail-closed stderr line naming %s:\n%s", ev, r.stderr)
	}
	if !strings.Contains(line, "harness "+mode) {
		t.Errorf("stderr line lacks the harness mode %q: %q", mode, line)
	}
	if !strings.Contains(line, "invalid JSON input") {
		t.Errorf("stderr line lacks the parse error cause: %q", line)
	}

	recs := readSink(t, r.root)
	if len(recs) != 1 {
		t.Fatalf("sink records = %d, want exactly 1: %+v", len(recs), recs)
	}
	rec := recs[0]
	if rec.Key != stdinParseFailClosedDiscardKey || rec.Key == hookFaultDiscardKey {
		t.Errorf("record key = %q, want the parse-failure key %q", rec.Key, stdinParseFailClosedDiscardKey)
	}
	if rec.Event != ev {
		t.Errorf("record event = %s, want %s", rec.Event, ev)
	}
	if want := expectedStdinBytes(form.payload); rec.ContentLength != want {
		t.Errorf("record content_length = %d, want the observed stdin bytes %d", rec.ContentLength, want)
	}

	sink, _ := os.ReadFile(filepath.Join(r.root, codexadapter.DiagnosticSinkRel))
	for name, surface := range map[string]string{"stdout": r.stdout, "stderr": r.stderr, "sink": string(sink)} {
		if strings.Contains(surface, form.canary) {
			t.Errorf("%s carries the %s canary — payload leaked", name, form.name)
		}
	}
}

// assertReasonShape checks AC-HSF-001(e1)-(e5) on a Claude reason field.
func assertReasonShape(t *testing.T, reason string, forms []brokenStdin) {
	t.Helper()
	if strings.TrimSpace(reason) == "" {
		t.Fatal("reason is empty")
	}
	if !strings.Contains(reason, "fail-closed") {
		t.Errorf("(e1) reason lacks fail-closed: %q", reason)
	}
	if !strings.Contains(reason, stdinParseFailureCause) {
		t.Errorf("(e2) reason lacks the cause constant: %q", reason)
	}
	if !strings.Contains(reason, stdinParseFailClosedDocID) {
		t.Errorf("(e3) reason lacks the document identifier: %q", reason)
	}
	if reason != expectedFailClosedReason() {
		t.Errorf("(e4) reason = %q, want exactly %q", reason, expectedFailClosedReason())
	}
	for _, f := range forms {
		if strings.Contains(reason, f.canary) {
			t.Errorf("(e5) reason carries the %s canary", f.name)
		}
	}
	if strings.Contains(reason, "disableAllHooks") {
		t.Errorf("(e5) reason carries recovery guidance: %q", reason)
	}
}

// failClosedPairs returns the decision events answered fail-closed under
// harness h, computed from the event set and the exemption predicate.
func failClosedEvents(h codexadapter.Harness) []hook.EventType {
	var out []hook.EventType
	for _, ev := range codexadapter.DecisionBearingEvents() {
		if !codexadapter.HostLacksStopBlockCap(h, ev) {
			out = append(out, ev)
		}
	}
	return out
}

// observationSubcommands returns the dispatcher subcommands whose event is not
// decision-bearing.
func observationSubcommands() []struct {
	use   string
	event hook.EventType
} {
	var out []struct {
		use   string
		event hook.EventType
	}
	for _, s := range hookEventSubcommands {
		if !codexadapter.IsDecisionBearing(s.event) {
			out = append(out, struct {
				use   string
				event hook.EventType
			}{s.use, s.event})
		}
	}
	return out
}

// TestStdinFailClosed_ClaudeDecisionEvents is AC-HSF-001 (+ AC-HSF-007 for the
// Claude pairs).
func TestStdinFailClosed_ClaudeDecisionEvents(t *testing.T) {
	for _, ev := range failClosedEvents(codexadapter.HarnessClaude) {
		forms := brokenStdinForms(t, ev)
		for _, f := range forms {
			t.Run(string(ev)+"/"+f.name, func(t *testing.T) {
				r := runHookWithStdin(t, subcommandFor(t, ev), nil, "", f.payload)
				if r.err != nil {
					t.Fatalf("(a) RunE = %v, want nil (exit 0)", r.err)
				}
				if r.dispatched != 0 {
					t.Fatalf("(b) dispatched %d times, want 0", r.dispatched)
				}
				assertSingleJSONValue(t, r.stdout)
				if got, want := strings.TrimSpace(r.stdout), expectedClaudeFailClosed(t, ev); got != want {
					t.Errorf("(d) stdout = %s\nwant       %s", got, want)
				}
				reason, ok := denyReason(ev, r.stdout)
				if !ok {
					t.Fatalf("(d) stdout lacks the %s deny field: %s", ev, r.stdout)
				}
				assertReasonShape(t, reason, forms)
				assertFailClosedObservability(t, ev, "", f, r)
			})
		}
	}
}

// TestStdinFailClosed_CodexDecisionEvents is AC-HSF-002 (+ AC-HSF-007 for the
// Codex pairs).
func TestStdinFailClosed_CodexDecisionEvents(t *testing.T) {
	events := failClosedEvents(codexadapter.HarnessCodex)
	if len(events) == 0 {
		t.Fatal("no Codex fail-closed events — the predicate degenerated")
	}
	for _, ev := range events {
		forms := brokenStdinForms(t, ev)
		for _, f := range forms {
			t.Run(string(ev)+"/"+f.name, func(t *testing.T) {
				r := runHookWithStdin(t, subcommandFor(t, ev), nil, "codex", f.payload)
				if r.err != nil {
					t.Fatalf("(a) RunE = %v, want nil (exit 0)", r.err)
				}
				if r.dispatched != 0 {
					t.Fatalf("(b) dispatched %d times, want 0", r.dispatched)
				}
				assertSingleJSONValue(t, r.stdout)
				if got, want := strings.TrimSpace(r.stdout), expectedCodexFailClosed(t, ev); got != want {
					t.Fatalf("(d) stdout = %s\nwant       %s", got, want)
				}
				if _, ok := denyReason(ev, r.stdout); !ok {
					t.Fatalf("(d) stdout lacks the %s deny field: %s", ev, r.stdout)
				}
				assertFailClosedObservability(t, ev, "codex", f, r)
			})
		}
	}
}

// TestStdinFailClosed_ObservedDecisionSetMatchesTheSource is AC-HSF-003(a):
// the events whose malformed-stdin output carries a deny field are exactly the
// source set (Claude) and the source set minus the exemption (Codex).
func TestStdinFailClosed_ObservedDecisionSetMatchesTheSource(t *testing.T) {
	for _, tc := range []struct {
		flag    string
		harness codexadapter.Harness
	}{{"", codexadapter.HarnessClaude}, {"codex", codexadapter.HarnessCodex}} {
		observed := map[hook.EventType]bool{}
		for _, s := range hookEventSubcommands {
			r := runHookWithStdin(t, s.use, nil, tc.flag, []byte(`{"broken`))
			if r.err != nil {
				t.Fatalf("%s/%s: RunE = %v", tc.harness, s.use, r.err)
			}
			if hasAnyDenyField(r.stdout) {
				observed[s.event] = true
			}
		}
		want := map[hook.EventType]bool{}
		for _, ev := range failClosedEvents(tc.harness) {
			want[ev] = true
		}
		if len(observed) != len(want) {
			t.Fatalf("%s: observed fail-closed set %v, want %v", tc.harness, observed, want)
		}
		for ev := range want {
			if !observed[ev] {
				t.Fatalf("%s: observed fail-closed set %v lacks %s (want %v)", tc.harness, observed, ev, want)
			}
		}
	}
}

// TestStdinFailClosed_HarnessDecidedBeforeStdin is AC-HSF-004.
func TestStdinFailClosed_HarnessDecidedBeforeStdin(t *testing.T) {
	r := runHookWithStdin(t, "pre-tool", nil, "bogus", []byte(`{"broken`))
	if r.err == nil || !strings.Contains(r.err.Error(), "invalid --harness value") {
		t.Fatalf("RunE = %v, want an invalid --harness value error", r.err)
	}
	if strings.Contains(r.stderr, "invalid stdin JSON") || strings.Contains(r.stderr, "fail-closed") {
		t.Fatalf("stdin was parsed before the harness was decided:\n%s", r.stderr)
	}
	if r.reads != 0 {
		t.Fatalf("ReadInput called %d times, want 0 — the harness decision must precede reading stdin", r.reads)
	}

	r = runHookWithStdin(t, "pre-tool", nil, "codex", []byte(`{"broken`))
	if r.err != nil {
		t.Fatalf("codex RunE = %v", r.err)
	}
	if got, want := strings.TrimSpace(r.stdout), expectedCodexFailClosed(t, hook.EventPreToolUse); got != want {
		t.Fatalf("codex stdout = %s\nwant %s", got, want)
	}
}

// TestStdinFailClosed_ObservationEventsPreserved is AC-HSF-005: 22 observation
// subcommands x 4 forms x 2 harness modes keep the default output.
func TestStdinFailClosed_ObservationEventsPreserved(t *testing.T) {
	subs := observationSubcommands()
	if len(subs) != 22 {
		t.Fatalf("observation subcommands = %d, want 22 (table changed?)", len(subs))
	}
	for _, s := range subs {
		forms := brokenStdinForms(t, s.event)
		for _, flag := range []string{"", "codex"} {
			for _, f := range forms {
				t.Run(s.use+"/"+flag+"/"+f.name, func(t *testing.T) {
					r := runHookWithStdin(t, s.use, nil, flag, f.payload)
					if r.err != nil {
						t.Fatalf("RunE = %v, want nil", r.err)
					}
					if r.dispatched != 0 {
						t.Fatalf("dispatched %d times, want 0", r.dispatched)
					}
					lines := strings.Split(strings.TrimRight(r.stderr, "\n"), "\n")
					prefix := "moai hook " + string(s.event) + ": invalid stdin JSON ("
					if len(lines) != 1 || !strings.HasPrefix(lines[0], prefix) || !strings.HasSuffix(lines[0], "); emitting default output") {
						t.Fatalf("stderr = %q, want exactly one %q warning line", r.stderr, prefix)
					}
					want := "{}\n"
					if s.event == hook.EventWorktreeCreate || s.event == hook.EventWorktreeRemove {
						want = ""
					}
					if r.stdout != want {
						t.Fatalf("stdout = %q, want %q", r.stdout, want)
					}
					if recs := readSink(t, r.root); len(recs) != 0 {
						t.Fatalf("an observation event wrote sink records: %+v", recs)
					}
				})
			}
		}
	}
}

// TestStdinFailClosed_StopIgnoresStopHookActive is AC-HSF-008.
func TestStdinFailClosed_StopIgnoresStopHookActive(t *testing.T) {
	for _, in := range []string{
		`{"hook_event_name":"Stop","stop_hook_active":true,`,
		`{"hook_event_name":"Stop","stop_hook_active":false,`,
	} {
		r := runHookWithStdin(t, "stop", nil, "", []byte(in))
		if r.err != nil {
			t.Fatalf("claude RunE = %v", r.err)
		}
		if _, ok := denyReason(hook.EventStop, r.stdout); !ok {
			t.Fatalf("claude Stop stdout = %q, want decision:block regardless of stop_hook_active", r.stdout)
		}

		r = runHookWithStdin(t, "stop", nil, "codex", []byte(in))
		if r.err != nil {
			t.Fatalf("codex RunE = %v", r.err)
		}
		if strings.TrimSpace(r.stdout) != "{}" {
			t.Fatalf("codex Stop stdout = %q, want {} (exemption)", r.stdout)
		}
		recs := readSink(t, r.root)
		if len(recs) != 1 || recs[0].Key != stdinParseExemptDiscardKey {
			t.Fatalf("codex Stop sink = %+v, want one exemption record", recs)
		}
	}
}

// TestStdinFailClosed_CodexStopExempt is AC-HSF-011.
func TestStdinFailClosed_CodexStopExempt(t *testing.T) {
	forms := brokenStdinForms(t, hook.EventStop)
	for _, f := range forms {
		t.Run(f.name, func(t *testing.T) {
			r := runHookWithStdin(t, "stop", nil, "codex", f.payload)
			if r.err != nil {
				t.Fatalf("(a) RunE = %v, want nil", r.err)
			}
			if r.dispatched != 0 {
				t.Fatalf("(a) dispatched %d times, want 0", r.dispatched)
			}
			if strings.TrimSpace(r.stdout) != "{}" {
				t.Fatalf("(b) stdout = %q, want exactly {}", r.stdout)
			}
			var line string
			for _, l := range strings.Split(r.stderr, "\n") {
				if strings.Contains(l, "exempt") {
					line = l
					break
				}
			}
			if line == "" {
				t.Fatalf("(c) no exemption stderr line:\n%s", r.stderr)
			}
			for _, want := range []string{string(hook.EventStop), "harness codex", "invalid JSON input", "no Stop block cap"} {
				if !strings.Contains(line, want) {
					t.Errorf("(c) exemption line lacks %q: %q", want, line)
				}
			}
			recs := readSink(t, r.root)
			if len(recs) != 1 {
				t.Fatalf("(c) sink records = %d, want 1: %+v", len(recs), recs)
			}
			if k := recs[0].Key; k != stdinParseExemptDiscardKey || k == stdinParseFailClosedDiscardKey || k == hookFaultDiscardKey {
				t.Errorf("(c) record key = %q, want the exemption key %q", k, stdinParseExemptDiscardKey)
			}
			if want := expectedStdinBytes(f.payload); recs[0].ContentLength != want {
				t.Errorf("(c) record content_length = %d, want %d", recs[0].ContentLength, want)
			}
			sink, _ := os.ReadFile(filepath.Join(r.root, codexadapter.DiagnosticSinkRel))
			for name, surface := range map[string]string{"stdout": r.stdout, "stderr": r.stderr, "sink": string(sink)} {
				if strings.Contains(surface, f.canary) {
					t.Errorf("(d) %s carries the canary", name)
				}
			}

			// The same input without --harness is blocked (AC-HSF-001).
			rc := runHookWithStdin(t, "stop", nil, "", f.payload)
			if _, ok := denyReason(hook.EventStop, rc.stdout); !ok {
				t.Fatalf("claude Stop stdout = %q, want decision:block", rc.stdout)
			}
		})
	}
}

// agentDecisionActions / agentObservationActions are the acceptance.md action
// fixtures; each action's event comes from the dispatcher's own mapping.
var (
	agentDecisionActions    = []string{"x-validation", "x-pre-transformation", "x-pre-implementation", "foo"}
	agentObservationActions = []string{"x-verification", "x-post-transformation", "x-post-implementation", "x-completion"}
)

// TestStdinFailClosed_AgentDecisionActions is AC-HSF-012 (+ AC-HSF-007 for the
// agent path).
func TestStdinFailClosed_AgentDecisionActions(t *testing.T) {
	for _, action := range agentDecisionActions {
		ev := agentActionEvent(action)
		if !codexadapter.IsDecisionBearing(ev) {
			t.Fatalf("fixture %s maps to %s, which is not decision-bearing", action, ev)
		}
		forms := brokenStdinForms(t, ev)
		for _, flag := range []string{"", "codex"} {
			for _, f := range forms {
				t.Run(action+"/"+flag+"/"+f.name, func(t *testing.T) {
					r := runHookWithStdin(t, "agent", []string{action}, flag, f.payload)
					if r.err != nil {
						t.Fatalf("(a) RunE = %v, want nil", r.err)
					}
					if r.dispatched != 0 {
						t.Fatalf("(b) dispatched %d times, want 0", r.dispatched)
					}
					want := expectedClaudeFailClosed(t, ev)
					if flag == "codex" {
						want = expectedCodexFailClosed(t, ev)
					}
					if got := strings.TrimSpace(r.stdout); got != want {
						t.Fatalf("(c) stdout = %s\nwant       %s", got, want)
					}
					if flag == "" {
						reason, ok := denyReason(ev, r.stdout)
						if !ok {
							t.Fatalf("(c) stdout lacks the deny field: %s", r.stdout)
						}
						assertReasonShape(t, reason, forms)
					}
					assertFailClosedObservability(t, ev, flag, f, r)
				})
			}
		}
	}

	r := runHookWithStdin(t, "agent", []string{"x-validation"}, "bogus", []byte(`{"broken`))
	if r.err == nil || !strings.Contains(r.err.Error(), "invalid --harness value") {
		t.Fatalf("bogus+malformed RunE = %v, want an invalid --harness value error", r.err)
	}
	if strings.Contains(r.stderr, "invalid stdin JSON") || strings.Contains(r.stderr, "fail-closed") {
		t.Fatalf("stdin was parsed before the harness was decided:\n%s", r.stderr)
	}
	if r.reads != 0 {
		t.Fatalf("ReadInput called %d times, want 0", r.reads)
	}
}

// TestStdinFailClosed_AgentRejectsBogusHarness closes N15: every agent action,
// decision- or observation-mapped, refuses an invalid --harness before stdin is
// read, even with a valid payload. This is a new contract (REQ-HSF-005): the
// pre-fix tree accepted all eight with exit 0.
func TestStdinFailClosed_AgentRejectsBogusHarness(t *testing.T) {
	for _, action := range append(append([]string{}, agentDecisionActions...), agentObservationActions...) {
		t.Run(action, func(t *testing.T) {
			r := runHookWithStdin(t, "agent", []string{action}, "bogus", []byte(`{}`))
			if r.err == nil || !strings.Contains(r.err.Error(), "invalid --harness value") {
				t.Fatalf("RunE = %v, want an invalid --harness value error", r.err)
			}
			if r.reads != 0 {
				t.Fatalf("ReadInput called %d times, want 0", r.reads)
			}
			if r.dispatched != 0 {
				t.Fatalf("dispatched %d times, want 0", r.dispatched)
			}
		})
	}
}

// TestStdinFailClosed_AgentObservationActionsPreserved is AC-HSF-013.
func TestStdinFailClosed_AgentObservationActionsPreserved(t *testing.T) {
	for _, action := range agentObservationActions {
		ev := agentActionEvent(action)
		if codexadapter.IsDecisionBearing(ev) {
			t.Fatalf("fixture %s maps to decision-bearing %s", action, ev)
		}
		forms := brokenStdinForms(t, ev)
		for _, flag := range []string{"", "codex"} {
			for _, f := range forms {
				t.Run(action+"/"+flag+"/"+f.name, func(t *testing.T) {
					r := runHookWithStdin(t, "agent", []string{action}, flag, f.payload)
					if r.err != nil {
						t.Fatalf("RunE = %v, want nil", r.err)
					}
					if r.dispatched != 0 {
						t.Fatalf("dispatched %d times, want 0", r.dispatched)
					}
					lines := strings.Split(strings.TrimRight(r.stderr, "\n"), "\n")
					prefix := "moai hook agent " + action + ": invalid stdin JSON ("
					if len(lines) != 1 || !strings.HasPrefix(lines[0], prefix) || !strings.HasSuffix(lines[0], "); emitting default output") {
						t.Fatalf("stderr = %q, want exactly one %q warning line", r.stderr, prefix)
					}
					if r.stdout != "{}\n" {
						t.Fatalf("stdout = %q, want %q", r.stdout, "{}\n")
					}
					if recs := readSink(t, r.root); len(recs) != 0 {
						t.Fatalf("an observation action wrote sink records: %+v", recs)
					}
				})
			}
		}
	}
	for _, action := range []string{"x-verification", "x-completion"} {
		r := runHookWithStdin(t, "agent", []string{action}, "bogus", []byte(`{}`))
		if r.err == nil || !strings.Contains(r.err.Error(), "invalid --harness value") {
			t.Fatalf("%s bogus RunE = %v, want an invalid --harness value error", action, r.err)
		}
		if r.reads != 0 {
			t.Fatalf("%s: ReadInput called %d times, want 0", action, r.reads)
		}
	}
}

// TestStdinFailClosed_DepthControl is the acceptance.md §0 control: depth 9000
// parses, depth 10001 does not, at the ReadInput level.
func TestStdinFailClosed_DepthControl(t *testing.T) {
	p := hook.NewProtocol()
	if _, err := p.ReadInput(bytes.NewReader(depthPayload(hook.EventPreToolUse, "ctl", 9000))); err != nil {
		t.Fatalf("depth 9000 control failed to parse: %v — the depth form cannot be attributed to the depth limit", err)
	}
	if _, err := p.ReadInput(bytes.NewReader(depthPayload(hook.EventPreToolUse, "ctl", 10001))); err == nil {
		t.Fatal("depth 10001 parsed; the depth form does not exercise a parse failure")
	}
}
