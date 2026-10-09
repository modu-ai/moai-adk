---
id: SPEC-ZONE-SHELL-CD-TRACKING-001
title: "Acceptance criteria — cd destination tracking (post-`--` operand incl. hyphen-leading, in-project absolute), cd-class regression cells"
version: "0.1.0"
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
tier: M
---

# SPEC-ZONE-SHELL-CD-TRACKING-001 — acceptance

## §A Scenario Conventions

Every criterion below is a Given-When-Then scenario, binary-testable: the Given names the tree state, the When names one runnable command, the Then names one observable output. Verification commands use the plain single-invocation form. The guard under test is `checkProtectedZoneShell` via the landed test harness (`newZoneRoot`/`zoneShippedDoc`/`zoneTestHandler` fixture family); a `deny` is `decision="deny"`, an `allow` is `decision="allow"` with an empty category.

## §B Evidence Ledger

Four-element rule (verification-completeness §2.1): a release-blocking RED cell carries the command, its verbatim stdout, its exit code, and the tree SHA — and RED must be red for the stated reason. Where a cited RED cannot be re-executed on the current tree, the criterion loses release-blocking eligibility, is classified regression-guard, and is NOT recorded as a pass (the undecidable disposition).

**EV-ZSCD-001 (inherited, DEMOTED — regression-guard-pending).** The predecessor's EV-6 probe: both cd shapes observed `decision="allow"` at tree `b9ef003808da2ac1dfe42e5374d5bde4302f46b3`, 2026-10-07 (verbatim record: `.moai/reports/t1584/red-reproduction.md`, inherited from `.moai/reports/t1574/`). The probe file (`internal/hook/zz_probe_cd_test.go`) was deleted after observation and never re-executed — this evidence CANNOT be re-executed on the current tree, so it is the demoted inheritance, never a pass. Its verbatim record:

```
=== RUN   TestZZProbeCdTracking
    zz_probe_cd_test.go:21: cmd="cd -- zone_dir && rm a.log" decision="allow" reason=""
    zz_probe_cd_test.go:21: cmd="cd /private/var/folders/.../TestZZProbeCdTracking1443033861/001/zone_dir && rm a.log" decision="allow" reason=""
--- PASS: TestZZProbeCdTracking (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.604s
```

**EV-ZSCD-002 (RED-now — TO BE CAPTURED at M1 step 1).** The three-shape reproduction set run against the PRE-FIX guard at the M1-entry tree (pin recorded at capture time), each shape expected in the reset state (`decision="allow"`): `cd -- zone_dir && rm a.log`, `cd <fixture-root>/zone_dir && rm a.log`, `cd -- -zone && rm a.log` (against a `-zone/` fixture). Until this capture exists with its four elements, AC-ZSCD-001..003 stay regression-guard-pending and no green claim is recorded against them. The capture is the ONLY re-establishment path (spec D5; plan M1 step 1 owns it — the predecessor's D8 lesson).

## §D AC Matrix (two-cell discipline: RED-now + green path)

| AC | Claim | RED-now | Green path | Classification |
|----|-------|---------|------------|----------------|
| AC-ZSCD-001 | The three-shape cd reproduction set is captured verbatim against the pre-fix guard, then flips to deny | EV-ZSCD-002 (to be captured at M1 step 1; the EV-ZSCD-001 inheritance is §2.1-demoted) | M1 fixes flip all three cells to deny; verbatim post-fix output recorded | release-blocking once captured; regression-guard-pending until then |
| AC-ZSCD-002 | `cd -- -zone && rm a.log` asserts DENY against a `-zone/` fixture | captured at M1 step 1 (pre-fix: allow — the reset) | M1 K2 fix flips it; permanent cell in the M2 cd group | release-blocking |
| AC-ZSCD-003 | The in-project absolute-destination cell asserts the destination is tracked (protected deletion denied) | captured at M1 step 1 (pre-fix: allow — the reset) | M1 K3 fix flips it; permanent cell in the M2 cd group | release-blocking |
| AC-ZSCD-004 | The cd matrix group exists in `TestProtectedZoneShellParsingMatrix`, non-empty, `-v` per-cell output, empty-list-fails intact | EV-ZSCD-003: `go test ./internal/hook/ -run '^TestProtectedZoneShellParsingMatrix$' -count=1 -v` at the pre-work tree prints no `cd_` cell — the cd class is unswept (measured baseline to pin at M1 entry) | M2 lands the cd group; the runner's empty-cell-list `t.Fatal` keeps guarding the catalogue | release-blocking |
| AC-ZSCD-005 | The landed `TestProtectedZone` family and the landed parsing matrix pass unchanged | green-now baseline (preserved-behavior — the two-cell pair's RED-now is the pre-work green pin, recorded at M1 entry) | stays green through M1–M3; `git diff` over the landed test files shows no assertion-line edit | release-blocking (preserved-behavior) |
| AC-ZSCD-006 | Scoped verification batch green: hook package suite, `go vet ./internal/hook/`, `golangci-lint run ./internal/hook/...` | baseline observed at M1 entry | M3 batch | release-blocking (quality-gate) |

## §D.1 Given-When-Then Scenarios

**AC-ZSCD-001** — RED re-establishment and flip.
- **Given** the pre-fix guard at the M1-entry tree (SHA pinned at capture), when the three-shape reproduction set runs with `-v` and per-shape verdict output, then each shape prints `decision="allow"` (the reset state — red for the STATED reason: the `:447` two-operand reset and the `:451` hyphen/absolute resets, never a selector miss), the commands' verbatim stdout and exit codes are recorded in acceptance.md §B, and after the M1 fixes the same three shapes print `decision="deny"`.

**AC-ZSCD-002** — the D7 hyphen-leading operand cell denies.
- **Given** a fixture root whose zone manifest covers `-zone/`, when `cd -- -zone && rm a.log` is judged, then the verdict is `deny` (category `probe_zone`) — post-fix; pre-fix the same command answers `allow` (captured at M1 step 1). A mutant that restores the hyphen-prefix reset for post-`--` operands fails this cell.

**AC-ZSCD-003** — the in-project absolute destination is tracked.
- **Given** a fixture root whose zone manifest covers `zone_dir/`, when `cd <fixture-root>/zone_dir && rm a.log` is judged (the absolute destination built from the fixture root), then the verdict is `deny` — the destination is tracked under its root-relative form; pre-fix the same command answers `allow`. A mutant that keeps the `zoneIsAbs` reset for in-project absolutes fails this cell; an outside-root absolute destination (`cd /tmp/outside && rm a.log`) stays `allow` and is NOT a matrix cell (REQ-ZSCD-003).

**AC-ZSCD-004** — the cd regression group is swept, never silent.
- **Given** the landed sweep runner `TestProtectedZoneShellParsingMatrix` (protected_zone_shell_matrix_test.go:205), when the suite runs with `-v -count=1`, then the catalogue contains the cd group (post-`--` operand across `rm`/`cp`/`mv`, the `cd -- -zone` deny cell, the absolute-destination cell, the plain-relative preserve cell), every zone-covered cd cell prints its own `--- PASS` line with a deny verdict, and the runner's empty-cell-list `t.Fatal` still fails a swept-empty catalogue (the `[no tests to run]` shape can never pass for a sweep).

**AC-ZSCD-005** — preserved behavior.
- **Given** tree state with this SPEC's fixes landed, when `go test ./internal/hook/ -run '^TestProtectedZone$' -count=1` and `go test ./internal/hook/ -run '^TestProtectedZoneShellParsingMatrix$' -count=1` run, then both pass, and `git diff` over the landed test files shows no assertion-line edit on pre-existing cells — no already-covered form changed verdict (REQ-ZSCD-004); the landed cd-chain budget cell and the bare-cd-only rule are untouched.

**AC-ZSCD-006** — scoped verification batch.
- **Given** the completed M1–M2 work, when the env-scrubbed batch runs (slot lease → `go test ./internal/hook/ -timeout 30m` → `go vet ./internal/hook/` → `golangci-lint run ./internal/hook/...`), then all three exit 0 with no NEW findings vs the M1-entry baseline, and the outputs are recorded verbatim in progress.md §E.2.

## §D.2 Edge Cases

- `cd -- dir1 dir2` (two operands after the separator): bash rejects the cd, the shell stays put — the reset is CORRECT and stays; the matrix may pin it as a preserve observation, never as an allow control for a hyphen-leading shape.
- `cd -` (OLDPWD): stays reset (the guard cannot know OLDPWD); not a matrix cell.
- `cd -- -L dir`-shaped option-before-separator forms: the hyphen reset stays in force BEFORE the `--`; only post-`--` operands become ordinary operands.
- `cd <root>/../<root>/zone_dir`-shaped absolutes: cleaned inside the root → tracked (REQ-ZSCD-002 lexical rule); cleaned outside → reset.
- Dynamic destination (`cd $d && rm a.log`): keeps the `?dynamic` under-match (parent §C.6) — unchanged by this SPEC.

## §D.3 Quality Gates

TRUST 5: Tested (AC-ZSCD-001..004, hook suite green); Readable/Unified (English round-style comments matching file density, gofmt/golangci-lint clean — AC-ZSCD-006); Secured (fail-closed direction, the D7 disposition — AC-ZSCD-002); Trackable (Conventional Commits referencing this SPEC ID, card t1584 in commit messages).

## §E Traceability

| AC | REQ | Evidence |
|----|-----|----------|
| AC-ZSCD-001 | REQ-ZSCD-001, REQ-ZSCD-002, REQ-ZSCD-003 | EV-ZSCD-001 (demoted), EV-ZSCD-002 (M1 step 1) |
| AC-ZSCD-002 | REQ-ZSCD-001, REQ-ZSCD-005 | EV-ZSCD-002 shape 3 + the M2 cell |
| AC-ZSCD-003 | REQ-ZSCD-002, REQ-ZSCD-005 | EV-ZSCD-002 shape 2 + the M2 cell |
| AC-ZSCD-004 | REQ-ZSCD-005 | EV-ZSCD-003 + the M2 runner run |
| AC-ZSCD-005 | REQ-ZSCD-004 | the M1-entry green pin + M3 re-run |
| AC-ZSCD-006 | all | the M3 batch |

## §F Indirect Verification

- REQ-ZSCD-003 (outside-root reset kept) is verified indirectly: AC-ZSCD-003's scenario asserts the outside-root control stays allow OUTSIDE the matrix, and AC-ZSCD-005's family re-run pins that no landed allow outside-root behavior flipped.
- REQ-ZSCD-004 (preserved behavior) is verified indirectly by AC-ZSCD-005's no-assertion-edit diff plus the family/matrix green.

## §G Closure Gates

- Definition of Done: AC-ZSCD-001..006 all GREEN with verbatim evidence; the three RED captures exist with four elements each (command, stdout, exit, tree SHA); no FOUNDER-blocked decision remains open in decision-index.md; progress.md §E.2/§E.3 populated; the D8b sweep conclusion (plan §B — fix-here 0) re-stated in the completion report.
- Forward-looking check: the cd cells' cell names are `cd_`-prefixed so a future selector audit can count the cd class without parsing the catalogue.
