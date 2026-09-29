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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- inputs: tier M · scope ~10 files (internal/cli/todo.go, internal/kanban store layer, doctor registry, storage doc, SPEC artifacts) · domains 2 (CLI+store Go, docs) · language mix Go+md · concurrency benefit LOW (coding-heavy)
- evaluation: direct — no (multi-file semantic change) · fanout — no (Anthropic coding-task caveat: sequential default) · sweep — no (not mechanical-uniform, inter-file dependencies) · agent-team — no (not operator-requested)
- Decision: serial
- justification: coding-heavy single-SPEC implementation with interdependent CLI/store changes — one manager-develop per milestone (M1 → M2 → M3) is the safe default.
