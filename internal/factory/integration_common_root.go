package factory

// integration_common_root.go — the one rule that names the directory whose .moai/state holds the
// release-integration window record (card t1680, sync-audit round 1 F1 and iter5).
//
// The record is read and written by three kinds of caller: the `moai integration` verbs, the MCP
// factory_complete handler, and the PreToolUse integration-lock guard. Each starts from a different
// "project directory" string, and a string naming a LINKED worktree would give each caller its own
// record. They all resolve it here, so the record has exactly one home.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// IntegrationCommonRoot returns the repository common root for dir: the parent of
// `git rev-parse --git-common-dir`, so a linked worktree, a subdirectory of one, and the primary
// checkout all answer the primary checkout.
//
// The answer is spelled by git (canonicalTreeSpelling, card t766), never as the caller wrote it.
// The question is asked with the repo-scoping environment removed (gitenv.Env, card t1478 M2),
// against dir itself, so it depends on neither this process's cwd nor its environment. Where git
// cannot answer — dir missing, not a repository, or a layout whose git directory is not named
// ".git" (bare, separate-git-dir) — dir is returned as canonicalTreeSpelling gives it, which is dir
// verbatim for a path git does not know.
//
// @MX:ANCHOR: [AUTO] the single window-root rule for the CLI verbs, the MCP handler and the PreToolUse guard
// @MX:REASON: fan_in 3 across internal/cli (integrationLockRoot, handleFactoryComplete) and internal/hook (checkIntegrationLock); a divergence here splits the lock record between its writers and its enforcer
func IntegrationCommonRoot(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-common-dir")
	cmd.Env = gitenv.Env()
	if raw, err := cmd.Output(); err == nil {
		if p := strings.TrimSpace(string(raw)); p != "" {
			if !filepath.IsAbs(p) {
				p = filepath.Join(dir, p) // git answers relative to the -C directory
			}
			if p = filepath.Clean(p); filepath.Base(p) == ".git" {
				return canonicalTreeSpelling(filepath.Dir(p))
			}
		}
	}
	return canonicalTreeSpelling(dir)
}

// canonicalTreeSpelling returns git's own spelling of dir's top level, and dir
// verbatim whenever git cannot answer for it (card t766).
//
// The spelling matters because everything downstream JOINS onto this string —
// the preserved copy's directory, and the `preserved_path` recorded beside it
// — while the sibling fields `source_path` and `worktree` come from
// `git rev-parse --show-toplevel`, which answers the on-disk spelling. One row
// could therefore name a tree two different ways, and the real ledger does:
// three rows from 2026-09-08 carry a `preserved_path` under `/Users/goos/moai/`
// against a `source_path` under `/Users/goos/MoAI/`. On a case-insensitive
// volume that is a record defect; on a case-sensitive one it is a second
// directory tree, and a preserved copy filed where nobody will look for it.
//
// Two properties bound the repair, and neither is incidental:
//
//   - git is the source, not `filepath.EvalSymlinks`. EvalSymlinks does NOT
//     correct case on macOS — it resolves links and returns whatever spelling
//     it was handed — so it cannot see this defect at all.
//
//   - The answer is accepted only when it names the SAME DIRECTORY, compared
//     by identity (os.SameFile) rather than by string. `--show-toplevel` walks
//     UP: handed a plain subdirectory of some repository, it answers that
//     repository's root, and relocating the lock root there would be a
//     different and much larger change than fixing a spelling. A path that is
//     not a repository, or is not its top level, keeps behaving exactly as it
//     does today.
func canonicalTreeSpelling(dir string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	cmd.Env = gitenv.Env() // an inherited GIT_DIR would outrank cmd.Dir
	out, err := cmd.Output()
	if err != nil {
		return dir
	}
	// Git uses forward slashes on Windows. Keep its case and identity,
	// but use the native separator used by the record root and branch lookup.
	top := filepath.FromSlash(strings.TrimSpace(string(out)))
	if top == "" || top == dir {
		return dir
	}
	given, givenErr := os.Stat(dir)
	found, foundErr := os.Stat(top)
	if givenErr != nil || foundErr != nil || !os.SameFile(given, found) {
		return dir
	}
	return top
}
