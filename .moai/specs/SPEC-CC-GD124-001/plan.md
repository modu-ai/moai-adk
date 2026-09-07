# SPEC-CC-GD124-001 — Implementation Plan

Tier M · harness level minimal · prose-only rule repair, 4 files (2 rules x local/template mirror).

## §A Context

Claude Code upstream released capability/fact changes that stale-dated two ALWAYS-LOADED MoAI rule files. Verified 2026-09-07 from CHANGELOG extracts and official docs (`.moai/reports/t491/upstream-verification-20260907.md`, V1-V10):

1. GD-1 — Fable = 1M (v2.1.257); Sonnet 5 = 1M (v2.1.247). Our cw table says `Fable (256K)` and buries Sonnet 5 in `Sonnet/Opus standard (200K)`.
2. GD-2 — same-machine messaging works on EVERY provider since v2.1.248; cross-machine exclusions on those providers REMAIN. Two axes, stated separately.
3. GD-3 — native-Windows same-machine messaging since v2.1.234 (named-pipe socket). Our rule denies native Windows entirely. Cross-machine reach from native Windows: not documented — claim nothing.
4. GD-4 — same-machine messaging works with feature-flag fetching off since v2.1.248 (class-level docs statement + CHANGELOG "and when telemetry is disabled"). Class-level fact ONLY.
5. Mirror state — cw pair byte-identical (`diff -q` rc=0); csm pair differs by exactly ONE hunk: template omits the local-only line `> Origin: SPEC-CODEX-SESSION-MSG-001 (design.md §8 mapping).` (added by `9ef2b91e1`, card t187). INTENTIONAL template-neutrality divergence (§25.1 class C1) — `diff -q` rc=0 is unreachable for csm BY DESIGN; the parity criterion is "no NEW hunks beyond the known one".

Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t491` (branch `WT-cc-upstream-sweep`, base `615d18c1f`). All writes stay inside this path.

## §B Known Issues

- The csm Providers bullet currently collapses two independent axes (same-machine / cross-machine) into one blanket unavailability — factually wrong on the same-machine axis for every provider since v2.1.248.
- The cw table's error direction is asymmetric: a 200K model wrongly on a 1M row makes sessions run past the real ceiling (stall); the inverse is merely an early handoff. REQ-003 exists to protect this direction.

## §C Pre-flight

- Confirm worktree HEAD is on `WT-cc-upstream-sweep` and base is `615d18c1f` (`git -C <wt> rev-parse --short HEAD`, `git -C <wt> branch --show-current`).
- Confirm the 4 target files exist and the RED-now greps from `acceptance.md` reproduce (all four measurements were reproduced in this tree at plan phase, 2026-09-07).
- No plan-audit run at authoring (delegation instruction); run-phase may request one per normal flow.

## §D Constraints

- Prose-only change — no Go source touched except the embedded template copies under `internal/template/templates/`.
- Preserve the existing prose register of both files: dense, declarative, English.
- Keep the csm § Availability section at exactly FIVE bullets (intro line "Five constraints" must not change).
- Do not touch cw L30-34 (GLM-5.3 section incl. `(Opus=1M, Sonnet/Haiku=200K)` parenthetical); csm L23, L25, L9.
- No per-flag claims for the four env flags (GD-4 is class-level only).
- No time estimates anywhere.

## §E Self-Verification

LSP gates: N/A — prose-only change to `.md` rule files; no Go symbols added or moved. Evidence obligation: every AC in `acceptance.md` is verified by the pinned single-invocation command with its verbatim output; judge printed `grep -c` counts, NOT exit codes (grep exits 1 on a zero count — the exit code gates the wrong way on zero-match assertions).

## §F Milestones

Ordered by decision-reversibility: the threshold-table semantics (most likely to change — data-model of the rule) first; mechanical verification last.

### M1 — context-window pair (Priority High)

Files (edit identically, keep byte-identical):

- `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t491/.claude/rules/moai/workflow/context-window-management.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t491/internal/template/templates/.claude/rules/moai/workflow/context-window-management.md`

Edits:

1. `| Fable (256K) | 256,000 tokens | **90%** | ~230,000 tokens |` → `| Fable (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |`
2. `| Sonnet/Opus standard (200K) | 200,000 tokens | **90%** | ~180,000 tokens |` → `| Sonnet 5 (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |`
3. Insert immediately after the new Sonnet 5 row: `| Sonnet 4.x / earlier standard (200K) | 200,000 tokens | **90%** | ~180,000 tokens |`
4. Both occurrences of `90% on 200K/256K` (L64 user-responsibilities, L80 pre-clear announcement) → `90% on 200K`.
5. Do NOT touch L30-34 (GLM-5.3 section).

### M2 — cross-session-messaging pair (Priority High)

Files (apply the same edit to both; the only permitted remaining delta is the known Origin-line hunk):

- `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t491/.claude/rules/moai/workflow/cross-session-messaging.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t491/internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md`

Edits (all inside § Availability constraints; keep FIVE bullets):

1. OS bullet — `- **Operating system** — macOS, Windows, and Linux (Linux inside WSL 2 included). Same-machine messaging works on native Windows since Claude Code v2.1.234, where the inbox socket is a named pipe. Cross-machine reach from native Windows is not documented — an explicit gap; claim nothing either way.`
2. Providers bullet (one bullet, two axes) — `- **Providers** — Two axes. Same machine: available on every provider — Amazon Bedrock, Claude Platform on AWS, Agent Platform on Google Cloud, and Microsoft Foundry included — since v2.1.248; delivery rides a per-session socket on the machine and never leaves it. Beyond this machine: still unavailable with an API key and on Amazon Bedrock, Claude Platform on AWS, Agent Platform on Google Cloud, and Microsoft Foundry.`
3. Flags bullet — `- **Flag evaluation** — any one of \`CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC\`, \`DISABLE_TELEMETRY\`, \`DO_NOT_TRACK\`, \`DISABLE_GROWTHBOOK\` disables the feature-flag evaluation. Since v2.1.248, same-machine messaging works in sessions with feature-flag fetching off, on every provider (class-level statement — no per-flag claims for the four flags). Below v2.1.248 such sessions had no same-machine messaging; the capability is new in that release. Diagnostic: \`/list-agents\` (alias \`/peers\`) recognized → present; unrecognized → absent.`
4. Do NOT touch: the versions bullet (L23), the shared-flag-slot bullet (L25), L9 "same platforms" phrasing, any other line.

### M3 — Verification + embed sanity (Priority Medium)

1. Run every AC command in `acceptance.md`; record verbatim outputs.
2. `make build` (from the worktree root) — exit 0; embeds the template edits into the binary.
3. `go test ./internal/template/...` — exit 0.
4. `git -C <wt> status --porcelain` — only the 4 target files + `.moai/specs/SPEC-CC-GD124-001/` + `.moai/reports/t491/`.

## §G Anti-Patterns

- Editing only the local copy (or only the template copy) — flips AC7 red. Both copies, always, in the same pass.
- "Fixing" the csm diff to zero (removing the Origin line from local or adding it to template) — the known hunk is intentional template neutrality; rc=0 for csm is unreachable by design.
- Moving Haiku 4.5 or Sonnet 4.x-era rows to 1M/50% — flips the error direction toward real-ceiling overrun (REQ-003).
- Touching the GLM-5.3 parenthetical, the csm versions/shared-flag-slot bullets, or the five-bullet count.
- Retrying a failed `grep -c` as if exit 1 meant failure — judge the printed count.

## §H Cross-References

- Verification record: `.moai/reports/t491/upstream-verification-20260907.md`
- CHANGELOG extracts: `.moai/reports/t491/t491-v247.md`, `t491-v248.md`, `t491-v257.md`
- Sweep report: `.moai/reports/t491/sweep-report-copy-20260906.md`
- Template isolation doctrine: `.moai/docs/template-internal-isolation-doctrine.md` §25.1 (class C1 — SPEC IDs forbidden in templates)
- Card: t491 (Factory Mode, lane-2)

## Plan notes for the lead (non-blocking, recorded at plan phase)

- (a) The csm pair can never reach `diff -q` rc=0 — parity is modulo the known neutrality hunk. Any audit surface that demands byte-identity for csm is misreading the doctrine; the correct criterion is AC7's "exactly 1 hunk, Origin-line only".
- (b) GD-4 is verified STRONGER than the sweep's partial verdict: the official docs carry a class-level statement and CHANGELOG 2.1.248 reads "and when telemetry is disabled" (2026-09-07). This supersedes the sweep's "3 flags unobserved" caveat at the class level — the class-level fact is sufficient and per-flag claims remain prohibited.
