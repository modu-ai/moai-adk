# C10b three-arm answers and raw context (#2186)

The repair implements the user decision recorded in [PR #2440 comment 6790809622](https://github.com/modu-ai/moai-adk/pull/2440#issuecomment-6790809622).
The 2026-11-11 measurement uses product source `4f4d5600cb274e8ca4716d9d9adbb2237b65abb0`
on macOS arm64, plus this runner repair. The fixed C10a corpus, source snapshot,
judge, ground truth and precision floor **0.9212** are unchanged.

## Execution and selection

`python3 tools/atlas_three_arm.py --binary <moai-cli> --output-dir <directory>`
materializes C10a's frozen source in an isolated git fixture, builds one index,
and runs all 42 scored tasks in each of three cumulative arms:

- `rg`: product `search` with the literal query, verified equal to pinned rg file sets.
- `+Map`: search, then `atlas-map` selecting those paths.
- `+Graph`: both stages, then `atlas-graph` from the Map file nodes, three hops,
  at most 5000 edges. The producer's `imports`/`calls` remain enabled.

Child processes use a synthetic HOME and an environment allowlist. Every answer
must be fresh, have the same generation, have no omissions, and remain within
the fixed snapshot. T-033 has no lexical seed: it is still scored, while its
Map/Graph calls are skipped to avoid an accidental all-files Map query.

The answer policy is `literal-evidence-v1`, applied identically to all tasks.
It receives **only the query pattern, returned Map/Graph records and frozen file
bytes**. It does not receive task id, task kind, ground truth or judge results.

1. Map's returned file nodes provide eligible starting files.
2. Graph may admit additional file candidates reachable from those starts in at
   most three hops, in either direction, through returned edges with nonempty
   `evidence_refs`. Returned but disconnected nodes are not eligible. Unknown
   files and dangling evidence endpoints are refused.
3. A candidate is an answer only if its bytes contain the query's exact UTF-8
   literal. No oracle-based reranking, task-specific exceptions or threshold
   changes are used. Each accepted answer records its supporting edge-id path
   (empty for a Map start).
4. The unchanged C10a judge scores only these answer files. Literal selection is
   intentionally conservative; it is not a semantic relevance model.

Thus Graph evidence can discover an answer absent from the preceding seed set;
removing its connecting edge or evidence changes the selected answers. The fixed
corpus happens to contain all known relevant files in 32/33 lexical task results;
T-033 has no seed. Equal answer scores in this corpus do not establish improved
retrieval or justify discarding Graph output.

## Raw context remains separately inspectable

Every task report keeps `hits` (answers), `selection` (eligible files and accepted
support paths) and `raw_context` (Map/Graph document references, complete raw file
lists, and Graph hop distances). [raw-contexts.json](evidence/atlas-c10b/raw-contexts.json)
stores the **complete** returned node/edge/missing documents. Identical documents
are deduplicated by SHA-256 of canonical JSON; every reference resolves within
that file. Filtering answers never edits the raw documents or removes unrelated
context. The generation id binds the registry to the task reports.

## Measured results

All arms score 33 RETRIEVAL and 9 GRAPH tasks using fixed inclusive macro-mean
floors. Reports and summary are in [evidence/atlas-c10b](evidence/atlas-c10b/summary.json).
Response bytes count full product stdout, including unrelated raw context; they
are **not model tokens**. Calls exclude the shared index build and judge.

| Arm | Answer precision | Recall | GRAPH file proxy | Judge | Answer files | Raw file context | Response bytes | Calls |
| --- | ---: | ---: | ---: | --- | ---: | ---: | ---: | ---: |
| rg | 0.9212 | 0.9697 | 1.0000 | passed | 140 | 140 | 124478 | 42 |
| +Map | 0.9212 | 0.9697 | 1.0000 | passed | 140 | 140 | 242224 | 83 |
| +Graph | 0.9212 | 0.9697 | 1.0000 | passed | 140 | 1805 | 7612796 | 124 |

Graph retains 1805 task/file context pairs, including 1665 files excluded from
answers. Its 9696 returned edges, plus 85 Map edges, remain 9781 cumulative view
edges. Each arm observed 376 fresh lexical spans and stale ratio 0. The registry
preserves these records; the test compares saved raw edge totals to the actual
operation totals and checks complete node-distance coverage.

Before selection, the same product's indiscriminate context union scored
precision 0.1958. Read-only in-memory ablation of code-reference edges restored
0.9212, with 41/41 native traversal replay matches. This was diagnostic evidence,
not the adopted repair: disabling those edges would discard the new capability.
See [measurement and ablation report](https://github.com/modu-ai/moai-adk/pull/2440#issuecomment-6790282125).

## Regression and mutation checks

`tests/atlas_three_arm.rs` executes all arms, requires search42/Map41/Graph41
calls, verifies scored answers and independently resolves saved raw documents,
files, distances and edge counts. It also runs the Python selection regressions.
`tests/atlas_three_arm_selection_test.py` uses small synthetic examples separate
from the frozen corpus: a relevant Graph-only file must become an answer, an
unrelated connected file stays raw context, a disconnected literal match cannot
leak in, and removing Map/Graph evidence changes the result. A synthetic end-to-end adapter
with a deliberately restricted search seed set also tests the actual runner call
site: a Graph-only relevant file must reach answers while unrelated context is
retained. It supplies no ground truth to the runner and replaces the judge only
for that pipeline regression; it is not a native retrieval-quality measurement.
Fourth-hop and
out-of-scope nodes are checked. No fixed fixture is rewritten.

Mutation runs in the author's evidence directory check Graph discard, edge
evidence discard, relevance removal, raw-edge discard, and Map/Graph-result
discard at the actual runner call site against those semantic
assertions, not only invocation counts. Actual command results, restored controls,
full checks and exact-head CI are recorded in the PR. The shared main.rs stack
limit tracked by #2452 is not bypassed with a stack override.

## Scope and provenance

GRAPH's 1.0 is C10a's referring-file inclusion proxy, not direct code-reference
correctness. The producer's separate direct-edge contract is 7/9 plus two stated
gaps (#2442/#2450); this runner does not reclassify it as 9/9. Tokens, C10b's eight
FRESHNESS mutations, live model hosts and semantic relation scoring remain
`not_run`. The original 2026-11-10 table is superseded by this measurement; git
history retains it. No retrieval improvement is claimed from equal scores.

Python 3.14 official [unittest documentation](https://docs.python.org/3.14/library/unittest.html)
and [hashlib documentation](https://docs.python.org/3.14/library/hashlib.html) were
successfully read 2026-11-11T08:24:53Z. Adopted stdlib discovery/assertions and
SHA-256 document identity; verified with the local regressions. Received HTML
SHA-256: `4fc602e998d0f8172d2ebc4e46282fa8b03fd7cf1f762a10e0d7eda14a461d56`
and `797996dc3699c45c8f7009773946a1c877729165b20b54de2cef99f50138fb01`.
No new dependency is introduced. Child execution retains the existing argv and
explicit-env subprocess protocol.
