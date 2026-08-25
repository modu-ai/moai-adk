---
id: SPEC-LSEL-DRAIN-STALL-001
title: "LSEL 드레인 3주 정지 — 내구 기계 트리거(session_drain.sh) + 정지 신호 + 3.5k 백로그 일괄 드레인"
version: "0.2.0"
status: draft
created: 2026-08-25
updated: 2026-08-25
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: lsel
lifecycle: spec-anchored
tags: "lsel, drain, stall, sessionstart-trigger, flock, archive-before-overwrite, backlog, trigger-absence, user-owned"
era: V3R6
tier: M
depends_on: [SPEC-LSEL-LOCAL-EVOLUTION-001]
related_specs: [SPEC-V3R6-GRAPH-FRESHNESS-001]
---

# SPEC-LSEL-DRAIN-STALL-001 — LSEL 드레인 정지 복구 (트리거 내구화 + 백로그 소진)

## HISTORY

- 2026-08-25 (plan-phase, v0.1.0) 최초 작성. 카드 t259 (Class B, Tier S~M → Tier M 확정). 카드 전문이 범위 권위: "LSEL 드레인이 3주째 정지 — 인박스 3,420행 미처리, 아무도 통지받지 않음". 리드 실측(2026-08-25)과 plan-phase 재측정(본인, 동일 primary 체크아웃 live state)이 원인 규명을 확정했다: **트리거 부재** (session-scoped `/loop` 레시피가 소유 세션과 함께 소멸, 그리고 그 레시피는 애초에 아무것도 실행하지 않는 `console.log` 묶음) — 드레인 엔진 자체의 실패가 아니다 (`drain.sh` 실백로그 dry-run rc=0으로 건재 확인). SPEC-LSEL-LOCAL-EVOLUTION-001 (completed)의 M1 불변식(드레인은 `.moai/state/lsel/`에만 쓴다, memory/ 무쓰기)을 그대로 계승하며, 본 SPEC은 그 드레인에 **내구 트리거와 정지 가시성**을 부여한다.
- 2026-08-25 (plan-phase, iter-1 수정, v0.2.0) — plan-audit review-1 반영 (PASS-WITH-DEBT 0.92, Tier M 문턱 0.80). **D1** (MUST): §28 grep 관측 정정 — 실측 2건(backlog_check.sh:6 헤더 주석 + :50 리마인더 본문; 종전 1건 기록은 :50만 집계한 과소계상). **D2** (MUST): AC-LDS-010 — `629` 하드코딩 폐지(M2 재측정 `$OFFSET_BEFORE` 파라미터화) + 검증 창 취약성 폐쇄(배선 후 no-op 세션 시작 드레인도 clusters.json을 덮어쓰므로 후보·자기일관 조건은 clusters-history archived 사본으로 판정; 드레인 직전 $LIVE/$OFFSET_BEFORE 캡처 절차 명시). **D6** (MUST): PRESERVE diff 기준 ref를 명시적 `origin/main`(three-dot)에 결속 — 스테일 로컬 main ref에 의한 false-FAIL 차단. **D3** (SHOULD): 직접 `drain.sh` 호출 경로의 잔여위험 기록 + REQ-LDS-010 확장(모든 드레인 wrapper 경유, SKILL.md 검증 레시피 `session_drain.sh` 경유로 교체, PROPOSE는 archived 사본 판독). **D4** (SHOULD): REQ-LDS-006/§A "정확히 발화" 완화 — 임계 초과 시 발화(죽은 트리거의 관측 가능한 서명), 긴 세션 중반 누적에서도 발화 가능, 같은 SessionStart 내 wrapper 드레인과의 순서는 미지정. **D5** (SHOULD): SessionStart 배선 항목에 명시적 `"timeout": 30`(live 선례 형식) + REQ-LDS-005에 예산 수치 명시. D7(전방 참조)·D8(kickoff 상급 사항) 무변경.

## §A. User Story

**As a** MoAI-ADK 유지자(GOOS) whose dogfood repository의 관측 수집은 살아 있으나(lessons-inbox 매일 수백 행 적재) 드레인은 2026-08-04 이후 한 번도 돌지 않았고,
**I want** (1) 어떤 Claude 세션의 생사에도 의존하지 않는 기계 트리거, (2) 정지·지연을 드러내는 신호, (3) 밀린 ~3.5k 행의 일괄 처리,
**so that** 관측→후보 클러스터 파이프라인이 다시 흐르고, 다음 정지가 3주가 아니라 다음 세션 시작에 발각된다.

**결과 가설:**
- 드레인 오프셋이 live 인박스 끝까지 따라잡는다 (bulk drain, mutant-guard 검증 포함).
- 세션 시작마다 잠금-보호된 드레인이 실행되며, 동시 다발 세션 시작(lane 병렬이 상식인 이 저장소)에서도 상태가 부패하지 않는다.
- 다시 멈추면 backlog advisory가 같은 SessionStart 표면에 이를 알린다 — 발화 조건은 "세션 시작 시점에 unread 백로그가 임계 초과"(죽은 트리거의 관측 가능한 서명)이며, 긴 세션의 중반 누적으로도 발화할 수 있고 같은 SessionStart 안에서 wrapper 드레인과의 실행 순서는 미지정이다(경합 허용 — D4 완화).

## §B. Context and Background

### §B.1 실측 정지 상태 (모두 2026-08-25, primary 체크아웃 live state — 워크트리엔 존재하지 않는 runtime state)

| 신호 | 값 | 측정 |
|---|---|---|
| `.moai/lessons-inbox.jsonl` | 4,204행 (plan-phase 재측정; 리드 측정 시점 4,190행/1.1MB, 1분 창 +2행 — 컬렉션 살아 있음) | `wc -l` |
| `.moai/state/lsel/drain-offset.json` | `{"offset": 629, "updated": "2026-08-04T05:21:03Z"}` — 21일 동결 | 파일 판독 |
| `.moai/state/lsel/clusters.json` | 같은 날짜(mtime 2026-08-04 14:21:03 KST) 이후 무갱신 | `ls -la` |
| 미처리 백로그 | 4,204 − 629 = **3,575행** (카드 기술 3,420 → 리드 3,561+ → plan-phase 3,575; 측정마다 증가) | 산술 |

### §B.2 원인 규명 — 트리거 부재 (확정, 증거 귀속)

직접 원인: 드레인의 유일한 "스케줄"은 네이티브 `/loop` 스케줄러에 의해 구동되는 `.claude/workflows/lsel-drain-loop.js`였고, `/loop` 스케줄은 **session-scoped** (인메모리, 소유 세션과 함께 소멸, 디스크 기록 없음)이다. SPEC-LSEL-LOCAL-EVOLUTION-001 작업 세션이 2026-08-04/05에 종료되며 스케줄이 조용히 죽었다 — 마지막 드레인 기록(2026-08-04T05:21:03Z)과 SPEC 종료 시점이 일치한다.

가해 결함: 그 레시피는 **애초에 아무것도 실행하지 않는다**. 전체 본문이 드레인/백로그 명령을 출력하는 `console.log` (45-49행)인데, 헤더 주석(12행)은 "The recipe runs drain.sh"라고 주장한다. 즉 살아 있던 시절조차 "예정 드레인"은 모델 매개 출력이지 기계 트리거가 아니었다.

분류: **트리거 부재** — 실패 후 무재시가 아니다. `drain.sh`는 실백로그 dry-run(3,563행 delta, jq 파이프라인, rc=0)으로 건재가 확인됐다.

### §B.3 정지가 보이지 않았던 기여 결함 3건

1. **미배선 advisory**: `backlog_check.sh`는 SessionStart advisory로 설계됐으나(헤더: "AC-LSEG-007 SessionStart backlog-check ... see CLAUDE.local.md §28") 어디에도 배선돼 있지 않다. live `.claude/settings.json` `.hooks.SessionStart`는 정확히 1개 항목(handle-session-start.sh session-attribution). settings.json + settings.local.json hooks에 backlog/lsel grep → 0 matches.
2. **허위 자기기술**: §B.2의 레시피 헤더 주장(실행한다) vs 본문(출력만 한다).
3. **운영 문서 소실**: 종료된 SPEC의 progress.md는 "CLAUDE.local.md §28"을 LSEL 운영 지침 앵커로 참조하는데, live CLAUDE.local.md(35,355 bytes, mtime 2026-08-25 — 활성 유지보수 중)에는 §28/LSEL이 0회 등장한다. 운영자-facing 리마인더 표면이 없다.

### §B.4 백로그 구성 (실측 — 드레인 부하 우려에 대한 답)

3,563행 delta에 대한 dry-run: noise(severity filter: Bash:UnknownFailure, Bash:SandboxViolation, *:TimeoutError) 3,095 (86.9% — 카드의 ~65% 추정은 낮음), singleton 6, 후보 클러스터(빈도≥2) ≈13-15. 상위: `tool_failure:Read:UnknownFailure` (146), `Bash:OOMKilled` (88), `Bash:PermissionDenied` (58), `Bash:ExitError` (45), `Monitor:UnknownFailure` (36). 무효 JSON 크래시 없음. **클러스터링 부하는 실측으로 사소함이 입증됐다** — 카드의 부하 우려는 측정으로 종결.

### §B.5 2차 결함 — clusters.json 무조건 덮어쓰기 (본 SPEC 범위로 편입)

`drain.sh`는 매 실행마다 clusters.json을 덮어쓴다(122행). 심지어 empty-delta no-op 경로(63-76행)조차 빈-후보 clusters.json을 쓴다. 잦은 트리거 하에서 모델 매개 PROPOSE 단계에 대기 중인 후보가 다음 드레인에 조용히 유실된다. 본 SPEC의 wrapper는 drain.sh를 수정하지 않고 **호출 전 보존(archive-before-overwrite)**으로 이 결함을 닫는다(범위-최소).

> **잔여위험 + 휘발성 상호작용 (iter-1 D3)**: (a) 보존은 wrapper 매개다 — `drain.sh`를 **직접** 호출하면 여전히 덮어쓴다(레시피·문서는 REQ-LDS-010에 따라 전부 wrapper 경유로 교체하나, 직접 호출 가능성 자체는 남는다). (b) 세션 시작마다 드레인이 도는 체제에서 live `clusters.json`은 **휘발성**이다 — no-op 드레인조차 `candidates: [], total_read: 0`으로 덮어쓰므로(63-76행), PROPOSE 단계는 live 파일이 아니라 **archived `clusters-history/` 사본**을 읽어야 한다(REQ-LDS-010, AC-LDS-010 검증도 같은 근거).

## §C. Requirements (GEARS)

> 요구 레이어는 GEARS. 검증 레이어(Given-When-Then)는 acceptance.md가 소유 — 여기에 GWT를 재술하지 않는다.

- **REQ-LDS-001 (event-driven, 내구 트리거)** — **When** a Claude session starts in this repository, the session-drain wrapper (tracked at `.claude/skills/hns-lsel-curator/session_drain.sh`) shall acquire an exclusive lock on the LSEL state directory, archive any staged candidate clusters, execute `drain.sh`, and emit a one-line drain status (stubs read, candidates found, new offset).
- **REQ-LDS-002 (event-detected, 동시성)** — **When** another session holds the drain lock, the wrapper shall skip the drain, emit a contention notice, and exit successfully (concurrent lane-session starts are the normal case in this repository, not an error).
- **REQ-LDS-003 (event-detected + unwanted, 보존)** — **When** `clusters.json` holds one or more candidate clusters at wrapper invocation, the wrapper shall copy it to the history store (`.moai/state/lsel/clusters-history/`) before any overwrite occurs. The wrapper shall not silently discard staged candidates — `drain.sh` overwrites `clusters.json` on BOTH the drain path (line 122) and the empty-delta no-op path (lines 63-76), so the archive must precede the `drain.sh` invocation unconditionally.
- **REQ-LDS-004 (state-driven, no-op)** — **While** the recorded drain offset equals the inbox tail, the wrapper shall complete without advancing the offset and without reporting a failure.
- **REQ-LDS-005 (unwanted, fail-open 예산)** — The wrapper shall not block or fail session start on its own errors: wrapper errors degrade to a stderr notice and exit 0, and the wrapper's runtime stays within an EXPLICIT SessionStart hook timeout set by the wiring (30 seconds, matching the live SessionStart hook precedent in `.claude/settings.json`) — against a measured full-backlog drain runtime under 1 second for a 1.1MB / ~4.2k-line inbox (plan-phase dry-run, §B.4), leaving two orders of magnitude of headroom (advisory-check discipline, `.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline).
- **REQ-LDS-006 (event-driven, 정지 신호)** — **When** unread inbox backlog exceeds the configured threshold (default 25, `LSEL_BACKLOG_THRESHOLD`) at session start, `backlog_check.sh` shall surface the unread count, offset, and drain invocation as a SessionStart advisory — wired on the same local SessionStart surface as the wrapper (belt-and-suspenders: a threshold-exceeding backlog at session start is the observable signature of a dead mechanical trigger; the advisory may also fire on mid-long-session accumulation, and its ordering relative to the wrapper drain within the same SessionStart is unspecified — the two hooks may race).
- **REQ-LDS-007 (event-driven + unwanted, 백로그 방침)** — **When** the backlog drain executes, the drain shall process the FULL unread backlog to the live inbox tail in one pass (bulk drain), advancing the offset to the current live line count. The implementation shall not advance the offset without clustering the stubs the advance covers (카드 명명 mutant: "오프셋만 최신으로 밀고 실제 클러스터링은 안 하는 구현").
- **REQ-LDS-008 (event-driven, 검증된 완료)** — **When** the backlog drain completes, verification shall confirm — via the `hns-lsel-curator/SKILL.md` §Verification recipe (lines 134-145) — that the offset equals the live inbox line count AND at least one candidate cluster is staged AND `clusters.json` is self-consistent (`offset_after` == live count, `total_read` == delta size).
- **REQ-LDS-009 (ubiquitous, 배선 정직성)** — The SPEC artifacts shall record the SessionStart wiring (`.claude/settings.local.json`) and the CLAUDE.local.md LSEL operating-instructions section as maintainer-machine local deliverables with their apply steps stated, and shall state explicitly that the PR does not carry them — a tracked `.claude/settings.json` entry would be wiped by every `moai update` (CLAUDE.local.md §2.3), so tracked wiring is affirmatively wrong, not merely unconventional. The wrapper, its tests, the SKILL.md mirror, and the recipe/comment corrections DO ride the PR.
- **REQ-LDS-010 (capability gate, 위생)** — **Where** the `lsel-drain-loop.js` recipe remains in the tree, its header and body shall describe it truthfully as a model-mediated reminder that prints commands and does not execute anything, and shall point at the wrapper as the durable trigger; the `hns-lsel-curator/SKILL.md` shall carry a durable-ops section (trigger + drain + verification) mirroring the local instructions, shall route ALL drains through the wrapper — the §Verification recipe shall invoke `session_drain.sh`, not `drain.sh` directly (wrapper-mediated archiving does not extend to direct `drain.sh` invocations) — and shall state that the PROPOSE stage reads the archived `clusters-history/` copy (live `clusters.json` is ephemeral under per-session-start drains, §B.5); the dead `CLAUDE.local.md §28` anchor shall be removed from `backlog_check.sh`'s reminder text and header comment (both occurrences: line 6 header + line 50 body).

## §D. Constraints (HARD)

1. `.claude/skills/hns-lsel-*` 및 `.claude/workflows/lsel-*` 파일은 `internal/template/templates/**`에 절대 미러링되지 않는다 — CI guard 3종(template-neutrality-check.yaml, lsel-leak-guard.yaml, `internal/template/internal_content_leak_test.go`)이 오늘 초록이고 그대로 초록이어야 한다.
2. 동결 Go applier 무손상: `internal/harness/applier.go` write-flag 무수정. `internal/harness/curator_dispatch.go`도 무수정.
3. `.claude/rules/moai/**`, CLAUDE.md, 배포 템플릿 무수정.
4. 신규 `.moai/config/sections/` 파일 금지 (`moai update`가 지운다) — 루프 상태는 `.moai/state/lsel/` 하위만.
5. 드레인 경로는 `.moai/state/lsel/` 하위에만 쓴다 — memory/ 무쓰기 (선행 SPEC M1 불변식 계승).
6. `.moai/lessons-inbox.jsonl`은 불변(append-only) — 드레인은 오프셋만 전진.
7. `drain.sh` 코어는 무수정 — wrapper가 보존/잠금/호출을 소유한다 (기존 `drain_test.sh` characterization 초록 유지).
8. plan-phase 중 live 드레인 실행·settings 변경 금지 — 실행은 Implementation Kickoff Approval 이후 run-phase.

## §E. Local Deliverables (PR 미탑재 — 유지자 머신 적용)

| 항목 | 대상 | 이유 |
|---|---|---|
| SessionStart wrapper 배선 | `.claude/settings.local.json` `.hooks.SessionStart`에 `session_drain.sh` 항목 추가 — **명시적 `"timeout": 30`**(live SessionStart 선례 형식; REQ-LDS-005 예산) | settings.json(tracked)은 `moai update`가 통째 재배포(§2.3) — 배선이 매 업데이트마다 유실된다. settings.local.json은 runtime-managed/personal, update가 지우지 않는다. |
| SessionStart advisory 배선 | 같은 표면에 `backlog_check.sh` 항목 추가 — 동일하게 `"timeout": 30` 명시 | REQ-LDS-006의 발화 표면. |
| LSEL 운영 지침 복원 | CLAUDE.local.md에 LSEL 섹션(트리거·드레인·검증·PROPOSE 연계, 죽은 §28 앵커의 실체) | §B.3-3의 소실 복원. CLAUDE.local.md는 로컬 전용(템플릿 금지). |

적용 증거(적용 후 jq/grep 출력)는 progress.md §E.2에 기록한다. PR diff에는 위 3종이 0건 포함된다(AC-LDS-012).

## §F. Success Criteria

- 백로그 소진: offset == live `wc -l` && candidates ≥ 1 (mutant-guard AC-LDS-010 통과).
- 내구성: 세션 시작 트리거가 잠금 하에 동작, 동시성 no-op 확인(AC-LDS-002).
- 가시성: 트리거가 다시 죽으면 backlog advisory가 같은 세션 시작에 발화(AC-LDS-007).
- 정직성: 레시피 주장=본문 동작, 죽은 앵커 0, 로컬 배선이 SPEC에 문서화(AC-LDS-008/009/012).

## §G. Out of Scope

### Out of Scope — 분산 표면 (moai doctor / statusline 정지 신호)

- `moai doctor` 진단 항목·statusline 게이지에 unread 백로그/마지막 드레인 시각 표시는 본 카드 비목표다 — 분산 표면 + dev-local 상태라는 이유로 리드 브리프가 명시적으로 non-goal로 지정했다. 운영자가 kickoff에서 이의를 제기하면 후속 카드 후보로 기록한다 (REQ는 본 SPEC에 두지 않는다).

### Out of Scope — 오프셋 리셋 (recent-only 백로그 방침)

- "오프셋을 현재로 리셋하고 최근분만 드레인"하는 대안은 기각된 디폴트로 기록한다 — mutant-인접(카드 명명 mutant와 같은 형태: 오프셋만 전진)이며, 측정된 클러스터링 부하가 사소해서(§B.4) 정직한 일괄 드레인을 포기할 근거가 없다. kickoff에서 운영자가 명시 선택 시에만 재검토.

### Out of Scope — PROPOSE/APPLY 단계 및 drain.sh 코어 수정

- 후보 clusters.json 스테이징까지만 (선행 SPEC M1 불변식). PROPOSE(섀도 제안)·APPLY·memory/ 기록은 선행 SPEC의 M2/M3 소관이며 본 SPEC은 트리거·신호·백로그에 한정한다. `drain.sh` 파이프라인 수정(예: no-op 경로의 덮어쓰기 자체 제거)도 하지 않는다 — wrapper의 호출-전 보존으로 닫는다 (Constraint 7).

### Out of Scope — 템플릿 배포 (16-언어 graduation)

- `hns-lsel-*`의 `moai-lsel-*` graduation은 선행 SPEC 당시부터 별도 SPEC 과제로 분리돼 있다 — 그대로 유지.

## §H. Cross-references

- 선행: `.moai/specs/SPEC-LSEL-LOCAL-EVOLUTION-001/` (drain.sh·backlog_check.sh·clusters.json 스키마·6개 가용 표면의 원천)
- 연관 카드: t250 (드리프트 게이트, SPEC-V3R6-GRAPH-FRESHNESS-001 — "스테일 무통지" 같은 결함 형태)
- 엔진: `.claude/skills/hns-lsel-curator/{SKILL.md,drain.sh,drain_test.sh,backlog_check.sh,backlog_check_test.sh}`
- 트리거 결함 원천: `.claude/workflows/lsel-drain-loop.js`
- 배선 근거: CLAUDE.local.md §2.3 (moai update가 관리 뿌리를 통째 삭제 — tracked settings.json 배선 불가), `.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline (fail-open 예산)
- CI guard: `.github/workflows/{template-neutrality-check,lsel-leak-guard}.yaml`, `internal/template/internal_content_leak_test.go`
