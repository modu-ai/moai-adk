---
id: SPEC-AGENTS-IGNORE-POLICY-001
title: ".agents .gitignore policy ruling — template default-allow vs dev-repo whitelist"
version: "0.1.0"
status: draft
created: 2026-09-14
updated: 2026-09-14
author: manager-spec
priority: P1
phase: "v3.2.1"
module: ".gitignore"
lifecycle: spec-anchored
tags: "gitignore, agents-skills, policy-ruling, dogfooding, template-parity"
tier: M
related_specs: [SPEC-GITIGNORE-ROOT-GUARD-001, SPEC-CODEX-DUAL-AGENTS-001]
---

# SPEC-AGENTS-IGNORE-POLICY-001 — .agents .gitignore Policy Ruling

## §A History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-14 | manager-spec | Initial plan-phase artifacts (Tier M, card t738). Ruling authored; application deferred. |

## §B Overview

This repository carries two mutually exclusive `.agents` ignore policies:

1. **Dev-repo policy** (root `.gitignore`, lines 133-170) — a default-deny whitelist chain.
2. **Template policy** (`internal/template/templates/.gitignore`, lines 193-226) — a default-allow policy that ignores only the regenerated deploy-time skill mirror.

The dev repo ignores `.agents/*` more aggressively than the policy it ships to its users. The consequence is two-fold: (a) the user-side `.agents` state cannot be reproduced inside this repo, which muddied card t498's mirror-absence investigation; and (b) nobody ever ruled which of the two policies is correct.

**The product of this SPEC is a POLICY RULING, not a code repair.** The SPEC records the ruling with evidence, presents both remediation options with blast radius for operator adjudication, and defers all application until cards t498/t510 close.

## §C Problem Statement

Two failure axes, per the lead-issued card t738 (t498-derived):

- **Axis 1 — self-dogfooding broken.** A user project created by `moai init` receives the template `.gitignore`, whose `.agents` policy lets user-authored entries under `.agents/` stay tracked. A maintainer working in this dev repo cannot reproduce that state: the root `.gitignore` whitelist chain hides everything except 16 published SKILL.md files, so any locally created `.agents` artifact is invisible to `git status` and to every grep-based observation.
- **Axis 2 — unruled divergence.** One of the two policies is wrong, and no decision record says which. The divergence is invisible in normal operation because git silently hides the affected files.

Background input (card-supplied, t701, landed): the root `.agents/*` pattern contains a non-trailing `/` and is therefore anchored to the repository root; it never covered `internal/template/templates/.agents/`, and the removed 171-174 re-include block was a no-op. This SPEC did not re-derive the t701 archaeology (flagged as background input, not independently re-measured); the anchoring semantics are consistent with the rules as read (§D).

## §D The Two Policies — Evidence

All line citations and probe outputs below were measured in this worktree (`.claude/worktrees/t738`, HEAD `99e02ac52`, branch `worktree-t738`) on 2026-09-14. The probe command is `git check-ignore -v <paths>` (read-only; creates no files). A path absent from the probe output is NOT ignored.

### §D.1 Dev-repo policy — default-deny whitelist (root `.gitignore`)

| Line(s) | Rule | Effect |
|---------|------|--------|
| 133 | `.agents/*` | Ignore everything under `.agents/` |
| 134 | `!.agents/skills/` | Re-include the skills directory itself |
| 135 | `.agents/skills/*` | Re-ignore every entry under `.agents/skills/` |
| 138-153 | `!.agents/skills/moai-<cmd>/` ×16 | Re-include the 16 published command-skill directories |
| 154 | `.agents/skills/*/*` | Re-ignore their contents |
| 155-170 | `!.agents/skills/moai-<cmd>/SKILL.md` ×16 | Re-include only the 16 SKILL.md files |

Observed probe output (verbatim):

```
.gitignore:154:.agents/skills/*/*	.agents/skills/moai-clean/manifest.json
.gitignore:135:.agents/skills/*	.agents/skills/moai-workflow-tdd/SKILL.md
.gitignore:135:.agents/skills/*	.agents/skills/my-custom/SKILL.md
.gitignore:154:.agents/skills/*/*	.agents/skills/moai-clean/extra/SKILL.md
.gitignore:133:.agents/*	.agents/notes.md
```

Net semantics: only the 16 `.agents/skills/moai-<command>/SKILL.md` files are trackable. A sidecar file inside a published directory (`manifest.json`, `extra/SKILL.md`), a user-authored skill (`my-custom/`), and any root-level `.agents` file (`notes.md`) are all silently invisible. The deploy-time mirror (`moai-workflow-*`) is also ignored — correctly.

### §D.2 Template policy — default-allow, ignore build products only (`internal/template/templates/.gitignore`)

| Line(s) | Rule | Effect |
|---------|------|--------|
| 193-201 | comment block | States the intent: the mirror is a build product regenerated on every deploy |
| 202-203 | comment | "Only the generated entries are ignored. The `.agents/` root itself is NOT: entries you create there, and source files placed there later, stay tracked." |
| 204 | `.agents/skills/moai*` | Ignore only the regenerated deploy-time mirror entries |
| 205-210 | comment | The publisher's collision guard keeps published `moai-<command>` names disjoint from mirror names |
| 211-226 | `!.agents/skills/moai-<cmd>/` ×16 | Re-include the 16 published command-skill directories — full contents, no content re-exclusion |

Observed probe output (verbatim, run inside `internal/template/templates/`):

```
internal/template/templates/.gitignore:204:.agents/skills/moai*	.agents/skills/moai-workflow-tdd/SKILL.md
```

Net semantics: the full contents of the 16 published directories are trackable, user-authored entries (`my-custom/`, `notes.md`) are trackable, and only the regenerated mirror (`moai-workflow-*`) is ignored.

### §D.3 Semantic difference table

| Path class | Dev-repo policy (L133-170) | Template policy (L193-226) |
|---|---|---|
| 16 published `SKILL.md` files | tracked | tracked |
| Other files inside the 16 published dirs | **ignored** (L154) | tracked |
| User-authored entries (`my-custom/`, `notes.md`) | **ignored** (L133/135) | tracked |
| Deploy-time mirror (`moai-workflow-*`) | ignored | ignored |
| `.agents/` root itself | ignored-by-default territory | tracked territory |

The two policies agree on exactly one class: the deploy-time mirror is a build product and is ignored. They disagree on everything else.

### §D.4 RULING

**The template policy is the correct policy. The dev-repo root `.gitignore` (lines 133-170) is the divergent policy.**

Reasons:

1. **The template policy is a documented decision** (lines 193-203): it draws the tracked/ignored boundary on the only distinction that has a principled basis — regenerated build product vs. everything else. The dev-repo policy has no decision record; it draws the boundary at "SKILL.md files that were already tracked when the rules were written", which is an artifact of history, not a policy.
2. **The template policy makes dogfooding possible.** A maintainer can reproduce the user-side `.agents` state (axis 1). The dev-repo policy cannot, by construction.
3. **The dev-repo policy creates silent-invisibility traps** in exactly the shape this repository's recurring defect family warns about: a file that exists on disk but is absent from `git status` reads as evidence of absence. The t498 mirror-absence investigation was muddied by precisely this.
4. **Adopting the template policy in the dev repo is dev-side only.** The template already encodes the correct policy, so the ruling requires NO template-side change and carries zero user impact (blast radius and the rejected alternatives are laid out in `plan.md` §F for operator adjudication).

The 16 published SKILL.md files remain tracked under either policy; the ruling changes nothing that is currently tracked.

## §E Requirements (GEARS)

- REQ-POL-001 (Ubiquitous): The SPEC record shall name the template `.agents` ignore policy (`internal/template/templates/.gitignore` lines 193-226) as the canonical policy and the dev-repo root `.gitignore` lines 133-170 whitelist chain as the divergent policy, with exact line citations for both.
- REQ-POL-002 (Event): When the operator adjudicates this ruling, the SPEC record shall capture the operator decision (option chosen) and its rationale in `progress.md`.
- REQ-POL-003 (Event): When the application window opens (after cards t498 and t510 close), the dev-repo root `.gitignore` shall be aligned to the template semantics: ignore only `.agents/skills/moai*` regenerated mirror entries, re-include the 16 published `.agents/skills/moai-<command>/` directories with full contents, and leave the `.agents/` root tracked.
- REQ-POL-004 (State-driven): While SPEC-AGENTS-IGNORE-POLICY-001 remains in `draft` or `in-progress`, no `.gitignore` rule — root or template — shall be modified by this SPEC. The application step in REQ-POL-003 is a separately-gated later act, owned by the follow-up card that runs after t498/t510 close.
- REQ-POL-005 (Event-detected): When the operator selects a remediation option that requires a template-side change (all-user impact), the follow-up card shall carry an explicit migration plan for existing user projects before any template rule is redeployed.

## §F Constraints

- This SPEC changes no `.gitignore` rule. Its deliverables are documentation (the ruling, the option analysis, the operator-gate question) only.
- Cards t498/t510 use the current absence of local `.agents` mirror state as an observation target; application timing is theirs, not this SPEC's.
- Verified present-state measurement (2026-09-14, HEAD `99e02ac52`): `.agents/skills/` contains exactly the 16 tracked published directories; `git status --porcelain .agents/` is empty; no ignored-but-present `.agents` content exists. Aligning the dev repo to the template policy would therefore surface no untracked noise in this repository today.

## §G Out of Scope

### Out of Scope — application of the ruling
- Any edit to the root `.gitignore` or `internal/template/templates/.gitignore`. This SPEC delivers the ruling and the application plan; the follow-up card applies it after t498/t510 close.

### Out of Scope — unrelated queue-path cards
- Cards t704/t705/t706 (queue-path concerns) are unrelated to this policy question and are not addressed here.

### Out of Scope — implementation detail
- Go code changes, publisher/collision-guard logic changes, and `moai update` deploy behavior. The ruling is a gitignore-policy decision; tooling changes, if the ruling's application ever demands them, belong to the follow-up card.

## §H Cross-References

- `plan.md` §F — remediation options with blast radius (operator-gate material)
- `acceptance.md` — machine-verifiable acceptance criteria
- Related: SPEC-GITIGNORE-ROOT-GUARD-001 (root .gitignore guard), SPEC-CODEX-DUAL-AGENTS-001 (the `.codex/` re-include analog, which chose the "template-distributed artifacts ARE committed" direction)
- Cards: t738 (this), t498/t510 (application timing owners), t701 (anchoring background)
