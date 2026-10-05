---
id: SPEC-CLI-TEST-TIMEOUT-001
title: "Implementation Plan — Explicit go test timeouts on local entry points"
version: "0.1.0"
created: 2026-09-26
updated: 2026-09-26
---

# plan.md — SPEC-CLI-TEST-TIMEOUT-001

## §A Context

- Card t1253, Tier M, no Go code change. Three files: `Makefile`, `CLAUDE.local.md`, and `scripts/ci-mirror/lib/go.sh`.
- Problem: `go test`'s 10-minute default per-binary timeout is exceeded by `./internal/cli/`
  on loaded local machines (measured 1118.093s no-race, exit 0; incidents t1171 601s panic,
  t1232 971s/1597s). CI already carries explicit timeouts; local entry points do not.
- Baseline attribution: `.moai/reports/t1253/measure-meta.txt` (HEAD `b4f798dcc`,
  go1.26.8 darwin/arm64, env-scrubbed, slot-serialized). Slowest-test distribution:
  `.moai/reports/t1253/aggregate-top25.txt`.

## §B Known Issues

- The 60m value rests on a projection (CI race × measured local amplification), not on a
  direct local race measurement; headroom is 1.12x by construction (spec.md §B.1 D1).
  REQ-TIMEOUT-006 owns the re-derivation trigger.
- `test-codex-live`'s Live subset has no duration measurement (disclosed gap, D3); the 10m
  value is an explicit pin, not a derived ceiling.
- The `go-test-heavy` slot lease (`moai slot`) bounds when the M3 re-verification may run;
  the run takes ~19 min wall and must not run concurrently with another lane's heavy test.

## §C Pre-flight

1. Confirm worktree branch is `WT-cli-test-duration` and HEAD is the expected base before the first edit (re-read `git rev-parse --short HEAD` immediately before any commit — the orchestrator owns commits; this plan performs none).
2. Confirm `internal/cli/main_test.go` state is unchanged from base (t1252 is file-disjoint; any surprise overlap is a stop-and-report).
3. Capture RED-now evidence for AC-001/AC-002: run the §D grep on the unmodified tree and record the zero-`-timeout` result in `progress.md` §E.1 evidence paths.

## §D Constraints

- Writes only under `.moai/specs/SPEC-CLI-TEST-TIMEOUT-001/` (artifacts) plus the three target files `Makefile`, `CLAUDE.local.md`, and `scripts/ci-mirror/lib/go.sh`.
- No commit, no branch, no worktree creation, no GitHub issue (late-branch opt-out default) — the orchestrator handles git.
- `.github/workflows/**` untouched (REQ-SCOPE-008).
- CLAUDE.local.md §6 edit is limited to the package-test recipe line(s); the full-suite lines are t1219's surface (REQ-COORD-009 / spec.md §D).
- Makefile edits must preserve the `##` help-text convention (`make help` extraction).
- Makefile `-timeout` flags get a trailing comment naming the derivation (D1/D3) and `.moai/reports/t1253/measure-meta.txt` (REQ-DOC-007); CLAUDE.local.md recipe lines get the same pointer inline.

## §E Self-Verification

- E1: AC matrix PASS/FAIL — run every AC in `acceptance.md` §D, quote verbatim output.
- E2: `git diff --stat` shows exactly `Makefile` + `scripts/ci-mirror/lib/go.sh` + `CLAUDE.local.md` + `.moai/specs/SPEC-CLI-TEST-TIMEOUT-001/**` and nothing else.
- E3: `make test-race-short` (the cheapest changed target) executes successfully with the explicit flag visible in dry-run (`make -n test-race-short` shows `-timeout 60m`).
- E4: M3 bounded re-verification completes exit 0, no `panic: test timed out` in the stream.
- E5: `make help` still lists the four targets (help-extraction not broken).

## §F Milestones (priority-ordered; decision-reversibility ordering — the timeout values are the highest-change-likelihood decisions and land first)

### M1 — Makefile + ci-mirror explicit timeouts (Priority High)

Add explicit `-timeout` to the five COVERED surfaces, each with its derivation comment (D1/D3):

- `test`: `go test -race -coverprofile=coverage.out -covermode=atomic -timeout 60m ./...`
- `test-verbose`: `go test -race -v -coverprofile=coverage.out -covermode=atomic -timeout 60m ./...`
- `test-race-short`: `go test -race -short -timeout 60m ./...`
- `test-codex-live`: `... go test -timeout 10m ./internal/cli/ -run 'Live' -v -count=1`
- `scripts/ci-mirror/lib/go.sh` test step: `go test -race -count=1 -short -timeout 60m ./...` (added at plan-audit iter-1; same D1 derivation and `-short` disclosure as `test-race-short`)

The EXCLUDED rows of the §B.3 inventory receive no edits — their rationale lives in the table, not in the touched files.

Accepts: AC-001, AC-003 (grep + derivation-comment presence), AC-009 (inventory cross-check basis).

### M2 — CLAUDE.local.md recipe edits (Priority High)

- §4 Before-Commit checklist line: `go test ./internal/<pkg>/...` → `go test -timeout 30m ./internal/<pkg>/...` (with D2 pointer).
- §6 [HARD] rule line (line "run the AFFECTED packages"): same substitution. Do NOT touch the §6 full-suite lines (`-count=1 ./...`, `-race ./...`) — t1219's surface.
- Do NOT touch the §13 full-suite mention (line ~532, "Unit tests: dev project (`go test ./...`)") either — context prose inside the GLM-testing ban, outside the sanctioned definition; wording ownership assigned to t1219 (spec.md §D, added at plan-audit iter-1).

Accepts: AC-002, AC-006 (coordination note verifiable in spec.md §D and the edited lines).

### M3 — Bounded post-change re-verification (Priority Medium)

Env-scrubbed, slot-serialized, single compound invocation:

```
unset <factory,kanban,GIT_* env> && go test -json -count=1 -timeout 35m ./internal/cli/
```

Expected: exit 0, no `panic: test timed out`, all leaf tests green. Persist verbatim output
tail + exit code under `.moai/state/verify/` per the evidence-persistence contract and record
the path in `progress.md` §E.2-ready form. Accepts: AC-004. (35m chosen for the probe itself:
30m recipe value plus probe headroom — the probe validates the recipe class, and its own
timeout must exceed the recipe value it validates.)

### M4 — Final AC sweep and artifact hygiene (Priority Low)

- AC-005 (workflows untouched), AC-007 (rejections recorded), AC-008 (no bare constant) verified; §E self-verification batch executed; `progress.md` §E.1 signal finalized.

## §G Anti-Patterns

- Do NOT "fix" by deleting slow tests or adding `t.Parallel` — out of scope (spec.md §C.2).
- Do NOT touch CI workflows even though their values (20m/25m) differ from the local values (60m) — different load regimes, different derivations (REQ-SCOPE-008).
- Do NOT raise a timeout value without re-deriving it (REQ-TIMEOUT-006).
- Do NOT put the derivation only in the SPEC and leave a bare constant in the file (REQ-DOC-007) — the mutant "comment removed, flag kept" fails AC-008.
- Do NOT edit CLAUDE.local.md §6 full-suite lines — t1219 owns them (merge-order hazard).

## §H Cross-References

- spec.md §A.1/§A.2 — measured baseline and incident record.
- `.moai/reports/t1253/measure-meta.txt`, `.moai/reports/t1253/aggregate-top25.txt` — evidence.
- `.moai/reports/t1253/go-test-cli.json` — raw 8.15MB `go test -json` stream (jq on demand; do not read wholesale).
- GitHub Actions run 36228023389 — CI comparator artifacts.
- SPEC-HEAVY-TEST-SLOT-001 — the slot-serialization mechanism the M3 probe uses.
- Cards: t1252 (TestMain fix, file-disjoint), t1219 (CLAUDE.local.md §6 full-suite lines, queued).
