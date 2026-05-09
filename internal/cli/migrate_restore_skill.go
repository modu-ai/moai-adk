// SPEC-V3R3-HARNESS-001 / T-M4-03
// migrate_restore_skill.go — moai migrate restore-skill <skill-id> subcommand.
//
// Restore 16 legacy skills archived in BC-V3R3-007.
// Restore path: .moai/archive/skills/v2.16/<id>/ → .claude/skills/<id>/
//
// Error codes (MigrateError pattern):
// - RESTORE_ARCHIVE_MISSING : archive entry not found
// - RESTORE_TARGET_EXISTS : target already exists (--force required)
// - RESTORE_INVALID_SKILL_ID : path traversal or empty skillID

package cli

import (
"fmt"
"os"
"path/filepath"

"github.com/spf13/cobra"
)

// restoreSkill .moai/archive/skills/v2.16/<skillID>/
// recursively copy to .claude/skills/<skillID>/.
//
// if force=true, overwrite target even if it exists.
//
// @MX:ANCHOR: [AUTO] entry point for archive → skills restore contract
// @MX:REASON: [AUTO] called from runRestoreSkill, TestRestoreSkill_*, round-trip tests; fan_in >= 3
func restoreSkill(projectRoot, skillID string, force bool) error {
// prevent path traversal
if err := validateSkillID(skillID); err != nil {
return &MigrateError{
Code: "RESTORE_INVALID_SKILL_ID",
Message: err.Error(),
}
}
if skillID == "" {
return &MigrateError{
Code: "RESTORE_INVALID_SKILL_ID",
Message: "skillID must not be empty",
}
}

archiveDir := filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion, skillID)
targetDir := filepath.Join(projectRoot, ".claude", "skills", skillID)

// verify archive exists
if _, err := os.Stat(archiveDir); err != nil {
if os.IsNotExist(err) {
return &MigrateError{
Code: "RESTORE_ARCHIVE_MISSING",
Message: fmt.Sprintf("no archive entry for skill %q at %s", skillID, archiveDir),
}
}
return fmt.Errorf("stat archive %s: %w", skillID, err)
}

// verify target exists
if _, err := os.Stat(targetDir); err == nil {
if !force {
return &MigrateError{
Code: "RESTORE_TARGET_EXISTS",
Message: fmt.Sprintf(
".claude/skills/%s already exists; use --force to overwrite", skillID,
),
}
}
// --force: remove existing target before restore
if err := os.RemoveAll(targetDir); err != nil {
return fmt.Errorf("remove existing target %s: %w", skillID, err)
}
}

// create target directory
if err := os.MkdirAll(targetDir, 0o755); err != nil {
return fmt.Errorf("create target directory for %s: %w", skillID, err)
}

// recursive copy (archive → skill directory))
if err := copyDirAll(archiveDir, targetDir); err != nil {
_ = os.RemoveAll(targetDir)
return fmt.Errorf("restore skill %s: %w", skillID, err)
}

return nil
}

// --- Cobra command wiring ---

var restoreSkillCmd = &cobra.Command{
Use: "restore-skill <skill-id>",
Short: "Restore a legacy skill from the v2.16 archive",
Long: `Restore a skill that was archived during moai update (BC-V3R3-007).

Copies .moai/archive/skills/v2.16/<skill-id>/ back to .claude/skills/<skill-id>/.

Error codes:
RESTORE_ARCHIVE_MISSING – no archive entry found for the given skill-id
RESTORE_TARGET_EXISTS – target directory already exists (use --force)
RESTORE_INVALID_SKILL_ID – skill-id contains path traversal characters

Example:
moai migrate restore-skill moai-domain-backend
moai migrate restore-skill moai-domain-backend --force`,
Args: cobra.ExactArgs(1),
RunE: runRestoreSkill,
}

func init() {
migrateCmd.AddCommand(restoreSkillCmd)
restoreSkillCmd.Flags().Bool("force", false, "Overwrite existing target directory")
}

func runRestoreSkill(cmd *cobra.Command, args []string) error {
skillID := args[0]
force, _ := cmd.Flags().GetBool("force")
out := cmd.OutOrStdout()

cwd, err := os.Getwd()
if err != nil {
return fmt.Errorf("get working directory: %w", err)
}

if err := restoreSkill(cwd, skillID, force); err != nil {
if me, ok := err.(*MigrateError); ok {
_, _ = fmt.Fprintf(out, "Error [%s]: %s\n", me.Code, me.Message)
return me
}
return err
}

dstRel := filepath.Join(".claude", "skills", skillID)
_, _ = fmt.Fprintf(out, "restore: %s → %s\n", skillID, dstRel)
return nil
}
