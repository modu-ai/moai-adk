auditor-model: claude-opus-5-5

# t2109 sync-audit, iteration 2 (independent, Claude)

Card: t2109 (Class B, no SPEC): YAML key removal ignores anchors.
Scope judged: a35c589b94..fff52b38601995a69a7522ae5c9a9cb7dd62aeb4 (commits 35911488199b, the guard, and fff52b3860, the F1 repair).
Tree: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2109, branch WT-user-yaml-yaml, HEAD fff52b38601995a69a7522ae5c9a9cb7dd62aeb4, working tree clean before and after this audit.

## Verdict

FAIL. The only cause is the required codex gate, which went unmet. The code itself passes: every check of the card change came out clean, and no blocking code defect was found.

- Code-merit score: 0.93 (Functionality 95, Security 95, Craft 90, Consistency 88, all at or above threshold).
- Overall: FAIL, because `moai verify audit-plan` for this tree reports `codex` as an explicit `required` gate (`enforced_required: ["codex"]`), and the one allowed `mcp__moai__codex_audit` call was refused with `stale_server`. A required backend that does not answer stays fail-closed, so this audit cannot yield PASS. The receipt from iteration 1 (rcpt-011dcf46cb04b0827be82f37) was minted for tree 35911488199b, before this audit began. It does not cover this head.
- To clear it: once the moai MCP server is reconnected, run one codex audit over the delta 35911488199b..fff52b3860. If that review raises no blocking finding, the code-merit result above stands as PASS. No code change is requested.

verdict: FAIL
audited_sha: fff52b38601995a69a7522ae5c9a9cb7dd62aeb4

## Findings

| id | severity | blocking | location | one line |
|---|---|---|---|---|
| G1 | gate | blocking (gate only, not code) | codex required gate | `codex_audit` was refused with `stale_server`, so the required codex gate is unmet and the gate has no receipt for this head |
| F1 (iter 1) | High | resolved | in-diff yamlpatch.go:451 | Absent-key removal under an anchored mapping used to be refused. It is now a no-op: reproduced as refused at 35911488199b and accepted at fff52b3860, with anchor and alias kept |
| F2 (iter 1) | Low | resolved | in-diff test file:42-43 | The key-node-anchor and grandparent-anchor cases were added. Mutants that ignore the key node, or look only at the direct parent, are now killed |
| F3 | Low | optional (limitation, pre-existing) | touched-file, removeKey / merge keys | Removing a key that overrides a `<<: *base` key makes it read as the inherited value rather than absent. Removing a key that exists only through the merge is a silent no-op. Base a35c589b94 behaves the same. Aliases and parseability are not affected, so this is not a defect the card's text covers |
| F6 | Medium | optional (outside, pre-existing) | touched-file, patchFile re-encode | Any write on the re-encode path drops every YAML document after the first `---`. Base a35c589b94 behaves the same, and the card diff does not touch patchFile. Real data loss: it needs its own card |
| F7 | Low | optional (in-diff, process) | .moai/reports/t2109/sync-audit.md | The iteration-1 verdict file was force-added (`.gitignore:235 .moai/reports/*` ignores it) in fff52b3860. A tracked FAIL verdict file with its own machine lines will reach main if the card lands as is. Audit exports are local by design. Leader's call whether to keep it |
| F8 | Info | optional (touched-file, pre-existing) | encode | Re-encoding writes an explicit `!!merge <<: *b` tag. The output still parses. Same on base |

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim 1: The F1 repair holds. A launcher-shaped write now succeeds and keeps anchor and alias. At 35911488199b it was refused.

Evidence: a probe copy of the package (head yamlpatch.go plus the package tests plus `probe_matrix_test.go`) in the scratchpad, built against gopkg.in/yaml.v3 v3.0.1 from the module cache with GOPROXY=off. The edit list was: set `llm.team_mode`, upsert `llm.new_key`, remove the absent `llm.glm.models.{opus,sonnet,haiku}` under `models: &m`, with alias `other.shared: *m`.

Head, `go test -count=1 -run TestProbe -v .`:
```
=== RUN   TestProbeLauncherShape
    probe_matrix_test.go:34: err=<nil>
        llm:
          team_mode: glm
          glm:
            models: &m
              high: glm-5.3
              low: glm-5.3
          new_key: v
        other:
          shared: *m
--- PASS: TestProbeLauncherShape (0.00s)
```
The same probe with the 35911488199b yamlpatch.go swapped in:
```
--- FAIL: TestProbeLauncherShape (0.00s)
    probe_matrix_test.go:36: PROBE-F1 refused: yamlpatch: .../f.yaml: path "llm.glm.models.opus": segment "opus" parent carries anchor &m, so removing a key would change every alias of it; edit the file by hand
--- FAIL: TestPatchFileStillRemovesUnanchoredKeys (0.00s)
    --- FAIL: TestPatchFileStillRemovesUnanchoredKeys/absent_key_under_an_anchored_mapping_beside_a_real_edit (0.00s)
        yamlpatch_remove_anchor_t2109_test.go:96: a write whose removals are absent keys was refused: ... path "llm.glm.models.opus": segment "opus" parent carries anchor &m, ...
FAIL
```
The second failure shows that the author's new RED control really fails on the pre-repair code.

Reasoning about the early return (code read, yamlpatch.go:426-439 and 498-508): `removeKey` mutates only when `lookupNode(root, path[:n-1])` is a mapping that holds the last key. That is exactly when `lookupNode(root, path)` is non-nil, because yaml.v3 never stores a nil value node (null is a ScalarNode) and it rejects duplicate keys. So the early return fires if and only if `removeKey` is a no-op, and no present-key removal can slip through it.

### Claim 2: The full refusal matrix still holds with the early return. Every refused case leaves the file byte-identical.

Evidence, head probe `TestProbeRefusalMatrix` (17 subtests, each asserting `err != nil` and output equal to input). Result `--- PASS: TestProbeRefusalMatrix (0.01s)`, with refusal messages including:
```
anchor_on_key_node: path "user.name": the key carries anchor &k, so removing it would leave its aliases undefined
anchor_on_scalar_value: path "user.name": the target (or a node inside it) carries anchor &n, ...
anchor_on_mapping_value: path "user": the target (or a node inside it) carries anchor &u, ...
anchor_deep_in_removed_value: path "user": the target (or a node inside it) carries anchor &d, ...
anchor_in_seq_inside_removed: path "user.list": the target (or a node inside it) carries anchor &s, ...
anchored_direct_parent: path "user.name": segment "name" parent carries anchor &u, ...
anchored_grandparent: path "a.b.c": segment "b" parent carries anchor &x, ...
anchored_great_grandparent: path "a.b.c.d": segment "b" parent carries anchor &x, ...
anchored_root: path "user.name": segment "user" parent carries anchor &r, ...
flow_anchored_parent: path "user.name": segment "name" parent carries anchor &u, ...
flow_anchor_on_value: path "user.name": the target (or a node inside it) carries anchor &n, ...
merge_base_key_removed: path "base.k": segment "k" parent carries anchor &b, ...
merge_base_removed_whole: path "base": the target (or a node inside it) carries anchor &b, ...
fine_edit_then_refused_removal: path "user.name": the target ... carries anchor &n, ...
fine_upsert_then_refused_removal: path "user.name": segment "name" parent carries anchor &u, ...
fine_removal_then_refused_removal: path "user.name": the target ... carries anchor &n, ...
absent_then_refused: path "user.name": segment "name" parent carries anchor &u, ...
```
Allowed matrix, head `TestProbeAllowedMatrix`: all 7 subtests PASS with `err=<nil> parseErr=<nil>`. The cases were: remove an alias-valued key, remove the `<<` merge key itself, remove a merge-override key (F3), remove through an alias (no-op, file `tmpl: &u ... user: *u` intact), remove an absent key under an anchored mapping, remove a sibling of an anchored node, and remove an absent key deep under an anchored ancestor.

### Claim 3: The card's tests kill every mutant asked about, plus three more.

Evidence: `python3 -I .../scratchpad/probe/mutate.py`. It applies each mutant to a fresh copy of the head package, with my probe file removed so that only the card's and the package's own tests run, and then runs `go test -count=1 .`:
```
M1_ignore_key_node: KILLED (exit 1); failing: [... TestPatchFileRefusesRemovingAnchoredKeys/the_key_node_itself_carries_the_anchor]
M2_direct_parent_only: KILLED (exit 1); failing: [... /an_anchored_ancestor_above_the_direct_parent]
M3_no_deep_anchor: KILLED (exit 1); failing: [... /removed_subtree_holds_an_anchor_definition]
M4_refuse_absent_again: KILLED (exit 1); failing: [... TestPatchFileStillRemovesUnanchoredKeys/absent_key_under_an_anchored_mapping_beside_a_real_edit]
M5_no_value_anchor_check: KILLED (exit 1); failing: [... key_whose_mapping_value_defines_an_anchor, key_whose_scalar_value_defines_an_anchor, removed_subtree_holds_an_anchor_definition]
M6_no_parent_check: KILLED (exit 1); failing: [... key_inside_an_anchored_mapping_shared_by_an_alias, an_anchored_ancestor_above_the_direct_parent]
M7_guard_not_called: KILLED (exit 1); failing: [... 5 refusal subtests]
```
The refusal tests also assert `out != tc.in`, so they check the file stays unchanged (test file lines 51-53).

### Claim 4: F3 and F6 are pre-existing and do not come from this card.

Evidence: `probe_f6_test.go`, run against the head copy and against a copy holding `git show a35c589b94:internal/settings/yamlpatch/yamlpatch.go`:
```
== head
    probe_f6_test.go:18: err=<nil> out="a: 1\nc: 3\n"
    probe_f6_test.go:20: F6: second document dropped
    probe_f6_test.go:32: err=<nil> out="base: &b\n  k: 1\nchild:\n  !!merge <<: *b\n"
== base a35c
    probe_f6_test.go:18: err=<nil> out="a: 1\nc: 3\n"
    probe_f6_test.go:20: F6: second document dropped
    probe_f6_test.go:32: err=<nil> out="base: &b\n  k: 1\nchild:\n  !!merge <<: *b\n"
```
The outputs are identical. Code reading agrees: `patchFile` unmarshals one `yaml.Node` and encodes only that document (yamlpatch.go:126-165), and the card's hunks are only at applyEdit (:359) and the new functions (:438+). Input `a: 1\n---\nb: 2\n` loses `b: 2`. F3: `child.k: 9` is removed and child now reads `k: 1` through the merge. The card text asks to protect aliases and keep the file parseable, and both still hold, so F3 is a limitation, not a defect of this card.

### Claim 5: The test suites named in the brief pass at head, and formatting is clean.

Evidence, each run as a plain separate command in the tree:
```
$ timeout 300 go test -count=1 ./internal/settings/...
ok  	github.com/modu-ai/moai-adk/internal/settings	1.496s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	0.302s
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.585s

$ timeout 400 go test -count=1 ./internal/config/
ok  	github.com/modu-ai/moai-adk/internal/config	15.233s

$ timeout 500 go test -count=1 ./internal/cli/ -run 'GLM|Glm|LLM|Llm'
ok  	github.com/modu-ai/moai-adk/internal/cli	39.214s

$ gofmt -l internal/settings/yamlpatch
(no output) exit=0

$ timeout 200 go test -count=1 -cover ./internal/settings/yamlpatch/
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.341s	coverage: 87.9% of statements

$ timeout 200 go vet ./internal/settings/yamlpatch/
(no output)
```
No run printed `No space left on device`. The data volume had 68Gi free at the start (`df -h`: `/dev/disk3s5 3.6Ti 3.5Ti 68Gi 99%`).

### Claim 6: There is no new security surface.

Evidence: `git diff a35c589b94 fff52b3860 -- internal/ | grep -nE '^\+.*(exec\.|http\.|os\.Remove|token|secret|password|unsafe)'` gave no matches (grep-exit=1). The change only reads in-memory YAML nodes. It writes nothing new, adds no command execution, and adds no network calls. The two callers (internal/config/llm_save.go diffLLMNodes, internal/cli/glm.go legacyLLMAliasEdits) only build `Remove` edits, and the guard sits in the shared applyEdit, so both are covered.

### Claim 7: The audit plan requires codex, and the codex leg did not answer.

Evidence: `moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2109` returned `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, `"cross_model_required": true`, `"enforced_required": ["codex"]`. The one call to `mcp__moai__codex_audit` (project_root set to the tree, target baseBranch, delta_base 35911488199b0b19d9a367b3ab67118dfaebec67, mode native) was refused:
```
{"_server":{"commit":"646955327e","version":"v3.2.0-rc.42"},"error":"codex_audit: the serving moai build is stale — stale_server: the serving moai build (/Users/goos/go/bin/moai) differs from the installed build (/Users/goos/go/bin/moai) — reconnect the moai MCP server (restart the host session) so the installed build serves","stale_server":true}
```
I did not retry it, did not work around it, and called no other audit tool.

## Baseline-attribution

- Tree: HEAD fff52b38601995a69a7522ae5c9a9cb7dd62aeb4, read with `git rev-parse HEAD` at the start. The card diff base is a35c589b94, and the earlier-iteration head is 35911488199b0b19d9a367b3ab67118dfaebec67. Base and iteration-1 code were obtained with `git show <sha>:internal/settings/yamlpatch/yamlpatch.go` into the scratchpad.
- Toolchain: the Go toolchain on PATH. All `go test` runs compile from the tree (or from scratch copies of it), so the code judged is the code at the named SHAs. The only moai-binary measurement is `moai verify audit-plan`, which ran on the installed `/Users/goos/go/bin/moai`. Its commit was not compared with the tree's HEAD (see Gaps). It only reads the tree's config, and its output agrees with the gate settings the codex refusal implies.
- Probe copies: package directories under the scratchpad, using module `probe` and yaml.v3 v3.0.1 (the version in go.mod). Deleting them afterwards failed: the permission system denied `rm -rf` on that directory, so they remain at `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/10944804-d6fb-4180-9533-737a355e2e6e/scratchpad/probe`. It holds about four small copies of the package source plus logs, and no build cache. I did not try another way to delete it.

## Gaps

- G1: the codex leg was refused (`stale_server`, quoted verbatim above). The required codex gate is unmet, and there is no audit receipt for this head.
- Not run: the full `internal/cli` suite, the race detector, `golangci-lint`, other operating systems.
- The probes ran on a scratch module copy, not on the in-tree package path, so a test file in the tree that is not in `internal/settings/yamlpatch/` cannot have influenced them. The copy held every file in that directory, testdata included.
- The callers were not driven end to end into the refusal. No test makes `moai glm` hit an anchored `llm.yaml`. Coverage of that path rests on reading the code (the shared applyEdit) and on the cli and config suites passing.
- The installed `moai` build's commit was not compared with the tree's HEAD for the audit-plan read (VCI §2.2).
- In my first two `go test ... | tail` commands the exit status was not captured, because `${PIPESTATUS[0]}` is empty under zsh. Pass is read from the `ok` lines, and no `FAIL` line was printed.

## Residual-risk

- F6 (multi-document files lose every document after the first on any re-encode write) is real, pre-existing data loss in the same seam. A user's `llm.yaml` with `---` documents is at risk on any launch that writes.
- F3: merge-key files can silently keep a value the caller believed it removed.
- Unanchored removal that crosses an alias (path through `user: *u`) is a silent no-op rather than a refusal. That matches base behaviour, but a caller expecting removal gets none.
- The refusal is conservative: a removal is still refused when its aliases were removed earlier in the same edit list, or when an anchor has no alias at all. A user then edits by hand. This is acceptable under the card text.

## Iteration history

- Iteration 1: FAIL 0.80 at 35911488199b0b19d9a367b3ab67118dfaebec67 (`.moai/reports/t2109/sync-audit.md`). Blocking F1 (absent key refused); optional F2, F3, F5, F6. One codex_audit, receipt rcpt-011dcf46cb04b0827be82f37.
- Iteration 2 (this file): at fff52b38601995a69a7522ae5c9a9cb7dd62aeb4. F1 is resolved and re-verified both ways, and F2 is resolved, with mutants killed. No repair-born blocking finding: the delta since 35911488199b is the 5-line early return and the tests, and no defect was found in it. Overall FAIL only on the unmet required codex gate (G1).

## Operational Notes (unverified)

- Reconnect the moai MCP server: `stale_server` says the serving build 646955327e differs from the installed binary. Then run the single allowed codex audit for delta 35911488199b..fff52b3860 and record its receipt. If it raises no blocking finding, a confirming re-audit can be scoped to that receipt only.
- F7: consider dropping the force-added `.moai/reports/t2109/sync-audit.md` from the card before it lands, or keeping it deliberately. Either way, the factory evidence reader will see two verdict files for t2109 in that directory.
- F6 deserves its own card (multi-document preservation or refusal on the re-encode path).
