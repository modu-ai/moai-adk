package defs

import "time"

// Hook timeout defaults (in seconds, matching settings.json generation).
const (
	// HookDefaultTimeout is the default timeout for most hook events (seconds).
	HookDefaultTimeout = 5

	// HookPostToolTimeout is the timeout for PostToolUse hooks (seconds).
	HookPostToolTimeout = 60
)

// Git command timeouts.
const (
	// GitShortTimeout is used for quick git operations (status, config, etc.).
	GitShortTimeout = 5 * time.Second

	// GitLongTimeout is used for longer git operations (fetch, push, etc.).
	GitLongTimeout = 10 * time.Second
)
