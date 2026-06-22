# Plan — SPEC-STEERING-ALIGN-RULE-SCOPING-001

## §A. Context

Entry SPEC of Epic Steering-Align. Aligns moai-adk's always-loaded rule set to Anthropic's official "Steering Claude Code" `paths:`-scoping guidance and the best-practices per-line test. Scope is mechanical, frontmatter-only editing of 4 rules (3 Class-A path-scope additions + 1 Class-D always-load exclusion). 3 targets are MIRRORED (edited in BOTH template SSOT tree + live deployed tree) and 1 target is LIVE-ONLY (live tree only — see spec.md §A.1b). See spec.md §A for the full classification and §F for live-verified ground-truth.

## §B. Known Issues / Pre-existing State

- **D1 (iter-1 BLOCKING fix): the template tree is a SUBSET of the live tree.** LIVE has 61 total / 15 always-loaded; TEMPLATE has 59 total / 13 always-loaded. The 2 live-only files are `workflow/lifecycle-sync-gate.md` AND `workflow/runtime-recovery-doctrine.md` (both template-ABSENT). `lifecycle-sync-gate.md` is a Class-A edit target → it is LIVE-ONLY: live-file edit only, NO template edit, NO `make build` on its behalf, NO parity check. The other 3 targets (hook-independence, prompting-best-practices, NOTICE) ARE template-mirrored → template-first applies. Per-target presence verified in spec.md §F.1.
- `.claude/rules/` is embedded via `//go:embed all:templates` (`internal/template/embed.go:28`). The template tree is the SSOT for MIRRORED files; the live tree is the deployed copy. A `make build` is required after editing the template tree to re-embed — but ONLY because of the MIRRORED edits; the LIVE-ONLY edit does not require it.
- `NOTICE.md` body DIFFERS between trees (template 4639 B vs live 9580 B). NOTICE.md IS mirrored; the Class-D exclusion mechanism touches frontmatter/residency only and applies to both trees regardless of body divergence (the parity check compares only the `paths:` line, not the divergent body).
- None of the 4 edit targets currently has ANY frontmatter (verified: all always-loaded rules lack a `^paths:` line). The edit for each target is therefore a clean 3-line frontmatter-block insertion at the top of the file.

## §C. Pre-flight Checklist (run-phase entry)

- [ ] `git fetch origin main` + divergence check (multi-session race mitigation).
- [ ] Confirm the LIVE baseline still holds: `for f in $(find .claude/rules/moai -name '*.md'); do grep -q '^paths:' "$f" || echo "$f"; done | wc -l` → 15.
- [ ] Confirm the TEMPLATE baseline still holds: `for f in $(find internal/template/templates/.claude/rules -name '*.md'); do grep -q '^paths:' "$f" || echo "$f"; done | wc -l` → 13.
- [ ] Re-confirm per-target template presence (the MIRRORED-vs-LIVE-ONLY split has not drifted): `for r in development/hook-independence.md development/prompting-best-practices.md workflow/lifecycle-sync-gate.md NOTICE.md; do [ -f "internal/template/templates/.claude/rules/moai/$r" ] && echo "MIRRORED $r" || echo "LIVE-ONLY $r"; done` → 3 MIRRORED + 1 LIVE-ONLY (lifecycle-sync-gate).
- [ ] Capture the before byte-sum of BOTH always-loaded sets (LIVE 211495 B baseline, TEMPLATE 156308 B baseline; acceptance.md AC-SARS-006).

## §D. Constraints & Key Decisions

### D.1 Tier S justification

This SPEC is **Tier S (minimal)** (frontmatter `tier: S`, the dominant in-repo convention — 58 SPECs use `tier: S`) because:
- The change is **frontmatter-only** — a 3-line glob insertion per target + one exclusion mechanism for the Class-D rule. No body content, no Go code, no new lint rule, no test code.
- Low blast radius: 3 MIRRORED targets × 2 trees + 1 LIVE-ONLY target × 1 tree = **7 file edits**, all mechanical and reviewable as a tiny diff.
- The `paths:` mechanism is battle-tested (46 in-repo rules already use it per tree); this SPEC copies established precedent globs verbatim.
- Reversible: removing a `paths:` line restores always-load behavior instantly.

### D.2 Per-rule `paths:` glob decision (Class-A) — all `**/`-prefixed (D5)

The exact globs mirror the closest in-repo precedent (verified live). All carry the `**/` precedent prefix (D5 fix — iter-1 omitted it on the lifecycle-sync-gate glob):

| Rule | Class / Mirror | `paths:` glob to add | Precedent mirrored |
|------|----------------|----------------------|--------------------|
| `development/hook-independence.md` | A / MIRRORED | `paths: "**/.claude/hooks/**"` | `core/hooks-system.md` → `"**/.claude/hooks/**,..."` |
| `development/prompting-best-practices.md` | A / MIRRORED | `paths: "**/.claude/agents/**,**/.claude/skills/**"` | `agent-authoring.md` (`"**/.claude/agents/**"`) + `skill-authoring.md` (`"**/.claude/skills/**/SKILL.md"`) |
| `workflow/lifecycle-sync-gate.md` | A / **LIVE-ONLY** | `paths: "**/internal/spec/**,**/.moai/specs/**"` | enforcement target `internal/spec/era.go` + SPEC dir `.moai/specs/` |

> Note on the `**/` normalization (D5): the iter-1 `internal/spec/**,.moai/specs/**` lacked the `**/` prefix used by every precedent glob (`**/.claude/agents/**`, `**/*.go`). Normalized to `**/internal/spec/**,**/.moai/specs/**`. Post-normalization value-context re-confirmation: `**/internal/spec/**` matches `internal/spec/era.go` (the `ClassifyEra` enforcement surface) and `**/.moai/specs/**` matches the SPEC directory the lifecycle gate governs. The `**/` prefix broadens (anchor-anywhere), never narrows — both contexts remain covered.

> Note on `prompting-best-practices.md`: the rule serves agent-prompt AND skill-body authoring; the glob covers both `.claude/agents/**` and `.claude/skills/**`. The `skill-authoring.md` precedent narrows to `/SKILL.md`, but prompting guidance applies to any skill body file, so the broader `.claude/skills/**` is used deliberately (documented here so the run-phase agent does not "correct" it to `/SKILL.md`).

### D.3 Class-D exclusion mechanism decision (`NOTICE.md`)

**Decision: add a trivially-rare `paths:` scope rather than relocating the file.** Rationale:
- Relocating `NOTICE.md` out of `.claude/rules/` would change its on-disk location, risking breakage of any tooling that expects the third-party notice at the canonical rules path, and would be a larger (non-frontmatter) change — contrary to Tier S / REQ-SARS-007.
- A trivially-rare `paths:` (e.g. `paths: "**/NOTICE.md"` — matches only when a NOTICE.md file is itself opened, which is essentially never during normal development) removes always-load residency while keeping the file in place and keeping the change frontmatter-only.
- Run-phase agent applies `paths: "**/NOTICE.md"` to BOTH trees' `NOTICE.md`. The legal attribution stays on disk (REQ-SARS-006), satisfying the distribution requirement; it simply stops loading into context on every turn.

> If the run-phase agent finds a cleaner runtime-supported "exclude from always-load" mechanism (e.g. a config-level exclusion), it MAY return a blocker for re-delegation rather than silently choosing a different mechanism. Default: the trivially-rare `paths:` approach above.

### D.4 Honesty guardrail

Class-B and Class-C rules MUST NOT be scoped (REQ-SARS-010). The run-phase agent edits exactly 4 rules (3 MIRRORED + 1 LIVE-ONLY). Touching any of the other 11 live always-loaded rules is a scope violation.

## §E. Self-Verification (plan-phase audit-ready signal)

- [x] LIVE + TEMPLATE counts re-verified live (LIVE 61/15/46; TEMPLATE 59/13/46) — commands in spec.md §F.1.
- [x] D1 BLOCKING resolved: per-target template presence verified — 3 MIRRORED (hook-independence, prompting-best-practices, NOTICE) + 1 LIVE-ONLY (lifecycle-sync-gate, template ABSENT). The false "both trees carry 15" premise is corrected in spec.md §A.1.
- [x] Class-A trigger grounding: each rule's self-declared Loading-scope / opening line quoted in spec.md §A.2.
- [x] Class-D grounding: best-practices per-line test applied to NOTICE.md (answer: No → exclude).
- [x] `paths:` syntax precedent verified (go.md, agent-authoring.md, skill-authoring.md, hooks-system.md); all 4 target globs `**/`-prefixed (D5).
- [x] Embedding mechanism confirmed (`//go:embed all:templates`) → template-first + make build milestone is real for MIRRORED targets only.
- [x] D2 resolved: `tier: S` frontmatter added (58-SPEC convention).
- [x] SPEC ID self-check PASS (spec.md §G).
- [x] Per-tree delta target: LIVE 15 → 11 (−4: 3 Class-A live + 1 Class-D); TEMPLATE 13 → 10 (−3: 2 Class-A mirrored + 1 Class-D; lifecycle-sync-gate not in template).

## §F. Milestones (priority-ordered, no time estimates)

### M1 — Template-tree edits for MIRRORED targets (Priority High)
Edit the 3 MIRRORED targets in `internal/template/templates/.claude/rules/moai/...`: insert the `paths:` frontmatter block per the §D.2 table for `hook-independence.md` + `prompting-best-practices.md`, and `paths: "**/NOTICE.md"` for `NOTICE.md` (§D.3). Frontmatter-only. The LIVE-ONLY target (`lifecycle-sync-gate.md`) is NOT touched here — it has no template mirror.

### M2 — Re-embed (Priority High, depends on M1)
`make build` to regenerate the embedded template (`internal/template/embedded.go` via `go:embed all:templates`). Verify build succeeds. This re-embeds the 3 MIRRORED edits only.

### M3 — Live-tree edits + parity for MIRRORED targets (Priority High, depends on M2)
Apply the identical `paths:` frontmatter to the 3 MIRRORED targets' live copies under `.claude/rules/moai/...`. The live `paths:` values MUST be byte-identical to the template values (REQ-SARS-008 / AC-SARS-003). This is the explicit **template + local parity** milestone (MIRRORED targets only).

### M4 — Live-only edit for `lifecycle-sync-gate.md` (Priority High)
Apply `paths: "**/internal/spec/**,**/.moai/specs/**"` to the LIVE file `.claude/rules/moai/workflow/lifecycle-sync-gate.md` ONLY (REQ-SARS-008b). Do NOT create or edit a template file for it; do NOT run `make build` on its behalf; do NOT include it in the parity check. Independent of M1-M3 (no template dependency).

### M5 — Verification (Priority High, depends on M3+M4)
Run the acceptance.md AC verification batch: per-tree always-loaded count (LIVE 15 → 11, TEMPLATE 13 → 10); each Class-A rule still loads on matching file touch (glob sanity, `**/`-prefixed); MIRRORED parity diff empty with both-files-exist guard (D4); LIVE-ONLY assertion for lifecycle-sync-gate (live PRESENT + scoped, template ABSENT); no rule body content changed (`git diff` shows only frontmatter insertions); NOTICE.md retained on disk both trees; byte-sum reduction proven per tree.

## §G. Anti-Patterns to avoid

- AP-1: Editing the live tree first for a MIRRORED target (violates Template-First C-1).
- AP-2: Adding `paths:` to a Class-B or Class-C rule (violates REQ-SARS-010 honesty guardrail).
- AP-3: Editing rule body content while inserting frontmatter (violates REQ-SARS-007).
- AP-4: Template/live `paths:` values diverging on a MIRRORED target (violates REQ-SARS-008).
- AP-5: Deleting NOTICE.md instead of excluding it from always-load (violates REQ-SARS-006 — legal attribution must stay on disk).
- AP-6: Narrowing `prompting-best-practices.md` to `/SKILL.md` only (per §D.2 note, the broader `.claude/skills/**` is deliberate).
- AP-7 (D1): Fabricating a template edit / `make build` re-embed / parity check for the LIVE-ONLY target `lifecycle-sync-gate.md` (it has no template mirror — REQ-SARS-008b). Editing `internal/template/templates/.claude/rules/moai/workflow/lifecycle-sync-gate.md` would CREATE a stray template file that does not belong there.
- AP-8 (D5): Adding a glob without the `**/` precedent prefix (e.g. `internal/spec/**` instead of `**/internal/spec/**`).

## §H. Cross-References

- spec.md §A.2 (classification), §F (evidence), §G (ID self-check).
- acceptance.md (GEARS ACs + verification commands).
- CLAUDE.local.md §2 (Template-First), §15 (template neutrality — frontmatter additions are language-neutral).
- `internal/template/embed.go:28` (`go:embed all:templates`).
- In-repo `paths:` precedents: go.md, agent-authoring.md, skill-authoring.md, hooks-system.md.
