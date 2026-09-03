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

```yaml
sync_complete_at: 2026-09-03
sync_commit_sha: 305a39bd6
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-QUEUE-UPGRADE-PROOF-001' CHANGELOG.md -> 0 (no duplicate; emission proceeded)"
b12_self_test_b: "distinct AC identifiers in acceptance.md -> 11; CHANGELOG entry states 11 (match)"
b12_self_test_c: "every path named in the CHANGELOG entry verified present on disk (ls)"
changelog_entry_position: "[Unreleased] -> ### Added, first bullet"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed (carried on this single sync commit)"
  plan_md: "no status: field (artifact-stateless on the status axis)"
  acceptance_md: "no status: field (artifact-stateless on the status axis)"
  updated_field: "2026-09-03 (already current; unchanged)"
mx_tag_validation:
  scope: "internal/cli/todo_composed_upgrade_test.go (the single delivered file)"
  observed: "0 @MX: annotations; file declares only test functions, no exported production symbol"
  action: "none — no annotation added; the MX quality gate targets exported production functions, high fan-in symbols, and dangerous patterns, none of which this file carries"
docs_surfaces_checked:
  readme: "README.md / README.ko.md / README.ja.md / README.zh.md — no queue-layout or storage-path content; nothing to change"
  docs_site: "docs-site/content/{en,ko,ja,zh}/utility-commands/moai-todo.md describe the queue storage path and export-json; this card changes no behavior, so no edit"
production_files_changed_in_sync: 0
```

**CHANGELOG decision (stated, not silent).** An entry WAS written. The card is a
test-only regression guard with no behavior change, so the entry is not
automatic — but the project records exactly this class: the `[Unreleased]`
`### Added` section already carries `SPEC-LLMCFG-PRESERVE-001` ("test-only
close, zero production changes") and `SPEC-CODEX-SIDECAR-GUARD-001` (added
assertions only). What the entry tells a reader is not a new capability but a
newly-guarded one, plus the boundary of what the guard does NOT cover (G3 / G4 /
G5, and the downgrade-export directory hole) — information that exists nowhere
a release reader would find it if the entry were omitted.

**`sync_commit_sha` placeholder.** Written as the canonical `pending-backfill`
in the sync commit itself and backfilled to `305a39bd6` in this following
commit. A commit cannot
cite its own hash; leaving the slot EMPTY is not the alternative, because the
SPEC is `completed` once this commit lands and nothing would ever schedule the
repair.

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

## §H Post-sync Corrections and Attributions

Three records the sync commit could not carry, because each arrived after it.

### AC-QUP-008 — the foreign actor is now named, and my inference was wrong

The run-wide live-queue change (`4b4656fd…` → `ecefa722…`, mtime `1788421830` →
`1788422246`) is attributed to the **`lead-1` session**, which stated it performed
three writes to the primary checkout's live queue during the run window:
`moai todo done t278`, `moai todo edit t446`, `moai todo edit t472`.

**Disposition: attribution recorded, the Gap RETAINED.** A statement of authorship
is not byte-level causal proof, and the lane declined to promote it by reproducing
the causation in a controlled window: this card's judgment is that the upgrade path
changed no production file, not that the queue is immutable, and a reproduction
would widen the card's scope for a fact it does not need.

**A correction against this lane's own earlier reasoning.** The run-phase report
narrowed the candidate set to the three lanes observed running
`go test ./internal/cli/...` concurrently (`t446`, `t454`, `t410`). That inference
was WRONG — the writer was the lead, which the process listing could not have
revealed, because the lead was not running the tests the listing was filtered on.

The distinction matters and is recorded rather than smoothed over: **recording the
verdict as a Gap was correct; the circumstantial reasoning inside it was not.** The
two are separate claims, and the first being right did not make the second right.
Had the run-wide window been written up as a PASS on the strength of that
plausible-looking inference, the lead's own writes would never have surfaced —
the restraint, not the reasoning, is what preserved the correction.

### Shared-rule defect — the B12 acceptance-criterion counter under-counts

Found by `manager-docs` during the CHANGELOG AC-count self-test, then reproduced
independently by this lane rather than taken on report. The canonical counter in
`.claude/rules/moai/development/manager-develop-prompt-template.md` § B12 collapses
sub-lettered criteria into a single token:

| Pattern | Command | Observed |
|---|---|---|
| B12 canonical | `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md \| sort -u \| wc -l` | `10` |
| With trailing `[a-z]?` | `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+[a-z]?' acceptance.md \| sort -u \| wc -l` | `11` |

`AC-QUP-001a` and `AC-QUP-001b` both reduce to `AC-QUP-001`. The CHANGELOG entry
states 11, the measured truth. **Not repaired here** — the defect is in a shared
rule, not in this SPEC, and repairing it from inside a proof card would be the
scope creep `REQ-QUP-009` and `spec.md §E` exist to prevent. Relayed to the lead
as a card candidate.

### Docs surfaces — independently swept, and not lagging

`manager-docs` reported README and docs-site as requiring no change. This lane
swept the same surfaces separately rather than accepting the claim:

| Check | Command | Observed |
|---|---|---|
| SPEC referenced in README (4 locales) | `grep -rln 'SPEC-QUEUE-UPGRADE-PROOF-001\|state/kanban' README*.md \| wc -l` | `0` |
| SPEC referenced in docs-site | `grep -rln 'SPEC-QUEUE-UPGRADE-PROOF-001' docs-site/content/ \| wc -l` | `0` |
| Queue path documented | per-file `grep -o 'state/kanban\|state/todo'` over the 8 matching pages | `state/todo` on all 8; `state/kanban` on none |

The eight pages are 4 locales × (`utility-commands/moai-todo.md`,
`advanced/moai-web-console.md`), and every one documents the CURRENT path only.
So "no edit" is right for a stronger reason than "nothing changed": the docs are
not lagging the migration this card proves, and an edit would have introduced
drift rather than removed it.

The `export-json` row's silence about the downgrade export landing in the NEW
directory is left uncorrected on purpose — that belongs to the separate-card
finding recorded in `spec.md §E`, and reaching it from here would widen this card.

## §I Sync-audit Verdict and the Corrections It Forced

`sync-auditor`, adversarial stance, verdict at `.moai/reports/t470/sync-audit.md`.

**PASS-WITH-DEBT — 91.6** (weighted harmonic mean, Tier M): Functionality 92
(40%), Security 96 (25%), Craft 84 (20%), Consistency 95 (15%). Both must-pass
dimensions cleared independently.

The auditor re-ran the RED mutation itself rather than reading `red-mutation.log`
and observed the same four failures at the same line numbers (236/236/236/249),
including the symptom `AC-QUP-010` names by name. It also independently
re-measured the controlled live-queue window — a THIRD clean observation.

### F1 (blocking) — a committed verdict sentence that was false

`verdict.md §E6` asserted every changed path was a `_test.go` file, a SPEC
artifact, or a report. At final HEAD that is FALSE: `CHANGELOG.md`, added by the
sync commit `305a39bd6`, sits outside the `acceptance.md` allowlist. Worse, the
diff block quoted beneath the claim was measured BEFORE the M1 commit, so it did
not even contain `internal/cli/todo_composed_upgrade_test.go` — **the claim
outran its own evidence.**

Verified independently by this lane before accepting the finding:

```
$ git diff --name-only 6765a75c0..HEAD | grep -v '_test\.go$' \
    | grep -v '^\.moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/' \
    | grep -v '^\.moai/reports/t470/'
CHANGELOG.md
```

**Corrected in place at `verdict.md §E6`, with the defect recorded rather than
quietly overwritten.** The requirement the criterion serves is separately intact:
`git diff --name-only 6765a75c0..HEAD -- internal/ pkg/ cmd/ | grep -v '_test\.go$' | wc -l`
returns `0`. What was wrong is the RECORD, not the work — and `AC-QUP-009`'s
allowlist, written in plan-phase against a run-phase diff, simply did not
anticipate a sync-phase artifact.

### A correction against this lane's own reasoning (second one this card)

`§G` justified leaving the mutation seam unguarded by saying repairing it "would
add production-shaped scaffolding to a card whose `REQ-QUP-009` forbids exactly
that." **That rationale is WRONG and the auditor was right to reject it.** Both
parameters, and any guard over them, live inside `_test.go`; `REQ-QUP-009`
governs production behavior and reaches none of it. The honest reason to leave
the seam as delivered is narrower and sufficient: reproducing the RED costs two
literal flips, so the seam earns its keep without a guard.

The earlier wording is left standing in `§G` rather than edited out, so the
correction is visible as a correction.

### The seam's real hazard is asymmetric, which the earlier record missed

The two parameters do not fail the same way, and only one of them is dangerous:

| Flipped alone | Result |
|---|---|
| `preCreateCurrentDir` | **Loud RED** — self-exposing, caught immediately |
| `currentDirPreCreated` | **Silent GREEN** — the early return at `:115-117` disarms `AC-QUP-002`'s precondition while the suite stays green |

Nothing catches the second: no observable change in the committed configuration,
so review, CI, and `unparam` all pass it (the auditor ran lint to confirm).
Classified optional because the harm is latent, not active — but the earlier
"either flag" framing understated it, and the asymmetry is the part worth
carrying forward.

### Auditor's own stated gaps, carried not resolved

- It did NOT re-run the full `./internal/cli/` package (~935s, past the Bash
  ceiling). **The package-level verdict in every report here rests on the
  implementer's `cli_suite.log`, not on the auditor's own measurement.** It ran
  targeted selectors twice plus `-race`, and `./internal/kanban/...` in full
  (146.7s).
- No `GOOS=windows` build, no coverage (zero production lines to attribute), no
  causal reproduction of the AC-QUP-008 run-wide change, no plan-phase audit.
- Its RED used the SAME mutation as the implementer's. It did not systematically
  hunt a SECOND mutation that ought to go red and does not — the one it caught by
  eye is that `AC-QUP-004` passes under the mutation and duplicates `AC-QUP-002`'s
  check.

### Residual the auditor flagged about its own finding

A one-line text repair is the kind that gets deferred and forgotten. Once the
card reaches `done` nothing schedules it, and the false sentence outlives the
card. That is why it was classified blocking despite being one line — and it is
why the repair landed in this same commit rather than being queued.
