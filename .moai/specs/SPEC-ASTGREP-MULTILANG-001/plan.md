# Implementation Plan — SPEC-ASTGREP-MULTILANG-001

> Ast-grep multi-language ruleset: Template-First curated production baseline. Tier M. No time estimates (priority-ordered milestones).

## §A Context

The distributed template ships only `go-hardcoding.yml`; the dogfood tree is experimental (empty stubs, demonstrative scaffolds, Korean messages, a `sgconfig.yml` declaring a nonexistent `utils` ruleDir + an internal SPEC-ID). Goal: ship a curated, neutral, `sg`-verified production baseline to `internal/template/templates/.moai/config/astgrep-rules/` without over-scoping into exhaustive 16-language authoring. All Ground Truth (GT-1..GT-5) is recorded in spec.md and was observed this session (`sg` = ast-grep 0.40.5 present).

## §B Known Issues / Risks

- **R1 — §15 elevation tension.** Shipping Go rules (`go-hardcoding` + Go security) could read as elevating Go. Mitigation: rule files are language-scoped by `sg` (a Go rule is inert on non-Go files), so retaining a Go rule file does not force a hierarchy on other users; the neutral contribution is the cross-language security set treated with equal opportunity. Recorded in §D.
- **R2 — over-scope creep.** Cross-language security authoring can balloon. Mitigation: bound to 2-3 pattern-families across a verifiability-gated language set; record a coverage matrix; defer the rest as equal-priority future work.
- **R3 — false-positive rules.** Several dogfood rules are noisy (`pattern: return $ERR`). Mitigation: NFR-AMR-001 requires positive+negative fixtures per shipped rule; unvetted rules are not shipped.
- **R4 — neutrality regression.** A mirrored SPEC-ID/Korean comment would violate §25/§15. Mitigation: M4 runs the leak-test + template-neutrality CI guards; every shipped file is English + token-free.
- **R5 — deployment assumption.** Relies on the blanket `//go:embed all:templates` (GT-5). Mitigation: M4 verifies a fresh `moai init`/`moai update` actually materializes the new files.

## §C Pre-flight (must hold before run-phase)

1. `sg --version` succeeds (ast-grep ≥ 0.40 present). [verified this session: 0.40.5]
2. `go test ./internal/astgrep/... ./internal/template/...` green on the base tree.
3. `internal_content_leak_test.go` + `template-neutrality-check.yaml` green on the base tree.

## §D Constraints & Decisions

- **D1 (neutrality)**: shipped files are English-only, zero internal-tracking tokens (§25). SPEC/REQ tokens live only in this `.moai/specs/` document, never in shipped rule files.
- **D2 (§15 equal opportunity)**: within each security pattern-family, every covered language is authored identically; no language is marked PRIMARY; uncovered languages are listed as equal-priority future additions, never as "unsupported."
- **D3 (self-hosting exception)**: `go-hardcoding.yml` is retained (pre-existing, proven) as a language-scoped artifact, not as an elevation of Go.
- **D4 (bounded scope)**: no exhaustive per-language authoring; no per-language domain/idiom rules; no dogfood cleanup; no gate config-wiring fix (all Out of Scope in spec.md).
- **D5 (Template-First)**: all edits are made in `internal/template/templates/…` first; `make build` and local sync are the maintainer's step and are explicitly NOT performed in this SPEC's authoring (per task instruction: do not `make build`, do not commit).

## §E Self-Verification (plan-phase)

- [x] SPEC ID passes canonical regex (`SPEC-ASTGREP-MULTILANG-001` — see spec.md HISTORY self-check; digit-leading `16LANG` corrected).
- [x] 12 canonical frontmatter fields present + `era: V3R6` + `tier: M`.
- [x] Requirements in GEARS form (Ubiquitous / When / While / Where / shall-not).
- [x] Exclusions section present with ≥1 `### Out of Scope — <topic>` H3 + `-` bullets.
- [x] Ground Truth recorded from observation, not assumption (GT-1..GT-5); task premise "gate OFF by default" corrected to a verified statement.
- [x] No implementation detail (exact rule patterns) prescribed in spec.md — deferred to run-phase.

## §F Milestones (priority-ordered)

### M1 — Neutral `sgconfig.yml` + ruleDir integrity
- Author a NEW template `sgconfig.yml`: English comments, no SPEC-ID/token, `ruleDirs` listing only dirs that ship with ≥1 vetted rule (drop `utils`, drop empty-stub dirs).
- Decide + record the shipped directory layout (which dirs ship). Confirm no empty `.gitkeep`-only dir and no demonstrative scaffold is carried into the template.
- Verify: `sg scan --config` parses the file with no missing-ruleDir error (REQ-AMR-005, REQ-AMR-008, REQ-AMR-011).

### M2 — Retain + re-verify `go-hardcoding.yml`
- Confirm the template `go-hardcoding.yml` is English + SPEC-ID-free (GT-1).
- Add a positive fixture (triggers each rule) + a negative fixture (clean tree → zero findings); `sg`-verify no false positives (NFR-AMR-001, NFR-AMR-003, REQ-AMR-001).

### M3 — English-translated + vetted cross-language security set
- Translate the existing Go security rules (crypto/injection/secrets/web) to English; re-vet each for true-positive quality; drop or fix noisy patterns.
- Extend the bounded security pattern-families (hardcoded-secret literals + injection: SQL/command/path) to additional languages gated by `sg` parse-reliability + genuine expressibility, with §15 equal treatment.
- Per-rule positive + negative fixtures; record the coverage matrix (pattern-family × language) in `progress.md` §E.2 (REQ-AMR-002, REQ-AMR-004, NFR-AMR-001).

### M4 — Template deployment + neutrality guards
- Verify the curated files deploy via the blanket embed: a fresh `moai init` (or `moai update`) into a scratch dir materializes `.moai/config/astgrep-rules/` with the curated baseline and no stub/scaffold (REQ-AMR-006, REQ-AMR-007, REQ-AMR-009).
- Run `internal_content_leak_test.go` + `template-neutrality-check.yaml`; both green (NFR-AMR-002, REQ-AMR-003).

### M5 — End-to-end verification + evidence
- `sg scan --config <shipped sgconfig.yml>` config-mode E2E against a mixed fixture tree: parse-clean, findings only from vetted rules (REQ-AMR-005, REQ-AMR-007).
- `moai ast-grep` CLI smoke against the fixture (exit-code behavior sane).
- `go test ./internal/astgrep/... ./internal/template/...` green.
- Record the verified gate blast-radius statement (GT-4) in `progress.md` §E.3 as closing evidence.

## §G Anti-Patterns to avoid

- Copying the dogfood tree verbatim into the template (propagates `utils`, SPEC-ID, Korean, stubs).
- Shipping empty `.gitkeep` language dirs or demonstrative scaffolds (REQ-AMR-009/012).
- Shipping an unvetted noisy rule (e.g. `return $ERR`) without a negative fixture (NFR-AMR-001).
- Marking any language PRIMARY, or labeling uncovered languages "unsupported" (§15 / REQ-AMR-004).
- Running `make build` / committing during plan-phase (task constraint).

## §H Cross-References

- spec.md §Ground Truth (GT-1..GT-5), §Requirements, §Exclusions.
- acceptance.md (AC-AMR-*, Given-When-Then, quality gate, DoD).
- `internal/astgrep/scanner.go`, `rules.go`; `internal/cli/astgrep.go`; `internal/hook/quality/astgrep_gate.go`.
- `internal/template/internal_content_leak_test.go`; `.github/workflows/template-neutrality-check.yaml`.
- CLAUDE.local.md §2 / §2.2 / §15 / §25.
