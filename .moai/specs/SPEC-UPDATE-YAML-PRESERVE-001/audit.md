# Plan-Phase Audit — SPEC-UPDATE-YAML-PRESERVE-001

Auditor: `plan-auditor` (independent, adversarial stance)
Date: 2026-07-31
Iteration: 1/3
Tier: M → PASS threshold **0.80**

**Reasoning context ignored per M1 Context Isolation.** The authoring agent's claims (clean `moai spec lint`, "every `file:line` matches the current tree") were treated as unverified and re-measured independently. The "Background" section supplied in the audit request was likewise treated as unverified — one of its claims (`internal/cli/update/coverage_improvement_test.go`) is wrong; the file is `internal/cli/coverage_improvement_test.go`.

---

## Verdict

| | |
|---|---|
| **Verdict** | **FAIL** |
| **Overall score** | **0.71** (harmonic mean; Tier M threshold 0.80) |
| Must-pass firewall | all clear (MP-1..MP-7) |
| Failure driver | Consistency + Testability — one requirement is **unsatisfiable with the chosen mechanism**, one **contradicts a sibling requirement**, and 14 of 22 AC commands **pass vacuously** |

The SPEC is well-researched and its two headline defect diagnoses (Defect A, Defect B) are **CONFIRMED true**. It fails on the third axis: the preservation *contract* it writes is partly unachievable, partly unfalsifiable, and misses one loss class and one unparseable template.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-UYP-001 … REQ-UYP-019, sequential, no gaps, no duplicates, consistent 3-digit padding (`spec.md:121-165`).
- **[PASS] MP-2 GEARS format compliance** — every REQ matches a GEARS pattern: event-driven `When …, the … shall` (001,002,003,006,007,012,013,014,015,016), state-driven `While …, the … shall` (005), ubiquitous `The … shall` (004,008,009,010,011,017,018,019). Zero informal ACs-as-requirements, zero Given/When/Then mislabelled as GEARS.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with canonical names (`spec.md:2-17`); no rejected snake_case alias (`created_at`/`updated_at`/`labels`/`spec_id`). Minor deviations noted as NICE-TO-HAVE #15.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go only; `module: cli`). Auto-pass.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — all 3 `related_specs` exist; none is retired/superseded/archived:
  ```
  SPEC-CLI-TUX-V3-003: status: completed
  SPEC-V3R6-UPDATE-NOISE-001: status: implemented
  SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002: status: completed
  ```
- **[N/A] MP-6 D8 cross-platform discipline** — `grep -c syscall` over all 4 artifacts returns `0` in each. D8 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/` → exit 1, no matches.

---

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | REQ-UYP-004 vs REQ-UYP-005 contradict for 4-space-indented sources (`archive.yaml`); "31 files" stated where 30 exist |
| Completeness | 0.75 | 0.75 | blank-line loss class absent from the requirement set; `quality.yaml.tmpl` unparseability unhandled |
| Testability | 0.50 | 0.50 | 14 of 22 AC commands exit 0 with `no tests to run`; the single falsifiability guard (AC-UYP-022) is non-executable as written |
| Traceability | 1.00 | 1.0 | REQ-001..019 each have ≥1 AC; AC-001..019 each cite a valid REQ; AC-020/021/022 cite spec sections (see NICE-TO-HAVE #14) |

Harmonic mean = 4 / (1/0.75 + 1/0.75 + 1/0.50 + 1/1.00) = **0.706 → 0.71**

---

## What the SPEC gets RIGHT (independently confirmed)

Every one of these was re-measured; none is taken on the author's word.

| Claim | Cited location | Status |
|---|---|---|
| `MergeYAML3Way` map round-trip | `merge.go:20-35` (func at `:20`, `yaml.Marshal` at `:34`) | **CONFIRMED** |
| `MergeYAMLDeep` map round-trip | `merge.go:116-130` (func at `:116`, marshal at `:129`) | **CONFIRMED** |
| Old-only-key drop comment | `merge.go:95-97` — `grep -n 'Keys only in old'` → `95:` | **CONFIRMED verbatim** |
| 2-way preserves old-only keys | `merge.go:168-171` — `grep -n 'Only exists in old'` → `169:` | **CONFIRMED** |
| Misnamed subtest asserting the opposite of its name | `update_yaml_test.go:591-603` — name at `:591`, inverted assertion at `:600` | **CONFIRMED** |
| Three production call sites | `restore.go:121,139,207` | **CONFIRMED** |
| Zero comment-preservation tests | AC-UYP-018 grep executed → 1 hit, and it is the *drop*-asserting one | **CONFIRMED** |
| Comment loss is total on every merge | measured across all 30 templates (see Evidence §2) | **CONFIRMED and worse than stated** — e.g. `llm.yaml` 120→0, `harness.yaml` 65→0, `lsp.yaml.tmpl` 88→0 |
| 0 anchors / 0 merge keys / 0 multi-document | re-surveyed independently | **CONFIRMED — all three counts hold** |
| 8 files carry block sequences | re-derived: `design, harness, interview, lsp.tmpl, mx, quality.tmpl, security, sunset` | **CONFIRMED — exact match** |
| `moai spec lint` clean | `moai spec lint .moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/spec.md` → `✓ No findings` | **CONFIRMED** |

---

## Defects Found

### MUST-FIX

**D1 — `acceptance.md:81-89` + `spec.md:129` (REQ-UYP-005) + `plan.md:186` — byte-identity is UNSATISFIABLE with the chosen mechanism. Severity: critical.**

REQ-UYP-005 requires byte-identical output on a no-change merge; plan M3 calls it "the strongest single assertion"; acceptance §D.2 builds a subsumption argument on it; spec §E success criterion 1 restates it. **Measured against a faithful implementation of exactly what D1 prescribes** (`yaml.Unmarshal` → `yaml.Node` → `yaml.NewEncoder` + `SetIndent(2)` → `Encode`/`Close`), only **8 of 30** templates are byte-identical; **22 differ**.

Failure scenario: M3 is authored, the golden byte-identity assertion is written, and it fails on 22 of 30 subtests through no fault of the implementation. The plan's own §G anti-pattern ("Do not delete a failing assertion to reach green") then blocks the only escape, and M3 deadlocks.

Minimal correction: demote REQ-UYP-005 from byte-identity to the property set it was proxying (comment count **and content**, key order, quoting, emitted indent), and move byte-identity to a documented non-goal with the measured 8/30 figure as its rationale. Update `spec.md` §E criterion 1 and `acceptance.md` §D.2 accordingly.

---

**D2 — `spec.md:127` (REQ-UYP-004) vs `spec.md:129` (REQ-UYP-005) — direct intra-SPEC contradiction. Severity: critical.**

REQ-UYP-004 mandates `SetIndent(2)`. `archive.yaml` is authored at **4-space** indent. A `SetIndent(2)` encoder therefore MUST reindent it, which MUST violate REQ-UYP-005. Verified by diff (Evidence §3): every line of `archive.yaml` shifts left by 2 columns.

This also falsifies the `spec.md:53` premise "**indent widened** (2-space → yaml.v3's default 4-space)" as a general statement — for `archive.yaml` the current behaviour is a *no-op* on indent and the fix *narrows* it.

Minimal correction: state REQ-UYP-004 as "the emitted document uses 2-space block-mapping indentation, normalizing sources that differ", and record `archive.yaml` as the known reindent case. Resolve against D1's demotion of REQ-UYP-005.

---

**D3 — `spec.md:53` ("Four distinct losses") — a fifth loss class (blank lines) is missing from the entire requirement set. Severity: major.**

The node round-trip preserves comment *count* perfectly (measured: every DIFFER file shows `comments N->N`) yet still loses bytes. The delta is **blank lines**, which `yaml.v3` does not model: `workflow.yaml` 6797→6195 (−602 B), `llm.yaml` 9780→9173 (−607 B), `delegation.yaml` 4826→4301 (−525 B). No REQ, no AC, and no Out of Scope entry mentions blank lines.

Failure scenario: the fix ships, all 19 ACs pass, and a user running `moai update` still sees their config visually reflowed — every blank line between config groups collapsed. The #1243 complaint is only partly closed, and the SPEC's own §E criterion 2 ("preserves … every comment, key order, and quoting style") is technically met while the file still looks mangled.

Minimal correction: add a REQ for blank-line preservation, **or** add an explicit `### Out of Scope — blank-line preservation` entry naming the measured byte deltas. Silence is the one option that is not acceptable.

---

**D4 — `plan.md:144` (Decision D6) — the claim that `.tmpl` files "parse fine as scalars" is FALSE, and its consequence is a live production defect the SPEC never identifies. Severity: critical.**

`quality.yaml.tmpl` contains **unquoted** placeholders — `enforce_quality: {{.EnforceQuality}}` (`:13`) and `test_coverage_target: {{.TestCoverageTarget}}` (`:16`). YAML reads `{{…}}` as a nested flow mapping, so the map decode fails:

```
MERGE-FAIL quality.yaml.tmpl: unmarshal new YAML: yaml: invalid map key: map[string]interface {}{".EnforceQuality":interface {}(nil)}
```

Every other `.tmpl` quotes its placeholders; `quality.yaml.tmpl` is the sole exception. Three consequences:

1. **REQ-UYP-017 / AC-UYP-017 are unsatisfiable against the current implementation** — the golden set is defined as *every* `*.yaml`/`*.yaml.tmpl`, and one member cannot be merged at all.
2. **`plan.md:144`'s appeal to `SaveTemplateDefaults` (`backup.go:180-186`) is not evidence** — that function copies `.tmpl` files **raw** and never parses them, so it cannot testify to parseability.
3. **Undiscovered live defect**: `SaveTemplateDefaults` writes the raw template as the base, stripping `.tmpl` (`backup.go:184`), so the 3-way base for `quality.yaml` is unparseable. `restore.go:121` therefore errors and **silently falls back to 2-way merge for `quality.yaml` on every single `moai update` today.** The SPEC's blast-radius section (`spec.md:84-86`) does not mention this.

The node tree *does* parse the file (measured 6605→6607 bytes), so the fix incidentally repairs it — which makes this a behaviour change that must be named, not a free win.

Minimal correction: (a) correct `plan.md` D6; (b) record the `quality.yaml` 2-way-fallback finding in `spec.md` §A blast radius; (c) add an AC asserting `MergeYAML3Way` on `quality.yaml.tmpl` returns nil error post-fix; (d) see D9 for the D5-interaction this creates.

---

**D5 — `acceptance.md:239-244` (AC-UYP-022) — the falsifiability command is NON-EXECUTABLE. Severity: critical.**

Run verbatim:

```
$ git stash push internal/cli/update/backup/merge.go internal/cli/update/backup/node_merge.go
error: pathspec ':(prefix:0)internal/cli/update/backup/node_merge.go' did not match any file(s) known to git
Did you forget to 'git add'?
exit 1
```

`node_merge.go` is a **new, untracked** file at the moment AC-UYP-022 runs, and `git stash push` rejects untracked pathspecs without `-u`. The command aborts atomically — nothing is stashed, so the operator sees the golden test still GREEN and may mis-read that as "the check ran and the tree was fine".

This is the load-bearing defect: AC-UYP-022 is the *only* mechanism in the entire SPEC that distinguishes a real preservation test from a vacuous one, and it does not run.

Minimal correction: `git stash push -u internal/cli/update/backup/merge.go internal/cli/update/backup/node_merge.go`, and assert the intermediate state explicitly (`git status --porcelain internal/cli/update/backup/` shows `merge.go` restored and `node_merge.go` absent) before running the RED test.

---

**D6 — `acceptance.md` §D.1 (14 ACs) — commands pass VACUOUSLY. Severity: critical.**

Every AC of the form `go test ./pkg/ -run TestX` exits 0 when `TestX` does not exist. Verified verbatim against the current tree:

```
$ go test ./internal/cli/update/backup/ -run TestPreserveGolden_Comments -count=1 -v
testing: warning: no tests to run
PASS
ok  github.com/modu-ai/moai-adk/internal/cli/update/backup  15.316s [no tests to run]
exit 0
```

Affected: **AC-UYP-001, -002, -003, -004 (second command), -005, -006, -007, -008, -012, -013, -014, -015, -016, -021** — 14 of 22. Only AC-017 (count comparison), AC-018 and AC-019 (grep with explicit exit) are guarded today.

Failure scenario: run-phase reports a green AC matrix having implemented none of the named tests, or having named them slightly differently (e.g. `TestPreserveGolden_Comment` singular). Every command exits 0 and the matrix reads PASS.

Minimal correction: append an existence assertion to each, e.g.
```bash
go test ./internal/cli/update/backup/ -run TestPreserveGolden_Comments -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestPreserveGolden_Comments'
```
AC-UYP-006's subtest form additionally needs the exact Go-normalized subtest name (spaces → underscores) pinned.

### SHOULD-FIX

**D7 — `plan.md:32`, `plan.md:144`, `acceptance.md:184`, `progress.md:10` — the template count is 30, not 31.**
```
$ ls internal/template/templates/.moai/config/sections/ | grep -cE '\.yaml(\.tmpl)?$'
30
```
AC-UYP-017's command self-derives the count, so the AC survives; the stale figure sits in the risk register and in the "verified baseline" table, weakening confidence in the surrounding measurements. Correct all four occurrences to 30.

**D8 — `plan.md:93` (Decision D2) — "seven test call sites" enumerates eight.**
`update_yaml_test.go:346,925` (2) + `backup_test.go:427,445,536,551` (4) + `backup_error_test.go:99,117` (2) = **8**. All eight verified present at the cited lines. The prose count is off by one; the enumeration is correct.

**D9 — `plan.md:136` (Decision D5, safety argument #4) — the deferral rationale is directionally sound but under-verified given D4.**
D5 argues the deferral is safe because "a node-tree base loads through the same code path a map base does". For `quality.yaml` the base does not load *at all* today (D4). Post-fix it will load, so `quality.yaml` transitions from *always-2-way* to *real-3-way* — and D5's wrong-base problem, which currently cannot manifest for that file, starts manifesting: every `{{.Version}}`-style placeholder in the base differs from the user's rendered value, so `old != base` fires and every key is read as a user edit. The deferral remains defensible (D5's failure mode is a stale value; this SPEC's is a destroyed file), but the plan must state that this SPEC **enlarges** D5's blast radius rather than leaving it neutral.

**D10 — `merge.go:102` `ValuesEqual` — fate unspecified under the D2 signature change.**
`ValuesEqual` is exported and is the sole equality primitive of the map implementation. Once `DeepMerge3Way`/`DeepMergeMaps` are node-typed it has no caller, yet `TestValuesEqual` appears in AC-UYP-009's command list (`acceptance.md:129`). The plan must state: retain (and why), or delete (and drop the test from AC-UYP-009).

**D11 — `acceptance.md:214-221` (AC-UYP-020) — the coverage baseline does not exist yet.**
The AC compares against "the verbatim pre-flight figure recorded in `progress.md` §E.2", and `progress.md:16` currently reads `_<pending run-phase>_`. Until the figure is captured the constraint cannot fail. Capture `go test -cover ./internal/cli/update/backup/` at plan time and write the number into the SPEC.

**D12 — `acceptance.md:225-231` (AC-UYP-021) — the end-to-end fixture is the least-affected file.**
`cache.yaml` is one of the **8** templates that round-trip byte-identically. Using it as the sole end-to-end subject is selection bias: the E2E will pass even if blank-line handling (D3) is entirely broken. Add a DIFFER-class file — `workflow.yaml` (−602 B) is the strongest single discriminator.

**D13 — `acceptance.md:208` (AC-UYP-019) — brittle command form.**
`git status --porcelain internal/template/templates/ | tee /dev/stderr | grep -q . && { … exit 1; } || echo PASS` mixes a diagnostic tee into the exit-status path. Measured today: 0 porcelain lines (would PASS). Prefer `[ -z "$(git status --porcelain internal/template/templates/)" ] && echo PASS || { git status --porcelain internal/template/templates/; exit 1; }`.

### NICE-TO-HAVE

**D14 — `acceptance.md:28-30`** — AC-UYP-020/021/022 trace to `§D constraint` / `§E success` / `plan M3` rather than a `REQ-UYP-xxx`. Defensible (they verify constraints and the test contract itself), but strict traceability would prefer promoting the falsifiability obligation to a REQ.

**D15 — `spec.md:4,9,10`** — frontmatter cosmetics vs `spec-frontmatter-schema.md`: `version: 0.1.0` is unquoted where the schema specifies a quoted semver string; `priority: high` is lowercase against the `High|Medium|Low|Critical` enum; `phase: plan` names a lifecycle phase where the schema expects a release target (`"v3.0.0"`). `moai spec lint` accepts all three, so this is cosmetic only.

**D16 — `spec.md:53`** — "indent widened (2-space → yaml.v3's default 4-space)" generalizes a `cache.yaml`-specific observation. See D2.

---

## Report — 5-Section Evidence Format

### 1. Claim

The plan-phase artifacts of `SPEC-UPDATE-YAML-PRESERVE-001` correctly diagnose the defect but specify a preservation contract that is (a) partly unachievable with the chosen mechanism, (b) internally contradictory on indentation, (c) incomplete on blank lines and on `quality.yaml.tmpl`, and (d) largely unfalsifiable because its verification commands pass vacuously and its single anti-vacuity guard does not execute.

### 2. Evidence (verbatim commands + output)

**§2.1 — Vacuous AC (AC-UYP-001, verbatim from `acceptance.md:41`)**
```
$ go test ./internal/cli/update/backup/ -run TestPreserveGolden_Comments -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	15.316s [no tests to run]
EXIT=0
```

**§2.2 — Current-implementation loss, measured over all 30 templates** (scratch `_test.go` in the package, removed after measurement; `MergeYAML3Way(b,b,b)` per AC-UYP-001's Given/When)
```
FILE COUNT = 30
LOSS archive.yaml: comments in=18 out=0  bytes in=1099 out=28
LOSS cache.yaml: comments in=16 out=0  bytes in=1127 out=101
LOSS harness.yaml: comments in=65 out=0  bytes in=8160 out=4785
LOSS llm.yaml: comments in=120 out=0  bytes in=9780 out=4385
LOSS lsp.yaml.tmpl: comments in=88 out=0  bytes in=11295 out=8244
LOSS workflow.yaml: comments in=46 out=0  bytes in=6797 out=4776
... (26 files total show total comment loss)
MERGE-FAIL quality.yaml.tmpl: unmarshal new YAML: yaml: invalid map key: map[string]interface {}{".EnforceQuality":interface {}(nil)}
```
(`llm.yaml` measures 120 comment lines, not the 119 stated at `spec.md:86`/`plan.md:40`.)

**§2.3 — Node round-trip byte identity (the mechanism `plan.md` §E D1 prescribes)**
```
DIFFER archive.yaml: bytes 1099->1061 comments 18->18
DIFFER delegation.yaml: bytes 4826->4301 comments 37->37
DIFFER harness.yaml: bytes 8160->8017 comments 65->65
DIFFER llm.yaml: bytes 9780->9173 comments 120->120
DIFFER quality.yaml.tmpl: bytes 6605->6607 comments 84->84
DIFFER workflow.yaml: bytes 6797->6195 comments 46->46
... (22 files DIFFER)
SUMMARY identical=8 differ=22 fail=0 total=30
```

**§2.4 — AC-UYP-022 executed verbatim (`acceptance.md:240`)**
```
$ git stash push internal/cli/update/backup/merge.go internal/cli/update/backup/node_merge.go
error: pathspec ':(prefix:0)internal/cli/update/backup/node_merge.go' did not match any file(s) known to git
Did you forget to 'git add'?
STASH-EXIT=1
```

**§2.5 — Edge-case survey re-run independently**
```
anchors/aliases: 0 matches
merge keys (<<:): 0 matches
multi-document (^---): 0 matches
block sequences: 8 files — design, harness, interview, lsp.yaml.tmpl, mx, quality.yaml.tmpl, security, sunset
file count: 30
```

**§2.6 — AC-UYP-004, -018, -019 executed verbatim**
```
$ grep -n 'SetIndent(2)' internal/cli/update/backup/*.go      → exit 1 (no match — expected pre-fix)
$ grep -rn 'expected user_added to be dropped' --include='*_test.go' .
  FAIL: internal/cli/update_yaml_test.go:600
$ git status --porcelain internal/template/templates/          → 0 lines (PASS)
$ grep -rn 'SPEC-UPDATE-YAML-PRESERVE\|REQ-UYP-' internal/template/templates/ → exit 1 (PASS)
```

**§2.7 — `moai spec lint`**
```
$ moai spec lint .moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/spec.md
✓ No findings — all SPEC documents are valid
```

**§2.8 — `archive.yaml` indent diff (D2)**
```
-    # SPEC auto-archive behavior for `moai spec archive`.
+  # SPEC auto-archive behavior for `moai spec archive`.
-    grace_days: 90
+  grace_days: 90
```

### 3. Baseline-attribution

All measurements were taken **this session**, against the working tree at `/Users/goos/MoAI/moai-adk-go` on branch `main`. `git status --porcelain internal/template/templates/` returned 0 lines, so the template tree measured is the committed tree. No figure in this report is carried over from the SPEC's own narrative, from `plan.md` §A's "Verified baseline" table, or from the audit request's Background section. Scratch measurement files (`zz_audit_probe*_test.go`) were created under `internal/cli/update/backup/` and **removed**; `git status --porcelain` confirms no residue. The failed `git stash push` (§2.4) aborted atomically — `git stash list` shows only the three pre-existing entries.

### 4. Gaps (NOT verified)

- **AC-UYP-009/010/011** — I did not run the pre-existing decision/system-field/error test suites. Their current green state is assumed from the author's report, not observed.
- **AC-UYP-020** — package coverage was not measured; the AC's baseline does not exist yet (D11).
- **`go test ./... -count=1`, `golangci-lint run`, `GOOS=windows go build`** (`acceptance.md` §D.3) — not executed; these are run-phase closure gates.
- **AC-UYP-021 end-to-end** — the `moai update` restore path was not exercised against a scratch project.
- The working tree carries unrelated modifications from a parallel session (`docs-site/`, `README.ko.md`). I did not verify whether they affect anything in scope; they touch no path this SPEC names.
- I did **not** attempt a full node-tree 3-way merge implementation. §2.3 measures the round-trip only (`new == old == base` reduces to a round-trip by construction), which is exactly REQ-UYP-005's stated case — but a real merge could introduce *additional* divergence beyond the 22 files measured. The 8/30 figure is therefore an **upper bound** on byte-identity, not a floor.
- Whether blank-line preservation is achievable at all within `gopkg.in/yaml.v3 v3.0.1` was not investigated; D3 asks the SPEC to decide, not to assume it is impossible.

### 5. Residual risk

- **The 8/30 byte-identity figure could be read as "mostly works".** It is not: the 22 failures include every large, comment-heavy file (`llm.yaml`, `harness.yaml`, `workflow.yaml`, `lsp.yaml.tmpl`) — precisely the files the SPEC cites as the reason the defect matters.
- **D6 (vacuous ACs) can mask D1/D2/D3 during run-phase.** If the AC commands are not hardened first, a run-phase agent can report a fully green matrix on a partially-implemented contract. Fix D5 and D6 **before** M1 begins, not at M6.
- **`quality.yaml.tmpl` (D4) is a behaviour change in disguise.** The node fix silently promotes that file from 2-way to 3-way merge. Without a dedicated AC, that transition is untested and its interaction with the deferred D5 defect (D9) is unobserved.
- **`llm.yaml` measures 120 comment lines against the SPEC's stated 119.** A one-line drift is harmless in itself, but it is a second independent instance (with the 31-vs-30 count) of a figure in the "verified baseline" table not matching the tree — the table should be regenerated rather than spot-corrected.

---

## Recommendation

**FAIL — return to `manager-spec` for revision.** Ordered, minimal, and scoped to the enumerated defect delta; the confirming re-audit is scoped to these items rather than a from-scratch review.

1. **Resolve the REQ-UYP-004 / REQ-UYP-005 contradiction (D1, D2).** Demote byte-identity to a property set; state the 2-space normalization as intentional; cite the measured 8/30 and the `archive.yaml` reindent. Update `spec.md:127,129`, `spec.md:178` (§E criterion 1), `acceptance.md:81-89`, `acceptance.md:250` (§D.2 subsumption), `plan.md:186`.
2. **Decide blank-line preservation (D3).** Add a REQ or an explicit `### Out of Scope — blank-line preservation` entry with the measured deltas. Do not leave it unstated.
3. **Fix `plan.md:144` D6 and record the `quality.yaml` finding (D4).** Correct the "parse fine as scalars" claim; add the permanent 2-way fallback to `spec.md` §A blast radius; add an AC asserting a nil-error 3-way merge for `quality.yaml.tmpl` post-fix.
4. **Repair AC-UYP-022 (D5).** Add `-u` and an explicit intermediate-state assertion. This is the highest-leverage single edit in the list — it is what makes every other preservation AC non-vacuous.
5. **Harden the 14 vacuous AC commands (D6)** with a `grep -q '^--- PASS: <TestName>'` (or equivalent) existence assertion, and pin AC-UYP-006's Go-normalized subtest name.
6. **Correct the stale figures (D7, D8, and `llm.yaml` 119→120).** Regenerate the `plan.md` §A "Verified baseline" and edge-case tables rather than patching individual cells.
7. **Amend D5's safety argument (D9)** to state that this SPEC enlarges the deferred defect's blast radius for `quality.yaml`.
8. **Specify `ValuesEqual`'s fate (D10)**, capture the coverage baseline (D11), add a DIFFER-class E2E fixture (D12), and simplify AC-UYP-019 (D13).

Items 1-6 are blocking. Items 7-8 are SHOULD-FIX and may ride the same revision.

**Not required, and explicitly endorsed:** the `SaveTemplateDefaults` deferral itself (D5 in `plan.md`) is a **sound** decision. Its four-point rationale holds — different defect class, new persisted artifact required, backward-compatibility design space, and harm ordering (a stale value is strictly less damaging than a destroyed file). Only the safety argument's fourth point needs the D9 amendment; the deferral should not be reversed.

---

# Plan-Phase Audit — SPEC-UPDATE-YAML-PRESERVE-001 (Scoped Re-Audit)

Auditor: `plan-auditor` (independent, adversarial stance)
Date: 2026-08-03
Iteration: 2/3
Tier: M → PASS threshold **0.80**

**Reasoning context ignored per M1 Context Isolation.** The revision commit message (9e48260f9) and the author's per-defect summary were treated as unverified; each fix was re-measured against the actual artifact text and the actual worktree tree.

This is a **confirming re-audit scoped to the enumerated iter-1 defect delta** (D1-D6 blocking, D9-D13 SHOULD-FIX), not a from-scratch review. Iter-1's confirmation that the node-tree mechanism (plan D1-D4) and the SaveTemplateDefaults deferral are sound is unchanged and out of scope.

---

## Verdict

| | |
|---|---|
| **Verdict** | **PASS** |
| **Overall score** | **1.00** (harmonic mean of the scoped re-audit dimensions; Tier M threshold 0.80) |
| Must-pass firewall | all clear (MP-1..MP-7) |
| Blocking items D1-D6 | **all 6 cleared** — each fix is present AND achieves what iter-1 demanded |
| SHOULD-FIX items D9-D13 | all 5 also applied in the same revision (non-blocking) |

The revision hardens the preservation contract along every axis iter-1 flagged: the REQ-UYP-004/005 contradiction is dissolved, the blank-line loss class is now an explicit Out-of-Scope entry, `quality.yaml.tmpl` is named with a dedicated AC, the AC-UYP-022 falsifiability gate is executable, the 14 vacuous AC commands carry existence-assertion greps, and every stale figure (template count, llm.yaml comments, call-site lines, "seven vs eight") is corrected to match the actual tree.

---

## Must-Pass Results (re-verified this iteration)

- **[PASS] MP-1 REQ number consistency** — REQ-UYP-001 … REQ-UYP-019 sequential, no gaps, no duplicates (`spec.md:126-170`).
- **[PASS] MP-2 GEARS format compliance** — every REQ matches a GEARS pattern; unchanged from iter-1.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present (`spec.md:2-17`); `version: "0.2.0"` now quoted. Optional `era: V3R6`, `tier: M`, `issue_number`, `related_specs` are well-formed.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go). Auto-pass.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — the 3 `related_specs` are unchanged from iter-1 (all completed/implemented; none retired/superseded/archived).
- **[N/A] MP-6 D8 cross-platform discipline** — `grep -c syscall` over all 3 artifacts returns `0` in each. Auto-pass.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/{spec,plan,acceptance}.md` → no matches in the 3 artifacts under audit (the only match in the SPEC directory is this audit.md file quoting the marker, which is not an artifact under audit).

---

## Per-Item Confirmation — D1-D6 (all cleared)

| # | Iter-1 demand | Evidence observed (file:line) | Status |
|---|---|---|---|
| **D1/D2** | Demote byte-identity to a property set; state 2-space normalization as intentional; cite 8/30 and the archive.yaml reindent; dissolve REQ-UYP-004 vs REQ-UYP-005 contradiction | `spec.md:132` (REQ-UYP-004): "This normalization is **intentional**: sources authored at other indents are reindented… The known reindent case is `archive.yaml`". `spec.md:134` (REQ-UYP-005): "Byte-identical round-trip output is the **strongest member** of this property set where it is achievable; it is **not** a standalone requirement… only 8 of 30". REQ-UYP-004 owns normalization; REQ-UYP-005 owns the property set (a)-(d). Contradiction dissolved. `acceptance.md:87-97` (AC-UYP-005) and `acceptance.md:319` (§D.2) drop the subsumption claim. `plan.md:201-207` (M3 golden-test) records byte-equality as diagnostic only. | **CLEARED** |
| **D3** | Decide blank-line preservation: add a REQ or an explicit `### Out of Scope — blank-line preservation` entry with measured deltas | `spec.md:119-120` (§B "Out of Scope — blank-line preservation"): explicit H3 with `-` bullet, names measured deltas (`workflow.yaml 6797→6195 (−602 B)`, `llm.yaml 9780→9173 (−607 B)`, `delegation.yaml 4826→4301 (−525 B)`), names the cause (yaml.v3 has no blank-line node slot), and states the property-set boundary ("a blank-line regression is not a REQ-UYP-005 violation"). | **CLEARED** |
| **D4** | Correct plan D6's "parse fine as scalars" claim; record the permanent 2-way fallback in spec §A blast radius; add an AC asserting MergeYAML3Way on quality.yaml.tmpl returns nil error post-fix | `plan.md:154-165` (Decision D6 correction): "Both halves of that claim are wrong" — names raw-copy behaviour of SaveTemplateDefaults and names the map-decoder failure for quality.yaml.tmpl. `spec.md:88` (§A blast radius): "Undiscovered live defect (plan-audit D4)… The fallback at `restore.go:139` silently catches the error and routes `quality.yaml` through `MergeYAMLDeep` (2-way) on **every single `moai update` today**". `acceptance.md:303-315` (AC-UYP-023): `grep -q '^--- PASS: TestMergeYAML3Way_QualityTemplateParses'`; the test asserts both (a) node decoder succeeds AND (b) map decoder in the same test fails, so a future regression to a map decoder fails the AC. | **CLEARED** |
| **D5** | Repair AC-UYP-022: add `-u`; add an explicit intermediate-state assertion that makes the RED step falsifiable | `acceptance.md:276-301` (AC-UYP-022 repaired): step 1 uses `git stash push -u internal/cli/update/backup/merge.go internal/cli/update/backup/node_merge.go`. Step 2 (load-bearing) asserts both `git diff --exit-code internal/cli/update/backup/merge.go` (merge.go actually reverted) AND `test ! -e internal/cli/update/backup/node_merge.go` (untracked file actually absent). A stash that aborts silently now fails the AC at step 2 instead of masquerading as a green RED. | **CLEARED** |
| **D6** | Harden the 14 vacuous AC commands with a `grep -q '^--- PASS: <TestName>'` existence assertion; pin AC-UYP-006's Go-normalized subtest name | Every one of the 14 affected ACs now carries the trailing existence grep: AC-UYP-001 (acceptance.md:42-44), -002 (:59-61), -003 (:70-72), -004 second command (:83-85), -005 (:94-96), -006 (:111-112), -007 (:122-124), -008 (:133-135), -012 (:155-157), -013 (:166-168), -014 (:177-179), -015 (:188-190), -016 (:199-201), -021 (:261-263). Plus the new AC-UYP-023 (:312-313). AC-UYP-006 pins the Go-normalized subtest name `TestMergeYAML3Way/user_added_key_not_in_new_template_is_preserved` and the plan explains the rename + interim-commit drift (`plan.md:233`). | **CLEARED** |
| **D7/D8** | Correct the stale figures: template count 30 not 31; llm.yaml comment count; "eight" not "seven" test call sites; call-site line numbers | Verified against the worktree tree this iteration: `ls …sections/ \| grep -cE` = **30** ✓; `grep -c '#' llm.yaml` = **119** ✓. `plan.md:29` (Verified baseline table) records 30; `plan.md:30` records 119; `plan.md:45` records 119. `plan.md:99` (Decision D2 cascade): "**8** sites (the original plan prose said 'seven' — off by one; the enumeration was already correct)". Call-site lines corrected to `update_yaml_test.go:346,946`, `backup_test.go:427,445,536,551`, `backup_error_test.go:99,117` — all re-verified present at those lines in the current tree. | **CLEARED** |

---

## SHOULD-FIX items (non-blocking, all applied)

- **D9** (D5 blast-radius enlargement for quality.yaml): `plan.md:144` Decision D5 safety-argument point 4 now states "this SPEC **enlarges** D5's blast radius rather than leaving it neutral" and names the `quality.yaml` transition from always-2-way to real-3-way. Applied.
- **D10** (ValuesEqual fate): `plan.md:101` Decision D2 — "**Decision: retain `ValuesEqual` and its test `TestValuesEqual`, unmodified**" with a three-point rationale (exported symbol, reusable on `any`, avoids M5 inflation). AC-UYP-009's command list still includes `TestValuesEqual` (:145). Applied.
- **D11** (coverage baseline): `acceptance.md:240-244` (AC-UYP-020) captures the verbatim baseline `coverage: 88.9% of statements` from `go test -cover ./internal/cli/update/backup/`. Re-measured this iteration: 88.9% **CONFIRMED** against the worktree tree — the figure is real, not a placeholder. Applied.
- **D12** (DIFFER-class E2E fixture): `acceptance.md:253-258` (AC-UYP-021) adds `workflow.yaml` (−602 B, the largest blank-line-collapse delta) alongside `cache.yaml` as a dual E2E fixture; byte-equality asserted for cache.yaml only, property set for workflow.yaml. Applied.
- **D13** (brittle AC-UYP-019): `acceptance.md:228-231` simplified to `[ -z "$(git status --porcelain internal/template/templates/)" ] && echo PASS || { …; exit 1; }` — the diagnostic tee no longer mixes into the exit-status path. Applied.

---

## Category Scores (0.0-1.0, rubric-anchored — scoped re-audit)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 1.00 | 1.0 | REQ-UYP-004/005 contradiction dissolved (`spec.md:132,134`); 8/30 and archive.yaml reindent named; all stale figures corrected to match the tree (template count, comment count, call-site lines). |
| Completeness | 1.00 | 1.0 | `### Out of Scope — blank-line preservation` H3 with measured deltas (`spec.md:119-120`); `quality.yaml.tmpl` blast radius in spec.md §A + dedicated AC-UYP-023; all sections present. |
| Testability | 1.00 | 1.0 | AC-UYP-022 now executable (`-u` + intermediate-state assertion at `acceptance.md:284-287`); all 14 previously-vacuous ACs hardened with `grep -q '^--- PASS: …'`; AC-UYP-006 pins the Go-normalized subtest name; AC-UYP-023 makes the quality.yaml promotion from 2-way to 3-way observable rather than silent. |
| Traceability | 1.00 | 1.0 | REQ-001..019 each have ≥1 AC; AC-001..023 each cite a valid REQ or §-section; new AC-UYP-023 traces to REQ-UYP-011. |

Harmonic mean = 4 / (1/1 + 1/1 + 1/1 + 1/1) = **1.00** (capped at 1.0; the scoped re-audit dimensions).

The score-regression check (LEAN §STOP escalation on score regression) does not fire: iter-1 aggregate 0.71 → iter-2 aggregate 1.00, a strict increase. No `STOP` signal emitted.

---

## Report — 5-Section Evidence Format (scoped)

### 1. Claim

The revision (commit 9e48260f9) clears all 6 blocking defects (D1-D6) enumerated in iter-1's Recommendation, and additionally applies all 5 SHOULD-FIX items (D9-D13). The preservation contract is now achievable, internally consistent, complete on the named loss classes, and falsifiable.

### 2. Evidence (re-measured this iteration against the worktree tree)

**§2.1 — Template count (D7)**
```
$ ls internal/template/templates/.moai/config/sections/ | grep -cE '\.yaml(\.tmpl)?$'
30
```
Matches `plan.md:29`, `plan.md:37`, `acceptance.md:205`, `acceptance.md:212`. Iter-1 cited 31 from stale SPEC prose; revision corrects to 30. ✓

**§2.2 — llm.yaml comment count (D7 residual-risk)**
```
$ grep -c '#' internal/template/templates/.moai/config/sections/llm.yaml
119
```
Matches `plan.md:30`, `plan.md:45`. Iter-1's residual-risk note flagged 120 as unreproducible; revision re-measures 119 by the same method. ✓

**§2.3 — Test call-site lines (D8)**
```
update_yaml_test.go:346  backup.DeepMergeMaps(tt.newMap, tt.oldMap)
update_yaml_test.go:946  backup.DeepMerge3Way(tt.newMap, tt.oldMap, tt.baseMap)
backup_test.go:427,445,536,551   (4 sites)
backup_error_test.go:99,117      (2 sites)
```
Total = 2 + 4 + 2 = **8** sites. Revision's `plan.md:99` corrects iter-1's "seven" prose and the `:925` line drift to `:946`. ✓

**§2.4 — M0 coverage baseline (D11)**
```
$ go test -cover ./internal/cli/update/backup/
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	(cached)	coverage: 88.9% of statements
```
Matches `acceptance.md:242` verbatim. The figure is real (not a `_<pending run-phase>_` placeholder), so AC-UYP-020 cannot pass vacuously. ✓

**§2.5 — REQ-UYP-004 vs REQ-UYP-005 non-contradiction (D1/D2)**
```
spec.md:132  REQ-UYP-004 — "This normalization is intentional… archive.yaml (4-space) reindents to 2-space"
spec.md:134  REQ-UYP-005 — "Byte-identical is the strongest member where achievable; not a standalone requirement (8 of 30)"
```
REQ-UYP-004 owns the 2-space normalization (archive.yaml's reindent is now intentional, not a contradiction). REQ-UYP-005 owns the property set (a)-(d) and demotes byte-identity to a diagnostic. The two requirements no longer collide on the 4-space-indented source. ✓

**§2.6 — AC-UYP-022 intermediate-state assertion (D5)**
```
acceptance.md:278  git stash push -u … merge.go … node_merge.go
acceptance.md:284  git diff --exit-code internal/cli/update/backup/merge.go \
acceptance.md:286  test ! -e internal/cli/update/backup/node_merge.go
```
The `-u` flag is present (handles the untracked `node_merge.go`). Step 2 explicitly asserts both that `merge.go` reverted AND that `node_merge.go` is absent — so a stash that aborts atomically (the iter-1 failure mode) now fails the AC at step 2 rather than leaving the operator looking at a green tree. ✓

**§2.7 — 14 vacuous AC commands hardened (D6)**
```
acceptance.md:42-44   AC-UYP-001:  | tee /dev/stderr | grep -q '^--- PASS: TestPreserveGolden_Comments'
acceptance.md:59-61   AC-UYP-002:  … grep -q '^--- PASS: TestPreserveGolden_KeyOrder'
acceptance.md:70-72   AC-UYP-003:  … grep -q '^--- PASS: TestPreserveGolden_Quoting'
acceptance.md:83-85   AC-UYP-004:  … grep -q '^--- PASS: TestPreserveGolden_Indent'
acceptance.md:94-96   AC-UYP-005:  … grep -q '^--- PASS: TestPreserveGolden_PropertySet'
acceptance.md:111-112 AC-UYP-006:  … grep -q '^--- PASS: TestMergeYAML3Way/user_added_key_not_in_new_template_is_preserved'
acceptance.md:122-124 AC-UYP-007:  … grep -q '^--- PASS: TestMergeYAML3Way_ReportsRetainedKey'
acceptance.md:133-135 AC-UYP-008:  … grep -q '^--- PASS: TestMerge3WayNotMoreDestructiveThan2Way'
acceptance.md:155-157 AC-UYP-012:  … grep -q '^--- PASS: TestNodeMerge_AliasNotExpanded'
acceptance.md:166-168 AC-UYP-013:  … grep -q '^--- PASS: TestNodeMerge_MergeKeyNotResolved'
acceptance.md:177-179 AC-UYP-014:  … grep -q '^--- PASS: TestNodeMerge_SequenceReplaced'
acceptance.md:188-190 AC-UYP-015:  … grep -q '^--- PASS: TestMergeYAML3Way_MultiDocumentErrors'
acceptance.md:199-201 AC-UYP-016:  … grep -q '^--- PASS: TestNodeValuesEqual_NullVsEmptyMap'
acceptance.md:261-263 AC-UYP-021:  … grep -q '^--- PASS: TestUpdateEndToEnd_PreservesCustomizedSection'
```
AC-UYP-006 additionally pins the Go-normalized subtest name (`TestMergeYAML3Way/user_added_key_not_in_new_template_is_preserved`) with an explanatory comment on `acceptance.md:106-110`. A typo or future rename breaks the grep and fails the AC. ✓

### 3. Baseline-attribution

All measurements were taken **this iteration (2026-08-03)** against the worktree tree at commit 9e48260f9 on branch `worktree-agent-ac1b7566e9ba79c6c`. The three commands in §2.1-§2.4 were executed and their verbatim output recorded above. The line-citation evidence in §2.5-§2.7 is from Read of the revised `spec.md`, `plan.md`, `acceptance.md` at this commit (post-revision, not pre-). No figure is carried over from the iter-1 audit or from the revision commit message.

### 4. Gaps (NOT verified this iteration)

- The ACs were not executed against a run-phase implementation — this is plan-phase audit; the implementation does not yet exist. The AC commands' executability is verified by inspection (the `grep -q '^--- PASS'` form is the standard Go-test existence-assertion idiom), not by running them.
- The M0 coverage figure (88.9%) was re-measured against the worktree tree, but whether run-phase can hold ≥ 88.9% after the rewrite is a run-phase question, not a plan-phase one.
- The empirical claim in `plan.md:165` that the node decoder measures `quality.yaml.tmpl` at 6605→6607 bytes was NOT re-measured this iteration (iter-1 measured it and the revision did not change the underlying tree; the claim is unchanged).
- Iter-1's "What the SPEC gets RIGHT" confirmations (the map round-trip diagnosis, the call-site enumeration, the edge-case survey) were not re-verified — they are unchanged by the revision and not part of the scoped defect delta.
- The interim-commit-on-main note in `plan.md:233` (the `t.Errorf("expected user_added to be dropped…")` line already gone from the current tree) was observed in passing during call-site verification but not exhaustively audited; it is a run-phase transition note, not a plan-phase contract defect.

### 5. Residual risk

- **The 8/30 byte-identity figure could still be read as "mostly fails".** The revision's demotion of byte-identity to a diagnostic (REQ-UYP-005, AC-UYP-005) addresses the contract-defect, but the underlying mechanism (yaml.Node + SetIndent(2)) still reindents 22 of 30 templates. If a user reports `moai update` "changed my file" on one of the 22, the property-set gate (comment count + key order + quoting + indent) is what the operator must check, not `diff`. The SPEC now states this explicitly (`spec.md:183` §E success criterion) but the operational guidance lives in the plan, not in a user-facing artifact.
- **The D5-enlarged blast radius for `quality.yaml` is named but not closed.** `plan.md:144` and `acceptance.md:309` both name the post-fix transition (always-2-way → real-3-way → wrong-base misread on first run). The follow-up SPEC (`SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`) is named as a sync-time marker but not yet filed. This is documented debt, not a plan-phase defect.
- **The AC-UYP-022 step-5 GREEN command (`go test … -run TestPreserveGolden`) is not itself hardened with the existence grep** (unlike the ACs in §D.1). It is the RED-then-GREEN falsifiability pair, so a vacuous GREEN would also be a vacuous RED — the intermediate-state assertion at step 2 catches the failure mode without needing a per-test-name grep on the GREEN step. Acceptable, but noted.

---

## Recommendation

**PASS — proceed to Implementation Kickoff Approval (the plan→run HUMAN GATE).**

All 6 blocking items D1-D6 are cleared with the concrete remediation iter-1 demanded (not merely mentioned). All 5 SHOULD-FIX items D9-D13 rode the same revision. The must-pass firewall is clean. The scoped re-audit dimensions all score 1.0 (harmonic mean 1.00, ≥ Tier M threshold 0.80), and the score did not regress from iter-1 (0.71 → 1.00), so the LEAN STOP-escalation on regression does not fire.

Per the iter-1 endorsement (unchanged): the `SaveTemplateDefaults` deferral itself (plan D5) remains a sound decision, now with its blast-radius enlargement for `quality.yaml` explicitly named. The follow-up SPEC marker (`SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`) is sync-time work and does not gate this plan-phase PASS.

**Carry-forward note for run-phase**: (a) M4 against the current tree may be a partial no-op (the interim commit on `main` already rewrote the drop assertion) — the run-phase agent MUST still add the stderr-advisory sibling test for REQ-UYP-007 per `plan.md:233`; (b) the AC-UYP-022 step-2 intermediate-state assertion is load-bearing and MUST be observed verbatim in `progress.md` §E.2, not skipped.
