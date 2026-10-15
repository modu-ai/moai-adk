# SPEC-JEV-RESOLVED-MODEL-001 — Acceptance Criteria

Given-When-Then is the verification layer's format. The requirement wording lives in `spec.md` §C and is not restated here.

Each criterion names its verification command. A release-blocking criterion carries a RED-now cell and a green-path cell in the evidence ledger at the end of this file (`verification-completeness.md` §2.1: command, verbatim stdout, exit code, tree SHA). The RED-now cells were observed on the base tree `53a8c5764` on 2026-10-15, before any run-stage change. Read-only `grep` cells are single invocations.

Run each `go test` as a plain invocation: no pipes and no shell-variable constructs. The worktree guard refused a `go test` command that combined a pipe with a `${PIPESTATUS}` construct; plain invocations with anchored selectors were accepted. Every `-run` selector is anchored (`'^Name$'`), because an unanchored selector also selects longer test names, and the spec linter flags it as a vacuous assertion.

---

## §A. Request and attribution

**AC-JRM-001** (REQ-JRM-001) — Given a request built for any question set, When its body is serialized, Then the body's `model` is `jev-latest`, and no versioned identifier appears in the body.
Verification: `go test -count=1 -v -run '^TestRequestBody_CarriesTheAliasAndNoInterpretedVersion$' ./internal/jev/`
Evidence: EV-RED-01 (the constant is the versioned pin), EV-RED-02 (the test does not exist yet).

**AC-JRM-002** (REQ-JRM-002) — Given an endpoint answer whose `model` field is `jev-1.13.0`, When the call completes, Then the result's `Usage.Model` is `jev-1.13.0`, and the client's usage log carries the same value.
Verification: `go test -count=1 -v -run '^TestResult_CarriesInterpretedModelFromResponse$' ./internal/jev/`
Evidence: EV-RED-03 (the record names the constant before the response is read).

**AC-JRM-003** (REQ-JRM-003) — Given an OK response whose `model` is absent, is `jev-latest`, or is `jev-1.13`, When it is decoded, Then the result's availability is `unattributed`, no answer is returned, and the usage record carries an empty interpreted model.
Verification: `go test -count=1 -v -run '^TestResponse_WithoutInterpretedModelIsUnattributed$' ./internal/jev/`
Evidence: EV-RED-04 (no `unattributed` availability exists).

**AC-JRM-004** (REQ-JRM-004) — Given a completed answered call, When its usage record is read, Then the request alias and the interpreted model are held in distinct fields, with the values `jev-latest` and `jev-1.13.0`.
Verification: `go test -count=1 -v -run '^TestUsage_RequestAliasAndInterpretedModelAreSeparateFields$' ./internal/jev/`
Evidence: EV-RED-05 (no `RequestModel` field exists).

**AC-JRM-005** (REQ-JRM-005) — Given the Jev answer label, When a rendered answer is read, Then the label names the interpreted model and not the request alias. Given the Go source, When `ModelID` is searched outside the tests, Then the only remaining matches are the declaration and the request construction (`toWire`).
Verification: `go test -count=1 -v -run '^TestAnswerLabel_NamesTheInterpretedModel$' ./internal/jev/`, then `grep -rn 'ModelID' internal/jev internal/jevmeasure internal/cli internal/contract --include=*.go` (review each non-test match: it is the declaration, `toWire`, or a test).
Evidence: EV-RED-06 (the label renders the constant).

---

## §B. Measurement and surfaces

**AC-JRM-006** (REQ-JRM-006) — Given two answered calls in one measurement run, reporting `jev-1.13.0` and `jev-1.14.0`, When the run completes, Then no artifact is returned, and the error names both ids.
Verification: `go test -count=1 -v -run '^TestRun_RefusesAStraddledMeasurement$' ./internal/jevmeasure/`
Evidence: EV-RED-07 (no straddle check exists).

**AC-JRM-007** (REQ-JRM-007) — Given a run whose answered calls all report `jev-1.13.0`, When its artifact is rendered, Then the rendering carries `model: jev-1.13.0` and `request model: jev-latest` as separate lines, and no model line names the alias. Positive control: the same check fires on a rendering whose model line is the alias.
Verification: `go test -count=1 -v -run '^TestReport_RenderCitesInterpretedModelAndAliasOnlyInRequestField$' ./internal/jevmeasure/`
Evidence: EV-RED-08 (the render line says `(pinned)`).

**AC-JRM-008** (REQ-JRM-008) — Given the doctor check's verbose detail, When it is read, Then it contains `request model jev-latest`, and it contains no phrase of the form `model jev-<n>.<n>.<n>`.
Verification: `go test -count=1 -v -run '^TestDoctorJev_VerboseNamesAliasAndNoInterpretedVersion$' ./internal/cli/`
Evidence: EV-RED-09 (the detail has no `request model` wording).

**AC-JRM-009** (REQ-JRM-009) — Given a contract-mode Kickoff with the `llm+jev` decider whose Jev answer reports `jev-1.13.0`, When the decision is recorded, Then the receipt's Jev reason names `jev-1.13.0` and does not name the alias.
Verification: `go test -count=1 -v -run '^TestDecide_JevCrossCheckReasonNamesInterpretedModel$' ./internal/contract/kickoff/`
Evidence: EV-RED-10 (the reason cites the constant).

---

## §C. Version change and rule

**AC-JRM-010** (REQ-JRM-010) — **UNADOPTED.** Gated on OQ-1 (plan.md §G). Given a validated interpreted id of `jev-1.13.0` and an answered call reporting `jev-1.14.0`, When the first such call completes in a client, Then its result carries exactly one notice line naming both ids and the shadow-return obligation, and a second call in the same client carries none.
Verification, once adopted: `go test -count=1 -v -run '^TestChangedInterpretedModel_NoticeOncePerClient$' ./internal/jev/`
Evidence: EV-RED-12 (no notice exists). The criterion is recorded neither as passing nor as failing until OQ-1 is settled.

**AC-JRM-011** (REQ-JRM-011) — Given `.claude/rules/moai/development/model-policy.md` and its template copy, When each is searched for the section `Jev model attribution`, Then both carry it, and the two files are byte-identical.
Verification: `grep -c 'Jev model attribution' .claude/rules/moai/development/model-policy.md`, then `cmp .claude/rules/moai/development/model-policy.md internal/template/templates/.claude/rules/moai/development/model-policy.md`, then `go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$/^model-policy.md$' ./internal/template/`.
Evidence: EV-RED-11 (the section is absent); EV-GRN-02 (the parity subtest passes on the base tree, so it is a guard, not a red-to-green criterion).

---

## §D. Gate, amendments, and lint

**AC-JRM-012** (REQ-JRM-012) — Given `workflow.jev.enabled` is false, When a call is made, Then no request is constructed and no network call is made. This is a regression guard.
Verification: `go test -count=1 -v -run '^TestDisabled_ConstructsNoRequest$' ./internal/jev/`
Evidence: EV-GRN-01 (passes on the base tree).

**AC-JRM-013** (REQ-JRM-013) — Given the two amended SPECs and their progress records, When their `sync_commit_sha` lines are read, Then they are `c8731f965` (CORE) and `52c460d30` (OPTIN). The regression guard is the value, and it must not move.
Verification: `grep -n 'sync_commit_sha: c8731f965' .moai/specs/SPEC-JEV-CORE-001/progress.md` and `grep -n 'sync_commit_sha: 52c460d30' .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/progress.md`
Evidence: EV-GRN-03, EV-GRN-04.

**AC-JRM-014** (REQ-JRM-013) — Given the two amended SPEC bodies, When searched for the amendment keys, Then CORE's `spec.md` carries `amendment_of: SPEC-JEV-CORE-001` and `prior_completed_sha: c8731f965`, and OPTIN's `spec.md` carries `amendment_of: SPEC-JEV-OPTIN-MEASURE-001` and `prior_completed_sha: 52c460d30`.
Verification: `grep -n 'prior_completed_sha: c8731f965' .moai/specs/SPEC-JEV-CORE-001/spec.md`, `grep -n 'amendment_of: SPEC-JEV-CORE-001' .moai/specs/SPEC-JEV-CORE-001/spec.md`, `grep -n 'prior_completed_sha: 52c460d30' .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md`, `grep -n 'amendment_of: SPEC-JEV-OPTIN-MEASURE-001' .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md`
Evidence: EV-RED-13 and EV-RED-14 (the key is absent on the base tree).

**AC-JRM-015** (lint for this SPEC and its amendments) — Given the three SPEC directories, When `moai spec lint` runs against a binary built from this tree, Then it reports no errors, and `moai spec audit` reports no drift finding for the two amended SPECs.
Verification: build the binary from the tree and invoke it by path, then `<moai-tree> spec lint .moai/specs/SPEC-JEV-RESOLVED-MODEL-001 .moai/specs/SPEC-JEV-CORE-001 .moai/specs/SPEC-JEV-OPTIN-MEASURE-001`, then the audit for each amended SPEC.
Evidence: EV-GRN-05 (the installed build `69c1ffc1f4`, a lagging build; a provenance caveat applies). The tree-built results are recorded in progress.md §E.1 at plan close.

---

## Evidence ledger

Each entry records the command as run, the verbatim stdout, the exit code, and the tree SHA. Tree `53a8c5764` is the base for every RED-now cell. A tab is shown as a tab in the stdout lines below.

```text
EV-RED-01
  command: grep -n 'const ModelID = "jev-1.13.0"' internal/jev/jev.go
  stdout:  77:const ModelID = "jev-1.13.0"
  exit:    0
  tree:    53a8c5764

EV-RED-02
  command: grep -c 'func TestRequestBody_CarriesTheAliasAndNoInterpretedVersion' internal/jev/jev_test.go
  stdout:  0
  exit:    1
  tree:    53a8c5764

EV-RED-03
  command: grep -n 'Usage{InputTokens: requestTokens(req), Model: ModelID}' internal/jev/jev.go
  stdout:  523:	usage := Usage{InputTokens: requestTokens(req), Model: ModelID}
  exit:    0
  tree:    53a8c5764

EV-RED-04
  command: grep -c 'Unattributed' internal/jev/jev.go
  stdout:  0
  exit:    1
  tree:    53a8c5764

EV-RED-05
  command: grep -c 'RequestModel' internal/jev/jev.go
  stdout:  0
  exit:    1
  tree:    53a8c5764

EV-RED-06
  command: grep -n 'SignalLabel, ModelID' internal/jev/jev.go
  stdout:  218:	prefix := fmt.Sprintf("[%s · %s] %s", SignalLabel, ModelID, a.QuestionID)
  exit:    0
  tree:    53a8c5764

EV-RED-07
  command: grep -c 'straddle' internal/jevmeasure/measure.go
  stdout:  0
  exit:    1
  tree:    53a8c5764

EV-RED-08
  command: grep -n 'model: %s (pinned)' internal/jevmeasure/measure.go
  stdout:  250:	fmt.Fprintf(&b, "model: %s (pinned)\n", r.ModelID)
  exit:    0
  tree:    53a8c5764

EV-RED-09
  command: grep -c 'request model' internal/cli/doctor_jev.go
  stdout:  0
  exit:    1
  tree:    53a8c5764

EV-RED-10
  command: grep -n 'jev.ModelID' internal/contract/kickoff/decide.go
  stdout:  183:					Reason: "Jev cross-check signal (" + jev.ModelID + ").", ReasonRefs: j.ReasonRefs},
  exit:    0
  tree:    53a8c5764

EV-RED-11
  command: grep -c 'Jev model attribution' .claude/rules/moai/development/model-policy.md
  stdout:  0
  exit:    1
  tree:    53a8c5764

EV-RED-12
  command: grep -c 'interpreted model changed' internal/jev/jev.go
  stdout:  0
  exit:    1
  tree:    53a8c5764

EV-RED-13
  command: grep -c 'prior_completed_sha' .moai/specs/SPEC-JEV-CORE-001/spec.md
  stdout:  0
  exit:    1
  tree:    53a8c5764

EV-RED-14
  command: grep -c 'prior_completed_sha' .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md
  stdout:  0
  exit:    1
  tree:    53a8c5764

EV-GRN-01
  command: go test -count=1 -v -run '^TestDisabled_ConstructsNoRequest$' ./internal/jev/
  stdout:  === RUN   TestDisabled_ConstructsNoRequest
           --- PASS: TestDisabled_ConstructsNoRequest (0.00s)
           PASS
           ok  	github.com/modu-ai/moai-adk/internal/jev	0.308s
  exit:    0
  tree:    53a8c5764

EV-GRN-02
  command: go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$/^model-policy.md$' ./internal/template/
  stdout:  === RUN   TestRuleTemplateMirrorDrift
           === PAUSE TestRuleTemplateMirrorDrift
           === CONT  TestRuleTemplateMirrorDrift
           === RUN   TestRuleTemplateMirrorDrift/model-policy.md
           === PAUSE TestRuleTemplateMirrorDrift/model-policy.md
           === CONT  TestRuleTemplateMirrorDrift/model-policy.md
           --- PASS: TestRuleTemplateMirrorDrift (0.00s)
               --- PASS: TestRuleTemplateMirrorDrift/model-policy.md (0.01s)
           PASS
           ok  	github.com/modu-ai/moai-adk/internal/template	0.484s
  exit:    0
  tree:    53a8c5764

EV-GRN-03
  command: grep -n 'sync_commit_sha: c8731f965' .moai/specs/SPEC-JEV-CORE-001/progress.md
  stdout:  320:sync_commit_sha: c8731f965
  exit:    0
  tree:    53a8c5764

EV-GRN-04
  command: grep -n 'sync_commit_sha: 52c460d30' .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/progress.md
  stdout:  277:sync_commit_sha: 52c460d30
  exit:    0
  tree:    53a8c5764

EV-GRN-05
  command: moai spec lint .moai/specs/SPEC-JEV-CORE-001 .moai/specs/SPEC-JEV-OPTIN-MEASURE-001
  stdout:  ✓ No findings — all SPEC documents are valid
  exit:    0
  tree:    53a8c5764
  provenance: installed build v3.2.0-rc.32, commit 69c1ffc1f4, a strict ancestor of 53a8c5764 (lagging build; re-measured with a tree build for AC-JRM-015)
```

---

## Gaps (not observed by this plan stage)

- The alias behaviour (`jev-latest` answered as `jev-1.13.0`, with the same answer and tokens) is the leader's observation of 2026-10-15. This stage made no call to the endpoint, and no credential is present in the tree. Every run-stage test uses fixtures or a local server. A live observation is a separate record, with its own command and output.
- AC-JRM-010 is unadopted (OQ-1). Its behaviour is not claimed.
- The rule's shadow and regression-set terms (OQ-2, OQ-3) have not been confirmed.
- The installed build lags the tree (`69c1ffc1f4` is a strict ancestor of `53a8c5764`). The tree-built lint and audit are the evidence for AC-JRM-015.

## Residual risk (could be wrong despite the observations above)

- A vendor change to the `model` field's shape makes every consumer unavailable (`unattributed`) until the shape is accepted. That is the fail-closed design, and it is visible to users.
- The alias can move to a new version at any time. Until OQ-1 is settled, nothing announces the move at runtime; the rule (REQ-JRM-011) describes the response.
- `model-policy.md` is path-scoped to agent files (OQ-4), so its Jev section will not load for `internal/jev` work unless the lead widens the scope.

## Definition of Done (run stage)

- AC-JRM-001 through AC-JRM-009, AC-JRM-011 through AC-JRM-015: each verification command exits 0 with the stated output, and each RED-now cell has a recorded green counterpart in the run-stage progress record.
- AC-JRM-010: adopted only after OQ-1 is settled. Otherwise it stays unadopted in the run-stage record.
- The §F gates of plan.md pass: gofmt empty, vet clean, the package test runs, the Jev family in `internal/cli`, `make build`, and the Windows build.
- The rule text (plan.md §D.4) is byte-identical in both copies.
- `sync_commit_sha` values are unchanged: `c8731f965` in CORE, `52c460d30` in OPTIN.
