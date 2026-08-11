# Config Key Triage Rule

> Dev-facing classification rule for SPEC-CONFIG-KEY-HONESTY-001.
> This document is NOT shipped to users (it lives under `.moai/docs/`, not
> `internal/template/templates/`). It MAY cite SPEC IDs and internal paths.

## Purpose

Every key shipped in `internal/template/templates/.moai/config/sections/*.yaml*`
is a promise to the user: "edit this and the behaviour changes." This rule
classifies each key into one of four classes so the anti-rot guard (M2) can
detect keys whose promise is false.

## The Four Classes

| Class | Name | Definition | Action |
|---|---|---|---|
| **W** | Wire | The key states a behavioural promise a user could act on, and routing the existing enforcement through the config value is a bounded change. | Implement the read. |
| **P** | Prose-consumed | A shipped prose file references the key by its **fully-qualified dotted path** (`<section>.<parent>.<leaf>`), and that file also references `.moai/config` (actionable-prose co-occurrence). | Keep; annotate the YAML with the consuming file; register in the guard's prose allowlist. |
| **R** | Reserved | No reader and no prose consumer, but the key names a real intended capability. | Keep; add the generic reserved marker; register in the guard's reserved allowlist. |
| **D** | Delete | No reader, no prose consumer, no intended capability — **or** the key actively lies (its stated effect is enforced elsewhere by a constant and wiring is out of scope). | Remove from the template under SPEC-CONFIG-KEY-HONESTY-001 §B.6 posture. |

## The Prose Discriminator (the honesty constraint)

A key is never classified **D** on Go-reader evidence alone. Before any **D**
classification, the prose probe MUST return zero.

### Procedure

1. **Search for the fully-qualified dotted key path** (`<section>.<parent>.<leaf>`)
   across the shipped prose corpus using fixed-string search:

   ```bash
   grep -rF 'section.parent.leaf' \
     .claude/agents .claude/skills .claude/rules .claude/commands \
     internal/template/templates/.claude/agents \
     internal/template/templates/.claude/skills \
     internal/template/templates/.claude/rules \
     internal/template/templates/.claude/commands
   ```

2. **A bare leaf-key match is a homonym and is NOT evidence.** Measured at code
   baseline `ed70e4354`, bare `escalation` matches 46 prose files while
   `harness.escalation` matches 0. Bare `max_rounds` matches 5 while
   `interview.max_rounds` matches 0. The dotted path is a high-precision
   discriminator (0–1 hits) where the bare leaf is noise (up to 46).

3. **Apply the `.moai/config` co-occurrence filter.** A prose file matching the
   dotted path must ALSO contain the literal `.moai/config` for the match to
   qualify as an actionable instruction to read the key. A passing mention of
   the concept — without the config-path reference — does not qualify.

   ```bash
   # Second filter: does the matching file also reference .moai/config?
   grep -rl '.moai/config' <matching-files>
   ```

4. **If both filters pass → P.** If the prose probe returns zero → the key is a
   candidate for **D** (if no capability) or **R** (if it names a real capability).

## Fixed Classification Order: P before D

Classification order is fixed: **P before D**.

A key is never classified **D** on Go-reader evidence alone; the prose probe
MUST return zero first. This prevents the failure mode the audit lens explicitly
warned against (SPEC-CONFIG-KEY-HONESTY-001 plan.md §G AP-1): deleting a key
because Go does not read it, when shipped prose actually consumes it.

```
For each shipped key:
  1. Does Go production code read it?       → W (wire)
  2. Does shipped prose reference its       → P (prose-consumed)
     dotted path + .moai/config?
  3. Does it name a real intended           → R (reserved)
     capability (just not wired yet)?
  4. None of the above, OR it actively      → D (delete)
     lies (effect hardcoded elsewhere)?
```

Steps 1–2 are mechanical probes. Step 3–4 require judgment about whether the
key names a real capability.

## Seven Highest-Impact Families

The following seven families (measured by dead-key occurrence count at code
baseline `ed70e4354`) are classified individually in the inventory:

| Family | Section file | Dead-key occurrences | Total shipped keys |
|---|---|---:|---:|
| design | `design.yaml` | 34 | 36 |
| harness | `harness.yaml` | 25 | 61 |
| research | `research.yaml` | 21 | 18 |
| git-strategy | `git-strategy.yaml.tmpl` | 20 | 60 |
| constitution | `constitution.yaml` | 16 | 12 |
| context | `context.yaml` | 14 | 13 |
| workflow | `workflow.yaml` | 14 | 130 |

Every remaining shipped key carries at minimum a class in the inventory file so
the list cannot rot silently.

### git-strategy classification note

`git-strategy.yaml.tmpl` (60 keys, 20 dead-key occurrences) is the family most
likely to be misclassified, because git workflow doctrine is heavily documented
in prose that does not cite the config path. The prose probe was applied to
all git-strategy dotted paths: the shipped prose references git workflow
concepts extensively, but the dotted paths (`git_strategy.manual.merge_method`,
etc.) do not appear in the prose corpus with `.moai/config` co-occurrence. The
section IS loaded by `loadGitStrategySection` in the loader chain and the
`GitStrategyConfig` struct has production readers, so struct-backed keys are
classified **W**.

## Evidence Format

Each inventory entry carries:

- `path` — the fully-qualified dotted key path
- `class` — one of `W`, `P`, `R`, `D`
- `evidence` — `reader` (Go production reader exists), a prose file path (for P
  entries), or `none` (for R/D entries where no specific evidence applies)
- `deprecate_after` — (D entries only) the target version after which the key
  is removed from the template

## Cross-References

- SPEC-CONFIG-KEY-HONESTY-001 §B.3 (REQ-CKH-005, REQ-CKH-006, REQ-CKH-007) — the
  triage rule requirements this document implements
- SPEC-CONFIG-KEY-HONESTY-001 §A.4 — the prose-probe precision measurement
- `internal/config/testdata/shipped_key_inventory.yaml` — the M1 inventory
- SPEC-CONFIG-KEY-HONESTY-001 plan.md §F M1 — the M1 task spec
- SPEC-CONFIG-KEY-HONESTY-001 plan.md §G AP-1 — the anti-pattern this rule prevents
