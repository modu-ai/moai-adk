# SPEC Review Report: SPEC-GIT-DELIVERY-PROCEDURE-001 (reduced, revision 0.2.2)
Iteration: 2/2 (Tier M ceiling — `.moai/config/sections/harness.yaml:77` `M: 2`)
Verdict: **FAIL**
Overall Score: **0.75** (harmonic mean; Tier M PASS threshold 0.80). Iteration 1 was 0.71, so the score went up and no STOP-on-regression applies.

Reasoning context ignored per M1 Context Isolation. The binding decisions in the dispatch (scope, flag semantics, approval rule, X1–X4 ruling, Tier M) were treated as fixed inputs and not re-argued.

## Tree attribution

```
$ git rev-parse --show-toplevel      → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t622
$ git branch --show-current          → WT-git-procedure-fixes
$ git log --format='%h %s' -3        → 3512b9f75 feat(SPEC-GIT-DELIVERY-PROCEDURE-001): complete revision 0.2.2 consumer scope
                                       ea09ca650 wip(t622): preserve partial 0.2.2 revision …
                                       caa601d7c feat(SPEC-GIT-DELIVERY-PROCEDURE-001): define sync merge flag semantics
$ git status --porcelain             → (empty)
$ git show --stat --format='%h parents=%p' HEAD
  3512b9f75 parents=ea09ca650 — acceptance.md 33, plan.md 125, progress.md 54, spec.md 38; 4 files changed
$ git diff --stat b412f8a33 HEAD -- <9 scope files (local), published moai-sync SKILL.md (local), zone-registry.md (local), internal/template, Makefile, docs-site/content>
  → (no output)
```

On HEAD `3512b9f75`, every scope file, the generated `.toml`, both published copies, both emitters, the registry, `Makefile`, and `docs-site/content` are byte-identical to `b412f8a33`. The whole `internal/template` subtree is included in that diff. Every detector result below measured on the working tree is therefore a base-tree measurement.

Guard note: compound Bash scripts that contain the path `manager-git.md` are refused by the worktree session guard ("names git in a form too complex"). Single commands as written in acceptance.md run: AC-028's `awk … .claude/agents/moai/manager-git.md > file` gave `exit=0` for both copies, AC-025's registry detector ran, AC-014's `sed` extractions ran, and `git log --format=%H b412f8a33…..HEAD -- .claude/agents/moai/manager-git.md …` ran.

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency** (placeholder numbering judged as in iteration 1).
  ```
  $ /usr/bin/grep -o -E '^\*\*REQ-GDP-[0-9]{3}\*\*' spec.md | sort | uniq -c
  → 1 each for **REQ-GDP-001** … **REQ-GDP-026** (26 lines, no count other than 1)
  $ /usr/bin/grep -o -E '^### AC-GDP-[0-9]{3}' acceptance.md
  → 001 002 003 004 005 006 007(range 007~012) 013 014 015 016 017(range 017~024) 025 026 027 028 029 030
  $ /usr/bin/grep -o -E '^\| AC-GDP-[0-9]{3}' acceptance.md | wc -l → 23 (16 judged rows + 6 placeholder rows 007–012 + 1 range row 017~024)
  ```
- [PASS] **MP-2 GEARS** (requirement layer, spec.md only). There are 12 judged entries, each with a pattern label that matches its form:
  - Ubiquitous: 001, 002, 004, 013, 014, 025
  - Unwanted: 003, 005, 015, 024
  - Event-driven: 006, 026 (spec.md:L154 "When sync 단계가…", L190 "When `--auto-merge`(또는 폐기된 별칭 `--merge`)가 주어져…")

  The 14 placeholders carry no normative text. ACs (Given-When-Then) were not graded here.
- [PASS] **MP-3 Frontmatter** — spec.md:L2-L13 carry all 12 fields.
  - `version: "0.2.2"` (quoted), `status: draft`, `created: 2026-09-10`, `updated: 2026-09-11`
  - `priority: P1`, `phase: "v3.2.0 target"`, `lifecycle: spec-anchored`, `tags` is a string
  - No rejected aliases.
- [N/A] **MP-4** — not multi-language tooling.
- [PASS] **MP-5 D7** — `grep -o -E 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` finds SPEC-GIT-DELIVERY-PROCEDURE-001 (self), SPEC-MERGE-METHOD-CONFIG-001 (`status: completed`), and SPEC-WORKTREE-BRANCH-GUARD-001 (`status: completed`). No retired, superseded, or archived reference. The iteration-1 SHOULD item about SPEC-AUTH-001 no longer applies: the ID no longer appears.
- [PASS] **MP-6 D8** — `/usr/bin/grep -c -e syscall -e '\[NEEDS CLARIFICATION' spec.md plan.md acceptance.md progress.md` → `0 0 0 0`.
- [PASS] **MP-7** — same command, 0 markers in plan.md. The directory holds no research.md (`ls` → acceptance, plan, progress, spec), which is correct for Tier M.
- Lint cross-check: `mcp__moai__spec_audit(project_root=<this worktree>, filter_spec=SPEC-GIT-DELIVERY-PROCEDURE-001)` → `modern_era_clean: 1, drift_findings: []`.
- Hygiene: Unicode Cf count with perl `\p{Cf}` is `count=0` over all four files. A planted U+200B control gives `control=1`. (`/usr/bin/grep -P` is unsupported on this host.)

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | minor ambiguity | REQ-GDP-025/026 now specify the flag semantics and the per-mode approval rule unambiguously (spec.md:L188, L190), which closes B1's clarity half. One count inconsistency remains in normative text: "범위 파일 열 개" (REQ-GDP-005 L141, REQ-GDP-024 L184, acceptance.md:L17) against nine enumerated scope files (D5) |
| Completeness | 1.00 | full | All sections present, including 5 `### Out of Scope — …` H3s with bullets (L243–278). §E.1 declares the orchestrator-interpretation gap for flags (L292). The consumer sweep found no uncovered surface (see X1–X3 below). The X4 inputs in §D match the tree exactly |
| Testability | 0.50 | several ACs cannot decide PASS/FAIL correctly | Four of 16 judged ACs have no reading step (AC-026, 027, 028, 029). Three of them pass measured wrong implementations that acceptance.md:L55-61 claims they catch (D1, D2, D4), and one fails correct wordings (D3). The other 12 judged ACs have runnable controls and working mutants |
| Traceability | 1.00 | full | Every judged REQ maps to an AC: 001–006→001–006, 013→013, 014→014·030, 015→015, 024→025, 025→026·027, 026→028·029. AC-016 is declared a process check (acceptance.md:L40, L47). Moved IDs are tracked in spec.md §G.1 |

Harmonic mean: 4 / (1/0.75 + 1/1 + 1/0.5 + 1/1) = 4 / 5.333 = **0.75**.

## Defects Found (structured defect-list)

D1. AC026-MERGE-PRESENCE-AND-DIRECTION — acceptance.md:L571-L648 (AC-GDP-026), L55-61 (the claim table) — Severity: major — Class: **blocking**
- **What passes that should fail.** AC-026 passes an implementation that deletes `--merge` from every surface, and one that un-deprecates it or reverses the alias direction.
- **Why.** (ii) only requires that any line containing `--merge` also contains "deprecat" and `--auto-merge`, in any order. No criterion requires `--merge` to remain. Contrast AC-027, which guards `--no-merge` presence with `ac027-dl-count.txt` ≥ 1.
- **What the SPEC requires.** Operator decision (spec.md:L201): "`--merge` 는 `--auto-merge` 의 폐기된 별칭으로 남기고 폐기 경고를 낸다". REQ-GDP-026 (L190) names "폐기된 별칭 `--merge`" as a trigger. acceptance.md:L57 claims AC-026 (ii) catches "`--merge` 를 별개이거나 폐기되지 않은 auto-merge 플래그로 서술".
- **Measured** (scratch fixtures, detector copied verbatim from L632):
  ```
  removal fixture "Modes: auto, force, status, project. Flags: --auto-merge, --skip-mx"
    → (i)=1  (ii)=0                                   → AC-026 PASS  (--merge alias silently dropped)
  m026.md line 2 "- `--auto-merge`: deprecated alias of `--merge`; use `--merge` to auto-merge the PR"      → not hit (PASS)
  m026.md line 3 "- `--merge`: auto-merge the PR after sync (the older `--auto-merge` spelling is deprecated)" → not hit (PASS)
  m026.md line 4 "- `--merge`: alias of `--auto-merge`"                                                     → hit
  m026.md line 5 "- `--merge`: deprecated alias of `--auto-merge` (logs a warning)" (correct)               → not hit
  026 (ii) hits: 4
  ```
  Line 3 is iteration 1's wrong implementation (a), "un-deprecate `--merge` as the opt-in", with one extra clause.
- **Required fix.** Extend AC-026 without adding a new AC (the judged AC count is already at the Tier M cap of 16, spec.md:L233):
  1. Add a presence guard mirroring AC-027: `/usr/bin/grep -c -E '(^|[^A-Za-z-])--merge([^A-Za-z-]|$)' $E/ac026-dl.md` ≥ 1. Where the SPEC intends it, also require ≥ 1 in the `**Flags**:` line and the Supported Flags section.
  2. Either make (ii) directional (require `--merge` … `deprecat` … `alias` … `--auto-merge` in that order on the line), or add a reading record `$E/ac026-reading.md` with the per-line binary question "does this line state that `--merge` is a deprecated alias of `--auto-merge`, and not the reverse or a separate flag?".
  3. List AC-026 in §D.3's reading-gated set.
  4. Record the new mutants (removal, reversed, un-deprecated) in the AC body.

D2. AC028-TEAM-ALL-APPROVALS — acceptance.md:L678-L712 (AC-GDP-028), L59 (claim "team 모드가 승인 없이 병합 → AC-GDP-028 (a)·(b)") — Severity: major — Class: **blocking**
- **What passes that should fail.** Team-mode wording that drops or weakens the approval condition. (a) only needs "team mode", `--auto-merge`, and the substring "approv" on one line; (b) knows four negative phrasings. REQ-GDP-026 (L190) requires "team 모드는 전원 승인을 조건으로 병합".
- **Measured:**
  ```
  m028.md line 1 "In team mode, `--auto-merge` merges once CI passes; approvals are optional."  → (a) hit, (b) not hit → PASS
  m028.md line 2 "In team mode, `--auto-merge` merges after at least one approval."             → (a) hit, (b) not hit → PASS
  m028.md line 3 "In team mode, `--auto-merge` merges only after all approvals are obtained." (correct) → (a) hit, (b) not hit
  028 (a): 1 2 3    028 (b): (none)
  ```
- **Required fix.** Tighten (a) to require the "all" condition (e.g. `/[Aa]ll( required)? approv/` on the team-mode line). Either add `optional|not needed|at least one|any approv|single approv` to (b), or add `$E/ac028-reading.md` to the §D.3 reading-gated set. Keep the change inside AC-028.

D3. AC029-B-FALSE-FAIL — acceptance.md:L714-L736 (AC-GDP-029 (b)) — Severity: minor — Class: optional
- **Problem.** (b) fails correct personal/manual wordings that avoid the four whitelisted phrasings, and passes a wording that adds a review requirement without the token "approv".
- **Measured:**
  ```
  m028.md line 5 "In personal and manual modes, `--auto-merge` merges without requiring approval." (correct) → (b) hit → FAIL
  m028.md line 6 "In personal and manual modes, `--auto-merge` merges; approval is not needed."       (correct) → (b) hit → FAIL
  m028.md line 4 "In personal and manual modes, `--auto-merge` merges after a code review."          (wrong)   → (a) hit, (b) not hit → PASS
  029 (a): 4 5 6    029 (b): 5 6
  ```
- **Why optional.** The false-fail side fails closed, and plan.md §B.14 steers the wording.
- **Required fix.**
  - Either name the accepted personal/manual phrasings in plan.md §B.14 / M1 so run-phase writes one of them, or widen the (b) exemptions (`without requiring approv`, `not needed`).
  - Add `review` to the (b) trigger, or cover AC-029 with the same reading record as D2.

D4. AC027-EFFECT-VERBS — acceptance.md:L650-L676 (AC-GDP-027 (ii)), L58 (claim "`--no-merge` 가 여전히 동작을 바꿈(건너뜀·트리거 조건) → AC-GDP-027 (i)·(ii)") — Severity: minor — Class: optional
- **Problem.** Lines that say "Deprecated no-op" but also assert an effect pass both (i) and (ii).
- **Measured:**
  ```
  "- `--no-merge`: Deprecated no-op; disables auto-merge for this run."                        → (i) none, (ii) none → PASS
  "- `--no-merge`: Deprecated no-op that overrides `--auto-merge`."                              → PASS
  "- `--no-merge`: Deprecated no-op; turns off merging even when `--auto-merge` is given."      → PASS
  ```
- **Why optional.** Each line contradicts itself ("no-op" plus an effect), which makes it a less likely implementation than D1/D2. It still falsifies the L58 claim.
- **Required fix.** Add `disabl|overrid|turns? off|suppress|bypass|cancel` to (ii), or fold AC-027 into the reading record proposed in D1/D2.

D5. SCOPE-FILE-COUNT — spec.md:L47, L88, L141 (REQ-GDP-005), L184 (REQ-GDP-024), L234, L289; acceptance.md:L17, L38, L231, L335, L454, L532 — Severity: minor — Class: optional (clarity)
- **Problem.** "범위 파일 열 개" is asserted in two REQs and six AC/convention sites, but every enumeration names nine path pairs: manager-git, agent-common-protocol, delivery, doc-execution, SKILL, reference, quality-gates-context, workflows/sync.md, and the command source `.md`/`.md.tmpl`.
- **Where the SPEC contradicts itself.**
  - AC-005 says "나머지 여덟 파일" (9 − manager-git).
  - spec.md:L289 says "아홉 파일" and lists eight names.
  - §C.4 L234 computes "열 개 × 로컬·템플릿 20 + 1 = 21". Nine pairs give 18 + 1 = 19 (+2 in case B).
- **Effect.** A reader of REQ-GDP-024 cannot tell whether a tenth file is protected.
- **Required fix.** Change "열 개" to "아홉 개" everywhere (and "나머지 아홉" to "나머지 여덟" where agent-common-protocol is excluded), correct §C.4 to 19 (+2), or name the tenth file if one was intended.

D6. STALE-HEAD-NOTE — spec.md §E.2 last bullet (L303, "HEAD `caa601d7c` 에서 확인") — Severity: minor — Class: optional
- **Problem.** Every other 0.2.2 measurement names `ea09ca650`, and the tree is now `3512b9f75`.
- **Measured.** `git diff --stat b412f8a33 HEAD -- <scope…>` gives no output on `3512b9f75`, so the premise still holds.
- **Required fix.** Update the HEAD citation.

D7. VARIABLE-SCOPE-CONVENTION — acceptance.md:L7-L18 (공통 변수와 명령 관례) — Severity: minor — Class: optional
- **Problem.** `BASE`, `E`, and `T` are defined once, but each Bash invocation is a fresh process, so a judge command run on its own sees them unset.
- **Consequences when unset.**
  - `git log --format=%H $BASE..HEAD` becomes `..HEAD`, which is empty. AC-016 and AC-030 `test -s` guards then fail closed, so this is safe.
  - `git diff $BASE -- $T/` becomes `git diff -- /`, which errors rather than passing.
- **Measured.** The guard accepts the same-invocation form: `BASE=b412f8a…; git log --format=%H $BASE..HEAD -- Makefile` → ran, no refusal.
- **Required fix.** One sentence: every judge command carries the three assignments in the same invocation. Also note that compound scripts naming `manager-git.md` are refused by the guard, so run commands one per invocation.

D8. AC016-TEMPLATE-ACP-ORDER — acceptance.md:L507-L523 — Severity: minor — Class: optional
- **Problem.** The "last commit" anchor reads only the local `agent-common-protocol.md` path, and the after-list omits the template `agent-common-protocol.md`. If the template copy were committed in a later commit than the local one, the check still passes.
- **Also.** The `<ac016-acp-commit.txt 의 SHA>` placeholder needs manual substitution.
- **Required fix.** Anchor on `git log -1 … -- <local acp> <template acp>` and state the substitution step explicitly.

## Iteration-1 items — re-verification

- **B1 — RESOLVED (semantics), PARTIALLY RESOLVED (AC resistance).**
  - REQ-GDP-025 (L188) states the exposed flag, the alias status, and the no-op status, together with prohibitions. REQ-GDP-026 (L190) states team = all approvals and personal/manual = no approval condition, requires both documents to carry mode-named sentences, and keeps CI and conflict checks.
  - The decided surface is in scope, and the §C.1 decision rows (L201–203) match the dispatch.
  - §E.1 L292 carries the gap that real sync execution is not observed.
  - The mutant sweep shows AC-026/028 still pass report-named wrong implementations (D1, D2).
- **N1 — RESOLVED.**
  - §D.3 (L836-837) makes AC-001 and AC-006 non-PASS without complete reading files.
  - The AC-006 detector was extended (L270-273). Re-measured on base: `awk <extended> dl.md` → `8 20`; doc-execution subsection (9 lines) hit at line 8 "This affects auto-merge behavior: worktree contexts defaul…". Both match acceptance.md:L295-297.
- **N2 — RESOLVED.**
  - Positive control: `/usr/bin/grep -c -E 'file: \.claude/rules/moai/core/agent-common-protocol\.md' <registry L> <registry T>` → `13` / `13`. Others detector as written → `0` / `0` (exit 1).
  - Own mutant fixture (10 `file:` lines, the literal manager-git path written via perl `\x67`): hits `1 2 3 4 5 6 7 8`; agent-common-protocol (line 9) and the quoted path (line 10) not hit.
  - The Frozen judge is now `^[-+].*`. Own fixture with `-+[ZONE:Frozen] leading plus content` → hits `1 2 5` (removed, added, removed-starting-with-plus); context line and `@@` header not hit.
- **N3 — NOT RE-VERIFIED.** The fix was a lead action on card t658's text, not a SPEC edit. §E.2 L298 still records the ordering risk. The card text was not read in this iteration (Gap).
- **N4 — RESOLVED.**
  - AC-014 (L437-446) now diffs both the `## Synchronization` and `## PR Auto-Merge` sections between the template `.md` and the `.toml`.
  - Re-measured: sync 11/11 lines `diff exit=0`; PR Auto-Merge 9/9 lines `diff exit=0`.
  - `grep -c -F 'gh pr merge <PR> --<merge_method> --delete-branch' manager-git.toml` → `0`, as expected at base.
- **N5 — RESOLVED.** AC-001 now has explicit toml commands (L101-110): `sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' $T/.codex/…/manager-git.toml`, `test -s`, the paragraph count, and autofail.
- **N6 — RESOLVED.** spec.md:L224 now reads "폐기 예정으로 분류된 키에 첫 소비자를 만든다".
- **N7 — RESOLVED.** SPEC-AUTH-001 is no longer referenced.

## New since iteration 1 — verification

**X1–X3 fragments (AC-026/027), base.** Measured on both copies with the acceptance extraction commands verbatim:
```
lines:  skill 1; ref 3; qgc-args 4; qgc-flags 6; sync 1; dl 52; sync-usage 1; hint 1; dl-next 24   (local = template)
(i) --auto-merge: all 9 fragments 0                                                                 → RED
026 (ii): skill:1 ref:2 qgc-args:4 qgc-flags:4 sync:1 dl:9 dl:20 sync-usage:1 hint:1 dl-next:9      → 10 lines, RED
027 dl --no-merge count: 2;  027 (i): dl:8 dl:19;  027 (ii): dl:8 dl:19                             → RED
028/029 (a) on dl: 0 / 0
dl-next L-T diff exit=0; hint L-T diff exit=0; dl L-T diff exit=0
```
All match acceptance.md:L608-611 and L661. The X1–X3 fixture (L641-646) is consistent with the detector.

**Consumer completeness sweep** (not only the author's list):
```
/usr/bin/grep -rn -E -e '--no-merge' -e '(^|[^A-Za-z-])--merge([^A-Za-z-]|$)' .claude internal/template/templates CLAUDE.md AGENTS.md README*.md --include='*.md' --include='*.tmpl' --include='*.toml' --include='*.yaml'
  | grep -v -e '^.claude/worktrees/' -e agent-memory -e 'gh pr merge'  → exit=0, 26 lines
```
- 24 lines are the in-scope sites: SKILL:140, reference:161, sync.md 95/114 (L) and 85/104 (T), delivery 337/338/348/349/404, qgc 30/101, command source :3, in both copies.
- 2 lines are `hns-release-specialist.md` 371/379, which use the "다른 뜻" meaning and are recorded in §D L266.
- No uncovered `/moai sync` flag consumer.

**AC-GDP-030 grounding** (read, not executed):
- `loader.go:4-6`: "argument-hint, allowed-tools (the latter two are Claude-only keys and are NOT carried into the published skill)" ✓.
- `Makefile:51-52` `commands-emit` = `COMMAND_EMIT_UPDATE=1 go test ./internal/template/commandemit/... -run TestGoldenCommittedArtifactsMatchEmission` ✓.
- `Makefile:58-60` `commands-emit-check` scrubs the env var, runs with `-count=1`, and prints "command-skill drift" before `exit 1` ✓.
- `golden_test.go:28` `templatesDir = "../templates"` ✓.
- `golden_test.go:57-90`: the update branch writes `committedSkillPath(p)` under `templatesDir`; the check branch compares **sha256** of committed vs emitted bytes. The positive control `printf '\n' >> …/SKILL.md` therefore produces a mismatch, so control (c) is valid.
- `:94-106` `TestCommandSourcesUnmodified` compares before/after hashes in one run ✓.
- `emit.go:122-130` `renderSkill` writes the header, `name`, `description`, and body only ✓.
- Published template copy is 7 lines, with no `merge`.
- `git ls-files` lists all four paths. `git status --porcelain -- internal/template/templates/.agents/skills/ .agents/skills/` → empty, so control-clean is achievable.

**AC-GDP-030 subset check mutants** (own fixtures): src{aaa1111} art{bbb2222} → `exit=0`, prints `bbb2222` (FAIL detected); src{aaa1111} art{aaa1111} → `exit=1`; empty src with art{aaa1111} → `exit=0`, prints `aaa1111` (FAIL); src{aaa1111} with empty art → `exit=1` (case A). This matches L808, and case B adds `test -s artifact-commits`.

Minor observation, folded into no defect: case (B) "변경 경로가 두 파일뿐" is judged by reading `ac030-changed.txt`, not by a count command.

**Command-source parity (`.md` / `.md.tmpl`).**
- The base diff is `2c2` only (spec.md:L87); the `argument-hint` lines are identical (the L-T fragment diff above gives exit=0).
- AC-013 strips diff range headers before the body comparison, so a divergent line-3 edit turns `2c2` into `2,3c2,3` with an extra body pair, which is caught. It adds a separate hint diff and its own mutant (L413).
- The design is sound for files whose names differ.

**AC-013/015/016/025 file lists.**
- AC-013 covers all nine pairs plus the published copy.
- AC-015 diffs the whole template root (and adds the non-empty guard plus the added-line count ≥ 1).
- AC-016's after-list has 16 paths (8 local + 8 template, command source included).
- AC-025's scope diff has 18 paths.
- The AC-025 registry pattern change to `manager-[g]it` runs under the guard and matches the literal path (fixture above).

**X4 §D counts.** `git grep -n -e --merge -e --no-merge -e --auto-merge b412f8a33 -- <8 docs-site paths>` →
- en moai-sync 5 (80, 82, 188, 303, 480)
- ko 4 (80, 82, 299, 473)
- ja 6 (80, 82, 299, 476, 477, 482)
- zh 9 (77, 92, 97, 108, 189, 304, 485, 486, 491)
- what-is-moai-adk 1 each (en 388, ko 391, ja 391, zh 389)
- `--auto-merge` hits 0

Exact match with the spec.md:L254-256 table. The quoted line texts (ko 473, en 480, ja 477/482, zh 486/491) match.

**Cross-document consistency.**
- Judged counts are 12 / 16 in spec §C.4 L232-233, plan §A L7, progress §E.1 counts, and acceptance "판정 대상: 16개" L47.
- progress §E.1 numbering arithmetic: REQ 26 = 12 + 12 + 2; AC 30 = 16 + 12 + 2 ✓.
- Milestone placement: `commands-emit` sits in M1 with the command-source commit (plan L92-96, §D L55-56), `agents-emit` in M4 (L111-112), and the always-loaded rule edit in M5 as the last instruction edit (L119-123), with M6 edit-free. This agrees with AC-016, AC-030, and §D.3.
- Pre-check §C steps 4–5 carry both emit-check baselines and the AC-030 control (c) before the source edit ✓.
- The only inconsistencies found are D5 and D6.

## Regression Check (Iteration 2)

Defects from reduced iteration 1:
- B1: [RESOLVED at the requirement layer / UNRESOLVED at the verification layer] — see D1, D2.
- N1: [RESOLVED] — §D.3 reading gate plus extended detector, re-measured `8 20`.
- N2: [RESOLVED] — control 13/13, own registry mutant 8/8, Frozen judge hits a removed line that starts with `+`.
- N3: [NOT RE-VERIFIED] — lead-side card text.
- N4: [RESOLVED] — two section diffs, 11/11 and 9/9, exit 0.
- N5: [RESOLVED] — toml commands present.
- N6: [RESOLVED] — L224 wording.
- N7: [RESOLVED] — reference removed.

Delta since the audited 0.2.0 (`git diff 24e20ec50 HEAD -- acceptance.md`, 22 hunks):
- AC-002, AC-003, and AC-004 changed only in expectation comments and mutant labels.
- AC-005's control and judge file lists grew to the new scope files.
- AC-006 gained the extended detector and reading record.
- No regression to a previously PASS criterion was observed.

No defect is stagnant across three iterations.

## Recommendation

FAIL on two blocking verification defects. The requirement layer is sound; the fixes are confined to acceptance.md and add no AC, so Tier M's 16-AC cap holds.

1. **D1 (AC-026).**
   - Add a `--merge` presence guard (≥ 1 in the delivery Step 3.4 fragment; also in Flags/Supported Flags if intended).
   - Make (ii) directional or reading-gated (`$E/ac026-reading.md`).
   - Record the removal, reversed-alias, and un-deprecated mutants.
2. **D2 (AC-028).**
   - Require "all … approv" on the team-mode line.
   - Extend (b) with `optional|not needed|at least one|any approv`, or reading-gate it.
   - Record the "approvals are optional" and "at least one approval" mutants.
3. Optional in the same pass:
   - D3: accepted phrasings or widened exemptions for AC-029 (b), plus a `review` trigger.
   - D4: effect verbs for AC-027 (ii).
   - D5: "열 개" → "아홉 개" and §C.4 19 (+2).
   - D6: HEAD citation.
   - D7: same-invocation variable sentence.
   - D8: AC-016 template-acp anchor.
4. Update §D.3's reading-gated list if reading records are chosen for AC-026–029.

**Iteration ceiling.** Tier M allows 2 plan-audit iterations (`harness.yaml:77`), and this is iteration 2. Per the retry-loop contract the orchestrator escalates to the operator. The options are PASS-with-debt (documenting D1 and D2), a targeted acceptance.md fix followed by an explicit override for a delta-scoped iteration 3, or a scope change. Both blocking defects are small, mechanical, and confined to one file.

## Gaps (not observed)

- `make commands-emit`, `make commands-emit-check`, `make agents-emit-check`, and `go test` were not run (dispatch prohibits make and go test). The AC-030 control (c) validity rests on reading `golden_test.go:57-90` (sha256 compare), not on execution.
- Card t658's text was not read (N3).
- AC-015's hex-token fixture and AC-002's block detectors were not re-executed in this iteration. Their hunks since 0.2.0 changed only comments and file lists (read from the delta diff), and iteration 1 executed them.
- `mcp__moai__audit_multi` / codex / GLM second opinions were not requested; the project `audit_model` was not read. This is a Claude-only verdict.
- Whether the orchestrator actually emits a deprecation warning for `--merge`/`--no-merge` at runtime is not observable from the SPEC (already declared in spec.md §E.1).

## Residual risk

- Even after the D1 and D2 fixes, AC-026–029 remain line-level lexical judges. A wrong implementation phrased outside the anticipated vocabulary can still pass, unless those ACs become reading-gated like AC-001 and AC-006.
- The AC-030 control temporarily mutates a tracked published file. A concurrent writer in the same worktree between `printf` and `cp` would mix. `cmp` and `git status` detect it after the fact but do not prevent it.
- The line citations assume scope files stay byte-identical to `b412f8a33`. That holds on `3512b9f75` but must be re-measured if run-phase absorbs develop first (spec.md §E.2).
- Until the X4 follow-up docs card lands, the published docs describe the old flags and defaults (spec.md §D) — an accepted, recorded risk.
