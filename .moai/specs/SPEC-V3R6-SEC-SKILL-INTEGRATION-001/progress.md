# progress.md — SPEC-V3R6-SEC-SKILL-INTEGRATION-001

> Run-phase evidence + audit-ready signals.
> §E.1 is owned by manager-spec (plan-phase). §E.2/§E.3 are owned by manager-develop (run-phase). §E.4 is owned by manager-docs (sync-phase). The era-classification engine greps for the literal `§E.2` / `§E.3` / `§E.4` substrings to classify this SPEC's era — these headings MUST be present verbatim.

## §E.1 Plan-phase Audit-Ready Signal

**Status**: plan-phase artifacts complete (2026-06-24); re-iterated after plan-auditor iter-1 PASS-WITH-DEBT (0.82) to resolve 1 BLOCKING + SHOULD-FIX defects + open-decisions.

**Artifacts authored (Tier L — 5-artifact set)**:
- `spec.md` — **23 GEARS requirements** (REQ-SI-001..023; REQ-SI-023 added for research.md containment), Out-of-Scope section (4 `### Out of Scope —` H3 sub-headings), 12-field frontmatter validated + `tier: "L"` optional field, `version: "0.2.0"`, `status: draft`.
- `plan.md` — **Tier L classification** (upgraded from initial Tier M; justification in §B.4) + 5 milestones (M1-M5) + §E self-verification incl. skill-specific E-REAUTHOR / E-DUALUSE / E-LANGNEUTRAL / E-NOATTRIB / **E-CONTAIN** checks.
- `acceptance.md` — **18 ACs** (AC-SI-001..018; AC-SI-018 added for research.md containment) with Given-When-Then scenarios + §A REQ↔AC traceability matrix + severity matrix + Definition of Done. AC-SI-008 scope clarified (distributed paths only) + literal grep patterns enumerated. AC-SI-009 expanded offensive-keyword list + defensive-verb proxy. AC-SI-010 positive multi-ecosystem reframe. AC-SI-013 wc -c on de-YAML-folded-scalar. AC-SI-014 HARD 400-line secops split threshold.
- `design.md` — **Tier L 5th artifact** (NEW). Documents: §A re-authorship procedure (concept-read → close → write-fresh), §B legal-integrity rationale (17 USC §102(b) / Korean Copyright Act §7 / Apache-2.0 §4(c)), §C upstream-containment boundary (research.md internal-only, never embedded), §D dual-use defensive-framing policy, §E design decisions log.
- `research.md` — internal-only; §A names upstream `mukul975/Anthropic-Cybersecurity-Skills` verbatim (D2 inversion); §B.3 re-authoring notes apply "standard \<category\> tool" framing consistently (D3); §E maintainer note cross-references AC-SI-008 scope + AC-SI-018 containment.
- `progress.md` — this file (§E.1 populated; §E.2-§E.4 placeholder headings only).

**SPEC ID Pre-Write Self-Check** (re-verified at re-iteration):
```
decomposition: SPEC ✓ | V3R6 ✓ | SEC ✓ | SKILL ✓ | INTEGRATION ✓ | 001 ✓ → PASS
```

**Frontmatter 12-canonical-field check + tier**: all fields present (`id`, `title`, `version: "0.2.0"`, `status: draft`, `created: 2026-06-24`, `updated: 2026-06-24`, `author: manager-spec`, `priority: P1`, `phase: "v3.0.0"`, `module`, `lifecycle: spec-anchored`, `tier: "L"`, `tags`). No snake_case aliases. `tags` is comma-separated string.

**spec-lint result**: `moai spec lint .moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/spec.md` → `✓ No findings — all SPEC documents are valid` (clean, re-verified post-re-iteration).

**Plan-auditor iter-1 defect-resolution status**:
- **D7 (BLOCKING — tier field + Tier L upgrade)**: RESOLVED. `tier: "L"` added to frontmatter; plan.md §B.4 documents the Tier-L upgrade justification (3 compounding factors); design.md authored as the Tier-L 5th artifact.
- **D3 (SHOULD-FIX — REQ-SI-013 internal contradiction)**: RESOLVED. REQ-SI-013 amended with the explicit de-facto-standard-tool carve-out; research.md §B.2/§B.3 re-authoring notes apply "standard \<category\> tool" framing consistently.
- **D2 (SHOULD-FIX — name upstream verbatim + enumerate AC-SI-008 grep patterns)**: RESOLVED. research.md §A names `mukul975/Anthropic-Cybersecurity-Skills` + agentskills.io verbatim; AC-SI-008 documents the scope boundary (distributed paths only, NOT `.moai/specs/`) + enumerates the literal grep patterns.
- **D4 (SHOULD-FIX — add AC-SI-018 research.md containment)**: RESOLVED. AC-SI-018 (P0) added with literal grep command; REQ-SI-023 added to map to it; §A traceability matrix + §E P0 count + §F Definition of Done updated.
- **D8 (AC-SI-014 tilde ~500 → resolved by HARD 400)**: RESOLVED. AC-SI-014 now specifies HARD 400-line secops split threshold (not tilde).
- **D9 (AC-SI-013 measurement)**: RESOLVED. AC-SI-013 specifies `wc -c` on the de-YAML-folded-scalar string (with a reference python3 snippet).
- **Open-decision (secops split threshold)**: APPLIED. HARD 400 lines (not ~500) in plan.md §B.4c + AC-SI-014.
- **Open-decision (offensive-keyword grep list)**: APPLIED. AC-SI-009 adds bloodhound/rubeus/impacket/evil-winrm/responder/seatbelt/sharpersist/certipy + `-i` flag + defensive-verb proxy.
- **Open-decision (16-language neutrality)**: APPLIED. AC-SI-010 reframed as positive multi-ecosystem balance (≥3 ecosystems per section).
- **D1 (AC-SI-007 spot-check coverage)**: DEFERRED to run-phase M4 (debt). At M4, raise sample to ≥10 passages/skill OR add mechanical full-sentence grep vs upstream. Recorded in plan.md §F M4 + §E E-REAUTHOR.
- **D5/D6 partial (keyword enumeration gap + semantic-coverage limit)**: keyword enumeration RESOLVED (AC-SI-009 expanded list). Semantic-coverage limit (re-framed offense evading grep) REMAINS a run-phase M4 human-verdict gate — recorded as residual risk in AC-SI-009 + plan.md §E E-DUALUSE.

**Parallel-session baseline**: local HEAD == `origin/main` (`0 0` divergence as of 2026-06-24 re-iteration), no active session on this SPEC (`moai session list --json --filter-spec=...` returns `[]`).

**Ready for**: plan-auditor Phase 2 re-review (iter-2) → Implementation Kickoff Approval (plan-to-implement human gate) → manager-develop run-phase M1. NOT proceeding to run-phase autonomously (per the Implementation Kickoff Approval mandatory-restoration policy).

## §F — Phase 0.95 Mode Selection (orchestrator)

**Input parameters**:
- Tier: L (PASS threshold 0.85; plan-auditor iter-2 PASS-WITH-DEBT 0.86, monotonic ↑ vs iter-1 0.82)
- Scope: 3 SKILL.md (local) + 3 template mirrors + possible `moai-ref-secops/modules/` ≈ 6-9 files, markdown-only
- Domain count: 1 (defensive-security reference-skill authoring)
- File language mix: 100% markdown
- Concurrency benefit: LOW (content authoring is coding-heavy per Anthropic coding-task parallelism caveat; re-authorship integrity + cross-skill style consistency favor sequential)

**Mode evaluation**:
- Mode 1 trivial — not selected (multi-file semantic authoring)
- Mode 2 background — not selected (Write-heavy, foreground required)
- Mode 3 agent-team — not selected (single domain; prereqs not targeted)
- Mode 4 parallel — not selected (content authoring is coding-heavy; cross-skill consistency needs sequential)
- Mode 5 sub-agent — **SELECTED** (sequential per milestone)
- Mode 6 workflow — not selected (not a ≥30-file mechanical-uniform transform)

**Decision**: sub-agent (Mode 5) — sequential manager-develop per milestone (M1 → M2 → M3 → M4); M5 = manager-docs sync.

**Justification**: Per Anthropic's coding-task parallelism caveat, content/coding-heavy work defaults to sequential sub-agents. M1 (`moai-ref-llm-security`) is delegated first to establish the `moai-ref-*` structural pattern + the re-authorship discipline; a checkpoint gate after M1 validates AC-SI-001/002 + E-DUALUSE/E-NOATTRIB/E-LANGNEUTRAL before the pattern propagates to M2/M3. Implementation Kickoff Approval PASSED (user-approved 2026-06-24; explicit-gate branch — Tier L + security domain per IGGDA §H conditions (c)/(d)).

## §E.2 Run-phase Evidence

### M1 — `moai-ref-llm-security` skill (2026-06-24)

**Milestone scope**: authored the first of three defensive-cybersecurity reference skills — `moai-ref-llm-security` (AI/LLM defensive security). Local copy + template mirror byte-identical. spec.md frontmatter `status: draft → in-progress` on this commit. Establishes the `moai-ref-*` structural pattern for M2/M3.

**Files touched (M1 scope only)**:
- `.claude/skills/moai-ref-llm-security/SKILL.md` (NEW, 289 lines)
- `internal/template/templates/.claude/skills/moai-ref-llm-security/SKILL.md` (NEW, byte-identical mirror)
- `internal/template/catalog.yaml` (registered new skill under `optional-pack:devops`; regen via `make build`)
- `internal/template/catalog_tier_audit_test.go` (`expectedSkillCount` 31 → 32)
- `internal/template/catalog_loader_test.go` (`expectedTotal` 38 → 39)
- `internal/template/embed_catalog_test.go` (`wantTotal` 38 → 39)
- `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/spec.md` (frontmatter `status` → `in-progress` ONLY)
- `.moai/specs/SPEC-V3R6-SEC-SKILL-INTEGRATION-001/progress.md` (this §E.2 evidence)

> **Catalog registration note (in-scope cascade)**: the CI guard `TestAllSkillsInCatalog` (sentinel `CATALOG_ENTRY_MISSING`) requires every `.claude/skills/<name>/` directory on disk to be registered in `internal/template/catalog.yaml`. Registering the new skill (+ updating the 3 count constants) is the same-SPEC deployment cascade that makes M1 deployable via `moai init` / `moai update`. This is the catalog-hash-regen cascade pattern (L46-attributable to this SPEC's scope envelope).

**E1 — AC PASS/FAIL Matrix (M1 ACs)**:

| AC | Subject | Status | Evidence command | Actual output |
|----|---------|--------|------------------|---------------|
| AC-SI-001 | Skill 1 frontmatter shape | PASS | `python3 yaml-parse SKILL.md` | `name: moai-ref-llm-security`, `user-invocable: False`, metadata keys `[category, status, tags, updated, version]` all-string=True, `progressive_disclosure.enabled: True` |
| AC-SI-002 | LLM defensive topic coverage | PASS | body scan + section grep | prompt-injection defense ✓; OWASP LLM Top 10 all 10 (LLM01-10) ✓ (≥5 req); MCP/agentic tool-call hardening ✓; training-data poisoning detection ✓; model-output validation/guardrails ✓; MITRE ATLAS 4 T-IDs (AML.T0051/T0020/T0053/T0054/T0057) ✓ (≥3 req, all defensive-correlation); NIST AI RMF GOVERN/MAP/MEASURE/MANAGE ✓ |
| AC-SI-012 | Template-mirror parity | PASS | `diff local template` + `diff -r` | empty (byte-identical, dir-level) |
| AC-SI-013 | desc + when_to_use ≤1536 each | PASS | `python3 len() on de-folded scalar` | description: 710 chars; when_to_use: 424 chars (both ≤1536) |
| AC-SI-014 | Skill size ≤500 lines | PASS | `wc -l SKILL.md` | 289 (≤500; llm-security is not the 400-line-split skill) |
| AC-SI-016 | 3 evolvable blocks present | PASS | `grep -c moai:evolvable-start` | 3 (ids: rationalizations, red-flags, verification) |

**Topic-coverage table (AC-SI-002 detail — defensive topic → body section)**:

| Defensive topic | Body section | Framing |
|-----------------|--------------|---------|
| Prompt-injection defense | `## Prompt-Injection Defense` (direct + indirect) | defense/detect/prevent |
| OWASP LLM Top 10 (≥5/10) | `## OWASP LLM Top 10 — Defensive Mapping` (all 10) | defensive check + hardening control |
| MCP/agentic tool-call hardening | `## MCP / Agentic Tool-Call Hardening` | least-privilege + output validation |
| Training-data poisoning detection | `## Training-Data and Model Poisoning Detection` | provenance/canary/lineage |
| Model-output validation | `## Model-Output Validation and Guardrails` | schema/filtering/encoding |
| MITRE ATLAS (≥3 T-IDs, defensive) | `## MITRE ATLAS — Defensive Correlation` | technique → counter-control |
| NIST AI RMF | `## NIST AI RMF — Governance Mapping` | GOVERN/MAP/MEASURE/MANAGE |

**E-REAUTHOR (independence note, 3-passage)**: skill body was authored fresh from public primary sources (OWASP LLM Top 10 2025, MITRE ATLAS, NIST AI RMF 1.0) — concept-level only, no verbatim upstream copy. Sample passages: (1) "an LLM does not distinguish 'data' from 'instructions' the way a parser does" — original framing of the trust-boundary concept; (2) "an output-only guardrail misses injection that has already changed the model's plan" — original guardrail-placement argument; (3) "treat every tool result as untrusted input on its way back into the model" — original re-validation framing. No standard tool-invocation syntax was even required (no CLI examples in this skill). Independence confirmed.

**E-DUALUSE** (`grep -niE 'mimikatz|kerberoast|ntlm.relay|cobalt strike|sliver|mythic|metasploit|pass-the-hash|pass-the-ticket|lsass.dump|credential dump|c2 framework|command and control|red team|pen test|penetration test|bloodhound|rubeus|impacket|evil-winrm|responder|seatbelt|sharpersist|certipy' SKILL.md`): **0 matches** (exit 1). Cleanest possible — zero offensive keywords. All attack-vector mentions (prompt injection, jailbreak, poisoning) appear strictly in detect/prevent/defend framing.

**E-LANGNEUTRAL** (`grep -niE 'pip install|npm install|go get|cargo add' SKILL.md`): **0 matches** (exit 1). No package-manager references at all; tools named ecosystem-neutrally ("LLM guardrail framework", "schema validator"). AC-SI-010 (≥3-ecosystem rule) vacuously satisfied — no package-manager-specific reference exists in the body.

**E-NOATTRIB** (`grep -rniE 'mukul975|anthropic-cybersecurity|agentskills\.io' <local> <template>` + `based on|adapted from|inspired by|github\.com/mukul`): **0 matches** (exit 1) across both distributed paths. No upstream attribution anywhere.

**E2 — build/parity**: `make build` exit 0 (catalog regenerated, hash `0d8ec550...` for new skill); `diff local template` empty; `go test ./internal/template/...` → `ok` (all guards pass including `TestAllSkillsInCatalog`, `internal_content_leak_test.go`, `template_neutrality_audit_test.go`).

**E5 — lint**: `golangci-lint run --timeout=2m` → `0 issues` (no Go logic change; count-constant + catalog edits only). `moai spec lint spec.md` → `0 error(s), 1 warning(s)` — the lone warning is the pre-existing `OwnershipTransitionInvalid` on the plan-phase `(none)→draft` transition (orchestrator-direct committed the plan artifacts), unchanged from baseline, NOT introduced by M1.

**E6 — template-neutrality** (`grep -niE 'SPEC-V3R6|REQ-SI-|AC-SI-|Finding [A-Z][0-9]' SKILL.md`): **0 matches** (exit 1). No internal SPEC/REQ/AC tokens, dates, or SHAs in the distributed skill body. macOS-bias absolute path grep + date/SHA body grep also 0.

**E-CONTAIN** (AC-SI-018, `grep -rniE 'mukul975|anthropic-cybersecurity' internal/template/templates/ internal/template/embed.go internal/template/embed_catalog.go`): **0 matches** (exit 1). research.md never reaches the embed-source paths — containment holds structurally (`.moai/specs/` is outside the `//go:embed all:templates` subtree).

**Full-suite note**: `go test ./...` showed 2 transient FAILs in `internal/hook` (`TestHookWrapper_ValidJSON`, `TestHookWrapper_MoaiBinaryFallback`) with `signal: killed`. Confirmed NOT a regression — `internal/hook` is unmodified by M1, and both tests PASS in isolation (`go test -count=1 ./internal/hook/` → `ok 1.5s`). The full-suite failures are parallel-load subprocess-spawn timeouts in the worktree sandbox, unrelated to this SPEC's markdown + catalog changes.

**Residual risk (M1)**: E-REAUTHOR is a 3-passage sampling floor (debt D1 — raise to ≥10 passages OR mechanical full-sentence grep deferred to M4). E-DUALUSE syntactic grep cannot catch a novel-phrasing re-framed offense (debt D5/D6 — M4 human verdict is the final gate). Both deferred to M4 cross-skill review per plan.md §F.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
