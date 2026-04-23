# MoAI-ADK v3 — Utility Layer Review Synthesis

> Scope: @MX TAG system + ast-grep integration + LSP subsystem
> Based on: 4 audit reports (~20K words), 2025-2026 research corpus
> Date: 2026-04-24
> Authors: synthesis-strategist (over A1/A2/A3/A4 audits)

---

## Executive Summary

### Aggregate health

| Layer | LOC | Health | Verdict | Critical count |
|-------|-----|--------|---------|----------------|
| MX (internal/hook/mx) | 2,264 | 5.5/10 | REFACTOR | 1 (B1 method blindspot) + 4 high |
| ast-grep (internal/astgrep + hook/{quality,security}) | 6,261 | 6/10 | REFACTOR | 0 critical, 5 high |
| LSP (internal/lsp/*, 9 sub-packages) | 18,718 | 7/10 | MAINTAIN+targeted | 3 (stderr leak, spawn race, type fragmentation) |
| **Aggregate utility layer** | **27,243** | **~6/10** | **REFACTOR with safe pocket (LSP core)** | **8 concrete Critical bugs** |

The utility layer is functional but below the production bar for a v3 release. The weakest link is the MX validator: its core analysis engine (`extractFunctions` + `countFanIn`) has four concrete bugs that silently drop signal on the majority of idiomatic Go code (method receivers). The ast-grep layer has sound ingress but loses OWASP/CWE metadata between parse and emit, and both it and the MX validator spawn unbounded goroutines. The LSP `core/` path is the strongest of the three — race-clean, powernap-backed, 92.9% coverage — but the `gopls/` bridge is a 632-LOC parallel JSON-RPC universe that must be soft-deprecated, and three type duplications (Diagnostic × 3, Phase × 2, extension-map × 3) create integration friction across the codebase.

All three utilities share the same two latent risks: **(a) unbounded goroutine fan-out** under load and **(b) subprocess lifecycle gaps** that the 2026 Go idiom corpus (A4 T5) considers table-stakes — `cmd.WaitDelay`, `Setpgid`, stderr drain. Fixing these three cross-cutting themes unlocks production-grade reliability without a full rewrite.

### Top 5 cross-cutting findings

1. **Unbounded goroutine fan-out** appears in all three audits: `ValidateFiles` (A1 P1), `ScanMultiple` (A2 5-3), and the implicit race in `getOrSpawn` (A3 D10-1). A single worker-pool pattern (`sem := make(chan struct{}, runtime.NumCPU()*2)`) resolves the first two; `singleflight.Group` resolves the third.
2. **Type fragmentation across package boundaries**: `Diagnostic` exists in three variants (A3 #4), `Phase` in two (A3 #5), extension-to-language maps in three (A3 D3-1, A3 D6-2, A2 7-3). Each duplication is a drift vector — when `lsp.yaml` adds a language, the audit confirms at least one consumer will miss it.
3. **Subprocess lifecycle discipline is missing**, validated externally by A4 T5: no `cmd.WaitDelay`, no `Setpgid`, stderr pipe left unread (A3 D2-3, Critical). The external corpus treats the full `exec.CommandContext` + `WaitDelay` + `Setpgid` + stderr-drain pattern as 2026 baseline — moai currently ships none of it.
4. **Version pinning / detection gaps**: `detectSGVersion()` always returns `"unknown"` (A2 REC-03), no minimum ast-grep version declared, SARIF `tool.driver.version` permanently empty. powernap is pinned (v0.1.4) but has no upgrade cadence documentation in the audit scope.
5. **16-language neutrality is aspirational**, not implemented: 11 of 16 ast-grep rule directories contain only `.gitkeep` (A2 7-1); fallback diagnostics cover only 5 languages (A3 D3); security rules are Go-only. The template-language-neutrality HARD rule in CLAUDE.local.md §15 is violated in practice.

### Verdict

**REFACTOR with targeted surgery**. No full rewrites. Estimated Tier-1 fix set (7 items) is achievable in the Phase 6 window; Tier-2 architectural unification (15 items) carries v3.0 GA. Tier-3 (10 items) grounds v3.1 on 2025-2026 research corpus.

---

## 1. MX System Findings (A1)

Source: A1 — `internal/hook/mx/` (types.go, config.go, validator.go) + integration tests. 2,264 LOC, 92.6% test coverage (exceeds 85% target), race-clean.

### 1.1 Critical bugs (concrete)

- **B1 (Critical) — Method receiver blindspot** (A1 §D1, validator.go:19). `exportedFuncRe = ^func\s+([A-Z]\w+)` never matches `func (r *T) Method()`. In idiomatic Go, most exported API is method receivers; P2/P3/P4 violations inside receiver methods are invisible to the validator. This is the single largest correctness failure in the utility layer.
- **B2 (High) — `countFanIn` substring false positives** (A1 §D1, validator.go:264,284). `strings.Count(data, funcName)` and `grep -l funcName` both match substrings: `funcName="New"` counts `NewContext`, `Renew`, `RenewToken`. For short names (`New`, `Get`, `Set`, `Parse`) the inflation regularly exceeds the fan_in ≥ 3 ANCHOR threshold, producing spurious P1 violations.
- **B3 (High) — Brace-in-string-literal** (A1 §D1, validator.go:224-230). The depth counter iterates over every rune including those inside string literals. `s := "{"` increments depth, offsetting function end detection and producing false P3 (>=100 lines) violations.
- **B4 (Medium) — Blank-line gap between tag and function** (A1 §D1, validator.go:194-210). A single blank line between `@MX:WARN` and the function declaration causes the tag to be missed during backward scan.
- **B5 (Medium) — Goroutine pattern in comments/strings** (A1 §D1, validator.go:233). `strings.Contains(bodyLine, "\tgo ")` matches `// \tgo func` in comments and `"\tgo "` inside string literals.

### 1.2 Architectural issues

- **A1 (High) — `analyzer any` parameter is dead weight** (A1 §D2, validator.go:40). `NewValidator(analyzer any, projectRoot string)` takes `any` that is permanently `nil` in all call sites. The placeholder for future AST-grep integration leaks an unimplemented extension point into the public API.
- **A2 (Medium) — No `FanInCounter` interface** (A1 §D2, validator.go:258). `countFanIn` directly `exec.CommandContext("grep", ...)` — no mock injection point for unit testing, and Windows has no `grep` binary.
- **P1/P2 (High) — Unbounded goroutine × per-function grep fan-out** (A1 §D3). `ValidateFiles` spawns one goroutine per file, each of which may spawn N `grep` subprocesses (one per exported function without ANCHOR). For a 500-file / 10-func-per-file project this produces 5,000 grep subprocesses with no semaphore. Current complexity `O(F × G × P)` vs target `O(F × 1)` via pre-built inverted index.
- **M1 (High) — `extractFunctions` is a 74-LOC God function** (A1 §D7) with cyclomatic complexity ~12, mixing: scanning function declarations, backward MX tag scan, brace counting, goroutine detection.

### 1.3 Feature gaps vs MX protocol

Source `.claude/rules/moai/workflow/mx-tag-protocol.md`:

- **F1 Critical** — method receiver detection (protocol-mandated, not implemented; see B1).
- **F2 High** — `@MX:REASON` enforcement for WARN/ANCHOR (mandated by protocol, never checked).
- **F3 High** — Per-file limits (`anchor_per_file: 3`, `warn_per_file: 5` in `mx.yaml`) never enforced. Demotion logic from the reference doc (`mx-tag.md`) is undefined behavior.
- **F4 High** — Cyclomatic complexity check for P2 (protocol: WARN for complexity >= 15, if_branches >= 8) entirely absent. Only goroutine patterns are detected.
- **F5–F9** — `@MX:LEGACY`, `@MX:SPEC`, `@MX:TEST`, `auto_tag` config flag, `code_comments` language — all missing or partial.

### 1.4 External research validation

- A4 T1.3 rec-2: promote `@MX:REASON` to first-class typed tag with controlled vocabulary (rationale/invariant/hazard/temporary/external-constraint). Supports F2 closure.
- A4 T1.3 rec-3: adopt `key=value` syntax (`@MX:ANCHOR fan_in=7 invariant="non-nil"`). Aligns with Rust attribute macros, Java annotations, tree-sitter-extractable.
- A4 T3.4: LSP 3.17 `callHierarchy/incomingCalls` is the authoritative fan-in source — replaces the fragile grep-substring heuristic entirely.
- A4 T1.2 (arXiv 2510.22956 "Tagging-Augmented Generation"): validates moai's inline-first `@MX` over JSON sidecars. A4 recommends generating sidecar as a derived index, not as primary storage.

---

## 2. ast-grep Findings (A2)

Source: A2 — `internal/astgrep/`, `internal/cli/astgrep.go`, `internal/hook/quality/astgrep_gate*.go`, `internal/hook/security/ast_grep.go`. 6,261 LOC. ast-grep 0.40.5 installed; no minimum pin. A4 T2.1 confirms ast-grep 0.42.1 is latest (2026-04-04).

### 2.1 Critical bugs

- **REC-01 (High) — `Rule.Metadata` / `Rule.Note` fields missing from struct** (A2 §D2-2, models.go:72). YAML security rules use `metadata: {owasp: "A03:2021", cwe: "CWE-89"}` and `note: "..."`; the `Rule` struct has no fields for either. `Finding` has them, but `scanWithRules` (scanner.go:361-374) never propagates Rule→Finding. SARIF output `ruleProperties` is always empty. All 5 security rules' OWASP/CWE classifications are silently dropped.
- **REC-02 (High) — `ScanMultiple` unbounded goroutines** (A2 §D5-3, ast_grep.go:186). 10,000 input files → 10,000 concurrent `sg` subprocesses. `@MX:WARN` tag documents the issue but no semaphore exists.
- **REC-03 (High) — `detectSGVersion()` is a placeholder** (A2 §D1-3, cli/astgrep.go:193). Always returns `"unknown"`. SARIF `tool.driver.version` never populated; no minimum-version validation; CI drift risk.
- **REC-04 (Medium) — Context cancel leak in analyzer.go** (A2 §D1-1, analyzer.go:242, 298). Loop body: `ctx, cancel := context.WithTimeout(ctx, ...); ...; cancel()` without `defer`. If `executor.Execute` panics before `cancel()`, the timeout leaks.

### 2.2 Architectural issues

- **JSON parser triplication** (A2 §D1-2). `parseSGFindings` (scanner.go), `parseSGOutput` (analyzer.go), `parseASTGrepJSON` (security/ast_grep.go) — three separate parsers for the same `[]sgMatch` structure. 0-indexed→1-indexed conversion duplicated at scanner.go:401 and analyzer.go:122.
- **Extension-to-language maps duplicated** (A2 §7-3). `analyzer.DetectLanguage` uses `foundation.DefaultRegistry`; `security/ast_grep.go` builds its own `extensionToLanguage` map (15 languages, missing Dart and R).
- **V1/V2 gate coexistence** (A2 §D4-1). `RunAstGrepGate` and `RunAstGrepGateV2` both live in the same file. V1 uses non-recursive `LoadFromDirectory`, V2 uses recursive `LoadFromDir`. Intended-behavior divergence for nested rule directories is undocumented.

### 2.3 Rule catalog gaps (16-lang coverage)

- **11 of 16 language rule directories contain only `.gitkeep`** (A2 §D7-1): cpp, csharp, elixir, flutter, java, javascript, kotlin, php, python, r, ruby, rust, scala, swift, typescript — 15 of which are dead.
- Security rules are Go-only (A2 §D3-1). OWASP coverage for other 15 supported languages is zero.
- `security/ast_grep.go:supportedLanguages` lists 15 languages, missing `flutter` and `r` (A2 §7-2) — silent discrepancy with sgconfig.yml declaration.
- No graceful-degradation message when a language has zero rules: "scan failed" vs "no rules" are indistinguishable to the user.

### 2.4 External research validation

- A4 T2.1 rec-3: adopt `no-suppress-all` built-in rule (ast-grep 0.42.x) to block agents from silently disabling scans.
- A4 T2.1 rec-4: parameterized utils (0.42.x) enable shared pattern fragments (`function-with-high-fan-in($N)`) without copy-paste.
- A4 T2.2: Semgrep license tightening (2024-12, Opengrep fork Jan 2025) and CodeQL repackage behind GHAS (April 2025) confirm ast-grep's Apache 2.0 stability as a strategic advantage — no action item, but de-risks the choice.
- A4 T2.3: GOEXPERIMENT=jsonv2 (Go 1.25) gives 5.6-12× unmarshal speedup — directly relevant to ast-grep JSON parsing hot path. Prepare now, default-on in Go 1.26.

---

## 3. LSP Findings (A3)

Source: A3 — `internal/lsp/*` (9 sub-packages, 73 source files, 18,718 LOC). `go test -race ./internal/lsp/...` passes clean. Coverage 80.6%–100% per package; only `gopls/` below the 85% threshold.

### 3.1 Critical bugs

- **D2-3 Critical — Stderr pipe never consumed** (A3, core/client.go:213-217). `readWriteCloser{r: result.Stdout, w: result.Stdin}` — `result.Stderr` is never read. When a subprocess writes more than 64KB to stderr (gopls verbose logging, common), the pipe buffer fills and the subprocess blocks on its next stderr write, deadlocking the entire LSP client. Fix is one goroutine: `go io.Copy(io.Discard, result.Stderr)`.
- **D10-1 / getOrSpawn Critical — Concurrent client spawn race** (A3, core/manager.go:332-363). Two goroutines race on the same language: both see an empty cache, both call `clientFactory`, the client is inserted into the cache in state `StateSpawning` *before* `Start()` returns. The second caller receives the not-yet-ready client. If it calls `OpenFile` before the first `Start()` completes, `c.tr` is nil → panic.
- **D1-3 + type fragmentation #4 Critical — Three parallel diagnostic types** (A3 Top-5 #4). `lsp.Diagnostic` (models), `hook.Diagnostic` (string severity for JSON), `gopls.Diagnostic` (in gopls/protocol.go). `internal/loop/` and `internal/ralph/` import `gopls.Diagnostic` directly — a full migration path is required before the gopls bridge can retire.

### 3.2 gopls bridge decision (parallel universe)

`gopls/bridge.go` is 632 LOC of hand-rolled Content-Length framing + PendingRegistry + NotificationDispatcher + inline circuit-breaker (A3 §D7). It is functionally equivalent to the `transport/` package (powernap-backed) that powers `core/`, but it shares zero code.

- **D7-1 High** — inline circuit breaker (`cbMu/cbFailures/cbOpenUntil`) duplicates `resilience.CircuitBreaker` — drift risk when the shared implementation evolves (#679 half-open retry improvements stay out of reach).
- **D7-2 Critical-semantic** — `LanguageID: "go"` hardcoded at gopls/bridge.go:259 — the bridge can never generalize to TypeScript/Python.
- **D7-3 High** — `cmd.Wait()` goroutine leak on double-close (bridge.go:477-479).
- **A3 R4 recommendation**: soft-deprecate with `@MX:WARN DEPRECATED` on `NewBridge`; migrate `ralph/` and `loop/` to `aggregator.GetDiagnostics`; flip `ResolveClientImpl()` default from `"gopls_bridge"` to `"powernap_core"` (A3 §D9 notes the current default contradicts lsp-client.md v1.0.0).

### 3.3 Type fragmentation

- **Diagnostic × 3** (A3 Top-5 #4). `lsp.Diagnostic` uses `int` severity; `hook.Diagnostic` uses `string` severity for JSON consumer compatibility. Any migration to `lsp.Diagnostic` in hook output is a breaking change for external hook consumers — needs a BC entry.
- **Phase × 2** (A3 Top-5 #5). `hook.PhaseType` (hook/types.go:66) and `core/quality.WorkflowPhase` (core/quality/trust.go:72) — identical string constants `"plan"`, `"run"`, `"sync"`. Propose a new `internal/lsp/phase` package.
- **Extension-map × 3** (A3 D3-1, D6-2). `aggregator/aggregator.go:244` reimplements `filepath.Ext`+switch in 46 lines; `core/manager.detectLanguage` reads from `config.ServerConfig.FileExtensions`; `security/ast_grep.go` has its own. Propose `config.LanguageForExt(ext string) string` as the single source.

### 3.4 External research validation (LSP 3.17, powernap, gopls MCP)

- A4 T3.1: powernap v0.1.4 verified as crush's production LSP layer (23K+ stars). Zero API/ABI changes since v0.1.3. No credible alternative in 2026 supersedes powernap for a multi-language Go LSP client. Keep.
- A4 T3.4: LSP 3.17 `callHierarchy/incomingCalls` is the authoritative source for `@MX:ANCHOR` auto-fan-in — eliminates MX's grep-substring heuristic entirely. `textDocument/inlayHint` lets non-moai editors render `@MX` tags without reading source. Pull diagnostics — **do not adopt yet** (gopls warns performance not on par with push).
- A4 T3.5 + T5: subprocess hygiene (WaitDelay + Setpgid + stderr-drain-goroutine) is 2026 baseline. A3 D4-1 (syscall.SIGKILL portability) and D2-3 (stderr) are both resolved by adopting the T5 idiom set.
- A4 T5 rec-10: `GOEXPERIMENT=jsonv2` drop-in for all JSON-RPC transports. No code changes; 5-12× unmarshal speedup. Default in Go 1.26.
- A4 T3.5 rec-6: gopls built-in MCP server (v0.20+, 12 tools incl. `go_diagnostics`, `go_rename_symbol`) justifies keeping gopls-bridge as an *opt-in* path for token-efficient agentic Go workflows — coexistence with powernap_core as default remains the right architecture.

---

## 4. Cross-Cutting Themes

### 4.1 Unbounded goroutine spawn (all 3 audits)

Every utility subsystem has at least one unbounded goroutine spawn site:

| Audit | Location | Fan-out |
|-------|----------|---------|
| A1 P1 | `internal/hook/mx/validator.go:317` `ValidateFiles` | 1 goroutine per file × N grep subprocesses per function |
| A2 5-3 | `internal/hook/security/ast_grep.go:186` `ScanMultiple` | 1 goroutine per file × 1 sg subprocess each |
| A3 D10-1 | `internal/lsp/core/manager.go:332` `getOrSpawn` | Race on cache insert — not truly fan-out but same root cause: no coordination primitive |

**Single fix pattern** for A1/A2: bounded worker pool with `sem := make(chan struct{}, runtime.NumCPU()*2)`. **Fix for A3**: `golang.org/x/sync/singleflight.Group` per-language. Both are minimal imports, no new dependencies (singleflight is in `go.sum` already via `aggregator`).

### 4.2 Type duplication (Diagnostic × 3, Phase × 2, Extension × 3, JSON parser × 3)

A3 flags this as the dominant architectural debt. Quantified:

- **Diagnostic × 3** (A3 Top-5 #4)
- **Phase × 2** (A3 Top-5 #5)
- **Extension-to-language × 3** (A3 D3-1 / D6-2 + A2 7-3)
- **JSON parser × 3** (A2 §D1-2)

Cost: every time a new language is added to `lsp.yaml`, at least one of the three extension maps drifts. Every time sg JSON format changes, three parsers need updating — confirmed likely by A4 T2.3 noting ast-grep 0.31+ already had a JSON format change. Fix: unify each family into a single canonical definition; see Tier-2 items IMP-V3U-008/009/010/011.

### 4.3 Missing subprocess lifecycle discipline

A4 T5 enumerates the 2026 baseline:

1. `exec.CommandContext` with deadline — **✓ present** (all three audits)
2. `cmd.WaitDelay` for grace period — **✗ missing** (A3 D4-2 notes the supervisor leaves goroutine leak footguns)
3. `SysProcAttr.Setpgid: true` (POSIX) / `CREATE_NEW_PROCESS_GROUP` (Windows) — **✗ missing everywhere**
4. `SIGTERM → WaitDelay → SIGKILL` escalation — **✗ only SIGKILL** (A3 D4-1, supervisor.go:102)
5. Start pipe reader goroutines before `cmd.Start()` — **✓ stdout** ; **✗ stderr leak on core/client.go:213** (A3 D2-3, Critical)
6. Cross-platform process kill via `os.Process.Kill()` not `syscall.SIGKILL` — **✗** (A3 D4-1)

Scope: `internal/lsp/subprocess/launcher.go` and `supervisor.go` are the single owner. Fix once, all downstream consumers benefit. Estimated M effort with build-tagged POSIX/Windows split.

### 4.4 Version pinning / detection gaps

| Tool | Pin present? | Runtime check? | Audit |
|------|---|---|---|
| ast-grep (`sg`) | No (uses installed 0.40.5) | No (detectSGVersion() returns "unknown") | A2 REC-03 |
| powernap | Yes (v0.1.4) | N/A (library) | lsp-client.md + A4 T3.1 |
| gopls | No | No | A4 T3.5 (version-dependent features like MCP v0.20+) |
| pyright / tsserver | No | No | — |

Action: adopt A4 T3.5 recommendation — quarterly upgrade cadence documented in `lsp-client.md` already; extend to an ast-grep minimum pin (>= 0.42.1) with CI version check.

### 4.5 16-language neutrality gap

A2 §D7 finds 11 of 16 ast-grep rule directories empty. A3 §D3 finds fallback diagnostics cover only 5 of 16. The security-scanner directory is Go-only. This contradicts CLAUDE.local.md §15 HARD rule: "`internal/template/templates/` 하위는 16개 언어 동등 취급".

However, `internal/template/templates/` is for project scaffolding; `.moai/config/astgrep-rules/` is project-instance config. A reading: the HARD rule applies to *templates*, so the empty dirs in the installed project are acceptable, but the template deployment itself must seed at minimum schema-valid placeholder rule sets or document "no rules yet" clearly. Confirm with user in Open Question #1.

---

## 5. Improvement Roadmap

Priority scoring:
- **P0** — Critical bug, blocks production, <1 day fix
- **P1** — High-value architectural debt, <1 week including tests
- **P2** — Strategic (Phase 6/7 planning)
- **P3** — Research / v3.1+

Effort labels: XS (<2h) · S (half-day) · M (1-2 days) · L (week+) · XL (multi-SPEC).

### Tier 1: Critical bugs (block production)

| ID | Source | Problem | File:Line | Proposed fix (Go signature) | Effort | Depends on |
|----|--------|---------|-----------|------------------------------|--------|------------|
| IMP-V3U-001 | A1 B1 | MX validator misses all exported method receivers → P2/P3/P4 violations invisible in idiomatic Go | `internal/hook/mx/validator.go:19` | `var exportedMethodRe = regexp.MustCompile(\`^func\s+\(\w+\s+\*?\w+\)\s+([A-Z]\w+)\`)` — extend `extractFunctions` to match both regexes | S | — |
| IMP-V3U-002 | A3 D2-3 | LSP subprocess stderr pipe never consumed → deadlock when buffer >64KB | `internal/lsp/core/client.go:213` | `if result.Stderr != nil { go func() { defer result.Stderr.Close(); io.Copy(io.Discard, result.Stderr) }() }` after `NewSupervisor` | XS | — |
| IMP-V3U-003 | A3 D10-1 | `getOrSpawn` concurrent client race inserts half-ready client into cache | `internal/lsp/core/manager.go:332` | Add `sf singleflight.Group` field; replace naive cache insert with `m.sf.Do(language, func() (any, error) { c := m.clientFactory(sc); if err := c.Start(ctx); err != nil { return nil, err }; m.mu.Lock(); m.clients[language] = c; m.mu.Unlock(); return c, nil })` | S | — |
| IMP-V3U-004 | A2 REC-01 | ast-grep `Rule.Metadata` + `Rule.Note` dropped during YAML → Finding pipeline; OWASP/CWE lost | `internal/astgrep/models.go:72`, `scanner.go:361` | Add to `Rule`: `Note string \`yaml:"note,omitempty"\``; `Metadata map[string]string \`yaml:"metadata,omitempty"\``. In `scanWithRules`: `findings[i].Note = rule.Note; if rule.Metadata != nil && findings[i].Metadata == nil { findings[i].Metadata = rule.Metadata }` | S | — |
| IMP-V3U-005 | A1 B2 | `countFanIn` substring match: `funcName="New"` counts `NewContext`, `Renew` | `internal/hook/mx/validator.go:264,284` | Grep: add `-w` flag. Go side: `re := regexp.MustCompile(\`\b\` + regexp.QuoteMeta(funcName) + \`\b\`); count := len(re.FindAllIndex(data, -1))` | XS | — |
| IMP-V3U-006 | A2 REC-02 | `ScanMultiple` unbounded goroutines: 10K files → 10K `sg` subprocesses | `internal/hook/security/ast_grep.go:186` | `const maxConcurrent = runtime.NumCPU() * 2; sem := make(chan struct{}, maxConcurrent)`; inside loop: `sem <- struct{}{}; defer func(){ <-sem }()` | XS | — |
| IMP-V3U-007 | A1 P1 | MX `ValidateFiles` same unbounded goroutine pattern | `internal/hook/mx/validator.go:317` | Same semaphore pattern as IMP-V3U-006 | XS | — |

**Tier 1 total: 7 items, 1×S+2×S+2×XS+2×XS → well under 1 week aggregate.**

### Tier 2: High-value architectural fixes

| ID | Source | Scope | Effort | Notes |
|----|--------|-------|--------|-------|
| IMP-V3U-008 | A3 R3 | Unify `Diagnostic` types on `lsp.Diagnostic`. Retire `hook.Diagnostic` (string severity) and `gopls.Diagnostic`. Migrate `internal/ralph/engine.go` and `internal/loop/go_feedback.go` off `gopls.Diagnostic`. | M | BC for hook JSON consumers — coordinate with SPEC-V3R2-RT-001 |
| IMP-V3U-009 | A3 R5 | Unify `PhaseType` × 2 into `internal/lsp/phase` package | S | Touches `hook/types.go:66`, `core/quality/trust.go:72` |
| IMP-V3U-010 | A3 R6 + A2 7-3 | Canonicalize extension-to-language map on `config.LanguageForExt(ext string) string`; remove `aggregator.detectLanguageFromPath` (replace with `filepath.Ext` + lookup), remove `security/ast_grep.go:extensionToLanguage` map | S | Include Dart (`.dart` → `flutter`) and R (`.r` → `r`) to close A2 §7-2 gap |
| IMP-V3U-011 | A2 REC-05 | Consolidate three ast-grep JSON parsers into one public `astgrep.ParseSGOutput([]byte) ([]Match, error)` | S | Also centralizes 0-indexed→1-indexed conversion |
| IMP-V3U-012 | A3 R4 | Soft-deprecate `gopls/bridge.go`: add `@MX:WARN DEPRECATED` on `NewBridge`; migrate `ralph/` and `loop/` to `aggregator.GetDiagnostics`; flip `ResolveClientImpl()` default from `"gopls_bridge"` to `"powernap_core"` | M | Breaking change: add `BC-V3R2-UTIL-001` |
| IMP-V3U-013 | A3 R7 | `execTool` — respect caller deadline, only apply 30s default when context has no deadline: `if _, ok := ctx.Deadline(); !ok { ctx, cancel = context.WithTimeout(ctx, 30*time.Second); defer cancel() }` | XS | `internal/lsp/hook/fallback.go:258` |
| IMP-V3U-014 | A1 B3 | MX brace-in-string bug — use `go/scanner` token-aware parser in `extractFunctions`, or simple state machine that skips runes inside `"` … `"` and `\`` … `\`` | S | Resolves F1 feature gap too (proper Go parsing) |
| IMP-V3U-015 | A1 R4 / F2 | `@MX:REASON` presence check for every WARN/ANCHOR. Add `hasAnchorReason`, `hasWarnReason` to `funcInfo`; emit violation if WARN/ANCHOR without paired REASON | S | Protocol-mandated (currently unchecked) |
| IMP-V3U-016 | A1 R6 / F3 | Per-file limit enforcement. `func capViolations(vs []Violation, anchorLimit, warnLimit int) []Violation`; read thresholds from `ValidationConfig.Thresholds` | S | Requires IMP-V3U-017 |
| IMP-V3U-017 | A1 R5 / A4 | Wire `ValidationConfig` into `mxValidator`. `func NewValidatorWithConfig(cfg *ValidationConfig, projectRoot string) Validator` | XS | Preserve backward-compat `NewValidator` for existing callers |
| IMP-V3U-018 | A2 REC-04 | ast-grep context cancel leak: extract `(a *SGAnalyzer) executeWithTimeout(ctx, timeout, ...)` helper using `defer cancel()`; apply to `analyzer.go:242, 298` | XS | — |
| IMP-V3U-019 | A2 REC-03 | Implement `detectSGVersion()`: `cmd := exec.CommandContext(ctx, "sg", "--version"); out, err := cmd.Output(); return strings.TrimSpace(string(out))` | XS | Populates SARIF `tool.driver.version` |
| IMP-V3U-020 | A3 R8 | Cross-platform process kill: replace `syscall.SIGKILL` with `s.cmd.Process.Kill()` in `subprocess/supervisor.go:102`; use `os.Interrupt` instead of `syscall.SIGTERM` in `client.Shutdown` | XS | — |
| IMP-V3U-021 | A4 T5 + A3 Cross-cut | Subprocess hygiene baseline: set `cmd.WaitDelay = 10*time.Second`; POSIX: `cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}` (build tag `//go:build unix`); Windows: `CREATE_NEW_PROCESS_GROUP` | M | `internal/lsp/subprocess/launcher.go` — single owner for all subprocess spawn |
| IMP-V3U-022 | A3 §D9 / A4 T3.1 | Flip `config.ResolveClientImpl()` default from `"gopls_bridge"` to `"powernap_core"` to match `lsp-client.md v1.0.0` intent | XS | BC — coordinate with IMP-V3U-012 |

**Tier 2 total: 15 items. Largest risk item is IMP-V3U-012 (gopls bridge retirement path) — requires call-site migration before default flip.**

### Tier 3: Strategic improvements (A4 external research)

| ID | Source | Scope |
|----|--------|-------|
| IMP-V3U-023 | A4 T1.3 rec-2 | Promote `@MX:REASON` to first-class typed tag with controlled vocabulary: `rationale`, `invariant`, `hazard`, `temporary`, `external-constraint`. Draws on arXiv 2502.08172 intent taxonomy. |
| IMP-V3U-024 | A4 T3.4 | Use LSP 3.17 `textDocument/prepareCallHierarchy` + `callHierarchy/incomingCalls` to replace grep-substring fan-in. Authoritative, language-neutral, zero false positives. |
| IMP-V3U-025 | A4 T3.4 + T1.3 | Render `@MX` tags as LSP `textDocument/inlayHint` so non-moai editors display tag context without reading source. Feature-flagged. |
| IMP-V3U-026 | A4 T1.3 rec-3 | Adopt `key=value` syntax for all `@MX:*` tags. Tree-sitter-extractable. Matches Rust attribute macros. |
| IMP-V3U-027 | A4 T2.4 rec-3,4 | Adopt ast-grep `no-suppress-all` built-in rule + parameterized utils. Blocks silent `// ast-grep: disable` in agent output; enables shared patterns like `function-with-high-fan-in($N)`. |
| IMP-V3U-028 | A2 §D7 + CLAUDE.local.md §15 | Seed minimum ast-grep rule set across 15 currently-empty language dirs: at minimum one "TODO-must-have-issue" + one "no-secret-in-source" rule per language. Closes 16-language neutrality gap. |
| IMP-V3U-029 | A4 T2.4 rec-6 | Pre-commit hook integration: adopt `boidolr/ast-grep-pre-commit` (pinned SHA) + `pre-commit run --all-files` in CI. |
| IMP-V3U-030 | A2 REC-08 | SARIF `artifactLocation.uri` must be a URI — add `file://` scheme for absolute paths in `internal/astgrep/sarif.go:toFileURI`. |
| IMP-V3U-031 | A4 T3.5 rec-6 | Opt-in integration with gopls built-in MCP server (v0.20+, 12 tools including `go_diagnostics`, `go_rename_symbol`). Token-efficient for agentic Go workflows. Behind feature flag. |
| IMP-V3U-032 | A4 T3.5 rec-1 | Document powernap quarterly upgrade cadence (already partially in lsp-client.md). Extend to ast-grep (quarterly, minimum 0.42.1) with CI version-check smoke test. |

### Tier 4: Research / explore (v3.1+)

| ID | Source | Scope |
|----|--------|-------|
| IMP-V3U-033 | A4 T4 (arXiv 2510.22210 LSPRAG) | LSPRAG adoption for test coverage: LSP-guided RAG gives +174% line coverage for Go. Explore integration with manager-tdd/manager-ddd for auto-generated characterization tests. |
| IMP-V3U-034 | A4 T2.3, T5 rec-10 | `GOEXPERIMENT=jsonv2` migration prep. No code change needed; 5-12× JSON unmarshal speedup. Default in Go 1.26. Add build-tag CI job. |
| IMP-V3U-035 | A1 F4 | Cyclomatic complexity check (P2 WARN @ complexity >= 15, if_branches >= 8) — currently entirely absent. Either ast-grep rule or dedicated AST walker. |
| IMP-V3U-036 | A4 T2.1 rec-5 | Explore `ast-grep-wasm` browser execution path for docs-site interactive rule preview (adk.mo.ai.kr). |
| IMP-V3U-037 | A1 Q3 | Cross-file `@MX:ANCHOR` demotion when per-file limit exceeded (currently undefined behavior per mx-tag.md). Requires cross-file state, low priority. |

---

## 6. SPEC Mapping — Integration with v3R2

The v3R2 bundle ships 35 SPECs across 9 theme prefixes (CON / EXT / HRN / MIG / ORC / RT / SPC / WF). Cross-check of the utility-layer improvements against these SPECs:

| Improvement ID | Closest v3R2 SPEC | Gap assessment |
|----------------|-------------------|-----------------|
| IMP-V3U-001 (MX method receiver) | SPEC-V3R2-SPC-002 (@MX TAG v2 w/ hook JSON + sidecar index) | **GAP** — SPC-002 is about sidecar JSON persistence, not validator core-engine bugs. Either amend SPC-002 with a §Validator Correctness section, OR propose new **SPEC-V3R2-UTIL-001**. |
| IMP-V3U-002 (LSP stderr leak) | SPEC-V3R2-RT-006 (Hook Handler Completeness) | **GAP** — RT-006 covers 27-event hook coverage, not LSP subsystem. Propose new **SPEC-V3R2-UTIL-003**. |
| IMP-V3U-003 (getOrSpawn race) | — | **UNCOVERED** — no LSP correctness SPEC in v3R2. → SPEC-V3R2-UTIL-003. |
| IMP-V3U-004 (ast-grep metadata) | SPEC-V3R2-EXT-001..004 (Extension) | **LIKELY GAP** — EXT series appears to be about extension architecture, not ast-grep rule-loading correctness. → Propose **SPEC-V3R2-UTIL-002**. |
| IMP-V3U-005/007 (MX word-boundary + goroutine pool) | SPEC-V3R2-SPC-002 | **GAP** — same as IMP-V3U-001. → SPEC-V3R2-UTIL-001. |
| IMP-V3U-006 (ast-grep ScanMultiple semaphore) | — | **UNCOVERED** — no ast-grep concurrency SPEC. → SPEC-V3R2-UTIL-002. |
| IMP-V3U-008/009/010/011 (type unification) | — | **UNCOVERED** — cross-package refactor not owned by any v3R2 SPEC. → SPEC-V3R2-UTIL-003 or standalone. |
| IMP-V3U-012/022 (gopls bridge soft-deprecate + default flip) | SPEC-GOPLS-BRIDGE-001 (legacy) + SPEC-LSP-CORE-002 | **PARTIAL** — LSP-CORE-002 owns the dual-client architecture; amend it with the retirement timeline + default flip. New BC-V3R2-UTIL-001 needed. |
| IMP-V3U-013 (execTool deadline) | — | **UNCOVERED**. → SPEC-V3R2-UTIL-003. |
| IMP-V3U-014 (brace-in-string) | SPEC-V3R2-SPC-002 | **GAP** — core validator bug. → SPEC-V3R2-UTIL-001. |
| IMP-V3U-015/016/017 (REASON, per-file, config wiring) | SPEC-V3R2-SPC-002 | **PARTIAL** — SPC-002 mentions sidecar; amend with feature-completeness requirements from A1 §D6. |
| IMP-V3U-018 (context cancel leak) | — | → SPEC-V3R2-UTIL-002. |
| IMP-V3U-019 (detectSGVersion) | — | → SPEC-V3R2-UTIL-002. |
| IMP-V3U-020/021 (subprocess hygiene, cross-platform) | SPEC-V3R2-RT-006 | **PARTIAL** — RT-006 about hook handler coverage; subprocess layer is `internal/lsp/subprocess/`. Amend RT-006 or → SPEC-V3R2-UTIL-003. |
| IMP-V3U-023-026 (typed REASON, key=value, inlay hints, callHierarchy) | SPEC-V3R2-SPC-002 | **AMEND SPC-002** with protocol v2 syntax decisions. |
| IMP-V3U-027-032 (Tier 3 strategic) | SPEC-V3R2-EXT-*, WF-* | **PLAN PHASE** — route into Phase 6/7 planning, not a single SPEC. |

### Proposed new SPECs (3 total)

1. **SPEC-V3R2-UTIL-001: MX Validator Correctness + Protocol v2 Enforcement** — owns IMP-V3U-001, 005, 007, 014, 015, 016, 017. Scope: core-engine correctness bugs + protocol feature gaps (REASON/per-file-limits/config wiring). Priority P0. Does NOT duplicate SPC-002 (which owns sidecar JSON + hook JSON integration).
2. **SPEC-V3R2-UTIL-002: ast-grep Integration Hardening** — owns IMP-V3U-004, 006, 011, 018, 019, 028, 030. Scope: Rule metadata propagation, concurrency, version detection, JSON parser consolidation, 16-language rule seeding, SARIF compliance. Priority P1.
3. **SPEC-V3R2-UTIL-003: LSP Subsystem Consolidation + Subprocess Hygiene** — owns IMP-V3U-002, 003, 008, 009, 010, 013, 020, 021, plus coordination with IMP-V3U-012 (gopls soft-deprecate). Scope: three Critical bugs + type unification + subprocess lifecycle baseline. Priority P0. BC-V3R2-UTIL-001 for default ClientImpl flip.

**Amendment to existing SPECs** (no new SPECs needed):
- **SPEC-V3R2-SPC-002** — amend with Tier-3 `@MX` protocol v2 syntax (IMP-V3U-023, 026), inlay hints (IMP-V3U-025), callHierarchy (IMP-V3U-024).
- **SPEC-LSP-CORE-002** — amend with gopls-bridge retirement timeline (IMP-V3U-012, 022).
- **SPEC-V3R2-RT-006** — optionally absorb IMP-V3U-021 (subprocess hygiene) since RT-006 owns hook handler completeness and hook handlers are subprocess-adjacent.

---

## 7. Recommendations for v3 Master Design

Proposed updates to `docs/design/major-v3-master.md`:

### 7.1 New Section 4.3.x — Utility Layer Runtime Invariants

Insert into Layer 3 Runtime coverage:

> **§4.3.x Utility-Layer Subprocess Hygiene (HARD)**. Every subprocess spawned from `internal/*` MUST apply the 2026 Go-idiom baseline: `exec.CommandContext` with caller-propagated deadline, `cmd.WaitDelay` non-zero (default 10s), `SysProcAttr.Setpgid: true` on POSIX / `CREATE_NEW_PROCESS_GROUP` on Windows, and a goroutine draining stderr before `cmd.Start()`. Rationale: A3 §D2-3 documents a Critical deadlock path on unread stderr; A4 T5 codifies the full pattern. Owner: `internal/lsp/subprocess/launcher.go` — single location for all compliant spawns.

### 7.2 Problem Resolution Matrix Addition

Add a utility-layer row:

| Problem ID | Description | Affected Layer | Source | Resolution |
|------------|-------------|----------------|--------|------------|
| P-U01 | Unbounded goroutine fan-out in MX validator and ast-grep ScanMultiple | Layer 3 Runtime (utility) | A1 P1 + A2 5-3 | SPEC-V3R2-UTIL-001 + 002 — bounded worker pool pattern |
| P-U02 | LSP subsystem type fragmentation (Diagnostic × 3, Phase × 2, extension-map × 3) | Layer 3 Runtime | A3 Top-5 #4, #5 | SPEC-V3R2-UTIL-003 — canonical types in `lsp/` + new `lsp/phase` package |
| P-U03 | MX validator blind to exported method receivers — majority of idiomatic Go invisible | Layer 2 SPEC & TAG | A1 B1 | SPEC-V3R2-UTIL-001 — dual regex (top-level + receiver) |

### 7.3 BC Catalog Addition

- **BC-V3R2-UTIL-001** (breaking): Default `ResolveClientImpl()` flips from `"gopls_bridge"` to `"powernap_core"`. Existing deployments must explicitly set `lsp.client: gopls_bridge` in `.moai/config/sections/lsp.yaml` to retain the legacy bridge path.

---

## 8. Open Questions

1. **16-language rule seeding scope** — should SPEC-V3R2-UTIL-002 seed actual rules (e.g., one "TODO-must-have-issue" + one "no-secret" per language), or just schema-valid placeholders? Trade-off: real rules require per-language expertise; placeholders preserve neutrality without pretending coverage. (A2 §D7, CLAUDE.local.md §15)
2. **Diagnostic type migration — breaking for external hook consumers?** — `hook.Diagnostic` uses string severity for JSON compat with external hook payloads. Migrating to `lsp.Diagnostic` (int severity) changes the wire format. Is this a BC worth taking, or should hook emit a JSON-projection view while keeping `lsp.Diagnostic` internal? (A3 R3 + OQ #4)
3. **gopls bridge retirement timeline** — `internal/ralph/engine.go` and `internal/loop/go_feedback.go` currently import `gopls.Diagnostic` directly. Who owns their migration to `aggregator.GetDiagnostics` — is it in-scope for SPEC-V3R2-UTIL-003, or does it need a separate migration SPEC? (A3 OQ #5)
4. **Stderr logging vs discard** — gopls emits structured log events on stderr during startup. Discarding (IMP-V3U-002) fixes the deadlock but loses diagnostics. Should the default be `io.Discard` with an opt-in "capture stderr to slog" config flag, or log-by-default? (A3 OQ #3)
5. **Windows support scope for MX `countFanIn`** — A1 Q4 notes grep is not available on Windows. Is Windows a supported target for MX validation, or is the session_end hook POSIX-only? If supported, replace grep with pure-Go `filepath.WalkDir` + `regexp.FindAll`. (A1 Q4)
6. **Cyclomatic complexity check ownership** — IMP-V3U-035 (F4 gap). Is this owned by ast-grep (structural rule) or MX (dedicated AST walker)? Protocol mandates the check; neither layer currently implements it. (A1 Q5)
7. **Ast-grep suppression policy** — adopt `no-suppress-all` (IMP-V3U-027) project-wide, or allow `# ast-grep: disable-next-line` with mandatory issue URL (IntelliJ/Dart pattern)? (A4 T2.4)
8. **v3.0 vs v3.1 cutline** — which of Tier-3 items land in v3.0 GA vs v3.1? Suggest minimum v3.0 includes IMP-V3U-023 (@MX:REASON typed vocabulary) and IMP-V3U-028 (16-lang rule seeding) for protocol completeness, deferring 024/025/031 to v3.1.

---

## 9. References

All file:line citations from the four source audits are preserved here; see the source audits for full context.

### A1 — MX (internal/hook/mx)
- `validator.go:19` — `exportedFuncRe` definition (B1, F1)
- `validator.go:40` — `NewValidator` signature (A1, M4)
- `validator.go:178-252` — `extractFunctions` God function (M1)
- `validator.go:194-210` — backward tag scan (B4, F2)
- `validator.go:224-230` — brace depth counter (B3)
- `validator.go:233` — goroutine detection (B5)
- `validator.go:258-293` — `countFanIn` (A2, B2, P2, M2)
- `validator.go:302-380` — `ValidateFiles` unbounded goroutines (P1)
- `config.go:98-121` — `ParseValidationConfig` silent errors (B7)
- `session_end.go:101-104, 167` — integration issues (H1, H2)

### A2 — ast-grep (internal/astgrep, internal/hook/quality, internal/hook/security)
- `internal/astgrep/models.go:72` — `Rule` missing Metadata/Note (REC-01)
- `internal/astgrep/scanner.go:302-306, 361-374` — Scanner metadata flow
- `internal/astgrep/analyzer.go:242, 298` — context cancel leak (REC-04)
- `internal/astgrep/rules.go:186-208` — multi-doc YAML split
- `internal/astgrep/sarif.go:197-206` — `toFileURI` incomplete (REC-08)
- `internal/cli/astgrep.go:193-198` — `detectSGVersion` placeholder (REC-03)
- `internal/hook/quality/astgrep_gate.go:107-191` — V1 path
- `internal/hook/security/ast_grep.go:172-213` — `ScanMultiple` goroutine flood (REC-02)
- `internal/hook/security/ast_grep.go:331-347` — supportedLanguages (Flutter/R missing)
- `.moai/config/astgrep-rules/sgconfig.yml:1-24` — 16-lang decl, 11 empty

### A3 — LSP (internal/lsp/*)
- `transport/transport.go:10, 65` — powernap wrapper; background context (D1-1)
- `core/client.go:199-204, 213-217` — StateShutdown after binary-not-found (D2-1); stderr leak (D2-3 Critical)
- `core/manager.go:227, 332-363` — detectLanguage duplicate; getOrSpawn race (D3-1, D10-1)
- `subprocess/supervisor.go:102` — `syscall.SIGKILL` portability (D4-1)
- `aggregator/aggregator.go:244-293` — detectLanguageFromPath 46-LOC (D6-2)
- `hook/gate.go:74-132, 329-335` — parallel config paths (D5-1); regression (D5-2)
- `hook/fallback.go:258-260` — hardcoded 30s timeout (D5-3)
- `hook/types.go:46, 66` — hook.Diagnostic, hook.PhaseType
- `hook/tracker.go:147-148` — typo `compareDignostics` (D8-2)
- `gopls/bridge.go:21-26, 259, 477-479, 591-622` — bridge duplication (D7-1, 7-2, 7-3)
- `core/quality/trust.go:72` — WorkflowPhase duplication
- `lsp/config/types.go:82-93` — default ClientImpl mismatch

### A4 — External corpus
- ast-grep 0.42.1 (2026-04-04): github.com/ast-grep/ast-grep/releases
- powernap v0.1.4: github.com/charmbracelet/x (crush production usage, 23K+ stars)
- LSP 3.17 spec: microsoft.github.io/language-server-protocol
- Key papers: arXiv 2510.22210 (LSPRAG), 2502.08172 (Intention), 2510.22956 (TAG), 2401.03003 (AST-T5), 2511.12884 (Agent READMEs)
- Go subprocess idioms: pkg.go.dev/os/exec, sigmoid.at, go-json-experiment/jsonbench

---

## Appendix: Priority Score Matrix

| ID | Title | Impact | Effort | Risk | Priority |
|----|-------|--------|--------|------|----------|
| IMP-V3U-001 | MX method receiver detection | High | S | Low | **P0** |
| IMP-V3U-002 | LSP stderr drain | Critical | XS | Low | **P0** |
| IMP-V3U-003 | LSP getOrSpawn singleflight | Critical | S | Low | **P0** |
| IMP-V3U-004 | ast-grep Rule metadata propagation | High | S | Low | **P0** |
| IMP-V3U-005 | MX countFanIn word-boundary | High | XS | Low | **P0** |
| IMP-V3U-006 | ast-grep ScanMultiple semaphore | High | XS | Low | **P0** |
| IMP-V3U-007 | MX ValidateFiles worker pool | High | XS | Low | **P0** |
| IMP-V3U-008 | Unify Diagnostic types | High | M | Medium (BC) | **P1** |
| IMP-V3U-009 | Unify Phase types | Medium | S | Low | **P1** |
| IMP-V3U-010 | Unify extension-to-language map | Medium | S | Low | **P1** |
| IMP-V3U-011 | Unify ast-grep JSON parser | Medium | S | Low | **P1** |
| IMP-V3U-012 | Soft-deprecate gopls bridge | High | M | Medium | **P1** |
| IMP-V3U-013 | execTool respect caller deadline | Medium | XS | Low | **P1** |
| IMP-V3U-014 | MX brace-in-string | Medium | S | Low | **P1** |
| IMP-V3U-015 | MX @MX:REASON enforcement | High | S | Low | **P1** |
| IMP-V3U-016 | MX per-file limit enforcement | Medium | S | Low | **P1** |
| IMP-V3U-017 | MX ValidationConfig wiring | Medium | XS | Low | **P1** |
| IMP-V3U-018 | ast-grep context cancel leak | Medium | XS | Low | **P1** |
| IMP-V3U-019 | detectSGVersion implementation | Medium | XS | Low | **P1** |
| IMP-V3U-020 | Cross-platform process kill | Medium | XS | Low | **P1** |
| IMP-V3U-021 | Subprocess hygiene baseline (WaitDelay, Setpgid) | High | M | Medium | **P1** |
| IMP-V3U-022 | Default ClientImpl flip | High | XS | Medium (BC) | **P1** |
| IMP-V3U-023 | @MX:REASON typed vocabulary | High | M | Low | **P2** |
| IMP-V3U-024 | LSP 3.17 callHierarchy for ANCHOR | High | L | Low | **P2** |
| IMP-V3U-025 | @MX inlay hints | Medium | L | Low | **P2** |
| IMP-V3U-026 | @MX key=value syntax | Medium | M | Medium (protocol change) | **P2** |
| IMP-V3U-027 | ast-grep no-suppress-all + param utils | Medium | S | Low | **P2** |
| IMP-V3U-028 | 16-language rule seeding | High | L | Low | **P2** |
| IMP-V3U-029 | Pre-commit hook integration | Low | S | Low | **P2** |
| IMP-V3U-030 | SARIF file:// URI compliance | Low | XS | Low | **P2** |
| IMP-V3U-031 | gopls MCP opt-in | Medium | M | Low | **P2** |
| IMP-V3U-032 | powernap/ast-grep upgrade cadence doc | Low | XS | Low | **P2** |
| IMP-V3U-033 | LSPRAG adoption | High | XL | High | **P3** |
| IMP-V3U-034 | jsonv2 migration prep | Medium | XS | Low (until Go 1.26) | **P3** |
| IMP-V3U-035 | Cyclomatic complexity check | Medium | L | Medium | **P3** |
| IMP-V3U-036 | ast-grep-wasm exploration | Low | XL | Low | **P3** |
| IMP-V3U-037 | Cross-file ANCHOR demotion | Low | M | Low | **P3** |

Totals: **P0 = 7** · **P1 = 15** · **P2 = 10** · **P3 = 5** · **Aggregate = 37 improvements** tracked against **8 Critical bugs** and **~20K words** of audit source.

---

*Synthesis produced from A1/A2/A3/A4 audit reports. No source code read beyond the citations in the four audits. Every claim carries a file:line or paper citation from the source material. Assumptions flagged as "likely" vs "confirmed" in §6 SPEC mapping pending user verification of v3R2 SPEC coverage.*
