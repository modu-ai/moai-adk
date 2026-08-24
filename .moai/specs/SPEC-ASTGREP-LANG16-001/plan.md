# SPEC-ASTGREP-LANG16-001 — Implementation Plan

> Companion to `spec.md`. Requirement IDs (`REQ-A16-*`) and acceptance IDs (`AC-A16-*`) are
> defined there and in `acceptance.md`; this document does not restate them.

## §A. Context

### A.1 Where the work happens

| field | value |
|---|---|
| worktree | `.claude/worktrees/t228` |
| branch | `WT-astgrep-16-langs` |
| HEAD | `294b4b6ab` |
| tool | `ast-grep 0.40.5` (`/opt/homebrew/bin/sg`) |
| card | t228, Class C, Tier L |

Route B (PR) per the repo-local policy: `main` is protected with `enforce_admins: true`, so every
tier lands through a PR. No direct push to `main` at any milestone.

### A.2 The split, and why this SPEC is the first half

Plan-audit iteration 1 returned FAIL 0.68 with three must-pass failures. The decisive finding was
not any single defect but a budget fact: requirements and criteria were saturated at 25/25, so the
audit's own corrections could not be added without breaching the Tier L ceiling. A SPEC that
cannot absorb its own audit findings is over-budget by definition.

The seam taken is the one the previous revision already named:

| | SPEC | Carries |
|---|---|---|
| First | **SPEC-ASTGREP-LANG16-001** (this) | M1-M3 — the contract: `sg test` harness, coverage matrix + checker, severity reclassification over the existing 26 rules |
| Second | **SPEC-ASTGREP-BREADTH-001** | M4-M10 — up to 80 rules across ten languages, per-language idiom rules, differential-corpus fixture pairs |

**Total scope is unchanged.** Card t228 requirements (2) and (4) are delivered in full, across two
SPECs in sequence, with no external gate between them. This is sequencing, not trimming.

Two things make contract-first the correct order rather than a convenience:

1. **Two of this SPEC's defenses were falsified during audit** (`spec.md` §A.8) — R2's
   anti-contrivance mitigation and the corpus gate's forcing function. Correcting them cost one
   requirement each. Had 80 rules already been written against those defenses, the same correction
   would have invalidated 80 rules.
2. **The implement-vs-exempt split across the 80 cells is unknown at plan time.** The feasibility
   probe covered 6 patterns of 80 and says so; `research.md` Q2 defers per-cell resolution to the
   breadth milestones. Sizing the breadth half honestly requires the matrix this SPEC builds.

### A.3 Primary write surface

```
internal/template/templates/.moai/config/astgrep-rules/
├── sgconfig.yml     # EDIT   — testConfigs pointing OUTSIDE this tree (REQ-A16-019)
├── security/*.yml   # EDIT   — severity pass + metadata.cwe anchors
└── go/*.yml         # EDIT   — severity pass only

<repo-side rule-tests root>       # CREATE — sg test fixtures + snapshots, NOT under the template
.moai/specs/SPEC-ASTGREP-LANG16-001/coverage-matrix.md   # CREATE — the 112-cell matrix
internal/hook/astgrep_corpus_pin_test.go                 # CREATE — the only new file under internal/hook/
```

The matrix lives under `.moai/specs/`, not under the template tree, so its rationale prose cannot
violate neutrality (REQ-A16-021).

**No embed artifact is produced.** `make build` is run so the binary carries the template edit, but
it emits nothing the commit must carry — see §D.

### A.4 PRESERVE list

Do not modify, at any milestone:

- **Every existing file under `internal/hook/**`** — including `pre_tool_scan_differential_test.go`
  and every file under `internal/hook/security/testdata/scan-corpus/`. This SPEC reads the deny
  wiring (`spec.md` §A.5) and the corpus gate (§A.7); it changes neither.
  - **Carve-out, exactly one file.** M1 adds `internal/hook/astgrep_corpus_pin_test.go` — a new
    file, never an edit to an existing one. It is the only permitted addition under this tree, and
    the §E standing diff command excludes it by path so the untouchedness proof stays exact rather
    than being widened. AC-A16-007 and AC-A16-008 are what that test asserts; AC-A16-008 pins the
    twelve corpus fixtures and their `wantDeny` values byte-identical to `294b4b6ab`.
  - The test is placed here rather than outside the tree because what it pins is package-internal:
    the `t.Skip` position relative to the assertion loop, and the fixture table the harness reads.
    A pin sitting outside the package would assert against a copy, which is the thing it exists to
    detect drifting.
- `.moai/astgrep-rules/**` — the local dogfood tree (REQ-A16-020)
- `.moai/config/sections/gate.yaml` and its template mirror
- every other SPEC directory under `.moai/specs/`

### A.5 Pre-flight hygiene

Before M1, remove the stray empty directory tree
`internal/template/templates/.moai/config/astgrep-rules/security/.claude/agent-memory/manager-spec/`.
It is untracked (git does not track empty directories, which is why `git status` reads clean) and
is physical residue of the cwd-drift incident `research.md` §8 records — sitting inside the exact
directory this SPEC edits. If any file is ever written there it ships to every user.

```bash
rmdir -p internal/template/templates/.moai/config/astgrep-rules/security/.claude/agent-memory/manager-spec 2>/dev/null || true
```

### A.6 Milestone ordering rationale

Three milestones, ordered by decision-reversibility. M1 fixes a mechanism (how rules are tested,
how ids are keyed, where test assets live), M2 fixes the scope contract (what "covered" means and
what evidence an exemption owes), M3 changes user-visible behaviour (which rules refuse a write).

Every one is cheap to redo now and expensive to redo once the successor writes 80 rules against
it. This SPEC contains no mechanical-breadth milestone at all — that is the point of the split,
and it means a reviewer's attention is well spent on all three.

---

## §B. Milestones

### M1 — `sg test` harness, id keying, and asset placement

**Priority: High. Blocks M2 and M3, and blocks the successor SPEC entirely.**

Work:

1. Run the §A.5 pre-flight `rmdir`.
2. Decide the rule-tests root **outside** the template tree (REQ-A16-019) and wire
   `testConfigs` in `sgconfig.yml` to point at it repo-side.
   - The decision is forced: everything under the template tree deploys into every user project,
     and by REQ-A16-008 these fixtures contain deliberately dangerous constructs. Shipping them
     would place our test assets inside the very tree the deployed ruleset scans, so a user's own
     pre-write gate would fire on them.
   - Verify the deployed `sgconfig.yml` remains valid for a user who has no rule-tests directory:
     `sg scan --config` must still work where `testConfigs` points at an absent path.
3. Create the rule-tests tree with a `valid`/`invalid` pair for each of the existing 26 rules, and
   generate the snapshot baseline.
4. **Settle R1 (rule-id collision).** The shipped convention repeats one id across languages —
   `sec-hardcoded-credential` appears four times with four `language:` values. Determine
   empirically whether `sg test` keys snapshots by id alone or by id+language.
   - id-alone → switch to per-language suffixed ids and record the decision in `design.md`. This
     renames existing rules, which is exactly why it is settled at M1.
   - id+language → keep the convention unchanged.
5. Reproduce **both** mutants on this ruleset, not only in a scratch directory: rewrite one rule
   to match nothing (expect `[Missing]`, non-zero exit), and one to match its own `valid` case
   (expect `[Noisy]`, non-zero exit). Revert both.
6. **Write the corpus pinning test** at `internal/hook/astgrep_corpus_pin_test.go` — the one file
   this SPEC adds under the PRESERVE tree (§A.4 carve-out). It asserts three things mechanically:
   - the covered-language validity gate calls `t.Skip` rather than failing, so `spec.md` §A.7's
     enforcement table cannot go stale silently (AC-A16-007 (a));
   - the differential run did **not** skip — output carries `--- PASS` and neither `--- SKIP` nor
     `corpus rejected:` (AC-A16-007 (b), applying REQ-A16-017 inside this SPEC rather than leaving
     it as a rule written only for the successor);
   - the twelve fixtures under `internal/hook/security/testdata/scan-corpus/` and every recorded
     `wantDeny` value are byte-unchanged from `294b4b6ab` (AC-A16-008).
   - It is a **test, not a reviewer's `git diff`**, because with the validity gate silent a
     `wantDeny` edit is a live way to turn red green and is invisible at close.

Stop condition: `sg test` green over 26 rules from a repo-side fixture root; both mutants observed
failing; the id-keying answer recorded; the pinning test present and green. A mutant that does
**not** fail means the harness is mis-wired — a milestone blocker, not a rule defect.

Verified by: AC-A16-001 … AC-A16-008.

### M2 — Coverage matrix, evidence rule, and the four-class checker

**Priority: High. Defines what the successor SPEC is measured against.**

Work:

1. Re-run the parser probe against the installed ast-grep version — do not copy the 0.40.5 result
   forward (R6, which this SPEC now owns since `SPEC-ASTG-UPGRADE-001` is archived). Record the
   version alongside the result.
2. Author `coverage-matrix.md`: 112 cells (8 families × 14 parseable languages), plus a separate
   excluded-languages record for `r` and `flutter` carrying verbatim probe output and the version.
3. Seed the 14 already-implemented cells from the measured inventory; leave the other 98 for the
   successor, marked pending.
4. Write the checker with four distinguishable failure classes:
   - **key-set mismatch** — a *set* comparison against the Cartesian product, not a count
   - **unresolved cell** — neither rule id nor rationale
   - **unevidenced exemption** — a rationale with neither citation nor probe
   - **dangling rule id** — a cell naming a rule absent from the ruleset
5. Wire the checker into the Go test suite so it runs in CI rather than by hand.
6. Prove all four classes fire, using four synthetic defects — including the substitution case
   (one cell duplicated, another deleted, count still 112), which a count-based checker passes
   silently.

Two of these classes exist because of specific holes audit found. Class 1 is a set comparison
because a substituted cell has the right cardinality and the wrong contents, and presents as
complete. Class 4 exists because the matrix is an authored document that can drift from the
ruleset, and nothing else reads both — `sg test` never reads the matrix.

Stop condition: checker fails on all four synthetic defects with four distinct class names, and
passes on the seeded 14.

Verified by: AC-A16-009 … AC-A16-014.

### M3 — Severity reclassification and the CWE anchor

**Priority: High. Changes which writes are refused — user-visible in every project that installs
the template.**

Work:

1. For each of the 26 existing rules, apply REQ-A16-012's two-clause predicate: security family
   **and** a benign same-shape `valid` case producing zero findings.
2. Apply the outcome. Two known cases:
   - `sec-csrf-no-token-check` matches every Go HTTP handler — audit confirmed the match is
     AST-based, so gofmt indentation does not narrow it. No plausible benign same-shape
     counterpart exists, so it **stays `warning`**, with the limitation recorded as its matrix
     annotation (REQ-A16-015).
   - `sec-log-injection-unsanitized` matches every `log.Printf` with a format string. Evaluate it
     against the predicate rather than assuming the answer.
3. Add `metadata.cwe` to every security rule lacking it, and verify each `invalid` case
   instantiates that weakness class in idiomatic code (REQ-A16-011). This is the anchor replacing
   R2's falsified mitigation, and it must be in place **before** the successor writes 80 rules.
4. **Record a citation or a probe for every security rule's matched head symbol** (REQ-A16-011,
   third clause) — measured today at 0 of 14. For each rule, either name the language / stdlib /
   framework reference that documents the symbol its pattern matches, or record a probe invocation
   showing the pattern matching real code from that ecosystem, verbatim with its output. Store it
   alongside the rule's `metadata` so a reviewer reads it with the pattern rather than hunting for
   it.
   - This is the **mechanical** half of the anchor; `metadata.cwe` alone is a reviewer-checkable
     label that `acmeInternalVaultFetch($X)` survives. It applies to a presence claim exactly the
     standard REQ-A16-005 imposes on an absence claim.
5. Confirm `sec-template-injection-html` measures `error`, correcting the handed measurement
   record which listed it as `warning`.

The direction is not preset: a rule moves to `error` when it acquires the evidence and to
`warning` when it lacks it. What is fixed is the predicate, not the count.

Stop condition: every one of the 26 carries a severity justified by its own test cases; every
security rule carries a CWE anchor **and** a citation or probe for its matched head symbol; each
`warning`-classified security rule carries its recorded reason.

Verified by: AC-A16-015 … AC-A16-018, plus AC-A16-019 at close.

---

## §C. Dependency graph

```
pre-flight rmdir
 └─> M1 (harness + id keying + asset placement)
      ├─> M2 (matrix + four-class checker)
      └─> M3 (severity + CWE anchor)
           └─> [close: neutrality, catalog.yaml, config mode]
                └─> SPEC-ASTGREP-BREADTH-001 (M4-M10)
```

M2 and M3 are independent of each other once M1 lands: M2 writes a document and a checker, M3
edits rule severities. They may be worked in either order or in parallel. The close steps
(neutrality scan, `make build` + `catalog.yaml`, `sg scan --config`) ride M3's commit rather than
forming a fourth milestone — there are only three substantive milestones here, and a
close-only milestone would be ceremony.

---

## §D. Constraints

- **No direct push to `main`.** Route B for every milestone.
- **`make build` rides the same commit as any template edit — and produces no artifact to commit.**
  The template tree is compiled into the binary by `//go:embed all:templates`, which emits no file,
  and `internal/template/catalog.yaml` tracks skills and agents only: measured,
  `grep -c astgrep internal/template/catalog.yaml` → **0**. So the build is required and the
  catalog diff must be **empty**. A non-empty `catalog.yaml` diff on a ruleset-only commit is the
  defect, not the deliverable — hand-editing the catalog to manufacture one is the failure mode
  this constraint is worded to prevent (AC-A16-019's last clause asserts the file is unchanged).
- **Never `git add -A` in the primary checkout.** Stage by explicit pathspec; this worktree is one
  of several live lanes.
- **Lane-local verification only.** Run affected package tests plus `sg test`; do not run the full
  Go suite locally — push and read CI for the full verdict.
- **No existing file under `internal/hook/**` is modified.** Exactly one new file is added —
  `internal/hook/astgrep_corpus_pin_test.go` (§A.4 carve-out). The §E standing diff command,
  which excludes that one path, is the proof that every pre-existing file is byte-unchanged;
  AC-A16-008 additionally pins the twelve corpus fixtures and their `wantDeny` values.
- **No new security families.** The family axis is fixed at the measured eight.

---

## §E. Self-verification

Each milestone reports in the 5-section form (Claim / Evidence / Baseline-attribution / Gaps /
Residual-risk), with the command and its verbatim output, attributed to the tree and HEAD measured.

Standing commands:

```bash
sg test --config internal/template/templates/.moai/config/astgrep-rules/sgconfig.yml
go test ./internal/template/...
go test ./internal/hook/...
git diff --stat 294b4b6ab -- internal/hook/ ':(exclude)internal/hook/astgrep_corpus_pin_test.go'
```

> **The baseline is the pinned SHA, never the branch name.** `origin/main` moves: measured on this
> tree, `git rev-list --count --left-right origin/main...HEAD` → `10 0`, and the same command
> written against `origin/main` reports `18 files changed, 23 insertions(+), 2825 deletions(-)`
> before any work exists. A PRESERVE proof stated against a moving ref is false on arrival and
> gets worse with every upstream commit. If this branch is later rebased, re-measure §H and
> re-pin this SHA — the baseline follows the tree the evidence was taken on, not the branch tip.

The last is not incidental: it is the standing proof that this SPEC's PRESERVE list held. The
exclusion pathspec is what keeps it **true** after M1 — the carve-out adds one file, and without
the exclusion this command would report that addition and the proof would read as a violation of
the very list it is proving. Excluding exactly one named path is narrower than dropping the check;
every other file under the tree is still asserted byte-unchanged.

---

## §F. Risks carried into run-phase

`spec.md` §F holds the register. The three that bind milestone ordering:

- **R1 (id collision)** is settled at M1. Discovering it later means renaming rules that already
  have committed snapshots and matrix references.
- **R2 (contrived rules)** — its original mitigation was falsified. The replacement anchor
  (`metadata.cwe`) lands at M3, before the successor writes any new rule against it.
- **R6 (version drift)** is owned here, because the SPEC that owned ast-grep version work is
  archived. M2 re-probes rather than copying §A.3 forward.

---

## §G. Anti-patterns for this SPEC

- Writing a matrix rationale that restates the cell is empty without the citation or probe that
  establishes it. That is the exemption-table failure REQ-A16-005 exists to close.
- Building the checker's class 1 as a count. A substituted cell passes a count and is worse than a
  missing one, because it presents as complete.
- Promoting every security rule to `error` because "security is error". The predicate has two
  clauses and the second is what keeps the gate usable.
- Placing rule-test fixtures under the template tree because that is where the ruleset lives. They
  would ship to every user, carrying deliberately dangerous constructs into the tree the deployed
  ruleset scans.
- Describing the corpus validity gate as failing. It skips (`spec.md` §A.7). Three prior revisions
  of this SPEC got this wrong.
- Copying `.moai/astgrep-rules/` into the template because it already has per-language files. It
  carries mixed-locale messages and SPEC IDs and would fail neutrality on contact.
- Weakening AC-A16-019's ranking grep until it matches nothing. The `sgconfig.yml`
  equal-opportunity paragraph is a **named sanctioned exemption** (REQ-A16-001), not a reason to
  blunt the check.
- Measuring an absence claim from a drifted working directory. `git ls-tree -r <ref>` scopes to
  the cwd prefix and returns empty with no diagnostic — the trap that produced two wrong scopings
  of this SPEC (`research.md` §8).

---

## §H. Cross-references

- `spec.md` — requirements, exclusions, risk register
- `acceptance.md` — the 19 criteria each milestone is verified against
- `design.md` — matrix schema, checker contract, id-keying decision record
- `SPEC-ASTGREP-BREADTH-001` — the successor consuming this contract
- `.moai/reports/t228/plan-audit-iter1.md` / `-iter2.md` / `-iter3.md` — the three audits; iter-3
  drove this propagation pass
- `.moai/reports/t228/plan-measurements.md` — M1-M4 measured basis (M5 superseded)
- `internal/hook/pre_tool_scan_differential_test.go:242` — the `t.Skip` corrected in `spec.md` §A.7
