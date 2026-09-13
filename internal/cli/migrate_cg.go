package cli

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/spf13/cobra"
)

func init() { migrateCmd.AddCommand(newMigrateCGCommand(findProjectRoot)) }

func newMigrateCGCommand(rootFn func() (string, error)) *cobra.Command {
	var target string
	var apply, accept bool
	cmd := &cobra.Command{Use: "cg", Short: "Preview migration of legacy CG teammate roles", Args: cobra.NoArgs, SilenceUsage: true}
	cmd.Flags().StringVar(&target, "target", "", "Migration target: claude-only or claude-glm")
	cmd.Flags().BoolVar(&apply, "apply", false, "Apply the reviewed migration")
	cmd.Flags().BoolVar(&accept, "accept-role-change", false, "Accept removal of automatic GLM teammate assignment")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		if target != "" && target != "claude-only" && target != "claude-glm" {
			return errors.New("target must be claude-only or claude-glm")
		}
		if accept && target != "claude-only" {
			return errors.New("--accept-role-change requires --target claude-only")
		}
		if apply && target == "" {
			return errors.New("--apply requires --target")
		}
		if apply && target == "claude-only" && !accept {
			return errors.New("claude-only removes automatic GLM teammate assignment; add --accept-role-change to apply")
		}
		if apply && target == "claude-glm" {
			return errors.New("claude-glm migration is unavailable until teammate routing capability is verified")
		}
		root, err := rootFn()
		if err != nil {
			return err
		}
		if !apply {
			raw, err := readCGSource(filepath.Join(root, ".moai/config/sections/llm.yaml"))
			if err != nil {
				return err
			}
			targets := []string{target}
			if target == "" {
				targets = []string{"claude-only", "claude-glm"}
			}
			for _, candidate := range targets {
				plan, err := config.PlanCGMigration(raw, candidate)
				if err != nil {
					return err
				}
				if plan.Unchanged {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Unchanged: %s already selected.\n", candidate); err != nil {
						return err
					}
					continue
				}
				if candidate == "claude-only" {
					if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Preview claude-only: team_mode=claude, teammates=in-process/inherit. Automatic GLM teammate assignment is removed. Applying requires --accept-role-change."); err != nil {
						return err
					}
				} else {
					if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Preview claude-glm: team_mode=claude, teammates=tmux/glm. Applying and launching require verified teammate routing capability; currently unavailable."); err != nil {
						return err
					}
				}
			}
			return nil
		}
		result, err := applyCGMigration(root, target, cgMigrationIO{})
		if err != nil {
			if result.Backup != "" {
				return fmt.Errorf("migration did not confirm success; exact backup: %s: %w", result.Backup, err)
			}
			return err
		}
		if result.Unchanged {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Unchanged: the requested teammate policy is already saved.")
		} else {
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Applied claude-only: automatic GLM teammate assignment removed. Backup: %s\n", result.Backup)
		}
		return err
	}
	return cmd
}

func guardCGLaunchMode(mode string) error {
	root, err := findProjectRootFn()
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}
	return guardCGLaunchAt(root, mode)
}
func guardCGLaunchAt(root, mode string) error {
	raw, err := readCGSource(filepath.Join(root, ".moai/config/sections/llm.yaml"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := config.GuardLegacyCG(raw); err != nil {
		return err
	}
	policy, err := config.ReadGatewayTeammatePolicy(raw)
	if err != nil {
		return err
	}
	if policy.Present && policy.Mode == "tmux" {
		if mode != "claude" {
			return errors.New("saved GLM teammate policy requires moai cc")
		}
		return errors.New("saved GLM teammate policy requires verified teammate routing capability; launch is unavailable")
	}
	return nil
}

// Test-local fault injection does not expose capability overrides or user flags.
type cgMigrationIO struct {
	BeforeLock func()
	WriteTemp  func(*os.File, []byte) error
	Replace    func(string, string) error
	ReadBack   func(string) ([]byte, error)
}
type cgMigrationResult struct {
	Backup    string
	Unchanged bool
}

func readCGSource(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("migration source must be a regular file")
	}
	if info.Size() > 1<<20 {
		return nil, errors.New("migration source exceeds 1 MiB")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }() // read-only source; Close carries no write-back to lose
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return nil, errors.New("migration source changed while opening")
	}
	raw, err := io.ReadAll(io.LimitReader(f, (1<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > 1<<20 {
		return nil, errors.New("migration source exceeds 1 MiB")
	}
	after, err := os.Lstat(path)
	if err != nil || !os.SameFile(opened, after) {
		return nil, errors.New("migration source changed while reading")
	}
	return raw, nil
}

func cgPrivateDirs(root, path string) error {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	current := root
	for _, part := range splitCGPath(relative) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			if err = os.Mkdir(current, 0700); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("migration directory must not be a symlink")
		}
	}
	return nil
}
func splitCGPath(path string) []string {
	var parts []string
	for path != "." && path != "" {
		dir, base := filepath.Split(path)
		parts = append([]string{base}, parts...)
		path = filepath.Clean(dir)
	}
	return parts
}

// applyCGMigration serializes only this migration. Hash rechecks detect writers that
// do not honor its lock. An existing lock is never stolen automatically.
func applyCGMigration(root, target string, ops cgMigrationIO) (result cgMigrationResult, err error) {
	if target != "claude-only" {
		return result, errors.New("only claude-only is currently applicable")
	}
	source := filepath.Join(root, ".moai/config/sections/llm.yaml")
	original, err := readCGSource(source)
	if err != nil {
		return result, err
	}
	plan, err := config.PlanCGMigration(original, target)
	if err != nil {
		return result, err
	}
	if plan.Unchanged {
		result.Unchanged = true
		return result, nil
	}
	hash := sha256.Sum256(original)
	if ops.BeforeLock != nil {
		ops.BeforeLock()
	}
	lockPath := source + ".cg-migration.lock"
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return result, errors.New("migration lock is already held or unavailable")
	}
	lockInfo, err := lock.Stat()
	if err != nil {
		_ = lock.Close() // lock release is governed by Remove + process exit, not Close
		return result, err
	}
	defer func() {
		_ = lock.Close() // lock release is governed by Remove + process exit, not Close
		if info, e := os.Lstat(lockPath); e == nil && os.SameFile(lockInfo, info) {
			_ = os.Remove(lockPath)
		}
	}()
	current, err := readCGSource(source)
	if err != nil {
		return result, err
	}
	if sha256.Sum256(current) != hash {
		return result, errors.New("configuration changed before migration lock")
	}
	backupDir := filepath.Join(root, ".moai/backups/cg-migration")
	if err := cgPrivateDirs(root, backupDir); err != nil {
		return result, err
	}
	backup := filepath.Join(backupDir, fmt.Sprintf("%x.yaml", hash))
	backupFile, err := os.OpenFile(backup, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		info, e := os.Lstat(backup)
		if e != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
			return result, errors.New("existing backup is not private and regular")
		}
		saved, e := readCGSource(backup)
		if e != nil || !bytes.Equal(saved, original) {
			return result, errors.New("existing backup does not match original bytes")
		}
	} else if err != nil {
		return result, err
	} else {
		_, err = backupFile.Write(original)
		if err == nil {
			err = backupFile.Sync()
		}
		closeErr := backupFile.Close()
		if err != nil {
			return result, err
		}
		if closeErr != nil {
			return result, closeErr
		}
	}
	result.Backup = backup
	temp, err := os.CreateTemp(filepath.Dir(source), "llm.yaml.tmp-")
	if err != nil {
		return result, err
	}
	tempPath := temp.Name()
	// The explicit Close below owns the meaningful error; this deferred Close is
	// only an error-path safety net (a second Close on the same file is a no-op).
	defer func() { _ = temp.Close(); _ = os.Remove(tempPath) }()
	if ops.WriteTemp != nil {
		err = ops.WriteTemp(temp, plan.Bytes)
	} else {
		_, err = temp.Write(plan.Bytes)
	}
	if err != nil {
		return result, err
	}
	if err = temp.Sync(); err != nil {
		return result, err
	}
	if err = temp.Close(); err != nil {
		return result, err
	}
	prepared, err := readCGSource(tempPath)
	if err != nil {
		return result, err
	}
	if !bytes.Equal(prepared, plan.Bytes) {
		return result, errors.New("migration temporary bytes differ")
	}
	if _, err = config.ReadGatewayTeammatePolicy(prepared); err != nil {
		return result, err
	}
	current, err = readCGSource(source)
	if err != nil {
		return result, err
	}
	if sha256.Sum256(current) != hash {
		return result, errors.New("configuration changed before atomic replacement")
	}
	replace := ops.Replace
	if replace == nil {
		replace = atomicfile.Replace
	}
	if err = replace(tempPath, source); err != nil {
		return result, err
	}
	readBack := ops.ReadBack
	if readBack == nil {
		readBack = readCGSource
	}
	written, err := readBack(source)
	if err != nil {
		return result, err
	}
	if !bytes.Equal(written, plan.Bytes) {
		return result, errors.New("migration readback differs from intended bytes")
	}
	_, err = config.ReadGatewayTeammatePolicy(written)
	return result, err
}
