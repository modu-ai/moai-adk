# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — Acceptance Criteria

Every judgment command below was **executed against this tree at `c7309aeb6`** during plan phase, and its observed baseline is recorded verbatim. Commands are given in fenced blocks, never inside table cells, so no regex metacharacter is reinterpreted by the markdown table parser.

**Scope convention.** Unless a criterion states otherwise, commands run from `internal/template/` and paths are relative to it, so `templates/...` denotes the distributed template tree. This is the same root the guard uses (`const templatesRoot = "templates"`). Where a criterion asserts something about the *guard's own* finding set, its scope is `(file, distinct date literal)` pairs per `research.md` §A.4 — not raw line occurrences. The two scopes differ numerically (see AC-TDN-007 / AC-TDN-008 notes) and must not be conflated.

---

## §A Summary matrix

| AC | Subject | Milestone | Baseline run? |
|---|---|---|---|
| AC-TDN-001 | Strict tier baseline captured | M1 | yes |
| AC-TDN-002 | Narrow tier stays green | M7 | yes |
| AC-TDN-003 | Inventory row count reconciles with the guard | M2 | yes (absence) |
| AC-TDN-004 | No unadjudicated inventory row | M2 | yes (absence) |
| AC-TDN-005 | DC-3 deadline dates preserved | M3 | yes |
| AC-TDN-006 | DC-4 attribution dates preserved | M3 | yes |
| AC-TDN-007 | DC-2 prose stamps removed | M3 | yes |
| AC-TDN-008 | DC-1 frontmatter fields preserved | M3 | yes |
| AC-TDN-009 | Carve-out is content-anchored | M4 | yes |
| AC-TDN-010 | Report cap no longer hides findings | M5 | yes |
| AC-TDN-011 | CI enforcement wired (conditional) | M6 | yes |
| AC-TDN-012 | Strict tier reaches zero findings | M4 | yes |
| AC-TDN-013 | Neutrality workflow test stays green | M7 | yes |
| AC-TDN-014 | Build stays green | M7 | yes |
| AC-TDN-015 | Future legitimate date is not blocked | M6 | yes |
| AC-TDN-016 | Mirror-parity allowlist not intersected | M3 | yes |

---

## §B Criteria

### AC-TDN-001 — Strict-tier baseline captured

**Given** the strict guard tier and **when** it is run against the unremediated tree, **then** it shall report a finding count that the triage inventory reconciles against.

```bash
MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ \
  -run TestTemplateNoInternalContentLeak -count=1
```

Observed baseline (from repo root):

```
exit=1
--- FAIL: TestTemplateNoInternalContentLeak (0.54s)
    internal_content_leak_test.go:725: template internal-content leak detected (135 occurrences, mode=strict):
    ...
    internal_content_leak_test.go:738:   ... 85 more (capped)
FAIL	github.com/modu-ai/moai-adk/internal/template	0.936s
```

Falsifiable: the assertion is on a specific non-zero count; a tree with a different finding set produces a different number.

---

### AC-TDN-002 — Narrow tier stays green (non-regression)

**While** the strict tier is being remediated, **the narrow tier shall** remain green.

```bash
go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1
```

Observed baseline:

```
narrow exit=0
ok  	github.com/modu-ai/moai-adk/internal/template	0.677s
```

---

### AC-TDN-003 — Inventory row count reconciles with the guard

**When** the triage inventory exists, **the inventory shall** carry exactly one row per guard finding.

```bash
# from internal/template/
find templates -type f \( -name '*.md' -o -name '*.tmpl' -o -name '*.yaml' -o -name '*.yml' \
  -o -name '*.sh' -o -name '*.json' -o -name '*.js' -o -name '.gitignore' -o -name '.gitattributes' \) -print0 \
| while IFS= read -r -d '' f; do
    grep -oE '\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b' "$f" 2>/dev/null | sort -u \
      | while read -r d; do echo "$f|$d"; done
  done | wc -l
```

Observed baseline: `135` — matching the guard's own reported count exactly, which is what makes this recipe an adequate regeneration mechanism (REQ-TDN-006).

PASS condition at M2: the row count of `.moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/triage.md` data rows equals the count this recipe reports at that time.

```bash
test -f .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/triage.md; echo "exit=$?"
```

Observed baseline (from repo root): `exit=1` — the inventory does not yet exist, so the criterion is currently FAIL and is not vacuously satisfiable.

---

### AC-TDN-004 — No unadjudicated inventory row

**The inventory shall not** contain a row whose disposition is unset, `TODO`, `UNTRIAGED`, or `TBD`.

```bash
grep -cEi 'UNTRIAGED|[[:space:]]TBD[[:space:]]|\bTODO\b' \
  .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/triage.md
```

Observed baseline (from repo root):

```
grep: .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/triage.md: No such file or directory
exit=2
```

PASS condition: file exists and the count is `0`.

Falsifiable: a placeholder disposition left in any row raises the count above zero.

---

### AC-TDN-005 — DC-3 deadline dates preserved

**The remediation shall not** remove the `2026-11-22` backward-compatibility expiry date from any file that currently carries it.

```bash
# from internal/template/
grep -rlE '\b2026-11-22\b' templates --include='*.md' --include='*.tmpl' --include='*.yaml' \
  --include='*.yml' --include='*.sh' --include='*.json' --include='*.js' | wc -l
```

Observed baseline: `9`

PASS condition: still `9` after remediation.

**Falsifiability probe (executed).** On a disposable copy of the template tree, stripping the date from a single file:

```
before: 9
after removing the date from 1 file: 8
```

The assertion therefore detects loss.

---

### AC-TDN-006 — DC-4 attribution dates preserved

**The remediation shall not** alter the import-date records in the attribution notice.

```bash
# from internal/template/
grep -cE 'imported 202[6-9]-[0-1][0-9]-[0-3][0-9]' templates/.claude/rules/moai/NOTICE.md
```

Observed baseline: `3`

PASS condition: still `3` after remediation.

**Falsifiability probe (executed).** On a disposable copy, replacing one `imported 2026-04-26` with `imported`:

```
before: 3
after: 2
```

---

### AC-TDN-007 — DC-2 prose authoring stamps removed

**When** remediation completes, **standalone document header/footer authoring stamps shall not** remain in the template tree.

```bash
# from internal/template/
grep -rhE '^[[:space:]]*(\*\*)?Last Updated(\*\*)?:[[:space:]]*202[6-9]-' templates --include='*.md' | wc -l
```

Observed baseline: `70`

PASS condition: `0`.

Note on scope: `70` counts **lines**, whereas DC-2 accounts for `58` **findings**. The difference is the guard's per-file date dedup (`research.md` §A.4 / §C.3) — a file whose prose stamp and frontmatter field carry the same date yields one finding but two lines. This criterion deliberately asserts on the line scope, because the line is the edit unit; AC-TDN-012 asserts on the finding scope.

Falsifiable: the baseline is non-zero, so a `-eq 0` assertion is a real state change rather than a vacuous pass.

---

### AC-TDN-008 — DC-1 frontmatter fields preserved

**The remediation shall not** delete a frontmatter `updated:` date field (deleting the key is a schema break, REQ-TDN-009).

```bash
# from internal/template/
grep -rhE '^[[:space:]]+updated:[[:space:]]*"?202[6-9]-[0-1][0-9]-[0-3][0-9]"?[[:space:]]*$' \
  templates --include='*.md' | wc -l
```

Observed baseline: `49`

PASS condition: `49` if the DC-1 disposition resolves to preserve-as-is. If the open question resolves to value-neutralization instead, this criterion is superseded by a replacement criterion asserting the key's continued presence with a non-date value — that substitution must be recorded, not silently dropped.

Note on scope: `49` lines vs `48` DC-1 findings, same dedup reason as AC-TDN-007.

---

### AC-TDN-009 — Carve-out is content-anchored

**The carve-out mechanism shall not** identify an allowed finding by line number.

```bash
# from internal/template/ — inspects the allowlist enforcement function body only
awk '/^func isPedagogicallyAllowed/,/^}/' internal_content_leak_test.go | grep -cE 'LineStart|LineEnd'
```

Observed baseline: `0` — the existing precedent is already content-anchored; `LineStart`/`LineEnd` are documented diagnostic-only fields and are not consulted in enforcement.

PASS condition: `0` for the new date carve-out's enforcement function as well (the command is repeated against that function's name at M4).

**Falsifiability probe (executed).** Injecting `if entry.LineStart > 0 && false { return false }` into a scratch copy of the function:

```
injected-copy count: 1
real-file count: 0
```

The check detects a line-number anchor if one is introduced.

---

### AC-TDN-010 — Report cap no longer hides findings silently

**When** the finding count exceeds the console cap, **the guard shall** either report all findings or name a path holding the full listing.

```bash
# from internal/template/
grep -c 'limit := 50' internal_content_leak_test.go
```

Observed baseline: `1` — the hard-coded cap is present.

PASS condition: `0`, **and** the truncation branch either does not truncate or emits a resolvable path. The second half is verified by injecting a synthetic finding at M5 and reading the emitted message; a bare `grep -c ... -eq 0` alone would pass if the literal were merely renamed, so both halves are required.

---

### AC-TDN-011 — CI enforcement wired (conditional on M6 adoption)

**Where** preconditions P1-P3 (`design.md` §C) all hold, **the CI configuration shall** run the strict tier.

```bash
# from repo root
grep -rn "LEAK_STRICT" --include="*.yaml" --include="*.yml" --include="Makefile" . \
  | grep -v "internal_content_leak_test.go"
```

Observed baseline: `exit=1 (no matches)` — strict mode is wired nowhere.

PASS condition (if M6 is adopted): at least one match inside `.github/workflows/`.
PASS condition (if M6 is not adopted): unchanged baseline **and** the failing precondition recorded in `progress.md`. "Not adopted" is a valid closure, not a partial pass.

---

### AC-TDN-012 — Strict tier reaches zero findings

**When** remediation and carve-out are complete, **the strict tier shall** report no findings.

```bash
MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ \
  -run TestTemplateNoInternalContentLeak -count=1
```

Observed baseline: `exit=1`, `135 occurrences, mode=strict`.

PASS condition: `exit=0`.

Falsifiable: the baseline is a failing run, so a passing run is a genuine state change.

---

### AC-TDN-013 — Neutrality workflow test stays green (non-regression)

```bash
go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1
```

Observed baseline:

```
AC-013 neutrality exit=0
ok  	github.com/modu-ai/moai-adk/internal/template	0.508s
```

---

### AC-TDN-014 — Build stays green (non-regression)

```bash
go build ./...
```

Observed baseline: `AC-014 build exit=0` (no output).

---

### AC-TDN-015 — A future legitimate date is not blocked

**Where** a new legitimate attribution entry is added after remediation, **the guard shall not** report it (REQ-TDN-012 / `design.md` §C precondition P2).

Probe of the *current* regex against a synthetic future attribution line:

```bash
printf 'The following docs (imported 2027-03-04) are incorporated:\n' \
  | grep -cE '\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b'
```

Observed baseline: `1` — the current S1 pattern **does** match a future attribution date, which is exactly why a naive CI enforcement would block a legitimate future `NOTICE.md` entry. This baseline establishes the hazard as real rather than hypothetical.

PASS condition at M6: after the carve-out is in place, appending a synthetic future import line to `templates/.claude/rules/moai/NOTICE.md` and re-running the strict tier yields `exit=0`; the synthetic line is then reverted.

---

### AC-TDN-016 — Mirror-parity allowlist not intersected

**Where** a remediating edit lands on a rule file, **the edit shall** be paired with an identical local-tree edit if and only if that file is enrolled in the byte-parity allowlist.

```bash
# from internal/template/
for f in NOTICE.md development/spec-frontmatter-schema.md \
         development/skill-authoring.md workflow/archived-agent-rejection.md; do
  p=".claude/rules/moai/$f"
  if grep -qF "\"$p\"" rule_template_mirror_test.go; then echo "IN-ALLOWLIST: $p"
  else echo "not-enrolled: $p"; fi
done
```

Observed baseline:

```
not-enrolled: .claude/rules/moai/NOTICE.md
not-enrolled: .claude/rules/moai/development/spec-frontmatter-schema.md
not-enrolled: .claude/rules/moai/development/skill-authoring.md
not-enrolled: .claude/rules/moai/workflow/archived-agent-rejection.md
```

**Positive control (executed)** — the check is not vacuously "not-enrolled":

```bash
grep -qF '".claude/rules/moai/workflow/session-handoff.md"' rule_template_mirror_test.go \
  && echo "IN-ALLOWLIST: session-handoff.md (control detects correctly)"
```

Observed: `IN-ALLOWLIST: session-handoff.md (control detects correctly)`.

PASS condition: the check is re-run at M3 against the **actual** edited-file list, not this four-file sample; any `IN-ALLOWLIST` result obliges a paired local edit in the same commit.

---

## §C Definition of Done

1. All open questions in `plan.md` §B resolved and recorded.
2. `triage.md` committed, reconciling with AC-TDN-003 and clean under AC-TDN-004.
3. AC-TDN-005, 006, 007, 008, 016 pass after remediation.
4. AC-TDN-009, 010, 012 pass after the carve-out and cap change.
5. AC-TDN-011 and 015 either both pass (M6 adopted) or M6 is closed as not-adopted with the failing precondition recorded.
6. AC-TDN-002, 013, 014 pass (non-regression).
7. Every AC result recorded in `progress.md` §E.2 with the verbatim command output.

---

## §D Known AC limitations

- **AC-TDN-007 / AC-TDN-008 scope asymmetry.** These assert on line counts while the guard counts findings. The mapping is documented in each criterion; a reviewer comparing `70`/`49` against `58`/`48` without reading the note will think the numbers disagree.
- **AC-TDN-005 uses file-count, not occurrence-count.** A file that carries `2026-11-22` twice and loses one occurrence still passes. This was chosen because the guard's own identity is per-file-per-literal; an occurrence-level assertion would be stricter than the guard and would fail for reasons the guard does not care about.
- **AC-TDN-010's grep half is weak alone.** Renaming the literal would pass the grep while preserving the behaviour. The injected-finding check is the load-bearing half.
- **No AC covers `research.md` gap G5** (orphaned header blocks after DC-2 removal). That is a review-quality property this criteria set does not mechanize.
