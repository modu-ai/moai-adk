---
paths:
  - "internal/cli/worktree/**/*.go"
  - "internal/cli/status.go"
  - "internal/bodp/**/*.go"
  - ".claude/skills/moai/workflows/plan.md"
---

# Branch Origin Decision Protocol (BODP)

> Source: SPEC-V3R3-CI-AUTONOMY-001 Wave 7 (T8). REQ-CIAUT-042 ~ 051.
> Status: HARD operational rule for the 3 BODP entry points.

## Identity

BODP standardises the base-branch decision when a new SPEC plan or worktree is created. The protocol is **embedded into 3 existing entry points**; no new slash commands or CLI subcommands are introduced.

## Three Entry Points (Shared Library)

All three paths consume `internal/bodp.Check()` and `internal/bodp.WriteDecision()`:

| Path | EntryPoint | Audit | Prompts user? |
|------|------------|-------|----------------|
| `/moai plan --branch` (skill body) | `EntryPlanBranch` | yes | yes (orchestrator AskUserQuestion) |
| `/moai plan --worktree` (skill body) | `EntryPlanWorktree` | yes | yes (orchestrator AskUserQuestion) |
| `moai worktree new <SPEC-ID>` (CLI) | `EntryWorktreeCLI` | yes | **no** (orchestrator-only HARD) |

## HARD Rules

- [HARD] CLI path (`moai worktree new`) MUST NOT invoke `AskUserQuestion` — see `agent-common-protocol.md` § User Interaction Boundary. Static check: `internal/cli/worktree/new_test.go` `TestNew_NoAskUserQuestion`.
- [HARD] Default base for `moai worktree new` is `origin/main` (from `internal/bodp.DefaultBase`). The legacy default `"main"` is replaced.
- [HARD] `--base` and `--from-current` are mutually exclusive flags on `moai worktree new`.
- [HARD] Every BODP decision (skill or CLI) MUST be persisted to `.moai/branches/decisions/<normalized-branch>.md` via `bodp.WriteDecision`. Failure is non-fatal; absence of the file is the diagnostic signal for the off-protocol reminder.
- [HARD] Skill body BODP gate MUST follow the askuser-protocol Socratic structure: `(권장)` first, ≤4 options, conversation_language match, "Other" auto-appended.
- [HARD] `bodp.HasAuditTrail` MUST return false when the audit directory itself is absent (fresh project). This prevents the off-protocol reminder from firing on freshly-cloned repositories.

## Algorithm (3-Signal Evaluation)

| Signal | Detection | Source |
|--------|-----------|--------|
| A — Code dependency | SPEC `depends_on` list intersects with currentBranch name OR git diff overlaps NewSpecID path | `bodp.checkSignalA` |
| B — Working tree co-location | `git status --porcelain` contains `.moai/specs/<NewSpecID>/` | `bodp.checkSignalB` |
| C — Open PR head | `gh pr list --head <currentBranch> --state open --json number` returns ≥ 1 entry | `bodp.checkSignalC` (graceful skip on gh missing) |

## Decision Matrix (verbatim 8-row truth table)

```
¬a ¬b ¬c → main      @ origin/main
 a ¬b ¬c → stacked   @ currentBranch
¬a  b ¬c → continue  @ ""
¬a ¬b  c → stacked   @ currentBranch
 a  b ¬c → continue  @ "" (b dominates)
 a ¬b  c → stacked   @ currentBranch
¬a  b  c → continue  @ "" (b dominates)
 a  b  c → continue  @ "" (b dominates)
```

When SignalC fires, Rationale is suffixed with the parent-merge gotcha pointer (`§18.11 Case Study`) per REQ-CIAUT-047b.

## Off-Protocol Reminder

`internal/cli/status.go` `emitOffProtocolReminder()` runs at the end of `moai status`. It writes a friendly notice when **all** of the following are true:

- `MOAI_NO_BODP_REMINDER` env var is unset / not "1".
- Current branch is not `main` / `master`.
- `bodp.HasAuditTrail(repoRoot, currentBranch)` returns false.
- `.moai/branches/decisions/` directory exists.

The notice mentions the branch name and the opt-out env var. Exit code is unaffected.

## Cross-References

- SPEC artifacts: `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/` strategy-wave7.md + tasks-wave7.md.
- CLAUDE.local.md §18.12 — dev-project specific notes (stacked PR Case Study reference is §18.11).
- `agent-common-protocol.md` § User Interaction Boundary — orchestrator-only AskUserQuestion HARD.
- `askuser-protocol.md` § Socratic Interview Structure — option label/order rules.

---

Version: 1.0.0
Source: SPEC-V3R3-CI-AUTONOMY-001 Wave 7 (W7-T06)
