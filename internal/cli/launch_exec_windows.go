//go:build windows

package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// execOrSpawnClaude spawns the claude binary as a child process, waits for it,
// and propagates its exit code. Windows lacks the execve(2) primitive that
// syscall.Exec relies on (it returns syscall.EWINDOWS at runtime), so instead of
// replacing the current process we spawn-and-exit — mirroring the reexecNewBinary
// pattern in update.go:461.
//
// REQ-CGH-001: this Windows path makes `moai cc` / `moai glm` launch correctly on
// Windows rather than failing with EWINDOWS at the unguarded syscall.Exec call.
func execOrSpawnClaude(claudeBin string, args, env []string) error {
	// args[0] is the program name (argv[0] convention); the child's actual
	// arguments are args[1:].
	var childArgs []string
	if len(args) > 1 {
		childArgs = args[1:]
	}

	child := exec.Command(claudeBin, childArgs...)
	child.Stdin = os.Stdin
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	child.Env = env

	if err := child.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			os.Exit(ee.ExitCode())
		}
		return fmt.Errorf("launch claude on windows: %w", err)
	}
	os.Exit(0)
	return nil // unreachable
}
