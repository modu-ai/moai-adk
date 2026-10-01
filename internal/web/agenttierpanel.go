package web

import (
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// agenttierpanel.go — the agent-tier sub-section of the workflow settings
// panel (SPEC-AGENT-TIER-001 M3, REQ-TIER-011 / AC-TIER-010). It renders the
// three-tier chart grounding and the per-class tier radios as ONE surface:
// the chart table and the controls travel together, so an operator reading a
// tier's figures selects from exactly the closed set those figures describe.
//
// Same render-placement shape as the Jev sub-section (jevkey.go): the fields
// partition out of the workflow panel's generic loop, persistence stays on
// the SectionWorkflow seam, and the sub-section marker lets tests slice the
// rendered markup.

// agentTierSectionMarker is the attribute value that identifies the tier
// sub-section in the rendered markup, so a test can slice it the way
// panelHTML slices a panel.
const agentTierSectionMarker = "agent-tiers"

// agentTierSectionFields returns the schema fields the tier sub-section
// renders: one closed-set radio per known class.
func agentTierSectionFields() []settings.FieldDef {
	_, _, _, _, tiers := partitionWorkflowFields()
	return tiers
}

// agentTierChartRows returns the chart-grounding table the sub-section header
// renders (the figures' single source is config.AgentTierChartTable — this is
// a display pass-through, never a restatement).
func agentTierChartRows() []config.AgentTierChart {
	return config.AgentTierChartTable()
}
