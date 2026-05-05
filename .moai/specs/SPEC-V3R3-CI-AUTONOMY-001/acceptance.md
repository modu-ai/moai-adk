---
id: SPEC-V3R3-CI-AUTONOMY-001
version: "0.1.0"
status: draft
created_at: 2026-05-05
updated_at: 2026-05-05
author: manager-spec
priority: P0
labels: [ci-cd, automation, acceptance, v3r3]
issue_number: null
breaking: false
---

# SPEC-V3R3-CI-AUTONOMY-001: Acceptance Criteria

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-05 | manager-spec | Initial. Per-Tier acceptance scenarios + DoD checklist + 5-PR sweep replay + current session replay. |

---

## 1. Acceptance Scenario Index

| ID | Tier | Theme |
|----|------|-------|
| AC-CIAUT-001 | T1 | Pre-push hook blocks lint failure (replay PR #739 P2) |
| AC-CIAUT-002 | T1 | `make ci-local` 16-language detection |
| AC-CIAUT-003 | T1 | `--no-verify` bypass logging |
| AC-CIAUT-004 | T2 | CI watch loop auto-invocation post `/moai sync` |
| AC-CIAUT-005 | T2 | Required vs auxiliary discrimination |
| AC-CIAUT-006 | T3 | Mechanical fix auto-resolution (replay PR #739 errcheck) |
| AC-CIAUT-007 | T3 | Semantic fail immediate escalation (replay PR #747 ETXTBSY) |
| AC-CIAUT-008 | T3 | Iteration cap 3 + mandatory escalation |
| AC-CIAUT-009 | T4 | Auxiliary workflow non-blocking |
| AC-CIAUT-010 | T4 | Release Drafter stale cleanup |
| AC-CIAUT-011 | T5 | Branch protection blocks force-push to main |
| AC-CIAUT-012 | T5 | No required check → no merge |
| AC-CIAUT-013 | T5 | No release/tag automation |
| AC-CIAUT-014 | T6 | Worktree state divergence detection |
| AC-CIAUT-015 | T6 | Empty worktreePath suspect handling |
| AC-CIAUT-016 | T7 | i18n validator blocks PR #783 mockReleaseData regression |
| AC-CIAUT-017 | T7 | i18n:translatable magic comment exempt |
| AC-CIAUT-018 | T8 | BODP recommends "main에서 분기" (current session replay) |
| AC-CIAUT-019 | T8 | BODP recommends stacked PR via `/moai plan --branch` (signal a positive) |
| AC-CIAUT-019b | T8 | BODP via `moai worktree new` CLI (default origin/main) |
| AC-CIAUT-020 | Cross | 5-PR sweep replay (manual validation) |
| AC-CIAUT-021 | T5 | required_status_checks SSoT in `.github/required-checks.yml` (REQ-027) |
| AC-CIAUT-022 | T5 | `gh` auth-failure graceful exit with manual command (REQ-028) |
| AC-CIAUT-023 | T7 | i18n validator 30s wall-clock budget (REQ-040) |
| AC-CIAUT-024 | T8 | `moai status` off-protocol branch reminder (REQ-050) |
| AC-CIAUT-025 | T8 | CLAUDE.local.md §18.12 BODP subsection added (REQ-051) |

---

## 2. T1 — Pre-Push Hook + ci-local

### AC-CIAUT-001: Pre-push hook blocks lint failure

**Given**: 사용자 워크스페이스에 `internal/cache/cache.go`가 존재하고, `golangci-lint errcheck` 위반 (PR #739 시나리오)이 있음. Pre-push hook이 `.git/hooks/pre-push`에 설치됨.

**When**: 사용자가 `git push origin feat/SPEC-XXX` 실행.

**Then**:
1. Pre-push hook이 `make ci-local` 호출
2. `golangci-lint run`이 errcheck 위반을 보고
3. Push가 non-zero exit로 차단됨
4. stderr에 remediation hint ("Run `make fmt && make lint && make test`") 출력
5. 원격 브랜치는 변경되지 않음 (`git ls-remote origin <branch>` 검증)

**Failure Mode**: lint가 통과하지만 `go test`가 실패하는 경우 — hook이 동일하게 차단해야 함.

**Cross-Platform**: macOS (POSIX sh), Linux (POSIX sh), Windows git-bash 환경에서 동일 동작.

### AC-CIAUT-002: `make ci-local` 16-language detection

**Given**: 비-Go 프로젝트 (예: Python `pyproject.toml`만 존재).

**When**: `make ci-local` 실행.

**Then**: `scripts/ci-mirror/run.sh`가 language marker를 감지하여 `lib/python.sh` (`ruff`, `pytest`)로 dispatch. Go 도구는 호출되지 않음. `lib/<lang>.sh`가 존재하지 않는 언어는 silent skip.

**Idempotency**: 동일 프로젝트에서 2회 실행 시 동일 결과 (캐시 영향 없음).

### AC-CIAUT-003: `--no-verify` bypass logging

**Given**: Pre-push hook 설치된 상태.

**When**: `git push --no-verify origin feature/x` 실행.

**Then**:
1. Push는 정상 진행
2. `.moai/logs/prepush-bypass.log`에 timestamp + user + branch 기록
3. 다음 `moai status` 시 bypass 횟수 friendly reminder

---

## 3. T2 — CI Watch Loop

### AC-CIAUT-004: CI watch auto-invocation post `/moai sync`

**Given**: 사용자가 `/moai sync SPEC-XXX` 실행하여 PR #N 생성됨.

**When**: `/moai sync` 마지막 단계 (push + PR create) 완료.

**Then**:
1. 30초 이내 watch loop 진입 (`scripts/ci-watch/run.sh` invocation 또는 skill activation)
2. `gh pr checks <N> --watch --json` polling 시작
3. 첫 status update가 30-60초 이내 자연어로 보고

**Backward Compatibility**: T2 미적용 PR (이전 세션 PR)에 대해 `/moai pr watch <N>` 수동 호출 가능.

### AC-CIAUT-005: Required vs auxiliary discrimination

**Given**: PR #N의 checks가 다음 상태:
- `Lint`: success
- `Test (ubuntu-latest)`: success
- `Test (windows-latest)`: failure
- `claude-code-review`: failure (org quota)
- `docs-i18n-check`: failure (advisory)

**When**: Watch loop이 모든 check 완료를 감지.

**Then**:
1. Required failure (`Test (windows-latest)`)이 T3 auto-fix 트리거를 발화
2. Auxiliary failure (`claude-code-review`, `docs-i18n-check`)는 advisory로 표시되지만 ready-to-merge 상태에 영향 없음
3. 사용자 보고에 "1 required failure, 2 advisory" 명확히 분리

---

## 4. T3 — Auto-Fix Loop

### AC-CIAUT-006: Mechanical fix auto-resolution

**Given**: PR #739 replay — `internal/cache/cache.go`에 errcheck 위반 (`os.Setenv` return value 미체크).

**When**: T2 watch가 lint failure 감지 → T3 진입.

**Then**:
1. `expert-debug` agent가 failed log + diff 분석
2. mechanical 분류 → patch 제안 (`if err := os.Setenv(...); err != nil { return err }`)
3. AskUserQuestion으로 patch apply 동의 (OQ2 결정 따라 trivial은 silent 또는 confirm)
4. Apply → 새 commit + push
5. 1 iteration 내 CI green 달성
6. `.moai/reports/ci-autofix/PR-739-2026-05-05.md`에 기록

**Failure Mode**: patch가 새로운 lint violation을 도입하면 iteration 2 진입.

### AC-CIAUT-007: Semantic fail immediate escalation

**Given**: PR #747 replay — `TestLauncher_Launch_HappyPath`에서 ETXTBSY race condition (Linux + race detector).

**When**: T2 watch가 test failure 감지 → T3 진입.

**Then**:
1. Classifier가 "data race" pattern 감지
2. semantic 분류 → patch 시도 SKIP
3. 즉시 AskUserQuestion 에스컬레이션 with diagnosis report ("ETXTBSY race in TestLauncher_Launch_HappyPath, suggested mitigations: ...")
4. 사용자 결정 대기 (continue manually / revise SPEC / abandon PR)

### AC-CIAUT-008: Iteration cap 3 + mandatory escalation

**Given**: 어떤 mechanical failure가 3 iteration 모두 실패 (각 iteration이 다른 lint 위반 도입).

**When**: iteration 3 실패.

**Then**:
1. Loop halt
2. AskUserQuestion으로 mandatory escalation (3 options + Other)
3. `.moai/reports/ci-autofix/<PR>-<DATE>.md`에 모든 iteration 기록
4. 사용자 응답 없으면 무한 대기 (silent timeout 금지 — 사용자가 진행해야 함)

---

## 5. T4 — Auxiliary Workflow Hygiene

### AC-CIAUT-009: Auxiliary workflow non-blocking

**Given**: T4 적용 후 PR이 docs-i18n-check failure 포함 (모든 required는 success).

**When**: PR ready-to-merge 평가.

**Then**:
1. GitHub UI가 "All checks have passed" 또는 동등한 ready 상태 표시
2. Auto-merge (T5) 활성화 시 자동 머지됨
3. Advisory PR comment에 docs-i18n-check 결과 informational로 표시

### AC-CIAUT-010: Release Drafter stale cleanup

**Given**: 30일+ stale Release Drafter draft 80개 존재.

**When**: 주간 cron `release-drafter-cleanup.yml` 실행.

**Then**:
1. 30일 미만 active draft는 보존
2. 30일+ + 연관 release 브랜치 없음 → 자동 close
3. Cleanup report가 `.github/release-drafter-cleanup-log.md`에 append

---

## 6. T5 — Branch Protection + Auto-Merge

### AC-CIAUT-011: Branch protection blocks force-push to main

**Given**: T5 적용 (`gh api -X PUT .../branches/main/protection`).

**When**: 사용자가 `git push --force origin main` 시도.

**Then**:
1. GitHub이 push 거부 (`remote rejected`)
2. Error message에 "branch protection rule" 명시
3. 로컬 브랜치는 변경되지만 원격은 무영향

### AC-CIAUT-012: No required check → no merge

**Given**: PR이 required check (`Lint`) 실패한 채 사용자가 `gh pr merge <N>` 시도.

**When**: 머지 시도.

**Then**: GitHub이 거부 — "Required statuses must pass before merging".

### AC-CIAUT-013: No release/tag automation

**Given**: T5 적용 후 PR이 main으로 머지됨.

**When**: 머지 완료.

**Then**:
1. `git tag` 자동 생성 안 됨
2. `gh release create` 자동 호출 안 됨
3. GoReleaser workflow는 사용자가 수동으로 tag push 시에만 트리거 (`feedback_release_no_autoexec.md` 준수)

### AC-CIAUT-021: required_status_checks SSoT validation (REQ-CIAUT-027)

**Given**: `.github/required-checks.yml` SSoT 파일이 존재하고 다음 contexts를 정의:
```yaml
required:
  - Lint
  - Test (ubuntu-latest)
  - Test (macos-latest)
  - Test (windows-latest)
  - Build (linux/amd64)
  - Build (linux/arm64)
  - Build (darwin/amd64)
  - Build (darwin/arm64)
  - Build (windows/amd64)
  - CodeQL
auxiliary:
  - claude-code-review
  - llm-panel
  - docs-i18n-check
```

**When**: `moai github init` 또는 branch protection 설정이 트리거됨.

**Then**:
1. `gh api -X PUT` JSON payload의 `required_status_checks.contexts` 배열이 SSoT YAML의 `required:` 섹션에서 동적으로 빌드됨
2. SSoT에 새 context가 추가되면 다음 `moai update` 시 branch protection도 자동 갱신 (또는 사용자에게 갱신 안내)
3. SSoT에 정의되지 않은 context를 hardcode한 코드 위치 제로 (`grep -r 'Test (ubuntu-latest)' internal/` 결과는 빈 검색)

**Failure Mode**: SSoT 파일이 missing이면 `moai github init`가 명확한 에러 메시지로 실패 ("required-checks.yml not found, run `moai update` to generate from template").

### AC-CIAUT-022: `gh` auth-failure graceful exit (REQ-CIAUT-028)

**Given**: `gh` CLI가 auth되지 않음 (`gh auth status` returns non-zero) OR 사용자가 admin permission 없음.

**When**: `moai github init` 또는 `gh api -X PUT .../branches/main/protection` 자동 호출 시도.

**Then**:
1. False success 메시지 출력 금지 ("Branch protection applied!" 같은 텍스트는 절대 출력 안 됨)
2. 실패 사실을 명확히 보고: "gh CLI not authenticated" 또는 "admin permission required"
3. 사용자가 수동으로 실행할 정확한 `gh api -X PUT ...` 명령어를 stdout에 출력 (copy-paste 가능)
4. exit code non-zero (로그 자동화 도구가 감지 가능)
5. `.moai/logs/branch-protection-attempts.log`에 실패 timestamp + 원인 기록

**Failure Mode**: `gh` 자체가 미설치된 경우 — 동일하게 graceful exit + install 안내 ("install gh: https://cli.github.com").

---

## 7. T6 — Worktree State Guard

### AC-CIAUT-014: Worktree state divergence detection

**Given**: 오케스트레이터가 `Agent(isolation: "worktree", ...)` 4회 invocation을 준비.

**When**: 첫 invocation 후 working tree state가 변경됨 (HEAD 변경 또는 untracked file 사라짐).

**Then**:
1. Pre-call snapshot vs post-call state diff 감지
2. 후속 invocation 차단
3. `.moai/reports/worktree-guard/2026-05-05.md`에 divergence detail 기록
4. AskUserQuestion으로 사용자 결정 (restore / accept / abort)

### AC-CIAUT-015: Empty worktreePath suspect handling

**Given**: `Agent(isolation: "worktree")` 호출.

**When**: Agent 응답에 `worktreePath: {}` (empty).

**Then**:
1. Suspect flag 설정
2. Agent 결과 활용 시 사용자 경고
3. Subsequent push 차단 (사용자 확인 필수)
4. `claude-code-guide` 위임으로 upstream investigation 트리거

---

## 8. T7 — i18n Validator

### AC-CIAUT-016: i18n validator blocks PR #783 regression

**Given**: 사용자가 PR #783 replay scenario 시도 — `internal/cli/release_test.go`의 `mockReleaseData`에서 `"유효한 YAML 문서가 아닙니다"` → `"Not a valid YAML document"` 변경.

**When**: `make ci-local` 실행 (또는 CI에서 validator job).

**Then**:
1. Validator가 `mockReleaseData` 키를 test에서 참조하는 것 감지
2. Translation-locked literal 변경 보고
3. Non-zero exit
4. Error message: "string literal at <file>:<line> is referenced by <test_name>:<line>, translation requires test update"

### AC-CIAUT-017: i18n:translatable magic comment exempt

**Given**: 사용자가 정당한 CLI 메시지를 번역.

```go
const errMsg = "Failed to load config" // i18n:translatable
```

**When**: Validator scan.

**Then**: Magic comment 감지하여 lock check에서 제외. 번역 통과.

### AC-CIAUT-023: i18n validator 30s wall-clock budget (REQ-CIAUT-040)

**Given**: moai-adk-go 전체 codebase (~100K LOC, 1000+ Go 파일) 상태에서 i18n validator 실행.

**When**: `scripts/i18n-validator/main.go` 또는 `make ci-local` 내 validator step 호출 (cold cache, 첫 실행).

**Then**:
1. wall-clock duration ≤ 30초 (CI runner 기준; pre-push hook 통합 시 사용자 경험 보호)
2. progress streaming: 5초 이상 silent 금지 (현재 처리 중 파일 또는 단계 stderr 출력)
3. 30s 초과 시 timeout exit + error message: "validator exceeded 30s budget, consider scoping to changed files only"
4. caching enabled mode (재실행, hot cache): ≤ 5초 목표

**Verification**:
- `time scripts/i18n-validator/main.go ./...` benchmark in CI
- `internal/scripts/i18n_validator_perf_test.go` (or shell-based perf harness) verifies budget compliance
- 30s 초과 회귀 발생 시 CI fail (regression guard)

**Failure Mode**: 매우 큰 string literal 파일 (수만 개)이 있는 경우 — `--max-files` CLI flag로 partial scan 옵션 제공.

---

## 9. T8 — Branch Origin Decision Protocol

### AC-CIAUT-018: BODP recommends "main에서 분기" via `/moai plan --branch` (current session replay)

**Given**: 현재 세션 상태 (실제 발생):
- 현재 브랜치: `chore/translation-batch-b`
- Working tree untracked: `.moai/specs/SPEC-V3R3-MX-INJECT-001/`
- 사용자 요청: "create SPEC-V3R3-CI-AUTONOMY-001"
- 현재 브랜치의 open PR 없음 (이미 머지됨 또는 closed)

**When**: 사용자가 `/moai plan "Autonomous CI/CD pipeline" --branch` 실행 (또는 SPEC ID와 함께 `/moai plan SPEC-V3R3-CI-AUTONOMY-001 --branch --resume`). **새 명령어 추가 없이 기존 `/moai plan --branch` 흐름 사용**.

**Then**:
1. plan workflow Phase 3 Branch Path 진입 직전, `internal/bodp/relatedness.go`가 3-signal 검사 실행
2. Signal (a) check: SPEC-V3R3-CI-AUTONOMY-001은 chore/translation-batch-b의 diff와 무관 → negative
3. Signal (b) check: untracked SPEC-V3R3-MX-INJECT-001은 다른 SPEC ID → negative
4. Signal (c) check: open PR with head=chore/translation-batch-b 없음 → negative
5. orchestrator AskUserQuestion: 1번째 옵션 `main에서 분기 (권장)`, 2번째 `현재 브랜치에서 분기 (stacked)`
6. 사용자 confirm 후 `manager-git` 호출이 base를 `origin/main`으로 강제: `git fetch origin main && git checkout -B feat/SPEC-V3R3-CI-AUTONOMY-001 origin/main`
7. `.moai/branches/decisions/feat-SPEC-V3R3-CI-AUTONOMY-001.md` audit trail 생성 (entry point: `plan-branch`)

### AC-CIAUT-019: BODP recommends stacked PR via `/moai plan --branch` (signal a positive)

**Given**:
- 현재 브랜치: `feat/SPEC-AUTH-001-base`
- 사용자가 `/moai plan SPEC-AUTH-002 --branch --resume` 실행
- SPEC-AUTH-002 plan에 `depends_on: [SPEC-AUTH-001]` 명시

**When**: plan workflow Phase 3 진입.

**Then**:
1. BODP signal (a): SPEC-AUTH-002.depends_on에 SPEC-AUTH-001 명시 → 현재 브랜치(SPEC-AUTH-001-base)와 매칭 → positive
2. AskUserQuestion 1번째 옵션: `현재 브랜치에서 분기 (stacked PR, base=feat/SPEC-AUTH-001-base) (권장)`
3. 사용자 confirm 후 `manager-git`이 stacked branch 생성: `git checkout -B feat/SPEC-AUTH-002` (현재 HEAD 유지)
4. PR 생성 시 base가 `feat/SPEC-AUTH-001-base`로 자동 설정 (manager-git에 parameter 전달)
5. Audit trail에 "stacked PR rationale: depends_on detected" + entry point `plan-branch` 기록

### AC-CIAUT-019b: BODP via `moai worktree new` CLI (default origin/main)

**Given**: 현재 브랜치 `chore/translation-batch-b`, 사용자가 `moai worktree new SPEC-AUTH-001` 직접 실행 (slash 우회).

**When**: CLI 호출.

**Then**:
1. CLI는 AskUserQuestion 호출하지 않음 (orchestrator-only HARD 준수)
2. Default behavior: base = `origin/main`. `git fetch origin main` 실행 후 `git worktree add ... origin/main`
3. Audit trail entry point: `worktree-cli` 기록
4. `--from-current` flag 사용 시 기존 동작 (current HEAD에서 worktree) — opt-out 명시적

### AC-CIAUT-024: `moai status` off-protocol branch reminder (REQ-CIAUT-050)

**Given**: 사용자가 BODP-aware 경로 (`/moai plan --branch`, `/moai plan --worktree`, `moai worktree new`)를 우회하고 raw `git checkout -b feat/quick-fix` 실행. BODP는 호출되지 않음. `.moai/branches/decisions/feat-quick-fix.md` audit trail 파일은 미생성.

**When**: 사용자가 다음 `moai status` 또는 `/moai sync` 등 MoAI 명령 실행.

**Then**:
1. `moai status`가 현재 브랜치 (`feat/quick-fix`) 감지 후 audit trail 디렉토리에서 매칭 파일 검색
2. 매칭 파일 없음 → "off-protocol branch detected" friendly reminder 출력 (warning level, error 아님)
3. Reminder 내용: "Branch `feat/quick-fix` was created without going through MoAI entry points. Future branches: use `/moai plan --branch <name>` (SPEC-tied) or `moai worktree new <SPEC-ID>` for relatedness check + audit trail. Skip with `MOAI_NO_BODP_REMINDER=1` if intentional."
4. Reminder는 hard block 안 함 — 작업 흐름 유지
5. 사용자가 `--no-bodp-reminder` 환경변수 설정 시 reminder 비활성화 (`MOAI_NO_BODP_REMINDER=1`)

**Failure Mode**: audit trail 디렉토리 자체가 없는 경우 (신규 프로젝트) — reminder 출력 안 함 (false positive 방지).

### AC-CIAUT-025: CLAUDE.local.md §18.12 BODP subsection added (REQ-CIAUT-051)

**Given**: 본 SPEC Wave 7 완료 후 CLAUDE.local.md 갱신.

**When**: `grep -A 10 '§18.12' CLAUDE.local.md` 또는 `grep '## 18.12' CLAUDE.local.md` 실행.

**Then**:
1. §18 (Enhanced GitHub Flow) 하위에 §18.12 (Branch Origin Decision Protocol) subsection 존재
2. subsection 내용:
   - BODP 알고리즘 요약 (3-axis relatedness check)
   - **3개 invocation path 명시 (모두 기존 entry, 새 명령어 ZERO)**:
     - `/moai plan --branch [name]` — SPEC-tied feature 브랜치
     - `/moai plan --worktree` — SPEC-tied worktree
     - `moai worktree new <SPEC-ID>` — 직접 worktree CLI
   - default recommendation = "main에서 분기"
   - audit trail 위치 (`.moai/branches/decisions/`)
   - opt-out 방법: `MOAI_NO_BODP_REMINDER=1` 환경변수 또는 raw `git checkout -b` 사용 (자기 책임)
3. CLAUDE.local.md table of contents 또는 §18 인덱스에 §18.12 entry 추가
4. PR description에 CLAUDE.local.md diff 포함

**Verification**: `internal/template/templates/CLAUDE.local.md`에도 동일 subsection 미러됨 (Template-First 준수, CLAUDE.local.md §2). `.claude/rules/moai/development/branch-origin-protocol.md` 신규 rule 파일 생성 (paths frontmatter로 자동 적용 범위 정의).

---

## 10. Cross-Tier Acceptance

### AC-CIAUT-020: 5-PR Sweep Replay (Manual Validation)

> **Audit finding F-005 resolution**: 본 AC는 fixture/runner 기반 자동화가 아닌 **수동 검증 (manual validation)** 으로 정의됨. 사용자 확정 (2026-05-05): "F-005 → AC-020을 manual validation으로 다운그레이드". 이유: 원래 5-PR 실패는 시간 종속 실제 CI 이벤트로 fixture로 정확히 재현 어려움 + 5개 fixture 디렉토리 + replay runner 구축이 audit 가치 대비 noise 큼.

**Given**: 본 SPEC 모든 Tier (T1-T8) main 머지 완료 후 다음 dev cycle.

**Validation Method (Manual)**:
- SPEC 작성자 + 1 reviewer가 본 SPEC 머지 후 30일 이내 발생한 PR sweep 또는 multi-PR development cycle을 회고
- 본 SPEC 머지 전 vs 후의 인적 개입 횟수 (manual debug + fix + push 사이클) 정량 비교
- PR description 또는 retrospective 문서에 다음 표 작성:

| 비교 차원 | SPEC 머지 전 (5-PR sweep 2026-05-05) | SPEC 머지 후 (next sweep) | 개선 |
|----------|--------------------------------------|---------------------------|------|
| 평균 manual debug 횟수 / PR | ~3-4회 | (실측) | (계산) |
| Pre-push 차단으로 사전 검출된 실패 | 0건 | (실측) | (계산) |
| Auto-fix loop이 처리한 mechanical failure | 0건 (loop 미존재) | (실측) | (계산) |
| Semantic 실패 즉시 escalation | 수동 (사용자가 진단) | (실측) | (계산) |
| 평균 PR ready→merge 시간 | (사용자 추정) | (실측) | (계산) |

**Acceptance Criterion**:
1. 본 SPEC 머지 후 첫 dev cycle (3+ PR 포함)에서 위 5개 dimension 모두 측정값 기록됨
2. 평균 manual debug 횟수 / PR이 머지 전 baseline (~3-4회) 대비 50% 이상 감소
3. PR description에 측정 결과 + run links 첨부
4. 회고 문서가 `.moai/reports/post-merge-validation/SPEC-V3R3-CI-AUTONOMY-001.md`에 저장

**원래 5-PR 정성적 분석 (참고용, normative 아님)**:

| PR | 원래 (manual debug+fix+push) | T1-T8 적용 후 (예상) |
|----|---|---|
| #783 (i18n) | 사용자가 ST1005 + map 데이터 손상 검출하고 revert | T7 validator가 push 전 차단 (~0 manual) |
| #744 (CACHE-ORDER) | golangci-lint errcheck를 수동 fix | T1 pre-push 차단 → 사용자 fix → T3 watch (~1 manual) |
| #739 (MCP) | unused/QF1003를 수동 fix | T1 차단 → fix → T3 mechanical autofix (~0-1 manual) |
| #747 (HOOK) | ETXTBSY race를 사용자가 retry 결정 | T3 semantic 즉시 escalation (~사용자 결정만) |
| #743 (STATUSLINE) | Windows-only failure를 수동 fix | T1 cross-compile build 사전 검출 (~0 manual) |

**Failure Mode**: 본 SPEC 머지 후 첫 dev cycle에서 측정 불가 (PR 부족, 외부 차단 등) → 60일 grace window 후 재측정 권고.

---

## 11. Definition of Done (DoD) Checklist

### Per-Wave DoD

각 Wave 종료 시 모두 통과:
- [ ] Wave의 모든 task 완료 (plan.md task 목록 모두 체크)
- [ ] 신규 추가 파일이 `internal/template/templates/`에 mirror됨 (Template-First, CLAUDE.local.md §2)
- [ ] `make build` 후 embedded.go 갱신
- [ ] `make ci-local` 통과
- [ ] `go test ./...` 통과
- [ ] `golangci-lint run` 통과
- [ ] Wave-specific acceptance scenarios 통과
- [ ] Conventional Commits 형식 + 🗿 MoAI co-author trailer
- [ ] PR 생성 시 type/priority/area 3축 label 부착 (CLAUDE.local.md §18.6)

### SPEC-Level DoD (전체)

본 SPEC 종료 시:
- [ ] 모든 7 Wave 완료
- [ ] 20개 acceptance scenario 모두 통과
- [ ] 5-PR sweep replay (AC-CIAUT-020) 정량적 개선 확인
- [ ] CLAUDE.local.md §18.12 BODP subsection 추가
- [ ] `.claude/rules/moai/development/branch-origin-protocol.md` 신규 규칙
- [ ] `.claude/rules/moai/workflow/worktree-state-guard.md` 신규 규칙
- [ ] `.claude/rules/moai/workflow/ci-watch-protocol.md` 신규 규칙
- [ ] `.claude/rules/moai/workflow/ci-autofix-protocol.md` 신규 규칙
- [ ] `.github/required-checks.yml` SSoT 작성
- [ ] CHANGELOG.md에 SPEC 머지 entry
- [ ] PR이 본 SPEC 자체로 auto-watch + (해당하는 경우) auto-fix 사이클 한 번 통과 — eat your own dogfood
- [ ] AskUserQuestion 4개 OQ 모두 사용자 확정 + spec.md 반영
- [ ] No release/tag automation 검증 (T5 후 main 머지 시 tag/release 자동 생성 없음)
- [ ] No hardcoded URLs/models (CLAUDE.local.md §14)
- [ ] 16-language neutrality 검증 (T1, T7만 해당)

### Quality Gate

- [ ] **Tested**: 각 Wave의 코드는 unit test + integration test 보유, coverage ≥85%
- [ ] **Readable**: ruff/golangci-lint zero warning
- [ ] **Unified**: gofmt/goimports 일관성
- [ ] **Secured**: pre-push hook이 secrets 노출 차단; auto-fix loop이 token/key 변경 안 함
- [ ] **Trackable**: 모든 commit이 SPEC-V3R3-CI-AUTONOMY-001 reference 포함

---

## 12. Out-of-Scope Verification (의도적 미검증)

본 SPEC이 다루지 않는 것을 명시:

- [ ] (의도적 SKIP) Anthropic upstream `Agent(isolation:)` 회귀 fix — claude-code-guide 위임 결과만 보고
- [ ] (의도적 SKIP) 16개 언어 전체 i18n validator — Go만 검증
- [ ] (의도적 SKIP) Release tag/GoReleaser 자동화 — 영구 prohibited
- [ ] (의도적 SKIP) `git checkout -b` raw 호출 가로채기 — opt-in 유지
- [ ] (의도적 SKIP) Auxiliary workflow의 fix — disable만, 사용자 별도 SPEC
- [ ] (의도적 SKIP) Multi-repo branch protection 일괄 적용

---

Version: 0.1.0
Status: draft
Last Updated: 2026-05-05
