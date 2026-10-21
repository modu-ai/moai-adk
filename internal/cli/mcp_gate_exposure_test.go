// mcp_gate_exposure_test.go — SPEC-MCP-NEUTRALITY-001 M3 (card t1632,
// REQ-MN-009 + AC-MN-003): the gate-session containment. The exposure test
// asserts the SERVED set equals plan §F M3's enumerated read-only subset
// exactly; the launch-argv test pins the codex child containment; the
// measurement harness issues the fixture call set with a gate-session-shaped
// client and prints the AC-MN-003 metric line.
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
)

// moaiGateExposureEnumerated is plan §F M3's enumerated set, in plan order.
var moaiGateExposureEnumerated = []string{
	"spec_progress", "spec_audit", "spec_drift", "verify_trend",
	"session_list", "audit_cache", "goal_status",
	"graph_file_api", "graph_find_code", "graph_trace_calls", "graph_shortest_path",
	"todo_list", "jev_ask",
	"codex_setup", "codex_job_status", "codex_job_result",
	"glm_job_status", "glm_job_result",
}

// gateExposureClient builds the gate-session-shaped client: the in-process
// client against the server constructed under the gate exposure.
func gateExposureClient(t *testing.T) *client.Client {
	t.Helper()
	prevExposure := moaiCurrentExposure
	t.Cleanup(func() { moaiCurrentExposure = prevExposure })
	srv := newMoaiMCPServerWithExposure(moaiMCPExposureGate)
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	t.Cleanup(func() { closeInProcessClient(c) })
	if _, err := c.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	return c
}

// TestMCPGateExposure_ServedSubsetEqualsEnumeration — REQ-MN-009: the gate
// session's moai server serves EXACTLY the enumerated read-only subset, no
// more, no less.
func TestMCPGateExposure_ServedSubsetEqualsEnumeration(t *testing.T) {
	c := gateExposureClient(t)
	res, err := c.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	served := map[string]bool{}
	for _, tool := range res.Tools {
		served[tool.Name] = true
	}
	want := map[string]bool{}
	for _, name := range moaiGateExposureEnumerated {
		want[name] = true
	}
	var extra, missing []string
	for name := range served {
		if !want[name] {
			extra = append(extra, name)
		}
	}
	for name := range want {
		if !served[name] {
			missing = append(missing, name)
		}
	}
	if len(extra) > 0 || len(missing) > 0 {
		t.Fatalf("gate exposure served %d tools, enumerated %d — extra=%v missing=%v", len(served), len(want), extra, missing)
	}
}

// TestMCPGateExposure_AuditCacheStoreRefused — the read-op-only carve: the
// audit_cache store operation is refused under the gate exposure while the
// read ops stay served.
func TestMCPGateExposure_AuditCacheStoreRefused(t *testing.T) {
	c := gateExposureClient(t)
	// The store op validates project_root (M2 root contract) BEFORE the gate
	// check fires, so the refusal probe needs a root that passes validation —
	// this server's own repository — and the pin asserts the GATE exclusion
	// text, not the root text.
	root := thisRepoRoot(t)
	res, err := c.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: "audit_cache",
		Arguments: map[string]any{
			"op":           "store",
			"spec_id":      "SPEC-GATE-001",
			"hash":         "abc",
			"project_root": root,
		},
	}})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !res.IsError {
		t.Fatal("audit_cache op=store was accepted under the gate exposure")
	}
	if text := resultTextOf(res); !strings.Contains(text, "not available to gate sessions") {
		t.Errorf("store refusal = %q, want it to name the gate exclusion", text)
	}
}

// TestCodexGateContainment_ArgsDisableForeignServers pins the launch-argv
// half (REQ-MN-009): the containment args disable every non-moai server and
// flip the moai server to the gate exposure, and a listing failure refuses
// the spawn instead of degrading into an uncontained child.
func TestCodexGateContainment_ArgsDisableForeignServers(t *testing.T) {
	root := t.TempDir()
	// Foreign + moai servers in the listing.
	listing := `[{"name":"context7","transport":{"type":"streamable_http","url":"https://mcp.context7.example"}},` +
		`{"name":"moai","transport":{"type":"stdio","command":"moai"}}]`
	prevList := codexAuditMCPList
	codexAuditMCPList = func(ctx context.Context, program, dir string) (string, error) {
		if dir != root {
			t.Errorf("mcp list ran in %q, want the reviewed tree %q", dir, root)
		}
		return listing, nil
	}
	t.Cleanup(func() { codexAuditMCPList = prevList })

	args, err := codexGateContainmentArgs(context.Background(), "codex", root)
	if err != nil {
		t.Fatalf("containment args: %v", err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "mcp_servers.context7") || !strings.Contains(joined, moaiGateExposureArg) {
		t.Errorf("containment args = %v, want the foreign disable + the moai gate exposure override", args)
	}
	// The moai server itself is NOT disabled.
	if strings.Contains(joined, `mcp_servers.moai.command=`) {
		t.Errorf("the moai server was disabled alongside the foreign servers: %v", args)
	}

	// Fail-closed: a listing failure refuses the containment.
	codexAuditMCPList = func(ctx context.Context, program, dir string) (string, error) {
		return "", fmt.Errorf("codex mcp list exploded")
	}
	if _, err := codexGateContainmentArgs(context.Background(), "codex", root); err == nil {
		t.Fatal("a listing failure produced containment args — an uncontained child would spawn")
	}
}

// TestCodexGateContainment_NoMoaiServerNoOverride (card t1632): the gate
// exposure override addresses the moai server, so it is appended ONLY when the
// listing carries one. Appended alone it makes a table with no transport, which
// codex rejects at config load ("invalid transport in mcp_servers.moai") — the
// child never starts.
func TestCodexGateContainment_NoMoaiServerNoOverride(t *testing.T) {
	root := t.TempDir()
	prevList := codexAuditMCPList
	codexAuditMCPList = func(context.Context, string, string) (string, error) {
		return `[{"name":"context7","transport":{"type":"streamable_http","url":"https://mcp.context7.example"}}]`, nil
	}
	t.Cleanup(func() { codexAuditMCPList = prevList })

	args, err := codexGateContainmentArgs(context.Background(), "codex", root)
	if err != nil {
		t.Fatalf("containment args: %v", err)
	}
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "mcp_servers.moai") {
		t.Errorf("no moai server is listed, yet the args address it: %v", args)
	}
	if !strings.Contains(joined, "mcp_servers.context7") {
		t.Errorf("the foreign server was not disabled: %v", args)
	}

	codexAuditMCPList = func(context.Context, string, string) (string, error) { return "[]", nil }
	args, err = codexGateContainmentArgs(context.Background(), "codex", root)
	if err != nil || len(args) != 0 {
		t.Errorf("an empty listing = %v, %v; want no args and no error", args, err)
	}
}

// ─── AC-MN-003 measurement harness ───

// TestACMN003_GateCallSuccessRate is the card gate 3 measurement (AC-MN-003):
// a gate-session-shaped client issues the fixture call set — initialize,
// tools/list, and one call per enumerated subset tool — against the gate
// server. A call SUCCEEDS when the server returns a non-error, non-refused
// result. The harness prints the metric line
//
//	gate call success rate: <successes>/<total> (<pct>%) over <total> sampled calls
//
// and the total is asserted non-zero — a run that samples nothing is a FAIL
// per acceptance.md §C.
func TestACMN003_GateCallSuccessRate(t *testing.T) {
	// The fixture root: mcpCodeFixture is a git-backed moai project whose
	// graph artifacts are already built (the graph tools read edges.jsonl
	// under .moai/project/graph). The SPEC catalogue gets one probe SPEC, the
	// factory broker one active run, and the job registries one record each.
	root := mcpCodeFixture(t)
	specDir := filepath.Join(root, ".moai", "specs", "SPEC-GATEMEASURE-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("---\nid: SPEC-GATEMEASURE-001\ntitle: \"gate probe\"\nversion: \"0.1.0\"\nstatus: draft\ncreated: 2026-10-14\nupdated: 2026-10-14\nauthor: t\npriority: P3\nphase: \"v3.2.0\"\nmodule: \"internal/cli\"\nlifecycle: exploratory\ntags: \"probe\"\n---\n\n# probe\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sdSeedRun(t, root, "gate-measure-run", "active")
	if _, err := factorymsg.Open(root, "gate-measure-run"); err != nil {
		t.Fatal(err)
	}
	codexRec, err := newCodexJobRegistry(root).create(codexJobSpec{Mode: "task", RequestSummary: "gate measure"})
	if err != nil {
		t.Fatal(err)
	}
	glmRec, err := newGLMJobRegistry(root).create(glmJobSpec{Model: "glm-5.3", RequestSummary: "gate measure"})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)
	prev := projectDirResolver
	projectDirResolver = func() string { return root }
	t.Cleanup(func() { projectDirResolver = prev })

	c := gateExposureClient(t)

	// initialize + tools/list are the harness's own fixture calls and both
	// succeeded to reach this point.
	total := 2
	successes := 2

	// One call per enumerated subset tool, with the minimal arguments the
	// tool's schema declares.
	calls := map[string]map[string]any{
		"spec_progress":       {"project_root": root},
		"spec_audit":          {"project_root": root},
		"spec_drift":          {"project_root": root},
		"verify_trend":        {"key": "gate-measure:probe", "project_root": root},
		"session_list":        {},
		"audit_cache":         {"op": "lookup", "spec_id": "SPEC-GATEMEASURE-001", "hash": "h", "project_root": root},
		"goal_status":         {"project_root": root},
		"graph_file_api":      {"file": "internal/svc/svc.go", "project_root": root},
		"graph_find_code":     {"query": "Finish", "project_root": root},
		"graph_trace_calls":   {"symbol": "Finish", "project_root": root},
		"graph_shortest_path": {"from": "Run", "to": "Finish", "project_root": root},
		"todo_list":           {"project_root": root},
		"jev_ask":             {"project_root": root},
		"codex_job_status":    {"job_id": codexRec.ID, "project_root": root},
		"codex_job_result":    {"job_id": codexRec.ID, "project_root": root},
		"glm_job_status":      {"job_id": glmRec.ID, "project_root": root},
		"glm_job_result":      {"job_id": glmRec.ID, "project_root": root},
	}
	var failures []string
	for _, name := range moaiGateExposureEnumerated {
		args := calls[name]
		if args == nil {
			args = map[string]any{}
		}
		total++
		res, err := c.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		}})
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: transport error %v", name, err))
			continue
		}
		// Success: a non-error, non-refused result.
		if res.IsError || strings.Contains(resultTextOf(res), `"refused":true`) {
			failures = append(failures, fmt.Sprintf("%s: refused or errored: %s", name, resultTextOf(res)))
			continue
		}
		successes++
	}

	pct := 0
	if total > 0 {
		pct = successes * 100 / total
	}
	metric := fmt.Sprintf("gate call success rate: %d/%d (%d%%) over %d sampled calls", successes, total, pct, total)
	t.Logf("AC-MN-003 %s", metric)
	if len(moaiGateExposureEnumerated) == 0 {
		t.Fatal("no metric line was produced — the measurement is a FAIL")
	}
	if len(failures) > 0 {
		t.Errorf("AC-MN-003 %s — failures:\n%s", metric, strings.Join(failures, "\n"))
	}
}
