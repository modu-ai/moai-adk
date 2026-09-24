package spec

// lint_tier_artifacts.go — TierArtifactMissingRule (card t1121). Reports a
// SPEC directory that lacks a file its frontmatter `tier:` requires.
//
// The required set per tier is NOT hardcoded here. Its single source of truth
// is the Tier table in `.claude/rules/moai/workflow/spec-workflow.md`
// § SPEC Complexity Tier (S/M/L), which ships to user projects through the
// embedded templates. The rule reads that table from the project root at lint
// time, so a table edit changes enforcement without a second copy to update.
//
// Three states, each with a distinct outcome:
//   - rule file absent       → no findings (the project does not carry the
//     tier doctrine; fail-open, lint exit unaffected)
//   - rule file present, but the table does not yield all three S/M/L sets
//     → ONE corpus warning, TierArtifactSetUnreadable, so a broken SSOT is
//     visible instead of silently disabling the check
//   - table parsed           → per-SPEC TierArtifactMissing warnings
//
// One TierArtifactMissing finding per SPEC lists every missing file. Severity
// is warning; the code is deliberately NOT in eraDemotableCodes (that map
// demotes errors only). Per-SPEC findings still pass through lint.skip and
// the era/terminal-status advisory marking in Linter.Lint.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// tierArtifactRuleRelPath is the rule file's path relative to the project
// root, slash-separated (it doubles as the embedded-template path).
const tierArtifactRuleRelPath = ".claude/rules/moai/workflow/spec-workflow.md"

var (
	// tierRowPattern matches the first cell of a Tier row: "S (Simple)" etc.
	tierRowPattern = regexp.MustCompile(`^([SML]) \(`)
	// tierParenPattern strips parenthetical prose from the artifact cell so
	// "(AC inline in spec.md §3)" does not contribute a file.
	tierParenPattern = regexp.MustCompile(`\([^)]*\)`)
	// tierFilePattern extracts artifact file names from the artifact cell.
	tierFilePattern = regexp.MustCompile(`[A-Za-z0-9_-]+\.md`)
)

// parseTierArtifactSets reads the Tier table from spec-workflow.md content.
// It locates a table header carrying an "Artifact set" column, then takes the
// S/M/L rows beneath it. ok is true only when all three tiers yield a
// non-empty file list; any partial parse is reported as not-ok so the caller
// surfaces it rather than enforcing a subset.
func parseTierArtifactSets(content string) (map[string][]string, bool) {
	sets := map[string][]string{}
	col := -1
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			col = -1 // a table ended; the header binding ends with it
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		if col < 0 {
			for i, c := range cells {
				if strings.EqualFold(c, "Artifact set") {
					col = i
				}
			}
			continue
		}
		m := tierRowPattern.FindStringSubmatch(cells[0])
		if m == nil || col >= len(cells) {
			continue
		}
		if _, seen := sets[m[1]]; seen {
			continue // first occurrence wins
		}
		var files []string
		dedup := map[string]bool{}
		for _, f := range tierFilePattern.FindAllString(tierParenPattern.ReplaceAllString(cells[col], ""), -1) {
			if !dedup[f] {
				dedup[f] = true
				files = append(files, f)
			}
		}
		if len(files) > 0 {
			sets[m[1]] = files
		}
	}
	for _, tier := range []string{"S", "M", "L"} {
		if len(sets[tier]) == 0 {
			return nil, false
		}
	}
	return sets, true
}

// tierTableState is the parsed Tier table of one project root.
type tierTableState struct {
	present  bool
	sets     map[string][]string
	readable bool
}

// tierArtifactTable resolves and caches the Tier table per project root for
// one Linter, shared by the per-SPEC rule and the corpus-warning rule. The
// root is derived from each SPEC's own spec.md path, so the result does not
// depend on the working directory the CLI ran from (its BaseDir is
// cwd-derived); fallbackRoot (from BaseDir) is used only when the SPEC path
// does not sit under <root>/.moai/specs/<SPEC>/.
type tierArtifactTable struct {
	fallbackRoot string
	mu           sync.Mutex
	byRoot       map[string]*tierTableState
}

// stateFor returns the (cached) table state for root.
func (t *tierArtifactTable) stateFor(root string) *tierTableState {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.byRoot == nil {
		t.byRoot = map[string]*tierTableState{}
	}
	if st, ok := t.byRoot[root]; ok {
		return st
	}
	st := &tierTableState{}
	if data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(tierArtifactRuleRelPath))); err == nil {
		st.present = true
		st.sets, st.readable = parseTierArtifactSets(string(data))
	}
	// A read error (absent or unreadable file) leaves the feature unconfigured.
	t.byRoot[root] = st
	return st
}

// rootFor maps a SPEC's spec.md path to its project root: the directory that
// contains .moai/specs/<SPEC>/. Falls back to the BaseDir-derived root.
func (t *tierArtifactTable) rootFor(specPath string) string {
	if abs, err := filepath.Abs(specPath); err == nil {
		specsDir := filepath.Dir(filepath.Dir(abs))
		if filepath.Base(specsDir) == "specs" && filepath.Base(filepath.Dir(specsDir)) == ".moai" {
			return filepath.Dir(filepath.Dir(specsDir))
		}
	}
	return t.fallbackRoot
}

// lintProjectRoot maps the Linter BaseDir to the project root. The CLI passes
// <root>/.moai/specs when it exists and the cwd otherwise, so this is only the
// fallback when a SPEC's own path cannot name its root.
func lintProjectRoot(baseDir string) string {
	if baseDir == "" {
		baseDir = "."
	}
	clean := filepath.Clean(baseDir)
	if filepath.Base(clean) == "specs" && filepath.Base(filepath.Dir(clean)) == ".moai" {
		return filepath.Dir(filepath.Dir(clean))
	}
	return clean
}

// normalizeTier folds a frontmatter tier value to S, M, or L; anything else
// (including empty and numeric values) returns "", and the SPEC is not checked.
//
// DOCTRINE DIVERGENCE — absent tier. Two clauses say a SPEC without `tier:` is
// treated as Tier L: `.claude/rules/moai/workflow/spec-workflow.md`
// § SPEC Complexity Tier (S/M/L) (the backward-compatibility sentence on the
// optional `tier:` field) and `.claude/rules/moai/development/spec-frontmatter-schema.md`
// § Optional Fields (the `tier` row). This rule deliberately does NOT check an
// absent tier: the card's scope was SPECs whose frontmatter declares a tier.
// Whether to align the code to the doctrine (check absent tier as L) or the
// doctrine to the code is left to a follow-up.
func normalizeTier(v string) string {
	v = strings.ToUpper(strings.Trim(strings.TrimSpace(v), `"'`))
	switch v {
	case "S", "M", "L":
		return v
	}
	return ""
}

// TierArtifactMissingRule reports SPEC directories missing an artifact their
// tier requires.
type TierArtifactMissingRule struct {
	table *tierArtifactTable
}

// Code returns the stable lint finding code.
func (r *TierArtifactMissingRule) Code() string { return "TierArtifactMissing" }

// Check emits at most one finding per SPEC, listing every missing file.
func (r *TierArtifactMissingRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	tier := normalizeTier(doc.Frontmatter.Tier)
	if tier == "" {
		return nil
	}
	st := r.table.stateFor(r.table.rootFor(doc.Path))
	if !st.readable {
		return nil
	}
	dir := filepath.Dir(doc.Path)
	var missing []string
	for _, name := range st.sets[tier] {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return []Finding{{
		File:     doc.Path,
		Line:     1,
		Severity: SeverityWarning,
		Code:     r.Code(),
		Message: fmt.Sprintf("%s: tier %s requires %s; missing: %s (artifact set per spec-workflow.md § SPEC Complexity Tier (S/M/L))",
			filepath.Base(dir), tier, strings.Join(st.sets[tier], " + "), strings.Join(missing, ", ")),
	}}
}

// TierArtifactTableRule is the corpus-level companion: when the rule file is
// present but its Tier table cannot be read, it emits exactly one warning.
type TierArtifactTableRule struct {
	table *tierArtifactTable
}

// Code returns the stable lint finding code.
func (r *TierArtifactTableRule) Code() string { return "TierArtifactSetUnreadable" }

// Check is a no-op; the rule is cross-SPEC (CheckAll).
func (r *TierArtifactTableRule) Check(_ *SPECDoc, _ []*SPECDoc) []Finding { return nil }

// CheckAll emits one warning per project root whose SSOT is present but
// unparseable. A normal run spans one root, so this is one corpus warning; with
// no SPECs it checks the BaseDir-derived root.
func (r *TierArtifactTableRule) CheckAll(docs []*SPECDoc) []Finding {
	var roots []string
	seen := map[string]bool{}
	for _, doc := range docs {
		if root := r.table.rootFor(doc.Path); !seen[root] {
			seen[root] = true
			roots = append(roots, root)
		}
	}
	if len(roots) == 0 {
		roots = []string{r.table.fallbackRoot}
	}
	var findings []Finding
	for _, root := range roots {
		if st := r.table.stateFor(root); st.present && !st.readable {
			findings = append(findings, r.unreadable(root))
		}
	}
	return findings
}

func (r *TierArtifactTableRule) unreadable(root string) Finding {
	return Finding{
		File:     filepath.Join(root, filepath.FromSlash(tierArtifactRuleRelPath)),
		Line:     1,
		Severity: SeverityWarning,
		Code:     r.Code(),
		Message:  "tier artifact set could not be read from spec-workflow.md § SPEC Complexity Tier (S/M/L): expected S, M, and L rows under an \"Artifact set\" column; TierArtifactMissing is inactive until the table parses",
	}
}
