# SPEC-WEB-CONSOLE-013 — Acceptance Criteria

> AC SSOT. 모든 AC는 기계 검증 가능(grep / go test / go build + 기대 출력 명시).
> AC sub-ID 규약: `AC-WC13-NNN` (+ 필요 시 소문자 접미 a/b — acceptance.md 한정).

## §D AC Matrix

### M1 — 라우팅 기반

| AC | REQ | 검증 명령 | 기대 결과 |
|----|-----|-----------|----------|
| AC-WC13-001 | 001/002 | `grep -n '"handoff"\|"cache"' internal/settings/sectionroute.go` | 두 이름 모두 `RouteSeam` 매핑 행 존재 |
| AC-WC13-002 | 002 | `grep -c '"cache"' <(go run 없이) — ExcludedSections 함수 본문 grep: awk '/func ExcludedSections/,/^}/' internal/settings/sectionroute.go \| grep -c '"cache"'` | `0` (cache가 제외군 열거에서 제거) |
| AC-WC13-003 | 002 | `awk '/func SeamSections/,/^}/' internal/settings/sectionroute.go \| grep -c 'handoff\|cache'` | ≥ 1행 (양 섹션 포함; 정확 목록은 테스트가 검증) |
| AC-WC13-004 | 003 | `awk '/sectionRootKeys = map/,/^}/' internal/settings/sectionwrite.go \| grep -c '"handoff"\|"cache"'` | `2` |
| AC-WC13-005 | 004 | `go test ./internal/settings/ -run 'TestRouteForSection\|TestSeamSections\|TestExcludedSections\|TestWriteSectionViaSeam' -v` | PASS — 신규 스코프 accept + 잔여 제외군(state/system/project/sunset/tool-policy/lsp/mx/constitution/context/design/interview/db) reject 케이스 포함 |
| AC-WC13-006 | 005 | `grep -rn 'ConfigManager\|\.Save(' internal/settings/sectionwrite.go internal/settings/sectionvalues.go \| grep -iv 'seam\|comment\|//'` | 무매치 (typed re-marshal 미도입) |
| AC-WC13-007 | 006 | seam 쓰기 characterization 테스트: cache.yaml fixture에 `enabled` 편집 후 `spec_ttl`/`min_cacheable_tokens`/주석 원문 보존 assert (`go test ./internal/settings/ -run TestCacheSeamPreservesUnexposedKeys -v`) | PASS |

### M2 — handoff + cache 노출

| AC | REQ | 검증 명령 | 기대 결과 |
|----|-----|-----------|----------|
| AC-WC13-010 | 010 | `grep -n 'SectionHandoff\|SectionCache' internal/settings/schema.go` | SectionID 2건 + AllSections 등재 |
| AC-WC13-011 | 010/012 | `go test ./internal/settings/ -run TestSchemaSections -v` (파생 방식 필드 검증 — 총수 하드코딩 금지) | PASS — handoff 2필드(mode select, guide bool) + cache 2필드(enabled bool, session_ttl select), 전부 PersistSeam kind |
| AC-WC13-012 | 011 | handlers 테스트: `handoff.mode=bogus` POST → 4xx + 파일 무변경 assert (`go test ./internal/web/ -run TestHandoffModeValidation -v`) | PASS (accept: manual/auto만) |
| AC-WC13-013 | 013 | `go test ./internal/settings/ -run TestSessionTTLClosedSetSymmetry -v` — settings측 옵션 집합 ≡ config측 validSessionTTLs {1h,5m,off} | PASS (export 재사용 또는 미러+대칭 가드) |
| AC-WC13-014a | 014 | `for k in sec.handoff sec.cache; do grep -c "\"$k\." internal/web/assets/i18n.js; done` | 각 키가 4 locale 블록 전부에 존재 (`go test ./internal/web/ -run TestI18n` PASS로 4-locale 파리티 검증) |
| AC-WC13-014b | 014 | `go test ./internal/web/ -run TestI18n -v` | PASS — 신규 키 4-locale 완전성 |
| AC-WC13-015 | 015 | `grep -c 'spec_ttl\|min_cacheable_tokens' internal/settings/schema_sections.go` | `0` (FieldDef 미생성) |
| AC-WC13-016 | 016 | `go test ./internal/cli/ -run TestSchemaBridge -v` — **schema_bridge_test.go 무변경 상태에서** 실행 | PASS (신규 필드가 PersistSeam 제외 술어(line 33 앵커: `f.Persist.Kind == settings.PersistSeam`)를 그대로 탐 — TUI측 변경 0) |
| AC-WC13-017 | 017 | `git diff --stat <M1-base>..HEAD -- internal/web/ \| grep -c 'handoff\|cache'` 전용 신규 templ 부재 확인 + generic fieldset 경로 렌더 테스트 | 신규 전용 템플릿 파일 0건 (Model Policy 뷰 제외) |

### M3 — Model Policy READ-ONLY 뷰

| AC | REQ | 검증 명령 | 기대 결과 |
|----|-----|-----------|----------|
| AC-WC13-020 | 020 | `go test ./internal/web/ -run TestModelPolicyView -v` — performance_tier 표시 + 3 perfTier × 12셀 표 렌더 assert | PASS |
| AC-WC13-021 | 021 | Model Policy 뷰 소스에서 form/input/POST 요소 grep: `grep -c '<form\|<input\|<select\|hx-post' internal/web/<modelpolicy view 파일>` | `0` (읽기 전용 — form 요소 부재) |
| AC-WC13-022 | 021 | `grep -rn 'model_routing_profiles\|performance_tier' internal/settings/schema_sections.go internal/settings/sectionapply.go` | 무매치 또는 주석뿐 (persist 바인딩 미생성) |
| AC-WC13-023 | 022 | 뷰 렌더 출력에 legacy flat 블록 부재 assert (`TestModelPolicyView_LegacyFlatHidden`) | PASS |
| AC-WC13-024 | 023 | `awk '/func agentSettingsFields/,/^}/' internal/settings/schema_sections.go \| grep -c 'workflow_agents'` | `0` (M5-a B1 유지 — FieldDef 재추가 없음) |
| AC-WC13-025 | 024 | performance_tier 빈 문자열 fixture 렌더 → "(runtime default" 문자열 assert (`TestModelPolicyView_EmptyTier`) | PASS |
| AC-WC13-026 | 026 | model_routing_profiles 부재 fixture 렌더 → fallback 상태 문구 assert, 오류 없음 (`TestModelPolicyView_AbsentBlock`) | PASS |
| AC-WC13-027 | 025 | `go test ./internal/web/ -run TestI18n -v` (M3 키 포함 재실행) + 구분 주석 키(`model_policy` vs `performance_tier`) 존재 grep | PASS |

### M4 — 검증 및 무접촉

| AC | REQ | 검증 명령 | 기대 결과 |
|----|-----|-----------|----------|
| AC-WC13-030 | 030 | `git diff --stat <base>..HEAD -- internal/statusline/` | 빈 출력 (무접촉) |
| AC-WC13-031 | 031 | `git diff --stat <base>..HEAD -- internal/template/templates/` | 빈 출력 (무접촉; 접촉 시 scope delta 보고 필수) |
| AC-WC13-032a | 032 | `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` | exit 0 × 2 |
| AC-WC13-032b | 032 | `go test ./internal/settings/... ./internal/web/... ./internal/cli/...` | PASS (전체) |
| AC-WC13-032c | 032 | `golangci-lint run --timeout=2m` | NEW 이슈 0 (선재 baseline은 별도 표기) |

## §D.1 키별 노출 판정 증거 표 (verification-claim-integrity §1.1 surface 3 — SSOT)

plan-phase(2026-07-10) grep 실측. run-phase pre-flight(§C-4)가 재실행해 판정 전제가 유지되는지 확인한다 — 전제가 바뀌면(예: InjectCacheControl 배선 착지) blocker report로 판정 재검토.

| 키 | 판정 | 증거 명령 | 관측 결과 (2026-07-10) |
|----|------|-----------|------------------------|
| `handoff.mode` | **editable** | `grep -rn "Handoff.Mode" internal/ --include="*.go" \| grep -v _test` | `internal/hook/handoff_inject.go:154` — SessionStart 핸들러가 매 세션 시작마다 소비 |
| `handoff.guide` | **editable** | `grep -rn "Handoff.Guide" internal/ --include="*.go" \| grep -v _test` | `internal/hook/handoff_inject.go:159` — notice-only 셀 stderr 힌트 게이트 |
| `cacheStrategy.enabled` | **editable (caveat)** | `grep -rn "LoadCacheConfig" internal/ cmd/ --include="*.go" \| grep -v _test` | `internal/cli/doctor_cache.go:48` 소비. **Caveat**: `InjectCacheControl` 비테스트 호출자 0건 — 현재 유효 반경은 doctor 표시 (배선은 PROMPT-CACHE-001 계보 부채) |
| `cacheStrategy.session_ttl` | **editable (caveat)** | 상동 + `grep -rn "SessionTTL" internal/runtime/cache_control.go` | 주입기 내부 소비 구현은 실존하나 주입기 자체 미배선 — 동일 caveat |
| `cacheStrategy.spec_ttl` | **omit** (미노출) | 브리프 스코프 2키 + 상동 | seam unmodeled-key 보존으로 파일 값 무손상 (AC-WC13-007) |
| `cacheStrategy.min_cacheable_tokens` | **omit** (미노출) | 상동 | 상동 |
| `llm.performance_tier` | **read-only 표시** | `grep -rn "PerformanceTier" internal/ --include="*.go" \| grep -v _test` | 기록자: `moai init --model-policy` → `template.ApplyPerformanceTier`. 라우팅 소비자: 없음 (RouteModelFor 미배선) → 편집 금지 |
| `workflow.model_routing_profiles.*` (3×12) | **read-only 표시** | `grep -rn "RouteModelFor" internal/ cmd/ --include="*.go" \| grep -v _test \| grep -v "internal/config/"` | **0건** — TOKEN-ROUTING-001 D1 wiring 부채. 편집 소비 런타임 부재 |
| `workflow.model_routing` (legacy flat) | **hidden** (완전 미렌더) | `grep -rln "ModelRouting\b" internal/ cmd/ --include="*.go" \| grep -v _test` | `internal/config/{model_routing,types}.go`뿐 — DEPRECATED alias, 외부 소비자 0 |
| `workflow.workflow_agents` | **hidden 유지** (M5-a B1) | `grep -rn "workflow_agents" internal/settings/*.go \| grep -v _test` | `schema_sections.go:262` 주석 — 웹 렌더 숨김, dynamic-workflow JS 직접 읽기. 재노출은 별도 사용자 결정 |

## §D.2 Given-When-Then 시나리오

**시나리오 1 — handoff.mode 편집 왕복 (happy path)**
- Given: t.TempDir() 프로젝트에 template 기본 `handoff.yaml` (mode: manual, guide: false, 주석 포함)
- When: 웹 콘솔 save 경로로 `handoff.mode=auto` 제출
- Then: `handoff.yaml`의 `mode:` 스칼라만 `auto`로 변경; 파일 주석 원문 보존; `internal/config` Loader 재로드 시 `cfg.Handoff.Mode == "auto"`; `handoff_inject.go` 소비 경로가 auto 분기 진입 가능 상태

**시나리오 2 — cache session_ttl 잘못된 값 거부 (unwanted path)**
- Given: 기본 `cache.yaml` (enabled: false, session_ttl: "1h", spec_ttl: "5m", min_cacheable_tokens: 2048)
- When: `cacheStrategy.session_ttl=2h` (닫힌 집합 밖) 제출
- Then: 4xx 응답 + `cache.yaml` 바이트 무변경 (spec_ttl/min_cacheable_tokens 포함 전체 보존)

**시나리오 3 — Model Policy 뷰 빈-값/부재 강건성 (edge)**
- Given: `llm.performance_tier: ""` + workflow.yaml에서 model_routing_profiles 블록 제거한 fixture
- When: Model Policy 뷰 렌더
- Then: "(runtime default: medium)" 표시 + fallback 상태 문구 렌더, 오류/패닉 없음, form 요소 0

## §D.3 Edge Cases

- handoff.yaml 파일 자체 부재 프로젝트: 섹션 렌더는 기본값 표시, 저장 시 seam upsert가 최상위 `handoff:` 키 생성 (sectionRootKeys 가드 내) — 011 REQ-WC11-073 upsert 선례.
- `handoff.mode`에 대소문자 변형(`Auto`) 제출 → 닫힌 집합 불일치로 4xx (대소문자 정규화 도입 금지 — handoff_inject.go 소비 값과 문자 그대로 일치해야 함).
- 로컬 dogfood `mode: auto` 상태에서 테스트 실행 — 실 프로젝트 파일 접촉 금지 (t.TempDir 격리).
- model_routing_profiles에 미지 perfTier 키가 존재하는 fixture → read-only 표는 있는 그대로 표시 (validation은 loader 소관, 뷰는 렌더만).

## §D.4 Definition of Done

- [ ] §D AC 매트릭스 전 항목 PASS (E1 verbatim 출력 인용)
- [ ] §D.1 증거 표의 전제 grep이 run-phase 시점 재실행으로 유지 확인 (변동 시 blocker report)
- [ ] cross-platform 빌드 2종 exit 0
- [ ] i18n 4-locale 파리티 테스트 PASS (M2 + M3 키)
- [ ] statusline / template 트리 diff 빈 출력
- [ ] golangci-lint NEW 0
- [ ] 커밋 subject `feat(SPEC-WEB-CONSOLE-013): M{N} ...` + `🗿 MoAI` trailer, main 직진 push (Route A)
