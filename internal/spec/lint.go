// Package spec provides SPEC document parsing and validation functionality
// lint.go is the core engine of moai spec lint CLI, validating SPEC documents
// for EARS compliance, coverage, DAG, etc. through Rule interface and Linter struct
package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// Severity represents the severity of a finding.
type Severity string

const (
	// SeverityError is a critical error that causes linter abnormal termination.
	SeverityError Severity = "error"
	// SeverityWarning is a warning that does not affect exit code in default mode.
	// Escalated to error when --strict flag is used.
	SeverityWarning Severity = "warning"
	// SeverityInfo is an informational message.
	SeverityInfo Severity = "info"
)

// JSON serialization includes file, line, severity, code, message fields.
type Finding struct {
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	// Advisory marks a warning that --strict must NOT escalate to error.
	// Set for (a) findings on grandfather-era SPECs (V2.x / V3R2-R4 / V3R5 —
	// same era classification as `moai spec audit`) and (b) inherently
	// heuristic rules such as StatusGitConsistency whose git-implied signal
	// is environment-dependent.
	Advisory bool `json:"advisory,omitempty"`
}

// Report represents lint execution results
type Report struct {
	// Findings is a list of all findings
	Findings []Finding
	// Strict is the state of --strict flag. Affects HasErrors() calculation
	Strict bool
}

// In strict mode, warnings are also considered errors
func (r *Report) HasErrors() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityError {
			return true
		}
		if r.Strict && f.Severity == SeverityWarning && !f.Advisory {
			return true
		}
	}
	return false
}

// ToJSON serializes findings to JSON byte slice
// Returns empty JSON array ([]) if findings is nil
func (r *Report) ToJSON() ([]byte, error) {
	findings := r.Findings
	if findings == nil {
		findings = []Finding{}
	}
	return json.Marshal(findings)
}

// ToSARIF serializes findings to JSON byte slice in SARIF 2.1.0 format
func (r *Report) ToSARIF() ([]byte, error) {
	return marshalSARIF(r.Findings)
}

// LinterOptions is the options for creating Linter
type LinterOptions struct {
	// RegistryPath is the path to zone registry markdown file
	// Skips DanglingRuleReference check
	RegistryPath string
	// BaseDir is the base directory for SPEC file search in no-args execution
	BaseDir string
	// Strict is the state of --strict flag
	Strict bool
}

// @MX:ANCHOR: [AUTO] Linter is the central lint engine; all lint rules are dispatched through it.
// @MX:REASON: [AUTO] Fan-in hub — CLI, tests, and future integrations all call Linter.Lint.
type Linter struct {
	opts     LinterOptions
	registry *constitution.Registry
	rules    []Rule
}

// NewLinter creates a new Linter instance
// Loads zone registry if options.RegistryPath is specified
func NewLinter(opts LinterOptions) *Linter {
	l := &Linter{opts: opts}

	// Load zone registry
	if opts.RegistryPath != "" {
		projectDir := opts.BaseDir
		if projectDir == "" {
			projectDir = "."
		}
		reg, err := constitution.LoadRegistry(opts.RegistryPath, projectDir)
		if err == nil {
			l.registry = reg
		}
		// registry load failure is silent — skip DanglingRuleReference check
	}

	// HaikuResidualRule scans the project tree (not a SPEC document), so it
	// needs the project root. Default to "." matching discoverSPECs behavior.
	haikuBaseDir := opts.BaseDir
	if haikuBaseDir == "" {
		haikuBaseDir = "."
	}

	l.rules = []Rule{
		&EARSModalityRule{},
		&REQIDUniquenessRule{},
		&CoverageRule{},
		&FrontmatterSchemaRule{},
		&DependencyExistsRule{},
		&OutOfScopeRule{},
		&BreakingChangeIDRule{},
		&StatusValueEnumRule{},
		&StatusCaseNormalizationRule{},
		&StatusGitConsistencyRule{},
		&OwnershipTransitionRule{},
		// StatusTransitionValidityRule — SPEC-STATUS-TRANSITION-VALIDITY-001
		// (card t376). Sits beside OwnershipTransitionRule and answers a
		// different question: is the (prev, curr) pair itself a legal edge,
		// regardless of who signed the commit. Emits two codes —
		// StatusTransitionInvalid and StatusTokenUnrecognized — neither of
		// which belongs in eraDemotableCodes: that map is consulted only for
		// SeverityError findings, so an entry there would be inert while
		// reading as intent.
		&StatusTransitionValidityRule{},
		// ArtifactStatusFieldForbiddenRule — SPEC-ARTIFACT-STATELESS-001 M2,
		// REQ-AST-001-004. Per-SPEC: it reads the SPEC's own sibling artifacts
		// via filepath.Dir(doc.Path). Severity is `error` and the code is
		// deliberately NOT in eraDemotableCodes — that absence holds only
		// because the D1 corpus cleanup lands in the same SPEC (REQ-AST-001-006
		// / -010); splitting the cleanup out inverts the decision.
		&ArtifactStatusFieldForbiddenRule{},
		// MovingRefUnpinnedRule — SPEC-MOVING-REF-GUARD-001 M3, REQ-MRG-001.
		// Per-SPEC (not cross-SPEC): it reads the SPEC's own sibling artifacts
		// via filepath.Dir(doc.Path), so lint.skip and era demotion both apply.
		// Severity is warning only (spec.md §D.5) and the code is deliberately
		// NOT in eraDemotableCodes.
		&MovingRefUnpinnedRule{},
		// SyncSHASlotFormatRule — SPEC-SYNC-SHA-SLOT-FORMAT-001 M3, REQ-SSF-004.
		// Per-SPEC (not cross-SPEC): it reads the SPEC's own sibling progress.md
		// via filepath.Dir(doc.Path), so lint.skip and era demotion both apply.
		// Severity is warning only (spec.md §D.3) — all five corpus findings sit
		// in `completed` SPECs, which terminalStatusEnum already shelters, so the
		// rule contributes nothing to the --strict exit status on the corpus as
		// it stands; an `error` would put five closed SPECs' history into the
		// strict path with no shelter and make lint.skip the rational response.
		// The code is deliberately NOT in eraDemotableCodes (REQ-SSF-007): that
		// map demotes ERRORS, so the entry would be inert for a warning, and an
		// inert entry in a policy map reads as intent. AC-SSF-010 guards it.
		&SyncSHASlotFormatRule{},
		// cross-SPEC rules
		&DependencyCycleRule{},
		&DuplicateSPECIDRule{},
		// HaikuResidualRule — cross-SPEC HARD gate (NOT skip-able; CheckAll
		// findings bypass applylintSkip). SPEC-AGENT-ARCH-V2-001 M3c REQ-AA2-012.
		&HaikuResidualRule{baseDir: haikuBaseDir},
		// Registry required
		&ZoneRegistryRule{registry: l.registry},
	}

	return l
}

// If paths is nil or empty, automatically discover spec.md files under opts.BaseDir
//
// @MX:ANCHOR: [AUTO] Lint is the primary entry point; orchestrates rule execution across all SPECs
// @MX:REASON: [AUTO] Fan-in hub — all callers (CLI, tests) go through this method
func (l *Linter) Lint(paths []string) (*Report, error) {
	// REQ-PERF-001-A/B: initialize per-run git-query cache. This memoizes
	// git rev-parse environment checks so they run once per Lint() instead
	// of once per SPEC (eliminating ~2×N redundant spawns). The cache is
	// discarded at Lint() exit (per-run invalidation — REQ-PERF-001-B).
	startGitQueryCache()
	defer stopGitQueryCache()

	// dirFindings collects directory-level findings emitted only on the
	// auto-discovery path (paths empty). When a caller passes explicit paths
	// it bypasses directory scanning, so root-integrity is not checked.
	var dirFindings []Finding
	if len(paths) == 0 {
		discovered, err := discoverSPECs(l.opts.BaseDir)
		if err != nil {
			return nil, fmt.Errorf("SPEC discovery failed: %w", err)
		}
		paths = discovered
		// Surface non-SPEC entries (loose files, non-SPEC directories) that
		// discoverSPECs silently skips. Design choice: a standalone function
		// (option 6b) rather than a Rule, because this is a directory-level
		// concern that does not fit the per-SPECDoc Rule interface, and it
		// keeps discoverSPECs' signature unchanged for its many callers.
		dirFindings = lintSpecsDirRootIntegrity(l.opts.BaseDir)
	}

	// Parse SPEC documents
	var docs []*SPECDoc
	var findings []Finding

	for _, path := range paths {
		doc := parseSPECDoc(path)
		docs = append(docs, doc)
		if doc.ParseError != nil {
			findings = append(findings, Finding{
				File:     path,
				Line:     1,
				Severity: SeverityError,
				Code:     "ParseFailure",
				Message:  fmt.Sprintf("SPEC parsing failed: %v", doc.ParseError),
			})
		}
	}

	for _, doc := range docs {
		if doc.ParseError != nil {
			continue // Skip rules for failed SPEC
		}
		var docFindings []Finding
		for _, rule := range l.rules {
			// cross-SPEC rules are processed later
			if _, ok := rule.(crossSPECRule); ok {
				continue
			}
			ruleFindings := rule.Check(doc, docs)
			ruleFindings = applylintSkip(ruleFindings, doc.LintSkip)
			docFindings = append(docFindings, ruleFindings...)
		}
		// The two disjuncts are kept separate so the demotion annotation can
		// name the one that fired (REQ-STV-008). The decision itself — demote
		// when EITHER holds — is unchanged.
		cause := demotionCause{
			GrandfatheredEra: isGrandfatheredSpecDir(filepath.Dir(doc.Path)),
			TerminalStatus:   terminalStatusEnum[doc.Frontmatter.Status],
		}
		findings = append(findings, applyEraDemotion(docFindings, cause)...)
	}

	for _, rule := range l.rules {
		if cr, ok := rule.(crossSPECRule); ok {
			crossFindings := cr.CheckAll(docs)
			findings = append(findings, crossFindings...)
		}
	}

	return &Report{
		Findings: append(findings, dirFindings...),
		Strict:   l.opts.Strict,
	}, nil
}

type crossSPECRule interface {
	CheckAll(docs []*SPECDoc) []Finding
}

// @MX:NOTE: [AUTO] Rule interface inspects a single SPEC document
type Rule interface {
	Code() string
	// Check inspects a single SPEC document and returns findings
	Check(doc *SPECDoc, all []*SPECDoc) []Finding
}

// eraDemotableCodes are the structural rules that grandfather-era SPECs are
// exempt from failing on. The findings stay visible (as advisory warnings)
// but no longer gate the lint exit code — mirroring the grandfather clause
// `moai spec audit` applies (AC-LSG-017): pre-V3R6 SPECs predate the
// structural requirements these rules enforce.
var eraDemotableCodes = map[string]bool{
	"MissingExclusions":  true,
	"FrontmatterInvalid": true,
}

// isGrandfatheredSpecDir reuses the era classifier (era.go — the SSOT shared
// with `moai spec audit` and drift detection; NOT a parallel heuristic) to
// report whether the SPEC directory is grandfather-clause-protected.
func isGrandfatheredSpecDir(specDir string) bool {
	signals, err := LoadEraSignalsFromDir(specDir)
	if err != nil {
		return false
	}
	era, _ := ClassifyEra(signals)
	return era.EraFinal()
}

// demotionCause records WHY a SPEC's findings are being demoted. The two
// reasons are independent and either alone is sufficient, so the annotation
// appended to a demoted finding must be able to name the one that actually
// fired: a document demoted solely because its status is terminal is not
// grandfathered, and saying so misstates the finding's own cause
// (REQ-STV-008, SPEC-STATUS-TRANSITION-VALIDITY-001).
type demotionCause struct {
	// GrandfatheredEra: the SPEC directory classifies as V2.x / V3R2-R4 / V3R5.
	GrandfatheredEra bool
	// TerminalStatus: the frontmatter status is in terminalStatusEnum.
	TerminalStatus bool
}

// demoted reports whether any cause fired.
func (c demotionCause) demoted() bool { return c.GrandfatheredEra || c.TerminalStatus }

// String names the cause(s) that fired, for the demotion annotation.
func (c demotionCause) String() string {
	switch {
	case c.GrandfatheredEra && c.TerminalStatus:
		return "grandfathered era + terminal lifecycle status"
	case c.TerminalStatus:
		return "terminal lifecycle status"
	case c.GrandfatheredEra:
		return "grandfathered era"
	}
	return ""
}

// applyEraDemotion downgrades structural ERROR findings on protected SPECs
// to advisory warnings, and marks the SPEC's remaining warnings advisory so
// --strict does not escalate them. Protected = grandfather-era (V2.x /
// V3R2-R4 / V3R5) OR terminal lifecycle status (terminalStatusEnum: closed
// history — retro-enforcing later structural rules on closed SPECs is the
// same false-positive class the grandfather clause exists for). Active
// modern-era SPECs pass through untouched — full enforcement.
//
// The appended annotation names the cause that fired (REQ-STV-008). Only the
// message changed; the demotion DECISION and its blanket Advisory marking are
// unchanged (spec.md §C non-goal).
func applyEraDemotion(findings []Finding, cause demotionCause) []Finding {
	if !cause.demoted() {
		return findings
	}
	annotation := " [" + cause.String() + " — downgraded to warning]"
	for i := range findings {
		f := &findings[i]
		switch {
		case f.Severity == SeverityError && eraDemotableCodes[f.Code]:
			f.Severity = SeverityWarning
			f.Advisory = true
			f.Message += annotation
		case f.Severity == SeverityWarning:
			f.Advisory = true
		}
	}
	return findings
}

// applylintSkip removes findings that match doc's lint.skip code list
func applylintSkip(findings []Finding, skipCodes []string) []Finding {
	if len(skipCodes) == 0 {
		return findings
	}
	skipSet := make(map[string]bool, len(skipCodes))
	for _, code := range skipCodes {
		skipSet[code] = true
	}
	var result []Finding
	for _, f := range findings {
		if !skipSet[f.Code] {
			result = append(result, f)
		}
	}
	return result
}

// discoverSPECs finds spec.md files matching baseDir/.moai/specs/SPEC-*/spec.md or baseDir/SPEC-*/spec.md pattern
func discoverSPECs(baseDir string) ([]string, error) {
	if baseDir == "" {
		baseDir = "."
	}

	var paths []string

	// SPEC-*/spec.md pattern directly under baseDir
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %q: %w", baseDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "SPEC-") {
			continue
		}
		candidate := filepath.Join(baseDir, entry.Name(), "spec.md")
		if _, err := os.Stat(candidate); err == nil {
			paths = append(paths, candidate)
		}
	}

	return paths, nil
}

// lintSpecsDirRootIntegrity scans baseDir (typically .moai/specs) for entries
// that do not belong in a SPEC root and emits a SpecsDirForeignEntry warning
// for each. discoverSPECs silently skips these; this function surfaces them so
// accidentally-committed roadmap documents or non-SPEC directories are visible.
//
// Whitelist: entries whose name starts with "_" (e.g. "_archive") are exempt —
// the underscore prefix is the established ignore convention.
//
// Severity is warning (not error) so that existing repositories with stray
// entries (e.g. this dev repo's own UPGRADE-HARNESS-DESIGN/) are surfaced
// without breaking their lint exit code.
//
// It returns nil (no findings) if baseDir cannot be read — discoverSPECs
// already surfaces directory-read failures as a hard error upstream.
func lintSpecsDirRootIntegrity(baseDir string) []Finding {
	if baseDir == "" {
		baseDir = "."
	}
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil
	}

	const msgTmpl = "%q: non-SPEC entry in .moai/specs/ — only SPEC-<DOMAIN>-<NNN>/ directories allowed; ROADMAP/planning docs belong in .moai/plans/ or project root (see spec-frontmatter-schema.md § Root Integrity)"

	var findings []Finding
	for _, entry := range entries {
		name := entry.Name()
		// Underscore-prefixed entries are whitelisted (_archive convention).
		if strings.HasPrefix(name, "_") {
			continue
		}
		// Valid SPEC-* directories are the expected occupants.
		if entry.IsDir() && strings.HasPrefix(name, "SPEC-") {
			continue
		}
		// Anything else is foreign: loose files (any extension) and
		// non-SPEC directories (e.g. NOTES/, UPGRADE-HARNESS-DESIGN/).
		findings = append(findings, Finding{
			File:     filepath.Join(baseDir, name),
			Line:     1,
			Severity: SeverityWarning,
			Code:     "SpecsDirForeignEntry",
			Message:  fmt.Sprintf(msgTmpl, name),
		})
	}
	return findings
}

// --- SPECDoc ---

// SPECFrontmatter represents the YAML frontmatter of a SPEC document
type SPECFrontmatter struct {
	ID           string   `yaml:"id"`
	Title        string   `yaml:"title"`
	Version      string   `yaml:"version"`
	Status       string   `yaml:"status"`
	Created      string   `yaml:"created"`
	Updated      string   `yaml:"updated"`
	Author       string   `yaml:"author"`
	Priority     string   `yaml:"priority"`
	Phase        string   `yaml:"phase"`
	Module       string   `yaml:"module"`
	Dependencies []string `yaml:"dependencies"`
	BcID         []string `yaml:"bc_id"`
	Lifecycle    string   `yaml:"lifecycle"`
	Tags         string   `yaml:"tags"`
	Breaking     bool     `yaml:"breaking"`
	RelatedRule  []string `yaml:"related_rule"`
	// LintConfig is a nested structure containing lint.skip code list.
	LintConfig struct {
		Skip []string `yaml:"skip"`
	} `yaml:"lint"`
	// HarnessLevel is the optional per-SPEC harness level override.
	// REQ-HRN-001-015: when present, harness_level: minimal|standard|thorough takes
	// precedence over auto-detection. It is not one of the 12 required fields, so
	// FrontmatterSchemaRule does not report its absence as a finding.
	HarnessLevel string `yaml:"harness_level,omitempty"`
	// Era is the optional era classification override (V2.x | V3R2-R4 | V3R5 | V3R6).
	// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 REQ-LSG-002 / AC-LSG-013: when present,
	// overrides auto-detection in ClassifyEra. Not one of the 12 required fields.
	Era string `yaml:"era,omitempty"`
	// Tier is the optional SPEC complexity tier (S | M | L). SPEC-WEB-CONSOLE-011
	// REQ-WC11-042: the read-only web SPEC board renders it as an OPTIONAL badge
	// (absent → no badge, not an error). Not one of the 12 required fields, so
	// FrontmatterSchemaRule does not report its absence.
	Tier string `yaml:"tier,omitempty"`
}

type REQEntry struct {
	ID   string
	Text string
	Line int

	// Widened records that this entry reached doc.REQs ONLY because the
	// extraction pattern was widened by SPEC-COVERAGE-RULE-SCOPE-001 — the
	// narrow reqLinePattern did not collect it. It is the provenance flag the
	// severity treatment keys on: a widened-only entry is one the linter was
	// blind to until now, so a finding against it is a newly-surfaced corpus
	// fact rather than a regression the author introduced, and it reports
	// without gating. An entry the narrow pattern already collected carries
	// false here and its rules behave exactly as before.
	Widened bool
}

// SPECDoc represents a parsed SPEC document.
type SPECDoc struct {
	Path        string
	Frontmatter SPECFrontmatter
	Body        string
	Criteria    []Acceptance
	REQs        []REQEntry
	ParseError  error
	LintSkip    []string
}

// reqIDPattern validates a REQ ID. It is the VALIDATION half of a pair whose
// other half is the extraction pattern that populates doc.REQs; the two must be
// kept in a deliberate relationship, because InvalidREQIDRule can only ever
// judge IDs the extraction hands it.
//
// SPEC-COVERAGE-RULE-SCOPE-001 M2 widened this from `^REQ-[A-Z]{2,5}-\d{3}-\d{3}$`.
// That shape admitted 260 of the 1,085 REQ definition lines the corpus actually
// carries; the widened extraction (reqLineWidePattern) would have made the other
// 825 fire InvalidREQID corpus-wide. The shapes now accepted, all measured live:
//
//	REQ-HOOK-001          three-segment
//	REQ-WF001-001         digits inside the domain segment
//	REQ-VNRN-RT-001-001   five-segment, two-part domain
//	REQ-HRN-FND-001       two-part alpha domain
//	REQ-TUX1-001          domain ending in a digit
//	REQ-WC01-001          alphanumeric domain
//
// It is deliberately NARROWER than the extraction, and that gap is load-bearing.
// Aligning validation exactly to extraction would make InvalidREQIDRule vacuous:
// every ID it judges would pass by construction, and the rule's non-execution
// would be indistinguishable from its success
// (`.claude/rules/moai/development/verification-completeness.md` §1.1). The
// retained rejection class — a domain segment not starting with a letter, a
// domain of three or more segments, a numeric tail that is not one or two groups
// of exactly three digits — is reachable through the extraction, and
// TestReqIDPattern_RejectsShapesTheExtractionAccepts is the mutant probe that
// keeps it reachable.
//
// The rejection class is REACHABLE, and that is measured rather than inferred
// from this pattern's shape. Corpus-level mutant probe
// (TestCorpusRejectedREQIDDecomposition section [F]): validation aligned exactly
// to the extraction — the option-(ii) mutant — fires InvalidREQID 0 times across
// the corpus, while this pattern fires 6. A delta of 0 would have meant the
// rejection class is unreachable on real documents whatever the regexp says;
// the delta is 6.
//
// Those 6 are REQ-256K-001..006 in SPEC-HANDOFF-CTXGUIDE-001, whose domain
// segment starts with a digit. They are genuine convention violations, also
// measured: 0 of 706 SPEC directories carry a digit-initial domain segment, and
// specIDPattern codifies the same letter-initial rule for SPEC IDs. Six
// digit-initial domain tokens out of 1,085 REQ definitions is an outlier, not an
// unrepresented convention.
//
// Consequently InvalidREQIDRule is NOT a deletion candidate: it fires on real
// input, and `moai spec lint --help` advertises "REQ ID uniqueness" against a
// rule that still checks something. Had the residual been 0, the rule would have
// been a deletion candidate — an advertised check that cannot fire is the exact
// defect SPEC-COVERAGE-RULE-SCOPE-001 was opened to document, and leaving one in
// place while repairing a vacuous parser would reproduce the defect inside its
// own repair. Deletion is not this SPEC's decision to make either way; the
// disposition is recorded so it stays visible.
var reqIDPattern = regexp.MustCompile(`^REQ-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)?-\d{3}(?:-\d{3})?$`)

// REQ-SPC-003-001: The system SHALL do X."
var reqLinePattern = regexp.MustCompile(`-\s+(REQ-[A-Z]{2,5}-\d{3}-\d{3})\s*:\s*(.+)`)

// parseSPECDoc parses the SPEC document at the given path
func parseSPECDoc(path string) *SPECDoc {
	doc := &SPECDoc{Path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		doc.ParseError = fmt.Errorf("failed to read file: %w", err)
		return doc
	}

	content := string(data)

	fm, body, err := extractFrontmatter(content)
	if err != nil {
		doc.ParseError = fmt.Errorf("frontmatter parsing error: %w", err)
		return doc
	}

	doc.Frontmatter = fm
	doc.Body = body
	doc.LintSkip = fm.LintConfig.Skip

	// Parse REQ list with the widened pattern, marking every entry the narrow
	// pattern would NOT have collected. SPEC-COVERAGE-RULE-SCOPE-001 M3.
	doc.REQs = parseREQsWithProvenance(body)

	// Parse Acceptance Criteria
	criteria, _ := ParseAcceptanceCriteria(body, false)
	doc.Criteria = criteria

	return doc
}

// ExtractFrontmatter splits a SPEC document's content into its YAML frontmatter and body and returns both.
// REQ-HRN-001-003: HarnessRouter consumes SPECFrontmatter.Priority, Tags, and HarnessLevel.
func ExtractFrontmatter(content string) (SPECFrontmatter, string, error) {
	return extractFrontmatter(content)
}

func extractFrontmatter(content string) (SPECFrontmatter, string, error) {
	var fm SPECFrontmatter

	if !strings.HasPrefix(content, "---") {
		return fm, content, fmt.Errorf("YAML frontmatter missing or does not start with '---'")
	}

	rest := content[3:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return fm, content, fmt.Errorf("could not find closing '---' for frontmatter")
	}

	yamlPart := rest[:endIdx]
	body := rest[endIdx+4:] // After "\n---"

	if err := yaml.Unmarshal([]byte(yamlPart), &fm); err != nil {
		return fm, body, fmt.Errorf("YAML parsing error: %w", err)
	}

	return fm, body, nil
}

func parseREQs(body string) []REQEntry {
	var reqs []REQEntry
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		matches := reqLinePattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			reqs = append(reqs, REQEntry{
				ID:   matches[1],
				Text: strings.TrimSpace(matches[2]),
				Line: i + 1,
			})
		}
	}
	return reqs
}

// reqFindingSeverity resolves the severity and advisory flag for a finding
// emitted against a single REQ entry.
//
// A finding on a WIDENED-ONLY entry (one the narrow reqLinePattern never
// collected) reports without gating: the linter was blind to that line until
// SPEC-COVERAGE-RULE-SCOPE-001 M3 wired the widened collector, so the finding
// is a newly-surfaced corpus fact rather than a regression the author
// introduced. Landing 25 ModalityMalformed and 6 InvalidREQID errors — the
// measured live counts — on a corpus that was never linted against them would
// make bulk suppression the rational response, which is the outcome the
// widening exists to avoid.
//
// A finding on an entry the narrow pattern already collected is untouched, so
// every pre-existing behavior is byte-identical.
//
// Advisory is set at the emission site because eraDemotableCodes is consulted
// only for SeverityError findings — a warning can never reach it.
//
// THE DEBT: like CoverageRule's advisory severity, this is a guard that
// declares without enforcing on the widened population. The promotion condition
// is that the widened-only corpus findings are remediated or exempted, after
// which the `Widened` branch is deleted and every finding gates. That condition
// is prose and nothing here fires when it is met; forgetting it leaves a check
// whose non-execution is indistinguishable from its success — the defect class
// this SPEC exists to document.
func reqFindingSeverity(req REQEntry, base Severity) (Severity, bool) {
	if req.Widened {
		return SeverityWarning, true
	}
	return base, false
}

// collectAllREQIDs collects REQ IDs from all nodes (leaf + non-leaf) in Acceptance tree
func collectAllREQIDs(criteria []Acceptance) map[string]bool {
	covered := make(map[string]bool)
	var visit func(ac *Acceptance)
	visit = func(ac *Acceptance) {
		for _, reqID := range ac.RequirementIDs {
			covered["REQ-"+reqID] = true
		}
		for i := range ac.Children {
			visit(&ac.Children[i])
		}
	}
	for i := range criteria {
		visit(&criteria[i])
	}
	return covered
}

// --- Rule implementations ---

// EARSModalityRule checks REQ text for EARS modality compliance
// AND emits GEARS migration warnings for legacy IF/THEN patterns.
//
// GEARS migration policy (SPEC-V3R6-GEARS-MIGRATION-001):
//   - WHEN/WHILE/WHERE/Ubiquitous "The system shall" remain canonical.
//   - IF ... THEN is deprecated; flagged with LegacyEARSKeyword (warning).
//   - Backward-compat window: 6 months from v3.0.0 OR until
//     SPEC-V3R6-GEARS-SWEEP-001 bulk-rewrites all 88 existing SPECs,
//     whichever comes first. After window: SPEC-V3R6-V3-CUTOVER-001
//     promotes the warning to error.
//   - --strict mode escalates warnings to errors immediately (existing
//     Report.HasErrors() behavior; opt-in for CI authors).
//
// Implements REQ-SPC-003-003, REQ-SPC-003-050, REQ-GM-002, REQ-GM-006, REQ-GM-009
type EARSModalityRule struct{}

func (r *EARSModalityRule) Code() string { return "ModalityMalformed" }

func (r *EARSModalityRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	var findings []Finding
	for _, req := range doc.REQs {
		sev, adv := reqFindingSeverity(req, SeverityError)
		// Existing legacy check (unchanged) — emits error when SHALL is missing.
		if isModalityMalformed(req.Text) {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: sev,
				Advisory: adv,
				Code:     "ModalityMalformed",
				Message:  fmt.Sprintf("REQ %s: EARS modality violation — SHALL missing or format mismatch: %q", req.ID, req.Text),
			})
		}
		// NEW: GEARS migration warning for legacy IF/THEN patterns.
		// SPEC-V3R6-GEARS-MIGRATION-001 REQ-GM-002 + REQ-GM-006.
		if isLegacyEARSPattern(req.Text) {
			_, legacyAdv := reqFindingSeverity(req, SeverityWarning)
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: SeverityWarning,
				Advisory: legacyAdv,
				Code:     "LegacyEARSKeyword",
				Message:  fmt.Sprintf("REQ %s: GEARS migration: replace IF/THEN with WHEN/event normalization; see https://adk.mo.ai.kr/en/workflow-commands/moai-plan/#gears-notation", req.ID),
			})
		}
	}
	return findings
}

// isModalityMalformed checks if REQ text violates EARS modality
func isModalityMalformed(text string) bool {
	upper := strings.ToUpper(text)

	if strings.HasPrefix(upper, "WHEN ") && !strings.Contains(upper, " SHALL") {
		return true
	}
	if strings.HasPrefix(upper, "WHILE ") && !strings.Contains(upper, " SHALL") {
		return true
	}
	if strings.HasPrefix(upper, "WHERE ") && !strings.Contains(upper, " SHALL") {
		return true
	}
	if strings.HasPrefix(upper, "IF ") && !strings.Contains(upper, " SHALL") {
		return true
	}
	// Ubiquitous format: Must start with "The [system] SHALL"
	if strings.HasPrefix(upper, "THE ") && !strings.Contains(upper, " SHALL") {
		return true
	}
	return false
}

// isLegacyEARSPattern returns true ONLY for IF ... THEN REQs.
// Other EARS keywords (WHEN/WHILE/WHERE/Ubiquitous) are GEARS-compatible.
// SPEC-V3R6-GEARS-MIGRATION-001 REQ-GM-002.
//
// GEARS migration policy:
//   - WHEN/WHILE/WHERE/Ubiquitous "The system shall" remain canonical.
//   - IF ... THEN is deprecated; flagged with LegacyEARSKeyword (warning).
//   - Backward-compat window: 6 months from v3.0.0 OR until
//     SPEC-V3R6-GEARS-SWEEP-001 bulk-rewrites all 88 existing SPECs,
//     whichever comes first. After window: SPEC-V3R6-V3-CUTOVER-001
//     promotes the warning to error.
//   - --strict mode escalates warnings to errors immediately (existing
//     Report.HasErrors() behavior; opt-in for CI authors).
//
// Reference: SPEC-V3R6-GEARS-MIGRATION-001/spec.md §1 + §3 REQ-GM-002/006/009.
func isLegacyEARSPattern(text string) bool {
	upper := strings.ToUpper(text)
	return strings.HasPrefix(upper, "IF ") && strings.Contains(upper, " THEN ")
}

// REQIDUniquenessRule checks REQ ID uniqueness within SPEC
// Implements REQ-SPC-003-004
type REQIDUniquenessRule struct{}

func (r *REQIDUniquenessRule) Code() string { return "DuplicateREQID" }

func (r *REQIDUniquenessRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	var findings []Finding
	seen := make(map[string]int) // ID → first occurrence line

	for _, req := range doc.REQs {
		sev, adv := reqFindingSeverity(req, SeverityError)
		if !reqIDPattern.MatchString(req.ID) {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: sev,
				Advisory: adv,
				Code:     "InvalidREQID",
				Message:  fmt.Sprintf("REQ ID %q does not match pattern REQ-<DOMAIN>[-<DOMAIN>]-NNN[-NNN] (each DOMAIN segment starts with an uppercase letter; each numeric group is exactly three digits)", req.ID),
			})
			continue
		}
		if firstLine, exists := seen[req.ID]; exists {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: sev,
				Advisory: adv,
				Code:     "DuplicateREQID",
				Message:  fmt.Sprintf("REQ ID %q is duplicated (first occurrence: line %d)", req.ID, firstLine),
			})
		} else {
			seen[req.ID] = req.Line
		}
	}
	return findings
}

// CoverageRule checks AC→REQ coverage
// Implements REQ-SPC-003-005
type CoverageRule struct{}

func (r *CoverageRule) Code() string { return "CoverageIncomplete" }

// Severity is `warning` with `Advisory: true` set at the EMISSION SITE, NEVER
// `error` (SPEC-COVERAGE-RULE-SCOPE-001 M3, plan.md §D option A).
//
// Before M3 this rule was an `error` that fired 0 times on the live corpus —
// not because the corpus was covered, but because the narrow reqLinePattern
// collected REQ definition lines from 16 of 704 spec.md files, so the rule's
// `len(doc.REQs) == 0` early return took almost every document. Wiring the
// widened collector turns the same rule on across 47 SPECs and 846 uncovered
// REQs. Landing that as `error` would redden the corpus on the first run and
// make bulk suppression the rational response — the outcome this SPEC exists to
// prevent.
//
// The mechanism is deliberately the emission site, NOT eraDemotableCodes. That
// map is consulted only for `SeverityError` findings, so a warning can never
// reach it, and the findings sit on modern-era SPECs no era path would demote.
// MovingRefUnpinnedRule (lint_movingref.go) reached the same emission-site
// conclusion by the same route; ArtifactStatusFieldForbiddenRule sits outside
// the same map for the OPPOSITE reason (it emits `error`, so the map WOULD
// reach it, and staying out is affordable only because its corpus cleanup lands
// alongside). Reading either as precedent for the other inverts both.
//
// THE DEBT: this guard now declares without enforcing. The promotion condition
// is that the ~846 corpus findings are remediated or exempted, after which the
// severity returns to `error`. That condition is prose, and prose does not
// fire — nothing in this file will notice when it is met. A guard left advisory
// past its promotion point is a check whose non-execution is indistinguishable
// from its success, which is precisely the defect class this SPEC was opened to
// document. MovingRefUnpinnedRule is the FIRST rule sleeping on a prose
// promotion condition in this package; this is the second.
func (r *CoverageRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	if len(doc.REQs) == 0 {
		return nil
	}

	// The covered set is the UNION of the inline AC section and the sibling
	// acceptance.md, which is the AC SSOT for Tier M/L. See
	// lint_coverage_sibling.go for why the sibling is read here rather than
	// merged into doc.Criteria, and why it is read whole.
	covered := collectAllREQIDs(doc.Criteria)
	for id := range siblingAcceptanceCoveredREQIDs(doc.Path) {
		covered[id] = true
	}

	var findings []Finding
	for _, req := range doc.REQs {
		if !covered[req.ID] {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: SeverityWarning,
				Advisory: true, // reports, never gates — see the rule doc above
				Code:     "CoverageIncomplete",
				Message:  fmt.Sprintf("REQ %s is not referenced by any AC", req.ID),
			})
		}
	}
	return findings
}

// FrontmatterSchemaRule validates the SPEC frontmatter schema.
// It emits a FrontmatterInvalid finding when any of the canonical 12 fields
// (id, title, version, status, created, updated, author, priority, phase,
// module, lifecycle, tags) is missing.
//
// snake_case aliases (created_at, updated_at, labels) do not match the yaml tags
// on the SPECFrontmatter struct, so the YAML decoder ignores them; they end up
// as empty strings and produce the same finding.
//
// SSOT: .claude/rules/moai/development/spec-frontmatter-schema.md
// Implements REQ-SPC-003-006
type FrontmatterSchemaRule struct{}

func (r *FrontmatterSchemaRule) Code() string { return "FrontmatterInvalid" }

// specIDPattern is a regular expression to validate SPEC ID format
var specIDPattern = regexp.MustCompile(`^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`)

var semverPattern = regexp.MustCompile(`^\d+\.\d+\.\d+`)

// phaseWorkflowStageTokens are the workflow-stage tokens that MUST NOT appear as a
// `phase:` value. The schema defines phase as a release target (e.g. "v3.0.2"), not
// a lifecycle-stage field, so a bare stage token there is an authoring mistake.
//
// Matching is exact equality on the trimmed, case-folded value — deliberately NOT
// substring containment, which would false-flag legitimate release targets that
// carry a token inside a longer word ("Run" within "Runtime Hardening"), and
// deliberately NOT a version-shape allowlist, which would false-flag the many
// legitimate free-form targets in the existing corpus. "mx" covers the retired
// fourth stage and adds no false positives.
//
// The finding this drives (FrontmatterPhaseInvalid) is deliberately absent from
// eraDemotableCodes: the guard must fire at authoring time, when the era heuristic
// classifies almost every in-flight SPEC as grandfathered. Registering it there
// would demote it to an advisory warning on exactly the SPECs it exists to catch.
var phaseWorkflowStageTokens = map[string]bool{
	"plan": true,
	"run":  true,
	"sync": true,
	"mx":   true,
}

func (r *FrontmatterSchemaRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	fm := doc.Frontmatter
	var findings []Finding

	required := []struct {
		name  string
		value string
	}{
		{"id", fm.ID},
		{"title", fm.Title},
		{"version", fm.Version},
		{"status", fm.Status},
		{"created", fm.Created},
		{"updated", fm.Updated},
		{"author", fm.Author},
		{"priority", fm.Priority},
		{"phase", fm.Phase},
		{"module", fm.Module},
		{"lifecycle", fm.Lifecycle},
		{"tags", fm.Tags},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityError,
				Code:     "FrontmatterInvalid",
				Message:  fmt.Sprintf("Frontmatter required field missing: %s", field.name),
			})
		}
	}

	if fm.ID != "" && !specIDPattern.MatchString(fm.ID) {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "FrontmatterInvalid",
			Message:  fmt.Sprintf("id %q does not match SPEC-<PREFIX>-<DOMAIN>-<NNN> format", fm.ID),
		})
	}

	// version semantic version verification
	if fm.Version != "" && !semverPattern.MatchString(fm.Version) {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "FrontmatterInvalid",
			Message:  fmt.Sprintf("version %q does not match semantic version format (X.Y.Z)", fm.Version),
		})
	}

	// phase value shape: a workflow-stage token is not a release target.
	// Placed after the required-field emptiness loop above so an empty phase yields
	// only the required-field finding and never a duplicate here.
	if phase := strings.TrimSpace(fm.Phase); phase != "" && phaseWorkflowStageTokens[strings.ToLower(phase)] {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "FrontmatterPhaseInvalid",
			Message: fmt.Sprintf(
				"phase %q is a workflow-stage token, not a release target; use the target release version (e.g. \"v3.0.2\")",
				fm.Phase),
		})
	}

	return findings
}

// DependencyExistsRule checks if SPECs in dependencies field actually exist
// Implements REQ-SPC-003-007
type DependencyExistsRule struct{}

func (r *DependencyExistsRule) Code() string { return "MissingDependency" }

func (r *DependencyExistsRule) Check(doc *SPECDoc, all []*SPECDoc) []Finding {
	if len(doc.Frontmatter.Dependencies) == 0 {
		return nil
	}

	knownIDs := make(map[string]bool, len(all))
	for _, d := range all {
		if d.Frontmatter.ID != "" {
			knownIDs[d.Frontmatter.ID] = true
		}
	}

	var findings []Finding
	for _, dep := range doc.Frontmatter.Dependencies {
		if knownIDs[dep] {
			continue
		}

		// Search based on doc.Path's parent directory
		docDir := filepath.Dir(filepath.Dir(doc.Path))
		depDir := filepath.Join(docDir, dep)
		depSpec := filepath.Join(depDir, "spec.md")
		if _, err := os.Stat(depSpec); os.IsNotExist(err) {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityError,
				Code:     "MissingDependency",
				Message:  fmt.Sprintf("Dependency SPEC %q not found", dep),
			})
		}
	}
	return findings
}

// OutOfScopeRule checks existence of "Out of Scope" section
// Implements REQ-SPC-003-009
type OutOfScopeRule struct{}

func (r *OutOfScopeRule) Code() string { return "MissingExclusions" }

func (r *OutOfScopeRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	body := strings.ToLower(doc.Body)
	// "out of scope" or "2.2 out of scope" pattern
	hasOutOfScope := strings.Contains(body, "out of scope")
	if !hasOutOfScope {
		return []Finding{{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "MissingExclusions",
			Message:  "'Out of Scope' section missing — minimum one item in Out of Scope subsection required",
		}}
	}

	lines := strings.Split(doc.Body, "\n")
	inOutOfScope := false
	hasContent := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lowerTrimmed := strings.ToLower(trimmed)

		if strings.HasPrefix(lowerTrimmed, "###") && strings.Contains(lowerTrimmed, "out of scope") {
			inOutOfScope = true
			continue
		}
		if strings.HasPrefix(lowerTrimmed, "##") && !strings.Contains(lowerTrimmed, "out of scope") && inOutOfScope {
			break
		}
		if inOutOfScope && strings.HasPrefix(trimmed, "-") && len(strings.TrimPrefix(trimmed, "-")) > 0 {
			hasContent = true
			break
		}
	}

	if !hasContent {
		return []Finding{{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "MissingExclusions",
			Message:  "'Out of Scope' section has no items — minimum one item required",
		}}
	}

	return nil
}

// BreakingChangeIDRule reports error when breaking:true but bc_id is empty
// Implements REQ-SPC-003-052
type BreakingChangeIDRule struct{}

func (r *BreakingChangeIDRule) Code() string { return "BreakingChangeMissingID" }

func (r *BreakingChangeIDRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	fm := doc.Frontmatter
	var findings []Finding

	if fm.Breaking && len(fm.BcID) == 0 {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "BreakingChangeMissingID",
			Message:  "breaking: true but bc_id is empty — breaking change requires bc_id",
		})
	}

	if !fm.Breaking && len(fm.BcID) > 0 {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityWarning,
			Code:     "OrphanBCID",
			Message:  "breaking: false이지만 bc_id가 비어 있지 않음 — orphan breaking change ID",
		})
	}

	return findings
}

// StatusValueEnumRule checks if status value is in the canonical 8-value enum
// Implements REQ-STATUS-LIFECYCLE-001-03.1
type StatusValueEnumRule struct{}

func (r *StatusValueEnumRule) Code() string { return "StatusValueInvalid" }

func (r *StatusValueEnumRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	fm := doc.Frontmatter
	var findings []Finding

	if fm.Status == "" {
		// Empty status is handled by FrontmatterSchemaRule
		return nil
	}

	if !IsValidStatus(fm.Status) {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "StatusValueInvalid",
			Message:  fmt.Sprintf("status %q is not in the canonical 8-value enum %v", fm.Status, ValidStatuses),
		})
	}

	return findings
}

// StatusCaseNormalizationRule checks if status value contains uppercase letters
// Implements REQ-STATUS-LIFECYCLE-001-03.2
type StatusCaseNormalizationRule struct{}

func (r *StatusCaseNormalizationRule) Code() string { return "StatusCaseInvalid" }

func (r *StatusCaseNormalizationRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	fm := doc.Frontmatter
	var findings []Finding

	if fm.Status == "" {
		// Empty status is handled by FrontmatterSchemaRule
		return nil
	}

	// Check if status contains uppercase
	if fm.Status != strings.ToLower(fm.Status) {
		lowerStatus := strings.ToLower(fm.Status)
		// Only report error if the lowercased version is valid
		if IsValidStatus(lowerStatus) {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityError,
				Code:     "StatusCaseInvalid",
				Message:  fmt.Sprintf("status %q contains uppercase; use lowercase %q instead", fm.Status, lowerStatus),
			})
		}
	}

	return findings
}

// ZoneRegistryRule checks if CONST-V3R2-NNN references in related_rule field exist in zone registry
// Implements REQ-SPC-003-010
type ZoneRegistryRule struct {
	registry *constitution.Registry
}

func (r *ZoneRegistryRule) Code() string { return "DanglingRuleReference" }

func (r *ZoneRegistryRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	if r.registry == nil || len(doc.Frontmatter.RelatedRule) == 0 {
		return nil
	}

	var findings []Finding
	for _, ruleID := range doc.Frontmatter.RelatedRule {
		if _, ok := r.registry.Get(ruleID); !ok {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityWarning,
				Code:     "DanglingRuleReference",
				Message:  fmt.Sprintf("related_rule %q not found in zone registry", ruleID),
			})
		}
	}
	return findings
}

// DependencyCycleRule detects cycles in SPEC dependency DAG
// Implements REQ-SPC-003-008
// Implements crossSPECRule interface, executed in cross-SPEC phase of Linter.Lint
type DependencyCycleRule struct{}

func (r *DependencyCycleRule) Code() string { return "DependencyCycle" }

func (r *DependencyCycleRule) Check(_ *SPECDoc, _ []*SPECDoc) []Finding {
	// single-spec check not used; processed in CheckAll
	return nil
}

func (r *DependencyCycleRule) CheckAll(docs []*SPECDoc) []Finding {
	idToIdx := make(map[string]int, len(docs))
	for i, doc := range docs {
		if doc.Frontmatter.ID != "" {
			idToIdx[doc.Frontmatter.ID] = i
		}
	}

	adj := make([][]int, len(docs))
	for i, doc := range docs {
		for _, dep := range doc.Frontmatter.Dependencies {
			if j, ok := idToIdx[dep]; ok {
				adj[i] = append(adj[i], j)
			}
		}
	}

	// Cycle detection via Tarjan SCC
	cycles := findCyclesTarjan(adj, len(docs))

	if len(cycles) == 0 {
		return nil
	}

	var findings []Finding
	for _, cycle := range cycles {
		names := make([]string, 0, len(cycle))
		for _, idx := range cycle {
			names = append(names, docs[idx].Frontmatter.ID)
			if names[len(names)-1] == "" {
				names[len(names)-1] = docs[idx].Path
			}
		}
		findings = append(findings, Finding{
			File:     docs[cycle[0]].Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "DependencyCycle",
			Message:  fmt.Sprintf("Dependency cycle detected: %s", strings.Join(names, " → ")),
		})
	}
	return findings
}

// DuplicateSPECIDRule checks if multiple SPECs declare the same id
// Implements REQ-SPC-003-031
// Implements crossSPECRule interface
type DuplicateSPECIDRule struct{}

func (r *DuplicateSPECIDRule) Code() string { return "DuplicateSPECID" }

func (r *DuplicateSPECIDRule) Check(_ *SPECDoc, _ []*SPECDoc) []Finding {
	return nil
}

func (r *DuplicateSPECIDRule) CheckAll(docs []*SPECDoc) []Finding {
	seen := make(map[string]string) // ID → first file path
	var findings []Finding

	for _, doc := range docs {
		id := doc.Frontmatter.ID
		if id == "" {
			continue
		}
		if firstPath, exists := seen[id]; exists {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityError,
				Code:     "DuplicateSPECID",
				Message:  fmt.Sprintf("SPEC ID %q declared multiple times (first location: %s)", id, firstPath),
			})
		} else {
			seen[id] = doc.Path
		}
	}
	return findings
}

// terminalStatusEnum is the set of lifecycle terminal states for which a mismatch with the git-implied status is normal.
// SPECs in a terminal state may no longer have active work commits in git history, and
// such cases must not be treated as a drift false positive.
//
// @MX:NOTE: [AUTO] terminal lifecycle state — status values where a mismatch with git history is considered normal
// @MX:REASON: resolves Pattern D/E/F/G false positives from SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001; extend this map only when adding future states
var terminalStatusEnum = map[string]bool{
	"superseded": true,
	"archived":   true,
	"rejected":   true, // future-proof: not currently used, kept for future extension
	"completed":  true, // V3R6 3-phase lifecycle terminal state (reached via the sync commit)
}

// StatusGitConsistencyRule checks if SPEC frontmatter status agrees with git log
// Implements Round 3: W3-T4
// Severity: advisory warning — the git-implied status is a heuristic over the
// available git history (shallow CI checkouts routinely disagree), so this
// finding is surfaced but never escalated to error under --strict.
type StatusGitConsistencyRule struct{}

func (r *StatusGitConsistencyRule) Code() string { return "StatusGitConsistency" }

func (r *StatusGitConsistencyRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	fm := doc.Frontmatter
	var findings []Finding

	if fm.ID == "" || fm.Status == "" {
		// Skip if ID or status is missing (handled by other rules)
		return nil
	}

	// ★ terminal lifecycle states are expected to mismatch the git-implied status — prevents false positives
	// Covers Pattern D (superseded/completed), E (superseded/implemented),
	// F (archived/implemented), and G (archived/in-progress).
	if terminalStatusEnum[fm.Status] {
		return nil
	}

	// Get git-implied status
	gitStatus, err := getGitImpliedStatus(fm.ID)
	if err != nil {
		// Observation failed. "Not observed" and "observed-and-matching" must
		// not share one output (REQ-SLGB-001, SPEC-SPECLINT-GITBLIND-001):
		// where the failure means the git signal is unobservable for this
		// repository, surface the Info finding instead of silently skipping.
		// Either way there is no StatusGitConsistency verdict — there is no
		// gitStatus to compare against.
		if gitObservationUnreachable(err) && takeUnreachableEmission() {
			return []Finding{statusGitUnreachableFinding(doc, err)}
		}
		// Observed-and-harmless failure (shapes ②/③ in a full repository),
		// or a repeat occurrence after this run's single emission (§2.2).
		return nil
	}

	// Check for drift
	if fm.Status != gitStatus {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityWarning,
			Advisory: true, // heuristic git-implied signal — never strict-escalated
			Code:     "StatusGitConsistency",
			Message:  fmt.Sprintf("SPEC %s frontmatter status '%s' disagrees with git-implied status '%s'", fm.ID, fm.Status, gitStatus),
		})
	}

	return findings
}

// gitObservationUnreachable decides whether a getGitImpliedStatus error
// means the git signal is UNOBSERVED for this repository, as opposed to
// observed-and-harmless (SPEC-SPECLINT-GITBLIND-001 §2.1):
//   - shape ① (errGitQueryFailed): base ref unresolvable / git unusable — a
//     repository-level blindness that fires unconditionally;
//   - shapes ② and ③: harmless in a full repository (no lifecycle commits /
//     cosmetic-only commits), but in a shallow clone the truncated window can
//     fabricate them — observation failure only while shallow.
//
// The deciding predicate for ②/③ is repository-level (shallow state), so it
// rides the same per-run cache as cachedMainBranch: the shape decision never
// spawns a per-SPEC subprocess.
func gitObservationUnreachable(err error) bool {
	switch {
	case errors.Is(err, errGitQueryFailed):
		return true
	case errors.Is(err, errNoGitHistory), errors.Is(err, errNoClassifiableCommit):
		return cachedIsShallowRepository()
	default:
		return false
	}
}

// statusGitUnreachableFinding renders the Info-severity StatusGitUnreachable
// finding — the observability surface of SPEC-SPECLINT-GITBLIND-001 M1.
// REQ-SLGB-002: for the ref-resolution shape the message names the candidate
// base refs whose resolution was attempted. §2.2: the message states both the
// repository-wide scope and the resulting rule-wide skip explicitly, because
// the one finding stands in for every SPEC the rule could not observe.
// REQ-SLGB-005: Info severity, never changes the --strict exit code.
func statusGitUnreachableFinding(doc *SPECDoc, err error) Finding {
	detail := "shallow clone window makes the git signal unreliable"
	if errors.Is(err, errGitQueryFailed) {
		detail = fmt.Sprintf("no usable base ref (tried: %s)", triedBaseRefsSummary())
	}
	return Finding{
		File:     doc.Path,
		Line:     1,
		Severity: SeverityInfo,
		Code:     "StatusGitUnreachable",
		Message: fmt.Sprintf(
			"SPEC %s git status NOT OBSERVED — %s; this condition is repository-wide: StatusGitConsistency is skipped for every SPEC in this lint run (%v)",
			doc.Frontmatter.ID, detail, err),
	}
}

// triedBaseRefsSummary names the base refs the resolution chain consults, for
// the REQ-SLGB-002 message. Sourced from the single mainBranchCandidates
// chain (gitquery_cache.go) so the message and the walk cannot drift apart.
func triedBaseRefsSummary() string {
	return strings.Join(mainBranchCandidates, ", ")
}
