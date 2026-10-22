package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// authenticateMigrationAdmin binds a journal identity to this repository's
// real administrative directory even while Git's backlink names the old path.
// It does not create a token or change either tree.
func authenticateMigrationAdmin(root, path string, id TreeIdentity, endpoints ...string) error {
	observed, err := TreeIdent(path)
	if err != nil {
		return err
	}
	if id.Token == "" || observed.Ident() != id.Ident() {
		return fmt.Errorf("migration identity mismatch at %s", path)
	}
	admin, err := gitDirPointer(path)
	if err != nil {
		return err
	}
	raw, err := listGitOutput(root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return err
	}
	parent, err := os.Stat(filepath.Dir(admin))
	if err != nil {
		return err
	}
	expected, err := os.Stat(filepath.Join(filepath.FromSlash(strings.TrimSpace(string(raw))), "worktrees"))
	if err != nil {
		return err
	}
	if !os.SameFile(parent, expected) {
		return fmt.Errorf("migration admin directory does not belong to %s", root)
	}
	backlink, err := os.ReadFile(filepath.Join(admin, "gitdir"))
	if err != nil {
		return err
	}
	registered := filepath.FromSlash(strings.TrimSpace(string(backlink)))
	for _, candidate := range append([]string{path}, endpoints...) {
		if worktreeroot.Same(registered, filepath.Join(candidate, ".git")) {
			return nil
		}
	}
	return fmt.Errorf("migration admin backlink names another tree: %s", registered)
}

// ensureMigrationRegistration repairs only an authenticated journal tree.
// Git's repair is idempotent; a permission or data error remains recoverable.
func ensureMigrationRegistration(root, path, other string, id TreeIdentity) error {
	if err := authenticateMigrationAdmin(root, path, id, other); err != nil {
		return err
	}
	paths, err := registeredWorktreePaths(root)
	if err != nil {
		return err
	}
	if isRegistered(paths, path) {
		return nil
	}
	if _, err := listGitOutput(root, "worktree", "repair", path); err != nil {
		return fmt.Errorf("repair partially moved worktree registration: %w", err)
	}
	paths, err = registeredWorktreePaths(root)
	if err != nil {
		return err
	}
	if !isRegistered(paths, path) {
		return fmt.Errorf("git did not register repaired worktree %s", path)
	}
	return nil
}
