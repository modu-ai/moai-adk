# Acceptance Criteria — SPEC-STEERING-ALIGN-LOCAL-DIET-001

Tier S. CONSERVATIVE diet of the git-tracked `CLAUDE.local.md`. Every AC is **behavioral** (verify the right content moved/stayed) and **mechanically verifiable** via grep / line-count / git-status against the live on-disk file. No AC is numeric-proxy (no "must remove exactly N lines" pass condition — P5 over-cut lesson).

> **Verification surface note**: all `CLAUDE.local.md` ACs read the LIVE on-disk file. `CLAUDE.local.md` is git-tracked (§F.1 of plan.md), so the live working-tree state equals the committed state once the diet edit lands; `wc -l` / `grep` against the working-tree file is the authoritative evidence.

---

## §A. Given-When-Then Scenarios

### Scenario 1 — §19.1 HUMAN GATE body survives the diet

- **Given** `CLAUDE.local.md` contains the §19.1 구현 착수 승인 (Implementation Kickoff Approval, plan-to-implement HUMAN GATE) `[HARD]` body at L702-718,
- **When** the CONSERVATIVE diet is applied (only the two §2 candidate edits + optional §19.1 minor header compression),
- **Then** the §19.1 `[HARD]` directive line, the 4-step orchestrator obligation list, and the violation anti-pattern are all still present and renderable in the live always-loaded `CLAUDE.local.md`.

### Scenario 2 — Pre-PR checklist duplicated body is pointer-ized

- **Given** the §2 Pre-PR Verification checklist (L108-118) is a 7-item bullet list duplicating the canonical 5-item self-check in `template-internal-isolation-doctrine.md §25.3` (with the §25.3 pointer already at L118),
- **When** the M-POINTER edit (REQ-LD-001) is applied,
- **Then** the duplicated 7-item `- [ ]` bullet list is gone, the §25.3 pointer survives (and names §25.1 as the content-class catalogue), and the §2 section + lead sentence remain.

### Scenario 3 — §2.1 stale cross-ref is corrected

- **Given** §2.1 (L106) cites the non-existent `coding-standards.md § MUST` section and "C1/C2/C4/C5/C6/C8 per coding-standards.md MUST constraints",
- **When** the CORRECTION + M-POINTER edit (REQ-LD-003) is applied,
- **Then** §2.1 no longer cites `coding-standards.md § MUST`, the corrected cross-ref points at `template-internal-isolation-doctrine.md §25.1 / §25` (verified to exist on disk), and the neutrality obligation + `template-neutrality-check.yaml` CI-guard mention are preserved.

---

## §B. Edge Cases

- **EC-1 (candidate SSOT moved/removed between plan and run)**: if the §25.3 or §25.1 anchor is absent at run-time (M1 re-grep returns 0), the run-phase BLOCKS the candidate #1 edit and reclassifies it KEEP (REQ-LD-002) — returns a blocker report rather than pointing at a non-existent target.
- **EC-2 (§2.1 corrected target itself stale)**: if `template-internal-isolation-doctrine.md §25.1` does not exist at run-time, the run-phase BLOCKS the §2.1 rewrite (REQ-LD-004) rather than introducing a new broken cross-ref.
- **EC-3 (only one candidate qualifies)**: if exactly one of the two candidates passes its gate, the run-phase applies that one and reports the other as KEEP-with-rationale; the line-count band (AC-LD-005) accommodates a smaller reduction via the behavioral-PASS escape.
- **EC-4 (no clean §19.1 minor compression)**: if no clean minor-header compression of §19.1 is available, the run-phase leaves §19.1 verbatim (KEEP is the safe default); AC-LD-001 still PASSes (body present).
- **EC-5 (tracked-file commit diff)**: `CLAUDE.local.md` is git-tracked, so the diet edit lands in the run/sync commit diff alongside the SPEC artifacts (standard P2/P5 close, plan.md §F.1). A verifier inspecting `git show <sync_commit_sha>` SHOULD see `CLAUDE.local.md` in the diff. The run-phase MUST `git add CLAUDE.local.md` before commit (AP-LD-006); forgetting to stage the tracked edit would leave the diet uncommitted.

---

## §C. Quality Gate Criteria

- spec-lint: `## D. Out of Scope` carries ≥1 `### Out of Scope — <topic>` h3 sub-heading with `-` bullets (4 present) → `OutOfScopeRule` (MissingExclusions) satisfied.
- Frontmatter: 12 canonical fields + `tier: S` + `era: V3R6`; `created:`/`updated:`/`tags:` canonical (no snake_case alias).
- No Go test impact (no Go code change, no template edit) — `go test ./...` baseline unchanged is NOT an AC of this SPEC (out of scope, C-5); only `git status` scope discipline (AC-LD-007) applies.
- `moai spec audit` (run after binary rebuild if stale): drift 0 expected for this V3R6 SPEC.

---

## §D. AC Binary Matrix (run-phase fills Status + Actual Output)

| AC | GEARS form | Verification command (re-runnable) | PASS condition |
|----|-----------|-------------------------------------|----------------|
| **AC-LD-001** | The diet SHALL NOT remove the §19.1 [HARD] HUMAN GATE body. | `grep -c '\[HARD\].*구현 착수 승인.*plan-to-implement HUMAN GATE' CLAUDE.local.md` ; `grep -c '오케스트레이터 의무 (구현 착수 승인 entry)' CLAUDE.local.md` ; `grep -c '위반 anti-pattern' CLAUDE.local.md` | all three ≥1 (the `[HARD]` directive line, the 4-step obligation header, and the violation anti-pattern survive) |
| **AC-LD-002** | The Pre-PR checklist duplicated 7-item body SHALL be replaced by the one-line §25.3 pointer (ALL 7 bullets removed, not a partial subset). | (pointer survives) `grep -c 'template-internal-isolation-doctrine.md.*§25.3' CLAUDE.local.md` ; (first bullet gone — C1) `grep -c 'No .* OS-specific absolute path (C1)' CLAUDE.local.md` ; (a SECOND distinct bullet gone — C8, guards against partial removal) `grep -c 'GOOS=.*cross-compile env vars preserved (C8)' CLAUDE.local.md` ; (whole-checklist bullet count gone — no `- [ ]` checklist bullets remain in the §2 Pre-PR region) `grep -c '^- \[ \] No \`/Users/\`' CLAUDE.local.md` | pointer ≥1 AND first-bullet (C1) == 0 AND second-bullet (C8) == 0 AND the `- [ ]` checklist bullets removed. (Plan-phase live verification: pointer→1, C1-bullet→1, C8-bullet→1, 7 `- [ ]` bullets present pre-diet; C1-bullet + C8-bullet both flip to 0 post-diet. Two independent bullet assertions + the pointer survival prevent a partial-removal false-PASS — D3 strengthening.) |
| **AC-LD-003** | The §2.1 cross-ref SHALL no longer cite the non-existent coding-standards.md §MUST / C1-C8 constraints. | (broken citation gone — TWO independent greps, BOTH must reach 0 post-diet) `grep -c 'C1/C2/C4/C5/C6/C8 per' CLAUDE.local.md` AND `grep -c '§ MUST' CLAUDE.local.md` ; (corrected target present) `grep -c 'template-internal-isolation-doctrine.md §25.1' CLAUDE.local.md` | both broken-citation greps == 0 AND corrected-target ≥1. (Plan-phase live verification: `C1/C2/C4/C5/C6/C8 per`→1, `§ MUST`→1, corrected-target→0 pre-diet; all flip to 0/0/≥1 post-diet. The two stale citations on L106 are the `C1/C2/C4/C5/C6/C8 per ... coding-standards.md MUST constraints` parenthetical AND the `coding-standards.md § MUST` cross-ref.) |
| **AC-LD-004** | Dev-local-unique sections SHALL be preserved (§A.6 preserve map). | `for h in '## 1. Quick Start' '## 5. Version Management' '## 6. Testing Guidelines' '## 13. GLM Integration Testing' '## 16. 오케스트레이터 자가 점검' '## 20. Vercel Build Cost Guard' '## 22. Dev Settings Intent' '## 23. Local Git Workflows' '## 24. Harness Namespace' '## 25. Template Internal-Content Isolation'; do grep -qF "$h" CLAUDE.local.md && echo "OK $h" || echo "MISSING $h"; done` | every line prints `OK` (zero `MISSING`) |
| **AC-LD-005** | The final line count SHALL land in the derived range (SOFT band, behavioral-PASS escape). | `wc -l < CLAUDE.local.md` | result in `[771, 806]` (soft band — 806 is the unchanged ceiling; a small honest reduction or near-zero reduction is a legitimate behavioral-PASS if only the two candidates qualify per REQ-LD-007; over-cut below ~771 to hit a number is FORBIDDEN). Report the actual count + which candidates applied; NO "must remove ≥N" gate. |
| **AC-LD-006** | The §2.1 correction SHALL NOT introduce a NEW broken cross-ref. | `grep -c '§25.1 정의 — Allowed vs Forbidden' .moai/docs/template-internal-isolation-doctrine.md` (new target exists on disk) | ≥1 (the rewritten §2.1 target anchor is verified present in the SSOT doctrine file) |
| **AC-LD-007** | Scope discipline: ONLY CLAUDE.local.md + SPEC artifacts SHALL change; SSOT doctrine files byte-unchanged. | `git status --porcelain .moai/docs/template-internal-isolation-doctrine.md .claude/rules/moai/development/coding-standards.md` (must be empty) ; `git status --porcelain CLAUDE.local.md` (tracked modification expected) ; `git status --porcelain .moai/specs/SPEC-STEERING-ALIGN-LOCAL-DIET-001/` (SPEC artifacts present) | doctrine-files status EMPTY (byte-unchanged) AND `CLAUDE.local.md` shows a tracked modification (` M CLAUDE.local.md`, since it is git-tracked — the diet edit IS expected in the working-tree/commit diff) AND SPEC-artifact changes present. (`CLAUDE.local.md` is git-tracked — its modification DOES appear in `git status`; this is the standard P2/P5 tracked-file path per plan.md §F.1.) |

### §D.1 Severity / Blocking classification

| AC | Severity | Blocking? | Rationale |
|----|----------|-----------|-----------|
| AC-LD-001 | MUST | **BLOCKING** | The §19.1 HUMAN GATE preservation is the load-bearing user decision (REQ-LD-005 / C-3). Loss of the `[HARD]` body is a hard FAIL. |
| AC-LD-002 | MUST | **BLOCKING** | The Pre-PR pointer-ization is the primary diet deliverable (REQ-LD-001). |
| AC-LD-003 | MUST | **BLOCKING** | The §2.1 stale-cross-ref correction is the second diet deliverable (REQ-LD-003) — a broken cross-ref is a live defect. |
| AC-LD-004 | MUST | **BLOCKING** | Dev-local knowledge preservation is the CONSERVATIVE-bound invariant (REQ-LD-006). Loss of any preserve-map section is over-cut FAIL. |
| AC-LD-005 | SHOULD | non-blocking | SOFT band with behavioral-PASS escape (REQ-LD-007). A small/near-zero honest reduction PASSES; the band exists to catch accidental over-cut, not to force a number. |
| AC-LD-006 | MUST | **BLOCKING** | No NEW broken cross-ref (REQ-LD-004) — the correction must not regress the very class of defect it fixes. |
| AC-LD-007 | MUST | **BLOCKING** | Scope discipline (C-7) — the SSOT doctrine files MUST stay byte-unchanged (M-POINTER points at, never edits, the SSOT). |

6 MUST-BLOCKING ACs + 1 SHOULD (the line-count band, with behavioral-PASS escape per the P5 over-cut lesson).

### §D.2 REQ → AC Traceability

| REQ | Covered by AC |
|-----|---------------|
| REQ-LD-001 (Pre-PR checklist pointer-ization) | AC-LD-002 |
| REQ-LD-002 (pre-edit duplication re-grep gate) | AC-LD-002 (verification command re-run) + EC-1 |
| REQ-LD-003 (§2.1 stale cross-ref correction) | AC-LD-003 |
| REQ-LD-004 (no NEW broken cross-ref) | AC-LD-006 + EC-2 |
| REQ-LD-005 (§19.1 HUMAN GATE body preserved) | AC-LD-001 |
| REQ-LD-006 (dev-local knowledge preserved) | AC-LD-004 |
| REQ-LD-007 (derived range, behavioral-PASS escape) | AC-LD-005 |
| C-7 (scope discipline, SSOT byte-unchanged) | AC-LD-007 |

Every REQ + the scope-discipline constraint is covered by at least one mechanically-verifiable AC. No AC depends on a numeric-proxy line target (AC-LD-005 is a soft band with an explicit behavioral-PASS escape).

---

## §E. Definition of Done

- [ ] AC-LD-001..004, AC-LD-006, AC-LD-007 (6 MUST-BLOCKING) all PASS with live command output recorded in progress.md §E.2.
- [ ] AC-LD-005 (SHOULD) reported with the actual line count + which candidates applied (behavioral-PASS rationale if the reduction is small).
- [ ] §19.1 `[HARD]` HUMAN GATE body verified present (AC-LD-001).
- [ ] Both verified candidates re-grepped live before edit (REQ-LD-002 / REQ-LD-004 gates); any failed gate → candidate reclassified KEEP + blocker reported.
- [ ] SSOT doctrine files (`template-internal-isolation-doctrine.md`, `coding-standards.md`) byte-unchanged (AC-LD-007).
- [ ] SPEC artifacts + progress.md + the `CLAUDE.local.md` diet edit committed together (Conventional Commits + `🗿 MoAI` trailer); `CLAUDE.local.md` is git-tracked and IS in the commit diff (standard P2/P5 close — plan.md §F.1); `sync_commit_sha` points at a commit whose diff contains `CLAUDE.local.md`.
- [ ] progress.md §E.2/§E.3 (run) + §E.4 (sync) populated by manager-develop / manager-docs respectively.
- [ ] `moai spec audit` (post-rebuild) shows drift 0 for this V3R6 SPEC.
