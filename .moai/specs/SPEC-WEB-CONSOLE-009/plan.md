# SPEC-WEB-CONSOLE-009 — Implementation Plan

## §A. Context

- **작업 위치**: `/Users/goos/MoAI/moai-adk-go` (project root)
- **현재 branch**: `main`(현재) → run-phase는 `feat/SPEC-WEB-CONSOLE-009` 신규 브랜치에서.
- **SPEC 산출물**: `.moai/specs/SPEC-WEB-CONSOLE-009/{spec.md, plan.md, acceptance.md}` (Tier L, 3-artifact — design.md/research.md는 audit 보고서가 대체, §F 참조).
- **연구·설계 근거**: `.moai/reports/web-console-statusline-gitconvention-audit.md` — git_convention 25-defect inventory + "Redesign Proposal — git_convention" + 8-step Implementation Plan + Test & Sentinel Impact. **권위 있는 ground-truth; run-phase에서 재발견 금지.** (audit 본문 일부 라인은 008 이전 값 — spec.md §A의 post-008 재검증 라인을 우선 사용.)
- **cycle_type**: tdd (RED→GREEN→REFACTOR). development_mode: tdd.
- **코호트**: web-console-v4 git_convention track. 006(HTMX+Templ enabler) + 007(write-seam `writeProjectNestedConfig`/위젯 재사용) + 008(statusline honest-hybrid 선례) 위에 작성.
- **처분 철학**: honest hybrid — REMOVE(custom 엔진·dead·formatting·validation.{enabled,enforce_on_commit}) + WIRE(Fix A auto-detection / Fix B max_length) + FIX(Fix C template flat→nested / symmetry guard). `custom` 엔진은 부활 금지(제거만, audit policy decision 1). GCR-5는 분리 deferred.

## §B. Known Issues / Constraints from Ground-Truth

- **B1. config surface vs runtime engine 구분**: `models.CustomConventionConfig`(config.go:246-254) 제거는 web/config/CLI/validator 레이어. 그러나 convention 패키지 내부 `ConventionConfig`(types.go:44-52)는 `Parse`/`ParseBuiltin`/`templates.go`가 built-in 엔진 구동에 쓰는 **별개 struct** — `Parse`/`ParseBuiltin`/built-in templates는 **무변경**. `custom` 제거 = `models` 레이어 + dead `LoadFromConfig`만. `Parse`/built-in 엔진 손대지 말 것.
- **B2. `LoadConvention` 시그니처 변경 cascade(Fix A, HARD-6)**: 유일 production caller = hook_pre_push.go:54. convention 패키지 테스트(manager_test.go) 갱신. **Builder 외 다른 production caller 0 확인**(`grep -rn 'LoadConvention' internal/ --include='*.go' | grep -v _test.go`). LoadFromConfig는 제거되므로 그 production caller(0) 확인 후 삭제.
- **B3. `SetMaxLength` 부재(Fix B, GCR-7)**: `Convention.MaxLength`(types.go:60) 필드는 존재하나 setter 없음 → **신규 추가**(`Convention` 또는 `Manager`에 `SetMaxLength(int)`). LoadConvention 후 `validation.max_length`를 forward. built-in `Parse`가 MaxLength=100을 굳히는 것을 setter가 override.
- **B4. `enforce_on_push` 보존(HARD-7/HARD-5)**: `validation.enforce_on_push`는 삭제 금지(유일 live gate). `isEnforceOnPushEnabled()`(hook_pre_push.go:121-134)가 `Validation.EnforceOnPush`(129)를 읽는 경로 무변경. 삭제는 `validation.{enabled, enforce_on_commit}`만.
- **B5. 006 sentinel(integration_test.go:191-205)**: workflow/harness/git-strategy만 보호; git_convention 미포함 → `SetSection("git_convention", ...)` write는 기대된 거동. sentinel **무수정** GREEN(HARD-3). 설계가 변경 강제 시 → STOP, blocker.
- **B6. Template-first**: 템플릿 YAML 재작성(`internal/template/templates/.moai/config/sections/git-convention.yaml`) → `make build`(go:embed embedded.go 재생성). `.templ` 편집 → `templ generate`(fieldsets_templ.go, zero-Node).
- **B7. Template neutrality(§15/§25)**: git-convention.yaml 템플릿 주석은 SPEC ID/REQ 토큰/내부 인용 무포함 — generic("Convention: auto | conventional-commits | angular | karma" 류).
- **B8. symmetry test 순서 게이트(HARD-10)**: `GitConventionConfig` symmetryCases는 **M1(struct trim) + M4(template nested 재작성)가 완료되어 struct↔YAML 대칭이 확보된 뒤** 마지막(M8)에 추가. 순서 위반 시 RED for the wrong reason(audit step 6 sequence note).
- **B9. AskUserQuestion 금지(subagent boundary)**: blocker 발견 시 structured blocker report 반환. orchestrator가 사용자 상호작용 수행.
- **B10. Untouched paths PRESERVE**: statusline 섹션/struct/검증/위젯 일절 무변경(008 스코프, HARD-2). workflow/harness/llm 무변경. GCR-5 관련 파일(.git_hooks/pre-push, hook_install.go:27-68 prePushHookContent, git-strategy.yaml hooks.pre_push) **무변경**(HARD-4). runtime-managed files 무변경.
- **B11. `convention=auto` empty-preserve 매핑(GCW-7)**: 빈 옵션 라벨 `"(unchanged)"`로 변경하되 on-save preserve 매핑(빈 값 → 기존 보존)은 **이미 올바름** — 라벨 string만 변경, write 로직 무변경.
- **B12. defaults_test fail-loud**: `TestNewDefaultGitConventionConfig`(defaults_test.go:570+)는 `Validation.Enabled`/`EnforceOnCommit`/`Formatting` 단언을 포함 → struct trim 시 이 단언들이 **컴파일 실패/RED**가 정상(fail-loud 신호). 테스트를 trimmed shape로 갱신.

## §C. Pre-flight (run-phase 진입 전 검증)

```bash
# 1. branch + baseline
git checkout -b feat/SPEC-WEB-CONSOLE-009    # plan PR merge 후
go test ./internal/web/... ./internal/cli/... ./internal/config/... ./internal/git/convention/... ./pkg/models/... -count=1   # baseline GREEN

# 2. cross-platform build
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. templ codegen 가용
which templ || go install github.com/a-h/templ/cmd/templ@latest

# 4. LoadConvention/LoadFromConfig production caller 인벤토리(Fix A cascade scope 확정)
grep -rn 'LoadConvention\|LoadFromConfig' internal/ --include='*.go' | grep -v '_test.go'   # 기대: hook_pre_push.go:54(LoadConvention) + manager.go 정의 only

# 5. SetMaxLength 부재 확인(Fix B 신규 대상)
grep -rn 'func.*SetMaxLength' internal/git/convention/ || echo "absent — Fix B adds it"

# 6. statusline 무관 baseline 캡처(이후 git diff로 무변경 증명)
git rev-parse HEAD
```

## §D. Constraints (HARD — DO NOT VIOLATE)

- spec.md §B HARD-1..10 전부.
- **`custom` 엔진 완전 제거 — 부활 금지**(HARD-1): 4 lockstep site + struct + YAML + 웹 위젯 + CLI 옵션 + dead LoadFromConfig 삭제. wiring 0.
- **statusline 무관**(HARD-2): `git diff` 상 statusline 관련 파일/struct/validation/web 위젯 변경 0.
- **006 sentinel 무수정**(HARD-3): integration_test.go:191-205 byte-identical. 강제 변경 시 STOP+blocker.
- **GCR-5 wiring 금지**(HARD-4): .git_hooks/pre-push / hook_install.go:27-68 prePushHookContent / git-strategy.yaml hooks.pre_push 무변경 `git diff` 0줄.
- **`enforce_on_push` 보존**(HARD-7/HARD-5): `validation.enforce_on_push` + `isEnforceOnPushEnabled()` 무변경. 삭제는 `validation.{enabled,enforce_on_commit}`만.
- **`Detect()` 시그니처 무변경**(HARD-5): LoadConvention의 인자 전달만 변경.
- **`Parse`/`ParseBuiltin`/built-in templates 무변경**(B1): convention 패키지 내부 엔진은 PRESERVE.
- 금지 명령: `--no-verify`, `--amend`, force-push to main.
- 사용 의무: Conventional Commits(`feat(SPEC-WEB-CONSOLE-009): M{N} ...`), `🗿 MoAI` trailer.

## §E. Self-Verification (각 milestone GREEN 기준)

- `go test ./internal/web/... ./internal/cli/... ./internal/config/... ./internal/git/convention/... ./pkg/models/... -count=1` exit 0
- `templ generate && git diff --exit-code internal/web/*_templ.go` (drift-free)
- `make build` 후 embedded.go 갱신 반영
- 4 lockstep site에 `custom` 0: `grep -rn '"custom"' pkg/models/config.go internal/config/validation.go internal/web/validate.go internal/cli/profile_setup.go` (comment 제외)
- 006 sentinel(`TestRealServerRoundTrip` 류 in integration_test.go) PASS 무수정
- statusline 관련 파일 `git diff` 0줄; GCR-5 관련 파일 `git diff` 0줄

## §F. Tier 결정 + 정당화

**선택: Tier L (large).** plan-auditor PASS threshold 0.85.

**정당화 (right-size 판단)**:
- **Tier M으로 축소 불가** 이유(008은 M이었으나 009는 한 단계 위):
  1. **런타임 시그니처 변경 cascade(Fix A)** — `LoadConvention(name string)` → `AutoDetectionConfig` 수용으로 시그니처가 바뀌며 `internal/git/convention` 패키지 + 유일 caller(hook_pre_push.go) + 패키지 테스트(manager_test.go)로 cascade한다. 008은 edit-only nested 작업이었고 런타임 시그니처 변경이 없었다 — 이것이 L-tier driver.
  2. **신규 런타임 동작(Fix B)** — `SetMaxLength` setter 신설 + LoadConvention 후 적용은 신규 코드/거동(삭제·교정이 아님).
  3. **다중 패키지 + struct 4종 trim** — pkg/models(struct 2개 삭제 + 필드 4개 삭제) + internal/config(defaults + validator) + internal/web(위젯 3종 변경) + internal/cli(wizard) + internal/git/convention(LoadConvention/SetMaxLength/LoadFromConfig 삭제) + 템플릿 + 4-site enum lockstep. ~16 파일.
  4. **dead code + 테스트 정리 광범위** — LoadFromConfig + custom 테스트(parser_test.go, manager_test.go) + defaults_test.go fail-loud 갱신 + validator custom 테스트 + web characterization 갱신.
  5. **TDD 사이클 다종 신규 거동** — auto-detection 4-knob honor / max_length forward / 템플릿 nested / symmetry case / web sample_size+enforce_on_push 추가 + custom 제거 / 조건부 grey UX.
- **그러나 design.md/research.md는 불요** — `.moai/reports/web-console-statusline-gitconvention-audit.md`가 25-defect inventory + Corrected Schema + Web UX + Runtime Wiring Fixes(A/B/C) + 8-step plan + Test & Sentinel Impact를 모두 담은 권위 있는 design+research 산출물이다(008과 동일 처리). 따라서 Tier L이되 **3-artifact**(spec/plan/acceptance) + audit 보고서 참조로 충분 — design.md/research.md를 새로 쓰면 audit 보고서를 중복 복제하는 것. 이 결정을 본 §F에 명시 기록한다.
- **SSE stall 임계 미달**: 8 milestone(< 30 task) → Round 분할 불요(`sprint-round-naming.md` 기준 Milestone만 사용).

## §G. Milestones (priority-ordered, no time estimates — audit 1-8 step에 매핑)

> **순서 게이트(HARD-10/B8)**: M1(struct trim) + M4(template flat→nested 재작성)가 struct↔YAML 대칭을 확보해야 M8(symmetry case)이 올바른 이유로 GREEN. M8은 반드시 마지막.

| Mn | Owner | Objective | TDD class | defect ids | Key files |
|----|-------|-----------|-----------|------------|-----------|
| **M1** | manager-develop (tdd) | **Struct trim**: `models.GitConventionConfig`에서 `Custom`/`Formatting` 필드 + `CustomConventionConfig`/`FormattingConfig` struct + `Validation.{Enabled, EnforceOnCommit}` 필드 제거; oneof 태그 `oneof=auto conventional-commits angular karma`(custom 제거); range 태그 정비. (audit step 1) | RED(필드 부재 단언) → GREEN | GCM-8/GCM-9, REQ-WC9-002 | pkg/models/config.go:205-254 |
| **M2** | manager-develop (tdd) | **Defaults trim**: `NewDefaultGitConventionConfig`에서 `Validation.{Enabled, EnforceOnCommit}`, `Formatting`, `Custom` seed 제거(retained 필드만). `TestNewDefaultGitConventionConfig`(defaults_test.go:570+) fail-loud 갱신. (audit step 2) | RED(defaults_test fail-loud) → GREEN | GCM-9, REQ-WC9-005 | internal/config/defaults.go:456-478, defaults_test.go |
| **M3** | manager-develop (tdd) | **Validator trim + custom enum 4-site 제거**: `validGitConventionNames`에서 `custom` 제거 + 메시지 정정(192) + custom-required block(226-232) 제거; web `conventionCanonical`(validate.go:57) custom 제거; CLI wizard `huh.NewOption("custom",...)`(profile_setup.go:511) 제거. (4-site lockstep — audit step 3) | RED(custom 부재/거부 단언) → GREEN | GCW-1/GCR-2, REQ-WC9-003 | internal/config/validation.go, internal/web/validate.go, internal/cli/profile_setup.go, validation_test.go |
| **M4** | manager-develop (tdd) | **Template + live YAML nested 재작성(Fix C)**: 템플릿 `git-convention.yaml`을 trimmed nested shape로 재작성(`convention`, `auto_detection.{enabled,sample_size,confidence_threshold,fallback}`, `validation.{enforce_on_push,max_length}`), `max_length`=100, 주석 generic. local YAML도 trimmed nested로 정렬. `make build`로 embedded.go 재생성. (audit step 4) | RED(nested 단언/flat 부재) → GREEN | GCM-1/GCM-2/GCR-1, REQ-WC9-006, HARD-9 | template git-convention.yaml, .moai/config/sections/git-convention.yaml, embedded.go(생성), neutrality test |
| **M5** | manager-develop (tdd) | **Runtime wiring Fix A**: `LoadConvention`이 `AutoDetectionConfig`를 honor — `enabled` 게이트, `sample_size`→`Detect()`, `confidence_threshold` gate, configured `fallback`. 시그니처 변경 cascade를 유일 caller(hook_pre_push.go:54) + manager_test.go에 전파. dead `LoadFromConfig` + custom 테스트 삭제(GCR-4). **`Parse`/built-in 엔진/`Detect()` 시그니처 무변경 확인.** (audit step 5 part 1 + Custom removal) | RED(auto-detection honor 단언) → GREEN → REFACTOR | GCM-4/5/6, GCR-3/GCR-4, REQ-WC9-008/004/018, HARD-5/HARD-6 | internal/git/convention/manager.go, detector.go(read-only), hook_pre_push.go, manager_test.go, parser_test.go |
| **M6** | manager-develop (tdd) | **Runtime wiring Fix B**: `SetMaxLength` setter 신규(`Convention` 또는 `Manager`) + LoadConvention 후 `validation.max_length` forward. (audit step 5 part 2) | RED(max_length forward 단언) → GREEN | GCR-7, REQ-WC9-009 | internal/git/convention/manager.go(or convention.go), hook_pre_push.go, manager_test.go |
| **M7** | manager-develop (tdd) | **Web 재구성**: `custom.pattern` 위젯(fieldsets.templ:247) + view-model(handlers.go:52/176/202) + nested-form parse(projectconfig.go custom 경로) 제거; `auto_detection.sample_size` number field + `validation.enforce_on_push` toggle 추가(view-model + nested-form + write seam); 빈 옵션 라벨 git_convention select `"(project default)"`→`"(unchanged)"`(241); 카운트 라벨 `"8 fields"`→`"9 fields"`(234); convention≠auto일 때 detection sub-field grey(client hint). `templ generate`. (audit step 7) | RED(위젯 부재/존재/parity) → GREEN | GCW-3/GCW-5/GCW-7/GCR-9, REQ-WC9-010/011/012/013 | fieldsets.templ, handlers.go, projectconfig.go, validate.go(view list), fieldsets_templ.go(생성), web characterization test |
| **M8** | manager-develop (tdd) | **symmetry CI guard 추가(LAST)**: `GitConventionConfig`를 `audit_struct_yaml_symmetry_test.go` symmetryCases에 추가(struct trim + template nested 재작성으로 이미 대칭). 006 sentinel 무수정 확인 + statusline 무변경 git diff + GCR-5 파일 무변경 git diff + full suite + `templ generate` drift-free. REFACTOR(위젯 chrome 정리). (audit step 6 — sequence note에 따라 마지막) | RED(symmetry) → GREEN → REFACTOR | GCM-3, REQ-WC9-007, HARD-3/HARD-4/HARD-10 | internal/config/audit_struct_yaml_symmetry_test.go, integration_test.go(READ-only 단언), 전체 |

> Round 분할 없음(Tier L이나 8 milestone < 30 task — SSE stall 임계 미달, Milestone만 사용).
> GCR-5(audit step 9)는 본 milestone 집합에 **없음** — §F deferred(maintainer-decision 의존).

## §H. Cross-References

- spec.md §B(HARD) / §D(변경 인벤토리) / §E(테스트 영향) / §F(Exclusions + GCR-5 deferred 메모)
- `.moai/reports/web-console-statusline-gitconvention-audit.md`(전체 file:line + 8-step plan + Test & Sentinel Impact — Tier L design/research 대체)
- acceptance.md (AC-WC9-NNN)
- SPEC-WEB-CONSOLE-008 plan.md(statusline honest-hybrid 선례 — REMOVE/WIRE/FIX milestone 구조)

## §I. Anti-Patterns (회피 목록)

- `custom` 엔진 wiring/부활(HARD-1 위반) → 제거만(audit policy decision 1).
- convention 패키지 내부 `Parse`/`ParseBuiltin`/built-in templates 변경(B1 위반) → built-in 엔진 PRESERVE, `custom` 제거는 models 레이어 + dead LoadFromConfig만.
- `Detect()` 시그니처 변경(HARD-5 위반) → LoadConvention의 인자 전달만 변경.
- `validation.enforce_on_push` 삭제 / `isEnforceOnPushEnabled()` 변경(HARD-7/HARD-5 위반) → enforce_on_push는 live gate, 보존.
- GCR-5 wiring(.git_hooks/pre-push, hook_install.go prePushHookContent, git-strategy.hooks.pre_push)(HARD-4 위반) → §F deferred, git diff 0줄.
- statusline 손대기(HARD-2 위반, 008 스코프) → git diff 0줄.
- 006 sentinel(integration_test.go:191-205) 수정(HARD-3 위반) → 무수정, 강제 시 blocker.
- symmetry case를 struct trim/template 재작성 전에 추가(HARD-10/B8 위반) → M8(LAST) 순서 준수.
- 신규 git_convention validator 함수 신설(scope creep) → 기존 `validateGitConventionConfig` trim만.
- defaults_test fail-loud를 "stale baseline"으로 합리화(B12) → 제거된 필드 단언을 trimmed shape로 갱신.
