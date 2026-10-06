auditor-model: glm-5.3-flash[1m]

# SPEC Review Report: SPEC-HARNESS-DETACHED-PRUNE-001 (card t1497) — DELTA AUDIT ROUND 1
Verdict: PASS
Overall Score: 0.94
Score Note: iter2 base preserved; the deltas strengthen Testability and Clarity, no deduction.
Plan Artifact Hash: 6298a4c2fce41cb2ea45863b6f9400895e99fe85eae20e082a8dbeaef94b85ff
Auditor Version: plan-auditor/v1 (GLM lane)

verdict: PASS
audited_sha: 8a430d101e40796490cfd11fb1fc0d24617f7397
overall_score: 0.94
must_pass_failed: 0
blocking_count: 0
plan_artifact_hash: 6298a4c2fce41cb2ea45863b6f9400895e99fe85eae20e082a8dbeaef94b85ff
scope: delta
fix_scope: none (PASS)
defect_class: none (no blocking findings)
reread_hunks: plan.md#file-3-seam-signature, plan.md#file-7-wrapper-var, plan.md#file-8-cli-test-fixture, plan.md#M2-exit, plan.md#M3-wrapper-spawn-source, acceptance.md#LEDGER-DP-GREEN-B, acceptance.md#AC-DP-003-row

Header facts: auditor model glm-5.3-flash[1m] (GLM lane); tree SHA `8a430d101e40796490cfd11fb1fc0d24617f7397` (HEAD re-read this run, branch `WT-harness-prune-detached`); delta scope = BOTH post-iter2 amendment commits: `eccd62531` (2 files +41/−4) and `8a430d101` (plan.md wrapper-layer amendment; `spec.md`/`acceptance.md`/`progress.md` untouched by the second commit — `git show --stat` verified); date 2026-10-05. Base attribution: iter2 PASS 0.94 at `2d29b0509` (plan-audit-iter2.md, hash `cbc931b3…ea8846`) rides unchanged for everything outside the amended hunks. Plan-artifact hash pinned at the final tree with the `internal/runtime/audit_cache.go` `ComputeHash` algorithm replicated (Tier M set, 3 files).

Delta provenance: the two leader-side fail receipts (`rcpt-b1711dc2551a0500a04de518` 13:45, `rcpt-25a06b9e9ee7c784137d0615` 14:02) are explained — the rc.27 codex reviews found real P2s (round 1: cross-package seam unreachability + exit-0-only registration blindness; round 2: handler-driven CLI tests not isolatable by a gate parameter alone), and both rounds are repaired in this delta's two commits.

## Delta Scope

1. `eccd62531` — (a) plan.md file 3: gate redesigned to pure parameter injection `MaybeSpawnRetentionPruner(logPath string, spawn func(executable string, args []string) error) error`; (b) plan.md file 8 + acceptance.md LEDGER-DP-GREEN-B + AC-DP-003 cell: registration half re-keyed on the verb's OWN help LISTING CONTENT (usage line carrying `retention-prune`; parent help excluded — the verb registers cobra-Hidden), measured red 0/exit 1 on `d05d1d5f0`; effect half = TestDetachedChildPrunes count-first asserting observable prune effects (kept-line shrink + archive member).
2. `8a430d101` — plan.md file 3/7/8 + M3: injection point moved to the CLI wrapper layer — package-level `var retentionSpawnImpl = <platform real spawn>` in `internal/cli` (file 7), the wrapper draws its spawn from the var and passes it as the gate parameter; the EXISTING handler-driven CLI tests (`runHarnessObserveStop(cmd, nil)` shape — `TestRunHarnessObserveStop_RecordsBaseline` at `hook_harness_multi_event_test.go:46`, 9 observe-handler invocations in `hook_harness_observe_test.go`) override the var in-package (deferred restore, non-parallel — package-var override discipline); `internal/harness` keeps pure parameter injection (gate imports no caller state).

## Delta Verification (re-run this round)

- **Round-2 finding adjudicated CORRECT**: the existing CLI tests drive the real handlers, whose internal gate call passes the real spawn — a fake at the gate's spawn parameter isolates only direct gate calls, not handler-driven paths. The wrapper-layer var is the only source inside `internal/cli` that can isolate them. The repair is the minimal correct shape: the gate stays pure (harness imports no caller state), the caller package owns its replaceable source, and the override discipline (deferred restore via `t.Cleanup`, non-parallel) is explicitly stated in files 7 and 8 — the standard Go package-var test seam, correctly bounded.
- **REQ-DP-007 from BOTH packages**: harness tests → parameter fakes; cli handler tests → wrapper-var override (fake, deferred restore, non-parallel); production → wrapper var defaulting to the platform real spawn passed as the gate parameter. No test spawns a real detached child; anchors verified this round and unchanged (`hook_harness_multi_event_test.go:46`; 9 invocations, `grep -c` = 9).
- **Registration-half probe AS WRITTEN**: `go run ./cmd/moai hook retention-prune --help | grep -c "retention-prune"` → `0`, grep exit 1 — the measured RED reproduces exactly (run at `eccd62531` this round; the amendment did not touch acceptance.md, so the cell and its red stand).
- **Mutant probe — registration half**: unregistered → red (measured); typo'd registration → red (usage line lacks the full string); exit-code escape hatch closed ("an exit code alone is not the gate" — the dispatcher exits 0 on unknown event names, the codex P2's exact defect); unrelated-change residual (parent-help text planting the string) is low-probability and independently covered by the hiddenness half in `TestHookRetentionPruneVerb`. Advisory DA-2.
- **Effect half**: count-first + observable effects (kept-line shrink + archive member) — exit-code-only mutants cannot satisfy it. Sound.
- **Trace verb (re-run post-amendment)**: `COLLECTED: 8 REQ definitions (acceptance input: read)`, zero UNCOVERED, zero ORPHAN.
- **CN-4 verb (mandatory full re-run, post-amendment)**: `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 0 exit bindings, 9 ordering candidates`; zero CONFLICT lines; all 9 candidates the previously-adjudicated benign classes (the amendment added no ordering-language records). MP-9 re-decided PASS.
- **spec lint --strict (post-amendment)**: `✓ No findings — all SPEC documents are valid`, exit 0, source-built.
- **Plan-artifact hash**: `6298a4c2…b85ff` pinned at `8a430d101` (3 files, ComputeHash replicated). The gate reviewer's interim pin `e0cd9f75…c1e7` was measured on `eccd62531` — superseded by the wrapper-layer amendment, as expected.

## Carry-Over (from iter2 base, named per procedure)

- MP-1/MP-2/MP-3: REQ ids and frontmatter untouched by both amendment commits (spec.md and acceptance.md untouched by `8a430d101`); lint re-run confirms.
- MP-4: N/A carried (single-language Go scope).
- MP-5 D7: carried — no new SPEC references in either amendment; owning SPEC still `completed`.
- MP-6 D8: carried — plan §E untouched; syscall placement unchanged.
- MP-7: carried — neither amendment introduces clarification markers (full diffs read).
- MP-8: RED ledger LEDGER-DP-A..H, GREEN-A, FORM byte-identical across both amendments (they only append/none touch the ledger's A..F entries) → iter2 reproductions carried with the identical-source-state attribution; the new surfaces verified: GREEN-B's red probe re-executed this round (reproduces); GREEN-A's controls re-executed at iter2 and the ledger is unchanged since.
- GOOS=windows baseline: carried (no code change in either amendment; artifacts only).

## Defects Found (delta)

None blocking. Advisory (optional class, M6):
- DA-1 (minor, EXTENDED): stale mechanism wording after two seam redesigns — REQ-DP-007 (spec.md) still says "a function field replaced by a recording fake", and AC-DP-006's cell (acceptance.md) still says "the seam is a function field"; the design is now parameter injection (gate) + wrapper-layer var override (cli). The normative core (never spawn a real child; injectable seam) is unchanged and strengthened across both rounds; a short sweep to "injectable spawn seam (gate parameter; cli wrapper-layer var)" closes it. No behavior fork — not blocking.
- DA-2 (minor): the registration-half single-count probe's theoretical unrelated-change mutant (parent-help text mentioning the verb name); pair the GREEN with an explicit parent-help count=0 assertion. The hiddenness half already asserts it in test form.
- DA-3 (nit): the `LEDGER-ACR-J` disposition label referenced in GREEN-B's form note is unlocatable in this tree (grep across `.moai/docs/` and `.claude/rules/`: no match). The substance is independent of the label — the four elements are present and the probe was re-executed this round.

## Receipt

The leader's rc.27 mint landed at 14:02:18Z — `rcpt-25a06b9e9ee7c784137d0615` (tool `codex_audit`, tree_root this worktree, post-marker) — but its `codex_verdict` is **fail** (no `gate_unmet`: per the tree's receipt code, `applyGateUnmet` runs before `recordAuditReceipt`, an empty `gate_unmet` on a fail means a reviewed fail). This is the round-2 review whose findings produced the wrapper-layer amendment — its findings are thereby adjudicated (adopted, repaired in `8a430d101`), but its receipt records fail and cannot corroborate a PASS. Two further mints by this auditor through the installed binary's stdio MCP surface (`rcpt-b58c078c10b08c840293b603` 14:10Z, `rcpt-d59eddb51a4f46ecf1a5c71c` 14:16Z) are **gate-unmet fails** (`gate_unmet` set: "codex stream closed before the turn completed", twice consecutively) — the codex backend could not complete an adversarial turn from this lane, most plausibly contention with the leader's concurrent codex usage. No verdict was delivered by either. None of the three fail receipts is cited. Coverage residual, stated per the agreed protocol: the receipt review snapshots predate or miss the wrapper-layer amendment (`8a430d101`); the amendment itself was re-read and judged directly by this auditor (full diff read + the verification batch above). The standing handoff: the leader's rc.27 server demonstrably completes codex turns — one `codex_audit` there via the adversarial surface (which returns pass on this tree) mints the citable PASS receipt.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: both post-iter2 amendments are sound — the two-round seam evolution (package-private var → parameter → parameter + wrapper-layer var) is the correct minimal path to a test design that spawns no real children from any package, and the re-keyed registration gate is mutation-sensitive where the exit-0 form was blind.
- **Evidence**: all commands and outputs cited inline (probe 0/exit 1; anchors verified; verbs re-run post-amendment; lint 0 findings; hash pinned).
- **Baseline-attribution**: everything measured this run, this tree (`8a430d101`), source-built `go run ./cmd/moai`; carried items carry their iter2 attribution.
- **Gaps**: the effect-half GREEN commands (TestDetachedChildPrunes) are form-verified only — no implementation exists to execute them; Windows vet/lint/coverage remain run-phase M4 obligations; the receipt corroboration is pending a pass-verdict mint (both post-marker mints to date are fails — one reviewed-and-adopted, one gate-unmet).
- **Residual-risk**: DA-2's unrelated-change mutant (low, test-half-covered); DA-3's unlocatable label (cosmetic); the package-var override discipline depends on implementer adherence to the stated non-parallel constraint (stated in files 7/8, testable at review).
