// permission.go — the role permission contract.
//
// Every emitted Codex role carries a contract over a fixed axis set. The
// restrictions a role's contract requires are derived from the role's Claude
// tool list and its contract sandbox; each required restriction must map to
// exactly one row of the manifest's axis table, and that row is either
// `enforced` (a Codex field this emitter writes) or `UNSUPPORTED` (the host
// cannot express it). A required restriction with no row fails emission, so
// no restriction is dropped silently, and an `UNSUPPORTED` row is never
// counted as passing.
package agentemit

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// Mapping values of the axis table.
const (
	MappingEnforced    = "enforced"
	MappingUnsupported = "UNSUPPORTED"
)

// PermissionAxes is the fixed axis set every contract covers.
var PermissionAxes = []string{"sandbox", "write-path-scope", "shell", "mcp-server", "mcp-tool", "subagent", "web"}

// enforceableFields names, per (axis, kind), the only Codex field this
// emitter writes to enforce that restriction. An `enforced` row naming any
// other field would claim enforcement the emitter never performs.
var enforceableFields = map[Requirement]string{
	{Axis: "sandbox", Kind: "mode"}:     "sandbox_mode",
	{Axis: "mcp-server", Kind: "grant"}: "mcp_servers",
}

// AxisMapping is one row of the axis table: how one restriction kind on one
// axis is carried on Codex.
type AxisMapping struct {
	Axis     string `yaml:"axis"`
	Kind     string `yaml:"kind"`
	Mapping  string `yaml:"mapping"`  // enforced | UNSUPPORTED
	Field    string `yaml:"field"`    // enforced only: the Codex field
	Basis    string `yaml:"basis"`    // measured | documented | unmeasured
	Reason   string `yaml:"reason"`   // UNSUPPORTED only: host-expressivity reason
	Evidence string `yaml:"evidence"` // what was observed or documented
}

// PermissionContract is the manifest's contract section.
type PermissionContract struct {
	Axes           []AxisMapping     `yaml:"axes"`
	DefaultSandbox string            `yaml:"default_sandbox"`
	RoleSandbox    map[string]string `yaml:"role_sandbox"`
}

// Requirement is one restriction a role's contract requires.
type Requirement struct {
	Axis string
	Kind string
}

// AxisVerdict is one (role, requirement) row of the permission report.
type AxisVerdict struct {
	Role    string
	Axis    string
	Kind    string
	Mapping string
	Field   string
	Basis   string
	Reason  string
}

// PermissionReport is the static role verification report. Pass holds only
// `enforced` verdicts; Unsupported holds every `UNSUPPORTED` verdict. A
// static report cannot establish runtime blocking — that rests on the live
// role verification alone.
type PermissionReport struct {
	Verdicts    []AxisVerdict
	Pass        []AxisVerdict
	Unsupported []AxisVerdict
}

// validateContract is the manifest-level self-validation of the contract.
func validateContract(pc *PermissionContract, accepted map[string]bool) error {
	knownAxis := map[string]bool{}
	for _, a := range PermissionAxes {
		knownAxis[a] = true
	}
	seen := map[Requirement]bool{}
	covered := map[string]bool{}
	for _, m := range pc.Axes {
		key := Requirement{Axis: m.Axis, Kind: m.Kind}
		if !knownAxis[m.Axis] {
			return fmt.Errorf("agentemit: permission contract axis %q is not one of %v", m.Axis, PermissionAxes)
		}
		if m.Kind == "" {
			return fmt.Errorf("agentemit: permission contract axis %s carries no restriction kind", m.Axis)
		}
		if seen[key] {
			return fmt.Errorf("agentemit: permission contract maps %s/%s more than once", m.Axis, m.Kind)
		}
		seen[key] = true
		covered[m.Axis] = true
		if strings.TrimSpace(m.Evidence) == "" {
			return fmt.Errorf("agentemit: permission contract %s/%s carries no evidence", m.Axis, m.Kind)
		}
		switch m.Mapping {
		case MappingEnforced:
			if m.Basis == "unmeasured" {
				return fmt.Errorf("agentemit: permission contract %s/%s is enforced on an unmeasured basis — an unmeasured axis is never enforced", m.Axis, m.Kind)
			}
			if m.Basis != "measured" && m.Basis != "documented" {
				return fmt.Errorf("agentemit: permission contract %s/%s enforced basis %q, want measured or documented", m.Axis, m.Kind, m.Basis)
			}
			if want, ok := enforceableFields[key]; !ok || m.Field != want {
				return fmt.Errorf("agentemit: permission contract %s/%s claims enforcement through field %q, which this emitter does not write for that restriction", m.Axis, m.Kind, m.Field)
			}
		case MappingUnsupported:
			if strings.TrimSpace(m.Reason) == "" {
				return fmt.Errorf("agentemit: permission contract %s/%s is UNSUPPORTED with no host-expressivity reason", m.Axis, m.Kind)
			}
			switch m.Basis {
			case "measured", "documented", "unmeasured":
			default:
				return fmt.Errorf("agentemit: permission contract %s/%s basis %q, want measured, documented, or unmeasured", m.Axis, m.Kind, m.Basis)
			}
		default:
			return fmt.Errorf("agentemit: permission contract %s/%s mapping %q is neither %s nor %s", m.Axis, m.Kind, m.Mapping, MappingEnforced, MappingUnsupported)
		}
	}
	for _, a := range PermissionAxes {
		if !covered[a] {
			return fmt.Errorf("agentemit: permission contract does not cover axis %s", a)
		}
	}
	if !accepted[pc.DefaultSandbox] {
		return fmt.Errorf("agentemit: permission contract default_sandbox %q is outside the measured sandbox value set", pc.DefaultSandbox)
	}
	for role, v := range pc.RoleSandbox {
		if role == "" || !accepted[v] {
			return fmt.Errorf("agentemit: permission contract role %q sandbox %q is outside the measured sandbox value set", role, v)
		}
	}
	return nil
}

// ContractSandbox returns the sandbox value a role's contract states.
func (pc *PermissionContract) ContractSandbox(role string) string {
	if v, ok := pc.RoleSandbox[role]; ok {
		return v
	}
	return pc.DefaultSandbox
}

// RoleRequirements derives the restrictions a role's contract requires from
// its Claude tool list and its contract sandbox. Each rule restricts what the
// Claude definition itself withholds or scopes:
//
//   - sandbox/mode: always (the role's filesystem mode).
//   - write-path-scope/path-scope: a writing role, whose write duty is
//     narrower than the whole workspace.
//   - shell/deny, subagent/deny, web/deny: the Claude role lacks that tool
//     class.
//   - mcp-server/grant and mcp-tool/subset: the role carries moai MCP tools
//     (grant exactly that server; only the named tools of it).
//   - mcp-server/deny: the role carries no moai MCP tools.
//
// @MX:ANCHOR: [AUTO] single derivation of per-role contract restrictions
// @MX:REASON: emission, the permission report, and the contract tests all read this; divergence would reintroduce a silent drop
func RoleRequirements(doc AgentDoc, man Manifest) ([]Requirement, error) {
	pc := man.PermissionContract
	if pc == nil {
		return nil, fmt.Errorf("%s: manifest carries no permission contract", doc.File)
	}
	classes := map[string]bool{}
	for _, tok := range doc.Tools {
		class, ok := classifyToken(man, tok)
		if !ok {
			return nil, fmt.Errorf("%s: unknown tool token %q — not mapped to any class in the Codex mapping manifest", doc.File, tok)
		}
		classes[class] = true
	}
	reqs := []Requirement{{Axis: "sandbox", Kind: "mode"}}
	if pc.ContractSandbox(doc.Name) != "read-only" {
		reqs = append(reqs, Requirement{Axis: "write-path-scope", Kind: "path-scope"})
	}
	if !classes["shell"] {
		reqs = append(reqs, Requirement{Axis: "shell", Kind: "deny"})
	}
	if classes["moai-mcp"] {
		reqs = append(reqs, Requirement{Axis: "mcp-server", Kind: "grant"}, Requirement{Axis: "mcp-tool", Kind: "subset"})
	} else {
		reqs = append(reqs, Requirement{Axis: "mcp-server", Kind: "deny"})
	}
	if !classes["subagent-spawn"] {
		reqs = append(reqs, Requirement{Axis: "subagent", Kind: "deny"})
	}
	if !classes["web"] {
		reqs = append(reqs, Requirement{Axis: "web", Kind: "deny"})
	}
	return reqs, nil
}

// roleVerdicts maps every requirement of one role onto the axis table and
// checks that each enforced restriction is actually written by this
// emission. emittedSandbox is "" when sandbox_mode is not emitted.
func roleVerdicts(doc AgentDoc, man Manifest, emittedSandbox string, hasMCP bool) ([]AxisVerdict, error) {
	reqs, err := RoleRequirements(doc, man)
	if err != nil {
		return nil, err
	}
	rows := map[Requirement]AxisMapping{}
	for _, m := range man.PermissionContract.Axes {
		rows[Requirement{Axis: m.Axis, Kind: m.Kind}] = m
	}
	out := make([]AxisVerdict, 0, len(reqs))
	for _, r := range reqs {
		m, ok := rows[r]
		if !ok {
			return nil, fmt.Errorf("%s: role %q requires restriction %s/%s, which is neither enforced nor declared UNSUPPORTED in the permission contract", doc.File, doc.Name, r.Axis, r.Kind)
		}
		if m.Mapping == MappingEnforced {
			if m.Basis != "measured" && m.Basis != "documented" {
				return nil, fmt.Errorf("%s: role %q restriction %s/%s is enforced on basis %q — an unmeasured axis is never enforced", doc.File, doc.Name, r.Axis, r.Kind, m.Basis)
			}
			want, known := enforceableFields[r]
			if !known || m.Field != want {
				return nil, fmt.Errorf("%s: role %q restriction %s/%s claims enforcement through field %q, which this emitter does not write", doc.File, doc.Name, r.Axis, r.Kind, m.Field)
			}
			if r == (Requirement{Axis: "mcp-server", Kind: "grant"}) && !hasMCP {
				return nil, fmt.Errorf("%s: role %q requires the moai server grant but none is emitted", doc.File, doc.Name)
			}
		}
		// The emitted sandbox_mode must equal the contract's value whatever
		// the sandbox row's mapping: an UNSUPPORTED row records that the host
		// does not enforce the field, never a licence to emit a wider mode.
		if r == (Requirement{Axis: "sandbox", Kind: "mode"}) {
			contract := man.PermissionContract.ContractSandbox(doc.Name)
			if emittedSandbox != contract {
				return nil, fmt.Errorf("%s: role %q emits sandbox_mode %q but its permission contract states %q — never emit a mode other than the contract's", doc.File, doc.Name, emittedSandbox, contract)
			}
		}
		out = append(out, AxisVerdict{Role: doc.Name, Axis: r.Axis, Kind: r.Kind, Mapping: m.Mapping, Field: m.Field, Basis: m.Basis, Reason: m.Reason})
	}
	return out, nil
}

// BuildPermissionReport emits the agent set under agentsRoot and reports the
// contract verdict of every (role, requirement). The verdicts are the ones
// the emission itself checked, so the report cannot diverge from what was
// emitted; an emission failure is the report's failure.
func BuildPermissionReport(fsys fs.FS, agentsRoot string, man Manifest) (PermissionReport, error) {
	if man.PermissionContract == nil {
		return PermissionReport{}, fmt.Errorf("agentemit: manifest carries no permission contract")
	}
	_, verdicts, err := emitAll(fsys, agentsRoot, man)
	if err != nil {
		return PermissionReport{}, err
	}
	sort.Slice(verdicts, func(i, j int) bool {
		a, b := verdicts[i], verdicts[j]
		return a.Role+"/"+a.Axis+"/"+a.Kind < b.Role+"/"+b.Axis+"/"+b.Kind
	})
	rep := PermissionReport{Verdicts: verdicts}
	for _, v := range verdicts {
		if v.Mapping == MappingEnforced {
			rep.Pass = append(rep.Pass, v)
		} else {
			rep.Unsupported = append(rep.Unsupported, v)
		}
	}
	return rep, nil
}
