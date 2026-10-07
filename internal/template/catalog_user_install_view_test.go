// catalog_user_install_view_test.go — M0 drift guards for the user-install
// catalog view (SPEC-USER-ASSET-INSTALL-001).
//
// The catalog is the single membership SSOT for the user-folder install
// (plan.md §B.2: no second name list in code — the publishedSkillNames
// hardening pattern is why drift guards exist). These guards DERIVE the L0
// closure and the install-coverage matrix from the template sources — the L0
// agent bodies' frontmatter skills unions and Skill() invoke sites, the
// dispatcher's plan/run/sync routing rows, the published command skills'
// dispatcher references, and the default-chain workflow documents' loading
// instructions — and pin the catalog's L0 view to the DERIVED set. The
// derived set, never a hand list, is the guard's source of truth (class
// repair R-c: THE MATRIX IS THIS DRIFT GUARD'S SOURCE OF TRUTH).
//
// Sweep scope (design §2.3): the L0 agent bodies + the default plan/run/sync
// chain (the dispatcher, the three published command skills, and the
// workflows plan/run/sync documents incl. their subdirectories). Conditional
// per-mission injections on that chain are the install-coverage matrix's
// BUNDLE ROWS, each carrying the `moai bundle add <bundle>` remediation.
package template

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// l0CoreAgents is the resolved D-Q1 answer (decision-index.md, 2026-10-05):
// the five core agents — the plan→run→sync chain plus its two auditors.
// Factory is separately in L0 (premise P2) and is not counted in the five.
var l0CoreAgents = []string{
	"manager-spec", "manager-develop", "manager-docs", "plan-auditor", "sync-auditor",
}

// l0ClassifiedOut is the EXPLICIT classification-out list under the
// DEFAULT-FLOW REACHABILITY criterion (final-class item 7 + JD-4: exclusions
// are explicit names, NEVER a moai-ref-*/moai-domain-* prefix — three
// moai-ref-* skills are IN the closure).
var l0ClassifiedOut = map[string]string{
	"moai-domain-html-report": "mission-type injection (manager-docs:224 HTML-render missions only); no default plan/run/sync path renders HTML",
}

// l0FactorySkills ride L0 per operator decision D3 (premise P2): the factory
// entry's skills. The factory entry carries manager-lead as its own declared
// agent dependency (in-round extension E1 — factory-dispatch.md:104 resident
// deputy), NOT a sixth core agent.
var l0FactorySkills = []string{"moai-factory-foreman", "moai-lane-watchdog"}

// l0CommandSkills ride L0 per D-Q4: the published command skills naming the
// plan/run/sync surface.
var l0CommandSkills = []string{"moai-plan", "moai-run", "moai-sync"}

// nonMatrixRoles are roles whose membership is not a default-chain catalog
// concern: the Anthropic built-in, the per-spawn generic, and the
// harness-generated builder (its rows are the non-default harness surface).
var nonMatrixRoles = map[string]bool{
	"Explore": true, "general-purpose": true, "builder-harness": true,
}

var (
	moaiSkillToken  = regexp.MustCompile(`\b(moai-[a-z0-9][a-z0-9-]*)\b`)
	roleToken       = regexp.MustCompile(`\b(manager-[a-z]+|plan-auditor|sync-auditor|super-advisor|e2e-tester|builder-harness|general-purpose|Explore)\b`)
	skillInvokeRe   = regexp.MustCompile(`Skill\("([a-z0-9][a-z0-9-]*)"\)`)
	dispatcherRefRe = regexp.MustCompile(`skills/moai/SKILL\.md`)
)

// extractFrontmatterSkills parses an agent body's YAML frontmatter `skills:`
// list (the static preload union).
func extractFrontmatterSkills(body string) []string {
	var skills []string
	lines := strings.Split(body, "\n")
	inFM, inSkills := false, false
	for i, ln := range lines {
		if i == 0 && strings.TrimSpace(ln) == "---" {
			inFM = true
			continue
		}
		if !inFM {
			continue
		}
		if strings.TrimSpace(ln) == "---" {
			break
		}
		if strings.HasPrefix(ln, "skills:") {
			inSkills = true
			continue
		}
		if inSkills {
			if strings.HasPrefix(ln, " ") || strings.HasPrefix(ln, "\t") {
				if m := moaiSkillToken.FindStringSubmatch(strings.TrimSpace(ln)); m != nil {
					skills = append(skills, m[1])
				}
				continue
			}
			inSkills = false
		}
	}
	return skills
}

// loadingInstructionLine reports whether a body line is a loading
// instruction the derivation sweep reads: a Skill("...") invocation, the
// quality-gates security reviewer's "loading the retained ..." instruction,
// or a dispatcher routing-table `Skills:` row.
func loadingInstructionLine(line string) bool {
	return strings.Contains(line, `Skill("`) ||
		strings.Contains(line, "loading the retained") ||
		strings.HasPrefix(strings.TrimSpace(line), "Skills:")
}

// delegationLine reports whether a line carries an agent delegation (a
// role-loading instruction): an `Agent: <role>` line, a frontmatter
// `agents: [...]` row, or a Route row of the delivery decomposition.
func delegationLine(line string) bool {
	t := strings.TrimSpace(line)
	return strings.HasPrefix(t, "Agent: ") ||
		strings.HasPrefix(t, "agents: [") ||
		strings.HasPrefix(t, "- **Route B") ||
		strings.HasPrefix(t, "- **Route A")
}

func collectSkillTokens(line string) []string {
	var out []string
	for _, loc := range moaiSkillToken.FindAllStringIndex(line, -1) {
		tok := line[loc[0]:loc[1]]
		// Skip family-stem wildcards (`moai-ref-*`, `moai-domain-*`) — the
		// capture stops at the trailing hyphen of the stem.
		if strings.HasSuffix(tok, "-") || strings.HasPrefix(line[loc[1]:], "-*") {
			continue
		}
		out = append(out, tok)
	}
	return out
}

// sweepSkillRefs returns the moai-* skill tokens named by the loading
// instructions of the given embedded-FS file.
func sweepSkillRefs(t *testing.T, fsys fs.FS, path string) map[string]bool {
	t.Helper()
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	refs := map[string]bool{}
	for _, ln := range strings.Split(string(data), "\n") {
		if !loadingInstructionLine(ln) {
			continue
		}
		for _, name := range collectSkillTokens(ln) {
			refs[name] = true
		}
		if m := skillInvokeRe.FindStringSubmatch(ln); m != nil && !strings.HasPrefix(m[1], "moai-") {
			// Non-moai Skill() names (harness built-ins) are out of the
			// catalog matrix's scope by definition.
			delete(refs, m[1])
		}
	}
	return refs
}

// sweepRoleRefs returns the role tokens named by the delegation lines of the
// given embedded-FS file.
func sweepRoleRefs(t *testing.T, fsys fs.FS, path string) map[string]bool {
	t.Helper()
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	refs := map[string]bool{}
	for _, ln := range strings.Split(string(data), "\n") {
		if !delegationLine(ln) {
			continue
		}
		for _, m := range roleToken.FindAllStringSubmatch(ln, -1) {
			if !nonMatrixRoles[m[1]] {
				refs[m[1]] = true
			}
		}
	}
	return refs
}

// sweepDefaultDelegateSkillRefs returns the moai-* tokens named ONLY by the
// default security delegate's "loading the retained ..." instruction (the
// sync flow's unconditional Phase 8 path — R-c's measured carrier for
// moai-ref-secops).
func sweepDefaultDelegateSkillRefs(t *testing.T, fsys fs.FS, path string) map[string]bool {
	t.Helper()
	refs := map[string]bool{}
	for _, ln := range strings.Split(mustRead(t, fsys, path), "\n") {
		if !strings.Contains(ln, "loading the retained") {
			continue
		}
		for _, name := range collectSkillTokens(ln) {
			refs[name] = true
		}
	}
	return refs
}

// catalogFSPath converts a catalog entry path to embedded-FS form.
func catalogFSPath(e *Entry) string {
	return strings.TrimSuffix(strings.TrimPrefix(e.Path, "templates/"), "/")
}

// l0ChainDocPaths enumerates the default plan/run/sync chain documents: the
// three workflow roots and everything under their subdirectories. The three
// roots ARE the D-Q4 answer's workflow surface.
func l0ChainDocPaths(t *testing.T, fsys fs.FS) []string {
	t.Helper()
	roots := []string{
		".claude/skills/moai/workflows/plan.md",
		".claude/skills/moai/workflows/run.md",
		".claude/skills/moai/workflows/sync.md",
	}
	for _, dir := range []string{"plan", "run", "sync"} {
		prefix := ".claude/skills/moai/workflows/" + dir
		names, err := fs.ReadDir(fsys, prefix)
		if err != nil {
			t.Fatalf("ReadDir %s: %v", prefix, err)
		}
		for _, n := range names {
			if !n.IsDir() && strings.HasSuffix(n.Name(), ".md") {
				roots = append(roots, prefix+"/"+n.Name())
			}
		}
	}
	return roots
}

// deriveL0Closure mechanically derives the L0 runtime skill closure from the
// template sources (design §2.3 three-tier union): L0 agent frontmatter
// skills unions + on-demand Skill() invoke sites in the L0 agent bodies + the
// dispatcher's plan/run/sync routing rows + the chain documents' loading
// instructions + the published command skills' dispatcher references. The
// classified-out entries (l0ClassifiedOut, explicit names only) are removed.
func deriveL0Closure(t *testing.T, fsys fs.FS, cat *Catalog) map[string]bool {
	t.Helper()
	derived := map[string]bool{}

	// Tier 1+3 — L0 agent bodies: frontmatter static preload + on-demand
	// Skill() invoke sites (fold B1).
	for i := range cat.Catalog.Core.Agents {
		e := &cat.Catalog.Core.Agents[i]
		for _, s := range extractFrontmatterSkills(mustRead(t, fsys, catalogFSPath(e))) {
			derived[s] = true
		}
		for name := range sweepSkillRefs(t, fsys, catalogFSPath(e)) {
			derived[name] = true
		}
	}

	// Tier 2 — the dispatcher's plan/run/sync routing rows. The rows live
	// under `### <command> - <title>` sections; only the D-Q4 three plus the
	// (default) row are the L0 surface — the other rows are the non-L0
	// command surface their bundles carry.
	l0Sections := map[string]bool{"plan": true, "run": true, "sync": true, "(default)": true}
	current := ""
	for _, ln := range strings.Split(mustRead(t, fsys, ".claude/skills/moai/SKILL.md"), "\n") {
		if strings.HasPrefix(ln, "### ") {
			current = strings.TrimSpace(strings.TrimPrefix(ln, "### "))
			if idx := strings.Index(current, " - "); idx >= 0 {
				current = current[:idx]
			}
			continue
		}
		if l0Sections[current] && loadingInstructionLine(ln) {
			for _, name := range collectSkillTokens(ln) {
				derived[name] = true
			}
		}
	}

	// Chain documents: only the default security delegate's loading line
	// class ("loading the retained ... skills" — R-c's measured default
	// Phase 8 carrier) is an UNCONDITIONAL default-path loading instruction
	// and contributes to the closure. The chain's conditional per-mission
	// injections are the derivation MATRIX's rows, not closure members —
	// they are swept by the matrix guard, not here.
	for _, p := range l0ChainDocPaths(t, fsys) {
		for name := range sweepDefaultDelegateSkillRefs(t, fsys, p) {
			derived[name] = true
		}
	}

	// The published command skills reference the dispatcher (the moai skill).
	for _, name := range l0CommandSkills {
		e, ok := cat.LookupSkill(name)
		if !ok {
			t.Fatalf("L0 command skill %q missing from catalog", name)
		}
		body := mustRead(t, fsys, catalogFSPath(e)+"/SKILL.md")
		if !dispatcherRefRe.MatchString(body) {
			t.Errorf("command skill %s does not reference the dispatcher (skills/moai/SKILL.md)", name)
		}
		derived["moai"] = true
	}

	// The classification-out list is EXPLICIT — only named entries leave.
	for name := range l0ClassifiedOut {
		delete(derived, name)
	}
	return derived
}

func mustRead(t *testing.T, fsys fs.FS, path string) string {
	t.Helper()
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// catalogL0SkillSet returns the catalog's L0 skill-view names minus the
// factory pair and the published command trio (which are pinned by their own
// decision constants, not by the runtime-skill closure sweep).
func catalogL0SkillSet(t *testing.T, cat *Catalog) map[string]bool {
	t.Helper()
	set := map[string]bool{}
	for _, e := range cat.Catalog.Core.Skills {
		set[e.Name] = true
	}
	for _, name := range l0FactorySkills {
		if !set[name] {
			t.Errorf("L0 factory skill %q missing from catalog core view", name)
		}
		delete(set, name)
	}
	for _, name := range l0CommandSkills {
		if !set[name] {
			t.Errorf("L0 command skill %q missing from catalog core view", name)
		}
		delete(set, name)
	}
	return set
}

// matrixRow is one install-coverage row: a referenced-but-not-L0 skill or
// role, its owning bundle (the `moai bundle add <bundle>` remediation
// target), and the referencing source (the reason).
type matrixRow struct {
	name   string
	kind   string // skill | role
	bundle string
	reason string
}

// TestUserInstallView_L0ClosureEqualsDerivedSet pins the catalog L0 view to
// the mechanically derived closure (EV-017 verdict basis; class repair R-c).
// A skill added to a workflow body or an L0 agent body flips this guard until
// the catalog view carries it — the drift the guard exists to catch.
func TestUserInstallView_L0ClosureEqualsDerivedSet(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	cat := loadCatalog(t)

	derived := deriveL0Closure(t, fsys, cat)
	listed := catalogL0SkillSet(t, cat)

	var missing, extra []string
	for name := range derived {
		if !listed[name] {
			missing = append(missing, name)
		}
	}
	for name := range listed {
		if !derived[name] {
			extra = append(extra, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 || len(extra) > 0 {
		t.Fatalf("L0 closure drift: derived-not-listed=%v listed-not-derived=%v (derived %d, listed %d)",
			missing, extra, len(derived), len(listed))
	}
	if len(derived) != 14 {
		t.Errorf("derived L0 closure holds %d skills, want the fourteen-skill union (design §2.3): %v", len(derived), sortedKeys(derived))
	}
}

// TestUserInstallView_L0AgentsPinned pins the L0 agent view to the resolved
// gate answer: the D-Q1 five plus manager-lead, who rides the catalog's
// per-entry dependency mechanism under the factory entry (extension E1) —
// not a sixth core agent.
func TestUserInstallView_L0AgentsPinned(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t)

	agentNames := map[string]bool{}
	for _, e := range cat.Catalog.Core.Agents {
		agentNames[e.Name] = true
	}
	for _, name := range l0CoreAgents {
		if !agentNames[name] {
			t.Errorf("D-Q1 core agent %q missing from catalog core.agents", name)
		}
	}
	if !agentNames["manager-lead"] {
		t.Fatalf("manager-lead (the factory entry's declared agent dependency) missing from catalog core.agents")
	}
	delete(agentNames, "manager-lead")
	for _, name := range l0CoreAgents {
		delete(agentNames, name)
	}
	if len(agentNames) > 0 {
		t.Errorf("catalog core.agents carries non-L0 agents %v — reclassify them into bundles (D-Q5)", sortedKeys(agentNames))
	}

	// The dependency edge: the factory entry declares manager-lead.
	for _, name := range l0FactorySkills {
		e, ok := cat.LookupSkill(name)
		if !ok {
			t.Fatalf("factory skill %q missing", name)
		}
		found := false
		for _, dep := range e.DependsAgents {
			if dep == "manager-lead" {
				found = true
			}
		}
		if !found {
			t.Errorf("factory entry %s does not declare its manager-lead agent dependency (factory-dispatch.md:104 resident deputy)", name)
		}
	}
}

// TestUserInstallView_PublishedCommandsFoldIntoCatalog asserts the catalog is
// the single membership SSOT over the published command skills: all
// seventeen appear in the catalog (the D-Q4 three in L0, the remaining
// fourteen in bundles per D-Q5) — no second name list.
func TestUserInstallView_PublishedCommandsFoldIntoCatalog(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	cat := loadCatalog(t)

	// The committed published-skill tree is the population (the same set
	// publishedSkillNames is pinned to by its own guard).
	var published []string
	names, err := fs.ReadDir(fsys, ".agents/skills")
	if err != nil {
		t.Fatalf("ReadDir .agents/skills: %v", err)
	}
	for _, n := range names {
		if n.IsDir() && strings.HasPrefix(n.Name(), "moai-") {
			published = append(published, n.Name())
		}
	}
	if len(published) != 17 {
		t.Fatalf("committed published-skill tree holds %d dirs, want 17: %v", len(published), published)
	}

	l0Trio := map[string]bool{}
	for _, name := range l0CommandSkills {
		l0Trio[name] = true
	}

	for _, name := range published {
		e, ok := cat.LookupSkill(name)
		if !ok {
			t.Errorf("published command skill %q is not a catalog entry — the catalog is the membership SSOT", name)
			continue
		}
		if l0Trio[name] != (e.Tier == TierCore) {
			t.Errorf("published command skill %q L0 membership mismatch: tier=%q expected-in-L0=%v", name, e.Tier, l0Trio[name])
		}
	}
}

// TestUserInstallView_DerivationMatrix is the install-coverage matrix (class
// repair R-c): sweep EVERY loading instruction across the default chain and
// the L0 agent bodies, cross each invoked role and loaded skill against the
// L0+bundle install set, and require every gap to resolve to a bundle row
// carrying the `moai bundle add <bundle>` remediation. The known rows the
// plan names (api-patterns / react-patterns / domain-database; manager-git
// the R-b precheck row) MUST be present; any UNKNOWN reference (no catalog
// entry at all) fails the guard.
func TestUserInstallView_DerivationMatrix(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	cat := loadCatalog(t)

	l0Skills := map[string]bool{}
	for _, e := range cat.Catalog.Core.Skills {
		l0Skills[e.Name] = true
	}
	l0Agents := map[string]bool{}
	for _, e := range cat.Catalog.Core.Agents {
		l0Agents[e.Name] = true
	}

	rows := map[string]matrixRow{}
	scan := func(path string) {
		for name := range sweepSkillRefs(t, fsys, path) {
			if l0Skills[name] {
				continue
			}
			if _, out := l0ClassifiedOut[name]; out {
				continue
			}
			e, ok := cat.LookupSkill(name)
			if !ok {
				t.Errorf("MATRIX_GAP_UNRESOLVED: %s loads skill %q which has no catalog entry", path, name)
				continue
			}
			rows[name] = matrixRow{name: name, kind: "skill", bundle: bundleOfTier(e.Tier), reason: path}
		}
		for name := range sweepRoleRefs(t, fsys, path) {
			if l0Agents[name] {
				continue
			}
			e, ok := cat.LookupAgent(name)
			if !ok {
				t.Errorf("MATRIX_GAP_UNRESOLVED: %s delegates role %q which has no catalog entry", path, name)
				continue
			}
			rows[name] = matrixRow{name: name, kind: "role", bundle: bundleOfTier(e.Tier), reason: path}
		}
	}

	// (i) the L0 agent bodies; (ii) the default-chain documents.
	for i := range cat.Catalog.Core.Agents {
		scan(catalogFSPath(&cat.Catalog.Core.Agents[i]))
	}
	for _, p := range l0ChainDocPaths(t, fsys) {
		scan(p)
	}

	// Known rows the plan names (measured this tree).
	known := map[string]string{
		"moai-ref-api-patterns":   "backend",
		"moai-ref-react-patterns": "frontend",
		"moai-domain-database":    "backend",
		"manager-git":             "delivery",
	}
	for name, wantBundle := range known {
		got, ok := rows[name]
		if !ok {
			t.Errorf("matrix row missing for %q — the plan's known row set must be captured", name)
			continue
		}
		if got.bundle != wantBundle {
			t.Errorf("matrix row %q remediation bundle = %q, want %q", name, got.bundle, wantBundle)
		}
	}

	// Every row's remediation names an existing bundle.
	for _, r := range rows {
		if r.bundle == TierCore {
			t.Errorf("matrix row %q classified into core but absent from the L0 view", r.name)
			continue
		}
		if _, ok := cat.Catalog.OptionalPacks[r.bundle]; !ok {
			t.Errorf("matrix row %q remediation bundle %q does not exist in the catalog", r.name, r.bundle)
		}
	}

	// The html-report exclusion is EXPLICIT: it is referenced on the chain
	// (manager-docs:224) and must NOT have leaked into L0.
	if l0Skills["moai-domain-html-report"] {
		t.Error("moai-domain-html-report is classified OUT (mission-type injection) but sits in the L0 view (JD-4: the exclusion is the explicit list, never a prefix)")
	}

	names := make([]string, 0, len(rows))
	for name, r := range rows {
		names = append(names, fmt.Sprintf("%s→%s(%s)", name, r.bundle, r.kind))
	}
	sort.Strings(names)
	t.Logf("derived matrix rows (%d): %v", len(names), names)
}

// TestUserInstallView_NoEntryLost asserts the reclassification conserved the
// catalog population. An entry rides at most ONE optional-pack section; an
// entry MAY additionally sit in core (the E3 shared-asset case — the
// historical devops pack carries the L0 trio), which is not a duplicate.
func TestUserInstallView_NoEntryLost(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t)

	packCount := map[string]int{}
	record := func(name string) { packCount[name]++ }
	for _, pack := range cat.Catalog.OptionalPacks {
		for _, e := range pack.Skills {
			record(e.Name)
		}
		for _, e := range pack.Agents {
			record(e.Name)
		}
	}
	for name, count := range packCount {
		if count > 1 {
			t.Errorf("CATALOG_DUPLICATE_ENTRY: %s appears in %d pack sections — an entry rides at most one pack", name, count)
		}
	}

	// The population is conserved: the catalog still carries every embedded
	// .claude/skills directory, every .claude/agents/moai body, and the
	// published command skills (the same populations TestAllSkillsInCatalog
	// and TestAllAgentsInCatalog audit — this is the no-loss half).
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	all := cat.AllEntries()
	if len(all) == 0 {
		t.Fatal("catalog carries no entries")
	}
	names := map[string]bool{}
	for _, e := range all {
		names[e.Name] = true
	}
	for _, dir := range []string{".claude/skills", ".claude/agents/moai", ".agents/skills"} {
		entries, err := fs.ReadDir(fsys, dir)
		if err != nil {
			t.Fatalf("ReadDir %s: %v", dir, err)
		}
		for _, n := range entries {
			if !n.IsDir() || !strings.HasPrefix(n.Name(), "moai") {
				continue
			}
			if !names[n.Name()] {
				t.Errorf("CATALOG_ENTRY_MISSING: %s/%s exists on disk but is not a catalog entry", dir, n.Name())
			}
		}
	}
}

// bundleOfTier maps an entry tier to its owning bundle name ("" result only
// for the core tier).
func bundleOfTier(tier string) string {
	if tier == TierCore {
		return TierCore
	}
	if name, ok := strings.CutPrefix(tier, TierOptionalPackPrefix); ok {
		return name
	}
	return ""
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
