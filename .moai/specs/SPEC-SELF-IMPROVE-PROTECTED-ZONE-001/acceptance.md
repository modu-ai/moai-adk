# SPEC-SELF-IMPROVE-PROTECTED-ZONE-001 — Acceptance

> Verification layer. Each `AC-SIPZ-NNN` is a binary-testable Given-When-Then entry; the GEARS obligation lives in `spec.md` §B. Document-level pin: every criterion and every ledger entry below was measured on tree SHA `e497f693608ac7ea45a08b06304dc585e927ff49` (clean tree, binary built from it in this run); a criterion-level pin, where present, wins. The pin binds every criterion that carries none of its own.

## §A Classification

| AC | Covers | Class | Why this class |
|---|---|---|---|
| AC-SIPZ-001 | REQ-SIPZ-001, REQ-SIPZ-002, REQ-SIPZ-004, REQ-SIPZ-005, REQ-SIPZ-011, REQ-SIPZ-012 | **release-blocking** | RED reproduced at base: 20 of 20 rows fail (ledger L-1) |
| AC-SIPZ-002 | REQ-SIPZ-006 | **release-blocking** | RED reproduced at base: 5 of 5 rows fail (L-2) |
| AC-SIPZ-003 | REQ-SIPZ-007 | **release-blocking** | RED reproduced at base: 3 of 3 rows fail (L-3) |
| AC-SIPZ-004 | REQ-SIPZ-002, REQ-SIPZ-009, REQ-SIPZ-010, REQ-SIPZ-013 | **release-blocking** | RED reproduced at base: 4 of 7 rows fail, 3 pass by design (L-4) |
| AC-SIPZ-009 | REQ-SIPZ-001, REQ-SIPZ-003 | **release-blocking** | RED reproduced at base: both manifest files absent, exit 1 (L-9) |
| AC-SIPZ-005 | REQ-SIPZ-008, REQ-SIPZ-016 | regression-guard | Green at base by design (12 of 12, L-5): it asserts nothing about the new work alone and is useful only paired with 001–004 by the mutants in §C |
| AC-SIPZ-006 | REQ-SIPZ-011, REQ-SIPZ-012 | regression-guard | The artifact under test (a manifest-sentinel reason) cannot exist before implementation, so no starting observation can be reproduced |
| AC-SIPZ-007 | REQ-SIPZ-008 | regression-guard | A before/after measurement; the base half is recorded (L-7) but is not a red |
| AC-SIPZ-008 | REQ-SIPZ-010, REQ-SIPZ-014 | regression-guard | The audit log does not exist before implementation |
| AC-SIPZ-010 | REQ-SIPZ-004 | regression-guard | The drift test does not exist before implementation |
| AC-SIPZ-011 | REQ-SIPZ-015 | regression-guard | The liveness check does not exist before implementation; its red is observed by mutation in M3 |
| AC-SIPZ-012 | REQ-SIPZ-006 | regression-guard | Pure-function table; red observed by mutation |
| AC-SIPZ-013 | spec.md §C.9, §E | regression-guard | An absence claim, green at base; carries a positive control |

A regression-guard is **not recorded as a pass on the strength of a base-tree green**. For the Go-test criteria the base-tree observation is the selector hazard in L-6: the run selects zero tests and prints `ok` with exit 0. Each criterion therefore states a swept-count floor and reads `[no tests to run]` as a fail.

**Test naming.** Every test this SPEC adds to the hook package is a named subtest of one top-level test, `TestProtectedZone`, selected with the anchored form `-run '^TestProtectedZone$'`. The subtests are `FileTools`, `Normalization`, `ShellMutation`, `ManifestStates`, `NonRegression`, `DenyReason`, `NoManifestReadForOthers`, `AuditRow`, `BaselineCovered` and `Liveness`; a subtest is swept when its `--- PASS: TestProtectedZone/<Sub> ` line (name followed by a space) is printed. The anchored selector prevents a longer test name sharing the prefix from satisfying the count.

## §B Criteria

### AC-SIPZ-001 — a self-improvement identity cannot modify any zone member through a file tool

**Covers**: maps REQ-SIPZ-001, REQ-SIPZ-002, REQ-SIPZ-004, REQ-SIPZ-005, REQ-SIPZ-011, REQ-SIPZ-012

**Given** the dogfood tree carrying the shipped manifest and the project overlay, **when** the real PreToolUse handler receives a Write or Edit from agent type `harness-learner` for each of the 20 zone-member paths R1–R20 (safety-guard wiring in the settings file, gate-policy and threshold YAML, evaluator profiles, the sync-audit workflow, Go guard code, Go tests, the CI workflow, apply/rollback code, the LSEL applier and allowlist, learning-history logs, budget state, and the two manifest files themselves), **then** every call is denied with the sentinel `HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION` and a reason containing `route=human`, and the judge prints `JUDGE swept=20 expected=20 fail=0` and exits 0.

| Cell | Content |
|---|---|
| RED-now | L-1: judge `-o '^R'` at base exits **1**, 20 `FAIL` rows (19 `got=allow/-`, 1 `got=ask/-`), `JUDGE swept=20 expected=20 fail=20`. |
| Why red | Right reason: none of the 20 paths is in the compiled lists and no manifest exists. R10 returns `ask` from the unrelated critical-config pattern — a decision a subagent cannot answer, so it is red against the required `deny`. |
| Green path | M1 lands the manifests and loader, M2 wires the guard; the same command then prints `JUDGE swept=20 expected=20 fail=0`, exit 0. |
| Mutant probe | §C rows a, d, j. |

### AC-SIPZ-002 — path form cannot be used to step around a protected entry

**Covers**: maps REQ-SIPZ-006

**Given** the legacy-protected `.claude/hooks/` directory, **when** the handler receives a `harness-learner` Write whose path is absolute (P1), has a leading `./` (P2), contains a `docs/../` detour (P3), differs only in letter case (P4), or goes through a symlink into the directory (S1), **then** each is denied with the same legacy sentinel `HARNESS_FROZEN_HOOK_VIOLATION` the plain relative path already earns, and the judge prints `JUDGE swept=5 expected=5 fail=0` and exits 0.

| Cell | Content |
|---|---|
| RED-now | L-2: judge `-o '^(P\|S)'` at base exits **1**, 5 `FAIL` rows all `got=allow/-`, `JUDGE swept=5 expected=5 fail=5`. |
| Why red | Right reason: the existing match is a raw prefix test on the string as received. The relative control C1 on the same directory is denied (L-5), so the path form is the only difference. |
| Green path | M2 flips it by normalizing before comparison; the legacy sentinel is preserved because baseline matches keep their sentinel (REQ-SIPZ-011, REQ-SIPZ-016). |
| Mutant probe | §C rows b, m. |

### AC-SIPZ-003 — a mutating shell command on a zone member is refused

**Covers**: maps REQ-SIPZ-007

**Given** an identity Bash call, **when** the command is `rm .moai/logs/lsel-reject.log` (B1), `echo x > internal/hook/pre_tool.go` (B2) or `sed -i s/a/b/ .moai/config/sections/harness.yaml` (B3), **then** each is denied with the protected-zone sentinel and `route=human`, while the read-only `cat` of the same log (N6) and the same `rm` from a non-identity caller (N3) are still allowed, and the judge prints `JUDGE swept=3 expected=3 fail=0` for `-o '^B'`.

| Cell | Content |
|---|---|
| RED-now | L-3: judge `-o '^B'` at base exits **1**, 3 `FAIL` rows all `got=allow/-`, `JUDGE swept=3 expected=3 fail=3`. |
| Why red | Right reason: the handler has no shell-mutation rule for the identity; the matching Write is also allowed (R14, R6). |
| Green path | M3 adds the rule; B-set flips, N3/N6 stay allowed (AC-SIPZ-005). |
| Mutant probe | §C rows k, d. |

### AC-SIPZ-004 — the manifest is what is read, an unreadable one fails closed, and nothing the identity controls unlocks it

**Covers**: maps REQ-SIPZ-002, REQ-SIPZ-009, REQ-SIPZ-010, REQ-SIPZ-013

**Given** four throwaway project roots — one whose manifest is garbage (M1, M5), one whose valid manifest lists `docs/` (M4), one with no manifest (M2, M3), one with only an overlay listing `docs/` (M6) — and the identity's own process environment carrying `MOAI_BRANCH_GUARD_EXEMPT=1` (E1), **when** the handler receives each call, **then** M1 is denied with the protected-zone sentinel and `manifest=invalid` in the reason; M4 and M6 are denied with `category=probe_docs`; M2 is denied by the legacy `.claude/hooks/` floor; M3 is allowed; M5 (invalid manifest, caller `manager-develop`) is allowed; E1 is denied with `route=human`; and the judge prints `JUDGE swept=7 expected=7 fail=0` for `-o '^(M|E)'`.

| Cell | Content |
|---|---|
| RED-now | L-4: judge `-o '^(M\|E)'` at base exits **1**, 4 `FAIL` rows (M1, M4, M6, E1), `JUDGE swept=7 expected=7 fail=4`; M2, M3, M5 pass at base by design. |
| Why red | Right reason: at base a manifest file of any content is ignored, so a garbage file, a `docs/` entry and an overlay all leave `docs/a.md` allowed and the exemption variable changes nothing because nothing is denied. M2/M3/M5 are the pair that stops an all-matching or everyone-affected mutant. |
| Green path | M1 (loader and validation) and M2 (guard behaviour); the same command prints `fail=0`, exit 0. |
| Mutant probe | §C rows a, e, f, h, i. |

### AC-SIPZ-005 — nothing outside the identity set or the zone changes

**Covers**: maps REQ-SIPZ-008, REQ-SIPZ-016

**Given** the baseline-protected controls C1–C4 and the identity's legitimate surface plus non-identity callers N1–N8 (`.claude/agents/harness/`, `.claude/skills/hns-*`, `.moai/harness/main.md`, `.moai/specs/`, a read-only `cat`, a non-identity Write and Bash to zone paths), **when** the handler receives them, **then** C1–C4 keep their four legacy sentinels and N1–N8 are allowed, and the judge prints `JUDGE swept=12 expected=12 fail=0` for `-o '^(C|N)'`.

| Cell | Content |
|---|---|
| Base observation | L-5: exit **0**, `JUDGE swept=12 expected=12 fail=0` — green at base, so this is a guard, not a gate for new work. |
| Pairing | Fails the all-matching mutant (§C c, k) that AC-001..004 alone would pass, and the nothing-changed mutant is failed by 001..004. |
| Continued firing | Same command each time the CI probe leg runs (AC-SIPZ-011). |

### AC-SIPZ-006 — the denial is a legible human-routing artifact

**Covers**: maps REQ-SIPZ-011, REQ-SIPZ-012

**Given** a zone-member denial, **when** the reason is read, **then** it begins with `HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION:`, names the identity, the project-relative path, `category=<declared name>` and `route=human`, contains no absolute path and no manifest content, is at most 240 bytes, and a baseline match still carries its pre-existing reason text byte for byte; and the hook package's existing subagent-boundary guard still finds no `AskUserQuestion` call in non-test code.

| Cell | Content |
|---|---|
| Verification | Go table subtest `TestProtectedZone/DenyReason` plus the probe's reason column checked by the judge (`route=human` substring on every protected-zone row). |
| Swept floor | Selector `-run '^TestProtectedZone$'` in `./internal/hook/` must print the parent `--- PASS: TestProtectedZone ` line and one `--- PASS: TestProtectedZone/<Sub> ` line for each of the ten subtests (11 lines); `[no tests to run]` is a fail (L-6). |
| Mutant probe | §C row j. |

### AC-SIPZ-007 — the guard costs nothing for anyone it does not apply to

**Covers**: maps REQ-SIPZ-008

**Given** a call from any non-identity agent or a tool other than Write/Edit/Bash, **when** the handler runs, **then** no manifest file is opened (asserted by a read-counting seam in a Go test), and in a paired A/B in one session the median wall time of `moai hook pre-tool` for a non-identity Write differs from the pre-change build's median by at most 50 ms and for an identity Write outside the zone by at most 50 ms.

| Cell | Content |
|---|---|
| Baseline-first | L-7: two base runs, medians 109/117 ms (non-identity) and 110/239 ms (identity, outside zone). The identity median moved 129 ms between runs with no code change, so absolute numbers are not a baseline: the comparison runs both builds back to back, 15 samples each, in one session, and records both. |
| Budget | The hook entry is `"timeout": 10` seconds in the shipped settings; 50 ms is 0.5% of it. |
| Verification | `TestProtectedZone/NoManifestReadForOthers` plus `baseline-latency.sh` run on the base build and the after build. |
| Mutant probe | §C row g. |

### AC-SIPZ-008 — every denial and every degraded evaluation leaves a row

**Covers**: maps REQ-SIPZ-010, REQ-SIPZ-014

**Given** a denial, an absent manifest, or an invalid manifest, **when** the handler decides, **then** one line is appended to `.moai/logs/protected-zone-audit.jsonl` carrying the seven fields of REQ-SIPZ-014, and **when** the append fails (unwritable directory), the decision is unchanged and a stderr notice is printed.

| Cell | Content |
|---|---|
| Verification | `TestProtectedZone/AuditRow` over a temporary directory, including the unwritable-directory case. |
| Swept floor | As AC-006; the three states (deny, absent, invalid) are three named nested subtests of `AuditRow`, each must print its `--- PASS: TestProtectedZone/AuditRow/<State> ` line. |
| Mutant probe | §C row n. |

### AC-SIPZ-009 — the manifest ships neutral and arrives with a build

**Covers**: maps REQ-SIPZ-001, REQ-SIPZ-003

**Given** the template tree, **when** `internal/template/templates/.moai/config/sections/protected-zone.yaml` is read and the template tests run, **then** the file exists, parses, lists the seven required categories, ships `regression_tests` empty, lists both manifest files, and trips none of the template-neutrality audits (no SPEC identifier, requirement token, audit citation, internal date or commit hash); the config completeness audit classifies the section as a dedicated-loader section; and the local dogfood manifest is a superset of the shipped one.

| Cell | Content |
|---|---|
| RED-now | L-9: `ls internal/template/templates/.moai/config/sections/protected-zone.yaml` exits **1**: `No such file or directory`; same for the local `.moai/config/sections/protected-zone.yaml`. |
| Why red | Right reason: the file is the deliverable; the config completeness audit would also fail (`YAML_SECTION_NO_LOADER`) the moment it appears without a registration, which is why M1 registers it in the same commit. |
| Green path | M1: the `ls` exits 0 and `go test ./internal/template/... ./internal/config/...` runs the template and audit tests green after `make build` regenerates the embed. |
| Mutant probe | §C row l (empty sweep), plus a manifest carrying a SPEC identifier must fail the neutrality audit. |

### AC-SIPZ-010 — the manifest cannot silently drift away from the lists it replaces

**Covers**: maps REQ-SIPZ-004

**Given** the compiled baseline and the three hard-coded frozen lists, **when** `TestProtectedZone/BaselineCovered` runs against the dogfood effective zone, **then** every entry — 6 prefixes and 4 basenames in `internal/hook`, 7 in `internal/harness/safety`, 4 in `internal/harness` — is covered, and the test reports the number of entries it swept (21).

| Cell | Content |
|---|---|
| Verification | The test prints the swept count; a count other than 21 or zero fails. |
| Mutant probe | Deleting one manifest entry must fail the test by name; adding a new entry to a compiled list without the manifest must fail it. |

### AC-SIPZ-011 — if the guard stops firing, a build goes red without anyone asking

**Covers**: maps REQ-SIPZ-015

**Given** the repository's CI, **when** any of the four conditions of REQ-SIPZ-015 holds — the PreToolUse matcher group lacks Write, Edit or Bash in the local or template settings; a shipped or dogfood manifest fails to parse; a non-`runtime` entry matches no existing path in its tree; the real handler does not deny a known zone-member input — **then** `TestProtectedZone/Liveness` fails and names which condition.

| Cell | Content |
|---|---|
| Continued-firing answer | The reader learns the guard went quiet from a red CI run on the next pull request or integration push, with the failing condition named in the test output. Nothing needs to be queried. |
| Swept floor | Each of the four conditions is a named nested subtest of `Liveness`; each must print its `--- PASS: TestProtectedZone/Liveness/<Condition> ` line; the dead-entry sweep must report a non-zero count of entries resolved. |
| Mutant probe | Run-phase M3 executes each mutation below and records the observed red in `progress.md` §E.2: (i) remove `Bash` from the template matcher; (ii) corrupt the shipped manifest; (iii) add an entry for a path that does not exist; (iv) make the handler return allow for R6. All four must fail the test. |
| Known blind spot | A binary older than the fix (§F G6 of the spec). |

### AC-SIPZ-012 — path matching behaves the same on every platform CI covers

**Covers**: maps REQ-SIPZ-006

**Given** the pure matching function and a table of Windows-shaped inputs, **when** the table runs on each CI operating system, **then** backslash separators, a drive-letter-prefixed absolute path under the project root, mixed case, a `.\` prefix, and a UNC-style prefix outside the root resolve as the table states (the first four to the zone entry, the last to outside-project), identically on darwin, linux and windows.

| Cell | Content |
|---|---|
| Verification | `TestProtectedZone/Normalization`, run by the existing CI matrix; the table is data, not conditional on `runtime.GOOS`. |
| Not covered | 8.3 short names, junctions, alternate data streams — disclosed in `spec.md` §E, not tested. |
| Mutant probe | §C row m. |

### AC-SIPZ-013 — the work stays inside its scope

**Covers**: maps REQ-SIPZ-016

**Given** the run-phase commits, **when** the changed-path list against the base SHA is read, **then** it contains no file under `.claude/agents/`, `internal/template/templates/.claude/agents/`, no change to either `settings.json` matcher, and no edit to `.moai/hooks/lsel-apply.sh`.

| Cell | Content |
|---|---|
| Command | `git diff --name-only e497f693608ac7ea45a08b06304dc585e927ff49 -- .claude/agents internal/template/templates/.claude/agents .moai/hooks` |
| Positive control | The same form with `-- internal/hook` must be non-empty after M2; an empty result for the control means the command measured nothing and the absence is "unmeasured", not "clean". |
| Base observation | Empty at base, trivially — which is why this is a regression-guard. |

## §C Mutation table

Each row is a mutant that satisfies a one-sided reading of the criteria and violates the requirement; the paired criterion kills it. Run-phase executes every row and records the observed red in `progress.md` §E.2.

| # | Mutant | Killed by |
|---|---|---|
| a | Zone list hard-coded in Go; the manifest file is ignored | AC-004 M4, M6 (a project-only entry `docs/` is denied only if the manifest is read) |
| b | Raw-prefix match with no normalization (today's behaviour) | AC-002 P1–P4, S1 |
| c | Deny every identity Write | AC-005 N4, N5, N7, N8 |
| d | Guard never emits the new sentinel for manifest-only entries | AC-001 all 20 rows |
| e | Invalid manifest treated as absent (fail open) | AC-004 M1 |
| f | Invalid manifest denies every caller | AC-004 M5 |
| g | Manifest read before the identity gate | AC-007 read-counter test |
| h | A bypass variable is honoured | AC-004 E1 |
| i | Overlay replaces instead of adds | `TestProtectedZone/ManifestStates` overlay-cannot-narrow subtest; AC-004 M6 pairs it |
| j | Sentinel present, `route=human` or `category=` missing | AC-001 judge reason column; AC-006 |
| k | Shell rule matches read-only verbs | AC-003 / AC-005 N6 |
| l | Liveness or dead-entry sweep passes with zero entries swept | AC-009 and AC-011 swept-count floors |
| m | Case folding only where the host is case-insensitive | AC-002 P4 on the linux CI leg; AC-012 |
| n | A failed audit append changes the decision | AC-008 unwritable-directory subtest |
| o | PreToolUse matcher drops `Bash` or `Edit` | AC-011 (i) |

## §D Counts, baseline-first note, cross-platform, definition of done

**Swept-set counts the judge must print after implementation** (the probe is one table, so the whole-table form is the release check): `JUDGE swept=47 expected=47 fail=0`; by group R=20, P/S=5, B=3, M/E=7, C/N=12. A selection that matches no row is itself a failure (`empty-selection`).

**Baseline-first.** The base-tree measurements (`evidence/red-probe-base-e497f6936.tsv`, `judge-base-e497f6936.txt`, `latency-base-e497f6936.txt`) are plan-phase artifacts committed ahead of any run-phase commit, so the commit graph — not a commit message — witnesses that the red preceded the change (`verification-claim-integrity.md` §2.3). Re-measure on rebase; do not re-quote.

**Cross-platform.** The matching function is pure string work and the same table runs on every CI leg. The shell rule uses the existing shell-segment splitter and applies to Bash only; PowerShell is not covered (spec §E). Relative paths are judged against the process working directory exactly as the existing file-access check does.

**Edge cases the unit tests cover:** a Korean path in NFD against an NFC entry; an empty or directory `file_path`; a manifest with a duplicate entry, a middle wildcard, an entry containing `..`, an unknown top-level `version`, or an empty category (all `manifest=invalid` except duplicate and empty category, which are valid); an overlay that is invalid while the shipped manifest is valid (fail closed); a deny reason for a path longer than the 240-byte cap (truncated, sentinel intact).

**Quality gates.** `go vet ./internal/hook/... ./internal/config/... ./internal/template/...`, `golangci-lint run` on the changed packages, `go test -race ./internal/hook/...` (the guard touches shared state only through the audit append), `make build` so the embed regenerates, and TRUST 5 sync-audit as usual. Full-suite verdict is CI's.

**Definition of done.** AC-SIPZ-001..004 and 009 green with the RED→GREEN pair recorded; AC-005..013 green with swept counts shown; every §C mutant executed and its red recorded; `moai spec lint --strict SPEC-SELF-IMPROVE-PROTECTED-ZONE-001` clean; no edit outside the scope of AC-SIPZ-013.

## §E Evidence ledger

All entries: tree SHA `e497f693608ac7ea45a08b06304dc585e927ff49`; binary built from that clean tree with `go build ./cmd/moai` in this run; commands are repository-root relative. Raw files are under `.moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/`.

```ledger
L-1
  cmd:    bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh -o '^R' .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/red-probe-base-e497f6936.tsv
  re-measure (live, builds a binary): bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh -o '^R'
  stdout:
    FAIL R1 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R2 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R3 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R4 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R5 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R6 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R7 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R8 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R9 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R10 got=ask/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R11 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R12 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R13 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R14 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R15 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R16 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R17 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R18 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R19 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL R20 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    JUDGE swept=20 expected=20 fail=20
  exit:   1

L-2
  cmd:    bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh -o '^(P|S)' .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/red-probe-base-e497f6936.tsv
  stdout:
    FAIL P1 got=allow/- want=deny/HARNESS_FROZEN_HOOK_VIOLATION
    FAIL P2 got=allow/- want=deny/HARNESS_FROZEN_HOOK_VIOLATION
    FAIL P3 got=allow/- want=deny/HARNESS_FROZEN_HOOK_VIOLATION
    FAIL P4 got=allow/- want=deny/HARNESS_FROZEN_HOOK_VIOLATION
    FAIL S1 got=allow/- want=deny/HARNESS_FROZEN_HOOK_VIOLATION
    JUDGE swept=5 expected=5 fail=5
  exit:   1

L-3
  cmd:    bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh -o '^B' .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/red-probe-base-e497f6936.tsv
  stdout:
    FAIL B1 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B2 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B3 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    JUDGE swept=3 expected=3 fail=3
  exit:   1

L-4
  cmd:    bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh -o '^(M|E)' .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/red-probe-base-e497f6936.tsv
  stdout:
    FAIL M1 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL M4 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL M6 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL E1 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    JUDGE swept=7 expected=7 fail=4
  exit:   1

L-5
  cmd:    bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh -o '^(C|N)' .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/red-probe-base-e497f6936.tsv
  stdout:
    JUDGE swept=12 expected=12 fail=0
  exit:   0

L-6
  cmd:    go test -count=1 -timeout 30m -run '^TestProtectedZone$' -v ./internal/hook/
  stdout:
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/hook	0.573s [no tests to run]
  exit:   0
  reading: zero tests swept; an exit of 0 and the word PASS are not a pass for any AC that names this selector.

L-7
  cmd:    bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/baseline-latency.sh <binary built from the tree> 15
  stdout (run 1, then run 2; the file evidence/latency-base-e497f6936.txt holds both):
    identity-write-outside-zone	n=15	min=104	median=110	max=204
    identity-write-inside-zone	n=15	min=72	median=78	max=102
    non-identity-write	n=15	min=104	median=109	max=143
    identity-write-outside-zone	n=15	min=116	median=239	max=308
    identity-write-inside-zone	n=15	min=83	median=92	max=118
    non-identity-write	n=15	min=108	median=117	max=155
  exit:   0

L-9
  cmd:    ls internal/template/templates/.moai/config/sections/protected-zone.yaml
  stdout: (empty)
  stderr: ls: internal/template/templates/.moai/config/sections/protected-zone.yaml: No such file or directory
  exit:   1
  cmd:    ls .moai/config/sections/protected-zone.yaml
  stdout: (empty)
  stderr: ls: .moai/config/sections/protected-zone.yaml: No such file or directory
  exit:   1
```
