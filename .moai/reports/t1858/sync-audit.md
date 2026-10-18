auditor-model: claude-sonnet-5-5

# Sync audit: card t1858 (Class B, no SPEC)

verdict: FAIL
audited_sha: f9458ae18998afe5d4e4a2d9deea8b2138dd2f96

Overall: FAIL. The required backend `codex` returned `fail`, so the overall verdict cannot be PASS whatever my own reading says. My own mechanical probes independently reproduce the codex finding (F1).

Scores (0.00-1.00): Functionality 0.80, Security 0.60, Craft 0.85, Consistency 0.90. The must-pass dimension Security is below its pass bar because of F1.

## Audit plan and required backends
`moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1858`: `config_status: ok`, `cross_model_active: true`, `enforced_required: ["codex"]`, claude gate `off`, glm gate `off`, codex gate `required`.

`mcp__moai__audit_multi` (card_id t1858, project_root = the worktree, target baseBranch), digest members:
- overall_verdict: `fail`
- gate_unmet: not present in the result (`residual_risk_note`: "required-backend FAIL: codex")
- plan_source: `config`
- per_backend_verdicts: backend `codex`, gate `required`, verdict `fail`
- audit_receipt: `rcpt-b6a0f22a9eee365cd060f299`
- participant_count 1, fail_open_backends []
- build_commit `ee981a932` (an ancestor of tree HEAD; the MCP server reports "moai binary lag").

codex finding, verbatim (Korean as returned):
> [P1] 잘못되거나 축소된 목록으로 필수 CI 검사 두 개를 생략할 수 있음 — scripts/release.sh:270. 신뢰도: 높음. HEAD `f9458ae18`의 스크립트 복사본으로 재현했어. 블록 목록에 `Lint`만 넣거나 세 번 반복하면, 성공한 `Lint` 응답 하나만으로 종료 코드 0과 `CI gate: verified`가 나와. 중첩 매핑과 중복 `release_gate` 키도 같은 우회가 가능해. 정상 세 이름 목록은 동일한 응답에서 거부됐어. 원인: awk가 YAML 구조를 검증하지 않고 항목을 수집하며, 이후에는 비어 있는지만 확인해. 추가된 설정 테스트도 비어 있지 않은 부분집합을 허용해. 권고: 최상위 문자열 목록을 구조적으로 파싱하고 중복 키·중복 이름·필수 이름 누락을 거부해. 위 사례를 `Lint` 단독 응답으로 실행하는 회귀 테스트를 추가해.

The orchestrator must write the digest to `<toplevel>/.moai/state/audit-plan-result.json` and run `moai verify audit-plan --result-file`; my verdict is not final until that check passes.

## Claim / Evidence per acceptance criterion

### A1 TestNoHardcodedContexts passes, existing test unchanged: PASS
- `git diff --stat 3642a3a9b..HEAD -- internal/config/required_checks_test.go` printed nothing (byte-unchanged). `git diff --name-only ... | grep _test.go` lists only the two NEW files `internal/config/required_checks_release_gate_test.go` and `internal/template/release_ci_gate_ssot_test.go`. No existing test was weakened.
- `go test ./internal/config -run TestNoHardcodedContexts -count=1` -> `ok  	github.com/modu-ai/moai-adk/internal/config	0.333s`
- `go test ./internal/config -count=1` -> `ok  	github.com/modu-ai/moai-adk/internal/config	4.980s`

### A2 Same three names, same error texts: PASS (with the F1 caveat on the contract)
- Diff of `scripts/release.sh`: the only changes are the `CI_REQUIRED_CHECKS=()` initialisation, the SSoT read block and its two new `die` texts. The existing `expected check run(s) not reported` text, the loop at lines 292-297 and the `log_ok` line are unchanged.
- The real SSoT block parsed by the script's own awk (extracted verbatim) yields `Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`.
- `go test ./internal/template -run 'RequiredChecks|Required|AutoMerge|ReleaseScript|CIGate|TestTestInputFilters' -count=1` -> `ok  	github.com/modu-ai/moai-adk/internal/template	29.447s` (existing release_* tests pass with no edits).
- Caveat: nothing pins the demanded SET to those three names (see F1).

### A3 Fail-closed: FAIL (partial)
Refusals work for: missing file, no `release_gate` block, empty block (driven end to end by `TestReleaseScriptCIGateSSoTFailsClosed`, which passes; the progress record also shows it was sensitivity-checked against neutered `die` guards). `--skip-ci-check` branch is untouched by the diff (the `if SKIP ... else` structure is unchanged).

Probes of the script's own awk program against crafted SSoT files (output via `cat -vet`):

| Input | awk yields | Outcome |
|---|---|---|
| flow list `release_gate: [Lint, "Test (ubuntu-latest)"]` | nothing | SAFE (refusal: zero names) |
| trailing comment `- Lint # c` | `Lint # c` | SAFE (name never matches a check run, so the gate refuses as "not reported") |
| CRLF line endings | `Lint^M`, `"Test (ubuntu-latest)"^M` | SAFE (names never match; quotes not stripped; refuses, but with a misleading "not reported" message) |
| `release_gate` nested under `branches:` | nothing | SAFE (refusal) |
| duplicate `release_gate` keys | all entries of both blocks, `Lint` twice | SAFE for dup-of-correct (stricter, harmless) |
| single-quoted entry | `'Test (ubuntu-latest)'` (quotes kept) | SAFE (never matches, refuses) |
| tab-indented, column-0 list items | `Lint` / nothing | indifferent / SAFE |
| name with spaces/parens (double-quoted) | stripped correctly | correct |
| **list reduced to `- Lint`** | `Lint` | **UNSAFE: gate demands one name; a lone green `Lint` from github-actions passes it** |
| **`Lint` repeated three times** | `Lint` x3 | **UNSAFE: same** |

So the awk cannot yield a wrong NAME that passes (wrong names over-refuse), but it yields a SHORT list that passes, and nothing checks that the list holds the three required names. The only guard is "non-empty"; the Go test asserts "non-empty subset of branches.main.contexts" (a subset allows `Lint` alone); the Go-side fixtures for the release gate carry extra check runs and cannot detect a shortened SSoT. This matches codex's finding. The attack surface is a deliberate or careless edit of a repo-controlled file in a PR, not a remote input; but the card's own contract is "same three names, fail closed", and a tag is irreversible, so a silent weakening of this gate by editing one YAML line is a real regression in safety relative to the previous hard-coded array (which was pinned by tests with 3-name fixtures).

### A4 No behavior change elsewhere: PASS
- Line 343 `for leg in "Release Verify (ubuntu-latest)" "Release Verify (macos-latest)" "Release Verify (windows-latest)"` is outside the diff (`git diff ... -- scripts/release.sh | grep Release Verify` printed nothing).
- `./scripts/ci-mirror/validate-required-checks.sh` -> `✅ All validations passed` (final line; the captured exit variable printed empty because of the pipe, the success banner is the observed result).
- Consumers: `git grep -l 'required-checks' -- '*_test.go'` -> internal/ciwatch/{classifier,coverage}_test.go, internal/cli/candidate_landing_gate_test.go, internal/config/required_checks_test.go, internal/template/auto_merge_checks_test.go, internal/template/release_ci_gate_ssot_test.go.
  - `go test ./internal/ciwatch ./internal/config ./internal/template -run 'RequiredChecks|Required|AutoMerge|ReleaseScript|CIGate|TestTestInputFilters' -count=1` -> ok, ok, ok (0.104s / 0.125s / 29.447s).
  - `go test ./internal/cli -run 'BranchProtection|CandidateLanding|RequiredChecks' -count=1` -> `ok  	github.com/modu-ai/moai-adk/internal/cli	1.105s`.
- The new top-level `release_gate` key is tolerated by the Go loader (the struct field was added; the unchanged strictness tests pass) and by the validator script.

### A5 Docs accuracy: PASS with one stale-text finding
- CHANGELOG entry matches the code (key name, struct field, subset test, relative resolution, refusal, same three names).
- Runbook (three statements) now names `release_gate` in `.github/required-checks.yml`; consistent with the code.
- Stale: `scripts/release.sh` line 235 comment (`each name in CI_REQUIRED_CHECKS below must be present among them`) and the die text at line 297 (`The repository CI workflow (CI_REQUIRED_CHECKS in this script) has not run ...`) still say the names live in the script. The error text is kept verbatim on purpose (A2), but it now points readers to a place that no longer holds the names. The CHANGELOG states the error texts are unchanged, which is accurate.

### Lint / format
- `gofmt -l internal/config internal/template` -> empty; `go vet ./internal/config ./internal/template` clean.
- `golangci-lint run ./internal/config/ ./internal/template/` -> `0 issues.`

## Baseline-attribution
Tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1858`, HEAD `f9458ae18998afe5d4e4a2d9deea8b2138dd2f96`, merge base `3642a3a9b`; every command above run in this audit against this tree. `go`/`golangci-lint` are tree-independent toolchains. The `moai` CLI used for `audit-plan` and the MCP server were built from `ee981a932`, an ancestor of HEAD (reported by the tool as build lag); their outputs describe the audit-plan surface, not repo code under test. The awk probes used the script's awk program text copied from `scripts/release.sh`, run on macOS BSD awk.

## Defect list
- F1 [BLOCKING] scripts/release.sh:~262-270 and .github/required-checks.yml `release_gate` - gate checks only "non-empty"; a list reduced to one name (e.g. only `Lint`, or one name repeated) lets a tag be cut when only that check is green. Reproduced by my awk probe (`Lint` / `Lint` x3) and by codex (exit 0, `CI gate: verified`). Required fix: pin the demanded set so a shortened SSoT cannot pass: (a) add a test asserting `release_gate` contains exactly/at least `Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)` (Go test next to the existing subset test), and (b) add a test driving release.sh with a one-name SSoT and a `Lint`-only check-run set that must refuse. Optionally make the script dedupe names and reject a list shorter than the pinned minimum.
- F2 [SHOULD-FIX] scripts/release.sh awk (lines ~257-266) - tolerates CRLF, trailing comment, single quotes poorly: yields names with `\r`, ` # c`, or quotes, which over-refuse with a misleading "not reported" message instead of "malformed release_gate". Required fix: strip `\r` and trailing ` #` comments, and either support or explicitly reject single-quoted entries, so a malformed list fails with the intended message.
- F3 [MINOR] scripts/release.sh:235 comment and :297 die text still say the names live in `CI_REQUIRED_CHECKS` of the script. Required fix: reword to "the release_gate list of .github/required-checks.yml" (changing the die text needs the A2 tests updated, so the lane may keep the text and only fix the comment).
- F4 [MINOR] internal/config/required_checks_release_gate_test.go - passes for any non-empty subset; document or tighten (folds into F1 fix).

## Gaps
- No live GitHub call was made; real check-run names published by ci.yml on a push to main are unobserved (carried from the card).
- I did not run the full repository test suite; the A4 consumer tests were scoped to the packages/tests named above. `internal/cli` was run with a `-run` filter only.
- I did not drive release.sh end to end myself; the script-level refusal evidence for A3 comes from the card's `TestReleaseScriptCIGateSSoTFailsClosed` (re-run green in the `ReleaseScript|CIGate` filter above) plus my direct awk probes. The unsafe cases are shown at the awk-output level and corroborated by codex's end-to-end reproduction, not by my own full script run.
- `validate-required-checks.sh` exit code was swallowed by the pipe; the success banner was observed.
- Windows/Git-Bash awk behavior was not observed.
- claude and glm backends are `off` by the audit plan; only codex voted (participant_count 1).

## Residual-risk
- Even after F1 is fixed, the awk is a line-pattern reader, not a YAML parser; an exotic but valid YAML shape (anchors, flow style, multi-line scalars) will either refuse or be misread. Flow style refuses safely today.
- The gate remains only as correct as the check names in ci.yml; the first real `--dry-run` release is the first live measurement.

## Iteration history
Iteration 1 (this audit): FAIL, driven by required-backend codex `fail` and F1 (reproduced independently). Re-audit scope on fix: F1 (+F2/F3 if taken), delta only.

AUDIT-VERDICT: FAIL spec=t1858 receipts=rcpt-b6a0f22a9eee365cd060f299
