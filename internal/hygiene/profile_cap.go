package hygiene

// SPEC-CAPACITY-RECLAIM-001 REQ-CAP-009: the profile-cap advisory check.
// The cap is warning-only by operator verdict (A11): when a profile under
// the home's claude-profiles grows above state.profile_max_bytes, the
// SessionStart deferred pass logs an advisory naming the profile and the
// key — nothing is ever deleted by this path. The loader has NO compiled
// fallback: an absent key yields no warning at all. The check is
// time-boxed per the Advisory-Check Discipline (coding-standards.md): it
// walks profile trees to size them, so a deadline bounds it and a timeout
// degrades to a skip notice instead of blocking the launch path.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// profileCapYAML reads state.profile_max_bytes from a tier's state.yaml.
// Every field is a pointer: nil means the key is absent.
type profileCapYAML struct {
	State struct {
		ProfileMaxBytes *int64 `yaml:"profile_max_bytes"`
	} `yaml:"state"`
}

// LoadProfileCapMaxBytes reads state.profile_max_bytes for the advisory
// check. The HOME tier — ~/.moai/config/sections/state.yaml resolved via
// paths.UserConfigSectionsDir() — is read first; when the file or the key is
// absent there and projectDir is non-empty, the PROJECT tier
// (<projectDir>/.moai/config/sections/state.yaml) is read, because the
// shipped default (profile_max_bytes in the template state.yaml) lands in
// the project tier and a project that never configured the home tier still
// carries a cap. Both tiers absent, or the key absent in both: ok is false —
// there is no compiled fallback, so an absent key yields no warning
// (REQ-CAP-009 / D11). A non-positive explicit value is config-invalid and
// is refused, mirroring the D30 fail-closed floors. Read and parse failures
// are returned as errors so callers choose their degradation.
//
// @MX:NOTE: [AUTO] No-compiled-fallback contract: an absent key returns
// ok=false and callers must NOT substitute a compiled cap (the former
// config.DefaultProfileMaxBytes was removed, card t1665) — a compiled
// fallback would warn on projects that never configured a cap.
// @MX:SPEC: SPEC-CAPACITY-RECLAIM-001
func LoadProfileCapMaxBytes(projectDir string) (capBytes int64, ok bool, err error) {
	// HOME tier first — semantics unchanged from the home-only loader.
	capBytes, ok, err = readProfileCapTier(func() (string, error) { return paths.UserConfigSectionsDir() })
	if err != nil || ok {
		return capBytes, ok, err
	}
	if projectDir == "" {
		return 0, false, nil
	}
	// PROJECT tier fallback: the canonical project sections directory
	// <projectDir>/.moai/config/sections, joined from the defs constants.
	projectSections := filepath.Join(projectDir, defs.MoAIDir, defs.SectionsSubdir)
	return readProfileCapTier(func() (string, error) { return projectSections, nil })
}

// readProfileCapTier reads state.profile_max_bytes from one tier's sections
// directory. File-absent or key-absent yields ok=false; a non-positive
// explicit value is config-invalid; other read/parse failures are errors.
func readProfileCapTier(sectionsDir func() (string, error)) (int64, bool, error) {
	dir, err := sectionsDir()
	if err != nil {
		return 0, false, fmt.Errorf("resolve config sections dir: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("read state.yaml: %w", err)
	}
	var wrapper profileCapYAML
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return 0, false, fmt.Errorf("parse state.yaml: %w", err)
	}
	if wrapper.State.ProfileMaxBytes == nil {
		return 0, false, nil
	}
	if *wrapper.State.ProfileMaxBytes <= 0 {
		return 0, false, fmt.Errorf("state.yaml: state.profile_max_bytes must be strictly positive, got %d",
			*wrapper.State.ProfileMaxBytes)
	}
	return *wrapper.State.ProfileMaxBytes, true, nil
}

// ProfileCapAdvisories sizes every profile directory under profilesDir and
// returns one advisory message per profile whose total size exceeds
// capBytes. Each message names the profile and the state.profile_max_bytes
// key. Nothing is deleted. Symlinks are not followed (a conservative
// undercount for a warning). Unreadable entries are skipped and counted,
// never fatal: a profile with unreadable sub-paths still warns on its
// partial sum, and the advisory names the skipped count. The walk is
// bounded by timeout (Advisory-Check Discipline): on deadline the
// advisories measured so far are returned together with
// context.DeadlineExceeded, which the caller degrades on without blocking.
//
// @MX:NOTE: [AUTO] Warning-only invariant: this function deletes nothing —
// space is reclaimed with `moai clean --home --force`. A repair that turns
// an advisory into a deletion breaks REQ-CAP-009.
// @MX:SPEC: SPEC-CAPACITY-RECLAIM-001
func ProfileCapAdvisories(profilesDir string, capBytes int64, timeout time.Duration) ([]string, error) {
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read profiles dir: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var advisories []string
	timedOut := false
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		size, skipped, walkErr := profileSizeWithDeadline(ctx, filepath.Join(profilesDir, e.Name()))
		if walkErr != nil {
			if errors.Is(walkErr, context.DeadlineExceeded) {
				timedOut = true
				break
			}
			continue // unreadable profile: skipped, advisory never blocks
		}
		if size > capBytes {
			msg := fmt.Sprintf(
				"profile %s is %d bytes, above state.profile_max_bytes (%d bytes); advisory only — nothing was deleted",
				e.Name(), size, capBytes)
			if skipped > 0 {
				// Partial-size semantics: the unreadable entries were
				// skipped, so the measured sum is a lower bound — the
				// advisory says so instead of staying silent.
				msg += fmt.Sprintf("; measurement partial: %d entries unreadable (skipped)", skipped)
			}
			advisories = append(advisories, msg)
		}
	}
	if timedOut {
		return advisories, context.DeadlineExceeded
	}
	return advisories, nil
}

// profileSizeWithDeadline walks root summing file sizes, aborting with
// ctx.Err() once the deadline passes. Unreadable sub-paths (a chmod-000
// directory, a vanished entry) are SKIPPED and counted — they never abort
// the walk — so an over-cap profile still warns on the partial sum; only a
// deadline/context error aborts.
func profileSizeWithDeadline(ctx context.Context, root string) (total int64, skipped int, err error) {
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			skipped++ // unreadable sub-path: skipped, counted — never fatal
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			skipped++ // unreadable entry: skipped, counted
			return nil
		}
		total += info.Size()
		return nil
	})
	return total, skipped, err
}
