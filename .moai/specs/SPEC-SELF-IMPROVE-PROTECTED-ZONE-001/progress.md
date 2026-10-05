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

