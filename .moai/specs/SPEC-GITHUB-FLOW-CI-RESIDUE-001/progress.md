# SPEC-GITHUB-FLOW-CI-RESIDUE-001 — Progress

- SPEC: SPEC-GITHUB-FLOW-CI-RESIDUE-001 (card t1535)
- phase 진행: plan 완료(2026-10-07) → run 완료(2026-10-07, §E.2) → sync 완료(2026-10-07, §E.4)
- worktree: `.moai/worktrees/t1535` · branch `WT-github-flow-ci-residue`

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-07
tier: L
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md, progress.md
req_count: 17
ac_count: 17
count_note: iter3 분할 제안서의 "15건"은 산술 오판 — 19 − 2(스코프 4) = 17이고 grep 계수 3곳(spec REQ 불릿·acceptance AC 섹션·§D 추적표)이 17로 일치한다(v0.5.0 재편 시점 실측).
split_history: v0.5.0 — 스코프 4(분할 전 REQ-GFC-013·014(스코프 4) + design D-3 + AC-GFC-013·014 + decision-index Q3)를 iter3 D16(보호 설정과의 구조적 불양립)으로 분할, decision-index Q3 hold(운영자 3택 대기). 잔여 REQ·AC는 015-019 → 013-017 번호 재배열만, 내용 무수정.
measured_head: 5a9d34fbb
measured_branch: WT-github-flow-ci-residue
```

## §E.2 Run-phase Evidence

기록자: lane-8 (manager-develop 위임 대상이 429 4회 연속 종료 → 워치독 §3.5 복구 경로로 레인이 계수·측정. 위임 커밋에 Authored-By-Agent: manager-develop 트레일러, 레인 수리 커밋은 본문에 명기).

**커밋 궤적 (plan 베이스 04e8a6ae6 위):**
- `9404bd8b2` M1 — sweep·done 착지 기준을 LandedRefFor 사슬로 통일 (REQ-GFC-001·002·004·005)
- `7578a54bb` M2 — pr-multi-os-gate 신설 + spec-lint develop 잔여 제거 (REQ-GFC-006~009·011)
- `8b098c508` revert — REQ-GFC-016 patch-id 수리분을 t1561로 이관(리더 판정) — **net diff: landing_predicate.go = 빈 값(04e8a6ae6 대비, 실측)**
- `462080fc5` 게이트 수리 — #30(cancelled 성공 오판)·#31(paths-filter 권한)
- `d0800bd93` M3 — develop push 트리거 9건 제거 (REQ-GFC-013·014·015)
- `1845ce122` 레인 수리 — probe 픽스처 origin ref + #32 NUL 생존 경로 판독

**레인 실측 (이 실행·이 트리 HEAD 1845ce122):**

| 검증 | 명령 | 결과 |
|---|---|---|
| vet | `go vet ./internal/cli/... ./internal/factory/...` | exit 0 |
| worktree+factory | `go test -count=1 -timeout 10m ./internal/cli/worktree/... ./internal/factory/...` | `ok internal/cli/worktree 230.849s` · `ok internal/factory 252.230s` |
| probe(REQ-GFC-017) | `go test ./internal/cli -run TestProductionLaneFilesProbe` | 2/2 PASS (수리 전 ConfiguredBaseWins FAIL — origin ref 미설치 픽스처 결함, 수리 후 GREEN) |
| lint | `golangci-lint run internal/cli/...` | `0 issues` |
| #32 하네스 | /tmp 3케이스(bad object fail-closed · 비ASCII 경로 매칭 · docs-only 대조) | 3/3 PASS |

**Gaps(레인 좌석에서 미측정):**
- REQ-GFC-016(patch-id 공백) — **t1561 위임**(리더 판정), 본 카드 AC 매트릭스에서 "delegated to t1561"
- internal/cli 실바이너리 스폰 테스트(TestTodoAddRefusesExactDuplicate 등)는 전축 MOAI_* scrub 후에도 레인 감지 지속(t1350·t1356 클래스의 레인 env 거짓 적색) — CI 판정 대상
- YML 검증은 수동 하네스로 수행(actionlint 미설치 좌석)

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: complete
run_complete_at: 2026-10-07
milestones: M1 done, M2 done, M3 done
delegated: REQ-GFC-016 patch-id whitespace repair → card t1561 (leader ruling 2026-10-07)
out_of_scope: 스코프 4 전체(분할 → t1557, 운영자 결정 (a) 라벨+수동 병합)
run_evidence_recorded_by: lane-8 (delegate 429-terminated, recovery path)
measured_head: 1845ce122
measured_branch: WT-github-flow-ci-residue
```

## §E.4 Sync-phase Audit-Ready Signal

- sync_complete_at: 2026-10-07
- sync_status: audit-ready
- sync_commit_sha: "fbe98865b"  # backfilled from the pending-backfill-sync placeholder (D3 exemption: the sync commit cannot cite its own hash; backfill commit follows)
- sync scope (single sync commit): CHANGELOG.md `### Fixed` entry (card t1535,
  B12 duplicate pre-check measured 0) + spec.md frontmatter
  `in-progress → implemented → completed` merged transition (status only —
  `updated:` already carried 2026-10-07; SPEC body untouched; plan.md and
  acceptance.md carry no frontmatter block, so nothing to refresh there) +
  this §E.4 signal
- b12_self_test_a: pre-emission `grep -c 'SPEC-GITHUB-FLOW-CI-RESIDUE-001' CHANGELOG.md` → 0 적중 (duplicate guard PASS)
- b12_self_test_b: canonical AC counter on acceptance.md → `live=17 excluded=0 ambiguous=0`; CHANGELOG entry references 17 AC (16 closed + AC-GFC-016 delegated to t1561) — count match PASS
- b12_self_test_c: every file path named in the entry verified via `git diff --stat 04e8a6ae6..HEAD` (9 workflow files, worktree 3 files, factory 1, cli 4, probe test) + direct reads of sweep.go·done.go·prlink_landedref.go·pr-multi-os-gate.yml and the M2/M3 spec-lint diffs — PASS
- changelog_entry_position: [Unreleased] → `### Fixed`, first entry
- frontmatter_status_transitions.spec_md: in-progress → implemented → completed (merged close, single sync commit)
- MX validation (sync sub-step): LandedRefForWithLevel의 기존 `@MX:ANCHOR`가 run 단계에서 갱신된 상태로 검증됨(사유: worktree 표면 편입으로 fan_in 3, 파일 본문 직접 판독); 신규 위험 패턴(goroutine·복잡도) 없음, 제거할 `@MX:TODO` 없음
- no-op dispositions: codemaps rotation — 패키지 구조 변화 없음(기존 파일 함수 내부 수정 + 테스트 2건); README/docs-site sync — 사용자 대면 제품 문서 표면 아님(CI 워크플로 + 내부 CLI 동작)
- delegated: AC-GFC-016 patch-id whitespace repair → card t1561 (leader ruling 2026-10-07, revert commit `8b098c508`)
- out_of_scope: 스코프 4(SPEC 상태 자동 동기화 PR 전환) — v0.5.0 분할 → 카드 t1557(운영자 결정 (a) 라벨+수동 병합); 본 sync가 shipped로 기술하지 않는 범위
- open_blockers: none
