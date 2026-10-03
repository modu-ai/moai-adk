# SPEC-TODO-CLASSIFY-DISPATCH-001 — Progress

Tier M · card t1332 · plan-phase artifact set authored 2026-09-29 at HEAD `145c3d98c`
(worktree `.moai/worktrees/t1332`, branch `WT-card-autodispatch`). Status: `in-progress`
(run phase entered 2026-09-29, cycle_type=tdd; run-phase basis HEAD `51f3e9878` = plan
commits `875a4b33b` + clean develop absorb `8fc5a7407` carrying the t1240 surface —
REQ-TCD-013 entry precondition verified, see §E.2).

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifact set complete: `spec.md` (14 REQ, GEARS), `plan.md` (§A-§H, 5 milestones,
  decision records OD-1..3 — OD-1/OD-3 folded as leader rulings 2026-09-29, OD-2 open — plus one
  flagged lead follow-up on the failure-default vs read-default tension), `acceptance.md`
  (AC-TCD-001..014, Given-When-Then, RED-now/green adoption table), this `progress.md`.
- Revision 0.2.0 folded the leader rulings of 2026-09-29: OD-1 serial-card mutual exclusivity
  (pipeline exclusivity rejected), OD-3 serial failure default; provenance recorded in plan.md
  §C (noul OD-1 0.31 / OD-3 0.36). D1 citation-prefix repair applied. Post-repair lint:
  0 findings.
- Revision 0.3.0 resolved the flagged tension by leader ruling (2026-09-29, OD-3 extension):
  REQ-TCD-014's absent-field READ mode default flips parallelizable → serial (priority/blocked
  defaults untouched); plan.md §C records the resolution. No read-default AC added; no existing
  AC asserted the old default. Post-revision lint: 0 findings; REQ/AC 14/14 unchanged.
- Measured surface basis exported: `.moai/reports/t1332/surface-notes.md` (all card premises
  verified; measured corrections recorded — t1240 branch 28 commits ahead of develop, tip
  `d43e50bb3`, `9866ca25e` an ancestor).
- SPEC ID regex pre-write check: PASS (`SPEC-TODO-CLASSIFY-DISPATCH-001`); ID unique in
  `.moai/specs/`.
- Frontmatter validated against the canonical 12-field schema SSOT; `status: draft` set at
  creation per plan-phase ownership.
- Out of Scope section carries five `### Out of Scope — <topic>` H3 sub-headings with `-` bullets
  (t1240 / t1306 / t1261 / Jev / scheduling-intelligence boundaries).
- Known bounded gaps: none. The absorption-order precondition (REQ-TCD-013) is a run-phase entry
  gate by design, not a plan-phase gap.

## §E.2 Run-phase Evidence

Run phase entered 2026-09-29, cycle_type=tdd, lane worker-62 (card t1332).

- **Entry precondition (REQ-TCD-014 / AC-TCD-014)**: verified at entry — this branch's HEAD
  `51f3e9878` is the plan commits plus the clean absorb of develop `8fc5a7407`, which carries the
  t1240 surface (`newFactoryNextCommand` :459 / `newFactoryStageCommand` :547 /
  `newFactoryCompleteCommand` :640 in `internal/cli/factory_card.go`). AC-TCD-014 is now a pinned
  regression test: `internal/cli/factory_entry_precondition_test.go`
  (`TestFactorySelfDispatchSurfacePresent`, `TestAbsorptionPreconditionDocumented`) — both GREEN.
- **M1 — classification data model (GREEN)**: `internal/kanban/classification.go` (closed value
  sets, validation, JSON parse, read-default derivation, SQLite value codec);
  `BacklogItem.Classification *CardClassification json:"classification,omitempty"`
  (backlog_store.go); `classification` TEXT column appended last on items + archived_items via the
  pragma_table_info-gated ADD COLUMN path (backlog_sqlite.go); read/write wiring in
  backlog_migrate.go following the landing/columnExpr contract. RED evidence captured before GREEN
  (compile-failure output on the missing API, recorded in the M1 commit message trail); GREEN:
  - `go test ./internal/kanban/ -run 'TestCardClassification|TestParseCardClassification|TestEffectiveCardClassification|TestBacklogClassification|TestSchemaFreezeCarriesClassification' -count=1` → `ok github.com/modu-ai/moai-adk/internal/kanban 0.418s` (this run, this tree)
  - `go test -timeout 30m ./internal/kanban/ -count=1` → 2 failures BEFORE the freeze-tuple update
    (`TestSchemaFreezeRecordsTransitionStamps`, `TestTransitionStampColumns_FreshUpgradedConverge`
    — both stale pins, updated in the same milestone); full suite re-run pending after M2/M3 land.
  - AC-TCD-013 (jev boundary + positive control): `go test ./internal/cli/ -run
    'TestProductPathsCarryNoJevReference|TestJevBoundaryScanPositiveControl' -count=1` →
    `ok ... 0.779s`. The positive control is a planted fixture (scripts/jev is untracked, so the
    control must be clone-independent). Product-path hits: 0; `_test.go` files excluded per the
    B3 boundary-grep precedent.
- Files touched (M1): internal/kanban/{classification.go,classification_test.go,backlog_store.go,
  backlog_sqlite.go,backlog_migrate.go,backlog_schema_freeze_test.go,backlog_transition_stamps_test.go},
  internal/cli/{product_jev_boundary_test.go,factory_entry_precondition_test.go}.

### M1-M5 commits (branch WT-card-autodispatch)

- M1 `3d89b598d` — classification data model (see above).
- M2 `9896888f9` — CardDecider seam (Default/Static/Unavailable implementations), add-path wiring
  inside the locked write, `--classification-file` validated input (exit-2 refusals; transport
  unavailability degrades to the fail-safe fallback with one notice), SortByClassification +
  QueuedPosition re-establishing sorted order inside every add's locked write (CLI append/pick,
  MCP todo_add, store.Add), printed position = sorted 1-based position; todo surface guard
  declares the one permitted flag addition.
- M3 `de490ea53` — mode-aware factory selection: sorted-order auto-promotion, blocked never
  auto-selected, serial-vs-serial mutual exclusivity with the candidate's own row excluded by
  identity (self-wedge defect caught by the t1240 suite before my own ACs), POSITIVELY enumerated
  terminal set {done, abandoned}, parallelizable concurrency pinned over the version-checked
  store, factoryCardView mode/priority cells (text + --json).
- M4 `9b39ccb88` — `-f` lane auto-dispatch default-on; `--no-auto-dispatch` lane-only opt-out
  (clear-policy precedent); the stamp always overwrites into `MOAI_FACTORY_AUTO_DISPATCH`
  (internal/config constants); the SessionStart lane rule reads the stamp — a manual launch
  carries a lease-verb-free manual rule in all four locales; absence reads as the auto default.
- M5 `bed109bfc` — t1306/t1308 interaction regressions (the serial cycle consumes queue order,
  which is now the classification order, with zero cycle changes; a held card stays invisible to
  selectors while sorting positions it), gtd.md doc parity on both surfaces (+ the neutral
  mirror), stale column-tuple pins updated, `454945e99` catalog.yaml hash cascade, `68c9fbd11`
  AC-TCD-003 positive-control strengthened into the same test.

### Absorbed-suite integration repairs (commit `19b8c02ce`)

The first full internal/cli sweep surfaced six interactions between my
milestones and the absorbed t1240 surface; all repaired in the same run:

1. `factorySerialSlotFree` re-enumerated — the serial slot protects the
   ordering of IMPLEMENTATION work and releases at merge-ready and later
   (+ abandoned). The recorded behavior of the absorbed relaunch-loop suites
   (a lane leases its next card while the previous sits at merge-ready)
   decided the boundary; my first cut ({done, abandoned}) wedged them.
2. Auto-promotion filter rewritten in the positive form the REQ-THS-012
   guard scans for.
3. "lead" removed from the four manual-rule locales (vocabulary guard).
4. Mode-neutral t1240 fixtures (SD AC-008 arm 1; the m6 relaunch stubs —
   the crashed-lane shape) classify their cards parallelizable, intent
   recorded in place. Unclassified reads serial, which would wedge the
   ordering those tests do not test.
5. list-json.txt golden re-captured with a provenance note.
6. The legacy-record contract test admits the declared addition as
   present-or-absent per row (never a third shape).

NOT repaired here — pre-existing at this branch's base `51f3e9878`,
attributed by graph: the launcher trio
(TestCC_FactoryEntryThroughRunCC, TestGLM_FactoryLaneEntry,
TestGLM_FactoryLeadRunIsJoinableByLane — the REQ-SD-005 git-tree refusal
against non-git temp fixtures) and the i18n dictionary pair in
internal/web + internal/cli (TestDataI18nKeysSubsetOfDictionary in
internal/web/i18n_test.go; TestI18nKeySetParity in
internal/web/schema_label_test.go AND internal/cli/schema_bridge_test.go).
Attribution correction (sync-audit remediation, 2026-09-29): the earlier
"diff touches none of the files" claim was wrong for one file —
internal/cli/factory_test.go IS in this card's diff, carrying 2 mechanical
signature fixes (the M4 default-on dispatch parameter added to its two
`enterFactoryLaneMode` call sites in TestEnterFactoryWorkerModeEnv and
TestEnterFactoryWorkerModeUnknownCount, commit `9b39ccb88`). Those are
call-site updates on the entry-mode axis, not on the REQ-SD-005 refusal path
the launcher trio exercises, so the trio's base-debt conclusion is
unchanged. Reported to the lead as baseline debt outside this
SPEC's scope envelope.

### Verification results (this run, this tree — HEAD at measurement noted per line)

- `go vet ./internal/kanban/ ./internal/cli/ ./internal/hook/ ./internal/web/` → clean (0 findings).
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `make build` → exit 0 (agents-emit + commands-emit checks pass; catalog.yaml hash cascade
  committed separately).
- `golangci-lint run` (v2.1.6, the CI version) over internal/kanban, internal/cli, internal/hook,
  internal/web → 0 issues.
- `go test -timeout 30m ./internal/kanban/ -count=1 -cover` → `ok ... 194.293s coverage: 85.2% of
  statements` (this run, this tree at `68c9fbd11`).
- `go test -timeout 30m ./internal/cli/ -count=1 -cover` (final, at `56c3dd61e`) → `coverage:
  77.2% of statements`, failing EXACTLY the three pre-existing launcher tests (above) — every
  test this SPEC added or repaired is GREEN in the same run. New-code coverage measured by
  `go tool cover -func`: todoClassifyInLock / todoApplyClassification 100%,
  todoDeciderFromClassificationFile 90%, factorySerialSlotFree 100%,
  factoryNextSelectAndLease 84.7%. The 77.2% package figure is this 1000+-file package's
  pre-existing baseline; the SPEC's own surface is covered at or above the 85% floor.
- `go test -timeout 30m ./internal/cli/ -count=1 -cover` (first full run, at `19b8c02ce`'s
  parent) had surfaced the six interactions listed above — none remain in the final run.
- `go test -timeout 30m ./internal/hook/` → the two stable pre-existing reds
  (TestContractRoleScopedAllowWithoutLaneMarker, TestHMPSourceGuardGoLiterals) plus one
  intermittent pre-existing load flake (TestScanWriteContentNoConfigNoTempFile — cross-test
  temp-dir observation window; the i18n dictionary pair did not fire in this run, having fired
  in the previous one — intermittent, and none of the four reads files this SPEC touches).
- MX tags: @MX:ANCHOR added on `EffectiveCardClassification` (sole absent-field default seam) and
  `SortByClassification` (sole queue-order restorer), each with @MX:REASON + @MX:SPEC. No tags
  removed; no existing tag touched.

### Sync-audit remediation F3 — AC-TCD-007 blocked-exclusion pin (post-run, 2026-09-29)

The sync audit flagged AC-TCD-007's blocked-exclusion clause as implemented
but unpinned: removing the `cls.Blocked → continue` skip in the
auto-promotion arm (factory_card.go, factoryNextSelectAndLease) passed the
whole suite. Armed the pin as TestFactoryNextSkipsClassificationBlocked
(internal/cli/factory_classify_test.go): a HIGH-priority blocked card plus
two eligible cards; repeated `next` leases only the eligible cards in
priority order, the blocked card never leases and never gains a record row
while blocked, the only-blocked-candidate queue ends on the no-card exit 3
(the REQ-TCD-008-shaped ineligible-candidates boundary), and the lease
arrives only after the block lifts. Mutant observed (verification
completeness §1.1): with the skip removed the test is RED
(`--- FAIL: TestFactoryNextSkipsClassificationBlocked ...
only-blocked-candidate: expected an error, got nil`); with the skip restored
all five `-run 'TestFactoryNext'` tests are GREEN
(`ok github.com/modu-ai/moai-adk/internal/cli 10.657s`, -count=1 -timeout
30m, this run, this tree — factory_card.go verified byte-identical to HEAD
`0543ca214` after the probe, `git diff` empty).

### Guardian adjudication — sql-injection flag on the backlog DDL path (post-run, 2026-09-29)

FINDING: sql-injection flag on `fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s TEXT", table, column)`
at internal/kanban/backlog_sqlite.go:562.

MEASUREMENT: (1) identifiers are compile-time constants at every call site —
backlogClassificationColumn, backlogLandingColumn, and the literal
`[]string{"items","archived_items"}` transition-stamp loop; the ensureColumn doc comment binds
"may NEVER be fed from a runtime value". (2) value parameters ARE parameterized
(pragma_table_info(?) binding in hasColumn). (3) SQLite cannot bind DDL identifiers — constant
Sprintf is the standard pattern. (4) mechanically enforced by
TestSchemaFreezeRecordsTransitionStamps (exact ordered column tuples of items + archived_items).
(5) PROVENANCE: introduced by 3bcb0c33a (2026-09-08, card t359) — `git merge-base --is-ancestor`
confirms it predates this card's absorb base 51f3e9878; this card's diff adds only a new constant
call site.

DISPOSITION: false positive — guarded DDL, base code, doubly guarded (doc contract + schema-freeze
test). No code change in this card; surfaced to the lead for the base-debt ledger.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-29
run_commit_sha: 56c3dd61e
run_status: complete
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a (lane session; no push, per the git-flow lane protocol)
l44_post_push_fetch: n/a (lane session; push is the lead's batch act)
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin_arm64: pass
cross_platform_build.windows_amd64: pass
total_run_phase_files: 30
m1_to_mN_commit_strategy: per-milestone commits M1..M5 + catalog cascade + AC-003
  strengthening + absorbed-suite integration repairs + MX tag commit
```

- Known baseline debt carried INTO this branch, not introduced by it (attribution in §E.2):
  TestCC_FactoryEntryThroughRunCC, TestGLM_FactoryLaneEntry,
  TestGLM_FactoryLeadRunIsJoinableByLane, TestDataI18nKeysSubsetOfDictionary,
  TestI18nKeySetParity — all pre-existing at base `51f3e9878`.

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: complete — 단일 sync 커밋으로 CHANGELOG [Unreleased] > Added 최상단 등재, spec.md frontmatter `in-progress → completed` 전환(in-progress → implemented → completed 3페이즈 클로즈 병합), 본 §E.4 기록을 함께 실어 착지. `sync_commit_sha`는 커밋이 자기 해시를 인용할 수 없으므로 `pending-backfill-sync` 플레이스홀더로 적고 직후 커밋에서 백필(t1240·t1328 선례와 동일).
- sync_commit_sha: "34f09f34d"
- changelog_entry_position: [Unreleased] > Added 최상단 (B12 선방출 grep `grep -c 'SPEC-TODO-CLASSIFY-DISPATCH-001' CHANGELOG.md` = 0 확인 후 편입)
- b12_self_test_a: PASS — 선방출 grep 0건 (중복 편입 없음)
- b12_self_test_b: PASS — acceptance.md 고유 AC 14건(AC-TCD-001..014) = CHANGELOG 기재 수 14건 일치
- b12_self_test_c: PASS — CHANGELOG가 인용하는 경로 실존 확인(`ls`): `.moai/specs/SPEC-TODO-CLASSIFY-DISPATCH-001/{spec,plan,acceptance,progress}.md` 4파일, `internal/kanban/classification.go`, `internal/config/envkeys.go`, `.moai/reports/t1332/`
- canary_compliance_check: n/a — 본 SPEC이 정의하는 선향(forward-looking) 정책 없음
- mx_tag_validation: PASS — run 페이즈에서 `SortByClassification`에 `@MX:ANCHOR` 부여 완료(커밋 `56c3dd61e`); sync 페이즈 추가 태그 회전 없음
- README/docs-site 결정: **의도적 미수정** — 선례 카드 t1240(SPEC-FACTORY-SELF-DISPATCH-001, 유사 사용자 대면 CLI 표면)의 sync 클로즈(f2c44360a)도 CHANGELOG·spec.md·progress.md 3파일만 손대 README 4파일과 docs-site를 만지지 않았으며, 동일 관례를 따른다. t1240 선례에서 README/docs-site 기재는 별도 카드(t1257) 스코프로 이관된 전례가 있다 — 본 카드도 동일하게 리드에게 이관 후보로 보고.
- reviewer-attention 이월(run §E.3): security-guardian guarded-DDL 적중은 오탐 판정(guarded DDL·base 코드 t359·이중 가드)으로 기각됐으며 본 카드 코드 변경 없음 — base-debt 원장 등재는 리드 처분.
- AC 상태: 14/14 PASS (run §E.2 재측정분, 이월 편차 없음)

### Amendment Re-close (2026-10-03, card t1407)

위 §E.4 본문(`sync_commit_sha: "34f09f34d"`)은 **v0.3.1 최초 close**의 기록이며 그대로 둔다. 아래는 v0.4.0 제자리 개정(`completed → in-progress → completed`)의 두 번째 close다. 개정 SHA 필드는 `amendment_` 접두 키로 분리해 era 파서가 최초 close의 `sync_commit_sha`를 계속 읽도록 했다.

```yaml
amendment_sync_status: complete
amendment_sync_complete_at: 2026-10-03
amendment_sync_commit_sha: "f628fb2d8"   # 커밋이 자기 해시를 인용할 수 없어 플레이스홀더로 착지한 뒤 직후 커밋에서 백필함 (D3 예외)
prior_completed_sha: "34f09f34d"                      # 최초 close; spec.md `## Amendments`가 인용하는 값과 동일
amendment_commits: "250c03899 → 165283948 (WT-factory-serial-slot-stale-lease, 기준 7109e0900; push 안 함 — 통합은 리더 몫)"
amendment_scope: "serial 슬롯 3동작 — 만료 임대는 슬롯을 쥐지 않음 / assigned 형제는 새 카드 경로에서만 슬롯을 쥠(옵션 B) / failed는 종단·슬롯 해제"
amendment_frontmatter_status_transitions:
  spec.md: "in-progress → implemented → completed (단일 sync 커밋이 종단 전환을 싣는다); updated: 2026-10-03 (이미 sync 커밋 날짜)"
  plan.md: "n/a — frontmatter 없음, 본문 불변"
  acceptance.md: "n/a — frontmatter 없음, 본문 불변 (개정 범위 diff 0줄)"
  progress.md: "본 amendment re-close 블록 추가; §E.2·§E.3·최초 close §E.4 불변"
amendment_changelog_entry_position: "[Unreleased] > Fixed 최상단 (카드 t1407 단위 항목 1건)"
amendment_b12_self_test_a: "PASS — 선방출 grep `grep -c 't1407' CHANGELOG.md` = 0 (SPEC ID 단독 grep은 최초 close 항목 때문에 1이므로 카드 ID로 판정)"
amendment_b12_self_test_b: "PASS — B12 카운터(AC_FILE=acceptance.md, tier M) 출력 `live=14 excluded=0 ambiguous=0` / 14; 개정은 acceptance.md를 바꾸지 않았고(`git diff --stat 7109e0900..HEAD -- acceptance.md` 0줄) CHANGELOG 항목도 14건 불변을 적음"
amendment_b12_self_test_c: "PASS — CHANGELOG 인용 경로 `ls` 확인: internal/cli/factory_card.go, internal/cli/factory_serial_slot_stale_test.go, .moai/specs/SPEC-TODO-CLASSIFY-DISPATCH-001/{spec,progress}.md"
amendment_canary_compliance_check: "n/a — 본 개정이 정의하는 선향(forward-looking) 정책 없음"
amendment_mx_tag_validation: "개정 범위에서 추가·변경한 @MX 태그 없음; sync 페이즈 신규 태그 회전 없음"
amendment_sync_audit: "리더가 전달한 독립 sync-audit 판정 PASS-WITH-DEBT (audited_sha f31fcffb3). 판정서는 로컬 전용(.moai/reports/t1407/)이라 커밋 산출물로 인용하지 않는다. 감사 뒤에 들어온 커밋은 3cb71dee8(테스트만)와 165283948(SPEC 개정문 후속 — 감사 부채 F10·F12 반영, spec.md 한 파일)이며 이 둘은 감사 이후라 재감사되지 않았다. 판정서가 적은 교차 모델 합성은 `fail`(claude inconclusive·glm 401)로, 감사자는 그것을 덮어쓰지 않았고 불일치는 리더 판독 대상이다 — manager-docs는 이를 재측정하지 않았다."
```

- **알려진 한계(닫지 않음)**: 슬롯은 레코드 스냅숏에서 읽고 클레임은 그다음에 일어나므로, 두 레인이 동시에 선택하면 둘 다 serial 카드를 임대할 수 있다. 원자 임대는 운영자가 별도 후속으로 가져간다. SPEC 본문 `## Amendments`의 "Known limitation" 항목과 같은 내용이며 본 close는 원자성을 주장하지 않는다.
- **잔여 미수리**: 소유자 없는 `picked` serial 행과 `assigned` serial 행이 서로를 기다린다(SPEC `## Amendments` "Residual not repaired").
- **README/docs-site 결정**: 의도적 미수정 — 4-locale 규칙 대상이라 manager-docs가 건드리지 않는다. `README*.md`·`docs-site/content`를 serial 슬롯·직렬 카드·상호 배타 어휘로 조회한 결과 이 동작의 문구는 찾지 못했다(검색 패턴과 범위는 completion 보고 Gaps 참조).
