---
id: SPEC-LANE-PUSH-BATCH-001
title: "Lane push prohibition — develop integration ends at the local merge, lead-side batched push"
version: "0.1.0"
status: draft
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
priority: P1
phase: "v3.1.5"
module: "CLAUDE.local.md; .claude/rules/local/gitflow-lane-protocol.md; .claude/rules/local/repo-local-pr-policy.md; .moai/docs/git-workflow-doctrine.md; .moai/docs/git-local-workflow-doctrine.md"
lifecycle: spec-anchored
tags: "gitflow, develop, lanes, batch-push, vercel, docs"
era: V3R6
tier: S
---

# SPEC-LANE-PUSH-BATCH-001 — Lane push prohibition; lead-side batched develop push

## §1 Background & Problem

Operator directive (2026-09-02, kanban card t430): **lanes no longer push the `develop`
branch.** A lane's integration ends at the LOCAL merge (`git merge --no-ff` into the
integration worktree `.claude/worktrees/develop`, inside the `moai integration
acquire`/`release` window); the lane then reports its local merge SHA to the lead. The lead
collects lane merge SHAs from completion reports and performs ONE batched
`git push origin develop`.

Measured justification (orchestrator measurement, 2026-09-02): 3 cards (t336, t372, t413),
22 commits, absorbed in ONE push (`09bf452c0..ad272be20`) — Vercel builds went from 3 to 1.
Every push to `origin/develop` triggers a Vercel build, so per-lane pushes multiply deploy
cost and noise for zero integration value.

The doctrine documents still instruct the old world. Measured at tree `ad272be20` — the
enumeration closed at **8 point-locations across 5 files**:

- `CLAUDE.local.md` §4.1 carries lane-facing `git push origin develop` instructions in the
  창을 받으면 window chain (line 348) and the 운영 절차 code block (line 364); 규율 2
  (line 327) names no push actor; 규율 4 (line 329) places push inside the lane window.
- `.claude/rules/local/gitflow-lane-protocol.md` §4 (line 69) instructs lanes to push
  directly ("병합 후 레인이 직접 올린다") and carries a lane-side rejected-push-retry
  scenario (line 71); §2 delegates the merge procedure to the distributed
  `delivery.md` Step 3.2 — whose step 6 IS a lane-facing integration-branch push (line 278
  of that file) — with no repo-local exception; §6 (line 81) and §7 predate the batch model.
- `.moai/docs/git-workflow-doctrine.md` §18.8 Patch Release step 1 (line 329) includes
  `git push origin develop` in the release-preparation chain, and the §18.3 merge-strategy
  callout (line 103) says the lane merges then pushes ("합친 뒤 `origin/develop`에
  push한다").
- `.claude/rules/local/repo-local-pr-policy.md` card-workflow bullet (line 12) ends
  "Remote CI on `origin/develop` is the verdict surface; lanes push `origin/develop`."
- `.moai/docs/git-local-workflow-doctrine.md` §23.7 [HARD] bullet (line 150) says the lane
  merges then pushes ("합친 뒤 push한다(`origin/develop` CI가 판정)").

Enumeration basis (orchestrator wide sweep 2026-09-02, corroborated at plan-phase):
sweep pattern = integration merge co-mentioned with push across `.moai/docs/`,
`.claude/rules/local/`, and `CLAUDE.local.md` — zero further hits; a plan-phase
`grep -rln` over the push-phrase class (`합친 뒤 push`, `에 push한다`, `lanes push`,
`git push origin develop`) returns exactly the 5 in-scope files. `delivery.md` 278/299 and
the primary working tree stay Out of Scope (§5).

Adjudicated ground truth this SPEC inherits (orchestrator adjudication — do NOT re-litigate):
canonical copy = `origin/develop @ ad272be20`; the 2026-08-29 chain + 2026-09-01 prohibition
landed on develop via `9a161687a`; the primary working-tree `CLAUDE.local.md` copy is a stale
derivative of develop's copy with ZERO unique lines (full diff measured, 35 primary-only
lines all older variants). All edits target the develop-base copies in the worktree.

## §2 Scope

**In scope (doc targets — run-phase writes; this SPEC specifies WHAT, not HOW):**

- `CLAUDE.local.md` §4.1 — lane window chain, 운영 절차 code block, 규율 2, 규율 4, NEW lead
  batch-push procedure (REQ-LPB-001, 002, 003).
- `.claude/rules/local/gitflow-lane-protocol.md` — §2 repo-local exception, §4 rewrite, §6
  rewording, §7 lead duty (REQ-LPB-004, 005, 006); §3/§5/§8/§9-§11 preserved (REQ-LPB-008).
- `.moai/docs/git-workflow-doctrine.md` — §18.8 step 1 push removal + lead-batch reference,
  §18.3 merge-strategy callout (line 103) lead-batch rewording (REQ-LPB-007).
- `.claude/rules/local/repo-local-pr-policy.md` — card-workflow bullet (line 12)
  lanes-push clause → lead-batch wording (REQ-LPB-004).
- `.moai/docs/git-local-workflow-doctrine.md` — §23.7 [HARD] bullet (line 150) lead-batch
  rewording (REQ-LPB-007).

Verification is grep-based (docs-only); no Go code, no tests, no template mirrors (all five
target files verified absent from `internal/template/templates/` at tree `ad272be20`).

## §3 Requirements (GEARS)

**REQ-LPB-001** (Event-driven) — **When** a lane's develop integration window closes
(`moai integration release`), the lane's public-facing duty shall end at the LOCAL merge:
the lane reports the local merge SHA (with card id) to the lead and performs no
`git push origin develop`. The lane's completion report shall carry the local merge SHA as a
new mandatory field alongside the existing card id · branch/HEAD · unpushed-commit-count
fields.

**REQ-LPB-002** (Ubiquitous) — `CLAUDE.local.md` §4.1 shall contain no lane-facing
`git push origin develop` instruction: the 창을 받으면 chain bullet shall end at
`moai integration release` followed by reporting the local merge SHA to the lead; the 운영
절차 code block shall end at `moai integration release` with a comment that push happens
OUTSIDE the window, by the lead, batched; 규율 2 shall name the lead as the push actor
(batched, 2026-09-02); 규율 4's window chain shall be `acquire → merge → release` with push
excluded from the lane window. The literal string `git push origin develop` shall survive in
§4.1 only (i) inside the [HARD] 2026-09-01 WT-branch prohibition bullet (kept verbatim per
operator instruction — its parenthetical is the historical issuance context) and (ii) inside
the lead batch-push context of REQ-LPB-003.

**REQ-LPB-003** (Ubiquitous) — `CLAUDE.local.md` §4.1 shall carry a lead batch-push
procedure containing all three elements: (1) **collection basis** — lane completion reports
carry card id + local merge SHA; (2) **push-time decision** — the lead closes the batch and
pushes once; (3) **remote-landing verification** — after push, verify `origin/develop`
moved (fetch + rev-parse), then card done + worktree disposal approval; and shall cite the
measured justification (3 cards t336 + t372 + t413, 22 commits, one push
`09bf452c0..ad272be20`, Vercel builds 3→1).

**REQ-LPB-004** (Capability gate) — **Where** a lane follows the distributed
`delivery.md` Step 3.2 `git-flow` `WT-*` integration path, `.claude/rules/local/gitflow-lane-protocol.md`
§2 shall carry an explicit repo-local exception stating that delivery.md Step 3.2 step 6
(integration-branch push) is **EXCLUDED** in this repository — the push is the lead's
batched act (2026-09-02). The distributed skill stays untouched; the local answer is
authoritative for lanes. The same repo-local-authoritative rule extends to
`.claude/rules/local/repo-local-pr-policy.md`: its card-workflow bullet (line 12) shall
replace the "; lanes push `origin/develop`" clause with lead-batch wording (the lead
collects lane merge SHAs and batch-pushes, 2026-09-02) while keeping the "Remote CI on
`origin/develop` is the verdict surface" sentence intact.

**REQ-LPB-005** (Ubiquitous) — `.claude/rules/local/gitflow-lane-protocol.md` §4 shall not
instruct lanes to push: the lane-side wording ("병합 후 레인이 직접 올린다") and the
lane-side rejected-push-retry scenario ("거부되면 다른 레인이 먼저 올린 것이다") are
removed, replaced by the lead-side batch push + remote-landing verification. The statement
that remote CI on `origin/develop` is the integration verdict surface is RETAINED — the
verdict surface is unchanged; only the push actor and cadence change. §6 shall use lane
wording without push (the lane's exit duty ends at the merge + report), and its
card-worktree disposal gate remains "after the work has landed on `origin/develop`", now
gated on the lead's push landing.

**REQ-LPB-006** (Ubiquitous) — `.claude/rules/local/gitflow-lane-protocol.md` §7 shall carry
the lead's batch-push duty: collect merge SHAs from lane completion reports → batch
decision → single push → verify remote landing → card done + disposal approval.

**REQ-LPB-007** (Ubiquitous) — The doctrine documents shall carry the lead-batch wording at
all three doctrine locations. `.moai/docs/git-workflow-doctrine.md` §18.8 Patch Release
step 1 shall drop `git push origin develop` and reference the lead batch (`CLAUDE.local.md`
§4.1) in its place; the §18.3 merge-strategy callout (line 103) shall reword "합친 뒤
`origin/develop`에 push한다" to lead-batch wording while keeping the
`gitflow-lane-protocol.md` §2·§4 pointer; `.moai/docs/git-local-workflow-doctrine.md` §23.7
(line 150) shall reword "합친 뒤 push한다(`origin/develop` CI가 판정)" to lead-batch
wording while keeping its [RETIRED 2026-08-27] markers and the `enforce_admins` note
untouched.

**REQ-LPB-008** (Ubiquitous, preservation) — The following clauses shall remain present,
byte-identical where so marked: (1) the [HARD] WT 브랜치 push·CI 직접 요청 금지 bullet
(운영자 지시 2026-09-01, `CLAUDE.local.md` ~line 349 — verbatim); (2) the window-only-after-
lead-designation rule ("`moai integration status`가 `free`인 것은 **승인이 아니다**");
(3) the sync-before-merge rule ("sync는 병합 **전에** 워크트리 안에서 끝낸다"); (4) the
no-disposal-before-remote-landing rule ("워크트리는 원격 머지가 확인되기 전까지 폐기하지
않는다"); (5) `gitflow-lane-protocol.md` §3 (serialization), §5 (conflicts belong to the
integrating lane), §8 (lane-local verification), §9-§11 (rc build / release / develop
refresh) — all eleven numbered sections intact.

## §4 Constraints

- **No template mirrors** (verified at tree `ad272be20`): none of the five target files
  exists under `internal/template/templates/` — no `make build` is needed or permitted.
- **Docs-only**: zero Go code, zero workflow/hook edits, zero test changes. Verification is
  grep-based.
- **Language**: SPEC body English; Korean operator-directive phrases quoted in this SPEC and
  in the prescribed grep anchors stay verbatim Korean (they are the AC evidence anchors and
  match the target files' existing ko/en mixed convention).
- **Adjudication inherited**: the canonical-copy question is settled (`origin/develop @
  ad272be20`); run-phase edits the worktree copies and does not re-adjudicate.
- **The 2026-09-01 [HARD] bullet is kept verbatim** — including its now-historical
  parenthetical "(창 경유 `git push origin develop`)". It describes the rule at its issuance
  date; the lead batch-push section supersedes the mechanism it names.

## §5 Out of Scope

### Out of Scope — distributed skill delivery.md

- `.claude/skills/moai/workflows/sync/delivery.md` lines 278 / 299 (the two lane-facing
  `git push origin develop` instructions) are NOT edited. It is a distributed skill; the
  deployed copy is t303-era stale vs template (t335 parameterized the branch to
  `<integration-branch>`; self-heals at the next `moai update`). The File 2 §2 repo-local
  exception (REQ-LPB-004) makes the local answer authoritative for lanes without touching
  the distributed file.

### Out of Scope — primary checkout working tree

- The primary checkout's uncommitted stale `CLAUDE.local.md` copy (660 lines vs develop's
  710, zero unique lines — full diff measured by the orchestrator) is NOT repaired here.
  Operator reconciliation post-landing.

### Out of Scope — push tooling and automation

- No new `moai` CLI subcommand, script, or hook for batch collection or the batched push.
  The lead batch push is a doctrine act (manual, judgment-bearing), not tooling.

### Out of Scope — Go code, CI workflows, hooks

- Zero code changes; no `.github/workflows/` edits; no `internal/` or `cmd/` surface.

## §6 HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-09-02 | manager-spec | Plan-phase creation (card t430, Tier S). All premises verified at tree `ad272be20abff9e4f3b1b363fce3e48dac4c5132` (branch `WT-lead-batch-push`): RED-now greps R1-R3 observed, 10 replacement/preservation anchor counts measured, delivery.md Step 3.2 step 6 confirmed at line 278, template-mirror absence confirmed for all three targets. Sibling surfaces (repo-local-pr-policy.md, git-local-workflow-doctrine.md:150) reported in §5, out of scope. |
| 2026-09-02 | manager-spec | Pinned `era: V3R6` (H-override): `moai spec audit` (running server v3.1.2, commit 64bba61aa — predates the tree's SPEC-ERA-H3-NARROWING-001 H-3 deferral at internal/spec/era.go:161) classified this SPEC V3R5 via H-3 despite `created: 2026-09-02` ≥ `modernEraThreshold`. The pin makes classification deterministic across server versions; per the era-field SSOT § When to set era explicitly (newly created plan-phase SPECs). |
| 2026-09-02 | manager-spec | Scope amendment (orchestrator wide sweep closed the enumeration): +3 point-locations — `repo-local-pr-policy.md:12`, `git-local-workflow-doctrine.md:150`, `git-workflow-doctrine.md:103`. REQ-LPB-004 and REQ-LPB-007 extended (no new REQs — Tier S 8-REQ ceiling kept); AC-LPB-006 added, AC-LPB-005 extended; enumeration basis (8 points / 5 files, zero further hits) recorded in §1. All three cited locations + both replacement-anchor baselines verified in the worktree at tree `ad272be20`; template-mirror absence re-verified for the two new files. |
| 2026-09-02 | manager-spec | Plan-audit iter-1 delta fixes (PASS-WITH-DEBT 0.92, D1-D4; artifacts-only, zero target-file edits): D1 — AC-LPB-001 now covers REQ-LPB-001 (matrix cell `REQ-LPB-001/002` + GWT + mapping). D2 — R1 records line 349 in full verbatim (no mid-line elision). D3 (orchestrator direction override of the auditor's formatting ban) — G2 expectation 0 → EXACTLY 1: the lead batch-push block's runnable code line is the single sanctioned bare-line site; the count is 1 at both RED (lane-facing line 364) and GREEN, so the flip is carried by G1 + G3's whitelist + G28, with G2 bounding the ceiling. D4 — chain-scoped positive anchor G28 (`^- 창을 받으면: .*병합 SHA` ≥ 1, RED measured 0 at tree `ad272be20`) added so the release-without-SHA-report evasion mutant fails. Ledger preamble documents that →0 exit codes are semantics-derived (count 0 ⇒ exit 1), with per-command `$?` captured at run-phase. |
