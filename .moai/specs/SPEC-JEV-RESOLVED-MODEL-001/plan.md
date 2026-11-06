# SPEC-JEV-RESOLVED-MODEL-001 — Plan

Card t1768 · Class C · Template-First · P1 · Tier M. Plan-phase author: manager-spec. Tree at authoring: `53a8c5764` (clean). Repair round 1 applied after plan-audit iteration 1 (FAIL, overall 0.65, audited commit `cfb5517962088b09936dfbb8a2ff5b80e63a8893`). The per-finding status of that round is in progress.md.

The decisions come first, ordered by how likely each is to change: the data model, the change notice, the rule wording, the amendment mechanics, and the dependency declaration. Mechanical steps (comments, citation sites, fixtures) come last.

---

## §A. Context and tier

- The request model moves from the pinned id `jev-1.13.0` to the alias `jev-latest`. The response's `model` field becomes the recorded attribution (spec.md §A).
- Tier M. Run-stage file count is 17 (§D.1): 12 files that import `internal/jev` and change, and 5 files that do not import it. The tier table places 5–15 files in Tier M. The count is two over that range, and the line below records why Tier M is kept. The 300 LOC Tier S band does not hold, because the file count is over 5.
- 영향 파일 16~17개는 가이드 상한을 1~2개 넘지만 단일 도메인·단일 결정이라 research/design이 더하는 정보가 없음. card:t1768 (brain, 2026-10-15)
- Requirements: 14, ceiling 16. Acceptance criteria: 16, ceiling 16.
- Artifact set: spec.md, plan.md, acceptance.md, progress.md. No design.md or research.md. Tier M does not name them. The three open decisions are settled, and §G records them.

---

## §B. Decisions

### D1 — Attribution data model (decided; card-driven)

- The `ModelID` constant keeps its identifier, as the card names it, and takes the value `jev-latest`. It is read only at the permitted sites of AC-JRM-005: its declaration, `toWire`, the `RequestModel` assignments in `(c *Client) Ask` and in `Run`, and the doctor detail. No attribution field receives its value. `Usage.Model`, `Answer.Model`, `Report.ModelID`, and the kickoff reason carry the interpreted id, or nothing.
- `Usage` gains `RequestModel`. `Usage.Model` holds the interpreted id taken from the response. It is empty when the call is not an answered call.
- The usage record is written after the send, not before it. Every call that was sent is recorded, whatever its outcome. That keeps `TestUsage_RecordedEvenWhenTheCallDidNotSucceed` true. A call refused before the send (gate, credential, size bound, secret screen) is not recorded, which keeps `TestUsage_NotRecordedWhenNoRequestWasConstructed` true.
- `Answer` gains `Model` (the interpreted id), so that `Label()` can name the version that answered. An empty `Model` renders `unattributed`.
- New availability `unattributed`. It applies to an OK HTTP response whose `model` is missing or is outside the versioned shape. No answers are returned with it.
- Rejected: (a) keeping the pre-send record and patching it after the response, which writes one call twice; (b) renaming the constant to `RequestModel`, which is a larger diff, and the card names `ModelID`.

### D2 — Change notice (decided: OQ-1 option A, card t1768)

- A compiled constant `VerifiedInterpretedModel` in `internal/jev/jev.go` holds the verified id `jev-1.13.0`. Its comment is dated 2026-10-15 and names the request alias `jev-latest` and the response id `jev-1.13.0` (REQ-JRM-010, condition 1). The comment does not name `ModelID`.
- The first answered call in a client whose interpreted id differs from the constant gets one notice line on its result, `Result.Notice`. `NoticeLine()` returns the notice when it is set, so existing callers read it through the method they already use. The notice names both ids and `internal/jev/jev.go:VerifiedInterpretedModel` (REQ-JRM-010, condition 2). It carries no generic instruction.
- Per client: a `changeNoticed` flag on `Client`, read and set under `c.mu`, the mutex that already guards the usage log. A second answered call in the same client carries no notice.
- No file is written. REQ-JEVC-011 and `internal/jev/display_only_test.go` stay as they are.
- The constant is bumped by a code change, after the measurement regression set is re-run against the new id (REQ-JRM-011). The cost of option A is the one the option states: a process that starts after a vendor change emits the notice until someone bumps the constant.
- Options B and C are not chosen. §G records why.

### D3 — Rule wording (decided: OQ-2 option (a), OQ-3 option (b), card t1768)

- The section sits in `.claude/rules/moai/development/model-policy.md`, as the card requires, and is copied byte-identically to the template tree.
- The shadow and advise-or-above terms are defined by behaviour, with no configuration key (OQ-2 (a)). The text in §D.4 is the decided wording.
- The shipping-prohibition clause is not in the rule and not in REQ-JRM-011 (OQ-3 (b)). The re-run procedure compares against the verified id: re-run the regression set against the new id, then change `VerifiedInterpretedModel` to that id.
- Reachability: the rule's frontmatter is `paths: "**/.claude/agents/**"`. The section therefore loads for agent-file edits and not for `internal/jev` work. OQ-4 keeps the placement, and it is non-blocking.

### D4 — Amendment transition (card-instructed)

- Both amended SPECs move `completed → in-progress (amendment)`. Each carries `amendment_of` set to itself in its frontmatter, a HISTORY row, and a `## Amendments` section with a `prior_completed_sha` line. This follows the card and the transition row of `spec-frontmatter-schema.md` § Status Enum.
- The audit exemption needs all four conditions: `status: in-progress`, `amendment_of` set, an `## Amendments` section, and `prior_completed_sha` equal to the prior close's `sync_commit_sha`. The tree-built audit verifies them after the edit (§F).
- Precedent differs: CORE 0.2.0–0.4.0 kept `status: completed`, as a non-transition correction. This card follows the card's instruction and not that precedent (OQ-5).
- `sync_commit_sha` is untouched in this commit. It stays `c8731f965` for CORE and `52c460d30` for OPTIN. REQ-JRM-013 keeps it so through the sync phase.
- At the sync phase, the amended SPECs return to `completed` by a change confined to their frontmatter (REQ-JRM-014, AC-JRM-016). The sync record of this amendment is `progress.md` §E.4 of this SPEC. That sentence is procedure rather than requirement, so it stays here and in spec.md §D.

### D5 — Dependency declaration: `related_specs`, not `depends_on`

- The Depends_on pre-flight (`spec-workflow-detail.md` § Depends_on Pre-flight Check) counts a dependency as fulfilled only at `status: completed`. The two amended SPECs are `in-progress` until this card's sync. A `depends_on` declaration would therefore block this card's own run-phase entry. The relation is declared as `related_specs`, which is non-blocking.

### D6 — Commit shape

- One commit carries the four new SPEC files and the three edited SPEC files (CORE spec.md; OPTIN spec.md and acceptance.md). The card asks for the plan artifacts in one commit, and the amendments must travel with it. The repair round changes only the SPEC directory of this card and the OPTIN acceptance file; the lane's single commit carries them.
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

### §D.1 Files (17)

| # | File | Change |
|---|---|---|
| 1 | `internal/jev/jev.go` | Data model, attribution, the `VerifiedInterpretedModel` constant with its dated comment and the notice (D2), comments |
| 2 | `internal/jev/jev_test.go` | Replace pin tests; add attribution and notice tests; fixture ids per §D.7 |
| 3 | `internal/jevmeasure/measure.go` | Report fields, Run attribution and straddle refusal, Validate, Render |
| 4 | `internal/jevmeasure/measure_test.go` | Fixtures and the render test |
| 5 | `internal/jevmeasure/edges_test.go` | Fixtures; Validate rejects the alias in the model field |
| 6 | `internal/jevmeasure/gate_demo_test.go` | Fixture and the cited id |
| 7 | `internal/jevmeasure/request.go` | Comment only (line 7, "under a pinned") |
| 8 | `internal/cli/doctor_jev.go` | Verbose detail wording (line 114) |
| 9 | `internal/cli/doctor_jev_test.go` | New detail assertion |
| 10 | `internal/cli/mcp_jev_test.go` | Fixture id and assertions on `usage.model` and `usage.request_model`; the `notice` key assertion of AC-JRM-010 (§D.6). `internal/cli/mcp_jev.go` writes `res.NoticeLine()` into `notice` (lines 114–116) and marshals `res.Usage` whole (line 119), so it needs no code change |
| 11 | `internal/contract/kickoff/decide.go` | Reason uses the answer's interpreted id (line 183) |
| 12 | `internal/contract/kickoff/decide_test.go` | New reason assertion; unattributed fallback assertion |
| 13 | `internal/config/types.go` | Comment only (lines 616 and 958, "pinned model id") |
| 14 | `.claude/rules/moai/development/model-policy.md` | New section and one Rules bullet (§D.4) |
| 15 | `internal/template/templates/.claude/rules/moai/development/model-policy.md` | Byte-identical copy of file 14 |
| 16 | `internal/cli/todo_auto_rank_test.go` | Fixture gains a versioned model; an unattributed case (D11) |
| 17 | `internal/cli/todo_auto_rank.go` | New fallback reason `jev-unattributed` for the new availability (D11) |

**Importer recount (this repair round).** `grep -rln 'moai-adk/internal/jev"' internal cmd` lists 23 files. Twelve of them change: rows 3–12, 16, and 17. The other five files in the list are not importers and change as well: rows 1, 2, 13, 14, and 15. The 11 importers that do not change are `internal/contract/receipt/store_test.go`; `internal/contract/kickoff/jev.go`, `kickoff.go`, and `units_test.go`; `internal/jevmeasure/request_test.go`; `internal/cli/mcp_jev.go`, `todo_jev_finding.go`, `todo_jev_finding_test.go`, `todo_auto_test.go`, `todo_auto_lane_test.go`, and `factory_nominate_test.go`. Each one is covered by a §F gate. Their `jev.` uses are fixtures with zero `Usage` and `Answer` values, or injected results with no model value (checked by `grep -n 'Usage\|\.Model\b'` in this round, which found no assertion on a model value).

### §D.2 Functions and types (internal/jev/jev.go)

- `const ModelID = "jev-latest"` (was line 77). Its doc comment names the constant on exactly one line, which states that it is the request alias and never a citation of the answering model. This keeps the AC-JRM-005 allow-list exact.
- `const VerifiedInterpretedModel = "jev-1.13.0"`, with the dated comment of D2. The comment does not name `ModelID`. A comment next to the constant says that it is bumped by a code change after the measurement regression set is re-run.
- `var interpretedModelShape = regexp.MustCompile(`^jev-[0-9]+\.[0-9]+\.[0-9]+$`)` and `func ValidInterpretedModel(s string) bool`. The doc comment of `ValidInterpretedModel` does not name `ModelID`.
- `Availability`: add `Unattributed Availability = "unattributed"`.
- `Answer`: add `Model string` with `json:"model,omitempty"`.
- `Usage`: add `RequestModel string` with `json:"request_model"`. `Model string json:"model"` documents the interpreted id, which is empty when the call is not an answered call.
- `Result`: add `Notice string`. It is set only on the change notice. `NoticeLine()` returns `Notice` when it is set, and otherwise behaves as it does now.
- `Client`: add `changeNoticed bool`, read and set under `c.mu`.
- `wireResponse.Model` already exists (jev.go:355) and is discarded at `decode` (line 626). `decode(resp)` and `(c *Client) send(...)` return the response's `model` as a third result. `fromWire(questions, answers, model)` sets `Answer.Model`.
- `(c *Client) Ask`: build the request as before. After `send`, decide availability and attribution: if the response is OK and `ValidInterpretedModel(model)` is false, the result is `Unattributed` with `Condition: ...` and no answers. Otherwise set `res.Answers = fromWire(...)`, and set `res.Notice` when the change notice fires. Then, for every sent call, whatever its availability, build and record the usage before the result is returned: `usage := Usage{InputTokens: requestTokens(req), RequestModel: ModelID, Model: <validated id, or empty when the call is unattributed or not answered>}`, call `c.record(usage)` once, set `res.Usage = usage`, and only then return `res`. An unattributed result therefore carries `Result.Usage` with `RequestModel` set and `Model` empty, and its record is in the usage log.
- `(a Answer) Label()`: the prefix names `a.Model` (line 218 changes from `ModelID`). An empty `Model` renders `unattributed`.
- The package doc and the `@MX:ANCHOR` reason (lines 5–6), and the comments at lines 60–77, 214–218, 293, and 330–331, are rewritten. The MX reason changes from "the model pin" to "the request alias and the attribution". No comment in `internal/jev/jev.go` names `ModelID` except the doc comment of the constant.

### §D.3 Functions (internal/jevmeasure/measure.go)

- `Report`: add `RequestModel string`. The existing field `ModelID` (line 175) remains, and it holds the interpreted id. Its comment says so.
- `Run` (line 282): the `Report` literal (line 306) sets `RequestModel: jev.ModelID` instead of `ModelID: jev.ModelID`, so no attribution field receives the constant. Collect `result.Usage.Model` for each answered call (OK with answers) into one set across both arms. After the loop: more than one id returns `fmt.Errorf` naming the ids in sorted order (REQ-JRM-006). Exactly one id sets `rep.ModelID`. Zero leaves it empty, and `Validate` then refuses the artifact (REQ-JRM-007).
- `Validate` (line 220): `ModelID` must satisfy `jev.ValidInterpretedModel` (REQ-JEVO-014 as amended; REQ-JRM-007). An empty value and the alias are both refused.
- `Render` (line 247): `model: <ModelID> (interpreted by the endpoint)` and `request model: <RequestModel> (alias sent)` replace the `(pinned)` line at 250.

### §D.4 Rule wording — `.claude/rules/moai/development/model-policy.md` (file 14)

Insert the section below before `## Rules`. Add the bullet at the end of `## Rules`. Copy the whole file to file 15 and confirm `cmp` exit 0. The heading uses Title Case, which matches the other headings of this file.

```markdown
## Jev Model Attribution

[ZONE:Evolvable] [HARD] A Jev answer names the model that produced it. The request alias is what the package asks for, and it is never the name of the model that answered.

- **Request.** The `internal/jev` package sends the request alias `jev-latest`. The alias is a request, not an attribution.
- **Attribution.** The `model` field of the response is the interpreted model id, for example `jev-1.13.0`. Every model field a consumer reads carries the interpreted model id: the usage record's `model`, a measurement artifact's `model`, a rendered answer label, and a kickoff cross-check reason. The request alias appears only in its own field: `request_model` in a usage record, and `request model` in a measurement artifact.
- **Unattributed answers are not answers.** A response with no interpreted model id in the shape `jev-<major>.<minor>.<patch>` is an unavailable result (`unattributed`). No consumer reads an answer from it.
- **No straddled measurement.** A measurement run whose answered calls report more than one interpreted model id is refused. Re-run it under one id.
- **Version change.** When the interpreted model id of an answered call differs from the verified id held in code (`internal/jev/jev.go:VerifiedInterpretedModel`), the first such call in each client reports one notice line naming both ids and that location. Until the measurement regression set has been re-run against the new id and the verified id has been changed to it, every consumer that shows a Jev answer to a person, or selects on one (an advise-or-above consumer), returns to shadow: its answer is recorded, and it is neither shown nor selected on.
- **Regression set.** The regression set is the labelled answer set that `internal/jevmeasure` runs for each consumer.
- **Terms.** Both terms describe behaviour and introduce no configuration key. *advise-or-above* names a consumer behaviour that shows a Jev answer to a person, or selects on one. *shadow* names the state in which a Jev answer is recorded and neither shown nor selected on.

Why: the prohibition on the alias rested on the risk that the answering model could not be attributed. Recording the response's `model` field removes that risk. The vendor can move the version behind the alias at any time, and the notice and the shadow return make such a move visible rather than silent.
```

Rules bullet: `- Jev answers name the interpreted model the endpoint returns; the request alias \`jev-latest\` is never an attribution (§ Jev Model Attribution)`

### §D.5 Surfaces

- **doctor_jev.go (file 8).** Verbose detail: `request model <jev.ModelID> (alias sent; the interpreted version is recorded per answered call and is not probed here); endpoint <jev.EndpointURL>; readiness only — no judgment request is sent by this check`. The constant is cited here as the request alias, which is one of the permitted sites of AC-JRM-005.
- **decide.go (file 11).** The reason is `"Jev cross-check signal (" + jres.Usage.Model + ")."`. `jres` is in scope at that call site (decide.go:170). `jevVerdict` (line 477) returns the fallback `jev_call_failed` (line 490) for every availability other than `Available`, `Disabled`, `NoCredential`, `Unauthorized`, and `Malformed`. An `Available` result carries a valid `Usage.Model` by construction, because `Ask` converts an unattributable response to `Unattributed` before any consumer sees it. The receipt already stores the marshalled `jev.Result` as `RawResponse`, so the interpreted id reaches the receipt without further change.
- **mcp_jev.go (no code change).** `out["usage"] = res.Usage` (line 119) marshals the whole `Usage`, so the usage object gains `request_model`. The change notice reaches MCP callers through `out["notice"]`: `handleJevAsk` writes `res.NoticeLine()` for every result, available results included (lines 114–116, outside the `res.OK()` block that starts at line 117). Each `jev_ask` call builds its own client (`newJevAskClient`, lines 41 and 100), so the notice appears on every changed-version call and not once per process. File 10 asserts the usage keys and the `notice` key (AC-JRM-010). The CLI consumers do not surface the notice on success (§H).

### §D.6 Tests (names are the run-stage contract)

Fixture rule (§D.7): every response id in a test fixture that asserts a positive attribution is `jev-9.8.7` or `jev-9.8.8`, never `jev-1.13.0` or `jev-latest`, so that a value copied from the constant cannot pass. Three exceptions: (a) the negative cases of AC-JRM-003 (a missing id, `jev-latest`, and `jev-1.13`), which assert that no attribution is made; (b) the verified-id control of AC-JRM-010 (`jev-1.13.0`, which must carry no notice); (c) the validator inputs of `TestInterpretedModel_MatchesVendorVersionedShape`, which test the shape check itself.

internal/jev/jev_test.go:
- Replace `TestModelID_IsPinnedNotAnAlias` (line 74) with `TestRequestModel_IsTheAliasNotAVersion`: asserts `ModelID == "jev-latest"`.
- Replace `TestModelID_MatchesVendorVersionedShape` (line 92) with `TestInterpretedModel_MatchesVendorVersionedShape`: accepts `jev-1.13.0` (the verified id) and `jev-9.8.7`; rejects `jev-1.13`, `jev-latest`, and the empty string.
- Replace `TestRequestBody_CarriesPinnedIDAndNoAlias` (line 107) with `TestRequestBody_CarriesTheAliasAndNoInterpretedVersion`. It parses the body: the `model` is `jev-latest` (a positive control: the alias is what was sent), and it is not a versioned identifier. The question text is not scanned, which matches the narrowed REQ-JRM-001.
- Replace `TestUsage_RecordsInputTokensAndPinnedModelPerCall` (line 387) with `TestUsage_RecordsInterpretedModelAndRequestAliasPerCall`.
- Update `TestUsage_RecordedEvenWhenTheCallDidNotSucceed` (line 413): the record exists, `Model` is empty, and `RequestModel` is `jev-latest`.
- New: `TestResult_CarriesInterpretedModelFromResponse` (AC-JRM-002; subtests for one call, and for two calls answered `jev-9.8.7` then `jev-9.8.8`, where each record carries its own call's id, in order); `TestResponse_WithoutInterpretedModelIsUnattributed` (table: missing, `jev-latest`, `jev-1.13`; AC-JRM-003; asserts `Result.Usage` and `UsageLog()`); `TestUsage_RequestAliasAndInterpretedModelAreSeparateFields` (AC-JRM-004); `TestAnswerLabel_NamesTheInterpretedModel` (AC-JRM-005; an empty `Model` renders `unattributed`); `TestChangedInterpretedModel_NoticeOncePerClient` (AC-JRM-010).
- Update `TestAnswerLabel_IsDistinguishableFromMeasurementAndJudgement` (line 536): the fixture carries `Model: "jev-9.8.7"`, and the label must contain it.
- Update `TestDecode_DocumentedResponseShape` (line 768) to assert the decoded model (`jev-9.8.7`).
- Adjust `TestRequestBody_IsTheDocumentedMapShape` (line 685) wording only. Its comparison already reads `ModelID`.
- The fixture bodies at lines 51 (`okBody`), 357, and 769 carry `"model":"jev-9.8.7"`.

internal/jevmeasure/measure_test.go:
- Fixture at line 55: the answerer's `Usage` carries `Model: "jev-9.8.7"` and `RequestModel: jev.ModelID`.
- `TestRun_ProducesBothArmsAndDelta` (line 140): assert `rep.ModelID == "jev-9.8.7"` and `rep.RequestModel == jev.ModelID` (the assertions at lines 152–153 change).
- Fixtures at lines 199 and 228: `ModelID: "jev-9.8.7"`.
- `TestReport_RenderCitesPinnedModelAndCarriesNoAlias` (line 379) is renamed `TestReport_RenderCitesInterpretedModelAndAliasOnlyInRequestField` (AC-JRM-007). Its positive control: a rendering whose model line is the alias fails the same check.
- New: `TestRun_RefusesAStraddledMeasurement` (AC-JRM-006; fixture ids `jev-9.8.7` and `jev-9.8.8`) and `TestRun_CitesTheInterpretedModelAcrossArms`.

internal/jevmeasure/edges_test.go and gate_demo_test.go: the fixtures (edges lines 61, 81, and 107; gate_demo lines 61–62) use `"jev-9.8.7"`. `TestValidate_RejectsMissingModelAndMismatchedThreshold` (line 72) gains a case: the alias in the model field is rejected.

internal/cli/doctor_jev_test.go: new `TestDoctorJev_VerboseNamesAliasAndNoInterpretedVersion` (AC-JRM-008).

internal/cli/mcp_jev_test.go: the fake server body (line 93) carries `"model":"jev-9.8.7"`. Assertions: `usage.model` is `jev-9.8.7`, and `usage.request_model` is `jev-latest`. New: `TestJevAsk_NoticeKeyFollowsTheChangeNotice` (AC-JRM-010, the MCP surface): a `jev_ask` call answered `jev-9.8.7` returns the `notice` key, and a call answered `jev-1.13.0` (the verified-id control) returns none. Each call builds its own client (`newJevAskClient`), so the key is present on every changed-version call.

internal/contract/kickoff/decide_test.go: new `TestDecide_JevCrossCheckReasonNamesInterpretedModel` (AC-JRM-009) and `TestDecide_UnattributedJevTakesTheFallback` (an unattributed result takes the `jev_call_failed` fallback on the R2 path; for each of the two LLM answers, `approve` and `reject`, the test asserts that the final outcome equals that answer, the rule is `R2`, the effective decider is `llm`, and the receipt's fallback reason is `jev_call_failed`). The helpers stay in this file, so no new file is added.

internal/cli/todo_auto_rank_test.go:
- `TestAutoLiveJevRanker` (line 1407): the body carries `"model":"jev-9.8.7"`. A new subtest, "gate on with a credential and no versioned model is unattributed", sends a body with no model and expects `jev.Unattributed` with no answers.
- `TestAutoFallbackReasonMapping` (line 621): new case `{jev.Unattributed, "jev-unattributed"}`.

### §D.7 Naming and wording rules for the run stage

- Use the plain identifiers in §D.2. Do not introduce `Pinned*`, `pinned`, or `resolved` in new identifiers or comments. The existing `pinned` comments in the files of §D.1 are rewritten at M6.
- Fixture response ids that assert a positive attribution are `jev-9.8.7` and `jev-9.8.8`, never `jev-1.13.0` or `jev-latest`; the three exceptions named in §D.6 apply. A fixture that carries the constant's value while asserting a positive attribution cannot tell a correct implementation from one that copies the constant into the attribution field.
- Run each `go test` as a plain single invocation: no pipes, no shell-variable constructs. The worktree guard refused a command that combined a pipe with `${PIPESTATUS}`; plain invocations with anchored selectors were accepted. Anchor every `-run` selector (`'^Name$'`, and `'^(A|B)$'` for a family), because the spec linter flags an unanchored selector as a vacuous assertion.
- No emoji. No time estimates. English only in the rule and the code comments.

---

## §E. Milestones (ordered by decision reversibility)

- **M1 — data model and attribution** (`internal/jev`). Write the §D.6 tests first, observe them red (the RED-now cells of acceptance.md), then implement §D.2 without the notice. Gates: the `internal/jev` package tests. Flips AC-JRM-001, 002, 003, and 004, and the label test of AC-JRM-005.
- **M2 — measurement attribution** (`internal/jevmeasure`). Straddle refusal, Validate, Render. Gates: the `internal/jevmeasure` package tests. Flips AC-JRM-006 and 007.
- **M3 — change notice** (`internal/jev`, and the MCP assertion in `internal/cli`). The `VerifiedInterpretedModel` constant with its dated comment, `Result.Notice`, `NoticeLine()`, and the per-client flag. The `notice` key assertion of AC-JRM-010 (file 10) lands here too: `handleJevAsk` already writes `NoticeLine()` into `notice`, so the handler needs no code change. Flips AC-JRM-010.
- **M4 — rule text and parity.** Edit file 14, copy to file 15, `cmp`, then `make build`. Gate: `TestRuleTemplateMirrorDrift/model-policy.md` passes. Flips AC-JRM-011.
- **M5 — surfaces.** Doctor (files 8–9), kickoff (files 11–12), MCP (file 10), and todo auto-rank (files 16–17). Gates: the §F selectors. Flips AC-JRM-008 and 009.
- **M6 — mechanical sweep.** Comments (files 7 and 13; `jev.go` comments; the `pinned` wording in the test comments of files 2, 4, and 6), citation sites, and fixtures. Gate: the AC-JRM-005 allow-list check.
- **M7 — verification and evidence.** The §F gates. Run-phase evidence (§E.2, §E.3) is written by manager-develop. Sync evidence (§E.4) is written by manager-docs, which also flips AC-JRM-016.

---

## §F. Verification gates (run stage; plain invocations)

```text
gofmt -l internal/jev internal/jevmeasure internal/cli/doctor_jev.go internal/cli/doctor_jev_test.go internal/cli/mcp_jev_test.go internal/cli/todo_auto_rank.go internal/cli/todo_auto_rank_test.go internal/contract/kickoff internal/config/types.go
go vet ./internal/jev/ ./internal/jevmeasure/ ./internal/contract/kickoff/ ./internal/cli/
go test -count=1 ./internal/jev/
go test -count=1 ./internal/jevmeasure/
go test -count=1 ./internal/contract/kickoff/
go test -count=1 ./internal/contract/receipt/
go test -count=1 ./internal/cli/ -run '^(TestJev.*|TestDoctorJev.*|TestCheckJev.*|TestAutoRank.*|TestAutoLive.*|TestAutoFallbackReasonMapping$|TestAutoDefaultSeamsAreInert$|TestAutoLaneCycle.*|TestTodoAuto.*|TestFactoryNextBareUnchanged$)$'
go test -count=1 ./internal/template/ -run '^TestRuleTemplateMirrorDrift$'
make build
GOOS=windows GOARCH=amd64 go build ./...
```

The `internal/cli` selector is the anchored union of the test families in `internal/cli` that touch `internal/jev`: the changed ones (doctor, MCP, and todo auto-rank), and the unchanged ones (`todo_jev_finding*`, `todo_auto_test.go`, `todo_auto_lane_test.go`, and `TestFactoryNextBareUnchanged`, which injects a Jev result). The kickoff and receipt packages have their own gates above. The selector is not the whole package, because the card-worktree CLI suite is structurally red (§H).

Lint and audit gates, with the tree-built binary (built from the tree in this run, and invoked by path):

```text
go build -o <scratchpad>/moai-tree ./cmd/moai
<scratchpad>/moai-tree spec lint .moai/specs/SPEC-JEV-RESOLVED-MODEL-001 .moai/specs/SPEC-JEV-CORE-001 .moai/specs/SPEC-JEV-OPTIN-MEASURE-001
<scratchpad>/moai-tree spec audit --filter-spec SPEC-JEV-CORE-001
<scratchpad>/moai-tree spec audit --filter-spec SPEC-JEV-OPTIN-MEASURE-001
```

The `spec audit` flag form is checked against `moai-tree spec audit --help` before use. The MCP audit is cited as corroboration only, with its build commit.

Clarification-marker check (plan-close and run-close): `command grep -rnie 'needs[[:space:]]clarification' .moai/specs/SPEC-JEV-RESOLVED-MODEL-001` prints nothing and exits 1.

---

## §G. Open questions — settled, and the rest non-blocking

### Settled (card:t1768, board, brain, 2026-10-15)

- **OQ-1 — the mechanism behind REQ-JRM-010 (the change notice).** Decided: option A. A compiled verified-id constant (`internal/jev/jev.go:VerifiedInterpretedModel`), one notice line per client, no file write, and REQ-JEVC-011 untouched. The conditions are met: the constant's comment carries the measurement date 2026-10-15, the request alias `jev-latest`, and the response id `jev-1.13.0`; the notice names the constant's location. It fixes REQ-JRM-010, and AC-JRM-010 is adopted. Reference: card:t1768 (board). Option B (a state record under `.moai/state/`) is not chosen, because it would need a REQ-JEVC-011 carve-out. Option C (in-process only) is not chosen, because it does not satisfy REQ-JRM-010 for CLI consumers.
- **OQ-2 — the shadow-return terms.** Decided: option (a). *advise-or-above* and *shadow* are defined by behaviour, with no configuration key (§D.4; REQ-JRM-011). Reference: card:t1768 (board).
- **OQ-3 — the regression-set reference.** Decided: option (b). The reference that REQ-JRM-010 compares against is the verified id. The REQ-JRM-011 re-run procedure compares against it: re-run the regression set against the new id, then change the verified id. The shipping-prohibition clause is dropped from §D.4 and from REQ-JRM-011. Reference: card:t1768 (board).

### Decided (card:t1768, board d-20261015T052851Z-42ed; brain, Jev, and leader agree)

- **D11 — the fallback reason for an unattributed answer.** Decided: the new reason token `jev-unattributed` (file 17). It is not merged into `jev-unreachable`, and the invariant of `TestAutoFallbackReasonMapping` (no two conditions share a fallback reason) stays; §D.6 adds one case row to that test. The reason is operator-visible. Reference: board d-20261015T052851Z-42ed.
- **Kickoff — the unattributed result.** Decided: keep the R2 fallback, stated as what it does. An unattributed result takes the same fallback path as every other Jev unavailability: `jevVerdict` returns a fallback reason, here `jev_call_failed` from its default branch (decide.go:489–490). Rule R2 applies, the effective decider is `llm`, the outcome equals the LLM's answer, and the receipt records the fallback reason `jev_call_failed`. The unattributed answer itself is never read. REQ-GR-010 in SPEC-AUTONOMY-GATE-REWIRE-001 defines R2 as substituting `llm` alone and deriving by R4; it stays as it is, and this path is not fail-closed. Reference: board d-20261015T075953Z-63b9, which corrects d-20261015T052851Z-42ed (that record carried the label "fail-closed").

### Non-blocking items (recorded, not gating)

- **OQ-4 — reachability.** `model-policy.md` is path-scoped to `**/.claude/agents/**`. The Jev section does not load for `internal/jev` work. Either keep the card's placement and accept the gap, or widen the `paths:` key to include `**/internal/jev/**` and `**/internal/jevmeasure/**`. The card names the file, so this plan keeps the placement.
- **OQ-5 — amendment transition precedent.** CORE 0.2.0–0.4.0 kept `status: completed` as a non-transition correction. This card follows its instruction (`completed → in-progress (amendment)`, §B D4). The precedent is not an option for this SPEC.
- **OQ-6 — sync-phase documentation.** `.moai/docs/jev-local-operations.md`, `kickoff-autonomy.md`, and `stale-card-premise-check.md` mention Jev. None states the pin or the alias. manager-docs reviews them at sync, and writes the CHANGELOG entry.
- **OQ-7 — dispatch.** Card t1762 depends on this card's attribution fields. Its plan must cite them and must be planned after this card lands.
- **D18 — pin wording in the other SPEC bodies.** The pin wording of SPEC-JEV-CONSUMERS-001 and SPEC-JEV-GOAL-DIST-001 is deferred to card t1803. Reference: board d-20261015T052851Z-42ed.

---

## §H. Residual risk and gaps

- **Live confirmation is a gap.** The alias-to-version behaviour (`jev-latest` → `jev-1.13.0`, same answer and tokens) is the leader's observation of 2026-10-15. This stage made no call to the endpoint, and no credential is present in this tree. The run-stage tests use fixtures and a local server. Live confirmation is a separate observation, recorded by the run stage with its own command and output, or it stays a gap.
- **Shape change.** If the vendor changes the `model` field's shape, every consumer becomes unavailable (`unattributed`) until the shape is accepted, so no unattributed answer is shown or selected on. On the Kickoff path the decision falls to the LLM alone under rule R2 (§G), not to a closed gate. That is the designed behaviour, and it is visible to the service.
- **Version changes reach the MCP output, not the CLI consumers.** The change notice is carried on `Result.Notice` and read through `NoticeLine()`. The MCP output writes it into `notice` for every result (mcp_jev.go:114–116), and each `jev_ask` call builds its own client, so the notice appears on every changed-version call; AC-JRM-010 asserts that key. The gap is handed to follow-up card t1803 (board d-20261015T052851Z-42ed): the CLI consumers, the `--auto` ranking and the Kickoff cross-check, do not surface the change notice on success, and this SPEC does not design that surfacing. No other surface changes in this SPEC (the file count stays at 17), and the Tier M basis stays as recorded in §A.
- **Consumer breakage outside the named families is possible.** The consumer sweep (§D.1) read every fixture that carries an `answers` body. Injected `jev.Result` fixtures whose empty `Usage.Model` changes a rendered reason remain possible in the importers that §F runs by name.
- **The CLI suite is not a gate.** The card-worktree CLI suite is structurally red, so §F runs anchored families only. A family-level pass does not prove the rest of `internal/cli`.

---

## §I. Cross-references

- `.moai/specs/SPEC-JEV-CORE-001/spec.md` — REQ-JEVC-003, REQ-JEVC-004 (amended in this commit); REQ-JEVC-011 (display-only, unchanged); REQ-JEVC-017 (gate, unchanged); REQ-JEVC-022 (doctor, unchanged).
- `.moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md` — REQ-JEVO-009, REQ-JEVO-014 (amended); §B (amended).
- `.moai/specs/SPEC-JEV-OPTIN-MEASURE-001/acceptance.md` — AC-JEVO-010, AC-JEVO-011 (amended; the straddle clause is owned here by REQ-JRM-006 and AC-JRM-006).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum (amendment exemption), § Status Transition Ownership Matrix.
- `.claude/rules/moai/workflow/spec-workflow-detail.md` § Depends_on Pre-flight Check (why D5).
- `.claude/rules/moai/development/verification-completeness.md` §2 (RED-now and green-path cells; the evidence ledger in acceptance.md).
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 (tool provenance; §C above).
- `.claude/rules/moai/development/model-policy.md` (file 14) and its template copy (file 15).
- `.claude/rules/moai/workflow/rule-loading-budget.md` — the file is path-scoped; its size is 29,950 bytes before this change (both copies). The run stage records the byte size after the change.
