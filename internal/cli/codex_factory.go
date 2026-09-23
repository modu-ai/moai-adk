package cli

// codex_factory.go — `moai codex -f` / `moai codex -f worker`: the factory
// entry for the Codex door (card t865, operator goal 2026-09-16 decision 4:
// the first factory combination is the fixed 1:1 — codex lead + cc worker).
//
// The codex launcher uses DisableFlagParsing with a closed-set verb table,
// so the factory token is intercepted BEFORE the verb lookup exactly as
// --spawn and -w are: stripped from the head, resolved into the factory
// mode env, and never forwarded to the codex child (a forwarded -f would be
// an unknown token in the child's argv — the same failure shape a forwarded
// synthesized verb has).
//
// Shapes accepted (mirroring the cc surface post-N-removal):
//
//	-f                → the factory lead (env only; the cli verb launches)
//	-f worker         → join the running factory as the next free worker-<n>
//	-f worker-<n>     → join as exactly worker n
//
// Legacy spellings still parse (keep-alias): `-f agent` behaves as `-f
// worker`, and `-f lane-<n>` / `-f agent-<n>` as `-f worker-<n>`; the claim
// canonicalizes them and prints a deprecation hint.
//
// A numeric count is rejected with the same error text the cc surface emits.

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

const codexFactoryBackend = "codex"

// stripCodexFactoryFlag removes the -f/--factory tokens from the head and
// reports the requested factory role. role is the role token as typed
// (factoryWorkerRoleToken, the legacy factoryLegacyAgentRoleToken, or "");
// lane is a worker label as typed. It scans only the head (pre---) because
// tokens after -- belong to the codex child.
func stripCodexFactoryFlag(head []string) (rest []string, lead bool, role string, lane string, err error) {
	rest = make([]string, 0, len(head))
	for i := 0; i < len(head); i++ {
		token := head[i]
		value := ""
		hasValue := false
		switch {
		case token == "-f" || token == "--factory":
			// Consume the next token ONLY when it names a factory role or is
			// definitely not a verb: the codex verb position is the same token
			// slot ("cli", "app"), so a space-separated verb must never be
			// eaten as a -f value — but a typo'd value like "4" errors HERE
			// with the factory usage text rather than as an unknown verb.
			if i+1 < len(head) && !strings.HasPrefix(head[i+1], "-") {
				next := head[i+1]
				if isFactoryRoleToken(next) || isFactoryLaneShape(next) {
					value, hasValue = next, true
					i++
				} else if !codexHeadTokenIsVerb(next) {
					return nil, false, "", "", fmt.Errorf("%s, got %q", factoryFlagUsageError, next)
				}
			}
		case strings.HasPrefix(token, "--factory="):
			value = strings.TrimPrefix(token, "--factory=")
			hasValue = true
		case strings.HasPrefix(token, "-f="):
			value = strings.TrimPrefix(token, "-f=")
			hasValue = true
		default:
			rest = append(rest, token)
			continue
		}
		if !hasValue {
			lead = true
			continue
		}
		switch {
		case isFactoryRoleToken(value):
			role = value
		case isFactoryLaneShape(value):
			lane = value
		default:
			return nil, false, "", "", fmt.Errorf("%s, got %q", factoryFlagUsageError, value)
		}
	}
	return rest, lead, role, lane, nil
}

// isFactoryLaneShape reports the worker label shapes only: the canonical
// worker-<n> and the legacy lane-<n> / agent-<n> spellings.
func isFactoryLaneShape(v string) bool {
	if _, ok := kanban.SplitFactoryLaneLabel(v); ok {
		return true
	}
	_, ok := kanban.SplitFactoryAgentLabel(v)
	return ok
}

// isFactoryRoleToken reports the role token: `worker`, or its legacy
// spelling `agent`.
func isFactoryRoleToken(v string) bool {
	return v == factoryWorkerRoleToken || v == factoryLegacyAgentRoleToken
}

// codexHeadTokenIsVerb reports whether the token could legally occupy the
// verb position (one of the closed routing set's keys).
func codexHeadTokenIsVerb(v string) bool {
	_, ok := codexVerbRouting[v]
	return ok
}

// applyCodexFactoryEntry resolves the stripped factory tokens into the
// session's factory mode env and returns the restore func. The label join
// path claims its slot through the same registry the cc launcher uses, so
// bump and liveness rules are one implementation across doors.
func applyCodexFactoryEntry(cmd *cobra.Command, role string, lane string) (func(), error) {
	noop := func() {}
	restoreFacts := exportFactoryLaunchFacts("", codexFactoryBackend)
	if role != "" {
		next := kanban.NextFactoryWorkerNumber(loadFactoryRegistry(factoryRegistryPath(launchProjectRoot())), factoryProcessAlive)
		lane = kanban.FactoryLaneLabel(next)
		if role == factoryLegacyAgentRoleToken {
			// Routed through the legacy label so the claim prints the
			// deprecation hint, exactly as the cc/glm doors do.
			lane = kanban.FactoryAgentLabel(next)
		}
	}
	if lane != "" {
		final, claimErr := resolveFactoryWorkerName(launchProjectRoot(), lane, cmd.ErrOrStderr())
		if claimErr != nil {
			restoreFacts()
			return noop, claimErr
		}
		restoreMode := enterFactoryWorkerMode(final, 0)
		return func() {
			restoreMode()
			restoreFacts()
		}, nil
	}
	restoreMode := enterFactoryLeadMode(config.DefaultFactoryLeadWorkers, "")
	return func() {
		restoreMode()
		restoreFacts()
	}, nil
}
