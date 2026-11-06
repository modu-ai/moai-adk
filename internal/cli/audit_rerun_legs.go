package cli

// audit_rerun_legs.go — the re-run of named audit_multi legs
// (SPEC-AUDIT-ENGINE-CONVERGENCE-001 REQ-AEC-008).
//
// A call naming legs through `rerun_legs` runs only those legs. Every other leg's
// verdict is carried from the result the SERVER persisted for the same session —
// never from an argument of the call — and only when that persisted result records
// the same head commit, the same tree digest and the same scope (target, resolved
// base, delta base). Anything else, a missing result included, runs every leg, and
// the result says so.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// rerunLegsArg is the audit_multi argument naming the legs to run again.
const rerunLegsArg = "rerun_legs"

// rerunLegsDescription is the tool-schema text of the rerun_legs input.
const rerunLegsDescription = "Optional legs to run again, as backend names (claude, codex, glm). Requires session_id. Only the named legs run; every other leg's verdict is carried from the result this server persisted for the same session, and a verdict the call supplies for a carried leg is ignored. The carry applies only when that persisted result records the same head commit, tree digest and scope (target, resolved base, delta base); otherwise every leg runs and the result says why. Omitted or empty ⇒ every leg runs."

// readRerunLegs reads the optional rerun_legs argument: a JSON array of backend
// names. Absent or empty ⇒ nil. A value that is not an array of known backend
// names is an error, not an absent request: reading it as absent would quietly
// re-run every leg.
func readRerunLegs(req mcp.CallToolRequest) ([]string, error) {
	raw, present := req.GetArguments()[rerunLegsArg]
	if !present || raw == nil {
		return nil, nil
	}
	var items []any
	switch v := raw.(type) {
	case []any:
		items = v
	case []string:
		for _, s := range v {
			items = append(items, s)
		}
	default:
		return nil, fmt.Errorf("%s must be an array of backend names (%s, %s, %s), got %T", rerunLegsArg, BackendClaude, BackendCodex, BackendGLM, raw)
	}
	seen := map[string]bool{}
	for _, it := range items {
		name, ok := it.(string)
		if !ok {
			return nil, fmt.Errorf("%s must hold backend names (%s, %s, %s), got %T", rerunLegsArg, BackendClaude, BackendCodex, BackendGLM, it)
		}
		name = strings.ToLower(strings.TrimSpace(name))
		switch name {
		case BackendClaude, BackendCodex, BackendGLM:
			seen[name] = true
		default:
			return nil, fmt.Errorf("%s names an unknown backend %q (want %s, %s or %s)", rerunLegsArg, name, BackendClaude, BackendCodex, BackendGLM)
		}
	}
	var out []string
	for _, b := range []string{BackendClaude, BackendCodex, BackendGLM} {
		if seen[b] {
			out = append(out, b)
		}
	}
	return out, nil
}

// convergencePersistPath returns the directory and the file the convergence result
// of sessionID is persisted under for projectRoot — the one place both the writer
// and the re-run's reader take the location from.
func convergencePersistPath(sessionID, projectRoot string) (dir, path string, err error) {
	stateDir, name := convergenceStateDirFor(projectRoot), sessionID+".json"
	if root := strings.TrimSpace(projectRoot); root != "" {
		store, serr := auditreceipt.StoreRoot(root)
		if serr != nil {
			return "", "", serr
		}
		stateDir = filepath.Join(store, ".moai", "state")
		if store != root {
			name = sessionID + "--" + auditreceipt.TreeKey(root) + ".json"
		}
	}
	dir = filepath.Join(stateDir, "audit-multi")
	return dir, filepath.Join(dir, name), nil
}

// auditedScopeOf is the scope a call measures: the target, the merge base a
// baseBranch target resolves to, and the delta base. nil when the scope cannot be
// established (no tree, or a baseBranch target whose base does not resolve): a
// result persisted without a scope is never carried from.
func auditedScopeOf(root, target, deltaBase string) *AuditedScope {
	if strings.TrimSpace(root) == "" {
		return nil
	}
	scope := AuditedScope{Target: strings.TrimSpace(target), DeltaBase: deltaBase}
	if scope.Target == codexTargetBaseBranch {
		base, err := resolveReviewBase(root)
		if err != nil {
			return nil
		}
		scope.Base = base.MergeBase
	}
	return &scope
}

// carryPersistedLegs decides which legs of the call are carried. active lists the
// backends whose gate is not off. It returns the persisted result the legs come
// from, the carried verdicts keyed by backend, and — when named legs were asked for
// but nothing could be carried — the note saying every leg runs and why. An empty
// carried map with an empty note means every active leg was named.
func carryPersistedLegs(cfg MultiAuditConfig, head, digest string, captureFailed bool, scope *AuditedScope, active []string) (ConvergenceResult, map[string]PerBackendVerdict, string) {
	refuse := func(why string) (ConvergenceResult, map[string]PerBackendVerdict, string) {
		return ConvergenceResult{}, nil, rerunLegsArg + " not applied (" + why + "): every leg ran"
	}
	if strings.TrimSpace(cfg.SessionID) == "" {
		return refuse("no session_id, so no persisted result to carry from")
	}
	if captureFailed || head == "" || digest == "" || scope == nil {
		return refuse("the head, tree digest or scope of this call could not be established")
	}
	prior, ok := loadPersistedConvergence(cfg.SessionID, cfg.ProjectRoot)
	if !ok {
		return refuse("no persisted result for session " + cfg.SessionID)
	}
	if prior.AuditedHead == "" || prior.AuditedDigest == "" || prior.AuditedScope == nil {
		return refuse("the persisted result records no head, tree digest or scope")
	}
	switch {
	case prior.AuditedHead != head:
		return refuse("the head commit differs from the persisted result's")
	case prior.AuditedDigest != digest:
		return refuse("the tree digest differs from the persisted result's")
	case *prior.AuditedScope != *scope:
		return refuse("the scope (target, resolved base or delta base) differs from the persisted result's")
	}
	named := map[string]bool{}
	for _, n := range cfg.RerunLegs {
		named[n] = true
	}
	isActive := map[string]bool{}
	for _, a := range active {
		isActive[a] = true
	}
	carried := map[string]PerBackendVerdict{}
	for _, v := range prior.PerBackendVerdicts {
		if isActive[v.Backend] && !named[v.Backend] {
			carried[v.Backend] = v
		}
	}
	return prior, carried, ""
}

// loadPersistedConvergence reads the result the server persisted for the session
// in the location persistConvergenceResult writes to. Any failure reads as "no
// persisted result" — the caller then runs every leg.
func loadPersistedConvergence(sessionID, projectRoot string) (ConvergenceResult, bool) {
	_, path, err := convergencePersistPath(sessionID, projectRoot)
	if err != nil {
		return ConvergenceResult{}, false
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ConvergenceResult{}, false
	}
	var r ConvergenceResult
	if err := json.Unmarshal(b, &r); err != nil {
		return ConvergenceResult{}, false
	}
	return r, true
}

// carriedLegNames lists the carried backends in the canonical order.
func carriedLegNames(carried map[string]PerBackendVerdict) []string {
	var out []string
	for _, b := range []string{BackendClaude, BackendCodex, BackendGLM} {
		if _, ok := carried[b]; ok {
			out = append(out, b)
		}
	}
	return out
}

// rerunLegsRan lists the backends of the assembled verdicts that were run, not
// carried.
func rerunLegsRan(verdicts []PerBackendVerdict, carried map[string]PerBackendVerdict) []string {
	var out []string
	for _, v := range verdicts {
		if _, ok := carried[v.Backend]; !ok {
			out = append(out, v.Backend)
		}
	}
	return out
}
