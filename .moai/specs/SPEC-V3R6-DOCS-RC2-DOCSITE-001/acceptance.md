---
id: SPEC-V3R6-DOCS-RC2-DOCSITE-001
title: "v3.0.0-rc2 docs-site version SSOT + lifecycle/era + harness-namespace — acceptance criteria"
version: "0.2.0"
status: in-progress
created: 2026-06-19
updated: 2026-06-22
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "docs-site"
lifecycle: spec-anchored
tags: "docs-site, version-ssot, 4-locale, lifecycle, era, harness-namespace, i18n"
era: V3R6
---

# Acceptance Criteria — SPEC-V3R6-DOCS-RC2-DOCSITE-001

## §A. Verification Philosophy

Every AC below is **independently reproducible** by running the exact grep/shell command in the **Evidence** column against the post-change tree. The verifier MUST observe the literal expected output (not a summary, not a carry-over from a prior SPEC) per the verification-claim-integrity doctrine (`evidence absent ≠ evidence of success`).

The baseline (pre-change) values are recorded in spec.md §B.4. The verifier reproduces the baseline first, then applies the change, then reproduces the post-change Evidence. A PASS requires the post-change Evidence to match.

## §B. Severity Convention

- **MUST-PASS**: blocks merge / sync-phase close if FAIL.
- **SHOULD-PASS**: does not block merge; surfaces as debt if FAIL.
- **INFO**: observed-only; no PASS/FAIL verdict.

Every AC below labels its severity. The MUST-PASS set is the sync-phase close gate.

## §C. Traceability Matrix

| AC ID | Severity | REQ binding | Milestone |
|-------|----------|-------------|-----------|
| AC-VLN-001 | MUST-PASS | REQ-VLN-001 | M1 |
| AC-VLN-002 | MUST-PASS | REQ-VLN-002 | M1, M2 |
| AC-VLN-003 | MUST-PASS | REQ-VLN-003 | M2 |
| AC-VLN-004 | MUST-PASS | REQ-VLN-004 | M2 |
| AC-LCY-001 | MUST-PASS | REQ-LCY-001 | M3 |
| AC-LCY-002 | MUST-PASS | REQ-LCY-002 | M4 |
| AC-LCY-003 | SHOULD-PASS | REQ-LCY-003 | M3 → M4 |
| AC-HNS-001 | MUST-PASS | REQ-HNS-001 | M5 |
| AC-HNS-002 | MUST-PASS | REQ-HNS-002 | M5 |
| AC-HNS-003 | MUST-PASS | REQ-HNS-003 | M5 |
| AC-PAR-001 | MUST-PASS | REQ-PAR-001 | M6 |
| AC-PAR-002 | MUST-PASS | REQ-PAR-002 | M2, M4, M5 |
| AC-NTR-001 | MUST-PASS | REQ-NTR-001 | M6 |
| AC-NTR-002 | MUST-PASS | REQ-NTR-002 | M6 |
| AC-VWD-001 | MUST-PASS | REQ-VWD-001 | M1, M2 |

## §D. Acceptance Criteria

### AC-VLN-001 — hugo.toml version SSOT aligned to v3.0.0-rc2 (MUST-PASS)

**Given** the docs-site `hugo.toml` `[params]` block,
**When** the verifier runs `grep -nE '^  version|^  releaseDate' docs-site/hugo.toml`,
**Then** the `version` line reads `version = "v3.0.0-rc2"` AND the `releaseDate` line reads `releaseDate = "2026-06-03"` (the verified rc2 tag date: `git log -1 --format=%ci v3.0.0-rc2` → `2026-06-03 08:00:59 +0900`).

**Evidence (post-change)**:
```bash
grep -nE '^  version|^  releaseDate' docs-site/hugo.toml
# Expected:
#   55:  version = "v3.0.0-rc2"
#   56:  releaseDate = "2026-06-03"

grep -nE '^  version' docs-site/hugo.toml | grep -c 'v3\.0\.0-rc2'
# Expected: 1

grep -nE '^  releaseDate' docs-site/hugo.toml | grep -c '2026-06-03'
# Expected: 1
```

**Baseline (pre-change)**: L55 `version = "0.2.0"`; L56 `releaseDate = "2026-05-21"`.

### AC-VLN-002 — `{{< version >}}` shortcode wired into ≥1 page per locale (MUST-PASS)

**Given** the `{{< version >}}` shortcode exists at `docs-site/layouts/shortcodes/version.html`,
**When** the verifier greps for shortcode usage,
**Then** at least one page in EACH of the four locales references `{{< version >}}` (intro `_index.md` and/or `getting-started/installation.md`).

**Evidence (post-change)**:
```bash
for loc in ko en ja zh; do
  echo -n "$loc: "
  grep -rln '{{< version >}}' docs-site/content/$loc/ | wc -l | tr -d ' '
done
# Expected: each locale ≥ 1
```

**Baseline**: `0` (dead shortcode, no references).

### AC-VLN-003 — 4-locale installation pages: stale literals removed, v3.0.0-rc2 identical (MUST-PASS)

**Given** the 4-locale installation pages,
**When** the verifier greps for ALL stale version literals (`2.9.0` or `2.0.0`) across the four installation pages,
**Then** the only remaining hit is the grandfathered historical-license mention at ko L16 (`v2.0.0부터 Go 언어로 재작성`), and the install command presents `v3.0.0-rc2` identically across locales.

**Evidence (post-change)**:
```bash
# ALL stale version literals across the 4 installation pages (install commands + prose + output examples):
grep -rn '2\.9\.0\|2\.0\.0' docs-site/content/*/getting-started/installation.md
# Expected: exactly 1 match — ko L16 historical-license context:
#   docs-site/content/ko/getting-started/installation.md:16:... v2.0.0부터 Go 언어로 재작성되며 ...
# All other stale literals (ko L11/L91/L157; en/ja/zh L81) removed.

# v3.0.0-rc2 present in each locale's installation page:
for loc in ko en ja zh; do
  echo -n "$loc: "
  grep -c 'v3\.0\.0-rc2' docs-site/content/$loc/getting-started/installation.md
done
# Expected: each locale ≥ 1
```

**Baseline (pre-change)**: ko L11 `v2.9.0` (Apache-2.0 prose), L91 `--version 2.9.0` (install command), L157 `moai v2.9.0` (output example), L16 `v2.0.0` (grandfathered historical-license — preserved); en/ja/zh L81 `--version 2.0.0` (install command). Total pre-change stale-literal hits = 7 (6 to remove + 1 grandfathered).

### AC-VLN-004 — atomic 4-locale commit for version fix (MUST-PASS)

**Given** the version SSOT + installation fix work lands,
**When** the verifier inspects the commit history for the version-fix commit,
**Then** a single commit contains the `hugo.toml` change AND all 4-locale installation page changes together (no intermediate commit where some locales are updated and others are not).

**Evidence (post-change)**:
```bash
# Locate the commit via the installation-page change (more stable than hugo.toml
# pickaxe — hugo.toml may be edited more than once, e.g. a releaseDate follow-up,
# which would bind -S 'version = "v3.0.0-rc2"' to the wrong commit). The install-command
# literal change is unique to this SPEC's version-fix commit.
COMMIT=$(git log --format='%H' -1 -S '--version 2.0.0' -- docs-site/content/en/getting-started/installation.md)
echo "version-fix commit: $COMMIT"

# Verify hugo.toml is in the SAME commit's --stat:
git show --stat $COMMIT | grep -c 'hugo\.toml'
# Expected: 1

# Verify all 4 locale installation pages are in the SAME commit's --stat:
git show --stat $COMMIT | grep -E 'content/(ko|en|ja|zh)/getting-started/installation\.md' | wc -l
# Expected: 4 (one per locale)
```

**Note**: the run-phase SHOULD land `hugo.toml` `version` + `releaseDate` in one shot (no separate releaseDate follow-up commit) to keep the pickaxe unambiguous; see plan.md §C pre-flight item 2.

### AC-LCY-001 — lifecycle/era section present in Korean (MUST-PASS)

**Given** the lifecycle/era concept section is authored in Korean,
**When** the verifier greps `content/ko/core-concepts/spec-based-dev.md`,
**Then** the file contains the 5-bucket era taxonomy, the grandfather clause, the 3-phase close, the `sync_commit_sha` reference, and the Mx-phase retirement.

**Evidence (post-change)**:
```bash
FILE=docs-site/content/ko/core-concepts/spec-based-dev.md
for term in 'grandfather' 'era_final' 'V3R6' 'V3R5' 'sync_commit_sha' 'Mx-phase' 'plan' 'run' 'sync'; do
  echo -n "$term: "
  grep -c "$term" "$FILE"
done
# Expected: each term ≥ 1 (the 5-bucket taxonomy requires all four V2.x/V3R2-R4/V3R5/V3R6/unclassified markers)
```

**Baseline**: `0` for each term (section absent).

### AC-LCY-002 — lifecycle/era section present in en/ja/zh, structurally identical (MUST-PASS)

**Given** the lifecycle/era section is translated from the finalized Korean,
**When** the verifier greps all four locales,
**Then** each locale's `core-concepts/spec-based-dev.md` contains the same concept set (era buckets, grandfather clause, 3-phase close, sync_commit_sha, Mx-phase retirement).

**Evidence (post-change)**:
```bash
for loc in ko en ja zh; do
  FILE=docs-site/content/$loc/core-concepts/spec-based-dev.md
  echo "=== $loc ==="
  for term in 'grandfather' 'era_final' 'sync_commit_sha' 'Mx-phase'; do
    # Note: 'grandfather'/'era_final' are technical terms likely preserved verbatim across locales.
    echo -n "  $term: "
    grep -c "$term" "$FILE"
  done
done
# Expected: each locale each term ≥ 1 (the technical terms are kept verbatim; surrounding prose is localized)
```

**Note**: the four locales' sections must have the same HEADING STRUCTURE. The verifier manually confirms heading-level parity (e.g., the era-classification H2/H3 headings appear in all four).

### AC-LCY-003 — ko-first authoring order preserved (SHOULD-PASS)

**Given** the ko-first workflow constraint,
**When** the verifier inspects the commit history for the lifecycle/era section,
**Then** the ko section commit precedes (or is co-located in an atomic commit with) the en/ja/zh translation commit — NOT the reverse.

**Evidence (post-change)**:
```bash
KO_COMMIT=$(git log --format='%H %ci' -1 -S 'grandfather' -- docs-site/content/ko/core-concepts/spec-based-dev.md | awk '{print $1}')
EN_COMMIT=$(git log --format='%H %ci' -1 -S 'grandfather' -- docs-site/content/en/core-concepts/spec-based-dev.md | awk '{print $1}')
echo "ko: $KO_COMMIT"
echo "en: $EN_COMMIT"
# Expected: ko commit timestamp ≤ en commit timestamp (ko authored first or same atomic commit)
```

**Severity rationale**: SHOULD-PASS because the ko-first constraint is a QUALITY/process discipline; a parallel-from-scratch execution that still produces correct 4-locale content does not break user-facing value, only the process discipline.

### AC-HNS-001 — harness-namespace section present in Korean (MUST-PASS)

**Given** the harness-namespace policy section is authored in Korean,
**When** the verifier greps `content/ko/core-concepts/harness-engineering.md`,
**Then** the file contains the template-managed (`moai-*`, `moai-harness-*`, `moai-meta-harness`) vs user-owned (`harness-*`) distinction.

**Evidence (post-change)**:
```bash
FILE=docs-site/content/ko/core-concepts/harness-engineering.md
for term in 'harness-*' 'moai-harness' 'moai-meta-harness' 'template-managed' 'user-owned'; do
  echo -n "$term: "
  grep -c "$term" "$FILE"
done
# Expected: each term ≥ 1
```

**Baseline**: section absent.

### AC-HNS-002 — harness-namespace section present in en/ja/zh (MUST-PASS)

**Given** the harness-namespace section is translated,
**When** the verifier greps all four locales,
**Then** each locale's `core-concepts/harness-engineering.md` contains the template-managed vs user-owned distinction.

**Evidence (post-change)**:
```bash
for loc in ko en ja zh; do
  FILE=docs-site/content/$loc/core-concepts/harness-engineering.md
  echo -n "$loc moai-harness: "
  grep -c 'moai-harness' "$FILE"
  echo -n "$loc harness-*: "
  grep -c 'harness-\*' "$FILE"
done
# Expected: each locale each term ≥ 1
```

### AC-HNS-003 — cites final doctrine SPEC, not SUPERSEDED (MUST-PASS)

**Given** the SUPERSEDED `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` is retired,
**When** the verifier greps the docs-site content for both SPEC IDs,
**Then** `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` appears (the final doctrine) and `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` does NOT appear.

**Evidence (post-change)**:
```bash
echo -n "V2 (final) cited: "
grep -rln 'SPEC-V3R6-HARNESS-NAMESPACE-V2-001' docs-site/content/ | wc -l | tr -d ' '
# Expected: ≥ 1

echo -n "SUPERSEDED cited: "
grep -rln 'SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001' docs-site/content/ | wc -l | tr -d ' '
# Expected: 0
```

### AC-PAR-001 — 4-locale page-set parity preserved (MUST-PASS)

**Given** the hard 4-locale parity invariant,
**When** the verifier counts `.md` files per locale after all additions,
**Then** the per-locale counts are identical across ko/en/ja/zh.

**Evidence (post-change)**:
```bash
for loc in ko en ja zh; do
  echo -n "$loc: "
  find docs-site/content/$loc -name '*.md' | wc -l | tr -d ' '
done
# Expected: all four identical (105 if both additions are in-file sections; 106 if a new page was added × 4)
```

**Baseline**: `105 105 105 105`.

### AC-PAR-002 — atomic 4-locale commits for additive content (MUST-PASS)

**Given** the atomic-commit-per-addition constraint,
**When** the verifier inspects the commit history for each new section,
**Then** the lifecycle/era translation commit and the harness-namespace translation commit each contain all four locales together (no intermediate commit with partial locales).

**Evidence (post-change)**:
```bash
# Lifecycle/era translation commit:
LCY_COMMIT=$(git log --format='%H' -1 -S 'grandfather' -- docs-site/content/en/core-concepts/spec-based-dev.md)
git show --stat $LCY_COMMIT | grep -E 'content/(ko|en|ja|zh)/core-concepts/spec-based-dev\.md' | wc -l
# Expected: 4

# Harness-namespace commit:
HNS_COMMIT=$(git log --format='%H' -1 -S 'harness-*' -- docs-site/content/en/core-concepts/harness-engineering.md)
git show --stat $HNS_COMMIT | grep -E 'content/(ko|en|ja|zh)/core-concepts/harness-engineering\.md' | wc -l
# Expected: 4 (or 3 if ko landed separately as M5a — in that case verify ko commit + translation commit together restore parity)
```

### AC-NTR-001 — no NEW forbidden internal content leaked by this SPEC (MUST-PASS)

**Given** the template-neutrality-style content discipline (even though `docs-site/` is project-owned and the §25 CI guard does not bind) AND a known pre-existing baseline of legitimate `/Users/` paths (WSL path mapping `C:\Users\name\...` in windows-guide, worktree examples, faq, and installation pages across 4 locales — 21 files total, NONE of which this SPEC touches for `/Users/` removal),
**When** the verifier greps docs-site content for forbidden artifacts,
**Then** (a) the `/Users/` count does NOT exceed the pre-existing baseline of 21 (this SPEC adds zero new `/Users/` paths — its added sections are in `spec-based-dev.md` and `harness-engineering.md` which carry no `/Users/` content), AND (b) zero matches for `"v3.0.0 stable"` wording, AND (c) the two files this SPEC adds sections to (`spec-based-dev.md`, `harness-engineering.md`) contain zero `/Users/` paths post-change.

**Evidence (post-change)**:
```bash
# (a) /Users/ count does not exceed the pre-existing baseline of 21
#     (21 pre-existing legitimate hits across 4 locales × 5 file-types + 1 session-memo;
#      this SPEC removes none and adds none):
echo -n "/Users/ total (baseline 21, must not exceed): "
grep -rln '/Users/' docs-site/content/ | wc -l | tr -d ' '
# Expected: 21 (unchanged baseline — this SPEC adds no new /Users/ paths)

# (b) 'v3.0.0 stable' wording absent:
echo -n "'v3.0.0 stable' wording: "
grep -rln 'v3\.0\.0 stable' docs-site/content/ | wc -l | tr -d ' '
# Expected: 0

# (c) the two files this SPEC adds sections to carry no /Users/ paths:
for loc in ko en ja zh; do
  echo -n "$loc spec-based-dev /Users/: "
  grep -c '/Users/' docs-site/content/$loc/core-concepts/spec-based-dev.md
  echo -n "$loc harness-engineering /Users/: "
  grep -c '/Users/' docs-site/content/$loc/core-concepts/harness-engineering.md
done
# Expected: all 0
```

**Baseline (pre-change, reproduced 2026-06-19)**: `/Users/` = 21 files (4 locales × {installation.md, windows-guide.md, faq.md, worktree/examples.md, worktree/faq.md} + `content/.moai/state/session-memo.md`). These are legitimate `C:\Users\...` WSL-path / worktree-example references in user-facing guides; this SPEC does NOT remove them. The AC guards against NEW `/Users/` paths introduced by this SPEC's added sections, not the pre-existing baseline.

**Note**: the two named originating SPECs (`SPEC-V3R6-LIFECYCLE-REDESIGN-001`, `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`) are explicitly ALLOWED in user-facing content (they explain the Mx-phase retirement and the namespace doctrine provenance). Other internal SPEC IDs are forbidden.

### AC-NTR-002 — already-correct content unchanged (MUST-PASS)

**Given** the already-correct model/agent content (§B.2 of spec.md),
**When** the verifier greps for the current markers,
**Then** `glm-5.2[1m]` still has 2 hits/locale, the 8-retained-agent content is intact, and no stale `glm-5.1` references were reintroduced.

**Evidence (post-change)**:
```bash
for loc in ko en ja zh; do
  echo -n "$loc glm-5.2[1m]: "
  grep -rln 'glm-5\.2\[1m\]' docs-site/content/$loc/ | wc -l | tr -d ' '
done
# Expected: each locale ≥ 2

echo -n "stale glm-5.1: "
grep -rln 'glm-5\.1' docs-site/content/ | wc -l | tr -d ' '
# Expected: 0
```

### AC-VWD-001 — version wording exactly v3.0.0-rc2 (MUST-PASS)

**Given** the version-wording constraint (no stable `v3.0.0` tag exists),
**When** the verifier greps for the version string,
**Then** the docs-site uses `"v3.0.0-rc2"` and does NOT use `"v3.0.0 stable"` or imply a stable `v3.0.0` tag.

**Evidence (post-change)**:
```bash
echo -n "'v3.0.0 stable': "
grep -rln 'v3\.0\.0 stable' docs-site/content/ docs-site/hugo.toml | wc -l | tr -d ' '
# Expected: 0

echo -n "v3.0.0-rc2 in hugo.toml: "
grep -c 'v3\.0\.0-rc2' docs-site/hugo.toml
# Expected: ≥ 1
```

## §E. Edge Cases

1. **Shortcode in install command line (known second touch-point)**: Hugo shortcodes do NOT render inside fenced code blocks, so the install command (`curl ... | bash -s -- --version <VER>`) keeps its literal version and the `{{< version >}}` shortcode is wired into the surrounding PROSE/intro instead. This means a future version bump still requires editing `hugo.toml` + the fenced install-command literal ×4 locales — the shortcode reduces duplicate version literals in PROSE only, NOT in the fenced command. This is the "known second touch-point" referenced by REQ-VLN-002. The run-phase records which prose surface carries the shortcode in the commit body. AC-VLN-002 remains satisfied as long as ≥1 page per locale uses the shortcode.
2. **Interim ko-only commit parity**: M3 (ko-only) and M5a (ko-only) temporarily break 4-locale parity at the commit level. AC-PAR-001 is evaluated at end-of-SPEC (M6), not at every intermediate commit. If a CI parity gate is commit-time, the ko-only commits stay on the run-phase branch and only the atomic-4-locale translation commits merge to main.
3. **Mermaid diagram in lifecycle section**: if the era-classification table is rendered as a Mermaid diagram instead of a Markdown table, the diagram MUST stay under the hugo-geekdoc render ceiling (no oversized graph). The verifier confirms `hugo` build succeeds without render warnings on the new section.
4. **Historical version mentions in prose**: ko `installation.md` L16 mentions `v2.0.0` in a historical-license-change context ("v2.0.0부터 Go 언어로 재작성"). This is a legitimate historical reference, NOT an install command. AC-VLN-003 targets install-command lines (`--version N.N.N`), not historical prose. The verifier distinguishes the two.
5. **`releaseDate` exact value**: the rc2 tag date is `git log -1 --format=%ci v3.0.0-rc2`. If the tag is annotated, `%ci` returns the tagger date; if lightweight, the commit date. The run-phase uses whichever Hugo/Vercel conventionally renders; the AC checks that `releaseDate` is NOT the stale `2026-05-21` and IS a plausible rc2 date.

## §F. Definition of Done

All MUST-PASS AC (AC-VLN-001/002/003/004, AC-LCY-001/002, AC-HNS-001/002/003, AC-PAR-001/002, AC-NTR-001/002, AC-VWD-001) are independently reproducible by the verifier with the exact Evidence commands, against the post-change tree, in this run. SHOULD-PASS AC (AC-LCY-003) is reported but does not block close.

The run-phase §E.2 (Run-phase Evidence) and §E.3 (Run-phase Audit-Ready Signal) in plan.md are populated with observed evidence per the 5-Section Evidence-Bearing Report format, including the Gaps section (what was NOT observed).

## §G. Forward-Looking Checks (non-blocking, recorded for future SPECs)

- **v3.0.0 stable release**: when a stable `v3.0.0` tag is cut, a future SPEC will bump `hugo.toml` to `"v3.0.0"`, drop the `-rc2` suffix, and update the shortcode-wired pages automatically (since they reference `{{< version >}}`, only `hugo.toml` changes). The forward-looking check: confirm the shortcode wiring from AC-VLN-002 makes that future bump a one-line edit.
- **Mermaid render ceiling**: if the lifecycle/era Mermaid diagram (if used) is near the render ceiling now, future additions to the era table may push it over. Recorded as forward-looking; not actionable in this SPEC.
- **Sibling README SPEC coordination**: the sibling `SPEC-V3R6-DOCS-RC2-README-001` owns README + CHANGELOG. The two SPECs should close together as the v3.0.0-rc2 docs cohort. Forward-looking: confirm cohort-level parity (docs-site + README + CHANGELOG all report `v3.0.0-rc2`).
