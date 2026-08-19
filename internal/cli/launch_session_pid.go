package cli

// launch_session_pid.go — stamps the session PID into the launch environment.
//
// The coordination registry is written from a hook subprocess that exits within
// milliseconds, so the resolver in internal/session normally has to walk the
// process ancestry to find the long-lived session process. The launcher does
// not have to guess: on POSIX it hands its own process to claude via
// execve(2), so its PID *becomes* the session's. Exporting that PID removes the
// ancestry walk from the resolution path entirely (it is step 1 of
// resolveSessionPID, ahead of the walk).
//
// The stamp belongs to callers that KNOW the session PID. A hook must never set
// it: a hook subprocess's own PID is dead on arrival, and recording it
// reintroduces the very defect the resolver exists to avoid.

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// withSessionPID returns env with MOAI_SESSION_PID set to pid.
//
// Any inherited value is stripped first: a nested launch (`--spawn`, a
// re-exec) carries the outer launcher's PID in its environment, and leaving
// both entries in place would make the child's reading of the variable depend
// on duplicate-key resolution order. A non-positive pid is a no-op — an
// unusable stamp is worse than none, since the resolver's ancestry fallback
// still works.
func withSessionPID(env []string, pid int) []string {
	if pid <= 0 {
		return env
	}
	prefix := config.EnvMoaiSessionPID + "="
	out := make([]string, 0, len(env)+1)
	for _, kv := range env {
		if strings.HasPrefix(kv, prefix) {
			continue
		}
		out = append(out, kv)
	}
	return append(out, fmt.Sprintf("%s%d", prefix, pid))
}
