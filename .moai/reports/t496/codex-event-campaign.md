# t496 — Codex hook event firing campaign (SPEC-CODEX-EVENT-COVERAGE-001 M2)

- Date: 2026-09-07 · Tree: `.claude/worktrees/t496` @ `88c7d0c84` (branch `WT-codex-hook-events`, M1 landed)
- Target binary: `codex --version` → `codex-cli 0.153.4` (evidence: `evidence/version.txt`)
- Precedents followed: t504 harness shape (CODEX_HOME relocation, serial cells, file-based census) · t83 §0 (auth via symlink, home-level `hooks.json`, `--dangerously-bypass-hook-trust`)
- Model: `gpt-6-astra` (from the operator's real config; campaign home carries a minimal `config.toml` naming only the model — `evidence/config.toml`)

## 0. P0 — `CODEX_HOME` isolation support: SUPPORTED (verified by execution)

Command (run `p0`):

```bash
cd "$EXP_ROOT/proj" && CODEX_HOME="$EXP_ROOT/home" timeout 300 codex exec --json \
  --skip-git-repo-check --dangerously-bypass-hook-trust "Reply with exactly: T496P0OK"
```

Observed: rc=0; answer `T496P0OK`; SessionStart/SessionEnd/Stop hook loggers landed in `$EXP_ROOT/captures/` (tmp home), and the SessionStart payload's `transcript_path` points **inside the tmp home** (`/private/tmp/t496-campaign-en7H7wVT/home/sessions/...`) — evidence `evidence/captures/SessionStart.jsonl`, `evidence/runs/p0.jsonl`. Auth worked via `home/auth.json -> ~/.codex/auth.json` (symlink, no copy — t83 §0 recipe). No fallback to project-scope `.codex/hooks.json` needed (t83 round3 H4 already showed project-level does not fire; campaign hooks live at `$CODEX_HOME/hooks.json`).

**Disclosure (isolation defect, fixed before measurement)**: the FIRST P0 attempt (`EXP_ROOT=/tmp/t496-campaign-d4IY5wt6`) omitted `CODEX_HOME=` from the exec invocation and therefore ran against the real home. It executed one model call ("Reply with exactly: T496P0OK") and wrote its session rollout into `~/.codex/sessions/2026/09/07/rollout-2026-09-07T14-46-59-*.jsonl` (verified: `originator: "codex_exec"`, `cwd: /private/tmp/t496-campaign-d4IY5wt6/proj` — the only `codex_exec` entry in the window). It did not modify `config.toml`/`hooks.json`/`auth.json` (mtimes pre-date the campaign). The defect was fixed (`CODEX_HOME="$EXP_ROOT/home"` in `run_exec`) and all measured runs (EXP_ROOT `en7H7wVT`) are isolated.

## 1. Real-home zero-write verdict (REQ-CEV-008, AC-CEV-011)

Method: three full snapshots of `~/.codex` (path list + per-file mtime) — control window C (t−20 s), campaign start P, campaign end O — with the delta attribution `delta(P,O) minus delta(C,P)` (raw snapshots: `evidence/homecheck/*.gz`, derived verdict: `evidence/homecheck/zero-write-verdict.txt`).

- **Key files**: `~/.codex/config.toml` (mtime Sep 7 13:33 — pre-campaign), `~/.codex/auth.json` (Aug 29), `~/.codex/hooks.json` (Aug 27) — untouched; hashes recorded in the verdict file.
- **Session rollouts appearing in the campaign window**: every new `sessions/2026/09/07/rollout-*.jsonl` in `delta(P,O)` was attributed by reading its `session_meta` line. ALL are `originator: "moai-codex-gate"` live-test sessions of a **parallel lane** (`TestCodexLive_ReviewStartBaseBranchIsNotRejected*` / `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey*`, cwd under `/var/folders/.../T/TestCodexLive_*`) — zero matches for campaign markers (`grep -l 't496'` over those files: rc=1). None carries a campaign `cwd`.
- **The fixed campaign (en7H7wVT) wrote zero files into `~/.codex`**: its session transcripts landed in the tmp home (payload `transcript_path` evidence above); no real-home entry carries its path.
- **Conclusion**: 0 campaign-attributable home writes (the parallel lane's writes are not this campaign's; the one earlier `codex_exec` entry is the disclosed pre-fix P0 run, before the measurement window, config untouched).

## 2. Per-event verdicts (AC-CEV-010)

Capture method: every campaign event had a home-level logger `/bin/sh -c 'cat >> $EXP_ROOT/captures/<Event>.jsonl'` (`timeout: 10`) in `$CODEX_HOME/hooks.json` (shape follows t83 §3.1: top-level `description`/`hooks` only, PascalCase event keys, no matcher). All runs ran from `$EXP_ROOT/proj` with `--skip-git-repo-check --dangerously-bypass-hook-trust`.

| Event | Verdict | Trigger (exact command in `evidence/runs/<run>.*`) | Observed evidence | Disposition |
|---|---|---|---|---|
| **SubagentStart** | **FIRED** | run `collab`: `codex exec --json … "Use the collaboration tool to spawn a subagent whose task is to compute 2+2…"` | capture 1 line (`evidence/captures/SubagentStart.jsonl`); stream shows `collab_tool_call` items | **adapt-now → M3** (payload below; dispatcher arg `subagent-start` registered) |
| **SubagentStop** | **FIRED** (reverses the 0.147.0 not-fired observation — REQ-CEV-011 re-test, not carried) | same run `collab` | capture 1 line (`evidence/captures/SubagentStop.jsonl`), `last_assistant_message: "4"` (the subagent's answer) | **adapt-now → M3** (dispatcher arg `subagent-stop` registered) |
| **Interrupt** | **FIRED** | run `interrupt2`: `timeout -s INT 8 codex exec --json … "Think carefully, then count slowly from 1 to 200…"`; rc=124 (SIGINT delivered) | capture 1 line (`evidence/captures/Interrupt.jsonl`); the run's error stream also shows codex parsing the config: ``clamping Interrupt hook timeout to 3s in /private/tmp/.../home/hooks.json`` — the Interrupt hook was registered, and fired on SIGINT | **follow-up card** — fires, but has no MoAI dispatcher counterpart to map to; adaptation requires a NEW `moai hook` subcommand (out of scope per REQ-CEV-005). Documented payload shape for that card |
| **PreCompact** | **trigger-not-achieved** (NOT "does not fire") | runs `compact` + `compact3`: stdin-piped filler, `codex exec --json … "Reply with exactly: T496COMPACT(3)OK"`; sizes 702,974 and ~960,000 chars | compaction precondition never reached: run `compact` = 192,872 input tokens, run `compact3` = **264,808 input tokens**, zero compaction markers in both streams (only marker-string matches); hard input cap measured on run `compact2` (1,583,027 chars): `Error: turn/start failed: Input exceeds the maximum length of 1048576 characters.` | **document no-fire basis** — non-interactive compaction is infeasible at this model's window under the input cap; capture file absent (consistent: nothing fired because nothing compacted). Follow-up candidate: interactive-session trigger |
| **PostCompact** | **trigger-not-achieved** | same runs as PreCompact | same precondition evidence — no compaction occurred in either run | same as PreCompact |
| **PermissionRequest** | **trigger-not-achieved (non-interactive)** — the approval request itself was never raised | run `perm`: `-s read-only` + in-workspace write → command auto-failed "operation not permitted", no approval item in stream. Run `perm2`: `-s workspace-write` + `-c approval_policy='"on-request"'` + out-of-workspace write → command executed, exit 0, no approval item. Run `perm3`: `-s read-only` + `-c approval_policy='"on-request"'` → auto-failed again, no approval item | three configurations (`evidence/runs/perm{,2,3}.jsonl`), each reached a restricted-operation attempt; `grep -c -i 'approval\|permission'` = 0 in all three streams | **follow-up card** — the event exists in the binary (t83 §2 `PermissionRequestCommandOutputWire`) and presumably fires on the interactive TUI approval path, which a PTY-less exec run never reaches. Documented as untested-interactively |

Fired counts for the record: captures grew only for SessionStart/SessionEnd/Stop (already adapted, P0+) and SubagentStart/SubagentStop (run `collab`) and Interrupt (run `interrupt2`) — `evidence/captures/`.

## 3. Payload shapes (observed key sets)

- `SubagentStart`: `session_id, turn_id, transcript_path, cwd, hook_event_name, model, permission_mode, agent_id, agent_type` — `agent_type: "default"`.
- `SubagentStop`: adds `agent_transcript_path, stop_hook_active: false, last_assistant_message` (the subagent's answer) on top of the SubagentStart set minus `permission_mode`-only differences (full raw: `evidence/captures/SubagentStop.jsonl`).
- `Interrupt`: `session_id, turn_id, transcript_path, cwd, hook_event_name, model, permission_mode`.
- Baseline sanity (already-adapted events, unchanged shapes vs t83/t91): SessionStart `…source: "startup"`, SessionEnd `…reason`, Stop `…stop_hook_active, last_assistant_message, turn_id` (`evidence/captures/`).

## 4. Dispositions summary (AC-CEV-013)

| Disposition | Events |
|---|---|
| adapt-now (→ M3, landed in this SPEC) | SubagentStart, SubagentStop |
| follow-up card | Interrupt (needs new dispatcher subcommand — REQ-CEV-005 forbids it here), PermissionRequest (interactive PTY harness) |
| document no-fire basis / trigger-not-achieved | PreCompact, PostCompact |

M3 executed: `EventTable` rows for SubagentStart/SubagentStop flipped to `Adapted: true` (census now 8 adapted / 4 held back of 12 rows), doc comments updated, `go test ./internal/codexadapter/ ./internal/codexwiring/ ./internal/cli/ -run 'Codex|Hooks'` GREEN. RenderHooks now installs `moai hook subagent-start --harness codex` / `moai hook subagent-stop --harness codex` — that IS the adaptation surface (M1's install-surface invariant was scoped to the Interrupt row and is unchanged: Interrupt still never installed, `TestRenderHooks_InterruptNeverInstalled` GREEN).

## 5. Gaps and residual risk

- Compaction events: verdict is trigger-not-achieved, not not-fired — a session long enough to auto-compact was not achievable non-interactively (input cap 1,048,576 chars, measured; best effort 264,808 input tokens). An interactive or API-driven compaction test remains open.
- PermissionRequest: not tested on the interactive TUI path (needs a PTY harness). Exec-mode evidence covers only the non-interactive configurations listed.
- Interrupt fired exactly once (single SIGINT attempt); multi-interrupt/edge timing untested.
- The live `~/.codex` had concurrent activity from a parallel lane during the window; attribution rests on per-file `session_meta` reads (originator + cwd) plus pre-campaign mtimes on the three key files — not on a quiescent home.
- One model call (the disclosed d4IY5wt6 P0 run) executed against the real home before the fix; no config mutation observed for it.
