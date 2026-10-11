package hook

// lane_spawn_authority.go — the standing spawn authority every lane session
// carries in its bootstrap context (card t224).
//
// Why this exists: the tk8hce factory run (2026-08-24) produced two lanes that
// REFUSED to spawn the phase-required specialist (manager-spec) because the
// runtime's default agent-usage guidance — "do not spawn subagents unless the
// user asks" — stood unoverridden in their bootstrap context. The leader's
// approval could not lift that instruction, because the leader is not the lane's
// user; a peer message is inert against a session instruction. Both lanes fell
// back to direct edits, which routed SPEC-body writes around the Status
// Transition Ownership Matrix — the exact outcome the matrix exists to
// prevent. The authority must therefore be part of what the lane reads at
// startup, in its own voice, as a STANDING grant rather than a per-dispatch
// exception.
//
// The three design decisions the sentence encodes (card t224; decision 1
// reworded for the notice diet, card t1335) — and which every locale's entry
// in the message table preserves semantically
// (SPEC-SESSION-START-GUIDE-I18N-001 REQ-002):
//
//  1. Scope — the authority does not inline the per-phase specialist mapping;
//     it delegates the mapping to the pointer: the Status Transition Ownership
//     Matrix (.claude/rules/moai/development/spec-frontmatter-schema.md)
//     names which specialist each work phase routes to, and the authority
//     adds the workflow chain's prescribed auditors on top. Not arbitrary
//     spawning — the pointer, not this sentence, is the mapping's home.
//  2. Depth — depth-1 only: agents a lane spawns are leaf workers and never
//     spawn further agents, the same flat-hierarchy seal
//     manager_lead_depth_test.go enforces for the leader's own fan-out.
//  3. Placement — the bootstrap context is the operative layer (what the lane
//     actually reads; a peer message cannot override a session instruction).
//     The normative text lives in the doctrine files (the Factory Dispatch Protocol,
//     agent-common-protocol.md, moai-constitution.md, manager-lead.md), and
//     the runtime wiring already permits the spawn: Agent ships in the
//     template's permissions.allow and the launcher seeds the per-lane
//     concurrent-subagent cap (seedLaneAgentCap, t118 axis).
//
// The sentence text itself lives in the message table
// (session_start_factory_i18n.go, the laneSpawnAuthority field). Card t1603
// moved it there from the English-only const that used to sit at this spot:
// the two-audience English rule is amended for the SessionStart bootstrap
// guide surface ONLY (decision-index Q1), so the authority renders in the
// session's conversation_language and the en entry keeps the canonical
// sentence verbatim. The matrix pointer path stays a protocol token in every
// locale.
