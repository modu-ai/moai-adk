# Plan — SPEC-PROJECT-NAVIGATOR-002

> Plan-phase artifact. Order: highest-change-likelihood decisions first (algorithm, format, boundary) → mechanical/refactor steps last (template mirror, CI guards). The user approves structure + mechanism at the Implementation Kickoff Approval gate.

## §A. Context

### §A.1 Branch + state

- Main checkout, branch `main`, HEAD `fb6072a22`, synced with `origin/main` (verified `0 0`, no concurrent sessions). This is a 1-person OSS repo with PR-mandatory policy (`enforce_admins:true`); plan-phase artifacts are markdown only and land via the standard plan→PR flow.
- SPEC dir: `.moai/specs/SPEC-PROJECT-NAVIGATOR-002/` (expanding the existing stub in-place; no new SPEC-ID).
- Today's date: 2026-08-05.

### §A.2 PRESERVE list (this SPEC writes NOTHING here)

- `.claude/rules/moai/**`, `CLAUDE.md`, `CLAUDE.local.md` — FROZEN doctrine.
- `.claude/agents/moai/**` — FROZEN retained agents.
- `.moai/project/{product,structure,tech}.md` — design docs are audit INPUTS (REQ-NA-002 read-only); never written.
- `.moai/project/navigator/{navigator,capability-map,progress-map}.md` — 001's outputs; audit is read-only over them (REQ-NA-002).
- `.moai/specs/SPEC-PROJECT-NAVIGATOR-001/` — completed; do not modify.
- `.moai/specs/SPEC-PROJECT-NAVIGATOR-003/` — sibling stub; do not modify.
- `.moai/specs/SPEC-LSEL-LOCAL-EVOLUTION-001/` and all other SPEC directories.
- `.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*` — LSEL surfaces; REQ-NA-011 forbids touching them.
- `.moai/specs/SPEC-*/spec.md` — the audit reads frontmatter only; NEVER writes SPECs.

### §A.3 EXTEND targets (the only surfaces this SPEC adds to)

User-project surfaces (generated, not scaffolded — these are the audit's *output*, created at runtime by the skill mode):

- NEW (runtime-generated): `.moai/project/navigator/audit-report.md` — human-readable drift report.
- NEW (runtime-generated): `.moai/project/navigator/audit-report.json` — machine-readable drift report.
- NEW (optional, user-authored): `.moai/project/navigator/audit-known-matches.yaml` — manual override file (the user writes this when they want to silence deliberate naming divergence).
- NEW (runtime log): warnings continue into `.moai/logs/navigator-warnings.log` (shared with 001).

Template-distributed surfaces (the *mechanism*, shipped to all users via `internal/template/templates/`):

- EDIT: `internal/template/templates/.claude/skills/moai-workflow-project/SKILL.md` — extend with a `--audit` skill mode (sibling to 001's `--brief`).
- NEW: `internal/template/templates/.claude/skills/moai-workflow-project/references/navigator-audit.md` — Level-3 reference for the audit procedure + matching heuristic + override-file schema.
- NEW: `internal/template/templates/.claude/skills/moai-workflow-project/scripts/navigator-audit.sh` — self-contained bash audit script (sibling to `navigator-regen.sh`).
- EDIT (CI guard): `internal/template/internal_content_leak_test.go` — extend the forbidden-token sentinel set with the 002-specific leak guard (`SPEC-PROJECT-NAVIGATOR-002`, `REQ-NA-`).

### §A.4 Tier classification

**Tier M** (spec.md + plan.md + acceptance.md). Estimate: ~300–700 LOC, ~5–8 affected files. The audit algorithm is non-trivial but well-bounded (no new infrastructure category — reuses 001's script pattern, atomic-rename, fail-open, idempotence). REQ/AC counts (12 each) sit comfortably inside the Tier M ceilings (≤16).

## §B. Known Issues + [DECISION] markers

### §B.1 [DECISION] Matching heuristic — token normalization, not semantic

The audit's matching problem is: given a design-doc feature name (e.g. "Project Initialization (`moai init`)") and a capability-map row title (e.g. "CLI Tool — project init template selection"), decide whether they refer to the same feature. Three options surveyed:

| Heuristic | Fit | Verdict |
|-----------|-----|---------|
| Exact-string equality | Too brittle — phrasing always diverges between design docs and SPEC titles | REJECTED |
| Token-normalized fuzzy match (lowercase kebab tokens, substring containment, module-path token match) | Tolerates common phrasing divergence without semantic machinery; runs in pure bash | **CHOSEN (REQ-NA-007)** |
| Embeddings / semantic similarity | Requires a model call or a vector index; violates "self-contained bash, no model dependency" + Advisory-Check Discipline | REJECTED (out of scope per §E) |

The chosen heuristic is deliberately conservative: it produces **candidates**, not verdicts. The override file (REQ-NA-008) is the authoritative escape hatch for any divergence the heuristic cannot resolve, and the user is always the final adjudicator.

### §B.2 [DECISION] Standalone script vs extend `navigator-regen.sh`

**CHOSEN: standalone `scripts/navigator-audit.sh`.** Rationale recorded in spec.md §H.1. Briefly: regeneration (001) and analysis (002) are orthogonal concerns with different inputs and different idempotence contracts; coupling them in one script would muddy 001's completed surface and complicate debugging. The two scripts share the atomic-rename + fail-open + idempotence patterns by convention, not by shared code.

### §B.3 [DECISION] Dual output (md + json)

**CHOSEN: dual.** The markdown is the human surface; the JSON is the machine surface. Both must stay byte-identical across idempotent re-runs. The JSON schema is fixed at:

```json
{
  "audit_at": "<ISO-8601 from git log HEAD committer date>",
  "audit_commit": "<HEAD SHA>",
  "inputs": {
    "design_docs": ["product.md", "structure.md", "tech.md"],
    "capability_map": "capability-map.md",
    "override_file": "audit-known-matches.yaml | null"
  },
  "missing": [{"design_name": "...", "source": {"file": "...", "heading_path": "..."}, "closest_match": "spec-id | null"}],
  "orphan":  [{"spec_id": "...", "title": "...", "implementation_path": "..."}],
  "matched": [{"design_name": "...", "spec_id": "...", "match_basis": "exact | substring | module-token | override"}]
}
```

The schema is small and stable; a future LSEL cross-reference can consume `missing[]` directly.

### §B.4 [DECISION] Override file format — YAML

**CHOSEN: YAML** (`.moai/project/navigator/audit-known-matches.yaml`). Schema:

```yaml
# Optional user-authored override file for the Navigator audit.
# The audit loads this BEFORE applying its matching heuristic.
match:
  - design_name: "Autonomy Loop"
    spec_id: "SPEC-AUTONOMY-WORKFLOW-001"
ignore:
  - "SPEC-DEPRECATED-001"      # design-doc reference still present, deliberately ignored
  - "Old Feature Name"          # design-doc bullet that is no longer a real feature
```

Rationale (spec.md §H.4): YAML is the project config language, the schema is small, and the file stays human-editable.

### §B.5 [DECISION] Audit-time regeneration — NO

The audit does NOT trigger `navigator-regen.sh` (REQ-NA-002). Rationale: (1) keeps the audit read-only over its inputs, so the capability-map the user sees in the report is exactly what was on disk before the audit ran; (2) avoids a surprising side-effect (writing to 001's outputs) inside what the user expects to be a read-only analysis; (3) the recommended invocation order is `/moai project` (or `/moai sync`) → `/moai project --audit`, which already guarantees a fresh capability-map.

## §C. Integration map (constraint — compose with existing primitives, do not invent)

### §C.1 Component ↔ primitive

| Audit component | Existing primitive | Relationship |
|-----------------|--------------------|--------------|
| `--audit` skill mode | `/moai project` skill (`moai-workflow-project/SKILL.md`) | EXTENDS — a new mode alongside 001's `--brief` |
| `scripts/navigator-audit.sh` | `scripts/navigator-regen.sh` (001) | SIBLING — shares patterns (atomic-rename, fail-open, idempotence, self-contained bash) by convention, NOT by shared code |
| Design-doc inputs | `.moai/project/{product,structure,tech}.md` (generated by `/moai project` + hand-edited by user) | READ-ONLY consumer — never modifies design docs |
| Capability-map input | `.moai/project/navigator/capability-map.md` (generated by 001) | READ-ONLY consumer — never modifies 001's output |
| SPEC frontmatter input | `.moai/specs/SPEC-*/spec.md` (SPEC registry) | READ-ONLY consumer — used only to resolve override `spec_id` references and to cross-check capability-map rows |
| Provenance (audit_commit, audit_at) | `git log` | READ-ONLY consumer — identical to 001's provenance contract |
| Override file (optional) | `.moai/project/navigator/audit-known-matches.yaml` (user-authored) | READ-ONLY consumer — user writes it; audit loads it if present |
| Output report | `.moai/project/navigator/audit-report.{md,json}` | WRITES — atomic-rename, sibling directory to 001's outputs |
| Warning log | `.moai/logs/navigator-warnings.log` (shared with 001) | APPENDS — same log path as 001's malformed-frontmatter warnings |

### §C.2 Cross-subcommand relationship (Navigator as shared read primitive — extends 001 §C.2)

001 established that `/moai project` OWNS maintenance (write / regenerate / audit) and other subcommands CONSUME the Navigator at defined orientation phases. 002 adds the audit maintenance verb to `/moai project`. It does NOT add new cross-subcommand READ points — the audit output lives alongside 001's outputs under `.moai/project/navigator/` and is available to whoever reads it, but 002 introduces no new consultation requirements on `/moai plan`, `/moai run`, `/moai sync`, or the SessionStart hook. A future cross-reference (audit missing → LSEL propose) is recorded in §E but is NOT in scope.

### §C.3 Boundary vs LSEL + SPEC-003

| Axis | 002 (this SPEC) | LSEL | SPEC-003 |
|------|------------------|------|----------|
| Question answered | "what is missing / orphan between design and implementation?" | "how does the harness itself refine?" | "what symbols does the AST surface for the capability-map?" |
| Primary input | design docs + capability-map | lessons-inbox + usage-log | source code + tree-sitter grammars |
| Primary output | `audit-report.{md,json}` | `feedback_*.md` + `hns-*` proposals | enriched capability-map rows |
| Write surface | `.moai/project/navigator/audit-report.*` | `memory/`, `hns-lsel-*`, `CLAUDE.local.md` | (003's run-phase scope) |
| Read surface | design docs + capability-map + SPEC frontmatter | lessons-inbox + usage-log | source code |

REQ-NA-011 mechanically enforces the non-overlap: the audit script's read-set + write-set grep MUST NOT touch any LSEL or SPEC-003 surface.

## §D. Algorithm specification (the audit step-by-step)

This is the contract `scripts/navigator-audit.sh` implements. Documented here so reviewers can challenge the algorithm before run-phase codes it.

### §D.1 Step 1 — Extract design-intent features

For each of `product.md`, `structure.md`, `tech.md`:

1. Parse markdown headings (`#`, `##`, `###`, `####`) into a heading path stack.
2. Identify "feature-listing" sections by sentinel phrases in the heading text: `Core Features`, `Features`, `Capabilities`, `Modules`, `Functionality`, `Subsystems`, `Components`. (Sentinel set is closed and documented in the Level-3 reference; the user can extend via override.)
3. Inside a feature-listing section, treat each sub-heading OR each bolded bullet (`- **<name>** ...`) as a named feature. The feature name is the heading text or the bolded text (without `**`), with trailing parenthetical (e.g. ` (moai init)`) stripped to a separate `notes` field.
4. Record each feature as `(name, source_file, heading_path)`.

Output: a TSV stream `name<TAB>file<TAB>heading_path`.

### §D.2 Step 2 — Extract capability-map features (header-driven)

Parse `.moai/project/navigator/capability-map.md`. Column resolution is **header-driven** (REQ-NA-007) — this is the design choice that immunizes 002 against 001's unfrozen/inconsistent column schema (001's spec.md declares column 1 as `capability`, capability-first; 001's acceptance.md AC-PN-013 enumerates rows spec-id-first; 001's plan defers the exact schema to its own M1).

1. Parse the capability-map's **header row** (the `| ... | ... |` row above the `|---|---|` separator) and resolve columns **by name**: the feature/name column (recognize any spelling 001 may use: `capability`, `name`, `feature`, `title`), the spec-id column (`owning-spec`, `spec-id`, `spec_id`), the status column (`status`), and the implementation-path column (`implementation-path`, `path`, `module-path`). Match header cells case-insensitively, trimmed, treating `-`/`_`/space as equivalent.
2. If the header lacks a required column (no feature/name column OR no spec-id column), skip the row and emit a warning to `.moai/logs/navigator-warnings.log` naming the missing column.
3. For each remaining data row, extract feature/name, spec-id, status, and implementation-path **by header-resolved position** — NOT by fixed index.
4. Filter OUT rows whose `status` is `superseded`, `archived`, or `rejected` (those represent retired features; flagging them as orphans would be noise). Keep `draft`, `planned`, `in-progress`, `implemented`, `completed`.

Output: a TSV stream `spec-id<TAB>title<TAB>implementation-path`. (The previous draft's `module` alias is dropped — `module` in 001 is a SPEC-frontmatter field, a distinct concept from 001's column-4 `implementation-path`. 002 uses 001's actual column name.)

### §D.3 Step 3 — Load overrides (if present)

If `.moai/project/navigator/audit-known-matches.yaml` exists, parse it (a minimal YAML reader — the schema is small enough that a pure-bash awk/sed extractor suffices, no `yq`/`jq`):

- `match`: list of `{design_name, spec_id}` pairs → forced matches.
- `ignore`: list of design names OR spec-ids → excluded from missing / orphan candidate lists.

### §D.4 Step 4 — Compute the diff

For each design feature `D` and each capability-map row `C`:

1. If `(D.name, C.spec_id)` is in `override.match` → record as `matched` with `match_basis: override`.
2. Else if `D.name` or `C.spec_id` is in `override.ignore` → skip (not emitted anywhere).
3. Else apply the heuristic (REQ-NA-007):
   - Normalize `D.name` and `C`'s title-column value to lowercase kebab tokens.
   - (a) exact normalized equality → match_basis `exact`.
   - (b) substring containment (shorter ≥4 chars) → match_basis `substring`.
   - (c) the last path segment of `C`'s implementation-path column (the path token) appears as a token in normalized `D.name` AND that path token is ≥4 characters → match_basis `module-token`. The ≥4-character floor rejects trivially short segments (`cmd`, `pkg`, `src`, `crm`) that would otherwise match unrelated design-doc features and inflate false positives.
4. If no heuristic match for `D` after iterating all `C` rows and no override → `D` is a **Missing SPEC** candidate.
5. If no heuristic match for `C` after iterating all `D` features and no override → `C` is an **Orphan SPEC** candidate.

### §D.5 Step 5 — Emit the report

Render `audit-report.md` (human-readable) + `audit-report.json` (machine-readable) per §B.3 / REQ-NA-005, write both via atomic-rename, then write the JSON `audit_commit` sentinel last.

## §E. Boundary vs LSEL + SPEC-003 (forward-looking, NOT in scope)

The audit's `missing[]` list is a natural input to LSEL's PROPOSE phase ("missing SPEC detected → propose creating it"). This is a **forward-looking cross-reference**, NOT in scope for 002 or for LSEL today:

- 002 emits `audit-report.json` with `missing[]`; that is the full extent of 002's contribution to this cross-reference.
- LSEL does NOT currently consume `audit-report.json`. A future SPEC (or an LSEL amendment) MAY wire that consumption; doing so is the LSEL owner's prerogative and is out of scope here.

## §F. Milestones (priority-ordered; no time estimates)

Order rationale: highest-change-likelihood first. M1 (algorithm + output format) is the most likely to be revised during review; M4 (CI guards) is the most mechanical.

### M1 — Audit algorithm + output format [Priority High]

- Specify the exact JSON schema (§B.3) and markdown layout.
- Specify the matching heuristic (§D / REQ-NA-007).
- Specify the override-file schema (§B.4).
- Specify idempotence (REQ-NA-009) + fail-open (REQ-NA-010) contracts.
- Deliverable: a documented schema + a fixture design-doc set + a fixture capability-map that exercises match / missing / orphan / override.

### M2 — Audit script (`scripts/navigator-audit.sh`) [Priority High]

- Implement the §D algorithm in self-contained bash (`git` + `awk` + `sed` + `grep` only).
- Implement atomic-rename writes (inherited from 001 §B.2).
- Implement fail-open on missing inputs (REQ-NA-010).
- Implement override-file loading (REQ-NA-008).
- Deliverable: the script runs end-to-end on the fixture set, producing byte-identical output across two runs (idempotence AC-NA-009).

### M3 — Skill mode + Level-3 reference [Priority High]

- Extend `moai-workflow-project/SKILL.md` with a `--audit` mode (sibling to `--brief`).
- Author `references/navigator-audit.md` (Level 3) documenting the procedure, the heuristic, the override schema, the fail-open semantics.
- Wire the skill mode to invoke `scripts/navigator-audit.sh`.
- Deliverable: `/moai project --audit` on a fixture project produces the expected report; the skill body + reference are template-neutral (AC-NA-012).

### M4 — Template mirror + CI guards [Priority Medium]

- Mirror all template edits to `internal/template/templates/` + run `make build`.
- Extend `internal/template/internal_content_leak_test.go` with the `SPEC-PROJECT-NAVIGATOR-002` + `REQ-NA-` sentinels.
- Verify §25 neutrality checklist (5-item pre-commit) on each new template file.
- Verify 16-language neutrality (run audit on a non-Go fixture).
- Deliverable: `make build` clean; CI guards green; neutrality checklist passes.

### M5 — Plan-audit gate + Implementation Kickoff Approval [Priority Medium]

- Run plan-auditor on this plan-phase artifact set; resolve findings (max 3 iterations per the plan-auditor retry contract).
- Implementation Kickoff Approval human gate at plan→run (constraint: PR-mandatory + `enforce_admins:true`).
- Deliverable: plan-auditor PASS; user-confirmed progression-mode + gate pass.

## §G. Anti-Patterns (what NOT to do)

- **AP-NA-001** — Trigger regeneration inside the audit (violates REQ-NA-002 read-only).
- **AP-NA-002** — Modify design docs or capability-map as a side-effect (violates REQ-NA-002 + REQ-NA-006 provenance integrity).
- **AP-NA-003** — Emit a wall-clock timestamp in the report (violates REQ-NA-009 idempotence — timestamps must come from `git log`).
- **AP-NA-004** — Auto-create a SPEC for a Missing-SPEC candidate (violates §E "auto-creating / auto-retiring SPECs" Out of Scope).
- **AP-NA-005** — Wire the audit as a SessionStart / Stop / PostToolUse hook (violates §E "hook-based audit gating" Out of Scope + Advisory-Check Discipline).
- **AP-NA-006** — Use `jq` / `yq` / `python` / `moai` binary in the audit script (violates the self-contained-bash portability contract inherited from 001).
- **AP-NA-007** — Hardcode this repo's SPEC IDs or commit SHAs into template-distributed files (violates §25.1 C2/C3/C7).
- **AP-NA-008** — Couple the audit to SPEC-003's AST surface (violates REQ-NA-011 boundary).
- **AP-NA-009** — Read any LSEL surface (`.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*`) inside the audit (violates REQ-NA-011 boundary).

## §H. Cross-References

- `acceptance.md` (this SPEC dir) — Given-When-Then ACs for every REQ.
- `.moai/specs/SPEC-PROJECT-NAVIGATOR-001/` — the substrate SPEC. 001's `capability-map.md` output is the surface this audit diffs against; the column schema is parsed **header-driven** (REQ-NA-007) because 001's spec.md / acceptance.md / plan.md do not agree on a frozen column order.
- `.claude/skills/moai-workflow-project/SKILL.md` + `references/navigator.md` + `scripts/navigator-regen.sh` — 001's shipped skill / reference / script surfaces that 002 extends with a sibling.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — `(none) → draft` emitted here (002 was a stub; this plan-phase expansion keeps `status: draft`).
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` — `era: V3R6` explicit override rationale.
- `.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline — fail-open + time-boxed rationale (REQ-NA-010).
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — provenance attribution rationale (REQ-NA-006).
- CLAUDE.local.md §2 (Template-First), §15 (16-language neutrality), §23 (PR-mandatory), §25 (template-neutrality) — constraints inherited.
