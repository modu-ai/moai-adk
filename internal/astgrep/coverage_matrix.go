package astgrep

// Coverage-matrix contract primitives for SPEC-ASTGREP-LANG16-001 M2.
//
// The coverage matrix (.moai/specs/SPEC-ASTGREP-LANG16-001/coverage-matrix.md)
// is an authored document: one cell per (security family, parseable language)
// pair. Because the matrix can drift from the shipped ruleset and nothing else
// reads both, this checker is wired into the Go test suite so it runs in CI.
//
// It reports FOUR distinguishable failure classes (REQ-A16-006):
//
//  1. key-set mismatch — the cell key set differs from the Cartesian product
//     of the two axes. A SET comparison, never a count: a substituted cell
//     has the right cardinality and the wrong contents.
//  2. unresolved cell — a cell carrying neither a rule id nor a rationale.
//  3. unevidenced exemption — an EXEMPT cell whose evidence carries neither a
//     citation ("cite:") nor a recorded probe ("probe:").
//  4. dangling rule id — a cell naming a rule absent from the shipped ruleset.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// FailureClass names one distinguishable checker failure class. Classes are
// reported separately so their output distinguishes "a row is missing" from
// "a row was substituted" from "an absence was asserted unchecked" from "a
// named rule does not exist".
type FailureClass string

const (
	ClassKeySetMismatch       FailureClass = "key-set mismatch"
	ClassUnresolvedCell       FailureClass = "unresolved cell"
	ClassUnevidencedExemption FailureClass = "unevidenced exemption"
	ClassDanglingRuleID       FailureClass = "dangling rule id"
)

// Cell states. PENDING is the plan-authorized interim marker: plan.md §B M2
// item 3 seeds only the already-implemented cells and leaves the rest marked
// pending for the successor SPEC. A PENDING cell contributes its key to the
// Cartesian-product comparison (class 1 still sees it) but cannot satisfy
// REQ-A16-004 until filled; the checker therefore surfaces PENDING keys in
// its output summary rather than emitting an UnresolvedCell failure that
// would keep CI red during the entire seed phase.
const (
	StateImplemented = "IMPLEMENTED"
	StateExempt      = "EXEMPT"
	StatePending     = "PENDING"
)

// MatrixAxes lists the two fixed axes (design.md §1.2): eight security
// families crossed with fourteen languages parseable under the pinned
// ast-grep version.
var MatrixAxes = struct {
	Families   []string
	Languages  []string
	ASTGrepVer string
}{
	Families:   []string{"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8"},
	Languages:  []string{"go", "javascript", "python", "typescript", "rust", "java", "kotlin", "csharp", "ruby", "php", "elixir", "cpp", "scala", "swift"},
	ASTGrepVer: "0.40.5",
}

// MatrixCell is one row of the matrix document (design.md §1.3 schema).
type MatrixCell struct {
	Family    string // "F1".."F8"
	Language  string
	State     string // IMPLEMENTED | EXEMPT | PENDING
	RuleID    string // non-empty when State == IMPLEMENTED
	Rationale string
	Evidence  string // rule-test path for IMPLEMENTED; cite:/probe: record for EXEMPT
}

// Key renders the cell's (family, language) identity.
func (c MatrixCell) Key() string { return c.Family + "/" + c.Language }

// CheckerFinding is one reported violation, tagged with its class and the
// offending cell key so each class is individually auditable.
type CheckerFinding struct {
	Class  FailureClass
	Key    string
	Detail string
}

// CheckCoverageMatrix runs the four-class contract over the parsed cells.
// ruleIDs maps every rule id declared by the shipped ruleset (LoadRulesetIDs);
// class 4 resolves IMPLEMENTED cells against it because nothing else reads
// both the matrix and the ruleset.
func CheckCoverageMatrix(cells []MatrixCell, ruleIDs map[string]bool) []CheckerFinding {
	var findings []CheckerFinding

	// Class 1 — key-set mismatch: a SET comparison against the Cartesian
	// product, deliberately never a count. A substituted cell has correct
	// cardinality and wrong contents; this branch still sees it.
	expected := CartesianProduct()
	actual := map[string]int{}
	for _, c := range cells {
		actual[c.Key()]++
	}
	var missing, foreignOrDup []string
	for key := range expected {
		if _, ok := actual[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	for key, n := range actual {
		switch {
		case !expected[key]:
			foreignOrDup = append(foreignOrDup, fmt.Sprintf("%s x%d (not on either axis)", key, n))
		case n > 1:
			foreignOrDup = append(foreignOrDup, fmt.Sprintf("%s x%d", key, n))
		}
	}
	sort.Strings(foreignOrDup)
	if len(missing)+len(foreignOrDup) > 0 {
		findings = append(findings, CheckerFinding{
			Class: ClassKeySetMismatch,
			Key:   "matrix",
			Detail: fmt.Sprintf("cell key set differs from the %d-key Cartesian product: missing=[%s]; duplicated/substituted=[%s]",
				len(expected), strings.Join(missing, ", "), strings.Join(foreignOrDup, ", ")),
		})
	}

	// Classes 2-4 — per-cell validation.
	for _, c := range cells {
		if c.RuleID != "" && !ruleIDs[c.RuleID] {
			findings = append(findings, CheckerFinding{
				Class: ClassDanglingRuleID,
				Key:   c.Key(),
				Detail: fmt.Sprintf("names rule id %q which is absent from the shipped ruleset",
					c.RuleID),
			})
		}
		hasID := c.RuleID != ""
		hasRationale := c.Rationale != "" && c.Rationale != "-"
		switch c.State {
		case StateImplemented:
			if !hasID {
				findings = append(findings, CheckerFinding{Class: ClassUnresolvedCell,
					Key: c.Key(), Detail: "IMPLEMENTED cell carries no rule id"})
			}
			if hasRationale {
				findings = append(findings, CheckerFinding{Class: ClassUnresolvedCell,
					Key: c.Key(), Detail: "carries BOTH a rule id and a rationale; exactly one is allowed"})
			}
		case StateExempt:
			if !hasRationale {
				findings = append(findings, CheckerFinding{Class: ClassUnresolvedCell,
					Key: c.Key(), Detail: "EXEMPT cell carries no rationale"})
			}
			if hasID {
				findings = append(findings, CheckerFinding{Class: ClassUnresolvedCell,
					Key: c.Key(), Detail: "carries BOTH a rule id and a rationale; exactly one is allowed"})
			}
			lower := strings.ToLower(c.Evidence)
			if !strings.Contains(lower, "cite:") && !strings.Contains(lower, "probe:") {
				findings = append(findings, CheckerFinding{Class: ClassUnevidencedExemption,
					Key: c.Key(),
					Detail: fmt.Sprintf("EXEMPT evidence carries neither cite: nor probe: (%q)",
						c.Evidence)})
			}
		case StatePending:
			// Plan-authorized interim marker (plan.md §B M2 item 3): keys stay
			// under the class-1 comparison above, but the cell owes neither a
			// rule id nor an exemption rationale until its fill milestone.
			// Deliberately NOT an UnresolvedCell failure while unfilled.
		default:
			findings = append(findings, CheckerFinding{Class: ClassUnresolvedCell,
				Key: c.Key(),
				Detail: fmt.Sprintf("state %q is none of IMPLEMENTED/EXEMPT/PENDING; carries neither a usable rule id nor a rationale",
					c.State)})
		}
	}

	return findings
}

// CartesianProduct renders all (family/language) keys demanded by REQ-A16-003.
func CartesianProduct() map[string]bool {
	out := make(map[string]bool, len(MatrixAxes.Families)*len(MatrixAxes.Languages))
	for _, fam := range MatrixAxes.Families {
		for _, lang := range MatrixAxes.Languages {
			out[fam+"/"+lang] = true
		}
	}
	return out
}

// ParseCoverageMatrix extracts cell rows from the markdown matrix document.
// Only pipe rows whose first field matches the family axis are treated as
// data; fenced blocks and prose are skipped so recorded probe output cannot
// corrupt extraction.
func ParseCoverageMatrix(doc string) ([]MatrixCell, error) {
	cellKey := regexp.MustCompile(`^F([1-9]|1[0-9])$`)
	lines := strings.Split(doc, "\n")
	var cells []MatrixCell
	inFence := false
	for i, raw := range lines {
		lineNo := i + 1
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		fields := splitPipeRow(trimmed)
		if len(fields) != 5 {
			if head := strings.Fields(first(fields)); len(head) > 0 && cellKey.MatchString(head[0]) {
				return nil, fmt.Errorf("coverage-matrix.md:%d: malformed table row (%d fields, want 5): %s",
					lineNo, len(fields), trimmed)
			}
			continue
		}
		if strings.EqualFold(fields[0], "family") {
			continue // header row
		}
		if !cellKey.MatchString(fields[0]) {
			continue // prose or excluded-record row, not a cell
		}
		// Column 4 is shared ("rule id / rationale"), so its meaning follows
		// the state: an EXEMPT row carries a rationale there, and reading it
		// as a rule id would turn an unevidenced exemption into a dangling id
		// — the wrong failure class for the same defect.
		cell := MatrixCell{
			Family:   fields[0],
			Language: fields[1],
			State:    norm(fields[2]),
			Evidence: norm(fields[4]),
		}
		if cell.State == StateExempt {
			cell.Rationale = norm(fields[3])
		} else {
			cell.RuleID = norm(fields[3])
		}
		cells = append(cells, cell)
	}
	return cells, nil
}

func first(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	return ss[0]
}

// norm normalizes a table field: trimmed case-kept text, with the documented
// placeholder "-" read as an absent value.
func norm(s string) string {
	t := strings.TrimSpace(s)
	if t == "-" {
		return ""
	}
	return t
}

// splitPipeRow splits a markdown table row into trimmed field values.
func splitPipeRow(row string) []string {
	body := strings.TrimSuffix(strings.TrimPrefix(row, "|"), "|")
	parts := strings.Split(body, "|")
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = strings.TrimSpace(p)
	}
	return out
}

// SeededImplementedCells returns the 14 already-implemented cells measured on
// the shipped ruleset: four sec-hardcoded-credential language variants, one
// weak-hash, api-key and jwt-signing key each, two command-injection families
// across four languages, csrf, log injection and template injection, all Go.
func SeededImplementedCells() []MatrixCell {
	const testRoot = "internal/astgrep/testdata/rule-tests/ (sg-test case pair)"
	return []MatrixCell{
		{"F1", "go", StateImplemented, "sec-command-injection-shell", "", testRoot},
		{"F1", "python", StateImplemented, "sec-command-injection-shell", "", testRoot},
		{"F1", "javascript", StateImplemented, "sec-command-injection-exec", "", testRoot},
		{"F1", "typescript", StateImplemented, "sec-command-injection-exec", "", testRoot},
		{"F2", "go", StateImplemented, "sec-hardcoded-credential", "", testRoot},
		{"F2", "python", StateImplemented, "sec-hardcoded-credential", "", testRoot},
		{"F2", "javascript", StateImplemented, "sec-hardcoded-credential", "", testRoot},
		{"F2", "typescript", StateImplemented, "sec-hardcoded-credential", "", testRoot},
		{"F3", "go", StateImplemented, "sec-weak-hash-md5", "", testRoot},
		{"F4", "go", StateImplemented, "sec-hardcoded-api-key", "", testRoot},
		{"F5", "go", StateImplemented, "sec-hardcoded-jwt-signing-key", "", testRoot},
		{"F6", "go", StateImplemented, "sec-csrf-no-token-check", "", testRoot},
		{"F7", "go", StateImplemented, "sec-log-injection-unsanitized", "", testRoot},
		{"F8", "go", StateImplemented, "sec-template-injection-html", "", testRoot},
	}
}

// LoadRulesetIDs collects every ast-grep rule id declared under a ruleset
// directory tree, keyed for membership lookup by the dangling-rule-id class.
// Rule documents use list items whose first entry is `id:`; indentation varies.
func LoadRulesetIDs(rulesetDir string) (map[string]bool, error) {
	idLine := regexp.MustCompile(`(?m)^\s*-?\s*id:\s*(\S+)\s*$`)
	ids := map[string]bool{}
	err := filepath.WalkDir(rulesetDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk %s: %w", path, walkErr)
		}
		if d.IsDir() || (filepath.Ext(path) != ".yml" && filepath.Ext(path) != ".yaml") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}
		for _, m := range idLine.FindAllStringSubmatch(string(data), -1) {
			ids[m[1]] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}
