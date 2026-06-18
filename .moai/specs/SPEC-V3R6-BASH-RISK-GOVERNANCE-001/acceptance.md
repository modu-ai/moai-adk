# Acceptance — SPEC-V3R6-BASH-RISK-GOVERNANCE-001

> Acceptance criteria for the Bash Risk-Amplifier Doctrine. Tier S, Era V3R6. Each AC maps to one or more REQ-BRG-* requirements in spec.md §B.

## §D. Acceptance Criteria Matrix

### AC-BRG-001 — Bash risk-tier classification present in doctrine

**Requirement**: REQ-BRG-001 (Ubiquitous).

**Given** the file `.claude/rules/moai/development/coding-standards.md` has been edited,
**When** a reviewer greps for the Bash risk-tier classification,
**Then** the doctrine MUST classify Bash as risk-tier "write/irreversible" by default, DISTINCT from Read/Glob/Grep risk-tier "read", AND cite book1 ch04 (Bash as risk amplifier / 风险放大器) AND the appendix A.3 litmus.

**Observable evidence**:
```bash
grep -niE 'write.irreversible|risk.amplifier|风险放大器|appendix.a.3|litmus' .claude/rules/moai/development/coding-standards.md
# Expected: ≥4 hits covering (1) the write/irreversible tier, (2) the risk-amplifier term, (3) the appendix A.3 litmus citation, (4) the read-tier distinction.
```

### AC-BRG-002 — Named soft-cap constant + compound-command doctrine

**Requirement**: REQ-BRG-002 (State-driven).

**Given** the doctrine section is authored,
**When** a reviewer greps for the soft-cap constant,
**Then** the doctrine MUST define a named, tunable threshold constant `BASH_SUBCOMMAND_SOFT_CAP = 5` (NOT prose like "around five"), AND state that exceeding it triggers a warn condition (split-to-script OR delegate), AND state that the cap is SOFT (warn-only, not hard-fail).

**Observable evidence**:
```bash
grep -nE 'BASH_SUBCOMMAND_SOFT_CAP' .claude/rules/moai/development/coding-standards.md
# Expected: ≥1 hit with the literal constant definition.

grep -niE 'soft|warn' .claude/rules/moai/development/coding-standards.md | grep -i bash
# Expected: ≥1 hit stating the cap is soft / warn-only.
```

### AC-BRG-003 — Destructive-primitive confirmation doctrine

**Requirement**: REQ-BRG-003 (Event-driven).

**Given** the doctrine section is authored,
**When** a reviewer greps for destructive primitives,
**Then** the doctrine MUST list the destructive-primitive set (`rm -rf`, `git push --force`, `git push --no-verify`, `git reset --hard`, `DROP TABLE`/`TRUNCATE`, `chmod -R 777`) AND require explicit confirmation EVEN in `bypassPermissions` mode AND cross-reference `CLAUDE.local.md` §19.1 (Implementation Kickoff Approval human-gate).

**Observable evidence**:
```bash
grep -nE 'rm -rf|git push --force|git push --no-verify|git reset --hard|DROP TABLE|TRUNCATE|chmod -R 777' .claude/rules/moai/development/coding-standards.md
# Expected: ≥4 distinct primitive hits.

grep -niE 'bypassPermissions|§19.1|Implementation Kickoff' .claude/rules/moai/development/coding-standards.md
# Expected: ≥1 hit cross-referencing §19.1 / bypassPermissions.
```

### AC-BRG-004 — Hook is warn-only and fail-open (NON-NEGOTIABLE)

**Requirement**: REQ-BRG-004 (event-detected form). This is the non-negotiable design constraint.

**Given** `handle-pre-tool.sh` has been edited to add the subcommand-count warn signal,
**When** a 7-segment compound Bash command is fed to the hook,
**Then** the hook MUST emit a warn line to stderr AND MUST exit 0 (fail-open). The hook MUST NOT exit non-zero, MUST NOT block the Bash call, MUST NOT instantiate the book1 ch06 death-spiral hazard.

**Observable evidence**:
```bash
# Smoke test — 7-segment compound command (exceeds soft cap of 5):
echo '{"tool_name":"Bash","tool_input":{"command":"a && b && c && d && e && f && g"}}' \
  | .claude/hooks/moai/handle-pre-tool.sh
echo "EXIT=$?"
# Expected: a stderr line containing a warn marker (e.g., [moai:bash-risk] WARN)
#           AND EXIT=0 (fail-open).
```

```bash
# Negative test — hook MUST NOT hard-fail on the warn path:
grep -nE 'exit 2|exit 1' .claude/hooks/moai/handle-pre-tool.sh | grep -v '^#'
# Expected: zero hits on the warn path (fail-open preserved).
```

### AC-BRG-005 — No regression to CLAUDE.md §7/§14 (compatibility / Non-Goal)

**Requirement**: REQ-BRG-005 (Unwanted behavior).

**Given** the doctrine section is authored,
**When** a reviewer reads the doctrine,
**Then** the doctrine MUST state it is strictly additive AND cite `CLAUDE.md §7 Safe Development Protocol` AND `CLAUDE.md §14 Parallel Execution Safeguards` as compatibility targets AND MUST NOT phrase itself as overriding, weakening, or superseding either section.

**Observable evidence**:
```bash
grep -nE '§7|Safe Development Protocol|§14|Parallel Execution Safeguards' .claude/rules/moai/development/coding-standards.md
# Expected: ≥2 hits (one for §7, one for §14).

grep -niE 'additive|supersede|override|weaken' .claude/rules/moai/development/coding-standards.md | grep -i bash
# Expected: ≥1 hit with "additive" framing; zero hits phrasing the doctrine as overriding §7/§14.
```

### AC-BRG-006 — Grep reproducibility / vocabulary introduced

**Requirement**: REQ-BRG-006 (Ubiquitous).

**Given** both files (doctrine + hook) have been edited,
**When** a reviewer runs the original gap grep,
**Then** the grep MUST return non-zero hits, closing the verified zero-hit gap from §A.2 of spec.md.

**Observable evidence**:
```bash
grep -rniE 'risk.?amplifier|subcommand.?count|subcommand.?cap|compound.*command|isConcurrencySafe' .claude/rules/moai/
# Expected: ≥3 hits (was 0 pre-SPEC; verified 2026-06-18 baseline).
```

## §D.1 Severity

All ACs are **MUST-PASS**. AC-BRG-004 (fail-open) is the non-negotiable design-constraint AC — failure here blocks the SPEC regardless of other AC status.

## §D.2 Traceability

| AC | REQ | Milestone |
|----|-----|-----------|
| AC-BRG-001 | REQ-BRG-001 | M1 |
| AC-BRG-002 | REQ-BRG-002 | M1 |
| AC-BRG-003 | REQ-BRG-003 | M1 |
| AC-BRG-004 | REQ-BRG-004 | M2 (non-negotiable) |
| AC-BRG-005 | REQ-BRG-005 | M1 |
| AC-BRG-006 | REQ-BRG-006 | M3 |

## §D.3 Edge Cases

- **EC-1 — Quoted-string false positive**: `echo "a && b && c && d && e && f"` (all inside quotes) would over-count. The counter is a heuristic warn signal, not a parser — this is acceptable AND documented as known issue B-2 in plan.md. The doctrine MUST acknowledge the heuristic nature.
- **EC-2 — `moai` binary missing**: the existing three-binary fallback chain (`moai` → `~/go/bin/moai` → `~/.local/bin/moai`) MUST still exit 0 silently. The warn signal is emitted regardless of whether the Go forwarder is found.
- **EC-3 — Non-Bash tool (Write/Edit)**: the warn logic MUST NOT fire on Write/Edit tool calls. The matcher scope stays `Write|Edit|Bash`; the subcommand count is only meaningful for Bash.
- **EC-4 — Single-segment Bash**: `git status` (0 compound metacharacters) MUST NOT trigger the warn. Count threshold is > 5, not ≥ 0.

## §D.4 Closure Gates

- All 6 ACs PASS with observable evidence (grep output reproduced independently).
- AC-BRG-004 fail-open verified by smoke test (exit code 0 on warn path).
- No regression: `go test ./...` unaffected (no Go changes), `moai spec lint` clean.
- plan-auditor verdict recorded before run-phase entry.

## §D.5 Forward-Looking Checks

- **FL-1**: if a future SPEC promotes the subcommand counter to a Go-binary-level parser, this SPEC's shell-layer warn signal becomes redundant — the future SPEC MUST supersede REQ-BRG-004's shell-layer constraint explicitly (frontmatter `superseded_by`).
- **FL-2**: if a future SPEC introduces an `isConcurrencySafe()` tool-partition taxonomy (book1 ch04 names it), this SPEC's Bash-specific risk tier becomes a subset — the future SPEC MUST reference this SPEC as the Bash-tier origin.

## §D.6 Quality Gate Criteria

- **Functionality**: doctrine is readable, hook is executable, warn signal fires on the documented condition.
- **Craft**: named constant (not prose), grep-reproducible vocabulary, book citations present.
- **Consistency**: additive to CLAUDE.md §7/§14, no contradiction with glm-web-tooling.md.

## §D.7 Definition of Done

- 6/6 ACs PASS with independently reproduced grep/smoke evidence.
- AC-BRG-004 fail-open constraint verified (the non-negotiable one).
- M1 + M2 + M3 milestones closed.
- spec-lint clean, go test unaffected, shellcheck clean (or inline-suppressed).
