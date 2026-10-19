# SPEC-AGENTS-CONTRACT-REVISE-001 — Implementation Plan

Tier M · docs card · manager-docs-led (no manager-develop leg anticipated — no Go code changes in scope; return a blocker report if a template/build-code issue surfaces).

## §A Context

- Worktree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1900` (branch `WT-agents-agents-md`, HEAD `8249bff8c3`, clean except the untracked SPEC directory). All writes stay inside this worktree. Never touch the primary checkout or any other tree (including `.moai/worktrees/t1882` — another card's paused tree).
- SPEC artifacts: `.moai/specs/SPEC-AGENTS-CONTRACT-REVISE-001/{spec.md,plan.md,acceptance.md,progress.md,decision-index.md}`.
- plan-auditor verdict: pending (this SPEC is at plan close; the verdict lands via the independent plan-audit gate).
- Touched surfaces (11 files + new companion docs): `AGENTS.md`, `AGENTS.local.md`, `internal/template/templates/AGENTS.md.tmpl`, `.claude/rules/local/repo-local-pr-policy.md`, `.claude/rules/local/gitflow-lane-protocol.md`, `.claude/rules/moai/workflow/factory-dispatch.md` (+ mirrors), `.moai/docs/git-workflow-doctrine.md`, `.moai/docs/git-local-workflow-doctrine.md`, `.moai/docs/local-dev-guide.md`, `.moai/docs/update-local-file-survival.md`, `.moai/docs/harness-namespace-doctrine.md`, `.moai/docs/version-management.md`.

## §B Known Issues (SPEC-specific)

- **B-ACR-1 Blind patch-paste hazard**: `inputs/proposed.patch` was authored against a fixed review baseline (`1cc1c854…`), not this tree. Apply it as a basis, re-verified hunk by hunk against this tree's code and docs; never paste blindly.
- **B-ACR-2 Canonical-copy hazard**: `AGENTS.local.md` §0.2 — an uncommitted working copy cannot be cited as canonical. The diet baseline is the committed copy in this tree (50,579 chars, MF-5), not the primary working copy (31,971, operator memo).
- **B-ACR-3 Size-diet model (leader-corrected)**: the governing target is the always-loaded instruction TOTAL ≤ 140,000 chars (hard 150,000) — AGENTS.local.md's share is derived, not fixed; the withdrawn ≤12,000 figure came from a bad baseline. See the decision-index measurement table (HEAD `9b1dcde024`) and card t1887's required core-rules share (≈87,615 chars residual).
- **B-ACR-4 Ownership boundary**: t1868 (review gate), t1882 (draft-PR tooling), t1895 (config defaults), t1874 (Codex contract docs) own their surfaces. This card's prose must cross-reference, not restate or pre-claim (spec §G groups 3/4/7; REQ-ACR-004/007).
- **B-ACR-5 Batch-release retention**: policy 11 keeps `release/main-batch-YYYYMMDD` current until #1830 — the patch's wholesale §4.0-H supersession must be narrowed (spec §G group 2).
- **B-ACR-6 hotfix/* omission trap**: an earlier SPEC omitted `hotfix/*` from the merge-flag rule and failed its audit; MF-2 pins the measured rule (spec REQ-ACR-006).
- **B-ACR-7 Template-First coupling**: edits to `AGENTS.md` and the factory-dispatch family must land in their `internal/template/templates/` mirrors and pass `make build` (MF-7); the local-only files (AGENTS.local.md, `.claude/rules/local/*`, `.moai/docs/*` targets) have no mirror.
- **B-ACR-8 Frontmatter discipline**: SPEC frontmatter uses canonical fields only (`created:`/`updated:`/`tags:` — no snake_case aliases; B4 of the delegation template).
- **B-ACR-9 Heading convention**: the exclusions section needs `### Out of Scope —` H3 sub-headings with `-` bullets (B6; spec.md §H satisfies this).
- **B-ACR-10 Pattern-kill discipline (2fe7)**: never `pkill` / `killall` / `kill $(pgrep)`; stop things only via TaskStop or recorded PIDs; a guard refusal is final — no retry around it.

## §C Pre-flight

```bash
git branch --show-current && git rev-parse --short HEAD   # expect WT-agents-agents-md @ 8249bff8c3
wc -m AGENTS.local.md                                     # expect ≈ 50,579 (pre-card pin @ 9b1dcde024); a different number means the tree moved — re-pin before proceeding, never substitute into the projection
grep -rn -e "lanes do not push" -e "Permitted direct execution" \
  -e "templates/CLAUDE.md" -e "CLAUDE.md Section 4" \
  AGENTS.md AGENTS.local.md .claude/rules/local/ .claude/rules/moai/workflow/factory-dispatch*.md .moai/docs/   # conflict baseline before edits
ls internal/template/templates/AGENTS.md.tmpl internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md   # mirror targets exist
```

## §D Constraints

- PRESERVE: `inputs/` (committed evidence — never delete or edit); `AGENTS.local.md` §0 canonical-copy section (§0.1-§0.4 stay); all [HARD] obligations (moved, never deleted); historical records (marked, not erased); batch-release procedure §4.0 items 6-7 until #1830.
- Forbidden: Go code changes; remote/management setting changes; `git add -A` / `git add .` / `git commit -a`; `--no-verify`; `--amend`; force-push; time estimates in any artifact; AskUserQuestion (leaf — blocker report instead); describing t1868/t1882/t1895 capabilities as implemented.
- Required: `make build` exit 0 after any template mirror edit; Conventional Commits with the card id (`docs(SPEC-AGENTS-CONTRACT-REVISE-001): M<N> …`) and the `🗿 MoAI` trailer; stage by explicit pathspec.

## §E Self-Verification (manager-docs deliverables)

Each item reported per the 5-section evidence format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk), with command + verbatim output + tree SHA:

- **E1** — AC binary PASS/FAIL matrix over AC-ACR-001…015 (`acceptance.md`), each row naming its verification command and observed output.
- **E2** — Template cycle: `make build` exit 0 after the mirror edits (verbatim tail).
- **E3** — Size model: projection = 252,650 (pre-card total @ `9b1dcde024`) − this card's AGENTS.local.md reduction − t1887 core-rules share ≤ 140,000, computed from the PIN; fresh close-time `wc -m` values are recorded as a separate decision-index evidence row and never overwrite the baseline row nor enter the projection; reconciliation table row count vs moved-clause count.
- **E4** — Conflict grep: the AC-ACR-005 command output; every surviving hit shown inside its historical marker block.
- **E5** — Diff list: the changed-file + diff list artifact path, one row per touched doc and mirror.
- **E6** — Branch HEAD + push state (new commit SHAs; push per the card's delivery path).
- **E7** — Blocker report if any required decision was missing (no prose questions).

## §F Milestones (priority-ordered; decision-reversibility first)

### M1 (Priority High) — AGENTS.local.md restructure (highest-change decisions)

Rewrite §4.0 as the current card policy (draft-PR flow, safe handoff, receiver confirmation, preservation-push clarification; REQ-ACR-001) and reframe the superseded card-delivery clauses as the §4.0-H historical block (REQ-ACR-005), keeping items 6-7 (batch release + batch-closing) current per policy 11. Execute the size diet (REQ-ACR-009, leader-corrected model): move §4.0/§4.1/§2.0/References procedure-and-reference mass into the companion doc(s) with one-line essences + pointers — this four-chunk move is confirmed scope that proceeds regardless — then size the result against the decision-index projection (always-loaded total with t1887's share), recording any further compression in decision-index as "리더 정정"; produce the reconciliation table; write the companion doc(s) with `paths:` scoping; verify each glob loads for its claimed scope. Fold in the pointer/mechanism fixes (REQ-ACR-010), leader-boundary §16 rewrite (part of REQ-ACR-002), primary-main preservation language (REQ-ACR-013), and the CLI/slash form distinction (REQ-ACR-016).

### M2 (Priority High) — AGENTS.md + template mirror (review gate + leader boundary + review contract)

Rewrite §0 leader boundary and the agent-spawning capability row (REQ-ACR-002); replace the CodeRabbit-only clause with the producer-neutral current-head contract (REQ-ACR-003); add the review-gate policy block with the incident-derived rules, the measured trigger/observation narration (spec REQ-ACR-004 — observed triggers separate from rules, no exclusive-start claim), and the t1868 in-progress marker; add the distributed-automation policy block (REQ-ACR-007/008 limitations). Mirror only the producer-neutral changes into `internal/template/templates/AGENTS.md.tmpl` — incident references, card-id markers, and t1868 in-progress/status markers stay LOCAL-ONLY (C9 card-id leak guard, `internal/template/card_id_leak_test.go:33-56`, removal-only per-(file, literal) baseline) — then run `make build` (REQ-ACR-014).

### M3 (Priority High) — Local rules + factory-dispatch family reconciliation

Rewrite `.claude/rules/local/repo-local-pr-policy.md` (card delivery + auto-merge conditions + leader boundary) and `gitflow-lane-protocol.md` §2/§4 (worker card-branch push through the draft-PR path; release-path batch retained; preservation-push clarification; candidate-path exception kept). Reconcile the factory-dispatch family sites (factory-dispatch.md:20/183/219/222; factory-dispatch-mechanics.md §125) to the per-card flow with transition-scoped historical framing — with their template mirrors (REQ-ACR-005, REQ-ACR-014).

### M4 (Priority Medium) — .moai/docs six-file update

Apply the header notices and in-body historical scoping to `git-workflow-doctrine.md` and `git-local-workflow-doctrine.md` (REQ-ACR-011); update `local-dev-guide.md` (embed mechanism, REQ-ACR-010), `update-local-file-survival.md` + `harness-namespace-doctrine.md` + `version-management.md` (preservation contract, REQ-ACR-012). Fold the review-gate rules and policy-vs-implementation markers into the flow-doctrine prose where they are read (REQ-ACR-004/006/007/008).

### M5 (Priority Medium) — Verification battery + evidence artifacts

Run the full acceptance battery (acceptance.md AC-ACR-001…016): conflict grep, size-model projection check, reconciliation-table audit, paths-glob load checks, pointer greps, mirror-parity diff, `make build`, merge-flag `hotfix/*` check, auto-merge both-conditions wording greps, CLI/slash form check, implementation-status wording greps. Produce the changed-file + diff list artifact (REQ-ACR-014) and populate §E self-verification.

## §G Anti-Patterns

- Pasting `proposed.patch` hunks unverified (B-ACR-1).
- Marking superseded content as deleted history instead of transition-scoped records (REQ-ACR-005; AC-ACR-005 would fail).
- Writing "the tool opens draft PRs" before t1882 lands, or "the gate blocks merges" before t1868 lands (REQ-ACR-007; B-ACR-4).
- Stating the merge-flag rule without `hotfix/*` (B-ACR-6).
- Deleting a [HARD] clause instead of moving it with a pointer (REQ-ACR-009).
- Editing the root `AGENTS.md` without the `.tmpl` mirror (REQ-ACR-014).
- Editing `inputs/` (PRESERVE).

## §H Cross-References

- spec.md §C (measured facts), §D (policy baseline), §G (disposition), §H (out of scope).
- acceptance.md AC-ACR-001…015; decision-index.md Q1-Q3 (implementation-level defaults, applied).
- Related cards: t1868, t1882, t1874, t1895. Related SPECs: SPEC-AGENTS-MD-CANON-001, SPEC-LOCAL-MAIN-FLOW-001.
