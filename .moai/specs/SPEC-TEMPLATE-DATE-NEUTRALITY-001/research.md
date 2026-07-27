# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — Research

Every figure below was produced by a command actually run against this worktree at `c7309aeb6`. Figures that were not measured are stated as gaps in §H.

---

## §A Guard behaviour (measured)

### A.1 Strict tier fails; narrow tier passes; package is green

```
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1
exit=1
    internal_content_leak_test.go:725: template internal-content leak detected (135 occurrences, mode=strict):
    ...
    internal_content_leak_test.go:738:   ... 85 more (capped)
FAIL	github.com/modu-ai/moai-adk/internal/template	0.936s
```

```
$ go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1
narrow exit=0
ok  	github.com/modu-ai/moai-adk/internal/template	0.677s
```

```
$ go test ./internal/template/...
package-wide exit=0
ok  	github.com/modu-ai/moai-adk/internal/template	1.292s
```

The third result **withdraws** an iteration-1 claim. `design.md` §C previously justified the isolated-target CI convention by citing "pre-existing unrelated failures" in `internal/template`. The package is green; that rationale is false. It originated in the workflow file's own in-file comment and was repeated without verification.

### A.2 Strict mode is wired nowhere

```
$ grep -rn "LEAK_STRICT" --include="*.yaml" --include="*.yml" --include="Makefile" . | grep -v "internal_content_leak_test.go"
exit=1 (1 = no matches)
```

### A.3 Class attribution of the reported rows

```
$ grep -o "class=[A-Za-z0-9-]*" <strict.log> | sort | uniq -c
  50 class=S1-internal-date
```

Only 50 rows are visible (the cap). See §H gap G1 — the guard's own classification of rows 51-135 was not observed, though §B's S1-only enumeration reproducing exactly 135 is strong circumstantial evidence.

### A.4 Finding identity is `(file, distinct match literal)`

From `collectLeakViolations`:

```go
matches := class.pattern.FindAllString(text, -1)
...
seen := map[string]struct{}{}
for _, m := range matches {
    trimmed := strings.TrimSpace(m)
    ...
    if _, ok := seen[trimmed]; ok { continue }
    seen[trimmed] = struct{}{}
```

Consequence: repeated occurrences of the same date in one file collapse to one finding. This is why line counts exceed finding counts, and why this SPEC introduces the finer **occurrence-class row** unit (§C).

### A.5 Scan scope

`shouldScanForLeak` admits extensions `.md .tmpl .yaml .yml .sh .json .js`, plus the extensionless dotfiles `.gitignore` and `.gitattributes` by basename. `.gitkeep` is deliberately excluded.

### A.6 Report cap

`limit := 50` at `internal_content_leak_test.go:730`, with `... %d more (capped)` at :738. Observed truncation line: `... 85 more (capped)`.

---

## §B The committed classifier

`classify.sh` (this directory) is the committed classification recipe required by REQ-TDN-006, and its finding-set equivalence is REQ-TDN-020.

```
$ bash classify.sh                                  # from internal/template/
rows: 180
distinct (file,date) findings: 135
$ diff <classifier-findings> <independent-grep-findings>
(no output — IDENTICAL to the 135-finding guard-equivalent set)
```

The independent enumeration it is diffed against reimplements the guard's scan scope and dedup semantics in shell:

```bash
find templates -type f \( -name '*.md' -o -name '*.tmpl' -o -name '*.yaml' -o -name '*.yml' \
  -o -name '*.sh' -o -name '*.json' -o -name '*.js' -o -name '.gitignore' -o -name '.gitattributes' \) -print0 \
| while IFS= read -r -d '' f; do
    grep -oE '\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b' "$f" | sort -u | while read -r d; do echo "$f|$d"; done
  done
```

Result: **135** rows — matching the guard's reported count exactly.

### B.1 A real defect the equivalence check caught

The classifier's first revision emitted **137**, not 135. Cause: `awk` ERE has no `\b`, so the date pattern matched inside longer tokens. The two extras were `"expiresAt": "2026-01-06T10:00:00Z"` in `moai-workflow-spec/references/examples.md` and one analogous embedded literal in `moai-harness-learner/SKILL.md`. Explicit non-word-character boundary checks on the characters flanking each match brought the count to exactly 135.

This is the reason REQ-TDN-020 asserts *set equality*, not "the classifier runs": the broken revision ran fine and produced a plausible-looking number.

### B.2 Distinct files

116.

---

## §C Finding and row distribution (measured)

### C.1 By date literal (top values, finding-level)

| Date | Findings |
|---|---:|
| 2026-01-06 | 43 |
| 2026-11-22 | 9 |
| 2026-03-30 | 6 |
| 2026-02-21 | 6 |
| 2026-07-10 | 4 |
| 2026-01-11 | 4 |

### C.2 By category — occurrence-class rows (the normative partition)

```
$ bash classify.sh | cut -f5 | sort | uniq -c | sort -rn
  80 DC-2a
  48 DC-1
  22 DC-5
  13 DC-3
  11 DC-2b
   6 DC-4
```

Sum: **180** = the total row count. This is the partition REQ-TDN-001 states.

### C.3 By category — collapsed to findings (spans allowed)

```
$ bash classify.sh | awk -F'\t' '{print $1"|"$2"\t"$5}' | sort -u | cut -f2 | sort | uniq -c | sort -rn
  60 DC-2a
  48 DC-1
  17 DC-5
  11 DC-2b
   9 DC-3
   3 DC-4
```

Sum: 148, not 135, because 13 findings span two categories (§C.5).

### C.4 By line shape

```
$ bash classify.sh | cut -f3 | sort | uniq -c | sort -rn
  91 LS-PROSE-STAMP
  48 LS-FM
  44 LS-OTHER
   1 LS-FM-FENCED
```

`LS-PROSE-STAMP` 91 = `DC-2a` 80 + `DC-2b` 11 (the arithmetic is exact only when read against the ordered rule, since DC-3 and DC-4 outrank shape — which is why the category table, not this one, is normative).

**Line → row delta (91 rows from 89 lines).** §C.6 measures the DC-2 line total as **89**; this table reports **91** `LS-PROSE-STAMP` rows. The two-row surplus is not a unit slip — it is the row unit doing its job:

```
$ classify.sh … | awk -F'\t' '$3=="LS-PROSE-STAMP"' | wc -l                     → 91
$ classify.sh … | awk -F'\t' '$3=="LS-PROSE-STAMP"{print $1"\t"$4}' | sort -u | wc -l → 89
$ … | sort | uniq -c | awk '$1>1'
   3 .claude/skills/moai/workflows/loop.md	380
```

A single line — `moai/workflows/loop.md:380` — carries **three** distinct date literals (`2026-07-12` in the leading `Updated:` token, plus `2026-07-09` and `2026-03-02` in the changelog prose that follows). Per REQ-TDN-003 each literal binds to its own row, so that one line yields 3 rows: `89 − 1 + 3 = 91`. This is the same line that motivated REQ-TDN-003 in the first place, and it is the only line in the tree carrying more than one literal.

### C.5 Dual-shape findings

```
$ bash classify.sh | awk -F'\t' '{print $1"|"$2"\t"$5}' | sort -u | cut -f1 | uniq -d | wc -l
13
```

Twelve are `DC-1`/`DC-2a` conflicts — a file whose frontmatter `updated:` and whose prose `Updated:` footer carry the *same* date literal:

```
.claude/skills/moai-domain-backend/SKILL.md|2026-01-11
.claude/skills/moai-domain-frontend/SKILL.md|2026-03-28
.claude/skills/moai-workflow-ddd/SKILL.md|2026-01-16
.claude/skills/moai-workflow-loop/SKILL.md|2026-01-11
.claude/skills/moai-workflow-tdd/SKILL.md|2026-02-03
.claude/skills/moai-workflow-testing/SKILL.md|2026-07-10
.claude/skills/moai/references/reference.md|2026-02-22
.claude/skills/moai/workflows/gate.md|2026-03-29
.claude/skills/moai/workflows/loop.md|2026-07-12
.claude/skills/moai/workflows/mx.md|2026-02-22
.claude/skills/moai/workflows/plan.md|2026-05-25
.claude/skills/moai/workflows/project.md|2026-02-21
```

The thirteenth is a `DC-1`/`DC-5` span: `moai-foundation-cc/SKILL.md|2026-01-11` (frontmatter line 21 + a changelog line 242 `- v5.0.0 (2026-01-11): …`).

These 13 are why the classification unit is the occurrence-class row. A finding-level disposition cannot express "delete one of my two lines, keep the other" — it must pick one, and either choice is wrong for half the file.

### C.6 The 18-line DC-2 blind spot (iteration-1 defect, now closed)

Iteration 1's `AC-TDN-007` regex matched only the `Last Updated:` shape (70 lines), while `REQ-TDN-001`'s DC-2 rule named three shapes. Measured:

| Shape | Lines |
|---|---:|
| `Last Updated:` | 70 |
| bare `Updated:` | 18 |
| `# Updated:` (in `.gitignore`) | 1 |
| **DC-2 total (all three shapes)** | **89** |

The 18 bare-`Updated:` lines were the blind spot. Their locations:

- 11 in `.claude/skills/moai-foundation-cc/reference/` — the mirror-capture stamps, now `DC-2b` PRESERVE
- 7 in `.claude/skills/moai/workflows/` — `gate.md:195`, `loop.md:380`, `clean.md:268`, `moai.md:276`, `plan.md:167`, `run/mode-orchestration.md:112`, `sync/delivery.md:465` — all `DC-2a` REMOVE

Iteration 2 closes this by expressing AC-TDN-007 against the committed classifier rather than a hand-written grep, so the rule and the criterion cannot diverge again.

### C.7 Where the 43 `2026-01-06` findings live

| Directory | Findings |
|---|---:|
| `.claude/skills/moai-foundation-cc/reference` | 11 |
| `.claude/skills/moai-workflow-worktree/modules` | 10 |
| `.claude/skills/moai-foundation-core/modules` | 7 |
| `.claude/skills/moai-workflow-testing/modules/automated-code-review/trust5-framework` | 6 |
| `.claude/skills/moai-workflow-testing/modules` | 6 |
| `.claude/skills/moai-workflow-testing/modules/automated-code-review` | 2 |
| `.claude/skills/moai-foundation-cc` | 1 |

The largest cluster is `moai-foundation-cc/reference` (11), whose files mirror third-party documentation and whose `Updated: 2026-01-06` line is a mirror-capture stamp — not an authoring stamp. This is the origin of `DC-2b` and REQ-TDN-011.

Naming precision: 10 of those 11 files carry the `-official.md` suffix; the eleventh is `advanced-agent-patterns.md`. The directory holds 15 `-official.md` files, only 10 of which bear a stamp. An `*-official.md` glob would therefore both over- and under-select; the category rule is scoped by directory instead.

### C.8 Corrected misclassifications (iteration 1 → iteration 2)

Three rows were misclassified in iteration 1 and are corrected by the committed classifier:

| Row | Iter-1 | Iter-2 | Evidence |
|---|---|---|---|
| `.gitignore \| 2026-01-10` | DC-5 | **DC-2a** | line 5 is `# Updated: 2026-01-10` — a shape `spec.md` explicitly names as DC-2 |
| `skill-authoring.md \| 2026-01-28` | DC-5 | **DC-5** (confirmed, for a different reason) | line 89 is `  updated: "2026-01-28"` *inside a fenced block* → `LS-FM-FENCED`, and line 45 is `- updated: ISO date as string (e.g., "2026-01-28")` → `LS-OTHER`. Both pedagogical. Not real frontmatter, so not DC-1 |
| `anti-patterns.md \| 2026-04-28` | DC-5 (gap G3, "context not re-read") | **DC-2a** | line 425 is `**Last Updated**: 2026-04-28` — an unambiguous prose stamp. Gap G3 is now CLOSED |

The `DC-1` count is 48 rather than 49 precisely because of the fenced-example exclusion above.

---

## §D Precedent mechanisms in the guard

| Mechanism | Shape | Anchoring |
|---|---|---|
| `skillBodyScoped` | class applies only under `.claude/skills/` | path prefix |
| `skillMoaiScoped` | class applies only under `.claude/skills/moai/` | path prefix |
| `requireHexLetter` | match must contain `[a-f]` | match content |
| `pedagogicalAllowlist` | `(File, SpecID)` pair skip, 5 entries | **file + match content** |

The frequently-repeated claim that `pedagogicalAllowlist` is line-number-anchored and therefore drift-prone is **not supported by the code**. `isPedagogicallyAllowed` compares `entry.File == relPath && entry.SpecID == matched` only; `LineStart`/`LineEnd` are documented diagnostic-only fields, never read in enforcement.

```
$ awk '/^func isPedagogicallyAllowed/,/^}/' internal_content_leak_test.go | grep -cE 'LineStart|LineEnd'
0
```

---

## §E Mirror and CI surface

### E.1 Byte-parity allowlist does not intersect this SPEC's scope

`rule_template_mirror_test.go` enforces byte-identity for an explicit non-glob list:

```
.claude/rules/moai/core/hooks-system.md
.claude/rules/moai/workflow/spec-workflow.md
.claude/rules/moai/workflow/session-handoff.md
.claude/rules/moai/development/model-policy.md
.moai/config/evaluator-profiles/default.md
.moai/config/evaluator-profiles/frontend.md
```

The four date-bearing files under `templates/.claude/rules/moai/` — `NOTICE.md`, `development/spec-frontmatter-schema.md`, `development/skill-authoring.md`, `workflow/archived-agent-rejection.md` — do not appear. A positive control confirms the membership check is not vacuous: `session-handoff.md` is correctly detected as `IN-ALLOWLIST`.

### E.2 Local-counterpart overlap

Of the 116 affected template files, **115** have a counterpart at the corresponding local working-tree path; 1 is template-only.

### E.3 Existing CI workflow convention

`.github/workflows/template-neutrality-check.yaml` runs `TestTemplateNeutralityAudit` in isolation by test name:

```
$ grep -c -- '-run TestTemplateNeutralityAudit' .github/workflows/template-neutrality-check.yaml
1
$ grep -c -- '-run TestTemplateNoInternalContentLeak' .github/workflows/template-neutrality-check.yaml
0
```

Its trigger paths already include `internal/template/internal_content_leak_test.go`. Its in-file comment states a package-wide green "is NOT required" because of pre-existing failures. §A.1 measures the package green **from the repo root in a clean checkout** — which withdraws the claim as *this SPEC's* rationale but does not establish it false in general: §F shows the same package failing two tests when a stray `.moai` marker sits inside `internal/template/`. The comment is left in place; deleting an environment-dependent claim on the strength of one green run would be the larger error.

---

## §F The `internal/template/.moai` cwd trap (measured)

`classify.sh` originally documented "run from `internal/template/`". Any command run from that cwd causes the moai statusline/memory subsystem to create a `.moai/` marker directory there. `output_styles_audit_test.go:137-149` `findProjectRoot()` ascends looking for a `.moai` marker; with one present at `internal/template/`, the ascent halts there instead of at the repo root, `liveDir` resolves to a nonexistent path, and two tests fail:

```
--- FAIL: TestOutputStylesTemplateLiveParity
    output_styles_audit_test.go:415: ReadDir(".../debt-clear/internal/template/.claude/output-styles/moai") error: ... no such file or directory
--- FAIL: TestOutputStylesFallbackDocsContract
    output_styles_audit_test.go:513: ReadFile(".../debt-clear/internal/template/.claude/rules/moai/core/settings-management.md") error: ... no such file or directory
```

The marker is recreated automatically on any later command run from that cwd, so deleting it once does not fix the trap — the documented invocation has to change. `classify.sh` already accepted the template root as `$1` (`ROOT="${1:-templates}"`), so the fix is free. Measured from the repo root:

```
$ bash .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/classify.sh internal/template/templates | wc -l
180
$ … | awk -F'\t' 'NF>=5{print $5}' | sort | uniq -c
  80 DC-2a   48 DC-1   22 DC-5   13 DC-3   11 DC-2b   6 DC-4
$ ls -d internal/template/.moai
ls: internal/template/.moai: No such file or directory
```

Identical output, no marker created. The repo-root form is now canonical in `classify.sh`'s header, in every `acceptance.md` command block, and as a `[WATCH]` note in `plan.md` M2/M3.

This is also why §E.3 does not assert the workflow's "pre-existing failures" comment is false in general: the same package is green from the repo root and red from a polluted cwd.

---

## §G Frontmatter key reach (DC-1 scope correction)

```
$ grep -rhE '^[[:space:]]+created:[[:space:]]*"?202[6-9]-…' templates … | wc -l
0
$ grep -rhE '^[[:space:]]+version:[[:space:]]*"?202[6-9]-…' templates … | wc -l
0
```

The tree contains **zero** dated `created:` and **zero** dated `version:` frontmatter lines. `DC-1` reaches the `updated:` key only. Iteration-1 prose describing DC-1 as `updated:` / `created:` / `version:` overstated the rule; corrected in `spec.md` REQ-TDN-001 and `plan.md`.

---

## §H Gaps (not observed)

- **G1** — That *all 135* findings carry `class=S1-internal-date` is **not** verified. The guard caps at 50 rows; all 50 visible rows are S1, and the independent S1-only enumeration in §B reproduces exactly 135 — strong circumstantial evidence — but the guard's own classification of rows 51-135 was not observed. The cap was never raised.
- **G2** — `S2-short-sha-sentence-final` contributes 0 findings *among the visible 50*. Its contribution to rows 51-135 is unobserved.
- **G3** — **CLOSED.** `anti-patterns.md | 2026-04-28` was re-read: the file carries exactly one line with that literal, line 425 `**Last Updated**: 2026-04-28`. Reclassified `DC-2a`.
- **G4** — **CLOSED by design change.** Iteration 1's classifier bound each finding to the *first* line in the file containing that literal, which mis-binds a finding on a line carrying two literals. The committed classifier operates per-line-per-literal with explicit boundary checks, so this class of error is structurally excluded. REQ-TDN-003 states the rule; §B.1 documents the boundary-check defect that the equivalence assertion caught.
- **G5** — No measurement of how many `DC-2a` removals would leave an orphaned or empty header/footer block. No AC mechanizes this (`acceptance.md` §D).
- **G6** — Whether the carve-out actually satisfies REQ-TDN-012 is unproven at plan phase; it is precondition P2, probed at M6 (AC-TDN-015). What *is* measured today is the hazard's reality: the current S1 pattern matches both a synthetic future attribution date and a synthetic future frontmatter bump.
- **G7** — The 7 `DC-2a` bare-`Updated:` rows in `moai/workflows/` (§C.6) were classified but not individually adjudicated for surrounding-block damage. They are ordinary REMOVE rows; the G5 caveat applies to them as it does to the other 73.
