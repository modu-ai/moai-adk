---
id: SPEC-JEV-NATIVE-PRIVACY-001
title: "Jev native credential privacy successor"
version: "0.1.0"
status: draft
created: 2026-10-22
updated: 2026-10-22
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "internal/jevcred + internal/privatefile"
lifecycle: spec-anchored
tags: "jev, credential, native-privacy, windows"
tier: M
amendment_of: SPEC-JEV-CORE-001
depends_on:
  - SPEC-JEV-CORE-001
  - SPEC-GLM-KEY-INPUT-001
---

# SPEC-JEV-NATIVE-PRIVACY-001

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-10-22 | 0.1.0 | Transfer the unfinished core v0.6.0 native-privacy amendment into a draft successor; restore the historically completed core v0.5.0 without claiming that Windows work is complete. | manager-spec |

## Amendments

- Parent completed version: SPEC-JEV-CORE-001 v0.5.0. Its four core artifacts match main `bf0a35f10e423d8f0768e637b74a58e143a5d7d6` and the model-alias sync close `f009522d162ea5ce72905a54ccd22a1f2fbe47a4`.
- prior_completed_sha: c8731f965. This is the unchanged original core progress §E.4 sync record for the initial capability, not the later v0.5.0 body snapshot. The later amendments and their separate sync records remain historical facts.
- The prior in-place native-privacy attempt is commit `a41b099f4541a428923b887a29fd934cb6dc0e7c` (2026-10-21, core v0.6.0). Its code and Git history are retained; its unfinished obligations move here, rather than being marked completed. The original core progress and research are not rewritten.
- Rationale: SPEC-AUTONOMY-GATE-REWIRE-001 AC-GR-022 requires the completed core contract to remain closed. The generic completed-to-in-progress amendment mechanism cannot discharge that specific obligation. The actual doctrine test failed on the open core status; no detector or assertion is relaxed.
- Scope: this successor qualifies core REQ-JEVC-018, plan B3/M1b, design §1 and AC-JEVC-010/011 only for credential storage and its downward dependency boundary. Its native Windows clauses supersede the historical storage clauses where Unix mode bits or a transitive stdlib-only credential-helper rule would contradict native protection. The `internal/jev` client remains standard-library-only. No Jev authority, model, gate or disclosure permission expands.
- Shared predecessor: the active SPEC-GLM-KEY-INPUT-001 v0.3.0 native-privacy amendment owns the corresponding GLM credential boundary. Its pending status and historical unreachable `d06771f07` sync reference, with reachable batch close `2f449e189e6ddda6d61740aa8445386a9cb210e1`, remain unchanged. This dependency does not attest that GKI is complete.
- Closure remains pending real native integration evidence, preserved Unix regressions, code/security review, independent audit and required CI. Existing implementation commits precede this transfer; this draft is not a retroactive plan audit or new completion receipt.

## §A. Problem and goal

Unix permission emulation does not establish Windows credential privacy. A credential replacement must be owner-only before receiving its payload, and a successful save must be supported by native access-policy readback. Preserve the single owner, location, format and disclosure boundary while keeping the completed Jev authority contract intact.

## §B. Requirements (GEARS)

**REQ-JNP-001** (Ubiquitous) The credential owner shall remain the single reader/writer of `~/.moai/.env.typesafe`, outside the repository, with unchanged credential format, escaping and path resolution.

**REQ-JNP-002** (Ubiquitous) On Unix the credential writer shall create and replace the credential with exact mode 0600, including narrowing an existing 0644 file.

**REQ-JNP-003** (Event-driven) When writing a Windows credential, the credential writer shall establish a protected, non-inheriting DACL with the current user as owner and exactly one explicit owner access grant before credential bytes enter the replacement.

**REQ-JNP-004** (Event-driven) When publishing or reading back a Windows credential save, the credential writer shall verify native owner and access policy; a null, broader, inherited, unknown or unreadable policy shall produce a save error rather than a success result.

**REQ-JNP-005** (Ubiquitous) The credential owner shall remain below CLI/web/client callers and may use shared low-level paths, definitions and private-file leaves. Only the Windows private-file leaf may use the existing `golang.org/x/sys/windows` dependency for filesystem security. The helper shall not import credential owners, CLI/web, application configuration, Jev/model or network layers; `internal/jev` shall retain its standard-library-only import contract.

**REQ-JNP-006** (Ubiquitous) The amendment shall preserve the default-off gate, fail-open model-call behavior, resolved-model attribution, schema exclusion, single credential ownership, short-key disclosure floor, masking and all completed core authority restrictions and exceptions.

**REQ-JNP-007** (Ubiquitous) The lifecycle repair shall preserve the four completed core v0.5.0 artifacts byte-for-byte, retain original progress/research and all prior history, and record the unfinished native-privacy work only in this successor without asserting new completion or independent audit.

**REQ-JNP-008** (Capability gate) Where native privacy work is proposed for closure, the successor shall require actual Windows fresh/replacement credential integration and independent native readback evidence, unchanged Unix narrowing and disclosure regressions, current-head code/security review, independent audit and required CI before a completion claim.

## §C. Exclusions

### Out of Scope — authority and completion

- No Jev-alone gate decision, new ranking permission, model/config behavior change, schema credential field or alternate credential location is introduced.
- No checker removal, test skip, permission-bit substitute, relaxed baseline, fabricated receipt, retrospective audit PASS or claim that the old v0.6.0 work is complete is permitted.
- The native filesystem helper does not own credential content or disclosure, and no stdlib exception extends to the Jev client.

## §D. Traceability

| Historical clause [REF] | Successor obligation |
|---|---|
| Core REQ-JEVC-018 [REF] | REQ-JNP-001 through REQ-JNP-004 |
| Core plan B3 and design §1 [REF] | REQ-JNP-005 |
| Core plan M1b and AC-JEVC-010/011 [REF] | REQ-JNP-002 through REQ-JNP-004, REQ-JNP-008 |
| Core authority/model/schema/disclosure contract [REF] | REQ-JNP-006 |
| Core v0.6.0 amendment history [REF] | REQ-JNP-007 |
