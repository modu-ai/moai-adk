# Coverage Matrix — SPEC-ASTGREP-LANG16-001

> Authored contract artifact (plan.md §A.3): one cell per (security family,
> parseable language) pair, verified by the checker wired into the Go test
> suite (`internal/astgrep/coverage_matrix_test.go`). The matrix document may
> drift from the ruleset it describes because `sg test` never reads it; the
> checker's fourth failure class exists precisely to catch that drift.

## Axes

**Family axis (8, fixed)** — F1 command injection, F2 hardcoded credential,
F3 weak hash, F4 hardcoded API key, F5 hardcoded JWT signing key, F6 CSRF
token absence, F7 log injection, F8 template injection / XSS.

**Language axis (14 parseable)** — go, javascript, python, typescript, rust,
java, kotlin, csharp, ruby, php, elixir, cpp, scala, swift. Derived under
ast-grep **0.40.5**, re-probed at M2 rather than copied forward.

## Cell states

- `IMPLEMENTED` — carries the shipped rule id; its `sg test` case pair is the
  structural proof. No rationale column content beyond `-`.
- `EXEMPT` — carries a rationale naming the missing construct AND an evidence
  entry starting `cite:` (a named language / stdlib / framework reference) or
  `probe:` (the invocation run and its observed output).
- `PENDING` — reserved key awaiting its fill milestone. This is the interim
  marker authorized by plan.md §B M2 item 3: only the 14 already-implemented
  cells are resolved now; the remaining 98 belong to SPEC-ASTGREP-BREADTH-001,
  which fills them as either IMPLEMENTED or EXEMPT-with-evidence. A PENDING
  cell satisfies nothing on its own.

| family | language | state | rule id / rationale | evidence |
|---|---|---|---|---|
| F1 | go | IMPLEMENTED | sec-command-injection-shell | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F1 | javascript | IMPLEMENTED | sec-command-injection-exec | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F1 | python | IMPLEMENTED | sec-command-injection-shell | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F1 | typescript | IMPLEMENTED | sec-command-injection-exec | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F1 | rust | PENDING | - | - |
| F1 | java | PENDING | - | - |
| F1 | kotlin | PENDING | - | - |
| F1 | csharp | PENDING | - | - |
| F1 | ruby | PENDING | - | - |
| F1 | php | PENDING | - | - |
| F1 | elixir | PENDING | - | - |
| F1 | cpp | PENDING | - | - |
| F1 | scala | PENDING | - | - |
| F1 | swift | PENDING | - | - |
| F2 | go | IMPLEMENTED | sec-hardcoded-credential | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F2 | javascript | IMPLEMENTED | sec-hardcoded-credential | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F2 | python | IMPLEMENTED | sec-hardcoded-credential | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F2 | typescript | IMPLEMENTED | sec-hardcoded-credential | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F2 | rust | PENDING | - | - |
| F2 | java | PENDING | - | - |
| F2 | kotlin | PENDING | - | - |
| F2 | csharp | PENDING | - | - |
| F2 | ruby | PENDING | - | - |
| F2 | php | PENDING | - | - |
| F2 | elixir | PENDING | - | - |
| F2 | cpp | PENDING | - | - |
| F2 | scala | PENDING | - | - |
| F2 | swift | PENDING | - | - |
| F3 | go | IMPLEMENTED | sec-weak-hash-md5 | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F3 | javascript | PENDING | - | - |
| F3 | python | PENDING | - | - |
| F3 | typescript | PENDING | - | - |
| F3 | rust | PENDING | - | - |
| F3 | java | PENDING | - | - |
| F3 | kotlin | PENDING | - | - |
| F3 | csharp | PENDING | - | - |
| F3 | ruby | PENDING | - | - |
| F3 | php | PENDING | - | - |
| F3 | elixir | PENDING | - | - |
| F3 | cpp | PENDING | - | - |
| F3 | scala | PENDING | - | - |
| F3 | swift | PENDING | - | - |
| F4 | go | IMPLEMENTED | sec-hardcoded-api-key | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F4 | javascript | PENDING | - | - |
| F4 | python | PENDING | - | - |
| F4 | typescript | PENDING | - | - |
| F4 | rust | PENDING | - | - |
| F4 | java | PENDING | - | - |
| F4 | kotlin | PENDING | - | - |
| F4 | csharp | PENDING | - | - |
| F4 | ruby | PENDING | - | - |
| F4 | php | PENDING | - | - |
| F4 | elixir | PENDING | - | - |
| F4 | cpp | PENDING | - | - |
| F4 | scala | PENDING | - | - |
| F4 | swift | PENDING | - | - |
| F5 | go | IMPLEMENTED | sec-hardcoded-jwt-signing-key | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F5 | javascript | PENDING | - | - |
| F5 | python | PENDING | - | - |
| F5 | typescript | PENDING | - | - |
| F5 | rust | PENDING | - | - |
| F5 | java | PENDING | - | - |
| F5 | kotlin | PENDING | - | - |
| F5 | csharp | PENDING | - | - |
| F5 | ruby | PENDING | - | - |
| F5 | php | PENDING | - | - |
| F5 | elixir | PENDING | - | - |
| F5 | cpp | PENDING | - | - |
| F5 | scala | PENDING | - | - |
| F5 | swift | PENDING | - | - |
| F6 | go | IMPLEMENTED | sec-csrf-no-token-check | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F6 | javascript | PENDING | - | - |
| F6 | python | PENDING | - | - |
| F6 | typescript | PENDING | - | - |
| F6 | rust | PENDING | - | - |
| F6 | java | PENDING | - | - |
| F6 | kotlin | PENDING | - | - |
| F6 | csharp | PENDING | - | - |
| F6 | ruby | PENDING | - | - |
| F6 | php | PENDING | - | - |
| F6 | elixir | PENDING | - | - |
| F6 | cpp | PENDING | - | - |
| F6 | scala | PENDING | - | - |
| F6 | swift | PENDING | - | - |
| F7 | go | IMPLEMENTED | sec-log-injection-unsanitized | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F7 | javascript | PENDING | - | - |
| F7 | python | PENDING | - | - |
| F7 | typescript | PENDING | - | - |
| F7 | rust | PENDING | - | - |
| F7 | java | PENDING | - | - |
| F7 | kotlin | PENDING | - | - |
| F7 | csharp | PENDING | - | - |
| F7 | ruby | PENDING | - | - |
| F7 | php | PENDING | - | - |
| F7 | elixir | PENDING | - | - |
| F7 | cpp | PENDING | - | - |
| F7 | scala | PENDING | - | - |
| F7 | swift | PENDING | - | - |
| F8 | go | IMPLEMENTED | sec-template-injection-html | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |
| F8 | javascript | PENDING | - | - |
| F8 | python | PENDING | - | - |
| F8 | typescript | PENDING | - | - |
| F8 | rust | PENDING | - | - |
| F8 | java | PENDING | - | - |
| F8 | kotlin | PENDING | - | - |
| F8 | csharp | PENDING | - | - |
| F8 | ruby | PENDING | - | - |
| F8 | php | PENDING | - | - |
| F8 | elixir | PENDING | - | - |
| F8 | cpp | PENDING | - | - |
| F8 | scala | PENDING | - | - |
| F8 | swift | PENDING | - | - |

## Excluded languages

Two of the sixteen supported languages have no parser under ast-grep 0.40.5
and therefore contribute no cells above. Like every other absence here, this
is a statement of current scope, recorded verbatim below: they remain equal-priority future additions once their parsers arrive, never unsupported ones.

| language | reason | evidence |
|---|---|---|
| r | no parser registered in this ast-grep build (`r is not supported!`) | see probe record below |
| flutter | no parser registered in this ast-grep build (`flutter is not supported!`) | see probe record below |

Verbatim probe records, re-run at M2 (both exit rc=2):

```text
$ printf 'x <- 1\n' | sg run -l r --stdin
error: invalid value 'r' for '--lang <LANG>': r is not supported!

For more information, try '--help'.

$ printf 'void main() {}\n' | sg run -l flutter --stdin
error: invalid value 'flutter' for '--lang <LANG>': flutter is not supported!

For more information, try '--help'.
```

## Fill accounting

14 / 112 cells IMPLEMENTED at M2 seed; 98 PENDING. The fill obligation and
the exemption-evidence standard are inherited whole by
SPEC-ASTGREP-BREADTH-001 (its plan.md §C treats this as non-renegotiable
contract).
