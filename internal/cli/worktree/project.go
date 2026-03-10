package worktree

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// gitRemoteFunc resolves git remote origin URL for a directory.
// Overridable in tests.
var gitRemoteFunc = func(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// detectProjectName returns a short project identifier for use in worktree paths.
// Priority:
//  1. Last segment of the module path from go.mod
//  2. Repository name from git remote origin
//  3. Base name of the directory
func detectProjectName(dir string) string {
	if name := readGoModName(dir); name != "" {
		return name
	}
	if url, err := gitRemoteFunc(dir); err == nil {
		if name := repoNameFromURL(url); name != "" {
			return name
		}
	}
	return filepath.Base(dir)
}

// readGoModName reads the module name from go.mod and returns the last path segment.
func readGoModName(dir string) string {
	f, err := os.Open(filepath.Join(dir, "go.mod"))
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			module := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			parts := strings.Split(module, "/")
			return parts[len(parts)-1]
		}
	}
	return ""
}

// repoNameFromURL extracts the repository name from a git remote URL.
// Handles SSH (git@github.com:user/repo.git) and HTTPS (https://github.com/user/repo.git).
func repoNameFromURL(url string) string {
	url = strings.TrimSuffix(url, ".git")
	// Normalize SSH colon separator so filepath.Base works uniformly
	url = strings.ReplaceAll(url, ":", "/")
	return filepath.Base(url)
}
