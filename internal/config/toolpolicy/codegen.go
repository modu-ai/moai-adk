package toolpolicy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CodegenResult reports the outcome of a BuildInto call: which target file
// was written, how many entries were emitted into each permissions list, and
// whether the file's template directive was preserved (for .tmpl targets).
type CodegenResult struct {
	Path           string
	AllowEmitted   int
	AskEmitted     int
	DenyEmitted    int
	EnvGatedSkipped int // entries with env_gate — emitted as audit only, not static permissions
	TargetKind     TargetKind
}

// TargetKind distinguishes the two codegen strategies (design.md §B.1 / D7).
type TargetKind string

const (
	// TargetJSON is a pure-JSON settings file (the local .claude/settings.json).
	// Strategy: parse-modify-serialize.
	TargetJSON TargetKind = "json"
	// TargetTemplate is a mixed JSON + Go-template-directive settings file
	// (internal/template/templates/.claude/settings.json.tmpl). Strategy:
	// raw-text region replacement (the file cannot be JSON-parsed because of
	// the {{jsonEscape .SmartPATH}} directive in the env block).
	TargetTemplate TargetKind = "tmpl"
)

// BuildPermissions renders the YAML entries into a Claude Code permissions
// object (the `{ "allow": [...], "deny": [...], "ask": [...] }` value).
//
// Entries carrying an EnvGate are NOT emitted as static permission specifiers
// — a static deny would over-block the cg-leader exception. They are counted
// in the result as EnvGatedSkipped and recorded only in the YAML audit
// surface (design.md §C.4). The codegen does not emit a hook matcher for
// env-gated rules in this SPEC (OQ-3 deferred).
//
// The output is deterministic: entries are sorted by specifier, and within a
// list no duplicates are emitted. This is the load-bearing property for
// AC-TPS-013 (idempotency — BuildPermissions(YAML) == BuildPermissions(YAML)
// byte-for-byte across runs).
func BuildPermissions(doc *PolicyDocument, defaultMode string) (*PermissionsBlock, CodegenResult, error) {
	if doc == nil {
		return nil, CodegenResult{}, fmt.Errorf("nil policy document")
	}
	block := &PermissionsBlock{
		DefaultMode: defaultMode,
		Raw:         map[string]json.RawMessage{},
	}
	seenAllow := map[string]bool{}
	seenDeny := map[string]bool{}
	seenAsk := map[string]bool{}
	var res CodegenResult

	for _, e := range doc.SortedBySpecifier() {
		if e.EnvGate != nil {
			// Env-gated rules are audit-only (the cg-leader exception would be
			// over-blocked by a static deny). design.md §C.4 / OQ-3 deferred.
			res.EnvGatedSkipped++
			continue
		}
		spec := e.SettingsSpecifier()
		switch e.Decision {
		case DecisionAllow:
			if !seenAllow[spec] {
				block.Allow = append(block.Allow, spec)
				seenAllow[spec] = true
				res.AllowEmitted++
			}
		case DecisionDeny:
			if !seenDeny[spec] {
				block.Deny = append(block.Deny, spec)
				seenDeny[spec] = true
				res.DenyEmitted++
			}
		case DecisionAsk:
			if !seenAsk[spec] {
				block.Ask = append(block.Ask, spec)
				seenAsk[spec] = true
				res.AskEmitted++
			}
		default:
			return nil, res, fmt.Errorf("entry tool=%q: unknown decision %q", e.Tool, e.Decision)
		}
	}
	return block, res, nil
}

// RenderSettingsJSON takes the current .claude/settings.json body and a
// regenerated permissions block, and returns the new file body with ONLY the
// permissions object replaced (parse-modify-serialize strategy — preserves
// all non-permissions regions: PATH, hooks, env, defaultMode at top level if
// any, etc.). Used for the pure-JSON target (TargetJSON).
//
// The permissions key is located via raw-text brace matching (so the strategy
// degrades gracefully if the file has trailing comments or unusual
// whitespace), then the new object is serialized and spliced in. The result
// is valid JSON.
func RenderSettingsJSON(body []byte, block *PermissionsBlock) ([]byte, error) {
	region, err := locatePermissionsRegion(body)
	if err != nil {
		return nil, err
	}
	rendered, err := renderPermissionsObject(block)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.Write(body[:region.start])
	out.WriteString(`"permissions": `)
	out.Write(rendered)
	out.Write(body[region.end:])
	return out.Bytes(), nil
}

// RenderSettingsTmpl takes the current settings.json.tmpl body and a
// regenerated permissions block, and returns the new file body with ONLY the
// permissions object replaced via raw-text region replacement. The rest of
// the file (including Go-template directives like {{jsonEscape .SmartPATH}}
// in the env block) is preserved verbatim because the region matcher only
// touches the "permissions": { ... } span (design.md §B.1 / D7).
//
// Post-condition (AC-TPS-014 sentinel): the regenerated permissions block
// MUST contain zero `{{` or `}}` substrings. AssertPermissionBlockNoTemplate
// verifies this after rendering.
func RenderSettingsTmpl(body []byte, block *PermissionsBlock) ([]byte, error) {
	region, err := locatePermissionsRegion(body)
	if err != nil {
		return nil, err
	}
	rendered, err := renderPermissionsObject(block)
	if err != nil {
		return nil, err
	}
	if err := AssertPermissionBlockNoTemplate(rendered); err != nil {
		return nil, fmt.Errorf("codegen post-condition violated (template directive leaked into permissions block): %w", err)
	}
	var out bytes.Buffer
	out.Write(body[:region.start])
	out.WriteString(`"permissions": `)
	out.Write(rendered)
	out.Write(body[region.end:])
	return out.Bytes(), nil
}

// AssertPermissionBlockNoTemplate enforces the AC-TPS-014 sentinel: the
// regenerated permissions block must contain zero `{{` or `}}` substrings.
// If a template directive accidentally lands inside the permissions block,
// the codegen is buggy.
func AssertPermissionBlockNoTemplate(rendered []byte) error {
	if bytes.Contains(rendered, []byte("{{")) || bytes.Contains(rendered, []byte("}}")) {
		return fmt.Errorf("permissions block contains template directive: %s", truncateForError(rendered, 200))
	}
	return nil
}

func truncateForError(b []byte, max int) string {
	s := string(b)
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

// BuildInto runs the codegen against one target file, writing the result
// back in place. The targetKind selects the strategy (JSON parse-modify-
// serialize vs .tmpl raw-text region replacement). The defaultMode is
// preserved from the existing block unless overridden.
func BuildInto(path string, doc *PolicyDocument, targetKind TargetKind, defaultModeOverride string) (*CodegenResult, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("codegen read %q: %w", path, err)
	}

	// Preserve the existing defaultMode unless the caller overrides it. This
	// respects the documented dev-vs-template split (CLAUDE.local.md §22.1 —
	// local defaultMode=bypassPermissions, template=acceptEdits are both
	// intentional divergences the codegen MUST NOT normalize).
	existing, err := extractPermissions(body)
	if err != nil {
		return nil, fmt.Errorf("codegen extract existing permissions %q: %w", path, err)
	}
	defaultMode := existing.DefaultMode
	if defaultModeOverride != "" {
		defaultMode = defaultModeOverride
	}

	block, res, err := BuildPermissions(doc, defaultMode)
	if err != nil {
		return nil, err
	}
	// Preserve any extra keys (additionalDirectories etc.) from the existing
	// block so the codegen does not drop non-list permission settings.
	block.Raw = existing.Raw

	var out []byte
	switch targetKind {
	case TargetJSON:
		out, err = RenderSettingsJSON(body, block)
	case TargetTemplate:
		out, err = RenderSettingsTmpl(body, block)
	default:
		return nil, fmt.Errorf("unknown target kind %q", targetKind)
	}
	if err != nil {
		return nil, fmt.Errorf("codegen render %q: %w", path, err)
	}

	if err := os.WriteFile(path, out, 0o644); err != nil {
		return nil, fmt.Errorf("codegen write %q: %w", path, err)
	}
	res.Path = path
	res.TargetKind = targetKind
	return &res, nil
}

// DetectTargetKind infers whether a settings file is a pure-JSON target or a
// mixed Go-template .tmpl target by checking for any `{{` directive in the
// body. Used by the CLI when the caller does not pass an explicit kind.
func DetectTargetKind(body []byte) TargetKind {
	if bytes.Contains(body, []byte("{{")) {
		return TargetTemplate
	}
	return TargetJSON
}

// BuildIntoAuto detects the target kind from the file contents and dispatches
// to the appropriate strategy. Convenience for the CLI.
func BuildIntoAuto(path string, doc *PolicyDocument, defaultModeOverride string) (*CodegenResult, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("codegen read %q: %w", path, err)
	}
	kind := DetectTargetKind(body)
	return BuildInto(path, doc, kind, defaultModeOverride)
}

// SettingsPath returns the conventional settings.json path for a project
// root. Exported so the CLI and tests share one resolution.
func SettingsPath(projectDir string) string {
	return filepath.Join(projectDir, ".claude", "settings.json")
}

// TemplateSettingsPath returns the conventional settings.json.tmpl path under
// the template source tree. Exported for symmetry with SettingsPath.
func TemplateSettingsPath(repoRoot string) string {
	return filepath.Join(repoRoot, "internal", "template", "templates", ".claude", "settings.json.tmpl")
}

// PolicyYAMLPath returns the conventional tool-policy.yaml path for a project.
func PolicyYAMLPath(projectDir string) string {
	return filepath.Join(projectDir, ".moai", "config", "sections", "tool-policy.yaml")
}

// DecisionFor looks up the decision declared for a given settings specifier
// (e.g., "Bash(git push:*)") in the document. Returns the decision and true
// if found, or "" and false otherwise. Used by the round-trip equivalence
// test (AC-TPS-004) to assert codegenDecision(entry) == entry.Decision for
// every seeded entry.
func (d *PolicyDocument) DecisionFor(specifier string) (Decision, bool) {
	for _, e := range d.Entries {
		if e.EnvGate != nil {
			continue // env-gated rules are not emitted as static specifiers
		}
		if e.SettingsSpecifier() == specifier {
			return e.Decision, true
		}
	}
	return "", false
}

// ensureNoCR normalizes \r\n to \n in a byte slice. The .tmpl file may have
// been authored on Windows; the codegen output uses \n exclusively for
// determinism (AC-TPS-013).
func ensureNoCR(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
}

// CountSpecifierByDecision tallies the entries (excluding env-gated) by
// decision. Used by the codegen result reporting and by tests.
func (d *PolicyDocument) CountSpecifierByDecision() map[Decision]int {
	out := map[Decision]int{}
	for _, e := range d.Entries {
		if e.EnvGate != nil {
			continue
		}
		out[e.Decision]++
	}
	return out
}

// splitSpecifier is the inverse of SettingsSpecifier: given "Tool" or
// "Tool(args)", returns (tool, args). Exported for the round-trip test.
func splitSpecifier(spec string) (tool, args string) {
	spec = strings.TrimSpace(spec)
	if !strings.HasSuffix(spec, ")") {
		return spec, ""
	}
	open := strings.Index(spec, "(")
	if open < 0 {
		return spec, ""
	}
	return spec[:open], spec[open+1 : len(spec)-1]
}
