// skill_mirror.go — dual-harness skill exposure.
//
// The canonical skill catalog lives at .claude/skills/. Codex CLI does not scan
// that path; it scans <repo>/.agents/skills/. This file makes every skill this
// run deployed reachable from BOTH paths by creating .agents/skills/<name> as a
// relative symlink to ../../.claude/skills/<name>, falling back to a real
// directory copy where symlink creation is unavailable.
//
// Two constraints shape the design and MUST NOT be "simplified" away:
//
//  1. //go:embed silently drops symlinks from a directory-pattern embed — no
//     build error, no warning. So the template source tree carries no symlink
//     and the link is created here, at deploy time.
//  2. manifest.Track hashes through the link (os.Open + io.Copy), which fails
//     EISDIR on a directory symlink. Mirror entries are therefore never tracked
//     in the manifest; the canonical files already are, so no information is
//     lost.
//
// Lifecycle management of the mirror (removing a mirror whose skill was renamed
// or retired) is deliberately NOT handled here — it belongs to the clean path.
package template

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// MirrorMode records how one skill's mirror entry was materialized.
type MirrorMode string

const (
	// MirrorModeSymlink — a relative symlink to the canonical directory.
	MirrorModeSymlink MirrorMode = "symlink"
	// MirrorModeCopy — a real directory copy (symlink creation unavailable).
	MirrorModeCopy MirrorMode = "copy"
	// MirrorModeSkipped — a non-symlink entry already occupied the path and
	// was left untouched.
	MirrorModeSkipped MirrorMode = "skipped"
	// MirrorModeFailed — both symlink and copy failed; deployment continued.
	MirrorModeFailed MirrorMode = "failed"
)

// canonicalSkillsPrefix is the deploy-relative directory holding the canonical
// skill catalog. Paths from fs.WalkDir are always slash-separated.
const canonicalSkillsPrefix = ".claude/skills/"

// mirrorSkillsRelDir is the deploy-relative directory holding the mirror.
var mirrorSkillsRelDir = filepath.Join(".agents", "skills")

// SkillMirrorEntry is the per-skill outcome of mirror creation.
type SkillMirrorEntry struct {
	// Skill is the skill directory name (e.g. "moai-workflow-tdd").
	Skill string
	// Mode is how the mirror entry was materialized.
	Mode MirrorMode
	// Warning is a human-readable note, empty when the entry needs none.
	Warning string
}

// DeployResult carries the observable outcome of a deployment back to the
// caller. The deployer never prints: displaying this is the caller's job.
type DeployResult struct {
	// SkillMirrors holds one entry per skill the run attempted to mirror.
	SkillMirrors []SkillMirrorEntry
}

// CopyFallbackUsed reports whether any skill fell back to a directory copy.
func (r *DeployResult) CopyFallbackUsed() bool {
	if r == nil {
		return false
	}
	for _, e := range r.SkillMirrors {
		if e.Mode == MirrorModeCopy {
			return true
		}
	}
	return false
}

// Warnings returns every non-empty warning collected during the run.
func (r *DeployResult) Warnings() []string {
	if r == nil {
		return nil
	}
	var out []string
	for _, e := range r.SkillMirrors {
		if e.Warning != "" {
			out = append(out, e.Warning)
		}
	}
	return out
}

// MirrorMode returns the recorded mode for a skill, and whether it was found.
func (r *DeployResult) MirrorMode(skill string) (MirrorMode, bool) {
	if r == nil {
		return "", false
	}
	for _, e := range r.SkillMirrors {
		if e.Skill == skill {
			return e.Mode, true
		}
	}
	return "", false
}

// DeployerOption customizes a Deployer at construction time.
type DeployerOption func(*deployer)

// WithSkillMirror enables or disables .agents/skills mirror creation.
// Mirroring is enabled by default; disabling it yields a deployment whose
// .claude/ output is byte-for-byte what it would be without this feature.
func WithSkillMirror(enabled bool) DeployerOption {
	return func(d *deployer) { d.skillMirrorDisabled = !enabled }
}

// withSymlinkFunc replaces the symlink syscall (test seam).
func withSymlinkFunc(fn func(oldname, newname string) error) DeployerOption {
	return func(d *deployer) { d.symlinkFn = fn }
}

// withMirrorCopyFunc replaces the copy fallback (test seam).
func withMirrorCopyFunc(fn func(srcDir, dstDir string) error) DeployerOption {
	return func(d *deployer) { d.mirrorCopyFn = fn }
}

// skillNameFromDeployPath extracts the skill directory name from a
// deploy-relative path under the canonical skills directory. It returns false
// for any path outside that directory, and for the directory itself.
func skillNameFromDeployPath(deployRelPath string) (string, bool) {
	rest, ok := strings.CutPrefix(deployRelPath, canonicalSkillsPrefix)
	if !ok {
		return "", false
	}
	name, remainder, found := strings.Cut(rest, "/")
	// A bare ".claude/skills/<name>" with no remainder is the skill directory
	// entry itself; the walk only reports files, so require a remainder.
	if !found || name == "" || remainder == "" {
		return "", false
	}
	return name, true
}

// mirrorLinkTarget is the relative symlink body for a skill mirror. Relative
// (never absolute) so the project directory can be moved or copied wholesale
// without breaking the link.
func mirrorLinkTarget(skill string) string {
	return path.Join("..", "..", ".claude", "skills", skill)
}

// symlink invokes the configured symlink function (os.Symlink by default).
func (d *deployer) symlink(oldname, newname string) error {
	if d.symlinkFn != nil {
		return d.symlinkFn(oldname, newname)
	}
	return os.Symlink(oldname, newname)
}

// mirrorCopy invokes the configured copy function (copyTree by default).
func (d *deployer) mirrorCopy(srcDir, dstDir string) error {
	if d.mirrorCopyFn != nil {
		return d.mirrorCopyFn(srcDir, dstDir)
	}
	return copyTree(srcDir, dstDir)
}

// @MX:ANCHOR: [AUTO] Sole entry point for .agents/skills mirror creation; called once per Deploy.
// @MX:REASON: [AUTO] Governs a destructive-adjacent path (pre-occupied mirror targets); every branch here is bound by a MUST acceptance criterion.
//
// mirrorSkills materializes one mirror entry per deployed skill. It is
// fail-open by contract: no failure here is propagated as a deployment error,
// because losing the Claude Code path over a Codex-visibility convenience would
// be a wildly disproportionate reaction.
func (d *deployer) mirrorSkills(projectRoot string, skills []string) []SkillMirrorEntry {
	if len(skills) == 0 {
		return nil
	}
	mirrorDir := filepath.Join(projectRoot, mirrorSkillsRelDir)
	entries := make([]SkillMirrorEntry, 0, len(skills))
	for _, skill := range skills {
		entries = append(entries, d.mirrorOneSkill(projectRoot, mirrorDir, skill))
	}
	return entries
}

func (d *deployer) mirrorOneSkill(projectRoot, mirrorDir, skill string) SkillMirrorEntry {
	mirrorPath := filepath.Join(mirrorDir, skill)
	srcDir := filepath.Join(projectRoot, ".claude", "skills", skill)
	want := mirrorLinkTarget(skill)

	// Lstat, not Stat: Stat follows the link and cannot tell a link from a real
	// directory, which is exactly the distinction that decides whether the path
	// may be replaced.
	if info, err := os.Lstat(mirrorPath); err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			// A real file or directory occupies the path. It may be the user's.
			// Never remove or overwrite it — skip and report.
			return SkillMirrorEntry{
				Skill:   skill,
				Mode:    MirrorModeSkipped,
				Warning: fmt.Sprintf("skill mirror %s: a non-symlink entry already exists at %s — left untouched", skill, filepath.Join(mirrorSkillsRelDir, skill)),
			}
		}
		if current, readErr := os.Readlink(mirrorPath); readErr == nil && current == want {
			// Already correct — leave it exactly as it is (idempotent).
			return SkillMirrorEntry{Skill: skill, Mode: MirrorModeSymlink}
		}
		// A symlink pointing somewhere else: replace it with the correct one.
		if rmErr := os.Remove(mirrorPath); rmErr != nil {
			return SkillMirrorEntry{
				Skill:   skill,
				Mode:    MirrorModeFailed,
				Warning: fmt.Sprintf("skill mirror %s: cannot replace stale link: %v", skill, rmErr),
			}
		}
	}

	if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
		return SkillMirrorEntry{
			Skill:   skill,
			Mode:    MirrorModeFailed,
			Warning: fmt.Sprintf("skill mirror %s: cannot create %s: %v", skill, mirrorSkillsRelDir, err),
		}
	}

	if err := d.symlink(want, mirrorPath); err == nil {
		return SkillMirrorEntry{Skill: skill, Mode: MirrorModeSymlink}
	}

	// Symlink creation is unavailable (Windows without the privilege, an
	// exotic filesystem, a sandbox). Fall back to a real copy — and say so,
	// because a silent copy looks like a link that stopped tracking the
	// canonical directory.
	_ = os.RemoveAll(mirrorPath)
	if err := d.mirrorCopy(srcDir, mirrorPath); err != nil {
		return SkillMirrorEntry{
			Skill:   skill,
			Mode:    MirrorModeFailed,
			Warning: fmt.Sprintf("skill mirror %s: symlink and copy both failed: %v — skill is reachable via .claude/skills only", skill, err),
		}
	}
	return SkillMirrorEntry{
		Skill:   skill,
		Mode:    MirrorModeCopy,
		Warning: fmt.Sprintf("skill mirror %s: copied instead of linked (symlink unavailable) — the copy does not follow later updates to .claude/skills/%s", skill, skill),
	}
}

// copyTree copies srcDir into dstDir recursively, preserving the executable
// bit. It is the symlink fallback, so it never follows symlinks it finds in
// the source: the canonical skill tree contains none.
func copyTree(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(srcPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(srcDir, srcPath)
		if relErr != nil {
			return relErr
		}
		dstPath := filepath.Join(dstDir, rel)
		if entry.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}
		if !entry.Type().IsRegular() {
			// Skip anything that is not a regular file (symlink, socket, …).
			return nil
		}
		content, readErr := os.ReadFile(srcPath)
		if readErr != nil {
			return readErr
		}
		perm := fs.FileMode(0o644)
		if info, statErr := entry.Info(); statErr == nil {
			perm = info.Mode().Perm()
		}
		if mkErr := os.MkdirAll(filepath.Dir(dstPath), 0o755); mkErr != nil {
			return mkErr
		}
		return os.WriteFile(dstPath, content, perm)
	})
}
