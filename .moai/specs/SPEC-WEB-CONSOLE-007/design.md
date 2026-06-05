# SPEC-WEB-CONSOLE-007 — Design

> WHAT/WHY는 spec.md. 본 파일은 nested 직렬화 + curated 인벤토리 + per-field validation map + 신규 위젯 + write-seam load-modify-write + nested isolation 메커니즘의 설계 디테일.
> 모든 ground-truth 인용은 research.md.

---

## §A. Nested form field naming (dot-path serialization)

### A.1 폼 필드 명명 규약
신규 nested 필드는 **dot-path `name=`**를 사용한다(기존 스칼라 `development_mode`/`git_convention`와 충돌 없음):

| 폼 필드 name= | 대상 struct path |
|---------------|------------------|
| `quality.test_coverage_target` | `Config.Quality.TestCoverageTarget` |
| `quality.enforce_quality` | `Config.Quality.EnforceQuality` |
| `quality.tdd_settings.min_coverage_per_commit` | `Config.Quality.TDDSettings.MinCoveragePerCommit` |
| `git_convention.auto_detection.confidence_threshold` | `Config.GitConvention.AutoDetection.ConfidenceThreshold` |
| `git_convention.auto_detection.enabled` | `Config.GitConvention.AutoDetection.Enabled` |
| `git_convention.custom.pattern` | `Config.GitConvention.Custom.Pattern` |

기존 `development_mode` / `git_convention`(convention) 스칼라 name=는 **무변경 유지**(006/003 계약).

### A.2 서버 파싱
`handleSave`는 `r.PostFormValue("quality.test_coverage_target")` 등으로 각 dot-path를 직접 읽는다. 동적 reflection path-walker는 **만들지 않는다**(over-engineering 회피 — 필드 수가 6개로 한정). 각 필드는 명시적 `PostFormValue` + 타입 변환(`strconv.Atoi`/`ParseFloat`/checkbox presence)으로 바인딩. 빈 문자열은 "미제출/보존"(EC-1).

새 타입 `projectNestedForm` 구조체(internal/web)가 파싱된 raw string + 변환값 + per-field 파싱 에러를 운반한다. 이는 ProfilePreferences가 아니며(HARD-7 project-scope), pageView에 echo-back용 현재값으로도 사용된다.

---

## §B. Per-field validation map

| 필드 | 변환 | 검증 위치 (확장) | 규칙 |
|------|------|------------------|------|
| Q1 `test_coverage_target` | Atoi | `validateQualityConfig` (validation.go:99-106, 기존) | 0-100; 폼 배선만 |
| Q2 `enforce_quality` | checkbox presence | (검증 무관) | bool — `validateQualityConfig`에 no-op 명시 주석 |
| Q3 `tdd_settings.min_coverage_per_commit` | Atoi | `validateQualityConfig` (validation.go:108-115, 기존) | 0-100; 폼 배선만 |
| G1 `auto_detection.confidence_threshold` | ParseFloat | `validateGitConventionConfig` (validation.go:184-191, 기존) | [0.0,1.0]; 폼 배선만 |
| G2 `auto_detection.enabled` | checkbox presence | (검증 무관) | bool — no-op |
| G3 `custom.pattern` | string | `validateGitConventionConfig` (validation.go:202-209, 기존) | `convention=="custom"`일 때 required; 폼 배선만 |

### B.1 검증 배선 방식 (CRITICAL SCOPE CONSTRAINT)
- 두 기존 검증기(`validateQualityConfig` / `validateGitConventionConfig`)는 **이미 위 int/float 규칙을 보유**한다. 본 SPEC은 새 규칙을 추가하지 않고, web 레이어가 폼 → struct 변환 후 **이 두 검증기를 호출**하도록 배선한다.
- web `validateProjectConfig`(validate.go:162-173)를 확장: 현재 2-인자(devMode, convention) → curated nested 값을 받는 형태로 확장하되, **내부 검증은 config 패키지의 두 기존 검증기를 재사용**한다. 신규 validator 함수 0개(HARD-3).
- 파싱 에러(예: `test_coverage_target=abc`)는 web 레이어에서 FieldErrors에 매핑("must be an integer") — 이는 검증 규칙이 아니라 타입 변환 가드(EC-4 류, 기존 exact-match 패턴 연장).

### B.2 검증 호출 흐름
```
handleSave
  → bindForm(profile) → validatePrefs                       (기존, 무변경)
  → parseProjectNestedForm(r) → projectNestedForm           (신규 파싱)
  → validateProjectConfigExtended(form)                     (validate.go 확장)
        ├─ models.DevelopmentMode(devMode).IsValid()        (기존 재사용)
        ├─ config.IsValidConvention(convention)             (기존 재사용)
        ├─ build QualityConfig from LoadRaw + form deltas
        │    → config.validateQualityConfig(quality)        (기존 재사용, export 필요 시 thin wrapper)
        └─ build GitConventionConfig from LoadRaw + deltas
             → config.validateGitConventionConfig(gc)       (기존 재사용)
  → merge FieldErrors; len>0 → atomic reject (EC-2)         (handlers.go:190 패턴)
```

> 주의: `validateQualityConfig`/`validateGitConventionConfig`는 현재 unexported(internal/config). 재사용을 위해 (a) thin exported wrapper(`config.ValidateQualitySection`/`config.ValidateGitConventionSection`)를 추가하거나 (b) 기존 export된 `IsValidConvention`처럼 필드별 export 예측자를 추가한다. **신규 validator 함수가 아니라 기존 검증기 위의 export seam** — CRITICAL SCOPE CONSTRAINT 준수(새 규칙 0개). 선택은 run-phase RED에서 확정(가장 작은 변경 우선).

---

## §C. 신규 위젯 설계 (Templ)

### C.1 `toggle` (boolean)
```
type toggleArgs struct { Name, Title, Key, Desc string; Checked bool; Errors map[string]string }
templ toggle(a toggleArgs) {
  <div class={ "field", templ.KV("has-error", hasFieldError(a.Errors, a.Name)) }>
    <span class="field__label"><span class="field__title" data-i18n={ "f."+a.Name+".title" }>{ a.Title }</span><code class="field__key">{ a.Key }</code></span>
    <span class="field__desc" data-i18n={ "f."+a.Name+".desc" }>{ a.Desc }</span>
    // checkbox: segmentCheckbox 패턴 재사용 (name= value="1" checked)
    if a.Checked { <input type="checkbox" id={ a.Name } name={ a.Name } value="1" checked/> }
    else { <input type="checkbox" id={ a.Name } name={ a.Name } value="1"/> }
    if msg := fieldErrorMsg(a.Errors, a.Name); msg != "" { <span class="field-error">@icon("alert-circle")<span>{ msg }</span></span> }
  </div>
}
```
chrome는 `optSelect`(page.templ:90), checkbox는 `segmentCheckbox`(fieldsets.templ:198) 패턴. checkbox presence = true(value="1"), 미존재 = false → EC-1: bool은 "미제출=false"가 아니라 "미제출=보존"이 되도록 write-seam에서 처리(§D 참조 — bool은 hidden companion 또는 명시적 "변경됨" 마커로 보존 의미 유지).

### C.2 `numberField` (int/float)
```
type numberFieldArgs struct { Name, Title, Key, Desc, Value string; Min, Max, Step string; Errors map[string]string }
templ numberField(a numberFieldArgs) {
  <div class={ "field", templ.KV("has-error", hasFieldError(a.Errors, a.Name)) }>
    <span class="field__label">...</span>
    <span class="field__desc" ...>{ a.Desc }</span>
    if hasFieldError(a.Errors, a.Name) {
      <input class="input" type="number" id={ a.Name } name={ a.Name } value={ a.Value } min={ a.Min } max={ a.Max } step={ a.Step } aria-invalid="true" aria-describedby={ "err_"+a.Name }/>
      <span class="field-error" id={ "err_"+a.Name }>@icon("alert-circle")<span>{ fieldErrorMsg(a.Errors, a.Name) }</span></span>
    } else {
      <input class="input" type="number" id={ a.Name } name={ a.Name } value={ a.Value } min={ a.Min } max={ a.Max } step={ a.Step }/>
    }
  </div>
}
```
chrome는 fieldsetIdentity의 text input 패턴(fieldsets.templ:21-30) + `optSelect` field chrome. `class="input"` 재사용. `min/max/step`는 클라이언트 힌트일 뿐 — **server-canonical 검증이 권위**(HARD-3): 브라우저 number input 우회 제출도 server가 reject.

### C.3 위젯 codegen
`.templ` 추가 → `templ generate` → `page_templ.go`/`fieldsets_templ.go` 재생성(zero Node, HARD-8). 위치: `toggle`/`numberField`는 page.templ(공유 위젯 영역).

---

## §D. Write-seam load-modify-write + nested isolation 메커니즘 (HARD-4 crux)

### D.1 메커니즘
`writeProjectConfig`(projectconfig.go:37-70)를 확장한다. 핵심 불변식: **SetSection은 섹션 구조체 전체를 교체**하고 Save는 전체를 직렬화하므로(manager.go:138-196), **LoadRaw가 반환한 전체 구조체에서 타깃 nested 필드만 변경**하면 sibling nested 필드는 LoadRaw값 그대로 보존된다.

```go
func writeProjectConfig(projectRoot string, form projectNestedForm) error {
    mgr := config.NewConfigManager()
    cfg, err := mgr.LoadRaw(projectRoot)   // 전체 트리 로드 (모든 nested 필드 포함)
    if err != nil { return ... }
    changed := false

    // --- quality: 전체 구조체 복사 → 타깃 nested 필드만 변경 ---
    if form.touchesQuality() {
        q := cfg.Quality                    // ★ 전체 구조체 복사 (DDD/TDD/Coverage/LSP/... 전부 포함)
        if form.devModeSet { q.DevelopmentMode = models.DevelopmentMode(form.devMode) }
        if form.coverageTargetSet { q.TestCoverageTarget = form.coverageTarget }
        if form.enforceQualitySet { q.EnforceQuality = form.enforceQuality }
        if form.minCovSet { q.TDDSettings.MinCoveragePerCommit = form.minCov }  // nested-of-nested: TDDSettings 전체 보존 후 한 필드만
        mgr.SetSection("quality", q)        // 변경 안 된 nested 필드는 q에 LoadRaw값 그대로 → 보존
        changed = true
    }

    // --- git_convention: 동일 패턴 ---
    if form.touchesGitConvention() {
        gc := cfg.GitConvention
        if form.conventionSet { gc.Convention = form.convention }
        if form.confidenceSet { gc.AutoDetection.ConfidenceThreshold = form.confidence }  // AutoDetection 전체 보존 후 한 필드
        if form.autoEnabledSet { gc.AutoDetection.Enabled = form.autoEnabled }
        if form.customPatternSet { gc.Custom.Pattern = form.customPattern }
        mgr.SetSection("git_convention", gc)
        changed = true
    }

    if changed { return mgr.Save() }
    return nil
}
```

### D.2 EC-1 (empty=보존) for nested 필드
- **string/enum/int/float**: `*Set` 플래그(form 파싱 시 빈 제출이면 false)로 "미제출"을 구분 → 미제출 필드는 LoadRaw값을 그대로 둠.
- **bool (checkbox) 특수성**: HTML checkbox는 미체크 시 폼에 `name`을 전혀 보내지 않으므로 "미제출"과 "false로 변경"을 구분 불가. EC-1 "empty=보존" 의미를 bool에 일관 적용하기 위해 **hidden companion 필드** 패턴 사용: `<input type="hidden" name="quality.enforce_quality__present" value="1">`를 toggle 옆에 렌더 → companion이 있고 checkbox가 없으면 "false로 변경", companion 자체가 없으면 "보존". 이로써 bool도 EC-1 의미 보존. (run-phase RED에서 companion 패턴 vs "bool은 항상 변경"의 단순화 중 선택; 기본은 companion으로 EC-1 일관성 유지.)

### D.3 nested isolation 증명 테스트 (HARD-4 must-pass)
- 픽스처: `coverage_exemptions.max_exempt_percentage=42` + `tdd_settings.test_first_required=true` + `lsp_quality_gates.enabled=true` 등 **편집 대상이 아닌** nested 필드를 채운 quality.yaml.
- 폼: `quality.test_coverage_target=85`만 제출.
- 단언: write 후 quality.yaml 재로드 → `test_coverage_target==85` AND `coverage_exemptions.max_exempt_percentage==42`(보존) AND `tdd_settings.test_first_required==true`(보존) AND `lsp_quality_gates.enabled==true`(보존). git_convention도 동형 픽스처(`formatting.verbose=true` + `validation.max_length=72` 보존).

---

## §E. View-model 확장 (pageView)

`pageView`(handlers.go:16-55)에 신규 현재값 + 옵션 힌트 추가:
```
// project nested (SPEC-WEB-CONSOLE-007)
CurTestCoverageTarget string  // echo-back 문자열 (int → strconv)
CurEnforceQuality     bool
CurMinCoveragePerCommit string
CurConfidenceThreshold  string  // float → strconv
CurAutoDetectionEnabled bool
CurCustomPattern        string
```
GET 시 readProjectConfig 확장으로 디스크 현재값을, rejected POST 시 제출값 echo-back(FieldErrors와 함께). 옵션 리스트 불요(number/bool/text는 enum 아님).

---

## §F. fieldsetProject 확장
`fieldsetProject`(fieldsets.templ:209-226)에 신규 필드 추가(섹션 격리 — Project fieldset 내부에만, 다른 fieldset 무변경). `count.project` data-i18n 카운트 "2 fields" → 신규 카운트로 갱신. quality 필드와 git_convention 필드를 시각적으로 sub-group(grid)로 배치. root.templ:105-109 합성 라인은 무변경(fieldsetProject 내부만 확장).

---

## §G. Anti-over-engineering 가드
- 동적 path-walker/reflection 금지 — 6개 필드 명시적 PostFormValue.
- 신규 validator 함수 금지 — 기존 두 검증기 export seam만.
- 동적 섹션 레지스트리 금지 — root.templ static 합성 유지.
- partial-swap fragment 금지 — hx-boost full-page 유지.
- bool companion은 EC-1 일관성을 위해서만; 단순화 가능하면 run-phase에서 재평가.
