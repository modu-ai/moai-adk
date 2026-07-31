# SPEC-CONFIG-KEY-HONESTY-001 — Implementation Plan

Milestones are ordered by **decision-reversibility**: the classification contract and the guard
contract come first because they are the decisions most likely to change under review, and the
mechanical edits that consume them come last.

## §A Context

Code baseline `d5336214e`, branch `plan/epic-update-config-audit` (merged with `origin/main`); the
worktree HEAD is a descendant that changes SPEC documents only. All `file:line` references in `spec.md` §A were
re-verified against this tree while authoring; one drift was found and recorded (spec.md §A.8 — the
shipped `workflow.yaml` worktree toggles contradict `internal/config/defaults.go`).

Two measurements this plan depends on, both re-derived at code baseline `d5336214e`:

- 287 distinct `yaml:`-tagged Go field names in `internal/config/types.go`; **174** with zero
  production reads and 4 read only through a `types.go` accessor, measured **path-resolved** (every
  `X.F` selector resolved to its receiver type; a read counts only when the receiver is a
  `types.go`-declared struct). 161 of the 174 surface as a key in a shipped section YAML, across 188
  (file, key) occurrences. The superseded bare-field-name figures were 122 / 5 / 121; 43 names flip
  live→dead under path resolution, `auto_merge` among them. See spec.md §A.3.
- Fixed-string dotted-path prose search yields 0-1 hits per key; bare leaf-key search yields up to
  46. This precision gap is the load-bearing mechanism of the M1 triage rule.

Both measurements share one root cause and one consequence. The root cause is that a **bare name**
— whether a Go field name or a YAML leaf key — is not an identifier; only the fully-qualified path
is. The consequence is that M2's guard must reproduce the path-resolved number mechanically rather
than trust the figure recorded above: the guard is the durable implementation of §A.3's probe, so a
divergence between the guard's output and 174 is a finding about one of the two, not a licence to
adjust the SPEC to match.

## §B Known issues carried into this SPEC

1. The parity guard (`TestAuditParity`) checks name registration, not binding reachability, and
   therefore currently reports the `"system"` mapping GREEN while no loader exists. Any new guard
   must not repeat this shape — see M2 D3.
2. `main-fork/` exists on disk and is untracked (`git ls-files main-fork` → 0). Every inventory step
   must walk git-tracked paths.
3. This checkout is shared with concurrent sessions. `git stash` is prohibited; falsification uses
   `go test -overlay` or a scratch worktree driven by `go -C` (acceptance.md §C).

## §C Pre-flight

```bash
git rev-parse --short HEAD                       # expect d5336214e or a recorded successor
go build ./... && go vet ./internal/config/...
go test -count=1 ./internal/config/... ./internal/template/... ./internal/lsp/...
```

## §D Constraints

- **D1 — template neutrality.** No internal SPEC ID, REQ/AC token, internal date, commit SHA, or
  internal artifact citation may enter `internal/template/templates/**`. A reserved marker must
  therefore be phrased generically ("not yet implemented"), never with a SPEC ID.
- **D2 — Template-First.** Any template edit is followed by `make build` before the change is
  observable in a local project.
- **D3 — no user-key deletion.** Removing a key from the template must not delete it from an
  existing user file; the `SPEC-UPDATE-YAML-PRESERVE-001` merge contract preserves user keys, and
  this SPEC relies on that rather than adding a removal path.
- **D4 — test isolation.** `t.TempDir()` only (NFR-CKH-001).
- **D5 — guard must not delete prose-consumed keys.** The M2 guard reports; it never edits.

## §E Self-verification

Each milestone closes only when its acceptance.md criteria print the stated observable output. §E
of `progress.md` carries the run-phase evidence.

NFR-CKH-003 (snake_case filenames, `%w` error wrapping, English comments) is cross-cutting rather
than milestone-scoped and is absorbed into AC-CKH-021, which checks it over the set of `.go` files
this SPEC adds. The check is scoped to added files deliberately: `gofmt -l internal pkg cmd` lists
114 pre-existing files at the code baseline (alignment-only diffs from a different toolchain
version), so a repository-wide formatting expectation would be unsatisfiable rather than a gate.

## §F Milestones

### M1 — Triage rule and full key inventory (REQ-CKH-005, REQ-CKH-006, REQ-CKH-007)

The highest-change-likelihood decision in this SPEC: the classification contract that every later
milestone consumes.

**Deliverable 1 — the rule**, written to `.moai/docs/config-key-triage-rule.md` (dev-facing, not
shipped, so it may cite SPEC IDs):

| Class | Definition | Action |
|---|---|---|
| **W** — wire | The key states a behavioural promise a user could act on, and routing the existing enforcement through the config value is a bounded change. | Implement the read. |
| **P** — prose-consumed | A shipped prose file references the key by its **fully-qualified dotted path**, and that file also references `.moai/config` (actionable-prose co-occurrence). | Keep; annotate the YAML with the consuming file; register in the guard's prose allowlist. |
| **R** — reserved | No reader and no prose consumer, but the key names a real intended capability. | Keep; add the generic reserved marker; register in the guard's reserved allowlist. |
| **D** — delete | No reader, no prose consumer, no intended capability — **or** the key actively lies (its stated effect is enforced elsewhere by a constant and wiring is out of scope). | Remove from the template under §B.6 posture. |

**The prose discriminator** (the honesty constraint, spec.md §A.4): search the shipped prose corpus
— `.claude/agents`, `.claude/skills`, `.claude/rules`, `.claude/commands`, plus their
`internal/template/templates/` mirrors — with `grep -rF` for the **dotted key path**
(`<section>.<parent>.<leaf>`). A bare leaf-key match is a **homonym and is not evidence**: measured
at code baseline `d5336214e`, bare `escalation` matches 46 prose files while `harness.escalation` matches 0.
A second filter requires the matching file to also contain the literal `.moai/config`, so a passing
mention of the concept does not qualify as an instruction to read the key.

Classification order is fixed: **P before D**. A key is never classified **D** on Go-reader evidence
alone; the prose probe must return zero first (REQ-CKH-007).

**Deliverable 2 — the inventory**, `internal/config/testdata/shipped_key_inventory.yaml`: one entry
per shipped key with `path`, `class`, `evidence` (reader file:line, prose file, or `none`), and for
**D** entries a `deprecate_after` version. The seven highest-impact families named in REQ-CKH-006
(`design`, `harness`, `research`, `git-strategy`, `constitution`, `context`, `workflow` — still the
top seven under the path-resolved counts in spec.md §A.3) are classified individually; every
remaining shipped key carries at minimum a class so the list cannot rot.

**Decision to surface at review**: whether `git-strategy.yaml.tmpl` (15 dead keys) is
prose-consumed by the git workflow rules or genuinely dead — it is the family most likely to be
misclassified, because git workflow doctrine is heavily documented in prose that does not cite the
config path.

### M2 — Anti-rot guard (REQ-CKH-008, NFR-CKH-002)

The second contract decision: what the guard treats as evidence of liveness.

`internal/config/shipped_key_reader_test.go`, `TestShippedConfigKeysHaveReaders`:

1. Enumerate shipped keys by parsing every `internal/template/templates/.moai/config/sections/*.yaml*`
   into dotted paths. Source the file list from `git ls-files`, not `os.ReadDir` (B2).
2. Resolve each dotted path to a **struct field path** by walking `Config` reflectively from the
   section root, so `Worktree.AutoCreate` and an unrelated `AutoCreate` on a different struct are
   distinct lookups. This is the field-name-collision defence: the guard keys on the path, never on
   the bare name.
3. Classify each key into one of **five** classes. Four apply once the dotted path resolves to a
   struct field: **direct-live** (a production read of that field outside `types.go`);
   **accessor-live** (read only by a `types.go` accessor **while** that accessor has a production
   caller — `GateTimeouts.Vet` → `GateConfig.VetTimeoutDuration()` → `internal/hook/pre_tool.go:657`
   is the reference case); **unresolved** (accessor exists but has no production caller); **dead**
   (no read resolves to the field). The fifth applies when step 2's walk finds **no field at all**
   for the dotted path: **unbound**. `unbound` is a distinct class, not a synonym for `unresolved`
   — `unresolved` presupposes a resolved field. Without it, the 25 keys under `github.*` and
   `document_management.*` never enter the partition and are skipped in silence, so REQ-CKH-008
   would miss its own headline case (spec.md §A.2). The `moai.*` block resolves to no `SystemConfig`
   field either, but has three ad-hoc readers recorded as its M1 evidence, so it is `unbound` with
   an allowlist entry rather than a failure.
4. Fail on any **dead**, **unresolved**, or **unbound** key not present in the M1 inventory's **P**
   or **R** allowlists.
5. **Non-vacuity**: fail if the shipped-key inventory has fewer than 200 entries or the reflective
   struct walk yields fewer than 200 fields (NFR-CKH-002). A guard that inventories zero must fail,
   not pass.

Falsification (acceptance.md §C): remove one **R** marker via `go test -overlay` and observe the
guard FAIL naming that key.

### M3 — `quality.yaml` resolution (REQ-CKH-001, REQ-CKH-002, REQ-CKH-003)

1. Decide per key group using the M1 rule. Expected outcome, subject to the prose probe:
   `session_effort_default` → **W** (a cost lever with a real consumer in model routing) or **D**;
   `memory_guard` → **R**; `report_generation` and `lsp_state_tracking` → **R**.
2. Fix the misnomer either way (REQ-CKH-002): if the blocks are parsed, switch
   `parseFullQualityConfig` to `models.FullQualityConfig` and return all three; if not, rename it to
   `parseQualityConstitution` and update `internal/lsp/hook/gate.go:312`.
3. Apply the resulting markers to `quality.yaml.tmpl` (D1: generic phrasing), then `make build`.

### M4 — `system.yaml` resolution (REQ-CKH-004)

The registry must stop asserting a binding that does not exist. Chosen decision, with its rationale
recorded here for review:

- **Move `"system"` from `yamlToStructRegistry` to `yamlAuditExceptions`**, with a reason naming the
  two real readers (`internal/cli/hook.go:508` `isHookOptInEnabled`, `internal/cli/update.go:1011`)
  and stating that `Loader.Load()` does not read the file. This is the honest description of today's
  code and immediately unblinds the parity guard.
- **Add a narrow `loadSystemSection`** binding the one block with a genuine consumer (`hook.*`), and
  replace `isHookOptInEnabled`'s inline anonymous struct with it — so the exception can later be
  retired for a real reason rather than a paperwork one.
- **Classify `github.*` and `document_management.*` through M1.** `document_management` retention is
  the sharpest case: it promises file deletion that nothing performs, so it is **R** with an
  explicit marker at minimum. Implementing retention is out of scope (spec.md §C).

Alternative considered and rejected: adding a full `SystemConfig` loader that binds `moai:`,
`github:`, and `document_management:`. Rejected because it would bind ~25 keys to a struct that
still no code reads — converting an unbound lie into a bound one, with more code.

### M5 — Documented-but-unenforced reconciliation (REQ-CKH-009, REQ-CKH-010)

- `design.evolution.max_active_learnings`: not wired (two independent constants in two packages, a
  refactor beyond this SPEC's scope). Document the real enforcement sites at the config declaration
  (`internal/config/types.go:1092`) and at `internal/config/defaults.go:753` per REQ-CKH-010, and
  classify the key **R** or **D**.
- `workflow.worktree.auto_cleanup` / `auto_merge`: unwired. Correct the shipped template to match
  `internal/config/defaults.go:545-547` (resolving the §A.8 contradiction: shipped `true` vs
  default `false`) and correct `CLAUDE.local.md` §22.8 to state that these two toggles are declared
  but not read.
- `workflow.worktree.auto_create`: read, but only for advisory wording. `CLAUDE.local.md` §22.8 is
  corrected to say so precisely.
- `workflow.worktree.session_name_pattern`: unwired; marked, and the shipped comment corrected so it
  no longer implies session names are built from it.

### M6 — Deprecation posture, neutrality leak removal, E5 handoff (REQ-CKH-011, REQ-CKH-012, REQ-CKH-013)

This milestone carries the whole of spec.md §B.6, REQ-CKH-011 included. REQ-CKH-011 is the guard on
M1's **D** class: template removal is the only deletion this SPEC performs, and REQ-CKH-011 states
what that removal must not do to an existing user file. It is verified by AC-CKH-023 — a
`t.TempDir()` project whose config carries a key the template no longer ships must retain the key
after the merge and be reported exactly once. The merge engine itself stays owned by
`SPEC-UPDATE-YAML-PRESERVE-001` (§D3); what M6 adds is the **report-once, delete-never** assertion
over that engine's behaviour for keys this SPEC removes.

- Rewrite `internal/template/templates/.moai/config/sections/workflow.yaml:65` and `:85` to drop
  `SPEC-AGENT-ARCH-V2-001` and `plan.md §D D6`, preserving the mechanism description.
- Rewrite `llm.yaml:179` to drop `issue #653`, preserving the upstream-misreport explanation.
- `make build`; confirm the §A.9 greps return only the generic `{SPEC-ID}` / `<SPEC-ID>` /
  `SPEC-XXX` placeholders.
- Write the E5 handoff note (`.moai/docs/` or the E5 SPEC's research input) recording the three
  measured evidence points: `SPEC-AGENT-ARCH-V2-001` matches no registered C1/C1c family; C6 matches
  `PR #N` but not `issue #N`; no class covers `plan.md §D D6`-shaped artifact citations.
  **This SPEC does not edit `internal/template/internal_content_leak_test.go`.**

## §G Anti-patterns

- **AP-1 — deleting a key because Go does not read it.** The prose probe must return zero first
  (REQ-CKH-007). This is the failure the audit lens explicitly warned against.
- **AP-2 — a guard that passes on an empty inventory.** NFR-CKH-002 exists because a
  key-inventory test that silently inventories zero keys is green and worthless.
- **AP-3 — bare-field-name reader lookup.** Two structs with an `AutoCreate` field would make one
  key's liveness leak onto the other. M2 step 2 keys on the resolved path.
- **AP-4 — filesystem-walked inventory.** `main-fork/` is on disk and untracked; walking the
  filesystem falsely marks fields live.
- **AP-5 — converting an unbound lie into a bound one.** Adding a loader for keys nothing reads
  (the rejected M4 alternative) satisfies the parity guard while changing nothing for the user.
- **AP-6 — a reserved marker citing a SPEC ID.** D1 forbids internal IDs in shipped templates; the
  marker is generic.
- **AP-7 — `git stash` for falsification.** Prohibited (§B3).

## §H Cross-references

- `.claude/rules/moai/core/verification-claim-integrity.md` — every claim in `progress.md` §E must
  cite an observed command.
- `CLAUDE.local.md` §2 (Template-First), §15 (16-language neutrality), §22.8 (worktree toggle
  intent — corrected by M5), §25 (template internal-content isolation).
- `internal/config/audit_registry.go`, `internal/config/audit_test.go` — the parity guard M4 unblinds.
- `internal/template/internal_content_leak_test.go` — read-only reference for M6; not edited.
