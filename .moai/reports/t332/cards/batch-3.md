# t332 card sweep — batch 3

Cards: t252 t253 t254 t255 t260 t262 t263 t264 t280 t281 t284  (11 entries)
Worktree HEAD: 6165f9f5e
Pinned develop: ee50984abe4f11ac337382b48a26328f091e200a
Pinned main:    48239c7dc7428c8751a04f6321887c2d36123884

### t252

**Premise (one sentence).** At least one SPEC (SPEC-V3R6-AUDIT-MODEL-PIN-001, t225) is stuck in `implemented` status with a `pending-backfill-sync` §E.4 `sync_commit_sha`, waiting to be batched into a single frontmatter-transition PR once ≥2 such items accumulate.

**Premise verdict.** `holds` — the cited SPEC is still exactly in the state the card describes.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t252 in `00-worktree-list.txt`)

**Claim.** The card's target SPEC (SPEC-V3R6-AUDIT-MODEL-PIN-001) still carries `status: implemented` (not `completed`) and `sync_commit_sha: "pending-backfill-sync"` at worktree HEAD; the batch-backfill action the card describes has not been performed for this item.

**Evidence.**
```
$ sed -n '1,7p' .moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/spec.md
---
id: SPEC-V3R6-AUDIT-MODEL-PIN-001
title: "Pin cross-model audit backend model+effort via the workflow.audit config block"
version: 1.1.0
status: implemented
created: 2026-08-24
updated: 2026-08-24

$ grep -n "sync_commit_sha\|run_commit_sha" .moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/progress.md
484:run_commit_sha: "a7c5c3833"   # backfilled (M5 commit) — placeholder-per-D3 pattern
526:sync_commit_sha: "pending-backfill-sync"   # D3 self-reference exemption — a commit cannot know its own SHA; backfilled after the PR merges (lead owns the implemented → completed transition + this backfill)
551:- `sync_commit_sha` is a pending-backfill placeholder per the D3 exemption — resolves only after the commit lands (lead's post-merge step).

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt252\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt252\b' --oneline
(no output)
```

**Baseline-attribution.** `spec.md` and `progress.md` read at worktree HEAD (`6165f9f5e`, tracking pinned develop `ee50984a`). Landing queries run against the two pinned refs.

**Gaps.** Did not check whether any *other* SPEC besides SPEC-V3R6-AUDIT-MODEL-PIN-001 has since joined the "pending item" pool the card describes (the card is explicitly a running list; only the one named item was verified). Did not check the exact `run_commit_sha` discrepancy (card cites `8d60fb5e4`; progress.md shows `a7c5c3833`) — could be a different commit in a since-superseded state, not verified further (out of the card's central premise).

**Residual-risk.** If a second qualifying SPEC has since appeared, the card's own launch condition ("2+ 건 모였을 때") may already be satisfied, which this sweep did not check for.

**Proposed disposition.** `keep` — premise holds, card is well-formed and still actionable; rests on the spec.md/progress.md read above.

**Overlap candidates.** none observed among in-scope ids — the SPEC-closure-batch mechanism this card describes is not clearly touched by any other in-scope card's text.

---

### t253

**Premise (one sentence).** `internal/sessionmsg/store.go`'s per-recipient pending mailbox has no depth ceiling, so an unconsumed recipient's mailbox can grow without bound.

**Premise verdict.** `holds` — `Store.Send` (store.go, `func (s *Store) Send`) writes unconditionally into `s.pendingDir(toAgentID)` via `writeJSONAtomic`, with no check against an existing pending-count before the write.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t253)

**Claim.** No depth/size cap exists on the pending mailbox at HEAD; the send path in `internal/sessionmsg/store.go` enqueues every validated envelope regardless of how many are already pending for that recipient.

**Evidence.**
```
$ grep -n "pending" internal/sessionmsg/store.go | head -20
103:func (s *Store) pendingDir(id string) string {
...
263:		return writeJSONAtomic(filepath.Join(s.pendingDir(toAgentID), msgID+".json"), env)

$ sed -n '206,266p' internal/sessionmsg/store.go
(func Send — validates from/to agent ids, sender/receiver existence, message shape via env.Validate(),
then unconditionally: s.withAgentLock(toAgentID, func() error {
    return writeJSONAtomic(filepath.Join(s.pendingDir(toAgentID), msgID+".json"), env)
}) — no pending-count read or ceiling check anywhere in this path)
```

**Baseline-attribution.** `internal/sessionmsg/store.go` read at worktree HEAD (`6165f9f5e`).

**Gaps.** Did not check `Poll`'s sweep logic (store.go:280-380) for any indirect throttling (e.g. a claimed-batch ceiling that might incidentally bound growth) — read only enough to confirm `Send` itself has no gate. Did not check whether a size/TTL-based passive cleanup exists elsewhere (e.g. expired-envelope sweep on `Poll`) that would bound *effective* growth even without an explicit cap.

**Residual-risk.** If TTL-based expiry (mentioned in the file's own comments, `ExpiresAt`) effectively bounds worst-case growth in practice, the "unbounded" framing may overstate risk somewhat — the card's own scope item (1) ("근거를 먼저 세울 것") anticipates exactly this kind of measurement gap.

**Proposed disposition.** `keep` — premise holds on direct code read; the card correctly identifies an unbounded-write path.

**Overlap candidates.** t254 (same delivering SPEC, SPEC-CODEX-SESSION-MSG-001, different artifact — spec.md/research.md vs store.go); t262 (uses SPEC-CODEX-SESSION-MSG-001 as its own worked example of a template-neutrality-guard gap).

---

### t254

**Premise (one sentence).** `research.md:32` and `spec.md:55` of SPEC-CODEX-SESSION-MSG-001 each contain a backslash-escaped grep-alternation pipe inside a GFM table cell whose recorded "0-hit" verification is unfalsifiable because GFM silently unescapes `\|` when rendering table cells.

**Premise verdict.** `falsified` — neither cited line currently has the hazard the card describes. `research.md:32` is a table cell but already uses the `-e "a" -e "b"` repeated-flag form (no `\|` present to unescape). `spec.md:55` does contain `"session_msg\|session-msg"`, but the line is **not** inside a table — it is a standalone prose paragraph — and GFM does not process backslash escapes inside inline code spans at all (per this project's own memory note `feedback-gfm-unescapes-pipes-in-table-cells`: "this affects table cells only. A plain `|` inside inline code in a bullet list or paragraph needs no escaping and renders correctly as-is").

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t254)

**Claim.** The specific defect the card names is not present at either cited location at worktree HEAD; `research.md:32` was already written in the hazard-avoiding form, and `spec.md:55` was never in a table cell in the first place.

**Evidence.**
```
$ awk 'NR==32{print NR": "$0}' .moai/specs/SPEC-CODEX-SESSION-MSG-001/research.md
32: | 부재 확인 | `grep -rn -e "session_msg" -e "session-msg" internal/ .moai/specs/` → 구현 0건 / SPEC 충돌 0건 (2026-08-23) | 신규 네임스페이스 무충돌 |

$ awk 'NR==55{print NR": "$0}' .moai/specs/SPEC-CODEX-SESSION-MSG-001/spec.md
55: `grep -rn "session_msg\|session-msg" internal/ .moai/specs/` (2026-08-23, 본 워크트리) — 구현 히트 0건, 기존 SPEC 충돌 0건. ...

$ sed -n '48,56p' .moai/specs/SPEC-CODEX-SESSION-MSG-001/spec.md
| 원자적 파일 쓰기 | ... |
| codex_task 패밀리 (위임) | ... |
| C-HRA-008 정적 가드 선례 | ... |
| 임계값 단일 원천 | ... |
                                    <- blank line, table ends here
### §A.3 부재 확인 (이 SPEC이 만들 것)
                                    <- blank line
`grep -rn "session_msg\|session-msg" internal/ .moai/specs/` (2026-08-23, ...)   <- line 55, standalone paragraph, not a table row
```

**Baseline-attribution.** Both files read at worktree HEAD (`6165f9f5e`). Line numbers matched exactly to the card's citation (32 and 55).

**Gaps.** Did not run the file through an actual GFM renderer (`gh api markdown`) to double-confirm code-span backslash preservation — relied on CommonMark spec behavior and this project's own recorded memory note, both of which agree code spans are unaffected. Did not re-verify the underlying "0-hit" grep claims themselves (not the card's premise; the card is about escape-safety of the *recorded* verification, not the verification's current truth).

**Residual-risk.** If some other renderer (not GitHub's GFM) processes backslashes inside code spans differently, the spec.md:55 instance could still be at risk in that renderer — this sweep verified GFM behavior only, consistent with the card's own framing ("GFM 표 셀 안에서는").

**Proposed disposition.** `drop` — the specific defect claimed does not exist at either cited location; rests on the line-55/line-32 reads above plus the project's own prior memory finding on code-span vs table-cell escape behavior.

**Overlap candidates.** t253 (same SPEC dir, different artifact); t262 (same SPEC used as the guard's worked example).

---

### t255

**Premise (one sentence).** A `.git/hooks/pre-commit.local` delegation extension point (deliberately deferred out of t230's scope, per t230's own spec.md §D Out of Scope) remains unimplemented and is a valid follow-up now that t230 has landed and established the provenance-recording precondition it depends on.

**Premise verdict.** `holds` — t230 has landed (backup+disclose for pre-commit hooks), and the codebase explicitly still does NOT implement `pre-commit.local` delegation; a test asserts the disclosure notice must not even name it.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t255 itself)

**Claim.** t255's precondition (t230 landed) is satisfied, and its target gap (pre-commit.local delegation) is confirmed still absent and deliberately excluded, matching the card's own account of why it was split out of t230.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt230\b' --oneline
6786c3fa4 t250: graph freshness, symbol layer, and MCP code queries (SPEC-V3R6-GRAPH-FRESHNESS-001) (#1648)
db1362739 feat(cli): back up and disclose user-modified pre-push hooks (t257) (#1650)
539349c5b docs(t230): t230 sync-audit evidence — SPEC-PRECOMMIT-PRESERVE-001 PASS 95/100 (#1649)
32d2221fa feat(cli): back up and disclose user-modified pre-commit hooks (t230) (#1647)

$ grep -rn "pre-commit.local" internal/cli/hook_install_precommit.go internal/cli/hook_install_precommit_disclosure_test.go
internal/cli/hook_install_precommit.go:255:				// not name pre-commit.local, a facility this SPEC does not ship.
internal/cli/hook_install_precommit_disclosure_test.go:293:	if strings.Contains(warn.String(), "pre-commit.local") {
internal/cli/hook_install_precommit_disclosure_test.go:294:		t.Errorf("the notice must not name pre-commit.local — a recovery path the installed hook never reads (REQ-PCP-004)")

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt255\b' --oneline
(no output)
```

**Baseline-attribution.** `internal/cli/hook_install_precommit.go` and its disclosure test read at worktree HEAD (`6165f9f5e`); t230 landing checked against pinned develop `ee50984a`.

**Gaps.** Did not verify t230's `--is-ancestor` exit code explicitly (not required — t255 itself, not t230, is this card's landing target, and t230 is out of the in-scope batch). Did not check t237/#1641 (named by the card as sharing the same release-level constraint) for its current state.

**Residual-risk.** None specific beyond the card's own stated risk (bundling this with a release that also ships the provenance discriminator would trip every install base's first-upgrade backup+warn) — this sweep did not need to re-litigate that reasoning, only confirm the precondition and the gap.

**Proposed disposition.** `keep` — premise holds; well-formed, correctly gated follow-up.

**Overlap candidates.** t237 (named by the card as sharing the identical release-level sequencing constraint — REQ-PCP-015).

---

### t260

**Premise (one sentence).** The lessons-inbox auto-collection channel captures only tool-call failures and therefore structurally cannot record the more expensive defect class this repository actually catches most often — checks that pass while observing nothing.

**Premise verdict.** `unverified` — the card's specific composition numbers ("최근 400행 중 390행이 tool_failure:Bash:UnknownFailure") could not be re-measured: `.moai/lessons-inbox.jsonl` does not exist in this worktree (it is local/gitignored runtime state, consistent with CLAUDE.local.md's Local-Only Files list). The structural claim (collection is tool-failure-triggered only) is corroborated by the hook wiring but the numeric composition claim itself is unverified here.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t260)

**Claim.** The card is fundamentally a decision-request ("무엇이 실제로 학습되는지를 정직하게 정하는 것"), not a mechanical-defect report; its structural premise (collection is tool-failure-only) is consistent with the codebase's hook wiring, but the specific inbox-composition measurement is stale/unverifiable in this worktree.

**Evidence.**
```
$ ls -la .moai/lessons-inbox.jsonl
ls: .moai/lessons-inbox.jsonl: No such file or directory

$ grep -rln "lessons-inbox" internal/ pkg/ cmd/ | grep -v _test.go
internal/hook/failure_observer.go
(+ 4 template/skill-doc references, not code)

$ grep -n "func \|lessons-inbox" internal/hook/failure_observer.go | head -8
46:func recordToolFailureEvent(input *HookInput, category ErrorCategory) {
87:func recordTestFailEvent(input *HookInput, pkg string) {
129:func appendLessonsInboxStub(root, eventKey, summary, source string) {
```

**Baseline-attribution.** `internal/hook/failure_observer.go` read at worktree HEAD (`6165f9f5e`). The `.moai/lessons-inbox.jsonl` absence check is scoped to this worktree only, at the time of this sweep.

**Gaps.** Did not check `recordTestFailEvent` for whether it captures a broader class than pure tool-call failure (e.g. whether a failed `go test` run itself gets logged, which would be closer to but still distinct from the "check passed while observing nothing" class the card names). Did not check the primary checkout's or lead machine's actual inbox file (out of scope for a worktree-isolated read-only sweep).

**Residual-risk.** If `recordTestFailEvent` or another handler already captures a wider class than "Bash tool_failure", the card's "구조적으로 못 잡는다" framing could be narrower than stated — this sweep did not fully enumerate every event type `failure_observer.go` handles.

**Proposed disposition.** `needs-operator-decision` — the card explicitly asks for a decision on channel design/scope, not a mechanical fix; the structural premise is plausible but the specific numbers are unverified here.

**Overlap candidates.** t280 (same underlying file, `.moai/lessons-inbox.jsonl` — t260 is about collection *scope*, t280 is about drain *deployment*; both share the mechanism).

---

### t262

**Premise (one sentence).** The template neutrality guard's SPEC-ID detection pattern only matches a fixed prefix allowlist (`V3R2-6`, `AGENCY`, `WORKTREE`), which is narrower than the doctrine's blanket prohibition on any SPEC ID in template content, so guard-pass does not imply doctrine compliance.

**Premise verdict.** `holds` — confirmed directly in the guard's source.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t262)

**Claim.** `internal/template/internal_content_leak_test.go`'s C1 pattern is a fixed alternation over a small prefix set; a SPEC ID with any other prefix (the card's own example, `SPEC-CODEX-SESSION-MSG-001`) does not match and would pass the guard undetected.

**Evidence.**
```
$ grep -n "SPEC-V3R6\|regexp.MustCompile" internal/template/internal_content_leak_test.go | grep -i spec
171:		pattern: regexp.MustCompile(`\bSPEC-(V3R[2-6]|AGENCY|WORKTREE)-[A-Z0-9-]+\b`),
207:		pattern:         regexp.MustCompile(`\bSPEC-V3R[0-9]-[A-Z0-9-]+\b|\bCONST-V3R[0-9]-[0-9]+\b|\bSPEC-WF-AUDIT-GATE-001\b|\bSPEC-MX-001\b`),
```

`SPEC-CODEX-SESSION-MSG-001` matches neither alternation (prefix `CODEX` is not `V3R[2-6]`, `AGENCY`, or `WORKTREE`, nor any of the other literal SPEC-ID alternatives at line 207).

**Baseline-attribution.** `internal/template/internal_content_leak_test.go` read at worktree HEAD (`6165f9f5e`).

**Gaps.** Did not enumerate every pattern in the file (10 spec-ID-adjacent entries total per an earlier grep) to build the full "what the guard currently matches" vs "what doctrine forbids" difference set the card's scope item (2) asks for — that is a larger task than this sweep's restraint budget allows; confirmed only that the gap exists, not its full extent.

**Residual-risk.** A broadened general-form pattern (`SPEC-<DOMAIN>-NNN`) risks false positives on prose mentions, exactly as the card's own "대표 mutant" note anticipates — the fix itself is nontrivial, not just the detection.

**Proposed disposition.** `keep` — premise holds on direct source inspection; genuine gap.

**Overlap candidates.** t253, t254 (both use SPEC-CODEX-SESSION-MSG-001, the card's own worked example, as their subject).

---

### t263

**Premise (one sentence).** `file_changed.go:110-118`'s incremental MX-scan path still has the same "die when the sidecar index is missing" defect that t216 M4 already fixed in the `mx query` cold-start path, so M4's self-build fix should be reused there rather than reinvented.

**Premise verdict.** `falsified` — the premise's own precondition is false: t216 has **not landed** (no commit for `\bt216\b` found against either pinned ref), and `mx_query.go`'s cold-start path at HEAD still exhibits the exact pre-fix die behavior the card describes — it is not merely unrepaired in the incremental path, it is *also* still unrepaired in the very path the card cites as already-fixed. A passing test (`TestMxQueryCmd_SidecarUnavailable`) currently asserts this die is the intended, current behavior.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t263; t216 itself has a live worktree — `agent-a62468d0d1a7040cf`/`.claude/worktrees/t216` @ `8aa96bfb1` `[WT-hook-wiring-drift]`, per `00-worktree-list.txt` — but t216 is a separate card, not t263)

**Claim.** The card's premise assumes t216 M4's self-build fix already landed in `mx_query.go`; it has not. `mx_query.go`'s cold-start path still Stat()s for the sidecar file and immediately returns a `SidecarUnavailable` error with no self-build attempt, and `resolver_query.go` mirrors the same shape. There is therefore no landed self-build pattern yet to "reuse" in `file_changed.go`.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt216\b' --oneline
(no output)

$ sed -n '92,104p' internal/cli/mx_query.go
			// verify sidecar file exists (REQ-SPC-004-013)
			sidecarPath := filepath.Join(stateDir, mx.SidecarFileName)
			if _, err := os.Stat(sidecarPath); os.IsNotExist(err) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
					"SidecarUnavailable: sidecar index does not exist — run 'moai mx scan' to build the index\n")
				return fmt.Errorf("SidecarUnavailable: no sidecar index")
			}

$ grep -n "TestMxQueryCmd_SidecarUnavailable" internal/cli/mx_query_test.go
90:func TestMxQueryCmd_SidecarUnavailable(t *testing.T) {
(test asserts the "sidecar missing → SidecarUnavailable error" behavior IS correct, per AC-SPC-004-04)

$ sed -n '155,183p' internal/mx/sidecar.go
(Manager.UpdateFile → loadWithoutLock: os.IsNotExist(err) → returns an EMPTY Sidecar with no error — the incremental path file_changed.go actually calls degrades gracefully, unlike mx_query.go)
```

**Baseline-attribution.** `internal/cli/mx_query.go`, `internal/cli/mx_query_test.go`, `internal/mx/sidecar.go`, `internal/mx/resolver_query.go`, and `internal/hook/file_changed.go` all read at worktree HEAD (`6165f9f5e`).

**Gaps.** Did not open t216's own worktree (`.claude/worktrees/t216`, `WT-hook-wiring-drift`) to check whether M4's self-build fix exists there in-flight — that would confirm the fix is written but simply unmerged, versus not yet written at all. Did not fully trace whether `file_changed.go`'s actual call chain (via `mx.Manager.UpdateFile`) has any *other* die path besides the one traced (only `loadWithoutLock` and `writeWithoutLock` were read).

**Residual-risk.** If t216's worktree does contain a landed-but-unmerged self-build fix, the card's overall direction (reuse M4's pattern once it lands) is still sound — only the "already fixed, therefore reuse it" framing is premature. The card's own scope item (2) ("M4가 세운 자가빌드 경로를 재사용") depends on M4 actually merging first.

**Proposed disposition.** `needs-operator-decision` — the premise as stated is false today (t216 hasn't landed), but the underlying direction may still be correct once t216 does land; the operator should decide whether to defer this card behind t216 or drop/rewrite it now.

**Overlap candidates.** t216 (hard dependency — the card's own cited fix lives there and has not landed).

---

### t264

**Premise (one sentence).** A large and growing set of local `WT-*` branches from already-merged (squash-merged) cards has accumulated with no worktree occupying them, and cleanup is blocked in the primary checkout by BranchGuard, whose two exemptions are unreachable from a tool-spawned subagent, so an operator must run the cleanup by hand.

**Premise verdict.** `holds` — and the situation has grown since the card's own 2026-08-25 measurement (129 local `WT-*` branches / ~60 orphaned) to 196 local `WT-*` branches / 89 orphaned as of this sweep.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t264 itself)

**Claim.** The orphaned-branch accumulation the card describes is real and has worsened; of the card's own 5 named example branches, 2 (`WT-security-scan-surface`, `WT-web-live-todo`) are still present, while 3 (`WT-codex-session-msg`, `WT-constitution-retire`, `WT-audit-model-pin`) have already been cleaned up since the card was written — partial progress, not full resolution.

**Evidence.**
```
$ git branch --list "WT-*" | wc -l
     196

$ grep -o '\[WT-[^]]*\]' .moai/reports/t332/00-worktree-list.txt | wc -l
     107

# 196 total - 107 worktree-occupied = 89 orphaned local branches

$ git branch --list "WT-codex-session-msg" "WT-web-live-todo" "WT-constitution-retire" "WT-audit-model-pin" "WT-security-scan-surface"
  WT-security-scan-surface
  WT-web-live-todo
```

**Baseline-attribution.** `git branch --list` run from this worktree (`6165f9f5e`) against the shared repository's ref namespace (branches are shared across all worktrees of one `.git`). `00-worktree-list.txt` read as-is per instructions (not re-run).

**Gaps.** Did not verify landing-proof for any specific orphaned branch (per the card's own scope item (1), "저작 경로 좁힌 diff 공집합 확인") — that per-branch verification is explicitly the card's own future work, not something this sweep should pre-empt. Did not check remote branches.

**Residual-risk.** Some of the 89 "orphaned" branches counted here may not actually be landed/mergeable-safe to delete (e.g. an abandoned experiment) — the raw count is not itself a landing proof, only a scale indicator, exactly as the card's own "대표 mutant" warning anticipates.

**Proposed disposition.** `keep` — premise holds and has strengthened; rests on the branch-count and worktree-occupancy comparison above.

**Overlap candidates.** none observed among in-scope ids beyond general worktree-lifecycle housekeeping.

---

### t280

**Premise (one sentence).** The lessons-inbox *collection* mechanism ships to every distributed user via the compiled `moai` binary and the template-shipped `PostToolUseFailure` hook wiring, but the *drain* trigger that consumes it only exists in this repository's local, untracked `settings.local.json`, so a deployed user's `lessons-inbox.jsonl` grows without any drain path at all.

**Premise verdict.** `holds` — confirmed on both halves: collection ships (Go code + template hook wiring), drain does not (absent from both tracked `settings.json` and the distributed template).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t280)

**Claim.** `internal/hook/failure_observer.go` (compiled into the `moai` binary, therefore shipped to every install) writes to `.moai/lessons-inbox.jsonl`, and the distributed template's `settings.json.tmpl` wires `PostToolUseFailure` to a handler that reaches it; no drain trigger (`lsel`/`drain`/`session_drain`) appears anywhere in the tracked `.claude/settings.json` or the template's `settings.json.tmpl`.

**Evidence.**
```
$ grep -n "func \|lessons-inbox" internal/hook/failure_observer.go | head -6
46:func recordToolFailureEvent(input *HookInput, category ErrorCategory) {
129:func appendLessonsInboxStub(root, eventKey, summary, source string) {
139: ... marshal lessons-inbox stub ...

$ grep -n "PostToolUseFailure" internal/template/templates/.claude/settings.json.tmpl
209:    "PostToolUseFailure": [
214:            ... "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-post-tool-failure.sh"

$ grep -n "lsel\|drain" .claude/settings.json
(no output)

$ grep -n "lsel\|drain\|lessons-inbox" internal/template/templates/.claude/settings.json.tmpl
(no output)
```

**Baseline-attribution.** `internal/hook/failure_observer.go`, `.claude/settings.json` (tracked), and `internal/template/templates/.claude/settings.json.tmpl` all read at worktree HEAD (`6165f9f5e`).

**Gaps.** Did not check whether `handle-post-tool-failure.sh` actually invokes the `moai hook` subcommand that reaches `recordToolFailureEvent` (traced the Go side and the settings wiring, not the intermediate shell wrapper). Did not verify the t259 (SPEC-LSEL-DRAIN-STALL-001) fix's exact scope boundary beyond what CLAUDE.local.md §28 already documents (this card explicitly treats that as its starting fact, not something to re-derive).

**Residual-risk.** If some other, undiscovered drain path exists for distributed users (e.g. a periodic `moai` subcommand run manually), the "permanent absence" framing could be too strong — this sweep found no such path in the settings/template surfaces checked, but did not exhaustively search all CLI subcommands.

**Proposed disposition.** `keep` — premise holds; well-evidenced structural gap between what ships and what doesn't.

**Overlap candidates.** t260 (same underlying `lessons-inbox.jsonl` mechanism, different concern — collection *scope* vs drain *deployment*).

---

### t281

**Premise (one sentence).** The operator decided on 2026-08-26 (C안) to promote the local `develop` branch to a standing, local-only integration/RC-testbed branch while keeping per-card PRs against `main` and never pushing `develop` to origin.

**Premise verdict.** `falsified` — this decision was explicitly and completely reversed the very next day. `CLAUDE.local.md` §4.1, as it exists at this worktree's HEAD, documents a full switch to git-flow: `develop` is now pushed to origin (`origin/develop`), card PRs against `main` are abolished in favor of direct merges into a shared `develop` integration worktree, and `release/vX.Y.Z` branches from `develop` is the only path to `main`. §4.1.3 of the same file states this in its own words: "이 절은 2026-08-26 백로그 카드 t281이 정한 '로컬 전용·일회용 develop, 원격 push 금지'를 명시적으로 뒤집는다."

**Landing verdict.** `not-landed`
- commit: `11216d13f612f7e7161487a4e1369a47612f0b4c` — **not a delivery of t281; the opposite: an explicit reversal of the t281 decision**, per its own commit message ("Operator-directed reversal of the 2026-08-26 t281 decision")
- pinned ref: `ee50984abe4f11ac337382b48a26328f091e200a`
- `--is-ancestor` exit: `0`
- branch + tip (in-flight only): —

**Claim.** No commit implements t281's C안 as described (local-only, disposable develop with per-card PRs). The one commit whose message mentions `t281` (`11216d13f6`) is a reversal, dated one day after the card's own cited decision date, ancestor-confirmed in pinned develop. t281's action items (rc.N numbering convention, `make build VERSION=` procedure documentation, BranchGuard-routing decision, etc.) target a model that CLAUDE.local.md itself now documents as superseded.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt281\b' --oneline
11216d13f docs(workflow): switch development model from GitHub Flow to git-flow

$ git show --stat 11216d13f | head -14
commit 11216d13f612f7e7161487a4e1369a47612f0b4c
Author: t <t@t.t>
Date:   Thu Aug 27 12:55:23 2026 +0900

    docs(workflow): switch development model from GitHub Flow to git-flow

    Operator-directed reversal of the 2026-08-26 t281 decision (local-only,
    disposable develop). Card worktrees now branch from develop, lanes merge
    directly into a single develop integration worktree with no per-card PR,
    origin/develop is pushed and CI-verified, rc.N builds are cut from
    develop, and release/vX.Y.Z branched from develop is the only path to
    main.

$ git merge-base --is-ancestor 11216d13f612f7e7161487a4e1369a47612f0b4c ee50984abe4f11ac337382b48a26328f091e200a ; echo "exit:$?"
exit:0

$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt281\b' --oneline
(no output)
```

**Baseline-attribution.** `CLAUDE.local.md` §4.1/§4.1.3 as loaded into this session's system context at worktree HEAD (`6165f9f5e`); commit `11216d13f` checked against pinned develop via `--grep` + `--is-ancestor`.

**Gaps.** Did not diff `CLAUDE.local.md`'s exact wording against t281's five numbered scope items one-by-one (e.g. whether the rc.N numbering convention t281 asked for was separately addressed by the git-flow rewrite) — the supersession is total enough (opposite branch-push policy) that a line-by-line reconciliation did not seem load-bearing within this sweep's restraint budget.

**Residual-risk.** If any of t281's five scope sub-items (rc.N reset convention, `make build`/reinstall procedure documentation) were NOT actually covered by the git-flow rewrite, there could be a genuine residual gap hiding behind the "superseded" verdict — this sweep did not check for that.

**Proposed disposition.** `drop` — the decision the card documents and proposes to formalize has already been explicitly reversed by a later, more comprehensive operator decision recorded in the same file the card would have edited.

**Overlap candidates.** t310 (`WT-gitflow-doc-align`, live worktree per `00-worktree-list.txt` — very likely the card already carrying git-flow documentation-alignment work that supersedes t281's scope).

---

### t284

**Premise (one sentence).** `audit_multi`'s `disagreement_flag` derivation counts only the number of *distinct verdict values* among required backends, not the number of backends that actually participated on-target, so a convergence with only one genuine on-target verdict (others excluded as off-target or inconclusive) reports `disagreement_flag=false` — "agreement" — when there was nothing to compare.

**Premise verdict.** `holds` — confirmed directly in `converge()`'s Step 2.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t284)

**Claim.** `internal/cli/mcp_convergence.go`'s `disagreement := len(distinctRequired) > 1` derives the flag purely from the *count of distinct pass/fail values*, never from the count of participating/on-target backends; `ConvergenceResult` carries no participant-count or on-target-count field at all.

**Evidence.**
```
$ sed -n '166,171p' internal/cli/mcp_convergence.go
	// ── Step 2: disagreement_flag derivation ──
	distinctRequired := distinctVerdicts(required, "pass", "fail")
	disagreement := len(distinctRequired) > 1

$ sed -n '104,113p' internal/cli/mcp_convergence.go
type ConvergenceResult struct {
	PerBackendVerdicts []PerBackendVerdict `json:"per_backend_verdicts"`
	OverallVerdict   string   `json:"overall_verdict"`
	DisagreementFlag bool     `json:"disagreement_flag"`
	ResidualRiskNote string   `json:"residual_risk_note"`
	FailOpenBackends []string `json:"fail_open_backends"`
}
```
No field for participant/on-target count exists in the struct; with exactly one required backend returning `pass` (others excluded/inconclusive), `distinctRequired = ["pass"]`, `len(distinctRequired) == 1`, so `disagreement = false`.

**Baseline-attribution.** `internal/cli/mcp_convergence.go` read at worktree HEAD (`6165f9f5e`).

**Gaps.** Did not check `.moai/reports/t229/succession.md` directly (the card cites it as its observation source) — took the card's own quoted description of the codex-off-target/GLM-inconclusive scenario as given and verified only that the current code has no structural defense against it. Did not check whether `PerBackendVerdicts` (which IS exposed) lets a *consumer* reconstruct the on-target count themselves, which would partially mitigate the gap even without a dedicated field.

**Residual-risk.** If `PerBackendVerdicts` already lets a careful reader count on-target backends themselves, the practical severity may be lower than "disagreement_flag lies outright" — the flag itself is still misleading in isolation, which is the card's core point.

**Proposed disposition.** `keep` — premise holds on direct code read.

**Overlap candidates.** none observed among in-scope ids (t229, the succession-record source, is not in the in-scope list — already closed per repository memory).
