# progress.md — SPEC-V3R6-DOCS-COVERAGE-001

> Plan-phase skeleton. §E.1 populated with plan-phase audit-ready signal. §E.2–§E.5 are placeholder headings — populated by manager-develop (run) and manager-docs (sync/Mx) per the artifact ownership matrix.

---

## Plan-phase Status

- **SPEC ID**: SPEC-V3R6-DOCS-COVERAGE-001
- **Tier**: L (4-locale × multi-page-family + ja structural rewrite)
- **Artifacts authored**: spec.md, plan.md, acceptance.md, research.md (5-artifact set; design.md omitted — no architectural decisions)
- **Status**: draft (initial frontmatter, plan-phase)
- **Depends on**: SPEC-V3R6-DOCS-DOCSITE-001 (completed), SPEC-V3R6-DOCS-CODEMAPS-V3-001 (completed)

---

## §E.1 Plan-phase Audit-Ready Signal

- **SPEC ID regex decomposition**: `SPEC ✓ | V3R6 ✓ | DOCS ✓ | COVERAGE ✓ | 001 ✓ → PASS`
- **Frontmatter schema**: 12 canonical fields present; `era: V3R6` explicit (H-2 avoidance); `depends_on` set; snake_case aliases absent.
- **GEARS compliance**: REQ-001~008 use Ubiquitous / Event-detected / State-driven / Capability gate / Unwanted patterns. No legacy IF/THEN modality.
- **Exclusions**: §E lists 5 out-of-scope items (DOCSITE-001 6 axes / IA redesign / build config / user-owned harness skills / skill descriptions).
- **AC matrix**: 10 ACs (9 MUST + 1 SHOULD), per-locale digit-boundary-anchored grep verification.
- **Primary-source evidence**: research.md §1 carries verbatim `find`/`ls` output establishing canonical count = 32.
- **Coverage map**: research.md §3 enumerates 18 facts-bearing pages across 4 locales (en:5, ko:4, ja:4, zh:5).
- **spec-lint**: pending verification (see §E.2 note below — run-phase or plan-phase-final gate).

**Audit verdict**: plan-phase artifacts are audit-ready pending spec-lint confirmation.

---

## §E.2 Run-phase Evidence

Run-phase executed as `cycle_type=ddd` ADAPTED (docs reconciliation, no Go characterization tests; grep-based AC verification is primary). All 4 milestones (M1 en / M2 ko / M3 ja / M4 zh) landed in a single run-phase commit per REQ-006 4-locale parity.

### Files modified (docs-site/content/{en,ko,ja,zh}/)

- `advanced/builder-agents.md` × 4 locales — count 31→32
- `advanced/skill-guide.md` × 4 locales — count correction + Domain humanize add (8→9) + Meta/Harness slash normalization (zh) + ko/ja structural rewrite (9 fictional categories → 6 canonical, 37+37=74 fictional names eliminated) + en/zh in-body fictional-name cleanup (10+11=21 names eliminated)
- `getting-started/introduction.md` × 4 locales — count correction (L133/L156/L163)
- `getting-started/update.md` × 2 locales (en/zh) — statusline string count (L396)
- `core-concepts/what-is-moai-adk.md` × 4 locales — count correction (L7/L48/L267/L652 or L661)

### Primary-source evidence (canonical count)

```
$ find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d | wc -l
32
```

6-category breakdown: Foundation=4, Workflow=10, Domain=9 (incl. humanize), Reference=5, Meta/Harness=2, Design=1 → 31 specialized + 1 umbrella = 32.

### AC verification summary (verbatim grep output captured in completion report)

- AC-002 (32 present): en=12, ko=11, ja=11, zh=11 — all ≥1 ✓
- AC-003 (humanize): all 4 locales ≥1 ✓
- AC-004 (Domain=9): en=1, ko=1, zh=2 — all ≥1 ✓
- AC-005 (sub-count sum): 4+10+9+5+2+1 = 31 specialized ✓
- AC-006 (ja+ko structural rewrite): both 0 fictional, 0 nine-cat, 1 six-cat, 3 canonical headers ✓
- AC-007 (4-locale parity): all 4 locales 6 canonical headers ✓
- AC-009 (primary-source): `find ... | wc -l` → 32 ✓
- AC-010 (spec-lint): 0 errors, 1 warning (StatusGitConsistency — resolves on draft→in-progress transition)
- AC-011 (en/zh in-body fictional): en=0, zh=0 ✓
- AC-001 (31 residual): 1 per locale — each is the REQ-002-mandated "31 specialized" sub-count (correct), NOT stale "31 total" (see §E.3 PASS-WITH-DEBT note)
- AC-008 (locale-native idiom, SHOULD): all 4 locale-native idioms present ✓

### Go code changes

Zero. This is a docs-only SPEC; `internal/`, `pkg/`, `cmd/` untouched (verified via diff inspection).

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-18
run_commit_sha: <to-be-filled-post-commit>
run_status: PASS-WITH-DEBT
ac_pass_count: 10
ac_fail_count: 0
ac_pass_with_debt_count: 1
preserve_list_post_run_count: 0
l44_pre_commit_fetch: pending
l44_post_push_fetch: pending
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_test: N/A (docs-only SPEC, no Go changes)
  go_build: N/A
  golangci_lint: N/A
total_run_phase_files: 18
m1_to_mN_commit_strategy: single-run-phase-commit (REQ-006 4-locale parity)
```

### PASS-WITH-DEBT note — AC-001 / REQ-002 design overlap

AC-001 (digit-boundary grep for "31" + skill-kw = 0) and REQ-002 (express count as "1 umbrella + 31 specialized") are in inherent tension. Every locale's `advanced/skill-guide.md` intro line now reads "32 skills total = 1 umbrella + 31 specialized" — the digit "31" appears because REQ-002 mandates expressing the specialized sub-count. The mechanical AC-001 grep flags this 1 residual per locale, but the residual is the REQ-002-mandated correct sub-count, not a stale "31 total" claim. The stale "31 total" claims (12 in en, 10 in ko, 7 in ja, 11 in zh at baseline) were ALL eliminated. This is a known AC design overlap that the plan-auditor iter-2 (0.88 PASS) accepted; the debt is the mechanical grep's inability to distinguish "31 specialized" (correct) from "31 total" (stale). Functionally, the AC's intent (eliminate stale total claims) is fully met.

### Blocker report

None. No scope-expansion needed; no SPEC body modification required; no user decision pending.

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

---

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase>_
