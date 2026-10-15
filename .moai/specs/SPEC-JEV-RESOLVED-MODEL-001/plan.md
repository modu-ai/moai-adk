# SPEC-JEV-RESOLVED-MODEL-001 — Plan

Card t1768 · Class C · Template-First · P1 · Tier M. Plan-phase author: manager-spec. Tree at authoring: `53a8c5764` (clean).

The decisions come first, ordered by how likely each is to change: the data model, the change notice, the rule wording, the amendment mechanics, and the dependency declaration. Mechanical steps (comments, citation sites, fixtures) come last.

---

## §A. Context and tier

- The request model moves from the pinned id `jev-1.13.0` to the alias `jev-latest`. The response's `model` field becomes the recorded attribution (spec.md §A).
- Tier M. Run-stage file count is 15 (§D.1). The tier table places 5–15 files in Tier M. The 300 LOC Tier S band does not hold, because the file count is over 5. Requirements: 13, ceiling 16. Acceptance criteria: 15, ceiling 16.
- Artifact set: spec.md, plan.md, acceptance.md, progress.md. No design.md or research.md. Tier M does not name them, and the open questions are carried as markers in §G.

---

## §B. Decisions

### D1 — Attribution data model (decided; card-driven)

- The `ModelID` constant keeps its identifier, as the card names it, and takes the value `jev-latest`. It is read only where the request is built (`toWire`). Every citation site moves to the interpreted id. AC-JRM-005 enforces that with a grep.
- `Usage` gains `RequestModel`. `Usage.Model` holds the interpreted id taken from the response. It is empty when the response is not attributable.
- The usage record is written after the send, not before it. Every call that was sent is recorded, whatever its outcome. That keeps `TestUsage_RecordedEvenWhenTheCallDidNotSucceed` true. A call refused before the send (gate, credential, size bound, secret screen) is not recorded, which keeps `TestUsage_NotRecordedWhenNoRequestWasConstructed` true.
- `Answer` gains `Model` (the interpreted id), so that `Label()` can name the version that answered.
- New availability `unattributed`. It applies to an OK HTTP response whose `model` is missing or is outside the versioned shape. No answers are returned with it.
- Rejected: (a) keeping the pre-send record and patching it after the response, which writes one call twice; (b) renaming the constant to `RequestModel`, which is a larger diff, and the card names `ModelID`.

### D2 — Change notice (OPEN: OQ-1; blocks AC-JRM-010 only)

Three mechanisms can meet REQ-JRM-010. The choice changes whether a file is written.

- **Option A (recommended).** A compiled validated id in `internal/jev`, the interpreted id the regression set was last run against. The first answered call in a client whose interpreted id differs gets one notice line on its result. The validated id is bumped by a code change, after the regression set is re-run. No file is written, so REQ-JEVC-011 and the display-only test (`display_only_test.go`) stay as they are. The cost: a process that starts after a vendor change emits the notice until someone bumps the constant.
- **Option B.** A state record under `.moai/state/`, written by the CLI layer after each answered call and compared on the next call. This writes a file as a consequence of an answer. REQ-JEVC-011 would need an explicit carve-out. The inert-sibling carve-out does not fit, because the record is not a sibling of any object it names.
- **Option C.** An in-process comparison only. It sees a change inside a long-lived process such as the MCP server, and it never sees one between CLI invocations. It does not satisfy REQ-JRM-010 for CLI consumers.

Until OQ-1 is answered, the run stage does not start AC-JRM-010. spec.md §E ("persisted model history") assumes Option A. If the lead chooses Option B, that exclusion and REQ-JRM-010 change together.

### D3 — Rule wording (text in §D.4; shadow and regression-set terms carry OQ-2 and OQ-3)

- The section sits in `.claude/rules/moai/development/model-policy.md`, as the card requires, and is copied byte-identically to the template tree.
- Reachability: the rule's frontmatter is `paths: "**/.claude/agents/**"`. The section therefore loads for agent-file edits and not for `internal/jev` work. See OQ-4.

### D4 — Amendment transition (card-instructed)

- Both amended SPECs move `completed → in-progress (amendment)`. Each carries `amendment_of` set to itself, a HISTORY row, and a `## Amendments` section with a `prior_completed_sha` line. This follows the card and the transition row of `spec-frontmatter-schema.md` § Status Enum.
- The audit exemption needs all four conditions: `status: in-progress`, `amendment_of` set, an `## Amendments` section, and `prior_completed_sha` equal to the prior close's `sync_commit_sha`. The tree-built audit verifies them after the edit (§F).
- Precedent differs: CORE 0.2.0–0.4.0 kept `status: completed`, as a non-transition correction. This card follows the card's instruction and not that precedent (OQ-5).
- `sync_commit_sha` is untouched in this commit. It stays `c8731f965` for CORE and `52c460d30` for OPTIN, and REQ-JRM-013 keeps it so through the sync phase.

### D5 — Dependency declaration: `related_specs`, not `depends_on`

- The Depends_on pre-flight (`spec-workflow-detail.md` § Depends_on Pre-flight Check) counts a dependency as fulfilled only at `status: completed`. The two amended SPECs are `in-progress` until this card's sync. A `depends_on` declaration would therefore block this card's own run-phase entry. The relation is declared as `related_specs`, which is non-blocking.

### D6 — Commit shape

- One commit carries the four new SPEC files and the three edited SPEC files (CORE spec.md; OPTIN spec.md and acceptance.md). The card asks for the plan artifacts in one commit, and the amendments must travel with it.
- Trailer layout: the body paragraphs, then `Authored-By-Agent: manager-spec` alone in its own paragraph, then `🗿 MoAI` as the last paragraph. The lint reads the trailer with its own regex over the commit body (`internal/spec/lint_ownership.go`), so the layout is for git's parser and the MoAI marker convention.

### D7 — Vision (not triggered)

- `.moai/project/product.md` contains no `jev`, no `pinned`, and no `pin` in its Vision section. The card's conditional vision revision does not apply.

---

## §C. Pre-flight (plan-stage evidence; verbatim)

Executed on tree `53a8c5764`, worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1768`.

| Check | Command | Verbatim result |
|---|---|---|
| SPEC ID pattern | `ID=SPEC-JEV-RESOLVED-MODEL-001; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS \|\| echo FAIL` | `PASS` |
| Target directory absent before write | `ls -d .moai/specs/SPEC-JEV-RESOLVED-MODEL-001` | `No such file or directory` (exit 2) |
| Baseline lint, installed build | `moai spec lint .moai/specs/SPEC-JEV-CORE-001 .moai/specs/SPEC-JEV-OPTIN-MEASURE-001` | `✓ No findings — all SPEC documents are valid` |
| Baseline audit, CORE (MCP) | `mcp__moai__spec_audit`, filter SPEC-JEV-CORE-001 | `modern_era_clean: 1`; one `EraAutoDetected` INFO; no drift finding |
| Baseline audit, OPTIN (MCP) | `mcp__moai__spec_audit`, filter SPEC-JEV-OPTIN-MEASURE-001 | `modern_era_clean: 1`; one `EraAutoDetected` INFO; no drift finding |
| Installed build provenance | `git merge-base --is-ancestor 69c1ffc1f4 HEAD && echo ANCESTOR_YES \|\| echo ANCESTOR_NO` | `ANCESTOR_YES` |

**Provenance statement (verification-claim-integrity §2.2).** The installed `moai` is `v3.2.0-rc.32`, build commit `69c1ffc1f4` (`moai version`). That commit is a strict ancestor of tree HEAD `53a8c5764`, so the installed build lags the tree. The baseline results above are therefore citations of a lagging build. The lint and audit results that gate this card are re-measured with a binary built from this tree (§F, `moai-tree`). The MCP server reports build `v3.2.0-rc.31`, which is a separate build; its results are cited as corroboration only.

**Measured on the base tree (red-now and green-now evidence).** The full ledger is in acceptance.md § Evidence ledger. The gate-off and mirror guards were run as plain invocations with anchored selectors: `go test -count=1 -v -run '^TestDisabled_ConstructsNoRequest$' ./internal/jev/` gave `--- PASS: TestDisabled_ConstructsNoRequest (0.00s)` with `ok`, and `go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$/^model-policy.md$' ./internal/template/` gave `--- PASS: TestRuleTemplateMirrorDrift/model-policy.md (0.01s)` with `ok`.

---

## §D. Run-stage work list

### §D.1 Files (15)

| # | File | Change |
|---|---|---|
| 1 | `internal/jev/jev.go` | Data model, attribution, notice seam (D2), comments |
| 2 | `internal/jev/jev_test.go` | Replace pin tests; add attribution tests |
| 3 | `internal/jevmeasure/measure.go` | Report fields, Run attribution and straddle refusal, Validate, Render |
| 4 | `internal/jevmeasure/measure_test.go` | Fixtures and the render test |
| 5 | `internal/jevmeasure/edges_test.go` | Fixtures; Validate rejects the alias in the model field |
| 6 | `internal/jevmeasure/gate_demo_test.go` | Fixture and the cited id |
| 7 | `internal/jevmeasure/request.go` | Comment only (line 7, "under a pinned") |
| 8 | `internal/cli/doctor_jev.go` | Verbose detail wording (around line 114) |
| 9 | `internal/cli/doctor_jev_test.go` | New detail assertion |
| 10 | `internal/cli/mcp_jev_test.go` | Assert `usage.model` and `usage.request_model` on the tool output |
| 11 | `internal/contract/kickoff/decide.go` | Reason uses the answer's interpreted id (around line 183) |
| 12 | `internal/contract/kickoff/decide_test.go` | New reason assertion |
| 13 | `internal/config/types.go` | Comment only (lines 616 and 958, "pinned model id") |
| 14 | `.claude/rules/moai/development/model-policy.md` | New section and one Rules bullet (§D.4) |
| 15 | `internal/template/templates/.claude/rules/moai/development/model-policy.md` | Byte-identical copy of file 14 |

Not in the run stage: `CHANGELOG.md`, `docs-site/`, the README family, `.moai/docs/*` (sync phase; manager-docs), and every other SPEC directory's body (historical records, not rewritten).

### §D.2 Functions and types (internal/jev/jev.go)

- `const ModelID = "jev-latest"` (was line 77). Doc comment rewritten: the request alias, never a citation of the answering model.
- `var interpretedModelShape = regexp.MustCompile(`^jev-[0-9]+\.[0-9]+\.[0-9]+$`)` and `func ValidInterpretedModel(s string) bool`.
- `Availability`: add `Unattributed Availability = "unattributed"`.
- `Answer`: add `Model string` with `json:"model,omitempty"`.
- `Usage`: add `RequestModel string` with `json:"request_model"`. `Model string json:"model"` documents the interpreted id.
- `decode(resp)` and `(c *Client) send(...)` return the response's `model` value as a third result. `fromWire(questions, answers, model)` sets `Answer.Model`.
- `(c *Client) Ask`: build the request as before. After `send`, if the response is OK and `ValidInterpretedModel(model)` is false, return `Result{Availability: Unattributed, Condition: ...}` with no answers. Then `usage := Usage{InputTokens: requestTokens(req), RequestModel: ModelID, Model: validated-or-empty}`, `c.record(usage)` once, and `res.Usage = usage`.
- `(a Answer) Label()`: prefix uses `a.Model`; an empty `Model` renders `unattributed`.
- Package doc and the `@MX:ANCHOR` reason (lines 5–6) and comments at lines 60–77, 214–218, 293, 330–331: rewritten. The MX reason changes from "the model pin" to "the request alias and the attribution".

### §D.3 Functions (internal/jevmeasure/measure.go)

- `Report`: add `RequestModel string`.
- `Run`: `rep.RequestModel = jev.ModelID`. Collect `result.Usage.Model` for each answered call into a set. After the loop: more than one id returns `fmt.Errorf` naming the ids (REQ-JRM-006). Exactly one sets `rep.ModelID`. Zero leaves it empty.
- `Validate`: `ModelID` must satisfy `jev.ValidInterpretedModel` (REQ-JEVO-014 as amended, REQ-JRM-007).
- `Render`: `model: <ModelID> (interpreted by the endpoint)` and `request model: <RequestModel> (alias sent)` replace the `(pinned)` line at 250.

### §D.4 Rule wording — `.claude/rules/moai/development/model-policy.md` (file 14)

Insert the section below before `## Rules`. Add the bullet at the end of `## Rules`. Copy the whole file to file 15 and confirm `cmp` exit 0.

```markdown
## Jev Model Attribution

[ZONE:Evolvable] [HARD] A Jev answer names the model that produced it. The request alias is what the package asks for, and it is never the name of the model that answered.

- **Request.** The `internal/jev` package sends the request alias `jev-latest`. The alias is a request, not an attribution.
- **Attribution.** The `model` field of the response is the interpreted model id, for example `jev-1.13.0`. Every model field a consumer reads carries the interpreted model id: the usage record's `model`, a measurement artifact's `model`, a rendered answer label, and a kickoff cross-check reason. The request alias appears only in its own field: `request_model` in a usage record, and `request model` in a measurement artifact.
- **Unattributed answers are not answers.** A response with no interpreted model id in the shape `jev-<major>.<minor>.<patch>` is an unavailable result (`unattributed`). No consumer reads an answer from it.
- **No straddled measurement.** A measurement run whose answered calls report more than one interpreted model id is refused. Re-run it under one id.
- **Version change.** When the interpreted model id differs from the id the measurement regression set was last run against, the first answered call in each client reports one notice line naming both ids. Until the regression set has been re-run against the new id, every consumer that shows a Jev answer to a person, or selects on one, returns to shadow: its answer is recorded, and it is neither shown nor selected on.
- **Regression set.** The regression set is the labelled answer set that `internal/jevmeasure` runs for each consumer. A consumer with no committed labelled set has no measurement, and it does not ship.

Why: the prohibition on the alias rested on the risk that the answering model could not be attributed. Recording the response's `model` field removes that risk. The vendor can move the version behind the alias at any time, and the notice and the shadow return make such a move visible rather than silent.
```

Rules bullet: `- Jev answers name the interpreted model the endpoint returns; the request alias \`jev-latest\` is never an attribution (§ Jev Model Attribution)`

The terms "shadow", "shows or selects", and "regression set" are the subject of OQ-2 and OQ-3 (§G). The section text is written so that it can be confirmed or amended without changing the attribution rules.

### §D.5 Surfaces

- **doctor_jev.go (file 8).** Verbose detail: `request model <jev.ModelID> (alias sent; the interpreted version is recorded per answered call and is not probed here); endpoint <jev.EndpointURL>; readiness only — no judgment request is sent by this check`.
- **decide.go (file 11).** The reason is `"Jev cross-check signal (" + jres.Usage.Model + ")."`. `jres` is in scope at that call site. The `fallback == ""` branch implies an OK result, and so an attributable one. The receipt already stores the marshalled `jev.Result` as `RawResponse`, so the interpreted id reaches the receipt without further change.
- **mcp_jev.go (no code change).** The output's `usage` object gains `request_model`. The tool description has no pin wording. File 10 asserts the new keys.

### §D.6 Tests (names are the run-stage contract)

internal/jev/jev_test.go:
- Replace `TestModelID_IsPinnedNotAnAlias` (line 74) with `TestRequestModel_IsTheAliasNotAVersion`: asserts `ModelID == "jev-latest"`.
- Replace `TestModelID_MatchesVendorVersionedShape` (line 92) with `TestInterpretedModel_MatchesVendorVersionedShape`: accepts `jev-1.13.0`; rejects `jev-1.13`, `jev-latest`, and the empty string.
- Replace `TestRequestBody_CarriesPinnedIDAndNoAlias` (line 107) with `TestRequestBody_CarriesTheAliasAndNoInterpretedVersion`: the body carries `"model":"jev-latest"` (positive control) and no `jev-<n>.<n>.<n>`.
- Replace `TestUsage_RecordsInputTokensAndPinnedModelPerCall` (line 387) with `TestUsage_RecordsInterpretedModelAndRequestAliasPerCall`.
- Update `TestUsage_RecordedEvenWhenTheCallDidNotSucceed` (line 413): the record exists, `Model` is empty, `RequestModel` is set.
- New: `TestResult_CarriesInterpretedModelFromResponse`, `TestResponse_WithoutInterpretedModelIsUnattributed` (table: missing, `jev-latest`, `jev-1.13`), `TestUsage_RequestAliasAndInterpretedModelAreSeparateFields`, `TestAnswerLabel_NamesTheInterpretedModel`.
- Update `TestAnswerLabel_IsDistinguishableFromMeasurementAndJudgement` (line 536) to keep its distinguishability property with the interpreted id.
- Update `TestDecode_DocumentedResponseShape` (line 768) to assert the decoded `model`.
- Adjust `TestRequestBody_IsTheDocumentedMapShape` (line 685) wording only; its comparison already reads `ModelID`.

internal/jevmeasure/measure_test.go:
- Fixture at line 55: the answerer's `Usage` carries the interpreted id and `RequestModel` carries the alias.
- `TestRun_ProducesBothArmsAndDelta` (line 140): assert `rep.ModelID` equals the fixture's interpreted id and `rep.RequestModel` equals `jev.ModelID`.
- Fixtures at lines 199 and 228: versioned fixture constant.
- `TestReport_RenderCitesPinnedModelAndCarriesNoAlias` (line 379) renamed `TestReport_RenderCitesInterpretedModelAndAliasOnlyInRequestField`.
- New: `TestRun_RefusesAStraddledMeasurement`, `TestRun_CitesTheInterpretedModelAcrossArms`.

internal/jevmeasure/edges_test.go and gate_demo_test.go: fixtures (lines 61, 81, 107; 61–62). `TestValidate_RejectsMissingModelAndMismatchedThreshold` (line 72) gains a case: the alias in the model field is rejected.

internal/cli/doctor_jev_test.go: new `TestDoctorJev_VerboseNamesAliasAndNoInterpretedVersion`.

internal/cli/mcp_jev_test.go: add assertions on `usage.model` (`jev-1.13.0`, from the fake server at line 93) and `usage.request_model` (`jev-latest`).

internal/contract/kickoff/decide_test.go: new `TestDecide_JevCrossCheckReasonNamesInterpretedModel` (fixture helpers in `fixture_test.go`).

internal/jev (OQ-1 option A only, blocked): new `TestChangedInterpretedModel_NoticeOncePerClient`.

### §D.7 Naming and wording rules for the run stage

- Use the plain identifiers in §D.2. Do not introduce `Pinned*`, `pinned`, or `resolved` in new identifiers or comments.
- Run each `go test` as a plain single invocation: no pipes, no shell-variable constructs. The worktree guard refused a command that combined a pipe with `${PIPESTATUS}`; plain invocations with anchored selectors were accepted. Anchor every `-run` selector (`'^Name$'`, and `'^A|^B'` for a family), because the spec linter flags an unanchored selector as a vacuous assertion.
- No emoji. No time estimates. English only in the rule and the code comments.

---

## §E. Milestones (ordered by decision reversibility)

- **M1 — data model and attribution** (`internal/jev`). Write the §D.6 tests first, observe them red (the E8 RED evidence), then implement §D.2. Gates: the `internal/jev` package tests.
- **M2 — measurement attribution** (`internal/jevmeasure`). Straddle refusal, Validate, Render. Gates: the `internal/jevmeasure` package tests.
- **M3 — change notice** (`internal/jev`, option A only). Blocked until OQ-1 is answered. If it is not answered before the run stage closes, the run stage records AC-JRM-010 as unadopted and does not claim it.
- **M4 — rule text and parity.** Edit file 14, copy to file 15, `cmp`, then `make build`. Gate: `TestRuleTemplateMirrorDrift/model-policy.md` passes.
- **M5 — surfaces.** Doctor (files 8–9), kickoff (files 11–12), MCP (file 10). Gates: the `internal/cli` jev family and `internal/contract/kickoff`.
- **M6 — mechanical sweep.** Comments (files 7, 13; `jev.go` comments), citation sites, and fixtures. Gate: the AC-JRM-005 grep, and no `jev.ModelID` citation left outside request construction.
- **M7 — verification and evidence.** §F gates. Run-phase evidence (§E.2, §E.3) is written by manager-develop. Sync evidence (§E.4) is written by manager-docs.

---

## §F. Verification gates (run stage; plain invocations)

```text
gofmt -l internal/jev internal/jevmeasure internal/cli/doctor_jev.go internal/cli/doctor_jev_test.go internal/cli/mcp_jev_test.go internal/contract/kickoff internal/config/types.go
go vet ./internal/jev/ ./internal/jevmeasure/ ./internal/contract/kickoff/
go test -count=1 ./internal/jev/
go test -count=1 ./internal/jevmeasure/
go test -count=1 ./internal/contract/kickoff/
go test -count=1 ./internal/cli/ -run '^(TestJev.*|TestDoctorJev.*)$'
go test -count=1 ./internal/template/ -run '^TestRuleTemplateMirrorDrift$'
make build
GOOS=windows GOARCH=amd64 go build ./...
```

The affected `internal/cli` family is run by name, not as the whole suite. The card-worktree CLI suite is known to be structurally red in a worktree, so a whole-suite run cannot serve as evidence here.

Lint and audit gates, with the tree-built binary (built from the tree in this run, and invoked by path):

```text
go build -o <scratchpad>/moai-tree ./cmd/moai
<scratchpad>/moai-tree spec lint .moai/specs/SPEC-JEV-RESOLVED-MODEL-001 .moai/specs/SPEC-JEV-CORE-001 .moai/specs/SPEC-JEV-OPTIN-MEASURE-001
<scratchpad>/moai-tree spec audit --filter-spec SPEC-JEV-CORE-001
```

The `spec audit` flag form is checked against `moai-tree spec audit --help` before use. The MCP audit is cited as corroboration only, with its build commit.

---

## §G. Open questions

These markers are the open decisions. Each is a clarification the orchestrator resolves before the run stage enters the affected item. The plan-auditor treats an unresolved marker as a clarification gate.

[NEEDS CLARIFICATION: OQ-1 — the mechanism behind REQ-JRM-010 (the change notice). Option A (compiled validated id; one notice per client; no file write; REQ-JEVC-011 untouched) is recommended. Option B (a CLI-written state record under `.moai/state/`) needs an explicit REQ-JEVC-011 carve-out, because it writes a file as a consequence of an answer. Option C (in-process only) does not satisfy REQ-JRM-010 for CLI consumers. Blocks AC-JRM-010 only.]

[NEEDS CLARIFICATION: OQ-2 — the shadow-return clause uses terms that have no Jev definition anywhere in the repository. "Advise" exists only as a layer name in `internal/hook/agent_model_guard.go` (an unrelated observation layer), and "shadow" appears only in a launcher comment. The §D.4 text defines the terms by behaviour (a consumer that shows an answer to a person, or selects on it) and introduces no config key. Confirm the definitions, or supply the mode vocabulary the lead intends. Blocks the wording of REQ-JRM-011, not the attribution rules.]

[NEEDS CLARIFICATION: OQ-3 — the "measurement regression set" that the shadow-return clause names is not a committed artifact. No labelled set exists in the repository, and the live measurement has not been run (SPEC-JEV-OPTIN-MEASURE-001 Q3 is OPEN). The §D.4 text defines it as the per-consumer labelled sets, with the consequence that a consumer without a committed set does not ship. Confirm. Blocks the re-run procedure in REQ-JRM-011.]

Non-blocking items (recorded, not gating):

- **OQ-4 — reachability.** `model-policy.md` is path-scoped to `**/.claude/agents/**`. The Jev section does not load for `internal/jev` work. Either keep the card's placement and accept the gap, or widen the `paths:` key to include `**/internal/jev/**` and `**/internal/jevmeasure/**`. The card names the file, so this plan keeps the placement.
- **OQ-5 — amendment transition precedent.** CORE 0.2.0–0.4.0 kept `status: completed` as a non-transition correction. This card follows its instruction (`completed → in-progress (amendment)`). If the lead prefers the precedent, the change is limited to the status line and `amendment_of`, and REQ-JRM-013 still holds.
- **OQ-6 — sync-phase documentation.** `.moai/docs/jev-local-operations.md`, `kickoff-autonomy.md`, and `stale-card-premise-check.md` mention Jev. None states the pin or the alias. manager-docs reviews them at sync, and writes the CHANGELOG entry.
- **OQ-7 — dispatch.** Card t1762 depends on this card's attribution fields. Its plan must cite them and must be planned after this card lands.

---

## §H. Residual risk and gaps

- **Live confirmation is a gap.** The alias-to-version behaviour (`jev-latest` → `jev-1.13.0`, same answer and tokens) is the leader's observation of 2026-10-15. This stage made no call to the endpoint, and no credential is present in this tree. The run-stage tests use fixtures and a local server. Live confirmation is a separate observation, recorded by the run stage with its own command and output, or it stays a gap.
- **Fail-closed on a shape change.** If the vendor changes the `model` field's shape, every consumer becomes unavailable (`unattributed`) until the shape is accepted. That is the designed behaviour, and it is service-visible.
- **Version changes are noticed, not prevented.** The alias can move at any time. Until OQ-1 is settled, nothing announces a change.

---

## §I. Cross-references

- `.moai/specs/SPEC-JEV-CORE-001/spec.md` — REQ-JEVC-003, REQ-JEVC-004 (amended in this commit); REQ-JEVC-011 (display-only, unchanged); REQ-JEVC-017 (gate, unchanged); REQ-JEVC-022 (doctor, unchanged).
- `.moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md` — REQ-JEVO-009, REQ-JEVO-014 (amended); §B (amended).
- `.moai/specs/SPEC-JEV-OPTIN-MEASURE-001/acceptance.md` — AC-JEVO-010, AC-JEVO-011 (amended).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum (amendment exemption), § Status Transition Ownership Matrix.
- `.claude/rules/moai/workflow/spec-workflow-detail.md` § Depends_on Pre-flight Check (why D5).
- `.claude/rules/moai/development/verification-completeness.md` §2 (RED-now cells; evidence ledger in acceptance.md).
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 (tool provenance; §C above).
- `.claude/rules/moai/development/model-policy.md` (file 14) and its template copy (file 15).
- `.claude/rules/moai/workflow/rule-loading-budget.md` — the file is path-scoped; its size is 29,950 bytes before this change (both copies). The run stage records the byte size after the change.
