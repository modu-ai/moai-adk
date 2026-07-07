---
id: SPEC-INTERNAL-ARCH-001
status: draft
updated: 2026-07-08
---

# SPEC-INTERNAL-ARCH-001 — Research (결합 증거 + 선행 사례)

> 본 문서의 모든 증거는 2026-07-08 SPEC 저작 세션에서 read-only 재검증한 실측이다(커맨드 + verbatim 출력). 감사 보고서의 주장과 실측 사이의 delta는 §B에 별도 기록한다 — verification-claim-integrity §2 (baseline 귀속).

## §A 실측 증거 (finding별)

### F1 — CLI 전역 deps singleton + import-cycle 회피 주석

```
$ grep -n "var deps" internal/cli/deps.go
76:var deps *Dependencies

$ grep -rn -i "import cycle" internal/cli/agentlint/agent_lint.go internal/cli/specid/specid.go internal/cli/preference/cmd.go
internal/cli/agentlint/agent_lint.go:21:// cli package, so it cannot import internal/cli without an import cycle. These
internal/cli/specid/specid.go:6:// worktree는 cli를 import할 수 없다(import cycle). 본 package는 cli와 worktree
internal/cli/preference/cmd.go:141:  // does not depend on internal/hook (avoiding an import cycle with a
```

`internal/cli/update_preserve_inventory.go` ~L176 인근 주석(실측 원문 발췌):

```
// inject the baseline reader through `baselineReader` rather than directly
// import the template package here (avoids circular imports — internal/cli
// already imports internal/template; the template package is the consumer
// of cli for some adjacent helpers).
```

→ 표현이 "import cycle"이 아닌 "**circular imports**"임. AC-ARCH-002c의 grep은 `import cycle|circular import` 양쪽을 커버하도록 정의했다. 또한 4곳의 cycle은 축이 다르다: agentlint/specid는 **cli↔subpackage** 축(본 SPEC M1이 직접 해소), preference는 **cli↔hook** 축, update_preserve_inventory는 **cli↔template** 축(둘은 seam 2안 승격 또는 후속 SPEC 연결점 — design.md §A.2/§G). pilot AC는 cli↔subpackage 축 파일에만 binding한다.

### F2 — monolith 파일 크기

```
$ wc -l internal/cli/update.go internal/cli/hook.go
    3172 internal/cli/update.go
    1182 internal/cli/hook.go
```

update.go는 차상위 파일(hook.go)의 ~2.7배. 하나의 컴파일 단위에 binary self-update / config 3-way merge(≈L1637-2177) / archive-drift / namespace 보호 / SEC-HARDEN-003·004·005 path-guard 가족(restoreTargetContained·parentChainContained, ≈L2199-2400)이 혼재.

### F3 — internal/core bare namespace

```
$ ls internal/core/*.go
(eval):1: no matches found        # 직접 .go 파일 0개 — Go 패키지가 아님

$ find internal/core -type f | sort   (요약)
internal/core/git/         → 13개 .go (worktree/branch/conflict/event/manager 등 VCS ops) + .gitkeep
internal/core/project/     → 16개 .go (detector/initializer/validator 등 scaffolding) + .gitkeep
internal/core/quality/     → 6개 .go (trust/validators — TRUST5) + .gitkeep
internal/core/integration/ → .gitkeep 만 (Go 코드 0)
internal/core/migration/   → .gitkeep 만 (Go 코드 0)
```

git/project/quality 3개 subpackage는 각각 내부 응집적이며 공통 도메인이 없다. 감사 fan-in 수치(5/3/2 files, disjoint)는 run-phase 이동 커밋 시 재실측한다(이동 커밋의 call-site 갱신 목록이 곧 재실측 증거).

### F4 — config 이중 pipeline

```
$ grep -rn "NewResolver(" --include="*.go" internal/ cmd/ | grep -v _test
internal/config/resolver.go:56:func NewResolver() SettingsResolver {
internal/cli/mx_query.go:106:      resolver := mx.NewResolver(mgr)      ← 별개 심볼 (internal/mx)
internal/cli/doctor_config.go:50:  resolver := config.NewResolver()
internal/cli/doctor_config.go:102: resolver := config.NewResolver()
internal/mx/resolver.go:22:func NewResolver(manager *Manager) *Resolver {

$ wc -l internal/config/resolver.go internal/config/loader.go internal/config/manager.go
    1156 resolver.go     474 loader.go     438 manager.go
```

- `config.NewResolver`의 production 소비처는 **doctor_config.go 단일 파일(호출 2회, L50/L102)** 뿐.
- `applyEnvOverrides`(manager.go, 실측 본문 확인)는 정확히 4종만 처리: `EnvDevelopmentMode`, `EnvLogLevel`, `EnvLogFormat`, `EnvNoColor`. resolver.go의 tier loader들은 이들 env를 읽지 않음 → **doctor 진단과 런타임 해석의 precedence 불일치** 성립.
- grep 시 `mx.NewResolver`(internal/mx — MX tag resolver) 오검출 주의: AC-ARCH-005a는 `config\.NewResolver(` 수식 grep으로 정의.

### F5 — loader/manager boilerplate

```
$ grep -n "^func " internal/config/loader.go   (발췌)
115 loadUserSection / 129 loadLanguageSection / 145 loadQualitySection /
159 loadGitConventionSection / 175 loadGitStrategySection / 189 loadLLMSection /
203 loadStateSection / 220 loadWorkflowSection / 234 loadStatuslineSection /
250 loadRalphSection / 270 loadResearchSection / 287 loadFeedbackSection /
306 loadHandoffSection                          → 계 13개 near-identical
```

각 메서드는 wrapper 구성 → `loadYAMLFile`(L459) → 필드 대입 → `loadedSections[name]=true`의 동형 패턴. manager.go에는 Save의 saveSection 반복 호출 + getSectionLocked/setSectionLocked 병렬 switch가 병존(감사 anchor L181-221/L284-389 — run-phase에서 심볼 기준 재확정).

### F6 — env-var 문서 drift

```
$ grep -n "MOAI_USER_NAME\|MOAI_CONVERSATION_LANG" internal/config/CLAUDE.md
12: ... (`MOAI_USER_NAME`, `MOAI_CONVERSATION_LANG`, ...) override file values ...
13: ... (e.g., `EnvUserName = "MOAI_USER_NAME"`) ...

$ grep -n "MOAI_USER_NAME\|EnvUserName" internal/config/envkeys.go
(exit 1 — 0건. envkeys.go 실존 상수는 MOAI_CONFIG_DIR/DEVELOPMENT_MODE/LOG_LEVEL/
LOG_FORMAT/NO_COLOR + statusline/update/git/test 계열 — user/lang 계열 부재)
```

CLAUDE.md는 미구현 env override를 사실처럼 서술하고(L12), 존재하지 않는 상수 `EnvUserName`을 예시로 인용한다(L13). CLAUDE.local.md §9의 "NOT currently implemented" 기재와 정합 — 문서 측 결함 확정.

## §B 감사 주장 vs 실측 delta (정직 기록)

| # | 감사 주장 | 실측 | 처리 |
|---|-----------|------|------|
| D1 | section loader "~17개" (L114-317), `loadDesignSection` 존재 | **13개** (L115-306), loadDesignSection 부재 | 감사 시점 이후 축소 추정(config-diet 계열). REQ-ARCH-005 성립에 영향 없음 — AC baseline을 13으로 고정 |
| D2 | `flattenStruct`가 loader.go:562 | **resolver.go:562** (loader.go는 474줄) | resolver 은퇴 시 flatten 로직 처리 결정이 필요해짐 — design.md §D에 반영 |
| D3 | update_preserve_inventory.go:176 "import cycle" 주석 | 표현은 "**circular imports**" (의미 동일) | AC grep 패턴을 양쪽 커버로 확장 |
| D4 | NewResolver "exactly ONE production call site" | 단일 **파일**(doctor_config.go), 호출 **2회**(L50/L102) | 실질 동일(소비처 1곳). AC는 "call site 0건" grep이므로 무영향 |

## §C 현재 test baseline (2026-07-08 실측 — 부채 메모리 인용)

`go test ./...` 재실측 결과 2건 FAIL 잔존:

1. `internal/spec` `TestCloseSubjectDoctrineAmendment/lifecycle-sync-gate.md` — AC-DLC-011 위반(rules audit 잔여 regression 추정)
2. `internal/statusline` `TestBuild_WritesContextUsageWithSessionID` — `context_window_size=1000000, want 256000` (env/state flaky, pre-existing)

→ sibling **SPEC-INTERNAL-TEST-001**이 이 복구를 담당. 본 SPEC은 REQ-ARCH-008로 green baseline을 선행 gate화했다. **red baseline 위 리팩터링은 "행위 보존" 판정 불능**이므로 절대 조건.

## §D 선행 사례 (prior art)

### SPEC-CLI-SUBPKG-SPLIT-001 — M1 CHECKPOINT STOP

- Tier L로 internal/cli flat 패키지 분할을 시도, **M1(agentlint cluster)만 추출 후 plan.md §F CHECKPOINT(REQ-CSS-010, spec.md:123)에서 "STOP (ship M1 only)"** 판정. 나머지 cluster(uikit/profile/migrate/update)는 tri-axis coupling으로 추출 불가 → 별도 SPEC 이관.
- 관련 커밋: run 0ee246ad9(agentlint 추출), c9abe2ddb(M1 coupling blocker re-sequence), 85adab898(close backfill).
- **본 SPEC과의 관계**: 그 STOP의 근본 원인 중 deps-축(전역 singleton과 그것을 캡처하는 커맨드 파일들)이 REQ-ARCH-001의 직접 대상이다. 즉 본 SPEC은 SUBPKG-SPLIT 후속 cluster 추출의 **전제 조건**을 만든다(추출 자체는 Out of Scope).
- 재사용 기법: 행위 보존 검증에 `moai --help` byte-identical + host/windows cross-build exit 0 사용(M1 AC 7/7 PASS 실적) → AC-ARCH-008b / plan.md §E E2에 승계.
- CHECKPOINT 패턴 승계: plan.md CHECKPOINT-1은 REQ-CSS-010의 PROCEED/RESCOPE/STOP 3-way 판정 구조를 그대로 따른다. STOP은 실패가 아닌 정당한 종결 경로라는 선례 확립.

### 리포 운영 교훈 (memory 승계)

- 공유 체크아웃 동시 커밋 레이스(양방향 staged 삼킴) → specific-path commit + spawn 전 겹침 확인 (plan.md §C/§D 반영)
- 격리 worktree의 병렬 main 전진 마스킹 → landing 직전 rebase 재확인 (plan.md §C)
- defect-claim은 도구 검증 선행 → 본 research 전체가 그 규율의 산물(§A 전 앵커 실측)

## §E 잔여 불확실성 (Gaps / Residual risk)

- **미검증**: update.go/hook.go 내부의 정확한 심볼-관심사 경계(L앵커는 근사치) — run-phase ANALYZE에서 심볼 단위 확정 필요.
- **미검증**: internal/core 3개 subpackage의 fan-in 정확 수치(감사 5/3/2 미재실측) — 이동 커밋에서 컴파일러가 전량 강제하므로 리스크 낮음.
- **미검증**: resolver.go 8-tier 의미의 잠재 소비 예정처(로드맵상) — design.md §D decision gate가 흡수.
- **미검증**: quality.yaml `development_mode: tdd` pin 하에서 cycle_type=ddd 명시 위임의 운영 정합 — orchestrator 위임 문구로 해결(plan.md §D), 코드 영향 없음.
- **환경 리스크**: 병렬 세션이 internal/cli를 접촉 중일 가능성 — pre-flight에서 매회 재확인(정적 보장 불가).
