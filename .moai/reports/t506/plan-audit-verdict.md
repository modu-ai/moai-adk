# SPEC Review Report: SPEC-CODEX-GHOST-SKILLS-PRUNE-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.76** (Tier M PASS threshold 0.80 — `spec-workflow.md:141`)

Auditor: plan-auditor (leaf worker; no SendMessage, no AskUserQuestion used).
Reasoning context ignored per M1 Context Isolation — every judgment below is
made against the four artifacts and the cited sources, not against the author's
description of them.

Tree: `.claude/worktrees/t506`, branch `WT-codex-ghost-skills`, HEAD `ace1c5440`.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-CGP-001 … REQ-CGP-017`, 17 ids,
  sequential, no gap, no duplicate, uniform 3-digit padding.
  Evidence: `grep -o 'REQ-CGP-[0-9]*' spec.md | sort -u` → the 17 consecutive ids.

- **[FAIL] MP-2 GEARS format compliance (requirement layer)** — 15 of 17 REQ
  entries carry a GEARS pattern; **two do not**.
  - `spec.md:65` REQ-CGP-004 — "항목은 다음 세 조건을 모두 만족할 때에만 제거 적격
    이다" is a *definition* ("an entry IS eligible when…"), carrying no
    `shall`/`해야 한다` modality and no actor. It states a predicate, not a
    behaviour the system shall exhibit.
  - `spec.md:69` REQ-CGP-005 — "`enabled` 플래그는 적격 판정의 게이트가 **아니다**"
    is likewise a negative definition, not the GEARS Unwanted form
    ("시스템은 … 해서는 안 된다"), which every other never-prune REQ (006-011)
    does use correctly.
  Judgment layer: this verdict is made against the **REQ-XXX requirement layer**
  in `spec.md`. The `Given/When/Then` entries in `acceptance.md` are the
  verification layer and were NOT graded here (M3 § Scope).
  Cheapest fix on this report: rewrite as "시스템은 다음 세 조건을 모두 만족하는
  항목만 제거해야 한다" / "시스템은 `enabled` 값을 적격 판정의 게이트로 사용해서는
  안 된다". Substance unchanged.

  **Attribution caveat (do not read the linter's green as covering this).**
  `moai spec lint` (built from this tree, `go build -o /tmp/t506-audit-moai
  ./cmd/moai`, rc 0) reports `✓ No findings`, rc 0. That green is **vacuous on
  the modality axis for a Korean SPEC**: `internal/spec/lint.go:790`
  `isModalityMalformed` only fires on text whose prefix is `WHEN `/`WHILE `/
  `WHERE `/`IF `/`THE ` — Korean REQ text matches no prefix and returns false
  unconditionally. The lint run is therefore evidence about frontmatter and REQ
  ids, not about GEARS modality. Control run proving the linter did read the
  file (non-vacuous on the axes it does cover): a copy with `lifecycle:` removed
  yields `WARNING FrontmatterInvalid … missing: lifecycle`.

- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with
  correct types (`spec.md:1-16`): `id, title, version("0.1.0"), status(draft),
  created/updated(2026-09-07), author, priority(P2), phase, module, lifecycle
  (spec-anchored), tags`. No rejected snake_case alias. `tier: M` and
  `related_specs` are extra, not violations.
  The id `SPEC-CODEX-GHOST-SKILLS-PRUNE-001` is multi-segment; it matches the
  **enforced** pattern `internal/spec/lint.go:952`
  `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`. `progress.md:8`'s self-check claim is
  therefore correct. (Observation, not a defect of this SPEC: the SSOT prose at
  `.claude/rules/moai/development/spec-frontmatter-schema.md` states the narrower
  `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`, which this id would fail. Doc/code
  divergence in the rule file, out of this card's scope.)

- **[N/A] MP-4 language neutrality** — the SPEC is scoped to this repository's own
  Go source (`internal/cli`, `internal/codexwiring`), not to template-bound or
  multi-language tooling. Auto-passes.

- **[PASS] MP-5 D7 cross-SPEC reconciliation** — every referenced SPEC exists and
  none is retired/superseded/archived: `SPEC-CODEX-SKILL-PATH-001`,
  `SPEC-CODEX-WIRING-001`, `SPEC-V3R6-MOAI-CLEAN-HOME-001`,
  `SPEC-CODEX-LAUNCHER-001`, `SPEC-CODEX-SKILL-LOADER-001` — all
  `status: completed`. No BLOCKING finding.

- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`.
  Auto-PASS per D8-4. (See Residual-risk: the *indeterminate* fixture is a
  platform-dependent construct even though the word `syscall` never appears.)

- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION'` over the
  SPEC directory returns rc 1, no match. `research.md` does not exist (Tier M
  artifact set is 3 files + progress); the marker check is satisfied on
  `plan.md`.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | REQ-CGP-001..017 each admit one reading; the two soft spots are `plan.md:52` (`EndLine` decides blank lines but leaves comments and non-`path`/`enabled` keys undecided) and `acceptance.md:138` ("실행 전후" names no boundary). |
| Completeness | 0.75 | 0.75-1.0 | All sections present; `§D` carries five `### Out of Scope — …` H3 sub-headings each with specific bullets (`spec.md:101,106,110,114,118`). Deducted for the missing command-surface requirement (D3 below) and for the un-specified entry-internal content cases (D2). |
| Testability | 0.65 | 0.50-0.75 | No weasel words; `acceptance.md:34` explicitly forbids an empty-operand PASS — good. But two ACs can pass while asserting nothing (D1, D4), which is the exact hazard the SPEC lectures about elsewhere. |
| Traceability | 0.80 | 0.75-1.0 | `§D.2` maps all 17 REQs to ACs. One orphan: AC-CGP-012 (`acceptance.md:95`) appears in no row of the REQ→AC table and no REQ requires it (D3). |

Aggregate (mean) = **0.76**. Tier M threshold 0.80 → below threshold independently
of the MP-2 failure.

---

## Defects Found

**D1. AC-CGP-013 can pass without exercising reassembly at all** —
`acceptance.md:101-105` — Severity: **critical** — Class: **blocking**.
The lossless round-trip is the criterion the whole byte-preservation design rests
on (`plan.md:58` says so: "이 무손실 왕복이 성립하지 않으면 §C 의 접근 자체를 다시
정해야 한다"). As written it is: *given a config with **0 eligible entries**, run
the pruner with `--force`, the file is byte-identical.* Any correct implementation
short-circuits when there is nothing to remove — REQ-CGP-017 (`spec.md:95`)
positively **requires** the no-change path for the 0-entry case — so the file is
byte-identical because **nothing was written**, and `splitLines`→reassembly is
never executed. It is a vacuous operand of exactly the shape `acceptance.md:34`
prohibits one paragraph earlier.
Compounding, the underlying loss is real and knowable **now**:
`internal/codexwiring/configtoml.go:204-209` `splitLines` does
`strings.Split(strings.TrimSuffix(body, "\n"), "\n")`, so `"a\n"` and `"a"`
produce the identical slice `["a"]`. Whichever rejoin the implementation picks,
one of the two trailing-newline states comes out byte-wrong — and CRLF input
loses its final `\r\n` the same way. `plan.md:58` defers this to "픽스처로 확인"
without recording the answer, and the AC that was supposed to catch it cannot.
Required fix: (a) restate AC-CGP-013 as a **unit** criterion on the
parse→reassemble function (not the end-to-end `--force` run), and (b) require
**both** fixtures — one config ending with a trailing newline, one without —
plus a CRLF case, each byte-identical after a round trip. Then state in
`plan.md` that `splitLines` alone is insufficient and the trailing-newline state
must be carried separately.

**D2. Span semantics are under-specified for entry-internal content, and
`plan.md:49`'s claim about the parser is factually wrong** —
`plan.md:49`, `plan.md:52` — Severity: **major** — Class: **blocking**.
`plan.md:49` states the parser "헤더에서 항목을 열어 다음 테이블 헤더에서 닫는
extent 개념을 **이미 갖고 있다** — 그 extent 를 버리지 않고 기록하기만 하면 된다".
Read against `internal/codexwiring/skills.go:106-143`, that is not what the parser
holds. It holds a *close event* (`inEntry = false` at `skills.go:131`) and nothing
else: it records no line index, and it never observes which lines belong to the
entry. Three concrete cases where a span derived as the plan describes would be
wrong:
  1. **Multi-line literal inside an entry.** `skills.go:113-125` `continue`s
     every line while `openDelim != ""`, bypassing the switch entirely. An entry
     carrying `notes = """…"""` contributes lines the entry logic never sees; a
     span that stops at the last *matched* key line truncates the literal and
     leaves an orphan closing `"""` behind after deletion.
  2. **Keys other than `path` / `enabled`.** `skills.go:135-139` matches only
     those two; every other assignment inside the entry is dropped on the floor.
     Deleting header+matched-keys re-parents the surviving assignment onto the
     *preceding* table — a silent semantic corruption of valid TOML, which
     REQ-CGP-015 (`spec.md:90`) forbids in spirit and no AC detects.
  3. **A line that `anyTableRe` mis-reads.**
     `internal/codexwiring/configtoml.go:76` is
     `^\[\[?[^\]]*\]\]?\s*(#.*)?$`, so a continuation line such as
     `["x", "y"]` inside an entry satisfies it and closes the extent early.
  **Entry at EOF** (asked about explicitly): this one is safe **provided** the
  span logic treats end-of-input as a close — the parser simply exits the loop
  with `inEntry` still true (`skills.go:141`), so nothing goes wrong, but
  `plan.md` never states that EOF closes an extent.
Required fix: in `plan.md §B`, replace the "already has it" sentence with the
actual obligation — the span must track **every** line from the header until the
close event *including* lines consumed by the `openDelim` branch, then trim
trailing blank lines — and decide, explicitly, the disposition of (a) comment
lines between the last key and the close, and (b) non-`path`/`enabled` keys
inside the entry. Then add an AC pairing each: an eligible entry containing a
multi-line literal, and one carrying an extra key.

**D3. AC-CGP-012 is an orphan — no requirement backs the command surface** —
`acceptance.md:95-99`, `acceptance.md:117-130` — Severity: **major** — Class:
**blocking**.
AC-CGP-012 tests that `moai clean --home --codex-skills` is rejected. It appears
in no row of the `§D.2` REQ→AC table, and `spec.md §C` carries no requirement
about the verb's surface, its flags, or their exclusivity — the decision lives
only in `plan.md:28`. An AC with no requirement is an untraced test; conversely
the surface decision (a destructive flag on an existing command) is exactly the
kind of thing a requirement should pin, since `plan.md` is revisable without
touching the SPEC.
Required fix: add a REQ (e.g. REQ-CGP-018, Unwanted: "시스템은 `--home` 과
`--codex-skills` 가 동시에 지정된 호출을 실행해서는 안 된다") and map it in `§D.2`.

**D4. AC-CGP-004's mutant does not have to fail the *right* test** —
`acceptance.md:36-44` — Severity: **major** — Class: **blocking**.
The must-pass criterion requires only that `go test ./internal/cli/...` goes RED
after the mutant, with a GREEN observation before it. The GREEN/RED pair proves
the mutant *caused* a failure; it does not prove the failure came from the
pruner's never-prune boundary. `internal/cli` also contains the doctor's own
indeterminate handling (`doctor_codex.go:442-450`), and a mutant applied there —
or a pruner mutant that happens to trip an unrelated doctor golden — turns the
suite RED while the boundary this SPEC is defending stays untested. That is a
vacuous pass of the one AC the SPEC declares must-pass.
Required fix: name the failing test in the AC — "the RED must include the
never-prune indeterminate case of AC-CGP-003 class 4, identified by test name in
the verdict record" — and require the mutant's diff site to be the pruner's own
eligibility branch.

**D5. `plan.md §F` offers `skip` as a mitigation for the very fixture the
must-pass AC depends on** — `plan.md:86` vs `acceptance.md:111` — Severity:
**major** — Class: **blocking**.
The risk row says the indeterminate fixture may be unportable (permission-denied
does not reproduce as root, and not on Windows) and offers "씨앗을 통한 stat 주입
**또는 플랫폼 조건부 skip**". Taking the second branch makes AC-CGP-003 class 4
unfired, which `acceptance.md:34` records as **미검증**, which in turn makes
AC-CGP-004 — declared must-pass at `acceptance.md:111` — unachievable: a skipped
class cannot be the thing the mutant flips. The plan presents as an option
something that, if chosen, fails the SPEC.
Note the seam does not exist yet: `doctor_codex.go:426` calls `os.Stat` directly
with no injection point, so the pruner must introduce its own stat seam — which
is new code, not reuse, and the plan should say so.
Required fix: make stat-seam injection the required route in `plan.md §F`, and
state that a skipped indeterminate fixture is a **FAIL** of AC-CGP-004, not a
recorded gap.

**D6. `path = ""` is a distinct on-disk form no AC fixture covers** —
`spec.md:77`, `acceptance.md:31` — Severity: **minor** — Class: **optional**.
REQ-CGP-010 and AC-CGP-003 class 5 both say "`path` 키를 선언하지 않은 항목". The
detector distinguishes a third form: `skillPathKeyRe` (`skills.go:66`) matches
`path = ""` and yields `Path == ""`, which `doctor_codex.go:396` skips by the same
early-continue. The SPEC is *safe* here twice over (REQ-CGP-004 condition 1
requires a non-empty path, and `classifyCodexSkillPath("")` returns
`codexPathRelative`), so this is coverage, not a hole. Worth one extra fixture
row.

**D7. `moai clean`'s help text states a blast radius the new scope contradicts** —
`internal/cli/clean.go:34` — Severity: **minor** — Class: **optional**.
The existing Long text says "Only ~/.moai is touched — ~/.claude is never
modified", written when `--home` was the only home-scoped branch. Adding a scope
that mutates `~/.codex/config.toml` makes that sentence false for the command as
a whole. `plan.md:68` M5 mentions 도움말 generically; it should name this sentence.

**D8. `§D.3`'s live-config proof names no boundary** — `acceptance.md:138` —
Severity: **minor** — Class: **optional**.
"실행 전후 sha256" — before and after *what*? The run phase as a whole, each test
binary invocation, or each `--force` fixture? Pin it (recommended: once before
the first run-phase command and once after the last, both recorded in the verdict
record).

---

## What I checked and found nothing wrong with

Stated positively, per the instruction not to assert cleanliness without saying
what was examined.

**Never-prune completeness (attack #1) — the six classes are complete and each is
correctly derived.** I enumerated every outcome
`codexStaleSkillFinding` (`doctor_codex.go:376-451`) distinguishes for one entry
and mapped it onto the SPEC:

| Detector outcome | Source | SPEC |
|---|---|---|
| `e.Path == ""` → skipped | `:396` | REQ-CGP-010 (+ REQ-CGP-004.1) |
| relative → `relativeCount++`, never stat'ed | `:412-416` | REQ-CGP-006 |
| oddly-formed (`default:`) | `:419-424` | REQ-CGP-007 |
| home-relative, `ok == false` → `indeterminate++` | `:404-410` | REQ-CGP-008 |
| `serr == nil` (incl. a directory) | `:428-432` | REQ-CGP-011 |
| stat error not `ErrNotExist` → `indeterminate++` | `:442-450` | REQ-CGP-009 |
| `errors.Is(serr, fs.ErrNotExist)` → missing, split by tri-state `Enabled` | `:433-441` | REQ-CGP-004 eligible + REQ-CGP-005 (enabled is not a gate — matches the code, which counts `SkillEnabledTrue` into `missing` too) |
No outcome the detector distinguishes is absent from the SPEC's set, and no listed
class is wrong. REQ-CGP-011's insistence that a **directory** counts as resolved
is a verbatim match for the code's own comment at `:429-432`.

**Surface decision facts (attack #5) — all three rejections rest on facts I
re-read, and every one is accurate.**
- `codexCmd` really is `DisableFlagParsing: true` (`codex_launcher.go:327`), and
  the head comment really does say "The launcher never writes: no directory is
  created, no file mutated (REQ-CL-013)" (`codex_launcher.go:24-25`). The
  hand-parsing helpers named in `plan.md:34` exist.
- `doctor.go:58` carries `Bool("fix", false, "Suggest fixes for detected issues")`
  — the help text is quoted correctly to the character — and I searched the whole
  `--fix` path: `doctor.go:105-114` only appends a static string
  `"- %s: run 'moai init' to initialize project"` into an info card. There is no
  repair-execution path anywhere in `internal/cli/doctor*.go` (`grep -rn
  'suggestFixes'` → no matches). `plan.md:39`'s stronger-than-the-brief reading
  ("doctor 에 없던 실행 표면을 새로 만드는 일") is correct.
- `newCleanCmd` (`clean.go:19-53`) is a plain cobra command, `GroupID: "tools"`,
  with exactly the two bools `--force` and `--home`, and `--home` branches to
  `runCleanHome` — the precedent `plan.md:13` claims. `clean_home.go` does carry
  the single-predicate guard (`isCarvedOut`, `:63`) consulted in the producing
  scan (`:155,289,321`) and the `@MX:WARN` markers (`:61,366`).
The chosen surface is, in my judgment, the right one, and it is right for the
reasons given.

**Scope discipline (attack #6) — axis 2 does not leak.** `spec.md:106-108`
excludes the `codex_measured_version` restamp explicitly; no REQ mentions
`agents-codex.yaml`, the manifest, or a version stamp
(`grep 'codex_measured_version\|agents-codex' spec.md` → only the exclusion line).
`plan.md`'s milestone table touches four files, none of them the manifest.

**Non-goal integrity (attack #7) — stated, and no AC would execute against the
live config.** `spec.md:101-104` says the card does **not** prune this machine's
49 entries and names the deliverable as the verb. `plan.md:78` requires every
test to stay inside `t.TempDir()` and to reach fixtures through the
`codexUserHomeDir` / `resolveCodexHomeDir` seams. I read all 13 ACs looking for
one that requires a real-config run; none does — AC-CGP-009's "홈이 해석되지 않음"
and "설정 파일 없음" are seam-driven states. `acceptance.md:138` adds a positive
proof obligation (unchanged sha256), which strengthens rather than weakens this.

**Structural / mechanical checks that came back clean, with the command run.**
`grep -o 'AC-CGP-[0-9]*' acceptance.md | sort -u` → 13 ids, sequential, no
duplicates. `grep -rn 'NEEDS CLARIFICATION'` → rc 1. `grep -c syscall spec.md` →
`0`. `moai spec lint` from a tree-built binary → rc 0, `✓ No findings` (with the
vacuity caveat recorded under MP-2, and a mutant control proving the run was not
empty).

---

## Baseline-attribution

Every figure above was produced in this run, in this tree (`ace1c5440`,
`.claude/worktrees/t506`). The linter used is `/tmp/t506-audit-moai`, built in
this run from this tree (`go build -o /tmp/t506-audit-moai ./cmd/moai`, rc 0) —
not the installed `~/go/bin/moai`, whose lag is unknown. Source line numbers are
read from the working tree at that HEAD. The baseline population figures (49
entries, all absolute, all `enabled = false`, all missing; four never-classes at
0) are **cited, not re-measured** — their attribution is
`.moai/reports/t506/baseline-measurement.md`, and I did not re-run its commands.

## Gaps — what I did NOT observe

- I did not re-measure this machine's `~/.codex/config.toml`. The SPEC's §A.1
  table is taken from the baseline report on its own attribution.
- I ran no Go tests. No implementation exists yet, so there was nothing to run;
  `plan.md §E`'s verification scope is a run-phase obligation and is unverified
  here.
- I did not check `internal/cli`'s existing test files for a name collision with
  the fixtures the plan implies, nor whether an `internal/cli` stat seam already
  exists under another name.
- I did not evaluate whether `moai clean --codex-skills` collides with any
  in-flight card's flag work.
- The `splitLines` trailing-newline loss (D1) is derived by reading
  `configtoml.go:204-209`, not by executing a round trip. It is a reading of the
  code, and I state it as such.

## Residual risk

- **The boundary this SPEC exists to defend is unobservable on this machine.**
  Four never-classes are 0 here, so every safety guarantee rests on fixtures that
  do not exist yet. D4 and D5 are precisely the two ways those fixtures can end
  up proving nothing while the record reads GREEN.
- **The indeterminate class is platform-shaped even though `syscall` never
  appears.** A permission-denied stat does not reproduce as root and behaves
  differently on Windows; D8 auto-passed on a literal-substring test that cannot
  see this.
- **Even with D1-D5 fixed, byte-preservation over arbitrary hand-written TOML is
  the residual hazard of the whole design.** The parser was written to
  under-report for an advisory diagnostic (`skills.go:78-82`, "Silence is the
  safe direction"); under-reporting is safe for a counter and is *also* safe for
  a pruner (an unseen entry is simply not pruned) — but the span logic inherits
  no such bias, and a span that is one line wrong deletes a line the user wrote.
  `plan.md §F` row 1 names this; the mitigation it names (D2) is not yet
  sufficient.

---

## Recommendation

FAIL. Five blocking defects, in the order I would fix them:

1. **D1** — rewrite AC-CGP-013 as a unit round-trip over the reassembly function,
   with trailing-newline-present, trailing-newline-absent, and CRLF fixtures.
   Record in `plan.md §C` that `splitLines` alone loses the trailing-newline
   state, so the reassembler must carry it separately. This is the one defect
   that can invalidate the whole `§C` approach, which is why it is first.
2. **D2** — correct `plan.md:49`'s claim about the parser and specify the span
   over entry-internal content (multi-line literals, unknown keys, comments,
   EOF); add the two missing ACs.
3. **D5** — make stat-seam injection required and declare a skipped indeterminate
   fixture a FAIL of the must-pass AC.
4. **D4** — bind the mutant's RED to the named never-prune test and to the
   pruner's own eligibility branch.
5. **D3** — add the surface requirement backing AC-CGP-012 and map it in `§D.2`.
6. **MP-2** — reword REQ-CGP-004 and REQ-CGP-005 into modal GEARS form. Trivial;
   do it in the same pass.

D6, D7, D8 are optional-class: surface them to the orchestrator, do not gate on
them.

Nothing in this report argues against the card's design. The surface choice is
sound and its rejections are correctly grounded; the never-prune set is complete
against the detector. What fails is the evidence layer — the two criteria that
were supposed to prove the boundary and the byte-preservation hold can both be
satisfied while proving neither.
