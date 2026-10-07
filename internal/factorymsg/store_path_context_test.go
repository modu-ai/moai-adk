package factorymsg

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The copied test executable acts as Git only in subprocesses explicitly armed by this fixture.
func brokerPathGitHelper() {
	realGit := os.Getenv("MOAI_TEST_BROKER_REAL_GIT")
	if realGit == "" || !strings.EqualFold(strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe"), "git") {
		return
	}
	mode := os.Getenv("MOAI_TEST_BROKER_GIT_MODE")
	if len(os.Args) > 1 && os.Args[1] == "--hold-pipe" {
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}
	if mode == "locale" {
		if logPath := os.Getenv("MOAI_TEST_BROKER_GIT_LOCALE_LOG"); logPath != "" {
			if file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600); err == nil {
				_, _ = file.WriteString(os.Getenv("LC_ALL") + "|" + os.Getenv("LANGUAGE") + "|" + strings.Join(os.Args[1:], " ") + "\n")
				_ = file.Close()
			}
		}
		if os.Getenv("LC_ALL") != "C" || os.Getenv("LANGUAGE") != "C" {
			_, _ = os.Stderr.WriteString("fatal: pas un dépôt git (ni aucun de ses parents) : .git\n")
			os.Exit(128)
		}
	}
	primary := strings.Contains(strings.Join(os.Args[1:], " "), "--path-format=absolute")
	if mode == "unknown" || (mode == "legacy" && primary) {
		_, _ = os.Stderr.WriteString("controlled Git failure\n")
		os.Exit(1)
	}
	if mode == "pipe" && primary {
		child := exec.Command(os.Args[0], "--hold-pipe")
		child.Stdout, child.Stderr = os.Stdout, os.Stderr
		if err := child.Start(); err != nil {
			os.Exit(1)
		}
		_ = os.WriteFile(os.Getenv("MOAI_TEST_BROKER_GIT_CHILD_PID"), []byte(strconv.Itoa(child.Process.Pid)), 0o600)
	}
	delay, _ := strconv.Atoi(os.Getenv("MOAI_TEST_BROKER_GIT_DELAY_MS"))
	if primary || (mode == "legacy" && strings.Contains(strings.Join(os.Args[1:], " "), "--absolute-git-dir")) {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	cmd := exec.Command(realGit, os.Args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			os.Exit(exit.ExitCode())
		}
		os.Exit(1)
	}
	os.Exit(0)
}

func slowBrokerGit(t *testing.T, delay int) {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	name := "git"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o700); err != nil {
		t.Fatal(err)
	}
	// The race instrumented helper must not spend the caller's budget in the
	// race runtime's artificial exit sleep; instrumentation remains enabled.
	t.Setenv("GORACE", os.Getenv("GORACE")+" atexit_sleep_ms=0")
	t.Setenv("MOAI_TEST_BROKER_REAL_GIT", realGit)
	t.Setenv("MOAI_TEST_BROKER_GIT_DELAY_MS", strconv.Itoa(delay))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func brokerGitRoot(t *testing.T) string {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	root := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	return root
}

func TestBrokerOpenContextBoundsGitPathResolution(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	root := brokerGitRoot(t)
	slowBrokerGit(t, 600)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	store, err := OpenWithContext(ctx, root, "run-slow")
	elapsed := time.Since(started)
	if store != nil {
		_ = store.Close()
		t.Fatal("expired path returned a store")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error=%v, want expired context", err)
	}
	t.Logf("slow Git path elapsed=%v", elapsed)
	if elapsed > 400*time.Millisecond {
		t.Errorf("Git path exceeded caller budget allowance: %v", elapsed)
	}
	if _, err := os.Stat(filepath.Join(home, "db")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expired path created state directory: %v", err)
	}
}

func TestBrokerOpenContextRechecksBudgetAfterGit(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := brokerGitRoot(t)
	path, err := BrokerPath(root, "run-lock")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.Exec("ROLLBACK"); _ = db.Close() })
	if _, err := db.Exec("PRAGMA journal_mode=WAL; BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	slowBrokerGit(t, 400)
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Millisecond)
	defer cancel()
	started := time.Now()
	store, err := OpenWithContext(ctx, root, "run-lock")
	elapsed := time.Since(started)
	if store != nil {
		_ = store.Close()
		t.Fatal("locked path returned a store")
	}
	if err == nil {
		t.Fatal("locked initialization succeeded")
	}
	t.Logf("delayed Git plus WAL lock elapsed=%v", elapsed)
	if elapsed > 650*time.Millisecond {
		t.Fatalf("busy wait used stale pre-path budget: %v", elapsed)
	}
}

func TestBrokerOpenContextUsesRemainingBusyTimeout(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := brokerGitRoot(t)
	slowBrokerGit(t, 600)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	store, err := OpenWithContext(ctx, root, "run-remaining")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	var busy int
	if err := store.db.QueryRow("PRAGMA busy_timeout").Scan(&busy); err != nil {
		t.Fatal(err)
	}
	t.Logf("actual busy_timeout after delayed path=%dms", busy)
	if busy >= 800 {
		t.Fatalf("busy_timeout still uses initial rather than remaining caller budget: %dms", busy)
	}
}

func TestBrokerOpenContextGitFailuresDoNotFallbackIdentity(t *testing.T) {
	for _, mode := range []string{"unknown", "legacy", "pipe"} {
		t.Run(mode, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("MOAI_HOME", home)
			root := brokerGitRoot(t)
			slowBrokerGit(t, 600)
			pidPath := filepath.Join(t.TempDir(), "child.pid")
			t.Setenv("MOAI_TEST_BROKER_GIT_MODE", mode)
			t.Setenv("MOAI_TEST_BROKER_GIT_CHILD_PID", pidPath)
			t.Cleanup(func() {
				if data, err := os.ReadFile(pidPath); err == nil {
					if pid, err := strconv.Atoi(string(data)); err == nil {
						if process, err := os.FindProcess(pid); err == nil {
							_ = process.Kill()
						}
					}
				}
			})
			budget := 200 * time.Millisecond
			if mode == "pipe" {
				budget = 500 * time.Millisecond
			}
			if mode == "unknown" {
				budget = 2 * time.Second
			}
			ctx, cancel := context.WithTimeout(context.Background(), budget)
			defer cancel()
			started := time.Now()
			store, err := OpenWithContext(ctx, root, "run-error")
			if store != nil {
				_ = store.Close()
				t.Fatal("uncertain Git identity returned a store")
			}
			if err == nil {
				t.Fatal("uncertain Git identity silently fell back")
			}
			elapsed := time.Since(started)
			t.Logf("mode=%s elapsed=%v error=%v", mode, elapsed, err)
			if mode == "unknown" && errors.Is(err, context.DeadlineExceeded) {
				t.Fatal("unknown Git error was masked by timeout")
			}
			if elapsed > budget+250*time.Millisecond {
				t.Fatalf("Git cancellation/pipe wait exceeded bound: %v", elapsed)
			}
			if mode == "pipe" {
				if _, err := os.Stat(pidPath); err != nil {
					t.Fatalf("inherited-pipe positive control did not start: %v", err)
				}
			}
			if _, err := os.Stat(filepath.Join(home, "db")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("uncertain Git identity created state: %v", err)
			}
		})
	}
}

func TestBrokerOpenContextRejectsCorruptGitFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	root := t.TempDir()
	missing := filepath.Join(t.TempDir(), "missing-metadata")
	if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: "+missing+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, nativeErr := exec.Command("git", "-C", root, "rev-parse", "--absolute-git-dir").CombinedOutput()
	if nativeErr == nil {
		t.Fatal("corrupt .git positive control unexpectedly succeeded")
	}
	t.Logf("actual Git corrupt .git error=%v output=%s", nativeErr, out)
	store, err := OpenWithContext(context.Background(), root, "run-corrupt")
	if store != nil {
		_ = store.Close()
		t.Error("corrupt Git identity returned a broker")
	}
	if err == nil {
		t.Error("corrupt Git identity silently fell back to a root hash")
	}
	if _, err := os.Stat(filepath.Join(home, "db")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("corrupt Git identity created broker state: %v", err)
	}
}

func TestBrokerOpenContextPinsGitErrorLocaleOnlyInChildren(t *testing.T) {
	t.Setenv("MOAI_HOME", "")
	root := t.TempDir()
	slowBrokerGit(t, 0)
	logPath := filepath.Join(t.TempDir(), "locale.log")
	t.Setenv("MOAI_TEST_BROKER_GIT_MODE", "locale")
	t.Setenv("MOAI_TEST_BROKER_GIT_LOCALE_LOG", logPath)
	t.Setenv("LC_ALL", "fr_FR.UTF-8")
	t.Setenv("LANGUAGE", "fr")
	out, err := exec.Command("git", "-C", root, "rev-parse", "--absolute-git-dir").CombinedOutput()
	if err == nil || !strings.Contains(string(out), "pas un dépôt git") {
		t.Fatalf("localized positive control absent: %v %s", err, out)
	}
	if err := os.WriteFile(logPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := OpenWithContext(context.Background(), root, "run-locale")
	if err != nil {
		t.Fatalf("ordinary nonGit directory rejected under caller locale: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if os.Getenv("LC_ALL") != "fr_FR.UTF-8" || os.Getenv("LANGUAGE") != "fr" {
		t.Fatal("caller locale was globally mutated")
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected core primary+fallback and homestate layout probes, got %q", data)
	}
	for _, line := range lines {
		if !strings.HasPrefix(line, "C|C|") {
			t.Fatalf("Git child locale not pinned: %q", line)
		}
	}
}
