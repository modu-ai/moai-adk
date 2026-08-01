package cli

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// withSpawnStubs installs deterministic tmux stubs and restores the production
// bindings when the test ends. Returns a pointer to the captured invocation.
type spawnCapture struct {
	cwd     string
	command string
	calls   int
}

func withSpawnStubs(t *testing.T, inTmux bool, paneID string, spawnErr error) *spawnCapture {
	t.Helper()

	origInTmux := inTmuxFn
	origSpawn := tmuxSpawnFn
	t.Cleanup(func() {
		inTmuxFn = origInTmux
		tmuxSpawnFn = origSpawn
	})

	cap := &spawnCapture{}
	inTmuxFn = func() bool { return inTmux }
	tmuxSpawnFn = func(cwd, command string) (string, error) {
		cap.cwd = cwd
		cap.command = command
		cap.calls++
		return paneID, spawnErr
	}
	return cap
}

// requireTmuxBinary skips when tmux is absent. spawnLaunch checks LookPath
// before doing anything else, so the happy-path assertions cannot run without
// it — and stubbing LookPath would stop the test from covering that check.
func requireTmuxBinary(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not installed; spawnLaunch cannot reach the tmux stub")
	}
	if _, err := exec.LookPath("moai"); err != nil {
		t.Skip("moai not in PATH; spawnLaunch cannot reach the tmux stub")
	}
}

func TestStripSpawnFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		wantArgs  []string
		wantFound bool
	}{
		{
			name:      "absent",
			args:      []string{"-w", "feat"},
			wantArgs:  []string{"-w", "feat"},
			wantFound: false,
		},
		{
			name:      "trailing",
			args:      []string{"-w", "feat", "--spawn"},
			wantArgs:  []string{"-w", "feat"},
			wantFound: true,
		},
		{
			name:      "leading",
			args:      []string{"--spawn", "-w", "feat"},
			wantArgs:  []string{"-w", "feat"},
			wantFound: true,
		},
		{
			name:      "middle preserves order",
			args:      []string{"-p", "work", "--spawn", "-w", "feat"},
			wantArgs:  []string{"-p", "work", "-w", "feat"},
			wantFound: true,
		},
		{
			name:      "after pass-through marker is claude's, not ours",
			args:      []string{"-w", "feat", "--", "--spawn"},
			wantArgs:  []string{"-w", "feat", "--", "--spawn"},
			wantFound: false,
		},
		{
			name:      "before marker stripped, after marker kept",
			args:      []string{"--spawn", "--", "--spawn", "--print"},
			wantArgs:  []string{"--", "--spawn", "--print"},
			wantFound: true,
		},
		{
			name:      "empty args",
			args:      []string{},
			wantArgs:  []string{},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotArgs, gotFound := stripSpawnFlag(tt.args)
			if gotFound != tt.wantFound {
				t.Errorf("found = %v, want %v", gotFound, tt.wantFound)
			}
			if strings.Join(gotArgs, " ") != strings.Join(tt.wantArgs, " ") {
				t.Errorf("args = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestBuildSpawnCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		subcommand string
		args       []string
		want       string
	}{
		{
			name:       "no args",
			subcommand: "cg",
			args:       nil,
			want:       "moai cg",
		},
		{
			name:       "worktree short name",
			subcommand: "cg",
			args:       []string{"-w", "feat-auth"},
			want:       "moai cg -w feat-auth",
		},
		{
			name:       "profile and worktree",
			subcommand: "cc",
			args:       []string{"-p", "work", "-w", "feat-auth"},
			want:       "moai cc -p work -w feat-auth",
		},
		{
			name:       "name with space is quoted",
			subcommand: "cc",
			args:       []string{"-w", "my feat"},
			want:       "moai cc -w 'my feat'",
		},
		{
			name:       "shell metacharacters are neutralized",
			subcommand: "cc",
			args:       []string{"-w", "a;rm -rf /"},
			want:       `moai cc -w 'a;rm -rf /'`,
		},
		{
			name:       "embedded single quote",
			subcommand: "cc",
			args:       []string{"-w", "it's"},
			want:       `moai cc -w 'it'\''s'`,
		},
		{
			name:       "pass-through marker survives",
			subcommand: "glm",
			args:       []string{"-w", "feat", "--", "--print"},
			want:       "moai glm -w feat -- --print",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := buildSpawnCommand(tt.subcommand, tt.args); got != tt.want {
				t.Errorf("buildSpawnCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSpawnLaunchOpensWindow is the core contract: --spawn must hand tmux the
// same command minus --spawn, rooted at the project directory, and must NOT
// replace the current process (spawnLaunch returns normally).
func TestSpawnLaunchOpensWindow(t *testing.T) {
	requireTmuxBinary(t)

	tmpRoot := t.TempDir()
	origRoot := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tmpRoot, nil }
	t.Cleanup(func() { findProjectRootFn = origRoot })

	cap := withSpawnStubs(t, true, "%7", nil)

	var out bytes.Buffer
	if err := spawnLaunch(&out, "cg", []string{"-w", "feat-auth"}); err != nil {
		t.Fatalf("spawnLaunch() error = %v, want nil", err)
	}

	if cap.calls != 1 {
		t.Fatalf("tmux spawn calls = %d, want 1", cap.calls)
	}
	if cap.command != "moai cg -w feat-auth" {
		t.Errorf("command = %q, want %q", cap.command, "moai cg -w feat-auth")
	}
	if cap.cwd != tmpRoot {
		t.Errorf("cwd = %q, want project root %q", cap.cwd, tmpRoot)
	}
	if !strings.Contains(out.String(), "%7") {
		t.Errorf("stdout should report the pane id for the user; got:\n%s", out.String())
	}
}

// TestSpawnLaunchRequiresTmuxSession pins the refusal: outside tmux there is no
// session to open a window in, so --spawn must error rather than silently
// falling back to an in-place launch (which would replace the user's session —
// the exact outcome --spawn exists to avoid).
func TestSpawnLaunchRequiresTmuxSession(t *testing.T) {
	cap := withSpawnStubs(t, false, "%1", nil)

	var out bytes.Buffer
	err := spawnLaunch(&out, "cc", []string{"-w", "feat"})
	if err == nil {
		t.Fatal("expected an error outside tmux; got nil")
	}
	if !strings.Contains(err.Error(), "tmux") {
		t.Errorf("error should name the tmux requirement; got %v", err)
	}
	if cap.calls != 0 {
		t.Errorf("tmux must not be invoked outside a session; calls = %d", cap.calls)
	}
	if out.Len() != 0 {
		t.Errorf("no success output expected on refusal; got %q", out.String())
	}
}

// TestSpawnLaunchPropagatesTmuxFailure verifies a failed window spawn surfaces
// as an error (non-zero exit) rather than a silent no-op. A silent no-op would
// leave the user believing a teammate is running when none is.
func TestSpawnLaunchPropagatesTmuxFailure(t *testing.T) {
	requireTmuxBinary(t)

	withSpawnStubs(t, true, "", os.ErrPermission)

	var out bytes.Buffer
	err := spawnLaunch(&out, "cc", []string{"-w", "feat"})
	if err == nil {
		t.Fatal("expected the tmux failure to propagate; got nil")
	}
	if !strings.Contains(err.Error(), "spawn tmux window") {
		t.Errorf("error should identify the spawn step; got %v", err)
	}
}

func TestShellQuote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: "''"},
		{name: "plain", in: "feat-auth", want: "feat-auth"},
		{name: "path", in: "/tmp/wt/feat", want: "/tmp/wt/feat"},
		{name: "space", in: "a b", want: "'a b'"},
		{name: "semicolon", in: "a;b", want: "'a;b'"},
		{name: "dollar", in: "$HOME", want: "'$HOME'"},
		{name: "backtick", in: "a`b`", want: "'a`b`'"},
		{name: "single quote", in: "it's", want: `'it'\''s'`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := shellQuote(tt.in); got != tt.want {
				t.Errorf("shellQuote(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParsePaneID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		out     string
		want    string
		wantErr bool
	}{
		{name: "pane id", out: "%7", want: "%7"},
		{name: "trailing newline", out: "%12\n", want: "%12"},
		{name: "surrounding whitespace", out: "  %3  \n", want: "%3"},
		{name: "empty output", out: "", wantErr: true},
		{name: "window index instead of pane id", out: "2", wantErr: true},
		{name: "error text on stdout", out: "no server running", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parsePaneID(tt.out)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parsePaneID(%q) = %q, want error", tt.out, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parsePaneID(%q) error = %v", tt.out, err)
			}
			if got != tt.want {
				t.Errorf("parsePaneID(%q) = %q, want %q", tt.out, got, tt.want)
			}
		})
	}
}

// TestSpawnWorkingDirFallsBackToCwd covers the no-project-root path: the spawned
// window still needs a directory, so spawnWorkingDir degrades to the process cwd
// rather than failing the launch.
func TestSpawnWorkingDirFallsBackToCwd(t *testing.T) {
	origRoot := findProjectRootFn
	findProjectRootFn = func() (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { findProjectRootFn = origRoot })

	got, err := spawnWorkingDir()
	if err != nil {
		t.Fatalf("spawnWorkingDir() error = %v, want nil", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd(): %v", err)
	}
	if got != cwd {
		t.Errorf("spawnWorkingDir() = %q, want process cwd %q", got, cwd)
	}
}
