package sync

import (
	"bytes"
	"os/exec"
	"strings"

	"log/slog"
)

// provenanceUnknown is the fail-open placeholder for git failures.
const provenanceUnknown = "<unknown>"

// CurrentProvenance returns the git provenance for the working-tree HEAD
// (REQ-NS-009). ExtractCommitSHA = `git rev-parse HEAD`; CapturedAt = the
// committer date of that SHA (`git log -1 --format=%cI`). Fail-open: on any
// git error, returns "<unknown>" values (never aborts).
//
// No wall-clock timestamp is used, so two runs on the same HEAD produce
// byte-identical output. Carried forward from 003's
// `internal/navigator/astx/enrich.go:83 CurrentProvenance`.
func CurrentProvenance(projectRoot string) Provenance {
	sha := gitOutput(projectRoot, "rev-parse", "HEAD")
	if sha == "" {
		sha = provenanceUnknown
	}
	captured := gitOutput(projectRoot, "log", "-1", "--format=%cI", sha)
	if captured == "" {
		captured = gitOutput(projectRoot, "log", "-1", "--format=%cI")
	}
	if captured == "" {
		captured = provenanceUnknown
	}
	return Provenance{ExtractCommitSHA: sha, CapturedAt: captured}
}

// gitOutput runs a git command in dir (empty dir = inherit cwd) and returns
// the trimmed stdout. Any error → "" (fail-open, logged at debug).
func gitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		slog.Debug("sync: git command failed", "args", args, "dir", dir, "error", err)
		return ""
	}
	return strings.TrimSpace(out.String())
}
