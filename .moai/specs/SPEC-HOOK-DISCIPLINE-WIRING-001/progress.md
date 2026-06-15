# SPEC-HOOK-DISCIPLINE-WIRING-001 — Progress

## §0 Plan Audit Gate (Phase 0.5)

| Field | Value |
|-------|-------|
| Auditor | plan-auditor (iter-1) |
| Verdict | PASS (PASS-WITH-DEBT) |
| Score | 0.83 (Tier M threshold 0.80; not skip-eligible <0.90) |
| MUST-PASS | MP-1..MP-4 all PASS |
| Dimensions | Clarity 0.90 / Completeness 0.92 / Testability 0.78 / Traceability 1.0 |

Carried debt (documented):
- **D1 (SHOULD-FIX)** — AC-HDW-009 `total == inblk` equality can false-fail when language-agnostic reporting strings (`"check":`, `BLOCKED_REASON=`, `unavailable`) stay in the common path. The neutrality property is independently + robustly guarded by AC-HDW-002's runtime fixture, so run-phase verifies "Go invocations confined to the `go)` branch" by **invocation-site inspection**, not by the raw `total==inblk` count. (Sanctioned by plan-auditor.)
- **D2 (MINOR)** — exit-2-deferral: commit to design §1.6 **option 1 (BLOCKING env-gate, dormant `exit 2` path)** at M1; AC-HDW-008 treats a dormant `exit 2` presence as acceptable.
- **D3 (MINOR)** — resolved: `tier: M` added to spec.md frontmatter (orchestrator-direct, prevents backward-compat Tier-L 0.85 misclassification).

## §E — Phase 0.95 Mode Selection

Input parameters:
- tier: M
- scope (file count): ~5 (settings.json.tmpl + local .claude/settings.json mirror + sync-phase-quality-gate.sh local+template mirror + embedded.go regen via make build)
- domain count: 1 (hooks/settings configuration + shell-script language generalization)
- file language mix: JSON (settings) + shell (hook script) + Go (embedded.go generated)
- concurrency benefit: LOW (single coherent coding unit, sequential)
- Agent Teams prereqs status: not evaluated (single-domain, below multi-domain threshold)

Mode evaluation:
| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | multi-file + semantic (language generalization) |
| 2 background | no | involves Write/Edit (not read-only) |
| 3 agent-team | no | single domain, <3 domains, prereqs not met |
| 4 parallel | no | not multi-domain research-heavy |
| 5 sub-agent | **yes** | coding-heavy single coherent unit; default fallback |
| 6 workflow | no | not ≥30-file mechanical transform |

Decision: sub-agent (Mode 5), cycle_type=tdd, single sequential manager-develop spawn.

Justification: Per Anthropic's coding-task parallelism caveat, coding-heavy single-domain work uses the sequential sub-agent path. Scope (~5 files, 1 domain) is well below Mode 3/4 multi-domain thresholds and the Mode 6 ≥30-file mechanical-transform threshold. Implementation Kickoff Approval was obtained from the user before this Phase 0.95 selection.

## §D — Run-phase Evidence

Run by manager-develop (cycle_type=tdd, Mode 5 sub-agent). Status transition: draft → in-progress (this run; updated 2026-06-15).

### Scope delivered

1. `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` — language-generalized (Go-only → marker-driven detect_language for Go/Node/Python/Rust). Go tooling confined to the `go)` case branch. Graceful `command -v` skip + silent-pass on no marker. Warn-first: dormant `exit 2` gated behind `MOAI_SYNC_GATE_BLOCKING=1` (OFF by default — D2 design §1.6 option 1).
2. `internal/template/templates/.claude/settings.json.tmpl` — status-transition-ownership.sh wired into PostToolUse (advisory, timeout 5), sync-phase-quality-gate.sh wired into Stop (warn-first, timeout 10). Platform-conditional quoting via `$CLAUDE_PROJECT_DIR`. team-ac-verify.sh NOT registered.
3. `.claude/hooks/moai/sync-phase-quality-gate.sh` — manual mirror (byte-identical to template).
4. `.claude/settings.json` — local mirror of both wiring entries; dev-intent keys (defaultMode/env.PATH/enableAllProjectMcpServers) untouched.

Note on embedding: this project uses `//go:embed all:templates` (compile-time, `internal/template/embed.go`). There is NO generated `embedded.go` — `make build` recompiles the binary which embeds the updated templates directly. `moai init` smoke confirmed the deployed sync-gate is the 9199-byte generalized version with `detect_language`.

In-run fix (Rule 4 reproduction-first): AC-HDW-002(b) single-commit fixture exposed two latent bugs in the generalized hook — (a) `git diff HEAD~1..HEAD` on an initial commit (no HEAD~1) caused a `grep -c || echo "0"` double-emit → fixed with empty-tree-SHA fallback + `${VAR:-0}` normalization; (b) `run_step`'s `( "$@"; echo $? )` subshell aborted under `set -e` when a tool legitimately failed → fixed with the `&& rc=0 || rc=$?` capture idiom. Both fixes mirrored to template + local, parity re-verified.

### E1 — AC PASS/FAIL matrix (AC-HDW-001..009)

| AC | Status | Verification command | Actual output |
|----|--------|----------------------|---------------|
| AC-HDW-001 | PASS | `grep -c 'status-transition-ownership.sh' <tmpl>` + block placement grep | count 2 (win+unix render); near-CLAUDE_PROJECT_DIR 2; line 82-84 between PostToolUse (67) and Stop (93) |
| AC-HDW-002 | PASS | (a) `source <hook>; detect_language` per marker; (b) real-git node fixture; (c) real-git no-marker | (a) go/node/python/rust/empty all correct; (b) exit 0, language node, `"go vet"` count 0; (c) exit 0 decision skip "no recognized language marker"; static: `command -v` count 3, non-go markers 6 |
| AC-HDW-003 | PASS | `grep -c sync-phase-quality-gate.sh` + Stop-block awk | count 2; in Stop block 2; handle-stop coexist 2; harness-observe coexist 2 |
| AC-HDW-004 | PASS | `grep -c team-ac-verify` + `test -f` both copies | grep 0 (both files); both `test -f` present |
| AC-HDW-005 | PASS | handler coexistence greps + `moai init` smoke | handle-post-tool/handle-stop/handle-task-completed each 2; `moai init --non-interactive` → valid JSON, all 3 hook scripts deployed |
| AC-HDW-006 | PASS | `go test ./internal/template/ -run 'TestTemplateNeutralityAudit\|TestTemplateNoInternalContentLeak' -v` (bare pipe) + targeted internal-token scan | both `--- PASS`, no "no tests to run"; targeted grep exit 1 (no match) |
| AC-HDW-007 | PASS | `diff -q` local vs template sync-gate + local settings grep + dev-intent git-diff | diff:0 (byte-identical via manual mirror); local settings hooks grep ≥2; dev-intent keys absent from diff |
| AC-HDW-008 | PASS | status-transition Edit payload exit + dormant exit-2 + passing-scenario sync-gate | status-transition exit 0; `exit 2` gated behind `MOAI_SYNC_GATE_BLOCKING=1` (OFF default); passing go fixture exit 0 decision allow; spec.md §Exclusions item 1 documents deferral |
| AC-HDW-009 | PASS (invocation-site per D1) | `awk` go-block extract + word-boundary grep | raw total/inblk 6/5 (D1 false-positive: `go test ` substring-matches `cargo test ` at rust branch); **word-boundary total/inblk 5/5** (zero Go-tool leak); 8 markers equal-weight, 0 PRIMARY designation |

### E2 — Cross-platform build

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

### E3 — Neutrality tests (bare pipe, GREEN)

```
$ go test ./internal/template/... -run 'TestTemplateNoInternalContentLeak|TestTemplateNeutralityAudit' -v
--- PASS: TestTemplateNoInternalContentLeak
--- PASS: TestTemplateNeutralityAuditC8Preserve
--- PASS: TestTemplateNeutralityAudit
ok  github.com/modu-ai/moai-adk/internal/template
```

### E4 — Language-neutrality proof

Go-tool tokens (`go vet`/`golangci-lint`/`go test`) appear ONLY inside the `go)` toolchain case branch (lines 144-149) in both template and local copies. Runtime: real-git node fixture (package.json marker, app.js delta) → hook detects `node`, runs eslint/npm test path, never invokes `go vet`/`go test`, exits 0.

### E5 — Lint status (no NEW issues)

```
$ golangci-lint run --timeout=2m
0 issues.   (== baseline 0)
```

### E6 — local↔template parity

```
$ diff -q .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh
(identical — diff:0)
```
settings.json hooks-block entries mirror settings.json.tmpl (modulo template-var rendering). Binary recompiled via `make build` (embeds templates at compile time; no embedded.go file exists).

### E7 — Full regression

```
$ go test ./...   → exit 0, zero FAIL lines
```

### E8 — Commit + push

Run-phase commits (Hybrid Trunk Tier M, main 직진):
- `6028c4419` — M1 language-generalize sync-phase-quality-gate hook (both copies)
- `4a3f41dec` — M2 wire discipline hooks into settings + status draft→in-progress + progress.md §D

Both carry `Authored-By-Agent: manager-develop` + `🗿 MoAI` trailers. Push to origin/main: see final agent report (pushed after this progress.md amendment commit).
