---
id: SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001
document: spec-compact
version: 0.1.0
created: 2026-08-26
updated: 2026-08-26
source: spec.md (Requirements + Acceptance Criteria + Exclusions sections)
---

# SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001: Compact View

본 문서는 spec.md의 Requirements (GEARS), Acceptance Criteria, Exclusions 세 섹션만 발췌한 컴팩트 뷰이다. 자동 생성된 참고용이며, 계약의 단일 원천은 spec.md이다.

## Requirements (GEARS)

- **REQ-TRSW-001** (Unwanted): **The MoAI 세션**은(는) `TaskStop`으로 정지된 티메이트를 이름으로 호명하는 메시지를 발신해서는 안 된다 — 정지 티메이트의 이름 주소는 살아 있고, 배달된 메시지 한 통이 transcript로부터 해당 에이전트를 재개(부활)시켜 무소유 작성자로 되살린다. [cross-session-messaging.md + 트윈]
- **REQ-TRSW-002** (Event-driven): **When** 이름으로 티메이트에게 메시지를 구성할 때, **the 발신 세션**은(는) 호명 대상이 정지 티메이트가 아님을 송신 전에 확인하고, 정지 티메이트 관련 조율은 소유 오케스트레이터(the owning orchestrator)·리드를 경유해서만 전달한다. [cross-session-messaging.md + 트윈]
- **REQ-TRSW-003** (State-driven): **While** 한 워크트리에 대한 감사가 활성 상태인 동안(개시 측정부터 판정 착지까지), **the 해당 워크트리**는(은) 정확히 한 명의 작성자(트리 소유 세션)만을 가지며, 이전 감사 세션의 보고서·스크립트 산출물을 포함한 타 세션 커밋은 감사 창이 닫힐 때까지 지연된다. [agent-common-protocol.md §Background Agent Execution + 트윈]
- **REQ-TRSW-004** (Event-detected): **When** 활성 감사 중인 워크트리에서 예상 밖 HEAD 이동·타 세션 커밋(foreign commit)이 관측되거나 정지 티메이트의 재실행이 관측되면, **the 관측 세션**은(는) 이를 리드에게 프로세스 결함으로 즉시 보고하고 단계 기록에 남긴다 — 조용히 계속하거나 추가 메시지를 발신하지 않는다. [양 파일쌍]
- **REQ-TRSW-005** (Ubiquitous): **The 본 SPEC이 추가하는 모든 절**은(는) 템플릿 트윈에 중립 형태로 동일 반영된다 — 트윈에 SPEC ID·커밋 SHA·내부 일자 토큰 금지, 사고 근거 인용은 dev 전용 SPEC 산출물에만, 로컬↔트윈 차이는 기존 의도적 차이(Origin 라인)만 허용. [4개 파일 전부]

## Acceptance Criteria

- **AC-TRSW-001**: `grep -c 'stopped teammate'` ≥2 및 `grep -c 'owning orchestrator'` ≥1 — CSM 로컬·트윈 각각. RED 실측(2026-08-26): 전부 0. → M1
- **AC-TRSW-002**: `grep -c 'actively audited'` ≥2 및 `grep -c 'foreign commit'` ≥1 — ACP 로컬·트윈 각각. RED 실측: 전부 0. → M1
- **AC-TRSW-003** (PRESERVE): `cmp ACP ACPT` 무출력 유지; `diff CSM CSMT` 트윈 전용 라인(`^>`) 0, 로컬 전용은 기존 2라인(Origin+공백, hunk `113,114d112`)만 — 신규 hunk 0. 오늘 실측과 동일 형태. → M1
- **AC-TRSW-004** (PRESERVE): 정밀 패턴 `SPEC-[A-Z][A-Z0-9]*-[0-9]{3}|202[0-9]-[0-9]{2}|[0-9a-f]{40}` 스캔 양 트윈 0/0 (기준선 0/0 — 조악 패턴 유일 적중은 ACPT:292 CLI 플레이스홀더). → M1
- **AC-TRSW-005** (PRESERVE): `go test -count=1 ./internal/constitution/...` → `ok` 유지(기준 실측 `ok … 0.708s`, 2026-08-26) — 레지스트리 101 엔트리·digest·ACP 등록 clause 13개 무변경. → M2
- **AC-TRSW-006** (예산): CSM 순증 ≤16행(현재 126), ACP 순증 ≤10행(현재 362) — `git diff --stat` 대 plan-phase base. → M1

REQ-TRSW-004는 별도 AC 없이 AC-001/002의 토큰 앵커가 같은 절 블록 내 표면화 문장을 함께 고정한다(교리 계층 — 검증 대상은 산출물, 런타임은 t267 경계 밖).

## Exclusions

- **t267 메커니즘 계층** — 정지 티메이트 이름 주소 회수(TaskStop 주소 회수 / reject-not-revive / 표면화)는 동반 카드 t267 소관; 본 SPEC은 교리만. hook/stop-goal-evaluator 가설은 t267이 배제.
- **Claude Code 런타임 동작 변경** — auto-resume은 런타임 동작(2.1.77); Go 코드 변경 없음.
- **기계적·훅 강제** — 정지-상태 신호가 없는 지금 강제하면 추측의 코드화; t267 결정의 하류. 교리 + REQ-004 표면화가 Tier S 범위.
- **추가 편집 표면** — `sync-auditor.md`(+.codex .toml 미검토), `manager-lead.md`, `kanban-dispatch.md`(+detail), `orchestration-mode-selection.md` §C.1, `cross-session-messaging-detail.md`, `sync.md`, `worktree-integration.md`, `.claude/rules/local/` 전부 무편집 — always-loaded 2 서식지가 전 세션 도달; 부족 판정 시 후속 카드.
