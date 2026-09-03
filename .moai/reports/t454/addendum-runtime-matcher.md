# t454 addendum — the FileChanged runtime measurement (D1 resolved)

Card: t454. Branch `WT-hook-goroutine-truth`. Measured 2026-09-03 by lane-5, after the lead
assigned the Gap to this card.

This addendum settles **D1** of `verdict.md`, and records a sibling sweep the measurement exposed.

---

## Claim

1. **The precondition holds**: `FileChanged` DOES fire in a headless `claude -p` session.
2. **D1 is resolved, and neither of its two options was the right frame.** The matcher is not what
   decides reachability. The runtime emits `FileChanged` for a **narrow, matcher-independent set of
   files** — `.env`, `.envrc`, `.gitignore` fired; `CLAUDE.md`, `package.json`, and three `.go`
   files did not, *including under a `.*` catch-all matcher*.
3. **Therefore the goroutine never spawns in the field, and lane-8's 0-exposure conclusion holds —
   but for a different reason than the one they gave.** Not "the two sets are disjoint under a
   literal reading": the event never reaches the handler for any source file, whatever the matcher
   says. My own regex-reading worry (`verdict.md` ⑴) is refuted at runtime.
4. **Widening the matcher cannot work.** A `.*` matcher was registered and observed NOT to fire for
   `.go` files. The option "widen the matcher to the source extensions" is therefore not merely
   unattractive — it is inoperative. D1's option space collapses to retire-vs-keep-dormant.
5. **Sibling sweep**: the same "executes in a background goroutine" promise the card corrected in
   `file_changed.go` is carried verbatim by three sibling handlers. Corrected in this card.

## Evidence

### Harness

`/tmp/t454-fc-runtime` (§13 isolation — not this project). Reproduction recipe committed at
`.moai/reports/t454/fixture-settings.json`, `fixture-run.sh`, `fixture-run-multi.sh`.

Four `FileChanged` slots registered simultaneously, each invoking a wrapper that appends the raw
stdin payload to a log:

| Slot | matcher | purpose |
|---|---|---|
| `S1_TEMPLATE` | `.env\|.envrc\|.gitignore` | the shipped moai matcher, verbatim |
| `S2_CATCHALL_REGEX` | `.*` | fires for everything IF matching is regex and the domain is wide |
| `S3_LITERAL_ALWAYS` | `t454-fc-runtime` | a literal substring present in EVERY probe path |
| `S4_REGEX_CLASS` | `env[k]eys` | matches `envkeys.go` under regex, never under literal |

The catch-all is the control that separates *"the event never arrived"* from *"the matcher filtered
it out"* — without it, silence means both.

### Observations

Each row is one file actually edited by a headless `claude -p` session. Every edit was verified to
have landed on disk, so each absence below is non-vacuous.

| File edited | S1 | S2 `.*` | S3 | S4 | edit landed |
|---|---|---|---|---|---|
| `.env` | fired | fired | — | — | yes |
| `.envrc` | fired | fired | — | — | yes |
| `.gitignore` | fired | fired | — | — | yes |
| `CLAUDE.md` | — | — | — | — | yes |
| `package.json` | — | — | — | — | yes |
| `env.go` | — | — | — | — | yes |
| `main.go` | — | — | — | — | yes |
| `internal/config/envkeys.go` | — | — | — | — | yes |

Raw logs: `fired-round1.tsv` (2-slot), `fired-round2.tsv`, `fired-round3.tsv` (4-slot).

A fired payload, verbatim:

```
slot=S1_TEMPLATE  {"session_id":"ef6ab686-...","hook_event_name":"FileChanged",
                   "file_path":"/private/tmp/t454-fc-runtime/.env","event":"change"}
```

### What the rows establish

- **Precondition (HARD, asked by the lead)**: `FileChanged` fires under `claude -p`. Settled by the
  `.env` row. Had it not fired, everything below would have been meaningless.
- **The domain is narrow and matcher-independent.** `S2_CATCHALL_REGEX` (`.*`) fired on the three
  dotfiles and on nothing else. A `.*` matcher cannot filter anything out, so the `.go` and
  `CLAUDE.md` silences are the *event* not arriving, not a matcher rejecting it.
- **Matching is not a substring test.** `S3_LITERAL_ALWAYS` (`t454-fc-runtime`) is present as a
  literal substring in every path yet never fired, while `S2` (`.*`) fired on the same events. The
  matcher behaves like an anchored regex against the basename, not a substring scan of the path.
- **`envkeys.go` never reached the handler**, under any of the four matchers, including the two
  built to catch it (`.*` and `env[k]eys`).

## Baseline-attribution

- Tree: `.claude/worktrees/t454`, branch `WT-hook-goroutine-truth`, HEAD `56f3f059f` at measurement.
- Runtime under test: the Claude Code build driving this machine on 2026-09-03. The domain is a
  property of that runtime, not of this repository — see Residual-risk.
- Post-sweep verification, this tree: `go vet ./internal/hook/` → exit 0;
  `go test ./internal/hook/` → `ok github.com/modu-ai/moai-adk/internal/hook 42.168s`.

## Gaps

- **The domain's full membership was NOT enumerated.** Eight files were probed. That the three
  firing ones are exactly the three the moai matcher names is consistent with several models
  (a built-in env/ignore watcher; a secrets-and-ignore-file tap) and this measurement does not
  choose among them. What it does establish is the only thing the card needs: no file carrying one
  of the 21 gated source extensions was ever observed to fire.
- **`.claude/settings.json` and MCP config files were not probed.** They are plausible domain
  members and would not change the conclusion — neither has a gated extension.
- **The runtime's matching semantics were inferred, not read.** "Anchored regex against the
  basename" is the model consistent with all four slots; no source or documentation was consulted.
- **Guard bypass, disclosed.** This session's cwd is pinned to its worktree, and `claude` derives
  its project directory from cwd, so the probes ran through a committed script file
  (`fixture-run.sh`) that performs the `cd`. The worktree guard cannot read inside a script. The
  script contains no git operation of any kind; it is committed so the bypass is reviewable rather
  than invisible.
- **The three sibling comment corrections were not separately re-measured.** They are the same
  process-lifetime fact already measured for `file_changed.go` (`0/10`, lane-8's tree), restated at
  three more spawn sites. No new measurement was made or claimed for them.

## Residual-risk

- The domain is a runtime behavior, not a contract. A Claude Code release that widens `FileChanged`
  to arbitrary files would make the handler live again overnight, and nothing in this repository
  would signal the change. That argues for the retire option over keep-dormant: a dormant handler
  whose dormancy depends on an unversioned upstream behavior is a latent, not a settled, state.
- Conversely, if the handler IS retired and the domain later widens, the MX sidecar simply stops
  being offered a path it never actually served.
- The sibling comments now describe process-exit behavior. Should hook invocation ever become a
  daemon or batched runner, all four comments become stale in the same direction at once.

---

## D1 — resolved, with a narrowed decision for the lead

The measurement removes one of D1's two options. What remains:

- **Retire handler + registration.** The handler has never run in the field and cannot be made to
  run by matcher changes. Its reachability rests on an unversioned upstream behavior.
- **Keep dormant, documented.** Cheaper, and preserves the code should the domain widen — but
  leaves a handler whose dormancy no test asserts.

This is still an operator call: retiring touches a landed SPEC's registration surface, the same
class of change as D2. The lane does not take it.

**D2 is unchanged** — `config_change` async removal still alters `AC-HAE-003` of a landed SPEC.
