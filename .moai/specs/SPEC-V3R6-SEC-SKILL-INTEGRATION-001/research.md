# research.md — SPEC-V3R6-SEC-SKILL-INTEGRATION-001 (INTERNAL — maintainer traceability only)

> **This file is INTERNAL to the moai-adk repository.** It is NOT distributed with the skills. It records the upstream repository as the initial knowledge-inspiration source and enumerates the defensive concepts to be re-authored. It contains NO upstream verbatim text (no copied sentences, paragraphs, or scripts — only concept labels, public-standard references, and re-authoring notes).
>
> **Re-authorship discipline**: the upstream is a concept/procedure reference source ONLY. Run-phase authors read the upstream for concepts, procedure ordering, tool names, technique IDs, and verification-step ideas — then write fresh prose in MoAI-ADK style. The HARD constraint (the only legal duty) is: do not copy the upstream's specific sentences, paragraphs, or script code. Independent re-authorship is not a derivative work (17 USC §102(b) / Korean Copyright Act §7 — ideas/procedures/facts are not copyrightable; Apache-2.0 §4(c) attribution duty applies only to derivative works). See design.md §A-B for the full legal-integrity rationale.
>
> **AC-SI-008 scope clarification (plan-auditor iter-1 D2 inversion)**: this file names the upstream verbatim because the AC-SI-008 no-attribution grep covers the DISTRIBUTED paths only (`.claude/skills/moai-ref-{llm-security,supply-chain,secops}/` and `internal/template/templates/.claude/skills/moai-ref-{llm-security,supply-chain,secops}/`), NOT `.moai/specs/`. This file is the SOLE location where the upstream name appears (per design.md §C upstream-containment boundary). AC-SI-018 verifies mechanically that this file is never embedded in the template binary.

## §A. Upstream Reference Source (knowledge-inspiration, NOT vendored)

- **Upstream repository**: `mukul975/Anthropic-Cybersecurity-Skills` — a community cybersecurity-skills repository (817 skills, Apache-2.0, follows the `agentskills.io` standard). GitHub URL: `https://github.com/mukul975/Anthropic-Cybersecurity-Skills`. The `agentskills.io` standard reference is at `https://agentskills.io`.
- **License**: Apache-2.0. The re-authored moai-adk skills are independent works (no verbatim reproduction of upstream expression per design.md §A procedure) and therefore NOT derivative works; no attribution duty attaches (Apache-2.0 §4(c) binds derivatives only).
- **Use**: concept/procedure reference ONLY. The upstream provides the topic inventory, the procedure-ordering, and the tool/technique-ID index. The moai-adk skills re-author the defensive subset in MoAI-ADK style.
- **Maintenance**: moai-adk does NOT auto-track upstream. Future upstream changes are re-evaluated manually against MoAI-ADK's quality bar via a follow-up SPEC if warranted.
- **Why named here verbatim (not in a memory file)**: plan-auditor iter-1 Open Decision #3 inverted the upstream-naming posture. The prior draft recorded the upstream only in the maintainer memory file `project_cybersec_skill_integration_decision.md`, obscuring the re-authorship source from the SPEC directory. Naming it verbatim here is SAFE (AC-SI-008 greps distributed paths only; AC-SI-018 verifies containment) and improves auditability — a future maintainer or external auditor reads this file to verify the re-authorship provenance without consulting an external memory file.

## §B. Defensive-Domain Concept Inventory (concept-level only — no upstream prose)

The following defensive concepts will be re-authored across the three skills. Each entry is a concept label + public-standard reference (OWASP / MITRE / NIST), NOT an upstream passage.

### B.1 `moai-ref-llm-security` — AI/LLM defensive security

| Concept | Public-standard reference | Re-authoring note |
|---------|---------------------------|-------------------|
| Prompt-injection defense | OWASP LLM Top 10 (LLM01), MITRE ATLAS | Frame as defense: system-prompt isolation, input sanitization, output validation, instruction-hierarchy enforcement |
| LLM application hardening | OWASP LLM Top 10 (LLM01-10) | Cover at least LLM01 (prompt injection), LLM02 (insecure output), LLM03 (training-data poisoning), LLM05 (supply chain), LLM06 (sensitive disclosure), LLM07 (insecure plugin design) — defensive mapping |
| MCP / agentic tool-call hardening | OWASP LLM07, Anthropic MCP security guidance | Tool-permission scoping, output-validation on tool results, least-privilege tool design |
| Training-data poisoning detection | OWASP LLM03, MITRE ATLAS | Data-lineage, canary artifacts, provenance verification |
| Model-output validation | OWASP LLM02 | Guardrails, structured-output enforcement, content filtering, schema validation |
| MITRE ATLAS defensive correlation | MITRE ATLAS (public) | Cite ≥3 technique IDs for DEFENSIVE correlation (here is how to detect/defend), never as attack recipe |
| NIST AI RMF defensive mapping | NIST AI RMF 1.0 (public) | Govern / Map / Measure / Manage functions referenced for defensive governance |

### B.2 `moai-ref-supply-chain` — Software supply-chain defensive security

| Concept | Public-standard reference | Re-authoring note |
|---------|---------------------------|-------------------|
| SBOM generation/verification | SPDX, CycloneDX, NTIA minimum-elements | Ecosystem-neutral SBOM-tool naming; SPDX vs CycloneDX format comparison; syft as "the standard SBOM-generation tool" if named (de-facto cross-ecosystem standard per REQ-SI-013 carve-out), framed as category |
| Dependency-confusion defense | Alex Birsan disclosure (2021), package-registry docs | Private-namespace scoping, install-time verification, registry-side routing controls |
| Malicious-package triage | CISA guidance, package-registry security advisories | Triage signals (typo, maintainer change, install-script), quarantine, upstream-reporting channel |
| SLSA provenance | SLSA framework (slsa.dev, public) | The 4 levels, what each guarantees, how to verify provenance |
| Sigstore / cosign signing | Sigstore project (CNCF, public) | cosign as "the standard Sigstore signing tool" (de-facto cross-ecosystem standard per REQ-SI-013 carve-out); keyless signing, identity-based verification, signature-verification at install time; framed as category, not as Go-ecosystem tool |
| Package-registry hardening | Registry-vendor security docs | Registry-side signing, mirror-pinning, 2FA admin |
| Typosquatting defense | Dependency-name research, allowlist patterns | Automated typo-detection, allowlist enforcement, namespace registration |
| Transitive-dependency auditing | OWASP A06, SCA-tool docs | Depth limits, license/vuln scanning, lockfile-pinning |

### B.3 `moai-ref-secops` — DevSecOps + Container + API operational security

| Concept | Public-standard reference | Re-authoring note |
|---------|---------------------------|-------------------|
| CI/CD pipeline hardening | CISA CI/CD security guide, OWASP | Secret-scanning, signed-pipeline, runner-isolation, least-privilege pipeline tokens |
| IaC scanning | CIS Benchmarks; KICS/Checkov as "the standard IaC-scanning tools" (de-facto cross-ecosystem standards per REQ-SI-013 carve-out) | Terraform/CloudFormation misconfiguration detection; name KICS + Checkov as "the standard IaC-scanning tools" (not as ecosystem-specific); frame as category, not vendor |
| Image scanning | CIS Docker Benchmark; Trivy/Grype as "the standard container-image-scanning tools" (de-facto cross-ecosystem standards per REQ-SI-013 carve-out) | Base-image hygiene, layer-scanning, admission-control (sign + verify at deploy); name Trivy + Grype as "the standard container-image-scanning tools" (not as ecosystem-specific) |
| K8s RBAC hardening | CIS Kubernetes Benchmark, NSA K8s hardening guide | Least-privilege RoleBindings, ServiceAccount scoping, PodSecurity admission |
| Container-escape defense | NSA K8s hardening guide, MITRE ATT&CK (defensive correlation) | seccomp, AppArmor, read-only root fs, non-root user, defensive framing only |
| Runtime detection | Falco as "the standard runtime-detection tool for K8s" (de-facto cross-ecosystem standard per REQ-SI-013 carve-out); Cilium/Tetragon as alternatives in the same category | Rule-shape for runtime-detection (syscall-level); name Falco as "the standard K8s runtime-detection tool" framed as category (not vendor-lock); show rule-shape so the concept transfers to any runtime-detection tool |
| OWASP API Top 10 operational | OWASP API Security Top 10 (public) | BOLA detection in production, broken-auth detection, rate-limit enforcement, server-side authorization — operational distinct from `moai-ref-owasp-checklist` dev-time patterns |
| WAF rule-tuning | OWASP ModSecurity CRS (rule-shape) | Rule-shape, not vendor-specific; false-positive tuning, signal-to-noise |
| GraphQL/REST depth/complexity limiting | OWASP, GraphQL security guidance | Depth limiting, complexity limiting, persisted-query allowlisting, rate limiting |

## §C. Re-authorship Procedure (for run-phase authors)

> See design.md §A for the canonical 5-step procedure (concept-read → close upstream → write fresh → defensive-framing check → spot-check at M4). The procedure is binding on run-phase authors; deviation is AP-SI-001 (copy-paste from upstream).

1. **Read for concepts**: open the upstream skill for the topic; extract the concept outline (which tool, which order, which verification, which technique ID). Do NOT copy sentences.
2. **Close the upstream**: before writing the moai-adk skill body, close the upstream file. Write from the concept.
3. **Write fresh in MoAI-ADK style**: use the existing `moai-ref-owasp-checklist` / `moai-ref-api-patterns` body style — markdown tables, checklists, verification sections, the three evolvable blocks.
4. **Cite public-standard references**: each defensive claim cites a public-standard source (OWASP page, MITRE page, NIST pub, vendor advisory) at the concept level — NOT the upstream.
5. **Spot-check at M4**: select ≥3 passages per skill (≥3 sentences each); compare against upstream; confirm no verbatim reproduction. Standard tool-invocation syntax is exempt. Debt D1: raise to ≥10 passages/skill OR add mechanical full-sentence grep at M4.

## §D. Out-of-Scope (defensive-only enforcement)

The following upstream topics are EXCLUDED from re-authoring (offensive / dual-use):
- Red Teaming execution playbooks
- C2 framework operational guides (Cobalt Strike, Sliver, Mythic)
- Penetration Testing exploitation execution (Metasploit modules)
- Credential dumping (Mimikatz, Kerberoasting, NTLM-relay, pass-the-hash, pass-the-ticket, LSASS dumping)
- Any step-by-step exploitation procedure for a technique cited defensively

## §E. Maintainer Note

This file is the SOLE location in the repository where the upstream repository is identified by name (per design.md §C upstream-containment boundary). The distributed skill artifacts (`.claude/skills/moai-ref-*/` and `internal/template/templates/.claude/skills/moai-ref-*/`) carry NO upstream attribution.

- **AC-SI-008** verifies this with a grep over the DISTRIBUTED paths only (`.claude/skills/moai-ref-{llm-security,supply-chain,secops}/` and `internal/template/templates/.claude/skills/moai-ref-{llm-security,supply-chain,secops}/`) — this file (`.moai/specs/`) is excluded from that grep's scope, which is why naming the upstream verbatim here is SAFE.
- **AC-SI-018** verifies mechanically that this file is never embedded in the template binary — it greps `internal/template/embed.go` (the `//go:embed all:templates` source file) and `internal/template/templates/` for `research.md` / `mukul975` / `Anthropic-Cybersecurity` → expected zero matches.
