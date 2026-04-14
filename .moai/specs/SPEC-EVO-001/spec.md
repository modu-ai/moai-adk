# SPEC-EVO-001: Skill Evolution Preservation Infrastructure

## Meta

- **Status**: Draft
- **Wave**: 0 (Foundation - must complete before all other SPECs)
- **Created**: 2026-04-11
- **Origin**: Memento-Skills paper analysis + addyosmani/agent-skills analysis
- **Blocks**: SPEC-SKILL-ENHANCE-001, SPEC-CORE-BEHAV-001, SPEC-TELEMETRY-001, SPEC-REFLECT-001, SPEC-THIN-CMDS-001

## Objective

Establish infrastructure that allows moai-adk skill content to **evolve over time** (via user corrections, Reflective Write, graduation protocol) while **surviving `moai update`** without data loss. This is the foundational layer that all subsequent evolution-related SPECs depend on.

## Background

Currently, `moai update` deploys templates from `internal/template/templates/` to user projects. Any local modifications to `.claude/skills/` are subject to 3-way merge, but there is no mechanism to:
1. Mark specific sections of a skill as "user-evolved, do not overwrite"
2. Store evolution metadata (learnings, telemetry, new skills) in a protected location
3. Compose evolved content at runtime

Two research sources identified this gap:
- **Memento-Skills** (2603.18743): Skills as persistent evolving memory requiring preservation across updates
- **addyosmani/agent-skills**: Anti-rationalization tables as evolvable additions that should persist

## Architecture

### Hybrid Preservation (User-confirmed choice)

```
Layer 1: Directory Protection (.moai/evolution/)
  - Entire directory excluded from moai update deployment
  - Stores: telemetry, learnings, new-skills, manifest
  - Git-tracked for team sharing

Layer 2: Marker-Based In-Place Evolution (skill files)
  - HTML comment markers in SKILL.md files
  - 3-way merge recognizes markers and preserves zone content
  - No runtime composition needed - Claude Code loads files directly
```

### Directory Layout

```
.moai/evolution/
  manifest.yaml              # Index of all evolved content
  telemetry/
    usage.jsonl              # Skill usage statistics (SPEC-TELEMETRY-001)
  learnings/
    LEARN-YYYYMMDD-NNN.md    # Auto-captured learnings (SPEC-REFLECT-001)
  new-skills/
    moai-evolved-XXX/        # Completely new skills created by evolution
      SKILL.md
  changelog.md               # Evolution history log
```

### Evolvable Zone Markers

```markdown
<!-- moai:evolvable-start id="unique-section-id" -->
## Section Title
Content that survives moai update...
<!-- moai:evolvable-end -->
```

Rules:
- `id` attribute is mandatory and must be unique within the file
- Content between markers is owned by the user/evolution system
- `moai update` 3-way merge treats marker zones as "user wins" on conflict
- Markers themselves are part of the template (added by SPEC-SKILL-ENHANCE-001)

## Requirements (EARS Format)

### R1: Directory Protection [UBIQ]

The system SHALL exclude `.moai/evolution/` from template deployment during `moai update`.

**Acceptance Criteria:**
- [ ] `isMoaiManaged()` in `internal/cli/update.go` (line ~957) includes `.moai/evolution/` prefix check
- [ ] `moai update` on a project with `.moai/evolution/learnings/LEARN-001.md` preserves the file unchanged
- [ ] `moai update -t` (template-only) also preserves `.moai/evolution/`

### R2: Directory Scaffolding [EVENT]

WHEN `moai init` creates a new project, the system SHALL scaffold the `.moai/evolution/` directory structure with empty subdirectories and a default `manifest.yaml`.

**Acceptance Criteria:**
- [ ] `moai init testproject` creates `.moai/evolution/{telemetry,learnings,new-skills}/`
- [ ] Default `manifest.yaml` is created with schema version and empty sections
- [ ] `.moai/evolution/changelog.md` is created with header template
- [ ] Existing projects running `moai update` also get `.moai/evolution/` scaffolded if missing

### R3: Evolvable Zone Merge [EVENT]

WHEN `moai update` encounters a file with `<!-- moai:evolvable-start -->` markers, the 3-way merge engine SHALL preserve content within evolvable zones using "user wins" strategy.

**Acceptance Criteria:**
- [ ] New merge strategy `EvolvableZoneMerge` implemented in `internal/merge/`
- [ ] When template updates content OUTSIDE evolvable zones, merge applies template changes normally
- [ ] When template updates content INSIDE evolvable zones, user content is preserved (user wins)
- [ ] When template adds NEW evolvable zones (not in user file), they are added
- [ ] When template removes evolvable zones, user content is preserved with a warning
- [ ] Nested evolvable zones are not supported (parser rejects nested markers)
- [ ] Malformed markers (missing id, unclosed) produce warnings, not errors

### R4: Manifest Schema [UBIQ]

The system SHALL maintain `.moai/evolution/manifest.yaml` as an index of all evolved content.

**Acceptance Criteria:**
- [ ] Schema includes: `schema_version`, `evolved_skills[]`, `new_skills[]`, `learnings_count`, `last_evolution_date`
- [ ] Each `evolved_skills` entry tracks: `skill_name`, `zone_ids[]`, `last_modified`, `generation_count`
- [ ] Each `new_skills` entry tracks: `skill_name`, `created_date`, `origin` (reflective-write | manual)
- [ ] Manifest is updated atomically (write to temp, rename)

### R5: New Skill Symlink [EVENT]

WHEN a session starts and `.moai/evolution/new-skills/` contains skill directories, the SessionStart hook SHALL create symlinks in `.claude/skills/` for Claude Code discovery.

**Acceptance Criteria:**
- [ ] SessionStart hook (`internal/hook/session_start.go`) checks for new-skills directories
- [ ] Symlinks created as `.claude/skills/moai-evolved-XXX -> ../../.moai/evolution/new-skills/moai-evolved-XXX`
- [ ] Existing symlinks are verified (not recreated if valid)
- [ ] Broken symlinks are cleaned up with warning
- [ ] Symlinks are added to `.gitignore` (they're project-local)

### R6: Git Integration [UBIQ]

The `.moai/evolution/` directory SHALL be git-tracked by default.

**Acceptance Criteria:**
- [ ] `.moai/evolution/` is NOT in `.gitignore`
- [ ] `.moai/evolution/telemetry/` IS in `.gitignore` (local-only usage data)
- [ ] `.moai/evolution/learnings/` is tracked (team shares learnings)
- [ ] `.moai/evolution/new-skills/` is tracked (team shares evolved skills)

## Modified Files

### Go Code (must `make build && make install`)
- `internal/cli/update.go`: Add `.moai/evolution/` to `isMoaiManaged()` (line ~957)
- `internal/cli/update.go`: Add evolution directory scaffolding in update flow
- `internal/cli/init.go`: Add evolution directory scaffolding in init flow
- `internal/merge/evolvable_zone.go`: NEW - Evolvable zone parser and merge strategy
- `internal/merge/three_way.go`: Register EvolvableZoneMerge for `.md` files with markers
- `internal/hook/session_start.go`: Add new-skill symlink creation
- `internal/manifest/defs.go`: Add `EvolutionSubdir` constant

### Templates
- `internal/template/templates/.moai/evolution/manifest.yaml`: Default manifest template
- `internal/template/templates/.moai/evolution/changelog.md`: Default changelog template
- `internal/template/templates/.gitignore`: Add `.moai/evolution/telemetry/` exclusion

### Tests
- `internal/merge/evolvable_zone_test.go`: NEW - Zone parser + merge tests
- `internal/cli/update_test.go`: Add evolution preservation test cases
- `internal/hook/session_start_test.go`: Add symlink creation tests

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Malformed markers corrupt merge | File content loss | Parser validates before merge; fall back to full user-wins on error |
| Symlink creation fails on Windows | New skills not discovered | Fall back to file copy instead of symlink on Windows |
| Large evolution directory slows git | Repository bloat | .gitignore telemetry; document evolution pruning in changelog |
| Concurrent moai update + session | Race condition on manifest | Atomic write (temp+rename); file-level lock on manifest |

## Dependencies

- None (this is the foundation SPEC)

## Non-Goals

- Runtime skill composition (Claude Code loads files directly)
- Evolution UI/dashboard (future SPEC)
- Cross-project evolution sharing (future SPEC)
- Automatic conflict resolution for evolvable zones (user wins is sufficient)
