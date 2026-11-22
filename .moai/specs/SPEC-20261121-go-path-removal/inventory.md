# Inventory: what can still reach the Go build (2026-11-21 refresh)

Read 2026-11-21 (UTC clock read with `date -u` at 17:45–19:4xZ). Everything here comes from commands and file reads run for this file; nothing was changed. Baseline: the 2026-11-19/20 inventory in `.moai/specs/SPEC-20261119-rust-only-moai/inventory.md` (branch `spec/20261119-rust-only-moai`, head `903061f05233a42fa8c014d3e0819b70871fe19b`) — rows carried from it say "baseline" and its precise per-hit searches stand unless a re-check below says otherwise. Tip sources read at `6e1bedb4f3a47418f01c887175e7d5fdf8ec4f27`. Probe tool: the installed Rust binary `4.2647.15`, every routed word with `--help` in a temporary HOME; the probe and caller-count scripts live in the worktree's ignored `target/c160/` and are not committed.

Dispositions: `rust-has` (the Rust build already does the job — the Go path can go), `port` (a caller exists and Rust lacks it, or the port exists but is not reachable yet), `retire` (no caller, or a recorded user/leader decision retires it).

## Machine state, 2026-11-21 (changed since the baseline)

- `moai` resolves straight to the Rust build: `~/.local/bin/moai` → `~/.moai/bin/moai` → `~/.moai/versions/4.2647.15/bin/moai`. `moai --version` prints `moai-cli 4.2647.15 (rust)`. A login shell's `which -a moai` reports exactly those two PATH entries and nothing else — the shim folder `~/.moai-cli/path` is not on PATH any more.
- `~/.moai-cli/path/moai` is still the v2 shim script (sha256 `36dd242b…26e4`, equal to `V2_SHA256` in `src/update/switch.rs:27`) but is an off-PATH orphan now, with four saved copies beside it: `moai.before-20261119T210226Z`, `moai.before-4.2647.5-20261117T101820Z`, `moai.before-3021-20261115T093310Z`, `moai.before-rust-update-20261114`. The v2 update step classifies it `V2` and offers no action.
- `~/go/bin` holds no Go moai binary: there is no Go moai binary left there (`moai` and `moai.old` absent; the user moved them to the Trash). Other Go tools remain (`actionlint`, `apidiff`, `buf`, `deadcode`, …), so the folder and its PATH lines cannot be dropped wholesale.
- `~/.moai-cli/path/moai-cli` is a symlink into `~/.moai/versions/4.2647.15/bin/moai-cli`; the bare name `moai-cli` is not on PATH in a login shell, but the statusline wrapper finds the symlink.
- Shell rc state: `~/.zshenv:30-33` carries a "Go binaries (moai, etc.)" comment plus a `~/go/bin` PATH line, `~/.zshrc:153` exports `~/go/bin` again, `~/.zshrc:163-164` still cites `internal/cli/glm.go` in a comment about the GLM key, and both files gained `~/.moai/bin` front-of-PATH lines on the user's 2026-11-18/21 edits. No rc line names `~/.moai-cli/path` any more. `.bash_profile` has only the cargo line.
- The Go toolchain itself is present (`/opt/homebrew/bin/go` 1.27.3, `gofmt`) — the hooks' fast subset (gofmt + go vet) still works; only the Go moai build is gone.
- This repository's shared hooks folder switched on 2026-11-21 (the leader, on the user's order): the Go-generation `pre-commit` and `pre-push` and their two `.moai-*.sha256` hash files are retired and moved aside, and two small Rust hooks are installed — `commit-msg` (plain `exec moai trailer-check --message-file "$1"` with a no-moai pass-through; the leader measured exit 0 for a normal message, a message without trailers and a merge subject, exit 1 for a message carrying a card number, about 9 ms) and `pre-commit` (`exec git diff --cached --check`). There is no `pre-push` hook now. The go-gate pre-commit count over `/Users/goos/MoAI` drops from 13 to 12 with this change.

## Words (Go line of the shim, one row per word)

Probe summary (one run, 60 words, temporary HOME, build 15): `bundle` answers with the M-Flow help line and `moai bundle list` answers with the Go-shaped 11-pack summary (exit 0); `hook` is swallowed (exit 0, one stderr line); `init` prints the removal text; every other word prints `error: unrecognized subcommand '<word>'` — sample exit codes recorded for gate, goal, doctor, verify, status, mcp-server, version, worktree, decision (all exit 2). Caller re-count below is a literal-string file count over `skills install docs src tests tools templates` (2026-11-21); the baseline's precise searches stand as cited.

| word | what the shim sent to Go | Rust equivalent | disposition | evidence |
| --- | --- | --- | --- | --- |
| af | card flow: claim, move, evidence, enqueue, release, land, reconcile | `card`, `lane-next`, `gate-merge` (lacks enqueue/land/reconcile) | rust-has | baseline; probe unrecognized; re-count 0 files |
| agent | agent definition lint | none | retire | baseline 0 callers; re-count 5 files, all prose/history per baseline classification |
| ast-edit | ast-grep rewrites in place | none | retire | baseline 0 callers; re-count 0 |
| ast-grep | rule scan | none | retire | baseline 0 callers (1 closed-issue fixture); re-count 0 outside fixtures |
| bundle | asset packs add/remove/list; the word shadowed Rust's M-Flow bundle | `bundle` (M-Flow) plus the user-asset port `src/user_assets.rs` mounted by `src/ops/bundle.rs:344-346` | rust-has | probe `moai bundle list` exit 0 (11-pack Go-shaped summary); `bundle --help` = M-Flow line; SPEC-20261120-user-asset-bundles is at tip |
| chain | worktree origin trail | none | retire | baseline 0 callers; re-count 0 |
| ci-verdict | one CI verdict per head SHA | none (audit verdicts live in `audit-gate`) | retire | baseline 0 callers (corpus fixture); re-count 0 outside fixtures |
| clean | prune `.moai/state/runs` | none for the same paths | retire | baseline 0 callers; re-count 4 files, prose per baseline |
| completion | shell completion script | none | retire | baseline 0 callers; re-count 0 |
| config | only `config mode-migrate` | none (Rust moves Go-era folders with `go-state-plan`) | retire | baseline 0 callers; re-count 0 |
| constitution | zone registry | none | retire | baseline 0 callers; re-count 0 outside fixtures |
| contract | SPEC autonomy contract | none | retire | baseline 0 callers; re-count 0 outside fixtures |
| decision | decision board record/read | `decision-record`, `decision-append`, `decision-due` | rust-has | baseline; probe unrecognized (equivalents are separate words); re-count 0 |
| doctor | health checks (config/hook/permission/sandbox) | none (narrow: `claude-doctor`, `codex-doctor`, `install-inspect`) | retire | baseline 0 callers; probe exit 2; re-count 0 outside fixtures |
| epic | milestone map from SPEC frontmatter | none | retire | baseline 0 callers; re-count 0 |
| factory | factory runtime (runs, leases, relaunch) | `card`, `lane-next`, `claim-*`, `leader-tick`, `message-record` | rust-has | baseline; probe unrecognized; re-count 0 |
| feedback | report scrub/queue | none | retire | baseline 0 callers; re-count 0 |
| fetch | Jev-scored web excerpts | none (GitHub-only `forge-read`) | retire | baseline 0 callers; re-count 0 |
| gate | heavy vet/lint/test gate (16 languages), `MOAI_PRECOMMIT=1` pre-commit path | full port exists on branch `spec/20261120-rust-quality-gate` (head `9f3de82dc7`, audit pending — absent at tip `6e1bedb4f3`); `main.rs` wiring is a separate piece | port | 12 repos' `.git/hooks/pre-commit` still call `MOAI_PRECOMMIT=1 moai gate` (re-count 2026-11-21 after this repository's hooks were retired; was 13); reproduced today: exit 2, `error: unrecognized subcommand 'gate'` — every commit in those repos is blocked until the port lands or the hooks are retired by the confirmed git-hook slice; bypass `SKIP_MOAI_PRECOMMIT=1` |
| github | parse-issue, link-spec | `forge-read`, `card spec` | rust-has | baseline; probe unrecognized; re-count 0 |
| goal | session goal loop (arm/status/stop-goal) | none of the same object (host `/goal`; Rust `goal-show|plan|apply|block` is the clone goal) | retire | baseline 0 callers; probe exit 2; re-count 5 files, prose per baseline |
| graph | edges.jsonl build/query/check | `index-build`, `atlas-graph`, `atlas-map`, `index-query` | rust-has | baseline; probe unrecognized; re-count 0 outside fixtures |
| gtd | alias of `todo` | `card` | rust-has | baseline; probe unrecognized; re-count 0 |
| handoff | resume record save/clear/show | `lane-resume-note`, `lane-resume-check`, `card handoff` | rust-has | baseline; probe unrecognized; re-count 0 |
| harness | v4 harness learning subsystem (21 subcommands) | none | retire | baseline 0 callers (only Go `harness-observe` wrote its log); re-count 0; user decision retires harness-observe with it |
| hook | 45 subcommands: 27 host events + Go-only duties | the swallow (exit 0, probe-confirmed), `hook-dispatch`, `hook-receipt`, `hooks-plan|apply|revert`, `go-hooks-plan|apply|revert` | rust-has | the channel is covered; duties split by decision: security 3-layer = port (piece 4, before any Go guard drops), `pre-push` = ported on the quality-gate branch, harness-observe/review gates/chain-event/stop-goal/spec-status = retire |
| inbox | lessons inbox status/drain | none | retire | baseline 0 callers; re-count 1 docs research note |
| init | create a project | removal message (probe-confirmed) + `moai cc|glm|codex`, `project-detect|plan|apply` | rust-has | baseline; probe prints the removal text |
| integration | release-integration window | none (opt-in in Go, default off) | retire | baseline 0 callers; re-count 0 |
| inventory | read-only session/worktree view | `worktree-report`, `install-inspect`, `lane-status` | rust-has | baseline; probe unrecognized; re-count 0 outside fixtures |
| jev | TypeSafe credential store | none (marked Go-only in `src/home_layout.rs:149`) | retire | baseline 0 callers; re-count 0 |
| jevmeasure | token benchmark | none | retire | baseline 0 callers; re-count 0 |
| loop | per-SPEC feedback loop | none (host `/loop`, Rust `lane-loop-line`) | retire | baseline 0 callers; re-count 0 |
| lsp | LSP readiness doctor | none | retire | baseline 0 callers; re-count 0 outside fixtures |
| mcp | third-party MCP entry editor | none | retire | baseline 0 callers; re-count 4 files, prose per baseline |
| mcp-server | 22-tool stdio JSON-RPC server | none | retire | user decision 2026-11-20: retire, do not port; the registrations are shown in the slice 2 preview and removed by token only; machine count now 13 `.mcp.json` files (see execution paths) |
| memory | Go memory doctor/archive/drain | none | retire | baseline 0 callers; re-count 0 |
| migrate | agency/cg/home-state/profiles/restore-skill | `home-state-*`, `go-state-*`, `migrate-plan|apply|recovery-plan|recover`, `agents-merge-*` | rust-has | baseline; probe unrecognized; re-count 0 |
| migration | run/status/rollback | none | retire | baseline 0 callers; re-count 0 |
| moai | dead entry (Go answered "Unknown command") | none | retire | baseline; probe unrecognized |
| mx | query/scan over tags | `atlas-tag-scan`, `atlas-map`, `search`, `index-build` | rust-has | baseline; probe unrecognized; re-count 0 outside fixtures |
| paths | DB paths view | none | retire | baseline 0 callers; re-count 0 |
| plan | plan render-html | none | retire | baseline 0 callers; re-count 1 docs line |
| pr | pr watch | `wait ci|comment|pr` | rust-has | baseline; probe unrecognized; re-count counts prose and the Rust `wait pr` docs together |
| preference | decay-scan/toggle | none | retire | baseline 0 callers; re-count 0 |
| profile | profile list/current/delete/setup | none | retire | user decision: retire; `~/.moai/claude-profiles` still holds the 4 Go profile dirs (re-check 2026-11-21), untouched |
| research | research status/list | none | retire | baseline 0 callers; re-count 0 |
| slot | named resource lease | `heavy-slot-acquire`, `heavy-slot-status` | rust-has | baseline; probe unrecognized; re-count 0 |
| state | phase dump/show-blocker | none | retire | baseline 0 callers; re-count 0 |
| status | project state overview | none (`ledger-status`, `statusline` are other jobs) | retire | baseline 0 callers; probe exit 2; re-count 18 files, prose per baseline |
| telemetry | report | none | retire | baseline 0 callers; re-count 0 |
| that | dead entry | none | retire | baseline; probe unrecognized; re-count 1 src prose line |
| todo | backlog queue (29 subcommands) | `card` (lacks hold/move/merge/classify/… per baseline) | rust-has | baseline; probe unrecognized; re-count counts prose only |
| tokens | token accounting | none (`cost-report` is receipts) | retire | baseline 0 callers; re-count 0 |
| tool-policy | YAML to permissions block | none | retire | baseline 0 callers; re-count 0 |
| verify | verify record/check/run/audit-plan | none (`run-verify` is a different job) | port | one live caller (baseline: moai-cowork `.claude/workflows/sync-audit-4dim.js:185` `moai verify check --key-current`); leader decision 2026-11-20: the update slice checks exactly which of those verbs the caller uses and ports the small missing part |
| version | `version` | `--version` flag (probe-confirmed first line `moai-cli 4.2647.15 (rust)`) | rust-has | baseline; probe unrecognized as a word; re-count counts prose only |
| web | 127.0.0.1 settings console | none | retire | baseline 0 callers; re-count 0 |
| workflow | lint | none | retire | baseline 0 callers; re-count 0 |
| worktree | 13 subcommands (new/sync/remove/…/restore) | `worktree-plan|apply`, `worktree-gc-plan|apply`, `worktree-move-plan|apply`, `worktree-report`, `git worktree list` | rust-has | baseline; probe unrecognized as a word; re-count counts prose and git forms |

## Actions the shim splits (`spec`, `session`, `bundle`)

| action | what Go offered | Rust equivalent | disposition | evidence |
| --- | --- | --- | --- | --- |
| spec status | set a SPEC status | none (`spec close` sets completed) | retire | baseline 0 callers; probe reaches Rust which lacks it |
| spec drift | drift report | none | retire | baseline 0 callers |
| spec view | acceptance tree | `spec show` (JSON) | rust-has | baseline; probe reaches Rust |
| spec lint | SPEC lint | `spec check` | rust-has | baseline; probe reaches Rust (`spec check --help` exit 0 today) |
| spec audit | era classification | none | retire | baseline callers are comments and Go-repo CI |
| spec archive | archive a SPEC | `archive-plan|apply|restore|search` (top-level words) | rust-has | baseline; a shim sending `spec archive` to Rust would fail on the name — the split row goes with the shim |
| spec ceiling | ceiling evidence | none | retire | baseline 0 callers |
| session register | register session id | `session-register --name --instance` | rust-has | baseline; arguments differ (recorded in baseline) |
| session heartbeat | heartbeat | `session-register` renew | rust-has | baseline |
| session deregister | deregister | `session-register --release --instance` | rust-has | baseline |
| session list | list sessions | none (`session-register --show` is one name) | retire | baseline 0 callers outside Go-repo docs |
| session purge | purge old sessions | none (leases expire) | retire | baseline 0 callers |
| session current | current session id | `session-current` | rust-has | baseline |
| session doctor | session health | none | retire | baseline 0 callers |
| bundle add | install a pack into the host skill folders | `bundle add` (`src/user_assets.rs`, mounted at `src/ops/bundle.rs:344`) | rust-has | port landed at tip (SPEC-20261120-user-asset-bundles); `moai bundle list` probe exit 0 |
| bundle remove | remove a pack (survivor rules) | `bundle remove` (same mount, line 345) | rust-has | same port |
| bundle list | list packs | `bundle list` (same mount, line 346) | rust-has | probe exit 0, Go-shaped summary lines |

## Execution paths

| path | what can run or install Go today | disposition | evidence |
| --- | --- | --- | --- |
| shim script (product source) | `install/moai-switch.sh` is embedded in the binary (`include_bytes!`, `src/update/switch.rs:29`) and is installable by `moai update --confirm-switch <token>`; the routing word list and the Go notes text live there | retire | file reads at tip; removal is slice 3; slice 2 replaces the update part with Go-link removal |
| update switch step | `src/update.rs` wires the switch layer into update (lines 37-38, 1623, 1656, 1702-1710, 1924-2019); `src/update_tux.rs` draws its screens (576-618, 1268-1315, 1592) and still carries the switch-screen handed-over items (pinned build numbers in the screen text) | port | the layer exists and works; slices 2/3 rework it into Go-link removal per the dispatch; the note text "moai hook still goes to the Go build" (`switch.rs:149`) is false on this machine now |
| machine shim orphans | `~/.moai-cli/path/moai` (v2, sha-matched) plus four `moai.before-*` saved copies, all off PATH | retire | user files — the slice 2 preview lists them and removes only with a token, keeping a revert record |
| machine moai resolution | `moai` = `~/.local/bin/moai` → `~/.moai/bin/moai` → `versions/4.2647.15` — no shim, no Go | rust-has | `which -a` login shell; `moai --version` |
| installer | `install/install.sh` and `install/install.ps1` carry no Go references; `install/README.md:47` still documents the Go locations (`GOBIN`, `GOPATH/bin`); `src/install.rs:13,641,1049` mention `~/go/bin/moai`, the Go version template and the `moai.old` names it leaves alone | rust-has | rg over `install/` and `src/install.rs` at tip; the README row and comment wording are slice 3 text |
| launcher | `moai cc|glm|codex|omp|-l|-a` go to Rust; no Rust source spawns `moai.old` or `go` (the only matches, `src/help.rs:2372,2587`, build clap commands, not processes) | rust-has | baseline grep, unchanged at tip; `src/launcher_init.rs` has no PATH-line writer |
| hooks (project hook folders) | 74 projects keep `.claude/hooks/moai/` (re-count 2026-11-21, equals the baseline); their `moai hook <event>` calls land on the Rust swallow (exit 0) — the Go-only duties stopped when the Go binary left PATH | port | swallow probe; replacement entries come from `go-hooks-apply` per project (slice 2, user-confirmed per project); the security 3-layer port card lands first per the dispatch |
| hooks (git pre-commit) | 12 repos' `.git/hooks/pre-commit` run `MOAI_PRECOMMIT=1 moai gate` guarded only by `command -v moai` — the guard now passes for the Rust binary and the call exits 2 (this repository's hooks were retired on 2026-11-21 and replaced by the small Rust hooks) | port | re-count 2026-11-21 (12, was 13); reproduced: exit=2, `error: unrecognized subcommand 'gate'`; the fast gofmt/vet subset still works (Go toolchain present, staged-`.go` guarded); the user-approved git-hook slice (card notes 16-18) retires and replaces them through `moai update`, recognising the hooks by the hash files and header text and never touching a hook the user wrote |
| hooks (git pre-push) | the Go-era repos' pre-push calls `moai hook pre-push` — swallowed exit 0, so the message-convention check silently passes today; this repository has no pre-push hook now (retired 2026-11-21) | port | hook text read at the Go-era repos' `.git/hooks/pre-push`; swallow probe; whether the Rust generation runs a pre-push check (proposed default: `trailer-check` on the pushed range) is decided at the git-hook slice's revision |
| statusline | `tools/claude-statusline.sh` still falls back to `$HOME/go/bin/moai.old statusline` (lines 25-29) — dead on this machine, still a Go path by text; the project generator `src/status_line.rs` is Rust-only | retire | file reads at tip; delete the fallback in slice 3 |
| commit-msg | `install/commit-msg` execs `${MOAI_BIN:-$HOME/.moai-cli/path/moai-cli}` `trailer-check --message-file` — it calls Rust, not Go, but its default names the old install home; the hook installed in this repository on 2026-11-21 instead calls plain `moai trailer-check --message-file` with a no-moai pass-through | rust-has | file reads at tip and in `.git/hooks/commit-msg`; the in-tree template must drop the hard-coded old launcher path (card note 17) — slice text work |
| bundle-packs (templates/) | `templates/bundle-packs/` (151 files, 25 base names) is embedded by `build.rs` and read by `src/user_assets.rs` — the live catalog of the landed user-asset port | rust-has | rg readers (`src/user_assets.rs`, `tests/bundle_user_assets_cli.rs`, `tools/repo_scans.json`); probe |
| MCP registrations | 13 project `.mcp.json` files name `moai mcp-server` — the baseline counted 10; the re-count adds alfred, snake-game-react and moai-adk-codex, and finds the archived copy renamed `moai-t648-archive.ZsMbQx` | retire | `find /Users/goos/MoAI -maxdepth 2 -name .mcp.json` + grep, 2026-11-21; the slice 2 preview lists them; removal by token only (settings files are guarded) |
| tests | `tests/diff_go.rs` looks up `MOAI_GO_BIN` or `$HOME/go/bin/moai.old` as a path (read-only) and skips with a printed reason when absent; the switch layer is exercised by `tests/switch_script_v2.rs`, `tests/update_switch_cli.rs`, `tests/update_switch_layer.rs`; `src/shim_check.rs` keeps the V1 script text for the `install-inspect` mcp-server routing check | retire | file reads at tip; `diff_go` and the switch-layer tests go with slices 2/3; the shim check stays useful until the shims are gone |
| Go migration tools | `src/go_state.rs`, `src/go_settings*.rs`, `src/go_hooks.rs`, `src/ops/go_*` (6 files) and `src/home_layout.rs` classify and migrate Go-era state | rust-has | dispatch rule: they stay while the update migration needs them; slice 2 builds its Go-link removal on them |
| Go-era docs and comments | 9 `docs/` files cite moai-adk-go (history/research notes), src comments cite Go parity (`src/statusline/godecode.rs:7`, `src/statusline/segments.rs:1`, `src/install.rs:641`, `src/main.rs:2643-2645`), and the shell rc carries `~/go/bin` lines with a "Go binaries (moai, etc.)" comment and a `glm.go` citation | retire | rg re-count and rc reads 2026-11-21; slice 3 wording; the `~/go/bin` PATH lines keep non-moai tools (buf, actionlint) — never drop them blindly |

## Findings

- F1 (live breakage, being fixed): with the Go build gone from PATH, `moai` resolves to Rust and the remaining 12 repos' pre-commit gate call exits 2 — `git commit` in those repos fails today. Reproduced with the exact call shape. This repository's hooks were retired on 2026-11-21 by the leader, and the user approved the fix for every repository: the git-hook slice (right after this inventory) has `moai update` retire the Go hooks (recognised by the hash files and header text) and install the Rust hooks in their place, with preview, apply, revert record and never touching a user-written hook. Until it runs per repo, `SKIP_MOAI_PRECOMMIT=1 git commit` bypasses.
- F2 (silent): the Go-era repos' pre-push convention check passes swallowed (exit 0) — no failure, but the check has been off since the Go binary left PATH. This repository has no pre-push now. Whether the Rust generation runs a pre-push check (proposed default: `trailer-check` on the pushed range, every hook under about 100 ms) is decided at the git-hook slice's revision of this SPEC.
- F3 (order rule): the security 3-layer hooks (security-scan/turn/commit, present in 60 projects per the baseline) are silenced project-by-project once `go-hooks-apply` replaces them — the port card must land first (dispatch rule; piece 4 comparison table is note 5 on the card).
- F4 (count drift): MCP registrations grew from 10 to 13 since the baseline scan.
- F5 (orphans): the v2 shim and four saved copies sit off PATH in `~/.moai-cli/path/` — user-visible files; slice 2 removes them with preview, token and a revert record.
- F6 (stale text): the update preview note "after the switch `moai hook` still goes to the Go build" (`src/update/switch.rs:149`) and the v1→v2 framing are false on a machine with no Go — slices 2/3 replace them.
- F7 (machine text): the rc files carry Go-era lines and comments; the `~/.moai/bin` front-of-PATH lines are already in place (user edit 2026-11-18/21), so `moai` needs no further PATH work — only the Go-era lines are preview material, and the `~/go/bin` lines must stay for the other Go tools.
- F8 (stale default): `install/commit-msg` defaults `MOAI_BIN` to `~/.moai-cli/path/moai-cli`, which still resolves through the versioned symlink but names the pre-move install home; the user's order (card note 17) is that it calls plain `moai`, and the git-hook slice's install command carries the fixed template.

## Not checked

Other machines; the Trash contents (the user moved the Go binaries there — their deletion is the user's step); Orca's own app bundle; `settings.local.json` files; a fresh run of the moai-cowork sync-audit caller (the baseline citation stands); Go subcommand behaviour (no Go moai binary on this machine to run).
