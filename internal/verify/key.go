package verify

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/execerr"
)

// digestHexLen is the hex length of the tree-state digest suffix in a key
// ("<head-sha>:<digest[:16]>").
const digestHexLen = 16

// Key computes the working-tree snapshot key binding a snapshot to the exact
// tree state it measured. Four inputs, each covering a distinct invalidation
// class:
//
//   - HEAD commit SHA — commit advance/switch.
//   - `git status --porcelain=v2` digest — file-set shape: staged/unstaged
//     delta listing and untracked non-ignored paths.
//   - `git diff HEAD` content hash — worktree content deltas of tracked files.
//     This leg is load-bearing: porcelain-v2 output is byte-identical across
//     successive edits to an already-dirty tracked file, so without the diff
//     hash a re-edit of a dirty file would falsely read as fresh.
//   - non-ignored untracked path/content hash — untracked file contents. Git
//     status only exposes the path, so without this leg re-editing an untracked
//     file would falsely read as fresh.
//
// Cost is constant w.r.t. repository history size (four git subprocesses;
// `git diff HEAD` scales with the dirty-delta size, not history). The caller
// owns the time-box: pass a deadline-bound ctx on latency-sensitive paths and
// fall back to re-execution on error.
func Key(ctx context.Context, repoDir string) (string, error) {
	head, err := gitOutput(ctx, repoDir, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("verify key: rev-parse HEAD: %w", err)
	}
	porcelain, err := gitOutput(ctx, repoDir, "status", "--porcelain=v2")
	if err != nil {
		return "", fmt.Errorf("verify key: status porcelain: %w", err)
	}
	diff, err := gitOutput(ctx, repoDir, "diff", "HEAD")
	if err != nil {
		return "", fmt.Errorf("verify key: diff HEAD: %w", err)
	}
	untracked, err := gitOutput(ctx, repoDir, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return "", fmt.Errorf("verify key: untracked files: %w", err)
	}
	h := sha256.New()
	h.Write([]byte(porcelain))
	h.Write([]byte{0}) // separator: porcelain/diff boundary is unambiguous
	h.Write([]byte(diff))
	for _, name := range strings.Split(untracked, "\x00") {
		if name == "" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(repoDir, filepath.FromSlash(name)))
		if err != nil {
			// A file can disappear between ls-files and ReadFile. Returning an
			// error makes the caller re-execute rather than cache a partial key.
			return "", fmt.Errorf("verify key: read untracked %q: %w", name, err)
		}
		h.Write([]byte(name))
		h.Write([]byte{':'})
		h.Write(content)
		h.Write([]byte{0})
	}
	digest := hex.EncodeToString(h.Sum(nil))[:digestHexLen]
	return strings.TrimSpace(head) + ":" + digest, nil
}

// gitOutput runs one git subcommand in repoDir and returns its stdout. The
// context bounds the subprocess (CommandContext kills it on deadline/cancel).
func gitOutput(ctx context.Context, repoDir string, args ...string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", repoDir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		// execerr.StatusDetail, not %w: a raw *exec.ExitError chain would be
		// mistaken for an intentional ExitCoder at the cmd/moai seam (t130).
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), execerr.StatusDetail(err))
	}
	return string(out), nil
}
