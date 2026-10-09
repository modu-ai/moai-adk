package template

// Binding ledger integrity test of SPEC-ALWAYS-LOADED-BUDGET-001
// (REQ-ALB-015, AC-ALB-019, AC-ALB-020, AC-ALB-021).
//
// The binding ledger (testdata/binding_ledger.json) is a TEST FIXTURE. It is
// never read at runtime and never deployed (REQ-ALB-023): every consumer in
// this file runs under _test.go only.

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// ---------------------------------------------------------------------------
// Ledger fixture types (full shape — shared with always_loaded_budget_test.go)
// ---------------------------------------------------------------------------

type ledgerEntryPoint struct {
	EntryPoint string `json:"entry_point"`
	Delivery   string `json:"delivery"`
}

type ledgerSource struct {
	File      string `json:"file"`
	Heading   string `json:"heading"`
	StartLine int    `json:"start_line"`
}

type bindingLedgerRow struct {
	ID          string              `json:"id"`
	Kind        string              `json:"kind"`
	AnchorKind  string              `json:"anchor_kind"`
	Source      ledgerSource        `json:"source"`
	BeforeText  string              `json:"before_text"`
	AfterText   string              `json:"after_text"`
	Location    string              `json:"location"`
	Processing  string              `json:"processing"`
	Reason      string              `json:"reason"`
	RewriteNote string              `json:"rewrite_note"`
	EntryPoints *[]ledgerEntryPoint `json:"entry_points"`
}

type bindingLedgerScope struct {
	Definition string   `json:"definition"`
	Members    []string `json:"members"`
}

type bindingLedger struct {
	Head struct {
		Spec        string             `json:"spec"`
		Card        string             `json:"card"`
		AnchorSHA   string             `json:"anchor_sha"`
		LedgerScope bindingLedgerScope `json:"ledger_scope"`
	} `json:"head"`
	Rows []bindingLedgerRow `json:"rows"`
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

// cloneLedger deep-copies the parsed ledger for in-memory mutation fixtures.
func cloneLedger(led *bindingLedger) *bindingLedger {
	out := &bindingLedger{Rows: make([]bindingLedgerRow, len(led.Rows))}
	out.Head = led.Head
	for i, r := range led.Rows {
		out.Rows[i] = r
		if r.EntryPoints != nil {
			eps := make([]ledgerEntryPoint, len(*r.EntryPoints))
			copy(eps, *r.EntryPoints)
			out.Rows[i].EntryPoints = &eps
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Anchor content — the anchor-tree bytes the coverage check extracts units
// from ("re-run the unit extractor against the anchor texts"). Primary source
// is `git show <anchor_sha>:<path>`; if git is unavailable the fallback reads
// the current tree copy and says so (valid only while template bytes equal
// the anchor — named, never silent).
// ---------------------------------------------------------------------------

var (
	anchorGitRootOnce sync.Once
	anchorGitRoot     string
	anchorGitRootErr  error
)

func anchorGitWorktreeRoot() (string, error) {
	anchorGitRootOnce.Do(func() {
		out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			anchorGitRootErr = err
			return
		}
		anchorGitRoot = strings.TrimSpace(string(out))
	})
	return anchorGitRoot, anchorGitRootErr
}

// newAnchorContentFn returns a fetcher for anchor-tree file content by
// repo-relative path, memoized per call (mutation fixtures re-run the checks
// many times; the anchor bytes never change within one test). Every fallback
// use is logged so the substitution stays visible.
func newAnchorContentFn(t *testing.T, anchorSHA string) func(repoRel string) ([]byte, bool) {
	t.Helper()
	if anchorSHA == "" {
		t.Fatal("ledger head carries no anchor_sha")
	}
	root, gitErr := anchorGitWorktreeRoot()
	var mu sync.Mutex
	cache := map[string][]byte{}
	return func(repoRel string) ([]byte, bool) {
		mu.Lock()
		if hit, ok := cache[repoRel]; ok {
			mu.Unlock()
			return hit, true
		}
		mu.Unlock()
		var data []byte
		if gitErr == nil {
			out, err := exec.Command("git", "-C", root, "show", anchorSHA+":"+repoRel).Output()
			if err == nil {
				data = out
			} else {
				// git is available and the object lookup failed: the file does
				// not exist at the anchor (exit 128 for a missing path). The
				// current-tree fallback here would fill anchor content with
				// TODAY's bytes and let a later-added binding fragment pass
				// the AC-ALB-021(2) "already in anchor copy" exemption — the
				// absent-at-anchor case must read as absent.
				t.Logf("[anchor-content] %s absent at anchor %s (git show: %v); anchorLines stays empty", repoRel, anchorSHA, err)
				return nil, false
			}
		} else {
			t.Logf("[anchor-content] git unavailable (%v); falling back to the current tree copy for %s", gitErr, repoRel)
		}
		if data == nil {
			read, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(repoRel)))
			if readErr != nil {
				return nil, false
			}
			data = read
		}
		mu.Lock()
		cache[repoRel] = data
		mu.Unlock()
		return data, true
	}
}

// ---------------------------------------------------------------------------
// Unit extractor — spec.md §B 단위 경계, applied to anchor bytes.
//
// Frontmatter (leading --- block) is metadata: skipped, and source.start_line
// counts body lines after it (file line minus frontmatter line count).
// Heading lines close units and are not units. A unit starts on a non-blank
// line and continues until a heading, or until — after a blank line — a line
// arrives that is not a list item, sub-item, table row, or code fence. A
// list/table/fence immediately following belongs even across one blank line
// (the blank stays inside the unit text, matching the ledger's before_text).
// Code fences are opaque: # lines, blanks, and constraint tokens inside a
// fence neither enumerate nor break units.
// ---------------------------------------------------------------------------

type anchorUnit struct {
	StartLine int // body-relative, 1-based (ledger source.start_line semantics)
	Heading   string
	Text      string
}

func isHeadingLine(line string) bool {
	if !strings.HasPrefix(line, "#") {
		return false
	}
	// n counts ADDITIONAL '#' after the first: a bare H1 has n == 0 and is a
	// heading; the ATX bound is 1-6 total ('#'+n <= 6), then space or EOL.
	rest := line[1:]
	n := 0
	for n < len(rest) && rest[n] == '#' {
		n++
	}
	if n > 5 {
		return false
	}
	return n == len(rest) || rest[n] == ' '
}

func isListItemLine(line string) bool {
	t := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") || strings.HasPrefix(t, "+ ") {
		return true
	}
	i := 0 // ordered list: digits then '.' or ')' then space
	for i < len(t) && t[i] >= '0' && t[i] <= '9' {
		i++
	}
	return i > 0 && i < len(t) && (t[i] == '.' || t[i] == ')') && i+1 < len(t) && t[i+1] == ' '
}

func isTableRowLine(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " \t"), "|")
}

func isFenceLine(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " \t"), "```")
}

func extractUnits(content []byte) (units []anchorUnit, frontmatterLines int) {
	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	start := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				start = i + 1
				frontmatterLines = i + 1
				break
			}
		}
	}

	type openUnit struct {
		lines        []string
		startLine    int
		pendingBlank int
	}
	flushBlanks := func(cur *openUnit) {
		for k := 0; k < cur.pendingBlank; k++ {
			cur.lines = append(cur.lines, "")
		}
		cur.pendingBlank = 0
	}
	var cur *openUnit
	heading := ""
	inFence := false

	closeUnit := func() {
		if cur != nil && len(cur.lines) > 0 {
			units = append(units, anchorUnit{
				StartLine: cur.startLine,
				Heading:   heading,
				Text:      strings.Join(cur.lines, "\n"),
			})
		}
		cur = nil
	}

	for i := start; i < len(lines); i++ {
		line := lines[i]
		bodyLine := i + 1 - frontmatterLines

		if isFenceLine(line) {
			if inFence {
				if cur != nil {
					cur.lines = append(cur.lines, line)
				}
				inFence = false
				continue
			}
			if cur == nil {
				cur = &openUnit{startLine: bodyLine}
			} else {
				flushBlanks(cur)
			}
			cur.lines = append(cur.lines, line)
			inFence = true
			continue
		}
		if inFence {
			if cur != nil {
				cur.lines = append(cur.lines, line)
			}
			continue
		}

		trimmed := strings.TrimSpace(line)
		if isHeadingLine(line) {
			closeUnit()
			heading = trimmed
			continue
		}
		if trimmed == "" {
			if cur != nil {
				cur.pendingBlank++
			}
			continue
		}
		continues := isListItemLine(line) || isTableRowLine(line) || line[0] == ' ' || line[0] == '\t'
		if cur == nil {
			cur = &openUnit{startLine: bodyLine}
		} else if cur.pendingBlank > 0 {
			if continues {
				flushBlanks(cur)
			} else {
				closeUnit()
				cur = &openUnit{startLine: bodyLine}
			}
		} else {
			cur.pendingBlank = 0
		}
		cur.lines = append(cur.lines, line)
	}
	closeUnit()
	return units, frontmatterLines
}

// ---------------------------------------------------------------------------
// Check environment — runs the REQ-ALB-015 condition set over a ledger and a
// deployed tree, returning error strings. Mutation fixtures re-run it over
// mutated copies and observe the failures.
// ---------------------------------------------------------------------------

const templatePrefix = "internal/template/templates/"

// deployedRelOfTemplate maps a repo-relative template path to its deployed
// tree-relative path.
func deployedRelOfTemplate(repoRel string) string {
	return strings.TrimPrefix(repoRel, templatePrefix)
}

type ledgerCheckEnv struct {
	t       *testing.T
	root    string
	ledger  *bindingLedger
	members []deployedSurfaceMember

	memberSet map[string]bool

	// anchor fetches anchor-tree content by repo-relative template path.
	anchor func(repoRel string) ([]byte, bool)
	// readFile reads deployed-tree content by root-relative slash path;
	// fixtures override it to mutate one file's content.
	readFile func(relPath string) ([]byte, error)
	// roleCoreContent is the M2 seam: returns the content the role-core
	// builder will produce for a role-gated rule (rule-relative path). At M1
	// the role-core text is still physically inside the deployed rule files,
	// so the seam returns the full deployed file; M2 re-points it at the
	// builder output. The sweep count is logged either way.
	roleCoreContent func(ruleRel string) string

	errors []string

	// normMemo caches whitespace-normalized deployed-file content so the
	// after-text containment check normalizes each file once per env.
	normMemo map[string]string
}

func newLedgerCheckEnv(t *testing.T, root string, led *bindingLedger, members []deployedSurfaceMember, anchor func(string) ([]byte, bool)) *ledgerCheckEnv {
	env := &ledgerCheckEnv{
		t:         t,
		root:      root,
		ledger:    led,
		members:   members,
		memberSet: make(map[string]bool, len(members)),
		anchor:    anchor,
	}
	for _, m := range members {
		env.memberSet[m.Path] = true
	}
	env.readFile = func(relPath string) ([]byte, error) {
		return os.ReadFile(filepath.Join(root, filepath.FromSlash(relPath)))
	}

	// Role-core seam (M2 form): resolve each ledger role-core: rule-relative
	// path against the deployed rules tree and return the BUILDER OUTPUT —
	// the regions the neutral moai:role-core markers enclose in the deployed
	// file (config.ExtractRoleCoreRegions, the extraction the SessionStart
	// hook's role-core builder runs; REQ-ALB-023). An unmarked or unreadable
	// file resolves to "" so the row falls to the after-text failure paths —
	// this seam is the marker-accuracy check: every role-core row's
	// after-text must live inside a marked region.
	resolved := map[string]string{}
	roleSuffixes := map[string]bool{}
	for _, row := range led.Rows {
		if strings.HasPrefix(row.Location, "role-core:") {
			roleSuffixes[strings.TrimPrefix(row.Location, "role-core:")] = true
		}
	}
	_ = filepath.Walk(filepath.Join(root, ".claude", "rules"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil //nolint:nilerr
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil //nolint:nilerr
		}
		slash := filepath.ToSlash(rel)
		for suf := range roleSuffixes {
			if strings.HasSuffix(slash, "/"+suf) {
				resolved[suf] = slash
			}
		}
		return nil
	})
	roleSweepCount := 0
	roleUnmarked := 0
	env.roleCoreContent = func(ruleRel string) string {
		roleSweepCount++
		deployed, ok := resolved[ruleRel]
		if !ok {
			return ""
		}
		data, err := env.readFile(deployed)
		if err != nil {
			return ""
		}
		regions, marked := config.ExtractRoleCoreRegions(string(data))
		if !marked {
			roleUnmarked++
			return ""
		}
		return strings.Join(regions, "\n\n")
	}
	t.Cleanup(func() {
		t.Logf("[role-core seam] sweep count: %d rows checked against builder output (M2 form: marker-enclosed regions of the deployed file); unmarked=%d",
			roleSweepCount, roleUnmarked)
	})
	return env
}

// normWS collapses whitespace runs to single spaces so after-text containment
// tolerates reflow only where the text itself is unchanged word-for-word.
func normWS(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// runLedgerChecks executes every REQ-ALB-015 condition and returns the
// collected error strings (empty = green).
func (env *ledgerCheckEnv) runLedgerChecks() []string {
	env.errors = nil
	env.checkRowCoverage()
	env.checkKindMatchesAnchorKind()
	env.checkEntryPoints()
	env.checkLocationVocabulary()
	env.checkAlwaysRowsInMembers()
	env.checkCompanionRowsPathsScoped()
	env.checkAfterTextPresence()
	env.checkAC021NoBindingOnDemand()
	env.checkAC021FragmentRule()
	return env.errors
}

func (env *ledgerCheckEnv) errf(format string, args ...any) {
	env.errors = append(env.errors, fmt.Sprintf(format, args...))
}

// checkRowCoverage re-runs the unit extractor against the anchor texts and
// proves every extracted unit has exactly one row at the same position and
// text, and no row is orphaned (REQ-ALB-015 row-coverage condition). The
// sweep enumerates the DECLARED ledger scope (Head.LedgerScope.Members),
// never a row-derived file set: a file whose rows were all deleted must
// still be swept, or its obligations silently leave the check.
func (env *ledgerCheckEnv) checkRowCoverage() {
	byFile := map[string][]*bindingLedgerRow{}
	for i := range env.ledger.Rows {
		r := &env.ledger.Rows[i]
		member := templatePathToMember(r.Source.File)
		byFile[member] = append(byFile[member], r)
	}
	unitCount, matched, uncovered, orphans, textMismatch, headingDrift, dupes := 0, 0, 0, 0, 0, 0, 0
	filesExtracted := 0
	inScope := map[string]bool{}
	for _, member := range env.ledger.Head.LedgerScope.Members {
		inScope[member] = true
		rows := byFile[member]
		content, ok := env.anchor(memberTemplatePath(member))
		if !ok {
			env.errf("row coverage: anchor content unavailable for declared member %s", member)
			continue
		}
		units, _ := extractUnits(content)
		filesExtracted++
		unitCount += len(units)
		byStart := map[int]anchorUnit{}
		claimed := map[int]bool{}
		for _, u := range units {
			byStart[u.StartLine] = u
		}
		for _, r := range rows {
			u, ok := byStart[r.Source.StartLine]
			if !ok {
				env.errf("row coverage: row %s (%s:%d) has no extracted anchor unit", r.ID, member, r.Source.StartLine)
				orphans++
				continue
			}
			if claimed[r.Source.StartLine] {
				env.errf("row coverage: row %s duplicates a claim on unit %s:%d — exactly one row per unit", r.ID, member, r.Source.StartLine)
				dupes++
				continue
			}
			claimed[r.Source.StartLine] = true
			if u.Text != r.BeforeText {
				env.errf("row coverage: row %s before_text differs from the anchor unit at %s:%d", r.ID, member, r.Source.StartLine)
				textMismatch++
				continue
			}
			if u.Heading != r.Source.Heading {
				headingDrift++ // attribution metadata; counted, not fatal
			}
			matched++
		}
		for _, u := range units {
			if !claimed[u.StartLine] {
				env.errf("row coverage: uncovered anchor unit at %s:%d in declared member %s (%q...)", member, u.StartLine, member, firstN(u.Text, 60))
				uncovered++
			}
		}
	}
	for member, rows := range byFile {
		if !inScope[member] {
			env.errf("row coverage: %d rows reference %s which is not in the declared ledger scope", len(rows), member)
		}
	}
	env.t.Logf("[row coverage] swept files=%d units=%d rows matched=%d uncovered=%d orphan rows=%d text mismatches=%d heading drift=%d duplicate claims=%d",
		filesExtracted, unitCount, matched, uncovered, orphans, textMismatch, headingDrift, dupes)
	if filesExtracted == 0 || unitCount == 0 {
		env.errf("row coverage swept nothing (files=%d units=%d) — empty sweep is a failure, not a pass", filesExtracted, unitCount)
	}
}

// memberTemplatePath maps a declared ledger-scope member (rules-dir-relative,
// e.g. "core/agent-common-protocol.md") to its full template path — the form
// the anchor-content function takes and rows' Source.File carries.
func memberTemplatePath(member string) string {
	return templatePrefix + ".claude/rules/moai/" + member
}

// templatePathToMember is the inverse of memberTemplatePath; paths outside
// the rules tree come back unchanged (and are reported as out-of-scope rows).
func templatePathToMember(repoRel string) string {
	return strings.TrimPrefix(repoRel, templatePrefix+".claude/rules/moai/")
}

func firstN(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

// checkKindMatchesAnchorKind fails any row whose kind differs from its frozen
// anchor kind (changes only in a §C.2 re-anchor commit).
func (env *ledgerCheckEnv) checkKindMatchesAnchorKind() {
	n := 0
	for i := range env.ledger.Rows {
		r := &env.ledger.Rows[i]
		n++
		if r.Kind != r.AnchorKind {
			env.errf("anchor kind mismatch: row %s kind=%s anchor_kind=%s (kind changes only in a §C.2 re-anchor commit)", r.ID, r.Kind, r.AnchorKind)
		}
	}
	env.t.Logf("[kind check] swept %d rows", n)
}

// checkEntryPoints verifies every binding/normative row carries a present,
// non-empty, well-formed entry_points field; role-core rows may only carry
// REQ-ALB-007/REQ-ALB-024 deliveries (REQ-ALB-025 — a delivery=always entry
// point on a role-core row means an uncovered entry point). The registry /
// marker RESOLUTION of individual REQ-ALB-007/REQ-ALB-024 identifiers is the
// named M2 seam (the registry and the moai:role-rules-required markers do
// not exist before M2); the distinct identifiers are logged here so the seam
// is concrete.
func (env *ledgerCheckEnv) checkEntryPoints() {
	present, missing, empty := 0, 0, 0
	identifiers := map[string]int{}
	roleCoreAlways := 0
	for i := range env.ledger.Rows {
		r := &env.ledger.Rows[i]
		if r.Kind != "binding" && r.Kind != "normative" {
			continue
		}
		switch {
		case r.EntryPoints == nil:
			env.errf("entry_points missing on %s row %s (%s)", r.Kind, r.ID, r.Location)
			missing++
			continue
		case len(*r.EntryPoints) == 0:
			env.errf("entry_points empty on %s row %s (%s)", r.Kind, r.ID, r.Location)
			empty++
			continue
		}
		present++
		isRoleCore := strings.HasPrefix(r.Location, "role-core:")
		for _, ep := range *r.EntryPoints {
			if ep.EntryPoint == "" {
				env.errf("entry_points carries an empty entry_point identifier on row %s", r.ID)
			}
			switch ep.Delivery {
			case "always", "REQ-ALB-007", "REQ-ALB-024":
			default:
				env.errf("entry_points delivery %q outside the vocabulary on row %s", ep.Delivery, r.ID)
			}
			if isRoleCore && ep.Delivery == "always" {
				env.errf("REQ-ALB-025 violation: role-core row %s carries a delivery=always entry point (%s) — role-core requires every entry point covered by REQ-ALB-007/REQ-ALB-024", r.ID, ep.EntryPoint)
				roleCoreAlways++
			}
			identifiers[ep.Delivery+" "+ep.EntryPoint]++
		}
	}
	env.t.Logf("[entry_points] swept binding+normative rows: present=%d missing=%d empty=%d role-core delivery=always violations=%d; distinct identifiers=%d",
		present, missing, empty, roleCoreAlways, len(identifiers))
	if present+missing+empty == 0 {
		env.errf("entry_points check swept 0 binding/normative rows — empty sweep is a failure, not a pass")
	}
	// Identifier resolution (gate finding): checking the delivery TYPE alone
	// let a never-registered role and an undeployed entry-point file pass.
	// Every identifier resolves against its real surface — REQ-ALB-007 names
	// a registry marker, REQ-ALB-024 names a deployed file carrying the
	// moai:role-rules-required marker (the M2 seam this check named, now
	// closed: both landed).
	registry := map[string]bool{}
	for _, m := range config.RoleMarkerRegistry() {
		registry[m.Name] = true
	}
	resolved, unresolved := 0, 0
	for id := range identifiers {
		delivery, ident, found := strings.Cut(id, " ")
		if !found {
			continue // the empty-identifier error already fired per row
		}
		switch delivery {
		case "REQ-ALB-007":
			if registry[ident] {
				resolved++
				continue
			}
			env.errf("entry_point %q (REQ-ALB-007) is not a role-marker registry name", ident)
			unresolved++
		case "REQ-ALB-024":
			// The entry-point files live outside the test deployment (the
			// deployer excludes .claude/agents and .claude/skills), so the
			// marker resolves against the deployment ORIGIN — the embedded
			// template set, the same surface role_entry_points_test.go
			// sweeps.
			efs, ferr := EmbeddedTemplates()
			if ferr != nil {
				env.errf("entry_point %q (REQ-ALB-024): embedded templates unavailable: %v", ident, ferr)
				unresolved++
				continue
			}
			data, rerr := fs.ReadFile(efs, filepath.FromSlash(ident))
			if rerr != nil {
				env.errf("entry_point %q (REQ-ALB-024) is not a shipped template file: %v", ident, rerr)
				unresolved++
				continue
			}
			if !strings.Contains(string(data), config.RoleCoreMarkerRequired) {
				env.errf("entry_point %q (REQ-ALB-024) is shipped but carries no %s marker", ident, config.RoleCoreMarkerRequired)
				unresolved++
				continue
			}
			resolved++
		}
	}
	env.t.Logf("[entry_points] identifier resolution: resolved=%d unresolved=%d", resolved, unresolved)
	if unresolved > 0 {
		env.errf("entry_points carried %d unresolved identifiers — delivery type alone is not REQ-ALB-015 conformance", unresolved)
	}
}

// checkLocationVocabulary verifies every location is one of always:/role-core:/
// companion: with a non-empty path, and companion: appears on rationale rows
// only (AC-ALB-020 g, AC-ALB-021(1) location half).
func (env *ledgerCheckEnv) checkLocationVocabulary() {
	counts := map[string]int{}
	for i := range env.ledger.Rows {
		r := &env.ledger.Rows[i]
		prefix, path, ok := splitLocation(r.Location)
		if !ok {
			env.errf("location %q outside the vocabulary on row %s", r.Location, r.ID)
			continue
		}
		if path == "" {
			env.errf("location %q carries an empty path on row %s", r.Location, r.ID)
		}
		counts[prefix]++
		if prefix == "companion" && r.Kind != "rationale" {
			env.errf("companion: location on %s row %s — companion: is rationale-only (REQ-ALB-012/015)", r.Kind, r.ID)
		}
	}
	env.t.Logf("[location vocabulary] swept %d rows: %v", len(env.ledger.Rows), counts)
}

func splitLocation(loc string) (prefix, path string, ok bool) {
	for _, p := range []string{"always:", "role-core:", "companion:"} {
		if strings.HasPrefix(loc, p) {
			return strings.TrimSuffix(p, ":"), strings.TrimPrefix(loc, p), true
		}
	}
	return "", "", false
}

// checkAlwaysRowsInMembers fails any always: row whose path is absent from
// the deployed-surface member list the budget test derives (REQ-ALB-004/
// REQ-ALB-015; AC-ALB-019 fixture f — a binding row located at a paths:-
// carrying full-body path is not an always member).
func (env *ledgerCheckEnv) checkAlwaysRowsInMembers() {
	checked := 0
	for i := range env.ledger.Rows {
		r := &env.ledger.Rows[i]
		if !strings.HasPrefix(r.Location, "always:") {
			continue
		}
		checked++
		path := strings.TrimPrefix(r.Location, "always:")
		if !env.memberSet[path] {
			env.errf("always: row %s path %s is absent from the derived deployed-surface member list (not an always member)", r.ID, path)
		}
	}
	env.t.Logf("[always membership] swept %d always: rows against %d derived members", checked, len(env.members))
	if checked == 0 {
		env.errf("always membership check swept 0 rows — empty sweep is a failure, not a pass")
	}
}

// checkCompanionRowsPathsScoped verifies each companion: row's path is a
// deployed file carrying a top-level paths: key. M0-baseline accommodation:
// a planned companion that does not exist yet is permitted only while the row
// sits in the origin-hold state (after_text == before_text placeholder); the
// accommodation is counted and logged, never silent. Once M3 lands, rows
// leave that state and the strict check binds.
func (env *ledgerCheckEnv) checkCompanionRowsPathsScoped() {
	existsStrict, missingPending, defects := 0, 0, 0
	for i := range env.ledger.Rows {
		r := &env.ledger.Rows[i]
		if !strings.HasPrefix(r.Location, "companion:") {
			continue
		}
		path := strings.TrimPrefix(r.Location, "companion:")
		data, err := env.readFile(path)
		if err == nil {
			if !frontmatterPathsScoped(data) {
				env.errf("companion: row %s path %s is a deployed file without a top-level paths: key", r.ID, path)
				defects++
				continue
			}
			existsStrict++
		} else {
			if r.Kind == "rationale" && r.AfterText == r.BeforeText {
				missingPending++ // planned companion, origin-hold placeholder state
				continue
			}
			env.errf("companion: row %s path %s is not a deployed file", r.ID, path)
			defects++
		}
	}
	env.t.Logf("[companion paths:] swept companion rows: destination verified=%d planned-but-missing (origin-hold, counted)=%d defects=%d",
		existsStrict, missingPending, defects)
}

// checkAfterTextPresence verifies after-text presence for every row:
// always:/companion: rows against the deployed file at their location path;
// role-core: rows against the M2-seam role-core content. A row whose
// after-text is not yet at its location passes ONLY through the counted
// M0-baseline pending branch (its before_text is still present in its source
// file — the planned relocation/rewrite has not executed yet); anything else
// is a loss and fails.
func (env *ledgerCheckEnv) checkAfterTextPresence() {
	direct, pending, failures := 0, 0, 0
	perClass := map[string]int{}
	for i := range env.ledger.Rows {
		r := &env.ledger.Rows[i]
		prefix, path, ok := splitLocation(r.Location)
		if !ok {
			continue // vocabulary check already failed this row
		}
		perClass[prefix]++
		// An empty after_text matches every strings.Contains probe — a
		// deleted directive with its after_text blanked passed the whole
		// integrity sweep (gate finding). Binding/normative rows must carry
		// the deployed obligation's text.
		if (r.Kind == "binding" || r.Kind == "normative") && strings.TrimSpace(r.AfterText) == "" {
			env.errf("after-text empty on %s row %s — an empty after_text contains-matches everything and hides a deleted obligation", r.Kind, r.ID)
			failures++
			continue
		}
		var target string
		switch prefix {
		case "always", "companion":
			if env.containsDeployedNorm(path, r.AfterText) {
				direct++
				continue
			}
		case "role-core":
			target = env.roleCoreContent(path)
			if target != "" && containsNorm(target, r.AfterText) {
				direct++
				continue
			}
		}
		// M0-baseline pending branch — companion: rows only. A planned
		// rationale relocation passes while its anchor text still lives in
		// the row's source file. always: rows take no pending branch: since
		// M3 the canonical always-surface home of a role-gated rule's general
		// blocks is its stub, while the paths:-scoped full body keeps a
		// delivery copy of the same text — letting source presence rescue an
		// always: row would excuse a lost always-surface obligation
		// (mutation fixture (b) depends on this distinction).
		if prefix == "companion" {
			srcRel := deployedRelOfTemplate(r.Source.File)
			if env.containsDeployedNorm(srcRel, r.BeforeText) {
				pending++
				continue
			}
		}
		env.errf("after-text absent for row %s: after_text not at %s and before_text no longer in source %s (obligation lost)", r.ID, r.Location, deployedRelOfTemplate(r.Source.File))
		failures++
	}
	env.t.Logf("[after-text presence] swept %d rows (always=%d companion=%d role-core=%d): direct=%d pending-M0-baseline=%d failures=%d",
		len(env.ledger.Rows), perClass["always"], perClass["companion"], perClass["role-core"], direct, pending, failures)
	if direct+pending+failures == 0 {
		env.errf("after-text presence swept 0 rows — empty sweep is a failure, not a pass")
	}
}

func containsNorm(haystack, needle string) bool {
	if strings.Contains(haystack, needle) {
		return true
	}
	return strings.Contains(normWS(haystack), normWS(needle))
}

// containsDeployedNorm reports whether needle is contained in the deployed
// file at relPath (exact first, then whitespace-normalized), with a per-path
// normalization memo so each file is normalized once per env.
func (env *ledgerCheckEnv) containsDeployedNorm(relPath, needle string) bool {
	if env.normMemo == nil {
		env.normMemo = map[string]string{}
	}
	norm, ok := env.normMemo[relPath]
	if !ok {
		data, err := env.readFile(relPath)
		if err != nil {
			env.normMemo[relPath] = ""
			return false
		}
		norm = normWS(string(data))
		env.normMemo[relPath] = norm
	}
	if norm == "" {
		return false
	}
	return strings.Contains(norm, normWS(needle))
}

// checkAC021NoBindingOnDemand asserts zero binding/normative rows are located
// companion: or in skill paths (AC-ALB-021(1); role-core rows pointing at the
// two role-gated rules are exempt by definition — they are rule paths).
func (env *ledgerCheckEnv) checkAC021NoBindingOnDemand() {
	examined := 0
	for i := range env.ledger.Rows {
		r := &env.ledger.Rows[i]
		if r.Kind != "binding" && r.Kind != "normative" {
			continue
		}
		examined++
		prefix, path, _ := splitLocation(r.Location)
		if prefix == "companion" {
			env.errf("AC-ALB-021(1): %s row %s located companion: (%s)", r.Kind, r.ID, path)
		}
		if strings.Contains(path, ".claude/skills/") {
			env.errf("AC-ALB-021(1): %s row %s located in a skill path (%s)", r.Kind, r.ID, path)
		}
	}
	env.t.Logf("[AC-ALB-021(1)] swept %d binding/normative rows for companion:/skill locations", examined)
}

// onDemandFile is one deployed on-demand surface file with the set of its
// trimmed lines (current) and of its anchor copy's trimmed lines.
type onDemandFile struct {
	rel         string
	lines       map[string]bool
	anchorLines map[string]bool
}

// trimmedLineSet maps every trimmed non-blank line of content.
func trimmedLineSet(content string) map[string]bool {
	set := make(map[string]bool)
	for _, line := range strings.Split(content, "\n") {
		t := strings.TrimSpace(line)
		if t != "" {
			set[t] = true
		}
	}
	return set
}

// stripRoleCoreRegions removes the marker-enclosed regions of content (the
// lines between moai:role-core-start/end, markers included). The role core
// of a role-gated rule is its SANCTIONED delivery path — the SessionStart
// hook injects exactly those regions (REQ-ALB-007) — so a binding row's
// after-text appearing there is delivery, not a forbidden migration onto an
// on-demand file. The AC-ALB-021(2) fragment sweep therefore never sees
// marker-enclosed content as on-demand material.
func stripRoleCoreRegions(content string) string {
	if !strings.Contains(content, config.RoleCoreMarkerStart) {
		return content
	}
	var out []string
	inRegion := false
	for _, line := range strings.Split(content, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.Contains(t, config.RoleCoreMarkerStart):
			inRegion = true
		case strings.Contains(t, config.RoleCoreMarkerEnd):
			inRegion = false
		case inRegion:
			// marker-enclosed: delivered by injection, not swept
		default:
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// checkAC021FragmentRule asserts each line of >=40 UTF-16 units in a
// binding/normative row's after_text appears as no line of any deployed
// on-demand file — top-level paths:-scoped rules (minus carve-out members)
// and deployed skill files — except where the line already existed in that
// file's anchor copy (AC-ALB-021(2) — 「앵커 트리의 같은 companion 에 이미 있던
// 줄은 제외」). Sweep counts are logged; empty sweeps fail.
func (env *ledgerCheckEnv) checkAC021FragmentRule() {
	var files []onDemandFile
	// paths:-scoped rules, minus the carve-out members (role-gated rules stay
	// always-surface members through the M2/M3 arc and are not on-demand).
	_ = filepath.Walk(filepath.Join(env.root, ".claude", "rules"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil //nolint:nilerr
		}
		rel, relErr := filepath.Rel(env.root, path)
		if relErr != nil {
			return nil //nolint:nilerr
		}
		slash := filepath.ToSlash(rel)
		if env.memberSet[slash] {
			return nil // always-surface member, not on-demand
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil || !frontmatterPathsScoped(data) {
			return nil //nolint:nilerr
		}
		files = append(files, onDemandFile{rel: slash, lines: trimmedLineSet(stripRoleCoreRegions(string(data)))})
		return nil
	})
	// Skill files. The deployer excludes the project .claude/skills/ tree from
	// the deployed project, so sweeping env.root/.claude/skills is void — the
	// skills sweep must walk the deployment ORIGIN (the embedded template
	// source) instead. The catalog installs TWO skill origins — .claude/skills
	// and .agents/skills (the command skills moai-plan/moai-run/moai-sync
	// among them) — so both template-source trees are swept; rules, .claude
	// skill, and .agents skill counts are reported separately so no sweep can
	// go silently empty (M1-inheritance repair: the .agents/skills origin was
	// entirely unswept).
	skillRoot := filepath.Join(repoRootFromTemplatePkg(env.t), "internal", "template", "templates")
	claudeSkillFiles, agentsSkillFiles := 0, 0
	for _, skillOrigin := range []struct {
		dir   string
		count *int
	}{
		{".claude/skills", &claudeSkillFiles},
		{".agents/skills", &agentsSkillFiles},
	} {
		skillsRoot := filepath.Join(skillRoot, filepath.FromSlash(skillOrigin.dir))
		_ = filepath.Walk(skillsRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
				return nil //nolint:nilerr
			}
			rel, relErr := filepath.Rel(repoRootFromTemplatePkg(env.t), path)
			if relErr != nil {
				return nil //nolint:nilerr
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil //nolint:nilerr
			}
			files = append(files, onDemandFile{rel: filepath.ToSlash(rel), lines: trimmedLineSet(string(data))})
			*skillOrigin.count++
			return nil
		})
	}
	skillFiles := claudeSkillFiles + agentsSkillFiles
	for j := range files {
		files[j].anchorLines = map[string]bool{}
		anchorRel := files[j].rel
		if !strings.HasPrefix(anchorRel, templatePrefix) {
			anchorRel = templatePrefix + anchorRel
		}
		if anchorData, ok := env.anchor(anchorRel); ok {
			files[j].anchorLines = trimmedLineSet(string(anchorData))
		}
	}

	linesChecked, hits, exemptAtAnchor := env.sweepFragmentLines(files)
	env.t.Logf("[AC-ALB-021(2)] fragment sweep: binding/normative qualifying lines=%d on-demand rule files=%d .claude skill files=%d .agents skill files=%d violations=%d lines exempt (already in anchor copy)=%d",
		linesChecked, len(files)-skillFiles, claudeSkillFiles, agentsSkillFiles, hits, exemptAtAnchor)
	if linesChecked == 0 || len(files) == 0 || skillFiles == 0 {
		env.errf("AC-ALB-021(2) fragment sweep empty (lines=%d files=%d skill files=%d: .claude=%d .agents=%d) — empty sweep is a failure, not a pass",
			linesChecked, len(files), skillFiles, claudeSkillFiles, agentsSkillFiles)
	}
}

// sweepFragmentLines scans every binding/normative after-text line of 40+
// UTF-16 units against the on-demand file line sets; returns the swept counts.
func (env *ledgerCheckEnv) sweepFragmentLines(files []onDemandFile) (linesChecked, hits, exemptAtAnchor int) {
	for i := range env.ledger.Rows {
		r := &env.ledger.Rows[i]
		if r.Kind != "binding" && r.Kind != "normative" {
			continue
		}
		for _, line := range strings.Split(r.AfterText, "\n") {
			trimmed := strings.TrimSpace(line)
			if utf16CodeUnits(trimmed) < 40 {
				continue
			}
			linesChecked++
			for _, f := range files {
				if !f.lines[trimmed] {
					continue
				}
				if f.anchorLines[trimmed] {
					exemptAtAnchor++ // line already in this file's anchor copy
					continue
				}
				env.errf("AC-ALB-021(2): binding fragment of row %s appears in on-demand file %s: %q", r.ID, f.rel, firstN(trimmed, 60))
				hits++
			}
		}
	}
	return linesChecked, hits, exemptAtAnchor
}

// ---------------------------------------------------------------------------
// The test
// ---------------------------------------------------------------------------

// TestBindingLedgerIntegrity runs the full REQ-ALB-015 condition set against
// the committed ledger fixture and the deployed tree, then observes each
// AC-ALB-020 mutation fixture turning a specific check red. The committed
// ledger is never modified: fixtures mutate in-memory copies or an overridden
// file read.
func TestBindingLedgerIntegrity(t *testing.T) {
	root := deployEmbeddedTemplatesForTest(t)
	led := loadBindingLedger(t)
	members := deriveDeployedAlwaysLoadedMembers(t, root)
	anchor := newAnchorContentFn(t, led.Head.AnchorSHA)

	env := newLedgerCheckEnv(t, root, led, members, anchor)
	if errs := env.runLedgerChecks(); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("ledger integrity: %s", e)
		}
		return
	}

	t.Run("mutation_fixtures", func(t *testing.T) {
		// (a) deleting one token-less continuation line of a binding block
		// must fail the ledger test naming that block ID (AC-ALB-020b).
		// M3 form: the mutation lands on the row's canonical location file
		// (an always: member post-split, not the row's anchor source), and
		// the deleted line comes from the row's after-text — the text the
		// ledger actually holds at that location.
		t.Run("a_binding_continuation_line_delete_names_block", func(t *testing.T) {
			row, line := findBindingRowTokenlessLine(led)
			if row == nil {
				t.Fatal("no multi-line binding row with a token-less line found for fixture (a)")
			}
			locRel, ok := rowLocationRel(row)
			if !ok {
				t.Fatalf("fixture (a): row %s has no file-backed location", row.ID)
			}
			mutated := newLedgerCheckEnv(t, root, led, members, anchor)
			base, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(locRel)))
			if err != nil {
				t.Fatalf("read location for mutation: %v", err)
			}
			content := strings.Replace(string(base), "\n"+line, "", 1)
			mutated.readFile = func(relPath string) ([]byte, error) {
				if relPath == locRel {
					return []byte(content), nil
				}
				return os.ReadFile(filepath.Join(root, filepath.FromSlash(relPath)))
			}
			observeMutationFailure(t, mutated, "fixture (a): deleting a token-less continuation line of binding block", row.ID)
		})

		// (b) deleting the STOPPED_TEAMMATE_VIOLATION normative paragraph in
		// cross-session-messaging must fail (AC-ALB-020c first half). M3
		// form: the paragraph's canonical always-surface home is the
		// cross-session-messaging stub, so the deletion lands there; the
		// paths:-scoped full body keeps a delivery copy, which must NOT
		// excuse the lost always-surface obligation.
		t.Run("b_stopped_teammate_delete_fails", func(t *testing.T) {
			row := findRowByText(led, "STOPPED_TEAMMATE_VIOLATION", "workflow/cross-session-messaging.md")
			if row == nil {
				t.Fatal("STOPPED_TEAMMATE_VIOLATION row not found for fixture (b)")
			}
			locRel, ok := rowLocationRel(row)
			if !ok {
				t.Fatalf("fixture (b): row %s has no file-backed location", row.ID)
			}
			mutated := newLedgerCheckEnv(t, root, led, members, anchor)
			base, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(locRel)))
			if err != nil {
				t.Fatalf("read location for mutation: %v", err)
			}
			content := strings.Replace(string(base), row.AfterText, "", 1)
			mutated.readFile = func(relPath string) ([]byte, error) {
				if relPath == locRel {
					return []byte(content), nil
				}
				return os.ReadFile(filepath.Join(root, filepath.FromSlash(relPath)))
			}
			observeMutationFailure(t, mutated, "fixture (b): deleting the STOPPED_TEAMMATE_VIOLATION normative paragraph", row.ID)
		})

		// (b2) reclassifying it rationale and moving it to a companion must
		// fail on the anchor-kind mismatch (AC-ALB-020c second half).
		t.Run("b2_stopped_teammate_reclassified_fails_on_anchor_kind", func(t *testing.T) {
			row := findRowByText(led, "STOPPED_TEAMMATE_VIOLATION", "workflow/cross-session-messaging.md")
			if row == nil {
				t.Fatal("STOPPED_TEAMMATE_VIOLATION row not found for fixture (b2)")
			}
			mutatedLedger := cloneLedger(led)
			for i := range mutatedLedger.Rows {
				if mutatedLedger.Rows[i].ID == row.ID {
					mutatedLedger.Rows[i].Kind = "rationale"
					mutatedLedger.Rows[i].Location = "companion:.claude/rules/moai/workflow/cross-session-messaging-detail.md"
					mutatedLedger.Rows[i].Processing = "relocated"
				}
			}
			mutated := newLedgerCheckEnv(t, root, mutatedLedger, members, anchor)
			observeMutationFailure(t, mutated, "fixture (b2): reclassify + move to companion must fail on anchor-kind mismatch", row.ID, "anchor kind mismatch")
		})

		// (c) deleting the "Arming a goal does not authorize..." paragraph in
		// goal-directive must fail; moving it to a companion must fail
		// (AC-ALB-020d). M3 form: the deletion lands on the row's canonical
		// location file using its after-text.
		t.Run("c_goal_directive_delete_fails", func(t *testing.T) {
			row := findRowByText(led, "Arming a goal does not authorize", "workflow/goal-directive.md")
			if row == nil {
				t.Fatal("goal-directive arming-authorization row not found for fixture (c)")
			}
			locRel, ok := rowLocationRel(row)
			if !ok {
				t.Fatalf("fixture (c): row %s has no file-backed location", row.ID)
			}
			mutated := newLedgerCheckEnv(t, root, led, members, anchor)
			base, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(locRel)))
			if err != nil {
				t.Fatalf("read location for mutation: %v", err)
			}
			content := strings.Replace(string(base), row.AfterText, "", 1)
			mutated.readFile = func(relPath string) ([]byte, error) {
				if relPath == locRel {
					return []byte(content), nil
				}
				return os.ReadFile(filepath.Join(root, filepath.FromSlash(relPath)))
			}
			observeMutationFailure(t, mutated, "fixture (c): deleting the arming-authorization normative paragraph", row.ID)
		})

		t.Run("c2_goal_directive_to_companion_fails", func(t *testing.T) {
			row := findRowByText(led, "Arming a goal does not authorize", "workflow/goal-directive.md")
			if row == nil {
				t.Fatal("goal-directive arming-authorization row not found for fixture (c2)")
			}
			mutatedLedger := cloneLedger(led)
			for i := range mutatedLedger.Rows {
				if mutatedLedger.Rows[i].ID == row.ID {
					mutatedLedger.Rows[i].Location = "companion:.claude/rules/moai/workflow/goal-directive-detail.md"
					mutatedLedger.Rows[i].Processing = "relocated"
				}
			}
			mutated := newLedgerCheckEnv(t, root, mutatedLedger, members, anchor)
			observeMutationFailure(t, mutated, "fixture (c2): moving the normative paragraph to a companion must fail (companion: is rationale-only)", row.ID)
		})

		// (d) binding/normative row with missing/empty entry_points fails.
		t.Run("d_entry_points_missing_fails", func(t *testing.T) {
			row := firstRowOfKind(led, "binding")
			mutatedLedger := cloneLedger(led)
			for i := range mutatedLedger.Rows {
				if mutatedLedger.Rows[i].ID == row.ID {
					mutatedLedger.Rows[i].EntryPoints = nil
				}
			}
			mutated := newLedgerCheckEnv(t, root, mutatedLedger, members, anchor)
			observeMutationFailure(t, mutated, "fixture (d): missing entry_points must fail", row.ID)
		})

		t.Run("d2_entry_points_empty_fails", func(t *testing.T) {
			row := firstRowOfKind(led, "normative")
			mutatedLedger := cloneLedger(led)
			for i := range mutatedLedger.Rows {
				if mutatedLedger.Rows[i].ID == row.ID {
					empty := []ledgerEntryPoint{}
					mutatedLedger.Rows[i].EntryPoints = &empty
				}
			}
			mutated := newLedgerCheckEnv(t, root, mutatedLedger, members, anchor)
			observeMutationFailure(t, mutated, "fixture (d2): empty entry_points must fail", row.ID)
		})

		// (e) a role-core: row with an entry point delivery=always fails
		// (REQ-ALB-025 violation).
		t.Run("e_role_core_delivery_always_fails", func(t *testing.T) {
			row := firstRoleCoreRow(led)
			if row == nil {
				t.Fatal("no role-core row found for fixture (e)")
			}
			mutatedLedger := cloneLedger(led)
			for i := range mutatedLedger.Rows {
				if mutatedLedger.Rows[i].ID == row.ID {
					eps := append([]ledgerEntryPoint{}, *mutatedLedger.Rows[i].EntryPoints...)
					eps = append(eps, ledgerEntryPoint{EntryPoint: "any-session", Delivery: "always"})
					mutatedLedger.Rows[i].EntryPoints = &eps
				}
			}
			mutated := newLedgerCheckEnv(t, root, mutatedLedger, members, anchor)
			observeMutationFailure(t, mutated, "fixture (e): role-core row with delivery=always entry point must fail (REQ-ALB-025)", row.ID)
		})

		// (f) a binding row located at a paths:-carrying full-body path fails
		// (not an always member — REQ-ALB-015).
		t.Run("f_binding_location_on_paths_file_fails", func(t *testing.T) {
			row := firstAlwaysBindingRow(led)
			mutatedLedger := cloneLedger(led)
			for i := range mutatedLedger.Rows {
				if mutatedLedger.Rows[i].ID == row.ID {
					mutatedLedger.Rows[i].Location = "always:.claude/rules/moai/workflow/factory-dispatch-detail.md"
				}
			}
			mutated := newLedgerCheckEnv(t, root, mutatedLedger, members, anchor)
			observeMutationFailure(t, mutated, "fixture (f): binding row located at a paths:-carrying full-body path must fail (not an always member)", row.ID)
		})

		// (g) a companion: row on a binding row fails.
		t.Run("g_companion_on_binding_fails", func(t *testing.T) {
			row := firstRowOfKind(led, "binding")
			mutatedLedger := cloneLedger(led)
			for i := range mutatedLedger.Rows {
				if mutatedLedger.Rows[i].ID == row.ID {
					mutatedLedger.Rows[i].Location = "companion:.claude/rules/moai/workflow/goal-directive-detail.md"
				}
			}
			mutated := newLedgerCheckEnv(t, root, mutatedLedger, members, anchor)
			observeMutationFailure(t, mutated, "fixture (g): companion: location on a binding row must fail", row.ID)
		})

		// (h) deleting ALL rows of one declared member must not remove that
		// member from the sweep: the coverage check enumerates the DECLARED
		// ledger scope, so the file's units come back uncovered naming the
		// member (gate finding — a row-derived file list let a whole file
		// vanish from the check).
		t.Run("h_member_rows_all_deleted_still_swept", func(t *testing.T) {
			member := led.Head.LedgerScope.Members[0]
			mutatedLedger := cloneLedger(led)
			kept := mutatedLedger.Rows[:0]
			for _, r := range mutatedLedger.Rows {
				if templatePathToMember(r.Source.File) != member {
					kept = append(kept, r)
				}
			}
			mutatedLedger.Rows = kept
			mutated := newLedgerCheckEnv(t, root, mutatedLedger, members, anchor)
			observeMutationFailure(t, mutated, "fixture (h): all rows of a declared member deleted — its units must be reported uncovered", member)
		})

		// (i) two rows claiming the same anchor unit must fail — exactly one
		// row per unit (REQ-ALB-015); the claim map must reject, not
		// overwrite (gate finding — a duplicated row passed with errors=[]).
		t.Run("i_duplicate_unit_claim_fails", func(t *testing.T) {
			row := firstRowOfKind(led, "binding")
			dup := *row
			dup.ID = row.ID + "-duplicate"
			mutatedLedger := cloneLedger(led)
			mutatedLedger.Rows = append(mutatedLedger.Rows, dup)
			mutated := newLedgerCheckEnv(t, root, mutatedLedger, members, anchor)
			observeMutationFailure(t, mutated, "fixture (i): a duplicated unit claim must fail", "duplicates a claim")
		})

		// (j) a binding fragment planted on the skill sweep surface must be
		// caught (AC-ALB-021(2)) — the skills sweep walks the deployment
		// origin (the template source), not the void project .claude/skills
		// path the deployer excludes (gate finding — the skill scan was
		// entirely hollow).
		t.Run("j_fragment_in_skill_surface_caught", func(t *testing.T) {
			row := firstRowOfKind(led, "binding")
			fragment := ""
			for _, line := range strings.Split(row.AfterText, "\n") {
				if utf16CodeUnits(strings.TrimSpace(line)) >= 40 {
					fragment = strings.TrimSpace(line)
					break
				}
			}
			if fragment == "" {
				t.Skipf("row %s has no 40+ unit line; this fixture needs a multi-line binding row", row.ID)
			}
			mutated := newLedgerCheckEnv(t, root, led, members, anchor)
			files := []onDemandFile{{
				rel:   "internal/template/templates/.claude/skills/moai/workflows/planted-fixture.md",
				lines: map[string]bool{fragment: true},
			}}
			_, hits, _ := mutated.sweepFragmentLines(files)
			if hits == 0 {
				t.Fatalf("fixture (j): planted skill-surface fragment of row %s was NOT caught", row.ID)
			}
			t.Logf("fixture (j) observed red: planted skill-surface fragment of row %s caught (hits=%d)", row.ID, hits)
		})

		// (j2) the same fragment planted on the SECOND skill origin the
		// catalog installs (.agents/skills — the command skills) must also be
		// caught (M1-inheritance repair: that origin was entirely unswept and
		// a planted fragment passed with hits=0).
		t.Run("j2_fragment_in_agents_skill_origin_caught", func(t *testing.T) {
			row := firstRowOfKind(led, "binding")
			fragment := ""
			for _, line := range strings.Split(row.AfterText, "\n") {
				if utf16CodeUnits(strings.TrimSpace(line)) >= 40 {
					fragment = strings.TrimSpace(line)
					break
				}
			}
			if fragment == "" {
				t.Skipf("row %s has no 40+ unit line; this fixture needs a multi-line binding row", row.ID)
			}
			mutated := newLedgerCheckEnv(t, root, led, members, anchor)
			files := []onDemandFile{{
				rel:   "internal/template/templates/.agents/skills/moai-plan/planted-fixture.md",
				lines: map[string]bool{fragment: true},
			}}
			_, hits, _ := mutated.sweepFragmentLines(files)
			if hits == 0 {
				t.Fatalf("fixture (j2): planted .agents/skills-origin fragment of row %s was NOT caught", row.ID)
			}
			t.Logf("fixture (j2) observed red: planted .agents/skills-origin fragment of row %s caught (hits=%d)", row.ID, hits)
		})

		// (k) a binding fragment planted in a NEW companion — a file that did
		// not exist at the anchor — must be caught. The anchor fetcher reads
		// absent-at-anchor as ABSENT (empty anchorLines), so the planted line
		// cannot slip through the "already in anchor copy" exemption; before
		// that repair the fetcher fell back to the current-tree copy and the
		// fixture passed with hits=0.
		t.Run("k_fragment_in_anchor_absent_companion_caught", func(t *testing.T) {
			row := firstRowOfKind(led, "binding")
			fragment := ""
			for _, line := range strings.Split(row.AfterText, "\n") {
				if utf16CodeUnits(strings.TrimSpace(line)) >= 40 {
					fragment = strings.TrimSpace(line)
					break
				}
			}
			if fragment == "" {
				t.Skipf("row %s has no 40+ unit line; this fixture needs a multi-line binding row", row.ID)
			}
			if _, ok := anchor("internal/template/templates/.claude/rules/moai/core/agent-common-protocol-detail.md"); ok {
				t.Skip("the companion exists at the anchor — this fixture needs an anchor-absent file")
			}
			const companionRel = ".claude/rules/moai/core/agent-common-protocol-detail.md"
			abs := filepath.Join(root, filepath.FromSlash(companionRel))
			base, err := os.ReadFile(abs)
			if err != nil {
				t.Fatalf("read new companion in deployed tree: %v", err)
			}
			mutatedBody := string(base) + "\n" + fragment + "\n"
			if err := os.WriteFile(abs, []byte(mutatedBody), 0o644); err != nil {
				t.Fatalf("plant fragment: %v", err)
			}
			t.Cleanup(func() { _ = os.WriteFile(abs, base, 0o644) })

			mutated := newLedgerCheckEnv(t, root, led, members, anchor)
			var ruleFiles []onDemandFile
			data, readErr := os.ReadFile(abs)
			if readErr != nil {
				t.Fatalf("re-read planted companion: %v", readErr)
			}
			if frontmatterPathsScoped(data) {
				ruleFiles = append(ruleFiles, onDemandFile{
					rel:   templatePrefix + companionRel,
					lines: trimmedLineSet(stripRoleCoreRegions(string(data))),
				})
			}
			linesChecked, hits, exempt := mutated.sweepFragmentLines(ruleFiles)
			if linesChecked == 0 {
				t.Fatal("fixture (k): swept 0 lines — the fixture asserts nothing")
			}
			if hits == 0 {
				t.Fatalf("fixture (k): planted fragment in the anchor-absent companion was NOT caught (exempt=%d) — the anchor-absent case is leaking current content as anchor content", exempt)
			}
			t.Logf("fixture (k) observed red: planted fragment of row %s caught in %s (hits=%d, exempt-at-anchor=%d)", row.ID, companionRel, hits, exempt)
		})

		// (m) blanking a binding row's after_text must fail: an empty
		// after_text contains-matches every probe, so a deleted directive
		// with its after_text blanked passed the whole sweep (gate finding).
		t.Run("m_empty_after_text_fails", func(t *testing.T) {
			row := firstRowOfKind(led, "binding")
			mutatedLedger := cloneLedger(led)
			for i := range mutatedLedger.Rows {
				if mutatedLedger.Rows[i].ID == row.ID {
					mutatedLedger.Rows[i].AfterText = "   \n  "
				}
			}
			mutated := newLedgerCheckEnv(t, root, mutatedLedger, members, anchor)
			observeMutationFailure(t, mutated, "fixture (m): a blanked binding after_text must fail — empty matches every Contains probe", row.ID, "after-text empty")
		})

		// (n) an entry_point that resolves to nothing must fail: a
		// never-registered REQ-ALB-007 role name passed delivery-type
		// checking alone (gate finding).
		t.Run("n_unresolved_entry_point_fails", func(t *testing.T) {
			row := firstRowOfKind(led, "binding")
			var target *bindingLedgerRow
			mutatedLedger := cloneLedger(led)
			for i := range mutatedLedger.Rows {
				r := &mutatedLedger.Rows[i]
				if r.ID == row.ID && r.EntryPoints != nil && len(*r.EntryPoints) > 0 {
					target = r
					break
				}
			}
			if target == nil {
				t.Fatal("no binding row with entry_points found for fixture (n)")
			}
			eps := *target.EntryPoints
			eps[0].Delivery = "REQ-ALB-007"
			eps[0].EntryPoint = "never-registered-role"
			mutated := newLedgerCheckEnv(t, root, mutatedLedger, members, anchor)
			observeMutationFailure(t, mutated, "fixture (n): a never-registered role name must fail identifier resolution", "never-registered-role", "not a role-marker registry name")
		})
	})
}

// observeMutationFailure runs the checks and requires that every want
// fragment appears in at least one collected error (one want for a simple
// fixture; several when the label names distinct expected aspects, e.g. the
// row ID and the anchor-kind mismatch). The observed red line is logged —
// the completion act: every fixture's failure is seen, not assumed.
func observeMutationFailure(t *testing.T, env *ledgerCheckEnv, label string, wants ...string) {
	t.Helper()
	errs := env.runLedgerChecks()
	if len(errs) == 0 {
		t.Errorf("%s: expected the mutated ledger/tree to FAIL, got zero failures", label)
		return
	}
	joined := strings.Join(errs, "\n")
	for _, want := range wants {
		if !strings.Contains(joined, want) {
			t.Errorf("%s: expected a failure naming %q, got %d failures: %v", label, want, len(errs), errs)
			return
		}
	}
	t.Logf("mutation observed RED (%s): %s", label, firstN(errs[0], 200))
}

// rowLocationRel returns the project-root-relative file a row's obligation
// canonically lives in after the M3 split — the location path for always:/
// companion: rows, the deployed source file for role-core: rows (the marked
// regions the builder reads live there). Returns false for rows whose
// location carries no file path.
func rowLocationRel(r *bindingLedgerRow) (string, bool) {
	prefix, path, ok := splitLocation(r.Location)
	if !ok || path == "" {
		return "", false
	}
	if prefix == "role-core" {
		return deployedRelOfTemplate(r.Source.File), true
	}
	return path, true
}

// findBindingRowTokenlessLine returns the first (by ID order is not
// guaranteed — first match in row order) file-backed binding row whose
// after_text spans multiple lines and contains at least one continuation line
// carrying no constraint token, plus that line. The after-text is the row's
// canonical current content at its location (M3 form).
func findBindingRowTokenlessLine(led *bindingLedger) (*bindingLedgerRow, string) {
	for i := range led.Rows {
		r := &led.Rows[i]
		if r.Kind != "binding" || !strings.Contains(r.AfterText, "\n") {
			continue
		}
		if _, ok := rowLocationRel(r); !ok {
			continue
		}
		for _, line := range strings.Split(r.AfterText, "\n") {
			if line == "" || strings.Contains(line, "[HARD]") || strings.Contains(line, "MUST") || strings.Contains(line, "shall ") {
				continue
			}
			return r, line
		}
	}
	return nil, ""
}

// findRowByText returns the first row whose before_text contains text and
// whose source file ends with fileSuffix.
func findRowByText(led *bindingLedger, text, fileSuffix string) *bindingLedgerRow {
	for i := range led.Rows {
		r := &led.Rows[i]
		if strings.Contains(r.BeforeText, text) && strings.HasSuffix(r.Source.File, fileSuffix) {
			return r
		}
	}
	return nil
}

func firstRowOfKind(led *bindingLedger, kind string) *bindingLedgerRow {
	for i := range led.Rows {
		if led.Rows[i].Kind == kind {
			return &led.Rows[i]
		}
	}
	return nil
}

func firstRoleCoreRow(led *bindingLedger) *bindingLedgerRow {
	for i := range led.Rows {
		if strings.HasPrefix(led.Rows[i].Location, "role-core:") {
			return &led.Rows[i]
		}
	}
	return nil
}

func firstAlwaysBindingRow(led *bindingLedger) *bindingLedgerRow {
	for i := range led.Rows {
		r := &led.Rows[i]
		if r.Kind == "binding" && strings.HasPrefix(r.Location, "always:") {
			return r
		}
	}
	return nil
}
