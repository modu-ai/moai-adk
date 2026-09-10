# Sync Audit Report — SPEC-AUDIT-PARTICIPANT-COUNT-001

**Auditor**: sync-auditor (independent, fresh-context)
**Tree / HEAD attribution**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t284`, branch `WT-audit-participant-count`, HEAD `299c40f8c` (audit performed 2026-09-02 on this tree; run-phase code audited at its landing commit `5713f82fa` with zero drift verified to HEAD).
**Audit scope**: SPEC artifacts v0.2.1 (closed), run diff `53a3fc1dd..5713f82fa -- internal/cli/`, sync diff `53a3fc1dd..HEAD`, committed evidence under `.moai/reports/t284/`.

## Verdict

| Dimension | Weight | Score | Verdict |
|---|---|---|---|
| Functionality | 40% | 98 | PASS (must-pass) |
| Security | 25% | 100 | PASS (must-pass) |
| Craft | 20% | 96 | PASS |
| Consistency | 15% | 98 | PASS |
| **Overall (harmonic mean)** | | **97.9 ≈ 98** | **PASS** |

Must-pass firewall: Functionality and Security both independently above threshold; no blocking findings. One optional cosmetic finding (F1).

---

## Claim / Evidence / Baseline-attribution

Every load-bearing claim below was **re-measured by this auditor on this tree in this run** (commands and observed outputs inline). Committed evidence logs were additionally cross-checked for authenticity (pre-GREEN byte signatures, line-number correspondence against the current test file).

### Check 1 — AC sweep (8 criteria): all 8 PASS, re-executed

Deciding command re-run by auditor:

```
$ go test ./internal/cli/ -run 'AC_APC' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli	1.047s   (exit 0)
```

Non-empty sweep confirmed (8 test functions in `e1-ac-matrix.log`; subtests enumerated a–h / per-case). Per-criterion substance assessment (mutant-shaped reasoning per verification-completeness §2):

| AC | Deciding test | Pins the substance? |
|---|---|---|
| AC-APC-001 | `TestConverge_ParticipantCount_Table_AC_APC_001` | **Yes** — exact-count equality on all 8 rows including row g (gate-`off` carrying `pass`, the REQ-APC-002 boundary) and row h (all-inconclusive). Counting-inconclusive mutant dies on rows c/h; gate-check-removed mutant dies on row g; advisory-excluding mutant dies on row d. Row c doubles as the witness for acceptance §D's second mutant. |
| AC-APC-002 | `TestConverge_BelowTwo_NoDivergence_FlagNilNotFalse_AC_APC_002` | **Yes** — asserts `r.DisagreementFlag != nil → fail` (test line 215), the exact nil-vs-pointer distinction; NOT the non-compliant `flag != nil && !*flag` shape the criterion forbids. Covers §A.2 cases 1/3, empty slice, all-inconclusive, `RequiredFailWithInconclusive` (the deliberate false→null move), and the DQ-2 refusal path via real `runMultiAudit` with a stubbed `backendCall`. |
| AC-APC-003 | `TestConverge_BelowTwo_NoDivergence_JSONNull_AC_APC_003` | **Yes** — triple assertion: member present (`present` check, line 70), value nil (line 73), raw bytes free of `"disagreement_flag":false` (line 76). An `omitempty` mutant dies on presence; a false-emitting mutant dies on value AND bytes. Fails independently of AC-APC-002 as designed. |
| AC-APC-004 | `TestConverge_TwoMeasuredInputs_DerivedSummaryDiffers_AC_APC_004` | **Yes** — asserts both positions (count 3-vs-1, flag false-vs-null) plus two inequality witnesses asserting the derived summaries differ. Asserts the derived positions, not byte inequality (which already held pre-change) — exactly the §A.3 premise. |
| AC-APC-005 | `TestConverge_TwoOrMore_BooleanUnchanged_AC_APC_005` | **Yes** — six ≥2 cases each carrying its measured count, asserting non-nil (`nil → Fatal`) and the exact pre-change boolean; includes the inclusive-2 boundary (`NoRequiredBackends_VacuousPass`). The two excluded cases are excluded exactly as acceptance §A specifies, each owned elsewhere (002/003 and 008). |
| AC-APC-006 | `TestConverge_Undetermined_GatesNothing_AC_APC_006` | **Yes** — real `persistConvergenceResult` → `HandleMultiReviewGate` round trip with gate enabled + change detected: ALLOW on undetermined; and the negative direction (overall fail still BLOCKs through the one existing path). A mutant making the undetermined state block would fail the ALLOW assertion; a new block category would fail the BLOCK-only-on-fail assertion. |
| AC-APC-007 | `TestLoadConvergenceResult_OldStateFile_Decodes_AC_APC_007` | **Yes** — hand-written old-shape fixture (boolean flag, no `participant_count` member — a shape the current struct can no longer produce, so hand-authoring is the only honest source); asserts `ok==true`, non-nil pointer holding the recorded `true`, zero-value count, and unchanged gate BLOCK decision. |
| AC-APC-008 | `TestConverge_SingleParticipantDivergence_CarveOut_AC_APC_008` | **Yes** — count 1; `nil || !*flag → Fatal` (cannot pass with null OR with false — the carve-out's two failure directions); serialized boolean `true`; overall/fail_open unchanged vs a note-free baseline; gate ALLOW through a real persist round trip. The pre-existing regression guard also re-verified green (below). |

Regression guards and invariants re-executed by auditor (all PASS, exit 0): `TestConverge_SurfacesSignalDivergence_WithoutBlocking` (landed REQ-CVS-003 behaviour kept through the carve-out), `TestConverge_RequiredFailWithInconclusive_Case2` (now asserts nil with rationale), `TestNoVerdictDisagreementEnum_EC7_AC_AMM_011` (REQ-AMM-008), `TestConvergence_NoAskUserQuestion_AC_AMM_024` (C5), `TestBuildIdentityOmittedWhenAbsent` (t248 key-set surface, F-1's second extra).

### Check 2 — Implementation semantics: exact against REQ-APC-001..005

Read `converge(...)` Step 2c (`mcp_convergence.go:233-249`) and `countParticipants` (`:311-322`) at HEAD:

- **Participant definition (REQ-APC-002)**: `gate == config.AuditGateOff → skip`; count iff `verdict == "pass" || verdict == "fail"`. Exact.
- **≥2 (REQ-APC-004)**: `disagreementFlag = &disagreement` — the three-pass boolean (Steps 2/2a/2b) untouched; Step 1 (overall verdict) untouched; note derivation untouched. `multi_review_gate.go`, `mcp_audit_multi.go`, `mcp_server.go` carry **zero diff** `53a3fc1dd..HEAD` (verified by `git diff --stat` over each path — empty).
- **<2 + synthesis notes (carve-out)**: non-nil `true`. **<2 without**: `disagreementFlag` stays nil. **DQ-2 refusal** (`runMultiAudit`, `:607-620`): literal omits both fields → count serializes as visible `0` (no omitempty) and flag as `null`.
- **No `omitempty`** on either tag: `DisagreementFlag *bool \`json:"disagreement_flag"\`` and `ParticipantCount int \`json:"participant_count"\`` — verified in the struct at `:121` / `:135`.
- **Gate reads neither field**: `grep -n 'DisagreementFlag\|ParticipantCount' internal/cli/multi_review_gate.go` → 0 hits. C3 holds structurally.
- **Result-literal census**: all non-test `ConvergenceResult{` sites are `converge` (`:274`, both fields set), the DQ-2 refusal (`:608`, intended zero values), and three zero-value error returns in the gate's loader (`multi_review_gate.go:112/117/121`, `ok=false` fail-open paths — never gated on). No producer was missed.
- **audit_multi declares no output schema** — verified at `mcp_server.go:405-425`: the registration carries no `WithOutputSchema`; the two schema sites (`:259`, `:400`) belong to `codex_audit`/`glm_audit`. Adding a field breaks no declared schema.

### Check 3 — Evidence chain: authentic and resolvable on this tree

All cited paths exist as committed files on this tree (`ls .moai/reports/t284/`). Authenticity cross-checks:

- **RED Block 1 (runtime, pre-GREEN at `53a3fc1dd` + new test file)**: the quoted JSON bytes in `red-stage1-runtime.log` contain **no `participant_count` member** and `"disagreement_flag":false` — a byte signature only the pre-implementation engine can produce. Not post-hoc.
- **RED Block 2 (compile, same tree)**: `red-stage2-compile.log` shows `r.ParticipantCount undefined` / `mismatched types bool and untyped nil` against the unchanged engine — the compile-coupling justification for the M1+M2 atomic commit is real.
- **Mutant discharge (`mutant-red.log`)**: failure set is exactly acceptance §D's witness table — AC-APC-002/003/008 red — **plus** the disclosed extra red (AC-APC-001's folded §B empty-slice subtest asserting the nil flag, `:174`), which progress.md §E.2 reports rather than hides; the criterion's own count rows stay green under the mutant ("the mutant counts correctly"). The mutant's signature is visible in the bytes (`"participant_count":1` alongside `"disagreement_flag":false` on a sub-2 input). All cited line numbers (`:74`, `:77`, `:174`, `:216`, `:423`) match the **current** test file exactly. Post-revert green re-verified independently (auditor's own E1 re-run, exit 0).
- **No drift**: `git diff 5713f82fa..HEAD -- internal/cli/mcp_convergence.go` → empty.
- **Stage-1 line numbers (`:67/:70/:92...`) differ from the final file** — expected: stage 1 ran while the test file still contained only the map-based tests; the file grew to its final layout before stage 2 (whose citations match the final layout). Internally consistent, not a defect.
- **premise-probe.log** (`.moai/state/verify/t284/premise-probe.log`) resolves on this machine — see Residual-risk for its machine-local nature.

### Check 4 — Docs honesty: figures match, locales faithful, mirrors byte-identical, neutral

- **CHANGELOG ↔ progress.md §E.2**: 8/8 ACs, three mutant witnesses observed RED, F-1 (inventory 12 → measured 14, both extras in `mcp_build_identity_test.go`), 100% on the four named functions, root package 80.1% pre-existing (≤ +0.12pp bound), 11 docs surfaces, `ko/autonomous-loops.md` exclusion — every figure in the CHANGELOG entry traces to progress.md and to evidence this auditor re-measured. Position verified: first entry under `[Unreleased] → Added`, above SPEC-AUDIT-BUILD-IDENTITY-001. B12 self-tests re-verified: pre-emission count 0 → now `grep -c` = 1; acceptance.md distinct AC identifiers = 8 = entry's "8/8"; all file paths named in the entry exist on this tree (ls-verified).
- **Coverage re-derived from the committed profile** (not trusted from prose): `go tool cover -func=.moai/reports/t284/e3-cover.out` → `converge 100.0%`, `countParticipants 100.0%`, `runMultiAudit 100.0%`, `HandleMultiReviewGate 100.0%`, `total: (statements) 80.2%` — matching progress.md §E.2's figures exactly (root package 80.1% per the run log line).
- **docs-site (7 files)**: `multi-model-audit.md` ×4 locales each gained a same-position paragraph stating participant_count + the three-valued flag + the carve-out in that locale's own prose (en/ko/ja/zh read independently — native phrasing, not calque); `autonomous-loops.md` ×3 (en/ja/zh) gained the three-state sentence. `ko/autonomous-loops.md` untouched exactly as plan §G directs (pre-existing 4-locale parity gap, recorded in spec.md §E).
- **Skill + workflow surfaces**: `moai-ref-cross-model-audit/SKILL.md` (JSON example `"participant_count": 3`, field bullets, three-state definition, case-table sub-2 invariant, carve-out) and `workflows/review.md` (folding-table `null` row) updated.
- **Template mirrors byte-identical**: `diff` local↔template for both SKILL.md and review.md → rc=0, no output. **Neutrality**: `grep -rn 'SPEC-AUDIT-PARTICIPANT-COUNT-001' internal/template/templates/` → **0 hits**.
- **catalog.yaml**: hash for `moai-ref-cross-model-audit` updated; `go test ./internal/template/ -run 'Catalog'` → `ok` (exit 0) — the committed hash matches the tree.

### Check 5 — Scope discipline: nothing outside declared surfaces

`git diff 53a3fc1dd..HEAD --stat` (33 files) decomposes exactly into: `internal/cli` ×6 (run scope), SPEC artifacts ×4, `.moai/reports/t284/` ×10 (evidence + plan-audit records), `CHANGELOG.md`, docs-site ×7, local skills ×2, template mirrors ×2, `internal/template/catalog.yaml` — all declared surfaces (plan §B/§E/§G). **Run-phase doc abstinence** (acceptance §C DoD item 5): `git diff 53a3fc1dd..5713f82fa --stat -- docs-site .claude/skills internal/template/templates` → **empty**. The sync commit `918d65366` is docs-only (15 files, all sync scope); backfill `299c40f8c` is the one progress.md line (canonical D3 exemption pattern, and `sync_commit_sha: "918d65366"` in §E.4 matches the real sync commit SHA). Worktree clean (`git status --porcelain` empty pre-audit).

### Additional mechanical verifications (this run)

```
$ go vet ./internal/cli/...                          → exit 0
$ go build ./...                                     → exit 0 (darwin/arm64)
$ GOOS=windows GOARCH=amd64 go build ./...           → exit 0
$ golangci-lint run --timeout=2m ./internal/cli/...  → "0 issues." exit 0
```

---

## Findings

- **F1** [MINOR] [optional] `CHANGELOG.md:12` + `progress.md` §E.2 (E3 bullet) — the phrase "every function this SPEC added or changed at 100% (… `HandleMultiReviewGate`)" overstates the changed-set: `multi_review_gate.go` carries zero diff across `53a3fc1dd..HEAD`, so `HandleMultiReviewGate` is an *exercised consumer* of the narrowed struct, not a function this SPEC changed. The 100% figure itself is real (re-derived from the committed profile). Required fix (if ever touched again): reword to "every function this SPEC added or changed (`converge`, `countParticipants`, `runMultiAudit`), plus the gate consumer `HandleMultiReviewGate`, all at 100%". Cosmetic; no behavior implication.

No other findings. No blocking findings.

## Gaps (explicitly NOT observed)

- **Mutant re-execution not re-performed by this auditor**: the constraint "do not modify code" binds this audit, so the representative mutant was not re-applied by me. Its discharge is instead verified by (a) assertion-shape analysis (each witness's failure under the mutant is entailed by the assertion text read at HEAD), (b) byte-signature and line-number correspondence of `mutant-red.log` against the current tree, and (c) the independently re-run green suite. Confidence high; re-execution would have made it mechanical certainty.
- **CI verdict**: the card branch is unpushed (lanes do not push); no `origin/develop` CI run exists for `299c40f8c`. The full-suite verdict belongs to the integration window (§E.4 `push: not-performed`).
- **`make build` not re-executed by this auditor** (it writes `bin/`); catalog consistency was instead verified via the template package's Catalog tests against the committed hash, which is the same invariant the embed axis needs at commit time.

## Residual-risk

- `spec.md` §A.2/§F cite `.moai/state/verify/t284/premise-probe.log` — machine-local scratch that resolves on this machine but will not survive a fresh clone. The §A.2 table is quoted verbatim inside spec.md itself, so the substance is durable; only the raw bytes are machine-bound.
- AC-APC-001's table (the contract) exercises gate-`off` carrying `pass` (row g) but not gate-`off` carrying `fail`. The implementation's gate check is verdict-independent (`continue` before the verdict test), and gate-`off` entries never reach `PerBackendVerdicts` through the fan-out (`mcp_convergence.go:571-573`), so the combination is test-constructible only. Noted as a residual, not a finding — the criterion's table is satisfied as written.
- Non-Go consumers reading `disagreement_flag` as a strict boolean will now see `null` in the sub-2 case. Plan §H measured the only Go reader as the engine itself (re-confirmed: gate reads neither field; `audit_multi` declares no output schema); JSON readers see the docs-site/SKILL surfaces this SPEC updated. Any third-party strict-boolean reader outside this repository is inherently outside measurement.

## Recommendation

PASS as audited. F1 is cosmetic and safely deferrable to any future touch of the CHANGELOG entry; nothing requires action before the lead's develop integration window.
