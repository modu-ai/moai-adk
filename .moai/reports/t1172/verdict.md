# t1172 — verdict (SPEC-CODEX-PREAPPROVAL-PROBE-001)

card: t1172
branch: WT-codex-preapproval-probe (base develop 0356e8117)

## Plan

Plan-phase author: manager-spec. Model requests in this phase: 0. Every `codex exec` below used a scratch `CODEX_HOME` with no login file and a model provider pointing at `http://127.0.0.1:9/v1` (unreachable). Scratch fixtures lived under the session scratchpad; raw outputs were copied to `plan-checks/` (sha256 below).

### Claim 1 — SPEC ID passes the canonical regex

Evidence:

```
$ ID="SPEC-CODEX-PREAPPROVAL-PROBE-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
$ ls .moai/specs/ | grep -c PREAPPROVAL
0
```

### Claim 2 — `codex mcp get --json` does not expose approval settings

Evidence (scratch project with `default_tools_approval_mode = "writes"` and `[mcp_servers.moai.tools.codex_role_audit] approval_mode = "approve"`, trust entry in scratch `CODEX_HOME`):

```
$ codex mcp get moai --json
{ "name": "moai", "enabled": true, "disabled_reason": null,
  "transport": { "type": "stdio", "command": "moai", "args": ["mcp-server"], "env": null, "env_vars": [], "cwd": null },
  "enabled_tools": null, "disabled_tools": null, "startup_timeout_sec": null, "tool_timeout_sec": null }
exit 0
```

(whitespace compacted; field set verbatim.) With the field misspelled (`approval_modex`) the same command also exits 0. `codex --strict-config mcp get moai --json` → `Error: \`--strict-config\` is not supported for \`codex mcp\``, exit 1. `codex --strict-config debug prompt-input hi` → `Error: \`--strict-config\` is not supported for \`codex debug\``, exit 1.

### Claim 3 — `codex exec --strict-config` rejects a misspelled field at load, before any session

Evidence (`plan-checks/typo-strict-exec.err`, exit 1, stdout 0 bytes):

```
Error loading config.toml:
<scratch>/plaintrusted/.codex/config.toml:7:1: unknown configuration field `mcp_servers.moai.tools.codex_role_audit.approval_modex`
  |
7 | approval_modex = "approve"
  | ^^^^^^^^^^^^^^
```

Also: `-c stream_max_retries=0` under `--strict-config` → `Error loading config.toml: unknown configuration field \`stream_max_retries\` in -c/--config override` (overrides are validated too).

### Claim 4 — non-git directory: loader accepts the trust entry, `exec` start gate does not

Evidence: `codex debug prompt-input hi` renders `` `sandbox_mode` is `workspace-write` `` in the directory that has a `CODEX_HOME` trust entry (`plan-checks/prompt-input-trusted.txt`) and `` `sandbox_mode` is `read-only` `` in one without (`plan-checks/prompt-input-untrusted.txt`); both exit 0. In the trusted non-git directory, with the correct key, `codex exec --strict-config … --json -` exits 1 with stdout empty and stderr (`plan-checks/ok-strict-exec.err`):

```
Not inside a trusted directory and --skip-git-repo-check was not specified.
```

This reproduces #39 without a model request. Config load precedes the gate (Claim 3 failed at load in the same directory).

### Claim 5 — git repository root passes the start gate with no trust entry; unreachable provider yields zero model output

Evidence (`-C` = this worktree, scratch `CODEX_HOME`, `timeout 90`, exit 124; `plan-checks/repo-exec.out` head):

```
{"type":"thread.started","thread_id":"01a0d5be-ec0f-7383-be6a-14df42f429b5"}
{"type":"turn.started"}
{"type":"error","message":"Reconnecting... waiting for network (Connection failed: error sending request)"}
```

`grep -c '"type":"item'` → `0`; `grep -c thread.started` → `1`. After exit, `ps -Ao pid,ppid,command | grep -F 'scratchpad/tp'` and `| grep -F '127.0.0.1:9'` → no rows.

### Claim 6 — precondition (t1143 F1/F2/F3 landed) holds on this base

```
$ git merge-base --is-ancestor 51d3e5be6 HEAD && git merge-base --is-ancestor a0b78213d HEAD && echo true
true
```

### Claim 7 — carried AC-CAR-010/011 differ from the source only by the declared path mapping (AC-CPP-011)

The AC-CPP-011 judge, run verbatim from `acceptance.md` against the working tree:

```
true
```

Source hashes after mapping: `carry-010.src` `6a02e64e8a356ac15e28380d40c7b399440b23202bd0ff68998b3296cd374332`, `carry-011.src` `3f497521f338213a9039484b34c73f60263d649a73c6474001d0e8102fde98cb`.

### Claim 8 — spec lint is clean

```
$ moai spec lint .moai/specs/SPEC-CODEX-PREAPPROVAL-PROBE-001
✓ No findings — all SPEC documents are valid
```

### Baseline-attribution

All commands above: this run, this worktree at HEAD `0356e8117` (plan artifacts uncommitted at measurement time), codex-cli 0.156.1.

### Evidence file hashes (`plan-checks/`)

```
d509b04c5e7fde6f53a3e25b4ed089b346bf1a4147331dcd1e0550f3e7fa2dcd  codexhome-config.toml
dd701e712037bba81445e9a92376e3d3af69acac5f0187ec904e9ed3c401324d  ok-strict-exec.err
8c763cf23bca0ab80d68da9787beb54fbd4e9de09c912500824d7c6be84d7f0c  prompt-input-trusted.txt
25113e4e14c4c759bd092c8b89dc1b84a99abe3248bc701ce0bcac6bc8021681  prompt-input-untrusted.txt
6d484eb0043b3408f18cb70042812cb7c3d1a5f1fb38e017b6886b4cbd2e1a7d  repo-exec.out
97c5f37f209cef82840992f29653bda1fa3362aa8273b0df20773ef917a679ee  t1143-route.txt   (= primary-checkout .moai/reports/t1143/m1-route/route.txt)
505f00b4020a87eb6c8d2ed307867f9e5a4f3b26bc62534a0e027aad4cc241ac  typo-strict-exec.err
```

### Gaps

- Why the non-git trust entry fails the `exec` gate (path canonicalization vs git-root-only) is not separated. The SPEC sidesteps it by making every LIVE fixture a git repository.
- Whether the moai MCP server launches at session start under an unreachable provider is not measured (Claim 5 ran in an untrusted root, so the project MCP layer may not have loaded). Run milestone M1-a measures it first.
- Whether `approval_mode = "approve"` overrides `approval_policy = "never"` remains NOT MEASURED. That is M1-b.
- The worktree-session guard refused, in this session, any single command that combines `git` with a `$(...)` substitution. The AC judges that contain `git` were rewritten without substitutions; AC-CPP-001 and AC-CPP-011 were run verbatim here and printed `true`. AC-CPP-007 and AC-CPP-010 contain `git` and were not run (their inputs do not exist yet).

### Residual risk

- The start check depends on codex-cli 0.156.1 behaviour (reconnect loop, error text). A version change can alter it; plan §C Pre-flight 3 pins the version.
- The carried AC-CAR-010 positive control (decoy launched ≥1) assumes the decoy still launches with `enabled_tools = ["spec_progress"]`. Not measured; AC-CPP-004 measures it before any LIVE call.

## Plan — amendment after plan-audit iter 1 (SPEC 0.2.0)

Input: `.moai/reports/t1172/plan-audit-iter1.md` (FAIL 0.70, D1–D27). Model requests in this step: 0. No code, template, or LIVE change.

### Claim A — facts the amendment relies on were re-read in this run

- `internal/cli/mcp_server.go:975` is `mcp.TextContent{Type: "text", Text: tool + ": " + err.Error()},` inside `toolErr` (read with `sed -n 975p`).
- `internal/cli/codex_audit_launch.go:165` formats refusals as `"codex audit %s: "+format`; `:171` `working root rejected: %v`; `:175` `destination rejected: %v`; `:350` `%s is outside the report tree`.
- `.moai/reports/t1143/sync-audit.md` line 118 begins `- **F4** [MEDIUM] …` (137 was wrong).
- In this Claude Code shell, `type diff` → `diff is a shell function from …/shell-snapshots/snapshot-zsh-1790122009427-qzf3rt.sh` (the auditor's D5 cause). Judges now use `command diff`.
- This session's worktree guard refused every tried command that mixed `git` with `$(...)`, `{ … }`, a loop, or a variable in an option position (verbatim refusal texts seen: "names git in a form too complex to verify", "is too complex to verify"). Git-bearing judges were rewritten without those forms; AC-CPP-007 no longer uses git (the `ReadOnlyHint` check reads the tool declaration block with `awk`).

### Claim B — every touched judge was run on synthetic true and false inputs

Toolkit (copied to `plan-checks/iter2/`): `gen.py` (sha256 `db9931229e5a78f3be6882c5a1a96087a4d0621eaee862b03771830fa83b1c8b`) builds one evidence tree per case; `mkvar.sh`, `mkaud.sh`, `mkgo.sh`, `mksrc.sh` build the remaining variants; `run.sh <judge> <case>…` runs a judge file from the case root and prints its last stdout line and exit code. The `jNNN.sh` files are byte-identical to the judge blocks in `acceptance.md` (the file was assembled from them by `assemble.py`).

Command: `sh run.sh jNNN.sh <cases>` from the synthetic root. Output (verbatim lines):

```
# AC-CPP-002
good: true exit=0
missing-input: (empty) exit=1
argv-used-differs: (empty) exit=1
live-before-export: (empty) exit=1
# AC-CPP-003
good: true exit=0
argv-extra: (empty) exit=1
home-extra-line: (empty) exit=1
table-auto: (empty) exit=1
diff-color: (empty) exit=1
diff-after-live: (empty) exit=1
missing-input: (empty) exit=1
# AC-CPP-004 (D2 jq binding fixed, D9, D10, D11, D12)
good: true exit=0
startup-nonmodel-error-item: true exit=0
startup-agent-item: (empty) exit=1
startup-err-missing: (empty) exit=1
startup-auth: (empty) exit=1
model-key: (empty) exit=1
startup-late: (empty) exit=1
# AC-CPP-005 (D3)
good: true exit=0
ran-timeout: false exit=1
auditor: false exit=1
# AC-CPP-006
good: true exit=0
adopt-mismatch: (empty) exit=1
decision-bare: (empty) exit=1
refused: true exit=0
# AC-CPP-008
good: true exit=0
gofalse: (empty) exit=1
# AC-CPP-009
good: true exit=0
gofalse: (empty) exit=1
# AC-CPP-012
good: true exit=0
gofalse: (empty) exit=1
# AC-CPP-013
good: true exit=0
refused: true exit=0
decoy-exposes: (empty) exit=1
routed-decoy: (empty) exit=1
route-missing: (empty) exit=2
# AC-CPP-014
good: true exit=0
over-330: (empty) exit=1
stop-then-live: (empty) exit=1
# AC-CPP-007 source-side variants
src-readonly: (empty) exit=1
src-other-table: (empty) exit=1
src-approve: (empty) exit=1
```

**Correction (added in iter 3, plan-audit iter 2 D-N7).** The two `missing-input: (empty) exit=1` lines above (AC-CPP-002 and AC-CPP-003) came from a half-built tree: the iter-2 `gen.py` deleted `prompt.txt` and then crashed with `FileNotFoundError` when hashing it, leaving no ledger or arm diff. The judges were red, but not for the stated reason. The iter-3 generator builds the complete tree first and deletes the file afterwards; the re-run is in the iter-3 section below (still `false`, now for the stated reason).

`auditor` = the plan-auditor's own synthetic evidence (`plan-audit-iter1-probes/ac-cpp-005-synthetic-evidence.json`, treatment `result_text` "tool call timed out after 60s"). Same input through the auditor's copy of the old judge (`plan-audit-iter1-probes/ac-cpp-005.jq`): `old judge (auditor ac-cpp-005.jq): true`. New judge: `false`.

AC-CPP-007 run from the worktree root (`sh …/j007.sh`) with synthetic streams staged in `.moai/reports/t1172/` and then removed: default streams → `true` exit 0; streams without the `…/default` subtest → exit 1 (no output).

AC-CPP-010, run verbatim from the worktree root on the current tree → no output, exit 1 (RED now: the row has no refusal sentence and no template change exists). The sentence extraction on a synthetic row (current row + one appended sentence, `plan-checks/iter2/row-synthetic.txt`) → `synthetic row: true`. Current row `spawn_agent` count: `1`, in a sentence without `requires approval` (D19 check no longer vacuous).

AC-CPP-011 verbatim after reassembly → `true`. AC-CPP-001 verbatim → `true`. `moai spec lint .moai/specs/SPEC-CODEX-PREAPPROVAL-PROBE-001` → `✓ No findings — all SPEC documents are valid`. `grep -rc 'NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-PREAPPROVAL-PROBE-001/` → 0 in all four files.

No-push gate (acceptance §D): `git ls-remote --heads origin WT-codex-preapproval-probe | grep -c . | grep -qx 0 && echo true` → `true`. Same pipeline on `develop` → no output, exit 1.

### Baseline-attribution

This run, worktree `t1172`, HEAD `258a26596` plus the uncommitted amendment; codex-cli 0.156.1; jq 1.8.1.

### Gaps

- Go-test-based judges (AC-CPP-007/008/009/012) were exercised on synthetic `go test -json` streams, not on real tests (the tests do not exist yet).
- AC-CPP-010 has no full-true run: its true path needs the run-phase template edit.
- AC-CPP-007's `not-adopted` source grep was not run on a mutated template tree.
- Whether `enabled_tools = ["spec_progress"]` on the decoy still launches the decoy is still not measured (AC-CPP-004 measures it before LIVE).
- REQ-CPP-004's "terminate only recorded pids" clause has no mechanical AC.

### Residual risk

- Guard behaviour was inferred from refusals in this session; a different guard version could accept or refuse other shapes.
- Sentence splitting on `.` in AC-CPP-010 assumes the new instruction sentence contains no `.`; plan §F M3 states that constraint. (Superseded in iter 3: AC-CPP-010 now requires a fixed sentence.)

## Plan — iter-3 revision (SPEC 0.3.0, after plan-audit iter 2)

Input: `.moai/reports/t1172/plan-audit-iter2.md` (FAIL 0.79, D-N1–D-N10) and its probes `plan-audit-iter2-probes/` (`newmut.py`, `polarity-row.txt`). Model requests: 0. No code, template, or LIVE change.

### Claim A — design changes and where they live

- D-N1 rerun: LIVE rows carry `attempt`; the lead writes an `invalidate` row between attempts; judges read only the last attempt per group; per-attempt counts/windows; absolute cap 14 counts every attempt (spec REQ-CPP-010, plan §D3 "재실행", plan §D, acceptance §A).
- D-N2 car010 ordering: per-group export manifests; after M2 the car010 inputs are re-exported and the startup check re-run before car010 LIVE; AC-CPP-004 compares only the last startup row per fixture, which must start after that fixture's last `exported_ns` (spec REQ-CPP-003, plan §D3, plan §F M4).
- D-N5: REQ-CPP-008 fixes the exact sentence; AC-CPP-010 requires it byte-for-byte, exactly one `requires approval`, and no `skip the audit and fall back` / `no blocker`.
- D-N6: fixture reset only via `git -C <root> clean -ffdx` after asserting non-empty absolute path under `t.TempDir()`, `rev-parse --show-toplevel` == root, not under any worktree, sentinel file matches (plan §D3); ledger rows carry `fixture_root` + `fixture_sentinel` (AC-CPP-014); subtest `clean_refuses_non_fixture_root` (AC-CPP-012).
- D-N3: startup rows must select the dead provider (`model_provider="dead"` in argv), `env_auth_present:false`; no `model` key in either exported config; no `-c model=…` in any argv (AC-CPP-004).
- D-N4: every LIVE row must have `bound_by` in {test, launcher} (AC-CPP-014).
- D-N8: `evidence.json` carries `recorded_pids`/`killed_pids` (subset check in AC-CPP-005); subtest `kills_only_recorded_pids` (AC-CPP-012).
- D-N9: `OPERATOR-CONFIRM d1sub=… d2=… card=t1172 recorded_by=lead at=…` line required by AC-CPP-006 when the outcome is `RAN` (plan §B).
- D-N10: AC-CPP-013 checks car010/car011 manifests against files and last-attempt LIVE row hashes against the manifests; AC-CPP-004 requires startup `fixture` to be one of the four names.

### Claim B — every touched judge on true and false cases

Toolkit (copied to `plan-checks/iter3/`): `gen.py` (sha256 `d6bbc3b79d188bb6e8eb71d61792404cd996dbb145eda87cec76d51af2642bf5`) now builds the complete tree and applies removal mutations afterwards; `mut3.py` (sha256 `12f316a6d82d806d05c8c6b6f082cd69f6582cb3a40daf936cf594a7b5dd8c99`) repeats the operations of plan-audit iter 2's `newmut.py` (`rerun-disc`, `car010-restart`, `startup-c-model`, `startup-no-select`, `boundby-missing`; `proj-model-key` comes from `gen.py` with consistent hashes) and adds the iter-3 pairs; `matrix.sh` runs every judge file on its cases. The `jNNN.sh` files are byte-identical to the judge blocks in `acceptance.md` (assembled from them by `assemble.py`).

Command: `sh matrix.sh` → `matrix.out` (sha256 `e29573b969683c507cfdde08eebad147e34d1cf3f247c962a3d0b5e337f65768`), verbatim:

```
# j002
good: true exit=0
rerun-recorded: true exit=0
rerun-disc: (empty) exit=1
rerun-unrecorded-attempt2: (empty) exit=1
missing-input: (empty) exit=1
argv-used-differs: (empty) exit=1
live-before-export: (empty) exit=1
# j003
good: true exit=0
rerun-recorded: true exit=0
argv-extra: (empty) exit=1
home-extra-line: (empty) exit=1
table-auto: (empty) exit=1
diff-color: (empty) exit=1
diff-after-live: (empty) exit=1
missing-input: (empty) exit=1
# j004
good: true exit=0
car010-restart: true exit=0
car010-reexport: true exit=0
rerun-recorded: true exit=0
startup-nonmodel-error-item: true exit=0
car010-stale: (empty) exit=1
car010-reexport-nocheck: (empty) exit=1
startup-agent-item: (empty) exit=1
startup-err-missing: (empty) exit=1
startup-auth: (empty) exit=1
startup-env-auth: (empty) exit=1
model-key: (empty) exit=1
proj-model-key: (empty) exit=1
startup-c-model: (empty) exit=1
startup-no-select: (empty) exit=1
startup-late: (empty) exit=1
# j005
good: true exit=0
rerun-recorded: true exit=0
ran-timeout: false exit=1
kill-foreign-pid: false exit=1
rerun-unrecorded-attempt2: false exit=1
# j006
good: true exit=0
confirm-missing: (empty) exit=1
adopt-mismatch: (empty) exit=1
decision-bare: (empty) exit=1
# j008
good: true exit=0
gofalse: (empty) exit=1
# j009
good: true exit=0
gofalse: (empty) exit=1
# j012
good: true exit=0
gofalse: (empty) exit=1
# j013
good: true exit=0
rerun-recorded: true exit=0
decoy-exposes: (empty) exit=1
routed-decoy: (empty) exit=1
route-missing: (empty) exit=2
car-live-hash-drift: (empty) exit=1
# j014
good: true exit=0
rerun-recorded: true exit=0
rerun-disc: (empty) exit=1
rerun-unrecorded-attempt2: (empty) exit=1
rerun-third: (empty) exit=1
boundby-missing: (empty) exit=1
root-in-worktree: (empty) exit=1
over-330: (empty) exit=1
stop-then-live: (empty) exit=1
```

Case key for the coordinator's checks: `rerun-disc` is plan-audit iter 2's rerun mutation verbatim (no attempt index, no record) → AC-CPP-002 and AC-CPP-014 `false`; `rerun-recorded` is the same rerun with attempt 2 and a lead `invalidate` row between the attempts → `true` on AC-CPP-002/003/004/005/013/014. `car010-restart` is plan-audit iter 2's reproduction verbatim → AC-CPP-004 `true`; `car010-stale` (last startup row carries a stale hash) and `car010-reexport-nocheck` (re-exported without a new startup check) → `false`.

A first matrix run, before the synthetic AC-CPP-012 stream was regenerated, fed the judge the older three-subtest stream: `good: (empty) exit=1` — the five-subtest judge rejects it.

AC-CPP-010 row semantics (`sh j010row.sh`: the judge text up to `unset`, byte-identical, run on synthetic template trees): `correct: true exit=0`, `current: (empty) exit=1`, `polarity: (empty) exit=1` (plan-audit iter 2's `polarity-row.txt`), `both: (empty) exit=1` (fixed sentence plus the inverted sentence in one row).

Verbatim on the worktree: AC-CPP-001 → `true`; AC-CPP-011 → `true`; AC-CPP-010 (full judge) → no output, exit 1 (RED now: no fixed sentence, no template change); no-push gate → `true`. `moai spec lint .moai/specs/SPEC-CODEX-PREAPPROVAL-PROBE-001` → `✓ No findings — all SPEC documents are valid`. `grep -c '^### AC-' acceptance.md` → `16`.

### Baseline-attribution

This run, worktree `t1172`, HEAD `8d1aae656` plus the uncommitted iter-3 amendment; jq 1.8.1; `/bin/sh`.

### Gaps

- AC-CPP-007/008/009/012 still judged only on synthetic `go test -json` streams (the tests do not exist yet); AC-CPP-007 was not re-run in iter 3 (its judge is unchanged from iter 2).
- AC-CPP-010's full true path needs the run-phase template edit.
- "One user turn per invocation" (REQ-CPP-004) still has no mechanical judge.
- The `fixture_root` worktree check in AC-CPP-014 matches the path segment `/.claude/worktrees/`; the harness-side assertion (plan §D3) compares against `git worktree list`, which the judge cannot see.

### Residual risk

- The attempt/invalidate protocol depends on the lead writing the `invalidate` row; a rerun without it fails AC-CPP-002/014 by design, which costs the evidence set rather than producing a false pass.

## Plan iter-4 gate (SPEC 0.4.0, after plan-audit iter 3)

plan-audit iter-3 FAIL 0.81 → lead conditional PASS-WITH-DEBT (2026-09-25); gate = mechanical check (no iter-4); sync-audit must re-substitute these judges.

Toolkit: `.moai/reports/t1172/plan-checks/iter4/` (judges `j0*.sh` = the blocks assembled verbatim into acceptance.md, `gen.py`/`mut3.py`/`mut4.py`/`build4.sh`/`matrix4.sh` author matrix, `gen3.py` + `gen3-adapted.py` + `mutgate.py` + `gate.sh` gate, `run4.py` auditor-scenario replay, `d12.py` AC-CPP-010 split run). Cases were built under the session scratchpad; judges run from each case root.

### Claim G1 — the gate generator differs from the auditor's only in the per-label export layout (plus `temp_root`)

Evidence: `shasum -a 256` → `47b98bcbb70781e18428aef94fcaceb37cca639cf6070ea98de93c64b96922d0` for both `plan-checks/iter4/gen3.py` and `/Users/goos/MoAI/moai-adk-go/.moai/reports/t1172/plan-audit-iter3-probes/gen3.py` (byte copy). `diff -u gen3.py gen3-adapted.py > gen3-adapted.diff` (122 lines, sha256 `eb98209eb61a09a733ddbdfa8bd66ce6b78c7054274555a66d3bbbe3d1ceb2ff`). The hunks, verbatim in that file, do exactly this:

- `export_inputs(..., labels=None)`: with labels, writes `<prefix>/<label>/argv.txt` and `<prefix>/<label>/prompt.txt` instead of `<prefix>/argv.txt`/`prompt.txt`; common five files unchanged.
- car manifests list `COMMON` + per-label keys (`K10`, `K11`).
- `live(..., label=None)`: row `label` = invocation label; default argv/prompt hashes read from the label's files.
- car010 rows labelled `direct`, `parent`, and the launcher child `parent` (`bound_by: launcher`, excluded from the hash check).
- car011 realistic rows: **argv, `argv_sha`, `prompt_sha` expressions are the auditor's, unchanged**; only `label=role` is added. The per-label files are written from the same `RARGV`/`RTASK` strings, so the file hashes equal the auditor's row hashes (reason check below).
- Disclosed addition outside the layout: `TEMP_ROOT = os.path.dirname(ROOT)` recorded as `temp_root` on startup and live rows — AC-CPP-014 now requires it (D-N13), and without it every case would be red for that reason instead of the layout.

### Claim G2 — gate cases (a)–(d)

Commands: `python3 gen3-adapted.py c-gate-<x> '{"car011_realistic": true}'` for x ∈ a,b,c,d; then `python3 mutgate.py c-gate-b label-crossed` (swap the two car011 final-attempt row labels), `python3 mutgate.py c-gate-c prompt-byte` (flip bit 0 of byte 0 of `car011/inputs/super-advisor/prompt.txt`, rewrite manifest), `python3 mutgate.py c-gate-d argv-byte` (same on `argv.txt`); then `sh run.sh jNNN.sh gate-a gate-b gate-c gate-d` for AC-CPP-013, 002, 004, 014. Verbatim `gate.out` (sha256 `46cc99cbde27302ff325e0669238f6a10a0f7dd39d24a73e6b5f7a055696dd9f`):

```
mutated label-crossed
mutated prompt-byte
mutated argv-byte
== j013
gate-a: true exit=0
gate-b: (empty) exit=1
gate-c: (empty) exit=1
gate-d: (empty) exit=1
== j002
gate-a: true exit=0
gate-b: true exit=0
gate-c: true exit=0
gate-d: true exit=0
== j004
gate-a: true exit=0
gate-b: true exit=0
gate-c: true exit=0
gate-d: true exit=0
== j014
gate-a: true exit=0
gate-b: true exit=0
gate-c: true exit=0
gate-d: true exit=0
```

| Case | Expected AC-CPP-013 | Observed AC-CPP-013 | AC-002 / 004 / 014 (shared tree) |
|---|---|---|---|
| (a) `car011_realistic` | true | `true exit=0` | true / true / true |
| (b) labels crossed | false | `(empty) exit=1` | true / true / true |
| (c) one byte in `super-advisor/prompt.txt` | false | `(empty) exit=1` | true / true / true |
| (d) one byte in `super-advisor/argv.txt` | false | `(empty) exit=1` | true / true / true |

Every result matches the expectation; no stop condition was hit in the recorded run. Reason check (`reason.py`, verbatim `reason.out`) — the false cases fail on the row-vs-own-label comparison, not on the manifest or the label set:

```
c-gate-a manifest==files: True labels==rows: True rows(label,argv_ok,prompt_ok): [('mission-governor', True, True), ('super-advisor', True, True)]
c-gate-b manifest==files: True labels==rows: True rows(label,argv_ok,prompt_ok): [('super-advisor', False, False), ('mission-governor', False, False)]
c-gate-c manifest==files: True labels==rows: True rows(label,argv_ok,prompt_ok): [('mission-governor', True, True), ('super-advisor', True, False)]
c-gate-d manifest==files: True labels==rows: True rows(label,argv_ok,prompt_ok): [('mission-governor', True, True), ('super-advisor', False, True)]
```

Disclosure — a first gate run used a draft AC-CPP-014 whose new clause was `(.fixture_root|startswith(.temp_root+"/"))`; jq indexes `.temp_root` on the string and errors: all four cases `(empty) exit=5`, reproduced as `jq: error … Cannot index string with string "temp_root"`. The clause was rewritten to `(.temp_root as $t | .fixture_root|startswith($t+"/"))` before anything was assembled or committed; the table above is the rerun. The erroring draft was never committed.

Auditor-scenario replay on adapted trees (`run4.py`, verbatim `run4.out`, sha256 `38444349c28d2e6c35690f5bbfb6cfb0e185ddc9c44f24e62a228bc22be4eae8`): every scenario the auditor's `run3.out` reported for AC-002/003/004/013/014 gives the same result, except `N-car011-realistic-adopt` and `N-car011-realistic-nonadopt`, which flip from `(empty) exit=1` to `true exit=0` on AC-CPP-013 — the defect D-N11 names. AC-CPP-005/006 are excluded from this replay because gen3 emits neither `user_turns` nor a pre-LIVE CONFIRM line; they are covered by the author matrix.

### Claim G3 — D-N12..D-N16 negative controls

| D | Judge | Negative control | Observed | Paired true case |
|---|---|---|---|---|
| D-N12 | AC-CPP-010 | the three p3_ac010 negation-adjacent rows | `neg-prefix-ignore -> (empty) exit=1`, `neg-suffix-unless -> (empty) exit=1`, `neg-suffix-fallback-synonym -> (empty) exit=1` | `canonical -> ROWTRUE exit=0`; `current-row -> (empty) exit=1` |
| D-N13 | AC-CPP-014 | car011 `super-advisor` row `fixture_root` = `/Users/goos/MoAI/moai-adk-go` (`root-primary`); = `/Users/goos/.moai/worktrees/moai-adk-go/t1172` (`root-l2`) | `root-primary: (empty) exit=1`, `root-l2: (empty) exit=1` | `good: true exit=0` |
| D-N14 | AC-CPP-004 | live-row argv gains `--model=gpt-5` / `-mgpt-5` / one token `-c model="gpt-5"` | `argv-model-eq: (empty) exit=1`, `argv-m-joined: (empty) exit=1`, `argv-c-model-one: (empty) exit=1` | `good: true exit=0` |
| D-N15 | AC-CPP-006 | CONFIRM `at=1970-01-01T00:50:00Z` (3000 s; first LIVE at 2000 s) (`confirm-late`) | `confirm-late: (empty) exit=1` | `good: true exit=0` (CONFIRM 00:15:00 = 900 s, DECISION 00:40:00 = 2400 s, final disc end 2120 s) |
| D-N16 | AC-CPP-005 | treatment `user_turns` 2 (`user-turns-2`) | `user-turns-2: false exit=1` | `good: true exit=0` |

D-N12 method (`d12.py`, verbatim `d12.out`, sha256 `609a44759ae45d4c189e25f6ab5d7299b3a2537b9979e38b017eef41136b5dda`): the judge's part A (`git show 0356e8117:…AGENTS.md.tmpl` → base row → expected row → `test -s` → sentinel count 1) ran once in the worktree and printed `PARTA-true`; part B (from the current-template `grep -c` up to ` && unset `, byte-identical, `&& echo ROWTRUE` appended — printed at the head of `d12.out`) ran in synthetic trees with the expected file copied in. The split exists because the synthetic trees are not git repositories.

Author matrix (`build4.sh` then `matrix4.sh`, verbatim `matrix.out`, sha256 `9b255a25cb9bb2480f40f86d8bf61e35caf3fd1fa2502a649e4b2ae13bf26dc1`): every iter-3 true case stays `true` (AC-002 `good`/`rerun-recorded`; AC-003 same; AC-004 `good`, `car010-restart`, `car010-reexport`, `rerun-recorded`, `startup-nonmodel-error-item`; AC-005 `good`/`rerun-recorded`; AC-006 `good`; AC-008/009/012 `good`; AC-013 `good`/`rerun-recorded`; AC-014 `good`/`rerun-recorded`) and every iter-3 false case stays false, now on the per-label layout, `temp_root` rows, `user_turns` arms and 1970-based verdict times.

### Claim G4 — the worktree state after the amendment

Verbatim runs in the worktree: AC-CPP-001 → `true`; AC-CPP-011 → `true`; AC-CPP-010 (full judge, assembled text) → no output, `exit=1` (RED before run: template unchanged); no-push gate → `true`. `moai spec lint .moai/specs/SPEC-CODEX-PREAPPROVAL-PROBE-001` → `✓ No findings — all SPEC documents are valid`. `grep -c '^### AC-' acceptance.md` → `16`. The AC-CPP-010 block extracted from acceptance.md is `cmp`-identical to `plan-checks/iter4/j010.sh`.

### Baseline-attribution

This run, worktree `t1172`, HEAD `b3e22b9f4` plus the uncommitted iter-4 amendment; jq-1.8.1; `/bin/sh`; python3.

### Gaps

- AC-CPP-005/006 were not replayed on the auditor's generator (it emits no `user_turns` and no pre-LIVE CONFIRM line); only the author generator exercises them.
- AC-CPP-007/008/009/012 are still judged only on synthetic `go test -json` streams.
- D-N12's part B was run outside git; part A ran once against the real base commit. The full judge's true path needs the run-phase template edit.
- The gate's car010 labels (`direct`, `parent`) use the same argv/prompt per label in `gen3-adapted.py` (the auditor's car010 values were a single shared pair); only car011 carries distinct per-role values.

### Residual risk

- The label set is taken from the manifest; a harness that exports a label it never runs fails AC-CPP-013 (label set ≠ row labels) — intended, but it means the harness must know its invocation list before the export.
- `fromdateiso8601` treats the CONFIRM/DECISION `at=` values as UTC seconds; a lead writing a local time without `Z` fails the line regex rather than mis-ordering.

## Operator Kickoff record — 2026-09-25

The operator confirmed D1-sub (`config-section-boolean`) and D2 (`doctor-warning`) in this session. The operator also chose to revise the SPEC and use Codex CLI 0.157.0 for LIVE. This records the decision; it is not LIVE evidence.

OPERATOR-CONFIRM d1sub=config-section-boolean d2=doctor-warning card=t1172 recorded_by=lead at=2026-09-25T04:25:59Z

## M1-a 비모델 선행 점검 — 2026-09-25

### Claim

Codex CLI 0.157.0의 설정 로더는 현재 writer 출력과 오타·잘못된 enum을 구분했다. 네 시작 검사 픽스처에서 moai MCP가 각각 한 번 기동했고, car010의 decoy도 한 번 기동했다. 이 결과는 모델을 사용한 사전 승인 판별이나 첫 LIVE 허가를 뜻하지 않는다.

### Evidence

실행 명령과 관측 출력:

```text
go test ./internal/codexwiring -run '^TestGeneratedConfigKeysLoadInCodex$' -count=1 -v -timeout=30s
--- PASS: TestGeneratedConfigKeysLoadInCodex (0.09s)
    --- PASS: TestGeneratedConfigKeysLoadInCodex/generated_config (0.05s)
    --- PASS: TestGeneratedConfigKeysLoadInCodex/positive_control_misspelled_key (0.02s)
    --- PASS: TestGeneratedConfigKeysLoadInCodex/positive_control_bad_enum (0.02s)
    --- PASS: TestGeneratedConfigKeysLoadInCodex/diagnostic_matcher_controls (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/codexwiring 0.555s

unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexPreApprovalGateRefusals$' -count=1 -v > .moai/reports/t1172/ac-cpp-012.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexPreApprovalGateRefusals")]|length)==1 and ([.[]|select(.Action=="pass" and (.Test=="TestCodexPreApprovalGateRefusals/missing_input" or .Test=="TestCodexPreApprovalGateRefusals/arm_diff_exceeds" or .Test=="TestCodexPreApprovalGateRefusals/startup_check_fails" or .Test=="TestCodexPreApprovalGateRefusals/clean_refuses_non_fixture_root" or .Test=="TestCodexPreApprovalGateRefusals/kills_only_recorded_pids"))]|length)==5 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0' .moai/reports/t1172/ac-cpp-012.jsonl
true

unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && MOAI_CODEX_PREAPPROVAL_LIVE=1 MOAI_T1172_EVIDENCE_DIR=.moai/reports/t1172 go test ./internal/cli -run '^TestCodexPreApprovalStartup$' -count=1 -v -timeout=150s
    codex_preapproval_startup_test.go:343: disc-control startup exit=-1 moai=1 decoy=0 items=0
    codex_preapproval_startup_test.go:343: disc-treatment startup exit=-1 moai=1 decoy=0 items=0
    codex_preapproval_startup_test.go:343: car010 startup exit=-1 moai=1 decoy=1 items=0
    codex_preapproval_startup_test.go:343: car011 startup exit=-1 moai=1 decoy=0 items=0
--- PASS: TestCodexPreApprovalStartup (87.23s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 87.810s
```

위 AC-CPP-012 판정식의 전체 명령은 `acceptance.md` 해당 항목에 있으며, 로컬 재실행은 `true`를 냈다. 각 시작 검사 출력은 `startup/<fixture>.jsonl`·`.err`·`.launches.json`, 호출 시각과 입력 해시는 `ledger.json`에 있다. 실제 `jq` 재확인에서 네 파일 모두 `thread.started` 1개, 비오류 `item.*` 0개, 표준 오류 39바이트였다. 매니페스트 파일 수는 discriminator 14, car010 9, car011 9이고, 파일 해시 불일치는 각각 0이었다. `arm-diff.txt`에는 `project-config.toml`의 빈 줄 하나와 도구별 표 두 줄만 있었다.

### Baseline-attribution

이 worktree `WT-codex-preapproval-probe`, 점검 직전 HEAD `7ff8b9f52`, Codex CLI 0.157.0, 이번 실행의 `bin/moai` 빌드 결과를 기준으로 측정했다. 시작 검사에서는 로그인 파일과 인증 환경 변수를 넣지 않았고, 제공자 주소를 `http://127.0.0.1:9/v1`로 고정했다. 네 호출은 각각 20초 상한에서 종료됐다(`exit=-1`); 비모델 이벤트 0과 MCP 기동 수는 종료 전 출력에서 확인했다.

### Gaps

- LIVE 모델 호출을 실행하지 않았다. `approval_mode`가 실제로 승인 요구를 없애는지는 미측정이다.
- 현재 `argv.txt`와 car010·car011 라벨 프롬프트는 M1-a 준비용 값이다. 실제 LIVE 호출 인자가 정해지면 재반출하고, 매니페스트를 갱신하고, 마지막 시작 검사를 다시 수행해야 한다. 따라서 AC-CPP-002·004의 LIVE 전제는 아직 충족했다고 주장하지 않는다.
- ignored 원시 산출물은 로컬 `.moai/reports/t1172/`에 있으며 이 증거 커밋에는 포함되지 않는다.

### Residual-risk

가짜 Codex의 중단 경로가 통과한 것과 실제 Codex의 배차·승인 결과는 별개다. 종료 상한에서 MCP가 한 번 기동해도 이후 도구 호출의 승인 정책은 바뀌지 않을 수 있다. 첫 LIVE 전에 실사용 인자 재반출, 마지막 네 시작 검사, 장부 해시·시각 판정을 다시 닫아야 한다.

## M1-a 감사 F1·F2 수리 증거 — 2026-09-25

### Claim

독립 감사 `sync-audit.md`의 F1·F2 재현 입력을 이제 LIVE 콜백 전에 거부한다. F1의 실제 `diff -r` 바이너리 차이 추가와 F2의 빈·임의 시작 검사, MCP 기동 0회, car010 decoy 0회, 설정 적재 오류, 비허용 도구 이벤트를 음성 대조로 확인했다. 이 기록은 독립 재감사의 PASS 판정이 아니다.

### Evidence

```text
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test ./internal/cli -run '^TestCodexPreApprovalGateRefusals$' -count=1 -v -timeout=40s
--- PASS: TestCodexPreApprovalGateRefusals (0.57s)
    --- PASS: TestCodexPreApprovalGateRefusals/missing_input (0.00s)
    --- PASS: TestCodexPreApprovalGateRefusals/arm_diff_exceeds (0.00s)
    --- PASS: TestCodexPreApprovalGateRefusals/startup_check_fails (0.56s)
    --- PASS: TestCodexPreApprovalGateRefusals/clean_refuses_non_fixture_root (0.00s)
    --- PASS: TestCodexPreApprovalGateRefusals/kills_only_recorded_pids (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 1.378s

AC-CPP-012 JSON 판정 명령(acceptance.md) 출력
true

unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && MOAI_CODEX_PREAPPROVAL_LIVE=1 MOAI_T1172_EVIDENCE_DIR=.moai/reports/t1172 go test ./internal/cli -run '^TestCodexPreApprovalStartup$' -count=1 -v -timeout=150s
    codex_preapproval_startup_test.go:342: disc-control startup exit=-1 moai=1 decoy=0 items=0
    codex_preapproval_startup_test.go:342: disc-treatment startup exit=-1 moai=1 decoy=0 items=0
    codex_preapproval_startup_test.go:342: car010 startup exit=-1 moai=1 decoy=1 items=0
    codex_preapproval_startup_test.go:342: car011 startup exit=-1 moai=1 decoy=0 items=0
--- PASS: TestCodexPreApprovalStartup (87.25s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 87.840s
```

F1 대조는 두 팔의 `project-config.toml`에 허용 표를 추가하고 `z.bin`에 다른 바이트를 넣어 실제 `diff -r`로 생성했다. 전체 출력에 `Binary files ... differ`가 포함된 것을 확인한 후 게이트의 `stop` 행과 LIVE 기동 0회를 단언했다. F2 대조는 구조화된 시작 검사 결과의 `thread.started` 정확히 1개와 moai 기동 1회 이상을 양성 조건으로 삼았다. car010에는 decoy 기동도 필요하다.

### Baseline-attribution

이번 수리는 worktree `WT-codex-preapproval-probe`의 독립 감사 커밋 `abb5d158c` 위에서 측정했다. 실제 시작 검사는 설치된 Codex CLI 0.157.0, 인증 없는 `CODEX_HOME`, 접속 불가 제공자 주소 `http://127.0.0.1:9/v1`에서 재실행했다.

### Gaps

- 독립 재감사 결과를 아직 받지 않았다. M1-b LIVE는 실행하지 않았다.
- 실제 LIVE 호출 인자·프롬프트로 다시 반출하고 마지막 시작 검사·장부 해시 판정을 닫는 작업이 남았다.
- 원시 JSONL·장부·매니페스트는 이 worktree의 ignored 로컬 산출물이다.

### Residual-risk

테스트 하네스의 양성 시작 검사와 실제 사전 승인 효과는 서로 다른 판정이다. 독립 재감사와 LIVE 전 입력 일치 검사를 통과하기 전에는 승인 동작을 주장할 수 없다.

## M1-b 비모델 하네스 준비 — 2026-09-25

### Claim

M1-b 판별 하네스의 비모델 선행 경로가 기존 M1-a 장부와 원자료를 보존하면서 고유 시작 검사 원자료와 반출 스냅샷을 만든다. 실모델 판별 호출은 실행하지 않았다.

### Evidence

```
$ go test ./internal/cli -run '^(TestCodexPreApprovalGateRefusals|TestCodexPreApprovalArgvShape|TestCodexPreApprovalLivePreflight|TestCodexPreApprovalM1aBaseline)$' -count=1 -v -timeout=60s
--- PASS: TestCodexPreApprovalLivePreflight (0.00s)
--- PASS: TestCodexPreApprovalM1aBaseline (0.00s)
--- PASS: TestCodexPreApprovalGateRefusals (0.24s)
--- PASS: TestCodexPreApprovalArgvShape (0.03s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 1.094s
```

`TestCodexPreApprovalM1aBaseline`은 M1-a 원본을 복사한 뒤 첫 장부 행 변이와 원자료 변이를 각각 거부했다. `TestCodexPreApprovalLivePreflight`는 재시도 무효 기록·상한·시작 검사 해시와 루트·반출 해시·승인 거부와 서버 실행 문구의 합성 대조를 통과했다.

```
$ MOAI_CODEX_PREAPPROVAL_LIVE=1 MOAI_T1172_EVIDENCE_DIR=/tmp/t1172-export.iBxWjH/.moai/reports/t1172 go test ./internal/cli -run '^TestCodexPreApprovalStartupPreserve$' -count=1 -v -timeout=150s
disc-control startup exit=-1 moai=1 decoy=0 items=0
disc-treatment startup exit=-1 moai=1 decoy=0 items=0
car010 startup exit=-1 moai=1 decoy=1 items=0
car011 startup exit=-1 moai=1 decoy=0 items=0
--- PASS: TestCodexPreApprovalStartupPreserve (87.07s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 87.667s
```

실행 직전 `/tmp` 증거 디렉터리에 기존 `ledger.json`과 M1-a 원자료 12개를 복사했다. 실행 후 장부 8행(원본 4행 + 신규 4행), 신규 `startup_id` 네 개, 고유 시작 검사 원자료 12개, 묶음별 반출 스냅샷 세 개를 확인했다. 기존 원자료 12개의 경로·SHA 출력 해시는 `c8653794d64014487c9f05d3e54f07e4b5fb61cf783f5e18f6460bb0e6a4b937`로 원본과 같았다. 신규 네 행의 stdout/stderr/launches 및 export 해시는 각 원자료와 모두 일치했다(`rows=4 export_snapshots=3 startup_raw=12 hashes=all_match`). `git diff --check`는 출력 없이 exit 0이었다.

### Baseline-attribution

위 검사는 worktree `WT-codex-preapproval-probe`, 설치된 Codex CLI 0.157.0, 접속 불가 제공자 `http://127.0.0.1:9/v1`을 사용한 이 실행의 관측이다. M1-a 장부·원자료 원본은 `.moai/reports/t1172/`에서 격리된 `/tmp/t1172-export.iBxWjH/`로 복사했으며 원본에는 이 실측이 쓰지 않았다.

### Gaps

- `TestCodexPreApprovalDiscriminatorLive`를 실행하지 않았다. 승인 정책 효과, 실제 MCP 도구 결과, LIVE 장부와 증거 해시는 미측정이다.
- 독립 plan audit의 수정 문서 최종 판정과 실제 LIVE 허가가 남았다. 이 절의 PASS는 비모델 하네스 경로에만 해당한다.

### Residual-risk

실제 Codex `--json`의 MCP 결과 구조가 합성 분류 입력과 다를 수 있다. LIVE 도중 실패하면 `stop` 행을 기록하고 같은 증거 디렉터리를 재사용하지 않도록 하네스가 거부한다.

## 2026-09-25 현재 실측·M4 준비 상태

### Claim

Codex CLI 0.157.0의 첫 판별 시도에서 대조 호출은 도구 승인을 거부했고 처치 호출은 `codex_role_audit` 서버까지 도달했다. 이 측정은 `approval_mode = "approve"`의 판별 효과를 보여 주지만, 채택 형태 A1/A2는 아직 운영자가 정하지 않았다. 이월 AC-CAR-011은 실모델 호출 전 준비 코드와 결정적 경계 테스트만 작성했으며 LIVE 결과는 없다. 이 절은 위 M1-a·M1-b 준비 당시의 미측정 문구보다 나중 상태를 기록한다.

### Evidence

```text
$ jq -r '"rows=\(length) startup=\(map(select(.kind=="startup"))|length) live=\(map(select(.kind=="live"))|length) stop=\(map(select(.kind=="stop"))|length)",(.[] | select(.kind=="live") | "\(.fixture) attempt=\(.attempt) exit=\(.exit) auth_equal=\(.auth_sha256_before==.auth_sha256_after)")' .moai/reports/t1172/ledger.json
rows=10 startup=8 live=2 stop=0
disc-control attempt=1 exit=0 auth_equal=true
disc-treatment attempt=1 exit=0 auth_equal=true

$ jq -r '"version=\(.codex_version) attempt=\(.attempt) calls=\(.invocations) aborted=\(.aborted) export=\(.export_complete) diff=\(.arm_diff_ok) auth_equal=\(.auth_sha256_before==.auth_sha256_after)", (.arms | to_entries[] | "\(.key) calls=\(.value.moai_role_audit_calls) outcome=\(.value.outcome) children=\(.value.child_processes) turns=\(.value.user_turns) text=\(.value.error_text // .value.result_text)")' .moai/reports/t1172/discriminator/evidence.json
version=0.157.0 attempt=1 calls=2 aborted=false export=true diff=true auth_equal=true
control calls=1 outcome=REFUSED children=0 turns=1 text=MCP tool call requires approval, but approval policy is never
treatment calls=1 outcome=RAN children=0 turns=1 text=codex_role_audit: codex audit sync-auditor: destination rejected: AGENTS.md is outside the report tree

$ cat .moai/reports/t1172/discriminator/outcome.txt
RAN

$ go test ./internal/cli -run '^TestCodexPreApprovalCar011(PreparationBoundary|ExportIntegrity)$' -count=1 -v
--- PASS: TestCodexPreApprovalCar011PreparationBoundary (0.00s)
--- PASS: TestCodexPreApprovalCar011ExportIntegrity (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 0.816s
```

### Baseline-attribution

장부와 판별 원자료는 이 worktree의 `.moai/reports/t1172/`에서 이번 확인 때 읽었다. M4 결정적 테스트는 `WT-codex-preapproval-probe`의 HEAD `6bff41fd1` 위에 아직 커밋되지 않은 하네스 변경을 대상으로 실행했다. 호출 상한은 SPEC plan §D의 LIVE 14회, 시작 검사 16회다. 현재 장부는 각각 2회와 8회다. 모델에 닿는 신규 호출은 이 M4 작업에서 실행하지 않았다.

### Gaps

- A1/A2 채택 결정 행은 아직 없으며 M2 writer·doctor 변경과 채택 갈래의 AC-CAR-010 실행도 남았다. M3 지시면의 독립 감사 기록은 별도 보고서에 있다.
- AC-CAR-011의 두 역할 LIVE 호출과 그 전용 시작 검사 1회는 아직 실행하지 않았다. 이월 AC의 결과를 PASS로 세지 않는다.
- Windows 런타임과 실제 운영 프로젝트 설정 전파는 이 측정 범위 밖이다.

### Residual-risk

처치의 `RAN`은 안전한 거부 목적지까지 서버 함수가 실행된 관측이다. 실제 감사 자식이 완료되는지와 다른 역할의 read-only 샌드박스가 쓰기를 거부하는지는 각 이월 AC의 LIVE 호출로 따로 측정해야 한다. M4 준비 코드도 독립 감사가 남아 있다.
