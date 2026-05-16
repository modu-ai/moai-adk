---
id: SPEC-V3R4-CI-INFRA-FIX-001
title: "CI infrastructure 3-defect bundle fix — detect-language SIGPIPE + spec-status-sync 403 + checkout fetch-depth 0"
version: "0.1.1"
status: draft
created: 2026-05-16
updated: 2026-05-16
author: manager-spec
priority: P1
phase: "v3.0.0 R4 — Foundation Cleanup / CI Infrastructure"
module: ".github/actions/detect-language, .github/workflows"
lifecycle: spec-anchored
tags: "v3r4, ci-infra, github-actions, sigpipe, fetch-depth, spec-status-sync, release-readiness, foundation"
issue_number: null
related_specs:
  - SPEC-V3R4-LINT-SPECID-GREP-FIX-001
  - SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002
  - SPEC-V3R4-HARNESS-NAMESPACE-001
depends_on: []
breaking: false
bc_id: []
related_theme: "Foundation Cleanup — CI Infrastructure for v3.0.0-rc1 release-readiness"
target_release: v3.0.0-rc1
---

# SPEC-V3R4-CI-INFRA-FIX-001 — CI infrastructure 3-defect bundle fix

## HISTORY

| Version | Date       | Author          | Description |
|---------|------------|-----------------|-------------|
| 0.1.1   | 2026-05-16 | manager-develop | Run-phase hotfix. design.md D-1 A2 (awk) 가 SIGPIPE 미해결로 확인됨 (PR #955 4 review workflow broken-pipe exit 2 재현). A3 (find -print -quit 단일-step) 로 hotfix 적용. AC-CIIF-001 재검증 예정. |
| 0.1.0   | 2026-05-16 | manager-spec    | 초기 draft. 3 CI infrastructure defect (A: detect-language SIGPIPE / B: spec-status-sync 403 / C: checkout fetch-depth 0) 를 단일 SPEC scope으로 묶음. AC-CIIF-001~004 binary (CI run-N 검증 + `moai spec lint --strict` 0/0 유지). v3.0.0-rc1 release tagging 최종 precondition. LSGF-001 PR #948의 `GITHUB_ACTIONS` env skip 우회 영구 제거. |

---

## 1. Goal

GitHub Actions CI infrastructure의 3개 결함을 **단일 SPEC scope** 으로 묶어 영구 해소하여 v3.0.0-rc1 release tagging 의 최종 precondition을 충족한다. 본 SPEC은 LSGF-001 + FOLLOWUP-002 lifecycle COMPLETE 후 main HEAD `9be0ef03b` 상태에서 `moai spec lint --strict` `✓ No findings (0/0)` 기준점을 회귀시키지 않으면서 CI workflow YAML + composite action shell script + GitHub Actions permissions 세 layer를 동시 개선한다.

### 1.1 배경 — 3 Sub-fix Root Causes

#### Fix A — detect-language SIGPIPE (4건 incident)

**Target**: `.github/actions/detect-language/action.yml` composite action의 shell script.

**Defect**: SPEC-V3R4-HARNESS-NAMESPACE-001 plan PR #944 처리 중 CI 4건이 SIGPIPE (exit 141) 로 실패. 해당 composite action은 다음 shell snippet 을 사용한다:

```yaml
LANG_COUNT=$(find . -name "*.go" -o -name "*.py" -o -name "*.ts" -o -name "*.js" | \
  grep -v node_modules | grep -v vendor | grep -v ".git" | \
  head -1 | sed 's/.*\.//' | sort | uniq -c | sort -rn | head -1 | \
  awk '{print $2}')
```

**Root cause**: `find ... | grep ... | head -1` 파이프 체인에서 `head -1` 이 첫 줄 읽은 후 stdin 을 close 한다. find/grep 은 계속 출력을 시도하다가 broken pipe 시그널 (SIGPIPE) 을 받음. `set -o pipefail` (GitHub Actions bash default) 가 활성화된 경우 SIGPIPE 가 exit 141 로 전체 파이프라인을 실패로 처리. 결과는 비결정적 — find 가 첫 결과를 빠르게 산출하면 SIGPIPE 회피, 느리면 발생.

**Evidence**: NAMESPACE-001 PR #944 4 CI run failures (project_v3r4_namespace_001_plan_merged.md memory 기록). 4건 모두 detect-language step 에서 exit 141.

#### Fix B — spec-status-sync 403 (push permission 결함)

**Target**: `.github/workflows/spec-status-auto-sync.yml`.

**Defect**: PR 머지 후 SPEC frontmatter status 자동 sync 시도 시 `git push origin main` 이 403 Forbidden 으로 실패. NAMESPACE-001 PR #944 force-push 후 동일 incident 관측됨 (memory `project_v3r4_namespace_001_plan_merged.md` 의 "spec-status-sync 403 push 권한" 기록).

**Root cause**: workflow 가 `permissions:` 블록을 명시하지 않아 GITHUB_TOKEN 의 default scope (workflow 단위 `contents: read`) 만 사용. `git push origin main` 은 `contents: write` 가 필요. Branch protection rule on `main` (CLAUDE.local.md §18.7) 가 활성화된 경우 추가로 PR-via-bot 우회 경로가 필요할 수도 있음 — run-phase 에서 정확한 권한 stack 검증 후 minimal-privilege 솔루션 채택.

**Evidence**: spec-status-auto-sync.yml 현재 상태 — `permissions:` 블록 부재. workflow 의 마지막 step `git push origin main` 이 직접 main 에 push.

#### Fix C — checkout fetch-depth 0 (LSGF-001 우회 영구 제거)

**Target**: `.github/workflows/ci.yml` (5 checkout step) + 후속 검토 대상 (codeql.yml, test-install.yml 등 — 본 SPEC scope 결정은 design.md D-3 참조).

**Defect**: `actions/checkout@v4` (또는 v5) 가 명시적 `fetch-depth` 미지정 시 default `fetch-depth: 1` (shallow clone) 을 사용. `internal/spec/drift.go` `getGitImpliedStatus` walker 는 `git log --grep=<SPEC-ID>` 로 full git history 가 필요. shallow clone 환경에서는 SPEC commits 가 부재 → walker 실패. LSGF-001 PR #948 에서 `TestGetGitImpliedStatus_HARNESS001Resolution` 이 `GITHUB_ACTIONS=true` env 일 때 skip 되도록 workaround 적용됨 (`internal/spec/drift_specid_grep_test.go:23-27`).

**Root cause**: shallow clone 은 GitHub Actions checkout v3+ default. drift walker 는 full history 가 contract 의 일부.

**Evidence**:
- `internal/spec/drift_specid_grep_test.go:21-27`:
  ```go
  if os.Getenv("GITHUB_ACTIONS") == "true" {
      t.Skip("requires full git history; CI uses actions/checkout@v4 default fetch-depth: 1 (shallow). " +
          "Word-boundary logic is fully covered by TestGetGitImpliedStatus_SPECIDWordBoundary 5 sub-cases. " +
          "Follow-up: SPEC-V3R4-CI-INFRA-FIX-001 to set fetch-depth: 0 for full-history tests.")
  }
  ```
- ci.yml: 5개 `actions/checkout@v5` step 모두 `with: fetch-depth: 0` 미지정 (test job × 1, test-integration × 1, lint × 1, build × 1, constitution-check × 1).
- 참조: 다른 workflow 는 이미 `fetch-depth: 0` 설정됨 (claude-code-review.yml, release.yml, spec-status-auto-sync.yml).

### 1.2 v3.0.0-rc1 Release Gate

본 SPEC은 v3.0.0-rc1 태그 발행의 **최종 precondition**. LSGF-001 + FOLLOWUP-002 가 lint correctness 를 확립했고, 본 SPEC은 CI infrastructure 신뢰성 (SIGPIPE 0 / 403 0 / shallow-clone skip 제거) 을 확립한다. 본 SPEC closure 후 `./scripts/release.sh v3.0.0-rc1` (CLAUDE.local.md §18.8) 호출 가능.

---

## 2. Scope

### 2.1 In Scope

- `.github/actions/detect-language/action.yml` shell script SIGPIPE remediation.
- `.github/workflows/spec-status-auto-sync.yml` permissions 블록 + trigger 정책 정정.
- `.github/workflows/ci.yml` 5개 checkout step 의 `fetch-depth: 0` 명시화.
- 후속 정리: `internal/spec/drift_specid_grep_test.go` 의 `GITHUB_ACTIONS` env skip guard 제거 (run-phase 에서 Wave 3 task).
- 본 SPEC의 plan/run/sync 3-PR cycle.

### 2.2 Out of Scope

- `internal/spec/drift.go` walker 로직 자체 변경 (LSGF-001 territory).
- 새 lint rule 또는 새 GitHub Actions workflow 생성.
- 다른 CI workflow 전체 audit (claude.yml, codex-review.yml 등 — 본 SPEC은 ci.yml + detect-language + spec-status-auto-sync 3 파일에 한정).
- GoReleaser, label-sync, release-drafter, docs-i18n-check workflows 검토.
- `actions/checkout@v4` → `v5` 일괄 upgrade.
- Branch protection rule (CLAUDE.local.md §18.7) 변경.
- v3.0.0-rc1 release tag push (별도 release 작업).

---

## 3. Requirements (EARS Format)

### REQ-CIIF-001 (Ubiquitous)

The `.github/actions/detect-language/action.yml` composite action SHALL execute its shell pipeline without producing SIGPIPE exit code (141) under any GitHub Actions runner condition (ubuntu-latest / macos-latest / windows-latest with bash).

**Acceptance**: AC-CIIF-001

### REQ-CIIF-002 (Event-Driven)

**WHEN** a pull request is closed with `merged == true`, **THE** `spec-status-auto-sync` workflow SHALL successfully push the resulting SPEC frontmatter status updates to `origin/main` without producing HTTP 403 authorization failures.

**Acceptance**: AC-CIIF-002

### REQ-CIIF-003 (Ubiquitous)

All `actions/checkout` steps in `.github/workflows/ci.yml` SHALL explicitly specify `fetch-depth: 0` so that subsequent `git log` / `git rev-list` based test cases have access to full git history.

**Acceptance**: AC-CIIF-003

### REQ-CIIF-004 (State-Driven)

**WHILE** the CI runs on a GitHub-hosted runner with `GITHUB_ACTIONS=true`, **THE** test `TestGetGitImpliedStatus_HARNESS001Resolution` in `internal/spec/drift_specid_grep_test.go` SHALL execute to completion (PASS) without skipping via the `GITHUB_ACTIONS` env probe.

**Acceptance**: AC-CIIF-003

### REQ-CIIF-005 (Unwanted)

**IF** a CI workflow step needs `git log --grep` over full project history, **THEN** that workflow's checkout step SHALL NOT rely on the default shallow clone (`fetch-depth: 1`).

**Acceptance**: AC-CIIF-003

### REQ-CIIF-006 (Constraints — least privilege)

The `spec-status-auto-sync` workflow `permissions:` block SHALL grant the minimum scope required for the push operation (`contents: write`), and SHALL NOT grant unrelated broader scopes (`actions: write`, `id-token: write`, `pull-requests: write` 등) unless the run-phase analysis documents a concrete necessity.

**Acceptance**: AC-CIIF-002

### REQ-CIIF-007 (Ubiquitous — regression baseline)

After all three sub-fixes are merged, `moai spec lint --strict` SHALL continue to output `✓ No findings (0/0)` (zero errors, zero warnings) — i.e., the lint baseline established by LSGF-001 + FOLLOWUP-002 SHALL NOT regress.

**Acceptance**: AC-CIIF-004

---

## 4. Exclusions (What NOT to Build)

1. **walker code 변경 금지** — `internal/spec/drift.go` `getGitImpliedStatus` 로직은 LSGF-001 으로 영구 fix 완료. 본 SPEC은 CI 환경만 정정.
2. **새 workflow 신설 금지** — 3 파일 (`detect-language/action.yml`, `spec-status-auto-sync.yml`, `ci.yml`) 만 수정. 새 workflow YAML 생성 없음.
3. **`actions/checkout` 버전 업그레이드 금지** — `@v4` 또는 `@v5` 현행 버전 유지. `fetch-depth: 0` 추가만 수행.
4. **광범위 permissions 부여 금지** — REQ-CIIF-006 에 따라 minimum scope (`contents: write`) 만. 다른 scope 는 documented necessity 없이 추가 금지.
5. **drift walker 의 shallow-clone fallback 도입 금지** — `if !fullClone { return early }` 같은 conditional 추가는 본 SPEC 범위 외. CI 환경을 fix 하여 walker contract 를 만족시키는 방향만.
6. **다른 SPEC frontmatter 수정 금지** — 본 SPEC 자체 spec.md 외 SPEC frontmatter 변경 0건.
7. **release tag push 금지** — `v3.0.0-rc1` 태그는 본 SPEC closure 후 별도 작업.
8. **Branch protection rule 변경 금지** — CLAUDE.local.md §18.7 에 정의된 main + release/* 보호 규칙은 불변.
9. **detect-language action 의 언어 detection 로직 자체 변경 금지** — SIGPIPE remediation 만. find/grep/awk pipeline 의 detection 의도 (16 supported languages 첫 1건 추출) 는 보존.

---

## 5. Constraints

- **Code language**: shell script (bash), YAML (GitHub Actions schema).
- **Comment language**: Korean (per `.moai/config/sections/language.yaml` `code_comments: ko`).
- **Schema**: 12-field canonical frontmatter (per `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT).
- **Test discipline**: YAML/shell change는 TDD 직접 적용 어려움 → **evidence-driven verification** (CI run N=3 consecutive 검증). DDD 모드로 진행 (manager-develop with cycle_type=ddd).
- **Performance budget**: CI 추가 비용 미미. fetch-depth: 0 은 ~2-3s 추가 (full clone vs shallow) — 무시 가능.
- **MX tag**: drift_specid_grep_test.go 의 skip guard 제거 시 주변 주석 정리 + `@MX:NOTE` 갱신 (LSGF-001 follow-up 완료 명시).
- **PR merge strategy**: CLAUDE.local.md §18.3 — plan/run/sync 모두 `--squash`.
- **Branch base**: BODP signals `¬a ¬b ¬c` (no depends_on / no co-located worktree / no open PR head on current branch) → base = `origin/main`. 현재 branch `feat/SPEC-V3R4-CI-INFRA-FIX-001-plan` 는 origin/main 에서 fork 완료.

---

## 6. Risks

### Risk-1: SIGPIPE remediation 이 detection 정확도 변경

**상황**: detect-language action 의 `head -1` 제거 또는 대체 (sed/mapfile) 가 첫 매칭 동작을 변경하여 다른 언어를 잘못 detect.

**Mitigation**: design.md D-1 에서 선택한 솔루션은 detection semantics (첫 1건 추출) 를 보존. run-phase 에서 16 supported languages 각각에 대해 fixture 기반 검증 task 수행.

### Risk-2: permissions 변경이 다른 secret/action 사용을 break

**상황**: `permissions:` 명시화가 workflow 의 다른 step 에서 의존하던 implicit scope 를 제거.

**Mitigation**: spec-status-auto-sync.yml 은 4 step 만 (checkout / setup-go / make build / sync 단일 step) 사용. 각 step 의 required scope 명시적 audit 후 union 적용.

### Risk-3: fetch-depth: 0 가 CI runtime 증가

**상황**: 대규모 모노레포에서 full clone 이 수십 초 추가.

**Mitigation**: moai-adk-go 는 중소 규모 (~1000 commits). full clone 추가 2-3 초 무시 가능. design.md D-3 에서 cache strategy 미변경.

### Risk-4: CI infrastructure 변경이 다른 workflow 회귀 유발

**상황**: ci.yml 의 fetch-depth: 0 변경이 다른 workflow (codeql.yml, test-install.yml) 와 분리되어 inconsistent state 노출.

**Mitigation**: 본 SPEC은 ci.yml 만 수정 (REQ-CIIF-003 명시). 다른 workflow 검토는 design.md D-3 에서 explicit scope decision. 별도 follow-up SPEC 후보로 기록.

### Risk-5: drift_specid_grep_test.go skip guard 제거 후 fetch-depth: 0 가 missing 인 다른 workflow 에서 동일 test 실패

**상황**: skip guard 제거 (Wave 3 task) 후 test 가 ci.yml 외 다른 workflow 에서 실패.

**Mitigation**: `TestGetGitImpliedStatus_HARNESS001Resolution` 는 `internal/spec/` package 의 standard `go test` 로만 실행. ci.yml 의 test job 만 영향. 다른 workflow (claude-code-review.yml 등) 는 `go test` 미실행.

---

## 7. References

- Memory entry (trigger): `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/project_v3r4_ci_infra_fix_001_entry.md`
- Memory entry (LSGF-001 complete): `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/project_v3r4_lsgf_001_complete.md`
- Memory entry (FOLLOWUP-002 complete): `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/project_v3r4_followup_002_complete.md`
- Memory entry (NAMESPACE-001 incident): `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/project_v3r4_namespace_001_plan_merged.md` (4 SIGPIPE + 403 incident)
- File: `.github/actions/detect-language/action.yml` (37 lines, 1718 bytes — full read in plan-phase)
- File: `.github/workflows/spec-status-auto-sync.yml` (109 lines)
- File: `.github/workflows/ci.yml` (217 lines, 5 checkout steps)
- File: `internal/spec/drift_specid_grep_test.go:21-27` (CI skip guard — Wave 3 removal target)
- File: `internal/spec/drift.go:120-180` (walker — 무수정 대상, 참조 only)
- Related SPECs:
  - `.moai/specs/SPEC-V3R4-LINT-SPECID-GREP-FIX-001/` (walker word-boundary fix — 본 SPEC 직접 trigger)
  - `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002/` (17 drift 해소 — 본 SPEC 이전 wave)
  - `.moai/specs/SPEC-V3R4-HARNESS-NAMESPACE-001/` (NAMESPACE incident — 본 SPEC 의 root trigger)
- Rules:
  - `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical schema SSOT
  - `.claude/rules/moai/workflow/spec-workflow.md` — SPEC phase discipline (plan-in-main + squash)
  - `CLAUDE.local.md §18.3` — merge strategy (feature → squash, release → merge commit)
  - `CLAUDE.local.md §18.7` — branch protection rule (read-only reference)
- External:
  - GitHub Actions docs: <https://docs.github.com/en/actions/security-for-github-actions/security-guides/automatic-token-authentication#permissions-for-the-github_token> (permissions scope reference)
  - actions/checkout docs: <https://github.com/actions/checkout#fetch-all-history-for-all-tags-and-branches> (fetch-depth: 0 rationale)
  - POSIX SIGPIPE rationale: <https://www.gnu.org/software/coreutils/manual/html_node/head-invocation.html> (head behavior)

---

## 8. AC ↔ REQ Mapping (summary)

| AC ID         | REQ ID(s)                       | Verification scope |
|---------------|---------------------------------|--------------------|
| AC-CIIF-001   | REQ-CIIF-001                    | detect-language action 의 CI 3-run SIGPIPE 부재 확인 |
| AC-CIIF-002   | REQ-CIIF-002, REQ-CIIF-006      | spec-status-auto-sync workflow run success on merged PR |
| AC-CIIF-003   | REQ-CIIF-003, REQ-CIIF-004, REQ-CIIF-005 | ci.yml checkout fetch-depth: 0 + skip guard 제거 후 test PASS |
| AC-CIIF-004   | REQ-CIIF-007                    | `moai spec lint --strict` 0/0 유지 (regression baseline) |

상세 시나리오 및 verification commands는 `acceptance.md` 참조.

### 8.1 Sub-fix ↔ REQ/AC Cross-reference

| Sub-fix | REQ(s) | AC(s) | Target file |
|---------|--------|-------|-------------|
| Fix A — SIGPIPE | REQ-CIIF-001 | AC-CIIF-001 | `.github/actions/detect-language/action.yml` |
| Fix B — 403 | REQ-CIIF-002, REQ-CIIF-006 | AC-CIIF-002 | `.github/workflows/spec-status-auto-sync.yml` |
| Fix C — fetch-depth | REQ-CIIF-003, REQ-CIIF-004, REQ-CIIF-005 | AC-CIIF-003 | `.github/workflows/ci.yml` + `internal/spec/drift_specid_grep_test.go` |
| Cross-cutting | REQ-CIIF-007 | AC-CIIF-004 | (regression baseline) |
