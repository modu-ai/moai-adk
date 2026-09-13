# Sync-Audit Report — SPEC-GATEWAY-ENVELOPE-REPAIR-001 (card t708)

- **FINAL VERDICT (iteration 2): PASS — overall 92/100** (harmonic mean; delta scope F1+F2+F7)
- Iteration 1 verdict (history, below): FAIL 78/100 — blocking findings F1, F2, F7
- **Iteration 2 overall score: 92/100** (harmonic mean of the re-scored dimensions)
- Auditor: sync-auditor (in-session; no `audit_model: multi` key in `.moai/config/` — the `workflow.yaml` `audit:` block carries backend model pins only, so per contract this audit ran without a cross-backend fan-out; same disposition as the plan-audit)
- Tree: worktree `.claude/worktrees/t708`, branch `WT-envelope-persist`, HEAD `7d7b29a1e` (re-verified `git rev-parse --show-toplevel` / `--short HEAD` at audit time); card diff measured against `CARD_BASE = git merge-base develop HEAD = 93ae49ce7` (the branch absorbed develop twice — merge-base is the discriminator, per the t543 lesson)
- Uncommitted tree state at audit time: `M .moai/specs/SPEC-GATEWAY-ENVELOPE-REPAIR-001/progress.md` (one modified tracked file)

## 1. Dimension Scores

| Dimension | Weight | Score | Verdict | Evidence (verbatim-command basis, this run, this tree) |
|---|---|---|---|---|
| Functionality | 40% | 70/100 | **FAIL** | `go test ./internal/gateway/conversation/ -run 'TestRepair' -count=1 -v` → 7/7 PASS; `go test ./internal/gateway/translate/ -run 'TestWedgePolicy|TestReceiptHistory' -count=1` → ok 1.265s; `go test ./internal/cli/ -run 'Gateway|Repair' -count=1` → ok 2.118s; `go test ./internal/gateway/receipt/ -count=1` → ok 1.389s; preserve lock `TestGatewayRepairCardDiffTouchesNoPreservedFile` + discriminator both PASS. 11 of 12 ACs verified green with real tests behind them. AC-EVR-004's byte-fidelity clause fails on real-world input and its stated evidence test does not exist (F1); AC-EVR-007's two named refusal cells are structurally unimplementable at the repair layer and §E.3 E1 substitutes different cells without recording the deviation (F7). |
| Security | 25% | 95/100 | PASS | Repair path imports only `opaque` (repair.go:10-23); single flag-gated call site (`gateway_repair.go:57`, only caller `gateway_session.go` `prepareGatewayConversation`); `TestRepairPathNeverReadsReceiptStore` PASS. Marker self-attestation sha256 before any injection (repair.go:190-195, 220-226); `opaque.Decode` canonicality ⇒ exactly one byte sequence satisfies a marker. All-or-nothing plan-then-inject (repair.go:155-236); refusal cells assert zero modification byte-for-byte. PRESERVE lock live: 0 violations vs merge-base + both-direction discriminator. Adversarial probes fail closed: non-hex 64-char marker digest → pool miss → refusal; tampered carrier → refusal; concurrent repairs arbitrated by O_EXCL aside. No validator/publish/guidance change (guidance built from `translate.HistoryReplayError`, gateway_repair.go:24-26). |
| Craft | 20% | 72/100 | PASS | `gofmt -l` on all 8 card Go files → clean; `go vet` (3 packages) → exit 0. @MX:NOTE+@MX:SPEC present on the exported surface (repair.go:69-72). Error wrapping `%w` + sentinel errors throughout. BUT: 5 new errcheck findings break the repo's clean lint baseline and the CI required check (F2); coverage gate per acceptance §D.3 (≥85% touched packages) not demonstrated — this run: conversation whole-package 79.6%, `go tool cover -func` repair.go: RepairEnvelope 75.5%, repairPlan 87.0%, writeExclusive 44.4% (matches §E.3's own figures; honestly recorded there with a caveat, but the gate as written is unmet); cli package coverage unmeasurable in-bounds (10m timeout, ~1583s baseline). |
| Consistency | 15% | 78/100 | PASS | AC matrix ↔ tests map 1:1 at the behavior level; t700 seam (spec §4, Reading B) verified against the sibling's actual landed text — REQ-WRR-008 scopes to gateway-state-changing recoveries and this repair changes none; REQ-WRR-007's rejection mandate is scoped to its own re-rooting policy. t703 verdict ground truth (`invalid conversation receipt` at the marker-without-envelope site) matches the repair target shape. CHANGELOG entry accurate (paths verified against `ls`; AC count 12/12; refusal-path byte claims true). Operator doc present, section checklist test green. Deductions: doc §3 byte promise overstated (F1), §3.3/AC-EVR-007 refusal claims unimplementable as written (F7), E1 label swap 008/009 (F6). |

**Harmonic mean**: 4 / (1/70 + 1/95 + 1/72 + 1/78) ≈ **78/100**.

**Must-pass firewall**: Functionality (must-pass) carries a criterion-level FAIL (AC-EVR-004 letter, AC-EVR-007 cells) → overall FAIL regardless of the other dimensions.

## 2. Findings

### F1 [High] [blocking] — AC-EVR-004 "no other byte" clause violated on real transcripts; JSON re-marshal rewrites the entire injected row; the stated evidence test does not exist

- `internal/gateway/conversation/repair.go:275-294` (`repairInjectLine`): unmarshals the row into `map[string]any` and re-marshals with `json.Marshal` — Go sorts map keys alphabetically and HTML-escapes `<`, `>`, `&`. **Mechanically proven this run** (standalone probe, /tmp/t708-probe): a client-shaped row with insertion-ordered keys and literal `a < b && c > d` round-trips 479 → 499 bytes, `byte-identical: false` — key order fully re-sorted, `<`→`<`, `&`→`&`.
- Real transcripts are client-written Node `JSON.stringify` rows (ground truth: `/Users/goos/.claude/projects/-Users-goos-MoAI-copythat-sweeper/1bfe6218-….jsonl` rows begin `{"type":"custom-title","customTitle":…,"sessionId":…}` — non-alphabetical insertion order). `refreshNative` walks `ConfigDir/projects` (native.go:24) — the client's own directory. So every repaired boundary row is byte-rewritten far beyond the injected envelope.
- Violates the letter of AC-EVR-004 ("no other byte of the replayed history changes") whose stated Evidence is "full-history byte-diff confined to injected `redacted_thinking` blocks", and the operator doc §3 claim "공개 콘텐츠(사용자 발화, 도구 결과, 텍스트, 도구 사용 블록)는 한 바이트도 바뀌지 않습니다" at file level — the injected row CONTAINS the tool_use (public) block being rewritten.
- Test gap: `internal/gateway/conversation/repair_test.go:210-215` comments "Every non-injected row is byte-identical" but asserts only `len(beforeLines) != len(afterLines)`. No byte-diff assertion exists; and the fixture builds rows with Go's own `json.Marshal` (repair_test.go:87-91), so the fixture is pre-canonicalized and cannot catch this defect class even if the assertion existed.
- Impact bound (honesty note): the re-marshal is semantically equivalent JSON — the client parses and re-canonicalizes into its own request, the Prefix hash excludes thinking content and is computed over canonicalized content, and the envelope carrier bytes survive verbatim (digest-verified, test green). The security determination (spec §3.2) is untouched. Residual semantic risk: `map[string]any` float64 conversion can corrupt integers > 2^53 in tool inputs (rare but real).
- **Required fix (either)**: (a) surgical splice — reconstruct the row preserving original bytes and insert only the envelope block into the content array; or (b) manager-spec amendment rescoping AC-EVR-004 + the doc to semantic equivalence, plus a real byte-diff test over a non-canonical (client-shaped) fixture row so the property is actually pinned. As written, the AC and the doc promise are not met.

### F2 [Medium] [blocking] — 5 new errcheck lint findings break the clean baseline and the CI required check; §E.3/E.5 lint claims were measured on a scope that omitted the package

- `internal/gateway/translate/receipt_history_wedge_policy_test.go:61, 88, 105, 115, 126` — errcheck: "Error return value is not checked" on `h.Check(...)` / `foreignRoot.Check(...)` in `goldenReplayOf(...)` argument position. All five are in the card's new M2 test file.
- Measured: `golangci-lint run --allow-parallel-runners --new-from-rev=93ae49ce7 internal/cli/ internal/gateway/conversation/ internal/gateway/translate/` → `5 issues: errcheck: 5`; full-package `./internal/gateway/translate/` → the same 5 (package was clean before the card); baseline spot-check `./internal/gateway/opaque/` → `0 issues.` CI gates golangci-lint as a required check (`.github/workflows/ci.yml:422,441-459`) → branch CI lint fails.
- Evidence-claim gap: progress.md §E.3 `new_warnings_or_lints_introduced: 0` and §E.5 "golangci-lint run --new-from-rev=4da5d1c4e internal/cli/ internal/gateway/conversation/ → 0 issues" are true as scoped but the scope omitted `internal/gateway/translate/` — the only package holding a new M2 file. No milestone ever linted the M2 file after creating it (M1 lint predates M2; M2's verification used `go vet` only; M3/M5 scoped to cli+conversation). VCI §1.1 surface 2: the aggregate "0 new lints for the card" is an unobserved claim.
- **Required fix**: capture `err := h.Check(...)` (or `foreignRoot`) before passing to `goldenReplayOf`, 5 sites — mechanical; then re-run full-package lint on all three packages.

### F7 [Medium] [blocking] — AC-EVR-007's "missing marker" and "Prefix-class public-content mismatch" refusal cells cannot exist at the repair layer; spec §3.3 overstates; §E.3 E1 substitutes cells without recording the deviation

- AC-EVR-007 names four refusal cells: source-gone, digest mismatch, **missing marker**, **Prefix-class public-content mismatch**, each "refuses with zero modification". Implementation reality:
  - *missing marker*: a boundary stripped of both envelope and marker is client-side indistinguishable from an ordinary text-only assistant turn — the repair cannot see it, so it cannot abort on it (t700's REQ-WRR-007 explicitly codifies this indistinguishability). For the CauseReasoning target class the marker provably survived (the classification fires at the marker-without-envelope site — t703 verdict §1), making the cell unreachable/vacuous for the repair path.
  - *Prefix-class public-content mismatch*: the repair has no public-content reference (receipt reads are forbidden by AC-EVR-010 — correctly so). It cannot refuse on Prefix mismatch; it injects, and the **unchanged Check** rejects downstream with the classified guidance. spec.md §3.3 states "the repair correctly refuses (Prefix-class mismatch, REQ-EVR-007)" — false as implemented; the refusal is composite (repair modifies, Check rejects), not a repair-side zero-modification refusal.
- E1 evidence substitution: the §E.3 matrix row for AC-EVR-007 lists "source-gone / digest-mismatch / **intact / incomplete** refusal cells PASS" — the last two are real, good tests (`TestRepairEnvelopeIntactHistoryNeedsNoRepair`, `TestRepairEnvelopeRequiresCompletedTransient`) but are NOT the AC's named cells. The M2 deviation note covers only the validator-side tamper shape, not this AC-EVR-007 substitution. Similarly AC-EVR-006's "no-flag no-op" cell has no test (behavior is structurally guaranteed — the sole `RepairEnvelope` call site is behind the flag scan at gateway_repair.go:32-40 — but the claimed cell does not exist), and AC-EVR-005's "marker-less boundary → refusal" names no existing cell.
- The code is arguably RIGHT and the SPEC text wrong — the proportionate fix is likely spec-side: an manager-spec amendment recording the composite-refusal semantics (repair injects only provable envelopes; the unchanged Check adjudicates everything else; aside preserves the preimage so the state is recoverable) + honest E1 cells. As written, AC-EVR-007's letter and §3.3's claim do not describe the implemented system.

### F3 [Medium] [optional] — aside-crash wedge: crash between aside write and record write permanently wedges repair; recovery undocumented

- `repair.go:116-136`: if a crash lands after `writeExclusive(asidePath)` (line 118) but before `writeExclusive(statePath)` (line 134), every retry passes the record-stat check, re-plans, then fails at the aside O_EXCL with "aside already exists from a prior partial attempt" wrapped in `ErrRepairNotRepairable` — refused forever, the single shot never burns, and nothing progresses. Manual aside deletion is the only exit.
- The operator doc §4 refusal table lists four shapes; this fifth (aside-exists) is absent, as is any row in spec §3.4. Conservative (zero modification, non-destructive), hence optional — but an undocumented permanent wedge with a manual recovery path nobody documented. Fix: document the recovery (delete `<transcript>.moai-repair-aside`, re-run) in the operator doc §4.

### F4 [Low] [optional] — multi-envelope single-boundary injection order is nondeterministic

- `repair.go:205-217`: `missing := map[string]bool{}` iterated to build `digests` — Go map order is random. When one row lost 2+ envelopes, the head-injection order (each prepend lands at content head) and the durable record's boundary order vary per run. Single-envelope boundaries (the t703 shape) are unaffected; the Check binds per-digest so acceptance is unaffected. Fix: sort digests (or order by first carrier appearance) for deterministic output and provenance.

### F5 [Low] [optional] — repair path confinement weaker than the package's own read path

- `repair.go:89-96` confines with `within()` + final-component `Lstat`; `transcriptModel` (native.go:100-111) additionally Lstats every ancestor component for symlinks. Exposure is TOCTOU-only (the path was discovered symlink-free by `refreshNative`'s WalkDir, which rejects symlink entries, native.go:38-40) inside a launcher-local trust domain. Hardening note for consistency with the sibling read path.

### F6 [Low] [optional] — progress.md §E.3 E1 matrix rows AC-EVR-008/009 swapped vs acceptance.md

- acceptance.md: 008 = single-shot durable termination; 009 = non-destructive aside + provenance. The §E.3 E1 table cites the aside/provenance evidence on the 008 row and `TestRepairEnvelopeSingleShotTerminatesAcrossRestart` on the 009 row. Both behaviors are tested and green — the row labels are swapped. Cosmetic.

### Positive findings (verified sound)

- Marker self-attestation is cryptographically sound: `BindToolID`-embedded digest + `opaque.Decode` canonicality mean exactly one byte sequence satisfies a marker; the pool is keyed by that digest, so cross-boundary transplant and manufactured envelopes fail closed.
- The preserve lock is exemplary per verification-completeness: live assertion skips LOUDLY post-merge (never vacuous green), the empty-diff case fails ("unmeasurable, not clean"), and the discriminator's red is observed on a synthetic violating input (`gateway_preserve_test.go:62-117`).
- The M2 golden bodies are full literals, deliberately duplicated from receipt_history.go so an edit to either half of the composed body cannot slip a substring check — correct characterization-lock discipline.
- Guidance string is constructed from the validator's own error type (`translate.HistoryReplayError{Cause: CauseReasoning}.Error()`) — no string duplication, REQ-EVR-007 honored.
- t700 seam handled exactly as the plan promised: adjudicated at M0 with two inputs, Reading B verified sound against the sibling's actual text, final confirmation correctly deferred to the landing-time absorb.

## 3. Verdict rationale

The card's core mechanism — explicit-invocation, single-shot, all-or-nothing verbatim re-injection verified by marker self-attestation, aside-before-replace, no receipt-store access, unchanged-Check adjudication — is real, well-tested, and security-sound. 11 of 12 ACs are green behind genuine tests.

The FAIL is driven by three blocking findings, all with small, mechanical remediation paths:

1. **F1**: the byte-fidelity clause of a MUST criterion is violated on real-world input (proven by probe) and the claimed evidence test does not exist — fix is a surgical splice or an honest SPEC amendment + a real client-shaped byte-diff test.
2. **F2**: 5 new errcheck findings fail the repo's clean lint baseline and the CI required check; the "0 new lints" evidence claim was measured on a scope that omitted the affected package — fix is 5 one-line captures.
3. **F7**: two of AC-EVR-007's four named refusal cells are structurally unimplementable at the repair layer and spec §3.3 claims a refusal the repair cannot perform — fix is a SPEC amendment recording the composite-refusal semantics + honest E1 cells.

Re-audit scope on resubmission: the F1/F2/F7 delta only — re-verify the three fixes, re-run the lint batch and the conversation/translate/cli scoped tests. F3-F6 are optional and do not gate.

## 4. Gaps (what this audit did NOT observe)

- `internal/cli` full-package test run (10m timeout at ~601s twice for the card; ~1583s package baseline) — CI owns the full verdict; this audit ran the scoped `Gateway|Repair` selector (38-surface) only.
- No local full-suite run (lane load discipline); CI owns it.
- `GOOS=windows` cross-build not re-executed by the auditor (relied on §E.2 records; native build is implicitly proven by test compilation).
- Real-transcript byte-shape sampled from two unrelated local sessions' transcripts, not from the t703 production family itself (that family's transcript was not available in this tree).
- M5 live probe remains skipped (card-recorded): live client encode-time behavior of a repaired transcript against a real replay is unmeasured — the in-repo tests prove the transcript-side machinery and unchanged-Check composition, not external encode-time behavior.
- Coverage gate (acceptance §D.3 ≥85% on touched packages) not demonstrated: conversation 79.6% whole-package this run; cli unmeasured in-bounds. Recorded honestly by the card; unresolved as a gate.
- t700 remains unlanded (draft in the sibling worktree); the §4 seam's final confirmation re-rides the landing-time absorb per M0.

## 5. Residual risk

Even after the F1/F2/F7 fixes, the repair's real-world efficacy rests on the unmeasured premise (card-recorded) that client compaction/`--continue` slicing retains the verbatim carriers in the family transcript; the repair refuses safely (source-gone) when they are gone, so the residual is "repair unavailable", never "repair corrupts". The F1 semantic-equivalence exception (float64 > 2^53 in tool inputs) survives fix option (b) and is only closed by fix option (a).

---

Auditor: sync-auditor (in-session, skeptical stance; no cross-backend fan-out — `audit_model` not set to `multi`)
Baseline attribution: all commands run in this audit session against worktree `.claude/worktrees/t708` @ `7d7b29a1e`, card diff vs merge-base `93ae49ce7`.

---

## Iteration 2 — Delta Re-Audit (F1 + F2 + F7) — PASS 92/100

- **FINAL VERDICT (iteration 2): PASS — overall 92/100** (harmonic mean; no must-pass dimension below threshold; no blocking findings outstanding)
- Auditor: sync-auditor (in-session; delta scope per the re-audit request — F1+F2+F7 only; passing areas not re-litigated except where the delta touches them)
- Tree: same worktree, HEAD now `9e71492f7`; delta commits verified: `3c48598b7` (iter-1 report preserved as-is for history), `c9e98df4c` (F7 spec amendment, manager-spec), `b46d33271` (F1 splice + F2 lint + F3 doc line), `9e71492f7` (reopen status flip + F6 row swap)

### Delta verification (each command run by the auditor this session, this tree)

**F1 — byte-preserving splice: VERIFIED FIXED.**
- `repairInjectLine` replaced by a streaming `json.Decoder` token-walk (`repairContentInsertOffset`, repair.go:282-334) that captures `dec.InputOffset()` just past the `message.content` opening bracket, skipping sibling values via `json.RawMessage` (no re-serialization); the injection is a pure splice: `line[:pos] + canonicalBlock (+ "," unless empty array) + line[pos:]` (repair.go:340-355). Every byte outside the inserted span is the input's own byte — key order, escaping, whitespace preserved. The envelope block is marshaled from a fixed-field struct (deterministic); the carrier digest binds decoded canonical bytes, so block spelling is irrelevant to the Check.
- The new test `TestRepairInjectLinePreservesRowBytesOutsideInjection` (+ empty-array subtest) uses a genuinely client-shaped row (`"type"` first, `"sessionId"` last, literal `a<b&c` inside tool input) and asserts the **reconstruction invariant**: removing exactly the injected span from the output reproduces the input byte-for-byte — precisely the property AC-EVR-004's evidence line demands, now actually asserted (the iter-1 gap). The literal-`a<b&c` survival assertion closes the HTML-escape vector. RED was captured against the old splice (commit `b46d33271` message records the alphabetical-reorder + escape reproduction).
- Run this session: `go test ./internal/gateway/conversation/ -run 'TestRepair' -count=1 -v` → **8/8 PASS** including the new test and its empty-array subtest; the integrated `TestRepairEnvelopeRestoresStrippedEnvelopeAtOriginalBoundary` still passes through the new splice. AC-EVR-004's byte-fidelity clause is now met on real-world input shape AND pinned by a test that could fail.

**F2 — lint gate: VERIFIED FIXED.**
- All 5 call sites now capture the Check error and discard the `goldenReplayOf` return explicitly (`_ = goldenReplayOf(t, err, …)`), per the diff.
- Independent measurement this session (auditor-run, unscoped, not the orchestrator's): `golangci-lint run --allow-parallel-runners ./internal/gateway/conversation/ ./internal/gateway/translate/ ./internal/cli/` → **`0 issues.`** The translate package is back to its pre-card clean baseline; the CI required check is unblocked.

**F7 — composite-refusal semantics: VERIFIED FIXED.**
- spec.md v0.2.2 (`c9e98df4c`, authored by manager-spec — correct ownership for body content): §3.3 retitled "Refusal semantics (repair-layer vs composite)"; repair-layer triggers enumerated as exactly four (source-gone, digest-mismatch, already-attempted, incomplete-transcript); missing-marker and Prefix-class mismatch reclassified COMPOSITE with the honest rationale (client-side indistinguishability; receipt reads forbidden by REQ-EVR-003-4); §3.4 cells reworded minimally; HISTORY row recorded.
- acceptance.md AC-EVR-007 aligned to the same split (lines 47-50): repair-layer preconditions refuse with zero modification; composite cells delegate final adjudication to the unchanged Check via the M2 characterization matrix (AC-EVR-003). This matches the implemented system.
- **§E.3 E1 row compatibility assessment (the coordinator's specific question)**: progress.md:119 reads "source-gone / digest-mismatch / intact / incomplete refusal cells PASS, zero-mod asserted". Ruling: **compatible subset, no contradiction.** Three of the four named cells (source-gone, digest-mismatch, incomplete) map 1:1 to the amended repair-layer triggers; the fourth ("intact" = no-stripped-boundary) is a real, green, zero-mod refusal cell (`TestRepairEnvelopeIntactHistoryNeedsNoRepair`) that §3.4's shape table and the operator doc both retain — the amended AC does not forbid it. The amended row's fourth trigger (already-attempted) is cited on the AC-EVR-008 row (`TestRepairEnvelopeSingleShotTerminatesAcrossRestart`, green) — coverage exists, distributed across rows rather than contradicted. The composite cells' adjudication rides the M2 matrix per the amended Evidence line. No finding raised.

**Also verified (bonus repairs on optional findings):** F3 aside-crash wedge now documented with manual recovery in the operator doc refusal table (gateway-envelope-repair.md:59); F6 E1 rows 008/009 label swap fixed (progress.md:120-121 now matches acceptance.md's definitions). Status reopened to `in-progress` (`9e71492f7`) — correct lifecycle for re-close after this PASS.

### Iteration-2 delta findings

No new blocking findings. Residual notes (non-gating, carried or new):
- **N1 [Low] [optional]** — The new `TestRepairInjectLinePreservesRowBytesOutsideInjection` pins the splice on a synthetic client-shaped row; it does not feed a non-canonical row through the full `RepairEnvelope` → transcript-write → re-read pipeline. The integrated test still runs on Go-canonical fixtures. The property proven (pure splice) plus the integrated digest test jointly cover the risk; noted for a future hardening pass, not required.
- **N2 [Low] [optional]** — spec §3.3's "exactly four" repair-layer triggers is a simplification: the code also refuses on non-JSON rows, undecodable envelopes, path escape, unreadable transcript, and the aside-exists wedge — all fail-closed variants of the four (or of the no-work case). Honest enough at the semantic level; a future amendment could say "four trigger classes".

### Iteration-2 dimension scores (delta-rescored)

| Dimension | Iter-1 | Iter-2 | Basis |
|---|---|---|---|
| Functionality | 70 FAIL | **92 PASS** | AC-EVR-004 byte-fidelity now met and tested (F1); AC-EVR-007 letter aligned to implemented semantics (F7). Residual deduction: coverage gate per acceptance §D.3 still not demonstrated at package level (unchanged Gap); AC-EVR-006 no-flag cell still structurally-guaranteed-only. |
| Security | 95 PASS | **95 PASS** | Splice reviewed: fail-closed offset walk, insertion of a fixed-field struct only, no new surface; digest binding unaffected. |
| Craft | 72 PASS | **88 PASS** | Lint gate restored (`0 issues.` unscoped, auditor-run); byte-identity test is exemplary verification discipline. Residual: coverage figures unchanged. |
| Consistency | 78 PASS | **93 PASS** | Doc byte-promise now true at file level; F7 spec/code/doc aligned; F6 fixed; CHANGELOG unchanged and still accurate. |

Harmonic mean: 4 / (1/92 + 1/95 + 1/88 + 1/93) ≈ **92/100**.

### Iteration-2 verdict

**PASS.** All three blocking findings from iteration 1 are repaired and independently re-verified; the delta introduced no new defects; the optional F3/F6 items were closed as well. The orchestrator may re-close the SPEC (status `in-progress` → `completed` riding the close commit). Carry-forward Gaps (non-gating, unchanged from iteration 1): coverage ≥85% on touched packages not demonstrated at package level; `internal/cli` full-package run unmeasured locally; M5 live probe skipped (live client encode-time behavior unmeasured); t700 landing-order confirmation re-rides the develop absorb.
