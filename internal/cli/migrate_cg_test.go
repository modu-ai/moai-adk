package cli

import (
	"bytes"
	"errors"
	"github.com/modu-ai/moai-adk/internal/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const migrationCGInput = "llm:\n  team_mode: cg # retain\n  credential_ref: KEEP_REFERENCE\n"

func cgProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	p := filepath.Join(root, ".moai/config/sections")
	if err := os.MkdirAll(p, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "llm.yaml"), []byte(migrationCGInput), 0600); err != nil {
		t.Fatal(err)
	}
	return root
}
func TestCGMigrationPreviewApplyAndGuards(t *testing.T) {
	root := cgProject(t)
	var out bytes.Buffer
	run := func(args ...string) error {
		cmd := newMigrateCGCommand(func() (string, error) { return root, nil })
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs(args)
		return cmd.Execute()
	}
	for _, args := range [][]string{nil, {"--target", "claude-only"}} {
		if err := run(args...); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".moai/backups")); !os.IsNotExist(err) {
		t.Fatal("preview wrote backup")
	}
	for _, args := range [][]string{{"--apply"}, {"--apply", "--target", "claude-only"}, {"--apply", "--target", "claude-glm"}, {"--target", "claude-glm", "--accept-role-change"}, {"--wat"}, {"--target"}} {
		if err := run(args...); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	original, _ := os.ReadFile(filepath.Join(root, ".moai/config/sections/llm.yaml"))
	if string(original) != migrationCGInput {
		t.Fatal("failed gate wrote source")
	}
	if err := run("--target", "claude-only", "--apply", "--accept-role-change"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Applied") || strings.Contains(out.String(), "KEEP_REFERENCE") {
		t.Fatalf("output %s", out.String())
	}
	result, err := applyCGMigration(root, "claude-only", cgMigrationIO{})
	if err != nil || !result.Unchanged {
		t.Fatalf("idempotence %+v %v", result, err)
	}
	backups, err := filepath.Glob(filepath.Join(root, ".moai/backups/cg-migration/*.yaml"))
	if err != nil || len(backups) != 1 {
		t.Fatalf("backups %v %v", backups, err)
	}
	backup, _ := os.ReadFile(backups[0])
	info, _ := os.Stat(backups[0])
	if string(backup) != migrationCGInput || info.Mode().Perm() != 0600 {
		t.Fatal("not exact private backup")
	}
	if err := guardCGLaunchAt(root, "claude"); err != nil {
		t.Fatal(err)
	}
	if _, err := applyCGMigration(root, "claude-glm", cgMigrationIO{}); err == nil {
		t.Fatal("opposite accepted")
	}
}
func TestCGMigrationTransactionFailures(t *testing.T) {
	for _, stage := range []string{"concurrent", "lock", "backup", "write", "replace", "readback"} {
		t.Run(stage, func(t *testing.T) {
			root := cgProject(t)
			source := filepath.Join(root, ".moai/config/sections/llm.yaml")
			failure := errors.New("injected failure")
			ops := cgMigrationIO{}
			switch stage {
			case "concurrent":
				ops.BeforeLock = func() { _ = os.WriteFile(source, []byte("llm: {team_mode: glm}\n"), 0600) }
			case "lock":
				_ = os.WriteFile(source+".cg-migration.lock", []byte("owned elsewhere"), 0600)
			case "backup":
				_ = os.MkdirAll(filepath.Join(root, ".moai/backups"), 0700)
				_ = os.Symlink(t.TempDir(), filepath.Join(root, ".moai/backups/cg-migration"))
			case "write":
				ops.WriteTemp = func(*os.File, []byte) error { return failure }
			case "replace":
				ops.Replace = func(string, string) error { return failure }
			case "readback":
				ops.ReadBack = func(string) ([]byte, error) { return nil, failure }
			}
			result, err := applyCGMigration(root, "claude-only", ops)
			if err == nil {
				t.Fatal("failure claimed success")
			}
			actual, _ := os.ReadFile(source)
			if stage == "readback" {
				if result.Backup == "" || !strings.Contains(string(actual), "team_mode: claude") {
					t.Fatalf("readback boundary %+v %s", result, actual)
				}
			} else if stage == "concurrent" {
				if !strings.Contains(string(actual), "team_mode: glm") {
					t.Fatal("overwrote concurrent source")
				}
			} else if string(actual) != migrationCGInput {
				t.Fatal("failed transaction altered source")
			}
			temps, _ := filepath.Glob(source + ".tmp-*")
			if len(temps) != 0 {
				t.Fatal("temp left behind")
			}
		})
	}
}
func TestCGLaunchGuardPrecedesEntryEffects(t *testing.T) {
	root := cgProject(t)
	t.Chdir(root)
	for _, mode := range []string{"claude", "glm", "gpt"} {
		if err := guardCGLaunchAt(root, mode); err == nil {
			t.Fatal("legacy accepted")
		}
	}
	raw := []byte("llm: {team_mode: claude, gateway: {teammate_mode: tmux, teammate_provider: glm, verified: true}}\n")
	_ = os.WriteFile(filepath.Join(root, ".moai/config/sections/llm.yaml"), raw, 0600)
	for _, mode := range []string{"claude", "glm", "gpt"} {
		if err := guardCGLaunchAt(root, mode); err == nil {
			t.Fatal("unverified hybrid accepted")
		}
	}
}

func TestCGEntryGuardRunsBeforeLaunchAndSpawn(t *testing.T) {
	root := cgProject(t)
	t.Chdir(root)
	oldRoot, oldLaunch, oldTmux, oldSpawn, oldLook := findProjectRootFn, unifiedLaunchFunc, inTmuxFn, tmuxSpawnFn, spawnLookPath
	defer func() {
		findProjectRootFn, unifiedLaunchFunc, inTmuxFn, tmuxSpawnFn, spawnLookPath = oldRoot, oldLaunch, oldTmux, oldSpawn, oldLook
	}()
	findProjectRootFn = func() (string, error) { return root, nil }
	launches, spawns := 0, 0
	launch := func(string, string, []string) error { launches++; return nil }
	unifiedLaunchFunc = launch
	inTmuxFn = func() bool { return true }
	spawnLookPath = func(string) (string, error) { return "/mock/moai", nil }
	tmuxSpawnFn = func(string, string) (string, error) { spawns++; return "%1", nil }
	for _, name := range []string{"cc", "glm", "gpt"} {
		for _, args := range [][]string{{"--model", "opus"}, {"--continue"}, {"--resume", "session"}, {"--spawn"}, {"-w"}, {"-k", "-p"}, {"-f", "-p"}} {
			var err error
			cmd := newGPTCommand(gptCommandServices{Launch: launch})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			switch name {
			case "cc":
				err = runCC(cmd, args)
			case "glm":
				err = runGLM(cmd, args)
			case "gpt":
				err = cmd.RunE(cmd, args)
			}
			if !errors.Is(err, config.ErrLegacyCG) {
				t.Fatalf("%s %v: expected legacy guard, got %v (launch=%d spawn=%d)", name, args, err, launches, spawns)
			}
		}
	}
	if launches != 0 || spawns != 0 {
		t.Fatalf("guard allowed effects: launch=%d spawn=%d", launches, spawns)
	}
	// A normal control reaches the same prepared launch seam. This also shows
	// the counter is live and would detect removing the entry guard.
	if err := os.WriteFile(filepath.Join(root, ".moai/config/sections/llm.yaml"), []byte("llm: {team_mode: claude}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := newGPTCommand(gptCommandServices{Launch: launch})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := runCC(cmd, []string{"--model", "opus"}); err != nil {
		t.Fatal(err)
	}
	if launches != 1 {
		t.Fatalf("normal control did not reach launch seam: %d", launches)
	}
}

func TestCGMigrationAvoidsTypedDependencyDecode(t *testing.T) {
	if !isTrivialCommand([]string{"migrate", "cg", "--apply"}) {
		t.Fatal("cg migration initializes typed dependency graph")
	}
	if isTrivialCommand([]string{"migrate", "agency"}) {
		t.Fatal("unrelated migration lost dependency initialization")
	}
}

func TestCGMigrationBackupReuseAndCorruption(t *testing.T) {
	root := cgProject(t)
	source := filepath.Join(root, ".moai/config/sections/llm.yaml")
	first, err := applyCGMigration(root, "claude-only", cgMigrationIO{Replace: func(string, string) error { return errors.New("stop before replace") }})
	if err == nil || first.Backup == "" {
		t.Fatal("expected backup before failure")
	}
	result, err := applyCGMigration(root, "claude-only", cgMigrationIO{})
	if err != nil || result.Backup != first.Backup {
		t.Fatalf("exact backup not reused %+v %v", result, err)
	}
	for _, kind := range []string{"different", "public", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			_ = os.WriteFile(source, []byte(migrationCGInput), 0600)
			switch kind {
			case "different":
				_ = os.WriteFile(first.Backup, []byte("wrong"), 0600)
			case "public":
				_ = os.Chmod(first.Backup, 0644)
			case "symlink":
				_ = os.Remove(first.Backup)
				_ = os.Symlink(source, first.Backup)
			}
			if _, err := applyCGMigration(root, "claude-only", cgMigrationIO{}); err == nil {
				t.Fatal("unsafe backup accepted")
			}
		})
	}
}
func TestCGMigrationSourceAndCommandErrors(t *testing.T) {
	root := cgProject(t)
	source := filepath.Join(root, ".moai/config/sections/llm.yaml")
	var out bytes.Buffer
	command := func(rootFn func() (string, error), args ...string) error {
		cmd := newMigrateCGCommand(rootFn)
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs(args)
		return cmd.Execute()
	}
	rootFn := func() (string, error) { return root, nil }
	if err := command(rootFn, "--target", "invalid"); err == nil {
		t.Fatal("invalid target")
	}
	if err := command(func() (string, error) { return "", errors.New("missing root") }); err == nil {
		t.Fatal("root failure")
	}
	if _, err := applyCGMigration(root, "claude-only", cgMigrationIO{}); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"--target", "claude-only"}, {"--target", "claude-only", "--apply", "--accept-role-change"}} {
		if err := command(rootFn, args...); err != nil {
			t.Fatal(err)
		}
	}
	_ = os.WriteFile(source, []byte("llm: {team_mode: cg, mode: glm}"), 0600)
	if err := command(rootFn); err == nil {
		t.Fatal("conflicting preview")
	}
	if err := command(rootFn, "--target", "claude-only", "--apply", "--accept-role-change"); err == nil {
		t.Fatal("conflicting apply")
	}
	if err := guardCGLaunchAt(root, "claude"); err == nil {
		t.Fatal("guard cg")
	}
	_ = os.WriteFile(source, []byte("llm: {team_mode: claude, gateway: {teammate_mode: tmux}}"), 0600)
	if err := guardCGLaunchAt(root, "claude"); err == nil {
		t.Fatal("invalid policy")
	}
	_ = os.WriteFile(source, bytes.Repeat([]byte("x"), 1<<20+1), 0600)
	if _, err := readCGSource(source); err == nil {
		t.Fatal("unbounded source")
	}
	_ = os.Remove(source)
	if err := guardCGLaunchAt(root, "claude"); err != nil {
		t.Fatal(err)
	}
	if err := command(rootFn); err == nil {
		t.Fatal("missing source preview")
	}
	if _, err := applyCGMigration(root, "claude-only", cgMigrationIO{}); err == nil {
		t.Fatal("missing source apply")
	}
	_ = os.Symlink(filepath.Join(root, "missing"), source)
	if _, err := readCGSource(source); err == nil {
		t.Fatal("symlink source")
	}
	if err := guardCGLaunchAt(root, "claude"); err == nil {
		t.Fatal("symlink guard")
	}
	old := findProjectRootFn
	defer func() { findProjectRootFn = old }()
	findProjectRootFn = func() (string, error) { return "", errors.New("root failed") }
	if err := guardCGLaunchMode("claude"); err == nil {
		t.Fatal("root guard")
	}
}
func TestCGMigrationPreparedByteAndLateWriteFailures(t *testing.T) {
	for _, stage := range []string{"temp-diff", "late-write", "readback-diff", "private-parent"} {
		t.Run(stage, func(t *testing.T) {
			root := cgProject(t)
			source := filepath.Join(root, ".moai/config/sections/llm.yaml")
			ops := cgMigrationIO{}
			switch stage {
			case "temp-diff":
				ops.WriteTemp = func(f *os.File, _ []byte) error { _, err := f.WriteString("wrong"); return err }
			case "late-write":
				ops.WriteTemp = func(f *os.File, b []byte) error {
					_ = os.WriteFile(source, []byte("llm: {team_mode: glm}"), 0600)
					_, err := f.Write(b)
					return err
				}
			case "readback-diff":
				ops.ReadBack = func(string) ([]byte, error) { return []byte("wrong"), nil }
			case "private-parent":
				_ = os.WriteFile(filepath.Join(root, ".moai/backups"), []byte("file"), 0600)
			}
			if _, err := applyCGMigration(root, "claude-only", ops); err == nil {
				t.Fatal("invalid transaction accepted")
			}
		})
	}
}
