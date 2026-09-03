# t462 — Codex-axis e2e measurement: plan-phase inventory baseline

Tree SHA: `e9c6a8564` (branch `WT-codex-e2e`, worktree `.claude/worktrees/t462`)
Measured: 2026-09-03, plan phase. SPEC: `.moai/specs/SPEC-CODEX-E2E-MEASURE-001/`

## 1. Three-axis test inventory (per spec.md §A; axis 3 added in audit iter-1 repair)

### Axis 1 — filename glob (raised to repo-wide in iter-2 R5)

Commands (verbatim, as run; the first is the ORIGINAL cli-scoped form whose output is preserved
for lineage, the second is the repo-wide form that is now the axis-1 definition):

```
find internal/cli -name '*codex*_test.go' | wc -l          → 38   (original, cli-scoped)
find internal -name '*codex*_test.go' | wc -l              → 41   (repo-wide: 38 cli + 3 below)
find internal/codexwiring -name '*_test.go' | wc -l        → 6
```

The 3 codex-named files outside `internal/cli` (invisible to the original cli-scoped glob):
`internal/web/mcp_codex_surface_test.go`, `internal/web/codex_card_sentinel_test.go`,
`internal/template/codex_agents_deploy_test.go` — execution-covered via the recursive
`./internal/web/...` / `./internal/template/...` patterns (REQ-CEM-013).

38 `internal/cli` files: codex_audit_gate_unmet, codex_auth_ladder, codex_contract_fixture_unix,
codex_contract_fixture_windows, codex_contract_link, codex_contract, codex_findings_parse,
codex_init, codex_job_control, codex_jobs, codex_launch_verb, codex_launcher_cross,
codex_launcher_guards, codex_launcher_readout, codex_launcher, codex_live_protocol_probe,
codex_protocol_liveness, codex_readiness, codex_registration, codex_review_gate_live,
codex_review_gate, codex_review_rpc, codex_review_target_live, codex_review_target,
codex_rpc_error, codex_session, codex_task, codex_verdict_divergence, codex_verdict_fallthrough,
codex_verdict_regression, codex_verdict_scored, doctor_codex, hook_harness_codex,
mcp_codex_audit_pin, mcp_codex_consoletest, mcp_codex, mcp_project_root_codex,
update_codex_wiring (all `_test.go`).

6 `internal/codexwiring` files: configtoml, hooks, inspect, sidecar, statusline, wire
(42 test funcs total).

### Axis 2 — dependency/symbol sweep (additions the glob misses)

Command (verbatim, as run):

```
find internal/codexadapter -name '*_test.go' | wc -l       → 7
grep -rlE 'runCodexReviewGate|CodexAudit|codex_task|codex_setup|codex_job|codex-review-gate|CodexWiring|resolveCodexHomeDir|codexCmd|codexadapter|codexwiring|codexRunner' internal/ --include='*_test.go' \
  | grep -v -E '/[a-z0-9_]*codex[a-z0-9_]*_test\.go$' \
  | grep -v '^internal/codexwiring/\|^internal/codexadapter/'      → 22 files
```

(a) `internal/codexadapter` — 7 files (config, diagnostics, dispatcher_registration, events,
golden, output, stderr): a WHOLE codex package the filename glob never sees. It is the adapter
whose `EventTable`/`ValidateConfig` the codexwiring package derives from
(`internal/codexwiring/codexwiring.go` package doc).

(b) 22 symbol/import-referencing files outside codex-named files and codex packages:

- `internal/cli` (16): audit_pin_live, glm_job_control, glm_task, hook_e2e, hook_pre_push, hook,
  init_agent_flag, init_agent_wizard (imports codexwiring), init_audit, init_audit_wiring,
  mcp_annotation_guard, mcp_build_identity, mcp_convergence, mcp_glm, multi_review_gate,
  multi_review_gate_wiring (all `_test.go`)
- `internal/core/project` (2): initializer_audit, initializer_audit_wiring
- `internal/hook` (1): review_gate_selfgate
- `internal/mcp` (1): catalog
- `internal/template` (1): review_gate_registration
- `internal/web` (1): mcp_console

### Axis 3 — lexicon delta (added in audit iter-1 repair; 50 files, classified)

Command shape (verbatim, as run): for each file matching `grep -ril 'codex' internal/
--include='*_test.go'` that is NOT filename-axis, NOT in `internal/codexwiring|codexadapter`,
and NOT dependency-axis, record `grep -ci codex <file>`.

Classification rule (measured): **≥5 refs = behavioral candidate (27 files)** — included in the
(A) surface and named below; **1–4 refs = incidental mention (23 files)** — config strings,
registration lists, one-line guards; recorded with counts, not individually asserted.

Behavioral (27), by ref count: `internal/cli/mcp_convergence_participant_test.go` (29),
`internal/template/agentemit/agentemit_test.go` (25),
`internal/template/agentemit/golden_test.go` (22, incl. `TestRealSetCodexShape` at
golden_test.go:172), `internal/config/audit_models_test.go` (22),
`internal/cli/review_gate_config_key_test.go` (20), `internal/cli/wizard/mcp_audit_test.go`
(17), `internal/settings/audit_pin_fields_test.go` (16),
`internal/config/token_budget_guard_test.go` (15), `internal/cli/mcp_audit_multi_test.go` (11),
`internal/cli/mcp_session_msg_test.go` (10), `internal/cli/mcp_glm_audit_pin_test.go` (9),
`internal/web/mcp_secret_hygiene_test.go` (8), `internal/sessionmsg/store_test.go` (8),
`internal/sessionmsg/ids_test.go` (8), `internal/spec/ac_count_clause_test.go` (7),
`internal/sessionmsg/edge_test.go` (7), `internal/github/workflow/validator_test.go` (7),
`internal/config/mcp_audit_config_test.go` (7),
`internal/cli/init_agent_wizard_precedence_test.go` (7),
`internal/cli/audit_blind_verdict_test.go` (7), `internal/web/widget_policy_test.go` (6),
`internal/template/llm_panel_test.go` (6), `internal/template/agentemit/agentemit_edge_test.go`
(6), `internal/cli/audit_pin_test.go` (6),
`internal/template/internal_content_leak_test.go` (5), `internal/sessionmsg/agent_test.go` (5),
`internal/cli/mcp_audit_test.go` (5).

Incidental (23), ref counts 1–4: `internal/cli/{version_stamp_registry, glm_registration,
mcp_session_msg_data, review_gate_projectdir, mcp_session_msg_boundary, mcp_project_root,
mcp_console, init_update_notice, doctor_agentemit_embed}_test.go`,
`internal/cli/wizard/{agent_wiring_question, expansion, restructure}_test.go`,
`internal/sessionmsg/{store_maxpending, round4, envelope, stoprule}_test.go`,
`internal/web/{mcp_glmkey_surface, tab_layout, option_desc, mcp_audit_surface,
audit_option_desc}_test.go`, `internal/template/{skill_dir_token_guard}_test.go`,
`internal/template/agentemit/tomldecodertest_test.go`.

### Union

**126 test files** = 47 (filename: 41 codex-named repo-wide + 6 codexwiring) + 7 (codexadapter)
+ 22 (symbol axis) + 50 (lexicon delta). The 3 codex-named files outside `internal/cli`
(`internal/web/mcp_codex_surface_test.go`, `internal/web/codex_card_sentinel_test.go`,
`internal/template/codex_agents_deploy_test.go` — added in iter-2 R5; the original axis-1 glob
was scoped to `internal/cli`) sit in packages the (A) execution reaches via its recursive
patterns, so they are execution-covered by construction.
The (A) execution surface is defined by RECURCIVE package patterns
(`./internal/cli/...`, `./internal/template/...`, plus the named whole packages), so every
subpackage the lexicon axis reaches runs by construction.

## 2. Machine-state ground truth (re-verified; drift vs relayed numbers)

| Claim (relayed) | Measured at 2026-09-03 | Verdict |
|---|---|---|
| 49 moai skills in `~/.codex/config.toml`, all `enabled = false` | `grep -c '^\[skills\.moai' ~/.codex/config.toml` → `0`; zero `[skills.*]` sections of any shape; 18 `[plugins."moai-*@moai-cowork"]` blocks, ALL `enabled = true` | NOT reproduced — numeric detail drifts |
| `~/.codex/skills/` contains only `.system` + `hatch-pet` | `ls ~/.codex/skills/` → `.system`, `hatch-pet` | CONFIRMED |
| moai skill deployment to codex never happened | follows from the above | CONFIRMED |

Load-bearing consequence: the vacuous-green trap is real — any "all wiring checks pass" verdict
is true of an empty subject until a positive control shows the check detects a deliberate break
(spec REQ-CEM-005 / AC-CEM-006).

## 3. Named e2e journey gaps (prong B baseline, measured at `e9c6a8564`)

Establishing command for all of G1–G3, G5: `grep -ric codex e2e/cli/tux3_journeys.sh` → `0`
(the repo's only e2e file is `e2e/cli/tux3_journeys.sh`, journeys J1–J6: init, update
change-preview, doctor, status+spec view, --help banner, regression matrix+GOOS build).
NOTE: run phase must precede the quotable 0 with the grep positive control (AC-CEM-008).

- **G1** — No e2e journey covers `moai codex app` (Codex CLI launch verb,
  `codex_launcher.go:293` `Use: "codex [cli | status | app]"`) against a scratch `CODEX_HOME`.
- **G2** — No e2e journey covers `moai codex status` (six-row readiness readout: codex binary,
  CODEX_HOME, auth provider, project wiring, generated agent TOMLs, …) rendering on the real
  CLI surface.
- **G3** — No e2e journey covers `moai codex cli` verb passthrough behavior.
- **G4** — No e2e journey covers `moai hook codex-review-gate` (`hook.go:221`) ALLOW/BLOCK
  Stop-hook behavior against a scratch project + `CODEX_HOME` override.
- **G5** — No e2e journey asserts codex-side deployment (`moai init`/`moai update` writing
  `.codex/hooks.json` + `.codex/config.toml`): J1/J2 exist but assert nothing codex-side.
- **G6** — No e2e journey covers the MCP codex delegation surface (`codex_setup`, `codex_task`,
  `codex_job_status/result/cancel` — `mcp_server.go:270-306`) via the running binary.
- **G7** — Live protocol probes (JSON-RPC review liveness) exist only as opt-in Go tests
  (`MOAI_CODEX_LIVE_PROBE=1` gate, `codex_live_protocol_probe_test.go:31`); no e2e journey
  exercises the codex JSON-RPC protocol.
- **G8** — Machine-state calibration (not a journey gap per se): with zero moai skills deployed
  (`~/.codex/skills/` = `.system` + `hatch-pet`), any future journey asserting deployed skills
  would be red today — journey authoring (follow-up card) must fixture its own `CODEX_HOME`.

## 4. Live-test gate inventory (corrected in audit iter-1 repair — three gates, verified)

| Gate | Semantics | Files | Default-suite behavior |
|---|---|---|---|
| `MOAI_CODEX_LIVE_PROBE` | OPT-IN (skipped unless `=1`; `MOAI_CODEX_LIVE_BIN` overrides binary) — `codex_live_protocol_probe_test.go:31` | codex_live_protocol_probe_test.go | skipped |
| `MOAI_SKIP_LIVE_CODEX` | OPT-OUT (RUNS unless `=1`) — `codex_review_gate_live_test.go:35`, `codex_review_target_live_test.go:37` | codex_review_gate_live_test.go, codex_review_target_live_test.go | would RUN — codex IS on PATH (`/Users/goos/.local/bin/codex`) — so the default suite sets `MOAI_SKIP_LIVE_CODEX=1` (kickoff item M1-D2) |
| `MOAI_AUDIT_PIN_LIVE` | OPT-IN (skipped unless `=1`) — `audit_pin_live_test.go:38` | audit_pin_live_test.go | skipped |

In all cases a skip is UNOBSERVED, never a pass.

## 5. Isolation seam confirmation

`resolveCodexHomeDir()` (`internal/cli/mcp_codex.go:1758`): reads env `CODEX_HOME`
(`codexHomeEnvVar`, declared `mcp_codex.go:1730`) first; falls back to `os.UserHomeDir()` +
`.codex`. Env axis is the isolation seam for every integration-style check (spec REQ-CEM-004).
