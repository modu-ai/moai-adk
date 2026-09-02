# Sync-Audit Review 1 — SPEC-V3R6-GRAPH-FRESHNESS-001 (card t250)

- Auditor: t250-syncaudit (sync-auditor, opus/high), read-only, tree @ 2fc4b40a6 (base baa100ce5, 11 commits)
- Verdict: **PASS** — harmonic mean 85.2; must-pass Functionality + Security both pass
- Scores: Functionality 92 / Security 78 / Craft 85 / Consistency 87

## Independently measured (this run, this tree)

- MUTANT A reproduced through the real binary: /tmp fixture, `go run ./cmd/moai graph check` → exit 1 observed (matches §E.2 verbatim)
- MUTANT B: TestRefreshIndex_ReflectsUncommittedEdits catches the stamp-only variant precisely
- Tests green (graph/symbol/astx/mx-refresh/gate/cli-subset); vet rc=0; coverage graph 86.1% / symbol 88.2% / astx 88.5%
- Real edges.jsonl grep: `disagrees_with` count exactly 1 (the documented lsp→astgrep) — 1,505→1 claim measured-consistent; `--all-disagreements` revival wired
- Additivity trap: relation-key preservation asserted (kind+source+target+line) — not mere co-existence
- mtime ban / template neutrality / go.mod unchanged / day-one stamp / 4-locale parity (7 sections each, new section :55) all verified

## Findings

| # | Severity | Finding | Disposition |
|---|----------|---------|-------------|
| F1 | MUST-FIX Medium | `internal/graph/codequery.go:37` — `filepath.Join` does not block `..`; PoC Join("/root/project","../../../etc/secrets/creds.go") → escape. LLM-facing MCP param = trust boundary. Fix ~3 lines (filepath.Rel containment rejection). Below Security-FAIL bar (local stdio, signatures-only) — verdict stands. | Lane verified firsthand; fix delegated (containment check + regression test) |
| F2 | MUST-FIX | gofmt violations 2 files (check.go:33 const alignment, mcp_code_tools.go:46-67 map alignment); nothing catches them (no gofmt in .golangci.yml/CI) | fix delegated; repo-wide gofmt CI = follow-up candidate |
| F3 | MUST-FIX doc | progress.md:178 + m5-baseline.md:3 claim "git history proves ordering" — FALSE (baseline landed in SAME commit 7f2e9e77d as first M5 impl; same-commit = unprovable). Gap itself honest. | fix delegated (reword to authoring-session basis) |
| F4 | NICE | threshold dual source: graph/check.go:51 literals vs config/defaults.go constants — future divergence risk | follow-up |
| F5 | OBS | CI red conclusion unobserved locally (push-forbidden) — honestly recorded; workflow YAML valid (bootstrap order, triggers, contents:read). Integration PR CI exercises it. | accepted |
| F6 | OBS | worktree check honestly stale now (mx 9 files, edges specs moved — M4/M5 commits newer than artifacts). Gate working as designed. | accepted |
| F7 | OBS | gate nil-config silence deviation — honestly recorded in §E.2 | accepted |

Settled items (asymmetric rule / baseline gap / threshold 40) not relitigated — honesty verified, all compliant. Codex/GLM not invoked (audit_model unset → claude anchor solo, expected path).

---
Preserved from the auditor's delivered verdict by the lane. Fix delegation 2026-08-25 (d773b203).
