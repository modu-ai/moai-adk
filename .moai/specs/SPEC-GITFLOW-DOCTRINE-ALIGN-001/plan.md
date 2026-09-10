# plan.md — SPEC-GITFLOW-DOCTRINE-ALIGN-001

Tier: S (see §D.1 for the classification justification) · Artifacts: spec.md + plan.md + acceptance.md + research.md + progress.md (lane dispatch explicitly mandates the full five-file set; the Tier S 2-file budget is knowingly exceeded — recorded, deliberate).

## §A Context

Card t310 (kanban dispatch from lead-1 to lane-5). The repo switched GitHub Flow -> git-flow on 2026-08-27; three local-only doctrine documents received header notices but retain pre-transition bodies that now read as inverted [HARD] rules when a reader greps past the header. Scope, defect locations, and the annotation convention to extend are fixed in research.md (measured at d29b8942e). This SPEC is documentation-only: no code, config, workflow behavior, or CI changes.

## §B Known Issues

- **KI-1**: The body defects are actively harmful, not cosmetic — `repo-local-pr-policy.md` loads via `paths:` on run/sync/moai workflow touches and currently tells loaded sessions that all change needs a PR against `main`.
- **KI-2**: Both ❌ develop bullets sit inside forbidden LISTS; strike-through is not an acceptable repair there (list context has no protective frame) — the list slot must be corrected in place to hold a true prohibition.
- **KI-3**: A naive "delete the old content" pass would destroy the audit trail convention both doctrine files already follow (`[RETIRED <date>]`, `~~strike-through~~`, dated `POLICY CHANGE` blockquotes). History stays where it explains.

## §C Pre-flight (run-phase entry conditions)

1. Worktree check: `pwd` -> `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t310`; `git branch --show-current` -> `WT-git-doctrine-align`. All edits via absolute paths under this tree.
2. Re-measure baselines before editing (they decay): re-run the §B greps in research.md; if line numbers drifted vs d29b8942e, update research.md with a fresh tree SHA before editing.
3. Confirm card t308 has NOT landed CLAUDE.local.md edits into this tree mid-flight (it is out of scope here regardless — the scope fence AC-GDA-008 enforces it mechanically).
4. No user channel exists on this lane: any genuinely operator-level decision returns a structured blocker report instead of planting an unresolved clarification-gate marker. None were needed at plan time (research.md §F-6 resolved conservatively toward the canon).

## §D Constraints

### §D.1 Tier justification (S vs M)

By `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier: target-file churn estimated ~120–180 changed lines across exactly 3 files — under BOTH the Tier S axes (<300 LOC, <5 files). REQ count 6 and AC count 8 are within the Tier S ceilings (8/8) EXACTLY at the AC edge — one more criterion forces a split or a tier-up. M was considered because three separate files each carry their own wording decisions, but none involves cross-subsystem consistency risk beyond what acceptance.md's grep battery already pins; S stands.

### §D.2 Editorial conventions (binding on the implementer)

- Prefer ANNOTATION + correction where history is explanatory (routing tables, rationale prose): retired text survives behind `[RETIRED 2026-08-27]` blockquote annotations or `~~strike-through~~`, matching the `[RETIRED 2026-07-20]` precedent already in these files.
- Plain replacement ONLY where the old bullet would stand alone as a dangerous falsehood: the two ❌ develop bullets (§18.0.1 line 52, §18.10 line 351).
- Edited prose is written in Korean native register matching each file's voice (the doctrine files' language); annotation markers stay in the established English marker form.
- Do not renumber §18.x / §23.x headings; existing cross-references must keep resolving.
- Native UTF-8 throughout — no `\uXXXX` escapes in any edit payload.

## §E Self-Verification (E1–E7)

| # | Check | Method |
|---|-------|--------|
| E1 | AC matrix | Run all eight commands in acceptance.md; every one must return its stated PASS shape. Capture verbatim outputs to `.moai/state/verify/<session>/t310-ac-battery.txt`. |
| E2 | Build/tests | N/A — documentation-only; no Go code touched. Instead: nothing under `internal/`, `pkg/`, `cmd/` appears in the diff (AC-GDA-008 proves it). |
| E3 | Coverage | N/A (no code). |
| E4 | Subagent boundary | Not applicable (single-lane doc edits, no spawns planned). |
| E5 | Lint/format | Markdown only — verify UTF-8 integrity and no mojibake: `grep -n $'�' <target>` returns 0 per edited file. |
| E6 | Push state | Lane merges to local `develop` + pushes `origin/develop` per gitflow-lane-protocol §4; AC-GDA-008 runs against this tree BEFORE integration. |
| E7 | Count = 7 checks executed, evidence path cited in progress.md §E.1 at close. |

## §F Milestones (ordered by decision-reversibility; priority labels, no time estimates)

### M1 — git-workflow-doctrine.md alignment [Priority High]

The wording decisions most likely to receive review pushback land first. Edits:

1. **Line 15 rationale** — wrap as retained-as-history and add the operative correction beneath it:
   - Prefix the paragraph with `> **[RETIRED 2026-08-27]** 아래 단락은 v2.14 시점의 기각 판단이다 — 2026-08-27 운영자 지시로 Gitflow 전환하면서 뒤집혔다(상단 고지 참조).`
   - Append after it one operative sentence: current model = git-flow; branching/integration rules live in the header notice's canonical sources.
   *Class*: annotated-retention (history explains WHY develop support exists today despite the old rejection).
2. **Line 52 bullet (§18.0.1)** — replace IN PLACE:
   `- ❌ \`develop\` 브랜치 생성 (Gitflow 패턴)`
   →
   `- ❌ \`main\`(또는 \`origin/main\`)에서 카드 브랜치 생성 — 카드 워크트리는 \`develop\`에서 분기한다`
3. **Line 351 bullet (§18.10)** — replace IN PLACE:
   `- ❌ \`develop\` 브랜치 생성 (Gitflow 패턴 금지)`
   →
   `- ❌ \`main\`에서 따서 카드 작업 시작 (분기점은 \`develop\`) · 카드 단위 PR은 존재하지 않는다 — 통합은 로컬 \`develop\` 병합(\`git merge --no-ff\`)이다`
4. **§18.3.1 routing rows** — keep the table but prepend a section-top addendum `> **[HARD] POLICY CHANGE (2026-08-27) — 카드 단위 PR 폐지.** 이 표의 PR routing은 릴리스 경로(`release/vX.Y.Z` → `main`)에만 적용된다…`, mark the four tier rows' card-applicability with `[RETIRED 2026-08-27]` annotations, and add a two-line current-rule statement (cards branch from `develop`, merge into `develop`, CI on `origin/develop` judges).

ACs flipped by M1: AC-GDA-001, -002, -003, -004.

### M2 — git-local-workflow-doctrine.md §23.7 / §23.9 re-scoping [Priority High]

1. **§23.7 lines 150–151 bullets** — TWO repairs in the SAME [HARD] list (the list must not end up contradicting itself):
   - **Line 150 (PR-mandatory bullet)** — annotate `~~[RETIRED 2026-08-27 — 카드 변경은 더 이상 PR을 만들지 않는다]~~` style strike-through AND replace with the scoped rule: main direct push prohibition unchanged; daily card changes integrate into local `develop` without a PR; the PR path now exists only for the release route. Preserve the self-merge conditions sentence as applying to release-path PRs.
   - **Line 151 (`git push origin main` 금지 bullet — plan-audit iter-1 D1)** — KEEP its true half verbatim (`시도 시 server-side rejected`); strike-and-annotate ONLY the route-prescription tail `항상 feat/fix/chore/docs/release 브랜치 → gh pr create → CI green → gh pr merge 흐름` behind ~~…~~ **[RETIRED 2026-08-27 — 카드 변경에는 더 이상 적용되지 않는다]**, appending: 릴리스 경로(`release/vX.Y.Z` → `main`, manager-git 위임)만 PR을 경유하며 카드 변경은 로컬 `develop` 병합으로 통합된다. Result: no standing always-PR clause survives two slots below the repaired L150.
2. **§23.9** — insert a `> **[RETIRED 2026-08-27]**` blockquote preserving the four-row tier table and the routing-flow items as historical record; write the current routing above it in ≤6 lines: (a) card/SPEC work → branch from `develop`, merge into `develop`, no PR; (b) `release/vX.Y.Z` → `main` → manager-git PR, merge-commit strategy, 0-approval self-merge once required checks pass; (c) explicit `--pr`: not part of card flow any more (canon: cards make no PRs — see research.md F-6).

ACs flipped by M2: AC-GDA-005, -006.

### M3 — repo-local-pr-policy.md rewrite [Priority Medium]

Full-body rewrite (14 lines) of a file whose entire premise inverted. Draft carried in plan so review sees the exact new contract:

```markdown
---
description: "Repo-local policy override — moai-adk-go runs git-flow (operator directive 2026-08-27): card work branches from develop and merges into local develop with NO card-level PR; main direct push remains prohibited"
paths: ".moai/specs/**,.claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/sync.md,.claude/skills/moai/workflows/moai.md"
---

# Repo-Local Branch Policy (moai-adk-go maintainer override)

[HARD] In THIS repository, direct push to `main` is DISABLED. `main` is protected with `enforce_admins: true` + required PR (verified via `gh api repos/modu-ai/moai-adk/branches/main/protection`), so a direct push to `main` is rejected even for admins. `main` advances ONLY through release pull requests (`release/vX.Y.Z` → `main`, merge-commit strategy; self-merge allowed at 0 required approvals once the required status checks pass).

[HARD] Card workflow (git-flow transition 2026-08-27; model: CLAUDE.local.md §4.1, rules: .claude/rules/local/gitflow-lane-protocol.md):
- Card worktrees branch FROM `develop` (never `main`).
- Completed cards integrate into LOCAL `develop` via `git merge --no-ff` inside the single integration worktree (`.claude/worktrees/develop`). There are NO card-level PRs. Remote CI on `origin/develop` is the verdict surface; lanes push `origin/develop`.
- The orchestrator MUST NOT instruct lane agents (`manager-develop` / `manager-docs` / per-spawn workers) to open card PRs against `main`, regardless of Tier.
- PR-based ceremony (spec-workflow Route B tier routing) applies ONLY to the release path above.

> ~~Prior policy (2026-07-20 → superseded 2026-08-26): ALL tiers (S / M / L) use Route B (PR): work lands on a feature branch and merges via PR~~ [RETIRED 2026-08-27 — replaced by the git-flow transition]

- This is a repo-local addendum (local-only, NOT mirrored to `internal/template/templates/`). The distributed template keeps the generic Route A / Route B choice for downstream users who have no branch protection.

Cross-reference: CLAUDE.local.md §4.1; `.claude/rules/local/gitflow-lane-protocol.md`; `.moai/docs/git-workflow-doctrine.md` §18 / `.moai/docs/git-local-workflow-doctrine.md` §23 (header notices carry the transition framing).
```

Note: `~~prior-policy~~` line keeps D7's old assertion behind strike-through for history while removing it as standing instruction (this file is a RULE, not a narrative doc — one strike-through line, not a preserved table).

AC flipped by M3: AC-GDA-007.

### M4 — verification battery + closure bookkeeping [Priority Low]

Mechanical: run the eight-command AC battery (§E1), capture outputs to `.moai/state/verify/<session>/t310-ac-battery.txt`, complete progress.md §E.1. Suggested batching: independent greps batched in ONE Bash turn (per agent-common-protocol parallel-execution discipline); each grep is read-only and idempotent.

Flips: AC-GDA-008 (+ confirms every earlier AC).

Delegation map (run-phase): single executor, NO sub-agents. Lane executes M1→M2→M3 directly (or delegates once to manager-docs-style doc editing); M4 is orchestrator/lane-direct. No manager-git involvement — card integrates into local `develop` per lane protocol §2–§4.

## §G Anti-Patterns

- Deleting historical passages wholesale (breaks the repo's annotation convention and loses the WHY).
- Leaving a strike-through ❌ develop bullet inside a forbidden LIST (list slots must state truths).
- Renumbering §18.x/§23.x headings ("while I'm in there" drift breaks external cross-references).
- Editing the header notices themselves (~7 lines added at transition time are canonical framing — leave byte-identical unless a fix REQUIRES touching them; prefer not).
- Mirroring anything to `internal/template/templates/**` (these files are deliberately local-only).
- Running a build/test suite to "verify" documentation (the battery IS the verification).

## §H Cross-References

- Model canon: `CLAUDE.local.md` §4.1 · `.claude/rules/local/gitflow-lane-protocol.md`
- Defect evidence + measured line numbers: research.md (pinned to d29b8942e)
- Precedent SPEC directories consulted: SPEC-CC-DOCS-ALIGNMENT-001, SPEC-V3R6-WORKFLOW-DOCS-001 (docs-only artifact sets, no design.md)
- Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map (this SPEC's progress.md carries the canonical §E.1–§E.4 skeleton; era.go matches the literal `§E.2..§E.4` tokens)
