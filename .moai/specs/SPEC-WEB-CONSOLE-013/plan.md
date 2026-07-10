# SPEC-WEB-CONSOLE-013 — Implementation Plan

> Tier M · module `internal/web` (+ `internal/settings`) · status draft
> SSOT: spec.md (요구), acceptance.md (AC + 키별 증거 표). 본 문서는 실행 계획.

## §A Context

### §A.1 작업 위치 / 기준

- Project root: `/Users/goos/MoAI/moai-adk-go`, branch `main` (Hybrid Trunk Route A — Tier M main 직진).
- SPEC 산출물: `.moai/specs/SPEC-WEB-CONSOLE-013/{spec,plan,acceptance,progress}.md`.
- 병렬 세션 주의: `internal/statusline/renderer.go` + `cache_hit_test.go`가 타 세션 수정 중 (2026-07-10 git status) — 무접촉 (REQ-WC13-030).

### §A.2 선행 의존 (depends_on 게이트)

- `depends_on: [SPEC-WEB-CONSOLE-012]` — Track 1 cleanup이 `internal/settings/schema_sections.go` + `sectionroute.go`를 먼저 접촉한다. **2026-07-10 현재 `.moai/specs/SPEC-WEB-CONSOLE-012/` 디렉터리 미작성** — run-phase Depends_on Pre-flight Check가 `status: completed` 미충족으로 차단하는 것이 의도된 동작이다 (wait/override/abort 3-option은 orchestrator 소관).

### §A.3 파일 접촉 인벤토리 (예상; run-phase pre-flight에서 재확정)

| 파일 | 변경 종류 | 밀스톤 |
|------|----------|--------|
| `internal/settings/sectionroute.go` | handoff/cache RouteSeam 등재, ExcludedSections에서 cache 제거, SeamSections 확장 | M1 |
| `internal/settings/sectionwrite.go` | `sectionRootKeys` 2건 추가 (handoff / cacheStrategy) | M1 |
| `internal/settings/schema.go` | `SectionHandoff` / `SectionCache` SectionID + `AllSections()` 렌더 순서 | M2 |
| `internal/settings/schema_sections.go` | `handoffFields()` / `cacheFields()` (PersistSeam kind; select/bool) | M2 |
| `internal/settings/sectionvalues.go` | 새 섹션 현재값 로딩 | M2 |
| `internal/web/assets/i18n.js` | sec/필드/옵션 키 ×4 locale (M2: handoff+cache, M3: Model Policy) | M2/M3 |
| `internal/web/` Model Policy read-only 뷰 (신규 파일, board.go READ-ONLY 대시보드 선례) + templ | M3 |
| 테스트: `sectionroute_test.go`, `sectionwrite_test.go`, `schema_sections_test.go`, `internal/web/coverage_test.go` 계열, `i18n_test.go`, `handlers_test.go` | M1-M3 각 동반 | — |
| `internal/cli/schema_bridge_test.go` | **무변경 예상** — persist-kind 제외 술어가 신규 seam 필드를 자동 스코프 (§B-5) | M2 검증만 |
| `internal/template/templates/**` | **무변경 예상** (REQ-WC13-031) | M4 검증만 |

### §A.4 설계 결정 요약

1. **handoff/cache = RouteSeam.** 두 섹션 모두 typed READ struct는 있으나(`HandoffConfig` types.go:656, `CacheConfig` cache_config.go:33) `ConfigManager.Save()` 쓰기 경로가 없다 → 011 M2a 확립 패턴대로 yamlpatch seam이 유일한 쓰기 경로 (REQ-WC13-005). 주석·미노출 키(spec_ttl 등) 보존은 seam 불변식이 담보 (REQ-WC13-006).
2. **Model Policy = FieldDef 없는 READ-ONLY 전용 뷰.** SPEC 보드(board.go) 선례를 따라 schema/persist 파이프라인 외부에서 렌더 — 편집 불가 키에 FieldDef를 만들면 schema_bridge/persist 표면이 오염된다 (REQ-WC13-021).
3. **닫힌 집합 재사용.** session_ttl 검증은 `cache_config.go` `validSessionTTLs`를 export 또는 대칭 테스트 가드 미러로 소비 (REQ-WC13-013); handoff.mode는 {manual, auto} (handoff_inject.go 소비 값과 일치).
4. **필드 수 하드코딩 금지.** 011 v0.2.0 교훈 — 총 필드 수 assertion은 파생(derived) 방식만.

## §B Known Issues (사전 주입)

- **B-1 (선재 결함 후보 — 본 SPEC 무수정):** `internal/config/types.go:245` `PerformanceTier` validator `oneof=high medium low` vs `template.ValidPerformanceTiers()` {max, medium, low} + `moai init --model-policy max|medium|low`(init.go:94) 불일치. `performance_tier: "max"` 기록 시 config validation 경고 가능성. Model Policy 뷰는 raw 값 표시라 차단되지 않을 것으로 예상; 표시가 차단되면 blocker report (수정은 Out of Scope).
- **B-2 (cache caveat):** `InjectCacheControl` 비테스트 호출자 0건 — cacheStrategy 편집의 현재 유효 반경은 `moai doctor` 표시 경로. UI에 허위 효능 문구 금지(i18n desc는 중립 서술).
- **B-3 (동일 파일 레이스):** 012(Track 1)와 `schema_sections.go`/`sectionroute.go` 접촉이 겹친다. depends_on 게이트가 직렬화를 강제하나, run 진입 시 `git log --oneline -5 -- internal/settings/`로 012 착지 여부 실측.
- **B-4 (spec-lint 규약):** `## §3 Exclusions` 아래 `### Out of Scope — <topic>` H3 필수 (MissingExclusions) — spec.md 준수 확인됨.
- **B-5 (TUI 브리지):** `internal/cli/schema_bridge_test.go:33` 술어 `f.Persist.Kind == PersistSeam || PersistTypedSection`이 웹 전용 필드를 TUI 파리티에서 제외한다. **신규 handoff/cache 필드는 PersistSeam kind로 이 술어를 그대로 탄다 → TUI측 코드/테스트 무변경.** 단 run-phase에서 `go test ./internal/cli/ -run TestSchemaBridge` 실행으로 술어 커버를 실증 (REQ-WC13-016). (011 M2b 회귀 교훈: 스키마 필드 추가는 웹+TUI 양측 파리티 테스트를 반드시 실행-확인.)
- **B-6 (frontmatter 규약):** `created:`/`updated:`/`tags:` canonical — snake_case alias 금지.
- **B-7 (working tree 위생):** 무관 untracked/병렬 산출물 commit 금지 — `git add` specific path만.
- **B-8 (crossplatform):** 신규 Go 파일에 syscall 미사용 예상이나 `GOOS=windows` 빌드 검증 의무.

## §C Pre-flight Check List (run-phase 착수 전 의무)

```bash
# 1. depends_on 게이트: 012 완결 확인
test -f .moai/specs/SPEC-WEB-CONSOLE-012/spec.md && grep -m1 "^status:" .moai/specs/SPEC-WEB-CONSOLE-012/spec.md
# 기대: status: completed (미충족 시 wait/override/abort — orchestrator)

# 2. 병렬 세션 파일 상태
git status --porcelain internal/statusline/

# 3. 기준 브랜치/HEAD + 빌드 baseline
git branch --show-current && git rev-parse HEAD
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. 앵커 재검증 (content-token 기준; 라인 번호 드리프트 허용)
grep -n "RouteExcluded\|sectionRoutes" internal/settings/sectionroute.go | head -5
grep -n "handoffConfig" internal/hook/handoff_inject.go
grep -rn "InjectCacheControl" internal/ --include="*.go" | grep -v _test | grep -cv cache_control.go   # 기대 0 (배선 부채 유지 확인 — 달라지면 editable 판정 재검토)
grep -rn "RouteModelFor" internal/ cmd/ --include="*.go" | grep -v _test | grep -cv "internal/config/"  # 기대 0 (read-only 판정 전제)

# 5. i18n 키 규약 + generic fieldset 경로 열거
grep -n "sec\." internal/web/assets/i18n.js | head -10
grep -n "AllSections\|schemaform" internal/web/*.go | head -10

# 6. 012 접촉 흔적 (동일 파일 병합 상태)
git log --oneline -5 -- internal/settings/sectionroute.go internal/settings/schema_sections.go
```

## §D Constraints (DO NOT VIOLATE)

- `internal/statusline/**` 무접촉 (REQ-WC13-030).
- `internal/template/templates/**` 무접촉 예상; 접촉 필요 발견 시 Template-First + `make build` + §25 neutrality + scope delta 보고 (REQ-WC13-031).
- handoff.yaml / cache.yaml에 typed re-marshal 금지 — seam only (REQ-WC13-005).
- model_routing_profiles / performance_tier / model_routing / workflow_agents에 쓰기 경로·FieldDef persist 바인딩 신설 금지 (REQ-WC13-021/022/023).
- 로컬 dogfood `handoff.mode: auto` 값을 run-phase 테스트가 덮어쓰지 않도록 t.TempDir() 격리 (CLAUDE.local.md §6).
- `--no-verify` / force-push 금지; Conventional Commits `feat(SPEC-WEB-CONSOLE-013): M{N} <subject>` + `🗿 MoAI` trailer.
- 총 필드 수 하드코딩 assertion 금지 (파생 방식만 — 011 교훈).

## §E Self-Verification Deliverables

manager-develop 완료 보고는 vci 5-section 형식(Claim/Evidence/Baseline/Gaps/Residual-risk)으로 다음을 포함:

- E1: acceptance.md §D AC 전항목 binary PASS/FAIL 매트릭스 (검증 명령 + verbatim 출력).
- E2: cross-platform 빌드 결과 (darwin + GOOS=windows).
- E3: `go test -cover ./internal/settings/... ./internal/web/...` (≥85% per package).
- E4: subagent boundary grep (해당 없음 예상 — 웹/설정 패키지).
- E5: golangci-lint NEW vs baseline 구분.
- E6: 커밋 SHA 목록 + push 상태.
- E7: blocker report (B-1 validator 불일치가 표시를 차단하는 경우 등).

## §F Milestones (우선순위 순서 — 시간 예측 금지)

### M1 — 라우팅 기반 (REQ-WC13-001..006)
- sectionroute.go: handoff/cache RouteSeam + ExcludedSections cache 제거 + SeamSections 확장 + REQ-WC11-018 부분 supersede 주석 갱신.
- sectionwrite.go: sectionRootKeys 2건.
- 가드 테스트 갱신 (accept 신규 스코프 + reject 잔여 제외군 + cache.yaml 미노출 키 보존 characterization).
- 커밋: `feat(SPEC-WEB-CONSOLE-013): M1 routing foundation — handoff/cache seam routes`

### M2 — handoff + cache 노출 (REQ-WC13-010..017)
- SectionID 2건 + AllSections 순서 + handoffFields()/cacheFields() (PersistSeam) + sectionvalues 로딩.
- 검증 닫힌 집합: mode {manual,auto}, session_ttl {1h,5m,off} (+대칭 테스트).
- i18n ×4 (en/ko/ja/zh) + i18n_test 갱신.
- TUI 브리지 무변경 실증: `go test ./internal/cli/ -run TestSchemaBridge`.
- 커밋: `feat(SPEC-WEB-CONSOLE-013): M2 handoff+cache sections`

### M3 — Model Policy READ-ONLY 뷰 (REQ-WC13-020..026)
- 신규 read-only 뷰 (board.go 선례): performance_tier 표시(빈 값 → "(runtime default: medium)") + 3×12 라우팅 표 + 블록 부재 fallback 상태 + model_policy/performance_tier 구분 주석.
- legacy model_routing / workflow_agents 미렌더 가드 테스트.
- i18n ×4.
- 커밋: `feat(SPEC-WEB-CONSOLE-013): M3 model-policy read-only view`

### M4 — 검증 스윕 (REQ-WC13-030..032)
- 전체 빌드/테스트/린트 배치 + statusline/template 무접촉 diff 검증 + acceptance.md §D.1 증거 표 최종 채움.
- 커밋(필요 시): `test(SPEC-WEB-CONSOLE-013): M4 verification sweep`

의존: M1 → M2 → M3 → M4 직렬 (M2/M3는 동일 웹 표면 접촉으로 병렬 금지).

## §G Anti-Patterns

- 편집 불가 키(라우팅 셀)에 disabled input을 렌더해 "편집 가능해 보이는" UI — read-only 표는 form 요소 없이 렌더.
- validSessionTTLs 재선언 후 대칭 테스트 누락 — 닫힌 집합 드리프트 재생산.
- ExcludedSections에서 cache 제거 시 reject 테스트 케이스까지 삭제 — 잔여 제외군 가드 약화.
- Model Policy 뷰를 schema FieldDef로 구현 — persist 표면 오염 + schema_bridge 파급.
- 브리프의 stale 앵커(workflow_agents 14필드, MOAI_MODEL_POLICY env)를 재인용 — spec.md §1.1 정정이 SSOT.

## §H Cross-References

- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter/status 소유권.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 — 키별 증거 의무 (acceptance.md §D.1).
- SPEC-WEB-CONSOLE-011 design.md §A.3 — 라우팅 표 원 설계.
- `internal/settings/yamlpatch` — seam 불변식.
