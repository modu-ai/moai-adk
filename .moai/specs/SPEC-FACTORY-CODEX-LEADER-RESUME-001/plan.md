# plan.md — SPEC-FACTORY-CODEX-LEADER-RESUME-001

## §A Context

Card t1745 runs on branch `WT-codex-leader-resume`. HEAD was `1ced8cab9` when this plan was authored, and the tree was clean. The card asks for the Codex factory leader entry to re-adopt the prior run on a resume-shaped start, reusing the shared adoption path with no second copy of its logic.

The plan starts with a decision gate, not with code. The re-verified coordinates below show that the target branch has no reachable production caller (spec.md §0).

Coordinates re-read in this run (file and line on this tree):

| Item | Coordinate | Observed on this tree |
|---|---|---|
| Candidate finder | `internal/homestate/factory_run_adopt.go:72` | present |
| Adoption write | `internal/homestate/factory_run_adopt.go:112` | present |
| Adoption helper and its call of the finder and the write | `internal/cli/factory_run_adopt.go:64`; calls at `:73` and `:106` | present |
| Resume-carrier judge | `internal/cli/factory_run_adopt.go:126` | present |
| Shared combined gate | `internal/cli/factory_run_adopt.go:163-166` | present |
| Shared start recorder | `internal/cli/factory_run_adopt.go:173-184` | present |
| cc leader branch | `internal/cli/cc.go:199` (prior id), `:201` (publication), `:214` (gate), `:218` (recorder) | present |
| glm leader branch | `internal/cli/glm.go:262`, `:264`, `:274`, `:281` | present |
| Codex entry parse and enable | `internal/cli/codex_factory.go:21`; enable at `:68` (lane token only) | present |
| Codex entry function and leader branch | `internal/cli/codex_factory.go:90`; leader publication `:131`, start record `:135` | present; no production path reaches the leader branch |
| Codex entry non-test caller | `internal/cli/codex_launcher.go:1214` | reached only when the parsed entry is enabled |
| Lane return before the Codex launch | `internal/cli/codex_launcher.go:753` | lane entries return here |
| Codex `-f` refusal line | `internal/cli/codex_launcher.go:801-802` | present |
| Codex-led run refusal | `internal/cli/factory.go:674-689` | present; checks only the run's stored lead backend |
| Commit `0ae31523f` (card t1707) | `internal/cli/cc.go`, `internal/cli/factory_run_adopt.go`, `internal/cli/glm.go` | matches the brief |

The brief's coordinates match except one: the combined gate sits at `:163-166`, not "around 160-164". The brief's line counts for `codex_factory.go` (231 lines), `parseCodexFactoryEntry` (~21), and `enterCodexFactory` (~90) also match.

## §B Known Issues

- **B1 — Reachability (blocking).** spec.md §0. Q1 settles it before any milestone below runs.
- **B2 — Carrier vocabulary.** The judge at `internal/cli/factory_run_adopt.go:126` reads Claude option spellings only. Applied to a Codex start, it would never fire for a Codex resume. A judge that fires on a Codex-only spelling is a new user-visible surface. [NEEDS CLARIFICATION: which Codex launch spelling counts as a resume-shaped start (decision Q2)]
- **B3 — Backend-blind refusal of Codex-led runs.** The automatic path calls `refuseCodexLeaderRun` (`internal/cli/factory.go:674-689`), which refuses any run whose stored lead backend is codex. Its warning text names a claude leader. A reopened Codex leader would be refused re-adoption of its own Codex-led run by this check. The check must become backend-aware, and the change runs through the cc and glm path too, so their outcomes for claude-led and glm-led runs must stay identical (AC-007, AC-008).
- **B4 — Untested shared gate, and an untested refusal.** Nothing drives the combined gate or the start recorder directly (no test references them). The automatic path's live-legacy-peer refusal (`internal/cli/factory_run_adopt.go:100-105`) has no test; only the explicit path's refusal is tested (`internal/cli/factory_run_boundary_m1_test.go:57`). The RED set adds characterization coverage for both before the Codex entry is wired, so the shared behavior is pinned for every leader.
- **B5 — Ordering.** The Codex branch must read the prior run id before the publication at `internal/cli/factory.go:786-807` runs, as `internal/cli/cc.go:199` reads it before `:201`. `leaderRunID` reads the environment before publishing (`internal/cli/factory_launch_helpers.go:72-80`). The Codex branch captures nothing today.
- **B6 — Start-record arguments.** The Codex branch records with the derived-capacity marker and an empty spec id (`internal/cli/codex_factory.go:135`). The shared recorder takes its values from the caller, so the Codex call must pass the same values the cc call passes for its arguments, with the Codex backend and the derived-capacity marker.
- **B7 — Signature.** The Codex entry function takes the root, the parsed entry, and the timing collector (`internal/cli/codex_factory.go:90`). It receives no launch arguments, so it cannot judge a carrier. Under Q1 `reopen`, the carrier can reach it through a new parameter, which changes the one production call site (`internal/cli/codex_launcher.go:1214`) and two existing test call sites (`internal/cli/factory_join_discovery_test.go:404`, `internal/cli/factory_lane_backend_test.go:125`), or through a field on the parsed entry. Choosing between them is part of M1. A RED test that names a symbol absent from the tree is red for a build reason and does not satisfy AC-006(c) until it observes an assertion.
- **B8 — Codex hook coverage.** spec.md Out of Scope. It matters only if Q1 is `reopen`.
- **B9 — Conflicting prior text.** SPEC-QUOTA-AWARE-SCHEDULING-001 `plan.md:150` (U22) describes the Codex factory-entry path as a live path and says "Not run". The later overlay measurement in SPEC-LAUNCHER-ENTRY-FLAGS-001 supersedes it for this path. Do not cite U22 as evidence.

## §C Pre-flight (the gate before any RED commit)

- **C1.** Q1 carries an operator verdict in `decision-index.md`.
  - Verdict `keep`: stop. No RED or GREEN commit lands. The SPEC moves to `rejected` through the Status Transition Ownership Matrix (`* → rejected`, an orchestrator decision recorded by manager-docs), with a HISTORY row that cites REQ-SD-004.
  - Verdict `reopen`: continue to C2.
- **C2.** Q2 carries an operator verdict (reopen only).
- **C3.** Re-trace reachability on the implementation tree. `git grep -n "enterCodexFactory(" -- internal/cli` must list one non-test call site, and the parse must enable the entry only at the lane token. Record the verbatim output in progress §E.2.
- **C4.** The SPEC ID is still unique in the catalogue. Check it with `mcp__moai__spec_progress` and the same `project_root` the session uses.
- **C5.** Baseline, observed and not claimed. Before any edit, run the existing adoption tests and record the verbatim output in progress §E.2:

```
go test ./internal/cli -run '^(TestAdoptResumedLeaderRunAdoptsSingleCandidate|TestAdoptResumedLeaderRunWarnsOnMultipleCandidates|TestStrictResumeCarrierSeen)$' -count=1 -v
```

## §D Constraints

- Test selectors are anchored (`-run '^Name$'`). A whole-package run of `./internal/cli` is heavy and goes through `moai slot acquire --resource <name>` and `moai slot release` (`.claude/rules/local/gitflow-lane-protocol.md` §8).
- Every new test builds its root under `t.TempDir()` with the home variables redirected, so no test touches the real `~/.moai`. The pattern is `adoptSandbox` (`internal/cli/factory_run_adopt_test.go:20-31`).
- No test starts a real `codex` or `claude` process. The launch seams are substituted.
- Development mode is TDD (`.moai/config/sections/quality.yaml`, `constitution.development_mode: tdd`). The order is RED commit, GREEN commit, then REFACTOR commit.
- New tests go into a new file, `internal/cli/codex_factory_resume_test.go`. The RED commit edits no existing test.
- Each commit names the card (t1745) and the SPEC ID, which carry the traceability that the branch name no longer does (AGENTS.md §3).
- The plan phase writes no Go code and makes no commit. Staging is by explicit pathspec only.

## §E Self-verification (plan audit and run-phase checks)

- **E1.** SPEC ID regex: `[[ "SPEC-FACTORY-CODEX-LEADER-RESUME-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` printed `PASS` at authoring.
- **E2.** spec.md carries the 12 canonical frontmatter fields. plan.md and acceptance.md carry no `status:` field.
- **E3.** The requirements use no `IF` / `THEN` modality (lint `LegacyEARSKeyword`).
- **E4.** spec.md carries `### Out of Scope — <topic>` H3 headings, each with `-` bullets.
- **E5.** Requirement count 8 (ceiling 16). Acceptance count 9 (ceiling 16).
- **E6.** decision-index.md carries one `### Q<N>:` row per decision. Each `Operator verdict:` line is empty at authoring, and each FOUNDER row carries a `Class:` line.

## §F Milestones

Ordered by decision reversibility, so the decisions most likely to change come first. No milestone carries a time estimate.

Priority High — user-facing decisions

- **M0 — Decision Q1.** Output: the Q1 verdict. Reversible until the RED commit exists.
- **M1 — Decision Q2 and the signature choice (B7)**, only if Q1 is `reopen`. Output: the carrier wording for REQ-CLR-003, and the parameter or field choice.

Priority High — tests before behavior

- **M2 — RED commit.** Test-only, in its own commit. Contents:
  - Characterization tests, passing on the tree, for the shared combined gate across the cc-shaped inputs (B4), and for the automatic path's live-legacy-peer refusal (B4). Named `TestLeaderEntryRunAdoptionCombinedGate` and `TestAdoptResumedLeaderRunWarnsLiveLegacyPeer`.
  - The new Codex-leg tests of AC-001, AC-002, AC-003, AC-004, AC-005, and AC-008. Each fails on an assertion, not on a build error (AC-006(c)); AC-002 and AC-004 are regression guards and pass on the tree.
  - The mutant probe for AC-002 and AC-004: a deliberately over-adopting entry must fail those tests. The mutant is built and run locally and is not committed.
  - Output: the verbatim RED output and the verbatim baseline output, both recorded in progress §E.2.
- **M3 — GREEN commit.** Contents:
  - Read the prior run id before the publication (B5).
  - Call the shared combined gate with the Codex carrier and the explicit-run selector, and call the shared recorder with the Codex backend and the derived-capacity marker (B6).
  - Make the Codex-led refusal backend-aware (B3), so a Codex leader may re-adopt its own run while a cc or glm leader still refuses it.
  - Output: AC-001 through AC-008 pass, each with verbatim output.

Priority Medium — structure

- **M4 — REFACTOR commit.** Remove any argument plumbing that M3 duplicated, and keep behavior identical. Output: unchanged PASS output for AC-007 and the RED set.

Priority Low — outside this SPEC's required scope

- **M5 — Dead-branch removal (Q3).** Only by operator decision, in a separate SPEC or card.

## §G Anti-Patterns

- **AP-1.** A second adoption decision or a second candidate scan in a Codex file. That copy is forbidden.
- **AP-2.** A Codex-only resume judge that reimplements the Claude spelling logic, in place of the carrier decision Q2 makes.
- **AP-3.** Choosing the first of several candidates. The rule is warn-and-offer.
- **AP-4.** Adopting an indeterminate owner so that a Codex start appears to succeed.
- **AP-5.** Changing the REQ-SD-004 refusal without a recorded Q1 verdict.
- **AP-6.** Committing the RED tests and the implementation together, or reordering them afterwards.
- **AP-7.** Reporting a `go test` result without the verbatim output that carries the `--- PASS:` line.
- **AP-8.** Treating a characterization test that passes on the tree as evidence of new behavior. It proves only that the old behavior is kept.

## §H Cross-References

- spec.md §0 (blocking premise), §3 (requirements), §4 (out of scope)
- acceptance.md (AC-000 through AC-008, edge cases, known gaps)
- decision-index.md (Q1, Q2, Q3)
- `.moai/specs/SPEC-FACTORY-SELF-DISPATCH-001/spec.md` (REQ-SD-004; Out of Scope — leader-side surfaces)
- `.moai/specs/SPEC-CODEX-FACTORY-RETIRE-001/spec.md` (REQ-CFR-002, partially superseded)
- `.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/spec.md` (C-13, O-5) and its `progress.md` (Codex lane reachability)
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.3 (commit-graph ordering) and §3.1 (refused-tool Gaps)
- `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell adoption)
- `.claude/rules/moai/workflow/factory-dispatch-core.md` (leader and lane model)
- AGENTS.md §9 (hook event coverage)
