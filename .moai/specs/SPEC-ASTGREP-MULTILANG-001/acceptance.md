# Acceptance Criteria — SPEC-ASTGREP-MULTILANG-001

> Given-When-Then scenarios, edge cases, quality-gate criteria, and Definition of Done for the curated ast-grep multi-language template baseline. Tier M.

## §D AC Matrix

| AC | Requirement(s) | Verification | Severity |
|----|----------------|--------------|----------|
| AC-AMR-001 | REQ-AMR-001, REQ-AMR-012 | Every shipped rule file is production-vetted (positive+negative fixture); no demonstrative scaffold ships | MUST |
| AC-AMR-002 | REQ-AMR-002 | Every shipped file/message/comment is English (grep for non-ASCII Hangul/CJK in shipped tree → 0) | MUST |
| AC-AMR-003 | REQ-AMR-003, NFR-AMR-002 | No internal SPEC-ID/REQ/AC/audit/date/SHA token in shipped tree; leak-test + neutrality CI green | MUST |
| AC-AMR-004 | REQ-AMR-008, REQ-AMR-011 | Shipped `sgconfig.yml` declares only existing ruleDirs; no `utils`; every ruleDir has ≥1 vetted rule | MUST |
| AC-AMR-005 | REQ-AMR-005 | `sg scan --config` on shipped sgconfig completes with no parse error / no missing-ruleDir error | MUST |
| AC-AMR-006 | REQ-AMR-006, REQ-AMR-009 | Fresh `moai init`/`update` deploys the baseline; no empty `.gitkeep` language dir, no scaffold | MUST |
| AC-AMR-007 | REQ-AMR-007 | `moai ast-grep` on a fixture yields findings only from vetted rules | MUST |
| AC-AMR-008 | REQ-AMR-004 | Security pattern-families treat every covered language identically; no language marked PRIMARY; coverage matrix recorded | MUST |
| AC-AMR-009 | NFR-AMR-001 | Each shipped rule matches its positive fixture and NOT its negative fixture | MUST |
| AC-AMR-010 | NFR-AMR-003 | `go-hardcoding.yml` behavior preserved; clean tree → zero findings | MUST |
| AC-AMR-011 | REQ-AMR-010 | Local dogfood tree unchanged by this SPEC (divergence permitted) | SHOULD |

## §D.1 Given-When-Then Scenarios

### Scenario 1 — Neutral sgconfig, no phantom ruleDir (REQ-AMR-005/008/011)
- **Given** the curated template `sgconfig.yml`,
- **When** `sg scan --config internal/template/templates/.moai/config/astgrep-rules/sgconfig.yml <fixture>` is run,
- **Then** the command completes without a config-parse error and without a "ruleDir not found" error, and every `ruleDirs` entry resolves to a shipped directory containing ≥1 rule.

### Scenario 2 — Deployed baseline is production-only (REQ-AMR-006/007/009)
- **Given** a fresh scratch project,
- **When** the maintainer runs `moai init <scratch>` (or `moai update`),
- **Then** `.moai/config/astgrep-rules/` is materialized with the curated baseline, contains **no** empty `.gitkeep` language directory and **no** demonstrative scaffold, and a subsequent `moai ast-grep <scratch>` (with `sg` present) reports findings only from vetted rules.

### Scenario 3 — Neutrality preserved (REQ-AMR-002/003)
- **Given** the shipped baseline files,
- **When** the neutrality guards run (`go test -run TestInternalContentLeak ./internal/template/...` and the `template-neutrality-check.yaml` grep),
- **Then** both pass, and a grep for Hangul/CJK and for internal SPEC-ID/REQ/AC tokens across the shipped tree returns zero matches.

### Scenario 4 — Cross-language equal treatment (REQ-AMR-004)
- **Given** the shipped security set covering ≥2 languages for a pattern-family,
- **When** the coverage matrix in `progress.md` §E.2 is inspected,
- **Then** each covered language carries the identical pattern-family, no language is labeled PRIMARY, and uncovered languages are listed as equal-priority future additions.

## §D.2 Edge Cases

- **EC-1 — `sg` absent**: with `sg` not in PATH, `moai ast-grep` and the gate return empty findings gracefully (existing scanner behavior — must not regress).
- **EC-2 — empty rules dir**: if a user deletes all rules, the scanner returns `[]` (no crash).
- **EC-3 — noisy pattern guard**: any candidate rule that matches its negative fixture is rejected (not shipped) — e.g. a `return $ERR`-style pattern must NOT ship.
- **EC-4 — multi-doc YAML**: rule files using `---` separators load correctly via `RuleLoader` split.
- **EC-5 — non-Go project**: a Go rule file is inert on a Python/JS project (language-scoped) — confirms retaining `go-hardcoding.yml` does not affect non-Go users.

## §D.3 Quality Gate Criteria

- `go test ./internal/astgrep/... ./internal/template/...` → PASS.
- `internal_content_leak_test.go` → PASS.
- `template-neutrality-check.yaml` grep → 0 forbidden matches.
- `sg scan --config <shipped sgconfig>` → exit without config error.
- Every shipped rule has a paired positive + negative fixture that passes NFR-AMR-001.

## §D.4 Definition of Done

- [ ] Curated baseline shipped to `internal/template/templates/.moai/config/astgrep-rules/`: neutral `sgconfig.yml` + retained `go-hardcoding.yml` + English-translated vetted security set (Core), plus bounded cross-language security layer.
- [ ] No empty `.gitkeep` language dir, no demonstrative scaffold in the shipped tree.
- [ ] All 11 AC (AC-AMR-001..011) satisfied; all MUST-severity ACs pass.
- [ ] Quality-gate criteria (§D.3) all green.
- [ ] Coverage matrix recorded in `progress.md` §E.2; verified gate blast-radius statement (GT-4) recorded in `progress.md` §E.3.
- [ ] Local dogfood tree unchanged; CLAUDE.local.md §2.2 remains accurate (no edit required).
- [ ] `make build` and commit performed by maintainer as a separate step (NOT part of this SPEC's plan-phase authoring).

## §D.5 Forward-Looking (deferred, tracked)

- Exhaustive per-language rules and per-language domain/idiom rules — equal-priority future SPEC.
- ast-grep commit-gate config key-path mismatch (`constitution.ast_grep_gate` vs `gate.ast_grep_gate`) — separate config-wiring follow-up SPEC.
- Local dogfood-tree cleanup — optional separate track.
