# plan.md — SPEC-V3R6-SEC-SKILL-INTEGRATION-001 Implementation Plan

> **Tier classification: L (large).** Upgraded from initial Tier M during plan-auditor iter-1 re-iteration (PASS-WITH-DEBT 0.82, BLOCKING D7). Justification in §B.4. Tier L mandates the 5-artifact set (spec.md + plan.md + acceptance.md + design.md + research.md); design.md is authored at `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/design.md`.

## §A. Context

This SPEC authors three new `moai-ref-*` defensive cybersecurity reference skills by **independently re-authoring** defensive knowledge from a community upstream repo used ONLY as a concept/procedure reference source. The upstream is Apache-2.0, but because the skills are independently re-authored (not copied), they are MoAI-ADK's own work and carry no upstream attribution.

The three deliverables:

1. `moai-ref-llm-security` — AI/LLM defensive security (~14 topics)
2. `moai-ref-supply-chain` — Supply chain defensive security (~8 topics)
3. `moai-ref-secops` — DevSecOps + Container + API operational security (~79 topics condensed)

Each skill is structurally identical to `moai-ref-owasp-checklist` / `moai-ref-api-patterns` / `moai-ref-testing-pyramid` (the existing `moai-ref-*` family).

**Parallel-session baseline (pre-write)**: local HEAD == `origin/main` (`0 0` divergence as of 2026-06-24), no active session on this SPEC, SPEC dir newly created. No race risk at plan-phase.

## §B. Known Issues

### §B.1 Upstream-as-reference-only discipline (binding item 1)

The upstream repository is a **knowledge/procedure reference source ONLY**. The run-phase author reads it for concepts, procedure ordering, tool names, technique IDs, and verification-step ideas — then writes fresh prose in MoAI-ADK style. The HARD constraint (the only legal duty) is: **do not copy the upstream's specific sentences, paragraphs, or script code.**

This is enforced three ways:
- **At authoring time**: the author treats each upstream skill as a concept outline, not a text source. They close the upstream file before writing the moai-adk skill body and write from the concept.
- **At review time (Phase 2 plan-auditor)**: the auditor spot-checks 3-5 passages of each finished skill body against the upstream and confirms no verbatim passage longer than standard tool-invocation syntax.
- **In research.md**: the upstream is recorded as the initial knowledge-inspiration source for maintainer traceability, with a note that verbatim reproduction is prohibited and that the skills are independent works.

### §B.2 16-language neutrality (binding item 2)

Per CLAUDE.local.md §15, the skills are distributed to user projects in any of the 16 MoAI-supported languages. The bodies must not privilege one language's security-tooling ecosystem. Concretely:
- Name tools at the ecosystem-neutral level: "container image scanner" (not "trivy" as the sole option), "SBOM generator" (not "syft" as the sole option), "SAST tool" (not "semgrep" as the sole option).
- Where a tool is the de-facto standard across ecosystems (e.g., `cosign` for Sigstore signing — a CNCF project with no real alternative), name it but frame it as "the standard Sigstore signing tool" rather than as a Python-specific or Go-specific tool.
- Examples in the body should be pseudo-code or CLI-shape (e.g., `tool --flag value`) rather than language-specific source code.
- The skill-body review (§E self-verification E5) greps for language-ecosystem bias markers (e.g., `pip install`, `npm install`, `go get`, `cargo add` appearing as the sole pattern).

### §B.3 Dual-use boundary (binding item 3)

Strictly defensive-only. The §F Out-of-Scope section in spec.md enumerates the excluded offensive categories (Red Teaming execution, C2 operation, Pen Testing exploitation, Mimikatz/Kerberoasting/NTLM-relay, credential dumping). The authoring discipline:
- Each skill body section frames content as **defense / hardening / detection / verification** — never as exploitation.
- Where a defensive topic inherently names an attack vector (e.g., "container escape defense" must mention that the escape exists), the skill describes the misconfiguration that enables it, how to detect it, and how to prevent it — but omits any step-by-step exploitation procedure.
- MITRE ATT&CK / ATLAS technique IDs and OWASP attack-pattern names MAY be cited for defensive correlation ("this hardening defends against T1xxx"), without providing the attack recipe.

### §B.4 Tier classification — Tier L (upgraded from initial Tier M)

This SPEC is classified **Tier L (large)**. The initial draft classified it Tier M based on file count (3 new SKILL.md files + 3 template mirrors + 1 internal research.md = 7 files; no Go code, hook, or config-schema changes; each skill body ~150-250 lines). During plan-auditor iter-1 review (PASS-WITH-DEBT 0.82), the BLOCKING defect D7 surfaced: the `tier:` frontmatter field was absent, which per the backward-compatibility rule defaults to Tier L. Rather than downgrade to Tier M, the re-iteration **upgraded explicitly to Tier L** based on three compounding factors that the initial M classification under-weighted:

**(a) The re-authorship pattern is novel in moai-adk.** No prior SPEC in this repository independently re-authors knowledge from a named upstream while maintaining an internal-only research.md as the sole attribution trace. This is a first-of-kind procedure pattern; first-of-kind patterns warrant the full 5-artifact Tier-L treatment (including design.md documenting the procedure) so future SPECs that follow the same pattern can reference this one as the canonical precedent.

**(b) Dual-use legal-integrity is a P0 surface.** The skills are distributed to user projects (template-managed, shipped via `moai init` / `moai update`). A re-authorship integrity failure or an offensive-procedure leak is a legal AND safety breach, not a quality regression. The legal-integrity argument (17 USC §102(b) / Korean Copyright Act §7 — ideas/procedures/facts not copyrightable; Apache-2.0 §4(c) binds derivatives only) must be documented in a dedicated design.md so a future maintainer (or an external auditor) can verify the legal basis without re-deriving it from first principles.

**(c) 16-language neutrality audit overlaps.** The skills must satisfy the template-internal-isolation doctrine (CLAUDE.local.md §25) AND the 16-language neutrality doctrine (CLAUDE.local.md §15) simultaneously. The overlap is non-trivial: a tool name that is "standard" in one dimension (e.g., `cosign` as the de-facto Sigstore signing tool) may still require the de-facto-standard carve-out framing to avoid ecosystem bias. The design.md documents the carve-out rule so the run-phase author applies it consistently across all three skills.

**Consequence of Tier L**: the artifact set expands from 3 files (spec.md + plan.md + acceptance.md) to 5 files (+ design.md + research.md). design.md lives at `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/design.md` and documents: (1) the re-authorship procedure (concept-read → close upstream → write fresh); (2) the legal-integrity rationale; (3) the upstream-containment boundary (research.md is internal-only, never embedded); (4) the dual-use defensive-framing policy. The plan-auditor PASS threshold rises from 0.80 (Tier M) to 0.85 (Tier L).

**Residual Tier-L escalation triggers (if run-phase reveals further scope)**:
- (a) The dual-use legal review reveals a topic that cannot be framed defensively without including exploitation material → split into a separate SPEC and re-audit.
- (b) The 16-language neutrality audit reveals that one of the three domains (e.g., supply-chain tooling) is so dominated by one ecosystem's tools that language-neutral framing is not achievable → expand the plan to address each ecosystem and re-audit.
- (c) The skill-body size for `moai-ref-secops` (the largest, ~79 topics) exceeds the HARD 400-line threshold (AC-SI-014) → split into `modules/devsecops.md` + `modules/container.md` + `modules/api-ops.md` with SKILL.md as a ~150-line overview (the split is already anticipated in AC-SI-014 and EC3).

## §C. Pre-flight Checks (before run-phase M1)

Before manager-develop starts M1 (first skill body), the following must be verified:

1. **Upstream reference access**: the author can reach the upstream repository (or a local clone) as a concept reference. If the upstream is unavailable, alternative public references (OWASP cheat-sheets, MITRE ATT&CK/ATLAS pages, NIST publications, vendor security advisories) cover the same defensive concepts.
2. **Existing `moai-ref-*` structure loaded**: the author has read `moai-ref-owasp-checklist/SKILL.md` and `moai-ref-api-patterns/SKILL.md` as the structural template.
3. **Skill-authoring standard loaded**: `.claude/rules/moai/development/skill-authoring.md` and `.claude/rules/moai/development/skill-writing-craft.md` are in context.
4. **Template-neutrality contract loaded**: CLAUDE.local.md §25 + `.moai/docs/template-internal-isolation-doctrine.md` §25.1 (forbidden/allowed content-class catalogue) are in context.
5. **Spot-check 3-5 MITRE/OWASP technique IDs**: the author independently verifies the technique-ID mappings the skills will cite against the canonical MITRE/OWASP source (not against the upstream).
6. **research.md scaffold exists**: this plan-phase includes a research.md skeleton recording the upstream as the inspiration source and enumerating the defensive concepts to be re-authored per skill (concept-level only, no upstream prose).

## §D. Constraints (restated from spec.md §D)

- **Legal**: ideas/procedures/facts not copyrightable; re-authored works are not derivative; no upstream attribution required; the only duty is to not copy verbatim text/scripts.
- **Dual-use**: strictly defensive; offensive categories excluded outright.
- **Template-neutrality §25**: no internal SPEC IDs / REQ tokens / audit citations / internal dates / commit SHAs in skill bodies.
- **16-language neutrality §15**: no language-ecosystem bias; tools named ecosystem-neutrally.
- **Skill size**: each SKILL.md ≤ ~500 lines.
- **Frontmatter listing cap**: `description` + `when_to_use` ≤ 1,536 chars each.

## §E. Self-Verification (manager-develop run-phase §E checklist)

The run-phase §E self-verification will include the standard E1-E7 plus skill-specific checks:

- **E1 (AC matrix)**: every AC in acceptance.md has a PASS/FAIL verdict with evidence.
- **E2 (build/CI)**: `make build` regenerates embedded template files; `go test ./internal/template/...` passes (skill-template parity tests).
- **E3 (coverage)**: N/A (no Go code changes); the coverage analog is "every skill body covers its declared defensive topics".
- **E4 (subagent-boundary grep)**: skills do not reference `AskUserQuestion` / `mcp__askuser` (they are `user-invocable: false` reference material, no prompting surface).
- **E5 (lint + skill-authoring lint)**: `golangci-lint run` passes; `moai spec lint` on this SPEC passes; the skill-frontmatter schema is valid per `skill-authoring.md`.
- **E6 (template-neutrality self-check)**: the 5-item template-neutrality pre-commit self-check (CLAUDE.local.md §25.3) passes for each skill file.
- **E7 (push state)**: the run-phase commit + the sync-phase commit are pushed; the skills are mirrored in `internal/template/templates/.claude/skills/moai-ref-*/`.

**Skill-specific self-checks** (added to §E):
- **E-REAUTHOR**: a spot-check of 3 passages from each skill body against the upstream confirms no verbatim reproduction of upstream expression beyond standard tool-invocation syntax. **Debt D1 (run-phase M4)**: raise the sample to ≥10 passages/skill OR add a mechanical full-sentence grep vs upstream — the 3-passage spot-check is a sampling floor, not a ceiling.
- **E-DUALUSE**: grep each skill body (case-insensitive `-i`) for the excluded offensive keywords per the expanded AC-SI-009 list (`mimikatz`, `kerberoast`, `ntlm relay`, `cobalt strike`, `sliver`, `metasploit`, `pass-the-hash`, `pass-the-ticket`, `lsass dump`, `bloodhound`, `rubeus`, `impacket`, `evil-winrm`, `responder`, `seatbelt`, `sharpersist`, `certipy`); expected: zero matches, OR only defensive-framing mentions (each match MUST appear within 1 sentence of a defensive verb: `detect|prevent|defend|block|alert|scan|harden|monitor|triage|investigate`). **Residual risk (debt D5/D6)**: the keyword grep is syntactic; a re-framed offense evading the grep remains a run-phase M4 human-verdict gate — the grep is necessary but not sufficient.
- **E-LANGNEUTRAL**: grep each skill body for language-ecosystem-bias patterns (`pip install`, `npm install`, `go get`, `cargo add`); expected: zero matches as the sole ecosystem pattern, OR balanced coverage (each package-manager reference appears in a balanced multi-ecosystem context — ≥3 ecosystems named in the same section per AC-SI-010's positive reframe).
- **E-NOATTRIB**: grep the distributed skill artifacts (`.claude/skills/moai-ref-{llm-security,supply-chain,secops}/` and `internal/template/templates/.claude/skills/moai-ref-{llm-security,supply-chain,secops}/`) for the upstream repo's name (`mukul975`, `anthropic-cybersecurity-skills`, `agentskills.io`), URL, or maintainer name (verbatim strings, case-insensitive); expected: zero matches. The internal research.md is excluded from this grep (AC-SI-008 scope: greps distributed paths only, NOT `.moai/specs/`).
- **E-CONTAIN** (new per AC-SI-018): grep `internal/template/embed.go` (the `//go:embed all:templates` source file) and `internal/template/templates/` for `research.md` / `mukul975` / `Anthropic-Cybersecurity` → expected: zero matches. Verifies research.md is never embedded in the template binary (upstream-containment boundary per design.md §C).

## §F. Milestones (Tier L — 5 milestones, no time estimates)

- **M1 — research.md + first skill (`moai-ref-llm-security`)**
  - The internal research.md already exists from plan-phase (records upstream as inspiration source + concept list for all three skills). Run-phase M1 authors the first skill against it.
  - Author `moai-ref-llm-security/SKILL.md` (local + template mirror).
  - Run the §E self-verification on this skill.
  - Commit: `feat(SPEC-V3R6-SEC-SKILL-INTEGRATION-001): M1 moai-ref-llm-security skill`

- **M2 — second skill (`moai-ref-supply-chain`)**
  - Author `moai-ref-supply-chain/SKILL.md` (local + template mirror).
  - Run §E on this skill.
  - Commit: `feat(SPEC-V3R6-SEC-SKILL-INTEGRATION-001): M2 moai-ref-supply-chain skill`

- **M3 — third skill (`moai-ref-secops`)**
  - Author `moai-ref-secops/SKILL.md` (local + template mirror). This is the largest (~79 upstream topics condensed). Per AC-SI-014, the HARD 400-line split threshold applies: if `wc -l > 400` for `moai-ref-secops/SKILL.md`, the milestone MUST split into `modules/devsecops.md` + `modules/container.md` + `modules/api-ops.md` with SKILL.md as a ~150-line overview (see EC3). The 400-line threshold is a HARD gate, not a tilde guidance — exceeding it without splitting is a FAIL.
  - Run §E on this skill.
  - Commit: `feat(SPEC-V3R6-SEC-SKILL-INTEGRATION-001): M3 moai-ref-secops skill`

- **M4 — cross-skill consistency + dual-use/legal review**
  - Cross-reference consistency between the three skills and the existing `moai-ref-owasp-checklist` / `moai-ref-api-patterns`.
  - Dual-use review: verify no offensive procedures leaked in (E-DUALUSE with the expanded keyword list per AC-SI-009).
  - Re-authorship integrity review: spot-check ≥3 passages per skill against upstream (E-REAUTHOR). **Debt D1**: at M4, raise the sample to ≥10 passages/skill OR add a mechanical full-sentence grep vs upstream (deferred to run-phase per plan-auditor debt tracking).
  - Template-neutrality self-check across all three skill files (E6 + E-LANGNEUTRAL + E-NOATTRIB). The E-NOATTRIB grep covers the distributed paths only (`.claude/skills/moai-ref-*/` and `internal/template/templates/.claude/skills/moai-ref-*/`); the internal research.md is excluded (it names the upstream verbatim per AC-SI-008 scope clarification).
  - Research.md containment check (E-CONTAIN, new per AC-SI-018): grep `internal/template/embed.go` + `internal/template/templates/` for `research.md` / upstream name → zero matches.
  - Commit: `feat(SPEC-V3R6-SEC-SKILL-INTEGRATION-001): M4 cross-skill consistency + dual-use + re-authorship review`

- **M5 — sync-phase (manager-docs)**
  - CHANGELOG entry.
  - README update if the skill catalog is listed.
  - Frontmatter `status: in-progress → implemented → completed` on the single sync commit.
  - Commit: `docs(SPEC-V3R6-SEC-SKILL-INTEGRATION-001): sync-phase artifacts` (carries the `completed` transition per the Status Transition Ownership Matrix).

## §G. Anti-Patterns

- **AP-SI-001 — Copy-paste from upstream**: copying the upstream's prose/script into the moai-adk skill body and lightly editing. This is a derivative work and triggers Apache-2.0 §4(c) attribution duty + the §25 template-neutrality breach (the upstream's moai-adk-internal language would leak). Correct: read upstream for concepts, close it, write fresh.
- **AP-SI-002 — Offensive leak**: including a "how to exploit this" section for completeness. Violates REQ-SI-006. Correct: defense/hardening framing only.
- **AP-SI-003 — Ecosystem bias**: naming one language's tools as the default (e.g., a supply-chain skill that only covers `pip-audit` and `safety`). Violates REQ-SI-013. Correct: name tools at the ecosystem-neutral level.
- **AP-SI-004 — Upstream attribution leak**: including "based on <upstream-repo>" in a comment or README. Violates REQ-SI-002/021. Correct: no attribution; the upstream is recorded ONLY in the internal research.md.
- **AP-SI-005 — Over-broad skill description**: a description that triggers on "security" generically, causing the skill to load on unrelated requests. Correct: narrow `description` + explicit `## NOT for` line per REQ-SI-012.
- **AP-SI-006 — Skill-body encyclopedia creep**: trying to cover all ~100 upstream topics verbatim-condensed, producing a 1500-line SKILL.md. Correct: select the defensive practitioner essentials; defer deep-dives to a follow-up SPEC.
- **AP-SI-007 — Treating MITRE IDs as attack recipes**: "T1xxx — here is how to do it" framing. Violates REQ-SI-008. Correct: "T1xxx — here is how to defend against it."

## §H. Cross-References

- **spec.md** — `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/spec.md` (this SPEC's requirements + Out-of-Scope)
- **acceptance.md** — `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/acceptance.md` (Given-When-Then scenarios + closure gates)
- **design.md** — `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/design.md` (Tier L 5th artifact — re-authorship procedure + legal-integrity rationale + upstream-containment boundary + dual-use defensive-framing policy)
- **research.md** — `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/research.md` (internal-only — records upstream `mukul975/Anthropic-Cybersecurity-Skills` as inspiration source + concept list; NOT distributed)
- **Structural template skills**: `.claude/skills/moai-ref-owasp-checklist/SKILL.md`, `.claude/skills/moai-ref-api-patterns/SKILL.md`, `.claude/skills/moai-ref-testing-pyramid/SKILL.md`
- **Skill-authoring standard**: `.claude/rules/moai/development/skill-authoring.md` (frontmatter schema + namespace policy), `.claude/rules/moai/development/skill-writing-craft.md` (description craft + body structure)
- **Template-neutrality doctrine**: `.moai/docs/template-internal-isolation-doctrine.md` §25.1 (forbidden/allowed content-class catalogue C1-C8) + CLAUDE.local.md §25
- **16-language neutrality**: CLAUDE.local.md §15
- **Frontmatter schema (SPEC)**: `.claude/rules/moai/development/spec-frontmatter-schema.md` (12 canonical fields + `tier:` optional field)
- **Status Transition Ownership Matrix**: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (this SPEC's status transitions: manager-spec authors draft; manager-develop M1 starts in-progress; manager-docs M5 single-sync-commit carries implemented→completed)
