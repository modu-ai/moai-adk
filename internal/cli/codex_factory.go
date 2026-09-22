package cli

// codex_factory.go — `moai codex -f` / `moai codex -f agent`: the factory
// entry for the Codex door (card t865, operator goal 2026-09-16 decision 4:
// the first factory combination is the fixed 1:1 — codex lead + cc agent).
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
//	-f agent          → join the running factory as an agent-<n> lane
//	-f lane-<n>       → join as exactly lane n (compat spelling)
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
// reports the requested factory role. It scans only the head (pre---)
// because tokens after -- belong to the codex child.
func stripCodexFactoryFlag(head []string) (rest []string, lead bool, agent bool, lane string, err error) {
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
				if next == factoryAgentRoleToken || isFactoryLaneShape(next) {
					value, hasValue = next, true
					i++
				} else if !codexHeadTokenIsVerb(next) {
					return nil, false, false, "", fmt.Errorf("%s, got %q", factoryFlagUsageError, next)
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
		case value == factoryAgentRoleToken:
			agent = true
		case isFactoryLaneShape(value):
			lane = value
		default:
			return nil, false, false, "", fmt.Errorf("%s, got %q", factoryFlagUsageError, value)
		}
	}
	return rest, lead, agent, lane, nil
}

// isFactoryLaneShape reports the lane-<n>/agent-<n> shapes only.
func isFactoryLaneShape(v string) bool {
	if _, ok := kanban.SplitFactoryLaneLabel(v); ok {
		return true
	}
	_, ok := kanban.SplitFactoryAgentLabel(v)
	return ok
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
func applyCodexFactoryEntry(cmd *cobra.Command, agent bool, lane string) (func(), error) {
	noop := func() {}
	restoreFacts := exportFactoryLaunchFacts("", codexFactoryBackend)
	if agent {
		next := kanban.NextFactoryAgentNumber(loadFactoryRegistry(factoryRegistryPath(launchProjectRoot())), factoryProcessAlive)
		lane = kanban.FactoryAgentLabel(next)
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
