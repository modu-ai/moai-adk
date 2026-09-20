# SPEC-JEV-INTEGRATION-001 — Implementation Plan

> Sections are ordered by decision-reversibility: the choices most likely to change on review come first (§B data-model, §C interfaces, §D user-facing flow), and the mechanical work is deferred to the end (§F milestones M8, §G). The milestone list in §F stays in execution order because its dependencies are real.

---

## §A. Context

Jev is a System One judgment model reached over one HTTP endpoint. It returns typed answers with probabilities and generates no text. This SPEC makes it an official, opt-in, default-off, display-only capability of moai-adk, with three consumers each gated on its own measurement.

Verified tree facts this plan rests on (measured in this worktree at `fd75cf692`):

| Fact | Command | Observed |
|---|---|---|
| No Go reference to the model today | `grep -ril "typesafe\|jev" internal/ pkg/ cmd/` | 0 files (positive control: `grep -ril glmcred` → 5 files) |
| `scripts/jev/` absent on develop | `ls scripts/` | no `jev` entry among 16 |
| SPEC ID free | `ls -d .moai/specs/SPEC-JEV-INTEGRATION-001` | absent (control: `SPEC-AGENT-001` resolves) |
| Only two finding sources exist | `internal/kanban/backlog_store.go:117-120` | `mechanical`, `agent` |
| `Source` has no validator | grep over write paths | free string; `todo relate` hardcodes `agent`, `todo add` hardcodes `mechanical` |
| No init route in the console | `internal/web/app.go:154-201` | 11 routes, none for init |
| MCP tool counts | `grep -n "30 tools\|26 of the 30"` in both copies | both say 30 / 26 of 30, byte-identical mirrors |

---

## §B. Data-model decisions (highest change likelihood — review these first)

**B1 — A third finding source constant.** `BacklogFinding.Source` is a free string, but only two constants exist and they are **not interchangeable**. `HasAgentFindingForPair` selects on `Source == BacklogSourceAgent` and is, by its own doc comment, the predicate behind the `machine-only` mark — the mark records the *absence* of an agent-sourced record for a pair, never that any review took place. Writing a Jev finding as `agent` would therefore make the queue claim a review happened for that pair when none did: a false review record, and precisely the display-becomes-verdict failure this card exists to prevent.

So a Jev finding carries a third constant, `BacklogSourceJev = "jev"`. Two consequences follow and both need a decision on review:

- **Dedup precedence.** `AppendFindingOnce` uses `HasFindingTuple`; `SamePairAs` compares unordered pairs. Proposed rule: a Jev finding is appended only when no finding of **any** source already names that unordered pair with the same relation, and a later mechanical or agent finding for the same pair is appended normally alongside it. Rationale: a Jev signal must never suppress a measurement or a human judgement, and must never be suppressed into invisibility by one either — both are records, and the render distinguishes them.
- **Render.** `todoFindingLine` prints a score only for `mechanical` (an agent judgement carries no measurement, and rendering `0.00` would read as a measured dissimilarity). A Jev finding carries a probability, which is neither. Proposed: render it in a distinct form — a labelled probability, not a bare `score N.NN` — so a reader cannot read a model's confidence as the text analyser's similarity.

**B2 — Config key shape.** `workflow.jev.enabled`, inside the existing `workflow.yaml`, following `codex.review_gate.enabled` / `multi.review_gate.enabled` / `slot_lease.enabled`: shipped `false`, compiled default in `internal/config/defaults.go`, template block documenting it. An alternative shape (`workflow.jev: {enabled, model, …}`) is available if the pinned model id should be operator-visible rather than compiled — that is a review decision, not a settled one.

**B3 — Credential is out of the settings schema.** Modelled on `internal/glmcred` and `internal/web/glmkey.go`: a dedicated package, mode 0600 with the existing wider-mode tightening, a hand-built parse/validate/view path, and absence from `settings.AllFields()` enforced by a regression test in the shape of `TestGLMKeyField_AbsentFromSchema`. The four-character disclosure floor is inherited verbatim: a key of four characters or fewer discloses nothing but `Configured`.

---

## §C. Interface decisions

**C1 — Package boundary.** `internal/jev` depends only on the standard library and stdlib-only internal leaf packages (`internal/paths`, `internal/defs`), exactly as `internal/glmcred` does, so it can be imported by both `internal/cli` and `internal/web` without creating a cycle. The credential reader may live inside `internal/jev` or as a sibling `internal/jevcred`; the second is closer to the existing precedent and is the proposal.

**C2 — The unavailable result is a value, not an error.** REQ-JEV-005/006/007 require that no absence reaches a caller's exit status. The proposed surface returns `(Answer, Availability)` where `Availability` names the observed condition (disabled / no-credential / unauthorized / rate-limited / overloaded / unreachable / oversize). A caller that ignores `Availability` degrades to "no answer", never to a failure.

**C3 — Batching is in the type.** REQ-JEV-010 is enforced by shape: a request carries one state and a *list* of questions. A single-question convenience wrapper is fine; a path that sends two requests over one state is not.

**C4 — Size refusal precedes transport.** The 32k state-plus-longest-question and 64k whole-request bounds are checked before the HTTP call and refuse rather than truncate, so an oversize state is a visible condition rather than a silently shortened one.

---

## §D. User-facing flow decisions

**D1 — One wizard question.** A `Question{ID: "jev_enabled", Type: Select, Group: …}` in `internal/cli/wizard/questions.go`, with `QuestionTranslation` entries in all four locales in `translations.go`. Which constructor it joins is a review decision: `Page3Questions` (init-only, alongside `agent_wiring` / `autonomy_tier`) keeps the reconfigure set unchanged; `DefaultQuestions` would also reach `moai update --reconfigure`. The proposal is `Page3Questions` plus an explicit reconfigure entry, so the setting is changeable later without adding a fifth init question — but note `InitQuestions` currently asks four, and the card's goal is one question, so this needs a decision rather than an assumption.

**D2 — One web section, no new route.** The Jev section renders inside the existing `/settings` screen and persists through the existing `/save` handler. The console has no init route and gains none. A key-reveal route is *not* proposed: the GLM precedent has one (`glmKeyRevealPath`), but reveal is a separable decision and the disclosure requirement (REQ-JEV-020) is satisfied without it.

**D3 — One persistence path.** Both entrances call the neutral `internal/settings` seam, per the AP-2 rule recorded at `internal/web/projectconfig.go:197` and `:260`: the load-modify-write body lives in `internal/settings` so the console and the TUI wizard drive one writer. The nested-isolation crux applies — `SetSection` replaces the whole section struct, so the seam copies the entire struct and mutates only the targeted field, and every sibling rides through byte-identical.

**D4 — Privacy disclosure at the point of choice.** Both the wizard question's `Description` and the web section's body state that enabling sends card text or request text to a third-party server. This is a string in four locales, not a link.

---

## §E. Constraints

- **Display-only** is the load-bearing invariant. Every consumer is verified against a queue-file hash taken before and after (the SHA-256 method already in `internal/cli/todo_triage_test.go:649`).
- **Default off.** Template default `false`, compiled default in `defaults.go`.
- **Fail-open.** No absence becomes a failure; retry only where repetition is provably side-effect-free.
- **Measurement is binding.** A consumer that does not beat its own constant baseline is not shipped — and the SPEC says so, so its absence reads as a decision.
- **Thresholds are fitted, never copied.** The `0.6` / `0.85` figures in vendor documentation are illustrative. A Noul threshold and a Choice threshold are different quantities and are never transferred.
- **Pinned model id everywhere.** `jev-latest` appears in no request and in no measurement citation.
- **Template-First.** Every `.claude/` or `.moai/` file lands in `internal/template/templates/` first; `make agents-emit` if an agent definition changed, `make commands-emit` if a command source changed; template content carries no SPEC ID, card id, date, price, or measurement figure.
- **Verification scope.** Affected packages only (`go test ./internal/jev/... ./internal/kanban/... ./internal/cli/... ./internal/web/... ./internal/config/...`); the full-suite verdict comes from CI. No full local `go test ./...`.

---

## §F. Milestones (execution order; priority High unless noted)

**M1 — Foundation.** `internal/jev` package: request/response types, pinned model id, size bounds, secret screening, HTTP transport, usage accounting. Credential package at `~/.moai/.env.typesafe` mode 0600. `workflow.jev.enabled` in `defaults.go`, local `workflow.yaml`, and the template block. `moai doctor` Jev check (a `DiagnosticCheck{Name: "Jev"}` following `doctor_codex.go` / `doctor_disk.go`) that probes reachability without sending a judgment request. Nothing consumes the package yet.

**M2 — Opt-in surfaces.** Wizard question + four-locale translations; web settings section with toggle and credential field; the shared `internal/settings` persistence seam; the anti-leak regression test asserting absence from `settings.AllFields()`; the four-character disclosure floor.

**M3 — Measurement harness.** A labelled-answer-set runner with two language arms, a constant-answer baseline per question shape, per-consumer threshold fitting, pinned-model-id citation, and a positive control for every absence-shaped result. No consumer ships before its measurement exists.

**M4 — Consumer C (near-duplicate marking).** `BacklogSourceJev` constant; the write path modelled on `todo_analysis.go:58-76`; the dedup precedence rule from §B1; the render form from §B1; the `HasAgentFindingForPair`-stays-false assertion; the queue-hash assertion. Declaration-only updates in the tests that enumerate the source set: `internal/kanban/backlog_findings_test.go`, `backlog_archive_test.go`, `todo_merge_procedure_test.go`, `todo_queue_merge_test.go`, `internal/cli/todo_analysis_test.go`, `todo_analysis_add_test.go`, `todo_relate_test.go`. Gated on M3's C-measurement.

**M5 — Consumer A (lane-question routing).** One batched request: a Choice naming the decision owner (lead / operator / worker, plus a no-match option) and two Nouls (needs-a-measurement, cheap-to-reverse). Operator-owned or below-threshold routes to operator. Gated on M3's A-measurement. Priority Medium relative to M4 — M7 seat (i) depends on it.

**M6 — Consumer B (skill suggestion).** Two requests: a wide rank over all skills batched with a needs-a-skill-at-all Noul, then a top-3 rerank under fuller text. Presented to the orchestrator as a ranked signal; the `/moai` intent router keeps its selection authority. Gated on M3's B-measurement. Priority Medium.

**M7 — `/moai goal --auto` seats.** Seat (i): inside the sealed-scope loop, a lane's blocking question is classified through M5's routing; lead-owned and cheap-to-reverse items are answered by the lead, operator-owned items remain a persisted blocked result exactly as today, and no new user question is introduced. Seat (ii): a Jev Noul answer is recorded as a separate auxiliary item in the governance receipt under `.moai/state/mission/governance/`, with the governor's authority, read-only scope, output shape, and existing receipt bindings unchanged, and with no Jev answer entering a completion predicate. Depends on M5.

**M8 — MCP wrapper, reference skill, docs.** A thin `mcp__moai__jev_*` tool over `internal/jev`; both tool-count figures updated in **both** copies of `moai-mcp-tools.md` (`.claude/rules/moai/core/` and its template mirror, currently saying "30 tools" at line 3 and "26 of the 30" at line 63, byte-identical today); one reference skill carrying question-design rules only; the recorded disposition of `scripts/jev/`; the preserved negative result. Priority Low — mechanical, and last because the tool counts must match the shipped set, which M4-M6 decide.

---

## §G. Anti-patterns

- **Writing a Jev finding as `agent`.** Produces a false review record (§B1). Mechanically caught by the M4 assertion that `HasAgentFindingForPair` stays false.
- **Copying a vendor threshold.** `0.6` and `0.85` are documentation examples. A threshold not fitted on this repository's data is an unmeasured claim.
- **Transferring a threshold between Noul and Choice.** Different question shapes, different calibration.
- **Citing `jev-latest`.** A measurement under a moving alias is unattributable.
- **Reading a zero-hit as an absence.** Every absence-shaped result needs a positive control, or it is a gap rather than a finding.
- **Treating a shipped-but-unmeasured consumer as gated.** The gate is the committed measurement, not the intention to take one.
- **Letting the display become the verdict.** The whole capability is one signal among others; a code path that acts on it is the defect.

---

## §H. Cross-references

- `.claude/rules/moai/core/verification-claim-integrity.md` — why a measurement's baseline attribution and positive control are obligations rather than courtesies.
- `.claude/rules/moai/core/moai-mcp-tools.md` (+ template mirror) — the tool catalogue M8 updates.
- `.claude/skills/moai/workflows/goal.md` § `/moai goal --auto` — the sealed-scope loop M7 seat (i) enters.
- `.claude/agents/moai/mission-governor.md` — the read-only decision agent M7 seat (ii) supplies an auxiliary input to.
- `internal/glmcred/glmcred.go`, `internal/web/glmkey.go` — the credential and disclosure precedent.
- `internal/web/projectconfig.go` §AP-2 — the one-persistence-path rule D3 follows.
- `internal/kanban/backlog_store.go:108-150` — the finding contract §B1 extends.
