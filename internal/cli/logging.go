package cli

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// defaultLogLevel is the minimum record level admitted on the non-hook path when
// MOAI_LOG_LEVEL is unset or unrecognized. warn-and-above reaches the operator;
// info and debug stay silent unless the variable lowers the bar.
const defaultLogLevel = slog.LevelWarn

// loggingDecision records where log records go and the minimum level admitted
// for one CLI invocation.
//
// The decision is produced as a value, separately from installing it, because a
// slog.Handler does not expose its writer: routing the choice straight into the
// default-logger installation would leave a test able to prove only that *a*
// handler was installed, never *which* destination was picked.
type loggingDecision struct {
	dest  io.Writer
	level slog.Level
}

// resolveLogLevel maps MOAI_LOG_LEVEL onto a slog.Level, falling back to
// defaultLogLevel when the variable is unset or holds a value slog cannot parse.
// A typo must not fail the invocation, so a parse error degrades to the default
// rather than propagating.
//
// The variable name is read from config.EnvLogLevel and is deliberately never
// restated as a literal here.
func resolveLogLevel() slog.Level {
	raw := strings.TrimSpace(os.Getenv(config.EnvLogLevel))
	if raw == "" {
		return defaultLogLevel
	}
	// slog.Level.UnmarshalText accepts DEBUG/INFO/WARN/ERROR case-insensitively,
	// and offset forms such as "WARN+2".
	var level slog.Level
	if err := level.UnmarshalText([]byte(raw)); err != nil {
		return defaultLogLevel
	}
	return level
}

// resolveLoggingDecision chooses the log destination for one CLI invocation.
//
// The `moai hook` path discards every record: stdout carries the hook's
// structured JSON contract and stderr is read by the Claude Code runtime, so a
// stray record corrupts the exchange. That carve-out is unconditional —
// MOAI_LOG_LEVEL does not re-open it.
//
// Every other subcommand writes to stderr, never stdout, which stays reserved
// for machine-readable output (--format=json / sarif payloads).
func resolveLoggingDecision(args []string) loggingDecision {
	if isHookCommand(args) {
		return loggingDecision{dest: io.Discard, level: defaultLogLevel}
	}
	return loggingDecision{dest: os.Stderr, level: resolveLogLevel()}
}

// configureLogging installs the process-wide default slog handler.
//
// This is the CLI's single default-logger installation site. It runs once from
// Execute(), ahead of the trivial/full initialization branch, so both paths
// share one logging decision.
func configureLogging(args []string) {
	d := resolveLoggingDecision(args)
	slog.SetDefault(slog.New(slog.NewTextHandler(d.dest, &slog.HandlerOptions{Level: d.level})))
}

// isHookCommand reports whether the CLI args name the `hook` subcommand. It
// mirrors the isTrivialCommand walk in root.go: leading flags are skipped, and
// the first non-flag argument is the subcommand.
//
// The name comes from hookCmd rather than a literal, so renaming the subcommand
// cannot silently detach the discard carve-out from it.
func isHookCommand(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		return arg == hookCmd.Name()
	}
	return false
}
