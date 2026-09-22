# t221 — `moai session current` answered with a foreign session's id

**Card**: t221 (Class B, cause established below) · **Branch**: `WT-session-id-sidecar` · **Worktree**: `.claude/worktrees/t221`

## Cause, established

`.moai/state/current-session-id.txt` is **one slot per project**. The SessionStart hook overwrites it unconditionally (`internal/hook/session_start.go:313`), and `resolveCurrentSessionID` (`internal/cli/session.go`) read that file as its first and only real source. In a checkout with several sessions, every session read back whichever id was written last.

The precise shape matters, and the card states it correctly: the key was **not a function of the asking session**. It was right sometimes, and nothing on the read side separated a right answer from a wrong one.

Reproduced in this session, in the primary checkout:

```
$ echo $CLAUDE_CODE_SESSION_ID          # this session, from the runtime
29ac14db-a1f0-4e0a-964b-fe31571f44d6
$ cat .moai/state/current-session-id.txt  # the shared slot
29ac14db-a1f0-4e0a-964b-fe31571f44d6      # matches only because this session started last
```

and from inside the worktree, where the slot does not exist at all:

```
$ moai session current                    # shipped binary
source_session_id: <not-available — environment-fallback, ...>
```

Two failures in one read path: a foreign id in a shared checkout, and no id at all from a worktree.

## The fix — the authoritative source already exists

Claude Code stamps `CLAUDE_CODE_SESSION_ID` into every subprocess it spawns, matching the `session_id` it hands hooks on stdin (CC changelog, recorded in `.moai/research/cc-changelog-snapshot-2.1.233.md`: *"Added `CLAUDE_CODE_SESSION_ID` environment variable to the Bash tool subprocess environment, matching the `session_id` passed to hooks"*). MoAI sets it nowhere — verified by grep across the tree.

That value travels with the process, so it names the asking session by construction. The code already anticipated it: `resolveCurrentSessionID` carried a comment reading *"Stage 2 (future): env var if the runtime ever exposes session.id"*. It does now.

Precedence is now **env → side-channel file → canonical fallback**. Nothing was removed: the file remains the degraded path for a runtime that does not stamp the variable, and `--session` still overrides everything.

| File | Change |
|---|---|
| `internal/config/envkeys.go` | `EnvClaudeCodeSessionID` constant (§14: no inline env names) |
| `internal/cli/session.go` | Stage 0 env read in `resolveCurrentSessionID`; new `sessionIDSourceIsAuthoritative`; `session current` help + `session doctor` now report which source answered |
| `internal/cli/goal.go` | The multi-session warning is skipped when the id came from the env — concurrency cannot make that id foreign |
| `internal/cli/integration.go` | `integrationSessionID` prefers the env var: a release-integration lock recorded under a foreign id is held by nobody who can release it |
| `.claude/rules/.../goal-directive-detail.md` + template mirror | Arming doctrine rewritten to the new precedence |

Every other sidecar consumer routes through `resolveCurrentSessionID` and was fixed by the one change: `moai session current`, goal arm (`goal.go`), worktree branch naming (`resolveSessionShortReal`), the block-cap launcher, and the MCP arm path.

## Verification

Execution-based, not grep — `internal/cli/session_env_attribution_test.go`:

- `TestResolveCurrentSessionID_EnvNamesTheAskingSession` — two simulated sessions share one project dir whose single sidecar slot holds a *third*, foreign id; each resolves its own id. This is the regression the card asked for.
- `TestResolveCurrentSessionID_SidecarStaysAsDegradedPath` / `_NoSourceFallsBack` — the removed-nothing guarantee.
- `TestResolveArmSessionID_EnvSkipsMultiSessionWarning` — two registry entries + a foreign slot, and the arm path still arms under this session with no warning.
- `TestIntegrationSessionID_EnvBeatsSidecar` — lock identity, with `--session` still winning.

**A finding worth carrying forward.** Ten existing tests began failing the moment the env var became authoritative — they staged the sidecar as their fixture while a *real* session id sat in the process environment, so they measured the developer's session instead. They passed in CI (no var) and failed locally. Each now calls `scrubSessionIDEnv(t)`; this is the environment-control lesson (`feedback_reproduction_must_control_environment`) reappearing in a new place.

End-to-end, with a binary built from this tree, run inside this session:

```
$ moai session current --json
{"session_id": "29ac14db-a1f0-4e0a-964b-fe31571f44d6", "source": "env:CLAUDE_CODE_SESSION_ID", "available": true}
```

The id matches the `source_session_id` this session's SessionStart hook injected — the first time the two agree from a worktree.

| Check | Command | Result |
|---|---|---|
| Build | `go build ./...` | exit 0 |
| Vet | `go vet ./internal/cli/ ./internal/config/` | exit 0 |
| Full package | `go test ./internal/cli/ -count=1` | `ok … 770.995s` |
| Affected subset (after the doc/format edits) | `go test ./internal/cli/ -run 'Session\|Goal\|Integration\|CC_' -count=1` | `ok … 2.628s` |
| Neighbours | `go test ./internal/config/... ./internal/session/... ./internal/hook/...` | all `ok` |
| Template parity + neutrality | `go test ./internal/template/...` | `ok … 70.006s` |

## Gaps and residual risk

- **Not measured: two live sessions.** The regression test simulates concurrency with `t.Setenv`; no two real Claude Code sessions were run side by side against this binary. The mechanism is per-process env inheritance, so the simulation is faithful, but that is reasoning, not an observation.
- **Windows unverified at runtime.** No platform-specific code was added (`os.Getenv` only); CI's cross-build covers compilation.
- **A runtime that stops stamping the variable** silently falls back to the shared slot — the old behaviour, warning intact. `moai session doctor` now names which source is live, so the degrade is visible rather than inferred.
- **Out of scope, deliberately**: `SPEC-KANBAN-RECORD-SESSION-KEY-001` (t207-D), the parent-launcher keying of `kanban.Record`. Different surface, as the card says. One observable side effect here: `moai cc --kanban` now keys its record by the real session id when the env var is present, which narrows that defect without addressing it.
- **The hook still writes the sidecar.** Kept on purpose — it is the degraded path's only source. Retiring it is a separate decision.
