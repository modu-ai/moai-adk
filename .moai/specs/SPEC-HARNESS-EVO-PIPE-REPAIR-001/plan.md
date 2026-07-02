# SPEC-HARNESS-EVO-PIPE-REPAIR-001 — Implementation Plan

Tier: **M** (6 milestones) · development_mode: **tdd** (quality.yaml 실측) · status: draft

---

## §A 컨텍스트

### §A.1 Epic Harness-Evolution 개요 (4 SPECs, 순서 1→2→3→4)

| # | SPEC ID | 범위 | 상태 |
|---|---------|------|------|
| 1 | **SPEC-HARNESS-EVO-PIPE-REPAIR-001 (본 SPEC)** | 파이프 수리(어휘/스키마/자동화) + v4 스모크 게이트 + 디스패처/문서 드리프트 수리 | draft |
| 2 | SPEC-HARNESS-EVO-RUN-REPORT-001 | 실행→학습 배선: manifest `learning` 블록, Runner return-schema `findings`, specialist 필수 "improvement findings" 최종 단계, post-run findings 수집 → 즉시 AskUserQuestion push(현행 pull-only apply 대체), `learner.go` confidence 하드코딩 1.0의 outcome 이벤트 기반 실측화 | 미작성 |
| 3 | SPEC-HARNESS-EVO-WRITE-SURFACE-001 | frozen_guard allowlist 단계적 확장(M1 manifest-first — 기승인 edit-verb 갱신 경로 활용 → M2 `.claude/commands/harness/` + `.claude/workflows/harness-*.js` + specialists; snapshot+rollback+regression-gate 의무), LOOP-CLOSURE C1 헌법 제약(auto_apply:false per-item 승인)의 티어별 표면 자율로의 명시적 supersede/amendment(저위험 표면=auto-apply+세션말 배치 보고, 고위험=per-item human gate), harness-namespace-doctrine.md "legitimate learning-loop writes" 조항(§24는 현재 moai-update 보존만 규율 — evolution-write 소유권 미정의) | 미작성 |
| 4 | SPEC-HARNESS-EVO-REQ-ARTIFACT-001 | manifest `source_request` → 구조화 요구사항 스키마(domain/goal/constraints/scope + AC + Discovery 응답 기록, 하네스 dir 영속화, 실행 시 재참조, drift 감지 → 재-Discovery), Builder 개선, 레거시 5-layer marker 경로 retire(/moai project meta-harness route, archived agent 이름 방출 layer5 scaffold), harness-delivery-strategy.md Model B rejection의 v4 command-as-harness 현실에 의한 supersede 선언 | 미작성 |

사용자 결정 기록(AFK — 권장 옵션 채택, **수정 가능(revisable)으로 기록**): ① 티어별 표면 자율(목표 상태), ② 단계적 write-surface 개방(M1 manifest-first → M2 full), ③ Epic + 4 split SPECs, ④ cleanup 4건 전부 포함.

### §A.2 8-표면 분석 요약

- R1(요구사항 유도)/R2(슬래시 커맨드 전달): v4 Builder(Context-First Discovery + 4-phase ANALYZE/PLAN/GENERATE/ACTIVATE)로 대부분 구현
- R3(end-to-end 실행): 관례(convention) 의존 — command→Runner→manifest→agent 참조가 어느 지점에서도 기계 검증되지 않음, live E2E 0건 → 본 SPEC의 스모크 게이트가 담당
- R4(재귀 자기개선): 6개 끊어진 링크로 구조적 0 커버리지 → 그중 파이프 계층 3개(B1 어휘, B2 스키마, B3 자동화)를 본 SPEC이 수리; 나머지(실행→학습 배선, write-surface, 요구사항 아티팩트)는 SPEC-2/3/4

---

## §B Known Issues (검증 앵커 요약)

| # | 결함 | 앵커 (2026-07-02 실측) | REQ |
|---|------|------------------------|-----|
| B1 | tier 어휘 불일치 → 제안 생성 0 | types.go:217-245 vs mapper.go:41-44; 실데이터 `to_tier: auto_update/observation` | REQ-HEP-001 |
| B2 | pattern_key 스키마 불일치 (prefix + 빈 segment) | mapper.go:36 vs learner.go:98-99 + EventType enum; 실데이터 `user_prompt::` | REQ-HEP-002 |
| B3 | classify 수동 전용 + generic observer 미등록 | hook.go:144-162; harness.md:135-136; settings.json/tmpl grep | REQ-HEP-003/004 |
| B4 | v4 참조 무결성 게이트 부재 | v4manifest/validate.go(단일 파일만), v4lifecycle.go, doctor_harness.go(legacy 5-layer만) | REQ-HEP-005/006 |
| B5 | exemplar Runner 부패 (wrong path + 비실존 agent) | harness-release-update-run.js:31,:56; 실존: release-update/manifest.json, harness-release-update-specialist.md | REQ-HEP-007 |
| B6 | 가드 4-verb 고착 / help 고착 / manifest 3-way 분기 | build-entry.md:53-55; commands/moai/harness.md:3; SKILL.md:70,:195-216; harness-builder.md:257 vs v4lifecycle.go:33-34 | REQ-HEP-008/009/010 |
| B7 | FROZEN·rate-limit 문서-코드 모순 | frozen_guard.go:21-37 vs harness.md §2.2 + learner SKILL.md; harness.yaml:119-121 + cli/harness.go:486 vs harness.md:91,:173 | REQ-HEP-011/012 |

---

## §C Pre-flight

1. `git fetch origin main` + divergence 확인 (agent-common-protocol § Pre-Spawn Sync Check)
2. `go build ./...` + `go test ./internal/harness/... ./internal/cli/...` 그린 베이스라인 확보
3. B1/B2 앵커를 content-token으로 재고정 (`actionableTiers`, `actionablePatternRE`, `buildPatternKey`, `Tier.String`)
4. baseline 재측정: `wc -l .moai/harness/usage-log.jsonl`, tier-promotions.jsonl 레코드 수, `moai hook harness-classify` 수동 1회 실행 결과 기록 (progress.md §E.2)

---

## §D Constraints + 설계 결정 (방향 권고 — run-phase에서 검증 후 확정)

- **D1 (어휘 정렬 방향)**: mapper의 `actionableTiers`를 tier 어휘 `{rule, auto_update}`로 재작성한다 (ToTier 방출부를 바꾸지 않음). 근거: `Tier` enum은 `@MX:ANCHOR` fan_in≥3 (learner/applier)의 SSOT이고, ToTier 방출 어휘 변경은 기존 tier-promotions.jsonl 데이터와의 호환을 깨뜨림. 채택 임계: tier ∈ {rule(5+), auto_update(10+)} = proposal-eligible; {observation, heuristic}는 pre-actionable 유지.
- **D2 (스키마 정렬 방향)**: `actionablePatternRE`의 수기 prefix 목록을 제거하고 EventType enum SSOT에서 파생한다 (prefix ∈ EventType values; subject/context_hash 빈 문자열 허용). `apply_outcome`은 pattern_key를 갖지 않으므로(cluster.go 주석, REQ-OBL-005) 파생 집합에서 제외. 실질 필터는 tier + confidence가 담당 — regex는 형식 검증만.
- **D3 (classify 배선 방식)**: 후보 (a) 기존 `runHarnessObserveStop`(Stop 경로 Go 핸들러)이 observe 기록 후 classify를 연쇄 호출 — settings 변경 불요, 훅 5s 예산 내, fail-open 용이 (**권고**); 후보 (b) settings.json에 별도 Stop hook 항목 등록 — 훅 수 증가, template 변경 필요. run-phase에서 (a)의 시간 예산 실측 후 확정.
- **D4 (smoke gate 배치)**: 신규 `moai harness doctor` subcommand — `internal/cli/harness/`의 v4lifecycle 스캔(list의 command↔manifest 조인) 재사용 + `v4manifest` validate 확장 + Runner 정적 파싱(manifest 경로 상수/agent 이름 참조 추출은 정규식 기반 heuristic으로 시작, JS 파싱 도입 금지 — Enforce Simplicity). 기존 `moai doctor`의 legacy 5-layer check와 별개 유지.
- **Template-First**: 문서 표면 수정은 live + `internal/template/templates/` mirror 동시 편집 → `make build`. mirror 실측 확인 완료(8개 표면 전부 MIRRORED). Exemplar Runner는 **dev-only 로컬 전용 — 절대 template 반입 금지**.
- **§25 중립성**: template mirror에 SPEC ID / REQ-HEP 토큰 / 감사 인용 기입 금지. 정정 문구는 generic prose로 작성 (예: rate-limit 정정은 "per harness.yaml rate_limit" 형태).
- **TDD**: development_mode=tdd — M1/M2/M3는 RED(실데이터 fixture 재현) → GREEN → REFACTOR 순.
- **하위 호환**: mapper 수리는 기존 reader tolerance(malformed 라인 skip)를 보존; observe 훅 등록은 기존 3종 wrapper와 공존(hook.go §A.3 cohabitation contract 준수).

---

## §E Self-Verification (run-phase 검증 배치 명세)

run-phase 완료 시 오케스트레이터/manager-develop이 단일 턴 병렬 배치로 실행:

1. `go test ./...` (full suite)
2. `go test -coverprofile=cover.out ./internal/harness/... ./internal/cli/...` — touched pkg ≥85% (cli/hook 90% 목표)
3. `go test -race ./internal/harness/... ./internal/cli/...`
4. `golangci-lint run --timeout=2m`
5. `go test -run TestSplitHarnessNamespaceNoLeak ./internal/template/`
6. 중립성 grep: `grep -rn "HARNESS-EVO-PIPE-REPAIR\|REQ-HEP" internal/template/templates/` → 0 matches
7. `moai harness doctor` 실행 스모크 (exemplar 수리 후 0 findings)
8. template↔live mirror parity diff (8개 표면)

---

## §F Milestones

### M1 — Go 어휘/스키마 수리 (REQ-HEP-001, 002) [TDD]

- RED: tier-promotions.jsonl 실데이터 라인을 verbatim fixture로 사용 — `{"pattern_key":"user_prompt::","to_tier":"auto_update","observation_count":196,"confidence":1}` → 현행 `MapPromotions` 0 candidates 재현
- GREEN: `actionableTiers` → `{rule, auto_update}`; `actionablePatternRE` → EventType enum 파생(빈 segment 허용) → ≥1 candidate
- REFACTOR: mapper.go 문서 주석(구 어휘 서술) 정정; scaffolder/DraftID 경로 회귀 확인
- 산출: `internal/harness/proposalgen/mapper.go` + `mapper_test.go` (+ `types.go` 파생 상수 필요 시)

### M2 — classify 자동화 + observer 등록 (REQ-HEP-003, 004) [TDD]

- RED: Stop 경로 e2e 테스트 — stop 이벤트 입력 시 promotions 미생성 재현
- GREEN: D3-(a) 채택 시 `runHarnessObserveStop`에 classify 연쇄(fail-open, 시간 예산 내); settings.json.tmpl + 로컬 settings.json에 PostToolUse `handle-harness-observe.sh` 등록
- `make build` (템플릿 임베드 재컴파일) + 렌더 검증
- 산출: `internal/cli/hook.go`(+test), `settings.json.tmpl`, 로컬 `settings.json`

### M3 — `moai harness doctor` 스모크 게이트 (REQ-HEP-005) [TDD]

- RED: fixture 하네스(B5 결함 클래스 재현 — 잘못된 manifest 경로 상수 + 비실존 agent 이름) → doctor가 ≥2 findings 반환하는 테스트 먼저 작성
- GREEN: v4lifecycle 스캔 재사용 + 4축 cross-ref 검사 구현; 0-harness 프로젝트 exit 0
- 산출: `internal/cli/harness/` doctor 구현(+test), `internal/harness/v4manifest/` 확장 필요분

### M4 — exemplar Runner 수리 + 게이트 회귀 증명 (REQ-HEP-007) [로컬 전용]

- 수리 전: 실제 `.claude/workflows/harness-release-update-run.js`에 doctor 실행 → 2 findings 관측 기록 (게이트 1호 실전 검증)
- 수리: MANIFEST_PATH → `.claude/commands/harness/release-update/manifest.json`; agent 참조 → `harness-release-update-specialist`; 헤더 구명칭 정정
- 수리 후: doctor 재실행 → 0 findings
- **template 반입 금지** (dev-only user-owned artifact)

### M5 — 디스패처/문서 드리프트 수리 (REQ-HEP-006, 008, 009, 010, 011, 012)

- build-entry Phase 0 가드 → 8-verb 집합(doctor 포함); commands/moai/harness.md argument-hint + SKILL.md :70 및 §harness 절 현행화
- harness-builder.md: manifest 단일 정본 선언(OR-분기 제거) + ACTIVATE 절에 스모크 게이트 계약 추가
- harness.md: Layer 1 FROZEN 목록에서 `.claude/agents/harness/` 제거(허용 명시) + rate-limit 3/week+24h 정정 + provenance note; learner SKILL.md L1 목록 동기 수정
- 전 표면 live + template mirror 동시 편집 → `make build`; 중립성 self-check

### M6 — 통합 검증 배치 + 마감

- §E 검증 배치 전체 실행(단일 턴 병렬)
- progress.md §E.2/§E.3 증거 기록 (manager-develop 소관)
- 커밋 분할: M1-M3 Go(테스트 포함), M4 로컬 exemplar, M5 docs+template — pathspec 제한 커밋

---

## §G Anti-Patterns

- **AP-1**: ToTier 방출 어휘를 바꿔 mapper에 맞추는 방향 — 기존 JSONL 데이터 호환 파괴 + `@MX:ANCHOR` 계약 위반 (D1 역방향 금지)
- **AP-2**: doctor에 JS AST 파서 도입 — Runner 참조 추출은 정규식 heuristic으로 충분 (Enforce Simplicity; 게이트는 결함 클래스 탐지가 목적, 완전 파싱 아님)
- **AP-3**: classify를 Stop hook에서 blocking으로 실행 — 훅 5s 예산 초과 시 세션 종료 차단 = death-spiral 위험 (fail-open 필수)
- **AP-4**: exemplar Runner 수정을 template에 미러링 — dev-only 격리 위반 (CLAUDE.local.md §21)
- **AP-5**: 정정 문서에 SPEC ID/REQ 토큰 기입 — §25 template 중립성 위반
- **AP-6**: frozen_guard 자체를 이번에 확장 — SPEC-3 범위 침범 (본 SPEC은 문서를 코드에 맞출 뿐, 코드 정책 불변)
- **AP-7**: "게이트 통과" 주장을 실행 출력 없이 보고 — verification-claim-integrity §1 위반

---

## §H Cross-References

- `.moai/specs/SPEC-V3R6-HARNESS-V4-001/` — v4 Builder 4-phase + manifest Runner (ACTIVATE 계약의 소유 SPEC)
- `.moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/` — mapper 원 계약(REQ-PGN-004..005; 본 SPEC이 어휘/스키마 조항을 supersede)
- `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/` — harness-classify subcommand 도입 SPEC (본 SPEC이 자동 배선으로 확장)
- `.moai/specs/SPEC-HARNESS-LOOP-CLOSURE-001/` — C1 헌법 제약(auto_apply:false)의 원 소유 SPEC (개정은 SPEC-3 소관)
- `.moai/docs/harness-namespace-doctrine.md` §24 — user-owned vs template-managed 네임스페이스 (조항 추가는 SPEC-3 소관)
- `.moai/docs/template-internal-isolation-doctrine.md` §25 — 중립성 콘텐츠 클래스
- CLAUDE.local.md §2 Template-First / §21 dev-only isolation
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — status transition ownership
