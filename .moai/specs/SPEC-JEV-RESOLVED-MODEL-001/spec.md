---
id: SPEC-JEV-RESOLVED-MODEL-001
title: "Jev request alias jev-latest with the interpreted model recorded per answer"
version: "0.1.0"
status: draft
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

## Position in the chain

This SPEC amends `SPEC-JEV-CORE-001` and `SPEC-JEV-OPTIN-MEASURE-001`. It is not a fifth member of their chain (`SPEC-JEV-CORE-001` → `SPEC-JEV-OPTIN-MEASURE-001` → `SPEC-JEV-CONSUMERS-001` → `SPEC-JEV-GOAL-DIST-001`). Dispatch order is t1687, then this card (t1768), then t1762, then t1767; this card must land before card t1762.

Definitions used throughout this SPEC:

- **Request alias**: `jev-latest`, the model value the package sends.
- **Interpreted model id**: the `model` value the endpoint returns in its response, which names the version that answered. It is in the versioned shape `jev-<major>.<minor>.<patch>`.
- **Measurement regression set**: the labelled answer sets that `internal/jevmeasure` runs for each consumer. No such set is committed yet (see plan.md OQ-3).

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

**REQ-JRM-001** (Ubiquitous) The `internal/jev` package shall send the request alias `jev-latest` as the model of every request it constructs, and no request it constructs shall carry a versioned model identifier.

**REQ-JRM-002** (Event-driven) **When** an answered call returns, the `internal/jev` package shall record the interpreted model id that the response reports in that call's result and in its usage record.

**REQ-JRM-003** (Event-detected) **When** a response reports no interpreted model id in the versioned shape, the `internal/jev` package shall return a typed unavailable result with availability `unattributed`, and no answers.

**REQ-JRM-004** (Ubiquitous) The usage record shall hold the request alias and the interpreted model id in separate fields, so that a reader cannot take the alias for the version that answered.

**REQ-JRM-005** (Ubiquitous) The model field of every usage record, every measurement artifact, and every rendered Jev answer label shall carry the interpreted model id, and shall never carry the request alias.

### C.2 — Measurement and surfaces

**REQ-JRM-006** (Event-detected) **When** the answered calls of one measurement run report more than one interpreted model id, the measurement harness shall refuse to produce an artifact for that run.

**REQ-JRM-007** (Ubiquitous) Every measurement artifact shall cite the interpreted model id observed across its answered calls, and shall name the request alias in a separately labelled field.

**REQ-JRM-008** (Ubiquitous) The doctor check shall name the request alias and shall not state an interpreted model version, because the check sends no judgment request and cannot observe one.

**REQ-JRM-009** (Ubiquitous) The contract-mode Kickoff `llm+jev` cross-check reason shall name the interpreted model id carried by the Jev answer it reports.

### C.3 — Version change and rule

**REQ-JRM-010** (Event-driven) **When** the interpreted model id of an answered call differs from the interpreted model id that the measurement regression set was last run against, the `internal/jev` package shall report one notice line for the client, naming both ids and the shadow-return obligation of REQ-JRM-011.

**REQ-JRM-011** (Ubiquitous) `.claude/rules/moai/development/model-policy.md` and its template copy shall carry one identical section, titled `Jev model attribution`. The section shall state the attribution rules of this SPEC and the shadow-return obligation: until the measurement regression set has been re-run against a new interpreted model id, every consumer that shows a Jev answer to a person, or selects on one, returns to shadow, where its answer is recorded and neither shown nor selected on.

### C.4 — Gate and amendment discipline

**REQ-JRM-012** (State-driven) **While** `workflow.jev.enabled` is false, the `internal/jev` package shall construct no request and make no network call, as REQ-JEVC-017 requires. This SPEC shall not alter that property.

**REQ-JRM-013** (Unwanted) An amendment of a completed SPEC shall not change that SPEC's `sync_commit_sha`. The amended SPEC returns to `completed` with only its frontmatter `status` changed, and the sync record of the amendment is kept in this SPEC's progress record.

---

## §D. Amendments to completed SPECs

This SPEC carries two in-place amendments. Each amendment is recorded in the amended SPEC's own body, as a HISTORY row and a `## Amendments` section. The amended text itself lives in the amended SPEC.

| Amended SPEC | Version | Amended items | Status transition | prior_completed_sha |
|---|---|---|---|---|
| `SPEC-JEV-CORE-001` | 0.4.0 → 0.5.0 | REQ-JEVC-003 (an alias request plus a mandatory recording of the response model), REQ-JEVC-004 (the interpreted model and the request alias in separate fields), §E Q2 wording | `completed → in-progress (amendment)`; `amendment_of: SPEC-JEV-CORE-001` | `c8731f965` |
| `SPEC-JEV-OPTIN-MEASURE-001` | 0.1.0 → 0.2.0 | §B goal phrase, REQ-JEVO-014 (cite the interpreted model id), AC-JEVO-010 wording, AC-JEVO-011 (outputs cite the interpreted version; the alias only in its own field) | `completed → in-progress (amendment)`; `amendment_of: SPEC-JEV-OPTIN-MEASURE-001` | `52c460d30` |

**Sync rule (REQ-JRM-013).** Neither amended SPEC's `sync_commit_sha` changes in this commit or in this SPEC's sync phase. At the sync phase of this SPEC, the two amended SPECs return to `completed` by a frontmatter-only status change. The sync record of the amendment is the §E.4 of this SPEC's `progress.md`.

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

- A persisted record of interpreted model ids on disk. REQ-JRM-010 is met by comparing each answered call against the interpreted model id recorded in code, and no state file is written. If the lead chooses a persisted record instead (plan.md OQ-1), this exclusion and REQ-JRM-010 change together.

### Out of Scope — documentation and changelog

- `CHANGELOG.md`, `docs-site/`, and the README family. The changelog entry and any documentation correction belong to the sync phase (manager-docs).
- The `.moai/docs/` pages that describe Jev operation. None of them states the pinned id or the alias.

### Out of Scope — consumer mode vocabulary

- A configuration key for consumer modes. The terms the rule uses (shadow, the decision of a consumer that shows or selects on an answer) describe existing consumer behaviour. They are not introduced as settings.

---

## §F. Open decisions

Open decisions are recorded in plan.md §G with `[NEEDS CLARIFICATION]` markers, as the marker convention requires. None of them changes the wording of a requirement in this file. OQ-1 decides the mechanism behind REQ-JRM-010, and it must be settled before the run phase enters that requirement.
