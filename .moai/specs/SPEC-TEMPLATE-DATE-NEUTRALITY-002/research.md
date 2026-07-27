# SPEC-TEMPLATE-DATE-NEUTRALITY-002 — Research

Measurement record for the plan phase. Every figure in `spec.md`, `plan.md`, and `acceptance.md` traces to a command in this file. All commands were executed from the worktree root `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/date2025` at HEAD `760f09f73` (== `origin/main`).

---

## §A Measurement instrument

### §A.1 Why a replica rather than the guard itself

The authoritative finding set belongs to the Go guard, but reading it requires either editing the guard (forbidden in plan phase) or running it with a widened pattern (same problem). The predecessor solved the equivalent problem by committing a shell classifier whose output reproduces the guard's finding set exactly, and asserted that equivalence as a requirement.

This SPEC reuses that instrument: the predecessor's `classify.sh` was copied to a scratch path and its single `DATE_RE` line rewritten from `202[6-9]-[0-1][0-9]-[0-3][0-9]` to `202[5-9]-[0-1][0-9]-[0-3][0-9]`. No other line was changed. The copy lives outside the repository; nothing in the tree was modified.

```bash
sed 's/^DATE_RE=.*/DATE_RE='"'"'202[5-9]-[0-1][0-9]-[0-3][0-9]'"'"'/' \
  .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/classify.sh > <scratch>/classify-2025.sh
grep -n '^DATE_RE=' <scratch>/classify-2025.sh
# → 27:DATE_RE='202[5-9]-[0-1][0-9]-[0-3][0-9]'
```

### §A.2 File-scope equivalence check

The replica's `find` filter and the guard's scanned-extension set must agree, or the replica would over- or under-count.

- Classifier filter: `*.md`, `*.tmpl`, `*.yaml`, `*.yml`, `*.sh`, `*.json`, `*.js`, `.gitignore`, `.gitattributes`.
- Guard `leakTextExtensions`: `.md`, `.tmpl`, `.yaml`, `.yml`, `.sh`, `.json`, `.js`; guard `leakScannedDotfiles`: `.gitignore`, `.gitattributes`.

These agree exactly. `.gitkeep` is deliberately excluded on both sides.

### §A.3 Instrument validation against the predecessor's arithmetic

Running the replica and filtering to the `202[6-9]` range reproduces the *post-remediation* residual set:

```bash
awk -F'\t' '$2 ~ /^202[6-9]-/' <scratch>/rows-all.tsv | wc -l
# → 88
```

The predecessor's disposition arithmetic states `carved out = 100 − k`, where `k` is the number of `DC-5` rows adjudicated REMOVE, and its close record fixes `k = 12`. That yields 88. The replica reproduces this independently, which validates it as a measurement instrument for this SPEC.

Category breakdown of those 88 residual rows:

| Category | Shape | Rows |
|---|---|---:|
| DC-1 | LS-FM | 48 |
| DC-3 | LS-OTHER | 13 |
| DC-2b | LS-PROSE-STAMP | 11 |
| DC-5 | LS-OTHER | 9 |
| DC-4 | LS-OTHER | 6 |
| DC-5 | LS-FM-FENCED | 1 |

---

## §B Headline measurements

```bash
grep -rEn '\b2025-[0-9]{2}-[0-9]{2}\b' internal/template/templates | wc -l
# → 74
grep -rlE '\b2025-[0-9]{2}-[0-9]{2}\b' internal/template/templates | wc -l
# → 34
grep -rnE '^[[:space:]]*updated:[[:space:]]*"?2025-' internal/template/templates | wc -l
# → 10
```

Replica-derived:

```bash
awk -F'\t' '$2 ~ /^2025-/' <scratch>/rows-all.tsv | wc -l          # rows      → 74
cut -f1,2 <scratch>/rows-2025.tsv | sort -u | wc -l                # findings  → 48
cut -f1   <scratch>/rows-2025.tsv | sort -u | wc -l                # files     → 34
```

| Measurement | Observed | Agrees with the task brief? |
|---|---:|---|
| Occurrences | 74 | Yes |
| Findings | 48 | Yes |
| Files | 34 | Yes |
| Frontmatter-shaped `updated:` | 10 | Yes (count), **No** (classification — see §D) |

---

## §C Category and shape distribution

```bash
awk -F'\t' '{print $5"\t"$3}' <scratch>/rows-2025.tsv | sort | uniq -c | sort -rn
#   33 DC-5    LS-OTHER
#   28 DC-2a   LS-PROSE-STAMP
#   13 DC-2b   LS-PROSE-STAMP
```

Per-category row and finding counts:

```bash
# DC-2a: rows=28 findings=22
# DC-2b: rows=13 findings=7
# DC-5:  rows=33 findings=23
# total distinct findings = 48
```

`22 + 7 + 23 = 52 ≠ 48`, the difference being the 4 dual-category findings enumerated in §E.

**`DC-1`, `DC-3`, and `DC-4` are each empty for this set.** §D and §F explain why, and why each emptiness is a substantive finding rather than an absence of interest.

---

## §D Refuted hypothesis — the frontmatter carve-out does not apply

The task brief hypothesised that the 10 frontmatter-shaped `updated:` occurrences correspond to the predecessor's `DC-1` class, whose structural gate would already carve them, so that widening the year class would probably not flag them — and asked for this to be verified rather than assumed.

**The hypothesis is refuted.** The classifier assigns zero `2025` rows to `DC-1`.

### Root cause

Both the classifier's `LS-FM` rule and the guard's `DC-1` structural gate require **at least one leading whitespace character**:

- classifier: `^[[:space:]]+updated:[[:space:]]*"?20`
- guard `dc1FrontmatterUpdatedRe`: `^[ \t]+updated:[ \t]*"?20`

Every `2025` `updated:` line sits at column 0:

```bash
grep -rhE '^updated:[[:space:]]*"?2025-'      internal/template/templates | wc -l   # → 10
grep -rhE '^[[:space:]]+updated:[[:space:]]*"?2025-' internal/template/templates | wc -l   # → 0
```

The verification grep in the brief used `^[[:space:]]*` (zero-or-more), which matches both indented and column-0 lines and therefore cannot distinguish the two.

### Why the indentation differs

The 48 `DC-1` rows in the `202[6-9]` set are **nested** keys — the `updated:` field appears indented under a parent mapping:

```
.claude/rules/moai/development/skill-authoring.md:89:  updated: "2026-01-28"
.claude/skills/moai-domain-database/SKILL.md:22:      updated: "2026-04-25"
```

All 10 `2025` occurrences are top-level keys inside frontmatter blocks that are themselves **documentation examples**, concentrated in three files:

```
.claude/skills/moai-foundation-cc/reference/examples.md:27
.claude/skills/moai-foundation-cc/reference/skill-examples.md:31,162,304,517,853,1161,1597
.claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md:27,173
```

The clearest instance carries its own teaching annotation:

```
skill-formatting-guide.md:27:updated: 2025-11-25 # Optional: last update date
```

### Consequences

1. A widened guard **will** flag all 10 — they are not carved by anything today.
2. They are not real frontmatter; they are frontmatter *shown as an example*. Deleting them damages the example rather than neutralizing a leak.
3. Each therefore requires an explicit allowlist entry, and the SPEC needs a stated disposition for the documentation-example construct as a class (`spec.md` REQ-TDN2-011).

This is the single largest correction the plan phase produced.

---

## §E Dual-category findings

```bash
# same (file,date) appearing under two categories
.claude/skills/moai-foundation-cc/reference/skill-examples.md         2025-11-25  => DC-2b DC-5
.claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md 2025-11-25  => DC-2b DC-5
.claude/skills/moai-workflow-spec/references/examples.md              2025-12-07  => DC-2a DC-5
.claude/skills/moai-workflow-project/references/examples.md           2025-12-06  => DC-2a DC-5
```

Two of the four pair a REMOVE category (`DC-2a`) with an adjudicated category (`DC-5`) in the same file under the same date literal. These are the rows that make the allowlist-masking hazard concrete (`design.md` §C).

---

## §F Sibling year-bearing patterns

The guard file contains **four** year-bearing sites — three regex constructs plus one prose doc comment. An earlier revision of this section counted only three and omitted the comment; the omission mattered, because AC-018's `202[6-9]` target of `3 → 1` cannot be met without editing it.

| # | Site | Current text | Tier | Measured relevance to 2025 | Disposition |
|---|---|---|---|---|---|
| 1 | `S1-internal-date` class pattern | `\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b` | strict | the entire 48-finding set | **widen** (REQ-TDN2-016) |
| 2 | `TestLeakClassNoDateShaInDefaultTier` doc comment | prose: `…generic-date regex (202[6-9]-MM-DD)…` | n/a (comment) | describes site 1's shape; goes stale when site 1 widens | **update** (REQ-TDN2-024) |
| 3 | Attribution line matcher | `\(imported 20[0-9]{2}-\|^\*\*Import Date\b` | strict (structural gate) | already spans all `20XX` | leave unchanged |
| 4 | Archive-path class | `Finding A[1-6]\|archive-202[6-9]-…` | narrow | zero occurrences at either year range | leave unchanged |

### Site 2 — cross-SPEC ownership

Site 2 is not merely a comment in a shared file: it is the doc comment of `TestLeakClassNoDateShaInDefaultTier`, a **different test function enforcing a different SPEC's acceptance criterion**. Its own text names the owner:

```go
// TestLeakClassNoDateShaInDefaultTier enforces AC-SBN-018(a): the SKILL-BODY
// additions to the DEFAULT-tier leakClasses MUST NOT include a generic-date
// regex (202[6-9]-MM-DD) or a short-sha regex ([0-9a-f]{7,8}). Those classes
// are owned exclusively by SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001's strict
// tier (S1/S2); duplicating them here would create dual-allow-list drift
```

Two facts make the edit safe, and both were measured rather than assumed:

- **No behavioural coupling.** The test scans `leakClasses` (the *default* tier) with `dateProbe := "2026-06-04"`; this SPEC widens a *strict*-tier class. Widening `S1` cannot change what the default tier matches, and `2026-06-04` matches `202[5-9]` exactly as it matched `202[6-9]`.
- **Green at baseline**, so a post-change re-run is a meaningful comparison:

```bash
go test ./internal/template/... -run 'TestLeakClassNoDateShaInDefaultTier' -v
# --- PASS: TestLeakClassNoDateShaInDefaultTier (0.00s)
# ok  	github.com/modu-ai/moai-adk/internal/template	0.558s
```

The edit is therefore confined to the prose parenthetical naming the year range. AC-032 re-runs this test so the cross-SPEC risk is discharged by evidence rather than by argument.

```bash
grep -rE 'archive-2025-[0-1][0-9]-[0-3][0-9]'   internal/template/templates | wc -l   # → 0
grep -rE 'archive-202[6-9]-[0-1][0-9]-[0-3][0-9]' internal/template/templates | wc -l # → 0
grep -rnE '\(imported 2025-|^\*\*Import Date'   internal/template/templates
#   .claude/rules/moai/NOTICE.md:91:**Import Date (harness)**: 2026-04-26
#   .claude/rules/moai/NOTICE.md:92:**Import Date (Karpathy)**: 2026-04-28
#   .claude/rules/moai/NOTICE.md:93:**Import Date (im-not-ai)**: 2026-06-15
grep -cE '2025-[0-9]{2}-[0-9]{2}' internal/template/templates/.claude/rules/moai/NOTICE.md   # → 0
```

The attribution matcher is already year-agnostic, so `DC-4` would carve a `2025` attribution record if one existed — but none does. The archive-path class matches a distinct `archive-`-prefixed construct that is absent from the tree at both year ranges, so widening it would be a no-op with no observable effect. Both are recorded as out of scope with the measurement attached, rather than silently skipped.

---

## §G Reconciling the "41 stamps / 29 files" figure

A prior memory note and the predecessor's own §5 recorded "41 `Last Updated: 2025-*` prose authoring stamps across 29 files". The task brief flagged this figure as wrong.

**Both the figure and the brief's caution are partially correct, and the distinction matters.** The figure is arithmetically exact but scope-incomplete: it counts the `LS-PROSE-STAMP` shape only.

```bash
grep -rhE '^[[:space:]]*(#[[:space:]]*)?(\*\*)?(Last )?Updated(\*\*)?:[[:space:]]*"?2025-' \
  internal/template/templates | wc -l                                   # → 41
grep -rlE '^[[:space:]]*(#[[:space:]]*)?(\*\*)?(Last )?Updated(\*\*)?:[[:space:]]*"?2025-' \
  internal/template/templates | wc -l                                   # → 29
awk -F'\t' '$3=="LS-PROSE-STAMP"' <scratch>/rows-2025.tsv | wc -l        # → 41
awk -F'\t' '$3=="LS-PROSE-STAMP"' <scratch>/rows-2025.tsv | cut -f1 | sort -u | wc -l  # → 29
```

The replica and the predecessor's grep agree exactly at 41 / 29. What the figure omits is the 33 `LS-OTHER` rows — version-history records, documentation-example values, `Created:` stamps, and the composite footers.

Two further corrections follow:

- **41 is not the REMOVE count.** Of those 41 prose stamps, 13 fall under the mirror-capture directory and are PRESERVE. The REMOVE-class prose stamps number **28**.
- **The full triage surface is 74 rows / 48 findings / 34 files**, not 41 / 29.

The figure should not be carried forward in any form; the SPEC uses 74 / 48 / 34 with 28 REMOVE.

---

## §H Fence state is not a usable discriminator

A tempting mechanical rule for "documentation-example value" is "the line is inside a fenced code block". The measurement rejects it.

```bash
awk -F'\t' '{print $5"\tfenced="$6}' <scratch>/rows-fence.tsv | sort | uniq -c
#   27 DC-2a   fenced=0
#    1 DC-2a   fenced=1
#   11 DC-2b   fenced=0
#    2 DC-2b   fenced=1
#   18 DC-5    fenced=0
#   15 DC-5    fenced=1
```

`DC-5` splits 18 / 15, so fence state neither selects the documentation-example rows nor excludes the others. Two independent failure modes:

- The JSON schema example lives in a `.json` file, where no fence markers exist at all, so a fence-based rule cannot reach it.
- `DC-2a` and `DC-2b` each contain fenced rows, so a fence-based rule would also reclassify prose stamps.

This is why `spec.md` introduces sub-shape *rationale codes* within `DC-5` rather than a seventh classifier category (`design.md` §B).

### The fenced REMOVE-class row

Exactly one `DC-2a` row is fenced:

```
.claude/skills/moai-workflow-project/references/examples.md:192  Last Updated: 2025-12-06
```

Its context is a fenced block demonstrating generated documentation output:

```
### POST /api/auth/refresh
Token refresh endpoint

---
Generated from: SPEC-001
Last Updated: 2025-12-06
```

The `DC-2a` decision rule does not inspect fence state, so a mechanical sweep of the 28 `DC-2a` rows would edit this sample without review. `spec.md` REQ-TDN2-012 requires it to be adjudicated explicitly.

---

## §I Baselines captured for acceptance criteria

All commands run from the worktree root.

| # | Command | Observed |
|---|---|---|
| 1 | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -run TestTemplateNoInternalContentLeak -v` | `--- PASS: TestTemplateNoInternalContentLeak (0.52s)`, `ok`, exit 0 |
| 2 | REMOVE-scoped prose stamps (prose-stamp grep, excluding the mirror directory) | 28 |
| 3 | PRESERVE-scoped prose stamps (prose-stamp grep, mirror directory only) | 13 |
| 4 | `grep -c '202\[5-9\]' internal/template/internal_content_leak_test.go` | 0 |
| 5 | `grep -c 'MOAI_TEMPLATE_LEAK_STRICT' .github/workflows/template-neutrality-check.yaml` | 1 |
| 6 | unquoted `-run <Name>` invocations in the workflow | 1 |
| 7 | quoted `-run '<Name>'` invocations in the workflow | 2 |
| 8 | `grep -cE '2025-[0-9]{2}-[0-9]{2}' internal/template/catalog.yaml` | 0 |
| 9 | placeholder-token scan over stamp values (AC-011, corrected form) | **9** |
| 10 | edit-scope pathspec diff against the merge-base (AC-029) | 0 |
| 11 | line-number `awk` predicates in `acceptance.md` (AC-031) | 1 |
| 12 | `go test … -run 'TestLeakClassNoDateShaInDefaultTier' -v` (AC-032) | `--- PASS`, exit 0 |
| 13 | `grep -c '202[6-9]' internal/template/internal_content_leak_test.go` | 3 |

Baseline 8 confirms the generated catalog artifact carries no `2025` date, so it is neither a remediation target nor a source of false positives — recorded because the predecessor's acceptance record names generated-path exclusion as a recurring criterion defect.

### Correction to baseline 9 (AC-011)

An earlier revision published AC-011 with a baseline of `0`. That figure was measured with a two-stage pipe whose second stage anchored `^[[:space:]]*updated:` against the `path:lineno:` prefix that `grep -rn` emits, so that alternative could never fire. Falsified in a scratch fixture carrying three injected placeholder shapes:

```
Last Updated: YYYY-MM-DD
updated: YYYY-MM-DD
Version: 5.0.0 | Last Updated: YYYY-MM-DD | Enterprise Ready:
```

| Command form | Detected of 3 |
|---|---:|
| original two-stage pipe | 2 (the `updated:` line never fires) |
| single-stage, `^`-anchored | 2 (misses the mid-line composite) |
| **single-stage, unanchored (adopted)** | **3** |

The adopted form is unanchored precisely because the `COMPOSITE` sub-shape carries its stamp mid-line — the shape most at risk of a placeholder substitution under REQ-TDN2-009, and the one an anchored pattern cannot see.

Re-measured over the real tree, the corrected command returns **9**, not `0`. All nine are pre-existing schema documentation teaching the ISO date format, and all must be preserved:

```
.claude/rules/moai/development/spec-frontmatter-schema.md:23,24,178,179   created:/updated: YYYY-MM-DD
.claude/skills/moai/workflows/plan/spec-assembly.md:87,88                  created:/updated: YYYY-MM-DD
.claude/skills/moai-workflow-spec/references/reference.md:40,91,170        Created: YYYY-MM-DD
```

AC-011's target is therefore `9` (invariant), not `0`. This is strictly stronger than the original: a target equal to a non-zero baseline cannot be satisfied by a broken path — an unresolvable `$TPL` returns `0 ≠ 9` and fails — whereas the original target of `0` was its own false-pass value.

---

## §J Open questions carried into plan phase

1. **`HIST` disposition** (14 rows). A version-history record pairs a released version with its release date. It is arguably a factual record rather than an authoring stamp, but it is also internal project history shipping to users. Recorded as `PER-ROW` with no default; `plan.md` M2 resolves it.

   **The predecessor's own precedent is the most constraining input, and it must be on the table before M2 decides.** SPEC-TEMPLATE-DATE-NEUTRALITY-001 `spec.md` §5 records, under "Known cosmetic residue":

   > In `internal/template/templates/.claude/skills/moai-foundation-cc/SKILL.md` the version-history list now reads unevenly — the in-scope `v5.0.0` / `v4.0.0` entries lost their dates while `v3.0.0 (2025-12-06)` / `v2.0.0 (2025-11-26)` retain theirs. This is a consequence of the class boundary, not a defect in the remediation.

   Confirmed live in the tree — these are `HIST` rows 1-2 of the 14:

   ```bash
   grep -n '202[5-9]-' internal/template/templates/.claude/skills/moai-foundation-cc/SKILL.md
   # 244:- v3.0.0 (2025-12-06): Added progressive disclosure, sub-agent details, integration patterns
   # 245:- v2.0.0 (2025-11-26): Initial comprehensive release
   ```

   So the predecessor **already removed the 2026 halves of this same list**. The two surviving entries are not a neutral starting position — they are the residue the predecessor flagged and deferred. Preserving them perpetuates a known cosmetic defect in a file the predecessor already edited; removing them resolves it. This does not decide the other 12 `HIST` rows (which sit in version-history *tables* in three `references/reference.md` files and carry no such precedent), and M2 may legitimately split the disposition between the two shapes.

2. **`CREATED` disposition** (3 rows). `Created:` is shaped like an authoring stamp but names a different event than `Last Updated:`. Recorded as `PER-ROW`. All three sit in one file under one date (`moai-workflow-spec/references/examples.md`, `2025-12-07`), alongside the `DC-2a` prose stamp at line 955 under that same literal — so this is one of the two dual-category findings and either disposition is covered: PRESERVE routes through AC-016's file-scoped check, REMOVE leaves no allowlist entry and the post-widening guard catches any survivor.

3. **`COMPOSITE`** (2 rows) — **re-posed as an editing instruction, not a decision.** The earlier phrasing ("does removing a mid-line stamp constitute a placeholder substitution?") was mis-posed: REQ-TDN2-009 already answers it — removal is never itself a substitution. The real gap was that the two rows were never located, so the *shape of the edit* could not be reviewed. Both are now named, and both carry byte-identical text:

   ```
   .claude/skills/moai-workflow-testing/modules/advanced-patterns.md:576
   .claude/skills/moai-workflow-testing/modules/optimization.md:505
   → Version: 5.0.0 | Last Updated: 2025-11-22 | Enterprise Ready:
   ```

   Verified by `od -c`, the line ends `Enterprise Ready:` + newline — a trailing label with no value, and it is the final line of each file. The date-bearing *construct* under REQ-TDN2-009 is the ` Last Updated: 2025-11-22 |` segment, not the whole line: the line also carries `Version: 5.0.0` and `Enterprise Ready:`, neither of which is date-bearing, so deleting the line would remove content outside the SPEC's scope.

   **What M2 owes is therefore the residual-line text**, and the candidate is:

   ```
   Version: 5.0.0 | Enterprise Ready:
   ```

   i.e. excise the stamp plus exactly one `|` delimiter, leaving no doubled separator and no dangling pipe. M2 confirms or replaces this residual; whichever it chooses, AC-011 verifies no placeholder token was introduced in its place.

None of the three blocks plan-phase completion; each is an M2 adjudication item with its measurement attached.

---

## §K Sub-shape partition derivation (all 33 `DC-5` rows)

The `EX-FM 10 / EX-DATA 3 / HIST 14 / CREATED 3 / DEADLINE 1 / COMPOSITE 2` partition is enumerated here per row, so `spec.md` REQ-TDN2-003's table and AC-004 have a stated basis rather than an asserted count. Generated by joining each classifier row's `(file, line_no)` back to its line text:

```bash
awk -F'\t' '$5=="DC-5"{print $1"\t"$4}' <rows-2025> | sort |
  while IFS=$'\t' read -r f l; do printf '%s:%s => %s\n' "$f" "$l" "$(sed -n "${l}p" "internal/template/templates/$f")"; done
```

| Code | Rows | File : line numbers |
|---|---:|---|
| `EX-FM` | 10 | `moai-foundation-cc/reference/examples.md:27`; `…/skill-examples.md:31,162,304,517,853,1161,1597`; `…/skill-formatting-guide.md:27,173` |
| `EX-DATA` | 3 | `moai-workflow-project/references/examples.md:250,500`; `moai-workflow-project/schemas/tab_schema.json:3` |
| `HIST` | 14 | `moai-foundation-cc/SKILL.md:244,245` (bullet form); `moai-foundation-core/references/reference.md:469,470,471,472,473`; `moai-workflow-project/references/reference.md:267,268,269`; `moai-workflow-testing/references/reference.md:431,432,433,434` (table form) |
| `CREATED` | 3 | `moai-workflow-spec/references/examples.md:69,386,673` |
| `DEADLINE` | 1 | `moai-foundation-cc/reference/skill-formatting-guide.md:716` |
| `COMPOSITE` | 2 | `moai-workflow-testing/modules/advanced-patterns.md:576`; `moai-workflow-testing/modules/optimization.md:505` |

`10 + 3 + 14 + 3 + 1 + 2 = 33`, matching the `DC-5` row total in §C.

**`HIST` splits into two shapes** that M2 may dispose of differently (see §J item 1): a **bullet** form (2 rows, `- v3.0.0 (2025-12-06): …`) carrying the predecessor's flagged residue, and a **table** form (12 rows, `| 2.3.0 | 2025-12-03 | … |`) spread across three `references/reference.md` files with no such precedent.

The line numbers in this table are a plan-phase locator for review, not a criterion anchor — no acceptance criterion depends on them (AC-031 enforces that mechanically).
