# t1259 — run-phase blocking dependencies, measured

Measured by the lane (orchestrator) on 2026-09-26, in worktree
`.claude/worktrees/t1259` (branch `WT-local-instructions`), against `origin/develop`
freshly fetched in this run. Closes the gap plan-audit iter1 left as "the lead's read
at dispatch".

`plan.md` §B declares two [HARD] blocking dependencies for the **run** phase. Both are
**unmet** at the time of measurement. The plan phase itself is unaffected.

## D-1 — parent `SPEC-INSTRUCTION-FILES-UNIFY-001` M2 has not landed

Claim: the parent SPEC's iteration-order change (REQ-IFU-006) is not on `origin/develop`.

Evidence — two independent reads, both negative:

```
$ git show origin/develop:.moai/specs/SPEC-INSTRUCTION-FILES-UNIFY-001/spec.md | grep -m1 -i '^status:'
status: draft

$ git show origin/develop:internal/cli/codex_launcher.go | grep -n 'local.md\|localInstruction\|LocalInstruction'
125:	for _, name := range []string{codexClaudeLocalName, codexLocalInstructionName} {
126:		body, err := readCodexLocalInstruction(projectRoot, name)
```

The loop at :125 still reads `codexClaudeLocalName` **before** `codexLocalInstructionName`.
REQ-IFU-006 requires the reverse order, so the change is absent from the code as well as
from the SPEC status — the status read and the code read agree.

Consequence: this SPEC's M-Codex-advisory milestone edits the fallback branch of that same
loop and builds on the reordered form. Starting it now would write against an order that is
about to change, in the one function both SPECs touch.

## D-2 — t1175 (`SPEC-ALWAYS-LOADED-DIET-002`) is not on `origin/develop`

Claim: the rules-diet SPEC this one inherits a dependency on is absent.

Evidence:

```
$ git ls-tree -r --name-only origin/develop -- .moai/specs/ | grep -ci 'ALWAYS-LOADED-DIET-002'
0
```

Note the near-miss that must not be read as a pass: `SPEC-ALWAYS-LOADED-DIET-001` IS present
and reads `status: completed`. It is a **different SPEC** — the `-001`/`-002` distinction is
the whole of the difference, and taking the completed sibling as the dependency would be a
false clear.

## What this does and does not establish

- **Establishes:** neither dependency was satisfied on `origin/develop` at this measurement.
- **Does not establish:** that either is still unsatisfied later. Both are moving; the read
  decays and is re-run at run-phase dispatch, not carried from here.
- **Not measured:** whether either SPEC's work exists unmerged in its own card worktree. The
  dependency `plan.md` §B declares is on landing, so an unmerged branch does not satisfy it —
  but it does mean "absent from develop" is not the same as "not yet done".

## Residual risk

`origin/develop` advances continuously under other lanes. A dependency can land between this
measurement and dispatch, which would make this record stale in the permissive direction —
the reason the re-read at dispatch is required rather than optional.
