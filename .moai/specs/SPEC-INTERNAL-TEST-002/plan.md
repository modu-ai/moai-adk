---
id: SPEC-INTERNAL-TEST-002
version: "0.1.0"
status: completed
created: 2026-07-09
updated: 2026-07-09
---

# SPEC-INTERNAL-TEST-002 — plan

## §A Context

### §A.1 Epic 위치 (internal/ 독립 감사 Epic, 5/5 → 6/6 확장)

본 SPEC은 internal/ 독립 감사 Epic의 5번째 SPEC으로 편성된다. Epic scope SSOT는 본 §A.1.

| # | SPEC | 상태 | 본 SPEC과의 관계 |
|---|------|------|-----------------|
| 1 | SPEC-INTERNAL-TEST-001 | completed (PASS-WITH-DEBT, 2026-07-08) | `depends_on` 부모. 본 SPEC의 3개 부채 명명 원천 (`progress.md §E.4`) |
| 2 | SPEC-INTERNAL-SECURITY-001 | completed | 형제. 독립 감사 코호트 |
| 3 | SPEC-INTERNAL-PERF-001 | completed | 형제. 독립 감사 코호트 |
| 4 | SPEC-INTERNAL-ARCH-001 | draft (blocked on TEST-002 M1 + web-i18n SPEC per plan-audit iter-1 BLOCKING D1) | **본 SPEC M1 + 후속 web-i18n SPEC을 선행 조건으로 대기**. 본 SPEC M1은 necessary not sufficient (ARCH-001 AC-ARCH-001 `go test ./... exit 0` whole-repo가 잔여 2개 internal/web i18n FAIL로 미달성). 두 SPEC 모두 머지 후 re-entry |
| 5 | **SPEC-INTERNAL-TEST-002 (본 SPEC)** | **draft (저작 중)** | debt-owner for DEBT-TEST-001/002/003 |
| 6 | web-i18n SPEC (TEST-003 또는 특화 SPEC) | **미저작 (future SPEC)** | `internal/web` i18n 2개 FAIL(`TestDataI18nKeysSubsetOfDictionary`, `TestI18nKeySetParity`) 청소. ARCH-001 re-entry의 두 번째 선행 조건. 본 SPEC scope 밖 — 참조만 |

### §A.2 현재 git 상태 (저작 시점 HEAD)

- branch: `main`
- 최근 HEAD 커밋: `cb8580ecb chore(SPEC-MOAI-SKILL-DOCTRINE-FIX-001): record sync_commit_sha d0d953894`
- working tree: 18 modified + 5 untracked (IGGDA Path-B retire 잔류 — 본 SPEC scope 밖, C-4 보호 대상)
- 본 SPEC이 `depends_on`하는 TEST-001 sync: `80dea9684` (2026-07-08 03:37:36 +0900)
- 본 SPEC M2 already-resolved 근거: `794bb4f84` (2026-07-09 00:41:12 +0900, HEAD에 포함)

### §A.3 Tier 및 산출물 세트

- **Tier M (3 artifacts)**: spec.md + plan.md + acceptance.md (본 3개 + progress.md §E skeleton = 4개 파일)
- Tier L(5 artifacts)로 승격하지 않는 근거: 부채 (a) root-cause 재검증 결과 stale-golden(단일 문자 diff) 확인 → renderer 재설계 필요 없음 → design.md 불필요. 부채 (c) pipeline integration test는 표준 `t.TempDir()` 패턴으로 커버 가능 → research.md 불필요.
- plan-auditor PASS threshold: 0.80 (Tier M 기준)

### §A.4 plan-auditorKICKOFF Predicate ( Phase 0.95 )
- (a) intent clarity: PASS (task 위임 프롬프트가 3 부채 + Milestone 구조까지 명시)
- (b) Tier M: PASS (본 plan.md §A.3)
- (c) no dangerous keywords: PASS (database migration 아님 — internal/migration 패키지 NOT in scope)
- (d) plan-auditorKICKOFF verdict: **plan-auditor 심사 대기** (본 파일 작성 시점)

### §A.5 PRESERVE list (변경 금지)

- SPEC-INTERNAL-TEST-001 산출물 전부 (`.moai/specs/SPEC-INTERNAL-TEST-001/{spec,plan,acceptance,progress}.md`) — read-only evidence source
- SPEC-INTERNAL-ARCH-001 산출물 전부 — `status: draft` SPEC (blocked on TEST-002 M1 + web-i18n SPEC per plan-audit iter-1 BLOCKING D1; "paused"는 8-value status enum에 없는 비정형 용어), 본 SPEC이 수정 금지
- working tree의 IGGDA Path-B 잔류 파일 18개 + untracked 5개 전부 (C-4)
- `internal/statusline/renderer.go`, `internal/statusline/cache_hit_test.go` (dirty-tree 변경사항 — REQ-TEST-008은 HEAD 베이스라인만 사용)

## §B Known Issues (manager-develop 위임 시 주입)

- **B1 (Cross-platform build tags)**: 본 SPEC은 Go 코드 신규 추가 없음 (REQ-TEST-007/008은 코드 변경 없음, REQ-TEST-009는 `internal/constitution/*_test.go` 신규 추가만). syscall build tag 해당 없음.
- **B4 (Frontmatter Canonical Schema)**: 본 SPEC 산출물은 이미 12 canonical fields 준수 (manager-spec 저작 시 검증 완료).
- **B5 (CI 3-tier 인지)**: spec-lint / golangci-lint / Test 각각 별도 fail 가능. REQ-TEST-009 신규 테스트가 lint 통과해야 함.
- **B6 (spec-lint Heading 규약)**: 본 SPEC은 `### Out of Scope — <topic>` H3 sub-heading 6개로 `MissingExclusions` ERROR 회피 (spec.md 확인).
- **B9 (Git Commit 직접 수행)**: Hybrid Trunk 1-person OSS 정책 — manager-develop이 main 직진 commit + push 수행 (Tier M default).
- **B10 (Untouched Paths PRESERVE)**: §A.5 PRESERVE list 엄수. 특히 다른 Claude session(1d3c155b quiescent)과의 race 회피 위해 specific pathspec commit 의무 (C-6).
- **B11 (AskUserQuestion 금지)**: subagent boundary 준수. blocker 발생 시 structured report 반환.
- **B12 (CHANGELOG emission)**: 본 SPEC sync-phase 소관 (manager-docs). plan-phase 해당 없음.

## §C Pre-flight Check List (run-phase 착수 전)

```bash
# 1. branch + baseline
git branch --show-current                         # → main
git rev-parse HEAD                                # → 현재 HEAD (저작 시점 cb8580ecb 계승)

# 2. Cross-platform build (REGRESSment 없음 확인)
go build ./...                                    # → exit 0
GOOS=windows GOARCH=amd64 go build ./...          # → exit 0

# 3. lint baseline (NEW vs pre-existing 구분)
golangci-lint run --timeout=2m 2>&1 | tail -5     # → 0 issues (TEST-001 E5 baseline)

# 4. PRESERVE list 확인
ls .moai/specs/SPEC-INTERNAL-TEST-001/ .moai/specs/SPEC-INTERNAL-ARCH-001/  # → read-only

# 5. debt (a) 현재 상태 — 6 FAIL 확인
go test ./internal/cli/ -run 'TestDoctor_Current_|TestStatus_Current_' -count=1 2>&1 | tail -10
# → 6 FAIL 확인 (stale rc6 → rc8)

# 6. debt (b) 현재 상태 — PASS 확인 (이미 해결됨)
go test -run TestBuild_WritesContextUsageWithSessionID -count=3 ./internal/statusline/ 2>&1 | tail -3
# → ok (PASS) — 794bb4f84 already-resolved

# 7. debt (c) 현재 상태 — coverage 67.5%
go test -cover ./internal/constitution/ 2>&1 | tail -2
# → coverage: 67.5% of statements
```

## §D Constraints (manager-develop 위임 시 전달)

- **C-1 ~ C-7** (spec.md §C 그대로 주입)
- 특히 **C-4 (working-tree preservation)**: `git status --porcelain` 출력의 전부가 보호 대상. 본 SPEC commit은 specific pathspec만 (`git add .moai/specs/SPEC-INTERNAL-TEST-002/ internal/cli/testdata/ internal/constitution/*_test.go internal/constitution/coverage_test.go`)
- **C-6 (pathspec discipline)**: `git add -A` / `git add .` 절대 금지
- 금지 명령: `--no-verify`, `--amend`, force-push to main, `git stash` (working-tree 보호)
- 사용 의무: Conventional Commits (`feat(SPEC-INTERNAL-TEST-002): M{N} <subject>`), `🗿 MoAI` trailer

## §E Self-Verification Deliverables (manager-develop §E1-E7)

> Each E-item은 verification-claim-integrity 5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)로 보고. 본 §E는 run-phase manager-develop이 작성 — 본 plan-phase 문서에서는 placeholder.

- E1. AC Binary PASS/FAIL Matrix (AC-TEST-007/008/009)
- E2. Cross-Platform Build (`go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`)
- E3. Coverage (`go test -cover ./internal/cli/`, `./internal/statusline/`, `./internal/constitution/`)
- E4. Subagent Boundary (해당 없음 — 본 SPEC은 hook/harness 코드 무접촉)
- E5. Lint (`golangci-lint run --timeout=2m`)
- E6. Branch HEAD + Push state (commit SHAs + `git push origin main` result)
- E7. Blocker Report (있을 경우 — 예: REQ-TEST-009 테스트 작성 중 pipeline.go 실제 결함 발견 시)

## §F Milestones (priority-based, no time estimates)

### M1 — debt (a): internal/cli golden testdata 재생성 (ARCH-001 M0 prerequisite)

**Priority: P0 (ARCH-001 re-entry를 gate하는 critical path)**

- 범위: 6개 golden testdata 파일 갱신 (`internal/cli/testdata/{doctor,status}-{light,dark,nocolor}.golden`)
- 메커니즘: `UPDATE_GOLDEN=1 go test ./internal/cli/ -run 'TestDoctor_Current_|TestStatus_Current_' -count=1`
- 변경 파일 수: 6 (testdata 만)
- AC: AC-TEST-007 (acceptance.md 참조)
- Commit strategy: 단일 commit `test(SPEC-INTERNAL-TEST-002): M1 regenerate 6 cli golden testdata (rc6→rc8)`

### M2 — debt (b): statusline env-isolation verify-only (already-resolved)

**Priority: P2 (이미 해결됨, evidence 재생성만)**

- 범위: 코드 변경 없음. `TestBuild_WritesContextUsageWithSessionID -count=10` 연속 PASS 증거 수집 + acceptance.md에 증거 블록 기록
- 변경 파일 수: 0 (코드) + 1 (acceptance.md 증거 블록 업데이트 — 단 본 파일은 manager-docs가 아닌 manager-develop이 run-phase에 업데이트 가능: AC evidence column은 run-phase 소유)
- AC: AC-TEST-008 (acceptance.md 참조)
- Commit strategy: M3와 동일 commit 또는 별도 소형 commit `docs(SPEC-INTERNAL-TEST-002): M2 record statusline already-resolved evidence (794bb4f84)`

### M3 — debt (c): internal/constitution pipeline.go integration coverage 67.5%→85%

**Priority: P1 (coverage floor 상향 — non-trivial integration test 저작)**

- 범위: `internal/constitution/*_test.go` 신규 추가 (pipeline.go 8함수 + human_oversight.go 2함수 커버)
- 변경 파일 수: 1-2 신규 _test.go 파일 (예: `internal/constitution/pipeline_test.go`, 필요 시 `internal/constitution/human_oversight_test.go` 확장)
- AC: AC-TEST-009 (acceptance.md 참조)
- Commit strategy: 단일 commit `test(SPEC-INTERNAL-TEST-002): M3 add pipeline.go integration tests (67.5%→≥85%)`
- **Risk**: pipeline.go가 filesystem lock + source-file amend + registry update를 수행하므로 `t.TempDir()` 외 실제 fs 경로 침범 주의. 잠재적 production 결함 발견 시 blocker report (C-3).

### Milestone 순서

```
M1 (P0, ARCH-001 gate) → M2 (P2, verify-only) → M3 (P1, coverage)
```

M1이 critical path이므로 가장 먼저 ship. M2는 코드 변경 없으므로 M1과 동일 commit에 묶거나 M3 직전 소형 commit으로 분리. M3는 integration test 저작으로 가장 effort 집중.

## §G Anti-Patterns / Risks

- **AP-1 (M1 over-engineering)**: REQ-TEST-007은 `UPDATE_GOLDEN=1` 단일 명령으로 끝나야 한다. golden 파일을 수동 byte-edit하거나 renderer를 수정하려는 시도 금지 (C-1).
- **AP-2 (M2 false-debt-claim)**: REQ-TEST-008은 이미 해결된 debt의 evidence 재생성만 한다. "deterministic flake 재현 실패"를 이유로 새 fix를 시도하거나 already-resolved 상태를 의심해 code를 수정하려 하지 않는다 — root cause가 commit 794bb4f84로 확정되었으므로 (vci §1.1 surface 3).
- **AP-3 (M3 production drift)**: REQ-TEST-009 테스트 저작 중 pipeline.go의 실제 결함을 발견하더라도 silent fix 하지 않는다 (C-3). blocker report → 후속 SPEC 소관.
- **AP-4 (working-tree race)**: IGGDA Path-B 잔류 파일 18+5개가 본 SPEC commit에 섞여 들어가지 않도록 specific pathspec only (C-4, C-6). 특히 `internal/statusline/renderer.go` dirty 변경이 본 SPEC M2 commit에 딸려 들어가면 REQ-TEST-008의 "코드 변경 없음" 단언이 깨진다.
- **AP-5 (headline gate over-claim)**: 3개 per-debt AC가 전부 PASS해도 `go test ./internal/... exit 0`는 달성되지 않는다 (2× internal/web i18n 잔류). headline gate를 본 SPEC이 claim하지 않도록 acceptance.md §D.1을 per-debt gate로 한정 (spec.md §A "Headline gate 범위" 단락 참조).

## §H Cross-References

- `.moai/specs/SPEC-INTERNAL-TEST-001/progress.md §E.4` — 3개 부채 원본 명명
- `.moai/specs/SPEC-INTERNAL-ARCH-001/plan.md` — M0 gate (AC-ARCH-001: `go test ./... exit 0` whole-repo headline) 및 REQ-ARCH-002 update.go/hook.go split (본 SPEC M1은 필요 조건 necessary not sufficient; 후속 web-i18n SPEC TEST-003 완료가 추가 필요)
- `.moai/specs/SPEC-INTERNAL-PERF-001/spec.md` — internal/ 감사 코호트 형제
- CLAUDE.local.md §6 — Test Isolation / Coverage Targets
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — §A-E delegation template (Tier M 의무)
- `.claude/rules/moai/core/verification-claim-integrity.md §1.1 surface 3` — 부채/결함 단언 도구 검증 의무 (본 SPEC §A 재검증 근거)
