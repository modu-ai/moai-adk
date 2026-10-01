---
id: SPEC-SESSION-ANCHOR-ATTR-001
title: "Bash 워크트리 세션 앵커 교차 레인 오판독 — 귀속 가능 계측기·재배치 감사·회피 규율 (card t1339, t1337+t1339 통합)"
version: "1.0.0"
status: completed
created: 2026-09-30
updated: 2026-10-01
author: "MoAI lane (card t1339)"
priority: P1
phase: "v3.2.0"
module: "internal/hook"
lifecycle: spec-anchored
tags: "session-anchor, worktree, observability, audit-log, worktree-guard"
tier: M
related_specs: [SPEC-SESSION-REGISTRY-READ-ANCHOR-001, SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001, SPEC-WORKTREE-REAPER-001, SPEC-WORKTREE-BRANCH-GUARD-001]
---

# SPEC-SESSION-ANCHOR-ATTR-001 — Bash 워크트리 세션 앵커 교차 레인 오판독 수리

## HISTORY

| 날짜 | 변경 | 근거 |
|------|------|------|
| 2026-09-30 | 최초 작성 (Tier M, plan-phase) | 리드 배차 — 카드 t1339, t1337+t1339 동일 근원 통합 수리 지시 |

---

## §A 문제 정의 및 분석 (측정 기반)

### §A.1 배경

카드 t1339(2026-09-29 등록): worker-66의 t1314 run 중 Bash 워크트리 세션 앵커가 다른 레인 트리(`.claude/worktrees/develop`, t1312의 트리)로 수 분간 잘못 해상 — 전 Bash 거절 후 공유 앵커가 되돌아왔다. 오실행은 없었으나(가드가 cwd 일치 강제), last-writer-wins 공유 앵커 구조 자체가 결함이다. 동일 근원의 선행 사례 t1337(2026-09-29, worker-69): 부모의 `EnterWorktree`(t1315) 이동 직후 서브에이전트의 쓰기 전면 거부 — 거부는 **서브 재시작 없이** 지속됐다(메모리 교훈 `feedback_subagent_anchor_follows_session_worktree.md` 축어: "거부는 서브 재시작 없이 지속됐다" — 조사 과정의 일부 렌즈가 퍼레프레이즈한 "across subagent restart" 표현은 1차 기록과 어긋나며 본 SPEC은 채택하지 않는다).

### §A.2 측정 확립 사실 (코드 실측, HIGH 신뢰)

1. **거부를 방출하는 가드는 Claude Code 런타임 상태다.** `internal/hook/post_tool_failure.go:40-48`이 축어 명시: "The guard belongs to the Claude Code runtime, not to this codebase". 저장소는 거부 토큰 `worktreeGuardAnchor = "isolated in the worktree"`를 고정하고(`worktree_guard_refusal_test.go`,살생 변이 2건 문서화) `WorktreeGuardRefusal`로 분류할 뿐이다. 런타임 앵커의 내부 저장/keying은 미관측(anchor-mechanism 렌즈 공백 — API 429).
2. **저장소가 소유한 모든 앵커 저장소는 공유되고 비세션 키다.**
   - git worktree lock: 트리당 1슬롯, card-id+pid 이유, claude 레인엔 세션 신원 없음(`internal/session/anchor_lock.go` — AUTHORITATIVE 선언, fail-closed `AnchorDecision` 합집합)
   - `.moai/state/active-sessions.json`: 체크아웃당 1파일(stateanchor git common-dir 수렴); session_id 키이나 각 항목의 `cwd`는 `Registry.RelocateSession`(`internal/session/anchor.go:126-142`)이 session_id 일치만 보고 덮어쓰는 last-writer-wins 필드 — 타당성 가드 없음, 감사 없음
   - `.moai/state/worktrees.json`: 경로 키 upsert
   - `.moai/state/current-session-id.txt`: 프로젝트당 1슬롯 매 SessionStart 덮어씀(`internal/cli/session.go:233-252` + `internal/cli/integration.go:121-140` — "외인 id로 잠긴 lock은 풀 수 있는 자가 없다")
   - `CLAUDE_ENV_FILE`의 `MOAI_PROJECT_DIR` 스탬프(`internal/hook/cwd_changed.go:77-84`): Go 소비자 없음 — 미증명 후보 표면
3. **재배치는 플랫폼 공백 위를 탄다.** EnterWorktree/ExitWorktree엔 `CwdChanged`가 방출되지 않고(t236/#1640), 재배치는 PostToolUse `handleWorktreeMove` + `CwdChanged` `relocateSessionCwd`(`internal/hook/cwd_changed_relocate.go`, `post_tool_worktree.go`)에 실려 fail-open이며, 재배치 로그가 전무하다. SPEC-WORKTREE-REAPER-001 §B.3: 레지스트리 cwd 5레인 중 4개 stale 실측. 한 훅 이벤트의 오트리 cwd가 이후 올바른 값으로 되돌아가는 형상은 t1339의 "수 분 오판 후 복귀"와 형상 일치(형상 일치, 재현 아님).
4. **디스크 상 생존 병리**: pid 41489가 active-sessions.json에서 3개 session_id로 존재; 폐기된 `.moai/worktrees/t1337`을 가리키는 stale 항목 3건; worktrees.json에 stale launcher 행 5건.
5. **재발 클래스**: 부모 이동형 t281/t508/t526/t530/t664 — t741 실측("cwd pinned at launch는 거짓 — 경로 없는 명령은 호출 시점에 세션 현재 앵커로 재해석: 12/12 통과, 두 번의 중간 이동을 걸쳐"; 경로 없는 명령은 거절 없이 **잘못된 트리에서 조용히 실행**). 교차 레인 재바인딩형: t707(1건 선행 — 공유 저장소 인과 주장은 사건 시 가설, 코드 검증 안 됨)과 t1339. t667: 카드 id 트리 이름도 소유 고유 아님.
6. **귀속 불능(t1064 결함 클래스)**: `.moai/harness/usage-log.jsonl`의 `WorktreeGuardRefusal` 행 4,358건(09-29 하루 173건)이 `subject:Bash` + `context_hash`만 실음 — session_id·트리 경로 없음. 09-29 행은 t1337/t1339/합법적 병합창 거절로 분리 불가. 재배치 역시 로그 없음.
7. **아직 미증명(정직하게 unknown으로 기록)**: 런타임 앵커의 저장/keying; t1339의 오트리 쓰기 방어가 어느 층이었는지(moai branch-guard Seam A vs 런타임 가드 — 둘 다 존재, 사건 시점 어느 쪽도 검증 안 됨); t1339/t707의 간헐성·자기복귀 역학.

### §A.3 심각도 논거 (W5 — Medium-High)

관측된 두 사건은 모두 거부형(denial-only, 오트리 쓰기 0건)이었으나, 클래스가 단순 거부 소음인 것은 아니다:

- **(a) 무음 변이 실측**: t741이 경로 없는 명령이 거절 없이 잘못된 트리에서 실행되는 SILENT 변이를 측정했다. 거부가 항상 오는 보증이 아니다.
- **(b) 처분 가드 오귀속**: 잘못된 레지스트리 cwd 항목은 `LiveAnchoredSessions`(처분 가드)를 먹인다 — 잘못 향한 앵커는 산 세션의 트리를 "비어 있음"으로 보일 수 있고, 그 가드는 "미푸시 브랜치의 유일본 파괴와 산 레인 사이의 마지막 방어선"이다.
- **(c) 재현 조건이 평상 운용 모드**: 재현 조건은 에지 케이스가 아니라 lane-parallel 표준 운용(factory/kanban) 그 자체다.

결론: 심각도 Medium-High. 리드의 병목 제안(P7, "앵커 네임스페이스 수리")이 이미 본 수리를 추가 자율화 작업의 전제로 시퀀싱하고 있다(`.moai/reports/autonomy-bottleneck-proposal-20260929.md` bottleneck #7).

### §A.4 수리 가능성의 경계

오염의 근원이자 거부를 방출하는 런타임 앵커 자체는 저장소 밖에 있다. 본 SPEC이 하는 것은 (1) 이 저장소가 **소유한** 결함 표면(비귀속 로그, 무감사 재배치)을 수리하고, (2) 런타임 결함의 다음 발생을 세션 귀속 가능하게 측정할 계측 스위치를 두며, (3) 회피 규율을 문서화하는 것이다 — 부분 산출이며, 그 한계를 명시한다.

---

## §B 용어

| 용어 | 정의 |
|------|------|
| 런타임 앵커 | Claude Code 런타임이 소유한 세션↔트리 격리 상태. 거부문("This session is isolated in the worktree …")의 방출자. 저장소 밖. |
| 앵커 저장소(anchor store) | 본 저장소가 소유한 세션↔트리 기록: git worktree lock, `active-sessions.json`, `worktrees.json`, `current-session-id.txt`, `MOAI_PROJECT_DIR` 스탬프 |
| 재배치(relocation) | `Registry.RelocateSession`이 레지스트리 항목의 `cwd` 필드를 훅 이벤트의 cwd로 갱신하는 행위 |
| 귀속(attribution) | 로그 행을 발생 세션(session_id)과 해상 cwd/트리로 식별 가능하게 하는 성질 |
| 앵커 트레이스 | `MOAI_ANCHOR_TRACE=1`로 켜지는 앵커 의사결정별 상세 로그 |

---

## §C 요구사항 (GEARS)

### §C.1 W1 — 거부 귀속 가능화

- **REQ-SAA-001** (Event-driven): **When** the failure observer records a `WorktreeGuardRefusal` classification row, the failure observer shall include `session_id`, the resolved working directory (cwd), and the rejection-quoted worktree path in the log row, in addition to the existing `subject` and `context_hash` fields.
- **REQ-SAA-002** (Event-detected): **When** a session identifier or tree path cannot be resolved at refusal-record time, the failure observer shall write the row with explicit `session_id: "unknown"` / `cwd: "unknown"` markers and shall not drop or defer the row.

### §C.2 W2 — 재배치 감사 + 소유 타당성

- **REQ-SAA-003** (Event-driven): **When** `Registry.RelocateSession` rewrites an entry's `cwd` field, `RelocateSession` shall append an audit row recording `session_id`, the previous cwd, the new cwd, the trigger hook event, and a timestamp.
- **REQ-SAA-004** (Event-driven): **When** a relocation's new cwd resolves to a tree whose git worktree lock owner, compared against the relocating session per the ownership rule below, falls into the "other live card" or "registry-only" case, the relocation path shall emit an advisory flag row to the relocation audit log and proceed with the relocation.

**REQ-SAA-004 소유 비교 규칙 (케이스별 플래그 판정표).** 카드 id 원천은 git worktree lock reason 문자열 파싱(`claude session <card-id> (pid <n> …)` 형식 — `internal/session/anchor_lock.go`의 lock-reason 파싱 seam과 `anchor_lock_holder.go`의 pid 토큰 해석을 재사용). `AnchorVerdict`는 카드 신원 필드를 갖지 않으므로({Anchored, Source, Detail} — anchor_lock.go:50-57), 비교는 (i) lock reason의 card-id 토큰과 (ii) lock reason의 pid 토큰 대상 세션 레지스트리 항목의 pid 위에서 수행한다:

| 케이스 | 판정 조건 | 기본 동작(advisory 도크트린) | REQ-SAA-005 config on 시 |
|--------|-----------|------------------------------|--------------------------|
| self-owned | lock reason pid == 재배치 대상 세션 항목의 pid (같은 프로세스가 쓴 lock) | 감사 행만, 플래그 없음 | 허용 |
| other live card | lock reason card-id가 세션 자신과 다르고 lock holder가 생존(`LockHolderConfirmedDead` 아님) | advisory 플래그 행 + 재배치 진행 | 거부 + 거부 기록 |
| unreadable | lock reason 파싱 실패 — 기존 fail-closed("treated as anchored") 준수 | 플래그 없음, 감사 행에 `owner: unreadable` 기록 | 허용 (판독 불가를 오판으로 삼아 산 lock을 차단하는 일 방지) |
| registry-only | 목표 트리에 lock 부재, 앵커 근거가 보조 소스(레지스트리)뿐 | 소유 판정 불가 → advisory 플래그(측정 유지) + 진행 | 거부 + 거부 기록 |

알려진 한계(기록용): pid 비교는 t958의 pid 재사용 한계를 물려받는다 — self-owned 오판 가능성이 있으나 본 REQ의 출력은 advisory 기본이므로 오류 비용은 플래그 누락 1행이다.
- **REQ-SAA-005** (Capability gate): **Where** the opt-in guard config (`workflow.anchor_relocation_guard.enabled`, mirroring the `workflow.branch_guard` / `agent_stop_guard` pattern, default false) is enabled, the relocation path shall refuse the flagged relocation and record the refusal to the relocation audit log.
- **REQ-SAA-006** (Ubiquitous invariant): The two-pass relocation candidate ordering (pass 1 upward walk, pass 2 primary registry) and fail-open error behavior pinned by SPEC-SESSION-REGISTRY-READ-ANCHOR-001 REQ-RAR-002/003/004 shall remain unchanged.

### §C.3 W3 — 앵커 트레이스 스위치

- **REQ-SAA-007** (Capability gate): **Where** `MOAI_ANCHOR_TRACE=1` is set, the anchor decision path shall emit one verbose log line per anchor decision — anchor reads in `checkBranchState` (Seam A), `RelocateSession` calls, and `AnchorDecision` outcomes — each carrying `session_id`, `pid`, `cwd`, and a timestamp.
- **REQ-SAA-008** (Event-detected default): **When** `MOAI_ANCHOR_TRACE` is unset or falsy, the anchor decision path shall emit no trace output and shall add no per-decision overhead beyond a single environment lookup.
- **REQ-SAA-009** (Ubiquitous): The trace switch's environment-variable name constant shall be declared in `internal/config/envkeys.go` (no hardcoded env-var names in the codebase).

### §C.4 불변식 (하드 제약)

- **REQ-SAA-010** (shall not): The `WorktreeGuardRefusal` classification anchor (`worktreeGuardAnchor = "isolated in the worktree"`) and its killed-mutant coverage in `worktree_guard_refusal_test.go` shall not be altered.
- **REQ-SAA-011** (shall not): The `LiveAnchoredSessions` read path shall not be single-sourced to the registry, and disposal-guard semantics shall not be altered (SPEC-SESSION-REGISTRY-READ-ANCHOR-001 REQ-RAR-005/006/008 — R1 landing gate).
- **REQ-SAA-012** (shall not): Seam A separation shall persist: `checkBranchState` shall query the git context at the command's actual cwd (`input.CWD` chain), and the audit-log directory anchoring shall remain `CLAUDE_PROJECT_DIR`-based (SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001).
- **REQ-SAA-013** (shall not): Anchor instrumentation shall live only where it actually executes (the Go hook layer; hook scripts resolve through the primary checkout's copy per the t1064 measurement) — no instrument shall be placed in a worktree-local hook copy — and `RefuseMutationFromNonCanonicalTree` / `MOAI_HOME` mutation-target semantics shall remain unchanged.

### §C.5 W4 — 회피 규율 문서

- **REQ-SAA-014** (Ubiquitous): `.claude/rules/moai/workflow/worktree-integration-ops.md` (template source 우선 — Template-First) shall document: (a) the anchor-misresolution recovery procedure (`ExitWorktree` keep + re-`EnterWorktree`, per the `feedback_worktree_cd_wedge` precedent), (b) the runtime-boundary statement (the emitting anchor is Claude Code runtime state), (c) the silent path-less-command hazard (t741: a path-less command silently runs in the wrong tree), and (d) the corrected t1337/t1339 incident records (the t1337 rejection persisted **without** subagent restart).

---

## §D 제외 사항

### Out of Scope — Claude Code 런타임 앵커 내부

- 런타임 앵커의 저장 위치·keying·쓰기 역학의 수리 또는 변경 — 저장소 밖이며 본 SPEC의 도달 범위가 아니다. (도크트린 명시 거절: per-agent anchor registration — `worktree-integration-ops.md:143-156` "Claude Code 바이너리 변경을 요하고 가드의 취지를 뒤집는다")
- 런타임 앵커 내부 keying에 대한 단정적 기술 — 미관측 사항이며 본 SPEC은 unknown으로 기록한다.

### Out of Scope — 처분 가드 시맨틱 변경

- `LiveAnchoredSessions` 읽기 경로의 레지스트리 단일 소스화 — R1 착지 게이트(SPEC-SESSION-REGISTRY-READ-ANCHOR-001 REQ-RAR-005/006/008)가 산 live orphan 항목 이관 또는 운영자 판정 기록을 전제로 하며, 본 SPEC은 그 전제를 다루지 않는다.
- worktree 폐기·sweep 생명주기 변경(t1369 소관).

### Out of Scope — 세션 종료 시점 메모리 갱신

- 세션 종료 시 자동메모리(`feedback_subagent_anchor_follows_session_worktree.md` 등) 갱신은 run-phase 후속 작업이며 SPEC 아티팩트 저작이 아니다.

### Out of Scope — 거부 문구 자체

- `worktreeGuardAnchor` 상류 거부 문구의 변경 — 상류 의존 고정값이며 REQ-SAA-010으로 보존한다.

---

## §E 참조

- 연구 종합: 본 SPEC 디렉터리 `research.md` (영속 운반체 — 축어 보존 완료). 3개 렌즈 원본은 /tmp 기계 로컬 세션 스크래치(`` `/tmp/t1339-lens-incident.md` ``, `` `/tmp/t1339-lens-constraints.md` ``, `` `/tmp/t1339-lens-priorspec.md` ``) — **OS 청소 대상 휘발 경로로 소멸 가능**; 내용은 research.md가 전재하므로 정보 손실 없음. 재열람이 필요하면 본 SPEC 디렉터리의 research.md를 1차 근거로 사용한다.
- 선행 SPEC: SPEC-SESSION-REGISTRY-READ-ANCHOR-001 (재배치 읽기 경로·R1 게이트), SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001 (Seam A), SPEC-WORKTREE-REAPER-001 (§B.3 stale 실측), SPEC-WORKTREE-BRANCH-GUARD-001 (1-작성자 규율)
- 메모리 교훈: `feedback_subagent_anchor_follows_session_worktree.md` (t1337), `feedback_worktree_cd_wedge.md` (복구 절차), `feedback_a_shared_log_marker_must_be_unique_before_it_can_attribute.md` (t1064), `feedback_switching_worktrees_degrades_live_background_agents.md` (t741 실측), `feedback_lane_agent_anchor_rebinding.md` (t707 — 가설로만 취급), `feedback_the_hook_copy_that_runs_is_the_primary_checkouts.md` (훅 사본 제약)
- 리드 보고: `.moai/reports/lead/autonomy-report-inputs-20260929.md` (row 2), `.moai/reports/autonomy-bottleneck-proposal-20260929.md` (bottleneck #7, P7)
