---
id: SPEC-CI-VERDICT-PRODUCER-001
title: "CI verdict producer — acceptance criteria"
version: "0.1.0"
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
tier: M
---

# acceptance.md — SPEC-CI-VERDICT-PRODUCER-001

All scenarios use fabricated inputs; no test touches the network. Fixtures build trees under `t.TempDir()`.

## §A — Producer

- **AC-CV-001** (maps REQ-CV-001, REQ-CV-002, REQ-CV-004) — **Given** a project tree under `t.TempDir()` and an offline input file recording `head_sha` = a fixture SHA, `conclusion` = `failure`, `run_id`, and `observed_at`, **When** the producer verb is invoked with that input file, **Then** exactly one file `<head-sha>.json` exists under `<tree>/.moai/state/ci-verdicts/` whose fields match the input schema byte-for-byte (`producer` carries the writing verb's identity); **And Given** the same invocation re-run with identical input, **Then** the file content is byte-identical (idempotent rewrite, no second file), and a re-run with a changed conclusion overwrites the record (last-writer-wins).
  - Test: producer tests in `internal/cli` (change-scoped).
- **AC-CV-002** (maps REQ-CV-001) — **Given** a fabricated `gh` output fixture delivered through the injected runner (no real gh, no network), **When** the producer verb is invoked for a head in fetch mode, **Then** the recorded file's `conclusion` and `run_id` equal the fabricated CI result for that head and `observed_at` reflects the observation time.
  - Test: producer tests in `internal/cli`.
- **AC-CV-003** (maps REQ-CV-003) — **Given** a runner fixture where `gh` is absent (LookPath failure), where it exits non-zero, and where it prints unparseable output — three separate fixtures — **When** the producer verb runs in fetch mode in each, **Then** each prints one clear message naming the fault, exits 0, and writes no file under `.moai/state/ci-verdicts/`.
  - Test: producer degradation tests in `internal/cli`.
- **AC-CV-004** (maps REQ-CV-005) — **Given** any producer invocation (fetch, offline, or degraded), **When** its call path is inspected, **Then** it contains no reference to the escalation checkpoint/hook entry points (grep discipline: no `Checkpoint` symbol reachable from the producer package path), and a record write alone causes no escalation record to appear in the tree.
  - Test: producer tests + boundary grep.

## §B — Detector CI limb

- **AC-CV-005** (maps REQ-CV-006, REQ-CV-007) — **Given** an armed card whose checkpoint head is H, a verify snapshot at H containing one check entry with exit code 0, and a CI verdict record pinning H with `conclusion: failure`, **When** a commit checkpoint runs, **Then** exactly one `contradictory-evidence` record is written whose observation names the local pass and the CI failure at H, written through the existing `contradiction(...)` fingerprint path.
  - Test: extended `TestContradictoryEvidenceTrips`, limb (c) same-head case, `internal/escalation/operational_m5_test.go`.
- **AC-CV-006** (maps REQ-CV-008) — **Given** the same fixture shape with, in separate runs, (a) the CI record pinning a different head H′, (b) the CI record at H with `conclusion: success`, (c) no local-pass snapshot at H (CI failure at H), (d) no CI record at all, (e) the CI record at H with `conclusion: neutral`, and (f) the CI record at H with `conclusion: success` and no local-pass snapshot at H — **When** a commit checkpoint runs in each, **Then** (a), (b), (c), (e), and (f) each write no record, with (a) listing the foreign-head CI verdict under `not_observed`, (c) and (f) each listing the missing local pass under `not_observed`, and (b) and (e) NOT listing any ci-verdict entry under `not_observed` (a success or neutral conclusion at the head completes the observation per REQ-CV-008); and **Then** (d) preserves the existing limb-(e) behavior byte-for-byte: no record and the ci-verdict entry still listed under `not_observed`, never as agreement.
  - Test: extended `TestContradictoryEvidenceTrips` (different-head, success, no-local-pass, no-record cases).
- **AC-CV-007** (maps REQ-CV-009, REQ-CV-004) — **Given** the AC-CV-005 trip already fired and its record resolved, **When** the producer re-records the byte-identical verdict for H and a further commit checkpoint runs, **Then** no duplicate `contradictory-evidence` record is written (the `freshEvidence` content-hash gate treats identical record bytes as consumed); a changed CI record (new run_id or flipped conclusion and back) re-trips once.
  - Test: extended `TestContradictoryEvidenceTrips` freshness case.

## §C — Re-judgement and constraints

- **AC-CV-008** (maps REQ-CV-007) — **Given** the completed implementation on this branch, **When** the re-judgement of AC-AE-012(c) of SPEC-AUTONOMY-ESCALATION-001 is run, **Then** the mechanical form is the extended `TestContradictoryEvidenceTrips` limb-(c) case passing (`go test ./internal/escalation/... -run TestContradictoryEvidenceTrips` → ok) — the criterion's own wording ("a local verification pass and a recorded CI failure on the same head" trips; (e) stays not-observed) is now exercised, not assumed; **And** `git diff --name-only "$(git merge-base develop HEAD)"..HEAD -- go.mod` is empty (no new dependency), and `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
  - Test: the run-phase §E evidence carries the verbatim command outputs.

## §D — Traceability

| AC | Requirements | Test |
|----|--------------|------|
| AC-CV-001 | REQ-CV-001, REQ-CV-002, REQ-CV-004 | `internal/cli` producer tests |
| AC-CV-002 | REQ-CV-001 | `internal/cli` producer fetch test |
| AC-CV-003 | REQ-CV-003 | `internal/cli` degradation tests |
| AC-CV-004 | REQ-CV-005 | producer boundary grep + test |
| AC-CV-005 | REQ-CV-006, REQ-CV-007 | `TestContradictoryEvidenceTrips` limb (c) same-head |
| AC-CV-006 | REQ-CV-008 | `TestContradictoryEvidenceTrips` non-trip cases |
| AC-CV-007 | REQ-CV-009, REQ-CV-004 | `TestContradictoryEvidenceTrips` freshness case |
| AC-CV-008 | REQ-CV-007 | §E evidence + cross-platform build |

## §E — Quality gates

- Coverage ≥ 85% for touched packages (quality.yaml `test_coverage_target`).
- No new `go.mod` dependency; no new syscall surface; Windows cross-build green.
- Definition of Done: all AC PASS with verbatim evidence; SPEC-AUTONOMY-ESCALATION-001 untouched (`git diff --name-only "$(git merge-base develop HEAD)"..HEAD -- .moai/specs/SPEC-AUTONOMY-ESCALATION-001/` empty); AC-AE-012(c) re-judged via AC-CV-008.
