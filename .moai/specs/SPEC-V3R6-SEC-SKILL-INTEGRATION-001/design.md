# design.md — SPEC-V3R6-SEC-SKILL-INTEGRATION-001

> **Tier L 5th artifact.** This document records the design decisions that govern the three defensive cybersecurity reference skills' authoring. It is the canonical reference for the re-authorship procedure, the legal-integrity rationale, the upstream-containment boundary, and the dual-use defensive-framing policy. Run-phase authors (manager-develop M1-M4) and future maintainers consult this file before touching the skill bodies.
>
> **Why this file exists (plan-auditor iter-1 BLOCKING D7)**: the initial SPEC classified itself Tier M (3-artifact set: spec/plan/acceptance). The plan-auditor flagged the absent `tier:` frontmatter field, which defaults to Tier L. The re-iteration upgraded explicitly to Tier L based on: (a) the re-authorship pattern is novel in moai-adk (first-of-kind independent re-authorship from a named upstream with an internal-only research.md); (b) dual-use legal-integrity is a P0 surface; (c) 16-language neutrality audit overlaps. Tier L mandates design.md as the 5th artifact documenting the procedure + rationale so future SPECs following the same pattern can reference this one as the canonical precedent.

---

## §A. Re-authorship Procedure (the concept-read → close → write-fresh pattern)

The three skills are authored by **independently re-authoring** defensive-cybersecurity knowledge from a named community upstream repository. The upstream is used ONLY as a **knowledge/procedure reference source** — concepts, tool names, technique IDs, procedure ordering. No upstream sentences, paragraphs, or script code are copied.

### §A.1 The 5-step procedure (binding on run-phase authors)

[ZONE:Evolvable] [HARD] Each defensive topic in the three skills is authored via the following 5-step procedure. Deviation from this procedure is AP-SI-001 (copy-paste from upstream) — a legal-integrity breach.

| Step | Action | Mechanical check |
|------|--------|------------------|
| 1. **Concept-read** | Open the upstream skill for the topic. Extract the concept outline: which tool, which order, which verification step, which technique ID. Do NOT copy sentences. Record the concept label + public-standard reference (OWASP / MITRE / NIST page) in research.md. | research.md §B concept-inventory table populated |
| 2. **Close the upstream** | Before writing the moai-adk skill body, close the upstream file. The author writes from the concept, not from the upstream text. | (discipline; no mechanical check at this step) |
| 3. **Write fresh in MoAI-ADK style** | Use the existing `moai-ref-owasp-checklist` / `moai-ref-api-patterns` body style — markdown tables, checklists, verification sections, the three evolvable blocks (rationalizations / red-flags / verification). Cite the public-standard reference (OWASP page, MITRE page, NIST pub, vendor advisory) at the concept level, NOT the upstream. | skill body follows `skill-authoring.md` + `skill-writing-craft.md` |
| 4. **Defensive-framing check** | Each section frames content as defense / hardening / detection / verification — never as exploitation. Where a topic inherently names an attack vector (e.g., container-escape defense names the escape), describe the misconfiguration that enables it + how to detect it + how to prevent it, omitting any step-by-step exploitation procedure. | E-DUALUSE grep (AC-SI-009) + defensive-verb proxy |
| 5. **Spot-check at M4** | Select ≥3 passages per skill (≥3 sentences each); compare against the upstream; confirm no verbatim reproduction beyond standard tool-invocation syntax. Debt D1: raise to ≥10 passages/skill OR add mechanical full-sentence grep. | E-REAUTHOR (AC-SI-007) |

### §A.2 Why the procedure works (the legal basis)

The procedure produces an **independent work**, not a derivative work, because:

- **Ideas, procedures, methods, and facts are not copyrightable** (17 USC §102(b) — United States; Korean Copyright Act §7 — "이 법에 의하여 보호되는 저작물은 다음 각호의 1에 해당하는 것을 제외한다. ... 아이디어, 규칙, 발견, 원리, 조작방법"). The defensive cybersecurity knowledge (which tool, which command, which order, which verification, MITRE technique IDs, CVE numbers, command syntax) is overwhelmingly facts/procedures/standards and is freely referenceable.
- **Apache-2.0 §4(c) attribution duty applies only to derivative works.** An independently re-authored work that does not reproduce upstream expression is NOT a derivative work; no attribution duty attaches. The moai-adk skills do not reproduce upstream expression; therefore Apache-2.0 §4(c) does not bind them.
- **The ONLY legal duty is: do not copy upstream verbatim text/scripts.** This is the HARD constraint enforced by REQ-SI-004 + AC-SI-007 + E-REAUTHOR.

### §A.3 Standard tool-invocation syntax exemption

A short procedural sentence (e.g., "run `cosign verify --key cosign.pub imageref`") may coincidentally match the upstream's wording because the tool's standard invocation is functionally fixed. Functional syntax is NOT protectable expression (it is part of the procedure/method category of §102(b)). AC-SI-007 explicitly exempts "standard tool-invocation syntax (e.g., `cosign verify --key ...`)" from the verbatim-reproduction check. The reviewer distinguishes functional syntax from expressive prose.

### §A.4 Anti-pattern: copy-paste-with-light-editing (AP-SI-001)

The canonical anti-pattern is copying the upstream's prose/script into the moai-adk skill body and lightly editing it. This IS a derivative work and triggers Apache-2.0 §4(c) attribution duty + the §25 template-neutrality breach (upstream's stylistic choices would leak into the distributed artifact). The correct pattern is: read upstream for concepts, close it, write fresh.

---

## §B. Legal-Integrity Rationale (17 USC §102(b) / Korean Copyright Act §7 / Apache-2.0 §4(c))

This section is the canonical legal-integrity argument. A future maintainer (or an external auditor) reads this section to verify the legal basis for the re-authorship pattern without re-deriving it from first principles.

### §B.1 The three legal pillars

| Pillar | Citation | Application to this SPEC |
|--------|----------|--------------------------|
| **Ideas/procedures/facts are not copyrightable** | 17 USC §102(b) (United States Code, Title 17, §102(b)): "In no case does copyright protection for an original work of authorship extend to any idea, procedure, process, system, method of operation, concept, principle, or discovery..." | The defensive cybersecurity knowledge (which tool, which command, which order, which verification, MITRE technique IDs, CVE numbers) is ideas/procedures/processes/methods/concepts — not protectable expression. The moai-adk skills reference these freely. |
| **Korean equivalent** | Korean Copyright Act (저작권법) §7 (works not protected): "이 법에 의하여 보호되는 저작물은 다음 각호의 1에 해당하는 것을 제외한다. ... 3. 아이디어, 규칙, 발견, 원리, 조작방법" (ideas, rules, discoveries, principles, methods of operation) | Same as §102(b) — ideas/procedures/methods are not protectable. Cited because the maintainer jurisdiction is Korea. |
| **Apache-2.0 §4(c) binds derivatives only** | Apache License 2.0 §4(c): "You must retain, in the Source form of any Derivative Works that You distribute, all copyright, patent, trademark, and attribution notices from the Source form of the Work..." | The attribution duty attaches ONLY to Derivative Works. An independently re-authored work that does not reproduce upstream expression is NOT a derivative work; the duty does not attach. |

### §B.2 The derivative-work test

A work is a "derivative work" (17 USC §101) if it is "based upon one or more preexisting works, such as a translation, musical arrangement, dramatization, fictionalization, motion picture version, sound recording, art reproduction, abridgment, condensation, or any other form in which a work may be recast, transformed, or adapted." The key phrase is **"recast, transformed, or adapted"** — this requires reproduction of protectable expression.

The moai-adk skills:
- Do NOT reproduce upstream sentences, paragraphs, or scripts (REQ-SI-004, AC-SI-007).
- DO reference the same public-standard facts (MITRE technique IDs, OWASP labels, CVE numbers, tool names, command syntax) — but facts/procedures/standards are not protectable expression per §102(b).
- DO use MoAI-ADK's own stylistic choices (markdown tables, checklists, evolvable blocks) — not the upstream's.

Therefore the moai-adk skills are NOT derivative works; Apache-2.0 §4(c) attribution duty does NOT attach; the skills carry NO upstream attribution (REQ-SI-002, REQ-SI-021, AC-SI-008).

### §B.3 The upstream is recorded for maintainer traceability — NOT for attribution

The upstream repository (`mukul975/Anthropic-Cybersecurity-Skills`) is recorded in the internal research.md as the **initial knowledge-inspiration source** for maintainer traceability (REQ-SI-020). This is NOT Apache-2.0 §4(c) attribution — it is an internal provenance record that the skills were independently re-authored (not copied) from a known upstream, so a future maintainer can re-evaluate upstream changes against MoAI-ADK's quality bar. The research.md is internal-only and never distributed (AC-SI-018 enforces containment).

### §B.4 What would change the legal analysis

If a future run-phase author copies upstream verbatim text/scripts into a skill body, that skill becomes a derivative work and Apache-2.0 §4(c) attribution duty attaches retroactively. This is why AC-SI-007 (re-authorship integrity) is a P0 gate and E-REAUTHOR is a mandatory run-phase self-check. The legal analysis in this section PRESUMES the procedure in §A is followed; if it is not followed, the analysis no longer holds and the skills must either (a) add upstream attribution (violating REQ-SI-002's no-attribution rule, forcing a scope reduction to remove the copied material) or (b) be re-authored per §A.

---

## §C. Upstream-Containment Boundary (research.md is internal-only, never embedded)

### §C.1 The containment invariant

[ZONE:Evolvable] [HARD] The internal research.md (`/Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/research.md`) is the SOLE location in the repository where the upstream repository is identified by name. It records the upstream as the knowledge-inspiration source for maintainer traceability. It is NOT distributed with the skills — it lives under `.moai/specs/` (a moai-adk-internal SPEC directory), NOT under `.claude/skills/` (the distributed skill directory) or `internal/template/templates/.claude/skills/` (the template-mirror distributed directory).

### §C.2 Why the containment matters

- **AC-SI-008 scope**: the no-attribution grep covers the DISTRIBUTED paths only (`.claude/skills/moai-ref-{llm-security,supply-chain,secops}/` and `internal/template/templates/.claude/skills/moai-ref-{llm-security,supply-chain,secops}/`). It does NOT cover `.moai/specs/`. The research.md names the upstream verbatim — this is SAFE because research.md is never embedded in the template binary and never ships to user projects.
- **AC-SI-018 containment check**: the new P0 AC greps `internal/template/embed.go` (the Go `//go:embed all:templates` source file) and `internal/template/templates/` (the template source tree) for `research.md` / `mukul975` / `Anthropic-Cybersecurity` → expected: zero matches. This verifies mechanically that the research.md is never embedded in the `moai init` / `moai update` binary payload.
- **E-CONTAIN self-check**: the run-phase §E self-verification (plan.md §E) includes E-CONTAIN, the per-skill grep that operationalizes AC-SI-018.

### §C.3 The dual-directory discipline

| Directory | Distributed? | Carries upstream name? | Grep exclusion |
|-----------|--------------|------------------------|----------------|
| `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/research.md` | NO (internal SPEC dir) | YES (records upstream verbatim for maintainer traceability) | Excluded from AC-SI-008 / E-NOATTRIB grep |
| `.claude/skills/moai-ref-{llm-security,supply-chain,secops}/SKILL.md` | YES (dev-project skill) | NO | IN AC-SI-008 / E-NOATTRIB grep scope |
| `internal/template/templates/.claude/skills/moai-ref-{llm-security,supply-chain,secops}/SKILL.md` | YES (template mirror → ships via `moai init`/`moai update`) | NO | IN AC-SI-008 / E-NOATTRIB grep scope |
| `internal/template/embed.go` + `internal/template/templates/` | YES (template binary + source) | NO (research.md never embedded) | IN AC-SI-018 / E-CONTAIN grep scope |

### §C.4 Template-parity invariant

The local skill files (`.claude/skills/moai-ref-*/SKILL.md`) and the template-mirror files (`internal/template/templates/.claude/skills/moai-ref-*/SKILL.md`) MUST be byte-identical (AC-SI-012). This ensures the dev-project experience and the user-project experience (via `moai init`/`moai update`) are the same. The template-mirror is the source of truth for what ships; the local copy is the source of truth for what runs in dev; they must match.

---

## §D. Dual-Use Defensive-Framing Policy

### §D.1 The defensive-only invariant

[ZONE:Evolvable] [HARD] The three skills cover defensive practitioner playbooks ONLY — detection, hardening, response, verification, and defense-in-depth guidance. Offensive procedures (Red Teaming execution, C2 framework operation, Pen Testing exploitation, credential dumping) are EXCLUDED entirely (REQ-SI-005, REQ-SI-006, spec.md §F Out-of-Scope).

### §D.2 The defensive-framing rule for dual-use topics

Many defensive topics inherently name an attack vector — "container-escape defense" must mention that the escape exists; "Mimikatz-style credential dumping detection" must name Mimikatz. The defensive-framing rule (REQ-SI-007) governs how these topics are presented:

- **Frame as defense/hardening/detection/verification** — never as exploitation.
- **Describe the misconfiguration that enables the attack** (here is the K8s RBAC misconfiguration that enables container escape).
- **Describe how to detect it** (here is how Falco/runtime-detection surfaces the escape attempt).
- **Describe how to prevent it** (here is how seccomp/AppArmor/read-only-root-filesystem/non-root-user prevents the escape).
- **OMIT any step-by-step exploitation procedure** (no "run this command to escape the container").

### §D.3 The defensive-verb proxy (mechanical check for AC-SI-009)

Because a keyword grep for offensive tool names (Mimikatz, BloodHound, Rubeus, etc.) is syntactic — it cannot distinguish "Mimikatz-style credential dumping is defended against by enforcing LSASS protection" (defensive framing, acceptable) from "run Mimikatz with these flags to dump credentials" (exploitation recipe, FAIL) — AC-SI-009 adds a **defensive-verb proxy**:

> Each offensive-keyword match MUST appear within 1 sentence of a defensive verb (`detect | prevent | defend | block | alert | scan | harden | monitor | triage | investigate`).

This proxy is necessary-but-not-sufficient: a re-framed offense that evades both the keyword list and the defensive-verb proximity (e.g., novel phrasing) can still slip through. The residual semantic-coverage gap is recorded as run-phase M4 residual risk (debt D5/D6) — the M4 human verdict is the final gate, not the grep.

### §D.4 MITRE ATT&CK / ATLAS / OWASP technique IDs as defensive correlation

Technique IDs (MITRE ATT&CK T-IDs, ATLAS technique IDs, OWASP attack-pattern labels, CVE numbers) are public-standard identifiers and MAY be cited freely for defensive correlation (REQ-SI-008). The skills cite them as "this hardening step defends against technique T1xxx" — NOT as "here is how to execute technique T1xxx." The framing is load-bearing: AP-SI-007 (treating MITRE IDs as attack recipes) is the anti-pattern.

### §D.5 Excluded categories (the hard boundary)

The following upstream categories are EXCLUDED from re-authoring entirely (research.md §D, spec.md §F Out-of-Scope):

- Red Teaming execution playbooks (recon, initial-access, lateral-movement procedures)
- C2 framework operational guides (Cobalt Strike, Sliver, Mythic operation)
- Penetration Testing exploitation execution (Metasploit modules, exploit chaining)
- Credential dumping (Mimikatz usage, Kerberoasting, NTLM-relay attack steps, pass-the-hash, pass-the-ticket, LSASS dumping)
- Any step-by-step exploitation procedure for a technique cited defensively

A defensive topic that cannot be framed without including exploitation material is removed from scope and recorded as a follow-up SPEC candidate (EC2).

---

## §E. Design Decisions Log

| Decision | Chosen | Rejected alternatives | Rationale |
|----------|--------|----------------------|-----------|
| Re-author vs vendor upstream | Independent re-authorship (Option B') | Direct vendoring (Option A) — too large attribution surface + imports upstream style + pulls dual-use offensive content | Re-authorship produces MoAI-ADK's own work; no attribution duty; defensive-only subset is cleanly selectable |
| Three-skill split | LLM / Supply Chain / SecOps (DevSecOps+Container+API) | Single skill (too broad, >500 lines); five skills (too narrow, overlapping vocab); two skills (supply-chain dilutes into SecOps) | Matches defensive-practitioner mental model; natural domain boundaries; description-trigger precision |
| research.md internal-only | Single internal file under `.moai/specs/` records upstream verbatim | (a) embed upstream name in skill body (fails AC-SI-008), (b) omit upstream entirely (loses maintainer traceability), (c) ship research.md with skills (fails AC-SI-018 containment) | Maintainer traceability + legal-integrity audit trail without distributing the upstream name |
| Defensive-only subset | Exclude offensive categories outright | Include with warnings (normalizes offensive content in a defensive tool) | MoAI-ADK is a defensive practitioner tool; distributing offensive procedures is a safety breach |
| 400-line HARD split threshold for secops | HARD 400 lines → split into modules | Tilde ~500 (ambiguous, invites over-cut before splitting) | The secops skill is the largest (~79 upstream topics); a HARD threshold forces the split decision before the file becomes unmanageable. The ~150-line overview + 3 modules pattern keeps skill-listing matcher precision high. |
| De-facto-standard tool carve-out (REQ-SI-013) | Name de-facto cross-ecosystem standards (cosign, trivy/grype, semgrep) framed as "the standard <category> tool" | (a) forbid all tool names (impractical — cosign has no real alternative), (b) allow all tool names (privileges one ecosystem) | The carve-out reflects reality (some tools ARE the cross-ecosystem standard) while keeping the framing ecosystem-neutral |

---

## §F. Cross-References

- **spec.md** — `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/spec.md` (§C requirements, §D constraints legal-integrity summary, §F Out-of-Scope, §H HISTORY recording the Tier-L upgrade)
- **plan.md** — `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/plan.md` §B.4 (Tier-L justification), §E (E-REAUTHOR / E-DUALUSE / E-LANGNEUTRAL / E-NOATTRIB / E-CONTAIN self-checks), §F (M1-M5 milestones)
- **acceptance.md** — `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/acceptance.md` (AC-SI-007 re-authorship, AC-SI-008 no-attribution + scope, AC-SI-009 dual-use + defensive-verb proxy, AC-SI-010 neutrality positive reframe, AC-SI-013 wc -c measurement, AC-SI-014 400-line HARD threshold, AC-SI-018 research.md containment)
- **research.md** — `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/research.md` (internal-only — §A names upstream verbatim, §B concept inventory, §C re-authorship procedure, §D excluded categories)
- **Legal authorities**: 17 USC §102(b) (ideas/procedures/facts not copyrightable); Korean Copyright Act §7 (저작권법 제7조); Apache License 2.0 §4(c) (attribution duty binds derivatives only)
- **Structural template skills**: `.claude/skills/moai-ref-owasp-checklist/SKILL.md`, `.claude/skills/moai-ref-api-patterns/SKILL.md`, `.claude/skills/moai-ref-testing-pyramid/SKILL.md`
- **Template-neutrality doctrine**: `.moai/docs/template-internal-isolation-doctrine.md` §25.1 (C1-C8 content-class catalogue); CLAUDE.local.md §25
- **16-language neutrality**: CLAUDE.local.md §15

---

Version: 0.2.0 (plan-auditor iter-1 re-iteration — Tier L upgrade, BLOCKING D7 resolved)
Status: Active — binds run-phase M1-M4 authoring and future maintainers
