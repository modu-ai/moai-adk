# SPEC Review Report: SPEC-ASTGREP-LANG16-001

Iteration: 2/3
Verdict: **FAIL**
Overall Score: **0.69** (iteration 1: 0.68 — delta **+0.01**). Tier L PASS threshold: 0.85.
STOP signal: **not raised** (no score regression; iteration 3 remains available).

Auditor: plan-auditor. Bias-prevention M1-M6 active.

**M1 Context Isolation statement.** Reasoning context ignored per M1 Context Isolation. The author's
rationale as relayed in the invocation prompt was treated as a set of *claims to falsify*, not as
evidence. Every judgement below rests on the six SPEC A artifacts, the three SPEC B artifacts (read
only for the two permitted split questions), and my own measurements against the tree.

## Measurement attribution

| field | value |
|---|---|
| worktree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t228` |
| branch | `WT-astgrep-16-langs` |
| HEAD | `294b4b6ab` (`git rev-parse --short HEAD`) |
| tool | `ast-grep 0.40.5` (`sg --version`) |
| probe scratch | `/tmp/t228i2/` — outside the repo; one in-tree probe edit to `security/web.yml` was made and reverted with `git checkout --`, leaving `git status --porcelain internal/template/` empty |

Every command below was run with cwd = worktree root, per the R7/§A.0 hygiene rule the SPEC itself
records.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency.** `grep -on '^\*\*REQ-A16-[0-9a-z]*\*\*' spec.md` returns
  exactly 23 definitions, `001`…`023`, contiguous, zero-padded to three digits, no duplicate, no
  sub-number. `grep -rhoE 'REQ-A16-[0-9a-z]+'` across both SPEC directories returns the same 23 plus
  `REQ-A16-022a` — which appears **only** in the `spec.md:26` HISTORY row recording its removal, i.e.
  an audit trail, not a live reference. Iteration-1 defects D4 and D10 are closed.
- **[PASS] MP-2 GEARS format compliance — judged against the `REQ-XXX` requirement layer in
  `spec.md` §C only.** (The `AC-A16-*` entries are Given-When-Then by design and are graded under
  Group 4, not here.) All 23 entries carry a named `<subject>`, a `shall`, and a modality matching
  their label. The iteration-1 failure is repaired verbatim at `spec.md:329-330`: *"**Where** a rule
  is promoted to severity `error`, the promotion **shall** satisfy **both** of:"*, with the
  `sec-csrf-no-token-check` counterexample preserved (`spec.md:341-353`). Two entries fuse a
  ubiquitous and a conditional clause in one paragraph (`REQ-A16-001`, `REQ-A16-018`) but both halves
  are well-formed; not a failure.
- **[PASS] MP-3 YAML frontmatter validity.** `spec.md:1-15` carries all 12 canonical fields with
  correct types: `id` / `title` (quoted) / `version` `"0.4.0"` (quoted semver) / `status: draft` /
  `created: 2026-08-24` / `updated: 2026-08-25` (ISO) / `author` / `priority: P2` / `phase` /
  `module` / `lifecycle: spec-anchored` / `tags` (comma-separated string), plus optional `tier: L`.
  No rejected snake_case alias. `related_specs:` (iteration-1 D17) is gone. **Correction to the
  author's account**: it was *removed*, not "replaced by `depends_on`" — SPEC A carries no
  `depends_on`. The outcome is compliant either way; the claim as stated is inaccurate.
- **[PASS] MP-4 language neutrality.** All 16 enumerated with equal weight (`spec.md:35-50`,
  `REQ-A16-001`); the two exclusions are version-scoped tool facts re-confirmed under
  `ast-grep 0.40.5`. `grep -rniE 'SPEC-[A-Z]|20[0-9]{2}-[0-9]{2}-[0-9]{2}'` over
  `internal/template/templates/.moai/config/astgrep-rules/` → rc=1, no match. The SPEC creates no
  neutrality violation. (Its *criterion* AC-A16-019 is defective — see E3 — but that is a
  verification defect, not a neutrality defect.)
- **[PASS] MP-5 D7 cross-SPEC reconciliation — no BLOCKING finding.** Extraction returns
  `SPEC-ASTG-UPGRADE-001`, `SPEC-ASTGREP-BREADTH-001`, `SPEC-ASTGREP-DOGFOOD-CLEANUP-001`,
  `SPEC-ASTGREP-MULTILANG-001`. Measured statuses: `archived`, `draft`, `completed`, `completed`. The
  one archived reference is now reconciled at **every** site: `spec.md:470-473` (§D, *"is **archived**
  and owns nothing active; this SPEC absorbs no part of it and pins `0.40.5` independently"*),
  `spec.md:496` (R6, ownership transferred here), `spec.md:516-517` (§H), `research.md:106-108`,
  `plan.md:141`, `plan.md:259`. R6 now has a live owner. Iteration-1 D3 is closed.
- **[PASS] MP-6 D8 cross-platform discipline.** `grep -rn 'syscall' .moai/specs/SPEC-ASTGREP-LANG16-001/`
  → rc=1. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate.** `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-ASTGREP-LANG16-001/`
  → rc=1, across all six artifacts.

**All seven must-pass criteria now pass** (iteration 1 had three failures: MP-1, MP-2, MP-5). The
FAIL verdict is therefore driven entirely by the aggregate score and the five blocking defects below,
not by the firewall.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.72 | between 0.50 and 0.75, nearer 0.75 | Gains are real: `REQ-A16-012` is now a requirement in form; the checker's cell-identity semantics are specified as a set comparison (`spec.md:278-279`, `design.md:196-199`); `AC-A16-020`'s sanctioned exemption is named. Against that: `design.md:119` and `design.md:154` still place `testConfigs`/`rule-tests` **inside** the template tree, contradicting `spec.md:395-396` and `design.md:163-169` — the load-bearing D13 decision is stated both ways in one artifact (E4). `REQ-A16-018` asserts a `catalog.yaml` coupling that does not exist (E2, measured). Four stale cross-references survive the renumbering (E6). |
| Completeness | 0.78 | 0.75+ | All required sections present. §D carries seven `### Out of Scope — <topic>` H3 sub-headings, each with specific `-` bullets and a named successor where one applies (`spec.md:420-473`). Frontmatter complete. Tier L 5-artifact set present and all five substantive. Headroom is real: 23/25 REQ and 23/25 AC. Deductions: `REQ-A16-010` has no acceptance criterion (E5); `AC-A16-018` has no requirement (E8); `AC-A16-016` and `AC-A16-017` are owned by no milestone (E7). |
| Testability | 0.60 | between 0.50 and 0.75 | Most criteria are binary and command-shaped, and the checker criteria (AC-006…010) are markedly sharper than iteration 1 — AC-A16-010's case (b) explicitly targets the substitution defect a count-based checker misses. Against that, five criteria are miscalibrated and I measured each: **AC-A16-001 passes on the untouched tree with zero implementation** (`sg test` on the shipped `sgconfig.yml` → `Running 0 tests … ok`, EXIT=0 — E1); **AC-A16-021's first clause is unsatisfiable by correct work** (a ruleset edit produces no `catalog.yaml` diff — E2); **AC-A16-019 fails today on 22 SPEC-ID files and 90 date-bearing files this SPEC does not touch** (E3); `AC-A16-013`'s second clause ("instantiates that class in idiomatic code") is a human judgement I defeated with a counterexample (D2-residual); `AC-A16-015` rests on the same judgement. A further four (AC-016/017/018/022) plus AC-020 and AC-023 hold today with zero implementation — legitimate as guards, but they mean 6 of 23 criteria observe nothing new. |
| Traceability | 0.70 | between 0.50 and 0.75, nearer 0.75 | Contiguous numbering is a genuine gain, and 22 of 23 requirements map to at least one criterion. Against that: `REQ-A16-010` ("A rule shall not ship without both test cases") maps to none, and I measured that `sg test` silently ignores a rule with no cases, so nothing else covers it (E5); `AC-A16-018` maps to no requirement — the former corpus-baseline requirement was deleted in the renumbering and only §D prose survives (E8); `plan.md` §B assigns AC-001…015 and AC-019…023 to milestones and leaves AC-016/017 unowned (E7); `design.md` carries four cites made stale by the renumbering — `AC-A16-007` for the unevidenced-exemption class (correct: `008`), `REQ-A16-014` for the matrix annotation (correct: `015`), `REQ-A16-021` for the `ruleDirs` obligation (correct: `023`), and `spec.md` §A.4 for the deny wiring (correct: §A.5) (E6). |

Aggregate = harmonic mean of (0.72, 0.78, 0.60, 0.70) = **0.69**. Tier L threshold 0.85.

---

## Answers to the six audit questions

### 1. D1 — really fixed, or narrated? **Fixed as a factual correction; the doctrine it produced was not applied to this SPEC's own harness criterion.**

The correction is real and I re-verified it against the code rather than against the prose. At
`internal/hook/pre_tool_scan_differential_test.go` the sequence is: the string builder closes with
the `CombinedOutput` cause message, then `t.Skip(b.String())` at line 242, then the assertion loop
`for _, fx := range scanCorpus {` at line 245 with `t.Errorf` on `gotDeny != fx.wantDeny` below it.
Every row of §A.7's enforcement table matches that code, including the row that reads *"Does not
exist. It is a skip."* The four passages that previously asserted a failure now assert a skip
(`spec.md:135-174`, `spec.md:221`, `design.md:236-248`, `plan.md:275-276`). This is not rewording —
it reverses the claim and cites the mechanism.

But the forcing-function half does not land where the author says it does. `REQ-A16-017` binds *"any
acceptance criterion [that] cites the differential corpus as evidence"*. SPEC A cites the corpus in
exactly one criterion — `AC-A16-018` — and that is a `git diff` emptiness check, not a corpus *run*.
So **`REQ-A16-017` has zero application inside SPEC A**; its only consumer is SPEC B, which carries
no requirements yet. `AC-A16-017` verifies only that the requirement *is stated in an applicable
form* — a documentation read that passes on the tree as it stands today.

That is defensible on its own (the corpus work moved to SPEC B with the hazard), and I do not fault
it. What I do fault is that **the identical vacuous-green failure reappears in SPEC A's own central
criterion, unguarded**. Measured:

```
$ sg test --config internal/template/templates/.moai/config/astgrep-rules/sgconfig.yml
Running 0 tests
test result: ok. 0 passed; 0 failed;
EXIT=0
```

`AC-A16-001`'s Then clause is *"it exits 0 and reports zero failures and zero missing snapshots"* —
satisfied, today, with nothing built. Its title promises "green over the existing 26 rules"; its
assertion never requires that 26 tests, or any tests, ran. This is D1's exact shape — a green run
that observes nothing — inside the SPEC that wrote the doctrine against it. **PARTIAL.**

### 2. D2 — anchor strength. **The anchor adds a field a reviewer can check; it does not raise a mechanical floor. The author's framing is accurate in its disclaimer and overstated in one clause.**

I tried to defeat it and succeeded on the first attempt. A rule carrying a **real** CWE id, a
plausible message, an `invalid` case that reads as ordinary Go, and a benign same-shape `valid` case
— whose pattern names an API that exists nowhere:

```yaml
id: probe-cwe-anchored
language: go
severity: error
message: Hardcoded credential detected.
metadata:
  cwe: "CWE-798"
  owasp: "A07:2021"
rule:
  pattern: acmeInternalVaultFetch($X)
# valid:   func ok()  { acmeSafeFetch("x") }
# invalid: func bad() { acmeInternalVaultFetch("x") }
```
```
$ sg test --config /tmp/t228i2/proj/sgconfig.yml
PASS probe-cwe-anchored  ..
test result: ok. 1 passed; 0 failed;    EXIT=0
```

This rule satisfies `REQ-A16-011` textually (it carries `metadata.cwe` naming a weakness class, and
its `invalid` case instantiates hardcoded-credential retrieval in syntactically idiomatic Go),
satisfies both clauses of `REQ-A16-012`, qualifies for `severity: error`, and records as IMPLEMENTED.
`AC-A16-013` cannot reject it without a human deciding that `acmeInternalVaultFetch` is not a real
API — the same unverifiable judgement `REQ-A16-008` already carried.

**Is the framing accurate?** Mostly. `spec.md:322-325` states the anchor is *"necessary, not
sufficient"* and that *"a determined author could still write a CWE-labelled rule matching nothing
real"* — that is honest and it is exactly what I measured. R2 is marked `[MITIGATION WITHDRAWN]` with
the counterexample reproduced verbatim, which is the correct disposition. **One clause is
overstated**: *"it is the strongest mechanical check available short of a human reading every
pattern."* It is not a mechanical check at all — nothing relates the label to the pattern — and a
strictly stronger one is available and already in the SPEC's own vocabulary. `REQ-A16-005` makes an
**absence** claim owe a citation or a probe; a **presence** claim owes nothing. Extending
citation-or-probe to the IMPLEMENTED side — the rule's matched head symbol resolves to a named,
cited stdlib/framework/API surface, or a recorded probe shows it matching real code — is
mechanically checkable by the same checker, closes `probe-cwe-anchored` (there is no reference to
cite), and costs one criterion of the two in reserve. **PARTIAL.**

### 3. The split — **not a scope loss. This is genuine sequencing.**

Checked against both SPECs:

- SPEC B carries scope (`§B:81-88`), eight provisional milestones with language groupings and an
  explicit dependency note (`§E:161-180`), exclusions naming the predecessor (`§D:124-157`), and
  `depends_on: [SPEC-ASTGREP-LANG16-001]` in frontmatter.
- Card requirement (2) and requirement (4) are named as discharged **there, in full**, in three
  independent places: SPEC B `§A.1:46-49`, SPEC B `progress.md:23-24`, and SPEC A `§D:430-439`.
- SPEC A `acceptance.md` §F requires the PR body to state the split at review time, so it is visible
  rather than inferred.
- The inherited-contract table (SPEC B `§C:100-118`) is executable, not gestural: nine rows, each
  citing specific requirement ids, plus the M1 id-keying decision. `REQ-A16-017` is singled out with
  a paragraph explaining why the eight languages without an accidental literal have no forcing
  function otherwise. A successor can consume this without re-deriving anything.
- SPEC B's deliberate absence of requirements is argued (`§C:94-98`) rather than left blank, and its
  `progress.md` records `plan_status: skeleton`, `requirements: 0`, `NOT audit-ready`.

I looked for what could have vanished in the move and found nothing dropped: the 98 open cells, the
10 idiom rules, the 20 fixture pairs, the two `_uncovered` promotions, and the
`coveredCorpusLanguages` extension are all enumerated on the SPEC B side. Two small seams: the
inherited table does not list `REQ-A16-009` / `REQ-A16-022` / `REQ-A16-023`, and it inherits
`REQ-A16-010` — whose verification gap (E5) therefore propagates into the successor's contract. Both
are addressable in SPEC B's own plan phase. **This is not a silent trim wearing a sequencing
costume.**

### 4. Monotonic improvement — per iteration-1 blocking finding

| iter-1 | Status | Evidence checked |
|---|---|---|
| D1 corpus-gate skips | **PARTIAL** | Four passages corrected and verified against `pre_tool_scan_differential_test.go` (t.Skip 242, assertion loop 245); §A.7 enforcement table row-accurate. But `REQ-A16-017` binds no SPEC A criterion, and the same vacuous-green shape reappears unguarded in `AC-A16-001` (measured, E1). |
| D2 contrived rule | **PARTIAL** | `metadata.cwe` anchor added (`REQ-A16-011`), R2 marked `[MITIGATION WITHDRAWN]` with the counterexample verbatim. Defeated on first attempt by `probe-cwe-anchored` (exit 0). Honestly disclaimed; one clause overstated. |
| D3 archived SPEC | **FIXED** | `SPEC-ASTG-UPGRADE-001` measured `status: archived`; reconciled at six sites; R6 ownership transferred to this SPEC; dropped from frontmatter. |
| D4 numbering gap | **FIXED** | 23 contiguous definitions, no `005a`, no `017` hole. |
| D5 REQ-011 not a requirement | **FIXED** | `spec.md:329-330` is `Where … shall`, counterexample preserved. |
| D6 REQ-025 unverified | **FIXED, with a new inverse gap** | `AC-A16-018` gives the corpus baseline two `git diff` commands and §G's exclusion is gone. But the requirement it verified was deleted in the renumbering, so the AC is now orphaned (E8). |
| D7 checker cardinality | **FIXED** | `REQ-A16-006` class 1 is a set comparison against the Cartesian product; class 4 (dangling rule id) added; `AC-A16-010` exercises four synthetic defects including the count-preserving substitution. Residue: `design.md:192` still says "three classes" above four (E6). |
| D8 split required | **FIXED** | See question 3. |
| D9 AC-015 collision | **FIXED for AC-020, reintroduced at AC-019** | Measured: `grep -rniE "primary\|planned\|unsupported" $T` returns exactly `sgconfig.yml:10`, which `AC-A16-020` names as the sanctioned exemption — the criterion is now correct. But the neutrality widening produced the same failure shape one criterion later (E3). |
| D10 dangling ref | **FIXED** | `REQ-A16-022a` survives only in the HISTORY row documenting its removal. |
| D11 tautological AC | **FIXED, partially undone** | The F8-severity criterion is retired. Six criteria still hold on the untouched tree, and `AC-A16-001` is newly among them (E1). |
| D12 catalog.yaml (optional) | **APPLIED, AND WRONG** | See E2 — measured. |
| D13 rule-tests placement (optional) | **APPLIED in `spec.md`/`plan.md`, NOT in `design.md`** | See E4. |
| D14 neutrality scope (optional) | **OVER-APPLIED** | I asked to widen the *text kind*; the scope was widened to the whole template tree as well (E3). |
| D15 unobservable obligation (optional) | **FIXED** | `grep -rn 'shall read'` returns only the HISTORY/progress audit trail. |
| D16 stray directory (optional) | **FIXED** | `plan.md` §A.5 carries the pre-flight `rmdir` with the rationale. |
| D17 frontmatter field (optional) | **FIXED** | `related_specs:` removed. |

### 5. New defects introduced by the revision — E1…E8 below.

### 6. The D13 relocation — **the hazard is real, the new location is correct, and the harness still works. The reasoning was not propagated into `design.md`.**

- **Hazard premise verified.** `internal/template/templates/**` is the distributed tree, and the
  deployed ruleset's scanned config resolves to the project's own
  `.moai/config/astgrep-rules/sgconfig.yml` — the deploy target of that same path. So fixtures placed
  under `$T/rule-tests/` would land inside the tree the user's own pre-write gate scans, and by
  `REQ-A16-008` each contains a deliberately dangerous construct. The reasoning holds.
- **`sg test` still resolves tests from outside.** I built a config whose `testConfigs` points above
  itself and ran it:
  ```
  # /tmp/t228i2/proj/sgconfig.yml → testConfigs: [{testDir: ../outside-tests}]
  $ sg test --config /tmp/t228i2/proj/sgconfig.yml
  PASS probe-cwe-anchored  ..
  test result: ok. 1 passed; 0 failed;    EXIT=0
  ```
  `testDir` resolves relative to the config file and traverses upward without complaint. The
  relocation does **not** break the harness. `plan.md:114-115`'s additional check — that the deployed
  config stays valid where the path is absent — is the right residual concern and is stated.
- **Not propagated.** `design.md:119` still reads *"`testConfigs: [{testDir: rule-tests}]` in the
  template `sgconfig.yml`, with fixtures under `$T/rule-tests/`"*, and `design.md:154`'s layout
  diagram repeats it inside the `$T/` tree — 44 lines above `design.md:163-169`, which states the
  opposite and cites `REQ-A16-019`. An implementer reading §3.1 first builds the shipping variant.
  This is E4.

### 7. Template neutrality — **no violation created by the SPEC.**

`grep -rniE 'SPEC-[A-Z]|20[0-9]{2}-[0-9]{2}-[0-9]{2}'` over the astgrep-rules template dir → rc=1.
All 16 languages carry equal weight and no rule directory is described as primary or planned;
`r` and `flutter` are excluded by a version-scoped parser fact under `ast-grep 0.40.5`, restated as
non-permanent at `spec.md:77-78`. Existing rule messages are English. The defect is in the
*criterion* that guards this (E3), not in the SPEC's own content.

### Criteria that would pass on a tree where nothing was implemented

Flagged as requested. `AC-A16-001` (measured: `Running 0 tests … ok`, EXIT=0 — this is E1 and it is
the serious one), `AC-A16-016` and `AC-A16-017` (documentation reads against prose already written in
this revision), `AC-A16-018` (both diffs empty today by construction), `AC-A16-020` (measured: the
single sanctioned match), `AC-A16-022` (true today by construction), `AC-A16-023` (already measured
working). Four of these are legitimate regression guards and should stay. `AC-A16-001` is not a
guard of anything — it is the milestone-1 stop condition, and it observes nothing.

---

## Defects Found (structured defect-list)

E1. **central-harness-criterion-passes-vacuously** — `acceptance.md:20-23` (AC-A16-001),
`plan.md:128-132` (M1 stop condition) — **The SPEC's own no-vacuous-green doctrine is not applied to
its central criterion.** Measured on the untouched tree:
`sg test --config internal/template/templates/.moai/config/astgrep-rules/sgconfig.yml` →
`Running 0 tests` / `test result: ok. 0 passed; 0 failed;` / EXIT=0. The criterion's Then clause is
*"it exits 0 and reports zero failures and zero missing snapshots"*, all three of which hold with
nothing built, because the shipped `sgconfig.yml` declares no `testConfigs` and `sg test` reports a
zero-test run as success. The criterion's own title claims coverage of "the existing 26 rules" that
the assertion never checks. This is the identical failure shape `REQ-A16-017` exists to forbid for
the corpus — a green run that observes nothing — reproduced inside the harness milestone.
  — Severity: **critical** — Class: **blocking** — Required fix: extend `AC-A16-001`'s Then clause to
  assert the run was non-empty and complete — the output reports **26** passing cases (or the count
  the M1 id-keying decision yields), and `sg test`'s reported test count equals the number of rules in
  the shipped ruleset. Mirror `REQ-A16-017`'s wording so the two read as one discipline.

E2. **catalog.yaml obligation is factually wrong** — `spec.md:386-393` (REQ-A16-018),
`acceptance.md:170-174` (AC-A16-021), `plan.md:222-224` — **`internal/template/catalog.yaml` does not
track the ruleset, so a ruleset edit produces no catalog diff and `AC-A16-021` cannot pass on correct
work.** Measured: `grep -c 'astgrep' internal/template/catalog.yaml` → 0; `grep -c 'config'` → 0. The
catalog's entries are skills and agents only (`catalog:` → `core` / `optional_packs` /
`harness_generated`, each holding `skills:` / `agents:` lists), and
`internal/template/scripts/gen-catalog-hashes.go` hashes a skill's root `SKILL.md` or an agent's
`*.md` — nothing else. I appended a comment to
`internal/template/templates/.moai/config/astgrep-rules/security/web.yml`, ran the generator, and
`git status --porcelain internal/template/catalog.yaml` was **empty** (probe reverted). `AC-A16-021`'s
first clause — *"it also carries `internal/template/catalog.yaml`"* — is therefore unsatisfiable by a
correct implementation, and the predictable response is to hand-edit the catalog to manufacture a
diff. (Its second clause, "a fresh `make build` produces no further diff", is sound: I confirmed the
generator is idempotent — regeneration on an unchanged tree produced a byte-identical file.)
  — Severity: **critical** — Class: **blocking** — Required fix: restate `REQ-A16-018` and
  `AC-A16-021` as what is actually true and load-bearing: the template ruleset is the source of truth,
  it is compiled in by `//go:embed all:templates` with **no** committed artifact, and `make build`
  must leave `internal/template/catalog.yaml` unchanged — i.e. the criterion becomes *"a fresh
  `make build` on that commit produces no diff at all, including none in `catalog.yaml`"*. Keep the
  §A.5-style note explaining why a reader might have expected an artifact and why there is none.

E3. **AC-A16-019 fails today on content this SPEC does not touch, and misstates its own CI guard** —
`acceptance.md:151-156`, `spec.md:408-410` (REQ-A16-021) — the criterion's Given is *"every file under
`internal/template/templates/`"* and its Then requires *"no SPEC-ID / date / SHA match is found"*.
Measured: `grep -rlE 'SPEC-[A-Z][A-Z0-9]+-[0-9]{3}' internal/template/templates/` → **22 files**
(including `.claude/agents/moai/plan-auditor.md`, `.claude/skills/moai/SKILL.md`,
`.claude/rules/moai/development/spec-frontmatter-schema.md`);
`grep -rlE '20[0-9]{2}-[0-9]{2}-[0-9]{2}'` → **90 files**. The criterion fails on the untouched tree
by a wide margin, against files in no way owned by this SPEC. It also conjoins that manual scan with
*"the CI neutrality guard passes"* as though they were one standard: the guard
(`.github/workflows/template-neutrality-check.yaml`) runs `TestTemplateNeutralityAudit` over classes
C1/C2/C4/C5/C6/C8 — a SPEC-ID is **not** among them, and the date class C3 and hash class C7 belong to
a different test. This is the iteration-1 D9 shape reintroduced by over-widening my own D14, which
asked to widen the *kind of text* covered, not the *set of files* scanned.
  — Severity: **major** — Class: **blocking** — Required fix: narrow the Given of `AC-A16-019` and the
  scope of `REQ-A16-021` to the files this SPEC writes —
  `internal/template/templates/.moai/config/astgrep-rules/**` — keeping D14's widening of the *text
  kind* (`message:` / `note:` / YAML comments / any human-language string). State the CI guard's role
  accurately: it runs its own class set on changed template paths, and it is not the same check as the
  SPEC-ID scan.

E4. **design.md contradicts the D13 placement decision it elsewhere states** —
`design.md:119` (§3.1), `design.md:154` (§4 layout) vs `design.md:163-169`, `spec.md:395-402`
(REQ-A16-019), `acceptance.md:46-50` (AC-A16-005) — §3.1 reads *"`testConfigs: [{testDir: rule-tests}]`
in the template `sgconfig.yml`, with fixtures under `$T/rule-tests/`. Snapshots are generated and
committed"*, and the §4 layout diagram places `testConfigs: [{testDir: rule-tests}]` on the `$T/`
`sgconfig.yml` line. Both are the pre-D13 design, and both survive 44 lines above the paragraph
stating the opposite and citing `REQ-A16-019`. An implementer reading the design in order builds the
shipping variant and is corrected only by an acceptance criterion at close. The relocation itself is
sound — I verified `sg test` resolves `testDir: ../outside-tests` from the config's own directory,
exit 0 — so this is unpropagated residue, not a wrong decision.
  — Severity: **major** — Class: **blocking** — Required fix: rewrite `design.md` §3.1's wiring
  sentence and §4's layout comment to name the repo-side root, matching `design.md:163-169`. State
  in §3.1 the measured fact that `testDir` resolves relative to `sgconfig.yml` and traverses upward,
  so the placement is known-feasible rather than assumed.

E5. **REQ-A16-010 has no criterion, and the tool does not enforce it** — `spec.md:307-309`,
`acceptance.md` (absent) — *"A rule shall not ship without both test cases"* maps to no acceptance
criterion. `AC-A16-001` does not cover it (E1); `AC-A16-012` covers only `error`-severity rules;
`AC-A16-013` covers only security-family rules. That leaves the twelve non-security rules in the `go/`
directory unguarded. And the tool is silent about it — measured with a two-rule config where only one
rule had a test file:

```
Running 1 tests
PASS probe-cwe-anchored  ..
test result: ok. 1 passed; 0 failed;    EXIT=0
```

`probe-no-tests` shipped with no cases at all and produced no failure, no warning, and no mention.
The requirement is the one SPEC B explicitly inherits (`§C` table row 2), so the gap propagates.
  — Severity: **major** — Class: **blocking** — Required fix: add a criterion asserting that the count
  of rules carrying a `valid`/`invalid` pair equals the count of rules in the shipped ruleset (26 at
  M1) — the same enumeration E1 requires, which lets one criterion discharge both if worded to compare
  `sg test`'s reported case count against the rule count.

E6. **renumbering and class-count residue in `design.md`** — `design.md:70` cites `AC-A16-007` for the
unevidenced-exemption class (correct: `AC-A16-008`); `design.md:110` cites `REQ-A16-014` for the
matrix annotation obligation (correct: `REQ-A16-015` — `014` is the shape-matcher prohibition);
`design.md:180` cites `REQ-A16-021` for the `ruleDirs` obligation (correct: `REQ-A16-023` — `021` is
neutrality); `design.md:274` cites `spec.md` §A.4 for the `ShouldAlert` → `DecisionDeny` wiring
(correct: §A.5 — §A.4 is the family table); and `design.md:192` says *"any of the three classes
below"* immediately above four enumerated classes, left over from the D7 fix. Each resolves to a real
but **wrong** id, which is worse than a dangling one: it reads as correct.
  — Severity: **minor** — Class: **blocking** — Required fix: correct the four citations and the
  class count. Mechanical.

E7. **two criteria owned by no milestone** — `plan.md:132`, `plan.md:165`, `plan.md:196` — M1 declares
"Verified by: AC-A16-001 … AC-A16-005", M2 "AC-A16-006 … AC-A16-011", M3 "AC-A16-012 … AC-A16-015,
plus AC-A16-019 … AC-A16-023". `AC-A16-016` and `AC-A16-017` (acceptance.md §D, enforcement honesty)
appear in no milestone's verification set. `AC-A16-018` is at least covered by `plan.md:70` and the
§E standing command, though not by a "Verified by" line.
  — Severity: **minor** — Class: **blocking** — Required fix: attach §D's three criteria to a
  milestone — M1 is the natural home, since §A.7's table is the input the successor inherits before
  any breadth work starts.

E8. **AC-A16-018 has no requirement behind it** — `acceptance.md:140-145`, `spec.md` §C (absent) — the
corpus-baseline criterion added in response to iteration-1 D6 verifies an obligation that exists only
as §D prose (`spec.md:440`, *"This SPEC touches no file under …scan-corpus/"*) and a `plan.md` PRESERVE
entry. The requirement it used to verify (former `REQ-A16-025`) was deleted in the renumbering, so
D6's fix landed in the verification layer with nothing above it. Low practical risk — the criterion is
runnable and the exclusion is normative — but the REQ↔AC mapping is now one-directional.
  — Severity: **minor** — Class: **optional** — Required fix: either add a one-line unwanted-form
  requirement in §C.5 (*"This SPEC's commits shall not modify any file under
  `internal/hook/security/testdata/scan-corpus/` nor any recorded `wantDeny` value."*) — there is
  headroom for it at 23/25 — or note in `AC-A16-018` that it verifies the §D exclusion rather than a
  requirement.

E9. **D2 residual — the anchor is asymmetric with the SPEC's own evidence standard** —
`spec.md:311-325` (REQ-A16-011), `spec.md:322-325` (the necessary-not-sufficient note),
`acceptance.md:105-109` (AC-A16-013) — recorded separately from the iteration-1 D2 status because it
is a defect in the current text, not a restatement. `REQ-A16-005` requires an **absence** claim to
carry a citation or a probe; a **presence** claim carries nothing comparable, and my
`probe-cwe-anchored` counterexample (exit 0, real CWE id) exploits exactly that asymmetry. The claim
that the anchor is *"the strongest mechanical check available short of a human reading every
pattern"* is not accurate while the citation-or-probe standard the SPEC already operates remains
unapplied to the implemented side.
  — Severity: **major** — Class: **optional** — Required fix (one criterion, and headroom exists):
  extend `REQ-A16-011` / `AC-A16-013` so an IMPLEMENTED cell owes the same evidence class its EXEMPT
  neighbours owe — the pattern's matched head symbol resolves to a named language, stdlib, or
  framework reference, **or** a recorded probe shows the pattern matching real code from that
  ecosystem. Then soften the overstated clause to what remains true: the anchor is a reviewer-checkable
  label, and the citation is the mechanical part.

---

## Regression Check (iteration 1 → 2)

Per-finding status is the table under audit question 4. Summary: of the eight blocking findings,
**six FIXED** (D3, D4, D5, D6, D7, D8, plus D9/D10 of the remainder), **two PARTIAL** (D1, D2). Of the
six optional findings, three FIXED (D15, D16, D17), one applied incorrectly (D12 → E2), one applied
incompletely (D13 → E4), one over-applied (D14 → E3).

No defect appears unchanged across both iterations, so **no stagnation is flagged**. The score is flat
(0.68 → 0.69) rather than regressing, so **no STOP escalation is raised** — but the flatness is
diagnostic: the revision traded old defects for new ones of the same class. Both critical findings
this round (E1 vacuous green, E2 unverifiable-by-construction obligation) are instances of the exact
two failure modes iteration 1 named — a defense that reads convincingly and is never executed, and a
claim about a mechanism nobody measured. `spec.md:223-225` says so itself: *"Both failures share a
shape: a defense that reads convincingly in prose and is never executed."*

---

## Recommendation

FAIL at 0.69 against a Tier L threshold of 0.85. All seven must-pass criteria now pass, so the
remaining work is score-driven and bounded. **One iteration remains** (max 3). Fix in this order; the
first two are the ones that change what gets built.

1. **E1 — make `AC-A16-001` count.** Assert the number of test cases that ran, not merely that the
   run exited 0. Measured today: `Running 0 tests … ok`, EXIT=0 on the untouched tree. This is the
   single highest-value fix in the document, because M1's stop condition is currently satisfiable by
   doing nothing, and everything downstream rests on M1.
2. **E2 — correct the `catalog.yaml` claim.** Measured: a ruleset edit changes no catalog entry.
   Invert the obligation — `make build` must leave `catalog.yaml` **unchanged** — and keep the note
   explaining why a reader expects an artifact and why there is none. (This corrects a defect my own
   iteration-1 D12 invited; the SPEC is what is audited, but the origin is worth recording.)
3. **E3 — re-narrow `AC-A16-019` / `REQ-A16-021` to the ruleset directory**, keeping D14's widening of
   the text *kind*, and state the CI guard's actual class set rather than implying it checks SPEC IDs.
4. **E4 — propagate D13 into `design.md` §3.1 and §4.** The decision is right and I confirmed the
   harness still resolves an outside `testDir` at exit 0; the design simply still states the old one.
5. **E5 — give `REQ-A16-010` a criterion.** Measured: `sg test` ignores an untested rule in silence.
   If E1 is worded as a count comparison, one criterion discharges both.
6. **E9 (optional, but the highest-leverage optional)** — close the presence/absence evidence
   asymmetry so the anchor becomes partly mechanical, and soften the one overstated clause. Headroom
   exists at 23/25 in both layers.
7. **E6, E7, E8** — mechanical: four stale cross-references and a "three classes" count in
   `design.md`; attach `AC-A16-016`/`017` to a milestone; give `AC-A16-018` a requirement or relabel
   it as verifying the §D exclusion.

**What the revision got right**, since a second FAIL should not obscure a document that improved
substantially. Three must-pass failures were closed cleanly and I verified each against the tree
rather than the prose. The D8 split is the strongest part of this revision: SPEC B is a real skeleton
with an executable inherited-contract table, both card requirements are explicitly discharged there in
three independent places, and nothing measurable was dropped in the move — this is the opposite of a
silent trim. §A.7's enforcement table is row-accurate against the code, including the row that admits
*"Does not exist. It is a skip."* §A.8's record of two falsified defenses, and R2's
`[MITIGATION WITHDRAWN]` marker with the counterexample reproduced verbatim, are the correct
disposition for a claim that failed under test — most SPECs quietly delete such a claim. The D13
relocation reasoning is sound and I confirmed both halves of it: the hazard is real, and the new
location keeps `sg test` working. The checker's four classes are now genuinely sharp, and
`AC-A16-010` case (b) targets precisely the count-preserving substitution a lesser checker misses.
The defects that remain are, again, in the defenses' reach rather than in the SPEC's honesty.
