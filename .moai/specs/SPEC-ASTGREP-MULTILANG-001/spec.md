---
id: SPEC-ASTGREP-MULTILANG-001
title: "Ast-grep Multi-Language Ruleset — Template-First Curated Production Baseline"
version: "1.0.0"
status: completed
created: 2026-07-02
updated: 2026-07-02
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/template/templates/.moai/config/astgrep-rules"
lifecycle: spec-anchored
tags: "ast-grep, ruleset, template-first, neutrality, security, multi-language, curated-baseline"
era: V3R6
tier: M
related_specs: [SPEC-ASTG-UPGRADE-001]
---

# SPEC-ASTGREP-MULTILANG-001 — Ast-grep Multi-Language Ruleset (Template-First Curated Baseline)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-07-02 | 0.1.0 | manager-spec | Initial plan-phase authoring. Grounded in fresh source inspection (`internal/astgrep/scanner.go`, `rules.go`, `internal/cli/astgrep.go`, `internal/hook/pre_tool.go`, `internal/hook/quality/gate.go`, `internal/hook/quality/astgrep_gate.go`, both ruleset trees, `internal/config/types.go`, `internal/template/internal_content_leak_test.go`) plus `sg --version` (ast-grep 0.40.5 present). Corrects the orchestrator-provided SPEC ID `SPEC-ASTGREP-16LANG-001` → `SPEC-ASTGREP-MULTILANG-001` (segment `16LANG` is digit-leading and fails the canonical frontmatter regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` at `internal/spec/lint.go:578`; no digit-leading-segment precedent exists in the catalog). Corrects the task premise "gate is OFF by default" into a verified nuanced statement (see Ground Truth). |

## Overview (WHY)

The ast-grep ruleset is a **Template-First violation** and is **incomplete**. The local dogfood tree `.moai/config/astgrep-rules/` carries a full experimental multi-language scaffold, but the template source `internal/template/templates/.moai/config/astgrep-rules/` — the asset actually distributed to every user via `moai init` / `moai update` — ships **only `go-hardcoding.yml`**. Users receive a single-file, Go-only ruleset while the maintainer's dogfood tree pretends to be multi-language.

The dogfood tree is not production-ready either: it mixes empty `.gitkeep` stubs, demonstrative example rules, inconsistent message languages (Korean in `go/` and `security/`, English in `csharp/`), a `sgconfig.yml` that declares a **nonexistent `utils` ruleDir** and carries an internal **SPEC ID** (a template-neutrality §25 hazard if mirrored verbatim). Simply copying the dogfood tree into the template would propagate all of these defects to 16-language distribution.

This SPEC ships a **proper, curated, neutral production baseline** to the template: a corrected `sgconfig.yml`, the proven `go-hardcoding.yml`, and a small vetted cross-language security set — with uniform English messages, zero internal-tracking tokens, and `sg`-verified config-mode scanning. It deliberately does **not** attempt to author exhaustive rules for all 16 languages; disproportionate authoring is explicitly deferred as equal-priority future work.

## Ground Truth (verified this session — evidence over assumption)

Per `.claude/rules/moai/core/verification-claim-integrity.md`, the following are recorded as **observed**, not assumed. The task's premise "gate is OFF by default (no gate.yaml)" was refuted/refined by direct inspection.

**GT-1 — Template ships only one rule file.** `internal/template/templates/.moai/config/astgrep-rules/` contains exactly `go-hardcoding.yml` (no `sgconfig.yml`, no language subdirs, no `security/`). The template `go-hardcoding.yml` already has its SPEC-ID comment stripped (a `diff` shows the local copy carries `# SPEC: SPEC-SLQG-001` on line 2, the template copy does not).

**GT-2 — Local dogfood tree state.** 16 language dirs + `security/` (≈17 subdirs). Breakdown: `go/` has 5 domain rule files (Korean messages, several noisy — e.g. `go-error-not-wrapped` uses `pattern: return $ERR`, which matches every error return, a false-positive generator); `security/` has 4 files (crypto/injection/secrets/web), **all `language: go`** and Korean; 5 languages (csharp, elixir, kotlin, php, ruby) carry 3 demonstrative 9-line scaffolds each (null-deref/todo-marker/unused-var, English); 10 language dirs (cpp, flutter, java, javascript, python, r, rust, scala, swift, typescript) are empty `.gitkeep` stubs.

**GT-3 — `sgconfig.yml` defects.** The local `sgconfig.yml` declares a `utils` ruleDir that has no corresponding directory, and its header carries `SPEC-ASTG-UPGRADE-001` (Korean comments). Note: the current `internal_content_leak_test.go` C1 regex only matches `SPEC-(V3R[2-6]|AGENCY|WORKTREE)-`, so `SPEC-ASTG-UPGRADE-001` would *not* mechanically trip that test — but it is still a §25 forbidden-content-class (C1 SPEC-ID) violation and MUST be stripped from any shipped file.

**GT-4 — Gate blast-radius (task premise corrected).** Two independent consumers of the deployed ruleset:
- **`moai ast-grep` CLI** (`internal/cli/astgrep.go`): runs `astgrep.Scanner.Scan` **unconditionally** (no gate-config gating), `WarnOnlyMode: false`, `os.Exit(1)` on error-severity findings; requires `sg` in PATH (present here: ast-grep 0.40.5). This is an **always-on** consumer for any user with `sg` installed.
- **Commit-time quality gate** (`pre_tool.go` PreToolUse on `git commit` → `QualityGate.Run` → `RunAstGrepGateV2`): runs only when `config.Gate.Enabled` **and** `config.Gate.AstGrepGate.Enabled`. The Go struct reads a **top-level `gate:`** YAML key (`config/types.go:29`), but no template config section defines a top-level `gate:` — and the template's `ast_grep_gate:` block is nested under **`constitution:`** in `quality.yaml.tmpl` (a different config path, `constitution.ast_grep_gate`). Therefore for a config-loaded session `config.Gate.AstGrepGate.Enabled` resolves to the Go zero value `false` → **the commit-time ast-grep gate is effectively OFF for deployed users via the config path.** (`DefaultAstGrepGateConfig()` returns `Enabled: true`, but that is the config-nil fallback in `loadGateConfig()`, not the loaded-config path.)

**GT-4 conclusion:** the task's "OFF by default → only affects CLI" is **approximately true for the commit path** but for a *different, verified* reason (a config key-path mismatch, not a clean default), and it **omits** that the `moai ast-grep` CLI path is unconditionally active. Net: the deployed ruleset's quality matters regardless of the commit-gate default, because the CLI path always applies it and error-severity rules can block CLI usage (exit 1). The `constitution.ast_grep_gate` vs `gate.ast_grep_gate` key-path mismatch is recorded here as a separate config-wiring defect (Out of Scope — see below).

**GT-5 — Deployment mechanism.** Templates are embedded via `//go:embed all:templates` (`internal/template/embed.go`) — a blanket embed of the whole `templates/` tree. `catalog.yaml` tracks hashes, it does not gate deployment. Adding curated files under `internal/template/templates/.moai/config/astgrep-rules/` therefore ships them through the existing embed with no separate manifest wiring.

## Scope (WHAT)

This SPEC touches **`internal/template/templates/.moai/config/astgrep-rules/`** (the shipped baseline) and its verification fixtures/tests under `internal/astgrep/` and `internal/template/`. It does **not** modify the local dogfood tree, the scanner/CLI/gate Go code, or any config schema.

### The curated production baseline (the scope decision)

A firm minimal **Core** plus a bounded **Cross-language layer**; everything else is deferred equal-priority future work.

**Core (must ship):**
1. **Neutral `sgconfig.yml`** — English comments, zero internal SPEC-ID/REQ/AC token, `ruleDirs` listing **only** directories that ship with ≥1 vetted rule (no `utils`, no empty-stub dirs).
2. **Retained `go-hardcoding.yml`** (root) — already shipping, English, SPEC-ID-free; kept as the self-hosting artifact and re-verified against a fixture (no false positives on a clean tree). Rule files are language-scoped and inert for non-matching files, so retaining a Go rule file does not force a language hierarchy on non-Go users (§15 reasoning recorded in plan.md §D).
3. **English-translated + vetted security rules** — the existing Go security rules (crypto/injection/secrets/web) translated to English, re-vetted for true-positive quality, `sg`-verified against positive + negative fixtures.
4. **Removal of demonstrative scaffolds and empty stubs** from the shipped set (they are NOT copied into the template).

**Cross-language layer (bounded — proportional to Tier M):**
5. Extend the vetted security **pattern-families** (target: hardcoded-secret literals + injection-family: SQL / command / path) to additional languages where `sg` reliably parses the construct and the pattern is genuinely expressible, with **§15 equal-opportunity treatment** (identical pattern-family per covered language; no language marked PRIMARY). The exact covered-language set is a run-phase determination bounded by verifiability and recorded as a coverage matrix in `progress.md` §E.2 — the plan does NOT hard-commit an exhaustive per-language matrix.

## Requirements (GEARS)

### Ubiquitous

- **REQ-AMR-001**: The template ast-grep ruleset shall constitute a curated production baseline — every shipped rule file shall be production-vetted, not demonstrative or empty.
- **REQ-AMR-002**: Every shipped rule file, `sgconfig.yml`, message, note, and comment shall be written in English (template neutrality per CLAUDE.local.md §15/§25).
- **REQ-AMR-003**: No shipped file or comment shall contain an internal SPEC ID, REQ/AC token, audit citation, internal date, or commit SHA (§25 forbidden content classes).
- **REQ-AMR-004**: The shipped baseline shall not elevate any single language as PRIMARY; within each security pattern-family every covered language shall be treated with equal opportunity (§15).

### Event-driven (When)

- **REQ-AMR-005**: When `sg scan --config <shipped-sgconfig.yml> <tree>` is run, the scan shall complete without a config-parse error and without a missing-ruleDir error.
- **REQ-AMR-006**: When a user runs `moai init` or `moai update`, the curated `astgrep-rules` baseline shall be deployed to `.moai/config/astgrep-rules/` in the user's project.
- **REQ-AMR-007**: When `moai ast-grep <path>` is run with `sg` installed, the deployed baseline shall produce findings only from vetted rules — never from a demonstrative scaffold or an empty stub.

### State-driven (While)

- **REQ-AMR-008**: While the shipped `sgconfig.yml` declares any `ruleDirs` entry, that entry shall correspond to a directory that exists in the shipped baseline and contains at least one vetted rule.

### Where (capability gate)

- **REQ-AMR-009**: Where a language has no vetted rule in this baseline, the shipped baseline shall omit that language's directory entirely rather than ship an empty `.gitkeep` stub.
- **REQ-AMR-010**: Where the local dogfood set diverges from the shipped template baseline, that divergence shall be permitted per CLAUDE.local.md §2 and §2.2; this SPEC shall not require the local set to match the template.

### Unwanted (shall not)

- **REQ-AMR-011**: The shipped `sgconfig.yml` shall not declare a `utils` ruleDir, nor any ruleDir that does not exist in the shipped baseline.
- **REQ-AMR-012**: The shipped baseline shall not include demonstrative or placeholder rules (e.g. `todo-marker`, `unused-var`, `null-deref` scaffolds) unless each is individually re-vetted and confirmed to be a true-positive production rule.

### Non-Functional

- **NFR-AMR-001** (verifiability): Every shipped rule shall be `sg`-verified — it matches its intended construct against a positive fixture and does not match a negative fixture.
- **NFR-AMR-002** (neutrality CI): The `internal_content_leak_test.go` guard and the `template-neutrality-check.yaml` CI guard shall remain green after the baseline is added.
- **NFR-AMR-003** (no regression): `go-hardcoding.yml` behavior shall be preserved (retained functionally; re-verified, not rewritten in intent).

## Exclusions

The following are explicitly out of scope for this SPEC. Each excluded item is deferred as **equal-priority future work**, not deprecated.

### Out of Scope — Exhaustive 16-language rule authoring

- Authoring a complete rule set for every one of the 16 supported languages is disproportionate for a Tier M SPEC and is deferred. The baseline is curated + bounded, not exhaustive.

### Out of Scope — Per-language domain/idiom rules beyond security

- Concurrency, resource-safety, error-handling, and idiom rules (the local `go/` domain rules) are NOT shipped in this baseline. Several are unvetted and noisy (e.g. `pattern: return $ERR`). Per-language domain rules are future work applied equally across languages.

### Out of Scope — Local dogfood-experimental ruleset cleanup

- The local `.moai/config/astgrep-rules/` dogfood tree is left as-is (its divergence is documented in CLAUDE.local.md §2.2). Cleaning or aligning the dogfood tree is a separate optional track.

### Out of Scope — ast-grep commit-gate config key-path mismatch

- The `constitution.ast_grep_gate` (template `quality.yaml.tmpl`) vs `gate.ast_grep_gate` (Go struct `config/types.go`) key-path mismatch documented in GT-4 is a separate config-wiring defect. Fixing config wiring is out of scope here (this SPEC concerns ruleset content, not gate activation). It is recorded so a follow-up SPEC can address it.

### Out of Scope — Scanner / CLI / gate Go code changes

- `internal/astgrep/*.go`, `internal/cli/astgrep.go`, and the gate handlers are not modified. This SPEC changes shipped rule *content*, not scanning *behavior*.

## Cross-References

- `internal/astgrep/scanner.go` — `Scanner.Scan`: config-mode (`sgconfig.yml` present) vs recursive rule-load fallback; graceful no-op when `sg` absent or rules dir empty.
- `internal/astgrep/rules.go` — `RuleLoader.LoadFromDir` (recursive) + multi-doc YAML split.
- `internal/cli/astgrep.go` — `moai ast-grep` unconditional CLI scan path (`--rules-dir` default `.moai/config/astgrep-rules`).
- `internal/hook/quality/astgrep_gate.go` — `RunAstGrepGateV2` (suppression-policy check + sg scan; `BlockOnError`).
- `internal/config/types.go:29` — `config.Gate` reads top-level `gate:` (GT-4 key-path mismatch origin).
- `internal/template/internal_content_leak_test.go` — §25 neutrality guard (C1 SPEC-ID class regex).
- `CLAUDE.local.md` §2 (Template-First), §2.2 (astgrep-rules dogfood exception), §15 (language neutrality), §25 (template internal-content isolation).
- `SPEC-ASTG-UPGRADE-001` — the upgrade SPEC that created the current dogfood scaffold (provenance).
