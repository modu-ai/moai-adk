# SPEC-JEV-RESOLVED-MODEL-001 — Acceptance Criteria

Given-When-Then is the verification layer's format. The requirement wording lives in `spec.md` §C and is not restated here.

Each criterion names its verification command. Each criterion is one of four classes:

- **Release-blocking** (AC-JRM-001 to AC-JRM-011). Each carries a RED-now cell, a green-path cell that names the milestone that flips it, and the `--- PASS: <TestName>` line that the passing run prints (`verification-completeness.md` §2). Exit 0 alone is not a pass, because a selector that matches nothing exits 0.
- **Regression guards** (AC-JRM-012 to AC-JRM-015). Each states the output that must hold, and each carries a GRN cell in the ledger.
- **Sync-phase** (AC-JRM-016). Its flip is the sync commit that REQ-JRM-014 requires, not an M milestone.

Evidence rules for every cell in the ledger: a single read-only invocation, with no pipe, no `&&`, no `;`, and no subshell. Grep-family commands carry the `command` prefix, because the shell tool hides the exit status of a grep that finds nothing; with the prefix, a no-match exit 1 is shown as `Exit code 1`. An exit with no error shown is recorded as exit 0. Base-tree cells use `git grep` pinned to `53a8c5764`. Cells that depend on the final tree are re-observed after the lane's commit, and each such cell says so.

Run each `go test` as a plain invocation: no pipes and no shell-variable constructs. The worktree guard refused a `go test` command that combined a pipe with a `${PIPESTATUS}` construct; plain invocations with anchored selectors were accepted. Every `-run` selector is anchored (`'^Name$'`), because an unanchored selector also selects longer test names, and the spec linter flags it as a vacuous assertion.

---

## §A. Request and attribution

**AC-JRM-001** (REQ-JRM-001) — Given a request built for any question set, When its body is serialized, Then the body's `model` field is `jev-latest` and is not a versioned identifier. The question text is not scanned.
Verification: `go test -count=1 -v -run '^TestRequestBody_CarriesTheAliasAndNoInterpretedVersion$' ./internal/jev/`
Expected green line: `--- PASS: TestRequestBody_CarriesTheAliasAndNoInterpretedVersion (<n>s)`
RED-now: EV-RED-01 (the constant is the versioned pin), EV-RED-02 (the test does not exist yet). Green path: M1.

**AC-JRM-002** (REQ-JRM-002) — Given an endpoint answer whose `model` field is `jev-9.8.7`, When the call completes, Then the result's `Usage.Model` and `Answer.Model` are `jev-9.8.7`, and the client's usage log carries the same value. Given two answered calls in sequence, answered `jev-9.8.7` and then `jev-9.8.8`, Then each usage record carries its own call's id, in order.
Verification: `go test -count=1 -v -run '^TestResult_CarriesInterpretedModelFromResponse$' ./internal/jev/`
Expected green line: `--- PASS: TestResult_CarriesInterpretedModelFromResponse (<n>s)`
RED-now: EV-RED-03 (the record names the constant before the response is read). Green path: M1.

**AC-JRM-003** (REQ-JRM-003) — Given an OK response whose `model` is absent, is `jev-latest`, or is `jev-1.13`, When it is decoded, Then the result's availability is `unattributed`, no answer is returned, and both `Result.Usage` and the client's `UsageLog()` carry the call's record, with an empty interpreted model and `request_model` set to `jev-latest`.
Verification: `go test -count=1 -v -run '^TestResponse_WithoutInterpretedModelIsUnattributed$' ./internal/jev/`
Expected green line: `--- PASS: TestResponse_WithoutInterpretedModelIsUnattributed (<n>s)`
RED-now: EV-RED-04 (no `unattributed` availability exists). Green path: M1.

**AC-JRM-004** (REQ-JRM-004) — Given a completed answered call whose response is `jev-9.8.7`, When its usage record is read, Then the request alias and the interpreted model are held in distinct fields, with the values `jev-latest` and `jev-9.8.7`.
Verification: `go test -count=1 -v -run '^TestUsage_RequestAliasAndInterpretedModelAreSeparateFields$' ./internal/jev/`
Expected green line: `--- PASS: TestUsage_RequestAliasAndInterpretedModelAreSeparateFields (<n>s)`
RED-now: EV-RED-05 (no `RequestModel` field exists). Green path: M1.

**AC-JRM-005** (REQ-JRM-005) — Given an answered call whose response is `jev-9.8.7`, When its rendered answer label is read, Then the label names `jev-9.8.7` and not the request alias, and an answer with an empty `Model` renders `unattributed`. Given the non-test Go source of the four packages, When the constant `ModelID` is referenced, Then every reference is one of the permitted sites below, and no attribution field is assigned the constant.
Verification: `go test -count=1 -v -run '^TestAnswerLabel_NamesTheInterpretedModel$' ./internal/jev/`, then `command grep -rnw --include='*.go' --exclude='*_test.go' 'jev\.ModelID' internal cmd`, then `command grep -nw --include='*.go' --exclude='*_test.go' 'ModelID' internal/jev/jev.go`.
Permitted sites (the allow-list):
- Outside `internal/jev`, the qualified `jev.ModelID` matches exactly two lines: the `RequestModel` assignment in `Run` (`internal/jevmeasure/measure.go`), and the verbose detail of the doctor check (`internal/cli/doctor_jev.go`). Any other match fails the criterion.
- Inside `internal/jev/jev.go`, the unqualified `ModelID` matches exactly four lines: the declaration `const ModelID = "jev-latest"`, the single doc-comment line that names the constant, the request body in `toWire` (`wireRequest{Model: ModelID, ...}`), and the `RequestModel: ModelID` assignment in the usage literal of `(c *Client) Ask`. Any other match fails the criterion.
- The attribution fields `Usage.Model`, `Answer.Model`, `Report.ModelID`, and the kickoff reason carry the interpreted id, or nothing. No line assigns them the constant. This is checked by the behavioural tests of AC-JRM-002, 004, 006, 007, and 009, which assert that each field holds the interpreted id.
- `ModelIDOpus55` in `internal/template` is a different identifier. The word-anchored grep does not match it, so it is not an exception.
Expected: the first grep prints exactly two lines, and the second prints exactly four.
RED-now: EV-RED-06 (the label renders the constant). Green path: the label test at M1, and the allow-list check at M6.

**AC-JRM-006** (REQ-JRM-006) — Given two answered calls in one measurement run, reporting `jev-9.8.7` and `jev-9.8.8`, When the run completes, Then no artifact is returned, and the error names both ids.
Verification: `go test -count=1 -v -run '^TestRun_RefusesAStraddledMeasurement$' ./internal/jevmeasure/`
Expected green line: `--- PASS: TestRun_RefusesAStraddledMeasurement (<n>s)`
RED-now: EV-RED-07 (no straddle check exists). Green path: M2.

**AC-JRM-007** (REQ-JRM-007) — Given a run whose answered calls all report `jev-9.8.7`, When its artifact is rendered, Then the rendering carries `model: jev-9.8.7 (interpreted by the endpoint)` and `request model: jev-latest (alias sent)` as separate lines, and no model line names the alias. Positive control: the same check fires on a rendering whose model line is the alias.
Verification: `go test -count=1 -v -run '^TestReport_RenderCitesInterpretedModelAndAliasOnlyInRequestField$' ./internal/jevmeasure/`
Expected green line: `--- PASS: TestReport_RenderCitesInterpretedModelAndAliasOnlyInRequestField (<n>s)`
RED-now: EV-RED-08 (the render line says `(pinned)`). Green path: M2.

**AC-JRM-008** (REQ-JRM-008) — Given the doctor check's verbose detail, When it is read, Then it contains `request model jev-latest`, and it contains no phrase of the form `model jev-<n>.<n>.<n>`.
Verification: `go test -count=1 -v -run '^TestDoctorJev_VerboseNamesAliasAndNoInterpretedVersion$' ./internal/cli/`
Expected green line: `--- PASS: TestDoctorJev_VerboseNamesAliasAndNoInterpretedVersion (<n>s)`
RED-now: EV-RED-09 (the detail has no `request model` wording). Green path: M5.

**AC-JRM-009** (REQ-JRM-009) — Given a contract-mode Kickoff with the `llm+jev` decider whose Jev answer reports `jev-9.8.7`, When the decision is recorded, Then the receipt's Jev reason names `jev-9.8.7` and does not name the alias.
Verification: `go test -count=1 -v -run '^TestDecide_JevCrossCheckReasonNamesInterpretedModel$' ./internal/contract/kickoff/`
Expected green line: `--- PASS: TestDecide_JevCrossCheckReasonNamesInterpretedModel (<n>s)`
RED-now: EV-RED-10 (the reason cites the constant). Green path: M5.

---

## §C. Version change and rule

**AC-JRM-010** (REQ-JRM-010) — Given the verified id `jev-1.13.0` (`internal/jev/jev.go:VerifiedInterpretedModel`) and an answered call that reports `jev-9.8.7`, When the first such call completes in a client, Then its result carries exactly one notice line, which names both ids, the location `internal/jev/jev.go:VerifiedInterpretedModel`, and the shadow-return obligation. A second answered call in the same client carries no notice. A client whose answered calls report `jev-1.13.0` carries no notice. On the MCP surface, a `jev_ask` call answered `jev-9.8.7` returns a `notice` key, and a `jev_ask` call answered `jev-1.13.0` returns no `notice` key.
Verification: `go test -count=1 -v -run '^TestChangedInterpretedModel_NoticeOncePerClient$' ./internal/jev/`, and `go test -count=1 -v -run '^TestJevAsk_NoticeKeyFollowsTheChangeNotice$' ./internal/cli/`
Expected green lines: `--- PASS: TestChangedInterpretedModel_NoticeOncePerClient (<n>s)` and `--- PASS: TestJevAsk_NoticeKeyFollowsTheChangeNotice (<n>s)`
Status: **adopted** (OQ-1 option A, card t1768, board). RED-now: EV-RED-12 (the verified-id constant is absent on the base tree) and EV-RED-17 (the MCP test is absent on the base tree). Green path: M3, both lines.

**AC-JRM-011** (REQ-JRM-011) — Given `.claude/rules/moai/development/model-policy.md` and its template copy, When each is checked, Then each carries exactly one heading line `## Jev Model Attribution`, each carries the anchored content lines below, the two files are byte-identical, and the mirror test passes.
Anchored content checks (each prints `1` in each file, unless stated):
- `command grep -c '^## Jev Model Attribution$' <file>`
- `command grep -c '^- \*\*Request\.\*\* .*jev-latest' <file>`
- `command grep -c '^- \*\*Unattributed answers are not answers\.\*\*' <file>`
- `command grep -c '^- \*\*No straddled measurement\.\*\*' <file>`
- `command grep -c '^- \*\*Version change\.\*\* .*shadow' <file>`
- `command grep -c '^- \*\*Terms\.\*\* .*advise-or-above' <file>`
- `command grep -c 'does not ship' <file>` prints `0` (the shipping-prohibition clause is absent; OQ-3 (b)).
Then `cmp .claude/rules/moai/development/model-policy.md internal/template/templates/.claude/rules/moai/development/model-policy.md` exits 0, and `go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$/^model-policy.md$' ./internal/template/` prints the green line below.
Positive control: the same content check, run on a copy in the scratchpad from which the `Version change` bullet has been deleted, prints `0`.
Expected green line: `--- PASS: TestRuleTemplateMirrorDrift/model-policy.md (<n>s)`
RED-now: EV-RED-11 (the anchored heading is absent on the base tree). EV-GRN-02 is the guard for the copies' identity (the parity subtest passes on the base tree), so it is not the red-to-green criterion. Green path: M4.

---

## §D. Gate, amendments, and lint

**AC-JRM-012** (REQ-JRM-012) — Given `workflow.jev.enabled` is false, When a call is made, Then no request is constructed and no network call is made. Regression guard.
Verification: `go test -count=1 -v -run '^TestDisabled_ConstructsNoRequest$' ./internal/jev/`
Expected output: `--- PASS: TestDisabled_ConstructsNoRequest (<n>s)`
GRN: EV-GRN-01 (passes on the base tree and on the repair-round tree).

**AC-JRM-013** (REQ-JRM-013) — Given the two amended SPECs and their progress records, When their `sync_commit_sha` lines are read, Then they are `c8731f965` (CORE) and `52c460d30` (OPTIN), and the amendment has not moved them. Regression guard.
Verification: `grep -n 'sync_commit_sha: c8731f965' .moai/specs/SPEC-JEV-CORE-001/progress.md` and `grep -n 'sync_commit_sha: 52c460d30' .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/progress.md`
GRN: EV-GRN-03, EV-GRN-04.

**AC-JRM-014** (REQ-JRM-013) — Given the two amended SPEC bodies, When they are searched by line, Then CORE's `spec.md` carries the frontmatter line `amendment_of: SPEC-JEV-CORE-001` and the `## Amendments` bullet `- prior_completed_sha: c8731f965`, and OPTIN's `spec.md` carries the frontmatter line `amendment_of: SPEC-JEV-OPTIN-MEASURE-001` and the bullet `- prior_completed_sha: 52c460d30`. The anchors are line anchors, so the prose in the HISTORY rows does not match. The key is a bullet of the `## Amendments` section, which is why its anchor is `^- prior_completed_sha: `. Regression guard.
Verification: `grep -n '^amendment_of: SPEC-JEV-CORE-001$' .moai/specs/SPEC-JEV-CORE-001/spec.md`, `grep -n '^- prior_completed_sha: c8731f965$' .moai/specs/SPEC-JEV-CORE-001/spec.md`, `grep -n '^amendment_of: SPEC-JEV-OPTIN-MEASURE-001$' .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md`, and `grep -n '^- prior_completed_sha: 52c460d30$' .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md`.
Historical base-tree observations: EV-RED-13 and EV-RED-14. On `53a8c5764` the keys were absent. They are recorded as history, because the criterion is a guard, not a red-to-green gate.
GRN: EV-GRN-05 to EV-GRN-08 (pinned to `cfb551796`).

**AC-JRM-015** (lint for this SPEC and its amendments; REQ-JRM-013) — Given the three SPEC directories, When `moai spec lint` runs against a binary built from this tree, Then it reports no errors, and `moai spec audit` reports no drift finding for the two amended SPECs. Regression guard.
Verification: build the binary from the tree and invoke it by path, then `<moai-tree> spec lint .moai/specs/SPEC-JEV-RESOLVED-MODEL-001 .moai/specs/SPEC-JEV-CORE-001 .moai/specs/SPEC-JEV-OPTIN-MEASURE-001`, then the audit for each amended SPEC.
GRN: EV-GRN-09 (the tree-built lint, at the repair round; re-observed after the lane's commit) and EV-GRN-10 (the tree the build was made from). The installed build `69c1ffc1f4` lags the tree, and it is not evidence for this criterion.

---

## §E. Sync phase

**AC-JRM-016** (REQ-JRM-014) — Given the sync phase of this SPEC has closed, When each amended `spec.md` is diffed against its amendment-state version (the plan commit of this card, which is `cfb551796` or the lane's commit that carries these artifacts), Then its frontmatter `status` is `completed`, and every changed line lies between the two `---` lines of its frontmatter. The changed lines are the `status:` and `updated:` lines.
Verification (sync phase, after the sync commit): `grep -n '^status: completed$' .moai/specs/SPEC-JEV-CORE-001/spec.md`, the same grep for `.moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md`, and `git diff --unified=0 <plan-commit> HEAD -- .moai/specs/SPEC-JEV-CORE-001/spec.md` (and the OPTIN file). Each grep prints one line, and every hunk of each diff lies inside the frontmatter.
RED-now: EV-RED-15 and EV-RED-16 (pinned to `cfb551796`, where both amended SPECs are `in-progress`). Green path: the sync-phase commit of manager-docs, which is not an M milestone.

---

## Evidence ledger

Each entry records the command as run, the verbatim stdout, the exit status, and the tree the command read. `exit: 0` means the tool showed no error; `exit: 1` means the tool showed `Exit code 1`. A tab is shown as a tab.

### RED-now cells (base tree `53a8c5764`)

```text
EV-RED-01
  command: command git grep -n 'const ModelID = "jev-1.13.0"' 53a8c5764 -- internal/jev/jev.go
  stdout:  53a8c5764:internal/jev/jev.go:77:const ModelID = "jev-1.13.0"
  exit:    0
  tree:    53a8c5764

EV-RED-02
  command: command git grep -c 'func TestRequestBody_CarriesTheAliasAndNoInterpretedVersion' 53a8c5764 -- internal/jev/jev_test.go
  stdout:  (none)
  exit:    1
  tree:    53a8c5764

EV-RED-03
  command: command git grep -n 'Usage{InputTokens: requestTokens(req), Model: ModelID}' 53a8c5764 -- internal/jev/jev.go
  stdout:  53a8c5764:internal/jev/jev.go:523:	usage := Usage{InputTokens: requestTokens(req), Model: ModelID}
  exit:    0
  tree:    53a8c5764

EV-RED-04
  command: command git grep -c 'Unattributed' 53a8c5764 -- internal/jev/jev.go
  stdout:  (none)
  exit:    1
  tree:    53a8c5764

EV-RED-05
  command: command git grep -c 'RequestModel' 53a8c5764 -- internal/jev/jev.go
  stdout:  (none)
  exit:    1
  tree:    53a8c5764

EV-RED-06
  command: command git grep -n 'SignalLabel, ModelID' 53a8c5764 -- internal/jev/jev.go
  stdout:  53a8c5764:internal/jev/jev.go:218:	prefix := fmt.Sprintf("[%s · %s] %s", SignalLabel, ModelID, a.QuestionID)
  exit:    0
  tree:    53a8c5764

EV-RED-07
  command: command git grep -c 'straddle' 53a8c5764 -- internal/jevmeasure/measure.go
  stdout:  (none)
  exit:    1
  tree:    53a8c5764

EV-RED-08
  command: command git grep -n 'model: %s (pinned)' 53a8c5764 -- internal/jevmeasure/measure.go
  stdout:  53a8c5764:internal/jevmeasure/measure.go:250:	fmt.Fprintf(&b, "model: %s (pinned)\n", r.ModelID)
  exit:    0
  tree:    53a8c5764

EV-RED-09
  command: command git grep -c 'request model' 53a8c5764 -- internal/cli/doctor_jev.go
  stdout:  (none)
  exit:    1
  tree:    53a8c5764

EV-RED-10
  command: command git grep -n 'jev.ModelID' 53a8c5764 -- internal/contract/kickoff/decide.go
  stdout:  53a8c5764:internal/contract/kickoff/decide.go:183:					Reason: "Jev cross-check signal (" + jev.ModelID + ").", ReasonRefs: j.ReasonRefs},
  exit:    0
  tree:    53a8c5764

EV-RED-11
  command: command git grep -c '^## Jev Model Attribution$' 53a8c5764 -- .claude/rules/moai/development/model-policy.md
  stdout:  (none)
  exit:    1
  tree:    53a8c5764

EV-RED-12
  command: command git grep -c 'VerifiedInterpretedModel' 53a8c5764 -- internal/jev/jev.go
  stdout:  (none)
  exit:    1
  tree:    53a8c5764

EV-RED-13 (historical; AC-JRM-014 is a guard)
  command: command git grep -c 'prior_completed_sha' 53a8c5764 -- .moai/specs/SPEC-JEV-CORE-001/spec.md
  stdout:  (none)
  exit:    1
  tree:    53a8c5764

EV-RED-14 (historical; AC-JRM-014 is a guard)
  command: command git grep -c 'prior_completed_sha' 53a8c5764 -- .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md
  stdout:  (none)
  exit:    1
  tree:    53a8c5764

EV-RED-15 (AC-JRM-016; the pre-sync state)
  command: command git grep -n '^status: completed$' cfb551796 -- .moai/specs/SPEC-JEV-CORE-001/spec.md
  stdout:  (none)
  exit:    1
  tree:    cfb551796

EV-RED-16 (AC-JRM-016; the pre-sync state)
  command: command git grep -n '^status: completed$' cfb551796 -- .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md
  stdout:  (none)
  exit:    1
  tree:    cfb551796

EV-RED-17 (AC-JRM-010, the MCP sub-assertion; the test is absent on the base tree)
  command: command git grep -c 'func TestJevAsk_NoticeKeyFollowsTheChangeNotice' 53a8c5764 -- internal/cli/mcp_jev_test.go
  stdout:  (none)
  exit:    1
  tree:    53a8c5764
```

### GRN cells (guards)

```text
EV-GRN-01 (AC-JRM-012; re-observed at the repair round; the jev package has no .go change, so the tree is the base's code; re-observe after the lane's commit)
  command: command go test -count=1 -v -run '^TestDisabled_ConstructsNoRequest$' ./internal/jev/
  stdout:  === RUN   TestDisabled_ConstructsNoRequest
           --- PASS: TestDisabled_ConstructsNoRequest (0.00s)
           PASS
           ok  	github.com/modu-ai/moai-adk/internal/jev	0.381s
  exit:    0
  tree:    cfb551796 with the repair-round SPEC edits uncommitted

EV-GRN-02 (AC-JRM-011 guard, the copies' identity; re-observed at the repair round; re-observe after the lane's commit)
  command: command go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$/^model-policy.md$' ./internal/template/
  stdout:  === RUN   TestRuleTemplateMirrorDrift
           === PAUSE TestRuleTemplateMirrorDrift
           === CONT  TestRuleTemplateMirrorDrift
           === RUN   TestRuleTemplateMirrorDrift/model-policy.md
           === PAUSE TestRuleTemplateMirrorDrift/model-policy.md
           === CONT  TestRuleTemplateMirrorDrift/model-policy.md
           --- PASS: TestRuleTemplateMirrorDrift (0.13s)
               --- PASS: TestRuleTemplateMirrorDrift/model-policy.md (0.01s)
           PASS
           ok  	github.com/modu-ai/moai-adk/internal/template	0.963s
  exit:    0
  tree:    cfb551796 with the repair-round SPEC edits uncommitted

EV-GRN-03 (AC-JRM-013)
  command: command git grep -n 'sync_commit_sha: c8731f965' cfb551796 -- .moai/specs/SPEC-JEV-CORE-001/progress.md
  stdout:  cfb551796:.moai/specs/SPEC-JEV-CORE-001/progress.md:320:sync_commit_sha: c8731f965
  exit:    0
  tree:    cfb551796

EV-GRN-04 (AC-JRM-013)
  command: command git grep -n 'sync_commit_sha: 52c460d30' cfb551796 -- .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/progress.md
  stdout:  cfb551796:.moai/specs/SPEC-JEV-OPTIN-MEASURE-001/progress.md:277:sync_commit_sha: 52c460d30
  exit:    0
  tree:    cfb551796

EV-GRN-05 (AC-JRM-014)
  command: command git grep -n '^amendment_of: SPEC-JEV-CORE-001$' cfb551796 -- .moai/specs/SPEC-JEV-CORE-001/spec.md
  stdout:  cfb551796:.moai/specs/SPEC-JEV-CORE-001/spec.md:15:amendment_of: SPEC-JEV-CORE-001
  exit:    0
  tree:    cfb551796

EV-GRN-06 (AC-JRM-014)
  command: command git grep -n '^- prior_completed_sha: c8731f965$' cfb551796 -- .moai/specs/SPEC-JEV-CORE-001/spec.md
  stdout:  cfb551796:.moai/specs/SPEC-JEV-CORE-001/spec.md:34:- prior_completed_sha: c8731f965
  exit:    0
  tree:    cfb551796

EV-GRN-07 (AC-JRM-014)
  command: command git grep -n '^amendment_of: SPEC-JEV-OPTIN-MEASURE-001$' cfb551796 -- .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md
  stdout:  cfb551796:.moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md:16:amendment_of: SPEC-JEV-OPTIN-MEASURE-001
  exit:    0
  tree:    cfb551796

EV-GRN-08 (AC-JRM-014)
  command: command git grep -n '^- prior_completed_sha: 52c460d30$' cfb551796 -- .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md
  stdout:  cfb551796:.moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md:32:- prior_completed_sha: 52c460d30
  exit:    0
  tree:    cfb551796

EV-GRN-09 (AC-JRM-015; the tree-built lint; <scratchpad> is the session scratchpad directory, and the binary was built from the worktree by `go build -o <scratchpad>/moai-tree ./cmd/moai`)
  command: <scratchpad>/moai-tree spec lint .moai/specs/SPEC-JEV-RESOLVED-MODEL-001 .moai/specs/SPEC-JEV-CORE-001 .moai/specs/SPEC-JEV-OPTIN-MEASURE-001
  stdout:  SEVERITY  CODE                  FILE                                            LINE  MESSAGE
           --------  ----                  ----                                            ----  -------
           WARNING   StatusGitConsistency  .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md  1     SPEC SPEC-JEV-OPTIN-MEASURE-001 frontmatter status 'in-progress' disagrees with git-implied status 'implemented'
           
           0 error(s), 1 warning(s)
  exit:    0
  tree:    cfb551796 with the repair-round SPEC edits uncommitted; re-observe after the lane's commit
  note:    the first run at this round reported a VacuousTestAssertion warning on the Definition of Done line that quoted the AC-JRM-010 PASS line without a trailing space; that line was corrected, and the warning is gone in this run. The StatusGitConsistency warning is pre-existing and was reported by the audit at cfb551796.

EV-GRN-10 (AC-JRM-015; the build's tree)
  command: command git rev-parse --short HEAD
  stdout:  cfb551796
  exit:    0
  tree:    cfb551796
```

### Green-path cells (specified here; observed at the named milestone)

```text
AC-JRM-001  M1  --- PASS: TestRequestBody_CarriesTheAliasAndNoInterpretedVersion (<n>s)
AC-JRM-002  M1  --- PASS: TestResult_CarriesInterpretedModelFromResponse (<n>s)
AC-JRM-003  M1  --- PASS: TestResponse_WithoutInterpretedModelIsUnattributed (<n>s)
AC-JRM-004  M1  --- PASS: TestUsage_RequestAliasAndInterpretedModelAreSeparateFields (<n>s)
AC-JRM-005  M1  --- PASS: TestAnswerLabel_NamesTheInterpretedModel (<n>s); M6: the two allow-list counts (2 and 4)
AC-JRM-006  M2  --- PASS: TestRun_RefusesAStraddledMeasurement (<n>s)
AC-JRM-007  M2  --- PASS: TestReport_RenderCitesInterpretedModelAndAliasOnlyInRequestField (<n>s)
AC-JRM-008  M5  --- PASS: TestDoctorJev_VerboseNamesAliasAndNoInterpretedVersion (<n>s)
AC-JRM-009  M5  --- PASS: TestDecide_JevCrossCheckReasonNamesInterpretedModel (<n>s)
AC-JRM-010  M3  --- PASS: TestChangedInterpretedModel_NoticeOncePerClient (<n>s); --- PASS: TestJevAsk_NoticeKeyFollowsTheChangeNotice (<n>s)
AC-JRM-011  M4  --- PASS: TestRuleTemplateMirrorDrift/model-policy.md (<n>s); anchored checks print 1
AC-JRM-016  sync-phase commit (not an M milestone)  status: completed in both amended spec.md files; diff inside the frontmatter
```

Green-path cells are specified, not observed. Each is re-observed at its milestone and again after the lane's commit.

---

## Gaps (not observed by this plan stage)

- The alias behaviour (`jev-latest` answered as `jev-1.13.0`, with the same answer and tokens) is the leader's observation of 2026-10-15. This stage made no call to the endpoint, and no credential is present in the tree. Every run-stage test uses fixtures or a local server. A live observation is a separate record, with its own command and output.
- AC-JRM-010 is adopted, and its green lines are not yet observed (M3). AC-JRM-016 is a sync-phase criterion, so it is not observed at plan stage.
- The rule's content checks in AC-JRM-011 are specified; they are observed at M4.
- The installed build lags the tree (`69c1ffc1f4` is a strict ancestor of `53a8c5764`). The tree-built lint and audit are the evidence for AC-JRM-015.
- The `command` prefix makes a no-match exit status visible. The earlier plan-stage cells (plain `grep` on the working tree, with the installed build) are superseded by the cells above.
- The change notice reaches the MCP output (`mcp_jev.go:114–116`, read at plan stage), and AC-JRM-010 asserts its `notice` key there; that assertion is not observed before M3. The CLI consumers, the `--auto` ranking and the Kickoff cross-check, do not surface the notice on success: a gap for follow-up card t1803, not designed or closed in this SPEC (plan §H).

## Residual risk (could be wrong despite the observations above)

- A vendor change to the `model` field's shape makes every consumer unavailable (`unattributed`) until the shape is accepted. Consumers that show or select on a Jev answer then show or select nothing, and the Kickoff path falls to the LLM-alone decision of rule R2 (plan §G), not to a closed gate. That is the designed behaviour, and it is visible to users.
- The alias can move to a new version at any time. The notice is carried on the result of the first answered call in each client: a new process repeats it, and so does each `jev_ask` call, because each builds its own client. Until the constant is bumped it keeps repeating, and the rule (REQ-JRM-011) describes the shadow return that follows. The CLI consumers do not surface it on success (Gaps).
- `model-policy.md` is path-scoped to agent files (OQ-4), so its Jev section will not load for `internal/jev` work unless the lead widens the scope.
- AC-JRM-005's allow-list is a line-count check over `grep`. A new site that adds the same word on a permitted line would be caught only by the line count and by review.

## Definition of Done (run stage)

- AC-JRM-001 through AC-JRM-011: each verification command exits 0 and prints its expected `--- PASS: <TestName>` line. AC-JRM-011 also prints `1` for each anchored check, `cmp` exits 0, and the `TestRuleTemplateMirrorDrift/model-policy.md` PASS line is printed. Each RED-now cell has a recorded green counterpart in the run-stage record (§E.3 of progress.md).
- AC-JRM-010 is adopted (OQ-1 option A, card t1768). Its green lines `--- PASS: TestChangedInterpretedModel_NoticeOncePerClient (<n>s)` and `--- PASS: TestJevAsk_NoticeKeyFollowsTheChangeNotice (<n>s)` are required at M3. Exit 0 alone does not satisfy them.
- AC-JRM-012 through AC-JRM-015: each guard prints its stated output. The GRN cells are re-observed after the lane's commit.
- AC-JRM-016: at the sync phase, the two `status: completed` greps and the frontmatter-only diff check hold.
- The §F gates of plan.md pass: gofmt prints nothing, vet is clean, each package gate prints `ok`, the named `internal/cli` selector prints `ok`, the receipt gate prints `ok`, `make build` exits 0, and the Windows build exits 0.
- The rule text (plan.md §D.4) is byte-identical in both copies (`cmp` exit 0).
- `sync_commit_sha` values are unchanged: `c8731f965` in CORE and `52c460d30` in OPTIN.
- The clarification-marker check of plan.md §F prints nothing and exits 1 at plan close and at run close.
