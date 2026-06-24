# Research — SPEC-CC2186-BG-PERMISSION-RATIONALE-001

> Plan-phase research artifact. Captures the verified CC 2.1.186 behavior change, the surviving rationale basis, and the run-phase edit-time re-confirmation caveat. Cross-references the full harness research report.

## §1. The upstream behavior change (verified)

**Claude Code 2.1.186 changelog (verbatim)**: "Changed background subagents to surface permission prompts in the main session instead of auto-denying; the dialog shows which agent is asking, and Esc denies just that tool."

**Authoritative source provenance** (from `.moai/research/cc-update-2.1.183-to-2.1.186.md`):
- GitHub raw CHANGELOG (`raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md`) — top-3 entries confirmed as 2.1.186 / 2.1.185 / 2.1.183 (2.1.184 is a release gap).
- Session-provided `/release-notes` 2.1.186 text matches the GitHub raw CHANGELOG 2.1.186 entry.

**Before 2.1.186**: a background subagent (`run_in_background: true`) that hit a non-pre-approved permission prompt auto-denied because it could not interact with the user.

**As of 2.1.186**: the prompt is surfaced to the main session; the leader/user can respond, and Esc denies just that one tool.

## §2. What drifts vs what survives

### 2.1 The single genuine drift (the rationale, not the conclusion)

The MoAI doctrine's *stated mechanism* is now factually outdated. The clause that drifts (agent-common-protocol § Background Agent Execution):

> "Background agents auto-deny all non-pre-approved permission prompts **because they cannot interact with the user**."

The causal clause "because they cannot interact with the user" is the primary drift — as of 2.1.186 they CAN interact with the user via the main session.

The same outdated "auto-deny" *descriptor* appears in two more surfaces:
- CLAUDE.md §14 "Background Agent Write Restriction": "... `run_in_background: true`) auto-deny Write/Edit operations ..."
- zone-registry CONST-V3R2-020 clause: "Background subagents (run_in_background: true) auto-deny Write/Edit operations. Use run_in_background: false for agents that modify files."

### 2.2 What survives (the retained conclusion + its surviving basis)

The [HARD] behavioral conclusion — background subagents MUST NOT Write/Edit; use `run_in_background: false` for writes; read-only safe in background — REMAINS VALID on a separate, surviving basis. Two surviving justifications:

1. **Allowlist non-inheritance (already in the doctrine, retained verbatim)**: even with `mode: "bypassPermissions"`, the background execution context does not fully inherit the parent session's permission allowlist. This sentence in agent-common-protocol § Background Agent Execution was never dependent on the auto-deny mechanism — it survives as-is.

2. **Flow-interruption / predictability (the new survivor)**: surfacing a permission prompt in the main session for *each* background write interrupts the leader's flow, defeating the purpose of background (parallel) execution. MoAI retains `run_in_background: false` for writes by deliberate conservative policy.

### 2.3 zone-registry CONST-V3R2-044 — pure conclusion, NOT a drift

CONST-V3R2-044 clause = "Background subagents (run_in_background: true) MUST NOT perform Write/Edit operations." This is the pure behavioral conclusion with NO embedded false rationale (no "auto-deny", no "cannot interact"). It is the retained conclusion and is explicitly OUT OF SCOPE (KEEP AS-IS) — see spec.md §F.

## §3. Run-phase edit-time re-confirmation caveat (LOAD-BEARING precondition)

[HARD] The changelog bullet is terse. The exact 2.1.186 background-permission surface wording (which agent is shown asking, the Esc-denies-just-that-tool detail, the main-session surfacing mechanism) MUST be re-confirmed against the official sub-agents doc at run time BEFORE the [HARD] rationale rewrite is finalized:

- Official doc to re-confirm: `code.claude.com/docs/en/sub-agents`
- Do NOT rewrite the [HARD] rationale from the changelog bullet alone (research report §"Unverifiable/partial/deferred" notes the official settings/sub-agents doc was NOT re-fetched in the harness run — classification was sufficient from the verbatim release note, but the exact rewrite surface needs edit-time confirmation).
- This is REQ-BGR-008 / AC-BGR-008: the re-confirmation evidence (the verbatim 2.1.186 surface from the official doc) is recorded in progress.md §E.2 before the rewrite.

## §4. Grep evidence (plan-phase, reproduced by orchestrator 2026-06-23)

```
$ grep -n 'auto-deny\|Background Agent Write' CLAUDE.md internal/template/templates/CLAUDE.md
CLAUDE.md:289:- **Background Agent Write Restriction**: [ZONE:Frozen] [HARD] Background subagents (`run_in_background: true`) auto-deny Write/Edit operations. ...
internal/template/templates/CLAUDE.md:289:- **Background Agent Write Restriction**: ... auto-deny Write/Edit operations. ...  (byte-identical to live)

$ grep -n (Background Agent Execution paragraph) .claude/rules/moai/core/agent-common-protocol.md
190: Background agents auto-deny all non-pre-approved permission prompts because they cannot interact with the user. Even with `mode: "bypassPermissions"`, the background execution context does not fully inherit the parent session's permission allowlist.
(mirror: internal/template/templates/.../agent-common-protocol.md:159 — byte-identical)

$ grep -n 'id: CONST-V3R2-020\|id: CONST-V3R2-044' .claude/rules/moai/core/zone-registry.md
219:- id: CONST-V3R2-020      # the `clause:` line being corrected is at L224 (+5 from this id: line)
417:- id: CONST-V3R2-044      # the `clause:` line is at L422 (+5 from this id: line) — OUT OF SCOPE

# The CONST block layout is: id (+0) → zone (+1) → zone_class (+2) → file (+3) → anchor (+4) → clause (+5) → canary_gate (+6).
# The drifting text lives on the clause: line (+5), NOT the id: line. A grep -A1 on the id: line
# would STOP at +1 and MISS the clause: line — this is why AC-BGR-005 Locus 3 uses -A6 (or anchors
# directly on the clause: text).
$ sed -n '224p' .claude/rules/moai/core/zone-registry.md
  clause: "Background subagents (run_in_background: true) auto-deny Write/Edit operations. Use run_in_background: false for agents that modify files."   # CONST-V3R2-020 — drift target (auto-deny descriptor)
$ sed -n '422p' .claude/rules/moai/core/zone-registry.md
  clause: "Background subagents (run_in_background: true) MUST NOT perform Write/Edit operations."   # CONST-V3R2-044 — OUT OF SCOPE (pure conclusion, no drift token)
# (mirror zone-registry: CONST-V3R2-020 / CONST-V3R2-044 id: lines at the same offsets — clause: text byte-identical to live)
```

Observation: all 6 in-scope surfaces confirmed present and live==mirror identical at plan-phase. The CONST-V3R2-020 drift lives on the `clause:` line (+5 from the `id:` line), NOT on the `id:` line. CONST-V3R2-044's `clause:` line (+5 from its `id:`) is the pure-conclusion clause (no drift token), out of scope.

## §5. Sibling-pattern reference

SPEC-CC2178-TEAM-API-ALIGN-001 (completed, origin 8b1ea95eb) is the canonical precedent: a CC-version doctrine-alignment SPEC of the same shape (a CC changelog delta drove a doctrine-prose correction across live + template-mirror surfaces, with `make build` regeneration and mirror parity). This SPEC follows that pattern.

## §6. Cross-references

- Full harness research: `.moai/research/cc-update-2.1.183-to-2.1.186.md` (T1-1 GENUINE-DRIFT; everything else NO-OP/informational)
- Official doc to re-confirm at run time: `code.claude.com/docs/en/sub-agents`
- CC canonical changelog: `raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md`
- Template-First / neutrality doctrine: CLAUDE.local.md §2 / §2.1 / §15 / §25
- Mirror parity invariant: `internal/template/embedded_mirror_test.go`
- Verification-claim integrity: `.claude/rules/moai/core/verification-claim-integrity.md`
