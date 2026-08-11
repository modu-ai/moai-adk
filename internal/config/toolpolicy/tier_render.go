package toolpolicy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config/atomicfile"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// SPEC-AUTONOMY-TIERS-001 M3 — tier → permission-bundle renderer.
//
// This renderer is a NEW CALLER of the existing toolpolicy codegen
// (BuildPermissions + RenderSettingsJSON), NOT a parallel writer (AP-2). It
// splits the tier bundle across the two Claude Code config scopes per
// REQ-003 / spec.md §C:
//   - USER scope (~/.claude/settings.json): permissions.defaultMode.
//     auto/bypassPermissions are USER-scope-only per Claude Code v2.1.142+
//     (PROJECT/local cannot grant them downward) — writing them to PROJECT
//     would silently fail (AP-6).
//   - PROJECT scope (<project>/.claude/settings.json): permissions.deny +
//     permissions.ask. These merge across all sessions and are tier-INVARIANT
//     (REQ-004) — the deny/ask rule set is identical at every tier.
//
// The caller resolves the tier → defaultMode (via config.TierDefaultMode +
// config.EffectiveTierWithGates) and passes the resolved defaultMode string.
// This keeps the tier logic in internal/config and the rendering here, with no
// import cycle (config does not import toolpolicy).

// @MX:NOTE: [AUTO] tier→permission-bundle renderer — USER scope (defaultMode) + PROJECT scope (deny/ask) split per REQ-003 / v2.1.142+ scope rules

// TierRenderResult reports the two files written by RenderTierPermissions.
type TierRenderResult struct {
	UserPath    string
	ProjectPath string
}

// RenderTierPermissions writes the tier bundle across the two config scopes:
//
//   - userSettingsPath (USER scope, typically ~/.claude/settings.json):
//     permissions.defaultMode is set to defaultMode. Existing deny/ask arrays
//     in the USER file are CLEARED (deny/ask are PROJECT-scoped per REQ-003;
//     leaving a stale USER deny/ask would shadow the PROJECT rule set).
//
//   - projectSettingsPath (PROJECT scope, <project>/.claude/settings.json):
//     permissions.deny + permissions.ask are regenerated from doc. The
//     defaultMode in PROJECT is reset to "default" (or removed) — auto /
//     bypassPermissions MUST NOT land in PROJECT (AP-6 / v2.1.142+).
//
// deny/ask are tier-INVARIANT (AC-004): doc is loaded once and the same deny/
// ask arrays are written regardless of defaultMode. semi-auto (defaultMode=
// "default") produces byte-identical output to today's template (AC-007).
func RenderTierPermissions(projectSettingsPath, userSettingsPath, defaultMode string, doc *PolicyDocument) (*TierRenderResult, error) {
	if doc == nil {
		return nil, fmt.Errorf("nil policy document")
	}

	// Build the full permissions block once (defaultMode + deny + ask), then
	// split it across the two scopes.
	full, _, err := BuildPermissions(doc, defaultMode)
	if err != nil {
		return nil, fmt.Errorf("build permissions: %w", err)
	}

	// USER scope: defaultMode ONLY. deny/ask are PROJECT-scoped, so a USER
	// block carries just defaultMode (plus any preserved Raw extras).
	userBlock := &PermissionsBlock{
		DefaultMode: full.DefaultMode,
		Raw:         map[string]json.RawMessage{},
	}

	// PROJECT scope: deny + ask ONLY. defaultMode is reset to "" so auto/
	// bypassPermissions never land here (AP-6). Preserve Raw extras.
	projectBlock := &PermissionsBlock{
		Allow: full.Allow, // allow stays with the project allowlist
		Ask:   full.Ask,
		Deny:  full.Deny,
		Raw:   map[string]json.RawMessage{},
	}

	if err := renderIntoFile(userSettingsPath, userBlock); err != nil {
		return nil, fmt.Errorf("USER scope render: %w", err)
	}
	if err := renderIntoFile(projectSettingsPath, projectBlock); err != nil {
		return nil, fmt.Errorf("PROJECT scope render: %w", err)
	}

	return &TierRenderResult{UserPath: userSettingsPath, ProjectPath: projectSettingsPath}, nil
}

// WriteUserDefaultMode splices ONLY the defaultMode into the USER-scope
// settings.json at userPath, preserving every other region (PATH, hooks, env,
// allow, deny, ask). It is the init-time fallback used when no tool-policy.yaml
// is deployed: the deny/ask arrays already ship in the PROJECT template, so the
// only missing piece is the USER-scope defaultMode that the selected tier maps
// to. The codegen reuses extractPermissions + RenderSettingsJSON (the same
// rendering pipeline RenderTierPermissions uses), so this is NOT a parallel
// writer — it is a scoped entry point into the same splice logic.
//
// If the file does not exist, a minimal permissions skeleton is created so the
// region matcher has a target (mirrors renderIntoFile's create-on-absent path).
// The parent directory is created on demand so the init path can write the
// USER-scope ~/.claude/settings.json even when the directory does not yet exist.
func WriteUserDefaultMode(userPath, defaultMode string) error {
	if err := os.MkdirAll(filepath.Dir(userPath), 0o755); err != nil {
		return fmt.Errorf("create user settings dir: %w", err)
	}
	block := &PermissionsBlock{
		DefaultMode: defaultMode,
		Raw:         map[string]json.RawMessage{},
	}
	return renderIntoFile(userPath, block)
}

// renderIntoFile reads the target settings file, splices block into its
// permissions region, and writes it back. It preserves all non-permissions
// regions (PATH, hooks, env, Go-template directives). If the file does not
// exist, it is created with a minimal skeleton.
func renderIntoFile(path string, block *PermissionsBlock) error {
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create a minimal skeleton so the region matcher has a target.
			// Ensure the parent directory exists so the subsequent WriteFile
			// succeeds (the USER-scope ~/.claude/ dir may not exist yet at
			// init time).
			if mkErr := os.MkdirAll(filepath.Dir(path), 0o755); mkErr != nil {
				return fmt.Errorf("create dir for %q: %w", path, mkErr)
			}
			body = []byte("{\n  \"permissions\": {\n  }\n}")
		} else {
			return fmt.Errorf("read %q: %w", path, err)
		}
	}

	// Preserve any extra Raw keys from the existing block (additionalDirectories
	// etc.) so the codegen does not drop non-list permission settings.
	existing, extrErr := extractPermissions(body)
	if extrErr == nil && existing != nil {
		block.Raw = existing.Raw
	}

	out, err := RenderSettingsJSON(body, block)
	if err != nil {
		return fmt.Errorf("render %q: %w", path, err)
	}
	if err := atomicfile.Write(path, out, defs.FilePerm); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}
