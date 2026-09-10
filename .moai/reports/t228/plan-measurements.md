# t228 plan-phase measurements

Tree: worktree `.claude/worktrees/t228`, branch `WT-astgrep-16-langs`, base `a1b1ca696`.
Tool: `ast-grep 0.40.5` (`/opt/homebrew/bin/sg`).

## M1. Parser support for the 12 uncovered languages

Command (per language): `echo foo | sg run -p foo -l <lang> --stdin`

| language | result |
|---|---|
| rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, swift | parses (`STDIN:1:foo`) |
| r | `error: invalid value 'r' for '--lang <LANG>': r is not supported!` |
| flutter (dart) | `error: invalid value 'dart' for '--lang <LANG>': dart is not supported!` |

10 of 12 are reachable. `r` and `flutter` have no parser in this ast-grep build — a tool
limitation, not a priority ranking.

## M2. Current template rule inventory

`internal/template/templates/.moai/config/astgrep-rules/` — 26 rules:
go=20, javascript=2, python=2, typescript=2.

Security families and their existing language coverage:

| rule id | languages | severity |
|---|---|---|
| sec-command-injection-{shell,exec} | go, python, javascript, typescript | error |
| sec-hardcoded-credential | go, javascript, python, typescript | error |
| sec-weak-hash-md5 | go | error |
| sec-hardcoded-api-key | go | error |
| sec-hardcoded-jwt-signing-key | go | error |
| sec-csrf-no-token-check | go | warning |
| sec-log-injection-unsanitized | go | warning |
| sec-template-injection-html | go | warning |

## M3. sg config-mode already works

`sg scan --config sgconfig.yml <path>` against `internal/cli` returned
`Error: 11 error(s) found in code.` — config mode is functional today, not a pending gap.

## M4. `sg test` catches the "matches nothing" mutant

Prototype under `/tmp/t228probe` (rule + `valid`/`invalid` test case, `testConfigs.testDir`):

- rule matching the invalid fixture: snapshot flow works
- same rule rewritten to `NeverMatchesAnything::zzz("sh")` (syntactically valid, matches
  nothing): `[Missing] Expect rule probe-rust to report issues, but none found in: ...`,
  and the run exits `FAIL`.

`sg test` is therefore the domain tool that judges "this rule actually fires", and it needs
no dependency on any other card's assets.

## M5. Dependency observed: t217 owns the differential corpus

`internal/hook/security/testdata/scan-corpus/` does not exist on `main`. It exists only in
the unmerged worktree `.claude/worktrees/t217` (branch `WT-security-scan-surface`, HEAD
`efbb21196`), holding 12 fixtures (go/js/ts/python deny+clean, plus `java_uncovered.java`
and `rs_uncovered.rs` placeholders). Card requirement (4) targets a path this card cannot
reach until t217 merges.
