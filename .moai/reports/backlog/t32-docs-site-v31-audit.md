# t32 — docs-site v3.1 remainder: re-measurement

Read-only re-measurement of the four unaddressed findings on backlog card t32,
against the current tree. No files were edited.

- **Repo**: `/Users/goos/MoAI/moai-adk-go`, branch `main`, HEAD `3b9b3bf99`
- **Baseline the original audit predates**: PRs #1532, #1533, #1535 have merged since
- **Locale trees**: `docs-site/content/{en,ko,ja,zh}` — 143 `.md` files each (file-existence parity is clean)

---

## B5 — locale derivation gap

**Status: STILL OPEN** (all 8 priority pages unchanged; the gap is larger than the card records)

### Evidence — the 8 named pages

Section counts are `grep -c '^#\{2,\} '` (H2 and deeper); line counts are `wc -l`.

```
page                                          ko       en       ja       zh
multi-llm/model-policy                        24/349   13/153   13/150   13/144
multi-llm/cg-mode                             23/281   18/206   18/192   18/183
cost-optimization/prompt-caching              16/291    8/134    8/121    8/85
getting-started/introduction                  27/296   24/288   23/278   23/278
advanced/harness-profiles                     12/187   10/115   10/99    10/99
advanced/no-haiku-3tier                       12/136    9/91     9/91     9/91
advanced/analyze-first-routing                12/122   10/79    10/79    10/79
claude-code/agentic/best-practices            31/279   24/216   24/216   24/216
```

All 8 still have ko ahead of en/ja/zh on both section count and line count. None
was touched by #1532/#1533/#1535. The three worst gaps by content volume are
`prompt-caching` (291 ko lines vs 85 zh — a 3.4× spread), `model-policy` (349 vs
144), and `cg-mode` (281 vs 183).

### Evidence — the derivation gap is 16 pages, not 8

A whole-tree scan finds **16** pages where ko carries more sections than en. The
8 above are the largest, but the full ko-ahead set is:

```
advanced/analyze-first-routing          ko=12 en=10
advanced/harness-profiles               ko=12 en=10
advanced/no-haiku-3tier                 ko=12 en=9
advanced/token-budget                   ko=9  en=8
advanced/ultracode-workflows            ko=20 en=19
claude-code/agentic/best-practices      ko=31 en=24
claude-code/extensibility/plugins       ko=12 en=11
claude-code/extensibility/skills        ko=13 en=11
claude-code/foundations/features-overview ko=11 en=10
claude-code/foundations/interactive-mode  ko=19 en=17
cost-optimization/prompt-caching        ko=16 en=8
getting-started/introduction            ko=27 en=24
multi-llm/_index                        ko=11 en=5
multi-llm/cg-mode                       ko=23 en=18
multi-llm/model-policy                  ko=24 en=13
utility-commands/moai-codemaps          ko=24 en=23
```

`multi-llm/_index` (ko=11 en=5) is a section landing page not on the card's list
and is the second-largest proportional gap in the set.

### Evidence — the size of the underived rewrite

```
$ git diff --name-only b50c4de71~1 b50c4de71 -- docs-site/content/ko | wc -l
132                                    # M1-M5 "KO SSOT" rewrite touched 132 ko files

$ git diff --name-only 734ede821~1 734ede821 -- docs-site/content   # M6 derivation
  44 en
  14 ja
  14 zh
```

M6 derived 44 en pages against a 132-page ko rewrite, and only 14 each for ja/zh.
That asymmetry is the mechanical origin of the gap.

### Scope of work remaining

Derive en from ko for the 16 pages above, then ja/zh from en. Priority order by
content volume: `prompt-caching`, `model-policy`, `cg-mode`, `multi-llm/_index`,
`best-practices`, `introduction`, then the nine smaller deltas.

### A second gap the card does not record (counter-direction)

The ko v3.1 rewrite **condensed** a number of pages that en/ja/zh never followed,
so those locales still carry the pre-v3.1 sprawl. The largest:

```
advanced/hooks-guide      ko=7   en=49 ja=49 zh=49
advanced/settings-json    ko=16  en=73 ja=73 zh=73
advanced/statusline       ko=13  en=36 ja=36 zh=36
advanced/claude-md-guide  ko=20  en=37 ja=37 zh=37
advanced/security-notes   ko=12  en=30 ja=30 zh=30
advanced/skill-guide      ko=15  en=30 ja=30 zh=30
advanced/decision-memory  ko=9   en=29 ja=32 zh=29
advanced/agent-guide      ko=21  en=29 ja=29 zh=29
```

These are not "en behind ko" — they are en/ja/zh holding superseded structure.
A fix pass that only adds sections to en will not close them; they need the ko
structure applied, i.e. a rewrite, not a top-up. Flagging so the fix pass does
not mis-scope the work as uniformly additive.

---

## B6 — sidebar invisibility

**Status: STILL OPEN** (all 7 pages still absent; the icon concern is a non-issue)

### Evidence — absent from `data/menu/main.yaml`

```
$ comm -13 <(grep -o 'ref: /advanced/[a-z0-9-]*' data/menu/main.yaml | sed 's|.*/||' | sort -u) \
           <(ls content/ko/advanced/*.md | ... | sort -u)
autonomy-tier
bas-navigator
harness-learning
kanban-mode
manager-kanban
multi-model-audit
mx-scanner-internals
```

Exactly the 7 named. `main.yaml` carries 24 `/advanced/*` refs against 31 pages
on disk. The reverse check (menu refs with no ko page) returns empty — no dangling
refs, so the file is under-populated rather than stale.

### Evidence — NEW-badge targets

`new_items` in each locale's `content/<loc>/advanced/_meta.yaml` is identical
across all four locales:

```
- kanban-mode
- bas-navigator
- manager-kanban
- multi-model-audit
- autonomy-tier
```

**5** of these — not 4 — are in the missing-from-menu set (all five are). Since
`layouts/partials/menu.html` renders the NEW badge only inside the `range .sub`
loop over `main.yaml` entries (lines 76-102), a page absent from `main.yaml` has
no `<li>` at all, so its badge cannot render. All five v3.1 NEW badges are
currently unrenderable.

### Evidence — icon coupling is NOT a blocker

`layouts/partials/menu.html` reads `.icon` only at **section** level (line 48,
inside the `<summary>`), against a 17-case map (`home`, `star`, `plug`, `book`,
`chef`, `palette`, `tag`, `flash_on`, `build`, `check_circle`, `database`,
`fork_right`, `handshake`, `map`, `school`, `shuffle`, `terminal`). Sub-items
carry no icon field. The `advanced` section already uses `icon: school`.

**No new SVG case is needed.** Adding the 7 pages is a pure `main.yaml` edit.

### Evidence — a related `_meta.yaml` defect found while measuring

Title entries in `advanced/_meta.yaml` are ko-only for the v3.1 pages:

```
key counts:  en=23  ko=29  ja=23  zh=23

en/ja/zh each missing title keys for:
  autonomy-tier, bas-navigator, harness-learning, kanban-mode,
  manager-kanban, multi-model-audit, mx-scanner-internals, config-sections
ko missing:  config-sections, mx-scanner-internals
```

So en/ja/zh `_meta.yaml` list slugs under `new_items` that have no corresponding
title entry in the same file. Not fatal (Hugo falls back to page frontmatter),
but it means the 4-locale `_meta` sets are not congruent, and `config-sections`
has no title entry in **any** locale.

### Scope of work remaining

1. Add 7 `sub:` entries under the `/advanced` section of `data/menu/main.yaml`,
   each with the full 4-locale `name` map (ko titles already exist in
   `content/ko/advanced/_meta.yaml`; en/ja/zh titles must be authored).
2. Backfill the 8 missing title keys into `en/ja/zh advanced/_meta.yaml` and the
   2 missing into ko.
3. No `menu.html` change.

---

## B7 — `moai epic status` undocumented

**Status: STILL OPEN**

### Evidence — the command exists and is live

```
$ grep -n "AddCommand" internal/cli/epic.go
38:	cmd.AddCommand(newEpicStatusCmd())
122:	rootCmd.AddCommand(newEpicCmd())

$ go run ./cmd/moai epic --help
  COMMANDS
    status <prefix> [--flags]  Compute epic milestone progress from disk
  FLAGS
    -h --help
```

Producer at `internal/cli/epic.go:66` (`newEpicStatusCmd`). Flags: `--json`,
`--design-report`, `--marker`, `--base-dir`, `--locale`. Observation-only; no
file writes.

### Evidence — zero documentation

```
$ grep -rl "epic status" docs-site/content/ | wc -l
0
$ grep -rn "moai epic" docs-site/content/
(no output)
```

`content/ko/cli-reference/` holds 21 command pages (`ast-grep`, `constitution`,
`doctor`, `github`, `goal`, `handoff`, `harness`, `init`, `inventory`,
`launchers`, `loop`, `pr`, `profile`, `session`, `spec`, `status`, `tool-policy`,
`update`, `web`, `worktree`). No `epic.md`.

### Scope of work remaining

One new page `cli-reference/epic.md` × 4 locales, plus a `_meta.yaml` title entry
× 4 and a `data/menu/main.yaml` sub-entry with the 4-locale name map. Content
source: the `Long` help text at `internal/cli/epic.go:66-75` and the flag set.

---

## B8 — `branch_guard` config key undocumented

**Status: STILL OPEN**

### Evidence — the key's current name and location

```
internal/config/types.go:397   BranchGuard BranchGuardConfig `yaml:"branch_guard"`
internal/config/types.go:555   // BranchGuardConfig mirrors workflow.branch_guard.*
internal/settings/schema_sections.go:329
     s(SectionWorkflow, "workflow", TypeBool, "workflow", "branch_guard", "enabled")
.moai/config/sections/workflow.yaml:135   branch_guard:
```

Canonical key: **`workflow.branch_guard.enabled`** (bool). Default is OFF —
`internal/config/defaults_test.go:32` asserts `Workflow.BranchGuard.Enabled` is
false, citing "REQ-1 default-off / REQ-4 template neutrality". It is surfaced in
the web console (`internal/web/schemaform.go:151` special-cases it alongside
`workflow.worktree.*`), so it is a genuinely user-facing toggle.

### Evidence — zero documentation

```
$ grep -rl "branch_guard" docs-site/content/ | wc -l
0
$ grep -c branch_guard content/ko/advanced/settings-json.md \
                       content/en/advanced/settings-json.md \
                       content/ko/advanced/config-sections.md
0 / 0 / 0
```

The only `BranchGuard` string in the docs tree is in
`advanced/autonomous-loops.md:113` (all 4 locales), where it appears as a
*pattern* reference — "BranchGuard pattern sibling to
`workflow.codex.review_gate`, default off" — describing `multi_review_gate`. It
documents neither the key nor the guard.

### Scope of work remaining

A `workflow.branch_guard.enabled` subsection in `advanced/config-sections.md`
× 4 locales (the natural host — it is the config-key reference page). Content
source: `.claude/rules/moai/workflow/main-checkout-branch-guard.md`
(v1.2.0 §Mechanical Enforcement) for behavior, the opt-in rationale, the
`MOAI_BRANCH_GUARD_EXEMPT` escape hatch, and the fail-open norm.

---

## Gate weakness — the 4-locale parity check

**Status: CONFIRMED, with one correction to the original framing**

### Evidence — what the gate actually checks

The gate is the inlined recipe in `.claude/skills/hns-oss-docs-verify/SKILL.md`
§4 (`locale-parity`, must_pass, threshold 1.0). There is no gate script in
`docs-site/scripts/`. It runs three checks:

```bash
# (1) file-existence parity — per page
for f in $(cd ko && find . -name '*.md'); do
  for loc in en ja zh; do [ -f "$loc/$f" ] || echo "MISSING: $loc/$f"; done
done

# (2) section-count parity — TREE TOTALS ONLY
for loc in ko en ja zh; do
  printf '%s: %s\n' "$loc" "$(grep -rc '^## ' docs-site/content/$loc --include='*.md' | awk -F: '{s+=$2} END {print s}')"
done

# (3) README 4-file heading-count parity
grep -c '^## ' README.md README.ko.md README.ja.md README.zh.md
```

**Correction to the card's framing**: the gate does have a section-count check —
it is not absent. Its weakness is twofold:

1. **It aggregates over the whole tree.** Per-page divergences in opposite
   directions cancel. The 16 ko-ahead pages are partly netted out by the ko-behind
   pages (hooks-guide, settings-json, etc.), so a large per-page disagreement can
   present as a small total delta.
2. **It has no stated pass criterion.** The step says only "compare totals". The
   "Expected: identical counts" line applies to check (3), the README check. So
   even a visible total mismatch has nothing to fail against.

Run against the current tree, the totals are in fact **not** identical:

```
ko: 1200   en: 1166   ja: 1119   zh: 1118
```

Check (1) is clean — 143 files per locale, no MISSING output — which is what
keeps the gate green.

### Current structural-divergence count

Per-page comparison of ko against all three derived locales:

```
H2-and-deeper (^#{2,}):  143 pages, 64 divergent
H2-only (^## , matching the gate's own grep):  143 pages, 56 divergent
```

The original audit reported 67 of 142. The **64 of 143** figure is the
like-for-like re-measurement: 3 pages resolved since, 1 page added.

### Scope of work remaining

Replace check (2) with a per-page comparison and give it an explicit failure
condition. A single-pass form (the one used for this report) runs in a few
seconds over the whole tree; a naive per-file loop over 143×4 files does not
complete within a 2-minute budget, so the awk form matters.

---

## Prioritized work list

1. **`data/menu/main.yaml`** — add the 7 missing `/advanced/*` sub-entries with
   4-locale name maps. Unblocks 5 unrenderable v3.1 NEW badges. No `menu.html`
   change needed. *(smallest edit, largest user-visible effect)*
2. **`advanced/_meta.yaml` × 4** — backfill the 8 missing title keys in en/ja/zh
   and the 2 in ko, so the four `_meta` sets are congruent.
3. **B5 additive derivation, 6 largest** — derive en then ja/zh for
   `cost-optimization/prompt-caching`, `multi-llm/model-policy`,
   `multi-llm/cg-mode`, `multi-llm/_index`,
   `claude-code/agentic/best-practices`, `getting-started/introduction`.
4. **B5 additive derivation, remaining 10** — the smaller ko-ahead deltas.
5. **B8** — `workflow.branch_guard.enabled` section in
   `advanced/config-sections.md` × 4.
6. **B7** — new `cli-reference/epic.md` × 4 + `_meta.yaml` + `main.yaml` entry.
7. **Gate fix** — per-page section-count comparison with an explicit failure
   condition, replacing the tree-total check in
   `.claude/skills/hns-oss-docs-verify/SKILL.md` §4.
8. **B5 counter-direction (separate scope)** — apply the condensed ko v3.1
   structure to en/ja/zh for the 8 pages still carrying pre-v3.1 sprawl
   (`hooks-guide`, `settings-json`, `statusline`, `claude-md-guide`,
   `security-notes`, `skill-guide`, `decision-memory`, `agent-guide`). This is
   rewrite work, not top-up work, and should not be folded into items 3-4.
