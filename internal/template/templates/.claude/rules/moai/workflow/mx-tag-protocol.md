---
paths: "**/*.go,**/*.py,**/*.ts,**/*.js,**/*.java,**/*.rs,**/*.c,**/*.cpp,**/*.rb,**/*.php,**/*.kt,**/*.swift,**/*.dart,**/*.ex,**/*.scala,**/*.hs,**/*.zig"
---

# @MX TAG Protocol

Purpose: Define rules for @MX code-level annotations that enable AI agents to communicate context, invariants, and danger zones between development sessions.

## Scope

This rule applies to all agents working with source code in the supported programming languages. For full @MX protocol details, see @.claude/skills/moai/references/mx-tag.md.

## @MX Tag Syntax

```
// @MX:TAG_TYPE: [description]
// @MX:SUB_KEY: [sub-value]
```

**Tag Types:**
- `@MX:NOTE` -- Context and intent delivery
- `@MX:WARN` -- Danger zone (requires @MX:REASON)
- `@MX:ANCHOR` -- Invariant contract (requires @MX:REASON)
- `@MX:TODO` -- Incomplete work

**Sub-lines:** @MX:SPEC, @MX:LEGACY, @MX:REASON, @MX:TEST, @MX:PRIORITY

## When to Add Tags

**@MX:NOTE** -- Add when:
- Magic constant encountered
- Exported function lacks godoc and exceeds 100 lines
- Business rule is unexplained

**@MX:WARN** -- Add when:
- Goroutine/channel without context.Context
- Cyclomatic complexity >= 15
- Global state mutation detected
- If-branches >= 8

**@MX:ANCHOR** -- Add when:
- Function has fan_in >= 3 callers
- Public API boundary identified
- External system integration point detected

**@MX:TODO** -- Add when:
- Public function has no test file
- SPEC requirement is not implemented
- Error returned without handling

## When to Update Tags

- **ANCHOR**: Update when fan_in count changes or SPEC is updated
- **NOTE**: Re-review when function signature changes
- **WARN**: Remove when dangerous structure is improved
- **TODO**: Remove when completed (GREEN/IMPROVE phase)

## When to Remove Tags

- **TODO**: Remove when resolved (test passes or implementation complete)
- **WARN**: Remove when danger is eliminated
- **NOTE**: Remove when code is deleted
- **ANCHOR**: NEVER auto-delete; demote to NOTE via report

## Tag Lifecycle Rules

**TODO:**
- Created in RED/ANALYZE phase
- Resolved in GREEN/IMPROVE phase (removed)
- Escalates to WARN after > 3 iterations unresolved

**ANCHOR:**
- Created when fan_in >= 3
- Updated when caller count or SPEC changes
- Demoted to NOTE when fan_in drops below 3 (requires report)
- NEVER auto-deleted

**WARN:**
- Created when danger detected
- Persistent when structural (e.g., goroutine lifecycle)
- Removable when resolved

**NOTE:**
- Created when context needed
- Updated after signature changes
- Obsolete when code deleted

## File Exclusion Rules

Files matching patterns in `.moai/config/sections/mx.yaml` exclude list are not tagged:

Default exclude patterns:
- `**/*_generated.go`
- `**/vendor/**`
- `**/mock_*.go`

## Hard Limits

Per-file limits from `.moai/config/sections/mx.yaml` (defaults):
- `anchor_per_file`: 3
- `warn_per_file`: 5

When limits exceeded:
- ANCHOR: Demote excess by lowest fan_in
- WARN: Keep P1-P5 highest priority only

## Team Environment

In Agent Teams mode, @MX tag operations follow file ownership rules:
- Each teammate only modifies tags within owned file patterns
- Cross-file tag validation respects ownership boundaries
- Report summarizes tag changes across all teammates

## Mandatory Fields

- **@MX:REASON**: MANDATORY for WARN and ANCHOR tags
- **@MX:SPEC**: OPTIONAL -- only include when SPEC exists
- **[AUTO] prefix**: MANDATORY for agent-generated tags

## Comment Syntax by Language

| Language | Prefix | Example |
|----------|--------|---------|
| Go, Java, TS, Rust, C/C++, Swift, Kotlin, Dart, Zig, Scala | `//` | `// @MX:NOTE:` |
| Python, Ruby, Elixir | `#` | `# @MX:WARN:` |
| Haskell | `--` | `-- @MX:ANCHOR:` |

## Configuration

Project-level settings in `.moai/config/sections/mx.yaml`:
- thresholds: fan_in_anchor, complexity_warn, branch_warn
- limits: anchor_per_file, warn_per_file
- exclude: file patterns to skip
- auto_tag: enable/disable autonomous tagging
- require_reason_for: tag types requiring @MX:REASON

## Language Settings

**IMPORTANT**: @MX tag descriptions MUST respect the `code_comments` setting from `.moai/config/sections/language.yaml`.

The `code_comments` setting controls the language used for:
- @MX tag descriptions (NOTE, WARN, ANCHOR, TODO)
- @MX:REASON sub-lines
- Code comments and godoc

Available languages:
- `en` - English (default)
- `ko` - Korean
- `ja` - Japanese
- `zh` - Chinese

**How to read the setting:**
Before adding @MX tags, agents MUST read `.moai/config/sections/language.yaml` and use the `code_comments` value to determine the tag language.

**Example:**
```yaml
# .moai/config/sections/language.yaml
language:
  code_comments: ko  # Tags will be in Korean
```

If `code_comments` is not set, default to English (`en`).

## Agent Reporting

After any phase with tag changes, generate report:

```markdown
## @MX Tag Report -- [Phase] -- [Timestamp]

### Tags Added (N)
### Tags Removed (N)
### Tags Updated (N)
### Attention Required
```

---

Version: 1.0.0
Source: SPEC-MX-001
