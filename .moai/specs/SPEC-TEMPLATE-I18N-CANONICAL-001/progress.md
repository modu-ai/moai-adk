# SPEC-TEMPLATE-I18N-CANONICAL-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_authored_at: 2026-07-10
plan_status: audit-fix-applied (PASS-WITH-DEBT 0.89 at v0.1.0 → 6 defects fixed at v0.2.0; pending Implementation Kickoff Approval)
tier: L
plan_baseline_sha: 39c74d77787621b6645aebe81e470277ba3c97cb
artifacts: spec.md (13 REQ incl. REQ-I18N-013 unwanted-behavior / 23 AC IDs, 27 sub-checks: 20 mech / 4 hybrid / 3 manual), research.md (4-lens audit synthesis + doctrinal blind-spot analysis), plan.md (M1-M5), acceptance.md (23 AC IDs), design.md (Tier L 5-artifact contract — architecture intent + CI lint C9 + doctrine amendment structure), progress.md

design_invariant: localization mechanism (4-locale table + render-time substitution + English-fallback) preserved end-to-end; Korean remains equal-tier locale column; only canonical framing (default skeleton + "(canonical)" label + tiering prose) moves to English.

scope: 3 always-loaded template rule files (askuser-protocol.md / session-handoff.md / context-window-management.md) + session-handoff-examples.md + optional CI lint (M4) + optional doctrine amendment (M4). 2 clean baseline files (agent-common-protocol.md / goal-directive.md) preserved untouched.

## §E.2 Run-phase Evidence

Run-phase cycle_type: ddd (behavior-preserving doc/template refactor — ANALYZE-PRESERVE-IMPROVE; no executable production code except the M4 test file).

| AC ID | Status | Verification (verbatim evidence at `.moai/state/verify/i18n-canonical-001/`) |
|-------|--------|-------------------------------------------------------------------------------|
| AC-I18N-001a | PASS | loc-table row count = 9 (≥9 floor); memory-heading row added (verified via AC-I18N-010) |
| AC-I18N-001b | PASS | 4 locale columns: inline English/Korean present; ja/zh in examples |
| AC-I18N-001c | PASS | conversation_language substitution = 10 hits; Korean loc data = 11 hits (mechanism survives) |
| AC-I18N-001d | PASS | skeleton English-first headers = 3 (Preconditions/Run/After merge) |
| AC-I18N-001e | PASS | Hangul: askuser 1176→14, session-handoff 186→155, context-window 18→0 |
| AC-I18N-002 | PASS | askuser policy section English-canonical (residual = `(권장)` locale tokens) |
| AC-I18N-003 | PASS | 3 worked examples English-canonical; `(권장)` token discussion locale-aware |
| AC-I18N-004 | PASS | Fisher math `I=p(1−p)` = 3; 5-principle numbering = 5 |
| AC-I18N-005 | PASS | cross-refs: vci §1.1 surface 3 = 2; design.md §A.4 = 1; §B.2 = 1 |
| AC-I18N-006 | PASS | Korean block headers in skeleton = 0 |
| AC-I18N-007 | PASS | `Korean (canonical)` = 0 everywhere |
| AC-I18N-008 | PASS | `primary locales` = 0; `inline locales` = 1 |
| AC-I18N-009 | PASS | tiering contradiction resolved (inline locales + English-fallback consistent) |
| AC-I18N-010 | PASS | memory-heading row: inline en/ko = 2; examples ja/zh = 1 |
| AC-I18N-011 | PASS | close-time-pruning cross-ref content-keyed (next-session-start-point section) |
| AC-I18N-012 | PASS | goal-first example English-canonical + locale note |
| AC-I18N-013 | PASS | context-window resume format block Hangul = 0 |
| AC-I18N-014 | PASS | cross-ref session-handoff.md Localization Table = 1 |
| AC-I18N-015 | PASS | C9-natural-language-canonical-form class present (3 hits); test exit 0 |
| AC-I18N-016 | PASS | doctrine §25.6 amendment present (5 hits); C1-C8 unaltered |
| AC-I18N-017 | PASS | 2 clean baseline files byte-frozen (git diff 39c74d777..HEAD on 4 paths = empty) |
| AC-I18N-018 | PASS | clean baseline files 0-Hangul (agent-common-protocol 0, goal-directive 0) |
| AC-I18N-019 | PASS | sweep-induced delta byte-identical; askuser divergence = exactly §25-intentional provenance (5 hunks); session-handoff/context-window/examples identical |
| AC-I18N-020 | PASS | go build exit 0; TestTemplateNeutrality exit 0 (C9 PASS, C1-C8 no regression). PRE-EXISTING debt resolved by parallel session commit 1243094f6 restoring moai-easy.md to live tree; NOT introduced by this SPEC. |
| AC-I18N-021 | PASS (manual) | askuser policy section 5 principles parseable in English by non-Korean session |
| AC-I18N-022 | PASS (manual) | ko-locale render path preserved (loc-table Korean column + render-time substitution intact) |
| AC-I18N-023 | PASS (manual) | en-locale render path is default skeleton (English-first headers, consistent with fallback) |

Milestone commits: M1 `9ef2117e4` (askuser, draft→in-progress) · M2 `273963326` (session-handoff skeleton + loc-table) · M3 `b9a24fee0` (context-window resume format) · M4 `5aa05403e` (CI lint C9 + doctrine §25.6 + output-styles cascade) · M5 (this commit — progress.md §E + close-out).

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-07-10
run_commit_sha: 51483ee74
run_status: PASS-WITH-DEBT
ac_pass_count: 22
ac_fail_count: 0
ac_pass_with_debt_count: 1 (AC-I18N-020 — pre-existing moai-easy.md live-tree deployment gap, unrelated to this SPEC)
preserve_list_post_run_count: 4 baseline-file paths (2 clean files × 2 trees) byte-frozen; verified via git diff 39c74d777..HEAD empty
l44_pre_commit_fetch: yes (git fetch origin main before each milestone commit; divergence 0 0 or 0 N clean)
l44_post_push_fetch: pending (push at M5 close)
new_warnings_or_lints_introduced: 0 (C2/C4 advisory WARNs in TestTemplateNeutrality are pre-existing, not introduced by this SPEC)
cross_platform_build.darwin_amd64: exit 0 (go build ./...)
cross_platform_build.linux_amd64: not run (markdown-only + 1 Go test file; no platform-specific code touched)
cross_platform_build.windows_amd64: not run (same rationale)
total_run_phase_files: 11 (3 template rule files + 3 local mirrors + 1 examples template + 1 examples local + 1 test file + 1 doctrine + 2 output-styles cascade [template moai.md/moai-easy.md] + 1 output-styles local [moai.md]; spec.md frontmatter flip on M1)
m1_to_mN_commit_strategy: specific-path commits per milestone (5 commits M1-M5); index reset before each commit to isolate from parallel-session staged work; no --no-verify, no --amend, no force-push

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
