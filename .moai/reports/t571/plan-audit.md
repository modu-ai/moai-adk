# SPEC Review Report: SPEC-CODEX-HOME-BACKSLASH-001

Iteration: 1/1 (Tier S ceiling = 1 per `harness.plan_audit_tier_ceilings`)
Verdict: **FAIL**
Overall Score: **0.66** (Tier S PASS threshold = 0.75)

Reasoning context ignored per M1 Context Isolation. The card brief was read for *audit scope* (which
dimensions to test) only; every judgment below is made against the artifact files and the code, and
every PASS cites a command or a line.

Audit tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t571` (`git rev-parse --show-toplevel`).
Artifacts read (Tier S input contract + the card's split `acceptance.md`): `spec.md`, `plan.md`,
`acceptance.md`, `progress.md`.

**Process note — a foreign write landed mid-audit.** `acceptance.md` changed on disk between my first
and second read: `_maps REQ-CHB-001, REQ-CHB-002_` was inserted at AC-CHB-001 (line 11). The change
is correct in itself and I have graded the post-edit state, but per `agent-common-protocol.md`
§ Background Agent Execution an actively audited worktree has exactly one writer. Reporting it rather
than quietly absorbing it. The traceability finding D5 below is scored against the current file — the
edit repaired one AC's mapping and left five unmapped.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-CHB-001` … `REQ-CHB-007` at spec.md:L77-L83, sequential,
  no gaps, no duplicates, uniform 3-digit zero-padding. Tier S REQ ceiling is 8; 7 declared.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in
  `spec.md`), never against an AC. All seven match a GEARS pattern:
  L77 `The path classifier shall assign…` (Ubiquitous); L78 `**When** a declared skill path … shall
  classify…` (Event-driven); L79 Ubiquitous; L80 `**When** the prune verb judges…` (Event-driven);
  L81 `The classifier shall not change…`; L82 `The classifier shall not run…` (Unwanted, canonical
  `shall not`); L83 `**While** a test overrides a package-level seam …, that test shall not call…`
  (State-driven). The six `Given/When/Then` entries in `acceptance.md` are the verification layer and
  are graded under Group 4, not here. One label defect noted at D8 (severity minor) — L81 is tagged
  `(Ubiquitous)` but its form is `shall not` = Unwanted. The *form* is valid GEARS either way, so MP-2
  is not failed by it.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present at spec.md:L2-L13
  (`id`, `title`, `version` quoted `"0.1.0"`, `status: draft`, `created`/`updated` ISO,
  `author`, `priority: P1`, `phase: "v3.2.0 target"` — a release label, not a lifecycle stage —
  `module`, `lifecycle: spec-anchored`, `tags` comma-string). No rejected snake_case alias. Optional
  `tier: S` + `depends_on` present. Mechanically confirmed:

  ```
  $ moai spec lint .moai/specs/SPEC-CODEX-HOME-BACKSLASH-001/spec.md
  ✓ No findings — all SPEC documents are valid
  EXIT=0
  ```

  **Empty-set control run** (the green above is attributed, not vacuous):

  ```
  $ moai spec lint /tmp/definitely-not-a-spec.md
  ERROR  ParseFailure  /tmp/definitely-not-a-spec.md  1  SPEC parsing failed: failed to read file…
  EXIT=1
  ```

  The linter reports on an unreadable input rather than passing it, so `EXIT=0` on the real file means
  the file was read and judged. See D9 for the author's stated lint gap, which this closes.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go; `module: "internal/cli"`), no
  multi-language tooling surface. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — one external reference,
  `SPEC-CODEX-SKILL-PATH-SLASH-001` (spec.md:L15, L26; plan.md:L74).
  `grep '^status:' .moai/specs/SPEC-CODEX-SKILL-PATH-SLASH-001/spec.md` → `status: implemented`.
  Not in `{retired, superseded, archived}` → no BLOCKING finding. A **separate, non-D7** consequence of
  that same value is reported at D3.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall` over all four artifacts returns
  `0 0 0 0`. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-HOME-BACKSLASH-001/`
  returns exactly one line, `progress.md:12: - 미해결 [NEEDS CLARIFICATION] 0건`, which is a
  zero-claim sentence, not an unresolved topic marker (no `: <topic>` payload). No open marker. Nit at
  D11: that sentence trips the gate's own grep and will read as a hit to the next mechanical consumer.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | spec.md:L34-L42 states the four-step ordering and names the asymmetry as the defect, not the backslash; L63-L67 sizes the damage precisely and refuses the unmeasured claim ("정보 유출은 관측하지 않았고 주장하지 않는다"); L97-L99 gives a reasoned rejection with a named breakage. Deduction: REQ-CHB-006 (L82) is stated as a requirement about a function that never expands a home (D6), and AC-CHB-005 quotes a predicate expression where it means a function (D2). |
| Completeness | 0.80 | 0.75-1.0 | HISTORY L20-L26, WHY/context §1, WHAT/approach §3, REQUIREMENTS §2, ACCEPTANCE (split file, 6 AC + DoD), and four `### Out of Scope — <topic>` H3 headings at L112/L117/L122/L126 each carrying specific `-` bullets. Deduction: the adjacent `~/../..` escape shape is absent from all four (D1). |
| Testability | 0.50 | 0.50 | Half the ACs can pass while their requirement is violated or are not mechanically decidable: AC-CHB-005 is satisfied by a mutant that re-implements the predicate (D2); AC-CHB-003's load-bearing row never goes red under any change this SPEC contemplates (D4); AC-CHB-006's "newly-SKIP 0건" has no baseline to compare against (D7). AC-CHB-001/003/004 are precisely binary — acceptance.md:L18-L19 pins `--- PASS: <name>` rather than a bare exit code, which defends against a zero-match `-run` selector. |
| Traceability | 0.50 | 0.50 | REQ-CHB-006 and REQ-CHB-007 have no AC at all; AC-CHB-005 maps to no REQ; only AC-CHB-001 carries an explicit `_maps_` line (D5). Rubric: "Multiple REQs lack ACs" = 0.50. |

Aggregate (mean) = (0.85 + 0.80 + 0.50 + 0.50) / 4 = **0.6625 → 0.66**, below the Tier S threshold 0.75.

**Where the FAIL comes from.** No must-pass criterion failed. The verdict is the rubric score, and it is
driven by two dimensions only — testability and traceability — both of which are cheap to lift. The
defect analysis in §1 is the strongest part of this SPEC and none of the findings below touch it.

---

## Answers to the six directed audit questions

**1. TWO-ARM DISCRIMINANT — HOLDS.** acceptance.md:L15 requires *both* assertions and names (a)
`got(homeArm) == got(relArm)` as the body of the AC, explicitly because the defect is an asymmetry.
This is exactly the shape that survives the degenerate future the question asks about: if a later
change made everything oddly-formed, assertion (a) still passes but (b) is unaffected and the
preserved-cell AC (AC-CHB-003 rows 1, 3, 4) goes red on three cells. The SPEC did **not** degrade into
a one-sided assertion — plan.md:L12 records the author reasoning through precisely this hazard
("부재 가드의 함정") and choosing the symmetry form because of it. No finding.

**2. VACUOUS-GREEN HAZARDS — three found, two blocking.**
- *Zero-match `-run` selector*: **defended** on AC-001 (`--- PASS: TestCodexSkillPathBackslashSymmetry`
  required in output), AC-003 (four subtests all PASS, SKIP 0), AC-004/005 (`--- PASS` required). A
  selector matching nothing prints `ok … [no tests to run]` with no `--- PASS` line, so every one of
  these criteria fails on an empty sweep rather than passing. Good.
- *Absence-satisfied-by-nonexistence*: no grep-based AC exists in this SPEC. The absence-shaped
  criterion is AC-CHB-004's "`osStatFn` call count **exactly 0**", and acceptance.md:L58 anticipates the
  hazard in the right direction — it explains why `Eligible=false` alone is insufficient. But a
  call-count-zero assertion is satisfied by a test that never reaches the code at all; the counter must
  be proven live by a positive control (a clean home-relative path driving the count to exactly 1).
  → **D10**, minor.
- *Mutant-satisfiable AC*: AC-CHB-005 → **D2**, blocking.
- *Never-red AC*: AC-CHB-003 row 1 → **D4**, blocking.
- *Non-decidable criterion*: AC-CHB-006 → **D7**, blocking.
- Pass criteria are strict rather than "no error" throughout — every AC pins an exit code AND an exact
  output token. That part is well done.

**3. MUTANT AC REACHABILITY — REACHED. The mutant is real.** Verified by construction, not assumed:
`classifyCodexSkillPath` (doctor_codex.go:667) is called directly by the named test with a literal
string; there is no layer between the mutation site and the assertion, so the "unreached mutant prints
the same `ok`" failure mode is structurally unavailable here. Confirmed against Go semantics — on
darwin `filepath.IsAbs(p)` is `stringslite.HasPrefix(path, "/")`
(`$GOROOT/src/internal/filepathlite/path_unix.go:35-37`), so `~/x\SKILL.md` is not absolute and the old
ordering at doctor_codex.go:670-675 peels `~/` first and returns `codexPathHomeRelative` (=1), while the
AC asserts `codexPathOddlyFormed` (=3). The test goes red. Shape constants confirmed against the
`iota` block at doctor_codex.go:644-660 and cross-checked against the repro log's `shape=1` / `shape=3`.
One defect in how the mutant AC is *written*: → **D6**, minor.

**4. THE AUTHOR'S OWN DERIVATION — CORRECT, AND THE AC ENCODES IT.** Both halves verified against the
stdlib source, not from memory:
- `C:\Users\x` on darwin: `IsAbs` is `HasPrefix(path, "/")` → **false**. It therefore falls to the
  backslash branch under the current ordering AND under the proposed ordering AND under a
  backslash-hoisted-above-IsAbs ordering — `codexPathOddlyFormed` in all three. It discriminates
  nothing on this host. The author's claim is right.
- `/tmp/a\b`: `IsAbs` **true** → `codexPathAbsolute` under current and proposed; `codexPathOddlyFormed`
  the moment the backslash check is hoisted above `IsAbs`. It is the only cell on a POSIX host that
  separates those two orderings. The author's claim is right.
- acceptance.md:L43 encodes exactly this path with exactly this expectation, and L48 states the
  reasoning in the artifact. Encoded correctly.
- The residual is not the *choice* of discriminant but its *adoption*: it is never observed red. → **D4**.

**5. SCOPE — one real omission, no creep.**
- *Creep*: none. plan.md §F touches one function in one file (doctor_codex.go) plus one new test file;
  §D forbids new seams and new shape constants; spec.md:L122-L124 explicitly walls off the write side.
  The consumer claim at spec.md:L58-L61 is **verified**, not taken on trust:
  `grep -rn classifyCodexSkillPath internal` returns exactly the two production call sites the SPEC
  names, at the two line numbers it cites (`codex_skills_prune.go:76`, `doctor_codex.go:831`).
- *Omission*: the `~/../../etc/x` shape is **not declared out of scope anywhere** — I read all four
  `### Out of Scope` headings and neither `..`, nor `Clean`, nor "home escape" appears in any of them.
  → **D1**, blocking.

**6. UNVERIFIED PREMISES — the SPEC is unusually disciplined here.** I checked every asserted-as-measured
statement:
- spec.md:L46-L52 quotes the repro log and labels it "인용, 재유도 아님" — the quoted lines match
  `.moai/reports/t571/repro-asymmetry.log` byte-for-byte, including `--- PASS` and `EXIT`-equivalent
  `ok`. Attributed.
- spec.md:L58-L61 consumer table — verified above, correct.
- spec.md:L71 write-side guard — verified: `strings.ContainsAny(skillPath, "\"\\\n\r")` is at
  `codex_skills_disable.go:262` as claimed, inside `upsertCodexSkillDisable`, and the comment block
  above it (L250-L261) does say what the SPEC says it says, including the "prune … deletes" clause.
- spec.md:L26 / progress.md:L10 explicitly mark the t540 landing state as *narration, not re-verified in
  this card* — a correctly-labelled gap rather than an unobserved claim.
- spec.md:L67 refuses to claim information disclosure it did not observe.
- One genuine unverified premise remains: spec.md:L103's "이 칸이 즉시 붉어진다" is an assertion about a
  test's behaviour under a change nobody ran. → folded into **D4**.

---

## Defects Found (structured defect-list)

**D1.** OOS-HOME-ESCAPE — `spec.md:L110-L128` — The `~/../../etc/x` shape is silently omitted rather than
declared out of scope. This is not hypothetical: on this host `filepath.Join("/Users/goos", "../../etc/x")`
resolves through `Clean` (`$GOROOT/src/path/filepath/path_unix.go:33-40` — `join` calls `Clean`) to
`/etc/x`, so a home-relative declaration escapes the home, is stat'ed, and reaches `Eligible=true` by the
same route this SPEC is closing — **and it still does after M2**, because it carries no backslash. A
reader of §4 will reasonably conclude that home-relative escapes were considered and handled. Three
Out-of-Scope headings exist for narrower topics; the absence of this one is the load-bearing gap.
— Severity: **major** — Class: **blocking** — Required fix: add a fourth `### Out of Scope — 홈-상대
선언의 `..` 이탈` heading naming the shape, stating that `filepath.Join` cleans `..` so the declaration
escapes the home, that this survives M2 unchanged, and naming the follow-up card that owns it.

**D2.** AC005-PREDICATE-COPY — `acceptance.md:L68` — AC-CHB-005 names the write-side refusal as the
*expression* `strings.ContainsAny(skillPath, "\"\\\n\r")` rather than as the *function* that evaluates
it. A test that satisfies this AC by re-writing that expression inline asserts a copy against itself and
would stay green after the production predicate was deleted — the mutant probe of
`verification-completeness.md` §2 is writable, so the criterion is too shallow to adopt. The production
entry point exists and is in-package: `upsertCodexSkillDisable(content []byte, skillPath string)` at
`codex_skills_disable.go:231`, returning a verdict whose `SkipReason` is the observable.
— Severity: **major** — Class: **blocking** — Required fix: rewrite AC-CHB-005 to call
`upsertCodexSkillDisable` with a config fixture and assert its returned `codexSkillDisableVerdict`
carries a skip; forbid inline re-implementation of the predicate in the AC text.

**D3.** DEP-NOT-COMPLETED — `spec.md:L15` (`depends_on`) — The dependency's measured status is
`implemented`, not `completed`. Per `spec-workflow.md` § Depends_on Pre-flight Check, fulfillment is
**strictly** `completed` — "all other 7 status values … are considered unfulfilled", "no partial credit".
The Phase 1 pre-flight therefore blocks `/moai run` on this SPEC with a 3-option `AskUserQuestion` and,
on the override path, requires the unfulfilled dependency ID plus a rationale logged to
`.moai/logs/depends-on-override.log`. The SPEC's HISTORY discusses the *branch* landing state at length
and never mentions the *status-field* consequence, so the run session meets an unannounced gate.
— Severity: **major** — Class: **blocking** — Required fix: add one HISTORY or §B line stating the
dependency reads `status: implemented`, that this is unfulfilled by the strict predicate, and which
disposition the card takes (wait / logged override with rationale).

**D4.** AC003-NEVER-RED — `acceptance.md:L43, L48` and `spec.md:L103` — The `/tmp/a\b → codexPathAbsolute`
row is declared "이 AC의 핵심" and "유일한 판별식", but it is green on the pre-change tree, green after
M2, and green under AC-CHB-002's mutant (the old ordering keeps `IsAbs` first). No step in the workflow
observes it red, so under `verification-completeness.md` §1.1 it is a check whose failure has never been
seen, and §2's two-cell rule leaves it with a green path and no RED-now cell. The claim at spec.md:L103
that hoisting the check "즉시 붉어진다" is an unverified premise (audit question 6). The fix is one extra
mutant and costs nothing.
— Severity: **major** — Class: **blocking** — Required fix: add AC-CHB-007 — a second mutant that hoists
`strings.ContainsRune(p, '\\')` **above** `filepath.IsAbs`, asserts `TestCodexSkillPathPreservedShapes`
goes red naming the `/tmp/a\b` cell, logs to `.moai/reports/t571/mutant-isabs-hoist.log`, restores the
source, and re-runs green. Without it, plan.md:L68's anti-pattern is guarded by an unproven claim.

**D5.** REQ-AC-COVERAGE — `spec.md:L82-L83`, `acceptance.md` (whole file) — Two requirements have no
acceptance criterion. REQ-CHB-006 (no backslash check against an expanded home) is asserted by nothing;
worse, it is **unfalsifiable on this host** — the nearest cell, `~/ok/SKILL.md` at acceptance.md:L46,
stays `codexPathHomeRelative` whether or not the check runs on the expanded path, because
`/Users/goos` contains no backslash, so that row would pass over a violating implementation.
REQ-CHB-007 (seam tests non-parallel + `t.Cleanup`) is carried only by a DoD checkbox at
`acceptance.md:L96`, which is not a command. Separately, AC-CHB-005 maps to no REQ, and after the
mid-audit edit only AC-CHB-001 carries a `_maps_` line while five do not.
— Severity: **major** — Class: **blocking** — Required fix: (a) give REQ-CHB-006 an AC that is capable of
failing — assert against a stubbed `codexUserHomeDir` returning a backslash-bearing home (e.g.
`/tmp/ho\me`) that `~/ok/SKILL.md` still classifies `codexPathHomeRelative`; that cell goes red the moment
the check moves after expansion; (b) give REQ-CHB-007 a mechanical AC (a grep over the new test file
asserting 0 `t.Parallel()` and ≥1 `t.Cleanup` in the seam-overriding subtests) or delete the requirement
and keep it as a plan-level constraint; (c) add `_maps REQ-…_` lines to AC-002 through AC-006, and either
map AC-CHB-005 to REQ-CHB-002 or add the read/write-symmetry requirement it actually tests.

**D6.** AC002-THEN-UNENCODED — `acceptance.md:L27` vs `L30` — The **Then** requires the failure message to
name the observed value as `codexPathHomeRelative`, but the pass criterion below it requires only
`EXIT != 0` plus a `--- FAIL:` line. The gap is not cosmetic: `codexSkillPathShape` is a bare
`int` type with **no `String()` method** anywhere in the package
(`grep -n "codexSkillPathShape)" internal/cli/*.go` returns nothing), so a `%v` of the observed value
prints `1`, not the constant name, and a test written to the letter of the criterion satisfies it while
the Then goes unmet.
— Severity: **minor** — Class: **blocking** (cheap, and it is the mutant AC — the one criterion the SPEC
itself designates as the anti-vacuity backstop) — Required fix: either add the constant-name requirement
to the pass criterion and have the test map the shape to a name in its failure message, or drop the
naming clause from the Then so the two agree.

**D7.** AC006-NO-BASELINE — `acceptance.md:L83` — "새로 `SKIP` 으로 바뀐 기존 테스트 0건" is not decidable
from the run it prescribes: "newly SKIP" is a delta, and no pre-change SKIP count is measured, pinned, or
named anywhere in the SPEC. As written, the criterion is satisfied by recording any number and asserting
it is the same as an unrecorded one.
— Severity: **minor** — Class: **blocking** — Required fix: add a pre-M2 baseline step — run
`go test ./internal/cli/... -count=1 -timeout 600s -v` on the unmodified tree, record the SKIP count and
the HEAD SHA it was measured on into `.moai/reports/t571/`, and restate AC-CHB-006 as an exact equality
against that pinned number.

**D8.** REQ005-MISLABEL — `spec.md:L81` — REQ-CHB-005 is tagged `(Ubiquitous)` but its form is
`The classifier shall not change…` — the GEARS canonical **Unwanted** pattern, the same form correctly
tagged at L82. The requirement's form is valid GEARS either way, so MP-2 is unaffected; the label is what
is wrong.
— Severity: **minor** — Class: **optional** — Required fix: retag L81 as `(Unwanted)`.

**D9.** LINT-GAP-CLOSED — `progress.md` / author's stated gap — The author reported being unable to run
`moai spec lint` (no `--spec` flag; a corpus-wide run hit EXIT=124 at 280s over 818 SPECs) and fell back
to hand-grepping frontmatter and Out-of-Scope conformance. A bounded invocation does exist: the file path
is a **positional** argument (`moai spec lint [spec.md...]`). I ran it (evidence under MP-3): clean,
EXIT=0, in well under the timeout, with an empty-set control proving the green is attributable. The
conformance the author could only hand-check is now mechanically verified. Two scope notes on that green:
it covers `spec.md` only, and the linter's advertised "AC→REQ coverage (100% required)" rule therefore
never saw the split `acceptance.md` — passing `acceptance.md` or `plan.md` to the linter is a category
error (both correctly carry no frontmatter per the artifact-statelessness rule) and returns
`ParseFailure`. So the lint pass says nothing about D5, which is why D5 stands on manual analysis.
— Severity: **minor** — Class: **optional** — Required fix: record the bounded invocation
`moai spec lint <path-to-spec.md>` and its EXIT=0 in `progress.md §E.1`, replacing the hand-grep note.

**D10.** AC004-COUNTER-UNPROVEN — `acceptance.md:L56, L58` — "`osStatFn` 호출 횟수가 정확히 0" is a
call-count-zero assertion, and a counter that is never incremented in the test cannot be distinguished
from a counter that is wired to nothing. The AC reasons correctly about why `Eligible=false` alone is
insufficient but does not extend that reasoning to the counter itself.
— Severity: **minor** — Class: **optional** — Required fix: add a positive control to the same test — a
second sub-case driving a clean `~/ok/SKILL.md` through `judgeCodexSkillEntry` and asserting the count is
exactly 1 — so the zero in the primary case is a measurement rather than an absence.

**D11.** MP7-GREP-COLLISION — `progress.md:L12` — The zero-claim sentence `미해결 [NEEDS CLARIFICATION] 0건`
contains the literal token the MP-7 gate greps for (`\[NEEDS CLARIFICATION`), so the gate's own recipe
reports a hit on a SPEC that has no marker. I resolved it by reading the line; the next mechanical
consumer may not.
— Severity: **minor** — Class: **optional** — Required fix: reword to avoid the bracketed literal, e.g.
`미해결 clarification 마커 0건`.

**D12.** ADJACENT-GUARD-UNNAMED — `plan.md:L20-L25` (§D 제약) — An existing source-text guard,
`TestSinglePathShapeClassifier` at `internal/cli/codex_skills_prune_test.go:605-628`, fails any non-test
file under `internal/cli` or `internal/codexwiring` that contains both `filepath.IsAbs` and the literal
`HasPrefix(p, "~` **unless** that file also contains `func classifyCodexSkillPath`. The M2 rewrite at
plan.md:L44-L48 preserves both the location and the literal form, so it does not break — I checked. But
the constraint is invisible to the implementer: extracting the classifier into a helper file, or changing
`HasPrefix(p, "~/")` to a different spelling, turns this guard red for a reason plan.md never mentions.
AC-CHB-006 would catch it late; §D naming it would prevent it.
— Severity: **minor** — Class: **optional** — Required fix: add a §D constraint naming
`TestSinglePathShapeClassifier` and the two literals it keys on.

---

## Regression Check

Not applicable — iteration 1.

---

## Recommendation

Five blocking findings, all cheap, none touching the defect analysis or the two-arm discriminant. In
order:

1. **D1** — add the fourth `### Out of Scope` heading for the `~/../..` home-escape shape, stating that it
   survives M2 unchanged and naming the follow-up owner. (`spec.md` §4)
2. **D5** — close the two uncovered requirements. REQ-CHB-006 needs an AC that can actually fail: stub
   `codexUserHomeDir` to a backslash-bearing home and assert `~/ok/SKILL.md` still classifies
   home-relative. REQ-CHB-007 needs a grep-shaped AC or demotion to a plan constraint. Add `_maps_` lines
   to AC-002…AC-006 and give AC-CHB-005 a requirement to map to. (`spec.md:L82-L83`, `acceptance.md`)
3. **D4** — add the second mutant (backslash hoisted above `IsAbs`) so AC-CHB-003's load-bearing row is
   observed red once. This also converts `spec.md:L103` from an assertion into a measurement.
4. **D2** — rewrite AC-CHB-005 to call `upsertCodexSkillDisable` (`codex_skills_disable.go:231`) rather
   than quoting the predicate at line 262.
5. **D3** — state the `depends_on` disposition: the dependency reads `status: implemented`, the pre-flight
   predicate is strictly `completed`, and this card either waits or takes the logged override.
6. **D6** and **D7** — align AC-CHB-002's Then with its pass criterion (no `String()` method exists), and
   pin AC-CHB-006's SKIP baseline to a measured pre-M2 number and its HEAD SHA.

D8-D12 are optional and left to the orchestrator's discretion; D9 in particular is a *closed* gap
(evidence in this report) and needs only a `progress.md` correction rather than new work.

Once 1-6 land, the traceability band moves to 1.0 (all seven REQs covered, all six ACs mapped) and
testability to 0.75-1.0 (no mutant-satisfiable AC, no never-red load-bearing cell, no undecidable
criterion), which puts the aggregate comfortably above the Tier S 0.75 threshold. Tier S ceiling is one
iteration, so the confirming re-audit is scoped to this enumerated defect delta.

Rule applications on this verdict, per `verification-completeness.md` §7:
`.claude/rules/moai/development/verification-completeness.md` §1.1 (D4, D10), §2 mutant probe (D2, D4),
§2.1 (D7); `.claude/rules/moai/core/verification-claim-integrity.md` §1 (audit question 6, D4);
`.claude/rules/moai/workflow/spec-workflow.md` § Depends_on Pre-flight Check (D3);
`.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields +
§ Artifact Statelessness (MP-3, D9); `.claude/rules/moai/core/agent-common-protocol.md`
§ Background Agent Execution (the mid-audit foreign write).
