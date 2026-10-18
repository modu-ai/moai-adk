auditor-model: claude-sonnet-5-5

# Sync audit, card t1858, iteration 2 (delta re-audit)

verdict: FAIL
audited_sha: 16ba65ac9209789ae9b9560c84a4dc5cd7afa9b2

Overall Verdict: **FAIL**. Class B card, no SPEC. Tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1858`, branch WT-release-checks-ssot, HEAD 16ba65ac9.

The required backend `codex` returned `verdict: fail`, so the overall verdict cannot be PASS. I also independently confirmed its P1 finding (an unguarded CI path), and reproduced its P2 finding with my own probes.

## Scores (harmonic-style; must-pass Functionality and Security)

| Dimension | Score | Verdict | Basis |
|---|---|---|---|
| Functionality (40%) | 0.70 | FAIL | F1 is not closed end to end: the pin tests do not run on the PR shape that shortens the list (F5). Empty or unrecognised entries are dropped silently (F6). |
| Security (25%) | 0.60 | FAIL | The release gate can still be weakened to a lone green `Lint` through an SSoT-only PR. |
| Craft (20%) | 0.85 | PASS | All commands below are green. The new tests are sensitive (shown below). The awk handles CRLF, comments, quotes, and duplicates. |
| Consistency (15%) | 0.85 | PASS | `ptr` helper is intentional. No check-name literal in scripts/. Existing release tests unchanged. |

Must-pass firewall: Functionality and Security are both below threshold, so overall FAIL.

## Iteration history

- Iteration 1: FAIL @ f9458ae18. F1 BLOCKING (shortened `release_gate` lets a tag through on a lone green Lint); F2 SHOULD-FIX (CRLF, trailing comment, single quote, duplicate name, duplicate key); F3 MINOR (stale `CI_REQUIRED_CHECKS` wording); F4 MINOR (weak subset test).
- Iteration 2 (this file): FAIL @ 16ba65ac9. F2, F3, F4 closed. F1 only partly closed: the pins are real but a CI path-filter hole (F5) lets the exact attack PR skip them. One new SHOULD-FIX (F6).

## Defects

| Id | Severity | Status | Location | Description | Required fix |
|---|---|---|---|---|---|
| F1 | BLOCKING | PARTIALLY CLOSED, residual is F5 | internal/config/required_checks_release_gate_test.go, internal/template/release_ci_gate_ssot_test.go | Both pins fail correctly on a shortened list when they run (proved below). | See F5. |
| F5 | BLOCKING | NEW | .github/test-input-filters.yml (go_code list) | `.github/required-checks.yml` is not in `go_code`. A PR that edits only that file classifies as "no Go inputs", the real `Test (ubuntu-latest)` job does not run, and `test-skip-marker` (same check name) reports green. Neither pin runs, so a PR shortening `release_gate` to `- Lint` merges and the script alone then accepts a lone green Lint. | Add `'.github/required-checks.yml'` to `go_code` with an evidence comment, and pin it in internal/template/test_input_filter_parity_test.go (the existing filter guard). Defense in depth: a count floor in release.sh (`>= 3` names) names no check and passes TestNoHardcodedContexts, but it does not stop a rename, so it is optional. |
| F6 | SHOULD-FIX | NEW | scripts/release.sh:267-278 | An entry of `-` (no trailing space), `- ""`, or `- ` yields an empty name. The awk drops or skips it silently and the gate demands fewer checks (probes below: `Lint` only, rc=0). Combined with F5 it widens the shortening path. | Make the awk exit 1 with "empty release_gate entry" for an empty name, and refuse any line inside the block that is not a `- name` entry or a comment. Add rows for them in TestReleaseScriptCIGateSSoTFailsClosed. |
| F2 | SHOULD-FIX | CLOSED | scripts/release.sh:263-275 | CRLF stripped. Trailing comment, single quote, duplicate name and duplicate key die with `malformed release_gate in <file>: <reason>`. | none |
| F3 | MINOR | CLOSED | scripts/release.sh:235,306 | The die text now cites `.github/required-checks.yml`. The comment at line 242 still names the real variable `CI_REQUIRED_CHECKS`, which is accurate. | none |
| F4 | MINOR | CLOSED | required_checks_release_gate_test.go | Subset test replaced by an exact three-name pin. | none |
| F7 | MINOR, optional | NEW | scripts/release.sh:235 | One comment line is over-long after the edit. | Rewrap. |

## 1. Claim

1. F2, F3, F4 are closed and the regression checks hold.
2. F1 is closed by two CI-enforced pins. **Rejected**: the pins do not run on an SSoT-only PR (F5).
3. The awk is fail-closed on malformed input. **Partly rejected**: empty names are silently dropped (F6).

## 2. Evidence (commands run in this run, tree 16ba65ac9)

### 2a. audit-plan (verbatim, trimmed)

```
$ moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1858
"config_status": "ok", "model": "codex",
backends: claude gate off, codex gate required, glm gate off
"cross_model_active": true, "cross_model_required": true, "enforced_required": ["codex"]
```

### 2b. Regression checks

```
$ grep -c 'Test (ubuntu-latest)' scripts/release.sh
0
$ git diff --quiet 3642a3a9b HEAD -- internal/config/required_checks_test.go && echo UNCHANGED
UNCHANGED
$ git diff 3642a3a9b HEAD --stat -- internal/template/release_*test.go
 internal/template/release_ci_gate_ssot_test.go | 154 +++++++++++++++++++++++++   (the only file; other release_* tests untouched)
$ git diff 3642a3a9b HEAD -- scripts/release.sh | grep -n 'Release Verify'
(no output: Release Verify legs untouched)
$ go test ./internal/config -count=1
ok  	github.com/modu-ai/moai-adk/internal/config	8.538s
$ go test ./internal/template -run 'TestReleaseScript|TestTestInputFilters' -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	35.305s
$ gofmt -l internal/config internal/template        (no output)
$ golangci-lint run ./internal/config/ ./internal/template/
0 issues.
$ scripts/ci-mirror/validate-required-checks.sh
✅ All validations passed
$ git status --short        (clean)
```

### 2c. Sensitivity of the pins (actually run, on a scratch copy, not the tree)

`git archive HEAD` extracted to `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/3485ddff-4832-470a-9a17-ee4dac7ad178/scratchpad/copy2`; in the copy only, `release_gate` was shortened to `- Lint`.

```
$ go test ./internal/config -run ReleaseGate -count=1
--- FAIL: TestRequiredChecks_ReleaseGateSubsetOfMain
    release_gate lacks the pinned check "Test (ubuntu-latest)": [Lint]
    release_gate lacks the pinned check "Build (linux/amd64)": [Lint]
    release_gate must hold exactly [Lint Test (ubuntu-latest) Build (linux/amd64)], got [Lint]
FAIL
$ go test ./internal/template -run TestReleaseScriptRealSSoTDemandsEveryPinnedCheck -count=1
--- FAIL: .../test_missing   exit code = 0, want 1
--- FAIL: .../build_missing  exit code = 0, want 1
--- FAIL: .../only_lint_green exit code = 0, want 1
FAIL
```

Both pins are sensitive when they run. The unshortened copy passed both (first attempt, before the edit applied): `ok config 0.328s`, `ok template 4.407s`.

### 2d. The hole: the pins do not always run (read from the tree)

```
$ grep -n 'required-checks' .github/test-input-filters.yml
(no match)
go_code list: '**/*.go', go.mod, go.sum, Makefile, 4 workflow/filter files,
  '.github/branch-protection.json.gtmpl', '.moai/**', templates, .claude/*, scripts/ci*, scripts/release.sh, ...
.github/workflows/ci.yml:666  test-skip-marker: name: Test (${{ matrix.os }}), if: needs.detect.outputs.go_code != 'true'
```

`.github/required-checks.yml` matches no `go_code` pattern. A PR touching only that file gets the skip-marker stub as the required `Test (ubuntu-latest)`. I did not run GitHub Actions; this is derived from the committed filter and workflow definitions (also the codex reading, below).

### 2e. awk probes (the exact awk extracted from release.sh lines 263-274 and run on crafted files)

| Input | Outcome | Class |
|---|---|---|
| real SSoT | 3 names | SAFE |
| tab indentation | names read | SAFE |
| quoted name containing `#` (`"Lint #1"`) | refused (trailing-comment) | SAFE (over-strict, fails closed) |
| name with trailing spaces | trimmed, correct names | SAFE |
| entry `-` (no space) | silently skipped, `Lint` only, rc=0 | UNSAFE (F6) |
| entry `- ` (trailing space) | empty name dropped, `Lint` only, rc=0 | UNSAFE (F6) |
| entry `- ""` | empty name dropped, rc=0 | UNSAFE (F6) |
| `release_gate:` then comment line then list | names read | SAFE |
| list item at column 0 | block ends, zero names, die | SAFE |
| `release_gate_x:` key | no match, zero names, die | SAFE |
| flow list `release_gate: [Lint, Test]` | zero names, die | SAFE |
| `release_gate: # c` | names read | SAFE |
| indented `release_gate:` key | zero names, die | SAFE |
| `-Lint` (no space) | skipped, zero names, die | SAFE |
| unterminated quote `"Lint` | name `"Lint` demanded, never reported, die later | SAFE |
| anchor `&a Lint` | name `&a Lint` demanded, die later | SAFE |
| nested deeper item | extra name demanded | SAFE |
| blank line inside list | skipped, rest read | SAFE |
| `---` document marker mid-list | block ends, later names dropped, rc=0 | UNSAFE (unrealistic, same class as F6) |
| duplicate `release_gate:` key | refused | SAFE |
| CRLF file | names read | SAFE |

### 2f. Required backend result (verbatim fields from `mcp__moai__audit_multi`, card_id t1858, project_root as above, no `gates` argument)

```
overall_verdict: fail
gate_unmet: (absent)
plan_source: config
residual_risk_note: required-backend FAIL: codex
audit_receipt: rcpt-afa2e6a2fb9d201f141c43b7
per_backend_verdicts: backend=codex gate=required verdict=fail
build_lag: installed binary built from ee981a932, an ancestor of HEAD 16ba65ac9
```

Codex findings verbatim:

> [P1] 설정 파일만 변경하면 새 pin 테스트가 실행되지 않아 F1 우회가 남음 — internal/config/required_checks_release_gate_test.go:23
> 신뢰도: 높음. `.github/test-input-filters.yml`에는 `.github/required-checks.yml`과 매칭되는 규칙이 없어. CI 정의상 이 경우 실제 테스트 대신 같은 이름의 성공 skip-marker가 실행돼.
> HEAD `16ba65ac9`에서 축소 목록은 pin 테스트를 직접 실행하면 실패했지만, 스크립트는 성공한 `Lint` 응답 하나로 종료 코드 0과 `CI gate: verified`를 출력했어. 따라서 Go 테스트에 집합을 고정하는 것만으로는 충분하지 않아.
> 권고: 해당 YAML을 공통 CI 입력 필터와 필터 보호 테스트에 추가해. 실제 GitHub Actions 실행은 미측정이며, 경로 매칭과 실행 조건을 로컬에서 확인했어.
>
> [P2] 빈 이름과 인식하지 못한 항목을 조용히 버려 불완전한 목록을 승인함 — scripts/release.sh:267
> 신뢰도: 높음. `Lint` 뒤에 `- ""`, 값 없는 `-`, 또는 예상하지 않은 매핑을 넣은 스크립트 복사본을 각각 실행했어. 모두 `Lint`만 요구하며 종료 코드 0과 `CI gate: verified`를 출력했어. 정상 세 항목 설정은 동일한 응답에서 거부됐어.
> 원인: awk는 매칭되는 줄만 수집하고, 277행은 빈 이름을 버리며, 마지막에는 전체 목록이 비었는지만 검사해.
> 권고: 허용한 목록 문법 밖의 항목과 빈 이름을 즉시 거부하고, 축소 목록도 런타임에서 거부하도록 완전성 검증을 추가해. 위 입력을 직접 거부하는 회귀 테스트가 필요해.

The P1 matches my reading of the filter (2d). The P2 empty-name part matches my probes (2e). Its "unexpected mapping" variant I did not reproduce (my probes of a nested item and an anchor failed closed).

## 3. Baseline-attribution

- Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1858`, HEAD `16ba65ac9209789ae9b9560c84a4dc5cd7afa9b2`, clean status.
- Delta: `git diff 5f1e6583f..HEAD` (5 files).
- Judging build: installed moai built from `ee981a932`, an ancestor of HEAD (the tool's own build_lag note). The audit-plan and audit_multi outputs above come from that older build. The go, gofmt, golangci-lint and shell results do not depend on it.
- The scratch copy is a `git archive HEAD` extraction; the tree itself was not edited.

## 4. Gaps

- GitHub Actions was not run: F5 comes from reading `.github/test-input-filters.yml` and ci.yml, not from an actual CI run.
- Codex's "unexpected mapping" variant of F6 was not reproduced.
- The claude and glm backends are gate-off per the plan; no second opinion beyond codex.
- Initial attempt to create a scratch copy with an `rm -rf` was denied by the permission system; I used a fresh directory instead. No effect on results.
- `rlsRequireTools(t, "jq")`: the template pin is skipped where `jq` is absent. I ran it with jq present; CI presence of jq was not measured.

## 5. Residual-risk

- Until F5 is fixed, an SSoT-only PR can shorten `release_gate` and merge on the skip-marker, after which a lone green `Lint` lets a tag through.
- After F5 and F6, a deliberate two-place edit (SSoT plus the Go pin) remains a legitimate way to change the demanded set; that is by design.
- Admin-bypass merges skip the pins.

Addendum: a second single-backend adversarial `codex_audit` (receipt rcpt-4f0500c01469bec3f88471c2) also returned `verdict: fail` with the same two findings: P1 (`.github/test-input-filters.yml:58`, the SSoT-only PR skips the pins; codex says it evaluated the filter with picomatch, giving `go_code=false`) and P2 (`scripts/release.sh:271`, empty/bare-`-`/whitespace entries dropped, script exits 0 on a shortened list). Hosted CI was not run by codex either.

AUDIT-VERDICT: FAIL spec=t1858 receipts=rcpt-afa2e6a2fb9d201f141c43b7,rcpt-4f0500c01469bec3f88471c2
