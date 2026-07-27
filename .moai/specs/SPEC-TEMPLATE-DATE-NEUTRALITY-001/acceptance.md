# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — Acceptance Criteria

Every judgment command below was **executed against this tree at `c7309aeb6`** during plan phase, and its observed baseline is recorded verbatim. Commands live in fenced blocks, never inside table cells, so no regex metacharacter is reinterpreted by the markdown table parser.

**Scope convention.** **Every command below runs from the worktree root.** Criteria asserting on the classifier operate on **occurrence-class rows** (the edit unit, 180); criteria asserting on the guard operate on **findings** (the guard's `(file, date)` unit, 135). The two units are never conflated — each criterion names which it uses.

**[WATCH] Never `cd` into `internal/template/` to run these commands.** The moai statusline/memory subsystem creates a `.moai/` marker directory in the cwd of any command it observes. A marker at `internal/template/` makes `findProjectRoot()` in `output_styles_audit_test.go` halt its ascent there instead of at the repo root, and `TestOutputStylesTemplateLiveParity` + `TestOutputStylesFallbackDocsContract` then both FAIL with `no such file or directory` on a path rooted at `internal/template/`. The marker is recreated on any later command run from that cwd, so it is a recurring trap that presents as a phantom test failure unrelated to this SPEC. Every command in this file is written in the repo-root form for that reason; `classify.sh` takes the template root as `$1` precisely so no `cd` is ever needed. Verified equivalent: the repo-root form emits the same 180 rows and the same six category counts, and leaves no marker.

**Anti-tamper discipline.** `classify.sh` is a committed SPEC artifact that run phase can edit, so a criterion expressed *only* through it is satisfiable by editing the measuring instrument rather than the tree. Every classifier-expressed criterion below therefore carries a **classifier-independent second form** that greps the tree directly. Both forms must hold. See §E for the tamper evidence that motivated this.

---

## §A Summary matrix

| AC | Requirement(s) covered | Unit | Milestone | Baseline run? |
|---|---|---|---|---|
| AC-TDN-001 | REQ-TDN-001 | finding | M1 | yes |
| AC-TDN-002 | REQ-TDN-014 | finding | M7 | yes |
| AC-TDN-003 | REQ-TDN-004 | row | M2 | yes |
| AC-TDN-004 | REQ-TDN-005 | row | M2 | yes |
| AC-TDN-005 | REQ-TDN-008 (DC-3) | row | M3 | yes |
| AC-TDN-006 | REQ-TDN-008 (DC-4) | line | M3 | yes |
| AC-TDN-007 | REQ-TDN-001 (DC-2a) | row | M3 | yes |
| AC-TDN-008 | REQ-TDN-009 | row | M3 | yes |
| AC-TDN-009 | REQ-TDN-021 | code | M4 | yes |
| AC-TDN-010 | REQ-TDN-016 | code | M5 | yes |
| AC-TDN-011 | REQ-TDN-013 | CI | M6 | yes |
| AC-TDN-012 | REQ-TDN-010 | finding | M4 | yes |
| AC-TDN-013 | REQ-TDN-014 | test | M7 | yes |
| AC-TDN-014 | REQ-TDN-014 (build clause) | build | M7 | yes |
| AC-TDN-015 | REQ-TDN-012 | probe | M6 | yes |
| AC-TDN-016 | REQ-TDN-018 (mirror) | code | M3 | yes |
| AC-TDN-017 | REQ-TDN-011 | row | M3 | yes |
| AC-TDN-018 | REQ-TDN-007 | line | M3 | yes |
| AC-TDN-019 | REQ-TDN-015 | CI | M6 | yes |
| AC-TDN-020 | REQ-TDN-018 (no-copy) | git | M3 | yes |
| AC-TDN-021 | REQ-TDN-002, 003, 006, 020 | finding | M2 | yes |
| AC-TDN-022 | REQ-TDN-017 | git | M3 | yes |
| AC-TDN-023 | REQ-TDN-019 | row | M3 | yes |

---

## §B Criteria

### AC-TDN-001 — Strict-tier baseline captured

**When** the strict guard tier is run against the unremediated tree, it **shall** report a finding count that the triage inventory reconciles against.

```bash
MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ \
  -run TestTemplateNoInternalContentLeak -count=1
```

Observed baseline *(repo root)*:

```
exit=1
    internal_content_leak_test.go:725: template internal-content leak detected (135 occurrences, mode=strict):
    internal_content_leak_test.go:738:   ... 85 more (capped)
FAIL	github.com/modu-ai/moai-adk/internal/template	0.936s
```

---

### AC-TDN-002 — Narrow tier stays green

**While** the strict tier is being remediated, the narrow tier **shall** remain green.

```bash
go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1
```

Observed baseline: `narrow exit=0` / `ok  github.com/modu-ai/moai-adk/internal/template  0.677s`

---

### AC-TDN-003 — Inventory row count reconciles with the classifier

**When** `triage.tsv` exists, it **shall** carry exactly one data row per occurrence-class row emitted by the committed classifier.

A **data row** is any line after the single header line, so the count command is unambiguous:

```bash
TSV=.moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/triage.tsv
echo "classifier rows: $(bash .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/classify.sh internal/template/templates | wc -l)"
echo "tsv data rows:   $(( $(wc -l < "$TSV") - 1 ))"
```

Observed baseline: classifier rows `180`; `triage.tsv` does not yet exist —

```
$ test -f .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/triage.tsv; echo "exit=$?"
exit=1
```

PASS condition at M2: both numbers equal.

---

### AC-TDN-004 — No unadjudicated inventory row

The inventory **shall not** contain a row whose `disposition` column is unset, `TODO`, `UNTRIAGED`, or `TBD`.

```bash
awk -F'\t' 'NR>1 && ($6=="" || $6=="TODO" || $6=="UNTRIAGED" || $6=="TBD")' \
  .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/triage.tsv | wc -l
```

Observed baseline *(repo root)*: file absent (`exit=1` from the `test -f` above), so the criterion is currently FAIL and is not vacuously satisfiable. PASS condition: file exists and the count is `0`.

---

### AC-TDN-005 — DC-3 deadline rows preserved

The remediation **shall not** remove any `2026-11-22` occurrence.

```bash
bash .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/classify.sh internal/template/templates \
  | awk -F'\t' '$5=="DC-3"' | wc -l
```

Observed baseline: `13` rows (across 9 findings / 9 files). PASS condition: still `13`.

**Classifier-independent second form** (must also hold):

```bash
grep -rhE '\b2026-11-22\b' internal/template/templates --include='*.md' --include='*.tmpl' \
  --include='*.yaml' --include='*.yml' --include='*.sh' --include='*.json' --include='*.js' | wc -l
```

Observed baseline: `13` lines — equal to the row count here because no line carries `2026-11-22` twice.

**Falsifiability probes (executed).** File-level, on a disposable copy: `9 → 8` after stripping the date from one file. Line-level, on the second form: `13 → 11` after removing one occurrence.

---

### AC-TDN-006 — DC-4 attribution dates preserved

The remediation **shall not** alter the import-date records in the attribution notice.

```bash
grep -cE 'imported 202[6-9]-[0-1][0-9]-[0-3][0-9]' internal/template/templates/.claude/rules/moai/NOTICE.md
```

Observed baseline: `3`

**Falsifiability probe (executed).** On a disposable copy, replacing one `imported 2026-04-26` with `imported`:

```
before: 3
after: 2
```

---

### AC-TDN-007 — DC-2a prose authoring stamps removed

**When** remediation completes, no `DC-2a` occurrence-class row **shall** remain.

```bash
bash .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/classify.sh internal/template/templates \
  | awk -F'\t' '$5=="DC-2a"' | wc -l
```

Observed baseline: `80`. PASS condition: `0`.

**Classifier-independent second form** (must also hold) — all three DC-2 shapes, excluding the `DC-2b` mirror directory:

```bash
grep -rnE '^[[:space:]]*(#[[:space:]]*)?(\*\*)?(Last )?Updated(\*\*)?:[[:space:]]*"?202[6-9]-[0-1][0-9]-[0-3][0-9]' \
  internal/template/templates --include='*.md' --include='.gitignore' \
  | grep -v 'moai-foundation-cc/reference/' | wc -l
```

Observed baseline: `78` lines. PASS condition: `0`.

Reconciling `78` lines against `80` rows: the two-row surplus is `moai/workflows/loop.md:380`, a single prose-stamp line carrying three distinct date literals (see `research.md` §C.4). Both forms are non-zero at baseline, so neither `-eq 0` assertion is vacuous.

The classifier form exists so the requirement (REQ-TDN-001's full three-shape DC-2 rule) and the criterion cannot diverge — a narrower hand-written regex matching only `Last Updated:` reaches `0` while 18 `Updated:`-shaped dated lines remain. The second form exists because the classifier alone is tamper-satisfiable (§E).

---

### AC-TDN-008 — DC-1 frontmatter rows preserved

The remediation **shall not** delete a frontmatter `updated:` key or alter its value.

```bash
bash .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/classify.sh internal/template/templates \
  | awk -F'\t' '$5=="DC-1"' | wc -l
```

Observed baseline: `48`. PASS condition: `48`.

**Classifier-independent second form** (must also hold) — the whole-file superset, fenced example included:

```bash
grep -rhE '^[[:space:]]+updated:[[:space:]]*"?202[6-9]-[0-1][0-9]-[0-3][0-9]"?[[:space:]]*$' \
  internal/template/templates --include='*.md' | wc -l
```

Observed baseline: `49` lines. PASS condition: `49`.

**Falsifiability probe (executed).** On a disposable copy, deleting one frontmatter `updated:` key: `49 → 48`.

Scope note: the classifier form counts `LS-FM` rows only. Fenced pedagogical examples are classified `LS-FM-FENCED` → `DC-5` by construction and are therefore **excluded** — the one measured instance is `skill-authoring.md:89`, a `metadata:` block inside a fenced skill-frontmatter example, to which REQ-TDN-009's schema-break rationale does not apply. A whole-file grep for indented `updated:` lines returns `49` because it includes that fenced example; the one-row difference is the exclusion, not a discrepancy.

---

### AC-TDN-009 — Carve-out is content-anchored

The carve-out **shall not** identify a preserved row by line number.

The mechanism is now decided (hybrid, REQ-TDN-010), so the enforcement surfaces are named rather than left as an unbound parameter. Two functions are in scope:

- `collectLeakViolations` — hosts the structural gate for `DC-1` / `DC-4` (the per-class scan loop, `internal_content_leak_test.go:593-634`).
- `isDateAllowlisted` — the new allowlist gate for `DC-3` / `DC-2b` / `DC-5-PRESERVE`, to be added alongside the existing `isPedagogicallyAllowed`.

The window spans every function the carve-out touches:

```bash
awk '/^func (collectLeakViolations|isPedagogicallyAllowed|isDateAllowlisted)/,/^}/' \
  internal/template/internal_content_leak_test.go | grep -cE 'LineStart|LineEnd|lineNo|LineNumber'
```

Observed baseline: `0` — measured today over the two functions that currently exist. (`isDateAllowlisted` does not exist yet; `awk` matches nothing for it, which is why the command is an alternation rather than a single name.)

PASS condition: `0` after M4, with all three functions present.

**Non-vacuity control (mandatory alongside the count).** This criterion's PASS value is `0`, and `0` is exactly what an unresolvable `awk` operand also produces: `awk` errors with `can't open file`, the pipeline carries an empty stream, and `grep -c` prints `0` — a silent false pass. The count alone therefore cannot distinguish a real pass from a vacuous one. The distinguishing control is the **awk window itself**: it must be non-empty and must contain all three `^func ` headers. Run both alongside the count and require all three results together:

```bash
awk '/^func (collectLeakViolations|isPedagogicallyAllowed|isDateAllowlisted)/,/^}/' \
  internal/template/internal_content_leak_test.go | wc -l          # window non-empty
awk '/^func (collectLeakViolations|isPedagogicallyAllowed|isDateAllowlisted)/,/^}/' \
  internal/template/internal_content_leak_test.go | grep -c '^func '  # expect 3 after M4
```

Measured at M7 *(repo root)*: count `0`, window `68` lines, `3` `^func ` headers — a real pass, not a vacuous one.

**Falsifiability probe (executed).** Injecting `if entry.LineStart > 0 && false { return false }` into a scratch copy of `isPedagogicallyAllowed`:

```
injected-copy count: 1
real-file count: 0
```

Note the existing precedent is already content-anchored: `isPedagogicallyAllowed` compares `entry.File == relPath && entry.SpecID == matched`; its `LineStart`/`LineEnd` struct fields are documented diagnostic-only and are never read in enforcement.

---

### AC-TDN-010 — Report cap no longer hides findings silently

**When** the finding count exceeds the console report cap, the guard **shall** name a filesystem path holding the complete listing.

Two halves, both binary-testable.

Half 1 — the hard-coded literal is gone:

```bash
grep -c "limit := 50" internal/template/internal_content_leak_test.go
```

Observed baseline: `1`. PASS condition: `0`.

Half 2 — executable injection recipe. The guard reports zero findings after M4, so a synthetic finding is injected to exercise the truncation branch, then reverted:

```bash
# run AFTER M4 (strict tier at zero)
printf '\n<!-- probe 2029-12-31 -->\n' >> internal/template/templates/.claude/rules/moai/NOTICE.md
MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1 2>&1 \
  | grep -oE 'full listing: [^ ]+' || echo "NO-PATH-EMITTED"
git checkout -- internal/template/templates/.claude/rules/moai/NOTICE.md
```

Expected output at M5: a single `full listing: <path>` line, and `test -f <path>` succeeds.

**Non-vacuity control (mandatory alongside the emitted path).** This half PASSes on an emitted `full listing: <path>` line and FAILs on `NO-PATH-EMITTED` — but `NO-PATH-EMITTED` is also what an unresolvable package operand produces: the `go test` target never resolves, no guard runs, the grep matches nothing, and the `||` branch fires for the wrong reason (a false FAIL rather than a false pass, but equally uninformative). The distinguishing control is that the **emitted path must satisfy `test -f`** — a path that resolves to a real file proves the guard actually ran and actually wrote the listing:

```bash
test -f <emitted-path> && echo "PASS (listing exists)" || echo "FAIL"
```

Measured at M7 *(repo root)*: emitted `full listing: /var/folders/…/T/moai-template-leak-*.log`, and `test -f` on it succeeded.

Observed baseline for the truncation message today — the shape being replaced:

```
$ grep -oE '\.\.\. [0-9]+ more \(capped\)' <strict-run-log>
... 85 more (capped)
```

The current message names a count and no path, which is exactly the defect. PASS requires the `NO-PATH-EMITTED` branch never to fire.

---

### AC-TDN-011 — CI enforcement wired (conditional on M6 adoption)

**Where** preconditions P1-P3 (`design.md` §C) all hold, the CI configuration **shall** run the strict tier.

```bash
grep -rn "LEAK_STRICT" --include="*.yaml" --include="*.yml" --include="Makefile" . \
  | grep -v "internal_content_leak_test.go"
```

Observed baseline: `exit=1` (no matches) — strict mode is wired nowhere.

PASS condition (M6 adopted): at least one match inside `.github/workflows/`.
PASS condition (M6 not adopted): unchanged baseline **and** a recorded failing precondition. The not-adopted branch is closed by an executable check, not by prose:

```bash
# the not-adopted branch requires a recorded reason
grep -cE '^precondition_failed: (P1|P2|P3)' \
  .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/progress.md
```

Expected in the not-adopted branch: `1` (exactly one recorded failing precondition). Observed baseline *(repo root)*: `0` — no such line yet, so neither branch is vacuously satisfied today.

---

### AC-TDN-012 — Strict tier reaches zero findings

**When** remediation and carve-out are complete, the strict tier **shall** report no findings.

```bash
MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ \
  -run TestTemplateNoInternalContentLeak -count=1
```

Observed baseline: `exit=1`, `135 occurrences, mode=strict`. PASS condition: `exit=0`.

**Joint reachability with AC-TDN-007 — robust, not conditional.** Write `k` for the number of `DC-5` rows adjudicated REMOVE at M2 (`0 ≤ k ≤ 22`; `spec.md` REQ-TDN-005 requires per-row adjudication, so `k` is not fixed at plan time). Then:

- deleted by remediation = `DC-2a` 80 + `k` = **`80 + k`**
- carved out by the guard = `DC-1` 48 + `DC-3` 13 + `DC-2b` 11 + `DC-4` 6 + `(22 − k)` = **`100 − k`**
- total = `(80 + k) + (100 − k)` = **180 for every `k`**

Every row exits by exactly one mechanism whatever M2 decides, so AC-TDN-007 (`DC-2a → 0`) and AC-TDN-012 (`exit=0`) are simultaneously satisfiable across the whole range. An earlier draft argued this by asserting all 100 non-`DC-2a` rows were PRESERVE — i.e. `k = 0` — which contradicted REQ-TDN-005 and `plan.md` M3 step 2. The `k`-parameterised form is strictly stronger: it does not depend on an M2 outcome at all.

Under the iteration-1 formulation the two were genuinely unreachable together: AC-TDN-007's regex reached `0` while 18 DC-2-shaped dated lines survived, neither deleted nor carved out, so AC-TDN-012 could not reach `exit=0`.

---

### AC-TDN-013 — Neutrality workflow test stays green

The existing `TestTemplateNeutralityAudit` target **shall** remain green.

```bash
go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1
```

Observed baseline: `exit=0` / `ok  github.com/modu-ai/moai-adk/internal/template  0.508s`

---

### AC-TDN-014 — Build stays green

The tree **shall** continue to build. This closes REQ-TDN-014's build clause; REQ-TDN-014 is the SPEC's single non-regression requirement and owns all three checks (narrow tier → AC-TDN-002, neutrality target → AC-TDN-013, build → this criterion).

```bash
go build ./...
```

Observed baseline: `exit=0` (no output).

---

### AC-TDN-015 — A future legitimate date is not blocked

**Where** a new legitimate attribution record or a frontmatter `updated:` bump is added after remediation, the guard **shall not** report it.

Two probes, because `design.md` §B identifies the DC-1 per-commit tax as the larger recurring cost while DC-4 is the more visible one.

Probe A — future attribution record (DC-4):

```bash
printf 'The following docs (imported 2027-03-04) are incorporated:\n' \
  | grep -cE '\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b'
```

Observed baseline: `1` — the current S1 pattern **does** match a future attribution date. This establishes the hazard as real: a naive CI enforcement would block a legitimate future `NOTICE.md` entry.

Probe B — frontmatter bump (DC-1):

```bash
printf '  updated: "2028-05-09"\n' \
  | grep -cE '\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b'
```

Observed baseline: `1` — the same hazard applies to every future skill frontmatter bump.

PASS condition at M6: after the carve-out is in place, appending each synthetic line to its respective real file (`NOTICE.md` for A, any skill `SKILL.md` frontmatter for B) and re-running the strict tier yields `exit=0` for both; both synthetic lines are then reverted.

---

### AC-TDN-016 — Mirror-parity allowlist not intersected

**Where** a remediating edit lands on a rule file, it **shall** be paired with an identical local-tree edit if and only if that file is enrolled in the byte-parity allowlist.

```bash
# sample form; M3 re-runs it over the actual edited-file list
for f in NOTICE.md development/spec-frontmatter-schema.md \
         development/skill-authoring.md workflow/archived-agent-rejection.md; do
  p=".claude/rules/moai/$f"
  if grep -qF "\"$p\"" internal/template/rule_template_mirror_test.go; then echo "IN-ALLOWLIST: $p"
  else echo "not-enrolled: $p"; fi
done
```

Observed baseline: all four report `not-enrolled`.

**Positive control (executed)** — the check is not vacuously "not-enrolled":

```
$ grep -qF '".claude/rules/moai/workflow/session-handoff.md"' internal/template/rule_template_mirror_test.go && echo IN-ALLOWLIST
IN-ALLOWLIST
```

PASS condition: re-run at M3 over the actual edited-file list; any `IN-ALLOWLIST` result obliges a paired local edit in the same commit.

---

### AC-TDN-017 — Mirror-capture stamps preserved

The remediation **shall not** delete a `DC-2b` mirror-capture stamp.

```bash
bash .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/classify.sh internal/template/templates | awk -F'\t' '$5=="DC-2b"' | wc -l
```

Observed baseline: `11` rows across `11` distinct files.

PASS condition: `11` after remediation.

A second, classifier-independent form asserts the lines themselves still exist, so the criterion does not become vacuous if the classifier's DC-2b rule is ever narrowed:

```bash
grep -rlE '^(\*\*)?Updated(\*\*)?:[[:space:]]*202[6-9]-' \
  internal/template/templates/.claude/skills/moai-foundation-cc/reference --include='*.md' | wc -l
```

Observed baseline: `11`. PASS condition: `11`.

Naming note: 10 of the 11 files carry the `-official.md` suffix; the eleventh is `advanced-agent-patterns.md`. The directory holds 15 `-official.md` files in total, only 10 of which bear a stamp. The criterion is therefore scoped by **directory + stamp presence**, never by an `*-official.md` glob, which would both over- and under-select.

---

### AC-TDN-018 — No placeholder token introduced

**When** a row's disposition is REMOVE, the edit **shall** delete the construct rather than substitute a placeholder.

```bash
grep -rlE 'DATE-REDACTED|DATE-REMOVED|\{\{ *DATE *\}\}|YYYY-MM-DD-PLACEHOLDER|<date-elided>' \
  internal/template/templates | wc -l
```

Observed baseline: `0`.

PASS condition: `0` after remediation. This is a no-regression assertion — it can only be broken by the remediation itself, which is exactly the behaviour REQ-TDN-007 forbids.

---

### AC-TDN-019 — CI step is an isolated target

The strict-tier CI step **shall** invoke the guard by test name, not the package as a whole.

```bash
grep -c -- '-run TestTemplateNoInternalContentLeak' \
  .github/workflows/template-neutrality-check.yaml
```

Observed baseline: `0` (`exit=1`).

**Positive control (executed)** — the grep shape does find an isolated target when one is present:

```
$ grep -c -- '-run TestTemplateNeutralityAudit' .github/workflows/template-neutrality-check.yaml
1
```

PASS condition (M6 adopted): `1`. AC-TDN-011 alone cannot distinguish an isolated target from a package-wide run — it greps only for `LEAK_STRICT` anywhere — so this criterion is the one that binds REQ-TDN-015.

---

### AC-TDN-020 — No local-tree copy performed

The remediation **shall not** copy content between the template tree and the local working trees.

```bash
git status --porcelain -- '.claude/' '.moai/' \
  | grep -v 'SPEC-TEMPLATE-DATE-NEUTRALITY-001' | wc -l
```

Observed baseline: `0`.

PASS condition: `0`, **unless** AC-TDN-016 reported an `IN-ALLOWLIST` file — in which case the only permitted non-zero entries are that file's local counterpart, each individually justified in the commit body.

---

### AC-TDN-021 — Classifier reproduces the guard's finding set

The committed classifier **shall** emit exactly the guard's `(file, date)` finding set.

```bash
bash .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/classify.sh internal/template/templates | cut -f1,2 | sort -u | wc -l
```

Observed baseline: `135` — identical to the guard's reported occurrence count, verified by set diff against an independent `grep`-based enumeration of the same regex (`diff` returned no output).

PASS condition: the count equals the guard's reported count **and** the set diff is empty.

Falsifiable, and it caught a real defect during authoring: the classifier's first revision used `awk` ERE, which has no `\b`, and emitted `137` — two extra findings from embedded literals (`"2026-01-06T10:00:00Z"` and one other). Adding explicit non-word-character boundary checks brought it to exactly `135`. A criterion asserting only "the classifier runs" would have passed that broken revision.

---

### AC-TDN-022 — Template-First edit scope

Every remediating edit **shall** land under `internal/template/templates/`, the guard file, or the CI workflow — and nowhere else.

```bash
git diff --name-only "$(git merge-base origin/main HEAD)"..HEAD -- . \
  ':(exclude).moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/**' \
  ':(exclude)internal/template/templates/**' \
  ':(exclude)internal/template/internal_content_leak_test.go' \
  ':(exclude).github/workflows/template-neutrality-check.yaml' \
  ':(exclude)internal/template/catalog.yaml' | wc -l
```

**Amended during run phase — the `catalog.yaml` exclude.** `internal/template/catalog.yaml` is a **generated artifact of the `make build` that REQ-TDN-017 mandates**, not a hand edit: `make build` rewrites its 12 skill content hashes whenever a file under `internal/template/templates/` changes. Reverting the regeneration is not available — a stale catalog fails the `internal/template` suite (measured: stashing the regenerated `catalog.yaml` and re-running `go test ./internal/template/` yields FAIL), so committing it is the only correct action. Without this exclude, a run that correctly obeys REQ-TDN-017 necessarily fails AC-TDN-022 — the two requirements were in direct conflict. Excluding a mechanically-generated path **preserves** the criterion's intent, which is to detect *hand* edits outside the declared scope; it narrows nothing else, because every remaining path in the pathspec is still observed.

Observed baseline: `0` (plan phase has touched only the SPEC directory). The merge-base resolves to `c7309aeb6` today; `origin/main` has since advanced to `f99bf4b8a`. A hardcoded `c7309aeb6..HEAD` would keep working now but silently report unrelated files the moment this branch is rebased onto the advanced `origin/main` — a false positive the orchestrator hit this iteration. `$(git merge-base origin/main HEAD)` is rebase-stable.

PASS condition: `0`. Paired with `make build` exiting 0 at M3, this is the mechanical half of REQ-TDN-017; the `make build` invocation itself is deferred to run phase per the plan-phase constraint and is recorded in `progress.md` §E.2 when performed.

---

### AC-TDN-023 — Dual-shape conflicts resolved per row

**Where** a finding's rows fall into two conflicting categories, each row **shall** receive its own disposition.

```bash
# findings spanning more than one category
bash .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/classify.sh internal/template/templates \
  | awk -F'\t' '{print $1"|"$2"\t"$5}' | sort -u | cut -f1 | uniq -d | wc -l
```

Observed baseline: `13` dual-shape findings — 12 `DC-1`/`DC-2a` conflicts and 1 `DC-1`/`DC-5` span.

PASS condition after M3: `0` remaining `DC-1`/`DC-2a` conflicts —

```bash
bash .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/classify.sh internal/template/templates \
  | awk -F'\t' '$5=="DC-1"||$5=="DC-2a"{print $1"|"$2"\t"$5}' | sort -u | cut -f1 | uniq -d | wc -l
```

Observed baseline: `12`. PASS condition: `0` — each conflicting finding's `DC-2a` prose row is deleted while its `DC-1` frontmatter row is preserved, so no finding remains in both categories.

**Classifier-independent second form** (must also hold) — files carrying both a frontmatter `updated:` line and a prose stamp on the *same* literal:

```bash
find internal/template/templates -type f -name '*.md' -print0 | while IFS= read -r -d '' f; do
  grep -oE '\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b' "$f" 2>/dev/null | sort -u | while read -r d; do
    fm=$(grep -cE "^[[:space:]]+updated:[[:space:]]*\"?$d\"?[[:space:]]*$" "$f")
    pr=$(grep -cE "^[[:space:]]*(\*\*)?(Last )?Updated(\*\*)?:[[:space:]]*$d" "$f")
    [ "$fm" -ge 1 ] && [ "$pr" -ge 1 ] && echo "$f|$d"
  done
done | wc -l
```

Observed baseline: `12`. PASS condition: `0`.

**Falsifiability probe (executed).** On a disposable copy, deleting one prose stamp from a conflict file: `12 → 11`.

**Coverage of the 13th finding (`DC-1`/`DC-5` span).** The commands above scope to `DC-1`/`DC-2a`, so they do not observe `moai-foundation-cc/SKILL.md | 2026-01-11` (frontmatter line 21 + changelog line 242). That exclusion is sound only where line 242 is adjudicated PRESERVE; adjudicated REMOVE, it becomes a `DC-1`-PRESERVE / `DC-5`-REMOVE conflict of exactly the shape REQ-TDN-019 governs. REQ-TDN-022 closes the gap by requiring M2 to adjudicate any `DC-5` row sharing a finding with a `DC-1` row explicitly against REQ-TDN-019 and to record the determination in the row's `rationale`. The conditional check:

```bash
# run at M3 — every DC-5 row that shares a finding with a DC-1 row must carry a REQ-TDN-019 note
awk -F'\t' 'NR>1 && $5=="DC-5" && $7 !~ /REQ-TDN-019/ {print $1"\t"$2}' \
  .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/triage.tsv \
  | while IFS=$'\t' read -r f d; do
      grep -qP "^\Q$f\E\t\Q$d\E\t.*\tDC-1\t" .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/triage.tsv && echo "UNANNOTATED: $f $d"
    done | wc -l
```

Expected at M3: `0`. Observed baseline: not runnable — `triage.tsv` does not exist yet (`test -f` → `exit=1`), which is the same not-yet-satisfiable state as AC-TDN-003/004 and is not a vacuous pass.

A finding-level disposition could express none of this; that is what REQ-TDN-019 exists to prevent.

---

## §C Definition of Done

1. `triage.tsv` committed; AC-TDN-003, 004, 021 pass.
2. Remediation complete; AC-TDN-005, 006, 007, 008, 016, 017, 018, 020, 022, 023 pass.
3. Carve-out + cap complete; AC-TDN-009, 010, 012 pass.
4. AC-TDN-011, 015, 019 either all pass (M6 adopted) or M6 closes as not-adopted with a recorded failing precondition (checked by AC-TDN-011's second command).
5. Non-regression: AC-TDN-002, 013, 014 pass.
6. Every AC result recorded in `progress.md` §E.2 with verbatim command output.

---

## §D Known AC limitations

- **AC-TDN-006 and AC-TDN-018 assert on lines, not classifier rows.** They are deliberately classifier-independent so a future narrowing of a classifier rule cannot silently make them vacuous. The trade-off is that they do not reconcile numerically with the row counts.
- **AC-TDN-005 counts rows, not literal occurrences within a row.** A line carrying `2026-11-22` twice contributes one row. This matches the guard's own identity and is intentional.
- **AC-TDN-010 half 1 is weak alone** — renaming the literal would pass the grep. Half 2 (the injection recipe) is the load-bearing half; both are required.
- **AC-TDN-022 does not verify `make build` itself.** It verifies edit scope. The `make build` exit code is recorded in `progress.md` §E.2 at M3; no plan-phase command can assert it without violating the plan-phase constraint.
- **AC-TDN-022's exclude list was amended during run phase** to add `internal/template/catalog.yaml` (a fifth `:(exclude)` entry). The conflict was unobservable at plan time: §5 forbids plan phase from running `make build`, so the cascade `template edit → make build → catalog.yaml regenerated` could not be measured, and the note directly above — asserting only that AC-TDN-022 "does not verify `make build` itself" — did not anticipate that `make build` *produces a tracked-file change*. As written, the criterion contradicted REQ-TDN-017: obeying the mandated `make build` guaranteed a non-zero count. The amendment narrows nothing else — the four original excludes and the merge-base anchor are unchanged, and every non-generated path remains observed.
- **AC-TDN-009's `awk` operand and AC-TDN-010 half 2's `go test` target were corrected during run phase.** Both were published as bare, cwd-relative operands (`internal_content_leak_test.go` and `go test .`) that only resolve when run from `internal/template/` — the exact invocation this file's §Scope forbids and the `[WATCH]` note warns against. From the repo root, AC-TDN-009's `awk` errored with `can't open file`, yielding an empty stream and a `grep -c` of `0` — **which is that criterion's own PASS condition**, so the published form was a silent false pass that would have reported PASS regardless of the source's contents. AC-TDN-010 half 2 failed in the opposite direction (`no Go files in …` → a false `NO-PATH-EMITTED`), non-executable rather than falsely green. Both operands are now repo-root-relative (`internal/template/internal_content_leak_test.go`; `go test ./internal/template/`), and each criterion gained a non-vacuity control that distinguishes a real result from an unresolvable-path artifact. Same defect class as the AC-TDN-022 conflict recorded above: a plan-phase-authored command whose scope assumption diverged from the file's own scope convention. No PASS condition, baseline figure, or assertion changed — this was a command-correctness fix.
- **No AC covers orphaned header blocks** after a DC-2a removal (`research.md` gap G5). That remains a review-quality property this set does not mechanize.
- **The second forms are not exact duplicates of the classifier rules.** Each is a deliberately simpler tree grep, so their baselines differ from the row counts by explained deltas (78 vs 80, 49 vs 48). They are a cross-check, not a reimplementation; a divergence beyond the documented delta is itself a signal worth investigating.

---

## §E Tamper evidence (why every classifier-expressed criterion carries a second form)

`classify.sh` is a committed SPEC artifact, so run phase can edit it. Measured, on a copy with one category branch disabled (`else if (shape == "LS-PROSE-STAMP")` → `else if (0)`) and the template tree **completely untouched**:

```
real classifier      DC-2a: 80
tampered classifier  DC-2a: 0
tampered classifier  DC-2b: 0
tampered total rows:       180
tampered (file,date) findings: 135
```

AC-TDN-007's PASS condition of `DC-2a == 0` is reached with all 80 lines still present. **AC-TDN-021 does not backstop this**: the `(file, date)` set and the row total are both unchanged under the tamper, so category-rule tampering is invisible to it.

Scope of the problem, precisely: with respect to **tree** changes these criteria are not circular — the classifier greps the tree, so deleting a `DC-1` line does drop the count. With respect to **classifier** changes they are.

### Approach chosen: replicate the dual-form pattern (not a classifier pin)

AC-TDN-017 already carried a classifier-independent second form and was the only criterion immune — it held at `11` under the tampered classifier. That pattern is now applied to AC-TDN-005, 007, 008, and 023 as well.

The alternative — pinning `classify.sh`'s category-rule block by content hash or by grepping for the rule literals — was rejected on two measured grounds:

1. **A rule-literal grep is tamper-shape-dependent.** Against this very tamper, `grep -c 'DC-2a'` returns `1` on both the real and the tampered file — the tamper is invisible to it. (`grep -c 'LS-PROSE-STAMP'` does drop `2 → 1` here, but only because this particular edit happened to remove that literal; an edit renaming the emitted category, or negating a different condition, would leave both literal counts intact.) A pin that catches some tampers and not others provides false assurance.
2. **A content hash produces false failures on legitimate edits.** `classify.sh`'s header changed in this very iteration to document the repo-root invocation; a hash pin would have failed on that no-op-to-behaviour change, training a maintainer to re-baseline the hash reflexively — which defeats the pin.

The dual-form approach has neither failure mode: each second form independently reads the tree, so it is unaffected by any classifier edit, and it does not break when the classifier is legitimately maintained.
