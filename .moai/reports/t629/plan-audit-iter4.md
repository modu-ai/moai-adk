# SPEC Review Report: SPEC-REVIEW-SECRET-SCAN-REFS-001
Iteration: 4/2 (amendment audit) — **ceiling override**: the Tier M ceiling is 2
(`.moai/config/sections/harness.yaml` `plan_audit_tier_ceilings` M: 2). Iteration 3 already exceeded it
by one with operator approval and returned PASS 0.89 (`.moai/reports/t629/plan-audit-iter3.md`). The
operator approved this further audit, scoped to one amendment (answered in the lead session, relayed by
the lead, 2026-09-11). The operator also directed that if this audit FAILs, no fixes are made and the
question returns to the operator.
Verdict: PASS
Overall Score: 0.87 harmonic (arithmetic mean 0.875; Tier M PASS threshold 0.80)
STOP signal (LEAN rule): raised, advisory — see § Score movement.

card: t629 · auditor: plan-auditor (independent, read-only) · date: 2026-09-11
Reasoning context ignored per M1 Context Isolation — the lead's brief was used only as the scope
statement and a list of places to look; every reading relied on below was re-measured in this run
unless it is labelled as carried.

**Scope (binding, from the brief).** Only the amendment delta `git diff 6e56840d5 9dab82a4a`
(`spec.md`, `plan.md`, `acceptance.md`) and regressions it creates in the SPEC. Unchanged text is not
re-audited; serious observations about it are listed under "Out of scope, not scored". The stale pin
in `progress.md` §E.2 (still the command-substitution form) is deliberately unchanged by this commit
and is not graded as an amendment defect.

## Baseline-attribution

Tree measured: worktree `.claude/worktrees/t629`, branch `WT-secret-scan-refs`.

```
$ git rev-parse --show-toplevel   → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t629
$ git rev-parse --short HEAD      → 9dab82a4a   (read at start and again before writing this report)
$ git status --porcelain          → (empty, both times)
$ git diff --stat 6e56840d5 9dab82a4a → acceptance.md 62, plan.md 25, spec.md 50; 113 insertions(+), 24 deletions(-)
$ git show --stat 9dab82a4a       → the same three SPEC files only
```

Delta read in full (255 diff lines). Artifacts read: `spec.md` lines 1-64, 181-236, 270-425; `plan.md`
lines 29-208; `acceptance.md` lines 1-60, 190-386, 456-536; `progress.md` lines 138-300.

Judging build for `moai spec lint`: the installed `moai`, which reported `v3.2.0-rc.6` when lint ran and
`v3.2.0-rc.7` (build `moai_cp/20260910_130400-275-ged71054d3-dirty`) when re-read later in this run — the
binary was replaced by another actor mid-session. The rc.6 build's commit was not captured. Gap: lint
attribution names a version, not a commit.

Git: `git version 2.50.1 (Apple Git-155)`.

## Re-measurements this run

| Reading | Command | Observed |
|---|---|---|
| lint | `moai spec lint SPEC-REVIEW-SECRET-SCAN-REFS-001` | `No findings — all SPEC documents are valid`, exit 0 |
| REQ / AC numbering | `grep -oE '^- \*\*REQ-[0-9]{3}' spec.md`; `grep -oE '^\| AC-[0-9]{3}' acceptance.md` | REQ-001 … REQ-013 in order; AC-001 … AC-016 in order |
| decision line | `grep -c '^\*\*Decision:\*\* Option 2' spec.md` | `1` |
| credential regex | `/usr/bin/grep -cE -- '<REGEX>'` over `spec.md`, `plan.md`, `acceptance.md`, `progress.md` | `0` each, exit 1 |
| review.md copies | `git diff --stat feeecc980 9dab82a4a -- <LOC> <TPL>` to a file; `wc -c`; `diff -q LOC TPL` | diff exit 0, 0 bytes; `diff -q` exit 0 — HISTORY's "before any edit to either copy" holds |
| cross-SPEC / syscall / clarification | `grep -ohE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'`; `grep -c syscall`; `grep -rn 'NEEDS CLARIFICATION' plan.md` | only the SPEC's own ID; `0` each; exit 1, no `research.md` |
| remaining `--not` / substitution wording | `grep -nE -e '--not' -e '\$\(cat' -e stdin -e objectname -e 'command substitution'` over the three files | classified in § Check 1 |
| AC-006 baseline on `TPL` at `9dab82a4a` | the §D.6 set-up and five checks | section `20` lines; cmds `2`; `--all` inverse `1` exit 0; `--stdin` `0` exit 1; `^%(objectname)` `0` exit 1; `every ref` `0` exit 1; HEAD-SHA phrase `1` exit 0 — matches the recorded 0.3.0 baseline |
| AC-006 positive control | the `--stdin` and `^%(objectname)` counts over a two-line scratch file | `1` exit 0; `1` exit 0 |
| AC-006 mutant | the five checks over a scratch section: a scan line feeding `--stdin` from `/dev/null`, the caret format named only as obsolete, "every ref" in a display sentence | `--all` inverse `0`; `--stdin` `1`; `^%(objectname)` `1`; `every ref` `1`; HEAD-SHA phrase `0` — **the mutant passes all five** (§ Check 2) |
| refs on this repository | `git for-each-ref --format='%(refname)'` to a file; `wc -l` | `726` (plan §G cites the lead's `725`; refs move under concurrent lanes — not a defect) |
| caret store, this repository | `git for-each-ref --format='^%(objectname)' refs/heads/WT-secret-scan-refs` to a file | exit 0; `42` bytes; caret lines `1` |
| stdin exclusion, this repository | `git rev-list --count --all --stdin < <caret store>` and `git rev-list --count --all --not HEAD`, back to back | `5335` and `5335`, exit 0 each (an earlier non-adjacent pair read `5335`/`5334` while refs moved) |
| `A` form, this repository | `sed 's/^\^//'` to a plain file (`41` bytes, caret count `0` exit 1); `git rev-list --count --stdin < <plain>`; `git rev-list --count HEAD` | `7082`; `7082` |
| command-line `--not` vs stdin, this repository | `git rev-list --count --not HEAD~1 --stdin < <plain>` | `1` — the stdin tip stayed positive |
| `--stdin` documentation | `man git-rev-list`, `--stdin` entry | "Flags like --not which are read via standard input are only respected for arguments passed in the same way and will not influence any subsequent command line arguments" |

### Scratch fixture (session scratchpad, outside this repository)

`git init -b main`; plain words only; `NEEDLE` is a harmless pickaxe string, not a credential stand-in;
counts only, no object names recorded here. Refs at store time: `main`, `side`, `gone` (one commit
`g1` reachable from no other ref), annotated tag `v1` on `main`. Store:
`git for-each-ref --format='^%(objectname)'` → `4` lines, caret lines `4`. Then `s1` (`NEEDLE`) on `side`
and `c2` (`NEEDLE`) on `main`.

| Step | Command (fixture) | Exit | Reading |
|---|---|---|---|
| scope through caret store | `git log --format=%H --all --stdin < store` | 0 | `2` lines |
| full scope control | `git log --format=%H --all` | 0 | `4` lines |
| pickaxe through store | `git log -p --all -G NEEDLE --stdin < store` | 0 | `2` commits (full control also `2`; only new commits carry the string) |
| A / B / L | `rev-list --count --stdin < plain`; `--count --all`; `--count --not --all --stdin < plain` | 0 each | `2` / `4` / `0`; `B − A + L` = `2` = scope |
| `git log` command-line `--not` vs stdin | `git log --format=%H --not HEAD~1 --stdin < <plain name of g1>` | 0 | `1` line — stdin tip positive under `git log` too |
| AC-009 membership | `grep -cxF -f <plain g1> <stripped store>`; same over the caret store | 0; 1 | `1`; `0` — the membership check needs the stripped store, as §D.9 says |
| delete `gone` (tip still exists) | `git branch -D gone`; then A / B / L and the scope | 0 each | `2` / `3` / `1`; scope `2` = `B − A + L` (3 − 2 + 1) |
| prune | `reflog expire --expire=now --all`; `gc --prune=now --quiet`; `cat-file --batch-check < <plain g1>` | 0 each | `missing` count `1` |
| scan with missing tip | `git log -p --all -G NEEDLE --stdin < store > out 2> err` | **128** | out `0` bytes; err `59` bytes, `1` line, `fatal: bad object` count `1`; caret count in err `0`; `grep -cF -f <plain g1> err out` → `err:1`, `out:0` |
| `A` with missing tip | `rev-list --count --stdin < plain` | **128** | `bad object` `1`; plain name `1` — cell ③'s "A cannot be computed → re-taken, gap" path is what happens |

No `-G` history scan was run on this repository, no matched commit was opened, no access-key-shaped or
listed example value was written, no `go test` was run, nothing was committed. The only file written
inside the worktree is this report.

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency** — REQ-001..REQ-013 sequential, 3-digit padding; the delta
  touches no REQ line (`spec.md` §2 absent from the diff).
- [PASS] **MP-2 EARS/GEARS format** — judged on the requirement layer (`spec.md` §2) only; the delta
  changes `spec.md` at `version` (`spec.md:4`), one HISTORY bullet (`spec.md:54-63`), a new §3.4
  sub-section (`spec.md:287-312`), and §3.4 paragraph and cell text (`spec.md:323-325`, cell ② detection,
  cell ③ scope at `spec.md:380`). No REQ changes pattern. ACs were not graded here.
- [PASS] **MP-3 YAML frontmatter** — `spec.md:2-14`: all 12 canonical fields, `version: "0.3.0"` quoted
  semver, `updated: 2026-09-11`, `priority: P1`, `lifecycle: spec-anchored`, comma-separated `tags`,
  plus `tier: M`; no rejected alias; lint exit 0.
- [PASS] **MP-4 language neutrality** — the delta names only git, the VCS the review workflow already
  uses; no programming-language tool.
- [PASS] **MP-5 D7 cross-SPEC** — only the SPEC's own ID is referenced; no BLOCKING finding.
- [PASS] **MP-6 D8 cross-platform** — `syscall` count `0` in all three files; auto-pass.
- [PASS] **MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' plan.md` exit 1; `research.md`
  absent.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | Canonical form stated once and cited everywhere (`spec.md:289-296`; `plan.md` §C item 3; `acceptance.md:200-204`, `275-281`, `310-322`). Minor: §3.2's Option 2 paragraph still gives the retired shape and calls the missing-tip case unexamined (D21); the inference is attributed to cells ① and ② although only cell ③ observes exclusion (D22) |
| Completeness | 0.90 | 0.75-1.0 | HISTORY 0.3.0 bullet, measured / inferred / gap split (`spec.md:301-312`), non-worktree gap (`plan.md:160-161`). Minor: the 0.3.0 re-gate sentence does not name the M3 move-out (D23) |
| Testability | 0.85 | 0.75-1.0 | AC-009 derivation and AC-010 A/B/L mechanics hold on a fixture, including `L` > 0 and the missing-tip exit. Minor: AC-006's token checks accept a store-not-fed mutant (D25, backstopped by AC-016); cell ③'s store snapshot is not pinned (D26) |
| Traceability | 0.90 | 0.75-1.0 | Matrix row AC-006 updated (`acceptance.md:47`); new anchor `§ Tip store and scan command` resolves; DoD and §D numbering unchanged. Minor: the design-probe citation will not resolve after the prescribed M3 move-out (D24) |

Harmonic mean 0.874 → 0.87; arithmetic mean 0.875. Tier M threshold 0.80.

### Score movement

Iteration 3 scored 0.89; this iteration scores 0.87. The two numbers grade different deltas (iteration 3:
the D13-D16 fixes; iteration 4: the caret-stdin amendment), so the drop is not a regression of the same
defect set. The LEAN rule is applied literally anyway: STOP is raised as advisory — no unconditional
further audit round. The optional findings below are for the orchestrator's discretion, not grounds for
another iteration.

## Per-check findings

### Check 1 — caret store + `--stdin` consistency

Consistent across the prescribing surfaces:

- `spec.md:289-296` — store written by `git for-each-ref --format='^%(objectname)'`, recorded before the
  scan, replaced only after the final scan exits 0; scan `git log -p --all -G '<regex>' --stdin < <tip store>`;
  missing tip looked for without the caret.
- `plan.md` §C item 3 (lines 44-48) — same store, same command, "never through a command substitution",
  missing-tip handling without the caret.
- `acceptance.md:200-204` (AC-006), `:275-281` (AC-009), `:310-322` (AC-010), `:519` (§D.18),
  `spec.md:323-325`, `spec.md:380` — all use the stdin form or the caret-stripped plain file.
- AC-016 (`acceptance.md` §D.16) is unchanged and format-agnostic: it checks the pinned lines verbatim,
  so it follows whatever the new pin says. Correct as is.

Remaining `--not` / substitution wording, classified:

| Location | Text | Class |
|---|---|---|
| `spec.md:220` (§3.2) | "An example shape is `git log -p --all --not <previous tips> -G '<regex>'`" | pre-decision option record; not prescriptive, but it sits in the adopted option's paragraph with no pointer to §3.4, and `spec.md:222-224` still says the missing-tip case is "unexamined (see `plan.md` §G)" while `plan.md:146-149` now records the probe → **D21** (optional) |
| `spec.md:306` | `git rev-list --count --all --not HEAD` | measurement narration (probe comparator); correct |
| `acceptance.md:219` | "The earlier check for `--not` … is retired" | retirement note; correct |
| `acceptance.md:319-322` | `L` = `git rev-list --count --not --all --stdin < plain` | the formula; `--not` negates `--all` only — measured correct |
| `acceptance.md:325` | `git rev-list --count --not HEAD~1 --stdin` | mechanics check; correct |
| `plan.md:151-156`, `spec.md:56-58`, `:299-303` | "command substitution" | describes the retired form as retired; correct |
| `progress.md:170` | `--not $(cat …)` pin | excluded by the brief (stale by design until re-pin) |

### Check 2 — AC-006 mutants and controls

- Positive controls are sound: a two-line scratch file gave `--stdin` `1` and `^%(objectname)` `1`, exit 0
  each; the `-F` explanation (`acceptance.md:202-204`) matches — the caret sits mid-line in the control
  and still counted.
- **Mutant passes AC-006 alone** (measured, § Re-measurements): a section whose only scan line feeds
  `--stdin` from `/dev/null`, that mentions `^%(objectname)` only as an obsolete format, and that says
  "every ref" in a display sentence reads `0 / 1 / 1 / 1 / 0` on the five checks — the passing shape.
  The checks test token presence, not that the scan line reads the store written in the caret format.
- Backstop: AC-016 (a) requires the pinned `Tip recording:` and `Scan command:` lines verbatim in
  `SECTION`, so a section that lacks them fails the joint criteria. AC-016 proves containment only
  (iteration 3, D19a), so a section carrying the pinned lines plus a contrary scan line passes both.
- Relative to the retired check: `--not` ≥ 1 had the same token-level weakness (a `--not` with no tips
  after it passed). **No regression**; strength is unchanged → **D25** (optional).

### Check 3 — AC-009 decidability

- Derivation (`acceptance.md:277-281`): measured on the fixture — plain name recorded by `rev-parse`
  before deletion; membership `grep -cxF -f` `1` over the stripped store, `0` over the caret store.
  Mechanically decidable, and the membership check prevents `<G1>` from naming a tip the store did not
  record.
- Detection (`acceptance.md:273-276`): measured — the scan through a caret store with a pruned tip exits
  128, stdout 0 bytes, stderr one `fatal: bad object` line with caret count `0` and plain-name count `1`.
  A correct procedure can therefore reach detection ≥ 1 with the stripped name, whereas the 0.2.2 wording
  ("exactly as the tip store holds it") would have read `0` for a correct procedure under a caret store.
  The amendment fixes that.
- Silent skip: the predicate (`spec.md` cell ②; `acceptance.md:282-285`) still makes exit 0 with
  `GONECELL` 0 untrustworthy, and a detection total of 0 untrustworthy. Unchanged by the delta.

### Check 4 — AC-010 formulas

- `A` = `rev-list --count --stdin < plain`, `B` = `rev-list --count --all`,
  `L` = `rev-list --count --not --all --stdin < plain`: measured correct on this repository (`A` 7082 =
  `rev-list --count HEAD` 7082 for a one-tip plain file) and on the fixture in both cases that matter —
  `L` = 0 (all tips live) and `L` = 1 (a recorded tip that exists but no ref reaches).
- `B − A + L` is the newly reachable count: the scan scope is All \ Tips, and
  |All \ Tips| = |All| − |All ∩ Tips| = B − (A − L). Fixture: scope `2` = `4 − 2 + 0` and `2` = `3 − 2 + 1`.
- The plain file is necessary, as stated (`acceptance.md:319-321`): the caret lines read directly would
  negate the tips.
- Missing tip: `A` exits 128 with `bad object` — matching "cannot be computed → re-taken, gap".
- Scope consistency (`spec.md:380`; `acceptance.md:310-312`): the form matches the scan. Which store
  file the scope and "tips excluded" readings use is not pinned in time — see **D26** (optional).

### Check 5 — inferred semantics and gaps recorded

- Recorded: `spec.md:308-310` and `plan.md:157-159` state that `git log` reading a caret store through
  `--stdin` is inferred from `git rev-list`. `spec.md:311-312` and `plan.md:160-161` record the
  non-worktree-session gap. PASS on the question asked.
- This run measured the inference on one fixture (git 2.50.1): `git log --all --stdin` over a caret store
  excluded the recorded tips (scope `2` against `4` without the store), including an annotated-tag line,
  and a command-line `--not` left a stdin tip positive under `git log` as well. This supports the
  inference; it does not replace the gate re-run.
- Attribution overclaim → **D22** (optional): cells ① and ② do not observe exclusion. Cell ①'s predicate
  (base 0 bytes, exit 0, `HEADCELL` ≥ 1, `SIDECELL` ≥ 1) holds equally for a scan that ignored every store
  line; cell ② observes that the store lines are parsed as revisions (the missing-object error), not that
  they exclude. Only cell ③'s `scope ≤ B − A + L` observes exclusion. This is a reading of the predicates,
  not a measured mutant.

### Check 6 — re-gate path after the amendment

- Stated: `spec.md:321-322` — a changed pin after the gate is a new gate round, cells ①②③ re-taken, ③
  after a fresh lead approval, before either copy is edited; `spec.md:323-325` — 0.3.0 changes the pin,
  so ① and ② are re-taken and ③ is taken for the first time on the new pin. Pin-before-gate ordering
  (`spec.md:317-320`) and approval-before-evidence (`spec.md` cell ③; `plan.md` §C item 4; AC-010) are
  unchanged and apply.
- Move-out: only `plan.md:122-127` (M3) says one commit moves the standing round's pinned procedure, lead
  approval, and gate evidence out verbatim into `.moai/reports/t629/`. AC-010 and AC-016 point to "a new
  gate round (`plan.md` M3)". Neither the 0.3.0 sentence (`spec.md:323-325`) nor the HISTORY bullet
  (`spec.md:61-63`, "then the procedure is pinned again and gate cells ① and ② are re-run") names the
  move-out, and plan M1 is not amended. The path is derivable, not stated where the 0.3.0 reader looks.
- Mechanical consequence if the move-out is skipped: AC-011 requires `^verdict: trustworthy$` count `3`
  at `R~1`; the old round's two verdict lines (`progress.md:250`, `:281`) plus a new round's three give
  `5`, and AC-011 runs at `K`, after the document edit. The miss is caught (safe direction), but late →
  **D23** (optional).
- AC-016's `G` selection works on the M3 path: the move-out commit removes `### Gate evidence` and the new
  evidence commit re-adds it, so `-S` lists the re-add as newest. It would not, if move-out and new
  evidence landed in one commit (count 1 → 1, invisible to `-S`); M3's separate commits avoid this.

### Check 7 — regressions

- REQ 13, AC 16, both unchanged in number and order; matrix row AC-006 updated to the stdin checks
  (`acceptance.md:47`); AC-009 and AC-010 rows remain accurate; §D.1-§D.19 numbering unchanged; DoD text
  unchanged (`acceptance.md:524-536`).
- Cross-references: `spec.md` § Tip store and scan command exists (`spec.md:287`) and `plan.md` §C item 3
  cites it; `acceptance.md:287-290` and `plan.md:146-149` cite a design probe in `progress.md` §E.2 that
  resolves today (`progress.md:190-205`) — see D24 for after the move-out.
- HISTORY / version: `0.3.0`, `updated: 2026-09-11`; the bullet lists AC-006, AC-009, AC-010, §D.18, and
  plan §C item 3 and §G. The §3.4 cell ③ scope wording change (`spec.md:380`) is covered by "§3.4 now makes
  …" only implicitly — not graded.
- The commit touched only the three SPEC files; both review.md copies unchanged since `feeecc980` and
  byte-identical.

## Defects Found

No blocking defect. No must-pass failure. All findings are optional under M6.

D21. SPEC32-STALE-OPTION2-SHAPE — `spec.md:219-224` — the adopted option's paragraph in §3.2 still gives
`git log -p --all --not <previous tips> -G '<regex>'` as its example shape and says the missing-tip case
is unexamined (see `plan.md` §G), while `plan.md:146-149` now records the exit-128 probe and §3.4 carries
the canonical stdin form. A reader reaching §3.2 first meets a retired command and a superseded
statement. — Severity: minor — Class: optional — Suggested fix: append one sentence to the Option 2
paragraph: "Superseded for the adopted option by §3.4 § Tip store and scan command (0.3.0)."

D22. INFERENCE-MEASURED-BY-WRONG-CELLS — `spec.md:308-310`; `plan.md:157-159` — the SPEC says the re-run of
cells ① and ② measures that `git log` reads a caret store by `rev-list` rules. Cell ①'s predicate is also
met by a scan that ignores the store; cell ② shows the lines are parsed, not that they exclude. Exclusion
is observed only by cell ③'s scope bound. Risk: a lane closing ① and ② records the inference as measured.
Cell ③ is still mandatory before any edit, so no edit rests on the gap. — Severity: minor — Class:
optional — Suggested fix: "Cells ① and ② on the new pin measure that `git log` accepts the store and
reports a missing tip; cell ③'s scope bound measures that the store's lines exclude."

D23. REGATE-MOVEOUT-NOT-NAMED — `spec.md:323-325`, `spec.md:61-63`; `plan.md` M1 (line 97) — the 0.3.0 re-gate
instruction does not name the M3 move-out commit (`plan.md:122-127`). The path is derivable through "new
gate round" → M3, but an omitted move-out surfaces only at AC-011's count at `K`, after the document edit.
— Severity: minor — Class: optional — Suggested fix: in `spec.md:323-325`, add "following `plan.md` M3:
the standing round is moved out in its own commit, the new procedure is pinned, cells ① and ② are
re-taken, and cell ③ is taken after a lead approval committed alone."

D24. PROBE-CITATION-DANGLES-AFTER-MOVEOUT — `acceptance.md:287-290`; `plan.md:146-149` → `progress.md:190-205`
— the cited design probe sits inside `### Pinned procedure` (`progress.md:162-206`), which M3 moves out
verbatim. After the prescribed re-gate the citation names a section that no longer holds it. The cited
bullet also concludes "so the store holds plain object names" — the reasoning this amendment reverses —
so a reader following the citation meets the opposite conclusion without comment. — Severity: minor —
Class: optional — Suggested fix: cite the probe by its post-move location in `.moai/reports/t629/` (or
"moved with the standing round by the M3 move-out"), and note that only its exit-128 observation is
relied on.

D25. AC006-TOKEN-NOT-LINKAGE — `acceptance.md:200-204` — measured: a section whose scan line feeds
`--stdin` from `/dev/null` and names the caret format only as obsolete passes all five AC-006 checks.
AC-016 (a) is the backstop, and it proves containment only. Same strength as the retired `--not` check,
so no regression. — Severity: minor — Class: optional — Suggested fix: count scan lines that carry both
`--stdin` and the redirect from the store, e.g. `/usr/bin/grep -c -e '--stdin <' SP/cmds.txt` → ≥ 1.

D26. AC010-STORE-SNAPSHOT — `acceptance.md:310-322`; `spec.md:374-382` — cell ③ names one store,
`SP/g3-tips.txt`, that both runs "record and read". The scope listing and "tips excluded" come after the
scan in the When-list, and the pinned procedure replaces the store after the final scan exits 0, so both
readings can read the new store rather than the one the scan read. The scope listing also substitutes the
gate's store for whatever the run actually fed, so it is a reconstruction, not an observation. Correct
procedures still pass. Nothing measured shows a false pass, but the recorded scope and excluded-tip counts
can be wrong. — Severity: minor — Class: optional — Suggested fix: before each run, copy the store the
run will read to `SP/g3-<run>-tips-read.txt` and derive the scope, tips excluded, and the plain file for
`A` / `L` from that copy.

## Regression Check

Iteration 3 closed with no blocking defects (D13-D16 resolved; N1-N3 optional and recorded in
`progress.md` §E.2 Run cautions). This delta does not touch the text those findings concern:

- D13 (AC-013/AC-014 findings file) — not in the delta; unchanged.
- D14 (AC-016 tip-recording line) — AC-016 not in the delta; `plan.md` §C item 3 still requires the
  `Tip recording: ` line and now specifies its caret form. No regression.
- D15 (AC-001 construction readings) — not in the delta; unchanged.
- D16 (AC-004 anchor at `K`) — not in the delta; unchanged.
- N1-N3 — not in the delta; no change in status.

## Out of scope, not scored

- Cell ② detection (`spec.md` cell ②, pre-existing since 0.2.1): the detection reading is satisfied by
  git's own stderr line (fixture: `err:1` with no procedure involvement), so it does not show that the
  procedure itself reported the tip as missing. The amendment keeps that property.
- `progress.md` §E.2's gate-execution method and gaps describe the retired form; they are expected to
  move out with the standing round. Not graded, per the brief.
- Git stops at the first bad object (`progress.md:290-291`), so recovery path (b) — drop the missing tip and
  rescan — needs a loop over the error for a store with several missing tips. The canonical stdin form does
  not change this. Unexercised.

## Recommendation

PASS at 0.87 against the Tier M threshold 0.80. All seven must-pass criteria pass with the evidence above:

- MP-1: REQ-001..013 unchanged, no gap, no duplicate.
- MP-2: requirement layer untouched by the delta.
- MP-3: 12 canonical fields plus `tier`, `version: "0.3.0"` (`spec.md:2-14`), lint exit 0.
- MP-4-MP-7: no language tool, no foreign SPEC reference, no `syscall`, no clarification marker, no
  `research.md`.

The canonical caret store + `--stdin` form is specified consistently across §3.4, plan §C item 3,
AC-006, AC-009, AC-010, and §D.18, and its git semantics hold on a fixture: caret lines through `--stdin`
exclude under `git log`, a command-line `--not` leaves stdin tips positive, `B − A + L` equals the stdin
scope with `L` = 0 and `L` = 1, and a pruned tip exits 128 naming the plain object name.

D21-D26 are optional, one- or two-sentence changes. The two most useful before the re-pin are D23 (name
the move-out where the 0.3.0 re-gate is stated) and D22 (attribute the exclusion measurement to cell ③).
This verdict does not bypass the Implementation Kickoff Approval gate or the fresh lead approval cell ③
requires.

## Gaps

- Mutant and semantics checks ran on one scratch fixture and one git build (2.50.1, Apple Git-155); the
  real pinned procedure does not exist yet on the new form.
- The AC-006 mutant was a hand-built scratch section, not an edited `TPL`.
- Lint ran on an installed build whose commit was not captured; the binary changed (rc.6 → rc.7) during
  this run.
- Remote-tracking refs, `refs/stash`, and lightweight tags were not in the fixture store; one annotated
  tag was.
- The worktree-guard refusals named in the amendment were not re-measured by this auditor. The guard did
  refuse compound git commands in this run, which is consistent with the refusal claim; no
  command-substitution git form was attempted.
- `spec.md` and `acceptance.md` were read by region, not in full; the delta shows no change elsewhere.
- Cross-model audit backends were not invoked (no `audit_model` instruction in the brief).

## Residual-risk

The stdin form's exclusion is measured here on a fixture, not by the gate. Cells ① and ② on the new pin
will not observe it (D22), so until cell ③ runs, exclusion rests on this fixture and the orchestrator's
`rev-list` probes. The re-gate depends on a move-out step named only in plan M3 (D23). If it is skipped,
the miss shows up only after the document edit. AC-006 and AC-016 together still prove containment of
the prescribed lines, not exclusivity (D25, iteration 3 D19a).
