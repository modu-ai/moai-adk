---
id: SPEC-AGENT-PARALLEL-OPT-001
title: "Agent instruction diet + plan/run/sync parallelization maximization — Design"
version: "0.2.0"
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: ".claude/skills/moai/workflows/sync, .claude/workflows, internal/template/templates"
lifecycle: spec-anchored
tags: "agent-diet, parallelization, fan-out, workflow-wiring, template-first"
tier: L
---

## §A 설계 범위

본 문서는 Tier L 5-artifact 의무를 충족하는 설계 산출물이며, 요구사항(`spec.md`)이 **무엇을** 요구하는지가 아니라 **어떤 구조로** 성립하는지를 기술한다. 다루는 대상은 3가지다.

1. **§B — sync Phase 12 drafter/applier 데이터 흐름** (REQ-APO-024 / 024b)
2. **§C — `.claude/workflows/` SSOT 방향과 `moai update` 상호작용** (REQ-APO-069 / 073 / 077)
3. **§D — capability gate degradation 설계** (REQ-APO-011 / 063)

범위 밖: write-concurrency manifest 설계. 구 Group 1 철회로 후속 SPEC 소관이 되었다(`spec.md` §C `Out of Scope — write-concurrency rule relaxation`).

---

## §B sync Phase 12 — drafter / applier 데이터 흐름

### B.1 구조

```
                     ┌──────────────────────────────────────────┐
                     │ 오케스트레이터 (single launcher, flat)    │
                     └────────────────┬─────────────────────────┘
                                      │ 단일 턴 5× Agent() spawn (read-only)
        ┌──────────────┬──────────────┼──────────────┬──────────────┐
        ▼              ▼              ▼              ▼              ▼
   ┌─────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
   │ D1      │   │ D2       │   │ D3       │   │ D4       │   │ D5       │
   │CHANGELOG│   │README +  │   │project-  │   │SPEC-     │   │codemaps  │
   │ 초안    │   │docs-site │   │docs 초안 │   │artifacts │   │ 초안     │
   │         │   │ 초안     │   │          │   │ 초안     │   │          │
   └────┬────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘
        │             │              │              │              │
        └─────────────┴──────────────┼──────────────┴──────────────┘
                                     │ 5× 초안 텍스트 (파일 미기록)
                                     ▼
                        ┌────────────────────────────┐
                        │ manager-docs (단일 applier) │
                        │ 순차 적용 — 유일한 writer   │
                        └────────────┬───────────────┘
                                     ▼
                          작업 트리 (파일 쓰기 1주체)
```

### B.2 왜 이 구조가 write-concurrency 규칙과 무관한가

| 주체 | 도구 권한 | 파일 쓰기 |
|---|---|---|
| D1~D5 drafter | Read / Grep / Glob (read-only) | **없음** |
| manager-docs applier | Read / Write / Edit | 있음 — **동시에 1개만** |
| 오케스트레이터 | drafter 실행 중 read-only 유지 | 없음 |

동시에 쓰기를 수행하는 주체가 **어느 시점에도 1개 이하**이므로, 파일 쓰기 레이스가 구조적으로 발생할 수 없다. 현행 `agent-common-protocol.md`의 절대 금지형 규칙("does not run two write-capable agents concurrently")은 **위반되지 않는다** — write-capable 에이전트가 애초에 1개다. 이것이 REQ-APO-024b가 성립하는 근거이며, 구 Group 1 철회 이후에도 본 설계가 그대로 유효한 이유다.

### B.3 초안 전달 형태

drafter는 파일을 쓰지 않고 **초안 텍스트를 반환**한다. 반환값은 오케스트레이터 컨텍스트에 적재된 뒤 applier의 spawn 프롬프트로 주입된다.

- 초안이 큰 경우(docs-site 4-locale 등)에는 drafter가 `.moai/state/` 하위 임시 경로에 기록하고 경로만 반환하는 변형이 허용된다. 이 경로는 런타임 상태 디렉터리이며 SPEC 산출물이 아니므로 쓰기 주체 계수에 포함되지 않는다.
- 이 변형을 쓰는 경우에도 **최종 산출물(CHANGELOG / README / docs-site / project-docs / SPEC frontmatter / codemaps)에 대한 쓰기는 applier 단독**이다.

### B.4 실패 처리

| 상황 | 처리 |
|---|---|
| drafter 1개가 blocker report 반환 | 나머지 4개 초안은 유지. 오케스트레이터가 해당 항목만 재위임하거나 사용자에게 에스컬레이션 |
| drafter가 사용자 질문을 시도 | 규약 위반(REQ-APO-030). drafter 프롬프트에 blocker-report 반환 규범을 명시 |
| applier 적용 중 실패 | 기존 sync 실패 경로와 동일. 부분 적용 상태는 `progress.md`에 기록 |

### B.5 fan-out 폭

5개는 Mode 4 동시 spawn 상한(3-5)의 상단이다. 6번째 산출물군이 추가되면 상한 초과이므로 **분할하지 않고 기존 5군에 병합**한다.

### B.6 채택하지 않은 대안 — disjoint-writer

`spec.md` §C가 명시적으로 제외했다. 요약: 이득은 적용 단계까지 병렬화되는 것이나, (a) write-concurrency 규칙 완화에 실동작 의존이 생기고, (b) `manager-docs`의 frontmatter 전이 소유권(`in-progress → implemented → completed`)이 SPEC 산출물 경로와 다른 writer 경로에 걸칠 경우 Status Transition Ownership Matrix와 충돌한다. 후속 SPEC이 규칙 완화를 완료한 뒤 재검토할 수 있다.

---

## §C `.claude/workflows/` SSOT 방향과 `moai update` 상호작용

### C.1 배포 전 / 배포 후 대비

| 항목 | 배포 전 (v0.2.0 이전 상태) | 배포 후 (본 SPEC 이후) |
|---|---|---|
| 템플릿 경로 | 부재 | `internal/template/templates/.claude/workflows/` (3개 generic) |
| SSOT | 로컬 `.claude/workflows/` | **템플릿** |
| 로컬 파일의 성격 | 원본 | **파생본** (`make build` 산출) |
| `moai update` 동작 | 해당 없음 | generic 3개 **덮어쓰기** |
| `hns-*` / `harness-*` | user-owned 보존 | **변동 없음** — user-owned 보존 |

### C.2 두 집합의 분리는 접두사로 성립한다

```
.claude/workflows/
├── plan-research-fanout.js   ─┐
├── sync-audit-4dim.js         ├─ MoAI-shipped generic fan-out (template-managed, 덮어쓰기)
├── codemaps-extract.js       ─┘
├── hns-oss-docs-run.js       ─┐
└── hns-release-update-run.js ─┴─ user-owned Runner (보존, 절대 덮어쓰지 않음)
```

이 분리는 **이미 존재하는 두 개의 prefix 기반 판정**이 그대로 수행한다(`spec.md` §F.8.2 / §F.8.4 실측):

| 판정 지점 | 규칙 | generic 3개 | `hns-*` |
|---|---|---|---|
| `split_namespace_test.go` (배포 차단) | `splitHarnessAgentPrefixes` 6종 접두 매칭 | 미매치 → 배포 허용 | `hns-release-*` 매치 → 차단 |
| `update/plan/plan.go` (보존 판정) | `.claude/workflows/hns-` / `harness-` 접두 | 미매치 → template-managed | 매치 → user-owned |

따라서 **두 Go 파일 모두 수정하지 않는다**. 수정이 필요한 것은 스캔 범위가 좁은 `internal_content_leak_test.go`의 `leakTextExtensions`뿐이다(§C.4).

### C.3 사용자에게 노출되는 귀결

3개 generic 스크립트가 template-managed라는 것은 **사용자가 로컬에서 수정하면 다음 `moai update`에서 소실된다**는 뜻이다. 이 사실은 `dynamic-workflows.md` 개정문에 명시되어야 한다(REQ-APO-077). 사용자가 커스텀 fan-out을 원하면 `hns-` 접두 이름으로 별도 파일을 만드는 것이 보존되는 경로다.

### C.4 편집 순서 (Template-First)

```
1. internal/template/templates/.claude/workflows/*.js   ← 여기서 작성 + 중립화
2. make build                                            ← embedded 트리 재생성
3. .claude/workflows/*.js                                ← 파생본으로 동기화
4. diff (1) (3)                                          ← 0-diff 확인
```

역방향(로컬 선편집 → 템플릿 복사)은 금지된다. v0.2.0 계획이 로컬 라인 번호를 인용해 로컬 중립화를 먼저 지시했던 것은 이 방향 위반이었고, v0.3.0에서 교정되었다.

### C.5 중립성 스캐너 확장의 위치

`leakTextExtensions`에 `.js`를 추가하지 않으면 스캐너가 배포된 스크립트를 **열지 않는다**. 이 상태에서의 "가드 green"은 공허하다. 확장은 배포 이전에 선행되어야 하며, 확장이 실제로 동작함은 RED/GREEN 왕복으로 입증한다:

```
미중립 .js 배치 → leak 테스트 FAIL (C2가 REQ-ATR-* 매치)  ← 스캐너가 읽었다는 증거
중립화 후       → leak 테스트 PASS
```

C2 클래스가 실제 토큰을 매치하므로 이 왕복은 달성 가능하다. 다만 왕복이 입증하는 범위는 **C2 클래스에 한정**되며, AC-APO-071이 열거한 더 넓은 정규식(특히 SHA `{9,40}` — CI는 `{7,8}`)은 CI 강제 범위 밖의 수동 검사다(`spec.md` §F.8.3-a).

---

## §D capability gate degradation 설계

### D.1 gate 조건은 2항이다

```
IF   (파일 존재: .claude/workflows/<script>.js)
AND  (런타임 지원: dynamic workflow 사용 가능)
THEN launch workflow
ELSE fallback → 기존 단일 에이전트 경로
```

배포(REQ-APO-069)는 **1항만** 보장한다. 2항은 사용자 런타임 버전에 달려 있으므로 gate는 배포 이후에도 제거되지 않는다.

### D.2 3가지 환경별 동작

| 환경 | 파일 | 런타임 | 동작 |
|---|---|---|---|
| 최신 배포 사용자 | 있음 | 지원 | fan-out 실행 — 병렬 이득 |
| 구버전 런타임 사용자 | 있음 | 미지원 | fallback — 배포 이전과 동일 동작 |
| 스크립트 삭제 사용자 | 없음 | 무관 | fallback — 동일 |

세 경우 모두 **오류·경고·워크플로우 중단이 없어야 한다**(REQ-APO-063). fallback은 예외 처리가 아니라 정상 경로 중 하나다.

### D.3 fallback 경로는 기존 경로다

| 단계 | fan-out 경로 | fallback 경로 (= 현행) |
|---|---|---|
| plan Phase 2+6 | `plan-research-fanout.js` | 단일 `Explore` 리서치 |
| run Phase 13/16/17 · sync Phase 7 | `sync-audit-4dim.js` | 단일 `sync-auditor` 순차 심사 |
| codemaps | `codemaps-extract.js` (high-count 한정) | `go list -deps -json` + `go doc` 결정론적 추출 |

fallback은 새로 작성하는 코드가 아니라 **현재 동작 그대로**다. 따라서 배선 작업의 리스크는 "fallback이 깨질 위험"이 아니라 "gate 문구가 fan-out을 무조건 참조하게 잘못 쓰일 위험"에 국한된다 — AC-APO-011이 gate 문구 건수와 참조 건수의 일치를 요구하는 이유다.

### D.4 verdict 소유권은 gate와 무관하게 고정

fan-out 경로를 타든 fallback을 타든 구속력 있는 PASS/FAIL은 `plan-auditor` / `sync-auditor`가 산출한다(REQ-APO-013). 스크립트가 계산하는 집계 점수(harmonic mean 등)는 **증거**이지 verdict가 아니다. 이 불변식은 두 경로 사이에서 판정 일관성을 보장한다 — 사용자 환경에 따라 verdict 주체가 달라지지 않는다.

---

## §E 설계 수준 리스크

| # | 리스크 | 완화 |
|---|---|---|
| R1 | drafter 초안 품질이 낮아 applier가 사실상 재작성 | drafter 프롬프트에 산출물별 형식·SSOT 경로를 명시. 초안 품질은 run-phase에서 관측 |
| R2 | 5개 초안이 오케스트레이터 컨텍스트를 과점 | 큰 초안은 §B.3 경로 반환 변형 사용 |
| R3 | `moai update`가 사용자 로컬 수정을 덮어씀 | §C.3 문서화 + `hns-` 접두 대안 안내 |
| R4 | `.js` 확장자 추가가 기존 템플릿의 다른 `.js`를 새로 스캔해 예상치 못한 FAIL | 현재 템플릿 트리에 `.js`는 본 SPEC이 추가하는 3개뿐 — 부수 영향 없음(배포 전 `go test ./internal/template/...` green baseline으로 확인) |
| R5 | shipped rule 개정문이 과도하게 넓어져 user-owned 보존 계약까지 흔듦 | 개정 범위를 L80/L131 두 서술로 한정, `hns-*`/`harness-*` 보존 문구는 유지 |

---

## §F Cross-References

- `spec.md` §B.2(Group 2 재구조화), §B.5~§B.6(배포/정합성), §F.8(전제 실측)
- `plan.md` §B.1(D1 작업 7건), M1(배포 순서), M2(재구조화)
- `acceptance.md` §B 시나리오 1/5/6, §D.5~§D.6
- `research.md` §H(가드 스코프 실측)
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §C.2 — Mode 4 3-5 동시 spawn 상한(§B.5 근거)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — §B.6 disjoint-writer 배제 근거
