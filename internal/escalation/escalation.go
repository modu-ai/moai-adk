// Package escalation is the contract-mode escalation detector
// (SPEC-AUTONOMY-ESCALATION-001). It detects when a contract-mode run leaves
// its signed contract and reports each trip as an escalation record beside the
// queue; it never denies, asks about, or alters a tool call.
//
// This file holds the activation gate and the contract resolver. The gate is
// the first statement on every detector path: under any mode other than
// "contract" the detector reads nothing beyond the already-loaded
// configuration and writes nothing.
package escalation

import (
	"github.com/modu-ai/moai-adk/internal/config"
)

// Active reports whether the escalation detector runs for the given effective
// workflow.autonomy settings. Only the effective mode "contract" activates it;
// config.ResolveAutonomy already maps an absent, empty, or unrecognized mode
// to guided, so this is the single gate every detector path checks first
// (REQ-AE-001).
//
// @MX:ANCHOR: [AUTO] activation gate of the escalation detector — the first check on every hook and checkpoint path
// @MX:REASON: REQ-AE-001 requires byte-identical hook output under guided; every detector entry point (PreToolUse, PostToolUse, Stop, checkpoints) must return here before any file read, so the gate cannot be duplicated or reordered
func Active(s config.AutonomySettings) bool {
	return s.Mode == config.AutonomyModeContract
}
