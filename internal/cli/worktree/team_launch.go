// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M1 pattern dispatcher scaffolding.
//
// Defines the Pattern enum + decidePattern pure function + TeamLaunchConfig
// struct. Execution wiring (launchP1, launchP3, printHandoff) lands in M2/M3.

package worktree

import "time"

// Pattern enumerates the four canonical team-launch dispatch patterns.
//
//	P4Handoff       — no --team flag; print handoff guidance only.
//	P3InProgress    — --team + not in tmux; syscall.Exec the LLM in-process.
//	P2TmuxCC        — --team + in tmux + not CG mode; open `moai cc` window.
//	P1TmuxGLM       — --team + in tmux + CG mode; open `moai glm` window.
//
// Decision matrix is implemented in decidePattern. The default zero-value is
// PatternP4Handoff so an uninitialized TeamLaunchConfig defaults to the
// safest, most observable behavior.
type Pattern int

const (
	// PatternP4Handoff is the default no-flag handoff. Worktree is created and
	// stdout receives paste-ready guidance for the user to start a session.
	PatternP4Handoff Pattern = iota

	// PatternP3InProgress is the no-tmux launch path. The current process
	// chdirs to the worktree and exec's the LLM binary in-place.
	PatternP3InProgress

	// PatternP2TmuxCC is the in-tmux, non-CG launch path. A new tmux window is
	// spawned in the user's current session running `moai cc`.
	PatternP2TmuxCC

	// PatternP1TmuxGLM is the in-tmux, CG-mode launch path. A new tmux window
	// is spawned in the user's current session running `moai glm`.
	PatternP1TmuxGLM
)

// String returns a stable, human-readable identifier suitable for logs and
// swarm registry mode strings.
func (p Pattern) String() string {
	switch p {
	case PatternP4Handoff:
		return "P4-handoff"
	case PatternP3InProgress:
		return "P3-in-progress"
	case PatternP2TmuxCC:
		return "P2-tmux-cc"
	case PatternP1TmuxGLM:
		return "P1-tmux-glm"
	default:
		return "unknown"
	}
}

// decidePattern returns the dispatch pattern for the supplied input signals.
// It is a pure function — no I/O, no globals — and is intended to be invoked
// by runNew after the worktree is created and after tmux.IsCGMode evaluation.
//
// Decision matrix:
//
//	teamFlag=false               → P4Handoff   (handoff only)
//	teamFlag=true, inTmux=false  → P3InProgress (syscall.Exec)
//	teamFlag=true, inTmux=true,  cgMode=true   → P1TmuxGLM
//	teamFlag=true, inTmux=true,  cgMode=false  → P2TmuxCC
func decidePattern(teamFlag, inTmux, cgMode bool) Pattern {
	if !teamFlag {
		return PatternP4Handoff
	}
	if !inTmux {
		return PatternP3InProgress
	}
	if cgMode {
		return PatternP1TmuxGLM
	}
	return PatternP2TmuxCC
}

// TeamLaunchConfig carries the resolved launch context from decidePattern
// through to the M2+ dispatch functions. Fields are populated by runNew after
// worktree creation succeeds.
//
// LLM is "glm" for PatternP1TmuxGLM and "cc" otherwise. LaunchTime is the
// monotonic timestamp captured at dispatch — used by swarm registry entries
// (M4 REQ-WTL-008).
type TeamLaunchConfig struct {
	Pattern      Pattern
	WorktreePath string
	Branch       string
	SpecID       string
	LLM          string // "glm" or "cc"
	LaunchTime   time.Time
}
