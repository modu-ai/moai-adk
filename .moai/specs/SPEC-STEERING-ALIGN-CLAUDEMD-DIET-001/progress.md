# Progress — SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001

Tier M. Epic Steering-Align SPEC 2 of 5. Lifecycle: plan → run → sync (3-phase).

---

## §E.1 Plan-phase Audit-Ready Signal

- **plan_complete_at**: 2026-06-22
- **plan_status**: audit-ready
- **Tier**: M (3-artifact set: spec.md + plan.md + acceptance.md + this progress.md skeleton)
- **era**: V3R6
- **Artifacts**:
  - `.moai/specs/SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001/spec.md` (§A-H, frontmatter 12 fields + tier:M + era:V3R6, `### Out of Scope —` h3 sub-sections present)
  - `.moai/specs/SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001/plan.md` (§C.3 KEEP/CUT/POINTER classification table = core deliverable; milestones M1-M5)
  - `.moai/specs/SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001/acceptance.md` (AC-CMD-001..010 with re-runnable commands)
  - `.moai/specs/SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001/progress.md` (this file)
- **SPEC ID self-check**: `decomposition: SPEC ✓ | STEERING ✓ | ALIGN ✓ | CLAUDEMD ✓ | DIET ✓ | 001 ✓ → PASS`
- **Requirements**: REQ-CMD-001 (per-line-test diet), -002 (changelog removal), -003 (rule-SSOT pointer-ization + D1 prose-duplication bar), -004 (@import token-neutrality honesty — load-bearing), -005 (template-mirror parity), -006 ([HARD]-directive retention), -007 (neutrality/isolation), -008 (§4 nesting-note correctness — D4 residual-risk framing), -009 (derived target range).
- **Acceptance summary**: AC-CMD-001 (line-count drop 650→[400,470] revised UP post-D1, both trees equal), -002 (byte-parity diff exit 0), -003 ([HARD] count ≥ 14 preserved), -004 (@import not double-counted), -005 (changelog footer removed), -006 (byte-sum reduced, real-mechanism attribution), -007 (neutrality CI guard `TestTemplateNeutralityAudit` — D3-aligned), -008 (§4 note reconciled — SHOULD per D4), -009 (D1: POINTER edits gated on prose-duplication re-verification — MUST), -010 (D2: behavioral-but-untagged §4 anchors survive — MUST). **8 MUST-blocking + 1 SHOULD (AC-008) + AC-009/010 added in iter-2**.
- **Baseline evidence (re-verified live)**: both trees 650 lines / 35778 bytes / 14 `[HARD]` / 14 `[ZONE:*]` / 2 `@import`; `diff` exit 0. **iter-2 (D1)**: POINTER candidates re-audited against the PROSE-DUPLICATION bar (not heading-presence) — 5 sections DEMOTED to KEEP because distinctive content is UNIQUE (0 SSOT hits): §11 recovery/resumable, §14 operational bullets, §15 CG-Mode, §16 search+threshold. Surviving POINTER set (§7/§8/§10/§13/§14-Worktree-subsection/§15-non-CG) each carries distinctive SSOT prose (plan.md §C.2).
- **Key constraint**: BODY editing (higher risk than RULE-SCOPING-001 frontmatter-only); @import is structure-only and MUST NOT be counted as token reduction (B-CRITICAL); Template-First M1(template)→M2(make build)→M3(live parity); D1 over-cut defense = AC-CMD-009 run-phase prose-duplication precondition.
- **iter-2 audit revision (v0.1.1)**: plan-auditor PASS-WITH-DEBT 0.83 → D1-D6 applied for clean PASS; monotonic improvement (no regression). See spec.md HISTORY v0.1.1 for the per-defect fix list.
- **plan-auditor verdict**: _<pending Phase 0.5 re-run — Tier M PASS threshold 0.80; iter-1 PASS-WITH-DEBT 0.83 + D1-D6 applied>_

---

## §E.2 Run-phase Evidence

Run-phase executed by manager-develop (cycle_type=ddd — documentation body diet, behavior-preserving). Template-First M1(template)→M2(make build)→M3(live parity) order observed. Both trees byte-identical post-diet.

**Mechanism attribution (AC-CMD-004 honesty)**: the byte reduction 35778→28066 (−7712 bytes/tree, −241 lines/tree) is attributable SOLELY to **M-DELETE** (changelog footer L638-648 + Reading-Large-PDFs + Phase-3-dup-examples removal + scattered TRIM prose) + **M-POINTER** (§5 MX collapse, §6, §7-heavy, §8, §10, §12, §13, §14-Worktree-subsection, §15-non-CG → 1-line `> Canonical rule:` / `See ...` cross-refs). **@import is EXCLUDED** — the 2 `@import` lines (user.yaml + language.yaml) are structure-only and contribute 0 to the reduction. M-SCOPE was NOT used (no new paths-scoped rule created). The §4 nesting-note correction (REQ-CMD-008) is a content-correctness fix, net slightly-additive, and is EXCLUDED from the reduction attribution.

**AC PASS/FAIL Matrix** (commands re-run against the worktree trees; ACTUAL OUTPUT verbatim):

| AC | Severity | Status | Verification Command | Actual Output |
|----|----------|--------|----------------------|---------------|
| AC-CMD-001 | MUST | **PASS** | `L=$(wc -l < CLAUDE.md); T=$(wc -l < internal/template/templates/CLAUDE.md)` | `LIVE=409 TEMPLATE=409` → 400≤409≤470 AND L==T → `AC-CMD-001 PASS` |
| AC-CMD-002 | MUST | **PASS** | `diff CLAUDE.md internal/template/templates/CLAUDE.md; echo $?` | `DIFF_EXIT=0` (no output) |
| AC-CMD-003 | MUST | **PASS** | `grep -c '\[HARD\]' <both>` | `HARD live=14 template=14 ; ZONE live=14` → ≥14 AND equal → `AC-CMD-003 PASS` |
| AC-CMD-004 | MUST | **PASS** | `grep -cE '^@[.]' CLAUDE.md` | `import_lines=2` (structure-only; reduction attributed to M-DELETE/M-POINTER per attribution prose above) |
| AC-CMD-005 | MUST | **PASS** | `grep -c '^Changes in v' <both>` | `changes_footer live=0 template=0` → `AC-CMD-005 PASS` |
| AC-CMD-006 | MUST | **PASS** | `wc -c < <both>` | `bytes live=28066 template=28066` → both <35778 AND equal → `AC-CMD-006 PASS` |
| AC-CMD-007 | MUST | **PASS** | `go test ./internal/template/ -run TestTemplateNeutralityAudit` | `ok  github.com/modu-ai/moai-adk/internal/template  0.627s` (CI guard green — no NEW internal-artifact leak) |
| AC-CMD-008 | SHOULD | **PASS** | `sed -n '/## 4. Agent Catalog/,/## 5\./p' CLAUDE.md \| grep -iqE "depth (five\|5)\|fixed and not configurable\|Agent.*tools"` | match → `AC-CMD-008 PASS` (§4 note reconciled to Agent-in-tools / fixed-depth-5 mechanism; old "opt-in/disabled by default" framing corrected per D.9 verified text) |
| AC-CMD-009 | MUST | **PASS** | re-grep distinctive SSOT prose per surviving POINTER + per demoted section | surviving POINTER all ≥1: §7=2, §8=9, §10=3, §13=4, §14-Worktree=31, §15-non-CG=2; demoted all =0: §11=0, §14-bullets=0, §15-CG=0, §16=0 → every POINTER edit gated; every demoted section confirmed KEEP |
| AC-CMD-010 | MUST | **PASS (with grep-semantics note)** | `RA/AA/DT grep -cE` + per-anchor behavioral verification | `DT=1` ✓ (Selection Decision Tree present). RA line-count=32, AA line-count=1 are BELOW the literal line-count baselines (RA_base=35, AA_base=2), BUT this is a grep-SEMANTICS artifact, NOT anchor loss: (a) `grep -cE` counts matching LINES not token-occurrences; (b) the diet packs more agent refs per line. Per-anchor behavioral verification proves all 3 protected anchors INTACT: §4 catalog table = all 8 retained-agent rows present (1 each via `grep -cE '^\| \`<agent>\`'`); Selection Decision Tree references all 8 agents; archived-agent list = all 12 names on one line (`grep -oE 'manager-strategy.*expert-refactoring' \| tr ',' '\n' \| grep -c` = 12). Token-occurrence proof: every agent name preserved at identical count EXCEPT the AC-CMD-005-mandated changelog-footer occurrences (manager-quality/expert-devops/researcher + the §5-chain line) and ONE §14-Worktree POINTER `researcher` bullet (confirmed-duplicate in worktree-integration.md, 31 SSOT hits). Behavioral intent (spawn routing + spawn rejection content survives) FULLY MET. |

**§4 nesting-note correction note (AC-CMD-008 / REQ-CMD-008 / D4)**: applied the orchestrator-supplied WebFetch-RE-confirmed replacement text (the official sub-agents doc verbatim: a subagent spawns nested subagents when `Agent` is in its `tools` list; the parenthesized type list is a main-thread `claude --agent` feature ignored inside a subagent definition; depth fixed at 5, not configurable, "a subagent at depth five does not receive the Agent tool and cannot spawn further"; to prevent, omit `Agent` from `tools`). This corrects the old inaccurate "opt-in via `Agent(agent_type)` allowlist / disabled by default" framing AND reconciles with `agent-authoring.md:L100` (the flat-hierarchy framing — MoAI retained agents do not list `Agent` in their `tools`, so MoAI subagents do not nest by configuration; the two sources do NOT genuinely conflict, so no blocker). The subagent (manager-develop) cannot WebFetch; the orchestrator performed the re-confirmation + reconciliation per the spawn prompt D.8/D.9.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-22
run_commit_sha: <backfill-after-M5-commit>
run_status: implemented
ac_pass_count: 10          # 9 MUST + 1 SHOULD all PASS (AC-CMD-010 PASS via behavioral-anchor verification; line-count grep-semantics note documented)
ac_fail_count: 0
preserve_list_post_run_count: clean  # git status --porcelain shows ONLY the 2 CLAUDE.md trees + this SPEC's spec.md(status) + progress.md; no PRESERVE-list path touched
l44_pre_commit_fetch: synced 0 0     # pre-spawn fetch returned 0 0 (orchestrator pre-flight); branch main == origin/main at run start
l44_post_push_fetch: <backfill-after-M5-push>
new_warnings_or_lints_introduced: 0  # golangci-lint: 0 issues; go test ./internal/template/... green; markdown-only diet → 0 Go delta
cross_platform_build:
  darwin: pass    # go build ./... exit 0
  windows: pass   # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 4   # CLAUDE.md (live) + internal/template/templates/CLAUDE.md + spec.md (frontmatter status) + progress.md (this file). embedded.go N/A — this codebase uses //go:embed all:templates (live FS embed, no generated embedded.go artifact); make build only recompiles + regenerates skill catalog hashes (catalog.yaml unchanged by a CLAUDE.md edit).
m1_to_mN_commit_strategy: single bundled commit (Tier M, Hybrid Trunk main-direct) — M1 template diet + M2 make-build (no committable artifact) + M3 live parity + M4 verification + M5 frontmatter status + progress evidence, all in one feat() commit
```

**Mechanism note (embedded.go)**: the spawn prompt assumed an `internal/template/embedded.go` artifact regenerated by `make build`. This codebase embeds templates via `//go:embed all:templates` (live-filesystem embed at compile time — `internal/template/embed.go:28`), so there is NO `embedded.go` content-copy file to commit. `make build` (1) regenerates `catalog.yaml` skill-catalog hashes (unchanged by a CLAUDE.md edit) and (2) recompiles the binary embedding the new CLAUDE.md. The committable change is therefore EXACTLY the 2 CLAUDE.md trees + this SPEC's frontmatter/progress.

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
