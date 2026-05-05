// SPEC-V3R3-HARNESS-001 / T-M4-01
// update_archive.go — recursively copies from .claude/skills/<id>/ to
// provides archiveSkill function that recursively copies.
//
// Idempotency guarantee:
// - source absent → return nil immediately
// - archive does not exist + content same → return nil
// - archive does not exist + content differs (drift) → ARCHIVE_DRIFT error
//
// path traversal protection:
// - return error if skillID contains ".." or "/"

package cli

import (
"fmt"
"io"
"os"
"path/filepath"
"strings"
)

// archiveVersion version tag used for archive directory.
const archiveVersion = "v2.16"

// legacySkillIDs is a list of 16 skill IDs removed in BC-V3R3-007.
// skills are moved to .moai/archive/skills/v2.16/ when moai update is run.
var legacySkillIDs = []string{
"moai-domain-backend",
"moai-domain-frontend",
"moai-domain-database",
"moai-domain-db-docs",
"moai-domain-mobile",
"moai-framework-electron",
"moai-library-shadcn",
"moai-library-mermaid",
"moai-library-nextra",
"moai-tool-ast-grep",
"moai-platform-auth",
"moai-platform-deployment",
"moai-platform-chrome-extension",
"moai-workflow-research",
"moai-workflow-pencil-integration",
"moai-formats-data",
}

// archiveSkill projectRoot/.claude/skills/<skillID>/
// Copy projectRoot/.claude/skills/<skillID>/ to projectRoot/.moai/archive/skills/v2.16/<skillID>/.
//
// @MX:ANCHOR: [AUTO] entry point for legacy skill archival contract
// @MX:REASON: [AUTO] called from runUpdate flow, restore-skill, idempotency tests, etc.; fan_in >= 3
func archiveSkill(projectRoot, skillID string) error {
// prevent path traversal attacks: disallow ".." or "/" (separator) in skillID
if err := validateSkillID(skillID); err != nil {
return err
}

srcDir := filepath.Join(projectRoot, ".claude", "skills", skillID)

// return nil idempotently if source does not exist (handle not-yet-removed case)
if _, err := os.Stat(srcDir); err != nil {
if os.IsNotExist(err) {
return nil
}
return fmt.Errorf("stat source skill %s: %w", skillID, err)
}

dstDir := filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion, skillID)

// if archive does not exist, check for drift
if _, err := os.Stat(dstDir); err == nil {
if err := checkArchiveDrift(srcDir, dstDir); err != nil {
return err
}
// content same → idempotent success
return nil
}

// create archive target directory
if err := os.MkdirAll(dstDir, 0o755); err != nil {
return fmt.Errorf("create archive directory for %s: %w", skillID, err)
}

// recursive copy
if err := copyDirAll(srcDir, dstDir); err != nil {
// clean up incomplete archive on failure
_ = os.RemoveAll(dstDir)
return fmt.Errorf("copy skill %s to archive: %w", skillID, err)
}

return nil
}

// validateSkillID checks if skillID contains path traversal characters ("..", "/", absolute path)
// check if it includes.
func validateSkillID(skillID string) error {
// reject absolute path
if filepath.IsAbs(skillID) {
return &MigrateError{
Code: "ARCHIVE_INVALID_ID",
Message: fmt.Sprintf("skillID must be a simple name, not an absolute path: %q", skillID),
}
}
// reject ".." inclusion
if strings.Contains(skillID, "..") {
return &MigrateError{
Code: "ARCHIVE_INVALID_ID",
Message: fmt.Sprintf("skillID must not contain '..': %q", skillID),
}
}
// reject "/" inclusion (platform separators)
if strings.ContainsAny(skillID, "/\\") {
return &MigrateError{
Code: "ARCHIVE_INVALID_ID",
Message: fmt.Sprintf("skillID must not contain path separators: %q", skillID),
}
}
return nil
}

// checkArchiveDrift content of source directory and existing archive directory
// compare for equality using SHA-256.
// return nil if same, ARCHIVE_DRIFT error if different.
func checkArchiveDrift(srcDir, dstDir string) error {
srcHashes, err := computeDirHashes(srcDir)
if err != nil {
return fmt.Errorf("compute source hashes: %w", err)
}
dstHashes, err := computeDirHashes(dstDir)
if err != nil {
return fmt.Errorf("compute archive hashes: %w", err)
}

if len(srcHashes) != len(dstHashes) {
return &MigrateError{
Code: "ARCHIVE_DRIFT",
Message: fmt.Sprintf(
"archive already exists but file count differs (src=%d, dst=%d). "+
"Use --force to overwrite.",
len(srcHashes), len(dstHashes),
),
}
}

for rel, srcHash := range srcHashes {
dstHash, ok := dstHashes[rel]
if !ok || srcHash != dstHash {
return &MigrateError{
Code: "ARCHIVE_DRIFT",
Message: fmt.Sprintf(
"archive already exists but content differs for %s. "+
"Use --force to overwrite.",
rel,
),
}
}
}

return nil
}

// computeDirHashes SHA-256 hash of all files in directory
// return map of path→hash.
// reuse existing hashFile function (from design_folder.go).
func computeDirHashes(dir string) (map[string]string, error) {
hashes := make(map[string]string)
err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
if err != nil {
return err
}
if d.IsDir() {
return nil
}
rel, err := filepath.Rel(dir, path)
if err != nil {
return err
}
// hashFile declaration from design_folder.go, returns []byte
rawHash, err := hashFile(path)
if err != nil {
return err
}
hashes[rel] = fmt.Sprintf("%x", rawHash)
return nil
})
return hashes, err
}

// copyDirAll recursively copy all contents of srcDir to dstDir.
// preserve Unix permissions for each file.
func copyDirAll(srcDir, dstDir string) error {
return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
if err != nil {
return err
}

rel, err := filepath.Rel(srcDir, path)
if err != nil {
return err
}

dstPath := filepath.Join(dstDir, rel)

if d.IsDir() {
info, err := d.Info()
if err != nil {
return err
}
return os.MkdirAll(dstPath, info.Mode().Perm())
}

return copyFile(path, dstPath)
})
}

// archiveLegacySkills skills from legacySkillIDs list in project root
// iterate and call archiveSkill. Return count of newly archived skills.
//
// Idempotent: do not count if archive already exists with same content.
//
// Output format:
//
// archive: <id> → .moai/archive/skills/v2.16/<id>
// total: N skills archived, 0 user customizations modified
//
// @MX:ANCHOR: [AUTO] entry point for legacy skill archival in update flow
// @MX:REASON: [AUTO] called from runUpdate, dry-run, idempotency tests; fan_in >= 3
func archiveLegacySkills(projectRoot string, out io.Writer) (int, error) {
archived := 0
for _, id := range legacySkillIDs {
srcDir := filepath.Join(projectRoot, ".claude", "skills", id)
if _, err := os.Stat(srcDir); err != nil {
// source missing → skip
continue
}

// check if archive does not exist first (idempotency check))
dstDir := filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion, id)
alreadyArchived := false
if _, err := os.Stat(dstDir); err == nil {
alreadyArchived = true
}

if err := archiveSkill(projectRoot, id); err != nil {
return archived, fmt.Errorf("archive %s: %w", id, err)
}

// do not count if already existed (idempotent run)
if alreadyArchived {
continue
}

archiveDst := filepath.Join(".moai", "archive", "skills", archiveVersion, id)
_, _ = fmt.Fprintf(out, "archive: %s → %s\n", id, archiveDst)
archived++
}

_, _ = fmt.Fprintf(out, "total: %d skills archived, 0 user customizations modified\n", archived)
return archived, nil
}

// dryRunArchiveLegacySkills runs in --dry-run mode
// output planned work without actual filesystem changes.
func dryRunArchiveLegacySkills(projectRoot string, out io.Writer) error {
planned := 0
for _, id := range legacySkillIDs {
srcDir := filepath.Join(projectRoot, ".claude", "skills", id)
if _, err := os.Stat(srcDir); err != nil {
continue
}
archiveDst := filepath.Join(".moai", "archive", "skills", archiveVersion, id)
_, _ = fmt.Fprintf(out, "[dry-run] archive: %s → %s\n", id, archiveDst)
planned++
}
_, _ = fmt.Fprintf(out, "[dry-run] total: %d skills archived, 0 user customizations modified\n", planned)
return nil
}

// copyFile copies a single file from source to target.
// preserve source permission bits.
func copyFile(src, dst string) error {
srcInfo, err := os.Stat(src)
if err != nil {
return fmt.Errorf("stat %s: %w", src, err)
}

in, err := os.Open(src)
if err != nil {
return fmt.Errorf("open %s: %w", src, err)
}
defer func() { _ = in.Close() }()

out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode().Perm())
if err != nil {
return fmt.Errorf("create %s: %w", dst, err)
}
defer func() { _ = out.Close() }()

if _, err := io.Copy(out, in); err != nil {
return fmt.Errorf("copy %s → %s: %w", src, dst, err)
}

return nil
}
