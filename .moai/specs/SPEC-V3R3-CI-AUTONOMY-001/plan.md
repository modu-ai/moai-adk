---
id: SPEC-V3R3-CI-AUTONOMY-001
version: "0.1.0"
status: draft
created_at: 2026-05-05
updated_at: 2026-05-05
author: manager-spec
priority: P0
labels: [ci-cd, automation, worktree, branch-protection, plan, v3r3]
issue_number: null
breaking: false
---

# SPEC-V3R3-CI-AUTONOMY-001: Implementation Plan

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-05 | manager-spec | Initial 7-Wave plan. Wave-split per `feedback_large_spec_wave_split.md` to avoid Anthropic SSE stream stalls. Each wave produces a ~1.5KB delegation prompt with 5-10 atomic tasks. |

---

## 1. Wave Strategy Overview

본 SPEC은 8개 Tier(T1-T8)를 7개 Wave로 패키징한다. Wave 1은 Quick Wins (T1+T5), Wave 2-7는 각 Tier 단위. Wave 간 의존성:

```
Wave 1 (T1+T5)        ──┐
                        ├─→ Wave 2 (T2)  ──→ Wave 3 (T3)
                        │
                        ├─→ Wave 4 (T4)  (independent)
                        │
                        ├─→ Wave 5 (T6)  (independent)
                        │
                        ├─→ Wave 6 (T7)  ──→ depends on Wave 1 ci-local
                        │
                        └─→ Wave 7 (T8)  (independent)
```

Wave 1은 Quick Wins로 즉시 가치 제공. Wave 2 → Wave 3는 sequential dependency. Wave 4/5/7은 독립적, 병렬 가능. Wave 6은 Wave 1의 `ci-local` framework 의존.

각 Wave 진입 전 `/clear` 권장 (Opus 4.7 1M context 한계 활용). Wave 완료 시 progress.md 업데이트 후 다음 Wave 위임.

## 2. Milestones

priority-based, no time estimates (per CLAUDE.md §4 + agent-common-protocol §Time Estimation):

| Milestone | Tier | Deliverable | Priority |
|-----------|------|-------------|----------|
| M1 | T1+T5 | Pre-push hook + ci-local + branch protection | P0 |
| M2 | T2 | CI watch loop (skill or sync extension) | P0 |
| M3 | T3 | Auto-fix loop with expert-debug + escalation | P1 |
| M4 | T4 | Auxiliary workflow cleanup | P1 |
| M5 | T6 | Worktree state guard | P1 |
| M6 | T7 | i18n validator | P2 |
| M7 | T8 | BODP CLI command + audit trail | P0 |

---

## 3. Wave 1 — Quick Wins (T1 + T5) — ~7 tasks

**Goal**: 5-PR sweep의 P2 (lint failure 미검출) 즉시 해결 + branch protection 적용으로 force-push/CI-skip 차단.

**Files**:
- `internal/template/templates/.git_hooks/pre-push` (new)
- `Makefile` (extend)
- `scripts/ci-mirror/run.sh` (new, 16-language language detection)
- `scripts/ci-mirror/lib/go.sh`, `python.sh`, `node.sh`, `rust.sh` (new, language modules)
- `internal/cli/github_init.go` (extend with branch protection prompt)
- `.github/required-checks.yml` (new, SSoT for required check contexts)
- `internal/template/templates/.github/required-checks.yml` (mirror)

**Tasks**:
1. **W1-T01**: `scripts/ci-mirror/run.sh` 스켈레톤 작성 — language detection + dispatch to `lib/<lang>.sh`
2. **W1-T02**: `lib/go.sh` 구현 — `go vet`, `golangci-lint run`, `go test -race ./...`, cross-compile (linux/darwin/windows × amd64/arm64) with `GOOS=windows go build -o /dev/null`
3. **W1-T03**: `lib/python.sh`, `lib/node.sh`, `lib/rust.sh`, `lib/java.sh` 등 16개 언어 lightweight 분기 (각 5-15 lines)
4. **W1-T04**: Makefile `ci-local` target — calls `scripts/ci-mirror/run.sh` with progress streaming
5. **W1-T05**: `internal/template/templates/.git_hooks/pre-push` — POSIX sh script that calls `make ci-local`, supports `--no-verify` bypass logging to `.moai/logs/prepush-bypass.log`
6. **W1-T06**: `internal/cli/github_init.go` 확장 — AskUserQuestion으로 branch protection 적용 동의 받고 `gh api -X PUT` 실행 (CLAUDE.local.md §18.7 JSON SSoT 사용)
7. **W1-T07**: `.github/required-checks.yml` SSoT 작성 + `internal/cli/github_init.go`에서 동적으로 contexts 배열 생성

**Tests**:
- `internal/cli/github_init_test.go`: AskUserQuestion mock + `gh api` invocation verification (테이블 기반)
- `scripts/ci-mirror/test/run_test.sh`: bash test harness, 5개 언어 각각에서 `make ci-local` 동작 검증

**Acceptance**:
- 5-PR sweep replay (PR #739 errcheck violation) 시 push가 pre-push hook에서 차단되어야 함
- `gh api` 실행 후 `git push --force origin main` 거부 확인

**Dependencies**: 없음 (foundational)

---

## 4. Wave 2 — CI Watch Loop (T2) — ~10 tasks

**Goal**: PR 생성 후 사용자가 `gh pr checks` 폴링 없이 자동으로 CI 상태 모니터링 + 결과 보고. T3 진입 트리거 제공.

**Files**:
- `internal/template/templates/.claude/skills/moai-workflow-ci-watch/SKILL.md` (new) OR
- `internal/template/templates/.claude/skills/moai/workflows/sync.md` (extend Phase 4)
- `internal/cli/pr_watch.go` (new, optional CLI helper)
- `scripts/ci-watch/run.sh` (new, gh CLI wrapper)
- `internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md` (new)

**Tasks**:

> OQ1 사전 해결 (audit finding F-002): 사용자가 AskUserQuestion으로 "신규 skill `moai-workflow-ci-watch`" 선택 (관심사 분리). 본 Wave는 이 결정을 전제로 진행.

1. **W2-T01**: 신규 skill `internal/template/templates/.claude/skills/moai-workflow-ci-watch/SKILL.md` 스켈레톤 작성 (YAML frontmatter + Progressive Disclosure 구조: Quick Reference + Implementation + Advanced)
2. **W2-T02**: `scripts/ci-mirror/run.sh`에서 mirror 로직 추출 검증 (Wave 1 W1-T01과 같은 파일이지만 watch 진입을 위한 metadata 추가)
3. **W2-T03**: `scripts/ci-watch/run.sh` — `gh pr checks <PR> --watch --json` polling, required vs auxiliary 구분
4. **W2-T04**: required-checks list를 `.github/required-checks.yml`에서 동적 로딩 (SSoT 재사용)
5. **W2-T05**: 30초 간격 status report 포맷 정의 (natural-language summary, 짧게)
6. **W2-T06**: skill SKILL.md 작성 (Progressive Disclosure: Quick Reference + Implementation + Advanced) OR sync.md Phase 4 확장
7. **W2-T07**: `ci-watch-protocol.md` 규칙 — when to invoke, timeout, abort condition
8. **W2-T08**: 성공 시 ready-to-merge 알림 + (T5 적용 후) auto-merge 활성화 AskUserQuestion 트리거
9. **W2-T09**: 실패 시 T3 auto-fix loop 진입 메타데이터 (실패한 check 이름, run-id, log URL) 캡처
10. **W2-T10**: 30분 timeout + 사용자 abort 경로 (`Ctrl+C` 우회로 `.moai/state/ci-watch-active.flag` 파일 제거)

**Tests**:
- `scripts/ci-watch/test/run_test.sh`: mock `gh pr checks` 응답으로 required/auxiliary 분류 검증
- skill 적용 시 manual integration test (실제 PR로)

**Acceptance**:
- PR 생성 후 30초 이내 watch 진입 확인
- required check 실패 시 T3 트리거 메타데이터 정확
- auxiliary check 실패는 ready-to-merge 차단 안 함

**Dependencies**: Wave 1 (`scripts/ci-mirror/`, `.github/required-checks.yml`)

---

## 5. Wave 3 — Auto-Fix Loop (T3) — ~10 tasks

**Goal**: CI 실패 시 mechanical fix는 expert-debug가 자동 시도, semantic은 즉시 escalate. Max 3 iterations.

**Files**:
- `internal/template/templates/.claude/skills/moai-workflow-ci-autofix/SKILL.md` (new)
- `internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md` (new)
- `internal/template/templates/.claude/agents/expert-debug.md` (extend with CI failure interpretation)
- `scripts/ci-autofix/classify.sh` (new, mechanical vs semantic 분류기)
- `scripts/ci-autofix/log-fetch.sh` (new, `gh run view --log-failed` wrapper)

**Tasks**:

> OQ2 사전 해결 (audit finding F-002): 사용자가 AskUserQuestion으로 "첫 iter 강제 confirm + 2-3 iter는 trivial fix만 silent" 선택. "Trivial" 정의 = whitespace, gofmt/goimports, import order. 그 외 모든 mechanical fix도 confirm.

1. **W3-T01**: Iteration cadence implementation — orchestrator 측 state machine: iter 1 → confirm 강제, iter 2-3 → classifier가 "trivial" 분류 시 silent + log only, "non-trivial mechanical" 분류 시 confirm 강제
2. **W3-T02**: failure classifier — lint/format/missing-import/typo (mechanical) vs assertion/race/deadlock/panic (semantic) regex/pattern set + trivial sub-classifier (whitespace/gofmt/goimports/import-order)
3. **W3-T03**: `scripts/ci-autofix/log-fetch.sh` — failed check id에서 log 추출, PR diff와 결합
4. **W3-T04**: `expert-debug` agent prompt 확장 — CI failure log + diff context로 patch 제안
5. **W3-T05**: orchestrator iteration loop (max 3) — patch propose → AskUserQuestion → apply → push → re-watch
6. **W3-T06**: semantic 즉시 escalation 경로 — `expert-debug`는 진단만, patch 시도 안 함
7. **W3-T07**: iteration count == 3 시 mandatory escalation AskUserQuestion (continue manually / revise SPEC / abandon PR)
8. **W3-T08**: `.moai/reports/ci-autofix/<PR-NNN>-<YYYY-MM-DD>.md` 로깅 schema + writer
9. **W3-T09**: integration with Wave 2 watch loop (W2-T09 메타데이터 consumption)
10. **W3-T10**: ci-autofix-protocol.md 규칙 — max iteration, force-push 금지, 모든 patch는 새 commit

**Tests**:
- 5-PR sweep replay scenarios (#783 P1 i18n / #739 P2 errcheck / #747 P5 ETXTBSY)
- mechanical case는 1-2 iteration 내 해결, semantic case는 즉시 escalate

**Acceptance**:
- P2 errcheck 같은 mechanical → expert-debug가 patch 제안, 1 iteration 내 해결
- P5 ETXTBSY race → 즉시 escalate, AskUserQuestion에 "race condition detected" 포함
- 3 iteration 후 항상 사용자 결정 요청

**Dependencies**: Wave 2 (watch loop이 트리거 제공)

---

## 6. Wave 4 — Auxiliary Workflow Cleanup (T4) — ~5 tasks

**Goal**: claude-code-review/llm-panel/Release Drafter 노이즈 제거, docs-i18n-check non-blocking 전환. CI signal 회복.

**Files**:
- `.github/workflows/optional/claude-code-review.yml` (move from `.github/workflows/`)
- `.github/workflows/optional/llm-panel.yml` (move)
- `.github/workflows/docs-i18n-check.yml` (set `continue-on-error: true`)
- `.github/workflows/release-drafter-cleanup.yml` (new, scheduled)
- `Makefile` (add `ci-disable WORKFLOW=<name>` target)

**Tasks**:
1. **W4-T01**: claude-code-review.yml + llm-panel.yml을 `.github/workflows/optional/`로 이동 (트리거 그대로 유지하되 required-checks.yml에서 제외)
2. **W4-T02**: docs-i18n-check.yml에 `continue-on-error: true` + advisory PR comment 추가
3. **W4-T03**: release-drafter-cleanup.yml 스케줄 workflow — 30일+ stale draft 자동 close (cron 주 1회)
4. **W4-T04**: `make ci-disable WORKFLOW=<name>` Makefile target — workflow trigger를 comment-out + commit
5. **W4-T05**: `.github/required-checks.yml` 검증 — auxiliary가 required에 포함 안 됨

**Tests**:
- `gh workflow list` 후 optional/ 디렉토리 인식 확인
- docs-i18n-check 실패 PR이 머지 차단 안 됨 manual test

**Acceptance**:
- 본 SPEC 머지 후 PR에서 claude-review 실패 → ready-to-merge 유지
- 80+ stale Release Drafter draft 자동 정리

**Dependencies**: 독립 (Wave 1의 required-checks.yml SSoT만 참조)

---

## 7. Wave 5 — Worktree State Guard (T6) — ~8 tasks

**Goal**: `Agent(isolation: "worktree")` 호출 전후 working tree 상태 보존. 회귀 발생 시 자동 감지 + 사용자 알림.

**Files**:
- `internal/template/templates/.claude/rules/moai/workflow/worktree-state-guard.md` (new)
- `internal/orchestrator/worktree_guard.go` (new, state snapshot/diff)
- `internal/orchestrator/worktree_guard_test.go` (new)
- `internal/template/templates/.claude/agents/claude-code-guide.md` (extend with upstream investigation prompt)
- `.moai/reports/upstream/agent-isolation-regression.md` (placeholder for Wave 5 deliverable)

**Tasks**:
1. **W5-T01**: state snapshot 함수 — `git status --porcelain`, HEAD SHA, branch name, untracked under `.moai/specs/` 캡처
2. **W5-T02**: state diff 함수 — pre-call vs post-call 비교, divergence dimension 분류 (HEAD/branch/untracked)
3. **W5-T03**: orchestrator wrapper — `Agent()` 호출 인접 위치에 snapshot/verify hook
4. **W5-T04**: divergence 발생 시 `.moai/reports/worktree-guard/<YYYY-MM-DD>.md` 로깅 + AskUserQuestion (restore/accept/abort)
5. **W5-T05**: empty `worktreePath: {}` 응답 감지 + suspect flag 설정 + push 차단
6. **W5-T06**: state restore 옵션 — `git restore --source=<sha> --staged --worktree :/` + untracked path 재첨부
7. **W5-T07**: `claude-code-guide` 위임 — Anthropic upstream 회귀 조사 + bug report 작성 (`.moai/reports/upstream/agent-isolation-regression.md`)
8. **W5-T08**: worktree-state-guard.md 규칙 문서 — when to snapshot, divergence threshold, escalation path

**Tests**:
- `internal/orchestrator/worktree_guard_test.go` — 4개 case (no diff, untracked added, untracked removed, branch changed)
- manual integration: 알려진 회귀 시나리오 4회 호출 후 alert 동작 확인

**Acceptance**:
- 4회 연속 `Agent(isolation:)` 호출에서 worktreePath 비어있는 경우 → 1회차에서 alert
- divergence 발생 후 user "restore" 선택 시 working tree 완전 복원

**Dependencies**: 독립 (Wave 1과 무관)

---

## 8. Wave 6 — i18n Validator (T7) — ~7 tasks

**Goal**: `mockReleaseData` 같은 테스트 의존 string literal이 번역되어 손상되는 사고 차단.

**Files**:
- `scripts/i18n-validator/main.go` (new, Go AST static analyzer)
- `scripts/i18n-validator/lockset.go` (new, translation-lock detection)
- `scripts/i18n-validator/main_test.go` (new)
- `scripts/i18n-validator/testdata/` (test fixtures)
- `scripts/ci-mirror/lib/go.sh` (extend to invoke validator)

**Tasks**:
1. **W6-T01**: Go AST parser — `require.Equal`, `assert.Contains`, `assert.Equal` 등 testify call의 string literal 인자 추출
2. **W6-T02**: cross-file lookup — test가 참조하는 source string literal 위치 추적 (e.g., `mockReleaseData["key"]`)
3. **W6-T03**: translation-lock set 빌드 — 모든 test-referenced literal을 `(file, line, content)` triple로 기록
4. **W6-T04**: `// i18n:translatable` magic comment escape 처리
5. **W6-T05**: PR diff 입력 시 변경된 string literal 중 lock set과 교차 검사 → 충돌 시 non-zero exit
6. **W6-T06**: `scripts/ci-mirror/lib/go.sh`에서 validator 호출 통합 (T1 ci-local에서 자동 실행)
7. **W6-T07**: 30s wall-clock budget 검증 (전체 repo scan)

**Tests**:
- testdata: PR #783의 mockReleaseData 손상 시나리오 fixture
- magic comment exempt 시나리오
- 정상 translation (사용자 메시지)이 통과하는지 확인

**Acceptance**:
- PR #783 replay → validator가 `Not a valid YAML document` 변경을 차단
- 정상적인 사용자 메시지 i18n 변경은 통과

**Dependencies**: Wave 1 (`scripts/ci-mirror/lib/go.sh`)

---

## 9. Wave 7 — Branch Origin Decision Protocol (T8) — ~6 tasks

**Goal**: 새 브랜치 생성 시 main에서 분기를 default로 강제. **새 명령어 추가 없이** 기존 entry points (`/moai plan --branch`, `/moai plan --worktree`, `moai worktree new`)에 BODP 로직 내장.

**Files**:
- `internal/bodp/relatedness.go` (new) — 3-signal check + decision matrix (pure-Go library, no CLI)
- `internal/bodp/relatedness_test.go` (new)
- `internal/bodp/audit_trail.go` (new) — `.moai/branches/decisions/<branch>.md` writer
- `.claude/skills/moai/workflows/plan.md` Phase 3 (extend Branch Path + Worktree Path with BODP gate)
- `internal/cli/worktree/new.go` (extend with `--from-current` flag + default origin/main + BODP audit trail)
- `internal/cli/status.go` (extend with off-protocol branch detection)
- `internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md` (new)
- `internal/template/templates/CLAUDE.local.md` §18.12 추가 (mirror)
- `.moai/branches/decisions/.gitkeep` (new directory)

**Tasks**:

> Architecture (resolves user critique 2026-05-05): 새 슬래시 명령/CLI 서브명령 ZERO. 기존 `/moai plan --branch`, `/moai plan --worktree`, `moai worktree new` 핸들러에 BODP 로직 내장. T8 작업량 9 → 6 tasks (~33% 감소).

1. **W7-T01**: BODP 라이브러리 `internal/bodp/relatedness.go` — 3-signal check 함수 + decision matrix. Pure-Go 모듈, CLI 또는 슬래시 명령 ZERO. 단위 테스트 4-5 case (all-negative, signal-a-only, signal-b-only, signal-c-only).
2. **W7-T02**: `/moai plan` skill body Phase 3 확장 — Branch Path와 Worktree Path 모두에서 manager-git 위임 전 BODP 검사 + AskUserQuestion으로 base 결정 + 결정사항을 manager-git에게 parameter로 전달
3. **W7-T03**: `moai worktree new` CLI (`internal/cli/worktree/new.go`) 확장 — default base를 `origin/main`으로 변경, `--from-current` flag로 기존 동작 opt-out. 변경 전 audit trail 기록.
4. **W7-T04**: `internal/bodp/audit_trail.go` writer — `.moai/branches/decisions/<branch>.md` 생성 (timestamp, invocation path, signals, decision, command). W7-T02 + W7-T03에서 호출.
5. **W7-T05**: `moai status` 확장 — `.moai/branches/decisions/`에 audit trail 없는 off-protocol 브랜치 감지 시 friendly reminder ("이 브랜치는 BODP 경로 외에서 생성되었습니다. 향후 `/moai plan --branch` 또는 `moai worktree new` 사용 권장")
6. **W7-T06**: 문서화 — `CLAUDE.local.md` §18.12 신규 subsection (BODP 알고리즘, 3개 entry points, raw-git-checkout reminder) + `.claude/rules/moai/development/branch-origin-protocol.md` rule + Template-First mirror

**Tests**:
- `internal/cli/branch_new_test.go` — 4개 signal scenario 테이블 기반 (all negative / a only / b only / c only / a+c)
- 현재 세션 replay 시나리오: chore/translation-batch-b + untracked SPEC-MX-INJECT-001 + 신규 SPEC-CI-AUTONOMY-001 → BODP가 main 권장

**Acceptance**:
- 4개 signal 조합 모두 정확한 권장 옵션 출력
- audit trail 파일 자동 생성
- CLAUDE.local.md §18.12 추가 후 `moai update` 시 사용자 프로젝트에 동기화

**Dependencies**: 독립 (다른 Wave와 무관)

---

## 10. Technical Approach

### 10.1 Pre-Push Hook 구현 패턴 (T1)

POSIX sh 사용 (zsh/fish 호환성). `printf` 사용 (`echo`는 escape 불일치). progress streaming은 `tee /dev/stderr` 또는 `>&2`.

```sh
#!/bin/sh
# .git_hooks/pre-push
set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(git rev-parse --show-toplevel)"

if [ "$1" = "--no-verify-bypass" ]; then
  printf '[prepush] Bypass logged.\n' >&2
  echo "$(date -Iseconds) $USER bypass" >> "$PROJECT_ROOT/.moai/logs/prepush-bypass.log"
  exit 0
fi

cd "$PROJECT_ROOT"
make ci-local || {
  printf '\n[prepush] FAIL — see above. Hint: make fmt && make lint && make test\n' >&2
  exit 1
}
```

### 10.2 CI Watch Loop 패턴 (T2)

`gh pr checks <PR> --watch --json status,conclusion,name` JSON 폴링. required vs auxiliary는 SSoT YAML 매칭. 30초 간격은 `sleep 30`으로 단순 구현 (long-running orchestrator session 토큰 우려 시 `--interval 30` flag 활용).

### 10.3 Auto-Fix Loop 분류기 (T3)

regex/keyword set:
- mechanical: `golangci-lint.*errcheck`, `unused-import`, `gofmt`, `staticcheck.*ST1005`, `QF1003`
- semantic: `panic:`, `data race`, `deadlock`, `--- FAIL: Test.*assertion`, `goroutine.*goroutine`

### 10.4 Worktree State Snapshot (T6)

snapshot은 in-memory struct, divergence는 deterministic comparison. `untracked` set은 sorted slice.

### 10.5 BODP Embedded-in-Existing-Entries Pattern (T8)

User critique resolution (2026-05-05): "추가 명령어는 자제. 기존 명령어 옵션을 분석해서 제대로 사용하자". BODP는 **새 명령어 0개**, 기존 entry points 3개에 동작 내장.

**Existing entries reused**:

1. **`/moai plan --branch`** (skill, orchestrator):
   - 현재: `manager-git` 위임 → `feature/SPEC-{ID}-{desc}` 생성 (현재 HEAD에서)
   - 변경: 위임 직전 BODP 검사 → AskUserQuestion(권장 base) → 사용자 선택을 manager-git에 parameter로 전달

2. **`/moai plan --worktree`** (skill, orchestrator):
   - 현재: `moai worktree new <SPEC-ID>` 호출
   - 변경: 호출 직전 BODP → AskUserQuestion → `moai worktree new <SPEC-ID> --base <chosen>` (W7-T03 신규 flag)

3. **`moai worktree new <SPEC-ID>`** (CLI):
   - 현재: 현재 HEAD에서 worktree 생성
   - 변경: default base = `origin/main`. `--from-current` flag로 기존 동작 opt-out. AskUserQuestion 호출 안 함 (HARD orchestrator-only). 호출 시 자동 audit trail 기록.

**Library design**:
- `internal/bodp/relatedness.go` — 3-signal check 함수 (3개 entry에서 공유)
- `internal/bodp/audit_trail.go` — `.moai/branches/decisions/` writer

**Cognitive load**: 사용자는 **새 명령어를 학습할 필요 없음**. 기존 `/moai plan --branch` 흐름이 자동으로 더 안전해짐.

---

## 11. Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Pre-push hook 느림 | progress streaming + `--no-verify` 보존 + 점진 적용 (warning-only mode 1주) |
| Watch loop 토큰 소비 | 30s polling + 30분 timeout + manual abort flag |
| Auto-fix wrong patch | 모든 patch는 새 commit + AskUserQuestion 강제 + max 3 |
| Branch protection admin lockout | `enforce_admins: false` 시작 |
| Worktree guard false positive | `.gitignore` 항목 비교 제외 |
| i18n validator 30s 초과 | scope을 `_test.go` 의존 literal로만 한정 + caching |
| BODP 잘못된 권장 | 항상 사용자 최종 confirm + "Other" 옵션 |

---

## 12. Verification Plan

각 Wave 종료 시:
- [ ] `make ci-local` 통과
- [ ] 신규 추가 파일은 `internal/template/templates/`에 mirror (Template-First)
- [ ] `make build` 후 embedded.go 갱신 확인
- [ ] `go test ./...` 전체 통과
- [ ] 16-language neutrality 검증 (T1 / T7만 해당)
- [ ] Conventional Commits + 🗿 MoAI co-author trailer

전체 SPEC 종료 시:
- [ ] 5-PR sweep replay (acceptance.md 시나리오) 모두 통과
- [ ] 현재 세션 replay (BODP "main 분기" 권장) 통과
- [ ] CLAUDE.local.md §18 업데이트 머지 확인
- [ ] PR auto-watch + auto-fix loop이 새 PR에서 동작 확인

---

## 13. Out of Plan (Follow-up SPECs)

- 16개 언어 전체 i18n validator (현재 T7는 Go만)
- BODP를 raw `git checkout -b`도 가로채는 hook (현재 opt-in)
- Auxiliary workflow의 fix 시도 (현재는 disable만)
- T6 root cause fix (Anthropic upstream 측 — 본 SPEC은 guard만)
- Multi-repo branch protection 일괄 적용 (현재는 단일 repo)

---

Version: 0.1.0
Status: draft
Last Updated: 2026-05-05
