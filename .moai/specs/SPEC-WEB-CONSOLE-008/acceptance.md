# SPEC-WEB-CONSOLE-008 — Acceptance Criteria

> 각 AC는 GREEN-verifiable(grep/test command) + REQ 매핑 + defect id + TDD class.
> cycle_type=tdd → 신규 거동은 RED→GREEN. 모든 명령은 repo root(`/Users/goos/MoAI/moai-adk-go`)에서 실행.
> 3개 characterization/sentinel gate(AC-WC8-016/-017/-018)는 "must remain green" — 무수정 회귀 가드.

---

## §D AC Matrix

### REMOVE — config theater 제거 (Class B/C — source-coupled)

- **AC-WC8-001** (REQ-WC8-001/005, SLM-3/SLR-6) — 템플릿 `statusline.yaml`에 `mode:`/`refresh_interval:` 부재.
  ```bash
  test "$(grep -cE '^[[:space:]]*(mode|refresh_interval):' internal/template/templates/.moai/config/sections/statusline.yaml)" -eq 0
  ```

- **AC-WC8-002** (REQ-WC8-005, SLR-4) — 템플릿 theme seed가 `catppuccin-mocha`; 주석은 존재하는 2개 테마만 나열.
  ```bash
  grep -qE '^[[:space:]]*theme:[[:space:]]*"?catppuccin-mocha"?' internal/template/templates/.moai/config/sections/statusline.yaml && \
  test "$(grep -iE 'theme' internal/template/templates/.moai/config/sections/statusline.yaml | grep -c 'default')" -eq 0
  ```
  단언: theme seed=catppuccin-mocha; theme 라인/주석에 미존재 테마 `default` 미언급(SLR-4). D1-fix: `minimal`은 preset 名으로 정당 잔존하므로 grep을 theme 라인으로 한정(`grep theme | grep -c default`).

- **AC-WC8-003** (REQ-WC8-002, SLM-1/SLM-2/SLR-2, HARD-5) — 3-struct 통합: 2개 private struct에 `Mode` 부재 + canonical 모델에 `Mode` 미추가.
  ```bash
  test "$(grep -cE '^[[:space:]]*Mode[[:space:]]' internal/profile/sync.go)" -eq 0 && \
  test "$(grep -cE 'Mode[[:space:]]+string' internal/cli/statusline.go)" -eq 0 && \
  grep -qE 'Preset[[:space:]]+string' pkg/models/config.go && \
  grep -qE 'func NormalizeMode' internal/statusline/types.go && \
  grep -qE 'func .*Render\(' internal/statusline/renderer.go
  ```
  단언: `statuslineData`/`statuslineFileConfig`에 Mode 필드 0; `models.StatuslineConfig`는 `{Preset, Segments, Theme}` 유지(Mode 미추가). D5-fix: Builder API 보존 **양성 검증** 추가 — `NormalizeMode`(types.go) + `Render(`(renderer.go) 존재 확인(REQ-WC8-003 HARD-2/HARD-5; config surface 제거가 enum/Render 시그니처를 깨지 않음을 양성 단언).

- **AC-WC8-004** (REQ-WC8-009, SLR-1) — 웹 `statusline_mode` control + parse + 검증 부재.
  ```bash
  test "$(grep -rnE 'statusline_mode|statuslineModeCanonical|StatuslineModes' internal/web/*.go internal/web/*.templ | grep -vE ':[0-9]+:[[:space:]]*(//|/\*|\*)' | wc -l | tr -d ' ')" -eq 0
  ```
  D2-fix: comment 라인 제외(`grep -v` 주석 prefix) — 잔존 `@MX:NOTE` 주석이 토큰 명명 시 false-fail 방지.

- **AC-WC8-005** (REQ-WC8-012, SLR-7) — dead `loadSegmentConfig` + 테스트 refs 부재.
  ```bash
  test "$(grep -rnE 'loadSegmentConfig' internal/cli/*.go | grep -vE ':[0-9]+:[[:space:]]*(//|/\*|\*)' | wc -l | tr -d ' ')" -eq 0
  ```
  D2-fix: comment 라인 제외 — 제거 설명 주석이 토큰 명명 시 false-fail 방지.

- **AC-WC8-006** (REQ-WC8-013, HARD-2) — `renderFullV3` 호출 0(부활 금지). 정의 자체가 nolint:unused로 남거나 제거되어도 무방하나 **production 호출 사이트 0**.
  ```bash
  test "$(grep -rnE 'renderFullV3\(' internal/statusline/*.go | grep -v '_test.go' | grep -vE 'func .*renderFullV3' | wc -l | tr -d ' ')" -eq 0
  ```

### WIRE/FIX — live lever + drift 교정 (Class B/A)

- **AC-WC8-007** (REQ-WC8-004, SLM-5) — `defaultStatuslineSegments()`가 canonical 15-key 반환(presetToSegments("full", nil) 위임).
  ```bash
  go test ./internal/profile/ -run 'TestDefaultStatuslineSegments|TestSyncStatusline' -count=1
  ```
  단언: 반환 map이 15 키(model/context/output_style/claude_version/moai_version/session_time/effort_thinking/usage_5h/usage_7d/directory/git_status/git_branch/worktree/task/pr) 전부 포함.

- **AC-WC8-008** (REQ-WC8-007, SLR-3, HARD-6) — preset write-effective: 비-custom preset 저장 → 디스크 statusline.yaml에 15-key segments map 영속.
  ```bash
  go test ./internal/profile/ -run 'TestSyncStatusline.*PresetExpand|TestStatuslinePreset.*Effective' -count=1
  ```
  단언: preset=compact/minimal 저장 후 재로드 → 해당 preset의 15-key 확장 map이 segments에 영속(no-op 아님).

- **AC-WC8-009** (REQ-WC8-008, HARD-7) — **theme-only round-trip(characterization gate) 무수정 GREEN**: preset 미전송 시 기존 segments 보존.
  ```bash
  go test ./internal/web/ -run 'TestRealServerRoundTrip|TestGoldenPath' -count=1
  ```
  단언: integration_test.go:124-165 통과 무수정(theme만 변경, segments preserve). [MUST remain green]

- **AC-WC8-010** (REQ-WC8-017, SLR-5/SLM-7) — `builder.go` ThemeName doc가 2 real themes만; `SegmentRepo` intentional-exclusion 주석 존재.
  ```bash
  grep -qiE 'ThemeName.*catppuccin|catppuccin-(mocha|latte)' internal/statusline/builder.go && \
  test "$(grep -c 'nerd' internal/statusline/builder.go)" -eq 0 && \
  grep -qiE 'SegmentRepo.*(schema|intentional|excluded|out of schema|outside)' internal/statusline/types.go
  ```

### Web 조건부 노출 + 라벨 (Class A — markup parity)

- **AC-WC8-011** (REQ-WC8-016, SLW-4) — 카운트 라벨 `"3 fields"`→`"2 fields"`.
  ```bash
  grep -qE '2 fields' internal/web/fieldsets.templ && \
  test "$(grep -c '3 fields · segments' internal/web/fieldsets.templ)" -eq 0
  ```

- **AC-WC8-012** (REQ-WC8-010, SLR-3 UX) — segment 체크박스 조건부 편집(preset=custom일 때만): disabled/display-only when preset≠custom.
  ```bash
  go test ./internal/web/ -run 'TestFieldsetStatusline.*Conditional|TestSegment.*Custom' -count=1
  ```
  단언: 렌더된 markup이 preset≠custom일 때 segment checkbox에 disabled 속성(또는 동등 비편집 상태); preset=custom일 때 편집 가능.

- **AC-WC8-013** (REQ-WC8-015, SLR-3 UX) — `templ generate` drift-free(codegen 최신).
  ```bash
  templ generate && git diff --exit-code internal/web/*_templ.go
  ```

### symmetry CI guard (Class B — 신규, LAST)

- **AC-WC8-014** (REQ-WC8-006, SLM-4) — `StatuslineConfig`가 symmetryCases에 포함 + symmetry 테스트 GREEN.
  ```bash
  grep -q 'StatuslineConfig{}' internal/config/audit_struct_yaml_symmetry_test.go && \
  go test ./internal/config/ -run 'TestStructYAMLSymmetry|TestSymmetry' -count=1
  ```
  단언: symmetryCases에 `{structType: reflect.TypeOf(StatuslineConfig{}), templateYAML: "statusline.yaml", yamlTopKey: "statusline"}` 추가 + struct↔YAML 대칭 통과(M1 mode/refresh_interval 제거 + M3 struct Mode 제거 후 clean).

### scope guard — git_convention 무관 + 신규 validator 0 (Class B — boundary)

- **AC-WC8-015** (REQ-WC8-014, HARD-4) — git_convention 무변경: git_convention 관련 파일/struct/검증 변경 0.
  ```bash
  git diff origin/main -- internal/git/convention/ internal/config/validation.go pkg/models/config.go internal/web/validate.go | \
    grep -E '^\+' | grep -icE 'git_convention|GitConvention|convention|custom\.pattern|auto_detection' | grep -qx 0
  ```
  단언: 추가 라인에 git_convention 관련 토큰 변경 0(009 스코프 침범 0).

- **AC-WC8-016** (REQ-WC8-001, CRITICAL SCOPE) — 신규 validator 함수 0개 불변.
  ```bash
  test "$(grep -cE 'func validate(Workflow|GitStrategy|Harness|Llm|Statusline)Config' internal/config/validation.go)" -eq 0
  ```
  단언: 신규 statusline/workflow/git-strategy/harness/llm validator 함수 미생성(SLM-6 deferred — scope creep 회피).

### 006 boundary sentinel 무수정 (Class B — 무수정 sentinel gate)

- **AC-WC8-017** (REQ-WC8-014, HARD-3) — **006 scope-boundary sentinel(integration_test.go:197-205) 무수정 GREEN**: workflow/harness/git-strategy `DO_NOT_TOUCH` byte-identical.
  ```bash
  go test ./internal/web/ -run 'TestRealServerRoundTrip' -count=1 && \
  git diff origin/main -- internal/web/integration_test.go | grep -E '^\+' | grep -icE 'workflow|harness|git-strategy|DO_NOT_TOUCH' | grep -qx 0
  ```
  단언: sentinel 로직 라인(197-205) 무수정; 추가 라인에 workflow/harness/git-strategy 변경 0. [MUST remain green]

### offline / HTMX 계약 (Class B — 회귀 가드)

- **AC-WC8-018** (REQ-WC8-015, HARD-8) — **offline zero-network 무수정**: 신규 CDN/외부 asset 0.
  ```bash
  test "$(grep -rcE 'unpkg\.com|jsdelivr|cdn\.|googleapis\.com|fonts\.gstatic' internal/web/*.templ internal/web/assets/*.css internal/web/assets/*.js | awk -F: '{s+=$2} END{print s+0}')" -eq 0
  ```
  [MUST remain green]

- **AC-WC8-019** (REQ-WC8-015) — web 레이어 직접 YAML write 금지(grep cleanliness).
  ```bash
  test "$(grep -rnE 'yaml\.Marshal\(|os\.WriteFile\(' internal/web/handlers.go internal/web/validate.go | grep -vE ':[0-9]+:[[:space:]]*//' | wc -l | tr -d ' ')" -eq 0
  ```

### cross-platform + 전체 회귀 (Class B)

- **AC-WC8-020** (전 REQ) — cross-platform build.
  ```bash
  go build ./... && GOOS=windows GOARCH=amd64 go build ./...
  ```

- **AC-WC8-021** (REQ-WC8-005, HARD-9) — `make build` 후 embedded.go가 template 변경 반영(템플릿-임베드 drift 0).
  ```bash
  make build && git diff --exit-code internal/template/embedded.go || echo "embedded.go regenerated (commit it)"
  go test ./internal/template/ -run 'TestTemplateNeutralityAudit|TestEmbedded' -count=1
  ```
  단언: embedded.go가 mode/refresh_interval 제거 + theme seed 반영; template neutrality(주석에 SPEC ID/REQ 토큰 0) 통과.

- **AC-WC8-022** (전 REQ) — full suite GREEN + 커버리지 baseline 유지.
  ```bash
  go test ./internal/web/... ./internal/cli/... ./internal/config/... ./internal/profile/... ./internal/statusline/... ./pkg/models/... -count=1 && \
  go test ./internal/web/... -cover -count=1
  ```
  단언: 0 FAIL; internal/web 커버리지 ≥ 006/007 baseline 이내(하락 없음).

---

## §D.1 Severity

- **MUST-PASS (closure gate)**: AC-WC8-001, -003, -004, -005, -006, -007, -008, -009(characterization), -013, -014, -015, -016, -017(sentinel), -018(offline), -019, -020, -021, -022. (config theater 제거 완전성, struct 통합, preset 배선, 신규 validator 0, git_convention 무관, sentinel 무수정, offline 무수정, codegen drift-free, full suite.)
- **SHOULD-PASS**: AC-WC8-002, -010, -011, -012.

## §D.2 Characterization / Sentinel "Must Remain Green" Gates (명시)

본 3개는 변경의 부작용 부재를 증명하는 무수정 회귀 가드 — 어느 하나라도 RED면 closure 차단:

1. **AC-WC8-009** — theme-only round-trip(integration_test.go:124-165): preset write-effective(SLR-3) 변경이 preserve 경로(HARD-7)를 깨지 않음.
2. **AC-WC8-017** — 006 boundary sentinel(integration_test.go:197-205): workflow/harness/git-strategy byte-identical(HARD-3).
3. **AC-WC8-018** — offline zero-network: 신규 CDN/외부 asset 0(HARD-8).

추가로 **builder_test.go "every mode collapses to 3-line / full retired"**(spec.md §E)는 Builder API 무변경(HARD-1)으로 자연히 GREEN 유지 — M3에서 무수정 PASS 확인.

## §D.3 REQ ↔ AC Traceability

| REQ | AC |
|-----|-----|
| REQ-WC8-001 | -001, -016 |
| REQ-WC8-002 | -003 |
| REQ-WC8-003 | -003 (Builder-API 양성 grep: NormalizeMode + Render) + builder_test.go 무수정 PASS, spec.md §E |
| REQ-WC8-004 | -007 |
| REQ-WC8-005 | -001, -002, -021 |
| REQ-WC8-006 | -014 |
| REQ-WC8-007 | -008 |
| REQ-WC8-008 | -009 |
| REQ-WC8-009 | -004 |
| REQ-WC8-010 | -012 |
| REQ-WC8-011 | (coercion 안전망 유지 — -002 seed fix로 source 제거) |
| REQ-WC8-012 | -005 |
| REQ-WC8-013 | -006 |
| REQ-WC8-014 | -015, -017 |
| REQ-WC8-015 | -013, -018, -019 |
| REQ-WC8-016 | -011 |
| REQ-WC8-017 | -010 |

## §D.4 Defect → REQ → AC 커버리지 맵

| audit defect | REQ | AC |
|--------------|-----|-----|
| SLM-1 (canonical lacks Mode / loader drops mode) | REQ-WC8-002 | -003 |
| SLM-2 (3 divergent structs) | REQ-WC8-002 | -003 |
| SLM-3 (refresh_interval dead) | REQ-WC8-001/005 | -001 |
| SLM-4 (symmetry test omits StatuslineConfig) | REQ-WC8-006 | -014 |
| SLM-5 (seed 11 of 15) | REQ-WC8-004 | -007 |
| SLM-6 (no enum validator) | (deferred SHOULD — §F) | -016 (validator 0 guard) |
| SLM-7 (SegmentRepo 16th constant) | REQ-WC8-017 | -010 |
| SLW-4 (segment-map server validation low / "3 fields" label) | REQ-WC8-016 | -011 |
| SLW-6 (load-modify-write preserve by design) | REQ-WC8-008 | -009 |
| SLR-1 (mode=full inert / renderFullV3 never called) | REQ-WC8-009/013 | -004, -006 |
| SLR-2 (only CLI private struct reads mode) | REQ-WC8-002 | -003 |
| SLR-3 (preset near-vestigial — segments wins) | REQ-WC8-007/010 | -008, -012 |
| SLR-4 (theme "default" coerces; comment lists 3) | REQ-WC8-005/011 | -002 |
| SLR-5 (builder.go:50 doc wrong themes) | REQ-WC8-017 | -010 |
| SLR-6 (refresh_interval no consumer) | REQ-WC8-001 | -001 |
| SLR-7 (loadSegmentConfig dead) | REQ-WC8-012 | -005 |
| SLR-8 (working-tree drift) | REQ-WC8-005 (template seed fix; no code change) | -002, -021 |

## §D.5 Definition of Done

- [ ] MUST-PASS AC 전부 GREEN
- [ ] config theater 제거 완전(mode surface/refresh_interval/loadSegmentConfig — AC-WC8-001/-003/-004/-005)
- [ ] preset write-effective 배선(AC-WC8-008) + seed 15-key(AC-WC8-007)
- [ ] 3-struct 통합, canonical 모델에 Mode 미추가(AC-WC8-003)
- [ ] Builder API 무변경(builder_test.go 무수정 PASS)
- [ ] renderFullV3 부활 0(AC-WC8-006)
- [ ] symmetry CI guard 추가(AC-WC8-014, LAST)
- [ ] git_convention 무관(AC-WC8-015) + 신규 validator 0(AC-WC8-016)
- [ ] 006 sentinel 무수정(AC-WC8-017) + offline 무수정(AC-WC8-018) + theme-only round-trip 무수정(AC-WC8-009)
- [ ] `templ generate` drift-free(AC-WC8-013) + `make build` embedded.go 반영(AC-WC8-021)
- [ ] cross-platform build(AC-WC8-020) + full suite GREEN + 커버리지 baseline(AC-WC8-022)
- [ ] spec.md §F Exclusions의 어느 항목도 침범하지 않음(특히 009 git_convention / renderFullV3 / canonical Mode)
