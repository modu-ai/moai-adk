---
id: SPEC-ZONE-SHELL-CD-TRACKING-001
title: "Implementation plan — cd destination tracking (post-`--` operand incl. hyphen-leading, in-project absolute), cd-class regression cells, RED-first reproduction, D8b sweep"
version: "0.2.0"
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
tier: M
---

# SPEC-ZONE-SHELL-CD-TRACKING-001 — plan

## §A Context

Card t1584 (security P1, operator-approved expansion axis, Class C) — the t1574 ceiling-split follow-up under parent SPEC-SELF-IMPROVE-PROTECTED-ZONE-001. The protected-zone shell guard closed five shell-parsing classes in SPEC-ZONE-SHELL-PARSING-001 and separated the sixth (cd tracking) at its v0.4.0; the independent plan-audit's finding D7 then invalidated the separated draft's hyphen-leading allow control (`cd -- -zone` moves real bash into `./-zone` — a protected deletion, not an allow control), so this card gives the cd class its fresh fail-closed decision and lands the D8b residual sweep. Tree: branch `WT-zone-cd-deny` at `2aab5f797b75983e132af451da68f69e3426557b` (= origin/main; the FF absorb satisfied the card precondition "t1574 landed first"), status clean, observed 2026-10-09. Source: `internal/hook/protected_zone_shell.go`, 1569 lines post-t1574. Development mode: TDD (`constitution.development_mode: tdd`, quality.yaml:4) — the RED-first obligation is fully open (the inherited EV-6 probe was deleted after observation and never re-executed), so M1 step 1 owns the entire RED re-establishment. Plan phase writes NO implementation code — nothing under `internal/` changes in this phase.

## §B Known Issues

| # | Site | Class | RED evidence |
|---|------|-------|--------------|
| K1 | `protected_zone_shell.go:657-664` (cd branch dirs collection) + `:447-449` (`zoneNextCwd` `len(dirs) != 1` reset) | `cd -- <dir>` yields two directory words (`--` counted as one), the set degenerates to the root, and `cd -- zone_dir && rm a.log` answers allow | EV-6 shape 1 (allow observed at `b9ef00380`, probe deleted — re-captured at M1 step 1) |
| K2 | `protected_zone_shell.go:451` (`zoneNextCwd` `zoneIsAbs` reset) | an in-project absolute destination degenerates the set; `cd <project>/zone_dir && rm a.log` answers allow | EV-6 shape 2 (same demotion) |
| K3 | `protected_zone_shell.go:451` (`strings.HasPrefix(arg, "-")` reset) | a post-`--` hyphen-leading operand is misread as an option; real bash moves into `./-zone` and `cd -- -zone && rm a.log` deletes a protected path (audit D7 — the draft's allow control is withdrawn; the cell asserts DENY) | real-bash semantics per audit D7; not probe-observed — RED captured at M1 step 1 |
| D8b | stale-reference surface left by the predecessor's 0.3.0→0.4.0 fold-out | live surfaces still claiming t1574's SPEC owns the cd class, or citing the deleted EV-6 probe as evidence | this plan's measured sweep (below) — fix-here: 0 |

### D8b sweep — measured grep table (this tree `2aab5f797`, 2026-10-09)

Swept paths: `.moai/specs`, `.moai/docs`, `.moai/config`, `internal/`, `cmd/`, `.claude/rules`, `CHANGELOG.md`. Card evidence copies (`.moai/reports/t1574/`, `.moai/reports/t1584/`) are historical by design and are named, not counted as live.

| Pattern | Raw hits | Classification | Action |
|---------|----------|----------------|--------|
| `REQ-ZSP-009` | 4 | SPEC-ZONE-SHELL-PARSING-001 spec.md:26 (HISTORY 0.3.0), :27 (HISTORY 0.4.0), :86 (Out of Scope handoff naming t1584), plan.md:82 (§H separation record) — all historical or accurate-handoff; completed SPEC is read-only | none (historical) |
| `AC-ZSP-007` | 4 | the same four sites — historical or accurate-handoff | none (historical) |
| `EV-6` | 25 | 4 in-scope: the same SPEC-ZONE-SHELL-PARSING-001 sites (historical). The other 21 are UNRELATED same-number evidence ids owned by other SPECs (SPEC-AUTONOMY-GATE-REWIRE-001 plan/acceptance/research/design/progress/spec, SPEC-GPT-DOC-DRIFT-001, SPEC-UPDATE-ADD-CODEX-001 — each SPEC has its own EV-6 namespace; none cites t1574's cd probe) | none (historical + false-positive namespace collision) |
| `zz_probe_cd_test` | 0 | live surfaces clean; the deleted probe is referenced only inside card evidence (`.moai/reports/t1574/red-reproduction.md`, copied to `.moai/reports/t1584/red-reproduction.md`) — historical by design | none |
| `cd-tracking` | 5 | 3× SPEC-ZONE-SHELL-PARSING-001 (plan.md:82, spec.md:27, spec.md:86 — historical/handoff) + spec.md:84 (Out of Scope handoff heading) + `protected_zone_shell.go:860` (a live code comment describing why a wrapper-stripped cd must not be tracked — accurate semantics comment, NOT an ownership claim) | none |
| `cd tracking` | 3 | SPEC-ZONE-SHELL-PARSING-001 spec.md:26, :39, :84 — historical/handoff | none |
| `cd class` | 1 | SPEC-ZONE-SHELL-PARSING-001 spec.md:39 — handoff naming t1584 | none |
| `t1584` | 7 | ALL accurate handoff references naming t1584 as the cd-class owner (SPEC-ZONE-SHELL-PARSING-001 spec.md:27, :39, :84, :86, :87 + plan.md:82) — the desired end state | none |

**Sweep conclusion (measured, not assumed): fix-here = 0.** The 0.3.0→0.4.0 fold-out was clean — every live hit either records history in a completed (read-only) SPEC or already delegates the cd class to card t1584; the deleted EV-6 probe survives only inside card evidence. This SPEC's creation completes the handoff; no stale-reference edits are required, and the predecessor's Out of Scope handoff line stays as-is because it is correct.

## §C Pre-flight

1. `git rev-parse --short HEAD` → `2aab5f797`; branch `WT-zone-cd-deny`; `git status --short` clean (observed 2026-10-09).
2. Baselines to observe at M1 entry, before any edit: family `TestProtectedZone` (protected_zone_test.go:12) ok; landed sweep `TestProtectedZoneShellParsingMatrix` (protected_zone_shell_matrix_test.go:205) ok; `go vet ./internal/hook/` exit 0; `golangci-lint run ./internal/hook/...` 0 new issues vs the landed baseline.
3. Cross-SPEC conflict pre-scan: `grep -rn "Retired\|TestHarnessRetirement\|superseded" internal/hook` — confirm no retired-SPEC conflict in the zone surface before the first edit.
4. Slot lease before the minutes-long hook package suite (M3): `moai slot acquire --resource <shared-type> --max-duration <bound>` → run → `moai slot release` (lane-local verification rules).

## §D Constraints

- Files: `protected_zone_shell.go` + `protected_zone_shell_matrix_test.go` (+ an optional reproduction test file if the M1 step-1 RED capture needs a home before the matrix cells land). Nothing else.
- Completed SPEC bodies are read-only: zero edits to SPEC-SELF-IMPROVE-PROTECTED-ZONE-001 or SPEC-ZONE-SHELL-PARSING-001 (REQ-SIPZ-006 is cited provenance-only, spec D2).
- The retired id `REQ-ZSP-009` never becomes a live id (spec D4); fresh prefix REQ-ZSCD/AC-ZSCD only.
- Fail-closed direction inside the new cd classes (REQ-ZSCD-001..002); pre-existing under-match classes unchanged; no landed denial is weakened.
- No matrix cell asserts an allow verdict for a hyphen-leading directory operand (spec D1 — the D7 banned shape); the outside-root reset is never a matrix cell (REQ-ZSCD-003).
- Plan phase writes no implementation code — nothing under `internal/` changes this phase.
- Code comments English, matching the file's density (the file documents every fix with a round-style comment naming its card and requirement).

## §E Self-Verification

Run-phase closes with verbatim outputs for: E1 the AC matrix (acceptance.md §D) each row GREEN; E2 the RED-first capture (M1 step 1) — commands + verbatim stdout + exit code per shape, recorded in acceptance.md §B and progress.md §E.2; E3 the hook package suite `go test ./internal/hook/ -timeout 30m` under a slot lease; E4 `go vet ./internal/hook/`; E5 `golangci-lint run ./internal/hook/...`; E6 `git status --short` showing only the in-scope files; E7 preserved-behavior proof — the landed matrix cells and the `TestProtectedZone` family green with no assertion edits (`git diff` over the landed test files shows no assertion-line change); E8 the RED failure output (TDD — verbatim pre-GREEN evidence per the delegation template's E8 item).

## §F Milestones

Ordered by decision-reversibility: the world-semantics fix (the D7 disposition — the highest-design-uncertainty decision this card exists to make) leads; the matrix and mechanical verification follow.

### M1 (High) — cd interpretation fix (steps 1–4, RED-first)

1. **RED-first reproduction (the D8b landing point — this step OWNS the RED observation the ACs cite).** Author the three-shape reproduction set asserting the NEW deny semantics and run it against the CURRENT guard, capturing verbatim stdout + exit per shape: `cd -- zone_dir && rm a.log` (EV-6 shape 1), `cd <fixture-root>/zone_dir && rm a.log` with an in-project absolute destination (EV-6 shape 2 — the matrix harness builds the command from the fixture root, mirroring the deleted probe's `filepath.Join(root, "zone_dir")`), and `cd -- -zone && rm a.log` against a `-zone/` fixture (D7). Expected pre-fix verdicts: all three in the reset state (allow) — the reset sites are `:447` (two-operand) and `:451` (hyphen prefix, absolute). Home: cd cells added to `zoneParsingMatrixCells()` (they run RED inside the landed runner) or, if the runner's deny-assertion shape fights the RED capture, a temporary reproduction subtest mirroring the deleted probe's helper set — the run-phase implementer picks one home and records it. Record the verbatim RED in acceptance.md §B + progress.md §E.2. RED must be red for the STATED reason (the reset), never a selector miss — run with `-v` and confirm the per-cell verdicts print.
2. **K1 — drop the bare `--` in the cd branch.** In the cd branch's dirs collection (:657-664), skip the bare `--` separator word (first occurrence) so `cd -- zone_dir` yields exactly one directory operand and `zoneNextCwd`'s `len(dirs) != 1` reset no longer fires for the tracked form; more than one literal word after `--` keeps the reset (bash rejects the two-operand cd — the shell stays put). This mirrors REQ-ZSP-001's seenDashDash discipline in `zonePathCandidates`; the separator itself never becomes a candidate; dynamic words keep the `?dynamic` under-match path unchanged.
3. **K3 — a post-`--` hyphen-leading operand is an operand (D7).** `zoneNextCwd` (:451): the hyphen-prefix reset applies to option territory only — words BEFORE the `--` (real options like `cd -L dir`) and the bare `cd -` (OLDPWD unknown). A post-`--` operand is judged like any literal operand: `-zone` tracks as the relative directory `-zone`, and candidate resolution concatenates it, so the `-zone/` fixture makes `cd -- -zone && rm a.log` deny. Preserve `arg == "-"` and the dynamic-word under-match.
4. **K2 — track an in-project absolute destination.** `zoneNextCwd` (:451 `zoneIsAbs` → reset): when the absolute argument's lexical form is inside the project root — exact-prefix strip of the root followed by `path.Clean`, the REQ-SIPZ-006 lexical semantics (provenance-only, spec D2) — track its root-relative form as the cd destination; an absolute path outside the root, or whose cleaned form escapes the root, keeps the documented reset. The root used for the strip is the same root the zone manifest was loaded against (the walker's existing root context — run-phase picks the plumbing: a root parameter on `zoneNextCwd` or the conversion done in the cd branch before the call). The `..`-escape reset (:459-461) stays in force.
5. TDD invariants: the three reproduction shapes are observed RED (step 1) BEFORE the fixes land (steps 2–4); the landed matrix cells and the frozen `TestProtectedZone` family stay byte-identical; every fix is additive interpretation (REQ-ZSCD-004).

### M2 (High) — cd-class regression matrix

Extend `protected_zone_shell_matrix_test.go`'s cell catalogue with the cd group (decision-index Q3's default: the landed matrix file, no new file). Cells per REQ-ZSCD-005: the post-`--` operand form across `rm`/`cp`/`mv` (`cd -- zone_dir && <verb> <relative>` — zone-covered, deny); the `cd -- -zone` DENY cell (the `addDash` helper, `-zone/` fixture); the in-project absolute-destination cell (deny, fixture root injected into the command); a plain relative preserve cell (`cd zone_dir && rm a.log` — already denied by the landed guard, stays deny, asserts no regression; NOTE: the landed `bare_cd_tracking_control` cell at :127 already pins this shape — M2 counts it as the existing control and does not double-count cd coverage, audit O3); and the landed cd-chain budget cell (t1574 D6) untouched alongside. No allow controls: the outside-root reset is NOT a cell (REQ-ZSCD-003), and no hyphen-leading shape asserts allow (spec D1). Guard: the sweep runs with `-v` (the per-cell output is the observable), and the runner's empty-cell-list `t.Fatal` discipline keeps guarding the whole catalogue so a zero-match selector cannot pass for a sweep.

### M3 (Medium) — scoped verification batch

Slot lease → env-scrubbed single invocation `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -timeout 30m` → `go vet ./internal/hook/` → `golangci-lint run ./internal/hook/...` → flip acceptance.md cells → progress.md §E.2/§E.3. Full-suite judgment stays CI's; the local batch is scoped to `internal/hook` per lane-local rules.

## §G Anti-Patterns

- Do NOT freeze any allow control inside the matrix for a hyphen-leading directory operand — the D7 banned shape; an allow disposition for cd shapes lives outside the matrix only (REQ-ZSCD-003).
- Do NOT weaken an existing denial to make a new test pass; every fix is additive interpretation (REQ-ZSCD-004).
- Do NOT amend the parent SPEC or edit any completed SPEC body; REQ-SIPZ-006 is cited, never modified (spec D2).
- Do NOT reuse the id `REQ-ZSP-009` (retired number; predecessor HISTORY references it) — fresh REQ-ZSCD/AC-ZSCD prefix only (spec D4).
- Do NOT label the outside-root reset as semantically sound — it is an accepted under-match retained on an honest record (spec D3; the `..`-reaching-back shape is named in REQ-ZSCD-003). Hardening it is a future decision the decision-index Q2 operator pin may open — never a claim this SPEC makes for it.
- Do NOT skip the M1 step-1 RED capture (the predecessor's D8 lesson: an AC whose RED observation has no owning plan step is unadopted).
- Do NOT run the full hook package suite without the slot lease; do NOT run `go test ./...` locally.

## §H Cross-References

- Parent: SPEC-SELF-IMPROVE-PROTECTED-ZONE-001 (completed; REQ-SIPZ-006 lexical normalization — provenance-only citation, no requirement of it is amended).
- Predecessor: SPEC-ZONE-SHELL-PARSING-001 (completed; the cd class separated at v0.4.0; its §A.6 and Out of Scope lines name card t1584 as this class's owner).
- Audit evidence: `.moai/reports/t1584/plan-audit-iter2.md` (D7 full text, the SPLIT PROPOSAL this card executes, and the same-revision control-set warning); `.moai/reports/t1584/red-reproduction.md` (the EV-6 probe record — deleted probe, never re-executed; the §2.1 demotion and the M1 step-1 re-establishment own it from here).
- Probe provenance: `.moai/reports/t1574/` (the original card evidence path the probe record was preserved at).
- D8b sweep: measured zero fix-here hits (§B table) — the fold-out was clean; this SPEC's creation completes the handoff.
- Boundary: quote-class internals (`zoneWordText`) → card t1570; the path-layer backslash open question → t1556 reservation (both unchanged from the predecessor's §H).
