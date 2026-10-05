# SPEC Review Report: SPEC-ROLE-NAMING-DOCS-001
Iteration: 3/3 (Tier L ceiling 3, `.moai/config/sections/harness.yaml`:78; PASS threshold 0.85)
Verdict: PASS
Overall Score: 0.91 (iter-1 0.81 → iter-2 0.88 → iter-3 0.91; monotone, no STOP signal)

Audited tree: worktree `.claude/worktrees/t1257`, branch `WT-role-naming-docs`, HEAD `b110beb4c` (v0.4.0). Scope: delta re-audit of iter-2 defects N1-N5 plus a regression scan of the changed lines. Every resolution was checked against the files and by re-running the scripts; HISTORY claims were not relied on.
Reasoning context ignored per M1 Context Isolation.

## Must-Pass Results

- [PASS] MP-1: 25 REQ entries (`grep -cE '^- \*\*REQ-RND-[0-9]+' spec.md` → 25), sequential 001..025, unchanged since iter-2.
- [PASS] MP-2 (requirement layer): changed REQs (REQ-RND-020 spec.md, REQ-RND-021 table + note) keep their Event-driven / Ubiquitous forms.
- [PASS] MP-3: `moai spec lint spec.md` → `0 error(s), 0 warning(s)`; version "0.4.0".
- [N/A] MP-4: not multi-language tooling.
- [PASS] MP-5 D7: no referenced SPEC retired/superseded/archived; SPEC-ROLE-NAMING-CODE-001 still off-branch (SHOULD; gated by REQ-RND-002).
- [PASS] MP-6 D8: `syscall` count 0 in spec/plan/acceptance.
- [PASS] MP-7: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no output (exit 1).

## Category Scores

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.90 | 0.75-1.0 | REQ-RND-018 (a)-(e) and REQ-RND-021 table single-interpretation; qualifier spellings tied to a cited source (spec.md L96-L98). Residual descriptive "leader pane" wording in REQ-RND-005/014 (M3 below). |
| Completeness | 0.92 | 0.75-1.0 | Echo set now reproduces all previously-missed lines in all four locales (N1 resolved). |
| Testability | 0.85 | 0.75-1.0 | AC-RND-004 extraction scoped with a ledger exemption; AC-RND-020 group check is mechanical (reproduced below); AC-RND-008 uses the fixed script. Minor residuals M1, M2, M4. |
| Traceability | 1.0 | 1.0 | 25 AC headings, 25 REQ→AC rows. |

Aggregate (harmonic mean): 0.91.

## Regression Check (iter-2 defects)

- N1 ECHO-PATTERN-ASYMMETRY — **RESOLVED**. Re-run: `python3 .moai/reports/t1257/raw/scripts/echoes.py` → stderr `# total lines 148, files 60, mirror-added 18`, exit 0; `cmp` against committed `raw/q3-clause-echoes.tsv` → exit 0. `moai-todo.md`:143 and `factory-mode.md`:113 are present in en, ko, ja, zh (8 rows, pattern `who-picks`). The iter-2 probe (`one who picks|who picks|고르는 주체|挑选的主体|picks? (a|the) card|human does the picking`) re-joined against the new TSV: 12 hits, 12 IN, 0 OUT. AC-RND-020 group check reproduced: docs-site rows grouped by page path + line → `groups=23 bad=0` (every group has exactly four locales).
- N2 AC-004-FALSE-AT-ARRIVAL — **RESOLVED**. acceptance.md AC-RND-004 now extracts only from `moai cc` / `moai glm` invocations, exempts ledger "other meaning" matches, and cites the code-layer rows by their titles at `d0770b9cc` (`` `-f <role>` join token ``, `` `-f <label>` / `--name <label>` (lane) ``), which match `WT-role-naming-code:…/design.md` L30-L31. Running the new extraction on the current tree yields only `-f worker` 43, `-f worker-3` 8, `-f worker-<n>` 3 (all rename targets) plus `--name run` 36 / `plan` 28 / `sync` 24 / `mysession` 4 — no shell `-f` flag or prose remains (see M1 for the `--name` rows).
- N3 QUALIFIER-CROSS-SPEC-DRIFT — **RESOLVED**. REQ-RND-021 table uses "team lead" / "CG leader" (spec.md L95-L96), matching the code-layer D10 row ("`CG leader`, `team lead`"); AC-RND-021 (acceptance.md L138), design.md L33, progress.md L36, research.md L77-L80 all carry the same spellings; the withdrawn v0.3.0 spellings are recorded as withdrawn.
- N4 ANCHOR-SAME-FILE-BLIND — **RESOLVED**. `grep -c 'if f == path' anchors.py` → 0; the script runs against the edited-tree file list with the plan-time `headings.tsv` (`headings 133 referenced-elsewhere 11 total ref-files 63`). The self-count risk the docstring rules out holds: plan-time `headings.tsv` has 0 headings with an explicit `{#id}`.
- N5 HEADING-SPACING — **RESOLVED**. A blank line now precedes `### B.6` (spec.md, after the REQ-RND-017 bullet).

No regressions: the only changed normative lines are REQ-RND-020, REQ-RND-021 (table + note), AC-RND-004, 008, 020, 021, and each was checked above. Counts are consistent across artifacts (148 lines / 60 files in spec.md REQ-RND-020 and acceptance.md AC-RND-020).

## Defects Found (structured defect-list)

All remaining findings are minor and optional; none affects a must-pass criterion or makes a stated criterion wrong on a correct implementation.

M1. AC-004-COMPANION-NAMES — acceptance.md AC-RND-004 — the scoped extraction also returns Kanban companion and example session names (`--name run|plan|sync|mysession`, 92 matches). They are not lane labels, so each must be ledger-marked "other meaning" to pass; the exemption exists, so the criterion is satisfiable, but the ledger burden is avoidable. — Severity: minor — Class: optional — Required fix: restrict `--name` extraction to `-f`/`--factory` contexts, or state that companion names are exempt by class.

M2. GATE-QUALIFIER-CHECK-UNCARRIED — spec.md L98 says "The REQ-RND-002 gate compares these two spellings against the code-layer D10 row as landed", but REQ-RND-002 (spec.md L48) and AC-RND-001 (acceptance.md L13) check only the status and the legacy-spelling column. — Severity: minor — Class: optional — Required fix: add one AC-RND-001 row comparing the D10 qualifier spellings on develop with the REQ-RND-021 table.

M3. DESCRIPTIVE-SENSE-WORDING — spec.md L54 (REQ-RND-005) and L69 (REQ-RND-014), acceptance.md L91 (AC-RND-014) still name the sense "the `moai cg` leader pane" while the qualifier is now "CG leader". Descriptive, not a qualifier, so no criterion breaks. — Severity: minor — Class: optional — Required fix: align the wording.

M4. GATE-AC-OMITS-M3 — acceptance.md L11 (AC-RND-001) — Given lists "M4, M5, or M6", while REQ-RND-002 gates "any substitution milestone" and plan.md L5 gates M3-M6; the M3 HARD-amendment commit has no pass-path gate row (the halt path is covered by AC-RND-002). Pre-existing since iter-1 and missed in earlier iterations. — Severity: minor — Class: optional — Required fix: change the Given to "M3, M4, M5, or M6".

M5. MIRROR-ASSUMES-LINE-ALIGNMENT — `echoes.py` mirror pass (REQ-RND-020) maps locale copies by line number. Measured: every role-bearing echo page is line-aligned across en/ko/ja/zh (factory-mode 116, kanban-mode 321, manager-lead 200, moai-web-console 208, launchers 113, web 77, moai-gtd 59, moai-todo 264); only three `claude-code/**` pages differ (e.g. commands.md 262/262/203/203), where the mirrored rows are other-meaning. Holds today; a future page whose copies drift would get wrong mirror rows. — Severity: minor — Class: optional — Required fix: state the alignment assumption, or have the script flag a page whose four copies differ in line count.

## Recommendation

PASS at 0.91 (Tier L threshold 0.85). Must-pass evidence: MP-1 25 sequential REQs; MP-2 all REQs in GEARS forms on the requirement layer; MP-3 lint clean; MP-5/6/7 no D7 BLOCKING, no `syscall`, no clarification markers. All five iter-2 defects are resolved and verified by re-running `echoes.py` (byte-identical to the committed TSV, 148/60), the AC-020 group check (23 groups, 0 bad), the AC-004 extraction, and `anchors.py`. M1-M5 are optional and may be folded into M2 of the run phase at the orchestrator's discretion. The substitution milestones stay gated on SPEC-ROLE-NAMING-CODE-001 landing on develop (REQ-RND-002); this PASS does not relax the Implementation Kickoff Approval gate.
