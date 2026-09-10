# t332 card sweep — batch 2

Cards: t231 t233 t236 t237 t239 t240 t242 t243 t244 t247 t248  (11 entries)
Worktree HEAD: 6165f9f5e
Pinned develop: ee50984abe4f11ac337382b48a26328f091e200a
Pinned main:    48239c7dc7428c8751a04f6321887c2d36123884

### t231

**Premise (one sentence).** `worktree clean`'s lock-source-unreadable path gives a machine-readable
`anchored: undetermined` signal in `--json` but no distinguishable exit code on the human path, so
a caller relying on exit code cannot tell a degraded run from a clean success.

**Premise verdict.** `holds` — verified against the tree at HEAD. `internal/cli/worktree/clean.go`
`reportStaleWorktrees` (the `--json` path, around line 372-384) explicitly discards the lock error
(`candidates, _ := classifyStaleWorktrees(...)`) and always `return nil`. The human path
`cleanStaleWorktrees` (line ~213-280) prints a stderr notice via `lockSourceUnreadableNotice` on
`lockErr != nil` but still falls through to `return nil` at every exit point — no error, no
distinguishable code. `internal/cli/worktree/clean_lock_unreadable_test.go`
(`TestCleanStale_UnreadableLockSourceRemovesNothing`) asserts this is intentional:
`t.Fatalf("runClean must stay non-blocking (REQ-WR-016), got error: %v", err)` — the test fails if
the command ever returns a non-nil error on this path. So REQ-WR-016 (the requirement the card asks
to amend) is still in force, unmodified.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree named for t231 in `00-worktree-list.txt`)

**Claim.** Both the `--json` and human `worktree clean` paths return exit 0 on a lock-source-read
failure; only `--json`'s payload field distinguishes the degraded run. Note: the card describes the
human path as exit 1 ("exit code는 1로 뭉개짐"); the code as read returns exit 0 (nil error) on
every branch, which is a *stronger* form of the same defect (no discrimination at all, not merely a
generic 1) — this is a minor factual correction to the card's own diagnosis, not a rebuttal of its
premise.

**Evidence.**
```
$ grep -n "causeLockSourceUnreadable" internal/cli/worktree/clean.go
315: c.KeepReason = fmt.Sprintf("cause=%s; could not read the worktree lock state: %v", causeLockSourceUnreadable, lockErr)
494: // causeLockSourceUnreadable is the cause token for a lock source that could
496: const causeLockSourceUnreadable = "lock-source-unreadable"
502: return fmt.Sprintf("moai: worktree clean degraded (cause=%s; git worktree list --porcelain failed: %v): no worktree removed", causeLockSourceUnreadable, err)

$ sed -n '372,384p' internal/cli/worktree/clean.go
	candidates, _ := classifyStaleWorktrees(worktrees, base)
	...
	return enc.Encode(candidates)   # always nil unless JSON encoding itself fails

$ sed -n '228,280p' internal/cli/worktree/clean.go
	candidates, lockErr := classifyStaleWorktrees(worktrees, base)
	if lockErr != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), lockSourceUnreadableNotice(lockErr))
	}
	...
	return nil   # every terminal branch in this function

$ sed -n '46,58p' internal/cli/worktree/clean_lock_unreadable_test.go
	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean must stay non-blocking (REQ-WR-016), got error: %v", err)
	}
```

**Baseline-attribution.** All figures measured against worktree HEAD `6165f9f5e` by direct
`grep`/`sed`/`Read` of `internal/cli/worktree/clean.go` and
`internal/cli/worktree/clean_lock_unreadable_test.go` in this run.

**Gaps.** Did not run `main.go`'s `cli.ResolveExitCode` to trace whether any wrapping layer above
`cleanStaleWorktrees`/`reportStaleWorktrees` could still turn a nil error into something other than
0 (unlikely, since `main.go` only special-cases a non-nil error). Did not check whether a *different*
subcommand path (`moai worktree done`, `recover`) shares this classifier and inherits the gap.

**Residual-risk.** If a future refactor threads `lockErr` through as a returned error instead of a
side-channel notice, the exit code could change without this file's local tests catching a
regression in REQ-WR-016's stated invariant, since the invariant itself is what's being asked to
change.

**Proposed disposition.** `keep` — the premise is current and the requested exit-2 discrimination +
REQ-WR-016 amendment genuinely has not happened; rests on the `clean_lock_unreadable_test.go`
non-blocking assertion above.

**Overlap candidates.** none observed among the in-scope batch ids.

---

### t233

**Premise (one sentence).** `moai gate` silently passes (`exit 0`, no notice) the lint axis for a
non-eslint Node project (biome/oxlint) because the only Node lint step is eslint,
config-file-gated, and a config-gated skip previously produced no notice.

**Premise verdict.** `falsified` — the specific "무통지"(no-notice) half of the premise is
contradicted by the current tree. `internal/hook/quality/gate.go` still lists only one Node lint
step (`eslint`, gated on eslint config files — no biome/oxlint step exists, so that half of the
underlying capability gap is real), **but** `executeStep`'s config-gated skip branch now calls
`g.summary.markSkipped(step.name, fmt.Sprintf(reasonConfigFilesAbsentFmt, ...))`, and
`QualityGate.Run` returns `joinBlocks(out, g.summary.render())` — every run, pass or fail, renders a
per-step outcome/reason block including skipped steps and their reason. This is exactly the "①-④"
class of fix the card proposes under item ④ ("pass 시 1줄 요약 출력"), already implemented. The
card's own cited line numbers (gate.go:158-163/787-789/811-820/343) do not match current content —
line drift consistent with other work (a Node `typecheckStep` addition, referenced in an inline
comment as closing the *type*-check blind spot for biome-style projects) having landed on this file
since the card was written.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** The silent-pass shape the card's headline mutant describes (no notice at all) has already
been closed by an unrelated-looking summary/render mechanism; the narrower "biome/oxlint rules
themselves are not run" gap (candidates ②/① in the card) remains open.

**Evidence.**
```
$ grep -n "eslint\|biome\|lintSteps" internal/hook/quality/gate.go | sed -n '1,10p'
197-198: # comment: "Node had only lint and test, and a project whose linter config is
           absent (biome instead of eslint, say) skipped lint too — leaving a
           type-broken build to pass the gate. typecheckStep closes that hole."
204: name: "eslint", binary: "npx", args: []string{"eslint", "."}, optional: true,

$ grep -n "biome\|oxlint\|scripts.lint" internal/hook/quality/gate.go
(no matches)

$ sed -n '891-894p' internal/hook/quality/gate.go
	if len(step.configFiles) > 0 && !g.anyConfigFileExists(step.configFiles) {
		g.summary.markSkipped(step.name, fmt.Sprintf(reasonConfigFilesAbsentFmt, strings.Join(step.configFiles, ", ")))
		return true, ""
	}

$ sed -n '500-505p' internal/hook/quality/gate.go
	return joinBlocks(out, g.summary.render())
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`, file
`internal/hook/quality/gate.go` and `internal/hook/quality/gate_summary.go`, read directly (no
runtime execution of `moai gate` performed).

**Gaps.** Did not actually run `moai gate` against a synthetic biome-only fixture to observe the
literal stdout (static read only, per restraint). Did not check whether the rendered summary block
is suppressed under any output mode (`--json`, quiet flags) that might reproduce the "0 바이트
exit 0" the card's issue reproduction describes.

**Residual-risk.** The card's downstream complaint — `moai gate` still does not actually catch a
biome/oxlint-detected lint error (unused var etc.) because no biome lint step exists — is real and
unaddressed; only the notice/silence aspect is closed. An operator reading only "falsified" without
this residual note could wrongly assume the whole issue is resolved.

**Proposed disposition.** `needs-operator-decision` — the card as currently worded (centered on
silent/무통지 pass) is falsified, but the narrower remaining gap (no biome/oxlint lint step) may
still warrant a rewritten, narrower card. Rests on the `markSkipped`/`render()` evidence above.

**Overlap candidates.** none observed among the in-scope batch ids (card itself references #1639 as
a related-but-distinct issue, not in scope).

---

### t236

**Premise (one sentence).** `MOAI_PROJECT_DIR` goes stale after a worktree switch because
Enter/ExitWorktree do not fire the `CwdChanged` hook event the env var's sole producer depends on,
and `verify_snapshot`/`verify_trend` remain unregistered for the `project_root` parameter that would
let callers route around the stale fallback.

**Premise verdict.** `unverified` — the premise has two halves and only one could be decided, so the
premise **as a whole** is undecided. (Orchestrator normalization, M3 post-check: the worker wrote a
compound verdict here; AC-BH-011 admits exactly one of `holds` / `falsified` / `unverified`, and a
partially-decided premise is the case `unverified` exists for. The substance below is the worker's,
unchanged.) Checked and holding: the
`project_root` parameter registration gap. `internal/cli/mcp_server.go`'s `verify_snapshot` (line
192-198) and `verify_trend` (line 201-206) tool registrations carry no `project_root`
`mcp.WithString` parameter, matching `moai-mcp-tools.md`'s own catalogue of the 9 tools that DO
accept it (verify_snapshot/verify_trend absent from that list). Not checked: whether
`EnterWorktree`/`ExitWorktree` actually fail to emit `CwdChanged` at runtime — that is a live
Claude-Code-runtime behavior claim (the card's own citation is a "reporter's runtime trace... 정적
재현 불가"), not something a static grep of this repository can confirm or refute, since
`CwdChanged` is a host-runtime-emitted event this codebase only *handles*, never emits.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** The `project_root`-registration half of the card (3 of the 5 originally-named tools
landed project_root; verify_snapshot/verify_trend still lack it) is confirmed current.

**Evidence.**
```
$ sed -n '191,206p' internal/cli/mcp_server.go
	add("verify_snapshot", mcp.NewTool(
		"verify_snapshot",
		mcp.WithDescription(...),
		mcp.WithString("key", mcp.Required(), ...),
		mcp.WithString("command", ...),
		mcp.WithInteger("exit_code", ...),
	), handleVerifySnapshot)
	add("verify_trend", mcp.NewTool(
		"verify_trend",
		mcp.WithDescription(...),
		mcp.WithString("key", mcp.Required(), ...),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleVerifyTrend)
	# neither call carries mcp.WithString("project_root", ...)
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`,
`internal/cli/mcp_server.go`, and cross-checked against the always-loaded
`.claude/rules/moai/core/moai-mcp-tools.md` catalogue (9-tool `project_root` list, which also
excludes verify_snapshot/verify_trend), consistent between doc and code.

**Gaps.** Did not check `internal/hook/cwd_changed.go` or the Claude Code runtime spec for whether
`EnterWorktree`/`ExitWorktree` are documented to emit `CwdChanged` — this is outside what a static
repo read can settle, and the card's own evidence is a runtime trace, not a code citation.

**Residual-risk.** If `CwdChanged` actually does fire on Enter/ExitWorktree (contra the card's
runtime trace), the card's headline defect (stale `MOAI_PROJECT_DIR`) may already be resolved and
only the narrower `project_root`-registration gap remains live.

**Proposed disposition.** `needs-operator-decision` — the checkable half holds; the runtime half
cannot be settled from this tree and would need a live reproduction (enter/exit a worktree, inspect
whether `CwdChanged` fired) rather than static reading.

**Overlap candidates.** none observed among the in-scope batch ids.

---

### t237

**Premise (one sentence).** `moai gate`'s pre-commit hook's `go vet` invocation resolves Go
packages relative to the repository root rather than each staged file's owning module, so any
monorepo with a non-root `go.mod` (e.g. `apps/id`) has every Go commit blocked by a package-resolution
failure rather than a real vet finding.

**Premise verdict.** `holds` — the exact defect shape the card names is present verbatim in the
current tree.

**Landing verdict.** `not-landed`
- commit: — (no commit's message matches `\bt237\b`; the only hits are t230's SPEC narrative
  *mentioning* t237/#1641 as "the card about to change the hook body", which is a reference, not a
  delivery — per the worker instructions' explicit "mention is not a landing" caution)
- pinned ref: 48239c7dc7428c8751a04f6321887c2d36123884 (mention found there and in develop; not a
  delivering commit either way)
- `--is-ancestor` exit: — (not applicable; no delivering commit to test)

**Claim.** `preCommitHookContent`'s vet step still computes `./$(dirname "$f")` against the process
cwd (repo root) with no upward `go.mod` search, exactly the defect the card describes; the
independently-referenced verified patch (`t312-precommit-vet @ b6f478b1a`) is not merged into this
tree.

**Evidence.**
```
$ grep -n "go vet\|dirname\|go.mod" internal/cli/hook_install_precommit.go
77: printf './%s\n' "$(dirname "$f")"
90: if ! go vet $BT_TAGS $PKGS >/dev/null 2>&1; then

$ git log ee50984abe4f11ac337382b48a26328f091e200a --oneline -- internal/cli/hook_install_precommit.go
db1362739 feat(cli): back up and disclose user-modified pre-push hooks (t257) (#1650)
32d2221fa feat(cli): back up and disclose user-modified pre-commit hooks (t230) (#1647)
883d53852 feat(SPEC-PRETOOL-GATE-MOVE-001): relocate commit-quality gate ... (#1189)
a596d9e41 fix(hook): 커밋 게이트 goolm 빌드태그 주입 ...
52b5e4bf5 feat(SPEC-PRECOMMIT-001): 배포 계층 pre-commit 훅 설치기 추가 (fast-subset gofmt+vet)
# no commit touches module-relative package resolution
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`,
`internal/cli/hook_install_precommit.go`, and the pinned-develop file history for that path.

**Gaps.** Did not check the paired template twin
(`internal/template/templates/.git_hooks/pre-commit`) for byte-identity — the card names this as a
hard-required twin edit, but since neither side has changed, drift-checking is moot here. Did not
attempt local reproduction (this repo's own module is root-level, as the card itself notes it cannot
reproduce here).

**Residual-risk.** None beyond what the card already states — the referenced verified patch exists
on an unmerged worktree (`t312-precommit-vet`) and this sweep did not audit that patch's quality,
only confirmed the defect it targets is still live here.

**Proposed disposition.** `keep` — premise holds, not landed, patch reportedly exists but unmerged.

**Overlap candidates.** t239 (explicitly cross-referenced by its own text as "같은 결함형, 다른
표면" — the same silent-overwrite/오버라이트 problem class as this card's monorepo vet-resolution
defect, both rooted in the t230 SPEC-PRECOMMIT-PRESERVE-001 lineage).

---

### t239

**Premise (one sentence).** `moai update` redeploys `.moai/config` wholesale, so a user who hand-edits
`llm.yaml`'s `audit.codex`/`audit.glm` sections (added by t225) has that edit silently reverted on
the next update, the same defect class as t230's pre-commit-hook overwrite but a different surface.

**Premise verdict.** `unverified` — the structural mechanism is confirmed but the specific
`audit.codex`/`audit.glm` sub-key claim could not be located, so the premise **as worded** is
undecided. (Orchestrator normalization, M3 post-check: the worker wrote a compound verdict here;
AC-BH-011 admits exactly one token, and a premise whose named subject cannot be found is undecided
rather than holding. The substance below is the worker's, unchanged.) Confirmed: `internal/cli/update/deploy/deploy.go`'s
`CleanMoaiManagedPaths` deletes `.moai/config/` entirely before redeployment (comment: "Clean
.moai/config/ entirely - backup was already done by the Backup step"), and its own doc comment
states plainly: "Template-managed files are NOT backed up: deployment rewrites them moments later,
so their only copy is never at stake" — which is precisely the false assumption the card flags for a
file like `llm.yaml` whose per-user customization IS the thing at stake. However, grepping the
current template and local `llm.yaml` for an `audit:` key found none — I could not confirm the
specific `audit.codex`/`audit.glm` sub-keys the card describes currently exist under that name (the
audit-model-pin config may have moved to a different section/struct since t225 landed, or my grep
scope was too narrow for the bounded budget here).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** The general defect mechanism (wholesale `.moai/config` wipe, template-managed files
excluded from the pre-clean backup on the stated "never at stake" assumption) is confirmed current
and applies to any template-shipped, user-customizable config file, `llm.yaml` included in
principle.

**Evidence.**
```
$ sed -n '99,105p' internal/cli/update/deploy/deploy.go
// destruction it was taken to survive. Template-managed files are NOT backed
// up: deployment rewrites them moments later, so their only copy is never at
// stake, and skipping them keeps the backup to what would otherwise be lost

$ sed -n '187,190p' internal/cli/update/deploy/deploy.go
	# Clean .moai/config/ entirely - backup was already done by the Backup step.

$ grep -n "audit:" internal/template/templates/.moai/config/sections/llm.yaml
(no matches)
$ grep -n "audit:" .moai/config/sections/llm.yaml
(no matches)
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`,
`internal/cli/update/deploy/deploy.go`, `internal/template/templates/.moai/config/sections/llm.yaml`,
and this worktree's own `.moai/config/sections/llm.yaml`.

**Gaps.** Did not locate where t225's audit-model-pin values actually live in the current schema
(possibly `internal/config/audit_models.go`'s `AuditModel` struct backed by a different YAML
section/file than `llm.yaml`) — a targeted `Grep` for "audit.codex"/"audit.glm" as literal YAML
paths across `.moai/config/sections/*.yaml` was not run within the bounded per-card budget.

**Residual-risk.** If the audit-pin values in fact live in a config file this run did not check
(e.g., `workflow.yaml`), the specific file name in the card (`llm.yaml`) could be stale even though
the general mechanism is real — the card's proposed disposition would then need to be revised to
name the correct file rather than dropped.

**Proposed disposition.** `keep` — structural mechanism confirmed; recommend the operator re-verify
which file currently carries the audit-pin values before scoping a fix.

**Overlap candidates.** t237 (self-cross-referenced: "t230(pre-commit 무음 덮어쓰기)와 같은
결함형, 다른 표면").

---

### t240

**Premise (one sentence).** The §H overlay doctrine documents that z.ai's `thinking` field carries
reasoning effort for the Anthropic-compatible shim, but live measurement during t225 found the
opposite — top-level `reasoning_effort` is what's actually honored — so the doc needs correcting and
AC-MTP-032b's UNVERIFIED marker needs resolving.

**Premise verdict.** `falsified` — the correction and the marker resolution this card asks for
appear to have already happened, inside SPEC-V3R6-AUDIT-MODEL-PIN-001 (status: `implemented`).
`.moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/acceptance.md` AC-AMP-006 is explicitly headed "`[MUST —
closes AC-MTP-032b]`" and its recorded amendment states: "Delivery was PROVEN by hypothesis B
(top-level `reasoning_effort`) against the hypothesis-A null (1.02)" — i.e., the same reversed
finding the card describes (`thinking` field is the null/ignored one; `reasoning_effort` is the
live one) is already measured and recorded, with the AC that specifically closes the UNVERIFIED
marker the card names.

**Landing verdict.** `not-landed`
- commit: — (no commit message matches `\bt240\b`; this finding is attached to SPEC
  V3R6-AUDIT-MODEL-PIN-001 / card t225, not t240 by commit-message convention)
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** The measurement reversal and the AC-MTP-032b closure the card asks for exist in the
SPEC-V3R6-AUDIT-MODEL-PIN-001 acceptance criteria already; whatever remains is narrower than the
card states (I could not locate a "§H 오버레이 문서" file by that heading to confirm or refute
whether ITS specific prose was ever updated — see Gaps).

**Evidence.**
```
$ grep -n "032b\|UNVERIFIED" .moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/acceptance.md
70: ### AC-AMP-006 — live GLM reasoning-delivery proof, numeric rule (REQ-AMP-006, REQ-AMP-007) [MUST — closes AC-MTP-032b]

$ grep -n "hypothesis" .moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/acceptance.md
85: delivery from noise (measured null: the hypothesis-A thinking-budget run
100: (output-token ratio 1.40, consistent). Delivery was PROVEN by hypothesis B
101: (top-level `reasoning_effort`) against the hypothesis-A null (1.02) — the

$ grep -n "^status:" .moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/spec.md
5: status: implemented

$ grep -rln "§H " .claude/
.claude/agents/moai/manager-spec.md
.claude/rules/moai/development/spec-frontmatter-schema.md
# neither file's §H concerns reasoning_effort delivery — the card's "§H 오버레이 문서" target
# was not located in this repo by heading search
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`,
`.moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/{spec,acceptance}.md`, and a repo-wide grep for a "§H"
heading.

**Gaps.** Did not locate the specific "§H 오버레이 문서" the card names (it may live in
`.moai/reports/t225/sync-audit-review-2.md`, the card's own cited source, rather than in the rules
tree — that report file was not read in this run). Did not confirm whether the two "비차단 R1/R2"
items (insertAuditLeaf comment order, default step=2 unreachable) were addressed.

**Residual-risk.** If the actual overlay doctrine file the card means is a separate, still-unedited
document (distinct from the SPEC's own acceptance criteria), the card's core ask — editing THAT
file — could still be outstanding even though the underlying measurement and the AC marker are
resolved.

**Proposed disposition.** `needs-operator-decision` — the measurement/marker half is resolved;
whether a distinct doctrine file still needs a matching edit needs the operator (or a follow-up read
of `.moai/reports/t225/sync-audit-review-2.md`) to confirm.

**Overlap candidates.** none observed among the in-scope batch ids.

---

### t242

**Premise (one sentence).** The origin-trail chain-node-creation path (`CreateNodeAtSpawn`) has no
production caller and `MOAI_CHAIN_NODE_ID` is never set anywhere, so `events.jsonl` has never been
populated and the chain-event hook wiring is a permanent no-op — requiring a decision on whether
this is a bug to fix or an intentionally-incomplete Phase 1 to retire.

**Premise verdict.** `holds` — confirmed by a full-repo grep for both the function and the env-var
constant.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** `CreateNodeAtSpawn` (defined `internal/chain/populate.go:53`) has zero non-test callers
anywhere in the Go tree; `config.EnvChainNodeID` (`MOAI_CHAIN_NODE_ID`) is only ever *read*
(`os.Getenv`) in production code — no production call sets it via `os.Setenv` or otherwise. The gap
is real and total, not partial.

**Evidence.**
```
$ grep -rn "CreateNodeAtSpawn" --include="*.go" . | grep -v _test
internal/chain/populate.go:42: // CreateNodeAtSpawn creates a skeleton node-enter event at a spawn boundary.
internal/chain/populate.go:53: func (p *Populator) CreateNodeAtSpawn(worktreePath, specID, milestone string) (string, error) {
# only the definition itself — no call site

$ grep -rn "EnvChainNodeID" --include="*.go" . | grep -v _test
internal/config/envkeys.go:273: // EnvChainNodeID carries the origin-trail chain node ID from the spawning
internal/config/envkeys.go:279: EnvChainNodeID = "MOAI_CHAIN_NODE_ID"
internal/chain/populate.go:54: parentID := os.Getenv(config.EnvChainNodeID)
internal/chain/populate.go:155: if envID := os.Getenv(config.EnvChainNodeID); envID != "" {
internal/hook/chain_banner.go:78: envNodeID := os.Getenv(config.EnvChainNodeID)
# three reads, zero writes
```

**Baseline-attribution.** Full-repository `grep -rn --include="*.go" .` from worktree root
`6165f9f5e`, both queries.

**Gaps.** Did not read the SPEC-CHAIN-CORE-001 plan/spec to determine whether Phase 1 explicitly
scoped node-creation wiring out (the card frames this exact question as open and unresolved — I did
not attempt to resolve it, per the card's own framing that a human/owner decision is what's needed).

**Residual-risk.** None beyond the open ownership question the card itself names.

**Proposed disposition.** `keep` — premise holds strongly; this is a `needs-operator-decision`-shaped
card by its own design (asks for a judgment, not a fix), so `keep` as-is is the right disposition for
the sweep to propose, letting the operator make the bug-vs-intentional call.

**Overlap candidates.** t216 (source investigation this card was split from), t243, t244 (siblings
split from the same t216 lane-6 investigation, `d1-chain-event.md`/`d2-unwired-scripts.md`).

---

### t243

**Premise (one sentence).** `handle-session-start-navigator.sh` was deleted in a build-recovery
commit alongside sibling hooks, but only the siblings were restored two commits later, leaving this
one hook permanently missing and requiring a decision to restore or retire it.

**Premise verdict.** `falsified` — the file exists, at both the local and template locations, and
git history for this exact path shows exactly ONE commit ever touching it (its creation) — no
deletion, no restoration, contradicting the "deleted then left behind" claim outright.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** The file is present now and was never removed in this repository's history; the card's
premise (as measured against a *different* worktree — t216/lane-6, per its own citation) does not
hold against this tree.

**Evidence.**
```
$ find .claude/hooks/moai internal/template/templates/.claude/hooks/moai -iname "*navigator*"
.claude/hooks/moai/handle-session-start-navigator.sh
internal/template/templates/.claude/hooks/moai/handle-session-start-navigator.sh

$ git log --oneline -- .claude/hooks/moai/handle-session-start-navigator.sh
2c87d195f feat(SPEC-PROJECT-NAVIGATOR-001): Project Navigator — living nav + --brief + shared read primitive (#1354)
# single commit — file was never deleted in this repo's history

$ grep -n "navigator" .claude/settings.json
(no matches)
# confirms it is currently unwired (matches the card's OTHER premise — d2-unwired-scripts.md's
# "11 unwired, 4 truly dead" — but not the "was deleted and orphaned" premise)
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`, full `git log --oneline` for
the exact path (no `--follow` needed since there is only one entry regardless), and
`.claude/settings.json`.

**Gaps.** Did not check whether the card's cited investigation (`.moai/reports/t216/...
d2-unwired-scripts.md`, on a different worktree at a different point in time) was itself measuring
a genuinely different repository state that has since self-corrected (e.g., another card restoring
the file before this sweep ran) — the git log here shows no restoration commit, which argues against
that, but the investigation report itself was not re-read to reconcile the discrepancy.

**Residual-risk.** The card's *decision* framing ("restore vs retire") is moot if the file was never
actually deleted — but its still-unwired status (confirmed above) is a live, separate finding that a
revised card might still want to carry forward.

**Proposed disposition.** `already-landed` — the specific claimed defect (file missing, sibling
hooks restored, this one orphaned) does not describe the current tree; the file is present. Rests on
the single-commit `git log` result above. The still-open "restore vs retire the WIRING" question
(unwired, not missing) may warrant a differently-worded follow-up card — operator's call.

**Overlap candidates.** t216, t242, t244 (siblings split from the same t216 lane-6 investigation).

---

### t244

**Premise (one sentence).** `team-ac-verify.sh` is currently unwired (no hook registration) and is
one of the dormant hooks the t216 investigation found needing an explicit decision — wire it into
the harness-thorough + team-mode gate it was designed for, or retire it.

**Premise verdict.** `holds` — confirmed both that the file exists (locally and in template) and
that it carries no entry in `.claude/settings.json`'s hook wiring, matching the card's "currently
unwired" claim exactly.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** The dormant/unwired state the card describes is current; no decision (wire or retire) has
been made in this tree.

**Evidence.**
```
$ find .claude/hooks internal/template/templates/.claude/hooks -iname "*team-ac-verify*"
.claude/hooks/moai/team-ac-verify.sh
internal/template/templates/.claude/hooks/moai/team-ac-verify.sh

$ grep -n "team-ac-verify" .claude/settings.json internal/template/templates/.claude/settings.json.tmpl
(no matches in either file)
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`, both hook-directory trees and
both settings.json variants (local rendered + template source).

**Gaps.** Did not read the hook's own script body to independently confirm the "harness thorough +
team 전제" preconditions it references (took the card's framing of the trigger conditions at face
value, since the wiring-absence finding alone is sufficient to confirm the "currently unwired"
premise regardless of what its trigger conditions would be if wired).

**Residual-risk.** None beyond the open decision the card itself frames as needed.

**Proposed disposition.** `keep` — premise holds; this is a `needs-operator-decision`-shaped card,
same pattern as t242/t243.

**Overlap candidates.** t216, t242, t243 (siblings split from the same t216 lane-6 investigation;
also referenced together in `agent-common-protocol.md`'s own "Hook Invocation Surface" doctrine as
one of the three mechanically-enforcing hook scripts alongside `sync-phase-quality-gate.sh` and
`status-transition-ownership.sh`, though that doctrine describes it as "dormant" without resolving
wire-vs-retire either).

---

### t247

**Premise (one sentence).** PR #1600 (`feat(mcp): make a running MCP server's build info
observable`, branch `WT-server-version`) is unreviewable at 497 changed files (CodeRabbit skipped
review outright) and `CONFLICTING`, and needs to be split into review-sized, non-generated-vs-logic
separated PRs before it can land.

**Premise verdict.** `falsified` — PR #1600 is not in the state the card describes. It is currently
`MERGED`, with only 10 changed files (597 additions / 7 deletions), on the exact branch name the
card cites (`WT-server-version`), merged 2026-08-25T03:23:18Z via merge commit
`07a4ea0ed73f79aa45f3011cb53d278f2ee7ae8e` — nowhere near the 497-file, CONFLICTING state the card
was written against. The underlying work evidently was split/rebased down to a mergeable size before
merging, whether or not that split was driven by this card.

**Landing verdict.** `not-landed`
- commit: — (no commit message in either pinned ref matches `\bt247\b`; the resolution happened via
  the PR itself, not via a card-attributed commit)
- pinned ref: —
- `--is-ancestor` exit: —
- (Supplementary, not a `landed` grep-hit): merge commit `07a4ea0ed73f79aa45f3011cb53d278f2ee7ae8e`
  IS an ancestor of pinned main `48239c7dc7428c8751a04f6321887c2d36123884`
  (`git merge-base --is-ancestor` exit 0), confirming the small, merged PR is genuinely in the tree
  this sweep is measuring against — not a stale `gh` cache artifact.

**Claim.** The specific blocking condition the card names (497 files, review-skipped, CONFLICTING)
no longer describes PR #1600's actual state; the card's request has been satisfied by however this
PR ended up at 10 files.

**Evidence.**
```
$ gh pr view 1600 --json state,mergeable,additions,deletions,changedFiles,title
{"additions":597,"changedFiles":10,"deletions":7,"mergeable":"UNKNOWN","state":"MERGED","title":"feat(mcp): make a running MCP server's build version visible"}

$ gh pr view 1600 --json headRefName,mergedAt,mergeCommit
{"headRefName":"WT-server-version","mergedAt":"2026-08-25T03:23:18Z","mergeCommit":{"oid":"07a4ea0ed73f79aa45f3011cb53d278f2ee7ae8e"}}

$ git merge-base --is-ancestor 07a4ea0ed73f79aa45f3011cb53d278f2ee7ae8e 48239c7dc7428c8751a04f6321887c2d36123884 && echo YES_ANCESTOR || echo NOT_ANCESTOR
YES_ANCESTOR
```

**Baseline-attribution.** `gh pr view` measured live against GitHub at the time of this run (not a
tree-scoped measurement); the ancestry check measured against worktree HEAD `6165f9f5e`'s view of
pinned main `48239c7dc7428c8751a04f6321887c2d36123884`.

**Gaps.** Did not verify whether the 497-files/CONFLICTING state the card describes ever actually
existed for PR #1600 at some earlier point (i.e., whether the card's own dated measurement,
"리드 2026-08-24", was accurate at the time) — only that it does not describe the PR's current,
merged state. Did not check whether the "생성물 재생성 커밋과 로직 변경을 분리" quality bar the card
sets was actually met by however this PR was structured before merging (10 files could still mix
generated + hand-written changes).

**Residual-risk.** If the 10-file merged PR did NOT actually separate generated-artifact commits
from logic commits (the card's specific quality bar, and its named "mutant" to watch for), the
underlying review-quality complaint could still be valid even though the file-count/CONFLICTING
crisis is resolved.

**Proposed disposition.** `already-landed` — rests on the `gh pr view` state above (MERGED, 10
files, ancestor of pinned main).

**Overlap candidates.** none observed among the in-scope batch ids.

---

### t248

**Premise (one sentence).** MCP `audit_multi`/`codex_audit`/`glm_audit` judgment output does not
record which commit of the `moai` binary actually served the audit, so a stale-binary judgment (as
happened in the t229 investigation, 259 commits behind) is indistinguishable after the fact from a
current one.

**Premise verdict.** `holds`, on a bounded check. Grepped the primary audit-multi handler file for
any version/commit-recording field in the tool's registration or its handler and found none.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** No commit-SHA (or build-version) field was found being attached to the `audit_multi`
tool's output in the handler file that composes it.

**Evidence.**
```
$ grep -n "pkg/version\|BuildVersion\|version\.\|Commit\b" internal/cli/mcp_audit_multi.go
(no matches)
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`,
`internal/cli/mcp_audit_multi.go` only.

**Gaps.** Did not check `codex_audit`'s and `glm_audit`'s own separate handler files (only
`audit_multi`'s composing file was grepped, per the bounded-depth restraint) — the card names all
three tools, and this run verified only one of the three surfaces. Did not check whether the
persisted audit report files themselves (under `.moai/reports/`) carry a commit field written by a
layer outside the MCP handler (e.g., the calling agent stamping it separately) — that would
partially satisfy the card's intent through a different mechanism than the one grepped here.

**Residual-risk.** If `codex_audit`/`glm_audit` or the persisted-report layer already records a
commit SHA by some other path not covered by this grep, the premise as measured here would overstate
the gap; this run's evidence supports the `audit_multi` composing-handler surface specifically, not
an exhaustive claim across all three tools and the persistence layer.

**Proposed disposition.** `needs-operator-decision` — narrow single-file check supports `keep`, but
the gaps above (2 of 3 named tools + the persistence layer unchecked) mean a fuller read is needed
before committing to scope.

**Overlap candidates.** none observed among the in-scope batch ids (card's own text names t246 as
a related card on "감사가 다른 트리를 읽는 축", not in this batch's in-scope list).
