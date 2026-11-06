---
id: SPEC-JEV-LEDGER-003
title: "Jev decision-ledger cost report (`moai jev cost`)"
version: "0.1.0"
status: in-progress
created: 2026-11-06
updated: 2026-11-06
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "jev, typesafe, ledger, cost, report, read-only, display-only"
tier: S
related_specs: [SPEC-JEV-LEDGER-001, SPEC-JEV-LEDGER-002]
---

# SPEC-JEV-LEDGER-003

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-11-06 | 0.1.0 | Initial plan-phase draft for card t1971 (Class C, Tier S). Scope is the `moai jev cost` report that SPEC-JEV-LEDGER-001's 0.2.0 scope-reduction ruling (factory-leader, operator-delegated; LEDGER-001 HISTORY 0.2.0 and §D.1) migrated out of the original milestone M4. Requirement and criterion ids keep their original numbers (REQ-JEVL-014 cost-report part, REQ-JEVL-015, REQ-JEVL-016; AC-JEVL-014 part, AC-JEVL-015, AC-JEVL-016), carried from the pre-split text at `0c2d2aa4c2` with intent preserved and RED-now citations re-measured on this tree (`9cf8a1fc14`). Card-scope additions over the carried text: the `--json` output form, and the de-dup rule made explicit in REQ-JEVL-015. | manager-spec |

## Position in the chain

This SPEC is split "c" of the original "Jev 1/6" infrastructure card. It builds on SPEC-JEV-LEDGER-001 (card t1762, completed): its landed M1 consumer-config surface and M2 ledger recorder — the recorder now lives at `internal/jevledger` (`internal/jevledger/ledger.go`) after the t1979 debt disposal — write the ledger this report reads. It also builds on the `moai jev` command surface (SPEC-GLM-JEV-KEY-001), whose command group gains the `cost` subcommand. Split "b" (the hook adapter, SPEC-JEV-LEDGER-002, card t1970) has NOT landed, and this SPEC depends on nothing from it: the rows it aggregates are written by the recorder that already landed, and today nothing in the shipped build writes a ledger row (LEDGER-001 position note), so on a real tree the report reads an absent or empty ledger and says so.

The ledger row shape is consumed AS-IS: `internal/jevledger.Row` (`internal/jevledger/ledger.go:46-67`) is not edited, no field is added, renamed, or re-keyed, and no configuration key is added. This SPEC owns exactly one new reader.

---

## §A. Problem Statement

The landed recorder appends one JSON Lines row per question to `.moai/state/jev/decisions.jsonl`, carrying `call_id` and `input_tokens` (the request's byte-length token estimate, repeated on every row of one call). Nothing reads those rows back. An operator who wants to know how much request mass Jev calls have consumed — how many calls, how many estimated input tokens, and eventually how many estimated input tokens per correct decision once outcomes are labelled — has no surface. The original LEDGER-001 plan named this milestone M4 ("input tokens per correct decision"); the 0.2.0 ruling moved it to its own card so the landed SPEC could close.

The report must stay display-only by construction: aggregates over model signals, every token figure labelled an estimate, never a gate input, and never a billed figure.

## §B. Goal

One read-only subcommand, `moai jev cost`, that reads the decision ledger, prints labelled aggregates, exits 0 on every ledger state — absent, empty, or malformed — and makes no Jev call of any kind.

---

## §C. Requirements (GEARS)

### §C.1 The report

**REQ-JEVL-015** (Event-driven) **When** an operator runs `moai jev cost`, the command shall read the ledger `<project root>/.moai/state/jev/decisions.jsonl` and print, without any network call, the number of calls, the total estimated input tokens (`Usage.InputTokens`, a byte-length estimate), the number of rows labelled `correct`, and the estimated input tokens per correct decision. The command shall aggregate once per call: rows sharing a `call_id` are one call, and a call's `input_tokens` enters the total once. Each figure shall be labelled as an aggregate over model signals under the literal heading `model signal aggregate`, and each token figure shall be labelled with the literal unit `estimated input tokens`. With zero correct rows, the per-correct figure shall be reported as `undefined` rather than as a number. With the `--json` flag, the command shall print the same aggregates as one JSON object instead of the labelled text, carrying the skipped-line count as well. The command shall exit 0.

**REQ-JEVL-016** (Event-detected) **When** the ledger file is absent or empty, `moai jev cost` shall report zero calls and exit 0. **When** the ledger carries one or more malformed lines, `moai jev cost` shall skip each malformed line, aggregate the remaining valid rows exactly as REQ-JEVL-015 describes, print `skipped: N` where N is the number of malformed lines, and exit 0. In every case the command shall make no Jev call.

### §C.2 Authority and the boundary

**REQ-JEVL-014** (Ubiquitous) The cost report shall limit its effects to reading the ledger and printing aggregates: no Jev answer shall become an input through it to a permission decision, a completion verdict, a merge approval, a queue mutation, an operator gate, or any hook output. The command shall construct no Jev client, send no request, and add no endpoint, no credential read, no API key, and no cost-incurring call path of its own. Every token figure it prints shall stay labelled an estimate, and no figure shall be presented as charged or billed.

---

## §D. Exclusions

### Out of Scope — SPEC-JEV-LEDGER-002 (card t1970)

- The hook adapter, its call site, shadow byte identity, the deadline machinery, and the credential wiring. This SPEC depends on nothing from -002 and joins none of its files.

### Out of Scope — the ledger row shape and configuration

- The row shape (`internal/jevledger.Row`) is consumed as-is: no new field, no new key, no rename, no reinterpretation. A row-shape change belongs to a SPEC that owns the recorder.
- No new configuration key is added or read. The report takes no setting.

### Out of Scope — money

- Price constants or currency conversion. The cost unit is estimated input tokens (`Usage.InputTokens`, a byte-length estimate), the only cost quantity `internal/jev` records. It is an estimate of request size, not a billed amount. A price table would be new, unverified data in the binary.
- Reading the response's server-reported `usage.input_tokens` field. The recorder neither reads nor records it, so no figure in the report is presented as a billed token count.

### Out of Scope — real consumers and outcome labelling

- Any command, hook, or agent that rewrites `outcome_label` after the fact. Rows are written `unlabeled`, so on real data the report shows zero correct rows and `undefined` per correct until a later card supplies labels.
- Any compiled consumer, question set, or rubric.

### Out of Scope — ledger maintenance and other surfaces

- Ledger rotation, retention, or pruning. Repairing an unterminated final fragment: a fragment joined with a later appended row reads as one malformed line and is counted toward `skipped: N` — reported, not repaired.
- `moai doctor`, `moai init`, `moai web`, and `moai jev --key` are untouched. No TTY-detection output branching: plain text is the default on every terminal; `--json` is explicit.

---

## §E. Open Questions

None open at authoring time. The decisions the plan took while grounding are recorded in `decision-index.md` (D1 allow-list necessity, D2 subcommand placement, D3 `--json` shape, D4 duplicate-row resolution, D5 decode target, D6 non-absent read failures), each with its default rule and the alternative rejected.
