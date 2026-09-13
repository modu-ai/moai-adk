# Sync Audit — SPEC-GATEWAY-WEDGE-REROOT-001 (card t700)

- Auditor: sync-auditor (independent, fresh judgment)
- Tree: worktree `.claude/worktrees/t700`, branch `WT-wedge-reroot-policy`, HEAD `f5d865d38` (sync commit)
- Base: merge-base `develop HEAD` re-derived **now** = `643abfb8c` (matches the M4 literal pin; local `develop` has since advanced to `4f5a9006f` via other lanes — merge-base unchanged)
- Date: 2026-09-14

## Overall Verdict: **PASS** — score 94.9/100 (harmonic mean over the 4 dimensions)

Must-pass firewall: Functionality PASS, Security PASS — both above threshold independently. No blocking findings; no repair round required before the card leaves the lane.

## Dimension Scores

| Dimension | Score | Verdict | Evidence (verbatim mechanical output, this run, this tree, HEAD f5d865d38) |
|-----------|-------|---------|----------|
| Functionality (40%) | 98 | PASS | `go test ./internal/gateway/translate/ -run TestReceiptHistory -count=1 -v` → `ok ... 1.063s` (12/12, incl. the three new locks + `TestReceiptHistoryRejectsMidPairTruncationBeforeHistory` behavioral net); `go test ./internal/cli/ -run TestGatewayReroot -count=1 -v` → `ok ... 1.062s` (11/11). AC-WRR-012 re-run: `git diff --name-status 643abfb8c..HEAD` → 14 files, zero hits on the five frozen files and zero paths under `internal/gateway/receipt/`. AC-WRR-014 re-run: `grep -rn "\.Fork(" internal/ --include="*.go" \| grep -v _test` → exactly `internal/cli/gateway_session.go:232` (baseline `643abfb8c:internal/cli/gateway_session.go:204` — same call site, 1→1). AC-WRR-015: 6/6 element greps ≥1 on `.moai/docs/gateway-wedge-recovery.md`. AC-WRR-013 SKIPPED with recorded gap (MINOR, non-gating — orchestrator disposition §E.2 M1, not re-litigated). |
| Security (25%) | 97 | PASS | Golden-string lock verified by reading the test: `TestHistoryReplayErrorGoldenStrings` pins all three `HistoryReplayError` messages as full byte literals independent of production constants (any constant edit fails). Forged-tail lock re-run green (`TestGatewayRerootForgedTailNeverRecoverable`: remainder accepted, reintroduction rejected). Structural no-store lock: `grep OpenStore\|receipt internal/cli/gateway_reroot.go internal/gateway/conversation/reroot.go` → 0 code matches; the source-level test `TestGatewayRerootPathHoldsNoReceiptStore` + store-digest complement both green. Single-shot: marker `O_CREATE\|O_EXCL` written BEFORE any mutation (gateway_reroot.go:174), fresh-invocation refusal green. `TranscriptPath` reuses the resume path's `within()` containment + per-component `Lstat` symlink-escape walk (reroot.go:25-42), refusal tested (`ErrInvalid`). `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (E2 independently reproduced). |
| Craft (20%) | 90 | PASS | Coverage re-measured this run: `go test -cover ./internal/gateway/translate/ -count=1` → `coverage: 91.9%` (matches §E.2 exactly); `./internal/gateway/conversation/` → `80.9%` package (claimed 80.7%; pre-existing baseline 77.0% — improving trend, package figure dominated by pre-existing untested branches). Per-function (`go tool cover -func`): `TranscriptPath 89.5%`, `rerootGatewayTranscript 85.4%`, `rerootRefusalError 66.7%` — all three match §E.2's claims exactly. `golangci-lint run --new-from-rev=643abfb8c ./internal/cli/... ./internal/gateway/...` → `0 issues.` `go vet ./internal/cli/ ./internal/gateway/...` → exit 0. TDD RED evidence verbatim in progress.md M3 (build-failure listing naming `rerootGatewayTranscript` undefined — credible pre-GREEN shape). |
| Consistency (15%) | 95 | PASS | Conventional Commits with card id on all 9 commits; CHANGELOG entry honest ("16 defined / 15 PASS", AC-WRR-013 skip disclosed, `pending-backfill-sync` placeholder disclosed, cited paths verified); spec.md `status: completed` transition carried by the sync commit (version 0.1.1 bump is plan-phase work documented in the HISTORY table — no §E.4 contradiction); §E.2–§E.4 structure matches the progress.md Section Map; code style matches the surrounding tree (`%w` wrapping, English comments, unexported helpers, marker/aside sidecar naming). One cosmetic typo in the operator doc (F5). |

## Findings (coverage-first — all optional, none blocking)

- **F1** [LOW] [optional] `internal/cli/gateway_reroot.go:167-184` — The attempt marker records `Aside: asidePath` before the aside file is written. If the process crashes or the aside write fails inside that window, later invocations render guidance "the removed content remains preserved at <path>" pointing at a file that does not exist. The single-shot consumption itself is correct-by-design (marker-first prevents the worse double-removal failure); only the guidance string can go stale-false. Required fix (if taken): `os.Stat(asidePath)` existence check before appending the aside pointer to the guidance. Confidence: high (read from source); severity: low (narrow window, no data lost, transcript untouched).
- **F2** [LOW] [optional] `.moai/docs/gateway-wedge-recovery.md` §7 table — "무엇을 지우나: 끝의 발행 안 된 어시스턴트 경계" states the never-published qualifier the client cannot verify (spec §D + AC-WRR-016 acknowledge the indistinguishability; §5 of the same doc documents it correctly). On a non-wedge conversation, `--reroot` removes a published boundary (validator-accepted prefix, aside-preserved). Required fix (if taken): one clarifying sentence in §7 pointing to §5's indistinguishability row. Confidence: high; severity: low (documentation nuance; behavior is spec-sanctioned).
- **F3** [LOW] [optional] `internal/cli/gateway_reroot.go:196-200` — Transcript rewrite via `tmp + Rename` resets the file mode to 0600 regardless of the original transcript's mode. Native transcripts are 0600 so practically moot. Required fix (if taken): `os.Stat` the original and `os.Chmod` the tmp before rename. Confidence: high; severity: low.
- **F4** [INFO] [optional] Bookkeeping — `§E.3 total_run_phase_files: 13` and the dispatch's "13 files" vs the HEAD diff's 14 files (sync commit adds CHANGELOG.md; M7 added `conversation/reroot_test.go`). Each figure was correct at its own measurement time (M4 recorded 11); no substantive discrepancy. Required fix: none; record-only.
- **F5** [MINOR] [optional] `.moai/docs/gateway-wedge-recovery.md:59` — typo "제거된 경위" (should be "경계"). Required fix: one-character correction at next doc touch. Confidence: high; severity: cosmetic.
- **F6** [NOTE] [optional] AC-WRR-008's "fresh-process re-invocation" is tested as a second in-process call with a documented statelessness argument (the function holds no state; the on-disk marker is the only record — verified by source read: no package-level state in `gateway_reroot.go`). The code path is genuinely identical to a fresh process; a true exec-based test would be stronger at subprocess cost. Required fix: none; noted for the record.

## Recommendations

- Consider F1's existence check opportunistically if any follow-up card touches `gateway_reroot.go`; do not open a repair round for it.
- Fold F2/F5 into the next documentation pass over the operator doc.
- The full-suite verdict remains CI's on the lead's `origin/develop` push (lane doctrine); the local 600s-timeout UNRESOLVED classification in §E.3 is honest and correctly scoped.

## Gaps (explicitly NOT observed by this audit)

- Full `internal/cli` suite locally (600s default-timeout abort shape on this loaded machine; package historically ~1583s) — UNRESOLVED by design; change-scoped selectors green ×3 (this audit re-ran two of them).
- AC-WRR-013 live gateway probe (orchestrator skip; t672 C4b supplies the live acceptance-shape evidence).
- `prepareGatewayConversation` 65.1% and `gatewayConversationPassthrough` 100% figures — cited from §E.2, not isolated in this audit's per-function re-measure (the three reroot-specific figures were re-measured and matched).

## Residual risk

- The wedge→re-root→retry path has never run against a live gateway end-to-end (AC-WRR-013); mitigated by t672's live C4b cell plus the unit locks, but a live-probe follow-up remains the closure for that residual.
- CI full-suite on `origin/develop` is pending at audit time; a red there supersedes this verdict's Craft basis.

## Score derivation

Harmonic mean: 4 / (1/98 + 1/97 + 1/90 + 1/95) = **94.9**. (Weighted arithmetic equivalent: 95.7.) The harmonic form is used per the skeptical-evaluation stance; both forms agree on PASS with must-pass dimensions comfortably above threshold.
