---
id: SPEC-WEB-CONSOLE-014
status: draft
created: 2026-07-10
updated: 2026-07-10
---

# Acceptance — SPEC-WEB-CONSOLE-014 (v0.1.0)

> 검증 SSOT. 모든 AC는 기계 검증(go test 타깃 / grep assertion). 파일:라인 앵커는 run-phase에서 content-token 기준 재확인 후 사용.
> Severity: **[B]** = blocking (DoD 필수), **[N]** = non-blocking (권장).
> 신규 테스트 바인딩 AC는 `go test -run` 무매치 exit 0 함정 방지를 위해 **func-grep 선행**을 명령에 포함한다 (plan.md B7).

## §A Given-When-Then 시나리오

### S1 — merge_method 편집 저장 (REQ-WC14-010/011)
- **Given**: 콘솔이 로드되고 git-strategy 섹션에 3개 mode의 merge_method select가 렌더된 상태
- **When**: 사용자가 `git_strategy.team.merge_method`를 `rebase`로 변경해 저장
- **Then**: `.moai/config/sections/git-strategy.yaml`의 team 프로파일에 `merge_method: rebase`가 기록되고, 파일 내 기존 주석이 보존되며(dirty-flag typed 경로), enum 밖 값(`fast-forward` 등) 저장 시도는 거부된다

### S2 — auto_apply 쓰기 거부 (REQ-WC14-001)
- **Given**: `learning.auto_apply`가 read-only 표시 키로 전환된 상태
- **When**: POST /save 요청이 `learning.auto_apply=true`를 포함
- **Then**: 적용 계층이 해당 키 쓰기를 거부하고, `harness.yaml`의 `auto_apply: false` 디스크 값은 불변이며, 콘솔에는 거버넌스 설명 라벨이 표시된다

### S3 — dormant 가드 회귀 (REQ-WC14-050)
- **Given**: denylist 가드 테스트가 도입된 상태
- **When**: 미래 변경이 `sunset.` 또는 `workflow.model_routing` prefix의 편집 FieldDef를 추가
- **Then**: 가드 테스트가 FAIL하여 회귀가 CI에서 차단된다

### S4 — mx read-only raw view (REQ-WC14-030/031)
- **Given**: GET / 응답
- **When**: mx raw view 그룹을 렌더
- **Then**: `danger_categories`/`test_paths` 내용이 read-only로 표시되고, mx 섹션에 대한 쓰기는 계속 RouteExcluded로 거부된다

## §B Plan-phase 검증 증거 (2026-07-10 실측 — verification-claim-integrity §1.1 surface 3)

각 항목: 실행 명령 + 관찰된 핵심 출력(발췌 verbatim). 전체 노출 판정의 baseline 귀속.

**E-1 learning 블록 reader** — `grep -rn "TierThresholds\|tier_thresholds\|RateLimit\|rate_limit\|MaxPerWeek\|CooldownHours\|AutoApply\|auto_apply" internal/ --include="*.go" | grep -vi "_test.go" | grep -i "learn\|harness\|lesson"`:
```
internal/cli/hook.go:1066:// readTierThresholds reads learning.tier_thresholds from
internal/cli/hook.go:1084:	if len(doc.Learning.TierThresholds) == 0 {
internal/cli/harness.go:178:	_, _ = fmt.Fprintf(out, " max per week: %d times\n", cfg.RateLimit.MaxPerWeek)
internal/cli/harness.go:179:	_, _ = fmt.Fprintf(out, " cooldown : %d hours\n", cfg.RateLimit.CooldownHours)
internal/cli/harness/execute.go:89:// @MX:NOTE: [AUTO] AutoApply: true는 in-memory PipelineConfig 값일 뿐 harness.yaml
internal/cli/harness/execute.go:90:// 디스크 값(auto_apply: false)을 변경하지 않는다 (spec.md §B.2 FROZEN 불변식, C1).
```
→ tier_thresholds = 행동적(hook tier 분류), rate_limit = 표시 전용, auto_apply = FROZEN false.

**E-2 rate limiter enforcement 상수** — `sed -n '1,80p' internal/harness/safety/rate_limit.go`:
```
const rateLimitMaxUpdates = 3
const rateLimitWindow = 7 * 24 * time.Hour
const rateLimitCooldown = 24 * time.Hour
func NewRateLimiter(statePath string) *RateLimiter {
```
→ enforcement는 컴파일 상수, config 값 미소비. **브리프 정정 1의 결정 증거**.

**E-3 merge_method** — `grep -rn "merge_method\|MergeMethod" internal/ --include="*.go" | grep -v "_test.go"` (발췌):
```
internal/config/types.go:125:	MergeMethod string `yaml:"merge_method"`
internal/config/validation.go:271:var validMergeMethods = map[string]bool{
internal/config/validation.go:290:func ValidMergeMethods() []string {
```
+ SPEC-MERGE-METHOD-CONFIG-001 spec.md 원문: "The consumer of this new field is the **sync agent prose** (`gh pr merge`), not Go runtime code" → 문서화된 agent-prose consumer 계약 존재 → 편집 노출 정당.

**E-4 team late-branch 4키 reader 부재** — `grep -rn "BranchCreation\.\|Automation\.AutoBranch\|Automation\.AutoPR\|\.PromptAlways" internal/ --include="*.go" | grep -v "_test.go" | grep -v "internal/config/"`:
```
(출력 없음 — 공집합)
```
→ Go 행동적 reader 0건. types.go:78 주석 "Forward-compat scaffold" 부합 → 비노출 판정.

**E-5 security sandbox 리스트 키 reader** — `grep -rn "network_allowlist\|NetworkAllowlist\|env_scrub_extra\|EnvScrubExtra" internal/ --include="*.go" | grep -v "_test.go"` (발췌):
```
internal/config/types.go:498:	NetworkAllowlist []string `yaml:"network_allowlist"`
internal/config/types.go:502:	EnvScrubExtra []string `yaml:"env_scrub_extra"`
internal/sandbox/docker.go:68:	allHosts := append(DefaultNetworkAllowlist, opts.NetworkAllowlist...)
internal/sandbox/env.go:29:// security.yaml sandbox.env_scrub_extra (additive).
```
→ 행동적 소비(sandbox 실행 계층) → raw read-only view 정당.

**E-6 permission scaffold 키 바인딩 부재** — `grep -n "pre_allowlist\|PreAllowlist\|session_rules\|SessionRules" internal/config/types.go`:
```
(출력 없음 — 공집합; StrictMode만 존재)
```
`grep -rn "Security.Permission\." internal/ --include="*.go" | grep -v "_test.go"`:
```
(출력 없음 — 공집합)
```
+ security.yaml 자체 주석: "실제 패턴은 stack.go PreAllowlist() 에 정의" → Go-unbound scaffold. **브리프 정정 2의 결정 증거**.

**E-7 mx reader** — `grep -rn "danger_categories\|DangerCategories\|test_paths\|TestPaths" internal/ --include="*.go" | grep -v "_test.go"` (발췌):
```
internal/mx/danger_category.go:47:// LoadDangerConfig reads .moai/config/sections/mx.yaml under projectRoot
internal/cli/mx_query.go:118:	fanInCounter := mx.NewTextualFanInCounterWithTestPaths(dangerCfg.TestPaths)
```
+ mx.yaml:282-286 `test_paths:` 리스트 실존 → 행동적 reader 존재, 값은 맵/리스트 → raw read-only view.

**E-8 observability 기노출** — `internal/settings/schema_sections.go` seam 필드 실측 (Read, L241-242 실측 시점):
```
s(SectionObservability, "observability", TypeText, "observability", "hook_metrics", "output_path"),
s(SectionObservability, "observability", TypeInt, "observability", "hook_metrics", "slow_hook_threshold_ms"),
```
→ 브리프 항목 5 검증 완료: **no-op** (회귀 핀 AC-040만 추가).

**E-9 dormant 확인** — `.moai/config/sections/sunset.yaml` 헤더 verbatim:
```
# DORMANT: this config is typed and loaded, but no runtime hot path enforces these
# conditions yet. The Go loader emits a once-per-session DORMANT notice.
```
+ `grep -rn "ModelUpgradeReview" internal/ --include="*.go" | grep -v "_test.go"` → `internal/cli/harness_validate.go:150` 리마인더 출력만(비강제) → 가드 대상.

**E-10 라우팅/가드 기반** — `internal/settings/sectionroute.go` 실측: `ExcludedSections()`에 sunset/tool-policy/mx 포함, `sectionRoutes` 맵 미등재 이름 zero-value RouteExcluded. `sectionroute_test.go:58 TestExcludedSectionsAllRejected` (len=12 핀) 실존.

**E-11 parity 테스트 표면** — `internal/settings/accessors.go:8 func AllFields()`; `internal/cli/schema_bridge_test.go`: TestI18nKeySetParity/TestI18nSegmentParity/TestBridgeFieldDefResolver/TestTUIRendersSchemaFieldSet 실존 (grep `^func Test`).

**E-12 i18n 구조** — `internal/web/assets/i18n.js`: `window.MOAI_I18N = { en:(L20) ko:(L368) ja:(L716) zh:(L1064) }` 4-locale, `"f.` 키 1032건 실측.

### 미검증 (Gaps)

- `learning.rate_limit` 값의 harness 파이프라인 내 간접 소비 가능성 — grep 범위는 internal/ 전체였으나 PipelineConfig 조립 전 경로 전수 추적은 미수행 → pre-flight C-3가 재확인 (반증 발견 시 blocker).
- SPEC-WEB-CONSOLE-013 산출물 내용 — 미작성 상태라 필드 네임스페이스 충돌 여부 검증 불가 → pre-flight C-1b.
- raw-only(편집 필드 0) 섹션의 렌더 파이프라인 지원 여부 — 미실측 → pre-flight C-4.

## §C AC Matrix (21 AC)

| AC | Sev | REQ | 검증 명령 (run-phase) | 기대 결과 |
|----|-----|-----|----------------------|-----------|
| AC-WC14-000 | [B] | 전체 | `go build ./... && go test ./internal/settings/... ./internal/web/... ./internal/cli/...` | exit 0 |
| AC-WC14-001a | [B] | 001 | 가드 테스트 denylist에 `learning.auto_apply` 정확명 포함 + `grep -n "func TestDormantConfigNeverEditable" internal/settings/*_test.go` ≥1건 후 `go test ./internal/settings/ -run TestDormantConfigNeverEditable -v` | grep 매치 + test PASS |
| AC-WC14-001b | [B] | 001 | `go test ./internal/settings/ -run TestApplySchemaEditsRejectsUnknownAndReadOnly -v` (learning.auto_apply 거부 케이스 확장 포함) + `grep -n "auto_apply" internal/settings/schema_sections.go`에서 ReadOnly 등록부 내 1건 이상, seam 편집 생성부 0건 | test PASS + grep 조건 충족 |
| AC-WC14-002 | [B] | 002 | `grep -c "tier_thresholds" internal/settings/schema_sections.go` ≥1 + `go test ./internal/settings/ -run TestRawBlockValues -v` | grep ≥1 + PASS |
| AC-WC14-003 | [B] | 003 | `grep -c "rate_limit" internal/settings/schema_sections.go` ≥1 (RawViewBlocks 등록) + `go test ./internal/settings/ -run TestRawBlockValues -v` | grep ≥1 + PASS |
| AC-WC14-010a | [B] | 010 | `grep -c "merge_method" internal/settings/schema_sections.go` ≥1 + AllFields에 `git_strategy.{manual,personal,team}.merge_method` 3건 존재 단언 테스트(func-grep 선행) | grep ≥1 + test PASS |
| AC-WC14-010b | [B] | 010 | `go test ./internal/settings/ -run TestApplySchemaEditsGitStrategyTyped -v` (merge_method 라운드트립 확장: 저장 후 git-strategy.yaml에 값 반영 + 주석 보존) | PASS |
| AC-WC14-011 | [B] | 011 | `grep -n "ValidMergeMethods\|IsValidMergeMethod" internal/settings/schema_sections.go` ≥1 + enum 밖 값 거부 테스트 PASS + `grep -c '"squash"' internal/settings/schema_sections.go` = 0 (리터럴 재선언 금지) | 모두 충족 |
| AC-WC14-012 | [B] | 012 | 가드 denylist에 4개 정확명(`git_strategy.team.branch_creation.prompt_always`/`.auto_enabled`/`git_strategy.team.automation.auto_branch`/`.auto_pr` — manual/personal 변형 포함 prefix 매칭 허용) + 가드 test PASS | PASS |
| AC-WC14-020 | [B] | 020 | `grep -c "network_allowlist" internal/settings/schema_sections.go` ≥1 + `grep -c "env_scrub_extra"` ≥1 + `go test ./internal/settings/ -run TestRawBlockValues -v` | grep 각 ≥1 + PASS |
| AC-WC14-021 | [B] | 021 | `grep -c "pre_allowlist\|session_rules" internal/settings/schema_sections.go` = 0 + 가드 denylist 항목 포함 + 가드 test PASS | 모두 충족 |
| AC-WC14-030a | [B] | 030 | `grep -c "danger_categories" internal/settings/schema_sections.go` ≥1 + `grep -c "test_paths"` ≥1 (RawViewBlocks) + TestRawBlockValues PASS | 충족 |
| AC-WC14-030b | [B] | 030 | `grep -n "func TestMXRawViewRendered" internal/web/*_test.go` ≥1건 후 `go test ./internal/web/ -run TestMXRawViewRendered -v` (GET / 응답에 mx raw view 컨테이너 존재) | grep 매치 + PASS |
| AC-WC14-031 | [B] | 031 | `go test ./internal/settings/ -run TestExcludedSectionsAllRejected -v` PASS + 가드 denylist `mx.` prefix 포함 | PASS |
| AC-WC14-040 | [B] | 040 | AllFields에 `observability.hook_metrics.output_path` + `observability.hook_metrics.slow_hook_threshold_ms` 존재 단언(가드 테스트의 allowlist 핀 케이스, func-grep 선행) | PASS |
| AC-WC14-050 | [B] | 050 | `grep -n "func TestDormantConfigNeverEditable" internal/settings/*_test.go` ≥1건 후 `go test ./internal/settings/ -run TestDormantConfigNeverEditable -v` — denylist: `sunset.`/`model_upgrade_review`/`workflow.model_routing`/`mx.`/`tool_policy`/`tool-policy` prefix + §B.6 정확명 전부 | grep 매치 + PASS |
| AC-WC14-051 | [B] | 051 | `go test ./internal/settings/ -run TestExcludedSectionsAllRejected -v` + sectionroute_test.go에 sunset/tool-policy/mx 명시 핀 존재 (`grep -n "sunset\|tool-policy\|\"mx\"" internal/settings/sectionroute_test.go` ≥3) | PASS + grep ≥3 |
| AC-WC14-060 | [B] | 060 | 본 SPEC 신규 i18n 키 목록(run-phase에서 확정)에 대해 4-locale 각 존재: per-locale 구간 grep (en/ko/ja/zh 오브젝트 경계 내 키별 ≥1) + `go test ./internal/web/ -run 'TestDataI18nKeysSubsetOfDictionary|TestI18nDictionaryEmbedded' -v` | 키별 4/4 + PASS |
| AC-WC14-061 | [B] | 061 | `go test ./internal/cli/ -run 'TestI18nKeySetParity|TestI18nSegmentParity|TestBridgeFieldDefResolver|TestTUIRendersSchemaFieldSet' -v` | exit 0 |
| AC-WC14-062 | [B] | 062 | `git diff --name-only $(git merge-base origin/main HEAD)..HEAD -- internal/statusline/` | 출력 0줄 |
| AC-WC14-063 | [N] | 전체 | `golangci-lint run internal/settings/... internal/web/... internal/cli/...` | clean (또는 pre-existing만) |

## §D Quality Gates / Definition of Done

1. [B] AC 20건 전부 PASS (AC-063은 [N]).
2. 가드 테스트(M1)가 **현행 코드 기준 즉시 GREEN**으로 도입됨 — RED 발생 시 blocker report (plan.md §D).
3. plan.md §C pre-flight 7항 수행 기록이 progress.md §E.2에 남음.
4. 브리프 정정 2건(B1/B2)의 반증 재탐색(C-3) 결과가 §E.2에 기록됨.
5. frontmatter 전이는 소유권 매트릭스 준수 (draft→in-progress: manager-develop / →completed: manager-docs).
