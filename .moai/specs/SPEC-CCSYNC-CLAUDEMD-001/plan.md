# Plan — SPEC-CCSYNC-CLAUDEMD-001

> Implementation plan for the Claude Code instruction-layer doc sync. Plan-phase
> artifact authored by manager-spec. Run-phase execution owned by manager-develop.

## A. Context

This SPEC remediates three verified documentation-drift findings (H1/H2/H8) plus a
low-priority dead-reference bundle, all in the instruction layer (CLAUDE.md both
copies + `.claude/rules/moai/**` dev-root and template mirror). Baseline HEAD
`5042e309c`. The sync-phase auditor agent is named `sync-auditor` at this baseline.

The work is content-edit + build-regeneration only — no Go logic changes. The
primary risks are (1) template-neutrality leakage when porting dev-root content,
(2) mirror byte-parity drift, and (3) an intra-batch edit collision with the
sibling SPEC `SPEC-CCSYNC-TOOLCAT-001` over `agent-authoring.md`.

## B. Known Issues / Pre-existing State (verified at HEAD 5042e309c)

| Item | Location | Current state |
|------|----------|---------------|
| Template version | `internal/template/templates/CLAUDE.md` line ~619 | `Version: 14.1.0` (dev-root is `14.2.0` line ~621) |
| §2 archived ref | template CLAUDE.md line ~61 | `expert-backend` invocation example |
| §5 stale chain | template CLAUDE.md lines ~155-162 | `manager-strategy`/`expert-backend`/`expert-frontend`/`manager-quality` |
| §11 archived refs | template CLAUDE.md lines ~383, ~386 | `manager-quality`, `expert-devops` |
| 1M=75% (H2) | CLAUDE.md ~533, template CLAUDE.md ~531 | "1M = 75%, 200K = 90%" |
| 75% window (H2) | settings-management.md ~214 (both copies) | "approach 75% of the window" |
| v2.1.50 (H8) | CLAUDE.md ~463, template CLAUDE.md ~461 | "Claude Code v2.1.50 or later" |
| spawn claim (H8) | agent-authoring.md line ~212 (both copies) | "teammates CAN spawn other teammates ... v2.1.50+" |
| dead ref | CLAUDE.md ~626, template CLAUDE.md ~626 | `workflow/progressive-disclosure.md` (file absent) |

Note: line numbers were re-verified at HEAD `5042e309c` during plan authoring. The
run phase MUST Grep/Read each file to re-confirm exact lines before editing (a
parallel session may shift lines by a few).

## C. Pre-flight Checklist (run-phase entry)

- [ ] `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` — confirm no parallel-session divergence (per agent-common-protocol §Pre-Spawn Sync Check).
- [ ] Confirm `SPEC-CCSYNC-TOOLCAT-001` (sibling sharing `agent-authoring.md`) is NOT mid-run; this SPEC runs FIRST (CON-5 / §F sequencing).
- [ ] Re-Grep all 9 drift locations in §B to confirm current line numbers.
- [ ] Confirm `make build` works in a clean tree before starting (baseline green).

## D. Constraints (binding)

- **Template-First** (CLAUDE.local.md §2): edit `internal/template/templates/` source, then `make build`. Never hand-edit `embedded.go`.
- **Template Internal-Content Isolation** (CLAUDE.local.md §25): ported template content stays generic — strip internal SPEC IDs / REQ / AC / audit citations / dates / SHAs / archive paths / memory refs.
- **Mirror parity** (`internal/template/rule_template_mirror_test.go`): settings-management.md + agent-authoring.md each have a `templates/` mirror — edit both copies in the same commit.
- **Hybrid Trunk Tier M** (CLAUDE.local.md §23.9): main-direct, no PR, unless user escalates with `--pr`. No `manager-git` routing by default.

## E. Self-Verification (run-phase exit, read-only batch)

The run phase must execute the acceptance.md AC greps as a single parallel Bash
batch (per verification-batch-pattern.md), plus:

```bash
go test ./internal/template/...                                 # neutrality + leak + mirror-drift
go test ./internal/template/... -run TestTemplateNeutralityAudit  # isolated neutrality
make build && git diff --stat internal/template/embedded.go     # embedded regenerated
```

## F. Milestones (priority-ordered, no time estimates)

### M1 — Finding H1: template CLAUDE.md catch-up (Priority High)

Port the dev-root CLAUDE.md v14.2.0 content for §2/§5/§11 into the template copy and
bump the template `Version:` line to 14.2.0+.

- M1.1: Re-Grep template CLAUDE.md §2 (line ~61), §5 (lines ~155-162), §11 (lines ~383/386), version line (~619).
- M1.2: §2 Phase 3 Execute — replace the `expert-backend` example with the retained `manager-develop` invocation example (dev-root lines 61-63 are the reference wording: `cycle_type=tdd, domain context: backend`).
- M1.3: §5 Agent Chain — replace the 6-phase archived chain with the dev-root 4-phase/6-phase retained chain (manager-spec → plan-auditor → manager-develop → manager-docs/sync-auditor → optional manager-git). Use `sync-auditor` (NOT evaluator-active) per baseline.
- M1.4: §11 Error Recovery — replace `manager-quality` / `expert-devops` routing with the dev-root retained-agent + per-spawn `Agent(general-purpose)` wording.
- M1.5: Bump template CLAUDE.md `Version:` line to `14.2.0` (or higher), updating the change-note block analogously to the dev-root.
- M1.6: §25 neutrality pass on all ported text — strip any internal SPEC IDs / REQ / dates / SHAs that the dev-root carries but the template must not. (The dev-root change-note references `SPEC-V3R6-AGENT-TEAM-REBUILD-001`; the template port MUST genericize per §25.2 substitution patterns.)

Covers REQ-CCSYNC-001..006 (006 satisfied at M5 build step).

### M2 — Finding H2: context-window threshold alignment (Priority High)

Replace "1M = 75%" → "1M = 50%" in both CLAUDE.md copies; update both
settings-management.md copies to the model-specific threshold (1M=50%, 200K=90%).

- M2.1: CLAUDE.md §16 (line ~533) + template CLAUDE.md §16 (line ~531): `1M = 75%` → `1M = 50%`.
- M2.2: settings-management.md (line ~214) + template mirror: "approach 75% of the window" → model-specific phrasing referencing 1M=50% / 200K=90% (consistent with context-window-management.md SSOT).

Covers REQ-CCSYNC-007, REQ-CCSYNC-008. Mirror-parity (REQ-CCSYNC-013) applies to settings-management.md.

### M3 — Finding H8: Agent Teams version + spawn-capability reconciliation (Priority High)

- M3.1: CLAUDE.md §15 (line ~463) + template CLAUDE.md §15 (line ~461): `v2.1.50` → `v2.1.32`.
- M3.2: agent-authoring.md line ~212 (dev-root + template mirror): reconcile the spawn-capability wording to match the official limitation (no "teammates CAN spawn other teammates"); make it internally consistent with line 92 ("subagents cannot spawn other subagents"); remove/correct the co-located `v2.1.50+` string.

Covers REQ-CCSYNC-009, REQ-CCSYNC-010, REQ-CCSYNC-011. Mirror-parity (REQ-CCSYNC-013) applies to agent-authoring.md.

### M4 — Low-priority bundle: dead changelog reference (Priority Medium)

- M4.1: CLAUDE.md (line ~626) + template CLAUDE.md (line ~626): replace `workflow/progressive-disclosure.md` with `development/skill-authoring.md`, or remove the dead reference.

Covers REQ-CCSYNC-012.

### M5 — Build regeneration + verification (Priority High, must follow M1-M4)

- M5.1: `make build` to regenerate `internal/template/embedded.go`.
- M5.2: Run the acceptance.md AC grep batch + `go test ./internal/template/...` (neutrality + leak + mirror-drift) in a single parallel Bash batch.
- M5.3: Confirm `git diff --stat internal/template/embedded.go` shows the regeneration.

Covers REQ-CCSYNC-006 and the cross-cutting verification of all prior milestones.

## G. Anti-Patterns to Avoid (run-phase)

- **AP-1**: Editing `embedded.go` directly instead of regenerating via `make build`.
- **AP-2**: Porting the dev-root change-note verbatim into the template (would leak `SPEC-V3R6-AGENT-TEAM-REBUILD-001` SPEC ID + 2026-05-25 date — §25 violation). Genericize.
- **AP-3**: Editing only one side of a mirror pair (settings-management.md or agent-authoring.md) — breaks `rule_template_mirror_test.go`.
- **AP-4**: Touching the docs-site agent-guide.md (EXC-1 — sibling docs-site SPEC owns it).
- **AP-5**: Modifying `context-window-management.md` or `session-handoff.md` (EXC-3 — already canonical-correct; this SPEC aligns drifted copies TO them).
- **AP-6**: Running concurrently with `SPEC-CCSYNC-TOOLCAT-001` on `agent-authoring.md` (CON-5 — sequence this SPEC first).
- **AP-7**: Drive-by edits to unrelated CLAUDE.md sections (EXC-4).

## H. Shared-File Sequencing vs SPEC-CCSYNC-TOOLCAT-001 (CON-5)

`agent-authoring.md` (dev-root + template mirror) is a run-phase target of BOTH
SPECs, on the SAME physical line (line ~212): the `TodoWrite` token and the
co-located `teammates CAN spawn other teammates … v2.1.50+` parenthetical share
one line.

- THIS SPEC (M3.2 — spawn-capability + version reconciliation on line ~212: the
  `teammates CAN spawn other teammates` claim + the `v2.1.50+` string).
- Sibling `SPEC-CCSYNC-TOOLCAT-001` (M2.3 — the `TodoWrite` → Task* recommendation
  on the SAME line ~212).

Sequencing rule: **this SPEC (CCSYNC-CLAUDEMD-001) runs FIRST**. The sibling
SPEC-CCSYNC-TOOLCAT-001 rebases onto this SPEC's commit before its own run phase.
Because both SPECs edit the SAME line in the same file (source + mirror), running
them concurrently would cause an intra-batch edit collision. The orchestrator MUST
confirm the sibling is not mid-run at the pre-flight check (§C).

## I. Tier & Routing

- Tier M (standard). Estimated scope: ~6 files (CLAUDE.md ×2, settings-management.md ×2, agent-authoring.md ×2) + embedded.go regeneration, content-edit only.
- Hybrid Trunk Tier M → main-direct, no PR (CLAUDE.local.md §23.9), unless user escalates with `--pr` (then manager-git PR routing applies).

## J. Cross-References

- spec.md — requirements + exclusions.
- acceptance.md — Given-When-Then + AC grep matrix.
- CLAUDE.local.md §2 (Template-First), §15 (neutrality), §23.9 (Tier routing), §25 (internal-content isolation).
- `.claude/rules/moai/workflow/context-window-management.md` — threshold SSOT.
- `internal/template/rule_template_mirror_test.go` — mirror byte-parity invariant.
