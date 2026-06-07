# SPEC-WEB-CONSOLE-008 — Implementation Plan

## §A. Context

- **작업 위치**: `/Users/goos/MoAI/moai-adk-go` (project root)
- **현재 branch**: `feat/SPEC-WEB-CONSOLE-007`(현재) → run-phase는 `feat/SPEC-WEB-CONSOLE-008` 신규 브랜치에서.
- **SPEC 산출물**: `.moai/specs/SPEC-WEB-CONSOLE-008/{spec.md, plan.md, acceptance.md}` (Tier M, 3-artifact).
- **연구·설계 근거**: `.moai/reports/web-console-statusline-gitconvention-audit.md` — statusline 17-defect inventory + "Redesign Proposal — statusline" + 9-step Implementation Plan + Test & Sentinel Impact. **권위 있는 ground-truth; run-phase에서 재발견 금지.**
- **cycle_type**: tdd (RED→GREEN→REFACTOR). development_mode: tdd.
- **코호트**: web-console-v4 statusline track. 006(HTMX+Templ enabler) + 007(write-seam/위젯 재사용) 위에 작성. git_convention은 009로 분리.
- **처분 철학**: honest hybrid — REMOVE(config theater) + WIRE(live lever) + FIX(drift). `mode=full`/`renderFullV3`는 부활 금지(거짓 제거만).

## §B. Known Issues / Constraints from Ground-Truth

- **B1. config surface vs Builder API 혼동 위험**: `mode:` YAML surface는 제거하되, `StatuslineMode` enum/`NormalizeMode`/`Render(data, mode)`/Builder Config `Mode` 필드/`SetMode`는 **유지**(HARD-1). builder.go:54의 `Mode StatuslineMode`는 Builder Config 필드(PRESERVE)이지 YAML surface가 아님 — 삭제 금지.
- **B2. loadSegmentConfig 실위치**: audit 본문은 `internal/statusline`로 적었으나 실제는 `internal/cli/statusline.go:170-186`. 테스트 호출자 = coverage_improvement_test.go:1289+, statusline_test.go:105+ (production caller 0).
- **B3. SLR-3 fix는 런타임 우선순위 변경 아님**: `segments != nil`이 preset을 이기는 우선순위는 그대로(HARD-6). fix = **write 경로가 비-custom preset을 항상 15-key map으로 expand**하게 만들어 preset이 silently no-op하지 않게 함. `presetToSegments`(update.go:2536)가 확장기.
- **B4. theme-only round-trip(characterization gate)**: integration_test.go:124-165는 preset 미전송 → preserve 경로(sync.go:129 guard) 유지가 전제. preset write-effective 변경이 이 경로를 깨면 안 됨(HARD-7). preset이 폼에 있을 때만 expand, 없으면 기존 segments 보존.
- **B5. 006 sentinel(integration_test.go:197-205)**: workflow/harness/git-strategy만 보호; statusline 미포함 → statusline.yaml write는 기대된 거동(integration_test.go:158-165도 statusline write를 단언). sentinel **무수정** GREEN(HARD-3). 설계가 변경 강제 시 → STOP, blocker(009/boundary 침범).
- **B6. Template-first**: 템플릿 YAML 편집(`internal/template/templates/.moai/config/sections/statusline.yaml`) → `make build`(go:embed embedded.go 재생성). `.templ` 편집 → `templ generate`(fieldsets_templ.go, zero-Node).
- **B7. Template neutrality(§15/§25)**: statusline.yaml 템플릿 주석은 SPEC ID/REQ 토큰/`REQ-CC297` 류 내부 인용 무포함 — generic("Theme: catppuccin-mocha (default) | catppuccin-latte" 류).
- **B8. symmetry test 순서**: `StatuslineConfig` symmetryCases는 **마지막에**(struct에서 Mode 제거 + 템플릿 mode/refresh_interval 제거가 완료되어 struct↔YAML이 이미 대칭일 때) 추가해야 clean PASS(audit 9-step의 sequence note).
- **B9. AskUserQuestion 금지(subagent boundary)**: blocker 발견 시 structured blocker report 반환. orchestrator가 사용자 상호작용 수행.
- **B10. Untouched paths PRESERVE**: git_convention 섹션/struct/검증/위젯 일절 무변경(009 스코프, HARD-4). workflow/git-strategy/harness/llm 무변경. runtime-managed files 무변경.
- **B11. statusline.go:72 over-removal 방지(D6)**: line 72 `opts.Mode`(builder hand-off)는 Builder Config `Mode` 필드를 PRESERVE한다 — config surface 제거 대상 아님. CLI는 config-derived mode 계산(50-55)만 제거하고 builder에는 `ModeDefault`를 전달(또는 mode 인자 생략). Builder Config `Mode` 필드/`SetMode`/enum 삭제 금지(HARD-1/HARD-5). M3 site-list의 `statusline.go:50-55`만 removal 대상, `:72`는 PRESERVE-with-default.

## §C. Pre-flight (run-phase 진입 전 검증)

```bash
# 1. branch + baseline
git checkout -b feat/SPEC-WEB-CONSOLE-008    # plan PR merge 후
go test ./internal/web/... ./internal/cli/... ./internal/config/... ./internal/profile/... ./internal/statusline/... ./pkg/models/... -count=1   # baseline GREEN

# 2. cross-platform build
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. templ codegen 가용
which templ || go install github.com/a-h/templ/cmd/templ@latest

# 4. 신규 validator 0개(009 침범 가드) baseline
grep -cE 'func validate(Workflow|GitStrategy|Harness|Llm)Config' internal/config/validation.go || echo 0

# 5. git_convention 무관 baseline 캡처(이후 git diff로 무변경 증명)
git rev-parse HEAD
```

## §D. Constraints (HARD — DO NOT VIOLATE)

- spec.md §B HARD-1..9 전부.
- **Builder API PRESERVE**: `StatuslineMode` enum / `NormalizeMode` / `Render(data, mode)` / Builder Config `Mode` 필드 / `SetMode` 무변경(HARD-1).
- **renderFullV3 부활 금지**(HARD-2): 5-line layout 재구현·재호출 0.
- **006 sentinel 무수정**(HARD-3): integration_test.go:197-205 byte-identical. 강제 변경 시 STOP+blocker.
- **git_convention 무관**(HARD-4): `git diff` 상 git_convention 관련 파일/struct/validation 변경 0.
- **canonical 모델에 Mode 추가 금지**(HARD-5).
- **theme-only round-trip GREEN 유지**(HARD-7): integration_test.go:124-165 무수정 PASS.
- 금지 명령: `--no-verify`, `--amend`, force-push to main.
- 사용 의무: Conventional Commits(`feat(SPEC-WEB-CONSOLE-008): M{N} ...`), `🗿 MoAI` trailer.

## §E. Self-Verification (각 milestone GREEN 기준)

- `go test ./internal/web/... ./internal/cli/... ./internal/config/... ./internal/profile/... ./internal/statusline/... ./pkg/models/... -count=1` exit 0
- `templ generate && git diff --exit-code internal/web/*_templ.go` (drift-free)
- `make build` 후 embedded.go 갱신 반영
- `grep -cE 'func validate(Workflow|GitStrategy|Harness|Llm)Config' internal/config/validation.go` == 0 (009 침범 가드)
- 006 sentinel(`TestScopeBoundary` 류 in integration_test.go) PASS 무수정
- git_convention 관련 파일 `git diff` 0줄

## §F. Tier 결정 + 정당화

**선택: Tier M (standard).** plan-auditor PASS threshold 0.80.

**정당화 (right-size 판단)**:
- **Tier S로 축소 불가** 이유:
  1. **다중 패키지 변경** — internal/cli + internal/profile + internal/statusline + internal/web + internal/config + pkg/models + 템플릿. 단일 파일/패키지 아님.
  2. **3-struct 통합** — Mode 필드를 2개 private struct에서 제거 + 사용처(statusline.go:50-55/72, sync.go:70/120-121) 정리는 cross-file 일관성 작업.
  3. **live lever 배선(SLR-3)** — preset write-effective는 단순 삭제가 아니라 write 경로 동작 변경 + characterization gate(theme-only round-trip) 비-회귀 보장 필요.
  4. **dead code 제거 + 테스트 정리** — `loadSegmentConfig` + 4개 테스트 사이트(coverage_improvement_test.go, statusline_test.go) Class C source-coupled 갱신.
  5. **TDD 사이클 다종 신규 거동** — preset-expand / seed 15-key / symmetry case / web 조건부 노출 / mode surface 제거.
- **Tier L은 과대** 이유: 신규 섹션/validator 0개, 서버 계약 무변경(hx-boost full-page 유지), 변경은 대부분 삭제·교정(신규 기능 아님), AC ~22 이내, SSE stall 임계(30+ task) 미달 → Round 분할 불요, design.md/research.md 불요(audit 보고서가 그 역할 수행).

## §G. Milestones (priority-ordered, no time estimates — audit 9-step에 매핑)

| Mn | Owner | Objective | TDD class | defect ids | Key files |
|----|-------|-----------|-----------|------------|-----------|
| **M1** | manager-develop (tdd) | **Template schema correction**: `statusline.yaml`에서 `mode:`/`refresh_interval:` 제거, `theme: "default"`→`"catppuccin-mocha"`, 주석 generic 정정(2 themes only). `make build`로 embedded.go 재생성. (audit step 1) | RED(템플릿 단언) → GREEN | SLM-3/SLR-6/SLR-4/HARD-9 | template statusline.yaml, embedded.go(생성), neutrality test |
| **M2** | manager-develop (tdd) | **seed 15-key 완성**: `defaultStatuslineSegments()` → `return presetToSegments("full", nil)`. 15-key 단언 테스트. (audit step 2) | RED(11→15) → GREEN | SLM-5 | internal/profile/sync.go:148-162 |
| **M3** | manager-develop (tdd) | **3-struct 통합 — Mode 제거**: `statuslineData.Mode`(sync.go:80-85) + `statuslineFileConfig.Mode`(statusline.go:127-128) + config-derived mode 계산 사용처(sync.go:70/120-121, statusline.go:50-55/151/163) 제거. **statusline.go:72 `opts.Mode` builder hand-off는 PRESERVE-with-default**(Builder Config Mode 필드 유지, `ModeDefault` 전달 — B11/D6). Builder API(enum/NormalizeMode/Render/Config Mode/SetMode) **무변경 확인**. canonical 모델 무변경. (audit step 3) | RED(struct 단언) → GREEN → REFACTOR | SLM-1/SLM-2/SLR-2, HARD-1/HARD-5 | internal/cli/statusline.go, internal/profile/sync.go, builder_test.go(무수정 PASS 확인) |
| **M4** | manager-develop (tdd) | **preset write-effective(SLR-3)**: 비-custom preset 저장 시 `presetToSegments`로 15-key map expand(write 경로). 런타임 우선순위 불변. **theme-only round-trip(integration_test.go:124-165) 무수정 GREEN 확인**(HARD-7). (audit step 4) | RED(preset→segments) → GREEN | SLR-3, HARD-6/HARD-7 | internal/profile/sync.go `syncStatusline`(96-145), 프로젝트-config write 경로 |
| **M5** | manager-develop (tdd) | **dead code 제거**: `loadSegmentConfig`(statusline.go:170-186) + 테스트 refs(coverage_improvement_test.go:1289+, statusline_test.go:105+) 삭제. (audit step 5) | RED(부재 단언) → GREEN | SLR-7 | internal/cli/statusline.go, coverage_improvement_test.go, statusline_test.go |
| **M6** | manager-develop (tdd) | **doc/comment 정정**: `builder.go:50` ThemeName doc → 2 real themes; `SegmentRepo`(types.go:321) intentional-exclusion 주석 1줄. (audit step 6) | RED(doc grep) → GREEN | SLR-5/SLM-7, HARD-2 | internal/statusline/builder.go, internal/statusline/types.go |
| **M7** | manager-develop (tdd) | **web mode control 제거 + 라벨**: `statusline_mode` select(fieldsets.templ:134) 제거, `statuslineModeCanonical`+검증(validate.go:40-41/136-137) 제거, handlers.go(29/82/365) StatuslineMode view-model/parse 제거, 카운트 라벨 `"3 fields"`→`"2 fields"`(fieldsets.templ:128). `templ generate`. (audit step 7) | RED(부재/parity) → GREEN | SLR-1/SLW-4 | fieldsets.templ, handlers.go, validate.go, fieldsets_templ.go(생성) |
| **M8** | manager-develop (tdd) | **web 조건부 segment 노출**: segment 체크박스를 preset=custom일 때만 편집 가능(disabled when preset≠custom), 비-custom은 server expand. `templ generate`. (audit step 8) | RED(조건부 markup) → GREEN | SLR-3 UX | fieldsets.templ, fieldsets_templ.go(생성) |
| **M9** | manager-develop (tdd) | **symmetry CI guard 추가(LAST)**: `StatuslineConfig`를 `audit_struct_yaml_symmetry_test.go` symmetryCases에 추가(struct+template 이미 대칭). 006 sentinel 무수정 확인 + git_convention 무변경 git diff 확인 + full suite + `templ generate` drift-free. REFACTOR(위젯 chrome 정리). (audit step 9) | RED(symmetry) → GREEN → REFACTOR | SLM-4, HARD-3/HARD-4 | internal/config/audit_struct_yaml_symmetry_test.go, integration_test.go(READ-only 단언), 전체 |

> Round 분할 없음(Tier M, 30+ task 미만 — SSE stall 임계 미달, `sprint-round-naming.md` 기준 Milestone만 사용).
> M9 symmetry case가 마지막인 이유: M1(템플릿 mode/refresh_interval 제거) + M3(struct Mode 제거)가 완료되어 struct↔YAML 대칭이 확보된 뒤에야 clean PASS(B8 / audit sequence note).

## §H. Cross-References

- spec.md §B(HARD) / §D(변경 인벤토리) / §E(테스트 영향) / §F(Exclusions)
- `.moai/reports/web-console-statusline-gitconvention-audit.md`(전체 file:line + 9-step plan + Test & Sentinel Impact)
- acceptance.md (AC-WC8-NNN)

## §I. Anti-Patterns (회피 목록)

- Builder API의 `StatuslineMode` enum/Render/`Mode` Config 필드 삭제(HARD-1 위반) → config surface(`mode:` YAML + 웹)만 제거.
- `renderFullV3` 부활/wiring(HARD-2 위반) → 거짓 제거만.
- canonical `models.StatuslineConfig`에 Mode 추가(HARD-5 위반) → private struct에서 제거 방향.
- 런타임 segments-wins 우선순위 변경(HARD-6 위반) → write 경로 expand만.
- preset write-effective가 theme-only round-trip 깨뜨림(HARD-7 위반) → preset 폼 존재 시에만 expand.
- 006 sentinel(integration_test.go:197-205) 수정 → 무수정, 강제 시 blocker.
- git_convention 손대기(HARD-4 위반, 009 스코프) → git diff 0줄.
- symmetry case를 struct/template 정리 전에 추가 → M9(LAST) 순서 준수.
- 신규 statusline validator 함수 생성(scope creep) → SLM-6/SLW-4는 deferrable SHOULD; 기존 coercion 안전망 활용.
