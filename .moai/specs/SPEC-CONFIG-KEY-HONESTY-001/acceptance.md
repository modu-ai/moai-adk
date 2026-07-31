# SPEC-CONFIG-KEY-HONESTY-001 — Acceptance Criteria

## §A Discipline

1. **Every AC states a command and its expected observable output.** A criterion phrased as a
   property with no command is not an AC.
2. **`go test -run <pattern>` exits 0 on zero matches.** Every `-run` AC therefore also requires a
   verbatim `--- PASS: <exact test name>` line. Recorded vacuity baseline at code baseline `d5336214e`:
   ```
   $ go test -run 'TestShippedConfigKeysHaveReaders' ./internal/config/ ; echo "exit=$?"
   ok  	github.com/modu-ai/moai-adk/internal/config	0.443s [no tests to run]
   exit=0
   ```
   An AC whose only assertion is `exit 0` would pass against a tree with no test at all, and is
   rejected.
3. **Baselines are observed, and attributed to the code baseline `d5336214e`.** The worktree HEAD at
   which they were run is a descendant of `d5336214e` on branch `plan/epic-update-config-audit` that
   changes SPEC documents only — `git diff --name-only d5336214e HEAD | grep -v '\.md$'` returns
   zero lines and `git diff --stat d5336214e HEAD -- '*.go'` is empty — so no Go source differs
   between the two and every `file:line` and count below is attributable to `d5336214e`. Each AC
   carries its observed pre-change baseline so a reviewer can distinguish a real change from a
   no-op.
4. **Every new guard needs a falsification** proving it FAILS against unfixed code (§C).
5. **`git stash` is prohibited, and falsification does not use a worktree either.** This checkout is
   shared with concurrent sessions: `git stash` is repository-global, and a scratch `git worktree`
   costs a mutation of shared repository state for no benefit here. All five §C procedures therefore
   use `go test -overlay`, which substitutes files for one command invocation and leaves no state
   behind.
6. **All fixtures use `t.TempDir()`** and touch no path outside it (NFR-CKH-001).
7. **A command that cannot observe its own expectation is not an AC.** Clause 1 requires a command;
   this clause requires the command to be capable of seeing what the AC claims. Four criteria were
   rewritten under it — a `-run` selector naming a test that does not exist (AC-CKH-019), a `grep`
   that cannot tell which of two maps an entry sits in (AC-CKH-011), a `grep` whose window excludes
   the line it verifies (AC-CKH-012), and a `grep` with no context flags asserting something about
   adjacent comments (AC-CKH-014). Each rewrite records the observed before/after in place, so the
   failure mode stays documented rather than merely removed. The same test applies to a
   falsification: a mutation step expressed as a comment mutates nothing, which is why every §C
   procedure now carries a `diff -q` no-op guard.

## §B Acceptance criteria

### M1 — Triage rule and inventory

#### AC-CKH-001 — the triage rule exists and defines the dotted-path discriminator

```bash
test -f .moai/docs/config-key-triage-rule.md && \
  grep -cE 'dotted (key )?path|fully-qualified' .moai/docs/config-key-triage-rule.md && \
  grep -c 'homonym' .moai/docs/config-key-triage-rule.md
```

Expected: `test -f` succeeds, and both `grep -c` print a count `>= 1`, establishing that the rule
names the dotted-path discriminator and explicitly records that a bare leaf-key match is a homonym.

Baseline: the file does not exist —
`ls .moai/docs/config-key-triage-rule.md` prints
`ls: .moai/docs/config-key-triage-rule.md: No such file or directory` and exits 1.

#### AC-CKH-002 — the inventory covers every shipped key with a class

```bash
go test -run 'TestShippedKeyInventoryIsComplete' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestShippedKeyInventoryIsComplete` line. The test parses every
`internal/template/templates/.moai/config/sections/*.yaml*` into dotted paths, loads
`internal/config/testdata/shipped_key_inventory.yaml`, and asserts (a) every shipped key has an
inventory entry, (b) every entry's `class` is one of `W`/`P`/`R`/`D`, (c) every `P` and `W` entry
carries non-empty `evidence`.

Baseline: neither the test nor the inventory exists —
`ls internal/config/testdata/shipped_key_inventory.yaml` prints
`No such file or directory`, and the `-run` selector matches nothing (see §A clause 2).

#### AC-CKH-003 — the seven highest-impact families are individually classified

```bash
go test -run 'TestShippedKeyInventoryFamilyCoverage' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestShippedKeyInventoryFamilyCoverage` line. The test asserts that every
dead key measured in the `design`, `harness`, `research`, `git-strategy`, `constitution`, `context`,
and `workflow` sections carries an inventory entry whose `evidence` field is populated (not the
literal `unclassified`).

Baseline recorded at code baseline `d5336214e`, **path-resolved** (spec.md §A.3) — the dead-key
occurrence counts these families must cover:

```
design.yaml             34     harness.yaml            25
research.yaml           21     git-strategy.yaml.tmpl  20
constitution.yaml       16     workflow.yaml           14
context.yaml            14     quality.yaml.tmpl        9
security.yaml            7     sunset.yaml              6
interview.yaml           5     system.yaml.tmpl         4
mx.yaml                  4     lsp.yaml.tmpl / git-convention.yaml   2 each
state.yaml / ralph.yaml / project.yaml.tmpl / observability.yaml / delegation.yaml   1 each

dead field names total: 174     of which mapping to a shipped key: 161
(file, key) occurrences: 188    accessor-only (types.go): 4
```

These figures **replace** the superseded bare-field-name baseline (122 / 121, seven families) that
earlier drafts of this AC recorded. The replacement is not cosmetic: the bare-name method is the one
AC-CKH-007 and plan.md §G AP-3 forbid, and under it a correct M2 implementation would have
contradicted this AC's own baseline — the more precisely the guard resolved paths, the further its
dead set would drift above 122. Two guards against re-drift:

- The counts above are **occurrences**, not distinct keys, because the field-name → section-file
  mapping remains leaf-name based (spec.md §A.3, third methodology fact). M2's reflective section
  walk narrows this; when the guard's per-family output differs from the table, the difference is a
  finding to record in `progress.md` §E.2, not a licence to edit this baseline to match.
- The two figures the guard **must** reproduce exactly are `174` dead field names and `4`
  accessor-only, since both come from the same path-resolved rule M2 step 2-3 implements.

#### AC-CKH-004 — no key is classified D on Go-reader evidence alone

```bash
go test -run 'TestTriageRuleProseProbeBeforeDelete' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestTriageRuleProseProbeBeforeDelete` line. For every inventory entry with
`class: D`, the test runs the fixed-string dotted-path probe over the shipped prose corpus and
asserts zero matches — i.e. it re-proves the **P**-before-**D** ordering rather than trusting the
recorded class.

Baseline (the probe's precision, measured at code baseline `d5336214e` over `.claude/agents`,
`.claude/skills`, `.claude/rules`):

```
design.evolution.max_active_learnings -> 0     bare 'max_rounds'  -> 5
workflow.worktree.auto_cleanup        -> 1     bare 'escalation'  -> 46
research.budget_cap_tokens            -> 0     bare 'adaptation'  -> 6
interview.max_rounds                  -> 0
harness.escalation                    -> 0
```

### M2 — Anti-rot guard

#### AC-CKH-005 — the guard passes on the fixed tree

```bash
go test -run 'TestShippedConfigKeysHaveReaders' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestShippedConfigKeysHaveReaders` line.

Baseline: `[no tests to run]`, exit 0 (§A clause 2, recorded verbatim above).

#### AC-CKH-006 — the guard is not vacuous

```bash
go test -run 'TestShippedConfigKeysHaveReaders/non_vacuous_inventory' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestShippedConfigKeysHaveReaders/non_vacuous_inventory` line. The subtest
asserts the shipped-key inventory has `>= 200` entries and the reflective walk of `Config` yields
`>= 200` struct fields (NFR-CKH-002), so a guard that inventories zero fails instead of passing.

Baseline for plausibility: at code baseline `d5336214e`, `grep -c 'yaml:"' internal/config/types.go` prints
`371`, and the field-name-deduped walk yields `287` distinct names — both comfortably above the
floor, so the floor cannot be met by an accidentally-truncated walk.

#### AC-CKH-007 — the guard resolves paths, not bare field names

```bash
go test -run 'TestShippedConfigKeysHaveReaders/collision_resolution' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestShippedConfigKeysHaveReaders/collision_resolution` line. The subtest
asserts that `workflow.worktree.auto_create` (which has a production reader at
`internal/cli/worktree_advisory.go:29`) and a same-named field on a different struct resolve to
distinct lookups, so liveness cannot leak across the collision.

Baseline: `pkg/models/config.go:172` already declares a second `AutoCreate bool \`yaml:"auto_create"\``
independent of `internal/config/types.go:486` — the collision exists in the tree today, so the
subtest exercises a real case, not a synthetic one.

#### AC-CKH-008 — accessor indirection is followed, not treated as dead

```bash
go test -run 'TestShippedConfigKeysHaveReaders/accessor_indirection' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestShippedConfigKeysHaveReaders/accessor_indirection` line. The subtest
asserts `GateTimeouts.Vet` classifies **accessor-live**, on the evidence that it is read via
`GateConfig.VetTimeoutDuration()` whose production caller is `internal/hook/pre_tool.go:657`.

Baseline: a direct-read-only classifier marks this field dead — which is exactly the false positive
the audit lens flagged and refused to report.

### M3 — `quality.yaml`

#### AC-CKH-009 — the parse function no longer misnames its schema

```bash
grep -rn 'parseFullQualityConfig\|parseQualityConstitution\|FullQualityConfig' \
  --include='*.go' internal pkg cmd | grep -v '_test.go'
```

Expected: **either** every `parseFullQualityConfig` occurrence is gone (renamed to
`parseQualityConstitution`, with `internal/lsp/hook/gate.go:312` updated), **or** the function body
unmarshals into `models.FullQualityConfig` — i.e. `FullQualityConfig` appears in
`internal/lsp/hook/gate.go`, not only in `pkg/models/config.go`. A tree where the name says "Full"
while the body says `qualityFileWrapper` fails.

Baseline at code baseline `d5336214e` (the failing state):

```
internal/lsp/hook/gate.go:125:// parseFullQualityConfig parses the full quality config ...
internal/lsp/hook/gate.go:126:func parseFullQualityConfig(data []byte) (models.QualityConfig, error) {
internal/lsp/hook/gate.go:312:	return parseFullQualityConfig(data)
pkg/models/config.go:198:// FullQualityConfig represents the complete quality.yaml structure.
pkg/models/config.go:199:type FullQualityConfig struct {
```

`FullQualityConfig` appears only as a type declaration and a comment; `gate.go` never mentions it.

#### AC-CKH-010 — every shipped `quality.yaml` block is parsed or marked

```bash
go test -run 'TestQualityYAMLBlocksParsedOrMarked' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestQualityYAMLBlocksParsedOrMarked` line. For each of
`report_generation`, `lsp_state_tracking`, `constitution.memory_guard`, and
`constitution.session_effort_default`, the test asserts the block either round-trips through a
production parse path or carries the generic reserved marker in `quality.yaml.tmpl`.

Baseline: none of the four is reachable —

```
$ grep -rn 'session_effort_default\|memory_guard\|SessionEffortDefault\|MemoryGuard' \
    --include='*.go' internal pkg cmd | grep -v '_test.go' | grep -v main-fork
(no output)
$ grep -n 'session_effort_default\|memory_guard\|report_generation\|lsp_state_tracking' \
    internal/template/templates/.moai/config/sections/quality.yaml.tmpl
10:  session_effort_default: "xhigh"
190:  memory_guard:
198:report_generation:
212:lsp_state_tracking:
```

### M4 — `system.yaml`

#### AC-CKH-011 — the registry stops asserting an absent binding

```bash
go test -run 'TestAuditParity' -count=1 -v ./internal/config/
awk '/^var yamlAuditExceptions/,/^}/'  internal/config/audit_registry.go | grep -c '"system"'
awk '/^var yamlToStructRegistry/,/^}/' internal/config/audit_registry.go | grep -c '"system"'
grep -rn 'loadSystemSection' internal/config/
```

Expected: a `--- PASS: TestAuditParity` line, then **either**

- exceptions-block count `1`, registry-block count `0`, and `grep -rn 'loadSystemSection'` printing
  no output — the entry has moved to `yamlAuditExceptions` with a reason naming the real readers;
- **or** exceptions `0`, registry `1`, and `grep -rn 'loadSystemSection'` printing a declaration
  plus a `Loader.Load` call site — a real binding now exists.

Any other combination fails, including the current one.

Baseline at code baseline `d5336214e` (the concealed state) — observed:

```
$ awk '/^var yamlAuditExceptions/,/^}/'  internal/config/audit_registry.go | grep -c '"system"'
0
$ awk '/^var yamlToStructRegistry/,/^}/' internal/config/audit_registry.go | grep -c '"system"'
1
$ grep -rn 'loadSystemSection' internal/config/
(no output)
```

and `TestAuditParity` passes anyway.

**Why the commands changed.** The earlier form was `go test … && grep -n '"system"' audit_registry.go`,
which prints line 31 identically whether the entry sits in `yamlToStructRegistry` or in
`yamlAuditExceptions` — it cannot see the one fact the AC is about, so the whole chain exited 0 on
the unmodified tree and the AC proved nothing. The `awk` range restricts each count to one map, so
the pair `(0, 1)` distinguishes the pre-change state from both accepted post-change states. The
range anchors on `^var ` because the bare identifier also appears at `audit_registry.go:83-84` and
`:105`, which would open a second unwanted range.

#### AC-CKH-012 — the `hook.*` block binds through a loader, not an inline struct

```bash
go test -run 'TestSystemHookOptInLoadsViaLoader' -count=1 -v ./internal/config/
awk '/^func isHookOptInEnabled/,/^}/' internal/cli/hook.go | grep -c 'struct {'
```

Expected: a `--- PASS: TestSystemHookOptInLoadsViaLoader` line, and the `awk`-scoped count printing
`0` — `isHookOptInEnabled` no longer declares an inline anonymous struct for `system.yaml`.

Baseline at code baseline `d5336214e` — observed:

```
$ awk '/^func isHookOptInEnabled/,/^}/' internal/cli/hook.go | grep -c 'struct {'
3
```

`internal/cli/hook.go:507-515` reads `system.yaml` with `os.ReadFile` into an inline
`var doc struct { Hook struct { OptIn struct { ... } } }` — the three nested struct literals the
count reports — bypassing `SystemConfig` entirely.

**Why the command changed.** The earlier form was `grep -n 'struct {' internal/cli/hook.go | sed -n '1,3p'`,
which prints the file's first three matches — lines 41, 475, 476 — while the struct under
verification is at line 513. Deleting the target struct left that output byte-identical, so the
second half of the AC could not fail. The `awk` range restricts the scan to the function body; the
range terminator `/^}/` is column-anchored and the nested braces are indented, so it closes on the
function's own brace.

#### AC-CKH-013 — `github.*` and `document_management.*` are classified

```bash
go test -run 'TestShippedKeyInventoryFamilyCoverage/system' -count=1 -v ./internal/config/
go test -run 'TestShippedConfigKeysHaveReaders/unbound_classification' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestShippedKeyInventoryFamilyCoverage/system` line, asserting every key under
`github:` and `document_management:` in `system.yaml.tmpl` carries an inventory class, and that any
key stating a retention or deletion promise carries the reserved marker. Plus a
`--- PASS: TestShippedConfigKeysHaveReaders/unbound_classification` line, asserting those same 25
keys are reported by the guard in the `unbound` class (spec.md §B.4, plan.md §F M2 step 3) rather
than skipped — and that removing one of them from the **R** allowlist makes the guard fail (C-5).

Baseline at code baseline `d5336214e` — observed: no Go struct binds either block, so the keys
resolve to no field and a four-class partition never reaches them.

```
$ grep -rn 'yaml:"github"\|yaml:"document_management"' --include='*.go' internal pkg cmd
(no output, exit 1)
$ github.* subkeys in system.yaml.tmpl:              2
$ document_management.* subkeys in system.yaml.tmpl: 23
```

The neighbouring `moai:` block (6 subkeys) is deliberately **not** counted here. It also has no
`SystemConfig` field, but unlike the other two it has real readers —
`internal/statusline/version.go:31`, `internal/cli/v2_detection.go:153`, and
`internal/cli/update/plan/plan.go:313` each parse it through an ad-hoc inline `yaml:"moai"` struct.
It is therefore `unbound` with allowlist evidence, not an undetected promise, and the count this AC
governs is 25, not 31.

### M5 — Documented-but-unenforced reconciliation

#### AC-CKH-014 — no surface claims `max_active_learnings` has effect

```bash
grep -n -B3 -A1 'max_active_learnings\|MaxActiveLearnings' \
  internal/config/types.go internal/config/defaults.go
```

Expected: both sites carry a comment naming the actual enforcement constants, and both literals
`internal/evolution/types.go` and `internal/constitution/rate_limiter.go` appear **within the
printed context**, so a reader of the config cannot mistake the key for the lever.

Baseline at code baseline `d5336214e` — observed with the same context flags:

```
internal/config/types.go-1091-	GraduationCriteria      DesignGraduationCriteria `yaml:"graduation_criteria"`
internal/config/types.go:1092:	MaxActiveLearnings      int                      `yaml:"max_active_learnings"`
internal/config/types.go-1093-	MaxEvolutionRatePerWeek int                      `yaml:"max_evolution_rate_per_week"`
internal/config/defaults.go-752-			},
internal/config/defaults.go:753:			MaxActiveLearnings:      50,
internal/config/defaults.go-754-			MaxEvolutionRatePerWeek: 3,
```

Neither enforcement-site literal appears — the criterion fails today and can only be satisfied by
adding the comments. Enforcement sits in two unrelated packages
(`internal/evolution/types.go:170`, `internal/constitution/rate_limiter.go:14`).

**Why the command changed.** The earlier form had no `-B`/`-A`, so it printed only the two
declaration lines. A maintainer adding the required comment on the line above or below each
declaration — the natural placement — would leave that output byte-identical, making the stated
expectation unobservable by its own command. AC-CKH-017 already uses `-A2` for the same purpose;
this restores the consistency.

#### AC-CKH-015 — the shipped worktree toggles match the Go defaults

```bash
grep -n 'auto_create\|auto_merge\|auto_cleanup' \
  internal/template/templates/.moai/config/sections/workflow.yaml
sed -n '543,549p' internal/config/defaults.go
```

Expected: the three shipped values agree with `internal/config/defaults.go`, and each unwired toggle
carries the generic reserved marker.

Baseline — the contradiction recorded in spec.md §A.8:

```
workflow.yaml:35:        auto_create: false
workflow.yaml:36:        auto_merge: true          <-- defaults.go says false
workflow.yaml:37:        auto_cleanup: true        <-- defaults.go says false
defaults.go:545:			AutoCleanup:        false,
defaults.go:546:			AutoCreate:         false,
defaults.go:547:			AutoMerge:          false,
```

#### AC-CKH-016 — `CLAUDE.local.md` §22.8 states the real reader status

```bash
sed -n '/§22.8/,/§22.9/p' CLAUDE.local.md | grep -cE 'auto_cleanup|auto_merge|auto_create'
```

Expected: a count `>= 3`, with the surrounding prose stating that `auto_cleanup` and `auto_merge`
have no reader and that `auto_create` is read only to select advisory wording.

Baseline: §22.8 currently describes all three as governing web-console worktree automation, which
matches no code path — `internal/cli/worktree_advisory.go:29` is the only read of any of the three,
and it selects between two `fmt.Fprintln` strings.

#### AC-CKH-017 — `session_name_pattern` no longer implies it is used

```bash
grep -n -A2 'session_name_pattern' internal/template/templates/.moai/config/sections/workflow.yaml
```

Expected: the key carries the generic reserved marker, so the shipped file no longer presents a
pattern that nothing consumes as an active setting.

Baseline: `workflow.yaml:39` ships `session_name_pattern: "moai-{ProjectName}-{SPEC-ID}"` with no
marker, while `SessionNamePattern`'s only non-default hits are three test files asserting the
default value.

### M6 — Neutrality and handoff

#### AC-CKH-018 — the three leaks are gone from shipped content

```bash
grep -rnE 'SPEC-(AGENT-ARCH-V2|[A-Z0-9-]*)-[0-9]{3}' internal/template/templates/.moai/config/ ; \
grep -rn 'issue #' internal/template/templates/.moai/config/ ; \
grep -rn 'plan\.md §' internal/template/templates/.moai/config/
```

Expected: all three greps produce no output (each exits 1). Generic placeholders `{SPEC-ID}`,
`<SPEC-ID>`, and `SPEC-XXX` remain and are unaffected, because none matches the 3-digit-suffixed
pattern above.

Baseline at code baseline `d5336214e` — the three genuine leaks:

```
.../sections/workflow.yaml:65:  # cycle (plan.md §D D6, SPEC-AGENT-ARCH-V2-001 M3b). Values mirror the
.../sections/workflow.yaml:85:  # model_routing_profiles: No-Haiku 3-tier policy (SPEC-AGENT-ARCH-V2-001
.../sections/llm.yaml:179:      # (issue #653). Claude Code reports context_window_size based on the
```

alongside the generic placeholders at `workflow.yaml:39`, `statusline.yaml:28`, `cache.yaml:6`,
`cache.yaml:18`, which must survive.

#### AC-CKH-019 — the existing neutrality guard still passes and was not edited

```bash
go test -run 'TestTemplateNoInternalContentLeak' -count=1 -v ./internal/template/
git diff --stat d5336214e HEAD -- internal/template/internal_content_leak_test.go
```

Expected: a verbatim `--- PASS: TestTemplateNoInternalContentLeak` line (per §A clause 2), and
`git diff --stat` producing **no output** for that file — REQ-CKH-013 forbids this SPEC from
touching it.

**Why the diff is ref-anchored.** The earlier form was `git diff --stat -- <file>` with no ref,
which compares the working tree against the index. A forbidden edit that is **committed** then
produces empty output and the criterion passes — exactly the violation this half exists to catch.
Observed at the code baseline, with a file that *did* change between `d5336214e` and `HEAD`
substituted to make the two forms distinguishable:

```
$ git diff --stat -- .moai/specs/SPEC-CONFIG-KEY-HONESTY-001/acceptance.md
(no output)
$ git diff --stat d5336214e HEAD -- .moai/specs/SPEC-CONFIG-KEY-HONESTY-001/acceptance.md
 .../SPEC-CONFIG-KEY-HONESTY-001/acceptance.md      | 370 +++++++++++++++++----
 1 file changed, 298 insertions(+), 72 deletions(-)
```

The unanchored form cannot distinguish "unmodified" from "modified and committed"; the anchored
form can. AC-CKH-021 already uses the same `d5336214e HEAD` anchoring, so the two criteria are now
consistent. A **committed** edit to the guard file is detected by this form.

Baseline at code baseline `d5336214e` — observed:

```
$ go test -run 'TestTemplateNoInternalContentLeak' -count=1 -v ./internal/template/
=== RUN   TestTemplateNoInternalContentLeak
--- PASS: TestTemplateNoInternalContentLeak (0.76s)
ok  	github.com/modu-ai/moai-adk/internal/template	2.411s
```

The guard passes today (it cannot see the three §A.9 leaks), and the file is unmodified.

**Why the test name changed — this AC was inert.** The earlier form named
`TestInternalContentLeak`, which does not exist in the tree. Go's `-run` takes a regexp matched
against test names, and no test name contains that substring — the closest is
`TestTemplateNoInternalContentLeak`, in which `TestInternalContentLeak` is not a substring. Observed:

```
$ go test -run 'TestInternalContentLeak' -count=1 ./internal/template/ ; echo "exit=$?"
ok  	github.com/modu-ai/moai-adk/internal/template	0.334s [no tests to run]
exit=0
```

Deleting the entire neutrality guard would still have produced `exit=0`, so the AC asserted nothing
— the exact shape §A clause 2 of this file declares rejected. The corrected form binds the real
test name and demands the verbatim `--- PASS:` line, so the AC now fails if the guard is removed,
renamed, or made to fail.

#### AC-CKH-020 — the E5 handoff records the three measured gaps

```bash
grep -cE 'SPEC-AGENT-ARCH-V2-001|issue #N|plan\.md §' <handoff-note-path>
```

Expected: a count `>= 3`, covering the three evidence points — the unregistered SPEC family, the
`issue #N` vs `PR #N` C6 asymmetry, and the absent artifact-citation class. `<handoff-note-path>` is
the dev-facing note produced by M6 (its exact path is fixed during run-phase and recorded in
`progress.md` §E.2).

Baseline: no handoff note exists.

### M6 — deprecation posture

#### AC-CKH-023 — a template-removed key survives in the user's file and is reported once (REQ-CKH-011)

```bash
go test -run 'TestTemplateRemovedKeySurvivesUserConfig' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestTemplateRemovedKeySurvivesUserConfig` line. The test builds a
`t.TempDir()` project whose `.moai/config/sections/<section>.yaml` carries a user-set key that the
current template no longer ships, runs the merge, and asserts three things: (a) the key and its
user-set **value** are still present in the user's file afterwards; (b) the removal is reported
exactly once — a second merge over the same tree emits no further report; (c) no other user-set key
is dropped.

Fails when: the removal path deletes from the user file rather than the template only (a), when the
report is emitted per-merge rather than once (b), or when the merge drops sibling keys (c).

Baseline: the test does not exist, and the `-run` selector matches nothing (§A clause 2). No
production report-once mechanism is asserted anywhere in the tree today.

Scope note: the merge engine remains owned by `SPEC-UPDATE-YAML-PRESERVE-001` (plan.md §D3). This AC
does not re-specify merge semantics; it pins the **delete-never, report-once** posture over that
engine for the keys M1 classifies **D**, which is the only deletion this SPEC performs.

### Cross-cutting

#### AC-CKH-021 — the full suite is green, the build is clean, and new sources meet NFR-CKH-003

```bash
go build ./... && go vet ./... && go test -count=1 ./...

git diff --name-only --diff-filter=A d5336214e HEAD -- '*.go' > /tmp/ckh-new-go.txt
test -s /tmp/ckh-new-go.txt || { echo "VACUOUS: no added .go files"; exit 1; }
xargs gofmt -l < /tmp/ckh-new-go.txt
xargs -n1 basename < /tmp/ckh-new-go.txt | grep -vE '^[a-z0-9_]+\.go$'
```

Expected: exit 0 from `go build`, `go vet`, and `go test`; the `test -s` non-vacuity guard passing
silently; `xargs gofmt -l` printing nothing; and the basename filter printing nothing (exit 1) —
every added source uses a `snake_case.go` filename. Error wrapping (`fmt.Errorf("...: %w", err)`)
and English comments/godoc are verified by reading the added files, since neither is expressible as
a single reliable pattern match.

The file-list indirection is not stylistic: `gofmt -l` with an empty argument list reads **stdin**
and blocks, so a naive `gofmt -l $NEW` would hang rather than fail on the very tree state — no added
files — that the non-vacuity guard exists to catch. `test -s` short-circuits before that happens.

The filter is `--diff-filter=A` (added), deliberately not `AM`. M3 and M5 edit pre-existing files
that are already in the 114-file gofmt-listed set — `internal/config/types.go` is one of them,
observed directly — so including modified files would fail this AC on the pre-existing alignment
rather than on anything this SPEC wrote. NFR-CKH-003 governs "Go sources added by this SPEC"; the
filter matches that wording exactly.

Baseline at code baseline `d5336214e`: `go build` / `go vet` / `go test` green (pre-flight,
plan.md §C); `git diff --name-only --diff-filter=A d5336214e HEAD -- '*.go'` returns **0 files**, so
the non-vacuity guard prints `0` and this half of the AC fails until the SPEC adds a source file.
`git ls-files 'internal/config/*.go' | xargs -n1 basename | grep -vE '^[a-z0-9_]+\.go$'` returns
**0** across 82 tracked files, so the naming convention already holds and the filter is a regression
guard rather than a cleanup task.

**Why the check is scoped to added files.** `gofmt -l internal pkg cmd` lists **114** pre-existing
files at the code baseline — alignment-only diffs from a different toolchain version, e.g.
`internal/cli/mx_scan.go`'s comment column. A repository-wide `gofmt -l` expectation would therefore
be unsatisfiable on an unmodified tree: not a gate, just a permanently red criterion. The
non-vacuity guard is required because the complementary failure is equally real — with no added
files, `gofmt -l` over an empty argument list prints nothing and would pass while proving nothing.

#### AC-CKH-022 — no test writes outside `t.TempDir()`

```bash
go test -count=1 ./internal/config/ ./internal/template/ && git status --porcelain
```

Expected: the tests pass and `git status --porcelain` reports no files created or modified by the
test run (NFR-CKH-001).

Baseline: clean tree before the run.

## §C Falsification procedures

Each new guard must be shown to FAIL against unfixed code. `git stash` is prohibited (§A clause 5).

### C-1 — the anti-rot guard actually catches a dead key

Use `go test -overlay` to substitute an inventory fixture with one **R** entry removed:

```bash
mkdir -p /tmp/ckh-falsify
# copy the real inventory, delete a single R entry (e.g. quality.report_generation.enabled)
sed '/report_generation\.enabled/,+2d' internal/config/testdata/shipped_key_inventory.yaml \
  > /tmp/ckh-falsify/shipped_key_inventory.yaml
if diff -q internal/config/testdata/shipped_key_inventory.yaml \
     /tmp/ckh-falsify/shipped_key_inventory.yaml >/dev/null; then
  echo "MUTATION NO-OP — sed matched nothing"; exit 1
fi
printf '{"Replace":{"internal/config/testdata/shipped_key_inventory.yaml":"/tmp/ckh-falsify/shipped_key_inventory.yaml"}}' \
  > /tmp/ckh-falsify/overlay.json
go test -overlay=/tmp/ckh-falsify/overlay.json -run 'TestShippedConfigKeysHaveReaders' \
  -count=1 -v ./internal/config/
```

Expected: **FAIL**, with the failure message naming `report_generation.enabled` as a shipped key with
no reader, no prose consumer, and no reserved marker. A PASS here means the guard's allowlist lookup
is not load-bearing and the guard is inert.

The `diff -q` guard is required by §A clause 7 for the same reason it is required in C-3: the
run-phase inventory's exact formatting is not fixed at plan time, so a `sed` address that matches
nothing leaves the overlay byte-identical to the original. The procedure would then report PASS
while having mutated nothing — a falsification that refutes itself rather than the guard. Observed
on a scratch fixture (three-line file, address deliberately unmatched vs matched):

```
$ sed '/zzz_nomatch/,+2d' orig.yaml > out.yaml
$ if diff -q orig.yaml out.yaml >/dev/null; then echo "MUTATION NO-OP — sed matched nothing"; exit 1; fi; echo "mutation confirmed"
MUTATION NO-OP — sed matched nothing        exit=1

$ sed '/^b$/,+0d' orig.yaml > out2.yaml
$ if diff -q orig.yaml out2.yaml >/dev/null; then echo "MUTATION NO-OP — sed matched nothing"; exit 1; fi; echo "mutation confirmed"
mutation confirmed                           exit=0
```

The `if` form is used here rather than C-3's `diff -q … && { …; exit 1; }` form because the latter
leaves the *success* path carrying `diff -q`'s own exit 1 (files differ), so its exit code cannot
distinguish the two outcomes — only its printed message can. The `if` form discriminates by exit
code as well, as the transcript above shows. C-3 / C-4 / C-5 retain the `&&` form; their guards
still surface the no-op via the message, which is what §D asserts.

### C-2 — the non-vacuity floor actually fires

Same overlay mechanism, substituting an inventory truncated to a handful of entries:

```bash
head -12 internal/config/testdata/shipped_key_inventory.yaml > /tmp/ckh-falsify/tiny.yaml
if diff -q internal/config/testdata/shipped_key_inventory.yaml \
     /tmp/ckh-falsify/tiny.yaml >/dev/null; then
  echo "MUTATION NO-OP — inventory is already <= 12 lines, nothing was truncated"; exit 1
fi
printf '{"Replace":{"internal/config/testdata/shipped_key_inventory.yaml":"/tmp/ckh-falsify/tiny.yaml"}}' \
  > /tmp/ckh-falsify/overlay2.json
go test -overlay=/tmp/ckh-falsify/overlay2.json \
  -run 'TestShippedConfigKeysHaveReaders/non_vacuous_inventory' -count=1 -v ./internal/config/
```

Expected: **FAIL** naming the `>= 200` floor (NFR-CKH-002). A PASS means an empty or truncated
inventory would sail through, which is the exact vacuity hazard AC-CKH-006 exists to exclude.

The `diff -q` guard catches this procedure's own no-op shape (§A clause 7): if the inventory is
already 12 lines or fewer, `head -12` copies it verbatim and the overlay substitutes the file with
itself. The subsequent FAIL would then prove nothing about truncation — it would merely restate that
the unmutated inventory is below the floor. Observed on a three-line scratch fixture:

```
$ head -12 orig.yaml > tiny.yaml
$ if diff -q orig.yaml tiny.yaml >/dev/null; then echo "MUTATION NO-OP — inventory is already <= 12 lines, nothing was truncated"; exit 1; fi
MUTATION NO-OP — inventory is already <= 12 lines, nothing was truncated        exit=1
```

At the ≥ 200-entry inventory NFR-CKH-002 requires, the guard is expected to pass silently; it fires
only if the inventory itself has collapsed, which is the condition worth catching.

### C-3 — the prose probe genuinely blocks a delete

C-1's overlay mechanism, mutating the inventory rather than the source. The mutation is a command,
not a described intention:

```bash
mkdir -p /tmp/ckh-falsify
# reclassify the P key workflow.worktree.auto_cleanup to D, leaving its prose consumer in place
sed 's/^\(  *\)class: P\( *# workflow\.worktree\.auto_cleanup\)/\1class: D\2/' \
  internal/config/testdata/shipped_key_inventory.yaml > /tmp/ckh-falsify/reclassified.yaml
diff -q internal/config/testdata/shipped_key_inventory.yaml /tmp/ckh-falsify/reclassified.yaml \
  && { echo "MUTATION NO-OP — sed matched nothing"; exit 1; }
printf '{"Replace":{"internal/config/testdata/shipped_key_inventory.yaml":"/tmp/ckh-falsify/reclassified.yaml"}}' \
  > /tmp/ckh-falsify/overlay3.json
go test -overlay=/tmp/ckh-falsify/overlay3.json -run 'TestTriageRuleProseProbeBeforeDelete' \
  -count=1 -v ./internal/config/
```

Expected: **FAIL**, naming `workflow.worktree.auto_cleanup` and the prose file whose dotted-path
match contradicts the `D` class — `.claude` prose matched this dotted path once at the code
baseline (spec.md §A.4), which is why it is the chosen subject. A PASS means the probe trusts the
recorded class instead of re-proving it, and AP-1 (deleting a prose-consumed key) is unguarded.

The `diff -q` line is part of the procedure, not decoration: the run-phase inventory's exact
formatting is not fixed at plan time, so a `sed` that silently matches nothing would leave the
overlay identical to the original and the procedure would report PASS while having mutated nothing —
a falsification that falsifies itself. The guard makes that outcome an explicit failure.

**Why the worktree was removed.** The earlier form ran `git worktree add /tmp/ckh-wt HEAD`, stated
the mutation only as a shell comment ("reclassify one P key to D"), and then executed the test
against an unmutated tree — which after implementation yields PASS, the opposite of the predicted
FAIL. It also ran `git worktree remove /tmp/ckh-wt` at its end while C-4 below depended on that same
path, so C-4 could not run at all. Overlay removes both defects and the shared-checkout hazard
(§A clause 5) with it.

### C-4 — the collision defence is load-bearing

Overlay-replace the guard source itself with a variant whose step-2 lookup is by bare field name:

```bash
# run-phase deliverable: a copy of the guard identical except that resolveFieldPath is replaced
# by a bare field-name lookup over types.go's field set
cp internal/config/shipped_key_reader_test.go /tmp/ckh-falsify/bare_name_variant.go
#   ...apply the bare-name substitution to /tmp/ckh-falsify/bare_name_variant.go...
diff -q internal/config/shipped_key_reader_test.go /tmp/ckh-falsify/bare_name_variant.go \
  && { echo "MUTATION NO-OP — variant identical to original"; exit 1; }
printf '{"Replace":{"internal/config/shipped_key_reader_test.go":"/tmp/ckh-falsify/bare_name_variant.go"}}' \
  > /tmp/ckh-falsify/overlay4.json
go test -overlay=/tmp/ckh-falsify/overlay4.json \
  -run 'TestShippedConfigKeysHaveReaders/collision_resolution' -count=1 -v ./internal/config/
```

Expected: **FAIL** — the bare-name lookup marks a dead key live via a collision that exists in the
tree today. The collision is real and measured: `WorkflowWorktreeConfig.AutoMerge` has **no**
production read resolving to it, yet `.AutoMerge` selectors exist on
`internal/github.MergeOptions`, so a bare-name lookup reports it live (spec.md §A.3, §A.6). The
`AutoCreate` pair (`internal/config/types.go:486` / `pkg/models/config.go:172`) is the same shape.
A PASS means the collision defence is decorative.

### C-5 — the `unbound` class actually fires

```bash
# drop the document_management.* R-allowlist entries from the inventory
grep -v '^  *path: document_management\.' internal/config/testdata/shipped_key_inventory.yaml \
  > /tmp/ckh-falsify/no_docmgmt.yaml
diff -q internal/config/testdata/shipped_key_inventory.yaml /tmp/ckh-falsify/no_docmgmt.yaml \
  && { echo "MUTATION NO-OP — grep removed nothing"; exit 1; }
printf '{"Replace":{"internal/config/testdata/shipped_key_inventory.yaml":"/tmp/ckh-falsify/no_docmgmt.yaml"}}' \
  > /tmp/ckh-falsify/overlay5.json
go test -overlay=/tmp/ckh-falsify/overlay5.json \
  -run 'TestShippedConfigKeysHaveReaders' -count=1 -v ./internal/config/
```

Expected: **FAIL**, naming `document_management.*` keys as `unbound` with no allowlist entry. A PASS
means the `unbound` class (spec.md §B.4 fifth bullet, plan.md §F M2 step 3) is not reachable — the
keys are being skipped rather than classified, which is the defect D4 identified.

## §D Definition of Done

- All of AC-CKH-001 through AC-CKH-023 produce their stated observable output.
- All five falsification procedures C-1 through C-5 produce **FAIL** against unfixed code, and no
  procedure's `diff -q` mutation guard reports a no-op.
- `make build` has been run after every `internal/template/templates/**` edit (plan.md §D2).
- `internal/template/internal_content_leak_test.go` is unmodified (AC-CKH-019).
- `progress.md` §E.2 cites the observed command output for every claim, per
  `.claude/rules/moai/core/verification-claim-integrity.md`.
