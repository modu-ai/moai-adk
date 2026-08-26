---
id: SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001
title: "Doctrine: prevent SendMessage revival of stopped teammates + sole-writer discipline for worktrees during active audits"
version: "0.1.0"
status: completed
created: 2026-08-26
updated: 2026-08-26
author: manager-spec (card t269)
priority: P2
phase: "v3.1.4 target"
module: ".claude/rules/moai/workflow/cross-session-messaging.md, .claude/rules/moai/core/agent-common-protocol.md"
lifecycle: spec-anchored
tags: "doctrine, agent-teams, cross-session-messaging, audit-hygiene, template-mirror"
tier: S
---

# SPEC: 정지 티메이트 부활 방지 + 활성 감사 중 워크트리 단독-작성자 규율 (교리 계층)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-26 | manager-spec | 최초 작성 — t232 sync-audit verdict(2026-08-25, PASS 0.92)의 process Finding F4를 카드 t269 Tier S 교리 SPEC으로 구조화. 6개 AC의 RED 기준값 전부 2026-08-26 t269 워크트리에서 실측(§3). 동반 카드 t267(메커니즘 계층)과의 경계는 §7 |

## 1. 문제 — 측정된 형태

### 1.1 사고 A — TaskStop 2회에도 조율 SendMessage 1건이 정지 티메이트를 부활시켰다

SPEC-ZONE-REGISTRY-RESYNC-001(card t232) 진행 중, `zrr-spec-amend` 티메이트는 오케스트레이터의 "이 커밋 후 종료" 지시 이후 커밋(`2319df7ac`)을 착지해 TaskStop으로 정지되었다(§F:30). 이후 **TaskStop 2회에도 불구하고 타 요원의 조율 SendMessage 1건이 이를 부활**시켰고, 그 부활 창에서 M2 가드 초안(`49630cba2`, `adde4cfc9`), M3 최종 검증(`a74362427`), sync 3-phase close(`a35ff0c60`~`0d8e3ce32`)가 자발 실행됐다(§F:31). 리드가 재정지·영구 정지했고(§E.2:381) 착지물은 소유 manager-develop 검토·오케스트레이터 전량 재검증으로 수용됐으나, 소유권 위반은 프로세스 부채로 남았다.

메커니즘은 Claude Code 런타임 동작이지 본 저장소 코드가 아니다: 정지된 에이전트의 이름 주소는 그대로 살아 있으며, 그 이름으로 배달된 메시지는 transcript로부터 해당 에이전트를 재개한다(`SendMessage` auto-resume — `.moai/research/cc-changelog-snapshot-2.1.233.md` :3236-3237, 2.1.77). 교훈은 auto-memory `feedback_sendmessage_revives_stopped_teammate.md`(MoAI 프로젝트 메모리 루트, 존재 오케스트레이터 확인 2026-08-26)에 기록됐다 — "제지 조율은 오케스트레이터 경유로만".

### 1.2 사고 B — 활성 감사 중인 워크트리에 제2 작성자가 착지했다 (`11df9587a`)

t232 sync-audit 진행 중(감사 개시 HEAD `ef93a9d1e` → porcelain clean 관측 사이), **이전 sync-auditor 세션**이 자신의 보고서·스크립트 커밋 `11df9587a`(보고 4파일, +318줄)를 감사 대상 워크트리에 착지·푸시했다 — reflog로 동일 워크트리 발생 확인(verdict :4, :49). 판정문은 이를 "§F가 기록한 부활 패턴의 재발이며 병렬-작성자 위생 위반"으로 규정했다(§프로세스 판정 (c)). 벡터는 읽기 전용 in-process 감사 서브에이전트가 아니라 **peer 세션**이다 — 서브에이전트 읽기 전용 규칙(도구 제한)은 도달하지 않는 경로다.

### 1.3 왜 교리 계층인가 — t267 경계

부활을 막는 **메커니즘**(어느 계층이 정지 티메이트의 이름 주소를 회수하는지: TaskStop 주소 회수 / reject-not-revive / 로그·커밋 귀속 표면화)은 동반 카드 **t267**(queued, Class C, Tier S~M) 소관이다. 본 SPEC(t269)은 **교리 계층**만 담당한다 — 규칙 문서·에이전트 정의에 서술되는 규범. 메커니즘이 없는 동안 교리 + 표면화 의무(REQ-TRSW-004)가 유일한 방어선이며, 이 한계는 §4에 명시한다.

### 1.4 왜 이 2개 파일인가 — 편집 표면 최소화

두 규칙 모두 **always-loaded** 규칙 파일 2개에 각각 서식한다: (1) `cross-session-messaging.md` — SendMessage 교리 SSOT(레지스트리 엔트리 0개, 신규 절은 레지스트리 중립); (2) `agent-common-protocol.md` §Background Agent Execution — 기존 동시성 단독-작성자 불변식의 자연한 확장 지점(감사 세션을 포함한 전 에이전트가 로드). 각 파일의 템플릿 트윈(`internal/template/templates/` 하위)에 동일 중립 텍스트로 반영한다. 후보 표면 전수와 제외 사유는 §7에 있다.

## 2. 요구사항 (GEARS)

### 2.1 사용자 스토리

칸반 리드로서, 나는 정지한 티메이트를 실수로 되살리는 어떤 메시지도 나가지 않기를 원한다 — 부활 창의 무소유 커밋이 내 카드의 감사·sync 무결성을 파손하기 때문이다. 그리고 감사관으로서, 나는 감사 중인 트리에 내 이전 세션이 보고서를 착지하는 일 없이, 감사 창이 닫힐 때까지 트리에 정확히 한 명의 작성자만 있기를 원한다.

### 2.2 REQ 목록

- **REQ-TRSW-001** (Unwanted — `shall not`): **The MoAI 세션**은(는) `TaskStop`으로 정지된 티메이트를 이름으로 호명하는 메시지를 발신해서는 안 된다 — 정지된 티메이트의 이름 주소는 살아 있으며, 그 이름으로 배달된 메시지 한 통이 transcript로부터 해당 에이전트를 재개(부활)시켜 소유자 없는 작성자로 되살린다.
  - 대상 파일: `.claude/rules/moai/workflow/cross-session-messaging.md` + 템플릿 트윈 (Rules 절 신규 불릿 및 Anti-patterns 신규 엔트리)
  - 근거: §1.1 (progress.md §F:30-31, §E.2:381)

- **REQ-TRSW-002** (Event-driven — `When`): **When** 이름으로 티메이트에게 메시지를 구성할 때, **the 발신 세션**은(는) 호명 대상이 정지 티메이트가 아님을 송신 전에 확인하고(liveness 확인), 정지 티메이트에 관한 조율은 그 티메이트를 정지시킨 소유 오케스트레이터(the owning orchestrator)·리드를 경유해서만 전달하며, 정지된 이름으로 직접 보내지 않는다.
  - 대상 파일: `cross-session-messaging.md` + 트윈 (REQ-TRSW-001과 동일 불릿 내)
  - 근거: §1.1 + auto-memory 교훈("제지 조율은 오케스트레이터 경유로만")

- **REQ-TRSW-003** (State-driven — `While`): **While** 한 워크트리에 대한 감사가 활성 상태인 동안(감사 개시 측정 시점부터 판정 착지 시점까지), **the 해당 워크트리**는(은) 정확히 한 명의 작성자 — 해당 트리를 소유한 세션 — 만을 가진다. 이전 감사 세션이 자기 보고서·스크립트 산출물을 착지시키는 경우를 포함하여, 모든 타 세션의 해당 트리 커밋은 감사 창이 닫힐 때까지 지연된다.
  - 대상 파일: `.claude/rules/moai/core/agent-common-protocol.md` §Background Agent Execution + 템플릿 트윈 (기존 동시성 문단 뒤 신규 문단, 제목 변경 없음)
  - 근거: §1.2 (verdict :4, §(c):49, F4:78)

- **REQ-TRSW-004** (Event-detected — `When` 비원하는 상태 관측): **When** 활성 감사 중인 워크트리에서 예상 밖 HEAD 이동 또는 타 세션 커밋(foreign commit)이 관측되거나, 정지 티메이트의 재실행이 관측되면, **the 관측 세션**은(는) 이를 리드에게 프로세스 결함으로 즉시 보고하고 해당 단계 기록(progress record)에 남긴다 — 조용히 계속하거나 해당 티메이트에게 추가 메시지를 발신하지 않는다.
  - 대상 파일: 위 2개 파일쌍 (부활 표면화 문장은 CSM 불릿 내, 작성자 표면화 문장은 ACP 문단 내)
  - 근거: §1.1-1.2 (t232의 §F 공개·재정지 패턴의 교리화)

- **REQ-TRSW-005** (Ubiquitous): **The 본 SPEC이 추가하는 모든 절**은(는) 템플릿 트윈에 중립 형태로 동일 반영된다 — 트윈에 SPEC ID·커밋 SHA·내부 일자 토큰이 없어야 하며, 사고 근거 인용(t232, SHA)은 dev 전용 SPEC 산출물에만 존재한다. 편집 대상 2파일에서 로컬↔트윈 차이는 기존 의도적 차이(cross-session-messaging.md Origin 라인)만 허용된다.
  - 대상 파일: 4개 파일 전부 (신규 절 텍스트 자체를 양 사본 동일·중립으로 저작)
  - 근거: 템플릿 중립성 가드 `internal_content_leak_test.go` + `template_neutrality_audit_test.go` + CI; 쌍둥이 저작 선례(research 교정 C2)

## 3. 인수 기준 (AC — Tier S 인라인)

전 AC가 2026-08-26 t269 워크트리(계획 단계, M1 전)에서 RED 기준값을 실측했다(2-셀 채택 규율, `verification-completeness.md` §2). 경로 약어: CSM = `.claude/rules/moai/workflow/cross-session-messaging.md`, CSMT = `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md`, ACP = `.claude/rules/moai/core/agent-common-protocol.md`, ACPT = `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`.

### 개요 매트릭스

| AC | 유형 | 커버 REQ | Green 마일스톤 |
|----|------|----------|----------------|
| AC-TRSW-001 | 채택(RED→GREEN) | REQ-001, 002 | M1 |
| AC-TRSW-002 | 채택(RED→GREEN) | REQ-003, 004(절반) | M1 |
| AC-TRSW-003 | PRESERVE(비회귀) | REQ-005 | M1 |
| AC-TRSW-004 | PRESERVE(비회귀) | REQ-005 | M1 |
| AC-TRSW-005 | PRESERVE(비회귀) | REQ-005(간접) | M2 |
| AC-TRSW-006 | 예산 제약 | 전 REQ | M1 |

### AC-TRSW-001 — 부활 금지·오케스트레이터 경유 절 착지 (CSM)

- **Given** M1 전 t269 트리 **When** `grep -c 'stopped teammate'` 을 CSM 로컬·CSMT 각각 실행 **Then** 0 — 실측(2026-08-26): 로컬 0 / 트윈 0 (ACP 양측도 0). `grep -c 'owning orchestrator'` 역시 로컬 0 / 트윈 0.
- **Given** M1 완료 후 **When** 동일 2개 명령 **Then** `stopped teammate` 로컬·트윈 각 ≥2 (Rules 불릿 + Anti-patterns 엔트리), `owning orchestrator` 각 ≥1.

### AC-TRSW-002 — 감사창 단독-작성자 절 착지 (ACP)

- **Given** M1 전 **When** `grep -c 'actively audited'` 및 `grep -c 'foreign commit'` 을 ACP 로컬·ACPT 각각 실행 **Then** 전부 0 — 실측(2026-08-26): `actively audited` 0/0, `foreign commit` 0/0.
- **Given** M1 완료 후 **When** 동일 2개 명령 **Then** `actively audited` 로컬·트윈 각 ≥2 (주제문 + When 절), `foreign commit` 각 ≥1.

### AC-TRSW-003 — 트윈 패리티 (PRESERVE — 오늘 초록, M1 후에도 초록이어야)

- **Given** 오늘 트리 — 실측(2026-08-26): `cmp ACP ACPT` 무출력(바이트 동일); `diff CSM CSMT`는 단일 hunk `113,114d112` + 로컬 전용 2라인(Origin 라인과 그 공백)뿐, 트윈 전용 라인 없음.
- **Given** M1 완료 후 **When** `cmp ACP ACPT` **Then** 여전히 무출력(rc=0). **When** `diff CSM CSMT` **Then** 트윈 전용 라인(`^>`) 0개, 로컬 전용 라인(`^<`)은 기존 2라인(Origin 라인 + 공백)만 — 신규 hunk 0개.

### AC-TRSW-004 — 템플릿 중립성 (PRESERVE)

- **Given** 오늘 트리 — 실측(2026-08-26): 조악한 상위 패턴 `SPEC-[A-Z]|202[0-9]-[0-9]{2}|[0-9a-f]{9}` 로 CSMT 0 / ACPT 1(292행 `moai session list --json --filter-spec=<SPEC-ID>` — CLI 플레이스홀더, 숫자 꼬리 없음).
- **Given** M1 완료 후 **When** 정밀 패턴으로 양 트윈을 스캔:

  ```
  grep -cE 'SPEC-[A-Z][A-Z0-9]*-[0-9]{3}|202[0-9]-[0-9]{2}|[0-9a-f]{40}' <CSMT> <ACPT>
  ```

  **Then** 0/0 (정밀 패턴은 `<SPEC-ID>` 플레이스홀더에 미부합 — 기준선 0/0, 신규 절은 어떤 토큰도 추가하지 않는다).

### AC-TRSW-005 — zone-registry 비회귀 (PRESERVE — ACP는 등록된 파일, 엔트리 13개)

- **Given** 오늘 트리 — 실측(2026-08-26): `go test -count=1 ./internal/constitution/...` → `ok github.com/modu-ai/moai-adk/internal/constitution 0.708s`.
- **Given** M2 **When** 동일 명령 재실행 **Then** `ok` 유지 — 101 엔트리·튜플 digest `2edb5384…` 핀과 ACP의 등록 clause 13개가 무변경임을 기계적으로 확인.

### AC-TRSW-006 — always-loaded 예산 (제약 AC)

- **Given** 오늘 트리 — 실측: CSM 로컬 126행, ACP 로컬 362행; diff 0 (자명 충족).
- **Given** M1 완료 후 **When** plan-phase base 대비 `git diff --stat` **Then** CSM 순증 ≤16행, ACP 순증 ≤10행 (최악 합계 ~26행, always-loaded 문맥 비용 상한).

**돌연변이 탐침(§2 계론)**: AC-001/002는 방향성 존재 검사다 — "토큰은 있으나 구속력 없는 절" 돌연변이는 4개 파일 전부에 동시에 위조되어야 AC를 통과한다(절 텍스트가 REQ 문장 그 자체이고 AC-003/004가 트윈 형태를 고정). AC-005가 레지스트리 결합 파일의 기계적 하한선이다. REQ-TRSW-004는 별도 AC 없이 AC-001/002의 토큰 앵커가 같은 절 블록 내 표면화 문장을 함께 고정한다 — 교리 계층 SPEC의 검증 대상은 런타임 동작(§7 t267)이 아니라 산출물(텍스트 착지·패리티·중립성·비회귀)이다.

## 4. 제약

- **[HARD] Template-First 순서**: 템플릿 트윈을 먼저 편집 → `make build` → 로컬 미러. 로컬 먼저 편집 금지.
- **[HARD] 중립 저작**: 신규 절은 양 사본 동일·중립 텍스트(REQ-TRSW-005). t232·SHA 인용은 SPEC 산출물에만.
- **[HARD] 레지스트리 무손상**: ACP의 기존 등록 clause 13개 verbatim 무변경, 신규 텍스트에 등록 clause 리터럴 재인용 금지(once 의미론 파손 방지). heading 추가·변경 없음(anchor 안전).
- **의도적 트윈 diff 보존**: CSM Origin 라인(hunk `113,114d112`)은 "수리"하지 않고 보존 — orchestration-mode-selection.md:142/144와 함께 확립된 중립성 규율의 산물이다.
- **예산**: AC-TRSW-006 상한(CSM ≤16행, ACP ≤10행 순증). 두 파일 모두 always-loaded다.
- **Go 코드 변경 없음**: 본 SPEC의 산출물은 markdown 규칙 문서 4파일뿐이다.
- **워크트리 가드**: 검증 명령은 단일 단순 명령으로 발행(복합 shell 구조는 가드 거부 — 계획 단계 실측 2회).

## 5. Tier 분류

**Tier S** — 근거: 문서 전용(코드 0행), 순증 ~26행 상한, 논리적 변경 1건(2 룰 파일 + 트윈 = 물리 4파일, Tier S "파일 <5"는 트윈 이중화를 논리 1건으로 계수), REQ 5 / AC 6 (Tier S 천장 8/8 이내). 산출물 2종(spec.md + plan.md, AC는 §3 인라인) + progress.md 스켈레톤 + spec-compact.md(카탈로그 관례 발췌본).

## 6. 변경 대상 파일

| 파일 | 편집 | 트윈 |
|------|------|------|
| `.claude/rules/moai/workflow/cross-session-messaging.md` | Rules 절 신규 [ZONE:Evolvable] [HARD] 불릿 + Anti-patterns 신규 엔트리 `Reviving a stopped teammate` | `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md` (동일 텍스트) |
| `.claude/rules/moai/core/agent-common-protocol.md` | §Background Agent Execution 기존 문단 뒤 신규 [ZONE:Evolvable] [HARD] 문단 (제목 변경 없음) | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` (동일 텍스트) |

이외 전 표면 무편집(§7, plan.md §A.4 PRESERVE 목록).

## 7. Out of Scope

### Out of Scope — t267 메커니즘 계층 (정지 티메이트 이름 주소 회수)

- 어느 계층이 정지 티메이트의 이름 주소를 회수하는지(TaskStop 주소 회수 / reject-not-revive / 로그·커밋 귀속에 부활 표면화)는 동반 카드 **t267**(queued, Class C, Tier S~M) 소관이다.
- 본 SPEC은 교리 계층만 다루며 t267의 결정을 선취·가정하지 않는다. hook/stop-goal-evaluator 가설은 t267 카드 자체가 명시적으로 배제한다.

### Out of Scope — Claude Code 런타임 동작 변경

- SendMessage의 정지 에이전트 auto-resume은 런타임 동작이다(changelog 2.1.77). moai-adk-go Go 코드로는 고칠 수 없고, 본 SPEC은 Go 변경을 포함하지 않는다.
- 런타임 부활 차단 설정(네이티브 옵션)은 현재 문서화된 바 없다(research §4 NONE-found).

### Out of Scope — 기계적·훅 강제

- PreToolUse 가드·stop-hook 판정 등 기계적 강제를 만들지 않는다. 근거: 훅이 판단할 수 있는 정지-상태 신호가 현재 없고, 강제 가능한 신호는 t267 메커니즘 결정의 하류 산물이다 — 지금 강제하면 추측을 코드화한다. Tier S 범위는 교리(REQ-001..003) + 표면화 의무(REQ-004)다.

### Out of Scope — 추가 편집 표면

- `sync-auditor.md` 세션-벡터 절 — 고려했으나 Tier S에서 제외: always-loaded `agent-common-protocol.md`가 감사 세션을 포함한 전 에이전트에 이미 도달하며, 에이전트 파일 편집은 내용 미검토 `.codex` `.toml` 미러를 범위에 끌어들인다.
- `manager-lead.md`, `kanban-dispatch.md`(+`-detail.md`), `orchestration-mode-selection.md` §C.1, `cross-session-messaging-detail.md`, `sync.md`(skill), `worktree-integration.md`, `.claude/rules/local/` dev-only 경로 — 전부 무편집. 동기: 2개 always-loaded 서식지가 이미 전 세션 도달하며, 중복 절은 예산(AC-006)과 중립성 관리면만 늘린다. sync-audit가 커버리지 부족을 판정하면 후속 카드로 재개한다.
