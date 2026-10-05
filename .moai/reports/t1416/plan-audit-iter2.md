auditor-model: claude-sonnet-5-5[1m]

verdict: FAIL
audited_sha: d7e97e0f4b26e1f2f9d8e72561d3cba0982ab098

# SPEC Review Report: SPEC-CC-ULTRACODE-TOGGLE-001
Iteration: 2/3
Verdict: FAIL (score 0.86; no must-pass failure, aggregate above the Tier M threshold 0.80; failing on one unresolved prior-iteration defect, D2(a), per the retry-loop rule "unresolved defects from a prior iteration are automatically FAIL")
Overall Score: 0.86 (iteration 1: 0.75; no regression)
Plan Artifact Hash (per file, sha256, convenience only; the Go `ComputeHash` was not run): spec.md `90fd3143e226e5e92afea3a9a9c32d16ea921c2fc649f580b1aa78f25e56cef4`, plan.md `be5bfc7a36b5940930c5e491d20af849163e1c386bb2dda281dd655c95df6837`, acceptance.md `68ae306f6a6112ca5da74d7e29e39cd9d98e6e13bf606d639fabfe1fb0fab207`, progress.md `af9a4191b366c885bbb7df20ffcfe787c74885ad16e597377f14f6dc24375dfd`
Auditor Version: plan-auditor (Sonnet 5.5 session)

Reasoning context ignored per M1 Context Isolation. Read only the committed artifacts at `d7e97e0f4` (spec, plan, acceptance, progress) plus the iteration-1 verdict file named by the caller as the defect baseline.

## Claim

C1. Ten of the eleven iteration-1 defects are fixed (D1, D3-D11). Supported by the re-run evidence below.
C2. D2 is only partly fixed: the positive pin half is fixed; the "wording-independent" coupling regex R1 is a closed verb list, and realistic rewordings pass it while keeping the xhigh coupling. NOT fixed.
C3. All RED-now cells in the evidence ledger reproduce on this tree (counts and exit codes). Supported.
C4. Upstream facts F1-F7 and DEC-1 match raw page text. Supported.
C5. All nine must-pass criteria pass. Supported.

## Evidence

All commands run from `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1416`; `grep` on this machine is `ugrep`, which prints multi-file `-c` output in a nondeterministic file order (counts per path are stable).

### E1. Tree attribution and scope

```
git rev-parse HEAD                      -> d7e97e0f4b26e1f2f9d8e72561d3cba0982ab098
git status --short                      -> (empty)
git diff --stat c50da9c2f d7e97e0f4     -> 5 files changed, 774 insertions(+): .moai/reports/t1416/plan-audit.md,
                                           .moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/{acceptance,plan,progress,spec}.md   (exit 0)
git diff --stat c50da9c2f HEAD -- .claude internal docs-site  -> (empty, exit 0)
```
Nothing outside `.moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/` and `.moai/reports/t1416/` changed; the scope files are byte-identical to the pin `c50da9c2f`, so the ledger's tree `0e7b6af5b` measures the same scope content as the audited HEAD.

### E2. RED-now re-execution (this run, tree d7e97e0f4; ledger cells are inherited from `0e7b6af5b`, document-level pin)

Rule source RS / mirror RM (count per file, exit):
```
R1  coupling regex                      RS:1 RM:1  exit 0     (ledger: 1,1  exit 0)  match
R1b 'combines `xhigh` reasoning'        RS:1 RM:1  exit 0     match
R2  'leaves the effort level unchanged' RS:0 RM:0  exit 1     match
R3  'independent on/off toggle'         0,0 exit 1            match
R4  'v2.1.284'                          0,0 exit 1            match
R5  '/effort ultracode off'             0,0 exit 1            match
R6  '--effort ultracode'                0,0 exit 1            match
R7  'starts the session at `xhigh`'     0,0 exit 1            match
R8  '"ultracode": true'                 0,0 exit 1            match
R9  'ultrathink.'                       1,1 exit 0            match
R10 'current session'                   0,0 exit 1            match
R11 v2.1.284 + key on one line          0,0 exit 1            match
R12 'step back with `/effort high`'     1,1 exit 0            match
R13 slider                              0,0 exit 1            match
X-cmp RS RM                             empty, exit 0         match
```
Docs-site (per locale; the ledger's ko,en,ja,zh order differs from ugrep's print order, counts match path by path):
```
W1 --effort ultracode[^|]*xhigh   0,0,0,0 exit 1     W2 "ultracode": true   0,0,0,0 exit 1
W3 v2.1.284 + key one line        0,0,0,0 exit 1     W4 /effort ultracode off  0,0,0,0 exit 1
W5 xhigh.*xhigh|xhigh.*--effort   0,0,0,0 exit 1     W6 ultracode..effort high 1,1,1,1 exit 0
W7 row shape                      1,1,1,1 exit 0     W8 v2.1.284  0,0,0,0 exit 1     W9 (retired) 1,1,1,1 exit 0
W10 ko/en/ja/zh scope literal     1,1,1,1            W11 slider words, 16 files: all 0, exit 1
M1 comment xhigh 1,1,1,1 exit 0  M2 control 1,1,1,1 exit 0
U1 effects bullet xhigh 1,1,1,1 exit 0   U2 '/effort high' ko:1 en/ja/zh:0 exit 0   U3 '/effort ultracode off' 0,0,0,0 exit 1
U4 ko 1   U5 ko 1
C1en/ko/ja/zh 1 each   C2en 1  C2ko 1   C3-en/ko/ja/zh 0 each (exit 1)
H-WF 11/11/11/11  H-ML 13/13/13/13  H-UW 24/20/20/20  H-CM 23/23/18/18
X-baseline-diff: git diff --stat -- docs-site/.locale-parity-baseline -> empty, exit 0
```
Hugo control (AC-009): `hugo --minify --gc --source <docs-site> --destination <scratch>` -> exit 0; `grep -c -E 'WARN|ERROR'` of its log -> `0` (exit 1); `git status --short` afterwards empty. URL-blacklist grep exit 1 (no output); Mermaid LR/RL grep exit 1 (no output). D11's claim reproduces.

### E3. Mutant probes (scratch copies under the session scratchpad; the worktree was not touched)

Rule-source mutants are built from the real L111 bullet with the old coupling sentences removed and every positive pin R2-R11 present (flag clause "also starts the session at `xhigh`"), differing only in one extra sentence:

| Mutant (extra sentence) | R1 (want 0) | R1b | R12 | R7 | result vs AC-001/AC-002 |
|---|---|---|---|---|---|
| A "Ultracode pairs with `xhigh` reasoning and automatic orchestration." | 0 | 0 | 0 | 1 | PASSES wrongly |
| B "Turning ultracode on lifts reasoning effort to `xhigh` for every task." | 0 | 0 | 0 | 1 | PASSES wrongly |
| C "The mode defaults to `xhigh` effort while it is on." | 0 | 0 | 0 | 1 | PASSES wrongly |
| E "Ultracode is best used at `xhigh`; the mode stays on at that level." | 0 | 0 | 0 | 1 | PASSES wrongly |
| F "Ultracode forces `xhigh`." | 1 | 0 | 0 | 1 | caught by R1 |
| D (clean, no extra coupling) | 0 | 0 | 0 | 1 | correct passes (feasibility control) |

A, B, C, E keep a second `xhigh` coupling in the bullet, violating REQ-001 ("shall not couple ultracode to `xhigh` in any wording") and REQ-002 ("That clause shall be the only place in the bullet where `xhigh` appears"). Their `xhigh` count on the bullet line is 2 (`grep -c -E 'xhigh.*xhigh'` -> 1 for A, B, C, E; 0 for the clean D and 0 on the real file today), so one extra pin separates them.
False-positive control: a correct line "key sets it for every session ... the `--effort ultracode` launch flag also starts the session at `xhigh`" -> R1 count 1 (R1 fires on the word "sets" anywhere between the first `ultracode` and the last `xhigh` on the one-line bullet).

Docs-row mutants (en row, all other pins satisfied):

| Mutant | W1 (want >=1) | W5 (want 0) | W6 (want 0) | result |
|---|---|---|---|---|
| G "...; `--effort ultracode` to launch; ultracode forces `xhigh` at all times" (single xhigh after the flag literal) | 1 | 0 | 0 | PASSES wrongly |
| H flag mentioned in one sentence, "Ultracode always runs at `xhigh`." in the next | 1 | 0 | 0 | PASSES wrongly |
| I correct row (flag clause carries the only xhigh) | 1 | 0 | 0 | correct passes |

Other mutants: U1 relabel ("- Reasoning level: set to `xhigh`") and U1 prose form ("Reasoning effort is raised to `xhigh` ...") both print 0 and pass `[U1]` while keeping the claim. C1-style reword ("also an `/effort` level (a toggle)") prints C1en 0, C3en 1 and passes while still calling it a level. The delete-only, over-correction (R7/W1), one-side (X-cmp), slider-word and unqualified-settings-key mutants listed by the SPEC are each caught by the pin the SPEC names (R7/W1, X-cmp, R13/W11, R11/W3 are distinct from A-H above; R11/W3 are trivially satisfied on a one-line bullet/row by any text containing both literals).

### E4. Upstream facts (raw page text via `python3 urllib` fetch with tag stripping; `WebFetch` not used; 2026-10-02)

- F1 changelog: `## 2.1.284` (line 342), line 406: "Changed Ultracode into its own toggle in `/effort` (Tab, or `/effort ultracode [on|off]`): it no longer forces xhigh effort and stays on at any effort level". Latest entry is 2.1.287; no later entry changes ultracode semantics.
- F2/F3/F4/F6 model-config: "Ultracode is a Claude Code setting rather than a model effort level ... at whichever effort level the session runs at." "Turning ultracode on or off with /effort or the ultracode setting leaves the effort level unchanged. The --effort ultracode flag and the Agent SDK effortLevel: "ultracode" value turn it on and also set the level to xhigh. Picking a level in the /effort slider or the /model picker leaves ultracode as it was." "The /effort ultracode off form, the slider toggle, and keeping ultracode on at effort levels other than xhigh require Claude Code v2.1.284 or later. Before v2.1.284, turning on ultracode set the session to xhigh effort, picking another level turned it off ..." Also "For where it can be set persistently, see the ultracode setting" and "ultracode has its own ultracode key". Match.
- F5 workflows page: "/effort ultracode lasts for the current session; to have every session start with it, set the ultracode setting. Turn it off with /effort ultracode off ..." Match.
- F7 settings-reference: "The key doesn't change the session's effort level: ultracode runs at whichever level the session uses. Claude Code reads this key but never writes it: /effort ultracode turns ultracode on for the current session only." "This and the /effort ultracode off form require Claude Code v2.1.284 or later. Before v2.1.284, ultracode: true ran the session at xhigh effort, and an effort cap below xhigh kept ultracode off." "The --effort ultracode flag also turns it on for one session, at xhigh effort, and requires Claude Code v2.1.203 or later". Match.
- DEC-1 and the OQ-1 parenthetical (slider Enter on an effort level saves a default under `modelSettings`; model-config line "Enter in the /effort slider or the /model picker, or a level typed after /effort: save the level as your default") match. One precision note: the page's "This ... require v2.1.284" follows the effort-cap sentence; the key itself existed earlier and ran at `xhigh`. DEC-1 words it correctly ("before v2.1.284 the key ran sessions at `xhigh`"); the SPEC does not mandate the stronger "the key requires v2.1.284".

### E5. Structural verbs

- Traceability: the verbatim `awk` verb was not run (worktree guard refuses awk programs; see Gaps); a python port of the same three definition regexes and the cell/list shorthand rule was run (`trace.py`, scratchpad). `COLLECTED: 12 REQ definitions (acceptance input: read)`, `UNCOVERED: none`, `ORPHAN: none`. REQ to AC headings: REQ-001 AC-001; REQ-002 AC-002; REQ-003 AC-002; REQ-004 AC-001, AC-003, AC-011; REQ-005 AC-004; REQ-006 AC-005; REQ-007 AC-006; REQ-008 AC-007; REQ-009 AC-004, AC-008; REQ-010 AC-008, AC-009; REQ-011 AC-004, AC-010; REQ-012 AC-010. The REQ definitions are now `- **REQ-NNN**:` and are collected (iteration 1 collected 0).
- CN-4 ordering (python/grep port; awk refused): `COLLECTED: 4 milestones in plan order (M1, M2, M3, M4), 0 exit bindings, 7 ordering candidates`. Candidates (acceptance.md L3, L23, L33, L38, L56, L60, L64) are "before/after/first" words describing pin values or regex shape, none names a milestone on a forbidden side; no `CONFLICT:`. M1 (canonical wording) -> M2 (ko, then en, ja, zh) -> M3 (mirror) -> M4 (verification) is consistent with AC-003 "stay exit 0 after the edit" (the plan edits RS and RM in the same step) and AC-011. No `Exit:` bindings exist, so the mechanical check observes nothing to conflict; read by hand: no conflict.
- `moai spec lint SPEC-CC-ULTRACODE-TOGGLE-001`: installed `moai` (`moai_cp/20260925_122548-1711-gd194083fb`, built 2026-09-30T08:37:30Z; `git merge-base --is-ancestor d194083fb HEAD` -> 0, a strict ancestor of HEAD, so it lags) -> `No findings`, exit 0. A binary built from this tree (`go build -o <scratch>/moai-tree ./cmd/moai` at HEAD d7e97e0f4, exit 0; it prints `v3.1.3 none` because no ldflags were passed) -> `No findings`, exit 0. Positive control: a scratch copy with an extra uncovered `REQ-013` linted with the tree-built binary -> 0 errors, 0 warnings. The linter is therefore silent on traceability for this SPEC form and is NOT cited as corroboration of AC-4/AC-5; the traceability reading above is the auditor's own.
- D7: spec.md references only its own id; plan.md also cites `SPEC-CC-GD124-001` (status `completed`, not retired). D8: `grep -c syscall spec.md` -> 0, exit 1. MP-7: `grep -rn '\[NEEDS CLARIFICATION' plan.md` -> no output, exit 1; research.md absent.

## Baseline-attribution

- Tree measured: worktree `.moai/worktrees/t1416`, HEAD `d7e97e0f4b26e1f2f9d8e72561d3cba0982ab098`, clean, in this run. The ledger's RED cells are pinned to `0e7b6af5b`; scope files are identical to `c50da9c2f`, `0e7b6af5b` and HEAD (E1), so the ledger pin names a tree whose scope content equals the audited one.
- Judging builds: installed `moai` build `d194083fb` (strict ancestor of HEAD, lagging) and a tree-built binary from HEAD (clean tree, `go build` without ldflags, no commit stamp). Per `verification-claim-integrity.md` section 2.2 both lint results are attributed to those builds and neither is cited as corroboration (positive control silent).
- Upstream pages were fetched raw in this run; no figure is carried over from iteration 1.

## Gaps

1. The verbatim traceability and CN-4 `awk` verbs were not run: the worktree guard refuses awk programs (iteration 1 saw the same). Python/grep ports were used; they implement the verbs' regexes and are not the verbs' own output.
2. One compound command mixing `git diff` with other commands was refused by the worktree guard and re-issued as separate plain commands. One zsh `$W` word-splitting error produced no data and was rerun with literal paths. Nothing was substituted by reading.
3. AC-011 (`make build`, `go test ./internal/template/...`) not executed (the mirror is unedited; scoped verification).
4. OQ-1 (slider toggle persistence) not observed live. Wording quality and native idiom of the not-yet-written ko/en/ja/zh text are unobservable at plan time.
5. No cross-backend audit tool (`audit_multi`, `codex_audit`, `glm_audit`, `claude_audit`) was invoked; the caller did not request one. No audit receipt exists.
6. Ledger stdout is listed ko,en,ja,zh while `ugrep` prints in nondeterministic order; the counts match, the byte order is not reproducible.

## Residual-risk

- Pins are string-presence checks on prose; they cannot prove a locale's sentence reads correctly or states all of F2-F7.
- Even after fixing D2 the R1 verb list remains a blacklist; the count pin recommended below makes the rule-source guard structural.
- The settings-key sentence depends on upstream still saying what F7 says; a later upstream change would age the text (the SPEC pins v2.1.284 only as the introduction version).
- The slider's persistence (OQ-1) remains open; REQ-003 (see N2) states a session scope for "a toggle made through `/effort`", wider than the evidence.

## Must-Pass Results

- [PASS] MP-1 REQ numbering: `- **REQ-001**:` .. `- **REQ-012**:` at spec.md L53-L64, sequential, no gaps or duplicates, uniform zero padding (python collection: 12 ids).
- [PASS] MP-2 GEARS (layer judged: requirement layer, `REQ-XXX` in spec.md; ACs are verification-layer and were graded under testability): REQ-001..008 are ubiquitous "shall" statements (D9 fixed: REQ-002/003 no longer wear a trigger), REQ-009 "**When** the docs-site edits are made, ko shall be authored first ..." is an event-driven form, REQ-010..012 are legacy "shall not" negatives (allowed at 1.0, never canonical).
- [PASS] MP-3 frontmatter: id L2, title quoted L3, version `"0.1.1"` quoted L4, status `draft` L5, created/updated `2026-10-02` L6-L7, author L8, priority `P2` L9, phase `"v3.2.0 target"` (release label, not a lifecycle token) L10, module L11, lifecycle `spec-anchored` L12, tags comma string L13, optional `tier: M` L14; no snake_case alias; plan.md/acceptance.md/progress.md carry no `status:` frontmatter. Lint clean on two builds (not cited as corroboration).
- [N/A] MP-4: prose correction of a Claude Code feature description; no language-specific tool is named or required.
- [PASS] MP-5 D7: only self-reference in spec.md; plan.md's `SPEC-CC-GD124-001` is `completed`, not retired/superseded.
- [PASS] MP-6 D8: `syscall` appears 0 times in spec.md (exit 1).
- [PASS] MP-7: no `[NEEDS CLARIFICATION` marker in plan.md; research.md absent (Tier M), that half N/A.
- [PASS] MP-8: release-blocking ACs are now declared (AC-001, AC-002, AC-004, AC-005, AC-006, AC-007; acceptance.md L21, L26, L35, L41, L45, L50). Each cell carries command, stdout, exit code and a pinned tree SHA (document-level pin `0e7b6af5b`, acceptance.md L3-L5, ledger L86-L378). Every cited command is a single read-only invocation with quoted metacharacters only. All release-blocking RED cells were re-executed on this tree and reproduce (E2): positives 0 with exit 1, negatives non-zero with exit 0, controls unchanged. The `grep -c` runs executed real files (counts are the files' own matches, not an empty swept set).
- [PASS] MP-9: CN-4 verb equivalent: 4 milestones, 0 exit bindings, 7 non-binding candidates, 0 `CONFLICT:`; M1-M4 order consistent with the DoD and every ordering clause (E5).

## Per-defect disposition of iteration 1

| ID | Disposition | Evidence |
|---|---|---|
| D1 AC-004 per-locale pins | FIXED | W1, W2, W3, W4, W8 (positive, RED 0,0,0,0 reproduced, exit 1) and W10 scope controls added for all four locales (acceptance.md L37-L39). Residual placement weakness (mutants G, H) is filed as N3, not a reopening of D1's stated fix. |
| D2 AC-001 wording / phrase alignment | PARTIAL | (b) positive pin phrase is now identical to REQ-001 (`leaves the effort level unchanged`, backticked at spec.md L53; R2 pin). (a) NOT fixed: R1 is a closed verb list (forces, forced, combines, sets, raises, runs at, implies) called "wording-independent" at acceptance.md L24; mutants A, B, C, E (pairs with / lifts / defaults to / best used at `xhigh`) pass AC-001+AC-002 while violating REQ-001 and REQ-002; a correct sentence containing "sets" trips R1 (E3). |
| D3 AC-007 commands page | FIXED | Vacuous `ultracode` pin replaced by C3-en/ko/ja/zh (RED 0, exit 1, reproduced); C2en/C2ko zero pins for L137 added (RED 1, reproduced); C1 zero pins retained. |
| D4 settings-reference facts / REQ-011 | FIXED | OQ-2 retired into DEC-1 (spec.md L40-L42) and matches the settings-reference text (E4); OQ-1 now cites "reads this key but never writes it" (L46); REQ-011 decided (L63); F7 added (L38). Pins R11/W3 exist. |
| D5 RED cells | FIXED | Ledger gives command, stdout, exit code per cell (acceptance.md L86-L378), document-level SHA pin, literal paths in every command, per-AC class stated, abbreviations resolved via the Path ledger. All cells reproduce (E2). `ko-UW` is defined in the Path ledger though no ledger command uses it (harmless). |
| D6 plan cross-references | FIXED | plan.md L19 "AC-001..AC-008 cells", L64 "AC-008" and "(the hugo/URL/Mermaid gate itself is AC-009)". |
| D7 AC/REQ scope | FIXED | REQ-007 now requires `/effort ultracode off` (spec.md L59, matches U3); slider check is a command (R13, W11, 16 files, reproduced 0); AC-011 cites REQ-004 (acceptance.md L67). REQ-012's line-level half is honestly stated as read-by-diff (L65). |
| D8 REQ definition form | FIXED | `- **REQ-NNN**:` form; traceability port collected 12/12, 0 uncovered, 0 orphan. |
| D9 REQ-002/003 GEARS shape | FIXED | spec.md L54-L55 are plain ubiquitous statements. |
| D10 accuracy notes | FIXED | REQ-010 names both divergence-listed pages (L62); `session-handoff-appendix.md` pair listed under "remain accurate" (L86); strategy-doc notes match the files read (harness-delivery header "작성: 2026-06-03", autonomous-workflow header "전략 제안", L574 cost-risk row). |
| D11 unobserved hugo claim | FIXED | AC-009 records the observed values (L60); reproduced: hugo exit 0, 0 WARN/ERROR, URL and Mermaid greps empty. |

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.85 | between 0.75 and 1.0 | REQs carry verbatim backticked phrases and exact lines (spec.md L53-L64, change map L73-L79). REQ-003 "a toggle made through `/effort`" is wider than F2/F5/F7 (N2); REQ-005 is one very long compound sentence. |
| Completeness | 1.00 | 1.0 | HISTORY L19-L22, Context/WHY L24, Requirements L51, Non-Goals L66, Change Map L71, Out of Scope with three `### Out of Scope - <topic>` H3 sections and specific bullets L81-L99, plan, acceptance, progress, 12 frontmatter fields. |
| Testability | 0.70 | between 0.50 and 0.75 | AC-001 R1 admits four realistic wrong rewordings and false-fires on a correct line (E3); U1/M1/C1 pins are keyed to the current wording of one sentence each. AC-004/005/007/008-011 are binary. |
| Traceability | 0.95 | between 0.75 and 1.0 | 12 collected, 0 uncovered, 0 orphan; every REQ has a pin; REQ-012 line-level scope is not a command (stated openly, L65). |

Aggregate 0.86 against the Tier M threshold 0.80: above threshold. Verdict is FAIL for the unresolved D2(a) alone.

## Defects Found (structured defect-list)

D2 (carried, unresolved part) / N1. AC-001 R1 — acceptance.md:L23-L24 — Claimed wording-independent, but R1 is a verb blacklist. Realistic partial edits of the old sentence ("combines `xhigh` reasoning" -> "pairs with / uses / lifts reasoning to / defaults to `xhigh`") pass AC-001 and AC-002 while violating REQ-001 ("in any wording") and REQ-002 ("only place in the bullet where `xhigh` appears"); R1 also false-fires on any "sets/implies/raises" between the first `ultracode` and last `xhigh` of the one-line bullet. By `verification-completeness.md` section 2 (mutant probe) the criterion is too shallow to adopt as written. — Severity: major — Class: blocking — Required fix: add a count pin that pins REQ-002 structurally, for example `grep -c -E 'xhigh.*xhigh' .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` must be `0,0` (today `0,0` by this audit; together with R7 >= 1 it forces exactly one `xhigh`, inside the flag clause), and either drop R1 or restate AC-001's claim as "verb-list heuristic", removing "wording-independent". Re-run the mutant probe (A, B, C, E, plus the "sets" false-positive line) and record the values.

N2. REQ-003 vs OQ-1 — spec.md:L55, L46 — REQ-003 states that "a toggle made through `/effort` lasts for the `current session`". The `/effort` slider toggle is a toggle made through `/effort`, and its persistence is the SPEC's own open question OQ-1; F2/F5/F7 only support the command form `/effort ultracode` (model-config also says "For where it can be set persistently, see the ultracode setting", which supports but does not settle it). — Severity: minor — Class: optional — Required fix: name the command form (`/effort ultracode`) in REQ-003, or leave as is and note the dependency on OQ-1.

N3. AC-004 / AC-006 / AC-007 placement and label pins — acceptance.md:L37-L38, L47, L52 — W1 + W5 enforce exactly one `xhigh` per row, located after the first `--effort ultracode` literal, but not "inside the flag sentence": mutants G and H keep an "always at `xhigh`" claim after the flag literal and pass every AC-004 pin. U1 (`^- (Reasoning effort|추론 깊이).*xhigh`) is keyed to the present label; a relabelled bullet keeps `xhigh` and passes. C1 pins only the literal old sentence. These need a deliberate or unusual edit, unlike N1. — Severity: minor — Class: optional — Required fix: tighten W1 to a same-sentence bound (for example `--effort ultracode[^|.。]*xhigh`, valid because the flag clause contains no period) and make U1 label-independent (`^- .*xhigh`, zero today in the four UW pages outside the effects bullet; verify before adopting).

N4. Vacuity sentence — acceptance.md:L79 — says every zero pin is at least 1 today; W5 (and the other regression-guard-style zero pins inside release-blocking ACs, R13, W11) are 0 today by design. — Severity: minor — Class: optional — Required fix: say "every positive pin is 0 today; zero pins that are 0 today are regression-guards".

N5. Ledger size — acceptance.md:L86-L378 — the evidence ledger is 293 of 378 lines. It is the carrier the doctrine recommends (`verification-completeness.md` section 2.1), sits after the criteria, and is addressed by pin id, so navigation holds; the multi-file blocks repeat four-path commands. Not a defect against any stated criterion. — Severity: minor — Class: optional — Required fix: none required; if the document is touched again, compress the repeated per-locale blocks to one command plus a four-count line.

(Answers to the four reviewer questions.)
(i) AC-004 W1 + W5 enforce "exactly one `xhigh` in the row, positioned after the first `--effort ultracode`" and nothing about sentence membership (N3, mutants G, H); the natural partial edits (keep the old leading "`xhigh` ..." opening, drop the flag clause) are caught by W5 and W1 respectively.
(ii) English-only R1 for the rule source is NOT sufficient (N1, mutants A, B, C, E); the language-neutral structural pins used for the locales (W1/W5 count and order, M1 `.*xhigh`) are stronger than R1, and the same count idea (`xhigh.*xhigh` = 0) fixes the rule source without any verb list. The structural pins for the locales are sufficient against natural partial edits and weak only against deliberate relocation (N3).
(iii) REQ-011 forbidding any slider mention is stricter than the card needs (no unverified claim is possible without a slider sentence, and F3's "picking a level in the slider leaves ultracode as it was" is a verified fact the pages could carry), but it is cheap, pinned (R13/W11 reproduce 0 over 18 files), and consistent with the open OQ-1: acceptable, optional.
(iv) The 293-line ledger is a minor note, not a defect (N5).

## Regression Check (Iteration 2+)

Defects from iteration 1: D1 FIXED, D2 PARTIAL (unresolved: coupling regex), D3 FIXED, D4 FIXED, D5 FIXED, D6 FIXED, D7 FIXED, D8 FIXED, D9 FIXED, D10 FIXED, D11 FIXED (table above). No score regression (0.75 -> 0.86), so no STOP signal. The CN-4 ordering verb equivalent was re-run in full: no conflict.

## Recommendation

FAIL (0.86): fix one blocking item and re-audit; the re-audit is scoped to D2 plus the CN-4 ordering check in full.

1. Add the `xhigh.*xhigh` count pin on RS and RM (expected 0,0 now and after), keep R7 as the presence pin, and drop the "wording-independent" claim from AC-001 or keep R1 only as a heuristic. Record the mutant results (A, B, C, E now fail; the correct line D passes).
2. Optional, same edit round at no extra cost: N3 tighter W1/U1, N2 REQ-003 command-form wording, N4 vacuity sentence.
3. Do not add pins for the remaining optional items beyond that; the rest of the SPEC held up (see Evidence).

What held up: every RED cell and every control reproduces; F1-F7 and DEC-1 match raw page text; scope diff is exactly the SPEC and report directories; hugo, URL and Mermaid gates green on the unedited tree; traceability 12/12; D1, D3-D11 are genuinely fixed.

## Operational Notes (unverified)

- Measure the revised AC-001 against mutants A, B, C, E and the "sets" control with the same scratch-copy method (`.../scratchpad/mut/mk.py`) before adoption. Status: measured for the current text (E3); assumption for the revised text.
- Confirm that no other line in RS carries two `xhigh` before adopting the whole-file count pin: `grep -c -E 'xhigh.*xhigh' <RS>` -> 0 on this tree. Status: measured.
- Measure OQ-1 live (toggle in a fresh session, `grep -n ultracode` the settings files, start a new session) only if a slider claim is ever wanted. Status: assumption.
- Re-run `moai spec lint` with a ldflags-stamped build from the tree before citing a lint result next to a commit; the tree-built binary used here carries no commit stamp. Status: inferred (verification-claim-integrity section 2.2).

AUDIT-VERDICT: FAIL spec=SPEC-CC-ULTRACODE-TOGGLE-001 receipts=none
