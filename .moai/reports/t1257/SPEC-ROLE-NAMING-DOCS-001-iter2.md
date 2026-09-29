# SPEC Review Report: SPEC-ROLE-NAMING-DOCS-001
Iteration: 2/3 (Tier L ceiling 3 per `.moai/config/sections/harness.yaml`:78; PASS threshold 0.85)
Verdict: FAIL (score above threshold; two narrow blocking defects remain — see Recommendation)
Overall Score: 0.88 (iter-1 0.81 — no regression, no STOP signal)

Audited tree: worktree `.claude/worktrees/t1257`, branch `WT-role-naming-docs`, HEAD `f3f22b2f9` (v0.3.0). Scope: delta re-audit of iter-1 defects D1-D10 plus regression and new-defect scan. Every resolution below was checked against the files, not the HISTORY entry.
Reasoning context ignored per M1 Context Isolation. The operator's D1 decision (self-promoting lane performs every pre-dispatch obligation and reports before starting) is taken as binding; recorded at research.md:L76 and progress.md:L35.

## Must-Pass Results

- [PASS] MP-1: 25 entries REQ-RND-001..025, sequential (`grep -oE '^- \*\*REQ-RND-[0-9]+'` → 001..025). 25 is within the Tier L ceiling (25).
- [PASS] MP-2 (requirement layer only): REQ-003/016 relabelled Event-driven (spec.md:L48, L73), REQ-010 recast as `When` (L61), REQ-017 now two `shall not` clauses (L74), REQ-018 Ubiquitous with sub-items (L77-L82), REQ-025 Ubiquitous (L101). ACs remain Given-When-Then.
- [PASS] MP-3: frontmatter unchanged apart from version "0.3.0"; `moai spec lint` → `0 error(s), 0 warning(s)`.
- [N/A] MP-4: not multi-language tooling.
- [PASS] MP-5 D7: no referenced SPEC is retired/superseded/archived; SPEC-ROLE-NAMING-CODE-001 still absent on this branch (SHOULD, gated by REQ-RND-002).
- [PASS] MP-6 D8: `syscall` count 0.
- [PASS] MP-7: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no output (exit 1).

## Category Scores

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.90 | 0.75-1.0 | REQ-RND-018 (a)-(e) names one subject per obligation — "the dispatching party (the leader, or a lane that promoted the card itself)" — for L37/L39/L41/L49 and the report-before-work duty (spec.md:L77-L82); REQ-RND-021 tables three senses × four locales and excludes plain English and identifiers (L88-L94). |
| Completeness | 0.85 | 0.75-1.0 | All sections present; echo set widened to 124 lines / 51 files and reproduced byte-identically, but still misses five direct echoes (N1). |
| Testability | 0.80 | 0.75 | AC-RND-018 now mutant-resistant (phrase count ≥ 4 + subject per obligation, acceptance.md:L111-L120); AC-RND-008 invocation stated (L53-L55). AC-RND-004's second clause is false at arrival on correct work (N2). |
| Traceability | 1.0 | 1.0 | 25 AC headings, 25 table rows REQ→AC 1:1 (acceptance.md §B), REQ-025 → AC-025 (L158-L162). |

Aggregate (harmonic mean): 0.88.

## Regression Check (iter-1 defects)

- D1 SELF-PROMOTION-ACTOR — **RESOLVED**. spec.md:L79-L80 set the subject of the PR/landed cross-check (L37), completed-SPEC cross-check (L39), confirm/withdraw surfacing (L41), and class assignment (L49) to "the dispatching party (the leader, or a lane that promoted the card itself)" and require the lane to perform them and report each result to the leader before starting. AC-RND-018 checks the phrase count (≥ 4 per copy) and grammatical subject per obligation. Line numbers verified: both copies are byte-identical (`cmp` exit 0) and carry those clauses at L31/L33/L37/L39/L41/L49/L266.
- D2 ECHO-LIST-INCOMPLETE — **RESOLVED as specified, residual reported as N1**. All three named items are now in the set (`README{,.ko,.ja,.zh}.md` L161, `kanban-dispatch.md` L33 and L266 in both copies); REQ-RND-020 is now "script output + run-time re-run" (spec.md:L84); AC-RND-020 re-runs the script (acceptance.md:L130-L132). Re-run evidence: `python3 .moai/reports/t1257/raw/scripts/echoes.py` → stderr `# total lines 124, files 51`, exit 0; `cmp` against `raw/q3-clause-echoes.tsv` → exit 0.
- D3 LEADER-SENSE-QUALIFIERS — **RESOLVED**. Three senses × en/ko/ja/zh tabled (spec.md:L90-L94); plain English and identifiers excluded (L88); ja/zh fixed now with rationale and measurement (research.md:L77-L81). AC-RND-021 quotes the table (acceptance.md:L136-L138).
- D4 DOCS-CANONICAL-ORDER — **RESOLVED**. plan.md:L76 ko-first, citing `docs-site-i18n-rules.md` L51.
- D5 HARD-COUNT-TENSION — **RESOLVED**. REQ-RND-007 (spec.md:L55) and AC-RND-007 (acceptance.md:L47-L49) split equal vs not-less-than.
- D6 ANCHOR-SCAN-INVOCATION — **RESOLVED**. AC-RND-008 copies plan-time `headings.tsv` into `$R/raw-anchor-check/` and runs `anchors.py` against the edited tree's `files.txt` (acceptance.md:L53-L55). Residual (optional N4): `anchors.py` skips references inside the heading's own file (`if f == path: continue`), so a same-file reference to a renamed heading is not counted.
- D7 CLI-ACCEPTANCE-UNMEASURED — **RESOLVED in form, new defect N2**. A command now exists (acceptance.md:L31), but its output cannot pass on a correct implementation.
- D8 GEARS-LABELS — **RESOLVED** (see MP-2).
- D9 LEDGER-PATH-IGNORED — **RESOLVED**. REQ-RND-006 (spec.md:L54) and AC-RND-006 (acceptance.md:L43) state `git add -f` and check `git ls-files $R/`.
- D10 — no action required; unchanged.

No regressions found in unchanged REQs. `raw/q3-clause-echoes.txt` was deleted; no artifact still references it (grep of all six artifacts: only the v0.3.0 `.tsv` is cited).

## Defects Found (structured defect-list)

N1. ECHO-PATTERN-ASYMMETRY — spec.md:L84 (REQ-RND-020 "the echo set is closed by measurement"), `.moai/reports/t1257/raw/scripts/echoes.py`:L20-L29 — The patterns differ by locale, so the same sentence is caught in one locale and missed in its siblings. Measured misses (probe `grep -rnE "one who picks|who picks|고르는 주체|picks? (a|the) card|…"` then join against the re-run TSV):
  - `docs-site/content/en/utility-commands/moai-todo.md`:143 "**A human does the picking.**", `…/ko/…/moai-todo.md`:143 "**고르는 주체는 사람입니다.**", `…/zh/…/moai-todo.md`:143 "**挑选的主体是人。**" — only the ja line 143 is in the set (the ja pattern has `選ぶ主体`; en/ko/zh have no equivalent). With lane self-promotion these sentences are false.
  - `docs-site/content/en/advanced/factory-mode.md`:113 "The operator is the one who picks" and `…/ko/advanced/factory-mode.md`:113 "카드를 고르는 주체는 운영자" — ja/zh line 113 are in the set, en/ko are not.
  Because AC-RND-020's post-edit check reads only lines the same script prints, these five lines pass AC-RND-020 unamended while contradicting the amended promotion clause. Page-level parity (AC-RND-013) does not force the line change. — Severity: major — Class: blocking — Required fix: add en `human does the picking|the one who picks|who picks`, ko `고르는 주체`, zh `挑选的主体` (and any other locale-mirror of an existing pattern) to `echoes.py`, re-run, commit the new TSV, and update the counts in spec.md:L84, acceptance.md:L130, plan.md:L9, research.md:L72. A cheap guard: for every page path with an echo in one locale, check the same line region in the other three.

N2. AC-004-FALSE-AT-ARRIVAL — acceptance.md:L31 (AC-RND-004 second clause) — "every form printed by `grep -rhoE -- '-f [a-z]+(-<n>|-[0-9]+)?' <in-scope doc surfaces> | sort -u` appears in the join-token or numbered-label row" of the code-layer table. Measured on the current tree over docs-site, READMEs, templates, `.claude`, CLAUDE.md, AGENTS.md: forms include `-f https` (curl, `moai-foundation-cc/reference/examples.md`:434), `-f docs` / `-f memory` (`test -f`), `-f session` / `-f lead` / `-f factory` / `-f entry` (prose "-k/-f lead session", "-f factory lead", "-f entry switch"). None of these is a factory join form, none will be in the code-layer table, and this clause (unlike the first clause) has no ledger "other meaning" exception — so a correct implementation fails it. Also, the rows in the code-layer design (`WT-role-naming-code` HEAD `d0770b9cc`, design.md table L30-L33) are titled "`-f <role>` join token" and "`-f <label>` / `--name <label>` (lane)", not "numbered-label". — Severity: major — Class: blocking — Required fix: restrict the extraction to launcher invocations (e.g. `grep -rhoE -- 'moai (cc|glm|codex)[^`]* -f [a-z]+(-<n>|-[0-9]+)?'`) or add "excluding matches the ledger marks other meaning (shell `-f` flags, `-k/-f` prose)"; name the table rows by their actual titles.

N3. QUALIFIER-CROSS-SPEC-DRIFT — spec.md:L94 vs `WT-role-naming-code:.moai/specs/SPEC-ROLE-NAMING-CODE-001/design.md` D10 — the code layer writes the homonym qualifiers as `CG leader` and `team lead`; this SPEC tables `cg leader pane`. User-facing strings from the code layer and the docs would then name the same sense two ways. — Severity: minor — Class: optional — Required fix: align with the code layer (or note in REQ-RND-002's gate check that the qualifier spelling is compared too).

N4. ANCHOR-SAME-FILE-BLIND — acceptance.md:L55 with `raw/scripts/anchors.py` (`if f == path: continue`) — same-file references to a renamed heading are not counted. — Severity: minor — Class: optional — Required fix: drop the same-file skip for the AC-RND-008 run, or add a per-file `§ <old text>` grep.

N5. HEADING-SPACING — spec.md:L74-L75 — the REQ-RND-017 bullet is followed directly by `### B.6` with no blank line (renders, but diverges from the file's own spacing). — Severity: minor — Class: optional — Required fix: insert a blank line.

## Recommendation

Verdict FAIL at 0.88: the rubric score clears the Tier L threshold and every must-pass criterion passes, and all ten iter-1 defects were addressed as instructed — but N1 and N2 are blocking because each makes a stated criterion wrong on a correct implementation (N1: AC-RND-020 passes while five contradicting sentences survive; N2: AC-RND-004 fails on shell `-f` flags that must stay).

Both fixes are mechanical and small:
1. N1 — add the missing locale-mirror patterns to `echoes.py`, re-run, commit the TSV, update the four count citations.
2. N2 — scope the `-f` extraction to launcher invocations (or add the ledger exception) and cite the code-layer table rows by their real titles.

Iteration count: `harness.yaml`:78 sets the Tier L ceiling at 3, so an iteration-3 delta audit (N1/N2 only) is still within the contract. If the orchestrator nonetheless escalates now, PASS-with-debt is defensible on this evidence provided N1 and N2 are carried as explicit M2 tasks (M2 already re-runs `echoes.py` and seeds the ledger, plan.md:L50) — that choice belongs to the operator, not to this report.
