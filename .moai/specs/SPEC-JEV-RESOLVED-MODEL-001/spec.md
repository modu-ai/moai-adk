---
id: SPEC-JEV-RESOLVED-MODEL-001
title: "Jev request alias jev-latest with the interpreted model recorded per answer"
version: "0.1.2"
status: completed
created: 2026-10-15
updated: 2026-10-15
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/jev + internal/jevmeasure + internal/cli + internal/contract/kickoff"
lifecycle: spec-anchored
tags: "jev, typesafe, system-one, model-attribution, request-alias, interpreted-model, measurement"
tier: M
related_specs: [SPEC-JEV-CORE-001, SPEC-JEV-OPTIN-MEASURE-001]
---

# SPEC-JEV-RESOLVED-MODEL-001

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-10-15 | 0.1.0 | Plan-phase artifacts for card t1768 (Class C, Template-First, P1). Tier M. Amends two completed SPECs in place, in the same commit: `SPEC-JEV-CORE-001` to 0.5.0 (REQ-JEVC-003, REQ-JEVC-004, Q2 wording) and `SPEC-JEV-OPTIN-MEASURE-001` to 0.2.0 (§B goal phrase, REQ-JEVO-014, AC-JEVO-010, AC-JEVO-011). Both move `completed → in-progress (amendment)`. Neither SPEC's `sync_commit_sha` changes. | manager-spec |
| 2026-10-15 | 0.1.1 | Plan repair round 1 after plan-audit iteration 1 (FAIL, overall 0.65, audited commit `cfb5517962088b09936dfbb8a2ff5b80e63a8893`). The three open decisions are settled (card t1768, board, brain): OQ-1 option A (a verified-id constant, one notice per client, no file write); OQ-2 behavioural terms, no configuration key; OQ-3 the verified id is the reference, and the shipping-prohibition clause is dropped. REQ-JRM-001, 005, 010, 011, and 013 reworded; REQ-JRM-014 added; AC-JRM-005, 010, 011, and 014 reworked; AC-JRM-016 added; fixture ids are `jev-9.8.7`. The amended text of the two amended SPECs is unchanged in this round, except AC-JEVO-011 in SPEC-JEV-OPTIN-MEASURE-001. | manager-spec |
| 2026-10-15 | 0.1.2 | Plan repair round 2 after plan-audit iteration 2 (FAIL, overall 0.81, audited commit `a9406fd614d159fd520f839c6ef8ba35e3dcaa2b`; blocking findings D21 to D25; board record d-20261015T075953Z-63b9, which corrects d-20261015T052851Z-42ed). D21: the MCP output carries the change notice in the `notice` key (mcp_jev.go:114–116), so AC-JRM-010 asserts that key on the MCP surface, and the CLI-consumer gap is recorded for card t1803. D22: the unattributed Kickoff path is stated as rule R2 with the LLM's answer, not as fail-closed. D23: the `Ask` order records the usage before the result is returned, and AC-JRM-003 asserts `Result.Usage` and `UsageLog()`. D24: the fixture rule covers positive attributions only, with three named exceptions. D25: the OQ-5 precedent alternative is dropped, and §F matches the open items. The requirement count (14) and the acceptance count (16) are unchanged. | manager-spec |

## Position in the chain

This SPEC amends `SPEC-JEV-CORE-001` and `SPEC-JEV-OPTIN-MEASURE-001`. It is not a fifth member of their chain (`SPEC-JEV-CORE-001` → `SPEC-JEV-OPTIN-MEASURE-001` → `SPEC-JEV-CONSUMERS-001` → `SPEC-JEV-GOAL-DIST-001`). Dispatch order is t1687, then this card (t1768), then t1762, then t1767; this card must land before card t1762.

Definitions used throughout this SPEC:

- **Request alias**: `jev-latest`, the model value the package sends.
- **Interpreted model id**: the `model` value the endpoint returns in its response, which names the version that answered. It is in the versioned shape `jev-<major>.<minor>.<patch>`.
- **Answered call**: a call whose response is OK and reports a valid interpreted model id. An OK response without one is `unattributed`, and it is not an answered call.
- **Verified id**: the interpreted model id held in code as `VerifiedInterpretedModel` (`internal/jev/jev.go`). It is the reference that REQ-JRM-010 compares each answered call against (plan.md §G, OQ-1 and OQ-3).
- **Measurement regression set**: the labelled answer sets that `internal/jevmeasure` runs for each consumer. No such set is committed yet.

---

## §A. Problem Statement

The package sends the versioned id `jev-1.13.0`. It was pinned because the vendor's alias `jev-latest` was taken to make the answering model unattributable.

On 2026-10-15 the leader measured the endpoint with one call each. A request with `model=jev-latest` returned a response whose `model` field is `jev-1.13.0`, with the same answer and the same token count as a request with `model=jev-1.13.0`. The attribution risk that justified the prohibition on the alias is removed once the response's `model` field is recorded.

The current code asserts more than it reads. It writes the pinned id into the usage record before the request is sent, and it never reads the response's `model` field. The record therefore names the answering model without having read the response, and it keeps naming the pinned version when the vendor's version behind the alias moves.

## §B. Goal

Send the request alias. Record the interpreted model id in every answered call's result and usage record, in every measurement artifact, in every rendered Jev answer label, and in every kickoff cross-check reason. Refuse an answer that carries no attributable version. Refuse a measurement that straddles two versions. Announce a change of the interpreted version once. Carry the change into the two amended SPECs and into the model policy rule, in the rule tree and in its template copy.

---

## §C. Requirements (GEARS)

### C.1 — Request and attribution

**REQ-JRM-001** (Ubiquitous) The `internal/jev` package shall send the request alias `jev-latest` as the `model` of every request it constructs, never a versioned model identifier.

**REQ-JRM-002** (Event-driven) **When** an answered call returns, the `internal/jev` package shall record the interpreted model id that the response reports in that call's result and in its usage record.

**REQ-JRM-003** (Event-detected) **When** a response reports no interpreted model id in the versioned shape, the `internal/jev` package shall return a typed unavailable result with availability `unattributed`, and no answers.

**REQ-JRM-004** (Ubiquitous) The usage record shall hold the request alias and the interpreted model id in separate fields, so that a reader cannot take the alias for the version that answered.

**REQ-JRM-005** (Ubiquitous) The `model` field of every usage record, measurement artifact, and rendered Jev answer label shall never carry the request alias, and shall carry the interpreted model id wherever it records an answered call.

### C.2 — Measurement and surfaces

**REQ-JRM-006** (Event-detected) **When** the answered calls of one measurement run report more than one interpreted model id, the measurement harness shall refuse to produce an artifact for that run.

**REQ-JRM-007** (Ubiquitous) Every measurement artifact shall cite the interpreted model id observed across its answered calls, and shall name the request alias in a separately labelled field.

**REQ-JRM-008** (Ubiquitous) The doctor check shall name the request alias and shall not state an interpreted model version, because the check sends no judgment request and cannot observe one.

**REQ-JRM-009** (Ubiquitous) The contract-mode Kickoff `llm+jev` cross-check reason shall name the interpreted model id carried by the Jev answer it reports.

### C.3 — Version change and rule

**REQ-JRM-010** (Event-driven) **When** an answered call reports an interpreted model id that differs from the verified id, the `internal/jev` package shall report one notice line for the client, on the first such call only, naming both ids, the location `internal/jev/jev.go:VerifiedInterpretedModel`, and the shadow-return obligation of REQ-JRM-011.

**REQ-JRM-011** (Ubiquitous) `.claude/rules/moai/development/model-policy.md` and its template copy shall carry one identical section, titled `Jev Model Attribution`. The section shall state the attribution rules of this SPEC and the shadow-return obligation: until the measurement regression set has been re-run against a new interpreted model id and the verified id has been changed to that id, every consumer that shows a Jev answer to a person, or selects on one, returns to shadow, where its answer is recorded and neither shown nor selected on. The two terms are defined by behaviour and introduce no configuration key: *advise-or-above* is a consumer behaviour that shows a Jev answer to a person or selects on one, and *shadow* is the state in which a Jev answer is recorded and neither shown nor selected on.

### C.4 — Gate and amendment discipline

**REQ-JRM-012** (State-driven) **While** `workflow.jev.enabled` is false, the `internal/jev` package shall construct no request and make no network call, as REQ-JEVC-017 requires. This SPEC shall not alter that property.

**REQ-JRM-013** (Ubiquitous) The amendment shall leave each amended SPEC's `sync_commit_sha` unchanged, set its frontmatter `amendment_of` to its own ID, and record the prior completed value as `prior_completed_sha` in its `## Amendments` section.

**REQ-JRM-014** (Event-driven) **When** the sync phase of this SPEC closes, each amended SPEC shall return to `completed` by a change confined to its YAML frontmatter.

---

## §D. Amendments to completed SPECs

This SPEC carries two in-place amendments. Each amendment is recorded in the amended SPEC's own body, as a HISTORY row and a `## Amendments` section. The amended text itself lives in the amended SPEC.

| Amended SPEC | Version | Amended items | Status transition | prior_completed_sha |
|---|---|---|---|---|
| `SPEC-JEV-CORE-001` | 0.4.0 → 0.5.0 | REQ-JEVC-003 (an alias request plus a mandatory recording of the response model), REQ-JEVC-004 (the interpreted model and the request alias in separate fields), §E Q2 wording | `completed → in-progress (amendment)`; `amendment_of: SPEC-JEV-CORE-001` | `c8731f965` |
| `SPEC-JEV-OPTIN-MEASURE-001` | 0.1.0 → 0.2.0 | §B goal phrase, REQ-JEVO-014 (cite the interpreted model id), AC-JEVO-010 wording, AC-JEVO-011 (outputs cite the interpreted version; the alias only in its own field) | `completed → in-progress (amendment)`; `amendment_of: SPEC-JEV-OPTIN-MEASURE-001` | `52c460d30` |

**Sync rule (REQ-JRM-013, REQ-JRM-014).** Neither amended SPEC's `sync_commit_sha` changes in this commit or in this SPEC's sync phase. At the sync phase of this SPEC, the two amended SPECs return to `completed` by a frontmatter-only change. The sync record of the amendment is the §E.4 of this SPEC's `progress.md`.

**Vision.** Not revised. The vision section of `.moai/project/product.md` carries no statement of the model pin or of the alias, and the file contains no occurrence of `jev` or `pinned`. The card's conditional vision revision therefore does not apply.

---

## §E. Exclusions

### Out of Scope — endpoint, credential, and size bounds

- Any change to the endpoint URL, the credential reader, the request size bounds, the secret screen, or the retry policy.
- Any judgment request sent by the doctor check. REQ-JEVC-022 stays as written.

### Out of Scope — measurement results

- Running the measurement harness against the live endpoint, and any accuracy figure. This SPEC changes how a measurement is attributed. It produces no measurement.
- Enabling or disabling any consumer. The shadow-return obligation is doctrine, and this SPEC adds no mode setting.

### Out of Scope — persisted model history

- A persisted record of interpreted model ids on disk. REQ-JRM-010 is met by comparing each answered call against the verified id held in code (`internal/jev/jev.go:VerifiedInterpretedModel`), and no state file is written. That mechanism is decided (plan.md §G, OQ-1, card t1768).

### Out of Scope — documentation and changelog

- `CHANGELOG.md`, `docs-site/`, and the README family. The changelog entry and any documentation correction belong to the sync phase (manager-docs).
- The `.moai/docs/` pages that describe Jev operation. None of them states the pinned id or the alias.

### Out of Scope — consumer mode vocabulary

- A configuration key for consumer modes. The terms the rule uses (*advise-or-above* and *shadow*) describe existing consumer behaviour. They are not introduced as settings (plan.md §G, OQ-2).

---

## §F. Open decisions

The three open decisions of plan-audit iteration 1 are settled (card t1768, board, brain, 2026-10-15). Each decision and its reference is recorded in plan.md §G.

- **OQ-1, the change-notice mechanism.** Decided: option A. It changes REQ-JRM-010 and AC-JRM-010, which is adopted. It does not change the §E exclusion "persisted model history", because option A writes no file.
- **OQ-2, the shadow-return terms.** Decided: behavioural definitions with no configuration key. It fixes the wording of REQ-JRM-011 and the terms of plan.md §D.4.
- **OQ-3, the regression-set reference.** Decided: REQ-JRM-010 compares against the verified id, and the shipping-prohibition clause is not part of this SPEC. It fixes the reference of REQ-JRM-010 and the re-run procedure of REQ-JRM-011, which is to re-run the regression set against the new id and then change the verified id.

Non-blocking items still open (plan.md §G): OQ-4 (the path scope of the rule file), OQ-6 (sync-phase documentation), and OQ-7 (the dispatch of card t1762). None of them changes the wording of a requirement in this file. OQ-5 (the amendment transition precedent) is not open: this SPEC follows the card's instruction, and the precedent is not an option (plan.md §G).
