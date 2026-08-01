---
paths: ".claude/skills/moai/workflows/plan.md,.claude/skills/moai/workflows/plan/**/*.md"
---

# Branch Origin Decision Protocol (BODP)

> Status: HARD operational rule for the plan-phase base-branch decision.

## Identity

BODP standardises one decision: **which branch a new SPEC's work branches from**. It is embedded in the `/moai plan --branch` skill body; it introduces no slash command and no CLI subcommand of its own.

Getting this wrong is quiet and expensive. Branching from a stale local `main` silently omits teammates' merged work; branching from `origin/main` when the user has unpushed local commits silently omits their own. Neither produces an error — the divergence surfaces later as a confusing merge.

## Entry Point

| Path | Prompts user? |
|------|----------------|
| `/moai plan --branch` (skill body) | yes (orchestrator `AskUserQuestion`) |

The orchestrator evaluates the signals below, recommends a base, and asks. There is no CLI entry point: `moai worktree new` and its `--base` / `--from-current` flags are retired, and a worktree is now entered with `moai cc -w <name>` rather than created by a MoAI subcommand.

## HARD Rules

- [ZONE:Frozen] [HARD] Skill body BODP gate MUST follow the askuser-protocol Socratic structure: `(권장)` first, ≤4 options, conversation_language match, "Other" auto-appended.
- [ZONE:Evolvable] [HARD] The recommended base MUST be derived from the signals below, not assumed. When no signal fires, the recommendation is `origin/main` — team-safe, because it reflects the latest merged state rather than whatever the local checkout happens to hold.
- [ZONE:Evolvable] [HARD] `main` (the local ref) is the explicit opt-in for solo workflows where the user has committed locally without pushing. Recommend it only when `git log main` actually shows such commits — otherwise it silently reintroduces the stale-base hazard.

## Algorithm (3-Signal Evaluation)

The orchestrator runs these three checks directly; each is one command.

| Signal | Detection |
|--------|-----------|
| A — Code dependency | the SPEC's `depends_on` list names the current branch, OR `git diff` overlaps the new SPEC's path |
| B — Working tree co-location | `git status --porcelain` already contains `.moai/specs/<NewSpecID>/` |
| C — Open PR head | `gh pr list --head <currentBranch> --state open --json number` returns ≥ 1 entry (skip gracefully when `gh` is absent) |

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

Signal B dominates: when the SPEC's files are already in the working tree, the work has effectively started on the current branch and moving it elsewhere would strand them.

When Signal C fires, note the parent-merge gotcha in the option description — a stacked branch whose parent PR merges first needs a rebase before it reads correctly.

## Retired

The following were part of BODP and are gone. They are listed so their absence reads as a decision rather than an oversight:

- **`internal/bodp` Go library** (`Check`, `WriteDecision`, `HasAuditTrail`) — its only caller was `moai worktree new`, and the plan path never had a way to reach it (no CLI exposed it).
- **`.moai/branches/decisions/` audit trail** — written only by that same retired command. Existing files are left in place; nothing writes new ones.
- **`moai status` off-protocol reminder** and its `MOAI_NO_BODP_REMINDER` opt-out — the reminder keyed on the audit trail, so without a writer it would have fired on every branch forever.
- **`EntryWorktreeCLI` / `EntryPlanWorktree` entry points** — the CLI command and the `--worktree` plan flag are both retired.

## Cross-References

- `agent-common-protocol.md` § User Interaction Boundary — orchestrator-only AskUserQuestion HARD.
- `askuser-protocol.md` § Socratic Interview Structure — option label/order rules.

---

Version: 2.0.0
