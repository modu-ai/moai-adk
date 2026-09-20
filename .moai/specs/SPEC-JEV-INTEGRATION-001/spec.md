---
id: SPEC-JEV-INTEGRATION-001
title: "Jev (TypeSafe System One) — opt-in, display-only judgment capability"
version: "0.1.0"
status: draft
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: "internal/jev + internal/cli + internal/web + internal/template/templates"
lifecycle: spec-anchored
tags: "jev, typesafe, system-one, opt-in, display-only, measurement-gated"
tier: L
---

# SPEC-JEV-INTEGRATION-001

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-09-20 | 0.1.0 | Initial plan-phase authoring from backlog card t1020. Operator pre-decided: official integration, three consumers (C/A/B), both `--auto` seats, two opt-in surfaces with one persistence path, English-card measurement deferred to a later card, Go-binary core with a thin MCP wrapper. | manager-spec |

---

## §A. Problem Statement

TypeSafe **Jev** is a System One judgment model: it answers a typed question about a supplied state and returns a typed answer (Choice / Noul / Score) with a probability. It generates no text. Today it reaches this repository only as an uncommitted local script set in the primary checkout's working tree — `scripts/jev/` does not exist on `develop` (verified: `ls scripts/` shows no `jev` entry), and no Go file under `internal/`, `pkg/`, or `cmd/` mentions `typesafe` or `jev` (verified: 0 matches, against a `glmcred` positive control returning 5 files).

That leaves three problems.

**It is unreachable.** A capability living only in one machine's uncommitted working copy cannot be used by anyone else, cannot be tested, and cannot be cited — an uncommitted working copy is not canonical and may not be quoted as one. Every operator other than the one holding that working tree has no Jev at all.

**Its known-bad use has no written boundary.** Jev was measured for one job — judging whether a stale backlog card's premise is dead — and that use was **rejected on measurement**: dead-call precision 14/48 = 29.2% against a 25% base rate, 2-class accuracy 58.9% against 75.0% for a constant "always alive" answer, and an English control on 40 cards reaching 67.5%, still below the constant. The rejection currently lives in prose and in a gitignored local evidence file. Nothing in the product records it, so nothing stops the next person re-litigating it.

**Where it would genuinely help, nothing carries it.** Three decisions in MoAI are currently made with no signal at all: whether a newly admitted card is a near-duplicate of an existing one in a way the text analyser's score does not capture; who owns a lane's blocking question; and which skill, if any, a turn needs before the intent router picks. Each is a typed judgment over a small state — the shape Jev answers well — and each is a place where a *displayed* signal helps and a *decided* answer would be a defect.

The gap this SPEC closes is therefore not "add a model". It is: make one capability reachable, bound it to display, gate each consumer on its own measurement, and write the rejected use down so it stays rejected.

---

## §B. Goal

Promote Jev into an official, **opt-in**, **display-only**, **default-off** capability of moai-adk:

- one Go package (`internal/jev`) owning the single implementation — HTTP client, credential read, config gate, fail-open, pinned model id, usage accounting;
- one shared persistence path reached from two entrances (the `moai init` wizard question and the `moai web` settings section);
- three consumers, each shipped only after its own measurement beats its own constant-answer baseline on this repository's data;
- two `/moai goal --auto` seats, neither of which moves the governor's authority;
- a thin moai-mcp tool wrapper over the same implementation, plus one reference skill carrying question-design rules.

---

## §C. Requirements (GEARS)

### C.1 — Core capability (`internal/jev`)

**REQ-JEV-001** (Ubiquitous) The `internal/jev` package shall be the single implementation of the TypeSafe System One call path: request construction, credential read, HTTP transport to `POST https://api.typesafe.ai/v1/systemone`, response decoding, and usage accounting.

**REQ-JEV-002** (Ubiquitous) Every consumer — CLI, MCP tool wrapper, and web surface — shall reach the model only through `internal/jev`; no second client implementation shall exist.

**REQ-JEV-003** (Ubiquitous) The package shall send a **pinned** model id. The alias `jev-latest` shall not appear in any request this package constructs.

**REQ-JEV-004** (Ubiquitous) The package shall record, per call, the request's input-token count and the pinned model id, so a later cost or provenance question is answerable from the record rather than from recollection.

**REQ-JEV-005** (Capability gate) **Where** the credential file is absent, unreadable, or empty, **the `internal/jev` package shall** return a typed unavailable result rather than an error that propagates to a caller's exit status.

**REQ-JEV-006** (Event-detected) **When** the transport returns HTTP 401, 429, or 529, or the network is unreachable, **the `internal/jev` package shall** return a typed unavailable result carrying the observed condition.

**REQ-JEV-007** (Ubiquitous) Every consumer of an unavailable result shall emit at most one notice line, exit 0, and continue the workflow it was performing. Jev absence is graceful degradation, never failure.

**REQ-JEV-008** (Capability gate, unwanted) **Where** a call failed non-idempotently or ambiguously, **the `internal/jev` package shall not** retry it. Retry shall be limited to calls whose repetition is provably free of side effects.

**REQ-JEV-009** (Ubiquitous) The package shall keep a request's `state` payload and its longest question within 32k tokens, and the whole request within 64k tokens, refusing to send a request that exceeds either bound rather than truncating it silently.

**REQ-JEV-010** (Ubiquitous) Where several independent questions are asked over one state, the package shall batch them into a single request.

### C.2 — Display-only invariant

**REQ-JEV-011** (Ubiquitous) A Jev answer shall not mutate the backlog queue, a card's state, a card's text, a file, a branch, or a merge. The capability is display-only.

**REQ-JEV-012** (Unwanted) The system shall not consult Jev — not as a decision, and **not as an input** — for any of: a completion verdict, a merge approval, a `moai todo` or `moai gtd` mutation, an operator gate, a user-surface behaviour change, or a CodeRabbit slot-wait adjudication.

**REQ-JEV-013** (Ubiquitous) A Jev answer surfaced anywhere shall be labelled as a model-produced signal, distinguishable by a reader from a mechanical measurement and from a human or agent judgement.

**REQ-JEV-014** (Capability gate) **Where** Jev is disabled, uncredentialed, or unreachable, **every consumer shall** produce output identical to its pre-Jev behaviour apart from at most one notice line.

### C.3 — Opt-in, configuration, and credential

**REQ-JEV-015** (Ubiquitous) The capability shall be gated on `workflow.jev.enabled` inside the existing `.moai/config/sections/workflow.yaml`. No new config section file shall be created.

**REQ-JEV-016** (Ubiquitous) The shipped template default for `workflow.jev.enabled` shall be `false`, with `internal/config/defaults.go` as the compiled source of truth, following the shape of the existing opt-in switches in that file (`codex.review_gate.enabled`, `multi.review_gate.enabled`, `slot_lease.enabled`).

**REQ-JEV-017** (State-driven) **While** `workflow.jev.enabled` is false, **the system shall** make no network call to the TypeSafe endpoint and construct no request.

**REQ-JEV-018** (Ubiquitous) The API credential shall live at `~/.moai/.env.typesafe` at file mode 0600, outside the repository, read and written through one package modelled on `internal/glmcred` — including that package's tightening of a pre-existing wider mode on write.

**REQ-JEV-019** (Unwanted) The credential shall not enter `settings.AllFields()`, so no generic schema-walking loop can read, render, or write it.

**REQ-JEV-020** (Ubiquitous) A web view of the credential shall disclose only a `configured` boolean and, for a credential longer than four characters, its final four characters — never the credential itself, and never any part of a credential of four characters or fewer.

**REQ-JEV-021** (Ubiquitous) The `moai init` wizard shall carry exactly one Jev question, and the `moai web` settings screen shall carry one Jev section holding an enable toggle and a credential field.

**REQ-JEV-022** (Ubiquitous) Both entrances shall write through **one** shared persistence path in the neutral `internal/settings` package; no parallel writer shall exist.

**REQ-JEV-023** (Ubiquitous) Both entrances shall state, at the point of choice, that enabling the capability sends card text or user-request text to a third-party server.

**REQ-JEV-024** (Capability gate) **Where** a request payload would carry text, **the `internal/jev` package shall** apply a secret-screening step before sending, and refuse to send a payload in which a credential-shaped token is detected.

**REQ-JEV-025** (Ubiquitous) The wizard question and the web section shall carry strings in all four locales (ko / en / ja / zh).

**REQ-JEV-026** (Ubiquitous) `moai doctor` shall carry one Jev check reporting, without sending a judgment request, whether the capability is enabled, whether a credential is present, and whether the endpoint is reachable.

### C.4 — Measurement gate

**REQ-JEV-027** (Ubiquitous) Each consumer shall carry its own measurement: a labelled answer set drawn from this repository's data, scored against a **constant-answer baseline** for that consumer's question shape.

**REQ-JEV-028** (Unwanted) A consumer whose measured accuracy does not beat its own constant-answer baseline shall not be shipped. The measurement's verdict is binding, not advisory.

**REQ-JEV-029** (Ubiquitous) Every measurement shall run two arms — the card or request text in its Korean original, and the same text in English translation — and shall record the per-arm result and the delta.

**REQ-JEV-030** (Unwanted) This SPEC shall not change the card schema, the card issuance path, or the language any card is written in. The English arm is a measurement, not a migration.

**REQ-JEV-031** (Ubiquitous) A confidence threshold shall be fitted per consumer on this repository's measured data.

**REQ-JEV-032** (Unwanted) A threshold shall not be copied from vendor documentation — the figures `0.6` and `0.85` appearing there are illustrative — and shall not be transferred between a Noul question and a Choice question.

**REQ-JEV-033** (Ubiquitous) Every measurement result shall cite the pinned model id it was taken under.

**REQ-JEV-034** (Ubiquitous) A measurement reporting an absence — a zero-hit, a no-difference, a null result — shall be accompanied by a positive control demonstrating the measuring apparatus fires on that path.

### C.5 — Question design under known model weaknesses

**REQ-JEV-035** (Ubiquitous) All counting, arithmetic, numeric proximity, date ordering, and SHA comparison shall be performed in Go, and the results passed to the model as named JSON fields. The model shall not be asked to compute them.

**REQ-JEV-036** (Ubiquitous) Every Choice question shall carry an explicit no-match option.

**REQ-JEV-037** (Ubiquitous) A request's state shall carry only the fields its questions read, and shall favour identifiers and English-language fields over long prose.

**REQ-JEV-038** (Unwanted) The system shall not treat text inside a state payload as instruction. State is untrusted data.

**REQ-JEV-039** (Unwanted) Consumer logic shall not assume `P(yes)` equals `1 − P(no)`; where both are needed, both shall be read.

**REQ-JEV-040** (Ubiquitous) Questions shall be written to survive literal reading: negation, scoping words, and implied conditions shall be made explicit rather than left to inference, and multi-hop indirection shall be resolved in Go before the question is asked.

### C.6 — Consumer C: near-duplicate card marking

**REQ-JEV-041** (Event-driven) **When** a card is admitted to the backlog and the capability is enabled, **the system shall** record its Jev near-duplicate judgment as a `BacklogFinding`, changing no card.

**REQ-JEV-042** (Ubiquitous) A Jev finding shall carry a **third** source constant, distinct from both `BacklogSourceMechanical` and `BacklogSourceAgent`.

**REQ-JEV-043** (Unwanted) A Jev finding shall not be written with `Source: agent`. `HasAgentFindingForPair` is the predicate behind the `machine-only` mark, and that mark records the absence of an agent-sourced record for a pair — never that a review took place. Writing a Jev finding as `agent` would make the queue claim a review happened when none did.

**REQ-JEV-044** (Ubiquitous) `HasAgentFindingForPair` shall continue to return false for a pair whose only finding is Jev-sourced.

**REQ-JEV-045** (Ubiquitous) The rendering of a Jev finding shall follow the existing source-conditioned rules in `todoFindingLine`: a score is printed only where a measurement was taken, so a Jev finding's probability shall be rendered in a form a reader cannot mistake for the text analyser's similarity score.

**REQ-JEV-046** (Ubiquitous) The dedup precedence between a Jev finding and a mechanical or agent finding naming the same unordered pair shall be stated explicitly and enforced through the existing `AppendFindingOnce` / `SamePairAs` paths.

**REQ-JEV-047** (Unwanted) No code path shall write a card field as a consequence of a Jev finding.

### C.7 — Consumer A: lane-question routing

**REQ-JEV-048** (Event-driven) **When** a lane's blocking question is routed and the capability is enabled, **the system shall** ask one Choice question naming the decision owner (lead / operator / worker) and two Noul questions — whether a new measurement is needed first, and whether the decision is cheap to reverse.

**REQ-JEV-049** (Ubiquitous) The three questions shall be batched into one request over one state.

**REQ-JEV-050** (Ubiquitous) The routing answer shall be surfaced to the reader as a signal and shall not itself dispatch, answer, or close anything.

**REQ-JEV-051** (Capability gate) **Where** the Choice answer is `operator`, or the Choice confidence falls below this consumer's fitted threshold, **the system shall** treat the item as operator-owned.

### C.8 — Consumer B: skill suggestion

**REQ-JEV-052** (Event-driven) **When** a turn's intent is being routed and the capability is enabled, **the system shall** issue two requests: a wide rank over all skills paired with a Noul asking whether the turn needs a skill at all, then a rerank of the top three under fuller text.

**REQ-JEV-053** (Ubiquitous) The suggestion shall be presented to the orchestrator as a ranked signal; the `/moai` intent router shall retain its existing selection authority unchanged.

**REQ-JEV-054** (Capability gate) **Where** the "needs a skill at all" Noul answers negatively above this consumer's fitted threshold, **the system shall** suppress the ranked list rather than present a best-of-nothing.

### C.9 — `/moai goal --auto` seats

**REQ-JEV-055** (Event-driven) **When** a lane blocks with a question inside the sealed-scope loop and the capability is enabled, **the lead shall** classify the decision owner through Consumer A's routing, answer items that are both lead-owned and cheap-to-reverse, and leave operator-owned items as a persisted blocked result exactly as today.

**REQ-JEV-056** (Unwanted) Seat (i) shall not introduce a user question inside the sealed-scope loop. The single approval before the loop remains the only one.

**REQ-JEV-057** (Ubiquitous) A Jev Noul answer supplied to `mission-governor` shall be an **auxiliary input signal**, recorded as a separate item in the governance receipt.

**REQ-JEV-058** (Unwanted) Seat (ii) shall not alter the governor's authority, its read-only scope, its output shape, or the receipt's existing binding fields.

**REQ-JEV-059** (Unwanted) A Jev answer shall not be an element of a completion predicate, nor of the landed-ancestry or authoritative-readback evidence a completion receipt binds.

### C.10 — Surfaces, template, and distribution

**REQ-JEV-060** (Ubiquitous) The moai-mcp tool wrapper shall be a thin shell over `internal/jev`, carrying no second implementation of the call path.

**REQ-JEV-061** (Ubiquitous) `.claude/rules/moai/core/moai-mcp-tools.md` and its template mirror shall both be updated, including both tool-count figures each carries.

**REQ-JEV-062** (Ubiquitous) One reference skill shall carry question-design rules only, and shall not carry the call path.

**REQ-JEV-063** (Ubiquitous) Every new file under `.claude/` or `.moai/` shall land in `internal/template/templates/` first.

**REQ-JEV-064** (Unwanted) Template content shall not carry a SPEC ID, a card id, a date, a price, a measurement figure, or a reference to a local-only file.

**REQ-JEV-065** (Capability gate) **Where** a file under `internal/template/templates/.claude/agents/moai/` changes, **the build shall** require `make agents-emit`; **where** a command source changes, **the build shall** require `make commands-emit`.

**REQ-JEV-066** (Ubiquitous) The SPEC shall state the disposition of the uncommitted `scripts/jev/` working copy, with the Go package as the canonical implementation.

### C.11 — Preserved negative result

**REQ-JEV-067** (Ubiquitous) The system shall record that using Jev to judge whether a stale card's premise is dead was measured and rejected, with the measured figures preserved.

**REQ-JEV-068** (Unwanted) `moai todo triage` shall remain model-free. No requirement in this SPEC shall route a model answer into it.

---

## §D. Exclusions

The items below are deliberately NOT built by this SPEC. Each is named so a later reader can tell a decision from an omission.

### Out of Scope — premise-death judgment

- Using Jev to judge whether a stale backlog card's premise is dead. Measured and rejected: dead-call precision 14/48 = 29.2% against a 25% base rate; 2-class accuracy 58.9% against 75.0% for the constant "always alive" answer; an English control on 40 cards reached 67.5%, still below that constant. Origin of the figures: `.moai/reports/t943/verdict.md` — gitignored local evidence in the primary checkout, possibly absent from any given tree. This SPEC cites it as the origin and re-measures nothing.
- Re-running that experiment under a different gate, a different prompt, or a different language arm. The gate sweep was already flat across 0.30–0.80 and the English arm already measured; a new run needs a new reason, not a new attempt.
- Any change to `moai todo triage`, which stays model-free.

### Out of Scope — card schema and issuance

- Changing `BacklogItem`'s fields, JSON tags, or the frozen per-item contract.
- Changing how cards are issued, who may issue them, or the queue's producer rules.
- Switching card text to English. The English measurement arm records a delta so a later card can decide; this SPEC decides nothing about it.

### Out of Scope — authority and gates

- Moving any authority to Jev. Completion verdicts, merge approval, queue mutations, operator gates, user-surface behaviour changes, and CodeRabbit slot-wait adjudication are unreachable from this capability, by design and by acceptance criterion.
- Changing `mission-governor`'s scope, output shape, or receipt bindings.
- Adding a user question inside the `--auto` sealed-scope loop.

### Out of Scope — web console

- Any `moai init` flow inside `moai web`. The console has no init route (its route set is overview / kanban / monitor / todo / settings / events / save / specs / profile CRUD / GLM key reveal / shutdown), and designing one is a different piece of work.
- A Jev-driven write path anywhere in the console. The Jev section reads and persists configuration; it displays no judgment that mutates state.

### Out of Scope — other model uses

- Text generation of any kind. Jev produces typed answers and probabilities; nothing here asks it for prose.
- Replacing, ranking, or second-guessing any existing audit backend (claude / codex / glm). The cross-model audit path is untouched.
- Distributing `scripts/jev/` to user projects. The scripts are superseded by the Go package.

---

## §E. Success Criteria

1. With `workflow.jev.enabled: false` (the shipped default), every affected command's output is byte-identical to its pre-SPEC output.
2. With the capability enabled and no credential present, every affected command exits 0 with at most one notice line.
3. The backlog queue file's SHA-256 is unchanged across a Jev-consulting run that records a finding through the read path — asserted by the method already used in `internal/cli/todo_triage_test.go`.
4. `HasAgentFindingForPair` returns false for a pair whose only finding is Jev-sourced.
5. Each of the three consumers has a committed measurement citing the pinned model id, both language arms, and a constant-answer baseline; any consumer not beating its baseline is absent from the shipped build.
6. The credential is absent from `settings.AllFields()`, asserted by a regression test in the shape of the existing GLM anti-leak guard.
7. Both tool-count figures in both copies of `moai-mcp-tools.md` agree with the shipped tool set.
