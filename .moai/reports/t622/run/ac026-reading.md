# AC-GDP-026 reading record (M1)

Read by: manager-develop (run-phase part 1, card t622). Target list: every line of
`ac026-local-merge-lines.txt` and `ac026-template-merge-lines.txt` (5 lines each). The nine local
fragments and nine template fragments are byte-identical (`ac026-frag-lt.txt`: all nine
`lt-diff=0`), and the two merge-line lists carry the same five line bodies.

Question for every line: "Does this line say `--merge` is a deprecated alias of `--auto-merge` —
not a reversed relation, a separate flag, an un-deprecated flag, or a flag with its own behavior?"

| Copy | Fragment:line | Line | Answer |
|---|---|---|---|
| local | skill:1 | `Modes: auto, force, status, project. Flags: --auto-merge, --merge (deprecated alias of --auto-merge), --skip-mx` | Yes — `--merge` is marked as the deprecated alias of `--auto-merge`; `--auto-merge` is listed as the flag |
| local | qgc-flags:5 | `- --merge: deprecated alias of --auto-merge (logs a warning).` | Yes — alias direction correct, no own behavior beyond the warning |
| local | sync:1 | `**Flags**: ... \| --auto-merge (auto-merge 옵트인) \| --merge (deprecated alias of --auto-merge) \| ...` | Yes |
| local | dl:11 | `- OR --merge flag set (deprecated alias of --auto-merge, logged as warning)` | Yes — the trigger is the alias of the `--auto-merge` trigger, not a separate condition |
| local | dl:26 | `- --merge: deprecated alias of --auto-merge (logs a warning).` | Yes |
| template | skill:1 | same text as local skill:1 | Yes |
| template | qgc-flags:5 | same text as local qgc-flags:5 | Yes |
| template | sync:1 | same text as local sync:1 | Yes |
| template | dl:11 | same text as local dl:11 | Yes |
| template | dl:26 | same text as local dl:26 | Yes |

Fragments without `--merge` (ref, qgc-args, sync-usage, hint, dl-next) expose `--auto-merge` only
(per (i) counts in `ac026-{local,template}-i.txt`), which (iii) permits.

Result: all ten answers are "Yes". No line reverses the alias, marks `--auto-merge` deprecated, or
gives `--merge` separate behavior.
