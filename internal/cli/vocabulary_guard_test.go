package cli

// vocabulary_guard_test.go — the role-vocabulary guard (SPEC-ROLE-NAMING-CODE-001
// REQ-RNC-018, AC-RNC-018).
//
// It scans PRODUCTION .go string literals (non-_test.go files) under the five
// factory/kanban packages and fails when any literal presents a legacy factory
// role: `-f worker`, `-f agent`, a `worker-<n>` / `agent-<n>` label, or `lead`
// used as the leader noun. `lead` is matched on word boundaries in any letter
// case; an occurrence inside a `MOAI_[A-Z0-9_]+` environment-variable token is
// not counted, and `lead` directly preceded by the qualifier `team ` (any case)
// is not counted (REQ-RNC-023).
//
// Every surviving occurrence must sit in the allowlist below. An entry names
// the file and the EXACT allowlisted literal of the site it covers — a
// comparison site against a legacy spelling that exists only to be refused or
// detected, the text of its error message, or a non-role sense of the word.
// The guard fails when a named file no longer contains its literal (a stale
// entry is a failure, not a pass); an edit that only moves the literal to
// another line does NOT fail, because line numbers are recorded but not
// binding.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// guardPackages are the packages the guard scans, relative to this package's
// directory (internal/cli).
var guardPackages = []string{".", "../kanban", "../hook", "../factorymsg", "../web"}

// guardAllowlist is the complete set of permitted legacy-vocabulary literals.
// Each entry: the file it lives in and the exact literal that must still be
// present. Extend only for refuse/detect comparisons, their error text, or a
// documented non-role sense (REQ-RNC-018).
var guardAllowlist = []guardAllowlistEntry{
	// --- internal/kanban: the detection vocabulary -------------------------
	// legacyLeaderSpelling and the legacy factory role constants exist only to
	// be refused or detected (REQ-RNC-009); their values are the frozen
	// legacy spellings.
	{file: "../kanban/role.go", literal: "lead"},
	{file: "../kanban/bootstrap.go", literal: "worker"},
	{file: "../kanban/bootstrap.go", literal: "agent"},
	{file: "../kanban/bootstrap.go", literal: "worker-"},
	{file: "../kanban/bootstrap.go", literal: "agent-"},
	{file: "../kanban/factory_slots.go", literal: "worker-"},
	{file: "../kanban/factory_slots.go", literal: "agent-"},
	{file: "../kanban/record.go", literal: "worker-"},

	// --- internal/cli: refusal messages and detection comparisons ----------
	// factory.go: the leader-name refusal path composes the canonical form by
	// trimming the legacy prefix (`strings.TrimPrefix(name, "lead")`) so the
	// error names leader-<suffix> (REQ-RNC-007).
	{file: "factory.go", literal: "lead"},

	// --- internal/cli: refusal messages and detection comparisons ----------
	// doctor_factory_run.go: the doctor's chain reader classifies a persisted
	// `lead` record so it renders "legacy run: relaunch required" instead of a
	// present leader (REQ-RNC-009/-013, AC-RNC-013).
	{file: "doctor_factory_run.go", literal: "lead"},

	// --- internal/hook: stale-run detection --------------------------------
	// session_stale_run.go: isLegacyRecordRole compares a session record's
	// role against the legacy spellings to emit the stale-run notice instead
	// of adopting the session (REQ-RNC-009/-025).
	{file: "../hook/session_stale_run.go", literal: "lead"},

	// --- internal/factorymsg: legacy peer identity + slot error text -------
	// factory_run_retire.go: the run-retire owner lookup reads a legacy
	// `lead` peer as identity evidence only (REQ-RNC-024), and classifies
	// legacy peers for the same refuse/retire messaging (REQ-RNC-022).
	{file: "../factorymsg/factory_run_retire.go", literal: "lead"},
	// store.go: canonicalSlotName maps a legacy slot onto the canonical name
	// the refusal error must advertise (REQ-RNC-013).
	{file: "../factorymsg/store.go", literal: "lead"},

	// --- internal/web: legacy dashboard rendering --------------------------
	// viewmodel_ops.go: legacyLeaderRole is the persisted pre-rename value,
	// detected so the leader slot renders the relaunch label (REQ-RNC-009,
	// AC-RNC-013).
	{file: "../web/viewmodel_ops.go", literal: "lead"},
}

// guardAllowlistEntry binds one allowlisted literal to the file that must
// contain it.
type guardAllowlistEntry struct {
	file    string
	literal string
}

// guardLegacyPatterns are the forbidden role presentations, per REQ-RNC-018.
var guardLegacyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`-f worker`),
	regexp.MustCompile(`-f agent`),
	regexp.MustCompile(`worker-(\d|n)`),
	regexp.MustCompile(`agent-(\d|n)`),
	regexp.MustCompile(`(?i)\blead\b`),
}

// guardEnvToken matches a MOAI_ environment-variable name token; `lead`
// inside one is part of the FROZEN variable name (REQ-RNC-011), not a role
// noun.
var guardEnvToken = regexp.MustCompile(`MOAI_[A-Z0-9_]+`)

// guardTeamQualified marks `lead` directly preceded by the `team ` qualifier
// (any case) — the Agent Teams homonym, not the factory or kanban leader
// (REQ-RNC-023).
var guardTeamQualified = regexp.MustCompile(`(?i)team $`)

// scanStringForLegacyRoles returns the sub-match report for one string
// literal: every legacy-pattern hit that survives the env-token and team-
// qualifier exclusions. Exposed for the paired controls below.
func scanStringForLegacyRoles(s string) []string {
	var hits []string
	for _, pat := range guardLegacyPatterns {
		for _, loc := range pat.FindAllStringIndex(s, -1) {
			if insideGuardEnvToken(s, loc[0], loc[1]) {
				continue
			}
			if pat.String() == `(?i)\blead\b` && guardTeamQualified.MatchString(s[:loc[0]]) {
				continue
			}
			hits = append(hits, s[loc[0]:loc[1]])
		}
	}
	return hits
}

// insideGuardEnvToken reports whether [start,end) sits wholly inside a
// MOAI_[A-Z0-9_]+ token. Under word-boundary matching `_` is a word
// character, so LEAD inside MOAI_KANBAN_LEAD_ADDR is not a separate word; the
// explicit token scan makes that exclusion mechanical and keeps the paired
// control honest (a token never masks a free-standing lead in the same
// string).
func insideGuardEnvToken(s string, start, end int) bool {
	for _, loc := range guardEnvToken.FindAllStringIndex(s, -1) {
		if start >= loc[0] && end <= loc[1] {
			return true
		}
	}
	return false
}

// collectGuardLiterals parses one production .go file and returns its string
// literal contents: interpreted strings unquoted, raw strings verbatim.
func collectGuardLiterals(t *testing.T, path string) []string {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("guard: reading %s: %v", path, err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("guard: parsing %s: %v", path, err)
	}
	var literals []string
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		if strings.HasPrefix(lit.Value, `"`) {
			if unquoted, err := strconv.Unquote(lit.Value); err == nil {
				literals = append(literals, unquoted)
			}
			return true
		}
		if strings.HasPrefix(lit.Value, "`") {
			literals = append(literals, lit.Value[1:len(lit.Value)-1])
		}
		return true
	})
	return literals
}

func TestProductionStringLiteralsUseLeaderLaneVocabulary(t *testing.T) {
	type violation struct {
		file, literal, hit string
	}
	var violations []violation

	for _, pkg := range guardPackages {
		entries, err := os.ReadDir(pkg)
		if err != nil {
			t.Fatalf("guard: reading package dir %s: %v", pkg, err)
		}
		for _, e := range entries {
			name := e.Name()
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			path := filepath.Join(pkg, name)
			display := path
			for _, lit := range collectGuardLiterals(t, path) {
				for _, hit := range scanStringForLegacyRoles(lit) {
					violations = append(violations, violation{file: display, literal: lit, hit: hit})
				}
			}
		}
	}

	// Partition violations into allowlisted and unexpected. Coverage is
	// EXACT: an entry excuses only the violation whose whole string literal
	// equals the entry's literal — a substring match would let one broad
	// entry mask every future edit in the file (the allowlist-as-loophole
	// failure REQ-RNC-018 exists to prevent).
	var unexpected []violation
covered:
	for _, v := range violations {
		for _, entry := range guardAllowlist {
			if v.file == entry.file && v.literal == entry.literal {
				continue covered
			}
		}
		unexpected = append(unexpected, v)
	}

	// Stale-entry check: every allowlist entry's file must still contain its
	// exact literal. An entry whose site is gone is a failure — the allowlist
	// must shrink when the vocabulary work removes its reason.
	for _, entry := range guardAllowlist {
		data, err := os.ReadFile(entry.file)
		if err != nil {
			t.Fatalf("guard: reading allowlisted file %s: %v", entry.file, err)
		}
		if !strings.Contains(string(data), entry.literal) {
			t.Errorf("stale allowlist entry: %s no longer contains %q — remove the entry", entry.file, entry.literal)
		}
	}

	for _, v := range unexpected {
		t.Errorf("legacy role vocabulary in %s: hit %q inside literal %q (allowlist it with its refuse/detect reason, or rewrite the string)",
			v.file, v.hit, v.literal)
	}
}

// TestVocabularyGuardControls is the paired-control proof for the exclusion
// rules (AC-RNC-018): an env-var token alone passes, a token plus a
// free-standing lead fails on the free-standing occurrence, and `leader`
// never matches the word-boundary lead pattern.
func TestVocabularyGuardControls(t *testing.T) {
	t.Parallel()

	if hits := scanStringForLegacyRoles("MOAI_KANBAN_LEAD_ADDR"); len(hits) != 0 {
		t.Errorf("env token alone produced hits %v, want none", hits)
	}
	if hits := scanStringForLegacyRoles("MOAI_KANBAN_LEAD_ADDR names the lead address"); len(hits) != 1 || hits[0] != "lead" {
		t.Errorf("free-standing lead next to a token produced %v, want exactly one \"lead\" hit", hits)
	}
	if hits := scanStringForLegacyRoles("This session is the leader of the run"); len(hits) != 0 {
		t.Errorf("leader produced hits %v, want none", hits)
	}
	if hits := scanStringForLegacyRoles("The Lead dispatches cards"); len(hits) != 1 {
		t.Errorf("sentence-initial Lead produced %v, want one hit (case-insensitive)", hits)
	}
	if hits := scanStringForLegacyRoles("Ask the team lead to promote the card"); len(hits) != 0 {
		t.Errorf("team-qualified lead produced hits %v, want none", hits)
	}
	if hits := scanStringForLegacyRoles("use -f lane to join"); len(hits) != 0 {
		t.Errorf("canonical vocabulary produced hits %v, want none", hits)
	}
	if hits := scanStringForLegacyRoles("-f lane-<n> joins as lane n"); len(hits) != 0 {
		t.Errorf("canonical lane label produced hits %v, want none", hits)
	}
}
