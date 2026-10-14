# SPEC-USER-SETTINGS-PROTECT-001 — Progress Record

Card t1630 (backlog 3.2-0-1, priority P0, change class C). Lane lane-18, run tmnboq, branch WT-3-2-0, worktree `.moai/worktrees/t1630`. Tier M (lane decision, §G; spec.md §0). Artifact set: spec.md, plan.md, acceptance.md, decision-index.md (required by the decision gate), and this progress record.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-10T07:02:43Z (first authoring)
- plan_tree: revision 2 is written on top of HEAD 5dc6c4530 (short SHA measured in the repair round). The RED observations stay pinned at 3975fe3cc. The lane commits revision 2, and the re-audit runs on that commit.
- revision: revision 2 is the repair round after the revision 1 iteration 2 FAIL (aggregate 0.69; audited_sha 5dc6c45307d21e4b0d34c876631e2fa553fde57b). It applies the operator-delegated split under ruling d-20261010T091713Z-10fe (decision board, card t1630). Policy decisions Q1, Q2, and Q5 stay in scope with the text and evidence defects B1, B4, B6 (wording), B8, B9, B11, and B12. Moved to t1666: the test-isolation items, the clean --home run and db items, decision Q3, decision Q7, the ~/.moai/run growth-0 criterion, and the post-t1619 baseline. Audit defects B2 (the AC-009 run path), B3 (AC-007), B5 (the AC-006 cell 1 order), B7 (the REQ-011 and REQ-012 deletion clauses), and B10 (the AC-006 run-time verdict) move with their criteria.
- audit-ready signal: granted by the delta plan audit PASS on commit fb672d1723c485ac52d4b21f8e4f3bbceafc78c9 (.moai/reports/t1630/plan-audit-iter3.md). The withdrawal ended with that re-audit. The ceiling count restarts at this revision.
- counts (revision 2): in-scope normative requirements 4 (REQ-001, REQ-002, REQ-005, and REQ-013; REQ-003 and REQ-004 sit in the spec.md §2 note); in-scope acceptance criteria 6 (AC-001 to AC-005 and AC-013); both within the Tier M ceiling of 16. Moved and retained: REQ-006 to REQ-012 and AC-006 to AC-012 (7 of each).
- evidence_manifest_sha256: 628c676714fe4b3c040f7a6a14d84618e79d9e3b0bcd3e1f57e7ffca74fd1ba8 (`shasum -a 256 .moai/specs/SPEC-USER-SETTINGS-PROTECT-001/evidence-manifest.json`, observed in revision 2). The last commit that changed the manifest is 5dc6c4530, and revision 2 does not change it. The binding is by manifest hash; the originals stay local under `.moai/reports/t1630/` (plan.md Appendix A). The plan-artifact hash (the ComputeHash subject, internal/runtime/audit_cache.go:92-100) covers acceptance.md, decision-index.md, design.md, plan.md, research.md, spec.md, and tasks.md, and excludes progress.md and evidence-manifest.json.
- independent plan audit: revision 2 iteration 2 (.moai/reports/t1630/plan-audit-iter2.md) FAIL at 56bd364f6c821999d0de3a6ecf226454198ab251. The text repair of B1 to B3 under ruling d-20261010T110028Z-358a is commit fb672d1723c485ac52d4b21f8e4f3bbceafc78c9. The delta re-audit (.moai/reports/t1630/plan-audit-iter3.md) returned PASS at fb672d172, score 0.81, blocking 0. Its residual-risk section makes the PASS conditional on operator ratification of the non-normative scope of REQ-003 and REQ-004.
- decisions: Q8 DECIDED (anchor SPEC-INIT-WIZARD-REPAIR-001, HISTORY row 0.1.1). Q1, Q2, and Q5 DECIDED by the pinned board record `board:d-20261010T073547Z-07ae#9c93e47809f9` (decided_by 영실이 판단, operator-delegated). Q4 and Q6 EVIDENCE-NEEDED, closed at M2 without blocking run. Q3 and Q7 are Out of Scope - moved to t1666; their rows keep their text.
- Q7 and the Kickoff: Q7 (FOUNDER, implementation-level) has no subject once the sandbox items move, so it no longer gates this SPEC's Kickoff. The leader confirms this reading at re-audit.
- no default is applied at plan close.
- ordering gate REQ-013: open at plan close (AC-013 red, plan.md E-14); the run-phase gate is P-2.
- EC-1 to EC-4: observed in revision 2 by the EC observation command in acceptance.md (exit code 0 at HEAD 5dc6c4530).
- AC-003 part A RED: re-observed at HEAD 5dc6c4530 in revision 2 (verbatim in acceptance.md; exit code 1).
- open: G-18 (the digest covers the header copy only; disclosed).
- tier: the in-scope set is four files (plan.md §A.4 Basis), below the Tier M file band of 5 to 15. The tier is not re-decided in revision 2 (see §G, Tier decision).
- scope boundary: scope item S-7 (t1594) is blocked on card text (plan.md G-2). No requirement is authored for it in this revision.
- process note: HEAD moved during this repair round, from 5dc6c4530 to 169a780db (commit "docs(SPEC-USER-SETTINGS-PROTECT-001): revision 2 after split (card t1630)", author Goos Kim, 2026-10-10 18:52:13 +0900, which holds the revision-2 state as of the AC-013 header edit). The repair-round edits are uncommitted in the working tree; this pass did not commit or stage. Reported in the repair-round reply.

## §E.2 Run-phase Evidence

Run phase opened by manager-develop on factory lane-17, card t1630. Run base: HEAD ef206ba8c6 on WT-3-2-0 (the merge commit that absorbed local main ec193e4777 — 689 commits of post-plan mainline work). Toolchain go1.26.8 darwin/arm64. All commands below ran in this run, against this tree, with `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED` in the same invocation.

### Plan Audit Gate skip (three conditions, each observed in this run)

1. Verdict PASS: the delta re-audit (.moai/reports/t1630/plan-audit-iter3.md) returned PASS at fb672d1723, score 0.81, blocking 0.
2. Score 0.81 >= the Tier M threshold 0.80.
3. Artifact-hash unchanged since the verdict: `git diff cd93616f99..ef206ba8c6 -- .moai/specs/SPEC-USER-SETTINGS-PROTECT-001/` prints nothing (0 lines, re-run by the lane in this run; the ComputeHash subject excludes progress.md and evidence-manifest.json, per plan.md Appendix A). The uncommitted working-tree diffs touch progress.md (leader label corrections + §F) and spec.md (HISTORY label correction + the draft→in-progress frontmatter transition of this commit) only; plan.md and acceptance.md carry no uncommitted diff (`git diff --stat` over both: empty).

### P-3 — RED re-observation on the post-absorb run base (all six in-scope cells re-observed at ef206ba8c6; every cell failed with the same shape as the pinned 3975fe3cc observation)

- AC-001: `go test -overlay .moai/reports/t1630/evidence/overlay-ac001-003.json -count=1 -v -run '^TestAC001PreservesListsAndUnmodelledKeys$' ./internal/config/toolpolicy/` → allow/ask/deny each logged "was removed"; `--- FAIL: TestAC001PreservesListsAndUnmodelledKeys`; exit 1.
- AC-002: same overlay, `-run '^TestAC002RegionBytesUnchangedWhenDefaultModeMatches$'` → "file bytes changed" with the before/after bodies; `--- FAIL`; exit 1.
- AC-003 part A: same overlay, `-run '^TestAC003ListsPreservedUnderEitherDisposition$'` → both subtests `differing` and `matching` FAIL with the three "was removed" logs each; exit 1; executed 3.
- AC-003 part B: `go test -overlay .moai/reports/t1630/evidence/overlay-ac003b.json -count=1 -v -run '^TestAC003DefaultModeKeptOnBothPaths$' ./internal/core/project/` → `default-init`: "plan" overwritten with "acceptEdits"; `automatic`: "plan" overwritten with "auto"; exit 1; executed 3.
- AC-004: `grep -c '"defaultMode": *"default"' internal/template/templates/.claude/settings.json.tmpl` → `0`, exit 1.
- AC-005 case (i): `go test -overlay .moai/reports/t1630/evidence/overlay-ac005.json ... '^TestAC005PolicyPathKeepsUserAllowEntry$' ./internal/core/project/` → "user-added permissions.allow entry dropped"; exit 1.
- AC-005 case (ii): `go test -overlay .moai/reports/t1630/evidence/overlay-ac005b.json ... '^TestAC005WithRecordKeepsUserAddedEntry$' ./internal/core/project/` → "user-added permissions.allow entry dropped with a sidecar record present"; exit 1.

Evidence-manifest integrity at entry: `jq -r '.files[] | "\(.sha256)  \(.path)"' .moai/specs/SPEC-USER-SETTINGS-PROTECT-001/evidence-manifest.json | shasum -a 256 -c -` → 37 lines OK, 0 non-OK.

### AC-013 ordering gate — measured divergence from the dispatch text (named, not worked around)

Measured at ef206ba8c6 with the plan's representative landing SHAs (plan.md E-14):

```
git merge-base --is-ancestor 8108eb256 HEAD   # t1619 → exit 1 (NOT an ancestor)
git merge-base --is-ancestor e73a7cbf5 HEAD   # t1578 → exit 0 (ancestor)
git merge-base --is-ancestor a372a984c HEAD   # t1591 → exit 1 (NOT an ancestor)
```

Corroborating probes: no commit naming t1619/t1591 or SPEC-UPDATE-REPAIR-RC30-001/SPEC-USERASSET-DEPLOY-GUARD-001 exists in HEAD ancestry or in the absorbed main tip ec193e4777's history; the SPEC directories `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/` and `.moai/specs/SPEC-USERASSET-DEPLOY-GUARD-001/` are absent from HEAD; t1591's branch lands as PR #1818 (`gh pr view 1818` → state OPEN, mergedAt null). The dispatch statement "predecessor t1619 landed (inside the absorbed mainline)" is therefore not reproducible on this tree for the SPEC-artifact or ancestry identity.

Gate-purpose containment measurement (why the run proceeded): `MB=$(git merge-base 8108eb256 HEAD)` → 32035f28b; `git diff $MB 8108eb256 --stat -- internal/config/toolpolicy internal/core/project internal/template/templates` → only `templates/.moai/config/sections/protected-zone.yaml` (+8), zero overlap with this SPEC's target files (the t1619 branch delta touches internal/cli/init.go +89/-13 and 103 files total, none of them this SPEC's writer/template paths; no in-scope AC measures the init executor — plan.md R-1 carries that level as a Gap). The t1619/t1591 landing is cross-card work the lane cannot land itself; the leader owns its disposition (merge before this card's batch, or re-pin the gate).

### E1 — AC matrix (release-blocking in-scope set green; verbatim output from this run, tree e581bdbb20)

| AC | Status | Verification command | Actual output (verbatim deciding lines) |
|----|--------|---------------------|------------------------------------------|
| AC-001 | PASS | `go test -count=1 -v -run '^(TestAC001PreservesListsAndUnmodelledKeys)$' ./internal/config/toolpolicy/` (repository-authored; same assertion as the overlay probe) | `--- PASS: TestAC001PreservesListsAndUnmodelledKeys (0.01s)` + 4 EC subtests RUN (N=5 executed) |
| AC-002 | PASS | `go test -count=1 -v -run '^(TestAC002RegionBytesUnchangedWhenDefaultModeMatches)$' ./internal/config/toolpolicy/` | `--- PASS: TestAC002RegionBytesUnchangedWhenDefaultModeMatches (0.00s)` (N=1) |
| AC-003 part A | PASS | `go test -count=1 -v -run '^(TestAC003ListsPreservedUnderEitherDisposition)$' ./internal/config/toolpolicy/` | `--- PASS: TestAC003ListsPreservedUnderEitherDisposition (0.00s)` with `differing` + `matching` RUN (N=3) |
| AC-003 part B | PASS | `go test -count=1 -v -run '^(TestAC003DefaultModeKeptOnBothPaths)$' ./internal/core/project/` | `--- PASS: TestAC003DefaultModeKeptOnBothPaths (0.00s)` with `default-init` + `automatic` RUN (N=3) |
| AC-004 | PASS | `grep -c '"defaultMode": *"default"' internal/template/templates/.claude/settings.json.tmpl` | `1`; total `"defaultMode"` occurrences also `1` (N=1) |
| AC-005 case (i) | PASS | `go test -count=1 -v -run '^(TestAC005PolicyPathKeepsUserAllowEntry)$' ./internal/core/project/` | `--- PASS: TestAC005PolicyPathKeepsUserAllowEntry (0.00s)` (N=1) |
| AC-005 case (ii) | PASS | `go test -count=1 -v -run '^(TestAC005WithRecordKeepsUserAddedEntry)$' ./internal/core/project/` | `--- PASS: TestAC005WithRecordKeepsUserAddedEntry (0.00s)`; both assertions held (user-added kept, old-managed absent) (N=1) |
| AC-013 (gate) | FAIL (measured divergence, leader disposition pending) | `git merge-base --is-ancestor <sha> HEAD` per §E.2 gate block | t1619 8108eb256 exit 1; t1578 e73a7cbf5 exit 0; t1591 a372a984c exit 1 — see §E.2 for the containment measurement and the dispatch-vs-tree divergence |

Repository test names recorded per acceptance.md QG-4: `TestAC001PreservesListsAndUnmodelledKeys` (+ subtests `ec1-absent-file`, `ec2-no-permissions-key`, `ec3-invalid-json`, `ec4-duplicate-allow-entries`), `TestAC002RegionBytesUnchangedWhenDefaultModeMatches`, `TestAC003ListsPreservedUnderEitherDisposition` (+ `differing`, `matching`), `TestAC003DefaultModeKeptOnBothPaths` (+ `default-init`, `automatic`), `TestAC005PolicyPathKeepsUserAllowEntry`, `TestAC005WithRecordKeepsUserAddedEntry`.

Invariant rows (spec.md §4): C1 no probe wrote outside `.moai/reports/t1630/evidence/` and `t.TempDir()` roots (overlay probes only); C2 the writer targets only its settings file through the existing atomicfile path (unchanged); C4 every change landed test-first (RED observed in-repo before GREEN — `go test -count=1 -v -run '^(TestAC001...|TestAC002...|TestAC003Lists...)$'` FAILED at lines 257/260/263/334/361/403/406/409 and autonomy_bundle_test.go 332/391/441 before the implementation commit; verbatim RED bodies in the P-3 block above and in the overlay-cell outputs); C5 gate semantics untouched (`EffectiveTierWithGates`/`SandboxProofKind` unmodified — the five pre-existing autonomy_bundle tests pass unchanged); C6 every claim above carries its command and verbatim output.

### E2 — cross-platform build (tree e581bdbb20)

```
$ go build ./...                           → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

### E3 — coverage (affected packages, this run)

```
$ go test -cover ./internal/config/toolpolicy/ ./internal/core/project/
ok  github.com/modu-ai/moai-adk/internal/config/toolpolicy  0.738s  coverage: 88.6% of statements
ok  github.com/modu-ai/moai-adk/internal/core/project       2.808s  coverage: 89.0% of statements
```

Both above the 85% threshold.

### E4 — subagent-boundary grep (C-HRA-008 family)

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/config/toolpolicy internal/core/project | grep -v _test.go | grep -v '// '
(no output; grep exit 1 = zero matches)
```

### E5 — lint status

```
$ golangci-lint run --timeout=2m
0 issues.
```

Pre-existing baseline on this tree: 0 issues; NEW issues introduced by this run: 0.

### E6 — regression families and known baseline failure

- `go test -count=1 ./internal/config/toolpolicy/... ./internal/core/project/...` → both `ok` (post-GREEN, twice).
- `go test -count=1 ./internal/template/...` → `internal/template` FAIL, exactly one failing test: `TestRuleDateProvenance` — a date-provenance leak (`2026-10-09`, line 368) in `internal/template/templates/.claude/rules/moai/core/hooks-system.md`. PRE-EXISTING on this tree (B5 classification): that file was last touched by absorbed-mainline commit ec193e4777 ("rule + template" row edit), is not in this run's file set, and the test is a static scan of files this run does not modify. All settings/template tests this run CAN affect pass (`TestSettingsTemplatePermissions`, `TestTemplate_NoBypassPermissionsDefault`, `TestTemplate_NoFullyAutonomousPreSelection`, `TestSettingsTemplatePowerShellDenyParity` darwin/linux/windows, embedemit/agentemit/commandemit subpackages all ok). Remediation (removing the work-date string from the template rule row) is mainline work outside this SPEC's envelope — left to the leader. A first family run had also failed at exactly 600.5s (go test's default package timeout) under machine load; the 900s-timeout verbose re-run completed with the single failure above.
- `internal/cli` is not run as a suite (structurally red in card worktrees, per the t1542 measurement; CI is the full verdict). This run's launcher.go change is comment-only; both builds compile it, and no cli test reads the template content (verified by grep).

### E7 — commit list (no push performed)

- 057db17c41 `feat(SPEC-USER-SETTINGS-PROTECT-001): M1 run-phase entry, draft to in-progress (card t1630)`
- e581bdbb20 `feat(SPEC-USER-SETTINGS-PROTECT-001): M2+M3 writer preservation, policy-path union, template defaultMode (card t1630)`

No push, no PR, no force, no --amend, no --no-verify (git-flow lane protocol B9: the leader batches the push).

### Init-level end-to-end observation (the card completion criterion; plan.md R-1 gap carried to the run phase)

Binary built from this tree and invoked by path (verification-claim-integrity §2.2): `go build -o /tmp/t1630-moai-bin/moai ./cmd/moai` → exit 0. Three scenarios, each `HOME=<sandbox> MOAI_HOME=<sandbox> /tmp/t1630-moai-bin/moai init --non-interactive <tmp-project> </dev/null` (init exit 0 each):

1. Matching defaultMode (`acceptEdits`) + allow/ask/deny + sibling env, non-canonical layout: **permissions content diff is 0** — after holds the same allow/ask/deny/defaultMode values. The whole-file byte diff is NOT 0: `ensureGlobalSettingsEnv` (internal/cli/update.go:1057, called from init.go:1067 — a pre-existing writer outside this SPEC's REQ set) re-marshal-parses the file, adds `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`, `CLAUDE_DISABLE_PATH_WARNING`, and `PATH` to `env` (the user's own `env` keys survive), and the map re-marshal normalizes key order alphabetically. The toolpolicy codegen is not involved in that rewrite (its canonical order is defaultMode-first; the observed order is alphabetical).
2. Differing defaultMode (`"plan"`) + allow: **kept** — after init `defaultMode: plan`, `allow: ['Bash(make:*)']` (Q1 holds through the full init executor).
3. Fresh home (no user settings file): created with `defaultMode: acceptEdits`, no lists invented (EC-1 holds end-to-end).

Card completion criterion disposition: `init 전후 settings diff 0` is met at the content level the criterion's own parenthetical defines ("init preserves user permissions and an existing defaultMode") — both hold in every scenario. The strict byte-level no-diff is not met, and the cause is the pre-existing `ensureGlobalSettingsEnv` whole-file env merge, not the writer this SPEC fixes. Making that writer byte-preserving is un-decided surface outside this SPEC's REQ set; named for the leader to disposition (accept content-level, or cut a follow-up card for the env writer).

### Card-review repair (P1) — last_generated sidecar record had no writer

**Finding** (card-review P1, verified by the lane): `internal/config/toolpolicy/tier_render.go` read `<projectRoot>/.moai/state/tool-policy/managed-allow.json` for the Q5 union classification but nothing in the repository wrote it — `last_generated` was never recorded, so an allow entry later REMOVED from the policy document was misclassified as user-added and persisted forever (revoked permission resurrection). This was the missing write half of the already-decided Q5 union rule (AC-005), not new behavior.

**Disposition: fixed** (commit dd1285d066). A successful PROJECT-scope render now persists the generated managed allow list (`full.Allow` at that render) as `last_generated` into the sidecar record — creating directory and file when absent, overwriting when present, via `atomicfile.Write` (the same choke point the settings render uses). The write runs only after the PROJECT-scope `renderIntoFile` succeeds, so a failed or skipped render never touches the record. The read path and union classification are unchanged in intent; the read side now shares one `managedAllowRecordPath` derivation with the writer.

**RED evidence** (pre-fix run, this tree, `go test ./internal/config/toolpolicy/ -run 'TestRenderTierPermissions_PersistsManagedAllowRecord|TestRenderTierPermissions_RemovedGeneratedEntryIsDropped|TestRenderTierPermissions_FailedRenderLeavesRecordUntouched' -v`):

```
=== RUN   TestRenderTierPermissions_PersistsManagedAllowRecord
    tier_render_test.go:259: managed-allow record not written: open /var/folders/.../proj/.moai/state/tool-policy/managed-allow.json: no such file or directory
--- FAIL: TestRenderTierPermissions_PersistsManagedAllowRecord (0.00s)
=== RUN   TestRenderTierPermissions_RemovedGeneratedEntryIsDropped
    tier_render_test.go:302: removed generated entry not dropped: allow = [Bash(managed-a:*) Bash(managed-b:*)], want [Bash(managed-a:*)]
--- FAIL: TestRenderTierPermissions_RemovedGeneratedEntryIsDropped (0.00s)
--- PASS: TestRenderTierPermissions_FailedRenderLeavesRecordUntouched (0.00s)
FAIL    github.com/modu-ai/moai-adk/internal/config/toolpolicy  0.282s
```

The deciding cell is the second FAIL: with no writer, `Bash(managed-b:*)` — generated by the first render and removed from the document before the second — survived the second render, reproducing the exact resurrection defect. The third test passed pre-fix by construction (nothing wrote the record, so a failed render trivially left it untouched) and is retained as a regression guard that must KEEP passing now that the writer exists.

**GREEN evidence** (post-fix run, same selector):

```
--- PASS: TestRenderTierPermissions_PersistsManagedAllowRecord (0.01s)
--- PASS: TestRenderTierPermissions_RemovedGeneratedEntryIsDropped (0.01s)
--- PASS: TestRenderTierPermissions_FailedRenderLeavesRecordUntouched (0.01s)
PASS    ok      github.com/modu-ai/moai-adk/internal/config/toolpolicy  0.298s
```

**Post-fix verification** (all env-scrubbed `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED`, this run, this tree): `go test ./internal/config/toolpolicy/ ./internal/core/project/` → both `ok`; `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0; `go test ./internal/config/toolpolicy/ -cover` → `coverage: 90.4% of statements` (threshold ≥ 85%); `gofmt -l internal/config/toolpolicy/` → clean. The existing AC-005 union test in `internal/core/project/autonomy_bundle_test.go` (`TestAC005WithRecordKeepsUserAddedEntry`, which pre-seeds the record) passes unmodified — the classification reads the previous render's record before the write overwrites it.

### Card-review repair round 2 (P2-1 + P2-2) — record-maintenance completeness

**Findings** (card-review round 2, both verified by the lane as completeness gaps of the record-maintenance rule):

- **P2-1 failure-ordering**: the round-1 writer persisted the record AFTER the PROJECT settings write. A record save failing after a successful settings write left new managed entries in settings but not in `last_generated` — a later policy removal could not drop them (the same resurrection hazard class this SPEC fights).
- **P2-2 BuildInto path**: the `moai tool-policy build` local target (`BuildInto`/`BuildIntoAuto`, `TargetJSON` — `internal/cli/tool_policy.go:119`) generated the same project allow list but never refreshed the record, so entries entering settings through that path were equally unrecorded.

**Disposition: fixed** (commit 026ba839a), all inside `internal/config/toolpolicy`:

- **Stage/converge protocol** (P2-1): the render now stages `prior ∪ generated` into the record BEFORE the PROJECT settings mutation (`stageManagedAllowRecord`), and converges the record to exactly the generation after the settings write lands (`persistManagedAllowRecord`). Failure directions verified: a failed STAGE errors out before any settings mutation (settings untouched); a failed SETTINGS write after a successful stage leaves the staged union — the safe direction, since a superset can only mark more entries as managed, and the next successful render then re-adds still-generated entries as managed or drops the revoked one (the deciding test observes exactly this recovery).
- **BuildInto wiring** (P2-2): the same stage/converge protocol is wired into `BuildInto` for the `TargetJSON` target (the project settings file — the only production JSON target of the codegen); the `TargetTemplate` target is the distribution template, not a project settings file, and maintains no record.
- Round-1's failure-path test was REVISED to the new contract: a failed render may advance the record to the staged union (safe); the protected invariants are now "PROJECT settings untouched" + "record in the safe staged state". The round-1 P1 deciding test (`RemovedGeneratedEntryIsDropped`), the writer test, and both existing AC-005 tests in `internal/core/project` pass unmodified under the new semantics.

**RED evidence** (pre-fix run, this tree, selector `-run 'TestRenderTierPermissions_RecordStageFailureLeavesProjectSettingsUntouched|TestRenderTierPermissions_SettingsWriteFailureAfterRecordStage_ClassifiesCorrectlyNextRender|TestBuildIntoPersistsManagedAllowRecord_RemovedEntryNotKeptAsUserAdded|TestRenderTierPermissions_FailedProjectRenderLeavesSettingsUntouchedRecordSafe' -v`):

```
=== RUN   TestRenderTierPermissions_FailedProjectRenderLeavesSettingsUntouchedRecordSafe
    tier_render_test.go:349: record after failed render = [Bash(sentinel:*)], want the safe staged union [Bash(managed-a:*) Bash(sentinel:*)]
    tier_render_test.go:371: managed-allow record not written: open .../fresh/.moai/state/tool-policy/managed-allow.json: no such file or directory
--- FAIL: TestRenderTierPermissions_FailedProjectRenderLeavesSettingsUntouchedRecordSafe (0.00s)
=== RUN   TestRenderTierPermissions_RecordStageFailureLeavesProjectSettingsUntouched
    tier_render_test.go:416: PROJECT settings mutated although the record stage failed:
        before={"permissions": {"defaultMode": "default"}}
        after={"permissions": {"allow": ["Bash(managed-a:*)"], "ask": [...], "deny": [...]}}
--- FAIL: TestRenderTierPermissions_RecordStageFailureLeavesProjectSettingsUntouched (0.01s)
=== RUN   TestRenderTierPermissions_SettingsWriteFailureAfterRecordStage_ClassifiesCorrectlyNextRender
    tier_render_test.go:460: record after failed settings write = [Bash(managed-a:*) Bash(managed-b:*)], want staged union [Bash(managed-a:*) Bash(managed-b:*) Bash(managed-c:*)]
--- FAIL: TestRenderTierPermissions_SettingsWriteFailureAfterRecordStage_ClassifiesCorrectlyNextRender (0.01s)
=== RUN   TestBuildIntoPersistsManagedAllowRecord_RemovedEntryNotKeptAsUserAdded
    tier_render_test.go:508: BuildInto did not maintain the managed-allow record: open .../.moai/state/tool-policy/managed-allow.json: no such file or directory
    tier_render_test.go:533: entry built by BuildInto then removed from the policy was kept as user-added: allow = [Bash(managed-a:*) Bash(managed-b:*)], want [Bash(managed-a:*)]
--- FAIL: TestBuildIntoPersistsManagedAllowRecord_RemovedEntryNotKeptAsUserAdded (0.01s)
FAIL    github.com/modu-ai/moai-adk/internal/config/toolpolicy  0.314s
```

**GREEN evidence** (post-fix, same selector):

```
--- PASS: TestRenderTierPermissions_FailedProjectRenderLeavesSettingsUntouchedRecordSafe (0.01s)
--- PASS: TestRenderTierPermissions_RecordStageFailureLeavesProjectSettingsUntouched (0.00s)
--- PASS: TestRenderTierPermissions_SettingsWriteFailureAfterRecordStage_ClassifiesCorrectlyNextRender (0.01s)
--- PASS: TestBuildIntoPersistsManagedAllowRecord_RemovedEntryNotKeptAsUserAdded (0.01s)
PASS    ok      github.com/modu-ai/moai-adk/internal/config/toolpolicy  0.297s
```

**Post-fix verification** (env-scrubbed, this run, this tree, HEAD 026ba839a): `go test -count=1 ./internal/config/toolpolicy/ ./internal/core/project/` → both `ok` (0.450s / 2.238s); `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0; `go test -count=1 ./internal/config/toolpolicy/ -cover` → `coverage: 91.0% of statements` (threshold ≥ 85%; round 1 measured 90.4%); `golangci-lint run --timeout=2m ./internal/config/toolpolicy/` → `0 issues.`; `gofmt -l` → clean. Named re-confirmations: round-1 P1 tests `TestRenderTierPermissions_RemovedGeneratedEntryIsDropped` + `TestRenderTierPermissions_PersistsManagedAllowRecord` PASS; `TestAC005PolicyPathKeepsUserAllowEntry` + `TestAC005WithRecordKeepsUserAddedEntry` PASS unmodified.

**Residual corner (named, not worked around)**: the two-file update (settings + record) is not atomic, so a settings-write failure that coincides with a policy REVOCATION in the same render cannot keep the revoked entry's managed status under exact-generation converge — after recovery the entry is reclassified user-added and persists. The stage-ahead union narrows this to that single coincidence window (new entries are always recorded before settings can gain them; the record-ahead protocol removed the lag hazard P2-1 named). Closing the window fully would need a compensating record restore on the settings-write failure path — un-decided surface, named for the leader to disposition if wanted.

### Card-review repair round 3 (finding A) — init wiring test aligned with the preserved-defaultMode contract

**Finding** (round-3 finding A, CONFIRMED RED by lane measurement): the M1 keep-rule (Q1 — an existing USER `permissions.defaultMode` is preserved; the tier default is written only when the key is absent) regressed the pre-existing wizard wiring test `internal/cli/init_autonomy_wiring_test.go` `TestRunInit_WizardAutonomyTierAppliesBundle`, which asserted the tier OVERWRITES a seeded `defaultMode: "default"`. CI-blocking against a pre-existing test. RED (lane-captured): `go test -run TestRunInit_WizardAutonomyTierAppliesBundle ./internal/cli/` → `--- FAIL: TestRunInit_WizardAutonomyTierAppliesBundle` / `init_autonomy_wiring_test.go:157: USER permissions.defaultMode = "default", want "auto" (tier automatic)`.

**Disposition: fixed** (commit db70a4266) — test-only, no production change:

- `TestRunInit_WizardAutonomyTierAppliesBundle` re-scoped to the new contract as two subtests: **`existing-defaultMode-preserved`** — the seeded `defaultMode: "default"` survives the wizard's tier application (expect `"default"`; the teammateMode/env verbatim-preservation assertions are kept unchanged); **`absent-defaultMode-applies-tier`** — USER settings with a permissions block carrying NO defaultMode (only an allow list) receive `defaultMode: "auto"` from the tier, with the allow entry and teammateMode preserved (REQ-001). Both subtests reuse the file's existing seam helper (`runInitForAutonomyAtHomeCapturingOut`), so the isolated-HOME + wizard-seam mechanism is unchanged.

**Class sweep** (`grep -rn "WizardAutonomyTier\|defaultMode.*auto\|want.*auto" internal/cli/*_test.go internal/core/project/*_test.go`) — every hit read, dispositions:

- `init_autonomy_wiring_test.go:157` — the stale overwrite expectation → FIXED (this repair).
- `init_autonomy_wiring_test.go:182` (`FlagAutonomyTierNonInteractive`) and `:248` (`FlagFullyAutonomousWithoutProofDowngrades`) — both run against a FRESH temp HOME (no pre-seeded USER defaultMode) → absent-key path → expecting `"auto"` is CORRECT under Q1. No touch; they double as the flag-path absent-key confirmations.
- `launcher_test.go` `TestSyncPermissionModeToSettingsLocal` (incl. "mode overwrites existing defaultMode") — a DIFFERENT mechanism: the explicit launcher mode sync into `settings.local.json` (user-invoked mode flag, card t1414 lineage), not autonomy-tier application. Deliberate overwrite contract. No touch.
- `cg_mode_hardening_test.go`, `clifix_concurrency_repro_test.go` — concurrent-write preservation of `bypassPermissions` in `settings.local.json`; assert PRESERVATION, different mechanism. No touch.
- `doctor_settings_defaultmode_test.go` — the doctor dead-value check; reads/flags, applies nothing. No touch.
- Remaining regex hits (`factory_*`, `codex_*`, `coverage_fixes`, `doctor_hook`, `glm_flash_default`, `hook_*`, `init_quiet_wizard`, `init_workflow_*`, `lane_entry_m2`, launcher `--auto` directives) — matched "auto" in unrelated identifiers (auto-merge, --auto, auto-dispatch, auto-compact), not defaultMode-overwrite tests. No touch.
- `internal/core/project` test files — zero defaultMode hits. Clean.

**GREEN evidence** (post-fix, this run, this tree, HEAD db70a4266):

```
$ go test ./internal/cli/ -run 'TestRunInit_WizardAutonomyTierAppliesBundle' -v -count=1
--- PASS: TestRunInit_WizardAutonomyTierAppliesBundle (3.10s)
    --- PASS: TestRunInit_WizardAutonomyTierAppliesBundle/existing-defaultMode-preserved (1.28s)
    --- PASS: TestRunInit_WizardAutonomyTierAppliesBundle/absent-defaultMode-applies-tier (1.79s)
PASS
ok      github.com/modu-ai/moai-adk/internal/cli  4.368s
```

**Post-fix verification** (env-scrubbed): whole `TestRunInit` neighbor family `go test ./internal/cli/ -run 'TestRunInit' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 44.113s` (full internal/cli suite still out per the t1542 measurement — only the touched tests + prefixed neighbors run); `go test -count=1 ./internal/config/toolpolicy/ ./internal/core/project/` → both `ok` (0.431s / 2.573s); `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0; `gofmt -l internal/cli/init_autonomy_wiring_test.go` → clean.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-10-14T02:39:16Z"
run_commit_sha: e581bdbb20   # the M2+M3 implementation commit; this §E.3 block rides the following evidence-close commit
run_status: complete   # release-blocking in-scope set (AC-001..AC-005) fully green; AC-013 gate carries a measured divergence owned by the leader (§E.2)
ac_pass_count: 5
ac_fail_count: 1   # AC-013 (class: gate, not release-blocking); divergence + containment measurement recorded in §E.2
preserve_list_post_run_count: 0   # no plan.md §A.5 PRESERVE violation; changed files = plan items 1-3 + the Q2 template file + two invalidated comment sentences in launcher.go + one consumer test re-pinned to the Q2 contract
l44_pre_commit_fetch: not-performed   # lane-local run on a card worktree; no push per git-flow lane protocol (B9)
l44_post_push_fetch: not-performed   # no push performed by the lane
new_warnings_or_lints_introduced: 0   # golangci-lint: 0 issues
cross_platform_build.native: exit 0   # go build ./...
cross_platform_build.windows: exit 0   # GOOS=windows GOARCH=amd64 go build ./...
total_run_phase_files: 8   # 6 implementation/template/test files + spec.md + progress.md
m1_to_mN_commit_strategy: "M1 entry (057db17c41) + single M2+M3 implementation commit (e581bdbb20) + this evidence-close commit"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: "2026-10-14T03:59:35Z"
sync_commit_sha: "pending-backfill-sync"   # placeholder — a commit cannot cite its own SHA; backfilled with the real SHA in the following commit (D3)
sync_status: complete
ac_source: ".moai/specs/SPEC-USER-SETTINGS-PROTECT-001/acceptance.md"   # Tier M: acceptance.md is the AC source; spec.md ACs are inline-NA
b12_self_test_a: pass   # pre-emission grep: grep -c 'SPEC-USER-SETTINGS-PROTECT-001' CHANGELOG.md → 0 (exit 1, no prior entry)
b12_self_test_b: pass   # AC counter (reserved-token grammar) on acceptance.md: live=13, excluded=0, ambiguous=0 (exit 0); the file declares 13 identifiers of which the SPEC's own scope record moves AC-006..AC-012 out of this card (t1666 / gate) — the entry names the 6 in-scope ACs (AC-001..005 + AC-013) per acceptance.md scope record + §E.3
b12_self_test_c: pass   # every repo path named in the CHANGELOG entry verified to exist (spec.md, progress.md, tier_render.go, codegen.go, settings.json.tmpl)
changelog_entry_position: "CHANGELOG.md [Unreleased] § Changed, first entry"
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed (merged into the single sync commit; 3-phase close)"
  plan_md: "no frontmatter — artifact statelessly omits it (schema-permitted); nothing refreshed"
  acceptance_md: "no frontmatter — artifact statelessly omits it (schema-permitted); nothing refreshed"
canary_compliance_check:
  applicable: false   # this SPEC defines no forward-looking policy that its own sync tests enforce
mx_tag_validation: "sync sub-step pass — new exported writers carry @MX:NOTE [AUTO] tags (tier_render.go WriteUserDefaultMode / mergeProjectAllow); no TODO/WARN debt added"
```


## §G Lane design decisions (plan phase)

### Tier decision (spec.md §0)

Decision: Tier M was kept at the pre-split scope of 16 files, option (b), as ratified by ruling d-20261010T081810Z-49e5. Revision 2 does not re-decide the tier. The in-scope set is four files (plan.md §A.4 Basis: items 1 to 3 and the Q2 template file), below the Tier M file band of 5 to 15, which is a guide and not a gate. The leader's decision, board record d-20261010T094229Z-74f2, keeps Tier M for the split scope.

In-scope files after revision 2: 4 (plan.md section A.4 Basis items 1 to 3, and the template file). Tier M is kept as recorded; the 5 to 15 file band is a guide, not a gate.

decision record: decided_by=lane-18 evidence_refs=.moai/specs/SPEC-USER-SETTINGS-PROTECT-001/spec.md §0; .moai/specs/SPEC-USER-SETTINGS-PROTECT-001/plan.md §A.4 Basis; board:d-20261010T081810Z-49e5 ladder_path=ladder ⑤ (lane judgment; the leader's ruling confirms option (b))

### G-17 — template defaultMode value

Decision: the template carries `permissions.defaultMode` = `"default"`. `acceptEdits` is not chosen: it would auto-approve edits, a behaviour change beyond ruling d-20261010T073547Z-07ae.

decision record: decided_by=lane-18 evidence_refs=.claude/skills/moai-foundation-cc/reference/claude-code-settings-official.md:89; internal/cli/launcher.go:737-741; board:d-20261010T081810Z-49e5 ladder_path=ladder ⑤ (lane judgment; ruling item 2 places this choice in the lane's ladder)

### G-20 — managed-block detection

Decision: diff-based. The writer keeps the last generated managed allow list in `.moai/state/tool-policy/managed-allow.json` as `last_generated`. user_added = existing allow entries minus `last_generated`; result = regenerated ∪ user_added. With no record, every existing entry is kept, so nothing is removed. Only the user removes user-added entries (ruling d-20261010T073547Z-07ae).

decision record: decided_by=lane-18 evidence_refs=board:d-20261010T073547Z-07ae (marker- or diff-based); internal/config/toolpolicy/tier_render.go:77; .moai/reports/t1630/evidence/probe-ac005_test.go.txt; .moai/reports/t1630/evidence/probe-ac005b_test.go.txt ladder_path=ladder ⑤ (lane judgment within the record's "marker- or diff-based" allowance)

### G-23 — run and db in the --force allowlist

Out of Scope - moved to t1666 (decision Q3). The text is retained unchanged.

Decision: the `run` and `db` categories join the `--force` allowlist of `moai clean --home`, so `--force` can delete run and db candidates. Run items are candidates only under the REQ-011 rule, and `--force` deletes only candidates. Without `--force`, run and db items are listed only. The ruling 07ae reads db as list-only, with deletion only by explicit `--yes`, which ruling 49e5 maps to `--force`. No live-record exclusion is added for db, because the ruling does not state one. plan.md's G-22 (the tier_render.go:106 statement) is a separate gap.

decision record: decided_by=lane-18 evidence_refs=board:d-20261010T073547Z-07ae; .moai/specs/SPEC-USER-SETTINGS-PROTECT-001/spec.md REQ-011 and REQ-012 ladder_path=ladder ⑤ (lane judgment; the ruling's explicit-deletion wording)

## §G Override and Refusal Record

- 2026-10-10T11:22:33Z SPEC-USER-SETTINGS-PROTECT-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=.moai/reports/t1630/plan-audit-iter1.md,.moai/reports/t1630/plan-audit-iter2.md,.moai/reports/t1630/plan-audit-iter3.md

decision record: decided_by=lane-18 evidence_refs=.moai/reports/t1630/plan-audit-iter3.md (verdict PASS, audited_sha fb672d1723c485ac52d4b21f8e4f3bbceafc78c9, overall 0.81, blocking 0), the ceiling-outcome record above (pass-through at 11:22:33Z), factory record t1630 version 28 ladder_path=plan to run Kickoff gate row, autonomous form per auto-semantics 9.1: audit-cross PASS plus evidence and a decision record condition=PASS holds only if the operator accepts the non-normative scope of REQ-003 and REQ-004 (verdict section 5), otherwise FAIL

## §F Phase 4 Mode Selection

Input parameters: tier M (frontmatter); scope = 3 module paths (internal/config/toolpolicy, internal/core/project, internal/template/templates) plus tests; domain count = 1 code domain (Go CLI/config) + template mirror; file language mix = Go + YAML template; concurrency benefit = LOW (coding-heavy, sequential TDD milestones); Agent Teams prereqs = not requested (no operator --team).

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | not selected | Multi-file behavioral change, not a one-liner |
| serial | selected | Coding-heavy TDD work; Anthropic coding-task parallelism caveat; single writer per tree |
| fanout | not selected | No multi-domain research; coding-heavy work |
| sweep | not selected | Not mechanical-uniform; new-code work |

Decision: serial

Justification: the SPEC is a coding-heavy Tier M behavioral change across three Go/template paths with TDD milestones; parallel spawns would violate one-writer-per-tree on a single card tree, and there is no mechanical-uniform bulk transform. Serial manager-develop delegation per milestone is the safe default. Kickoff gate met in its autonomous form (§9.1): plan-audit iter3 PASS 0.81 (blocking 0, audited_sha fb672d1723) + committed decision record 947719542 + operator ratification d-20261010T113615Z-6892; recorded before the first run-phase Agent() spawn.
