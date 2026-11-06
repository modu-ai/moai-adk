auditor-model: claude-opus-5-5

# t2109 sync-audit — YAML key removal and anchors

## Evaluation Report

SPEC: none (Class B card t2109; acceptance basis = card text)
Audited range: a35c589b94 .. 35911488199b0b19d9a367b3ab67118dfaebec67 (one commit, 2 files: internal/settings/yamlpatch/yamlpatch.go +56/-1, internal/settings/yamlpatch/yamlpatch_remove_anchor_t2109_test.go +87)
Overall Verdict: FAIL (score 0.80)
Iteration: 1 (first sync audit of this card; there is no previous audited SHA, so no finding is repair-born)

verdict: FAIL
audited_sha: 35911488199b0b19d9a367b3ab67118dfaebec67

The fix does what the card asks for every anchored case I probed: a key that defines an anchor, a key inside an anchored mapping (direct parent, a higher ancestor, flow style), an anchor on the key node, and an anchor deeper inside the removed value are all refused with the file byte-identical, including when a fine edit comes first in the same PatchFile call. The FAIL rests on one blocking defect the change itself introduces: removing an ABSENT key under an anchored mapping is now refused. Before this commit that removal was a no-op and the other edits in the batch were written. The launcher's own edit list can hit this case. The repair is three lines and I verified it in a scratch copy (F1). Codex (one call, receipt below) found the same defect independently.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 70/100 | FAIL (must-pass) | Card cases are refused and the file stays unchanged (P5, P8, P12, P6, P6b; the card's 4 RED cases fail on base and pass on HEAD). Regression F1: P1b/P3 return a refusal on HEAD, while base wrote the correct result. |
| Security (25%) | 92/100 | PASS | Diff grep for exec/secret/token/perm patterns: `grep exit=1` (no match). The change only refuses writes and fails closed before atomicWriteMode. It opens no new input surface. |
| Craft (20%) | 75/100 | PASS (weak) | `gofmt exit=0`, `vet exit=0`, package tests ok. Two mutants survive the suite (F2). No control covers an absent key under an anchored parent (the gap behind F1). |
| Consistency (15%) | 88/100 | PASS | The refusal wording and placement match t1974's setScalar/applyEdit guards ("... carries anchor &x ...; edit the file by hand"). The doc comment contradicts the code's behaviour for absent keys (F1). |

Must-pass firewall: Functionality is below its threshold because of blocking finding F1, so the overall verdict is FAIL whatever the other scores are.

### Findings (structured defect-list)

- F1 [Medium] [blocking] [in-diff, reproduced] internal/settings/yamlpatch/yamlpatch.go:454 — `refuseAnchoredRemoval` checks `cur.Anchor` BEFORE `findKey`. Removing a key that does not exist (or whose intermediate segment does not exist) under any anchored mapping on the path is therefore refused, which contradicts the function's own doc comment ("nil when the removal is safe (including when the key is absent ...)") and the progress.md claim ("An absent key ... return nil as before"). lineSplice skips absent removals, so the defect shows only when another edit in the batch forces the re-encode path (an upsert or a non-scalar target). It then cancels the whole batch, where base wrote the correct result. Launcher-shaped reproduction P3: `llm.glm.models: &m` plus an upsert of `llm.team_mode` plus the three legacy alias removals (`legacyLLMAliasEdits`, glm.go:905). HEAD refuses with `path "llm.glm.models.opus": segment "opus" parent carries anchor &m`, while base wrote `team_mode: glm` with the anchor and alias intact. This also falsifies progress.md's claim that "an anchored llm.glm.models already made the t1974 set-refusal fire first": that holds only when a set edit targets llm.glm.models, not for llm.team_mode. Required fix: at the top of refuseAnchoredRemoval, `if lookupNode(root, path) == nil { return nil }` (or check the anchor only after the key is found). Then add a control test: an absent key under an anchored parent, together with an upsert in the same PatchFile call, must succeed and leave the anchor and alias intact. I verified this fix in scratch: the full yamlpatch package plus probes pass, and the P1b/P3 output is byte-identical to base's.
- F2 [Low] [optional] [in-diff, test file] yamlpatch_remove_anchor_t2109_test.go — two wrong implementations still pass the whole package suite: mutant A ("ignore the key node": the key-node `firstAnchor` check disabled) and mutant B ("refuse on a parent anchor only at the direct parent": `cur.Anchor != "" && i == len(path)-1`). Mutant C ("value's own anchor only, no deep scan") and mutant D ("no parent anchor check") are killed. Required fix (recommended together with F1): add the cases `"&k name: Old\nx: *k\n"` remove `name` and `"a: &g\n  b:\n    c: 1\n    d: 2\nz: *g\n"` remove `a.b.c`. Both are refused correctly today (P12, P5); the cases pin that behaviour.
- F3 [Low] [optional] [in-diff] merge keys — removing an override key in a mapping that merges an anchored one (`user: {<<: *b, x: 3}`, remove `user.x`) is allowed, and `user.x` then reads as the inherited 1, not as absent (P4). The anchored base and the other aliases are unchanged, so this is not the corruption the card describes. It is a semantic surprise that the code does not mention. Optional: refuse the removal when the parent mapping holds a `<<` key, or document the behaviour.
- F4 [Info] [optional] [in-diff] conservative over-refusal — removing a value that carries an anchor no alias uses (P13), or an anchor that a later node redefines (P14), is refused. This is safe and simple, and matches the "edit the file by hand" stance. No action needed.
- F5 [Info] [optional] [touched-file: callers] internal/cli/launcher.go:305/437, internal/cli/glm.go:922 — the refusal reaches the user as `Warning: failed to persist team mode: ... carries anchor &m ...; edit the file by hand` on stderr (glm path), or as the returned warning string `Warning: failed to disable team mode: ...` (cc path), and launch continues. For an llm.yaml whose `llm.glm.models` (or an ancestor) is anchored AND still holds a legacy opus/sonnet/haiku key, every launch that needs a write warns until the user edits the file. I judge this degradation acceptable (fail-closed, file intact, message actionable) and not a card defect, apart from the absent-key case F1 already covers. llm_save.go (ConfigManager.saveLLM) returns the error from Save().
- F6 [High, data loss] [optional for this card] [outside, pre-existing] internal/settings/yamlpatch/yamlpatch.go:126-165 (re-encode path) — `yaml.Unmarshal` reads only the first document, so any re-encoded write deletes every later `---` document: P7 `"a: 1\nb: 2\n---\nc: 3\n"` remove `a` → `"b: 2\n"`, identical on base. This predates the card and lies outside its diff, so it does not affect the verdict. It goes to the operator as a separate card candidate. Related pre-existing outside behaviour: the re-encode renders a merge key as `!!merge <<: *b` (P4, same on base).

### Recommendations

- Apply the F1 guard and its control test, add the F2 pinning cases in the same commit, then request a delta re-audit scoped to F1/F2.
- File F6 (multi-document loss on the re-encode path) as its own card. It is not this card's scope.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

1. The card's four anchored-removal cases fail on the pre-fix code (RED) and pass on HEAD, with the file unchanged.
2. Anchored cases beyond the card's tests are refused with the file unchanged: grandparent anchor, flow-style anchored mapping, key-node anchor, a mixed batch where a fine edit precedes the refused removal.
3. The change regresses a safe case: an absent-key removal under an anchored mapping, combined with any re-encode-forcing edit (F1).
4. The test file leaves two plausible wrong implementations alive (F2).
5. Package-scoped tests, gofmt and vet are clean on HEAD.
6. The tree is left unchanged by this audit.

### Evidence

Card commit scope (`git diff a35c589b94 35911488199b0b19d9a367b3ab67118dfaebec67 --stat`):
```
 internal/settings/yamlpatch/yamlpatch.go           | 56 +++++++++++++-
 .../yamlpatch_remove_anchor_t2109_test.go          | 87 ++++++++++++++++++++++
 2 files changed, 142 insertions(+), 1 deletion(-)
```

Required runs in /Users/goos/.moai/worktrees/projects/moai-adk-go/t2109 (HEAD 35911488199b0b19d9a367b3ab67118dfaebec67):
```
$ timeout 300 go test -count=1 ./internal/settings/...
ok  	github.com/modu-ai/moai-adk/internal/settings	6.135s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	0.845s
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	1.593s

$ timeout 400 go test -count=1 ./internal/config/
ok  	github.com/modu-ai/moai-adk/internal/config	86.659s

$ timeout 560 go test -count=1 ./internal/cli/ -run 'GLM|Glm|LLM|Llm'
ok  	github.com/modu-ai/moai-adk/internal/cli	196.345s
[exited with code 0]

$ gofmt -l internal/settings/yamlpatch; echo "gofmt exit=$?"; go vet ./internal/settings/yamlpatch/; echo "vet exit=$?"
gofmt exit=0
vet exit=0
```

RED on base. A scratch copy holds base `yamlpatch.go` (from `git show a35c589b94:internal/settings/yamlpatch/yamlpatch.go`) plus the card's test file:
```
    yamlpatch_remove_anchor_t2109_test.go:46: err = <nil>, want a refusal naming the anchor   (x4)
--- FAIL: TestPatchFileRefusesRemovingAnchoredKeys (0.02s)
    --- FAIL: .../key_whose_mapping_value_defines_an_anchor
    --- FAIL: .../key_whose_scalar_value_defines_an_anchor
    --- FAIL: .../key_inside_an_anchored_mapping_shared_by_an_alias
    --- FAIL: .../removed_subtree_holds_an_anchor_definition
--- PASS: TestPatchFileStillRemovesUnanchoredKeys (0.24s)   (plain_key, an_anchor_elsewhere_in_the_file_is_unrelated, absent_key)
```
The failures come from the stated cause: base returns `err = <nil>`, because base `applyEdit` calls `removeKey` without any anchor check. The controls pass on base and HEAD alike, so they are real controls for "removal is not a blanket ban".

Behaviour probes (scratch test `zz_probe_test.go` on a `git archive` copy of HEAD, then on the base copy). Format: ERR | parses after | file unchanged | output:
```
HEAD P1b absent key under anchored parent + upsert:
  ERR=... path "user.missing": segment "missing" parent carries anchor &u, so removing a key would change every alias of it; edit the file by hand | PARSE_AFTER=<nil> | UNCHANGED=true
BASE P1b: ERR=<nil> | UNCHANGED=false | OUT="user: &u\n  name: Old\nteam: *u\nmode: x\nnewkey: v\n"

HEAD P3 launcher-like (llm.team_mode upsert + 3 legacy removals, models: &m):
  ERR=... path "llm.glm.models.opus": segment "opus" parent carries anchor &m, so removing a key would change every alias of it; edit the file by hand | UNCHANGED=true
BASE P3: ERR=<nil> | OUT="llm:\n  glm:\n    models: &m\n      high: a\n      low: b\n  copy: *m\n  team_mode: glm\n"

HEAD P1 absent key alone (lineSplice path): ERR=<nil> | UNCHANGED=true
HEAD P2 absent deep path under anchored ancestor alone: ERR=<nil> | UNCHANGED=true
HEAD P5 grandparent anchor a.b.c: ERR=... segment "b" parent carries anchor &g ... | UNCHANGED=true   (BASE: removed c, alias z changed)
HEAD P6 set ok then refused remove: ERR=... carries anchor &n ... | UNCHANGED=true
HEAD P6b remove ok then refused remove: ERR=... carries anchor &n ... | UNCHANGED=true
HEAD P8 flow anchored mapping: ERR=... parent carries anchor &u ... | UNCHANGED=true
HEAD P9 alias as ancestor (user: *b, remove user.x): ERR=<nil> | UNCHANGED=true
HEAD P10 alias as removed value: ERR=<nil> | PARSE_AFTER=<nil> | OUT="base: &b 1\nk: 2\n"
HEAD P12 key-node anchor: ERR=... the key carries anchor &k ... | UNCHANGED=true
HEAD P13 unused anchor in value: refused | UNCHANGED=true
HEAD P14 redefined anchor: refused | UNCHANGED=true
HEAD P15 unanchored sibling of anchored value: ERR=<nil> | OUT="user:\n  nick: &n k\nteam: *n\n"
HEAD P4 merge-key override removal: ERR=<nil> | OUT="base: &b\n  x: 1\nuser:\n  !!merge <<: *b\n"   (BASE identical)
HEAD P7 multi-document: ERR=<nil> | OUT="b: 2\n"   (BASE identical: second document lost)
```

Mutant probes (scratch copy of HEAD, `go test -count=1 ./internal/settings/yamlpatch/`):
```
A  key-node check disabled                         -> ok   (survives)
B  parent anchor checked at last segment only      -> ok   (survives)
C  value.Anchor only, no firstAnchor deep scan     -> --- FAIL .../removed_subtree_holds_an_anchor_definition  (killed)
D  parent anchor check disabled                    -> --- FAIL .../key_inside_an_anchored_mapping_shared_by_an_alias  (killed)
```

F1 fix candidate (scratch only, `if lookupNode(root, path) == nil { return nil }` at the top of refuseAnchoredRemoval):
```
--- PASS: TestZZProbe/P1b_absent_key_under_anchored_parent_with_upsert
--- PASS: TestZZProbe/P3_launcher_like_teammode_upsert_models_anchored
P1b OUT="user: &u\n  name: Old\nteam: *u\nmode: x\nnewkey: v\n"   (= base)
P3  OUT="llm:\n  glm:\n    models: &m\n      high: a\n      low: b\n  copy: *m\n  team_mode: glm\n"   (= base)
P5/P8/P12 still refused, UNCHANGED=true
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.950s
```

Caller surface (read at HEAD): launcher.go:305 `if err := persistTeamMode(root, "glm"); err != nil { fmt.Fprintf(os.Stderr, "Warning: failed to persist team mode: %v\n", err) }`; launcher.go:437 `return fmt.Sprintf("Warning: failed to disable team mode: %v", err)`; glm.go:972 always appends `legacyLLMAliasEdits()` to a real write; llm_save.go:41 returns PatchFileWithMode's error.

Security probe: `git show 3591148819... | grep -nEi 'exec\.|os/exec|password|token|secret|api_?key|unsafe|0o7|0777'; echo "grep exit=$?"` → `grep exit=1`.

Cross-model (one call, operator allowance d-20261106T055816Z-449d): `mcp__moai__codex_audit` mode native, target baseBranch, delta_base a35c589b94, project_root = this tree. Result -> verdict=fail; one finding [P2] at yamlpatch.go:454: "없는 키의 삭제는 앵커 검사 전에 무시하라". It describes the same absent-key-under-anchor upsert cancellation as F1. The review was static (codex ran no commands). review_base "delta_base (merge base a35c589b9476b831dc445cff095bc903b1825297)", resolved_model gpt-6.1-sol, audit_receipt rcpt-011dcf46cb04b0827be82f37.

Tree state after audit: `git status --short` → (empty); `git rev-parse HEAD` → 35911488199b0b19d9a367b3ab67118dfaebec67.

### Baseline-attribution

- Every test, gofmt and vet run above was executed in this audit, against tree HEAD 35911488199b0b19d9a367b3ab67118dfaebec67, with the Go toolchain building the tree's own source (no installed moai binary was used to judge the code).
- Base behaviour was measured on a35c589b94's `yamlpatch.go` (git show), placed in a scratch copy beside HEAD's test files. The probes ran in scratch copies under /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/10944804-d6fb-4180-9533-737a355e2e6e/scratchpad/ (probe = HEAD, base = base code, mut = mutants and fix candidate). The scratch copies hold only `go.mod`, `go.sum` and `internal/settings/yamlpatch` from `git archive 35911488199b...`.
- The codex leg ran on moai mcp-server build 646955327e (not built from this tree). It produced a review only, which I use as a second opinion and not as a measurement.

### Gaps

- I measured F1 at the yamlpatch layer with a launcher-shaped edit list (P3), not by driving `saveLLMSection`/`persistTeamMode` end to end with an anchored llm.yaml. Whether a real persisted section produces an upsert edit depends on which llm keys the user's file lacks. That dependency is inferred from the code, not measured.
- I did not run: the full `internal/cli` suite, the race detector, golangci-lint, coverage measurement for the package, or other OSes.
- No command was refused during this audit, except one compound shell line that the worktree guard refused while I was setting up scratch probes. I re-ran it as separate plain commands and Write calls, and its result is not used as evidence.
- I did not verify whether the receipt guard accepts the codex receipt. The verdict is FAIL, so no PASS needs corroboration.

### Residual-risk

- The refusal gives an anchored llm.yaml a permanent per-launch warning when legacy alias keys sit under an anchored mapping (F5). It is acceptable but noisy, and the scenario is uncommon.
- Merge-key semantics (F3) and multi-document loss (F6) remain. F6 is pre-existing data loss on any re-encoded write to a multi-document file.
- Scratch-only material (probe test file, mutant copies) stays in the session scratchpad and is not exported. The deciding outputs are copied into this file.

## Operational Notes (unverified)

- The F1 repair is a single early return. A delta re-audit scoped to F1+F2 should be enough.
- F6 deserves its own card. A guard that refuses (or line-splices) when the input holds more than one document would close it. I have not designed or tested that guard.
- Receipt for the one allowed cross-model call: rcpt-011dcf46cb04b0827be82f37.
