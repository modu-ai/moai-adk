# SPEC-UPDATE-HOOK-DELIVERY-001 — Acceptance Criteria

> ACs are binary-testable Given-When-Then scenarios. Option-gated ACs (AC-UHD-004..007, AC-UHD-013) carry a gate marker; at run-phase entry the non-selected option's gated ACs are marked N/A in progress.md (decision recorded at M1). "Template" = the shipped `internal/template/templates/.claude/settings.json` of the build under test; "user file" = an existing project's `.claude/settings.json`.

## §1 Core acceptance scenarios

**AC-UHD-001 — New event key delivered (guard; green today, must stay green)**
Given a user file lacking the `hooks.SessionStart` key entirely, When `moai update` runs, Then the user file contains the template's `hooks.SessionStart` block and every non-hook key is byte-identical to before.
Guard AC — flips only if a fix regresses today's behavior; any RED here is a fix defect, not progress.

**AC-UHD-002 — User deletion preserved, no resurrection (guard; green today, must stay green)**
Given a user file whose carried event key array has had one template-origin entry deleted by the user, When `moai update` runs, Then that entry remains absent in the user file after update and the user's remaining entries are unchanged.
Guard AC — this is the protective behavior any resolution option MUST keep (REQ-UHD-002).

**AC-UHD-003 — The core gap is resolved (RED today; flips per selected option)**
Given a user file that carries a hook event key whose array is missing an entry present in the template's same key, When `moai update` runs, Then the gap is resolved: either the missing entry is delivered into the user file (Option A), or the gap is reported to the user with the event key named (Option B), or the no-op is explicit in the update output/docs (Option C) — and in NO case does update complete silently with neither delivery nor report.
RED-now cell: on tree d592b0551 the scenario today ends with the user file unchanged and zero output mentioning hooks — observed from the code chain (base.go:112-131 + strategies.go:427-429); the RED is re-executed mechanically at M2 as the characterization/RED test, pinning the failing command + output per verification-completeness §2.
Adoption gate: AC-UHD-003 counts as adopted only after M2 lands the four mechanical elements in progress.md §E.2 — (1) the failing test command, (2) its verbatim RED output and its exit code, (3) the fixture that produced it (user file + template pair), (4) the tree SHA the run executed against.
Green path: M3 (mechanism) flips the selected option's branch.

**AC-UHD-004 — [Option A gate] Delivery without resurrection**
Given Option A is selected and a user file that (a) previously deleted template entry E from a carried event key and (b) lacks a newly-shipped template entry F in that same key, When `moai update` runs, Then F is present in the user file and E remains absent.
Flips at M3 (mechanism) — depends on M1's identity scheme.

**AC-UHD-005 — [Option A gate] User-modified entries never overwritten**
Given Option A is selected and a user file carrying a template-origin entry whose fields the user edited (e.g. changed a matcher or command), When `moai update` runs, Then the user's edited version survives byte-identical.
Flips at M3.

**AC-UHD-006 — [Option B gate] Detection report is read-only and specific**
Given Option B is selected and a user file missing two template hook entries, When `moai doctor` (or `moai update`) runs, Then the output names each missing entry with its event key and remediation guidance, exits with the documented status, and the user file's mtime and content are unchanged.
Flips at M3/M4.

**AC-UHD-013 — [Option B gate] Missing / hook-free settings.json tolerated as informational**
Given Option B is selected and the project's settings.json is absent entirely (or carries no `hooks` key), When `moai doctor` (or `moai update`) runs the hook-delivery check, Then the check reports informational status (not an error state), `moai update` completes normally, and no settings.json is created or written by the check.
Covers REQ-UHD-009 (Option B branch — the tolerance half, complementing AC-UHD-006's reporting half). Flips at M4.

**AC-UHD-007 — [Option C gate] Explicit no-op documented**
Given Option C is selected and the gap scenario of AC-UHD-003, When the user consults the shipped documentation (and, if chosen, the update output), Then the limitation, its blast radius (entry classes not delivered), and the manual remediation path are stated.
Flips at M4.

## §2 Robustness and idempotence

**AC-UHD-008 — Invalid JSON handled gracefully**
Given a user file that is not valid JSON, When the resolution mechanism runs (update or doctor), Then `moai update` completes without corrupting the file, and the anomaly is reported (message naming settings.json); no partial write occurs.
Applies under every option; flips at M3.

**AC-UHD-009 — Non-array hook value handled gracefully**
Given a user file where a hook event key holds a non-array value, When the resolution mechanism runs, Then the mechanism reports and skips that key without modifying it; other keys resolve normally.
Applies under every option; flips at M3.

**AC-UHD-010 — Idempotence**
Given the gap scenario of AC-UHD-003 already resolved by a first `moai update`, When a second `moai update` runs with no template change, Then no further settings.json change occurs and no duplicate report/entry appears.
Applies under every option; flips at M3.

## §3 PRESERVE characterization

**AC-UHD-011 — Non-hook keys byte-identical**
Given a user file with non-hook keys (permissions, env, model, …) including user-only keys, When `moai update` runs under the selected option, Then all non-hook keys are byte-identical before and after (JSON-serialized comparison of the non-hook subtree).
Characterization — green today, must stay green; flips to RED only on fix defect.

**AC-UHD-012 — Boundary: pre-commit axis untouched**
Given the completed implementation, When `git diff <baseline>..HEAD -- internal/` is inspected and the affected packages are grepped, Then no diff touches `installPreCommitHookOptional` or any `.git/hooks/` write path (0 matches for changes to that function/callers), and the t461-owned surface behaves identically.
Scope guard — green at M5 by construction; any RED is scope drift, escalate per D-NEW-1.

## §4 Quality gates

- `go test ./internal/cli/update/... ./internal/merge/...` exit 0; touched packages ≥85% coverage (90%+ where `internal/cli` is touched, per its critical-package target).
- `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- golangci-lint: zero NEW findings vs the M1 pre-flight baseline.
- RED evidence for AC-UHD-003 captured verbatim at M2 before GREEN (verification-completeness two-cell adoption).

## §5 Definition of Done

All non-N/A ACs PASS with verbatim evidence in progress.md §E.2; the option decision and N/A markings recorded; boundary grep (AC-UHD-012) clean; sync-phase entry criteria met.
