# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — Acceptance Criteria

Every judgment command below was **executed against this tree at `c7309aeb6`** during plan phase, and its observed baseline is recorded verbatim. Commands live in fenced blocks, never inside table cells, so no regex metacharacter is reinterpreted by the markdown table parser.

**Scope convention.** Unless stated otherwise, commands run from `internal/template/` and paths are relative to it, so `templates/...` denotes the distributed template tree — the same root the guard uses (`const templatesRoot = "templates"`). Commands marked *(repo root)* run from the worktree root. Criteria asserting on the classifier operate on **occurrence-class rows** (the edit unit, 180); criteria asserting on the guard operate on **findings** (the guard's `(file, date)` unit, 135). The two units are never conflated — each criterion names which it uses.

**Classifier invocation.** `CL` denotes `../../.moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/classify.sh`, the committed classifier (REQ-TDN-006 / REQ-TDN-020).

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
| AC-TDN-009 | REQ-TDN-010b | code | M4 | yes |
| AC-TDN-010 | REQ-TDN-016 | code | M5 | yes |
| AC-TDN-011 | REQ-TDN-013 | CI | M6 | yes |
| AC-TDN-012 | REQ-TDN-010 | finding | M4 | yes |
| AC-TDN-013 | REQ-TDN-014 | test | M7 | yes |
| AC-TDN-014 | (build non-regression) | build | M7 | yes |
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
# from internal/template/
CL=../../.moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/classify.sh
TSV=../../.moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/triage.tsv
echo "classifier rows: $(bash "$CL" | wc -l)"
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
# from internal/template/
bash "$CL" | awk -F'\t' '$5=="DC-3"' | wc -l
```

Observed baseline: `13` rows (across 9 findings / 9 files).

PASS condition: still `13` after remediation.

**Falsifiability probe (executed).** On a disposable copy of the template tree, stripping the date from one file dropped the file-level count `9 → 8`:

```
before: 9
after removing the date from 1 file: 8
```

---

### AC-TDN-006 — DC-4 attribution dates preserved

The remediation **shall not** alter the import-date records in the attribution notice.

```bash
# from internal/template/
grep -cE 'imported 202[6-9]-[0-1][0-9]-[0-3][0-9]' templates/.claude/rules/moai/NOTICE.md
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
# from internal/template/
bash "$CL" | awk -F'\t' '$5=="DC-2a"' | wc -l
```

Observed baseline: `80`

PASS condition: `0`.

This criterion is expressed against the **classifier**, not a hand-written grep, precisely so the requirement (REQ-TDN-001's full three-shape DC-2 rule) and the criterion cannot diverge. A narrower hand-written regex — for example one matching only `Last Updated:` — reaches `0` while 18 `Updated:`-shaped dated lines remain in the tree; the classifier-based form cannot exhibit that blind spot because the same rule text drives both.

---

### AC-TDN-008 — DC-1 frontmatter rows preserved

The remediation **shall not** delete a frontmatter `updated:` key or alter its value.

```bash
# from internal/template/
bash "$CL" | awk -F'\t' '$5=="DC-1"' | wc -l
```

Observed baseline: `48`

PASS condition: `48`.

Scope note: this counts `LS-FM` rows only. Fenced pedagogical examples are classified `LS-FM-FENCED` → `DC-5` by construction and are therefore **excluded** — the one measured instance is `skill-authoring.md:89`, a `metadata:` block inside a fenced skill-frontmatter example, to which REQ-TDN-009's schema-break rationale does not apply. A whole-file grep for indented `updated:` lines returns `49` because it includes that fenced example; the one-row difference is the exclusion, not a discrepancy.

---

### AC-TDN-009 — Carve-out is content-anchored

The carve-out **shall not** identify a preserved row by line number.

The mechanism is now decided (hybrid, REQ-TDN-010), so the enforcement surfaces are named rather than left as an unbound parameter. Two functions are in scope:

- `collectLeakViolations` — hosts the structural gate for `DC-1` / `DC-4` (the per-class scan loop, `internal_content_leak_test.go:593-634`).
- `isDateAllowlisted` — the new allowlist gate for `DC-3` / `DC-2b` / `DC-5-PRESERVE`, to be added alongside the existing `isPedagogicallyAllowed`.

The window spans every function the carve-out touches:

```bash
# from internal/template/
awk '/^func (collectLeakViolations|isPedagogicallyAllowed|isDateAllowlisted)/,/^}/' \
  internal_content_leak_test.go | grep -cE 'LineStart|LineEnd|lineNo|LineNumber'
```

Observed baseline: `0` — measured today over the two functions that currently exist. (`isDateAllowlisted` does not exist yet; `awk` matches nothing for it, which is why the command is an alternation rather than a single name.)

PASS condition: `0` after M4, with all three functions present.

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
# from internal/template/
grep -c 'limit := 50' internal_content_leak_test.go
```

Observed baseline: `1`. PASS condition: `0`.

Half 2 — executable injection recipe. The guard reports zero findings after M4, so a synthetic finding is injected to exercise the truncation branch, then reverted:

```bash
# from internal/template/ — run AFTER M4 (strict tier at zero)
printf '\n<!-- probe 2029-12-31 -->\n' >> templates/.claude/rules/moai/NOTICE.md
MOAI_TEMPLATE_LEAK_STRICT=1 go test . -run TestTemplateNoInternalContentLeak -count=1 2>&1 \
  | grep -oE 'full listing: [^ ]+' || echo "NO-PATH-EMITTED"
git checkout -- templates/.claude/rules/moai/NOTICE.md
```

Expected output at M5: a single `full listing: <path>` line, and `test -f <path>` succeeds.

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
# from repo root
grep -rn "LEAK_STRICT" --include="*.yaml" --include="*.yml" --include="Makefile" . \
  | grep -v "internal_content_leak_test.go"
```

Observed baseline: `exit=1` (no matches) — strict mode is wired nowhere.

PASS condition (M6 adopted): at least one match inside `.github/workflows/`.
PASS condition (M6 not adopted): unchanged baseline **and** a recorded failing precondition. The not-adopted branch is closed by an executable check, not by prose:

```bash
# from repo root — the not-adopted branch requires a recorded reason
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

**Joint reachability with AC-TDN-007.** AC-TDN-007 drives `DC-2a` rows to `0` (80 deletions). The remaining 100 rows (`DC-1` 48 + `DC-5` 22 + `DC-3` 13 + `DC-2b` 11 + `DC-4` 6) are all PRESERVE-dispositioned and leave the guard's view via the M4 carve-out, not via deletion. `80 + 100 = 180` — every row is accounted for by exactly one of the two mechanisms, so both criteria are simultaneously satisfiable. Under the iteration-1 formulation they were not: AC-TDN-007's regex reached `0` while 18 DC-2-shaped dated lines survived, and those lines were neither deleted nor carved out, so AC-TDN-012 could not reach `exit=0`.

---

### AC-TDN-013 — Neutrality workflow test stays green

The existing `TestTemplateNeutralityAudit` target **shall** remain green.

```bash
go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1
```

Observed baseline: `exit=0` / `ok  github.com/modu-ai/moai-adk/internal/template  0.508s`

---

### AC-TDN-014 — Build stays green

The tree **shall** continue to build.

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
# from internal/template/ — sample form; M3 re-runs it over the actual edited-file list
for f in NOTICE.md development/spec-frontmatter-schema.md \
         development/skill-authoring.md workflow/archived-agent-rejection.md; do
  p=".claude/rules/moai/$f"
  if grep -qF "\"$p\"" rule_template_mirror_test.go; then echo "IN-ALLOWLIST: $p"
  else echo "not-enrolled: $p"; fi
done
```

Observed baseline: all four report `not-enrolled`.

**Positive control (executed)** — the check is not vacuously "not-enrolled":

```
$ grep -qF '".claude/rules/moai/workflow/session-handoff.md"' rule_template_mirror_test.go && echo IN-ALLOWLIST
IN-ALLOWLIST
```

PASS condition: re-run at M3 over the actual edited-file list; any `IN-ALLOWLIST` result obliges a paired local edit in the same commit.

---

### AC-TDN-017 — Mirror-capture stamps preserved

The remediation **shall not** delete a `DC-2b` mirror-capture stamp.

```bash
# from internal/template/
bash "$CL" | awk -F'\t' '$5=="DC-2b"' | wc -l
```

Observed baseline: `11` rows across `11` distinct files.

PASS condition: `11` after remediation.

A second, classifier-independent form asserts the lines themselves still exist, so the criterion does not become vacuous if the classifier's DC-2b rule is ever narrowed:

```bash
# from internal/template/
grep -rlE '^(\*\*)?Updated(\*\*)?:[[:space:]]*202[6-9]-' \
  templates/.claude/skills/moai-foundation-cc/reference --include='*.md' | wc -l
```

Observed baseline: `11`. PASS condition: `11`.

Naming note: 10 of the 11 files carry the `-official.md` suffix; the eleventh is `advanced-agent-patterns.md`. The directory holds 15 `-official.md` files in total, only 10 of which bear a stamp. The criterion is therefore scoped by **directory + stamp presence**, never by an `*-official.md` glob, which would both over- and under-select.

---

### AC-TDN-018 — No placeholder token introduced

**When** a row's disposition is REMOVE, the edit **shall** delete the construct rather than substitute a placeholder.

```bash
# from internal/template/
grep -rlE 'DATE-REDACTED|DATE-REMOVED|\{\{ *DATE *\}\}|YYYY-MM-DD-PLACEHOLDER|<date-elided>' \
  templates | wc -l
```

Observed baseline: `0`.

PASS condition: `0` after remediation. This is a no-regression assertion — it can only be broken by the remediation itself, which is exactly the behaviour REQ-TDN-007 forbids.

---

### AC-TDN-019 — CI step is an isolated target

The strict-tier CI step **shall** invoke the guard by test name, not the package as a whole.

```bash
# from repo root
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
# from repo root
git status --porcelain -- '.claude/' '.moai/' \
  | grep -v 'SPEC-TEMPLATE-DATE-NEUTRALITY-001' | wc -l
```

Observed baseline: `0`.

PASS condition: `0`, **unless** AC-TDN-016 reported an `IN-ALLOWLIST` file — in which case the only permitted non-zero entries are that file's local counterpart, each individually justified in the commit body.

---

### AC-TDN-021 — Classifier reproduces the guard's finding set

The committed classifier **shall** emit exactly the guard's `(file, date)` finding set.

```bash
# from internal/template/
bash "$CL" | cut -f1,2 | sort -u | wc -l
```

Observed baseline: `135` — identical to the guard's reported occurrence count, verified by set diff against an independent `grep`-based enumeration of the same regex (`diff` returned no output).

PASS condition: the count equals the guard's reported count **and** the set diff is empty.

Falsifiable, and it caught a real defect during authoring: the classifier's first revision used `awk` ERE, which has no `\b`, and emitted `137` — two extra findings from embedded literals (`"2026-01-06T10:00:00Z"` and one other). Adding explicit non-word-character boundary checks brought it to exactly `135`. A criterion asserting only "the classifier runs" would have passed that broken revision.

---

### AC-TDN-022 — Template-First edit scope

Every remediating edit **shall** land under `internal/template/templates/`, the guard file, or the CI workflow — and nowhere else.

```bash
# from repo root
git diff --name-only c7309aeb6..HEAD -- . \
  ':(exclude).moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/**' \
  ':(exclude)internal/template/templates/**' \
  ':(exclude)internal/template/internal_content_leak_test.go' \
  ':(exclude).github/workflows/template-neutrality-check.yaml' | wc -l
```

Observed baseline: `0` (plan phase has touched only the SPEC directory).

PASS condition: `0`. Paired with `make build` exiting 0 at M3, this is the mechanical half of REQ-TDN-017; the `make build` invocation itself is deferred to run phase per the plan-phase constraint and is recorded in `progress.md` §E.2 when performed.

---

### AC-TDN-023 — Dual-shape conflicts resolved per row

**Where** a finding's rows fall into two conflicting categories, each row **shall** receive its own disposition.

```bash
# from internal/template/ — findings spanning more than one category
bash "$CL" | awk -F'\t' '{print $1"|"$2"\t"$5}' | sort -u | cut -f1 | uniq -d | wc -l
```

Observed baseline: `13` dual-shape findings — 12 `DC-1`/`DC-2a` conflicts and 1 `DC-1`/`DC-5` span.

PASS condition after M3: `0` remaining `DC-1`/`DC-2a` conflicts, verified by:

```bash
# from internal/template/
bash "$CL" | awk -F'\t' '$5=="DC-1"||$5=="DC-2a"{print $1"|"$2"\t"$5}' \
  | sort -u | cut -f1 | uniq -d | wc -l
```

Observed baseline: `12`. PASS condition: `0` — reached because each conflicting finding's `DC-2a` prose row is deleted while its `DC-1` frontmatter row is preserved, so no finding remains in both categories. A finding-level disposition could not express this; that is what REQ-TDN-019 exists to prevent.

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
- **No AC covers orphaned header blocks** after a DC-2a removal (`research.md` gap G5). That remains a review-quality property this set does not mechanize.
