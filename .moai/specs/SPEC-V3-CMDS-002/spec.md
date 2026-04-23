---
id: SPEC-V3-CMDS-002
title: "Command $ARGUMENTS[N] and Named Args"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P2 Medium
phase: "v3.0.0 — Phase 3 Agent Runtime v2"
module: "internal/command/arguments.go"
dependencies:
  - SPEC-V3-CMDS-001
related_gap:
  - gm#99
  - gm#100
  - gm#101
related_theme: "Command argument substitution parity"
breaking: false
bc_id: null
lifecycle: spec-anchored
tags: "command, arguments, substitution, v3"
---

# SPEC-V3-CMDS-002: Command $ARGUMENTS[N] and Named Args

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-22 | GOOS | Initial v3 draft from Wave 4 bundle (Hooks/Commands) |

---

## 1. Goal (목적)

moai's command/skill substitution supports only `$ARGUMENTS` (the full args string). Claude Code's substitution system is richer:
- `$ARGUMENTS[N]` — 0-indexed positional arg.
- `$N` — shorthand for `$ARGUMENTS[N]` (integer N followed by non-word char).
- `$name` — named argument when frontmatter `arguments: "name1 name2 name3"` sets a named list.
- Progressive typeahead hint via `generateProgressiveArgumentHint(argNames, typedArgs)` (covered by SPEC-V3-CMDS-001).

Parsing uses `tryParseShellCommand` semantics (shell-quote) so quoted strings stay as single tokens. Without these primitives, moai commands cannot surgically use individual args (a common need for routing commands).

## 2. Scope (범위)

In-scope:
- Parse `$ARGUMENTS` (full string), `$ARGUMENTS[N]` (0-indexed), `$N` (shorthand when N is integer and not followed by a word char), `$name` (named arg).
- Frontmatter `arguments: "name1 name2 name3"` declares positional named args mapped to `$name1`, `$name2`, `$name3`.
- Shell-quote-aware parser for user's arg string: quoted tokens stay as one arg (`"hello world"` is one token).
- `appendIfNoPlaceholder` behavior: if the body contains no placeholder and `appendIfNoPlaceholder: true` frontmatter is set, append `\n\nARGUMENTS: {args}` to the prompt.
- Out-of-range `$ARGUMENTS[N]` (N ≥ len(args)) substitutes the empty string with a trace-level warning (matches CC behavior).

Out-of-scope:
- Progressive argument hint typeahead UX (covered by SPEC-V3-CMDS-001 `argument-hint` array form).
- Inline bash execution `!`cmd`` → SPEC-V3-CMDS-003.
- `${CLAUDE_SKILL_DIR}` / `${CLAUDE_SESSION_ID}` body substitution → SPEC-V3-SKL-001 / -002.
- `isSensitive` argument redaction.

## 3. Environment (환경)

Current moai-adk state:
- Moai commands use only `$ARGUMENTS` placeholder (findings-wave1-moai-current.md §8 — command catalog).
- No indexed or named argument substitution.
- No shell-quote-aware parser; args passed as whitespace-split tokens.

Claude Code reference:
- `utils/argumentSubstitution.ts:94-145` — core substitution implementation (findings-wave1-hooks-commands.md §8).
- `utils/argumentSubstitution.ts:124-127` — indexed substitution (findings-wave1-hooks-commands.md §14.5).
- `utils/argumentSubstitution.ts:50-67` — named arg substitution (findings-wave1-hooks-commands.md §14.5).
- `utils/argumentSubstitution.ts:76-83` — progressive hint generator (referenced by SPEC-V3-CMDS-001).

Affected modules:
- `internal/command/arguments.go` — new file: substitution engine.
- `internal/command/loader.go` — integrate substitution into command dispatch.
- `internal/template/templates/.claude/commands/` — no mandatory changes; commands may opt into new placeholders.

## 4. Assumptions (가정)

- Shell-quote parsing uses a library or minimal in-house parser that respects single quotes, double quotes, and escaped spaces. Go stdlib has no drop-in equivalent; a ~50-LOC parser suffices.
- Named arg names match `[a-zA-Z_][a-zA-Z0-9_]*` (C-like identifiers); characters outside this set in `arguments:` field are rejected.
- `$N` shorthand is only treated as indexed substitution when followed by a non-word character (space, punctuation, end of string) — matches CC semantics.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-CMDS-002-001: The substitution engine SHALL parse the user's raw argument string via shell-quote semantics (quoted tokens remain single args).
- REQ-CMDS-002-002: The `$ARGUMENTS` placeholder SHALL substitute the full (re-joined) argument string, preserving original user spacing where possible.
- REQ-CMDS-002-003: The `$ARGUMENTS[N]` placeholder SHALL substitute the N-th argument (0-indexed).
- REQ-CMDS-002-004: The `$N` shorthand SHALL behave identically to `$ARGUMENTS[N]` when N is an integer and the next character is not a word character.
- REQ-CMDS-002-005: When frontmatter contains `arguments: "name1 name2 name3"`, the substitution engine SHALL map `$name1` to the 0th arg, `$name2` to the 1st, etc.
- REQ-CMDS-002-006: Named-arg identifiers SHALL match regex `^[a-zA-Z_][a-zA-Z0-9_]*$`; other values SHALL be rejected at load time.
- REQ-CMDS-002-007: The substitution engine SHALL support simultaneous use of indexed, named, and full forms in the same command body.

### 5.2 Event-driven Requirements

- REQ-CMDS-002-010: WHEN a placeholder references an out-of-range index (`$ARGUMENTS[N]` or `$N` where N ≥ len(args)), the engine SHALL substitute the empty string and emit a trace-level warning.
- REQ-CMDS-002-011: WHEN a named-arg reference cannot be resolved (`$unknown` with `unknown` not in the `arguments:` list), the engine SHALL leave the literal text `$unknown` in place (no substitution) and emit a trace-level warning.
- REQ-CMDS-002-012: WHEN the body contains no substitution placeholder AND the frontmatter `appendIfNoPlaceholder: true` is set, the engine SHALL append `\n\nARGUMENTS: {args}` to the rendered body.
- REQ-CMDS-002-013: WHEN the body is rendered for the model turn, substitution SHALL be performed exactly once (no recursive expansion).

### 5.3 State-driven Requirements

- REQ-CMDS-002-020: WHILE frontmatter `arguments: ""` (empty) is set, the engine SHALL reject `$name`-form references and treat them as unknown per REQ-CMDS-002-011.
- REQ-CMDS-002-021: WHILE the configuration key `command.arguments.strict_mode: true` is set, the engine SHALL escalate trace-level warnings to errors (out-of-range index or unknown name → command aborted).

### 5.4 Optional Features

- REQ-CMDS-002-030: WHERE the configuration key `command.arguments.preserve_spacing: true` is set, the engine SHALL preserve the user's original whitespace between args when reconstructing `$ARGUMENTS`; otherwise args are joined with single spaces.

### 5.5 Complex Requirements

- REQ-CMDS-002-040: IF a body contains both `$ARGUMENTS` and `$ARGUMENTS[0]`, THEN both SHALL resolve independently; the full string and the 0th arg are substituted separately; ELSE only the matched placeholder is substituted.
- REQ-CMDS-002-041: IF a placeholder appears inside a code-fence block (```...```), THEN the engine SHALL still perform substitution (matches CC behavior); ELSE (no code-fence) normal substitution applies.

## 6. Acceptance Criteria (수용 기준 요약)

- AC-CMDS-002-01: Given a command body `Rename $ARGUMENTS[0] to $ARGUMENTS[1]` and user input `/rename foo.go bar.go`, When rendered, Then the body becomes `Rename foo.go to bar.go`. (maps REQ-CMDS-002-003)
- AC-CMDS-002-02: Given a command body `Process $1 and $2` and user input `/cmd a b`, When rendered, Then the body becomes `Process a and b`. (maps REQ-CMDS-002-004)
- AC-CMDS-002-03: Given frontmatter `arguments: "source target"` and body `cp $source $target`, user input `/cp /tmp/a /tmp/b`, When rendered, Then the body becomes `cp /tmp/a /tmp/b`. (maps REQ-CMDS-002-005)
- AC-CMDS-002-04: Given user input `/cmd "hello world" bar`, When parsed via shell-quote, Then args[0] is `hello world` (single token) and args[1] is `bar`. (maps REQ-CMDS-002-001)
- AC-CMDS-002-05: Given body `Do $ARGUMENTS[5]` and user input with 3 args, When rendered, Then the placeholder resolves to the empty string and a trace warning is emitted. (maps REQ-CMDS-002-010)
- AC-CMDS-002-06: Given `command.arguments.strict_mode: true`, body `Do $ARGUMENTS[5]`, 3 args, When rendered, Then the command aborts with `CommandArgumentIndexOutOfRange`. (maps REQ-CMDS-002-021)
- AC-CMDS-002-07: Given frontmatter `arguments: "source target"` and body references `$unknown`, When rendered, Then `$unknown` remains literal and a trace warning is emitted. (maps REQ-CMDS-002-011)
- AC-CMDS-002-08: Given body contains no placeholder, frontmatter `appendIfNoPlaceholder: true`, user input `foo bar`, When rendered, Then the body gains trailing `\n\nARGUMENTS: foo bar`. (maps REQ-CMDS-002-012)
- AC-CMDS-002-09: Given frontmatter `arguments: "9invalid name"` (starts with digit), When loaded, Then the command is rejected with `CommandArgumentNameInvalid`. (maps REQ-CMDS-002-006)
- AC-CMDS-002-10: Given body ` ```bash\necho $ARGUMENTS[0]\n``` `, When rendered with arg `foo`, Then the code fence body becomes `echo foo`. (maps REQ-CMDS-002-041)

## 7. Constraints (제약)

- Technical: Go 1.22+. Shell-quote parser hand-rolled (< 50 LOC) or via `github.com/kballard/go-shellquote` if already depended upon elsewhere (check at implementation time).
- Backward compat: `$ARGUMENTS` semantic unchanged for all existing moai commands.
- Platform: Pure-Go; no OS-specific behavior.
- Performance: Substitution runs in O(N) over body length; no regex backtracking (use single-pass scanner).
- Security: Substitution is pure-text; no eval, no file-read. Escaping responsibility stays with the user's command body.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Shell-quote parser drift from CC's behavior | M | M | Golden tests against a corpus of 50+ real command invocations from CC's test suite, normalized to Go stdlib semantics. |
| `$NN` (e.g. `$12`) ambiguity: is it `$1` followed by `2` or `$12`? | M | M | Scanner reads maximal digits (greedy match), matching CC; edge case documented in docs-site. |
| Named args with special characters in user input bypass sanitization | L | L | Substitution is pure text; user is responsible for shell-escaping within their command body. Documentation emphasizes this. |
| Out-of-range silent empty breaks user expectations | M | M | `command.arguments.strict_mode: true` opt-in converts to explicit error; default is silent-empty matching CC. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3-CMDS-001 (command frontmatter schema introduces `arguments:` field).

### 9.2 Blocks

- SPEC-V3-CMDS-003 (inline bash execution may reference `$ARGUMENTS[N]` for command substitution).

### 9.3 Related

- SPEC-V3-SKL-001 (skill frontmatter shares argument substitution engine).
- SPEC-V3-OUT-001 (progressive typeahead UX surfaces arg hints — covered separately).

## 10. Traceability (추적성)

- Theme: master-v3 §8.4 (Skill SPECs — shared command+skill substitution).
- Gap rows: gm#99 (Medium — `$ARGUMENTS[N]`), gm#100 (Medium — `$N` shorthand), gm#101 (Medium — named args `$name`).
- BC-ID: None (purely additive).
- Wave 1 sources: findings-wave1-hooks-commands.md §8 (argument substitution), §14.5 (command feature gap table), §12 (source references `utils/argumentSubstitution.ts:94-145, 124-127, 50-67`).
- Priority: P2 Medium (ergonomic improvement for command authors; ships Phase 3).
