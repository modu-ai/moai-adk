# plan.md — SPEC-CLI-TUX-V3-005

> **Implementation plan for AC-TUX3-020 Printer migration.** This is a DDD/behavior-preservation migration: characterization tests are written BEFORE any refactor, and the ratchet AC (count < 38) is the exit gate.

---

## §A. Context

The `printer.Printer` interface (`internal/cli/printer/printer.go`, SPEC-CLI-TUX-V3-001) is the single output gateway for CLI commands. 38 direct `fmt.Print*` call sites remain in non-test `internal/cli/` sources. This SPEC migrates them onto the Printer interface, fulfilling the @MX:UPGRADE trigger from SPEC-001.

**Key finding** (research.md §D): 24 of 38 calls are straightforward migration targets (state.go, migration.go, tmux_integration.go). 13 are gap sites requiring a scope decision (banner.go TUI render, branch_protection.go interactive prompt). 1 is dead code (commented-out godoc example).

---

## §B. Known Issues

| Issue | Impact | Disposition |
|-------|--------|-------------|
| `uikit/banner.go` has no compatible Printer method | 12 calls blocked | M1 architecture gate — [NEEDS CLARIFICATION] |
| `branch_protection.go` prompt is interactive (Printer is one-way) | 1 call blocked | M1 gate — exclude as dead code (recommended) |
| `state.go` human-format is multi-line stdout text | Channel semantics unclear | M2 characterization + [NEEDS CLARIFICATION] |
| `tmux_integration.go` functions lack Printer parameter | Wiring required | M4 signature change to accept Printer |
| Baseline includes commented-out line (`new.go:433`) | Count inflated by 1 | Document as dead code; ratchet satisfied regardless |

---

## §C. Pre-flight (run-phase entry)

- [ ] Baseline re-measured: `grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v _test.go | wc -l` → expected 38 (record verbatim in progress.md §E.2)
- [ ] `go test ./internal/cli/printer/...` passes (Printer package is green)
- [ ] `go build ./...` passes (clean build)
- [ ] M1 architecture gate resolved (banner.go + branch_protection.go scope decision)

---

## §D. Constraints

- **No interface redesign**: the Printer interface is stable (SPEC-001). New methods are proposed ONLY if M1 gate determines an unavoidable gap (non-default path).
- **Characterization before refactor**: every migration milestone writes characterization tests capturing pre-migration output BEFORE changing any `fmt.Print*` call.
- **Channel discipline**: stdout = data (Data method), stderr = status (Info/Warn/Error/Success). Migration may intentionally re-route output from stdout to stderr — this is a behavior change that characterization tests must document.
- **Specific-path commits**: each milestone commits ONLY the files it touches. No `git add -A` (working tree has 60+ unrelated modified files).

---

## §E. Self-Verification (run-phase §E)

Run-phase manager-develop will populate:
- E1: AC PASS/FAIL matrix
- E2: `go build ./...` + `go vet ./...`
- E3: `go test -cover ./internal/cli/printer/...` + migrated files
- E4: subagent-boundary grep (no AskUserQuestion in CLI code)
- E5: `golangci-lint run`
- E6: ratchet grep count (must be < 38)

---

## §F. Milestones

> Ordered by decision-reversibility: the architecture gate leads because it determines whether the Printer interface needs extension. Mechanical migrations follow in risk-ascending order.

### M1 — Architecture gate: banner.go + branch_protection.go scope decision

**Goal**: Resolve the two gap-site dispositions BEFORE any code migration.

**Deliverables**:
- Decision recorded for `uikit/banner.go` (12 calls): migrate via interface extension / exclude as TUI render / route through Data()
- Decision recorded for `branch_protection.go` (1 call): migrate / exclude as dead code / defer
- If interface extension is approved: updated research.md §A with the new method contract

**[NEEDS CLARIFICATION: should uikit/banner.go's 12 fmt.Print\* calls be migrated to Printer (requiring a new styled-stdout method), or excluded as TUI render out-of-scope? Default recommendation: exclude.]**

**[NEEDS CLARIFICATION: should branch_protection.go's ttyConfirmer dead-code fmt.Printf prompt be migrated for ratchet credit, or explicitly excluded as nolint:unused deferred code? Default recommendation: exclude.]**

**Reversibility**: HIGH — this decision shapes the interface surface for all subsequent milestones. Getting it wrong means reworking migrations.

### M2 — state.go migration (11 calls)

**Goal**: Migrate all 11 `fmt.Print*` calls in `internal/cli/state.go` to Printer methods.

**Steps** (DDD ANALYZE-PRESERVE-IMPROVE):
1. ANALYZE: write characterization tests capturing exact stdout/stderr output of `runShowCheckpoint`, `runShowBlocker`, `printPhaseStateHuman` for both `--json` and human formats.
2. PRESERVE: verify characterization tests pass against current code.
3. IMPROVE: inject Printer into the cobra RunE closures (construct via `printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))`), replace `fmt.Print*` with appropriate Printer methods.
4. Verify: characterization tests still pass (with documented channel re-routing for status messages stdout→stderr).

**Mapping** (from research.md §C.1):
- JSON output (lines 110, 185) → `p.Data(...)`
- Status messages (lines 99, 149, 175) → `p.Info(...)`
- Human-format block (lines 120-129) → [NEEDS CLARIFICATION: Data() with composed string, or restructure]

**[NEEDS CLARIFICATION: should the human-format state display (printPhaseStateHuman, lines 120-129) route to stdout via Data() as a composed multi-line string, or should it be reformatted as JSON-only with human rendering through stderr Info() calls?]**

**Reversibility**: MEDIUM — channel routing changes affect scripted consumers.

### M3 — migration.go migration (8 calls)

**Goal**: Migrate all 8 `fmt.Print*` calls in `internal/cli/migration.go` to Printer methods.

**Steps**: Same DDD cycle as M2. Korean strings are preserved verbatim — the Printer methods accept format strings in any language.

**Mapping** (from research.md §C.2):
- JSON output (line 99) → `p.Data(...)`
- Success messages (lines 58, 149) → `p.Success(...)`
- Status messages (line 54) → `p.Info(...)`
- Human-format block (lines 104-111) → same [NEEDS CLARIFICATION] as M2

**Reversibility**: MEDIUM — same channel routing concern as M2.

### M4 — worktree/tmux_integration.go migration (5 calls)

**Goal**: Migrate all 5 `fmt.Print*` calls in `internal/cli/worktree/tmux_integration.go` to Printer methods.

**Steps**:
1. ANALYZE: characterization tests for tmux session creation output.
2. PRESERVE: verify tests pass.
3. IMPROVE: add `printer.Printer` parameter to the function containing these calls (currently no Printer parameter). Wire from the cobra RunE closure.
4. Replace `fmt.Printf(...)` with `p.Info(...)` for all 5 lines.

**Reversibility**: LOW — pure status output, direct Info() mapping, no channel ambiguity.

### M5 — Ratchet verification + gap-site documentation

**Goal**: Verify the ratchet AC is met and all gap-site dispositions are recorded.

**Steps**:
1. Run the canonical ratchet grep: `grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v _test.go | wc -l`
2. Verify count < 38 (expected: 14 if all migratable sites are done, minus gap-site dispositions).
3. Record verbatim grep output in progress.md §E.2.
4. Verify gap-site dispositions from M1 are documented in spec.md Out-of-Scope or acceptance.md.

**Reversibility**: LOW — verification step, no code change.

---

## §G. Anti-Patterns

- **AP-1: Migrating without characterization tests** — changing `fmt.Print*` to Printer methods without capturing pre-migration output risks silent behavior changes (channel routing, formatting, newline handling).
- **AP-2: Using Data() for status messages** — Data() writes to stdout and is semantically machine-consumable data. Status messages ("No checkpoint found") belong on stderr via Info().
- **AP-3: Extending the Printer interface without M1 gate approval** — adding methods to Printer.go without the architecture decision is scope creep.
- **AP-4: Kitchen-sink commits** — committing unrelated working-tree files alongside migration changes. Use specific-path commits only.
- **AP-5: Migrating banner.go before M1 decision** — the 12 banner.go calls are gap sites; migrating them before the scope decision is premature.

---

## §H. Cross-References

- research.md §A — Printer interface method catalogue
- research.md §C — per-call-site migration mapping
- research.md §D — gap analysis summary
- spec.md §B — GEARS requirements (REQ-TUX3-020 through 024)
- acceptance.md §D — AC matrix with ratchet baseline 38
- `internal/cli/printer/printer.go:64-112` — the interface being migrated to
- `internal/cli/clean.go:33` — canonical Printer wiring pattern
- SPEC-CLI-TUX-V3-003 acceptance.md:33 — original AC-TUX3-020 grep command
