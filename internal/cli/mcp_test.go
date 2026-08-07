// mcp_test.go: SPEC-TREND-MCP-001 M2 — generic `moai mcp add|remove|list` CLI
// table-driven coverage.
//
// Exercise the new internal/cli/mcp.go subcommands against a t.TempDir fixture
// (project scope writes to <tmpdir>/.mcp.json). Verifies:
//
//   - AC-TMC-005: atomic-RMW add via mutateClaudeJSONAtomic (flock + backup +
//     publish); unrelated entries preserved.
//   - AC-TMC-005 (concurrent-writer sub-scenario): the claudeJSONGuardPreLockHook
//     injection simulates a non-cooperating writer; the compare-retry path
//     recovers within claudeJSONGuardMaxRetries attempts and BOTH the external
//     write + the new entry survive.
//   - AC-TMC-006: idempotent add (second invocation is a no-change skip; only
//     ONE .mcp.json.bak-* file exists after both runs).
//   - AC-TMC-007: partial-remove safety (only the named entry is removed; every
//     unrelated mcpServers entry is preserved).
//   - AC-TMC-008: list --json emits valid JSON, distinguishes stdio vs http,
//     and flags entries whose env carries a ${VAR} literal.
//   - AC-TMC-009: --env KEY=VAL rejects any VALUE that is not a ${VAR} literal;
//     exit is non-zero and NO entry is written.
//   - AC-TMC-010: TestMCP_NoAskUserQuestion (in mcp_boundary_test.go).
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// baselineProjectMCP is the distributed 3-entry default seeded into each test
// fixture. Tests assert additions/removals relative to this baseline.
const baselineProjectMCP = `{
  "$schema": "https://raw.githubusercontent.com/anthropics/claude-code/main/.mcp.schema.json",
  "mcpServers": {
    "context7": {
      "command": "/bin/bash",
      "args": ["-l", "-c", "exec npx -y @upstash/context7-mcp@latest"]
    },
    "chrome-devtools": {
      "command": "/bin/bash",
      "args": ["-l", "-c", "exec npx -y chrome-devtools-mcp@latest --headless"]
    },
    "playwright": {
      "command": "/bin/bash",
      "args": ["-l", "-c", "exec npx -y @playwright/mcp@latest"]
    }
  }
}
`

// writeBaselineMCP seeds <dir>/.mcp.json with the distributed 3-entry default.
// Tests then exercise moaiMcpAdd/moaiMcpRemove/moaiMcpList against it.
func writeBaselineMCP(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(path, []byte(baselineProjectMCP), 0o600); err != nil {
		t.Fatalf("seed .mcp.json: %v", err)
	}
	return path
}

// countBackupFiles returns the number of `.claude.json.bak-*` files in dir.
// `backupClaudeJSON` (glm_tools.go:655) names every backup `.claude.json.bak-<ts>`
// regardless of whether the file under management is `.mcp.json` (project) or
// `.claude.json` (user) — a pre-existing naming convention preserved unchanged
// per REQ-TMC-008. Tests assert against that actual prefix.
func countBackupFiles(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir %s: %v", dir, err)
	}
	n := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".claude.json.bak-") {
			n++
		}
	}
	return n
}

// readServers reads <dir>/.mcp.json and returns the mcpServers map.
func readServers(t *testing.T, dir string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".mcp.json"))
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}
	var doc struct {
		Servers map[string]any `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal .mcp.json: %v", err)
	}
	return doc.Servers
}

// TestMCP_Add_RegistersEntry_PreservesUnrelated — AC-TMC-005 baseline scenario.
func TestMCP_Add_RegistersEntry_PreservesUnrelated(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeBaselineMCP(t, dir)

	if err := moaiMcpAdd(dir, mcpAddArgs{
		name:    "my-tool",
		command: "npx",
		args:    []string{"-y", "my-tool-mcp"},
		scope:   "project",
	}); err != nil {
		t.Fatalf("moaiMcpAdd: %v", err)
	}

	servers := readServers(t, dir)
	if _, ok := servers["my-tool"]; !ok {
		t.Errorf("my-tool entry was not registered: %v", servers)
	}
	for _, required := range []string{"context7", "chrome-devtools", "playwright"} {
		if _, ok := servers[required]; !ok {
			t.Errorf("unrelated entry %q was not preserved: %v", required, servers)
		}
	}
	if n := countBackupFiles(t, dir); n < 1 {
		t.Errorf("expected at least one .mcp.json.bak-* after add, got %d", n)
	}
}

// TestMCP_Add_IdempotentSkip — AC-TMC-006.
func TestMCP_Add_IdempotentSkip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeBaselineMCP(t, dir)

	args := mcpAddArgs{
		name:    "my-tool",
		command: "npx",
		args:    []string{"-y", "my-tool-mcp"},
		scope:   "project",
	}
	if err := moaiMcpAdd(dir, args); err != nil {
		t.Fatalf("first moaiMcpAdd: %v", err)
	}
	// Second invocation MUST be an idempotent skip — no new backup, no error.
	if err := moaiMcpAdd(dir, args); err != nil {
		t.Fatalf("second (idempotent) moaiMcpAdd: %v", err)
	}
	if n := countBackupFiles(t, dir); n != 1 {
		t.Errorf("expected exactly ONE .mcp.json.bak-* after two identical adds, got %d", n)
	}
}

// TestMCP_Add_ConcurrentWriter — AC-TMC-005 concurrent-writer sub-scenario.
// Inject a non-cooperating external write between the prep-read and the in-lock
// compare via the package-global claudeJSONGuardPreLockHook; the guard's
// compare-retry path MUST recover and BOTH the external write + the new entry
// survive.
//
// Deliberately NOT t.Parallel(): writes the package-global hook.
func TestMCP_Add_ConcurrentWriter(t *testing.T) {
	dir := t.TempDir()
	writeBaselineMCP(t, dir)

	origHook := claudeJSONGuardPreLockHook
	claudeJSONGuardPreLockHook = func(path string) {
		data, err := os.ReadFile(path)
		if err != nil {
			return
		}
		var root map[string]any
		if json.Unmarshal(data, &root) != nil {
			return
		}
		root["externalSetting"] = "from-cc"
		out, err := json.MarshalIndent(root, "", "  ")
		if err != nil {
			return
		}
		_ = os.WriteFile(path, out, 0o600)
	}
	t.Cleanup(func() { claudeJSONGuardPreLockHook = origHook })

	if err := moaiMcpAdd(dir, mcpAddArgs{
		name:    "my-tool",
		command: "npx",
		args:    []string{"-y", "my-tool-mcp"},
		scope:   "project",
	}); err != nil {
		t.Fatalf("moaiMcpAdd under concurrent writer: %v", err)
	}

	// External write MUST survive.
	data, err := os.ReadFile(filepath.Join(dir, ".mcp.json"))
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}
	var doc struct {
		External interface{}    `json:"externalSetting"`
		Servers  map[string]any `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if doc.External != "from-cc" {
		t.Errorf("external concurrent write lost: got %v", doc.External)
	}
	if _, ok := doc.Servers["my-tool"]; !ok {
		t.Errorf("my-tool entry lost after guarded RMW: %v", doc.Servers)
	}
}

// TestMCP_Remove_PartialDelete — AC-TMC-007.
func TestMCP_Remove_PartialDelete(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Seed 4 entries: 3 baseline + zai-mcp-server (the GLM entry, to confirm
	// the partial-delete safety preserves the GLM entry alongside the
	// distributed defaults).
	seed := `{
  "mcpServers": {
    "context7": {"command":"npx","args":["-y","@upstash/context7-mcp@latest"]},
    "chrome-devtools": {"command":"npx","args":["-y","chrome-devtools-mcp@latest"]},
    "playwright": {"command":"npx","args":["-y","@playwright/mcp@latest"]},
    "zai-mcp-server": {"command":"npx","args":["-y","zai-mcp-server"]},
    "my-tool": {"command":"npx","args":["-y","my-tool-mcp"]}
  }
}
`
	path := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := moaiMcpRemove(dir, "my-tool", "project"); err != nil {
		t.Fatalf("moaiMcpRemove: %v", err)
	}

	servers := readServers(t, dir)
	if _, ok := servers["my-tool"]; ok {
		t.Errorf("my-tool was NOT removed: %v", servers)
	}
	for _, required := range []string{"context7", "chrome-devtools", "playwright", "zai-mcp-server"} {
		if _, ok := servers[required]; !ok {
			t.Errorf("unrelated entry %q was lost on partial-delete: %v", required, servers)
		}
	}
}

// TestMCP_Remove_NoEntryIsNoChange — partial-delete safety on a missing entry.
func TestMCP_Remove_NoEntryIsNoChange(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeBaselineMCP(t, dir)
	before, err := os.ReadFile(filepath.Join(dir, ".mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := moaiMcpRemove(dir, "does-not-exist", "project"); err != nil {
		t.Errorf("moaiMcpRemove on missing entry should be a no-op, got: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(dir, ".mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("moaiMcpRemove on missing entry rewrote .mcp.json (expected byte-identical)")
	}
}

// TestMCP_List_JSON — AC-TMC-008.
func TestMCP_List_JSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Seed 4 entries: 3 stdio + 1 HTTP with ${VAR} literal.
	seed := `{
  "mcpServers": {
    "context7": {"command":"npx","args":["-y","@upstash/context7-mcp@latest"]},
    "chrome-devtools": {"command":"npx","args":["-y","chrome-devtools-mcp@latest"]},
    "playwright": {"command":"npx","args":["-y","@playwright/mcp@latest"]},
    "semgrep": {
      "type":"http",
      "url":"https://semgrep.example.com/mcp",
      "headers":{"Authorization":"Bearer ${SEMGREP_API_TOKEN}"}
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := moaiMcpList(&buf, dir, "project", true); err != nil {
		t.Fatalf("moaiMcpList: %v", err)
	}

	var out struct {
		Scope   string `json:"scope"`
		Entries []struct {
			Name    string         `json:"name"`
			Type    string         `json:"type"`
			EnvRefs []string       `json:"env_refs,omitempty"`
			Entry   map[string]any `json:"entry"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("list output is not valid JSON: %v\nraw=%s", err, buf.String())
	}
	if len(out.Entries) != 4 {
		t.Errorf("list entries = %d, want 4", len(out.Entries))
	}
	// Distinguish stdio from http.
	httpSeen := false
	stdioSeen := false
	for _, e := range out.Entries {
		if e.Type == "http" {
			httpSeen = true
		}
		if e.Type == "stdio" {
			stdioSeen = true
		}
		if e.Name == "semgrep" {
			if e.Type != "http" {
				t.Errorf("semgrep entry type = %q, want http", e.Type)
			}
			// Flag the ${VAR}-literal env reference (in headers for HTTP).
			if !containsAny(buf.String(), "${SEMGREP_API_TOKEN}") {
				t.Errorf("semgrep ${VAR} literal not surfaced in list output: %s", buf.String())
			}
		}
	}
	if !httpSeen || !stdioSeen {
		t.Errorf("list did not distinguish http+stdio: http=%v stdio=%v", httpSeen, stdioSeen)
	}
}

func containsAny(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// TestMCP_Add_SecretRejection — AC-TMC-009.
func TestMCP_Add_SecretRejection(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeBaselineMCP(t, dir)

	err := moaiMcpAdd(dir, mcpAddArgs{
		name:    "bad",
		command: "npx",
		args:    []string{"-y", "my-tool-mcp"},
		env:     []string{"API_KEY=sk-secret-123"},
		scope:   "project",
	})
	if err == nil {
		t.Fatal("moaiMcpAdd with a positional secret value MUST fail; got nil")
	}
	// Error message MUST point to the ${VAR}-literal form.
	if !strings.Contains(err.Error(), "${") {
		t.Errorf("secret-rejection error does not point to ${VAR} form: %v", err)
	}
	// NO entry MUST be written.
	servers := readServers(t, dir)
	if _, ok := servers["bad"]; ok {
		t.Errorf("bad entry was written despite secret rejection: %v", servers)
	}
}

// TestMCP_Add_SecretLiteralAccepted — the ${VAR}-literal form is accepted.
func TestMCP_Add_SecretLiteralAccepted(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeBaselineMCP(t, dir)

	if err := moaiMcpAdd(dir, mcpAddArgs{
		name:    "ok",
		command: "npx",
		args:    []string{"-y", "my-tool-mcp"},
		env:     []string{"API_KEY=${MY_API_KEY}"},
		scope:   "project",
	}); err != nil {
		t.Fatalf("moaiMcpAdd with ${VAR} literal MUST succeed: %v", err)
	}
	servers := readServers(t, dir)
	entry, ok := servers["ok"].(map[string]any)
	if !ok {
		t.Fatalf("ok entry missing or wrong shape: %v", servers["ok"])
	}
	env, _ := entry["env"].(map[string]any)
	if env["API_KEY"] != "${MY_API_KEY}" {
		t.Errorf("env API_KEY = %v, want ${MY_API_KEY}", env["API_KEY"])
	}
}

// TestMCP_Add_HTTPType — --type http + --url + --headers constructs an HTTP entry.
func TestMCP_Add_HTTPType(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeBaselineMCP(t, dir)

	if err := moaiMcpAdd(dir, mcpAddArgs{
		name:    "my-http",
		typeArg: "http",
		url:     "https://my-http.example.com/mcp",
		scope:   "project",
	}); err != nil {
		t.Fatalf("moaiMcpAdd --type http: %v", err)
	}
	servers := readServers(t, dir)
	entry, ok := servers["my-http"].(map[string]any)
	if !ok {
		t.Fatalf("my-http entry missing: %v", servers["my-http"])
	}
	if entry["type"] != "http" {
		t.Errorf("type = %v, want http", entry["type"])
	}
	if entry["url"] != "https://my-http.example.com/mcp" {
		t.Errorf("url = %v, want https://my-http.example.com/mcp", entry["url"])
	}
}

// TestNewMCPCmd_Registered verifies the cobra command tree wiring. The
// `moai mcp` parent + add/remove/list children MUST all be registered.
func TestNewMCPCmd_Registered(t *testing.T) {
	t.Parallel()
	cmd := newMCPCmd()
	subs := cmd.Commands()
	if len(subs) < 3 {
		t.Errorf("newMCPCmd children = %d, want >= 3 (add/remove/list)", len(subs))
	}
	names := map[string]bool{}
	for _, c := range subs {
		names[c.Use] = true
	}
	for _, required := range []string{"add", "remove", "list"} {
		// cobra Use strings may include arg hints ("add <name>"); match prefix.
		found := false
		for n := range names {
			if strings.HasPrefix(n, required) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not registered under moai mcp: %v", required, names)
		}
	}
}

// verify the secret-literal regex matches the spec (^[A-Z_][A-Z0-9_]*$ inside
// ${...}). Keeping it local to the test guards against accidental widening.
var secretLiteralRe = regexp.MustCompile(`^\$\{[A-Z_][A-Z0-9_]*\}$`)

func TestSecretLiteralRegex(t *testing.T) {
	t.Parallel()
	positive := []string{"${API_KEY}", "${Z_AI_API_KEY}", "${A}", "${_X9}", "${INSIDE_VAR}"}
	negative := []string{"sk-123", "${lower-case}", "${nested-${x}}", "${}", "$VAR"}
	for _, p := range positive {
		if !secretLiteralRe.MatchString(p) {
			t.Errorf("regex should accept %q", p)
		}
	}
	for _, n := range negative {
		if secretLiteralRe.MatchString(n) {
			t.Errorf("regex should reject %q", n)
		}
	}
}
