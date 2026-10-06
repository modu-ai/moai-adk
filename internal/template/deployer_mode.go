package template

// deployer_mode.go — the deploy-mode split (SPEC-INIT-SHRINK-001 REQ-001 /
// REQ-003 / REQ-006, plan M2). One option on the existing deployer family —
// not a new deployer type:
//
//	plugin  — the default path: the deploy walk and ListTemplates carry no
//	          .claude/skills/** or .claude/commands/** file (the plugin is
//	          the carrier; REQ-001). The .agents/skills mirror follows the
//	          plugin mirror policy: MirrorPolicyNone deploys no mirror at
//	          all (taken only once Codex is verified to actually execute
//	          plugin-borne skills); MirrorPolicyRehome deploys the mirror
//	          with every entry a REAL DIRECTORY COPY rendered from the
//	          embedded tree — never a symlink into a .claude/skills the
//	          plugin path does not deploy (REQ-006 / OD-6 settled (a) +
//	          condition, the D-8 repair).
//	local   — the --no-plugin and --all paths: today's payload
//	          byte-for-byte (REQ-003 / REQ-007). The zero value is local,
//	          so every existing constructor call site is unchanged.
//
// Codex-only composition: the harnessFS re-homing already writes real
// directories under .agents/skills, so on plugin mode the same exclusion
// set applies — under MirrorPolicyNone the re-homed catalog is excluded
// too; under MirrorPolicyRehome it stays (real copies by construction,
// P-12).

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DeployMode selects the deploy file set.
type DeployMode string

const (
	// DeployModeLocal — the full local payload (today's deploy; the zero
	// value and every opt-out path).
	DeployModeLocal DeployMode = "local"
	// DeployModePlugin — the thin payload: the plugin carries skills and
	// commands.
	DeployModePlugin DeployMode = "plugin"
)

// PluginMirrorPolicy selects what plugin mode does with the .agents/skills
// mirror (REQ-006, OD-6 settled (a) + condition).
type PluginMirrorPolicy string

const (
	// MirrorPolicyNone — deploy no mirror at all. The post-verification
	// arm: taken only when Codex is verified to actually execute
	// plugin-borne skills.
	MirrorPolicyNone PluginMirrorPolicy = "none"
	// MirrorPolicyRehome — deploy the mirror with every entry a REAL
	// DIRECTORY COPY rendered from the embedded tree. The no-verification
	// fallback arm: never a symlink into a .claude/skills the plugin path
	// does not deploy.
	MirrorPolicyRehome PluginMirrorPolicy = "rehome"
)

// WithDeployMode sets the deploy mode. An empty value and DeployModeLocal
// are the same thing: today's full payload.
func WithDeployMode(mode DeployMode) DeployerOption {
	return func(d *deployer) { d.deployMode = mode }
}

// WithPluginMirrorPolicy sets the plugin-mode mirror policy. The zero value
// rehomes (the conservative OD-6 fallback), so an option-less plugin deploy
// never dangles a mirror link.
func WithPluginMirrorPolicy(policy PluginMirrorPolicy) DeployerOption {
	return func(d *deployer) { d.pluginMirrorPolicy = policy }
}

// isPluginExcludedPath reports whether a deploy-relative slash path is
// excluded from the plugin-mode payload: always the two dropped Claude
// roots; the .agents/skills tree only under MirrorPolicyNone (under rehome
// the walk itself writes the re-homed copies for a codex-only profile).
func isPluginExcludedPath(relPath string, policy PluginMirrorPolicy) bool {
	if strings.HasPrefix(relPath, ".claude/skills/") || strings.HasPrefix(relPath, ".claude/commands/") {
		return true
	}
	if policy == MirrorPolicyNone && strings.HasPrefix(relPath, ".agents/skills/") {
		return true
	}
	return false
}

// pluginModeExcluded reports whether the deployer's mode excludes the path.
func (d *deployer) pluginModeExcluded(relPath string) bool {
	if d.deployMode != DeployModePlugin {
		return false
	}
	policy := d.pluginMirrorPolicy
	if policy == "" {
		policy = MirrorPolicyRehome // conservative default (OD-6 fallback)
	}
	return isPluginExcludedPath(relPath, policy)
}

// isCommonAssetRoot reports whether a deploy-relative path lives under one
// of the project common-asset roots the user-folder install replaces
// (SPEC-USER-ASSET-INSTALL-001 REQ-005: the project payload carries no
// common skill or agent file in any mode).
func isCommonAssetRoot(relPath string) bool {
	for _, root := range []string{
		".claude/skills/",
		".claude/commands/moai/",
		".claude/agents/moai/",
		".agents/skills/",
		".codex/agents/moai/",
	} {
		if strings.HasPrefix(relPath, root) {
			return true
		}
	}
	return false
}

// stripMoaiFromMcpJSON removes the `moai` server from the mcpServers object
// of a rendered .mcp.json (SPEC-INIT-SHRINK-001 REQ-005, OD-1 settled (c)):
// on the plugin path the project render never carries the entry — the
// plugin is the carrier — and the post-install provision call writes it
// back as the fallback carrier when the probe reads not-demonstrated
// (design §2.3's render-time filter + sequenced provision call). context7
// and staggeredStartup are preserved byte-for-byte in meaning; the JSON is
// re-encoded with two-space indent. A malformed document returns the
// content unchanged — the deploy of a template defect fails elsewhere, and
// this filter never turns a render error into silent data loss.
func stripMoaiFromMcpJSON(content []byte) []byte {
	var doc map[string]any
	if err := json.Unmarshal(content, &doc); err != nil {
		return content
	}
	servers, ok := doc["mcpServers"].(map[string]any)
	if !ok {
		return content
	}
	if _, has := servers["moai"]; !has {
		return content // already absent — leave the bytes untouched
	}
	delete(servers, "moai")
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return content
	}
	return append(bytes.TrimRight(out, "\n"), '\n')
}

// pluginRehomedMirror materializes the plugin-mode mirror under the
// no-verification fallback: one REAL DIRECTORY COPY per skill the embedded
// tree carries under .claude/skills, written from the tree itself (rendered
// when a renderer + context are wired) — never a symlink into the
// undeployed canonical directory. Entries already on the mirror path are
// never clobbered: a non-link entry there is the user's (the same
// skip-and-report rule mirrorOneSkill applies).
func (d *deployer) pluginRehomedMirror(projectRoot string, tmplCtx *TemplateContext) []SkillMirrorEntry {
	const skillsRoot = CanonicalSkillsRelDir
	entries, err := fs.ReadDir(d.fsys, skillsRoot)
	if err != nil {
		// No skill catalog in this tree (never happens for the embedded
		// tree; a custom FS may lack it): nothing to rehome.
		return nil
	}
	var result []SkillMirrorEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		result = append(result, d.rehomeOneSkill(projectRoot, tmplCtx, e.Name()))
	}
	return result
}

// rehomeOneSkill writes one skill's real-directory copy under
// .agents/skills, rendering .tmpl sources when the renderer is wired.
func (d *deployer) rehomeOneSkill(projectRoot string, tmplCtx *TemplateContext, skill string) SkillMirrorEntry {
	mirrorPath := filepath.Join(projectRoot, MirrorSkillsRelDir, skill)

	// A non-link entry already occupies the mirror path: it may be the
	// user's — skip and report, never overwrite.
	if info, err := os.Lstat(mirrorPath); err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			return SkillMirrorEntry{
				Skill: skill,
				Mode:  MirrorModeSkipped,
				Warning: "plugin mirror " + skill + ": a non-symlink entry already exists at " +
					MirrorSkillsRelDir + "/" + skill + " — left untouched",
			}
		}
		// A stale link from an earlier local deploy: replace it with the
		// re-homed copy (a link into the undeployed canonical dir would
		// dangle).
		if rmErr := os.Remove(mirrorPath); rmErr != nil {
			return SkillMirrorEntry{Skill: skill, Mode: MirrorModeFailed,
				Warning: "plugin mirror " + skill + ": cannot replace stale link: " + rmErr.Error()}
		}
	}

	srcRoot := CanonicalSkillsRelDir + "/" + skill
	if err := os.MkdirAll(mirrorPath, 0o755); err != nil {
		return SkillMirrorEntry{Skill: skill, Mode: MirrorModeFailed,
			Warning: "plugin mirror " + skill + ": cannot create " + MirrorSkillsRelDir + "/" + skill + ": " + err.Error()}
	}
	copyErr := fs.WalkDir(d.fsys, srcRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := strings.CutPrefix(path, srcRoot+"/")
		if !relErr {
			return nil // the root itself
		}
		if entry.IsDir() {
			return os.MkdirAll(filepath.Join(mirrorPath, rel), 0o755)
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		data, readErr := fs.ReadFile(d.fsys, path)
		if readErr != nil {
			return readErr
		}
		if strings.HasSuffix(path, ".tmpl") && d.renderer != nil && tmplCtx != nil {
			rendered, renderErr := d.renderer.Render(path, tmplCtx)
			if renderErr == nil {
				data = rendered
			}
			// A render failure falls back to raw bytes: the mirror copy is
			// a preservation convenience, and the deploy walk's own render
			// error surface is where a template defect fails loudly.
		}
		return os.WriteFile(filepath.Join(mirrorPath, rel), data, 0o644)
	})
	if copyErr != nil {
		return SkillMirrorEntry{Skill: skill, Mode: MirrorModeFailed,
			Warning: "plugin mirror " + skill + ": copy failed: " + copyErr.Error()}
	}
	return SkillMirrorEntry{
		Skill: skill,
		Mode:  MirrorModeCopy,
		Warning: "plugin mirror " + skill + ": written as a real directory copy (Codex execution not yet verified) — " +
			"the copy does not follow later updates to .claude/skills/" + skill,
	}
}

// RehomeExistingMirrorEntries re-homes a project's EXISTING .agents/skills
// symlink entries to real directory copies rendered from the embedded tree
// (SPEC-INIT-SHRINK-001 design §3 mirror paragraph — the migration-path
// re-home, card t1438 review finding 2). Only entries that currently exist
// as symlinks are converted: a non-link entry is the user's and stays
// untouched (rehomeOneSkill's skip rule), and a catalog skill with no entry
// gains none — the re-home converts, it never provisions (REQ-019 holds
// mirror addition out of update runs). Failures are per-entry
// (MirrorModeFailed + Warning); the caller decides what the user sees.
func RehomeExistingMirrorEntries(projectRoot string, tmplCtx *TemplateContext) []SkillMirrorEntry {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		return []SkillMirrorEntry{{Mode: MirrorModeFailed,
			Warning: "cannot load embedded templates: " + err.Error()}}
	}
	entries, err := fs.ReadDir(fsys, CanonicalSkillsRelDir)
	if err != nil {
		// No skill catalog in this tree (never happens for the embedded
		// tree; a custom FS may lack it): nothing to rehome.
		return nil
	}
	d := &deployer{fsys: fsys, renderer: NewRenderer(fsys)}
	var result []SkillMirrorEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mirrorPath := filepath.Join(projectRoot, MirrorSkillsRelDir, e.Name())
		// Convert only a KEPT LINK: an absent entry is not provisioned, a
		// non-link entry is the user's.
		if info, statErr := os.Lstat(mirrorPath); statErr != nil || info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		result = append(result, d.rehomeOneSkill(projectRoot, tmplCtx, e.Name()))
	}
	return result
}
