# Sync-Audit Verdict — SPEC-WORKTREE-GUARD-HEREDOC-DOC-001 (card t512, GH #1659)

> Auditor: sync-auditor (independent spawn, per progress.md §E.4 note). Audit-only session:
> no code changes, no commits. Measured in worktree `.claude/worktrees/t512`, branch
> `WT-guard-heredoc`, HEAD `01ee9191a`. All commands below were run by this audit in this
> session against this tree. Multi-model fan-out skipped: `.moai/config/sections/llm.yaml`
> carries no `audit_model` key (checked; single-verdict path).

**최종 판정: PASS — 0.96 / 1.00** (Functionality 0.95 · Security 1.00 · Craft 0.92 · Consistency 0.98, harmonic mean 0.96)

## Claim

1. The SPEC is closed with an intact 3-phase chain: plan `d71fd5f33` → M1 `504d6457c` → M2 `b482942fc` → M3 `76c296623` → evidence `06900b1fc` → close `86c9e0eb2` → backfill `01ee9191a`, all above base `6a46c0edb`.
2. All 7 acceptance criteria are mechanically verified PASS: anchors GREEN at HEAD (3/2/1/0), RED-now cells re-verified red at the pinned base tree, mirror parity holds (cmp exit 0 + drift test ok), the mutant observation is recorded with the failure observed first, and both drafts carry their required elements with the tone/scope constraints respected.
3. Scope is documentation-only: zero `.go` files in the entire card diff, `branch_guard.go` untouched, the repro ledger lands exactly once (plan commit) and was never modified after.
4. Close integrity holds: after the sync commit only progress.md (backfill) changed; spec.md transitions are frontmatter-only (draft→in-progress in M1, in-progress→completed in sync); §E.4 carries the real sync SHA; the CHANGELOG entry exists exactly once and its numeric claims match independent measurement.
5. The orchestrator-direct execution of run+sync (manager-develop terminated 4x by GLM 429s; manager-docs skipped for the same reason, under the operator's explicit continue directive) is disclosed in §E.2/§E.4 attribution notes, was preceded by a real delegation attempt, and crossed executor identity only — never artifact scope. The phase contracts survived the substitution.
6. The doctrine section does not hand anyone a guard-weakening recipe.

## Evidence

**STEP 0**: `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t512`; `git branch --show-current` → `WT-guard-heredoc`; `git rev-parse --short HEAD` → `01ee9191a`. Match.

**Axis 1 — anchors, re-run pipe-free (HEAD `01ee9191a`):**

| Grep | Measured | rc | Threshold |
|---|---|---|---|
| `grep -c "brace expansion"` (local doctrine) | 3 | 0 | ≥2 (expect 3) — PASS |
| `grep -c "unquoted"` | 2 | 0 | ≥1 (expect 2) — PASS |
| `grep -c "re-measured"` | 1 | 0 | ≥1 — PASS |
| `grep -cE "1659\|SPEC-WORKTREE-GUARD-HEREDOC"` (TEMPLATE copy) | 0 | 1 | 0 — PASS |

**RED-now cells re-verified at the pinned base** (`git grep -c <anchor> 6a46c0edb -- <doctrine path>`): all three anchors → 0 matches, rc=1, matching the acceptance cells verbatim. The cells carry the four elements (§2.1 verification-completeness): command, verbatim output, exit code as its own field, tree SHA (`6a46c0edb`, document-level pin + per-cell restatement). AC-WGHD-006's ls cell additionally records the stream-attribution control (`2>/dev/null` → empty stdout + rc=1) — the empty-stdout-with-nonzero-exit form doctrine §2.1 admits. AC-WGHD-001 (invariant type — pre-RED impossible in principle) and AC-WGHD-004 (absence type) are honestly classified instead of force-fitted: the former adopted via a recorded mutant observation, the latter demoted to regression-guard with the vacuous-green hazard named. Both dispositions are the correct reading of the two-cell discipline.

**Axis 2 — mirror parity:**

- `cmp .claude/.../worktree-integration.md internal/template/templates/...` → exit 0.
- `go test ./internal/template/ -run TestRuleTemplateMirrorDrift -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 0.411s`.
- Test mechanism read (`internal/template/rule_template_mirror_test.go:137-177`): `bytes.Equal(srcContent, mirrorContent)` — byte-level, and the file IS listed in `workflowOptMirroredPaths` (line 57), so a 1-byte drift flips it red. The §E.2 mutant record (1-line drift → `cmp` differ char 44091 line 505 rc=1 → drift test FAIL 0.505s → restored by re-edit → exit 0 + ok 0.258s; also recorded in the M1 commit message) is mechanically sound: the failure was observed before the green, on a known input.
- The delegate-found drift (behaviour→behavior, line 505) and the mutant reuse the same spot — internally coherent.

**Axis 3 — scope:**

- `git diff --stat 6a46c0edb..HEAD` → 12 files: doctrine ×2 (each +29), SPEC ×5, reports ×4, CHANGELOG ×1. Note: the dispatch said "exactly 11 files (doctrine ×2, SPEC ×5, reports ×4)" — that arithmetic (2+5+4) omits the CHANGELOG.md its own axis 5 requires. Measured 12 is the correct population; dispatch-side slip, not a work defect.
- `.go` files in diff: 0. `branch_guard` mentions in diff: 0.
- `git log --oneline -- .moai/reports/t512/repro-heredoc-brace.md` (and `--all`) → exactly one commit, `d71fd5f33` (plan phase). Ledger unmodified since.

**Axis 4 — doctrine content (inserted "### Why the two heredoc delimiter forms differ", both copies byte-identical):**

- Both delimiter semantics present: quoted (`<<'EOF'`) → inert body, no expansion of any kind, brace is a literal; unquoted (`<<EOF`) → live body, all expansions apply.
- The unquoted bullet's guard sentence is intact: "a refusal that errs on the side of caution there is **correct behavior, not a defect**" — plus the strengthening "No guard — and no reader of this file — may treat an unquoted-delimiter body as inert."
- Observation register framed as observation, not specification, with the boundary explicitly left unmeasured twice ("Gap: the boundary … is unmeasured in both rows" / "the boundary of the guard's analyzer is still unmeasured").
- Neutrality of the inserted section: no `2026-` dates, no `1659`, no SPEC ID (grep → 0 hits); no `6a46c0edb` SHA and no repro-ledger path in either copy (REQ-WGHD-005 honored — the ledger path is not neutral text).
- Version-scoped: "Version measured: Claude Code 2.1.251" present.
- Insertion placement honored: the pre-existing observation table, Gap paragraph, and Workarounds section are byte-identical to base (base vs HEAD Workarounds compared — unchanged); the section was inserted between them as planned. The plan T1 draft matches the inserted text verbatim.
- Auxiliary loosening scan (AC-WGHD-004) `disable|weaken|suppress|reconfigure` over the inserted section → 0 (and the AC itself correctly records that this 0 proves nothing alone — the review cell carries the weight; this audit is that sync-phase review instance and the read concurs).

**Axis 5 — close integrity:**

- `git diff 86c9e0eb2..HEAD` → progress.md only (1 line). No `.go` files after close.
- `git show 86c9e0eb2 -- <spec.md>` → single hunk: `status: in-progress` → `completed`. Frontmatter only. M1's spec.md hunk: `status: draft` → `in-progress`, frontmatter only.
- Backfill `01ee9191a`: `sync_commit_sha: pending-backfill-sync` → `86c9e0eb2` — the real, independently measured sync SHA. D3 backfill pattern, not an ownership crossing.
- CHANGELOG: `grep -c SPEC-WORKTREE-GUARD-HEREDOC-DOC-001` → 1. Spot-checked claims, all matching independent measurement: anchors 0→3 / 0→2 (threshold ≥2 with the mutant rationale), cmp exit 0, drift-failure-observed-first sequence, plan-audit PASS 0.98 iter2 (verdict file confirms: iter1 0.91 → iter2 0.985→0.98), 7 REQ / 7 AC with "6 release-blocking + 1 honestly classified regression-guard" (matches the acceptance matrix), zero `.go` files, tree pin `6a46c0edb`, ledger path, draft-only posting boundary, executor note preserved rather than smoothed away.

**Axis 6 — ownership judgment:**

- Delegation happened first and its traces are in the artifacts: §E.2's M1 table attributes the mirror-drift correction to the delegated agent ("위임 에이전트의 드리프트 정정 후"), and the note states the delegate completed pre-flight, the M1 insertion, and first anchor measurements before terminating. The commit chain shows single-session authorship (`t <t@t.t>` — indistinguishable in this single-committer repo); nothing in the repo contradicts the note.
- The crossings were of executor identity, not artifact scope: M1 (manager-develop's transition slot) touched only doctrine ×2 + spec.md frontmatter; evidence commit touched only progress.md; sync commit (manager-docs's slot) touched only status line + CHANGELOG + §E.4 — all within each phase owner's allowed write surface. Both crossings carry truthful attribution notes (§E.2, §E.4), including the honest statement that E2/E3 (build/coverage) have no target in a docs-only SPEC.
- §E.4 names the residual: "독립 감사(sync-auditor)는 에이전트 스폰으로 수행한다" — this audit is that spawn, so sync-phase judgment is not self-judged.

**Axis 7 — drafts:**

- M2 reply (Korean, `reporter-reply-1659.md`): four elements all present — (a) reporter's diagnosis verified first with the paired-probe matrix and tree pin `6a46c0edb` (+ non-execution proof cited), (b) ownership = Claude Code binary with the repo doctrine quoted and `branch_guard.go` explicitly excluded, (c) two workarounds (Write tool / plain separate commands) presented as help, (d) upstream handoff + draft-only/post-after-landing explicit. Tone: passes the blame-free bar affirmatively — it opens by confirming the reporter was right, and it explicitly disclaims the blame-shift reading ("소속이 바이너리라는 것은 「저희 탓이 아니다」가 아니라…"). Contrast cell: `grep -c "1659"` → 3.
- M3 upstream (English, `upstream-draft-claude-code.md`): five elements all present — probe matrix, provable-inertness argument, fix direction 1 (fold braces like command substitutions), fix direction 2 (name the offending construct in the refusal), reporter credit (jjjh7401, #1659, credited twice including as the source of the argument). Both scope limits hold verbatim: no script-file-bypass ask, and an explicit "does NOT ask for any change to unquoted-delimiter bodies … erring toward refusal there is correct behavior."

**Axis 8 — security read (as adversary):** The section describes the analyzer's measured blind spot (braces treated as live in a provably inert position) without providing a smuggling path: the "passed" probe is the guard *working* (the substitution was folded/neutralized, not executed past isolation); the workarounds steer to the sanctioned Write-tool/plain-command conventions and explicitly de-legitimate ad-hoc detours; the boundary is declared unmeasured rather than mapped. Nothing in the section tells a reader how to get un-refused complex behavior past the guard. No weakening recipe.

## Baseline-attribution

Every figure above was measured in this run, in this session, against this worktree at HEAD `01ee9191a` (RED cells re-measured against pinned base `6a46c0edb` via `git grep <tree>`), except where marked a Gap. Anchors: pipe-free `grep -c` with rc captured separately. Mirror: `cmp` + `go test -run TestRuleTemplateMirrorDrift -count=1`. History: `git diff --stat`, `git log --oneline -- <path>`, `git show <sha> --stat|-- <path>`.

## Gaps

1. **The 4 GLM-429 terminations and the operator's continue directive are transcript-only.** They are claimed in §E.2/§E.4 and are consistent with every repo-observable trace, but no repo artifact can corroborate or refute them. Attributed claims, not audit-verified facts.
2. **The mutant observation was not re-executed by this audit** (audit-only, no tree mutation). Verified instead by mechanism: byte-level `bytes.Equal` + the file's presence in `workflowOptMirroredPaths`. The recorded rc=1/FAIL/restore sequence stands as the run-phase's observation, not this audit's.
3. **Probe 2's scratch file (55 bytes) and its cleanup are ledger claims** — the file is gone by design; not re-observable.
4. **The B12 pre-emission duplicate count (0) ran in the sync session** and is not re-observable; the post-state (exactly 1 entry) is consistent with a successful pre-check.
5. **plan.md, research.md, acceptance.md were read selectively** (T1 draft, frontmatter, acceptance matrix, §F) rather than line-by-line in full; the plan-audit (iter2, 0.98) is the authority for their internal quality.

## Residual-risk

1. **Most load-bearing: the doctrine documents an external binary's behavior with a version pin but no staleness signal.** The entire pivot rests on the refusal belonging to the Claude Code binary (`2.1.251` measured); when upstream changes the analyzer — fix, rework, or message rewording — the observation table and the why-section cannot announce their own staleness (verification-completeness §1.3's continued-firing concern, in prose form). The "Version measured" line is the mitigation; it works only if a future reader checks it. The natural repair is upstream-fix-driven (the M3 draft, once submitted, is the mechanism that gets the behavior corrected), which is why the unposted-draft state below matters.
2. **The card's user-facing promise is still un-discharged: both drafts are inert files.** #1659's reporter has received nothing; posting is the lead's post-landing act and the only carriers of that obligation are the draft headers and the CHANGELOG sentence. If the lead never posts, nothing in the repo goes red.
3. The M3 draft's "reproduces the refusal deterministically" rests on n=1 maintainer reproduction (plus the reporter's original report) — an overclaim relative to the SPEC's own observation discipline. Soften before submission (e.g. "reproduced deterministically in our session" or drop the adverb).
4. If the worktree-guard message ever changes upstream, the M2/M3 quotes become stale the same way the doctrine would; the drafts cite the message as measured at `2.1.251`.

## Scores

| Dimension | Score | Basis |
|---|---|---|
| Functionality | 0.95 | All 7 AC mechanically PASS with re-measured evidence; honesty classifications correct; −0.05 for the M3 "deterministically" overclaim inside a deliverable |
| Security | 1.00 | No weakening recipe; adversarial read clean; workarounds steer to sanctioned paths; template copy neutral (0 IDs / 0 dates / 0 SHAs in section) |
| Craft | 0.92 | Four-element RED cells, stream-attribution control, failure-first mutant, mutant-justified threshold (≥2), honest regression-guard demotion; −0.08 for the CHANGELOG "zero occurrences in this repository's source" wording (true of Go source — verified 0 in `.go` — but the doctrine's own markdown now quotes the message 4×, so a whole-repo grep contradicts the loose reading) |
| Consistency | 0.98 | plan↔acceptance↔doctrine↔§E aligned (F2's ≥2 promotion landed in both documents); commit messages match content; CHANGELOG claims match measurement; §F present; −0.02 for the dispatch's 11-vs-12 file arithmetic being left uncorrected anywhere in the card record |
| **Harmonic mean** | **0.96** | |

**Verdict: PASS** (Tier M threshold cleared; no must-fail finding. The three Minor findings — CHANGELOG source-scope wording, M3 "deterministically", draft-posting obligation — are recorded for the lead; none blocks the close. The M3 softening is worth one line before the draft is ever submitted.)

**Post-verdict tool re-verification (lead advisory, 2026-09-07)**: this shell's `grep` is a ugrep wrapper that silently skips binary-suspected and gitignored files, so every plain-`grep` 0 in this card was re-measured with `/usr/bin/grep` before window entry. Results identical: the refusal sentence hits exactly one file under `internal/` + `.claude/hooks/` — the doctrine markdown itself (`internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`), zero `.go`/`.sh` source hits; template neutrality 0/rc=1; anchors 3 and 2 unchanged. The ownership finding ("no repository source implements or configures this check") survives the tooling correction.
