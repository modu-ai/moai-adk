// Package spec provides SPEC document parsing and validation functionality
// lint.go is the core engine of moai spec lint CLI, validating SPEC documents
// for EARS compliance, coverage, DAG, etc. through Rule interface and Linter struct
package spec

import (
	"encoding/json"
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
		if r.Strict && f.Severity == SeverityWarning {
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

//
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

	l.rules = []Rule{
		&EARSModalityRule{},
		&REQIDUniquenessRule{},
		&CoverageRule{},
		&FrontmatterSchemaRule{},
		&DependencyExistsRule{},
		&OutOfScopeRule{},
		&BreakingChangeIDRule{},
		// cross-SPEC rules
		&DependencyCycleRule{},
		&DuplicateSPECIDRule{},
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
	if len(paths) == 0 {
		discovered, err := discoverSPECs(l.opts.BaseDir)
		if err != nil {
			return nil, fmt.Errorf("SPEC discovery failed: %w", err)
		}
		paths = discovered
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
		for _, rule := range l.rules {
			// cross-SPEC rules are processed later
			if _, ok := rule.(crossSPECRule); ok {
				continue
			}
			ruleFindings := rule.Check(doc, docs)
			ruleFindings = applylintSkip(ruleFindings, doc.LintSkip)
			findings = append(findings, ruleFindings...)
		}
	}

	for _, rule := range l.rules {
		if cr, ok := rule.(crossSPECRule); ok {
			crossFindings := cr.CheckAll(docs)
			findings = append(findings, crossFindings...)
		}
	}

	return &Report{
		Findings: findings,
		Strict:   l.opts.Strict,
	}, nil
}

type crossSPECRule interface {
	CheckAll(docs []*SPECDoc) []Finding
}

//
// @MX:NOTE: [AUTO] Rule interface inspects a single SPEC document
type Rule interface {
	Code() string
	// Check inspects a single SPEC document and returns findings
	Check(doc *SPECDoc, all []*SPECDoc) []Finding
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
}

type REQEntry struct {
	ID   string
	Text string
	Line int
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

// reqIDPattern is a regular expression to validate REQ-<DOMAIN>-<NNN>-<NNN> format
var reqIDPattern = regexp.MustCompile(`^REQ-[A-Z]{2,5}-\d{3}-\d{3}$`)

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

	// Parse REQ list
	doc.REQs = parseREQs(body)

	// Parse Acceptance Criteria
	criteria, _ := ParseAcceptanceCriteria(body, false)
	doc.Criteria = criteria

	return doc
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
// Implements REQ-SPC-003-003, REQ-SPC-003-050
type EARSModalityRule struct{}

func (r *EARSModalityRule) Code() string { return "ModalityMalformed" }

func (r *EARSModalityRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	var findings []Finding
	for _, req := range doc.REQs {
		if isModalityMalformed(req.Text) {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: SeverityError,
				Code:     "ModalityMalformed",
				Message:  fmt.Sprintf("REQ %s: EARS modality violation — SHALL missing or format mismatch: %q", req.ID, req.Text),
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

// REQIDUniquenessRule checks REQ ID uniqueness within SPEC
// Implements REQ-SPC-003-004
type REQIDUniquenessRule struct{}

func (r *REQIDUniquenessRule) Code() string { return "DuplicateREQID" }

func (r *REQIDUniquenessRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	var findings []Finding
	seen := make(map[string]int) // ID → first occurrence line

	for _, req := range doc.REQs {
		if !reqIDPattern.MatchString(req.ID) {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: SeverityError,
				Code:     "InvalidREQID",
				Message:  fmt.Sprintf("REQ ID %q does not match pattern REQ-[A-Z]{{2,5}}-NNN-NNN", req.ID),
			})
			continue
		}
		if firstLine, exists := seen[req.ID]; exists {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: SeverityError,
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

func (r *CoverageRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	if len(doc.REQs) == 0 {
		return nil
	}

	covered := collectAllREQIDs(doc.Criteria)

	var findings []Finding
	for _, req := range doc.REQs {
		if !covered[req.ID] {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: SeverityError,
				Code:     "CoverageIncomplete",
				Message:  fmt.Sprintf("REQ %s is not referenced by any AC", req.ID),
			})
		}
	}
	return findings
}

// FrontmatterSchemaRule checks SPEC frontmatter schema
// Implements REQ-SPC-003-006
type FrontmatterSchemaRule struct{}

func (r *FrontmatterSchemaRule) Code() string { return "FrontmatterInvalid" }

// specIDPattern is a regular expression to validate SPEC ID format
var specIDPattern = regexp.MustCompile(`^SPEC-[A-Z][A-Z0-9]+-[A-Z]{2,5}-\d{3}$`)

var semverPattern = regexp.MustCompile(`^\d+\.\d+\.\d+`)

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
