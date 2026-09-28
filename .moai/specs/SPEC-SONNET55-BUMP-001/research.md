# SPEC-SONNET55-BUMP-001 — Research Notes

## §1 Verified facts (announcement, https://www.anthropic.com/claude-sonnet-5-5, fetched 2026-09-29)

- Model id: `claude-sonnet-5-5`. Clear upgrade over Sonnet 5; 30%+ faster; up to 30% cheaper per task.
- Pricing per Mtok: input $2, output $10, cache read $0.20, cache write $2.50 — same input/output price as Sonnet 5.
- Effort: low→max supported. Defaults: Claude Code and apps = Medium; Claude Platform = High.
- Migration note: users running Sonnet with thinking off must switch to the `between_tools` thinking setting (up-front thinking stays off) before moving to Sonnet 5.5.
- Safety: first Sonnet with cyber safeguards (higher-risk cybersecurity tasks visibly fall back to Sonnet 5), biology safeguards as Sonnet 5, distillation safety classifiers, expanded preserved thinking.
- **Context window: NOT stated in the announcement.** See §2.

## §2 Context-window research procedure (gates plan.md M3, REQ-SSB-012, D-5)

1. Fetch the official models overview / pricing pages (docs.anthropic.com — models overview,
   pricing page). Record the context-window figure for `claude-sonnet-5-5` with the exact URL
   and fetch date, below.
2. Decision gate:
   - Figure stated → compare against the current moai tables (`context-window-management.md`
     "Sonnet 5 (1M)" row; docs-site tokenomics pages). Differ → update both, template-first.
     Same → record "no table change needed" with the citation.
   - Not stated → keep existing row values, record the citation gap here, and note it in the
     docs prose ("per Anthropic's docs as of <date>").
3. Never write 1M or 200K into any table without the §2.1 citation.

### §2.1 Research record (fill during M3)

- [ ] URL fetched: ___
- [ ] Context window stated: ___ (or "not stated")
- [ ] Fetch date: ___
- [ ] Decision: tables updated / unchanged — reason: ___

## §3 Tree measurements (a62a05764)

See spec.md §A.1 for the full evidence table. Headlines: alias table is the single canonical-id
home; GLM slot + web validation are alias-driven (no edit); `profile_matrix.go` absorbed into
`apply_harness.go` (t1246); launcher_test resolves dynamically (no edit); i18n.js picker labels
are the only hard-coded "Sonnet 5" user-facing strings.
