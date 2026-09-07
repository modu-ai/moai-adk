# Plan-Audit Verdict — SPEC-REMOVAL-GUARD-EXTRAS-001

- Card: t511 (GitHub #1658 + #1686 residual repair) · Tier M · class B · lens --security
- Auditor: plan-auditor (adversarial, plan-phase) · 2026-09-07
- Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t511` · branch `WT-danger-guard-regex` · HEAD `0b1e27877` (= `origin/develop`, exact)

## VERDICT: PASS — score 0.87 (Tier M threshold 0.80)

Finding count: **0 blocking (S1) · 0 major (S2) · 4 minor (S3) · 2 advisory observations.**
All four S3 findings are repairable by wording/one-AC additions; none blocks run-phase entry.
The SPEC's factual layer is exceptionally strong: every mechanical claim I independently
re-executed held exactly (20+ checks, §2 below).

---

## 1. Claim

The SPEC asserts: (a) the #1658/#1686 main body was already repaired on develop by `034dad48e`
(structural `dangerousRemovalTarget` check; its 4 tests pass today); (b) what remains is one
flawed extras regex `(?i)rm\s+-rf\s+/[^.]` shipped in THREE files, which fires on harmless data
mentions because heredoc bodies and unquoted spans are not quote-folded; (c) the repair is
deletion of that line from 3 copies plus a NEW deployed-policy (defaults + extras merge)
bidirectional regression surface, with two named mutants as discriminators; (d) 10 REQs measured
by 11 two-cell ACs. I audited each layer skeptically and re-executed the observable claims.

## 2. Evidence (all run by me, this session, this tree)

Baseline-attribution for every row: **tree `0b1e27877`, worktree `.claude/worktrees/t511`,
branch `WT-danger-guard-regex`, 2026-09-07.** The SPEC's own ledger pins the same SHA (document
level, acceptance.md:3, research.md:4) — pin matches measurement tree.

| # | Claim under test | Command | Observed output (verbatim) | Result |
|---|------------------|---------|---------------------------|--------|
| E1 | Tree/branch identity | `git rev-parse --show-toplevel` / `git branch --show-current` / `git rev-parse HEAD` | `.../.claude/worktrees/t511` / `WT-danger-guard-regex` / `0b1e27877259fef70079188bf7714029cbaa7ded` | MATCH |
| E2 | 3 copies at cited lines | `sed -n` on the three files | C1 `internal/template/templates/.moai/config/sections/security.yaml:12` = ``- 'rm\s+-rf\s+/[^.]'  # rm -rf targeting root paths``; C2 `.moai/config/sections/security.yaml:7`; C3 `internal/settings/testdata/sections/security.yaml:7` | MATCH |
| E3 | AC-RGE-008 RED | `grep -rn 'rm\s+-rf' internal/ .moai/config/` | exactly 3 matches: testdata:7, templates:12, dogfood:7 | MATCH |
| E4 | Baseline test (LEDGER-BASE) | `unset … && go test ./internal/hook/ -run 'TestDangerousRemoval' -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/hook	0.693s`, exit 0 (SPEC recorded 0.669s from its own run; same command/tree, timing varies) | MATCH |
| E5 | 034dad48e in this tree | `git merge-base --is-ancestor 034dad48e HEAD` + `git show --stat` | ancestor=yes; touched `dangerous_removal.go` (+252), `dangerous_removal_test.go` (+105), `pre_tool.go` (28), t286 verdict | MATCH |
| E6 | **Live RED re-execution (LEDGER-R1)** | my own Bash heredoc repro | **REFUSED: `Dangerous command blocked: (?i)rm\s+-rf\s+/[^.]` — command not executed.** Byte-identical to the SPEC's recorded refusal; message format identifies the regex scan path (not the structural check's `removal of protected path %q`) | CONFIRMED |
| E7 | §R.3 pattern table (6 rows) | `python3 -c` with subjects built via `chr(32)` (no firing form in command text) | deep path=True / bare root=False / /usr=True / /.=False / unquoted echo data=True / `$HOME`=False | 6/6 MATCH — `[^.]` coverage inversion confirmed |
| E8 | Source line claims | targeted `sed -n` / `grep -n` | `pre_tool.go`: func 934, structural check 947-949 (message `%q` at 948), quote-fold 956, regex scan 959-963 (message at 961); `MergeExtraPatterns` at 298 with extras append 302-304; `deps.go` deployed-policy combo 240-243; `security/config.go` 24-39 graceful-nil loader; `branch_guard.go` 197-199; `dangerous_removal.go:27` `shellSeparators` includes `"\n"`; `dangerous_removal_test.go:12` builtin-only `decideBash`; exactly 4 `TestDangerousRemoval_*` funcs (lines 25/57/74/93) | ALL MATCH |
| E9 | Import-cycle absence + M2 path-A feasibility | `grep -rln 'moai-adk/internal/hook' internal/template/` → 0 files; `internal/template/embed.go:42` | `EmbeddedTemplates() (fs.FS, error)` exists — plan M2 principle path A is implementable | MATCH |
| E10 | C3 consumers | grep sectionwrite_test/schema_sections_test | `seedSectionFixture` (sectionwrite_test.go:15) + schema_sections_test.go consume the fixture; not a policy consumer | MATCH |
| E11 | Build chain + neutrality guard | Makefile:34-36; `ls .github/workflows/template-neutrality-check.yaml` | build = agents-emit-check, commands-emit-check, templ-generate, gen-catalog-hashes --all, go build; CI guard exists, binary-fail classes C1/C5/C6 over `templates/**` | MATCH |
| E12 | Mutant B discriminating power | read builtin list `pre_tool.go:177+` | no `rm -rf` builtin regexes remain (supabase/neon/vercel/rmdir-Windows/format/terraform only) — with the structural call removed, deny tests MUST go red | CONFIRMED |
| E13 | Repo-wide copy sweep (beyond SPEC scope) | `grep -rlF 'rm\s+-rf' .` | 3 config copies + `CHANGELOG.md:426` (historical t286 entry — immutable, correct to keep) + **docs-site 4 locales** (see F4) | NEW FACT |

## 3. Findings

### F1 — S3 (minor) · REQ-RGE-010 has no acceptance criterion
- Where: `spec.md:133` (REQ-RGE-010) vs `acceptance.md:26-38` (AC matrix).
- REQ-RGE-010 (extend-never-replace contract + other 5 extras lines untouched) is the only REQ
  no AC cites (AC-011 is a discipline meta-criterion citing none). The requirement is carried
  only procedurally (plan §A.5 PRESERVE, §D constraints) — nothing measures it.
- Repair: add one AC (e.g., post-M1 `git diff --numstat` shows exactly 1 deleted line per file
  AND the other 5 pattern lines still present; or an M2 constant asserting the extras list the
  deployed-policy helper loads has exactly 5 entries) or record REQ-RGE-010 explicitly as
  verified-by-PRESERVE-procedure in acceptance.md.

### F2 — S3 (minor) · M2 test-content spec contradicts AC-RGE-008's green path
- Where: `plan.md:112` ("R1 형태 **전체** 명령 문자열 1건 이상") vs `acceptance.md:35` (AC-008
  green: same grep → 0 matches) + `acceptance.md:44` (채택 절차 #3 re-runs the grep at
  M2-completion).
- R1's full form includes the line `pattern literal: (?i)rm\s+-rf\s+/[^.]`. If the M2 heredoc
  test embeds that raw literal in `internal/hook/dangerous_removal_test.go`, the M2-completion
  grep finds ≥1 match inside `internal/` and AC-008 can never go green. The line that makes R1
  fire is the real-space example form, not the literal.
- Repair: one sentence in plan M2.2 — the heredoc reproduces the example-form line only; the
  pattern-literal line is deliberately omitted (AC-008 scope).

### F3 — S3 (minor) · AC-RGE-009 (release-blocking) carries no four-element RED cell
- Where: `acceptance.md:36` — RED cell is prose "없음 — 현재 임베디드 FS에 결함 라인 존재(문제
  상태)".
- Its red is materially the same underlying state AC-008's grep RED already observes, and its
  green is enforced transitively (a stale embedded FS keeps the extras loaded → the AC-001~003
  allow tests go red). Enforcement is real; the cell form deviates from the two-cell four-element
  rule for a release-blocking row.
- Repair: cite `[LEDGER-* AC-008 grep]` as AC-009's RED cell (same state, same tree pin), or
  reclassify AC-009 as regression-guard with the transitive-enforcement note.

### F4 — S3 (minor) · docs-site 4 locales document the removed pattern; no sync obligation recorded
- Where: `docs-site/content/{en,ko,ja,zh}/advanced/config-sections.md` (en line 131) show
  `'rm\s+-rf\s+/[^.]'` as an example entry of `extra_dangerous_bash_patterns` in the shipped
  config. Neither plan.md §F milestones nor acceptance.md nor the PRESERVE list mentions them.
- After M1 lands, 4 locale pages document an extras pattern the shipped template no longer
  carries — and sync-phase (manager-docs) works from plan.md's file list, so it will not be
  discovered there. (CHANGELOG.md:426 is the t286 historical entry — immutable history, correct
  to keep; not a finding.)
- Repair: add one sync-phase item to plan M3 ("remove the stale example line from the 4
  docs-site config-sections locales, 4-locale same-PR per docs i18n rules") or record the
  staleness as a scope-out with rationale.

### Advisory A1 — AC-005/006/007 RED cells are run-phase mutant placeholders
`acceptance.md:21-22` ([LEDGER-MUT-A/B] "런 페이즈 채움"). Disclosed and procedurally coherent —
mutant red is inherently post-test-authoring, and AC-007 makes failure-to-induce-red a blocking
rework trigger, so the discrimination is self-policing. Not a defect at plan phase. Sync-audit
must verify LEDGER-MUT-A/B are populated with verbatim RED output before PASS.

### Advisory A2 — AC-007's RED reading is conditional on M2 principle paths A/B
`plan.md:110`. Under fallback C (inline extras slice), re-adding the template line (mutant A)
flips nothing even with a fully correct, discriminative wiring — the plan requires E-verification
disclosure for C but does not state that a non-red mutant A then means "wrong path chosen", not
"test not discriminative". One clause in M2.4 closes it.

## 4. Per-axis assessment (audit axes from the dispatch)

1. **Two-cell discipline** — AC-001/002/003 exemplary (observed, tree-pinned, re-executable REDs
   — E6 re-executed one myself); AC-004 correctly regression-guard (green-now); F3 (AC-009 form);
   A1 (deferred mutant reds, disclosed). No single-cell release-blocking AC beyond F3.
2. **Bidirectional pairing** — relaxation ACs 001-004 ↔ tightening 005-006, discriminator 007,
   sweep discipline 011. I probed for a mutant satisfying all 11 ACs while restoring either
   direction: the builtin-only-policy wiring mutant is caught (mutant A stays green → AC-007
   fails → M2 rework); the structural-check-removal mutant is caught (E12: no builtin rm regexes
   remain, so deny tests must red). No hole found.
3. **Mutant executability** — both mutants are concrete, executable, revert-disciplined
   (re-add line + `make build`; comment out `pre_tool.go:947-949`), each mapping to named REQs.
   HARD #2 satisfied.
4. **Scope boundaries** — t512/#1659 excluded twice (spec §D hard constraint + §F); other extras
   lines protected (§D + PRESERVE); #1686 override-key is a recorded scope-out WITH rationale
   (spec §A.6 + §F, including a re-review trigger). Compliant.
5. **Template neutrality** — M1 is delete-only; AC-010 measures `deleted=1 added=0` + CI guard;
   plan §D/§G forbid repair comments. The SPEC artifacts quoting the literal live under
   `.moai/specs/` (outside the guard surface). Compliant.
6. **Traceability** — 9/10 REQs map to ACs (F1); every AC cites its REQs; plan milestones cover
   M1/M2/M3 with PRESERVE/EXTEND lists. F2 is a revised-scope wording mismatch between plan and
   acceptance layers.
7. **Evidence-ledger fact-check** — every ledger claim I could re-execute held exactly (E1-E13).
   The three-copy discovery (dispatch said two; SPEC found the testdata third) is itself
   verified (E3).
8. **Live RED** — re-executed; refusal recorded verbatim (E6); not retried, not routed around.

## 5. Gaps (explicitly NOT observed)

- Mutants A/B were NOT executed by me — their predicted REDs are run-phase obligations, unobserved
  at plan time (A1).
- `make build`, the M1 settings-suite run, and all M2 tests were NOT run (M1/M2 unexecuted; the
  tests do not exist yet).
- The deployed policy was verified via the live session guard + source read, NOT by driving a
  separately compiled binary through the hook — no second-process reproduction was performed.
- Cross-model audit skipped: `llm.yaml` is absent in this worktree and the primary copy carries no
  `audit_model` key → not `multi`.
- Two of my own probe attempts were guard-blocked (a `/tmp` Write refused as path traversal; a
  compound python heredoc refused by the worktree guard) — neither bears on the SPEC; E7 was
  obtained with a compliant single command instead.

## 6. Residual-risk

- If run-phase lands on M2 fallback C (inline slice), AC-007's mutant-A reading becomes
  conditionally invalid (A2) — a green-everything + non-red-mutant outcome would then be
  misdiagnosable. The E-verification disclosure the plan already requires must explicitly state
  the path chosen.
- AC-008's green is scope-bounded (`internal/` + `.moai/config/`); F4's docs-site copies sit
  outside it, so a green AC-008 does not mean the literal is gone from the repository.
- The refusal message as discriminator ("message is the pattern literal → regex path fired")
  rests on the current message formats at `pre_tool.go:948`/`961`; a future message rewording
  would weaken R1-R3's subject-attribution reasoning (the ACs themselves measure behavior, not
  the message, so exposure is limited to the ledger narrative).

---

Score composition: factual accuracy/evidence ledger 1.00 · mutant discipline 1.00 · scope &
neutrality 1.00 · plan executability 0.95 · traceability 0.85 · two-cell discipline 0.80 →
weighted verdict **0.87** after S3 deductions. PASS ≥ 0.80.

Audit surface: documents only — no code, template, or SPEC artifact was modified by this audit.
Evidence written to `.moai/reports/t511/plan-audit-verdict.md` (this file).

---
---

# ITERATION 2 — Resolution Verification (2026-09-07, same session)

Scope: verification-only re-audit of the four S3 resolutions. The full audit was NOT re-derived.
Tree re-read before verifying: `git rev-parse --short HEAD` → `0b1e27877`, branch
`WT-danger-guard-regex` unchanged, `git status --porcelain` → only `.moai/reports/t511/` +
`.moai/specs/SPEC-REMOVAL-GUARD-EXTRAS-001/` (artifact edits in place; no code/template touched).

## VERDICT (iter2): PASS — score 0.93 (Tier M threshold 0.80)

Finding count: **0 S1 · 0 S2 · 0 S3 · 2 advisory carried (A1, A2 — neither blocking).**
All four iter-1 S3 findings are verified RESOLVED against the current artifact set.

## Resolution verification (per coordinator check)

**Check 1 — each fix landed where claimed.** Verified by reading the cited lines:
- F2 → `plan.md:111` (M2.2 fixture discipline: real-space example forms only, pattern-literal
  line deliberately absent from ALL fixtures, programmatic Go construction when the pattern text
  is needed — "리터럴이 소스에 착지하지 않는다"); `plan.md:112` first allow-test now reads
  "R1의 예시 형태 행만 담은 heredoc … 패턴 리터럴 행 없음" (the iter-1 "전체" wording is gone);
  `acceptance.md:40` (AC-001 green cites the discipline); `acceptance.md:41` (AC-002);
  `acceptance.md:47` (AC-008 green states the reachability condition). LANDED.
- F2c → `acceptance.md:28-34` [LEDGER-GREP3]: scoped grep, fenced verbatim 3-match output,
  exit 0, tree `0b1e27877`, scope note (`.moai/specs/`/`.moai/reports/`/docs-site outside
  scope). Its verbatim output matches MY OWN iter-1 E3 measurement byte-for-byte (same 3
  files, same lines, same content) — independently corroborated. LANDED.
- F1 → `acceptance.md:49`: AC-RGE-010 now cites REQ-RGE-009 + **010**; green = per-copy
  `git diff --numstat` deleted=1 added=0 (one measurement judging both requirements);
  named discriminating mutant (change one other extras line → numstat deviates → RED);
  regression-guard classification WITH stated rationale (invariant-preserving direction has
  no RED-now). REQ coverage is now 10/10. LANDED.
- F3 → `acceptance.md:26` [LEDGER-R4] + `acceptance.md:48` (AC-009 row cites it). LANDED.
- F4 → `plan.md:128` (M3 §2): all 4 locale paths, line 131, 4-locale same-PR obligation per
  `docs-site-i18n-rules.md`, and a concrete sync gate. LANDED.
- `progress.md:14-20` §E.1 resolution map present, mapping all five fixes to file:line.
  `research.md:102` §R.7 (R.7.1/R.7.2/R.7.3) present, consistent with all claims. LANDED.

**Check 2 — F2 contradiction genuinely gone.** AC-008's green path walked as written: scope is
`internal/` + `.moai/config/`; M1 removes the 3 yaml copies (2 under `internal/`, 1 under
`.moai/config/`); M2.2 fixture discipline forbids the pattern-literal text landing anywhere in
the tree; the grep form empirically does NOT match real-space forms (iter-1 measurement: the
pre-existing `dangerous_removal_test.go` real-space `"rm -rf /"` lines did not match while the
yaml `\s` lines did), so real-space fixtures and Go-concatenated subjects add no matches.
**0-match is reachable as written**, and the green cell honestly conditions on the fixture
discipline. The docs-site copies (F4) sit outside the scope — no interaction.

**Check 3 — AC-010's mutant executable as described.** Yes: mutate one non-target extras line →
that file's numstat deviates from `0 1` → measurement RED. Concrete, run-phase-executable
(marked 선택), and genuinely discriminating — it proves the numstat measurement can distinguish
a REQ-RGE-010 violation.

**Check 4 — AC-009's RED cell carries all four elements.** [LEDGER-R4]: command
(`echo example: rm -rf /tmp/moai-spec-repro-456`) · verbatim output
(`Dangerous command blocked: (?i)rm\s+-rf\s+/[^.]`) · exit path (PreToolUse deny, not
executed) · tree SHA (`0b1e27877`, stated in the entry itself). I **independently re-executed
the echo form in this iteration** and received the byte-identical refusal, command not executed —
the cell's content and its stated meaning (deployed path still loads the defective extras) are
both confirmed live.

**Check 5 — F4 sync gate concrete.** `plan.md:128` closes with a checkable instruction:
sync 완료 판정 전 `grep -rn "rm\\s+-rf" docs-site/content/` → 0매치, on the 4 named locale
files under the 4-locale same-PR rule. An instruction manager-docs can execute and sync-audit
can read a verdict from.

## Evidence (iter2 measurements, this run, tree `0b1e27877`)

| # | Command | Observed output |
|---|---------|-----------------|
| I2-E1 | `git rev-parse --short HEAD` / `git branch --show-current` / `git status --porcelain` | `0b1e27877` / `WT-danger-guard-regex` / SPEC dir + reports dir only |
| I2-E2 | `echo example: rm -rf /tmp/moai-spec-repro-456` (LEDGER-R4 form) | **REFUSED: `Dangerous command blocked: (?i)rm\s+-rf\s+/[^.]`** — not executed |
| I2-E3 | Read of the five repaired surfaces at cited lines | all claims present as quoted above |
| I2-E4 | `sed -n '102,140p' research.md` | §R.7 with R.7.1/7.2/7.3 matching the ledger entries |

## Gaps (iter2)

- LEDGER-MUT-A/B remain unfilled placeholders — by design; they are run-phase obligations, not
  plan-phase defects. Sync-audit must verify they are populated with verbatim RED output before
  any run-phase PASS is recorded against AC-005/006/007.
- LEDGER-GREP3 and LEDGER-R4 were measured by manager-spec, not by me — but both are
  independently corroborated (I2-E2 re-executes R4's exact command; GREP3's output matches my
  iter-1 E3 measurement byte-for-byte on the same tree).
- Advisory A2 (iter-1) was not addressed: AC-007's mutant-A interpretation remains conditional
  on M2 principle paths A/B — under fallback C (inline slice) a non-red mutant A would mean
  "wrong path chosen", not "test not discriminative". plan M2.1 still requires the E-verification
  disclosure for C, which contains the exposure; advisory only, non-blocking.

## Residual-risk (iter2)

- The AC-008 green is conditioned on fixture discipline being OBSERVED at M2, not mechanically
  enforced; a run-phase executor pasting the R1 literal into a test string would reintroduce the
  contradiction. The plan's §G anti-pattern list and M2.2 wording make this visible to review.
- Unchanged from iter-1: refusal-message-as-discriminator rests on current message formats at
  `pre_tool.go:948`/`961`.

## Iter2 score composition

factual accuracy/evidence ledger 1.00 · two-cell discipline 0.95 (AC-009 now four-element;
AC-008 RED measured + conditioned green; mutant placeholders structurally deferred) ·
traceability 1.00 (10/10 REQs → ACs) · scope & neutrality 1.00 · mutant discipline 1.00 ·
plan executability 1.00 (F2 eliminated) → weighted **0.93** after the open A2 advisory and
placeholder deferral deductions. PASS ≥ 0.80.

**FINAL (current artifact set): PASS 0.93** — run-phase entry is unblocked. Skip-eligibility
note per spec-workflow: this iter2 verdict supersedes the iter-1 verdict for the CURRENT
artifact hash; any further plan-artifact edit invalidates it.
