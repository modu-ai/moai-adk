# SPEC-CONFIG-KEY-HONESTY-001 — Acceptance Criteria

## §A Discipline

1. **Every AC states a command and its expected observable output.** A criterion phrased as a
   property with no command is not an AC.
2. **`go test -run <pattern>` exits 0 on zero matches.** Every `-run` AC therefore also requires a
   verbatim `--- PASS: <exact test name>` line. Recorded vacuity baseline at HEAD `1d4e4f7da`:
   ```
   $ go test -run 'TestShippedConfigKeysHaveReaders' ./internal/config/ ; echo "exit=$?"
   ok  	github.com/modu-ai/moai-adk/internal/config	0.443s [no tests to run]
   exit=0
   ```
   An AC whose only assertion is `exit 0` would pass against a tree with no test at all, and is
   rejected.
3. **Baselines were recorded from this tree while authoring** — HEAD `1d4e4f7da`, branch `main`.
   Each AC carries its observed pre-change baseline so a reviewer can distinguish a real change from
   a no-op.
4. **Every new guard needs a falsification** proving it FAILS against unfixed code (§C).
5. **`git stash` is prohibited.** This checkout is shared with concurrent sessions and `git stash`
   is repository-global. Falsification uses `go test -overlay` or a scratch `git worktree` driven by
   `go -C`.
6. **All fixtures use `t.TempDir()`** and touch no path outside it (NFR-CKH-001).

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

Baseline recorded at HEAD `1d4e4f7da` — the dead-key counts these families must cover, derived by
matching each zero-production-read field's YAML key against the shipped section files:

```
design.yaml             29     harness.yaml            17
research.yaml           17     git-strategy.yaml.tmpl  15
constitution.yaml       12     context.yaml            12
workflow.yaml           10     interview.yaml           4
sunset.yaml              2     system.yaml.tmpl / security.yaml / mx.yaml  1 each
shipped dead keys: 121         dead field names total: 122
```

#### AC-CKH-004 — no key is classified D on Go-reader evidence alone

```bash
go test -run 'TestTriageRuleProseProbeBeforeDelete' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestTriageRuleProseProbeBeforeDelete` line. For every inventory entry with
`class: D`, the test runs the fixed-string dotted-path probe over the shipped prose corpus and
asserts zero matches — i.e. it re-proves the **P**-before-**D** ordering rather than trusting the
recorded class.

Baseline (the probe's precision, measured at HEAD `1d4e4f7da` over `.claude/agents`,
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

Baseline for plausibility: at HEAD `1d4e4f7da`, `grep -c 'yaml:"' internal/config/types.go` prints
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

Baseline at HEAD `1d4e4f7da` (the failing state):

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
go test -run 'TestAuditParity' -count=1 -v ./internal/config/ && \
  grep -n '"system"' internal/config/audit_registry.go
```

Expected: a `--- PASS: TestAuditParity` line, and the `"system"` key appearing in
`yamlAuditExceptions` with a reason naming the real readers — **or** appearing in
`yamlToStructRegistry` alongside a `loadSystemSection` function that `Loader.Load` calls.

Baseline at HEAD `1d4e4f7da` (the concealed state): the entry claims a binding,

```
$ grep -n 'system' internal/config/audit_registry.go
31:	"system":         "SystemConfig",
$ grep -rn 'loadSystemSection' internal/config/
(no output)
```

and `TestAuditParity` passes anyway.

#### AC-CKH-012 — the `hook.*` block binds through a loader, not an inline struct

```bash
go test -run 'TestSystemHookOptInLoadsViaLoader' -count=1 -v ./internal/config/ && \
  grep -n 'struct {' internal/cli/hook.go | sed -n '1,3p'
```

Expected: a `--- PASS: TestSystemHookOptInLoadsViaLoader` line, and
`internal/cli/hook.go`'s `isHookOptInEnabled` no longer declaring an inline anonymous struct for
`system.yaml`.

Baseline: `internal/cli/hook.go:508` reads `system.yaml` with `os.ReadFile` into an inline
`var doc struct { Hook struct { OptIn struct { ... } } }`, bypassing `SystemConfig` entirely.

#### AC-CKH-013 — `github.*` and `document_management.*` are classified

```bash
go test -run 'TestShippedKeyInventoryFamilyCoverage/system' -count=1 -v ./internal/config/
```

Expected: a `--- PASS: TestShippedKeyInventoryFamilyCoverage/system` line, asserting every key under
`github:` and `document_management:` in `system.yaml.tmpl` carries an inventory class, and that any
key stating a retention or deletion promise carries the reserved marker.

Baseline: `system.yaml.tmpl` declares `github:` (line 23) and `document_management:` (line 50) with
no Go binding of any kind.

### M5 — Documented-but-unenforced reconciliation

#### AC-CKH-014 — no surface claims `max_active_learnings` has effect

```bash
grep -n 'max_active_learnings\|MaxActiveLearnings' internal/config/types.go internal/config/defaults.go
```

Expected: both sites carry a comment naming the actual enforcement constants
(`internal/evolution/types.go` and `internal/constitution/rate_limiter.go`), so a reader of the
config cannot mistake the key for the lever.

Baseline at HEAD `1d4e4f7da` — declared and defaulted with no such note, while enforcement sits in
two unrelated packages:

```
internal/config/types.go:1092:	MaxActiveLearnings      int  `yaml:"max_active_learnings"`
internal/config/defaults.go:753:			MaxActiveLearnings:      50,
internal/evolution/types.go:170:	MaxActiveLearnings = 50
internal/constitution/rate_limiter.go:14:	rateLimitMaxActiveLearnings = 50
```

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

Baseline at HEAD `1d4e4f7da` — the three genuine leaks:

```
.../sections/workflow.yaml:65:  # cycle (plan.md §D D6, SPEC-AGENT-ARCH-V2-001 M3b). Values mirror the
.../sections/workflow.yaml:85:  # model_routing_profiles: No-Haiku 3-tier policy (SPEC-AGENT-ARCH-V2-001
.../sections/llm.yaml:179:      # (issue #653). Claude Code reports context_window_size based on the
```

alongside the generic placeholders at `workflow.yaml:39`, `statusline.yaml:28`, `cache.yaml:6`,
`cache.yaml:18`, which must survive.

#### AC-CKH-019 — the existing neutrality guard still passes and was not edited

```bash
go test -run 'TestInternalContentLeak' -count=1 ./internal/template/ ; \
git diff --stat -- internal/template/internal_content_leak_test.go
```

Expected: the test package passes, and `git diff --stat` produces **no output** for that file —
REQ-CKH-013 forbids this SPEC from touching it.

Baseline: the guard passes today (it cannot see the three leaks), and the file is unmodified.

#### AC-CKH-020 — the E5 handoff records the three measured gaps

```bash
grep -cE 'SPEC-AGENT-ARCH-V2-001|issue #N|plan\.md §' <handoff-note-path>
```

Expected: a count `>= 3`, covering the three evidence points — the unregistered SPEC family, the
`issue #N` vs `PR #N` C6 asymmetry, and the absent artifact-citation class. `<handoff-note-path>` is
the dev-facing note produced by M6 (its exact path is fixed during run-phase and recorded in
`progress.md` §E.2).

Baseline: no handoff note exists.

### Cross-cutting

#### AC-CKH-021 — the full suite is green and the build is clean

```bash
go build ./... && go vet ./... && go test -count=1 ./...
```

Expected: exit 0 from all three.

Baseline: green at HEAD `1d4e4f7da` (pre-flight, plan.md §C).

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
printf '{"Replace":{"internal/config/testdata/shipped_key_inventory.yaml":"/tmp/ckh-falsify/shipped_key_inventory.yaml"}}' \
  > /tmp/ckh-falsify/overlay.json
go test -overlay=/tmp/ckh-falsify/overlay.json -run 'TestShippedConfigKeysHaveReaders' \
  -count=1 -v ./internal/config/
```

Expected: **FAIL**, with the failure message naming `report_generation.enabled` as a shipped key with
no reader, no prose consumer, and no reserved marker. A PASS here means the guard's allowlist lookup
is not load-bearing and the guard is inert.

### C-2 — the non-vacuity floor actually fires

Same overlay mechanism, substituting an inventory truncated to a handful of entries:

```bash
head -12 internal/config/testdata/shipped_key_inventory.yaml > /tmp/ckh-falsify/tiny.yaml
printf '{"Replace":{"internal/config/testdata/shipped_key_inventory.yaml":"/tmp/ckh-falsify/tiny.yaml"}}' \
  > /tmp/ckh-falsify/overlay2.json
go test -overlay=/tmp/ckh-falsify/overlay2.json \
  -run 'TestShippedConfigKeysHaveReaders/non_vacuous_inventory' -count=1 -v ./internal/config/
```

Expected: **FAIL** naming the `>= 200` floor (NFR-CKH-002). A PASS means an empty or truncated
inventory would sail through, which is the exact vacuity hazard AC-CKH-006 exists to exclude.

### C-3 — the prose probe genuinely blocks a delete

Behavioural check in a scratch worktree, driven by `go -C` (no `cd`, no `git stash`):

```bash
git worktree add /tmp/ckh-wt HEAD
# in the scratch tree only: reclassify one P key to D without removing its prose consumer
go -C /tmp/ckh-wt test -run 'TestTriageRuleProseProbeBeforeDelete' -count=1 -v ./internal/config/
git worktree remove /tmp/ckh-wt
```

Expected: **FAIL**, naming the reclassified key and the prose file whose dotted-path match
contradicts the `D` class. A PASS means the probe trusts the recorded class instead of re-proving
it, and AP-1 (deleting a prose-consumed key) is unguarded.

### C-4 — the collision defence is load-bearing

```bash
go -C /tmp/ckh-wt test -run 'TestShippedConfigKeysHaveReaders/collision_resolution' \
  -count=1 -v ./internal/config/
```

with the guard's path resolution replaced by a bare field-name lookup in the scratch tree.

Expected: **FAIL** — the bare-name lookup marks a dead key live via the
`internal/config/types.go:486` / `pkg/models/config.go:172` `AutoCreate` collision that exists in the
tree today. A PASS means the collision defence is decorative.

## §D Definition of Done

- All of AC-CKH-001 through AC-CKH-022 produce their stated observable output.
- All four falsification procedures C-1 through C-4 produce **FAIL** against unfixed code.
- `make build` has been run after every `internal/template/templates/**` edit (plan.md §D2).
- `internal/template/internal_content_leak_test.go` is unmodified (AC-CKH-019).
- `progress.md` §E.2 cites the observed command output for every claim, per
  `.claude/rules/moai/core/verification-claim-integrity.md`.
