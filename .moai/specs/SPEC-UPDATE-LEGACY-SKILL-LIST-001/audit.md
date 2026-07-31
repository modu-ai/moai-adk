# Plan-Phase Audit — SPEC-UPDATE-LEGACY-SKILL-LIST-001

Iteration: 1
Auditor: plan-auditor (adversarial, M1-M5 bias-prevention protocol active)
Audited at: 2026-07-31, tree `main` @ `9a6b6c854`, working tree clean apart from this SPEC directory + one pre-existing untracked report HTML.
Artifacts read: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M input contract: spec + plan + acceptance; `research.md` correctly absent).

**Reasoning context ignored per M1 Context Isolation.** The delegation brief's framing was treated as a set of hypotheses to test, not as findings to accept. Every claim below was re-derived from the tree by a command whose verbatim output is quoted.

---

## Verdict

| | |
|---|---|
| **Verdict** | **FAIL** (marginal) |
| **Overall Score** | **0.79** (harmonic mean of 4 dimensions) |
| **Tier** | M — threshold **0.80** |
| **Must-pass firewall** | no MP failure; the FAIL is on the aggregate + 4 MUST-FIX defects |

This is a strong SPEC. Its evidence base is the best of the Epic so far: **28 of 31 independently re-run baseline commands reproduced the recorded value exactly**, every `file:line` citation resolves, and the `§A` verification-discipline preamble anticipates three of the four Epic defect classes correctly. It fails on a narrow, cheaply-fixable set: one stale decision premise, one internally contradictory AC pair, one provably-blind AC, and one destructive falsification step.

Three of the four Epic-wide defect classes recurred here in reduced form — that is the reason for FAIL rather than PASS-with-debt on a 0.79/0.80 margin.

---

## Must-Pass Results

- **[PASS] MP-1 — REQ number consistency.** `REQ-LSL-001 … REQ-LSL-017` sequential, no gaps, no duplicates, consistent 3-digit padding; `NFR-LSL-001 … 005` likewise; `AC-LSL-001 … 014` likewise.
  ```
  $ grep -o 'REQ-LSL-[0-9]*' spec.md | sort -u | sort -t- -k3 -n | tr '\n' ' '
  REQ-LSL-001 REQ-LSL-002 ... REQ-LSL-016 REQ-LSL-017
  $ grep -o '^### AC-LSL-[0-9]*' acceptance.md | tr '\n' ' '
  ### AC-LSL-001 ... ### AC-LSL-014
  ```
- **[PASS] MP-2 — GEARS format compliance.** 15/17 REQ are clean ubiquitous (`The <subject> shall …`), event-driven (`REQ-LSL-005`: "When the Go test suite runs, the guard test shall …"; `REQ-LSL-007`: "When `template.EmbeddedMoaiSkillNames()` returns an error, or returns an empty set, the guard shall skip …") or Where-form (`REQ-LSL-012`). Two minor deviations recorded as SHOULD-FIX (S5).
- **[PASS] MP-3 — YAML frontmatter validity.** All 12 canonical fields present with correct types (`spec.md:2-15`); no rejected snake_case alias. Enforced-lint confirmation:
  ```
  $ moai spec lint | grep -i 'LEGACY-SKILL-LIST'   → no match (exit 1)
  $ moai spec lint | tail -1
  0 error(s), 62 warning(s)
  ```
  The live ID rule is `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (`internal/spec/lint.go:715`), which the multi-segment `SPEC-UPDATE-LEGACY-SKILL-LIST-001` satisfies. (The rule-file `spec-frontmatter-schema.md` still prints the older single-segment regex — that is a doc/code divergence in the rule file, not a defect of this SPEC.) Three style deviations recorded as SHOULD-FIX (S6).
- **[N/A] MP-4 — Section 22 language neutrality.** Single-language (Go) internal-CLI SPEC; `NFR-LSL-002` explicitly forbids touching `internal/template/templates/`, verified `git status --porcelain internal/template/templates/ → 0`. Auto-passes.
- **[PASS] MP-5 — D7 cross-SPEC reconciliation.** All 8 referenced SPEC IDs resolve to existing directories; none is `retired` / `superseded` / `archived`:
  ```
  SPEC-CONFIG-KEY-HONESTY-001 status=draft        SPEC-CONFIG-TIER-PERSIST-001 status=draft
  SPEC-UPDATE-CI-GUARD-001 status=draft           SPEC-UPDATE-DATA-SURVIVAL-001 status=draft
  SPEC-UPDATE-DOC-DRIFT-001 status=draft          SPEC-UPDATE-REINSTALL-LOOP-002 status=draft
  SPEC-UPDATE-LEGACY-SKILL-LIST-001 status=draft  SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 status=implemented
  ```
  `SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001` (`implemented`) is explicitly reconciled in `spec.md` § "Relationship to …" — its out-of-scope declarations were re-verified and all three citations resolve exactly (see V9 below).
- **[PASS] MP-6 — D8 cross-platform discipline.** `grep -c 'syscall'` → `0` in all four artifacts. Auto-PASS. Note the SPEC nonetheless reasons about cross-platform failure injection (`plan.md` M4 step 1, `acceptance.md` AC-LSL-010 last bullet) and correctly rejects the `chmod 0o000` fixture in favour of a portable `ENOTDIR`-class one — a positive finding.
- **[PASS] MP-7 — clarification gate.** `grep -rn '\[NEEDS CLARIFICATION' plan.md` → no match (exit 1). `research.md` does not exist (correct for Tier M). The only literal occurrence anywhere is `progress.md:60`, which is prose *declaring* that no marker was needed — a self-description, not an open marker.

---

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|-----------|------:|-------------|----------|
| Clarity | 0.85 | 0.75–1.0 | Every decision in `plan.md §A` carries its rejected alternative (D2 option B, D4 hand-rolled string, D5 blunt-`.gitignore`). One unstated implementation constraint (M2 test shape — see **M2** below) and one D3↔AC-007 wording tension. |
| Completeness | 0.85 | 0.75–1.0 | All required sections present; 6 × `### Out of Scope — <topic>` H3 headings each with `-` bullets (`grep -c '^### Out of Scope —' → 6`), satisfying `OutOfScopeRule`. Gaps: `archiveVersion` invariance (§B Goal 5 / NFR-LSL-003) has no mechanical AC; `REQ-LSL-016` has no mechanical AC; M4 `continue`-semantics unspecified. |
| Testability | 0.65 | 0.50–0.75 | 9/14 ACs are solid and baseline-reproduced. Three carry hard defects (**M2**, **M3**, and AC-LSL-005's shared count assumption), two more are brittle (S1, S3). `REQ-LSL-016` is verified only by "checkable by reading it". |
| Traceability | 0.85 | 0.75–1.0 | All 17 REQ + 5 NFR covered by an AC. One orphan: `AC-LSL-006` traces to "M1 step 3-4", a plan milestone, not a REQ. One inversion: `REQ-LSL-012`'s Where-branch is contradicted by its own AC (**S4**). |

Harmonic mean = 4 / (1/0.85 + 1/0.85 + 1/0.65 + 1/0.85) = **0.789 → 0.79**.

---

## Baseline Re-Verification (Epic defect class 1)

I re-ran every AC's recorded pre-fix baseline against the current tree. **This is the SPEC's strongest area.** Results:

| AC | Recorded baseline | Re-measured | Match |
|----|-------------------|-------------|:-----:|
| AC-LSL-001(a) | `16` | `16` | ✓ |
| AC-LSL-001(b) | 13 IDs preceded by the 3 offenders | exact, same order | ✓ |
| AC-LSL-002 | `3` | `3` | ✓ |
| AC-LSL-002 companion | all 3 template `SKILL.md` present | no output | ✓ |
| AC-LSL-003(a) | `1` | `1` | ✓ |
| AC-LSL-003(b) | `"lists the 16 skill IDs removed in BC-V3R3-007"` | verbatim match | ✓ |
| AC-LSL-004(a) | `0` | `0` | ✓ |
| AC-LSL-005 | probe `legacySkillIDs=16 embedded=30 overlap=3` | intersection re-derived = 3 (see below) | ✓ |
| AC-LSL-006(a) | `ok … 0.776s`, 0 FAIL | `ok … 2.295s`, `--- FAIL` count `0` | ✓ |
| AC-LSL-006(b) | `[0][1][2][3] [:5] [:8]`, max idx 3, max bound 8 | exact — 7 sites, same values | ✓ |
| AC-LSL-006(c) | `2` and `2` | `2` and `2` | ✓ |
| AC-LSL-007(b) | `0` | `0` | ✓ |
| AC-LSL-008 | `3` | `3` | ✓ |
| AC-LSL-009(a) | `7` | `7` | ✓ |
| AC-LSL-009(b) | the 4 genuine + 3 offenders | exact | ✓ |
| AC-LSL-009(c) | `0` | `0` | ✓ |
| AC-LSL-009(d) | `0`, grep exits 1 | `0`, exit 1 | ✓ |
| AC-LSL-010(a) | `0` | `0` | ✓ |
| AC-LSL-012(a) | `3` (lines 302, 305, 320) | `3`; lines confirmed 302/305/320 | ✓ |
| AC-LSL-012(b) | `0` | `0` | ✓ |
| AC-LSL-013 | `1` and `2` | `1` and `2` | ✓ |
| AC-LSL-014(a) | exit 0, no output | exit 0, no output | ✓ |
| AC-LSL-014(b) | `0` | `0` | ✓ (but see **M3** — the value is uninformative) |
| AC-LSL-014(c) | `update_archive.go` IS listed | listed; `gofmt -d` shows exactly the claimed single import-order hunk | ✓ |
| AC-LSL-014(d) | `0` | `0` | ✓ |

Supporting-evidence claims outside the AC table, also re-run:

```
$ for s in moai-domain-backend moai-domain-frontend moai-domain-database; do
    grep -rl "$s" .claude/{agents,rules,skills} .moai/config \
      internal/template/templates/.claude/{agents,rules,skills} \
      internal/template/templates/.moai/config | grep -v "/$s/" | wc -l; done
moai-domain-backend  files=56 occ=68
moai-domain-frontend files=26 occ=34
moai-domain-database files=30 occ=30
```
→ `spec.md §A` Defect 1 table (56/68, 26/34, 30/30) reproduced **exactly**. The brief's 46/17/25 was indeed wrong and the SPEC corrected it correctly.

```
$ md5 -q .claude/skills/<id>/SKILL.md  vs  .moai/archive/skills/v2.16/<id>/SKILL.md
moai-domain-backend  live=e9edb94ad7fbdca57cff8a96e54fd11f arch=7469bb4e09bae816ea7268e940368095
moai-domain-database live=534149429f1db713b1df268dbde9dc60 arch=ea4d38eb507dd73a47e7da2ff80a63bc
moai-domain-frontend live=080bc37188d9e5ea7a5af7838b22013e arch=4f58f1b22314a95f36321ad1ad7f8bbc
```
→ `§A` Defect 2 md5 table reproduced **exactly, all six hashes**.

```
$ git log -S"moai-domain-backend" --oneline -- internal/cli/update_archive.go
ec0e9e257 feat(update,migrate): archive 마이그레이터 + restore-skill 서브커맨드 (M4)      ← exactly 1 ✓
$ git ls-tree -r --name-only 74bae50f4^ -- .claude/skills/moai-domain-backend
SKILL.md  references/examples.md  references/reference.md                                ← ✓
$ git log -1 --format='%h %ad' --date=short -- .moai/archive/.../moai-domain-backend/SKILL.md
9373e558f 2026-05-11                                                                     ← ✓
$ 74bae50f4 2026-04-27 / ec0e9e257 2026-04-27 / 697a6e2c7 2026-04-28                      ← ✓ all three dates
$ grep -rl 'legacySkillIDs' internal/cli/ | wc -l  → 7 (6 _test.go + the definition)      ← "six test files" ✓
$ ls -d internal/template/templates/.claude/skills/*/ | wc -l                      → 31
$ grep -c 'path: templates/.claude/skills/' internal/template/catalog.yaml         → 31
$ ls -d .claude/skills/moai-*/ | wc -l → 30 ; ls -d .claude/skills/hns-*/ | wc -l → 7     ← 31=31=31 ✓
$ for s in <the 13 retained>; do [ -d tmpl/$s ] || [ -d local/$s ] && echo PRESENT; done
(no output — all 13 genuinely absent from both trees)                                     ← ✓
```

**Conclusion on defect class 1 (baseline mis-attribution): substantially clean.** One stale premise found — see **M1** — and one number-ordering error — see **S7**.

---

## Defect-Locus Re-Verification (Epic defect class 4)

Every cited location was opened and checked against what the SPEC claims it contains. **All 13 citations resolve; several are exact to the line.**

| Citation | Claimed content | Actual | Verdict |
|----------|-----------------|--------|:-------:|
| `update_archive.go:33-50` | the `legacySkillIDs` list | `:33` `var legacySkillIDs = []string{`, `:50` `}` — exact block bounds | ✓ exact |
| `:66-71` | source-absent short-circuit | `:66` `if _, err := os.Stat(srcDir); err != nil {` … `:71` `}` | ✓ exact |
| `:129-165` | drift check | `:129` `func checkArchiveDrift(…)`, `:165` closing `}` | ✓ exact |
| `:271-335` | the loop / function | `:271` `func archiveLegacySkills(…)`, `:335` closing `}` | ✓ exact |
| `:319-321` | the early return | `:319` `if err := archiveSkill(…)`, `:320` `return archived, fmt.Errorf("archive %s: %w", …)`, `:321` `}` | ✓ exact |
| `:296-317` | `--force` drift-backup path | `:296` `if force && alreadyArchived {`, `:317` closing `}` | ✓ exact |
| `update.go:449-457` | post-sync archive call | `:442` `tui.Section("Post-sync steps"…)`; the call is at `:454`, inside the `{ … }` block spanning `:449-457` | ✓ |
| `deploy/deploy.go:38-68` | clean targets | `:38` `targets := []cleanTarget{`, `:68` closing `}` | ✓ exact |
| `deploy/deploy.go:51-55` | the `moai*` glob target | `:51-55` = the `SkillsSubdir, "moai*"` entry with `isGlob: true` at `:54` | ✓ exact |
| `deploy/deploy.go:115-127` | `.moai/config` removal | `:115` `configDir :=`, `:121` `os.RemoveAll(configDir)`, `:127` `plConfig.Done(…)` | ✓ exact |
| `migrate_restore_skill.go:29,44,68-83` | restore hazard | `:29` `func restoreSkill(…)`, `:44` `archiveDir :=`, `:68-83` = the `--force` `os.RemoveAll(targetDir)` → `MkdirAll` → `copyDirAll` region | ✓ |
| `skills_manifest.go` | `EmbeddedMoaiSkillNames` + contract | present; doc comment says verbatim "callers MUST treat an empty derived set as 'manifest unavailable' and degrade gracefully"; `moaiSkillPrefix = "moai-"` trailing-dash note in `plan.md §A D2` is correct | ✓ |
| `skills_removal_test.go` | narrowed 9-entry `removed` slice + stale comment | 9 entries; the "NOT removed (still exist)" comment lists 7 IDs of which 4 (`electron`, `auth`, `deployment`, `chrome-extension`) are in fact now absent — the SPEC's own staleness observation is correct | ✓ |
| `catalog.yaml` 157/159, 162/164, 220/222 | name+path per skill | `:157` `name: moai-domain-backend` / `:159` `path:`; `:162`/`:164` database; `:220`/`:222` frontend | ✓ exact |
| prior SPEC `spec.md:76`, `plan.md:44`, `acceptance.md:268` | list declared out of scope | `spec.md:76` "`legacySkillIDs` 목록의 변경은 다루지 않는다"; `plan.md:44` "`legacySkillIDs` 목록 (16개) 변경 없음."; `acceptance.md:268` "`legacySkillIDs` 목록 변경." | ✓ exact |

**Conclusion on defect class 4 (defect-locus mis-attribution): clean. Zero findings.** This is the best citation hygiene I have seen in this Epic.

---

## MUST-FIX

### M1 — `plan.md §A D1`: the decision's entire stated rationale is refuted by the SPEC's own declared baseline (Epic defect class 1)

**Artifact/location:** `plan.md:13` (`§A` D1), reinforced at `plan.md:11`.

**What is wrong.** D1 justifies not renumbering the six siblings with:

> "All six are mid plan-audit revision (the Epic audit returned FAIL on 6/6 and revision work is in flight on `plan/epic-update-config-audit`). Editing their HISTORY denominators **from a different branch** would collide with that revision work for zero functional gain."

Three independent facts at `9a6b6c854` — the SPEC's own declared baseline — contradict this:

```
$ git diff --stat main plan/epic-update-config-audit
(empty — zero content difference between main and the branch)

$ git rev-list --count --left-right main...plan/epic-update-config-audit
0	17
$ git log --oneline -2 plan/epic-update-config-audit
f89de81b9 Merge remote-tracking branch 'origin/main' into plan/epic-update-config-audit
9a6b6c854 docs(epic): update/config 4-lens audit Epic — 6 SPEC plan-phase artifacts (E1-E6) (#1258)
$ git merge-base --is-ancestor 9a6b6c854 plan/epic-update-config-audit && echo YES
YES
```
The 17 "ahead" commits are merge commits plus the pre-squash originals of what already landed; the trees are byte-identical.

```
$ git status --porcelain .moai/specs/
?? .moai/specs/SPEC-UPDATE-LEGACY-SKILL-LIST-001/
```
All six sibling directories are **committed on `main`** (`git ls-files .moai/specs/<sibling>/ | wc -l` → 4 for each). The revised artifacts are at v0.2.0 on main:
```
$ git show main:.moai/specs/SPEC-UPDATE-CI-GUARD-001/spec.md | grep -E '^(version|priority|phase):'
version: "0.2.0"
priority: P1
phase: "v3.0.2"
```

And the revision is not "in flight" — it is done and audited:
```
$ grep -m1 -iE 'verdict' .moai/reports/plan-audit/*epic-update-config-iter2.md
SPEC-CONFIG-KEY-HONESTY-001 …   Verdict: **PASS**   (0.81)
SPEC-CONFIG-TIER-PERSIST-001 …  Verdict: **PASS**   (0.82)
SPEC-UPDATE-CI-GUARD-001 …      Verdict: **PASS**   (0.85)
SPEC-UPDATE-DATA-SURVIVAL-001 … **Verdict: PASS**
SPEC-UPDATE-REINSTALL-LOOP-002 … | **Verdict** | **PASS** |   (and an iter3 PASS)
```

So: there is no different branch, there is no in-flight work, and "FAIL on 6/6" describes iteration 1, not the state at `9a6b6c854`. This SPEC is being authored **on `main`**, the same branch the siblings live on — the exact collision D1 invokes cannot occur.

**Why it matters beyond pedantry.** D1 is written in the same verified-evidence register as the rest of the plan, and `plan.md §C` opens with "Every `file:line`, count, and command in the delegation brief was re-run before being written into an artifact." D1's premise was not re-run. A reader — or the run-phase agent — will take it as measured. It is the one place in this SPEC where an unverified claim wears verified clothing, which is precisely Epic defect class 1.

**Required fix.** Either (a) restate D1's rationale on grounds that survive `9a6b6c854` — e.g. "renumbering six committed SPECs for a cosmetic denominator is churn with no functional gain, and belongs to whichever change lands last across the Epic" (this reason is already the second half of D1 and is sound on its own), or (b) if the collision argument is retained, replace it with a measured statement of the actual branch state. Do not leave the current text.

---

### M2 — `AC-LSL-004(b)` and `AC-LSL-007(a)/(b)` are mutually unsatisfiable as written (Epic defect class 3)

**Artifact/location:** `acceptance.md:109-112` (AC-LSL-004 b), `acceptance.md:189-199` (AC-LSL-007 a, b), and the same count assumption at `acceptance.md:140-142` (AC-LSL-005).

**What is wrong.** AC-LSL-007(a) requires subtests named `manifest_error` and `manifest_empty`, and 007(b) requires exactly 2 `--- SKIP` lines under `-run 'TestLegacySkillIDsNotEmbedded'`. AC-LSL-004(b) requires:

```bash
go test ./internal/cli/ -run 'TestLegacySkillIDsNotEmbedded' -count=1 -v 2>&1 \
  | grep -c -- '--- PASS: TestLegacySkillIDsNotEmbedded'
# EXPECT: 1
```

Go's `-v` output prints the **parent** `--- PASS: TestLegacySkillIDsNotEmbedded` *and* an indented `--- PASS: TestLegacySkillIDsNotEmbedded/<subtest>` for every passing subtest. The grep is unanchored, so both match. I built the exact shape the SPEC mandates and ran it:

```go
func TestLegacySkillIDsNotEmbedded(t *testing.T) {
	t.Run("production", func(t *testing.T) {})
	t.Run("manifest_error", func(t *testing.T) { t.Skip("manifest unavailable") })
	t.Run("manifest_empty", func(t *testing.T) { t.Skip("manifest empty") })
}
```
```
$ go test ./pkg/ -run 'TestLegacySkillIDsNotEmbedded' -count=1 -v
--- PASS: TestLegacySkillIDsNotEmbedded (0.00s)
    --- PASS: TestLegacySkillIDsNotEmbedded/production (0.00s)
    --- SKIP: TestLegacySkillIDsNotEmbedded/manifest_error (0.00s)
    --- SKIP: TestLegacySkillIDsNotEmbedded/manifest_empty (0.00s)

$ ... | grep -c -- '--- PASS: TestLegacySkillIDsNotEmbedded'
2                                    ← AC-LSL-004(b) EXPECT: 1  → FAILS
$ ... | grep -c -- '--- SKIP'
2                                    ← AC-LSL-007(b) EXPECT: 2  → passes
```

The two ACs only reconcile under one unstated implementation shape: the production assertion must live **directly in the parent function body**, never as its own subtest. Nothing in `spec.md`, `plan.md §A D3`, or `plan.md §E M2` states that. `AC-LSL-005` inherits the same assumption (`grep -c -- '--- FAIL: TestLegacySkillIDsNotEmbedded'` EXPECT 1) — under a subtest shape, falsification produces a parent FAIL *and* a subtest FAIL, i.e. 2.

The result is an AC set that reads as rigorous but will fail the run-phase agent for a structurally correct implementation, or force it to silently discover an unstated constraint.

**Required fix.** Either anchor the greps to the parent line (`grep -cE '^--- PASS: TestLegacySkillIDsNotEmbedded'` / `^--- FAIL:`), or state the required test shape explicitly in `plan.md §A D3` and keep the current counts. Anchoring is the cheaper and more robust choice; it also survives a later refactor into subtests.

---

### M3 — `AC-LSL-014(b)` cannot detect the failure it exists to detect (Epic defect class 3)

**Artifact/location:** `acceptance.md:356-359`.

```bash
go test ./internal/cli/... ./internal/template/... -count=1 2>&1 | grep -c -E '^(FAIL|---) '
# EXPECT: 0
# BASELINE: 0 (all packages "ok"; internal/template/scripts has no test files)
```

**What is wrong.** The pattern requires a literal **space** after `FAIL`. Go emits package failures as `FAIL\t<pkg>\t<time>` (tab) and a bare `FAIL` line with no trailing space — neither matches. Only `--- FAIL: …` lines match, and those exist only for *runtime* test failures. A **compile failure in a test file produces no `--- ` line at all**, so the AC returns 0 and passes.

This SPEC adds two brand-new test files to `internal/cli`. A compile error in either takes down the whole `internal/cli` test binary. Demonstrated on a minimal module:

```
$ go build ./...            ← go build ignores _test.go files
build-exit=0
$ go test ./... -count=1
# vac/pkg [vac/pkg.test]
pkg/a_test.go:9:6: undefined: undefinedSymbol
FAIL	vac/pkg [build failed]
FAIL
$ go test ./... -count=1 2>&1 | grep -c -E '^(FAIL|---) '
0                                    ← AC-LSL-014(b) EXPECT: 0  → PASSES vacuously
```

`AC-LSL-014(a)` does not cover this: `go build ./...` does not compile test files (exit 0 above). The partial safety net is `AC-LSL-004(b)` / `AC-LSL-010(b)`, which grep for `--- PASS:` lines that a broken build cannot produce — so a total escape is unlikely. But AC-LSL-014(b) is the *only* whole-suite regression check across `internal/cli/...` and `internal/template/...`, and as written it certifies nothing about compilation. That is exactly the vacuity class `acceptance.md §A` rule 1 promises to have eliminated, one level up from `-run` selectors.

**Required fix.** Assert on the exit code rather than a grep, e.g.
```bash
go test ./internal/cli/... ./internal/template/... -count=1 > /tmp/lsl-suite.log 2>&1; echo "exit=$?"
# EXPECT: exit=0
# BASELINE: exit=0
```
and keep the grep as a secondary readability check. Record the baseline by executing it (I did: exit 0, all packages `ok`, `internal/template/scripts [no test files]`).

---

### M4 — `acceptance.md §C.1` step 4 uses a destructive revert that can delete the fix it is validating (Epic defect class 2)

**Artifact/location:** `acceptance.md:408-410`.

```bash
# 4. Revert the edit; confirm green again and a clean tree.
git checkout -- internal/cli/update_archive.go
git status --porcelain internal/cli/update_archive.go   # EXPECT: no output
```

**Two problems.**

1. **It can destroy M1 and M4.** `git checkout -- <path>` restores from the index. `internal/cli/update_archive.go` is the single file carrying *both* the M1 list correction and the M4 loop refactor. If M1's correction is uncommitted and unstaged when §C.1 runs — nothing in `plan.md §E` requires an M1 commit before M2's falsification step, and M2 immediately follows M1 — this command silently reverts the file to the 16-entry pre-fix state. The subsequent `git status --porcelain` check then reports "no output" and the procedure declares success, because a clean tree is exactly what a full revert produces. The verification step cannot distinguish "temporary edit removed" from "the whole fix removed."

2. **It is a repo-forbidden command.** `.claude/rules/moai/workflow/main-checkout-branch-guard.md` lists `git checkout -- <path>` in its forbidden table for the primary checkout ("discards work the orchestrator cannot see the provenance of"), and this SPEC's own `progress.md §E.1` records the checkout as shared (`origin/main...HEAD → 0 0`, concurrent-session discipline noted in `§E.2` Epic run order). `§C.2` step 4 and `§C.4` repeat the same "Revert; confirm clean tree" instruction with no command at all, inheriting the hazard implicitly.

**Required fix.** Replace the revert with a non-destructive, fix-preserving form and make the post-condition discriminating. For example:

```bash
# 0. (precondition) M1 is committed; record the SHA.
FIXED=$(git rev-parse HEAD)
# 4. Revert ONLY the falsification edit.
git restore --source="$FIXED" -- internal/cli/update_archive.go
# 4b. Post-condition must prove the FIX is present, not merely that the tree is clean:
awk '/^var legacySkillIDs = \[\]string\{/{f=1;next} f&&/^\}/{f=0} f' \
  internal/cli/update_archive.go | grep -c '"moai-'      # EXPECT: 13
```
Add the same fix-present post-condition to `§C.2` and `§C.3`, and state in `plan.md §E M2` that M1 must be committed before falsification begins (this also gives `progress.md §E.2`'s `pre_fix_commit` field an unambiguous meaning — currently it is described as the *pre*-fix SHA yet is said to "bind §C.1 through §C.4", which are all post-fix procedures).

---

## SHOULD-FIX

### S1 — `AC-LSL-011`'s grep literal will reject a correct assertion

`acceptance.md:295`: `grep -c '"total:"' internal/cli/update_archive_continue_test.go`, EXPECT ≥1. This requires the exact 8-character sequence `"total:"` — a closing quote immediately after the colon. The natural assertion mirroring the existing contract test is `strings.Contains(output, "total: ")` or `"total: 2 skills"`, neither of which matches. (The existing `update_archive_flow_test.go:78` happens to use `"total:"` exactly, which is likely where the literal came from.) Use `grep -c 'total:'` or `grep -cF 'total:'`.

Related: `update_archive_continue_test.go` is named as a hard dependency by AC-LSL-011, AC-LSL-014(c) and AC-LSL-014(e), but `plan.md §E M4` never mandates the filename — it says only "Write the failing test first". If the implementer chooses another name, three ACs fail on `No such file or directory`. Pin the filename in M4.

### S2 — `REQ-LSL-016` has no mechanical AC; three of AC-LSL-010's obligations are "checkable by reading it"

`acceptance.md:276-281` lists four obligations — fail-the-*first*-entry ordering, later-entry-archived assertion, success-only count (`REQ-LSL-016`), portable injection — under "The test's own obligations, **checkable by reading it**". A reviewer-read obligation is not a binary AC. In particular `REQ-LSL-016` ("The reported archived count shall count successful archives only") is verified nowhere mechanically. Add a grep for the count assertion, or an explicitly-named subtest whose `--- PASS` line is asserted.

### S3 — `AC-LSL-009(d)` grep is over-broad and exit-code-fragile

`grep -c 'archive' .gitignore` EXPECT 0 fails on any future unrelated line containing the substring `archive` (e.g. a `*.tar.archive` rule), and `grep -c` exits **1** on zero matches — under `set -e` or a `&&` chain the AC aborts the run rather than passing. Scope the pattern (`grep -cE '^\.?moai/archive'`) and note the exit-1-on-zero-match behaviour, as `§A` rule 3 promises for the other greps.

### S4 — `REQ-LSL-012`'s Where-branch is contradicted by its own AC

`spec.md:193`: "**Where** a `.gitignore` rule for the archive tree is adopted, it shall not un-track the four genuine entries; otherwise the decision shall be recorded as a follow-up in plan.md." `AC-LSL-009(d)` requires zero `archive` matches in `.gitignore` — so the Where-branch can never be satisfied and the requirement's primary clause is dead. `plan.md §A D5` already decided "out of scope". Rewrite REQ-LSL-012 as the prohibition it actually is (e.g. "The SPEC shall not add a `.gitignore` rule for the archive tree; the decision shall be recorded as a follow-up in plan.md"), which also removes the compound two-subject shape flagged in S5.

### S5 — Two REQs deviate from GEARS

`REQ-LSL-008` ("The guard shall be demonstrated to fail against the pre-correction 16-entry list") is a verification-process requirement whose real actor is the implementer, not the system. `REQ-LSL-012` chains two requirements with different subjects across an "otherwise". Both are minor (15/17 are clean) and MP-2 passes, but S4's rewrite fixes REQ-LSL-012 for free.

### S6 — Frontmatter style regresses against the Epic's own remediated siblings

`spec.md` carries `version: 0.1.0` (unquoted; the schema specifies a quoted semver string), `priority: high` (lowercase; the schema enum is `P0|P1|P2|P3` or `High|Medium|Low|Critical`), `phase: plan` (the schema describes `phase` as a release target). The enforced lint accepts all three (`0 error(s)`, no finding for this SPEC), so this is style, not validity. But it is style the Epic already corrected:

```
SPEC-UPDATE-DATA-SURVIVAL-001   version: "0.2.0"  priority: P0  phase: "v3.0.2"
SPEC-CONFIG-TIER-PERSIST-001    version: "0.2.0"  priority: P0  phase: "v3.0.2"
SPEC-CONFIG-KEY-HONESTY-001     version: "0.2.0"  priority: P1  phase: "v3.0.2"
SPEC-UPDATE-CI-GUARD-001        version: "0.2.0"  priority: P1  phase: "v3.0.2"
SPEC-UPDATE-DOC-DRIFT-001       version: "0.2.0"  priority: P2  phase: "v3.0.2"
SPEC-UPDATE-REINSTALL-LOOP-002  version: 0.2.0    priority: high phase: plan       ← also unremediated
```
Five of six siblings converged on the canonical form during audit remediation. Adopting it here costs three characters and removes a known future audit finding.

### S7 — `progress.md` states one measured triple in the wrong order

`progress.md:17`: "the brief cited backend 46 / database 25 / frontend 17. Measured file counts are 56 / 30 / 26 (occurrence counts **68 / 34 / 30**)."

The file counts are in backend/database/frontend order and are correct (`56/30/26`). The occurrence counts in that same order are `68/30/34` — the line has frontend and database transposed. `spec.md §A` (backend/frontend/database order: `56/26/30`, `68/34/30`) and `plan.md §C` (backend/database/frontend: `56/30/26`, `68/30/34`) are each internally consistent and correct; only `progress.md` mixes the two orderings. Verified against the re-measured values above.

### S8 — `archiveVersion` invariance is a stated goal with no AC

`spec.md §B` Goal 5 and `NFR-LSL-003` both require `archiveVersion` and the archive directory scheme to be unchanged, and `§C` names them out of scope. No AC pins the constant. `AC-LSL-008`/`009` exercise `.moai/archive/skills/v2.16/` as *git paths*, which would not catch a change to the Go constant. Add `grep -c 'archiveVersion = "v2.16"' internal/cli/update_archive.go` → EXPECT 1 (baseline verified: the constant is at `update_archive.go:29`).

### S9 — M4 does not specify per-entry `continue` semantics

`plan.md §E` M4 step 2 says only "Convert all three in-loop `return archived, …` sites to accumulate." It does not say what happens after an accumulation — specifically whether control `continue`s to the next entry or falls through into the remaining per-entry work. `REQ-LSL-013` implies `continue`, but the drift-backup sites at `:302`/`:305` sit *above* `archiveSkill` at `:319`, so a literal `return → append` substitution without a `continue` would fall through into `os.Rename` and `archiveSkill` for an entry whose backup parent could not be created.

I checked the blast radius: the existing `TestArchiveForce/force_with_drift_backup_failure_preserves_original` (`update_archive_force_test.go:172-218`) asserts `strings.Contains(err.Error(), "drift backup")` and that the original archive still holds the stale content — both survive either control flow, because the subsequent `os.Rename` also fails. So this is latent, not immediately breaking. State the `continue` explicitly anyway; the AC set does not pin it either.

### S10 — `AC-LSL-006` has no owning REQ

`AC-LSL-006` traces to "(M1 step 3-4)", a plan milestone. Every other AC names REQ or NFR IDs. Either add a REQ for "the six pre-existing test files pass unmodified in behaviour + the `All16` naming decision is recorded" (it is already `spec.md §F` Success Criterion 3 and `§G` Risk 2), or cite those explicitly.

### S11 — the "7 of 7" denominator is not independently confirmable

`spec.md:26` records "Epic SPEC 7 of 7 (added after the initial six)". The six siblings each read "Epic SPEC N of 6", exactly as `plan.md §A D1` describes, so the documented inconsistency is real and deliberate (D1's *decision* is fine; only its rationale is stale — see M1). However the working tree also contains `SPEC-UPDATE-YAML-PRESERVE-001`: `created: 2026-07-31`, `author: manager-spec`, `module: cli`, `era: V3R6`, `tier: M`, `priority: high`, `phase: plan`, subject `moai update` config merge — the same day, author, module and frontmatter style as this SPEC. It carries no "Epic SPEC N of M" line and its `related_specs` point elsewhere (it is issue-driven, #1243), so it is plausibly not an Epic member. I cannot verify either way — there is no authoritative Epic roster artifact. **Flagged as unverified, not as a defect.** If YAML-PRESERVE is an Epic member, "of 7" is already wrong.

---

## Explicitly Checked and Found Clean

These were hunted for and **not** found — recorded so a later reader does not re-litigate them.

1. **Defect class 4 (defect-locus mis-attribution): zero findings.** All 13 cited locations resolve; six are exact to the line pair. See the table above.
2. **Scope discipline: clean.** No AC or milestone reaches into any of the five declared out-of-scope items. `restoreSkill` is referenced only as evidence; no AC touches `migrate_restore_skill.go`. `AC-LSL-001(b)` pins the 13 retained IDs verbatim **and in order**, mechanically guarding "no change to the other 13 entries". `AC-LSL-009(d)` is a negative AC enforcing the `.gitignore` out-of-scope decision. Nothing addresses downstream user archives or `moai-meta-harness`. (The one gap is `archiveVersion` — S8.)
3. **Positional-index safety: the claim holds, and identity-independence is real.** Re-derived all 7 sites: `[0]`,`[1]`,`[2]`,`[3]` (force), `[:5]` (flow), `[:8]` (idempotency), `[0]` (skip_sync) — max index 3, max bound 8, both in range at 13. On the semantics question: post-shrink `[0..3]` selects `db-docs, mobile, electron, shadcn` instead of `backend, frontend, database, db-docs`, and `[:5]`/`[:8]` shift correspondingly. I read every consuming test: each calls `makeSkillDir(t, root, id, …)` to seed a synthetic directory from whatever ID it drew, and asserts only on that same `id` — no test depends on the identity. `spec.md §G` Risk 1 states this argument correctly ("the tests seed synthetic dirs from whatever ID they draw"); `plan.md §C` and `progress.md` cover only the range half. The SPEC's conclusion is sound. Also confirmed: `TestArchiveSkill_All16Skills` and `TestRestoreSkill_All16RoundTrip` iterate the slice dynamically (16 subtests today, 13 after) and assert no cardinality, so the rename in M1 step 4 is genuinely optional.
4. **Output-contract preservation: pinned, not assumed.** `AC-LSL-013` greps the production source (`"archive: "` → 1, `total: %d skills archived` → 2, both baselines reproduced), and `AC-LSL-006(a)` runs the test that actually asserts the runtime output. I verified that test exists and asserts what the SPEC claims: `update_archive_flow_test.go:71` `expected := "archive: " + id`, `:78` `strings.Contains(output, "total:")`. The SPEC's own note ("the greps above are the cheap early-warning, not the authority") is accurate.
5. **Degradation-path reachability (`REQ-LSL-007`): reachable as designed.** `plan.md §A D3` correctly identifies that the real embedded FS never fails in a compiled binary and factors the comparison into a pure function driven by synthetic `(legacyIDs, embedded, err)` inputs. `§C.3`'s falsification (flip `t.Skip` → silent return, observe `--- SKIP` become `--- PASS`) is genuinely discriminating — my subtest probe above confirms the SKIP marker appears exactly as the AC expects. One wording tension: D3 describes a *pure function*, while AC-LSL-007 requires `--- SKIP` markers, which only the subtest body can emit; the two reconcile if the subtest calls `t.Skip` on the function's verdict, but D3 does not say so. Minor clarity, folded into the Clarity score rather than raised as a finding.
6. **Guard falsification (`§C.1` steps 1-3) is operative.** Re-adding `moai-domain-backend` to a corrected list would put it back in the intersection with `EmbeddedMoaiSkillNames()` (which returns it — the template directory exists, confirmed), so the guard genuinely goes red and the message genuinely names the ID. Only step 4's revert is defective (M4).
7. **MP-7 / clarification gate: genuinely clean**, not merely asserted. `plan.md` has zero markers, and every decision `progress.md:60` claims was resolvable is in fact recorded in `plan.md §A` D1-D5 with a rejected alternative.
8. **Tier M is supported.** Files touched: `update_archive.go` + 2 new test files + 3 archive deletions = 6 (> the Tier S `<5 files` bound). LOC is borderline S (~200-300 including tests). Tier M is the conservative and correct call, and the delivered 3-artifact set matches. No finding.
9. **The core defect claim itself is fully substantiated.** The three IDs are on the list (`3`), exist as live template skills (all three `SKILL.md` present), are registered in `catalog.yaml`, are the most-referenced skills in the repo (56/26/30 files), are re-deployed every update by the `moai*` glob (`deploy.go:51-55`, `isGlob: true`), are archived *after* redeploy (`update.go:454`, inside `Post-sync steps`), and their live/archive hashes differ (all six md5s). The loop is real and permanent. This SPEC is solving a genuine defect.

---

## Recommendation

Fix M1-M4, then re-audit **scoped to this defect delta** — a from-scratch re-audit is unnecessary. Every one of the four is a localized edit:

1. **M1** — rewrite `plan.md:13` D1's rationale on grounds that survive `9a6b6c854`. The second half of D1 ("zero functional gain … belongs to whichever change lands last") already stands on its own; deleting the branch-collision clause is sufficient.
2. **M2** — anchor the two greps: `acceptance.md:110` → `grep -cE '^--- PASS: TestLegacySkillIDsNotEmbedded'`, `acceptance.md:141` → `grep -cE '^--- FAIL: TestLegacySkillIDsNotEmbedded'`. Keep both EXPECT values at 1.
3. **M3** — replace `acceptance.md:357`'s grep with an exit-code assertion; the baseline is `exit=0` (measured).
4. **M4** — replace `acceptance.md:409`'s `git checkout --` with a `git restore --source=<M1-commit>` form, add a fix-present post-condition (`… | grep -c '"moai-'` → 13) to `§C.1`, `§C.2` and `§C.4`, and state in `plan.md §E` M2 that M1 must be committed before falsification begins.

S1-S10 are all one-line edits and should be folded into the same pass; S4 and S8 are the two worth prioritizing (a dead requirement clause, and a stated goal with no AC). S11 needs a decision, not an edit: confirm whether `SPEC-UPDATE-YAML-PRESERVE-001` belongs to this Epic before the "of 7" denominator is committed.

A note on the aggregate. 0.79 against a 0.80 threshold is a hair's breadth, and the temptation to round up is real — this SPEC's baseline hygiene and citation accuracy are genuinely the strongest in the Epic. I am declining to round up because the four MUST-FIX items are not cosmetic: M2 and M3 are ACs that will mislead the run-phase verdict (one blocks a correct implementation, one certifies nothing), M4 can destroy the implementation it is meant to validate, and M1 is the exact defect class this Epic's six prior FAILs were about. The SPEC promises in `acceptance.md §A` that "an AC whose baseline was not observed is not an AC" — it earned that claim on the baselines and lost it on the mechanics.

---

# Iteration 2

Iteration: 2 (of max 3)
Auditor: plan-auditor (adversarial, M1-M5 bias-prevention protocol active)
Audited at: 2026-07-31, working tree `main`, SPEC at `version: "0.2.0"`
Artifacts read: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M input contract).

**Reasoning context ignored per M1 Context Isolation.** manager-spec's claim that all 4 MUST-FIX and all 11 SHOULD-FIX are resolved was treated as a hypothesis. Every closure below was re-derived by executing the relevant command; every new artifact (AC-LSL-015, AC-LSL-016, REQ-LSL-018, §C.0-§C.4) was audited from scratch.

**Environment-trap discipline applied.** `ls | grep '^name'` was not used anywhere in this audit; file existence was established by glob and by `git ls-files`. The six sibling iteration-2 reports were located by glob and all six verdicts read directly — confirming the iteration-1 finding and the SPEC's own §A measurement note.

---

## Verdict

| | |
|---|---|
| **Verdict** | **FAIL** |
| **Overall Score** | **0.81** (harmonic mean of 4 dimensions; iteration 1: 0.79) |
| **Tier** | M — threshold **0.80** |
| **Must-pass firewall** | no MP failure (MP-1…MP-7 all PASS or N/A) |
| **Score trend** | 0.79 → 0.81, improving — no STOP escalation, no scope-reduction proposal |

**The aggregate clears the Tier M threshold. The FAIL is not driven by the aggregate.** It is driven by a proven contradiction inside the acceptance surface: `AC-LSL-010(b)` and the newly-added `AC-LSL-016(b)` are mutually unsatisfiable (measured, below), so `§D` Definition of Done — "AC-LSL-001 … AC-LSL-016 all pass" — cannot be achieved by any implementation. I decline to certify an acceptance surface that a structurally correct implementation must fail. This is the identical defect class as iteration-1 **M2**: fixed for the guard test's ACs, and re-created for the archive-loop test's ACs by the very AC that resolved S2.

v0.2.0 is materially better than v0.1.0. The falsification machinery is now genuinely operative — I executed all four §C procedures end-to-end on this platform and every one behaves as documented, which is rare. Three MUST-FIX items remain, all small edits, and a delta-scoped re-audit of those three should convert this to PASS.

---

## Must-Pass Results

- **[PASS] MP-1 — REQ number consistency.** `REQ-LSL-001 … REQ-LSL-018` sequential, no gaps, no duplicates, 3-digit padding. 18 bullet definitions for 18 distinct IDs. `NFR-LSL-001…005` and `AC-LSL-001…016` likewise.
  ```
  $ grep -o 'REQ-LSL-[0-9]*' spec.md | sort -u | wc -l → 18
  $ grep -c '^- \*\*REQ-LSL-' spec.md               → 18
  $ grep -c '^### AC-LSL-' acceptance.md             → 16
  ```
- **[PASS] MP-2 — GEARS format compliance.** `REQ-LSL-008` is now event-driven (`When the guard is evaluated against the pre-correction 16-entry list, the guard shall fail…`) — S5 closed. `REQ-LSL-012` is now a single-subject Unwanted-form prohibition (`The SPEC shall not add a .gitignore rule…`) — S4 closed. `REQ-LSL-018` is ubiquitous form with a compound "and" clause (minor, recorded as SHOULD-FIX-h).
- **[PASS] MP-3 — YAML frontmatter validity.** All 12 canonical fields present with correct types (`spec.md:2-15`); no rejected snake_case alias. S6 closed: `version: "0.2.0"` (quoted), `priority: P1`, `phase: "v3.0.2"` — now matching the Epic's remediated siblings. Optional `era: V3R6`, `tier: M`, `issue_number: null`, `related_specs` present and well-formed.
  *Gap:* `moai spec lint` did **not** complete — it exceeded a 200 s timeout (`exit=124`, no output). The progress.md claim `0 error(s), 62 warning(s)` and "no finding for this SPEC" could not be independently reproduced this iteration. MP-3 is passed on a field-by-field schema check against `spec-frontmatter-schema.md`, not on the lint run. Recorded as a gap, not as a pass-by-inference.
- **[N/A] MP-4 — Section 22 language neutrality.** Single-language (Go) internal-CLI SPEC; `NFR-LSL-002` forbids touching `internal/template/templates/`. Auto-passes.
- **[PASS] MP-5 — D7 cross-SPEC reconciliation.** All 8 referenced SPEC IDs resolve; none is retired/superseded/archived. `SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001` (`implemented`) is explicitly reconciled in spec.md § "Relationship to …". No BLOCKING finding.
- **[PASS] MP-6 — D8 cross-platform discipline.** `grep -c 'syscall'` → 0 in all four artifacts. Auto-PASS.
- **[PASS] MP-7 — clarification gate.** `grep -rn '\[NEEDS CLARIFICATION' plan.md` → no match. `research.md` correctly absent (Tier M). `progress.md:151-155` declares none open and names the resolution of each decision.

---

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|-----------|------:|-------------|----------|
| Clarity | 0.85 | 0.75–1.0 | Every §A decision still carries its rejected alternative; §C.0 now explains *why* the overlay mechanism was chosen over the auditor's own suggestion, with the reasoning stated rather than asserted. Deduction: `plan.md §A D1` leg (ii) still presents an unverified consequence in verified register (**MF-2**). |
| Completeness | 0.90 | 0.75–1.0 | All sections present; 6 × `### Out of Scope — <topic>` H3 with `-` bullets. The three iteration-1 completeness gaps are closed: `archiveVersion` → AC-LSL-015, `REQ-LSL-016` → AC-LSL-016, AC-LSL-006 orphan → REQ-LSL-018. M4 `continue` semantics now stated with rationale. |
| Testability | 0.65 | 0.50–0.75 | One proven contradiction (**MF-1**), one AC with an unobserved and prose-wrong baseline (**MF-3**), three ACs resting on reviewer judgment (AC-015(b), AC-016(a), AC-004(c)-negative), one AC missing the existence guard its own §A rule 1 mandates (AC-LSL-006). Offsetting: AC-LSL-014(b) now genuinely discriminates, and all four §C procedures were executed and confirmed operative. |
| Traceability | 0.90 | 0.75–1.0 | 18/18 REQ and 5/5 NFR literally cited in acceptance.md; zero uncovered; zero orphan ACs. One wrong cross-reference: §C.1's heading binds `AC-LSL-008` where `AC-LSL-004` is meant. |

Harmonic mean = 4 / (1/0.85 + 1/0.90 + 1/0.65 + 1/0.90) = 4 / 4.9372 = **0.810**.

---

## Iteration-1 closure verification

Each of the four MUST-FIX and eleven SHOULD-FIX items was re-checked against the artifacts and, where a mechanism was claimed, against the machine.

| Item | Status | Basis |
|------|--------|-------|
| **M1** — D1 rationale refuted | **PARTIALLY CLOSED** | The false branch-collision clause is gone (`grep -c 'in flight' plan.md` → 0). Legs (i) and (iii) are re-verified TRUE. Leg (ii)'s consequence claim is newly false — see **MF-2**. |
| **M2** — AC-004(b)/005/007 unsatisfiable | **CLOSED for the guard ACs; class RECURRED elsewhere** | Anchoring verified by execution (below). The identical trap now exists between AC-010(b) and AC-016(b) — see **MF-1**. |
| **M3** — AC-014(b) blind to build failure | **CLOSED (genuinely)** | Constructed the failure and ran the new commands — they detect it. |
| **M4** — destructive falsification | **CLOSED (genuinely)** | Overlay procedure executed end-to-end; §C.4 index round-trip executed in a throwaway repo. Both behave exactly as documented. |
| S1 (`"total:"` literal; filenames) | CLOSED | `acceptance.md:317` now `grep -cF 'total:'`; both filenames pinned in `plan.md §E M2` step 1 and `§E M4` step 1. |
| S2 (REQ-016 no mechanical AC) | CLOSED (with caveat) | AC-LSL-016 added; part (b) is mechanical and works, part (a) is judgment-based (SHOULD-FIX-e). |
| S3 (`.gitignore` grep over-broad) | CLOSED | `^\.?moai/archive` scoping + count captured into `n`; measured 0. |
| S4 (dead Where-branch) | CLOSED | REQ-LSL-012 is now a plain prohibition. |
| S5 (GEARS deviations) | CLOSED | REQ-LSL-008 restated event-driven; REQ-LSL-012 single-subject. |
| S6 (frontmatter style) | CLOSED | Verified field-by-field. |
| S7 (transposed triple) | CLOSED | `progress.md:17` now states the ordering explicitly and records the correction. |
| S8 (`archiveVersion` no AC) | CLOSED for (a), DEFECTIVE for (b) | AC-015(a) measured `1`. AC-015(b) — see **MF-3**. |
| S9 (`continue` semantics) | CLOSED | `plan.md:200` states it, with the fall-through hazard named. |
| S10 (AC-006 orphan) | CLOSED | REQ-LSL-018 added; AC-006 heading now cites it. |
| S11 (roster unverifiable) | CLOSED | Independently re-verified — see below. |

**Nothing was found merely reworded.** Every claimed closure changed the artifact in the way described; the two problems are a new false premise (MF-2) and a newly-introduced contradiction (MF-1), not cosmetic rewrites of old ones.

### Closure evidence — the four MUST-FIX items

**M2 anchoring (closed).** Built the exact shape `plan.md §E M2` pins (production assertion in the parent body; `manifest_error` / `manifest_empty` as subtests) and ran it:
```
--- PASS: TestLegacySkillIDsNotEmbedded (0.00s)
    --- SKIP: TestLegacySkillIDsNotEmbedded/manifest_error (0.00s)
    --- SKIP: TestLegacySkillIDsNotEmbedded/manifest_empty (0.00s)

grep -cE '^--- PASS: TestLegacySkillIDsNotEmbedded'  → 1   ← AC-LSL-004(b) EXPECT 1 ✓
grep -c -- '--- SKIP'                                → 2   ← AC-LSL-007(b) EXPECT 2 ✓
```
FAIL direction, under overlay with `moai-domain-backend` re-injected:
```
--- FAIL: TestLegacySkillIDsNotEmbedded (0.00s)
    --- SKIP: TestLegacySkillIDsNotEmbedded/manifest_error (0.00s)
    --- SKIP: TestLegacySkillIDsNotEmbedded/manifest_empty (0.00s)

grep -cE '^--- FAIL: TestLegacySkillIDsNotEmbedded' → 1   ← AC-LSL-005 EXPECT 1 ✓
grep -c 'moai-domain-backend' <log>                 → 1   ← AC-LSL-005 EXPECT >=1 ✓
```
AC-004(b), AC-005 and AC-007(b) are mutually consistent under the pinned shape. Closed.

**M3 build-failure detection (closed).** Constructed a module with one `_test.go` compile error and ran the exact replacement commands:
```
$ go build ./...                                   → build_exit=0   (still blind, as documented)
$ go test ./pkg/... -count=1 > /tmp/log 2>&1       → exit=1
$ grep -c 'build failed' /tmp/log                  → 1
$ grep -c -- '--- FAIL' /tmp/log                   → 0
$ cat /tmp/log
  # probe/pkg [probe/pkg.test]
  pkg/broken_test.go:5:33: undefined: undefinedSymbol
  FAIL	probe/pkg [build failed]
  FAIL
```
Both the exit code and the `build failed` marker fire; the retired `^(FAIL|---) ` form would still have returned 0. The AC now detects what it exists to detect. Baseline re-measured on the real tree:
```
$ go test ./internal/cli/... ./internal/template/... -count=1 > /tmp/lsl-suite.log 2>&1; echo "exit=$?"
exit=0
$ grep -c 'build failed' /tmp/lsl-suite.log → 0
$ grep -c -- '--- FAIL' /tmp/lsl-suite.log  → 0
```
Matches the recorded baseline exactly. Closed.

**M4 non-destructive falsification (closed).** The overlay mechanism was executed end-to-end, including the exact `sed` and `printf` forms in §C.0/§C.1:
```
$ sed 's|^var legacySkillIDs = \[\]string{|&\n\t"moai-domain-backend",|' "$FIX_SRC" > "$WORK/mutated.go"
$ wc -l < "$WORK/mutated.go"      → 6   (was 5 — the newline WAS interpreted)
$ printf '{"Replace": {"%s": "%s"}}\n' "$(cd "$(dirname "$FIX_SRC")" && pwd)/$(basename "$FIX_SRC")" "$WORK/mutated.go"
{"Replace": {"/…/probe/pkg/list.go": "/var/folders/…/mutated.go"}}
$ go test -overlay="$OVERLAY_JSON" … → --- FAIL, names moai-domain-backend
$ grep -c 'moai-domain-backend' pkg/list.go → 0     ← P3 holds; tree provably untouched
```
Two things worth recording because they are non-obvious and could have broken the procedure: (a) the `\n`/`\t` escapes in the replacement **are** honoured by this platform's `sed` — the mutation produces valid Go rather than a one-line compile error; (b) `-overlay` also works when the replaced file is a `_test.go`, which §C.3 depends on — verified: flipping `t.Skip` to a silent return under overlay turned `--- SKIP` into `--- PASS` while `grep -c 'manifest empty'` on the real file still returned 1. Both §C.3's mechanism and its discriminating property hold.

§C.4's index-only round trip was executed in a throwaway git repo:
```
before: 1
after rm --cached: 0
after reset: 1
present
porcelain after: (empty)
```
`git rm --cached` un-tracks without touching the working file; `git reset -- <path>` restores the index entry; the file survives and the tree ends clean. Also checked that neither command is on the primary-checkout guard's pattern list — `internal/hook/branch_guard.go:96` matches only `\bgit\s+reset\s+--hard\b`, and `git rm --cached` has no pattern at all (the guard is additionally default-OFF per `BranchGuardConfig.Enabled`). §C.4 is correct and safe. Closed.

**S11 roster (closed, and the answer is confirmed).**
```
SPEC-UPDATE-REINSTALL-LOOP-002     Epic SPEC 1 of 6
SPEC-UPDATE-DATA-SURVIVAL-001      Epic SPEC 2 of 6
SPEC-CONFIG-TIER-PERSIST-001       Epic SPEC 3 of 6
SPEC-CONFIG-KEY-HONESTY-001        Epic SPEC 4 of 6
SPEC-UPDATE-CI-GUARD-001           Epic SPEC 5 of 6
SPEC-UPDATE-DOC-DRIFT-001          Epic SPEC 6 of 6
SPEC-UPDATE-YAML-PRESERVE-001      (none)
```
and the two membership citations resolve verbatim:
```
$ sed -n '40p' SPEC-UPDATE-DATA-SURVIVAL-001/progress.md
| 3+ | remaining Epic SPECs (`SPEC-UPDATE-YAML-PRESERVE-001`, `SPEC-CONFIG-TIER-PERSIST-001`, …
$ sed -n '121p' SPEC-UPDATE-DOC-DRIFT-001/progress.md
| 5+ | `SPEC-UPDATE-CI-GUARD-001` (E5), `SPEC-UPDATE-YAML-PRESERVE-001`, **this SPEC** (E6) | …
```
The roster is 7 with 6 numbered; leg (i) of D1 is TRUE. **The ordinal decision creates no contradiction with any sibling artifact** — this SPEC asserts no ordinal for itself and edits no sibling, so the siblings remain internally consistent and this SPEC makes no claim they refute. Confirmed clean.

---

## MUST-FIX

### MF-1 — `AC-LSL-010(b)` and the new `AC-LSL-016(b)` are mutually unsatisfiable (iteration-1 M2's defect class, re-created)

**Artifact/location:** `acceptance.md:286-289` (AC-LSL-010 b) and `acceptance.md:455-459` (AC-LSL-016 b), interacting via `acceptance.md:462`.

**What is wrong.** AC-LSL-016 **mandates** a named subtest: "Part (b) requires the success-count check to live in a **named subtest** (`success_count_excludes_failures`)". AC-LSL-010(b) then counts PASS lines for the same test with an **unanchored** grep:

```bash
go test ./internal/cli/ -run 'TestArchiveLegacySkills_ContinuesAfterFailure' -count=1 -v 2>&1 \
  | grep -c -- '--- PASS: TestArchiveLegacySkills_ContinuesAfterFailure'
# EXPECT: 1
```

Go prints the parent PASS line *and* an indented PASS line for the mandated subtest, so the unanchored grep counts 2. Measured on the exact shape both ACs together require:

```
$ go test ./pkg/ -run 'TestArchiveLegacySkills_ContinuesAfterFailure' -count=1 -v
--- PASS: TestArchiveLegacySkills_ContinuesAfterFailure (0.00s)
    --- PASS: TestArchiveLegacySkills_ContinuesAfterFailure/success_count_excludes_failures (0.00s)

$ ... | grep -c -- '--- PASS: TestArchiveLegacySkills_ContinuesAfterFailure'
2                              ← AC-LSL-010(b) EXPECT: 1  → FAILS
$ ... | grep -cE '^    --- PASS: TestArchiveLegacySkills_ContinuesAfterFailure/success_count_excludes_failures'
1                              ← AC-LSL-016(b) EXPECT: 1  → passes
```

A correct implementation therefore fails AC-LSL-010(b), and `§D` Definition of Done ("AC-LSL-001 … AC-LSL-016 all pass") is unachievable. The run-phase agent must either fail a green implementation or silently discover that AC-016's mandated subtest must not exist — in which case AC-016 fails instead.

This is precisely the trap iteration 1 raised as M2, closed for the guard test, and re-opened for the archive-loop test by the AC that resolved S2. The SPEC's own `§A` rule 3 states the remedy it did not apply here: "Where a test-output line is counted, the pattern is `^`-anchored — Go indents subtest result lines, so an unanchored `--- PASS:` grep silently counts parent plus subtests."

**Required fix.** Anchor AC-LSL-010(b) the same way AC-LSL-004(b) is anchored:
```bash
  | grep -cE '^--- PASS: TestArchiveLegacySkills_ContinuesAfterFailure'
# EXPECT: 1
```
Measured under the mandated shape: `1`. Then re-check every remaining unanchored `--- PASS:` / `--- FAIL:` count in acceptance.md for the same hazard — the EXPECT-0 forms (AC-011, AC-012(c), AC-006(a)) are unaffected because 0 is 0 either way, but a future subtest addition would not change that, so anchoring them costs nothing.

---

### MF-2 — `plan.md §A D1` leg (ii): the cited mechanism is real, but the consequence it is used to assert is false (Epic defect class 1, recurring)

**Artifact/location:** `plan.md:38-50`, restated in `progress.md:69`.

**What the SPEC claims.** That renumbering the six siblings carries "a real, mechanical cost": editing a HISTORY denominator changes `spec.md`, "which changes that hash, which fails skip-eligibility condition 3 (artifact-hash unchanged) and forces a full Phase 1 plan-audit re-execution on each of the six", i.e. it "would invalidate six passing plan-audit verdicts".

**The mechanism half is TRUE and was verified against source.** `planArtifactNames` exists and hashes exactly the four files named:
```
$ grep -n "planArtifactNames" -A 6 internal/runtime/audit_cache.go
61:// planArtifactNames is the ordered list of plan artifact file names to hash.
63:var planArtifactNames = []string{
64-	"acceptance.md",
65-	"plan.md",
66-	"spec.md",
67-	"tasks.md",
68-}
```
and `ComputeHash` reads each and folds it into a whitespace-normalized SHA-256 (`audit_cache.go:90-105`). The citation to `spec-workflow.md` § Report Persistence is accurate. No defect-locus mis-attribution here.

**The consequence half is FALSE, on two independent grounds.**

1. **No skip-eligibility exists to lose.** The skip contract (`spec-workflow.md` § Phase Transitions, Plan Audit Gate skip policy) requires **all four** conditions, of which condition 2 is **overall score ≥ 0.90**. The six siblings' iteration-2 scores, read directly from the report files this SPEC itself cites:
   ```
   $ for f in .moai/reports/plan-audit/*epic-update-config-iter2.md; do grep -m2 -iE 'verdict|overall score' "$f"; done
   SPEC-CONFIG-KEY-HONESTY-001     PASS   Overall Score: **0.81**
   SPEC-CONFIG-TIER-PERSIST-001    PASS   Overall Score: **0.82**
   SPEC-UPDATE-CI-GUARD-001        PASS   Overall Score: **0.85**
   SPEC-UPDATE-DATA-SURVIVAL-001   PASS   Overall Score: 0.84
   SPEC-UPDATE-DOC-DRIFT-001       PASS   Overall Score: **0.82**
   SPEC-UPDATE-REINSTALL-LOOP-002  PASS   Overall Score: **0.88**
   ```
   All six are in `0.81 … 0.88`; **none reaches 0.90**. Every one of the six already fails skip-eligibility on condition 2 and will re-execute Phase 1 on its next `/moai run` regardless of whether its `spec.md` is ever touched. Condition 3 is moot. The claimed mechanical cost is **zero**.

2. **There is no persistent "cached PASS" to invalidate.** The Go cache the SPEC invokes is explicitly non-durable:
   ```
   $ sed -n '70,75p' internal/runtime/audit_cache.go
   // InMemoryCache implements AuditCache using an in-memory store.
   // Each GateConfig holds its own InMemoryCache instance, so cache entries do not
   // persist across separate /moai run invocations.
   ```
   A hash change cannot invalidate an entry that does not survive the process. (The durable surface is the daily report file, which is a verdict record, not a hash-keyed cache — `spec-workflow.md` § Report Persistence states this explicitly.)

**Why this matters.** Iteration 1's M1 was: "D1's rationale is stated in verified register but was never re-run." v0.2.0 removed the false branch-collision premise and replaced it with a *differently* false premise — a mechanism citation that is correct at the file-and-function level but whose stated consequence was never checked against the six scores sitting in the very report files the same paragraph cites. `plan.md §C` opens with "Every `file:line`, count, and command in the delegation brief was re-run before being written into an artifact"; this consequence was not. That is Epic defect class 1 recurring in a subtler form: verified locus, unverified implication.

**What survives.** The **decision** (no ordinal; no sibling renumbering) is sound and does not depend on this leg — leg (i), the roster inconsistency, is independently verified TRUE above and is sufficient on its own.

**Required fix.** Delete leg (ii) or restate it truthfully. A correct version would say something like: "the six siblings' iteration-2 scores are 0.81–0.88, all below the 0.90 skip-eligibility floor, so each already re-executes Phase 1 on its next run — editing their `spec.md` costs nothing mechanically, and the case against renumbering rests entirely on leg (i)." Do not leave a stated mechanical cost that measurement contradicts.

---

### MF-3 — `AC-LSL-015(b)` is non-binary and its prose baseline undercounts the measured value (new material, never audited)

**Artifact/location:** `acceptance.md:431-435`.

```bash
grep -cE '"\.moai", "archive", "skills"' internal/cli/update_archive.go
# EXPECT: unchanged from baseline
# BASELINE: the joins in archiveSkill / archiveLegacySkills / the force path
```

**Two problems.**

1. **"unchanged from baseline" is not a criterion.** No number is recorded, so a run-phase agent cannot evaluate it without re-deriving the pre-fix value from a tree that no longer exists. This is exactly what `acceptance.md §A` rule 4 forbids ("Baselines are stated, not assumed") and what §A's opening sentence claims to have eliminated ("An AC whose baseline was not observed is not an AC").

2. **The prose baseline is wrong.** It names three regions; the measured count is **six** matching lines across at least four regions, including `dryRunArchiveLegacySkills` which the prose does not mention:
   ```
   $ grep -cE '"\.moai", "archive", "skills"' internal/cli/update_archive.go
   6
   $ grep -nE '"\.moai".*archive' internal/cli/update_archive.go
   73:	dstDir := filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion, skillID)
   287:		dstDir := filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion, id)
   298:				backupDir := filepath.Join(projectRoot, ".moai", "archive", "skills",
   311:				archiveBackupRel := filepath.Join(".moai", "archive", "skills",
   328:		archiveDst := filepath.Join(".moai", "archive", "skills", archiveVersion, id)
   347:		archiveDst := filepath.Join(".moai", "archive", "skills", archiveVersion, id)
   ```

**Required fix.** `# EXPECT: 6` / `# BASELINE (2026-07-31): 6 (lines 73, 287, 298, 311, 328, 347 — archiveSkill, archiveLegacySkills, the two force-path joins, and dryRunArchiveLegacySkills)`. AC-LSL-015(a) is fine as written (measured `1`, EXPECT 1) and is the load-bearing pin; (b) is the supplementary half, so this is the least severe of the three MUST-FIX items — but it is the SPEC's own stated rule being broken in brand-new material.

---

## SHOULD-FIX

**SF-a — `§C.1`'s heading binds the wrong AC.** `acceptance.md:524`: "### C.1 Guard falsification (binds AC-LSL-005, **AC-LSL-008** → REQ-LSL-008)". AC-LSL-008 is the archive-file-untracked criterion and has nothing to do with guard falsification; §C.4 already binds it correctly. Almost certainly `AC-LSL-004` was meant. One-token fix.

**SF-b — `AC-LSL-006` omits the existence guard its own §A rule 1 mandates.** AC-004(a) and AC-010(a) each grep that the test name exists before running `-run`; AC-006(a) does not, and asserts only `grep -c -- '--- FAIL'` → 0. If the six-alternative `-run` regex ever selects nothing (e.g. a rename that drops the `TestArchiveSkill_` / `TestRestoreSkill_` prefixes — which AC-006(c) explicitly permits), the command exits 0 with zero FAIL lines and the AC passes vacuously. Verified the regex currently selects real tests (19 top-level PASS/FAIL lines, 0 FAIL, `ok … 0.660s`), so this is latent, not live. Add a PASS-line count or a name-existence grep.

**SF-c — `AC-LSL-014(b)`'s two greps exit 1 on the expected zero-match.** The same `grep -c` exit-code hazard that AC-009(d) documents in a NOTE applies verbatim here (`grep -c 'build failed'` → prints `0`, exits `1`; observed). Under `set -e` or a `&&` chain the AC aborts the run instead of passing. Apply AC-009(d)'s capture-then-compare form, or repeat the NOTE.

**SF-d — `progress.md §E.1` "Epic run order": the disjointness claim is not established.** It states "no sibling touches `internal/cli/update_archive.go`. It may therefore run in parallel". Measured:
```
$ grep -rl 'update_archive' <the seven sibling dirs>
SPEC-UPDATE-REINSTALL-LOOP-002/acceptance.md   (update_archive.go:351)
SPEC-UPDATE-DOC-DRIFT-001/spec.md, plan.md     (update_archive.go:339-353 — "M1 evidence sites")
SPEC-UPDATE-DATA-SURVIVAL-001/plan.md, acceptance.md
    | 3 | update_archive.go | archiveSkill        | 1 (`:92`)  | user data …
    | 4 | update_archive.go | archiveLegacySkills | 1 (`:304`) | user data …
```
`SPEC-UPDATE-DATA-SURVIVAL-001` M2 builds a destructive-target registry-as-code plus a static source-scan guard covering those two sites, and "M2 assigns them protection or exemption during implementation". This SPEC's M4 refactors `archiveLegacySkills` in the same file, which shifts `:304` and `:339-353` and restructures the function the sibling's guard scans. I did **not** establish that any sibling *edits* `update_archive.go` — I could not determine that from the plan text, so I am not calling the claim false. But it is asserted, not measured, and there is a concrete line-number coupling. Restate as measured, or narrow to "no sibling *modifies*" with the evidence for it.

**SF-e — three ACs rest on reviewer judgment.** `AC-LSL-016(a)` EXPECTs "at least one line comparing the returned count against the number of SUCCESSFUL entries … NOT against `len(legacySkillIDs)`" — the grep cannot make that distinction. `AC-LSL-015(b)` (MF-3). `AC-LSL-004(c)`-negative explicitly says "the reviewer must confirm it is fixture-only". Each is defensible individually, but §A rule 1's promise is binary evaluation; label them non-binary as AC-LSL-010's bullets now are, or mechanize them.

**SF-f — `§C.2` and `§C.3` are under-specified relative to `§C.0`.** §C.0's setup hard-codes `FIX_SRC=internal/cli/update_archive.go`, but §C.3 directs the mutation "over the GUARD TEST file" without saying to re-point `FIX_SRC` and rebuild `$OVERLAY_JSON`. §C.2 gives its mutation only as prose ("turn the accumulation … back into an in-loop return") with no command, unlike §C.1's concrete `sed`. Both are executable by a careful reader; both are a step below §C.1's standard.

**SF-g — `§C.0`'s P2 is described as discriminating but cannot discriminate within this procedure.** P2 ("the fix is still present — 13 entries") is billed as "the discriminating post-condition the retired procedure lacked". Under the overlay mechanism the source is untouched *by construction*, which the SPEC itself notes one line earlier ("Under overlay this is guaranteed by construction"). P2 is therefore a belt-and-braces check against unrelated damage, not a discriminator of this procedure's failure mode. Harmless, but the framing overstates it — and the framing is what a run-phase reader will rely on.

**SF-h — `REQ-LSL-018` is a compound requirement.** "shall continue to pass with their assertion behaviour unmodified, **and** the `…_All16…` disposition shall be applied consistently across both files" chains two obligations with different subjects under one ID. GEARS-acceptable as ubiquitous form (MP-2 passes), but splitting into 018/019 would let AC-006(a) and AC-006(c) each trace to one requirement.

---

## Fresh audit of the new material

**REQ-LSL-018 (new).** Owns AC-LSL-006, closing the iteration-1 orphan. Observable: (a) suite run, (b) index range, (c) `All16` consistency. Fails when the fix is absent? Partially — (a) would fail if the shrunk list broke a test, (c) fails on a half-applied rename. Baseline observed and reproduced (`2` and `2`; indices `[0][1][2][3] [:5] [:8]`, max index 3, max bound 8, all in range at 13). Defect: the missing existence guard (SF-b) and the compound shape (SF-h).

**AC-LSL-015 (new).** (a) is a clean source-level invariance pin, baseline reproduced (`1`). Note it passes at baseline as well as after the fix — correct and intended for an invariance criterion, not a vacuity defect, but the SPEC does not say so. (b) is defective (MF-3).

**AC-LSL-016 (new).** (b) is mechanically sound and I verified the exact grep — including that Go's subtest indentation is 4 spaces, which the anchor depends on:
```
$ grep -cE '^    --- PASS: TestArchiveLegacySkills_ContinuesAfterFailure/success_count_excludes_failures'
1
```
It also fails correctly when the fix is absent: an aborting loop yields the wrong success count, the subtest fails, and the PASS line disappears. (a) is judgment-based (SF-e). Its interaction with AC-010(b) is MF-1.

**Four Epic-wide defect classes, across all 16 ACs.**

- *Baseline mis-attribution* — one instance (MF-3). Every other recorded baseline was re-measured this iteration and reproduced exactly: AC-001(a) `16`, AC-002 `3`, AC-003(a) `1`, AC-006(c) `2`/`2`, AC-006(b) 7 sites unchanged, AC-008 `3`, AC-009(a) `7`, AC-009(b) the same 7 directory names, AC-009(d) `0`, AC-012(a) `3`, AC-013 `1`/`2`, AC-014(b) `exit=0,0,0`, AC-014(c) file listed with exactly the claimed single import-order hunk, AC-015(a) `1`.
- *Non-operative falsification* — **zero instances.** All four §C procedures were executed. This is the largest improvement over iteration 1.
- *Tautological reproduction* — zero harmful instances. AC-015(a) passes pre- and post-fix by design (invariance). Every fix-bearing AC has a baseline that differs from its EXPECT.
- *Defect-locus mis-attribution* — one instance, internal to the SPEC (SF-a). All source citations re-checked this iteration resolve: `planArtifactNames`, `archiveVersion`, the six `.moai/archive/skills` joins, the three in-loop returns (`3`), the branch-guard pattern list.

**Scope discipline: holds.** No new AC or milestone reaches into the four out-of-scope items. The only mentions of `migrate_restore_skill_test.go` are AC-006(c)'s read-only `All16` count and a `§D` line carrying the follow-up forward — neither touches `restoreSkill` or its `--force` overwrite hazard. Nothing addresses downstream user archives, `moai-meta-harness`, `archiveVersion` (which AC-015 *pins* rather than changes), or the other 13 entries (which AC-001(b) pins verbatim and in order). `.gitignore` remains guarded negatively by AC-009(d), measured `0`.

**Tier M remains correct at AC 16 / REQ 18.** File count is unchanged by the revision: `update_archive.go` + 2 new test files + at most 2 test files renamed + 3 archive deletions ≈ 8 files, above the Tier S `<5` bound and well within Tier M's 5-15. AC and REQ counts are not tier inputs. The delivered 3-artifact set matches Tier M. No change warranted.

---

## Gaps (not inferred)

1. **`moai spec lint` did not run to completion** — `timeout 200 moai spec lint` returned `exit=124` with no output. The progress.md lint claims (`0 error(s), 62 warning(s)`, no finding for this SPEC, and the `--strict` variant) are therefore **unverified this iteration**. MP-3 was passed on a manual field-by-field schema check instead. This is a gap in my verification, not a defect claim against the SPEC.
2. **Whether any sibling SPEC will *modify* `internal/cli/update_archive.go`** could not be determined from the sibling plan text (SF-d). I established the citations and the line-number coupling; I did not establish edit-intent either way.

---

## Recommendation

Three localized edits, then a **delta-scoped re-audit** of exactly those three plus a re-check that no other unanchored `--- PASS:` count was left behind. A from-scratch iteration-3 audit is unnecessary.

1. **MF-1** — anchor `acceptance.md:288` to `grep -cE '^--- PASS: TestArchiveLegacySkills_ContinuesAfterFailure'`, EXPECT 1. Sweep the remaining unanchored output greps while there.
2. **MF-2** — delete or truthfully restate `plan.md §A D1` leg (ii) (and its restatement at `progress.md:69`). The six sibling scores are 0.81–0.88; none is skip-eligible; the mechanical cost is zero. Leg (i) already carries the decision.
3. **MF-3** — give `AC-LSL-015(b)` the measured number: EXPECT 6, baseline 6 at lines 73/287/298/311/328/347.

SF-a through SF-h are one-line edits and should ride the same pass; SF-b and SF-d are the two worth prioritizing (a latent vacuity in the only whole-suite regression AC, and an asserted-not-measured parallelism claim in an Epic whose entire purpose is to stop asserted-not-measured claims).

**On the verdict.** The aggregate is 0.81 against a 0.80 threshold and the trend is positive (0.79 → 0.81), so nothing here triggers a STOP or a scope-reduction proposal. I am returning FAIL rather than PASS-with-debt for one reason: MF-1 makes the Definition of Done unachievable, and an acceptance surface that a correct implementation must fail is not a matter of degree. The other two MUST-FIX items are ordinary defects and would not, alone, have overridden the aggregate. If the orchestrator prefers to proceed, MF-1 is a single-character-class edit that can be applied and re-verified in one command — that is a faster path than a PASS-with-debt escalation.

A closing note on what this SPEC got right, since an adversarial report under-reports it. The falsification machinery in `§C.0`-`§C.4` is the first in this Epic that I could execute end-to-end and find operative in every particular — including two non-obvious platform behaviours (BSD `sed` honouring `\n`, and `-overlay` accepting `_test.go` substitution) that the procedure silently depends on and that happen to hold. The baseline hygiene remains the Epic's strongest: every recorded value I re-measured reproduced exactly, save the one in MF-3. The two remaining premise defects are both of the form "correct citation, unchecked implication" — a narrower failure than iteration 1's, and a sign the discipline is converging rather than drifting.

---

## Iteration 3

Iteration: 3 (of max 3 — final allowed iteration)
Auditor: plan-auditor (adversarial, M1-M5 bias-prevention protocol active)
Audited at: 2026-07-31, working tree `main`, SPEC at `version: "0.3.0"`
Artifacts read: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M input contract).

**Reasoning context ignored per M1 Context Isolation.** The v0.3.0 HISTORY row's claim that MF-1..MF-3 and SF-a..SF-h are all resolved was treated as a hypothesis. Each closure below was re-derived by executing a command against the post-revision tree.

**Scope.** Narrowed by orchestrator instruction under a provider rate limit. Not re-audited: the 13 defect-locus citations, the 28 reproduced baselines, scope discipline, the Epic-ordinal decision, the Tier M classification — all confirmed sound at iteration 2. Budget spent on (1) the three MUST-FIX closures, (2) the six claimed-not-verified SF items, (3) the two carried-forward gaps, (4) the orchestrator-authored v0.3.0 progress.md section.

**Environment-trap discipline applied.** `ls | grep '^name'` was not used. One of my own extraction commands returned a surprising zero-length result; I applied the broken-command-first rule to myself, found the fault was a zero-width `[0-9]*$` match in my own pattern, and re-ran with a corrected form rather than recording the artifact as a finding.

---

## Verdict

| | |
|---|---|
| **Verdict** | **PASS** |
| **Overall Score** | **0.89** (harmonic mean of 4 dimensions; iteration 1: 0.79, iteration 2: 0.81) |
| **Tier** | M — threshold **0.80** |
| **Must-pass firewall** | no MP failure (MP-1…MP-7 all PASS or N/A) |
| **Score trend** | 0.79 → 0.81 → 0.89, monotonically improving — no STOP escalation, no scope-reduction proposal |

**The iteration-2 FAIL is cleared at its cause.** MF-1 was not point-fixed; a mechanical sweep of every count-based criterion in `acceptance.md` confirms no unanchored specific-non-zero-EXPECT pattern remains anywhere in the acceptance surface. The `§D` Definition of Done is now achievable by a structurally correct implementation, which is the single thing that made iteration 2 uncertifiable. MF-2 and MF-3 are closed with their supporting citations re-verified to the line.

Both carried-forward gaps are **closed in this audit** rather than carried into run-phase — one of them (`moai spec lint`) in 0.46 s, by reading the command's own `--help`.

Four SHOULD-FIX findings remain. None blocks implementation; three are one-token edits. Two of them are claim-integrity defects inside the orchestrator-authored progress.md section, and are reported at full severity precisely because that section is the one that most loudly claims command-attribution.

---

## Must-Pass Results

- **[PASS] MP-1 — REQ number consistency.** Re-checked after the 018→018/019 split.
  ```
  $ grep -o '^- \*\*REQ-LSL-[0-9][0-9][0-9]' spec.md | sed 's/.*REQ-LSL-//' | sort | diff <(seq -f '%03g' 1 19) -
  (no output)
  → REQ-LSL-001..019 sequential, no gaps, no duplicates, 3-digit padded  (count: 19, dupes: none)

  $ grep -o '^### AC-LSL-[0-9][0-9][0-9]' acceptance.md | sed 's/.*AC-LSL-//' | sort | diff <(seq -f '%03g' 1 16) -
  (no output)
  → AC-LSL-001..016 sequential, no gaps, no duplicates
  ```
- **[PASS] MP-2 — GEARS format compliance.** `REQ-LSL-018`'s compound "and" clause (SF-h) is split; 018 and 019 are each single-subject ubiquitous form. No regression from iteration 2's PASS.
- **[PASS] MP-3 — YAML frontmatter validity.** Upgraded from iteration 2's pass-with-gap: the lint run that timed out at iterations 2 and at the orchestrator's re-attempt now completes.
  ```
  $ moai spec lint .moai/specs/SPEC-UPDATE-LEGACY-SKILL-LIST-001/spec.md
  ✓ No findings — all SPEC documents are valid
  $ echo $?
  0
  real 0.46s
  ```
  Zero findings against the 12-field canonical schema, EARS modality, REQ uniqueness, AC→REQ coverage, and the Out of Scope rule. MP-3 no longer rests on a manual field-by-field check.
- **[N/A] MP-4 — Section 22 language neutrality.** Single-language (Go) internal-CLI SPEC; `NFR-LSL-002` forbids touching `internal/template/templates/`. Auto-passes.
- **[PASS] MP-5 — D7 cross-SPEC reconciliation.** Unchanged from iteration 2; `related_specs` set is unmodified at v0.3.0. No BLOCKING finding.
- **[PASS] MP-6 — D8 cross-platform discipline.**
  ```
  $ grep -c 'syscall' spec.md plan.md acceptance.md progress.md
  spec.md:0
  plan.md:0
  acceptance.md:0
  progress.md:0
  ```
  Auto-PASS.
- **[PASS] MP-7 — clarification gate.**
  ```
  $ grep -c 'NEEDS CLARIFICATION' plan.md   → 0
  $ ls research.md                          → No such file or directory  (correct: Tier M)
  ```

---

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|-----------|------:|-------------|----------|
| Clarity | 0.88 | 0.75–1.0 | MF-2's retraction table is the strongest artifact in this SPEC — it records not just the corrections but the shared failure mode behind both, and derives a transferable rule from it. Deduction: `plan.md §A D1`'s lead-in still says "Three measured facts" over two legs (**D1** below), a stale count introduced by the very edit that closed MF-2. |
| Completeness | 0.90 | 0.75–1.0 | All sections present; 6 × `### Out of Scope — <topic>` H3 with `-` bullets. Both iteration-2 gaps are stated honestly rather than papered over, with an explicit "do not restate this as a current observation" on the stale lint value. |
| Testability | 0.85 | 0.75–1.0 | The proven contradiction is gone — swept across all 16 ACs, verified mechanically (below). MF-3's baseline is measured and exact. The three reviewer-judgment ACs are now either labelled `[NON-BINARY]` or mechanized. Deduction: §A 3a's own "state which row" obligation is met by 5 of ~9 executable count criteria (**D4**), and the table carries no row for the parent-FAIL shape three criteria actually use (**D5**). |
| Traceability | 0.95 | 0.75–1.0 | 19/19 REQ and 5/5 NFR literally cited in `acceptance.md`, zero uncovered, zero orphan ACs, verified by loop. SF-a's wrong cross-reference is corrected. `AC-LSL-006` parts (a-b)/(c) now trace to 018/019 separately. |

Harmonic mean = 4 / (1/0.88 + 1/0.90 + 1/0.85 + 1/0.95) = 4 / 4.4766 = **0.894**.

```
$ for id in $(grep -o 'REQ-LSL-[0-9]*\|NFR-LSL-[0-9]*' spec.md | sort -u); do grep -q "$id" acceptance.md || echo "UNCOVERED: $id"; done
(no output — zero uncovered)
```

---

## 1. The three iteration-2 MUST-FIX closures

### MF-1 — anchoring class recurrence — **CLOSED (swept, not reworded)**

This was iteration 2's sole FAIL driver, so it was audited as the pivot. Two independent questions: (a) does any unanchored non-zero count remain, and (b) would §A 3a as written prevent recurrence.

**(a) Mechanical sweep.** Every `grep -c` over `go test` output in `acceptance.md`, classified against §A 3a's own table:

```
$ grep -n -- 'grep -c.*---' acceptance.md
```

| Line | Pattern | EXPECT | Anchored? | Rule 3a row | Verdict |
|------|---------|-------:|-----------|-------------|---------|
| 130 | `^--- PASS: TestLegacySkillIDsNotEmbedded` | 1 | yes | row 1 | correct |
| 164 | `^--- FAIL: TestLegacySkillIDsNotEmbedded` | 1 | yes | (no row — **D5**) | correct |
| 195 | `^--- PASS` | 19 | yes | row 1 (cited) | correct |
| 205 | `--- FAIL` | 0 | no — deliberate | row 3 (cited) | correct |
| 245 | `^    --- SKIP: …/` | 2 | yes | row 2 (cited) | correct |
| 326 | `^--- PASS: TestArchiveLegacySkills_ContinuesAfterFailure` | 1 | yes | row 1 (cited) | **the MF-1 fix** |
| 349 | `--- FAIL` | 0 | no — deliberate | row 3 | correct |
| 383 | `--- FAIL` | 0 | no — deliberate | row 3 | correct |
| 433 | `--- FAIL` | 0 | no — deliberate | row 3 (cited) | correct |
| 517 | `^    --- PASS: …/success_count_excludes_failures` | 1 | yes | row 2 | **the MF-1 counterpart** |
| 600 | `^--- FAIL: …` (§C.1) | 1 | yes | (no row — **D5**) | correct |
| 631 | `^--- FAIL: …` (§C.2) | 1 | yes | (no row — **D5**) | correct |
| 658 | `^    --- SKIP: …/` (§C.3) | 0 | yes | n/a | correct — paired with 659, which catches a vacuous zero |
| 659 | `^    --- PASS: …/` (§C.3) | 2 | yes | row 2 | correct |

**Result: zero unanchored criteria with a specific non-zero EXPECT.** The two criteria that formed iteration 2's contradiction — `AC-LSL-010(b)` at 326 and `AC-LSL-016(b)` at 517 — are now anchored at parent depth and subtest depth respectively, and are mutually satisfiable: an implementation with the pinned shape yields 1 and 1. The `§D` Definition of Done is achievable.

```
$ grep -c '\^--- PASS\|\^--- SKIP\|\^--- FAIL' acceptance.md
8
```

**(b) Would the rule prevent recurrence?** Partially, and honestly so. Its strengths are real: it is `[HARD]`, it sits in §A where every AC author reads it, it carries a three-row table with rationale, it names the recurrence in its own text ("this rule exists because the defect it prevents recurred once"), and — the non-obvious part — it correctly resists over-correction by mandating that EXPECT-0 forms stay *unanchored* for sensitivity. A blanket "anchor everything" rule would have been the easy write and the wrong one.

Its weakness is that it is prose-only with no mechanical self-check, and its own row-citation obligation is unevenly applied (**D4**). Given this class already survived one point-fix, a rule that binds only future authors by convention is adequate but not airtight. It is not, however, a defect in the acceptance surface — it is a residual risk, recorded as SHOULD-FIX.

### MF-2 — D1 mechanism argument retracted — **CLOSED; no third causal claim; both table entries accurate**

**No third causal claim.** `plan.md §A D1` leg (ii) now reads in full: "Renumbering six sibling artifacts is outside this SPEC's scope. … That is the whole of the argument — it is a scope statement, and it needs no supporting mechanism." No mechanism about branches, caches, or re-audit costs is asserted. Leg (i) is a roster-arithmetic argument, not a causal one. The generalized rule ("when a decision is a scope statement, state the scope and stop") is stated and followed.

**Retraction table entry 1 (v0.1.0 — branch collision) — accurate.**
```
$ git diff --stat main plan/epic-update-config-audit
(empty)
$ git rev-list --count --left-right main...plan/epic-update-config-audit
0	17
```
Both values match the table verbatim. The branch is fully merged; no collision was possible. The retraction is correct.

**Retraction table entry 2 (v0.2.0 — plan-artifact-hash cost) — accurate in all three legs.**
```
$ grep -n 'planArtifactNames\|persist across separate' internal/runtime/audit_cache.go
61:// planArtifactNames is the ordered list of plan artifact file names to hash.
63:var planArtifactNames = []string{
74:// persist across separate /moai run invocations. For cross-invocation caching,
$ sed -n '63,68p' internal/runtime/audit_cache.go
var planArtifactNames = []string{
	"acceptance.md",
	"plan.md",
	"spec.md",
	"tasks.md",
}
```
Cited as `:63-68` hashing `{acceptance.md, plan.md, spec.md, tasks.md}` — exact, including the ordering. The `:73-74` verbatim quote is exact.
```
$ sed -n '319p' .claude/rules/moai/workflow/spec-workflow.md
    2. **Overall score ≥ 0.90.**
```
Cited as "skip-eligibility condition 2 at score ≥ 0.90, `spec-workflow.md:319`" — exact to the line.
```
$ for f in .moai/reports/plan-audit/*epic-update-config-iter2.md; do grep -m1 -i 'Overall Score' "$f"; done
CONFIG-KEY-HONESTY-001    0.81      CONFIG-TIER-PERSIST-001   0.82
UPDATE-CI-GUARD-001       0.85      UPDATE-DATA-SURVIVAL-001  0.84
UPDATE-DOC-DRIFT-001      0.82      UPDATE-REINSTALL-LOOP-002 0.88
```
Range 0.81–0.88 — exactly as the table states. None reaches 0.90, so none is skip-eligible and condition 3 is moot. **The claimed cost is indeed zero, and the retraction is correct.**

### MF-3 — `AC-LSL-015(b)` numeric EXPECT — **CLOSED; count and all six line numbers exact**

```
$ grep -cE '"\.moai", "archive", "skills"' internal/cli/update_archive.go
6
$ grep -nE '"\.moai", "archive", "skills"' internal/cli/update_archive.go
73:	dstDir := filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion, skillID)
287:		dstDir := filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion, id)
298:				backupDir := filepath.Join(projectRoot, ".moai", "archive", "skills",
311:				archiveBackupRel := filepath.Join(".moai", "archive", "skills",
328:		archiveDst := filepath.Join(".moai", "archive", "skills", archiveVersion, id)
347:		archiveDst := filepath.Join(".moai", "archive", "skills", archiveVersion, id)
```

`EXPECT: 6` matches. All six line numbers match, in order, including `347` in `dryRunArchiveLegacySkills` — the site the v0.2.0 prose baseline omitted. The function attribution in the AC's comment block is correct for each. `archiveVersion = "v2.16"` confirmed at `:29`, so AC-LSL-015(a)'s EXPECT of 1 also still holds.

---

## 2. SF-a … SF-h — all eight resolved

Six audited this iteration; SF-b and SF-g were confirmed by the orchestrator and re-confirmed incidentally while reading.

| Item | Status | Evidence |
|------|--------|----------|
| **SF-a** — §C.1 heading binds wrong AC | **RESOLVED** | `$ grep -n '^### C\.' acceptance.md` → `585:### C.1 Guard falsification (binds AC-LSL-004 and AC-LSL-005 → REQ-LSL-008)`. `AC-LSL-008` correctly removed; `§C.4` still binds it at `673`. |
| **SF-b** — AC-006 missing existence guard | **RESOLVED** | `acceptance.md:191-201` adds `(a-i)` — `grep -cE '^--- PASS'`, EXPECT 19, with a CONSTRAINT note explaining that a prefix-dropping rename reduces the count and that this is intended detection, not a false positive. The vacuity path SF-b identified is closed. |
| **SF-c** — AC-014(b) `grep -c` exit-1 hazard | **RESOLVED** | `acceptance.md:427-429` carries the NOTE, and the commands use the capture form `bf=$(grep -c …)` / `tf=$(grep -c …)`, matching AC-009(d)'s pattern. |
| **SF-d** — disjointness claim asserted, not measured | **RESOLVED** | `progress.md:194-214` replaces the assertion with a measured `grep -rn` across all seven sibling directories, then explicitly separates "**Established**: a live line-number coupling" from "**Not established**: whether any sibling *modifies* the file", and carries the latter forward as an open gap rather than resolving it by assertion. This is the correct handling. |
| **SF-e** — three ACs rest on reviewer judgment | **RESOLVED** | `$ grep -n 'NON-BINARY' acceptance.md` → `140` (AC-004(c)-negative), `503` (AC-016(a)). The third, AC-015(b), was *mechanized* rather than labelled — MF-3 gave it `EXPECT: 6`. Both dispositions are legitimate. |
| **SF-f** — §C.2/§C.3 under-specified | **RESOLVED** | §C.2 (`620-625`) now carries a concrete `sed` plus a `cmp -s` guard that surfaces a non-matching pattern instead of silently producing a byte-identical "mutation". §C.3 (`641-671`) adds step 0 (re-point `FIX_SRC`, rebuild `$OVERLAY_JSON`) and step 4 (re-point back). Both now meet §C.1's standard. |
| **SF-g** — §C.0 P2 framing overstated | **RESOLVED** | `acceptance.md:579` now states plainly that under the overlay P2 "cannot fail as a consequence of *this* procedure", is "belt-and-braces … not the load-bearing check", and explains why it is retained anyway — closing with the reason the framing matters: "a run-phase reader relying on 'P2 is the discriminating check' would over-trust a green P2." The overstatement is not merely removed but inverted into a warning. |
| **SF-h** — REQ-LSL-018 compound | **RESOLVED** | `spec.md:207-208` splits into 018 (six files still pass) and 019 (All16 disposition consistency + prefix preservation). `acceptance.md:180` traces parts (a-b) → 018 and part (c) → 019. |

---

## 3. The two carried-forward gaps — both closed

### Gap 1 — `moai spec lint` — **CLOSED. The gap statement is half accurate and half a false-absence claim.**

The **first** clause is accurate and well-handled: `progress.md:174` correctly marks the recorded `0 error(s), 62 warning(s)` as a v0.2.0-era result and instructs "Do not restate it as a current observation." That is exactly right, and it is the disciplined move.

The **second** clause is false. It states: "The `moai spec lint --spec <ID>` form does not exist (`Unknown flag: --spec`), so **no cheap per-SPEC narrowing is available**." The flag half is accurate — I reproduced it:
```
$ moai spec lint --spec SPEC-UPDATE-LEGACY-SKILL-LIST-001
ERROR  Unknown flag: --spec.
```
But the command's own `--help` carries the narrowing on its USAGE line:
```
$ moai spec lint --help
  USAGE
    moai spec lint [spec.md...] [--flags]
```
Positional paths are accepted. The narrowed run:
```
$ moai spec lint .moai/specs/SPEC-UPDATE-LEGACY-SKILL-LIST-001/spec.md
✓ No findings — all SPEC documents are valid
$ echo $?
0
real 0.46s
```
**0.46 seconds**, against the 120–200 s timeouts that produced `exit=143` / `exit=124` on the whole catalog. The gap did not need to be carried at all.

**This SPEC lints clean at v0.3.0 — zero findings, exit 0.** MP-3 is upgraded accordingly. The stale `62 warning(s)` figure belongs to the whole-catalog run and can now simply be dropped rather than caveated.

Recorded as **D3** below, because inferring "no narrowing is available" from one rejected flag — without reading the help text — is an unobserved-absence claim, the same `verification-claim-integrity.md` §1.1-surface-3 class this Epic exists to fix.

### Gap 2 — does any sibling *modify* `internal/cli/update_archive.go`? — **CLOSED: no sibling modifies it.**

```
$ grep -rn 'update_archive' <all seven sibling dirs>
SPEC-UPDATE-REINSTALL-LOOP-002/acceptance.md:238   table row: dryRunArchiveLegacySkills → "[dry-run] total:"
SPEC-UPDATE-DATA-SURVIVAL-001/acceptance.md:161-162  registry rows: archiveSkill (:92), archiveLegacySkills (:304)
SPEC-UPDATE-DATA-SURVIVAL-001/plan.md:108,116-117    same registry rows + a scope-list mention
SPEC-UPDATE-DOC-DRIFT-001/spec.md:299                dryRunArchiveLegacySkills (:339-353)
SPEC-UPDATE-DOC-DRIFT-001/plan.md:290                ":339-353 — M1 evidence sites"
```
Every reference is a table row or an evidence anchor. `SPEC-UPDATE-DATA-SURVIVAL-001` is the only sibling whose milestones touch the file's *functions*, and its M2 edit targets are explicit:
```
$ sed -n '207,219p' SPEC-UPDATE-DATA-SURVIVAL-001/plan.md
### M2 — Destructive-target registry + `.moai/memory/` backup + comment reconciliation
Introduces the registry of §C.0 as code (10 rows / 17 sites …), its drift guard,
the missing `.moai/memory/` backup, and the `dirs.go` brand+db group comment fix.
```
Registry code, drift guard, `.moai/memory/` backup, `dirs.go` comment. `update_archive.go` is **scanned**, not edited. Its other five milestones (M1, M3–M6) name unrelated targets.

**Conclusion: read-coupling, not write-coupling.** `progress.md`'s revised claim — scope disjoint, line numbers coupled — is confirmed correct, and so is its sequencing recommendation, which I can now strengthen from "lower-risk" to **materially recommended**: `SPEC-UPDATE-DATA-SURVIVAL-001` M2 lands a registry hard-coding `update_archive.go:92` and `:304` behind a drift guard that enumerates sites by static source scan. This SPEC's M1 shifts every line below the slice up by three, and its M4 restructures `archiveLegacySkills`. Landing this SPEC **second** would turn that guard red on arrival. Landing it **first** costs nothing, because DATA-SURVIVAL-001 has not yet been implemented and would simply register the post-restructure line numbers.

---

## 4. The orchestrator-authored v0.3.0 progress.md section

Audited as an artifact, not as a report from a trusted source.

**Attribution and honesty — strong, and unusually so.** The section:
- **discloses its own authorship anomaly** up front (`progress.md:160`), naming the `manager-spec` rate-limit termination, stating which artifacts had already been written and were complete, and asserting that no claim is carried over from the terminated agent's report. This is a correct ledger closure under `agent-common-protocol.md` § Ledger Closure clause (a).
- **draws the claimed/verified line explicitly** and against its own interest (`progress.md:170`): "the remaining six are recorded on the authoring agent's claim and were **not** independently re-measured by the orchestrator. Treat them as claimed-not-verified until iteration 3 audits them." Having now audited all six and found all six genuinely resolved, I note that the orchestrator's caution was *conservative* — it under-claimed rather than over-claimed. That is the correct direction of error.
- **refuses to restate a stale measurement** (`progress.md:174`), marking the v0.2.0 lint value as not re-confirmed with an explicit instruction not to quote it as current.
- **carries gap 2 forward without resolving it by assertion in either direction** (`progress.md:212`).

**The `plan_audit.iteration_2` yaml block is faithful to the real iteration-2 verdict.** Field by field against `audit.md` `# Iteration 2`: `verdict: FAIL` ✓; `score: 0.81` ✓; `threshold: 0.80` ✓; `trend: 0.79 -> 0.81 (rising; no STOP / scope-reduction triggered)` ✓ (iteration 2: "0.79 → 0.81, improving — no STOP escalation, no scope-reduction proposal"); the `note` ✓ (iteration 2: "The aggregate clears the Tier M threshold. The FAIL is not driven by the aggregate… Definition of Done … cannot be achieved by any implementation"); `report:` path resolves. `must_pass: 7/7` is a defensible shorthand for iteration 2's "MP-1…MP-7 all PASS or N/A" — MP-4 was N/A and N/A auto-passes under the firewall — though it does flatten iteration 2's recorded MP-3 lint gap (**D6**, minor).

**Its structural counts are all reproducible.**
```
$ grep -o 'REQ-LSL-[0-9]*' spec.md | sort -u | wc -l   → 19
$ grep -c '^- \*\*REQ-LSL-' spec.md                    → 19
$ grep -o 'NFR-LSL-[0-9]*' spec.md | sort -u | wc -l   → 5
$ grep -c '^### AC-LSL-' acceptance.md                 → 16
$ (uncovered-ID loop)                                  → no output
```
"REQ `19` unique IDs, NFR `5`, AC `16`, zero uncovered" — all four exact.

**Two of its supporting claims are defective** (D2, D3 below). Both sit in the evidence column rather than in the claimed/verified boundary, so neither undermines the section's integrity — but both are the same class of error the section is written to guard against, which is why they are reported rather than waved through.

---

## Defects Found

### MUST-FIX

None. No finding here makes the Definition of Done unachievable, contradicts a must-pass criterion, or blocks a structurally correct implementation.

### SHOULD-FIX

**D1 — `plan.md §A D1` says "Three measured facts" over two legs.** `plan.md:11` reads "Three measured facts drive that:", followed by exactly `**(i)**` (`:13`) and `**(ii)**` (`:38`), then "**Decision.**".
```
$ sed -n '9,55p' plan.md | grep -n '^\*\*('
5:**(i) The existing "of 6" numbering is already inconsistent …
30:**(ii) Renumbering six sibling artifacts is outside this SPEC's scope. …
$ grep -c '(iii)' plan.md
0
```
v0.2.0 carried three legs; MF-2's fix removed the mechanism leg and left the count word behind. The retraction table is not a third "fact driving the decision" — it is explicitly a record of two *retracted* rationales — and "Decision." is not a fact. Severity: **minor**, one-token fix. Reported because it is a stale numeric claim inside the one block that has now been revised three times and whose own stated lesson is "state the scope and stop", and because an implementer counting to three will look for a leg that is not there. **Required fix:** change "Three measured facts" to "Two measured facts" at `plan.md:11`.

**D2 — `progress.md` MF-1 evidence miscounts the residual unanchored occurrences.** `progress.md:164` claims: "Exactly **one** unanchored `--- PASS:` occurrence remains, and it sits inside §A rule 3a's own explanatory prose as the counter-example — not in an executable criterion." Measured:
```
$ grep -n -- '--- PASS:' acceptance.md | grep -v '\^--- PASS:\|\^    --- PASS:'
13:3a. **[HARD] Anchoring rule … An unanchored `grep -c -- '--- PASS: TestX'` therefore counts parent plus every passing subtest …
126:#     ("    --- PASS: Test.../production"), so an unanchored grep counts the
```
**Two, not one.** The second is inside AC-LSL-004(b)'s inline explanatory comment. The substantive conclusion survives — neither is an executable criterion, and my independent sweep (§1 above) confirms zero unanchored criteria — but the stated count is wrong in a row whose purpose is to evidence the MF-1 closure. Severity: **minor**. **Required fix:** restate as "Two unanchored `--- PASS:` occurrences remain, both in explanatory prose (`acceptance.md:13` rule 3a's counter-example, `:126` AC-LSL-004(b)'s inline note) — neither in an executable criterion."

**D3 — `progress.md` gap 1 asserts a capability absence that was never checked.** `progress.md:174`: "The `moai spec lint --spec <ID>` form does not exist (`Unknown flag: --spec`), so **no cheap per-SPEC narrowing is available**; a bounded re-run belongs to iteration 3." The premise is true; the conclusion is false. `moai spec lint --help` documents `USAGE: moai spec lint [spec.md...]`, and the narrowed run completes in **0.46 s** with `✓ No findings`, exit 0 (§3 above). One rejected flag was generalized into an absence of capability without reading the help text — an unobserved-absence claim under `verification-claim-integrity.md` §1.1 surface 3, and the same class the Epic exists to eliminate. Severity: **major** (the claim class, not the consequence — the consequence was one deferred iteration). **Required fix:** replace gap 1 with the observed result: `moai spec lint .moai/specs/SPEC-UPDATE-LEGACY-SKILL-LIST-001/spec.md` → `✓ No findings`, exit 0, 0.46 s. Drop the stale `62 warning(s)` figure rather than caveating it — it was a whole-catalog number and is not this SPEC's result.

**D4 — §A 3a's own row-citation obligation is unevenly applied.** The rule states: "Any new AC that counts `go test` output MUST state which row of the table above it falls under." Satisfied by AC-006(a-i), AC-006(a-ii), AC-007(b), AC-010(b), AC-014(b). Not satisfied by AC-004(b) (`:124-128`, gives the reasoning but names no row), AC-005 (`:166`, "for the same reason as AC-LSL-004(b)"), AC-016(b) (`:522`, describes the indent but names no row), or the four §C count steps (`:600`, `:631`, `:658`, `:659`). A strict reading of "**new** AC" grandfathers these, which is defensible — but the un-cited majority is the precedent a future author copies, and this class has already survived one point-fix. Severity: **minor**. **Required fix:** add the row citation to the four named ACs, or state explicitly in rule 3a that the obligation binds only ACs added after v0.3.0.

**D5 — rule 3a's table has no row for the parent-FAIL shape three criteria use.** The table covers parent-PASS (row 1), subtest-PASS/SKIP (row 2), and any-FAIL-EXPECT-0 (row 3). `AC-LSL-005` (`:164`), §C.1 (`:600`), and §C.2 (`:631`) all count `^--- FAIL: TestX` with EXPECT 1 — a parent result line with a specific non-zero EXPECT. They are correctly anchored under the rule's prose ("anchor when counting a specific non-zero number"), but the "state which row" obligation is unsatisfiable for them because no row exists. Severity: **minor**. **Required fix:** add a row — "A parent result line (PASS or FAIL), EXPECT ≥ 1 | `grep -cE '^--- (PASS|FAIL): TestX'`" — or widen row 1 to cover both verbs.

**D6 — `plan_audit` yaml `must_pass: 7/7` flattens an N/A and a recorded gap.** Both iteration blocks record `7/7`. Iteration 2's actual result was "MP-1…MP-7 all PASS or N/A", with MP-4 N/A and MP-3 passed on a manual schema check with the lint run explicitly recorded as a gap. N/A auto-passing is firewall-consistent, so `7/7` is not wrong — but a downstream reader cannot see that MP-3's basis was caveated. Severity: **minor**, fidelity only. **Required fix (optional):** annotate as `must_pass: 7/7 (MP-4 N/A; MP-3 passed on schema check, lint unverified at iter 2 — resolved at iter 3)`.

---

## Regression Check (iteration-2 defects)

| Item | Status | Evidence |
|------|--------|----------|
| MF-1 — anchoring class recurrence | **RESOLVED** | Full sweep of 14 count criteria (§1). Zero unanchored non-zero-EXPECT patterns. AC-010(b) and AC-016(b) mutually satisfiable. |
| MF-2 — D1 leg (ii) false consequence | **RESOLVED** | Mechanism argument removed, not replaced. Both retraction-table entries re-verified exact (branch `0 17` + empty diff; `audit_cache.go:63-68` / `:73-74`; `spec-workflow.md:319`; sibling scores 0.81–0.88). |
| MF-3 — AC-015(b) non-binary + undercounted baseline | **RESOLVED** | `EXPECT: 6` and all six line numbers verified exact, including the previously-omitted `:347`. |
| SF-a … SF-h | **ALL RESOLVED** | Per-item evidence in §2. Six audited this iteration, two previously confirmed. |

**No defect appears in all three iterations unchanged.** No stagnation. The score trend 0.79 → 0.81 → 0.89 is monotonically improving with no regression, so no STOP signal and no scope-reduction proposal are triggered.

---

## Gaps (not inferred)

1. **The §C.1–§C.4 falsification procedures were not re-executed this iteration.** Iteration 2 executed all four end-to-end and found them operative; v0.3.0's changes to §C.2 and §C.3 (SF-f) add a `cmp -s` mutation guard and an overlay re-point, both of which I verified by reading rather than by running. The SF-f edits are additive and cannot break a procedure that already worked, but I did not observe the amended forms execute. Scope-limited by the rate-limit budget, not by judgment.
2. **The whole-catalog `moai spec lint` still does not complete** within a 200 s bound. This SPEC lints clean in isolation (0.46 s), which is what MP-3 needs. Whether the catalog-wide timeout is a performance defect in `moai spec lint` is outside this SPEC and is not a finding against it — but it is worth a separate issue.
3. **`REQ-LSL-019`'s GEARS form was assessed by reading, not by lint rule.** `moai spec lint` reports no EARS-modality finding, which covers it mechanically; I did not separately hand-classify 019 against the five GEARS patterns beyond confirming single-subject ubiquitous form.

---

## Recommendation

**PASS at 0.89 against the Tier M threshold of 0.80. Proceed to Implementation Kickoff Approval.**

No escalation is required. Since this was the final allowed iteration, the three escalation routes are recorded here as not-taken, with the reason:

- **PASS-with-debt** — not applicable. This is an unqualified PASS. The six SHOULD-FIX findings are documentation defects, not debt against the implementation; none constrains what run-phase must build.
- **Scope reduction** — not applicable and would be actively harmful. The score rose monotonically across all three iterations (0.79 → 0.81 → 0.89) with no regression, which is the opposite of the signal that justifies scope reduction. The SPEC's scope has been stable since v0.1.0 (four milestones, one file plus two new test files plus three git deletions); nothing in three audits suggested the scope was the problem.
- **Explicit user override** — not required. The iteration cap is not being exceeded; iteration 3 reached PASS within it.

**Residual debt carried into run-phase: none blocking.** The six SHOULD-FIX items are, in priority order:

1. **D3** (`progress.md` gap 1's false-absence claim) — the only one worth fixing before run-phase, because it is a live claim-integrity defect in the Epic's own subject matter and because the correction *adds* a verified fact (this SPEC lints clean) rather than merely deleting a wrong one.
2. **D1** (`plan.md:11` "Three" → "Two") — one token.
3. **D2** (`progress.md:164` one → two) — one clause.
4. **D4 / D5** (rule 3a self-application and the missing FAIL row) — improve the recurrence guard for future ACs; no effect on the current 16.
5. **D6** (yaml `must_pass` annotation) — optional fidelity.

All six can ride a single v0.3.1 editorial pass, or be folded into the sync-phase artifact update. **None requires a re-audit.**

**One operational recommendation carried out of §3:** sequence this SPEC **before** `SPEC-UPDATE-DATA-SURVIVAL-001`. That SPEC's M2 lands a destructive-target registry hard-coding `update_archive.go:92` and `:304` behind a static-source-scan drift guard; this SPEC's M1 shifts those lines by three and its M4 restructures the enclosing function. Landing second turns that guard red on arrival; landing first costs nothing. `progress.md:214` already reaches this conclusion — it is now measured rather than inferred, and I would raise it from "lower-risk order" to a sequencing requirement.

**A closing note, since an adversarial report under-reports what went right.** Three things in this revision are better than the fix they were asked for. MF-1 was answered with a sweep and a standing rule that names its own recurrence, rather than the one-character edit that would have satisfied the finding — and the rule correctly resists over-correction by keeping EXPECT-0 forms unanchored, which is the harder and less obvious call. MF-2 was answered by *deleting* an argument rather than supplying a third one, and by recording both retractions with a diagnosis of the shared failure mode; that retraction table is the most transferable artifact this Epic has produced. And the orchestrator-authored progress.md section, written under a rate-limit interruption, disclosed its own authorship anomaly and marked six items claimed-not-verified that I have now confirmed were in fact all resolved — it under-claimed against its own interest. The two defects I found in it are of the same class it was written to guard against, which is worth stating plainly; but a section that errs by under-claiming is the one I would rather audit.
