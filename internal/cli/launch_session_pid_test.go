package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestWithSessionPID_StampsAndReplaces covers the three behaviors the launch
// stamp relies on: the variable is appended, an inherited value is replaced
// rather than duplicated, and unrelated entries survive untouched.
func TestWithSessionPID_StampsAndReplaces(t *testing.T) {
	want := config.EnvMoaiSessionPID + "=4242"

	t.Run("appends to a clean environment", func(t *testing.T) {
		got := withSessionPID([]string{"PATH=/usr/bin", "HOME=/home/x"}, 4242)
		if len(got) != 3 || got[2] != want {
			t.Fatalf("withSessionPID = %v, want PATH+HOME plus %q", got, want)
		}
	})

	t.Run("replaces an inherited value", func(t *testing.T) {
		in := []string{config.EnvMoaiSessionPID + "=999", "PATH=/usr/bin"}
		got := withSessionPID(in, 4242)

		var stamps []string
		for _, kv := range got {
			if strings.HasPrefix(kv, config.EnvMoaiSessionPID+"=") {
				stamps = append(stamps, kv)
			}
		}
		if len(stamps) != 1 || stamps[0] != want {
			t.Fatalf("stamps = %v, want exactly [%q] (a nested launch must not leave the outer PID behind)", stamps, want)
		}
		if !containsEnvEntry(got, "PATH=/usr/bin") {
			t.Errorf("unrelated entry dropped: %v", got)
		}
	})

	t.Run("non-positive pid is a no-op", func(t *testing.T) {
		in := []string{"PATH=/usr/bin"}
		for _, pid := range []int{0, -1} {
			got := withSessionPID(in, pid)
			if len(got) != 1 || got[0] != "PATH=/usr/bin" {
				t.Errorf("withSessionPID(pid=%d) = %v, want the input unchanged", pid, got)
			}
		}
	})
}

// TestWithSessionPID_ParsesBackAsSelf pins the stamp's format against the value
// the resolver actually reads: the recorded text must name this process's PID,
// which under execve(2) is the session's.
func TestWithSessionPID_ParsesBackAsSelf(t *testing.T) {
	got := withSessionPID(nil, os.Getpid())
	want := fmt.Sprintf("%s=%d", config.EnvMoaiSessionPID, os.Getpid())
	if len(got) != 1 || got[0] != want {
		t.Fatalf("withSessionPID(self) = %v, want [%q]", got, want)
	}
}

// TestSessionPIDStamp_PlatformBoundary is the structural guard for the rule
// that only a caller which KNOWS the session PID may stamp it.
//
//   - POSIX: execve(2) replaces this process, so os.Getpid() IS the session —
//     the stamp belongs there.
//   - Windows: the session is a spawned child whose PID does not exist yet, so
//     stamping would record the launcher, which exits with the child. That is
//     the dead-PID defect the resolver exists to avoid.
func TestSessionPIDStamp_PlatformBoundary(t *testing.T) {
	cliDir := sessionPIDPackageDir(t)

	posix := readSessionPIDSource(t, filepath.Join(cliDir, "launch_exec_posix.go"))
	if !strings.Contains(posix, "withSessionPID(env, os.Getpid())") {
		t.Errorf("launch_exec_posix.go must stamp the session PID via withSessionPID(env, os.Getpid()); " +
			"after execve(2) this process IS the session")
	}

	win := readSessionPIDSource(t, filepath.Join(cliDir, "launch_exec_windows.go"))
	if strings.Contains(win, "withSessionPID(") {
		t.Errorf("launch_exec_windows.go must NOT stamp the session PID: the session is the spawned " +
			"child, and this process's PID names a launcher that exits with it")
	}
}

// TestSessionPIDStamp_NotSetFromHooks encodes the hard boundary from the other
// side: a hook subprocess exits within milliseconds of running, so a PID it
// declares is dead before any reader probes it. No hook may write the variable.
func TestSessionPIDStamp_NotSetFromHooks(t *testing.T) {
	hookDir := filepath.Join(filepath.Dir(sessionPIDPackageDir(t)), "hook")

	var offenders []string
	err := filepath.WalkDir(hookDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		src := readSessionPIDSource(t, path)
		if strings.Contains(src, config.EnvMoaiSessionPID) || strings.Contains(src, "EnvMoaiSessionPID") {
			offenders = append(offenders, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", hookDir, err)
	}
	if len(offenders) > 0 {
		t.Errorf("hook sources must not set %s (a hook PID is dead on arrival): %v",
			config.EnvMoaiSessionPID, offenders)
	}
}

func sessionPIDPackageDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file path")
	}
	return filepath.Dir(thisFile)
}

func readSessionPIDSource(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func containsEnvEntry(env []string, want string) bool {
	for _, kv := range env {
		if kv == want {
			return true
		}
	}
	return false
}
