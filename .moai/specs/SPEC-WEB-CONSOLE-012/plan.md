# SPEC-WEB-CONSOLE-012 — Implementation Plan

> Tier S (delegation-classified). 파일 수(~12-17, 테스트 포함)는 Tier M 대역이나 LOC delta는 소규모 net-negative 기계적 제거다 — boundary case 기록: 위임 prompt는 Tier M 수준 rigor(Section A-E) 권장.

## §A Context

- **Base**: main checkout, Route A (Hybrid Trunk main-direct). 병렬 세션이 `internal/statusline/renderer.go` + `cache_hit_test.go` 수정 중 (git status 실측) — 해당 트리 절대 무접촉.
- **Origin**: 2026-07-10 orchestrator 감사 Track 1 (A1-A5). 모든 앵커 2026-07-10 plan-phase 재실측 완료 (§C 명령 기록). 라인 번호는 드리프트 가능 — run-phase 착수 시 content-token 기준 재검증 의무 (feedback_line_number_drift_asymmetry).
- **SPEC 산출물**: spec.md + plan.md + acceptance.md + progress.md (4 파일).

### §A.1 Verified change surface (plan-phase 실측)

| # | 파일 | 변경 | 앵커 (content-token) |
|---|------|------|----------------------|
| 1 | `internal/settings/schema_sections.go` | `llmFields()` tier 루프 `{high,medium,low,opus,sonnet,haiku}` → `{high,medium,low,fable}`; `WorkflowAgentPurposes()` 삭제; 헤더 주석 "7개 seam 섹션" → 6 | `func llmFields`, `func WorkflowAgentPurposes` |
| 2 | `internal/settings/sectionapply.go` | `applyLLMKey` case `glm.models.{opus,sonnet,haiku}` 삭제, `glm.models.fable` case 추가 | `func applyLLMKey` |
| 3 | `internal/settings/sectionroute.go` | `"research": RouteSeam` 삭제; `SeamSections()`에서 research 제거; 주석 "7개 섹션" 정정 | `sectionRoutes`, `func SeamSections` |
| 4 | `internal/settings/sectionwrite.go` | `sectionRootKeys`에서 `"research"` 삭제; 파일 doc comment 정정 | `sectionRootKeys` |
| 5 | `internal/settings/schema.go` | `SectionResearch` const + `AllSections()` 항목 삭제만. (0.2.0: auto_detection FieldDef 3블록은 USED 재분류로 **무접촉 잔류** — REQ-WC12-020) | `SectionResearch` |
| 6 | `internal/web/assets/i18n.js` | `f.llm.glm.models.{opus,sonnet,haiku}.*` 키 제거 (전 locale 블록); `f.llm.glm.models.fable.{title,desc}` 추가 (high와 동수 locale). (0.2.0: `f.git_convention.auto_detection.*` 키는 무접촉 잔류) | `f.llm.glm.models.` |
| 7 | `internal/web/schemaform.go` | `schemaSectionMetas` LLM desc "(high/medium/low/opus/sonnet/haiku)" → "(high/medium/low/fable)" | `SectionLLM, "rocket"` |
| 8 | `internal/web/server.go` | 패키지 doc scope-boundary 문단 재작성 (10섹션 열거 → 최종 계약; db/research 제외군 명기) | `Scope boundary` |
| 9 | `internal/web/projectconfig.go` | @MX:REASON 블록 동일 재작성 | `@MX:REASON` |
| 10 | `internal/web/assets.go` + `internal/web/validate.go` | Where-gate 통과 시 `errDictKey` sentinel + keep-alive 삭제 (REQ-WC12-030/031) | `errDictKey` |
| 11 | 테스트 | `internal/settings/{schema_test,schema_sections_test,accessors_test}.go`, `internal/web/{schema_sections_test,schema_label_test,agent_settings_test}.go`, `internal/cli/schema_bridge_test.go` — 필드셋 변화 반영 (양표면 동시, REQ-WC12-050). **M1 RED-turning 3건 (0.2.0 D3 추가)**: `internal/settings/sectionroute_test.go:22`(research→RouteSeam 기대), `internal/settings/sectionwrite_test.go:135`(research seam-write 성공 케이스), `internal/web/scope_contract_test.go:34`(seam 7종 루프에 research 포함) — M1에서 research 폐선 시 RED로 전환되므로 M1 커밋에 신규 기대(RouteExcluded + 거부)로 동시 갱신 | `sectionroute_test`, `sectionwrite_test`, `scope_contract_test` |
| 12 | `internal/cli/schema_bridge.go` (조건부) | fable TUI label 항목 필요 여부는 llm 필드의 persist-kind로 결정 (PersistTypedSection → web-only key-chip, TUI bridge 불요 — run-phase 실측). (0.2.0: auto_detection orphan 정리 항목은 무접촉 반전으로 소멸) | `func llmFields` persist-kind |

## §B Known Issues (위험 주입)

- **B1 — 양표면 스키마 회귀 (011 M2b 교훈)**: 공유 스키마 SSOT 변경은 web 파리티 테스트와 TUI `schema_bridge_test.go`를 반드시 동시 갱신. web만 고치면 TUI 3-test 파손 재발.
- **B2 — i18n 4-locale 파리티**: i18n.js는 locale 블록별 동일 키셋. fable 추가/ghost 제거는 전 locale 블록 동시 적용. AC는 hard-pin 대신 sibling-key 동수 파생 단언 (011 파생-assertion 선례; 근사 카운트 금지 교훈).
- **B3 — `-run` 무매치 exit-0 함정 (011 D4 교훈)**: AC 검증은 실존 테스트 함수명 명시 + full-package `go test` 병행.
- **B4 — validate.go byte-guard 불확실성**: i18n_test.go:23 주석은 "asserted by a sibling diff in M4"(시점 검증) — 상시 guard 실존 여부는 미확정. REQ-WC12-030 Where-gate가 pre-flight에서 판정 (§C-5). 발견 시 REQ-WC12-031 fallback (문서화-잔류, blocker 아님).
- **B5 — auto_detection 무접촉 (0.2.0 반전)**: A5 전원 USED 재분류로 `git_convention.auto_detection.*`은 스키마/i18n/bridge 전부 무접촉. `harness.auto_detection.enabled`(schema_sections.go:215)도 별개 live 필드로 무접촉 (REQ-WC12-023 방어 guard 잔류, AC-WC12-014).
- **B9 — dead 분류 grep 프로토콜 (0.2.0 D1/D8 교훈)**: field-dot 패턴(`\.Field\.` / `\.Struct\.`) 단독 grep은 whole-struct bind(`x := cfg.A.B` 후 `x.Field`)와 미러 struct 복사를 구조적으로 놓친다 — iter-1이 이 결함으로 live 필드 3개를 DEAD 오분류(CRITICAL). 모든 dead 판정은 field-dot grep + bare-symbol grep(`AutoDetection` 등 타입/필드 심볼 단독) 쌍으로 수행하고, bare-symbol의 추가 매치 전부를 설명해야 분류 확정 (acceptance.md §D.2 preamble).
- **B6 — 병렬 세션 race**: `internal/statusline/*` 무접촉; commit은 specific-path `git add`만 (shared-checkout 교훈). Pre-Spawn Sync Check (git fetch + rev-list) 준수.
- **B7 — WorkflowAgentPurposes Chesterton's fence**: 함수 주석이 의도적 유지를 주장 — spec §A(A4)에 검토-후-기각 근거 기록됨. 삭제 직전 re-grep 0-caller 재확인 (REQ-WC12-032).
- **B8 — 라이브 llm.yaml ghost 키**: 라이브 파일에 `models.{opus,sonnet,haiku}` 키가 값과 함께 실존. struct legacy 필드 보존으로 typed re-marshal이 이를 파괴하지 않음을 저장-roundtrip 테스트로 확인 (acceptance.md EC-2).

## §C Pre-flight (run-phase 착수 전 의무 검증)

```bash
# 1. baseline + 병렬 세션 상태
git branch --show-current && git rev-parse HEAD
git status --porcelain | grep 'internal/statusline' || echo "statusline clean-of-me"

# 2. build + lint baseline
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5

# 3. 앵커 content-token 재검증 (라인 드리프트 대비)
grep -n 'func llmFields' internal/settings/schema_sections.go
grep -n '"research"' internal/settings/sectionroute.go internal/settings/sectionwrite.go
grep -n 'SectionResearch' internal/settings/schema.go
grep -n 'git_convention.auto_detection' internal/settings/schema.go

# 4. A4b 0-caller 재확인 (삭제 게이트)
grep -rn 'WorkflowAgentPurposes' --include='*.go' internal/ cmd/ pkg/   # 기대: 정의부만

# 5. REQ-WC12-030 Where-gate: validate.go 상시 byte-guard 탐색
grep -rn 'validate\.go' internal/web/*_test.go internal/template/*_test.go
grep -rln 'sha256\|byte' internal/web/*_test.go | xargs grep -ln 'validate' 2>/dev/null || echo "no standing byte guard"
# (0.2.0 D6 지침) 매치 판정 시 주석-전용 언급과 기계적 단언을 구분할 것:
#   i18n_test.go:23의 "AC-WC5-010a: validate.go is byte-unchanged"는 //-주석 서술(역사 기록)이며
#   guard가 아니다. "상시 guard 존재" 판정은 validate.go의 바이트/해시/구조를 실제로 단언하는
#   테스트 코드 라인(비주석)이 있을 때만 성립 — 매치 각각을 열어 assertion 여부를 확인한다.

# 6. persist-kind 실측 (fable TUI bridge 항목 필요 여부 결정 — §A.1 row 12)
grep -n -B2 -A8 'func llmFields' internal/settings/schema_sections.go   # typedField persist-kind 확인
```

## §D Constraints (DO NOT VIOLATE)

- PRESERVE: `internal/config/types.go`(GLMModels legacy 필드, AutoDetectionConfig), `internal/config/{defaults,validation,loader}.go`, `internal/cli/glm.go`, `internal/cli/hook_pre_push.go`, `internal/core/quality/trust.go`, `internal/statusline/**`, `internal/template/templates/**`, 라이브 `.moai/config/sections/*.yaml`.
- 금지: `--no-verify`, force-push, `git add -A`/`git add .` (specific-path만), blind sed 일괄 치환 (B5).
- Conventional Commits: `refactor(SPEC-WEB-CONSOLE-012): M{N} <subject>` (+ `🗿 MoAI` trailer). 첫 run-phase 커밋에서 frontmatter `draft → in-progress` (manager-develop 소유).
- 매 milestone 후 full suite: `go test ./...` (부분 성공 선언 금지).

## §E Self-Verification

manager-develop 완료 보고는 manager-develop-prompt-template.md §E (E1 AC 매트릭스 / E2 cross-platform build / E3 coverage / E5 lint NEW-vs-baseline / E6 push state) + verification-claim-integrity 5-section 형식을 따른다. AC SSOT = acceptance.md.

## §F Milestones (priority-based, 의존 순서)

| M | 내용 | REQ | Priority |
|---|------|-----|----------|
| M1 | A2 research seam 폐선 (sectionroute/sectionwrite/schema.go + 주석) + 거부 회귀 테스트 | 010-012 | High |
| M2 | A1 llm ghost 제거 + fable 노출 (schema_sections/sectionapply/i18n.js/schemaform) + 양표면 테스트 | 001-005, 050 | High |
| M3 | A5 잔류 확인 게이트 (0.2.0 반전 — 코드 무접촉): §D.2 강화 프로토콜 분류 명령 재실행 + AC-WC12-011/012/014 확인만 | 020, 021, 023 | Low |
| M4 | A4 dead code (errDictKey Where-gate, WorkflowAgentPurposes) | 030-032 | Medium |
| M5 | A3 doc comment 최종 재작성 (M1/M3 결과 반영 — 마지막 순서 고정) | 040 | Medium |

M1/M2는 독립이나 순차 실행 (동일 파일 schema_sections.go 접촉). M5는 M1+M3 완료 후.

## §G Anti-Patterns

- 웹 파리티만 갱신하고 TUI bridge 테스트 방치 (B1).
- field-dot grep 단독을 dead 분류의 유일 근거로 삼기 — whole-struct bind/미러 struct를 놓쳐 live 필드를 오삭제 (B9, iter-1 D1 실증).
- `auto_detection` 계열 필드 접촉 — A5는 전원 USED 잔류, harness 필드 포함 무접촉 (B5).
- i18n 키 수 hard-pin (locale 수 가정) — sibling 파생 단언 사용 (B2).
- 완결 SPEC의 역사적 AC(BYTE-UNCHANGED)를 상시 제약으로 오독해 A4a를 무근거 skip — Where-gate 판정 의무.
- 라이브 llm.yaml ghost 키를 "정리" — 런타임 config 무접촉 (Out of Scope).

## §H Cross-References

- SPEC-WEB-CONSOLE-011 (M4 다이어트 선례: FieldDef 제거 + struct/yaml 보존 패턴; M2b 양표면 회귀 교훈)
- SPEC-WEB-CONSOLE-010 (스키마 SSOT 확립), SPEC-WEB-CONSOLE-006 (validate.go byte-unchanged 시점 제약 출처)
- SPEC-WEB-CONSOLE-009 (**completed** — git_convention honest hybrid: auto-detection + max_length 배선. A5 USED 분류의 배선 출처. iter-1의 "미작성" 기재는 stale memory 오류 — 0.2.0 정정)
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 (A5 분류 프로토콜)
