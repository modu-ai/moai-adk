# SPEC-SEC-HARDEN-005 — Progress

> 4-phase 진행 신호 보드. plan-phase는 manager-spec이 scaffold; `plan_complete_at` / `plan_status: audit-ready`는 plan-auditor PASS 후 orchestrator가 append.

## §A — Plan-phase Context

- **SPEC**: SPEC-SEC-HARDEN-005 — SEC-HARDEN §F residual containment (${IFS} shell-aware word-split + update env-trust allowlist)
- **Tier**: M (Medium) — plan-auditor PASS threshold 0.80
- **era**: V3R6
- **선행**: SEC-HARDEN-001/002/003/004 (전부 `completed`)
- **산출물 (5-file set)**: spec.md + plan.md + acceptance.md + design.md + progress.md
- **cycle_type (run-phase)**: tdd (quality.yaml development_mode: tdd)
- **branch 전략**: main 직진 (Hybrid Trunk 1-person OSS; --worktree/--branch 미사용)
- **수정 표면**:
  - §F.1 (PRIMARY): `internal/permission/stack.go` — `hasUnquotedShellSeparator`(L172) + `Matches` `:*` 브랜치(L100,127-136). NEW dep `mvdan.cc/sh/v3/syntax`.
  - §F.2 (PRIMARY): `internal/cli/deps.go` — `EnsureUpdate`(L250-309) env-read 블록. scheme+host allowlist.
  - §F.3 (OPTIONAL, 비요구): `restoreTargetContained`/`parentChainContained`/`runMXScan` godoc TOCTOU note. 코드 동작 변경 없음.

## §B — Plan-phase Self-Check

- [x] SPEC ID Pre-Write Self-Check: `decomposition: SPEC ✓ | SEC ✓ | HARDEN ✓ | 005 ✓ → PASS`
- [x] Frontmatter 12 canonical fields + era:V3R6 + tier:M, status: draft
- [x] `created:`/`updated:`/`tags:` (snake_case alias 없음)
- [x] GEARS format requirements (Ubiquitous / Event-driven; canonical 2-form per plan-auditor D4 정정)
- [x] Exclusions section §F with h3 sub-sections (MissingExclusions 회피)
- [x] design.md 포함 (신규 의존성 보안-게이트 통합 설계)
- [x] AC 13개 (재현 5 + 회귀 4 + fail-closed 1 + 의존성/범위 2 + 전역 NFR 1), 모든 grep AC `$` anchor + 명시 테스트 실행 (non-vacuous)
- [x] OPT-SEC5-001 (TOCTOU)는 OPTIONAL, AC 게이트 아님
- [x] anti-over-engineering: 새 패키지/플래그 금지(예외 mvdan.cc/sh + 검증 로직)
- [x] spec-lint clean (orchestrator 검증 — `✓ No findings`)
- [x] plan-auditor PASS-WITH-DEBT 0.86 ≥ 0.80 (Tier M); D1 BLOCKING + D2/D3/D4 전부 orchestrator-direct 정정

## §C — Milestone Tracker (run-phase, manager-develop)

| Milestone | 설명 | 상태 | commit SHA |
|-----------|------|------|------------|
| M1 | mvdan.cc/sh dep + §F.1 ${IFS} RED + legit baseline 고정 | done | 9648c7721 |
| M2 | §F.1 GREEN — hasIFSWordSplit 헬퍼 + Matches 배선 | done | bf5e2ee75 |
| M3 | §F.2 RED+GREEN — update env-trust allowlist | done | 8914af483 |
| M4 | §F.3 OPTIONAL godoc + 전체 검증 batch | done | (M-final, 본 progress.md update 포함 commit) |

## §D — Phase 0.5 SKIP Rationale (placeholder)

- plan-auditor verdict: PASS-WITH-DEBT 0.86 (Clarity 0.90 / Completeness 0.92 / Testability 0.74 / Traceability 1.00; MP-1..MP-4 PASS). 보고서: `.moai/reports/plan-audit/SPEC-SEC-HARDEN-005-2026-06-14.md`. 결함 D1 BLOCKING(C-HRA-008 grep idiom 불능) + D2 SHOULD-FIX(TestX$ trailing-`$` 경계 미고정) + D3/D4 MINOR 전부 orchestrator-direct 정정 (D1 canonical 필터 0 반환 + spec-lint clean 독립 검증).
- SKIP 적용 여부: **NOT skip-eligible (0.86 < 0.90)** → run-phase `/moai run` Phase 0.5 plan-auditor 재실행 필수.
- GATE-2 (plan→run HUMAN GATE): score 무관 — 사용자 명시 승인 필수(skip-eligible 0.90 autonomous bypass는 Phase 0.5 verdict 재실행에만 적용, GATE-2에는 미적용).
- **GATE-2 결과 (2026-06-14, run-phase 진입 세션)**: 사용자 명시 승인 → run-phase 진입 (AskUserQuestion "run-phase 진입 (권장)").
- **Phase 0.5 재감사 결과 (run-phase /moai run)**: plan-auditor 재실행 → **PASS 0.92** (0.86→0.92; Clarity 0.92 / Completeness 0.92 / Testability 0.88 / Traceability 1.00; MP-1..MP-4 PASS). D1/D2/D3/D4 정정 전부 landed 독립 확인 (canonical 필터 0 live-verified, 코드 앵커 전부 존재, 11 REQ↔13 AC 양방향 완전). R1 MINOR(비차단): AC-SEC5-007 grep `^\s`(GNU-specific)→`^[[:space:]]*` orchestrator-direct polish 적용. 0.92 ≥ 0.90 → 향후 Phase 0.5 skip-eligible.

## §D.2 — Phase 0.95 Mode Selection

- **Input parameters**: tier M / scope ~5-8 files (stack.go, deps.go, +tests, go.mod/go.sum, optional const + godoc) / domain count 2 (internal/permission, internal/cli) / file language mix 100% Go / concurrency benefit LOW (coding-heavy, dependent milestones M1→M2/M3) / Agent Teams prereqs NOT all met (standard harness).
- **Mode evaluation**: Mode 1 trivial=not selected (multi-file semantic) / Mode 2 background=not selected (writes code) / Mode 3 agent-team=not selected (2 domains <3, prereqs unmet) / Mode 4 parallel=not selected (coding-heavy per Anthropic caveat) / Mode 6 workflow=not selected (~8 files, coding-heavy, not mechanical-uniform) / **Mode 5 sub-agent=SELECTED**.
- **Decision: sub-agent (Mode 5)**
- **Justification**: Coding-heavy Go work with dependent milestones (M1 mvdan.cc/sh dep → M2 §F.1 GREEN; M1 dep → M3 §F.2). Anthropic coding-task parallelism caveat → sequential sub-agent is the safe default. Single manager-develop(cycle_type=tdd) spawn executed M1-M4 sequentially. GATE-2 already approved (above); Mode selection is strictly downstream. 실제 실행: manager-develop이 runtime 자율 L1 worktree(`worktree-agent-a5ef445596698bdf0`)에서 작업, orchestrator가 FF 통합·push (B9 예외a).

## §E — Audit-Ready Signals (4-phase, append-only)

> 각 phase 완료 후 해당 agent가 append. plan-phase signal은 plan-auditor PASS 후 orchestrator가 기록.

### §E.1 Plan-phase Audit-Ready Signal
- plan_complete_at: 2026-06-14
- plan_status: audit-ready
- plan_commit_sha: 328ff95e3

### §E.2 Run-phase Evidence

> manager-develop append (REQ-ARR-002). cycle_type=tdd (RED-GREEN-REFACTOR), Mode 5 (sequential sub-agent). L1 worktree(`worktree-agent-a5ef445596698bdf0`) 작업, 통합·push는 orchestrator(B9 예외a).

**AC PASS/FAIL Matrix (13 gating AC + 1 OPTIONAL)**

| AC | 유형 | Status | Verification Command | Actual Output |
|----|------|--------|----------------------|---------------|
| AC-SEC5-001 | 재현 | PASS | `go test -run 'TestMatches_IFSWordSplit_Reproduction$' ./internal/permission/` | `ok` (픽스 후 `${IFS}curl${IFS}evil`→false; 픽스 전 동일 테스트 FAIL 입증) |
| AC-SEC5-002 | 재현 | PASS | `go test -run 'TestMatches_IFSVariants$' ./internal/permission/` | `ok` (`${IFS}`/`$IFS/`/다중삽입 전부 false) |
| AC-SEC5-003 | 회귀 | PASS | `go test -run 'TestMatches_SeparatorVariants$' ./internal/permission/` | `ok` (separator DENY 스위트 green 유지) |
| AC-SEC5-004 | 회귀 | PASS | `go test -run 'TestMatches_PrefixChainBypass_Reproduction$' ./internal/permission/` | `ok` (SEC-HARDEN-001 M1 chain bypass DENY 유지) |
| AC-SEC5-005 | 회귀 | PASS | `go test -run 'TestMatches_IFSLegitNotRejected$' ./internal/permission/` | `ok` (9-sample legit set ALLOW 유지, `TestX$` trailing-`$` 포함) |
| AC-SEC5-006 | fail-closed | PASS | `go test -run 'TestMatches_MalformedShellFailClosed$' ./internal/permission/` | `ok` (malformed shell→false/DENY) |
| AC-SEC5-007 | 의존성 | PASS | `grep -E '^[[:space:]]*mvdan\.cc/sh/v3 ' go.mod` + `grep -c blacklist internal/permission/stack.go` | `mvdan.cc/sh/v3 v3.13.1` (1 match, direct require) + blacklist count `0` |
| AC-SEC5-008 | 재현 | PASS | `go test -run 'TestEnsureUpdate_RejectsNonHTTPSUpdateURL$' ./internal/cli/` | `ok` (non-https→fail-closed; 픽스 전 FAIL) |
| AC-SEC5-009 | 재현 | PASS | `go test -run 'TestEnsureUpdate_RejectsDisallowedHost$' ./internal/cli/` | `ok` (allowlist 외 host→fail-closed) |
| AC-SEC5-010 | 재현 | PASS | `go test -run 'TestEnsureUpdate_RejectsURLShapedReleasesDir$' ./internal/cli/` | `ok` (URL-shaped releases dir→fail-closed) |
| AC-SEC5-011 | 회귀 | PASS | `go test -run 'TestEnsureUpdate_DefaultPathNoRegression$' ./internal/cli/` | `ok` (env 미설정→api.github.com checker 정상 구성) |
| AC-SEC5-012 | 범위 | PASS | (1) `grep -nE '"https"' internal/cli/deps.go` (2) `grep -c 'EnvUpdateSource\|EnvUpdateURL\|EnvReleasesDir' internal/cli/deps.go` | (1) `allowedUpdateScheme = "https"` (deps.go:49) (2) `5` (3종 env만, 확장 없음) |
| AC-SEC5-013 | NFR | PASS | 4-command batch (build linux+win / full test / C-HRA-008 grep / lint) | linux=0, win=0, full test ok(all pkg), C-HRA-008 grep 0 매치, lint 0 issues |
| OPT-SEC5-001 | OPTIONAL | DONE (비게이트) | `grep -c 'TOCTOU\|check-vs-use' internal/cli/update.go internal/hook/file_changed.go` | update.go=4, file_changed.go=2 (godoc-only, 코드 동작 변경 0) |

**Cross-platform build (E2)**: `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (mvdan.cc/sh/v3/syntax는 pure-Go, NFR-SEC5-002).

**Coverage no-regression (E3, NFR-SEC5-003)**: `internal/permission` 88.0%→**89.5%** (개선), `internal/cli` 71.8% (baseline 동등). 신규 함수 전부 커버(hasIFSWordSplit/isTrailingDollarLiteral/commandHasUnquotedIFSOrSubst/wordPartsHaveUnquotedIFSOrSubst + validateUpdateURL/isLocalPath).

**C-HRA-008 (E4)**: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/permission/stack.go internal/cli/deps.go | grep -v '_test.go' | grep -v "^[^:]*:[0-9]*:[ \t]*//"` → 0 매치.

**Lint (E5)**: `golangci-lint run --timeout=2m ./internal/permission/... ./internal/cli/...` → 0 issues (NEW 0).

**Dependency (E7)**: `mvdan.cc/sh/v3 v3.13.1` 직접 require, `go mod tidy` clean, `go mod graph | grep mvdan` 최소(`go-quicktest/qt`만 transitive). `syntax` subpackage만 import(interp/expand 미사용).

**PRESERVE 불변 확인**: `hasUnquotedShellSeparator` 본체 + 모든 separator/redirect/unterminated-quote DENY 거동 불변(SEC-HARDEN-001 M1 + 002 M4 스위트 green). `restoreTargetContained`/`parentChainContained`/`runMXScan` 코드 동작 불변(godoc note만).

**Existing test 정정 (in-scope, 거동 변경 mandated by REQ-SEC5-007/008)**: `TestEnsureUpdate_CustomURL`(misc_coverage_test.go)가 취약 동작(`api.example.com` 임의 host 통과)을 encode → canonical `api.github.com`으로 정정. off-allowlist host 거부는 신규 `TestEnsureUpdate_RejectsDisallowedHost`가 커버.

### §E.3 Run-phase Audit-Ready Signal
- run_complete_at: 2026-06-14
- run_commit_sha: (M-final commit — 본 progress.md update 포함; M1 9648c7721 / M2 bf5e2ee75 / M3 8914af483 / M4 본 commit)
- run_status: implemented (13/13 AC PASS + OPTIONAL DONE)
- ac_pass_count: 13
- ac_fail_count: 0
- preserve_list_post_run_count: 3 (hasUnquotedShellSeparator 거동 + restoreTargetContained/parentChainContained/runMXScan 코드 동작 — 전부 불변)
- l44_pre_commit_fetch: orchestrator 책임 (L1 worktree 통합·push 시 pre-spawn fetch)
- l44_post_push_fetch: orchestrator 책임
- new_warnings_or_lints_introduced: 0
- cross_platform_build.linux: exit 0
- cross_platform_build.windows: exit 0
- total_run_phase_files: 7 (stack.go, stack_ifs_sec_harden_test.go, deps.go, deps_env_trust_test.go, misc_coverage_test.go, update.go, file_changed.go) + go.mod/go.sum + progress.md
- m1_to_mN_commit_strategy: M1(dep+RED) → M2(§F.1 GREEN) → M3(§F.2 RED+GREEN) → M4(OPTIONAL godoc + full verification + evidence). M별 분리 commit, L1 worktree 작업 후 orchestrator 통합·push.

### §E.4 Sync-phase Audit-Ready Signal
- sync_complete_at: 2026-06-15
- sync_commit_sha: (본 progress.md update 포함 commit — orchestrator backfill)
- sync_status: implemented
- frontmatter_transition: in-progress → implemented (spec.md)
- changelog_updated: true (### Security 섹션 1 항목)
- readme_updated: false (내부 권한/cli 봉쇄 — 사용자-facing 표면 변경 없음, README 갱신 불필요)
- docs_site_updated: false (docs-site는 사용자 가이드 — 내부 permission/cli 보안 봉쇄는 무관)
- sync_scope_files: 3 (CHANGELOG.md, spec.md frontmatter, progress.md §E.4)
- preserve_unrelated_uncommitted: 6 M files (settings-management workstream) + 7 untracked — NOT absorbed (explicit-path discipline)

### §E.5 Mx-phase Audit-Ready Signal
- mx_complete_at: 2026-06-15
- mx_commit_sha: b69f72dad
- mx_status: completed
- frontmatter_transition: implemented → completed (spec.md)
- sync_auditor_verdict: PASS-WITH-DEBT 0.90 (harmonic mean; Functionality 1.00 / Security 1.00 MUST-PASS 통과 / Craft 0.75 / Consistency 1.00). report: `.moai/reports/sync-audit/SPEC-SEC-HARDEN-005-2026-06-15.md`
- ac_reproduction: 13/13 독립 재확인 (progress.md §E.2 claim 미신뢰, 전부 재실행 GREEN)
- adversarial_probes: 다수 (§F.1 ${IFS} 변형 + §F.2 env-trust 변형), real bash + fake curl sentinel 계측, 0 genuine bypass — 두 위협모델 demonstrably closed
- deferred_should_fix: 1건 adjacent-class — nested `${x:-${IFS}}` default-value expansion AST walk 미재귀(stack.go). 비악용 입증(word-split→arg, command 합성 불가), design D.1.3 CallExpr-args-only scope 일치, command-chain 위협모델 닫힘. self-claim 위반 아님. SEC-HARDEN-003/004 §F adjacent-class 선례 → follow-up 후보(별도 SPEC). 사용자 결정: 이연.
- minor_findings: 2건 (design.md B.1 invariant 문구 광범위 doc-precision / AC-012(1) grep `internal/config` 경로 vacuous — const는 internal/cli, AC PASS 유지)
- pre_existing_diagnostics: 11건 (glm.go/team_spawn.go/update.go unused param 등) — SEC-HARDEN-005 commit 수 0, scope 밖 pre-existing baseline noise, 흡수/수정 안 함 (scope discipline)
- 4_phase_complete: plan(328ff95e3) → run(M1 9648c7721 / M2 bf5e2ee75 / M3 8914af483 / M4+orch a18bf798b) → sync(c5c4d0a36) → Mx(본 close commit)
