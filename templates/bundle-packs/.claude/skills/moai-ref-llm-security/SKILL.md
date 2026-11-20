---
name: moai-ref-llm-security
description: >
  AI/LLM defensive security reference: prompt-injection defense, OWASP LLM Top 10
  defensive mapping, MCP and agentic tool-call hardening, training-data poisoning
  detection, model-output validation and guardrails, MITRE ATLAS defensive
  correlation, and NIST AI RMF governance. Agent-extending skill that amplifies
  backend, security, and AI-application engineering with production-grade
  defensive patterns for LLM-backed systems.
  NOT for: offensive techniques (jailbreak authoring, attack-payload crafting,
  red-team exploitation), model training or fine-tuning methodology, prompt
  optimization for capability, web-app OWASP Top 10 (see moai-ref-owasp-checklist),
  or general API design (see moai-ref-api-patterns).

when_to_use: >
  Use when hardening an LLM-backed application or agent against prompt injection,
  designing guardrails or output validation, scoping MCP/tool permissions,
  detecting training-data or model poisoning, mapping a design against the OWASP
  LLM Top 10 or NIST AI RMF, or correlating defenses to MITRE ATLAS techniques.
  Loads as background knowledge for AI-security review, agentic-system design, and
  LLM-application hardening tasks.

user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-06-24"
  tags: "llm, ai-security, prompt-injection, guardrails, mcp, agentic, owasp-llm, mitre-atlas, nist-ai-rmf, reference"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000
---

# LLM / AI Defensive Security Reference

Defensive practitioner reference for hardening LLM-backed applications and agents.
Every section is framed as defense, hardening, detection, or verification — it
describes the misconfiguration, how to detect it, and how to prevent it, never how
to exploit it. Cross-domain web-app vulnerabilities live in
`moai-ref-owasp-checklist`; API design lives in `moai-ref-api-patterns`.

## Target Use

Apply when reviewing or building an LLM-backed system — a chat product, a
retrieval-augmented application, an autonomous agent, or an MCP server. The
material assumes an untrusted-input threat model: any text the model reads (user
turns, retrieved documents, tool results, file contents) may carry adversarial
instructions, and any text the model emits may be acted on downstream.

## Trust Boundaries in an LLM System

The core defensive insight: an LLM does not distinguish "data" from "instructions"
the way a parser does. Treat every text channel that reaches the model as a
boundary where adversarial instructions can enter.

| Channel | Entry risk | Primary defense |
|---------|-----------|-----------------|
| End-user prompt | Direct prompt injection | Instruction-hierarchy enforcement, input screening |
| Retrieved documents (RAG) | Indirect prompt injection | Provenance tagging, content isolation, retrieval allowlist |
| Tool / function results | Injected instructions in tool output | Treat tool output as untrusted data, re-validate before re-prompting |
| System / developer prompt | Leakage, override | Minimize secrets in prompt, assume prompt is recoverable |
| Model output | Improper downstream handling | Schema validation, encoding, never auto-execute raw output |
| Training / fine-tuning data | Poisoning | Data lineage, provenance verification, canary detection |

## OWASP LLM Top 10 — Defensive Mapping

The OWASP Top 10 for LLM Applications (2025 edition) is the canonical risk index
for LLM systems. Each row below states the risk, the defensive check, and the
hardening control. Citations are for defensive correlation only.

| ID | Risk | Defensive check | Hardening control |
|----|------|-----------------|-------------------|
| LLM01 | Prompt Injection | Can external text override system instructions? | Instruction-hierarchy enforcement, input/output screening, content isolation |
| LLM02 | Sensitive Information Disclosure | Can the model reveal secrets, PII, or system prompt? | Output filtering, minimize secrets in context, response redaction |
| LLM03 | Supply Chain | Are model weights, datasets, and plugins from verified sources? | Provenance verification, model/dataset signing, dependency pinning |
| LLM04 | Data and Model Poisoning | Could training/fine-tuning data carry adversarial content? | Data lineage, canary artifacts, anomaly detection on training sets |
| LLM05 | Improper Output Handling | Is model output executed/rendered without validation? | Schema validation, context-aware encoding, no auto-execution |
| LLM06 | Excessive Agency | Does the agent hold more capability/permission than needed? | Least-privilege tool design, human-in-loop for high-impact actions |
| LLM07 | System Prompt Leakage | Does the system prompt hold secrets that leak on extraction? | Never store secrets in the prompt; assume the prompt is recoverable |
| LLM08 | Vector and Embedding Weaknesses | Can embedding/RAG stores be poisoned or leak cross-tenant? | Per-tenant isolation, embedding-source validation, access control |
| LLM09 | Misinformation | Is unverified model output presented as authoritative? | Grounding, citation requirements, confidence signalling |
| LLM10 | Unbounded Consumption | Can a request exhaust tokens, cost, or compute? | Rate limiting, token budgets, request-size caps, timeout enforcement |

## Prompt-Injection Defense

Prompt injection is the highest-frequency LLM risk (LLM01). Defense is
defense-in-depth — no single control is sufficient, so layer the controls below.

### Direct injection (the user is the attacker)

| Layer | Control | Purpose |
|-------|---------|---------|
| Instruction hierarchy | Mark system instructions as highest-priority; instruct the model that downstream text cannot override them | Reduce override success rate |
| Input screening | Scan incoming prompts for known injection markers before they reach the model | Detect obvious override attempts |
| Privilege separation | Run a high-trust planning model separately from a low-trust task model that touches untrusted content | Limit blast radius of a successful injection |
| Output gating | Validate the model's response against an allowlist of expected shapes before acting on it | Block injected instructions from taking effect downstream |

### Indirect injection (the attacker poisons retrieved content)

Indirect injection arrives through documents, web pages, emails, or tool results
that the model reads. The defense is to treat all retrieved content as untrusted
data, never as instructions.

- **Provenance tagging**: wrap retrieved content in clearly delimited, labeled
  blocks so the model is instructed to treat the enclosed text as data to analyze,
  not commands to obey.
- **Content isolation**: keep retrieved content in a separate channel from the
  instruction channel; do not concatenate untrusted text directly after a system
  instruction.
- **Retrieval allowlist**: restrict RAG sources to verified corpora; reject or
  quarantine documents from unverified origins.
- **Re-validation after tool calls**: when a tool returns text that re-enters the
  prompt, re-screen it — a tool result is an untrusted channel, not a trusted one.

This class of attack maps to MITRE ATLAS **AML.T0051** (LLM Prompt Injection); the
defenses above are how you detect and prevent it, not an attack procedure.

## MCP / Agentic Tool-Call Hardening

Agentic systems that call tools (including MCP servers) expand the attack surface:
a successful injection can now drive real actions. The defense is least-privilege
tool design plus output re-validation (LLM06 Excessive Agency).

### Tool-permission scoping

| Control | Defensive rationale |
|---------|---------------------|
| Least-privilege tool set | Expose only the tools the task needs; an agent that cannot delete cannot be tricked into deleting |
| Scoped credentials per tool | A tool holds the minimum credential for its function, never a broad admin token |
| Human-in-the-loop gates | High-impact actions (payments, deletions, external posts, force-push) require explicit confirmation regardless of the model's confidence |
| Allowlist of callable targets | Constrain which hosts, paths, or resources a tool may reach; block internal/metadata endpoints |
| Idempotency + dry-run | Prefer reversible or preview-able actions; confirm before irreversible ones |

### Tool-output validation

Treat every tool result as untrusted input on its way back into the model:

- **Schema-validate tool results** before re-prompting — reject malformed or
  oversized payloads.
- **Strip or neutralize instruction-shaped content** in tool output (a tool that
  returns a web page may carry an injection in that page).
- **Bound tool-call loops** — cap the number of tool invocations per task to
  prevent an injected instruction from driving an unbounded action chain
  (LLM10 Unbounded Consumption).
- **Log every tool call** with its arguments and result for audit and anomaly
  detection.

This hardening defends against MITRE ATLAS **AML.T0053** (LLM Plugin Compromise)
class concerns by minimizing what a compromised tool path can reach.

## Training-Data and Model Poisoning Detection

Poisoning (LLM04) corrupts model behavior by inserting adversarial samples into
training or fine-tuning data, or by substituting a tampered model artifact. The
defenses are provenance and detection, not retraining recipes.

| Defense | What it detects | How to apply |
|---------|-----------------|--------------|
| Data lineage | Untracked or unverified data entering the training set | Record the source, hash, and acquisition path of every dataset |
| Canary artifacts | Whether a known marker sample influenced the model unexpectedly | Insert known canary records; verify the model's behavior on them post-training |
| Provenance verification | A substituted or tampered model/dataset | Verify cryptographic signatures on model weights and datasets before use |
| Distribution monitoring | Anomalous shifts in training-data statistics | Compare new data batches against the established distribution baseline |
| Source allowlisting | Data from unverified origins | Restrict training corpora to vetted, signed sources |

Poisoning maps to MITRE ATLAS **AML.T0020** (Poison Training Data); the detection
controls above are defensive correlation, never an attack method.

## Model-Output Validation and Guardrails

Improper output handling (LLM05) is when downstream systems trust raw model output.
The defense is to treat model output as untrusted until validated.

### Output validation layers

| Layer | Control | Prevents |
|-------|---------|----------|
| Structured output | Enforce a schema (typed object, constrained grammar) on the response | Free-form output carrying injected commands |
| Content filtering | Screen output for policy violations, leaked secrets, or disallowed content | Sensitive disclosure (LLM02), system-prompt leakage (LLM07) |
| Context-aware encoding | Encode output for its sink (HTML escape, shell-quote, SQL parameter) | Output-handling injection into a downstream interpreter |
| No auto-execution | Never feed raw model output directly into a shell, eval, or query | Remote code execution via the model |
| Grounding + citation | Require the model to cite sources for factual claims | Misinformation (LLM09) presented as authoritative |

### Guardrail placement

Guardrails belong on **both** sides of the model — an input guardrail screens what
enters, an output guardrail screens what leaves. An output-only guardrail misses
injection that has already changed the model's plan; an input-only guardrail misses
sensitive disclosure in the response. Layer both.

## MITRE ATLAS — Defensive Correlation

MITRE ATLAS catalogs adversarial techniques against AI systems. The skills cite
ATLAS technique IDs to correlate a defense with the technique it counters — never
to provide an attack recipe. Representative defensive correlations:

| ATLAS technique | Technique name | Defensive control that counters it |
|-----------------|----------------|------------------------------------|
| AML.T0051 | LLM Prompt Injection | Instruction-hierarchy enforcement, content isolation, input/output screening |
| AML.T0020 | Poison Training Data | Data lineage, canary artifacts, provenance verification |
| AML.T0054 | LLM Jailbreak | Output filtering, refusal-policy enforcement, response gating |
| AML.T0057 | LLM Data Leakage | Output redaction, minimize context secrets, per-tenant isolation |

For each technique, the skill states how to detect and defend, not how to execute.
Treating an ATLAS technique ID as an attack recipe is the anti-pattern this skill
explicitly forbids.

## NIST AI RMF — Governance Mapping

The NIST AI Risk Management Framework (AI RMF 1.0) organizes AI risk into four
functions. Map LLM-security controls onto them for defensive governance:

| Function | Defensive activity |
|----------|--------------------|
| GOVERN | Establish AI-security policy, accountability, and incident-response ownership for LLM systems |
| MAP | Inventory LLM components, data sources, tool permissions, and trust boundaries; identify the threat model |
| MEASURE | Test for prompt-injection resistance, output-validation coverage, and guardrail effectiveness; track findings |
| MANAGE | Prioritize and remediate identified risks; monitor production behavior; maintain the response plan |

The four functions give a defensive governance scaffold; the OWASP LLM Top 10 and
MITRE ATLAS supply the specific risks to map, measure, and manage.

## Cross-References

- `moai-ref-owasp-checklist` — web-application OWASP Top 10, authentication
  patterns, input validation, HTTP security headers (the dev-time application
  security surface; this skill is the LLM-specific surface).
- `moai-ref-api-patterns` — REST/GraphQL API design, error handling, rate limiting
  (the API-design surface that an LLM-backed service exposes).
- `moai-ref-supply-chain` — dependency and artifact provenance, signing, SBOMs
  (the broader supply-chain surface that LLM03/LLM04 model and data provenance
  draws on).

## Defensive Severity Levels

| Level | Label | Action | Example |
|-------|-------|--------|---------|
| P0 | CRITICAL | Block release | Raw model output executed in a shell; no output validation on an agent with destructive tools |
| P1 | HIGH | Fix before merge | No instruction-hierarchy enforcement on a system handling untrusted documents |
| P2 | MEDIUM | Fix within iteration | Tool-call loop has no upper bound; missing input screening |
| P3 | LOW | Track in backlog | System prompt slightly over-discloses internal structure (no secret leaked) |

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "The model is smart enough to ignore injected instructions" | Models do not reliably separate data from instructions. A capable model can still be steered by injected text; defense-in-depth, not model capability, is the control. |
| "We only read internal documents, so RAG content is trusted" | Internal documents can be edited by a compromised account or an upstream feed. Indirect injection does not require an external attacker — treat all retrieved content as untrusted. |
| "Output validation slows the response, users will not notice the risk" | Unvalidated output is how model responses become shell commands and XSS. Validation is the boundary between a chat reply and remote code execution. |
| "The agent needs broad permissions to be useful" | Excessive agency is an OWASP LLM Top 10 risk. Scope tools to the task; a capability the agent does not hold cannot be abused via injection. |
| "We put guardrails on the output, that is enough" | An output-only guardrail misses injection that already changed the model's plan and tool calls. Guardrails belong on both input and output. |
| "Prompt injection is a research problem, not a production one" | Indirect prompt injection is exploited in production RAG and agentic systems today. It is a current operational risk, not a theoretical one. |

**Untrusted-by-default**: every text channel reaching the model — user input,
retrieved documents, tool results — is untrusted until screened. The model is not
a trust boundary; your validation layer is.

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Model output fed directly into a shell, eval, query, or `innerHTML` without schema validation or encoding
- Retrieved RAG content concatenated directly after a system instruction with no isolation or provenance tagging
- An agent holds destructive or high-privilege tools with no human-in-the-loop gate on high-impact actions
- Tool results re-enter the prompt without being re-screened as untrusted input
- Secrets, API keys, or credentials embedded in the system prompt (assume the prompt is recoverable)
- No upper bound on tool-call loops or token consumption per request
- Guardrails present on only one side (input-only or output-only) of the model

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Every text channel reaching the model is classified as trusted or untrusted, and untrusted channels are screened
- [ ] System instructions use instruction-hierarchy enforcement against override by downstream text
- [ ] Retrieved/RAG content is isolated and provenance-tagged, not concatenated after instructions
- [ ] Model output is schema-validated and context-encoded before any downstream use; nothing auto-executes raw output
- [ ] Agent tools follow least-privilege scoping; high-impact actions have a human-in-the-loop gate
- [ ] Tool results are re-validated as untrusted input before re-prompting; tool-call loops are bounded
- [ ] The design is mapped against the OWASP LLM Top 10 (show which items were evaluated) and governed under the NIST AI RMF functions
- [ ] No secrets reside in the system prompt; output filtering screens for sensitive disclosure

<!-- moai:evolvable-end -->
