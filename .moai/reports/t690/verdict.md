# Card t690 Verdict — GH #1675 re-verification: gh-call counting on v3.1.2 vs develop

Date: 2026-09-13 · Lane: lane-3 · Branch: `WT-statusline-gh-count` (base `fac132d38` = local develop)
Card: t690 (GH #1675 re-verification, Tier S, Class B)

## Claim

1. **The report's central condition does NOT reproduce on the shipped v3.1.2 binary.** With the exact reported config (`statusline.forge: none` + `segments.github: false`), the released v3.1.2 artifact makes **0 gh calls**, with AND without the fallback-profile environment (`CLAUDE_CONFIG_DIR` pointing at a copy of the real `__no_such_profile__` dir). The measurement is proven live by positive controls: without the yaml, both binaries make gh calls (2 per render round).
2. **develop also makes 0 gh calls** under the same reported conditions — the repairs (t293 `62485c918`/`5a193fa4c`, t297 `5ed9d2d75`) hold.
3. **NEW finding — v3.1.2's segment flag gated the DRAWING, not the POLLING.** With `segments.github: false` alone (no `forge:` key), v3.1.2 still makes 2 gh calls per round; the polling stopper in v3.1.2 was `forge: none`. In develop, the segment gate reaches the spawn (REQ-001, SPEC-STATUSLINE-PROFILE-RESPECT-001) — this gap is already repaired.
4. **The `0/0` display path** is `internal/statusline/renderer.go:664` — `fmt.Sprintf(", %d/%d", data.GitHub.OpenIssues, data.GitHub.OpenPRs)` over the cached counts; a failed fetch preserves zeros, and the segment draws them.
5. **`__no_such_profile__` is not produced by any code path in the repo** (grep over internal/ pkg/ cmd/ templates/ scripts/: 0 hits). It is on-disk residue of an earlier probe-era defect, and the current code cannot select it (details below). No repair owed in code; the residue is operator-disposable.

## Evidence

Measurement: `/tmp/t690` isolated fixtures (git repo + `github.com` remote; statusline.yaml per case); gh replaced by a logging stub first on PATH (counts `gh api rate_limit …` and `gh repo view …` invocations); binaries named `moai` in per-binary directories because develop's `isSelfInvocable` guard only spawns the detached refresh child for basename `moai` (a guard my first measurement run tripped over — first run showed 0 everywhere including controls, i.e. a vacuous pass, discarded); stdin carries the Claude Code session JSON (`cwd`/`workspace.current_dir` = the fixture) because the github segment and its refresh ride `if input != nil` — with empty stdin the whole path is skipped (second vacuous pass, also discarded). Detached-child calls land in the log; 2s settle per case.

| Case | v3.1.2 (release artifact, `4b2f203fe`, built 2026-08-21) | develop (this branch tree) |
|---|---|---|
| opt-out yaml (`forge: none` + `segments.github: false`), fallback-profile env | **0** | **0** |
| opt-out yaml, no profile env | **0** | **0** |
| segments-only yaml (`github: false`, no `forge:`) | **2** ← polling survives | (not run — develop gates the spawn per REQ-001, source-verified `builder.go:269-271`) |
| no yaml (positive control) | **2** | **2** |

Logged stub invocations (identical shape in every non-zero case): `gh api rate_limit --jq .resources.graphql.remaining` + `gh repo view --json issues,pullRequests --jq …`.

**`__no_such_profile__` (task 4)** — on-disk evidence on this machine:

- `~/.moai/claude-profiles/__no_such_profile__/` exists, mtime 2026-08-28 09:01 — runtime state (`.claude.json`, `agent-memory/`, `backups/`) accumulated 2026-08-27~28, exactly the issue's observation window, and NOT since.
- `launch.yaml` (current): NO `projects:` entry maps anything to `__no_such_profile__`. Two sibling directories carry LLM-response text as profile names — `'Reply with the single word: ok'` (also the value of the global `last_profile` key) and `'Reply only GATEWAY_OK.'` — the signature of the probe-era flow that fed model/transport-probe response text into the profile ledger and let `EnsureDir` materialize it.
- Current code cannot select it: `projects[]` has no entry naming it; the global `last_profile` is no longer read in resolution (`internal/profile/profile.go:88-89` — "write-only on this binary"); and the stale-record guard drops recorded names whose directory is absent. A bare launch resolves to `default` or the project's mapped profile.

## Baseline-attribution

All runs executed today (2026-09-13) in `/tmp/t690`, against the RELEASED v3.1.2 artifact (`gh release download v3.1.2`, checksum asset published; version banner `3.1.2 4b2f203fe`) and against a build of this branch (`bf27af2cb` + t689 docs commit — no code delta vs develop). Source citations (`builder.go`, `github.go`, `renderer.go`, `profile.go`) read at this tree. The first measurement round (all-zero, controls failing) was DISCARDED as vacuous, not reported as a pass.

## Gaps (explicitly NOT observed)

- The ORIGINAL 2026-08-27 incident environment is gone: why the operator's session showed polling and `0/0` under a yaml that (per this measurement) v3.1.2 honours for polling via `forge: none` cannot be reconstructed from here. Candidate explanations — the yaml landed in the project AFTER the incident, the statusline's cwd/projectRoot resolution failed in that session, or a pre-3.1.2 binary was serving the statusline — are not adjudicated.
- `segments-only` was not re-run on develop: the spawn-side segment gate is source-verified (`builder.go:269-271`) and covered by SPEC-STATUSLINE-PROFILE-RESPECT-001's own tests.
- The `__no_such_profile__`-creating flow was not witnessed; the attribution to the probe-era ledger bug rests on the on-disk naming signature, not on a caught-in-the-act execution. If the gateway probe flow ever regenerates profile names from response text, that is a separate defect to file then.

## Residual-risk

- The v3.1.2 half-gate (segments flag does not stop polling) is live in every deployed v3.1.2: an operator who set ONLY `segments.github: false` (without `forge: none`) still gets polled at one spawn per TTL. The fix is already on develop; the residual is release-side (users on ≤3.1.2), not code-side.
- The stale probe-era profile directories (`__no_such_profile__`, `'Reply with the single word: ok'`, `'Reply only GATEWAY_OK.'`) are ~40 MB of residue under `~/.moai/claude-profiles/`; harmless under current resolution, disposable by the operator.

## Issue disposition (for the lead's Korean closing comment)

제보의 핵심 조건(폴백 프로필 + `github: false` + `forge: none`)에서 v3.1.2 배포 바이너리도 gh 호출 0회 — 폴링 부활은 그대로는 재현되지 않으며, develop의 수리(t293/t297)도 동일 조건에서 0회로 유지됩니다. 다만 v3.1.2에서 `segments.github: false`는 그리기만 끄고 폴링은 `forge: none`이 막는 부분 게이트였다는 것이 실측됐고, develop(REQ-001)에서는 세그먼트 스위치가 폴링까지 끊습니다. `__no_such_profile__`은 코드 경로가 아니라 8월 말 프로브 시기의 ledger 잔재이며 현재 코드로는 선택될 수 없습니다.

## Re-measurement scope for the merge window

Docs-only (verdict). Tree identity substitutes for re-measurement; the measurement script is preserved at `/tmp/t690/run2.sh` (machine-local).
