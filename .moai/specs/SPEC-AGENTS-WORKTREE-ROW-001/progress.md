# progress.md — SPEC-AGENTS-WORKTREE-ROW-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-22
- plan_auditor: pending (independent audit is the next act after this commit; not run by this
  session — the lead dispatches plan-audit separately)

### Plan-phase evidence (plan-phase, this run, worktree t1071 @ `cd99336bf`, 2026-09-22)

Re-measurement of the card's investigator citations (spec.md §A, C1–C8), executed before any
artifact was written:

- **C1** `AGENTS.md:18-25` read — capability table, three rows, no worktree row. Negative
  control `grep -c '^| worktree-entry |' AGENTS.md` → `0`; positive control
  `grep -c '^| question-channel |' AGENTS.md` → `1`.
- **C2** `AGENTS.md:106-111` read — §3 launcher enumeration lists `moai cc -w`, `--spawn`,
  `EnterWorktree(<path>)` only; the `moai codex` form is absent.
- **C3** `internal/template/templates/AGENTS.md.tmpl:18-25` read — same three-row table, same
  absence. Negative control grep → `0`.
- **C4** `internal/template/templates/AGENTS.md.tmpl:291-305` read — `## 11. moai CLI Verbs`
  at `:291`, nine verb rows `:295-303`, no `moai codex` row. Negative control
  `grep -c '^| \`moai codex\`' …` → `0`; positive control `grep -c '^| \`moai init' …` → `1`.
  The dispatch's `:293-302` citation is corrected to `:293-303` (re-measured range governs).
- **C5** `internal/cli/codex_launcher.go` read — verb registration `:324` (`Use:` line, code
  declaration), `DisableFlagParsing: true` `:360`, `resolveCodexWorktreeDir` `:272-299` (L2
  validation `:279`, L1 join `:286`, absent-directory diagnostic `:289-297`), named diagnostics
  `:47`/`:52`, help text `:341-342`. Entry exists; creation is impossible.
- **C6** The dispatch's `:16-17` citation is a package-doc comment block (`codex_launcher.go:16-19`),
  not a declaration — the card's own comment-hit warning applied to the citation itself.
  Declaration anchors above are code.
- **C7** `grep -n 'moai CLI Verbs' AGENTS.md` → no matches — the root copy carries no Verbs
  table; the registration fix is mirror-only.
- **C8** `internal/template/templates/AGENTS.md.tmpl:270` prose-references `moai codex` while
  the Verbs table omits it — the command is live; only the inventory lags.

Pre-edit control batch (one compound invocation, 2026-09-22):

```
G1-root-worktree-row: 0
G2-tmpl-worktree-row: 0
G3-tmpl-codex-verb: 0
G4-positive-control-root-table: 1
G5-positive-control-verbs-table: 1
exit=0
```

Interpretation: the three zeros are evidence of absence (both positive controls fired on the
same run), and the two ones prove the row-shape patterns fire against the tables as they stand —
so the post-edit guards measuring `1` will be a real flip, not a pattern that matches nothing.

Ambiguity decisions taken (for the plan-auditor):

1. The dispatch named the root capability table and the mirror Verbs table explicitly, but
   re-measurement found the mirror's *capability table* also lacks the row (C3). Constraint (c)
   ("judge each copy separately and fix each accordingly") was read as licensing the fix; it is
   REQ-AWR-002.
2. The dispatch's `:16-17` coordinate was replaced by code anchors (`:324`, `:272-299`) per the
   card's own comment-hit rule.
3. `tier: M` was chosen over `S` because the card requires `acceptance.md` explicitly and the
   verification chain (guards + build + lint) exceeds a two-artifact scope.

SPEC lint at plan-phase close (tree build `go run ./cmd/moai spec lint
SPEC-AGENTS-WORKTREE-ROW-001`, this run, this tree): exit 0, `0 error(s), 4 warning(s)` —
REQ collection FIRED (four REQ-AWR-* rows collected and modality-judged; REQ-AWR-004 emitted no
finding). The four findings are `ModalityMalformed` on REQ-AWR-001/002/003/005 — the same
prose-shape class card t1057 measured; they are warnings, not errors, and are left for the
plan-auditor to weigh rather than re-worded mid-flight by this plan phase.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

---

🗿 MoAI
