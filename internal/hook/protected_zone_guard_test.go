package hook

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/config"
)

// zoneSeven renders the seven required categories with empty lists: a manifest at the
// shipped path must declare them or it is invalid.
func zoneSeven() string {
	var b strings.Builder
	for _, c := range config.ProtectedZoneRequiredCategories {
		b.WriteString("  " + c + ":\n    paths: []\n")
	}
	return b.String()
}

// zoneShippedDoc is a valid shipped manifest carrying the seven categories plus the extra body.
func zoneShippedDoc(extra string) string {
	return "version: 1\ncategories:\n" + zoneSeven() + extra
}

// newZoneRoot creates a physical project root (macOS temp directories sit behind a
// symlink) holding a .moai directory and the given manifest files; an empty string
// leaves that manifest absent.
func newZoneRoot(t *testing.T, shipped, overlay string) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	for rel, body := range map[string]string{config.ProtectedZoneShippedRel: shipped, config.ProtectedZoneOverlayRel: overlay} {
		if body == "" {
			continue
		}
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// zoneTestHandler builds a PreToolUse handler pinned to root, with relative file
// paths resolved against root: the zone's own resolution reads zoneGetwd, and the
// existing file-access check reads the process working directory — the judge
// probe drives the binary from inside the root, so chdir gives the in-process
// handler the same resolution basis rather than a second one.
func zoneTestHandler(t *testing.T, root string) *preToolHandler {
	t.Helper()
	t.Chdir(root)
	prev := zoneGetwd
	zoneGetwd = func() (string, error) { return root, nil }
	t.Cleanup(func() { zoneGetwd = prev })
	return &preToolHandler{
		cfg:        &mockConfigProvider{cfg: newTestConfig()},
		policy:     DefaultSecurityPolicy(),
		projectDir: root,
	}
}

// zoneCall drives the real handler and returns the decision and its reason.
func zoneCall(t *testing.T, h *preToolHandler, tool, agent string, input map[string]any) (decision, reason string) {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	out, err := h.Handle(context.Background(), &HookInput{ToolName: tool, ToolInput: raw, AgentType: agent})
	if err != nil {
		t.Fatalf("Handle(%s): %v", tool, err)
	}
	if out == nil || out.HookSpecificOutput == nil {
		return "", ""
	}
	return out.HookSpecificOutput.PermissionDecision, out.HookSpecificOutput.PermissionDecisionReason
}

func zoneWrite(path string) map[string]any {
	return map[string]any{"file_path": path, "content": "x"}
}

func zoneEdit(path string) map[string]any {
	return map[string]any{"file_path": path, "old_string": "a", "new_string": "b"}
}

// wantZoneDeny asserts a deny carrying the new sentinel with the routing fields in order.
func wantZoneDeny(t *testing.T, name, decision, reason, identity, field, value string) {
	t.Helper()
	if decision != DecisionDeny {
		t.Errorf("%s: decision=%q reason=%q, want deny", name, decision, reason)
		return
	}
	prefix := SentinelHarnessFrozenProtectedZone + ": " + identity + " " + field + "=" + value + " route=human next=return-blocker-report path="
	if !strings.HasPrefix(reason, prefix) {
		t.Errorf("%s: reason %q does not start with %q", name, reason, prefix)
	}
}

const zoneProbeManifest = "  probe_zone:\n    paths: [\"zone_dir/\", \"**/*.zonefile\", \"한글_dir/\"]\n"

func testZoneFileTools(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc(zoneProbeManifest), "")
	h := zoneTestHandler(t, root)
	swept := 0

	deny := func(name, tool, path string) {
		t.Helper()
		swept++
		in := zoneWrite(path)
		if tool == "Edit" {
			in = zoneEdit(path)
		}
		d, r := zoneCall(t, h, tool, harnessLearnerIdentity, in)
		wantZoneDeny(t, name, d, r, harnessLearnerIdentity, "category", "probe_zone")
	}
	for _, tool := range []string{"Write", "Edit"} {
		deny(tool+" relative", tool, "zone_dir/a.md")
		deny(tool+" absolute", tool, filepath.Join(root, "zone_dir", "a.md"))
		deny(tool+" dot prefix", tool, "./zone_dir/a.md")
		deny(tool+" dot-dot detour", tool, "docs/../zone_dir/a.md")
		deny(tool+" letter case", tool, "ZONE_DIR/a.md")
		deny(tool+" basename glob", tool, "deep/er/x.zonefile")
		deny(tool+" basename glob case", tool, "deep/X.ZONEFILE")
		deny(tool+" nfd korean path", tool, "한글_dir/a.md")
	}

	allow := func(name, path string) {
		t.Helper()
		swept++
		d, r := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite(path))
		if strings.Contains(r, SentinelHarnessFrozenProtectedZone) || d == DecisionDeny {
			t.Errorf("%s: decision=%q reason=%q, want no zone deny", name, d, r)
		}
	}
	allow("outside every entry", "docs/a.md")
	allow("prefix-like sibling", "zone_dir_other/a.md")
	allow("empty file_path", "")
	allow("the project root itself", root)
	swept++
	if d, r := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("/definitely/outside/zone_dir/a.md")); strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("outside the root: decision=%q reason=%q, want the zone to leave it to the outside-project check", d, r)
	}

	// symlinks: a link into the zone, the root reached through a link, a target path reached through one
	t.Run("symlinks", func(t *testing.T) {
		real := newZoneRoot(t, zoneShippedDoc(zoneProbeManifest), "")
		if err := os.MkdirAll(filepath.Join(real, "zone_dir"), 0o755); err != nil {
			t.Fatal(err)
		}
		linkRoot := filepath.Join(filepath.Dir(real), filepath.Base(real)+"-link")
		if err := os.Symlink(real, linkRoot); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		if err := os.Symlink(filepath.Join(real, "zone_dir"), filepath.Join(real, "shortcut")); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		// a link into the zone directory
		hr := zoneTestHandler(t, real)
		swept++
		d, r := zoneCall(t, hr, "Write", harnessLearnerIdentity, zoneWrite("shortcut/a.md"))
		wantZoneDeny(t, "link into zone", d, r, harnessLearnerIdentity, "category", "probe_zone")
		// root reached through a symlink, path physical
		hl := zoneTestHandler(t, linkRoot)
		swept++
		d, r = zoneCall(t, hl, "Write", harnessLearnerIdentity, zoneWrite(filepath.Join(real, "zone_dir", "a.md")))
		wantZoneDeny(t, "root via symlink, path physical", d, r, harnessLearnerIdentity, "category", "probe_zone")
		// root physical, path reached through the symlinked root
		hr = zoneTestHandler(t, real)
		swept++
		d, r = zoneCall(t, hr, "Write", harnessLearnerIdentity, zoneWrite(filepath.Join(linkRoot, "zone_dir", "a.md")))
		wantZoneDeny(t, "root physical, path via symlink", d, r, harnessLearnerIdentity, "category", "probe_zone")
		// a dot-dot that crosses the link resolves physically (merge-gate round 1
		// P1-4, observed red first): lexically the file lands at the root, the
		// filesystem puts it inside the zone
		if err := os.MkdirAll(filepath.Join(real, "zone_dir", "sub"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(real, "zone_dir", "sub"), filepath.Join(real, "deep")); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		hr = zoneTestHandler(t, real)
		swept++
		d, r = zoneCall(t, hr, "Write", harnessLearnerIdentity, zoneWrite("deep/../secret.md"))
		wantZoneDeny(t, "dot-dot through link", d, r, harnessLearnerIdentity, "category", "probe_zone")
		swept++
		d, r = zoneCall(t, hr, "Bash", harnessLearnerIdentity, map[string]any{"command": "echo x > deep/../secret.md"})
		wantZoneDeny(t, "dot-dot through link, shell", d, r, harnessLearnerIdentity, "category", "probe_zone")
	})

	if swept < 26 {
		t.Fatalf("swept %d rows, want at least 26", swept)
	}
	t.Logf("swept=%d", swept)
}

func testZoneManifestStates(t *testing.T) {
	swept := 0
	hooksFile := ".claude/hooks/moai/x.sh"

	// present but invalid: every identity Write is denied, naming the failing file
	root := newZoneRoot(t, "::: not [valid yaml", "")
	h := zoneTestHandler(t, root)
	swept++
	d, r := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md"))
	wantZoneDeny(t, "invalid shipped", d, r, harnessLearnerIdentity, "manifest", "invalid")
	if !strings.HasSuffix(r, config.ProtectedZoneShippedRel) {
		t.Errorf("invalid shipped: reason %q does not end with the failing file %s", r, config.ProtectedZoneShippedRel)
	}
	// ... but a caller outside the identity set is untouched
	swept++
	if d, r := zoneCall(t, h, "Write", "manager-develop", zoneWrite("docs/a.md")); d == DecisionDeny {
		t.Errorf("non-identity caller with an invalid manifest: decision=%q reason=%q, want allow", d, r)
	}
	swept++
	if d, r := zoneCall(t, h, "Write", "", zoneWrite("docs/a.md")); d == DecisionDeny {
		t.Errorf("empty agent type with an invalid manifest: decision=%q reason=%q, want allow", d, r)
	}

	// absent: the compiled floor still holds, nothing beyond it
	root = newZoneRoot(t, "", "")
	h = zoneTestHandler(t, root)
	swept++
	d, r = zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite(hooksFile))
	if d != DecisionDeny || !strings.HasPrefix(r, SentinelHarnessFrozenHook+":") {
		t.Errorf("absent manifest, hooks file: decision=%q reason=%q, want the legacy hook sentinel", d, r)
	}
	swept++
	if d, r := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md")); d == DecisionDeny {
		t.Errorf("absent manifest, docs file: decision=%q reason=%q, want allow", d, r)
	}

	// valid: the manifest is what is read (docs/ is in no compiled list)
	root = newZoneRoot(t, zoneShippedDoc("  probe_docs:\n    paths: [\"docs/\"]\n"), "")
	h = zoneTestHandler(t, root)
	swept++
	d, r = zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md"))
	wantZoneDeny(t, "valid shipped lists docs/", d, r, harnessLearnerIdentity, "category", "probe_docs")

	// overlay only
	root = newZoneRoot(t, "", "version: 1\ncategories:\n  probe_docs:\n    paths: [\"docs/\"]\n")
	h = zoneTestHandler(t, root)
	swept++
	d, r = zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md"))
	wantZoneDeny(t, "overlay only", d, r, harnessLearnerIdentity, "category", "probe_docs")

	// overlay adds, never replaces: the base's entry is still protected
	root = newZoneRoot(t, zoneShippedDoc("  probe_base:\n    paths: [\"base_dir/\"]\n"), "version: 1\ncategories:\n  probe_docs:\n    paths: [\"docs/\"]\n")
	h = zoneTestHandler(t, root)
	swept++
	d, r = zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("base_dir/a.md"))
	wantZoneDeny(t, "overlay cannot narrow", d, r, harnessLearnerIdentity, "category", "probe_base")
	swept++
	d, r = zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md"))
	wantZoneDeny(t, "overlay entry", d, r, harnessLearnerIdentity, "category", "probe_docs")

	// overlay carrying an unknown key is invalid and the reason names the overlay
	root = newZoneRoot(t, zoneShippedDoc("  probe_base:\n    paths: [\"base_dir/\"]\n"),
		"version: 1\nexclude:\n  - base_dir/\ncategories:\n  probe_docs:\n    paths: [\"docs/\"]\n")
	h = zoneTestHandler(t, root)
	swept++
	d, r = zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/z.md"))
	wantZoneDeny(t, "overlay unknown key", d, r, harnessLearnerIdentity, "manifest", "invalid")
	if !strings.HasSuffix(r, config.ProtectedZoneOverlayRel) {
		t.Errorf("overlay unknown key: reason %q does not end with %s", r, config.ProtectedZoneOverlayRel)
	}

	// shipped file lacking one of the seven required categories is invalid (a shrunken zone must not load)
	root = newZoneRoot(t, "version: 1\ncategories:\n  logs:\n    paths: [\"x/\"]\n", "")
	h = zoneTestHandler(t, root)
	swept++
	d, r = zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md"))
	wantZoneDeny(t, "one-category shipped file", d, r, harnessLearnerIdentity, "manifest", "invalid")

	// a zero-byte overlay is valid and adds nothing (acceptance.md §D)
	root = newZoneRoot(t, zoneShippedDoc(""), "")
	overlayPath := filepath.Join(root, filepath.FromSlash(config.ProtectedZoneOverlayRel))
	if err := os.MkdirAll(filepath.Dir(overlayPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(overlayPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	h = zoneTestHandler(t, root)
	swept++
	if d, r = zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md")); d == DecisionDeny {
		t.Errorf("zero-byte overlay: decision=%q reason=%q, want allow", d, r)
	}

	// duplicate entries are valid (acceptance.md §D): the path is denied once either way
	root = newZoneRoot(t, zoneShippedDoc("  probe_docs:\n    paths: [\"docs/\", \"docs/\"]\n"), "")
	h = zoneTestHandler(t, root)
	swept++
	d, r = zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md"))
	wantZoneDeny(t, "duplicate entries", d, r, harnessLearnerIdentity, "category", "probe_docs")

	// no agent-invocable bypass: the environment, an extra tool-input field, nor a file the
	// identity could write changes a denial
	root = newZoneRoot(t, zoneShippedDoc("  probe_docs:\n    paths: [\"docs/\"]\n"), "")
	h = zoneTestHandler(t, root)
	t.Setenv("MOAI_BRANCH_GUARD_EXEMPT", "1")
	t.Setenv("MOAI_PROTECTED_ZONE", "off")
	swept++
	d, r = zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md"))
	wantZoneDeny(t, "environment variable", d, r, harnessLearnerIdentity, "category", "probe_docs")
	swept++
	in := zoneWrite("docs/a.md")
	in["override"] = true
	in["protected_zone"] = "off"
	d, r = zoneCall(t, h, "Write", harnessLearnerIdentity, in)
	wantZoneDeny(t, "extra tool-input field", d, r, harnessLearnerIdentity, "category", "probe_docs")

	if swept < 15 {
		t.Fatalf("swept %d rows, want at least 15", swept)
	}
	t.Logf("swept=%d", swept)
}

func testZoneNonRegression(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc(""), "")
	h := zoneTestHandler(t, root)
	swept := 0

	// baseline-matched controls keep their legacy sentinel and reason, byte for byte, with no routing field
	for _, c := range []struct {
		tool, path, sentinel, want string
	}{
		{"Write", ".claude/hooks/moai/x.sh", SentinelHarnessFrozenHook, SentinelHarnessFrozenHook + ": harness-learner cannot modify frozen path .claude/hooks/moai/x.sh"},
		{"Write", ".claude/rules/moai/core/moai-constitution.md", SentinelHarnessFrozenRule, SentinelHarnessFrozenRule + ": harness-learner cannot modify frozen path .claude/rules/moai/core/moai-constitution.md"},
		{"Write", "CLAUDE.md", SentinelHarnessFrozenInstruction, SentinelHarnessFrozenInstruction + ": harness-learner cannot modify frozen instruction file CLAUDE.md"},
		{"Edit", ".claude/agents/moai/plan-auditor.md", SentinelHarnessFrozenAgent, SentinelHarnessFrozenAgent + ": harness-learner cannot modify frozen path .claude/agents/moai/plan-auditor.md"},
	} {
		swept++
		in := zoneWrite(c.path)
		if c.tool == "Edit" {
			in = zoneEdit(c.path)
		}
		d, r := zoneCall(t, h, c.tool, harnessLearnerIdentity, in)
		if d != DecisionDeny || r != c.want {
			t.Errorf("%s %s: decision=%q reason=%q, want deny with exactly %q", c.tool, c.path, d, r, c.want)
		}
		if strings.Contains(r, "route=") || strings.Contains(r, "category=") {
			t.Errorf("%s %s: a baseline denial carries routing fields: %q", c.tool, c.path, r)
		}
	}

	// the identity's legitimate surface stays open
	for _, p := range []string{
		".moai/specs/SPEC-ANY-001/spec.md", ".claude/agents/harness/my-specialist.md",
		".claude/skills/hns-example/SKILL.md", ".moai/harness/main.md", "docs/a.md",
	} {
		swept++
		if d, r := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite(p)); d == DecisionDeny {
			t.Errorf("identity write to %s: decision=%q reason=%q, want allow", p, d, r)
		}
	}

	// callers outside the identity set, and tools other than Write/Edit/Bash, are unchanged
	for _, agent := range []string{"", "manager-develop", "manager-git", "general-purpose"} {
		swept++
		if d, r := zoneCall(t, h, "Write", agent, zoneWrite(".moai/config/sections/protected-zone.yaml")); strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
			t.Errorf("agent %q: decision=%q reason=%q, want no zone denial", agent, d, r)
		}
	}
	for _, tool := range []string{"Read", "Glob", "Grep", "MultiEdit", "NotebookEdit", "WebFetch"} {
		swept++
		if d, r := zoneCall(t, h, tool, harnessLearnerIdentity, map[string]any{"file_path": ".claude/settings.json"}); strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
			t.Errorf("tool %s: decision=%q reason=%q, want no zone denial", tool, d, r)
		}
	}
	if swept < 19 {
		t.Fatalf("swept %d rows, want at least 19", swept)
	}
	t.Logf("swept=%d", swept)
}

func testZoneDenyReason(t *testing.T) {
	cases := []struct {
		name, field, value, path string
	}{
		{"short", "category", "safety_guards", "internal/hook/pre_tool.go"},
		{"32 byte category", "category", strings.Repeat("a", 32), "internal/hook/pre_tool.go"},
		{"long ascii path", "category", "safety_guards", "internal/hook/" + strings.Repeat("a", 220) + ".go"},
		{"long multibyte path", "category", "safety_guards", "internal/hook/" + strings.Repeat("가", 90) + ".go"},
		{"invalid manifest long file", "manifest", "invalid", ".moai/" + strings.Repeat("d/", 200) + "protected-zone.yaml"},
		{"baseline pseudo category", "category", zoneBaselineCategory, ".claude/hooks/moai/x.sh"},
	}
	for _, c := range cases {
		reason := zoneDenyReason(harnessLearnerIdentity, c.field, c.value, c.path)
		prefix := SentinelHarnessFrozenProtectedZone + ": " + harnessLearnerIdentity + " " + c.field + "=" + c.value + " route=human next=return-blocker-report path="
		if !strings.HasPrefix(reason, prefix) {
			t.Errorf("%s: reason %q lost a routing field or the order", c.name, reason)
		}
		if len(reason) > zoneReasonMax {
			t.Errorf("%s: reason is %d bytes, want at most %d", c.name, len(reason), zoneReasonMax)
		}
		if !utf8.ValidString(reason) {
			t.Errorf("%s: reason is not valid UTF-8 (truncation cut a rune)", c.name)
		}
		if !strings.HasPrefix(c.path, strings.TrimPrefix(reason, prefix)) {
			t.Errorf("%s: the tail of the reason is not a prefix of the path: %q", c.name, reason)
		}
	}
	// a short path survives whole
	if got := zoneDenyReason(harnessLearnerIdentity, "category", "logs", ".moai/logs/a.log"); !strings.HasSuffix(got, "path=.moai/logs/a.log") {
		t.Errorf("short path was altered: %q", got)
	}

	// through the real handler: no absolute path, no manifest content, the same bound
	root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n"), "")
	h := zoneTestHandler(t, root)
	for _, p := range []string{
		"zone_dir/" + strings.Repeat("a", 220) + ".go",
		"zone_dir/" + strings.Repeat("가", 90) + ".go",
		filepath.Join(root, "zone_dir", "a.md"),
	} {
		d, r := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite(p))
		wantZoneDeny(t, "handler "+p[:min(len(p), 30)], d, r, harnessLearnerIdentity, "category", "probe_zone")
		if len(r) > zoneReasonMax || !utf8.ValidString(r) {
			t.Errorf("handler reason is %d bytes (valid=%v)", len(r), utf8.ValidString(r))
		}
		if strings.Contains(r, root) || strings.Contains(r, "version: 1") {
			t.Errorf("handler reason leaks an absolute path or manifest content: %q", r)
		}
	}
	if len(cases) != 6 {
		t.Fatalf("swept %d cases", len(cases))
	}
	t.Logf("swept=%d", len(cases)+3)
}

func testZoneNoManifestReadForOthers(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n"), "")
	h := zoneTestHandler(t, root)
	reads := 0
	h.zoneLoader = func(r string) config.ProtectedZoneLoad {
		reads++
		return config.LoadProtectedZone(r)
	}
	swept := 0
	for _, agent := range []string{"", "manager-develop", "manager-git", "general-purpose"} {
		for _, call := range []struct {
			tool  string
			input map[string]any
		}{
			{"Write", zoneWrite("zone_dir/a.md")},
			{"Edit", zoneEdit("zone_dir/a.md")},
			{"Bash", map[string]any{"command": "rm zone_dir/a.md"}},
		} {
			swept++
			zoneCall(t, h, call.tool, agent, call.input)
		}
	}
	for _, tool := range []string{"Read", "Glob", "Grep", "MultiEdit", "NotebookEdit"} {
		swept++
		zoneCall(t, h, tool, harnessLearnerIdentity, map[string]any{"file_path": "zone_dir/a.md"})
	}
	if reads != 0 {
		t.Fatalf("the manifest was read %d times for callers or tools the guard does not apply to", reads)
	}
	// positive control: the seam counts, so a zero above is a measurement
	swept++
	zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("zone_dir/a.md"))
	if reads != 1 {
		t.Fatalf("an identity Write read the manifest %d times, want exactly 1 (the read-counter seam is dead)", reads)
	}
	t.Logf("swept=%d", swept)
}

func testZoneAuditRow(t *testing.T) {
	readRows := func(root string) []map[string]string {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(zoneAuditRel)))
		if err != nil {
			return nil
		}
		var rows []map[string]string
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if line == "" {
				continue
			}
			var row map[string]string
			if err := json.Unmarshal([]byte(line), &row); err != nil {
				t.Fatalf("audit line is not JSON: %q: %v", line, err)
			}
			rows = append(rows, row)
		}
		return rows
	}
	wantFields := []string{"ts", "identity", "tool", "path", "category", "decision", "manifest_state"}
	checkRow := func(t *testing.T, row map[string]string, want map[string]string) {
		t.Helper()
		for _, f := range wantFields {
			if _, ok := row[f]; !ok {
				t.Errorf("audit row lacks field %q: %v", f, row)
			}
		}
		for k, v := range want {
			if row[k] != v {
				t.Errorf("audit row %s=%q, want %q (row %v)", k, row[k], v, row)
			}
		}
	}

	t.Run("Deny", func(t *testing.T) {
		root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n"), "")
		h := zoneTestHandler(t, root)
		d, _ := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("zone_dir/a.md"))
		rows := readRows(root)
		if d != DecisionDeny || len(rows) != 1 {
			t.Fatalf("decision=%q rows=%d, want a deny and exactly one row", d, len(rows))
		}
		checkRow(t, rows[0], map[string]string{"identity": harnessLearnerIdentity, "tool": "Write", "path": "zone_dir/a.md", "category": "probe_zone", "decision": "deny", "manifest_state": "ok"})
		// an allowed call under a healthy manifest leaves no row
		zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md"))
		if got := len(readRows(root)); got != 1 {
			t.Errorf("a healthy allow appended a row: %d rows", got)
		}
	})

	t.Run("Absent", func(t *testing.T) {
		root := newZoneRoot(t, "", "")
		h := zoneTestHandler(t, root)
		d, _ := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md"))
		rows := readRows(root)
		if d == DecisionDeny || len(rows) != 1 {
			t.Fatalf("decision=%q rows=%d, want an allow that records the degraded state once", d, len(rows))
		}
		checkRow(t, rows[0], map[string]string{"decision": "allow", "manifest_state": "absent", "path": "docs/a.md"})
	})

	t.Run("Invalid", func(t *testing.T) {
		root := newZoneRoot(t, "::: not [valid yaml", "")
		h := zoneTestHandler(t, root)
		zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md"))
		rows := readRows(root)
		if len(rows) != 1 {
			t.Fatalf("rows=%d, want 1", len(rows))
		}
		checkRow(t, rows[0], map[string]string{"decision": "deny", "manifest_state": "invalid"})
	})

	t.Run("AppendFailure", func(t *testing.T) {
		root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n"), "")
		// a regular file where the logs directory must be: the append cannot succeed
		if err := os.WriteFile(filepath.Join(root, ".moai", "logs"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		h := zoneTestHandler(t, root)
		d, r := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("zone_dir/a.md"))
		wantZoneDeny(t, "unwritable audit log", d, r, harnessLearnerIdentity, "category", "probe_zone")
		if err := zoneAppendAudit(root, zoneAuditRow{}); err == nil {
			t.Errorf("zoneAppendAudit into an unwritable location returned no error")
		}
		// a decision made while the log is writable is unchanged by the failure above
		d2, _ := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("docs/a.md"))
		if d2 == DecisionDeny {
			t.Errorf("allow turned into %q after a failed append", d2)
		}
	})
}

// zoneGoStringList reads the string elements of a package-level []string variable from a Go source file.
func zoneGoStringList(t *testing.T, file, name string) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	var out []string
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, id := range vs.Names {
				if id.Name != name || i >= len(vs.Values) {
					continue
				}
				lit, ok := vs.Values[i].(*ast.CompositeLit)
				if !ok {
					continue
				}
				for _, el := range lit.Elts {
					if bl, ok := el.(*ast.BasicLit); ok && bl.Kind == token.STRING {
						s, err := strconv.Unquote(bl.Value)
						if err != nil {
							t.Fatal(err)
						}
						out = append(out, s)
					}
				}
			}
		}
	}
	if len(out) == 0 {
		t.Fatalf("no string elements found for %s in %s", name, file)
	}
	return out
}

func testZoneBaselineCovered(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repo := filepath.Join(filepath.Dir(thisFile), "..", "..")

	type baselineEntry struct{ list, entry string }
	// The 21 list entries the manifest must cover, enumerated by name so a removed or
	// renamed compiled entry fails by that name and not only by a count.
	want := []baselineEntry{
		{"hook", ".claude/agents/moai/"}, {"hook", ".claude/skills/moai-"}, {"hook", ".claude/rules/moai/"},
		{"hook", ".claude/commands/"}, {"hook", ".claude/hooks/"}, {"hook", ".claude/output-styles/"},
		{"hook", "CLAUDE.md"}, {"hook", "CLAUDE.local.md"}, {"hook", "AGENTS.md"}, {"hook", "AGENTS.local.md"},
		{"safety", ".claude/agents/moai/"}, {"safety", ".claude/skills/moai-"}, {"safety", ".claude/rules/moai/"},
		{"safety", ".claude/settings.json"}, {"safety", ".claude/settings.local.json"},
		{"safety", "internal/harness/safety/frozen_guard.go"}, {"safety", "internal/harness/frozen_guard.go"},
		{"harness", ".claude/agents/moai/"}, {"harness", ".claude/skills/moai-"}, {"harness", ".claude/skills/moai/"},
		{"harness", ".claude/rules/moai/"},
	}

	// the lists as they are compiled, so adding an entry to one without the manifest fails
	var have []baselineEntry
	for _, fz := range frozenZonePrefixes {
		have = append(have, baselineEntry{"hook", fz.prefix})
	}
	for _, f := range frozenInstructionFiles {
		have = append(have, baselineEntry{"hook", f})
	}
	for _, e := range zoneGoStringList(t, filepath.Join(repo, "internal", "harness", "safety", "frozen_guard.go"), "frozenPrefixes") {
		have = append(have, baselineEntry{"safety", e})
	}
	for _, e := range zoneGoStringList(t, filepath.Join(repo, "internal", "harness", "frozen_guard.go"), "frozenPrefixes") {
		have = append(have, baselineEntry{"harness", e})
	}
	for _, w := range want {
		if !slices.Contains(have, w) {
			t.Errorf("enumerated baseline entry %s %q is no longer in the compiled list (renamed or removed)", w.list, w.entry)
		}
	}
	for _, h := range have {
		if !slices.Contains(want, h) {
			t.Errorf("compiled list entry %s %q is not in the enumerated 21: add it here and to the manifest", h.list, h.entry)
		}
	}

	load := config.LoadProtectedZone(repo)
	if load.State != config.ZoneStateOK {
		t.Fatalf("dogfood zone state=%q file=%q err=%v", load.State, load.InvalidFile, load.Err)
	}
	covered := func(rel string) bool {
		f := config.FoldZoneText(rel)
		for _, e := range load.Zone.Entries {
			if e.Match(f) {
				return true
			}
		}
		return false
	}
	swept := 0
	for _, w := range want {
		swept++
		var rep string
		switch {
		case slices.Contains(frozenInstructionFiles, w.entry):
			rep = "probe/" + w.entry
		case strings.HasSuffix(w.entry, "/"):
			rep = w.entry + "probe.txt"
		case strings.HasSuffix(w.entry, "-"):
			rep = w.entry + "probe"
		default:
			rep = w.entry
		}
		if !covered(rep) {
			t.Errorf("baseline entry %s %q (representative %q) is not covered by the dogfood effective zone", w.list, w.entry, rep)
		}
	}
	if swept != 21 {
		t.Fatalf("swept %d baseline entries, want 21", swept)
	}
	t.Logf("swept=%d", swept)
}

// testZoneShellMutation drives the thirteen mutating shell forms of REQ-SIPZ-007
// against a zone member, with the read-only and non-identity controls that keep
// the rule scoped, and the command-text prefix that must not suppress it.
func testZoneShellMutation(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n"), "")
	h := zoneTestHandler(t, root)
	swept := 0

	forms := []string{
		"rm zone_dir/a.log",                    // B1
		"unlink zone_dir/a.log",                // B2
		"mv zone_dir/a.log /tmp/elsewhere.log", // B3
		"cp docs/keep.md zone_dir/a.log",       // B4
		"echo x | tee zone_dir/a.log",          // B5
		"truncate -s 0 zone_dir/a.log",         // B6
		"sed -i s/a/b/ zone_dir/a.log",         // B7
		"echo x > zone_dir/a.log",              // B8
		"echo x >> zone_dir/a.log",             // B9
		"git rm zone_dir/a.log",                // B10
		"git checkout zone_dir/a.log",          // B11
		"git restore zone_dir/a.log",           // B12
		"git apply zone_dir/patch.diff",        // B13
	}
	if len(forms) != 13 {
		t.Fatalf("shell form table carries %d rows, want the thirteen of REQ-SIPZ-007", len(forms))
	}
	for _, cmd := range forms {
		swept++
		d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": cmd})
		wantZoneDeny(t, cmd, d, r, harnessLearnerIdentity, "category", "probe_zone")
	}

	// read-only verbs on the same path stay open (AC-SIPZ-005 N6 mirror)
	for _, cmd := range []string{"cat zone_dir/a.log", "ls zone_dir/"} {
		swept++
		if d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": cmd}); d == DecisionDeny {
			t.Errorf("%q: decision=%q reason=%q, want allow", cmd, d, r)
		}
	}
	// the same rm from a non-identity caller stays open (AC-SIPZ-005 N3 mirror)
	swept++
	if d, r := zoneCall(t, h, "Bash", "manager-develop", map[string]any{"command": "rm zone_dir/a.log"}); d == DecisionDeny {
		t.Errorf("non-identity rm: decision=%q reason=%q, want allow", d, r)
	}
	// a command-text prefix and a trailing comment do not suppress (REQ-SIPZ-013 mirror)
	swept++
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": "MOAI_PROTECTED_ZONE=off rm zone_dir/a.log # override"})
	wantZoneDeny(t, "assignment prefix and comment", d, r, harnessLearnerIdentity, "category", "probe_zone")

	// review-repair rows (merge-gate round 1, observed red first): the directory
	// itself, inline and repeated redirections, and the segment's cwd
	for _, cmd := range []string{
		"rm -rf zone_dir",                  // the directory itself, not only its children
		"mv zone_dir /tmp/elsewhere",       // moving the directory wholesale
		"echo x>zone_dir/a.log",            // inline redirection, no space
		"echo x > docs/a > zone_dir/a.log", // every target, not only the first
		"rm docs/missing >zone_dir/a.log",  // a redirect target after a mutating verb
		"cd zone_dir && rm a.log",          // the segment's working directory moved
	} {
		swept++
		d, r = zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": cmd})
		wantZoneDeny(t, cmd, d, r, harnessLearnerIdentity, "category", "probe_zone")
	}
	// control: an absolute candidate after a cd is judged where it lands
	swept++
	if d, r = zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": "cd zone_dir && rm /tmp/elsewhere"}); d == DecisionDeny {
		t.Errorf("absolute rm after cd: decision=%q reason=%q, want allow", d, r)
	}

	// review-repair round 2 rows (merge-gate re-verdict, observed red first)
	root = newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n  probe_moai:\n    paths: [\".moai/config/sections/\"]\n"), "")
	h = zoneTestHandler(t, root)
	for _, cmd := range []string{
		"true & rm zone_dir/a.log",                               // a single & separates commands
		"cd . > zone_dir/a.log",                                  // the cd segment's own redirection
		"sed -e s/a/b/ -i '' zone_dir/a.log",                     // -i not in the first option slot
		"git -c core.quotePath=false checkout -- zone_dir/a.log", // global options before the subcommand
	} {
		swept++
		d, r = zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": cmd})
		wantZoneDeny(t, cmd, d, r, harnessLearnerIdentity, "category", "probe_zone")
	}
	// moving a parent of protected entries denies with the contained entry's
	// category — the ancestor pass names what would be lost
	swept++
	d, r = zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": "mv .moai moved"})
	wantZoneDeny(t, "mv .moai moved", d, r, harnessLearnerIdentity, "category", "probe_moai")
	// control: a quoted > is string data, not a redirection operator
	swept++
	if d, r = zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": "echo '>zone_dir/a.log'"}); d == DecisionDeny {
		t.Errorf("quoted > data: decision=%q reason=%q, want allow", d, r)
	}

	if swept < 30 {
		t.Fatalf("swept %d rows, want at least 30", swept)
	}
	t.Logf("swept=%d", swept)
}

// testZoneLiveness is the continued-firing check (REQ-SIPZ-015): each way the
// guard can go quietly is a named condition that fails a CI run — the matcher
// group of both settings files, both manifests parsing, a dead paths entry in
// either sweep tree, and the real handler denying a known zone-member input.
func testZoneLiveness(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repo := filepath.Join(filepath.Dir(thisFile), "..", "..")

	t.Run("MatcherGroup", func(t *testing.T) {
		for _, name := range []string{
			filepath.Join(".claude", "settings.json"),
			filepath.Join("internal", "template", "templates", ".claude", "settings.json.tmpl"),
		} {
			data, err := os.ReadFile(filepath.Join(repo, name))
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			if !strings.Contains(string(data), `"matcher": "Write|Edit|Bash"`) {
				t.Errorf("%s: the PreToolUse matcher group no longer carries Write, Edit and Bash (REQ-SIPZ-015 i)", name)
			}
		}
	})

	t.Run("ManifestParse", func(t *testing.T) {
		load := config.LoadProtectedZone(repo)
		if load.State != config.ZoneStateOK {
			t.Errorf("dogfood manifests: state=%q file=%q err=%v, want ok", load.State, load.InvalidFile, load.Err)
		}
		data, err := os.ReadFile(filepath.Join(repo, "internal", "template", "templates", filepath.FromSlash(config.ProtectedZoneShippedRel)))
		if err != nil {
			t.Fatalf("template shipped manifest: %v", err)
		}
		if _, err := config.ParseProtectedZone(data, true, config.ProtectedZoneShippedRel); err != nil {
			t.Errorf("template shipped manifest does not parse (REQ-SIPZ-015 ii): %v", err)
		}
	})

	t.Run("DeadEntries", func(t *testing.T) {
		load := config.LoadProtectedZone(repo)
		if load.State != config.ZoneStateOK {
			t.Fatalf("dogfood manifests: state=%q", load.State)
		}
		wantPaths, wantRuntime := 0, 0
		for i := range load.Zone.Entries {
			if load.Zone.Entries[i].Runtime {
				wantRuntime++
			} else {
				wantPaths++
			}
		}
		templateTree := os.DirFS(filepath.Join(repo, "internal", "template", "templates"))
		shippedSweep, err := config.SweepZoneEntries(load.Zone, config.ProtectedZoneShippedRel, templateTree)
		if err != nil {
			t.Fatalf("shipped sweep: %v", err)
		}
		overlaySweep, err := config.SweepZoneEntries(load.Zone, config.ProtectedZoneOverlayRel, os.DirFS(repo))
		if err != nil {
			t.Fatalf("overlay sweep: %v", err)
		}
		for _, e := range shippedSweep.Dead {
			t.Errorf("dead shipped entry %q (category %s) matches nothing in the template tree (REQ-SIPZ-015 iii)", e.Raw, e.Category)
		}
		for _, e := range overlaySweep.Dead {
			t.Errorf("dead overlay entry %q (category %s) matches nothing in the local tree (REQ-SIPZ-015 iii)", e.Raw, e.Category)
		}
		resolved := shippedSweep.Resolved + overlaySweep.Resolved
		skipped := shippedSweep.Skipped + overlaySweep.Skipped
		if resolved != wantPaths {
			t.Errorf("sweep resolved %d paths entries, want %d (the number the files declare)", resolved, wantPaths)
		}
		if skipped != wantRuntime {
			t.Errorf("sweep skipped %d runtime_paths entries, want %d", skipped, wantRuntime)
		}
		t.Logf("resolved=%d skipped=%d", resolved, skipped)
	})

	t.Run("HandlerDenies", func(t *testing.T) {
		root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n"), "")
		h := zoneTestHandler(t, root)
		d, r := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite("zone_dir/a.md"))
		wantZoneDeny(t, "end to end", d, r, harnessLearnerIdentity, "category", "probe_zone")
	})
}
