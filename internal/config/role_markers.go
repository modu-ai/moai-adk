// role_markers.go is the single role-marker registry the SessionStart hook
// and the launcher both read (SPEC-ALWAYS-LOADED-BUDGET-001 REQ-ALB-011).
//
// A role session is a session a launcher marked as a factory leader or a
// factory lane. The mark is an environment variable, so the registry is a
// list of {name, env key} pairs over the SAME constants the launchers stamp
// (internal/cli's factory lane entry writes config.EnvMoaiFactoryWorker,
// config.EnvFactoryRole, and config.EnvMoaiFactoryWorkers — envkeys.go).
// Consumers derive their marker set from RoleMarkerRegistry(), never from a
// hand-written list: when a marker is added here, every guard and every
// detection path picks it up without further edits.
//
// The same file hosts the neutral region markers and their extractor. The
// extraction is pure text logic with no hook or template dependency, so both
// the hook (which builds the SessionStart role core from deployed rule
// files) and the template test suite (whose binding-ledger check verifies
// role-core rows against builder output) consume this one implementation —
// a second extractor would let the two disagree silently.
package config

import "strings"

// RoleCoreMarkerStart and RoleCoreMarkerEnd bound a role-core region inside
// a deployed role-gated rule file. HTML comments, same family as the
// moai:evolvable markers: invisible in rendered markdown, carrying no SPEC
// id, card id, date, or hash (template neutrality).
const (
	RoleCoreMarkerStart = "<!-- moai:role-core-start -->"
	RoleCoreMarkerEnd   = "<!-- moai:role-core-end -->"
)

// RoleMarker is one registered session-role marker: a registry name and the
// environment variable whose presence marks a session with that role.
type RoleMarker struct {
	// Name is the registry identifier (the spelling the binding ledger's
	// REQ-ALB-007 entry points use).
	Name string
	// EnvKey is the environment variable the launcher stamps and the hook
	// detects. Presence with a non-empty value marks the role.
	EnvKey string
}

// RoleMarkerRegistry returns the registered leader/lane markers. Seed: the
// Q8 measurement set — factory lane (MOAI_FACTORY_WORKER carrying the
// lane-<n> label) and factory leader (MOAI_FACTORY_WORKERS). The LANE entry
// comes first deliberately: a lane launch environment carries BOTH keys
// (the worker label and the run's lane count — internal/cli/factory.go
// stamps EnvMoaiFactoryWorker and EnvMoaiFactoryWorkers together), so the
// worker key is the lane discriminator and detection must consult it
// before the leader key. MOAI_FACTORY_ROLE=lane corroborates the lane role
// but is not itself a registry marker.
func RoleMarkerRegistry() []RoleMarker {
	return []RoleMarker{
		{Name: "factory-lane", EnvKey: EnvMoaiFactoryWorker},
		{Name: "factory-leader", EnvKey: EnvMoaiFactoryWorkers},
	}
}

// DetectRoleMarker returns the first registry role this lookup carries. A
// registry entry matches when its env key holds a non-empty value. lookup
// is injected (os.Getenv at the call site) so tests can drive the
// environment without touching process state.
func DetectRoleMarker(lookup func(string) string) (RoleMarker, bool) {
	if lookup == nil {
		return RoleMarker{}, false
	}
	for _, m := range RoleMarkerRegistry() {
		if lookup(m.EnvKey) != "" {
			return m, true
		}
	}
	return RoleMarker{}, false
}

// ExtractRoleCoreRegions splits content on the role-core region markers and
// returns the enclosed regions in document order.
//
// marked reports whether any start marker exists — the caller distinguishes
// "carries markers, core possibly empty" (a rule whose core is legitimately
// empty, e.g. a rule that binds every session) from "carries no markers"
// (a failure: the hook must not treat an unmarked rule file as an empty
// core). An unclosed start marker yields no region for it; production rule
// files are marker-complete, and the hook treats a truncated file through
// the same fail-visible path as any other unreadable core.
func ExtractRoleCoreRegions(content string) (regions []string, marked bool) {
	if !strings.Contains(content, RoleCoreMarkerStart) {
		return nil, false
	}
	for {
		i := strings.Index(content, RoleCoreMarkerStart)
		if i < 0 {
			break
		}
		rest := content[i+len(RoleCoreMarkerStart):]
		j := strings.Index(rest, RoleCoreMarkerEnd)
		if j < 0 {
			break // unclosed final region: ignored, named by marked=true
		}
		regions = append(regions, rest[:j])
		content = rest[j+len(RoleCoreMarkerEnd):]
	}
	return regions, true
}

// RoleCoreMarkerRequired is the neutral entry-point marker a deployed agent
// definition, skill, or workflow file carries when it cites a role-gated
// rule (REQ-ALB-024). The marker pairs with a binding directive to read
// both role-gated rule files in full before the file's first action.
const RoleCoreMarkerRequired = "moai:role-rules-required"
