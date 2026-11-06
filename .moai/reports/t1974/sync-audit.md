auditor-model: claude-opus-5-5

# Sync audit — card t1974 (Class B, yamlpatch anchor refusal on the user-name set path)

verdict: PASS
audited_sha: 9ec025e54f578a78021206594ed0fae03db42903

- Base: `9b632c1eb3` (card branched directly from main). Range audited: `9b632c1eb3..9ec025e54f`.
- Audit mode: Claude-only, single model, operator instruction. No cross-model, codex or GLM tool was called. Iteration: 1 (first audit of this card).
- Overall score: **0.91** (Functionality 92, Security 95, Craft 85, Consistency 92; flat weighted 40/25/20/15 = 91.3). Must-pass (Functionality, Security): both PASS.
- Blocking findings: **0**. All findings below are optional; F2 and F3 sit outside the diff and are listed for the operator, not queued.

## Dimension scores

| Dimension | Score | Verdict | Evidence (verbatim, commands in § Evidence) |
|---|---|---|---|
| Functionality (40%) | 92 | PASS | E3: `ok github.com/modu-ai/moai-adk/internal/settings/yamlpatch 0.407s` / `ok github.com/modu-ai/moai-adk/internal/profile 1.271s`; E6: the four refusal subtests and three profile subtests fail with `err = <nil>` against the base `yamlpatch.go` |
| Security (25%) | 95 | PASS | No new input surface, no secrets, no injection surface; the change turns a silent cross-key mutation into an explicit refusal with the file byte-identical (E5 `changed=false` on every refused case) |
| Craft (20%) | 85 | PASS | Tests drive the real `PatchFile` and real `SyncToProjectConfig`; RED proven (E6); one surviving mutant (F1, E7) |
| Consistency (15%) | 92 | PASS | E4 `go vet` empty output; E8 `gofmt -l` empty output; guard style and comments match the file's existing `card tNNNN:` English comments |

## Findings (structured defect list)

- **F1** [Low] [optional] [in-diff, test] `internal/settings/yamlpatch/yamlpatch_anchor_t1974_test.go` — the anchor guard on an intermediate or root mapping is not pinned by any test. A mutant that checks the anchor only on the LAST parent (`i == len(e.Path)-1 && cur.Anchor != ""` in `lineSplice`, `last && cur.Anchor != ""` in `applyEdit`) passes the whole yamlpatch suite (E7). The shipped code does check every level (E5 `deep_anchor_intermediate` and `root_anchor` refused). Required fix (optional): add one subtest with a three-segment path whose anchor sits on the first segment, e.g. `a: &x\n  b:\n    c: Old\nz: *x` editing `a.b.c`.
- **F2** [Medium] [optional] [outside — `removeKey`, not in the diff] The Remove path is not guarded. Proven in scratch (E5): removing the key that carries an anchor definition writes a file that no longer parses (`reparse_err=yaml: unknown anchor 'n' referenced`, output `user:\n  nick: *n\n`), and removing a key inside an anchored mapping silently removes it from every alias (`remove_inside_anchored_map`, `changed=true`). Remove callers are `internal/config/llm_save.go:121` and `internal/cli/glm.go:905` (llm.yaml), not the card's set path (`sync.go` issues no Remove). The lane disclosed this in progress.md line 35. Recommended: a follow-up card applying the same refusal to `removeKey` / the `lineSplice` Remove branch.
- **F3** [Low] [optional] [outside — `internal/profile/sync.go`, `internal/cli/launcher.go`] `SyncToProjectConfig` returns at the first failing step and the user-name step runs first, so an anchored user.yaml now also blocks the language and statusline sync of the same call. Nothing is half-written in the project (the user-name step precedes every write). What the user sees: `moai profile setup` prints `Warning: failed to sync profile to project config: sync user name: yamlpatch: <path>/user.yaml: path "user.name": target carries anchor &n, so an edit would change every alias of it; edit the file by hand` after `preferences.yaml` was already saved (profile_setup.go:394-415); `moai update` records a `profile sync failed: ...` ledger warning (update.go:688-689); the launcher discards the error entirely (`_ = profile.SyncToProjectConfig(root, prefs)`, launcher.go:682), so on every launch the sync is silently skipped. The abort-on-first-error structure predates this card (alias inputs were already refused at base, E6 shows the alias tests passing at base); the card widened the input set that reaches it.
- **F4** [Info] [optional] [in-diff] An anchor that no alias references is refused too (test `anchored_scalar_without_any_alias_known`; E5 `unused_anchor_on_parent`), and a document-root anchor refuses every edit (E5 `root_anchor`). Deliberate and documented in progress.md line 22; accepted as conservative. No shipped YAML uses anchors (E2), so the false-positive cost falls only on hand-written files, which receive an actionable error.
- **F5** [Info] [optional] [outside, pre-existing] Upsert into a mapping that holds a merge key re-renders `<<: *b` as `!!merge <<: *b` (E5 `merge_key_upsert`). Valid YAML, cosmetic, untouched by this diff; the lane observed the same (progress.md line 17).
- **F6** [Info] [optional] Card wording: there is no `moai profile set` subcommand. `profileCmd` registers `list`, `current`, `delete`, `setup` (E9). The user-facing set path is `moai profile setup` (plus `moai update`, the launcher, and init), all through `SyncToProjectConfig` → `sync.go:53`.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

1. The diff touches exactly three files: `yamlpatch.go` (production) and two new test files.
2. Lane claim (1): `clearProjectUserName` does not exist in this tree — confirmed.
3. Lane claim (2): at base, the three anchored inputs (anchored scalar with sibling alias, anchored parent shared by alias, upsert into anchored parent) write successfully; alias value node and alias parent were already refused — confirmed by running the new tests against the base `yamlpatch.go`.
4. Lane claim (3): the guards in `setScalar`, `applyEdit`, and `lineSplice` refuse, and the refused file stays byte-identical — confirmed, including multi-edit atomicity (a refused second edit leaves the first unapplied on disk).
5. Lane claim (4): no YAML the repository ships uses anchors — confirmed, with a positive control.
6. Each guard is necessary: removing the `lineSplice` guard lets the fast path splice `name: &n Old` → `name: &n New` (base behaviour, E6); the `applyEdit` guard is the only one that catches the upsert case; the `setScalar` guard is the only one that catches an anchored target under an unanchored parent on the re-encode path.
7. Merge keys (`<<: *x`) do not trip the guard: an explicit key next to a merge key edits normally (E5 `merge_key_override_present`).
8. Alias nodes carry no `Anchor` value, and their `Kind` (`AliasNode`) already fails the `MappingNode`/`ScalarNode` checks, so the new `Anchor` checks do not change alias handling (E6 alias tests pass at base and at HEAD).
9. The greenfield seed `{}` and `appendSection` build fresh nodes with no anchor, so neither path is affected.
10. The `sync.go` wrapper still reads `sync user name: ...` (sync.go:55 unchanged; the profile test asserts the substring and passes).

### Evidence

E1 — `git -C <worktree> diff 9b632c1eb3..HEAD --stat`
```
 internal/profile/sync_anchor_t1974_test.go         | 42 ++++++++++++
 internal/settings/yamlpatch/yamlpatch.go           | 17 ++++-
 .../yamlpatch/yamlpatch_anchor_t1974_test.go       | 77 ++++++++++++++++++++++
 3 files changed, 135 insertions(+), 1 deletion(-)
```
`git -C <worktree> rev-parse HEAD` → `9ec025e54f578a78021206594ed0fae03db42903`.
Every changed line of `yamlpatch.go` was read: `lineSplice` adds `if cur.Anchor != "" { return nil, false, nil }` per traversed mapping and `|| val.Anchor != ""` on the target scalar; `applyEdit` adds the mapping refusal before `findKey` (so it also covers the upsert branch); `setScalar` adds the scalar refusal after the Kind check.

E1b — `grep -rn "clearProjectUserName" <worktree> --include=*.go` → no output (exit 0, no hit).

E2 — anchors in shipped and caller YAML:
- `grep -rnE -e "(:|-)[[:space:]]+&[A-Za-z0-9_-]+" -e "(:|-)[[:space:]]+\*[A-Za-z0-9_-]+[[:space:]]*$" -e "<<:" <worktree>/.moai/config <worktree>/internal/template/templates/.moai/config` → no output.
- Positive control, same anchor pattern over `internal/settings internal/config/llm_save.go internal/cli/glm.go internal/cli/init_workflow_flags.go internal/profile` → 12 hits, all in the two new t1974 test files (e.g. `yamlpatch_anchor_t1974_test.go:33: "anchored_scalar_with_sibling_alias": "user:\n  name: &n Old\n  nick: *n\n",`). The pattern fires on anchors, so the empty config result is a measured absence. No non-test caller source or fixture carries an anchor.
- Note: the template tree has no `sections/user.yaml` or `sections/language.yaml` (`No such file or directory`); `workflow.yaml` (407 lines) and `llm.yaml` (178 lines) exist and were covered by the empty grep.

E3 — `go test ./internal/settings/yamlpatch ./internal/profile -count=1`
```
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.407s
ok  	github.com/modu-ai/moai-adk/internal/profile	1.271s
```
`go test ./internal/settings/... -count=1`
```
ok  	github.com/modu-ai/moai-adk/internal/settings	1.084s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	0.296s
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.233s
```

E4 — `go vet ./internal/settings/yamlpatch ./internal/profile` → no output (exit 0).

E5 — scratch probe (scratchpad test file injected with `go test -overlay`, no tracked file modified): `go test -overlay <scratch>/overlay.json ./internal/settings/yamlpatch -run TestAuditScratch -count=1 -v`
```
CASE merge_key_override_present   err=<nil>  changed=true reparse_err=<nil>  out="base: &b\n  name: B\n  tz: UTC\nuser:\n  <<: *b\n  name: New\n"
CASE merge_key_upsert             err=<nil>  changed=true reparse_err=<nil>  out="base: &b\n  name: B\n  tz: UTC\nuser:\n  !!merge <<: *b\n  name: New\n"
CASE root_anchor                  err=... path "user.name": segment "user" parent carries anchor &r, so an edit would change every alias of it; edit the file by hand  changed=false
CASE unused_anchor_on_parent      err=... path "user.name": segment "name" parent carries anchor &u, ...  changed=false
CASE multi_edit_second_refused    err=... path "user.tz": target carries anchor &t, ...  changed=false
CASE multi_edit_upsert_then_anchor err=... path "user.tz": target carries anchor &t, ...  changed=false
CASE remove_anchor_definition     err=<nil>  changed=true reparse_err=yaml: unknown anchor 'n' referenced  out="user:\n  nick: *n\n"
CASE remove_inside_anchored_map   err=<nil>  changed=true reparse_err=<nil>  out="user: &u\n  name: Old\nteam: *u\n"
CASE deep_anchor_intermediate     err=... path "a.b.c": segment "b" parent carries anchor &x, ...  changed=false
--- PASS: TestAuditScratch (0.03s)
```
(Error prefixes `yamlpatch: <tempdir>/x.yaml:` elided above; the full lines are in the scratch run.)

E6 — RED against base: `git show 9b632c1eb3:internal/settings/yamlpatch/yamlpatch.go` exported to the scratchpad and swapped in via `-overlay`.
`go test -overlay <scratch>/overlay_base.json ./internal/settings/yamlpatch -run 'TestPatchFileRefusesAnchoredNodesOnThePath|TestPatchFileKeepsRefusingAliasNodes|TestPatchFileEditsBesideAnAnchorNotOnThePath' -count=1`
```
--- FAIL: TestPatchFileRefusesAnchoredNodesOnThePath (0.02s)
    --- FAIL: .../anchored_scalar_without_any_alias_known   err = <nil>, want a refusal naming the anchor
    --- FAIL: .../anchored_scalar_with_sibling_alias        err = <nil>, want a refusal naming the anchor
    --- FAIL: .../anchored_parent_shared_by_alias           err = <nil>, want a refusal naming the anchor
    --- FAIL: .../upsert_into_an_anchored_parent            err = <nil>, want a refusal naming the anchor
FAIL	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.290s
```
(`TestPatchFileKeepsRefusingAliasNodes` and the control test do not appear as failures: they pass at base, matching lane claim (2).)
`go test -overlay <scratch>/overlay_base.json ./internal/profile -run TestSyncToProjectConfigUserNameRefusesAnchorsAndAliases -count=1`
```
--- FAIL: .../anchored_user_shared_by_alias        sync_anchor_t1974_test.go:31: err = <nil>, want a refusal from the user-name sync
--- FAIL: .../name_absent_in_anchored_user         sync_anchor_t1974_test.go:31: err = <nil>, want a refusal from the user-name sync
--- FAIL: .../anchored_name_with_sibling_alias     sync_anchor_t1974_test.go:31: err = <nil>, want a refusal from the user-name sync
FAIL	github.com/modu-ai/moai-adk/internal/profile	0.423s
```

E7 — surviving mutant (F1): HEAD `yamlpatch.go` copied to the scratchpad, both mapping guards narrowed to the last parent, swapped in via `-overlay`: `go test -overlay <scratch>/overlay_mutant.json ./internal/settings/yamlpatch -count=1` → `ok  github.com/modu-ai/moai-adk/internal/settings/yamlpatch 0.391s`.
Mutants reasoned from the diff and E6, not run: guard only in `applyEdit`/`setScalar` → `anchored_scalar_with_sibling_alias` fails, because the fast path splices as at base (the lane reports observing exactly this); guard only in `lineSplice` → the re-encode fallback succeeds as at base, so the refusal subtests fail; any-anchor-anywhere refusal → `TestPatchFileEditsBesideAnAnchorNotOnThePath` fails; mutate-then-refuse → the byte-identity assertion (`out != in`) fails.

E8 — `gofmt -l` on the three changed files → no output.

E9 — `grep -rn -e "profileCmd.AddCommand" -e "Use: *\"profile" <worktree>/internal/cli` → `profileListCmd`, `profileCurrentCmd`, `profileDeleteCmd` (profile.go:53-55), `profileSetupCmd` (profile_setup.go:218). No `set` subcommand.

E10 — callers of `SyncToProjectConfig` (grep): `internal/cli/profile_setup.go:414` (warning printed), `internal/cli/update.go:688` (ledger warning), `internal/cli/launcher.go:682` (`_ =`, error discarded), `internal/cli/init_profile_sync.go:26` (returned).

### Baseline-attribution

- Tree: worktree `/Users/goos/.moai/worktrees/projects/moai-adk-go/t1974`, HEAD `9ec025e54f578a78021206594ed0fae03db42903` (E1), clean at session start per git status; this audit modified no tracked file. The two `-overlay` runs replaced `yamlpatch.go` in the build only, from `git show 9b632c1eb3:...` (base) and from a scratch copy of HEAD (mutant).
- Every measurement above was produced in this session by `go test` / `go vet` / `gofmt` / `grep` from the Go toolchain on this machine, compiling the tree's sources directly (no installed `moai` build judged anything, so §2.2 tool provenance is not in play).

### Gaps

- Cross-model and Codex review: not run, by explicit operator instruction for this chain (Claude-only). The audit-plan verb (`moai verify audit-plan`) was not consulted, for the same reason. If a receipt guard asks for a codex receipt when this audit stops, it will not be met, deliberately.
- No end-to-end run of `moai profile setup` / `moai update` / the launcher with an anchored user.yaml. F3's description of user-visible output comes from reading the callers (E10) and the refusal text observed in E5, not from a CLI run.
- Remove-path behaviour (F2) and the merge-key re-render (F5) are measured in scratch only; no caller-level reproduction through `llm_save.go` / `glm.go`.
- The "resume stops on an alias node" symptom named in the card was not reproduced by the lane or by this audit; no resume code path was examined.
- The full repository suite and other packages (`internal/cli`, `internal/config`, `internal/web`) were not run, by instruction (step 3 limits). The `internal/settings/...` suite, which holds the web console seam callers, passed (E3).
- Coverage was not measured (`-cover` not in the allowed command set).
- No command was refused by a guard during this audit.

### Residual-risk

- A user with a hand-written config file using anchors in `workflow.yaml` or `llm.yaml` now gets an error from every seam write that crosses an anchored node (web console, `init` workflow flags, GLM setup), where before they got a silent alias change. This is the intended trade, but it is a visible behaviour change on the shared primitive beyond the card's user.yaml scope, and only the user-name path is covered by a caller-level test.
- Through the launcher (F3) the refusal is silent and recurs on every launch; a user may never learn why language preferences stop syncing into that project.
- Removal (F2) still corrupts or silently alters anchored files; it remains reachable through llm.yaml writes.
- A key node that itself carries an anchor (`&k name: Old`) is not checked; considered unreachable in practice and not probed.

## Judgement on the scope decision

- **Central guard vs a narrow pre-check in sync.go:** accepted. The defect lives in the primitive (it mutates a node that aliases share), every seam caller had the same silent-mutation hazard, and no shipped YAML uses anchors (E2). A sync.go-only pre-check would have left the web console and llm.yaml writes exposed to the identical defect. Cost: F4 false positives on unaliased anchors, which are explicit, actionable errors.
- **Unguarded Remove path:** acceptable for PASS on this card. It is outside the diff, unreachable from the set path the card names, and disclosed by the lane. It is a real defect (F2 proves an unparseable write), so it belongs in a follow-up card rather than a debt line on this one.

## Recommendations

1. (optional, this card or later) Add the three-segment intermediate-anchor subtest that kills the F1 mutant.
2. (follow-up card, operator decision) Guard `removeKey` and the `lineSplice` Remove branch with the same anchor refusal; add a test that removes an anchor-defining key.
3. (follow-up, optional) Consider surfacing the launcher's discarded `SyncToProjectConfig` error, or running the user-name step so its refusal does not block the language and statusline steps.
