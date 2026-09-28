# SPEC-LOCAL-INSTR-RECEPTION-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: pending-plan-audit
plan_complete_at: 2026-09-28
authored_by: card t1290 (manager-spec), worktree `.claude/worktrees/t1290`, HEAD `514ac7abe`
artifact_set: Tier L — spec.md, plan.md, acceptance.md, design.md, research.md (6 files with progress.md)
note: plan-audit has not run against this SPEC yet; no plan-artifact has been modified since authoring.

## §E.2 Run-phase Evidence

**M1 — the reception gate (2026-09-28 09:37–09:50 KST, caps per the §M1.4-style declaration above; GLM gateway per lead directive 2).** Raw verbatim outputs: `.moai/reports/t1290/m1-probes.md` (retained load-bearing artifact). Fixture token `TANGO9`, force-tracked per plan.md M1 step 2 (commit `6a6b32202` on WT-agents-local-migration).

| Leg | S1 file@HEAD | S2 exit 0 | S3 control (AGENTS.md loaded) | S4 local token | Verdict |
|---|---|---|---|---|---|
| RED (t1290 tree @ c05393bc2, file absent) | n/a — absent by design | ✓ | ✓ (answer describes AGENTS.md body) | `NOT_PRESENT` | **RED observed** (AC-LIR-005(a)) |
| A — materializer-created tree (`moai worktree new t1290probe-a`, base d5df9457c; fixture landed in-tree, `git show HEAD:AGENTS.local.md` exit 0) | ✓ | ✓ | ✓ | **`TANGO9`** | **PASS** (AC-LIR-001) |
| B — `claude -w t1290probe-b` native creation (ping turn `OK`; base 517ec51ba; pre-landing file@HEAD **exit 128** — the D1 geometry observed directly — fixture landed → S1 ✓) | ✓ | ✓ | ✓ | **`TANGO9`** | **PASS** (AC-LIR-002) |
| C — `EnterWorktree` re-entry into the leg-A tree | ✓ | ✓ | ✓ | **`TANGO9`** | **PASS** (AC-LIR-003) |
| D — `moai codex -w t1290probe-d -- exec` | gate-blocked | — | — | — | **측정 불가** (AC-LIR-004 open; attempts verbatim in m1-probes.md) |
| Negative control (git-archive export of `514ac7abe`, file-less by design; one cwd-error duplicate run recorded honestly before it) | absent-by-design (the opposite of S1) | ✓ | ✓ | `NOT_PRESENT` | **PASS** — probe fails for the right reason (AC-LIR-005(b)) |

**Central measurement result:** all three Claude-side entry paths receive the force-tracked `AGENTS.local.md` content (`TANGO9` observed, control present, exit 0, file at HEAD) — the first measured confirmation that with the file tracked and present at a worktree root the `@AGENTS.local.md` import fires in a linked worktree (`AGENTS.md:262`'s negative holds for the untracked geometry only). The negative control prints `NOT_PRESENT` from the file-less geometry — the probe is failable and discriminates.

**Leg D 측정 불가 (blocker chain, in firing order):** (1) the codex launcher's wiring gate refuses the launch while the resolved original project root (primary checkout — shared surface) lacks `.codex/hooks.json`; completing it requires `moai init` on the primary mid-measurement — outside lane authority (two invocation attempts + the non-interactive refusal recorded verbatim; `t1290probe-d` never created). (2) Behind the gate, untested today: codex account quota (t1203 record: blocked until 2026-09-28 14:37) and the login token-refresh pending state (t1280 record). No pass claimed for AC-LIR-004 — §8's claim stays a hypothesis until probed.

**M2 entry gate: STOPPED.** AC-LIR-001~005 are not all PASS (004 open) — the [HARD] entry condition fails and lead directive 3 applies: stop here, report. Caps reconciled 7/7 turns, ~13 min wall-clock.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §M1.4-style 측정 상한 선언 (실행 전 — 2026-09-28 09:37 KST, 리드 지침 1)

| 축 | 상한 |
|---|---|
| 프로브 턴 | 총 **7회**: RED 1 + leg 4(각 1) + leg B 트리 생성 핑 1 + 부정 대조 1. 재시도 없음(원문 기록) |
| 프로브별 시간 | `timeout 180` (plan.md 레시피 고정, /opt/homebrew/bin/timeout) |
| M1 전체 벽시계 | 30분 |
| 모델 경로 | **GLM 게이트웨이**(리드 지침 2 — Anthropic 쿼터 소진). 수신 측정 대상은 하네스의 지시파일 로딩이라 모델 백엔드 독립 |
| 레시피 이탈 | 레시피의 `--model claude-haiku-4-5-20251001` 핀이 게이트웨이에서 거부되면 핀 없이 재실행 — 두 시도 모두 원문 기록, 실패 파라미터 호출은 턴 수 미소모 |
| 쿼터/자격 불가 leg | 나머지 완료 후 그 leg 만 「측정 불가·사유 명시」 기록 → M2 진입 게이트에서 정지·리드 보고 (리드 지침 3, AC-LIR-006 준수) |

킥오프 승인 근거: 운영자 09-28 전력 완수 지시 + CLAUDE.local.md §31 (리드 메시지로 수령, iter2 PASS 1.00 판정서 .moai/reports/t1290/plan-audit-iter2.md).
