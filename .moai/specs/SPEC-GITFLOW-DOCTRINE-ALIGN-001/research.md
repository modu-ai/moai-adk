# research.md — SPEC-GITFLOW-DOCTRINE-ALIGN-001

Provenance: all measurements below taken on THIS worktree (`.claude/worktrees/t310`, branch `WT-git-doctrine-align`), base commit **d29b8942e**, 2026-08-27 session. Per the repo's tree-pinning discipline, line numbers are pinned to d29b8942e; a later rebase MUST re-measure before quoting.

## §A Investigation basis

Defect inventory supplied by the lane dispatch and re-measured directly by manager-spec (not carried over on trust). Every entry below was observed with the quoted command in this session.

## §B Target 1 — `.moai/docs/git-workflow-doctrine.md` (424 lines)

| # | Location (d29b8942e) | Observed defect | Command + verbatim observation |
|---|---|---|---|
| D1 | Line 15 | Retired rationale asserted as standing fact: "…Gitflow의 `develop` 이중 관리 부담을 상회하는 장점이 부재…" — contradicts the header's own line 7 ("서두의 Gitflow 기각 판단도 이 전환으로 뒤집혔다"). | `grep -n "장점이 부재"` → exactly 1 hit, line 15, no `[RETIRED …]` marker on or near it |
| D2 | Line 52 (§18.0.1 금지 사항) | `- ❌ \`develop\` 브랜치 생성 (Gitflow 패턴)` — inverse of canonical rule #1. | `grep -nE '❌.*develop.*브랜치 생성'` → line 52, line 351 |
| D3 | §18.3.1 (~lines 107–134) | Tier-based PR routing routes ALL tiers via feature branch + `gh pr create` -> self-merge against `main`, with a `[RETIRED 2026-07-20]` annotation describing the intermediate (pre-git-flow) model. Section-anchored sweep: `sed -n '/### §18.3.1/,/^### §18.4/p' \| grep -cE 'self-merge\|gh pr create'` → **7** hits, none aware of the 2026-08-27 model. |
| D4 | Line 351 (§18.10 [HARD]) | `- ❌ \`develop\` 브랜치 생성 (Gitflow 패턴 금지)` | same grep as D2 |

Annotation-marker census (green-path anchor): `grep -F -c '[RETIRED 2026-08-27]'` over the three targets → **0 / 0 / 0** (no 2026-08-27 markers exist yet); `[RETIRED 2026-07-20]` markers DO exist in both doctrine files, confirming the dated-annotation convention this SPEC extends. NOTE (plan-audit iter-1 D2): the `-F` flag is REQUIRED — the bare pattern `[RETIRED 2026-08-27]` read as a regex bracket-expression contains an invalid reverse character range and exits 2 under `/usr/bin/grep` (BSD) and ugrep alike; always re-measure with the fixed-string form.

## §C Target 2 — `.moai/docs/git-local-workflow-doctrine.md` (207 lines)

| # | Location | Observed defect |
|---|---|---|
| D5 | §23.7 bullet, line 150 | `- [HARD] **(2026-07-20 신규) PR-mandatory: 모든 tier (S/M/L) 변경은 PR 경유.** …` — presented as standing [HARD]; in the new model daily card changes produce NO PR. Substantive core worth retaining: `enforce_admins: true` blocks main direct push; self-merge conditions apply to whatever PR path remains (release). |
| D6 | §23.9 (lines 154–181) | "Tier-based PR Routing" heading + lead sentence "모든 경우 PR 생성·머지는 `manager-git` 서브에이전트가 담당한다"; four-row tier table (S/M/L/`--pr`) all premised on per-change PRs; routing flow items 1–3; Late-Branch references. All superseded for card work. |
| D9 | §23.7 bullet, line 151 (added at plan-audit iter-1, auditor finding D1) | `- [HARD] **\`git push origin main\` 금지** — 시도 시 server-side rejected. 항상 feat/fix/chore/docs/release 브랜치 → \`gh pr create\` → CI green → \`gh pr merge\` 흐름.` — its first half (server-side rejection of main direct push) is TRUE under the new model and survives; its route-prescription tail (`항상 … 흐름`) presents PR-opening as standing procedure for EVERY change type including cards — inverted, and inside the same [HARD] list as D5. Left out of the original inventory; caught by plan-audit iter-1. RED-now (measured this session): `sed -n '/^### §23.7/,/^### §23.9/p' $L \| grep -n '항상'` → single hit at range-line 9 (= file line 151), un-struck. |

Sections verified SOUND (no edit): §23.2 branch-protection table (describes `main` protection, which survives unchanged); §23.3–§23.6 operational patterns (A4/A5/A6, Late-Branch Phase D — generic git procedures); §23 pre-transition blockquote at line 14 (correctly frames the 2026-07-20 change historically).

## §D Target 3 — `.claude/rules/local/repo-local-pr-policy.md` (14 lines)

| # | Location | Observed defect |
|---|---|---|
| D7 | Line 10 | `ALL tiers (S / M / L) use **Route B (PR)**: work lands on a feature branch and merges via PR` — the origin premise this file exists to assert, now inverted for card work. |
| D8 | Line 8 + whole body | Body makes NO mention of `develop` anywhere (0 occurrences measured): the file cannot tell a rule-loading session that cards branch from and merge into `develop`. What SURVIVES verbatim-in-substance: `enforce_admins: true` + required PR on `main`, direct-push rejection for admins, and the template-neutrality paragraph (local-only, not mirrored). |

Note the frontmatter `paths:` reaches `.moai/specs/**`, `run.md`, `sync.md`, `moai.md` — this rule loads on exactly the surfaces a misrouted lane acts through, raising repair priority above ordinary prose staleness.

## §E Convention study — how this repo annotates superseded doctrine

Both doctrine files already carry the pattern this SPEC must extend (left by the 2026-07-20 PR-mandatory change):

- Header notice: `> **[HARD] POLICY CHANGE (<date>) — …**` blockquote at the affected section's top, naming superseded vs still-binding clauses (git-local-workflow-doctrine.md line 14; git-workflow-doctrine.md line 5).
- Inline retirement: `**[RETIRED 2026-07-20]**` trailing annotations and `~~strike-through~~` wrapping of the retired sentence, PRESERVING the old text as historical record (e.g., git-local-workflow-doctrine.md line 152, git-workflow-doctrine.md line 118).

Decision encoded into plan.md: prefer ANNOTATION + correction where history is explanatory (routing tables, rationale prose); plain in-place replacement only where a bullet would stand alone as a dangerous falsehood (the two ❌ develop bullets — a naked "- ❌ develop 브랜치 생성" must not survive even behind strike-through in a forbidden LIST, because list context gives strike-through no protective frame; the list slot must hold a TRUE prohibition instead).

## §F Additional findings — reported, deliberately OUT of scope

Found while sweeping the full bodies; each is covered reader-side by the existing header notice, and bundling them would grow this card's diff beyond the dispatched defect set:

- **F-1** git-workflow-doctrine.md §18.1 branch diagram (lines 58–72): ASCII art shows everything branching off `main` only; no `develop` rail.
- **F-2** §18.3 merge-strategy table (lines 90–100): rows named "feature/fix/chore/docs → **main**"; under git-flow, feature/card merges land in `develop`.
- **F-3** §18.5 hotfix workflow (lines 145–169): branches hotfix from the production TAG and PRs to `main` — defensible post-transition for main-touching hotfixes, but the CARD-equivalent flow is undocumented there.
- **F-4** §18.8 release process, line 299 (measured): `git checkout -b release/vX.Y.Z main` — release branches now cut from `develop` (canonical §10 of gitflow-lane-protocol).
- **F-5** git-local-workflow-doctrine.md §23.8 (line 191): multi-session mitigation cites `git rev-list --count --left-right origin/main...HEAD`; lanes now measure divergence against `origin/develop`.
- **F-6** OPEN QUESTION (non-blocking): whether an operator-explicit `--pr` card escape hatch survives. Canonical files state card work makes no PRs; SPEC body text follows the canon. If the operator later wants the hatch back, that is an amendment — not a reason to plant ambiguity here.

- **F-7** (plan-audit iter-1 finding D7): the H1/H2 title framings remain after M1–M3 — `$D` titled "Git Workflow Doctrine — Enhanced GitHub Flow", `$L` H2 suffixed "(PR-mandatory 1-person OSS)". Consciously EXCLUDED: each file's own 2026-08-27 header notice already names-and-limits those framings ("초과분"), and renaming headings risks breaking external cross-references (the no-renumber rule this SPEC adopts).

If a follow-up card picks up F-1..F-5 (and F-7's assessment), suggest SPEC-GITFLOW-DOCTRINE-SWEEP-002.

## §G design.md necessity ruling

NOT produced. Ruling (recorded per the lane dispatch requirement): pure documentation alignment has no design space — no interfaces, no data model, no behavioral contracts; the artifact only edits prose in three known files per a convention the files themselves already demonstrate (§E above). Repo precedent agrees: recent docs-only SPEC directories (SPEC-CC-DOCS-ALIGNMENT-001, SPEC-V3R6-WORKFLOW-DOCS-001, both Tier M) shipped spec/plan/acceptance(+progress) with no design.md. design.md here would be ceremony, contradicting the lean-artifact intent of the Tier taxonomy.
