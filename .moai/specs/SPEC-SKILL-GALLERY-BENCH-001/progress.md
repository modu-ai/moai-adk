# progress.md — SPEC-SKILL-GALLERY-BENCH-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-25

Tier M artifact set complete: spec.md + plan.md + acceptance.md authored by
manager-spec; run-phase evidence slots below are placeholders.

## §E.2 Run-phase Evidence

### M1 — 9-form benchmark execution + verdict table (2026-08-25)

**What ran.** All nine TypePack briefs of spec.md §D.2 were executed in the
pinned order through the skill's own method (numeric layout pass → SVG
authoring → G2 lint → G3 render), each with its own artifact and its own
gate logs. Environment preflight (plan.md §C): node v22.14.0; fixture render
probe exit 0 (Chrome 151.0.7922.174, `--headless=new`, IHDR 800x400
verified); evidence dir absent before run; branch `WT-skillstead-gallery`.

**What was observed.** 9/9 forms `PRODUCIBLE`, 0 PARTIAL, 0 NOT-PRODUCIBLE.
Per form: G1 artifact exists and is non-empty; G2 `check-svg.mjs` exit=0
with `0 errors, 0 warnings` (no warnings to triage — observed emptiness,
not omission); G3 `render.mjs` exit=0 with PNG IHDR dimensions matching the
2x target and the browser executable + version disclosed in every render
log. All nine PNGs were visually inspected (skill workflow step 6): no slop
symptoms, focal discipline held (≤ 1 accent element per diagram).
Recorded dial deviations: approval-gate + process-flow size fit (W 1660),
decision-matrix detail faithful-banded (19 nodes > balanced ceiling).
REQ-SGB-006 deviation sentences for all nine forms and the non-binding
follow-up observations live in `.moai/reports/t272/verdict.md`.

**Evidence tree** (all under `.moai/reports/t272/`): `verdict.md` (9 rows,
full REQ-SGB-004 schema incl. settled-dials), `layout-notes.md` (numeric
layout passes + lint calibration appendix), `artifacts/<form>.svg` ×9,
`artifacts/<form>.png` ×9, `logs/<form>-lint.txt` ×9 + `logs/<form>-render.txt`
×9 (each in the spec.md §D.2 log-format contract: command line + verbatim
output + `exit=N`).

**Skill immutability (REQ-SGB-008 / AC-007) at the M1 checkpoint:**
`git diff --stat origin/main -- .claude/skills/moai-domain-svg-infographic/ internal/template/templates/.claude/skills/moai-domain-svg-infographic/`
→ empty output.

**Gaps.** M2 (docs-site + README 4-locale emphasis) and M3 (verify recipe +
PR readiness) are not yet executed — this spawn's scope was M1 only.
G3 evidence is single-machine (one browser build, one run per form); no
cross-browser claim is made. KPI values in cards-kpi-grid are illustrative
and disclosed as such in the artifact's `<desc>` and in verdict.md.

### M2 — docs-site + README 4-locale emphasis (2026-08-25)

**What ran.** The producible-diagram-types emphasis authored in ko (canonical)
and derived into en/ja/zh, 8 files total: a new H3 section inserted inside
the Domain category of `docs-site/content/{ko,en,ja,zh}/advanced/skill-guide.md`
(after the Domain skill table, before the Reference heading) listing all 9
`PRODUCIBLE` forms with per-form evidence paths
(`.moai/reports/t272/artifacts/<form>.svg`, satisfying REQ-SGB-009's per-type
citation on the docs surface), and a new H3 feature entry in the 핵심 기능 /
Key Features section of `README.ko.md` + `README.md` + `README.ja.md` +
`README.zh.md` (after the ref/domain skills entry) naming the 9 kinds with a
single evidence pointer to `.moai/reports/t272/verdict.md` (REQ-SGB-011: ko
authored first, minimal-diff derivation, no new H2).

**What was observed.** Diff is purely additive: +18/-0 per skill-guide
(H3 + intro + 9-row table + closing), +4/-0 per README (H3 + paragraph).
Parity checks: README H2 12/12/12/12 identical; README H3 63/63/63/63
(all four shifted +1 together); skill-guide H3 counts each +1 (ko 5→6,
en/ja/zh 17→18 — the pre-existing ko↔derived structural divergence is
unchanged, so no NEW divergence for the AC-009 ratchet). URL blacklist grep
over all 8 touched files: 0 matches. Switcher header intact in all 4 READMEs.
No Mermaid added (TD-only trivially holds), no body emoji, no icon shortcode
needed (prose/table content), no version displays (plan.md §B.3 race
sidestepped). Hugo sanity build to /tmp: exit 0, and the new section
verified present in the rendered HTML of all 4 locales.

**Skill immutability (REQ-SGB-008 / AC-007) after M2:** `git diff --stat
origin/main -- .claude/skills/moai-domain-svg-infographic/
internal/template/templates/.claude/skills/moai-domain-svg-infographic/`
→ empty (re-checked after the docs edits; worktree fast-forwarded to
origin/main by the lead between M1 and M2, so the diff now measures clean
against the refreshed base).

**Gaps.** M3 (hns-oss-docs-verify full recipe + README parity checklist
record + PR readiness) not yet executed — the hugo run above was a syntax
sanity check, not the M3 gate. The emphasis cites repo paths under
`.moai/reports/t272/` which are untracked until the orchestrator commits
(AC-008's "cited paths exist" resolves at commit time on the branch).

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-08-25
run_commit_sha: pending-backfill-t272 (no commit made in run-phase — orchestrator reviews, then routes commit+PR through manager-git per the repo-local all-tier PR policy; backfill the SHA at commit time)
run_status: audit-ready (M1 + M2 + M3 complete; AC matrix below)

### AC matrix (from measured evidence only)

| AC | Status | Evidence (command → observed) |
|----|--------|-------------------------------|
| AC-001 coverage | PASS | `grep -c` of verdict-table data rows → `9`; form set equals spec.md §D.2; settled-dials column populated on every row |
| AC-002 G1 artifact | PASS | `ls -la .moai/reports/t272/artifacts/` → 9 SVGs, 2.5–5.4 KB each (non-empty) |
| AC-003 G2/G3 logs | PASS | 18 logs read; each carries command line + verbatim output + `exit=0`; render logs disclose Chrome 151.0.7922.174; lint logs read `0 errors, 0 warnings` (empty triage recorded as observed) |
| AC-004 taxonomy | PASS (vacuous) | No PARTIAL/NOT-PRODUCIBLE rows exist; acceptance.md §D.2 pre-authorizes the empty taxonomy |
| AC-005 deviation naming | PASS | 9 deviation sentences present in verdict.md §Deviation naming |
| AC-006 evidence location | PASS | Evidence tree entirely under `.moai/reports/t272/`; all verdict.md citations carry that prefix |
| AC-007 skill immutability | PASS | `git diff --stat origin/main -- .claude/skills/moai-domain-svg-infographic/ internal/template/templates/.claude/skills/moai-domain-svg-infographic/` → empty; measured at M1 checkpoint and re-measured post-M2 on HEAD 71781683c |
| AC-008 docs grounding | PASS | docs-site 9-row tables cite `.moai/reports/t272/artifacts/<form>.svg`; all 9 cited forms' verdict rows read PRODUCIBLE; cited paths exist in the working tree (resolve on-branch at commit) |
| AC-009 docs 4-locale parity | PASS | file-existence 150/150/150/150 identical (comm ×3 empty); section-count ratchet `comm -23 now base` → empty (0 NEW divergence); emphasis section verified in all 4 locales' rendered HTML |
| AC-010 README chain | PASS | `grep -c '^## '` → 12/12/12/12; switcher header in all 4; touched-section H3 +1 in each (63/63/63/63) |
| AC-011 verify recipe | PASS (flagged) | verify.md: §1 exit=0 warning-free + sitemap OK; §2/§3 0 matches; §4 both sub-checks clean; §5 added-lines emoji 0; §6 diff-attributed PASS with a pre-existing absolute divergence (origin/main Release badges v3.1.3 vs hugo.toml SSOT v3.1.2 — `git show origin/main` proves it pre-exists; this branch adds zero version displays). Decision routed to orchestrator |
| AC-012 env preflight blocker | PASS (vacuous) | Preflight passed (node v22.14.0, fixture render probe exit 0), so the blocker scenario never fired |

ac_pass_count: 12
ac_fail_count: 0

### Residual notes

- AC-011's version-sync flag is an attribution outcome, not a diff defect; the
  v3.1.3 release workstream owns the badge/SSOT reconciliation.
- AC-008's on-branch resolution and `run_commit_sha` backfill land when the
  orchestrator commits through manager-git.

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-08-25
sync_commit_sha: pending-backfill-t272 (single sync commit not yet made — manager-git backfills the SHA after the commit lands; that commit carries the draft → completed frontmatter transition and the 3-phase close)
sync_status: audit-ready

What sync verified (authored in worktree t272 at HEAD 71781683c, branch
WT-skillstead-gallery — same tree as the §E.2 AC-007 post-M2 measurement;
§E.4 cites the existing run-phase artifacts, it does not re-measure them):

- **M2 docs edits, 8 files, purely additive** — the producible-diagram-types
  emphasis in docs-site skill-guide ×4 locales (+18/-0 each) and README ×4
  locales (+4/-0 each); README H3 parity 63/63/63/63, H2 12/12/12/12,
  skill-guide section-count ratchet 0 NEW divergence (§E.2 M2 evidence).
- **M3 verify recipe — `.moai/reports/t272/verify.md` all-PASS** — §6
  version-display check is a diff-attributed PASS: the Release-badge
  divergence (origin/main badges v3.1.3 vs hugo.toml SSOT v3.1.2) pre-exists
  on origin/main, and the branch adds zero version displays. Gated per
  orchestrator decision (a): the v3.1.3 release workstream owns the
  badge/SSOT reconciliation.
- **AC matrix 12/12 PASS** — adopted verbatim from t272-run's §E.3
  self-check (`ac_pass_count: 12`, `ac_fail_count: 0`); sync adds no new AC
  surface.

`run_commit_sha` keeps the §E.3 placeholder (`pending-backfill-t272`); both
SHA backfills land with the sync commit via manager-git, which also commits
the `.moai/reports/t272/` evidence tree and the 8 M2 docs files.
