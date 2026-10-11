package settings

// participation.go — the ONE writer for the user-scoped consent file
// (SPEC-FEEDBACK-PARTICIPATION-001 design.md section 9).
//
// Three entrances drive it: the `moai init` wizard (applyParticipationFromWizard),
// the `moai update` ask (runParticipationStep), and the `moai web` console
// (ApplySchemaEdits' user-scoped branch). All three write the same shape
// through this one function — a yamlpatch.PatchFile over
// <moai home>/config/participation.yaml — so the file's schema, its unknown-key
// survival (a user-set participation.repository rides through untouched), and
// its atomicity have a single implementation.
//
// The file lives beside, not inside, config/sections/ precisely so the section
// resolver's tier merge can never decide a consent value (the reader's doc
// comment in internal/config carries the reasoning and the fail-closed rule).

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// ParticipationField is the schema field name of the participation toggle.
// Exported so the wizard, the update step, the console schema, and the tests
// name the same target.
const ParticipationField = "feedback.participation"

// WriteUserParticipation writes the consent state to the user-scoped file,
// creating the config directory 0700 when it does not exist and preserving
// every key it does not write (PatchFile's contract).
func WriteUserParticipation(p config.UserParticipation) error {
	path, err := config.UserParticipationFilePath()
	if err != nil {
		return fmt.Errorf("settings: resolve participation path: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("settings: create participation directory: %w", err)
	}
	edits := []yamlpatch.KeyEdit{
		{Path: []string{"participation", "enabled"}, Value: strconv.FormatBool(p.Enabled)},
		{Path: []string{"participation", "asked"}, Value: strconv.FormatBool(p.Asked)},
	}
	// The repository override is optional: an empty value writes nothing, so
	// the writer never plants an empty key, and PatchFile's no-delete contract
	// keeps any existing value alive when the callers pass it through.
	if p.Repository != "" {
		edits = append(edits, yamlpatch.KeyEdit{
			Path: []string{"participation", "repository"}, Value: p.Repository,
		})
	}
	if err := yamlpatch.PatchFile(path, edits); err != nil {
		return fmt.Errorf("settings: write participation file: %w", err)
	}
	// PatchFile's atomic write honors the file's existing mode; a file this
	// writer created must be user-only, because consent state names a person's
	// account posture. Best-effort chmod: an immutable filesystem surfaces the
	// next time the reader runs, not as a failed save.
	if info, statErr := os.Stat(path); statErr == nil && info.Mode().Perm() != 0o600 {
		_ = os.Chmod(path, 0o600)
	}
	return nil
}
