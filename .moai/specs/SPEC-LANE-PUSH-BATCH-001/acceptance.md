---
id: SPEC-LANE-PUSH-BATCH-001
title: "Acceptance criteria — lane push prohibition; lead-side batched develop push"
version: "0.1.0"
created: 2026-09-02
author: manager-spec
---

# SPEC-LANE-PUSH-BATCH-001 — acceptance.md

## §1 AC Matrix

| AC | Requirement | Subject | Classification |
|----|-------------|---------|----------------|
| AC-LPB-001 | REQ-LPB-001/002 | No lane-facing `git push origin develop` in §4.1; lane chain ends at release → local merge-SHA report; exactly ONE sanctioned lead push site | release-blocking |
| AC-LPB-002 | REQ-LPB-003 | Lead batch-push procedure: 3 elements + measured evidence | release-blocking |
| AC-LPB-003 | REQ-LPB-004/005/006 | Lane protocol §2 exception, §4 rewrite, §6 wording, §7 duty | release-blocking |
| AC-LPB-004 | REQ-LPB-008 | Five kept clauses + §3/§5/§8-§11 preserved | regression-guard (preservation) |
| AC-LPB-005 | REQ-LPB-007/003 | Doctrine surfaces reworded (workflow-doctrine §18.3 + §18.8, local-workflow-doctrine §23.7); §18.8 push dropped; measured justification cited | release-blocking |
| AC-LPB-006 | REQ-LPB-004 | repo-local-pr-policy.md card-workflow bullet: lanes-push clause → lead-batch | release-blocking |

## §2 Given-When-Then

**AC-LPB-001** — **Given** the §4.1 lane window chain and 운영 절차 code block still
instruct `git push origin develop` and name no merge-SHA report, **when** M1 steps 5-6
land, **then** the chain-arrow form is gone (G1 = 0); the bare-line form survives EXACTLY
once — inside the lead batch-push block's runnable code (G2 = 1; the count is 1 at both RED
and GREEN — the lane-facing line 364 at RED, the lead block at GREEN — so the site flip is
carried by G1 + G3's whitelist + G28, with G2 bounding the ceiling at one); the [HARD]
2026-09-01 bullet survives verbatim; and the 창을 받으면 chain now ends at
`moai integration release` followed by the local merge-SHA report (G28 chain-scoped
positive anchor — the evasion mutant that rephrases the chain to end at release without
the SHA-report step fails it). Covers REQ-LPB-001 (lane duty ends at the local merge +
mandatory merge-SHA report) and REQ-LPB-002 (no lane-facing push instruction).

**AC-LPB-002** — **Given** §4.1 today names no push actor and carries no batch procedure,
**when** M1 step 6 lands, **then** a lead batch-push block exists carrying (1) collection
basis — lane completion reports carry card id + local merge SHA, (2) push-time decision —
the lead closes the batch and pushes once, (3) remote-landing verification — fetch +
rev-parse `origin/develop` before card done + disposal approval, and cites the measured
evidence sentence.

**AC-LPB-003** — **Given** protocol §2 delegates to delivery.md without exception, §4
instructs lanes to push with a rejected-push-retry scenario, §6 says "병합·push를 마치면",
and §7 has no batch duty, **when** M1 steps 1-4 land, **then** §2 carries the EXCLUDED
exception for delivery.md Step 3.2 step 6, §4 assigns the push to the lead (retry scenario
gone, verdict-surface sentence kept), §6 uses lane wording without push with the disposal
gate re-anchored to the lead's push landing, and §7 carries the lead batch duty.

**AC-LPB-004** — **Given** the five kept clauses and protocol §3/§5/§8-§11 are present at
baseline (green at arrival — see ledger P1-P5), **when** M1 completes, **then** all anchors
still measure exactly their baseline counts (the work disturbs nothing it must not).

**AC-LPB-005** — **Given** `git-workflow-doctrine.md` §18.8 step 1 includes
`git push origin develop`, its §18.3 merge-strategy callout (line 103) says the lane merges
then pushes, and `git-local-workflow-doctrine.md` §23.7 (line 150) says the same, **when**
M1 steps 6-8 land, **then** the §18.8 step drops the push string and references the lead
batch (`CLAUDE.local.md` §4.1); the §18.3 callout is reworded to lead-batch with the
`gitflow-lane-protocol.md` §2·§4 pointer kept; the §23.7 bullet is reworded to lead-batch
with its [RETIRED] markers and `enforce_admins` note kept; and the measured justification
(3 cards, 22 commits, `09bf452c0..ad272be20`, Vercel 3→1) is citable from the §4.1 lead
batch block.

**AC-LPB-006** — **Given** `repo-local-pr-policy.md` line 12 ends "Remote CI on
`origin/develop` is the verdict surface; lanes push `origin/develop`.", **when** M1 lands,
**then** the "; lanes push `origin/develop`" clause is replaced by lead-batch wording (the
lead collects lane merge SHAs and batch-pushes, 2026-09-02) and the verdict-surface
sentence survives.

## §3 GREEN-path verification commands

Run from the worktree root; batch as parallel read-only Bash calls. `grep -c` exit codes:
0 when count > 0, 1 when count = 0 — for the → 0 checks, stdout `0` + exit 1 IS the PASS.

| # | Command | Expected stdout | Expected exit |
|---|---------|-----------------|---------------|
| G1 | `grep -n "→ \\\`git push origin develop\\\`" CLAUDE.local.md` | (empty) | 1 |
| G2 | `grep -c "^git push origin develop$" CLAUDE.local.md` | `1` (exactly one — the lead batch-push block; count is 1 at RED too, the site flip is carried by G3) | 0 |
| G3 | `grep -n "git push origin develop" CLAUDE.local.md` | hits ONLY on the 2026-09-01 [HARD] bullet and inside the lead batch-push block | 0 |
| G4 | `grep -c "일괄" CLAUDE.local.md` | `>= 1` | 0 |
| G5 | `grep -c "병합 SHA" CLAUDE.local.md` | `>= 1` | 0 |
| G6 | `grep -c "원격 착지" CLAUDE.local.md` | `>= 1` | 0 |
| G7 | `grep -c "09bf452c0" CLAUDE.local.md` | `>= 1` | 0 |
| G8 | `grep -c "병합 후 레인이 직접 올린다" .claude/rules/local/gitflow-lane-protocol.md` | `0` | 1 |
| G9 | `grep -c "거부되면" .claude/rules/local/gitflow-lane-protocol.md` | `0` | 1 |
| G10 | `grep -c "병합·push" .claude/rules/local/gitflow-lane-protocol.md` | `0` | 1 |
| G11 | `grep -c "EXCLUDED" .claude/rules/local/gitflow-lane-protocol.md` | `>= 1` | 0 |
| G12 | `grep -c "일괄" .claude/rules/local/gitflow-lane-protocol.md` | `>= 1` | 0 |
| G13 | `grep -c "원격 CI" .claude/rules/local/gitflow-lane-protocol.md` | `>= 1` (verdict surface kept) | 0 |
| G14 | `grep -c "git push origin develop" .moai/docs/git-workflow-doctrine.md` | `0` | 1 |
| G15 | `grep -c "리드 일괄" .moai/docs/git-workflow-doctrine.md` | `>= 1` | 0 |
| G16 | `grep -c "승인이 아니다" CLAUDE.local.md` | `1` | 0 |
| G17 | `grep -c "sync는 병합" CLAUDE.local.md` | `1` | 0 |
| G18 | `grep -c "원격 머지가 확인되기 전까지 폐기하지 않는다" CLAUDE.local.md` | `1` | 0 |
| G19 | `grep -c "WT 브랜치 push·CI 직접 요청 금지" CLAUDE.local.md` | `1` | 0 |
| G20 | `grep -c "^## [0-9]" .claude/rules/local/gitflow-lane-protocol.md` | `11` | 0 |
| G21 | `grep -c "lanes push" .claude/rules/local/repo-local-pr-policy.md` | `0` | 1 |
| G22 | `grep -c "일괄" .claude/rules/local/repo-local-pr-policy.md` | `>= 1` | 0 |
| G23 | `grep -c "verdict surface" .claude/rules/local/repo-local-pr-policy.md` | `>= 1` (kept) | 0 |
| G24 | `grep -c "합친 뒤 push한다" .moai/docs/git-local-workflow-doctrine.md` | `0` | 1 |
| G25 | `grep -c "일괄" .moai/docs/git-local-workflow-doctrine.md` | `>= 1` | 0 |
| G26 | `grep -c "에 push한다" .moai/docs/git-workflow-doctrine.md` | `0` | 1 |
| G27 | `grep -c "gitflow-lane-protocol.md" .moai/docs/git-workflow-doctrine.md` | `>= 1` (pointer kept) | 0 |
| G28 | `grep -c "^- 창을 받으면: .*병합 SHA" CLAUDE.local.md` | `>= 1` (chain-scoped positive anchor — the chain ends at release → merge-SHA report) | 0 |

AC mapping: AC-1 = G1+G2+G3+G28 (covers REQ-LPB-001 + REQ-LPB-002); AC-2 = G4+G5+G6+G7 (+
human read of the evidence sentence); AC-3 = G8-G13; AC-4 = G16-G20; AC-5 =
G14+G15+G24+G25+G26+G27 (+ G7 for the §4.1 citation); AC-6 = G21+G22+G23.

## §4 Evidence ledger — RED-now cells

Baseline attribution: measured in worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t430`
(branch `WT-lead-batch-push`), **tree SHA `ad272be20abff9e4f3b1b363fce3e48dac4c5132`**,
2026-09-02, by manager-spec at plan-phase. Commands recorded repo-root-relative; executed
against the same files via absolute paths resolving to this worktree. `grep -c` prints the
count and exits 0 iff count > 0 — a `0` line with exit 1 is the expected shape of the
not-yet-present anchors. The exit codes annotated on the →0 anchor lines in R4/R8 are
derived from this documented `grep -c` semantics (count 0 ⇒ exit 1), not separately
captured `$?`; per-command `$?` is captured at run-phase when the §3 GREEN batch executes.

### R1 — AC-LPB-001 RED (release-blocking)

```text
$ grep -n "git push origin develop" CLAUDE.local.md
348:- 창을 받으면: `moai integration acquire --name <lane>` → 본인 워크트리에서 `git merge origin/develop` 흡수 → **병합 트리에서 재측정** → `EnterWorktree(.claude/worktrees/develop)` → `git merge --no-ff <WT-브랜치>` → `git push origin develop` → `moai integration release` → `ExitWorktree keep` → 완료 보고
349:- **[HARD] WT 브랜치 push·CI 직접 요청 금지 (운영자 지시 2026-09-01).** 카드가 마감되면 로컬 develop 병합(창 경유 `git push origin develop`)이 **유일한** 공개 경로다. 레인은 `git push origin <WT-브랜치>`를 하지 않고, `gh run rerun`/`workflow dispatch` 등 CI를 직접 요청·재요청하지도 않는다 — CI 판정은 develop push가 일으키는 실행에 맡기고, 판독은 리드 몫이다. (당일 lane-2가 `WT-version-stamp-predicate`를 origin에 push한 전례로 추가)
364:git push origin develop
exit code: 0   (3 hits; 348 + 364 are the removal targets, 349 is the preserved bullet — recorded in full, no elision)

$ grep -c "^- 창을 받으면: .*병합 SHA" CLAUDE.local.md
0
exit code: 1   (G28 chain-scoped positive anchor RED — the merge-SHA report step is absent from the chain today)
```

### R2 — AC-LPB-003 RED (release-blocking)

```text
$ grep -n "병합 후 레인이 직접 올린다" .claude/rules/local/gitflow-lane-protocol.md
69:병합 후 레인이 직접 올린다: `git push origin develop`.
exit code: 0
```

### R3 — AC-LPB-005 RED (release-blocking)

```text
$ grep -n "git push origin develop" .moai/docs/git-workflow-doctrine.md
329:# 1. 카드 워크트리(WT-<slug>)에서 수정 → 통합 워크트리에서 develop 에 --no-ff 병합 → git push origin develop
exit code: 0
```

### R4 — AC-LPB-002 / AC-LPB-003 replacement-anchor RED (all counts measured at tree ad272be20)

```text
$ grep -c "일괄" CLAUDE.local.md                 → 0  (exit 1)
$ grep -c "병합 SHA" CLAUDE.local.md            → 0  (exit 1)
$ grep -c "원격 착지" CLAUDE.local.md           → 0  (exit 1)
$ grep -c "09bf452c0" CLAUDE.local.md           → 0  (exit 1)
$ grep -c "ad272be20" CLAUDE.local.md           → 0  (exit 1)
$ grep -c "EXCLUDED" .claude/rules/local/gitflow-lane-protocol.md → 0  (exit 1)
$ grep -c "일괄" .claude/rules/local/gitflow-lane-protocol.md     → 0  (exit 1)
$ grep -c "리드 일괄" .moai/docs/git-workflow-doctrine.md          → 0  (exit 1)
```

### R5 — AC-LPB-003 removal-target RED (protocol retry + §6 wording, present today)

```text
$ grep -c "거부되면" .claude/rules/local/gitflow-lane-protocol.md  → 1  (exit 0; line 71)
$ grep -c "병합·push" .claude/rules/local/gitflow-lane-protocol.md → 1  (exit 0; line 81)
```

### R6 — AC-LPB-006 RED (release-blocking; amendment location)

```text
$ grep -n "lanes push" .claude/rules/local/repo-local-pr-policy.md
12:- Completed cards integrate into LOCAL `develop` via `git merge --no-ff` inside the single integration worktree (`.claude/worktrees/develop`). There are NO card-level PRs. Remote CI on `origin/develop` is the verdict surface; lanes push `origin/develop`.
exit code: 0   (kept-anchor baseline: "verdict surface" = 1, exit 0)
```

### R7 — AC-LPB-005 doctrine-location RED (release-blocking; amendment locations)

```text
$ grep -n "합친 뒤 push한다" .moai/docs/git-local-workflow-doctrine.md
150:- [HARD] **(2026-08-27 개정)** 카드 변경은 PR 없이 통합된다 — 카드 워크트리는 `develop`에서 분기하고, 통합 워크트리에서 로컬 `develop`에 `git merge --no-ff`로 합친 뒤 push한다(`origin/develop` CI가 판정). ~~PR-mandatory: 모든 tier (S/M/L) 변경은 PR 경유~~ **[RETIRED 2026-08-27 — 카드 변경은 더 이상 PR을 만들지 않는다]**. main direct push 금지는 변함 없다(`enforce_admins: true`, 아래 불릿); PR 경로는 릴리스 전용이다(`release/vX.Y.Z` → `main`, manager-git 위임). self-merge 조건(0 approvals + 필수 status check 통과)은 릴리스 경로 PR에 그대로 적용되며, tier는 그 ceremony 무게만 결정한다.
exit code: 0

$ grep -n "에 push한다" .moai/docs/git-workflow-doctrine.md
103:> **[HARD] 2026-08-27 개정.** 카드 통합(`WT-<slug>` → `develop`)은 이 표의 대상이 **아니다** — PR도 squash도 없고, 통합 워크트리 안에서 `git merge --no-ff`로 합친 뒤 `origin/develop`에 push한다(`gitflow-lane-protocol.md` §2·§4). 아래 표는 **`main`으로 가는 PR**에만 적용된다.
exit code: 0   (single hit; kept-anchor baseline: "gitflow-lane-protocol.md" = 5, exit 0)
```

### R8 — amendment replacement-anchor RED (all counts measured at tree ad272be20)

```text
$ grep -c "일괄" .claude/rules/local/repo-local-pr-policy.md  → 0  (exit 1)
$ grep -c "일괄" .moai/docs/git-local-workflow-doctrine.md    → 0  (exit 1)
```

### P1-P5 — AC-LPB-004 preservation baseline (green at arrival; regression-guard)

```text
$ grep -c "승인이 아니다" CLAUDE.local.md                            → 1  (exit 0)
$ grep -c "sync는 병합" CLAUDE.local.md                              → 1  (exit 0)
$ grep -c "원격 머지가 확인되기 전까지 폐기하지 않는다" CLAUDE.local.md → 1  (exit 0)
$ grep -c "WT 브랜치 push·CI 직접 요청 금지" CLAUDE.local.md          → 1  (exit 0)
$ grep -c "^## [0-9]" .claude/rules/local/gitflow-lane-protocol.md   → 11 (exit 0)
```

Preservation criteria are green at arrival by design (the work must not flip them); per the
two-cell discipline they are classified regression-guard, not release-blocking. Their
mutant probe: an implementation that rewrites the 2026-09-01 bullet or renumbers the
protocol sections while satisfying AC-1's shape greps FAILS G16-G20.

## §5 Quality gates / Definition of Done

- All release-blocking ACs (001, 002, 003, 005, 006) PASS with verbatim command output
  recorded in progress.md §E.2, attributed to the run's tree SHA.
- AC-LPB-004 (preservation) PASS at baseline counts.
- No file outside the five targets modified (`git status --short` shows exactly the five
  paths + the SPEC directory).
- No `internal/template/templates/` change; no `make build` run.
- Gaps: remote Vercel build-count reduction is NOT re-measured by this card's verification
  (it is the operator's observed justification, cited as evidence, not re-proven here);
  `git-local-workflow-doctrine.md` §23.9(a) (line 157) is confirmed-leave (push-neutral as
  written) and therefore unverified beyond that reading.
