# A4 — External Best Practices Survey

**Research Date:** 2026-04-23
**Mission:** Outside-view survey of code-annotation, AST analysis, LSP client, and Go subprocess best practices for the MoAI utility layer (no moai source read).
**Verified versions:**
- `ast-grep` latest: **v0.42.1** (2026-04-04) — crate `ast-grep-cli` on GitHub Releases.
- `charmbracelet/x/powernap` latest: **v0.1.4** (used pinned by crush `go.mod`).
- LSP specification still at 3.17 (no 3.18 yet as of 2026-04); several extension ecosystems now converging around MCP bridges on top of LSP.

---

## Topic T1: Code Annotation Systems

### T1.1 Landscape (frameworks surveyed)

Across the 2025–2026 window, code-annotation tooling splits into four archetypes:

1. **Keyword / comment linters** — Tools like `todocheck`, `trunk-toolbox`, `ai-linter`, and `dkotter/fixme-linter` catch free-form `TODO/FIXME/HACK/XXX`. They typically require a specific syntax (issue ID, owner, deadline) and fail CI on unowned items. Trunk.io's own review ("7 Different Ways to Lint for TODO Comments") concludes that a useful TODO linter must enforce syntax, support multiple keywords, and offer per-file ignores — a superset of what moai's `@MX:TODO` currently enforces.
2. **Language-native annotations** — Python type hints (PEP 526/604), Java annotations, Rust attribute macros, TypeScript decorators, Go `//go:build` / `//go:generate` directives. These are compile-/load-time metadata rather than comment-level tags.
3. **LLM-consumable semantic tags** — Emerging 2025 standard: "Tagging-Augmented Generation" (TAG) injects LLM-generated semantic annotations (entities, topics, discourse roles) directly into the input (arXiv 2510.22956). KONDA (SCI-K 2025) extends this to ontology-aligned knowledge graphs. The message for moai: inline semantic tags beat external sidecars when the LLM is the main consumer, because attention naturally focuses on inline content.
4. **Data-annotation platforms** — Label Studio, Snorkel, Keymakr's "LLM Data Annotation 2025" stack. Not directly relevant to code, but they validate the multi-level schema pattern (surface tag + deeper assessment) that moai's two-layer `@MX:NOTE / @MX:REASON` already implements.

Key surveyed tools (≈10):
- ast-grep (structural pattern engine — doubles as annotation surface via custom rules).
- Semgrep (YAML rules, partly commercialized; Opengrep fork 2025).
- CodeQL (QL-based semantic queries; GitHub Code Security 2025).
- todocheck, ai-linter, trunk-toolbox (keyword/comment linters).
- IntelliJ IDEA TODO system (in-IDE `// TODO(username): message, <url>` enforcement).
- Dart's `todo_format` linter (canonical example of per-project TODO policy).
- PyCharm / JetBrains structural search (SSR) — shaped ast-grep's DSL.
- rust-analyzer inlay hints (LSP 3.17 path for "annotate-adjacent-to-code" UX without modifying source).
- Flutter DCM (commercial, 2025) — demonstrates monetized code annotation linting viable as SaaS.
- Aikido.dev's SAST list 2026 ("Best 6 Static Code Analysis Tools Like Semgrep").

### T1.2 Papers (2–5 cited)

1. **"Tagging-Augmented Generation"** — arXiv 2510.22956 (Oct 2025). Injecting LLM-visible semantic tags directly into input documents improves retention on long contexts without any architectural change. Validates moai's inline `@MX` approach over external JSON sidecars.
2. **"Intention is All You Need: Refining Your Code from Your Intention"** — arXiv 2502.08172 (Feb 2025). Proposes a two-phase pipeline (Intent Extraction → Intent-Guided Modification) that treats code-review comments as a structured intent surface. Models reach higher fix quality when they see canonical intent tags vs raw review text.
3. **"Automating Intention Mining"** (Huang et al., TSE). Foundational study on mining intent from developer discussions; the taxonomy it derives (`problem`, `solution proposal`, `usage`, `clarification`, `fix`) maps well onto a future `@MX:REASON` vocabulary.
4. **"Agent READMEs: An Empirical Study of Context Files for Agentic Coding"** — arXiv 2511.12884. 67.7% of agentic context files carry architecture hints; only 14.5% encode non-functional guardrails. Direct implication: moai's `@MX:WARN` + `@MX:ANCHOR` guardrail category is under-represented in the ecosystem — a differentiator if kept.
5. **"Do AI Coding Agents Log Like Humans?"** — arXiv 2604.09409. Cursor users saw a 30% rise in static-analysis warnings and +41% cyclomatic complexity post-adoption. Motivates aggressive static-tag + LSP gating for any agentic coding system.

### T1.3 Recommendations for moai `@MX` v2

1. **Keep inline-first**: Tagging-Augmented Generation research confirms inline tags outperform sidecars on attention retention. Do NOT add a JSON sidecar; if needed, generate it as a derived index.
2. **Promote to `@MX:REASON` as a first-class tag type** with a controlled vocabulary drawn from the Huang et al. intent taxonomy: `rationale`, `invariant`, `hazard`, `temporary`, `external-constraint`. The current free-form `@MX:WARN` body becomes typed.
3. **Adopt `key=value` syntax** (e.g. `@MX:ANCHOR fan_in=7 invariant="non-nil before return"`). Aligns with Rust attribute macro semantics and Java annotations — AST-style structure that any tree-sitter parser can extract as attributes.
4. **Canonical issue-tracker link** (IntelliJ + Dart style): `@MX:TODO(goos) https://github.com/modu-ai/moai-adk/issues/702 deadline=2026-05-01`. `todocheck` in pre-commit fails PRs with missing/invalid issue URLs.
5. **Fan-in detection algorithm**: use ast-grep's `report-threshold` or a Graphify-style call-graph computation (citations → node; references → edge) to auto-flag functions with fan_in ≥ 3 for `@MX:ANCHOR`. Avoid hard-coded thresholds; make them configurable per project (`design.yaml` precedent).
6. **Multi-line annotations via fenced blocks** — the Dart + Flutter convention `// TODO(user): description \n //  second line` is widely recognized; don't invent a new continuation syntax.

Score-card summary: moai's `@MX` already covers NOTE/WARN/ANCHOR/TODO/LEGACY — 5 categories; the landscape maxes at ~7 (add REASON and ASSUMPTION). Syntax is the main evolution target.

---

## Topic T2: AST-based Analysis

### T2.1 ast-grep 2026 state

- **Latest:** v0.42.1 (CLI), released 2026-04-04 ([GitHub Releases](https://github.com/ast-grep/ast-grep/releases)).
- **Language coverage:** 26 built-in tree-sitter grammars with the default feature flag `builtin-parser`. Dart support was re-enabled in 0.42.x (previously removed, now restored). Covers all 16 moai-supported languages plus CSS/HTML/Dockerfile/YAML niche ones. Python bindings (`ast-grep-py`) at 0.42.0 (2026-03-16).
- **2026 additions worth adopting:**
  - ESQuery-style selectors (`:nth-child`, `:is`, `:not`, `:has`).
  - Parameterized util support — reusable rule fragments with arguments.
  - LSP: diagnostics for injected languages (SQL-in-Go, HTML-in-JS, etc.); deadlock fix on change events.
  - `ast-grep-wasm` — browser/no-binary execution path.
  - Built-in rule `no-suppress-all` — prevents blanket `// ast-grep: disable`.
  - `--color` flag on `test` command — better CI output.
  - LLM-oriented features: AI-generated rules from prompts, multi-option interactive fixes (blog posts late-2025 / early-2026, ast-grep.github.io/blog.html).

### T2.2 Semgrep / CodeQL comparison

Summarized from Konvu, Aikido, Doyensec, and ACR 2025–2026 comparisons:

| Axis | ast-grep | Semgrep | CodeQL |
|------|---------|---------|--------|
| Architecture | Tree-sitter CST pattern match | AST pattern match, YAML rules | QL-on-relational-DB, full semantic analysis |
| Scan time | Fastest (Rust, zero-DB) | 10–30 s (fast) | 5–20+ min (deep) |
| Language count | 26 (incl. Dart 0.42.1) | 40+ (some experimental) | ~12 (PHP gap) |
| Cross-file analysis | Limited | Limited for Ruby/PHP/Swift/Rust | All supported languages |
| License trajectory | Apache 2.0 (stable) | Tightening 2024-12; Opengrep fork Jan 2025 | Commercial via GHAS (repackaged April 2025) |
| False-positive rate* | n/a (lightweight) | ≈12% | ≈5% (most accurate) |
| Best use case | Structural search, codemods, LLM agents | Security on every PR | Deep semantic audit on schedule |

*Accuracy numbers from an independent 2024 evaluation cited by sanj.dev.

Trend: many security-mature orgs now run Semgrep on every PR *and* CodeQL on schedule. ast-grep fits a distinct niche — fast developer-facing structural search, library-callable, ideal for LLM-agent code inspection.

### T2.3 Performance benchmarks

- Semgrep scans are ~3× faster than CodeQL on equivalent codebases (Semgrep vs GitHub comparison page, 2025).
- Incremental CodeQL PR analysis improved 5–40% in 2025, closing the speed gap but still minutes-per-run.
- ast-grep cites itself as "a faster alternative to Semgrep CLI" in official tool-comparison doc, though no numeric benchmark is published by the project itself; community reports (Dropstone Research "AST Parsing at Scale: Tree-sitter Across 40 Languages") measure tree-sitter (the engine under ast-grep) at approximately 10–100× the throughput of equivalent regex scans on 50k+-file repos.
- Go 1.25 `GOEXPERIMENT=jsonv2` (released 2025-08) provides 5.6–12× faster JSON unmarshal vs encoding/json v1 — directly relevant if moai handles ast-grep JSON output at scale. Several Rust-to-Go JSON-RPC consumers have reported material speedups.

### T2.4 Recommendations for moai

1. **Stay on ast-grep** — the chosen engine fits moai's agent-assisted use case better than Semgrep (commercial licensing risk) or CodeQL (ops burden).
2. **Pin to v0.42.1+** and add a CI smoke check that exercises all 16 language grammars monthly (reuse powernap's upgrade policy pattern).
3. **Adopt the `no-suppress-all` built-in rule** project-wide to prevent agents from disabling scans silently.
4. **Use parameterized utils** to factor shared patterns (e.g., `function-with-high-fan-in($N)`); this enables per-project tuning without copy-paste rules.
5. **LSP path**: ast-grep's own LSP server (lsp-deadlock fix in 0.42) can be co-hosted with moai's gopls bridge so editors see both sets of diagnostics through a single LSP client.
6. **Pre-commit hook**: adopt `boidolr/ast-grep-pre-commit` with pinned SHA; mirror the Semgrep pattern (same hook in CI via `pre-commit run --all-files`).
7. **False-positive triage**: the 2026 "multi-option interactive fixes" feature is ideal for agentic triage — agent proposes Fix A/B/C and the human or evaluator-active picks one; avoids the binary accept/reject UX.
8. **SARIF output** (already supported) should be the default export for CI integration — aligns with GitHub Advanced Security and reduces lock-in to any one vendor.

---

## Topic T3: LSP Clients

### T3.1 powernap 2026 state (verified)

- **Latest:** `github.com/charmbracelet/x/powernap v0.1.4`. Verified by (a) `gh api repos/charmbracelet/x/tags` listing `powernap/v0.1.4` at top, (b) crush's `go.mod` requires `powernap v0.1.4` as of the main branch fetch.
- **v0.1.3 → v0.1.4 delta:** As noted in moai's own `lsp-client.md`, the diff inside `powernap/` is limited to `pkg/config/lsps.json` (+11/-7) — a data-only refresh synced from `nvim-lspconfig`. Zero API or ABI change. Other 28 commits in the charmbracelet/x range touch unrelated sub-packages.
- **Core abstractions (unchanged since 0.1.x):** `Connection`, `Router`, `Transport`, `lsp.Client` with full initialize → ready → shutdown lifecycle. Wraps `github.com/sourcegraph/jsonrpc2` v0.2.1 internally.
- **Production evidence:** charmbracelet/crush (23k+ stars) uses powernap as its LSP layer and has done so for ~12 months. Per DeepWiki "LSP Integration" page, crush implements a Manager + Client pattern on top of powernap; Manager merges user config with powernap's default server configs for gopls, typescript-language-server, nil, and others.

### T3.2 Alternative Go LSP libraries

| Library | Status 2026 | Notes |
|---------|-------------|-------|
| `sourcegraph/jsonrpc2` (direct) | Maintained | Low-level; would require re-implementing all LSP-specific abstractions. |
| `bugst/go-lsp` | Maintained but niche | Client/server scaffold; fewer production users than powernap. |
| `golang.org/x/tools/internal/lsp` | Internal to gopls | Not usable as a library (internal package). |
| `tliron/glsp` | Maintained | Server-side only; not a client. |
| `dirien/glsp-server` | Derivative | Same caveat. |
| crush's `lsp` package | Embedded in crush | Tight coupling to crush runtime; inspiration only. |

Conclusion: **no credible alternative supersedes powernap** for a multi-language Go LSP client in 2026. Building in-house would duplicate 1–2 K LoC for no functional gain.

### T3.3 Multi-LSP coordination (zed, helix, crush)

From DeepWiki's "Zed Language Registry and Adapters" + Helix fork blog + crush:

- **Zed (Rust)** implements a `LanguageRegistry` central service with `LspAdapterDelegate` trait. Each language can carry multiple adapters (e.g., Go → gopls + copilot-language-server). `CachedLspAdapter` caches static configuration. Subsystems split into five concerns: Language Registry and Adapters, LSP Store Architecture, Language Server Lifecycle, Completions and Diagnostics, Multi-Language Server Coordination.
- **Helix (Rust, modal editor)** ships tree-sitter + LSP out of the box, philosophy "less configuration, more editing." Supports multiple LSPs per language via simple TOML config.
- **crush (Go)** uses powernap + a Manager/Client wrapper. Manager lazily starts servers and owns per-server diagnostic caches, open-file tracking, and state. `applyLSPDefaults()` enriches minimal user config with known-good defaults for gopls, typescript-language-server, nil, etc.

Common pattern: a Registry/Manager lazily spawns servers; each Client wraps the library primitive (powernap `Client`, tower-lsp `LanguageServer`, etc.) and adds editor-specific caches. None of them hand-roll JSON-RPC — the library layer handles the wire protocol.

### T3.4 LSP 3.17+ features moai should track

LSP 3.17 remains the current spec as of 2026-04. Key additions:

- **Pull diagnostics** (`textDocument/diagnostic`, `workspace/diagnostic`) — client requests diagnostics on-demand, reducing server push chatter. gopls supports this behind `pullDiagnostics: true` flag but keeps it off by default until parity with push.
- **Inlay hints** (`textDocument/inlayHint`) — small annotations adjacent to code (type hints, parameter names). Two kinds standardized: `Type` and `Parameter`. Relevant for moai's `@MX:ANCHOR` UX: could render MX tags as inlay hints instead of modifying source.
- **Type hierarchy** (`textDocument/prepareTypeHierarchy`, subtypes/supertypes) — useful for fan-in detection.
- **Call hierarchy** (`textDocument/prepareCallHierarchy`, incoming/outgoing calls) — direct source of fan-in data for `@MX:ANCHOR` promotion.
- **Inline values** (debug-adapter adjacent) — less relevant.
- **Notebook document support** — Jupyter/analogues; low priority for moai.
- **ServerCancelled** error code `-32802` — for long-running requests aborted by document change.

### T3.5 Recommendations for moai

1. **Keep powernap as the canonical client.** Upgrade to v0.1.4 (done per moai's `lsp-client.md`). Next upgrade: re-run the integration test matrix (gopls + pyright + tsserver) per the SPEC-LSP-CORE-002 policy.
2. **Adopt call hierarchy** from LSP 3.17 to populate `@MX:ANCHOR` auto-fan-in data. Use `textDocument/prepareCallHierarchy` + `callHierarchy/incomingCalls`.
3. **Don't pull diagnostics yet** — gopls warns performance isn't on par with push. Revisit when gopls flips the default.
4. **Render `@MX` tags as inlay hints** (optional feature flag). Avoids modifying source files for decorative annotations; the tag text lives in a sidecar while the LSP client displays it inline. Complements the inline-first recommendation in T1.3 rather than replacing it — keep canonical tags in source, mirror to inlay for non-MoAI editors.
5. **Subprocess lifecycle** (see T5): powernap already wraps `os/exec` with stdio pipes; ensure `exec.CommandContext` + `Setpgid: true` + `WaitDelay` are set at the moai wrapper layer to kill the whole LSP process tree on context cancellation (crush does this).
6. **gopls-specific bridge (SPEC-GOPLS-BRIDGE-001) remains justified** as an opt-in accelerator for `go` projects — gopls exposes a built-in MCP server since v0.20 (12 tools: go_diagnostics, go_rename_symbol, go_references, …) that cuts token cost for agentic workflows. Keep both paths coexisting via the `lsp.client` config key.
7. **Fallback strategy**: when LSP unavailable — degrade to tree-sitter (ast-grep already gives this for free). The research confirms tree-sitter CST captures enough to preserve most IDE-like features; regex is always a last resort. Do NOT skip checks silently; emit a single warning and fall back.

---

## Topic T4: Research Papers

1. **arXiv 2510.22210 — "LSPRAG: LSP-Guided RAG for Language-Agnostic Real-Time Unit Test Generation"** (Oct 2025; ICSE '26 accepted). Reuses LSP servers to give LLMs language-aware retrieval. Achieves +174.55% line coverage for Go, +213.31% for Java, +31.57% for Python vs best baseline. **Most directly applicable to moai**: validates that moai's LSP layer can feed test-gen agents for large coverage wins without per-language engineering.
2. **arXiv 2502.08172 — "Intention is All You Need: Refining Your Code from Your Intention"** (Feb 2025). Two-phase Intent Extraction + Intent-Guided Modification. Evidence that structured intent tags (templated) outperform raw review comments for driving LLM code edits — direct support for promoting `@MX:REASON` to typed vocabulary.
3. **arXiv 2510.22956 — "Tagging-Augmented Generation"** (Oct 2025). Inline semantic tags raise retention on long contexts — backs moai's inline-first `@MX` design over JSON sidecars.
4. **arXiv 2508.11126 — "AI Agentic Programming: A Survey"** (Aug 2025). Defines static vs adaptive agents and argues the big unsolved question is redesigning languages/compilers/tests for AI consumption, not just human consumption. Motivates moai's `@MX` as a machine-first annotation layer.
5. **arXiv 2511.12884 — "Agent READMEs: Context Files for Agentic Coding"** (Nov 2025). 67.7% of agentic context files carry architecture hints, only 14.5% carry NFR guardrails. Backs `@MX:WARN` / `@MX:ANCHOR` as differentiators.
6. **arXiv 2604.09409 — "Do AI Coding Agents Log Like Humans?"** (2026). Shows +30% static-analysis warnings and +41% complexity in post-Cursor codebases. Justifies aggressive LSP + ast-grep gating.
7. **arXiv 2401.03003 — "AST-T5: Structure-Aware Pretraining"**. CodeT5 + AST beats CodeT5 by 2–3 points on Bugs2Fix and Java-C# transpilation. Validates that AST structure is a genuine signal for code models, not a heuristic — supports investing in ast-grep-driven context for moai.
8. **arXiv 2510.04905 — "Retrieval-Augmented Code Generation: A Survey with Focus on Repository-Level Approaches"** (Oct 2025). Notes most deployed systems use shallow retrieval and "struggle with structural context." Direct architectural prompt for moai's future: structural > textual retrieval.

**Top 3 most relevant to moai utility layer:** LSPRAG (T3-focused), "Intention is All You Need" (T1), Tagging-Augmented Generation (T1).

---

## Topic T5: Go Subprocess Idioms 2026

Synthesis of pkg.go.dev `os/exec`, sigmoid.at, calmops, bigkevmcd, Mezhenskyi, and ramr/go-reaper sources:

1. **Always use `exec.CommandContext`** with a cancellable context (`context.WithTimeout` or derived from the caller). Go will invoke `os.Process.Kill()` on cancel; on POSIX this is SIGKILL by default.
2. **Always call `Wait()` (or `Run()`)** — otherwise the child becomes a zombie once it exits. Even if the parent only cares about stdout, never skip Wait.
3. **Set `cmd.WaitDelay`** (added in Go 1.20) to a non-zero grace period (5–30 s typical). Without it, `Wait()` can hang forever if the child spawned grandchildren holding the stdout/stderr pipe descriptors.
4. **Kill the whole process tree on POSIX**: set `SysProcAttr.Setpgid: true`, then on cancellation send `syscall.Kill(-pid, SIGKILL)`. Default `CommandContext` only kills the direct child — grandchildren are reparented to init. LSP servers sometimes spawn helper workers, so this matters.
5. **Prefer `SIGTERM → WaitDelay → SIGKILL`** escalation for graceful shutdown. Override `cmd.Cancel` to send SIGTERM, then let WaitDelay fire SIGKILL as the fallback.
6. **Start pipe reader goroutines before `cmd.Start()`**. Deadlocks occur if stdout or stderr buffers fill and no one is reading. Most LSP clients including powernap already do this.
7. **Windows**: Setpgid doesn't exist; use `syscall.SysProcAttr{CreationFlags: CREATE_NEW_PROCESS_GROUP}` and Job Objects for tree kill. SIGINT doesn't translate. Test matrix matters.
8. **PID 1 / containers**: if moai ever ships as PID 1 in a container, use `dumb-init` or `ramr/go-reaper` to reap orphaned grandchildren. Not an issue today, but an upgrade note.
9. **Race-detector clean concurrent client state**: use `sync.Mutex` around shared LSP state (open-file caches, diagnostic maps). Crush's Client does this; moai should too. `go test -race ./...` is mandatory per project's CLAUDE.local.md.
10. **JSON-RPC hot path**: consider `GOEXPERIMENT=jsonv2` (Go 1.25) for encode/decode. 5.6–12× vs v1 on unmarshal per the official jsonbench. For powernap consumers, this is a drop-in at the build level — no API changes. Expected to become default in Go 1.26.
11. **Testable subprocess design**: abstract powernap's `Transport.Connection` behind an interface in moai so unit tests can mock stdio without spawning real language servers. Crush does this with an internal `LSPClient` interface.
12. **OTEL caveat (from moai's own memory)**: avoid `t.Setenv("OTEL_…")` in parallel tests — OTEL SDK initializes globals on first use and env mutation races the initialization. Use a fake/no-op exporter.

---

## Synthesis: Top-10 Improvements for moai Utility Layer

Prioritized by return-on-effort, each with one source anchor.

1. **Promote `@MX:REASON` to a typed first-class tag** with controlled vocabulary (rationale/invariant/hazard/temporary/external-constraint). Source: arXiv 2502.08172 ("Intention is All You Need") + Huang et al. TSE intent taxonomy.
2. **Add call-hierarchy-driven `@MX:ANCHOR` auto-promotion** using LSP 3.17 `callHierarchy/incomingCalls` instead of static analysis heuristics. Source: [LSP 3.17 spec](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/).
3. **Render `@MX` tags as LSP inlay hints** so non-moai editors display tag context without reading the source as a string. Source: LSP 3.17 `InlayHintClientCapabilities`.
4. **Adopt `key=value` syntax for all `@MX:*`** — closer to Rust attribute macros / Java annotations and tree-sitter-extractable. Source: [Rust Reference, Attributes](https://doc.rust-lang.org/reference/attributes.html) + Leapcell decorator comparison.
5. **Upgrade powernap pin to v0.1.4 already; set upgrade cadence to quarterly** with gopls + pyright + tsserver regression matrix. Source: moai's own `lsp-client.md` v1.0.0 + crush production usage (DeepWiki).
6. **Enforce `cmd.WaitDelay` + `Setpgid: true` on every subprocess spawn** in moai's transport layer. Source: pkg.go.dev `os/exec.Cmd`; sigmoid.at "Killing a process and all of its descendants in Go."
7. **Adopt ast-grep parameterized utils + `no-suppress-all` built-in** in moai's rule catalog; keeps false-positive triage ergonomic. Source: ast-grep CHANGELOG 0.42.0.
8. **Mirror ast-grep into CI via pre-commit** (`boidolr/ast-grep-pre-commit`) plus `pre-commit run --all-files` job. Source: Trunk/Semgrep pattern + gatlenculp 2025 guide.
9. **Plan for `GOEXPERIMENT=jsonv2`** (default in Go 1.26) on all JSON-RPC paths — no code changes, 5–12× unmarshal speedup. Source: go-json-experiment/jsonbench + aran.dev "go 1.25's new json encoding package."
10. **Preserve gopls-bridge (SPEC-GOPLS-BRIDGE-001) as an opt-in** — gopls' built-in MCP server (v0.20+, 12 tools incl. `go_diagnostics`, `go_rename_symbol`) gives token-efficient agent access that a generic LSP client cannot match. Source: [Gopls features index](https://go.dev/gopls/features/).

---

## Sources (all URLs verified via WebFetch/WebSearch + gh api)

### ast-grep & AST analysis
- [ast-grep Releases (GitHub)](https://github.com/ast-grep/ast-grep/releases) — latest 0.42.1 (2026-04-04), verified via `gh api`.
- [ast-grep CHANGELOG](https://github.com/ast-grep/ast-grep/blob/main/CHANGELOG.md)
- [ast-grep Tool Comparison](https://ast-grep.github.io/advanced/tool-comparison.html)
- [ast-grep Blog](https://ast-grep.github.io/blog.html)
- [ast-grep-cli on PyPI](https://pypi.org/project/ast-grep-cli/)
- [Semgrep vs CodeQL (Konvu, 2026)](https://konvu.com/compare/semgrep-vs-codeql)
- [Semgrep vs CodeQL: Patterns vs Semantic (ACR)](https://aicodereview.cc/blog/semgrep-vs-codeql/)
- [Aikido: Best 6 Static Code Analysis Tools Like Semgrep in 2026](https://www.aikido.dev/blog/semgrep-alternatives)
- [Dropstone Research: AST Parsing at Scale across 40 Languages](https://www.dropstone.io/blog/ast-parsing-tree-sitter-40-languages)
- [tree-sitter GitHub](https://github.com/tree-sitter/tree-sitter)
- [boidolr/ast-grep-pre-commit](https://github.com/boidolr/ast-grep-pre-commit)
- [ast-grep-pre-commit (alt, lucasrcezimbra)](https://github.com/lucasrcezimbra/ast-grep-pre-commit)

### LSP clients & multi-LSP
- [charmbracelet/x GitHub](https://github.com/charmbracelet/x)
- [charmbracelet/crush GitHub](https://github.com/charmbracelet/crush)
- [Crush LSP Integration (DeepWiki)](https://deepwiki.com/charmbracelet/crush/6.4-lsp-integration)
- [Crush Language Server Protocol (DeepWiki)](https://deepwiki.com/charmbracelet/crush/7.1-language-server-protocol)
- [Zed Language Registry & Adapters (DeepWiki)](https://deepwiki.com/zed-industries/zed/5.1-assistant-panel-and-integration)
- [Helix ML: Forked Zed for Agent Fleet Orchestration](https://blog.helix.ml/p/how-we-forked-zed-to-run-a-fleet)
- [Tulio Cunha: Neovim vs VS Code vs JetBrains vs Zed vs Helix (2025)](https://www.tuliocunha.dev/blog/neovim-lazyvim-vs-vscode-jetbrains-zed-helix-2025/)
- [bugst/go-lsp](https://github.com/bugst/go-lsp)
- [LSP 3.17 Specification (Microsoft)](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/)
- [Gopls Diagnostics](https://go.dev/gopls/features/diagnostics)
- [Gopls Features Index](https://go.dev/gopls/features/)
- [Claude Go LSP Plugin](https://claude.com/plugins/gopls-lsp)
- [rust-analyzer Inlay Hints (DeepWiki)](https://deepwiki.com/rust-lang/rust-analyzer/3.3-inlay-hints)

### Code annotations & comment mining
- [Tagging-Augmented Generation (arXiv 2510.22956)](https://arxiv.org/html/2510.22956v1)
- [Intention is All You Need (arXiv 2502.08172)](https://arxiv.org/pdf/2502.08172)
- [Huang et al., Automating Intention Mining, TSE](https://xin-xia.github.io/publication/tse185.pdf)
- [TnT-LLM (arXiv 2403.12173)](https://arxiv.org/abs/2403.12173)
- [KONDA SCI-K 2025](https://sci-k.github.io/2025/papers/paper14.pdf)
- [Keymakr: LLM Data Annotation 2025](https://keymakr.com/blog/complete-guide-to-llm-data-annotation-best-practices-for-2025/)
- [Frontiers: LLMs extract metadata for neuroimaging (2025)](https://www.frontiersin.org/journals/neuroinformatics/articles/10.3389/fninf.2025.1609077/full)
- [Trunk: 7 Different Ways to Lint for TODO Comments](https://trunk.io/blog/fixme-please-an-exercise-in-todo-linters)
- [ismailyagci/ai-linter](https://github.com/ismailyagci/ai-linter)
- [Dart Analysis: customize static analysis](https://dart.dev/tools/analysis)
- [IntelliJ TODO comments](https://www.jetbrains.com/help/idea/using-todo.html)
- [Bitsea: digital check-up](https://bitsea.us/blog/2025/01/the-digital-check-up-static-analysis-as-a-doctor-for-your-code/)

### Research papers
- [LSPRAG (arXiv 2510.22210)](https://arxiv.org/html/2510.22210v1) and [PDF](https://arxiv.org/abs/2510.22210)
- [RACG Survey (arXiv 2510.04905)](https://arxiv.org/abs/2510.04905)
- [AI Agentic Programming Survey (arXiv 2508.11126)](https://arxiv.org/pdf/2508.11126)
- [Agent READMEs (arXiv 2511.12884)](https://arxiv.org/html/2511.12884v1)
- [Do AI Agents Log Like Humans (arXiv 2604.09409)](https://arxiv.org/html/2604.09409v1)
- [Investigating Autonomous Agent Contributions (arXiv 2604.00917)](https://arxiv.org/html/2604.00917v1)
- [DeepCode (arXiv 2512.07921)](https://arxiv.org/abs/2512.07921)
- [AST-T5 (arXiv 2401.03003)](https://arxiv.org/abs/2401.03003)
- [AST-guided SVRF (arXiv 2507.00352)](https://arxiv.org/html/2507.00352)
- [Code the Transforms (arXiv 2410.08806)](https://arxiv.org/html/2410.08806)
- [iSEngLab AwesomeLLM4SE](https://github.com/iSEngLab/AwesomeLLM4SE)
- [Codebase-Memory (arXiv 2603.27277)](https://arxiv.org/html/2603.27277v1)

### Go subprocess & JSON perf
- [pkg.go.dev os/exec](https://pkg.go.dev/os/exec)
- [sigmoid.at: Killing a process and its descendants in Go](https://sigmoid.at/post/2023/08/kill_process_descendants_golang/)
- [Calmops: Go process management](https://calmops.com/programming/golang/go-process-management-subprocess/)
- [Mezhenskyi: Managing Linux processes in Go](https://mezhenskyi.dev/posts/go-linux-processes/)
- [ramr/go-reaper](https://pkg.go.dev/github.com/ramr/go-reaper)
- [bigkevmcd: Terminating processes in Go](https://bigkevmcd.github.io/go/pgrp/context/2019/02/19/terminating-processes-in-go.html)
- [go-json-experiment/jsonbench](https://github.com/go-json-experiment/jsonbench)
- [Saraikin: Go JSON Performance Showdown (2025)](https://saraikin.com/posts/golang-json-marshalling/)
- [aran.dev: Go 1.25's new JSON encoding package](https://aran.dev/posts/go-125/go-125-new-json-encoding-pkg/)
- [Medium: Go json/v2 Benchmark Showdown](https://medium.com/@kevalsabhani123/gos-new-json-v2-api-is-it-really-faster-a-benchmark-showdown-%EF%B8%8F-44df8bf20fd5)
- [packagemain: gRPC vs HTTP+JSON benchmark](https://packagemain.tech/p/protobuf-grpc-vs-json-http)

### Pre-commit & CI patterns
- [gatlenculp: Ultimate Pre-Commit Hooks Guide for 2025](https://gatlenculp.medium.com/effortless-code-quality-the-ultimate-pre-commit-hooks-guide-for-2025-57ca501d9835)
- [Semgrep: pre-commit integration](https://semgrep.dev/docs/extensions/pre-commit)
- [MLOps Pre-Commit Hooks](https://mlops-coding-course.fmind.dev/5.%20Refining/5.2.%20Pre-Commit%20Hooks.html)

---

**Document statistics:** ~5,100 words across 5 topics and synthesis. All verifications done 2026-04-23 via WebSearch + WebFetch + `gh api`. No moai source code was read per mission rules.
