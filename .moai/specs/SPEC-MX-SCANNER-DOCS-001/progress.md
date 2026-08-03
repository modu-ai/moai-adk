# progress.md — SPEC-MX-SCANNER-DOCS-001

## §E.1 Plan-phase Audit-Ready Signal

_status: plan-phase artifacts emitted (spec.md + plan.md + acceptance.md + progress.md). Ready for plan-auditor independent audit._

- SPEC ID regex check: PASS (`SPEC-MX-SCANNER-DOCS-001`).
- Frontmatter: all 12 canonical fields present; `status: draft`; `tier: M`.
- GEARS notation: 9 REQ entries (8 Ubiquitous + 1 Unwanted/`shall not` via "shall not be modified" in REQ-MSD-009).
- Out of Scope: 5 `### Out of Scope — <topic>` H3 sub-headings with bullets (satisfies `OutOfScopeRule`).
- Research basis: 4 features documented in plan.md §C with source-file citations.
- Tier rationale: documentation-only, 4 locales + 4 README files → Tier M (full artifact set incl. acceptance.md).

## §F Phase 4 Mode Selection

**Input parameters**:
- tier: M (documentation-only, 3-artifact set)
- scope: ~16-20 files (1 new docs-site page × 4 locales + cross-ref edits in 2 existing pages × 4 locales + 4 README files + menu/_meta wiring × 4 locales)
- domain count: 3 (docs-site content, README, navigation config)
- file language mix: 100% markdown + YAML config (no Go source — REQ-MSD-009 forbids scanner source changes)
- concurrency benefit: LOW for the canonical→derived dependency (derived locales require the canonical locale first); HIGH within the derived-locale fan-out (3 disjoint locale dirs)

**Mode evaluation**:
- Mode 1 trivial — NO (16-20 files, multi-milestone)
- Mode 2 background — NO (write work, not read-only async)
- Mode 3 agent-team — RETIRED (never selected)
- Mode 4 parallel — partial (applies to the derived-locale translation fan-out only, not the whole phase)
- Mode 5 sub-agent — SELECTED (milestone-sequential: canonical authoring → derived-locale fan-out → verify; the canonical→derived dependency makes the structure fundamentally sequential)
- Mode 6 workflow — NO (<30 files, not a single uniform mechanical transform — locale translation preserves facts but adapts per-locale emphasis/idiom)

**Decision**: sub-agent (Mode 5). The derived-locale milestone (M2) fans out 3 locale-translator specialists in one turn on disjoint locale directories (parallel-safe by disjoint paths; the harness locale-translator is designed for exactly this fan-out).

**Justification**: Per Anthropic's coding-task parallelism caveat, sequential is the safe default; this docs SPEC's canonical→derived locale dependency reinforces it (derived content cannot be authored before the canonical exists). The work is markdown translation, not a single uniform mechanical transform, so Mode 6 does not apply. Mode 5 with a bounded parallel burst for the 3 derived locales is the proportionate choice.

## §E.2 Run-phase Evidence

Run-phase deliverables (documentation-only; authored by oss-docs specialists, AC-verified by orchestrator):

**New 4-locale docs-site page** (`docs-site/content/<locale>/advanced/mx-scanner-internals.md`):
- `docs-site/content/ko/advanced/mx-scanner-internals.md`
- `docs-site/content/en/advanced/mx-scanner-internals.md`
- `docs-site/content/ja/advanced/mx-scanner-internals.md`
- `docs-site/content/zh/advanced/mx-scanner-internals.md`

**Cross-reference edits (8 files — 2 existing pages × 4 locales)**:
- `docs-site/content/{ko,en,ja,zh}/advanced/mx-tags.md`
- `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-mx.md`

**README FAQ entries (4 files)**:
- `README.md` (English canonical)
- `README.ko.md`, `README.ja.md`, `README.zh.md` (derived)

**Navigation wiring**:
- `docs-site/data/menu/main.yaml` (4-locale name maps + icon value)

**Verification (run by orchestrator; verbatim evidence recorded)**:
- Hugo `--minify --gc` build: exit 0, 0 warnings.
- Page-count parity ko/en/ja/zh = 134/134/134/134.
- H2 section parity on the new page = 5/5/5/5.
- Mermaid diagram byte-identical across 4 locales (md5 `60191824…`).
- 0 LR/RL Mermaid direction matches.
- 0 body-text emoji matches (typographic arrows / branding-emoji inside code blocks permitted).
- 0 forbidden/blacklisted URLs across all 4 locale pages + 4 README files.
- 8 cross-ref files touched (4 locales × {mx-tags.md, moai-mx.md}).
- `internal/` source tree untouched (AC-MSD-010) — `git diff --name-only origin/main...HEAD` filtered for `internal/` is empty.
- AC-MSD-009 FAQ-scope PASS; total-H3 count 29/30/30/29 is a PRE-EXISTING cost-axis drift outside this SPEC's scope (not introduced, not fixed).

**AC verification matrix (12/12 PASS)**:

| AC | Status | Evidence |
|----|--------|----------|
| AC-MSD-001 | PASS | rotRisk values documented (no-trigger / empty) |
| AC-MSD-002 | PASS | LSP fan-in + textual fallback + strict failure shape documented |
| AC-MSD-003 | PASS | CGO-gated complexity path + Supported:false triggers documented |
| AC-MSD-004 | PASS | Scan automation lifecycle points + fail-open ceilings documented |
| AC-MSD-005 | PASS | Mermaid diagram is `flowchart TD`; source identical across 4 locales (md5 60191824…) |
| AC-MSD-006 | PASS | `{{< icon >}}` shortcodes used; 0 body-text emoji |
| AC-MSD-007 | PASS | 4-locale same-PR parity; section structure preserved; diagram byte-identical |
| AC-MSD-008 | PASS | README 4-file FAQ entries with heading + section-order parity |
| AC-MSD-009 | PASS | FAQ scope confirmed (cost-axis total-H3 drift 29/30/30/29 is pre-existing, out-of-scope) |
| AC-MSD-010 | PASS | `internal/` untouched — `git diff --name-only origin/main...HEAD \| grep internal/` empty |
| AC-MSD-011 | PASS | All linked URLs on adk.mo.ai.kr; 0 blacklisted-domain matches |
| AC-MSD-012 | PASS | New page wired into main.yaml + reachable via 8 cross-ref links |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-08-04
ac_pass: 12
ac_total: 12
ac_summary: "AC-MSD-001..012 all PASS (4-locale docs-site page + 8 cross-refs + 4 README FAQ + nav wiring; internal/ untouched; 0 LR/RL, 0 body-emoji, 0 forbidden URLs, Mermaid byte-identical)"
hugo_build: "exit 0, 0 warnings (--minify --gc)"
page_count_parity: "ko/en/ja/zh = 134/134/134/134"
h2_parity: "5/5/5/5"
mermaid_md5: "60191824… (byte-identical across 4 locales)"
internal_touched: false
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-08-04
sync_commit_sha: "3c65146623235dbb29112c5b6c50f1db6f9687b9"
changelog_entry: "CHANGELOG.md [Unreleased] ### Added — SPEC-MX-SCANNER-DOCS-001"
frontmatter_status_transitions:
  spec_md: "draft → in-progress → implemented → completed (3-phase close on the single sync commit per spec-frontmatter-schema.md)"
  updated_field: "2026-08-04 refreshed on all 4 SPEC artifacts"
canary_compliance_check:
  b12_changelog_self_test: "PASS (pre-emission grep count 0; 12 distinct AC IDs in acceptance.md; file paths verified)"
  close_subject_full_id: "PASS (sync commit subject carries exactly one full SPEC-ID SPEC-MX-SCANNER-DOCS-001)"
  internal_untouched: "PASS (AC-MSD-010 preserved)"
```

> §E.4 `sync_commit_sha` carries a `pending-backfill-*` placeholder at this commit because a commit cannot reference its own SHA (D3 self-referential-hazard exemption per `spec-frontmatter-schema.md` § Forbidden ownership crossings). The real SHA is backfilled in a follow-up chore commit by the same phase-owning agent (manager-docs).
