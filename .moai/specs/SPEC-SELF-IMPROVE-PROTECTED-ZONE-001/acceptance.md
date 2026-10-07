# SPEC-SELF-IMPROVE-PROTECTED-ZONE-001 — Acceptance

> Verification layer. Each `AC-SIPZ-NNN` is a binary-testable Given-When-Then entry; the GEARS obligation lives in `spec.md` §B. Document-level pin: every criterion and every ledger entry below was measured on tree SHA `e497f693608ac7ea45a08b06304dc585e927ff49`. The later commits change no code this SPEC reads or edits: `git diff --stat e497f693608ac7ea45a08b06304dc585e927ff49 HEAD -- internal/hook internal/config internal/harness internal/spec cmd .claude/settings.json internal/template/templates internal/template/template_neutrality_audit_test.go` printed nothing at HEAD `41127e8caded5c54507414fc6dd88fbf28d9517b` (the unscoped `-- internal cmd` form lists eleven unrelated files under `internal/cli`, `internal/web` and `internal/template/model_policy.go`, which is the positive control that the scoped command can print). After the 0.3.0 probe-fixture change (D21) the probe and judge were re-run with a binary built from HEAD `41127e8ca` with `go build ./cmd/moai`; the recorded TSV and both judge outputs came out byte-identical to the base recording, so they were not rewritten. A criterion-level pin, where present, wins; the pin binds every criterion that carries none of its own.

## §A Classification

Every release-blocking criterion carries **two commands**, and they are not interchangeable. The **live** form (`judge-probe.sh -o '<re>'` with no TSV argument) builds a binary from the tree, drives the real handler and judges what it just observed — it is the form whose output changes when the implementation lands, and the one every green-path cell names. The **replay** form judges the recorded base-tree output (`evidence/red-probe-base-e497f6936.tsv`); it is evidence of what the base tree did and can never flip. At base the two print identical rows (`evidence/judge-base-live.txt` and `judge-base-replay.txt` differ in nothing but the mode line).

| AC | Covers | Class | Why this class |
|---|---|---|---|
| AC-SIPZ-001 | REQ-SIPZ-001, REQ-SIPZ-002, REQ-SIPZ-004, REQ-SIPZ-005, REQ-SIPZ-012 | **release-blocking** | RED reproduced live at base: 21 of 21 rows fail (L-1) |
| AC-SIPZ-002 | REQ-SIPZ-006 | **release-blocking** | RED reproduced live at base: 7 of 7 rows fail (L-2) |
| AC-SIPZ-003 | REQ-SIPZ-007 | **release-blocking** | RED reproduced live at base: 13 of 13 rows fail (L-3) |
| AC-SIPZ-004 | REQ-SIPZ-002, REQ-SIPZ-009, REQ-SIPZ-010, REQ-SIPZ-013 | **release-blocking** | RED reproduced live at base: 8 of 11 rows fail, 3 pass by design (L-4) |
| AC-SIPZ-009 | REQ-SIPZ-001, REQ-SIPZ-003 | **release-blocking** | RED reproduced at base: both manifest files absent, exit 1 (L-9) |
| AC-SIPZ-005 | REQ-SIPZ-008, REQ-SIPZ-016 | regression-guard | Green at base by design (13 of 13, L-5): asserts nothing about the new work alone; useful only paired with 001–004 by the mutants in §C |
| AC-SIPZ-006 | REQ-SIPZ-011 | regression-guard | The live probe rows LP1, LP2 are red at base (L-6a) but for a different reason than the property: nothing is denied at all, so truncation is never exercised. The property is observed by mutation in M2 |
| AC-SIPZ-007 | REQ-SIPZ-008 | regression-guard | A before/after measurement; the base half (A/A control and A/B) is recorded (L-7) and is not a red |
| AC-SIPZ-008 | REQ-SIPZ-010, REQ-SIPZ-014 | regression-guard | The audit log does not exist before implementation |
| AC-SIPZ-010 | REQ-SIPZ-004 | regression-guard | The drift test does not exist before implementation; R21 is its live red-now sentinel for the one baseline entry no manifest list carried |
| AC-SIPZ-011 | REQ-SIPZ-015 | regression-guard | The liveness check does not exist before implementation; its red is observed by mutation in M3 |
| AC-SIPZ-012 | REQ-SIPZ-006 | regression-guard | Pure-function table; red observed by mutation |
| AC-SIPZ-013 | REQ-SIPZ-016 | regression-guard | An absence claim, green at base; carries a positive control |

A regression-guard is **not recorded as a pass on the strength of a base-tree green**. For the Go-test criteria the base-tree observation is the selector hazard in L-8: the run selects zero tests and prints `ok` with exit 0. Each criterion therefore states a swept-count floor and reads `[no tests to run]` as a fail.

**Test naming.** Every test this SPEC adds to a package is a named subtest of one top-level test, `TestProtectedZone`, selected with the anchored form `-run '^TestProtectedZone$'`. In `./internal/hook/` the subtests are `FileTools`, `Normalization`, `ShellMutation`, `ManifestStates`, `NonRegression`, `DenyReason`, `NoManifestReadForOthers`, `AuditRow`, `BaselineCovered` and `Liveness`; a subtest is swept when its `--- PASS: TestProtectedZone/<Sub> ` line (name followed by a space) is printed. In `./internal/config/` they are `Load`, `Overlay` and `Validation`; in `./internal/template/`, `Shipped`, `Neutrality` and `SupersetOfShipped`.

## §B Criteria

### AC-SIPZ-001 — a self-improvement identity cannot modify any zone member through a file tool

**Covers**: maps REQ-SIPZ-001, REQ-SIPZ-002, REQ-SIPZ-004, REQ-SIPZ-005, REQ-SIPZ-012

**Given** the dogfood tree carrying the shipped manifest and the project overlay, **when** the real PreToolUse handler receives a Write or Edit carrying agent type `harness-learner` for each of the 21 zone-member paths R1–R21 (safety-guard wiring in the settings file, gate-policy and threshold YAML, evaluator profiles, the sync-audit workflow, Go guard code, Go tests, the CI workflow, apply/rollback code, the LSEL applier and allowlist, learning-history logs, budget state, the two manifest files themselves, and the `/moai` skill directory), **then** every call is denied, R1–R20 with the sentinel `HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION` and a reason containing `category=`, `route=human` and `next=return-blocker-report`, and the live judge prints `JUDGE swept=21 expected=21 fail=0` and exits 0. This is the routed class of denial: R1–R20 are matched by the manifest only. The baseline-matched controls C1–C4 are the other side of the scope boundary (spec §C.7) and are asserted in AC-SIPZ-005.

| Cell | Content |
|---|---|
| RED-now (live) | L-1: `bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh -o '^R'` at base exits **1**, 21 `FAIL` rows (20 `got=allow/-`, R10 `got=ask/-`), `JUDGE swept=21 expected=21 fail=21`. The replay command on the recorded TSV prints the same rows. |
| Why red | Right reason: none of the 21 paths is in the compiled lists and no manifest exists. R10 returns `ask` from an unrelated critical-config pattern — a decision a subagent cannot answer, so it is red against the required `deny`. R21 is allowed because the compiled hook list has `.claude/skills/moai-` and not `.claude/skills/moai/`. |
| Green path | M1 lands the manifests and loader, M2 wires the guard. **The live command, unchanged, then prints `JUDGE swept=21 expected=21 fail=0` and exits 0**; the replay command keeps printing `fail=21` forever and is not the green check. |
| Mutant probe | §C rows a, d, j, p. |

### AC-SIPZ-002 — path form cannot be used to step around a protected entry

**Covers**: maps REQ-SIPZ-006

**Given** the legacy-protected `.claude/hooks/` directory, **when** the handler receives a `harness-learner` Write whose path is absolute (P1), has a leading `./` (P2), contains a `docs/../` detour (P3), differs only in letter case (P4), is the physical path while the project root is reached through a symlink (P5), is a symlinked path while the root is physical (P6), or goes through a symlink into the directory (S1), **then** each is denied with the same legacy sentinel `HARNESS_FROZEN_HOOK_VIOLATION` the plain relative path already earns, and the live judge prints `JUDGE swept=7 expected=7 fail=0` and exits 0. P5 and P6 are the rows the existing check's root-resolution asymmetry would fail: a root reached through a symlink (macOS `/var` to `/private/var` is the common case) makes an absolute in-project path look outside the project, which is the bypass this criterion exists to close, and the fixture does not avoid it.

| Cell | Content |
|---|---|
| RED-now (live) | L-2: `bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh -o '^(P\|S)'` at base exits **1**, 7 `FAIL` rows all `got=allow/-`, `JUDGE swept=7 expected=7 fail=7`. |
| Why red | Right reason: the existing match is a raw prefix test on the string as received. The relative control C1 on the same directory is denied (L-5), so the path form is the only difference. |
| Green path | M2 normalizes before comparison and resolves root and target by one function; the live command prints `fail=0`, exit 0. Legacy sentinels survive because baseline matches keep them (REQ-SIPZ-011, REQ-SIPZ-016). |
| Mutant probe | §C rows b, m, q. |

### AC-SIPZ-003 — every mutating shell form on a zone member is refused

**Covers**: maps REQ-SIPZ-007

**Given** an identity Bash call, **when** the command uses each of the thirteen forms — `rm` (B1), `unlink` (B2), `mv` (B3), `cp` (B4), `tee` (B5), `truncate` (B6), `sed -i` (B7), `>` (B8), `>>` (B9), `git rm` (B10), `git checkout` (B11), `git restore` (B12), `git apply` (B13) — against a zone member, **then** each is denied with the protected-zone sentinel and `route=human`, while the read-only `cat` of the same log (N6) and the same `rm` from a non-identity caller (N3) are still allowed, and the live judge prints `JUDGE swept=13 expected=13 fail=0` for `-o '^B'`. A `ShellMutation` subtest table carries the same thirteen forms and its swept count must equal 13.

| Cell | Content |
|---|---|
| RED-now (live) | L-3: `bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh -o '^B'` at base exits **1**, 13 `FAIL` rows all `got=allow/-`, `JUDGE swept=13 expected=13 fail=13`. |
| Why red | Right reason: the handler has no shell-mutation rule for the identity; the matching Write is also allowed (R14, R6). |
| Green path | M3 adds the rule; the live command prints `fail=0`, exit 0; N3/N6 stay allowed (AC-SIPZ-005). |
| Mutant probe | §C rows k, d, r. |

### AC-SIPZ-004 — the manifest is what is read, an unreadable one fails closed, the overlay only adds, and nothing the identity controls unlocks it

**Covers**: maps REQ-SIPZ-002, REQ-SIPZ-009, REQ-SIPZ-010, REQ-SIPZ-013

**Given** throwaway project roots — garbage manifest (MS1, MS5), valid manifest listing `docs/` (MS4), no manifest (MS2, MS3), overlay only (MS6), a valid base listing `base_dir/` plus an overlay listing `docs/` (MS7), a valid base plus an overlay carrying an unknown `exclude:` key (MS8); every valid manifest at the shipped path (MS4, MS7, MS8) declares all seven required categories with empty lists plus its probe category, because a shipped-path file lacking any of the seven is invalid (REQ-SIPZ-001) — and three bypass attempts against `internal/hook/pre_tool.go` or a log: `MOAI_BRANCH_GUARD_EXEMPT=1` in the hook's environment (E1), extra tool-input fields `override` and `protected_zone` (E2), and a command-text prefix plus a trailing comment token (E3), **when** the handler receives each call, **then** MS1 and MS8 are denied with `manifest=invalid`, `route=human`, `next=return-blocker-report` and the failing file's project-relative path in the reason (`.moai/config/sections/protected-zone.yaml` for MS1, `.moai/project/protected-zone.yaml` for MS8 — a reason naming the other file or none fails); MS4 and MS6 are denied with `category=probe_docs`; MS7 is denied with `category=probe_base` (the overlay did not replace the base); MS2 is denied by the legacy `.claude/hooks/` floor; MS3 is allowed; MS5 (invalid manifest, caller `manager-develop`) is allowed; E1–E3 are denied with `route=human`; and the live judge prints `JUDGE swept=11 expected=11 fail=0` for `-o '^(MS|E)'`.

| Cell | Content |
|---|---|
| RED-now (live) | L-4: `bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh -o '^(MS\|E)'` at base exits **1**, 8 `FAIL` rows (MS1, MS4, MS6, MS7, MS8, E1, E2, E3), `JUDGE swept=11 expected=11 fail=8`; MS2, MS3, MS5 pass at base by design. |
| Why red | Right reason: at base a manifest file of any content is ignored, so a garbage file, a `docs/` entry, an overlay and an overlay with an unknown key all leave the path allowed, and the bypass rows are red only because nothing is denied to bypass. MS2/MS3/MS5 are the pair that stops an all-matching or everyone-affected mutant. |
| Green path | M1 (loader and validation) and M2 (guard behaviour) flip the MS rows and E1, E2; **E3 is a Bash call that only the M3 shell rule can deny**, so the live `-o '^(MS\|E)'` prints `JUDGE swept=11 expected=11 fail=0` and exits 0 at the end of M3, not M2 (at the end of M2 it still prints one `FAIL E3`). |
| Mutant probe | §C rows a, e, f, h, i, s. |

### AC-SIPZ-005 — nothing outside the identity set or the zone changes

**Covers**: maps REQ-SIPZ-008, REQ-SIPZ-016

**Given** the baseline-protected controls C1–C4 and the identity's legitimate surface plus non-identity callers N1–N9 (`.claude/agents/harness/`, `.claude/skills/hns-*`, `.moai/harness/main.md`, `.moai/specs/`, `docs/a.md`, a read-only `cat`, a non-identity Write and Bash to zone paths), **when** the handler receives them, **then** C1–C4 keep their four legacy sentinels with their legacy reason text and carry no `route=` or `category=` field (the judge asserts both substrings absent), and N1–N9 are allowed, and the live judge prints `JUDGE swept=13 expected=13 fail=0` for `-o '^(C|N)'`.

| Cell | Content |
|---|---|
| Base observation | L-5: exit **0**, `JUDGE swept=13 expected=13 fail=0` — green at base, so this is a guard, not a gate for new work. |
| Pairing | Fails the all-matching mutant (§C c, k) that AC-SIPZ-001..004 alone would pass, and the mutant that appends routing fields to a baseline denial (§C w). Negative control observed: a copy of the recorded TSV with `route=human` appended to C1's reason makes `-o '^C'` print `FAIL C1 reason-has-forbidden 'route=human'` and exit 1. |
| Continued firing | The same command runs in the CI probe leg (AC-SIPZ-011). |

### AC-SIPZ-006 — the routing fields survive a long path

**Covers**: maps REQ-SIPZ-011

**Given** a zone-member denial for a very long ASCII path (LP1, 220 characters) and a very long non-ASCII path (LP2, 90 three-byte characters), **when** the reason is read, **then** it carries, in this order, the sentinel, the identity, `category=<declared name>`, `route=human` and `next=return-blocker-report`, followed by the project-relative path as the only truncated field; the whole reason is at most 240 bytes; it contains no absolute path and no manifest content; a baseline match still carries its pre-existing reason text byte for byte; and the hook package's existing subagent-boundary guard still finds no `AskUserQuestion` call in non-test code.

| Cell | Content |
|---|---|
| Observation | Live `bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh -o '^LP'` prints `fail=0` after M2; at base L-6a shows 2 `FAIL` rows (nothing is denied). The judge requires `route=human`, `category=` and `next=return-blocker-report` in the reason of every protected-zone row, R1–R20 and B1–B13 included, so a deny that lacks the instruction fails there too. |
| Verification | `TestProtectedZone/DenyReason`: a table over short, long, multi-byte and 32-byte-category inputs asserting field order, the 240-byte bound, and that the three marker tokens are present in every row; a truncation mutant (cut the whole reason) must fail it. |
| Swept floor | Selector `-run '^TestProtectedZone$'` in `./internal/hook/` must print the parent `--- PASS: TestProtectedZone ` line and one `--- PASS: TestProtectedZone/<Sub> ` line for each of the ten subtests (11 lines); `[no tests to run]` is a fail (L-8). |
| Mutant probe | §C rows j, t. |
| Residual | That a denied subagent then returns a blocker report is unobserved (spec §F G8); this criterion observes the instruction, not the obedience. |

### AC-SIPZ-007 — the guard costs nothing measurable for anyone it does not apply to

**Covers**: maps REQ-SIPZ-008

**Given** a call from any non-identity agent or a tool other than Write/Edit/Bash, **when** the handler runs, **then** no manifest file is opened (asserted by a read-counting seam in `TestProtectedZone/NoManifestReadForOthers`), and the interleaved paired A/B of `baseline-latency.sh` between the pre-change build (A) and the after build (B) prints `verdict=PASS` for all three classes.

| Cell | Content |
|---|---|
| Instrument | Each iteration times A then B back to back, 21 pairs; the statistic is the median of the per-iteration differences, which cancels drift that moves two separate runs apart (an earlier pair of same-build runs moved a median by 129 ms). The instrument is also run A against A as a control. |
| Pass rule | A/B median difference at most 250 ms per class (2.5% of the 10 s hook timeout in the shipped settings) and A/A median difference at most 100 ms (else `INCONCLUSIVE`, which is a fail). |
| Paths | `non-identity-write` on `internal/hook/pre_tool.go`; `identity-outside` on `docs/a.md` (outside every zone entry before and after); `identity-inside` on `.claude/hooks/moai/x.sh` (legacy-denied before and after). No sampled path changes zone membership between the builds. |
| Baseline | L-7: A=base, B=an identical-code build: A/A median differences 0, 1, 0 ms; A/B −1, 1, 0 ms; all `PASS`. The control resolves two orders of magnitude below the bound. |
| Mutant probe | §C row g. |

### AC-SIPZ-008 — every denial and every degraded evaluation leaves a row

**Covers**: maps REQ-SIPZ-010, REQ-SIPZ-014

**Given** a denial, an absent manifest, or an invalid manifest, **when** the handler decides, **then** one line is appended to `.moai/logs/protected-zone-audit.jsonl` carrying the seven fields of REQ-SIPZ-014, and **when** the append fails (unwritable directory), the decision is unchanged and a stderr notice is printed.

| Cell | Content |
|---|---|
| Verification | `TestProtectedZone/AuditRow` over a temporary directory, including the unwritable-directory case. |
| Swept floor | As AC-SIPZ-006; the three states are three nested subtests, each must print its `--- PASS: TestProtectedZone/AuditRow/<State> ` line. |
| Mutant probe | §C row n. |

### AC-SIPZ-009 — the manifest ships neutral and arrives with a build

**Covers**: maps REQ-SIPZ-001, REQ-SIPZ-003

**Given** the template tree, **when** `internal/template/templates/.moai/config/sections/protected-zone.yaml` is read and the template and config tests run, **then** the file exists and parses; lists all seven required categories (empty lists allowed); ships `regression_tests` empty; lists both manifest files; every `paths` entry matches an existing path in the template tree and every `runtime_paths` entry is one of the entries spec §D marks as created at runtime; trips none of the template-neutrality audits (no SPEC identifier, requirement token, audit citation, internal date, commit hash, or `CLAUDE.local.md` literal — that entry is carried by the dogfood overlay, and the `Neutrality` subtest asserts the shipped file contains no such literal); the config completeness audit classifies the section as a dedicated-loader section; and the local dogfood manifest is byte-identical to the shipped one.

| Cell | Content |
|---|---|
| RED-now | L-9: `ls internal/template/templates/.moai/config/sections/protected-zone.yaml` exits **1**: `No such file or directory`; same for the local `.moai/config/sections/protected-zone.yaml`. |
| Why red | Right reason: the file is the deliverable; the config completeness audit would also fail (`YAML_SECTION_NO_LOADER`) the moment it appears without a registration, which is why M1 registers it in the same commit. |
| Green path | M1: the `ls` exits 0 and `go test -count=1 -timeout 30m -run '^TestProtectedZone$' -v ./internal/template/ ./internal/config/` prints, for each of the two packages, the parent `--- PASS: TestProtectedZone ` line and its three subtests — 8 `--- PASS` lines in total — after `make build` regenerates the embed. |
| Swept floor | 8 `--- PASS` lines; the shipped-file test also prints the number of `paths` and `runtime_paths` entries it swept and the number must be at least the count in spec §D (non-zero is not enough). `[no tests to run]` for either package is a fail. |
| Mutant probe | §C rows l, u. |

### AC-SIPZ-010 — the manifest cannot silently drift away from the lists it replaces

**Covers**: maps REQ-SIPZ-004

**Given** the compiled baseline and the three hard-coded frozen lists, **when** `TestProtectedZone/BaselineCovered` runs against the dogfood effective zone, **then** each of the 21 list entries enumerated in spec §C.2 — `internal/hook` 10, `internal/harness/safety` 7, `internal/harness` 4, of which 15 are distinct strings — is covered by a representative path, the test lists the entries by name, and it reports the number it swept (21). The entry only the `internal/harness` list carries, `.claude/skills/moai/`, is covered by an explicit manifest entry.

| Cell | Content |
|---|---|
| Verification | A count other than 21, or zero, fails; so does an enumerated name that is absent. |
| Live red-now | Probe row R21 (`.claude/skills/moai/SKILL.md`) is `allow` at base (L-1) and must be `deny` after M1/M2. |
| Mutant probe | §C row v: delete one manifest entry — the test must fail by that entry's name; add an entry to a compiled list without the manifest — it must fail. |

### AC-SIPZ-011 — if the guard stops firing, a build goes red without anyone asking

**Covers**: maps REQ-SIPZ-015

**Given** the repository's CI, **when** any of the four conditions of REQ-SIPZ-015 holds — the PreToolUse matcher group lacks Write, Edit or Bash in the local or template settings; a shipped or dogfood manifest fails to parse; a `paths` entry matches no existing path in its sweep tree (shipped file against the template tree, overlay against the local tree; `runtime_paths` entries exempt); the real handler does not deny a known zone-member input — **then** `TestProtectedZone/Liveness` fails and names which condition.

| Cell | Content |
|---|---|
| Continued-firing answer | The reader learns the guard went quiet from a red CI run on the next pull request or integration push, with the failing condition named in the test output. Nothing needs to be queried. |
| Swept floor | Each of the four conditions is a named nested subtest of `Liveness`; each must print its `--- PASS: TestProtectedZone/Liveness/<Condition> ` line; the dead-entry sweep must report how many `paths` entries it resolved and how many `runtime_paths` entries it skipped, and the resolved count must equal the number of `paths` entries in the files. |
| Mutant probe | Run-phase M3 executes each mutation and records the observed red in `progress.md` §E.2: (i) remove `Bash` from the template matcher; (ii) corrupt the shipped manifest; (iii) add a `paths` entry for a path that does not exist **in a category that also holds a `runtime_paths` entry** (so a mutant that exempts whole categories survives (iii) and fails here); (iv) make the handler return allow for R6. All four must fail the test. |
| Known blind spot | A binary older than the fix (spec §F G6). |

### AC-SIPZ-012 — path matching behaves the same on every platform CI covers

**Covers**: maps REQ-SIPZ-006

**Given** the pure lexical function and a table of Windows-shaped inputs, **when** the table runs on each CI operating system, **then** backslash separators, a drive-letter absolute path under the project root (`C:\proj\.claude\hooks\x` against root `C:/proj`), mixed case, a `.\` prefix, and a `//host/share` prefix outside the root resolve as the table states (the first four to the zone entry, the last to outside the zone), identically on darwin, linux and windows. The function carries its own absolute-path rule (leading `/`, `X:/`, leading `//`) so a POSIX host does not call a drive-letter path relative; the symlink steps are tested separately in `FileTools` against real temporary directories.

| Cell | Content |
|---|---|
| Verification | `TestProtectedZone/Normalization`, run by the existing CI matrix; the table is data, not conditional on `runtime.GOOS`. |
| Not covered | 8.3 short names, junctions, alternate data streams — disclosed in `spec.md` §E, not tested. |
| Mutant probe | §C row m. |

### AC-SIPZ-013 — the work stays inside its scope

**Covers**: maps REQ-SIPZ-016

**Given** the run-phase commits, **when** the changed-path list against the base SHA is read, **then** it contains no file under `.claude/agents/` or `internal/template/templates/.claude/agents/`, no change to either settings file (`.claude/settings.json`, `internal/template/templates/.claude/settings.json.tmpl`), and no edit under `.moai/hooks/`.

| Cell | Content |
|---|---|
| Command | `git diff --name-only e497f693608ac7ea45a08b06304dc585e927ff49 -- .claude/agents internal/template/templates/.claude/agents .claude/settings.json internal/template/templates/.claude/settings.json.tmpl .moai/hooks` — L-13 records empty stdout, exit 0 at base. |
| Positive control | L-13 also records the same form over two SPEC files changed since base, which prints both paths; the control is re-run after M2 over `internal/hook`, and an empty control means the command measured nothing and the absence is "unmeasured", not "clean". |
| Base observation | Empty at base, trivially — which is why this is a regression-guard. |

## §C Mutation table

Each row is a mutant that satisfies a one-sided reading of the criteria and violates the requirement; the paired criterion kills it. Run-phase executes every row and records the observed red in `progress.md` §E.2.

| # | Mutant | Killed by |
|---|---|---|
| a | Zone list hard-coded in Go; the manifest file is ignored | AC-SIPZ-004 MS4, MS6, MS7 (project-only entries are denied only if the manifests are read) |
| b | Raw-prefix match with no normalization (today's behaviour) | AC-SIPZ-002 P1–P4, S1 |
| c | Deny every identity Write | AC-SIPZ-005 N4, N5, N7, N8, N9 |
| d | Guard never emits the new sentinel for manifest-only entries | AC-SIPZ-001 R1–R20 |
| e | Invalid manifest treated as absent (fail open) | AC-SIPZ-004 MS1, MS8 |
| f | Invalid manifest denies every caller | AC-SIPZ-004 MS5 |
| g | Manifest read before the identity gate | AC-SIPZ-007 read-counter test |
| h | A bypass is honoured — environment variable, tool-input field, or command-text token | AC-SIPZ-004 E1, E2, E3 (one row each; a mutant that honours only the environment variable survives none of E2/E3) |
| i | Overlay replaces instead of adds | AC-SIPZ-004 MS7 (the base's `base_dir/` is denied only if the union is taken), with `ManifestStates` overlay-cannot-narrow as the unit mirror |
| j | Sentinel present, a routing field missing or cut | AC-SIPZ-001 judge reason column; AC-SIPZ-006 LP1, LP2 and `DenyReason` |
| k | Shell rule matches read-only verbs | AC-SIPZ-003 / AC-SIPZ-005 N6 |
| l | A sweep or test passes with zero entries or zero tests swept | AC-SIPZ-009, AC-SIPZ-010 and AC-SIPZ-011 swept-count floors; the empty-selection and swept-count failures of the judge |
| m | Case folding only where the host is case-insensitive | AC-SIPZ-002 P4 on the linux CI leg; AC-SIPZ-012 |
| n | A failed audit append changes the decision | AC-SIPZ-008 unwritable-directory subtest |
| o | PreToolUse matcher drops `Bash` or `Edit` | AC-SIPZ-011 (i) |
| p | Manifest `paths` entries only; `runtime_paths` entries are not protected | AC-SIPZ-001 R1, R13–R16 (settings and learning-history entries are `runtime_paths`) |
| q | Root resolved only when the target resolved (the existing check's asymmetry) | AC-SIPZ-002 P5, P6 |
| r | Shell verb list truncated to three forms | AC-SIPZ-003 B1–B13 (a three-verb mutant fails ten rows) and the 13-row `ShellMutation` floor |
| s | Overlay with an unknown key is ignored instead of invalidating | AC-SIPZ-004 MS8 |
| t | Truncation cuts the whole reason instead of only the path | AC-SIPZ-006 LP1, LP2 and `DenyReason` |
| u | Dead-entry sweep exempts whole categories that hold any `runtime_paths` entry | AC-SIPZ-011 mutation (iii) |
| v | A baseline entry silently missing from the manifest | AC-SIPZ-010 by-name enumeration; R21 |
| w | A baseline-matched Write/Edit denial gains routing fields (violates the byte-for-byte legacy reason) | AC-SIPZ-005 C1–C4 forbidden-substring assertion |

## §D Counts, baseline-first note, cross-platform, definition of done

**Swept-set counts the live judge must print after implementation** (the probe is one table, so the whole-table form is the release check): `JUDGE swept=67 expected=67 fail=0`; by group R=21, P/S=7, LP=2, B=13, MS/E=11, C/N=13. A selection that matches no row is itself a failure (`empty-selection`). The group labels MS1–MS8 are probe fixtures and are unrelated to the milestone names M1–M3.

**Baseline-first.** The base-tree measurements (`evidence/red-probe-base-e497f6936.tsv`, `judge-base-replay.txt`, `judge-base-live.txt`, `latency-ab-base-e497f6936.txt`) are plan-phase artifacts committed ahead of any run-phase commit, so the commit graph — not a commit message — witnesses that the red preceded the change (`verification-claim-integrity.md` §2.3). Re-measure on rebase; do not re-quote.

**Cross-platform.** The lexical function is pure string work and the same table runs on every CI leg; the symlink steps use real temporary directories where the host supports symlinks. The shell rule uses the existing shell-segment splitter and applies to Bash only; PowerShell is not covered (spec §E). Relative paths are judged against the process working directory exactly as the existing file-access check does.

**Edge cases the unit tests cover:** a Korean path in NFD against an NFC entry; an empty or directory `file_path`; a manifest with a duplicate entry (valid), a middle wildcard, an entry containing `..`, an absolute entry, an unknown top-level key, an unknown `version`, or a category name longer than 32 bytes (all `manifest=invalid`); an empty category list (valid); a zero-byte, `categories: {}` or one-category file at the **shipped** path (invalid: any of the seven required categories missing — a shrunken zone must not load silently — while the same file with all seven plus an extra category is valid, which is what the MS4, MS7 and MS8 fixtures use); a zero-byte overlay (valid, no entries); an overlay that is invalid while the shipped manifest is valid (fail closed); a deny reason for a path longer than the 240-byte cap (path truncated, four routing fields and sentinel intact).

**Quality gates.** `go vet ./internal/hook/... ./internal/config/... ./internal/template/...`, `golangci-lint run` on the changed packages, `go test -race ./internal/hook/...`, `make build` so the embed regenerates, and TRUST 5 sync-audit as usual. Full-suite verdict is CI's.

**Definition of done.** AC-SIPZ-001..004 and 009 green via their **live** commands with the RED→GREEN pair recorded; AC-SIPZ-005..013 green with swept counts shown; every §C mutant executed and its red recorded; `moai spec lint --strict SPEC-SELF-IMPROVE-PROTECTED-ZONE-001` clean; no edit outside the scope of AC-SIPZ-013.

## §E Evidence ledger

All entries: tree SHA `e497f693608ac7ea45a08b06304dc585e927ff49` (code this SPEC reads identical at HEAD `41127e8caded5c54507414fc6dd88fbf28d9517b`, see the header). L-1 to L-5 and L-6a were re-run live on 2026-10-05 with a binary built from `41127e8ca` after the probe fixtures MS4/MS7/MS8 gained their seven categories and the judge gained the C1–C4 forbidden-substring and MS1/MS8 failing-file assertions: the whole-table live run prints `JUDGE swept=67 expected=67 fail=51` (the replay on the recorded TSV prints the same), the six group outputs equal `judge-base-live.txt` and `judge-base-replay.txt` byte for byte, and the probe TSV equals `red-probe-base-e497f6936.tsv` byte for byte. L-9 and L-13 were re-run at `41127e8ca` with the same outputs. L-7 (latency) and L-8 (selector hazard) were not re-run and carry the base observation. Commands are repository-root relative; `J` abbreviates `bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh`. Raw files are under `.moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/`. The recorded runs passed a prebuilt binary (`go build ./cmd/moai` of this tree) as the last argument; omitting it makes the judge build one itself, same code path. The replay stdout of every entry below is identical to its live stdout (`judge-base-live.txt`).

```ledger
L-1
  live cmd:   J -o '^R'
  replay cmd: J -o '^R' .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/red-probe-base-e497f6936.tsv
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
    FAIL R21 got=allow/- want=deny/<any sentinel>
    JUDGE swept=21 expected=21 fail=21
  exit:   1

L-2
  live cmd:   J -o '^(P|S)'
  replay cmd: J -o '^(P|S)' .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/red-probe-base-e497f6936.tsv
  stdout:
    FAIL P1 got=allow/- want=deny/HARNESS_FROZEN_HOOK_VIOLATION
    FAIL P2 got=allow/- want=deny/HARNESS_FROZEN_HOOK_VIOLATION
    FAIL P3 got=allow/- want=deny/HARNESS_FROZEN_HOOK_VIOLATION
    FAIL P4 got=allow/- want=deny/HARNESS_FROZEN_HOOK_VIOLATION
    FAIL P5 got=allow/- want=deny/HARNESS_FROZEN_HOOK_VIOLATION
    FAIL P6 got=allow/- want=deny/HARNESS_FROZEN_HOOK_VIOLATION
    FAIL S1 got=allow/- want=deny/HARNESS_FROZEN_HOOK_VIOLATION
    JUDGE swept=7 expected=7 fail=7
  exit:   1

L-3
  live cmd:   J -o '^B'
  replay cmd: J -o '^B' .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/red-probe-base-e497f6936.tsv
  stdout:
    FAIL B1 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B2 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B3 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B4 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B5 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B6 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B7 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B8 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B9 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B10 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B11 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B12 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL B13 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    JUDGE swept=13 expected=13 fail=13
  exit:   1

L-4
  live cmd:   J -o '^(MS|E)'
  replay cmd: J -o '^(MS|E)' .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/red-probe-base-e497f6936.tsv
  stdout:
    FAIL MS1 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL MS4 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL MS6 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL MS7 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL MS8 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL E1 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL E2 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL E3 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    JUDGE swept=11 expected=11 fail=8
  exit:   1

L-5
  live cmd:   J -o '^(C|N)'
  stdout:
    JUDGE swept=13 expected=13 fail=0
  exit:   0

L-6a
  live cmd:   J -o '^LP'
  stdout:
    FAIL LP1 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    FAIL LP2 got=allow/- want=deny/HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
    JUDGE swept=2 expected=2 fail=2
  exit:   1

L-7
  cmd:    bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/baseline-latency.sh <binary A> <binary B> 21
          (A = base build, B = a build of the identical code; recorded in evidence/latency-ab-base-e497f6936.txt)
  stdout:
    non-identity-write	A/A n=21 median_diff_ms=0	A/B n=21 median_diff_ms=-1	bound=250 verdict=PASS
    identity-outside	A/A n=21 median_diff_ms=1	A/B n=21 median_diff_ms=1	bound=250 verdict=PASS
    identity-inside	A/A n=21 median_diff_ms=0	A/B n=21 median_diff_ms=0	bound=250 verdict=PASS
  exit:   0

L-8
  cmd:    go test -count=1 -timeout 30m -run '^TestProtectedZone$' -v ./internal/hook/
  stdout:
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/hook	0.573s [no tests to run]
  exit:   0
  reading: zero tests swept; an exit of 0 and the word PASS are not a pass for any AC that names this selector.

L-9
  cmd:    ls internal/template/templates/.moai/config/sections/protected-zone.yaml
  stdout: (empty)
  stderr: ls: internal/template/templates/.moai/config/sections/protected-zone.yaml: No such file or directory
  exit:   1
  cmd:    ls .moai/config/sections/protected-zone.yaml
  stdout: (empty)
  stderr: ls: .moai/config/sections/protected-zone.yaml: No such file or directory
  exit:   1

L-13
  cmd:    git diff --name-only e497f693608ac7ea45a08b06304dc585e927ff49 -- .claude/agents internal/template/templates/.claude/agents .claude/settings.json internal/template/templates/.claude/settings.json.tmpl .moai/hooks
  stdout: (empty)
  exit:   0
  control cmd: git diff --name-only e497f693608ac7ea45a08b06304dc585e927ff49 -- .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/spec.md .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/plan.md
  control stdout:
    .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/plan.md
    .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/spec.md
  control exit: 0
```
