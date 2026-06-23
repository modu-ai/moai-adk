---
id: SPEC-V3R6-SEC-SKILL-INTEGRATION-001
title: "Defensive cybersecurity practitioner reference skills (independent re-authorship)"
version: "0.3.0"
status: completed
created: 2026-06-24
updated: 2026-06-24
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai-ref-llm-security, .claude/skills/moai-ref-supply-chain, .claude/skills/moai-ref-secops"
lifecycle: spec-anchored
tier: "L"
era: V3R6
tags: "security, skills, defensive-cybersecurity, reference-skills, re-authorship, supply-chain, llm-security, devsecops, container-security, api-security"
---

# SPEC-V3R6-SEC-SKILL-INTEGRATION-001 — Defensive Cybersecurity Reference Skills

## §A. Background and Problem

MoAI-ADK's existing security skill surface covers **development-time application security only** — `moai-ref-owasp-checklist` (web-app vulnerabilities + auth patterns), the `/moai security` command, and the (archived) `expert-security` agent. This is intentionally narrow: it helps a developer writing code avoid OWASP Top 10 mistakes.

However, a modern software practitioner — the MoAI-ADK target user — increasingly needs **defensive operational cybersecurity knowledge** that the current surface does not cover:

- **AI/LLM security**: defending LLM-backed applications against prompt injection, training-data poisoning, MCP/agentic tool abuse, and model-output attacks (OWASP LLM Top 10, MITRE ATLAS, NIST AI RMF).
- **Supply-chain security**: SBOMs, dependency-confusion, malicious-package triage, SLSA provenance, Sigstore signing.
- **DevSecOps / Container / API operations**: CI/CD hardening, IaC scanning, K8s RBAC, image scanning, OWASP API Top 10 (which overlaps but extends the dev-focused OWASP checklist), WAF tuning.

These domains are **complementary** to the existing dev-security surface, not duplicates. A user asking "how do I triage a suspicious npm package I just pulled?" or "how do I harden my K8s RBAC?" or "how do I detect prompt injection in my MCP server?" has no current MoAI-ADK reference to load.

A large external community repository of cybersecurity skills exists (817 skills, Apache-2.0, agentskills.io standard). It is a high-quality knowledge index of procedures, tool commands, verification steps, MITRE/OWASP technique IDs, and CVE references across 29 domains. **However**, directly vendoring (copying) it would (a) create a large attribution/license-maintenance surface, (b) import the upstream's stylistic choices rather than MoAI-ADK's, and (c) pull in dual-use offensive content (Red Teaming, C2, Pen Testing execution, credential dumping) that MoAI-ADK — a defensive practitioner tool — must not distribute.

## §B. Goal and Solution

Author **three new `moai-ref-*` defensive cybersecurity reference skills** by **independently re-authoring** defensive-cybersecurity knowledge in MoAI-ADK style. The external community repo is used ONLY as a **knowledge/procedure reference source** (concepts, tool names, technique IDs, procedure ordering). No upstream sentences, paragraphs, or script code are copied.

**The three skills (proposed naming — confirmed in §F):**

1. **`moai-ref-llm-security`** — AI / LLM Security (~14 upstream topics re-authored): LLM red-team *defense*, prompt-injection defense, training-data poisoning detection, MCP/agentic tool-call hardening, guardrails, model-output validation, OWASP LLM Top 10 defensive mapping, ATLAS technique detection.
2. **`moai-ref-supply-chain`** — Software Supply Chain Security (~8 topics): SBOM generation/verification, dependency-confusion defense, malicious-package triage playbooks, SLSA provenance levels, Sigstore/cosign signing, package-registry hardening, typosquatting defense.
3. **`moai-ref-secops`** — DevSecOps + Container + API operational security (~79 topics condensed): CI/CD pipeline hardening, secret scanning, IaC scanning (Terraform/CloudFormation), container image scanning, K8s RBAC hardening, container-escape defense, Falco runtime rules, OWASP API Top 10 defense (operationally distinct from the dev-focused `moai-ref-owasp-checklist`), WAF rule tuning, GraphQL/REST rate-limiting and authorization enforcement.

Each skill is a **`moai-ref-*` agent-extending reference skill** — structurally identical to the existing `moai-ref-owasp-checklist` / `moai-ref-api-patterns` / `moai-ref-testing-pyramid` family: `user-invocable: false`, MoAI progressive-disclosure frontmatter, ~150-250 line body, markdown tables / checklists / verification sections. They are **reference knowledge loaded on demand**, not user-invoked slash commands.

## §C. Requirements (GEARS notation — current standard)

### §C.1 Re-authorship and legal-integrity requirements

**REQ-SI-001 (Ubiquitous).** The skill bodies shall present defensive cybersecurity procedures, tool references, MITRE/OWASP/CVE identifiers, and verification steps re-authored in MoAI-ADK style without copying the upstream repository's specific sentences, paragraphs, or script code.

**REQ-SI-002 (Ubiquitous).** The distributed skill artifacts (SKILL.md files under `.claude/skills/moai-ref-*/` and their template mirrors under `internal/template/templates/`) shall carry no attribution, name, brand, URL, or other identifying marker of the external upstream repository.

**REQ-SI-003 (State-driven).** **While** a run-phase author is drafting skill body content, the author shall treat the upstream as a concept/procedure reference only and shall write fresh prose/examples in MoAI-ADK style, with the research.md maintainer-only memo recording the upstream as the initial knowledge-inspiration source.

**REQ-SI-004 (Unwanted behavior).** The skills shall not include verbatim reproduction of any upstream text or script longer than a standard tool invocation (tool name + standard flags) that would constitute copying protectable expression.

### §C.2 Dual-use boundary requirements (defensive-only subset)

**REQ-SI-005 (Ubiquitous).** The three skills shall cover defensive practitioner playbooks ONLY — detection, hardening, response, verification, and defense-in-depth guidance.

**REQ-SI-006 (Unwanted behavior).** The skills shall not contain offensive procedures, attack execution instructions, credential-dumping techniques (e.g., Mimikatz usage, Kerberoasting, NTLM-relay attack steps), command-and-control (C2) framework operational guides, or penetration-testing exploit execution.

**REQ-SI-007 (State-driven).** **While** documenting a technique that has both offensive and defensive uses (e.g., "how K8s RBAC misconfiguration leads to container escape"), the skill shall frame the content as defense/hardening (here is the misconfiguration, here is how to detect it, here is how to prevent it) and shall omit any step-by-step exploitation procedure.

**REQ-SI-008 (Where).** **Where** a defensive topic references a known offensive technique (MITRE ATT&CK technique ID, OWASP attack pattern name), the skill shall cite the technique by its public standard identifier (T-ID, OWASP label, CVE number) for defensive correlation without providing an attack recipe.

### §C.3 Structural and frontmatter requirements

**REQ-SI-009 (Ubiquitous).** Each of the three skills shall be a `moai-ref-*` reference skill with `user-invocable: false`, the MoAI `progressive_disclosure` extension block, and a `metadata` block carrying `version`, `category`, `status`, `updated`, and `tags` per `.claude/rules/moai/development/skill-authoring.md`.

**REQ-SI-010 (Event-driven).** **When** Claude Code's skill-listing matcher evaluates a security-domain request, each skill's `description` + `when_to_use` fields shall trigger on the appropriate defensive-security keyword set within the 1,536-character listing cap.

**REQ-SI-011 (Ubiquitous).** Each skill body shall follow the progressive-disclosure structure established by `moai-ref-owasp-checklist` and `moai-ref-api-patterns`: a domain overview, one or more markdown tables/checklists mapping threats to defenses, a verification checklist section, and the three MoAI-evolvable blocks (rationalizations / red-flags / verification).

**REQ-SI-012 (Ubiquitous).** Each skill shall include a `## NOT for` line in its `description` frontmatter explicitly listing out-of-scope domains (offensive techniques, other skills' scope) to prevent over-triggering.

### §C.4 Template-neutrality and 16-language neutrality requirements

**REQ-SI-013 (Ubiquitous).** The skill bodies shall remain language-neutral across the 16 MoAI-supported languages — examples and tool references shall use standard, ecosystem-agnostic tool names (e.g., "SAST tool of choice", "container scanner", "SBOM generator") rather than privileging one language's ecosystem, **except where a tool is the de-facto cross-ecosystem standard for its domain (e.g., cosign for Sigstore signing, trivy/grype for image scanning, semgrep for SAST), in which case it MAY be named but MUST be framed as "the standard \<category\> tool" rather than as a language-specific tool.** A tool that only exists for one ecosystem is named neutrally (its ecosystem-specificity is not hidden, but it is not privileged as the default). **De-facto-standard criterion (plan-auditor iter-2 D12)**: a tool qualifies as a de-facto cross-ecosystem standard if and only if it has no functional alternative across ≥3 of the 16 MoAI-supported language ecosystems; borderline cases defer to the run-phase author's neutrality judgment, with the judgment rationale recorded in progress.md §E.2 so a reviewer can audit the borderline call.

**REQ-SI-014 (Ubiquitous).** The skills shall not embed moai-adk internal SPEC IDs, REQ/AC tokens, audit citations ("Finding AX"), internal dates, or commit SHAs in their bodies — they are distributed to user projects and must satisfy the template-neutrality §25 contract.

**REQ-SI-015 (Where).** **Where** the skills reference the existing `moai-ref-owasp-checklist` or `moai-ref-api-patterns` for cross-domain coverage (e.g., the dev-time OWASP checklist vs the operational API security material), the skills shall cross-reference by skill name without duplicating content.

### §C.5 Scope-allocation requirements (3-skill split)

**REQ-SI-016 (Ubiquitous).** The skill `moai-ref-llm-security` shall cover AI/LLM defensive security: LLM application hardening, prompt-injection defense, training-data poisoning detection, MCP/agentic tool-call hardening, model-output validation, guardrails, OWASP LLM Top 10, MITRE ATLAS, NIST AI RMF defensive mapping.

**REQ-SI-017 (Ubiquitous).** The skill `moai-ref-supply-chain` shall cover software supply-chain defensive security: SBOMs, dependency confusion, malicious-package triage, SLSA provenance, Sigstore/cosign, package-registry hardening, typosquatting defense, transitive-dependency auditing.

**REQ-SI-018 (Ubiquitous).** The skill `moai-ref-secops` shall cover DevSecOps + Container + API operational security: CI/CD hardening, secret scanning, IaC scanning, image scanning, K8s RBAC, container-escape defense, runtime detection (Falco), OWASP API Top 10 operational defense, WAF, GraphQL/REST authorization enforcement.

**REQ-SI-019 (Ubiquitous).** The skill `moai-ref-secops` shall distinguish its operational API-security coverage from the development-focused `moai-ref-owasp-checklist` by emphasizing runtime/operational concerns (rate-limit enforcement, WAF tuning, broken-auth detection in production, server-side authorization) rather than the dev-time patterns already covered.

### §C.6 Upstream-tracking-lightness and verification requirements

**REQ-SI-020 (Ubiquitous).** A `research.md` file inside `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/` shall record the upstream repository as the initial knowledge-inspiration source for maintainer traceability; this file is internal-only and is NOT distributed with the skills.

**REQ-SI-021 (Unwanted behavior).** The skill artifacts shall not reference the upstream repository's URL, GitHub path, maintainer name, or any other identifier anywhere in `.claude/skills/` or `internal/template/templates/.claude/skills/`.

**REQ-SI-022 (Event-driven).** **When** a reviewer or auditor needs to verify the re-authorship integrity of a skill body, the reviewer shall compare a sample of skill-body prose against the upstream and confirm that no verbatim passage longer than standard tool-invocation syntax was reproduced.

**REQ-SI-023 (Unwanted behavior).** The internal `research.md` file (which records the upstream repository by name for maintainer traceability) shall not be embedded in the template binary — it shall not appear in `internal/template/embed.go` (the `//go:embed all:templates` source file) nor anywhere under `internal/template/templates/`, ensuring the upstream name never ships to user projects via `moai init` / `moai update`.

## §D. Constraints (Non-Functional)

- **Legal constraint (17 USC §102(b) / Korean Copyright Act §7)**: ideas, procedures, methods, facts are not copyrightable; the defensive cybersecurity skill content (which tool, which command, which order, which verification, MITRE technique IDs, CVE numbers, command syntax) is overwhelmingly facts/procedures/standards and is freely referenceable. Apache-2.0 §4(c) attribution duty applies only to derivative works; independently re-authored works that do not reproduce upstream expression are not derivative.
- **HARD constraint**: the skills MUST NOT copy upstream verbatim text/script. This is the only legal duty. Concept-level reference is permitted and expected.
- **Dual-use constraint**: strictly defensive subset. Offensive techniques excluded outright (no Red Teaming, C2, Pen Testing execution, Mimikatz, Kerberoasting, NTLM-relay, credential dumping).
- **Template-neutrality constraint**: skill bodies must pass `internal/template/internal_content_leak_test.go` + `.github/workflows/template-neutrality-check.yaml` (C3 internal dates + C7 commit-hashes + the C1-C8 forbidden/allowed content-class catalogue per CLAUDE.local.md §25).
- **16-language neutrality constraint**: no Go/Python/CLI-ecosystem bias; tools named at the ecosystem-neutral level.
- **Skill-size constraint**: each SKILL.md ≤ ~500 lines (progressive disclosure; overflow into `modules/` if needed, but the design intent is compact reference, not encyclopedia).
- **Frontmatter listing-cap constraint**: `description` + `when_to_use` combined ≤ 1,536 characters per skill.

## §E. Assumptions

1. The external upstream repository remains accessible as a knowledge reference during plan-phase research and run-phase authoring; if it becomes unavailable mid-run, the authoring can proceed from publicly available alternative defensive-security references (OWASP, MITRE, NIST primary sources).
2. The MoAI-ADK maintainer is willing to assume full editorial responsibility for the re-authored skill quality (no upstream quality inheritance, no upstream maintenance dependency).
3. The three-skill split (LLM / Supply Chain / SecOps) reflects the natural domain boundaries and matches how a defensive practitioner mentally organizes the material; alternative splits (e.g., 5 skills by domain, or 1 consolidated skill) were considered in plan.md §F and rejected.
4. MITRE ATT&CK / ATLAS / OWASP / NIST technique IDs and CVE numbers are public-standard identifiers and may be cited freely; the skills cite them for defensive correlation, not as attack recipes.
5. The skills are reference material for Claude Code to load on demand — they are not user-invoked commands and have no imperative workflow side-effects.

## §F. Scope Decisions

### Three-skill split (confirmed)

The three-skill split was chosen over alternatives:

- **Alternative A (single `moai-ref-cybersecurity` skill)**: rejected — a single skill covering ~100 defensive topics would exceed the ~500-line SKILL.md ceiling, force aggressive module-splitting, and weaken skill-listing matcher precision (one description triggers on too many keyword sets).
- **Alternative B (five skills by domain)**: rejected — splitting Container from DevSecOps from API Security would create three narrow skills with overlapping vocabulary; the operational-practitioner mental model treats these as one "SecOps" domain.
- **Alternative C (two skills: LLM-Security + everything-else)**: rejected — the supply-chain domain has a distinct vocabulary (SBOM, SLSA, Sigstore, provenance) that would dilute if merged into a general "SecOps" skill.

**Confirmed split**: `moai-ref-llm-security` (AI/LLM) / `moai-ref-supply-chain` (Supply Chain) / `moai-ref-secops` (DevSecOps + Container + API operational).

### Defensive-only subset (confirmed)

The skills cover ONLY defensive procedures. Offensive techniques (Red Teaming execution, C2 framework operation, Pen Testing exploitation, Mimikatz/Kerberoasting/NTLM-relay attack steps, credential dumping) are **excluded entirely**. MITRE ATT&CK technique IDs MAY be cited for defensive correlation (e.g., "this hardening step defends against technique T1xxx") without providing the attack recipe.

## §G. Risks

- **R1 — Unintentional verbatim similarity**: even with independent re-authorship, a short procedural sentence (e.g., "run `cosign verify-key ...` with the public key") may coincidentally match the upstream's wording because the tool's standard invocation is fixed. Mitigation: REQ-SI-004 permits standard tool-invocation syntax; the reviewer compares prose passages, not standard CLI invocations.
- **R2 — Re-authorship quality is entirely MoAI-ADK's responsibility**: no upstream quality inheritance. Mitigation: each skill body's defensive claims must cite a public-standard source (OWASP doc, MITRE page, NIST publication, vendor security advisory) at the concept level, recorded in research.md.
- **R3 — MITRE/OWASP mapping accuracy**: the re-authored skills must verify their technique-ID mapping independently of the upstream. Mitigation: plan.md §C pre-flight requires the author to spot-check 3-5 technique IDs against the canonical MITRE/OWASP source before finalizing each skill.
- **R4 — Over-triggering in skill-listing matcher**: a broad `description` keyword set could cause the skills to load on unrelated requests. Mitigation: REQ-SI-012 mandates a `## NOT for` line narrowing scope; the description craft follows `skill-writing-craft.md` Part 1.
- **R5 — Dual-use edge cases**: a defensive topic (e.g., "container escape hardening") inherently names the attack vector, which a malicious actor could read as reconnaissance. Mitigation: REQ-SI-007 mandates defense/hardening framing (misconfiguration → detect → prevent), never step-by-step exploitation.
- **R6 — Scope creep into offensive**: a practitioner reference naturally tempts the author to add "and here is how an attacker would exploit it" for completeness. Mitigation: REQ-SI-006 + the §B Out-of-Scope section make this a hard exclusion; the plan-auditor gate (Phase 2 of the workflow) is the backstop.

## §H. HISTORY

- **2026-06-24 (v0.1.0 draft)** — Initial plan-phase artifacts authored by manager-spec. Origin: user decision (Option B' — independent re-authorship, defensive-only subset, no upstream attribution) recorded in maintainer memory `project_cybersec_skill_integration_decision.md`. Three-skill split confirmed. Initially classified Tier M.
- **2026-06-24 (v0.2.0 draft)** — Plan-auditor iter-1 PASS-WITH-DEBT (0.82). Re-iteration resolved 1 BLOCKING defect (D7: tier field absent) + SHOULD-FIX defects (D2/D3/D4 + open-decisions). Tier upgraded from M to **L** based on: (a) the re-authorship pattern is novel in moai-adk (no prior SPEC independently re-authored from a named upstream with an internal-only research.md); (b) dual-use legal-integrity is a P0 surface; (c) 16-language neutrality audit overlaps. Added `design.md` (5th Tier-L artifact) documenting the re-authorship procedure, the legal-integrity rationale, the upstream-containment boundary, and the dual-use defensive-framing policy. Added REQ-SI-023 + AC-SI-018 (research.md containment). Amended REQ-SI-013 (de-facto-standard-tool carve-out). Named upstream verbatim in research.md (AC-SI-008 scope clarified: greps distributed paths only). AC-SI-009 expanded offensive-keyword list + defensive-verb proxy. AC-SI-010 reframed as positive multi-ecosystem balance AC. AC-SI-013 measurement specified (`wc -c` on de-YAML-folded-scalar). AC-SI-014 HARD 400-line secops split threshold.
- **2026-06-24 (v0.3.0 draft)** — Plan-auditor iter-2 verification-wording fixes (D10-D13). All four defects resolved with live-verification evidence pasted per verification-claim-integrity §1.1: **D11** (filename correction): `internal/template/embedded.go` does NOT exist at HEAD; the real `//go:embed all:templates` source file is `internal/template/embed.go` (confirmed via `ls internal/template/*.go` + Read of embed.go line 28). Replaced every occurrence of `embedded.go` → `embed.go` across spec.md (REQ-SI-023), acceptance.md (AC-SI-018), design.md (§C.2/§C.3), plan.md (§E E-CONTAIN ×2), research.md (§E). **D10** (AC-SI-018 grep unsatisfiable): the prior pattern `grep 'research\.md\|mukul975\|Anthropic-Cybersecurity'` produced 74 pre-existing `research.md` matches across `internal/template/` (legitimate design-doc references), making the AC unsatisfiable-by-design. Dropped the `research\.md` alternative; kept ONLY the upstream-name signals `mukul975|anthropic-cybersecurity`, scoped to embed-source paths (`internal/template/templates/`, `internal/template/embed.go`, `internal/template/embed_catalog.go`). Live run returned **0 matches** — containment holds structurally (`.moai/specs/` lives outside the `//go:embed all:templates` directive's `templates/` subtree by construction; the grep is the mechanical guard against future regression). **D12** (REQ-SI-013 de-facto-standard criterion): added a one-sentence rule after the example list — "a tool qualifies as a de-facto cross-ecosystem standard iff it has no functional alternative across ≥3 of the 16 MoAI-supported language ecosystems; borderline cases defer to the run-phase author's neutrality judgment recorded in progress.md §E.2." **D13** (REQ-SI-008 uncovered): REQ-SI-008 previously had zero AC mappings. Added REQ-SI-008 to the §A traceability matrix rows for AC-SI-002 (MITRE ATLAS defensive-correlation technique-ID count) and AC-SI-009 (no attack recipes / defensive framing). AC-SI-002 body annotates the ATLAS bullet as operationalizing REQ-SI-008; AC-SI-009 body adds explicit REQ-SI-008 coverage paragraph (technique-ID citation + attack recipe = FAIL). `grep -c 'REQ-SI-008' acceptance.md` confirmed ≥2 after edit. `moai spec lint spec.md` remains clean.

## Exclusions

### Out of Scope — Offensive / dual-use attack procedures

- Red Teaming execution playbooks (recon, initial-access, lateral-movement procedures)
- Command-and-Control (C2) framework operational guides (Cobalt Strike, Sliver, Mythic operation)
- Penetration Testing exploitation execution (Metasploit modules, exploit chaining steps)
- Credential dumping techniques (Mimikatz usage, LSASS dumping, Kerberoasting attack steps, NTLM-relay attack execution, pass-the-hash / pass-the-ticket execution)
- Attack recipes for the techniques cited defensively (a skill may name MITRE ATT&CK T1xxx as a defended-against technique, but MUST NOT provide the exploitation procedure)

### Out of Scope — Upstream vendoring / attribution

- Directly copying / porting / vendoring the external upstream repository's SKILL.md files, scripts, or prose into moai-adk
- Including the upstream repository's name, URL, GitHub path, maintainer name, license-notice text, or any other identifying marker in any distributed artifact
- Treating the upstream as a maintenance dependency (upstream updates do NOT auto-flow into moai-adk; future upstream changes are re-evaluated manually against MoAI-ADK's own quality bar)

### Out of Scope — Non-defensive cybersecurity domains

- Digital Forensics & Incident Response (DFIR) deep-dive procedures (evidence collection, memory forensics, disk forensics tool operation) — out of scope for this SPEC; may be a future SPEC if practitioner demand emerges
- Threat hunting methodology (hypothesis-driven hunt procedures) — out of scope for this SPEC
- Cloud-security operations beyond what naturally falls under Container/K8s and API (AWS/Azure/GCP CSPM deep-dives) — out of scope; the skills stay ecosystem-neutral
- Security-compliance / GRC (SOC2, ISO 27001, PCI-DSS audit procedures) — out of scope; not practitioner-playbook material
- Full-stack framework-specific security (Django security middleware, Spring Security config, Rails strong-params deep-dive) — out of scope; this is the domain of the 16 language rules + `moai-ref-owasp-checklist`, not the new defensive-ops skills

### Out of Scope — Skill-as-user-command

- The three skills are NOT user-invoked slash commands. They are `user-invocable: false` reference skills loaded by Claude Code on demand. A `/moai security-ops` or similar command would be a separate SPEC.
