package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// setupMCPProviderFixture points both path seams at t.TempDir() files so the
// check never reads the real home. It returns the project root and the state
// file path; the caller writes whichever fixtures the case needs.
func setupMCPProviderFixture(t *testing.T) (projectRoot, statePath string) {
	t.Helper()
	base := t.TempDir()
	projectRoot = filepath.Join(base, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	statePath = filepath.Join(base, "state", ".claude.json")
	globalPath := filepath.Join(base, "home", ".claude", ".mcp.json")

	prevState, prevGlobal := mcpProviderStatePath, mcpProviderGlobalMCPPath
	mcpProviderStatePath = func() string { return statePath }
	mcpProviderGlobalMCPPath = func() string { return globalPath }
	t.Cleanup(func() {
		mcpProviderStatePath, mcpProviderGlobalMCPPath = prevState, prevGlobal
	})
	return projectRoot, statePath
}

func writeMCPProviderFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

const providerProjectMCP = `{"mcpServers":{"context7":{"command":"npx"},"playwright":{"command":"npx"}}}`

func providerState(t *testing.T, projectRoot, disabled string) string {
	t.Helper()
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	projects := `{}`
	if disabled != "" {
		projects = `{"` + abs + `":{"disabledMcpServers":` + disabled + `}}`
	}
	return `{"claudeAiMcpEverConnected":["claude.ai Context7","claude.ai Notion"],"projects":` + projects + `}`
}

func TestCheckMCPProviderDuplicates_MatchWarns(t *testing.T) {
	root, state := setupMCPProviderFixture(t)
	writeMCPProviderFile(t, filepath.Join(root, ".mcp.json"), providerProjectMCP)
	writeMCPProviderFile(t, state, providerState(t, root, ""))

	check := checkMCPProviderDuplicates(root, false)

	if check.Name != mcpProviderDuplicatesCheckName {
		t.Errorf("Name = %q, want %q", check.Name, mcpProviderDuplicatesCheckName)
	}
	if check.Status != uikit.CheckWarn {
		t.Fatalf("Status = %q, want warn; msg=%s", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "context7 (.mcp.json) + claude.ai Context7") {
		t.Errorf("Message = %q, want the context7 pair", check.Message)
	}
	if strings.Contains(check.Message, "Notion") {
		t.Errorf("Message = %q, Notion has no local counterpart", check.Message)
	}
	if !strings.Contains(check.Detail, "/mcp") || !strings.Contains(check.Detail, "twice") {
		t.Errorf("Detail must be set on warn with remediation, got %q", check.Detail)
	}
}

func TestCheckMCPProviderDuplicates_DisabledLocalNameSuppresses(t *testing.T) {
	root, state := setupMCPProviderFixture(t)
	writeMCPProviderFile(t, filepath.Join(root, ".mcp.json"), providerProjectMCP)
	writeMCPProviderFile(t, state, providerState(t, root, `["context7"]`))

	check := checkMCPProviderDuplicates(root, false)
	if check.Status != uikit.CheckOK {
		t.Errorf("Status = %q, want ok; msg=%s", check.Status, check.Message)
	}
}

func TestCheckMCPProviderDuplicates_DisabledConnectorSuppresses(t *testing.T) {
	root, state := setupMCPProviderFixture(t)
	writeMCPProviderFile(t, filepath.Join(root, ".mcp.json"), providerProjectMCP)
	writeMCPProviderFile(t, state, providerState(t, root, `["claude.ai Context7"]`))

	check := checkMCPProviderDuplicates(root, false)
	if check.Status != uikit.CheckOK {
		t.Errorf("Status = %q, want ok; msg=%s", check.Status, check.Message)
	}
}

func TestCheckMCPProviderDuplicates_NullDisabledStillWarns(t *testing.T) {
	root, state := setupMCPProviderFixture(t)
	writeMCPProviderFile(t, filepath.Join(root, ".mcp.json"), providerProjectMCP)
	writeMCPProviderFile(t, state, providerState(t, root, `null`))

	check := checkMCPProviderDuplicates(root, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("Status = %q, want warn; msg=%s", check.Status, check.Message)
	}
}

func TestCheckMCPProviderDuplicates_UserScopeServerMatches(t *testing.T) {
	root, state := setupMCPProviderFixture(t)
	writeMCPProviderFile(t, state,
		`{"mcpServers":{"notion":{"command":"npx"}},"claudeAiMcpEverConnected":["claude.ai Notion"]}`)

	check := checkMCPProviderDuplicates(root, false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("Status = %q, want warn; msg=%s", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "notion") || !strings.Contains(check.Message, "claude.ai Notion") {
		t.Errorf("Message = %q, want the notion pair", check.Message)
	}
}

func TestCheckMCPProviderDuplicates_NoStateFileOK(t *testing.T) {
	root, _ := setupMCPProviderFixture(t)
	writeMCPProviderFile(t, filepath.Join(root, ".mcp.json"), providerProjectMCP)

	check := checkMCPProviderDuplicates(root, false)
	if check.Status != uikit.CheckOK {
		t.Errorf("Status = %q, want ok; msg=%s", check.Status, check.Message)
	}
	if check.Message == "" {
		t.Error("Message must not be empty")
	}
}

func TestCheckMCPProviderDuplicates_MalformedStateFileOK(t *testing.T) {
	root, state := setupMCPProviderFixture(t)
	writeMCPProviderFile(t, filepath.Join(root, ".mcp.json"), providerProjectMCP)
	writeMCPProviderFile(t, state, `{not json`)

	check := checkMCPProviderDuplicates(root, false)
	if check.Status != uikit.CheckOK {
		t.Errorf("Status = %q, want ok; msg=%s", check.Status, check.Message)
	}
}

func TestCheckMCPProviderDuplicates_NonMatchingNamesOK(t *testing.T) {
	root, state := setupMCPProviderFixture(t)
	writeMCPProviderFile(t, filepath.Join(root, ".mcp.json"),
		`{"mcpServers":{"playwright":{"command":"npx"}}}`)
	writeMCPProviderFile(t, state, providerState(t, root, ""))

	check := checkMCPProviderDuplicates(root, false)
	if check.Status != uikit.CheckOK {
		t.Errorf("Status = %q, want ok; msg=%s", check.Status, check.Message)
	}
}

func TestResolveClaudeStatePath_UsesConfigDirWhenSet(t *testing.T) {
	env := map[string]string{"CLAUDE_CONFIG_DIR": "/profiles/x"}
	got := resolveClaudeStatePath(func(k string) string { return env[k] }, "/home/u")
	want := filepath.Join("/profiles/x", ".claude.json")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveClaudeStatePath_FallsBackToHome(t *testing.T) {
	got := resolveClaudeStatePath(func(string) string { return "" }, "/home/u")
	want := filepath.Join("/home/u", ".claude.json")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCheckMCPProviderDuplicates_Registered(t *testing.T) {
	setupMCPProviderFixture(t)
	results := runDiagnosticChecks(false, mcpProviderDuplicatesCheckName)
	if len(results) != 1 || results[0].Name != mcpProviderDuplicatesCheckName {
		t.Fatalf("expected one %q result, got %+v", mcpProviderDuplicatesCheckName, results)
	}
}
