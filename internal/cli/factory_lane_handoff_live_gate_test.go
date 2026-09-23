package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The LIVE gates of AC-FLH-012/013 are jq programs written in acceptance.md.
// This test reads those exact programs from the SPEC file and evaluates them with
// the jq binary, so it exercises the production predicates rather than a copy
// that could drift looser than the acceptance text.
const liveGateAcceptancePath = "../../.moai/specs/SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001/acceptance.md"

type liveGatePredicates struct {
	events   string // jq -se program over the go test -json stream
	evidence string // jq -e program over the card-scoped evidence JSON
}

var (
	liveGateEventsRe   = regexp.MustCompile(`jq -se '([^']+)'`)
	liveGateEvidenceRe = regexp.MustCompile(`jq -e '([^']+)'`)
)

// loadLiveGatePredicates returns the predicates of the first bash block that
// follows the given AC heading.
func loadLiveGatePredicates(t *testing.T, doc, heading string) liveGatePredicates {
	t.Helper()
	start := strings.Index(doc, heading)
	if start < 0 {
		t.Fatalf("acceptance.md has no heading %q", heading)
	}
	rest := doc[start:]
	open := strings.Index(rest, "```bash")
	if open < 0 {
		t.Fatalf("%s has no bash block", heading)
	}
	block := rest[open+len("```bash"):]
	end := strings.Index(block, "```")
	if end < 0 {
		t.Fatalf("%s bash block is unterminated", heading)
	}
	block = block[:end]
	ev := liveGateEventsRe.FindStringSubmatch(block)
	evi := liveGateEvidenceRe.FindStringSubmatch(block)
	if ev == nil || evi == nil {
		t.Fatalf("%s bash block lacks a jq -se or jq -e predicate", heading)
	}
	return liveGatePredicates{events: ev[1], evidence: evi[1]}
}

// runLiveGate evaluates a jq program; it passes only when jq exits 0 and prints true.
func runLiveGate(t *testing.T, flag, program string, input []byte) bool {
	t.Helper()
	jq, err := exec.LookPath("jq")
	if err != nil {
		t.Fatalf("jq is required to evaluate the acceptance predicates: %v", err)
	}
	path := filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(path, input, 0o600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	cmd := exec.Command(jq, flag, program, path)
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	return err == nil && strings.TrimSpace(out.String()) == "true"
}

func liveGateContext(argv0 string, pid int) map[string]any {
	return map[string]any{
		"production_cli": true,
		"real_model":     true,
		"argv":           []any{argv0, "exec"},
		"pid":            float64(pid),
		"process_start":  "Wed Sep 24 10:00:00 2026",
		"binary_sha256":  strings.Repeat("ab", 32),
	}
}

func liveGateCommon(mode string) map[string]any {
	const head = "0123456789abcdef0123456789abcdef01234567"
	return map[string]any{
		"schema_version": float64(1),
		"card_id":        "t1082",
		"spec_id":        "SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001",
		"mode":           mode,
		"contexts": map[string]any{
			"separate": true,
			"lead":     liveGateContext("/bin/lead", 4101),
			"lane":     liveGateContext("/bin/lane", 4202),
		},
		"target": map[string]any{
			"reserved_cwd":    "/repo/.claude/worktrees/t1082",
			"reserved_branch": "WT-lane-handoff",
			"pinned_head":     head,
		},
		"empty_model_turn_count": float64(0),
		"receipts":               map[string]any{"bound_id": "bound-1", "message_id": "msg-1"},
		"old_endpoint":           map[string]any{"rejected": true, "code": "STALE"},
		"writes":                 map[string]any{"pre_bound": float64(0), "wrong_cwd": float64(0), "primary_branch_switches": float64(0)},
		"messages":               map[string]any{"sent": float64(2), "received": float64(2), "lost": float64(0)},
		"nonces":                 map[string]any{"lead_to_lane": "nonce-a", "lane_to_lead": "nonce-b"},
		"cleanup":                true,
		"bypass": map[string]any{
			"direct_register_peer": false, "db_seed": false, "mock": false,
			"fixture": false, "private_control": false,
		},
	}
}

func liveGateInteractiveDoc() map[string]any {
	doc := liveGateCommon("interactive")
	target := doc["target"].(map[string]any)
	target["observed_cwd"] = target["reserved_cwd"]
	target["observed_branch"] = target["reserved_branch"]
	target["observed_head"] = target["pinned_head"]
	doc["interactive"] = map[string]any{
		"actual_cd":                      true,
		"state_before_session_start":     "SWITCH_PENDING_INTERACTIVE",
		"next_normal_turn_session_start": true,
	}
	return doc
}

func liveGateHeadlessDoc() map[string]any {
	doc := liveGateCommon("headless")
	target := doc["target"].(map[string]any)
	doc["app_server"] = map[string]any{
		"method":             "thread/fork",
		"request_cwd":        target["reserved_cwd"],
		"returned_thread_id": "thread-new",
		"forked_from_id":     "thread-source",
		"thread_started":     true,
	}
	doc["controller_readback"] = map[string]any{
		"cwd":    target["reserved_cwd"],
		"branch": target["reserved_branch"],
		"head":   target["pinned_head"],
	}
	doc["session_start_wait_count"] = float64(0)
	return doc
}

// liveGateClone deep-copies a document through JSON so mutants never alias.
func liveGateClone(t *testing.T, doc map[string]any) map[string]any {
	t.Helper()
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

// liveGateLeafPaths lists every non-object path; arrays count as one leaf.
func liveGateLeafPaths(doc map[string]any, prefix []string) [][]string {
	var out [][]string
	keys := make([]string, 0, len(doc))
	for k := range doc {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		p := append(append([]string{}, prefix...), k)
		if m, ok := doc[k].(map[string]any); ok {
			out = append(out, liveGateLeafPaths(m, p)...)
			continue
		}
		out = append(out, p)
	}
	return out
}

func liveGateSet(doc map[string]any, path []string, value any, remove bool) {
	m := doc
	for _, k := range path[:len(path)-1] {
		m = m[k].(map[string]any)
	}
	if remove {
		delete(m, path[len(path)-1])
		return
	}
	m[path[len(path)-1]] = value
}

func liveGateJSON(t *testing.T, doc map[string]any) []byte {
	t.Helper()
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func liveGateEvents(t *testing.T, liveTest string, extra ...map[string]any) []byte {
	t.Helper()
	gate := "TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants"
	events := []map[string]any{
		{"Action": "run", "Test": liveTest},
		{"Action": "pass", "Test": liveTest},
		{"Action": "run", "Test": gate},
		{"Action": "pass", "Test": gate},
	}
	events = append(events, extra...)
	events = append(events, map[string]any{"Action": "pass", "Package": "github.com/modu-ai/moai-adk/internal/cli"})
	var buf bytes.Buffer
	for _, e := range events {
		b, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		buf.Write(b)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

func TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants(t *testing.T) {
	raw, err := os.ReadFile(liveGateAcceptancePath)
	if err != nil {
		t.Fatalf("read acceptance predicates: %v", err)
	}
	doc := string(raw)
	cases := []struct {
		heading, liveTest string
		valid             func() map[string]any
	}{
		{"### AC-FLH-012", "TestFactoryLiveCodexCodexWorktreeHandoff", liveGateInteractiveDoc},
		{"### AC-FLH-013", "TestFactoryLiveClaudeCodexWorktreeHandoff", liveGateHeadlessDoc},
	}
	for _, tc := range cases {
		pred := loadLiveGatePredicates(t, doc, tc.heading)
		valid := tc.valid()

		// Positive controls: without them every rejection below would be vacuous.
		if !runLiveGate(t, "-e", pred.evidence, liveGateJSON(t, valid)) {
			t.Fatalf("%s: valid evidence control rejected", tc.heading)
		}
		if !runLiveGate(t, "-se", pred.events, liveGateEvents(t, tc.liveTest)) {
			t.Fatalf("%s: valid event-stream control rejected", tc.heading)
		}

		rejected := 0
		expectReject := func(name string, flag, program string, input []byte) {
			if runLiveGate(t, flag, program, input) {
				t.Errorf("%s: mutant %s accepted", tc.heading, name)
				return
			}
			rejected++
		}

		for _, p := range liveGateLeafPaths(valid, nil) {
			m := liveGateClone(t, valid)
			liveGateSet(m, p, nil, true)
			expectReject("missing:"+strings.Join(p, "."), "-e", pred.evidence, liveGateJSON(t, m))
		}
		for _, field := range []string{"fixture", "mock", "direct_register_peer"} {
			m := liveGateClone(t, valid)
			liveGateSet(m, []string{"bypass", field}, true, false)
			expectReject("bypass."+field+"=true", "-e", pred.evidence, liveGateJSON(t, m))
		}
		expectReject("child_fail", "-se", pred.events, liveGateEvents(t, tc.liveTest,
			map[string]any{"Action": "fail", "Test": tc.liveTest + "/child"}))
		expectReject("child_skip", "-se", pred.events, liveGateEvents(t, tc.liveTest,
			map[string]any{"Action": "skip", "Test": tc.liveTest + "/child"}))
		expectReject("not_run_output", "-se", pred.events, liveGateEvents(t, tc.liveTest,
			map[string]any{"Action": "output", "Test": tc.liveTest, "Output": "NOT_RUN\n"}))

		if tc.heading == "### AC-FLH-013" {
			m := liveGateClone(t, valid)
			liveGateSet(m, []string{"app_server", "method"}, "thread/start", false)
			liveGateSet(m, []string{"app_server", "forked_from_id"}, nil, false)
			expectReject("wrong_method_thread_start", "-e", pred.evidence, liveGateJSON(t, m))
		}
		t.Logf("%s: 2 valid controls passed, %d mutants rejected", tc.heading, rejected)
	}
}
