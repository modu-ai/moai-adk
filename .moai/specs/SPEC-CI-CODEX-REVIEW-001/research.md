# research.md — SPEC-CI-CODEX-REVIEW-001

Read-only investigation, card t1868, tree `4f9dba14bf` (branch `WT-ci-codex-review`, clean), run on
2026-10-18. Every entry below is a command that was run and its observed output. Remote reads are
`gh api` GET only (no write, no secret set, no dispatch, no pull request). Nothing here reads a
committed file under `.moai/reports/`; that path is gitignored by operator order of 2026-09-14.

Tree pin for every local entry: `4f9dba14bf` unless the entry says otherwise. A remote entry names
its own reading date (2026-10-18) and is a measurement of a moving system, never a standing fact.

Plan revision 2 (same day, before the first audit): entries E19-E21 and U6 were added and E12 was
corrected; the RED-now cells of `acceptance.md` were re-measured on tree `8cd1d0d51d` (E21). No remote
read was made in revision 2. E13, E14, and E15 keep their measured content; after leader ruling (b) the
doctrine files they list are known debt KD-1 and follow-up candidates, not edits of this SPEC.

Plan revision 3 (same day, after plan audit iteration 1 and the operator-confirmed design correction): the
reviewer is the ChatGPT subscription's native Codex GitHub review, not a keyed `openai/codex-action`
workflow. Entries E22-E30 were added (all GET-only on the remote, run at tree `ae494aedb2`, commands and
outputs recorded in `.moai/reports/t1868/progress.md`); the U table was rewritten for the new design
(U2-U9); E7 and E8 now describe the **rejected** keyed alternative and E19 notes which of its rows no longer
apply. Nothing here reads a committed file under `.moai/reports/`.

Plan revision 4 (same day, after plan audit iteration 2): entries E31 (the changed-file entry binding) and E32
(the dated DS-1 re-read) and premise U10 were added; U2, U5, U7, and U9 were reworded; the stale
cross-references of revision 2 in E5, E11, and E19 were relabelled to the revision-3 and revision-4
identifiers (marked "formerly" where an identifier no longer exists). Commands and outputs of E31 and E32 are
recorded in `.moai/reports/t1868/progress.md`; E31 and E32 read no committed file under `.moai/reports/`.

## 1. How to read this file

- **E-entries** are measured facts: command, verbatim stdout, exit code.
- **U-entries** are premises taken from the card text or inferred, that could not be re-verified
  read-only. No requirement or criterion in the SPEC rests on a U-entry being true: each is paired
  with the fail-closed design that holds whether it is true or false.
- Output that is long is cut to the deciding lines and says so.

## 2. Premises that stay unverified (U)

| Id | Premise | Why unverified | Design that holds either way |
|----|---------|----------------|------------------------------|
| U1 | CodeRabbit answered PR 1826 `Too many files / Review skipped` | The PR comment was not fetched; only the card text states it | The gate withholds when no recognised review exists, regardless of who reviews (REQ-CCR-007); no requirement mentions PR 1826 |
| U2 | **Which artifact the native Codex review leaves on a pull request** (a pull-request review, a review comment, an issue comment, or a check run), under which author | Not observed: none of 120 recent pull requests of the two repositories carries any artifact of the Codex app (E29), and a pull request opened about 82 minutes before a later read still had none (E32; the leader's earlier read at about 47 minutes found none either); the review was switched on only recently per the operator | The gate recognises exactly one documented form (a pull-request review by the identity, bound by `commit_id`), counts only state APPROVED as clean, and fails closed on every other; deferred step DS-1 observes the form (REQ-CCR-001, REQ-CCR-003, REQ-CCR-004) |
| U3 | The pull-request files endpoint returns at most 3000 files | Documented limit recalled, not read in this session | `fetch-pages` refuses when a 30th page is full, so a truncated list can only withhold (REQ-CCR-010) |
| U4 | Whether the native review runs on a **draft** pull request, and whether it runs at the ready-for-review transition | The contract says "when a pull request opens"; nothing observed shows a draft or a transition case | The gate reads `draft` and reports NOT RUN at once whatever exists; nothing depends on the answer; deferred step DS-2 observes it (REQ-CCR-008) |
| U5 | How the native review represents **findings** (inline review comments versus review-body text) and a **clean** result (an APPROVED review, a COMMENTED review with no comment, or another form) | Not observed (E29, E32) | Findings are review comments by the identity on the head; body text is never read; only an APPROVED review is clean, so a COMMENTED review, a body-only finding, or a clean result in another form is never a pass and the gate withholds (REQ-CCR-001, REQ-CCR-004, REQ-CCR-005; design §3.2 residual) |
| U6 | How a fresh review would be requested for a head that moved after the pull request opened | Per-push review is excluded by the operator; no request path was exercised | A moved head is `stale-review` and withheld; DS-3 is a leader-coordinated step, never a deliverable (REQ-CCR-002) |
| U7 | A re-run of the `CI` workflow run fires the `workflow_run` event again, so auto-merge reads a late review | Platform behavior recalled, not observed | The gate withholds a late review as NOT RUN `review-absent` either way; the re-trigger path is an operator action (accepted limitation AL-1) |
| U8 | `CI`, and so auto-merge, fires for a **draft** head; and the native review reviews Dependabot-authored pull requests | Not observed | A missing auto-merge run on a draft is a gap, not a pass (design §12); an unreviewed bot pull request is withheld by the gate |
| U9 | The reviews list can be truncated without notice | No count of reviews exists on the pull request object (E23), so completeness rests on the page rules | `fetch-pages` collects every page, refuses at a full 30th page, on any page error, and on a repeated entry (REQ-CCR-010); a listing larger than 3000 reviews is refused; a silent shortening with no error is not detectable and is a named residual risk (design §6) |
| U10 | The `ref` in the `contents_url` of a changed-file entry equals the pull request head for every status other than `removed`, for statuses not observed (`copied`, `changed`, `unchanged`), and for a base that moved after the pull request opened | Measured only on three pull requests and 1,147 entries (E31); removed entries carry the base SHA | The reader binds every entry whose status is not `removed`, so an unobserved status that carried another ref would withhold (fail closed, not a pass); removed entries are not bound and are a named residual (design §6) |
| U11 | What the workflow platform reports on a runner for the amended `files` and `review` steps and their job: a non-zero step status as a failed step and a failed job, the later steps skipped through the implied `success()`, and a draft's `review` step as a non-success job | Platform behavior recalled, not observed: the amended workflow has never run on a runner, and the G6 draft pull request cannot run it (`workflow_run` runs the default branch's copy) | The local criteria assert the step exit status and a job result derived from the step statuses (AC-CCR-020 to AC-CCR-025); nothing claims the remote conclusion, and the remote state is "unverified" until a pull request after landing is observed (spec §5, G6 acceptance condition) |
| U12 | Re-running a failed Auto Merge job re-reads a review that arrived after the window | Platform behavior recalled, not observed; a failed job offers a re-run, a green one did not | The gate withholds either way; the re-trigger is an operator action outside this lane (AL-1) |

## 3. Measured facts (E)

### E1 — the gate script does not exist

```
$ test -f scripts/ci/codex-review-gate.sh
(no stdout)   exit 1
```

### E2 — the review workflow does not exist (revision 3: and none is built; the absence stays true and AC-CCR-013 pins that this SPEC adds no workflow)

```
$ test -f .github/workflows/codex-review.yml
(no stdout)   exit 1
```

### E3 — what `auto-merge.yml` reads today (tree 4f9dba14bf)

```
$ grep -c -i coderabbit .github/workflows/auto-merge.yml
14   exit 0
$ grep -c -F 'issues/$PR_NUMBER/comments' .github/workflows/auto-merge.yml
1    exit 0
$ grep -c 'gate-files-changed' .github/workflows/auto-merge.yml
0    exit 1
```

Structure read from the file (line numbers at the pinned tree): trigger `workflow_run` of `CI`
(5-7); permissions `contents: write`, `pull-requests: write` only (9-11); checkout step has no
`ref` (55-56); head re-verification (63-88); required-check observation named `Check all required CI
checks passed` (90-173); the CodeRabbit step reads the commit status `CodeRabbit` and the
`Merge Risk:` comment line, with a ceiling taken from variable `AUTOMERGE_MAX_MERGE_RISK` (175-293);
the merge step is pinned with `--match-head-commit` (295-359). `TestAutoMergeRequiredChecks`
(`internal/template/auto_merge_checks_test.go:18`) extracts the step by name `Check all required CI
checks passed` and runs it with a stub `gh`.

### E4 — the Claude surfaces present today

```
$ test -f .github/workflows/claude.yml
exit 0
$ test -f .github/workflows/review-quality-gate.yml
exit 0
$ grep -rn 'anthropics/claude-code-action' .github/workflows
.github/workflows/claude.yml:35:        uses: anthropics/claude-code-action@v1
.github/workflows/claude.yml:47:          # See https://github.com/anthropics/claude-code-action/blob/main/docs/usage.md
exit 0
```

`claude.yml` triggers on `issue_comment`, `pull_request_review_comment`, `issues`,
`pull_request_review` and runs only when the text contains `@claude` (lines 3-19); it reads
`secrets.CLAUDE_CODE_OAUTH_TOKEN` (line 37). `review-quality-gate.yml` triggers on `check_run`
`completed` and runs only for a check run named `Claude Code Review` (lines 9, 22).

### E5 — `claude-code-review.yml` is remote-only

```
$ test -f .github/workflows/claude-code-review.yml
(no stdout)   exit 1                  (absent from the local tree)
$ git ls-tree --name-only origin/main .github/workflows/claude-code-review.yml
.github/workflows/claude-code-review.yml   exit 0   (present on origin/main)
$ git merge-base --is-ancestor 610840bf63 HEAD
exit 1
$ git merge-base --is-ancestor 610840bf63 origin/main
exit 0
$ git rev-list --count --left-right origin/main...HEAD
10	400
```

Commit `610840bf63` (2026-10-14, "Claude Code Review workflow") added the file on the remote side
after commit `b3ab67ce86` ("remove claude-review workflow (CodeRabbit replaces it)") had removed it
locally. `origin/main` at the pin of this read is `6dde1c1c1b`. Consequence: a local `git rm` cannot
delete the file, and merging the local line into `origin/main` does not delete it either; the
deletion belongs to the leader's draft pull request (gate G6) and a census test makes a surviving
copy red (AC-CCR-014; formerly AC-CCR-016 in revision 2).

The file content on `origin/main` (read with `git show origin/main:.github/workflows/claude-code-review.yml`):
trigger `pull_request` types `opened, synchronize, ready_for_review, reopened`; job `claude-review`
runs `anthropics/claude-code-action@v1` with `secrets.CLAUDE_CODE_OAUTH_TOKEN` and the plugin
`code-review@claude-code-plugins`; `permissions` include `id-token: write`.

### E6 — main's measured required status checks (GET, 2026-10-18)

```
$ gh api repos/modu-ai/moai-adk/branches/main/protection/required_status_checks
{"url":"https://api.github.com/repos/modu-ai/moai-adk/branches/main/protection/required_status_checks","strict":false,"contexts":["Test (ubuntu-latest)","Lint","Build (linux/amd64)","Analyze (Go) (go)","Release PR Multi-OS Gate"],"contexts_url":"...","checks":[{"context":"Test (ubuntu-latest)","app_id":15368},{"context":"Lint","app_id":15368},{"context":"Build (linux/amd64)","app_id":15368},{"context":"Analyze (Go) (go)","app_id":15368},{"context":"Release PR Multi-OS Gate","app_id":15368}]}
```

Five required contexts, no review check, `strict: false`. The local `.github/required-checks.yml`
`branches.main.contexts` lists ten names (the five plus four `Build` matrix names and `Guard Bundle`),
and `release_gate` three (lines 34-50, 62-65). That local-versus-remote difference exists before
this SPEC and this SPEC does not touch it.

### E7 — secrets and workflows on the remote (GET, 2026-10-18; revision 3: relevant only to the rejected keyed alternative)

```
$ gh api repos/modu-ai/moai-adk/actions/secrets --jq '.secrets[].name'
CLAUDE_CODE_OAUTH_TOKEN
GLM_API_KEY
GROQ_API_KEY
PYPI_API_TOKEN
TESTPYPI_API_TOKEN
VERCEL_ORG_ID
VERCEL_PROJECT_ID
VERCEL_TOKEN
$ gh api orgs/modu-ai/actions/secrets --jq '.secrets[].name'
(no output, no error)          -> see U4
```

No `OPENAI_API_KEY` at repository level. The workflow list (GET `.../actions/workflows`) shows
active: `claude-code-review.yml` (204797886), `claude.yml` (231530404), `review-quality-gate.yml`
(254299093), `auto-merge.yml` (254644310), plus the CI and release workflows.

### E8 — the Codex action, as read (GET, 2026-10-18; revision 3: the REJECTED keyed alternative, decision-index Q1 — kept as the evidence for why it was weighed and rejected; nothing in the SPEC uses it)

```
$ gh api repos/openai/codex-action/git/ref/tags/v1
object: {"sha":"1d82a3dba7c7988ec71ecae1ed5b8e8db96495c5","type":"tag"}
$ gh api repos/openai/codex-action/git/tags/1d82a3dba7c7988ec71ecae1ed5b8e8db96495c5
object: {"sha":"bdf19a4a223ec2549a3e2274a0cf61556bc07675","type":"commit"}   tagger date 2026-10-05T23:23:32Z, "verification":{"verified":false,"reason":"unsigned"}
$ gh api repos/actions/checkout/git/ref/tags/v7
object: {"sha":"3d3c42e5aac5ba805825da76410c181273ba90b1","type":"commit"}
```

`action.yml` of `actions/checkout` read at its pinned commit (`contents/action.yml?ref=3d3c42e5...`)
declares the inputs the design uses: `ref` (line 7), `persist-credentials` (52), `sparse-checkout` (65),
`fetch-depth` (74); `runs.using: node24`.

So `openai/codex-action@v1` currently resolves to commit `bdf19a4a223ec2549a3e2274a0cf61556bc07675`
through an unsigned, movable tag, and `actions/checkout@v7` to
`3d3c42e5aac5ba805825da76410c181273ba90b1`. The repository's other workflows pin by tag (`actions/checkout@v7`);
the review workflow pins both by SHA because its Codex step holds a paid key (REQ-CCR-008). The run
phase re-resolves both SHAs by GET before writing them and stops with a blocker report if the GET is
refused.

`action.yml` read at the pinned commit (`contents/action.yml?ref=bdf19a4a...`, 382 lines) declares the
inputs the design uses: `openai-api-key`, `prompt`, `prompt-file`, `output-file`, `output-schema-file`,
`sandbox` (description: one of `workspace-write`, `read-only`, `danger-full-access`; "Mutually
exclusive" with `permission-profile`), `permission-profile`, `safety-strategy` (default `drop-sudo`;
`read-only` "could read `openai-api-key` from memory"; `unsafe` "NOT RECOMMENDED"), `allow-users`,
`allow-bots` (default `false`), `allow-bot-users`; and the output `final-message` ("Raw output
emitted by `codex exec`"). The README (default-branch read) states the key must be supplied as a
secret, the default action user must have write access, and recommends running the action as the
last step of a job and passing its output to a second job. `docs/security.md` lists pull-request
title/body, commit messages, repository instruction files, and screenshots as prompt-injection
sources, and says to use `drop-sudo` or `unprivileged-user` to protect the key.

### E9 — `.github/` is not templated for the files this SPEC edits

```
$ ls internal/template/templates/.github
actions  branch-protection.json.gtmpl  labels.yml  workflows
$ find internal/template/templates -path '*workflows*' -name '*.yml*'
internal/template/templates/.github/workflows/label-sync.yml
```

No template copy exists for `auto-merge.yml`, `claude.yml`, `review-quality-gate.yml`,
`.coderabbit.yaml`, or `required-checks.yml`. The one mirrored pair near this work is
`branch-protection.json.gtmpl` (`TestBranchProtectionParity`,
`internal/template/branch_protection_parity_test.go:52`), which this SPEC leaves byte-unchanged.

### E10 — CodeRabbit configuration today

```
$ grep -n 'enabled:' .coderabbit.yaml
65:    enabled: true
$ grep -n 'auto_reply' .coderabbit.yaml
74:  auto_reply: true
```

`reviews.auto_review.base_branches` lists `main`, `develop`, `moai-go-v2`, `release/*` (lines
67-71). Deleting the file would return the app to its defaults, which review; so the file stays with
the two flags off (decision-index Q8).

### E11 — contract files and the byte-identity probe

```
$ git diff --quiet 4f9dba14bf HEAD -- .github/required-checks.yml
exit 0
$ git diff --quiet 4f9dba14bf origin/main -- .github/required-checks.yml
exit 1
$ git diff --quiet 4f9dba14bf origin/main -- .github/workflows/ci.yml
exit 0
```

The same probe is green on the baseline and red on a tree known to differ (`origin/main` differs
from the baseline in `required-checks.yml`). That is the observed failure that makes AC-CCR-016's
command (formerly AC-CCR-018 in revision 2) a verdict rather than a print.

### E12 — guard-liveness manifest

`.moai/guards/liveness.yaml` declares one entry per watched workflow. Entry locators at lines
34 (`auto-merge.yml`), 48 (`claude.yml`, events `issue_comment`, `pull_request_review_comment`,
`issues`, `pull_request_review`), 146 (`review-quality-gate.yml`, events `check_run`). The
file header states why it is not mirrored into templates and what `measure` means.

```
$ grep -c 'review-quality-gate.yml' .moai/guards/liveness.yaml
1   exit 0
$ grep -c 'codex-review.yml' .moai/guards/liveness.yaml
0   exit 1
```

Only `internal/guardstate/manifest.go` parses it (`const ManifestPath = ".moai/guards/liveness.yaml"`, line 31).

**Correction (plan revision 2).** Revision 1 of this entry said "no test asserts that a locator file
exists". That was wrong. `internal/guardstate/census_test.go` loads the real manifest
(`loadRepoManifest`, lines 32-40) and `TestCensus_SetDifferenceEmptyBothDirections` (lines 49-91) asserts
that the manifest locators and the files matching `.github/workflows/*.yml` and `*.yaml` are equal as
sets in both directions, with a duplicate locator a failure; `TestCensus_EveryEntryCarriesFiveFields`
(96-106) validates every entry; `TestCensus_ReleaseOnlySubjectsDeclareTheirCondition` (112-136) requires
`release.yml` and `spec-status-auto-sync.yml` to carry `expected_when`. Measured counts at tree
`8cd1d0d51d`: 21 workflow files on disk (17 `.yml` plus 4 `.yaml`) and 21 `locator:` lines in the
manifest. Consequence: adding `codex-review.yml` or deleting `claude.yml` and `review-quality-gate.yml`
turns that existing test red unless the manifest changes in the same commit (plan §B.1).
Revision 3: this SPEC adds no workflow, so only the removal half applies (two entries removed with the two
deletions).

### E13 — surfaces that instruct a CodeRabbit read (census)

```
$ grep -rIl 'Merge Risk' . --exclude-dir=.git --exclude-dir=specs --exclude-dir=reports --exclude-dir=worktrees --exclude-dir=_archive --exclude=CHANGELOG.md
AGENTS.md
.moai/policies/workflow/kanban-dispatch.md
.moai/policies/workflow/factory-dispatch-mechanics.md
.moai/policies/workflow/factory-dispatch.md
.moai/policies/workflow/factory-dispatch-detail.md
.moai/policies/workflow/kanban-dispatch-detail.md
internal/template/testdata/binding_ledger.json
internal/template/templates/AGENTS.md.tmpl
internal/template/templates/.claude/agents/moai/manager-lead.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md
internal/template/templates/.codex/agents/moai/manager-lead.toml
internal/cli/testdata/codex-rollouts-t1171/roles-other-version/manager-lead.toml
internal/cli/testdata/codex-rollouts-t1171/real/rollout-2026-09-24T18-41-35-01a0d2ca-b7e2-7e50-8f7f-0ab053beb07f.jsonl
internal/cli/testdata/codex-rollouts-t1171/roles/manager-lead.toml
internal/hook/session_start_factory_i18n.go
internal/hook/role_core_elements_test.go
.claude/agents/moai/manager-lead.md
.claude/rules/moai/workflow/factory-dispatch-mechanics.md
.claude/rules/moai/workflow/factory-dispatch-core.md
.claude/rules/moai/workflow/factory-dispatch.md
.claude/rules/moai/workflow/factory-dispatch-detail.md
.github/workflows/auto-merge.yml
```

A second census on `Review completed|Review rate limited` adds `factory-dispatch-gates.md`
(deployed, template, and `.moai/policies/workflow/`). `.moai/policies/workflow/*` is a byte projection
of `.claude/rules/moai/workflow/*` (`cmp` of `factory-dispatch-gates.md` against its policies copy:
exit 0), tracked in this repository. The frozen fixture trio under
`internal/cli/testdata/codex-rollouts-t1171/` holds recorded role text and is kept as is. Also naming
CodeRabbit: `.moai/project/tech.md:157` (a tooling table row), `.moai/docs/jev-local-operations.md:26`
(a retained-powers list), `.moai/docs/batch-failure-bisect-runbook.md:62,75` (an explanation that
external statuses are not read by the required-check count — still true).

Scope after leader ruling (b): of these 25 paths, `.github/workflows/auto-merge.yml` is the only one this
SPEC edits; the three `internal/cli/testdata/codex-rollouts-t1171/` paths are frozen fixtures kept as is;
the other 21 are doctrine, role-prompt, template, projection, hook, test, and ledger files that are known
debt KD-1 and follow-up candidates (decision-index § Follow-up candidates), not edits of this SPEC.

### E14 — the session-start lane notice

```
$ grep -c 'CodeRabbit' internal/hook/session_start_factory_i18n.go
4   exit 0
$ grep -c 'Codex Review' internal/hook/session_start_factory_i18n.go
0   exit 1
```

The four hits are the four locale strings of the lane notice ("CodeRabbit via the commit-status API
plus the Merge Risk line", line 140 in English). `internal/hook/role_core_elements_test.go:51`
requires the injected role core to carry `Merge Risk:`. No existing test asserts the lane-notice
CodeRabbit phrase. The binding ledger `internal/template/testdata/binding_ledger.json` pins the
current text in rows ALB-0284 and ALB-0285 (lines 5777-5837) and two more rows (4680-4681,
5653-5654); `TestBindingLedgerIntegrity` (`internal/template/binding_ledger_test.go:1131`) reads it.

### E15 — existing tests that read the surfaces this SPEC edits (names verified by grep)

After ruling (b) the doctrine-reading tests in this list (`TestRuleTemplateMirrorDrift`,
`TestDeclaredRuleMirrorForks`, `TestBindingLedgerIntegrity`, `TestDeployedAlwaysLoadedCharBudget`,
`TestRoleCoreStubGate`, `TestDeployedRuleSectionRefsResolve`, `TestInjectedCoreCarriesObligationElements`)
read files this SPEC no longer edits, so they are not re-measured by this SPEC; the tests this SPEC does
reach are in E19.

`TestAutoMergeRequiredChecks` (`internal/template/auto_merge_checks_test.go:18`),
`TestBranchProtectionParity` (`branch_protection_parity_test.go:52`),
`TestRuleTemplateMirrorDrift` (`rule_template_mirror_test.go:378`), `TestDeclaredRuleMirrorForks`
(`:99`), `TestBindingLedgerIntegrity` (`binding_ledger_test.go:1131`),
`TestDeployedAlwaysLoadedCharBudget` (`always_loaded_budget_test.go:261`), `TestRoleCoreStubGate`
(`role_core_stub_gate_test.go:95`), `TestDeployedRuleSectionRefsResolve`
(`deployed_rule_section_refs_test.go:338`), `TestInjectedCoreCarriesObligationElements`
(`internal/hook/role_core_elements_test.go:17`), `TestNoHardcodedContexts`
(`internal/config/required_checks_test.go:81`), `TestRequiredChecks_ReleaseGateSubsetOfMain`
(`internal/config/required_checks_release_gate_test.go:10`).

### E16 — sibling cards (read with `moai todo show`)

- t1865 (picked): `moai update` ownership, manifest and hash provenance, empty-directory cleanup,
  unregistered `.agents/skills` handling. Its scope names `internal/cli/update*` territory and the
  t1668 branch; it names no workflow, doctrine, or CI file.
- t1632 (picked): MCP rejection structuring, project_root, todo_claim catalogue, inbox. Names no
  workflow, doctrine, or CI file.
- t1731 (picked, on hold per the dispatch): GitHub Actions duplicate calls — `release-pr-multi-os.yml`
  cancellation, duplicate manual CI re-invocation, fast test / race / browser fire guard / SPEC Lint
  job bottlenecks, proposal 7 (an always-running aggregate job for required checks; PR concurrency
  cancel; batch PR double `ci.yml` run). Overlap with this SPEC's edited files: none. Overlap with
  files this SPEC only guards unchanged (`ci.yml`, `release-pr-multi-os.yml`, `required-checks.yml`):
  possible. Rule recorded by the dispatch: this card goes first and t1731 absorbs it on resume.

### E17 — workflows that already use risky triggers (so REQ-CCR-007 binds only the new file)

```
$ grep -n 'pull_request_target' .github/workflows/graph-freshness.yml .github/workflows/community.yml
graph-freshness.yml:14:  pull_request_target:
community.yml:8:  pull_request_target:
```

Existing workflows use `pull_request_target`; the constraint (REQ-CCR-013 in revision 3, formerly
REQ-CCR-007) is scoped to the workflows this SPEC edits or adds — `auto-merge.yml` only — not
repository-wide.

### E18 — the repository's own precedent for "who published this check run"

`scripts/release.sh:289-307` reads check runs with `.app.slug` and requires `CI_APP="github-actions"`
(test: `internal/template/release_ci_gate_test.go` rows `rlsCheckRunsRightNameWrongApp`,
`rlsCheckRunsNoApp`). Revision 3: the consumer reads no check run; it applies the same "who published
this" discipline to a pull-request review (login, numeric id, and account type of the Codex app identity).

### E19 — existing tests and files that the deletions, the new workflow, or the config change touch (plan revision 2)

Tree `8cd1d0d51d`. Every row is a command that was run (grep or read; no `go test` was run) and what it
showed. The handling milestone is the one that carries the test in its re-measure list.

```
$ grep -rln --include='*_test.go' -e 'claude\.yml' -e 'review-quality-gate' -e 'claude-code-review' -e 'Claude Code Review' . (excluding .git, worktrees, node_modules)
internal/config/required_checks_auxiliary_test.go  internal/ciwatch/coverage_test.go  internal/ciwatch/handoff_test.go
internal/ciwatch/classifier_test.go  internal/ciwatch/formatter_test.go  internal/cli/pr/watch_test.go  internal/cli/pr/coverage_test.go   exit 0
$ grep -rln --include='*_test.go' -i -e 'coderabbit' -e 'auto_review' -e 'auto_reply' . (same exclusions)
internal/template/renderer_test.go  internal/template/batch_gate_summary_doctrine_test.go
internal/template/release_ci_gate_test.go  internal/spec/drift_close_body_test.go   exit 0
$ grep -rln --include='*_test.go' -e '\.coderabbit' . (same exclusions)
(no output)   exit 1
```

| Existing test or file | What it reads | Touched by | Handling |
|-----------------------|---------------|------------|----------|
| `internal/guardstate/census_test.go` `TestCensus_SetDifferenceEmptyBothDirections`, `TestCensus_EveryEntryCarriesFiveFields`, `TestCensus_ReleaseOnlySubjectsDeclareTheirCondition` | the real `liveness.yaml` and `.github/workflows/*.yml\|*.yaml` (E12) | revision 2: the new workflow (M2) and the two deletions (M5); revisions 3 and 4: the two deletions only (M4) | Revision 2 (formerly): manifest entry added in M2 with the workflow, two removed in M5 with the deletions, criteria formerly AC-CCR-020 and AC-CCR-021 (neither exists any more). Revisions 3 and 4: no entry is added because no workflow is added; the two entries are removed in M4 with the deletions; the three tests are re-measured at M4; the new criterion test is AC-CCR-018 |
| `internal/cli/codex_live_axis_declaration_test.go` `TestCodexLiveAxis_NotObservedInCI` (lines 204-257) | every `.github/workflows/*.yml\|*.yaml`, for the five `MOAI_*` live-axis switch names | the new workflow | The workflow carries none of the five names (a constraint in spec §4 and plan §B.2); re-measured by name at M2 |
| `internal/template/auto_merge_checks_test.go` `TestAutoMergeRequiredChecks` (line 18) | `.github/workflows/auto-merge.yml`, step named `Check all required CI checks passed` (line 29 reads the file) | the `auto-merge.yml` edit | M3 keeps the step name; AC-CCR-011 lists its 16 `--- PASS` rows (formerly M4 and AC-CCR-015 in revision 2) |
| `internal/ciwatch/*_test.go`, `internal/cli/pr/*_test.go`, `internal/config/required_checks_auxiliary_test.go` | the check-run name `claude-code-review` inside inline YAML or struct fixtures (for example `required_checks_auxiliary_test.go:27` and `ciwatch/coverage_test.go:22`) | nothing — fixtures are inline strings, not reads of a deleted file; `.github/required-checks.yml` carries no `claude-code-review` or `Claude` text (grep exit 1) | None needed; the file stays byte-unchanged anyway |
| `internal/template/release_ci_gate_test.go` | an inline `CodeRabbit` check-run row as an unrelated-app fixture for `scripts/release.sh` | nothing | None needed; `scripts/release.sh` is unchanged (its line 232 is a comment) |
| `internal/template/renderer_test.go`, `batch_gate_summary_doctrine_test.go`, `internal/spec/drift_close_body_test.go` | the word "CodeRabbit" in a comment or a quoted historical commit line | nothing | None needed |
| `.github/workflows/workflow-parse-guard.yaml` | parses `.claude/workflows/*.js` scripts (paths filter lines 16-20), not GitHub workflow files | nothing | None needed |
| `.github/workflows/template-neutrality-check.yaml` | runs `TestTemplateNeutralityAudit` and the leak guards over `internal/template/templates/**` | nothing: no template file is edited, and the new test files sit outside `templates/` | None needed |
| `internal/template/branch_protection_parity_test.go` `TestBranchProtectionParity` | `branch-protection.json.gtmpl` and its mirror | nothing: both stay byte-unchanged (AC-CCR-016; formerly AC-CCR-018 in revision 2) | Re-measured by name at landing |
| `internal/config/required_checks_test.go` `TestNoHardcodedContexts`, `required_checks_release_gate_test.go` `TestRequiredChecks_ReleaseGateSubsetOfMain` | `.github/required-checks.yml` | nothing: byte-unchanged | Re-measured by name at M4 (formerly M5 in revision 2) |

Revision 3 (no new workflow, no `MOAI_*` switch to re-measure): the `codex_live_axis_declaration_test.go`
row no longer applies, and the census row applies to the removals only. The `auto_merge_checks_test.go` row
applies to the `auto-merge.yml` edit as before.

Other consumers searched, none found: no non-test Go file reads CodeRabbit text except the four lane-notice
strings (E20); `CLAUDE_CODE_OAUTH_TOKEN`, `@claude`, and `AUTOMERGE_MAX_MERGE_RISK` appear only in the
workflows this SPEC edits or deletes (`grep` over `.github scripts Makefile .moai/config .moai/guards
.moai/docs`, exit 0, no other file); no workflow names another workflow by `workflows:` list; the only
CodeRabbit text outside `auto-merge.yml` under the scanned trees is `scripts/release.sh:232` (a comment,
outside `scripts/ci/`) and `.moai/guards/liveness.yaml` has no such text. `.github/workflows` holds 21
workflow files today (17 `.yml`, 4 `.yaml`); 19 are expected after the change in revisions 3 and 4 (21 minus
2; revision 2 had planned 20, 21 minus 2 plus 1, for a workflow that is no longer added), so the census floor
of AC-CCR-014 (at least 18; formerly AC-CCR-016) holds before and after. No mechanical consumer of the CodeRabbit
text exists in Go (`grep -rn -i -e coderabbit -e 'merge risk' --include='*.go' internal cmd pkg` outside
`_test.go` printed only `session_start_factory_i18n.go` lines 140, 200, 260, 319 and an unrelated word at
`update_template_sync.go:1095`).

### E20 — which doctrine sentences go stale (known debt KD-1)

Tree `8cd1d0d51d`, `grep -n` over the deployed files. These sentences describe CodeRabbit as the review
source or instruct the `Merge Risk:` / combined-status read; they stay unedited by this SPEC.

| Location | Sentence or content |
|----------|---------------------|
| `.claude/rules/moai/workflow/factory-dispatch-core.md:46` and `:48-50` | heading "CodeRabbit is not read from `gh pr checks`"; "[HARD] A CodeRabbit `gh pr checks` row is not evidence — counts only with combined-endpoint `success`+`Review completed` and a matching `Merge Risk:` line"; "`Review rate limited` means the review never started, and a card carrying it does not leave `sync`" |
| `.claude/rules/moai/workflow/factory-dispatch.md:146-150`, `:132` | the same rule in the full body; the deputy holds no power over "CodeRabbit slot-wait adjudication" |
| `AGENTS.md:200-203` (§4) | "A CodeRabbit row in `gh pr checks` is not evidence that a review ran ... (1) ... `CodeRabbit` context with `state == "success"` and description `Review completed`; (2) a `Merge Risk:` line ..." |
| `.claude/agents/moai/manager-lead.md:224, 236` | "CodeRabbit two-condition READS" and "CodeRabbit discipline adjudication" |
| `.claude/rules/moai/workflow/factory-dispatch-gates.md:10-14` | "CodeRabbit endpoint measurement" companion section |
| `internal/hook/session_start_factory_i18n.go:140, 200, 260, 319` | the lane notice, four locales: "CodeRabbit via the commit-status API plus the Merge Risk line" |
| Template mirrors, `.moai/policies/workflow/*` projections, `.moai/project/tech.md:157` | the same text, per the E13 census |

Existing tests that pin the stale text (read, not run): `internal/hook/role_core_elements_test.go:51-52`
(rows `ALB-0284 (Merge Risk line)` requiring `Merge Risk:` and `ALB-0284 (combined endpoint)` requiring
`/commits/{sha}/status`), and `internal/template/testdata/binding_ledger.json` rows `ALB-0284` and
`ALB-0285` (`grep -c` of the two ids prints 2 lines), read by `TestBindingLedgerIntegrity`
(`binding_ledger_test.go:1131`). Because the doctrine is left unedited, these inputs and assertions are
unchanged.

### E21 — revision-2 re-measurement at tree 8cd1d0d51d

```
$ git rev-parse --short HEAD                      8cd1d0d51d
$ git diff --name-only 4f9dba14bf HEAD
.moai/reports/t1868/progress.md and the seven files under .moai/specs/SPEC-CI-CODEX-REVIEW-001/   exit 0
$ git rev-parse origin/main                       6dde1c1c1bdc84bf82c1b8266c071c44756fe7c0
$ ls .github/workflows (entries)                  21 files
```

Every RED-now cell of `acceptance.md` §1 was run again at this tree and printed the output recorded
there (R1 exit 1, R2 exit 1, R3 14, R4 1, R5 0 exit 1, R6 two lines, R7 lines 65 and 74, R8 baseline exit 0
and control exit 1, R9 0 exit 1, R10 both exit 0, R11 2). The R8 control now names the commit
`6dde1c1c1b` instead of the moving ref `origin/main`. No remote read was made.

### E22 — the changed-file listing is paged at 30 by default, and its count agrees with the pull request's own (GET, 2026-10-18, tree ae494aedb2)

```
$ gh api "repos/modu-ai/moai-adk/pulls/1826/files"                         -> 30 entries
$ gh api "repos/modu-ai/moai-adk/pulls/1826/files?per_page=100&page=11"    -> 29 entries
$ gh api "repos/modu-ai/moai-adk/pulls/1826"   (.changed_files)            -> 1029
paginated by an explicit loop, per_page=100, pages 1..11                   -> 10 x 100 + 29 = 1029 entries; 9 entries carry previous_filename
```

The default page holds 30 entries, so a single call sees 30 of 1029 (audit defect D2 is real). The explicit
loop collected exactly the pull request's own `changed_files`, a renamed file counting once with
`previous_filename`. The 3000-entry ceiling (U3) was not reached by any measured pull request.

### E23 — review and review-comment objects, and the review-comment count proof (GET, 2026-10-18)

```
$ gh api "repos/modu-ai/moai-adk/pulls/1823/reviews"      (first review)
keys: _links, author_association, body, commit_id, html_url, id, node_id, pull_request_url, state, submitted_at, user
sample: id 5516835540, state COMMENTED, commit_id 610840bf6310ab0a11348393eac070ed4e2133bf, submitted_at 2026-10-14T10:57:15Z,
        user.login coderabbitai[bot], user.id 136622811, user.type Bot
$ gh api "repos/modu-ai/moai-adk/pulls/1823/comments"     (first review comment)
keys include: commit_id, original_commit_id, pull_request_review_id, user, path, line, body
sample: id 4269881788, commit_id == original_commit_id == 610840bf6310ab0a11348393eac070ed4e2133bf, user coderabbitai[bot] id 136622811 Bot
count proof, 30 most recent pull requests of modu-ai/moai-adk: object.review_comments == length of the paginated
review-comment list on 30 of 30 (mismatches: none)
```

A pull-request review carries the commit it is bound to (`commit_id`), a state, a submission time, and the
author's login, numeric id, and account type; a review comment carries `commit_id` and `original_commit_id`.
The pull request object carries `review_comments` (a count that matched the list on every pull request
measured) and `changed_files`, but no count of reviews (U9). The `draft` flag is a field of the pull request
object. These are the documented REST shapes the gate's fixtures follow; they were measured on a CodeRabbit
review, not on a Codex one.

### E24 — a `codex-review` check run exists on a mo.ai.kr pull request, and it is not the native app (GET, 2026-10-18)

```
$ gh api "repos/modu-ai/mo.ai.kr/commits/<head of PR 425>/check-runs?per_page=100"
Vercel Preview Comments  completed success  app vercel
codex-review             completed failure  app github-actions   details https://github.com/modu-ai/mo.ai.kr/actions/runs/39119153445/job/117651001452
```

That check run is published by `github-actions`, so it belongs to a workflow of modu-ai/mo.ai.kr, not to the
Codex app, and it is outside this card's repository. It shows only that a repository workflow can carry the
name `codex-review`; the gate reads no check run (decision-index Q6).

### E25 — workflow token defaults (GET and grep, 2026-10-18)

```
$ gh api "repos/modu-ai/moai-adk/actions/permissions/workflow"
{"default_workflow_permissions":"read","can_approve_pull_request_reviews":false}
$ grep -rn 'checks: write' .github/workflows           (no output)   exit 1
$ grep -L '^permissions:' .github/workflows/*.yml .github/workflows/*.yaml
.github/workflows/claude.yml
.github/workflows/review-quality-gate.yml            (both retired by this SPEC)
```

The repository default token permission is `read`, no workflow grants `checks: write`, and the only two
workflows without a top-level permissions block are the two this SPEC deletes. This supports the tamper
guard's rationale (a pull-request-added workflow gets a read token unless it grants more, and any such
workflow is withheld by the guard anyway) and is not a requirement premise.

### E26 — the merge base and the preserve probe (tree ae494aedb2)

```
$ git merge-base main HEAD
4f9dba14bf2834aa08bea79f9e72c2a7e75ab18e
$ git rev-parse main                                 7705b365e2789eaccfde8b1ff488862b56c5d0c0
$ git rev-list --count 4f9dba14bf..main              11
$ git diff --name-only 4f9dba14bf HEAD               9 paths (the seven SPEC files, plan-audit.md, progress.md under .moai/reports/t1868)
```

Local `main` is 11 commits past the card base, and the merge base of `HEAD` and `main` is still the card base,
so AC-CCR-016's left end `A` is the card base today; after the standard absorb-`main` merge it becomes the
absorbed `main` commit (audit defect D9).

### E27 — the Codex app and its bot account (GET, 2026-10-18)

```
$ gh api "apps/chatgpt-codex-connector"
id 1144995, slug chatgpt-codex-connector, name "ChatGPT Codex Connector", owner {login openai, type Organization},
description "Bring ChatGPT and Codex to your GitHub repositories.", permissions {pull_requests write, contents write, issues write,
actions write, workflows write, checks read, statuses read, emails read, metadata read}   (output cut to the deciding fields)
$ gh api "users/chatgpt-codex-connector%5Bbot%5D"
login chatgpt-codex-connector[bot], id 199175422, type Bot, html_url https://github.com/apps/chatgpt-codex-connector
(output cut to the deciding fields)
```

The identity the gate binds is login `chatgpt-codex-connector[bot]`, numeric user id 199175422, account type
Bot. The run-phase pre-flight re-reads it (plan §C).

### E28 — the app is installed on the organisation (GET, 2026-10-18)

```
$ gh api "orgs/modu-ai/installations"
total_count 6; the entry with app_slug chatgpt-codex-connector: installation id 66790488, app_id 1144995, repository_selection all,
suspended_at null, permissions {issues write, actions write, contents write, metadata read, workflows write, pull_requests write}
(also present: vercel, render, claude, coderabbitai, railway-app; output cut to the deciding entry)
```

The app is installed on `modu-ai` for all repositories and not suspended. That the native review is switched
on for the two repositories, runs when a pull request opens, and has the personal-review credit off is the
operator's browser-confirmed state relayed by the leader; this lane did not observe the settings page.

### E29 — no Codex artifact on any recent pull request (GET, 2026-10-18)

```
$ (explicit loop, per_page=100) reviews, review comments, and issue comments of the 60 most recent pull requests of each repository
modu-ai/moai-adk  : 60 pull requests, oldest created 2026-10-06T00:10:05Z
   authors seen: GoosLab, claude[bot], codecov-commenter, coderabbitai[bot], dependabot[bot], github-actions[bot], vercel[bot],
                 github-advanced-security[bot]
   authors whose login contains codex, chatgpt, or openai: none
modu-ai/mo.ai.kr  : 60 pull requests, oldest created 2026-08-26T16:02:12Z
   authors seen: GoosLab, coderabbitai[bot], dependabot[bot], github-actions[bot], vercel[bot]
   authors whose login contains codex, chatgpt, or openai: none
$ the newest pull request, modu-ai/mo.ai.kr#435 (created 2026-10-18T11:17:58Z, not draft): issue comments by vercel[bot] and coderabbitai[bot] only;
  no review, no review comment, no reaction (GET issues/435/reactions -> [])
$ GET commits/<head>/check-runs for the first 12 pull requests of each repository: no check run whose name or app slug contains codex,
  except the github-actions `codex-review` of mo.ai.kr#425 (E24)
```

No review, review comment, or issue comment by the Codex app exists on 120 pull requests, including one opened
the same day; the review was switched on only recently or leaves no artifact of these forms. Which artifact it
leaves is therefore UNVERIFIED (U2, U5), the gate recognises a pull-request review only and fails closed, and
deferred step DS-1 observes the form. A later dated observation (the leader at 2026-10-18T12:05Z, and this
lane at 12:40:08Z; E32 and `spec.md` §5 DS-1) found no Codex trace on a pull request 47 and 82 minutes after it
opened; the holder of that fact is the DS-1 row.

### E30 — revision-3 re-measurement of the RED-now ledger at tree ae494aedb2

```
$ git rev-parse --short HEAD                      ae494aedb2
R1 test -f scripts/ci/codex-review-gate.sh        exit 1
R2 grep -c -i coderabbit auto-merge.yml           14   exit 0
R3 grep -c -F 'issues/$PR_NUMBER/comments' ...    1    exit 0
R4 grep -c 'gate-files-changed' ...               0    exit 1
R5 grep -c 'read-native-review' ...               0    exit 1
R6 grep -c 'fetch-pages' ...                      0    exit 1
R7 grep -rn 'anthropics/claude-code-action' ...   claude.yml:35 and claude.yml:47   exit 0
R8 enabled: 65 ; auto_reply: 74                   exit 0 ; exit 0
R9 control exit 1 ; baseline exit 0
R10 claude.yml present ; review-quality-gate.yml present   exit 0 ; exit 0
R11 grep -c -e claude.yml -e review-quality-gate.yml liveness.yaml   2   exit 0
```

R1, R2, R3, R7, R8, R9, R10, and R11 carry over from the revision-2 measurement at `8cd1d0d51d` unchanged; R4
carries over; R5 and R6 are new greps measured at this tree. In revision 4 R1 to R11 were re-run at tree
`e11af67632` with identical output and R12 (`grep -c -F 'steps.review.outputs.ok'` on `auto-merge.yml`: 0,
exit 1) was added (acceptance §1).

### E31 — changed-file entries carry the commit they were read at (GET, 2026-10-18, revision 4)

Question: can a changed-file list be bound to one commit, so that a head that flipped while the list was read
is detected? Method: a scratchpad script called `gh api` GET and tallied each entry's `status` against the
`ref` in its `contents_url` (`...contents/<path>?ref=<sha>`), compared with the pull request object's
`head.sha` and `base.sha`.

```
modu-ai/moai-adk#1826  head 2e7d1bd373  base.sha f352c0827c  entries 1029 (= changed_files 1029)
   status added    ref == head : 554      status modified ref == head : 452
   status renamed  ref == head : 9        status removed  ref == base : 14
modu-ai/moai-adk#1823  head 610840bf63  base.sha 2aab5f797b  entries 2     (added 1, modified 1: ref == head)
modu-ai/moai-adk#1816  head b5d311c5a9  base.sha db0c514d3b  entries 116   (added 31, modified 85: ref == head)
```

1,147 entries on three pull requests: every non-removed entry (1,015 + 2 + 116) carries `ref` equal to the head;
every removed entry (14) carries the base SHA. The `raw_url` and `blob_url` of an entry carry the head too (the
first-page probe printed all three). Not observed: the statuses `copied`, `changed`, `unchanged`, and a pull
request whose base moved after it opened (U10). Consequence used by the design: `gate-files-changed` requires
the `ref` of every non-removed entry to equal the verified head (design §3.4), and removed entries are not
bound (design §6, residual 4).

### E32 — the DS-1 re-read at 2026-10-18T12:40:08Z (GET, this lane, revision 4)

Question: does the native review leave any artifact on the pull requests opened after the operator's
confirmation? Method: scratchpad script, `gh api` GET of `repos/modu-ai/mo.ai.kr/pulls/<N>`, `.../reviews`,
`.../pulls/<N>/comments`, `.../issues/<N>/comments`, `.../issues/<N>/timeline`, and
`commits/<head>/check-runs`, each with `per_page=100` and a page loop.

```
mo.ai.kr#435  open, draft False, created 2026-10-18T11:17:58Z, age 82 min, head 386f0d6c9c
   reviews 0, review comments 0, issue-comment authors [coderabbitai[bot], vercel[bot]]
   timeline actors [GoosLab, coderabbitai[bot], vercel[bot]], codex/chatgpt/openai actors []
   check-run apps [github-actions, vercel]
mo.ai.kr#436  open, draft False, created 2026-10-18T11:54:19Z, age 45 min, head 31c8074dcf
   reviews 0, review comments 0, issue-comment authors [coderabbitai[bot], vercel[bot]]
   timeline actors [GoosLab, coderabbitai[bot], vercel[bot]], codex/chatgpt/openai actors []
   check-run apps [github-actions, vercel]; names include codex-review (app github-actions: a workflow of that repository, E24)
```

Same result as the leader's read of #435 at 12:05Z (47 minutes: reviews 0, no Codex actor, check runs only
vercel and github-actions). The operator's "automatic review ON when a pull request opens" has not been
observed to produce a review on either pull request; a delay is possible. The form stays UNCONFIRMED (U2, U5)
and is confirmed only when a real review is observed. The dated fact and its direction binding are held by the
DS-1 row of `spec.md` §5.

### E33 — the review trigger and settings (relayed by the leader, revision 6; not measured by this lane)

The leader relayed, from a read-only observation of the native review's settings for this repository, on
2026-10-18 (this lane did not observe the settings page and changes no setting): the review trigger is "when a
pull request is opened"; automatic review is on for all pull requests; thorough review is on; automatic
security review is on; the minimum reported severity is High. What a review with no finding at or above High
leaves on a pull request, and whether a review runs again after the ready-for-review transition or after a push,
are unknown (DS-1, DS-2). Holder of this fact: this entry and the DS-2 row of `spec.md` §5.

### E34 — the `review` step ends failures with exit 0 and writes keys twice (delegated Codex observer reproduction and this lane's re-run, revision 6)

Delegated Codex observer reproduction (`/tmp/adk-queue-repair-20261018/review-step-repro-v2/result.json`, machine-local; run at
source head `67a5149ce2879ef47fa7d50478a3c927f9d22617`, workflow sha256 `8a32f372...`, gate sha256 `2d07cdc4...`;
the exact `run:` body of the step and the real gate script, with `gh` replaced by a fixture stub; no network):

```
case changes-requested  exit_code 0  stdout ::warning::native review gate: changes-requested - withholding merge   outputs "ok=false\nstate=fail\n"
case draft              exit_code 0  stdout ::notice::NOT RUN: draft - neither a failure nor a pass; ...             outputs "ok=false\nstate=fail\nstate=not-run\n"
```

This lane's re-run of the same method against the tree at HEAD `0955fb9058` (workflow sha256 `24cba625...`, gate
sha256 `2d07cdc4...`, the same gate script; the workflow differs from the delegated Codex observer's by the ruling-B recheck),
script in the lane scratchpad, not committed; the step body run with `bash --noprofile --norc -eo pipefail`:

```
changes-requested step_exit=0 outputs='ok=false\nstate=fail\n'
draft             step_exit=0 outputs='ok=false\nstate=fail\nstate=not-run\n'
clean-approved    step_exit=0 outputs='ok=false\nstate=fail\nok=true\nstate=pass\n'
```

The defect is therefore present at HEAD, and a clean pass also writes each key twice. Limits: the run used a
stub `gh` and no runner; the reproduction shows the step's exit status and output file, not the job conclusion
GitHub reports (U11). Re-executable cells for the tree: acceptance ledger R13 and R14.

### E35 — the committed tests assert the pre-ruling exit status (measured, revision 6)

```
$ grep -c -F 'not_run_step_exit_status_0_with_notice_annotation' internal/template/auto_merge_codex_test.go
1
exit: 0     (tree 0955fb9058)
```

Two subtests of the closed AC-CCR-008 (`not_run_step_exit_status_0_with_notice_annotation`,
`fail_step_exit_status_0_with_warning_annotation`) assert exit status 0, and the `files_step_*_not_ok` /
`review_step_*_not_ok` rows of AC-CCR-012 read `ok` only; no committed criterion asserts a non-zero exit or a job
result. Revision 6 replaces them (acceptance AC-CCR-020 to AC-CCR-025).

### E36 — the `files` step ends a missing gate script with exit 0 (delegated read-only observer, revision 6)

Observer's record `/tmp/adk-ci-followthrough-20261018/files-step-missing-script-repro/result.json`
(machine-local; the leader's board record d-20261018T155527Z-4e9e): at 2026-10-18T15:54:27Z, head
`0955fb90585444f4a9679afe82d994d3f1b1a80b`, workflow sha256 `24cba625b248b8050b6957b6b0ddbcaff166e67a7c7627c8e1e7e13be1832e1d`,
the exact `files` step body run with `bash --noprofile --norc -eo pipefail`, the gate script deliberately
absent, no `gh` reached:

```
exit_code 0
stdout    ::warning::changed-file check: gate-script-missing - withholding merge
outputs   "ok=false\nstate=fail\n"
```

The observer's own gap: the GitHub job conclusion is unobserved. Because the `review` step runs only when
`files.ok == 'true'`, a failure caught in `files` ended a green job even if the `review` step were fixed;
revision 6 therefore covers both steps (design `T-CHAIN` rows 10 to 13, AC-CCR-025). Re-executable cell:
acceptance ledger R15.

## 4. Options weighed

| Question | Chosen | Rejected | Evidence |
|----------|--------|----------|----------|
| Reviewer | The native Codex GitHub review of the existing subscription, read through the GitHub API | A key-based `openai/codex-action` workflow of our own (plan revisions 1 and 2) | The operator confirmed the native review is ON for both repositories and runs on pull-request open; the app is installed org-wide (E27, E28). The keyed path adds a secret, a usage cost, a deferred gate (G1), a workflow whose verdict carrier needed binding and a widened tamper guard (audit D1), and a third action-pinning surface; the native review needs none |
| Verdict carrier | A pull-request review by the Codex identity, bound by `commit_id` | An issue comment; a check run; a commit status; review body text | A review object carries `commit_id`, state, submission time, and the author's login, id, and type (E23); an issue comment has no commit binding and any user can post the marker text; a status is writable by any token holding `statuses: write`; no check run of the native app was observed (E29) and a same-named check run is forgeable by a repository workflow |
| Identity binding | Login, numeric id, and type together | Login only; id only | A lookalike or renamed account defeats a single-field match (AC-CCR-003); the three values were measured (E27) |
| Reading lists | An explicit `per_page=100` page loop with a count proof where one exists and a ceiling refusal | One call; `gh api --paginate` | The default page holds 30 entries (E22); the count proof agreed with the list on every pull request measured (E22, E23); `--paginate` carries unverified failure semantics a stub cannot reproduce |
| What counts as clean | State APPROVED only, newest bound review, no finding; a CHANGES_REQUESTED is superseded only by a later APPROVED | COMMENTED or APPROVED with no finding (revision 3: fail-open on a COMMENTED review whose body reports findings, a failure, or a usage-limit notice, and the newest review could override an earlier CHANGES_REQUESTED); any bound CHANGES_REQUESTED final with no supersession | The form of a clean result is unobserved (E29, E32); APPROVED is the one state whose meaning needs no reading of body text, and it makes the audit's defects D12 and D13 one rule (decision Q20) |
| No review at the end of the window | Gate exit 11 `not-run:review-absent`, which the consumer ends as a failed step (revision 6; the closed revision called it NOT RUN, neither failure nor pass) | A pass; a green step; staying pending | The operator's direction: no review is NOT RUN, never success, never failure; time only bounds a wait (E32: no review after 82 minutes, so no longer wait is assumed, and none is ever a pass) |
| Changed-file binding | The `ref` of each non-removed entry's `contents_url` equals the verified head, plus the head read before and after | List read by pull request number; the compare API | Entries carry the commit they were read at (E31); the compare endpoint's file limit was not measured and a three-dot compare is not the pull request's own file list |
| Findings | Any review comment by the identity on the head, or CHANGES_REQUESTED, withholds | Parse a severity from comment text | The native review's finding vocabulary was not observed (U5); a parser would depend on an unobserved text shape and be a widening path |
| Draft pull requests | NOT RUN (exit 11) at once, whatever exists | Read reviews on a draft; wait for a review on a draft | Whether the native review runs on a draft is unconfirmed (U4); a draft is never merged by automation, and a draft observation must not read as a review result |
| Gate script source | Default-branch checkout of one script | Pull-request head checkout; inline steps | `workflow_run` runs the default-branch copy of the workflow; a head copy of the script could pass itself |
| Tamper guard breadth | Any file under `.github/workflows/`, and the gate script | Only the named gate files | A pull request can add a workflow that merges itself or acts with the workflow token (audit D1); human merges are unaffected |
