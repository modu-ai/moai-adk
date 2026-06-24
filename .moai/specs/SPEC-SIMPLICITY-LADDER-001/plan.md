# SPEC-SIMPLICITY-LADDER-001 — Implementation Plan

## §A. Context

Doctrine absorption of two `DietrichGebert/ponytail` (v4.7.0, MIT) mechanisms into MoAI: an ordered simplicity decision ladder (REQ-1, doc-only) and a `@MX:DEBT` deferred-simplification tag (REQ-2, doc + small Go change). The user reviewed the ponytail findings report and approved this scope as a bundle (two REQs, one SPEC).

## §B. The @MX:DEBT-vs-@MX:TODO Justification Verdict (REQ-2 challenge — owned here)

The spawn prompt required a genuine challenge: justify why `@MX:DEBT` is NOT redundant with `@MX:TODO`, and if the distinction does not hold, propose extending `@MX:TODO` instead rather than manufacturing a tag to be thorough.

### B.1 The two candidates, stated precisely

- `@MX:TODO` (existing): "incomplete work, resolved in the GREEN phase." Its lifecycle (per `mx-tag-protocol.md`): created in RED/ANALYZE, **removed** in GREEN/IMPROVE, escalates to WARN after >3 iterations unresolved. The semantic invariant: **the code is not yet done; the TODO is a promise to finish.**
- `@MX:DEBT` (proposed): "a deliberate, working simplification with a named ceiling + an upgrade trigger." The semantic invariant: **the code IS done and works correctly within its ceiling; the DEBT is a record of a knowingly-bounded choice, not a promise to finish.**

### B.2 Falsification check — can @MX:TODO absorb @MX:DEBT?

I tried to falsify the need for a new tag by collapsing DEBT into TODO. Three properties break the collapse:

1. **Resolution semantics differ.** A `@MX:TODO` is resolved when the work is *completed* (the missing thing is built). A `@MX:DEBT` is resolved when an *external trigger fires* (load grows, a dependency arrives, a scale threshold is crossed) — the work was never "incomplete". Marking a working, deliberate simplification as `@MX:TODO` lies about the code's state: it says "unfinished" when the truth is "finished, but bounded".
2. **The escalation rule actively misfires.** `@MX:TODO` escalates to `@MX:WARN` after >3 unresolved iterations — the protocol treats a long-lived TODO as a problem. A `@MX:DEBT` is *designed* to persist across many GREEN phases until its trigger fires; auto-escalating it to WARN would generate false danger signals for code that is working as intended.
3. **The required sub-lines differ.** `@MX:DEBT` carries `@MX:CEILING` (the validity limit) + `@MX:UPGRADE` (the exit trigger). `@MX:TODO` has neither and needs neither. The rot-risk detection (REQ-2.3 — a DEBT lacking `@MX:UPGRADE` is "no-trigger") has no analogue in the TODO lifecycle.

### B.3 VERDICT

**The distinction holds. A new `@MX:DEBT` tag is justified — NOT an extension of `@MX:TODO`.** The decisive property is #2: the existing TODO→WARN escalation rule would actively mis-handle deliberate-simplification debt, producing false danger signals. This is not "manufacturing a tag to be thorough"; reusing `@MX:TODO` would corrupt both tags' semantics and trip the escalation machinery. The honest, simplest correct design is a distinct kind with its own (non-escalating) lifecycle.

> Self-application note: even this verdict obeys REQ-1 rung 1 ("does this need to be built at all?"). The answer is yes *because* the cheaper option (extend TODO) was tried first and falsified — not assumed.

## §C. Pre-flight Verification (research complete — see spec.md §B)

All four required research items are resolved and recorded in `spec.md §B` with Read/Grep evidence:

1. ✅ REQ-1 insertion point located (constitution `### 4. Enforce Simplicity`, inside the evolvable block).
2. ✅ @MX house-format studied (`mx-tag-protocol.md` per-tag structure).
3. ✅ `internal/mx/` code-touch confirmed (`scanner.go:247` hard validity gate REJECTS unknown kinds → REQ-2 needs Go change).
4. ✅ karpathy-quickref "Simplicity First" confirmed LOC/abstraction-oriented → ladder is a complementary axis, not a duplicate.

## §D. Tier Decision — **Tier M**

**Decision: Tier M.** Rationale, stated honestly per the spawn prompt's two-branch test:

- The spawn prompt said: "propose Tier S if both deliverables are pure doctrine/doc edits; propose Tier M if research shows @MX:DEBT requires touching @MX validation code in `internal/`."
- Research (spec.md §B.3) shows `internal/mx/scanner.go:247` has a **hard validity gate** that returns `"unknown tag kind"` for any kind not in `{NOTE,WARN,ANCHOR,TODO,LEGACY}`. Therefore a `@MX:DEBT` tag cannot be scanned/harvested without registering `MXDebt` in Go code.
- Minimum REQ-2 code surface: `internal/mx/tag.go` (const), `internal/mx/scanner.go` (validity case + rot-risk no-trigger detection), `internal/mx/resolver_query.go` (queryable-kinds map), `internal/cli/mx_query.go` (CLI mapping) — each with TDD test coverage (≥85%).

So the second branch fires: **Tier M**. The Go change is small and localized (one new enum member threaded through 4 recognition sites + tests), but it is real runtime code, so Tier S would be a mis-classification.

> The deliverables stay in ONE SPEC (two REQs) per user approval. They do NOT split: REQ-1 (ladder doctrine) and REQ-2 (@MX:DEBT) are thematically one absorption of ponytail, and REQ-2's doctrine half lives in the same `mx-tag-protocol.md` neighborhood the ladder cross-references for its safety carve-out.

## §E. Self-Verification (plan-phase audit-ready signal)

- [ ] spec.md frontmatter: all 12 canonical fields present + `era: V3R6` + `tier: M`.
- [ ] SPEC ID `SPEC-SIMPLICITY-LADDER-001` passes the canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (decomposition printed in the completion report → PASS).
- [ ] §J Exclusions present with ≥1 `### Out of Scope — <topic>` H3 + `-` bullets (6 sub-headings).
- [ ] REQ-1 + REQ-2 in GEARS form (Ubiquitous / Where / When patterns used).
- [ ] @MX:TODO-vs-@MX:DEBT verdict recorded (§B above).
- [ ] No implementation code in spec.md (file names cited as touch-points only, no Go bodies).

## §F. Milestones (priority order — no time estimates)

| Milestone | Scope | Deliverable |
|-----------|-------|-------------|
| **M1** | REQ-1 ladder doctrine | Insert the 6-rung ladder + safety carve-out into `moai-constitution.md` § Agent Core Behaviors #4; add cross-ref line to `karpathy-quickref.md`. |
| **M2** | REQ-1 mirror | Mirror M1 edits to `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` + `.../development/karpathy-quickref.md`. Verify template neutrality (language-neutral ladder, no JS/Python tokens). |
| **M3** | REQ-2 doctrine | Add `@MX:DEBT` definition (purpose, When-to-Add/Update/Remove, lifecycle, `@MX:CEILING`/`@MX:UPGRADE` sub-lines, no-trigger rot-risk) to `mx-tag-protocol.md` + the @MX list in `moai-constitution.md § MX Tag Quality Gates`. |
| **M4** | REQ-2 code (RED→GREEN) | (a) Register `MXDebt` in `internal/mx/tag.go` + add the `RotRisk string \`json:"rotRisk,omitempty"\`` field to the `Tag` struct. (b) Add the recognized-sub-line-kind set `{CEILING, UPGRADE, REASON, SPEC, TEST, PRIORITY}` — **LEGACY EXCLUDED (D-NEW-1: LEGACY is a real tag kind, would be silently dropped)** — to `parseTag` (scanner.go:235-251), consulted BEFORE the `default` branch, so those sub-lines are skipped, not errored (AC-SL-009b). (c) Accept `MXDebt` in the `scanner.go` validity case + populate `RotRisk="no-trigger"` when a DEBT tag has no following `@MX:UPGRADE`. (d) Add `MXDebt` to `resolver_query.go` queryable map; map in `cli/mx_query.go`. TDD: failing test first per touch-point, INCLUDING the AC-SL-009c regression test (standalone `@MX:LEGACY` → 1 `MXLegacy` tag). CONSTRAINT: existing `TestScanFileWithWarnReason` MUST still pass. |
| **M5** | REQ-2 mirror | Mirror M3 doctrine edits to `internal/template/templates/.claude/rules/moai/workflow/mx-tag-protocol.md`, `.../core/moai-constitution.md`, AND `internal/template/templates/.claude/skills/moai/references/mx-tag.md` (CONFIRMED touched — DEBT joins its tag-type grammar at line 39, D-NEW-2). Each mirror verified by per-file `diff -q` (AC-SL-011a/b/c). |
| **M6** | Verify | `go test ./internal/mx/... ./internal/cli/...` (≥85% on changed packages); `go test ./internal/template/...` (neutrality CI guard); per-file `diff -q` mirror parity; spec-lint clean; full `go test ./...` for cascade. |

> Milestone ordering note: M1-M3 are doc edits (independent), M4 is the Go change (depends on M3's doctrine for the CLI `--kind DEBT` semantics), M5 mirrors M3, M6 verifies all. M4 is the only Tier-raising milestone.

> **D1/D2 scope boundary (run-phase MUST honor)**: M4's recognized-sub-line-kind set is the minimal fix that makes `@MX:CEILING`/`@MX:UPGRADE` scan cleanly (spec.md §B.5). Adding `REASON` to that set incidentally improves `@MX:REASON` scanning, but the general REASON-path repair + the `scanner_test.go:170` vacuous-test fix are OUT OF SCOPE (spec.md §J) — owned by the proposed `SPEC-MX-SUBLINE-PARSE-REPAIR-001`. M4 acceptance is verified on CEILING/UPGRADE only.

## §G. Anti-Patterns (self-application irony guard)

This is a SPEC about simplicity. It MUST NOT over-engineer itself. Forbidden in the run-phase:

- **AP-SL-001** — Adding a config knob for the ladder ("ladder enforcement level"). REJECTED in spec.md §J. The ladder is doctrine prose.
- **AP-SL-002** — Building a separate `debt-ledger.json` for `@MX:DEBT`. REJECTED — inline comment is the source of truth; grep / `moai mx query` is the harvest.
- **AP-SL-003** — A new lint rule or PreToolUse hook to mechanically enforce the ladder. REJECTED — out of scope.
- **AP-SL-004** — Restating the ladder in karpathy-quickref instead of cross-referencing it. Violates single-source-of-truth.
- **AP-SL-005** — Threading `MXDebt` through MORE than the recognition sites named in §F M4 (e.g. re-architecting the scanner). Touch only the minimum recognition surface.
- **AP-SL-006** — Adding `@MX:DEBT` to the per-file hard-limit machinery (`anchor_per_file`/`warn_per_file`) unless a test proves it is needed. DEBT has no per-file cap requirement in this SPEC.
- **AP-SL-007** (D2 scope guard) — Repairing the general `@MX:REASON` sub-line path or rewriting the `scanner_test.go:170` vacuous-test guard inside THIS SPEC. That is the proposed `SPEC-MX-SUBLINE-PARSE-REPAIR-001`'s job. M4 adds the recognized-sub-line set for CEILING/UPGRADE; any REASON improvement is incidental, never a deliverable here. Scope creep into the REASON repair is the exact over-engineering this simplicity SPEC must avoid.
- **AP-SL-008** (D3 contract guard) — Inventing a separate harvest surface (a new `moai mx harvest` command or a `debt-report.json`) for the rot-risk signal. The contract is ONE field (`rotRisk`) on the existing `Tag` struct, emitted by the existing `moai mx query --kind DEBT --json`. No new command, no new file.
- **AP-SL-009** (D-NEW-1 collision guard) — Including `LEGACY` in the recognized-sub-line-kind set. `LEGACY` is a real standalone tag kind (`MXLegacy`, tag.go:25); putting it in the sub-line set makes `parseTag` silently drop standalone `@MX:LEGACY` tags — a regression. The set is exactly `{CEILING, UPGRADE, REASON, SPEC, TEST, PRIORITY}`. Do NOT re-derive the set by copying `mx-tag-protocol.md:26`'s "Sub-lines:" list verbatim (that list dual-classifies `@MX:LEGACY`). The AC-SL-009c regression test is the canary.

## §H. Cross-References

- spec.md §B (research findings), §C (REQ-1/REQ-2 GEARS), §J (exclusions).
- acceptance.md (Given-When-Then + DoD).
- `internal/mx/tag.go:11-26` (TagKind const block + `RotRisk` field — M4 edit site).
- `internal/mx/scanner.go:80-110` (per-line loop + REASON handler) + `:235-251` (`parseTag` validity gate — M4 edit site for the recognized-sub-line-kind set).
- `internal/mx/scanner_test.go:170` (the vacuous `&& len(tags) > 0`-guarded REASON assertion — NOT touched by this SPEC; owned by `SPEC-MX-SUBLINE-PARSE-REPAIR-001`).
- `internal/mx/resolver_query.go:183-187` + `internal/cli/mx_query.go:125-133` (queryable map + CLI mapping — M4 edit sites).
- `internal/mx/tag.go:25` (`MXLegacy TagKind = "LEGACY"` — the D-NEW-1 collision source: LEGACY is a tag kind, so it MUST be excluded from the recognized-sub-line set).
- `.claude/rules/moai/workflow/mx-tag-protocol.md` (M3 doctrine target) — line 26 "Sub-lines:" list dual-classifies `@MX:LEGACY` (the D-CARRY doctrine ambiguity, left to `SPEC-MX-SUBLINE-PARSE-REPAIR-001`).
- `.claude/skills/moai/references/mx-tag.md:39` (tag-type grammar `tag_type := "NOTE" | "WARN" | "ANCHOR" | "TODO"` — M5 mirror target, DEBT added here per D-NEW-2; its template mirror at `internal/template/templates/.claude/skills/moai/references/mx-tag.md`).
- `SPEC-MX-SUBLINE-PARSE-REPAIR-001` (PROPOSED follow-up, not authored here) — owns the REASON-path repair surfaced in spec.md §B.5.
- CLAUDE.local.md §2 (Template-First mirror duty), §15 + §25 (template neutrality).
