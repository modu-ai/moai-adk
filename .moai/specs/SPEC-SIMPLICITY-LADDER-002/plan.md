# SPEC-SIMPLICITY-LADDER-002 — Implementation Plan

## §A. Context

Tier S doctrine-only insertion. Repair the one rung MoAI dropped when SPEC-SIMPLICITY-LADDER-001 ported the `DietrichGebert/ponytail` (MIT) simplicity ladder: the in-codebase-reuse / DRY rung. Insert it at position 2 in the constitution ladder (LIVE) + its template mirror, renumber the five carried-over rungs, adjust one framing sentence, run `make build`, verify byte-parity + neutrality. No Go logic, no behavior change.

### A.1 PRESERVE list (DO NOT modify outside this scope)
- `moai-constitution.md` L238 intro line — keep verbatim.
- `moai-constitution.md` L249 safety carve-out ("Never simplify away …") — keep verbatim.
- `moai-constitution.md` L251 quantitative 3x-LOC trigger — keep verbatim.
- Everything in `moai-constitution.md` outside the ladder block (L238-251).
- `karpathy-quickref.md` — ONLY line ~33 (the cross-reference line) is touched (REQ-5: lead the arrow enumeration with in-codebase reuse + reuse-inclusive framing). Everything else in karpathy-quickref.md (the "Simplicity First" checkpoint questions, the rest of the file) is PRESERVED. The edit is surgical — one line, both LIVE and template mirror.
- All `internal/mx/` code, all `@MX` doctrine — NOT touched.
- Runtime-managed files (`.moai/state/*`, `.moai/cache/*`, `.moai/harness/*`), unrelated untracked files — NOT touched.

## §B. The "is the rung redundant?" challenge (owned here)

A genuine challenge: before inserting a 7th rung, prove it is NOT redundant with an existing rung (otherwise this is "manufacturing a rung to match ponytail's count").

### B.1 Falsification attempt — can an existing rung absorb in-codebase reuse?

| Candidate absorber | Why it does NOT cover in-codebase reuse |
|--------------------|------------------------------------------|
| Rung 1 (YAGNI) | YAGNI asks "should this exist AT ALL?" — a build/no-build decision. It does not ask "if it must exist, does it already exist HERE?". Different question. |
| Rung 3 (standard library, formerly rung 2) | Stdlib sources capability from the LANGUAGE RUNTIME, not the project. "Does Go's `strings` package do this?" is a different sourcing question from "Does a helper in `internal/` already do this?". |
| Rung 5 (installed dependency, formerly rung 4) | An installed dependency is an EXTERNAL package already in the dependency manifest. In-codebase reuse is FIRST-PARTY code in the project tree — no dependency, no import of a third party. Distinct source. |

None of the three absorb it. The reuse rung occupies a genuinely empty slot: **first-party, in-project capability reuse**, which is cheaper than the language stdlib (no new API surface to learn), cheaper than an installed dependency (no third-party coupling), and orthogonal to YAGNI (it presumes the thing must exist and asks where to source it).

### B.2 VERDICT

**The rung is NOT redundant. Inserting it at position 2 is justified.** Position 2 (after YAGNI, before stdlib) is correct because in-codebase reuse is the cheapest capability source of all — code that already lives in the project costs zero new code, zero new dependency, and zero new stdlib surface. This is not count-matching; it is repairing a real gap in the cheapest-capability-first ordering. MoAI already practices this informally ("grep the whole repo before reimplementing"); the rung codifies an existing discipline.

> Self-application note: this verdict obeys rung 1 ("does this rung need to exist?") — the answer is yes BECAUSE the cheaper option (fold reuse into an existing rung) was tried first and falsified, not assumed.

## §C. Pre-flight Verification (research complete — see spec.md §B)

All research items resolved with Read/Grep evidence in `spec.md §B`:

1. ✅ Edit-target lines located (LIVE + template both at L238-251, byte-identical — parallel `grep -n`).
2. ✅ Framing-prose distinction confirmed (L247 "dependency-avoidance ordering axis" needs one-clause reuse framing).
3. ✅ Mirror surface scoped (TWO file pairs — constitution LIVE+template AND karpathy-quickref LIVE+template; karpathy line 33 carries an inline ladder enumeration + framing, §B.3. No skill-reference / docs-site copy touched).
4. ✅ Template-neutrality precedent confirmed (existing ladder block carries no internal tokens; new rung is generic prose).

Run-phase pre-flight commands:
```bash
# 1. Confirm both files still byte-identical at the ladder block before editing
diff <(sed -n '238,251p' .claude/rules/moai/core/moai-constitution.md) \
     <(sed -n '238,251p' internal/template/templates/.claude/rules/moai/core/moai-constitution.md) \
  && echo "PRE-EDIT PARITY OK"

# 2. Confirm the ladder is currently 6 rungs (no "7." rung yet)
grep -c '^[1-6]\. ' .claude/rules/moai/core/moai-constitution.md   # informational
```

## §D. Tier Decision — **Tier S**

**Decision: Tier S.** Rationale:
- Pure doctrine/markdown insertion: one ladder block + two framing edits (constitution + karpathy) across TWO file pairs (constitution LIVE+template, karpathy LIVE+template). No Go code, no test, no behavior change.
- Scope: ~1 inserted rung + 5 renumbers + 1 constitution framing clause + 1 karpathy line ≈ within Tier S thresholds (< 300 LOC; 4 markdown files edited + catalog-hash regen via `make build`).
- Contrast with v001 (Tier M): v001 touched `internal/mx/` Go code (the scanner validity gate). This SPEC touches ZERO Go logic — `make build` runs `gen-catalog-hashes --all` (catalog-hash regen for the changed templates) + `go build`; there is NO `embedded.go` codegen and no code change.

Artifact set: Tier S nominally allows 2 files (spec.md + plan.md with AC inline). This SPEC ships the full 3-file set (spec.md + plan.md + acceptance.md) + progress.md skeleton for consistency with the v001 predecessor's structure and to keep the AC matrix in its own file. The Tier S minimal delegation form (~500-800 tokens) is appropriate for the run-phase prompt (per `manager-develop-prompt-template.md` § Applicability).

## §E. Self-Verification (plan-phase audit-ready signal)

- [ ] spec.md frontmatter: all 12 canonical fields present + `era: V3R6` + `tier: S`.
- [ ] SPEC ID `SPEC-SIMPLICITY-LADDER-002` passes the canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (decomposition printed in the completion report → PASS).
- [ ] §J Exclusions present with ≥1 `### Out of Scope — <topic>` H3 + `-` bullets (6 sub-headings).
- [ ] REQ-1..REQ-4 in GEARS form (Ubiquitous / Where / When patterns used).
- [ ] "Is the rung redundant?" challenge recorded with a falsification attempt + verdict (§B above).
- [ ] No implementation code in spec.md (only the verbatim ladder markdown + file/line cite as edit targets).
- [ ] PRESERVE list explicit (§A.1): safety carve-out + 3x-LOC trigger verbatim; intro line unchanged.

## §F. Milestones (priority order — no time estimates)

| Milestone | Scope | Deliverable |
|-----------|-------|-------------|
| **M1** | REQ-1 LIVE edit | In `.claude/rules/moai/core/moai-constitution.md`: insert the new rung 2 ("Does a helper, util, type, or pattern already exist in this codebase? Reuse it."), renumber the existing rungs 2-6 → 3-7. Preserve intro line, safety carve-out, 3x-LOC trigger verbatim (REQ-1.3). |
| **M2** | REQ-3 framing edit (LIVE) | Adjust the L247 framing sentence to the "reuse-and-dependency-avoidance ordering axis" form (REQ-3.1), keeping the language-neutral sentence intact. |
| **M3** | REQ-5 karpathy edit (LIVE) | In `.claude/rules/moai/development/karpathy-quickref.md` line ~33: lead the arrow enumeration with `in-codebase reuse` (`in-codebase reuse → stdlib → native platform feature → installed dependency → one line → minimum code`) AND change "dependency-avoidance decision ladder" → "reuse-and-dependency-avoidance decision ladder"; retain the "capability-source ordering axis …" tail. ONLY this one line. |
| **M4** | REQ-2/REQ-5 template mirrors | Apply M1+M2 to `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` AND M3 to `internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md`. Byte-parallel each. |
| **M5** | REQ-2.2 build | Run `make build` (`gen-catalog-hashes --all` regenerates the changed templates' catalog hashes, then `go build`). NO `embedded.go` exists — the directive `//go:embed all:templates` (`internal/template/embed.go:28`) embeds the live tree at compile time. |
| **M6** | REQ-4 verify | `diff -q` BOTH file pairs (constitution + karpathy) → clean (byte-parity, REQ-4.1). Template neutrality grep → 0 internal tokens in BOTH templates (REQ-4.2). 7-rung grep + per-rung content anchors (acceptance.md Scenario 1). PRESERVE grep (carve-out + trigger unchanged). karpathy enumeration-leads-with-reuse grep (Scenario 6). Gate: `go test ./internal/template/...` + `go build ./...`. |

> Milestone ordering note: M1-M3 are LIVE edits (constitution insert + framing, karpathy line), M4 mirrors all three to the templates, M5 runs `make build` (catalog-hash regen + compile; no `embedded.go` codegen), M6 verifies both file pairs. No milestone hand-edits a build artifact.

## §G. Anti-Patterns (self-application irony guard)

This is a SPEC about simplicity. It MUST NOT over-engineer itself. Forbidden in the run-phase:

- **AP-SL2-001** — Rewording the five carried-over rungs while renumbering them. They are renumbered ONLY; text unchanged (REQ-1.3).
- **AP-SL2-002** — Editing the safety carve-out (L249) or the 3x-LOC trigger (L251). PRESERVE verbatim.
- **AP-SL2-003** — Adding a ponytail citation, a SPEC ID, or any internal token to the constitution ladder block or its template mirror (neutrality violation — §B.4, REQ-4.2). The ponytail citation stays in the SPEC artifacts only.
- **AP-SL2-004** — Restating the FULL rung-by-rung ladder in karpathy-quickref, a skill reference, or docs-site. The REQ-5 karpathy edit is surgical (one cross-reference line: lead the arrow enumeration with in-codebase reuse + reuse-inclusive framing) — NOT a full ladder restatement. The rung-by-rung single source of truth stays in the constitution + its template mirror (§B.3).
- **AP-SL2-005** — Introducing a language-specific token in the new rung (e.g., "interface", "trait", "module" tied to one language; "helper, util, type, or pattern" is the language-neutral set). The new rung must read naturally for all 16 supported languages.
- **AP-SL2-006** — Skipping `make build` after the template edits, leaving the catalog hash (`catalog.yaml`) stale for the changed template files. `make build` runs `gen-catalog-hashes --all` then `go build`; the directive `//go:embed all:templates` embeds the live tree at compile time (no `embedded.go` to stale).
- **AP-SL2-007** — Hand-editing `catalog.yaml` template hashes (or inventing an `internal/template/embedded.go`) instead of running `gen-catalog-hashes --all` via `make build`. There is NO `embedded.go` — the embed is directive-based `//go:embed all:templates` (`embed.go:28`). Do not invent an embedded.go regeneration step.

## §H. Cross-References

- spec.md §B (research findings), §C (REQ-1..REQ-4 GEARS), §J (exclusions).
- acceptance.md (Given-When-Then + AC matrix + DoD).
- `.claude/rules/moai/core/moai-constitution.md` L238-251 (M1/M2 edit site).
- `.claude/rules/moai/development/karpathy-quickref.md` line ~33 (M3 edit site — cross-reference line).
- `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` L238-251 + `internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md` line ~33 (M4 mirror sites).
- `internal/template/embed.go:28` (`//go:embed all:templates` — directive embed; NO `embedded.go`. M5 `make build` runs `gen-catalog-hashes --all` + `go build`).
- `.moai/specs/SPEC-SIMPLICITY-LADDER-001/spec.md` §A (the v001 6-rung ladder this SPEC completes).
- CLAUDE.local.md §2 (Template-First), §15 (16-language neutrality), §25 (template internal-content isolation).
