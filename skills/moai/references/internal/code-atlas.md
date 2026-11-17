# Internal: code-atlas

When to read: in READ or PLAN, before you search the code for where something is, what depends on a file, or what a change could touch.

The code atlas is the project's published index generation under its state folder (`<state>/index/`). `/moai project` offers to build it (`moai index-build`); READ and PLAN only read it. Every command below takes the project folder as its argument and a JSON request on stdin, writes nothing and uses no network.

1. **Check freshness first.** Run `moai index-query <project>` with `{"protocol_version":1}`. Its `result.state` decides:
   - `fresh`: use the atlas.
   - `stale`: use it for every file not named in `result.changes` (`added`, `deleted`, `modified`, `new_files`), and read the named files directly. When `changes.branch_switched` is true, treat it as `unavailable`.
   - `partial` or `unavailable`: do not use it. Search with `rg` (or `grep`) and read the files.
2. **Never build or rebuild it here.** `moai index-build` writes the state folder; it runs only after the user agrees, from the PROJECT workflow. When the atlas is missing or stale, say so once in the answer and that `/moai project` can build it; never build it silently.
3. **Ask narrowly.**
   - What depends on a file, and what it uses: `moai atlas-graph` with `{"protocol_version":1,"query":{"start":["file:<path>"]}}` (one hop by default). Raise `limits.max_hops` only for one named file: on the moai-adk repository two hops from one file returned about 21 times the one-hop answer.
   - The files and their `@MX`/`@MOAI` tags under a path: `moai atlas-map` with `{"protocol_version":1,"query":{"paths":["<file>"]}}` or `"prefixes":["<folder>"]`. A whole top-level folder can be tens of kilobytes; name the files or the smallest folder.
   - Where a literal appears, with cited lines: `moai search` with `{"protocol_version":1,"query":{"literals":["<text>"]}}`.
4. **An atlas answer is a lead, not a fact.** Edges are `unreviewed`, imports and calls are file to file for Rust only, and there are no symbol records (`coverage.symbols` says so). Before you state a dependency, open the line its `evidence_refs` cite and read it. Re-check `state` in each answer: when it is no longer `fresh`, step 1 applies to that answer.
5. **Record what you used.** In the answer or plan, give the `generation_id` and `state` you read, or say that you searched without the atlas and why.
