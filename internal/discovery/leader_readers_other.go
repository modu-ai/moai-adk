//go:build !darwin && !linux

// Platforms without candidate readers (design.md §E): every reader declines,
// so discovery verifies no leader and the join degrades to today's honest
// NO_ACTIVE_FACTORY refusal — never to a wrong join. A later SPEC that wants
// discovery here widens the reader behind the same seam.
package discovery

import "context"

func platformArgv(ctx context.Context, pid int) ([]string, bool) { return nil, false }

func platformEnv(ctx context.Context, pid int) (map[string]string, bool) { return nil, false }

func platformCwd(ctx context.Context, pid int) (string, bool) { return "", false }
