# progress.md — SPEC-TODO-CLAIM-LEASE-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_phase:
  card: t1342
  spec: SPEC-TODO-CLAIM-LEASE-001
  tier: M
  artifacts: [spec.md, plan.md, acceptance.md, spec-compact.md, progress.md]
  dp1_decisions: [C1, C2, C4, C5, C6]
  dp1_approved: 2026-09-29
  research_evidence:
    - .moai/state/plan-research-t1342/spec-proposal.md
    - .moai/state/plan-research-t1342/research.md
    - .moai/state/plan-research-t1342/per-lens-reports.md
  status: draft
  plan_complete_at: 2026-09-29T22:03:37+0900
  plan_status: audit-ready
  plan_audit:
    iteration1: "FAIL 0.60 — .moai/reports/plan-audit/SPEC-TODO-CLAIM-LEASE-001-review-1.md (D1-D10)"
    iteration2: "PASS 0.86 — .moai/reports/plan-audit/SPEC-TODO-CLAIM-LEASE-001-review-2.md (must-pass 6/6, D1-D10 해소 재검증)"
    threshold: "0.80 (Tier M standard)"
    review_lenses: "FO-PLAN-2 4렌스 — .moai/state/plan-research-t1342/lens-{1-frontmatter,2-testability,3-scope,4-commands}.md"
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F.1 Plan-phase 기록 (Plan-Phase Record)

### 리서치 팬아웃 출처 (Research Fan-out Provenance)

- **팬아웃**: FO-PLAN-1 — 4개 읽기 전용 렌즈 탐색 + 1개 신디사이저. 렌즈: `codebase-precedent`, `constraints-risks`, `prior-SPEC-memory`, `queue-consumer-inventory`. 4렌즈 + 신디사이저 전부 정상 반환.
- **run wf id**: 세션 상태 파일에서 회수 불가(기록 없음) — 증거는 디스크 아티팩트 3건(`.moai/state/plan-research-t1342/{research.md, per-lens-reports.md, spec-proposal.md}`, 2026-09-29 20:55–21:03 작성)으로 대신한다. 미회수 항목을 Gaps로 명시.
- **manager-spec 제안**: spec-proposal.md (Phase 8 산출, Phase 10 이전) — REQ-TCL-001..015, AC 스케치 AC-TCL-001..010, M1-M5 계획, MX 계획 입력 포함.

### DP1 운영자 결정 (타임스탬프 포함)

- **승인 일자**: 2026-09-29 (구속력 있는 운영자 결정)
- C1 — 마이그레이션: additive `ensureColumn`, schema-stamp bump 없음("2" 유지); freeze 테스트 재기록 필요
- C2 — 클레임 원자성: Mutate 콜백 내 freshly-loaded 레코드 상태 술어 (flock = CAS 원시체); version 컬럼·Mutate 밖 엔진 접근 없음
- C4 — 파싱 불가 만료: EXPIRED (factory 기본값)
- C5 — 반환 감사: 사람이 읽는 claim/reclaim 출력 라인(id + 이전 보유자); events 테이블 없음
- C6 — `--lane <label>`: 운영자/리드 공급 레인 레이블; REQ-SD-015 거부 의미론 불변 — 레인 셀프 클레임 승인(carve)은 t1338 소유. (iteration-2 정정: 본 SPEC의 가드 변경은 flag-form guard extension inside todoRefuseLaneMutation이며 `factoryLaneRefusal()` 참 보유 시 `--lane`형도 기존 거부 텍스트로 거부)

### Phase 4 모드

- **serial** — 팬아웃 리서치는 별도 수행, SPEC 저작은 단일 시리얼 경로.

### Clarity-skip 사유

- 디스패치 및 카드 텍스트에 ≥5개 기술 키워드(`BacklogStore.Mutate`, `ensureColumn`, CAS, flock, `lease_expires_at`, REQ-SD-015) — Context-First Discovery Socratic 인터뷰 예외(기술 지시 완비) 적용, 클래리피케이션 라운드 스킵.

### Tier 산출 근거

- **dispatch 타이틀 "Tier M"** 기준 — Tier M 아티팩트 세트(spec.md + plan.md + acceptance.md + progress.md)에 dispatch가 명시 지정한 `spec-compact.md`를 추가 산출.

### 미회수/미검증 항목 (Gaps)

- FO-PLAN-1 run wf id — 상태 파일에서 회수 불가 (위 기술).
- 구버전 바이너리 쓰기의 임대 데이터 유실 및 Mutate 밖 CAS clobber는 코드 파생 판단으로, 런타임 재현 미실행 (research.md §8 고지 분).

### Iteration-2 수정 기록 (plan-audit review-1 대응)

- plan-audit iteration 1: FAIL 0.60 (기준 0.80) — `.moai/reports/plan-audit/SPEC-TODO-CLAIM-LEASE-001-review-1.md`.
- D1-D10 전건 수정. 결정적 2건의 구속 방향 반영: D2+D3 — REQ-TCL-013에 `factoryLaneRefusal()` 참 보유 시 `--lane`형도 기존 거부 텍스트로 거부하는 절 추가 + "carve" 용어를 "flag-form guard extension inside todoRefuseLaneMutation; REQ-SD-015 bare-refusal semantics unchanged"로 전면 치환(C6·REQ-013·Out of Scope·spec §E·plan M3 상호 일치). D1 — 모든 AC observable을 실존 테스트(9개 이름 트리 검증 완료) 또는 [NEW]+정확한 함수명으로 재고정.
- 기타: D4(신규 AC-TCL-011 live-lease 가드 + unpick carve-out 기존 가드 인용), D5(신규 AC-TCL-012 human 표면 + REQ-010 human-vs-JSON 재조정), D6(경합 AC에 카드당 N≥2·pre/post 전체 튜플 비교 검출기·raced byte-identity 고정), D7(claim-family 열거·cleared-field 집합·seq 순서·renew-expired observable), D8(C3을 DP1 결정이 아닌 연구 발견으로 재표기), D9(spec §E delta 마커가 plan [MODIFY] 표면 전체 커버 — todo_disclosure.go·golden/concurrency/migration 테스트 포함), D10(backlog_migrate.go 인용을 writeArchive ~:337-401/:381과 writeRecordArchive :422-515/:482 이원 범위로 정정, REQ-006 Event-driven 재표기, todo.go :303 정정).

## §F Phase 4 Mode Selection (run-phase)

- **Input parameters**: tier M · scope ~12 파일 (kanban 스키마/스토어 + cli verb/MCP + 테스트 6파일) · 도메인 2 (internal/kanban, internal/cli) · 언어 혼합 Go+markdown 없음(순 Go) · 병렬 이익 LOW (coding-heavy) · Agent Teams 사전 요건 미요청.
- **Mode evaluation**: direct — 아님(다중 파일 의미 변경) / serial — **selected** / fanout — 아님(coding-heavy, Anthropic 병렬화 경고) / sweep — 아님(기계적 균일 변환 아님) / agent-team — 미요청.
- **Decision**: serial
- **Justification**: 단일 도메인 밀착 코딩 작업(스키마→스토어→CLI→MCP→테스트가 순차 의존)이라 단일 manager-develop 스폰이 마일스톤 M1-M5를 순서로 수행하는 것이 병렬 팬아웃보다 낫다(Anthropic coding-task caveat). 팬아웃 리서치는 plan 단계에서 이미 별도 수행됐다.
- **Run-gate skip 기록 (Phase 1)**: plan-audit PASS 0.86 ≥ Tier M 임계 0.80 + plan-artifact 해시 불변(review-2 이후 spec/plan/acceptance 무변경 — progress.md는 해시 대상 아님) + depends_on SPEC-TODO-RUNTIME-STORE-001 completed → 3조건 충족으로 Phase 1 재실행 스킵.
- **진행 축**: 운영자가 Kickoff에서 자율(골 무장) 선택 — 기계 종료 조건 무장됨(세션 ed6e9938…, max_turns 30).
