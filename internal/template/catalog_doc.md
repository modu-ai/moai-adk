# catalog.yaml — Schema Documentation

`internal/template/catalog.yaml` is the MoAI-ADK 3-tier catalog manifest. It lists
every skill and agent that ships with or can be installed by `moai init`, together with
content hashes for integrity verification.

**SPEC**: SPEC-V3R4-CATALOG-001 T-024 (M5.1)
**Related proposals**: `.moai/brain/IDEA-003/proposal.md` (catalog slimming proposal)

---

## Top-Level Fields

| Field | Type | Description |
|-------|------|-------------|
| `version` | string (semver) | Manifest schema version (e.g. `"1.0.0"`). Bumped on breaking schema changes. |
| `generated_at` | string (ISO 8601) | Timestamp when hashes were last computed. Updated by `gen-catalog-hashes.go --all`. |
| `catalog` | object | Container for the three tier sections (see below). |

---

## Catalog Sections

```
catalog:
  core:             # Always-installed skills and agents
  optional_packs:   # Named packs installed on demand
  harness_generated: # Components created by builder-harness workflow
```

### `catalog.core`

Contains skills and agents that are installed by every `moai init` run. Entries in
this section carry `tier: core`.

### `catalog.optional_packs.<packName>`

A named pack containing related skills and agents for a specific domain. Each pack has:

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | One-line human-readable pack description. |
| `depends_on` | list of strings | Pack names this pack requires. Must form an acyclic graph (enforced by `TestPackDependencyDAG`). |
| `skills` | list of entries | Skills provided by this pack. |
| `agents` | list of entries | Agents provided by this pack. |
| `marketplace_id` | string (optional) | Reserved for future marketplace integration. Must be a string when present. |
| `marketplace_url` | string (optional) | Reserved for future marketplace integration. Must be a string when present. |
| `publisher` | string (optional) | Reserved for future marketplace integration. Must be a string when present. |

Tier strings for pack entries follow the pattern: `optional-pack:<packName>`.
Use `FormatOptionalPackTier(packName)` in Go code.

**Defined packs (9)**: `backend`, `frontend`, `mobile`, `chrome-extension`, `auth`,
`deployment`, `design`, `devops`, `testing`.

**Dependency graph** (acyclic):

```
backend   ← auth
backend   ← deployment
backend   ← devops
frontend  ← design
```

### `catalog.harness_generated`

Components that the builder-harness workflow creates at project init time. Skills and
agents in this section carry `tier: harness-generated`.

---

## Entry Schema

Each element in a `skills` or `agents` list has the following fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Unique identifier. Matches the directory name (skills) or `<name>.md` basename (agents). |
| `tier` | string | yes | See Tier Values below. |
| `path` | string | yes | Path relative to the `templates/` root (e.g. `templates/.claude/skills/moai/`). Skill directories end with `/`; agent paths end with `.md`. |
| `hash` | string | yes | 64-character lowercase hex sha256 of the normalized source file. See Hash Normalization below. |
| `version` | string | yes | Semver string (e.g. `"1.0.0"`). |

### Tier Values

| Value | Meaning |
|-------|---------|
| `core` | Installed by every `moai init`. |
| `optional-pack:<packName>` | Part of an optional named pack. `<packName>` matches a key in `catalog.optional_packs`. |
| `harness-generated` | Created by the builder-harness workflow at runtime. |

Tier validity is enforced by regex `^(core|optional-pack:[a-z][a-z0-9-]{1,30}|harness-generated)$`
in `TestCatalogTierValid`.

---

## Hash Normalization

Hashes are computed over a normalized form of the source file to ensure
cross-platform reproducibility (Windows CRLF vs Unix LF).

**Algorithm** (implemented in `NormalizeForHash` in `catalog_hash_norm.go`):

1. Replace all `\r\n` (CRLF) with `\n` (LF).
2. Replace any remaining lone `\r` (CR) with `\n`.
3. Strip trailing whitespace (spaces and tabs) from every line.
4. Join lines with `\n`, then ensure exactly **one** trailing newline.
5. Compute `sha256` of the resulting bytes; hex-encode to lowercase 64-char string.

**For skills**: hash only the root `SKILL.md` or `skill.md` file, not sub-files.  
**For agents**: hash the `.md` file at the given path.

Hash correctness is verified by `TestManifestHashFormat`, which re-computes each hash
at test time and compares against the stored value. This test is the Windows CI
stability check (REQ-022, REQ-023).

To regenerate hashes after editing template files, run:

```bash
go run internal/template/scripts/gen-catalog-hashes.go --all
```

---

## Go API

```go
// Load the catalog from the raw embedded FS (before "templates/" prefix strip).
cat, err := template.LoadCatalog(embeddedRaw)

// Look up a specific skill or agent by name.
entry, ok := cat.LookupSkill("moai-domain-backend")
entry, ok := cat.LookupAgent("evaluator-active")

// Flat list of all 65 entries.
all := cat.AllEntries()

// Build the optional-pack tier string for a known pack name.
tier := template.FormatOptionalPackTier("backend") // → "optional-pack:backend"
```

---

## Follow-up SPECs

This document covers SPEC-V3R4-CATALOG-001 (manifest + loader + audit suite).
Later SPECs in the CATALOG series extend this foundation:

| SPEC | Topic |
|------|-------|
| SPEC-V3R4-CATALOG-002 | `moai catalog list` CLI command |
| SPEC-V3R4-CATALOG-003 | `moai catalog verify` hash integrity check |
| SPEC-V3R4-CATALOG-004 | `moai catalog install <pack>` optional-pack installation |
| SPEC-V3R4-CATALOG-005 | `moai catalog update` selective template refresh |
| SPEC-V3R4-CATALOG-006 | Marketplace registry integration |
| SPEC-V3R4-CATALOG-007 | Harness-generated entries lifecycle management |
