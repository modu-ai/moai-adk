# t97 review verdict — D1 review-role retirement + -k entry unification

- Reviewer: lead session (operator-confirmed direct verdict; t96 lead-verified precedent)
- Card: t97 · Worktree: `.claude/worktrees/t97` · Branch: `WT-t97` @ `c5816c927` (base `5c3141372`)
- Delta reviewed: `5c3141372..c5816c927` (18 files, +178/−86)
- Lens: `--deep` (multi-file logic + runtime guidance change; no auth/secret surface)
- Evidence read: `.moai/reports/t97/run-evidence.md` (53 lines, 5-section)

## Verdict: PASS

## 1. Claims reviewed (evidence vs this review's direct reads)

| # | Claim | Check performed | Result |
|---|-------|-----------------|--------|
| 1 | Commit c5816c927 on WT-t97, base 5c3141372 | `git log` + `git show-ref refs/heads/WT-t97` = `c5816c927aeca…` ✓ | PASS |
| 2 | 18 files +178/−86 | `git diff --stat` — exact match | PASS |
| 3 | CompanionRoles 4→3, review labels no longer companion-parsed | `bootstrap.go`: `var CompanionRoles = []string{"plan", "run", "sync"}` + D1 rationale comment; retired-label test cases (`bootstrap_test.go:97,142`) and fail-open notice test (`session_start_kanban_test.go:322` asserts `kanbanCompanionNotice("review") == ""`) present | PASS |
| 4 | glmSubstitute rework + nameChoices, 4 locales | i18n diff read in full — entry-point unification ("launcher=backend, -k=role") + nameChoices (default/judge/worker-N) filled at en/ko/ja/zh | PASS |
| 5 | 3-phase prose cleanup | `kanban.go` header, `factory.go` genealogy + `rejectConflictingModes` text ("three-role"), `cc.go:45` + `:86` — all read directly; no "plan->run->verify->sync" survives on touched surfaces | PASS |
| 6 | No stray review-role remnants in code | `git grep '"review"'` @c5816c927 — remaining hits are: the retired-label tests (intended), board column test data, and `column.go:22 ColumnReview` (the declared out-of-scope board surface); `RoleReview` constant: 0 hits | PASS |
| 7 | TestHomeJoinSiteCountIsPinned is pre-existing at base | Reviewer-verified disjointness: the 5 implicated files (∉ t97's 18-file delta) — grep empty | PASS |

## 2. Evidence (this review's commands)

- `git log / show-ref / diff --stat / --name-only 5c3141372..c5816c927`
- Full diff read: `bootstrap.go`, `role.go`, `record.go`, `session_start_kanban_i18n.go`; targeted reads: `cc.go`, `kanban.go`, `factory.go`
- `git grep '"review"' / 'RoleReview'` @c5816c927 over internal/kanban + internal/hook + internal/
- Read: `.moai/reports/t97/run-evidence.md`

## 3. Baseline attribution

- Code reads: tree @ `c5816c927` (WT-t97 ref verified). Disjointness grep at the same commit.
- Test outputs (build/vet OK, `internal/kanban ok 15.2s`, `-run "Kanban|ACFB" ok 1.275s`, lint 0, hook suite failing ONLY on the pre-existing base defect): lane-attributed, not re-run — same rationale as the t85/t94 verdicts (lane-local load discipline; CI owns the full matrix on the integrated head).

## 4. Gaps

- Test/lint outputs lane-attributed (see §3).
- `judge` name has no launcher/hook recognition path yet — guidance text only (declared by the lane; a follow-up card wires it if adopted).
- Board `ColumnReview`, web `ChainRoles`, and `kanban-dispatch.md` review-column prose remain — declared scope boundary; a follow-up card owns the board-column retirement.

## 5. Residual risks

- **In-flight review-named sessions**: after this lands, a session labeled `review-*` gets an empty notice and no companion perks (fail-open by design). Exactly the situation this round encountered (no review lane existed); the lead-verified pattern is the sanctioned replacement.
- **Integration collision (material)**: WT-t97 branches from `5c3141372` — WITHOUT t85/t94 — and edits `cc.go`/`glm.go`/`factory.go`/`kanban.go` + kanban i18n. The operator-directed rework on WT-t85 (new base `9794f293d`, which includes t85+t94+t96) is retiring `-f` in favor of `-k N` on the SAME surfaces with different semantics. Both deltas PASS independently; the release integration must resolve these deliberately — flag semantics should follow the rework's `-k N` shape, while t97's D1 three-role direction is common to both.
- 4-phase prose may survive in surfaces outside t97's touched set (docs-site, rules) — the declared follow-up (docs bundle) covers those.
