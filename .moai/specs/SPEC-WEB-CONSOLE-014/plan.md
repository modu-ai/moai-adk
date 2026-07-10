---
id: SPEC-WEB-CONSOLE-014
status: completed
created: 2026-07-10
updated: 2026-07-11
---

# Plan — SPEC-WEB-CONSOLE-014 (v0.2.0)

## §A Context

- **작업 위치**: `/Users/goos/MoAI/moai-adk-go` (Tier M — PR 여부는 orchestrator의 Tier-based Routing 결정).
- **산출물**: `.moai/specs/SPEC-WEB-CONSOLE-014/{spec,plan,acceptance,progress}.md` (Tier M 3-artifact + progress §E skeleton).
- **스코프 SSOT**: spec.md §A.1 실측 Findings F1-F12 + §B 16 REQ. 노출 판정 기준은 §A.2 3-계층 원칙.
- **근거 조사**: 2026-07-10 manager-spec plan-phase 실측 (acceptance.md §B에 command+output 기록). 모든 파일:라인 앵커는 실측 시점 기준 — run-phase 착수 시 §C에서 content-token 기준 재검증 의무.

### §A.1 의존성 gate (depends_on: SPEC-WEB-CONSOLE-013)

- 013(Model Policy 페이지)과 **공유 파일**: `internal/settings/schema_sections.go`, `internal/settings/sectionroute.go`(+각 테스트), `internal/web/assets/i18n.js`. 병렬 진행 시 충돌 확실 → 직렬 강제.
- **iter-2 갱신(2026-07-10) 실측**: `.moai/specs/SPEC-WEB-CONSOLE-013/`은 디스크에 **존재한다** (`status: draft`, created 2026-07-10 — 본 SPEC iter-1 authoring 직후 병렬 authored; iter-1의 "미존재" 기술은 stale). gate 의미는 유지: 013의 run 완결(공유 파일 변경 landed) 이전에 014 run에 진입하지 않는다 — pre-flight C-1이 013의 status와 공유 파일 변경 상태를 확인하고, 미충족 && orchestrator의 명시적 순서 해제 없음 → blocker report.
- REQ-WC14-052에 해당하는 가드(=REQ-WC14-050의 `workflow.model_routing` 항목)는 013이 typed Model Policy 표면을 신설하더라도 **legacy flat 블록**이 편집 표면에 유입되지 않음을 고정하는 것 — 013 결과물과 충돌하지 않게 prefix를 `workflow.model_routing`(flat 경로)으로 한정한다. 013이 다른 필드 네임스페이스를 쓰는지 run 시점 재확인 (C-1b).

### §A.2 대상 파일 인벤토리 (run-phase 예상, 실측 기반)

| 파일 | 변경 성격 |
|------|-----------|
| `internal/settings/schema_sections.go` | merge_method 3필드 추가(typed), learning.auto_apply 편집 FieldDef 철거 + ReadOnlyDisplayFields 추가, observability.hook_metrics.output_path 편집 FieldDef 철거 + read-only 강등(iter-2 D2), RawViewBlocks **6건** 추가(learning.tier_thresholds / learning.rate_limit / security.sandbox.network_allowlist / security.sandbox.env_scrub_extra / mx.danger_categories / mx.test_paths — iter-2 D4 카운트 정정), mx 렌더 그룹 등재 |
| `internal/settings/schema_sections_test.go` | 라운드트립/거부/raw 테스트 확장 + 신규 가드 테스트(denylist) |
| `internal/settings/sectionroute_test.go` | RouteExcluded 핀 확장(sunset/tool-policy/mx 명시) |
| `internal/settings/schema.go` | (조건부) mx 렌더용 SectionID 추가 — 실측: SectionID는 string 타입, 기존 확장 섹션 상수 위치 확인 후 |
| `internal/web/fieldsets.templ` + `fieldsets_templ.go`(생성물) | mx raw-only 렌더 그룹 + read-only auto_apply 표시 |
| `internal/web/assets/i18n.js` | 신규 f.* / sec.mx.* / 설명 키 ×4 locale |
| `internal/web/*_test.go` | mx raw view 렌더 테스트(TestMXRawViewRendered 신설), i18n 키 검증 확장 |
| `internal/cli/schema_bridge_test.go` | TUI parity 갱신 (011 M2b 교훈 — 필드 추가/철거 양방향) |

**무접촉**: `internal/statusline/*` (REQ-WC14-062), `internal/settings/sectionroute.go` 본체(라우팅 맵 무변경 — 테스트만 확장), `.moai/config/sections/*.yaml`(값 무변경), `internal/config`/`internal/harness`/`internal/mx`/`internal/sandbox`(reader 측 무변경).

## §B Known Issues / Ground-truth 정정

- **B1 (브리프 정정 1)**: 위임 브리프는 `learning.rate_limit.{max_per_week,cooldown_hours}`를 편집 노출 대상으로 지목했으나, 실측상 enforcement rate limiter는 컴파일 상수(`rate_limit.go` `rateLimitMaxUpdates=3`, `rateLimitCooldown=24h`)를 쓰고 `NewRateLimiter(statePath)`는 config 값을 받지 않는다. 편집 노출 = config theater(008 "honest hybrid" 반례) → read-only raw view로 격하. 반박 증거 발견 시(파이프라인 어딘가 config 소비) blocker report로 스코프 재론.
- **B2 (브리프 정정 2)**: `security.permission.{pre_allowlist,session_rules}`는 internal/config에 struct 바인딩이 없고(grep 공집합) `stack.go PreAllowlist()`는 SrcBuiltin 내장이다. security.yaml 주석 자체가 "실제 패턴은 stack.go에 정의"라고 명시 → 노출 아닌 금지 가드 대상.
- **B3**: `config.ValidMergeMethods()`는 map range로 슬라이스를 만들어 **순서 비결정**. select 옵션 파생 시 정렬(sort) 후 사용 — 렌더/golden 안정성. 리터럴 재선언 금지 제약(REQ-WC14-011)과 양립: 정렬은 파생이지 재선언이 아님.
- **B4**: mx는 현재 렌더되지 않는 섹션 — raw-only(편집 필드 0개) 렌더 그룹을 기존 파이프라인이 지원하는지 미실측. 미지원이면 fieldsets.templ에 raw 블록 전용 렌더 분기 추가 (스칼라 필드 추가 금지 — REQ-WC14-031). Pre-flight C-4.
- **B5**: `fieldsets_templ.go`는 templ 생성물 — `.templ` 수정 후 재생성 필요 (011 run 절차 재사용; 생성기 부재 시 011 커밋 이력에서 재생성 방법 확인).
- **B6**: `TestExcludedSectionsAllRejected`는 `ExcludedSections()` 길이 12를 핀 — 본 SPEC은 ExcludedSections를 변경하지 않으므로(멤버십 그대로) 영향 없음. 길이 핀을 깨는 변경은 스코프 위반 신호.
- **B7**: `go test -run <이름>` 무매치 시 exit 0 함정 — 신규 테스트 바인딩 AC는 반드시 func-grep 선행 후 -run 실행 (acceptance.md AC 명령에 반영).
- **B8**: 라인 앵커 drift 비대칭 — 모든 앵커는 content-token(함수명/상수명/키 리터럴) 우선으로 재탐색.
- **B9**: `depends_on: SPEC-WEB-CONSOLE-013` forward reference의 spec-lint 거동은 authoring 직후 lint 실행 결과를 §E에 기록 (경고 발생 시 문서화된 debt로 수용 — 013 authoring 완료 시 자연 해소).
- **B10**: `learning.auto_apply` 편집 필드 **철거**는 기존 노출의 축소 — TUI bridge parity(TestTUIRendersSchemaFieldSet 등)와 웹 라운드트립 테스트(TestApplySchemaEditsAllFieldsRoundTrip)가 해당 필드를 순회 중이면 함께 갱신 필요 (REQ-WC14-061의 "양방향" 의미). iter-2 D2로 `observability.hook_metrics.output_path` 철거도 동일 클래스에 합류.
- **B11 (iter-2 D1)**: security.sandbox 리스트 키는 **config-미배선 scaffold** — `SecuritySandbox` 구조체는 로드되나 sandbox Options로의 브리지가 없다(Options 유일 populator = doctor_sandbox.go 내장 Default; `EnvScrubExtra`는 Options 필드 부재, `ScrubEnv(parent, passthrough)` 파라미터 없음 — env.go 주석은 미구현 약속). raw view 라벨은 F2-style 정직 라벨 필수. 반증 발견 시 blocker (C-3 확장분).
- **B12 (iter-2 D2)**: `observability.hook_metrics.output_path`는 dead config — non-schema reader 0건, 기록 경로는 `hookMetricsRelPath` 상수 고정(post_tool_duration.go). `slow_hook_threshold_ms`만 로컬 struct가 디코드. 편집 철거 + read-only 강등 + 쓰기 거부 (REQ-WC14-040).
- **B13 (iter-2 D5)**: live git-strategy.yaml에 `merge_method` 키 부재(grep -c = 0) vs 템플릿 .tmpl 3건 — absent-key 초기 표시는 검증 실패 없이 컴파일 기본(`squash`) 표시, **empty는 enum 비멤버**(저장은 항상 명시값 기록). 현재값 리더가 부재 키를 빈 문자열로 읽는 경로의 select 렌더 처리 확인 필요 (AC-010c).
- **B14 (iter-2 ordering)**: M1 denylist에 `learning.auto_apply`·`observability.hook_metrics.output_path`를 선탑재하면 M1 시점 RED(두 키는 아직 editable) — 두 항목은 **M2 강등과 함께** denylist에 추가한다(TDD RED→GREEN 쌍). M1 탑재분은 현행 코드 기준 즉시 GREEN 항목만.

## §C Pre-flight (run 착수 시, 순서대로)

1. **C-1 의존성 gate**: `ls .moai/specs | grep WEB-CONSOLE-013` + 013 frontmatter status 확인. 미존재/미완결 && orchestrator 순서 해제 없음 → blocker report. **C-1b**: 013 산출물의 필드 네임스페이스가 `workflow.model_routing` flat prefix와 충돌하지 않는지 확인.
2. **C-2 앵커 재실측** (content-token): `readTierThresholds`(internal/cli/hook.go), `rateLimitMaxUpdates`(internal/harness/safety/rate_limit.go), `validMergeMethods`(internal/config/validation.go), `SecuritySandbox`(internal/config/types.go), `LoadDangerConfig`(internal/mx/danger_category.go), `hook_metrics` seam 필드(schema_sections.go), `PreAllowlist()`(internal/permission/stack.go).
3. **C-3 B1/B2/B11/B12 반증 탐색**: `grep -rn "MaxPerWeek\|CooldownHours" internal/ --include="*.go" | grep -v _test` 재실행 — enforcement 소비자 신규 발견 시 blocker. `grep -rn "pre_allowlist\|session_rules" internal/config internal/permission --include="*.go"` 재실행 — struct 바인딩 신설 발견 시 blocker. **iter-2 추가 (F6)**: `grep -rn "Security.Sandbox\|Sandbox\.NetworkAllowlist\|Sandbox\.EnvScrubExtra" internal/ --include="*.go" | grep -v _test | grep -v "internal/config/types.go"` — config→Options 브리지 신설 발견 시 F6 재분류 blocker. **iter-2 추가 (F9)**: `grep -rn "output_path\|OutputPath" internal/hook internal/observability --include="*.go" 2>/dev/null | grep -v _test` — output_path reader 신설 발견 시 F9 재분류 blocker.
4. **C-4 렌더 파이프라인 확인**: RawViewBlocks가 Section 단위로 렌더되는 위치(fieldsets.templ) 실측 → 편집 필드 0개 섹션의 렌더 가능성 판정, 필요 시 raw 전용 분기 설계.
5. **C-5 병렬 세션 방어**: `git diff --cached --stat` + `git status --porcelain internal/settings internal/web` — 겹침 발견 시 중단 (공유 체크아웃 레이스 lesson).
6. **C-6 템플릿 미러 실측**: `ls internal/template/templates/` 하위에 본 SPEC 대상 파일(i18n.js, settings Go 소스) 미러가 없음을 확인 (Go 소스+embedded asset은 템플릿 트리 무관 — 매 SPEC 재실측 원칙).
7. **C-7 i18n locale 앵커**: i18n.js의 en/ko/ja/zh 오브젝트 경계 재실측 (실측 시점 line 20/368/716/1064 — drift 전제).

## §D Constraints

- spec.md §D 5항 승계. 추가 run-phase 제약:
- 커밋 단위: milestone별 Conventional Commits (`feat(SPEC-WEB-CONSOLE-014): M<N> ...`), 첫 run 커밋에서 frontmatter `draft → in-progress` (manager-develop 소유).
- 가드 테스트(M1)는 **현행 코드에서 즉시 GREEN이어야 정상** — regression pin/characterization이며 TDD RED 대상이 아니다. RED가 나오면 그 자체가 현행 노출 결함 발견 → blocker report. 단 (B14): `learning.auto_apply` / `observability.hook_metrics.output_path` denylist 항목은 M1 탑재 대상이 아니라 M2 강등과 함께 추가되는 TDD RED→GREEN 쌍이다.
- 편집 필드 추가(M3)는 TDD: 라운드트립 테스트 선행 RED → 구현 GREEN.
- i18n 키는 en 기준 작성 후 ko/ja/zh 번역 — 누락 locale 금지 (AC-060).

## §E Self-Verification

- E1: acceptance.md §C AC Matrix 23건 PASS/FAIL 표 (명령 verbatim 출력 첨부, 파일-redirect 계약 준수).
- E2: `go build ./...` + `go test ./internal/settings/... ./internal/web/... ./internal/cli/...` exit 0.
- E3: 커버리지 — `go test -cover ./internal/settings/` (기존 수준 유지 이상).
- E4: subagent-boundary grep (관례 준수).
- E5: `golangci-lint run` 대상 패키지 clean.
- E6: push 상태 보고 (orchestrator 지시 범위 내).
- E7: `git diff --name-only <base>..HEAD -- internal/statusline/` 출력 0줄 (REQ-WC14-062).

## §F Milestones (우선순위 순 — 시간 추정 금지)

- **M0 — Pre-flight**: §C 1-7 전부. C-1 실패 시 여기서 종료(blocker).
- **M1 — 노출 금지 가드 (P3 선행)**: REQ-WC14-050/051/012/021/031/040. denylist 테이블 주도 가드 테스트 + RouteExcluded 핀 + observability/hook_metrics 회귀 핀. 전부 즉시 GREEN 기대 (§D). 가드를 먼저 깔아 이후 M2-M4 자체 변경도 가드 아래서 진행.
- **M2 — learning + observability 정직화**: REQ-WC14-001/002/003 + REQ-WC14-040(output_path 강등 분). auto_apply·output_path 편집 철거→read-only 표시(+설명 i18n), tier_thresholds/rate_limit raw view. 두 철거 키의 denylist 항목 추가는 이 milestone에서 강등과 함께 수행 (B14 — TDD RED→GREEN 쌍).
- **M3 — merge_method**: REQ-WC14-010/011. 3필드 select(정렬된 파생 옵션) + typed 라운드트립/거부 테스트 (TDD).
- **M4 — raw views (security + mx)**: REQ-WC14-020/030. RawViewBlocks 4건(security 2 + mx 2; M2의 learning 2건과 합산 총 6건) + F6 정직 라벨 + mx 렌더 그룹(B4 판정 반영) + TestMXRawViewRendered.
- **M5 — i18n + parity sweep**: REQ-WC14-060/061. i18n.js 4-locale 키 일괄 + 웹 i18n 테스트 확장 + internal/cli/schema_bridge_test.go 갱신.
- **M6 — 최종 검증**: §E E1-E7 + REQ-WC14-062 무접촉 확인 + full suite.

의존 그래프: M0 → M1 → {M2, M3, M4 — 동일 파일(schema_sections.go) 순차} → M5 → M6.

## §G Anti-Patterns

- **AP-1**: enforcement에 배선되지 않은 config를 편집 가능으로 노출 (config theater — B1).
- **AP-2**: enum 옵션 리터럴 재선언 (`[]string{"squash",...}` 하드코딩) — 반드시 config SSOT 파생.
- **AP-3**: `internal/statusline/` 접촉.
- **AP-4**: 리스트 편집 위젯/FieldType 신설.
- **AP-5**: ExcludedSections 멤버십 또는 sectionRoutes 맵 변경.
- **AP-6**: func-grep 없는 `go test -run` AC (무매치 exit 0 공허 GREEN).
- **AP-7**: 일부 locale에만 i18n 키 추가 (en-only).
- **AP-8**: 013 미완결 상태에서 공유 파일 선행 수정 (직렬화 위반).
- **AP-9**: 가드 테스트 RED를 "수정해서 GREEN"으로 처리 — 가드 RED는 발견이며 blocker report 대상.
- **AP-10**: M1 denylist에 아직-editable 항목(learning.auto_apply, observability.hook_metrics.output_path) 선탑재 — M1 즉시-GREEN 원칙 위반 (B14).

## §H Cross-References

- SPEC-WEB-CONSOLE-011 — 필드 SSOT/seam/raw view/ReadOnly 메커니즘 확립 + M2b parity 교훈.
- SPEC-WEB-CONSOLE-013 — depends_on (Model Policy 페이지; 공유 파일 직렬).
- SPEC-MERGE-METHOD-CONFIG-001 — merge_method enum/consumer 계약 (REQ-MMC-007).
- SPEC-WEB-CONSOLE-008 — "honest hybrid" config theater 제거 선례.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 — 노출 판정 증거 의무.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — status 전이 소유권.
