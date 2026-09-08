# SPEC Review Report: SPEC-CODEX-HOME-BACKSLASH-001

Iteration: 2/2
Verdict: **PASS**
Overall Score: **0.88** (Tier S PASS threshold = 0.75)

Reasoning context ignored per M1 Context Isolation. The card brief and the coordinator's mid-audit
message were read for *audit scope* and *attribution obligations* only; every judgment below is made
against the artifact files and the code, and every PASS cites a command or a line.

---

## 0. Tree integrity — opening and closing digests

[HARD] first act, `shasum -a 256 .moai/specs/SPEC-CODEX-HOME-BACKSLASH-001/*.md`:

```
da68c0dbe550bb6f0fccf5b04cc9856e53d49808cf10a34e62657ee62e099c4f  acceptance.md
71d08a2d412e80a047292be65c1c9d7a185d355aa75eb0cbd82cc227fa869e8a  plan.md
07de922fea9482a8d6fd42f4d8e3a358ac7a9743e8fadb671d7008bb2170c5ce  progress.md
2abec682c7e0673284aa61a0f3130ed49a20ff6ec81e4b9a9d73e48d8ee13baa  spec.md
```

**All four match the supplied baseline byte-for-byte.** Closing digests are in §9 and are identical.
**No second writer entered the tree during this audit window.**

Revision graded: the four files at the digests above, in worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t571`, branch `WT-codex-home-backslash`,
HEAD `ee194493f`, working tree carrying only two untracked paths (`.moai/reports/t571/`,
`.moai/specs/SPEC-CODEX-HOME-BACKSLASH-001/`).

---

## 1. [HARD] Iteration 1 graded a STALE tree — required attribution

Per the lead's ruling: iteration 1's `FAIL 0.66` was produced against a revision that predated the
authoring agent's second write pass. The first agent's completion notification fired **before** it
stopped writing, so it was still editing `acceptance.md` while iteration 1 was reading it. Iteration
1's own mid-audit note ("`acceptance.md` changed on disk between my first and second read") is what
pinned the cause down; that self-report was correct.

**The 0.66 → 0.88 improvement MUST be split, or it reads larger than it is.**

**(a) What the repairs actually fixed** — twelve items, all verified landed in §3:
D1 (fifth Out-of-Scope heading), D2 (AC-CHB-005 rewritten onto the production entry point),
D3 (`depends_on` logged-override disposition), D4 (AC-CHB-009 + §3.3 converted from assertion to
citation), D5(c) (anchored `_maps` lines across all nine ACs), D6 (AC-CHB-002 Then/criterion
alignment), D7 (pinned SKIP baseline file), D8 (REQ-CHB-005 retag), D9 (bounded lint record),
D10 (AC-CHB-004 counter positive control), D11 (MP-7 grep collision removed), D12 (plan.md §D
adjacent-guard constraint).

**(b) What iteration 1 marked as a defect only because it was reading an older revision** —
**D5(a) and D5(b)**. Per the lead, `AC-CHB-007` (sole coverage for REQ-CHB-006) and `AC-CHB-008`
(sole coverage for REQ-CHB-007) were **already on disk, unseen**, at the moment iteration 1 graded.
Iteration 1 reported "REQ-CHB-006 and REQ-CHB-007 have no AC at all" and counted six ACs; the file now
carries nine.

**I cannot independently verify (b), and I say so rather than adopting it as measured.** Both artifact
files are untracked (`git status --short` → `?? .moai/specs/SPEC-CODEX-HOME-BACKSLASH-001/`), so no
prior revision exists in git to diff against, and no pre-edit digest of that revision was captured.
The split above is recorded **on the lead's authority as a process finding**, not as something this
audit observed. It is stated because omitting it would let the traceability move 0.50 → 1.00 read as
repair work when part of it was never broken at the moment it was graded.

**The split cuts both ways.** The two ACs that existed-but-unseen are precisely the two carrying this
iteration's two heaviest new findings (N1, N3). Their first real audit is in §4 — nobody had graded
them before, in either direction.

This attribution does not soften the audit. The current files are graded adversarially below.

### Baseline-attribution and a deliberately-open gap

Local and remote `develop` have both moved to `a4855f0b2` since this worktree last absorbed — t540
(this card's dependency), t562, and t577 have landed, and t540 is now on the remote too. The lead is
**deliberately not merging** while this audit window is open (one writer per audited tree). Every
judgment in this report is therefore attributed to **HEAD `ee194493f` without that absorb**. The
re-absorb happens after this report returns. One finding falls directly out of that gap — see **N9**,
which is the heaviest new finding in this report.

---

## 2. Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-CHB-001` … `REQ-CHB-007` at spec.md:L90-L96,
  sequential, no gaps, no duplicates, uniform 3-digit padding. Measured set (`grep -o` over §2)
  equals `{001,002,003,004,005,006,007}`.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in
  `spec.md`), never against an AC. L90 Ubiquitous; L91 `**When** … shall classify` Event-driven;
  L92 Ubiquitous; L93 Event-driven; L94 `shall not change` Unwanted (correctly retagged, D8 landed);
  L95 `shall not run` Unwanted; L96 `**While** … shall not call` State-driven. The nine
  Given/When/Then entries in `acceptance.md` are the verification layer and are graded under Group 4,
  not here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields at spec.md:L2-L13, no rejected
  snake_case alias, optional `tier: S` + `depends_on` present. Mechanically confirmed, and the green
  is attributable:

  ```
  $ moai spec lint .moai/specs/SPEC-CODEX-HOME-BACKSLASH-001/spec.md
  ✓ No findings — all SPEC documents are valid
  EXIT=0

  $ moai spec lint /tmp/not-a-spec-xyz.md          # empty-set control
  ERROR  ParseFailure  …  SPEC parsing failed: failed to read file … no such file or directory
  EXIT=1
  ```

  The control establishes the linter reports on an unreadable input rather than passing it, so
  `EXIT=0` on the real file means the file was read and judged.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go; `module: "internal/cli"`).
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — one external reference,
  `SPEC-CODEX-SKILL-PATH-SLASH-001`. `grep -n '^status:'` → `5:status: implemented`. Not in
  `{retired, superseded, archived}` → no BLOCKING finding. The separate, non-D7 consequence of that
  same value is D3, verified landed in §3.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall` over all four artifacts returns
  `0 0 0 0`. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — the gate's own recipe, run verbatim:

  ```
  $ grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-HOME-BACKSLASH-001/
  EXIT=1        # zero matches
  ```

  Iteration 1's D11 (the zero-claim sentence tripping the gate on itself) is **closed and
  mechanically re-verified** — progress.md:L14 now states the count without the bracketed literal and
  explains why.

No must-pass failure. The verdict is the rubric score.

---

## 3. Regression check against iteration 1 — all twelve findings verified independently

I did not accept the repairing agent's counters. Each row below is a measurement.

| # | Status | Evidence |
|---|---|---|
| **D1** | **RESOLVED** | Fifth `### Out of Scope` heading at spec.md:L141, `홈-상대 선언의 `..` 이탈 (`~/../../etc/x`)`, with four specific bullets. Its two load-bearing claims are both **correct**: (i) `filepath.Join` calls `Clean` — verified at `$GOROOT/src/path/filepath/path_unix.go:36`, `return Clean(strings.Join(...))`; (ii) **the shape survives M2 unchanged** — M2 hoists only `ContainsRune(p,'\\')` and this declaration carries no backslash, so the reorder cannot touch it. L144 states this explicitly and warns against the "considered and handled" misreading. One defect *inside* this fix → **N4**. |
| **D2** | **RESOLVED** | acceptance.md:L81 now calls the production entry point `upsertCodexSkillDisable(content []byte, skillPath string)`; signature verified at `codex_skills_disable.go:231`. L82 asserts on the returned `codexSkillDisableVerdict` (`Action == codexSkillDisableSkipped`, non-empty `Reason`, content unchanged) rather than on a re-written expression. **The appended "test body contains no `ContainsAny`" criterion is decorative, not load-bearing** → **N8**; the load-bearing part is the production call, and that is sound. |
| **D3** | **RESOLVED, and the rationale is sound** | spec.md:L28-L39. The premise is verified (`grep -n '^status:'` → `implemented`); the strict-`completed` predicate is stated correctly; the disposition is a **logged override** with the unfulfilled dependency ID and a rationale, and L39 makes the log entry a *precondition* of run-phase entry rather than an afterthought. **The stated rationale — that the dependency is on t540's code rather than its administrative state — holds.** The code is in this tree: I read `classifyCodexSkillPath` at `doctor_codex.go:667` and it carries the post-t540 four-shape `iota` block and the `ContainsRune(p,'\\')` branch. Only the status *field* lags. That is exactly the sanctioned override path, not a bypass. |
| **D4** | **RESOLVED** | AC-CHB-009 added at acceptance.md:L140, and spec.md:L118 now **cites** it — "이 문장은 그 AC 를 인용할 뿐 결과를 미리 단언하지 않는다" — instead of asserting "즉시 붉어진다". The conversion from assertion to citation is the correct repair shape. Mutant reachability verified in §4.3. Residual → **N7** (minor). |
| **D5** | **RESOLVED structurally** | Measured: `grep -c '^_maps ' acceptance.md` = **9**; `grep -c '^## AC-CHB-' acceptance.md` = **9**; the union of mapped REQs = exactly the seven declared in spec.md §2. No orphan AC, no uncovered REQ. (a)/(b) attribution per §1. The *substantive* weakness of the REQ-006/REQ-007 coverage is graded in §4 and scored under testability, not double-counted here. |
| **D6** | **RESOLVED** | acceptance.md:L33 now requires the string `codexPathHomeRelative` in the mutant output, aligning the criterion with the Then. The premise is re-verified: `grep -n "codexSkillPathShape)" internal/cli/*.go` → 0 matches, so no `String()` method exists and a `%v` prints `1`; L31 states this and requires the test to carry the mapping. **The same discipline was NOT applied to the two new ACs** → **N1, N3**. |
| **D7** | **RESOLVED, strongly** | `.moai/reports/t571/skip-baseline-pre-m2.log` exists. `SKIP_COUNT=30`; `MEASURED_ON_HEAD=ee194493ff14…` — **equals the current HEAD**; `EXIT=0`; `SKIP_COUNT_EXPR` pinned with the `^ *` anchor and a written rationale for why an unanchored `^--- SKIP:` would report a vacuous 0; 30 names enumerated (measured: `grep -c '^--- SKIP:'` = **30**, internally consistent); and an explicit host/environment portability caveat. AC-CHB-006 is now exact-equality to a pinned number. One factual slip inside → **N6**; one structural gap the baseline does not cover → **N9**. |
| **D8** | **RESOLVED** | spec.md:L94 now `(Unwanted)`. |
| **D9** | **RESOLVED** | progress.md:L15-L23 records the bounded positional invocation and its `EXIT=0`, **and** states the scope limit correctly — the linter parses only the passed `spec.md`, so its `AC→REQ coverage` rule never saw the split `acceptance.md`. I re-ran it myself (§2 MP-3) with a control. |
| **D10** | **RESOLVED, strongly** | acceptance.md:L68-L69 adds the positive control (clean `~/ok/SKILL.md` → count exactly 1), requires the two sub-cases to use **separate** counters with `t.Cleanup` restore so ordering cannot contaminate the 0, and L72 raises the criterion to 2 subtests / 4 assertions / SKIP 0. This is the best-executed repair in the set. |
| **D11** | **RESOLVED, mechanically re-verified** | progress.md:L14 reworded; the gate's own grep now returns EXIT=1 / zero matches (§2 MP-7). |
| **D12** | **RESOLVED** | plan.md:L26-L28 adds a `[HARD]` constraint naming `TestSinglePathShapeClassifier`, both literals it keys on, and the two forbidden refactors (extraction; respelling `HasPrefix(p, "~/")`). Line citation verified: `grep -n 'func TestSinglePathShapeClassifier'` → **`:606`**, which is what plan.md cites. |

**No previously-passing dimension got worse.** Every rubric dimension moved up (§7). But two *hazard
classes* the repair removed were **reproduced one section away in the new material** — the D6
Then/criterion mismatch (at AC-CHB-007 and AC-CHB-008) and the D4 no-RED-cell shape (at AC-CHB-007).
That is a qualitative regression in discipline consistency even though no score fell, and it is why
this report carries new blocking findings despite a PASS.

---

## 4. The three things nobody had graded

### 4.1 AC-CHB-007 — sole coverage for REQ-CHB-006

**The construction claim the repairing agent flagged as unverified: I verify it, and it is SOUND. The
cell CAN fail.** Established from the production code plus the stdlib source, not from memory:

- `filepath.IsAbs` on darwin is `stringslite.HasPrefix(path, "/")`
  (`$GOROOT/src/internal/filepathlite/path_unix.go:35-37`). So `C:\Users\x` is **not** absolute here,
  and `~/ok/SKILL.md` is not either.
- Under the adopted M2 ordering, `~/ok/SKILL.md` carries no backslash and matches the `~/` branch →
  `codexPathHomeRelative` → `judgeCodexSkillEntry` (`codex_skills_prune.go:76-90`) calls
  `expandCodexHomeRelativePath`, which is `filepath.Join(home, p[2:])` over the stubbed home →
  `Clean("C:\Users\x" + "/" + "ok/SKILL.md")`. On a `/`-separator host `\` is an ordinary byte, so the
  result is `C:\Users\x/ok/SKILL.md` — **backslash-bearing**. Then `osStatFn` runs exactly once.
- Under the rejected design (check moved *after* expansion), that expanded path is refused →
  oddly-formed → `skip(...)` → **stat count 0**. Both assertions redden.

So the discriminant is real and the AC is not passing for an unrelated reason at the *design* level.
Two defects at the *criterion* level:

- **N1 (major, blocking) — the stub is never observed.** The Then (L117) requires stat to be called
  **"확장 경로로"** — with the expanded path. The pass criterion (L122) drops the path and asks only
  for `stat 호출 계수 == 1`. **The entire discriminating power of this AC lives in the stubbed home**,
  and a count of 1 is produced identically whether the `codexUserHomeDir` override took effect or was
  silently inert (the real home also yields exactly one stat). A test that stubs the wrong seam, or
  whose `t.Cleanup` ordering restores the seam early, passes this criterion while exercising nothing
  the AC is about. This is the D10 hazard class — a counter that cannot be distinguished from an
  unwired counter — reproduced at the one AC where the repair did not apply it.
- **N2 (minor, blocking) — no prescribed RED-now cell.** L119 asserts "검사를 확장 뒤로 옮기면 이 AC가
  즉시 붉어진다". No step in this SPEC ever moves the check there — the design was *rejected*, so the
  mutant is never injected. This is textually the same assertion the repair just removed from
  spec.md:L103 by adding AC-CHB-009, reproduced one AC away. AC-CHB-007 is green pre-M2, green
  post-M2, and green under both prescribed mutants.

### 4.2 AC-CHB-008 — sole coverage for REQ-CHB-007

**N3 (major, blocking) — not mechanically decidable, and its command does not measure its criterion.**

- The command (L134) is `grep -c "t.Parallel()" internal/cli/codex_skills_path_shape_test.go` — a
  **whole-file** count. The pass criterion (L135) is explicitly **per-function** and explicitly
  *refuses* the whole-file count ("판정은 파일 전체 계수가 아니라 시접 재대입 함수 단위로 읽는다").
  The AC therefore prescribes a command that cannot produce its own verdict, and adjudication falls to
  a human reading the file. **For a binding AC that is not acceptable**: the Group 4 AC-2 criterion is
  "a tester can determine PASS/FAIL without judgment calls", and this one is a judgment call by
  construction.
- Three further gaps: (i) the Then also requires each seam function to carry `t.Cleanup`, and the pass
  criterion says **nothing** about `t.Cleanup` — a second Then/criterion mismatch of the D6 class;
  (ii) the criterion is a pure absence assertion with **no positive control** proving the expression
  can ever be non-zero on this file, and **no exit-code pin**, so a missing file, a renamed test, or a
  test file that never overrides a seam satisfies it — the repository's own standing rule is that an
  absence guard cannot be adopted on a green-only observation; (iii) `"t.Parallel()"` is an unanchored
  regex in which `.` matches any character.

The fix is cheap and there are two routes: make it decidable with an `awk`-bounded per-function
extraction plus an exit pin and an injected positive control, **or** simplify the requirement to "zero
`t.Parallel()` anywhere in this new file" — which the already-written command decides as-is.

### 4.3 AC-CHB-009 — mutant reachability

**REACHED. An unreached mutant and a real survivor print the same `ok`, so this was checked by
construction rather than assumed.**

`TestCodexSkillPathPreservedShapes` calls `classifyCodexSkillPath` **directly** with literal strings
(AC-CHB-003's table); there is no layer between the mutation site (`doctor_codex.go:667-678`) and the
assertion, so the unreached-mutant failure mode is structurally unavailable. The mutant hoists
`strings.ContainsRune(p,'\\')` above `filepath.IsAbs`. Effect, cell by cell: `/tmp/a\b` is
`IsAbs`-true today (`HasPrefix(p,"/")`) → `codexPathAbsolute`, and under the mutant matches the
backslash branch first → `codexPathOddlyFormed` → **that row reddens**. The other three rows
(`~someone/…`, `rel/…`, `~/ok/…`) carry no backslash and are untouched, so exactly one cell goes red —
which is what the AC asserts. AC-CHB-009 genuinely establishes §3.3 invariant 1.

One residual: **N7 (minor, optional)** — the criterion (L152) requires the output to contain `/tmp/a\b`
**and** `codexPathAbsolute` anywhere. Under `-v`, every subtest name prints regardless of outcome, so
if the subtest is named after the path the first conjunct is satisfied by the run *existing*, not by
that cell failing. The discriminating conjunct is `codexPathAbsolute` alone. The strict form pins
`--- FAIL: TestCodexSkillPathPreservedShapes/<the subtest naming that cell>`.

---

## 5. Vacuous-green sweep — all nine ACs

| AC | Zero-match `-run` defended? | Absence-shape? | RED-now cell |
|---|---|---|---|
| 001 | YES — requires `--- PASS: TestCodexSkillPathBackslashSymmetry` | no | AC-002 mutant |
| 002 | YES — requires `--- FAIL: <name>` + the constant-name token | no | is the mutant |
| 003 | YES — 4 subtests PASS, SKIP 0 | no | **AC-009 mutant** (was the iteration-1 gap; closed) |
| 004 | YES — 2 subtests `--- PASS`, SKIP 0 | count-0, but **positive control added** | inherits the AC-002 mutant (old ordering stats once → count 1 ≠ 0). Not stated in the SPEC — one free line, optional |
| 005 | YES — requires `--- PASS: <name>` | appended `ContainsAny`-0 grep → **N8** | body calls production; guard-deletion mutant present but marked optional. Acceptable |
| 006 | n/a — exact equality to a pinned 30 | no | any new SKIP |
| 007 | partial — criterion says bare `--- PASS` without naming the test. A zero-match selector still prints no `--- PASS` at all, so the hole does not open; sloppy, fold into N1's fix | count-1, **stub unobserved** → **N1** | **none** → **N2** |
| 008 | n/a — no `go test` | **pure absence, no exit pin, no positive control** → **N3** | **none** → **N3** |
| 009 | YES — requires `--- FAIL: <name>` | no | is the mutant |

Pass criteria are strict rather than "no error" throughout — every `go test` AC pins an exit code AND
an exact output token. Iteration 1's three vacuity findings (D2 mutant-satisfiable, D4 never-red, D7
undecidable delta) are all closed. The two surviving holes are both in the two ACs iteration 1 never
saw.

### Self-counting hazard (D11 class) — sibling sweep

The agent's self-caught fix is **verified working**. DoD acceptance.md:L161 uses the anchored
`grep -c '^_maps '`. Measured: anchored = **9** (= AC count), unanchored `grep -c '_maps '` = **10** —
the extra match is the DoD line's own text, exactly the contamination the anchor removes.

I swept all four files for siblings of that shape — a line whose own text contains the token it counts,
against the target the criterion actually names. **None found.** Checked: progress.md:L14 (counts the
bracketed marker in the SPEC dir; the line deliberately omits the bracket — re-verified by running the
gate grep, EXIT=1); acceptance.md L33/L89/L134/L152 (each counts a token in *test output* or a *test
file*, not in `acceptance.md` — no self-collision); the baseline log's `SKIP_COUNT_EXPR=` line (does
not match its own `^ *--- SKIP: ` expression).

---

## 6. Defects Found (structured defect-list)

**N1.** AC007-STUB-UNOBSERVED — `acceptance.md:L117 vs L122` — The Then requires `osStatFn` to be
called **with the expanded path**; the pass criterion requires only `계수 == 1`. The whole
discriminating power of this AC is the `codexUserHomeDir` → `C:\Users\x` stub, and a count of 1 is
produced identically when the stub is inert (the real home also yields exactly one stat). The AC that
is the *sole* coverage for REQ-CHB-006 therefore cannot distinguish "the backslash-bearing home
expanded correctly" from "the override never took effect". D10 hazard class, reproduced.
— Severity: **major** — Class: **blocking** — Required fix: assert the *observed stat argument*, not
just the count — expected `C:\Users\x/ok/SKILL.md` (verified: `filepath.Join` on a `/`-separator host
leaves `\` as an ordinary byte, so the expansion is backslash-bearing) — and put that value in the
pass criterion. While there, name the test in the `--- PASS` token as every other AC does.

**N2.** AC007-NO-RED-CELL — `acceptance.md:L119` — "검사를 확장 뒤로 옮기면 이 AC가 즉시 붉어진다" is
an assertion about a change no prescribed step performs; the design it describes was *rejected*, so
the mutant is never injected. AC-CHB-007 is green pre-M2, green post-M2, and green under both
prescribed mutants. This is textually the same shape the repair just removed from spec.md:L103.
— Severity: **minor** — Class: **blocking** — Required fix: either add a third mutant (move the check
below `expandCodexHomeRelativePath` and observe AC-CHB-007 red, log, restore, re-run green), **or** —
cheaper and equally honest — restate L119 in the citation form §3.3 now uses: a reasoned expectation
that is explicitly labelled as not-yet-observed. Do not leave it phrased as an observation.

**N3.** AC008-NOT-DECIDABLE — `acceptance.md:L134-L135` — The command is a whole-file `grep -c`; the
pass criterion is per-seam-function and explicitly refuses the whole-file count. The AC prescribes a
command that cannot produce its own verdict, so adjudication requires a human reading the file. Plus:
the Then requires `t.Cleanup` and the criterion never mentions it; the criterion is a pure absence
assertion with no exit-code pin and no positive control, satisfied by a missing file or by no seam
test existing at all; and `"t.Parallel()"` is an unanchored regex.
— Severity: **major** — Class: **blocking** — Required fix: pick one route. (a) Make it decidable:
`awk`-bounded per-function extraction, pin the exit code, add a `t.Cleanup` ≥1-per-seam-function
conjunct, and add a positive control (inject one `t.Parallel()` into a seam function, show the
expression reports ≥1, revert). (b) Simplify the requirement to "zero `t.Parallel()` anywhere in
`codex_skills_path_shape_test.go`" — the already-written command decides that as-is, and the file is
new so the cost of forgoing parallelism in it is nil.

**N9.** AC006-BASELINE-VS-ABSORB — `acceptance.md:L97-L99` vs `plan.md:L11` — **The pinned 30 does not
survive the mandatory re-absorb, and the two documents contradict each other on this.** plan.md §B
requires absorbing `origin/develop` in the merge window and **re-measuring in the merge tree**;
AC-CHB-006 requires exact equality to a number measured pre-absorb at `ee194493f`, and its Given
mentions only "M2 의 순서 재배치가 적용된 트리". This is now concrete rather than hypothetical:
`develop` has moved to `a4855f0b2` and among the landed cards **t577 edits `internal/cli/todo.go`** —
inside the very package AC-CHB-006 counts. The baseline file's own caveat covers *host/environment*
drift and is silent on *tree* drift. As written, AC-CHB-006 will read a legitimate post-absorb count
as a regression, or invite re-pinning the number without a rule for doing so.
— Severity: **major** — Class: **blocking** — Required fix: state the re-pin rule in AC-CHB-006. On
absorbing `origin/develop`, re-measure the baseline in the merge tree with the same command and the
same `^ *` expression, write a second pinned record naming the new HEAD, and make the exact-equality
comparison against **that** record; the `ee194493f`/30 pin remains valid only for a pre-absorb run.
Keep the existing "이름이 무엇인지가 판정" clause — it is what makes a delta diagnosable either way.

**N4.** TODO-QUERY-UNATTRIBUTED — `spec.md:L146` — "`moai todo` 전수 조회 … 매치 0건" is not
re-derivable, and the named invocation demonstrably does not perform an exhaustive query. Measured:
`moai todo` prints `list: 28 rows withheld — showing 20 of 48` plus `27 dropped (hidden)`; the
exhaustive form is `moai todo list --limit 0` (`moai todo --limit 0` is rejected: `Unknown flag`). No
command form, output, or row count is cited. This is the repository's standing truncated-listing
hazard reproduced inside a fix iteration 1 asked for. **The conclusion itself survives — I re-derived
it**: over the full 48-row listing the SPEC's pattern matches **0** case-sensitively; the single
case-insensitive hit is `moai worktree clean` in an unrelated card, a false positive.
— Severity: **minor** — Class: **blocking** (evidence integrity; one-line fix) — Required fix: cite
`moai todo list --limit 0`, the matched count (0), and the rows-scanned figure, so the claim is
attributable rather than remembered.

**N5.** STALE-LINE-CITATION — `spec.md:L84` **and** `plan.md:L83` — Both cite
`codex_skills_disable.go:262` for the write-side refusal predicate. Measured: `grep -n 'ContainsAny'
internal/cli/codex_skills_disable.go` → **`261:	if strings.ContainsAny(skillPath, "\"\\\n\r") {`**;
line 262 is the `return skip(...)` body. The predicate measures at 261. **Iteration 1 repeated `262`
as verified — that was an iteration-1 error, corrected here.** The sibling citation in the same AC
(`upsertCodexSkillDisable` at `:231`) is **correct**, as is plan.md's `:606` for
`TestSinglePathShapeClassifier`.
— Severity: **minor** — Class: **blocking** (two one-token fixes; a line citation decays like a HEAD
reading, and this one is cited twice) — Required fix: change both to `:261`.

**N6.** BASELINE-SUBTEST-MISCOUNT — `.moai/reports/t571/skip-baseline-pre-m2.log` and
`acceptance.md:L103` — Both state "3건이 하위 테스트". Measured: `grep '^--- SKIP:' … | grep -c '/'` =
**4** (`TestLegacySkillIDsNotEmbedded/manifest_empty`, `…/manifest_error`,
`TestMakeSymlink_SkipsWhenCreationFails/collision`,
`TestResolveRulesDir/windows_volume-letter_gate.yaml_value_passes_through_unjoined`). The binding
number (30) is unaffected and the 30 enumerated names are internally consistent; the parenthetical is
wrong in the evidence file that exists to make the criterion attributable.
— Severity: **minor** — Class: **optional** — Required fix: 3 → 4 in both places.

**N7.** AC009-WEAK-CONJUNCT — `acceptance.md:L152` — The criterion asks for `/tmp/a\b` and
`codexPathAbsolute` anywhere in the output. Under `-v` the subtest name prints regardless of outcome,
so the first conjunct is satisfied by the run existing rather than by that cell failing.
— Severity: **minor** — Class: **optional** — Required fix: pin
`--- FAIL: TestCodexSkillPathPreservedShapes/<subtest naming the /tmp/a\b cell>`.

**N8.** AC005-GREP-DECORATIVE — `acceptance.md:L89` — "테스트 파일의 해당 함수 본문에 `ContainsAny`
문자열이 0회" carries no command, is read per-function (the N3 decidability gap again), and is evadable
by spelling the copy differently (`strings.IndexAny`, a manual rune loop). It is not a completeness
guard. This does **not** re-open D2: the load-bearing repair is the call to `upsertCodexSkillDisable`,
and that is sound.
— Severity: **minor** — Class: **optional** — Required fix: drop it, or replace with the same
`awk`-bounded form chosen for N3.

**N10.** AC004-INHERITED-MUTANT-UNSTATED — `acceptance.md:L59-L72` — AC-CHB-004 does have a RED-now
cell: under AC-CHB-002's old-ordering mutant, `~/x\SKILL.md` classifies home-relative, expands, and
stats — so the count-0 assertion reddens in that same window. The SPEC does not say so, leaving the
count-0 looking unbacked when it is not.
— Severity: **minor** — Class: **optional** — Required fix: one line in AC-CHB-004 noting it is also
observed red inside the AC-CHB-002 mutant window.

---

## 7. Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Iter 1 | Iter 2 | Band | Evidence |
|---|---|---|---|---|
| Clarity | 0.85 | **0.88** | 0.75-1.0 | Improved: §3.3 invariant 1 converted from assertion to citation (spec.md:L118); the `depends_on` branch-landing axis and status-field axis explicitly separated (L30); §4's fifth heading states "고려됐고, **처리되지 않았다**" (L144), pre-empting the exact misreading iteration 1 predicted. Deductions: two stale line citations (N5), an unattributed measurement (N4), a wrong subtest count in the evidence file (N6) — precision defects in a document whose thesis is precision. |
| Completeness | 0.80 | **0.95** | 0.75-1.0 | All required sections; five `### Out of Scope — <topic>` H3 headings each carrying specific `-` bullets; the adjacent `..`-escape shape now named with an owner disposition and a pinned issuance text; a new pinned baseline evidence file. Deduction: the follow-up owner card is unissued — defensible (card issuance is the operator's act) and documented as such. |
| Testability | 0.50 | **0.70** | between 0.50 and 0.75 | Four of iteration 1's five testability defects are genuinely closed, and the two mutants plus the AC-004 positive control are strong work. Held below 0.75 because **two of nine ACs are not adoptable as written and both are the sole coverage for their REQ**: AC-CHB-008 requires a judgment call by construction (N3), AC-CHB-007's criterion cannot observe the stub its whole discriminating power rests on (N1). No weasel words anywhere; every `go test` AC pins exit + token. |
| Traceability | 0.50 | **1.00** | 1.0 | Measured, not asserted: 9 anchored `_maps` lines = 9 AC headings; mapped-REQ union = the seven declared; no orphan AC, no uncovered REQ. Rubric 1.0 exactly. The *substantive* weakness of the REQ-006/007 coverage is scored under testability and deliberately not double-counted here. |

Aggregate (mean) = (0.88 + 0.95 + 0.70 + 1.00) / 4 = **0.8825 → 0.88**, against the Tier S threshold
**0.75**.

**Robustness of the verdict.** The PASS does not depend on my testability judgment. Holding
testability at iteration 1's 0.50 gives 0.8325; holding testability at 0.50 *and* clarity at 0.75
gives 0.80. Both still clear 0.75. The verdict is stable across the full plausible range of the
dimension I graded most severely.

**Score regression check (LEAN clause).** iter1 0.66 → iter2 0.88 is an increase, so no `STOP`
signal. Per §1, part of that increase is attribution rather than repair.

**Regression against iteration 1: no dimension declined.** The qualitative regression — two repaired
hazard classes reproduced in new material — is recorded in §3 and priced into testability.

---

## 8. Recommendation

**PASS.** No must-pass failure; aggregate 0.88 well clear of the Tier S 0.75 threshold, and stable
under stress. The defect analysis in §1, the two-arm discriminant, the two mutants, and the
now-complete REQ↔AC mapping are all sound, and the `depends_on` override is correctly reasoned and
correctly gated on a log entry.

The four **blocking** findings are enumerated for the orchestrator to route **before run-phase
entry** — a PASS verdict is not a reason to carry them into implementation, because each one is a
verification-design defect that would produce a green run proving less than it appears to:

1. **N9** — heaviest, and the one the moved `develop` makes urgent. State AC-CHB-006's re-pin rule;
   as written it contradicts plan.md §B and will misread a legitimate post-absorb count.
2. **N3** — make AC-CHB-008 decidable, or simplify the requirement to the whole-file form its own
   command already decides. It is the sole coverage for REQ-CHB-007.
3. **N1** — assert the stat *argument* in AC-CHB-007, not just the count, so the stub is observed. It
   is the sole coverage for REQ-CHB-006.
4. **N4**, **N5**, **N2** — one-line each: cite `moai todo list --limit 0` with its measured 0;
   correct `:262` → `:261` in both spec.md and plan.md; and either add the third mutant for
   AC-CHB-007 or restate L119 in the citation form §3.3 already uses.

N6-N8 and N10 are optional and left to the orchestrator's discretion.

**Two process items for the record, not for the SPEC author.** (i) Iteration 1 graded a stale
revision because a completion notification preceded the writer's last write; the delta split is in §1
and the (b) half is recorded on the lead's authority because I could not verify it — the artifacts are
untracked, so no prior revision exists to diff. (ii) This audit graded `ee194493f` without the
`a4855f0b2` absorb, at the lead's instruction; N9 is the finding that falls out of that gap, and the
post-absorb re-measurement it prescribes is the thing to do first after the merge window opens.

Rule applications on this verdict:
`.claude/rules/moai/development/verification-completeness.md` §1.1 (N1, N2, N3), §2 mutant probe
(§4.3, N2), §2.1 (N3); `.claude/rules/moai/core/verification-claim-integrity.md` §1 (N2, N4), §2
baseline attribution (N4, N9, §1); `.claude/rules/moai/workflow/spec-workflow.md` § Depends_on
Pre-flight Check (D3); `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12
Required Fields (MP-3); `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent
Execution (§0, §1 — one writer per audited tree, held this window).

---

## 9. [HARD] Closing digest — last act

`shasum -a 256 .moai/specs/SPEC-CODEX-HOME-BACKSLASH-001/*.md`, re-run after all grading:

```
da68c0dbe550bb6f0fccf5b04cc9856e53d49808cf10a34e62657ee62e099c4f  acceptance.md
71d08a2d412e80a047292be65c1c9d7a185d355aa75eb0cbd82cc227fa869e8a  plan.md
07de922fea9482a8d6fd42f4d8e3a358ac7a9743e8fadb671d7008bb2170c5ce  progress.md
2abec682c7e0673284aa61a0f3130ed49a20ff6ec81e4b9a9d73e48d8ee13baa  spec.md
```

**Opening = closing = supplied baseline, all four files. No digest moved during this audit. No second
writer entered the tree.** The revision graded is the one named in §0.
