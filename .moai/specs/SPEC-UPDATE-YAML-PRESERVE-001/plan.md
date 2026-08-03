# Plan — SPEC-UPDATE-YAML-PRESERVE-001

Tier **M** (standard). Justification in §A.

## §A Context

### Tier assessment

Tier **M**, matching the prior estimate. The reasoning:

- **File count**: 6-9 files. One production file rewritten (`merge.go`), one new production file (`node_merge.go`), four-plus test files revised (`update_yaml_test.go`, `backup_test.go`, `backup_error_test.go`, `coverage_improvement_test.go`), one new golden test file.
- **Blast radius**: contained to `internal/cli/update/backup` plus its test consumers. No new package, no new dependency, no config schema change, no template change.
- **Risk**: medium — the change is behaviour-visible on every `moai update`, and the exported API is consumed from four packages. But the decision semantics are held constant, which bounds what can go wrong.
- **Not Tier S**: a Tier S SPEC would be a metadata-only or single-file edit with no test-contract work. Here the test contract does not exist and must be built (REQ-UYP-017 through REQ-UYP-019), and a large existing assertion table must be revised.
- **Not Tier L**: no architectural decision is open, no cross-subsystem coordination is needed, and no research phase is required — the mechanism (`yaml.Node`) is already available in the pinned dependency and its behaviour was reproduced directly.

### Verified baseline (re-measured this revision, 2026-08-03, against the worktree tree)

| Claim | Location | Status |
|---|---|---|
| `MergeYAML3Way` map round-trip | `merge.go:20-35` | CONFIRMED |
| `MergeYAMLDeep` map round-trip | `merge.go:116-130` | CONFIRMED |
| Old-only-key drop comment | `merge.go:95-97` | CONFIRMED verbatim |
| 2-way preserves old-only keys | `merge.go:168-171` | CONFIRMED |
| Misnamed subtest at audit (2026-07-31) | `update_yaml_test.go:591-603` | CONFIRMED at audit; **note**: an interim commit on `main` has since rewritten the assertion to preserve `user_added` (the `t.Errorf("expected user_added to be dropped…")` line is gone). M4 becomes a no-op when run-phase begins against the current tree — see M4 note and acceptance §D.1 AC-UYP-006. |
| `SaveTemplateDefaults` base from new template | `backup.go:149-192` | CONFIRMED |
| Zero comment-preservation tests | `grep -rin comment` over both test files | CONFIRMED, 0 matches |
| Round-trip loss | executed against `cache.yaml` | CONFIRMED — 16 comments → 0, keys alphabetized, `"1h"` → `1h`, indent normalized (widens for 2-space sources like `cache.yaml`; narrows for 4-space sources like `archive.yaml`) |
| Template count | `ls …sections/ \| grep -cE '\.yaml(\.tmpl)?$'` | **30** (not 31 — original SPEC miscounted) |
| `llm.yaml` comment-bearing lines | `grep -c '#' llm.yaml` | **119** (plan-audit §5 residual-risk flagged 120; this revision re-measures 119 by the same `grep -c '#'` method) |
| `lsp.yaml.tmpl` / `quality.yaml.tmpl` / `harness.yaml` | same method | 88 / 84 / 65 |
| `quality.yaml.tmpl` unquoted placeholders | `:13` `enforce_quality: {{.EnforceQuality}}`, `:16` `test_coverage_target: {{.TestCoverageTarget}}` | **CONFIRMED** — only `.tmpl` with unquoted placeholders; causes silent 2-way fallback today (see spec.md §A blast radius) |
| M0 coverage baseline | `go test -cover ./internal/cli/update/backup/` | **88.9%** — captured this revision; written into acceptance.md AC-UYP-020 |

### Edge-case survey of the shipped templates

Executed over `internal/template/templates/.moai/config/sections/` (30 files):

| Feature | Files | Consequence for the plan |
|---|---|---|
| Anchors / aliases (`&x` / `*x`) | 0 | REQ-UYP-012 is a defensive contract, not an active migration concern |
| Merge keys (`<<:`) | 0 | REQ-UYP-013 likewise defensive |
| Multi-document (`---`) | 0 | REQ-UYP-015 likewise defensive |
| Block sequences | 8 (`design`, `mx`, `interview`, `harness`, `sunset`, `security`, `lsp.tmpl`, `quality.tmpl`) | REQ-UYP-014 is **live** — sequence policy must be decided, not deferred |
| Heavy comment payload | `llm.yaml` 119, `lsp.yaml.tmpl` 88, `quality.yaml.tmpl` 84, `harness.yaml` 65 | golden test must cover all 30, not a sample |
| Unquoted-placeholder `.tmpl` (unparseable by map decoder) | 1 (`quality.yaml.tmpl`) | REQ-UYP-011/REQ-UYP-017 are **live** — the node decoder parses it where the map decoder does not; see spec.md §A blast radius and Decision D6 |

## §B Known Issues

- The existing `internal/cli/update_yaml_test.go` table (from `TestMergeYAML3Way` at `:471` through `:648`) asserts on `contains(string(result), ...)` substrings. Node output changes whitespace and quoting, so several of these assertions become brittle even where behaviour is unchanged. Budgeted as M5.
- `internal/cli/coverage_improvement_test.go` carries three further `MergeYAML3Way` tests (`:5716`, `:5733`, `:5750`) with the same substring style. Same treatment.
- `DeepMerge3Way` and `DeepMergeMaps` take and return `map[string]any` in their current exported signatures. A node-tree implementation cannot preserve formatting through a map-typed signature — see Decision D2.

## §C Pre-Flight (Run-Phase Entry Checks)

```bash
# 1. Confirm the pinned YAML version has not moved
grep 'gopkg.in/yaml.v3' /Users/goos/MoAI/moai-adk-go/go.mod

# 2. Confirm the baseline suite is green before any edit
go test ./internal/cli/... 2>&1 | tail -20

# 3. Capture the pre-change package coverage as the baseline for the coverage constraint
go test -cover ./internal/cli/update/backup/

# 4. Re-confirm the misnamed subtest is still at the cited location
sed -n '591,603p' internal/cli/update_yaml_test.go
```

## §D Constraints (Hard)

1. No new dependency. `gopkg.in/yaml.v3 v3.0.1` only.
2. No file added or modified under `internal/template/templates/` (template neutrality, CLAUDE.local.md §25). Fixtures live in `_test.go` bodies or `t.TempDir()`.
3. `t.TempDir()` for every temp path; no project-root writes from tests.
4. Decision semantics held constant (REQ-UYP-009, REQ-UYP-010) — this is a representation change, not a policy change, except for the single deliberate policy reversal in REQ-UYP-006.
5. Package coverage must not fall below the M0-captured baseline.

## §E Self-Verification (Design Decisions)

### Decision D1 — merge on `*yaml.Node`, not on a formatting-aware map wrapper

**Chosen**: decode each document with `yaml.Unmarshal(data, &node)` into a `yaml.Node`, walk the three node trees in parallel, and encode the result with a `yaml.Encoder` configured `SetIndent(2)`.

**Rejected — a custom `OrderedMap` type carrying comment metadata**: reimplements what `yaml.Node` already provides, and would still lose any node property not modelled by the custom type (style flags, tags, line/column). More code, strictly less fidelity.

**Rejected — textual patching of the template file**: preserves formatting perfectly but cannot express the 3-way decision, which requires structural comparison.

**Consequence**: the merged output's structure and comments come from whichever document supplied each node. Since the new template supplies the structure, template comments (the ones users read) are the ones preserved — which is the desired behaviour, because template comments explain the keys while a user's local edit is usually a value change.

### Decision D2 — signature change for `DeepMerge3Way` / `DeepMergeMaps` (highest-reversibility decision in this SPEC)

`DeepMerge3Way(newMap, oldMap, baseMap map[string]any) map[string]any` cannot preserve formatting: the map type has nowhere to store a comment. The signature must become node-typed:

```go
func DeepMerge3Way(newNode, oldNode, baseNode *yaml.Node) (*yaml.Node, error)
func DeepMergeMaps(newNode, oldNode *yaml.Node) (*yaml.Node, error)
```

**Cascade**: eight test call sites across three files — `update_yaml_test.go:346,946` (the plan-audit cited `:925`, but the actual `DeepMerge3Way` call sits at `:946` in the current tree; the enumeration count of 2 sites in this file is correct, only the line drifts), `backup_test.go:427,445,536,551`, `backup_error_test.go:99,117`. That is 2 + 4 + 2 = **8** sites (the original plan prose said "seven" — off by one; the enumeration was already correct).

**`ValuesEqual` fate (D10 resolution)**: the existing exported `ValuesEqual(a, b any) bool` (`merge.go:102`-ish) is the equality primitive of the map implementation. Under D2 it has **no production caller** after the rewrite (both `DeepMerge3Way` and `DeepMergeMaps` become node-typed). **Decision: retain `ValuesEqual` and its test `TestValuesEqual`, unmodified.** Rationale: (a) it is an exported symbol whose removal is a separate breaking-API decision this SPEC does not need to make; (b) it remains a valid equality on the `any` (map / scalar) domain and could be reused by a future non-node caller; (c) keeping it avoids a test-deletion change that would touch `coverage_improvement_test.go` and inflate M5 for no contract benefit. AC-UYP-009's command list continues to include `TestValuesEqual` so its survival is asserted.

**Rejected — keep the map signatures and add parallel node functions**: leaves two merge implementations in the tree, one of which is silently wrong. A future caller picks the wrong one. Dead-code risk with no offsetting benefit.

**Rejected — keep the exported names map-typed and make the node functions unexported**: `DeepMerge3Way` is exercised directly by tests specifically to reach branches `MergeYAML3Way` cannot; hiding it loses that coverage.

**Mitigation**: the two wrapper functions `MergeYAML3Way` / `MergeYAMLDeep` keep their `([]byte, ...) ([]byte, error)` signatures unchanged, so `restore.go` — the only production consumer — needs no edit at all. The cascade is entirely test-side.

### Decision D3 — old-only keys preserved AND reported (REQ-UYP-006, REQ-UYP-007)

Three candidate policies for a key present in the user file and absent from the new template:

| Policy | Behaviour | Verdict |
|---|---|---|
| Drop silently | current | Rejected — data loss with no signal; this is the #1243 complaint |
| Preserve silently | matches `DeepMergeMaps` | Rejected — a genuinely retired template key accumulates as invisible cruft |
| **Preserve + advisory stderr notice** | **chosen** | Data-safe by default, and the user learns which keys are no longer template-managed |

The asymmetry argument settles it: the 3-way path must not be more destructive than the fallback it degrades to (REQ-UYP-008). Reporting rides the existing stderr advisory channel already used by `restore.go:141` and the fallback recorder — no new output surface.

### Decision D4 — sequence replace, not append (REQ-UYP-014)

Eight shipped templates carry block sequences. Merge policy must be explicit or the behaviour is accidental.

**Chosen — replace wholesale**: a sequence value is treated as a single scalar-like unit for decision purposes. If the user's sequence equals base, take the template's; otherwise keep the user's, whole.

**Rejected — element-wise append**: unbounded growth across repeated `moai update` runs, and no identity notion for elements to deduplicate against.

**Rejected — element-wise merge by index**: meaningless for lists whose order carries no key semantics.

This matches the current behaviour, where a sequence lands in the non-map branch of `DeepMerge3Way` (`merge.go:80-92`) and is compared by `ValuesEqual`'s `fmt.Sprintf("%v", ...)` stringification — so replace-wholesale is a preservation of existing semantics, not a new policy.

### Decision D5 — `SaveTemplateDefaults` base derivation is DEFERRED (explicit, per the SPEC's fourth requirement)

**Decision: DEFER to a follow-up SPEC. Not fixed here.**

The defect is real and confirmed: `SaveTemplateDefaults` (`backup.go:149-192`) reads from `template.EmbeddedTemplates()` — the NEW embedded template — and writes it as the 3-way merge BASE. The 3-way premise is "base = what the user started from". Consequence: when a template default changes between versions and the user never touched that key, `old != base` evaluates true, the key is misread as a user edit, and the stale value sticks.

**Why defer**:

1. **Different defect class**. This SPEC fixes a *representation* defect (a lossy encoding). D5 is a *provenance* defect (the wrong document is being used as base). They share a file family but not a mechanism, and fixing them together conflates two independent risk surfaces in one diff.
2. **It requires a new persisted artifact.** A correct base is the template as it existed at the user's *previous* install. Nothing on disk records that today — `SaveTemplateDefaults` is called at backup time (`backup.go:115`), by which point the old template is already gone. A correct fix needs an install-time template snapshot written on every `moai init` / `moai update`, which is a new on-disk data model.
3. **It carries a backward-compatibility problem this SPEC does not.** Every existing user project has no such snapshot. The follow-up must define the behaviour for a first-run-after-upgrade with no recorded base — a design question with several defensible answers, warranting its own requirements and its own audit.
4. **The deferral is safe, but this SPEC enlarges D5's blast radius for `quality.yaml` (plan-audit D9).** D5's failure mode is a stale *value*; this SPEC's failure mode is a destroyed *file*. Fixing the destructive defect first strictly reduces harm, and does not make D5 harder — a node-tree base loads through the same code path a map base does. **However**, for `quality.yaml` the wrong-base problem currently *cannot manifest* (D4 finding): the file's unquoted `{{.EnforceQuality}}` / `{{.TestCoverageTarget}}` placeholders make the 3-way base unparseable, `MergeYAML3Way` errors out, and `restore.go:139` silently routes the file through the 2-way path on every `moai update` today. Post-fix the node decoder parses the base successfully, so `quality.yaml` transitions from *always-2-way* to *real-3-way* — and on the very first run every `{{.Version}}`-style placeholder in the base differs from the user's rendered value, so `old != base` fires and every key is misread as a user edit. The D5 follow-up SPEC therefore gains a new victim file that did not exist as a victim before. This is acceptable under the harm ordering above (a stale value is still less damaging than a destroyed file, and the 2-way fallback remains available), but the plan must name the blast-radius enlargement rather than treat the deferral as neutral.

**What this SPEC does owe D5**: the node-tree rewrite must not entrench the wrong-base assumption. `MergeYAML3Way` continues to take `baseData []byte` from its caller, so the follow-up changes only *which bytes* are passed, not the merge's shape. M2 must not add any code that assumes base and new originate from the same document.

**Follow-up marker**: file as `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001` at sync time. Not created by this SPEC.

### Decision D6 — golden test covers all 30 section templates, not a sample

The golden test (REQ-UYP-017) enumerates `internal/template/templates/.moai/config/sections/*.yaml` and `*.yaml.tmpl` at runtime with `filepath.Glob`, rather than hard-coding a fixture list. A new template added later is automatically covered; a hard-coded list silently stops covering it.

**`.tmpl` parseability (plan-audit D4 correction)**: the original plan claimed `.tmpl` files "parse fine as scalars" and cited `SaveTemplateDefaults` (`backup.go:180-186`) as evidence. Both halves of that claim are wrong.

1. `SaveTemplateDefaults` copies `.tmpl` files **raw** (stripping the `.tmpl` suffix, writing the body verbatim) — it never invokes a YAML parser, so it cannot testify to parseability. The appeal is not evidence.
2. The claim is itself **false for `quality.yaml.tmpl`**, the sole `.tmpl` that leaves Go-template placeholders unquoted. Its lines `enforce_quality: {{.EnforceQuality}}` and `test_coverage_target: {{.TestCoverageTarget}}` cause `{{…}}` to be read as a nested flow mapping, so the **map decoder** (`map[string]any`) fails: `yaml: invalid map key: map[string]interface {}{".EnforceQuality":interface {}(nil)}`. The other `.tmpl` files quote their placeholders (`"{{.X}}"`) and the map decoder parses them as plain strings.

What this means for the run-phase:

- The golden test **must not** assume every `.tmpl` parses through the map decoder — it must use the node decoder (`yaml.Unmarshal(data, &node)`) exclusively, which **does** parse `quality.yaml.tmpl` (the placeholder lands as scalar-nested-flow-mapping text rather than failing). The 22/30 byte-identity figure measured by plan-audit §2.3 was obtained with the node decoder.
- The permanent **2-way fallback** for unparseable templates is part of this SPEC's contract, not a regression: `restore.go:139` already catches `MergeYAML3Way`'s error and falls back to `MergeYAMLDeep`, and that behaviour is preserved (REQ-UYP-011 keeps the wrapped-error + caller-fallback contract). A future template author who introduces another unquoted-placeholder `.tmpl` will see the 3-way path decline to 2-way for that file — the contract is that 2-way fallback is always available.
- An **additional AC** (acceptance.md AC-UYP-023) asserts that `MergeYAML3Way` on `quality.yaml.tmpl` returns a nil error post-fix, so the file's transition from always-2-way to real-3-way is observable rather than silent.

The node decoder's success on `quality.yaml.tmpl` is also why the byte-identity figure for that file is `6605 → 6607` (a 2-byte gain) rather than a parse failure — see acceptance.md §D.1 AC-UYP-005's per-template table.

## §F Milestones (Ordered by Decision-Reversibility)

Highest-change-likelihood decisions first; mechanical revisions last.

### M1 — Node-tree merge core (the data-model decision)

**Files**: new `internal/cli/update/backup/node_merge.go`.

Implement the node-tree primitives before touching any existing function:

- `mergeMappingNode3Way(newNode, oldNode, baseNode *yaml.Node) (*yaml.Node, error)` — walks `Content` pairwise (yaml.v3 mapping nodes store `[k1,v1,k2,v2,...]`), preserving new-node key order, appending old-only keys at the end of the mapping.
- `nodeValuesEqual(a, b *yaml.Node) bool` — structural equality on `Kind`/`Tag`/`Value` and recursively on `Content`. Replaces the `fmt.Sprintf("%v", ...)` stringification of `ValuesEqual` for node inputs. Note: this must distinguish explicit null from empty mapping (REQ-UYP-016), which the stringified comparison does not.
- `encodeNode(n *yaml.Node) ([]byte, error)` — `yaml.NewEncoder` + `SetIndent(2)` + `Encode` + `Close`, errors wrapped.

Unit tests for each primitive land in this milestone, in `node_merge_test.go`.

**Exit**: `go test ./internal/cli/update/backup/ -run 'TestMergeMappingNode|TestNodeValuesEqual|TestEncodeNode' -count=1` green.

### M2 — Rewire the four merge entry points (the behaviour decision)

**Files**: `internal/cli/update/backup/merge.go`.

- `MergeYAML3Way` / `MergeYAMLDeep`: unmarshal into `yaml.Node` instead of `map[string]any`; delegate to the M1 primitives; encode via `encodeNode`. Wrapper signatures unchanged (D2 mitigation).
- `DeepMerge3Way` / `DeepMergeMaps`: signatures change to node-typed per D2.
- Old-only-key retention + stderr advisory per D3.
- Document-node unwrapping: `yaml.Unmarshal` into a `yaml.Node` yields a `DocumentNode` whose single `Content[0]` is the mapping. Multi-document input surfaces as extra content — detect and error per REQ-UYP-015.
- Empty-input handling: an empty `[]byte` yields a zero-Kind node. Preserve the current contract that empty input produces empty output (`update_yaml_test.go:604-616` asserts `{}\n` or `null\n`) — reconcile the exact expected bytes with node encoding and adjust that assertion in M5 if the byte form legitimately changes.

**Exit**: `go build ./...` clean; `go vet ./internal/cli/...` clean.

### M3 — Preservation test contract (the new contract, REQ-UYP-017)

**Files**: new `internal/cli/update/backup/preserve_golden_test.go`.

Golden round-trip over all section templates per D6. For each file, with `new == old == base` (the no-user-edit case), assert on the merged output — the **property set is the gate**, in this order:

- comment-line count equals the source's (primary)
- block-mapping indent width is 2 (primary)
- key order at every mapping level equals the source's (primary)
- quoted scalars remain quoted (primary)
- **byte-equality is a diagnostic, not the gate** (plan-audit D1): only 8 of the 30 section templates produce byte-identical no-edit output through the node encoder; the other 22 (including `llm.yaml`, `workflow.yaml`, `lsp.yaml.tmpl`, `harness.yaml`, `quality.yaml.tmpl`, `delegation.yaml`) reindent and reflow to the canonical 2-space shape. Asserting byte-equality as the sole gate would deadlock M3 against the plan's own §G anti-pattern ("Do not delete a failing assertion to reach green"). The golden test records `bytes_equal` as a per-template boolean for diagnostic localization only — when it fails, the four primary properties above name which axis broke.

A second test covers the user-edit case: change one value, assert the value changes and every comment survives.

**Falsifiability check** — before declaring this milestone done, revert `merge.go` to the map round-trip and confirm the golden test FAILS. A preservation test that passes against the broken implementation proves nothing.

**Exit**: golden test green on the new implementation, RED on the reverted one (both observations recorded verbatim in `progress.md` §E.2).

### M4 — Old-only-key policy reversal + its test (REQ-UYP-006/007/018)

**Files**: `internal/cli/update_yaml_test.go:591-603`.

Rename the subtest to describe the new behaviour and invert the assertion:

```go
name: "user added key not in new template is preserved",
...
if !contains(s, "user_added") {
    t.Errorf("expected user_added preserved (user-added key must survive), got %s", s)
}
```

Add a sibling assertion that the stderr advisory names the retained key path.

**Exit**: `go test ./internal/cli/ -run TestMergeYAML3Way -count=1 -v` green with the renamed subtest visible in output.

**Pre-flight note (plan-audit iter-1 fallout)**: at audit time (2026-07-31) the destructive `t.Errorf("expected user_added to be dropped…")` line was still present at `update_yaml_test.go:600`. An interim commit on `main` has since **rewritten the assertion to preserve `user_added`** and renamed the subtest name to `"user added key not in base preserved"` (the comment now references issue #1267). When run-phase begins against the current tree, M4's rename-and-invert step may therefore be a **no-op**: the run-phase agent MUST verify `grep -rn 'expected user_added to be dropped' --include='*_test.go' .` returns zero matches before treating M4 as completed, and add the sibling stderr-advisory assertion (which is NOT yet present in the interim commit) so REQ-UYP-007 has its own test. AC-UYP-018's grep will then pass against the green implementation, not against an empty match.

### M5 — Revise the conflicting assertion tables (mechanical)

**Files**: `internal/cli/update_yaml_test.go` (`:130-350` `TestDeepMergeMaps`, `:471-648` `TestMergeYAML3Way`, `:650-944` `TestDeepMerge3Way`), `internal/cli/update/backup/backup_test.go` (`:338-560`), `internal/cli/update/backup/backup_error_test.go` (`:61-140`, `:91-120`), `internal/cli/coverage_improvement_test.go` (`:5716`, `:5733`, `:5750`).

Convert map-literal inputs to node construction (a small `mustNode(t, yamlString) *yaml.Node` test helper), and loosen substring assertions that were coupled to map-marshal whitespace while keeping the semantic assertions intact. Do **not** weaken an assertion to make it pass — where output legitimately changed, update the expectation to the new correct value and note why in the test comment.

**Exit**: `go test ./internal/cli/... -count=1` fully green.

### M6 — Verification sweep

```bash
go test ./... -count=1
go test -cover ./internal/cli/update/backup/          # vs M0/pre-flight baseline
golangci-lint run ./internal/cli/...
GOOS=windows GOARCH=amd64 go build ./...
grep -rn "SPEC-UPDATE-YAML-PRESERVE" internal/template/templates/ ; # expect zero matches
```

Plus an end-to-end check: `moai update` against a scratch project in `t.TempDir()` with a comment-bearing, user-edited `cache.yaml`, asserting comments and the edit both survive.

### M7 — Commit

Conventional commit, one logical unit. `fix(update): preserve YAML comments, key order, and user-added keys across config merge (#1243)`.

## §G Anti-Patterns (Avoid)

- **Do not** delete a failing assertion to reach green. Every changed expectation needs a stated reason.
- **Do not** add YAML fixture files under `internal/template/templates/` — template neutrality (constraint 2).
- **Do not** silently change the 3-way decision semantics while changing the representation. If a decision outcome changes, that is a bug in M2, not an improvement.
- **Do not** expand alias nodes for convenience (REQ-UYP-012) — an expanded alias is an irreversible loss of the source's intent.
- **Do not** claim comment preservation without running the M3 falsifiability check. A green test on an unreverted tree does not distinguish "preserves" from "vacuous".
- **Do not** widen scope into `SaveTemplateDefaults` (D5). If the run phase finds the base-derivation defect blocking a test, report it as a blocker rather than fixing it inline.

## §H Cross-References

- `spec.md` §C — the 19 requirements this plan implements.
- `acceptance.md` §D — the AC matrix each milestone is verified against.
- `internal/cli/update/backup/merge.go` — primary change target.
- `internal/cli/update/backup/restore.go:121,139,207` — the three production call sites (unchanged by D2's mitigation).
- CLAUDE.local.md §6 (test isolation), §25 (template neutrality).
- `.claude/rules/moai/languages/go.md` — Go conventions.
