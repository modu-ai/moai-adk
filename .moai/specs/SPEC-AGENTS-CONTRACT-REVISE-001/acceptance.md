# SPEC-AGENTS-CONTRACT-REVISE-001 — Acceptance Criteria

Tier M · 16 ACs, 1:1 with REQ-ACR-001…016. Every AC is binary-testable; each verification names its command. All measurements are taken in the card worktree at close time and attributed to the then-current HEAD SHA (tree pin, never a moving branch name).

## AC-ACR-001 — §4.0 current card policy (REQ-ACR-001)

**Given** the revised `AGENTS.local.md`, **When** the flow-doc checks run, **Then**:
- `grep -n "§4.0" AGENTS.local.md` shows the current card-policy section carrying: draft-PR delivery for the verified head, recorded handoff fields (head SHA, evidence, review/repair owner, lease/test-slot disposition), receiver-side confirmation before the next independent card, PR-is-not-done closure, and the preservation-push clarification.
- `grep -n "receiver\|인수" AGENTS.local.md` finds the confirmation requirement (the lane/receiver confirmation completion criterion).
- The §4.0-H historical block exists and its header states the superseded clauses carry no execution authority over current card delivery.

## AC-ACR-002 — Leader boundary (REQ-ACR-002)

**Given** the revised `AGENTS.md` and `AGENTS.local.md`, **When**:
- `grep -n "handle simple, bounded operations directly\|Do the bounded work inline" AGENTS.md internal/template/templates/AGENTS.md.tmpl` → **0 hits**.
- `grep -n "Permitted direct execution" AGENTS.local.md` → **0 hits** (replaced by the assigned-worker boundary).
- `grep -n "never directly implement\|shall not implement, repair, build" AGENTS.md AGENTS.local.md` → hits present in §0, the capability-table fallback, and §16.
- `grep -inE "emergency|worker-absence|(^|[^-])trivial" AGENTS.md AGENTS.local.md` → every hit is inside a "no exception" sentence, never an exception grant (the `(^|[^-])` guard excludes "non-trivial", so the pre-existing AGENTS.md:111/211/230/358 occurrences — verified 0 hits under this pattern — cannot redden the check for the wrong reason).

## AC-ACR-003 — Review contract rewrite (REQ-ACR-003)

**Given** the revised `AGENTS.md` §4 and its mirror, **When**:
- `grep -n "CodeRabbit" AGENTS.md internal/template/templates/AGENTS.md.tmpl` → **0 hits** in the review-contract clause (the producer-neutral contract replaces it).
- The clause requires: exact current head SHA, configured review producer, completed review result, evidence URL; missing/skipped/failed/old-head review is a gap.
- A cross-reference to the standing Codex contract document exists, and no measurement recipe is duplicated.

## AC-ACR-004 — Review-gate policy block (REQ-ACR-004)

**Given** the revised flow documents, **When**:
- `grep -n "codex-review-state" <flow docs>` → the merge rule is stated: both `code_review` and `security_review` `completed` with zero unresolved findings **on the current head SHA**; the re-request rule on head change; unprotected base is no exception; leader coordination-only with the #1841 incident named.
- The trigger narration is observation-scoped: the ready transition (#1841) and the `@codex review` comment (#1831, #1832, #1834 — #1832/#1834 recorded results) both appear as observed triggers with no exclusive-start claim, and the head-changed-without-automatic-re-review observation (#1836, #1842) is recorded.
- The official-trigger enumeration is present (four triggers per the PR #1844 Codex summary comment observed 2026-10-19T06:05Z: open-for-review, draft→ready, `@codex review`, `@codex security review`, with 👀 in-progress / 👍 findings-free markers), and `grep -n "Failed" <flow docs>` finds the state-set rule: any state other than `completed` is not a pass and is a re-request target.
- `grep -in "only after the draft" <flow docs>` → **0 hits** (the exclusive ready-transition-start phrasing does not survive).
- `grep -n "1841" <flow docs>` → the incident is cited with the evidence paths (`inputs/REPORT.md`, `inputs/native-codex-rest-comment.json`).
- `grep -n "t1868" <flow docs>` → every gate mention carries the in-progress/limitation marker; `grep -in "gate.*implemented\|implemented.*gate" <flow docs>` → **0 hits** presenting the gate as shipped tooling.

## AC-ACR-005 — Superseded-flow confinement (conflict grep; REQ-ACR-005)

**Given** the revised tree, **When** the conflict grep runs:

```bash
grep -rn -e "lanes do not push" -e "do NOT push" -e "no card-level PR" \
  -e "leader collects lane merge SHA" -e "bulk action by the leader" \
  -e "local main.*landing" -e "waits until the batch" \
  AGENTS.local.md .claude/rules/local/repo-local-pr-policy.md \
  .claude/rules/local/gitflow-lane-protocol.md \
  .claude/rules/moai/workflow/factory-dispatch.md \
  .claude/rules/moai/workflow/factory-dispatch-core.md \
  .claude/rules/moai/workflow/factory-dispatch-mechanics.md
```

**Then** the count of hits outside transition-scoped historical blocks (each surviving hit sits in a section whose heading or immediately preceding marker states historical/transition scope — "전환 전/후", "Historical", "superseded", "not the current procedure") is **0**, and the swept hit count is reported (a zero-hit sweep is reported as "not measurable", not as pass). Mirrors under `internal/template/templates/` are swept with the same predicate.

## AC-ACR-006 — Batch-release retention + merge-flag accuracy (REQ-ACR-006)

**Given** the revised docs, **When**:
- The batch-release procedure (`release/main-batch-YYYYMMDD`, merge-commit, leader-executed) is present as the current release-path procedure with the until-#1830 scope note.
- `grep -rn "MERGE_FLAG\|--merge\|--squash" AGENTS.local.md .moai/docs/git-workflow-doctrine.md .moai/docs/git-local-workflow-doctrine.md` → every merge-flag statement includes the hotfix/* branch of the rule (measured source: `.github/workflows/auto-merge.yml:341-345`).
- `grep -rn "release/\*" <docs> | grep -v hotfix` → **0 hits** where the statement claims the complete rule.

## AC-ACR-007 — Policy-vs-implementation distinction (REQ-ACR-007)

**Given** the revised docs, **When**:
- `grep -rn "factory complete" <flow docs>` → any draft-PR claim about `moai factory complete` carries the t1882 not-yet-landed marker; the manual `gh pr create --draft` path is named as operative.
- `grep -rn "manual" <deployment docs>` → manual-by-default / staged-independent-opt-in is stated as policy; no sentence presents it as the shipped default while t1895 is open.
- `grep -rin "shipped default\|already implements" <flow docs>` → **0 hits** claiming unimplemented capability as shipped.

## AC-ACR-008 — Auto-merge mechanism accuracy and conditions (REQ-ACR-008)

**Given** the revised docs, **When** the auto-merge description is read against `.github/workflows/auto-merge.yml` (trigger L7-10; condition L23-26; merge call L355): **Then** the description names the `workflow_run` CI-completed trigger, no draft filter, no config gate, and `--match-head-commit`; **and**:
- The two non-own-PR conditions are stated together with a both-required qualifier: `grep -in "both\|모두" <docs>` at the auto-merge clause finds test verification AND explicit approval required together; a statement presenting either condition alone as sufficient fails.
- `grep -in "connected\|연결된" <docs>` shows identity resolved from the actual connected account; `grep -in "hardcod" <docs>` shows the no-hardcoded-login rule.
- Own-PR gate retention is stated: latest-SHA required CI, real Codex review, resolved findings, protection rules, third-party approvals.
- `grep -rn "delete-branch\|delete_branch_on_merge" <docs>` finds the mismatch recorded as a limitation with the repository-stays-public / no-automatic-branch-deletion rule (own PRs included).
- `grep -in "local operation choice\|로컬 운영 선택" <docs>` finds the not-a-product-default-change scope note.

## AC-ACR-009 — Size diet, leader-corrected model (REQ-ACR-009)

**Given** the restructured `AGENTS.local.md`, **When**:
- The decision-index measurement table exists with the always-loaded surfaces measured at the PRE-CARD baseline commit `9b1dcde024` (AGENTS.md, AGENTS.local.md, output style, always-loaded rules) and the projection pinned to that baseline: (252,650 pre-card total − this card's AGENTS.local.md reduction − t1887 core-rules share) ≤ **140,000** — the total is never re-measured into the formula at close time, so this card's own reduction cannot relax its own gate; the derived share bound and the "리더 정정" provenance are stated.
- The four named chunks are moved into companion docs (before/after section measurements recorded); the withdrawn figure appears nowhere as a target (`grep -rn "12,000" AGENTS.local.md <SPEC dir>` → only correction-provenance notes, never a governing target).
- The reconciliation table exists (heading present) and every row maps original section → new location; moved-clause count from the table equals the pointer count found by grep.
- For each [HARD] clause that moved, a one-line pointer or essence remains at the original location; no obligation is deleted or weakened (the [HARD] essence inventory recorded at M1 start does not decrease).
- Each companion doc's `paths:` glob is exercised against the files it claims to scope: for each claimed file, the glob matches (list the file + glob pair per companion in the evidence). A glob that fails to load for a claimed file fails this AC.

## AC-ACR-010 — Pointer and mechanism fixes (REQ-ACR-010)

**Given** the revised `AGENTS.local.md`, **When**:
- `grep -n "templates/CLAUDE.md" AGENTS.local.md` → **0 hits**; `grep -n "AGENTS.md.tmpl" AGENTS.local.md` → the template-source entry present.
- `grep -n "CLAUDE.md Section 4" AGENTS.local.md` → **0 hits**; `grep -n "delegation.yaml" AGENTS.local.md` → the catalogue pointer present.
- The quick-start/embed description names `embed_manifest_gen.go` and the explicit emit targets (`make agents-emit`, `make commands-emit`, `make embed-manifest`) with `make build` as drift-check-and-compile, matching `.moai/docs/local-dev-guide.md` §2b-2c as revised.

## AC-ACR-011 — Historical marking (REQ-ACR-011)

**Given** the revised `.moai/docs/git-workflow-doctrine.md` and `git-local-workflow-doctrine.md`, **When**: `grep -n "historical\|히스토리\|History" <file>` at the §86-153 / §253-307 / §367-375 regions shows the historical-record markers; **and** `grep -n "not.*current.*settings\|fresh remote evidence" <file>` finds the protection-receipt caution. No marked section retains imperative execution instructions for the bypass recipes.

## AC-ACR-012 — Update-preservation docs (REQ-ACR-012)

**Given** the revised update-survival docs, **When**: `grep -n "보존\|preserv" .moai/docs/update-local-file-survival.md .moai/docs/version-management.md` → the current classification/preservation contract is presented as operative; the historical wipe/restore sections carry incident-record framing (`grep -n "실행 금지\|incident record\|사고 기록"` hits); and the version-management `git-strategy.yaml` note states preservation-and-verify instead of wipe-and-reapply.

## AC-ACR-013 — Primary-main preservation guidance (REQ-ACR-013)

**Given** the revised guidance, **When**: `grep -n "preserv\|보존" AGENTS.local.md .claude/rules/local/*.md` at the recovery clauses shows inventory → preservation → handoff ordering; **and** `grep -rn "git restore\|git reset --hard\|git stash" AGENTS.local.md .claude/rules/local/*.md` at those clauses returns only prohibitions or historical markers, never recovery instructions for the shared primary.

## AC-ACR-014 — Template-First mirror parity + diff list (REQ-ACR-014)

**Given** the completed edits, **When**:
- The changed-file + diff list artifact exists (path recorded in the completion report) with one row per touched document AND its template mirror (mirror rows: `AGENTS.md.tmpl`, `templates/.claude/rules/moai/workflow/factory-dispatch*.md` where touched).
- `make build` → exit 0 (verbatim tail captured).
- C9 card-id guard, aligned with the guard's real semantics (per-(file, literal) removal-only baseline, `internal/template/card_id_leak_test.go:33-56`): (i) `grep -Ec "\bt[0-9]{3,4}\b" internal/template/templates/AGENTS.md.tmpl` → **0 hits** (no baseline entry exists for this file — an accurate zero target); (ii) for the factory-dispatch mirror directory, zero NON-BASELINED hits — every `\bt[0-9]{3,4}\b` occurrence must be a member of the guard's recorded baseline entries (the conforming tree carries baseline-member occurrences there; an unlisted literal, or a listed literal in an unlisted file, fails); (iii) `go test ./internal/template/ -run 'TestCardID|TestTemplateNoInternalContentLeak'` → `ok` at close. Incident references, card-id markers, and in-progress/status markers appear only in the local-only documents.
- For each mirrored pair, `diff <(grep -c <marker> root) <(grep -c <marker> mirror)` shows the same marker present on both sides for each M2/M3 change (or the hunk-level diff is shown equivalent).

## AC-ACR-015 — Kanban prohibition (REQ-ACR-015)

**Given** the revised docs, **When**: `grep -rin "kanban" AGENTS.md AGENTS.local.md .claude/rules/local/*.md .moai/docs/git*.md` → every hit states the prohibition (a Kanban simulation is not implementation approval; no authorization to implement or switch), and **0 hits** authorize Kanban work.

## AC-ACR-016 — CLI vs slash-command forms (REQ-ACR-016)

**Given** the revised documents, **When**: `grep -rn "todo --auto" AGENTS.local.md .claude/rules/local/*.md .moai/docs/git*.md .claude/rules/moai/workflow/factory-dispatch*.md` → every shell-form occurrence is exactly `moai todo --auto`; the slash forms appear only as `/moai todo --auto-lane` and `/moai todo --auto-leader`; **and** `grep -rn "todo --auto-lane" <docs> | grep -v "/moai todo --auto-lane"` → **0 hits** (slash-prefixed occurrences excluded — the invariant stands: `moai todo --auto-lane` never appears in shell form). Where the distinction is drawn, the t1695 delegation note for code-side strings is present.

## Quality gates

- Docs card: the Go coverage/test gates do not apply; the applicable gates are the size model (AC-ACR-009), conflict grep (AC-ACR-005), mirror parity + `make build` (AC-ACR-014), and the wording greps (AC-ACR-002/003/004/007/008/010/011/012/013/015/016).
- Every AC report follows the 5-section evidence format with command + verbatim output + tree SHA.

## Definition of Done

- All 16 ACs PASS with attributed evidence; zero unsuperseded conflict-grep hits; the leader-corrected size model recorded in decision-index (always-loaded total projection ≤ 140,000 with t1887's share; four-chunk move executed) with reconciliation table + loadable companion globs; `make build` green; changed-file + diff list delivered; inputs/ untouched; no capability described as implemented ahead of t1868/t1882/t1895; CLI/slash forms never conflated.
