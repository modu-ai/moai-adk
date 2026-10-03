package cli

// factory_skill_assertions_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M10, AC-021
// (card t1399): the factory skill documents only what the launcher observably
// does. The test parses the skill file itself — its entry-form table, its
// session-record path and field table, its block-cap value — and replays every
// row against the same seams the launcher acceptance tests drive:
//
//   - cc / glm rows go through parseLauncherEntry (the entry parse runCC and
//     runGLM call first; AC-001..AC-009 and TestLaneEntryComposesWithLaneOptions
//     use it);
//   - codex rows go through codexFactoryEntryClassify (the head classifier
//     runCodex calls first; TestCodexFactoryFlagRefusals reaches it);
//   - the retired rows go through runCG and the root command table.
//
// The same assertions run over the local skill and its template twin, the way
// the neighbouring document-versus-behavior test (docSurfaces) covers both
// surfaces. A table of mutants proves the checker can fail: each mutated skill
// text must earn at least one problem, so a vacuous checker cannot pass.
//
// No session is launched and no real home state is touched: every parse runs
// against a temp project root.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

const (
	factorySkillLocalRel    = ".claude/skills/moai/workflows/factory.md"
	factorySkillTemplateRel = "internal/template/templates/" + factorySkillLocalRel
)

// factorySkillDependents are the four skill files whose framing sentences once
// named a "factory contract" or a "factory chain" (design.md §6, A7).
var factorySkillDependents = []string{
	".claude/skills/moai/workflows/moai.md",
	".claude/skills/moai/workflows/run.md",
	".claude/skills/moai/workflows/run/mode-orchestration.md",
	".claude/skills/moai/workflows/sync/quality-gates-quality.md",
}

// factorySkillForbidden are the assertions the skill must not make: each
// pattern names something the tree does not hold.
var factorySkillForbidden = []struct {
	pattern *regexp.Regexp
	reason  string
}{
	{regexp.MustCompile(`factory_chain`), "names a goal preset that exists in no code or configuration"},
	{regexp.MustCompile(`(?i)goal preset`), "names a goal preset; no goal preset drives the factory"},
	{regexp.MustCompile(`\.moai/state/factory`), "places the session record under a directory RecordPath does not resolve"},
	{regexp.MustCompile(`FACTORY_MODE_UNSUPPORTED_BACKEND`), "names the moai cg sentinel; moai cg is retired and refuses with its retirement error"},
	{regexp.MustCompile(`(?i)chain head`), "names a plan-phase chain head no code implements"},
	{regexp.MustCompile(`(?i)SPEC identifier (may|can) follow|targeted SPEC|targets that SPEC`), "presents -f as taking a SPEC argument"},
	{regexp.MustCompile(`(?i)kanban|moai (cc|glm) -k\b|codex -f\b`), "names a removed entry form"},
}

// factorySkillDependentForbidden is what the four dependents may no longer say.
var factorySkillDependentForbidden = regexp.MustCompile("(?i)factory`?\\s+(contract|chain)")

var (
	entryRowPattern   = regexp.MustCompile("(?m)^\\|\\s*(accepted|refused|retired)\\s*\\|\\s*`([^`]+)`\\s*\\|\\s*([a-z-]*)\\s*\\|")
	recordRowPattern  = regexp.MustCompile("(?m)^\\|\\s*`([a-z_]+)`\\s*\\|\\s*([^|]+?)\\s*\\|")
	backtickSpan      = regexp.MustCompile("`([^`]+)`")
	launcherSpan      = regexp.MustCompile(`^moai (cc|glm|codex) -`)
	blockCapSpan      = regexp.MustCompile("`" + config.EnvClaudeCodeStopHookBlockCap + `=(\d+)` + "`")
	recordSectionHead = regexp.MustCompile(`(?m)^## Session record\s*$`)
	nextSectionHead   = regexp.MustCompile("(?m)^## ")
)

// entryRow is one parsed row of the skill's entry-form table.
type entryRow struct {
	verdict  string // accepted | refused | retired
	command  string // the full command text, e.g. "moai cc -f"
	role     string // leader | lane for accepted rows, empty otherwise
	launcher string // cc | glm | codex | cg | gpt
	args     []string
}

func parseEntryRows(text string) []entryRow {
	var rows []entryRow
	for _, m := range entryRowPattern.FindAllStringSubmatch(text, -1) {
		fields := strings.Fields(m[2])
		row := entryRow{verdict: m[1], command: m[2], role: m[3]}
		if len(fields) >= 2 && fields[0] == "moai" {
			row.launcher = fields[1]
			row.args = fields[2:]
		}
		rows = append(rows, row)
	}
	return rows
}

// entryShape classifies a row into the coverage key the skill table must
// carry at least once: which entry token, with or without a value.
func entryShape(row entryRow) string {
	if row.verdict == "retired" {
		return "retired:" + row.launcher
	}
	if len(row.args) == 0 {
		return ""
	}
	kind := ""
	switch row.args[0] {
	case "-f", "--factory":
		kind = "f"
	case "-l", "--lane":
		kind = "l"
	default:
		return ""
	}
	if len(row.args) >= 2 && !strings.HasPrefix(row.args[1], "-") {
		kind += "-arg"
	}
	return row.verdict + ":" + row.launcher + ":" + kind
}

// factorySkillRequiredShapes is the minimum the table must cover so that no
// launcher, entry token, or accepted/refused polarity goes unreplayed.
var factorySkillRequiredShapes = []string{
	"accepted:cc:f", "accepted:glm:f",
	"accepted:cc:l", "accepted:glm:l", "accepted:codex:l",
	"refused:cc:f-arg", "refused:glm:f-arg",
	"refused:cc:l-arg", "refused:glm:l-arg", "refused:codex:l-arg",
	"refused:codex:f",
	"retired:cg", "retired:gpt",
}

// checkEntryRow replays one row against the launcher seam and returns the
// complaint, or "" when the document and the launcher agree.
func checkEntryRow(row entryRow) string {
	switch row.launcher {
	case "cc", "glm":
		entry, err := parseLauncherEntry(row.args)
		switch row.verdict {
		case "accepted":
			if err != nil {
				return "documented as accepted but the launcher refuses it: " + err.Error()
			}
			if !entry.FactoryEnabled {
				return "documented as accepted but selects no factory entry"
			}
			switch row.role {
			case "leader":
				if entry.FactoryAutoNumber || entry.FactoryLanes != config.DefaultFactoryLeaderLanes {
					return "documented as the leader entry but the parse is not the leader shape"
				}
			case "lane":
				if !entry.FactoryAutoNumber {
					return "documented as the lane entry but the parse takes no automatic lane number"
				}
			default:
				return "accepted row names no role (leader or lane)"
			}
		case "refused":
			if err == nil {
				return "documented as refused but the launcher accepts it"
			}
		default:
			return "a " + row.verdict + " row cannot name a " + row.launcher + " launcher"
		}
	case "codex":
		class, diag := codexFactoryEntryClassify(row.args)
		switch row.verdict {
		case "accepted":
			if class != codexFactoryEntryLane || diag != "" || row.role != "lane" {
				return "documented as a codex lane entry but the head classifier disagrees (class=" + strconv.Itoa(int(class)) + " diag=" + strconv.Quote(diag) + ")"
			}
		case "refused":
			if class != codexFactoryEntryOther || diag == "" {
				return "documented as refused but the codex head classifier does not refuse it"
			}
		default:
			return "a " + row.verdict + " row cannot name the codex launcher"
		}
	case "cg":
		if row.verdict != "retired" {
			return "moai cg is retired, not " + row.verdict
		}
		if err := runCG(cgCmd, row.args); !errors.Is(err, errCGRetired) {
			return "documented as retired but runCG returns " + strconv.Quote(errString(err))
		}
	case "gpt":
		if row.verdict != "retired" {
			return "moai gpt does not exist, not " + row.verdict
		}
		for _, c := range rootCmd.Commands() {
			if c.Name() == "gpt" || c.HasAlias("gpt") {
				return "documented as nonexistent but the root command table registers gpt"
			}
		}
	default:
		return "unknown launcher " + strconv.Quote(row.launcher)
	}
	return ""
}

func errString(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}

// recordJSONKeys returns the JSON keys of factory.Record, sorted.
func recordJSONKeys() []string {
	typ := reflect.TypeOf(factory.Record{})
	keys := make([]string, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		tag := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
		if tag != "" && tag != "-" {
			keys = append(keys, tag)
		}
	}
	sort.Strings(keys)
	return keys
}

// factorySkillUnwritten are the record fields no code in the repository
// writes; the skill must say so for exactly these and no others.
var factorySkillUnwritten = []string{"deepscan_dir", "verify_reentries", "verify_rung"}

// recordNamedOutside returns, per unwritten field, the non-test Go files other
// than record.go that name the field's Go identifier or JSON key. The scan is
// the evidence behind "written by no code in this repository".
func recordNamedOutside(t *testing.T, root string) map[string][]string {
	t.Helper()
	idents := map[string][]string{
		"deepscan_dir":     {"DeepScanDir", "deepscan_dir"},
		"verify_rung":      {"VerifyRung", "verify_rung"},
		"verify_reentries": {"VerifyReentries", "verify_reentries"},
	}
	named := map[string][]string{}
	for _, top := range []string{"internal", "cmd"} {
		_ = filepath.WalkDir(filepath.Join(root, top), func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			if filepath.ToSlash(path) == filepath.ToSlash(filepath.Join(root, "internal/factory/record.go")) {
				return nil
			}
			raw, rerr := os.ReadFile(path)
			if rerr != nil {
				return nil
			}
			for field, names := range idents {
				for _, name := range names {
					if strings.Contains(string(raw), name) {
						rel, _ := filepath.Rel(root, path)
						named[field] = append(named[field], filepath.ToSlash(rel))
						break
					}
				}
			}
			return nil
		})
	}
	return named
}

// recordWriters returns the non-test Go files that build and persist a factory
// session record (the evidence behind the skill's "session-start hook").
func recordWriters(t *testing.T, root string) []string {
	t.Helper()
	var writers []string
	for _, top := range []string{"internal", "cmd"} {
		_ = filepath.WalkDir(filepath.Join(root, top), func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			raw, rerr := os.ReadFile(path)
			if rerr == nil && strings.Contains(string(raw), "factory.WriteBestEffort(") {
				rel, _ := filepath.Rel(root, path)
				writers = append(writers, filepath.ToSlash(rel))
			}
			return nil
		})
	}
	sort.Strings(writers)
	return writers
}

// factorySkillProblems checks one skill text against the tree and returns
// every disagreement. An empty slice is a pass.
func factorySkillProblems(t *testing.T, root, text string, named map[string][]string, writers []string) []string {
	t.Helper()
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	// 1. The entry-form table: every row replays against the launcher seam.
	rows := parseEntryRows(text)
	if len(rows) == 0 {
		add("the skill carries no machine-readable entry-form rows (| accepted/refused/retired | `moai ...` | role |)")
	}
	commands := map[string]bool{}
	shapes := map[string]bool{}
	for _, row := range rows {
		commands[row.command] = true
		shapes[entryShape(row)] = true
		if msg := checkEntryRow(row); msg != "" {
			add("entry row %q: %s", row.command, msg)
		}
	}
	for _, want := range factorySkillRequiredShapes {
		if !shapes[want] {
			add("entry table has no %s row", want)
		}
	}
	for _, span := range backtickSpan.FindAllStringSubmatch(text, -1) {
		if launcherSpan.MatchString(span[1]) && !commands[span[1]] {
			add("prose presents the entry form %q that no table row replays", span[1])
		}
	}

	// 2. The session record: the path equals the resolved RecordPath, the
	// field set equals the Record JSON keys, and the writer column holds.
	wantPath, err := filepath.Rel(root, factory.RecordPath(root, "<session-id>"))
	if err != nil {
		t.Fatalf("relative record path: %v", err)
	}
	wantPath = filepath.ToSlash(wantPath)
	pathSpans := 0
	for _, span := range backtickSpan.FindAllStringSubmatch(text, -1) {
		if strings.Contains(span[1], "<session-id>.json") {
			pathSpans++
			if span[1] != wantPath {
				add("record path %q does not equal the resolved RecordPath %q", span[1], wantPath)
			}
		}
	}
	if pathSpans == 0 {
		add("the skill states no session-record path (want `%s`)", wantPath)
	}
	problems = append(problems, recordTableProblems(text, named, writers)...)

	// 3. The block cap: the skill's value is the launcher's value.
	if m := blockCapSpan.FindStringSubmatch(text); m == nil {
		add("the skill states no `%s=<n>` block-cap value", config.EnvClaudeCodeStopHookBlockCap)
	} else if got, _ := strconv.Atoi(m[1]); got != DefaultRaisedStopHookBlockCap {
		add("block cap %d does not equal DefaultRaisedStopHookBlockCap %d", got, DefaultRaisedStopHookBlockCap)
	}

	// 4. What the skill must not assert.
	for _, f := range factorySkillForbidden {
		if loc := f.pattern.FindString(text); loc != "" {
			add("forbidden text %q: %s", loc, f.reason)
		}
	}
	if loc := removedFormPattern.FindString(text); loc != "" {
		add("names the removed entry form %q", loc)
	}
	return problems
}

// recordTableProblems checks the record section's field table.
func recordTableProblems(text string, named map[string][]string, writers []string) []string {
	var problems []string
	loc := recordSectionHead.FindStringIndex(text)
	if loc == nil {
		return []string{"the skill has no '## Session record' section"}
	}
	section := text[loc[1]:]
	if next := nextSectionHead.FindStringIndex(section); next != nil {
		section = section[:next[0]]
	}
	var fields, unwritten []string
	for _, m := range recordRowPattern.FindAllStringSubmatch(section, -1) {
		fields = append(fields, m[1])
		switch {
		case strings.HasPrefix(m[2], "no code"):
			unwritten = append(unwritten, m[1])
		case strings.HasPrefix(m[2], "session-start hook"):
		default:
			problems = append(problems, fmt.Sprintf("record field %q names writer %q, want 'session-start hook' or 'no code in this repository'", m[1], m[2]))
		}
	}
	sort.Strings(fields)
	sort.Strings(unwritten)
	if want := recordJSONKeys(); !reflect.DeepEqual(fields, want) {
		problems = append(problems, fmt.Sprintf("record fields %v do not equal the Record JSON keys %v", fields, want))
	}
	if !reflect.DeepEqual(unwritten, factorySkillUnwritten) {
		problems = append(problems, fmt.Sprintf("fields documented as written by no code are %v, want %v", unwritten, factorySkillUnwritten))
	}
	for _, field := range factorySkillUnwritten {
		if files := named[field]; len(files) != 0 {
			problems = append(problems, fmt.Sprintf("%s is named outside record.go (%v); re-examine the 'no code writes it' claim", field, files))
		}
	}
	if want := []string{"internal/hook/session_start_record.go"}; !reflect.DeepEqual(writers, want) {
		problems = append(problems, fmt.Sprintf("the session-start hook is not the sole record writer: %v", writers))
	}
	return problems
}

// factorySkillMutants are mutated skill texts: each must earn a problem, which
// proves the checker is not vacuous. A mutation that changes nothing fails the
// harness itself.
var factorySkillMutants = []struct {
	name   string
	mutate func(string) string
	want   string // a substring one problem must carry
}{
	{
		name: "-f documented as taking a SPEC argument",
		mutate: func(s string) string {
			return strings.Replace(s, "| refused | `moai cc -f <spec-id>`", "| accepted | `moai cc -f <spec-id>`", 1)
		},
		want: "documented as accepted but the launcher refuses it",
	},
	{
		name: "-l documented as taking a lane label",
		mutate: func(s string) string {
			return strings.Replace(s, "| refused | `moai glm --lane lane-2`", "| accepted | `moai glm --lane lane-2`", 1)
		},
		want: "documented as accepted but the launcher refuses it",
	},
	{
		name: "a lane entry documented as refused",
		mutate: func(s string) string {
			return strings.Replace(s, "| accepted | `moai cc -l`", "| refused | `moai cc -l`", 1)
		},
		want: "documented as refused but the launcher accepts it",
	},
	{
		name: "a codex leader entry documented as accepted",
		mutate: func(s string) string {
			return strings.Replace(s, "| refused | `moai codex --factory`", "| accepted | `moai codex --factory`", 1)
		},
		want: "documented as a codex lane entry but the head classifier disagrees",
	},
	{
		name: "the old record path",
		mutate: func(s string) string {
			return strings.ReplaceAll(s, ".moai/state/todo/<session-id>.json", ".moai/state/factory/<session-id>.json")
		},
		want: "does not equal the resolved RecordPath",
	},
	{
		name: "a verify field documented as written by the hook",
		mutate: func(s string) string {
			return strings.Replace(s, "| `verify_rung` | no code", "| `verify_rung` | session-start hook", 1)
		},
		want: "fields documented as written by no code",
	},
	{
		name:   "a goal preset named",
		mutate: func(s string) string { return s + "\nThe chain runs under the factory_chain preset.\n" },
		want:   "names a goal preset",
	},
	{
		name: "a wrong block-cap value",
		mutate: func(s string) string {
			return strings.ReplaceAll(s, config.EnvClaudeCodeStopHookBlockCap+"=200", config.EnvClaudeCodeStopHookBlockCap+"=8")
		},
		want: "does not equal DefaultRaisedStopHookBlockCap",
	},
	{
		name:   "an unlisted entry form in prose",
		mutate: func(s string) string { return s + "\nRun `moai glm -f <spec-id>` to target one SPEC.\n" },
		want:   "no table row replays",
	},
}

// TestFactorySkillAssertionsMatchBehavior is AC-021 (release-blocking).
func TestFactorySkillAssertionsMatchBehavior(t *testing.T) {
	root := repoRootForTest(t)
	// Every parse below reads the lane registry under the project root the
	// process resolves; pin it to a temp directory so no real state is read.
	projectDir := t.TempDir()
	t.Setenv(config.EnvClaudeProjectDir, projectDir)
	t.Setenv("MOAI_HOME", t.TempDir())
	netScrubLaneEnv(t)

	named := recordNamedOutside(t, root)
	writers := recordWriters(t, root)

	read := func(rel string) string {
		raw, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		return string(raw)
	}

	var rowSets [][]entryRow
	for _, rel := range []string{factorySkillLocalRel, factorySkillTemplateRel} {
		text := read(rel)
		t.Run("skill/"+rel, func(t *testing.T) {
			for _, p := range factorySkillProblems(t, projectDir, text, named, writers) {
				t.Errorf("%s: %s", rel, p)
			}
		})
		rowSets = append(rowSets, parseEntryRows(text))
	}
	t.Run("twin rows agree", func(t *testing.T) {
		if !reflect.DeepEqual(rowSets[0], rowSets[1]) {
			t.Errorf("the local and template entry-form rows differ:\nlocal    %v\ntemplate %v", rowSets[0], rowSets[1])
		}
	})

	t.Run("dependents", func(t *testing.T) {
		for _, tree := range []string{"", "internal/template/templates/"} {
			for _, rel := range factorySkillDependents {
				text := read(tree + rel)
				if loc := factorySkillDependentForbidden.FindString(text); loc != "" {
					t.Errorf("%s names %q; the verify exit gate and the dedup gate are specified in their own sections, no launcher flag enters them", tree+rel, loc)
				}
				if strings.HasSuffix(rel, "/moai.md") {
					if loc := regexp.MustCompile(`(?i)chain head`).FindString(text); loc != "" {
						t.Errorf("%s names %q; no code implements a plan-phase chain head", tree+rel, loc)
					}
				}
			}
		}
	})

	t.Run("block cap is raised for leader and lane sessions", func(t *testing.T) {
		for _, workers := range []string{"1", "0"} { // a leader's lane count; a lane's incremental 0
			t.Setenv(config.EnvMoaiFactoryWorkers, workers)
			got := injectStopHookBlockCapForGoal(context.Background(), nil, "", "")
			want := config.EnvClaudeCodeStopHookBlockCap + "=" + strconv.Itoa(DefaultRaisedStopHookBlockCap)
			if len(got) != 1 || got[0] != want {
				t.Errorf("workers=%q: block-cap env = %v, want [%s]", workers, got, want)
			}
		}
	})

	t.Run("mutants", func(t *testing.T) {
		base := read(factorySkillLocalRel)
		if p := factorySkillProblems(t, projectDir, base, named, writers); len(p) != 0 {
			t.Skipf("the local skill has problems (%d); mutants are meaningful only against a passing text", len(p))
		}
		for _, m := range factorySkillMutants {
			t.Run(m.name, func(t *testing.T) {
				mutated := m.mutate(base)
				if mutated == base {
					t.Fatalf("mutant %q changed nothing; the harness cannot prove the checker", m.name)
				}
				problems := factorySkillProblems(t, projectDir, mutated, named, writers)
				for _, p := range problems {
					if strings.Contains(p, m.want) {
						return
					}
				}
				t.Errorf("mutant %q was not rejected with %q; problems: %v", m.name, m.want, problems)
			})
		}
	})
}
