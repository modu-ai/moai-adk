# SPEC-WEB-CONSOLE-009 — Acceptance Criteria

> 각 AC는 GREEN-verifiable(precise single-file grep / `go test -run`) + REQ 매핑 + defect id + TDD class.
> cycle_type=tdd → 신규 거동은 RED→GREEN. 모든 명령은 repo root(`/Users/goos/MoAI/moai-adk-go`)에서 실행.
> **multi-file grep-c idiom 회피**(006/007 plan-audit 결함): 각 AC는 단일 파일 grep + comment 제외, 또는 `go test -run` 단언으로 검증. comment-naming false-fail 방지를 위해 grep은 주석 prefix(`//`/`#`) 제외.
> characterization/sentinel gate(AC-WC9-016/-017/-018)는 "must remain green" — 무수정 회귀 가드.

---

## §D AC Matrix

### REMOVE — custom 엔진 / config theater 제거 (Class B/C — source-coupled)

- **AC-WC9-001** (REQ-WC9-002, GCM-8) — `models.GitConventionConfig`에서 `Custom`/`Formatting` 필드 + `CustomConventionConfig`/`FormattingConfig` struct 부재.
  ```bash
  test "$(grep -cE '^[[:space:]]*(Custom|Formatting)[[:space:]]+[A-Z]' pkg/models/config.go)" -eq 0 && \
  test "$(grep -cE 'type (CustomConventionConfig|FormattingConfig) struct' pkg/models/config.go)" -eq 0
  ```
  단언: `GitConventionConfig`에 Custom/Formatting 필드 0 + 두 sub-struct 정의 0.

- **AC-WC9-002** (REQ-WC9-002, GCW-6/GCR-6) — `ConventionValidationConfig`에서 `Enabled`/`EnforceOnCommit` 부재; `EnforceOnPush`/`MaxLength` 보존.
  ```bash
  go test ./pkg/models/ -run 'TestGitConventionConfig|TestConventionValidationConfig' -count=1
  ```
  단언: struct가 `{EnforceOnPush, MaxLength}`만 — Enabled/EnforceOnCommit 제거; 컴파일·테스트 GREEN.

- **AC-WC9-003** (REQ-WC9-003, GCW-1/GCR-2) — convention oneof 태그에 `custom` 부재(4-site lockstep #1: models).
  ```bash
  grep -qE 'oneof=auto conventional-commits angular karma([^ ]|$)' pkg/models/config.go && \
  test "$(grep -E 'oneof=auto conventional-commits angular karma' pkg/models/config.go | grep -c custom)" -eq 0
  ```
  단언: oneof 태그가 `custom` 미포함(4-value).

- **AC-WC9-004** (REQ-WC9-003, GCW-1/GCR-2) — validator map + web slice + CLI wizard에 `custom` 부재(4-site lockstep #2/#3/#4) + custom-required block 제거.
  ```bash
  test "$(grep -E '"custom"' internal/config/validation.go | grep -vE '^[[:space:]]*//' | wc -l | tr -d ' ')" -eq 0 && \
  test "$(grep -E '"custom"' internal/web/validate.go | grep -vE '^[[:space:]]*//' | wc -l | tr -d ' ')" -eq 0 && \
  test "$(grep -E 'NewOption\("custom"' internal/cli/profile_setup.go | wc -l | tr -d ' ')" -eq 0
  ```
  단언: validator map / web `conventionCanonical` / CLI wizard 옵션에서 `custom` 0(comment 제외). custom-required block(`pattern is required when convention is 'custom'`) 부재는 AC-WC9-009 validator 테스트로 보강.

- **AC-WC9-005** (REQ-WC9-004, GCR-4) — dead `LoadFromConfig` + custom 테스트 refs 부재(production caller 0이었음).
  ```bash
  test "$(grep -rnE 'func .*LoadFromConfig' internal/git/convention/*.go | grep -v '_test.go' | wc -l | tr -d ' ')" -eq 0 && \
  test "$(grep -rnE 'LoadFromConfig' internal/ --include='*.go' | grep -vE ':[0-9]+:[[:space:]]*//' | wc -l | tr -d ' ')" -eq 0
  ```
  단언: `LoadFromConfig` 정의 0 + 전 패키지 호출/참조 0(comment 제외).

- **AC-WC9-006** (REQ-WC9-002/005, GCM-9) — defaults에서 `Validation.{Enabled, EnforceOnCommit}`/`Formatting`/`Custom` seed 부재; retained 필드만.
  ```bash
  go test ./internal/config/ -run 'TestNewDefaultGitConventionConfig' -count=1
  ```
  단언: `NewDefaultGitConventionConfig`가 `{Convention, AutoDetection{Enabled,SampleSize,ConfidenceThreshold,Fallback}, Validation{EnforceOnPush,MaxLength}}`만 반환; defaults_test fail-loud 갱신 후 GREEN.

- **AC-WC9-007** (REQ-WC9-010, GCR-9) — 웹 `git_convention.custom.pattern` 위젯 + view-model + nested-form parse 부재.
  ```bash
  test "$(grep -rnE 'custom\.pattern|CustomPattern' internal/web/*.go internal/web/*.templ | grep -vE ':[0-9]+:[[:space:]]*(//|/\*|\*)' | wc -l | tr -d ' ')" -eq 0
  ```
  D-fix: comment 라인 제외 — 제거 설명 주석이 토큰 명명 시 false-fail 방지.

### WIRE/FIX — live lever 배선 + drift 교정 (Class B/A)

- **AC-WC9-008** (REQ-WC9-008, GCM-4/5/6, GCR-3) — **Fix A**: `LoadConvention`이 `AutoDetectionConfig`를 honor(enabled gate / sample_size→Detect / confidence_threshold gate / configured fallback).
  ```bash
  go test ./internal/git/convention/ -run 'TestLoadConvention.*AutoDetect|TestLoadConvention.*SampleSize|TestLoadConvention.*Confidence|TestLoadConvention.*Fallback' -count=1
  ```
  단언: `enabled=false`면 Detect 미호출; `sample_size`가 Detect에 전달(100 하드코딩 아님); confidence < threshold면 fallback; fallback이 configured 값. 유일 production caller(hook_pre_push.go:54) 새 시그니처로 컴파일.

- **AC-WC9-009** (REQ-WC9-003, GCR-2) — validator가 `custom`을 거부 + custom-required block 부재.
  ```bash
  go test ./internal/config/ -run 'TestValidateGitConventionConfig|TestIsValidConvention' -count=1 && \
  test "$(grep -E "convention is 'custom'|required when convention" internal/config/validation.go | grep -vE '^[[:space:]]*//' | wc -l | tr -d ' ')" -eq 0
  ```
  단언: `IsValidConvention("custom")`==false; custom-required 메시지 부재(comment 제외).

- **AC-WC9-010** (REQ-WC9-009, GCR-7) — **Fix B**: `SetMaxLength` setter 존재 + LoadConvention 후 `validation.max_length` forward.
  ```bash
  grep -qE 'func .*SetMaxLength\(' internal/git/convention/*.go && \
  go test ./internal/git/convention/ -run 'TestSetMaxLength|TestLoadConvention.*MaxLength' -count=1
  ```
  단언: `SetMaxLength(int)` 정의 존재; 적용 후 `Convention.MaxLength`가 configured 값(built-in 100 override).

- **AC-WC9-011** (REQ-WC9-006, GCM-1/GCM-2/GCR-1) — 템플릿 `git-convention.yaml`이 nested shape + `max_length`=100; flat 키 부재.
  ```bash
  grep -qE '^[[:space:]]*auto_detection:' internal/template/templates/.moai/config/sections/git-convention.yaml && \
  grep -qE '^[[:space:]]*validation:' internal/template/templates/.moai/config/sections/git-convention.yaml && \
  test "$(grep -cE '^[[:space:]]*(auto_detect|enforce_on_push):[[:space:]]' internal/template/templates/.moai/config/sections/git-convention.yaml)" -eq 0 && \
  test "$(grep -cE '^[[:space:]]*max_length:[[:space:]]*72' internal/template/templates/.moai/config/sections/git-convention.yaml)" -eq 0
  ```
  단언: nested `auto_detection:`/`validation:` 존재; flat top-level `auto_detect:`/`enforce_on_push:` 부재(nested로 이동); `max_length: 72` 부재(100으로 표준화). (`enforce_on_push`는 `validation:` 하위로만 존재해야 하므로 top-level grep으로 flat 부재 확인.)

- **AC-WC9-012** (REQ-WC9-001, GCM-7/GCW-6/GCR-6/GCR-8) — 템플릿·local YAML에 `formatting.*`/`validation.enabled`/`validation.enforce_on_commit`/`custom.*` 부재.
  ```bash
  for f in internal/template/templates/.moai/config/sections/git-convention.yaml .moai/config/sections/git-convention.yaml; do \
    test "$(grep -cE '^[[:space:]]*(formatting|custom):' "$f")" -eq 0 && \
    test "$(grep -cE '^[[:space:]]*(enabled|enforce_on_commit):' "$f")" -eq 0 || exit 1; \
  done
  ```
  단언: 두 YAML 모두 formatting/custom 블록 0 + validation.{enabled,enforce_on_commit} 0. (`auto_detection.enabled`는 `enabled:` 매칭되나 `^[[:space:]]*enabled:`이 nested indent로 잡힘 — 주의: auto_detection.enabled는 유지 대상이므로 이 grep은 false-fail 위험. **D-fix 적용 → 아래 정밀 버전**.)
  ```bash
  # D-fix 정밀 버전: validation 블록의 enabled/enforce_on_commit만 검증(auto_detection.enabled는 보존)
  go test ./internal/config/ -run 'TestGitConventionYAMLLoad|TestLoadGitConvention' -count=1
  ```
  단언(정밀): trimmed YAML 로드 시 retained 필드만 unmarshal(auto_detection.enabled 보존, validation.enabled/enforce_on_commit/formatting/custom 부재). YAML-load 테스트로 검증(grep indent 모호성 회피).

### Web 재구성 + 라벨 (Class A — markup parity)

- **AC-WC9-013** (REQ-WC9-010, GCW-3/GCW-5) — 웹에 `sample_size` number field + `enforce_on_push` toggle 추가.
  ```bash
  grep -qE 'git_convention\.auto_detection\.sample_size' internal/web/fieldsets.templ && \
  grep -qE 'git_convention\.validation\.enforce_on_push' internal/web/fieldsets.templ
  ```
  단언: 두 신규 위젯의 form name이 fieldsets.templ에 존재.

- **AC-WC9-014** (REQ-WC9-011/013, GCW-7) — 카운트 라벨 `"9 fields"` + git_convention select 빈 옵션 `"(unchanged)"`.
  ```bash
  grep -qE '9 fields' internal/web/fieldsets.templ && \
  test "$(grep -c '8 fields' internal/web/fieldsets.templ)" -eq 0 && \
  grep -qE 'Empty:[[:space:]]*"\(unchanged\)"' internal/web/fieldsets.templ
  ```
  단언: count.project 라벨 `"9 fields"`; `"8 fields"` 0; git_convention optSelect의 `Empty:"(unchanged)"`. (model select의 `(project default)`는 다른 필드이므로 무관.)

- **AC-WC9-015** (REQ-WC9-012, GCR-3 UX) — convention≠auto일 때 detection sub-field grey + `templ generate` drift-free.
  ```bash
  go test ./internal/web/ -run 'TestFieldsetProject.*Convention|TestProjectNested.*RoundTrip' -count=1 && \
  templ generate && git diff --exit-code internal/web/*_templ.go
  ```
  단언: 렌더 markup이 convention≠auto일 때 detection sub-field에 de-emphasis 속성(client hint); nested round-trip(sample_size/enforce_on_push 추가 + custom.pattern 제거) 보존; codegen 최신.

### symmetry CI guard (Class B — 신규, LAST)

- **AC-WC9-016** (REQ-WC9-007, GCM-3) — `GitConventionConfig`가 symmetryCases에 포함 + symmetry 테스트 GREEN. [LAST gate]
  ```bash
  grep -qE 'GitConventionConfig\{\}' internal/config/audit_struct_yaml_symmetry_test.go && \
  go test ./internal/config/ -run 'TestStructYAMLSymmetry|TestSymmetry' -count=1
  ```
  단언: symmetryCases에 `{structType: reflect.TypeOf(models.GitConventionConfig{}), templateYAML: "git-convention.yaml", yamlTopKey: "git_convention"}` 추가 + struct↔YAML 대칭 통과(M1 struct trim + M4 template nested 재작성 후 clean — HARD-10 순서 게이트).

### scope guard — statusline 무관 + GCR-5 무변경 + enforce_on_push 보존 (Class B — boundary)

- **AC-WC9-017** (REQ-WC9-015, HARD-2) — statusline 무변경: statusline 관련 파일/struct/검증/위젯 변경 0.
  ```bash
  git diff origin/main -- internal/statusline/ internal/profile/sync.go internal/cli/statusline.go | \
    grep -E '^\+' | grep -icE 'statusline|StatuslineConfig|preset|segment' | grep -qx 0
  ```
  단언: 추가 라인에 statusline 관련 토큰 변경 0(008 스코프 침범 0). [MUST remain green]

- **AC-WC9-018** (REQ-WC9-014, HARD-4) — **GCR-5 wiring 무변경**: .git_hooks/pre-push / hook_install.go prePushHookContent / git-strategy.yaml hooks.pre_push 변경 0.
  ```bash
  test "$(git diff origin/main -- .git_hooks/pre-push internal/cli/hook_install.go internal/template/templates/.moai/config/sections/git-strategy.yaml | grep -cE '^[+-][^+-]')" -eq 0
  ```
  단언: GCR-5 deferred 경계의 세 파일에 diff 0줄(§F 침범 0). [MUST remain green]

- **AC-WC9-019** (REQ-WC9-017, HARD-7/HARD-5) — `validation.enforce_on_push` + `isEnforceOnPushEnabled()` read 경로 보존.
  ```bash
  grep -qE 'EnforceOnPush' pkg/models/config.go && \
  grep -qE 'cfg\.GitConvention\.Validation\.EnforceOnPush' internal/cli/hook_pre_push.go && \
  go test ./internal/cli/ -run 'TestIsEnforceOnPushEnabled|TestRunPrePush' -count=1
  ```
  단언: `EnforceOnPush` 필드 보존; `isEnforceOnPushEnabled`가 nested 값 읽는 경로 무변경 + 테스트 GREEN.

### 006 boundary sentinel 무수정 (Class B — 무수정 sentinel gate)

- **AC-WC9-020** (REQ-WC9-015, HARD-3) — **006 scope-boundary sentinel(integration_test.go:191-205) 무수정 GREEN**: workflow/harness/git-strategy `DO_NOT_TOUCH` byte-identical.
  ```bash
  go test ./internal/web/ -run 'TestRealServerRoundTrip' -count=1 && \
  git diff origin/main -- internal/web/integration_test.go | grep -E '^\+' | grep -icE 'workflow|harness|git-strategy|DO_NOT_TOUCH' | grep -qx 0
  ```
  단언: sentinel 로직 라인(191-205) 무수정; 추가 라인에 workflow/harness/git-strategy 변경 0. (git_convention round-trip 테스트 갱신은 sentinel 라인이 아니므로 허용.) [MUST remain green]

### offline / HTMX 계약 + 회귀 가드 (Class B)

- **AC-WC9-021** (REQ-WC9-016, HARD-8) — **offline zero-network 무수정**: 신규 CDN/외부 asset 0.
  ```bash
  test "$(grep -rcE 'unpkg\.com|jsdelivr|cdn\.|googleapis\.com|fonts\.gstatic' internal/web/*.templ internal/web/assets/*.css internal/web/assets/*.js | awk -F: '{s+=$2} END{print s+0}')" -eq 0
  ```
  [MUST remain green]

- **AC-WC9-022** (REQ-WC9-016) — web 레이어 직접 YAML write 금지(grep cleanliness).
  ```bash
  test "$(grep -rnE 'yaml\.Marshal\(|os\.WriteFile\(' internal/web/handlers.go internal/web/validate.go internal/web/projectconfig.go | grep -vE ':[0-9]+:[[:space:]]*//' | wc -l | tr -d ' ')" -eq 0
  ```
  단언: 영속은 `SetSection`/`Save` 통한다(직접 marshal/write 0, comment 제외).

- **AC-WC9-023** (REQ-WC9-006, HARD-9) — `make build` 후 embedded.go가 template nested 재작성 반영 + neutrality.
  ```bash
  make build && git diff --exit-code internal/template/embedded.go || echo "embedded.go regenerated (commit it)"
  go test ./internal/template/ -run 'TestTemplateNeutralityAudit|TestEmbedded' -count=1
  ```
  단언: embedded.go가 nested shape 반영; template neutrality(주석에 SPEC ID/REQ 토큰 0) 통과.

- **AC-WC9-024** (REQ-WC9-018, HARD-6) — `LoadConvention` cascade scope 확정: production caller 1개(hook_pre_push.go)만.
  ```bash
  test "$(grep -rnE '\.LoadConvention\(' internal/ --include='*.go' | grep -v '_test.go' | grep -vE ':[0-9]+:[[:space:]]*//' | wc -l | tr -d ' ')" -eq 1
  ```
  단언: `LoadConvention` production caller 정확히 1(hook_pre_push.go:54). 다른 caller 0 확인(Fix A cascade 통제).

- **AC-WC9-025** (전 REQ) — cross-platform build.
  ```bash
  go build ./... && GOOS=windows GOARCH=amd64 go build ./...
  ```

- **AC-WC9-026** (전 REQ) — full suite GREEN + 커버리지 baseline 유지.
  ```bash
  go test ./internal/web/... ./internal/cli/... ./internal/config/... ./internal/git/convention/... ./pkg/models/... -count=1 && \
  go test ./internal/web/... ./internal/git/convention/... -cover -count=1
  ```
  단언: 0 FAIL; internal/web + internal/git/convention 커버리지 ≥ 006/007/008 baseline 이내(하락 없음).

---

## §D.1 Severity

- **MUST-PASS (closure gate)**: AC-WC9-001, -002, -003, -004, -005, -006, -007, -008, -009, -010, -011, -012, -016, -017, -018(GCR-5 무변경), -019, -020(sentinel), -021(offline), -022, -023, -024, -025, -026. (custom 제거 완전성, struct trim, defaults trim, Fix A/B 배선, template nested, validator trim, symmetry guard, statusline 무관, GCR-5 무변경, enforce_on_push 보존, sentinel 무수정, offline 무수정, cascade 통제, full suite.)
- **SHOULD-PASS**: AC-WC9-013, -014, -015. (web sample_size/enforce_on_push 추가, 라벨, 조건부 grey UX.)

## §D.2 Characterization / Sentinel "Must Remain Green" Gates (명시)

본 4개는 변경의 부작용 부재를 증명하는 무수정 회귀 가드 — 어느 하나라도 RED면 closure 차단:

1. **AC-WC9-017** — statusline 무변경: git_convention 변경이 008 스코프(statusline)를 침범하지 않음(HARD-2).
2. **AC-WC9-018** — GCR-5 deferred 경계 무변경: .git_hooks/pre-push / hook_install.go prePushHookContent / git-strategy.yaml hooks.pre_push byte-identical(HARD-4).
3. **AC-WC9-020** — 006 boundary sentinel(integration_test.go:191-205): workflow/harness/git-strategy byte-identical(HARD-3).
4. **AC-WC9-021** — offline zero-network: 신규 CDN/외부 asset 0(HARD-8).

추가로 **AC-WC9-019**(enforce_on_push live gate 보존)는 honest-hybrid가 live lever를 보존함을 증명하는 양성 가드.

## §D.3 REQ ↔ AC Traceability

| REQ | AC |
|-----|-----|
| REQ-WC9-001 | -012 |
| REQ-WC9-002 | -001, -002, -003, -006 |
| REQ-WC9-003 | -003, -004, -009 |
| REQ-WC9-004 | -005 |
| REQ-WC9-005 | -006 |
| REQ-WC9-006 | -011, -023 |
| REQ-WC9-007 | -016 |
| REQ-WC9-008 | -008 |
| REQ-WC9-009 | -010 |
| REQ-WC9-010 | -007, -013 |
| REQ-WC9-011 | -014 |
| REQ-WC9-012 | -015 |
| REQ-WC9-013 | -014 |
| REQ-WC9-014 | -018 |
| REQ-WC9-015 | -017, -020 |
| REQ-WC9-016 | -021, -022 |
| REQ-WC9-017 | -019 |
| REQ-WC9-018 | -008, -024 |

## §D.4 Defect → REQ → AC 커버리지 맵

| audit defect | REQ | AC |
|--------------|-----|-----|
| GCM-1 (template flat never unmarshals) | REQ-WC9-006 | -011, -023 |
| GCM-2 (max_length 72 wrong path/value) | REQ-WC9-006 | -011 |
| GCM-3 (no symmetry test) | REQ-WC9-007 | -016 |
| GCM-4 (sample_size ignored — Detect hardcodes 100) | REQ-WC9-008 | -008 |
| GCM-5 (confidence_threshold never gates) | REQ-WC9-008 | -008 |
| GCM-6 (enabled never read) | REQ-WC9-008 | -008 |
| GCM-7 (formatting.* zero consumers) | REQ-WC9-001 | -012 |
| GCM-8 (custom sub-fields never compiled) | REQ-WC9-002/003 | -001 |
| GCM-9 (defaults omits Custom; trim removed fields) | REQ-WC9-005 | -006 |
| GCW-1 (web accepts custom but runtime can't load) | REQ-WC9-003 | -003, -004 |
| GCW-2 (confidence_threshold web-editable runtime-ignored) | REQ-WC9-008 | -008 (now wired) |
| GCW-3 (enforce_on_push not in web UI) | REQ-WC9-010 | -013 |
| GCW-4 (auto_detection.enabled web-editable no consumer) | REQ-WC9-008 | -008 (now wired) |
| GCW-5 (sample_size + fallback unexposed/hardcoded) | REQ-WC9-008/010 | -008, -013 |
| GCW-6 (validation.enforce_on_commit no consumer) | REQ-WC9-001 | -012 |
| GCW-7 ("(project default)" misleading) | REQ-WC9-013 | -014 |
| GCR-1 (template flat silently fails — runtime angle) | REQ-WC9-006 | -011 |
| GCR-2 (convention=custom rejected at runtime) | REQ-WC9-003 | -004, -009 |
| GCR-3 (auto-detection partial theater) | REQ-WC9-008/012 | -008, -015 |
| GCR-4 (custom.* dead — LoadFromConfig no caller) | REQ-WC9-004 | -005 |
| GCR-5 (deployed pre-push never invokes handler) | (deferred — §F) | -018 (무변경 guard) |
| GCR-6 (validation.enforce_on_commit + commit-msg path absent) | REQ-WC9-001 | -012 |
| GCR-7 (validation.max_length unwired) | REQ-WC9-009 | -010 |
| GCR-8 (formatting.* zero consumers — runtime angle) | REQ-WC9-001 | -012 |
| GCR-9 (web writes enabled/confidence/custom.pattern — dead) | REQ-WC9-008/010 | -007, -008 |

## §D.5 Definition of Done

- [ ] MUST-PASS AC 전부 GREEN
- [ ] custom 엔진 제거 완전(4-site enum + struct + YAML + 웹 위젯 + CLI 옵션 + dead LoadFromConfig — AC-WC9-001/-003/-004/-005/-007)
- [ ] struct trim(Custom/Formatting/Validation.{Enabled,EnforceOnCommit}) + defaults trim(AC-WC9-001/-002/-006)
- [ ] Fix A(auto-detection 4-knob honor, AC-WC9-008) + Fix B(SetMaxLength forward, AC-WC9-010)
- [ ] Fix C(template flat→nested + max_length=100 + neutrality, AC-WC9-011/-012/-023)
- [ ] symmetry CI guard 추가(AC-WC9-016, LAST — struct trim + template nested 후)
- [ ] `enforce_on_push` live gate 보존(AC-WC9-019) — 삭제는 validation.{enabled,enforce_on_commit}만
- [ ] LoadConvention cascade 통제(production caller 1개, AC-WC9-024)
- [ ] statusline 무관(AC-WC9-017) + GCR-5 deferred 경계 무변경(AC-WC9-018) + 006 sentinel 무수정(AC-WC9-020) + offline 무수정(AC-WC9-021)
- [ ] web sample_size + enforce_on_push 추가 + "(unchanged)" 라벨 + "9 fields"(AC-WC9-013/-014) + 조건부 grey(AC-WC9-015)
- [ ] `templ generate` drift-free(AC-WC9-015) + `make build` embedded.go 반영(AC-WC9-023)
- [ ] cross-platform build(AC-WC9-025) + full suite GREEN + 커버리지 baseline(AC-WC9-026)
- [ ] spec.md §F Exclusions의 어느 항목도 침범하지 않음(특히 GCR-5 wiring / custom 부활 / statusline / enforce_on_push 삭제)
