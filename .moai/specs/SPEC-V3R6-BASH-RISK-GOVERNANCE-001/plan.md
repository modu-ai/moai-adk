# Plan — SPEC-V3R6-BASH-RISK-GOVERNANCE-001

> Implementation plan for the Bash Risk-Amplifier Doctrine. Tier S (standard-small — single domain, 2-3 files). Era V3R6.

## §A. Context

Sprint 15 harness-books application cohort (P1b). Source doctrine: `github.com/wquguru/harness-books` book1 ch04 (Bash as 风险放大器 / risk amplifier) + appendix A.3 litmus + §9.4. The verified gap: `grep -rniE 'risk.?amplifier|subcommand.?count|subcommand.?cap|compound.*command|isConcurrencySafe' .claude/rules/moai/` returns zero hits; the PreToolUse hook (`handle-pre-tool.sh`) fires uniformly on `Write|Edit|Bash` with no Bash-specific risk tier or compound-command cap.

## §B. Known Issues

- **B-1 (hook is a thin wrapper)**: `handle-pre-tool.sh` (33 lines) is a forwarder to `moai hook pre-tool` (Go binary). The warn-only subcommand counter must live at the shell layer BEFORE the `exec` forward, because (a) REQ-BRG-004 forbids changing the Go-binary block logic, and (b) the shell layer can emit a stderr warning and still `exec` into the Go forwarder without changing the exit contract.
- **B-2 (subcommand counting is approximate)**: shell metacharacter counting (`|`, `&&`, `||`, `;`, backtick, `$(`) over-counts in edge cases (e.g., `echo "a && b"` inside a quoted string). The counter is a HEURISTIC warn signal, not a parser. This is acceptable for a warn-only doctrine; it MUST NOT be the basis of a hard block.
- **B-3 (5s timeout budget)**: the subcommand-count check MUST be trivially fast (single `grep -oE` + `wc -l`) to stay well within the 5s hook timeout. No `awk`/`sed`/external-process pipelines.

## §C. Pre-flight

- [ ] Confirm grep-zero-hit baseline: `grep -rniE 'risk.?amplifier|subcommand.?count|subcommand.?cap|compound.*command|isConcurrencySafe' .claude/rules/moai/` returns zero hits (verified 2026-06-18 pre-plan).
- [ ] Confirm `handle-pre-tool.sh` is 33 lines, forwards to `moai hook pre-tool`, exits 0 on missing binary (verified 2026-06-18 pre-plan).
- [ ] Confirm `coding-standards.md` is 99 lines with no existing `## Bash` section (verified 2026-06-18 pre-plan).

## §D. Constraints (carried from spec.md §D)

- WARN-only, fail-open (non-negotiable).
- Soft cap = warn threshold, not hard limit.
- Preserve 5s timeout + `Write|Edit|Bash` matcher scope.
- Additive only (no CLAUDE.md §7/§14 contradiction).
- Named constant (`BASH_SUBCOMMAND_SOFT_CAP = 5`).
- Book citations required (book1 ch04 + appendix A.3).

## §E. Self-Verification (M3 gate)

- [ ] `grep -rniE 'risk.?amplifier|subcommand|BASH_SUBCOMMAND_SOFT_CAP' .claude/rules/moai/development/coding-standards.md` returns ≥3 hits (closes the grep-zero-hit gap, REQ-BRG-006).
- [ ] `grep -n 'BASH_SUBCOMMAND_SOFT_CAP' .claude/hooks/moai/handle-pre-tool.sh` returns ≥1 hit (named constant present in hook).
- [ ] `grep -nE 'exit 0' .claude/hooks/moai/handle-pre-tool.sh` shows the hook still exits 0 on warn (fail-open preserved).
- [ ] Hook smoke test: `echo '{"tool_name":"Bash","tool_input":{"command":"a && b && c && d && e && f && rm -rf x"}}' | .claude/hooks/moai/handle-pre-tool.sh; echo "EXIT=$?"` prints the warn line AND `EXIT=0`.
- [ ] Doctrine does NOT mention hard-fail, block, or exit-non-zero on the subcommand warn path.

## §F. Milestones

### M1 — coding-standards.md §Bash Risk-Amplifier Doctrine (Priority High)

Author the new section in `.claude/rules/moai/development/coding-standards.md` covering REQ-BRG-001 through REQ-BRG-003 + REQ-BRG-005 + REQ-BRG-006:

- (a) Bash = risk-tier "write/irreversible", distinct from Read/Glob/Grep "read". Cite book1 ch04 + appendix A.3 litmus verbatim framing.
- (b) Compound subcommand soft cap: `BASH_SUBCOMMAND_SOFT_CAP = 5` named constant. Warn → split to script OR delegate. Soft = warn-only.
- (c) Destructive primitives list (`rm -rf`, `git push --force`, `git push --no-verify`, `git reset --hard`, `DROP TABLE`/`TRUNCATE`, `chmod -R 777`) require explicit confirmation even in `bypassPermissions`. Cross-ref CLAUDE.local.md §19.1.
- (d) Additive-only statement: explicitly cite CLAUDE.md §7 Safe Development Protocol + §14 Parallel Execution Safeguards as compatibility targets, NOT superseded.

**Exit gate**: grep reproducibility — `grep -rniE 'risk.?amplifier|subcommand' .claude/rules/moai/development/coding-standards.md` returns ≥3 hits.

### M2 — handle-pre-tool.sh warn-only signal (Priority High)

Edit `.claude/hooks/moai/handle-pre-tool.sh` to add a subcommand-count check BEFORE the `exec` forward:

- Parse the Bash `command` field from stdin JSON (best-effort, no `jq` dependency — use `grep -oE` on the raw stdin payload, or read stdin into a var first then forward).
- Count compound metacharacters (`|`, `&&`, `||`, `;`, backtick, `$(`).
- If count > `BASH_SUBCOMMAND_SOFT_CAP` (5): emit a stderr warn line (e.g., `[moai:bash-risk] WARN: subcommand count N exceeds soft cap 5 — consider splitting into a script`).
- **Exit 0 regardless** (fail-open). Then proceed to the existing `exec` forward.
- Preserve the 5s timeout, the `Write|Edit|Bash` matcher scope, the `head -c 65536` cap, and the three-binary fallback chain (`moai` → `~/go/bin/moai` → `~/.local/bin/moai`).

**Non-negotiable**: the hook MUST exit 0 even when the warning fires. No `exit 2`, no block.

**Exit gate**: smoke test — a 7-segment compound Bash command produces the warn line AND `EXIT=0`.

### M3 — lint/clean + grep reproducibility (Priority Medium)

- `shellcheck .claude/hooks/moai/handle-pre-tool.sh` clean (or inline-suppressed with reason).
- `moai spec lint SPEC-V3R6-BASH-RISK-GOVERNANCE-001` clean.
- `go test ./...` unaffected (no Go changes).
- Final grep reproducibility: the zero-hit gap is closed across `.claude/rules/moai/`.

## §G. Anti-Patterns

- **AP-BRG-001 — Hard-fail on subcommand count**: emitting `exit 2` or blocking the Bash call when count exceeds the soft cap. Violates REQ-BRG-004 non-negotiable fail-open constraint AND instantiates the book1 ch06 death-spiral hazard.
- **AP-BRG-002 — Burying the cap value in prose**: writing "around five subcommands" instead of a named `BASH_SUBCOMMAND_SOFT_CAP = 5` constant. Violates REQ-BRG-002.
- **AP-BRG-003 — Contradicting CLAUDE.md §7/§14**: phrasing the doctrine as overriding existing safeguards. Violates REQ-BRG-005.
- **AP-BRG-004 — Hook timeout regression**: introducing an `awk`/`sed`/multi-process pipeline in the subcommand counter that eats into the 5s budget. Violates the preserve-timeout constraint.
- **AP-BRG-005 — Go-binary block logic changes**: modifying `internal/cli/hook/` to add a hard block. Out of scope (§C.2) AND violates fail-open.
- **AP-BRG-006 — Quoted-string false positives without acknowledgment**: treating the heuristic counter as a precise parser. Mitigate by documenting B-2 (heuristic, warn-only) in the doctrine body.

## §H. Cross-References

- spec.md §B REQ-BRG-001..006 — requirement mapping.
- acceptance.md §D AC-BRG-001..006 — AC mapping.
- CLAUDE.local.md §19.1 — Implementation Kickoff Approval human-gate.
- CLAUDE.md §7, §14 — compatibility targets.
- `.claude/rules/moai/core/glm-web-tooling.md` — only other tool-category special-case (out of scope).
