---
id: SPEC-JEV-LEDGER-003
title: "Acceptance — Jev decision-ledger cost report"
version: "0.1.0"
created: 2026-11-06
updated: 2026-11-06
---

# Acceptance — SPEC-JEV-LEDGER-003

**Document-level tree pin:** every RED-now cell below was measured on `9cf8a1fc147605a8221206616d04fb43569961d8` (card worktree `t1971`, branch `WT-spec-jev-ledger-3`, 2026-11-06). RED-now cells cite an evidence-ledger entry (§L) that holds the command, the verbatim stdout, and the exit code.

Each criterion has the four parts the rule requires: the Given-When-Then, the RED-now cell (§L), the green path (milestone + passing output), and the three-part check spec: (a) WHEN, (b) failing INPUT, (c) who sees the red. All three criteria are final at M4 (plan.md §E closure map); there are no partial milestones.

**Carried-id note.** REQ/AC numbers 014/015/016 are SPEC-JEV-LEDGER-001's original numbers, carried unchanged per its §D.1. The old AC-JEVL-016 template-default clause ("the template workflow.yaml") lives in LEDGER-001 AC-JEVL-017 since the 0.2.0 ruling ("Its template partial moved to AC-JEVL-017 here") and is not restated here.

---

## §A. Criteria

### AC-JEVL-014 — boundary: the report is not a Jev caller (REQ-JEVL-014) [release-blocking]

- **Given** the tree after M4, `jev_cost.go` present.
- **When** `TestJevCallPath_HasExactlyTheDeclaredConsumers` runs **unchanged** — this SPEC adds no entry to its allow-list — together with `TestJevCost_Boundary_NoClientOrEndpoint`, which source-scans `jev_cost.go` for an `internal/jev` import, a URL literal, and a credential read (`jevcred` or credential-environment access).
- **Then** both PASS. The package guard passing with `jev_cost.go` present and absent from its allow-list is the by-name proof that the cost report imports no `internal/jev` — any such import without an allow-list entry fails that guard — hence constructs no client and sends no request (REQ-JEVL-014). `TestJevCallPath_UnreachableFromDecisionSurfaces` passes unchanged (`jev_cost.go` is not a decision surface and joins no list).
- **Carried-text deviation (recorded, decision D1).** LEDGER-001 §D.1 described this part as "the by-name allow-list entry for `jev_cost.go`", premised at split time on the report importing `internal/jev` as the adapter does. The reader's actual imports are `internal/jevledger` and the standard library (plan.md §D), so an allow-list entry would be a dead row misdescribing the file. The boundary is asserted as the non-importer proof above — mechanically equivalent (the same guard fires on the same mutants) and honest about the imports.
- RED-now: §L E1 (`ls internal/cli/jev_cost.go` → absent, exit 1). Red because the subject file does not exist, so the non-importer proof has nothing to prove.
- Green path: M4. `--- PASS: TestJevCallPath_HasExactlyTheDeclaredConsumers ` (unchanged) and `--- PASS: TestJevCost_Boundary_NoClientOrEndpoint `.
- Mutant probes: (1) importing `internal/jev` in `jev_cost.go` (e.g., to construct a client for a "live" estimate) → the package guard fails on the undeclared importer and the source scan fails → red. (2) adding a URL literal or a `jevcred` read → the source scan fails → red. (3) routing the import through a helper file so `jev_cost.go` itself stays clean — the same evasive mutant LEDGER-001 named — is caught by the package-wide importers scan: any file in `internal/cli` importing `internal/jev` without an allow-list entry fails the guard.
- Check spec: (a) CI (`go test ./internal/cli/`). (b) the mutants above. (c) Test job exit 1.

### AC-JEVL-015 — cost report aggregates (REQ-JEVL-015)

- **Given** a fixture ledger with 3 calls (`call_id` a, b, c; `input_tokens` 100, 200, 300, repeated across 2 rows each), of which 2 rows are labelled `correct`. A second fixture with zero `correct` rows.
- **When** `moai jev cost` runs against each fixture (plain text, and once with `--json`).
- **Then** the output reports calls = 3, estimated input tokens = 600, correct = 2, and estimated input tokens per correct decision = 300. It contains the literal label `estimated input tokens` on each token figure and the literal heading `model signal aggregate`. It exits 0. No Jev client is constructed on the path — the report imports no `internal/jev` (AC-JEVL-014's proof), so no call is possible; this replaces LEDGER-001's counting-transport assertion, which assumed the transport seam of the not-yet-landed adapter. With zero correct rows, the per-correct figure reads `undefined` (text) / `null` (`--json`). No figure is labelled as charged or billed. The `--json` object carries the same aggregates plus `skipped`.
- RED-now: §L E1 (file absent — the reader does not exist).
- Green path: M4. `--- PASS: TestJevCost_Aggregates `, `--- PASS: TestJevCost_ZeroCorrectUndefined `, `--- PASS: TestJevCost_JsonMatchesTextAggregates `.
- Mutant probes: (1) summing tokens per row instead of per call gives 1200, which the test rejects. (2) printing the token figures without the `estimated input tokens` label → red. (3) printing a number instead of `undefined` at zero correct → `TestJevCost_ZeroCorrectUndefined` red. (4) dropping the correct count or renaming the heading → the literal-label assertions fail → red.
- Check spec: (a) CI. (b) any mutant above. (c) Test job exit 1.

### AC-JEVL-016 — absent / empty / malformed ledger (REQ-JEVL-016)

- **Given** (i) no ledger file, (ii) an empty one, (iii) one with a non-JSON line between two valid rows of two different calls (`call_id` a with `input_tokens: 100`, `call_id` b with `input_tokens: 200`, one row labelled `correct`).
- **When** `moai jev cost` runs against each.
- **Then** (i)/(ii) report calls = 0, exit 0. (iii) reports calls = 2, estimated input tokens = 300, correct = 1, estimated input tokens per correct decision = 300, and `skipped: 1`, exit 0. The valid rows are aggregated, not discarded. No Jev call is made in any case (AC-JEVL-014's proof).
- RED-now: §L E1 (file absent).
- Green path: M4. `--- PASS: TestJevCost_AbsentAndMalformed ` (3 subtests), with the `skipped`-in-JSON assertion carried by `TestJevCost_JsonMatchesTextAggregates`.
- Mutant probes: (1) the reader aborts on the first malformed line → exit 1 → red. (2) the reader reports zero calls once any line is malformed → calls = 0 in (iii) → red. (3) the reader drops the malformed line without counting it → `skipped:` missing → red.
- Check spec: (a) CI. (b) any of the three mutants. (c) Test job exit 1.

## §B. Edge cases covered above

- An unterminated final fragment joined with a later appended row reads as one malformed line and counts toward `skipped: N` — reported, not repaired (spec §D carry-over of LEDGER-001's ledger-maintenance exclusion).
- Rows that decode but carry zero-valued fields count as valid (plan.md §G residual).

## §C. Continued-firing answers (§1.3)

| Check family | If it stopped firing tomorrow, what would be different? |
|---|---|
| Go tests in `internal/cli` (AC-014, -015, -016) | Each test name is unique and listed in `progress.md` §E.2 at run time; the self-verification matrix requires each named test's `--- PASS` line, and a dropped selector shows as a lower test count in `go test -v` output — a missing line is a gap, not a pass. |
| The package importers guard (AC-014) | The guard keeps its positive control (`doctor_jev.go` must be seen as an importer), so a scan that silently swept zero files fails. With `jev_cost.go` present and unlisted, a green run IS the non-importer proof; if the file were renamed or deleted, the boundary source scan fails on the missing subject. |

## §D. Quality gates and Definition of Done

- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go test -cover` for the new files: covered by their tests (package-wide coverage is not re-baselined).
- `golangci-lint run` shows no NEW findings against the baseline.
- No template change: no `make build` output drift is expected (nothing under `internal/template/` is touched).
- No user-facing question channel anywhere in the new code. The C-HRA-008 grep returns 0.
- All criteria PASS at M4 with verbatim `--- PASS` lines in `progress.md` §E.2.

## §L. Evidence ledger (RED-now cells)

**Tree:** `9cf8a1fc147605a8221206616d04fb43569961d8` (`git rev-parse HEAD` in card worktree `t1971`, branch `WT-spec-jev-ledger-3`, 2026-11-06, plan-phase authoring run; `git status --short` empty at measurement).

**Collection method.** The command was run exactly once, with its exit status captured inside the same shell invocation by a wrapper of the form `sh -c '<command>; echo "exit=$?"'`. The wrapper printed the cited command's own status; the wrapper is the measuring harness, not part of the cited command. The Command column is the bare single-invocation form a reader re-runs.

| Id | Command | Verbatim stdout | Exit | Tree |
|---|---|---|---|---|
| E1 | `ls internal/cli/jev_cost.go` | `ls: internal/cli/jev_cost.go: No such file or directory` | 1 | `9cf8a1fc14` |

E1 is shared by AC-JEVL-014, AC-JEVL-015, and AC-JEVL-016: the absent reader is the one honest red all three share. Each criterion's subject — the file, its boundary properties, its behaviour — does not exist yet, which is why the red is right-reason red: green requires M4's reader, and nothing short of this card's own work could make the file appear.
