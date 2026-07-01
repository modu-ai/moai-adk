---
id: SPEC-OWASP-CHECKLIST-GAP-001
title: "moai-ref-owasp-checklist: Generalized Trust-Boundary Principles + Stale Agent Reference Fix"
version: "0.2.1"
status: in-progress
created: 2026-07-01
updated: 2026-07-01
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai-ref-owasp-checklist"
lifecycle: spec-anchored
tier: "S"
era: V3R6
tags: "security, owasp, skill, reference, agent-catalog, checklist"
---

# SPEC-OWASP-CHECKLIST-GAP-001 — moai-ref-owasp-checklist: Generalized Trust-Boundary Principles + Stale Agent Reference Fix

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-07-01 | Initial draft created by manager-spec. Plan-phase artifacts only; no implementation. |
| 0.2.0 | 2026-07-01 | plan-auditor iteration 1 FAIL remediation (D1-D8): acceptance criteria rewritten in GEARS notation and folded inline as §3 (Tier S 2-file convention — `acceptance.md` removed per D7); AC-OCG-007 scope-containment command fixed to use `git diff --name-only` (D2); AC-OCG-002 split into AND-semantics sub-checks (D3); explicit `Traces to:` citations added to every AC plus 2 new formal requirements REQ-OCG-009/010 (D4); AC-OCG-006 (unchanged item number) given a concrete baseline-commit verification command (D5); REQ-OCG-005 relabeled from Ubiquitous to Unwanted behavior (D6); AC-OCG-008 (unchanged item number) supplemented with a post-commit check to close the pre-commit-only blind spot (D8). No implementation performed — plan-phase artifacts only. |
| 0.2.1 | 2026-07-01 | plan-auditor iteration 2 PASS-WITH-DEBT remediation (D9-D10): AC-OCG-007's scope-containment command anchored to the plan-phase baseline SHA `366e701af60bb789714efbe6068cac59788fb6bb` (matching the AC-OCG-006/008 pattern) so the check no longer vacuously passes once the change is staged or committed — reclassified from Event-driven to State-driven GEARS pattern to match its "While comparing against baseline" phrasing (D9); corrected the inaccurate "(renumbered §3 item)" parenthetical in the 0.2.0 HISTORY entry above — AC-OCG-006/008 retained their original item numbers; nothing was renumbered (D10). plan.md §E point 4 updated to mirror the same anchored command. No implementation performed — plan-phase artifacts only. |

## §A. Background and Problem

### A.1 Research provenance

This SPEC originates from a comparative audit of the external repository **https://github.com/AITutor3/vibecoding-security-skill** (a Codex/Antigravity skill scoped to Next.js + Vercel + Supabase security auditing) against moai-adk's existing defensive-security assets. The audit was performed via 4 parallel read-only exploration passes and verified directly against this repository's files with `Grep`/`Read` (not inferred).

### A.2 Finding 1 — moai-adk already has broad coverage; the external repo is a strict subset

moai-adk-go already ships broad defensive application-security coverage:

- The `/moai security` workflow (OWASP A01-A10 full audit)
- `moai-ref-owasp-checklist` (OWASP API Top 10, auth patterns, security headers, input validation, P0-P3 severity)
- `moai-ref-secops`, `moai-ref-llm-security`, `moai-ref-supply-chain` (defensive cybersecurity reference skills, per SPEC-V3R6-SEC-SKILL-INTEGRATION-001)

The external repo's content — verified during the audit — is a strict subset of this coverage in nearly every area it addresses (auth, headers, input validation, secret handling).

### A.3 Finding 2 — bundling the external repo as a new skill is architecturally forbidden

Two independent rules forbid vendoring the external repo as a bundled single-stack skill:

- `.claude/rules/moai/development/skill-authoring.md` § "Language Guidance Lives in Rules, Not Skills" — cross-framework/language guidance must go through `moai-ref-*` skills, never a stack-specific composite skill.
- `CLAUDE.local.md` §15 — anything under `internal/template/templates/` must treat all 16 supported languages equally; a skill scoped to one framework (Next.js), one host (Vercel), and one BaaS (Supabase) has zero precedent among the repository's 32 existing skills.

**This SPEC therefore does NOT create any new skill, and does NOT import any Next.js/Vercel/Supabase-specific file, script, or reference content.**

### A.4 Finding 3 — 5 generalizable security principles are absent from moai-ref-owasp-checklist

After stripping the external repo's Supabase/Vercel/Next.js-specific naming, 5 principles remain that are genuinely framework-neutral, generalizable, and confirmed ABSENT from `moai-ref-owasp-checklist` (verified via repository-wide `Grep`):

1. **Cached/client-supplied session identity is not proof of current server-side identity.** Generalizes Supabase's `getSession()` (locally decoded, cache-only) vs. `getUser()` (round-trips to the auth server) distinction. Applies to any framework that caches or locally decodes a session/JWT value.
2. **Edge/gateway/middleware-layer auth checks are a UX convenience, not a security boundary.** Generalizes the external repo's Middleware/Proxy warning. Applies to Express middleware, Next.js middleware, Django decorators, Spring filters, Cloudflare Workers, etc. Only implicitly present today in this skill's existing A5 "Broken Function Level Authorization" row — never stated as its own explicit principle.
3. **Scheduled/cron-triggered HTTP endpoints are still public URLs and need their own authentication.** Generalizes Vercel's `CRON_SECRET` pattern. Applies equally to AWS Lambda scheduled events, GCP Cloud Scheduler, k8s CronJob-triggered HTTP calls.
4. **Production builds must not expose source maps or equivalent debug artifacts.** Generalizes Next.js/Vercel's `productionBrowserSourceMaps` setting. Applies to any bundler/build tool that can emit debug artifacts alongside a production build.
5. **Webhook receivers must verify a signature/HMAC header before trusting the request body.** Generalizes Supabase's `x-supabase-signature` check. Applies equally to Stripe, GitHub, Slack, and any other webhook provider.

### A.5 Finding 4 — unrelated stale agent-reference defect

Independently of the gap-closure above, `Grep` confirms `.claude/skills/moai-ref-owasp-checklist/SKILL.md` (lines 6, 12, 34-35) names `expert-security` and `expert-backend` as "Target Agents". Both agents are **archived** per `CLAUDE.md` §4 (the 2026-05-25 8-agent catalog consolidation) and MUST NOT be referenced as live targets.

## §B. Goal and Solution

Update the single skill file `.claude/skills/moai-ref-owasp-checklist/SKILL.md` — which is template-managed, with the template source at `internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md` as the single source of truth, currently byte-identical to the local deployed copy — to:

1. Add the 5 generalized, framework/vendor-neutral security principles identified in §A.4, in whatever placement and exact wording best fits the skill's existing structure (implementer judgment; see plan.md for a recommended placement and drafted wording).
2. Correct the stale `expert-security` / `expert-backend` "Target Agents" references and the matching frontmatter `description` / `when_to_use` text so they reflect the current 8-agent retained catalog.

No other skill, agent, workflow, or rule file is modified by this SPEC.

## §C. Requirements (GEARS)

- **REQ-OCG-001** (Ubiquitous): The `moai-ref-owasp-checklist` skill shall document 5 additional framework-neutral security principles covering: (a) server-side identity re-verification against cached/client-supplied session state, (b) edge/middleware-layer authorization independence, (c) scheduled-job endpoint authentication, (d) production build source-map/debug-artifact exposure prevention, and (e) webhook signature verification.

- **REQ-OCG-002** (Unwanted behavior): The `moai-ref-owasp-checklist` skill shall not express any of the 5 new principles by naming a specific vendor, framework, or hosting platform — the skill body shall not contain the literal (case-insensitive) strings "Supabase", "Next.js", "Vercel" (or close variants such as "Nextjs", "next.js").

- **REQ-OCG-003** (Ubiquitous): The `moai-ref-owasp-checklist` skill's "Target Agents" section shall reference only agents and workflows present in the current 8-agent retained catalog (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `sync-auditor`, `builder-harness`, `Explore`) or documented replacement patterns (`/moai security` workflow, per-spawn `Agent(general-purpose)` per `archived-agent-rejection.md` §C).

- **REQ-OCG-004** (When — event-detected): When the "Target Agents" section is corrected, the `moai-ref-owasp-checklist` skill shall document BOTH (a) that backend-implementation context flows through `manager-develop` AND (b) that security-audit invocation flows through the `/moai security` workflow or a per-spawn `Agent(general-purpose)` security specialist per `archived-agent-rejection.md` §C. Both (a) and (b) are required — documenting only one does not satisfy this requirement.

- **REQ-OCG-005** (Unwanted behavior): The `moai-ref-owasp-checklist` skill shall not contain the literal strings `expert-security` or `expert-backend` anywhere in its frontmatter `description` or `when_to_use` fields, nor in its body "Target Agents" section.

- **REQ-OCG-006** (Where — capability gate on template-source SSOT): Where the skill file is template-managed with a single source of truth at `internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md`, the change process shall edit the template source first and then propagate byte-identical content to the local deployed copy at `.claude/skills/moai-ref-owasp-checklist/SKILL.md`, so both files remain byte-identical after the change (verified via `diff`).

- **REQ-OCG-007** (When — event-detected): When the `SKILL.md` content changes, the `internal/template/catalog.yaml` `hash` field for the `moai-ref-owasp-checklist` entry shall be regenerated so the recorded hash matches the new file content (verified via `gen-catalog-hashes.go --dry-run`).

- **REQ-OCG-008** (Unwanted behavior): The `moai-ref-owasp-checklist` skill shall not lose, delete, or semantically alter any pre-existing section, table row, or evolvable block (`Common Rationalizations`, `Red Flags`, `Verification`) as a side effect of this change — all edits shall be additive (new principles) or narrowly targeted (the stale agent-reference lines only, matched against the plan-phase baseline commit — see §3 AC-OCG-006).

- **REQ-OCG-009** (Unwanted behavior): The change process shall not modify any skill file other than `moai-ref-owasp-checklist`'s `SKILL.md` (its 2 copies) and `internal/template/catalog.yaml` — in particular, `moai-ref-api-patterns`, `moai-foundation-cc`, `moai-workflow-spec/references/reference.md`, and `moai-meta-harness/references/*.md` (which also contain stale `expert-security`/`expert-backend` references, per §D "Out of Scope — Other Skills") shall remain untouched by this SPEC.

- **REQ-OCG-010** (Unwanted behavior): The change process shall not create any new skill file, skill directory, or `SKILL.md` under `.claude/skills/` or `internal/template/templates/.claude/skills/`.

## §3. Acceptance Criteria (GEARS notation, inline per Tier S convention)

> Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier, Tier S SPECs use a 2-file artifact set (spec.md + plan.md) with AC inline in spec.md §3 — no separate `acceptance.md`. Each AC below states a GEARS-pattern criterion, a verification command, and a `Traces to:` REQ citation, per the plan-auditor MP-2/D4 remediation.

**Plan-phase baseline commit** (referenced by AC-OCG-006 and AC-OCG-008 for pre/post-commit git-diff comparisons): `366e701af60bb789714efbe6068cac59788fb6bb` (recorded at the time this revision was authored, before any M1-M5 implementation edit begins).

### AC-OCG-001 (Event-driven) — Vendor-neutral generalized principles added

**When** the 5 generalized principles (session-identity re-verification, edge/middleware auth independence, scheduled-job endpoint auth, production source-map/debug-artifact exposure prevention, webhook signature verification) are added to the `moai-ref-owasp-checklist` skill template source, **the skill body shall** contain zero case-insensitive matches of `supabase|next\.?js|vercel`.

Traces to: REQ-OCG-001, REQ-OCG-002

```bash
grep -c -i -E 'supabase|next\.?js|vercel' internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md
# Expected: 0
```

### AC-OCG-002 (Event-driven, compound AND) — Stale "Target Agents" section corrected

**When** the "Target Agents" section is corrected, **the skill shall** satisfy all three of the following (AND, not OR — per REQ-OCG-004's dual requirement):

1. The file shall contain zero matches of `expert-security|expert-backend`.
2. The file shall contain at least one match of `manager-develop`.
3. The file shall contain at least one match of `/moai security|Agent\(general-purpose`.

Traces to: REQ-OCG-003, REQ-OCG-004, REQ-OCG-005

```bash
# Sub-check 1 — archived-agent-name absence
grep -c -E 'expert-security|expert-backend' internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md
# Expected: 0

# Sub-check 2 — manager-develop presence (REQ-OCG-004(a))
grep -c -E 'manager-develop' internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md
# Expected: >= 1

# Sub-check 3 — security-invocation-flow presence (REQ-OCG-004(b))
grep -c -E '/moai security|Agent\(general-purpose' internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md
# Expected: >= 1

# PASS requires all 3 sub-checks to hold simultaneously (AND semantics).
```

### AC-OCG-003 (Unwanted behavior) — Frontmatter description/when_to_use corrected

**The frontmatter block shall not** contain `expert-security` or `expert-backend` in its `description` or `when_to_use` fields.

Traces to: REQ-OCG-005

```bash
awk '/^---$/{c++} c<=1' internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md \
  | grep -c -E 'expert-security|expert-backend'
# Expected: 0
```

### AC-OCG-004 (Event-driven) — Template source and local deployed copy byte-identical

**When** both the template source and the local deployed copy are edited, **the two files shall** be byte-identical.

Traces to: REQ-OCG-006

```bash
diff .claude/skills/moai-ref-owasp-checklist/SKILL.md \
     internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md
# Expected: (no output, exit 0)
```

### AC-OCG-005 (Event-driven) — catalog.yaml hash regenerated

**When** the skill's content changes, **the recorded `hash:` field shall** match the freshly computed hash (idempotent on a second dry-run).

Traces to: REQ-OCG-007

```bash
go run internal/template/scripts/gen-catalog-hashes.go --dry-run --entry moai-ref-owasp-checklist
# Expected output hash == the hash committed in internal/template/catalog.yaml for this entry
```

### AC-OCG-006 (State-driven) — No pre-existing content deleted or semantically altered

**While** comparing the template source against the plan-phase baseline commit `366e701af60bb789714efbe6068cac59788fb6bb`, **every deleted line in the diff shall** contain `expert-security` or `expert-backend` (i.e., no other pre-existing content — table rows, headings, evolvable blocks — was deleted).

Traces to: REQ-OCG-008

```bash
git diff 366e701af60bb789714efbe6068cac59788fb6bb -- internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md \
  | grep -E '^-[^-]' \
  | grep -v -E 'expert-security|expert-backend'
# Expected: no output (every deleted line matches the stale-reference terms;
# any other surviving output indicates an out-of-scope deletion)
```

### AC-OCG-007 (State-driven) — Scope containment (other skills untouched)

**While** comparing the working tree against the plan-phase baseline commit `366e701af60bb789714efbe6068cac59788fb6bb` (the same baseline AC-OCG-006 and AC-OCG-008 use), **the set of modified file paths (excluding this SPEC's own `.moai/specs/SPEC-OWASP-CHECKLIST-GAP-001/` artifacts) shall be** exactly the 3-path allowlist: `.claude/skills/moai-ref-owasp-checklist/SKILL.md`, `internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md`, `internal/template/catalog.yaml`.

Traces to: REQ-OCG-009

```bash
# --name-only has no trailing summary line (unlike --stat), so no leak is possible (D2 fix).
# Anchored to the baseline SHA (not a bare `git diff`) so the check remains valid whether the
# working tree is unstaged, staged, or already committed — a bare unanchored `git diff --name-only`
# compares only worktree-vs-index and vacuously passes once a violation is staged/committed (D9 fix).
git diff --name-only 366e701af60bb789714efbe6068cac59788fb6bb -- . ':!.moai/specs/SPEC-OWASP-CHECKLIST-GAP-001/' \
  | grep -vE '^(\.claude/skills/moai-ref-owasp-checklist/SKILL\.md|internal/template/templates/\.claude/skills/moai-ref-owasp-checklist/SKILL\.md|internal/template/catalog\.yaml)$'
# Expected: no output
```

### AC-OCG-008 (Event-driven, pre- and post-commit) — No new skill created

**When** implementation completes (checked both before and after the final commit, to avoid a post-commit blind spot), **no new skill directory shall** exist under `.claude/skills/` or `internal/template/templates/.claude/skills/` that did not exist at the plan-phase baseline commit `366e701af60bb789714efbe6068cac59788fb6bb`.

Traces to: REQ-OCG-010

```bash
# Pre-commit check — uncommitted new files/directories
git status --porcelain -- '.claude/skills/' 'internal/template/templates/.claude/skills/' | grep -E '^\?\?|^A '
# Expected: no output

# Post-commit check — committed new files/directories since the plan-phase baseline
git diff --name-only --diff-filter=A 366e701af60bb789714efbe6068cac59788fb6bb -- '.claude/skills/' 'internal/template/templates/.claude/skills/'
# Expected: no output
```

### §3.1 Quality Gate Criteria

- [ ] All 8 AC (AC-OCG-001 through AC-OCG-008) verified PASS with actual command output (not asserted).
- [ ] Every AC's `Traces to:` REQ-OCG-XXX citation resolves to an existing requirement in §C.
- [ ] Frontmatter YAML remains valid/parseable (no broken `---` delimiters, no unescaped special characters).
- [ ] No markdown table is left malformed (column count consistent per row) in either edited section.
- [ ] The 3 evolvable blocks (`rationalizations`, `red-flags`, `verification`) remain intact and unmodified.
- [ ] Commit message (future run-phase) follows Conventional Commits and references `SPEC-OWASP-CHECKLIST-GAP-001`.

### §3.2 Definition of Done

1. Both `SKILL.md` copies (template source + local deployed) contain the 5 new generalized principles and are byte-identical (AC-OCG-001, AC-OCG-004, AC-OCG-006).
2. All `expert-security`/`expert-backend` references in this skill (frontmatter + body) are corrected to reference the current 8-agent retained catalog / documented replacement patterns, with AND-semantics verified (AC-OCG-002, AC-OCG-003).
3. `internal/template/catalog.yaml`'s `moai-ref-owasp-checklist` entry hash is regenerated and verified stable (AC-OCG-005).
4. No other skill, agent, workflow, or rule file is modified, and no new skill is created (AC-OCG-007, AC-OCG-008).
5. `SPEC-OWASP-CHECKLIST-GAP-001` frontmatter `status` transitions `draft → in-progress` (on first run-phase commit, owned by `manager-develop` per the Status Transition Ownership Matrix) and eventually `implemented → completed` (sync-phase, owned by `manager-docs`) — both transitions deferred; NOT performed by this plan-phase artifact set.

### §3.3 Forward-Looking Checks (deferred to a future SPEC, not required for this SPEC's closure)

- Cleanup of the remaining `expert-security`/`expert-backend` references in `moai-ref-api-patterns`, `moai-foundation-cc`, `moai-workflow-spec/references/reference.md`, and `moai-meta-harness` (out of scope here; tracked per REQ-OCG-009 and spec.md §D).

## §D. Out of Scope

### Out of Scope — New Skill Creation
- Creating any new skill file, skill directory, or `SKILL.md` under `.claude/skills/` or `internal/template/templates/.claude/skills/`.

### Out of Scope — Vendor-Specific Content
- Importing, porting, or adapting any Next.js-specific, Vercel-specific, or Supabase-specific reference file, code sample, or configuration snippet from the external repository or elsewhere.

### Out of Scope — Other Skills
- Modifying `moai-ref-api-patterns`, `moai-foundation-cc`, `moai-workflow-spec/references/reference.md`, `moai-meta-harness`, or any other skill file that also contains stale `expert-security` / `expert-backend` references. Those are a separate, pre-existing cleanup surface and are explicitly deferred to a future SPEC.

### Out of Scope — Agent / Workflow / Rule Changes
- Modifying `/moai security` workflow logic, `.claude/rules/moai/workflow/archived-agent-rejection.md`, or any `.claude/agents/**/*.md` agent definition file.

### Out of Scope — Implementation
- Writing any code (Go, shell, etc.). This SPEC's deliverable is skill-content (markdown + YAML frontmatter) only.

## §E. Cross-References

- `.claude/rules/moai/development/skill-authoring.md` § Language Guidance Lives in Rules, Not Skills
- `CLAUDE.local.md` §15 (템플릿 언어 중립성)
- `.claude/rules/moai/workflow/archived-agent-rejection.md` §C (migration table for archived agents)
- `CLAUDE.md` §4 (8-agent retained catalog; archived agents list)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (Tier S 2-file artifact-set convention applied in this revision)
- `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/` (sibling defensive-security reference-skill SPEC; independent re-authorship precedent)
- External research source (provenance only, not a dependency): https://github.com/AITutor3/vibecoding-security-skill
