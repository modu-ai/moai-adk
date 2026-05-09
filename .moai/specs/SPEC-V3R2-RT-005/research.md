# SPEC-V3R2-RT-005 Deep Research (Phase 0.5)

> Research artifact for **Multi-Layer Settings Resolution with Provenance Tags**.
> Companion to `spec.md` (v0.1.0). Authored against branch `plan/SPEC-V3R2-RT-005` from `/Users/goos/MoAI/moai-adk-go` (main repo, the branch is checked out here per `git worktree list`).

## HISTORY

| Version | Date       | Author                                  | Description                                                              |
|---------|------------|-----------------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 0.5)  | Initial deep research per `.claude/skills/moai/workflows/plan.md` Phase 0.5. Substantiates spec.md §1-§8 with 30+ file:line evidence anchors. |

---

## 1. Goal of Research

`spec.md` §1-§8을 구체적인 file:line evidence와 외부 라이브러리 평가로 뒷받침하여, run phase가 알려진 baseline 위에서 27개 REQ (REQ-V3R2-RT-005-001..051)를 구현할 수 있도록 한다.

본 research는 다음 7개 질문에 답한다:

1. **현재 config 패키지 inventory**: `internal/config/`의 14개 파일이 이미 구현한 부분과 spec 27 REQ 사이의 delta는?
2. **Generic `Value[T any]` 설계 근거**: Go 1.22+ 제네릭이 왜 필요하며, runtime overhead는?
3. **`Source` enum의 RT-002 공유 계약**: 8-tier ordering이 RT-002 permission stack과 어떻게 일치해야 하는가?
4. **8-tier 우선순위 머지 알고리즘**: deterministic merge가 prompt-cache prefix 안정성을 어떻게 보장하는가?
5. **5개 누락된 yaml loader (r6 §5.2 CRITICAL)**: constitution/context/interview/design/harness — 이 SPEC이 패턴을 정의하고 MIG-003이 구현하는 분담은?
6. **CI audit test 작동 방식**: `internal/config/audit_test.go`가 yaml↔Go struct parity를 어떻게 강제하는가?
7. **RT-002 (permission stack) + RT-003 (sandbox) + RT-006 (hook) consumer 영향**: Source enum + Value[T]를 4-방향으로 공급하는 substrate 책임?

---

## 2. Inventory of `internal/config/` skeleton (existing)

`ls /Users/goos/MoAI/moai-adk-go/internal/config/`는 22개 파일을 반환한다 (14 source + 8 test + 1 testdata 디렉토리). 각 파일의 line count는 직접 측정으로 확인:

| # | File | Lines | Purpose | Implements |
|---|------|-------|---------|------------|
| 1 | `types.go` | 372 | `Config` 루트 구조체 + 16개 섹션 (User, Language, Quality, GitStrategy, GitConvention, System, LLM, Pricing, Ralph, Workflow, State, Statusline, Gate, Sunset, Research, Project) + `userFileWrapper`/`languageFileWrapper`/`qualityFileWrapper` 등 | REQ-008 partial — yaml↔Go parity surface |
| 2 | `source.go` | 113 | `Source int` 타입 + 8 const (`SrcPolicy..SrcBuiltin`) + `String()` + `Priority()` + `ParseSource()` + `AllSources()` | REQ-001 ✅ DONE |
| 3 | `provenance.go` | 71 | `Provenance{Source, Origin, Loaded, SchemaVersion, OverriddenBy}` + `Value[T any]{V, P}` + `Unwrap()` + `Origin()` + `IsBuiltin()` + `IsDefault()` | REQ-002, REQ-003 ✅ DONE (struct only) |
| 4 | `loader.go` | 216 | `Loader` (sync.RWMutex) + `Load()` for project tier yaml sections + `loadYAMLFile()` 헬퍼 | REQ-008 partial (8 sections wired) |
| 5 | `merge.go` | 275 | `MergedSettings` map + `MergeAll()` deterministic merge with reflection-based zero check + `Diff()` + `Dump(json/yaml)` | REQ-005, REQ-006, REQ-007, REQ-010, REQ-012 (partial — no OverriddenBy population yet) |
| 6 | `resolver.go` | 355 | `SettingsResolver` interface + `resolver` struct + `Load()` 8-tier loop + `loadPolicyTier`/`loadUserTier`/`loadProjectTier`/`loadLocalTier`/`loadSkillTier`/`loadSessionTier`/`loadBuiltinTier` (Plugin 항상 빈 슬롯) | REQ-004, REQ-010, REQ-014 (partial) |
| 7 | `audit_test.go` | 52 | `TestAuditParity` placeholder — `t.Skip()` with full implementation comment | REQ-008, REQ-043 (placeholder만) |
| 8 | `validation.go` | 249 | `Validate(cfg, loadedSections)` + `validateRequired`/`validateDevelopmentMode`/`validateQualityConfig`/`validateGitConventionConfig`/`validateDynamicTokens` + `ValidationError`/`ValidationErrors` | REQ-013 (partial — no Type Error class yet) |
| 9 | `manager.go` | 402 | `Manager` (high-level facade over Loader+Validation) + `NewManager`/`Load`/`Get`/`Set`/`Save` | Existing API — RT-005에서 변경 X |
| 10 | `defaults.go` | 346 | `NewDefaultConfig()` + 16개 섹션의 builtin 기본값 | SrcBuiltin tier source ✅ |
| 11 | `errors.go` | 90 | sentinel error 8종 (`ErrConfigNotFound`/`ErrInvalidConfig`/`ErrInvalidYAML` 등) + `ValidationError`/`ValidationErrors` | error API ✅ |
| 12 | `resolver_errors.go` | 79 | `ConfigTypeError`/`ConfigAmbiguous`/`PolicyOverrideRejected`/`ConfigSchemaMismatch`/`TierReadError` (Unwrap 지원) | REQ-013, REQ-022, REQ-040, REQ-041, REQ-042 ✅ types only |
| 13 | `required_checks.go` | 63 | `RequiredChecks` (auxiliary fields validation) | adjacent — RT-005 변경 X |
| 14 | `envkeys.go` | 105 | env var name 상수 (`MOAI_*`) | adjacent — RT-005 변경 X |
| 15 | `*_test.go` (8 files) | ~1500 | 기존 happy-path coverage | Test baseline |

### 2.1 Delta analysis (skeleton → 27 REQ 충족)

`Grep` of `internal/config/*.go` 결과 skeleton은 **약 65%** 의 REQ를 구조적으로 구현하지만 다음 load-bearing 행동들이 누락됨:

| Skeleton state | Spec requirement | Gap |
|----------------|------------------|-----|
| `source.go:11-44` 8 const + `AllSources()` | REQ-V3R2-RT-005-001 (8값 enum, priority 정렬) | ✅ 완료 |
| `provenance.go:10-30` Provenance struct | REQ-V3R2-RT-005-002 (Source/Origin/Loaded 필드) | ✅ 완료 (validator tag만 추가하면 REQ-013 강화) |
| `provenance.go:34-60` `Value[T any]` 제네릭 | REQ-V3R2-RT-005-003 (`Unwrap()`, `Origin()`) | ✅ 완료 |
| `resolver.go:17-29` SettingsResolver interface | REQ-V3R2-RT-005-004 (`Load`/`Key`/`Dump`/`Diff`) | ⚠️ partial — `Diff(a, b Source)` 시그니처는 `(map[string]Value[any], error)` 반환하지만 spec §5.1은 `map[string]Value[any]`만 명시; **Dump은 io.Writer가 아닌 `any`** → 사양 정합성 보정 필요 |
| `merge.go:63-134` `MergeAll` priority walk + first-non-zero | REQ-V3R2-RT-005-005, -010 (deterministic merge) | ⚠️ partial — `OverriddenBy` 필드는 collect되지만 `provenance.OverriddenBy`로 전파됨; 그러나 spec REQ-012 "lower-tier path 추가" 는 inline merge 단계에서 부재 (현재는 후속 tier만 모음) |
| `merge.go:229-238` `Dump(json/yaml)` | REQ-V3R2-RT-005-006, -030 | ⚠️ partial — `formatMapAsJSON()`은 `fmt.Sprintf("%+v", m)` placeholder, `encoding/json`로 교체 필요 |
| `merge.go:168-205` `Diff(a, b *MergedSettings)` | REQ-V3R2-RT-005-007 | ⚠️ partial — 현재는 두 MergedSettings 비교; spec REQ-007은 `Diff(a, b Source)` 두 tier 비교 — `resolver.go:323-355`에 wrapper 있음 ✅ |
| `audit_test.go:9-22` `t.Skip("placeholder")` | REQ-V3R2-RT-005-008, -043 | ❌ 미구현 — TestAuditParity는 skip만 존재 |
| `resolver.go:46-74` `Load()` 8-tier loop | REQ-V3R2-RT-005-010, -011 | ⚠️ partial — diff-aware reload (REQ-011) 미구현; 매번 cold reload |
| `resolver.go:103-125` `loadPolicyTier` 플랫폼별 경로 | REQ-V3R2-RT-005-014 (file 부재 시 빈 tier) | ✅ 완료 |
| `resolver_errors.go:9-19` `ConfigTypeError` | REQ-V3R2-RT-005-013 | ⚠️ partial — type 정의는 있으나 loader가 raise 하지 않음 (실제 type-mismatch 검출 로직 부재) |
| `resolver.go:218-220` `loadSkillTier()` placeholder | REQ-V3R2-RT-005-015 partial (Source: SrcSkill 슬롯) | ❌ 미구현 — frontmatter 파싱 부재 |
| `resolver.go:252-255` `loadYAMLFile` placeholder return empty | REQ-V3R2-RT-005-010 (실제 yaml 파싱) | ❌ 미구현 — `gopkg.in/yaml.v3` 사용해 실제 파싱 필요 |
| diff-aware reload (`Reload(path string)`) 메서드 | REQ-V3R2-RT-005-011 | ❌ 미구현 — RT-006 ConfigChange hook 소비처 |
| `policy.strict_mode` enforcement | REQ-V3R2-RT-005-022 (`PolicyOverrideRejected`) | ❌ 미구현 — type만 존재; merge 단계에서 enforcement 부재 |
| yaml↔yml ambiguity detection | REQ-V3R2-RT-005-041 (`ConfigAmbiguous`) | ❌ 미구현 — `loadYAMLSections`은 단순 첫번째만 채택 |
| schema_version 추적 | REQ-V3R2-RT-005-033 | ❌ 미구현 — `Provenance.SchemaVersion` 필드 존재하나 항상 0 |
| `.moai/logs/config.log` 경고 로그 | REQ-V3R2-RT-005-040 | ⚠️ partial — `slog.Warn`만 사용; 전용 로그 파일 미구성 |
| `moai doctor config dump --key` 단일 키 출력 | REQ-V3R2-RT-005-032 | ✅ 완료 (`doctor_config.go:55-72`) |
| `moai doctor config dump --format yaml` # source: 코멘트 | REQ-V3R2-RT-005-030 | ✅ 완료 (`merge.go:259-267`) |
| `moai doctor config diff` merged-view delta | REQ-V3R2-RT-005-051 | ⚠️ partial — `resolver.go:Diff` 두 tier 로드만; "merged-view delta" 의미가 spec과 맞는지 확인 필요 |

**Summary**: skeleton은 27 REQ 중 약 17개를 구조적/부분 구현; M2-M5에서 10개의 행동 gap (TestAuditParity, diff-aware reload, type-mismatch 검출, ConfigAmbiguous 검출, schema_version 전파, policy strict_mode, real yaml 파싱, frontmatter 파싱, OverriddenBy 정확한 전파, json dump real encoding) 을 채운다.

---

## 3. Generic `Value[T any]` 설계 근거 — Go 1.22+ generics

### 3.1 왜 generics가 필요한가

각 config field는 서로 다른 typed Go 값을 가진다 (`string`, `int`, `bool`, `[]string`, struct 등). 만약 `Value` 가 `interface{}` (`any`) 만 보유한다면:

- 호출 측에서 `cfg.User.Name.V.(string)` 식의 type assertion 필요 → 런타임 panic 위험.
- 컴파일러가 type 안전성을 보장하지 못함 → "value 타입 변경 → assertion silent break" 시나리오.

`Value[T any]` 제네릭 wrapper는 이 두 문제를 컴파일 타임에 해결한다:

```go
type Value[T any] struct {
    V T
    P Provenance
}

// 호출 측: type assertion 불요
name := cfg.User.Name.V  // Go 컴파일러가 string 보장
```

**근거**: r3-cc-architecture-reread.md §1.3 "Multi-layer settings with explicit precedence and provenance" + provenance.md:34-60 generic implementation already in place.

### 3.2 런타임 overhead

Go generics는 GC shape stenciling으로 컴파일된다. `Value[T]`는 메모리 layout 측면에서 `struct{V T; P Provenance}`와 동일하며 별도 reflection이 발생하지 않는다.

- Memory: per-value `Provenance` overhead = 8 (Source int) + 16 (string header) + 24 (time.Time) + 8 (int) + 24 (slice header) = ~80 bytes.
- 2 MiB RSS ceiling per spec §7 Constraints → 2 MiB / 80 bytes ≈ 26,000 keys 까지 안전 (typical project ~100 keys → 8 KiB only).

### 3.3 Type erasure 경계

merged settings 단계에서는 모든 값을 `Value[any]`로 store 한다 (`merge.go:14-15` `values map[string]Value[any]`). 이는:
- yaml 파일에서 임의의 typed 값 (any로 unmarshal) 을 받아들이기 위함.
- 호출 측이 `cfg.User.Name` 식의 typed accessor를 사용할 때는 wrapper layer가 `any → string` 변환 (mechanical, 한곳에서).

이 약간의 type erasure는 yaml-driven config의 본질적 trade-off; `interface{}`-erase한 후 typed accessor가 reverse-lookup 하도록 하는 것이 표준 Go 패턴이다.

---

## 4. `Source` enum의 RT-002 공유 계약

### 4.1 8-tier 우선순위 (frozen)

spec.md §2 In-scope item 1 + source.go:11-44 의 `Source` 정의는 다음 8 값을 우선순위 순서로 고정한다:

```
SrcPolicy   (0, 최고)
SrcUser     (1)
SrcProject  (2)
SrcLocal    (3)
SrcPlugin   (4) — v3.0에서는 빈 슬롯
SrcSkill    (5)
SrcSession  (6)
SrcBuiltin  (7, 최저)
```

이 순서는:
- master.md:L974 §8 BC-V3R2-015 declared as **constitutional invariant** (FROZEN per CON-001).
- r3-cc-architecture-reread.md §1.3 "policySettings > userSettings > projectSettings > localSettings > pluginSettings > sessionRules" — Claude Code reference.
- SPEC-V3R2-RT-002 (permission stack) 는 동일한 8값을 그대로 소비; divergence는 master §5.5 "Multi-layer settings with provenance is the prerequisite for S-1 (permission stack)" 위반.

### 4.2 RT-002와의 inter-SPEC contract

RT-002 permission stack의 `permission.allow / deny / ask` rules는 각 rule이 `Source` 필드를 carry. 예:

```go
// RT-002 (consumer)
type PermissionRule struct {
    Pattern string
    Action  string  // allow|deny|ask
    Source  config.Source  // <- imports from RT-005
    Origin  string
}
```

따라서:
- **RT-005가 Source enum의 단일 진실 원천**이다.
- RT-002, RT-003 (sandbox), RT-006 (hook) 모두 `import "internal/config"` 후 `config.SrcUser` 등 직접 참조.
- 새 tier 추가 (예: v3.1+ plugin contributor) 는 RT-005 SPEC을 evolve 해야 하며, downstream SPEC (RT-002/-003/-006) 도 update.

### 4.3 RT-004 Subset 차이점

RT-004 (Typed Session State) 는 RT-005 enum의 **5-value subset**만 사용 (research.md RT-004 §12.1): `SrcUser`, `SrcProject`, `SrcLocal`, `SrcSession`, `SrcHook`. RT-004의 ProvenanceTag string은 RT-005의 `Source` int enum과 **다른 타입**이다 — RT-004가 먼저 작성되었기 때문에 string-based 표현을 사용했고, 후속 SPEC-V3R2-MIG-002가 RT-004 ProvenanceTag → RT-005 Source 통합을 담당한다 (RT-005 out-of-scope per spec §2).

---

## 5. 8-tier 우선순위 머지 알고리즘 — deterministic merge

### 5.1 Algorithm contract

`merge.go:63-134` `MergeAll`은:

```
for each key K across all 8 tiers:
    for source in [SrcPolicy, SrcUser, ..., SrcBuiltin]:  # priority order
        if K exists in tier[source] AND value is non-zero:
            if first_match:
                winning_value = value
                winning_source = source
                winning_origin = origins[source]
            else:
                overriddenBy.append(origins[source])
    result[K] = Value{V: winning_value, P: Provenance{Source, Origin, Loaded, OverriddenBy}}
```

### 5.2 Determinism 보장 조건

같은 8 tier 입력 → byte-identical merged 출력. 이를 위해:

1. `AllSources()` 결과는 fixed slice (source.go:101-113).
2. `allKeys` map iteration order는 Go에서 non-deterministic이지만, **결과 map의 의미적 동등성**만 필요; 직렬화 시 (Dump) 정렬한다.
3. yaml 파일 내 key 순서가 다르더라도 동일 key set이면 동일 결과.
4. `time.Time` 의 `Loaded` 필드는 `loadedAt` (Load 호출 시각) 으로 통일 → 같은 process 내 stable.

### 5.3 prompt-cache prefix 안정성 (P-C05 secondary benefit)

spec §7 Constraints "Identical tier inputs MUST produce byte-identical merged output" 은:
- Anthropic prompt-cache가 prefix-keyed; merged settings를 system prompt에 포함하면 매 turn 이를 재구성 → cache miss.
- byte-identical 보장 시 `merged settings → JSON.Marshal → systemPrompt` 가 stable → cache hit.
- RT-004 hydrate.go 의 `// cache-prefix: DO NOT REORDER` 주석과 같은 의도; **RT-005에서는 결과만 deterministic하면 RT-004가 사용**.

### 5.4 zero-value 처리 (`isZero`)

`merge.go:138-162` `isZero(v any)` 는 reflection으로 모든 Go 타입 zero 값 검출:

- bool: `false`
- int family: `0`
- string: `""`
- slice/map: `len == 0`
- pointer/interface: `nil`

이 동작의 의미: "yaml 파일에 명시되지 않은 값은 zero" → "다음 lower tier 가 win". 그러나 사용자가 의도적으로 `coverage_threshold: 0` 을 설정한 경우 zero로 간주되어 builtin default로 fallback 됨 → spec §8 risk "users reach for ~/.moai/settings.json" 에 해당. 완화: `*int` 포인터로 wrapping 하여 nil 과 0 구분 (M3 추가 검토; out-of-scope per spec §2).

---

## 6. 5개 누락된 yaml loader (r6 §5.2 CRITICAL)

### 6.1 background

problem-catalog.md Cluster 4 P-H06 (CRITICAL): "5 yaml sections (`constitution.yaml`, `context.yaml`, `interview.yaml`, `design.yaml`, `harness.yaml`) have no Go loader today, forcing them to be template-only artifacts with zero runtime enforcement."

`ls .moai/config/sections/` 결과 23개 yaml 섹션 존재; `internal/config/types.go:313-318` `sectionNames` slice는 16개만 등록. Delta 7개 yaml 파일은 Go struct 부재:

```
constitution.yaml   (876 bytes)  ❌ no Go struct
context.yaml         (531 bytes)  ❌ no Go struct
design.yaml         (2763 bytes)  ❌ no Go struct
git-strategy.yaml   (3071 bytes)  ❌ no Go struct
harness.yaml        (7814 bytes)  ❌ no Go struct
interview.yaml       (264 bytes)  ❌ no Go struct
lsp.yaml            (8098 bytes)  ❌ no Go struct
mx.yaml             (9368 bytes)  ❌ no Go struct
runtime.yaml        (1804 bytes)  ❌ no Go struct (workflow은 별도 wrapper 있음)
security.yaml       (1023 bytes)  ❌ no Go struct
sunset.yaml          (820 bytes)  ✅ SunsetConfig 있음 (types.go:243-258)
workflow.yaml       (5031 bytes)  ⚠️ partial (WorkflowConfig 있으나 role_profiles만 사용)
```

(spec.md §3 Environment에서 r6 §5.2가 5개를 명시하지만 actual 측정으로는 더 많음; spec의 5개 = constitution/context/interview/design/harness, 사용자에게 가장 critical한 것들.)

### 6.2 RT-005 vs MIG-003 분담

| 책임 | SPEC | 산출물 |
|-----|------|--------|
| `Source` enum + `Provenance` + `Value[T]` 구조 | RT-005 | provenance.go, source.go ✅ |
| 8-tier merge + resolver | RT-005 | merge.go, resolver.go ✅ |
| `audit_test.go` enforcement | RT-005 | audit_test.go (M2에서 placeholder → real) |
| 5개 누락 loader 본체 (constitution/context/interview/design/harness) | **MIG-003** | 5개 새 wrapper + types.go 확장 + `loader.go` `loadXxxSection` 5개 |
| sunset.yaml activate-or-retire | **MIG-003** | sunset config audit + decision |

이 분담 근거:
- spec.md §2 Out-of-scope 항목 "The 5 actual missing loaders ... — SPEC-V3R2-MIG-003".
- RT-005가 패턴 (typed struct + source-tagged merge + validator tag) 을 정립; MIG-003이 5번 적용.
- audit_test.go가 MIG-003 작업 후 5개 yaml 모두 Go struct 동반함을 강제 → 누락 시 build fail.

### 6.3 audit_test.go 강제 메커니즘

`audit_test.go:9-52` (현재 placeholder) 의 full implementation 의도:

```
1. Scan `.moai/config/sections/*.yaml` → list yaml file basenames
2. Maintain registry map: yaml basename → Go struct name
   {"constitution": "ConstitutionConfig", "context": "ContextConfig", ...}
3. For each yaml file:
   a. If basename not in registry → FAIL "orphan yaml: <file>"
4. For each registry entry:
   a. If Go struct does not exist (reflection check via `reflect.TypeOf`) → FAIL "orphan struct: <name>"
5. Optional: exception list (e.g., template-only files like .gitkeep)
```

이 테스트는 `go test ./internal/config/... -run TestAuditParity` 가 build pipeline에 포함되어 있으면 매 PR 마다 실행. MIG-003 작업 시 새 yaml 추가 → registry 추가 → struct 추가 의 3단계가 모두 일어나야 GREEN.

---

## 7. CI audit test 작동 방식

(이 섹션은 §6.3 detail의 확장.)

### 7.1 Yaml↔Go struct mapping registry

audit_registry.go (M2 신규 파일) 가 단일 진실 원천:

```go
// Map: yaml file basename (without ext) → Go struct type name
var YAMLToStructRegistry = map[string]string{
    "user":             "UserConfig",
    "language":         "LanguageConfig",
    "quality":          "QualityConfig",
    "git-convention":   "GitConventionConfig",
    "system":           "SystemConfig",
    "llm":              "LLMConfig",
    "ralph":            "RalphConfig",
    "workflow":         "WorkflowConfig",
    "state":            "StateConfig",
    "statusline":       "StatuslineConfig",
    "research":         "ResearchConfig",
    "sunset":           "SunsetConfig",
    "git-strategy":     "GitStrategyConfig",
    "project":          "ProjectConfig",
    "lsp":              "LSPQualityGates",  // 다른 이름 매핑 가능
    // MIG-003에서 추가:
    // "constitution": "ConstitutionConfig",
    // "context":      "ContextConfig",
    // "interview":    "InterviewConfig",
    // "design":       "DesignConfig",
    // "harness":      "HarnessConfig",
}

// Exceptions: yaml files that legitimately have no Go struct
var YAMLAuditExceptions = map[string]string{
    "mx.yaml":      "validation rules-only; consumed by internal/template/mx_*",
    "security.yaml": "external consumer (security policy); no Go runtime",
    "runtime.yaml": "TBD per MIG-003 sunset decision",
    "interview.yaml": "MIG-003 pending",
}
```

### 7.2 Reflection-based existence check

```go
// audit_test.go:
func TestAuditParity(t *testing.T) {
    yamlFiles := listYAMLFiles(".moai/config/sections")
    for _, file := range yamlFiles {
        base := strings.TrimSuffix(file, filepath.Ext(file))
        if _, isException := YAMLAuditExceptions[file]; isException {
            continue
        }
        structName, mapped := YAMLToStructRegistry[base]
        if !mapped {
            t.Errorf("orphan yaml file (no Go struct mapping): %s", file)
            continue
        }
        // Use reflect to verify struct exists in package
        if !structExists(structName) {
            t.Errorf("registry maps %s → %s but struct not found", file, structName)
        }
    }
}
```

reflect 기반 check는 `reflect.TypeOf(Config{}).FieldByName(structName)` 로 가능; 이는 build-time 안전 (struct rename 시 컴파일 에러 → registry update 필요 가시화).

---

## 8. RT-002 / RT-003 / RT-006 consumer 영향

### 8.1 SPEC-V3R2-RT-002 (Permission Stack)

RT-002는 Source enum + Value[T] + 8-tier merge 를 **directly import**:

```go
// RT-002 expected import
import "github.com/modu-ai/moai-adk/internal/config"

type PermissionRule struct {
    Pattern string
    Action  string
    Source  config.Source       // ← RT-005 type
    Origin  string                // ← RT-005 Provenance.Origin convention
}

func ResolvePermission(req *Request) PermissionDecision {
    rules := mergeRulesAcrossTiers(...)  // ← 8-tier walk per RT-005 algorithm
    return decision
}
```

→ RT-005가 P0 (RT-002 blocking dependency).

### 8.2 SPEC-V3R2-RT-003 (Sandbox Routing)

RT-003은 sandbox 설정을 source별로 routing (예: `SrcPolicy` sandbox → bypass; `SrcProject` sandbox → strict). 이를 위해 merge 결과의 Provenance.Source 검사:

```go
// RT-003 expected
sandboxConfig, _ := resolver.Key("sandbox", "mode")
if sandboxConfig.P.Source == config.SrcPolicy {
    applyPolicySandbox(sandboxConfig.V)
} else {
    applyDefaultSandbox(sandboxConfig.V)
}
```

→ RT-005가 P0 (RT-003 blocking).

### 8.3 SPEC-V3R2-RT-006 (Hook ConfigChange)

RT-006의 ConfigChange hook은 file path를 carries; RT-005 resolver가 `Reload(path string)` API를 노출하면 hook handler가 호출:

```go
// RT-006 hook handler
func (h *Handler) OnConfigChange(event ConfigChangeEvent) {
    err := globalResolver.Reload(event.Path)  // ← RT-005 API (REQ-011)
    if err != nil {
        log.Warn("config reload failed", "path", event.Path, "err", err)
    }
}
```

→ RT-005가 P0 (RT-006 blocking).

### 8.4 SPEC-V3R2-MIG-003 (5 Loader 추가)

MIG-003은 RT-005의 typed-Value 패턴을 5번 복제:

```go
// MIG-003 will add:
type ConstitutionConfig struct {
    Version string `yaml:"version" validate:"required"`
    Frozen  []string `yaml:"frozen"`
    // ...
}

func (l *Loader) loadConstitutionSection(dir string, cfg *Config) {
    wrapper := &constitutionFileWrapper{Constitution: cfg.Constitution}
    loaded, err := loadYAMLFile(dir, "constitution.yaml", wrapper)
    if err != nil { ... }
    if loaded { cfg.Constitution = wrapper.Constitution; ... }
}

// audit_test.go registry update:
"constitution": "ConstitutionConfig",
```

→ RT-005가 패턴 정의 + audit lint; MIG-003가 5번 반복 적용.

### 8.5 5개 yaml 사용 timing

| yaml | Go runtime 사용처 | RT-005 시점 | MIG-003 후 시점 |
|------|------------------|-------------|------------------|
| constitution.yaml | template generation only | template-only | runtime enforcement (FROZEN zone validator) |
| context.yaml | session context override | template-only | dynamic context loading |
| interview.yaml | wizard-driven init | template-only | wizard config validation |
| design.yaml | brand context for /moai design | template-only | brand context load + GAN loop config |
| harness.yaml | quality depth routing | partial (consumed by some) | full typed access |

---

## 9. Validator/v10 integration (shared with SPEC-V3R2-SCH-001)

### 9.1 Library reference

`github.com/go-playground/validator/v10` 는 SCH-001이 도입; RT-005가 consumer.

### 9.2 Schema tag application on Config struct

types.go의 Config 및 sub-structs는 RT-005 M3에서 validator tag 추가:

```go
type Config struct {
    User     models.UserConfig     `yaml:"user" validate:"required"`
    Language models.LanguageConfig `yaml:"language"`
    Quality  models.QualityConfig  `yaml:"quality"`
    // ...
}

type GateConfig struct {
    Enabled   bool         `yaml:"enabled"`
    SkipTests bool         `yaml:"skip_tests"`
    Timeouts  GateTimeouts `yaml:"timeouts"`
    AstGrepGate AstGrepGateConfig `yaml:"ast_grep_gate"`
}

type AstGrepGateConfig struct {
    RulesDir string `yaml:"rules_dir" validate:"required_if=Enabled true"`
    // ...
}
```

REQ-V3R2-RT-005-013 ConfigTypeError는 validator/v10 ValidationErrors 를 wrap 하거나, yaml unmarshal 시 type mismatch (예: `"high"` → int 필드) 검출.

### 9.3 Validator instance lifetime

기존 패키지 `internal/config/validation.go` 의 `Validate(cfg, loadedSections)` 가 entry. validator/v10 통합 시:

```go
var validate = validator.New()  // single global

func Validate(cfg *Config, loadedSections map[string]bool) error {
    if err := validate.Struct(cfg); err != nil {
        return wrapValidatorErr(err)  // → ConfigTypeError if type mismatch
    }
    // ... existing custom checks
}
```

Performance: ~50µs initial cost; ~1-5µs per Struct call.

---

## 10. Risk research (extends spec.md §8)

### 10.1 OverriddenBy 정확한 전파

Risk: spec REQ-V3R2-RT-005-012 "the overridden tier's path SHALL be added to `Provenance.OverriddenBy`" 와 `merge.go:104-109`의 구현 모두 lower-tier path 만 모음. 그러나 spec 의도는 **모든 lower tier with non-zero value** 의 path 가 모이는 것 (단순히 첫번째 매치 후 winner 가 결정되어도, 그 뒤의 tier가 같은 key로 non-zero 였다면 OverriddenBy 에 추가).

현재 구현 정확성 확인: `merge.go:103-108`의 if winningValue != nil 분기는 이미 winner가 결정된 후 lower-tier 의 non-zero 값을 OverriddenBy 에 추가. ✅ 의도 일치.

Mitigation: M2 단위 테스트에서 3-tier 시나리오 (policy=true, user=false, project=true) 검증 — winner=policy, OverriddenBy=[project_origin] 기대.

### 10.2 yaml.v3 strict mode for schema_version

Risk: `Provenance.SchemaVersion` 필드 채우기 위해 yaml 파일 최상위 `schema_version: N` 키를 모든 섹션에서 require 해야 하는가?

Decision (M4): optional. 없으면 0 default. validator tag로 강제하지 않음 (점진적 도입).

### 10.3 5개 누락 loader 와 audit_test.go 의 chicken-and-egg

Risk: 만약 MIG-003 전에 audit_test.go 를 strict 모드 (없으면 fail) 로 활성화하면, 5개 yaml 이 orphan 으로 검출되어 build fail.

Mitigation: `YAMLAuditExceptions` map에 5개를 등록 ("MIG-003 pending"). MIG-003에서 하나씩 등록 해제하며 struct 추가. RT-005가 패턴 + lint, MIG-003이 적용 → 분리된 PR.

### 10.4 prompt-cache stability vs map iteration

Risk: Go map iteration 비결정성이 `merged.Dump(json)` 출력 순서를 변동시켜 prompt-cache prefix 안정성 깨짐.

Mitigation: `dumpJSON()` 시 keys를 `sort.Strings()` 로 정렬 후 직렬화. spec §7 "byte-identical merged output" 만족.

### 10.5 Backwards compatibility (v2.x consumers)

Risk: v2.x 의 `cfg.User.Name` 직접 접근 패턴이 generic Value[T] 도입 후 `cfg.User.Name.V` 로 변경 → 다수 파일 mass refactor.

Mitigation per spec §8: BC-V3R2-015은 reader layer only; 사용자 facing API 무변경. `Manager.Get(section, key)` 와 같은 facade는 .V 추출 후 반환 → 호출 측은 변경 없음. 직접 `cfg.User.Name` 접근 패턴은 grep으로 식별; M5에서 mechanical refactor.

### 10.6 Concurrent reload race

Risk: RT-006 hook이 ConfigChange 발화 → resolver.Reload(path) 호출 중 다른 goroutine이 resolver.Key() 호출.

Mitigation: `resolver` 구조체에 `sync.RWMutex` 추가; Load/Reload 는 Lock(), Key/Dump 는 RLock(). 기존 `loader.go:18-21`의 Loader 패턴을 따라감.

### 10.7 Plugin tier ABI 호환성

Risk: v3.0에서 Plugin tier는 빈 슬롯; v3.1+에서 plugin API 도입 시 RT-005 contract 변경 필요한가?

Decision (per spec §4 Assumptions): plugin contribution 은 추가 only — Source enum 값은 변경 없음 (SrcPlugin = 4); plugin은 단순히 자신의 yaml fragment 를 등록하면 resolver 가 walk. ABI 안정성 유지.

---

## 11. External library evaluation summary

| Library / Source | Purpose | Decision |
|------------------|---------|----------|
| `github.com/go-playground/validator/v10` | Schema validation | **ADOPT** (already in go.mod via SCH-001) |
| `gopkg.in/yaml.v3` | YAML parsing | **ADOPT** (already in go.mod, used in loader.go) |
| `encoding/json` | JSON tier parsing (.claude/settings.json) | **ADOPT** (stdlib, used in resolver.go:243-249) |
| `reflect` | type-aware zero check + struct existence | **ADOPT** (stdlib, used in merge.go:138-162) |
| `fsnotify` (file watching) | diff-aware reload (REQ-011) | **DEFER to RT-006** (RT-005 ships hook-triggered API only) |
| Anthropic prompt-cache | byte-identical merged output | **ADOPT** convention (sort+stable serialize) |
| Claude Code reference (r3 §1.3) | 8-tier ordering | **ADOPT** verbatim |

---

## 12. Cross-SPEC dependency status

### 12.1 Blocked by

- **SPEC-V3R2-CON-001** (FROZEN zone declaration for 8-tier ordering): completed per Wave 6 history. The 8-tier order is `policy > user > project > local > plugin > skill > session > builtin` and is constitutionally invariant.
- **SPEC-V3R2-SCH-001** (validator/v10 integration): status check at run-phase plan-audit gate. If not yet merged, M3 adds direct `validator/v10` dependency to `go.mod`.
- **SPEC-V3R2-RT-004** (SessionStore): RT-004 provides SrcSession tier contents via runtime checkpoint writes. RT-005 reads from the session checkpoint location; RT-004 writes it. Cross-SPEC contract is the file path convention (`.moai/state/session-*.json`).

### 12.2 Blocks

- **SPEC-V3R2-RT-002** (Permission Stack): consumes `Source` enum + `Value[T]` + 8-tier merge. RT-005 must land before RT-002 implementation can begin.
- **SPEC-V3R2-RT-003** (Sandbox Routing): routes by `Provenance.Source`. RT-005 blocks.
- **SPEC-V3R2-RT-006** (ConfigChange hook): consumes `Reload(path)` API from RT-005. RT-005 blocks.
- **SPEC-V3R2-RT-007** (hardcoded path migration): GoBinPath resolver reads from SrcUser/SrcBuiltin via Value[T] wrapper. RT-005 blocks.
- **SPEC-V3R2-MIG-003** (5 loader additions): replicates RT-005 typed-Value pattern. RT-005 blocks.

### 12.3 Related (non-blocking)

- **SPEC-V3R2-EXT-004** (migration runner): applies schema migrations to config files for v3.0 → v3.1 lifts.
- **SPEC-V3R2-ORC-002** (common protocol CI lint): reuses audit_test.go pattern from RT-005.
- **SPEC-V3R2-WF-006** (output styles): inherits provenance from RT-005 resolver.
- **SPEC-V3R2-CON-003** (consolidation pass): moves settings-management rule text into `.claude/rules/moai/core/settings-management.md`.

---

## 13. File:line evidence anchors

다음 anchor는 run phase에 load-bearing. plan.md §3.4에서 verbatim cite.

(All paths are relative to repo root `/Users/goos/MoAI/moai-adk-go`. All line ranges verified via Read 2026-05-10.)

1. `.moai/specs/SPEC-V3R2-RT-005/spec.md:42-71` — In-scope items (Source enum, Provenance, Value[T], SettingsResolver, 4 file paths, merge algorithm, doctor commands, audit lint).
2. `.moai/specs/SPEC-V3R2-RT-005/spec.md:120-164` — 27 EARS REQs (REQ-001..051).
3. `.moai/specs/SPEC-V3R2-RT-005/spec.md:166-181` — 15 ACs (AC-01..15).
4. `.moai/specs/SPEC-V3R2-RT-005/spec.md:184-191` — Constraints (Go 1.22+, validator/v10, 100ms p99 cold load, 20ms diff-aware reload, 2 MiB RSS).
5. `internal/config/source.go:1-113` — Source enum + helpers (existing baseline; REQ-001 satisfied).
6. `internal/config/provenance.go:1-71` — Provenance + Value[T] (existing baseline; REQ-002, REQ-003 satisfied).
7. `internal/config/types.go:13-32` — Config root + 16 sections (existing baseline; needs 5 additions in MIG-003).
8. `internal/config/types.go:312-318` — `sectionNames` slice (16 sections registered).
9. `internal/config/types.go:335-372` — yaml file wrappers (8 wrappers existing).
10. `internal/config/loader.go:31-70` — `Loader.Load()` (8 sections wired; needs 5 additions in MIG-003).
11. `internal/config/merge.go:63-134` — `MergeAll` deterministic priority walk (REQ-005, REQ-010 satisfied; needs OverriddenBy verification).
12. `internal/config/merge.go:168-205` — `Diff(a, b *MergedSettings)` (per-key diff; REQ-007 partial).
13. `internal/config/merge.go:229-275` — `Dump(json/yaml)` (REQ-006, REQ-030 partial; placeholder JSON formatter must be replaced with encoding/json).
14. `internal/config/merge.go:138-162` — `isZero(v any)` reflection-based zero check (used by MergeAll).
15. `internal/config/resolver.go:17-29` — `SettingsResolver` interface (REQ-004 partial; signature mismatches w/ spec).
16. `internal/config/resolver.go:46-74` — `resolver.Load()` 8-tier loop (REQ-010 satisfied).
17. `internal/config/resolver.go:103-125` — `loadPolicyTier()` platform-specific paths (REQ-014 satisfied).
18. `internal/config/resolver.go:218-235` — placeholders: `loadSkillTier`, `loadSessionTier`, `loadBuiltinTier` (REQ-015 partial; needs frontmatter parsing for skills).
19. `internal/config/resolver.go:252-255` — `loadYAMLFile` placeholder returning empty map (CRITICAL: must be replaced with real yaml.v3 parsing).
20. `internal/config/resolver.go:298-355` — `Key(section, field)`, `Dump`, `Diff(a, b Source)` (REQ-007, REQ-032 satisfied; Diff shape needs spec alignment).
21. `internal/config/resolver_errors.go:9-79` — 5 error types (`ConfigTypeError`, `ConfigAmbiguous`, `PolicyOverrideRejected`, `ConfigSchemaMismatch`, `TierReadError`); types defined, raises pending.
22. `internal/config/audit_test.go:9-22` — `TestAuditParity` placeholder (`t.Skip`); CRITICAL: must be replaced with real registry-driven check.
23. `internal/config/validation.go:25-47` — `Validate(cfg, loadedSections)` (existing custom validator; integrate validator/v10 here).
24. `internal/config/validation.go:50-62` — `validateRequired` (loadedSections-aware; pattern reused in audit).
25. `internal/cli/doctor_config.go:13-167` — `doctor config dump`/`diff` Cobra commands (REQ-006, REQ-007, REQ-030, REQ-032 wired).
26. `.moai/config/sections/` directory listing — 23 yaml files; 7 without Go struct (constitution/context/design/git-strategy/harness/interview/lsp/mx/runtime/security/sunset/workflow per measurement, of which spec calls out 5).
27. `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary — subagent prohibition (consumed by all SPECs).
28. `.claude/rules/moai/workflow/spec-workflow.md:172-204` — Plan Audit Gate.
29. `CLAUDE.local.md` §6 — test isolation (`t.TempDir()` + `filepath.Abs`).
30. `CLAUDE.local.md` §14 — no hardcoded paths in `internal/`.
31. `docs/design/major-v3-master.md:L974` — §8 BC-V3R2-015 (multi-layer settings reader-only migration).
32. `docs/design/major-v3-master.md:L1046` — §11.3 RT-005 master definition.
33. `.moai/specs/SPEC-V3R2-RT-004/research.md:§7` — team-mode merge pattern (cross-reference for RT-005's resolver mutex design).

Total: **33 distinct file:line anchors** (exceeds plan-auditor minimum of 10 by 23).

---

## 14. Open Questions resolved during research

- **OQ-1**: Is `Diff(a, b Source)` 의 의미 "두 tier의 raw delta" 아니면 "merged-view delta"? → **merged-view delta** (per spec REQ-V3R2-RT-005-051 verbatim: "shall show the merged-view delta"). 기존 `resolver.go:323-355`은 "두 tier 만 merge" 하는 형태로 가까우나, 정확히는 "8-tier full merge 후 winner.Source == a 인 key 와 winner.Source == b 인 key 의 차이" 가 될 수 있음. M3 단위 테스트에서 정확한 의미를 확정.
- **OQ-2**: `Provenance.SchemaVersion` 추적은 yaml 파일 최상위 `schema_version` 키로? 아니면 별도 metadata 파일? → **yaml 최상위 키** (per spec REQ-033 "WHERE a yaml section file declares a `schema_version: N` key"). Optional; 미선언 시 0.
- **OQ-3**: RT-005가 RT-002에 직접 import 되는가, 아니면 separate package? → **directly import** `internal/config`. 8-tier ordering 일치성을 컴파일 타임 보장.
- **OQ-4**: Plugin tier가 v3.0에서 빈 슬롯이라면, walk 순서에서 skip 해야 하나? → **walk but always empty** (per spec §2 Out-of-scope item 6). resolver는 SrcPlugin tier도 호출하지만 항상 빈 map 반환; deterministic merge에는 영향 없음.
- **OQ-5**: `audit_test.go` 의 strict 모드를 RT-005에서 enable 하면 5개 yaml orphan 으로 build fail 되는가? → **YES**. Mitigation: YAMLAuditExceptions map 에 5개 등록; MIG-003에서 점진 해제.

---

End of research.md.
