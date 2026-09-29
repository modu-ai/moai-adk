# progress.md — SPEC-TODO-SURFACE-POLISH-001 (카드 t1349)

## §E.1 Plan-phase Audit-Ready Signal

auditor_model: glm-5.3-flash[1m]
plan_status: audit-ready
plan_complete_at: 2026-09-30
plan_audit_verdict: PASS 0.90 (iter1 — Tier M threshold 0.80)
plan_audit_report: .moai/reports/t1349/plan-audit-iter1.md
findings_note: D2-D7/D9는 경미(minor)로 수용 — run 단계에서 사소한 경우에만 반영.

## §E.2 Run-phase Evidence

Baseline attribution: every measurement below was taken in this run, in the
card worktree `.moai/worktrees/t1349`, branch `WT-todo-surface-polish`. The
RED cells are pinned to tree `b33246d3e` (the plan-artifact commit, the
pre-fix tree); GREEN cells carry the commit SHA of the milestone that
flipped them. Test invocations are env-scrubbed single compound
invocations (`unset MOAI_KANBAN … && go test …`) per the lane discipline.

### M1 — CLI surface (① show, ② add -f, ③ list limit, ⑥ render completeness)

**RED(①) — show verb absent (AC-TSP-001 pre-work, tree b33246d3e).**
Command: `go test ./internal/cli/ -run '^TestTodoShow$' -count=1` — exit 1.
Verbatim failing output (abridged to the decisive lines; full transcript in
session log):

```
--- FAIL: TestTodoShow (3.47s)
    --- FAIL: TestTodoShow/live_line_carries_the_full_text (0.62s)
        todo_show_test.go:41: show t1: todo: "show" is not a todo verb and "t1" is a card id — refusing to create a card named "show t1".
            Known verbs: add, analyze, auto-done, done, drop, edit, export-json, history, hold, landed, list, move, next, pr, relate, triage, undone, undrop, unhold, unpick, unrelate, why
    --- FAIL: TestTodoShow/absent_id_prints_absent_and_the_issued-mark_qualifier (1.96s)
        todo_show_test.go:68: show t3: todo: "show" is not a todo verb ... (same refusal)
    --- FAIL: TestTodoShow/body_with_tabs_and_newlines_flattens_without_truncation (0.89s)
        todo_show_test.go:87: show t1: todo: "show" is not a todo verb ... (same refusal)
```
RED reason: `show` is not in the AddCommand registration, so the parent
mistyped-verb guard refuses the verb-shaped first token — exactly the
refusal the guard's own guidance cites as the address grammar.

**RED(②) — add `-f`-leading body misparsed (AC-TSP-010 pre-work, tree
b33246d3e).** Command: `go test ./internal/cli/ -run
'^TestTodoAddLeadingDashText$' -count=1` — exit 1. Verbatim output:

```
--- FAIL: TestTodoAddLeadingDashText (0.27s)
    todo_add_dash_test.go:33: add "-f hello world": unknown shorthand flag: 'f' in -f hello world
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.011s
```
RED reason: pflag's interspersed parser consumes every `-`-prefixed token
as flags; no `--force` shorthand is registered, so the quoted body dies as
an unknown shorthand flag.

**RED(③) — default limit 20 truncates (AC-TSP-020 pre-work, tree
b33246d3e).** Command: `go test ./internal/cli/ -run
'^TestTodoListDefaultLimit$' -count=1` — exit 1. Verbatim decisive lines:

```
todo_list_limit_test.go:62: bare todo rendered 20 ids, want all 60 (stdout "t1\tqueued\tlimit fixture card 1\n…"...)
todo_list_limit_test.go:65: bare todo stderr = "…\nlist: 40 rows withheld — showing 20 of 60 (--limit 0 lists all)\n", want no withheld notice at 60 rows under the 100 default
todo_list_limit_test.go:76: list rendered 20 ids, want exactly the 100-row bound
todo_list_limit_test.go:80: list stderr = "…\nlist: 100 rows withheld — showing 20 of 120 (--limit 0 lists all)\n", want the withheld line "list: 20 rows withheld — showing 100 of 120 (--limit 0 lists all)"
```

**RED(⑥) — render-completeness instrument observed failing (mutant probe,
AC-TSP-030; tree b33246d3e-based, pre-GREEN).** The property test passes on
the unmodified tree (the filter already admits hold); its observed failure
comes from the planted mutant — a hold-excluding branch inserted into
`runTodoListRoot`, probe run, then reverted. Command: `go test
./internal/cli/ -run '^TestTodoRenderCompleteness$' -count=1` against the
mutant — exit 1. Verbatim output:

```
--- FAIL: TestTodoRenderCompleteness (1.10s)
    todo_list_limit_test.go:197: default render id set = map[t1:true t2:true t5:true], want exactly map[t1:true t2:true t4:true t5:true] (render must equal the filter-satisfying set)
    todo_list_limit_test.go:201: default render is missing live card t4 — a store row the filter admits vanished from the render
    todo_list_limit_test.go:205: default render omits the hold card t4 — hold is a live state and belongs in the default view
```
The instrument's failure on a known-bad render is the observed-failure
evidence verification-completeness §1.1 requires.

**GREEN — M1 flips (tree: M1 milestone commit).** Command: `go test
./internal/cli/ -run '^(TestTodoShow|TestTodoShowReadOnly|TestTodoAddLeadingDashText|TestTodoListDefaultLimit|TestTodoRenderCompleteness|TestTodoAddPick|TestTodoAddForce|TestTodoAddClassificationFile|TestTodoBareFallthrough|TestTodoListLimitZero|TestTodoListJSONIgnoresLimit|TestTodoListNegativeLimit|TestTodoListDroppedOnly)$' -count=1`
— exit 0:

```
ok  	github.com/modu-ai/moai-adk/internal/cli	19.552s
```

**M1 regression sweep** — `go test ./internal/cli/ -run '^TestTodo'
-count=1` surfaced the frozen verb-surface guard (TestTodoVerbSurfaceZeroDelta,
AC-TOSQ-010), whose declaration mechanism (not the frozen table) records
post-freeze sanctioned changes. Declared: the `show` verb addition
(SPEC-TODO-SURFACE-POLISH-001 REQ-TSP-001/003) and the `list --limit`
default 20→100 (REQ-TSP-020). After the declarations the sweep exits 0:
`ok github.com/modu-ai/moai-adk/internal/cli 68.501s` (surface + add-path
regression groups, post-declaration run).

**SPEC-required test evolutions** (recorded per the no-silent-test-edit
discipline — each is a behavior change the SPEC itself mandates, not a
weakening):

1. `todo_verb_leak_test.go` (t555 matrix) — the matrix used `show` as the
   mistyped-token exemplar because `show` was unregistered; REQ-TSP-001
   registers exactly that grammar, so `show t401` now ANSWERS instead of
   being refused. The three `show`-led refusal rows moved to an
   unregistered stand-in token (`peek` — no near-miss of any verb), the
   `Show`-capitalized row stays (routing is name-exact), and the control
   test gained `{"show","t401"}` / `{"show","401"}` asserting the new
   registration creates no card. Post-fix: `go test ./internal/cli/ -run
   '^(TestTodoVerbGuardRefusesWidenedMistypes|TestTodoVerbGuardControlRegisteredVerbs|TestTodoVerbGuardStillAddsNonAddresses)$' -count=1`
   → `ok ... 20.751s`.
2. `todo_runtime_store_test.go` (kanban, M3) — the readback test asserted
   the seeded legacy label `worker-2` verbatim; REQ-TSP-052 mandates the
   write path records only the canonical vocabulary, so the expectation is
   `lane-2`. Post-fix: `go test ./internal/kanban/ -run
   '^TestTodoRuntimeStorePublicReadbackSurvivesCardEdit$' -count=1` →
   `ok ... 2.596s`.

### M2 — ghost stores (④ detection, once-notice, doctor inventory, doc fix)

**Diagnosis record (REQ-TSP-031 / AC-TSP-031 — the t1338 observation).**
Three legs, each mechanically observed in this run:

1. **Reproduction attempt** — the plan-phase re-measure (2026-09-29, this
   session's predecessor) ran `moai todo list --limit 60 | grep -c t1338`
   against the live home queue and counted 4 (the card renders in one row
   plus findings lines): the one-off non-render observation was NOT
   reproduced at plan time, and the run phase adds no counter-evidence —
   the render filter's only exclusion is `dropped`
   (`runTodoListRoot`, todo.go), so a queued row cannot be skipped by
   state.
2. **v1/v2 CHECK-constraint hold-write behavior** — observed with a
   hand-built v1-schema database (3-state CHECK, `schema_version=1`,
   the seq/landing column list the rebuild copies) driven through the
   store's own `Mutate` hold write. Result: the write SUCCEEDS — the
   adopting open's `ensureSchema` rebuilds the v1 items table to the
   4-state v2 layout first (`rebuildItemsTable`), the row lands as `hold`,
   and the post-write stamp reads `schema_version="2"`. Verbatim probe
   output (temporary diagnosis test, since deleted; evidence preserved
   here):
   ```
   hold write on hand-built v1 DB: err=<nil>
   post-write schema_version="2" items.t1.state="hold"
   LoadPure items=1 first-state="hold"
   ```
   A CHECK violation on the hold write path is unreachable through the
   verb surface.
3. **Conclusion** — the t1338 observation is a non-reproducible one-off
   with no reachable structural cause remaining: the render filter admits
   every live state (now pinned by the AC-TSP-030 property test, whose
   instrument was mutant-probed red), and the hold write path on a v1
   database is protected by the rebuild migration. The structural guard
   the SPEC demanded is the invariant test itself.

**Detector extension (AC-TSP-040)** — `TestInspectStaleLocalStoresGhostClasses`
+ `TestInspectStaleLocalStores_LiveJSONQueueIsNotAGhost` in
internal/kanban pass (`ok github.com/modu-ai/moai-adk/internal/kanban`),
covering: each ghost class reported as its own fact with path/class/bytes,
sha-unmodified evidence across the probe, the SQLite divergence verdict
uncontaminated, named artifacts (companions.json, leads.json) not
swallowed, and the live pre-SQLite JSON queue not misreported as a ghost.

**Once-per-class notice (AC-TSP-041)** — `TestGhostNoticeOnce` in
internal/cli: first `list` carries the ghost notice naming `legacy-json`;
the second `list` is silent; stdout is byte-identical across the two runs;
the marker lands in the alive state directory recording the class; ghost
file shas are unchanged and the pure ghost directory plus the home queue
directory gain no entry. Note (interpretation recorded per §B.4): the AC's
"no file added to the ghost directory" is judged over the ghost-only
snapshot directories; the alive state directory — which the launcher
already fills with live runtime artifacts (companions.json, leads.json,
the autodone log) — hosts the marker, because REQ-TSP-041 itself names the
alive state directory as the marker's home and the ghost evidence stays
byte-untouched.

**Doctor ghost inventory (AC-TSP-042)** — `TestDoctorTodoGhostInventory`
(3 states: absent OK / present WARN with path·class·bytes / unreadable
FAIL) passes; the same commit carries the constant registration
(`todoGhostInventoryCheckName`), the `namesAddedAfterBaseline` bare key,
and regenerated goldens — the t1251 bundle, verified by
`go test ./internal/cli/ -run
'^(TestBinaryLag_AllowlistKeysAreLiveNames|TestBinaryLag_DoctorCheckNameSetIsUnchanged)$'`
and the three `TestDoctorGolden_*` tests, all `ok`.

**Storage doc fix (AC-TSP-043)** — template source
`internal/template/templates/.moai/docs/todo-queue-storage.md` rewritten
(home-DB canonical layout; the `backlog.json` "safe to delete" row
corrected to the ghost framing; a ghost-artifact table added) and mirrored
byte-identical to the tracked `.moai/docs/todo-queue-storage.md` in this
branch (`cmp` → identical). Greps: `backlog.db` ×10, `~/.moai/db/` ×2,
both paths tracked by git.

### M3 — owner_label vocabulary (⑤ migration, doctor drift, write normalization)

**One-time locked migration (AC-TSP-050)** — `TestOwnerLabelMigration` in
internal/kanban: a fixture table carrying lead ×3 / worker-67 ×2 / lane-1
×1 migrates 5 rows in one locked write to leader ×3 / lane-67 ×2, leaves
the canonical row untouched, keeps the items table fingerprint
byte-identical, and a second run migrates 0 rows (idempotence). Command:
`go test ./internal/kanban/ -run '^(TestOwnerLabelMigration|TestOwnerLabelWritesCanonical)$' -count=1`
→ `ok github.com/modu-ai/moai-adk/internal/kanban 1.995s` (final tree).
Operator-facing counts for the measured production table (lead 453 /
worker-67 13, 2026-09-29): the migration is value-only (no column or
schema change), so it is safe to run against the home DB through
`BacklogStore.MigrateOwnerLabelVocabulary` whenever the operator invokes
it; the doctor check below is the reproducible verdict that it happened.

**Doctor label drift (AC-TSP-051)** — `TestDoctorOwnerLabelDrift` (0 rows
→ OK; legacy rows → WARN naming the labels and counts), counting through
the SAME detectors the refusal paths use (kanban.IsLegacyLeaderSpelling /
kanban.IsLegacyFactoryRoleValue — no second detector). The same commit
carries `ownerLabelDriftCheckName` registration + allowlist bare key +
regenerated goldens (verified with the binary_lag and TestDoctorGolden_*
commands above, final tree `ok`).

**Write normalization (AC-TSP-052)** — `TestOwnerLabelWritesCanonical`: a
legacy label carried into the write path lands relabeled (worker-12 →
lane-12); the table stays legacy-free. The existing readback test's
expectation moved to the canonical value (see the M1 test-evolution
record), post-fix `ok ... 2.995s` family above.

**Migration traceability note.** The migration rides
`BacklogStore.MigrateOwnerLabelVocabulary` (locked, idempotent,
value-only) and is additionally invoked inside `recordRuntime`'s locked
transaction, so the next runtime event on any queue completes the relabel;
the doctor "Owner Label Drift" check reports the remaining count either
way. No column, schema, or other table changes — REQ-TSP-050's scope is
exactly one column.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-30
run_commit_sha: "5d3389913"
run_status: complete
ac_pass_count: 15
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (card-worktree stacking; no push by lane per git-flow lane protocol)
l44_post_push_fetch: not-run (push is the leader's batch act)
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin_arm64: pass (go build ./... exit 0, final tree)
cross_platform_build.windows_amd64: pass (GOOS=windows GOARCH=amd64 go build ./... exit 0, final tree)
total_run_phase_files: 30
m1_to_m3_commit_strategy: one commit per milestone (M1 surface, M2 ghost stores, M3 label vocabulary), explicit pathspec staging; the doctor surface files shared by M2+M3 were landed in their respective milestone commits with the tree held at each milestone's state (goldens regenerated per state)
verification_scope: lane-local (internal/cli, internal/kanban full suites; no go test ./... per the lane load discipline)
suite_results: internal/cli 1 pre-existing load-sensitive timing flake (TestStopChainMemberCostWithinBudget, failed under concurrent-lane load, green on isolated re-run: --- PASS 8.63s, exit 0); every other package and scoped suite green
coverage: internal/cli 85.1%, internal/kanban 85.1% (measured with -cover on the full touched-package suites this run; >= 85% threshold met)
lint: golangci-lint v2.1.6 (CI-pinned) on internal/cli + internal/kanban — 0 issues, final tree
boundary_grep: AskUserQuestion/mcp__askuser in internal/cli+internal/kanban non-test non-comment lines: 44, all pre-existing baseline (help texts and transcript fixtures), 0 in files this SPEC touched

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-09-30
sync_commit_sha: "pending-backfill-sync"
sync_status: complete
changelog_entry_position: [Unreleased] > Added (top, above the SPEC-TODO-CLASSIFY-DISPATCH-001 row — most-recent-first convention)
b12_self_test_a: PASS — pre-emission grep `grep -c 'SPEC-TODO-SURFACE-POLISH-001' CHANGELOG.md` = 0 before the entry landed (no duplicate from a parallel BATCH-SYNC session)
b12_self_test_b: PASS — acceptance.md distinct live AC identifiers = 15 (AC-TSP-001..052; zero `[RETIRED]`/`[REF]` tokens) = CHANGELOG entry count 15
b12_self_test_c: PASS — every file path named in the CHANGELOG entry verified present via ls: `.moai/specs/SPEC-TODO-SURFACE-POLISH-001/{spec,progress}.md`, `internal/cli/{todo_show,todo_history,todo,todo_ghost_notice,doctor_todo_ghost,doctor_owner_label}.go`, `internal/kanban/{todo_stale_store,todo_owner_label,todo_runtime}.go`, `internal/template/templates/.moai/docs/todo-queue-storage.md`, `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md`
frontmatter_status_transitions: in-progress → implemented → completed (merged 3-phase close on this single sync commit); `updated:` 2026-09-30 (already the current date — value unchanged)
canary_compliance_check: n/a — this SPEC defines no forward-looking policy covered by its own sync tests
mx_tag_validation: PASS with 1 update — the `@MX:ANCHOR` on `discloseQueueLayout` (`internal/cli/todo_disclosure.go`) refreshed to fan_in=5 with its caller enumeration extended to (bare/list, show, why, pr, triage); this SPEC's show verb added the fifth caller, so the ANCHOR-update duty fired (MX protocol: update when fan_in changes). No other tag gap: the new functions carry no fan_in≥3 surface, no goroutines, no complexity≥15; every new exported symbol (`MigrateOwnerLabelVocabulary`, `GhostArtifact`, the three `GhostClass*` constants) is godoc'd, tested, and inside the already-ANCHORed `InspectStaleLocalStores` file
readme_docs_site_decision: default-limit-20 — NO statement exists in README*.md or docs-site/content (confirmed absent by grep; nothing to fix). Read-verb enumeration — `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md` line 248 enumerated the disclosing read verbs as list·why·pr·history, made stale by show's addition; FIXED in all 4 locales in this commit (show inserted after list, one token per locale, no structural change). README todo rows (ko L392 / en L393) are partial subcommand enumerations already missing pre-existing verbs (history, hold, why, pr, triage, export-json…) and claim no exhaustiveness — not made false by this SPEC, untouched. Pre-existing imprecision noted, not repaired (out of this SPEC's delta): the same docs enumeration omits `triage`, which also discloses via discloseQueueLayout since before this SPEC
sync_scope: CHANGELOG.md + progress.md §E.4 + spec.md frontmatter (status/updated only) + the MX ANCHOR refresh + the 4-locale docs fix; spec/plan/acceptance BODY content untouched (blocker-report boundary respected — no body edit was needed)

## §F Phase 4 Mode Selection

- inputs: tier M · scope ~10 files (internal/cli/todo.go, internal/kanban store layer, doctor registry, storage doc, SPEC artifacts) · domains 2 (CLI+store Go, docs) · language mix Go+md · concurrency benefit LOW (coding-heavy)
- evaluation: direct — no (multi-file semantic change) · fanout — no (Anthropic coding-task caveat: sequential default) · sweep — no (not mechanical-uniform, inter-file dependencies) · agent-team — no (not operator-requested)
- Decision: serial
- justification: coding-heavy single-SPEC implementation with interdependent CLI/store changes — one manager-develop per milestone (M1 → M2 → M3) is the safe default.
