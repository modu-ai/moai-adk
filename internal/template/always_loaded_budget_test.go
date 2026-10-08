package template

// TestDeployedAlwaysLoadedCharBudget is the deployed always-loaded instruction
// budget guard of SPEC-ALWAYS-LOADED-BUDGET-001 (REQ-ALB-001..005).
//
// The binding ledger fixture (testdata/binding_ledger.json) is a test fixture,
// never read at runtime (REQ-ALB-023). The expected surface values live in the
// ledger head and are re-derived here mechanically from the deployed tree.

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// deployedAlwaysLoadedCharBudget caps the deployed always-loaded instruction
// surface at 115000 UTF-16 code units (REQ-ALB-002). It sits 5000 units under
// the 120000-character runtime limit for 200K-context models (1M-context
// models carry a 150000-character limit), and the headroom is sized by the
// measured drift between diet periods: +4669 units over 7 days
// (b1ec8602c..b69cfe7ec) and +2627 units over 2 days (f4aa9bf99..b5815ca80)
// — the largest observed non-diet growth interval fits the headroom once, so
// this guard breaks in CI before a user-facing runtime limit is crossed
// (research.md §3; ledger head of binding_ledger.json).
const deployedAlwaysLoadedCharBudget = 115000

// deployedAlwaysLoadedPerFileCharBudget caps any single always-loaded rule
// file at 40000 UTF-16 code units (REQ-ALB-003, narrowed to always-surface
// members: top-level paths: companions are excluded from this check).
const deployedAlwaysLoadedPerFileCharBudget = 40000

// alwaysLoadedLedgerPath is the binding ledger fixture (test-only, never
// deployed, never read at runtime — REQ-ALB-023).
const alwaysLoadedLedgerPath = "testdata/binding_ledger.json"

// deployedSurfaceMember is one measured always-loaded surface file.
type deployedSurfaceMember struct {
	Path  string // project-root-relative slash path
	UTF16 int
}

// deployEmbeddedTemplatesForTest deploys the embedded template set into a
// fresh t.TempDir() project through the production deploy-and-render path
// (EmbeddedTemplates + NewRenderer + NewDeployerWithRenderer + manifest
// manager — the same wiring moai init uses) with the fixed render inputs
// (UserName="", ConversationLanguage="en", defaults otherwise).
func deployEmbeddedTemplatesForTest(t *testing.T) string {
	t.Helper()

	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	renderer := NewRenderer(embedded)
	deployer := NewDeployerWithRenderer(embedded, renderer)
	tmplCtx := NewTemplateContext(WithUser(""), WithLanguage("en"))

	ctx := context.Background()
	if err := deployer.ValidateAll(ctx, tmplCtx); err != nil {
		t.Fatalf("validate embedded templates: %v", err)
	}
	root := t.TempDir()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest for deploy: %v", err)
	}
	if err := deployer.Deploy(ctx, root, mgr, tmplCtx); err != nil {
		t.Fatalf("deploy embedded templates: %v", err)
	}
	return root
}

// utf16CodeUnits counts UTF-16 code units of s (the JS string-length unit:
// 1 per BMP rune, 2 per supplementary-plane rune).
func utf16CodeUnits(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// frontmatterPathsScoped reports whether the leading frontmatter block of a
// rules-tree markdown file carries a top-level `paths:` key. Mirrors the
// runtime rule (internal/hook/instructions_loaded.go ruleFileAlwaysLoaded):
// an indented `paths:` (leading whitespace) does NOT match the top-level
// prefix, so the file stays always-loaded; a missing closing `---` is also
// always-loaded (no top-level paths: is observable); no frontmatter at all is
// always-loaded.
func frontmatterPathsScoped(content []byte) bool {
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return false // no frontmatter block: always-loaded
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return false // closing delimiter reached: no top-level paths:
		}
		if strings.HasPrefix(lines[i], "paths:") {
			return true // paths:-scoped: loads on demand
		}
	}
	return false // unterminated frontmatter: always-loaded
}

// importTargets parses the @-import lines of a root-instruction file and
// returns the resolvable, in-project import paths (relative to the importing
// file's directory, slash form). Project-external imports (absolute, ~-,
// ..-escaping) and the AGENTS.local.md local instruction file are skipped.
func importTargets(importingFileRel string, content []byte) []string {
	dir := filepath.Dir(importingFileRel)
	var out []string
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "@") {
			continue
		}
		p := strings.TrimSpace(trimmed[1:])
		switch {
		case p == "", p == "AGENTS.local.md":
			continue
		case strings.HasPrefix(p, "/"), strings.HasPrefix(p, "~"):
			continue // project-external
		case p == ".." || strings.HasPrefix(p, "../"):
			continue // escapes the project
		}
		resolved := filepath.ToSlash(filepath.Join(dir, p))
		out = append(out, resolved)
	}
	return out
}

// deriveDeployedAlwaysLoadedMembers derives the always-loaded surface
// mechanically from the deployed tree at root (REQ-ALB-004):
//
//   - every .claude/rules/**/*.md carrying no top-level paths: key
//     (indented paths: and missing frontmatter closers stay always-loaded);
//   - the deployed root instruction file AGENTS.md (the SPEC-era CLAUDE.md
//     slot holder — CLAUDE.md was removed upstream; ledger head exclusion
//     note);
//   - the @-import transitive closure of the root instruction file,
//     interpreted in the deployed tree (project-external imports and
//     @AGENTS.local.md skipped);
//   - the AC-ALB-008 carve-out: the role-gated rules named by the binding
//     ledger's role-core: row locations count as surface members at their
//     deployed size (full body at this milestone, stub after M3), so the
//     budget keeps guarding them after M2/M3 gives them a top-level paths:.
//
// Members are returned sorted by path, deduplicated.
func deriveDeployedAlwaysLoadedMembers(t *testing.T, root string) []deployedSurfaceMember {
	t.Helper()

	members := make(map[string]int)

	add := func(rel string) {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return // not present in the deployed tree
		}
		members[rel] = utf16CodeUnits(string(data))
	}

	// 1. Always-loaded rules (frontmatter top-level paths: inspection).
	rulesRoot := filepath.Join(root, ".claude", "rules")
	_ = filepath.WalkDir(rulesRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil //nolint:nilerr // unreadable entries are skipped
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil //nolint:nilerr // unreadable files are skipped
		}
		if frontmatterPathsScoped(content) {
			return nil // paths:-scoped: on-demand, not session-start
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil //nolint:nilerr
		}
		members[filepath.ToSlash(rel)] = utf16CodeUnits(string(content))
		return nil
	})

	// 2. Root instruction file + 3. its @-import transitive closure.
	const rootInstruction = "AGENTS.md"
	add(rootInstruction)
	seen := map[string]bool{rootInstruction: true}
	queue := []string{rootInstruction}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(cur)))
		if err != nil {
			continue
		}
		for _, target := range importTargets(cur, data) {
			if seen[target] {
				continue
			}
			if _, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(target))); statErr != nil {
				continue // not part of the deployed tree
			}
			seen[target] = true
			add(target)
			queue = append(queue, target)
		}
	}

	// 4. AC-ALB-008 carve-out: role-gated rules from the ledger's role-core:
	// row locations stay surface members at their deployed size.
	for _, rulePath := range roleCoreRulePaths(t, root) {
		if _, ok := members[rulePath]; ok {
			continue
		}
		add(rulePath)
	}

	out := make([]deployedSurfaceMember, 0, len(members))
	for path, size := range members {
		out = append(out, deployedSurfaceMember{Path: path, UTF16: size})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// bindingLedgerHead is the subset of the ledger head the derivation reads.
type bindingLedgerHead struct {
	AnchorSHA string `json:"anchor_sha"`
}

// bindingLedger is the parsed fixture shape the tests need.
type bindingLedger struct {
	Head bindingLedgerHead `json:"head"`
	Rows []struct {
		ID         string `json:"id"`
		Kind       string `json:"kind"`
		AnchorKind string `json:"anchor_kind"`
		Location   string `json:"location"`
	} `json:"rows"`
}

// loadBindingLedger parses the committed ledger fixture.
func loadBindingLedger(t *testing.T) *bindingLedger {
	t.Helper()
	data, err := os.ReadFile(alwaysLoadedLedgerPath)
	if err != nil {
		t.Fatalf("read binding ledger fixture: %v", err)
	}
	var led bindingLedger
	if err := json.Unmarshal(data, &led); err != nil {
		t.Fatalf("parse binding ledger fixture: %v", err)
	}
	if len(led.Rows) == 0 {
		t.Fatalf("binding ledger fixture carries no rows")
	}
	return &led
}

// roleCoreRulePaths resolves the ledger's role-core: row locations to
// deployed rule paths (project-root-relative slash paths) inside the deployed
// tree at root. The ledger writes role-core locations as rule-relative paths
// (e.g. workflow/factory-dispatch.md); they resolve by suffix match against
// the deployed rules tree.
func roleCoreRulePaths(t *testing.T, root string) []string {
	t.Helper()
	led := loadBindingLedger(t)
	var suffixes []string
	for _, row := range led.Rows {
		if strings.HasPrefix(row.Location, "role-core:") {
			suffixes = append(suffixes, strings.TrimPrefix(row.Location, "role-core:"))
		}
	}
	if len(suffixes) == 0 {
		t.Logf("role-core carve-out sweep: 0 role-core rows in ledger (empty sweep is named, not silent)")
		return nil
	}
	var out []string
	seen := map[string]bool{}
	_ = filepath.WalkDir(filepath.Join(root, ".claude", "rules"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil //nolint:nilerr
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil //nolint:nilerr
		}
		slash := filepath.ToSlash(rel)
		for _, suf := range suffixes {
			if strings.HasSuffix(slash, "/"+suf) && !seen[suf] {
				out = append(out, slash)
				seen[suf] = true
			}
		}
		return nil
	})
	if len(out) < len(seen) { //nolint:staticcheck // defensive; seen dedupes by suffix
		t.Logf("role-core carve-out: %d of %d role-core suffixes unresolved in deployed tree", len(suffixes)-len(seen), len(suffixes))
	}
	t.Logf("role-core carve-out sweep: %d role-core rows -> %d deployed rule paths kept as surface members", len(suffixes), len(out))
	return out
}

// TestDeployedAlwaysLoadedCharBudget deploys the embedded template set through
// the production path, derives the always-loaded surface mechanically, and
// enforces the total and per-file char budgets (REQ-ALB-001..004).
//
// Expected state after M1: RED (anchor surface 180901 > 115000) and red until
// M3 lands — that is the deliverable, not a defect (plan.md M1).
func TestDeployedAlwaysLoadedCharBudget(t *testing.T) {
	root := deployEmbeddedTemplatesForTest(t)
	members := deriveDeployedAlwaysLoadedMembers(t, root)

	if len(members) == 0 {
		t.Fatal("derived surface is empty — derivation swept nothing, asserting nothing")
	}

	total := 0
	for _, m := range members {
		t.Logf("deployed-surface-member=%s %d", m.Path, m.UTF16)
		total += m.UTF16
	}
	t.Logf("deployed-surface-total=%d", total)

	// REQ-ALB-002: total budget failure names total, budget, file count, and
	// the five largest members with sizes.
	if total > deployedAlwaysLoadedCharBudget {
		largest := append([]deployedSurfaceMember(nil), members...)
		sort.Slice(largest, func(i, j int) bool { return largest[i].UTF16 > largest[j].UTF16 })
		if len(largest) > 5 {
			largest = largest[:5]
		}
		var sb strings.Builder
		for _, m := range largest {
			sb.WriteString(" " + m.Path + "=" + strconv.Itoa(m.UTF16) + ";")
		}
		t.Errorf("deployed always-loaded surface exceeds budget: total=%d budget=%d files=%d; five largest:%s",
			total, deployedAlwaysLoadedCharBudget, len(members), sb.String())
	}

	// REQ-ALB-003: any always-surface member over the per-file budget fails
	// naming the file and its size. Members only contains always-surface
	// files, so top-level paths: companions are excluded by derivation.
	for _, m := range members {
		if m.UTF16 > deployedAlwaysLoadedPerFileCharBudget {
			t.Errorf("always-loaded file exceeds per-file budget: file=%s size=%d budget=%d",
				m.Path, m.UTF16, deployedAlwaysLoadedPerFileCharBudget)
		}
	}
}
