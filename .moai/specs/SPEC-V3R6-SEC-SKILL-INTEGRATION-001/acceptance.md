# acceptance.md — SPEC-V3R6-SEC-SKILL-INTEGRATION-001

> Acceptance criteria for the three defensive cybersecurity reference skills.
> AC severity convention: P0 = MUST (blocker), P1 = SHOULD (fix before merge), P2 = NICE (track in backlog).

## §A. AC Matrix Overview

| AC ID | Subject | Severity | Milestone | Maps to REQ |
|-------|---------|----------|-----------|-------------|
| AC-SI-001 | Skill 1 exists with valid frontmatter | P0 | M1 | REQ-SI-009, REQ-SI-010, REQ-SI-012 |
| AC-SI-002 | Skill 1 body covers LLM defensive topics | P0 | M1 | REQ-SI-008, REQ-SI-011, REQ-SI-016 |
| AC-SI-003 | Skill 2 exists with valid frontmatter | P0 | M2 | REQ-SI-009, REQ-SI-010, REQ-SI-012 |
| AC-SI-004 | Skill 2 body covers supply-chain defensive topics | P0 | M2 | REQ-SI-011, REQ-SI-017 |
| AC-SI-005 | Skill 3 exists with valid frontmatter | P0 | M3 | REQ-SI-009, REQ-SI-010, REQ-SI-012 |
| AC-SI-006 | Skill 3 body covers SecOps defensive topics | P0 | M3 | REQ-SI-011, REQ-SI-018, REQ-SI-019 |
| AC-SI-007 | Re-authorship integrity (no verbatim upstream) | P0 | M4 | REQ-SI-001, REQ-SI-003, REQ-SI-004, REQ-SI-022 |
| AC-SI-008 | No upstream attribution in distributed artifacts | P0 | M4 | REQ-SI-002, REQ-SI-021 |
| AC-SI-009 | Dual-use boundary (no offensive procedures) | P0 | M4 | REQ-SI-005, REQ-SI-006, REQ-SI-007, REQ-SI-008 |
| AC-SI-010 | 16-language neutrality (positive multi-ecosystem balance) | P1 | M4 | REQ-SI-013 |
| AC-SI-011 | Template-neutrality §25 (no internal IDs leak) | P0 | M4 | REQ-SI-014 |
| AC-SI-012 | Template-mirror parity (local == template) | P0 | M4 | (structural) |
| AC-SI-013 | Skill-listing matcher descriptions within cap | P1 | M4 | REQ-SI-010 |
| AC-SI-014 | Skill size — HARD 400-line secops split threshold | P1 | M3 | (structural) |
| AC-SI-015 | Cross-references to existing moai-ref skills | P2 | M4 | REQ-SI-015 |
| AC-SI-016 | Verification + red-flags + rationalizations blocks present | P1 | M4 | REQ-SI-011 |
| AC-SI-017 | Closure gate (3-phase close) | P0 | M5 | (closure) |
| AC-SI-018 | research.md containment (never embedded in template binary) | P0 | M4 | REQ-SI-020, REQ-SI-023 |

## §B. Given-When-Then Scenarios

### AC-SI-001 — Skill 1 (`moai-ref-llm-security`) exists with valid frontmatter

**Severity**: P0  **Milestone**: M1

**Given** the `.claude/skills/moai-ref-llm-security/SKILL.md` file and its template mirror `internal/template/templates/.claude/skills/moai-ref-llm-security/SKILL.md` both exist,
**When** a reviewer parses the frontmatter,
**Then** all of the following hold:
- `name: moai-ref-llm-security` is present
- `description:` is a non-empty YAML folded scalar ≤ 1,536 chars combined with `when_to_use`
- `user-invocable: false`
- `metadata.version`, `metadata.category`, `metadata.status`, `metadata.updated`, `metadata.tags` are all present and string-typed
- the `progressive_disclosure` MoAI-extension block is present with `enabled: true`

**Evidence**: frontmatter dump + `moai spec lint` clean + the skill-authoring lint pass.

### AC-SI-002 — Skill 1 body covers LLM defensive topics

**Severity**: P0  **Milestone**: M1

**Given** the `moai-ref-llm-security/SKILL.md` body,
**When** a reviewer scans the body for topic coverage,
**Then** the body covers at least the following defensive topics (each framed as defense/hardening/detection, not exploitation):
- prompt-injection defense (system-prompt isolation, input sanitization, output validation)
- LLM application hardening (OWASP LLM Top 10 defensive mapping — at least 5 of the 10)
- MCP / agentic tool-call hardening (tool-permission scoping, output-validation on tool results)
- training-data poisoning detection (data-lineage, canary artifacts)
- model-output validation (guardrails, structured-output enforcement, content filtering)
- MITRE ATLAS defensive correlation (at least 3 technique IDs cited for defense, not attack) — this bullet operationalizes REQ-SI-008 (cite technique by public-standard ID for defensive correlation without providing an attack recipe); the "for defense, not attack" qualifier is the REQ-SI-008 invariant.
- NIST AI RMF defensive mapping (at least the Govern/Map/Measure/Manage functions referenced)

**Evidence**: the body itself + a topic-coverage table mapping each defensive topic to the body section that covers it.

### AC-SI-003 — Skill 2 (`moai-ref-supply-chain`) exists with valid frontmatter

**Severity**: P0  **Milestone**: M2

**Given** the `.claude/skills/moai-ref-supply-chain/SKILL.md` file and its template mirror both exist,
**When** a reviewer parses the frontmatter,
**Then** the same frontmatter shape as AC-SI-001 holds (with `name: moai-ref-supply-chain`).

**Evidence**: frontmatter dump + lint pass.

### AC-SI-004 — Skill 2 body covers supply-chain defensive topics

**Severity**: P0  **Milestone**: M2

**Given** the `moai-ref-supply-chain/SKILL.md` body,
**When** a reviewer scans the body for topic coverage,
**Then** the body covers at least:
- SBOM generation and verification (ecosystem-neutral SBOM tool naming, SPDX/CycloneDX formats)
- dependency-confusion defense (namespacing, private-registry scoping, install-time verification)
- malicious-package triage playbook (signals to look for, quarantine procedure, upstream-reporting channel)
- SLSA provenance levels (the 4 levels, what each guarantees, how to verify provenance at each level)
- Sigstore / cosign signing (keyless signing, identity-based verification, signature-verification at install time)
- package-registry hardening (registry-side signing, mirror-pinning, two-factor admin)
- typosquatting defense (dependency-name verification, allowlist, automated typo-detection tooling)
- transitive-dependency auditing (depth limits, license/vuln scanning, lockfile-pinning)

**Evidence**: the body + topic-coverage table.

### AC-SI-005 — Skill 3 (`moai-ref-secops`) exists with valid frontmatter

**Severity**: P0  **Milestone**: M3

**Given** the `.claude/skills/moai-ref-secops/SKILL.md` file and its template mirror both exist,
**When** a reviewer parses the frontmatter,
**Then** the same frontmatter shape as AC-SI-001 holds (with `name: moai-ref-secops`).

**Evidence**: frontmatter dump + lint pass.

### AC-SI-006 — Skill 3 body covers SecOps defensive topics

**Severity**: P0  **Milestone**: M3

**Given** the `moai-ref-secops/SKILL.md` body,
**When** a reviewer scans the body for topic coverage,
**Then** the body covers at least the following three sub-domains (each framed defensively):
- **DevSecOps**: CI/CD pipeline hardening (secret-scanning, signed-pipeline, runner-isolation), IaC scanning (Terraform/CloudFormation misconfiguration detection), SAST/DAST integration
- **Container**: image scanning (base-image hygiene, layer-scanning, admission-control), K8s RBAC hardening (least-privilege RoleBindings, ServiceAccount scoping, PodSecurity admission), container-escape defense (seccomp, AppArmor, read-only root filesystem, non-root user), runtime detection (Falco-rule shape, not Falco-as-sole-tool)
- **API operational**: OWASP API Top 10 operational defense (BOLA detection in production, broken-auth detection, rate-limit enforcement, server-side authorization) — operationally distinct from `moai-ref-owasp-checklist` which covers dev-time patterns; WAF rule-tuning (rule-shape, not vendor-specific); GraphQL/REST depth/complexity limiting

**Evidence**: the body + topic-coverage table.

### AC-SI-007 — Re-authorship integrity (no verbatim upstream reproduction)

**Severity**: P0  **Milestone**: M4

**Given** the three finished skill bodies and access to the upstream repository,
**When** the reviewer selects 3 random prose passages (each ≥ 3 sentences) from each skill body and compares them against the upstream,
**Then** none of the 9 passages is a verbatim or near-verbatim reproduction of upstream expression. Standard tool-invocation syntax (e.g., `cosign verify --key ...`) is exempt because it is functional, not expressive.

**Evidence**: the 9 passage-comparison table (moai-adk passage | closest upstream passage | verdict: independent / verbatim / standard-syntax-exempt). The research.md file records the upstream as inspiration source.

### AC-SI-008 — No upstream attribution in distributed artifacts

**Severity**: P0  **Milestone**: M4

**Given** the six distributed skill artifacts (3 SKILL.md under `.claude/skills/moai-ref-{llm-security,supply-chain,secops}/` + 3 SKILL.md under `internal/template/templates/.claude/skills/moai-ref-{llm-security,supply-chain,secops}/`),
**When** the reviewer greps the DISTRIBUTED paths (NOT `.moai/specs/`) for the upstream repository's name, GitHub path, maintainer handle, and the agentskills.io standard reference,
**Then** zero matches are found.

**Scope boundary (plan-auditor iter-1 D2 inversion)**: the upstream is named verbatim in the internal research.md (`.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/research.md` §A) as `mukul975/Anthropic-Cybersecurity-Skills` + its `agentskills.io` standard reference. This is SAFE because this AC's grep scope covers the DISTRIBUTED paths only — `.claude/skills/moai-ref-*/` and `internal/template/templates/.claude/skills/moai-ref-*/`. The internal `.moai/specs/` directory is OUT of scope for this AC (it is the SOLE location where the upstream name appears, per design.md §C upstream-containment boundary). AC-SI-018 separately verifies that research.md is never embedded in the template binary, closing the containment loop.

**Evidence** — the literal grep command + zero-match output:

```bash
grep -rniE 'mukul975|anthropic-cybersecurity-skills|agentskills\.io' \
  .claude/skills/moai-ref-{llm-security,supply-chain,secops}/ \
  internal/template/templates/.claude/skills/moai-ref-{llm-security,supply-chain,secops}/
# expected: zero matches
```

Additional grep patterns the reviewer SHOULD include: the upstream GitHub URL prefix (`github\.com/mukul975`), the upstream maintainer's GitHub handle, and attribution-text patterns ("based on", "adapted from", "inspired by" followed by the upstream name). Generic Apache-2.0 license text in `LICENSE` files is fine; attribution-text in skill bodies is NOT.

### AC-SI-009 — Dual-use boundary (no offensive procedures)

**Severity**: P0  **Milestone**: M4

**Given** the three finished skill bodies,
**When** the reviewer greps (case-insensitive `-i` flag) for the excluded offensive keywords — the expanded list:
`mimikatz`, `kerberoast`, `ntlm relay`, `ntlm-relay`, `cobalt strike`, `sliver`, `mythic`, `metasploit`, `pass-the-hash`, `pass-the-ticket`, `lsass dump`, `lsass-dump`, `credential dump`, `c2 framework`, `command and control`, `red team`, `pen test`, `penetration test`, `initial access`, `lateral movement`, `privilege escalation` (as an attack procedure), **plus** the additionally-enumerated offensive tools: `bloodhound`, `rubeus`, `impacket`, `evil-winrm`, `responder`, `seatbelt`, `sharpersist`, `certipy`,
**Then** the result is one of:
- (a) zero matches — preferred; OR
- (b) matches ONLY in defensive-framing sentences — acceptable, but minimized.

**Defensive-verb proxy (mechanical check)**: each offensive-keyword match MUST appear within 1 sentence of a defensive verb — `detect | prevent | defend | block | alert | scan | harden | monitor | triage | investigate`. A match that does NOT appear near a defensive verb is suspect and requires human verdict; a match that provides step-by-step exploitation is a FAIL.

A match that provides step-by-step exploitation is a FAIL regardless of defensive-verb proximity.

**REQ-SI-008 coverage (plan-auditor iter-2 D13)**: this AC also verifies REQ-SI-008 — any technique cited by public-standard ID (MITRE ATT&CK T-ID, OWASP attack-pattern label, CVE number) MUST be cited for defensive correlation ("this hardening defends against T1xxx") and MUST NOT be accompanied by an attack recipe. A skill body that cites a technique ID AND provides the exploitation procedure for that same technique is a FAIL under this AC via REQ-SI-008, even if no offensive keyword from the list above appears. The combination of (technique-ID citation + attack recipe) is the REQ-SI-008 violation shape this AC catches in addition to the keyword-list grep.

**Evidence**: the grep command (with `-i` flag explicit) + the per-match verdict table (keyword | surrounding sentence | nearest defensive verb | verdict: defensive-framing-acceptable / suspect-human-verdict / exploit-recipe-FAIL).

**Residual risk (debt D5/D6)**: the keyword grep is syntactic; a re-framed offense evading both the keyword list and the defensive-verb proximity can still slip through. The M4 human verdict is the final gate, not the grep — the grep is necessary-but-not-sufficient. This residual risk is recorded for run-phase M4 and not fixable in plan-phase.

### AC-SI-010 — 16-language neutrality (positive multi-ecosystem balance)

**Severity**: P1  **Milestone**: M4

**Given** the three finished skill bodies,
**When** the reviewer scans each section that references a package manager or ecosystem-specific tool,
**Then** each such reference appears in a balanced multi-ecosystem context — **≥3 ecosystems named in the same section** (e.g., a supply-chain section that references pip-audit AND npm-audit AND cargo-audit together is acceptable; a section that names only pip-audit is a FAIL).

**Positive reframe rationale (plan-auditor iter-1 open-decision)**: the prior negative-grep formulation (`grep` for `pip install` / `npm install` / etc. and assert zero-or-balanced) always lags the ecosystem list — a new package manager not in the grep list would silently pass. The positive reframe ("each package-manager reference MUST appear in a balanced multi-ecosystem context with ≥3 ecosystems named in the same section") is forward-proof: it does not depend on enumerating the 16 ecosystems. The de-facto cross-ecosystem standard carve-out (REQ-SI-013, e.g., cosign/trivy/semgrep named as "the standard \<category\> tool") is NOT a package-manager reference and is exempt from the ≥3-ecosystem rule — the carve-out is governed separately by REQ-SI-013.

**Evidence**: for each section that references a package manager, list the ecosystems named in that section; assert ≥3. The negative grep (`pip install`, `npm install`, `go get`, `cargo add`, `gem install`, `composer require`, `mvn`, `gradle`, `dotnet add` as the SOLE ecosystem pattern) is a secondary check — a section where one of these appears as the sole pattern is a FAIL.

### AC-SI-011 — Template-neutrality §25 (no internal IDs leak)

**Severity**: P0  **Milestone**: M4

**Given** the six distributed skill artifacts,
**When** the reviewer runs the template-neutrality 5-item pre-commit self-check (CLAUDE.local.md §25.3) and the `internal_content_leak_test.go` CI guard,
**Then** zero forbidden content-class items are found:
- no moai-adk internal SPEC IDs (e.g., `SPEC-V3R6-*`) in the skill bodies
- no REQ/AC tokens (e.g., `REQ-SI-001`, `AC-SI-001`) in the skill bodies
- no audit citations (e.g., "Finding AX") in the skill bodies
- no internal dates / commit SHAs in the skill bodies
- no macOS-bias absolute paths

**Evidence**: the 5-item self-check output + `go test ./internal/template/...` (the `internal_content_leak_test.go` guard) green.

### AC-SI-012 — Template-mirror parity (local == template)

**Severity**: P0  **Milestone**: M4

**Given** the three local skill files (`.claude/skills/moai-ref-{llm-security,supply-chain,secops}/SKILL.md`) and the three template-mirror files (`internal/template/templates/.claude/skills/moai-ref-.../SKILL.md`),
**When** the reviewer runs `diff` on each local-vs-template pair,
**Then** each pair is byte-identical (the template mirror is what ships to user projects via `moai init` / `moai update`; the local copy is what runs in the dev project; they must match).

**Evidence**: `diff -r .claude/skills/moai-ref-llm-security internal/template/templates/.claude/skills/moai-ref-llm-security` (and the same for the other two) — empty output.

### AC-SI-013 — Skill-listing matcher descriptions within cap (wc -c on de-YAML-folded-scalar)

**Severity**: P1  **Milestone**: M4

**Given** the three skill frontmatters,
**When** the reviewer measures the character count of the `description` field and the `when_to_use` field separately (each ≤ 1,536 characters, the `maxSkillDescriptionChars` cap),
**Then** each is ≤ 1,536 characters. The measurement is taken on the **de-YAML-folded-scalar string** — i.e., the literal string value after YAML parsing resolves folded scalars (`>-` / `>`) by replacing newlines with spaces and stripping the leading indentation. A raw `wc -c` on the YAML source line (which includes the YAML indentation and the fold marker) over-counts; the correct measurement is on the parsed string value.

**Evidence**: for each skill, extract the `description` and `when_to_use` string values (e.g., via a YAML parser or by manually resolving the folded scalar), then `wc -c` on each resolved string. Per-skill character-count table. A small reference script:

```bash
# Resolve the folded-scalar description and measure character count
python3 -c "
import yaml, sys
with open('SKILL.md') as f:
    # strip frontmatter delimiters for YAML parsing
    content = f.read()
    fm = content.split('---')[1]
    data = yaml.safe_load(fm)
    print('description:', len(data.get('description','')))
    print('when_to_use:', len(data.get('when_to_use','')))
"
```

### AC-SI-014 — Skill size — HARD 400-line secops split threshold

**Severity**: P1  **Milestone**: M3

**Given** the three finished skill bodies,
**When** the reviewer counts non-frontmatter lines via `wc -l` on each SKILL.md,
**Then**:
- `moai-ref-llm-security/SKILL.md` and `moai-ref-supply-chain/SKILL.md` are ≤ 500 lines (the standard skill-authoring ceiling).
- `moai-ref-secops/SKILL.md` is subject to the **HARD 400-line split threshold**: if `wc -l > 400` for `moai-ref-secops/SKILL.md`, the milestone MUST split into `modules/devsecops.md` + `modules/container.md` + `modules/api-ops.md` with `SKILL.md` as a ~150-line overview that cross-references the three modules. The split is MANDATORY above 400 lines — it is a HARD gate, not a tilde guidance (the prior `~500` tilde was ambiguous and invited over-cut before splitting).

**Evidence**: `wc -l` on each SKILL.md. For `moai-ref-secops`, if the count exceeds 400, evidence MUST show the post-split structure (the 4 files: SKILL.md ≤150 lines + 3 module files) AND the cross-reference grep confirming SKILL.md links to all 3 modules.

**Why 400 (not ~500)**: `moai-ref-secops` condenses ~79 upstream topics — the largest of the three skills. A HARD 400-line threshold forces the split decision BEFORE the file becomes unmanageable, preserving skill-listing matcher precision (a 500-line encyclopedia-style skill weakens the description-trigger matching). The ~150-line overview + 3 modules pattern keeps the SKILL.md entry-point compact while the deep-dive content lives in modules (per the skill-authoring file-splitting convention).

### AC-SI-015 — Cross-references to existing moai-ref skills

**Severity**: P2  **Milestone**: M4

**Given** the three finished skill bodies,
**When** the reviewer checks for cross-references to the existing `moai-ref-owasp-checklist` / `moai-ref-api-patterns` / `moai-ref-testing-pyramid`,
**Then** at least one cross-reference exists where the domains overlap — e.g., `moai-ref-secops` cross-references `moai-ref-owasp-checklist` for dev-time OWASP patterns and keeps its own coverage operational; `moai-ref-supply-chain` cross-references `moai-ref-owasp-checklist` for A06 (Vulnerable Components) if applicable.

**Evidence**: the cross-reference grep.

### AC-SI-016 — Verification + red-flags + rationalizations blocks present

**Severity**: P1  **Milestone**: M4

**Given** the three finished skill bodies,
**When** the reviewer scans the body structure,
**Then** each body contains the three MoAI-evolvable blocks (`<!-- moai:evolvable-start id="rationalizations" -->`, `id="red-flags"`, `id="verification"`) following the pattern established by `moai-ref-owasp-checklist`.

**Evidence**: the grep for the three evolvable-block markers per skill.

### AC-SI-017 — Closure gate (3-phase close)

**Severity**: P0  **Milestone**: M5

**Given** all of AC-SI-001 through AC-SI-016 PASS AND AC-SI-018 PASS,
**When** manager-docs produces the single sync commit carrying the `implemented → completed` transition,
**Then** the commit subject names exactly one SPEC-ID (`chore(SPEC-V3R6-SEC-SKILL-INTEGRATION-001): sync-phase artifacts` — NOT a combined/abbreviated scope), the frontmatter `status` transitions to `completed`, and the progress.md §E.4 sync-phase audit-ready signal is populated.

**Evidence**: the sync commit SHA + `git show <sha> --stat` + the `completed` frontmatter + progress.md §E.4.

### AC-SI-018 — research.md containment (never embedded in template binary)

**Severity**: P0  **Milestone**: M4

**Given** the internal research.md at `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/research.md` (which names the upstream `mukul975/Anthropic-Cybersecurity-Skills` verbatim per design.md §C upstream-containment boundary),
**When** the reviewer greps the embed-source paths (`internal/template/templates/` + `internal/template/embed.go`) for the upstream-name signals `mukul975` and `Anthropic-Cybersecurity` (case-insensitive),
**Then** zero matches are found — the research.md is never embedded in the template binary payload that ships to user projects via `moai init` / `moai update`.

**Containment is structural, with this grep as the mechanical guard** (plan-auditor iter-2 D10): the `.moai/specs/` directory lives OUTSIDE the `//go:embed all:templates` directive's `templates/` subtree (the directive at `internal/template/embed.go` line 28 embeds only `all:templates` + the sibling `catalog.yaml`), so `research.md` is structurally incapable of being embedded by construction. The grep is the mechanical guard that catches a future regression if someone were to copy `research.md` into `internal/template/templates/` or add a stray `//go:embed` directive covering `.moai/specs/`. The `research\.md` literal is deliberately DROPPED from the grep pattern (plan-auditor iter-2 D10): it produced 74 pre-existing matches across `internal/template/` from legitimate design-doc references to `research.md` as a generic MoAI artifact, making the old pattern unsatisfiable. The upstream-name signals `mukul975` and `Anthropic-Cybersecurity` are the precise identifiers that would indicate a real containment breach.

**Evidence** — the literal grep command + zero-match output:

```bash
grep -rniE 'mukul975|anthropic-cybersecurity' \
  internal/template/templates/ internal/template/embed.go internal/template/embed_catalog.go
# expected: zero matches (run it, paste actual output)
```

**Live-verification output (plan-auditor iter-2, 2026-06-24)**:

```
$ grep -rniE 'mukul975|anthropic-cybersecurity' \
    internal/template/templates/ internal/template/embed.go internal/template/embed_catalog.go
(0 matches — containment holds)
```

The 0-match result is the mechanical proof that the upstream name never reaches the embed-source paths, confirming the structural containment boundary holds.

**Why this is P0**: this AC closes the containment loop opened by AC-SI-008's scope clarification. AC-SI-008 verifies no attribution in the DISTRIBUTED skill artifacts (`.claude/skills/moai-ref-*/` + template mirrors). AC-SI-018 verifies the research.md itself (which DOES name the upstream) is never embedded in the template binary. Together, the two ACs guarantee: (a) distributed skills carry no upstream name (AC-SI-008), AND (b) the internal research.md that DOES name the upstream never reaches user projects (AC-SI-018). The upstream-containment boundary (design.md §C) holds iff both ACs PASS.

## §C. Edge Cases

- **EC1 — Upstream unavailable mid-run**: if the upstream repo becomes unreachable during M1-M3, the author falls back to public primary sources (OWASP cheat-sheets, MITRE ATT&CK/ATLAS pages, NIST pubs, vendor security advisories) that cover the same defensive concepts. The research.md records the fallback.
- **EC2 — A topic is too dual-use to frame defensively**: if the dual-use review (M4) finds a topic that cannot be framed defensively without including exploitation material, that topic is removed from scope and recorded as a follow-up SPEC candidate. The plan-auditor gate catches this.
- **EC3 — `moai-ref-secops` exceeds the HARD 400-line threshold**: per AC-SI-014, split MANDATORILY into `modules/devsecops.md`, `modules/container.md`, `modules/api-ops.md` with SKILL.md as a ~150-line overview that cross-references the three modules. The 400-line threshold is HARD (not tilde ~500) — exceeding it without splitting is a FAIL. If even the post-split overview cannot stay ≤150 lines, escalate to Tier L (already classified) and split into three separate skills (`moai-ref-secops-ci-cd` / `moai-ref-secops-container` / `moai-ref-secops-api`) per plan.md §B.4 residual-trigger (c), which requires a follow-up SPEC amendment.
- **EC4 — An upstream passage coincidentally matches a moai-adk passage**: standard tool-invocation syntax (e.g., `cosign verify --key cosign.pub imageref`) is functional, not expressive, and is exempt from AC-SI-007. The reviewer distinguishes functional syntax from expressive prose.
- **EC5 — A defensive topic inherently names an attack tool**: e.g., "Falco rules can detect Mimikatz-style credential dumping" names Mimikatz. AC-SI-009 permits this IF the framing is defensive (here is how Falco detects it); the FAIL case is providing the Mimikatz usage procedure.

## §D. Indirect Verification (where direct AC is impractical)

- **Skill-listing matcher triggering** cannot be verified by a unit test (it depends on Claude Code's runtime matcher). Indirect verification: the `description` + `when_to_use` keyword set is reviewed against `skill-writing-craft.md` Part 1 (description craft) and the `## NOT for` line narrows scope. A follow-up `/moai feedback`-driven user report channel captures real-world triggering issues.
- **Re-authorship integrity** (AC-SI-007) cannot be 100% verified mechanically (it is a prose-comparison judgment). Indirect verification: the 9-passage spot-check + the research.md record + the plan-auditor's independent review. The CI guard `internal_content_leak_test.go` catches the mechanical half (forbidden content classes); the spot-check catches the expressive half.

## §E. Severity Convention

- **P0 (MUST)**: AC-SI-001, 002, 003, 004, 005, 006, 007, 008, 009, 011, 012, 017, 018 — these are the legal-integrity (re-authorship AC-007, no-attribution AC-008, research.md-containment AC-018), dual-use-boundary (AC-009), template-parity (AC-011, AC-012), and closure (AC-017) gates. Failure blocks merge.
- **P1 (SHOULD)**: AC-SI-010, 013, 014, 016 — these are quality bars (16-language neutrality positive reframe, listing cap with wc -c measurement, HARD 400-line secops split, evolvable blocks). Failure should be fixed before merge but is not a legal/safety breach.
- **P2 (NICE)**: AC-SI-015 — cross-references. Failure tracks in backlog.

## §F. Definition of Done

This SPEC is **Done** when:
1. All P0 ACs PASS with evidence in progress.md §E.2 — including AC-SI-018 (research.md containment).
2. All P1 ACs PASS OR have a documented waiver in progress.md §E.2.
3. P2 ACs are tracked (PASS or backlog entry).
4. The three skills exist locally AND in the template mirror, byte-identical (AC-SI-012).
5. The 3-phase close (plan → run → sync) completes with the single sync commit carrying the `completed` transition.
6. `moai spec lint` on this SPEC is clean; `go test ./internal/template/...` is green; `golangci-lint run` is green (no Go code changes expected, so this is a baseline-pass).
7. The internal research.md records the upstream (`mukul975/Anthropic-Cybersecurity-Skills`) as inspiration source AND is verified NOT distributed (AC-SI-018 containment grep over `internal/template/` returns zero matches).

## §G. Forward-Looking Checks (post-close validation)

- **FL1 — Skill-listing matcher real-world triggering**: after the skills ship, monitor whether they over-trigger (load on unrelated requests) or under-trigger (fail to load on defensive-security requests). Tune `description` / `when_to_use` in a follow-up if needed.
- **FL2 — Defensive-topic coverage gaps**: practitioner feedback may reveal defensive topics the skills missed (e.g., a new OWASP LLM Top 10 edition). A follow-up SPEC extends coverage; this SPEC's skills are not expected to be encyclopedic.
- **FL3 — Upstream drift**: the upstream repo continues to evolve independently. moai-adk does NOT auto-track upstream. A periodic (manual) re-evaluation may pull new defensive concepts into a follow-up SPEC, subject to the same re-authorship discipline.
