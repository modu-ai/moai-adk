# Design — SPEC-CONFIG-AUDIT-REPAIR-001

All three decisions were RESOLVED at Implementation Kickoff (2026-07-25). Original option analyses are preserved below for audit trail; each decision header records the resolution.

## DECISION-1 (H3+M5) — ast-grep pre-tool gate — **RESOLVED: RESTORE (modified Option A), default OFF**

**Resolution (user directive)**: NOT removal. "재설계를 통해 ast-grep을 제대로 사용 가능하도록 게이트 복구" — restore the gate as a properly usable, opt-in feature. The original Option-B recommendation is superseded; its core objection (enabling over a broken ruleset) is answered by making ruleset loadability (REQ-CAR-020) a hard precondition of the enable path within this SPEC.

Resolved design points:
- **Loader placement**: a dedicated `gate.yaml` section file loaded via the existing sections loader — chosen because the `gate` registry key ALREADY exists in `audit_registry.go` (defaults-only) and `GateConfig` types exist (`types.go:569-570`); adding the loader completes an already-half-built pair with minimal new surface, versus overloading an unrelated existing section (rejected: obscures the SSOT and complicates the audit registry story).
- **Default posture**: OFF (`Enabled: false` explicit default in defaults.go) — deployed users see zero behavior change without explicit opt-in (REQ-CAR-019).
- **V1 `RunAstGrepGate` disposition**: DELETE/fold V1; V2 (`RunAstGrepGateV2`) is the sole gate entry. Rationale: two parallel entries with one newly-reachable invites divergence; V1 has no independent caller once V2 is config-reachable (REQ-CAR-021; verify caller-lessness at run pre-flight).
- **Ruleset**: fix local `sgconfig.yml:24` phantom `utils`; config-mode loads root `go-hardcoding.yml` (or the template's curated go/security set); graceful degrade when `sg` binary absent.
- **Backlog boundary**: SPEC-ASTGREP-DOGFOOD-CLEANUP-001 retains 16-language productization, empty stubs, message unification, SPEC-ID stripping.
- **Resequencing** (per the original interaction note): the new `gate` section registration lands in M-2 BEFORE M-4's registry/completeness reconciliation.

### Original analysis (superseded — audit trail)

Current state (audit H3): `AstGrepGate.Enabled` has no path to true — `defaults.go:286-296` never sets it, `loader.go:47-98` has no `gate` section, guard `hook/quality/gate.go:281` short-circuits, `pre_tool.go:566-575` and `RunAstGrepGateV2` unreachable. The explicit `moai ast-grep` CLI path is separate and alive.

### Option A — add a `gate` section loader (enable path)
- Work: new `gate.yaml` section schema + loader + defaults registration + template distribution + docs; tests proving Enabled=true reaches the guard.
- Pros: preserves the original design intent (pre-tool quality gate); the guard code was written for this.
- Cons: activates a feature whose ruleset story is broken (M5: local sgconfig phantom `utils` dir; root `go-hardcoding.yml` not loaded in config-mode; template ships only one rule file; 10/17 language stubs empty). Enabling the gate before ASTGREP-DOGFOOD-CLEANUP-001 lands ships a gate with no healthy rules to run — the gate would fire on `sg` availability with a defective config. Larger surface: new section, schema, template mirror, 16-language neutrality.

### Option B — remove the dead path (RECOMMENDED)
- Work: delete `AstGrepGate` struct wiring, the `gate.go:281` guard branch, `pre_tool.go:566-575`, and `RunAstGrepGateV2` if caller-less; adjust tests; keep `moai ast-grep` CLI untouched.
- Pros: eliminates wired-but-dead code (simplicity ladder step 1: does this need to exist now? — no consumer, no config path, no ruleset). Fully reversible: git history preserves the implementation; a future SPEC (post ASTGREP-DOGFOOD-CLEANUP-001) can reintroduce the gate WITH a loader and a healthy ruleset in one coherent effort.
- Cons: if the dogfood cleanup lands soon and wants the gate, the code must be restored from history.
- Recommendation rationale: dead-path removal is the reversible, low-risk move; enabling now couples this SPEC to an unresolved backlog SPEC's ruleset health. Coordination boundary (REQ-CAR-012): sgconfig repair stays in ASTGREP-DOGFOOD-CLEANUP-001 either way.

## DECISION-2 (H4) — tool-policy.yaml distribution status — **RESOLVED: Option A (dev-only + graceful CLI error)**

Current state: 47KB yaml, code-consumed (`toolpolicy/loader.go:78`, `moai tool-policy` CLI), hard error at `loader.go:19` when absent, not in template, user-facing docs 0. Dev-only intent already recorded in `audit_loader_completeness_test.go:39` allowlist comment.

### Option A — declare dev-only (RECOMMENDED)
- Work: CLAUDE.local.md §2 local-only list entry; graceful CLI message replacing the hard error ("tool-policy is a maintainer-only surface; not distributed to user projects"); SSOT doc note (REQ-CAR-016).
- Pros: matches the already-recorded intent (test allowlist comment); zero neutrality burden (47KB of internal policy would need heavy §25 scrubbing to ship); smallest diff.
- Cons: distributed users get a documented no-op CLI rather than a working feature.

### Option B — neutralize + distribute
- Work: scrub 47KB for §25 neutrality, add to template sections, register in completeness test expectations, document for users.
- Pros: `moai tool-policy` works everywhere.
- Cons: large neutralization effort on a file with no user-facing docs or demand signal; ongoing dual-tree maintenance of 47KB policy content.

## DECISION-3 (M3) — mcp-matrix.yaml dangling template reference — **RESOLVED: Option A (reword reference, keep dev-only)**

Current state: shipped template skill `project/doc-generation.md` references `mcp-matrix.yaml`; template does not distribute it; Go consumers 0 (prompt-consumed only).

### Option A — reword the template-skill reference, keep yaml dev-only (RECOMMENDED)
- Work: edit `project/doc-generation.md` (template + local) to drop/generalize the mcp-matrix reference; add mcp-matrix to the acknowledged local-only inventory.
- Pros: removes the dangling reference at its source; prompt-only consumption means users lose nothing functional; minimal diff.
- Cons: the doc-generation skill loses a (currently broken anyway) pointer.

### Option B — neutralize + distribute mcp-matrix.yaml
- Pros: reference becomes valid. Cons: distributes a maintainer-inventory file with no code consumer; neutrality scrub + registry/test updates for a prompt-only asset.

## Interaction notes

- DECISION-2 Option A + DECISION-3 Option A together produce one consolidated "maintainer-only config surfaces" documentation entry (REQ-CAR-016 acknowledged-orphan list) — preferred pairing.
- DECISION-1 restore applies the Option-A resequencing: the new `gate` section registration in `audit_registry.go` / completeness expectations lands in M-2, before M-4's reconciliation (now reflected in plan.md §F).
