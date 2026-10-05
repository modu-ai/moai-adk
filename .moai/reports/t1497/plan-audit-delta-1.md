auditor-model: glm-5.3-flash[1m]

# SPEC Review Report: SPEC-HARNESS-DETACHED-PRUNE-001 (card t1497) — DELTA AUDIT ROUND 1
Verdict: PASS
Overall Score: 0.94 (iter2 base preserved; delta strengthens Testability and Clarity, no deduction)
Plan Artifact Hash: e0cd9f75bc81805ad7459b24182c1a594d36530ed21618a741809063e9f5c1e7
Auditor Version: plan-auditor/v1 (GLM lane)

verdict: PASS
audited_sha: eccd6253176e1b4869e0a27e068d99076ab3f3fc
overall_score: 0.94
must_pass_failed: 0
blocking_count: 0
plan_artifact_hash: e0cd9f75bc81805ad7459b24182c1a594d36530ed21618a741809063e9f5c1e7
scope: delta
fix_scope: none (PASS)
defect_class: none (no blocking findings)
reread_hunks: plan.md#file-3-seam-signature, plan.md#file-8-cli-test-fixture, plan.md#M2-exit, acceptance.md#LEDGER-DP-GREEN-B, acceptance.md#AC-DP-003-row

Header facts: auditor model glm-5.3-flash[1m] (GLM lane); tree SHA `eccd6253176e1b4869e0a27e068d99076ab3f3fc` (HEAD re-read this run, branch `WT-harness-prune-detached`); delta commit `eccd62531` (2 files +41/−4: plan.md, acceptance.md; `progress.md` and `spec.md` untouched — `git show --stat` verified); date 2026-10-05. Base attribution: iter2 PASS 0.94 at `2d29b0509` (plan-audit-iter2.md, hash `cbc931b3…ea8846`) rides unchanged for everything outside the two amended hunks.

Delta provenance note: the 13:45 fail receipt (`rcpt-b1711dc2551a0500a04de518`) is now explained — the leader's rc.27 codex review found two real P2s (cross-package seam unreachability; exit-0-only registration gate blindness), and `eccd62531` is their repair. The iter2 conflict-of-record is closed by this adjudication.

## Delta Scope (leader-approved)

1. plan.md — seam redesigned: `MaybeSpawnRetentionPruner(logPath string, spawn func(executable string, args []string) error) error` — the spawn function is a PARAMETER (cross-package functional injection), replacing the package-private `retentionSpawnFn` var seam.
2. acceptance.md — AC-DP-003's registration half re-keyed on the verb's OWN help LISTING CONTENT (usage line carrying `retention-prune`; parent help excluded — the verb registers cobra-Hidden), with LEDGER-DP-GREEN-B carrying the commands, the measured red (grep count 0 / exit 1 on `d05d1d5f0`), and the effect half as TestDetachedChildPrunes count-first asserting observable prune effects (kept-line shrink + archive member).

## Delta Verification (re-run this round)

- **Registration-half probe AS WRITTEN**: `go run ./cmd/moai hook retention-prune --help | grep -c "retention-prune"` → `0`, grep exit 1 — the measured RED reproduces exactly on this tree (`eccd62531`, source-built). The RED is the unregistered state; the gate can only flip when the verb's own help carries the name.
- **Mutant probe — registration half**: (a) unregistered verb → measured red (above); (b) typo'd registration (`retention-prun`) → the usage line lacks the full string → stays red; (c) the exit-code escape hatch is closed — the cell explicitly demotes exit codes ("an exit code alone is not the gate") because the dispatcher exits 0 on unknown event names (the codex P2's exact defect); (d) unrelated-change residual: a hypothetical future edit planting the string "retention-prune" into parent-hook help text could false-flip the single count — low probability, and the hiddenness half (parent help carries no hidden verb) is independently asserted in `TestHookRetentionPruneVerb` (plan file 8). Advisory DA-2 recommends pairing the GREEN with an explicit parent-help count=0 assertion. No blocking mutant found.
- **Mutant probe — effect half**: count-first (`-list` lists exactly 1) + observable effects (kept-line count shrink + `<YYYY-MM>.jsonl.gz` archive member) — an exit-code-only mutant cannot satisfy it; the assertions bind the prune's actual outputs. Sound.
- **Seam design vs REQ-DP-007, both packages**: the parameter design satisfies the requirement from BOTH packages — `internal/harness` tests inject recording fakes as before, AND `internal/cli` tests inject through the same exported entry, which the old package-private `var` seam structurally could not (unreachable cross-package — the codex P2's exact hazard: handler tests driving unstamped temp logs would have spawned real detached children). Plan file 8 names the concrete fixture anchors, both verified this run: `TestRunHarnessObserveStop_RecordsBaseline` at `hook_harness_multi_event_test.go:46` and the observe-handler family at 9 invocations (`grep -c` = 9 in `hook_harness_observe_test.go`). Production wiring passes the platform real implementation; no test spawns a real child.
- **Trace verb (re-run)**: `COLLECTED: 8 REQ definitions (acceptance input: read)`, zero UNCOVERED, zero ORPHAN.
- **CN-4 verb (mandatory full re-run)**: `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 0 exit bindings, 9 ordering candidates`; zero CONFLICT lines; all 9 candidates hand-read — the benign carried classes (count-first form descriptors, stamp-before-work algorithm prose, "before run-phase exit" aligning with M4-last) plus the two new rows (AC-DP-003 cell, GREEN-B effect half — form descriptors, not orderings). MP-9 re-decided PASS.
- **spec lint --strict**: `✓ No findings — all SPEC documents are valid`, exit 0 (re-run on the amended artifacts, source-built).
- **Plan-artifact hash**: recomputed at `eccd62531` → `e0cd9f75…c1e7` (3 files, ComputeHash algorithm replicated).

## Carry-Over (from iter2 base, named per procedure)

- MP-1/MP-2/MP-3: REQ ids and frontmatter untouched by the diff (the diff's spec.md is untouched entirely); lint re-run confirms.
- MP-4: N/A carried (single-language Go scope).
- MP-5 D7: carried — no new SPEC references in the diff; owning SPEC still `completed`.
- MP-6 D8: carried — plan §E untouched by the diff; syscall placement unchanged.
- MP-7: carried — the diff introduces no clarification markers (full diff read).
- MP-8: RED ledger LEDGER-DP-A..H, GREEN-A, and FORM byte-identical (the diff only appends GREEN-B) → iter2's reproductions carried with the identical-source-state attribution; the NEW surface (GREEN-B's red probe) re-executed this round and reproduces.
- GOOS=windows baseline: carried (no code change in the delta; artifacts only).

## Defects Found (delta)

None blocking. Advisory (optional class, M6):
- DA-1 (minor): stale mechanism wording after the seam redesign — REQ-DP-007 (spec.md) still says "a function field replaced by a recording fake … ownerCheck precedent" and AC-DP-006's cell (acceptance.md) still says "the seam is a function field"; the plan now uses a spawn PARAMETER. The normative core (never spawn a real child; injectable seam) is unchanged and strengthened; one-word sweep ("injectable spawn seam") closes it. No behavior fork — not routed as blocking.
- DA-2 (minor): the registration-half single-count probe's theoretical unrelated-change mutant (parent-help text mentioning the verb name); pair the GREEN with an explicit parent-help count=0 assertion. The hiddenness half already asserts it in test form.
- DA-3 (nit): the `LEDGER-ACR-J` disposition label referenced in GREEN-B's form note is unlocatable in this tree (grep across `.moai/docs/` and `.claude/rules/`: no match). The substance is independent of the label — the four elements are present (command, out 0, exit 1, tree pinned) and the probe was re-executed this round. Fix the label or point it at a real record.

## Receipt

The leader's rc.27 mint landed at 14:02:18Z — `rcpt-25a06b9e9ee7c784137d0615` (tool `codex_audit`, tree_root this worktree, post-marker) — but its `codex_verdict` is **fail** again (no `gate_unmet`: per the tree's receipt code, a reviewed fail, not a rejection conversion). This is the second leader-side fail; both leader-side reviews differ from this auditor's three adversarial runs (pass/zero-findings at the same SPEC-artifact state) in scope — the leader's native/baseBranch surface reviews the whole branch diff, which includes these verdict files and the receipt-saga prose, where this auditor's runs were scoped to the SPEC artifacts. A fail receipt cannot corroborate a PASS verdict and is not cited. The findings text is needed from the leader's session to adjudicate: if it holds a real defect, the verdict flips honestly; if it flags audit-bookkeeping prose or configuration, the mint should be re-run via the adversarial+focus configuration that demonstrably returns pass on this tree. Until then the honest line is `receipts=none`.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: the two post-iter2 amendments are sound — the seam redesign closes a real cross-package test hazard, and the re-keyed registration gate is mutation-sensitive where the exit-0 form was blind.
- **Evidence**: all commands and outputs cited inline (probe 0/exit 1; anchors verified; verbs re-run; lint 0 findings).
- **Baseline-attribution**: everything measured this run, this tree (`eccd62531`), source-built `go run ./cmd/moai`; carried items carry their iter2 attribution.
- **Gaps**: the effect-half GREEN commands (TestDetachedChildPrunes) are form-verified only — no implementation exists to execute them (same class as iter2's noted residual); Windows vet/lint/coverage remain run-phase M4 obligations.
- **Residual-risk**: DA-2's unrelated-change mutant (low probability, test-half-covered); DA-3's unlocatable label (cosmetic).
