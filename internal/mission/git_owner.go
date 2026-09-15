package mission

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type IntegrationLease struct {
	SessionID string `json:"session_id"`
	BaseSHA   string `json:"base_sha"`
}

func WriteIntegrationLease(path string, lease IntegrationLease) error {
	if validateMissionSessionID(lease.SessionID) != nil || strings.TrimSpace(lease.BaseSHA) == "" {
		return errors.New("git owner: invalid lease")
	}
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("git owner: unsafe lease path")
	}
	data, err := json.Marshal(lease)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

type GitEffect struct {
	Action              Action
	Repository          string
	IntegrationWorktree string
	WorktreeBranch      string
	ExplicitPaths       []string
	CommitMessage       string
	CardSHA             string
	BaseSHA             string
	LeasePath           string
	SessionID           string
}

type GitOwnerAdapter struct {
	GitBinary string
	Effect    GitEffect
}

func (o GitOwnerAdapter) Snapshot(ctx context.Context) (string, string, error) {
	repo, err := o.repository()
	if err != nil {
		return "", "", err
	}
	head, err := o.git(ctx, repo, "rev-parse", "HEAD")
	if err != nil {
		return "", "", err
	}
	branch, err := o.git(ctx, repo, "branch", "--show-current")
	return head, branch, err
}

func (o GitOwnerAdapter) ValidateIntegrationLease() error {
	repo, err := o.repository()
	if err != nil {
		return err
	}
	return o.validateLease(repo)
}

func (o GitOwnerAdapter) binary() string {
	if o.GitBinary != "" {
		return o.GitBinary
	}
	return "git"
}
func (o GitOwnerAdapter) git(ctx context.Context, dir string, args ...string) (string, error) {
	all := append([]string{"-C", dir}, args...)
	out, err := exec.CommandContext(ctx, o.binary(), all...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
func (o GitOwnerAdapter) repository() (string, error) {
	root, err := filepath.EvalSymlinks(o.Effect.Repository)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(root) {
		return "", errors.New("git owner: repository must be absolute")
	}
	if info, err := os.Stat(filepath.Join(root, ".git")); err != nil || (!info.IsDir() && !info.Mode().IsRegular()) {
		return "", errors.New("git owner: repository not initialized")
	}
	return root, nil
}
func operationTrailer(id string) string { return "MoAI-Operation: " + id }

func (o GitOwnerAdapter) Readback(ctx context.Context, operationID string) (bool, error) {
	repo, err := o.repository()
	if err != nil {
		return false, err
	}
	switch o.Effect.Action {
	case ActionCommit:
		out, err := o.git(ctx, repo, "log", "-1", "--format=%B")
		if err != nil {
			return false, err
		}
		return strings.Contains(out, operationTrailer(operationID)), nil
	case ActionLocalMerge:
		if o.Effect.CardSHA == "" {
			return false, errors.New("git owner: card sha missing")
		}
		_, err := o.git(ctx, repo, "merge-base", "--is-ancestor", o.Effect.CardSHA, "HEAD")
		if err != nil {
			return false, nil
		}
		out, err := o.git(ctx, repo, "log", "-1", "--format=%B")
		if err != nil {
			return false, err
		}
		return strings.Contains(out, operationTrailer(operationID)), nil
	default:
		return false, errors.New("git owner: unsupported action")
	}
}

func validGitPath(path string) bool {
	clean := filepath.Clean(path)
	return path != "" && !filepath.IsAbs(clean) && clean != "." && clean != ".." && !strings.HasPrefix(filepath.ToSlash(clean), "../") && !strings.HasPrefix(filepath.ToSlash(clean), ".git/")
}
func (o GitOwnerAdapter) validateLease(repo string) error {
	if o.Effect.IntegrationWorktree != "" {
		resolved, err := filepath.EvalSymlinks(o.Effect.IntegrationWorktree)
		if err != nil || resolved != repo {
			return errors.New("git owner: integration worktree mismatch")
		}
	}
	commonDir, err := o.git(context.Background(), repo, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil || !filepath.IsAbs(commonDir) {
		return errors.New("git owner: common directory unavailable")
	}
	commonDir, err = filepath.EvalSymlinks(commonDir)
	if err != nil {
		return err
	}
	leasePath, err := filepath.EvalSymlinks(o.Effect.LeasePath)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(commonDir, leasePath)
	if err != nil || rel == ".." || strings.HasPrefix(filepath.ToSlash(rel), "../") {
		return errors.New("git owner: lease outside repository")
	}
	info, err := os.Lstat(leasePath)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0o077 != 0 {
		return errors.New("git owner: lease missing or unsafe")
	}
	raw, err := os.ReadFile(leasePath)
	if err != nil {
		return err
	}
	var lease IntegrationLease
	if json.Unmarshal(raw, &lease) != nil || lease.SessionID != o.Effect.SessionID || lease.BaseSHA != o.Effect.BaseSHA {
		return errors.New("git owner: lease mismatch")
	}
	return nil
}

func (o GitOwnerAdapter) Apply(ctx context.Context, operationID string) error {
	if applied, err := o.Readback(ctx, operationID); err != nil {
		return err
	} else if applied {
		return errors.New("git owner: operation already applied")
	}
	repo, err := o.repository()
	if err != nil {
		return err
	}
	branch, err := o.git(ctx, repo, "branch", "--show-current")
	if err != nil {
		return err
	}
	switch o.Effect.Action {
	case ActionCommit:
		if branch != o.Effect.WorktreeBranch || !strings.HasPrefix(branch, "WT-") || len(o.Effect.ExplicitPaths) == 0 || strings.TrimSpace(o.Effect.CommitMessage) == "" {
			return errors.New("git owner: commit boundary invalid")
		}
		for _, path := range o.Effect.ExplicitPaths {
			if !validGitPath(path) {
				return errors.New("git owner: unsafe explicit path")
			}
		}
		args := append([]string{"add", "--"}, o.Effect.ExplicitPaths...)
		if _, err := o.git(ctx, repo, args...); err != nil {
			return err
		}
		_, err = o.git(ctx, repo, "commit", "-m", o.Effect.CommitMessage, "-m", operationTrailer(operationID))
		return err
	case ActionLocalMerge:
		if branch != "develop" || !strings.HasPrefix(o.Effect.WorktreeBranch, "WT-") || o.Effect.CardSHA == "" || o.Effect.BaseSHA == "" {
			return errors.New("git owner: merge boundary invalid")
		}
		if err := o.validateLease(repo); err != nil {
			return err
		}
		head, err := o.git(ctx, repo, "rev-parse", "HEAD")
		if err != nil || head != o.Effect.BaseSHA {
			return fmt.Errorf("git owner: stale integration base")
		}
		_, err = o.git(ctx, repo, "merge", "--no-ff", o.Effect.CardSHA, "-m", "merge "+o.Effect.WorktreeBranch, "-m", operationTrailer(operationID))
		return err
	default:
		return errors.New("git owner: unsupported action")
	}
}
