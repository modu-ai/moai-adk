auditor-model: claude-opus-5-5

# t2108 sync audit: reconfigure keeps a stored self-hosted GitLab URL

verdict: PASS-WITH-DEBT
audited_sha: defa1f45ad1563f084d4715958c8cae086d35a5b

## Evaluation Report

- Card: t2108 (Class B, no SPEC). Scope audited: the single commit defa1f45ad (range 65d950bf5a..defa1f45ad). The absorbed t1972 work was not re-audited.
- Overall Verdict: PASS-WITH-DEBT. Score 0.90 (weighted 40/25/20/15 over the scores below).
- Profile: the built-in default (no SPEC, so no `evaluator_profile`). Must-pass dimensions are Functionality and Security, and both pass.
- Binding run conditions: none. No SPEC `progress.md` carries that heading; the card record `.moai/reports/t2108/progress.md` has none either.
- Iteration history: iteration 1 (this one). There is no prior audit of t2108.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | The card tests fail on the base for the stated reason and pass on the fix (RED probe below). The focused suites pass. The saver edge cases were probed and every one behaves as the card asks. |
| Security (25%) | 95/100 | PASS | The diff adds no new input surface. `validateWizardInput` still runs before any write, and the wizard path still sends the prefilled URL through `validateHTTPSURL`. The saver's blank-answer fallback writes back only the value already on disk. |
| Craft (20%) | 82/100 | PASS | The tests kill mutants M1 and M2, but M3 (the stored value beats a non-blank answer) survives every test (F1). Coverage of the changed functions under the focused tests: prefillReconfigureDefaults 55.6% (the username branches were already untested on the base), applyWizardConfig 81.8%, ReadGitLabInstanceURLFromConfig 85.7%. |
| Consistency (15%) | 92/100 | PASS | `gofmt -l` is empty and `go vet` on the wizard package is clean. The reader follows the existing helper pattern (`moaiSectionsDir`, empty string on error). The extraction keeps the username prefill unchanged (probe). |

### Findings (structured defect list)

- F1 [Medium] [optional] [in-diff] `internal/cli/update_wizard_gitlab_t1972_test.go` / `update_wizard_gitlab_prefill_t2108_test.go`: no test covers a non-blank answer that differs from a stored URL. Mutant M3 (the saver prefers the stored URL whenever the provider is gitlab, even over an explicit answer) passes all card tests. The shipped code is correct here (probe "nonblank-differs-from-stored" writes the new URL), so this is a missing regression guard, not a defect. Confidence high. Required fix: add a case with stored `https://gitlab.corp.example`, a result URL of `https://new.example` and provider gitlab, expecting `https://new.example`.
- F2 [Low] [optional] [in-diff] `internal/cli/update_wizard.go:116`: a stored URL that fails `validateHTTPSURL` (for example `http://…` from a hand edit or an older build) now becomes the question default. Pressing Enter sends it to `validateWizardInput`, and the reconfigure stops at save time, after every question has been answered, with `invalid gitlab_instance_url "http://old.example": scheme must be https`. On the base, Enter silently replaced it with gitlab.com. This fails closed and is arguably better than silent loss, but the user has to re-run the wizard. `moai init` validates the flag as https (init.go:314), so normal tooling cannot reach this state. Confidence high (probe). Fix if wanted: prefill only a value that passes `validateHTTPSURL`, or validate inline in the input field.
- F3 [Low] [optional] [in-diff] `internal/cli/wizard/config_helpers.go:54` and `update_wizard.go:300`: the two sides read a non-string scalar differently. With `instance_url: 5` the typed reader returns `"5"` (yaml.v3 decodes a scalar into a string), while the saver's `.(string)` assertion treats it as absent. Only a hand-edited file can reach this. Confidence high (probe). No fix required.
- F4 [Info] [optional] [outside] The deprecated flat key `git_strategy.gitlab_instance_url` is not read by the prefill. No production consumer reads it either (grep), so this is consistent with the codebase. No fix.
- F5 [Info] [optional] [in-diff] With a blank answer, the saver's fallback writes the stored value back without validating it (probe "stored-http-blank-result"). That value was already on disk, so nothing new is introduced. No fix.

No blocking findings. The debt is F1, a missing test for a behaviour the card implies (a non-blank answer must win).

## Claim

1. Mechanism: before the fix, `runInitWizard` prefilled the usernames but not the instance URL, so the question default was `https://gitlab.com`. Since `buildInputField` seeds the input value with `q.Default` and its validator maps a blank answer to `defVal`, Enter wrote gitlab.com. After the fix, `prefillReconfigureDefaults` sets the default to the stored URL, so both a pre-seeded Enter and a cleared blank yield the stored URL. Confirmed by reading `wizard.go:421-461` and by the RED and GREEN test outputs. The interactive TUI was not driven (see Gaps).
2. The "init" half of the card is correctly out of scope. `moai init` builds `InitQuestions` (conversation_language, user_name, page 3) with no Git questions (`questions.go` GitQuestions doc comment; `TestInitWizardQuestionSetHasNoGitCredentialQuestions` exists at `wizard/questions_test.go:155`). init takes the URL only through `--gitlab-instance-url` (init.go:312-315, validated as https).
3. Saver rule, probed case by case. A blank answer with provider gitlab keeps a stored URL. A `gitlab:` section without `instance_url` gets the default and keeps its siblings. A null `gitlab:` gets the default. A provider changed to github leaves the stored URL untouched. Manual mode with a blank provider writes nothing. A non-blank answer that differs from the stored URL wins. Compared with the base, no case regresses except the intended one; F2 is a behaviour change, not a data loss.
4. Reader: a BOM is tolerated. A missing, empty or invalid-YAML file, a map value, a scalar `gitlab:` and duplicate keys all return `""` (duplicate keys fall back to the gitlab.com default; the saver then fails on the same file with a parse error, which is pre-existing behaviour).
5. The tests fail on the pre-fix code for the stated reason, and M1 and M2 are killed. M3 survives (F1). M4 (stored value rewritten for any provider) survives, but it is equivalent on disk and is not a gap.
6. The t1972 test change is justified. The old case asserted the very overwrite this card removes; the other three cases are unchanged.
7. The extraction preserves behaviour: the username prefill logic is identical to the base text (`git show 65d950bf5a:internal/cli/update_wizard.go`), and the probe shows `github_username="gh-user" gitlab_username="gl.user"` prefilled.

## Evidence

Commands run in /Users/goos/.moai/worktrees/projects/moai-adk-go/t2108 at HEAD defa1f45ad1563f084d4715958c8cae086d35a5b:

```
$ timeout 560 go test -count=1 ./internal/cli/ -run 'TestReconfigure|TestApplyWizard|GitLab|TestInitCmd_|TestInitBoolFlag'
ok  	github.com/modu-ai/moai-adk/internal/cli	8.656s

$ timeout 300 go test -count=1 ./internal/cli/wizard/
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	6.505s
exit=0

$ gofmt -l internal/cli internal/cli/wizard ; go vet ./internal/cli/wizard/
exit=0
vet_exit=0

$ timeout 560 go test -count=1 ./internal/cli/ -run 'TestReconfigurePrefillsTheStoredGitLabInstanceURL|TestReconfigureBlankGitLabURL' -v
--- PASS: TestReconfigurePrefillsTheStoredGitLabInstanceURL (0.01s)   [4 subtests PASS]
--- PASS: TestReconfigureBlankGitLabURLKeepsTheStoredValue (0.00s)
--- PASS: TestReconfigureBlankGitLabURLRestoresTheDefault (0.01s)     [4 subtests PASS]
ok  	github.com/modu-ai/moai-adk/internal/cli	1.341s
```

Mutation and RED probes were run on a `git archive` copy of defa1f45ad in the session scratch directory (the audited tree was not modified). Each mode edits `internal/cli/update_wizard.go` and then runs the three card tests:

```
== red exit=1        (extraction kept, URL prefill removed, saver reverted to the base)
    update_wizard_gitlab_prefill_t2108_test.go:55: gitlab_instance_url default = "https://gitlab.com", want "https://gitlab.corp.example"
    update_wizard_gitlab_prefill_t2108_test.go:72: instance_url on disk = "https://gitlab.com", want the stored "https://gitlab.corp.example" kept
    update_wizard_gitlab_t1972_test.go:90: instance_url on disk = "https://gitlab.com", want "https://gitlab.corp.example"
== m1 exit=1         (prefill removed, saver fixed)
    update_wizard_gitlab_prefill_t2108_test.go:55: gitlab_instance_url default = "https://gitlab.com", want "https://gitlab.corp.example"
== m2 exit=1         (prefill fixed, saver reverted to the base)
    update_wizard_gitlab_prefill_t2108_test.go:72: instance_url on disk = "https://gitlab.com", want the stored "https://gitlab.corp.example" kept
    update_wizard_gitlab_t1972_test.go:90: instance_url on disk = "https://gitlab.com", want "https://gitlab.corp.example"
== m3 exit=0         (stored value beats a non-blank answer)        -> SURVIVES (F1)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.699s
== m4 exit=0         (stored value rewritten for any provider)      -> survives, equivalent on disk
== orig exit=0
```

Saver edge probe (scratch test calling `applyWizardConfig`; disk content after the call, abridged to `instance_url`):

```
nonblank-differs-from-stored: err=<nil>   instance_url: https://new.example
provider-to-github-keeps-stored: err=<nil> instance_url: https://old.example  provider: github
gitlab-section-no-url: err=<nil>          instance_url: https://gitlab.com  other: x
stored-int: err=<nil>                     instance_url: https://gitlab.com
gitlab-null: err=<nil>                    instance_url: https://gitlab.com
manual-mode-blank-provider: err=<nil>     instance_url: https://old.example  mode: manual
stored-http-via-wizard-default: err=invalid gitlab_instance_url "http://old.example": scheme must be https, got "http"   (file unchanged)
stored-http-blank-result: err=<nil>       instance_url: http://old.example
TestAuditProbePrefillUsernamesUnchanged: github_username="gh-user" gitlab_username="gl.user" url="https://gitlab.com"
```

Reader edge probe (`ReadGitLabInstanceURLFromConfig`):

```
plain -> "https://gl.example"   bom -> "https://gl.example"   empty-file -> ""   int-value -> "5"
map-value -> ""   null-gitlab -> ""   gitlab-scalar -> ""   dup-key -> ""
flat-deprecated-only -> ""   invalid-yaml -> ""   missing-file -> ""
```

Coverage under the focused tests (scratch copy, `-coverpkg=./internal/cli/,./internal/cli/wizard/`):

```
update_wizard.go:103:      prefillReconfigureDefaults        55.6%
update_wizard.go:142:      applyWizardConfig                 81.8%
config_helpers.go:54:      ReadGitLabInstanceURLFromConfig   85.7%
```

Audit plan and cross-model review:

```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2108
"config_status": "ok", "model": "codex", "cross_model_active": true, "enforced_required": ["codex"]
  (claude gate off, codex required, glm off)

mcp__moai__codex_audit (mode native, target baseBranch, delta_base 65d950bf5a, project_root = this tree), called once:
  verdict: pass, findings: [], resolved_model: gpt-6.1-sol, review_base: "delta_base (merge base 65d950bf5aeb14ad228f9bcedcb4466098528906)",
  build_commit: 646955327e, audit_receipt: rcpt-b5fdfd620ebfea24cb28527b
```

## Baseline-attribution

- Every test, gofmt and vet result above was measured in this run against tree HEAD defa1f45ad1563f084d4715958c8cae086d35a5b (re-read after the audit; `git status --short` was empty both before and after). Go toolchain: go1.26.8 darwin/arm64.
- The pre-fix baseline for RED was not the absorbed parent file, which lacks `prefillReconfigureDefaults` and so would fail at compile time rather than for the stated reason. It is the card tree with the URL prefill block removed and the saver reverted to the parent's text, read from `git show 65d950bf5a:internal/cli/update_wizard.go`. This matches the author's claim that the extraction landed before the RED test.
- The probes ran on a `git archive` copy of defa1f45ad, not the live tree; the copy is byte-identical to the commit.
- The audit-plan verb ran on the installed `moai` (v3.2.0-rc.42; its commit was not printed). The codex audit ran on MCP server build 646955327e. Neither build's ancestry relative to the tree HEAD was checked (see Gaps).

## Gaps

- Refused commands (worktree guard, §3.1). Three compound commands were refused as "names git in a form too complex": the first focused test run with a pipe and `pipestatus`, a `for` loop over mutants, and the coverage one-liner. Each was rerun as a plain command or as a script under the scratch directory; no measurement was replaced by reading the code.
- The interactive huh TUI was not driven end to end. The claim that Enter keeps the pre-seeded value rests on reading `buildInputField` (the value is seeded from `q.Default`, and a blank answer maps to `defVal`) plus the seam tests.
- Not run: the full `internal/cli` suite, the race detector, other OSes, and `golangci-lint`.
- Re-running `moai init` over an existing project (and what it does to a stored URL) was not examined. Init has no wizard Git question, so it is outside this card.
- The ancestry of the installed `moai` build and the MCP server build (646955327e) relative to the tree HEAD was not checked. They only produced the audit plan and the codex review; no test measurement depends on them.
- Per the operator allowance, no `audit_multi`, GLM or Claude-backend audit was run, although the plan reports `cross_model_active: true`. The single required codex leg is cited by its receipt. If the receipt guard refuses this verdict, that is recorded here and the verdict stands.

## Residual-risk

- The surviving mutant M3 means a future change could make a stored URL beat an explicit new URL with no test failing (F1).
- Users with a non-https stored URL get a save-time error after a full wizard pass (F2).
- If the huh library stops seeding the field value from a pre-set Value, or changes how it calls the validator, the behaviour seen in the real TUI could drift from the seam tests.
- The probe and coverage material in the session scratch directory is machine-local; the deciding lines are copied above, and nothing else from scratch is cited.

## Operational Notes (unverified)

- The codex_audit result reports `transport: codex app-server`; this is the MCP server's internal transport. The auditor called only the MCP tool, once, as the allowance permits.
- F1 is a one-case test addition; it fits a follow-up or the next touch of this file.

audit receipt: rcpt-b5fdfd620ebfea24cb28527b
