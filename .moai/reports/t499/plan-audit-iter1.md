# SPEC Review Report: SPEC-CODEX-PARTIAL-WIRING-001

Card: t499 · Tier M (PASS threshold 0.80) · Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t499`, branch `WT-codex-partial-wiring`, HEAD `ace1c5440`

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.76** (arithmetic mean 0.7625; harmonic mean 0.746) — below the Tier M threshold of 0.80

> Reasoning context ignored per M1 Context Isolation. The dispatch's framing of the un-nagging tension was read as a scoping instruction only; every judgment below is made against the four SPEC artifacts and the source files, and every claim carries the command that produced it.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `spec.md:86-93` carries REQ-CPW-001 … REQ-CPW-008: sequential, no gap, no duplicate, uniform 3-digit zero-padding.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md` §C `REQ-XXX` rows) only. All 8 match a GEARS pattern: Ubiquitous (001 `spec.md:86`, 005 `:90`, 008 `:93`), Event-driven (002 `:87`, 003 `:88` — both `~일 때, 검사는 … SHALL`), Unwanted (004 `:89`, 007 `:92` — both canonical `SHALL NOT`, no deprecated `If/then`), State-driven (006 `:91` — `~인 동안`). The Given-When-Then entries in `acceptance.md` are the **verification layer** and are correctly formatted as such; they are graded under Group 4, not here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types (`spec.md:2-14`); canonical names throughout, no rejected snake_case alias (`created:`/`updated:`/`tags:`/`id:` all correct). `phase: "v3.1.5 target"` is a release-target label, matching the schema's own `"vX.Y.Z target"` example form and violating no prohibition (`plan`/`run`/`sync`/`mx` are the prohibited whole-values). Optional `tier: M` valid. See F8 for a non-blocking note on `related_specs`.
- **[N/A] MP-4 Section 22 language neutrality** — single-language scope. `module: internal/cli`; the SPEC governs Go source in this repository and touches no `internal/template/templates/**` surface. The only occurrence of `internal/template` is `spec.md:137`, naming `internal/template/agentemit` as **out of scope**. Criterion auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — 3 external SPEC-IDs referenced; all exist and all read `status: completed`:
  ```
  SPEC-CODEX-SIDECAR-GUARD-001   completed
  SPEC-CODEX-WIRING-001          completed
  SPEC-INIT-HARNESS-PROMPT-001   completed
  ```
  None is retired/superseded/archived, so no reconciliation clause is owed. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall'` over all four artifacts returns `0` on every file. D8 auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-PARTIAL-WIRING-001/` → no match (exit 1). The three unverified premises are recorded as prose Gaps in `plan.md:100-104`, which is the correct disposition, not a marker.

**Mechanical corroboration.** `moai spec lint` over the whole `.moai/specs/` tree (exit 0, background run) emits **zero findings naming this SPEC** — `grep -c 'SPEC-CODEX-PARTIAL-WIRING-001' <lint-output>` → `0`, against a run that reported `ModalityMalformed` and `CoverageIncomplete` findings on several sibling SPECs. That is an independent check of MP-1, MP-2, MP-3, the `OutOfScopeRule`, and REQ→AC coverage.

**All seven must-pass criteria clear. The FAIL is carried entirely by the aggregate score**, and the score is carried by the Testability dimension.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.80 | 0.75 | Measured-baseline §A, named design decisions §D, explicit out-of-scope §E. Two under-specified requirements: `spec.md:88` (REQ-CPW-003 constrains only the *status*, never the message's content — see D3) and `spec.md:89` (REQ-CPW-004 says "어떤 분기에서도" but only one branch is measured — see D4). |
| Completeness | 0.90 | 1.0-adjacent | HISTORY `spec.md:22`, WHY `§B:73`, WHAT `§A:28`/`§C:80`, HOW `plan.md`, REQUIREMENTS `§C:80`, ACCEPTANCE `§F:139` + `acceptance.md`, Out of Scope — three `### Out of Scope — <topic>` H3 sub-headings at `spec.md:121,127,133`, each with specific `-` bullets (matches `OutOfScopeRule`; corroborated by the zero-finding lint run). Deduction: the existence-gate preservation has an AC but no requirement (D5). |
| Testability | 0.60 | 0.50 | Measured, not inferred. AC-CPW-007 is vacuous (D1); AC-CPW-001/002/005/006 carry no RED-now observation and their selectors are green-on-empty-sweep (D2); no AC cell carries the four §2.1 elements. No weasel words found (`grep -iE '적절\|합리적\|충분히\|reasonable\|appropriate\|adequate' acceptance.md` → no match). AC-CPW-009 is the only vacuity brake and does not cover the two axes D1/D3 name. |
| Traceability | 0.75 | 0.75 | Both directions complete — every REQ-CPW-001…008 has ≥1 AC, and every AC-CPW-001…009 names a REQ (`acceptance.md:10-18`). One mapping is indirect to the point of being wrong: AC-CPW-008 → REQ-CPW-007 (D5). |

---

## Defects Found

**D1. AC-CPW-007 is vacuous — REQ-CPW-008 has no live acceptance criterion** — `acceptance.md:89-94` (and the matrix row `:16`) — Severity: **major** — Class: **blocking**

AC-CPW-007's Then-clause asserts "새 문구의 rune 길이가 `codexMessageWidthCeiling`(113) 이하다", but its 판정 command is only the two pre-existing width tests. Both of those tests call `checkCodexWiring(t.TempDir(), …)` — an **empty** temp dir (`internal/cli/doctor_codex_test.go:394` and `:441`), i.e. the `unwired` path, which REQ-CPW-006 freezes byte-for-byte. **Neither test can ever reach a half-wired project, so neither can ever observe the new message's width.**

Measured on the pre-implementation tree:
```
$ go test ./internal/cli/ -run 'TestCheckCodexWiring_(MessageWidthStaysInBand|RenderedPanelStaysInBand)' -v
--- PASS: TestCheckCodexWiring_MessageWidthStaysInBand (0.00s)
--- PASS: TestCheckCodexWiring_RenderedPanelStaysInBand (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.784s
exit=0
```
Green before the work, green after, and green under a mutant that gives the half-wired branch a 200-rune Message. This is `verification-completeness.md` §1.1 report-not-verdict and §2's mutant probe: **a mutant satisfying AC-CPW-007 while violating REQ-CPW-008 is trivially writable.**

**Required fix:** add a new test to the AC-CPW-007 판정 that constructs a half-wired fixture in **both** PATH branches and asserts `utf8.RuneCountInString(check.Message) <= codexMessageWidthCeiling` on each — e.g. `TestCheckCodexWiring_HalfWiredMessageWidthStaysInBand`. Keep the two existing tests in the AC as regression guards, but they cannot be the whole judgment.

---

**D2. Four ACs carry no RED-now cell, and their judgment commands are provably green-on-empty-sweep** — `acceptance.md:27, 34, 80, 87` (AC-CPW-001, -002, -005, -006) — Severity: **major** — Class: **blocking**

Each of these four names a test that does not yet exist and records its expected output as `--- PASS`. Per `verification-completeness.md` §2, adoption takes **two cells**: a RED-now observation pinned to the pre-implementation tree, plus a green path. These four carry only the green path.

The failure is not theoretical. Measured on this tree:
```
$ go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCodexInstalled
ok  	github.com/modu-ai/moai-adk/internal/cli	0.854s [no tests to run]
exit=0
```
A selector matching zero tests **exits 0 and prints `ok`**. So today, AC-CPW-001's own judgment command reports success. That is precisely the evidence row §2.1 was written from ("nine release-blocking criteria … resting on a single premise — that an absent test turns its suite red — where a runner given a selector matching zero tests exits 0 and prints `ok`").

Related, same root: no AC cell in `acceptance.md` carries the four §2.1 elements (command / that command's verbatim stdout / its exit code / a tree SHA). `plan.md:3` and `progress.md:7` pin `ace1c5440` at document level, but `acceptance.md` carries no pin of its own, so §2.1's document-level-pin fallback does not reach these criteria.

**Required fix:** for each of AC-CPW-001/-002/-005/-006, add a RED-now cell that (a) states the tree SHA `ace1c5440`, (b) states the pre-implementation observation **with its `[no tests to run]` token quoted verbatim** and exit code `0`, and (c) states explicitly that *the absent test is why it is not red* — so the reader is not invited to mistake the `ok` for a passing baseline. Then state the green path: which milestone (`plan.md` §E M2) creates the test and what the passing line becomes. Pin `ace1c5440` at the top of `acceptance.md` so future criteria inherit it.

---

**D3. REQ-CPW-003 admits a mutant that nags every claude-only project while satisfying all nine ACs** — `spec.md:88` (REQ-CPW-003), `acceptance.md:29-34` (AC-CPW-002) — Severity: **major** — Class: **blocking**

This is the sharpest form of the un-nagging tension, and it is a defect rather than a disagreement.

REQ-CPW-003 requires only that the codex-absent branch "그 상태를 이름 붙여 보고" at `CheckOK`. AC-CPW-002 asserts exactly three things: `CheckOK`, the Message states the agent definitions exist, and `claude-only` appears zero times. **Nothing in the requirement or the criterion forbids a directive.** An implementation emitting

> `Codex agent definitions present, wiring absent — run moai init --agent codex`

satisfies REQ-CPW-003 and AC-CPW-002 in full, and — because §A.1 measured that plain `init` deploys the TOMLs into **every** project — it puts an unactionable instruction (codex is not installed) on essentially every project the tool creates. That is the outcome §D-1 (`spec.md:97-103`) says the split exists to prevent. The SPEC's own design intent is unenforced by its own criteria.

**Required fix:** add a constraint to REQ-CPW-003 — the codex-absent Message SHALL NOT carry an action directive — and add the matching assertion to AC-CPW-002: `check.Message` does not contain `initCodexAdvice`. This costs one line each and closes the mutant.

---

**D4. REQ-CPW-004's "어떤 분기에서도" is measured in one branch only** — `spec.md:89`, `acceptance.md:33` — Severity: **minor** — Class: **blocking**

REQ-CPW-004 forbids describing a half-wired project as `claude-only` in **any** branch. Only AC-CPW-002 (the codex-absent branch) asserts the zero-occurrence. AC-CPW-001 (`acceptance.md:22-27`) makes no such assertion, so a mutant placing `claude-only` in the Warn branch's Message satisfies every criterion while violating REQ-CPW-004.

**Required fix:** extend AC-CPW-001's Then-clause with the same `claude-only` zero-occurrence assertion over `Message + Detail`, or restate it once as a branch-parametrised criterion.

---

**D5. AC-CPW-008 traces to a requirement that does not state its subject** — `acceptance.md:17` (matrix), `acceptance.md:52-57` — Severity: **minor** — Class: **blocking**

AC-CPW-008 asserts the `internal/codexwiring` **existence gate** is untouched, and maps to "REQ-CPW-007(존재-게이트)". But REQ-CPW-007 (`spec.md:92`) is scoped to the doctor check: "검사는 `.codex/` 파일이나 `~/.codex/config.toml`을 생성·수정·삭제해서는 안 되며 … 종료 코드를 바꾸어서는 안 된다". Its subject is `checkCodexWiring`; it says nothing about `RefreshWiring` / `wireProject` / `wiringFilesExist`.

The existence-gate contract is stated only in `spec.md:62` (a baseline table row) and fenced at `spec.md:124` (out of scope). Neither is a requirement, so AC-CPW-008 is an acceptance criterion with no requirement behind it. Note this is a **mapping** defect, not a fencing defect — the out-of-scope section does fence the gate correctly and explicitly ("존재-게이트는 **그대로 보존**되며 이 SPEC은 그 계약을 건드리지 않는다").

**Required fix:** either add a REQ-CPW-009 (Unwanted: "이 SPEC의 구현은 `RefreshWiring`/`wireProject`/`wiringFilesExist`의 동작을 변경해서는 안 된다") and re-map AC-CPW-008 to it, or re-map AC-CPW-008 to the §E out-of-scope clause explicitly and stop citing REQ-CPW-007.

---

**D6. §A.3 cites the wrong line for the discriminant** — `spec.md:49` — Severity: **minor** — Class: **blocking**

`§A.3` cites `internal/cli/doctor_codex.go:105` and asserts, in its own heading, that this is "읽은 것이지 추론이 아님". The quoted **snippet is correct**; the **coordinate is not**. Measured on this tree:
```
$ git rev-parse HEAD
ace1c5440fad4d2a3787b0331059d4f3c4149d15
$ grep -n 'wired := hooksErr' internal/cli/doctor_codex.go
92:	wired := hooksErr == nil || cfgErr == nil
```
Line 105 is `var problems []codexFinding`. `git status --short` shows `doctor_codex.go` unmodified in this worktree, so the file the SPEC was written against is the file I read. The same wrong number appears in `.moai/reports/t499/repro.md:67` and propagated from there into the dispatch — worth correcting at the source so it stops travelling.

For contrast, the sibling citation is correct: `spec.md:62` cites `internal/codexwiring/wire.go:51-56`, and `RefreshWiring` spans exactly lines 51-56 (`51:func`, `52:if !wiringFilesExist`, `53:return Result{}, nil`, `55:return wireProject`).

**Required fix:** change `spec.md:49` to `internal/cli/doctor_codex.go:92`, and correct `repro.md:67` alongside it.

---

**D7. §F's golden-preservation risk row rests on an unstated population claim** — `plan.md:82`, `spec.md:91` (REQ-CPW-006) — Severity: **minor** — Class: **optional**

`plan.md:82` mitigates the golden-breakage risk with "진짜 claude-only 프로젝트의 침묵 보존", and REQ-CPW-006 is framed as preserving the un-nagging contract. The **code claim is correct** (see the §A.5 verification below), but the **framing is not**: given §A.1, the population REQ-CPW-006 protects is projects with no `.codex/` at all — pre-existing projects, hand-made ones, and the golden fixture. A claude-only user who runs `moai init` today gets a half-wired project and lands on the **new** message, not the preserved one. So REQ-CPW-006 is a *regression-radius* argument (which §D-2 at `spec.md:105-112` states correctly and well), not an *un-nagging* argument.

**Required fix (optional):** one sentence in §D-1 or the §F row stating that the preserved `unwired` wording now covers a near-empty live population, so no reader mistakes REQ-CPW-006 for the thing that protects claude-only users. What actually protects them is the `CheckOK` status of REQ-CPW-003 — plus, once D3 is fixed, the no-directive constraint.

---

**F8 (not a defect, recorded for the record).** `related_specs:` (`spec.md:15`) is not in the schema's Optional Fields list; the documented field for this relation is `depends_on`. It causes no lint finding (confirmed by the zero-finding lint run) and I do not recommend changing it inside this card.

---

## The three targeted questions

### 1. The un-nagging tension

**(a) Is REQ-CPW-006's preservation claim weaker than it reads? — Yes, and it is currently mis-framed.** Verified as a defect above (D7, minor/optional). The mechanism is exactly as the dispatch suspected: with plain `init` depositing the TOMLs into every project (`repro.md:15-38`, 11 files, no wiring), the `unwired` population is effectively legacy-and-fixture-only. REQ-CPW-006 remains **correct and worth keeping** — it is what holds the golden fixtures and the three named tests still — but it preserves *regression radius*, not *user experience*, and the SPEC currently advertises it as the latter.

**(b) Does a new informational Message on effectively every project violate the un-nagging contract? — Not as the SPEC intends it, but as written it permits a violation.**

The letter is satisfied cleanly, and more so than the framing suggests. The un-nagging invariant's mechanism (`doctor_codex.go:96-103`) does not suppress a *row* — the check always renders one; on a claude-only machine it renders `not wired (claude-only project) — skipped`, which the golden fixtures carry at `doctor-{light,dark,nocolor}.golden:43`. So the change adds **no new row and no new severity**: an OK row's text changes, the `ok` count is unchanged, the `warn` count is unchanged.

More importantly, the code comment states the invariant's *premise*: "Claude-only project on a claude-only machine: **nothing about Codex is in play**, so the check stays silent." §A.1 measured that premise false for every `moai init` project — Codex agent definitions **are** in play. Replacing a false assertion with a true one is not nagging; nagging is repeatedly urging an action the user has not chosen, and the codex-absent branch chooses `CheckOK` with no directive.

The spirit-violation the dispatch suspected is nonetheless **real and reachable — through the requirement, not through the design**: nothing in REQ-CPW-003 or AC-CPW-002 stops the implementer from putting `run moai init --agent codex` into that Message, at which point every claude-only project carries an unactionable instruction forever. That is D3, and it is why I graded it blocking rather than treating it as a matter of taste.

**(c) Is `CheckOK`-with-changed-text the right disposition? — Yes. I agree with §D-1, and I want to be explicit that this is agreement, not a finding.**

Escalating to `CheckWarn` would put a warning on every project the tool creates, which is the outcome §A.4's invariant exists to prevent. Staying silent would leave a measured false assertion in place. `CheckOK` + true text is the only option that does neither. *Judgement call, offered as an alternative rather than a defect:* a strictly smaller change would be to drop only the false clause — `not wired (claude-only project) — skipped` → `not wired — skipped` — satisfying REQ-CPW-004 without introducing any Codex-mentioning text on a codex-less machine. It is less informative and I do not recommend it over §D-1's choice; I record it because it is the minimum diff that clears the falsehood, and the SPEC never considers it.

### 2. AC-CPW-009's mutant probe

Judging each mutant against the criteria **as written**:

| Mutant | Goes RED? | Reasoning |
|---|---|---|
| M1 — revert the discriminant to `wired := hooksErr == nil \|\| cfgErr == nil` | **Yes**, both named tests | codex-absent: the half-wired fixture re-enters the `!wired && !codexInstalled` early return (`doctor_codex.go:96`) and the Message becomes `not wired (claude-only project) — skipped`, failing AC-CPW-002's zero-`claude-only` assertion. codex-installed: the Message reverts to the `unwired` wording, failing AC-CPW-001's "states the agent definitions exist" assertion. |
| M2 — directory-existence instead of file-existence | **Yes**, conditionally | Only if `TestCheckCodexWiring_HalfWiredCountIndependent` actually builds the empty-`.codex/agents/moai/` case AC-CPW-005 (`acceptance.md:77-79`) specifies. The criterion does specify it, so the mutant is caught — but the RED depends on a fixture the criterion describes and that does not yet exist (D2). |
| M3 — restore the old wording in the codex-absent branch | **Yes** | Directly contradicts AC-CPW-002's zero-`claude-only` assertion. |

All three would go RED. **The probe is sound as far as it goes, and it is the strongest item in the acceptance set** — it is what keeps M1/M3 from being uninterpreted greens.

**Does a mutant exist that satisfies all nine criteria while violating the requirements? Yes — two, and I would not adopt AC-CPW-009 as a sufficient brake until both are closed:**

- **Width mutant** (violates REQ-CPW-008): give the half-wired Message 200 runes. AC-CPW-007's named tests never touch that path (D1, measured), and no other AC measures width. All nine pass.
- **Directive mutant** (violates §D-1's stated design intent, and the un-nagging invariant in substance): put `run moai init --agent codex` into the codex-absent Message. AC-CPW-002 asserts only status, agent-definition mention, and no `claude-only`. All nine pass (D3).

**Two-cell adoption (`verification-completeness.md` §2), applied to all nine:**

| AC | RED-now cell? | Classification |
|---|---|---|
| AC-CPW-001, -002, -005, -006 | **Absent** — selector matches zero tests, exits 0, prints `ok … [no tests to run]` (measured) | Not adopted; currently vacuous-green. Fix per D2. |
| AC-CPW-003, -004, -008 | **Absent by construction** — preservation criteria, green before and green after | Per §2.1's undecidable disposition these are **regression-guards**, not release-blocking. The SPEC classifies them as "회귀 방지"/"금지 제약" in the matrix, which is the right instinct, but never says they are therefore not release gates. Recommend labelling them explicitly. |
| AC-CPW-007 | **Absent, and cannot exist** — green on a path frozen by REQ-CPW-006 | Vacuous. Fix per D1. |
| AC-CPW-009 | The mutant observation **is** its RED cell | The only properly two-celled item in the set. |

### 3. The golden-fixture safety claim (§A.5)

**Verified against the harness, and the claim holds — read, not accepted.**

`internal/cli/doctor_golden_test.go`:
- `:129` — `t.Chdir(t.TempDir())`, an empty temp dir with no `.codex/`.
- `:119-126` — `codexWiringLookPath` is replaced with a stub returning `os.ErrNotExist` for `"codex"` (and delegating for every other name), restored via `t.Cleanup`.

Both conditions §A.5 asserts are present and exactly as described. The golden project is therefore `!wired && !codexInstalled` and takes the early return at `doctor_codex.go:96-103`. It has no `.codex/agents/` directory, so it can never be classified half-wired under any correct implementation of REQ-CPW-001.

Confirming the encoded baseline:
```
$ grep -n 'Codex Wiring' internal/cli/testdata/doctor-*.golden
doctor-light.golden:43:    ok  Codex Wiring  not wired (claude-only project) — skipped
doctor-dark.golden:43:     ok  Codex Wiring  not wired (claude-only project) — skipped
doctor-nocolor.golden:43:  ok  Codex Wiring  not wired (claude-only project) — skipped
```
All three carry the `unwired` row that REQ-CPW-006 freezes. **§A.5 is accurate and AC-CPW-004 is a sound regression guard.** The harness comment at `:113-118` even states the intent verbatim ("Absent is the snapshot's encoded baseline (the claude-only skip)"), which is a stronger corroboration than the SPEC claims for itself.

---

## Other checks at normal depth

- **Traceability, both directions:** complete apart from D5. No orphan REQ (all of 001-008 appear in the `acceptance.md:10-18` matrix; the lint run's `CoverageIncomplete` rule reported nothing against this SPEC); no AC without a REQ (AC-CPW-009's "전체" is an acceptable set-wide mapping for a vacuity brake).
- **Out-of-scope fencing:** three H3 sub-headings with specific bullets, satisfying `OutOfScopeRule`. The existence-gate fence (`spec.md:124`) is explicit and correct. The `init`-side fence (`:133-137`) correctly declines to change plain-init behaviour — which matters, because that behaviour is the whole premise of §A.1 and re-opening it would invalidate the design. The wizard path is fenced as unmeasured rather than as decided, which is the honest disposition.
- **REQ/AC budget:** 8 requirements / 9 criteria, both well under the Tier M ceiling of 16/16.
- **Implementation leakage in requirements (RQ-3/RQ-4):** REQ-CPW-008 (`spec.md:93`) names the code identifier `codexMessageWidthCeiling` and its value. This is mild HOW leakage in a WHAT layer, but the constant is the contract's only stable name and the SPEC would be less precise without it. Recorded, not scored as a defect.
- **Unfalsifiable / off-target criteria:** none beyond D1. AC-CPW-006's extra `moai doctor; echo rc=$?` → `rc=0` check is weak but not vacuous — `runDoctor` returns `doctorExitStatus(failCount)` (`internal/cli/doctor.go:123`), counting only `CheckFail`, so the assertion would genuinely catch a `CheckFail` regression while being insensitive to `CheckWarn`. That is the correct sensitivity for REQ-CPW-007's exit-code clause.
- **Era-classification note (informational, not a defect):** `matchesModernPhase` (`internal/spec/era.go:288-300`) matches only `v3r6` or a `v3.0` prefix, so `phase: "v3.1.5 target"` does not satisfy it; classification falls to `created: 2026-09-07 >= 2026-04-01` and still resolves V3R6 on one signal. This is a limitation of the predicate against any `v3.1.x` label, not a defect in this SPEC — no correct release label for this SPEC would satisfy it. The author's flagged uncertainty about the release number (`plan.md:102`) is therefore **immaterial to schema validity and to era classification**; it matters only if someone later reads the field as a scheduling commitment.

---

## Recommendation

The SPEC is materially above the average artifact I audit: the measured-baseline discipline in §A is real (I re-derived §A.5 and the D7 statuses independently and both held), §D-2's condition-narrowing over wording-rewrite is the right call and well argued, and §D-3's refusal to encode the count 11 is correctly reasoned from the residual risk `repro.md:87` recorded. **The FAIL is a near-miss at 0.76 against 0.80, and it is concentrated entirely in the verification layer** — the requirement layer clears every must-pass criterion and draws zero mechanical lint findings.

Six blocking findings, in the order I would fix them:

1. **D1** — add a half-wired width test to AC-CPW-007's judgment. REQ-CPW-008 currently has no criterion that can fail.
2. **D3** — add "the codex-absent Message SHALL NOT carry an action directive" to REQ-CPW-003, and the matching `initCodexAdvice`-absence assertion to AC-CPW-002. This is the clause that makes the un-nagging design enforceable rather than merely intended.
3. **D2** — give AC-CPW-001/-002/-005/-006 RED-now cells pinned to `ace1c5440`, quoting the `[no tests to run]` token verbatim so no reader mistakes the current `ok` for a baseline. Pin the SHA at the top of `acceptance.md`.
4. **D4** — extend AC-CPW-001 with the `claude-only` zero-occurrence assertion.
5. **D5** — add REQ-CPW-009 for the existence gate, or re-map AC-CPW-008 to §E.
6. **D6** — correct `doctor_codex.go:105` → `:92` in `spec.md:49`, and in `repro.md:67`.

One optional finding (**D7**) and one judgement-call alternative (the minimum-diff `not wired — skipped` wording) are recorded above; both are the lead's to route, and neither should be treated as required work.

Also recommended, cheap: label AC-CPW-003/-004/-008 explicitly as **regression-guards** rather than release-blocking criteria, per §2.1's undecidable disposition. They are correct and worth keeping; they simply cannot be gates.

Every fix above is an edit to `spec.md` / `acceptance.md`. None requires re-opening §B's design decisions, none changes the milestone plan in `plan.md` §E, and none affects the §A measured baseline other than D6's coordinate. A revised iteration 2 would be a scoped re-audit over this enumerated defect delta.
