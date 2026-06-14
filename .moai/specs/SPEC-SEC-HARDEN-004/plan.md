# Implementation Plan — SPEC-SEC-HARDEN-004

> Tier S (minimal). cycle_type=tdd, reproduction-first. 2 files + their tests only.
> Authoritative source: `.moai/reports/sync-audit/SPEC-SEC-HARDEN-003-2026-06-14.md` §Findings.

## A. Context (위치 + 분기 + SPEC 산출물 경로)

- **작업 위치**: `/Users/goos/MoAI/moai-adk-go` (main checkout, NO worktree).
- **현재 branch**: main (orchestrator가 GATE-2 후 결정; Hybrid Trunk 1-person OSS → main 직진 per git-workflow-doctrine Tier S).
- **SPEC 산출물**: `.moai/specs/SPEC-SEC-HARDEN-004/{spec.md, plan.md, acceptance.md, progress.md}`.
- **선행 인프라(PRESERVE)**: SEC-HARDEN-003이 추가한 leaf 가드 전부 보존 — `isSymlinkEntry`(update.go L2125), `restoreTargetContained`의 leaf check(L2162), `pathContainedIn`(file_changed.go L237), `resolveProjectRootFromInputOrEnv`(path_resolve.go), `runMXScan`의 lexical 가드(L160) + sidecar re-root(L184).
- **EXTEND 대상**: `restoreTargetContained`에 parent-chain 봉쇄 추가(F1); `runMXScan`에 EvalSymlinks scan-target 재검사 추가(F2).

### A.1 PRESERVE list (변경 절대 금지)

- `internal/cli/update.go`의 `restoreTargetContained` **leaf 가드**(L2153-2164: filepath.Rel `..` 거부 + leaf isSymlinkEntry) — parent-chain 검사는 그 위 additive.
- `internal/cli/update.go`의 `isSymlinkEntry`(L2125), `restoreMoaiConfig`/`restoreMoaiConfigLegacy`의 walk 구조·머지 로직.
- `internal/hook/file_changed.go`의 `pathContainedIn`(L237), `Handle`의 async goroutine 구조(L84-122), sidecar re-root(L184).
- `internal/hook/path_resolve.go` 전체(resolveProjectRootFromInputOrEnv).
- runtime-managed: `.moai/harness/*`, `.moai/state/*`, `.moai/cache/*`.
- 무관 SPEC 디렉토리, `.moai/research/*`, 무관 untracked 파일.

## B. Known Issues (Tier S 관련 카테고리만)

- **B1 Cross-platform**: `filepath.EvalSymlinks` / `filepath.Rel` / `os.Lstat`는 cross-platform이나 Windows에서 symlink 동작이 다를 수 있다. `GOOS=windows GOARCH=amd64 go build ./...` 통과 의무. symlink 재현 테스트는 Windows에서 `os.Symlink` 권한 이슈 가능 → 필요 시 `runtime.GOOS == "windows"`에서 `t.Skip` (단, 봉쇄 로직 자체는 cross-platform build 통과).
- **B3 Subagent boundary (C-HRA-008)**: `internal/hook`, `internal/cli` 변경 코드에 `AskUserQuestion`/`mcp__askuser` 호출 금지. 검증(변경 2파일 한정 — 전체 디렉터리 grep은 harness.go/agent_lint.go의 pre-existing godoc/string-literal 매치 5건 포함, 본 SPEC scope 아님; plan-audit D2): `grep -n 'AskUserQuestion\|mcp__askuser' internal/cli/update.go internal/hook/file_changed.go | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"` → 0 매치.
- **B4 Frontmatter**: spec.md frontmatter는 canonical 12-field + `tier:` + `era:` (snake_case alias 금지) — 이미 적용됨.
- **B6 spec-lint Heading**: spec.md는 `## F. Exclusions (What NOT to Build)` + `### F.1 Out of Scope` (h3 sub-section)을 가짐 → `MissingExclusions` 회피. 이미 적용됨.
- **B7 path resolution**: F2는 `resolveProjectRootFromInputOrEnv`만 사용(os.Getenv 인라인 금지). 이미 SEC-HARDEN-003에서 확립.
- **B9 Git commit + push (Hybrid Trunk)**: manager-develop이 본 SPEC scope commit + push 자체 수행(main 직진). Conventional Commits + `🗿 MoAI` trailer. `--no-verify`/`--amend`/force-push 금지. parallel session race 시 orchestrator가 push.
- **B10 Untouched paths PRESERVE**: §A.1 list 외 working tree 변경 금지.
- **B11 AskUserQuestion 금지**: blocker 발견 시 structured blocker report 반환(orchestrator가 AskUserQuestion).

## C. Pre-flight (착수 전 의무 검증)

```bash
# 1. branch + baseline
git branch --show-current && git rev-parse HEAD

# 2. cross-platform build 사전 확인
go build ./... && GOOS=windows GOARCH=amd64 go build ./internal/cli/... ./internal/hook/...

# 3. lint baseline (NEW vs pre-existing 구분 위해)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. 변경 전 coverage baseline 측정
go test -cover ./internal/cli/... ./internal/hook/... 2>&1 | grep -E 'coverage|ok'

# 5. SEC-HARDEN-003 leaf 가드 회귀 baseline (반드시 GREEN이어야 변경 시작)
go test -run 'TestRunMXScan_RejectsUncontainedFilePath|TestRunMXScan_RejectsUncontainedSidecarCWD|TestRunMXScan_AllowsInProjectPath' ./internal/hook/
go test -run 'TestRestoreMoaiConfig_LegacyBackup|TestRestoreMoaiConfigLegacy_WithMerge' ./internal/cli/
```

## D. Constraints (DO NOT VIOLATE)

- PRESERVE: §A.1 list.
- 금지 명령: `--no-verify`, `--amend`, force-push to main.
- 사용 의무: Conventional Commits (`feat(SPEC-SEC-HARDEN-004): M{N} <subject>`), `🗿 MoAI` trailer, `Authored-By-Agent: manager-develop` trailer(첫 run-phase commit이 draft→in-progress 전이 owner).
- C-HRA-008: B3 grep 0 매치.
- specid import 금지(internal/hook): F2는 specid를 import하지 않는다.
- 새 패키지/플래그/추상화 도입 금지(REQ-SEC4-008).

## E. Self-Verification Deliverables

manager-develop 완료 보고 시 자체 검증:

- **E1 AC Binary PASS/FAIL Matrix**: acceptance.md AC-SEC4-001..010 전부 PASS/FAIL + verification command + actual output.
- **E2 Cross-Platform Build**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./internal/cli/... ./internal/hook/...` exit 0.
- **E3 Coverage (≥85% threshold, no-regression)**: `go test -cover ./internal/cli/... ./internal/hook/...`.
- **E4 Subagent Boundary Grep (C-HRA-008, 변경 2파일 한정)**: `grep -n 'AskUserQuestion\|mcp__askuser' internal/cli/update.go internal/hook/file_changed.go | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"` → 0. (전체 디렉터리 grep의 pre-existing 5 매치(harness.go/agent_lint.go godoc/string-literal)는 baseline, 본 SPEC 도입 아님.)
- **E5 Lint Status**: `golangci-lint run --timeout=2m` (NEW vs baseline 구분).
- **E6 Branch HEAD + Push**: 새 commits SHA + push 결과.
- **E7 Blocker Report**: 있을 시 structured(AskUserQuestion 호출 금지).

## F. Milestones

> 2 milestones (F1 / F2). 각 milestone은 reproduction-first(RED → GREEN).
> SSE stall 임계(≥30 tasks) 미달이므로 Round 분할 없음.

### M1 — F1: `restoreTargetContained` parent-chain 봉쇄 (update.go, 양 walk 동시)

1. **RED**: `internal/cli/` 테스트에 재현 테스트 추가 — `TestRestoreMoaiConfig_RejectsSymlinkedParentDir`. configDir 안에 `linkdir → /outside`(t.TempDir 밖 별도 temp) symlinked dir을 사전 생성하고, 백업이 `linkdir/evil.yaml` relPath를 산출하도록 fixture 구성 → 현 코드에서 쓰기가 `/outside/evil.yaml`에 안착함을 assert(현재 FAIL=취약점 재현). 모던 walk(`restoreMoaiConfig`)와 레거시 walk(`restoreMoaiConfigLegacy`) 둘 다 커버하는 sub-test 또는 2 테스트.
2. **GREEN**: `restoreTargetContained`(L2141)에 parent-chain 봉쇄 추가 — `filepath.EvalSymlinks(filepath.Dir(targetPath))` 해소 후 resolved parent의 configDir(역시 EvalSymlinks 정규화) 봉쇄 재검사. not-exist parent는 lexical 봉쇄만으로 통과(아직 symlink 없음), 그 외 resolution error는 fail-closed(false). 공유 헬퍼이므로 양 walk 동시 봉쇄.
3. **REGRESSION**: SEC-HARDEN-003 leaf 가드 회귀 확인 — `TestRestoreMoaiConfig_LegacyBackup`, `TestRestoreMoaiConfigLegacy_WithMerge`, `TestRestoreMoaiConfig_SkipsNonYAML` 등 기존 테스트 GREEN 유지.
4. **commit**: `feat(SPEC-SEC-HARDEN-004): M1 restoreTargetContained parent-chain symlink containment (F1)`.

### M2 — F2: `runMXScan` scan-target EvalSymlinks 재검사 (file_changed.go)

1. **RED**: `internal/hook/file_changed_test.go`에 재현 테스트 추가 — `TestRunMXScan_RejectsSymlinkInRootEscapingTarget`. project root 안에 `root/innocent.go → /secret/secret.go`(root 밖 secret, MX-tag 포함) symlink 생성 → 현 코드에서 `ScanFile`이 secret의 MX-tag를 읽어 사이드카에 기록함을 assert(현재 FAIL=취약점 재현). `mx.Manager` 사이드카 또는 scan 결과로 검증.
2. **GREEN**: `runMXScan`(L130)의 `scanner.ScanFile(input.FilePath)`(L170) 전에 `filepath.EvalSymlinks(input.FilePath)` 해소 + `pathContainedIn(root, resolved)` 재검사 추가. 위반 시 `slog.Warn` + early return(스캔 없음). EvalSymlinks error(not-exist 포함)는 fail-closed early return. 기존 lexical 가드(L160)는 보존.
3. **REGRESSION**: SEC-HARDEN-003 leaf 가드 회귀 확인 — `TestRunMXScan_RejectsUncontainedFilePath`, `TestRunMXScan_RejectsUncontainedSidecarCWD`, `TestRunMXScan_AllowsInProjectPath`, `TestFileChanged_AsyncReturn_Under100ms`(빈 payload 계약), `TestFileChanged_SideEffectsCompleted` GREEN 유지.
4. **commit**: `feat(SPEC-SEC-HARDEN-004): M2 runMXScan symlink-in-root scan-target containment (F2)`.

### M3 — 최종 검증 + push

1. 전체 검증 batch(E1-E6): `go test ./internal/cli/... ./internal/hook/...`, cross-platform build, coverage, C-HRA-008 grep, lint.
2. progress.md §E.2 run-phase evidence 기록(manager-develop owner).
3. push(main 직진) — parallel session race 시 orchestrator.

## G. Anti-Patterns (이 SPEC 한정)

- F1을 modern-only로만 봉쇄 → 공유 헬퍼이므로 modern-only 분기는 불필요하고 오히려 오류. 공유 헬퍼 1곳 수정이 정답.
- F2를 무조건 symlink-skip(os.Lstat)으로 → scan 의미상 정상 in-root symlink 누락. resolve-recheck가 정답.
- leaf 가드 제거/대체 → SEC-HARDEN-003 회귀. additive layer로만.
- specid import(internal/hook) → 책임 분리 위반.
- vacuous AC(`[no tests to run]`) → 테스트 명 infix가 실제 Test func와 매칭되는지 확인(acceptance.md 참조).

## H. Cross-References

- `.moai/reports/sync-audit/SPEC-SEC-HARDEN-003-2026-06-14.md` §Findings(F1=라인 64-80, F2=라인 82-93) + §Recommendations(라인 123-132).
- `.moai/specs/SPEC-SEC-HARDEN-003/spec.md` §B 위협 모델(leaf 표면), C-F1/C-F2 봉쇄.
- `internal/cli/update.go` `restoreTargetContained`(L2141), `restoreMoaiConfig`(L1947), `restoreMoaiConfigLegacy`(L2054).
- `internal/hook/file_changed.go` `runMXScan`(L130), `pathContainedIn`(L237).
- `internal/hook/path_resolve.go` `resolveProjectRootFromInputOrEnv`.
- `internal/cli/CLAUDE.md` (절대경로/cross-platform 규약), `internal/hook/CLAUDE.md` (B7 path resolution).
