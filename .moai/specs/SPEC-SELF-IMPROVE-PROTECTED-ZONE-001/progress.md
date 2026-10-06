# SPEC-SELF-IMPROVE-PROTECTED-ZONE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-04
tier: M
artifacts: spec.md (v0.3.0), plan.md, acceptance.md, progress.md, decision-index.md, evidence/ (8 files)
requirements: 16 (ceiling 16) — acceptance criteria: 13 (ceiling 16)
probe_cases: 67 (re-recorded after plan-audit iteration 1; live and replay judge outputs identical at base)
base_tree_sha: e497f693608ac7ea45a08b06304dc585e927ff49 (code this SPEC reads or edits unchanged at HEAD 41127e8caded5c54507414fc6dd88fbf28d9517b)
measurement_provenance: binary built with `go build ./cmd/moai` from the tree at the SHA above, in the authoring run (and from HEAD 41127e8ca for the 2026-10-05 re-run); the installed `moai` was not used as the judge
open_decisions: none held — Q1-Q7 in decision-index.md carry "default accepted — operator approval 2026-10-05 relayed by the leader"
known_gaps: spec.md §F G1-G8

### Delta round — 2026-10-05 (audit status, stated plainly)

- Plan-audit iteration 1: FAIL 0.74. Plan-audit iteration 2 (Tier M ceiling): FAIL 0.78 against the 0.80 threshold. No cross-model receipt exists (`receipts=none` on both verdicts).
- The operator approved proceeding to run as PASS-with-debt on 2026-10-05 (relayed by the leader). This is not a plan-audit PASS and does not stand in for one.
- The three blocking defects of iteration 2 were repaired in this delta **without a re-audit**: D20 (the legacy local-instruction basename moved from the shipped manifest to the dogfood overlay), D21 (the seven-category rule binds any file at the shipped path; probe fixtures MS4/MS7/MS8 carry all seven; the recorded probe and judge outputs came out byte-identical, so no evidence file was rewritten), D23 (the real scope of the human route is stated: a baseline-matched Write/Edit denial keeps its legacy reason and carries no routing field). The minor defects D22, D24, D25 and D26 and the bookkeeping item D27 were fixed where touched. D15 (scope breadth) and D28 remain open by design.
- The sync-phase audit under the new audit rules will re-read these repairs; until then they are unverified by any independent auditor.

## §E.2 Run-phase Evidence

Run executed 2026-10-05 by lane-5 (factory card t1510, leader-dispatched; lane-direct landing after the prior session stalled mid-M2 — its uncommitted WIP was adopted after a full read and a live-owner check: `lsof -a -c claude -d cwd` found no session inside the tree). Branch `WT-self-improve-protected-zone`; base of this run's own work = M1 commit `8f6a4475a` (plan-phase artifacts + manifest/loader/template already landed by the prior session, SPEC 0.3.0, Kickoff recorded).

### M2 — guard behaviour for file tools (commit `9ec0d557b`)

- RED (adopted WIP state, observed before any implementation): `go test -run '^TestProtectedZone$' ./internal/hook/` exit **1**, `[build failed]` — `undefined: SentinelHarnessFrozenProtectedZone / zoneDenyReason / zoneReasonMax / zoneBaselineCategory`; `go build ./...` exit 0.
- GREEN: all eight M2 subtests pass. Swept rows per subtest (each prints its own `swept=` line): Normalization 33, FileTools 24 (incl. the three symlink rows), ManifestStates 15, NonRegression 19, DenyReason 9, NoManifestReadForOthers 18 (0 reads for non-identity/non-file tools; exactly 1 for an identity Write — the cost seam is alive), AuditRow 4/4 nested (Deny 1 row exactly, healthy allow appends nothing, Absent records once, Invalid records, append failure leaves the decision unchanged), BaselineCovered 21/21 by name.
- Live judge flips with a tree-built binary (`go build -o /tmp/t1510-moai ./cmd/moai`, judge `judge-probe.sh <bin>`): R 21/21, P/S 7/7, C/N 13/13, LP 2/2 — all `fail=0`; MS/E `fail=1` (E3 red as acceptance.md §AC-SIPZ-004 predicts at end of M2).
- Defect found by the judge and repaired in the same milestone: P4 (letter-case variant `.CLAUDE/Hooks/…`) was denied with the new sentinel instead of the legacy `HARNESS_FROZEN_HOOK_VIOLATION` — the baseline step ran on the display-case form only; repaired with an exact-case-then-folded retry so canonical inputs keep their byte-for-byte reason and case variants earn the legacy sentinel.
- Test-authoring repairs to the adopted WIP: the three symlink rows now increment the shared swept counter (the floor of 24 was unreachable without them), two acceptance.md §D edge rows added to ManifestStates (zero-byte overlay valid, duplicate entries valid) bringing it to its stated floor of 15, and `zoneTestHandler` now `t.Chdir(root)` — the existing file-access check resolves relative paths against the process cwd and the judge probe runs from inside the fixture root, so chdir gives the in-process handler the same resolution basis instead of a second one.

### M3 — shell rule, liveness, evidence (this commit)

- `ShellMutation` subtest: the thirteen REQ-SIPZ-007 forms each denied with `category=probe_zone` + routing fields (swept 17: 13 forms + 2 read-only controls + 1 non-identity rm + 1 assignment-prefix/comment row). Implementation: `internal/hook/protected_zone_shell.go` reuses `splitShellSegments`, tokenizes with quote awareness, skips leading `VAR=value` (REQ-SIPZ-013), pairs verb with candidates; the splitter's comment exclusion handles trailing comments. Wired after every existing shell guard so an earlier deny is preserved.
- `Liveness` subtest (REQ-SIPZ-015), all four conditions named and passing: MatcherGroup (both settings files carry `"matcher": "Write|Edit|Bash"`), ManifestParse (dogfood load OK + template shipped file parses), DeadEntries (`resolved=34 skipped=10`, resolved equals the paths entries the files declare, zero dead), HandlerDenies (end-to-end deny through `Handle`).
- Live judge, final tree binary (`/tmp/t1510-moai-final`, built after every post-test repair): **`JUDGE swept=67 expected=67 fail=0`, exit 0**.
- Latency A/B (AC-SIPZ-007): A = binary built from `8f6a4475a` (pre-guard, extracted via `git archive`, built with `go build -C`), B = final-tree binary, interleaved 21 pairs per class, same session. A/A control: 43/12/37 ms median diffs (bound 100). A/B: 54/−25/−24 ms (bound 250). **verdict=PASS all three classes** (`non-identity-write`, `identity-outside`, `identity-inside`).
- Quality gates: `gofmt -l` clean; `go vet` exit 0 (hook, config, template, harness); `golangci-lint run` (v2.1.6 — the CI version) **0 issues** on hook+config+harness (repairs: checked `Close` error in `zoneAppendAudit`, de-Morganed `isZoneAssignment`); `go test -race -timeout 30m ./internal/hook/...` — 0 data races.
- Full-package sweep findings fixed here (both invisible to `-run`-scoped testing, per the swept-set hazard): `TestShippedConfigKeysHaveReaders` failed because M1 registered the section without its 13 keys in `internal/config/testdata/shipped_key_inventory.yaml` (REQ-CKH-008 anti-rot) — registered as `class: W, evidence: reader`; `TestHMPSourceGuardGoLiterals` failed on the new Bash literals — registered in `hmpGoLiteralExclusions` with the §C.6 design reason (the rule is deliberately Bash-only; PowerShell is a separate matcher).
- Known-red attribution (observed, out of this SPEC's scope): `TestStaleRunNoticeFactoryLegacyLabel` fails at the M1 base too (`go test` on the `git archive` of `8f6a4475a`) — pre-existing launcher-protocol debt (t1399 territory), untouched here. `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer` and `TestFactoryHookZeroTurnAndCapabilityTruth` fail only under `-race` (pass without it) — race-mode timing sensitivity, 0 data races reported.

### Mutant execution (acceptance.md §C — 23 rows)

Each mutant applied, its killing check run, the red observed, then reverted; tree verified clean by `git status` after the sweep. Two first attempts were discarded and redone honestly: the (t) attempt that merely broke the build (a compile failure is not the assertion's red — redone as a whole-reason cut, killed by `DenyReason`) and the (v) attempt that removed an overlay-only entry the 21-by-name list does not assert (redone against the real R21 entry, killed by name). Mutant (s)'s first form targeted a category-level key; MS8's `exclude:` is a top-level unknown key, so the mutant was redone at `KnownFields(false)`.

| # | Mutant | Killed by | Observed red (one line) |
|---|---|---|---|
| a | manifest ignored (load forced absent) | ManifestStates | MS4-class row `want deny` denied nothing (`exit 1`) |
| b | raw-prefix match, no normalization | FileTools | `Write absolute: decision="allow", want deny` |
| c | deny every identity Write | NonRegression | `identity write to .moai/specs/...: decision="deny", want allow` |
| d | new sentinel never emitted | FileTools | `--- FAIL: TestProtectedZone/FileTools` |
| e | invalid treated as absent | ManifestStates | `invalid shipped: decision="allow", want deny` |
| f | invalid denies every caller | ManifestStates | `non-identity caller with an invalid manifest: decision="deny", want allow` |
| g | manifest read before identity gate | NoManifestReadForOthers | read counter `!= 0` (`exit 1`) |
| h | env bypass honoured | ManifestStates | `environment variable: decision="allow", want deny` |
| i | overlay replaces instead of adds | ManifestStates | `overlay cannot narrow: decision="allow", want deny` |
| j | `route=human` field missing | DenyReason | reason `lost a routing field or the order` |
| k | read-only verbs matched | ShellMutation | `"cat zone_dir/a.log": decision="deny", want allow` |
| l | sweep counts nothing | Normalization floor | `swept 9 rows, want at least 33` |
| m | case folding only on case-insensitive hosts | linux CI leg | not executable on darwin — CI-owned (acceptance.md §AC-SIPZ-012) |
| n | failed audit append poisons decisions | AuditRow/AppendFailure | `allow turned into "deny" after a failed append` |
| o | matcher drops Bash | Liveness/MatcherGroup | tmpl matcher `no longer carries Write, Edit and Bash` |
| p | runtime_paths not protected | judge `-o '^R'` | 3 FAIL rows among R13–R16, `exit 1` |
| q | root never symlink-resolved | FileTools/symlinks | `root via symlink, path physical: decision="allow", want deny` |
| r | verb list truncated to three | ShellMutation | table rows red, `exit 1` |
| s | unknown top-level key ignored | ManifestStates | `overlay unknown key: reason … category=probe_docs … does not start with … manifest=invalid` |
| t | whole reason cut instead of the path | DenyReason | `long ascii path: reason "" lost a routing field or the order` |
| u | runtime-holding category exempt wholesale | Liveness/DeadEntries | `sweep resolved 13 paths entries, want 34` |
| v | baseline entry missing from manifest | BaselineCovered | `baseline entry harness ".claude/skills/moai/" … is not covered` (by name, R21) |
| w | baseline denial gains routing fields | NonRegression | `Write .claude/hooks/moai/x.sh: … want deny with exactly "HARNESS_FROZEN_HOOK_VIOLATION: …"` |

### Repair round 1 — merge-gate review findings (2026-10-05, post-sync)

The turn-end codex review gate — the independent review the card-review stage could not obtain — failed the card diff with 4×P1 and 2×P2, all real and all in the guard this SPEC added. Each finding was adopted as a test row first (RED observed verbatim: `rm -rf zone_dir: allow`, `echo x>zone_dir/a.log: allow`, `cd zone_dir && rm a.log: allow`, `dot-dot through link: allow`, `dot-prefixed dir entry: pattern="./docs/"`, `shipped/overlay second YAML document: valid`), then repaired:

- **P1 dir operand** — `rm -rf .claude/hooks` walked past every entry because the match never tried the directory spelling. `zoneShellCovered` now tries each form with and without a trailing slash against the baseline and the entries.
- **P1 redirections** — the first redirect target was the only one, an inline `x>file` was invisible, and a verb's segment dropped its own redirections. `zoneMutatingWords` now splits every word on `>` (all substrings are targets, a trailing `>` hands to the next word) and unions the targets with the verb's arguments.
- **P1 segment cwd** — `cd zone_dir && rm a.log` resolved `a.log` against the hook process cwd. `checkProtectedZoneShell` tracks a plain literal `cd` across segments (`zoneNextCwd`; substitution, glob, option and absolute arguments reset the tracking — the documented under-match) and prefixes the candidates of later segments.
- **P1 dots after a symlink** — measured, not assumed: `filepath.EvalSymlinks("deep/../secret.md")` with `deep -> zone_dir/sub` collapses the dots against the LEXICAL parent (go1.26), landing at the project root while the shell lands in `zone_dir`. `zoneResolve` now walks the components itself — each existing component symlink-resolved as encountered, a `..` pops the resolved prefix, the missing tail rejoins there. `resolveThroughExistingParent` (shared with `checkFileAccess`) is deliberately untouched: its lexical-`Dir` walk has the same shape of hole for `..`-after-symlink targets, recorded below as residual.
- **P2 manifest normalization** — an overlay entry `./docs/` loaded valid and protected nothing. `parseZoneEntry` now cleans the body lexically (a cleaned body naming the repository root is rejected like a bare `*`; the raw `..`-segment rejection is unchanged and runs first).
- **P2 second YAML document** — a valid manifest followed by `---` and malformed text loaded as if the tail did not exist. `ParseProtectedZone` now requires the first document to be the only one.

The shipped and template manifests gain `.moai/config/` under `gate_policy` (both copies, byte-identical): without the parent entry `mv .moai/config docs/config` removed the whole zone directory while every individual entry still matched. The liveness sweep counts move with it (resolved=35 of 35 declared paths entries). `.claude/` is deliberately NOT added wholesale — the identity's legitimate surface (`.claude/agents/harness/`, `.claude/skills/hns-*`) lives there and the controls N4/N5 must stay allowed.

GREEN on the whole family after the repairs: `TestProtectedZone` 10/10 subtests (hook, FileTools swept 26, ShellMutation 24) + config Load/Overlay/Validation (Validation swept 46), `go vet` exit 0, `gofmt` clean. Test rows added: 6 ShellMutation rows + 1 control, 2 symlink rows (file and shell form of the same finding), 2 shipped/overlay multi-document rows, 2 normalization assertions, and the multi-document table row.

Residual from the round, recorded: `resolveThroughExistingParent` (shared with `checkFileAccess`) still resolves a `..`-after-symlink target lexically — the same shape of hole this round fixed for the zone guard, living in code this SPEC's plan deliberately did not touch (the existing outside-project deny). Left for a follow-up card with the leader's routing.

### Repair round 2 — merge-gate re-verdict findings (2026-10-05, post-push)

The gate's re-verdict on the round-1 push failed again: 5×P1 + 1×P2, all in the shell rule, all adopted as rows first (RED observed: `mv .moai moved: allow`, `true & rm zone_dir/a.log: allow`, `cd . > zone_dir/a.log: allow`, `sed -e s/a/b/ -i '' …: allow`, `git -c core.quotePath=false checkout -- …: allow`, and the quoted-`>` control falsely denied):

- **P1 parent-directory move** — the coverage gained an ancestor pass: a candidate that is a parent directory of protected entries (or the project root itself) is denied with the contained entry's category, since moving or removing it removes them wholesale. `.claude/` is still deliberately excluded from wholesale protection (the identity's legitimate surface lives there).
- **P1 single `&`** — the shared splitter does not segment on an async `&`; `zoneAsyncGroups` now splits each segment's words at an unquoted `&` (including one embedded in a word, which the shell reads as a separator), and the working-directory tracking resets at every async boundary because a cd inside an async group runs in a subshell.
- **P1 cd's own redirection** — the cd branch judges its segment's redirection targets against the directory the shell evaluates them in (before the cd takes effect) instead of skipping the segment.
- **P1 sed option order** — in-place is signalled by any argument carrying `-i`, wherever it sits, not only the first option slot.
- **P1 git global options** — the subcommand is looked up past git's global options (`-c k=v`, `--git-dir=…`) instead of being assumed the second token.
- **P2 quoted `>`** — the tokenizer now records per word whether quoting contributed; only unquoted words carry redirection operators, while a quoted word after a bare `>` is still the target. `echo '>zone_dir/a.log'` returns to allow.

GREEN on the whole family after the repairs: `TestProtectedZone` hook+config 0 FAIL (ShellMutation swept 30), `golangci-lint` 0 issues, live judge `JUDGE swept=67 expected=67 fail=0` on a binary built from the repaired tree. The shell module's tokenizer, grouping, verb scan, and coverage are now round-2 shaped throughout; the token-cost seam (no manifest read before the identity and mutating gates) is preserved by construction — the async-group and redirect walks touch no files.

### Repair round 3 — merge-gate re-verdict findings (2026-10-06, post-push)

The gate's second re-verdict failed with 8×P1 — the shell rule's long tail, every finding reproduced and repaired with the same RED-first discipline:

- **P1 partially quoted redirection** (`echo changed >".moai/logs/secret"`) — the quote mask is now per character: an unquoted `>` inside a partially quoted word is an operator whose target is the quoted tail.
- **P1 quoted-value assignment** (`X='value' rm …`) — the assignment skip no longer requires the word to be unquoted; a quoted value is still a normal leading assignment.
- **P1 basename-glob parent** (`rm -rf tests` under an overlay `**/*_test.go`) — the ancestor pass now walks a BaseGlob-covered candidate directory (bounded at 5,000 entries, beyond which it under-matches) instead of skipping globs wholesale.
- **P1 `git -C <dir>`** — the git scan captures `-C <dir>` and resolves the subcommand's file arguments against that directory (absolute as given, relative through the tracked cwd).
- **P1 async boundary** (`cd .moai; true & rm logs/secret`) — the boundary now RESTORES the pre-group directory instead of resetting to the root: the group's own cd ran in a subshell, and the main shell never moved.
- **P1 piped cd** (`cd docs | cat; rm …`) — a segment followed by a `|` connector runs in a subshell; its cd no longer moves the main shell's tracked directory (its redirections and mutating commands are still judged).
- **P1 cd arguments vs redirection** (`cd .moai > docs/out; rm …`) — redirection tokens are stripped from the cd segment before the cwd update, so a normal cd to one directory is not misread as a multi-argument cd.
- **P1 sed in-place forms** (`sed -Ei '' …`, `sed '-i' '' …`) — the in-place scan accepts any option cluster containing "i" (plus `--in-place`) and reads quoted option words, which the shell passes through as options.

GREEN on the whole family: `TestProtectedZone` hook 10/10 subtests (ShellMutation swept 39) + config 3/3, `go build ./...` exit 0, `golangci-lint` 0 issues, live judge `JUDGE swept=67 expected=67 fail=0` on a binary built from the repaired tree. The quote mask is now the single mechanism deciding shell syntax (operators and async boundaries live outside quotes); a fourth round of NEW P1 findings would indicate the hand-rolled analyzer needs replacing with a real shell parser — an operator-level decision, flagged in the completion report.

### Repair round 4 — ESCALATION, no repairs executed (2026-10-06)

The gate's third re-verdict failed with 8×P1 + 1×P2. Per the escalation protocol the leader approved before this round ("4라운드에서 신규 P1이 또 나오면 수리 정지하고 '손수 만든 셸 분석기 → 실제 셸 파서 교체'를 운영자 판정 사안으로 올려라"), **no repairs were executed**; the findings are recorded verbatim as the operator-decision package's evidence. Three rounds and 20 findings on one module are the structural signal.

Findings as delivered by the gate (classified for the decision):

- **Shell-analyzer class (7×P1 — the operator decision):** partially quoted words hide real redirections (`>"target"`); multi-target redirections in one word (`>a>b`) keep only the last; quoted redirections escape the cd-argument strip; the last element of a pipeline is not recognized as subshell (`true | cd docs`); a failed `cd` (`||` arm) is assumed successful; an assignment-prefixed `cd` (`X=1 cd .moai`) is missed; the async restore point is taken after the group's own cd (`cd docs && true & rm`); consecutive/quoted `-C` options are mishandled.
- **Outside the analyzer (1×P1 + 1×P2 — trivial fixes, ready on a nod):** the dogfood overlay does not list the checker source (`internal/config/protected_zone.go`) or the manifest template source — editing them weakens every later build (the overlay is exactly where such entries belong); and a broken symlink at a manifest path reads as `os.ErrNotExist` → `absent` instead of `invalid`, silently weakening fail-closed.

The reviewer's own discipline note: every finding came with a real temporary-project execution where the protected file was actually modified — these are observed compromises of the zone, not hypotheses.

### Repair round 5 — second post-escalation verdict, still no repairs (2026-10-06)

The gate's verdict on the round-3 push failed again with 5×P1, all new — delivered AFTER the operator decision was requested, so it is addendum evidence, not a repair trigger: adjacent redirections in one word (`>a>b`) keep only the last target under the round-2 fragment logic; the LAST element of a pipeline (`true | cd docs`) is not recognized as a subshell (the check keyed on the FOLLOWING connector, but the pipe connector PRECEDES the segment); a partially quoted redirect escapes the cd-argument strip (the strip keyed on `QuotedAny()` instead of the `>` mask); a quoted `'-C'` is a real git option; consecutive relative `-C` paths accumulate (`-C a -C ..` resolves against a). Every one is the same class as rounds 1–4: hand-rolled analyzer vs POSIX semantics. Still no repairs — the operator decision owns the surface.







### Repair round 5b — operator decision A: re-platform on mvdan/sh (2026-10-06)

The operator chose option A: the lexical analysis moves to mvdan/sh (v3.14.0 — already a direct dependency in go.mod, so no new supply-chain surface enters; BSD-3-Clause). This commit also lands the two non-analyzer findings:

- **overlay content (round 4 P1)** — the dogfood overlay now lists the checker source (`internal/config/protected_zone.go`, safety_guards) and the manifest template source (`internal/template/templates/.moai/config/sections/protected-zone.yaml`, gate_policy): editing either weakens every later build, and the template→build→deploy chain makes the shipped copy alone insufficient. Both paths exist in the swept tree (liveness resolved=37, up from 35).
- **broken symlink fail-closed (round 5 P2)** — a dangling symlink at a manifest path surfaces the same ENOENT as true absence through `os.ReadFile`; `LoadProtectedZone` now Lstats the path and classifies present-but-unreadable as `invalid` (fail closed). Regression row in config Load; the finding's RED was the reviewer's observed real write.

The shell module's hand-rolled tokenizer, async-group splitter, and redirect scanner are replaced by an AST walker over mvdan's parse: async boundaries from `Stmt.Background`, pipelines and subshells never leak a cd to the main shell, `&&` carries the updated directory, `||` restores the pre-left directory (the right side runs only on failure), redirection targets come from `Stmt.Redirs` (write redirects only; input and here-docs skip), the git `-C` chain accumulates across consecutive and quoted options, and sed's in-place scan reads the literal option clusters. The policy layer — mutating verb set, zone coverage (direct + ancestor + glob-parent walk), deny reason, audit rows, cost seam — is unchanged. Dynamic words still under-match by construction; a parse failure under-matches like the old splitter's unclassifiable text.

All 13 round-4/5 findings are regression rows in `TestProtectedZone/ShellMutation` (swept 50; FileTools 26; config Load/Overlay/Validation 47). GREEN on the whole family: hook+config TestProtectedZone 0 FAIL, config/template/harness packages 0 FAIL, `golangci-lint` 0 issues, `go build ./...` exit 0, live judge `JUDGE swept=67 expected=67 fail=0` on a binary built from the repaired tree.

### Repair round 6 — walker-semantics findings on the mvdan head (2026-10-06)

The gate's first verdict on the mvdan head failed with 7×P1 + 1×P2 — qualitatively DIFFERENT from rounds 4–5: these are implementation defects in the walker's policy port (clear semantic rules with small provable fixes), not tokenizer-class gaps. Classified as ordinary repair, distinct from the analyzer-structural escalation; every row RED-first:

- **P1 conditions execute** — IfClause.Cond and WhileClause.Cond are `[]*Stmt` statement lists and are now walked (`if rm secret; then` deleted the file).
- **P1 statement-wide redirects** — `> zone_dir/secret` (nil Cmd) and `{ true; } > zone_dir/secret` (Block) were missed because the redirect judgment lived only in the CallExpr branch; the walk now judges `stmt.Redirs` for every statement shape, with the CallExpr branch marking itself handled.
- **P1 git resolves against the tracked cwd** — `gitDir` starts from each possible cwd (not ""), a relative `-C` accumulates onto it, and the subcommand's file arguments anchor to the accumulated `-C` (round-5 accumulation preserved).
- **P1 cd redirect ordering** — the cd branch judges its redirects inline BEFORE the cwd update (the round-4 deferred judgment ran after it, misreading `cd docs > zone_dir/secret`).
- **P1 pipeline side isolation** — each pipeline element starts from the same pre-pipe directory set (`cd docs | tee zone_dir/secret` no longer judges the tee target against docs).
- **P1 control-flow directory union** — the walker tracks a SET of possible directories: `||` restores the pre-left set while keeping the post-left set (a successful cd inside a failed `&&` chain persists: `cd zone_dir && false || rm secret` is a denial), and an if unions the condition-false, then, and else worlds (a skipped cd is never applied: `if false; then cd docs; fi; rm zone_dir/secret` is a denial).
- **P2 sed locality** — sed's in-place decision uses a local flag: an earlier mutating command (`rm docs/disposable; sed -n '1p' zone_dir/secret`) no longer turns a read-only sed into a denial.

GREEN on the whole family: `TestProtectedZone` hook 10/10 subtests (ShellMutation swept 60 — the 9 new deny rows plus the read-only-sed allow control) + config 3/3, `go build ./...` exit 0, `golangci-lint` 0 issues, live judge `JUDGE swept=67 expected=67 fail=0` on a binary built from the repaired tree. Escalation note: if the next verdict surfaces ANOTHER set of new P1s, the walker's control-flow model itself goes back to the operator with this round's union-vs-sequence data.

### Repair round 7 — ESCALATION per the pre-approved protocol, no repairs executed (2026-10-06)

The gate's verdict on the round-6 push failed with 8×P1, all new. The leader pre-approved this exact trigger before round 6 ("7차 재판정 신규 P1 시 통제흐름 모델(순서 vs 집합) 운영자 판정 상신 갱신도 사전 승인"), so **no repairs were executed**; the findings are recorded verbatim as the decision package's evidence:

- **P1 compound-command redirects** — `(true) > zone_dir/secret` allowed: Subshell/loop/case redirect judgment missing, and a Block's redirect was judged AFTER its body (`{ cd docs; } > f` misread).
- **P1 backslash escapes** — `rm zone\ dir/secret` allowed: `syntax.Lit.Value` keeps the backslash, and the later normalization reads it as a separator; unquoted escape interpretation is missing.
- **P1 cd failure** — `cd missing; rm …` allowed: the analyzer assumed cd success and dropped the pre-directory; both pre and post must stay possible.
- **P1 `&&` right-skip** — `false && cd docs; rm …` allowed: the right side may not run, leaving the pre-chain directory; the post-right and left-failed states must union.
- **P1 elif condition** — `if false; then true; elif rm secret; then true; fi` allowed: zoneWalkIf skipped the elif's Cond.
- **P1 else entry state** — `if cd zone_dir && false; then true; else rm secret; fi` allowed: the else world starts from AFTER the condition (a successful cd inside it persists), not from pre-if.
- **P1 loop zero-iteration** — `while false; do cd docs; done; rm …` allowed: the body was analyzed as always-run; the entry state and the post-iteration state must both judge later candidates.
- **P1 case arm isolation** — `case x in x) true;; y) cd docs;; esac; rm …` allowed: mutually exclusive arms were walked as a sequence; each arm analyzes from the same entry set and the exit states union.
- **P2 git subcommand misparse (added 2026-10-06, second observation of the same verdict)** — `git grep rm -- zone_dir/secret` (read-only) denied: the argument scan matched the SEARCH TERM "rm" as a mutating subcommand because every argument is checked. The subcommand is the first word after git's global options; later arguments must never be re-interpreted as the subcommand.

All eight are the same class: the walker's CONTROL-FLOW MODEL mixes sequence and set semantics inconsistently. The fix is bounded and mechanical — a sound possible-directory-set walk (redirects judged at every statement entry against the entry set; cd keeps pre∪post; `&&` unions post-right with post-left; `||` unions pre-left with post-left then walks right from the union; if/case walk each arm from the same entry set and union; while judges the body from pre and post once; Lit backslash unescaping). This is the operator decision the protocol reserved.

### Repair round 8 — operator option A executed: the sound possible-directory-set walker (2026-10-06)

The operator chose option A on the round-7 verdict; the leader ordered execution. All eight findings went in as RED regression rows first (observed red on the round-6 tree: 8 deny rows allowed + `git grep rm --` false-denied), then `protected_zone_shell.go` was reworked onto the sound set semantics in one pass:

- **Redirects at statement entry, every shape** — the redirect judgment moved to `zoneWalkStmt` before the command switch: a nil-command statement, a Block (its redirect opens BEFORE the body runs — `{ cd docs; } > f` truncates at the entry cwd), a Subshell, a For/While/Case/If clause all judge `stmt.Redirs` against the entry set (the old code judged a Block's redirect AFTER its body and missed the Subshell/loop/case statements entirely).
- **cd failure worlds** — the cd branch unions its pre-directory set with the post-directory set (a failed cd leaves the caller where it was): `cd missing; rm zone_dir/secret` judges the rm against the entry directory and is denied.
- **`&&`/`||` skip worlds** — after either operator the post-left set is re-unioned into the possible set (the right side may be skipped): `false && cd docs; rm …` judges the rm against the pre-chain directory.
- **if/elif/else** — `walkIfChain` (replacing `zoneWalkIf`) walks the condition, runs the then branch from the post-condition set, runs the else branch from the post-condition set too (a successful cd inside the condition persists into the else), and unions the then/else/post-condition worlds; an elif is an IfClause as the Else member, so its condition is walked as well.
- **loops** — a For/While keeps the zero-iteration world (the entry/post-condition set survives) and unrolls the body once, carrying the post-body set forward; the condition is walked (round 6 P1, kept).
- **case arms independent** — every arm is walked from a clone of the same entry set; the arm worlds union, and the no-arm-matches world (the entry set) survives.
- **Lit backslash unescaping** — `zoneUnescapeLit` resolves `\X` → `X` in literal words (`rm zone\ dir/secret` now names `zone dir/secret`; a documented slight over-approximation inside double quotes, where the shell would keep `\n` — the guard prefers matching a file the command cannot touch over missing one it can).
- **git subcommand misparse (P2)** — the argument scan finds the FIRST non-option word as the subcommand; valued global options (`-c`, `--git-dir`, `--work-tree`, `--namespace`, `--super-prefix`, plus the `-C` accumulator) consume their argument, so `git -c core.quotePath=false checkout -- …` resolves the subcommand correctly and `git grep rm -- …` can never match its search term. The subcommand's file arguments anchor to each possible shell directory, with an accumulated `-C` on top.

GREEN on the whole family: `TestProtectedZone` hook subtests (ShellMutation swept 73 — the 8 new deny rows plus the git grep allow control) + config 3/3, `go build ./...` exit 0, `golangci-lint` 0 issues on hook+config, live judge `JUDGE swept=67 expected=67 fail=0` exit 0 on a binary built from the repaired tree. Escalation contingency stands: another verdict with NEW P1 findings re-escalates with rounds 6/7/8 data.

### Repair round 9 — verdict 8 findings: state carried into iterations and fall-through arms (2026-10-06)

The gate's verdict on the round-8 head (`32ec393a9`) failed with 2×P1 + 1×P2. All three are state-carrying defects INSIDE the operator-approved sound-set model — the model already unions worlds; the round-8 walk just did not carry the post-arm / post-iteration state into the constructs that consume it. Per the leader's lane-autonomy directive (2026-10-06: lanes decide within the approved model; keep-set only for the operator), this repaired directly instead of re-escalating. All rows RED-first (observed allow on the round-8 tree, matching the verdict verbatim):

- **P1 loop iterations beyond the first** — the body was analyzed once, so `for i in 1 2; do cd a; printf changed > secret; done` (entry `a/a/`) missed iteration 2's directory `a/a`. ForClause with a fully literal item list (`WordIter.Items`, every item a plain unquoted word) now walks EXACTLY that many iterations, each from the accumulated set; a literal list above 64 items, a C-style loop, and any quoted/expanding item take the fixed-point path. WhileClause (and `until`, folded by `WhileClause.Until`) walks its body to a FIXED POINT — each pass starts from the accumulated set, stopping when a pass adds no new directory — so iteration k+1 is judged at iteration k's exit state.
- **P1 uncomputable loops fail closed** — the fixed point is bounded at 16 passes; a body still growing at the cap (an accumulating `cd a` chain never converges) sets the walker's `unbounded` flag, and a mutating command with no covered candidate is then DENIED with `category=loop-unbounded` instead of allowed on an incomplete walk (the verdict's own alternative: "계산할 수 없으면 차단"). The flag never fires when nothing mutates, and it does not fire with no manifest (ZoneStateAbsent keeps the baseline-floor degrade).
- **P1 case fall-through** — every arm was analyzed from the case's entry set only, so `;&` (unconditional fall-through) and `;;&` (resume) missed the prior arm's exit directory (`case x in x) cd zone_dir ;& y) cd zone_dir ;; z) printf … ;; esac`). mvdan's `CaseItem.Op` distinguishes Break/Fallthrough/Resume; the walk now starts every arm from the case's entry set UNION the accumulated earlier-arm worlds — a sound over-approximation that carries fall-through state while keeping arm independence (the no-match world still survives).
- **P2 latency instrument** — `evidence/baseline-latency.sh` discarded the hook's exit status and stdout, so a broken binary measured as three PASS rows with exit 0 (observed: `bash baseline-latency.sh /usr/bin/false /usr/bin/false 1` → all `verdict=PASS`, exit 0). A sample is now valid only when the hook exited 0, its stdout parsed as JSON, and `hookSpecificOutput.permissionDecision` equals the class's expected decision (deny for identity-inside — the guard must have engaged; allow for the other two). Any invalid sample makes the class INCONCLUSIVE and the script exits 1. Observed post-fix: the false/false probe reads all-INCONCLUSIVE with exit 1; the A/A positive control on the current-tree binary reads PASS with exit 0.

GREEN: `TestProtectedZone` hook+config (ShellMutation swept 78 — 5 new deny rows: for-iteration, while-accumulation, `;&`, `;;&`, and the loop-unbounded fail-closed row). One test-fixture correction during GREEN: the loop-unbounded row originally shared the `a/a/` manifest, whose directory entry already covers the deep-world candidates via prefix — the row moved to a zone_dir-only manifest so the deny provably comes from the bound itself.

### Repair round 10 — absorb origin/main (PR conflict) + the independent auditor's findings (2026-10-06)

Two inputs landed together after the round-9 push:

- **PR #1757 became CONFLICTING with base main** (observed: `gh pr view` → `mergeable: CONFLICTING`, `mergeStateStatus: DIRTY`; `graph-freshness-conflict-guard` failed with "mergeable state unknown after polling" — the merge ref could not be created, so the docs gate was ABSENT, not red). The main side had moved (the transition PR fleet merging). Absorbed `origin/main` at `9da74469a` via a merge commit into this branch (`d4d9c56dd`). Conflict surface was exactly two files (`comm -12` of the two changed-file sets from merge-base `1c2336de0`): `.github/workflows/spec-lint.yml` and `CHANGELOG.md` (auto-merged). spec-lint resolved to the MAIN side: its fetch-step is the completed cutover my round-3 interim repair had deferred to — under GitHub Flow only `origin/main` is fetched and the deleted develop ref is never named. My interim best-effort develop fetch is fully superseded by it.
- **The independent fallback auditor's verdict** (`.moai/reports/t1510/card-review-r8.md`, FAIL — 3×P2 + 2×P3 on the round-8 head) cross-checked against the round-9 tree: its case-fallthrough finding (P2 #1) and its `git -c "$CFG"` finding (P3 #5) were already repaired in round 9 — the gate and the auditor converged on the same holes from independent reads, and both repro strings are now regression rows. Three findings were new, repaired RED-first:
  - **P2 TimeClause** — `time rm zone_dir/a.log` parsed to `TimeClause` and fell to the under-match default; the clause's inner statement now walks (`time` runs the command, timed).
  - **P2 function decl+call** — `cleanup() { rm zone_dir/a.log; }; cleanup`: a `FuncDecl` registers its body under its name (a declaration alone runs nothing); a later `CallExpr` naming it walks the body in the caller's directory state, with a `calling` set breaking recursive declarations (documented under-match). mvdan has no FuncCall node — calls are plain CallExprs, which is why the call site needed the registry.
  - **P3 DplOut** — `2>&1` (`DplOut`) duplicates one descriptor onto another and writes nothing, but it was not in the redirect skip list: read-only commands were flagged mutating, so under an INVALID manifest even `echo hi 2>&1` hit the fail-closed deny (observed RED). DplOut now skips alongside RdrIn/Hdoc/DplIn; the invalid-manifest fail-closed keeps its positive control (`rm zone_dir/a.log` under the same loader still denies with `manifest=invalid`).

GREEN on the MERGED tree: `TestProtectedZone` hook+config (ShellMutation swept 84 — 4 new deny rows: TimeClause, function call, the auditor's `;;& *` fall-through repro, `git -c "$CFG" rm`; plus the DplOut allow control and its invalid-manifest positive control), `go build ./...` exit 0, GOOS=windows build OK, `golangci-lint` 0 issues, gofmt clean, `TestHMPSourceGuard` ok, live judge `JUDGE swept=67 expected=67 fail=0` exit 0.

### Repair round 11 — verdict findings: correction, dangling symlinks, loop conditions, subshell functions + a second main absorb (2026-10-06)

The gate's verdict on the round-10 head failed with 4×P1 + 4×P2. Scope split first:

- **P2 4건은 이 카드의 소관이 아니다 — 보고로 라우팅, 수리하지 않음.** `internal/cli/factory_card.go` 자동 힌트 덮어쓰기, `internal/cli/factory_bundle.go` 묶음 중복 ID, `internal/cli/todo.go` 신규 카드 `--files` 겹침 누락과 `add --pick` 발행 제시 누락 — 네 파일 모두 이 카드가 한 번도 건드리지 않은 main 착지 코드다(이 카드의 diff에 없음). 실측: main은 흡수 지점 `9da74469a`에서 `10df085da`(t1502, #1759)로 진행했고 그 delta는 `todo.go` 5줄뿐 — 네 발견은 main 자체의 결함으로, 소관 카드(팩토리 카드·묶음·todo 발행 경로)의 후속 수리 대상이다. 이 PR에서 외부 범위를 섞지 않는다(범위 규율). 리더 착지 보고에 전문을 실어 라우팅을 청구한다.
- **재흡수**: main 진행분을 다시 흡수했다(`10df085da`, 병합 커밋 `c419cc5a2`, 무충돌) — PR diff가 다시 카드 스코프만 담도록.

The 4 P1s are this card's protected-zone surface, repaired RED-first (all four observed allow on the round-10 tree, matching the verdict):

- **P1 `>&` file target** — round 10's blanket DplOut skip over-corrected: `printf changed >& zone_dir/secret` writes the file in dialects accepting the spelling. Only a NUMERIC descriptor target (`2>&1`) writes nothing now; a word target judges as a write (`isZoneDigits` gate). The round-10 `echo hi 2>&1` control stays allow (its target is the digit `1`).
- **P1 dangling symlink destination** — `zoneResolve` treated ANY `EvalSymlinks` failure at a component as "missing": `innocent.md -> zone_dir/new.md` (target absent) resolved to the bare link name and a Write through it created the file inside the zone (observed RED). The component walk now distinguishes: `Lstat` fails → genuinely missing → unresolved tail (unchanged); the component EXISTS → a real link the shell would follow → `Readlink` its destination, rejoin it onto the walk recursively, with a depth bound (`zoneSymlinkDepthBound` 32, fail closed beyond — standing in for ELOOP).
- **P1 loop condition every iteration** — `while cd a; do printf changed > secret; done` runs the CONDITION each round (its `cd a` accumulates: pass 2 writes at `a/a`), but the fixed point walked only the body. `walkBodyFixedPoint` now takes the condition and alternates cond+body per pass — one pass is one loop round; convergence and the `loop-unbounded` fail-closed bound are unchanged.
- **P1 subshell function redefinitions** — `( f(){ :; }; )` overwrote the shared `funcs` registry, so a later caller-side call `f` walked the EMPTY body. Every subshell-shaped walk now runs on its own registry copy (`cloneZoneFuncs`): explicit Subshell, each pipeline element, and background statements — redefinitions die with the subshell exactly as the directory set does.

GREEN on the re-absorbed tree: `TestProtectedZone` hook+config (ShellMutation swept 88 — 3 new deny rows plus the dangling-symlink Write row), `go build ./...` exit 0, GOOS=windows OK, `golangci-lint` 0 issues, gofmt clean, `TestHMPSourceGuard` ok, live judge `JUDGE swept=67 expected=67 fail=0` exit 0.

### Repair round 12 — git --work-tree anchoring + six more foreign-scope P2s recorded (2026-10-06)

The gate's verdict on the round-11 head failed with 1×P1 + 6×P2. The card's own surface produced exactly ONE finding — the review is reading the whole tree, and the foreign-scope findings keep accumulating for the routing report:

- **P2 6건 — 이 카드 소관 아님, 라우팅 누적**(라운드 11 목록과 2건 중복): `internal/factory/backlog_store.go:144` MergeBacklogRecords가 자식의 `SpawnedBy` 부모 참조를 매핑 누락 · `internal/cli/factory_bundle.go:211` 큐 재정렬 시 작업 중 허브 카드와의 선행 순서 붕괴 · `internal/cli/todo.go:940` 신규 카드 `--files` 겹침 미전달(라운드 11 중복) · `internal/cli/todo_issuance.go:216` 변경 파일 목록 공백 분리(`two words.go` 분리 — `-z` NUL 분리 필요) · `todo_issuance.go:212` 탐침 기준 브랜치 `develop` 하드코딩 · `todo_issuance.go:207` 명시적 프로젝트 루트 미전파. 네 파일 모두 이 카드 diff에 없는 main 착지 코드다. 리더 라우팅 보고에 누적된다.
- **P1 `git --work-tree`** — `git --work-tree=zone_dir restore -- secret` allowed: the option scan consumed `--work-tree` as a valued global option but DISCARDED its value, and only `-C` anchored the subcommand's paths. `--work-tree` is now captured in both spellings (inline `=` and separate value) and anchors the file arguments as its own base alongside `-C` — both anchorings are judged when both are present (sound union). RED rows observed allow first (both spellings).

GREEN: `TestProtectedZone` hook+config (ShellMutation swept 90 — 2 new deny rows), `go build ./...` exit 0, GOOS=windows OK, `golangci-lint` 0 issues, gofmt clean, live judge `JUDGE swept=67 expected=67 fail=0` exit 0. Separately observed: CodeRabbit completed a full review of the round-11 head (`state=success`, description `Review completed`) after an explicit re-review request — the merge-discipline's first condition held there and is re-checked per head.

### Repair round 13 — invocation spellings and a case-preserving glob walk (2026-10-06)

The gate's verdict on the round-12 head failed with 4×P1 + 2×P2. The card's own surface carried all four P1s — invocation-spelling class, each observed RED first (allow on the round-12 tree, matching the verdict):

- **P1 absolute executable path** — `/bin/rm zone_dir/sentinel.txt` allowed: the verb check was exact-match on the raw word. A literal executable path now folds to its base name before the verb lookup (`/bin/rm` → `rm`); a path to some OTHER binary folds to a base matching no verb, exactly as before (sound either way).
- **P1 case-preserving glob walk** — `rm -r Tests` allowed on Linux while `Tests/example_test.go` sits under a `**/*_test.go` entry: the ancestor pass walked the FOLDED form (`tests`), which finds nothing on a case-sensitive filesystem. The walk now joins `form.Display` (original case); the fold stays on the comparison side only. **RED를 로컬에서 관측할 수 없는 유일한 행** — darwin의 대소문자 무시 파일시스템은 접힌 경로로도 실제 디렉터리를 찾아주므로 이 행은 darwin에서 공허하게 초록이다. 판정서의 재현은 Linux에서 측정됐고, 이 행은 CI linux 레그의 회귀 가드로 기록된다(미관측 간격은 완료 보고의 Gaps 절에 이름을 올린다).
- **P1 GNU sed `--in-place=.bak`** — the suffixed long form is in-place too; the scan now accepts the `--in-place=` prefix alongside the exact form and the `-i` cluster.
- **P1 option-attached target paths** — `cp --target-directory=zone_dir source.txt` dropped the attached value with the flag: `zonePathCandidates` now extracts a long option's `=value` as a candidate (a non-path value matches nothing and costs one lookup). The separate-value form already worked (the bare value was never dropped); a GNU short option with an attached value (`-tDIR`) stays an accepted under-match, documented.

**P2 2건 — 외부 소관, 라우팅 누적(누적 12건)**: `internal/cli/factory_card.go:1314` 자동 허브 힌트가 기존 `HintAfter`를 재계산해 덮음(라운드 11의 factory_card 발견이 P1으로 격상된 것과 동일 결함군) · `factory_card.go:821` 레코드 없는 picked 후보가 선행조건 검사 전에 힌트를 기록하고 claim 실패로 종료. 둘 다 이 카드 diff에 없는 main 착지 코드.

GREEN: `TestProtectedZone` hook+config (ShellMutation swept 94 — 3 new deny rows plus the original-case glob row), `go build ./...` exit 0, GOOS=linux+windows OK, `golangci-lint` 0 issues, gofmt clean, live judge `JUDGE swept=67 expected=67 fail=0` exit 0.

### Repair round 14 — branch-local function definitions join the possible-worlds model (2026-10-06)

The gate's verdict on the round-13 head failed with 1×P1 (this card) + 6×P2 (foreign scope). The P1: `f(){ rm zone_dir/guard.go; }; if false; then f(){ true; }; fi; f` was ALLOWED — a branch's redefinition overwrote the shared function registry even though the branch may never run, so the caller-side call walked the EMPTY body. The registry now joins the same possible-worlds model the directory set uses:

- `funcs` maps a name to its POSSIBLE BODIES (`[]*syntax.Stmt`). A straight-line redefinition REPLACES (`f(){a}; f(){b}; f` walks only `b`); a BRANCH JOIN unions (`mergeZoneFuncs`, bodies deduped by node identity): if/elif/else clone the post-condition registry per branch and merge the worlds; case arms merge entry ∪ accumulated arm worlds per arm and union at the end; loops keep the zero-iteration world's pre-loop registry; subshells, pipeline elements, and background statements keep their snapshot/restore (round 11 P1, unchanged). A call walks EVERY body the name may hold, under the same recursion guard.
- RED rows observed first (both allow): the verdict's if-form repro and a case-form twin.

**P2 6건 — 외부 소관, 라우팅 누적(누적 18건; 4건은 기존 발견의 재지적)**: `internal/cli/factory_bundle.go:211` 다중 허브 선행 검사 누락(마지막 항목만 HintAfter에 반영) · `factory_card.go:821` 레코드 없는 picked 후보 선행조건 생략(재지적) · `todo_issuance.go:152` SPEC 읽기 자체의 시간 제한 부재(FIFO로 막힘 — 신규) · `todo.go:940` --files 겹침(재지적) · `todo.go:873` add --pick 제시(재지적) · `todo_issuance.go:274` engage가 자기 자신을 exact 이웃으로 표시(신규). 전부 이 카드 diff에 없는 main 착지 코드.

GREEN: `TestProtectedZone` hook+config (ShellMutation swept 96 — 2 new deny rows), `go build ./...` exit 0, GOOS=linux+windows OK, `golangci-lint` 0 issues, gofmt clean, live judge `JUDGE swept=67 expected=67 fail=0` exit 0.

### Repair round 15 — short-circuit function definitions union (2026-10-06)

The gate's verdict on the round-14 head failed with 1×P1 (this card) + 5×P2 (foreign scope, 4 re-flags + 1 new). The P1: `f(){ echo changed > zone_dir/a.md; }; false && f(){ :; }; f` was ALLOWED — the `&&` right side's redefinition REPLACED the registry even though the right side may be skipped. The `&&`/`||` branch now snapshots the post-left registry and merges it back after the right side (the same possible-worlds union the branch joins got in round 14). RED row observed first (allow, matching the verdict).

**P2 5건 — 외부 소관, 라우팅 누적(누적 23건; 4건 재지적)**: `factory_card.go` 자동 힌트 덮음(4번째 재지적)·picked 건너뛰기(재지적)·`factory_bundle.go` 중복 ID(재지적)·`todo.go` --files(재지적)·`todo_issuance.go:244` 탐침 시간 초과 후 git 프로세스 미종료(신규 — 취소 가능한 context 필요). 전부 이 카드 diff에 없는 main 착지 코드.

GREEN: `TestProtectedZone` hook+config (ShellMutation swept 97 — 1 new deny row), `go build ./...` exit 0, GOOS=linux+windows OK, `golangci-lint` 0 issues, gofmt clean, live judge `JUDGE swept=67 expected=67 fail=0` exit 0.

### CI repair — codemaps rotation (2026-10-06)

The graph-freshness required check failed on the round-15 head (`codemaps metric=described-source-diff value=57 threshold=40 verdict=stale`) — the window since the codemaps anchor `d0378d37c` had accumulated this card's 13 rounds plus the absorbed fleet commits without a rotation. The rotation followed the established convention (t1524 판 모양):

- 측정 먼저: 창 57개를 검사기와 같은 술어로 독립 재현(신규 17 · 수정 40 · 삭제 0), 규모 표 일곱 값 재측정(비테스트 1559→1576 · 테스트 2851→2885 · 패키지 169 불변 · 최상위 87 불변 · 엣지 476/301→476/313 · 템플릿 606→607 · go.mod 불변), 패키지 행 앵커 트리 직접 대조(cli 427→433 · hook 158→163 — 신규 5파일 삭제 0 · factory 61→63 · config 64→66 · homestate 32→33 · 나머지 불변). 측정 과정에서 "memo/taxonomy가 신규 패키지"라는 1차 판독이 틀렸음을 파일 집합 대조가 잡아냈다 — taxonomy는 앵커에 존재하고 창 신규는 +2파일.
- 저작: modules.md 새 창(일곱 착지분 카드별 산문) · overview.md 머리 재측정 단락(이전 판은 이전으로 강등) · entry-points.md(등록 줄 불변, 다중 등록 줄 안 표면 추가) · data-flow.md 보호 구역 거부 흐름 § · fold-judgments.txt(신규 판정 없음 — 전부 named row).
- 커밋 2벌(t1485 패턴): `docs(t1510): codemaps regenerate…`(`faa179e7e`) → `chore(t1510): stamp codemaps provenance at the refresh commit`(`f63af2034`, provenance가 갱신 커밋을 가리킴 + generated_at 실측시각).
- 검증: `go run ./cmd/moai mx scan --quiet && go run ./cmd/moai graph check` — codemaps fresh(value=0) · mx-index fresh · citations fresh · exit 0(CI 작업과 동일 순서 미러링). 첫 검사에서 citations가 적색로 전환됐는데 원인은 내가 쓴 중괄호 축약 `protected_zone_{guard,shell,path}.go`가 실존하지 않는 인용 경로로 읽힌 것 — 실경로 나열로 수리. 하나의 교훈: `$?`를 파이프 뒤에서 읽으면 head의 종료코드를 읽는다 — 검증 판정은 파이프 없이.

### Repair round 16 — function bodies are worlds, and the fixed point must see the registry (2026-10-06)

The gate's verdict on the rotation head failed with 2×P1 (this card) + 4×P2 (foreign). Both P1s are holes in the round-14 possible-bodies model, diagnosed with a throwaway AST+registry probe (deleted after use):

- **P1 nested definitions across alternatives** — `if true; then f(){ g(){ rm …; }; }; else f(){ g(){ :; }; }; fi; f; g` was ALLOWED: the call site walked f's two possible bodies SEQUENTIALLY, so the second walk's `g` declaration REPLACED the first's (straight-line replaces — but these are alternatives, not a sequence). The probe showed `registry g -> 1 body`. The call site now isolates each body: clone the registry per body, walk, merge back — bodies are worlds, exactly like branch joins. (The parse shape was verified correct first — the model, not the parser, was wrong.)
- **P1 loop fixed point ignored the registry** — the fixed point converged when the DIRECTORY set stopped changing; `while test -e x; do f; f(){ rm x; }; done` redefines f every round while the directory set stands still, so iteration 2's call walked a different body than the analysis saw. Convergence now requires the registry to be unchanged too (`zoneFuncsEqual`, body-pointer sets per name); non-convergence still lands in the `loop-unbounded` fail-closed.
- **P2 4건 — 외부 소관, 라우팅 누적**: `factory_card.go:821`(재지적)·`todo_issuance.go:109`(재지적 — 읽기 전용 조회의 스키마 마이그레이션)·`backlog_issuance.go:511`(신규 파일 — --files 겹침의 factory 측 절반)·`backlog_relation.go:215`(재지적 — archived 관계 투영). 전부 이 카드 diff에 없는 main 착지 코드.

RED rows observed first (both allow, matching the verdict verbatim). GREEN: `TestProtectedZone` hook+config (ShellMutation swept 99 — 2 new deny rows), `go build ./...` exit 0, GOOS=linux+windows OK, `golangci-lint` 0 issues, gofmt clean, `TestHMPSourceGuard` ok, live judge `JUDGE swept=67 expected=67 fail=0` exit 0.

### Repair round 17 — bounded recursion unrolling, precise call-result merge (2026-10-06)

The gate's verdict on the round-16 head failed with 1×P1 + 1×P2 (both this card, both function-model refinements) + 2×P2 (foreign re-flags):

- **P1 bounded recursion unrolling** — the `calling` guard returned immediately on a recursive call, so the recursion's directory-driven unrolling was never walked (`f(){ echo > secret.md; test -f stop || { cd ..; f; }; }; cd zone_dir/a/b; f` writes into `zone_dir/secret.md` on the third level). A recursive call now RE-ENTERS its body bounded — each re-entry walks from the walker's CURRENT state, so the `cd ..` between levels is real — with a per-name counter (`zoneRecursionBound` 8); past the bound the walk sets the unbounded flag (the fail-closed denial), exactly the verdict's own alternative. 진행 중 한 번 헛점을 잡았다: 첫 RED 행이 디렉터리 엔트리(`zone_dir/`)를 써서 1패스 커버로 구멍을 가렸고, 판정서처럼 정확 파일 엔트리(`zone_dir/secret.md`)로 고친 뒤에야 적색이 관측됐다 — 가림 마스킹은 RED-first의 사각지대다.
- **P2 call-result merge precision** — the call site seeded its result with the PRE-call registry, so a body that CERTAINLY redefines a name kept the old definition alive: `g(){ rm x; }; f(){ g(){ :; }; }; f; g` was DENIED on the retired body (false positive, observed). The call's result is now the union of the bodies' own outcomes — each walked from the pre-call entry state, merged bodies-wise — so a certain redefinition retires the old definition while genuinely-possible alternatives keep both.
- **P2 2건 — 외부 소관(재지적)**: `todo.go:1055` 읽기 전용 명령의 큐 디렉터리 실제 이관(신규 성격이나 todo_issuance 계열 재지적)·`todo.go:940` --files 겹침(누적 재지적).

RED rows observed first — the recursion row allow on the round-16 tree (after the fixture correction), the over-deny control deny. GREEN: `TestProtectedZone` hook+config (ShellMutation swept 101 — 1 new deny row + 1 allow control), `go build ./...` exit 0, GOOS=linux+windows OK, `golangci-lint` 0 issues, gofmt clean, live judge `JUDGE swept=67 expected=67 fail=0` exit 0.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-05
tier: M
implementation: M1 `8f6a4475a` + M2 `9ec0d557b` + M3 (this commit), branch `WT-self-improve-protected-zone`
definition_of_done: AC-SIPZ-001..004 and 009 green via live commands (RED→GREEN pairs above; L-9's `ls` reds pre-date M1 and its green is carried by the M1 commit plus this run's full-suite green); AC-SIPZ-005..013 green with swept counts shown (AC-SIPZ-007 via the latency A/B above; AC-SIPZ-011 via the four Liveness conditions; AC-SIPZ-012 via the GOOS-independent table passing on darwin with the linux leg in CI); every §C mutant executed and its red recorded above except m (CI-owned, disclosed)
judge: `JUDGE swept=67 expected=67 fail=0` live, final-tree binary, exit 0
quality_gates: gofmt clean; go vet exit 0; golangci-lint (v2.1.6) 0 issues on hook/config/harness; go test -race 0 data races; shipped-key inventory and HMP literal registries updated with reasons
unobserved_gaps: Windows behaviour (G6, table-tested only); a stale deployed binary (G6, binary-lag doctor's surface); whether a denied subagent returns a blocker report (G8); the linux-leg execution of mutant m


## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-10-05
sync_commit_sha: "180613914"
artifacts: CHANGELOG.md entry ([Unreleased]/Fixed, SPEC-ID count 0 before this entry — B12 checked); spec.md status transition in-progress → implemented → completed merged into this sync commit; codemaps not rotated (the guard module is documented in its package CLAUDE.md scope and the SPEC artifacts — no structural surface added)
sync_audit: deferred to the leader's integration review gates (factory card flow; the lane's card-review evidence lands at .moai/reports/t1510/card-review.md before integration)

