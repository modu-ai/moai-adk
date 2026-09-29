//go:build linux

// Linux candidate readers (design.md §E): /proc for argv, environment, and
// working directory of same-uid processes. Every reader declines (false)
// rather than guessing: a decline downstream is an honest NO_ACTIVE_FACTORY,
// never a wrong join.
package discovery

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

func procFields(path string, sep byte) ([]string, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	fields := strings.Split(strings.TrimRight(string(raw), string(sep)), string(sep))
	if len(fields) == 0 {
		return nil, false
	}
	return fields, true
}

// platformArgv reads the candidate's command line.
func platformArgv(ctx context.Context, pid int) ([]string, bool) {
	return procFields("/proc/"+strconv.Itoa(pid)+"/cmdline", 0)
}

// platformEnv reads the candidate's environment.
func platformEnv(ctx context.Context, pid int) (map[string]string, bool) {
	fields, ok := procFields("/proc/"+strconv.Itoa(pid)+"/environ", 0)
	if !ok {
		return nil, false
	}
	env := map[string]string{}
	for _, token := range fields {
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
	path, err := os.Readlink("/proc/" + strconv.Itoa(pid) + "/cwd")
	if err != nil || path == "" {
		return "", false
	}
	return path, true
}
