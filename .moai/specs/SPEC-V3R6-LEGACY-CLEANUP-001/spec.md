---
id: SPEC-V3R6-LEGACY-CLEANUP-001
title: "v2.x agency keyword residual cleanup (scope C — user-facing docs + skills + rules + docs-site)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: Medium
tags: "cleanup, legacy, v3-roadmap, sprint-2, docs, brand-design"
issue_number: null
tier: L
phase: "v3.0.0"
module: "docs"
lifecycle: spec-anchored
related_specs:
  - SPEC-AGENCY-ABSORB-001
  - SPEC-V3R6-CHANGELOG-CLEANUP-001
---

# SPEC-V3R6-LEGACY-CLEANUP-001 — v2.x agency Keyword Residual Cleanup

## §A — Pre-existing State Survey (Mandatory per `feedback_plan_auditor_codebase_state_blindspot` mitigation #1)

This section documents the verified state of the codebase BEFORE this SPEC's run-phase. It is the foundation for all subsequent REQ/AC scope decisions.

### §A.0 Reconnaissance Methodology

Verification command used (orchestrator, 2026-05-23):

```bash
grep -rln -E '\bagency\b|\.agency/|/agency/' . \
  --exclude-dir=.git \
  --exclude-dir=node_modules \
  --exclude-dir=vendor \
  --exclude-dir=.moai/design \
  --exclude-dir=.moai/brain \
  --exclude-dir=.moai/marketing \
  --exclude-dir=.moai/specs/SPEC-AGENCY-ABSORB-001 \
  --exclude-dir=.moai/archive \
  --exclude-dir=.moai/plans \
  | sort -u
```

Total matches after PRESERVE exclusions: **103 files**. Of these, this SPEC's scope C addresses only the **user-facing documentation subset** (see §A.1 inventory).

### §A.1 Inventory — scope C (29 files in scope) [CORRECTED from spawn prompt §C]

Spawn prompt §C claimed 31 files. Actual grep reveals **29 user-facing files** + 74 out-of-scope files (see §A.6 below). Discrepancy origin: spawn prompt §C overcounted docs-site by 2 (claimed ko:6 + en:6 with `code-based-path.md` containing agency keyword; verified `code-based-path.md` exists in all 4 locales but does NOT contain `agency` keyword). The corrected inventory:

**Root markdown (4 files)** — top-level project documentation:
- `CHANGELOG.md`
- `CLAUDE.md`
- `README.md`
- `README.ko.md`

(Note: `README.ja.md` and `README.zh.md` also contain `agency` keyword per actual grep — adding to root MD set. See §A.1.7 below for the +2 correction.)

**Skills (4 files)** — Claude Code skill bodies:
- `.claude/skills/moai-domain-brand-design/SKILL.md`
- `.claude/skills/moai-domain-copywriting/SKILL.md`
- `.claude/skills/moai-workflow-gan-loop/SKILL.md`
- `.claude/skills/moai/workflows/design.md`

**Rules (1 file)** — design constitution:
- `.claude/rules/moai/design/constitution.md`

**docs-site 4-locale (20 files, perfectly symmetric)** — verified via actual grep:
- ko (5): `_index.md`, `gan-loop.md`, `getting-started.md`, `migration-guide.md` (under `design/`) + `moai-design.md` (under `workflow-commands/`)
- en (5): same 5 files as ko
- ja (5): same 5 files as ko
- zh (5): same 5 files as ko

**4-locale parity finding**: `code-based-path.md` EXISTS in all 4 locales but does NOT contain `agency` keyword. The spawn prompt §C's "asymmetric coverage" claim was incorrect — the agency keyword distribution across locales is **perfectly symmetric (5/5/5/5)**.

### §A.1.7 README +2 correction

Actual grep includes `README.ja.md` and `README.zh.md` (Japanese + Chinese README), which spawn prompt §C omitted. Adding these:

- Corrected root markdown count: **6 files** (was 4): `CHANGELOG.md`, `CLAUDE.md`, `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`
- Corrected total scope C: **6 (root) + 4 (skills) + 1 (rule) + 20 (docs-site) = 31 files** ← matches spawn prompt §C numerically by coincidence, but the breakdown differs

**Final inventory total for this SPEC's scope: 31 files** (root 6 + skills 4 + rule 1 + docs-site 20).

### §A.2 PRESERVE list — paths intentionally untouched

The following paths contain `agency` keyword but MUST NOT be modified by this SPEC. The 74 out-of-scope files break down as:

**Archives (explicit exclusion)**:
- `.moai/design/v3-legacy/**` (historical architecture archive)
- `.moai/design/v3-research/**` (Wave 1 research artifacts)
- `.moai/design/v3-redesign/research/**` (Wave 1 research artifacts)
- `.moai/brain/IDEA-*/**` (ideation archive)
- `.moai/marketing/**` (separate decommissioning policy)
- `.moai/specs/SPEC-AGENCY-ABSORB-001/**` (self-reference archive)
- `.moai/archive/**` (general archive)
- `.moai/plans/DOCS-SITE/**` (migration log)
- `.moai/specs/_archive/**` (archived SPEC `SPEC-AGENCY-001`, `SPEC-THIN-CMDS-001`)

**Historical SPEC documents (treat as audit trail, PRESERVE)**:
- `.moai/specs/SPEC-AGENT-002/`, `SPEC-CORE-BEHAV-001/`, `SPEC-DESIGN-DOCS-001/`, `SPEC-OPUS47-COMPAT-001/`, `SPEC-SKILL-002/`, `SPEC-SKILL-ENHANCE-001/`, `SPEC-SRS-003/` (7 SPECs documenting v2.x→v3.x migration intent)
- `.moai/specs/SPEC-V3R2-*/`, `SPEC-V3R3-*/`, `SPEC-V3R4-LINT-SKIP-CLEANUP-001/`, `SPEC-V3R6-RULES-PATH-SCOPE-001/`, `SPEC-V3R6-SKILL-COMPRESS-001/`, `SPEC-V3R6-UPDATE-NOISE-001/` (~38 files; all completed SPECs whose body references agency as historical context)

**Production Go code (separate SPEC scope — out of legacy-cleanup)**:
- `internal/cli/migrate_agency*.go` (9 files: command implementation + tests + disk/posix/signal/windows variants) — the `moai migrate agency` CLI command must keep its identifier
- `internal/cli/migration.go`, `update_cleanup.go`, `update_cleanup_test.go`, `update_e2e_test.go` — agency keyword referenced as enum/migration target
- `internal/defs/dirs.go`, `internal/evolution/safety.go`, `internal/evolution/security_test.go`, `internal/research/safety/frozen.go`, `internal/research/safety/frozen_test.go` — safety guard code
- `internal/template/commands_audit_test.go`, `commands_root_audit_test.go`, `deployer_bench_test.go` — template audit tests

**Template mirror files (cascading edit — out of THIS SPEC's scope)**:
- `internal/template/templates/.claude/rules/moai/design/constitution.md`
- `internal/template/templates/.claude/skills/moai-domain-brand-design/SKILL.md`
- `internal/template/templates/.claude/skills/moai-domain-copywriting/SKILL.md`
- `internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md`
- `internal/template/templates/.claude/skills/moai/workflows/design.md`
- `internal/template/templates/.moai/config/sections/design.yaml`
- `internal/template/templates/CLAUDE.md`

These 7 template mirrors **must be updated in lockstep with the 5 source files** under `.claude/rules/` + `.claude/skills/` + `CLAUDE.md` per CLAUDE.local.md §2 [HARD] Template-First Rule. **DECISION**: This SPEC's scope includes the source files only; the run-phase implementer is responsible for invoking `make build` after editing source files, which will regenerate `internal/template/embedded.go` and pick up template changes IF the implementer ALSO copies edits to `internal/template/templates/`. Acceptance criterion AC-LCL-006 below enforces this.

**Miscellaneous (PRESERVE by category)**:
- `.moai/decisions/skill-rename-map.yaml` (rename audit trail)
- `.moai/project/structure.md` (auto-generated by `/moai project`)
- `.moai/research/v3.0-redesign-2026-05-23.md` (forbidden per spawn prompt B9)
- `docs-site/vercel.json` (build config, not user-facing prose)
- `docs/design/major-v3-master.md` (master design doc, separate cleanup SPEC candidate)

Total preserved (74 files): 9 archives + 38 historical SPECs + 19 production Go + 7 template mirrors + 1 yaml decision + 1 structure.md + 1 research doc + 1 vercel.json + 1 master design doc + ~marginal others.

### §A.3 Replacement semantics — per-file judgment required (PARTIAL VERIFICATION DISCLOSURE)

Per-file inspection of all 31 in-scope files was **NOT performed during plan-phase**; semantic judgment deferred to run-phase. The 4 replacement categories below provide the framework, but the precise edit per file is implementer-determined during M2/M3.

**Four replacement categories**:

1. **Reference to v2.x architecture concept** → replace with `"v2.x agency (retired, see SPEC-AGENCY-ABSORB-001)"` reference, OR remove the sentence if redundant. Applies to most skill body mentions.

2. **Reference to `.agency/` directory path** → remove or redirect to current path. The `.agency/` directory was removed in SPEC-AGENCY-ABSORB-001; any path reference is stale.

3. **Skill `description:` frontmatter mention** → review whether the skill is still aligned with absorption status; likely needs description rewrite. Applies to: `moai-domain-brand-design`, `moai-domain-copywriting`, `moai-workflow-gan-loop` (3 skill descriptions inferred to need refresh).

4. **Historical CHANGELOG entry** → preserve (CHANGELOG is append-only historical record). Per REQ-LCL-004, only `[Unreleased]` and v3.0+ entries are subject to cleanup.

**Partial verification statement**: This SPEC's plan-phase verified file-existence + scope-membership for all 31 files; it did NOT inspect the per-file semantic context required to decide between category 1, 2, 3, 4 for each occurrence. Run-phase MUST perform per-file inspection before edit.

### §A.4 Working tree state (verified via `git status` + `git log -1` at plan-phase entry)

- Current branch: `main`
- HEAD: `87dd61564` — `chore(SPEC-V3R6-CHANGELOG-CLEANUP-001): M3 B12 standing-rule guard — sync-phase CHANGELOG emission discipline`
- Modified: `.moai/harness/usage-log.jsonl` (1 file, runtime-managed — IGNORE)
- Untracked: `.moai/harness/observations.yaml`, `.moai/research/v3.0-redesign-2026-05-23.md` (forbidden per B9)
- Untracked: `.moai/specs/SPEC-V3R6-CHANGELOG-CLEANUP-001/` (sister SPEC in flight)
- Stale stash entries: 2 (`wave6-cce-residual` + `gears-migration-001`) — unrelated to this SPEC

**HEAD divergence note**: Spawn prompt §A.4 referenced HEAD `731aa0df5`; actual verified HEAD at plan-phase entry is `87dd61564` (4 commits ahead — parallel session race, `feedback_large_spec_wave_split` L9 reinforced). No impact on this SPEC's scope; the discrepancy is timestamp-only.

### §A.5 SPEC Tier classification + Hybrid Trunk policy override

**Tier reclassification (iter-2 fix-forward, plan-auditor B2 finding)**:
- Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (SSoT): Tier S < 5 files, Tier M 5-15 files, Tier L > 15 files OR constitutional. PASS thresholds: 0.75 / 0.80 / 0.85.
- This SPEC scope = **31 files** (root 6 + skills 4 + rule 1 + docs-site 20). 31 > 15 → **Tier L**.
- iter-1 spec.md mislabeled as Tier M citing "5-30 files coordinated" which contradicts the SSoT 5-15 range. Corrected in iter-2.

**Tier L 5-artifact obligation — relaxed per user decision 2026-05-23**:
- Standard Tier L requires 5 artifacts (spec/plan/acceptance/design/research). This SPEC produces only 4 (spec/plan/acceptance/progress) because the work is documentation-only (no functional code changes; no external research required).
- The `design.md` content is absorbed into spec.md §A.3 (4-category replacement framework) + §B.3 (surgical edit contract demoted to design notes in iter-2).
- The `research.md` content is absorbed into spec.md §A.1 (31-file precise inventory with provenance) + §A.2 (PRESERVE list with 74 out-of-scope files enumeration) + §A.6 (4 follow-up SPEC candidates).
- Lesson L31 candidate: Tier L documentation-only SPECs MAY consolidate design+research into spec.md §A when the work involves no functional code. Future canonical rule update candidate: `spec-workflow.md` § SPEC Complexity Tier — add note for doc-only Tier L exemption.

**Branch policy override (CLAUDE.local.md §23 [HARD]) — 1-person OSS explicit per-SPEC override 2026-05-23**:
- §23 [HARD] requires Tier M/L SPECs to use `feat/<SPEC-ID>` branch + auto PR. main direct push is allowed only for "Tier S 미만".
- This SPEC overrides §23 [HARD] explicitly: run-phase commits push directly to `main` (no PR) under 1-person OSS Hybrid Trunk policy. Rationale: (a) doc-only edits with byte-level reversibility via `.moai/backups/`, (b) 4 CI status checks (Test/Lint/Build/CodeQL) gate main push providing safety net equivalent to PR review for 1-person workflow, (c) pre-push hook warn-only + 5s sleep (CLAUDE.local.md §23.1) provides last-mile abort window.
- **Override scope**: Per-SPEC only. Default flow for future Tier L SPECs remains §23 [HARD] (feat-branch + auto PR). Override approval recorded: user decision 2026-05-23 AskUserQuestion Q2.
- Lesson L32 candidate: §23 [HARD] per-SPEC override pattern requires spec.md §A.5 documentation + commit message body annotation.

**Tier L policy validations**:
- File count: 31 (Tier L scope, 0.85 PASS threshold)
- Per-milestone files-touched: ≤10 (Hybrid Trunk Tier M files-per-M discipline retained as voluntary practice; reduces blast radius per commit)
- CI status checks: 4 (Test ubuntu / Lint / Build linux/amd64 / CodeQL) — all must pass for each milestone commit
- `pre-push` hook: warn-only + 5s sleep (CLAUDE.local.md §23.1)
- plan-auditor PASS threshold: 0.85 (Tier L per SSoT)

### §A.6 Out-of-scope but related — separate SPEC candidates

The following follow-up SPECs are explicitly out of THIS SPEC's scope and may be created post-merge:

- **SPEC-V3R6-LEGACY-CLEANUP-002** (template mirror cascade): Update the 7 `internal/template/templates/.claude/...` mirror files in lockstep with this SPEC's source edits.
- **SPEC-V3R6-LEGACY-CLEANUP-003** (production code): Audit `internal/cli/migrate_agency*.go` to decide if `moai migrate agency` command should be retired or rename to `moai migrate v2x-agency`.
- **SPEC-V3R6-LEGACY-CLEANUP-004** (master design doc): Cleanup `docs/design/major-v3-master.md` reference to v2.x agency.
- **SPEC-V3R6-LEGACY-CLEANUP-005** (historical SPEC archive): Decide whether the 38 `SPEC-V3R*` + 7 pre-v3.0 historical SPECs should remain in `.moai/specs/` or be moved to `.moai/specs/_archive/`.

---

## §B — Requirements (EARS Format)

### §B.1 Backup contract

**REQ-LCL-001** [Ubiquitous]: The cleanup process SHALL create a backup of all 31 in-scope files at `.moai/backups/legacy-cleanup-{ISO-DATE}/` before any modification, where `{ISO-DATE}` follows the format `YYYY-MM-DDTHHMMSSZ` (UTC).

**REQ-LCL-002** [Ubiquitous]: The backup manifest at `.moai/backups/legacy-cleanup-{ISO-DATE}/manifest.json` SHALL list all backed-up paths with their SHA256 hash and original byte-size.

**REQ-LCL-003** [Event-Driven]: WHEN the backup creation step fails for any file, THEN the cleanup process SHALL halt immediately and emit a structured error report to stderr listing the failed path and OS error.

### §B.2 PRESERVE contract

**REQ-LCL-004** [Ubiquitous]: The following paths SHALL remain byte-identical after run-phase completion (verified by SHA256 comparison pre/post):
- `.moai/design/**`
- `.moai/brain/**`
- `.moai/marketing/**`
- `.moai/specs/SPEC-AGENCY-ABSORB-001/**`
- `.moai/archive/**`
- `.moai/plans/**`
- `.moai/specs/_archive/**`
- `.moai/specs/SPEC-V3R*/` (38 files, historical audit trail)
- `.moai/specs/SPEC-AGENT-002/`, `SPEC-CORE-BEHAV-001/`, `SPEC-DESIGN-DOCS-001/`, `SPEC-OPUS47-COMPAT-001/`, `SPEC-SKILL-002/`, `SPEC-SKILL-ENHANCE-001/`, `SPEC-SRS-003/`
- `internal/**/*.go` (all production Go code)
- `internal/template/templates/**` (template mirror — separate SPEC scope)
- `.moai/decisions/skill-rename-map.yaml`
- `.moai/project/structure.md`
- `.moai/research/v3.0-redesign-2026-05-23.md`
- `docs-site/vercel.json`
- `docs/design/major-v3-master.md`

### §B.3 Surgical edit (design notes, not normative REQs)

**[iter-2 fix-forward, plan-auditor S4]**: The following 2 items were originally drafted as REQ-LCL-005 + REQ-LCL-006 but are not binary-testable. Per user decision 2026-05-23 (AskUserQuestion Q4), they are demoted to **design notes** governing run-phase implementer judgment. They are NOT REQ-LCL-XXX entries; no AC links to them.

**Design Note D-1 (formerly REQ-LCL-005)**: Each modification SHOULD preserve sentence and paragraph meaning; `agency` keyword removal SHOULD NOT delete factual claims about v2.x architecture history. Verification: implementer judgment + diff review during run-phase, NOT a binary AC.

**Design Note D-2 (formerly REQ-LCL-006)**: WHILE editing a file in the in-scope inventory, the implementer SHOULD classify each `agency` occurrence into one of four categories (per §A.3): (1) v2.x architecture concept, (2) `.agency/` directory path, (3) skill `description:` frontmatter mention, (4) historical CHANGELOG entry. Verification: per-commit message body annotation suggested but not enforced.

### §B.4 CHANGELOG append-only contract

**REQ-LCL-007** [Unwanted]: The cleanup process SHALL NOT remove `agency` keyword from CHANGELOG.md entries older than v3.0 (these are historical append-only records).

**REQ-LCL-008** [Optional]: Where the `agency` keyword appears in CHANGELOG.md `[Unreleased]` section or v3.0+ entries, the cleanup process SHOULD remove or refactor the reference to point at `SPEC-AGENCY-ABSORB-001` for traceability.

### §B.5 4-locale parity preservation

**REQ-LCL-009** [Ubiquitous]: The cleanup process SHALL NOT add or remove any docs-site locale files; only the `agency` keyword inside the 20 existing docs-site in-scope files SHALL be modified.

**REQ-LCL-010** [Ubiquitous]: The pre-existing 4-locale symmetry (5 files per locale × 4 locales = 20 files containing `agency` keyword) SHALL be preserved; any edit applied to a ko file SHALL be applied semantically equivalently to its en, ja, zh siblings.

### §B.6 Build and regression contracts

**REQ-LCL-011** [Ubiquitous]: After cleanup completion, `hugo --source docs-site` SHALL exit with status 0 (docs-site Hugo build must remain green).

**REQ-LCL-012** [Ubiquitous]: After cleanup completion, `go test ./...` PASS rate SHALL be maintained (this SPEC modifies no `.go` files; expected regression count = 0).

### §B.7 Verification visibility

**REQ-LCL-013** [Ubiquitous]: After cleanup completion, the grep query `grep -rln -E '\bagency\b|\.agency/|/agency/' <31-in-scope-paths>` SHALL return ≤5 occurrences. Threshold rationale: up to 1 retired-reference per top-level user-facing doc — {CHANGELOG, CLAUDE, README × any one canonical locale, 2 design skill bodies} = ~5 expected residuals from category 1 replacement strategy (Design Note D-2). Docs-site retired-references SHOULD be inlined into a shared paragraph rather than per-locale repetition.

### §B.8 Exclusion contracts (Unwanted requirements)

**[iter-2 fix-forward, plan-auditor S5 traceability]**: AC-LCL-009 and AC-LCL-010 in iter-1 linked to `§C exclusion #1/#2` (scope statements). iter-2 promotes those exclusions to formal Unwanted REQs so the AC traceability is REQ-to-AC clean.

**REQ-LCL-014** [Unwanted]: This SPEC's run-phase SHALL NOT modify any file under `internal/**/*.go` (all production Go code, including `internal/cli/migrate_agency*.go`). Production code audit is deferred to SPEC-V3R6-LEGACY-CLEANUP-003.

**REQ-LCL-015** [Unwanted]: This SPEC's run-phase SHALL NOT modify any file under `internal/template/templates/**` (template mirror files). Template mirror cascade is deferred to SPEC-V3R6-LEGACY-CLEANUP-002, which MUST follow IMMEDIATELY after this SPEC merges to restore template-source parity (CLAUDE.local.md §2 [HARD] Template-First Rule).

---

## §C — Exclusions (What NOT to Build)

This SPEC explicitly excludes the following work to maintain Tier M scope and avoid cascading concerns:

1. **Production Go code edits** — `internal/cli/migrate_agency*.go` and related (19 files) are excluded; the `moai migrate agency` CLI command keeps its identifier. Future SPEC: SPEC-V3R6-LEGACY-CLEANUP-003.

2. **Template mirror file edits** — `internal/template/templates/.claude/...` (7 files) are excluded from THIS SPEC. Future SPEC: SPEC-V3R6-LEGACY-CLEANUP-002. Note: This creates a temporary template-source desync; SPEC-V3R6-LEGACY-CLEANUP-002 must follow IMMEDIATELY after this SPEC merges.

3. **Historical SPEC archive moves** — The 38+ historical SPEC documents remain in `.moai/specs/` and are NOT moved to `.moai/specs/_archive/`. Future SPEC: SPEC-V3R6-LEGACY-CLEANUP-005.

4. **Docs-site `code-based-path.md` content** — These 4 locale files exist but do NOT contain `agency` keyword; they are out of scope by definition.

5. **vercel.json build config** — Out of scope (not user-facing prose).

6. **Master design doc cleanup** — `docs/design/major-v3-master.md` excluded; Future SPEC: SPEC-V3R6-LEGACY-CLEANUP-004.

7. **CHANGELOG historical entries** — Pre-v3.0 CHANGELOG entries preserved as append-only audit trail (REQ-LCL-007).

8. **Adding new locale files** — Strict 4-locale parity preservation (REQ-LCL-009/REQ-LCL-010).

9. **AGENCY-ABSORB-001 self-references** — The SPEC `.moai/specs/SPEC-AGENCY-ABSORB-001/` itself is the canonical reference; no edits to that directory.

10. **Run-phase execution** — This SPEC is plan-phase only. Implementation is deferred to a separate `/moai run` invocation.
