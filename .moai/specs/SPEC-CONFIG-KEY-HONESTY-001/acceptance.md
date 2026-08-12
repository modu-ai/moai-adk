# SPEC-CONFIG-KEY-HONESTY-001 — Acceptance Criteria

## §A Discipline

1. **Every AC states a command and its expected observable output.** A criterion phrased as a
   property with no command is not an AC.
2. **`go test -run <pattern>` exits 0 on zero matches.** Every `-run` AC therefore also requires a
   verbatim `--- PASS: <exact test name>` line. Recorded vacuity baseline at code baseline `ed70e4354`:
   ```
   $ go test -run 'TestShippedConfigKeysHaveReaders' ./internal/config/ ; echo "exit=$?"
   ok  	github.com/modu-ai/moai-adk/internal/config	0.443s [no tests to run]
   exit=0
   ```
   An AC whose only assertion is `exit 0` would pass against a tree with no test at all, and is
   rejected.
3. **Baselines are observed, and attributed to the code baseline `ed70e4354`.** The worktree HEAD at
   which they were run changes SPEC documents only — every `file:line` and count below is
   attributable to `ed70e4354` (iteration-3 refresh; the prior `d5336214e` baseline was 12 days
   stale). Each AC carries its observed pre-change baseline so a reviewer can distinguish a real
   change from a no-op.
4. **Every new guard needs a falsification** proving it FAILS against unfixed code (§C).
5. **`git stash` is prohibited, and falsification does not use a worktree either.** This checkout is
   shared with concurrent sessions: `git stash` is repository-global, and a scratch `git worktree`
   costs a mutation of shared repository state for no benefit here. All five §C procedures therefore
   use `go test -overlay`, which substitutes files for one command invocation and leaves no state
   behind.
6. **All fixtures use `t.TempDir()`** and touch no path outside it (NFR-CKH-001).
7. **A command that cannot observe its own expectation is not an AC.** Clause 1 requires a command;
   this clause requires the command to be capable of seeing what the AC claims. Four criteria were
   rewritten under it — a `-run` selector naming a test that does not exist, a `grep` that cannot
   tell which of two maps an entry sits in, a `grep` whose window excludes the line it verifies, and
   a `grep` with no context flags asserting something about adjacent comments. Each rewrite records
   the observed before/after in place, so the failure mode stays documented rather than merely
   removed. The same test applies to a falsification: a mutation step expressed as a comment mutates
   nothing, which is why every §C procedure now carries a `diff -q` no-op guard.
8. **Near-duplicate ACs sharing a deliverable are consolidated.** The iteration-2 SPEC carried 23
   ACs against the Tier M ceiling of 16. The iteration-3 refresh consolidates to 15 by merging ACs
   that test the same milestone deliverable from the same baseline — every original test assertion,
   baseline observation, and falsification procedure is preserved; only the AC *header count*
   shrinks. The consolidation map: AC-CKH-001+002→001 (M1 rule+inventory); AC-CKH-003+004→002
   (M1 classification quality); AC-CKH-005+006→003 (M2 guard runs non-vacuously);
   AC-CKH-007+008→004 (M2 path-resolved lookup design); AC-CKH-009→005, AC-CKH-010→006 (M3, held);
   AC-CKH-011+012→007 (M4 registry+loader, D4-rewritten); AC-CKH-013→008 (M4 unbound class);
   AC-CKH-014→009 (M5 max_active_learnings); AC-CKH-015+016→010 (M5 worktree toggles, both
   surfaces); AC-CKH-017→011 (M5 session_name_pattern); AC-CKH-018+019→012 (M6 leaks removed+guard
   untouched); AC-CKH-020→013 (M6 E5 handoff); AC-CKH-023→014 (M6 report-once/delete-never);
   AC-CKH-021+022→015 (cross-cutting suite green + NFR + t.TempDir). Every original test name is
   preserved in the merged AC body.

## §B Acceptance criteria

### M1 — Triage rule and inventory

#### AC-CKH-001 — the triage rule exists and the inventory covers every shipped key (was AC-CKH-001 + AC-CKH-002)

```bash
# Part A: the triage rule exists
test -f .moai/docs/config-key-triage-rule.md && \
  grep -cE 'dotted (key )?path|fully-qualified' .moai/docs/config-key-triage-rule.md && \
  grep -c 'homonym' .moai/docs/config-key-triage-rule.md

# Part B: the inventory is complete
go test -run 'TestShippedKeyInventoryIsComplete' -count=1 -v ./internal/config/
```

Expected: Part A — `test -f` succeeds, and both `grep -c` print a count `>= 1`, establishing that
the rule names the dotted-path discriminator and explicitly records that a bare leaf-key match is a
homonym. Part B — a `--- PASS: TestShippedKeyInventoryIsComplete` line. The test parses every
`internal/template/templates/.moai/config/sections/*.yaml*` into dotted paths, loads
`internal/config/testdata/shipped_key_inventory.yaml`, and asserts (a) every shipped key has an
inventory entry, (b) every entry's `class` is one of `W`/`P`/`R`/`D`, (c) every `P` and `W` entry
carries non-empty `evidence`.

Baseline: neither the rule file, the test, nor the inventory exists —
`ls .moai/docs/config-key-triage-rule.md` prints
`ls: .moai/docs/config-key-triage-rule.md: No such file or directory` and exits 1;
`ls internal/config/testdata/shipped_key_inventory.yaml` prints `No such file or directory`, and
the `-run` selector matches nothing (see §A clause 2).

#### AC-CKH-002 — the seven highest-impact families are individually classified and no key is D on Go evidence alone (was AC-CKH-003 + AC-CKH-004)

```bash
go test -run 'TestShippedKeyInventoryFamilyCoverage' -count=1 -v ./internal/config/
go test -run 'TestTriageRuleProseProbeBeforeDelete' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestShippedKeyInventoryFamilyCoverage` line, asserting every dead key in the
`design`, `harness`, `research`, `git-strategy`, `constitution`, `context`, and `workflow` sections
carries an inventory entry whose `evidence` field is populated. Plus a
`--- PASS: TestTriageRuleProseProbeBeforeDelete` line, asserting for every `class: D` entry the
fixed-string dotted-path probe over the shipped prose corpus returns zero matches — re-proving the
**P**-before-**D** ordering rather than trusting the recorded class.

Baseline recorded at code baseline `ed70e4354`, **path-resolved** (spec.md §A.3) — the dead-key
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
```

dead field names total: 174     of which mapping to a shipped key: 161
(file, key) occurrences: 188    accessor-only (types.go): 4

The counts above are **occurrences**, not distinct keys, because the field-name → section-file
mapping remains leaf-name based (spec.md §A.3, third methodology fact). M2's reflective section walk
narrows this. The two figures the guard **must** reproduce exactly are `174` dead field names and
`4` accessor-only.

Prose-probe precision measured at `ed70e4354` over `.claude/agents`, `.claude/skills`,
`.claude/rules`:

```
design.evolution.max_active_learnings -> 0     bare 'max_rounds'  -> 5
workflow.worktree.auto_cleanup        -> 1     bare 'escalation'  -> 46
research.budget_cap_tokens            -> 0     bare 'adaptation'  -> 6
interview.max_rounds                  -> 0
harness.escalation                    -> 0
```

### M2 — Anti-rot guard

#### AC-CKH-003 — the guard passes and is not vacuous (was AC-CKH-005 + AC-CKH-006)

```bash
go test -run 'TestShippedConfigKeysHaveReaders' -count=1 -v ./internal/config/
go test -run 'TestShippedConfigKeysHaveReaders/non_vacuous_inventory' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestShippedConfigKeysHaveReaders` line, and a
`--- PASS: TestShippedConfigKeysHaveReaders/non_vacuous_inventory` line. The non-vacuity subtest
asserts the shipped-key inventory has `>= 900` entries and the reflective walk of `Config` yields
`>= 250` struct fields (NFR-CKH-002), so a guard that inventories zero fails instead of passing.

Baseline: `[no tests to run]`, exit 0 (§A clause 2). For plausibility: at `ed70e4354`,
`grep -c 'yaml:"' internal/config/types.go` prints `371`, and the field-name-deduped walk yields
`287` distinct names — both comfortably above the floor.

#### AC-CKH-004 — the guard resolves paths, not bare field names (was AC-CKH-007 + AC-CKH-008)

```bash
go test -run 'TestShippedConfigKeysHaveReaders/collision_resolution' -count=1 -v ./internal/config/
go test -run 'TestShippedConfigKeysHaveReaders/accessor_indirection' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestShippedConfigKeysHaveReaders/collision_resolution` line, asserting
`workflow.worktree.auto_create` (production reader at `internal/cli/worktree_advisory.go:29`) and a
same-named field on a different struct resolve to distinct lookups. Plus a
`--- PASS: TestShippedConfigKeysHaveReaders/accessor_indirection` line, asserting `GateTimeouts.Vet`
classifies **accessor-live** via `GateConfig.VetTimeoutDuration()` whose production caller is
`internal/hook/pre_tool.go:657`.

Baseline: `pkg/models/config.go:172` already declares a second `AutoCreate bool \`yaml:"auto_create"\``
independent of `internal/config/types.go:544` — the collision exists in the tree today. A
direct-read-only classifier marks `GateTimeouts.Vet` dead — the false positive the audit lens flagged.

### M3 — `quality.yaml` (HELD — pending SPEC-CONFIG-TIER-PERSIST-001)

#### AC-CKH-005 — the parse function no longer misnames its schema (was AC-CKH-009)

```bash
grep -rn 'parseFullQualityConfig\|parseQualityConstitution\|FullQualityConfig' \
  --include='*.go' internal pkg cmd | grep -v '_test.go'
```

Expected: **either** every `parseFullQualityConfig` occurrence is gone (renamed to
`parseQualityConstitution`, with `internal/lsp/hook/gate.go:312` updated), **or** the function body
unmarshals into `models.FullQualityConfig` — i.e. `FullQualityConfig` appears in
`internal/lsp/hook/gate.go`, not only in `pkg/models/config.go`.

Baseline at `ed70e4354`:

```
internal/lsp/hook/gate.go:125:// parseFullQualityConfig parses the full quality config ...
internal/lsp/hook/gate.go:126:func parseFullQualityConfig(data []byte) (models.QualityConfig, error) {
internal/lsp/hook/gate.go:312:	return parseFullQualityConfig(data)
pkg/models/config.go:198:// FullQualityConfig represents the complete quality.yaml structure.
pkg/models/config.go:199:type FullQualityConfig struct {
```

#### AC-CKH-006 — every shipped `quality.yaml` block is parsed or marked (was AC-CKH-010)

```bash
go test -run 'TestQualityYAMLBlocksParsedOrMarked' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestQualityYAMLBlocksParsedOrMarked` line. For each of `report_generation`,
`lsp_state_tracking`, `constitution.memory_guard`, and `constitution.session_effort_default`, the
test asserts the block either round-trips through a production parse path or carries the generic
reserved marker.

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

#### AC-CKH-007 — the registry stops asserting an absent binding and the `hook.*` block binds through a loader (was AC-CKH-011 + AC-CKH-012; D4-rewritten at iteration 3)

```bash
go test -run 'TestAuditParity' -count=1 -v ./internal/config/
go test -run 'TestSystemHookOptInLoadsViaLoader' -count=1 -v ./internal/config/
awk '/^var yamlAuditExceptions/,/^}/'  internal/config/audit_registry.go | grep -c '"system"'
awk '/^var yamlToStructRegistry/,/^}/' internal/config/audit_registry.go | grep -c '"system"'
grep -rn 'loadSystemSection' internal/config/
# Part B: the hook.* inline-struct readers are consolidated into the loader
grep -rn 'var doc struct' internal/hook/routing_ledger.go internal/cli/update.go | grep -c 'Hook struct'
```

Expected: a `--- PASS: TestAuditParity` line, then **either**

- exceptions-block count `1`, registry-block count `0`, and `grep -rn 'loadSystemSection'` printing
  no output — the entry has moved to `yamlAuditExceptions` with a reason naming the real readers;
- **or** exceptions `0`, registry `1`, and `grep -rn 'loadSystemSection'` printing a declaration
  plus a `Loader.Load` call site — a real binding now exists.

Plus a `--- PASS: TestSystemHookOptInLoadsViaLoader` line, and Part B's count printing `0` — the two
inline anonymous structs that read `system.yaml`'s `hook.*` block at HEAD
(`internal/hook/routing_ledger.go:104` `HookObserveOptInEnabled` and `internal/cli/update.go:1140`
`readHookOptInEnabled`) are replaced by the narrow `loadSystemSection` loader.

Baseline at `ed70e4354` (the concealed state) — observed:

```
$ awk '/^var yamlAuditExceptions/,/^}/'  internal/config/audit_registry.go | grep -c '"system"'
0
$ awk '/^var yamlToStructRegistry/,/^}/' internal/config/audit_registry.go | grep -c '"system"'
1
$ grep -rn 'loadSystemSection' internal/config/
(no output)
$ grep -rn 'var doc struct' internal/hook/routing_ledger.go internal/cli/update.go | grep -c 'Hook struct'
2
```

and `TestAuditParity` passes anyway.

**D4 iteration-3 note — why the AC was rewritten.** At the prior baseline `d5336214e`, the AC's
Part B counted `struct {` occurrences inside `isHookOptInEnabled` (`internal/cli/hook.go`) and
expected `3` (three nested inline structs). At HEAD `ed70e4354` that function was refactored
(`e3f8dd463`, SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 M2) into a one-line delegator
`return hook.HookObserveOptInEnabled(projectRoot)` — the awk range returns `0` whether or not M4 is
implemented, so the old AC could not distinguish pre-change from post-change. The inline-struct
readers for `hook.*` survived the refactor but moved to two new sites
(`internal/hook/routing_ledger.go:104` and `internal/cli/update.go:1140`). The rewritten Part B
tracks the inline structs at their actual HEAD locations, so the M4 deliverable (consolidate both
into `loadSystemSection`) is again mechanically observable. The M4 mechanism itself (move `"system"`
to `yamlAuditExceptions`, add narrow `loadSystemSection`) is unchanged.

**Why the Part A commands changed.** The earlier form was `go test … && grep -n '"system"' audit_registry.go`,
which prints line 31 identically whether the entry sits in `yamlToStructRegistry` or in
`yamlAuditExceptions` — it cannot see the one fact the AC is about. The `awk` range restricts each
count to one map, so the pair `(0, 1)` distinguishes the pre-change state from both accepted
post-change states.

#### AC-CKH-008 — `github.*` and `document_management.*` are classified unbound (was AC-CKH-013)

```bash
go test -run 'TestShippedKeyInventoryFamilyCoverage/system' -count=1 -v ./internal/config/
go test -run 'TestShippedConfigKeysHaveReaders/unbound_classification' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestShippedKeyInventoryFamilyCoverage/system` line, asserting every key under
`github:` and `document_management:` in `system.yaml.tmpl` carries an inventory class, and that any
key stating a retention or deletion promise carries the reserved marker. Plus a
`--- PASS: TestShippedConfigKeysHaveReaders/unbound_classification` line, asserting those same 25
keys are reported by the guard in the `unbound` class rather than skipped — and that removing one of
them from the **R** allowlist makes the guard fail (C-5).

Baseline at `ed70e4354`:

```
$ grep -rn 'yaml:"github"\|yaml:"document_management"' --include='*.go' internal pkg cmd
(no output, exit 1)
$ github.* subkeys in system.yaml.tmpl:              2
$ document_management.* subkeys in system.yaml.tmpl: 23
```

The neighbouring `moai:` block (6 subkeys) is deliberately **not** counted here. It also has no
`SystemConfig` field, but unlike the other two it has real readers —
`internal/statusline/version.go:31`, `internal/cli/v2_detection.go:155`, and
`internal/cli/update/plan/plan.go:313` each parse it through an ad-hoc inline `yaml:"moai"` struct.
It is therefore `unbound` with allowlist evidence, not an undetected promise, and the count this AC
governs is 25, not 31.

### M5 — Documented-but-unenforced reconciliation

#### AC-CKH-009 — no surface claims `max_active_learnings` has effect (was AC-CKH-014)

```bash
grep -n -B3 -A1 'max_active_learnings\|MaxActiveLearnings' \
  internal/config/types.go internal/config/defaults.go
```

Expected: both sites carry a comment naming the actual enforcement constants, and both literals
`internal/evolution/types.go` and `internal/constitution/rate_limiter.go` appear **within the
printed context**, so a reader of the config cannot mistake the key for the lever.

Baseline at `ed70e4354`:

```
internal/config/types.go-1229-	GraduationCriteria      DesignGraduationCriteria `yaml:"graduation_criteria"`
internal/config/types.go:1230:	MaxActiveLearnings      int                      `yaml:"max_active_learnings"`
internal/config/types.go-1231-	MaxEvolutionRatePerWeek int                      `yaml:"max_evolution_rate_per_week"`
internal/config/defaults.go-929-			},
internal/config/defaults.go:930:			MaxActiveLearnings:      50,
internal/config/defaults.go-931-			MaxEvolutionRatePerWeek: 3,
```

Neither enforcement-site literal appears. Enforcement sits in two unrelated packages
(`internal/evolution/types.go:170`, `internal/constitution/rate_limiter.go:14`).

#### AC-CKH-010 — the shipped worktree toggles match the Go defaults and CLAUDE.local.md states the real reader status (was AC-CKH-015 + AC-CKH-016)

```bash
# Part A: shipped values match Go defaults
grep -n 'auto_create\|auto_merge\|auto_cleanup' \
  internal/template/templates/.moai/config/sections/workflow.yaml
grep -n 'AutoCleanup\|AutoCreate\|AutoMerge' internal/config/defaults.go | head -3

# Part B: CLAUDE.local.md §22.8 states the real reader status
sed -n '/§22.8/,/§22.9/p' CLAUDE.local.md | grep -cE 'auto_cleanup|auto_merge|auto_create'
```

Expected: Part A — the three shipped values agree with `internal/config/defaults.go:665-667`, and
each unwired toggle carries the generic reserved marker. Part B — a count `>= 3`, with the
surrounding prose stating that `auto_cleanup` and `auto_merge` have no reader and that `auto_create`
is read only to select advisory wording.

Baseline — the contradiction recorded in spec.md §A.8:

```
workflow.yaml:35:        auto_create: false
workflow.yaml:36:        auto_merge: true          <-- defaults.go says false
workflow.yaml:37:        auto_cleanup: true        <-- defaults.go says false
defaults.go:665:			AutoCleanup:        false,
defaults.go:666:			AutoCreate:         false,
defaults.go:667:			AutoMerge:          false,
```

§22.8 currently describes all three as governing web-console worktree automation, which matches no
code path — `internal/cli/worktree_advisory.go:29` is the only read of any of the three, and it
selects between two `fmt.Fprintln` strings.

#### AC-CKH-011 — `session_name_pattern` no longer implies it is used (was AC-CKH-017)

```bash
grep -n -A2 'session_name_pattern' internal/template/templates/.moai/config/sections/workflow.yaml
```

Expected: the key carries the generic reserved marker, so the shipped file no longer presents a
pattern that nothing consumes as an active setting.

Baseline: `workflow.yaml:39` ships `session_name_pattern: "moai-{ProjectName}-{SPEC-ID}"` with no
marker, while `SessionNamePattern` (`internal/config/types.go:546`) has no production reader — its
only non-default hits are three test files asserting the default value.

### M6 — Neutrality and handoff

#### AC-CKH-012 — the three leaks are gone and the existing guard was not edited (was AC-CKH-018 + AC-CKH-019)

```bash
# Part A: the leaks are gone
grep -rnE 'SPEC-(AGENT-ARCH-V2|[A-Z0-9-]*)-[0-9]{3}' internal/template/templates/.moai/config/ ; \
grep -rn 'issue #' internal/template/templates/.moai/config/ ; \
grep -rn 'plan\.md §' internal/template/templates/.moai/config/

# Part B: the existing guard still passes and was not edited
go test -run 'TestTemplateNoInternalContentLeak' -count=1 -v ./internal/template/
git diff --stat ed70e4354 HEAD -- internal/template/internal_content_leak_test.go
```

Expected: Part A — all three greps produce no output (each exits 1). Generic placeholders
`{SPEC-ID}`, `<SPEC-ID>`, and `SPEC-XXX` remain and are unaffected. Part B — a verbatim
`--- PASS: TestTemplateNoInternalContentLeak` line (per §A clause 2), and `git diff --stat`
producing **no output** for that file — REQ-CKH-013 forbids this SPEC from touching it.

Baseline at `ed70e4354` — the three genuine leaks:

```
.../sections/workflow.yaml:82:  # cycle (plan.md §D D6, SPEC-AGENT-ARCH-V2-001 M3b). Values mirror the
.../sections/workflow.yaml:102:  # model_routing_profiles: No-Haiku 3-tier policy (SPEC-AGENT-ARCH-V2-001
.../sections/llm.yaml:179:      # (issue #653). Claude Code reports context_window_size based on the
```

Part B baseline:

```
$ go test -run 'TestTemplateNoInternalContentLeak' -count=1 -v ./internal/template/
=== RUN   TestTemplateNoInternalContentLeak
--- PASS: TestTemplateNoInternalContentLeak (0.76s)
ok  	github.com/modu-ai/moai-adk/internal/template	2.411s
```

**Why the Part B diff is ref-anchored.** The earlier form was `git diff --stat -- <file>` with no
ref, which compares the working tree against the index. A forbidden edit that is **committed** then
produces empty output and the criterion passes — exactly the violation this half exists to catch.
The `ed70e4354 HEAD` anchoring detects a committed edit.

#### AC-CKH-013 — the E5 handoff records the three measured gaps (was AC-CKH-020)

```bash
grep -cE 'SPEC-AGENT-ARCH-V2-001|issue #N|plan\.md §' <handoff-note-path>
```

Expected: a count `>= 3`, covering the three evidence points — the unregistered SPEC family, the
`issue #N` vs `PR #N` C6 asymmetry, and the absent artifact-citation class. `<handoff-note-path>` is
the dev-facing note produced by M6 (its exact path is fixed during run-phase and recorded in
`progress.md` §E.2).

Baseline: no handoff note exists.

#### AC-CKH-014 — a template-removed key survives in the user's file and is reported once (was AC-CKH-023; REQ-CKH-011)

```bash
go test -run 'TestTemplateRemovedKeySurvivesUserConfig' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestTemplateRemovedKeySurvivesUserConfig` line. The test builds a
`t.TempDir()` project whose `.moai/config/sections/<section>.yaml` carries a user-set key that the
current template no longer ships, runs the merge, and asserts three things: (a) the key and its
user-set **value** are still present in the user's file afterwards; (b) the removal is reported
exactly once — a second merge over the same tree emits no further report; (c) no other user-set key
is dropped.

Baseline: the test does not exist, and the `-run` selector matches nothing (§A clause 2).

Scope note: the merge engine remains owned by `SPEC-UPDATE-YAML-PRESERVE-001` (plan.md §D3). This AC
pins the **delete-never, report-once** posture over that engine for the keys M1 classifies **D**.

### Cross-cutting

#### AC-CKH-015 — the full suite is green, new sources meet NFR-CKH-003, and no test writes outside `t.TempDir()` (was AC-CKH-021 + AC-CKH-022)

```bash
# Part A: build, vet, test green
go build ./... && go vet ./... && go test -count=1 ./...

# Part B: added sources meet NFR-CKH-003 (snake_case, gofmt-clean)
git diff --name-only --diff-filter=A ed70e4354 HEAD -- '*.go' > /tmp/ckh-new-go.txt
test -s /tmp/ckh-new-go.txt || { echo "VACUOUS: no added .go files"; exit 1; }
xargs gofmt -l < /tmp/ckh-new-go.txt
xargs -n1 basename < /tmp/ckh-new-go.txt | grep -vE '^[a-z0-9_]+\.go$'

# Part C: no test writes outside t.TempDir()
go test -count=1 ./internal/config/ ./internal/template/ && git status --porcelain
```

Expected: Part A — exit 0 from `go build`, `go vet`, and `go test`; Part B — the `test -s`
non-vacuity guard passing silently, `xargs gofmt -l` printing nothing, and the basename filter
printing nothing (every added source uses a `snake_case.go` filename); Part C — the tests pass and
`git status --porcelain` reports no files created or modified by the test run (NFR-CKH-001).

Baseline at `ed70e4354`: `go build` / `go vet` / `go test` green (pre-flight, plan.md §C);
`git diff --name-only --diff-filter=A ed70e4354 HEAD -- '*.go'` returns **0 files**, so the
non-vacuity guard prints `0` and Part B fails until the SPEC adds a source file.

**Why the check is scoped to added files.** `gofmt -l internal pkg cmd` lists **114** pre-existing
files at the code baseline — alignment-only diffs from a different toolchain version. A
repository-wide `gofmt -l` expectation would therefore be unsatisfiable on an unmodified tree. The
non-vacuity guard is required because the complementary failure is equally real — with no added
files, `gofmt -l` over an empty argument list prints nothing and would pass while proving nothing.

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

The `diff -q` guard is required by §A clause 7: the run-phase inventory's exact formatting is not
fixed at plan time, so a `sed` address that matches nothing leaves the overlay byte-identical to the
original and the procedure would report PASS while having mutated nothing. Observed on a scratch
fixture (three-line file, address deliberately unmatched vs matched):

```
$ sed '/zzz_nomatch/,+2d' orig.yaml > out.yaml
$ if diff -q orig.yaml out.yaml >/dev/null; then echo "MUTATION NO-OP — sed matched nothing"; exit 1; fi; echo "mutation confirmed"
MUTATION NO-OP — sed matched nothing        exit=1

$ sed '/^b$/,+0d' orig.yaml > out2.yaml
$ if diff -q orig.yaml out2.yaml >/dev/null; then echo "MUTATION NO-OP — sed matched nothing"; exit 1; fi; echo "mutation confirmed"
mutation confirmed                           exit=0
```

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

Expected: **FAIL** naming the `>= 900` floor (NFR-CKH-002). A PASS means an empty or truncated
inventory would sail through, which is the exact vacuity hazard AC-CKH-003 exists to exclude.

The `diff -q` guard catches this procedure's own no-op shape (§A clause 7): if the inventory is
already 12 lines or fewer, `head -12` copies it verbatim and the overlay substitutes the file with
itself.

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
overlay identical to the original and the procedure would report PASS while having mutated nothing.

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
`AutoCreate` pair (`internal/config/types.go:544` / `pkg/models/config.go:172`) is the same shape.
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

- All of AC-CKH-001 through AC-CKH-015 produce their stated observable output.
- All five falsification procedures C-1 through C-5 produce **FAIL** against unfixed code, and no
  procedure's `diff -q` mutation guard reports a no-op.
- `make build` has been run after every `internal/template/templates/**` edit (plan.md §D2).
- `internal/template/internal_content_leak_test.go` is unmodified (AC-CKH-012).
- `progress.md` §E.2 cites the observed command output for every claim, per
  `.claude/rules/moai/core/verification-claim-integrity.md`.
