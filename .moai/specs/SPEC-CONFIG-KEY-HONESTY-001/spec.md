---
id: SPEC-CONFIG-KEY-HONESTY-001
title: "config surface honesty: every key shipped to a user must be parsed, read, and enforced by the thing it claims to control — or be explicitly marked as not"
version: "0.1.0"
status: draft
created: 2026-07-31
updated: 2026-07-31
author: manager-spec
priority: P1
phase: "v3.0.2"
module: "internal/config"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "config, template, dead-config, parity-guard, deprecation, neutrality, quality-yaml, system-yaml, anti-rot"
related_specs: [SPEC-UPDATE-REINSTALL-LOOP-002, SPEC-UPDATE-DATA-SURVIVAL-001, SPEC-CONFIG-TIER-PERSIST-001, SPEC-UPDATE-YAML-PRESERVE-001, SPEC-V3R2-RT-005, SPEC-WORKTREE-ENTRY-STRATEGY-001, SPEC-AGENT-ARCH-V2-001]
depends_on: [SPEC-CONFIG-TIER-PERSIST-001]
---

# SPEC-CONFIG-KEY-HONESTY-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-07-31 | Initial draft. Epic SPEC 4 of 6 from the four-lens audit of `moai update` / `.moai/config`. Findings F1-F7 each re-verified against HEAD `1d4e4f7da` (branch `main`) while authoring; F3 independently re-derived; one drift recorded (§A.8). |

## §A Problem / Motivation

A configuration key that a user can edit is a promise. `.moai/config/sections/*.yaml` ships hundreds
of such promises to every project created by `moai init`. This SPEC concerns the subset that are
false: keys that are shipped and editable but that no code parses, no code reads, or whose stated
effect is actually decided elsewhere by a hardcoded constant.

The failure mode is silent in every direction. There is no parse error, no warning, no log line —
the user edits the key, the behaviour does not change, and nothing in the system indicates why.
This is worse than an unsupported key, because an unsupported key at least fails loudly.

Sibling SPECs own adjacent halves and are not restated here: `SPEC-CONFIG-TIER-PERSIST-001` (E3)
owns tier precedence, zero-value merging, atomic writes, and the loader's swallow-into-defaults
behaviour for keys that *are* in the struct; `SPEC-UPDATE-YAML-PRESERVE-001` owns the YAML merge
engine; `SPEC-UPDATE-DATA-SURVIVAL-001` (E2) owns backup coverage. **This SPEC owns what happens
when a key is absent from the struct entirely, or present in the struct but read by nobody.**

### A.1 A function named for a schema it does not use (F1)

`pkg/models/config.go:199` declares the correct schema:

```go
type FullQualityConfig struct {
	Constitution     QualityConfig    `yaml:"constitution"`
	ReportGeneration ReportGeneration `yaml:"report_generation"`
	LSPStateTracking LSPStateTracking `yaml:"lsp_state_tracking"`
}
```

It is never instantiated. Verified against HEAD `1d4e4f7da`:

```
$ grep -rn 'FullQualityConfig' --include='*.go' internal pkg cmd | grep -v '_test.go'
internal/lsp/hook/gate.go:125:// parseFullQualityConfig parses the full quality config from YAML data using pkg/models
internal/lsp/hook/gate.go:126:func parseFullQualityConfig(data []byte) (models.QualityConfig, error) {
internal/lsp/hook/gate.go:312:	return parseFullQualityConfig(data)
pkg/models/config.go:198:// FullQualityConfig represents the complete quality.yaml structure.
pkg/models/config.go:199:type FullQualityConfig struct {
```

The only two production hits are the *type declaration* and a *comment*. The function named for it
parses a narrower wrapper (`qualityFileWrapper`, `internal/config/types.go:1174`) and returns the
constitution block alone:

```go
func parseFullQualityConfig(data []byte) (models.QualityConfig, error) {
	var wrapper qualityFileWrapper          // NOT FullQualityConfig
	if err := yaml.Unmarshal(data, &wrapper); err != nil { ... }
	return wrapper.Constitution, nil        // constitution only
}
```

Consequence: `report_generation` (`quality.yaml.tmpl:198`, 4 keys), `lsp_state_tracking`
(`:212`, 12 keys), `constitution.memory_guard` (`:190`, 5 keys), and
`constitution.session_effort_default` (`:10`) are shipped and silently discarded. The last is a
cost lever — `session_effort_default: "xhigh"` reads as the knob a user lowers to reduce spend.
It reaches nothing:

```
$ grep -rn 'session_effort_default\|memory_guard\|SessionEffortDefault\|MemoryGuard' \
    --include='*.go' internal pkg cmd | grep -v '_test.go' | grep -v main-fork
(no output)
```

### A.2 A registry entry asserting a binding that does not exist (F2)

`internal/config/audit_registry.go:31` maps `"system"` to `"SystemConfig"`. There is no loader:

```
$ grep -rn 'loadSystemSection' internal/config/
(no output)
```

`SystemConfig` (`internal/config/types.go:219-228`) declares `version`, `log_level`, `log_format`,
`no_color`, `non_interactive`, `migrations`, `hook`. The shipped
`internal/template/templates/.moai/config/sections/system.yaml.tmpl` declares four different
top-level blocks:

```
$ grep -nE '^[a-z_]+:' internal/template/templates/.moai/config/sections/system.yaml.tmpl
4:moai:
23:github:
34:hook:
50:document_management:
```

Only `hook` overlaps, and even that is not bound through the struct, because `Loader.Load()` never
reads this file at all. `cfg.System` is populated solely from `internal/config/defaults.go` plus
the env overrides in `internal/config/manager.go`. The file *is* read — by ad-hoc parsers that
bypass the struct entirely: `internal/cli/hook.go:508` (`isHookOptInEnabled`, an inline anonymous
struct) and `internal/cli/update.go:989`.

Roughly 25 keys under `github.*` and `document_management.*` therefore have no implementation.
`document_management` is the sharpest case: it states a retention policy, which is a promise about
files being deleted. Nothing deletes them.

The compounding defect is that `TestAuditParity` reports this mapping GREEN. The parity guard
checks **name registration**, not **binding reachability** — so the one mechanism that exists to
catch this class of defect is the mechanism concealing it.

### A.3 One hundred twenty-two field names with no production reader (F3)

An audit lens parsed the `yaml:`-tagged fields of `internal/config/types.go` and searched every
production `.go` file for each. This SPEC re-derived the measurement independently at HEAD
`1d4e4f7da`, deduplicating by Go field name (371 `yaml:` tags collapse to 287 distinct field
names) and searching for `\.<FieldName>\b` across all non-test `.go` files under
`internal pkg cmd`, excluding `main-fork/`:

```
distinct Go field names: 287
zero production reads (.Field): 122
reads only inside types.go: 5
```

Mapping those 122 back onto the shipped template YAMLs gives **121 keys that a user can edit today
with no Go reader**, concentrated in seven families:

| shipped section | dead keys |
|---|---:|
| `design.yaml` | 29 |
| `harness.yaml` | 17 |
| `research.yaml` | 17 |
| `git-strategy.yaml.tmpl` | 15 |
| `constitution.yaml` | 12 |
| `context.yaml` | 12 |
| `workflow.yaml` | 10 |
| `interview.yaml` | 4 |
| `sunset.yaml` | 2 |
| `system.yaml.tmpl` / `security.yaml` / `mx.yaml` | 1 each |

Two methodology facts bind any re-derivation. First, `main-fork/` is a gitignored, untracked local
clone (`git ls-files main-fork` → 0 while the directory exists on disk); including it falsely marks
dozens of fields live, so the inventory must walk tracked files, not the filesystem. Second, a
smaller bucket of fields is read only through accessors declared inside `types.go` itself. At least
one such field is genuinely live — `GateTimeouts.Vet` reaches `internal/hook/pre_tool.go:657` via
`GateConfig.VetTimeoutDuration()` — so accessor-only fields are **unknown, not dead**, and must be
resolved by following the accessor rather than by the direct-read count.

### A.4 "No Go code reads this" is not "dead" (the honesty constraint)

Some of these keys may be consumed by **agent, skill, or rule prose** rather than by Go: an agent
body instructing the orchestrator to read `interview.max_rounds` is a real consumer even though no
Go code touches the key. The audit lens attempted a word-match over the shipped prose corpus and
discarded the result, because bare leaf keys like `search`, `performance`, `evolution`, and
`escalation` have unusable signal-to-noise.

Probing at HEAD `1d4e4f7da` shows the discarded method failed for a fixable reason — it matched the
wrong token. Fixed-string search for the **fully-qualified dotted key path** across
`.claude/agents`, `.claude/skills`, `.claude/rules`:

```
design.evolution.max_active_learnings -> 0
workflow.worktree.auto_cleanup        -> 1
research.budget_cap_tokens            -> 0
interview.max_rounds                  -> 0
harness.escalation                    -> 0
```

versus bare leaf keys over the same corpus:

```
bare 'max_rounds'  -> 5
bare 'escalation'  -> 46
bare 'adaptation'  -> 6
```

The dotted path is a high-precision discriminator (0-1 hits) where the bare leaf is noise (up to
46). That precision gap is the mechanism this SPEC builds its triage rule on (§B REQ-CKH-005) and
the reason the mechanical guard must never delete a key merely because Go does not read it.

### A.5 A key whose documented effect is decided by two constants (F4)

`design.evolution.max_active_learnings` is declared at `internal/config/types.go:1092` and
defaulted to 50 at `internal/config/defaults.go:753`. Enforcement lives elsewhere, twice:

```
internal/evolution/types.go:170:      MaxActiveLearnings = 50
internal/evolution/learning.go:65:    if active >= MaxActiveLearnings {
internal/constitution/rate_limiter.go:14:   rateLimitMaxActiveLearnings = 50
internal/constitution/rate_limiter.go:112:  if activeCount >= rateLimitMaxActiveLearnings {
```

Raising the config to 200 changes nothing. This is the sharpest shape of the defect: not an unused
key, but a key that *appears* to control a behaviour which is in fact hardcoded — and hardcoded
twice, in two packages, with no shared constant.

### A.6 Worktree toggles: two unread, one read for wording only (F5)

```
$ grep -rn 'AutoCleanup\|AutoMerge\|AutoCreate' --include='*.go' internal cmd pkg \
    | grep -v '_test.go' | grep -v main-fork
internal/config/types.go:485-487        (declarations)
internal/config/defaults.go:545-547     (defaults)
internal/cli/worktree_advisory.go:29    (readWorktreeAutoCreate)
```

`AutoCleanup` and `AutoMerge` have zero reads. `AutoCreate` is read exactly once, and only to
choose between two advisory sentences (`internal/cli/worktree_advisory.go:29-60`) — it does not
gate worktree creation. `CLAUDE.local.md` §22.8 describes all three as governing web-console
worktree automation.

### A.7 A shipped comment describing unimplemented behaviour (F7)

`internal/template/templates/.moai/config/sections/workflow.yaml:39` ships
`session_name_pattern: "moai-{ProjectName}-{SPEC-ID}"`. `WorkflowWorktreeConfig.SessionNamePattern`
(`internal/config/types.go:488`) has no production reader — its only non-default hits are three
test files asserting the default value. No code builds a session name from it.

### A.8 Drift recorded while authoring

The audit did not note this: the shipped template and the Go defaults **disagree** on the worktree
toggles. `internal/template/templates/.moai/config/sections/workflow.yaml:36-37` ships
`auto_merge: true` and `auto_cleanup: true`, while `internal/config/defaults.go:545-547` sets all
three to `false` per the EnterWorktree-first policy recorded in `CLAUDE.local.md` §22.8. Because
neither key is read, the contradiction has no runtime effect today — but it means the shipped file
states the opposite of the recorded policy, and any future wiring would inherit the wrong value.
This SPEC folds the contradiction into REQ-CKH-009.

### A.9 Three neutrality leaks the CI guard structurally cannot see (F6)

```
$ grep -rnE 'SPEC-[A-Z0-9]' internal/template/templates/.moai/config/
.../sections/workflow.yaml:39:  session_name_pattern: "moai-{ProjectName}-{SPEC-ID}"
.../sections/workflow.yaml:65:  # cycle (plan.md §D D6, SPEC-AGENT-ARCH-V2-001 M3b). Values mirror the
.../sections/workflow.yaml:85:  # model_routing_profiles: No-Haiku 3-tier policy (SPEC-AGENT-ARCH-V2-001
.../sections/statusline.yaml:28:  # "📋 [<command> <SPEC-ID>-<stage>]" only when
.../sections/cache.yaml:6:# `/moai run SPEC-XXX`.
.../sections/cache.yaml:18:  # Per-SPEC breakpoint TTL applied on `/moai run SPEC-XXX`. Enum: "5m" | "off".
$ grep -rn 'issue #' internal/template/templates/.moai/config/
.../sections/llm.yaml:179:  # (issue #653). Claude Code reports context_window_size based on the
```

`{SPEC-ID}`, `<SPEC-ID>`, and `SPEC-XXX` are generic placeholders and are not leaks. The three
genuine leaks are `workflow.yaml:65`, `workflow.yaml:85` (both citing the internal SPEC ID
`SPEC-AGENT-ARCH-V2-001`, one of them alongside the internal artifact citation `plan.md §D D6`),
and `llm.yaml:179` (`issue #653`).

The guard misses them by construction. `internal/template/internal_content_leak_test.go:276-283`
matches only `SPEC-(DB-SYNC-RELOC|PROJECT-DB-HINT)-[0-9]{3}`, and the tree-wide C1 class covers only
`SPEC-(V3R[2-6]|AGENCY|WORKTREE)-`. `SPEC-AGENT-ARCH-V2-001` belongs to no registered family. The
C6 class matches `PR #N` but not `issue #N`, and no class covers internal artifact citations of the
`plan.md §D D6` shape. The enumerated-family design is deliberate (a generic
`SPEC-[A-Z-]+-[0-9]+` wildcard would flag pedagogical placeholders throughout skill bodies), so the
pattern-coverage question is a real design decision — and it belongs to sibling E5, not here.

Measured clean in the same tree: `/Users/` paths, `REQ-` / `AC-` tokens, `CLAUDE.local` references,
ISO dates, and commit SHAs are all zero across `internal/template/templates/.moai/config/`.

## §B Requirements (GEARS)

### B.1 Parse-reachability

**REQ-CKH-001** — The `moai` binary shall parse every top-level block declared in the shipped
`quality.yaml` into a Go value that at least one production code path can read, or shall not ship
that block.

**REQ-CKH-002** — The `parseFullQualityConfig` function shall either unmarshal into
`models.FullQualityConfig` and return all three blocks, or be renamed to state the narrower scope it
actually parses. **When** a reviewer reads the function name, the name shall not assert a schema the
body does not use.

**REQ-CKH-003** — **Where** a shipped `quality.yaml` key group has no production consumer after
REQ-CKH-001 is applied (`report_generation`, `lsp_state_tracking`, `constitution.memory_guard`,
`constitution.session_effort_default`), the key group shall carry the §B.4 reserved marker or be
removed under the §B.6 deprecation posture.

### B.2 Registry honesty

**REQ-CKH-004** — The `yamlToStructRegistry` entry for a section shall assert a binding that a
production loader actually performs. **Where** a section has no such loader, its entry shall live in
`yamlAuditExceptions` with a reason naming the actual readers. Specifically, the `"system"` entry
shall stop asserting `SystemConfig` binding while `Loader.Load()` does not read `system.yaml`, and
the shipped `system.yaml` blocks with no implementation (`github.*`, `document_management.*`) shall
be classified under §B.3.

### B.3 Triage rule

**REQ-CKH-005** — The repository shall carry a written, mechanically applicable four-way triage rule
classifying every shipped config key as **W** (wire up), **P** (prose-consumed), **R** (reserved,
not yet implemented), or **D** (delete). The rule shall define its prose-consumer discriminator as a
fixed-string search for the **fully-qualified dotted key path** across the shipped prose corpus, and
shall state explicitly that a bare leaf-key match is a homonym and not evidence of consumption.

**REQ-CKH-006** — The triage rule shall be applied in full to the highest-impact families measured
in §A.3 (`design`, `harness`, `research`, `git-strategy`, `constitution`, `context`, `workflow`),
and every remaining shipped key shall carry a classification in a tracked inventory file so the
list cannot rot silently.

**REQ-CKH-007** — **When** the triage rule classifies a key as **P**, the shipped YAML shall carry a
comment naming the prose file that consumes it, and the classification shall not be revoked by any
Go-reader-based analysis.

### B.4 Anti-rot guard

**REQ-CKH-008** — The test suite shall contain a guard that fails **when** a key shipped in a
template section YAML has neither a production Go reader, nor a registered prose consumer, nor an
explicit reserved marker. The guard shall:

- build its inventory from git-tracked files only, excluding `main-fork/`;
- key its reader lookup on the dotted key path resolved to a struct field path, not on a bare Go
  field name, so that a field-name collision across two structs cannot bias a key toward live;
- follow accessor indirection, treating a field read only by a `types.go` accessor as live **while**
  that accessor itself has a production caller, and as unresolved otherwise;
- assert a non-empty, plausible inventory (see NFR-CKH-002) so that an inventory of zero fails
  rather than passes.

### B.5 Documented-but-unenforced reconciliation

**REQ-CKH-009** — For each of `design.evolution.max_active_learnings`,
`workflow.worktree.auto_cleanup`, `workflow.worktree.auto_merge`, `workflow.worktree.auto_create`,
and `workflow.worktree.session_name_pattern`, the repository shall either route the enforcement
through the config value, or correct every document that claims the key has effect — including the
shipped YAML comment and `CLAUDE.local.md` §22.8 — so that no surface describes an effect the code
does not produce. **Where** the key is not wired, the shipped template value shall not contradict
`internal/config/defaults.go` (§A.8).

**REQ-CKH-010** — **Where** `max_active_learnings` remains unwired, the two independent hardcoded
ceilings (`internal/evolution/types.go:170`, `internal/constitution/rate_limiter.go:14`) shall be
documented as the enforcement site at the config declaration, so a reader of the config cannot
mistake it for the lever.

### B.6 Deprecation posture and neutrality

**REQ-CKH-011** — The `moai` binary shall not remove a shipped config key from a user's existing
`.moai/config` as a side effect of an update. Removal shall apply to the **template** only; a
user-set key that disappears from the template shall survive in the user's file (per the
`SPEC-UPDATE-YAML-PRESERVE-001` merge contract) and shall be reported once, not deleted.

**REQ-CKH-012** — The `moai` binary shall not ship internal SPEC IDs, internal artifact citations,
or internal issue numbers in `internal/template/templates/.moai/config/**`. The three occurrences
identified in §A.9 shall be removed from the shipped content.

**REQ-CKH-013** — The pattern-coverage gap that let §A.9's leaks pass shall be handed to sibling E5
with the measured evidence; this SPEC shall not modify
`internal/template/internal_content_leak_test.go`.

### B.7 Non-functional

**NFR-CKH-001** — Every test added by this SPEC shall confine its filesystem writes to `t.TempDir()`.

**NFR-CKH-002** — The REQ-CKH-008 guard shall assert its inventory contains at least 200 shipped
keys and at least 200 struct fields, so a silently-empty inventory fails.

**NFR-CKH-003** — Go sources added by this SPEC shall use `snake_case.go` filenames, wrap errors
with `fmt.Errorf("...: %w", err)`, and carry English comments and godoc.

## §C Exclusions

### Out of Scope — sibling-owned config behaviour

- Tier precedence, `SrcLocal` ordering, zero-value merging, atomic and mode-preserving writes, and
  the malformed-config contract. Owned by `SPEC-CONFIG-TIER-PERSIST-001` (E3).
- The YAML merge engine's node-level preservation semantics. Owned by
  `SPEC-UPDATE-YAML-PRESERVE-001`.
- Backup coverage and the update failure contract. Owned by `SPEC-UPDATE-DATA-SURVIVAL-001` (E2).
- v2 detection and deprecated-path handling. Owned by `SPEC-UPDATE-REINSTALL-LOOP-002` (E1).

### Out of Scope — CI guard pattern design

- Extending, generalising, or restructuring the leak-detection patterns in
  `internal/template/internal_content_leak_test.go`, and the policy question of whether
  `SPEC-[A-Z-]+-[0-9]+` should become a wildcard class. Handed to sibling E5 per REQ-CKH-013. This
  SPEC removes the leaked content only.

### Out of Scope — implementing the capabilities the dead keys name

- Building the features that `design.*`, `research.*`, `interview.*`, or `document_management.*`
  describe. A key classified **R** (reserved) is marked, not implemented. Wiring is limited to the
  named cases in REQ-CKH-009 and to keys the triage rule classifies **W**.
- Any change to the `.moai/config/local/` gitignore gap recorded in `CLAUDE.local.md` §22.9.

### Out of Scope — runtime and distribution mechanics

- `moai update` control flow, clean-reinstall sequencing, and template deployment ordering.
- `.claude/settings.json` / `settings.local.json` key surfaces; this SPEC covers
  `.moai/config/sections/**` only.
- The 16-language template rule catalogue itself (`CLAUDE.local.md` §15/§25); this SPEC complies with
  it and does not amend it.
