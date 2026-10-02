package config

// audit_plan.go — the audit plan resolver (SPEC-AUDIT-MODEL-CONVERGE-001 M2,
// REQ-ACV-001..004, design.md §D.1-§D.3).
//
// ResolveAuditPlan turns the audit values a tree's workflow.yaml carries (the
// AuditConfig value its section file unmarshals to) plus the gates a caller
// supplied into one plan: for each of the three backends exactly one gate from
// off|advisory|required, the source that decided it, and whether the operator
// wrote it. Every surface that needs a gate reads it from here, so one tree
// configuration yields one reading.
//
// The resolver is a PURE function. It reads no file, environment variable or
// clock, and its input MUST be the raw section value: a merged, default-filled
// configuration pairs the model token "claude" with the default gates, which would
// hand "claude" to every project that never wrote one. The empty token (the key is
// absent) is therefore a distinct input from an explicit "claude".
//
// @MX:NOTE: [AUTO] ResolveAuditPlan — the one gate reading REQ-ACV-005 assigns to audit_multi, the convergence enforcement, codex_audit, the receipt predicate and the plan verb (consumers land in M3/M4; promote to ANCHOR once fan_in >= 3)
// @MX:SPEC: SPEC-AUDIT-MODEL-CONVERGE-001

import (
	"fmt"
	"strings"
)

// Source labels: which of the four precedence levels decided a backend's gate.
const (
	// AuditPlanSourceArgument is a gate the caller supplied for that backend.
	AuditPlanSourceArgument = "argument"
	// AuditPlanSourceConfigGates is the tree's audit.gates.<backend> value.
	AuditPlanSourceConfigGates = "config.gates"
	// AuditPlanSourceConfigModel is the cell the tree's audit.model token assigns.
	AuditPlanSourceConfigModel = "config.model"
	// AuditPlanSourceDefault is the distributed default profile
	// (claude required, codex required, glm advisory).
	AuditPlanSourceDefault = "default"
)

// auditPlanBackends is the stable backend order of AuditPlan.Backends and of
// the cell arrays below.
var auditPlanBackends = [3]string{"claude", "codex", "glm"}

// AuditPlanEntry is one backend's resolved gate.
type AuditPlanEntry struct {
	// Backend is claude, codex or glm.
	Backend string `json:"backend"`
	// Gate is exactly one of off, advisory, required.
	Gate string `json:"gate"`
	// Source is one of the AuditPlanSource* labels.
	Source string `json:"source"`
	// Explicit reports that an operator wrote this gate: its source is
	// config.gates or config.model. A caller-supplied gate is not explicit, and
	// neither is the distributed default. Only an explicit required gate makes
	// an unanswered backend fail the verdict.
	Explicit bool `json:"explicit"`
}

// AuditPlan is the resolved plan for one tree.
type AuditPlan struct {
	// Model is the tree's audit.model token with surrounding whitespace trimmed;
	// empty when the key is absent.
	Model string `json:"model"`
	// Backends holds one entry per backend, in claude, codex, glm order.
	Backends []AuditPlanEntry `json:"backends"`
}

// ModelSource reports where the model token came from: "config" when the tree
// wrote one, "default" when the key is absent.
func (p AuditPlan) ModelSource() string {
	if p.Model == "" {
		return AuditPlanSourceDefault
	}
	return "config"
}

// FromConfig reports whether any backend's gate came from the tree's
// configuration (audit.gates or the audit.model token) — the condition under
// which a result carries plan_source "config".
func (p AuditPlan) FromConfig() bool {
	for _, e := range p.Backends {
		if e.Source == AuditPlanSourceConfigGates || e.Source == AuditPlanSourceConfigModel {
			return true
		}
	}
	return false
}

// ExplicitGates returns the gates an operator wrote — the plan's explicit
// entries — with every other field left empty. Empty reads as "not configured":
// the fail-closed readers key on what the project wrote, never on the
// distributed default. An explicit off or advisory is returned too; only an
// explicit required enforces anything.
func (p AuditPlan) ExplicitGates() AuditGates {
	var g AuditGates
	for _, e := range p.Backends {
		if !e.Explicit {
			continue
		}
		switch e.Backend {
		case auditPlanBackends[0]:
			g.Claude = e.Gate
		case auditPlanBackends[1]:
			g.Codex = e.Gate
		case auditPlanBackends[2]:
			g.GLM = e.Gate
		}
	}
	return g
}

// auditPlanCell is what a model token assigns to one backend.
type auditPlanCell struct {
	gate string
	// fromModel marks the cell as the operator's choice (source config.model);
	// false leaves the backend on the distributed default.
	fromModel bool
}

// auditPlanCells is the table of design.md §D.1 for a validated token, in
// claude, codex, glm order.
func auditPlanCells(model string) [3]auditPlanCell {
	def := [3]auditPlanCell{
		{AuditGateRequired, false}, {AuditGateRequired, false}, {AuditGateAdvisory, false},
	}
	switch model {
	case AuditModelClaude:
		// Decision D7': Claude's gate is the operator's, codex and glm stay on
		// the default profile, so audit_multi behaves as it did before the token
		// carried meaning. Claude alone is spelled out with audit.gates.
		def[0].fromModel = true
		return def
	case AuditModelCodex:
		return [3]auditPlanCell{{AuditGateOff, true}, {AuditGateRequired, true}, {AuditGateOff, true}}
	case AuditModelGLM:
		return [3]auditPlanCell{{AuditGateOff, true}, {AuditGateOff, true}, {AuditGateRequired, true}}
	case AuditModelMulti:
		// The cross-model profile has the default's gates; the difference is the
		// source: an operator wrote it, so it is fail-closed.
		for i := range def {
			def[i].fromModel = true
		}
		return def
	default: // the empty token: the engine default profile, not an operator choice
		return def
	}
}

// gateAt returns the backend's gate in g by position in auditPlanBackends.
func gateAt(g AuditGates, i int) string {
	switch i {
	case 0:
		return g.Claude
	case 1:
		return g.Codex
	default:
		return g.GLM
	}
}

// normalizeAuditGate trims value and checks it against the closed gate set. An
// empty value is "not supplied" and returns "". The error names the key, the
// value and the accepted values (REQ-ACV-003).
func normalizeAuditGate(key, value string) (string, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return "", nil
	}
	valid := ValidAuditGates()
	for _, g := range valid {
		if v == g {
			return v, nil
		}
	}
	return "", fmt.Errorf("%s %q unknown (want one of %s)", key, v, strings.Join(valid, "|"))
}

// ResolveAuditPlan resolves the plan for one tree. audit is the AuditConfig the
// tree's workflow.yaml section file carries — read raw, never default-merged —
// and supplied is the gates the caller put in its call (zero fields mean "not
// supplied"). Per backend, highest first: the supplied gate, the tree's
// audit.gates value, the gate the audit.model token assigns, the distributed
// default.
//
// An audit.model outside the closed set, or any gate outside off|advisory|
// required (configured or supplied), returns an error and a zero plan: a
// mistyped value must reach the caller and never become the default.
func ResolveAuditPlan(audit AuditConfig, supplied AuditGates) (AuditPlan, error) {
	model := strings.TrimSpace(audit.Model)
	if model != "" {
		known := false
		for _, m := range ValidAuditModels() {
			if model == m {
				known = true
				break
			}
		}
		if !known {
			return AuditPlan{}, fmt.Errorf("audit_model %q unknown (want one of %s)", model, strings.Join(ValidAuditModels(), "|"))
		}
	}

	var configured, argument [3]string
	for i, backend := range auditPlanBackends {
		var err error
		if configured[i], err = normalizeAuditGate("audit.gates."+backend, gateAt(audit.Gates, i)); err != nil {
			return AuditPlan{}, err
		}
	}
	for i, backend := range auditPlanBackends {
		var err error
		if argument[i], err = normalizeAuditGate("gates."+backend, gateAt(supplied, i)); err != nil {
			return AuditPlan{}, err
		}
	}

	cells := auditPlanCells(model)
	plan := AuditPlan{Model: model, Backends: make([]AuditPlanEntry, 0, len(auditPlanBackends))}
	for i, backend := range auditPlanBackends {
		e := AuditPlanEntry{Backend: backend, Gate: cells[i].gate, Source: AuditPlanSourceDefault}
		switch {
		case argument[i] != "":
			e.Gate, e.Source = argument[i], AuditPlanSourceArgument
		case configured[i] != "":
			e.Gate, e.Source, e.Explicit = configured[i], AuditPlanSourceConfigGates, true
		case cells[i].fromModel:
			e.Source, e.Explicit = AuditPlanSourceConfigModel, true
		}
		plan.Backends = append(plan.Backends, e)
	}
	return plan, nil
}
