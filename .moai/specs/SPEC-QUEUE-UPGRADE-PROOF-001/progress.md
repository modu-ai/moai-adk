# Progress — SPEC-QUEUE-UPGRADE-PROOF-001

Card: `t470` · Branch `WT-queue-upgrade-proof` · Base `4e4607abe`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts written: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- Tier: M (4-file plan-phase set; no `design.md` / `research.md`)
- Status: `draft`
- Requirements: 10 (`REQ-QUP-001`..`010`), one optional (`REQ-QUP-007`)
- Acceptance criteria: 11 rows (`AC-QUP-001a`..`010`), one optional
  (`AC-QUP-007`); every row now names a requirement (no orphan)
- Open clarifications at v0.3.0: 2 — `[NEEDS CLARIFICATION: G2 definition]` and
  `[NEEDS CLARIFICATION: downgrade intent vs quarantine rename]`, both in
  `plan.md §A`; both RESOLVED at v0.4.0 (see the closing entry below)
- Plan audit: iteration 1 returned FAIL (score 0.875 vs Tier M threshold 0.80;
  cause was the MP-3 frontmatter defects and the MP-7 clarification gate, not
  the score). Verdict: `.moai/reports/t470/plan-audit.md`
- Remediation landed at SPEC `v0.2.0`: D1 (`tags` sequence → string), D2
  (`lifecycle` enum), D3 (`AC-QUP-010` mutation replaced with one that produces
  RED, plus a positive precondition and relocation sentinel on `AC-QUP-002`),
  D4 (`AC-QUP-008`'s gitignored `git status` limb replaced with a file digest
  comparison), D6 (`REQ-QUP-010` added; `AC-QUP-010` no longer an orphan), and
  the optional D7/D8/D9/D10. No production file touched — `REQ-QUP-009` holds
- Plan audit: iteration 2 returned FAIL (score 0.9625, monotonic up from 0.875;
  above the Tier M threshold 0.80). Cause was MP-7 alone. Verdict:
  `.moai/reports/t470/plan-audit-iter2.md`. Tier M iteration ceiling (2) reached
- Remediation landed at SPEC `v0.3.0`: D11 (`AC-QUP-008` + its twin constraint
  `C-1` named the live queue repository-relative, which from a linked worktree
  resolves to an absent file — both now derive the PRIMARY checkout's path the
  way `todo_root.go:95-99` does, and a failed derivation FAILS rather than
  passing) and the optional D12 (`AC-QUP-002`'s "holds the queue" limb given a
  stated observation). No production file touched — `REQ-QUP-009` holds
- Clarification gate CLOSED at SPEC `v0.4.0` (D5 resolved). Both markers in
  `plan.md §A` are converted to RESOLVED records — question retained, answer
  stated, source named (the dispatcher's ruling on card `t470`), consequence
  stated; neither marker was edited out. G2 is ABSORBED into G1 (carried by
  `AC-QUP-001a`/`001b`/`002`/`003`/`004`/`006`; the "closes as unstarted"
  contingency is withdrawn). The downgrade marker's earlier mechanism was WRONG
  and is corrected — the `.migrated` rename never contradicted the downgrade
  intent (`export-json` re-creates `backlog.json`); the real hole is that the
  export lands in the NEW directory while a v3.1.2 binary reads the legacy one,
  ruled OUT OF SCOPE as a separate-card candidate. G4 was newly supplied and is
  likewise OUT OF SCOPE, filed in `spec.md §E` beside G3 and G5. `AC-QUP-008`
  gained a hand-verification note (worktree guard refuses the nested `$(...)`
  form) with a matching pointer on its twin constraint `C-1`. **MP-7's blocking
  condition is now cleared.** No production file touched — `REQ-QUP-009` holds
- Open clarifications: 0 (was 2)

## §E.2 Run-phase Evidence

Deliverable: `internal/cli/todo_composed_upgrade_test.go` (one new test file, no
production file touched). Two tests:
`TestTodoComposedUpgrade_FromLegacyV312Layout` (M1 / G1) and
`TestTodoComposedUpgrade_ForwardCompatibleFieldsSurvive` (M2 / optional).

Full 5-section evidence, with every command's verbatim output, lives at
`.moai/reports/t470/verdict.md`. The matrix below is the AC roll-up.

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-QUP-001a | PASS | `go test ./internal/cli/ -run 'TestTodoComposedUpgrade' -count=1 -timeout 600s` | `ok github.com/modu-ai/moai-adk/internal/cli 3.068s` — the test decodes `todo list --json` and asserts ids `t2/t3/t5`, states `queued/picked/dropped`, texts, and seeded order |
| AC-QUP-001b | PASS | same run | post-upgrade `add` issued `t8` (seeded `last_seq` 7 + 1); a re-derived mark would have issued `t6` |
| AC-QUP-002 | PASS | same run | `todo/` exists; `todo/backlog.db` non-empty; `todo/companions.json` byte-identical to the seeded sentinel; `kanban/` gone. Preconditions asserted before the command: both legacy files present, `todo/` absent |
| AC-QUP-003 | PASS | same run | `todo/backlog.json.migrated` present, bytes identical to the F1 fixture; no `backlog.json` beside it |
| AC-QUP-004 | PASS | same run | `todo/backlog.db` exists, size > 0 |
| AC-QUP-005 | PASS | `go test ./internal/cli/ -run 'TestTodoComposedUpgrade' -v -timeout 600s` | both tests reported `=== RUN` + `--- PASS`; selector matched 2 tests (non-zero) |
| AC-QUP-006 | PASS | `sed -n '43,46p' internal/cli/todo_composed_upgrade_test.go \| grep -o '"\(version\|last_seq\|items\|findings\|archived\)"' \| sort -u` | `"items"` / `"last_seq"` / `"version"` — no `findings`, no `archived` |
| AC-QUP-007 | PASS (optional, attempted) | same targeted run | `TestTodoComposedUpgrade_ForwardCompatibleFieldsSurvive --- PASS`; 1 finding + 1 archived entry survive the composed upgrade |
| AC-QUP-008 | PASS on the controlled window; the run-wide window is a GAP | `shasum -a 256` + `stat -f '%m %z'` on the derived primary-checkout queue, before and after a window containing only this card's tests | before/after both `ecefa722…` / `1788422246 368640` — identical. The wider run-window comparison changed (`4b4656fd…` → `ecefa722…`) and is attributed to a foreign actor; see §E.3 and the verdict's Gaps section |
| AC-QUP-009 | PASS (baseline substituted) | `git diff --stat 6765a75c0..HEAD` | every changed path is a `_test.go`, a file under `.moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/`, or under `.moai/reports/t470/`. Substitution rationale in §E.3 |
| AC-QUP-010 | PASS | mutation applied, then reverted; both runs recorded | RED named AC-QUP-002 verbatim: sentinel absent under `todo/`, `kanban/` "must no longer exist … (stat err = <nil>)". Reverted GREEN recorded. Logs: `.moai/reports/t470/red-mutation.log`, `.moai/reports/t470/green.log` |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-03
run_commit_sha: d2675d57b
run_status: complete
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 0
production_files_changed: 0
new_warnings_or_lints_introduced: 0
total_run_phase_files: 1
m1_to_mN_commit_strategy: "M1 test commit + M3 evidence commit (2 commits)"
```

**AC-QUP-009 baseline substitution (recorded, not silent).** `acceptance.md`
names `git diff --stat 4e4607abe..HEAD`. That pin became an ANCESTOR of the
develop tip this branch absorbed (`6765a75c0`), so the named diff now spans 37
commits of unrelated develop work and no longer measures what this card
authored. The measurement was taken against `6765a75c0` instead — a pinned SHA,
not a moving ref — and the criterion's intent (no production file changed by
this card) is what was verified. The substitution changes the baseline, not the
assertion.

**AC-QUP-008 attribution.** The run-wide before/after comparison on the primary
checkout's `backlog.db` is NOT clean: the digest changed mid-run. It is
attributed to a foreign actor, not to this card's tests, on four observations —
`current-session-id.txt` names a different session; a session-start signature
(that file + `mcp-server/<pid>.json` + `lsel/clusters.json`) is stamped at
`1788421869`; `backlog.lock` carries the same mtime as the changed `backlog.db`
(`1788422246`), so a lock-taking write reached the live queue; and three other
lanes were running `go test ./internal/cli/...` concurrently. A controlled
window containing ONLY this card's tests left the file byte-identical. The
attribution is evidence-backed but not proof — recorded as a Gap in the verdict
rather than claimed as a clean PASS.

**Scope.** Verification scope was `./internal/kanban/...` + `./internal/cli/`
per C-2. No full local suite was run. No test spawns background load (C-4).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Decision: `serial` (one `manager-develop` spawn; `direct` / `fanout` / `sweep` all rejected — see the evaluation table below).

**Recorded late — stated rather than backdated.** The orchestration rule requires
this section be written BEFORE the first run-phase `Agent()` spawn. It was not;
it is written after the spawn returned. The record below is the decision that was
actually made at spawn time, and the ordering defect is recorded here rather than
concealed by silence.

**Input parameters**

| Parameter | Value |
|---|---|
| Tier | M |
| Scope (files) | 1 new `_test.go` + SPEC artifacts + evidence |
| Domain count | 1 (Go test authoring in `internal/cli/`) |
| File language mix | Go (test-only) + markdown |
| Concurrency benefit | LOW — single-file coding-heavy authoring |

**Mode evaluation**

| Mode | Selected | Rationale |
|---|---|---|
| `direct` | no | Not trivial: a composed-path proof plus a RED/GREEN mutation cycle is not a typo or single-line edit |
| `serial` | **yes** | Single domain, coding-heavy, one deliverable file — the default fallback, and the coding-task parallelism caveat points here |
| `fanout` | no | 1 domain, not ≥3; work is coding-heavy, not research-heavy |
| `sweep` | no | 1 file, not ≥~30; the work is semantic authoring, not a uniform mechanical transform |

**Decision: `serial`** — one `manager-develop` spawn (model `opus`, resolved via
`moai model profile --json` under the active `medium` profile).

**Justification.** The deliverable is a single Go test file whose fixture shape
and mutation seam are judgment calls, not mechanical edits. Fan-out would split
one coherent authoring decision across agents with nothing to gain, and the
coding-task parallelism caveat makes the sequential path the safe default for
coding work. Implementation Kickoff Approval was obtained before the spawn
(dispatcher's ruling on card `t470`, 2026-09-03, 4-card batch approval).

## §G Orchestrator Independent Verification

The lane orchestrator re-executed the deliverable rather than accepting the
implementer's report. Commands run in this worktree at HEAD `df2410ed5`:

| Check | Command | Observed |
|---|---|---|
| Non-production change | `git diff --name-only 6765a75c0..HEAD -- internal/ \| grep -v '_test\.go$' \| wc -l` | `0` |
| Changed-path classification | `git diff --stat 6765a75c0..HEAD` | 14 paths: 1 `_test.go`, 6 SPEC artifacts, 7 under `.moai/reports/t470/` |
| Test re-run (independent) | `go test ./internal/cli/ -run 'TestTodoComposedUpgrade' -count=1 -v -timeout 600s` | both tests `=== RUN` + `--- PASS`; `ok … 1.566s`; selector matched 2 (non-zero swept set) |
| Mutation seam disarmed | `grep -n 'seedLegacyV312Layout(t,\|assertPreUpgradeState(t,'` | 4 call sites, all passing `false` |
| F1 fixture fidelity | `grep -n '"version"\|"last_seq"\|"items"\|"findings"\|"archived"'` | F1 (`:43`) carries the three keys only; `findings`/`archived` appear solely in the F2 literal (`:283-286`) |
| Formatting | `gofmt -l internal/cli/todo_composed_upgrade_test.go` | no output |
| Verdict structure | `grep -n '^## ' .moai/reports/t470/verdict.md` | Claim / Evidence / Baseline-attribution / Gaps / Residual-risk — all five present |

**AC-QUP-008 corroboration (independent, orchestrator-side).** The live queue at
the derived primary checkout `/Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db`
measured `ecefa722b3b1181a8864e02ff0257301ed368d3f652b439f35b20521004d802f` /
`1788422246 368640` both immediately before and immediately after the independent
re-run above. A second controlled window, run by a different actor than the one
that authored the tests, therefore also left the file byte-identical. This does
not close the run-wide gap the verdict records — the mid-run change remains
attributed to a foreign actor by inference, not proof — but it is a second
independent observation on the same side.

**Residual defect the orchestrator is not closing.** The mutation seam
(`preCreateCurrentDir` / `currentDirPreCreated`) is committed code, `false` at
every call site, with no mechanical guard against one being flipped and left
flipped — which would silently disarm `AC-QUP-002`'s precondition. Recorded, not
repaired: repairing it would add production-shaped scaffolding to a card whose
`REQ-QUP-009` forbids exactly that.
