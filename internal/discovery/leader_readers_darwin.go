//go:build darwin

// Darwin candidate readers (design.md §E): `ps` for argv and the environment
// of same-uid processes (`ps eww` appends the environment to the command
// column), `lsof -d cwd` for the working directory — the t1038 probe chain's
// mechanism. Every reader declines (false) rather than guessing: a decline
// downstream is an honest NO_ACTIVE_FACTORY, never a wrong join.
package discovery

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// readerDeadline bounds one platform reader invocation (a `ps` or `lsof`
// subprocess).
const readerDeadline = 3 * time.Second

// readerContext bounds one platform reader invocation.
func readerContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, readerDeadline)
}

// platformArgv reads the candidate's command line.
func platformArgv(ctx context.Context, pid int) ([]string, bool) {
	ctx, cancel := readerContext(ctx)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ps", "-o", "command=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return nil, false
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) == 0 {
		return nil, false
	}
	return fields, true
}

// platformEnv reads the candidate's environment. `ps eww` renders the
// environment appended after the command in the COMMAND column, so the
// specific key is scanned out of the token stream: an argv never carries
// MOAI_KANBAN_ID, and a same-uid process is required for the environment to
// appear at all — a foreign or unreadable process simply declines here.
func platformEnv(ctx context.Context, pid int) (map[string]string, bool) {
	ctx, cancel := readerContext(ctx)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ps", "eww", "-p", strconv.Itoa(pid), "-o", "command=").Output()
	if err != nil {
		return nil, false
	}
	env := map[string]string{}
	for _, token := range strings.Fields(string(out)) {
		if key, value, found := strings.Cut(token, "="); found && key == config.EnvMoaiKanbanID {
			env[key] = value
		}
	}
	if env[config.EnvMoaiKanbanID] == "" {
		return nil, false
	}
	return env, true
}

// platformCwd reads the candidate's working directory.
func platformCwd(ctx context.Context, pid int) (string, bool) {
	ctx, cancel := readerContext(ctx)
	defer cancel()
	out, err := exec.CommandContext(ctx, "lsof", "-a", "-p", strconv.Itoa(pid), "-d", "cwd", "-Fn").Output()
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(out), "\n") {
		// lsof field output: p<pid>, then the pathname record n<path>.
		if path, found := strings.CutPrefix(line, "n"); found && path != "" {
			return path, true
		}
	}
	return "", false
}
