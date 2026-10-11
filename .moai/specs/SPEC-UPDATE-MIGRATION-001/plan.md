# plan.md — SPEC-UPDATE-MIGRATION-001 (card t1547)

## A. Context

- Worktree: `.moai/worktrees/t1547`, branch `WT-update-migration` (develop-based card worktree; commits land here, integration via the lane/leader flow — repo-local PR policy, no card PRs).
- SPEC: `.moai/specs/SPEC-UPDATE-MIGRATION-001/{spec,plan,acceptance,design,research,progress}.md` (Tier L).
- Key code sites (all verified 2026-10-07, research.md §1-6): `internal/cli/update/deploy/deploy.go` (`ManagedCleanTargets` :56, `CleanMoaiManagedPaths` :107, `WithTargets` :119, config wipe :201-215, `backupThenRemove` :437), `internal/cli/update/migrate_classify.go` (`ClassifyMigration`, `classifyCarried`), `internal/cli/update/plan/plan.go` (`IsUserOwnedNamespace` :152, `IsMoaiManaged` :236, count exclusion :73, `DetermineStrategy`), `internal/cli/update/merge/merge.go`, `internal/cli/update/report/outcome.go`.
- Methodology: DDD (brownfield; characterization-test-first per the card) — `cycle_type=ddd`.

## B. Known Issues (relevant subset)

- B2 (cross-SPEC conflict): SPEC-INIT-SHRINK-001 owns the classifier over DROPPED roots — this SPEC extends the same semantics to MANAGED roots; do not regress its foreign-preservation contract. SPEC-CLI-CLEAN-SYMLINK-001's Lstat-first ordering is load-bearing. SPEC-UPDATE-DATA-SURVIVAL-001 REQ-UDS-008 backup ordering is load-bearing.
- B3/B11 (subagent boundary): CLI code carries no AskUserQuestion-shaped code; `--dry-run` and summary output remain non-interactive.
- B6: spec.md carries `### Out of Scope — <topic>` H3 sub-headings (done).
- B8 (tree hygiene): do not touch `.moai/state/`, `.moai/harness/`, other SPEC directories; stage by explicit pathspec.
- B1 (cross-platform): classification walks must use `filepath` + Lstat semantics; verify `GOOS=windows GOARCH=amd64 go build ./...`.

## C. Pre-flight

```bash
git branch --show-current && git rev-parse HEAD
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
go test -count=1 ./internal/cli/update/...   # green baseline BEFORE any change
grep -rn 'Retired\|superseded' internal/cli/update/ | head   # cross-SPEC scan
```

## D. Constraints

- PRESERVE: `internal/merge` engine API, `internal/cli/update/deploy` leaf-ness (no new imports from root cli), symlink dispositions, REQ-UDS-008 backup-first ordering, t111 pre-clean backup, existing test intent (characterization tests extend, not rewrite).
- Forbidden: `git add -A`/`git commit -a`; local full-suite `go test ./...`; touching t1509 (user-folder install) or t1527 (UX) surfaces; template content changes without Template-First (`make build`).
- Test isolation: `t.TempDir()` everywhere; no OTEL env vars in tests.
- Commits: Conventional Commits + card id (t1547) + `🗿 MoAI` trailer; one commit per milestone.

## E. Self-Verification

Per-milestone: scoped `go test -count=1 ./internal/cli/update/...` (target packages only), then at close the §E matrix per manager-develop-prompt-template (E1-E8), coverage ≥85% per touched package, windows build, lint delta (NEW vs baseline), and the two hazard ACs (AC-UPM-020/021) run as fixture tests, not manual steps.

## F. Milestones (priority-ordered; no time estimates)

### M1 — Characterization pin (Priority High) — GREEN-BASE, no behavior change
Characterize current `CleanMoaiManagedPaths` behavior BEFORE touching it. New tests:
- local-only file under each managed root is backed up to pre-clean AND deleted (documents the hazard);
- template-carried but user-modified file is overwritten without merge (documents the overwrite hazard);
- clean-tree update end state (NFR-UPM-003 baseline).
Files: `internal/cli/update/deploy/deploy_characterization_test.go` (new).
Exit: characterization suite green on the UNMODIFIED tree.

### M2 — RED: hazard tests (Priority High)
Write the failing tests the fix must satisfy (RED evidence captured verbatim per E8, and per the acceptance.md measurement protocol — the two release-blocking ACs bind to these exact test names):
- `TestUpdate_LocalOnlyFileSurvives` — local-only file under a managed root SURVIVES update byte-for-byte (AC-UPM-020);
- `TestUpdate_GitStrategyValuesSurvive` — operator-set `git-strategy.yaml` VALUES survive update (`worktree_base_branch: develop` vs the template's empty default; `manual.workflow: git-flow` vs `github-flow`), no manual re-apply (AC-UPM-021);
- conflict path: user-modified template file preserved + first-unused numbered `.moai-new` sidecar (collision case included) + summary row (AC-UPM-030 subset);
- stale file archive-then-remove with summary listing.
Files: `internal/cli/update/reconcile_test.go` (new; table-driven fixtures under `testdata/`).
Exit: RED observed for each (these tests fail against the current wipe-first flow).

### M3 — Classifier extension (Priority High)
Extend the three-way classification to the managed-root set: carriage-gated classes, manifest-aware identical/modified split, `IsUserOwnedNamespace` force-preserve, Lstat-first symlink recording, deterministic output. Reuse `ClassifyMigration`/`classifyCarried` semantics (extract shared helpers rather than duplicating).
Files: `internal/cli/update/reconcile_classify.go` (new) + `reconcile_classify_test.go`; minimal refactor touching `migrate_classify.go` only if a helper extraction demands it.

### M4 — Reconciliation pipeline (Priority High)
Wire classify → reconcile → deploy in the update orchestration: per-class dispositions (refresh/merge/conflict/preserve/archive-remove), base-aware config 3-way merge replacing the `.moai/config` wipe (design.md §4 — operator values win where ours changed from base and the template side did not), REQ-UPM-015 guard (wholesale path refuses `user-owned`/unresolved-modified deletions), `CleanMoaiManagedPaths` call removed from the default existing-project path (retained for legacy fresh-install, REQ-UPM-040).
Files: `internal/cli/update/reconcile.go` (new), `internal/cli/update.go` (wiring), `internal/cli/update/deploy/deploy.go` (guard + target plumbing only).

### M5 — Summary + counting (Priority Medium)
Outcome categories (refreshed/merged/conflicted/preserved/archived-removed) with per-path lists; remove the `plan.go:73` managed-count exclusion; deletions reported (REQ-UPM-031); keep output on existing report structures (t1527 boundary). `--dry-run` previews the reconciliation plan using the read-only classifier.
Files: `internal/cli/update/report/outcome.go`, `report.go`, `internal/cli/update/plan/plan.go`, preview files as needed.

### M6 — Green + docs (Priority Medium)
All M2 tests green; coverage ≥85% on touched packages; windows build; lint delta clean. Update the local-only doctrine doc `.moai/docs/update-local-file-survival.md` (survival rule is now enforced by code — document the new guarantees + what still requires care: `.moai-new` sidecars, legacy path) and AGENTS.local.md §2.3 summary line (local file; update within the card, merge via develop). Files: docs (local-only, not template-mirrored per §25 neutrality), test finalization.

### M7 — Verification close (Priority Medium)
Full §E matrix, scoped re-measurement including `./internal/cli/update/...` + `./internal/manifest/...`, plan-audit evidence under `.moai/reports/t1547/`, progress.md §E.2/§E.3 population handoff to run phase records.

## G. Anti-Patterns

- Do NOT delete `CleanMoaiManagedPaths` (legacy callers depend on it); guard it instead.
- Do NOT classify by managed-name alone (D-15 loss path).
- Do NOT follow symlinks in any new walk.
- Do NOT let a merge engine conflict write to the user's file.
- Do NOT count a green hazard test as proof without the RED-first capture (M2).

## H. Cross-References

- spec.md (requirements REQ-UPM-*), acceptance.md (AC matrix), design.md (pipeline + dispositions), research.md (code evidence).
- Related SPECs: SPEC-INIT-SHRINK-001, SPEC-CLI-CLEAN-SYMLINK-001, SPEC-UPDATE-DATA-SURVIVAL-001, SPEC-CLI-TUX-V3-003. Boundary cards: t1509, t1527.
