---
title: moai constitution Constitution
weight: 84
draft: false
---

# moai constitution Constitution

`moai constitution` queries and validates the zone registry (the codification of FROZEN/EVOLVABLE zones). It is a command tree that governs which parts of the rules are FROZEN and cannot be casually changed, and which parts are EVOLVABLE.

When a single SPEC flows through multiple manager agents and harness rules, accidentally touching a core invariant breaks the consistency of the entire system. So this command tags each rule with its zone and mechanically verifies whether the FROZEN area has drifted from the source files. Hanging `moai constitution guard` off the CI pipeline catches violations before merge, which is why it serves as the safety net during large-scale rule overhauls.

## Subcommands

| Command | Description |
|--------|------|
| `moai constitution list` | List zone-registry entries |
| `moai constitution guard` | Check for FROZEN-zone violations (for CI integration) |
| `moai constitution amend` | Propose a constitutional amendment that passes the 5-layer safety gate |
| `moai constitution validate` | Validate zone-registry drift and invariants against the source files |

## moai constitution list

```bash
moai constitution list
moai constitution list --zone frozen --format json
```

| Flag | Description |
|--------|------|
| `--zone <frozen\|evolvable>` | Zone filter |
| `--file <path>` | File-path filter (partial match) |
| `--format <table\|json>` | Output format |

## moai constitution guard

```bash
moai constitution guard --violations CONST-V3R2-001,CONST-V3R2-002
```

Takes a list of changed rule IDs and checks for FROZEN-zone violations. For CI integration.

| Flag | Description |
|--------|------|
| `--violations <ids>` | List of changed rule IDs (comma-separated or repeated flag) |

## moai constitution amend

```bash
moai constitution amend --rule CONST-V3R2-001 --before "..." --after "..." --evidence "..."
```

Must pass the FrozenGuard → Canary → ContradictionDetector → RateLimiter → HumanOversight 5-layer safety gate to be applied.

| Flag | Description |
|--------|------|
| `--rule <id>` | Rule ID (CONST-V3R2-NNN) [required] |
| `--before <text>` | Current clause text [required] |
| `--after <text>` | New clause text [required] |
| `--evidence <text>` | Amendment rationale (required for the Frozen zone) |
| `--dry-run` | Simulate only, without modifying files |

## moai constitution validate

```bash
moai constitution validate
```

Confirms that each registry entry's clause exists in the source files, validates the zone_class enum and canary_gate invariants, and reports drift.

| Flag | Description |
|--------|------|
| `--strict` | Strict mode (enforce all checks) |
| `--fail-on-warning` | Treat warnings as errors (includes `--strict`) |
| `--format <text\|json>` | Output format |

Exit codes: 0 = OK, 1 = drift/error, 2 = fatal (source file missing).

## Related documents

- [CLI Overview](/en/getting-started/cli)
