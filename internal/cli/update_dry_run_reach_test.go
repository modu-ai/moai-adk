package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/pkg/version"
	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// SPEC-UPDATE-REINSTALL-LOOP-002 M4 — `--dry-run` reachability
// (REQ-RIL2-024..027 / AC-RIL2-013, AC-RIL2-014, AC-RIL2-015)
//
// Defect 4: runUpdate returned from its --dry-run branch before the v2-detection
// block, so the already-implemented non-mutating renderers inside
// runCleanReinstall (update_clean_install.go) and runV3ResidueCleanup
// (update_residue_cleanup.go) were unreachable from the CLI.
//
// The fix makes those renderers reachable. It MUST NOT relocate the --dry-run
// early return past stripRetiredV2DenyEntries (update.go), which rewrites
// .claude/settings.json — that would make a dry run write to disk.
// ---------------------------------------------------------------------------

// dryRunDeprecatedRel is a plain-file entry of defs.DeprecatedPaths used as the
// residue fixture, mirroring residueDeprecatedRel in update_residue_cleanup_test.go.
const dryRunDeprecatedRel = ".claude/agents/moai/planner.md"

// dryRunSeededDenyEntry is one of the twelve retiredV2DenyEntries literals
// (internal/cli/update_deny_migration.go). Seeding it into the fixture's
// permissions.deny array is what gives TestUpdateDryRun_ZeroMutation the power
// to fail: without it, stripRetiredV2DenyEntries short-circuits at
// `removed == 0` and returns before its os.WriteFile, so even an implementation
// that relocated the early return past the strip block would leave the tree
// byte-identical and the hash comparison would pass.
const dryRunSeededDenyEntry = "Write(~/.ssh/**)"

// newDryRunProject builds a t.TempDir() project carrying the system.yaml project
// marker at the requested moai.version, a deprecated-path residue, and a
// .claude/settings.json whose permissions.deny array seeds one retired v2 deny
// entry.
func newDryRunProject(t *testing.T, moaiVersion string) string {
	t.Helper()
	root := t.TempDir()
	residueWriteFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n  version: "+moaiVersion+"\n")
	residueWriteFile(t, root, dryRunDeprecatedRel, "legacy planner\n")
	residueWriteFile(t, root, ".claude/settings.json",
		`{"permissions":{"deny":["`+dryRunSeededDenyEntry+`","Read(./.env)"]}}`+"\n")
	return root
}

// newDryRunUpdateCmd returns a cobra command carrying every flag runUpdate
// reads, with --dry-run and --yes set.
func newDryRunUpdateCmd(out *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{Use: "update-dry-run-test"}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.Flags().Bool("check", false, "")
	cmd.Flags().Bool("shell-env", false, "")
	cmd.Flags().Bool("config", false, "")
	cmd.Flags().Bool("binary", false, "")
	cmd.Flags().Bool("templates-only", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("dry-run", true, "")
	cmd.Flags().Bool("no-hooks", false, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().String("profile", "", "")
	return cmd
}

// runDryRunUpdateInProject chdirs into root, invokes runUpdate, and returns the
// captured output. Chdir is required because runUpdate resolves the project from
// os.Getwd(); these tests are therefore NOT parallel-safe.
func runDryRunUpdateInProject(t *testing.T, root string) string {
	t.Helper()

	origDeps := deps
	t.Cleanup(func() { deps = origDeps })
	deps = &Dependencies{}

	origVersion := version.Version
	t.Cleanup(func() { version.Version = origVersion })
	version.Version = "dev" // dev build → the binary-update step is skipped

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir %s: %v", root, err)
	}

	var buf bytes.Buffer
	cmd := newDryRunUpdateCmd(&buf)
	if err := runUpdate(cmd, nil); err != nil {
		t.Fatalf("runUpdate --dry-run: %v\noutput:\n%s", err, buf.String())
	}
	return buf.String()
}

// hashDryRunTree returns a slash-relative path → sha256 map covering every
// regular file under root, plus a "<dir>/" key for every directory, so both
// content changes and directory creation (e.g. .moai/backups/) are observable.
func hashDryRunTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}
		key := filepath.ToSlash(rel)
		if d.IsDir() {
			out[key+"/"] = "<dir>"
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		sum := sha256.Sum256(data)
		out[key] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return out
}

// assertSeededDenyEntry reads back the fixture's .claude/settings.json, parses
// it, and asserts permissions.deny actually carries at least one
// retiredV2DenyEntries literal.
//
// This runtime assertion — not a static grep — is what makes the zero-mutation
// comparison non-vacuous: a grep is satisfied by a literal appearing in a
// trailing comment, which executes nothing. Parsing the written file cannot be
// satisfied by a comment in any position.
func assertSeededDenyEntry(t *testing.T, root string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("fixture seeds no readable .claude/settings.json: %v", err)
	}
	var parsed struct {
		Permissions struct {
			Deny []string `json:"deny"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("fixture .claude/settings.json is not valid JSON: %v", err)
	}
	retired := make(map[string]struct{}, len(retiredV2DenyEntries))
	for _, e := range retiredV2DenyEntries {
		retired[e] = struct{}{}
	}
	for _, entry := range parsed.Permissions.Deny {
		if _, ok := retired[entry]; ok {
			return
		}
	}
	t.Fatalf("fixture seeds no retiredV2DenyEntries literal in permissions.deny (%v); "+
		"the zero-mutation assertion would be vacuous, because stripRetiredV2DenyEntries "+
		"returns at `removed == 0` before its os.WriteFile", parsed.Permissions.Deny)
}

// AC-RIL2-013 (REQ-RIL2-024) — `moai update --dry-run` reaches the
// clean-reinstall plan on a v2 project and emits a non-zero removal count.
func TestUpdateDryRun_EmitsCleanReinstallPlan(t *testing.T) {
	root := newDryRunProject(t, "v2.16.0")

	out := runDryRunUpdateInProject(t, root)

	if !strings.Contains(out, "DRY-RUN") {
		t.Fatalf("dry-run output lacks the literal %q; the clean-reinstall plan renderer was not reached.\noutput:\n%s",
			"DRY-RUN", out)
	}
	re := regexp.MustCompile(`Would remove ([0-9]+) deprecated paths`)
	m := re.FindStringSubmatch(out)
	if m == nil {
		t.Fatalf("dry-run output lacks a %q line.\noutput:\n%s", "Would remove N deprecated paths", out)
	}
	if m[1] == "0" {
		t.Fatalf("dry-run reported a zero removal count; the fixture seeds %s.\noutput:\n%s",
			dryRunDeprecatedRel, out)
	}
}

// REQ-RIL2-025 — the v3-with-residue branch reaches the residue-cleanup plan
// renderer instead of the clean-reinstall one.
func TestUpdateDryRun_EmitsResidueCleanupPlan(t *testing.T) {
	root := newDryRunProject(t, "3.0.1")

	out := runDryRunUpdateInProject(t, root)

	re := regexp.MustCompile(`\[residue-cleanup\] Would remove ([0-9]+) deprecated paths`)
	m := re.FindStringSubmatch(out)
	if m == nil {
		t.Fatalf("dry-run output lacks the residue-cleanup plan line.\noutput:\n%s", out)
	}
	if m[1] == "0" {
		t.Fatalf("residue-cleanup plan reported a zero removal count.\noutput:\n%s", out)
	}
}

// AC-RIL2-014 (REQ-RIL2-026) — `--dry-run` mutates nothing on either branch.
func TestUpdateDryRun_ZeroMutation(t *testing.T) {
	for _, moaiVersion := range []string{"v2.16.0", "3.0.1"} {
		t.Run(moaiVersion, func(t *testing.T) {
			root := newDryRunProject(t, moaiVersion)

			// The seed must have reached the file on disk, or the comparison
			// below is vacuous — see assertSeededDenyEntry's doc comment.
			assertSeededDenyEntry(t, root)

			before := hashDryRunTree(t, root)
			_ = runDryRunUpdateInProject(t, root)
			after := hashDryRunTree(t, root)

			for key, want := range before {
				got, ok := after[key]
				if !ok {
					t.Errorf("dry-run removed %s", key)
					continue
				}
				if got != want {
					t.Errorf("dry-run modified %s", key)
				}
			}
			for key := range after {
				if _, ok := before[key]; !ok {
					t.Errorf("dry-run created %s", key)
				}
			}

			if _, err := os.Stat(filepath.Join(root, ".moai", "backups")); err == nil {
				t.Errorf("dry-run created .moai/backups/")
			}
		})
	}
}

// AC-RIL2-015 (REQ-RIL2-027) — the pre-existing dry-run emissions survive the
// M4 reordering: the legacy-skill archive summary and the worktree advisory.
func TestUpdateDryRun_PreservesExistingOutput(t *testing.T) {
	root := newDryRunProject(t, "v2.16.0")

	out := runDryRunUpdateInProject(t, root)

	for _, want := range []string{"[dry-run] total:", "use a worktree for isolation"} {
		if !strings.Contains(out, want) {
			t.Errorf("dry-run output lacks the literal %q.\noutput:\n%s", want, out)
		}
	}
}
