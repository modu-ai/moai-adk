---
id: SPEC-V3R6-TOOL-POLICY-SSOT-001
title: "Tool/Permission Policy SSOT — Design (schema, codegen approach, drift-prevention)"
version: "0.1.0"
status: in-progress
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/config"
lifecycle: spec-anchored
tags: "policy, tools, ssot, harness, codex"
tier: L
era: V3R6
---

# Design — SPEC-V3R6-TOOL-POLICY-SSOT-001

> Design decisions: (§B) codegen approach selected + alternatives considered; (§C) the YAML schema; (§D) the SSOT invariant + §24.5 drift-prevention mechanism; (§E) migration path; (§F) Template-First distribution decision.

---

## §A. Design Questions

- DQ-1: What codegen approach (template vs codegen vs hook-time read)?
- DQ-2: What is the YAML entry schema?
- DQ-3: How does the design prevent the §24.5 drift class BY CONSTRUCTION?
- DQ-4: How do existing scattered-policy entries migrate into the YAML without behavior change?
- DQ-5: Is the YAML template-distributed or local-only (Template-First Rule C4)?

---

## §B. Codegen Approach (DQ-1) — Selected + Alternatives

### §B.1 SELECTED — `moai tool-policy build` subcommand + Go codegen with block-region replacement (raw-text on .tmpl, parse-modify-serialize on .json)

**Decision**: a Go codegen function in `internal/config/toolpolicy/codegen.go`, invoked via a NEW `moai tool-policy build` subcommand (`internal/cli/tool_policy.go`). The codegen:
1. Reads `.moai/config/sections/tool-policy.yaml`.
2. Parses the YAML into `[]PolicyEntry`.
3. Emits the `permissions` block (allow-list, deny-list, ask-list) as JSON.
4. Performs **block-region replacement** on TWO targets via DIFFERENT strategies (justified below):
   - **`.claude/settings.json` (local)** — pure JSON → **parse-modify-serialize**: parse the full file as JSON, replace the `permissions` object value, re-serialize. Preserves non-permissions regions because JSON round-trip is structure-preserving (with stable key ordering via `json.MarshalIndent` + sorted keys, satisfying AC-TPS-013 idempotency).
   - **`internal/template/templates/.claude/settings.json.tmpl` (template)** — mixed JSON + Go-template directives (`{{jsonEscape .SmartPATH}}` etc.) → **raw-text region replacement**: locate the `"permissions": {` ... matching-close-brace region as a raw TEXT slice (regex/bracket-matching, NOT JSON parse), replace ONLY that text region with the regenerated block, preserve the rest of the file verbatim. The `.tmpl` CANNOT be parsed as JSON because of the `{{...}}` directives outside the permissions block.

**Justification for the two-strategy split (D7 decision)**:
- The local `settings.json` is pure JSON; parse-modify-serialize is safest (structure-preserving, no string-surgery bugs).
- The template `settings.json.tmpl` contains Go-template directives (`{{jsonEscape .SmartPATH}}` at L379 in the `env` block, outside permissions). Parsing `.tmpl` as JSON FAILS — the `{{...}}` syntax is not valid JSON. Raw-text region replacement is the only viable approach for `.tmpl`.
- **Verified (iter-2, D7)**: the permissions BLOCK region inside `.tmpl` (L380-520) contains NO `{{...}}` directives — the template directives live ONLY in the `env` block (L379 `{{jsonEscape .SmartPATH}}`) and are outside the permissions region. Therefore raw-text region replacement on the permissions block of `.tmpl` is SAFE: the codegen never touches the `{{...}}`-containing regions, and the permissions block it writes is plain JSON (no template directives needed inside it because the permissions entries are static literals, not template-rendered).
- The codegen MUST validate post-condition that the regenerated `.tmpl` permissions block contains zero `{{` or `}}` substrings (sentinel: if a template directive accidentally lands inside the permissions block, the codegen is buggy).

**Justification**:
- **Single invocation point** — `moai tool-policy build` is the explicit regenerate command. Maintainer edits YAML, runs build, commits the regenerated settings.json + .tmpl. The edit → regenerate → commit cycle is auditable.
- **Block-region replacement** — both files are partially non-permissions; a full rewrite would clobber PATH/hooks/env (.json) or template directives (.tmpl). Region replacement is surgical on both.
- **Go-native** — the codegen lives alongside the existing `internal/config` loaders and reuses Go's YAML + JSON libraries; no shell-script dependency.

### §B.2 Alternative 1 — REJECTED: Go-template rendering (`settings.json.tmpl` with `{{range}}`)

**Approach**: make the permissions block a `{{range .PolicyEntries}}` loop in `settings.json.tmpl`, rendered at `moai init`/`moai update` time.

**Rejected because**:
- The YAML would need to be loaded into the `TemplateContext` struct, coupling the template renderer to the policy domain.
- `moai init`/`moai update` run infrequently; a maintainer who edits the YAML and wants to see the regenerated settings.json WITHOUT a full `moai update` would have no path.
- The template engine is for file scaffolding, not policy regeneration. Overloading it conflates two concerns.

### §B.3 Alternative 2 — REJECTED: Hook-time YAML read (no codegen)

**Approach**: a `PreToolUse` hook that reads the YAML at runtime and returns allow/deny/ask per tool invocation. No generated settings.json permissions block.

**Rejected because**:
- Claude Code's native permission layer consumes the settings.json `permissions` block; bypassing it for a hook-based layer creates a SECOND enforcement path (the §24.5 drift class in a different guise — two enforcement surfaces).
- Hook latency on every tool call (YAML parse per invocation) is unacceptable.
- Loses Claude Code's built-in permission UX (the allow/deny prompt).

### §B.4 Alternative 3 — DEFERRED: full settings.json regeneration from a single template

**Approach**: the ENTIRE settings.json (PATH, hooks, env, permissions) is generated from a single template + YAML at `moai update` time.

**Deferred because**: out of scope (§X.6). This SPEC generates ONLY the permissions block; full template-engine replacement is a separate refactor.

---

## §C. YAML Schema (DQ-2)

### §C.1 Top-level structure

```yaml
# .moai/config/sections/tool-policy.yaml
# SSOT for tool/permission policy. Codegen target: .claude/settings.json permissions block.
# Drift class prevented: harness-namespace-doctrine.md §24.5 (doctrine/enforcement divergence).
# Pattern origin: book2 ch4.3-4.4 (execpolicy), appendix B.3 (independently-evaluable invariant).
# Query surface: moai constitution list (reused; this file does NOT define a parallel query system).

metadata:
  version: "1.0.0"
  generated_into:
    - ".claude/settings.json#permissions"
    - "internal/template/templates/.claude/settings.json.tmpl#permissions"
  cross_refs:
    - ".moai/docs/harness-namespace-doctrine.md#§24.5"
    - ".claude/rules/moai/core/zone-registry.md"
    - "book2 ch4.3-4.4, appendix B.3"

entries:
  - tool: "Bash"
    args_pattern: "git push --force*"
    risk_tier: irreversible
    decision: deny
    owner_agent: orchestrator
    audit: "force-push is destructive; codified as explicit deny (was implicit: absent from allow-list)"
    source: "derived (book2 ch8.4 highest-risk-first)"
  # ... (full seed in M1)
```

### §C.2 Entry schema (6 fields + 2 metadata)

| Field | Type | Required | Notes |
|---|---|---|---|
| `tool` | string | YES | Tool name (`Bash`, `Write`, `WebSearch`, `mcp__context7__*`) |
| `args_pattern` | string | YES | Glob or regex (`git push:*`, `*`, empty for tool-level) |
| `risk_tier` | enum | YES | `read \| write \| irreversible` |
| `decision` | enum | YES | `allow \| deny \| ask` |
| `owner_agent` | string | YES | Agent role that owns this rule (`orchestrator`, `manager-develop`, etc.) |
| `audit` | string | YES | Human-readable rationale (book1 ch08 "rationale" + "impact") |
| `source` | string | NO | Where the entry was extracted from (research.md §C file:line) |
| `env_gate` | object | NO | Optional environment condition (e.g., `{base_url_contains: "api.z.ai"}` for GLM rules) |

### §C.3 Environment-gated rules (§D.3 source — GLM backend)

For rules that apply only under an environment condition (GLM backend), the `env_gate` field expresses the condition:

```yaml
- tool: "WebSearch"
  args_pattern: "*"
  risk_tier: read
  decision: deny
  owner_agent: orchestrator
  audit: "PROHIBIT under GLM backend; routes through 529-prone gateway"
  source: ".claude/rules/moai/core/glm-web-tooling.md L25-35"
  env_gate:
    base_url_contains: "api.z.ai"   # ANTHROPIC_BASE_URL contains this substring
    exception_when: "cg_leader"      # moai cg leader pane is exempt (L39)
```

**Codegen behavior for env_gated rules**: the codegen emits a hook-matcher entry (or a settings.json conditional, if Claude Code supports it) rather than a static allow/deny. design.md §C.4 decides the exact emission.

### §C.4 Codegen emission rules

| `decision` | `env_gate` | Codegen emission |
|---|---|---|
| allow | none | add to `permissions.allow` list |
| deny | none | add to `permissions.deny` list |
| ask | none | add to `permissions.ask` list (Claude Code supports this) |
| deny | present | emit hook matcher (glm-web-tooling enforcement) — static `permissions.deny` would over-block the cg-leader exception |

---

## §D. SSOT Invariant + §24.5 Drift Prevention (DQ-3)

### §D.1 The SSOT invariant

The YAML is the SINGLE source. Two surfaces are GENERATED from it:

1. **Enforcement surface** — `.claude/settings.json` `permissions` block (what Claude Code consumes at runtime).
2. **Doctrine surface** — the cross-reference comments in the YAML header + the generated audit trail.

Because both are generated, they cannot diverge. A rule edit propagates to both in one codegen pass.

### §D.2 What drift class this prevents BY CONSTRUCTION (honest scope)

§24.5 (harness-namespace-doctrine.md L59-71) occurred because the doctrine (`harness-*` in markdown) and the enforcement (`my-harness-*` in Go) were MAINTAINED INDEPENDENTLY — two edit locations, no mechanical linkage. A maintainer changed one without the other.

**This SPEC prevents the ANALOGOUS drift class ONLY on the surfaces it generates.** The surfaces this SPEC generates from the YAML are:
1. The `.claude/settings.json` `permissions` block (enforcement surface — what Claude Code consumes).
2. The YAML header comment / audit trail (audit surface).

Because BOTH are generated from the YAML, a rule edit propagates to both in one codegen pass — YAML↔settings.json drift (enforcement surface drifting from the audit surface) is structurally impossible.

**What this SPEC does NOT prevent (out of scope, §X.8):** the markdown-doctrine-vs-Go-code drift that characterized §24.5 literally. This SPEC generates NEITHER markdown doctrine files NOR Go enforcement source code. A maintainer editing the YAML does NOT propagate to `harness-namespace-doctrine.md` (markdown) nor to `internal/harness/prefix_conflict.go` (Go) — those remain independently-maintained surfaces. Narrowing rationale (verification-claim-integrity doctrine): the SPEC must promise only what it delivers. What it delivers is YAML↔settings.json single-source generation. A follow-up SPEC MAY extend the codegen to also emit markdown doctrine and/or Go code, at which point the §24.5 markdown-vs-code drift class would ALSO become preventable by construction.

**§24.5 as an ANALOGY (not a literal prevention claim):** §24.5 is cross-referenced throughout this SPEC as the canonical EXAMPLE of the drift failure mode (two independently-maintained surfaces diverging), and this SPEC's YAML↔settings.json codegen is the prevention mechanism for the ANALOGOUS class on the surfaces it touches. The cross-reference is motivational/illustrative, not a claim that this SPEC retroactively prevents the §24.5 incident itself.

### §D.3 The drift-prevention test (AC-TPS-005) — YAML↔settings.json round-trip

The test asserts the by-construction property on the surfaces this SPEC actually generates:
1. Edit YAML entry R (e.g., flip `Bash(git push:*)` from allow to ask).
2. Run codegen (`moai tool-policy build`).
3. Assert the generated `.claude/settings.json` `permissions` block reflects R — specifically, `Bash(git push:*)` moved from the `allow` array to the `ask` array (enforcement surface updated). Verify via: parse the generated settings.json permissions block, evaluate the decision for entry R, assert `decision == "ask"`.
4. Assert the YAML↔settings.json round-trip is decision-equivalent: for EVERY seeded entry, `codegenDecision(yamlEntry) == yamlEntry.decision` holds (the real deliverable — single-source generation produces a settings.json whose semantics match the YAML).
5. Assert there is NO hand-maintained enforcement file: `.claude/settings.json` `permissions` block is regenerated (not hand-edited). Verify via: the codegen region carries a generated-marker comment, and `git log` shows the permissions block only changes via codegen commits.

If step 3 OR 4 fails, the codegen is buggy. If step 5 fails (a hand-edit lands in the codegen region), the SSOT invariant is violated. Either flips the test RED.

**What step 4 does NOT test (out of scope):** it does NOT test that markdown doctrine (e.g., `harness-namespace-doctrine.md`) or Go enforcement code reflects R — those surfaces are not generated by this SPEC (§X.8).

---

## §E. Migration Path (DQ-4)

### §E.1 Migration sequence (M4)

1. **Capture baseline** — characterize the pre-migration settings.json permissions behavior as a test fixture (the current allow-list, deny-list — mostly allow, no explicit deny). This is AC-TPS-007's baseline.
2. **Seed YAML** (M1) — extract entries from the 4 sources (research.md §C) into the YAML.
3. **Codegen** (M2) — run `moai tool-policy build`, regenerate settings.json permissions block.
4. **Verify equivalence** (M4) — the compat test asserts every pre-migration decision equals the post-migration decision. Allow-list entries that were implicit (e.g., `Bash` allowed because `Bash(git push:*)` is in the list) are made explicit in the YAML.
5. **Commit** — the migration commit carries the YAML + regenerated settings.json; the compat test is the gate.

### §E.2 Handling the 4 sources during migration

| Source | Migration action |
|---|---|
| settings.json permissions (§C.1) | Bulk-extracted into YAML read-tier + write-tier entries; settings.json permissions block becomes GENERATED (no longer hand-edited) |
| status-transition-ownership.sh (§C.2) | Mirrored into YAML as write-tier entries with `audit: "mirrored; hook not yet codegen-fed"` (KI-3 — the hook stays as-is this SPEC) |
| glm-web-tooling.md PROHIBIT (§C.3) | Extracted as env_gated deny entries (§C.3); the markdown table becomes a doctrine cross-reference, the enforcement is the generated hook matcher |
| agent-common-protocol.md L162 (§C.4) | Extracted as a deny entry with `audit: "mirrored from L162; hard enforcement deferred"` (AC-TPS-012 — declared machine-readably, enforcement is follow-up) |

### §E.3 No-behavior-change guarantee (C1)

The compat test (AC-TPS-007) is the guarantee. Every entry that was `allow` in settings.json remains `allow` in the YAML → regenerated as `allow`. No entry that was absent (implicit-deny by Claude Code default) becomes `allow`. The baseline fixture is the truth.

---

## §F. Template-First Distribution Decision (DQ-5)

### §F.1 Decision: YAML in BOTH local + template

The YAML exists in:
- `.moai/config/sections/tool-policy.yaml` (local — dev source of truth, per CLAUDE.local.md §2 dev location)
- `internal/template/templates/.moai/config/sections/tool-policy.yaml` (template — distributed to consumer projects)

`moai update` syncs template → local (consistent with other config sections like `quality.yaml`, `harness.yaml`).

### §F.2 Why both (not local-only)

- **Template-First Rule (CLAUDE.local.md §2)**: any new file under `.claude/`, `.moai/`, `.agency/` MUST be added to the template source first, then synced to local. A local-only YAML violates the rule.
- **Distribution**: consumer projects (16-language template users) get the SSOT mechanism, not just moai-adk dev.
- **Neutrality (CLAUDE.local.md §15, §25)**: the template YAML carries generic policy entries (the read-tier allow-list is universal); project-specific overrides (local-only entries, e.g., a project's custom MCP tool) live in the local YAML and are preserved by `moai update` (consistent with how local `quality.yaml` overrides work).

### §F.3 Codegen distribution

The codegen OUTPUT (settings.json permissions block) is ALSO in both:
- `.claude/settings.json` (local)
- `internal/template/templates/.claude/settings.json.tmpl` (template)

`moai tool-policy build` regenerates BOTH from the respective YAML (local YAML → local settings.json; template YAML → template settings.json.tmpl). This preserves the dev-vs-template split (CLAUDE.local.md §22.1 documents that local `defaultMode: bypassPermissions` is intentional dev divergence — that field is NOT in the codegen region; only the allow/deny/ask LISTS are codegen-fed).

---

## §G. Risks + Mitigations

| Risk | Mitigation |
|---|---|
| Codegen clobbers non-permissions settings.json regions (KI-1) | Block-region replacement (§B.1); M4 round-trip test verifies region isolation |
| Background-write deny rule breaks existing workflows (KI-2) | Seed as declared-only (AC-TPS-012); hard enforcement deferred to follow-up |
| status-transition hook drift from YAML (KI-3) | Explicit `audit: mirrored` annotation; hook not codegen-fed this SPEC |
| Template-neutrality violation (forbidden content classes per §25) | Template YAML carries generic entries only; no SPEC IDs / commit SHAs / internal dev references in the template copy (CI guard `template-neutrality-check.yaml` is the safety net) |

---

## §H. Open Questions (deferred to run-phase manager-develop)

- OQ-1: RESOLVED at plan-phase iter-2 (D7). The codegen uses TWO strategies: parse-modify-serialize on the local `settings.json` (pure JSON), raw-text region replacement on the template `settings.json.tmpl` (mixed JSON + Go-template directives). The `.tmpl` cannot be JSON-parsed (the `{{jsonEscape .SmartPATH}}` directive at L379 is invalid JSON), so raw-text region replacement is the only viable approach for the template. Verified: the permissions block region inside `.tmpl` (L380-520) contains NO `{{...}}` directives — template directives live only in the `env` block outside permissions — so raw-text replacement on the permissions block is safe. A post-condition sentinel asserts the regenerated permissions block contains zero `{{`/`}}` substrings. See design.md §B.1 for the full justification.
- OQ-2: RESOLVED at plan-phase iter-2 (D6). Claude Code's settings.json `permissions` schema SUPPORTS the `ask` array natively — the runtime prompts the user for any tool invocation matching an `ask` entry. Verified empirically: the local `.claude/settings.json` already carries a non-empty `ask` array (6 entries: `Bash(rm:*)`, `Bash(sudo:*)`, `Bash(chmod:*)`, `Bash(chown:*)`, `Read(./.env)`, `Read(./.env.*)`), and the template `settings.json.tmpl` carries its own `ask` array (L502+). REQ-TPS-002's `decision ∈ {allow, deny, ask}` is therefore fully emittable — `ask` maps directly to the `permissions.ask` array. No "neither allow nor deny" workaround is needed.
- OQ-3: For env_gated rules (GLM), is the hook-matcher emission a new hook script or an extension of an existing one? M3 decides; prefer extending `handle-pre-tool.sh` over a new script.

---

## §I. Cross-References

- spec.md — requirements (REQ-TPS-001..009 + REQ-TPS-008b), seed scope (§D), exclusions (§X.1-§X.8).
- plan.md — milestones (M1-M6), known issues (KI-1..3, KI-2 corrected iter-2).
- acceptance.md — AC matrix (AC-TPS-001..014), drift-prevention test (AC-TPS-005 YAML↔settings.json scope), idempotency (AC-TPS-013), template round-trip (AC-TPS-014).
- research.md — scattered-policy inventory (§C, iter-2 corrections), book2 survey (§B), §24.5 analysis (§E, narrowed scope).
- `.moai/docs/harness-namespace-doctrine.md` §24.5 — drift class ANALOGY (L59-71; narrowed to YAML↔settings.json scope per §D.2).
- `.claude/rules/moai/core/glm-web-tooling.md` — env_gated rule source (L25-35).
- `.claude/rules/moai/core/zone-registry.md` + `internal/cli/constitution.go` — query PATTERN model (schema is DISJOINT per D9; not a wrapper target).
- CLAUDE.local.md §2 (Template-First), §22.1 (dev-vs-template split), §25 (template neutrality).

---

## Out of Scope

### Out of Scope — Canonical exclusions live in spec.md

- This design.md is a companion artifact; the canonical exclusions (§X.1-§X.8) live in `spec.md`. This section satisfies the lint `MissingExclusions` rule for this file.
