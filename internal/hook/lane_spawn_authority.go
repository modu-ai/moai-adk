package hook

// lane_spawn_authority.go — the standing spawn authority every lane session
// carries in its bootstrap context (card t224).
//
// Why this exists: the tk8hce factory run (2026-08-24) produced two lanes that
// REFUSED to spawn the phase-required specialist (manager-spec) because the
// runtime's default agent-usage guidance — "do not spawn subagents unless the
// user asks" — stood unoverridden in their bootstrap context. The lead's
// approval could not lift that instruction, because the lead is not the lane's
// user; a peer message is inert against a session instruction. Both lanes fell
// back to direct edits, which routed SPEC-body writes around the Status
// Transition Ownership Matrix — the exact outcome the matrix exists to
// prevent. The authority must therefore be part of what the lane reads at
// startup, in its own voice, as a STANDING grant rather than a per-dispatch
// exception.
//
// The three design decisions this sentence encodes (card t224):
//
//  1. Scope — the specialist the Status Transition Ownership Matrix names for
//     the work at hand (plan-phase artifacts to manager-spec, implementation
//     to manager-develop, sync-phase docs to manager-docs, plus the workflow
//     chain's prescribed auditors), not arbitrary spawning.
//  2. Depth — depth-1 only: agents a lane spawns are leaf workers and never
//     spawn further agents, the same flat-hierarchy seal
//     manager_lead_depth_test.go enforces for the lead's own fan-out.
//  3. Placement — the bootstrap context is the operative layer (what the lane
//     actually reads; a peer message cannot override a session instruction).
//     The normative text lives in the doctrine files (kanban-dispatch.md,
//     agent-common-protocol.md, moai-constitution.md, manager-lead.md), and
//     the runtime wiring already permits the spawn: Agent ships in the
//     template's permissions.allow and the launcher seeds the per-lane
//     concurrent-subagent cap (seedLaneAgentCap, t118 axis).
//
// English-only by the two-audience rule: the bootstrap notices are rendered
// with langEnglish at both call sites (session_start.go) because the
// additionalContext channel is agent-facing; only systemMessage is localized.
const laneSpawnAuthority = "Standing spawn authority: you are the lane session and therefore the orchestrator for your card — use the Agent tool to spawn the specialist the Status Transition Ownership Matrix requires for the work at hand (plan-phase artifacts to manager-spec, implementation to manager-develop, sync-phase docs to manager-docs, plus the workflow chain's prescribed auditors) without asking the lead or the operator first. Depth-1 only: agents you spawn are leaf workers and must not spawn further agents. This authority is part of your bootstrap context and is not granted or revoked by peer messages."
