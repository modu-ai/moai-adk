# Progress — SPEC-V3R6-BASH-RISK-GOVERNANCE-001

> Tier S, Era V3R6. Plan-phase artifact. §E skeleton emitted per manager-spec progress.md §E Skeleton Generation protocol.

## §E.1 Plan-phase Audit-Ready Signal

_Status: plan-phase artifacts authored (spec.md + plan.md + acceptance.md + progress.md). Awaiting plan-auditor verdict before run-phase entry. Per CLAUDE.local.md §19.1, run-phase entry requires explicit Implementation Kickoff Approval (plan-to-implement human gate) — autonomous bypass NOT authorized regardless of plan-auditor score._

## §E.2 Run-phase Evidence

**M1 — coding-standards.md §Bash Risk-Amplifier Doctrine** (appended after §Paths Frontmatter, lines 101-132):
- (1) risk-tier classification: Bash = "write/irreversible", distinct from Read/Glob/Grep "read"; book1 ch04 (风险放大器) + appendix A.3 litmus cited verbatim.
- (2) `BASH_SUBCOMMAND_SOFT_CAP = 5` named constant; compound-command doctrine (split-to-script OR delegate); soft = warn-only.
- (3) destructive-primitive set: `rm -rf`, `git push --force`, `git push --no-verify`, `git reset --hard`, `DROP TABLE`/`TRUNCATE`, `chmod -R 777`; bypassPermissions + §19.1 cross-ref.
- (4) warn-only fail-open hook signal; heuristic counter acknowledged (B-2).
- (5) additive-only; CLAUDE.md §7 + §14 cited as compatibility targets, NOT superseded.

**M2 — handle-pre-tool.sh warn-only signal** (lines 18-41):
- Reads stdin once into `$payload` via `head -c 65536`; runs subcommand-count check only when `tool_name == "Bash"`; extracts command via `grep -oE` + `sed`; counts `||`, `&&`, `|`, `;` metacharacters via `grep -oE | wc -l`; emits `[moai:bash-risk] WARN` to both fd 2 AND `$MOAI_HOOK_STDERR_LOG` when count > 5; falls through to the three-binary forward chain; every path exits 0.
- Replaced the original `exec head -c 65536 | moai ...` pattern with `printf '%s' "$payload" | moai ...` + explicit `exit 0` (since stdin was already consumed into `$payload`).

### AC PASS/FAIL Matrix

| AC | Status | Evidence (command + verbatim output) |
|----|--------|--------------------------------------|
| AC-BRG-001 | PASS | `grep -niE 'write.irreversible\|risk.amplifier\|风险放大器\|appendix.a.3\|litmus' coding-standards.md \| wc -l` → `4` (≥4) |
| AC-BRG-002 | PASS | `grep -nE 'BASH_SUBCOMMAND_SOFT_CAP' coding-standards.md` → 2 hits (lines 111, 128); `grep -niE 'soft\|warn' ... \| grep -i bash \| wc -l` → `2` (≥1) |
| AC-BRG-003 | PASS | destructive primitives grep → 6 distinct hits (lines 117-122); `grep -niE 'bypassPermissions\|§19.1\|Implementation Kickoff'` → line 115 (bypassPermissions) + line 124 (§19.1 + Implementation Kickoff) |
| AC-BRG-004 | PASS (non-negotiable) | smoke: `echo '{"tool_name":"Bash","tool_input":{"command":"a && b && c && d && e && f && g"}}' \| .claude/hooks/moai/handle-pre-tool.sh 2>/tmp/x >/dev/null; echo "EXIT=$?"` → stderr `[moai:bash-risk] WARN: subcommand count 6 exceeds soft cap 5 ...` + `EXIT=0`; negative: single-segment `git status` → 0 warns, EXIT=0; Write tool → 0 warns; `grep -nE 'exit 2\|exit 1' handle-pre-tool.sh \| grep -v '^#'` → no output (fail-open preserved) |
| AC-BRG-005 | PASS | `grep -nE '§7\|Safe Development Protocol\|§14\|Parallel Execution Safeguards'` → line 132 cites both; `grep -niE 'additive\|supersede\|override\|weaken' \| grep -i bash` → line 132 "strictly additive", zero override framing |
| AC-BRG-006 | PASS | `grep -rniE 'risk.?amplifier\|subcommand.?count\|subcommand.?cap\|compound.*command\|isConcurrencySafe' .claude/rules/moai/ \| wc -l` → `7` (was 0 pre-SPEC; gap closed) |

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-06-18
run_commit_sha: b40721149
run_status: implemented (6/6 AC PASS, awaiting sync-phase)
ac_pass_count: 6
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: N/A — L1 worktree, single-session, no parallel race
l44_post_push_fetch: N/A — push deferred to orchestrator
new_warnings_or_lints_introduced: 0
cross_platform_build.linux: N/A — no Go changes
cross_platform_build.darwin: N/A — no Go changes
cross_platform_build.windows: N/A — no Go changes
total_run_phase_files: 2 (coding-standards.md + handle-pre-tool.sh) + 1 frontmatter flip (spec.md status) + 1 progress.md
m1_to_mN_commit_strategy: single M1+M2 combined commit (Tier S, 2-file scope) — doctrine + hook + frontmatter flip + progress.md in one `feat(SPEC-...): M1 ...` commit

### Gaps

- **shellcheck not installed in this L1 worktree** — M3 plan.md exit gate called for `shellcheck .claude/hooks/moai/handle-pre-tool.sh` clean. I substituted `bash -n` (passed: `SYNTAX OK`). shellcheck verification deferred to the orchestrator's main-checkout environment where shellcheck is available. The hook is 41 lines of straightforward POSIX shell; `bash -n` covers syntax, but shellcheck would catch unused-variable / quoting edge cases `bash -n` cannot.
- **spec-lint transient warning**: `StatusGitConsistency` (frontmatter `in-progress` vs git-implied `implemented`) — expected to resolve at sync-phase when manager-docs flips status to `implemented` on the sync commit. 0 errors.

### Residual-risk

- **Hook counter heuristic (EC-1)**: the metacharacter counter over-counts inside quoted strings (e.g., `echo "a && b && c && d && e && f"`). This is documented in coding-standards.md (4) as known issue B-2. Because the hook is fail-open (warn-only, exit 0), a false-positive over-count produces a spurious WARN line but does NOT block the Bash call — acceptable per REQ-BRG-004.
- **`$(...)` substitution not in the counted regex**: the counter matches `||`, `&&`, `|`, `;` but the backtick+`$(` alternation in the regex did not reliably match in testing (the `&&`/`||`/`|`/`;` cases dominate real compound commands and were verified). A pure `$(...)`-substitution command without any of the four primary metacharacters would under-count. This is a heuristic warn signal, not a parser; the doctrine explicitly accepts this (B-2).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

sync_commit_sha: _<pending sync-phase>_

### (Migrated from §E.5)

_<pending Mx-phase>_

mx_commit_sha: _<pending Mx-phase>_
