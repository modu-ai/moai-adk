# plan.md — SPEC-CODEX-SKILL-PATH-001

> Tier S, Class B (defect, cause established at plan-phase). Card t468, lane-8.
> Tree: worktree `.claude/worktrees/t468`, branch `WT-codex-path-parser`, base `d592b0551` (= origin/develop).

## §A. Context

### §A.1 Defect surface (measured, tree `d592b0551`)

- `internal/cli/doctor_codex.go` `codexStaleSkillFinding()` — stats `e.Path` raw (line ~338). No `~` expansion, no shape classification.
- `internal/codexwiring/skills.go` `ParseSkillEntries()` — raw TOML token into `SkillEntry.Path`; the struct comment asserts "absolute, as Codex writes it" — that premise is what breaks (hand-edited configs).
- Home resolution seam already exists: `internal/cli/mcp_codex.go` `resolveCodexHomeDir()` + `codexUserHomeDir = os.UserHomeDir` (t451). The `~` expansion MUST route through `codexUserHomeDir` (user home), NOT through codex home / `CODEX_HOME`.
- Existing fixture machinery to extend, not duplicate: `internal/cli/doctor_codex_test.go` — `stubCodexHome` (pins BOTH `CODEX_HOME` and the seam), `writeCodexHomeConfig` (synthetic `~/.codex/config.toml` under `t.TempDir()`), `codexSkillEntrySpec`.

### §A.2 RED-now baseline (measured this run, tree `d592b0551`)

- `os.Stat("~/definitely-not-a-real-path-t468")` → `no such file or directory`, `errors.Is(..., fs.ErrNotExist) == true` (throwaway /tmp probe).
- `os.Stat("C:\\Users\\goos\\SKILL.md")` → same ENOENT/`isNotExist=true` (darwin).
- Existing tests green: `go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported|TestCheckCodexWiring_HealthyHomeSkillsNoFinding' -count=1 -v` → both PASS.
- Latency fact (lead-verified, carried not re-measured): 0 false positives on this machine — all real entries are absolute. The defect is latent; the fixture is the only legitimate reproduction.

### §A.3 Design decision (single-pass, Tier S)

Insert a shape-classification step between `ParseSkillEntries` and `os.Stat`:

| Shape | Detection | Disposition |
|---|---|---|
| absolute | `filepath.IsAbs` (platform-native; catches `C:\...` on Windows) | stat as-is (current behavior) |
| home-relative | exactly `~` or `~/`-prefix | expand via `codexUserHomeDir` seam → stat expanded |
| relative | not absolute, not home-relative, no backslash | NOT missing; distinct "relative" classification (base unobserved — no guessed resolution) |
| oddly-formed | NOT absolute per `filepath.IsAbs` (checked FIRST — a Windows `C:\...` absolute stays absolute) AND (contains `\`, or `~user`-prefix) | NOT missing; distinct "oddly-formed" classification |

Stat outcomes keep the t451 taxonomy verbatim: `fs.ErrNotExist` (on a determinately-resolvable shape) → missing; anything else → indeterminate. The Detail line gains per-classification counts, and the "remove the stale entries" directive stays attached ONLY to the missing count.

Rejected alternative: decode TOML escapes + guess a relative base (rejected — unobserved premises become facts; false-finding risk moves rather than shrinks; see spec.md §F).

## §B. Known Issues (filtered, Tier S)

- **B5 CI 3-tier**: spec-lint / golangci-lint / Test can fail independently; classify NEW vs baseline before patching.
- **B6 spec-lint headings**: `### Out of Scope — ...` H3 sub-sections required (satisfied in spec.md §F).
- **B8 working-tree hygiene**: stage by explicit pathspec only; no `git add -A`.
- **B10 scope discipline**: touch only `internal/cli/doctor_codex.go`, `internal/cli/doctor_codex_test.go` (and at most a comment correction in `internal/codexwiring/skills.go` — see §F M3). t452/t462 surfaces are out of bounds.
- Cross-platform (NFC-3): the `\` classification must be guarded so a Windows host does not classify its own native absolute paths as odd — `filepath.IsAbs` runs BEFORE the backslash check.

## §C. Pre-flight (run before any code change)

```bash
git branch --show-current && git rev-parse --short HEAD
go build ./...
go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported|TestCheckCodexWiring_HealthyHomeSkillsNoFinding' -count=1
go test ./internal/codexwiring/ -count=1
golangci-lint run --timeout=2m 2>&1 | tail -5
```

## §D. Constraints (DO NOT VIOLATE)

- PRESERVE (t451): `CODEX_HOME` respected via `resolveCodexHomeDir()`; NO second home-resolution path (the `~` expansion reuses the `codexUserHomeDir` seam — it does not add a resolver); read-only checker; fail-open on every missing input.
- PRESERVE: stat error taxonomy — only `os.IsNotExist`-class feeds "missing"; permission-denied / symlink-loop / other stay indeterminate, surfaced in Detail, never acted on.
- ISOLATION (HARD): any reproduction or verification runs ONLY under the system temp root with `CODEX_HOME` isolated to the fixture (`stubCodexHome` + `t.TempDir()`). NEVER in the dev project, NEVER against the real `~/.codex`. No `t.Setenv("HOME", ...)` (parallel-test pollution — pin the seam instead, as the existing helper comment mandates).
- No new dependency (NFC-2); stdlib only.
- Do not modify: `internal/codexwiring/` parser logic (a doc-comment correction on `SkillEntry.Path` is the only allowed touch), any t452/t462-owned surface, `.moai/specs/` outside this SPEC directory.

## §E. Self-Verification (manager-develop deliverable)

- E-item matrix over spec.md §D AC-CSP-001-01 .. AC-CSP-001-08 with command + verbatim output + tree SHA per the 5-section attribution format.
- RED evidence for AC-CSP-001-01 / -03 / -04 / -08 captured BEFORE the implementation (verbatim failing-test output — the latent defect becomes red only through the fixture; -04 and -08 carry the D6-folded `~user` and bare-`~` claims respectively).
- **Test-count census (D3)**: every RED capture names the N test functions observed — `go test ./internal/cli/ -run '<selector>' -v 2>&1 | grep -c '^=== RUN'` prints the count alongside the capture. A selector matching 0 tests prints `ok ... [no tests to run]` and reads green — a RED capture without a positive test count is void.
- **Windows is a test-run verdict, not a build verdict (D1c / NFC-3)**: the CI windows matrix leg must RUN the affected tests (`go test ./internal/cli/ -run 'CodexSkillPath|TestCheckCodexWiring_StaleHomeSkillsReported|TestCheckCodexWiring_HealthyHomeSkillsNoFinding' -count=1`), not merely compile them — `GOOS=windows go build` alone proves nothing about classification. The reclassified-fixture case is covered structurally: the absolute-existing and absolute-missing fixture entries MUST be written as temp-root-derived absolute paths (`filepath.Join(t.TempDir(), ...)` / `<tempdir>/definitely-absent/SKILL.md`), NEVER as hardcoded `/nonexistent/...` literals — a leading-`/` literal is NOT absolute on Windows (`filepath.IsAbs` false), so it would silently reclassify as relative there and corrupt the regression judgment. The M1 fixture review asserts this before implementation lands.
- Regression: the two existing doctor stale-skill tests + `./internal/codexwiring/` suite stay green.
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (NFC-3).
- `golangci-lint run`: no NEW findings vs pre-flight baseline.

## §F. Milestones (single-pass, priority order — most reviewable decisions first)

- **M1 — RED: reproduction fixture + failing tests.** Extend `doctor_codex_test.go` with a `CodexSkillPath`-prefixed test set over `writeCodexHomeConfig` fixtures: home-relative-existing (AC-CSP-001-01), home-relative-missing (-02), relative (-03), backslash non-absolute fragment `skills\foo\SKILL.md` (-04), real-missing + symlink-loop pair (-06), seam-observability (-08), bare `~` + `~user` pair (asserted under -08 and -04 respectively — no separate AC, Tier S 8-AC budget). Absolute fixtures use temp-root-derived paths, never `/nonexistent/...` literals (§E, D1c). All RED tests assert the POST-fix contract and are captured failing before implementation, each with a positive test-count census (§E, D3). Fixture-isolation criterion AC-CSP-001-07 applies to this milestone's output.
- **M2 — GREEN: shape classification + `~` expansion.** Implement the §A.3 table in `codexStaleSkillFinding` (classification helper + expansion through the `codexUserHomeDir` seam). Absolute/indeterminate behavior untouched (AC-CSP-001-05/-06 regression guards). Flip M1's RED set green.
- **M3 — reporting + doc sweep + full verification.** Detail line carries per-classification counts; remove-directive bound to the missing count. Correct the `SkillEntry.Path` "absolute, as Codex writes it" comment in `internal/codexwiring/skills.go` to state the declared-as-written posture. Run §C pre-flight again + E-item matrix; affected-package tests only (`./internal/cli/`, `./internal/codexwiring/`) — no local full suite.

## §G. Anti-Patterns

- Guessing a relative-path base "to be helpful" — explicitly prohibited (REQ-CSP-003; unobserved premise).
- Collapsing the new shape classifications into the existing `indeterminate` bucket — they are declared classifications, distinct from stat-failure indeterminacy; merging them loses the non-destructive report.
- Expanding `~` against `CODEX_HOME` — wrong home; `CODEX_HOME` locates the config, the user home expands `~`.
- Writing the fixture under the dev project or reading the real `~/.codex` "just to sanity-check" — HARD isolation violation.
- Test-after GREEN for M1: the RED output must exist before the implementation lands (E8).

## §H. Cross-References

- spec.md §B.1 defect surface; §C REQ-CSP-001..007; §D AC-CSP-001..008.
- Lineage: t451 landed code (`doctor_codex.go`, `codexwiring/skills.go`); related SPEC-CODEX-WIRING-001 (doctor surface owner).
- Boundary: t452 (skill wiring), t462 (e2e sweep) — out of scope.
- `verification-claim-integrity.md` §1 (no unobserved premise) — grounds REQ-CSP-003's no-guessed-base rule.
