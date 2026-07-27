---
id: SPEC-GOAL-SURFACE-UNIFY-001
title: Unify the goal surface on /moai goal and relocate goal presentation to the Implementation Kickoff Approval gate
version: 1.3.0
status: in-progress
created: 2026-07-25
updated: 2026-07-27
author: manager-spec
priority: HIGH
phase: "v3.1.0"
module: doctrine
lifecycle: spec-anchored
tags: "goal, doctrine, session-handoff, slash-command, template-mirror"
tier: L
---

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifact set complete (Tier L): `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md`.

- SPEC ID regex check executed as Bash; observed output `SPEC-ID-CHECK: PASS`.
- **30** acceptance criteria retained after the iteration-2 scope reduction; every judgment command executed against the worktree at `origin/main` = `e306e21a9` and its verbatim output recorded as the baseline in `acceptance.md` §B.
- All 30 ACs FAIL at the recorded baseline.
- Milestone ownership verified single-owner across all **37** run-phase paths (`plan.md` §F.1). No sync-phase paths remain.
- Scope spans two layers: doctrine (M1-M6) and Go emission paths (M7, `cycle_type: tdd`). The public-docs layer was split out to `SPEC-GOAL-DOCS-RETIRE-001`.

Five brief corrections surfaced and resolved during authoring, each verified by an executed command (see `research.md`):

| # | Correction | Evidence |
|---|---|---|
| 1 | Doctrine scope is 28 existing files, not 26 — the root `CLAUDE.md` pair was missed by a `.claude/`-scoped grep | `research.md` §A.3 |
| 2 | Four retention surfaces exist, not one — `internal/goal/evaluate.go`'s native-`/goal` yield invariant was mis-classified in the D4 brief as "implementation" | `research.md` §F.2 |
| 3 | The D4 "17 implementation references" figure is not reproducible (nearest pattern gives 35); no AC is built on it | `research.md` §F.3 |
| 4 | Sync-phase affected set is 13 files, not 9 — `advanced/self-evolving.md` (×4) is MoAI-surface primitive naming, not factual contrast | `research.md` §G.1 |
| 5 | `advanced/hooks-reference.md` exists in all four locales, not only en/ko — the gap is a content gap inside existing pages | `research.md` §G.3 |

Two D4 preconditions resolved in the safe direction: `PrimitiveGoal` occupancy measured **0** (hard rename safe, no back-compat needed), and the renderer confirmed to have **no test** (M7 must author the RED gate).

### Plan-audit iteration 1 — FAIL 0.71, remediated

`plan-audit.md` returned FAIL (0.71 against the Tier L 0.85 threshold), forced by MP-3. The auditor re-executed all 29 judgment commands and **every recorded baseline reproduced verbatim, zero divergence**; the defects were structural, not measurement errors. All six MUST-FIX and the material SHOULD-FIX are addressed:

| Finding | Resolution |
|---|---|
| D1 — `tags:` YAML sequence broke SPEC parsing on all six artifacts (MP-3) | Converted to the comma-separated string form. `moai spec lint` now reports **0 errors** on all six (`spec.md` → `✓ No findings`; the other five carry one grandfathered `MissingExclusions` warning each, since Out-of-Scope correctly lives only in `spec.md`) |
| D2 — `plan.md` D3 [HARD] said "two" retention surfaces vs four elsewhere | §A.2 is now the canonical **retention register** with **six** rows; D3 reads "six" and adds a per-AC retention test obligation. `research.md` corrected. `design.md:99`'s different "two surfaces" (the two files owning the W2 rule halves) deliberately left alone |
| D3 — M7 breaks `runner_template_test.go`, owned by no milestone | Added to M7 (now 5 Go files); REQ-GSU-028 + AC-GSU-030 added |
| D4 — AC-GSU-028's blanket `0` would destroy a retention surface | `autonomous-loops.md` reclassified **split**, `autonomous-workflow-strategy.md` **retain + note**; AC-GSU-028 rewritten as emission-markers-positive + 3 retention pins; AC-GSU-032 added. Sweep total 89 → **18 emission markers** |
| D5 — 8 unbackticked refs invisible to every AC | Union detector adopted; AC-GSU-002 rescoped, AC-GSU-031 added over the five no-retention owned files |
| D6 — `moai-meta-harness/SKILL.md` + mirror owned by no milestone | Added to M4 (9 local) and M6 (15 mirrors) |
| S1/S2/S3/S12 | Cross-ref counts, `version: 1.2.0`, canonical **50**-path total, `phase: "v3.1.0"` |
| S4 | B1 corrected: mirror parity is CI-enforced for **2 of 15** pairs, raising AC-GSU-019's importance |
| S5/S6 | AC-GSU-023 rescoped to locale-named subtests; the ordering overclaim dropped and replaced by AC-GSU-033 (RED output recorded in §E.2) |
| S7 | REQ↔AC traceability matrix added at `acceptance.md` §F — **28/28 REQs** cited |
| S9/S10/S11 | `.moai/specs/**` exclusion stated; M4 must re-derive (not transpose) the v2.1.139 availability condition; M1 relocates the comparison-table native row into the prohibition section |

Iteration-2 entry state was 33 ACs / 28 REQs / 6 retention surfaces / 50 paths.

### Plan-audit iteration 2 — FAIL 0.64, STOP; scope reduction executed

Iteration 2 (`plan-audit-2.md`) returned FAIL 0.64 — a regression from 0.71 — and emitted **STOP** rather than a fix list. All must-pass criteria now pass (MP-3 genuinely closed: `spec lint --strict spec.md` → `✓ No findings`, and the SPEC contributes zero findings to the repo-wide run), and across both iterations the auditor re-executed **62 judgment commands with zero divergence**. The regression was consistency and coverage in a scope that had outgrown single-pass remediation: iteration 1's remediation updated some surfaces and missed others, and the misses were themselves new defects.

The user chose the auditor's **scope-reduction** proposal over iteration 3. The split runs along the run/sync seam:

| | SPEC-A (this SPEC) | SPEC-B |
|---|---|---|
| ID | `SPEC-GOAL-SURFACE-UNIFY-001` | `SPEC-GOAL-DOCS-RETIRE-001` |
| Scope | doctrine + Go, run-phase, M1-M7 | public docs, sync-phase |
| Tier | L | M |
| REQs | 25 (001-024 + 028) | 8 |
| ACs | 30 (001-027 + 030, 031, 033) | 12 |
| Retention surfaces | 3 (doctrine ×2, Go ×1) | 4 (docs) |

Iteration-2 findings closed in SPEC-A:

| Finding | Resolution |
|---|---|
| N1 — `spec.md` REQ-GSU-004 bound retention to FOUR while four other artifacts said SIX | **Re-derived, not edited.** REQ-GSU-004 now binds by reference to the `plan.md` §A.2 register instead of restating a table, so the two cannot drift again. Count derived from post-split layer membership: 6 − 3 docs-layer surfaces = **3**. All five surfaces agree (verified in §E.1 evidence below) |
| N2 — AC-028's a3 detector was English-only (`en:2 ja:0 ko:0 zh:0`) | Moved to SPEC-B and re-anchored on locale-invariant literals; every SPEC-B detector now carries a per-locale baseline proving symmetry |
| N3 — `plan.md` asserted a guard that did not exist for `orchestration-mode-selection.md`'s 4 residue tokens | File added to AC-GSU-031's list (baseline `99` → `112`), retention-tested first |
| N4 — the `moai-meta-harness` mirror had no guard | Added to AC-GSU-019's parity loop (`same=13` → `same=14`) |
| N5 — `acceptance.md` contradicted itself about AC-023 | The overclaim was **removed**, not disclaimed; the edge case now points at AC-GSU-033 |
| S-new-1 | AC-GSU-004 and AC-GSU-018 strengthened with content assertions, closing two proxy-coverage matrix rows |
| S-new-2 | Both duplicated paragraph pairs deleted |
| S-new-4/5/6/7 | Stale counts reconciled against `plan.md` §F.1 as the arithmetic SSOT; HISTORY rows 1.2.0 and 1.3.0 added; `version: 1.3.0` |
| S-old-S10 | AC-GSU-031 gained a `grep -c '2.1.139' run.md` → `0` component |

Also decided rather than deferred (the auditor's flagged judgment call): `session-handoff.md:163`'s Paste-Time Activation Matrix classification statement keeps its classification but drops the `/goal` token, since the classification survives canonically in retention register row 2 (`native-invocation-model.md`). Recorded in `plan.md` M2.

_Awaiting plan-audit iteration 3 (SPEC-A) and a first audit for SPEC-B, then Implementation Kickoff Approval._

## §E.2 Run-phase Evidence

### M7 RED gate — recorded before the GREEN change (AC-GSU-033)

`internal/hook/handoff_inject_render.go` had no test at plan time (recorded in §E.1). M7 authored one, and the RED state below was observed **before** the four `goalPrefix` values were changed. The RED state is transient and leaves no other durable trace, so this block is its only record.

Command:

```
go test -run 'TestHandoffInjectRender' ./internal/hook/ -v 2>&1
```

Observed output (verbatim, exit 1):

```
=== RUN   TestHandoffInjectRender
=== RUN   TestHandoffInjectRender/ko
    handoff_inject_render_test.go:24: handoffLocaleStrings("ko").goalPrefix = "  • /goal ", want "  • /moai goal "
    handoff_inject_render_test.go:37: rendered context does not carry "  • /moai goal the target test suite is green"
        --- rendered ---
        [MoAI 자동 재개 — 이전 세션 핸드오프]
        아래는 이전 세션이 저장한 재개 컨텍스트입니다. 이 주입은 컨텍스트 전달일 뿐이며, 확장 추론 모드를 자동으로 활성화하지 않습니다. 필요하면 아래 안내 줄을 직접 입력하세요.

        복원 안내(수동 입력):
          • /goal the target test suite is green ← 자율 continuation을 복원하려면 이 줄을 입력하세요

        ── 저장된 재개 메시지(참고) ──
        resume body
    handoff_inject_render_test.go:40: rendered context still emits the retired native-goal token "  • /goal "
        --- rendered ---
        [MoAI 자동 재개 — 이전 세션 핸드오프]
        아래는 이전 세션이 저장한 재개 컨텍스트입니다. 이 주입은 컨텍스트 전달일 뿐이며, 확장 추론 모드를 자동으로 활성화하지 않습니다. 필요하면 아래 안내 줄을 직접 입력하세요.

        복원 안내(수동 입력):
          • /goal the target test suite is green ← 자율 continuation을 복원하려면 이 줄을 입력하세요

        ── 저장된 재개 메시지(참고) ──
        resume body
=== RUN   TestHandoffInjectRender/ja
    handoff_inject_render_test.go:24: handoffLocaleStrings("ja").goalPrefix = "  • /goal ", want "  • /moai goal "
    handoff_inject_render_test.go:37: rendered context does not carry "  • /moai goal the target test suite is green"
        --- rendered ---
        [MoAI 自動再開 — 前セッションのハンドオフ]
        以下は前セッションが保存した再開コンテキストです。この注入はコンテキストの受け渡しのみで、拡張推論モードを自動的に有効化しません。必要に応じて下記の案内行を手動で入力してください。

        復元案内（手動入力）:
          • /goal the target test suite is green ← 自律継続を復元するにはこの行を入力してください

        ── 保存された再開メッセージ（参考）──
        resume body
    handoff_inject_render_test.go:40: rendered context still emits the retired native-goal token "  • /goal "
        --- rendered ---
        [MoAI 自動再開 — 前セッションのハンドオフ]
        以下は前セッションが保存した再開コンテキストです。この注入はコンテキストの受け渡しのみで、拡張推論モードを自動的に有効化しません。必要に応じて下記の案内行を手動で入力してください。

        復元案内（手動入力）:
          • /goal the target test suite is green ← 自律継続を復元するにはこの行を入力してください

        ── 保存された再開メッセージ（参考）──
        resume body
=== RUN   TestHandoffInjectRender/zh
    handoff_inject_render_test.go:24: handoffLocaleStrings("zh").goalPrefix = "  • /goal ", want "  • /moai goal "
    handoff_inject_render_test.go:37: rendered context does not carry "  • /moai goal the target test suite is green"
        --- rendered ---
        [MoAI 自动恢复 — 上一会话交接]
        以下是上一会话保存的恢复上下文。此注入仅传递上下文，不会自动启用任何扩展推理模式。如有需要，请手动输入下方的指引行。

        恢复指引（手动输入）:
          • /goal the target test suite is green ← 输入此行以恢复自主延续

        ── 已保存的恢复消息（参考）──
        resume body
    handoff_inject_render_test.go:40: rendered context still emits the retired native-goal token "  • /goal "
        --- rendered ---
        [MoAI 自动恢复 — 上一会话交接]
        以下是上一会话保存的恢复上下文。此注入仅传递上下文，不会自动启用任何扩展推理模式。如有需要，请手动输入下方的指引行。

        恢复指引（手动输入）:
          • /goal the target test suite is green ← 输入此行以恢复自主延续

        ── 已保存的恢复消息（参考）──
        resume body
=== RUN   TestHandoffInjectRender/en
    handoff_inject_render_test.go:24: handoffLocaleStrings("en").goalPrefix = "  • /goal ", want "  • /moai goal "
    handoff_inject_render_test.go:37: rendered context does not carry "  • /moai goal the target test suite is green"
        --- rendered ---
        [MoAI auto-resume — previous session handoff]
        The following is the resume context saved by the previous session. This injection only delivers context; it does not automatically enable any extended-reasoning mode. Enter the guidance lines below manually if you need them.

        Restoration guidance (enter manually):
          • /goal the target test suite is green ← enter this line to restore autonomous continuation

        ── saved resume message (reference) ──
        resume body
    handoff_inject_render_test.go:40: rendered context still emits the retired native-goal token "  • /goal "
        --- rendered ---
        [MoAI auto-resume — previous session handoff]
        The following is the resume context saved by the previous session. This injection only delivers context; it does not automatically enable any extended-reasoning mode. Enter the guidance lines below manually if you need them.

        Restoration guidance (enter manually):
          • /goal the target test suite is green ← enter this line to restore autonomous continuation

        ── saved resume message (reference) ──
        resume body
--- FAIL: TestHandoffInjectRender (0.00s)
    --- FAIL: TestHandoffInjectRender/ko (0.00s)
    --- FAIL: TestHandoffInjectRender/ja (0.00s)
    --- FAIL: TestHandoffInjectRender/zh (0.00s)
    --- FAIL: TestHandoffInjectRender/en (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.411s
FAIL
```

All four locales failed on both halves of the assertion — the strings-table value and its reachability through `renderHandoffContext`. The RED therefore proves the test observes the renderer's actual emission, not merely a table entry.

### AC PASS/FAIL matrix (M1-M7)

Every judgment command in `acceptance.md` §B was executed **verbatim** — unmodified, from the worktree root, with the §A shell prelude — at the end of M7. The `Observed` column carries the actual output of that run, not a carry-over from the milestone that landed the change. Every row is therefore attributed to a measurement taken against this tree, per `verification-claim-integrity.md` §2.

| AC | M | Judgment command (`acceptance.md` §B, verbatim) | Baseline | Target | Observed | Status |
|---|---|---|---|---|---|---|
| AC-GSU-001 | M1 | `head -1 "$G" \| grep -c '/moai goal'` | `0` | `1` | `1` | PASS |
| AC-GSU-002 | M1 | `awk` union-form, prohibition section excluded | `69` | `0` | `0` | PASS |
| AC-GSU-003 | M1 | `grep -c '^## Native \`/goal\` Prohibition$'` | `0` | `1` | `1` | PASS |
| AC-GSU-004 | M1 | heading count; then invariant sentence in section | `0` / `0` | `1` / `>= 1` | `1` / `1` | PASS |
| AC-GSU-005 | M1 | `grep -cF 'moai goal --run' "$G"` | `0` | `>= 1` | `1` | PASS |
| AC-GSU-006 | M2 | `grep -oiF 'post-paste'` over `$SH` `$EX` `$OS` | `18`/`6`/`4` | `0`/`0`/`0` | `0`/`0`/`0` | PASS |
| AC-GSU-007 | M2 | `grep -ciF 'Post-paste /goal instruction line'` ×3 | `1`/`1`/`1` | `0`/`0`/`0` | `0`/`0`/`0` | PASS |
| AC-GSU-008 | M2 | stated/actual/retired self-check triple | `stated=10 actual=10 retired=1` | `stated=9 actual=9 retired=0` | `stated=9 actual=9 retired=0` | PASS |
| AC-GSU-009 | M2 | `grep -ciF 'arm-only' "$SH"` | `0` | `>= 1` | `3` | PASS |
| AC-GSU-010 | M2 | `grep -cF 'Progression Mode' "$SH"` | `0` | `>= 1` | `1` | PASS |
| AC-GSU-011 | M2 | `grep -o 'Pre-emit self-check ([0-9]* items)' "$OS"` | `(12 items)` | `(11 items)` | `Pre-emit self-check (11 items)` | PASS |
| AC-GSU-012 | M3 | `grep -cF 'Post-Paste' "$GW"`; `'Canonical Format'` | `1` / `0` | `0` / `>= 1` | `0` / `2` | PASS |
| AC-GSU-013 | M3 | `HUMAN-ONLY` occurrences; prohibition cross-ref | `17` / `0` | `>= 17` / `1` | `18` / `1` | PASS |
| AC-GSU-014 | M3 | `session-handoff` inside `$GW` § Progression Mode | `0` | `>= 1` | `1` | PASS |
| AC-GSU-015 | M4 | `` grep -ohF '`/goal' `` over the 8 emission-path files | `38` | `0` | `0` | PASS |
| AC-GSU-016 | M5 | `TestCommandsThinPattern` names `commands/moai/goal.md.tmpl` | `0` | `>= 1` | `2` | PASS |
| AC-GSU-017 | M5 | wrapper count, local / template | `14` / `14` | `15` / `15` | `15` / `15` | PASS |
| AC-GSU-018 | M5 | `arguments: goal`; `argument-hint` advertises `resume` | (file absent) | `1` / `0` | `1` / `0` | PASS |
| AC-GSU-019 | M6 | 14-pair byte parity + `post-paste` leak in templates | `same=14 diff=0 post_paste_in_templates=34` | `same=14 diff=0 post_paste_in_templates=0` | `same=14 diff=0 post_paste_in_templates=0` | PASS |
| AC-GSU-020 | M6 | stage-⑤ clause on both `CLAUDE.md` sides | `0` / `0` | `1` / `1` | `CLAUDE.md:1` / `templates/CLAUDE.md:1` | PASS |
| AC-GSU-021 | M6 | internal-content leak; `Goal-Presentation Timing` in mirror | `0` / `0` | `0` / `>= 1` | `0` / `2` | PASS |
| AC-GSU-022 | M7 | `grep -coF '• /goal '` / `'• /moai goal '` in the renderer | `4` / `0` | `0` / `4` | `0` / `4` | PASS |
| AC-GSU-023 | M7 | locale-named `--- PASS: TestHandoffInjectRender/(ko\|ja\|zh\|en)` | `0` | `4` | `4` | PASS |
| AC-GSU-024 | M7 | `PrimitiveGoal` aligned token; harness occupancy | `0` / `0` | `1` / `0` | `1` / `0` | PASS |
| AC-GSU-025 | M7 | `case "/goal":` / `case "/moai goal":` in runner template | `2` / `0` | `0` / `2` | `0` / `2` | PASS |
| AC-GSU-026 | M7 | `"goal", ""` flag name; `a /moai goal condition` help text | `1` / `0` | `1` / `>= 1` | `1` / `1` | PASS |
| AC-GSU-027 | M7 | `NativeGoalActive`; `native /goal`; quoted `"/goal"` non-test | `2` / `3` / `3` | `>= 2` / `>= 3` / `0` | `2` / `3` / `0` | PASS |
| AC-GSU-030 | M7 | `v4manifest` suite exit; fixture marker moved | `exit=0` / `0` | `exit=0` / `1` | `exit=0` / `1` | PASS |
| AC-GSU-031 | M4/M2 | union-detector sweep over the 5 no-retention files | `112` | `0` | `0` | PASS |
| AC-GSU-031b | M4 | `grep -c '2\.1\.139'` in `run.md` (graceful-degradation) | `1` | `0` | `0` | PASS |
| AC-GSU-033 | M7 | `FAIL: TestHandoffInjectRender` recorded in this §E.2 | `0` | `>= 1` | `5` | PASS |

**30 / 30 PASS · 0 FAIL.** AC-GSU-031b is the additional graceful-degradation component of AC-GSU-031, listed separately because it carries its own command and target; it is not a 31st criterion.

Retention surfaces re-verified rather than assumed — the three `plan.md` §A.2 register rows all survive: the `goal-directive.md` prohibition section (AC-GSU-003 = `1`), the `native-invocation-model.md` HUMAN-ONLY classification (AC-GSU-013 = `18`, above the `>= 17` floor), and the `internal/goal/evaluate.go` yield invariant (AC-GSU-027 = `2` / `3`). No sweep reached target by deleting retained content.

### M7 scope deviation — a sixth Go file, not named in the milestone brief

The M7 edit-site list named five Go files. A sixth was required and is recorded here rather than absorbed silently:

`internal/hook/handoff_inject_cover_test.go:165` asserted the retired token via `for _, want := range []string{"ultrathink", "/effort ultracode", "/goal ", ...}`. Because `/moai goal ` does not contain the substring `/goal `, the GREEN change turned `TestRenderHandoffContext_AllDirectives` red — caught by the full-suite gate, not by any AC. The single string literal was moved to `"/moai goal "`, preserving the test's intent (the goal guidance line renders) while re-pinning it to the surviving token.

This is the same lockstep shape the SPEC already owns explicitly for `runner_template_test.go` (added at plan-audit iteration 1 as D3 / REQ-GSU-028 / AC-GSU-030): an existing test pinning a token the SPEC retires must move with the implementation or the suite goes red. It is attributable to this SPEC's scope envelope per L46 — same renderer, same token retirement, `internal/hook/` is M7-owned — so it was fixed in place rather than returned as a blocker. A repo-wide sweep confirmed it was the only such pin: `grep -rnoF '"/goal"' --include='*.go' internal/ cmd/ pkg/` returns matches only in `internal/template/implementation_kickoff_approval_preservation_test.go` (template surface, M6-complete, its own suite green) and this file.

### gofmt gate — pre-existing baseline, zero M7 delta

`gofmt -l internal/` is **not** empty: it lists **108** files. This is a pre-existing repo-wide baseline, not an M7 regression. Measured per-file by extracting each file's `HEAD` version and running `gofmt -l` on both:

| File | gofmt @ HEAD `5b5204483` | gofmt @ now |
|---|---|---|
| `internal/harness/v4manifest/schema.go` | DIRTY | DIRTY |
| `internal/harness/v4manifest/runner_template.go` | clean | clean |
| `internal/harness/v4manifest/runner_template_test.go` | DIRTY | DIRTY |
| `internal/hook/handoff_inject_render.go` | clean | clean |
| `internal/hook/handoff_inject_cover_test.go` | DIRTY | DIRTY |
| `internal/cli/handoff.go` | clean | clean |
| `internal/hook/handoff_inject_render_test.go` (new) | (did not exist) | clean |

Every file's status is identical before and after M7; the new test file is clean. The three DIRTY files were already dirty at HEAD at lines unrelated to this SPEC — `schema.go`'s complaint is the `"e2e-tester":      TierOrange,` alignment in the `agentTiers` map at line ~192, a different const/map block from the `PrimitiveGoal` edit at line 15. Fixing it would be a drive-by touching another SPEC's surface, so it was left alone.

**AC-GSU-024 realignment risk — disproven by experiment.** The concern was that gofmt might realign the `PrimitiveGoal` const block and break AC-GSU-024's exact-alignment detector. Copying `schema.go`, running `gofmt -w` on the copy, and re-running the detector returns `1` — the block's alignment column is set by the longest *name* (`PrimitiveAdversarialFanOut`), not by any value, so substituting `"/goal"` → `"/moai goal"` changes no padding and a future `gofmt -w` would not either.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-27
run_commit_sha: pending-backfill-m7   # a commit cannot cite its own hash; backfilled in a follow-up
run_status: audit-ready
ac_pass_count: 30
ac_fail_count: 0
preserve_list_post_run_count: 3       # all three plan.md §A.2 retention rows survive (AC-003/013/027)
l44_pre_commit_fetch: "git fetch origin main → rev-list --count --left-right origin/main...HEAD = 2 9 (feature branch behind by 2, ahead by 9); the 2 origin/main commits touch 0 of the 8 files in this milestone's scope (comm -12 overlap = empty)"
l44_post_push_fetch: n/a              # not pushed — push is out of M7 scope per the milestone brief
new_warnings_or_lints_introduced: 0   # golangci-lint run --timeout=2m → "0 issues."; gofmt delta 0 (see §E.2)
cross_platform_build:
  host: "go build ./... → exit 0"
  windows: "GOOS=windows GOARCH=amd64 go build ./... → exit 0"
total_run_phase_files: 8              # 6 Go edited + 1 Go created + progress.md; spec.md frontmatter transition additionally
m1_to_mN_commit_strategy: "one commit per milestone, M1..M7; M7 is the final run-phase commit"
```

Held-out gates (`acceptance.md` §C), all re-run at M7 close:

| Gate | Observed |
|---|---|
| Full suite — `go test ./... 2>&1 \| tail -5` | green; zero `FAIL` lines across the run |
| M7 package suites | `ok internal/hook` · `ok internal/harness/v4manifest` · `ok internal/cli` · `ok internal/goal` |
| Cross-platform build | host exit 0; `GOOS=windows GOARCH=amd64` exit 0 |
| Template guard suite | `ok github.com/modu-ai/moai-adk/internal/template` |
| Lint | `golangci-lint run --timeout=2m` → `0 issues.` |
| gofmt | `gofmt -l internal/` → 108 files, **pre-existing**; M7 delta 0 (per-file HEAD-vs-now table in §E.2) |

Run-phase complete: M1-M7 landed, 30/30 AC PASS, all held-out gates green apart from the pre-existing gofmt baseline documented above. The one scope deviation (a sixth Go file) is recorded in §E.2. Not pushed — the branch carries the work locally for sync-phase.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
