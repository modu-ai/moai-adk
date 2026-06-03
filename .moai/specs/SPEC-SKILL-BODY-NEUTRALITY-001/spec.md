---
id: SPEC-SKILL-BODY-NEUTRALITY-001
title: "Skill-Body Neutrality — purge moai-adk local-dev traces from deployed skills + extend neutrality CI guard"
version: "0.1.1"
status: draft
created: 2026-06-04
updated: 2026-06-04
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template/templates/.claude/skills"
lifecycle: spec-anchored
tags: "template-system, neutrality, skills, ci-guard, distribution"
tier: M
related_specs: [SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001, SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]
---

# SPEC-SKILL-BODY-NEUTRALITY-001 — Skill-Body Neutrality

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-04 | manager-spec | Initial plan-phase draft. Part A (purge skill-body internal traces) + Part B (extend neutrality CI guard to cover skill-body leak classes). |
| 0.1.1 | 2026-06-04 | manager-spec | plan-auditor PASS-WITH-DEBT 0.84 defect patch (D1-D5): D1 add `moai/references/reference.md:244` to §B.4; D2 new §B.5 4-locale-annotation inventory + REQ-SBN-019 + AC-SBN-019; D5 REQ-SBN-013 C7 pkg-restriction HARD + AC-SBN-020; D4 REQ-token partition-guard extension to AC-SBN-018; D3 binary positive-greps for AC-SBN-004/009. 18-REQ↔18-AC bijection extended to 20↔20. |

## §A Goal

Deployed skill bodies under `internal/template/templates/.claude/skills/` ship to end-user projects across all 16 supported languages. Orchestrator-verified grep sweep (2026-06-04, 221 deployed skill `.md` files) found extensive **moai-adk local-development traces** that are meaningless or misleading in end-user projects: references to moai-adk's own CI test files, its own 5-platform release build, its own internal Go implementation paths, and its own internal SPEC/REQ history.

This SPEC has two parts:

- **Part A — Purge internal traces from deployed skill bodies** (meaning-preserving generic-ization, NOT blind deletion). A user reading a deployed skill must understand the mechanism described without being told to look at a Go file or CI test that does not exist in their project.
- **Part B — Extend the neutrality CI guard to cover the skill-body leak classes** so the leaks cannot silently re-accumulate.

This complements (does not duplicate) two prior SPECs:
- `SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001` (implemented) — sanitized **rules-file** content for path/version/date/author classes. It did NOT cover skill-body Go-impl-path references or the CI-test self-reference paradox.
- `SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001` — owns the `internal_content_leak_test.go` strict-tier date/sha classes. Part B of this SPEC extends the **same** guard with skill-body-specific leak classes, partitioned to avoid dual-allow-list drift.

### §A.1 Why this matters (motivation for generic-ization over deletion)

An end-user who installs MoAI-ADK receives these skill bodies verbatim. When a skill says "CI guards in `internal/template/agentless_audit_test.go` enforce the literal `MODE_UNKNOWN` sentinel," the user has no `internal/template/agentless_audit_test.go` — the reference is dead and confusing. But the **information the user needs** ("`MODE_UNKNOWN` is a stable error key; do not remove it") is real and must be preserved. Blind deletion would strip the user-facing meaning; generic-ization keeps it.

## §B Background — Orchestrator-Verified Leak Inventory

All counts below are grep-verified against `internal/template/templates/.claude/skills/` at the plan-phase baseline.

### §B.1 CLASS 1 — moai-adk CI test path references (6 files)

Skill bodies name moai-adk's own CI test file `internal/template/agentless_audit_test.go`:

| File | Line | Leak |
|------|------|------|
| `moai/workflows/loop.md` | 45 | `test TestLoopAliasCrossReference in internal/template/agentless_audit_test.go` |
| `moai/workflows/plan.md` | 157 | `CI guards in internal/template/agentless_audit_test.go enforce ... MODE_PIPELINE_ONLY_UTILITY ... REQ-WF003-016 ↔ REQ-WF004-014 ... SPEC-V3R2-WF-004` |
| `moai/workflows/sync.md` | 116 | same pattern as plan.md:157 |
| `moai/workflows/design.md` | 27 | `CI guards in internal/template/agentless_audit_test.go enforce ... MODE_UNKNOWN` |
| `moai/workflows/run.md` | 113 | same + `REQ-WF003-010` |
| `moai/workflows/run/context-loading.md` | 83 | `CI guards in internal/template/agentless_audit_test.go enforce each literal sentinel` |

### §B.2 CLASS 2 — moai-adk release build / internal Go impl paths

**Release build (`moai/workflows/sync/delivery.md`):**
- Lines 139-143: `GOOS=linux/darwin/windows × amd64/arm64` cross-compile of `./cmd/moai/` — moai-adk's OWN 5-platform release build, not a user-project build.
- Lines 135, 174: `golangci-lint ... @v2.1.6` pinned-version install of moai-adk's own toolchain.

**Internal Go impl paths (point to real moai-adk source files):**

| File | Line | Real internal path |
|------|------|--------------------|
| `moai/workflows/sync/quality-gates-context.md` | 185 | `internal/hook/dbsync/db_schema_sync.go`, `internal/cli/hook.go` |
| `moai/workflows/plan/spec-assembly.md` | 78 | `internal/spec/lint.go` |
| `moai/workflows/harness.md` | 30, 35 | `internal/cli/harness.go` |
| `moai/team/plan.md` | 162 | `internal/spec/lint.go` |
| `moai-workflow-worktree/SKILL.md` | 235 | `internal/cli/worktree/team_launch.go` |
| `moai-workflow-spec/SKILL.md` | 205 | `internal/spec/lint.go` |
| `moai-workflow-spec/references/reference.md` | 11 | `internal/spec/lint.go` |
| `moai-workflow-design/SKILL.md` | 216 | `internal/design/dtcg/frozen_guard_test.go` |
| `moai-workflow-ci-loop/SKILL.md` | 99, 150 | `internal/ciwatch/handoff.go`, `internal/cli/pr/watch.go` |

### §B.3 CLASS 3 — internal SPEC IDs + REQ tokens (37 files with V3R-family SPEC IDs)

Real internal SPEC IDs present in skill prose (NOT format placeholders): `SPEC-V3R2-WF-001/003/004/005`, `SPEC-V3R2-HRN-003`, `SPEC-V3R3-HARNESS-001`, `SPEC-V3R3-PROJECT-HARNESS-001`, `SPEC-V3R3-HARNESS-LEARNING-001`, `SPEC-V3R3-CI-AUTONOMY-001`, `SPEC-V3R3-RETIRED-DDD-001`, `SPEC-V3R4-HARNESS-001/003/005/006`, `SPEC-V3R4-WORKFLOW-SPLIT-001`, `SPEC-V3R5-LATE-BRANCH-001`, `SPEC-V3R5-WORKFLOW-LEAN-001`, `CONST-V3R5-001/030`, plus `SPEC-WF-AUDIT-GATE-001`, `SPEC-MX-001`.

REQ token families present: `REQ-WF003`, `REQ-WF004`, `REQ-BRAIN`, `REQ-SKILL`, `REQ-ROUTE`, `REQ-PH`, `REQ-LB`, `REQ-DPL`, `REQ-BRIEF`, `REQ-WAG`, `REQ-FALLBACK`, `REQ-HL`, `REQ-HCW`, `REQ-DETECT`, `REQ-WTL`, `REQ-SLQG`, `REQ-HLF`, `REQ-HARNESS`, `REQ-CONST`. Worst-affected: `moai/workflows/design.md`, `brain.md`, `harness.md`, `plan/spec-assembly.md`.

### §B.4 CLASS 4 — dev-only self-reference + maintainer doctrine

| File | Line | Leak |
|------|------|------|
| `moai/SKILL.md` | 76, 238, 244-245 | `release-update (dev-only) ... NOT distributed to user projects (entry: .claude/commands/97-release-update.md)` — self-contradictory text shipped to users |
| `moai-foundation-core/modules/commands-reference.md` | 21, 264, 272, 329 | `/moai:99-release` (99-* prefix is dev-only-reserved) |
| `moai-foundation-core/modules/INDEX.md` | 143 | `/moai:99-release (Deployment)` |
| `moai/references/reference.md` | 244 | `Note: /moai:99-release is a separate local-only command, not part of the /moai skill.` — 99-* dev-only-reserved self-reference (6th `99-release` hit; required so AC-SBN-008 `grep -rn '99-release'` reaches 0) |
| `moai-meta-harness/SKILL.md` | 168 | Korean maintainer doctrine: `Doctrine-code drift (2026-05-26 ~ catch-up SPEC 완료 전) ... maintainer doctrine` — internal date + catch-up-SPEC reference |

`grep -rn '99-release' internal/template/templates/.claude/skills/` returns exactly **6 hits** at the plan-phase baseline: `commands-reference.md` (×4: lines 21, 264, 272, 329), `INDEX.md` (line 143), and `moai/references/reference.md` (line 244). All 6 are CLASS 4 dev-only self-references and ALL must be purged for AC-SBN-008's "0 matches".

### §B.5 CLASS 5 — docs-site URL "4-locale" maintainer annotation

The public docs-site GEARS URL `adk.mo.ai.kr` is legitimate and KEPT (EXCL-SBN-005). But several skill bodies attach a maintainer-facing "4-locale (en/ko/ja/zh)" annotation to that URL — a maintainer-internal concern (the docs-site's translation coverage) that is meaningless to an end user consulting the public URL. The annotation appears in two surface forms: `4-locale: en / ko / ja / zh` and `4-locale (en / ko / ja / zh)`.

| File | Line | Annotation surface form |
|------|------|-------------------------|
| `moai-foundation-core/modules/spec-ears-format.md` | 11 | `... (4-locale: en / ko / ja / zh)` |
| `moai-workflow-spec/SKILL.md` | 65 | `... — 4-locale (en / ko / ja / zh)` |
| `moai-workflow-spec/SKILL.md` | 146 | `... (4-locale: en / ko / ja / zh) ...` |
| `moai-workflow-spec/references/reference.md` | 27 | `... (4-locale) ...` |

Note: `moai/SKILL.md:76,240` also carry a 4-locale annotation but tied to the dev-only `release-update` entry — those are removed wholesale by CLASS 4 / REQ-SBN-008, so they are not separately enumerated here (removing the dev-only entry removes the annotation with it). The KEPT object is the `adk.mo.ai.kr` URL itself; the DROPPED object is the `4-locale...` annotation attached to it.

## §C Requirements (GEARS)

### §C.1 Part A — Purge skill-body internal traces

**REQ-SBN-001** (Ubiquitous): The deployed skill bodies shall not name moai-adk's internal CI test file `internal/template/agentless_audit_test.go`.

**REQ-SBN-002** (Ubiquitous): The deployed skill bodies shall retain every literal mode sentinel keyword value (`MODE_UNKNOWN`, `MODE_PIPELINE_ONLY_UTILITY`, `MODE_FLAG_IGNORED_FOR_UTILITY`, `MODE_TEAM_UNAVAILABLE`, `AGENTLESS_CONTROL_FLOW_VIOLATION`) so the existing sentinel-presence verification continues to pass.

**REQ-SBN-003** (Ubiquitous): The deployed skill bodies shall describe each sentinel's purpose generically (as a stable error key) without referencing the internal test that enforces it or the internal REQ token that owns it.

**REQ-SBN-004** (Ubiquitous): The `sync/delivery.md` skill body shall describe the release/CI build step as a generic multi-platform Go build pattern without hardcoding moai-adk's own `./cmd/moai/` target or its pinned `golangci-lint` version.

**REQ-SBN-005** (Ubiquitous): The deployed skill bodies shall not reference real moai-adk internal Go implementation paths (`internal/<pkg>/<file>.go` forms that map to actual moai-adk source); each such reference shall be generic-ized to a mechanism description (e.g., `internal/spec/lint.go FrontmatterSchemaRule` → "the SPEC frontmatter lint rule").

**REQ-SBN-006** (Ubiquitous): The deployed skill bodies shall not contain real internal SPEC IDs of the `SPEC-V3R[0-9]-*` / `CONST-V3R[0-9]-*` families, nor the named real internal SPEC IDs `SPEC-WF-AUDIT-GATE-001` and `SPEC-MX-001`; each shall be replaced by a generic description of the policy or mechanism it denoted (e.g., `REQ-LB-009 (SPEC-V3R5-LATE-BRANCH-001)` → "the late-branch opt-in policy").

**REQ-SBN-007** (Ubiquitous): The deployed skill bodies shall not contain internal REQ/AC token references in prose (`REQ-<PREFIX>-NNN`, `REQ-WF<NNN>-NNN`); each shall be generic-ized to the requirement's plain-language intent.

**REQ-SBN-008** (Ubiquitous): The deployed skill bodies shall not contain dev-only self-references to the `97-*` / `99-*` reserved command prefixes, nor self-contradictory "NOT distributed to user projects" text; the `release-update` / `99-release` entries shall be removed from user-facing command surfaces or generic-ized.

**REQ-SBN-009** (Ubiquitous): The `moai-meta-harness/SKILL.md` skill body shall not contain the maintainer-only doctrine note that carries an internal date and a "catch-up SPEC" reference; it shall be replaced by a generic statement of the namespace-separation policy.

**REQ-SBN-010** (Where capability gate): Where a skill-body SPEC ID, internal path, or REQ token is a **format-example placeholder** (`SPEC-AUTH-001`, `SPEC-{ID}`, `SPEC-XXX`, `SPEC-{DOMAIN}-{NUMBER}`, `SPEC-001`/`SPEC-002`/`SPEC-NNN` series, domain-example IDs such as `SPEC-PAY-001` / `SPEC-API-001` / `SPEC-FASTAPI-001` / `SPEC-DASH-001` / `SPEC-UI-001`, illustrative example Go paths in a fictional code-review or file-list example), the purge shall preserve it verbatim as a legitimate pedagogical illustration.

**REQ-SBN-011** (When event-driven): When Part A edits a skill file under `internal/template/templates/.claude/skills/`, the corresponding local mirror under `.claude/skills/` shall be synced to the same content (Template-First Rule, CLAUDE.local.md §2).

**REQ-SBN-019** (Ubiquitous): The deployed skill bodies shall not attach a maintainer-facing "4-locale" translation-coverage annotation (`4-locale: en / ko / ja / zh` or `4-locale (en / ko / ja / zh)`) to the public docs-site URL `adk.mo.ai.kr`; the annotation shall be dropped at every §B.5 site while the `adk.mo.ai.kr` URL itself is preserved.

### §C.2 Part B — Extend the neutrality CI guard

**REQ-SBN-012** (Ubiquitous): The neutrality guard test `internal/template/internal_content_leak_test.go` shall enforce a leak class that matches the internal CI-test-file reference `internal/template/agentless_audit_test.go` in any scanned skill body.

**REQ-SBN-013** (Ubiquitous): The neutrality guard shall enforce a Go-impl-path leak class (`C7`) that matches real moai-adk internal Go implementation paths in scanned skill bodies. [HARD] The `C7` regex MUST be **package-restricted** to the same real moai-adk top-level package set used by AC-SBN-005 — `internal/(spec|cli|hook|ciwatch|design)/...\.go` — and MUST NOT use the unrestricted `internal/.*\.go` form. The package restriction is required because the unrestricted form would match the EXCL-SBN-003 illustrative example paths (`internal/auth/login.go`, `internal/api/handler.go`, `internal/core/handler.go`), making the M6 GREEN state unreachable. As belt-and-suspenders, the 3 illustrative paths MUST ALSO be registered in the pedagogical allowlist (REQ-SBN-015).

**REQ-SBN-014** (Ubiquitous): The neutrality guard's internal-SPEC-ID leak class shall match the `SPEC-V3R[0-9]-*` family (currently the C1 class matches only `SPEC-(V3R6|AGENCY|WORKTREE)-*` and misses the V3R2-V3R5 families found in skill bodies).

**REQ-SBN-015** (Ubiquitous): The neutrality guard shall register every legitimate format-example placeholder and illustrative example path (per REQ-SBN-010) in the pedagogical allowlist so they do not produce false-positive findings.

**REQ-SBN-016** (When event-driven): When the neutrality guard runs after Part A is complete, it shall report zero findings for the skill-body leak classes (CLASS 1-4) excluding allow-listed placeholders.

**REQ-SBN-017** (When event-driven, unwanted-behavior form): When a future edit reintroduces a CLASS 1-4 leak into a skill body, the neutrality guard shall fail (the RED → GREEN guard is the recurrence backstop).

**REQ-SBN-018** (Where capability gate): Where the date-class (`S1-internal-date`) and short-sha-class (`S2-short-sha-sentence-final`) leak detection already lives in the strict tier owned by `SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001`, Part B shall NOT add overlapping date/sha classes — it adds only the skill-body-specific classes (CI-test path, Go-impl path, broadened SPEC-ID) to avoid dual-allow-list drift.

## §D Exclusions (What NOT to Build)

- **EXCL-SBN-001** — Date-class and commit-SHA-class enforcement. Owned by `SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001` strict tier (`S1-internal-date`, `S2-short-sha-sentence-final`). This SPEC does not add or duplicate those classes.
- **EXCL-SBN-002** — Rules-file (`.claude/rules/`), agent-file (`.claude/agents/`), and command-file (`.claude/commands/`) neutrality. Out of scope; this SPEC's surface is skill bodies (`.claude/skills/`) only. (Rules-file neutrality is `SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001`.)
- **EXCL-SBN-003** — Format-example placeholders and illustrative example IDs/paths. KEEP verbatim: `SPEC-AUTH-001`, `SPEC-{ID}`, `SPEC-XXX`, `SPEC-{DOMAIN}-{NUMBER}`, `SPEC-001`/`SPEC-002`/`SPEC-NNN` series, domain-example IDs (`SPEC-PAY-001`, `SPEC-API-001`, `SPEC-FASTAPI-001`, `SPEC-DASH-001`, `SPEC-UI-001`, etc.), and illustrative example Go paths inside fictional code-review / file-list examples (`internal/auth/login.go` and `internal/api/handler.go` in `pr-review-multi-agent.md`; `internal/core/handler.go` in `mx.md`'s mixed-language modified-files list). The 3 illustrative Go paths are protected by **two** independent mechanisms per REQ-SBN-013: (a) the `C7` regex is package-restricted to `internal/(spec|cli|hook|ciwatch|design)` and so does not match `internal/auth`, `internal/api`, or `internal/core`; AND (b) all 3 are additionally registered in the pedagogical allowlist (belt-and-suspenders, REQ-SBN-015).
- **EXCL-SBN-004** — Neutral multi-language tool lists. KEEP: `golangci-lint` when it appears as one item in a 16-language toolchain list (Go is 1 of 16); the user-facing install command `go install github.com/modu-ai/moai-adk/cmd/moai@latest`; Vercel in frontend/e2e/project/plugin domain-knowledge contexts.
- **EXCL-SBN-005** — The public docs-site GEARS URL `adk.mo.ai.kr`. KEEP the URL (users may legitimately consult it); however the maintainer-facing "4-locale (en/ko/ja/zh)" annotation attached to that URL is a maintainer concern and is dropped where it appears in a skill body (decided in plan §F).
- **EXCL-SBN-006** — Implementation of the Part B guard's exact regex literals, the precise generic-ized replacement prose for each file, and the pedagogical-allowlist entry list. These are run-phase (HOW) decisions deferred to plan.md milestones and the run phase; this spec.md defines only WHAT and WHY.
- **EXCL-SBN-007** — Sentinel keyword removal. The mode sentinels (`MODE_UNKNOWN`, etc.) are MoAI-ADK system identifiers (allowed content per §25). They are NOT leaks and MUST remain present (REQ-SBN-002).
- **EXCL-SBN-008** — The `internal_content_leak_test.go` walk-scope mechanism. The guard already walks the entire `templatesRoot` tree (including skills) and scans `.md` files; the gap is **pattern coverage**, not walk scope. This SPEC does not rewrite the walk; it adds leak classes + allowlist entries.

## §E Constraints

- **Template-First Rule** (CLAUDE.local.md §2): `internal/template/templates/` is the SSOT. `//go:embed all:templates` means `make build` picks up template edits directly (no `embedded.go` to regenerate). The local `.claude/skills/` mirror must be synced to match (REQ-SBN-011).
- **Meaning preservation** (CLAUDE.local.md §25 allowed content): generic-ize, do not blindly delete. Allowed: generic prose, mechanism descriptions, public-doc references, permanent-rule references, MoAI-ADK system identifiers (subcommand names, sentinel keyword values). Forbidden: internal SPEC IDs, REQ/AC tokens, internal Go impl paths, internal CI test file names, internal dates, commit SHAs.
- **Specific-path commit discipline**: the working tree has UNRELATED uncommitted edits (`language.yaml`, `statusline.yaml`, `user.yaml`, a deleted `progress.md`, untracked `design/` + `harness-delivery-strategy.md`). The plan commit MUST stage only `.moai/specs/SPEC-SKILL-BODY-NEUTRALITY-001/` and carry the `Authored-By-Agent: manager-spec` trailer.
- **GEARS notation**: requirements use GEARS (Ubiquitous / When / While / Where); no legacy `IF/THEN`.

## §F Success Criteria (summary; full AC in acceptance.md)

- Part A: every CLASS 1-4 leak in §B is generic-ized with meaning preserved; allow-listed placeholders untouched; local mirror synced.
- Part B: the neutrality guard (RED before Part A, GREEN after) fails on the current leaks and passes once they are purged; the skill-body leak classes (CI-test path, Go-impl path, broadened SPEC-ID) are added without overlapping the ISOLATION-001 date/sha classes; legitimate placeholders are allow-listed.
- Build + tests: `go test ./internal/template/...` passes (including the existing `agentless_audit_test.go` sentinel-presence tests and the extended leak test); `go build ./...` succeeds.
