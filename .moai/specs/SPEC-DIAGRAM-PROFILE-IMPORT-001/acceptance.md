# acceptance.md — SPEC-DIAGRAM-PROFILE-IMPORT-001

All criteria test DOCUMENTED behavior contracts in template content: greps,
diffs, a build run, and file reads. There is no Go code in this SPEC — a PASS
is always an observed command output, never an impression.

Template root shorthand: `TPL` = `internal/template/templates/.claude/skills`.

## §D AC Matrix

| AC | Group | Binary statement |
|----|-------|------------------|
| AC-DPI-001 | (a) | Profile reference exists with mechanism + marker-first + slug + schema-backfill contracts |
| AC-DPI-002 | (a) | design-dna SKILL.md points at it; existing phases intact |
| AC-DPI-003 | (a) | Write-safety contracts present (confirm + re-read) |
| AC-IMP-001 | (b) | import-mermaid.md exists; untrusted-source + no-carry-over stated |
| AC-IMP-002 | (b) | import-drawio.md exists; four containers + decode-first stated |
| AC-IMP-003 | (b) | Fidelity ledger (merged/collapsed/dropped + check-svg pass) in both |
| AC-IMP-004 | (b) | Bulk-replace / one-home in both, phrased as same-change |
| AC-IMP-005 | (b) | §2.5 numeric obligations (r=8, 6–10, ≥12, paint order) in both |
| AC-IMP-006 | (b) | SKILL.md table rows opt-in; default flow unchanged in substance |
| AC-ATTR-001 | (c) | Attribution line in each of the 3 new reference files |
| AC-TPL-001 | (c) | `make build` exits 0; both skills' catalog hashes changed |
| AC-TPL-002 | (c) | 5 local mirrors byte-identical to template source |
| AC-TPL-003 | (c) | Neutrality greps land zero in new bodies |
| AC-TPL-004 | (c) | SKILL.md ≤ 500 lines each; routing frontmatter untouched |
| AC-VERIFY-001 | (c) | Run-phase: importer constraints checked against lane-6's landed t166 code (or gap recorded) |

### AC-DPI-001 — profile reference contract

Given the template tree, When `TPL/moai-domain-design-dna/references/diagram-profiles.md`
is read, Then it exists and `grep -c` finds: (i) the marker-first resolution
statement (a project-root marker names the slug; the snapshot is read in place;
no marker → no profile is guessed), (ii) the slug grammar, (iii) the
project-scoped storage location outside the skill directory, and (iv) the
load-time schema check with explicit "not observed" backfill — each ≥ 1 match.

### AC-DPI-002 — SKILL.md pointer, phases intact

Given `TPL/moai-domain-design-dna/SKILL.md`, When `grep -c "diagram-profiles.md"`
runs, Then ≥ 1; When `grep -c "Phase 2 — Analyze"` and `grep -c "Phase 3 — Generate"`
run, Then each ≥ 1 (the pointer is additive; the existing flow is not rewritten).

### AC-DPI-003 — write safety

Given diagram-profiles.md, When `grep -ci` runs for the overwrite-confirmation
rule and the verify-by-re-read rule, Then each ≥ 1.

### AC-IMP-001 — mermaid importer

Given `TPL/moai-domain-svg-infographic/references/import-mermaid.md`, When read,
Then it exists and greps land for: untrusted-source treatment ("untrusted"),
and the never-carry-over list covering coordinates, colors, fonts, and layout.

### AC-IMP-002 — drawio importer containers

Given `TPL/moai-domain-svg-infographic/references/import-drawio.md`, When read,
Then it exists and greps land for: the four container shapes (`base64` +
`deflate` + the two embedded forms) and the decode-before-parse rule; and it
states compressed bytes are never read as structure.

### AC-IMP-003 — fidelity ledger

Given both importer files, When `grep -ci "ledger"` and greps for merged /
collapsed / dropped run, Then each file carries them; and each file states the
ledger records a `check-svg.mjs` pass (zero errors) per migrated diagram.

### AC-IMP-004 — bulk replace / one home

Given both importer files, When greps for the replacement obligation run
("one home" or equivalent + same-change replacement phrasing), Then both files
carry it and neither offers a keep-both option.

### AC-IMP-005 — connector obligations by construction

Given both importer files, When greps run for `r = 8`, `6–10` (or "6-10"),
`>= 12` (or "≥ 12"), a paint-order mention, and a citation of
`authoring.md` §2.5 (or "six mandatory connector rules"), Then each lands in
both files.

### AC-IMP-006 — opt-in discoverability, default unchanged

Given `TPL/moai-domain-svg-infographic/SKILL.md`, When (a) the
bundled-references table is grepped for both importer filenames plus an opt-in
marker ("opt-in" or equivalent), Then both rows present; and When (b) the
six-step workflow headings, the Step 0 routing table rows, and the four-dial
table are grepped, Then all are present unchanged in substance (additive rows
only — verified by reading the section, not by byte-diff, since the table gains
rows).

### AC-ATTR-001 — attribution

Given the 3 new reference files, When `grep -c "cathrynlavery/diagram-design"`
and `grep -c "MIT"` run per file, Then each file ≥ 1 for both.

### AC-TPL-001 — build + catalog

Given all edits landed, When `make build` runs, Then exit 0; When the catalog
hash entries for `moai-domain-design-dna` and `moai-domain-svg-infographic`
are compared against the §C pre-flight baseline, Then both changed (both
SKILL.md roots changed, so an unchanged hash is a regen failure, not a no-op).

### AC-TPL-002 — mirror parity

Given the 5 changed template files, When each is `diff`ed against its local
`.claude/skills/` mirror, Then all 5 byte-identical (exit 0 per diff).

### AC-TPL-003 — neutrality

Given the 3 new reference files (bodies only), When `grep -rn "SPEC-\|t167\|lane-"`
runs over them, Then 0 matches; When a date-shaped grep runs, Then matches are
confined to examples a user would write in their own profile data, never to
moai-adk-internal chronology.

### AC-TPL-004 — body budget + routing untouched

Given both SKILL.md files after edit, When `wc -l` runs, Then each ≤ 500; When
the `description:` and `when_to_use:` blocks are diffed against the pre-change
versions, Then byte-identical.

### AC-VERIFY-001 — t166 landed-code alignment (run-phase gate)

Given lane-6's t166 has landed on the integration base, When the landed
`check-svg.mjs` geometry checks are read, Then the importer files' restated
constraints agree with the landed check semantics (same numeric thresholds for
mask gap, attach spacing, paint-order/label placement); any disagreement is
corrected in the importer files before close. Where t166 has NOT landed by
M3, Then this AC records the gap (with the observed 0-match grep) and the
ledger's check-svg pass is verified against the t165 checker instead — an
honest partial, never a silent PASS.

## §D.1 Severity

- **Must-pass**: AC-DPI-001..003, AC-IMP-001..006, AC-ATTR-001, AC-TPL-001,
  AC-TPL-003 — these are the contracts the card bought.
- **Standard**: AC-TPL-002, AC-TPL-004.
- **Conditional**: AC-VERIFY-001 — gated on lane-6 landing; the gap path is a
  documented outcome, not a failure of this SPEC.

## §D.2 Traceability

| AC | REQ |
|----|-----|
| AC-DPI-001 | REQ-1, REQ-2, REQ-3, REQ-4 |
| AC-DPI-002 | REQ-1 |
| AC-DPI-003 | REQ-5 |
| AC-IMP-001 | REQ-6, REQ-7 |
| AC-IMP-002 | REQ-8 |
| AC-IMP-003 | REQ-9, REQ-11 (ledger half) |
| AC-IMP-004 | REQ-10 |
| AC-IMP-005 | REQ-11 (construction half) |
| AC-IMP-006 | REQ-6, REQ-12 |
| AC-ATTR-001 | REQ-13 |
| AC-TPL-001 | REQ-14 |
| AC-TPL-002 | REQ-14 |
| AC-TPL-003 | REQ-15 |
| AC-TPL-004 | REQ-6 (default unchanged) |
| AC-VERIFY-001 | REQ-11 + spec §5 t166 constraint |

Every REQ-1..REQ-15 is covered by ≥ 1 AC.

## §D.3 Indirect verification

Profile persistence and importer pipelines are prose contracts; their runtime
behavior is exercised the next time a session uses the skills. The ACs above
verify the CONTRACT TEXT is present, correct, neutral, attributed, mirrored,
and catalogued — the strongest mechanical evidence available for
template-content work (t165 precedent: 13 of 14 criteria closed the same way).

## §D.4 Closure gates

- All must-pass ACs PASS with observed outputs (grep counts, diff exits, build
  exit code).
- progress.md §E.2/§E.3 populated with per-AC evidence per the attribution
  discipline (command + verbatim output + this-run baseline).
- AC-VERIFY-001 either aligned or gap-recorded.

## §D.5 Edge cases to check while implementing

- A profile JSON missing an entire dimension (backfill says "not observed",
  never a fabricated dimension — REQ-4).
- A mermaid source with subgraphs (groups) — IR keeps the grouping, layout
  recomputes containment per the numeric layout pass.
- A drawio source in the base64-deflate container opened as text — the
  reference must make the mojibake path impossible to reach by accident.
- An import whose source exceeds the `faithful` 24-node ceiling — ledger
  records zone/split, not quiet truncation.
- A slug collision on save — confirmation, then verify-by-re-read (REQ-5).

## §D.6 Forward-looking checks

- The importer references must not name t166 diagnostics by code number (they
  do not exist on this branch); if a later edit hard-codes them, that edit
  owns keeping them in sync.
- If B-10 (icon normalization) is ever adopted, it arrives as its own change
  with THIRD_PARTY notice — nothing in these files may pre-vendor icons.

## §D.7 Definition of Done

5 template files + catalog landed, 5 mirrors byte-identical, `make build`
green, neutrality greps zero, 15 ACs evidenced in progress.md §E, HISTORY
updated to 1.0.0 at close.
