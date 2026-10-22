# Plan — SPEC-JEV-NATIVE-PRIVACY-001

## §A. Scope and boundary decisions

Tier M: one credential-storage seam across the shared private-file leaf and its existing callers; eight requirements and eight criteria. This transfers the entire unfinished core v0.6.0 amendment, not a completed feature or a new Jev client.

### Dependency and privacy design

The credential reader/writer remains in sibling `internal/jevcred`, below the client and CLI/web callers. Shared paths, definitions and native private-file leaves are permitted. Only the private-file Windows implementation may use the already present `golang.org/x/sys/windows` dependency for filesystem security. It must not import credential owners or application/configuration layers, CLI/web, Jev/model or networking. The downward boundary preserves acyclic composition; credential format and disclosure stay in `jevcred`. The `internal/jev` standard-library-only contract and its positive-control import test remain binding.

The old core design diagram's 0600 label describes Unix. On Windows owner-only privacy means a protected, non-inheriting current-owner DACL established before any payload and verified by native readback. Emulated Unix permission bits cannot prove it. This scoped successor qualifies the historical core plan B3/design §1 transitive credential-helper limitation without changing the client import boundary.

## §B. Milestones, in decision-reversibility order

### M1 — Preserve and verify credential integration

Adopt the existing native privacy implementation only after checking it against this successor. Actual code already exists in commits `4717f5034`, `29455779b`, `15c894c6b`, `704c7e324` and later Windows repairs; their existence is not native-platform validation or plan-audit evidence. Record any later implementation changes and run evidence under the implementation owner's phase; this draft does not populate that owner's progress sections.

For fresh and replaced `~/.moai/.env.typesafe`, verify that payload bytes never precede private replacement creation. Native owner/DACL inspection must independently prove the published policy; null, broad, inherited, unknown and unreadable policies fail the save. Round-trip through the real credential owner at its unchanged location. A helper-only check is insufficient. Coordinate the common helper with the active GKI native-privacy amendment while keeping both credential owners separate.

### M2 — Retain existing Unix and disclosure contracts

Keep exact Unix 0600 creation and existing 0644-to-0600 narrowing assertions. Keep schema absence, single location, short-key redaction and the four-character disclosure floor: a key of four characters or fewer discloses only `Configured`. Preserve default-off/fail-open behavior, model attribution and all completed core authority boundaries. No test skip, weaker assertion or new baseline is an acceptance method.

### M3 — Restore the closed historical contract and retain open work

Manager-spec restores core plan/design/acceptance and spec body/version to their v0.5.0 baseline while leaving the status transition to manager-docs. Manager-docs restores only the historical core `status: completed` in a separate commit. Verify all four artifacts against `bf0a35f10e423d8f0768e637b74a58e143a5d7d6` and `f009522d162ea5ce72905a54ccd22a1f2fbe47a4`; leave original progress/research unchanged. The v0.6.0 Git history is preserved and its unfinished obligations live here. Do not rewrite core or GKI sync records or claim Windows completion.

### M4 — Obtain current evidence and closure review

Run the unchanged doctrine and standard-library import checks, scoped SPEC validation and the official whole-corpus baseline ratchet. Obtain native Windows credential integration evidence, current-head code/security review, independent audit and required CI through their real owners. This draft has no completed plan/run/sync signal; the decision index lists measurements still needed. Prior core close evidence remains historical only.

## §C. Risks and evidence boundaries

- Windows mode emulation can falsely suggest privacy; inspect the actual owner and protected DACL independently.
- Shared-leaf broadening could accidentally affect the Jev client; retain its stdlib import contract and helper's downward boundary.
- Restoring historical files alone would lose native obligations; the mapping in spec.md and criteria below retain the full prior v0.6.0 scope.
- A passing doctrine check proves lifecycle consistency, not native Windows security or independent audit. Native platform and current-head review gaps remain until actual observations arrive.
