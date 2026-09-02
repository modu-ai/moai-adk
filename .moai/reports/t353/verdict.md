# t353 verdict — REQ-MRG-010 resumed: the R4-form lint exclusion, delivered and lane-verified

Card: t353 (G2a, lane-9) · Branch: WT-r4-imperative-exempt · Base: develop `9b89b8c5b` · Implementation commit: `0026dc7b9` (by the spawned manager-develop; lane-verified below)

## Claim

The deferral SPEC-MOVING-REF-GUARD-001 left behind (REQ-MRG-010 + AC-MRG-013 + CM-1/CM-2, §H Q0 option C) is delivered: `MovingRefUnpinned` now exempts a flagged-shape line IFF it carries BOTH an imperative measuring directive introducing a backticked command (run/measure/re-measure/측정/재측정 + backtick) AND a demoted dated reference (parenthetical with a calendar date + a demotion label) — imperative STRUCTURE, never a command token, per the [HARD] the deferral fixed. The resume condition was re-measured MET before implementation (live external R4-form occupants silenced via R3 markers: SPEC-IGNORED-EVIDENCE-CITATION-001/progress.md:480 self-declared R4; SPEC-SPECLINT-GITBLIND-001/progress.md:233 Korean-form).

## Evidence — lane's independent verification (executor claims re-executed, not accepted)

1. Commit scope: `git show --stat 0026dc7b9` → lint_movingref.go, lint_movingref_test.go, 4 testdata fixtures, impl-record.md — no SPEC artifact, no doctrine file.
2. Suite re-run by the lane: `go test ./internal/spec/ -run 'MovingRef' -count=1` → `ok ... 5.901s` (this run, this tree).
3. Corpus delta claim spot-checked: the two newly-exempted lines named (acceptance.md:357 = AC-MRG-013's own retained fixture; progress.md:158) — the :357 line was read directly in this session before implementation and carries the R4 structure (imperative + backticked command + dated parenthetical); count 115→113 recorded in impl-record.md (built-binary measurement by the executor, not re-built by the lane — see Gaps).
4. Working tree clean after the run — no unstaged residue.

## Findings on the executor's report (one defect withdrawn after lead correction, one attribution that stands)

1. **"Wrong test name" — RETRACTED.** The original finding claimed `TestCatalogHashParity` "does not exist"; that claim was an UNSCOPED absence assertion: the lane probed only `./internal/template/`, where the name is indeed absent, and overreached from that to non-existence. The lead's correction (measured on the develop tree) locates it at `internal/spec/catalog_hash_test.go:108` — a DIFFERENT package — and the lane has since re-run `go test ./internal/spec/ -run 'TestCatalogHashParity'` in a develop-tip worktree and observed it FAIL (on entry `moai-ref-seo`; the drift set differs per tip). Both guards named by the executor and the lead are real and both fail on stale catalog entries: `TestCatalogHashParity` (internal/spec) and `TestManifestHashFormat` (internal/template, `CATALOG_HASH_UNSTABLE: sync-auditor stored f1b4487f… ≠ computed 545d03d9…`, observed this run). Lesson recorded per verification-claim-integrity §1.1 surface 3: an absence claim names its sweep scope or it is a hypothesis.
2. **Attribution verified, repair NOT done** (per the standing lead instruction on pre-existing red): `git diff 9b89b8c5b..HEAD --stat -- internal/template/catalog.yaml .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md` → EMPTY (both sides byte-identical to base), so the drift exists identically at the develop tip this branch absorbed. The lead's own attribution (its `4244c4a06` batch refreshed `sync-auditor.md` without the two derived artifacts) owns the root cause; repair (`gen-catalog-hashes --entry sync-auditor` + `make agents-emit`) is queued with the operator, so this card repairs nothing.

## Baseline-attribution

All lane measurements this run, this tree (WT-r4-imperative-exempt @ 0026dc7b9):
- suite: `go test ./internal/spec/ -run 'MovingRef' -count=1` → ok (this run).
- selector probe: `-run 'TestCatalogHashParity'` → no tests to run (this run).
- manifest failure: `-run 'TestManifestHashFormat' -v` → FAIL with CATALOG_HASH_UNSTABLE (this run).
- base-attribution: empty diff --stat over the two hash-comparison inputs vs `9b89b8c5b` (this run).
- resume-condition measurement (pre-implementation, same session): greps over `.moai/specs/` for R4-form structures + applied `<!-- moving-ref-ok:` markers (17 name-mentions; ≥5 applied externally, ≥2 genuinely R4-form).

## Gaps (explicitly NOT observed)

- The 115→113 corpus numbers were produced by the executor's `/tmp` binary builds; the lane verified the two named lines' R4 form by reading, not by re-running the built binary.
- golangci-lint was not run (outside the mission's verification list).
- The SPEC body still says DEFERRED — the un-defer edit is manager-spec's surface, dispatched separately from this verdict; until it lands, SPEC-MOVING-REF-GUARD-001's artifacts describe the pre-resume state.
- CI's full-suite verdict on the merged develop head belongs to the lead's batch push.

## Residual-risk

- The structural key is bilingual by enumeration (EN imperatives + 측정/재측정 + EN/KR demotion labels); a future line in a THIRD language meeting R4's structure would not match — the failure direction is under-exemption (line stays flagged, silenced manually via R3 marker), which is the safe direction, and the finding message still names R4.
- The exclusion's two-conjunct conjunction is narrower than every real R4 line observed (both conjuncts present in all measured occupants) — over-exemption exposure measured zero on this corpus.
