# SPEC-LSEL-DRAIN-STALL-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-25
spec_id: SPEC-LSEL-DRAIN-STALL-001
tier: M
version: 0.2.0
artifact_set: [spec.md, plan.md, acceptance.md, progress.md]
baseline_sha: a739d04b4
worktree: .claude/worktrees/t259
branch: WT-lsel-drain-stall
plan_audit_iter1: "PASS-WITH-DEBT 0.92 (Tier M threshold 0.80) — MUST D1/D2/D6 + SHOULD D3/D4/D5 applied (v0.2.0, 2026-08-25)"
red_now_observations:
  wrapper_absent: "test -f .claude/skills/hns-lsel-curator/session_drain.sh → rc=1 (2026-08-25, worktree a739d04b4)"
  live_offset_gap: "wc -l → 4204행 vs drain-offset.json offset 629 (2026-08-25, primary 체크아웃 live state — RED for AC-LDS-010 predicate)"
  offset_frozen_since: "2026-08-04T05:21:03Z (21일) — clusters.json 같은 날짜"
  sessionstart_wiring_absent: "settings.json SessionStart 항목 = 1(session-attribution); settings.json+settings.local.json hooks에 lsel/backlog grep → 0 matches"
  loop_recipe_false_claim: "grep -c 'The recipe runs drain.sh' lsel-drain-loop.js → 1 (12행) vs 본문 45-49행 console.log뿐"
  dead_anchor: "grep -rn '§28' .claude/skills/hns-lsel-curator/ → 2건 (backlog_check.sh:6, :50 — iter-1 D1 정정, 종전 1건은 :50만 집계한 과소계상); live CLAUDE.local.md(35,355B)에 §28/LSEL 0회"
  backlog_composition_dry_run: "3,563행 delta — noise 3,095(86.9%), singleton 6, 후보 ≈13-15, top tool_failure:Read:UnknownFailure(146) — drain.sh 건재 rc=0"
id_precheck: "SPEC-LSEL-DRAIN-STALL-001 → bash regex → PASS (2026-08-25)"
```

계획 산출 4건 완료(v0.1.0, 2026-08-25). 카드 t259 (Class B). 원인 확정: **트리거 부재** — `/loop` 스케줄이 session-scoped로 소멸 + 레시피 본문이 실행 없는 console.log 묶음(헤더 주장과 불일치). Tier M 확정(AC 12 > Tier S 상한 8; tracked-코드/로컬-적용 이중 검증 표면). Milestone 2개: M1 wrapper+신호+위생, M2 일괄 드레인+검증(mutant guard)+로컬 인도물 적용. **iter-1 수정 적용(v0.2.0, 2026-08-25)**: plan-audit review-1 PASS-WITH-DEBT 0.92 — MUST D1(§28 관측 2건 정정)·D2(AC-010 파라미터화+archived 사본 판정)·D6(origin/main 결속) + SHOULD D3(전 드레인 wrapper 경유+PROPOSE archived 판독)·D4(advisory 발화 조건 완화)·D5(명시적 timeout 30+예산 수치) 전량 반영. 재감사 대기.

## §F Phase 4 Mode Selection

**M1 (내구 트리거 wrapper + 정지 신호 + 위생) — Phase 4 decision (logged before the M1 manager-develop spawn):**

Input parameters:
- tier: M (12 AC > Tier S ceiling 8)
- scope (file count): ~5 tracked (NEW `session_drain.sh`, NEW `session_drain_test.sh`, EDIT `lsel-drain-loop.js`, EDIT `backlog_check.sh`, EDIT `SKILL.md`) + SPEC artifacts
- domain count: 1 (LSEL drain tooling — bash scripts + skill docs, single cohesive mechanism)
- file language mix: bash + markdown; NO Go source changes
- concurrency benefit: LOW (coding-heavy, single-domain, M1→M2 sequential dependency — M2 consumes M1's wrapper + predicate)

Mode evaluation:
- Mode `direct`: NO — multi-file TDD implementation (wrapper + mutant-probe test), not a typo/single-line edit
- Mode `fanout`: NO — single domain, coding-heavy (Anthropic coding-task parallelism caveat)
- Mode `sweep`: NO — <30 files, semantic new-code (not a mechanical uniform transform); also unnecessary
- Mode `agent-team`: NO — not operator-requested (`--team` absent; kickoff selected 반자율 semi-autonomous progression, not the teams layer)
- Mode `serial`: YES

Decision: serial
Justification: M1 is coding-heavy single-domain implementation (flock wrapper + 5-path characterization test + mutant probe + 3 hygiene edits) with a hard M1→M2 dependency (M2's bulk drain routes through M1's wrapper and its verification predicate). Sequential manager-develop delegation (cycle_type=tdd, RED-first for the wrapper test) is the correct default per the coding-task parallelism caveat. Implementation Kickoff Approval passed 2026-08-25 (승인 / 일괄 드레인 / D8 비준 SessionStart 대체 / 반자율) — no goal armed (semi-autonomous milestone reporting selected).

Plan-audit gate note: iter-2 verdict PASS 0.95 ≥ Tier M threshold 0.80 on artifacts @ `70c9f7393` (hash subject = spec/plan/acceptance; this §F append touches progress.md only, outside the ComputeHash subject set) — skip-eligible, Phase 1 re-execution not required.

## §E.2 Run-phase Evidence

### M1 — 내구 트리거 wrapper + 정지 신호 + 위생 (2026-08-26, worktree WT-lsel-drain-stall, pre-commit HEAD `70c9f7393` + M1 tree)

**RED (AC-LDS-005 two-cell — wrapper 구현 직전, absent 상태 관측):**

```
FAIL: session_drain.sh not found or not executable at /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t259/.claude/skills/hns-lsel-curator/session_drain.sh (expected during RED phase; this line is the RED failure)
RED rc=1
```

**AC binary matrix (명령 → 관측 출력):**

| AC | 판정 | 명령 | 관측 출력 |
|---|---|---|---|
| AC-LDS-001 | PASS | `bash session_drain_test.sh` (path1) | `ok - path1 normal drain: offset 0->18, 3 candidates, one-line status emitted` |
| AC-LDS-002 | PASS | path2 | `ok - path2 lock contention: skip + exit 0 + notice; drain/archive NOT executed` |
| AC-LDS-003 | PASS | path3 | `ok - path3 archive-before-overwrite: prior candidate preserved to clusters-history, then overwritten` |
| AC-LDS-004 | PASS | path4 | `ok - path4 no-op: exit 0, offset unchanged, no-op status; bulk result archived before the wipe` |
| AC-LDS-005 | PASS | 전체 + mutant probe | `ALL SESSION DRAIN TESTS PASSED` (rc=0); `ok - mutant probe: predicate accepts the real bulk drain (true) and REJECTS the offset-only-advance mutant (false)` |
| AC-LDS-006 | PASS | path5 | `ok - path5 fail-open: inbox-absent and uncreatable state dir both exit 0 with a stderr notice` |
| AC-LDS-007 | PASS | `bash backlog_check_test.sh` | 3× PASS + `backlog_check_test: PASS` (rc=0) — 리마인더 텍스트 wrapper 지시로 교체 반영 |
| AC-LDS-008 | PASS | greps | `grep -c "The recipe runs drain.sh" .claude/workflows/lsel-drain-loop.js` → `0`; `grep -c "session_drain"` → `4` |
| AC-LDS-009 | PASS | greps | `grep -rn '§28' .claude/skills/hns-lsel-curator/` → 0 matches (rc=1); `grep -c 'Durable operations' SKILL.md` → `1` |
| AC-LDS-012(문서화 부) | PASS | 판독 | spec.md §E 3종 로컬 인도물이 "PR 미탑재·적용 단계 명시"로 기록(plan-phase 완료분 유지); SKILL.md 내구-운영 섹션에 M2 적용 안내 미러 |

**회귀 (E3):** `bash drain_test.sh` → `ALL DRAIN TESTS PASSED` (rc=0) · `bash backlog_check_test.sh` → `backlog_check_test: PASS` (rc=0). 구문: `bash -n` wrapper/backlog_check OK · `node --check` lsel-drain-loop.js OK.

**spec-lint:** `moai spec lint .moai/specs/SPEC-LSEL-DRAIN-STALL-001/spec.md` → `✓ No findings — all SPEC documents are valid` (rc=0, status=in-progress 상태).

**잠금 수단 결정 기록:** homebrew flock 0.4.0의 fd-form 지원을 실측했다(`flock -n 9` rc=0 — `.moai/reports/t259/flock_fd_probe.sh`, 카드 증거)에도 **mkdir 원자잠금**을 선택했다 — 훅 PATH에서 flock 바이너리 부재 시 `command not found`가 영구 경쟁으로 오독돼 모든 드레인이 조용히 skip되는 형태(=무음 정지 재현)가 결정적 위험. 120s stale-reap(`rmdir`은 빈 디렉터리만 삭제 — 데이터 자동삭제 없음). AC-LDS-002는 동작을 검증하고 수단은 자유(plan.md §B).

**구현 편차 1건 (하네스 수준, wrapper 무관):** GREEN 도달 전 `verify_predicate`가 jq 문자열 `"true"`(quote 포함)를 출력해 `[[ ]]` 비교가 실패 — bare boolean 식으로 교체(의미 불변, SKILL.md 검증 레시피와 동일 형태). 첫 GREEN 시도의 실패는 이 하네스 인코딩 결함이며 wrapper는 무수정.

**E4 PRESERVE:** 관측은 M1 커밋 직후 three-dot(`git diff --stat origin/main...HEAD`)로 완료 보고에 기록한다(관측 전 여기에 기입하지 않는다 — VCI §2).

M2 대기: live 일괄 드레인(AC-LDS-010 3조건, 캡처→드레인→검증 순서) + 로컬 인도물 적용 관측(AC-LDS-007 배선 부/AC-LDS-012) + AC-LDS-011 5검사 — 모두 유지자 머신(primary 체크아웃).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
