# SPEC-SIMPLICITY-AUDIT-001 — Implementation Plan

## §A. Context

Doctrine/skill-prose absorption of `DietrichGebert/ponytail`'s (MIT) THIRD mechanism — a narrow, over-engineering-ONLY audit (`ponytail-review` diff-scope + `ponytail-audit` repo-scope) — into MoAI as a flag-gated `--lean` MODE of the existing `/moai review` skill. Sibling to `SPEC-SIMPLICITY-LADDER-001` (decision ladder + `@MX:DEBT` tag, completed origin 9d7f7b266). The user approved absorbing this third mechanism and explicitly deferred ponytail's benchmark harness (§J).

This is a Tier S, zero-Go, skill-prose + template-mirror SPEC. The only files touched in the run-phase are `review.md` (live + template mirror).

## §B. Application-Point Verdict (the central plan-phase decision — owned here)

The spawn prompt required a genuine least-invasive-integration decision across four candidate homes. The verdict, each branch backed by file evidence (spec.md §B):

| Candidate home | Verdict | Reason |
|----------------|---------|--------|
| `--lean` MODE of `review.md` | **CHOSEN** | `review.md` already uses flag-gated lenses (`--security`/`--design`/`--critique`/`--staged`/`--branch`/`--file`). `--lean` reuses Phase 1 (change identification) + Phase 5 (next steps); adds only the scan body + report format. Lightest integration, zero new skill surface. |
| Standalone `lean-audit.md` skill | REJECTED | Duplicates `review.md`'s change-identification + report machinery — the exact over-engineering the SPEC is about (irony guard). |
| sync-auditor Craft-dimension sub-check | REJECTED | sync-auditor is a PASS/FAIL **gate** with a must-pass firewall; the lean audit is **finding-only, verdict-less, advisory**. Conflating a no-gate lens into the 4-dimension verdict scoring corrupts sync-auditor's semantics AND risks a false FAIL on working, correct code (over-engineering is a smell, not a defect). |
| New `anti-patterns.md` entries | REJECTED | The 5 lean tags map onto existing Karpathy Categories 1+2 (Behavior 4). Cross-reference (REQ-3.1), don't duplicate (single source of truth). |

### B.1 The `@MX:DEBT` connect-vs-independent decision (sub-verdict)

**Verdict: CONNECT one-directionally, NOT stay fully independent.** A `yagni:` finding (single-implementation abstraction) is exactly the deliberate-simplification case `@MX:DEBT` records. The cheaper option — stay fully independent — was tried first and falsified: an independent lean audit would re-surface a site already carrying an `@MX:DEBT` marker as a *fresh* `yagni:` finding, generating noise the sibling SPEC's `@MX:DEBT` mechanism was built to suppress. So the lean audit consults the existing `moai mx query --kind DEBT` harvest (a read, no new code) and annotates already-tracked sites as `[already tracked @MX:DEBT — deferred]`. The link is one-directional (read-only); the lean audit never WRITES `@MX:DEBT` (that would let an advisory lens mutate source — out of scope, §J).

> Self-application note: even this connect verdict obeys the simplicity discipline — the connection reuses a surface the sibling ALREADY shipped (`moai mx query --kind DEBT`), adding zero new code. The cheaper "fully independent" option was tried first and falsified, not assumed.

## §C. Pre-flight Verification (research complete — see spec.md §B)

All four required research items are resolved and recorded in `spec.md §B` with Read/Grep evidence:

1. ✅ Application point — `review.md` Read confirms flag-gated-lens pattern (`--security`/`--design`/`--critique`); `--lean` fits as a new flag-mode (lightest integration).
2. ✅ sync-auditor Read confirms 4-dimension PASS/FAIL gate, NO over-engineering dimension → lean audit stays SEPARATE (verdict-semantics protection).
3. ✅ `@MX:DEBT` harvest surface confirmed landed (mx-tag-protocol.md:57,65 — `moai mx query --kind DEBT --json`, `rotRisk: no-trigger`) → REQ-2 connects via a read, no new code.
4. ✅ `anti-patterns.md` Read confirms Karpathy Cat 1+2 already cover the over-engineering semantic space → REQ-3 cross-references, does not duplicate.

## §D. Tier Decision — **Tier S**

**Decision: Tier S.** Rationale per the spawn prompt's two-branch test (S if prose/skill-heavy minimal-to-zero Go; M if research shows a runtime-code touch):

- Research (spec.md §B) shows the mechanism lands ENTIRELY as skill prose: a `--lean` flag-mode section in `review.md` + its template mirror + cross-reference lines. There is NO `internal/mx/`-style validity-gate to register (contrast the sibling `SPEC-SIMPLICITY-LADDER-001`, which was Tier M precisely because `@MX:DEBT` required touching `internal/mx/` Go). The `@MX:DEBT` connection REUSES the sibling's already-shipped `moai mx query` surface — no new Go.
- File count: 2 files (`review.md` live + template mirror). LOC-equivalent: well under 300 (a flag entry + one mode section). Well within Tier S guidance (< 5 files, < 300 LOC).
- Artifact set per Tier S: **spec.md + plan.md (AC inline spec.md §3) + progress.md**. No acceptance.md, no design.md, no research.md.

So the first branch fires: **Tier S**. The plan-auditor PASS threshold for Tier S is 0.75.

## §E. Self-Verification (plan-phase audit-ready signal)

- [x] spec.md frontmatter: all 12 canonical fields present + `era: V3R6` + `tier: S`.
- [x] SPEC ID `SPEC-SIMPLICITY-AUDIT-001` passes the canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (decomposition printed in the completion report → PASS).
- [x] §J Exclusions present with ≥1 `### Out of Scope — <topic>` H3 + `-` bullets (7 sub-headings, incl. the mandated LLM-as-judge benchmark exclusion).
- [x] REQ-1..REQ-4 in GEARS form (Where / Ubiquitous / When patterns used; no legacy IF/THEN).
- [x] AC inline in spec.md §3 (Tier S), 8 ACs, each binary + independently verifiable with a grep/diff command.
- [x] Application-point verdict recorded (§B above) with the standalone-skill / sync-auditor / anti-patterns rejections.
- [x] `@MX:DEBT` connect-vs-independent sub-verdict recorded (§B.1).
- [x] No implementation code in spec.md (file names + flag literals cited as touch-points only; no skill-body prose pre-written).
- [x] Irony guard applied (§G) — zero new skill/agent/config/lint/hook/Go; mechanism is one flag-mode.

## §F. Milestones (priority order — no time estimates)

| Milestone | Scope | Deliverable |
|-----------|-------|-------------|
| **M1** | REQ-1 `--lean` mode | Add the `--lean` flag to `review.md` Supported Flags + a new "`--lean` Mode — Over-Engineering-Only Lean Audit" section: short-circuit-4-perspective statement, diff/repo scope (REQ-1.2), 5-tag table + output format (REQ-1.3), language-neutral tag phrasing (REQ-1.4), `net:`/`Lean already. Ship.` summary (REQ-1.5), read-only/advisory/no-verdict + Phase 5 routing (REQ-1.6). |
| **M2** | REQ-2 `@MX:DEBT` cross-link | In the `--lean` mode section, document the one-directional `yagni:`→`moai mx query --kind DEBT` annotation (`[already tracked @MX:DEBT — deferred]`) + the never-writes-markers boundary (REQ-2.1/REQ-2.2). |
| **M3** | REQ-3 cross-refs | Add the two cross-reference lines (`anti-patterns.md` Cat 1+2 / Behavior 4; `moai-constitution.md` #4 ladder — lean audit = post-hoc detection counterpart). Reuse, do not restate (REQ-3.1/REQ-3.2). |
| **M4** | REQ-4 mirror | Mirror the M1-M3 edits to `internal/template/templates/.claude/skills/moai/workflows/review.md`. Verify template neutrality: NO `SPEC-SIMPLICITY-AUDIT-001` / `REQ-SA` / dev-date leak (CLAUDE.local.md §25). The mirrored `--lean` content is generic ponytail-mechanism prose only. |
| **M5** | Verify | Run AC-SA-001..008 grep/diff checks; confirm `--lean` section present in BOTH live + template; confirm no internal-content leak in template; confirm 5 tags + both summary forms + cross-refs present. No Go test (zero Go change), but run `go test ./internal/template/...` neutrality CI guard to confirm the mirror passes. |

> Milestone ordering note: M1-M3 are sequential edits to the SAME `review.md` section (M1 creates the section, M2/M3 extend it). M4 mirrors the finished section. M5 verifies. There is no parallelism — all edits target one section of one file (+ its mirror). This is the minimal milestone set for a Tier S skill-prose SPEC.

## §G. Anti-Patterns (self-application irony guard)

This is a SPEC about over-engineering. It MUST NOT over-engineer itself. Forbidden in the run-phase:

- **AP-SA-001** — Creating a standalone `lean-audit.md` skill instead of a `--lean` mode of `review.md`. REJECTED in spec.md §B.1 / §J. The mechanism is a flag-mode, not a new skill.
- **AP-SA-002** — Adding an over-engineering dimension to sync-auditor's 4-dimension scoring. REJECTED — sync-auditor is a PASS/FAIL gate; the lean audit is verdict-less advisory. Conflation corrupts verdict semantics + risks false FAIL.
- **AP-SA-003** — Adding a new lint rule, hook, config file, or Go code to mechanically enforce leanness. REJECTED — out of scope (§J). The lean audit is skill-prose advisory.
- **AP-SA-004** — Restating the Karpathy 8-category catalogue or the 6-rung ladder in the `--lean` section instead of cross-referencing them. Violates single-source-of-truth (REQ-3).
- **AP-SA-005** — Making the `--lean` mode write/create `@MX:DEBT` markers from `yagni:` findings (bidirectional link). REJECTED — the link is one-directional read-only (REQ-2.2); an advisory lens must not mutate source.
- **AP-SA-006** — Porting ponytail's LLM-as-judge 0-3 benchmark harness. REJECTED — the mandated §J exclusion. Absorb the mechanism, not the measurement.
- **AP-SA-007** — Letting the `--lean` mode report correctness / security / performance findings (scope creep into the comprehensive review's domain). REJECTED — REQ-1.1's hard scope boundary IS the mechanism's value; diluting it reproduces the comprehensive review.
- **AP-SA-008** — Leaking moai-adk-internal tokens (SPEC ID, REQ codes, dev dates) into the template mirror. REJECTED — CLAUDE.local.md §25 template internal-content isolation; the mirrored `--lean` content is generic ponytail-mechanism prose.

## §H. Cross-References

- spec.md §B (research findings), §C (REQ-1..REQ-4 GEARS), §3 (inline AC), §J (exclusions).
- `.claude/skills/moai/workflows/review.md` (M1-M3 edit target) + `internal/template/templates/.claude/skills/moai/workflows/review.md` (M4 mirror target).
- `.claude/skills/moai/references/anti-patterns.md` (REQ-3.1 cross-ref — Karpathy Cat 1+2).
- `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors #4 (REQ-3.2 cross-ref — the ladder).
- `.claude/rules/moai/workflow/mx-tag-protocol.md:57,65` (REQ-2 `@MX:DEBT` harvest surface — read-only consume).
- `.claude/agents/moai/sync-auditor.md` (§B rejected-home evidence).
- `SPEC-SIMPLICITY-LADDER-001` (sibling — supplies the `@MX:DEBT` harvest this SPEC reads).
- `DietrichGebert/ponytail` (MIT) — mechanism source (benchmark excluded, §J).
- CLAUDE.local.md §2 (Template-First), §15 + §25 (neutrality + internal-content isolation).
- progress.md (§E.1 plan-phase audit-ready signal + §E.2-§E.4 placeholder headings).
