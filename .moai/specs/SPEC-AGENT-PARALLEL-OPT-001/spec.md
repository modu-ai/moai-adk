---
id: SPEC-AGENT-PARALLEL-OPT-001
title: "Agent instruction diet + plan/run/sync parallelization maximization"
version: "0.2.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: ".claude/agents/moai, .claude/skills/moai/workflows, .claude/rules/moai/core, internal/template/templates"
lifecycle: spec-anchored
tags: "agent-diet, parallelization, fan-out, write-concurrency, workflow-wiring, template-first"
tier: L
related_specs: [SPEC-DWF-CODEMAPS-PILOT-001, SPEC-WORKFLOW-CACHE-OPT-001, SPEC-V3R6-WORKFLOW-EFFORT-MAP-001, SPEC-SUBAGENT-NESTING-DOCTRINE-001]
partially_supersedes: [SPEC-DWF-CODEMAPS-PILOT-001]
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-25 | manager-spec | 최초 plan-phase draft. §F Ground Truth 전량 본 세션 실측(agent line count 10파일, orphan grep, template mirror 8경로, write-concurrency 3표면, run/sync/plan phase 번호). 브리프 2건 실측 정정(§F.7). Tier L(§E). |
| 0.2.0 | 2026-07-25 | manager-spec | 사용자 결정 D1/D2/D3 반영. D1=3개 fan-out 스크립트 **템플릿 배포**(REQ-069~073 신설, DWF-CODEMAPS-PILOT-001 비배포 AC 부분 supersede). D2=sync P12를 drafter+단일 적용자로 **확정**하고 M1 의존 해제. D3=SPEC-ID 마커는 **선행 grep 게이트** 후 결정. 배포 전제 3건 실측 정정(§F.8) — split_namespace 가드는 이미 prefix-scoped라 개정 불요, leak 가드는 `.js` 미스캔이라 확장 필수, `moai update` 보존목록도 prefix-scoped. |

---

## §A Context — 문제 정의

### A.1 배경

MoAI 오케스트레이션 표면은 두 축에서 동시에 비용을 지불하고 있다.

**축 1 — 에이전트 본문 과복잡도.** `.claude/agents/moai/` 10개 에이전트 본문 합계는 **2,417 라인**(§F.1 실측)이다. 이 중 상당량이 (a) 스스로 SSOT라고 선언한 규칙 파일의 내용을 본문에 다시 복제한 것, (b) `agent-authoring.md` § Prompt Craft가 명시적으로 금지한 Opus 4.6-era 방어적 스캐폴딩, (c) 동일 에이전트 안에서 2~3회 반복되는 동일 제약이다. 이 라인들은 **매 spawn마다 컨텍스트에 적재**되므로 비용은 호출 횟수에 비례해 누적된다.

**축 2 — 워크플로우 직렬 병목.** plan/run/sync 파이프라인의 다수 단계가 본질적으로 독립적인 작업을 단일 에이전트에 직렬로 통과시킨다. 동시에, 병렬 fan-out을 위해 이미 **작성·배포된 3개의 dynamic workflow 스크립트가 어느 워크플로우 문서에서도 호출되지 않는다**(§F.2 실측). 즉 병렬화 수단이 존재하는데 배선만 없다.

두 축은 하나의 근원을 공유한다. **병렬 배칭 지시가 카탈로그 전체에 단 1개 파일에만 존재한다** — `verification-batch-pattern.md` 또는 `agent-common-protocol.md § Parallel Execution`을 참조하는 에이전트 본문은 `plan-auditor.md` **단 1개**이며 나머지 9개는 참조가 전무하다(§F.3 실측). 병렬화 규범이 에이전트 계층에 도달하지 못한 채 규칙 파일 안에만 갇혀 있다.

### A.2 정책 차단 요인

`agent-common-protocol.md § Background Agent Execution`은 동시성 안전장치를 **절대 금지 형태**로 서술한다("MoAI does not run two write-capable agents concurrently"). 반면 이미 배포된 `e2e.md:251`은 **스코프 한정 형태**("never run concurrently **on overlapping scope**")를 쓰고 있다. 두 표면이 불일치하며, 절대 금지 형태가 sync Phase 12(서로 겹치지 않는 ~10개 산출물)의 병렬화를 원천 차단한다. 실제 위험은 "동시 쓰기"가 아니라 **"겹치는 경로에 대한 동시 쓰기"**다.

### A.3 목표

에이전트 본문에서 SSOT 중복·금지된 스캐폴딩·반복 제약을 제거하고(축 1), 기존 fan-out 자산을 배선하며 read-only fan-out + single-writer 패턴으로 plan/run/sync를 재구조화하고(축 2), 그 전제가 되는 write-concurrency 규칙을 안전 의도를 보존한 채 스코프 한정 형태로 정합화한다(축 3).

---

## §B Requirements (GEARS)

### B.1 Group 1 — Write-concurrency 스코프 한정 (기반 축)

#### REQ-APO-001 (Ubiquitous)
The `agent-common-protocol.md § Background Agent Execution` 절은 동시성 안전장치를 **스코프 한정 형태**로 서술 **shall** — 즉 "겹치는 스코프(overlapping scope)에 대해 두 write-capable 에이전트를 동시 실행하지 않는다"로, 절대 금지 서술을 대체한다.

#### REQ-APO-002 (Where)
**Where** 오케스트레이터가 2개 이상의 write-capable 에이전트를 동시 spawn하는 경우, 각 spawn 프롬프트는 **disjoint path manifest**(해당 에이전트가 쓸 수 있는 경로의 명시적 allow-list)를 선언 **shall**.

#### REQ-APO-003 (When)
**When** 두 개 이상의 선언된 manifest가 교집합을 가지는 것이 탐지되면, 오케스트레이터는 해당 spawn들을 동시 실행하지 **shall not** 하고 직렬화 **shall**.

#### REQ-APO-004 (Ubiquitous)
`CLAUDE.md §14` 미러 문장과 `internal/template/templates/CLAUDE.md`의 대응 문장은 REQ-APO-001과 동일한 스코프 한정 어휘를 담 **shall**, 두 파일의 해당 문장은 byte-parity를 유지 **shall**.

#### REQ-APO-005 (Ubiquitous)
본 개정은 read-only 오케스트레이터 안전장치를 원문 그대로 보존 **shall** — "orchestrator work concurrent with a write-capable agent stays read-only"는 스코프 한정 대상이 아니며 약화되지 **shall not**.

### B.2 Group 2 — 고아 fan-out 자산 배선

#### REQ-APO-010 (Ubiquitous)
`plan.md` Phase Routing Table은 `plan-research-fanout.js`를 Phase 2(Project Exploration)와 Phase 6(Deep Research)를 통합한 리서치 수행 수단으로 명시 **shall**.

#### REQ-APO-011 (Where)
**Where** `.claude/workflows/<script>.js`가 실행 환경에 존재하고 런타임이 dynamic workflow를 지원하는 경우 오케스트레이터는 해당 스크립트를 launch **shall**; **Where** 어느 한 조건이라도 부재한 경우 기존 단일 에이전트 경로로 fallback **shall** 하며 이때 기능 손실이 발생하지 **shall not**.

capability gate는 배포(REQ-APO-069) 이후에도 유지 **shall** — 배포는 파일 존재를 보장할 뿐 런타임 지원(Claude Code dynamic workflow 최소 버전)을 보장하지 않으므로 gate는 여전히 필요하다.

#### REQ-APO-012 (Ubiquitous)
`run.md` 품질 단계(Phase 13/16/17)와 `sync.md` 품질 단계(Phase 7)는 `sync-audit-4dim.js`를 4차원 증거 수집 수단으로 명시 **shall**.

#### REQ-APO-013 (Ubiquitous)
3개 workflow 스크립트는 **증거 수집 수단(evidence vehicle)**으로만 기능 **shall** — 구속력 있는 PASS/FAIL verdict 소유권은 `plan-auditor` / `sync-auditor`에 유지되며 스크립트 산출물이 verdict를 대체하지 **shall not**.

#### REQ-APO-014 (Where)
**Where** 대상 소스 패키지 수가 높은(high-count) 경우에 한해 codemaps 단계는 `codemaps-extract.js`를 아키텍처 인사이트 증강 수단으로 호출 **shall** — SPEC-DWF-CODEMAPS-PILOT-001이 확정한 "추출 대체가 아닌 증강" 스코핑을 유지 **shall**.

#### REQ-APO-015 (When)
**When** 배선이 완료되면, 3개 스크립트 각각이 최소 1개 워크플로우 문서(`.claude/skills/moai/workflows/**`)에서 참조되는 것이 grep으로 확인 **shall**.

#### REQ-APO-016 (When)
**When** 배선(REQ-APO-010/012/015)과 배포(REQ-APO-069)가 완료되면, docs-site 4-locale이 주장하는 "두 스크립트가 실제 파이프라인에 투입되어 있다"는 서술은 **참이 되며**, 그 참임이 검증 **shall** — 즉 본 요구사항은 "미검증 주장을 정정"이 아니라 **"주장을 참으로 만든 뒤 검증"**이다. 검증은 zero-orphan grep(AC-APO-015) + 배포 존재 grep(AC-APO-069)의 동시 통과로 성립 **shall**.

**When** 배선 또는 배포 중 하나라도 미완인 채 SPEC이 종료되면, 해당 서술은 실제 상태에 맞게 정정 **shall** — 미검증 주장을 잔존시키지 **shall not**.

### B.3 Group 3 — read-only fan-out + single-writer 재구조화

#### REQ-APO-020 (Ubiquitous)
plan Phase 11(Independent SPEC Review)은 복수의 read-only 심사 렌즈를 병렬 수집한 뒤 `plan-auditor` 단일 에이전트가 구속력 있는 verdict를 산출하는 구조 **shall**.

#### REQ-APO-021 (Ubiquitous)
plan Phase 10 산출물 생성은 **단일 writer(manager-spec)** 유지 **shall** 하되, 산출물 전량을 단일 턴 병렬 `Write` 호출로 생성 **shall**, 그리고 `manager-spec.md` 본문의 산출물 개수 서술은 실제 산출물 집합과 일치 **shall**.

#### REQ-APO-022 (Where)
**Where** development_mode가 `tdd`인 경우, RED 단계 테스트 초안은 복수 read-only drafter가 병렬 작성하고 단일 `manager-develop`이 적용 **shall** — 테스트 파일에 대한 동시 쓰기는 발생하지 **shall not**.

#### REQ-APO-023 (Ubiquitous)
run Phase 13 / 16 / 17의 3회 직렬 감사 패스는 1회 병렬 증거 수집 + 1회 구속력 있는 `sync-auditor` verdict로 축약 **shall** 하며, 기존 최대 3회 반복 상한은 보존 **shall**.

#### REQ-APO-024 (Ubiquitous)
sync Phase 12(Execute Document Synchronization)는 5개 산출물군(CHANGELOG / README+docs-site / project-docs / SPEC-artifacts / codemaps)에 대해 **병렬 read-only drafter**를 운용하고 **단일 `manager-docs`가 순차 적용** **shall**. 이 형태가 확정 설계이며 disjoint-writer 변형은 채택하지 **shall not**.

#### REQ-APO-024b (Ubiquitous)
REQ-APO-024의 구조는 Group 1(write-concurrency 개정) 결과와 **독립** **shall** — drafter는 전원 read-only이고 적용자는 단일이므로 동시 write가 발생하지 않는다. 따라서 M1이 지연·실패·철회되더라도 M3의 Phase 12 재구조화는 차단되지 **shall not**.

#### REQ-APO-025 (Where)
**Where** sync Phase 10 커버리지 갭이 복수 패키지에 걸쳐 있는 경우, 패키지별 테스트 생성은 병렬 fan-out으로 수행 **shall**.

#### REQ-APO-026 (When)
**When** sync Phase 1(`gate-sync-1`)과 Phase 7이 실행될 때, run Phase 15가 기록한 `moai verify` 스냅샷이 신선(키 일치 + TTL 이내)하면 해당 검사 범주를 재실행하지 **shall not** 하고 스냅샷을 증거로 인용 **shall**; 신선하지 않으면 재실행 **shall**.

#### REQ-APO-027 (Where)
**Where** MX 스캔 대상이 복수 패키지·디렉터리로 분할 가능한 경우, 스캔은 샤딩된 read-only fan-out으로 수행 **shall**.

#### REQ-APO-028 (Ubiquitous)
본 SPEC이 도입하는 모든 fan-out은 **오케스트레이터가 launch** **shall** — subagent nesting에 의존하지 **shall not** (평면 계층 유지).

#### REQ-APO-029 (Ubiquitous)
기존 HUMAN GATE는 전량 보존 **shall**: plan Decision Point 1, Implementation Kickoff Approval, `gate-sync-1`, `gate-sync-2`. 어떤 fan-out도 게이트를 우회하거나 무인 통과시키지 **shall not**.

#### REQ-APO-030 (Ubiquitous)
`AskUserQuestion` 오케스트레이터 전용 경계는 보존 **shall** — 모든 drafter/judge 서브에이전트는 사용자에게 질문하지 **shall not** 하고 구조화된 blocker report를 반환 **shall**.

### B.4 Group 4 — 에이전트 본문 다이어트

#### REQ-APO-040 (Ubiquitous)
`plan-auditor.md`는 12-field frontmatter 스키마를 본문에서 **최대 1회만** 서술 **shall** — 현재 MP-3와 FC-1..FC-12 두 곳에 존재하는 중복 열거 중 하나는 SSOT(`spec-frontmatter-schema.md`) 교차참조로 대체 **shall**.

#### REQ-APO-041 (Ubiquitous)
`manager-spec.md`의 12-field frontmatter 스키마 블록은 SSOT 교차참조로 대체 **shall**.

#### REQ-APO-042 (Ubiquitous)
`plan-auditor.md`의 M6 Chain-of-Verification 절과 그 보고 템플릿 섹션은 제거 **shall** — `agent-authoring.md § Prompt Craft`가 금지한 Opus 4.6-era 방어적 스캐폴딩에 해당한다.

#### REQ-APO-043 (Ubiquitous)
`manager-spec.md`의 SPEC-ID 사전 자가검사 프로토콜은 **실행 가능한 Bash 정규식 검사를 유지한 채** 의례적 서술(단계별 decomposition 출력 마커 강제, 예시 표, AC sub-ID 혼동 표)을 축약 **shall**.

#### REQ-APO-044 (Ubiquitous)
`manager-spec.md` Step 5 검증 체크리스트는 Step 4가 이미 서술한 제약을 재진술하지 **shall not**.

#### REQ-APO-045 (Ubiquitous)
`manager-spec.md`의 GEARS/EARS 문법 패턴 표는 `moai-workflow-spec` 스킬 교차참조로 대체 **shall**.

#### REQ-APO-046 (When)
**When** `manager-spec.md` Step 4가 산출물 병렬 생성을 지시할 때, 서술된 파일 개수는 실제 열거된 파일 개수와 일치 **shall**.

#### REQ-APO-047 (Ubiquitous)
`manager-develop.md`의 DDD / TDD 두 워크플로우는 공통 골격 1개 + 모드별 차이 서술로 통합 **shall** — 동형 워크플로우를 전문 2회 기술하지 **shall not**.

#### REQ-APO-048 (Where)
**Where** 대상 변경이 서로 독립적인 패키지에 걸쳐 있는 경우, `manager-develop.md`의 "one atomic change at a time" 제약은 패키지 내부 범위로 한정 **shall** — 독립 패키지 간 동시 진행을 금지하지 **shall not**.

#### REQ-APO-049 (Ubiquitous)
`sync-auditor.md`는 단일 scoring model과 단일 report template만 서술 **shall**.

#### REQ-APO-050 (Ubiquitous)
`manager-docs.md`에서 실제 소유 범위(CHANGELOG / README / docs-site / frontmatter 전이)와 무관한 레거시 서술(Nextra 프레임워크 설정, WCAG 접근성 점수, page-speed 지표)은 본문에서 제거 **shall**.

#### REQ-APO-051 (Where)
**Where** e2e 레시피가 현재 호스트 OS에서 실행 불가능한 경우, 해당 레시피는 에이전트 본문이 아니라 on-demand 스킬 레퍼런스에 위치 **shall**.

#### REQ-APO-052 (Ubiquitous)
`manager-git.md`의 `merge_method` 해석 규칙은 본문에서 1회만 서술 **shall**.

#### REQ-APO-053 (Ubiquitous)
`builder-harness.md`에서 `model-policy.md`를 재진술하는 블록과 이미 종료된 마이그레이션 안내는 제거 **shall**.

#### REQ-APO-054 (Ubiquitous)
다중 검증을 수행하는 모든 retained 에이전트 본문은 병렬 배칭 규범(`agent-common-protocol.md § Parallel Execution` / `verification-batch-pattern.md`)에 대한 교차참조를 1줄 보유 **shall**.

#### REQ-APO-055 (Ubiquitous)
`.claude/agents/moai/*.md` 10개 파일의 합계 라인 수는 지정된 상한 이하 **shall** 하며, 파일별 상한도 각각 충족 **shall**(§D.2 표).

### B.5 Group 5 — 불변식 및 Template-First

#### REQ-APO-060 (When)
**When** 템플릿 미러가 존재하는 파일을 수정할 때, 편집은 `internal/template/templates/` 원본에 먼저 수행하고 `make build` 후 로컬로 동기화 **shall**.

#### REQ-APO-061 (Ubiquitous)
템플릿 미러 산출물은 내부 개발 흔적(SPEC ID, REQ 토큰, 내부 날짜, commit SHA)을 포함하지 **shall not**.

#### REQ-APO-062 (Ubiquitous)
`.claude/workflows/` 하위의 **dev-only 및 user-owned harness Runner** 비배포 불변식은 유지 **shall** — `hns-*` / `harness-*` 접두 파일이 `internal/template/templates/.claude/workflows/`에 존재하지 **shall not**. 이 불변식은 3개 generic fan-out 스크립트의 배포(REQ-APO-069)와 **양립** **shall** — 두 집합은 접두사로 분리된다.

#### REQ-APO-063 (Where)
**Where** 배포 사용자 환경에 `.claude/workflows/`가 존재하지 않는 경우, 본 SPEC이 추가한 모든 참조는 graceful degradation **shall** — 오류·경고·워크플로우 중단을 유발하지 **shall not**.

#### REQ-APO-064 (When)
**When** 구현이 완료되면 `go test ./...`가 green **shall** 하고 template-neutrality CI guard가 green **shall**.

#### REQ-APO-065 (Ubiquitous)
에이전트 frontmatter의 동작 결정 필드(`name`, `description`, `tools`, `model`, `effort`, `skills`)는 변경되지 **shall not** — 본 SPEC은 본문(body) 범위 작업이다.

#### REQ-APO-066 (Ubiquitous)
archived 12개 에이전트 이름은 어떤 편집 산출물에도 신규 도입되지 **shall not**.

#### REQ-APO-067 (When)
**When** 템플릿 편집이 수행되면 `internal/template/split_namespace_test.go`와 `internal/template/internal_content_leak_test.go`가 green을 유지 **shall**.

#### REQ-APO-068 (Ubiquitous)
제거된 모든 본문 중복은 선언된 SSOT를 가리키는 교차참조로 대체 **shall** — 정보의 무성 소실(silent information loss)이 발생하지 **shall not**.

### B.6 Group 6 — dynamic workflow 스크립트 배포 (사용자 결정 D1)

#### REQ-APO-069 (Ubiquitous)
`plan-research-fanout.js`, `sync-audit-4dim.js`, `codemaps-extract.js` 3개 스크립트는 `internal/template/templates/.claude/workflows/`에 미러되어 배포 사용자에게 전달 **shall**.

#### REQ-APO-070 (Ubiquitous)
본 SPEC은 `SPEC-DWF-CODEMAPS-PILOT-001`의 **비배포 acceptance criterion을 명시적으로 supersede** **shall** — 해당 SPEC의 `grep -r "codemaps-extract" internal/template/templates/` → nothing 판정은 본 SPEC 이후 무효이며, 그 사실이 본 SPEC 산출물과 선행 SPEC 아티팩트 양쪽에 기록 **shall**. 선행 판정을 침묵 속에 위반하지 **shall not**.

#### REQ-APO-071 (Ubiquitous)
배포되는 3개 스크립트는 §25 템플릿 중립성을 충족 **shall** — 내부 SPEC ID, REQ/AC 토큰, 내부 날짜, commit SHA, 메인테이너 절대 경로가 스크립트 헤더·주석·문자열 어디에도 잔존하지 **shall not**.

#### REQ-APO-072 (When)
**When** `.js` 자산이 템플릿 트리에 최초 배포되면, 중립성 leak 스캐너의 대상 확장자 집합에 `.js`가 추가 **shall** — 스캐너가 `.js`를 읽지 않는 상태에서 통과하는 중립성 판정은 공허(vacuous)하며 증거로 인정되지 **shall not**.

#### REQ-APO-073 (Ubiquitous)
배포된 3개 스크립트는 `moai update`의 **template-managed**(덮어쓰기 가능) 집합에 속 **shall** — user-owned 보존 집합(`hns-*` / `harness-*`)에 편입되지 **shall not**. 사용자 개인 Runner Workflow의 보존 계약은 변경되지 **shall not**.

---

## §C Exclusions — 범위 제외

### Out of Scope — Go 프로덕션 코드 변경

- `internal/` 및 `pkg/` 하위 Go **프로덕션** 구현 변경은 본 SPEC 범위 밖이다. 본 SPEC은 지시문(instruction) 계층 — 에이전트 본문, 워크플로우 스킬 문서, 규칙 파일, 템플릿 자산 — 이 주 대상이다.
- `internal/template/templates/` 하위 **템플릿 자산** 편집은 Template-First 원칙상 필수이며 범위 내다.
- **유일한 예외**: `internal/template/internal_content_leak_test.go`의 `leakTextExtensions`에 `.js`를 추가하는 변경은 범위 내다(REQ-APO-072). 이것이 없으면 중립성 판정이 공허해진다(§F.8.3). 그 외 신규 CI 가드 테스트 작성은 범위 밖이다.
- `split_namespace_test.go` 가드 개정은 범위 밖이며 **불필요**하다 — 이미 prefix-scoped이다(§F.8.2). `internal/cli/update/plan/plan.go` 보존 분류 변경도 범위 밖이며 불필요하다(§F.8.4).

### Out of Scope — 에이전트 카탈로그 변경

- 11개 retained 에이전트의 추가·삭제·이름 변경은 범위 밖이다.
- 에이전트 frontmatter의 `model` / `effort` / `tools` 튜닝은 범위 밖 — 해당 축은 `moai model profile` 3-tier 프로파일 소관이다.
- `sync-auditor`의 read-only nesting 파일럿(`Agent` tool 보유) 설정 변경은 범위 밖이며, 본 SPEC의 모든 fan-out은 오케스트레이터 launch로 설계된다(REQ-APO-028).

### Out of Scope — 게이트 및 사용자 상호작용 의미 변경

- HUMAN GATE의 발화 조건·순서·필수성 변경은 범위 밖이다. 병렬화는 게이트 사이 구간의 실행 형태만 바꾼다.
- `AskUserQuestion` 채널 독점 규칙 및 orchestrator-subagent 비대칭 경계의 변경은 범위 밖이다.
- plan-auditor / sync-auditor의 verdict 소유권 이전은 범위 밖이다(REQ-APO-013).

### Out of Scope — 성능 실측 및 벤치마크

- 병렬화로 인한 wall-time 단축량의 정량 측정(A/B 벤치마크)은 범위 밖이다. 본 SPEC의 AC는 구조적 검증(grep / line count / 참조 도달성)만 요구한다.
- 토큰 절감량의 정량 측정 역시 범위 밖이다. 라인 수 감소는 대리 지표(proxy)로만 사용한다.

### Out of Scope — 신규 workflow 스크립트 작성

- 새 `.js` dynamic workflow 스크립트를 작성하지 않는다. 본 SPEC은 **이미 존재하는 3개** 스크립트를 배선하고 배포한다.
- 3개 스크립트 내부 **로직**의 기능 변경은 범위 밖이다. 배포를 위한 §25 중립화(헤더·주석의 내부 토큰 제거, REQ-APO-071)는 범위 내이며 실행 로직을 바꾸지 않는다.
- `hns-oss-docs-run.js` / `hns-release-update-run.js` 등 harness Runner의 배포는 범위 밖이다(비배포 유지, REQ-APO-062).

### Out of Scope — disjoint-writer 병렬 쓰기 변형

- sync Phase 12의 경로 소유 writer 병렬 실행 변형은 **채택하지 않는다**(사용자 결정 D2). 확정 설계는 read-only drafter + 단일 적용자다(REQ-APO-024).
- 해당 변형은 **문서화된 향후 선택지**로만 보존한다 — 후속 SPEC이 Group 1 규칙 정착 이후 재검토할 수 있으나 본 SPEC의 산출물·AC·마일스톤에 포함되지 않는다.
- Group 1(write-concurrency 개정)은 그 자체로 독립 가치를 가지므로 유지되나, Phase 12는 Group 1 결과에 의존하지 않는다(REQ-APO-024b).

### Out of Scope — docs-site 본문 재작성

- docs-site 4-locale 콘텐츠의 전면 개편은 범위 밖이다. REQ-APO-016은 오직 "파이프라인 투입" 주장 1건의 진위 정합화만 요구한다.

---

## §D AC Matrix

### D.1 REQ → AC 매핑

| REQ 그룹 | REQ 범위 | AC 범위 | 검증 성격 |
|---|---|---|---|
| Group 1 — write-concurrency | REQ-APO-001..005 | AC-APO-001..005 | grep(스코프 한정 어휘) + byte-parity diff |
| Group 2 — fan-out 배선 | REQ-APO-010..016 | AC-APO-010..016 | zero-orphan grep + fallback 문구 grep |
| Group 3 — 재구조화 | REQ-APO-020..030 (024b 포함) | AC-APO-020..030 (024b 포함) | 단계 서술 grep + 게이트 보존 grep + M1 독립성 서술 |
| Group 4 — 본문 다이어트 | REQ-APO-040..055 | AC-APO-040..055 | 중복 카운트 grep + line-count 상한 |
| Group 5 — 불변식 | REQ-APO-060..068 | AC-APO-060..068 | 미러 diff + CI green + 부재 grep |
| Group 6 — 배포 (D1) | REQ-APO-069..073 | AC-APO-069..073 | 배포 존재 grep + supersession 기록 + 중립성(비공허) + hns 차단 유지 + update 분류 |

전체 AC 열거·판정 명령·MUST/SHOULD 등급은 `acceptance.md` §D에 있다.

### D.2 라인 수 상한 (REQ-APO-055 판정 기준)

| 파일 | 현재(실측) | 상한 | 근거 |
|---|---:|---:|---|
| `plan-auditor.md` | 505 | **340** | FC-1..12 중복 열거 + M6 CoVe + 보고 템플릿 CoVe 섹션 제거 |
| `manager-spec.md` | 317 | **230** | frontmatter 스키마 블록 + ID 자가검사 의례 + GEARS 표 + Step5 중복 제거 |
| `manager-develop.md` | 311 | **240** | DDD/TDD 동형 워크플로우 통합 |
| `sync-auditor.md` | 221 | **150** | scoring model 2→1, report template 2→1 |
| `manager-git.md` | 211 | **190** | merge_method 3회→1회 |
| `manager-design.md` | 201 | **205** | 다이어트 대상 아님(REQ-APO-054 교차참조 1줄 허용) |
| `builder-harness.md` | 195 | **170** | model-policy 재진술 + stale 마이그레이션 안내 제거 |
| `e2e-tester.md` | 182 | **150** | 비-호스트 OS 레시피 스킬 레퍼런스 이관 |
| `manager-docs.md` | 167 | **120** | Nextra/WCAG/page-speed 레거시 제거 |
| `super-advisor.md` | 107 | **112** | 다이어트 대상 아님(교차참조 1줄 허용) |
| **합계** | **2,417** | **≤ 1,907** | 최소 21% 감축 |

상한은 각각 **이하(≤)** 판정이며, 합계 상한은 개별 상한의 합이다. 개별 파일이 상한보다 더 줄어드는 것은 허용된다(단 REQ-APO-068 정보 무성 소실 금지가 우선한다).

---

## §E Tier 판정 — L

Tier L로 판정한다. 근거:

- **표면 수**: 에이전트 본문 10 + 워크플로우 스킬 문서 3(plan/run/sync) + 하위 스킬 문서 최소 4(`run/phase-execution.md`, `run/task-decomposition.md`, `sync/doc-execution.md`, `sync/quality-gates-quality.md`) + 정규 규칙 2(`agent-common-protocol.md`, `CLAUDE.md`) + 각각의 템플릿 미러. 편집 파일 수는 **30개 이상**이다.
- **도메인 수**: 에이전트 정의 / 워크플로우 스킬 / 정규 규칙 / 템플릿 배포 / docs-site 콘텐츠 — 5개 도메인.
- **정책 위험**: Group 1은 [HARD] 안전 규칙의 의미를 바꾼다. 문구 오작성 시 파일 쓰기 레이스를 허용하게 된다 — 최고 심사 밀도가 필요하다.
- **불변식 밀도**: mirror byte-parity, 템플릿 중립성, 비배포 불변식, 게이트 보존, verdict 소유권 보존 — 5종 불변식이 동시에 걸린다.

Tier M은 30+ 파일 다중 미러 표면과 [HARD] 정책 개정을 과소 커버한다.

---

## §F Ground Truth — 본 세션 실측

모든 수치는 본 plan-phase 세션에서 직접 명령을 실행해 관측한 것이다. 전임 보고서·기억에서 이월한 값은 없다.

### F.1 에이전트 본문 라인 수

`wc -l .claude/agents/moai/*.md` 관측: plan-auditor 505 / manager-spec 317 / manager-develop 311 / sync-auditor 221 / manager-git 211 / manager-design 201 / builder-harness 195 / e2e-tester 182 / manager-docs 167 / super-advisor 107, 합계 **2,417**.

### F.2 고아 fan-out 자산

`.claude/workflows/` 실재 파일: `plan-research-fanout.js`, `sync-audit-4dim.js`, `codemaps-extract.js`, `hns-oss-docs-run.js`, `hns-release-update-run.js`.

`.claude/skills/moai/` 하위에서 앞의 3개 스크립트명을 검색한 결과 **0건**. 즉 plan/run/sync/codemaps 워크플로우 문서 어디에서도 호출되지 않는다.

### F.3 병렬 배칭 교차참조 분포

`grep -ln "verification-batch-pattern\|Parallel Execution" .claude/agents/moai/*.md` 관측: **`plan-auditor.md` 1개 파일**만 매치. 나머지 9개 에이전트는 병렬 배칭 규범 참조 전무.

### F.4 write-concurrency 3표면

- `agent-common-protocol.md` L191 / L193 / L198 — 절대 금지 형태.
- `CLAUDE.md` L250 및 `internal/template/templates/CLAUDE.md` L250 — 동일 절대 금지 문장(byte-parity 상태).
- `.claude/skills/moai/workflows/e2e.md` L251 — **이미 스코프 한정 형태** ("on overlapping scope").

### F.5 템플릿 미러 인벤토리

미러 존재: `.claude/agents/moai`, `.claude/skills/moai/workflows`(+ `run`/`sync`/`plan` 하위), `.claude/rules/moai/core`, `.claude/rules/moai/workflow`, `CLAUDE.md`.
미러 **부재**: `.claude/workflows/` (3개 `.js` 포함).

### F.6 단계 번호 실측

- plan: Phase 2 Project Exploration / Phase 6 Deep Research / Phase 10 SPEC Document Creation / Phase 11 Independent SPEC Review.
- run: Phase 13 Quality Validation / Phase 15 Pre-Review Quality Gate(`moai verify` 스냅샷 record 지점) / Phase 16 Active Quality Evaluation / Phase 17 TRUST 5 Static Verification / Phase 9 Pre-Implementation MX Context Scan / Phase 18 MX Tag Update.
- sync: Phase 1 Pre-Sync Quality Gate(`gate-sync-1`) / Phase 7 Quality Check / Phase 10 Coverage Analysis / Phase 11 Analysis + `gate-sync-2` / Phase 12 Execute Document Synchronization(Step 2.2 CHANGELOG·README, Step 2.2.5 project docs + codemaps, Step 2.3 post-sync quality, Step 2.4 SPEC status, Step 2.4.1 issue sync).

### F.7 브리프 대비 실측 정정 2건

**정정 1 — "orphaned = zero references"는 부분적으로 거짓.** 3개 스크립트는 워크플로우 문서에서는 참조 0건이 맞으나, `docs-site/content/{en,ko,ja,zh}/claude-code/agentic/workflows.md:120`이 4개 로케일 전부에서 두 스크립트를 "실제 파이프라인에 투입"했다고 서술하고 있으며, `dynamic-workflows.md:105`는 `codemaps-extract.js`를 canonical worked example로 문서화하고 있다. 따라서 정확한 진술은 **"워크플로우 호출 경로에서 고아이며, 동시에 공개 문서가 배선되었다고 주장하는 미검증 클레임이 존재한다"**이다. 이 사실이 REQ-APO-016을 신설하게 했다.

**정정 2 — `.claude/workflows/`는 템플릿 미러가 없다.** 브리프는 "어떤 파일이 미러를 갖는지 확인하고 스코프하라"고만 지시했으나, 실측 결과 3개 `.js`는 의도적 비배포이며 SPEC-DWF-CODEMAPS-PILOT-001이 이를 명시 AC로 고정해 두었다. 반면 배선 대상인 `plan.md`/`run.md`/`sync.md`는 **미러 대상**이다. 따라서 배선은 반드시 capability-gated 형태(REQ-APO-011/063)여야 하며, 무조건적 참조는 배포 사용자에게 존재하지 않는 파일을 가리키게 된다. 이 비대칭은 사용자 결정 D1(배포)로 해소되었으나 **capability gate 자체는 유지**된다 — 파일 배포가 런타임 지원까지 보장하지는 않기 때문이다(§F.8.4).

---

## §F.8 사용자 결정 D1/D2/D3 반영 및 배포 전제 실측 정정

### F.8.1 결정 요약

| 결정 | 내용 | 반영 |
|---|---|---|
| D1 | 3개 fan-out 스크립트를 **템플릿 배포** | Group 6 신설(REQ-APO-069..073), REQ-APO-062 재정의 |
| D2 | sync Phase 12 = **read-only drafter + 단일 적용자** 확정, disjoint-writer 불채택 | REQ-APO-024 확정형, REQ-APO-024b(M1 독립성) 신설 |
| D3 | SPEC-ID 마커 축약은 **선행 grep 게이트 후** 결정 | `plan.md` M4 첫 작업으로 게이트화, REQ-APO-043 조건부 |

### F.8.2 정정 1 — split_namespace 가드는 이미 prefix-scoped이며 개정 불요

`internal/template/split_namespace_test.go` L93-104는 `.claude/workflows/*.js` 전체를 차단하지 **않는다**. 실제 차단 조건은 `splitHarnessAgentPrefixes`(`harness-release-update`, `harness-github`, `harness-release`, `hns-release-update`, `hns-github`, `hns-release`) 중 하나로 시작하는 basename이다. 5개 스크립트명에 대해 접두사 매칭을 실행한 결과:

```
plan-research-fanout.js          NOT blocked
sync-audit-4dim.js               NOT blocked
codemaps-extract.js              NOT blocked
hns-oss-docs-run.js              NOT blocked (user-owned, 의도적)
hns-release-update-run.js        BLOCKED (prefix=hns-release)
```

즉 가드는 이미 정확히 원하는 분리를 수행하고 있다 — **가드 개정·축소 작업은 필요하지 않으며**, 요구되는 것은 배포 이후에도 `hns-*` 차단이 유지됨을 확인하는 **불변식 단언**(AC-APO-072b)뿐이다. 존재하지 않는 차단을 전제로 가드를 수정하면 오히려 dev-only 보호를 약화시킨다.

### F.8.3 정정 2 — 중립성 leak 가드는 `.js`를 스캔하지 않는다 (공허 통과 위험)

`internal/template/internal_content_leak_test.go`의 `leakTextExtensions`는 `.md` / `.tmpl` / `.yaml` / `.yml` / `.sh` / `.json` 6종이며 **`.js`가 없다**. 따라서 3개 스크립트를 템플릿에 추가한 뒤 "중립성 가드 green"을 근거로 삼으면 **가드가 파일을 읽지도 않은 채 통과한 공허 판정**이 된다.

실제 중립성 위반은 존재한다:

| 파일 | 라인 | 위반 |
|---|---|---|
| `codemaps-extract.js` (62줄) | — | **0건 — 이미 중립** |
| `plan-research-fanout.js` (132줄) | 35-36, 54 | `REQ-ATR-018/019/020`, `AC-ATR-023/024/025`, `design.md §D`, `acceptance.md` 내부 참조 |
| `sync-audit-4dim.js` (173줄) | 37-38, 42 | `REQ-ATR-015/016/017`, `AC-ATR-020/021/022`, 예시 문자열 `spec_id: "SPEC-FOO-001"`(C1 정규식 매칭 대상) |

따라서 REQ-APO-072(`.js` 확장자 추가)는 REQ-APO-071(중립화) 판정이 유효해지기 위한 **선행 조건**이다. 이것이 D1이 실제로 요구하는 유일한 Go 변경이며, 예상되었던 "가드 축소"와는 반대 방향이다.

### F.8.4 정정 3 — `moai update` 보존 목록도 prefix-scoped이며 배포와 양립

`internal/cli/update/plan/plan.go` L135-145 / L189-196은 `.claude/workflows/hns-` 및 `.claude/workflows/harness-` 접두 경로만 user-owned로 분류한다. 3개 generic fan-out은 이 집합에 속하지 않으므로 자동으로 **template-managed**(덮어쓰기 가능)가 되며, 이는 배포 자산에 요구되는 정확한 의미다. 보존 계약 변경은 필요하지 않다(REQ-APO-073은 이 상태의 유지를 단언한다).

또한 `internal/template/catalog.yaml`에는 `.claude/workflows` 항목이 존재하지 않는다(유일한 "workflows" 매치는 스킬 설명 문자열). 배포는 embedded 트리의 generic FS walk로 이루어지므로 catalog 등록은 불필요하다.

### F.8.5 capability gate 존속 근거

배포는 **파일 존재**를 보장하지만 **런타임 지원**을 보장하지 않는다. dynamic workflow 실행은 Claude Code 최소 버전 요구가 있으며, 구버전 런타임의 사용자는 파일을 받고도 실행할 수 없다. 따라서 REQ-APO-011의 capability gate는 배포 이후에도 유지되며, gate 조건이 "파일 존재"에서 "파일 존재 AND 런타임 지원"으로 확장된다.

---

## §G Cross-References

- `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution / § Parallel Execution — Group 1 개정 대상 및 REQ-APO-054 참조 대상 SSOT.
- `.claude/rules/moai/workflow/verification-batch-pattern.md` — 병렬 배칭 근거 및 클래스 분류.
- `.claude/rules/moai/development/agent-authoring.md` § Prompt Craft — REQ-APO-042 금지 근거.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — REQ-APO-040/041 교차참조 대상 SSOT.
- `.claude/rules/moai/workflow/dynamic-workflows.md` — workflow 원시 및 `codemaps-extract.js` worked example.
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §C.2 — Mode 4 동시 spawn 3-5 상한(REQ-APO-028 fan-out 폭 제약).
- `SPEC-DWF-CODEMAPS-PILOT-001` — `codemaps-extract.js` 스코핑 선례. 본 SPEC이 그 **비배포 acceptance criterion을 부분 supersede**한다(REQ-APO-070); high-count 증강 스코핑은 유지된다(REQ-APO-014).
- `internal/template/split_namespace_test.go` — `hns-*` / `harness-*` Runner 차단 가드(prefix-scoped, 개정 불요 — §F.8.2).
- `internal/template/internal_content_leak_test.go` — `leakTextExtensions`에 `.js` 추가 대상(REQ-APO-072, §F.8.3).
- `internal/cli/update/plan/plan.go` — user-owned 보존 분류(prefix-scoped, 변경 불요 — §F.8.4).
- `SPEC-WORKFLOW-CACHE-OPT-001` — `sync-audit-4dim` 병렬 심사 선례 인용.
- `CLAUDE.local.md` §2 / §25 — Template-First 및 템플릿 중립성.
