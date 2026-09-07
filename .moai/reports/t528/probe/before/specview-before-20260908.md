# t528 M0 before-image — pre-widening `moai spec view` corpus sweep

- Tree SHA: `52f863f36` (`internal/spec/parser.go` UNEDITED at capture time; only the
  promoted probe `internal/spec/zz_t528_anchor_probe_test.go` was added)
- Binary: `go build -o /tmp/t528-moai-before ./cmd/moai` — built from this tree, `BUILD_EXIT=0`
- Denominator: `.moai/reports/t528/probe/before/filelist.txt` (807)
- Sweep script: `/tmp/t528-specview-sweep.sh` (verbatim copy in this directory as `specview-sweep.sh`)

## [HARD] Correction 1 — the flag cited in spec.md / acceptance.md does not exist

`spec.md` §1.3, `plan.md` §E M7 and `acceptance.md` §D.12/§D.15 all name the command
`moai spec view <ID> --acceptance`. **There is no `--acceptance` flag.**
`internal/cli/spec_view.go:37` registers exactly one flag, `--shape-trace`; the acceptance
view *is* `moai spec view <ID>`.

Measured, verbatim (tree `52f863f36`, binary `/tmp/t528-moai-before`):

```
$ /tmp/t528-moai-before spec view SPEC-AC-COLLECTOR-ANCHOR-001 --acceptance
   ERROR
  Unknown flag: --acceptance.
  Try --help for usage.
RC=1
```

A corpus sweep using the documented flag returns `SWEPT=807  NONZERO_EXIT=807` — an
807/807 "failure" that measures argument parsing and says nothing about the collector.
That first sweep is recorded here precisely because it is the shape this card has been
caught by five times: a confident figure produced by a mis-shaped discriminator.
**All figures below use the real command.**

## Sweep result (real command, `moai spec view <ID>`)

```
SWEPT=807  NONZERO_EXIT=448
  parse error:            442   all of shape "acceptance criteria section not found"
  spec.md not found for:    6   sweep-addressability artifact, NOT a parse error
  addressable SPECs:      801
```

The 6 are corpus entries living under `.moai/specs/_archive/<ID>/spec.md`, which the CLI
cannot address by bare SPEC-ID (it joins `.moai/specs/<ID>/spec.md`). They are excluded
from the hard-error count because they never reach the parser.

## [HARD] Correction 2 — the hard-error before-image is 442, not 0

`plan.md` §E M7 and `acceptance.md` §D.15 pin the `spec view` hard-error before-image at
**0**, citing `probe/duplicate-20260908.txt`. That artifact measures **duplicate-ID
material**, which is genuinely 0 — it does not measure CLI hard errors. Two different
quantities; the 0 was carried from one to the other. `acceptance.md` §D.15 in fact says so
itself under **Gap** ("실제 `moai spec view`를 코퍼스 전체에 돌린 적은 아직 없다"), so this
measurement closes that gap rather than contradicting it.

Measured here for the first time by actually running the CLI over the corpus:

| quantity | value |
|---|---|
| hard errors (`parse error:`) pre-widening | **442** |
| of which `DuplicateAcceptanceID` / depth class | **0** |

All 442 are `acceptance criteria section not found`, raised by
`parseAcceptanceCriteriaInternal` when `findACSectionStart` returns -1 and delivered as a
fatal error by the `default:` branch at `internal/cli/spec_view.go:85`. That is the
**heading axis**, explicitly out of this card's scope (`spec.md` §4), and it is untouched
by the item-grammar widening.

The operative gate is therefore the one REQ-ACA-001-013 actually states — hard errors
**not present in the baseline** == 0 — measured as a delta against this 442, with the
duplicate/depth class (0) watched separately since that is the class the widening could
newly create.

## Sample lines (plan.md §D.1 — no corpus figure adopted without opening the lines)

```
SPEC-AC-COUNT-DISCRIMINATOR-001  RC=1  ERROR Parse error: acceptance criteria section not found.
SPEC-ADVISOR-RUNG-001            RC=1  ERROR Parse error: acceptance criteria section not found.
SPEC-AGENT-BOUNDARY-DRIFT-001    RC=1  ERROR Parse error: acceptance criteria section not found.
SPEC-AGENT-MEMORY-DRAIN-001      RC=1  ERROR Parse error: acceptance criteria section not found.
...
SPEC-AGENCY-001                  RC=1  ERROR Spec.md not found for SPEC-AGENCY-001 at .../.moai/specs/SPEC-AGENCY-001/spec.md.
SPEC-THIN-CMDS-001               RC=1  ERROR Spec.md not found for SPEC-THIN-CMDS-001 at .../.moai/specs/SPEC-THIN-CMDS-001/spec.md.
```

Shape tally over the 448 nonzero lines (`sort | uniq -c`):

```
442  ERROR Parse error: acceptance criteria section not found.
  6  ERROR Spec.md not found for <ID> at .../.moai/specs/<ID>/spec.md.
```

No other shape appears — in particular no `DuplicateAcceptanceID` and no depth error.
