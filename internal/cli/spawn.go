package cli

// --spawn: launch a MoAI session in a NEW tmux window instead of replacing the
// current process.
//
// The launch commands (`moai cc` / `moai glm` / `moai cg`) normally replace the
// running shell via syscall.Exec. That is the right default for "I want to work
// here now", but it cannot express "keep my session and start a teammate
// alongside it" — the case a parallel worktree session needs.
//
// `--spawn` supplies exactly that: the same command is re-issued verbatim in a
// new tmux window (`-d`, so focus stays put) and the current process exits 0.
// Combined with `-w <name>` this opens a teammate session in an isolated
// worktree without disturbing the caller.
//
//	moai cg -w feat-auth --spawn    # GLM teammate in .claude/worktrees/feat-auth
//	moai cc -w feat-auth --spawn    # Claude teammate, same worktree
//
// The flag is MoAI-side only: it is stripped before any argument reaches
// claude, and it is never consumed after the `--` pass-through marker.

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/execerr"
	"github.com/modu-ai/moai-adk/internal/tmux"
)

// spawnFlag is the MoAI-side token requesting a new-tmux-window launch.
const spawnFlag = "--spawn"

// inTmuxFn reports whether the current process runs inside a tmux session.
// Injection seam: tests override it instead of manipulating $TMUX, which would
// race with parallel tests sharing the process environment.
var inTmuxFn = func() bool { return tmux.NewDetector().InTmuxSession() }

// tmuxSpawnFn opens a new tmux window. Injection seam: CI containers have no
// tmux server, so tests capture (cwd, command) here rather than shelling out.
//
// @MX:ANCHOR: [AUTO] tmuxSpawnFn — sole tmux new-window call site for --spawn.
// Fan-in: spawnLaunch (runCC / runCG / runGLM).
// @MX:REASON: the real tmux invocation cannot run in CI (no server) and its
// argv correctness is the whole contract of --spawn; injecting the function
// pointer is the only way to assert cwd and command deterministically.
var tmuxSpawnFn = defaultTmuxSpawn

// stripSpawnFlag removes every --spawn token appearing BEFORE the `--`
// pass-through marker and reports whether one was found. Tokens at or after
// `--` belong to claude and are returned untouched.
func stripSpawnFlag(args []string) ([]string, bool) {
	filtered := make([]string, 0, len(args))
	found := false

	for i, arg := range args {
		if arg == "--" {
			filtered = append(filtered, args[i:]...)
			break
		}
		if arg == spawnFlag {
			found = true
			continue
		}
		filtered = append(filtered, arg)
	}

	return filtered, found
}

// defaultTmuxSpawn runs `tmux new-window` in the caller's current session and
// returns the new pane id (e.g. "%7").
//
//		tmux new-window -d -P -F '#{pane_id}' -c <cwd> '<command>'
//
//	  - -d creates the window without moving focus, so the caller keeps working.
//	  - -P -F '#{pane_id}' makes stdout the pane id alone, which the caller
//	    prints so the user can jump to the window.
//	  - -c <cwd> sets the new window's working directory.
func defaultTmuxSpawn(cwd, command string) (string, error) {
	out, err := exec.Command("tmux", "new-window", "-d", "-P", "-F", "#{pane_id}", "-c", cwd, command).Output()
	if err != nil {
		// execerr.StatusDetail, not %w: a raw *exec.ExitError chain would be
		// mistaken for an intentional ExitCoder at the cmd/moai seam (t130).
		return "", fmt.Errorf("tmux new-window: %s", execerr.StatusDetail(err))
	}
	return parsePaneID(string(out))
}

// parsePaneID validates the stdout of `tmux new-window -P -F '#{pane_id}'`.
// A pane id always starts with '%'; anything else means tmux printed something
// other than the requested format, and passing that through would hand the user
// a switch target that does not exist.
func parsePaneID(out string) (string, error) {
	paneID := strings.TrimSpace(out)
	if !strings.HasPrefix(paneID, "%") {
		return "", fmt.Errorf("tmux returned unexpected pane id: %q", paneID)
	}
	return paneID, nil
}

// spawnLaunch re-issues `moai <subcommand> <args...>` in a new tmux window and
// reports the pane id on out. args MUST already have --spawn stripped, so the
// spawned command launches normally rather than spawning again.
//
// All three failure modes are setup errors the user must fix, so each returns a
// non-nil error rather than degrading silently:
//
//   - not inside tmux — there is no session to open a window in;
//   - tmux binary missing — nothing to open the window with;
//   - moai binary missing from PATH — the window would open and immediately die.
//
// Nothing has been mutated at this point (the caller invokes spawnLaunch before
// any settings write), so an error here leaves the environment untouched.
func spawnLaunch(out io.Writer, subcommand string, args []string) error {
	if !inTmuxFn() {
		// Phrased so the flag name is not the first word: the CLI error card
		// capitalizes the leading character, which would render "--Spawn".
		return fmt.Errorf(
			"tmux session required for %s — no $TMUX detected\n"+
				"  start tmux first, or drop %s to launch in the current terminal",
			spawnFlag, spawnFlag)
	}
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("%s requires the tmux binary: %w", spawnFlag, err)
	}
	if _, err := exec.LookPath("moai"); err != nil {
		return fmt.Errorf("%s needs the moai binary in PATH: %w", spawnFlag, err)
	}

	cwd, err := spawnWorkingDir()
	if err != nil {
		return err
	}

	command := buildSpawnCommand(subcommand, args)
	paneID, err := tmuxSpawnFn(cwd, command)
	if err != nil {
		return fmt.Errorf("spawn tmux window: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Spawned pane %s running `%s` in %s\n", paneID, command, cwd)
	_, _ = fmt.Fprintln(out, "Switch to it with: tmux select-window -t "+paneID)
	return nil
}

// spawnWorkingDir returns the directory the spawned window starts in. The
// project root is preferred so a `-w <name>` value resolves against
// .claude/worktrees/ regardless of which subdirectory the caller ran from;
// the process cwd is the fallback when no project root is detectable.
func spawnWorkingDir() (string, error) {
	if root, err := findProjectRootFn(); err == nil && root != "" {
		return root, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory for %s: %w", spawnFlag, err)
	}
	return cwd, nil
}

// buildSpawnCommand renders the shell command string tmux runs in the new
// window. Every argument is quoted, so a worktree name containing spaces or
// shell metacharacters reaches the spawned moai process intact.
func buildSpawnCommand(subcommand string, args []string) string {
	parts := make([]string, 0, len(args)+2)
	parts = append(parts, "moai", subcommand)
	for _, arg := range args {
		parts = append(parts, shellQuote(arg))
	}
	return strings.Join(parts, " ")
}

// shellQuote wraps s in single quotes, which suppresses every form of shell
// interpretation. An embedded single quote is handled the standard POSIX way:
// close the quote, emit a backslash-escaped quote, reopen. Arguments made only
// of safe characters are returned bare to keep the printed command readable.
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n'\"\\$`&;|<>()*?[]{}!#~") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
