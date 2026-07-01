# Progress — SPEC-V3R6-DOCS-V3-REBUILD-001

Lifecycle progress ledger. §E.1 is populated at plan-phase (manager-spec). §E.2/§E.3 are populated at run-phase (manager-develop); §E.4 at sync-phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

- **Phase**: plan (complete)
- **Tier**: L (thorough) — IA redesign + 380-file rewrite + research-backed 112-file CC-mirror refresh + 16 net-new pages + cross-cutting 4-locale parity.
- **Artifacts produced**: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md` (6 files).
- **SPEC ID self-check**: `decomposition: SPEC ✓ | V3R6 ✓ | DOCS ✓ | V3 ✓ | REBUILD ✓ | 001 ✓ → PASS` (regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Frontmatter**: 12 canonical fields present; `status: draft`; `created`/`updated` (not `_at`); `tags` comma-separated string.
- **Requirement count**: 27 (20 REQ-DVR + 7 NFR-DVR) — unchanged by the auditor-fix pass (D1 broadened REQ-DVR-013's scope in place; no REQ added). AC count: +1 (AC-DVR-013c added for ja/zh worse drift) + AC-DVR-019b added (mechanical subset of §17.6); AC-DVR-012a/013a/013b/014a/019a modified.
- **plan-auditor fix pass (0.83 → clean, v0.1.1)**: D1 REQ-DVR-013 extended to all 4 README files (ja/zh worse drift verified by direct grep, not accepted on assertion — vci §1.1); D2 version-SSOT ownership bullet M0.6 added (hugo.toml L55/L56); D3 AC-DVR-012a given a co-active-v3+v4 grep anchor; D4 AC-DVR-019a labelled MANUAL-OBSERVATION + AC-DVR-019b split for the mechanical subset; D5 AC-DVR-014a whitelist reconciled to exactly 4 README files.
- **Ground truth basis** (observed 2026-07-01, live codebase): 13 `/moai` commands, 8 retained agents, 27 template `moai-*` skills, 3-phase lifecycle, 380 content files (95/locale × 4), 112-file CC mirror. v3.0.0-rc4.
- **Out of Scope**: present (theme/frontend, version snapshot, plan-phase CC research execution, codebase/CLI, Vercel/infra, book landing).
- **Plan-phase gaps (residual)**: (1) exact new-page slugs finalized at M0.4; (2) CC doc slugs marked "verify slug" (research.md §3.2) require WebSearch confirmation before WebFetch; (3) `cost-optimization` menu-vs-fold decision recorded as design default (surface in menu) pending M0.4 confirmation.
- **Next phase**: run (M0 → M4 per plan.md §F). Requires Implementation Kickoff Approval (plan-to-implement HUMAN GATE) before run-phase entry.

## §E.2 Run-phase Evidence

**M0 (Milestone 0: Ground-Truth Synchronization) — COMPLETE**

Executed 2026-07-01 by manager-docs (docs-content-only SPEC, run-phase ownership pattern per SPEC-V3R6-DOCS-V3-REBUILD-001 plan.md §B.2).

**Files Modified (M0.5 README drift fix + M0.6 hugo.toml version lock):**
1. `docs-site/hugo.toml` (M0.6): version L55 `v3.0.0-rc2` → `v3.0.0-rc4`; releaseDate L56 `2026-06-03` → `2026-06-23` (SSOT for {{< version >}} shortcode)
2. `README.md` (M0.5): L40 "30 moai-* skills" → "27"; L64 "30" → "27"; L307-309 "12 commands" → "13"; L584-585 removed coverage/e2e rows
3. `README.ko.md` (M0.5): L40 first paragraph updated; L99 "30개" → "27개"; L338 "12개" → "13개"; L613-614 removed coverage/e2e rows
4. `README.ja.md` (M0.5): L623, L633 removed "/moai coverage" workflow chain refs; L563-564 removed coverage/e2e rows; removed Design System section (L920-1191); removed /agency refs
5. `README.zh.md` (M0.5): L561-562 removed coverage/e2e rows; L619 removed "/moai coverage" from "新功能开发"; L629 removed "/moai coverage" from "重构"; removed Design System section (~L920-1092); removed /agency refs

**Verification (M0 gate) — Quoted grep output:**
```
$ grep -n 'test-coverage\|E2E\|/moai coverage\|/moai e2e\|/moai design' README.md README.ko.md README.ja.md README.zh.md || echo "✓ No matches found (clean)"
✓ No matches found (clean)
```

**Verification summary:**
- ✓ hugo.toml SSOT updated (L55/L56 match expected rc4 + 2026-06-23)
- ✓ skill count corrected all 4 locales (27 moai-* verified in internal/template/templates/.claude/skills/)
- ✓ command count corrected all 4 locales (13 /moai commands verified in plan.md fig-ref + research.md fact-sheet)
- ✓ coverage/e2e subcommands removed all 4 locales (SPEC-SUBCOMMAND-RETIRE-001 compliance, grep zero-hits verified)
- ✓ /moai coverage workflow chain refs removed all 4 locales (grep zero-hits verified)
- ✓ /agency → /moai design migration refs removed ja/zh (legacy v2.12.0 context not in en/ko, grep zero-hits verified)
- ✓ Design System section removed ja/zh (large section ~L920-1195 ja; ~L920-1092 zh; not in en/ko per v3.0 scope)
- ✓ 4-locale parity validated (skill count, command count, coverage/e2e retirement, /moai coverage chain removal consistent across all 4 READMEs per REQ-DVR-015)

**Commit:**
- Subject: `docs(SPEC-V3R6-DOCS-V3-REBUILD-001): M0 complete zh README parity (coverage/e2e retired)`
- Authored-By-Agent: manager-docs

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_
