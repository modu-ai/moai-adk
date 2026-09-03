# t462 — Named e2e journey gap inventory (prong B, spec REQ-CEM-007 / AC-CEM-007)

Tree SHA: `bd7c58201` (branch `WT-codex-e2e`). Re-established at run phase: 2026-09-03.
Adjacent cards: t451 LANDED (5 commits in `e9c6a8564..bd7c58201`), t452 NOT landed.

## 0. Method control (AC-CEM-008 — precedes every 0 below)

```
$ grep -c doctor e2e/cli/tux3_journeys.sh   → 7    (same file, same grep: sees present tokens)
```

The e2e surface is exactly one file (`find e2e -type f` → `e2e/cli/tux3_journeys.sh`, 280 lines),
journeys J1–J6: init (J1), update change-preview (J2), doctor (J3), status+spec view (J4),
--help banner (J5), regression matrix+GOOS build (J6). The script builds a fresh binary and
sandboxes all mutation under `/tmp/moai-e2e` — the harness a codex journey would slot into exists.

## 1. The quotable zero

```
$ grep -ric codex e2e/cli/tux3_journeys.sh   → 0
$ grep -ric codex e2e/                        → 0
$ grep -c 'codex\|\.codex' e2e/cli/tux3_journeys.sh → 0
```

No e2e journey anywhere in the repo references codex in any form.

## 2. Named gaps (G1–G8, each re-established at `bd7c58201`)

- **G1 — No e2e journey covers `moai codex app`** (the Codex CLI launch verb).
  Surface exists: `internal/cli/codex_launcher.go:293` `Use: "codex [cli | status | app]"`.
  Establishing: `grep -ric codex e2e/` → 0 (§1 above). A journey would build the binary, run
  `moai codex app` against a scratch `CODEX_HOME` + scratch project, and assert the launcher
  readout/exit behavior.

- **G2 — No e2e journey covers `moai codex status`** — the six-row readiness readout (codex
  binary, CODEX_HOME resolution, auth provider, project wiring, generated agent TOMLs, …) is
  never rendered on the real CLI surface in e2e. Same establishing zero as G1.

- **G3 — No e2e journey covers `moai codex cli` verb passthrough.** Same establishing zero.

- **G4 — No e2e journey covers `moai hook codex-review-gate`** ALLOW/BLOCK Stop-hook behavior.
  Surface exists: `internal/cli/hook.go:221` `Use: "codex-review-gate"`.
  Establishing: `grep -ric 'codex-review-gate' e2e/` → 0. A journey would wire a scratch project
  + `CODEX_HOME` override and assert both the ALLOW and BLOCK paths (this card's Go-test
  execution-log covers the unit surface, not the e2e journey).

- **G5 — No e2e journey asserts codex-side deployment effects of `moai init`/`moai update`**
  (`.codex/hooks.json` + `.codex/config.toml` being written). J1/J2 exist but assert nothing
  codex-side: `grep -c 'codex\|\.codex' e2e/cli/tux3_journeys.sh` → 0 — the string `.codex`
  appears nowhere in the assertions.

- **G6 — No e2e journey covers the MCP codex delegation surface** — `codex_setup`
  (`internal/cli/mcp_server.go:270`), `codex_task` (:294), `codex_job_status` (:305),
  `codex_job_result`, `codex_job_cancel`. Establishing: `grep -ric 'codex_setup\|codex_task\|codex_job' e2e/` → 0.

- **G7 — Live protocol probes (JSON-RPC review liveness) exist only as opt-in Go tests.**
  Gate confirmed at run SHA: `internal/cli/codex_live_protocol_probe_test.go:31`
  `const probeLiveEnv = "MOAI_CODEX_LIVE_PROBE"` — "It never runs on a plain [suite run]".
  No e2e journey exercises the codex JSON-RPC protocol (would require live-quota approval —
  same kickoff gate as M1-D2).

- **G8 — Machine-state calibration for any future journey** (not a journey gap per se): with
  zero moai skills deployed (`ls ~/.codex/skills/` → `.system`, `hatch-pet`; measured at run
  phase, `inventory-run.md` §5), a future journey asserting deployed skills would be red today
  against the real home. Journey authoring (follow-up card) must fixture its own `CODEX_HOME`
  (the seam is `resolveCodexHomeDir()` env-first resolution, `internal/cli/mcp_codex.go:1758`).

## 3. Summary for the lead

8 named gaps; 0 filled by this card (REQ-CEM-008 — this card measures, it does not author
journeys). The existing journey harness (`tux3_journeys.sh` sandbox/binary/log scaffolding) is
reusable for all of G1–G6; G7 additionally needs a live-quota decision. Priority signal from the
(A) execution results (see `execution-log.md`): the unit surface is broad (128 union files) and
green in this run — the gaps are journey-level, not coverage-level, except G7 where even the Go
surface is opt-in-only.
