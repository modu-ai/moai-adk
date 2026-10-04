// payload_test.go — SPEC-PLUGIN-MARKETPLACE-001 M2 payload guards (AC-004 to
// AC-008): derivation from the template tree independent of component names,
// copy/render/mode fidelity, the flat command layout, the allow-list and the
// payload MCP file.
package pluginemit_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/template/pluginemit"
)

const pluginRoot = "plugins/moai"

// nameSet is one set of synthetic component names: two core skills, one
// non-core skill, two commands (one a .tmpl).
type nameSet struct{ skillA, skillB, nonCore, cmdPlain, cmdTmpl string }

// derivationFixture builds a catalog and a template tree from a name set. The
// tree also carries scaffold-only content that must never reach the payload.
func derivationFixture(n nameSet) (string, map[string]string) {
	catalog := fmt.Sprintf(`version: 1.0.0
catalog:
  core:
    skills:
      - name: %[1]s
        tier: core
        path: templates/.claude/skills/%[1]s/
      - name: %[2]s
        tier: core
        path: templates/.claude/skills/%[2]s/
  optional_packs:
    extras:
      description: not shipped
      skills:
        - name: %[3]s
          tier: optional-pack:extras
          path: templates/.claude/skills/%[3]s/
`, n.skillA, n.skillB, n.nonCore)
	tree := map[string]string{
		".mcp.json": defaultMCPJSON,
		".claude/skills/" + n.skillA + "/SKILL.md":         "skill a {{ not a template }}\n",
		".claude/skills/" + n.skillA + "/scripts/tool.sh":  "#!/bin/sh\n",
		".claude/skills/" + n.skillA + "/modules/notes.md": "notes\n",
		".claude/skills/" + n.skillB + "/SKILL.md":         "skill b\n",
		".claude/skills/" + n.nonCore + "/SKILL.md":        "non-core\n",
		".claude/commands/moai/" + n.cmdPlain + ".md":      "plain command\n",
		".claude/commands/moai/" + n.cmdTmpl + ".md.tmpl":  "---\ndescription: {{if eq .ConversationLanguage \"ko\"}}korean{{else}}english{{end}}\n---\nbody\n",
		".claude/agents/moai/some-agent.md":                "agent\n",
		".claude/rules/moai/rule.md":                       "rule\n",
		".claude/hooks/moai/hook.sh":                       "#!/bin/sh\n",
		".claude/output-styles/moai/style.md":              "style\n",
		".claude/workflows/flow.md":                        "flow\n",
		"CLAUDE.md":                                        "instructions\n",
		"AGENTS.md":                                        "instructions\n",
	}
	return catalog, tree
}

func sortedKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// TestEmitDerivesFromTree is AC-004 (a): over a synthetic tree and catalog the
// emitted skill and command set equals exactly the synthetic core set, for two
// name sets that differ from each other and from the real tree.
func TestEmitDerivesFromTree(t *testing.T) {
	sets := map[string]nameSet{
		"first-names":  {"alpha-skill", "beta-skill", "gamma-skill", "kappa", "lambda"},
		"second-names": {"orion-skill", "vega-skill", "rigel-skill", "sirius", "altair"},
	}
	for name, n := range sets {
		t.Run(name, func(t *testing.T) {
			catalog, tree := derivationFixture(n)
			pub, err := pluginemit.Emit(syntheticRaw(catalog, tree), pluginemit.DefaultOptions())
			if err != nil {
				t.Fatalf("Emit: %v", err)
			}
			want := []string{
				pluginemit.ClaudeMarketplacePath, pluginemit.CodexMarketplacePath,
				pluginemit.ClaudePluginPath, pluginemit.CodexPluginPath,
				pluginRoot + "/.mcp.json",
				pluginRoot + "/skills/" + n.skillA + "/SKILL.md",
				pluginRoot + "/skills/" + n.skillA + "/scripts/tool.sh",
				pluginRoot + "/skills/" + n.skillA + "/modules/notes.md",
				pluginRoot + "/skills/" + n.skillB + "/SKILL.md",
				pluginRoot + "/commands/" + n.cmdPlain + ".md",
				pluginRoot + "/commands/" + n.cmdTmpl + ".md",
			}
			sort.Strings(want)
			got := sortedKeys(pub.Files)
			if strings.Join(got, "\n") != strings.Join(want, "\n") {
				t.Errorf("emitted set differs from the synthetic core set\n--- got ---\n%s\n--- want ---\n%s",
					strings.Join(got, "\n"), strings.Join(want, "\n"))
			}
			// The rendered command carries the English variant and no template action.
			rendered := string(pub.Files[pluginRoot+"/commands/"+n.cmdTmpl+".md"])
			if rendered != "---\ndescription: english\n---\nbody\n" {
				t.Errorf("rendered command = %q, want the English default render", rendered)
			}
			// A non-.tmpl file with brace text is copied byte for byte.
			if got := string(pub.Files[pluginRoot+"/skills/"+n.skillA+"/SKILL.md"]); got != "skill a {{ not a template }}\n" {
				t.Errorf("brace-bearing skill file = %q, want it copied unchanged", got)
			}
		})
	}
}

// TestEmitFidelity is AC-005 (a): over the real template tree every non-.tmpl
// payload file equals its source bytes, every .tmpl file equals the English
// default render with the suffix dropped, and every payload file carries the
// mode the deployer gives it.
func TestEmitFidelity(t *testing.T) {
	raw := os.DirFS(rawTemplateDir)
	pub, err := pluginemit.Emit(raw, pluginemit.DefaultOptions())
	if err != nil {
		t.Fatalf("Emit over the real template tree: %v", err)
	}
	cat, err := template.LoadCatalog(raw)
	if err != nil {
		t.Fatal(err)
	}
	view, err := template.SlimFS(raw, cat)
	if err != nil {
		t.Fatal(err)
	}
	renderer := template.NewRenderer(view)

	want := map[string]string{} // emitted path -> source path in the view
	err = fs.WalkDir(view, ".claude/skills", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		want[pluginRoot+"/skills/"+strings.TrimPrefix(p, ".claude/skills/")] = p
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	cmds, err := fs.ReadDir(view, ".claude/commands/moai")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cmds {
		stem := strings.TrimSuffix(strings.TrimSuffix(c.Name(), ".tmpl"), ".md")
		want[pluginRoot+"/commands/"+stem+".md"] = ".claude/commands/moai/" + c.Name()
	}
	if len(want) < 100 {
		t.Fatalf("source sweep found %d payload files; an empty sweep asserts nothing", len(want))
	}

	braceFiles, scriptFiles := 0, 0
	for dst, src := range want {
		got, ok := pub.Files[dst]
		if !ok {
			t.Errorf("%s: not emitted (source %s)", dst, src)
			continue
		}
		var expect []byte
		if strings.HasSuffix(src, ".tmpl") {
			expect, err = renderer.Render(src, template.NewTemplateContext())
			if err != nil {
				t.Fatalf("render %s: %v", src, err)
			}
		} else if expect, err = fs.ReadFile(view, src); err != nil {
			t.Fatal(err)
		}
		if string(got) != string(expect) {
			t.Errorf("%s: emitted bytes differ from the source %s", dst, src)
		}
		if strings.Contains(string(expect), "{{") && strings.HasPrefix(dst, pluginRoot+"/skills/") {
			braceFiles++
		}
		if strings.HasPrefix(dst, pluginRoot+"/commands/") && strings.Contains(string(got), "{{") {
			t.Errorf("%s: a rendered command carries a template action", dst)
		}
	}

	wantFiles := len(want) + 5 // the four manifests and the payload .mcp.json
	if len(pub.Files) != wantFiles {
		t.Errorf("emitted %d files, want %d (payload sources plus manifests plus .mcp.json)", len(pub.Files), wantFiles)
	}

	for p := range pub.Files {
		expectMode := fs.FileMode(0o644)
		if strings.HasSuffix(p, ".sh") {
			expectMode = 0o755
			scriptFiles++
		}
		if pub.Modes[p] != expectMode {
			t.Errorf("%s: mode %v, want %v (the deployer rule: .sh 0755, otherwise 0644)", p, pub.Modes[p], expectMode)
		}
	}
	// Controls: the sweep saw the live cases the criterion names.
	if braceFiles == 0 {
		t.Error("no brace-bearing skill file was seen; the byte-for-byte copy of such files was not exercised")
	}
	if scriptFiles < 3 {
		t.Errorf("saw %d .sh payload files, want at least the three navigator scripts; the mode rule was not exercised", scriptFiles)
	}
}

// TestPayloadCommandsFlat is AC-006 (a): one commands/<stem>.md per command
// stem of the template tree and no directory under commands/, in the emission
// and in the committed tree.
func TestPayloadCommandsFlat(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join(rawTemplateDir, "templates", ".claude", "commands", "moai"))
	if err != nil {
		t.Fatal(err)
	}
	var stems []string
	for _, e := range entries {
		stems = append(stems, strings.TrimSuffix(strings.TrimSuffix(e.Name(), ".tmpl"), ".md"))
	}
	if len(stems) == 0 {
		t.Fatal("no command stems found in the template tree; an empty sweep asserts nothing")
	}
	sort.Strings(stems)

	pub, err := pluginemit.Emit(os.DirFS(rawTemplateDir), pluginemit.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	var emitted []string
	for p := range pub.Files {
		rest, ok := strings.CutPrefix(p, pluginRoot+"/commands/")
		if !ok {
			continue
		}
		if strings.Contains(rest, "/") {
			t.Errorf("%s: a nested command path; the runtime inventory counts none of them", p)
		}
		emitted = append(emitted, strings.TrimSuffix(rest, ".md"))
	}
	sort.Strings(emitted)
	if strings.Join(emitted, ",") != strings.Join(stems, ",") {
		t.Errorf("emitted commands = %v, want one flat file per template stem %v", emitted, stems)
	}

	committed, err := os.ReadDir(filepath.Join(repoRoot, filepath.FromSlash(pluginRoot), "commands"))
	if err != nil {
		t.Fatalf("committed commands directory: %v — run `make plugin-emit`", err)
	}
	var onDisk []string
	for _, e := range committed {
		if e.IsDir() {
			t.Errorf("committed commands/%s is a directory", e.Name())
		}
		onDisk = append(onDisk, strings.TrimSuffix(e.Name(), ".md"))
	}
	sort.Strings(onDisk)
	if strings.Join(onDisk, ",") != strings.Join(stems, ",") {
		t.Errorf("committed commands = %v, want %v", onDisk, stems)
	}
}

// TestPayloadAllowList is AC-008 (b): a synthetic tree that gains scaffold-only
// directories emits only the five allowed top-level entries, and the committed
// plugin root holds exactly those five.
func TestPayloadAllowList(t *testing.T) {
	allowed := []string{".claude-plugin", ".codex-plugin", ".mcp.json", "commands", "skills"}

	catalog, tree := derivationFixture(nameSet{"alpha-skill", "beta-skill", "gamma-skill", "kappa", "lambda"})
	pub, err := pluginemit.Emit(syntheticRaw(catalog, tree), pluginemit.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	top := map[string]bool{}
	for p := range pub.Files {
		if rest, ok := strings.CutPrefix(p, pluginRoot+"/"); ok {
			top[strings.SplitN(rest, "/", 2)[0]] = true
		}
	}
	var got []string
	for k := range top {
		got = append(got, k)
	}
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(allowed, ",") {
		t.Errorf("emitted top-level entries = %v, want exactly %v", got, allowed)
	}

	entries, err := os.ReadDir(filepath.Join(repoRoot, filepath.FromSlash(pluginRoot)))
	if err != nil {
		t.Fatalf("committed plugin root: %v — run `make plugin-emit`", err)
	}
	var onDisk []string
	for _, e := range entries {
		onDisk = append(onDisk, e.Name())
	}
	sort.Strings(onDisk)
	if strings.Join(onDisk, ",") != strings.Join(allowed, ",") {
		t.Errorf("committed plugin root = %v, want exactly %v", onDisk, allowed)
	}
}

// TestPayloadMCPFile is AC-007 (a) and (b) at the generator level: the payload
// .mcp.json declares only moai and its command and args follow the template.
func TestPayloadMCPFile(t *testing.T) {
	const changed = `{"mcpServers": {
  "moai": {"command": "alt-launcher", "args": ["first", "second"]},
  "context7": {"command": "npx", "args": ["-y", "pkg"]}
}}`
	pub := emitSynthetic(t, changed)
	data, ok := pub.Files[pluginRoot+"/.mcp.json"]
	if !ok {
		t.Fatalf("emitted set lacks %s/.mcp.json", pluginRoot)
	}
	var doc struct {
		MCPServers map[string]pluginemit.MCPEntry `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("payload .mcp.json is not valid JSON: %v", err)
	}
	if len(doc.MCPServers) != 1 {
		t.Errorf("payload .mcp.json declares %d servers, want only moai", len(doc.MCPServers))
	}
	entry := doc.MCPServers["moai"]
	if entry.Command != "alt-launcher" || strings.Join(entry.Args, ",") != "first,second" {
		t.Errorf("payload entry = %+v, want the changed template entry copied", entry)
	}

	committed, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(pluginRoot), ".mcp.json"))
	if err != nil {
		t.Fatalf("committed payload .mcp.json: %v — run `make plugin-emit`", err)
	}
	tmplEntry, err := pluginemit.DeriveMCPEntry(os.DirFS(path.Join(rawTemplateDir, "templates")), ".mcp.json")
	if err != nil {
		t.Fatal(err)
	}
	var cdoc struct {
		MCPServers map[string]pluginemit.MCPEntry `json:"mcpServers"`
	}
	if err := json.Unmarshal(committed, &cdoc); err != nil {
		t.Fatal(err)
	}
	if len(cdoc.MCPServers) != 1 || cdoc.MCPServers["moai"].Command != tmplEntry.Command ||
		strings.Join(cdoc.MCPServers["moai"].Args, ",") != strings.Join(tmplEntry.Args, ",") {
		t.Errorf("committed payload .mcp.json = %v, want only the template moai entry %+v", cdoc.MCPServers, tmplEntry)
	}
}
