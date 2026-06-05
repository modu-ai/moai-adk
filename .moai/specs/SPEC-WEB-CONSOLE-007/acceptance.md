# SPEC-WEB-CONSOLE-007 — Acceptance Criteria

> 각 AC는 GREEN-verifiable(grep/test command) + REQ 매핑 + test-class. cycle_type=tdd → 신규 거동은 RED→GREEN.
> 모든 명령은 repo root(`/Users/goos/MoAI/moai-adk-go`)에서 실행.

---

## §D.1 AC Matrix

### 신규 위젯 (Class A — markup parity)

- **AC-WC7-001** (REQ-WC7-004, HARD-8) — `toggle` 위젯 markup-parity 테스트 존재 + GREEN.
  ```bash
  grep -q "func TestToggleHelperMarkupParity" internal/web/templ_helpers_test.go && \
  go test ./internal/web/ -run TestToggleHelperMarkupParity -count=1
  ```
  단언: clean 렌더에 `class="field"` + checkbox `name=`/`value="1"`, checked 상태 렌더에 `checked`, errored 렌더에 `class="field has-error"` + `field-error` span + `alert-circle`.

- **AC-WC7-002** (REQ-WC7-004, HARD-8) — `numberField` 위젯 markup-parity 테스트 존재 + GREEN.
  ```bash
  grep -q "func TestNumberFieldHelperMarkupParity" internal/web/templ_helpers_test.go && \
  go test ./internal/web/ -run TestNumberFieldHelperMarkupParity -count=1
  ```
  단언: `<input ... type="number"` + `name=` + `value=` + `min=`/`max=`/`step=`, errored 시 `aria-invalid="true"` + field-error span.

- **AC-WC7-003** (REQ-WC7-013, HARD-8) — `templ generate` drift-free(codegen 최신).
  ```bash
  templ generate && git diff --exit-code internal/web/*_templ.go
  ```

### 검증 export seam (Class B — 단위)

- **AC-WC7-004** (REQ-WC7-002, CRITICAL SCOPE CONSTRAINT) — 신규 validator 함수 0개 불변.
  ```bash
  test "$(grep -cE 'func validate(Workflow|GitStrategy|Harness|Llm)Config' internal/config/validation.go)" -eq 0
  ```

- **AC-WC7-005** (REQ-WC7-002, HARD-3) — quality nested 검증이 기존 `validateQualityConfig` 규칙(0-100)을 재사용(신규 규칙 0개). export seam 단위 테스트 GREEN.
  ```bash
  go test ./internal/config/ -run 'TestValidateQuality' -count=1
  ```
  단언: `test_coverage_target=150` → "must be between 0 and 100"(기존 메시지 재사용), `tdd_settings.min_coverage_per_commit=-5` → 동일 규칙.

- **AC-WC7-006** (REQ-WC7-008, HARD-3) — git_convention 검증이 기존 custom-required 규칙 재사용.
  ```bash
  go test ./internal/config/ -run 'TestValidateGitConvention' -count=1
  ```
  단언: `convention=custom` + `custom.pattern=""` → "pattern is required when convention is 'custom'"(기존 메시지); `auto_detection.confidence_threshold=1.5` → "must be between 0.0 and 1.0".

### 폼 round-trip + nested isolation (Class B — HTTP 계약)

- **AC-WC7-007** (REQ-WC7-005) — 단일 nested 필드 round-trip: `quality.test_coverage_target=85` 제출 → 디스크 quality.yaml에 `test_coverage_target: 85` 영속.
  ```bash
  go test ./internal/web/ -run 'TestProjectNested.*RoundTrip' -count=1
  ```

- **AC-WC7-008** (REQ-WC7-005, REQ-WC7-012, HARD-4) — **nested sibling 보존 증명**: `quality.test_coverage_target`만 변경 제출 시 `coverage_exemptions.max_exempt_percentage` + `tdd_settings.test_first_required` + `lsp_quality_gates.enabled`가 byte-identical 보존.
  ```bash
  go test ./internal/web/ -run 'TestProjectNested.*SiblingPreserved' -count=1
  ```
  단언: write 후 재로드 → 타깃 필드 변경 AND 3개 이상 비-편집 nested 필드 LoadRaw값 유지.

- **AC-WC7-009** (REQ-WC7-005, HARD-4) — git_convention nested isolation: `auto_detection.confidence_threshold`만 변경 시 `formatting.verbose` + `validation.max_length` 보존.
  ```bash
  go test ./internal/web/ -run 'TestProjectNested.*GitConventionSiblingPreserved' -count=1
  ```

### EC-1 / EC-2 (Class B)

- **AC-WC7-010** (REQ-WC7-006, HARD-5, EC-1) — empty 제출=보존: 빈 `quality.test_coverage_target` 제출 → 기존 영속값 무변경.
  ```bash
  go test ./internal/web/ -run 'TestProjectNested.*EmptyPreserves' -count=1
  ```

- **AC-WC7-011a** (REQ-WC7-006, HARD-5, EC-1 bool) — bool companion 보존: enforce_quality toggle 미제출(companion 없음) → 기존 bool값 보존.
  ```bash
  go test ./internal/web/ -run 'TestProjectNested.*ToggleEC1' -count=1
  ```

- **AC-WC7-011b** (REQ-WC7-005, EC bool 변경) — companion 존재 + checkbox 미체크 → false로 변경 영속.
  ```bash
  go test ./internal/web/ -run 'TestProjectNested.*ToggleUnchecked' -count=1
  ```

- **AC-WC7-012** (REQ-WC7-007, HARD-5, EC-2) — atomic reject: 한 nested 필드 유효 + 다른 필드 무효 동시 제출 → **어떤 섹션도 write 안 됨**(전부 무변경) + 폼 재렌더 + per-field 에러.
  ```bash
  go test ./internal/web/ -run 'TestProjectNested.*AtomicReject' -count=1
  ```

### Server-canonical 검증 (Class B)

- **AC-WC7-013** (REQ-WC7-014, HARD-3) — out-of-range server reject: `quality.test_coverage_target=150` POST → 400 + FieldErrors["quality.test_coverage_target"] 비어있지 않음 + write 0.
  ```bash
  go test ./internal/web/ -run 'TestProjectNested.*OutOfRangeReject' -count=1
  ```

- **AC-WC7-014** (REQ-WC7-008, HARD-3) — custom-required server reject: `git_convention.convention=custom` + 빈 pattern POST → 400 + FieldErrors["git_convention.custom.pattern"] 존재 + write 0.
  ```bash
  go test ./internal/web/ -run 'TestProjectNested.*CustomPatternRequired' -count=1
  ```

- **AC-WC7-015** (REQ-WC7-003, HARD-3) — web 레이어 직접 YAML write 금지(grep cleanliness).
  ```bash
  test "$(grep -cE 'yaml\.Marshal|os\.WriteFile' internal/web/projectconfig.go internal/web/handlers.go)" -eq 0
  ```

### HARD-1 / HARD-2 boundary 무수정 (Class B — 무수정 sentinel)

- **AC-WC7-016** (REQ-WC7-011, HARD-2) — 006 scope-boundary sentinel **무수정 GREEN**: workflow/harness/git-strategy가 `DO_NOT_TOUCH` byte-identical.
  ```bash
  go test ./internal/web/ -run 'TestScopeBoundary|TestPersistence.*Scope|TestSave.*Boundary' -count=1
  ```
  (integration_test.go:197-205 sentinel 로직 무수정 — `git diff --exit-code` 해당 라인 범위 변경 0.)

- **AC-WC7-017** (REQ-WC7-011, HARD-1) — handleSave가 quality/git_convention 외 섹션 미write(코드 경계 불변).
  ```bash
  git diff origin/main -- internal/web/integration_test.go | grep -E '^\+' | grep -cE 'workflow|harness|git-strategy' | grep -qx 0
  ```
  (sentinel 단언부 무수정 — 추가 라인에 workflow/harness/git-strategy 변경 0.)

### Offline / HTMX 계약 (Class B + Class A)

- **AC-WC7-018** (REQ-WC7-013, HARD-6) — offline zero-network: 신규 CDN/외부 asset 0.
  ```bash
  test "$(grep -rcE 'unpkg\.com|jsdelivr|cdn\.|googleapis\.com|fonts\.gstatic' internal/web/*.templ internal/web/assets/*.css internal/web/assets/*.js | awk -F: '{s+=$2} END{print s}')" -eq 0
  ```

- **AC-WC7-019** (REQ-WC7-009) — hx-boost full-page swap 유지(partial-swap fragment 미도입): `/save` 응답이 full HTML page(`<!DOCTYPE html>`).
  ```bash
  go test ./internal/web/ -run 'TestSave.*FullPage|TestHtmx.*FullPage' -count=1 && \
  test "$(grep -cE 'hx-target|hx-swap=' internal/web/root.templ internal/web/fieldsets.templ internal/web/page.templ)" -eq 0
  ```

### 전체 회귀 (Class B)

- **AC-WC7-020** (전 REQ) — full suite GREEN + 커버리지 baseline 유지.
  ```bash
  go test ./internal/web/... ./internal/config/... ./pkg/models/... -count=1 && \
  go test ./internal/web/... -cover -count=1
  ```
  단언: 0 FAIL; internal/web 커버리지 ≥ 006 baseline(90.9%) 이내(±하락 없음).

---

## §D.2 Severity
- **MUST-PASS (closure gate)**: AC-WC7-003, -004, -005, -006, -007, -008, -012, -013, -014, -015, -016, -017, -020. (위젯 codegen, 신규 validator 0개, nested isolation, atomic reject, server reject, boundary 무수정, full suite.)
- **SHOULD-PASS**: AC-WC7-001, -002, -009, -010, -011a, -011b, -018, -019.

## §D.3 REQ ↔ AC Traceability
| REQ | AC |
|-----|-----|
| REQ-WC7-001 | -017, -020 |
| REQ-WC7-002 | -004, -005 |
| REQ-WC7-003 | -005, -015 |
| REQ-WC7-004 | -001, -002 |
| REQ-WC7-005 | -007, -008, -009, -011b |
| REQ-WC7-006 | -010, -011a |
| REQ-WC7-007 | -012 |
| REQ-WC7-008 | -006, -014 |
| REQ-WC7-009 | -019 |
| REQ-WC7-010 | -007 (GET 현재값 경로) |
| REQ-WC7-011 | -016, -017 |
| REQ-WC7-012 | -008 |
| REQ-WC7-013 | -003, -018 |
| REQ-WC7-014 | -013 |

## §D.4 Definition of Done
- [ ] MUST-PASS AC 전부 GREEN
- [ ] `templ generate` drift-free
- [ ] 신규 validator 함수 0개 (AC-WC7-004)
- [ ] 006 sentinel 무수정 (AC-WC7-016, -017)
- [ ] nested isolation 증명 (AC-WC7-008, -009)
- [ ] full suite GREEN + 커버리지 baseline (AC-WC7-020)
- [ ] spec.md §F Exclusions의 어느 항목도 침범하지 않음
