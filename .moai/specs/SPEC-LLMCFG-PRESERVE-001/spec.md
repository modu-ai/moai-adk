---
id: SPEC-LLMCFG-PRESERVE-001
title: "User-edited llm.yaml survives moai update — keep the merge pipeline, pin the contract"
version: "0.1.0"
status: draft
created: 2026-09-02
updated: 2026-09-02
author: manager-spec (card t239)
priority: P2
phase: "v3.1.5 target"
module: "internal/cli, internal/cli/update/backup, internal/cli/update/merge, internal/cli/update/deploy"
lifecycle: spec-anchored
tags: "update, llm-yaml, config-preservation, three-way-merge, regression-guard, template-sync, clean-reinstall"
tier: M
related_specs:
  - SPEC-V3R6-AUDIT-MODEL-PIN-001
  - SPEC-UPDATE-YAML-PRESERVE-001
  - SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001
---

# SPEC-LLMCFG-PRESERVE-001 — User-edited llm.yaml survives `moai update`

> Card t239 (Class C — preservation-policy design decision, Tier M). The card's
> stated defect premise is **refuted in part by the tree**: the audit model pins
> t225 landed live in `workflow.yaml` `workflow.audit` (tracked), NOT in
> llm.yaml — SPEC-V3R6-AUDIT-MODEL-PIN-001 v1.1.0 MF1 relocated them off
> llm.yaml precisely because llm.yaml is gitignored and update-wiped. What the
> card correctly names is the defect CLASS: llm.yaml is user-editable,
> gitignored, machine-modified at runtime, and sits inside the `.moai/config`
> root that `CleanMoaiManagedPaths` deletes wholesale before redeploy. This SPEC
> records what actually protects that file today, decides the policy question
> the card asks, and closes the one gap that survives verification.

## §A Problem / Context

### §A.1 The mechanical constraint (verified)

`CleanMoaiManagedPaths` (`internal/cli/update/deploy/deploy.go:107`) removes the
ENTIRE `.moai/config` directory before template redeployment
(`deploy.go:187-204`). It is not diff-based. Its pre-clean backup (card t111,
`deploy.go:96-106`) saves only files the embedded template does NOT carry —
llm.yaml IS template-carried, so it never enters the pre-clean backup. A bare
"don't ship the template file" exclusion therefore cannot protect user content:
the directory wipe kills it regardless.

### §A.2 What already protects llm.yaml (verified on this tree)

The update pipeline already runs a full preservation cycle:

1. **Backup** step (`update_template_sync.go:398`): `BackupMoaiConfig`
   (`internal/cli/update/backup/backup.go:27`) copies ALL of
   `.moai/config` (the `excludedDirs` list is empty — the "excluding sections"
   comment above it is stale) to `.moai-backups/<timestamp>/`, and
   `SaveTemplateBase` (`backup/base_loader.go`) provisions the merge base —
   template-base snapshot first (SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001),
   falling back to `SaveTemplateDefaults` (the current embedded defaults).
2. **Restore Settings** step (`update_template_sync.go:472-524`):
   `RestoreMoaiConfigRetained` (`backup/restore.go`) walks the backup's
   `sections/*.yaml` — llm.yaml included, no exclusion — and 3-way merges
   (new template + user backup + base) into the freshly deployed file.
3. The merge engine is node-tree based (`backup/merge.go`, `backup/node_merge.go`):
   comments, key order, and scalar quoting survive (`SPEC-UPDATE-YAML-PRESERVE-001`,
   issue #1243, landed 2026-08-03); user-divergent values win over template
   defaults; template-only keys are delivered; user-only keys are retained with
   an advisory (issues #1265/#1267, landed 2026-08-01 — the `llm.agent_overrides`
   case). Scoped tests pass on this tree: `go test -count=1
   ./internal/cli/update/backup/` → `ok ... 0.626s` (this run, HEAD `b7462203a`).

llm.yaml itself is gitignored (`.gitignore`, commented block: "Per-machine
runtime LLM backend config … Expanded + modified at runtime by moai glm/cg;
the tracked template default lives at internal/template/templates/.moai/config/
sections/llm.yaml"). Its tracked template twin ships documentation comments for
~90% of its 247 lines.

### §A.3 The verified gap

llm.yaml is the ONLY major section file with no named end-to-end preservation
test on either update path:

- `TestCleanReinstall_AllSectionsYAMLPreserved`
  (`internal/cli/update_clean_install_config_preserve_test.go:195`) exercises
  user.yaml, language.yaml, design.yaml — not llm.yaml.
- The template-sync path has section-level merge unit tests (synthetic YAML)
  and the llm-specific `merge_useradd_test.go` (`llm.agent_overrides`,
  issue #1267) — but no test drives the REAL embedded template llm.yaml
  (247 lines, comment-dense) through the update cycle asserting that user
  values, comments, and new keys all survive together.
- The only llm.yaml reference in update tests outside the backup package is a
  synthetic retained-advisory key (`update_retained_advisory_test.go:33`).

A future refactor of the update flow that breaks llm.yaml preservation would
fail no llm.yaml-named test. The contract the card asks for exists but is
emergent, not pinned.

### §A.4 The policy question (card t239, Class C heart)

The card offers: (a) exclude llm.yaml from managed redistribution (first deploy
only), (b) detect user modification at update time → preserve + inform, or
(c) hybrid. §D below decides (c) — and records why (a) and (b)-as-new-machinery
lose.

## §B Goal

1. Decide and record the preservation policy for `.moai/config/sections/llm.yaml`:
   keep the existing template-deploy + backup + 3-way-merge + restore pipeline
   (which already implements "first deploy ships defaults; updates preserve user
   divergence and deliver new keys"), and pin it with named regression coverage.
2. Give llm.yaml the same named end-to-end protection its sibling section files
   already carry — through BOTH update paths (template-sync and clean-reinstall),
   against the REAL embedded template bytes.
3. Keep first-update behavior calm for untouched installs and preserve the
   no-provenance-sidecar property (conscious divergence from the t230 hook
   mechanism, rationale in §D).

## §C Scope

**In scope**: llm.yaml preservation contract tests (template-sync path;
clean-reinstall path parity), the design-decision record, and — only if the new
contracts expose a live break — the minimal repair the failing contract names.
No merge-engine semantics change is planned; the engine is exercised as-is.

**Out of scope** (§F): t280, t230, first-update imperfect-base over-preservation,
workflow.yaml audit pins, runtime llm.yaml writers.

## §D Requirements (GEARS)

- **REQ-LCP-001** (Ubiquitous) — `.moai/config/sections/llm.yaml` shall remain
  template-deployed and merge-protected: the update flow shall NOT exclude it
  from managed redistribution, shall NOT relocate it out of the managed root,
  and shall NOT carry a provenance sidecar for it. The preservation mechanism is
  the existing Backup → 3-way node merge → Restore pipeline.

- **REQ-LCP-002** (event-driven) — **When** the update template-sync cycle runs
  on a project whose llm.yaml carries values diverging from the merge base, the
  restore step shall write llm.yaml with every divergent user value preserved
  and the template's comment documentation retained. Verification is at value
  granularity (parsed keys) plus comment presence — never byte equality of the
  merged file.

- **REQ-LCP-003** (event-driven) — **When** the embedded template llm.yaml
  carries a key absent from the user's llm.yaml, the update shall deliver that
  key with its template default into the user's file. A preserved file keeps
  receiving new template keys; preservation must never fossilize the file.

- **REQ-LCP-004** (state-driven) — **While** a project carries no llm.yaml, the
  update shall deploy the template default verbatim. First-deploy behavior stays
  calm for untouched installs; no preservation machinery may alter what a fresh
  project receives.

- **REQ-LCP-005** (unwanted) — The update path shall NOT introduce a provenance
  discriminator for llm.yaml (a sidecar in the style of the pre-commit tier's
  `.moai-pre-commit.sha256`), and base comparison in the 3-way merge shall
  remain the only user-modification detector. Rationale: llm.yaml has a natural
  merge base (`.template-defaults` + the template-base snapshot), so a sidecar
  would import t230's first-release false-flag constraint (every pre-existing
  install reads as "user-modified" on the first update after the sidecar
  ships) while adding nothing the base comparison does not already provide.

- **REQ-LCP-006** (Ubiquitous) — The llm.yaml preservation contract shall be
  pinned by end-to-end regression tests that drive the REAL embedded template
  `.moai/config/sections/llm.yaml` through the update cycle (Backup → Clean
  Managed Paths → Deploy Templates → Restore Settings) on both the template-sync
  and clean-reinstall paths. Synthetic-YAML unit fixtures alone do not satisfy
  this requirement; every new contract shall have its teeth demonstrated (the
  mutant obligation in acceptance.md §Verification Strategy).

### §D.1 Design decision (the Class C record)

**Chosen: (c) hybrid — keep + pin.** The existing pipeline already delivers the
hybrid semantics the card sketches: the template ships defaults on first deploy
(REQ-LCP-004), and updates preserve user divergence while delivering new keys
(REQ-LCP-002/003). t239's deliverable is the contract pin + the decision record,
not new preservation machinery.

**(a) exclude llm.yaml from managed redistribution — REJECTED.** Three reasons:

1. Mechanically inert as stated: `CleanMoaiManagedPaths` wipes the whole
   `.moai/config` directory (`deploy.go:187-204`), so "don't ship the template"
   protects nothing — exclusion would require a protection list consulted before
   `RemoveAll` (a new special-case mechanism diverging from the section-file
   merge machinery) or relocating the file out of the managed root.
2. Relocation is a user-facing cost with zero preservation gain: the loader
   (`internal/config` llm section), the runtime writers (`moai glm`/`cg` via
   `internal/settings`), and every existing project's file layout would need to
   migrate, and the gitignore contract would need a transition window.
3. Exclusion severs new-key delivery forever: a never-redeployed file never
   gains template keys (REQ-LCP-003 becomes structurally impossible). t225's
   `workflow.audit` block is a concrete instance of templates gaining keys users
   must receive.

**(b) fresh detect-and-inform machinery — REJECTED as duplication.** Detection
(base comparison in the 3-way merge) and informing (retained-key advisories +
the merge-fallback ledger, REQ-UN-007/008/010) already exist. Rebuilding them
for one file adds a second mechanism where the ladder question is "does a
helper already exist here" — it does.

## §E Constraints

- **Test-only default posture**: the planned deliverable is regression coverage
  + the decision record. If M1/M2 expose a live break (RED that is not the
  mutant), the repair is scoped to exactly what the failing contract names —
  no engine redesign rides this SPEC.
- **Real-template fixture obligation**: the e2e fixtures read the embedded
  template llm.yaml (`template.EmbeddedTemplates()`), not a hand-copied
  miniature — the comment-dense real bytes are the point (issue #1243 class).
- **Value-granularity assertions**: assert parsed key values, comment presence,
  and (where load-bearing) key membership — never byte equality of merged
  output (comment/ordering-preserving merges make byte snapshots brittle).
- **Scoped verification**: run targeted packages and `-run` filters; never the
  local full suite (lane load discipline — 2026-08-15 incident). The
  `internal/cli` package is heavy: use `-run TestUpdateLLMYAML…` filters.
- **Mutant teeth**: each new contract demonstrates failure under a targeted
  mutation (e.g. bypassing the restore call) before being counted — a green
  contract that has never seen red proves nothing (verification-completeness).
- **Template neutrality**: no new template content ships; the existing template
  llm.yaml is unchanged (its content-neutrality status is untouched).

## §F Out of Scope

### Out of Scope — t280 (lessons-inbox.jsonl)
- Same defect-family suspicion, different surface; this plan did NOT
  investigate t280's wipe path (lessons-inbox.jsonl lives outside
  `.moai/config`, whose wipe is the only mechanism verified here). Cross-
  referenced, not absorbed; any t239+t280 merge decision belongs to the
  lead/operator.

### Out of Scope — t230 (pre-commit hook provenance)
- The hook provenance sidecar, `classifyHook`, and hook backup-before-overwrite
  are untouched; cited as prior art for the sidecar trade-off analysis only.
  The t255 precondition verdict (`.moai/reports/t255/precondition-verdict.md`,
  2026-09-02) establishes the discriminator has shipped in NO release yet
  (first release would be v3.1.4, pending) — which is exactly why llm.yaml
  must not adopt the mechanism now (REQ-LCP-005): adopting it here would put
  llm.yaml's first-discrimination event on the same release-cliff timeline
  the hook tier is deliberately spacing out.

### Out of Scope — first-update imperfect-base over-preservation
- When no template-base snapshot exists (first update under the preservation
  machinery), the base falls back to CURRENT embedded defaults, so an
  untouched key whose default changed between releases reads as "user-modified"
  and keeps its old value — conservative over-preservation, never data loss.
  Base provisioning is owned by SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001.

### Out of Scope — workflow.yaml audit pins
- `workflow.audit.{codex,glm}` values in this dev repo are TRACKED (git
  protects them) and the template ships empty defaults by design;
  SPEC-V3R6-AUDIT-MODEL-PIN-001 owns that schema. This SPEC neither touches
  nor tests workflow.yaml.

### Out of Scope — runtime llm.yaml writers
- `moai glm`/`moai cg` settings expansion and the web console's typed edit path
  mutate llm.yaml outside the update flow; `internal/settings` (yamlpatch)
  owns those surfaces.

### Out of Scope — merge engine semantics
- The 3-way node merge is exercised as-is. No precedence, retained-key, or
  comment-preservation behavior changes unless a new contract fails and names
  the defect.

## §G HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-02 | Draft authored (plan phase, Factory Mode card t239, lane-15). Card premise partially refuted by tree evidence: audit pins live in workflow.yaml (SPEC-V3R6-AUDIT-MODEL-PIN-001 MF1), and the backup→3-way-merge→restore pipeline already preserves llm.yaml values (verified: scoped backup-package tests green at HEAD `b7462203a`). Verified gap: llm.yaml is the only major section file without named end-to-end preservation coverage. Design decision (c) keep+pin recorded with (a)/(b) rejection rationale. |

## §H Cross-References

- `SPEC-UPDATE-YAML-PRESERVE-001` — the node-tree merge engine this SPEC pins
  (issue #1243: comments/order/quoting preservation).
- `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001` — merge-base provisioning; owns the
  first-update imperfect-base residual this SPEC explicitly excludes.
- `SPEC-V3R6-AUDIT-MODEL-PIN-001` — provenance SPEC for the audit pins; its
  MF1 relocation (llm.yaml → workflow.yaml) is the evidence that corrected the
  card's premise.
- `internal/cli/update/deploy/deploy.go` — `CleanMoaiManagedPaths`, the
  whole-directory wipe that makes per-file exclusion inert (§A.1).
- `internal/cli/update/backup/{backup,restore,merge,node_merge}.go` — the
  preservation pipeline under contract.
- `internal/cli/update_clean_install_config_preserve_test.go` — the sibling
  e2e pattern (user/language/design.yaml) llm.yaml joins.
- `internal/cli/hook_install_precommit.go` — the t230 provenance sidecar
  mechanism consciously NOT replicated here (REQ-LCP-005).
