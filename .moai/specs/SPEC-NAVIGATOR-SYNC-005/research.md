# research.md — SPEC-NAVIGATOR-SYNC-005 (BAS Epic M3 — Falconer Fix)

> Plan-phase research artifact. Documents the CodeWiki `--compare-to` pattern investigation
> (from the design report + the cited external references) and maps it onto moai-adk's actual
> architecture, producing the split-architecture decision defended in `design.md §A`.
> This is a design-report-derived investigation, not fresh codebase research — the authoritative
> source is `.moai/reports/navigator-redesign-bas-20260805.html` + the references it cites.

## §A. The CodeWiki `--compare-to` Pattern (the technique M3 adopts)

### §A.1 What CodeWiki does

CodeWiki (FSoft-AI4Code, ACL 2026) is a standalone documentation-generation tool. Its relevant feature for M3 is the incremental-update model (design report §5 Q2 line 327):

> *"CodeWiki의 --update / --compare-to <commit> + metadata.json 증분 모델이 '갱신'의 검증된 기법 — Fix 요소가 이 패턴을 가져온다."*

The pattern:
1. **Baseline**: a known commit (the last time the docs were generated/updated).
2. **Diff**: `git diff --name-only <baseline>..HEAD` → the set of source files changed since the baseline.
3. **Scope**: identify which doc subtrees correspond to the changed source files (via the doc↔source binding the tool maintains).
4. **Regenerate**: re-run the LLM on ONLY the changed subtrees, producing updated doc sections.
5. **Metadata**: a `metadata.json` records what was regenerated, when, against which baseline — enabling the next run to diff against the new baseline.

The key property: **incremental, not regenerate-and-replace**. Only the changed subtrees are regenerated; unchanged sections are preserved as-is (including any human edits).

### §A.2 Why the pattern is "verified"

The design report §5 Q2 line 327 calls it *"검증된 기법"* (a verified technique). The verification grounds:
- CodeWikiBench (cited in design report §5 Q2 line 326): measures doc-generation quality across tools; CodeWiki outperforms pure-LLM approaches (DeepWiki) on system languages (C/C++ +3.15%), precisely because the deterministic structure layer (the `--compare-to` scoping) constrains the LLM's weaker spots.
- The `metadata.json` provenance model: each regeneration is attributable to a baseline + a scope, so the doc's history is reconstructable — the same provenance property moai-adk's `verification-claim-integrity.md` requires.

### §A.3 The CodeWiki `--compare-to` ↔ moai-adk M3 mapping

| CodeWiki step | moai-adk M3 equivalent | Lives in |
|---|---|---|
| Baseline commit | `nav-graph.json` `provenance.extract_commit_sha` (the default) or `--compare-to` flag | M0 artifact / CLI flag |
| Diff (`git diff --name-only`) | `navigator-fix` layer 1 diff-scope engine | `internal/navigator/fix/scope.go` |
| Scope (doc↔source binding) | M0 graph edges (`source_path` → bound subtrees) + M1 `changed_path` + M2 `owner_path` | M0/M1/M2 artifacts (consumed read-only) |
| Regenerate (LLM on changed subtrees) | orchestrator → `manager-develop` delegation (layer 2) | orchestrator (NOT Go engine) |
| Metadata.json | `request.json` + `applied.json` provenance blocks | staging surface |

The one structural difference: CodeWiki calls its LLM in-process; M3 delegates the LLM call to the orchestrator's `Agent()` runtime (design.md §A.2 — the two load-bearing reasons).

## §B. How moai-adk's Architecture Differs from CodeWiki

### §B.1 CodeWiki is standalone; moai-adk is a harness

CodeWiki ships as a standalone tool with its own LLM client — it owns the entire pipeline (diff → scope → LLM → output). moai-adk is a harness ON TOP of Claude Code (CLAUDE.md §1: *"MoAI is the strategic orchestrator for Claude Code"*). The `moai` binary does deterministic work; the LLM is reached through Claude Code's `Agent()` subagent runtime. moai-adk has no Go-embedded LLM client in ANY subsystem — not navigator, not hook, not CLI.

### §B.2 The implication for M3

The CodeWiki pattern CANNOT be copied verbatim. The "Regenerate (LLM on changed subtrees)" step must be split out of the Go engine and delegated to the orchestrator. This is not a limitation — it is the architectural invariant that keeps moai-adk's model-config, secret-management, and permission/effort surfaces in ONE place (Claude Code's runtime + `llm.yaml`), rather than duplicated into every Go subsystem that might want an LLM call.

The split produces three benefits:
1. **No model-config duplication** — the AI draft uses whatever model/effort/thinking the session is configured for, not a parallel config in the Go binary.
2. **No new secret surface** — API keys stay in Claude Code's env, not in the `moai` binary's env.
3. **Write-delegation preserved** — the orchestrator delegates the draft (a write) to `manager-develop`, per CLAUDE.md's core invariant.

## §C. Asset Reuse from M0/M1/M2 (design report §4)

### §C.1 The design-report asset-reuse table

Design report §4 (line 287-308) lists the assets BAS reuses. For M3 specifically:

| Asset | Location | M3 reuse |
|---|---|---|
| M0 `atomicWrite` pattern | `internal/navigator/sync/write.go` | apply-on-approval atomic-rename |
| M0 `Provenance` model | `internal/navigator/sync/types.go` | `request.json` + `applied.json` provenance (no wall-clock) |
| M0 graph edges | `nav-graph.json` | diff-scope subtree identification (the binding that maps changed paths to doc subtrees) |
| M1 detect `changed_path` | `.moai/state/navigator-detect/*.jsonl` | diff-scope seed (the real-time touched-path supplement to git-diff) |
| M2 work-item `action` | `work-items.json` | per-subtree fix-strategy hint (the "what" the AI draft should do) |
| M2 work-item `owner_path` | `work-items.json` | diff-scope fan-in (the "where") |
| Hidden-subcommand sibling pattern | `internal/cli/navigator_{sync,route,tiers}.go` | `navigator-fix` registration |
| `nonoverlap_test.go` pattern | `internal/navigator/{sync,detect,route}/` | `internal/navigator/fix/nonoverlap_test.go` |

### §C.2 What M3 does NOT reuse (and why)

- **No reuse of the 001/003/M0 regenerate-and-replace engines.** M3 is the incremental complement, not a caller of full regen. The three chains remain independent (separable changeability — design report §4 callout line 306-308).
- **No reuse of `navigator-audit.sh`'s matching logic.** M1 already cited it as inspiration for the path→graph-edge mapping (REQ-NS2-010); M3 consumes M1's output (the detect JSONL), not the shell script. M3's path→subtree mapping is via M0 graph edges (deterministic), not via the audit heuristic.
- **No reuse of M4 tiers.** M3 does not consume `tiers.json`; the tier overlays are M4's surface. M3's non-overlap with M4 (REQ-NS5-006) carries forward M4's REQ-NS3-016/017/018.

## §D. The Design-Report §9 Success Metrics for M3

### §D.1 The ≥50% Fix automation rate

Design report §9 line 443:

> | Fix 자동화율 | AI 초안이 1-click 승인으로 들어가는 비율 (수정 없이) | ≥ 50% (나머지는 가벼운 수정) |

**Definition**: of all drafts M3 produces, the fraction that the engineer approves via option (a) "approve + apply" WITHOUT a prior edit (option (c)) or subset-selection (option (b)). The "수정 없이" (without modification) qualifier is load-bearing — a draft that needed editing counts toward the denominator (total drafts) but not the numerator (unmodified approvals).

**Why ≥ 50% (not higher)**: the design report's parenthetical *"나머지는 가벼운 수정"* (the rest are light edits) sets the expectation. M3 is an AI-drafted tool; expecting 100% unmodified approval would be naive. The 50% floor is the threshold at which M3 is worth using (if < 50% of drafts are usable as-is, the engineer is editing more than approving — the gate becomes overhead, not a speedup). The measurement procedure (acceptance.md §D AC-NS5-010) is mechanically deterministic against a fixture corpus.

**How M3 hits ≥ 50%**: the diff-scope is tight (only stale subtrees, REQ-NS5-003), the per-subtree strategy hints come from M2's `action` field (the hint is already human-validated at the route layer), and the draft is scoped to a single doc section (not a whole artifact). Tight scope + good hints + small surface = high unmodified-approval rate.

### §D.2 The other §9 metrics M3 does NOT own

- **Detect 커버리지 ≥ 80%** — M1's metric (REQ-NS2-007). M3 consumes M1's output; M3 does not re-measure detect coverage.
- **Route 정확도 ≥ 70%** — M2's metric (REQ-NS4-010). M3 consumes M2's work-items; M3 does not re-measure route accuracy.
- **drift 감소율 ≥ 30%/sprint** — a portfolio metric across M1+M2+M3, not attributable to M3 alone. Measured at the sync-phase portfolio review, not in an M3 AC.
- **LLM 컨텍스트 정밀도 ≥ 40% 감소** — M4's blueprint-first metric (the "오라버니 통찰의 직적 측정", design report §9 callout line 449-451). M3 does not own this; M4 does.

M3 owns exactly ONE §9 metric: the ≥50% Fix automation rate.

## §E. The Falconer Living-Documentation Principle (the "why")

### §E.1 The core quote

Design report §2 line 207 (falconer, living-documentation guide):

> *"한 번 생성되고 다시 갱신되지 않는 문서는 살아있는 게 아니라 출처가 더 좋은 스냅샷이다."*
> ("A document that is generated once and never updated is not a living document — it is just a snapshot with a better source.")

### §E.2 How M3 closes the loop

M1 (Detect) makes drift VISIBLE. M2 (Route) makes drift ADDRESSABLE. M3 (Fix) makes drift RESOLVABLE. Without M3, M1+M2 surface drift but never close it — the doc map is still a snapshot, just one with visible drift annotations. M3 is the **갱신 인프라** (renewal infrastructure) that converts the Navigator from a snapshot into a living document.

The design report §8 risk grid (line 430-433, "다이어그램≠문서") names the hazard M3 exists to close: CodeWiki-style tools that position themselves as "generate once = done" are selling snapshots, not living docs. BAS's Fix element forces renewal infrastructure ON TOP of generation — the generation (001/003/M0) is necessary but NOT sufficient; M3 is what makes the doc map alive.

### §E.3 The safety boundary (why M3 asks before applying)

The falconer principle cuts both ways: a living document must be RENEWED, but a renewed-with-wrong-content document is worse than a stale one (it is a confidently-wrong living document). M3's approval gate (REQ-NS5-008) is the trust boundary — the human confirms the AI draft is correct before it lands. This is the one structural advantage M3 holds over the regenerate-and-replace chains (which apply without asking): M3's renewal is TRUSTED, because a human signed off.

## §F. References

### §F.1 Design report (the authoritative source)

- `.moai/reports/navigator-redesign-bas-20260805.html` — the BAS design report. §3.3(C) Falconer 3-element loop, §4 asset reuse, §5 Q2/Q3 (the three-questions answers), §6 M3 milestone card, §7 slice table, §8 risk grid, §9 success metrics.
- The report's footer (line 463-471) cites the external references it verified during investigation: CodeWiki (FSoft-AI4Code, ACL 2026), DeepWiki/DeepWiki-Open (Cognition), falconer living-documentation, Martin Fowler SDD three levels, Augment Code SDD guide, Kiro Design sections, GitHub Spec Kit, Tessl (`document --code`), SCIP (Sourcegraph), tree-sitter, LSIF, log4brains (ADR), Cucumber, Bazel/Buck.

### §F.2 moai-adk internal (the consumed artifacts)

- `internal/navigator/sync/` (M0) — `schema.go`, `scan.go`, `join.go`, `write.go` (`atomicWrite`), `mx_bridge.go`. The graph-join layer M3 consumes read-only.
- `internal/navigator/detect/traverse.go` (M1) — the reverse-traversal engine whose JSONL output M3 consumes.
- `internal/navigator/route/` (M2) — the promotion engine whose `work-items.json` output M3 consumes.
- `internal/navigator/tiers/` (M4) — the 4-tier overlay (M3 does not consume, but carries forward the non-overlap contract).
- `internal/cli/navigator_{sync,route,tiers}.go` — the Hidden-subcommand sibling pattern.

### §F.3 Rule cross-references (the architectural invariants)

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §2 — the attribution discipline for the ≥50% automation-rate claim.
- `.claude/rules/moai/core/askuser-protocol.md` § Channel Monopoly + § Preview Field Standards — the approval-gate mechanics.
- `.claude/rules/moai/workflow/nav-tokens.md` — the binding-token trio (`@NAV:DEC` / `@NAV:SYM` / `@MX:SPEC`) whose graph edges M3's diff-scope traverses.
- CLAUDE.md §1 (Core Identity) + §2 (Request Processing Pipeline) — the orchestrator-delegates-writes invariant that grounds the split-architecture decision.
- CLAUDE.local.md §2 [HARD] Template-First Rule + §25 Template Internal-Content Isolation — the distribution discipline.
