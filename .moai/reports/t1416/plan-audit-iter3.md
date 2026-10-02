auditor-model: claude-sonnet-5-5[1m]

verdict: PASS-WITH-DEBT
audited_sha: 3d44379b86dc8d66ee3d9d38f40e29d5a095d186

# SPEC Review Report: SPEC-CC-ULTRACODE-TOGGLE-001
Iteration: 3/3 (final)
Verdict: PASS-WITH-DEBT (score 0.89; all nine must-pass criteria pass; aggregate above the Tier M threshold 0.80; every prior-iteration defect resolved; two pin-coverage gaps and four minor items are carried as debt, see Defects Found)
Overall Score: 0.89 (iteration 1: 0.75; iteration 2: 0.86; no regression, so no STOP signal)
Plan Artifact Hash (per file, sha256, convenience only; the Go `ComputeHash` was not run): spec.md `8c84764b0a89492e78092cc85f629f81c035fdf934f66eba5c8f948166fb06ac`, plan.md `2f051654b1f3da698f27ca302b2c1fa8db06a6df5dd4f370a5d58eaa2f79655b`, acceptance.md `4804aa57c035fb2af3674fdbb59e1c74610c3ae299d06b28e863ffee92d13f96`, progress.md `15c1579ccc9b286255449236da8ef57c4c670e95dd316c3aa3363e0d0e508f59`
Auditor Version: plan-auditor (Sonnet 5.5 session)

Reasoning context ignored per M1 Context Isolation. Read only the committed artifacts at `3d44379b8` (spec, plan, acceptance, progress) plus the iteration-2 verdict file named by the caller as the defect baseline. Scratch work lives under `<session scratchpad>/audit3/`; the worktree was not written (`git status --short` empty before and after).

## Claim

C1. D2/N1 (the blocking defect of iteration 2) is FIXED: the structural count pins `[R1s]` + `[R1c]` + `[R7]` catch every realistic rewording that keeps a second `xhigh` coupling, and do not false-fire on a correct draft containing "sets" / "runs at".
C2. N2, N3, N4 are FIXED; N5 needed no change.
C3. All 53 RED-now / control entries of the evidence ledger reproduce on this tree (stdout and exit code).
C4. Upstream facts F1-F7 and DEC-1 match raw page text (re-read this run).
C5. All nine must-pass criteria pass (MP-1..MP-9).
C6. The revision introduced no must-pass defect. It left two REQ-005 clauses without a machine pin (toggle / effort-level-unchanged statement in the four docs rows; "same column count"), one locale-dependence in `[W1]`, and four minor consistency items.

## Evidence

All commands run from `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1416`. `grep` is `ugrep 7.8.4`; `/usr/bin/grep` is BSD grep 2.6.0; session `LANG=en_US.UTF-8`.

### E1. Tree attribution and scope

```
git rev-parse HEAD                                    -> 3d44379b86dc8d66ee3d9d38f40e29d5a095d186
git status --short                                    -> (empty)
git diff --stat c50da9c2f 3d44379b8                   -> 6 files changed, 1031 insertions(+):
    .moai/reports/t1416/plan-audit-iter2.md, .moai/reports/t1416/plan-audit.md,
    .moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/{acceptance,plan,progress,spec}.md          (exit 0)
git diff --stat c50da9c2f HEAD -- .claude internal docs-site  -> (empty, exit 0)
```
Nothing outside `.moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/` and `.moai/reports/t1416/` changed since the base. The scope files at HEAD are byte-identical to the ledger tree (`ff7b64f6e`) and to the original pin `c50da9c2f`.

### E2. RED-now ledger replay

`replay_ledger.py` parses every `[ID] command / stdout / exit` entry of acceptance.md § Evidence ledger, runs the command via `subprocess` (shlex argv, no shell, cwd = worktree) and compares the stdout line set (order-insensitive: ugrep prints multi-file `-c` in nondeterministic order, as the SPEC says) and the exit code:

```
entries ok: 53 mismatch: 0
```
So every cell reproduces, including the decisive RED cells: `[R1b]` 1,1 exit 0; `[R2]`..`[R8]`, `[R10]`, `[R11]` 0,0 exit 1; `[R12]` 1,1 exit 0; `[W1]` 0,0,0,0 exit 1; `[W6]` 1,1,1,1; `[M1]` 1,1,1,1; `[U1]` 1,1,1,1; `[U1c]` ko 5 / en,ja,zh 1; `[U2]` ko 1; `[U3]` 0,0,0,0; `[C1*]` 1 each; `[C3-*]` 0 each exit 1; controls `[R1s]` 0,0; `[R1c]` 1,1; `[W5]` 0,0,0,0; `[X-cmp]` empty exit 0. Re-checked by hand (ugrep, direct Bash): `[R1s]` RS:0 RM:0; `[R1c]` 1,1; `[W1]` 0,0,0,0; `[W5]` 0,0,0,0; `[U1c]` ko:5 en:1 zh:1 ja:1. The ko `[U1c]` expectation of 4 after the edit is correct: `grep -n xhigh` on the ko page prints L67 (the effects bullet), L144, L151, L152, L153.

MP-8 per-AC: AC-001 (`[R1b]` 1,1 -> 0,0 plus positives), AC-002, AC-004 (`[W1]`..), AC-005 (`[M1]`), AC-006 (`[U1]`..), AC-007 (`[C1*]`, `[C3-*]`) are each classified release-blocking, each carries command + verbatim stdout + exit code + tree SHA (document-level pin `ff7b64f6e`, acceptance.md L5, L115), each command is a single read-only `grep -c` / `cmp` invocation with `|`, `;`, `(` only inside single quotes, and each has at least one RED cell that reproduces (E2). The runner is `grep`, whose swept set is real files (counts are the files' own matches, never `no tests to run`).

### E3. D2/N1 re-verification: rule-source mutants (the decisive probe)

`mut_rule.py` takes the real L111 bullet, substitutes a correct draft (all positive phrases present, one `xhigh` in the launch-flag clause), adds exactly one extra sentence per mutant, writes a scratch copy, and runs the AC-001/AC-002 pin set (`[R1b]` 0, `[R12]` 0, `[R1s]` 0, `[R1c]` 1, `[R2]`..`[R8]`,`[R10]`,`[R11]`,`[R9]` at least 1, `[R13]` 0) with the real `grep`:

```
D_correct                          PASS  tripped=[]
FP_sets  (correct, contains "sets") PASS  tripped=[]
A_pairs_with                       FAIL  tripped=['R1s']
B_lifts                            FAIL  tripped=['R1s']
C_defaults                         FAIL  tripped=['R1s']
E_best_used                        FAIL  tripped=['R1s']
F_forces                           FAIL  tripped=['R1s']
J_own_line (coupling on its own line) FAIL  tripped=['R1c']
N7_extra_xhigh_no_backtick         FAIL  tripped=['R1s']
U_upstream_phrase ("sets the level to `xhigh`") FAIL tripped=['R7']
OLD (today's line 111)             FAIL  tripped=['R1b','R12','R2','R3','R4','R5','R6','R7','R8','R10','R11']
--- coupling worded WITHOUT the token `xhigh` / case variant ---
N1_max_effort  ("raises the session to the maximum reasoning effort") PASS  (passes wrongly)
N2_highest     ("always uses the highest effort level")              PASS  (passes wrongly)
N3_implies     ("implies a very high reasoning level")               PASS  (passes wrongly)
N4_pair_with_max ("pairs with `/effort max`")                        PASS  (passes wrongly)
N5_x-high      ("pairs with x-high reasoning")                       PASS  (passes wrongly)
N6_XHIGH_upper ("pairs with XHIGH reasoning")                        PASS  (passes wrongly)
```
Reading: every `xhigh`-token coupling (A, B, C, E, F, J, N7) is now caught structurally, the correct draft D and the "sets" false-positive probe pass, the old line fails on 11 pins. The upstream-faithful phrase "sets the level to `xhigh`" fails `[R7]` because REQ-002 mandates the words "starts the session at" — a stated constraint, not a defect. The class that passes wrongly is a coupling stated without the token `xhigh` (N1-N5) or in another case (N6; `[R1s]`/`[R1c]` use case-sensitive `-F`/`-E` without `-i`).

### E4. Docs-row mutants (en row inside the real page; `mut_docs2.py`) and locale validity of `[W1]` (`mut_docs.py`)

Pins judged: `[W1]` >=1, `[W2]`,`[W3]`,`[W4]`,`[W8]`,`[W10]`,`[W7]` >=1, `[W5]`,`[W6]`,`[W11]` = 0.

```
I_correct                          PASS
G_semicolon                        FAIL tripped=[W1]
G2_comma                           FAIL tripped=[W1]
H_next_sentence                    FAIL tripped=[W1]
X_two_xhigh (old opening kept)     FAIL tripped=[W5]
Y_xhigh_before_literal             FAIL tripped=[W1, W5]
Z_xhigh_outside_flag_clause_only   FAIL tripped=[W1]
R_effort_high_kept                 FAIL tripped=[W6]
S_slider_mentioned                 FAIL tripped=[W11]
L_rowlabel_changed (`[on|off]` in the label cell) FAIL tripped=[W7]
OLD_real                           FAIL tripped=[W1,W2,W3,W4,W8,W6]
G3_commafree ("also starts the session and ultracode always runs at `xhigh`") PASS  <- passes wrongly (declared residual, acceptance.md L38, L99)
T_no_toggle_no_unchanged (row never says toggle / effort level unchanged; all other literals present) PASS  <- passes wrongly (NEW, D1 below)
P_pipe_in_cell (`/effort ultracode [on|off]` inside the description cell) PASS  <- passes wrongly (NEW, D2 below); table cells (naive split) 2 -> 3
```
Candidate pins measured: `^\| `/effort ultracode` \|[^|]*\|$` prints 1 on I_correct, T and OLD_real, 0 on P and L (it closes the column-count gap); a row-scoped toggle-word pin (`^\| `/effort ultracode` \|.*toggle`) closes T.

`[W1]` validity per locale (correct drafts, the flag clause worded without internal punctuation; ugrep and BSD grep, three environments):

```
              default(LANG=en_US.UTF-8)   LC_ALL=en_US.UTF-8   LC_ALL=C
ko   W1/W5          1 / 0                       1 / 0              0 / 0   <- false negative on correct text
en   W1/W5          1 / 0                       1 / 0              1 / 0
ja   W1/W5          1 / 0                       1 / 0              0 / 0   <- false negative
zh   W1/W5          1 / 0                       1 / 0              0 / 0   <- false negative
ja with "は、" (natural topic comma) between literal and xhigh: W1 = 0 in every environment
zh with "，" between literal and xhigh: W1 = 0 in every environment
```
Both greps agree. The bracket `[^|.。;；,，、]` is evaluated bytewise under `LC_ALL=C`, and the excluded multibyte punctuation shares lead / continuation bytes with ordinary Hangul and kana, so the `*` run stops early.

### E5. Ultracode-workflows page mutants (`mut_uw.py`, en page with the effects bullet replaced and the off route added)

```
correct_bullet_removed   U1=0 U1c=0 U3=1 PASS
relabelled_bullet        U1=1 U1c=1       FAIL
star_bullet ("* ...")    U1=0 U1c=1       FAIL (caught by U1c)
numbered ("1. ...")      U1=0 U1c=1       FAIL
indented_sub             U1=0 U1c=1       FAIL
prose_token              U1=0 U1c=1       FAIL
bold_label               U1=1 U1c=1       FAIL
prose_no_token ("raised to the highest level") U1=0 U1c=0 PASS  <- token-less residual, same class as E3 N1-N5
```

### E6. Upstream facts (raw fetch this run: `curl` for the three docs pages, `python3 urllib` for the changelog because a `curl` of the raw changelog URL was refused by the worktree guard; tags stripped; 2026-10-02)

- F1: `changelog.md:406` (inside `## 2.1.284`, lines 342-445): "Changed Ultracode into its own toggle in `/effort` (Tab, or `/effort ultracode [on|off]`): it no longer forces xhigh effort and stays on at any effort level". Newer entries 2.1.285-2.1.287 do not change ultracode semantics (`grep -n -i ultracode` lists only 347, 406, 416 (VSCode switch), 608 and older lines).
- F2: model-config: "Ultracode is a Claude Code setting rather than a model effort level: with it on, Claude orchestrates dynamic workflows for substantive tasks, at whichever effort level the session runs at." "Turning ultracode on or off with /effort or the ultracode setting leaves the effort level unchanged." Match.
- F3: "The --effort ultracode flag and the Agent SDK effortLevel: "ultracode" value turn it on and also set the level to xhigh. Picking a level in the /effort slider or the /model picker leaves ultracode as it was." Also "launch with claude --effort ultracode, which starts the session at xhigh effort with ultracode on" — the same words as REQ-002's mandated phrase. Match.
- F4: "The /effort ultracode off form, the slider toggle, and keeping ultracode on at effort levels other than xhigh require Claude Code v2.1.284 or later. Before v2.1.284, turning on ultracode set the session to xhigh effort, picking another level turned it off ..." Match.
- F5 (workflows page): "/effort ultracode lasts for the current session; to have every session start with it, set the ultracode setting. Turn it off with /effort ultracode off ..." Match.
- F6: "The persisted effortLevel setting and the CLAUDE_CODE_EFFORT_LEVEL environment variable don't accept ultracode. If CLAUDE_CODE_EFFORT_LEVEL or an effort cap sets the session's level, ultracode stays on at that level." Match.
- F7 (settings-reference `ultracode`): "The key doesn't change the session's effort level: ultracode runs at whichever level the session uses. Claude Code reads this key but never writes it: /effort ultracode turns ultracode on for the current session only." "... --effort ultracode flag also turns it on for one session, at xhigh effort, and requires Claude Code v2.1.203 or later". "This and the /effort ultracode off form require Claude Code v2.1.284 or later. Before v2.1.284, ultracode: true ran the session at xhigh effort, and an effort cap below xhigh kept ultracode off." Match; DEC-1 is correctly worded.
- Fact the SPEC does not use, found while reading: ultracode is unavailable when workflows are off or the model does not support `xhigh` (model-config), and the settings entry says sessions start with it on "when dynamic workflows are enabled for you and your model supports xhigh". Not asserted by the SPEC; relevant only to question (ii) below.

### E7. Structural verbs

- Traceability (the verbatim `awk` verb, run via a script file; it was not refused this time): `COLLECTED: 12 REQ definitions (acceptance input: read)`, no `UNCOVERED:`, no `ORPHAN:` lines. REQ to AC headings read by hand: REQ-001 AC-001; REQ-002/003 AC-002; REQ-004 AC-001/003/011; REQ-005 AC-004; REQ-006 AC-005; REQ-007 AC-006; REQ-008 AC-007; REQ-009 AC-004/008; REQ-010 AC-008/009; REQ-011 AC-004/010; REQ-012 AC-010. Collected N = 12 > 0. Note mapping by REQ id is coarse: the pin-level coverage of REQ-005 is weaker than the id mapping suggests (D1, D2).
- CN-4 ordering verb (verbatim `awk`, script file): `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 0 exit bindings, 6 ordering candidates`, no `CONFLICT:`. Candidates (acceptance.md L3, L33, L38, L56, L60, L64) are "after/before/first" words in pin descriptions (for example "must stay exit 0 after the edit"), none binds a milestone on a forbidden side. Read by hand: M1 (rule source) -> M2 (ko, en, ja, zh) -> M3 (mirror) -> M4 (verification) agrees with AC-003 (assertion after both edits) and AC-011; M3's "same step" wording and M1's RS-only edit leave a transient `cmp` difference between M1 and M3 that no AC asserts on.
- `moai spec lint SPEC-CC-ULTRACODE-TOGGLE-001`:
  - installed `moai` = `moai_cp/20260925_122548-1711-gd194083fb`, built 2026-09-30; `git merge-base --is-ancestor d194083fb HEAD` -> exit 0, a strict ancestor of HEAD (3d44379b8), so it lags -> `No findings`, exit 0.
  - binary built from this tree (`go build -o <scratch>/moai-tree ./cmd/moai`, exit 0; `moai-tree version` prints `v3.1.3 none built unknown` because no ldflags were passed, so it carries no commit stamp) -> `No findings`, exit 0.
  - positive control: a scratch SPEC copy with an extra uncovered `REQ-013` linted with the tree-built binary -> `0 error(s), 0 warning(s)` (only two INFO git-unreachable findings). The linter is silent on traceability for this SPEC form, so its clean result is NOT cited as corroboration of AC-4 / AC-5; the traceability reading in this section is the auditor's own.
- D7: spec.md names only its own id; plan.md also names `SPEC-CC-GD124-001`, `status: completed` (not retired / superseded / archived). D8: `grep -c syscall spec.md` -> 0. MP-7: `grep -rn '\[NEEDS CLARIFICATION' plan.md` -> no output; research.md absent (Tier M).
- Plan assertions spot-checked: `grep -c -E 'rules|dynamic-workflows' internal/template/scripts/gen-catalog-hashes.go` -> 0; `internal/template/embed.go:28: //go:embed all:templates`; `docs-site/.locale-parity-baseline` lists `advanced/ultracode-workflows.md` (L18) and `claude-code/foundations/commands.md` (L35). Match.
- AC-009 gate reproduced on a scratch copy of `docs-site` (a copy has no `.git`, so `--enableGitInfo=false` was required; the first attempt without it exited 1 on "failed to load Git data"): `hugo --minify --gc --enableGitInfo=false` -> exit 0, `grep -c -E 'WARN|ERROR'` -> 0. URL-blacklist grep over `docs-site/content` -> no output; Mermaid `flowchart/graph LR|RL` grep -> no output.

## Baseline-attribution

- Tree measured: worktree `.moai/worktrees/t1416`, HEAD `3d44379b86dc8d66ee3d9d38f40e29d5a095d186`, clean, in this run. The ledger's tree pin is `ff7b64f6e` (acceptance.md L5, L115); scope files at HEAD are identical to it and to `c50da9c2f` (E1), so the pin names a tree whose scope content equals the audited one.
- Judging builds for the lint measurement: installed `moai` build `d194083fb` (strict ancestor of HEAD, lagging) and a tree-built binary at HEAD (`go build` without ldflags, no commit stamp). Per `verification-claim-integrity.md` section 2.2 neither result is cited as corroboration (positive control silent).
- Upstream pages were fetched raw in this run; no figure is carried over from iteration 1 or 2.
- Mutant results are measurements of the real `grep` on scratch copies derived from this tree; the drafts are the auditor's, not the SPEC's.

## Gaps

1. No cross-backend audit (`audit_multi`, `codex_audit`, `glm_audit`, `claude_audit`) was invoked; the caller did not request one. No audit receipt exists.
2. Refused commands: (a) a compound `git diff ...; echo rc=$?` and (b) a `curl` of the raw changelog URL and (c) a `hugo` call with a computed shell variable were refused by the worktree guard; each was re-issued as a plain command, a script file with literal paths, or `python3 urllib`. Nothing was substituted by reading. The traceability and CN-4 `awk` verbs ran verbatim from script files this time.
3. AC-009 was run on a scratch copy with `--enableGitInfo=false`, not in-tree: an in-tree `hugo` would write `.hugo_build.lock` into the worktree.
4. AC-011 (`make build`, `go test ./internal/template/...`) was not run: the mirror is unedited and the AC describes the post-work state.
5. The ja / zh / ko drafts used for `[W1]` validity are the auditor's; native idiom of the real run-phase text is unobservable at plan time. GNU grep was not tested (ugrep and BSD grep only).
6. OQ-1 (slider-toggle persistence) was not observed live; the SPEC avoids the claim.
7. The ledger replay compares stdout line sets and exit codes, not byte order (ugrep order is nondeterministic).
8. The Go `ComputeHash` was not run; the sha256 values above are per-file conveniences.

## Residual-risk

- Pins are string-presence and count checks on prose. They cannot prove that a locale's sentence reads correctly, that a coupling worded without the token `xhigh` is absent (E3 N1-N5, E5 `prose_no_token`), or that a flag-clause relocation inside one comma-free clause is absent (G3).
- `[R1s]`/`[R1c]` are case-sensitive; `XHIGH` / `Xhigh` would pass.
- The settings-key sentence ages if upstream changes F7; the SPEC pins `v2.1.284` as the introduction version only.
- OQ-1 stays open; REQ-003 is now limited to the command form, so nothing asserted exceeds the evidence.
- Any change to acceptance.md or spec.md changes the plan-artifact hash and withdraws skip-eligibility for the run-gate (spec-workflow § Plan Audit Gate skip policy); a micro-edit to close D1/D2 therefore costs a fresh Phase 1 audit at run entry.

## Must-Pass Results

- [PASS] MP-1 REQ numbering: `- **REQ-001**` .. `- **REQ-012**` at spec.md L54-L65 (grep output above), sequential, no gap, no duplicate, uniform zero padding. 12 <= 16 (Tier M requirement ceiling); 11 ACs <= 16.
- [PASS] MP-2 GEARS (layer judged: requirement layer, `REQ-XXX` in spec.md; ACs are verification-layer and graded under testability): REQ-001..008 ubiquitous `shall` statements (spec.md L54-L61), REQ-009 "**When** the docs-site edits are made, ko shall be authored first ..." (L62, event-driven), REQ-010..012 legacy `shall not` negatives (L63-L65; allowed at 1.0, never canonical). No informal "should / must try".
- [PASS] MP-3 frontmatter (spec.md L1-L15): id, title (quoted), version `"0.1.2"` (quoted semver), status `draft`, created / updated `2026-10-02`, author, priority `P2`, phase `"v3.2.0 target"` (release label), module, lifecycle `spec-anchored`, tags (comma string), optional `tier: M`. No snake_case alias. plan.md, acceptance.md, progress.md carry no `status:` frontmatter. `moai spec lint` clean on two builds (not cited as corroboration, E7).
- [N/A] MP-4: prose wording correction of a Claude Code feature description; no language-specific tool is named or required (the four docs locales are the subject, not tooling).
- [PASS] MP-5 D7: only self-reference in spec.md; plan.md's `SPEC-CC-GD124-001` has `status: completed`, not retired / superseded / archived. No BLOCKING.
- [PASS] MP-6 D8: `syscall` appears 0 times in spec.md; no BLOCKING.
- [PASS] MP-7: no `[NEEDS CLARIFICATION` marker in plan.md (grep, no output); research.md absent (Tier M), that half N/A.
- [PASS] MP-8: every release-blocking AC (AC-001, AC-002, AC-004, AC-005, AC-006, AC-007) has command, verbatim stdout, exit code and a tree SHA in the ledger; all 53 cells re-executed and reproduce (E2); every RED is a real signal (positives 0 with exit 1, negatives non-zero with exit 0, grep over real files). Ledger commands conform to the single-invocation form.
- [PASS] MP-9: CN-4 verb: `COLLECTED: 4 milestones ... 0 exit bindings, 6 ordering candidates`, no `CONFLICT:`; candidates are pin descriptions; M1-M4 order consistent with the DoD (acceptance.md L113) and AC-003 / AC-011 (E7).

## Disposition of iteration-2 items

| ID | Disposition | Evidence |
|---|---|---|
| D2 / N1 (rule-source coupling regex was a closed verb list) | FIXED | `[R1]` retired; `[R1s]` (`xhigh.*xhigh` = 0,0) + `[R1c]` (lines containing `xhigh` = 1,1) + `[R7]` (>= 1) now force exactly one `xhigh`, inside the launch-flag clause, in any wording that uses the token (acceptance.md L24, L28). Re-measured: mutants A (pairs with), B (lifts), C (defaults to), E (best used at), F (forces), J (own line), N7 all FAIL (R1s / R1c); correct draft D and the "sets" false-positive line PASS (E3). The "wording-independent" claim is gone and the limit is stated at L24 ("does not catch a coupling claim worded without the token `xhigh`"). Residual: token-less and case-variant couplings (see answer (i)). |
| N2 (REQ-003 wider than the evidence) | FIXED | spec.md L56: "a toggle made with the `/effort ultracode` command lasts for the `current session` ... (the slider toggle is not covered by this sentence; its persistence is OQ-1)". Matches F5 / F7. |
| N3a (W1 not bound to the flag clause) | FIXED | `[W1]` is now `--effort ultracode[^|.。;；,，、]*xhigh` (acceptance.md L37). Mutants G (semicolon), G2 (comma), H (next sentence), Y, Z all FAIL; I (correct) passes; G3 comma-free relocation passes and is declared (L38, L99). New caveat on locale dependence: D3. |
| N3b (U1 keyed to the label) | FIXED | `[U1]` is `^- .*xhigh` and `[U1c]` counts `xhigh` lines per page (L47). Relabel, star, numbered, indented, prose-with-token, bold-label mutants all FAIL; the correct page passes (E5). Token-less prose passes (same residual class). |
| N4 (vacuity sentence) | FIXED | acceptance.md L108: "every positive pin above is 0 today (RED) per the ledger; zero pins that are 0 today (`[R1s]`, `[W5]`, `[R13]`, `[W11]`) are regression-guards". Checked against the ledger: every positive pin (R2-R8, R10, R11, W1-W4, W8, U3, C3-*) is 0 today; the four named zero pins are 0 today. |
| N5 (ledger size) | NOT FIXED, no change required | Ledger is now L115-L419 (305 lines; was 293). It is the recommended carrier and is addressed by pin id; no stated criterion is violated. Optional compaction only. |

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.90 | between 0.75 and 1.0 | REQs carry verbatim backticked phrases and exact lines (spec.md L54-L65, L74-L80); REQ-003 now names the command form (L56). REQ-005 is one very long compound sentence (L58) and leaves the order "literal before `xhigh`" implicit although `[W5]` enforces it (D5). |
| Completeness | 0.95 | between 0.75 and 1.0 | HISTORY L19-L23, Context L25, Requirements L52, Non-Goals L67, Change Map L72, three `### Out of Scope — <topic>` H3 sections with specific bullets L84-L100, plan, acceptance, progress, 12 frontmatter fields. Two stale-pin items (D4). |
| Testability | 0.80 | between 0.75 and 1.0 | Every pin is a binary command with printed counts; mutant probe: A-F, J, N7, G, G2, H, X, Y, Z, U-mutants all caught, correct drafts pass. Writable mutants that satisfy AC-004 while violating REQ-005: T (no toggle / effort-unchanged statement) and P (`[on|off]` splits the cell), plus `[W1]` locale dependence (D1-D3). |
| Traceability | 0.90 | between 0.75 and 1.0 | 12 collected, 0 uncovered, 0 orphan by the verbatim verb (E7); every REQ names an AC; two REQ-005 clauses and one REQ-010 clause (emoji) have no pin of their own (D1, D2, D6). |

Aggregate (0.90 + 0.95 + 0.80 + 0.90) / 4 = 0.8875, reported 0.89, against the Tier M threshold 0.80: above.

## Defects Found (structured defect-list)

D1. REQ-005 "toggle" and "effort level unchanged" statements have no pin in the four docs rows — acceptance.md:L35-L39 (spec.md:L58) — REQ-005 requires each locale's row to "state the facts of REQ-001 to REQ-003 — toggle, effort level unchanged, ...". The AC-004 pins cover `/effort ultracode off`, the settings key, `v2.1.284`, the flag exception and the scope literal, but nothing pins the toggle statement or the effort-level-unchanged statement on the row (they are pinned only for the English rule source, `[R2]`/`[R3]`, and the commands pages carry a toggle-word pin `[C3-*]`). Mutant T (a row that drops the false claim, adds off / key / version / flag exception, and never says toggle or that the level is unchanged) passes every AC-004 pin (E4). The SPEC's central corrective fact is therefore unpinned on its card-named surface. — Severity: major — Class: blocking — Required fix: add a row-scoped toggle-word pin per locale, for example `[W12]` `grep -c -E -- '^\| `/effort ultracode` \|.*(toggle|토글|トグル|开关)'` per file (one pattern per locale, `0,0,0,0` today, at least `1` each after), and, if wanted, a locale phrase pin for "effort level unchanged". Otherwise record in plan.md §E M4 that REQ-005's toggle / unchanged clauses are judged by reading the diff.

D2. REQ-005 "single table row with the same column count" is not pinned — acceptance.md:L39 (`[W7]` "row shape"), spec.md:L58 — `[W7]` (`^\| `/effort ultracode` \|`) proves only that the row still starts with its label cell; it cannot see a pipe inside the description cell. The natural upstream notation is `/effort ultracode [on|off]` (changelog L406, model-config); in GFM a pipe inside a code span still splits the cell, so a row containing it renders with a third cell and the text after the pipe is dropped. Mutant P passes every AC-004 pin while the row has 3 cells instead of 2 (E4), and `hugo` does not warn on it (AC-009 is silent). plan.md §B (known traps) does not warn about it. — Severity: minor — Class: blocking (a stated REQ clause with no pin) — Required fix: strengthen `[W7]` to `^\| `/effort ultracode` \|[^|]*\|$` (measured: 1 on a correct row, 0 on mutants P and L, 1 on today's row) and add the pipe hazard to plan.md §B.

D3. `[W1]` is locale-dependent and the AC does not say so — acceptance.md:L37 — the bracket `[^|.。;；,，、]` is evaluated bytewise under `LC_ALL=C`; on correct ko / ja / zh drafts `[W1]` printed 0 under `LC_ALL=C` and 1 under UTF-8 for both ugrep and BSD grep (E4). A run-phase agent or CI shell with `LC_ALL=C` would see a wrong-reason red on correct work. The pin is valid for all four locales only under a UTF-8 locale (this session's `LANG=en_US.UTF-8` is one). REQ-005's no-`、`/`，` rule also forces ja and zh clause punctuation (a natural "は、" breaks it), which is workable but is a constraint on native prose that plan.md §D does not mention beyond "worded without internal commas". — Severity: minor — Class: optional — Required fix: state in AC-004 that the ledger and the green path are produced under a UTF-8 locale (`LC_ALL=en_US.UTF-8` or `LANG=en_US.UTF-8`), and note that the flag clause must avoid the topic comma in ja / zh.

D4. Stale evidence pin and HISTORY order in spec.md — spec.md:L50 and L21-L23 — L50 still says "Evidence baseline: measured on this worktree at HEAD `0e7b6af5b` ... Cells are in `acceptance.md` § Evidence ledger", while acceptance.md L5 and L115 (and progress.md) say the ledger was regenerated at `ff7b64f6e`. The scope content is identical (E1), so no measurement is wrong, but the spec names a tree the cited cells were not measured on. HISTORY lists v0.1.2 (L22) before v0.1.1 (L23), against the oldest-first order of L21. — Severity: minor — Class: blocking by the internal-consistency definition, trivial in effect — Required fix: change L50 to `ff7b64f6e` (or "the ledger tree named in acceptance.md L5") and swap L22 / L23.

D5. REQ-005 does not state the order "literal before `xhigh`" that `[W5]` enforces — spec.md:L58 vs acceptance.md:L38 — `[W5]` fails any row where `xhigh` precedes the `--effort ultracode` literal (mutant Y fails W1 and W5), although REQ-005 only requires one clause with no break between "the literal and `xhigh`". A REQ-compliant sentence with `xhigh` first is rejected by the AC. Natural word order in all four locales puts the literal first, so the practical risk is small. — Severity: minor — Class: optional — Required fix: add "the literal precedes `xhigh`" to REQ-005.

D6. REQ-010 "no emoji in body text" has no command — acceptance.md:L58-L60 (AC-009 lists hugo, URL blacklist, Mermaid) — plan.md §M4.3 delegates to `hns-oss-docs-verify` §1-§4 but AC-009 does not carry the emoji scan, so the clause rests on diff review. — Severity: minor — Class: optional — Required fix: add one emoji-scan command to AC-009 or state that the clause is judged by reading the diff.

D7. (informational, carried from the SPEC's own declaration) coupling stated without the token `xhigh`, a case variant (`XHIGH`), and a comma-free in-clause relocation (G3) pass the pin set — acceptance.md:L24, L38. — Severity: minor — Class: optional — Required fix: none required (acknowledged as a diff-review residual; a case-insensitive `-i` on `[R1s]`/`[R1c]` and a zero pin on `/effort (low|medium|high|max)` in the rule files would be a cheap partial closure; measured: the rule file has `/effort` only on L111 today).

D8. Minor format note — acceptance.md:L21-L69 — most ACs are RED-now/green-path pin lists, not Given-When-Then scenarios (only AC-010/AC-011 read G-W-T). They are binary-testable, which is the property that matters, so no fix is required. — Severity: minor — Class: optional.

## Regression Check (Iteration 2+)

Defects from iteration 2: D2/N1 FIXED, N2 FIXED, N3 FIXED, N4 FIXED, N5 no change required (table above). Defects of iteration 1 (D1, D3-D11) were closed in iteration 2 and their cells still reproduce (E2). No score regression (0.75 -> 0.86 -> 0.89), so no STOP signal. The CN-4 ordering verb was re-run in full (E7): no conflict.

## Answers to the author's review questions

(i) Token-counting structural pin plus the stated limitation: acceptable as a residual. The count pins are language-neutral, wording-independent for every coupling that uses the token, and survive the false-positive probe; the measured pass-through class (N1-N6) is coupling worded without `xhigh`. Such a sentence would sit on the same single bullet as the mandated "leaves the effort level unchanged" and contradict it in plain view of the diff, and AC-010 already routes line-level review to the diff. Two cheap partial closures exist and are optional (D7): `-i` on `[R1s]`/`[R1c]`, and a zero pin on `/effort (low|medium|high|max)`; neither closes N1-N3.

(ii) REQ-002's whole-file limit of exactly one `xhigh`: sound for this SPEC. The rule file has exactly one `xhigh` today (`grep -c` 1,1) and the SPEC replaces it with exactly one, so the constraint forbids only additions. It is a point-in-time pin, not a standing gate, so it cannot "force a SPEC change later" by itself. The only latent tension found: upstream now documents an availability prerequisite that mentions `xhigh` (ultracode is unavailable when the model does not support `xhigh`), so a later edit that wants to state it would have to relax the pin first. The SPEC does not claim that fact, so no false statement results; optional note only.

(iii) `[W1]` for all four locales: valid when the grep runs under a UTF-8 locale and the flag clause carries no internal punctuation (E4: 1,1,1,1 for ko/en/ja/zh under the default environment, ugrep and BSD grep). Invalid under `LC_ALL=C` for ko / ja / zh (false negative on correct text) — D3. Because the flag clause does not yet exist, W1 is correctly a constraint on sentence construction (REQ-005 says so), but the ja / zh topic comma makes the constraint non-trivial for native prose and REQ-005 should also state the literal-before-`xhigh` order (D5).

## Recommendation

PASS-WITH-DEBT. All nine must-pass criteria pass, the aggregate 0.89 exceeds the Tier M threshold 0.80, every prior-iteration defect is resolved, and the iteration cap (3) is reached. The SPEC is executable by a run-phase agent without further decisions: plan.md names the exact phrases, lines, order and mirror step. Carried debt, in priority order:

1. D1 and D2 are two real REQ-005 clauses that no command pins (mutants T and P pass AC-004). Either accept them as debt with the run-phase agent and the sync audit reading those two clauses in the diff (record that in progress.md §E.2), or have manager-spec add the two pins above. The second route edits acceptance.md, changes the plan-artifact hash and therefore withdraws skip-eligibility for the run-gate (Phase 1 re-audits at run entry); the first route costs nothing now.
2. D3 (UTF-8 locale note) and D4 (stale pin, HISTORY order) are one-line fixes if the spec is touched at all.
3. D5-D8 are optional.

## Operational Notes (unverified)

- Measure the final docs rows against the pins under both `LANG=en_US.UTF-8` and `LC_ALL=C` before reading a red `[W1]` as a defect in the text. Status: measured for the auditor's drafts (E4); assumption for the real rows.
- Before the run-phase agent writes the en row, avoid `/effort ultracode [on|off]` inside the table cell (a pipe splits a GFM cell even inside a code span); measure the cell count after the edit, for example `grep -c -E -- '^\| `/effort ultracode` \|[^|]*\|$' <page>` should print 1 per locale. Status: measured on scratch rows (E4); inferred for GFM behaviour (the rendered page was not inspected).
- If the orchestrator chooses the manager-spec micro-edit route (D1, D2), expect a fresh Phase 1 audit at run entry because the plan-artifact hash changes. Status: inferred (spec-workflow § Phase 1 Plan Audit Gate, skip contract condition 3); whether `PASS-WITH-DEBT` counts as `PASS` for skip-eligibility condition 1 was not checked in the runtime code.
- Measure OQ-1 live (toggle in a fresh session, `grep -n ultracode` the settings files, start a new session) only if a slider claim is ever wanted. Status: assumption.
- Re-run `moai spec lint` with an ldflags-stamped build from the tree before citing a lint result next to a commit. Status: inferred (verification-claim-integrity section 2.2).

AUDIT-VERDICT: PASS-WITH-DEBT spec=SPEC-CC-ULTRACODE-TOGGLE-001 receipts=none
