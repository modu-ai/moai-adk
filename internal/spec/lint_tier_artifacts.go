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

// tierArtifactTable loads the project's Tier table once per Linter and is
// shared by the per-SPEC rule and the corpus-warning rule.
type tierArtifactTable struct {
	root     string
	once     sync.Once
	present  bool
	sets     map[string][]string
	readable bool
}

func (t *tierArtifactTable) load() {
	t.once.Do(func() {
		data, err := os.ReadFile(filepath.Join(t.root, filepath.FromSlash(tierArtifactRuleRelPath)))
		if err != nil {
			return // absent or unreadable file: feature not configured
		}
		t.present = true
		t.sets, t.readable = parseTierArtifactSets(string(data))
	})
}

// lintProjectRoot maps the Linter BaseDir to the project root. The CLI passes
// <root>/.moai/specs when it exists and <root> otherwise.
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
// (including empty and numeric values) returns "".
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
	r.table.load()
	if !r.table.readable {
		return nil
	}
	tier := normalizeTier(doc.Frontmatter.Tier)
	if tier == "" {
		return nil
	}
	dir := filepath.Dir(doc.Path)
	var missing []string
	for _, name := range r.table.sets[tier] {
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
			filepath.Base(dir), tier, strings.Join(r.table.sets[tier], " + "), strings.Join(missing, ", ")),
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

// CheckAll emits one warning when the SSOT is present but unparseable.
func (r *TierArtifactTableRule) CheckAll(_ []*SPECDoc) []Finding {
	r.table.load()
	if !r.table.present || r.table.readable {
		return nil
	}
	return []Finding{{
		File:     filepath.Join(r.table.root, filepath.FromSlash(tierArtifactRuleRelPath)),
		Line:     1,
		Severity: SeverityWarning,
		Code:     r.Code(),
		Message:  "tier artifact set could not be read from spec-workflow.md § SPEC Complexity Tier (S/M/L): expected S, M, and L rows under an \"Artifact set\" column; TierArtifactMissing is inactive until the table parses",
	}}
}
