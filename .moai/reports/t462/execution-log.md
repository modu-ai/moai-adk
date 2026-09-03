# t462 — Prong A execution log (plan M4)

Tree SHA: `bd7c58201` (branch `WT-codex-e2e`, worktree `.claude/worktrees/t462`).
Executed 2026-09-03, serial, after M3 positive-control PASS. Adjacent cards: t451 LANDED,
t452 NOT landed. All runs carried `MOAI_SKIP_LIVE_CODEX=1` (M1-D2 default; kickoff relayed no
live-quota approval). Extracts under `logs/`; full verbose logs machine-local under `/tmp/`
(see Residual-risk in the completion report).

## 1. Pre/post isolation + tree snapshots (plan §C, AC-CEM-009)

| Snapshot | Before (17:1x KST) | After (17:45 KST) | Verdict |
|---|---|---|---|
| `git rev-parse --short HEAD` | `bd7c58201` | `bd7c58201` | identical |
| `git status --porcelain` | 0 lines | 0 lines (all t462 outputs untracked-only) | no foreign mutation |
| `shasum ~/.codex/config.toml` | `ad8c8593a5d89937b9786f1b706384a532361120` | `ad8c8593a5d89937b9786f1b706384a532361120` | **real `~/.codex` not mutated** |
| `ls ~/.codex/skills/` | `.system`, `hatch-pet` | `.system`, `hatch-pet` | unchanged |

Every integration-style check (M3 controls) ran with `CODEX_HOME` command-scoped to
`/tmp/t462-codex-home` (throwaway). No check resolved `CODEX_HOME` to `~/.codex`.

Machine-state probes: `command -v codex` → `/Users/goos/.local/bin/codex` (`codex-cli 0.152.1`,
functional — the OPT-OUT gate therefore mattered and was set on every run). Installed
`/Users/goos/go/bin/moai` = v3.2.0-rc.0 — recorded, NOT used for the control (AP-6).

## 2. Execution census — the "N pass, M uncovered (reason each)" shape

| Run | Command (recursive, per REQ-CEM-003/012/013) | Exit | Duration | Swept (`=== RUN`) | Pass | Skip | Fail |
|---|---|---|---|---|---|---|---|
| 1 | `MOAI_SKIP_LIVE_CODEX=1 go test -count=1 -v ./internal/codexwiring/... ./internal/codexadapter/...` | 0 | ~4s | 103 (56 + 47) | 103 | 0 | 0 |
| 2 | `MOAI_SKIP_LIVE_CODEX=1 go test -count=1 -timeout 1800s -v ./internal/cli/...` (STANDALONE) | 0 | 773s | 6660 | 6629 | 31 | 0 |
| 3a | `MOAI_SKIP_LIVE_CODEX=1 go test -count=1 -v ./internal/config/... ./internal/core/project/... ./internal/github/workflow/... ./internal/hook/... ./internal/mcp/... ./internal/sessionmsg/... ./internal/settings/... ./internal/spec/... ./internal/web/...` | 1 | 589s | 5730 | 5709 | 18 | 3 subtests (2 tests) |
| 3b | `MOAI_SKIP_LIVE_CODEX=1 go test -count=1 -v ./internal/template/...` | 1 | 35s | 1119 | 1102 | 15 | 2 |
| — | **Total** | — | — | **13612** | **13543** | **64** | **5** |

Evidence files: `logs/step1-codexwiring-codexadapter.verbose.log` (full, 16K),
`logs/step2-internal-cli.extract.txt`, `logs/step3a-peripheral.extract.txt`,
`logs/step3b-template.extract.txt`, `logs/hook-flake-rerun.verbose.log`.

Package verdicts: step 2 — all 17 `internal/cli/**` packages `ok` (incl. `wizard` — the
recursive pattern held, AP-7 avoided). Step 3a — 24 packages `ok`, `internal/hook` FAIL,
`internal/spec` FAIL. Step 3b — `internal/template` FAIL, `internal/template/agentemit` FAIL.

### Process observation — lane contention (recorded, not repaired)

Step 2's FIRST attempt (17:21 KST) was killed by this lane on purpose: `pgrep` showed FIVE
concurrent `go test ./internal/cli...` suites on the machine (t446, t470, t454, t410, and this
card's). REQ-CEM-010 forbids concurrency with another lane's verification, so this lane killed
its own run (still in compile phase, 0 tests swept, nothing lost), waited for the machine to
drain (`pgrep -f 'go test'` → 0 at 17:58), and re-ran serially. The recorded step-2 numbers are
from the serial re-run. Attribution of the contention itself: lead's concern, not repaired here.

## 3. The 5 failures — each named, attributed, and classified

1. **`internal/hook` — `TestScanWriteContentNoConfigNoTempFile/control:_resolvable_config_creates_exactly_one_temp_file`**
   (pre_tool_scan_config_test.go:147): "expected exactly 1 security-scan temp file during the
   call, got [2 files]". Suite duration for this test: 428s. **Targeted rerun in isolation:
   PASS in 2.61s** (`logs/hook-flake-rerun.verbose.log`, exit 0). Classification:
   load/environment-sensitive flake (the suite ran while the machine's load average was ~28
   decaying from the 5-lane burst), NOT a deterministic red. Not codex-surface.
2. **`internal/spec` — `TestCatalogHashParity`** (catalog_hash_test.go:192):
   `CATALOG_HASH_DRIFT: entry "sync-auditor" stored=f1b4487f… computed=545d03d9…`.
3. **`internal/template` — `TestManifestHashFormat`** (catalog_tier_audit_test.go:451):
   `CATALOG_HASH_UNSTABLE: sync-auditor stored=f1b4487f…, computed=545d03d9…` — same root as #2.
4. **`internal/template/agentemit` — `TestGoldenCommittedArtifactsMatchEmission`**
   (golden_test.go:109): `.codex/agents/moai/sync-auditor.toml: committed artifact differs from
   emission (sha256 mismatch) — regenerate or stop hand-editing`.
5. (subtest counted under #1's parent; total `--- FAIL` lines = 3 in 3a + 2 in 3b.)

**Shared root for #2–#4** (one defect, three surfaces): the catalog hash + C3 emission for
`sync-auditor` are stale relative to the source `.claude/agents/moai/sync-auditor.md`.
Attribution, measured: stored catalog hash is IDENTICAL at plan base `e9c6a8564` and HEAD
(`git show e9c6a8564:internal/template/catalog.yaml` vs HEAD — both `f1b4487f…`); the `.md`
and the agentemit code are UNCHANGED in `e9c6a8564..bd7c58201` (`git log -- <path>` → 0 for
both; catalog.yaml changed 3× — t367/t348/t447 — but not this entry). **The red predates the
plan base: inherited, not introduced by t462** (whose diff touches only `.moai/reports/t462/`
and `.moai/specs/SPEC-CODEX-E2E-MEASURE-001/`). Repair (`make agents-emit` +
`gen-catalog-hashes.go --all`) is OUT OF SCOPE (REQ-CEM-008, §G) — follow-up card material for
the lead. Deterministic (hash equality on unchanged inputs), so it will redden CI on this tree.

## 4. SKIP inventory — live-gated tests (AC-CEM-005; a skip is UNOBSERVED, never a pass)

Codex-axis live-gated skips (10, all in step 2):

| Gate | Test | File:line of skip reason | Reason |
|---|---|---|---|
| `MOAI_SKIP_LIVE_CODEX=1` (opt-out, set by this run per M1-D2) | `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` | codex_review_gate_live_test.go:36 | gate set — live run would spend real codex quota; lead did not approve quota at kickoff |
| `MOAI_SKIP_LIVE_CODEX=1` | review-target live test | codex_review_target_live_test.go:137 | same ("AC-CRT-010 UNOBSERVED") |
| `MOAI_CODEX_LIVE_PROBE` (opt-in, unset) | `TestCodexLive_{ThreadReuseAndTurnInterrupt, SandboxPolicyStickiness, ReviewStartEmitsTurnStarted, ReviewStartBaseBranchIsNotRejected, OmittedSandboxPolicyBaseline, ExplicitReadOnlyApprovalStall}` (6) | codex_live_protocol_probe_test.go:177/330/395/461/508 | "live protocol probe is opt-in (it spends real codex quota)" — no approval, stays unset (REQ-CEM-010) |
| `MOAI_AUDIT_PIN_LIVE` (opt-in, unset) | `TestAuditPinLive_{GLMDifferential, CodexPinConfirmation}` (2) | audit_pin_live_test.go:182/301 | "the live audit-pin gates are opt-in (they spend real z.ai / codex quota)" — no approval, stays unset |

Non-codex environmental skips (54 across steps 2/3a/3b): migration-runner pending (7),
helper-process-only tests, `go.mod not found` integration gates, ast-grep-absent-requiring
tests, Windows volume-letter gate, corpus-scan opt-in (`MOAI_T362_CORPUS_SCAN`), read-only-dir
environment limits, etc. — full per-skip enumeration with reason lines in the `logs/*.extract.txt`
files. None of these is a codex-surface skip.

## 5. Judgment (prong A conclusion)

Of the enumerated 128-file codex-axis union (inventory-run.md §4), every file sat in an executed
recursive package pattern. Swept 13,612 tests: 13,543 pass, 64 skip (10 codex-live-gated, each
reasoned above; 54 environmental), 5 fail (1 contention flake — passes in isolation; 4 one
inherited sync-auditor hash/emission drift predating this card's base, out of scope to repair).
**The codex-axis unit surface is green on this tree except the named inherited drift; the
uncovered codex surface is exactly the live-quota-gated tests (10) and the e2e journey layer
(prong B, `gap-inventory.md` G1–G8).**
