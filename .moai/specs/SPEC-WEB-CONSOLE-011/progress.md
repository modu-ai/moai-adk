# Progress — SPEC-WEB-CONSOLE-011

Lifecycle tracking for the 3-phase plan → run → sync flow. §E carries the audit-ready signals.

## §E.1 Plan-phase Audit-Ready Signal

- **Phase**: plan (artifacts authored)
- **Tier**: L (5-artifact: spec.md + plan.md + acceptance.md + design.md + research.md)
- **Status**: `draft`
- **plan_complete_at**: 2026-07-03
- **plan_status**: audit-ready (plan-auditor Phase 0.5 pending)
- **Version**: 0.2.0 — 사용자 4결정 확정 (2026-07-03, AskUserQuestion 직접 선택): M2 10섹션 전면 확장 + M3 쓰기 전면 지원으로 개정. v0.1.0의 "선택 확장 / staged write"는 supersede. → **0.2.1** — plan-audit iter-1 D1-D6 + D8d/D9d 정정 (REQ 총계 42→49 오산 정정 포함).
- **Artifacts created (v0.2.0 amended)**:
  - `spec.md` — 12-field canonical frontmatter (+ tier: L, era: V3R6), HISTORY (0.1.0 + 0.2.0 + 0.2.1), 49 GEARS requirements (REQ-WC11-001..005, 010..019, 020..029, 030..034, 040..046, 050..053, 060..062, 070..074; v0.2.0 개정 ID: 001/002/022/025/026/050/053; v0.2.1 정정 ID: 016/032/033), Exclusions with 7 `### Out of Scope —` H3 sub-headings.
  - `plan.md` — §A Context (+§A.5 PRESERVE), §B Known Issues B1..B13 (B10 i18n 인플레이션, B11 파생 카운트, B12 섹션 키 미실측, B13 frontmatter 견고성), §C Pre-flight 11-command (섹션 키 열거 + 미러 재확인 포함), §D Constraints, §E Self-Verification pointer, §F Milestones M1 → M2a → {M2b, M3} → M4/M5/M6 (의존 그래프 명시), §G Anti-Patterns AP-1..11, §H Cross-References.
  - `acceptance.md` — 52 AC (AC-WC11-001..005, 010..019, 020..029, 030a/030b..034, 040..046, 050..054, 060..063, 070..074), 7 Given-When-Then, 8 edge cases, Definition of Done (agent body byte-무변경 + 미러 3종 parity 포함).
  - `design.md` — §A yaml.Node patch seam (+upsert 확장, §A.3 10섹션 라우팅 표, §A.4 정규화 리스크), §B effort opaque-node 결정 (Option A — 전면 쓰기 하 유지 확인), §C.1 frontmatter patch layer + template-mirror policy (live-only + 지속 경고), §C.2 workflow_agents typed 표면, §D 검증 배치, §E 보드 데이터 흐름, §F segment SSOT 테스트, §G risks.
  - `research.md` — §A-§H survey (wf_d19d522a-d39) + §I 결정 변경 기록 (사용자 직접 선택, AskUserQuestion, 2026-07-03) + plan-phase 추가 실측 (10섹션/미러/workflow_agents grep 0/미지명 섹션 4종).
- **SPEC ID self-check**: `decomposition: SPEC ✓ | WEB ✓ | CONSOLE ✓ | 011 ✓ → PASS` (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Scope SSOT**: 사용자 승인 4결정 (2026-07-03) — spec.md §1.1.
- **Plan-auditor gate**: iter-1 **PASS-WITH-DEBT 0.85** (6 SHOULD-FIX + 4 MINOR, 전부 mechanical) → D1-D6 + D8d/D9d 정정 완료 (v0.2.1); D10d(house-convention orphan AC)는 noticed-only 잔존.
- **Implementation Kickoff Approval**: pending (plan-to-implement human gate — run-phase 진입 전 필수).

## §E.2 Run-phase Evidence

> **범위 주의**: 본 run-phase 위임은 **M1 + M2a만** 커버한다 (orchestrator 위임 명시). M2b/M3/M4/M5/M6은 후속 위임 소관 — 아래 evidence는 M1/M2a 바인딩 AC(AC-WC11-001..005, 017)에 국한된다. 작업 트리 = L1 runtime worktree (`worktree-agent-a20044e6263f21afb` 브랜치, base 2510c2775 = plan-phase 커밋; main과의 차분은 무관 SPEC-HANDOFF-CTXGUIDE-001 plan 커밋 1개뿐).

### Pre-flight 결과 (plan.md §C subset — 1, 3, 4, 5, 7, 8)

| # | 항목 | 명령 | 실측 결과 |
|---|------|------|----------|
| C-1 | baseline | `git branch --show-current && git rev-parse HEAD` | worktree 브랜치 base `2510c2775` (plan-phase 커밋); main checkout HEAD는 `b303d9916` (병렬 세션 미push 커밋 — 무접촉) |
| C-3 | cross-platform build | `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` | `BUILD_OK` + `WIN_BUILD_OK` (worktree 기준 재측정 포함) |
| C-4 | lint baseline | `golangci-lint run --timeout=3m` | `0 issues.` (baseline 0) |
| C-5 | 앵커 재검증 | content-token grep | `REQ-WC-012` @ server.go:10, `REQ-WC3-007` @ projectconfig.go:159(드리프트 +1행), `isValidProfileName` @ profile.go:118/126, `REQ-WC-009` @ app.go — 전부 실존 |
| C-7 | 10섹션 키 열거 | `cat .moai/config/sections/{8섹션}.yaml` | 8파일 전부 실존. harness.yaml은 최상위 키 2개(`harness:` + `learning:`) — sectionRootKeys에 반영. db.yaml 8키 = system 5(enabled, dir, auto_sync, migration_patterns, engine) + interview 3(orm, multi_tenant, migration_tool) |
| C-8 | no-Save-path 전제 기계 검증 | `internal/config/manager.go` `Save()` 실독 (L166-228) | **전제 성립**: Save()는 정확히 6파일만 기록 — user/language/quality/git-convention/git-strategy(dirty-flag)/llm. 8개 seam 섹션(workflow, harness, ralph, research, feedback, observability, security, db)은 Save 경로 부재 (typed 읽기 struct는 일부 존재하나 쓰기 경로 없음 — B12 해소, 불일치 0건) |

### M1 — Foundation (커밋 81f8ee11e)

- **Scope contract supersede**: server.go 패키지 doc(구 REQ-WC-012) + projectconfig.go @MX:REASON(구 REQ-WC3-007)을 10섹션 계약으로 재작성. `grep -rn "SPEC-WEB-CONSOLE-011" internal/web/server.go internal/web/projectconfig.go` = **2 매치** (AC-WC11-001 충족).
- **라우팅 SSOT 신설**: `internal/settings/sectionroute.go` — RouteForSection (typed 6 / seam 8 / statusline / 미등재 기본 거부 zero-value) + SeamSections() + ExcludedSections() (12종).
- **Guard test 신설**: `internal/web/scope_contract_test.go` — TestScopeContractTenSections + TestScopeContractExclusions (제외군 12종 + 임의 미등재 4종 거부).
- **yamlpatch seam (TDD RED→GREEN)**: `internal/settings/yamlpatch/` — KeyEdit/PatchFile (스칼라 교체 + 명시 경로 upsert + Style/Tag 보존 + indent 검출(4 기본/2 db형) + temp+rename 원자 기록, 삭제 미지원). **RED 관측**: stub 상태에서 6개 테스트 FAIL (`yamlpatch: not implemented`) → 구현 → GREEN. 실제 workflow.yaml fixture round-trip: 1-line diff + 주석/`team.patterns`/effort 전량 보존 + upsert additive-only.
- frontmatter 전환: spec/plan/acceptance `status: draft → in-progress` (M1 커밋 동승).

### M2a — 8섹션 seam persistence 인프라

- **per-section 쓰기 라우팅**: `internal/settings/sectionwrite.go` — WriteSectionViaSeam (RouteSeam 아닌 섹션 전부 오류 거부 + 파일 무접촉, sectionRootKeys로 섹션 파일 밖 최상위 키 주입 차단, db는 validateDBEdit로 REQ-WC11-019 3/5 분리 강제).
- **golden fixture ×8**: `internal/settings/testdata/sections/*.yaml` (실파일 사본 8개) + TestYAMLPatchGoldenSections — 섹션별 스칼라 1개 편집 후 (i) 1-line diff (ii) 주석 전량 보존 (iii) 키 순서 보존 검증. `go test -run TestYAMLPatchGolden ./...` **8/8 PASS** (AC-WC11-017).
- **yaml.v3 공백-only 정규화 실측 + golden 고정 (design.md §A.4 사전 승인 경로)**: 재인코딩이 매핑 항목-사이 빈 줄을 제거함을 관측 — 8섹션 중 항목-사이 빈 줄을 가진 **security.yaml / db.yaml 2개에서만** 발생. 주석/키/값/순서는 전량 보존(공백-only) → blocker 아님. golden 테스트에 `blankNormalized` 플래그로 범위 고정: 빈 줄 제거 외 라인 변동 0 + 라인 추가 0 + 비-빈 줄 1-line diff. 나머지 6섹션(workflow, harness, ralph, research, feedback, observability)은 strict byte-level 1-line diff 유지.
- **db 3/5 분리 (data-level)**: DBEditableKeys()={orm, multi_tenant, migration_tool} / DBSystemKeys()={enabled, dir, auto_sync, migration_patterns, engine} — 실측 db.yaml 헤더 주석("5 system-fixed + 3 interview-input")과 정합, 드리프트 가드 테스트(TestWriteSectionViaSeamKeySplitConstantsMatchFixture) 포함. system 키/미지명 키/중첩 경로 write 시도 → 오류 + 파일 불변 (AC-WC11-019 데이터 계층 준비).
- harness `learning:` 최상위 키 seam 기록 검증 (TestYAMLPatchGoldenHarnessLearningRoot).
- 커버리지 보강: yamlpatch 오류 경로 직접 테스트 (atomicWrite stat/read-only-dir, encode unknown-kind, 매핑/시퀀스 대상 거부) → 77.9% → **86.3%**.

### 발견 사항 / 특이 기록

1. **yaml.v3 blank-line 정규화** (위 M2a 항 — design.md §A.4 예상 리스크의 실측 확정, 허용 범위 내 golden 고정).
2. **선재 실패 (내 변경 무관)**: `go test ./...`에서 `internal/cli` `TestRunHookEvent_ReadInputError` FAIL (nil pointer panic @ coverage_test.go:77). **base 시점 선재** — 본 worktree에서 internal/cli는 무변경(`git status --porcelain internal/cli/` = empty)이고 M1/M2a 커밋은 해당 패키지 무접촉. main checkout에는 병렬 세션(SPEC-CLI-SUBPKG-SPLIT-001)의 internal/cli/coverage_test.go 미커밋 수정이 존재 — 해당 세션 소관. PRESERVE 제약상 무접촉 유지.
3. **선재 gofmt 드리프트 (내 변경 무관)**: `gofmt -l`이 internal/web/handlers.go + internal/web/projectnested_error_test.go 지적 — 양쪽 모두 git 무변경 base 파일. 내 신규/수정 파일은 전부 gofmt clean.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-04
run_commit_sha: "81f8ee11e (M1) + <M2a — 본 progress.md 동승 커밋, 커밋 후 백필 불요: 커밋 목록은 §E.2/E6 보고 기준>"
run_status: in-progress — M1 + M2a 완료 (본 위임 범위 전체); M2b/M3/M4/M5/M6 후속 위임 대기
ac_pass_count: 5   # AC-WC11-001, 002, 003, 004(인프라 계층), 017 — M1/M2a 바인딩 AC
ac_fail_count: 0
ac_partial_notes: "AC-WC11-004/005는 M1 인프라 계층에서 충족(seam 라우팅 + allowlist grep); 웹 핸들러 경유 행동 바인딩은 M2b/M3에서 완결"
preserve_list_post_run_count: 0   # PRESERVE 위반 0 — statusline/app.go@MX:NOTE/병렬 세션 산출물/agent 파일 전부 무접촉
l44_pre_commit_fetch: "n/a — L1 runtime worktree 격리 실행 (전용 브랜치, push 금지 지시); landing은 orchestrator gate 검증 후 수행"
l44_post_push_fetch: "n/a — push 미수행 (지시 사항)"
new_warnings_or_lints_introduced: 0   # golangci-lint baseline 0 → 이후에도 "0 issues."
cross_platform_build:
  darwin_arm64: PASS (go build ./... → BUILD_OK)
  windows_amd64: PASS (GOOS=windows GOARCH=amd64 go build ./... → WIN_BUILD_OK)
  go_vet: PASS (VET_OK)
coverage:
  internal_settings: "96.1%"
  internal_settings_yamlpatch: "86.3%"
  internal_web: "73.0% — 선재 baseline (본 위임의 web 변경은 주석 + 테스트만, 신규 실행문 0)"
total_run_phase_files: 11   # M1: server.go, projectconfig.go, scope_contract_test.go, sectionroute{,_test}.go, yamlpatch{,_test}.go, testdata/workflow.yaml + M2a: sectionwrite{,_test}.go, testdata/sections/×8, yamlpatch_test.go(보강) — 신규 12 + 수정 3 (SPEC frontmatter 3종 별도)
m1_to_mN_commit_strategy: "마일스톤별 path-limited 커밋 (M1: 81f8ee11e / M2a: 후속 커밋), 전용 worktree 브랜치, push는 orchestrator 소관"
known_preexisting_failures: "internal/cli TestRunHookEvent_ReadInputError (base 선재, 병렬 세션 SPEC-CLI-SUBPKG-SPLIT-001 도메인, 무접촉); gofmt 드리프트 2파일 (base 선재, 무접촉)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
