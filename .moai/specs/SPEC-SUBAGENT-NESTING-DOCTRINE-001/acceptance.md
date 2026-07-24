---
id: SPEC-SUBAGENT-NESTING-DOCTRINE-001
title: "Subagent-nesting doctrine correction + auditor read-only nesting pilot — Acceptance"
version: "0.1.0"
status: draft
created: 2026-07-24
updated: 2026-07-24
author: manager-spec
priority: P2
phase: "v3.0.2 target"
module: ".claude"
lifecycle: spec-anchored
tags: "doctrine, subagent-nesting, claude-code, agent-authoring, sync-auditor"
tier: M
---

# Acceptance Criteria — SPEC-SUBAGENT-NESTING-DOCTRINE-001

All greps run from repo root `/Users/goos/MoAI/moai-adk-go`. Doc-accuracy ACs assert BOTH the stale phrasing is gone AND the v2.1.217 facts are present (delta, not presence-only). Every M1 doc-accuracy AC applies to BOTH the live surface AND its `internal/template/templates/` mirror.

## §D AC Matrix

| AC | Requirement | Type | Verification |
|----|-------------|------|--------------|
| AC-SND-001 | REQ-SND-002 | doc-accuracy | CLAUDE.md §4 Watch note |
| AC-SND-002 | REQ-SND-003 | doc-accuracy | CLAUDE.md §14 concurrency caps |
| AC-SND-003 | REQ-SND-004 | doc-accuracy | agent-authoring §Agent(agent_type) Restrictions |
| AC-SND-004 | REQ-SND-005 | doc-accuracy | agent-authoring §Fork Subagents |
| AC-SND-005 | REQ-SND-006 | doc-accuracy | agent-authoring §Tool Permissions |
| AC-SND-006 | REQ-SND-007 | doc-accuracy | agent-patterns §Deprecated |
| AC-SND-007 | REQ-SND-008 | doc-accuracy | orchestration-mode-selection §Mode 6 |
| AC-SND-008 | REQ-SND-009 | conditional | zone-registry CONST clauses re-sync |
| AC-SND-009 | REQ-SND-010/022 | mirror-parity | live == template + make build |
| AC-SND-010 | REQ-SND-011 | neutrality | no internal content in template |
| AC-SND-011 | REQ-SND-013/014/021 | held-in | sync-auditor CAN nest (gated) |
| AC-SND-012 | REQ-SND-015/019 | held-out | env unset ⇒ flat default |
| AC-SND-013 | REQ-SND-018 | boundary guard | no AskUserQuestion in auditor path |
| AC-SND-014 | REQ-SND-017 | read-only children | no write-capable child |
| AC-SND-015 | REQ-SND-016 | verdict ownership | verdict not delegated |
| AC-SND-016 | REQ-SND-019 | env dev-only | env absent from template settings.json |

## §D.1 Doc-accuracy scenarios (M1)

### AC-SND-001 — Watch note rewrite (Given-When-Then)

- **Given** `CLAUDE.md` §4 has been corrected,
- **When** the Watch note is inspected,
- **Then** the stale phrasing is gone AND the v2.1.217 facts + double-guarantee + M2-exception reference are present:

```bash
# stale phrasing gone (live + template)
grep -rn "fixed and not configurable\|depth five\|a subagent at depth five" CLAUDE.md internal/template/templates/CLAUDE.md
# → 0 matches
# v2.1.217 facts present
grep -c "CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH" CLAUDE.md   # >= 1
grep -c "v2.1.217" CLAUDE.md                                # >= 1
grep -ciE "default.*off|off by default" CLAUDE.md          # >= 1 (Watch-note vicinity)
grep -ciE "double guarantee|both" CLAUDE.md                # >= 1
grep -ci "sync-auditor" CLAUDE.md                          # >= 1 (pilot exception reference)
```

### AC-SND-002 — §14 concurrency caps

Presence greps are §14-anchored (REQ-SND-003 documents the caps in §14), not global — a global grep would PASS even if the caps landed in the wrong section. The §14 body is bounded between the `## 14.` and `## 15.` headings.

```bash
# §14-anchored: extract the §14 section, assert the caps are present WITHIN it
S14() { awk '/^## 14\. /{f=1} /^## 15\. /{f=0} f' CLAUDE.md; }
S14 | grep -c "CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS"    # >= 1 (within §14)
S14 | grep -c "CLAUDE_CODE_MAX_SUBAGENTS_PER_SESSION"   # >= 1 (within §14)
S14 | grep -A2 "MAX_CONCURRENT_SUBAGENTS" | grep -c "20"    # >= 1
S14 | grep -A2 "MAX_SUBAGENTS_PER_SESSION" | grep -c "200"  # >= 1
# mirror parity (template §14)
awk '/^## 14\. /{f=1} /^## 15\. /{f=0} f' internal/template/templates/CLAUDE.md | grep -c "CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS"  # >= 1
```

### AC-SND-003 — agent-authoring §Agent(agent_type) Restrictions

```bash
F=".claude/rules/moai/development/agent-authoring.md"
grep -n "nesting depth is fixed, not configurable" "$F"   # → 0 matches
grep -c "CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH" "$F"        # >= 1
grep -ciE "v2\.1\.217|default off" "$F"                    # >= 1
# + template mirror parity for the same file
```

### AC-SND-004 — agent-authoring §Fork Subagents

```bash
F=".claude/rules/moai/development/agent-authoring.md"
# the stale "fixed depth cap (depth 5) as of ... v2.1.187" is marked superseded, not left as current fact
grep -n "fixed depth cap (depth 5)" "$F" | grep -viE "supersed|previously|until v2\.1\.216"   # → 0 unmarked matches
grep -ciE "supersed|configurable" "$F"   # >= 1 (near Fork Subagents)
```

### AC-SND-005 — agent-authoring §Tool Permissions

```bash
F=".claude/rules/moai/development/agent-authoring.md"
# note that the runtime default is ALSO off (flat no longer rests on tools-omission alone)
grep -ciE "runtime default.*off|default.*off.*v2\.1\.217" "$F"   # >= 1 in Tool Permissions vicinity
```

### AC-SND-006 — agent-patterns §Deprecated Hierarchical

Delta-form (not presence-only, per the acceptance.md header "delta, not presence-only" principle): asserts BOTH the v2.1.217 facts are present AND the stale-as-current "nesting DOES exist ... as of v2.1.172" framing REQ-SND-007 targets cannot survive a PASS unless it is explicitly marked historical/superseded. Applies to BOTH the live file AND its template mirror.

```bash
F=".claude/rules/moai/development/agent-patterns.md"
T="internal/template/templates/.claude/rules/moai/development/agent-patterns.md"
# presence: v2.1.217 fact + M2 pilot exception reference present
grep -c "v2.1.217" "$F"        # >= 1
grep -ci "sync-auditor" "$F"   # >= 1 (M2 pilot exception reference)
# absence (delta): the present-tense stale framing cannot survive a PASS unless marked historical.
# NOTE: the historical-marker filter deliberately OMITS the generic word "previously" — it collides
# with an unrelated clause ("previously documented as a viable Pattern 6") on the very target line,
# which would falsely exclude the stale match. Use phrase-specific historical markers only.
grep -nE "nesting DOES exist|as of Claude Code v2\.1\.172" "$F" | grep -viE "supersed|until v2\.1\.216|no longer|introduced in"   # → 0 unmarked matches
# template mirror: same gone-check
grep -nE "nesting DOES exist|as of Claude Code v2\.1\.172" "$T" | grep -viE "supersed|until v2\.1\.216|no longer|introduced in"   # → 0 unmarked matches
```

### AC-SND-007 — orchestration-mode-selection §Mode 6

```bash
F=".claude/rules/moai/workflow/orchestration-mode-selection.md"
grep -c "scaling NOT nesting" "$F"   # >= 1 (distinction RETAINED)
grep -ciE "v2\.1\.217|default off" "$F"   # >= 1 (version note refreshed)
```

## §D.2 Conditional AC (M1)

### AC-SND-008 — zone-registry conditional re-sync

- **Given** the M1 §14 concurrency-cap sentence has (or has not) been authored inside the CONST-V3R2-020 mirrored clause span,
- **When** the zone-registry entries are checked,
- **Then** either the CONST-V3R2-020/044 clause text is unchanged from baseline (nesting facts did not touch the background/concurrency clauses — the recommended outcome), OR if the mirrored clause text changed, the zone-registry entry matches it:

```bash
Z=".claude/rules/moai/core/zone-registry.md"
# CONST clauses remain about background/concurrency, NOT nesting (no nesting terms injected)
grep -A6 "id: CONST-V3R2-020" "$Z" | grep -ciE "nest|spawn.depth"   # 0 (nesting facts stay out of these clauses)
grep -A6 "id: CONST-V3R2-044" "$Z" | grep -ciE "nest|spawn.depth"   # 0
```

## §D.3 Mirror-parity + neutrality (M1 + M2)

### AC-SND-009 — Template-First mirror parity + make build

- **Given** all live edits are complete,
- **When** each edited live file is compared to its template mirror,
- **Then** the mirror is byte-identical (modulo the known template neutralization set), `make build` exits 0, and the mirror-parity test suite is green:

```bash
make build; echo "make build exit=$?"        # exit 0
go test ./internal/template/... 2>&1 | tail -20; echo "exit=$?"   # exit 0 (mirror-parity + neutrality tests)
```

### AC-SND-010 — Template neutrality (no internal-content leak)

```bash
# no SPEC ID / internal date / commit SHA in the edited template content
grep -rn "SPEC-SUBAGENT-NESTING-DOCTRINE-001" internal/template/templates/   # → 0 matches
go test ./internal/template/ -run 'TestTemplateNeutralityAudit|TestTemplateNoInternalContentLeak' 2>&1 | tail; echo "exit=$?"  # exit 0
```

## §D.4 M2 held-in / held-out / boundary

### AC-SND-011 — Held-in: sync-auditor CAN nest when gated (Given-When-Then)

- **Given** `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` is set to a positive integer AND `Agent` is in `sync-auditor` `tools`,
- **When** `sync-auditor` runs,
- **Then** it is configured to spawn a read-only per-dimension verifier child — evidence is grep-observable (tools list + body documentation):

```bash
F=".claude/agents/moai/sync-auditor.md"
grep -n "^tools:" "$F" | grep -qw "Agent" && echo "Agent-in-tools OK"    # present
grep -ciE "per-dimension|Functionality.*Security.*Craft.*Consistency|verifier child" "$F"   # >= 1 (body documents the pattern)
grep -ciE "Explore|mode: \"plan\"|mode: 'plan'|general-purpose" "$F"     # >= 1 (read-only child mechanism)
```

Residual-risk (not a mechanical AC): a full runtime nested execution is exercisable only in a dev session with the env set — recorded as residual risk (spec.md §E Out of Scope — Live nested-execution).

### AC-SND-012 — Held-out: env unset ⇒ flat, byte-identical default

- **Given** the shipped distribution (env unset),
- **When** the distributed configuration and the other 10 agents are inspected,
- **Then** nesting is off and the fleet is flat:

```bash
# env absent from distributed template settings.json
grep -rn "CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH" internal/template/templates/.claude/settings.json internal/template/templates/.claude/settings.json.tmpl 2>/dev/null   # → 0 matches
# the other 10 retained agents carry NO standalone Agent in tools (pilot set = sync-auditor only; plan-auditor deferred, carries no Agent)
for f in .claude/agents/moai/*.md; do
  case "$f" in *sync-auditor.md) continue;; esac
  grep -n "^tools:" "$f" | grep -qw "Agent" && echo "UNEXPECTED Agent in $f"
done
# → no "UNEXPECTED" lines
```

### AC-SND-013 — Boundary guard: no AskUserQuestion invocation in auditor path

Verifies the ABSENCE of an actual `AskUserQuestion` invocation in the sync-auditor / child path — NOT the mere presence of the word. The exclusion filter is widened so REQ-SND-021's mandated inline boundary sentence (REQ-SND-018 phrasing: "no ... shall invoke AskUserQuestion ... the single-point-of-contact boundary holds at every depth") does not produce a false result; any surviving match would be an actual invocation instruction.

```bash
grep -rn 'AskUserQuestion\|mcp__askuser' .claude/agents/moai/sync-auditor.md \
  | grep -v "^[^:]*:[0-9]*:[ \t]*#" \
  | grep -viE "MUST NOT|never|prohibited|boundary|barred|shall not invoke|no .*shall invoke|single-point-of-contact|holds at every depth|nesting-independent"
# → 0 matches (every remaining occurrence would be an actual invocation, never a prohibition statement)
```

### AC-SND-014 — No write-capable child (read-only enforcement)

The no-write-child check targets tool-GRANT syntax — a `tools:`-line scope or an explicit `Write, Edit` tool-list adjacency — NOT free prose. REQ-SND-021's mandated inline read-only-child documentation (REQ-SND-017) naturally contains prose like "a verifier child is NOT granted Write/Edit", which a free-prose `child.*Write` grep would false-match; the tool-grant-syntax grep does not.

```bash
F=".claude/agents/moai/sync-auditor.md"
# read-only enforcement mechanism documented (Explore / mode:plan)
grep -ciE "read-only|Explore|mode: \"plan\"" "$F"   # >= 1
# no write-capable grant: no tools:-line granting Write/Edit, and no explicit Write,Edit tool-list adjacency
grep -nE "^[[:space:]]*tools:.*(Write|Edit)|Write, ?Edit|Edit, ?Write" "$F"   # → 0 matches (no write-capable child/agent tool grant)
```

### AC-SND-015 — Verdict ownership retained by top-level sync-auditor

```bash
F=".claude/agents/moai/sync-auditor.md"
grep -ciE "binding.*verdict.*(sync-auditor|top-level|not delegat)|verdict.*(remains|stays).*owned" "$F"   # >= 1
```

### AC-SND-016 — Pilot env is local/dev-only

```bash
# distributed template settings.json does NOT set the pilot env (== AC-SND-012 first grep, restated as an explicit dev-only AC)
grep -rn "CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH" internal/template/templates/   # → 0 matches
```

## §D.5 Edge Cases

- **plan-auditor (D5 RESOLVED — deferred)**: `plan-auditor` is NOT in the pilot scope (sync-auditor only). It stays in the "other 10 flat" set — AC-SND-012's loop does NOT skip `plan-auditor.md`, and `plan-auditor.md` MUST carry no `Agent` in `tools` and remain unmodified. The `plan-auditor` read-only nesting pilot is deferred to a future SPEC (spec.md §E Out of Scope — plan-auditor nesting pilot).
- **Concurrency-cap sentence placement (D4)**: if the §14 caps land inside the CONST-V3R2-020 mirrored span, AC-SND-008's re-sync branch applies; recommended authoring keeps the caps as a distinct sentence so no re-sync fires.
- **Line-number drift**: all live anchors are 2026-07-24 reads; re-anchor by content token before edit (AP-4).

## §D.6 Definition of Done

- [ ] AC-SND-001..008 doc-accuracy PASS on BOTH live and template mirror.
- [ ] AC-SND-009/010 mirror-parity + neutrality PASS (`make build` + `go test ./internal/template/...` exit 0).
- [ ] AC-SND-011..016 M2 held-in / held-out / boundary / read-only / verdict / env-dev-only PASS.
- [ ] `moai spec lint` (or repo spec-lint) MUST-FIX 0 for this SPEC (repo-global exit code filtered via `--json`).
- [x] Clarification markers (D5, D6) RESOLVED at plan finalization — D5: M2 pilot = `sync-auditor` only (`plan-auditor` deferred to a future SPEC); D6: M1 + M2 both ship in v3.0.2 (shipped default flat).
- [ ] No `model-policy.md` edit (scope discipline).

## §D.7 Traceability

Every REQ-SND-001..022 maps to at least one AC-SND row above (M1 doc-accuracy REQ-001..008 → AC-001..008; mirror/neutrality REQ-010/011/022 → AC-009/010; M2 REQ-013..021 → AC-011..016). REQ-SND-012 (no runtime behavior change) and REQ-SND-020 (plan-auditor NOT piloted — deferred; no `Agent` tool) are verified by AC-SND-012 (held-out flat: plan-auditor is one of the "other 10" agents the loop confirms carry no `Agent`) + the §D.5 plan-auditor edge case.
