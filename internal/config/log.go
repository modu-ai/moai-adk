package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// logTierReadFailure appends a structured error entry to .moai/logs/config.log.
// This is a best-effort logger — if the log file cannot be opened or written,
// the error falls back to slog.Warn on stderr and never panics.
//
// Entry format (RFC3339 | source=<tier> | path=<file> | err=<message>):
//
//	2026-05-10T12:00:00+09:00 | source=project | path=.moai/config/sections/quality.yaml | err=...
//
// @MX:NOTE: [AUTO] SPEC-V3R2-RT-005 M4 GREEN — REQ-040: dedicated config.log for tier read failures
//
// REQ-V3R2-RT-005-040
func logTierReadFailure(source Source, path string, err error) {
	const logDir = ".moai/logs"
	const logFile = ".moai/logs/config.log"

	// Best-effort directory creation.
	if mkErr := os.MkdirAll(logDir, 0o755); mkErr != nil {
		slog.Warn("config: cannot create log directory",
			"dir", logDir,
			"err", mkErr,
			"source", source.String(),
			"path", path,
			"original_err", err,
		)
		return
	}

	f, openErr := os.OpenFile(filepath.Clean(logFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if openErr != nil {
		slog.Warn("config: cannot open config.log",
			"file", logFile,
			"err", openErr,
			"source", source.String(),
			"path", path,
			"original_err", err,
		)
		return
	}
	defer func() { _ = f.Close() }()

	entry := fmt.Sprintf("%s | source=%s | path=%s | err=%v\n",
		time.Now().Format(time.RFC3339),
		source.String(),
		path,
		err,
	)

	if _, writeErr := f.WriteString(entry); writeErr != nil {
		slog.Warn("config: cannot write to config.log",
			"file", logFile,
			"err", writeErr,
			"source", source.String(),
			"path", path,
			"original_err", err,
		)
	}
}
