# Implementation Plan — SPEC-STEERING-ALIGN-LOCAL-DIET-001

Tier S. Epic Steering-Align SPEC 5 of 5 (P6, FINAL). Lifecycle: plan → run → sync (3-phase). CONSERVATIVE diet bound (user-confirmed). Primary edited artifact `CLAUDE.local.md` is git-tracked — standard P2/P5 close mechanics (§F lifecycle).

---

## §A. Context

- **Work location**: project root `/Users/goos/MoAI/moai-adk-go/` (absolute path is dev-local; the maintainer-local `CLAUDE.local.md` lives at repo root).
- **SPEC artifacts**: `.moai/specs/SPEC-STEERING-ALIGN-LOCAL-DIET-001/{spec,plan,acceptance,progress}.md`.
- **Diet target**: `CLAUDE.local.md` (806 lines / 33939 bytes, 25 sections — re-verified live, spec.md §F.1). **git-tracked** (verified: `git ls-files` exit 0, committed history `8e78530bb` / `de13ecc4c` / `96fad88ff`, `git check-ignore` exit 1) — maintainer-local guide but a tracked file.
- **Existing infra (PRESERVE)**: all 25 sections except the two verified §2 candidates; `.moai/docs/template-internal-isolation-doctrine.md` §25.1/§25.3 (the M-POINTER targets — PRESERVE, do NOT edit); `.claude/rules/moai/development/coding-standards.md` (the file the stale ref wrongly cited — PRESERVE, do NOT edit).
- **Diet bound**: CONSERVATIVE (user-confirmed) — pointer-ize ONLY the two verified external-SSOT-duplication candidates; preserve ALL dev-local-unique knowledge; §19.1 HUMAN GATE body KEEP.

---

## §B. Known Issues (filtered to relevant categories — Tier S)

Per `manager-develop-prompt-template.md` § Applicability, Tier S MAY filter B1-B12 to relevant categories. Relevant here:

- **B6 (spec-lint heading)**: this SPEC's own `spec.md` uses `### Out of Scope — <topic>` h3 sub-headings under `## D. Out of Scope` to satisfy `OutOfScopeRule` (MissingExclusions). VERIFIED present (4 `### Out of Scope —` h3 sub-sections).
- **B4 (frontmatter schema)**: `created:`/`updated:`/`tags:` canonical (no snake_case alias); 12 required fields + `tier: S` + `era: V3R6`.
- **B8 (working tree hygiene)**: runtime-managed files (`.moai/state/`, `.moai/cache/`, `.moai/logs/`) untouched. `CLAUDE.local.md` is git-tracked — its edit lands in the run/sync commit diff alongside the SPEC artifacts (§F).
- **B10 (untouched paths PRESERVE)**: ONLY `CLAUDE.local.md` §2 (two edits) + the 4 SPEC artifacts change. No SSOT doctrine / Go / template edits. Parallel-session SPEC dirs (e.g. `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001`) untouched.
- **B9 (commit + push self-perform)**: Hybrid Trunk 1-person OSS — run-phase commits + pushes the SPEC artifacts. Conventional Commits + `🗿 MoAI` trailer. NEVER `--no-verify`.
- B1/B2/B3/B5/B7/B11/B12: N/A (no Go code, no syscall, no subagent-domain code, no CHANGELOG emission this phase).

---

## §C. KEEP / CUT / POINTER Classification (core deliverable)

The CONSERVATIVE bound admits exactly TWO content edits. Every other passage is KEEP. This table is the precise run-phase map (C-2 makes run-phase mechanical).

| # | Passage | Class | Mechanism | SSOT target (verified) | Verification command (run-phase MUST re-run, AC gate) |
|---|---------|-------|-----------|------------------------|--------------------------------------------------------|
| 1 | §2 Pre-PR Verification checklist (L108-118): the 7-item `- [ ]` bullet list | **POINTER** | M-POINTER — collapse the 7-item duplicated bullet list to the one-line pointer already at L118; MAY expand the pointer to name §25.1 as the catalogue. Keep the `**Pre-PR Verification ...**` lead sentence + the `template-neutrality-check.yaml` safety-net mention. | `.moai/docs/template-internal-isolation-doctrine.md §25.3` (5-item self-check) + §25.1 (content-class catalogue) | `grep -c '§25.3 Pre-commit Self-Check (5-item' .moai/docs/template-internal-isolation-doctrine.md` ≥1 AND `grep -c '§25.1 정의 — Allowed vs Forbidden' .moai/docs/template-internal-isolation-doctrine.md` ≥1 → else BLOCK + KEEP (REQ-LD-002) |
| 2 | §2.1 Template Content Neutrality cross-ref (L106): cites non-existent `coding-standards.md § MUST` + "C1/C2/C4/C5/C6/C8 per coding-standards.md MUST constraints" | **CORRECTION + POINTER** | Rewrite the cross-ref to point at `template-internal-isolation-doctrine.md §25.1 / §25` (the actual C1-C8 owner). Remove the two broken `coding-standards.md § MUST` citations. MAY compress prose; PRESERVE the neutrality obligation + `template-neutrality-check.yaml` CI-guard mention. | `.moai/docs/template-internal-isolation-doctrine.md §25.1` (actual C1-C8 owner) | (a) confirm STALE: `grep -cE 'C1/C2\|MUST constraints\|## MUST\|### MUST' .claude/rules/moai/development/coding-standards.md` == 0; (b) confirm NEW target exists: `grep -c '§25.1 정의 — Allowed vs Forbidden' .moai/docs/template-internal-isolation-doctrine.md` ≥1 → else BLOCK (REQ-LD-004) |
| 3 | §19.1 구현 착수 승인 body (L702-718) | **KEEP** (user decision) | minor header compression ONLY (e.g. trim "(renamed from GATE-2)" parenthetical OR the trailing 상위 SPEC 참조 footnote). `[HARD]` directive line + 4-step orchestrator obligation + violation anti-pattern PRESERVE. | n/a (KEEP) | `grep -c '\[HARD\].*구현 착수 승인.*plan-to-implement HUMAN GATE' CLAUDE.local.md` ≥1 (AC-LD-001) |
| 4 | §1, §3–§16, §17/§18/§21/§23/§24/§25, §19 Quick Pointer + Local Notes, §20, §22 (all dev-local-unique) | **KEEP** | none — out of CONSERVATIVE diet scope (§A.6 preserve map) | n/a (KEEP) | §1/§5/§22 heading anchors present (AC-LD-004) |

### §C.1 Derived target range (REQ-LD-007 — behavioral, NOT numeric-proxy)

The two qualifying edits remove approximately:
- Candidate #1: the 7-item bullet list (~7 lines) collapsed to a 1-2 line pointer → net ≈ −5 to −6 lines.
- Candidate #2: §2.1 prose compression of the corrected cross-ref → net ≈ −0 to −3 lines (correction may not shrink much; the broken citation is replaced, not deleted wholesale).
- Optional §19.1 minor header compression → net ≈ −0 to −2 lines.

**Derived range: 806 → ~771-781 lines (soft band).** This is GUIDANCE. Per REQ-LD-007 + the P5 over-cut lesson, the band is SOFT: if only the two candidates qualify and §19.1 compression is minimal, landing at ~781-800L is a legitimate behavioral-PASS — NOT an under-cut. The real target is "the two verified candidates corrected/pointer-ized + everything dev-local-unique preserved", NOT a line number. Over-cutting dev-local knowledge to reach ~771 is FORBIDDEN (REQ-LD-002/006 gate).

> **P5 lesson applied verbatim**: P5 OUTPUT-STYLE-SLIM estimated −150~250L but honestly landed at −26L because preservation forced fewer cuts, and that was the CORRECT outcome. This SPEC's CONSERVATIVE bound is even tighter — a small honest reduction is the expected, correct result. AC-LD-005 carries the behavioral-PASS escape so a small reduction does NOT fail the SPEC.

---

## §D. Constraints (DO NOT VIOLATE)

- PRESERVE: all 25 sections except the two §2 candidate edits; §19.1 body (KEEP, header-compress only); the SSOT doctrine files (`template-internal-isolation-doctrine.md`, `coding-standards.md` — do NOT edit, C-7); all dev-local-unique sections (§A.6 preserve map).
- FORBIDDEN: aggressive multi-section condense; §19.1 body pointer-ization/removal; neutralizing legitimate internal refs in `CLAUDE.local.md` (C-6); editing any file other than `CLAUDE.local.md` + the SPEC artifacts; `--no-verify`, `--amend`, force-push; introducing a NEW broken cross-ref (REQ-LD-004).
- REQUIRED: Conventional Commits + `🗿 MoAI` trailer; re-run the duplication/existence greps (REQ-LD-002/004) before each edit; verify the corrected §2.1 target exists on disk before committing.
- BINARY constraint: §19.1 `[HARD]` HUMAN GATE body MUST remain grep-present post-diet (AC-LD-001).

---

## §E. Self-Verification (run-phase deliverable preview)

Run-phase reports the AC Binary PASS/FAIL matrix (acceptance.md §D) with actually-observed command output per row. Key rows:

- AC-LD-001: `grep` §19.1 `[HARD]` HUMAN GATE body + 4-step obligation present.
- AC-LD-002: `grep` the §25.3 pointer survives AND the duplicated 7-item `[ ]` bullet list is gone.
- AC-LD-003: `grep` §2.1 no longer cites `coding-standards.md § MUST` / "C1/C2/...MUST constraints"; new ref to `template-internal-isolation-doctrine.md §25.1` present.
- AC-LD-004: dev-local section headings present (§1/§5/§22 + §3-§16 + §17/§18/§21/§23/§24/§25).
- AC-LD-005: `wc -l CLAUDE.local.md` in soft band (behavioral-PASS escape if minimal).
- AC-LD-006: new §2.1 target anchor exists on disk.
- AC-LD-007: SSOT doctrine files byte-unchanged (`git status` shows `CLAUDE.local.md` as a tracked modification (` M`) alongside the SPEC artifacts; doctrine files clean).

---

## §F. Milestones (priority-based, no time estimates) + Lifecycle close approach

### Milestones

- **M1 (priority High)**: Re-run the REQ-LD-002 / REQ-LD-004 verification greps live (the plan-phase grep snapshot could be stale). Confirm: §25.3 + §25.1 anchors exist; coding-standards.md §MUST genuinely absent. If any gate fails → blocker report, reclassify candidate as KEEP.
- **M2 (priority High)**: Apply candidate #1 (Pre-PR checklist → pointer). Edit `CLAUDE.local.md` §2: collapse the 7-item bullet list to the one-line §25.3 pointer (already at L118), naming §25.1 as the content-class catalogue.
- **M3 (priority High)**: Apply candidate #2 (§2.1 stale-cross-ref correction). Rewrite the §2.1 cross-ref to point at `template-internal-isolation-doctrine.md §25.1 / §25`; remove the two broken `coding-standards.md § MUST` citations; preserve the neutrality obligation + CI-guard mention.
- **M4 (priority Medium)**: §19.1 minor header compression ONLY (optional — trim "(renamed from GATE-2)" parenthetical or the 상위 SPEC 참조 footnote). The `[HARD]` body + 4-step obligation + anti-pattern MUST survive. If no clean minor compression is available, leave §19.1 verbatim (KEEP is the safe default).
- **M5 (priority High)**: Self-verify the full AC matrix (acceptance.md §D) with live command output. Confirm dev-local sections survive (AC-LD-004), §19.1 body survives (AC-LD-001), no new broken cross-ref (AC-LD-006), scope discipline holds (AC-LD-007).
- **M6 (priority High)**: Commit the SPEC artifacts + progress.md AND the `CLAUDE.local.md` diet edit in the same commit (Conventional Commits, `🗿 MoAI` trailer) — `CLAUDE.local.md` is git-tracked, so the edit is part of the commit diff (standard P2/P5 close mechanics). Push to main (Hybrid Trunk).

### §F.1 Lifecycle close approach — git-tracked CLAUDE.local.md (standard P2/P5 path)

`CLAUDE.local.md` is git-tracked (§A.5 / spec.md §F.1 — verified: `git ls-files CLAUDE.local.md` exit 0; committed history `8e78530bb` / `de13ecc4c` / `96fad88ff`; `git check-ignore` exit 1). The 3-phase close is the SIMPLE, standard case, identical to P2 (CLAUDE.md in-diff) and P5 (moai.md in-diff):

1. **What the commits carry**: the run-phase commit and the sync-phase close commit touch the SPEC artifacts (`.moai/specs/SPEC-STEERING-ALIGN-LOCAL-DIET-001/*.md`) + `progress.md` **AND** `CLAUDE.local.md` (the diet edit). `CLAUDE.local.md` is a tracked file, so its edit appears in the commit diff — exactly like P2's CLAUDE.md edit and P5's moai.md edit.
2. **`sync_commit_sha`**: the §E.4 `sync_commit_sha` is the SHA of the sync close commit whose diff DOES contain `CLAUDE.local.md` (plus the SPEC artifacts). This is the ordinary tracked-file pattern — the same as every prior Epic SPEC.
3. **AC verification reads the live file**: every AC (acceptance.md) verifies against the live on-disk `CLAUDE.local.md`. For a tracked file, the live working-tree state equals the committed state once the edit lands; `wc -l` / `grep` against the working-tree file is the authoritative evidence.
4. **Era classification**: `era: V3R6` is set explicitly in `spec.md` frontmatter (H-override per lifecycle-sync-gate.md) so auto-detection is bypassed; the progress.md §E.2/§E.4 markers + `sync_commit_sha` confirm the V3R6 3-phase layout for `moai spec audit` (drift 0 expected, grandfather-clause N/A — this is a modern-era SPEC).
5. **How prior Epic SPECs handled close**: P2/P5 closed with the SPEC artifacts in the commit diff AND their primary edited file (CLAUDE.md / moai.md) ALSO in the diff. Both of those files are git-tracked (`git ls-files CLAUDE.md` and `git ls-files .claude/output-styles/moai/moai.md` both return the path) — IDENTICAL tracked-ness to `CLAUDE.local.md`. There is NO tracked-ness difference; the lineage is the same standard path. The close-subject full-ID mandate (`chore(SPEC-STEERING-ALIGN-LOCAL-DIET-001): ... 3-phase close`) + the `3-phase close` infix are unchanged.

> **Run-phase note**: `CLAUDE.local.md` + the SPEC artifacts are all tracked and land in the same commit — the git-race surface is the standard tracked-file working tree. The pre-spawn fetch discipline (agent-common-protocol.md § Pre-Spawn Sync Check) applies before any spawn that commits.

---

## §G. Anti-Patterns (this SPEC's specific traps)

- **AP-LD-001**: treating "CONSERVATIVE" as license to also condense a verbose-looking dev-local section (e.g. §6 Testing Guidelines, §23 git workflow). FORBIDDEN — only the two verified §2 candidates qualify (REQ-LD-006). The §A.6 preserve map is the binding scope.
- **AP-LD-002**: pointer-izing §19.1 because "§19 is a pointer-index section". FORBIDDEN — §19.1 body is the user-confirmed KEEP (REQ-LD-005 / C-3). The §19 Quick Pointer TABLE is a pointer index; §19.1 is a [HARD] policy body and stays.
- **AP-LD-003**: numeric-proxy pass condition ("must remove ≥25 lines"). FORBIDDEN — P5 over-cut lesson. AC-LD-005 is a SOFT band with a behavioral-PASS escape (REQ-LD-007).
- **AP-LD-004**: "neutralizing" the legitimate `SPEC-V3R6-AGENT-TEAM-REBUILD-001` / `REQ-ATR-015` refs in §19.1 (or other internal refs across the file) because they look like template-neutrality violations. FORBIDDEN — `CLAUDE.local.md` is maintainer-LOCAL, the neutrality rules bind `internal/template/templates/**` not this file (C-6).
- **AP-LD-005**: editing the SSOT doctrine file (`template-internal-isolation-doctrine.md`) to "make the pointer match" instead of pointing AT it. FORBIDDEN — M-POINTER points at an existing SSOT; it never modifies the SSOT (C-7). If a genuine SSOT gap surfaces, return a blocker.
- **AP-LD-006**: forgetting to stage the `CLAUDE.local.md` edit into the run/sync commit. `CLAUDE.local.md` is git-tracked (§F.1) — the diet edit MUST land in the commit diff alongside the SPEC artifacts (standard P2/P5 close mechanics), and `sync_commit_sha` MUST point at a commit whose diff contains `CLAUDE.local.md`. Do NOT treat it as a local-only/untracked file. Verify `git status --porcelain CLAUDE.local.md` shows ` M CLAUDE.local.md` before commit, then confirm it is staged (`git add CLAUDE.local.md`).

---

## §H. Cross-References

- `spec.md` §A-H (this SPEC's requirements, constraints, evidence).
- `acceptance.md` §D (the AC matrix + REQ→AC traceability + severity).
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1 (C1-C8 catalogue) + §25.3 (5-item self-check) — the M-POINTER / correction targets.
- `.claude/rules/moai/development/coding-standards.md` — the file the §2.1 stale cross-ref wrongly cited (NO §MUST section; correction removes the broken citation).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — `(none) → draft` owned by manager-spec; close-subject full-ID mandate.
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` — era classification (era: V3R6 H-override); 3-phase close + `sync_commit_sha`.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S minimal delegation form.
- SPEC-STEERING-ALIGN-OUTPUT-STYLE-SLIM-001 (P5, COMPLETED, origin ed8172482) — the over-cut-avoidance lesson + behavioral-PASS escape pattern inherited.
- SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001 (P2, COMPLETED, origin d0ca1f214) — the M-DELETE/M-POINTER + over-cut-gate pattern mirrored.
