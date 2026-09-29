# Progress — SPEC-REPORTS-LIFECYCLE-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
status: draft
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
basis_tree: "6879cfa5e (develop-based card worktree .moai/worktrees/t1320)"
spec_id_check: "PASS (bash regex, this run)"
req_count: 9
ac_count: 12
open_decisions: []
needs_clarification: 0
audit: "iter1 FAIL 0.88 (MP-7 — 미해결 요구 확인 마커 2건) -> iter2 delta pass"
evidence_dir: ".moai/reports/t1320/ (gitignored, never committed)"
red_baseline: ".moai/reports/t1320/red-baseline-delta.txt + red-blobs-before.txt (511, digest 113a9ea0)"
```

### 결정 채택 기록 (iter2, 레인 채택)

카드 t1320 은 운영자 게이트를 명시하지 않으므로 kickoff-autonomy 정책(AGENTS.local.md §31)에 따라 레인이 결정 포인트를 채택했다. **D1 = tracked 연속성(git mv, 재포함 없음)** · **D3 = skill 기본 경로 무조건 `.moai/reports/`** · **D4 = 연령 기반 90일 mtime + `<YYYY-MM>` 분할 + 플래그/defaults.go(신설 설정 파일 없음)** — 세 건은 plan-audit iter1 이 측정한 전제 위에서 작성자 권장안을 그대로 채택한 것이며, D4 의 세부 3값(창 길이·충돌 정책 skip-and-report·1GB 경고)은 열린 마커가 아니라 보수적으로 채택된 문서화 기본값이다. **D2 = hoist 독립 동사 재설계** — iter1 structural 발견(done 이 L1 트리를 `:80`·`:284` 에서 `--force` 로도 거부, SPEC-WORKTREE-DONE-TIER-001)을 본 러닝에서 grep 으로 재검증해 채택; 선정 사유·기각 기록은 plan.md §F.0 D2.

iter2 delta 수리 내역: AC-RLC-002 판정 명령 교체(16진 SHA 수치 강제 폐기 → 정렬 blob 목록 + staged rename 검사) · REQ-RLC-005/AC-RLC-007/012 재설계(동사 경유) · REQ-RLC-007 술어+보호 집합 정의 · Block 등급 RED 셀 3건 본 러닝 전사로 충전. 미해결 요구 확인 마커 잔존 0건.

카드 본문 대비 정정 5건은 spec.md §A.1·§A.3 에 근거와 함께 기록됐다. 다음 단계: plan-audit iter2 재판정 → Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

증거 디렉터리: `.moai/reports/t1320/` (gitignored). 기선 재측정(M1 착수 직전): `git ls-files reports/ | wc -l` → `511`, `git ls-files -s reports/ | awk '{print $2}' | sort | shasum` → `113a9ea0037d4df5141ebb4b3c90ef8b6f7159d9  -` — plan-phase 핀과 동일.

### M1 — 루트 reports/ → .moai/reports/historical/ (전수 전사: `.moai/reports/t1320/m1-migration.txt`)

| AC | 판정 명령 | 관측 출력 |
|---|---|---|
| AC-RLC-001 | `git ls-files reports/ \| wc -l` | `0` |
| AC-RLC-001 | `git ls-files .moai/reports/historical/ \| wc -l` | `511` |
| AC-RLC-001 | 양변 repo-relative 정규화 `diff` | `0` |
| AC-RLC-002 | 이동 전 blob 정렬 digest (`git ls-tree -r HEAD -- reports/ \| awk '{print $3}' \| sort \| shasum`) | `113a9ea0037d4df5141ebb4b3c90ef8b6f7159d9  -` |
| AC-RLC-002 | 이동 후 (`git ls-files -s .moai/reports/historical/ …`) | `113a9ea0037d4df5141ebb4b3c90ef8b6f7159d9  -` (불변) |
| AC-RLC-002 | `git diff --cached --summary \| grep -cv '^ rename'` | `0` |
| AC-RLC-002 | `git diff --cached --summary \| grep '^ rename' \| grep -cv '(100%)'` | `0` |
| AC-RLC-002 | `git diff --cached --stat=350 \| grep -cv '=>'` | `1` (요약행 1건 — rename 행 전체가 화살표 보유. 평범 `--stat`은 이 git이 장경로 축약 시 화살표를 생략하므로 `--stat=350` 이 축약 없는 동의 형태) |
| AC-RLC-003 | `git status --porcelain \| grep -v '^R'` | `?? internal/cli/worktree/evidence_ignore_guard_test.go` 만 존재 — `D` 0건 |
| AC-RLC-004 | `go test -run TestEvidenceIgnoreMatrixGuard ./internal/cli/worktree/` | PASS (매트릭스 6케이스 + 변이 3건 관측 적색 + 양성대조) |

AC-RLC-004 판정 보강 — 가드의 RED 관측 기록(구현 과정에서 실측): ① 최초 변이 셀 설계(방어행 단독 삭제)에서는 매트릭스가 뒤집히지 않았다(`.moai/reports/*`가 전 케이스를 이미 커버) — 방어행의 존재 이유(spec §A.2)대로 **재포함 규칙 삽입 하에서의 방어행 부재** 변이로 재설계하자 뒤집힘을 관측. ② 이 git의 `check-ignore -v`는 결정 규칙이 negation이어도 exit 0을 낸다(실측: negation 매치 출력 `.gitignore:399:!.moai/reports/**/*.md` + exit 0, `-v` 없이는 exit 1) — 판정축을 exit code에서 `--non-matching` 출력 해석으로 교체. 가드의 실패 관측성 자체가 이 과정에서 두 번 검증됐다.

### M2 — html-report skill 기본 경로 (전수 전사)

| AC | 판정 명령 | 관측 출력 |
|---|---|---|
| AC-RLC-005 | `grep -c '<cwd>/reports/' .claude/skills/moai-domain-html-report/SKILL.md internal/template/templates/.claude/skills/moai-domain-html-report/SKILL.md` | `…:0` / `…:0` (exit 1 — 매치 없음) |
| AC-RLC-005 | 양 미러 바이트 동일성 `diff` | 무출력 (BYTE-IDENTICAL) |
| AC-RLC-005 | `make build` | emit-check 통과 + `catalog.yaml updated successfully (13400 bytes)` + 바이너리 빌드 성공 |
| AC-RLC-005 | `go test -run TestHtmlReportOutputPathParity ./internal/template/` | PASS (양 미러 관례 고정 + 변이 셀: 구 기본값 복원 변이가 실패 집합을 내는 것 관측) |
| AC-RLC-006 | 임시 프로젝트 렌더 관측 (`/tmp/rlc006-render-observe.sh`, 전사: `.moai/reports/t1320/m2-render-observation.txt`) | pre-condition `.moai exists? no` → `mkdir -p` 후 `.moai/reports/status-report-20260929.html`·`.md` 생성 확인 |

AC-RLC-006 성격 주기: html-report skill 은 AI 실행 지침이라 Go 테스트 불가점이며, 본 관측은 편집된 SKILL.md 의 절차(기본 경로 해석 → 디렉터리 자동 생성 → html+md 트윈 출력)를 임시 프로젝트에서 그대로 실행해 경로 관례와 자동 생성 동작을 확인한 것이다 — 렌더 품질(spec §E 예외: 렌더링 동작 전부 현행 유지)은 대상 밖이다.

### M3 — hoist 동사 + done(L2) 배선 + 폐기 플로우 문서 (전수 전사)

RED 증거(GREEN 이전 실측, `.moai/reports/t1320/m3-red-evidence.txt` 전사): `go test -run 'TestHoist|TestDoneL2RemovalHoists' ./internal/cli/worktree/` → `undefined: hoistWorktreeReports / newHoistCmd / runDoneWorktreeCleanupWithOptions` + `[build failed]`.

| AC | 판정 명령 | 관측 출력 |
|---|---|---|
| AC-RLC-007 | `go test -run 'TestHoist\|TestDoneL2RemovalHoists' ./internal/cli/worktree/` (구현 후) | PASS — 동사 정상 인출(`Hoisted 1 file(s), 14 byte(s)` + 목적지 파일 관측) · no-op · 루트 밖 거부 · done L2 제거 전 루틴 호출(실제 git repo+worktree 픽스처) · `--no-hoist` 제외 |
| AC-RLC-008 | 동일 스위트 `TestHoistVerb_ConflictSkipAndReport` | PASS — 목적지 상이 내용 보존(`OLD evidence` 불변) + `Skipped (differs at destination): evidence.md` 보고 + 비충돌 파일 인출 |
| AC-RLC-007(보강) | `go test -count=1 -timeout 30m ./internal/cli/worktree/` | `ok … 28.865s` — 기존 9동사 레지스트리 가드 2건은 신규 동사 등록의 필연 수반 갱신(9→10, Long 텍스트 hoist 명시) 후 전수 GREEN |
| AC-RLC-012 | `grep -c hoist .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | `…:1` / `…:1` — 양 사본 동일 조항(「hoist, then remove」 의무), 조항은 AC-RLC-007 이 검증한 `moai worktree hoist <tree-path>` 동사를 참조 |

구현 형상(REQ-RLC-005 재설계 반영): hoist 코어는 `hoistWorktreeReports(treePath, mainRoot)` 단일 함수 — ① 동사 `moai worktree hoist <tree-path>`(`internal/cli/worktree/hoist.go`, main 루트는 대상 트리 경로에서 역산 — `gitMainRootFromTargetFunc` 재사용) ② done 의 L2 제거 경로 2곳 모두 제거 전 호출(`runDoneWorktreeCleanupWithOptions`, 실패 시 제거 차단, `--no-hoist` 배제) ③ L1 세션 종료 폐기 플로우는 문서 의무로 동사 호출 명시. 충돌 정책은 REQ-RLC-006 대로 skip-and-report.

판정 과정 실측 정정 2건: ① 이 git의 `check-ignore -v` negation exit 0 동작(M1 기록 재확인)과 무관하게, done 배선 테스트의 초기 실패 2건은 테스트 픽스처 설계 오류(tree 가 main temp 밖 — 거부가 정답 / 선행 서브테스트의 목적지 재사용)로 구현과 무관. ② AC-RLC-012 조항은 사본 간 바이트 동일 유지(cp 동기화, 편집 전 IDENTICAL 확인).

### M4 — moai clean reports-archive 액션 (전수 전사)

RED 증거(GREEN 이전 실측, `.moai/reports/t1320/m4-red-evidence.txt` 전사): `go test -run 'TestCleanReportsArchive' ./internal/cli/` → `undefined: runCleanReportsArchiveWithRoot / reportsArchiveWarnNeeded` + `[build failed]`.

| AC | 판정 명령 | 관측 출력 |
|---|---|---|
| AC-RLC-009 | `go test -run TestCleanReportsArchive_MovesOnlyEligibleCandidates ./internal/cli/` | PASS — 술어 적합 후보가 `archive/<YYYY-MM>/` 로 move(원본 부재 + shard 내용 관측), 신선한 항목·이름 비적합 항목 무영향, 출력 건수/바이트 |
| AC-RLC-009(dry-run) | `TestCleanReportsArchive_DryRunMovesNothing` | PASS — 무 force 시 이동·shard 생성 없음 + 후보 건수 보고 |
| AC-RLC-010 | `TestCleanReportsArchive_ProtectedEntriesUntouched` | PASS — historical/·plan-audit/.gitkeep·worktrees/·archive/ 기존분 전수 제자리 |
| AC-RLC-010 | `TestCleanReportsArchive_TrackedEntriesAutoProtected` | PASS — 이름·나이 적중 t338형도 tracked 조건으로 배제 |
| AC-RLC-010(1GB) | `TestCleanReportsArchive_WarnAboveThreshold` | PASS — 경고 임계 경계 양방향 (`config.DefaultReportsArchiveWarnBytes` = 1GiB) |
| 보조 | `go test -count=1 -timeout 30m ./internal/config/` | `ok … 5.719s` — defaults.go 신설 상수 2건 (`DefaultReportsArchiveRetentionDays=90`, `DefaultReportsArchiveWarnBytes=1<<30`) |
| 불변 | `grep -n 'RemoveAll\|os.Remove' clean_reports_archive.go hoist.go` | 0건 — move 전용 불변의 정적 확인 |

구현 노트: ① dry-run 기본은 `moai clean` 의 기존 스코프 관례(--home/--codex-skills)와의 일관성이며 `--force` 가 move 를 수행한다 — REQ-RLC-007 의 move 의무는 force 경로에서 이행된다. ② AC-RLC-009 픽스처명 `t-old`/`t-new` 은 술어 1조(`^t[0-9]+$`)가 이름 형상을 요구하므로 술어 적합 실현명 `t100`/`t101` 로 구현했다(리터럴 t-old 는 술어에 의해 배제돼 AC 의도「술어 적합 후보」와 모순). ③ shard 는 후보 mtime 의 `YYYY-MM` 이다. ④ tracked 판정은 `git ls-files` — 판독 불가도 tracked 로 보는 보수적 default-deny.

### M5 — 템플릿·문서 반영 및 전체 재검증 (전수 전사)

| 항목 | 판정 명령 | 관측 출력 |
|---|---|---|
| AC-RLC-011 | `git diff 6879cfa5e..HEAD -- internal/template/templates/ \| grep '^+' \| grep -cE 't1320\|SPEC-REPORTS-LIFECYCLE\|2026-09-29\|[0-9a-f]{9,}'` | `0` (미러 변경분 전체 추가행에서 카드 유래물 0건) |
| 빌드 축 | `make build` | catalog.yaml 재생성(워킹 diff 0 — 해시 일치) + 바이너리 빌드 성공 |
| 빌드 축 | `make embed-check` | `Pass 1  Warn 0  Fail 0` |
| 코드 품질 | `golangci-lint run ./internal/cli/... ./internal/config/... ./internal/template/` (v2.1.6 = CI 판) | 초기 4건(hoist.go errcheck — 신규 코드) → `_, _ =` 수리 후 `0 issues.` |
| 정적 분석 | `go vet ./internal/cli/... ./internal/config/...` | 클린 |
| 크로스 플랫폼 | `GOOS=windows GOARCH=amd64 go build ./internal/cli/... ./internal/config/...` | 성공 |
| 재검증 | `go test -count=1 -timeout 30m ./internal/cli/worktree/` | `ok … 35.202s` (lint 수리 후 재측정) |
| 재검증 | `go test -count=1 -timeout 30m ./internal/template/` | `ok … 73.758s` |
| 커버리지 | `go test -cover ./internal/cli/worktree/` | `coverage: 86.5% of statements` (목표 85% 충족) |
| 커버리지 | `go test -cover ./internal/template/` | `coverage: 83.8%` — 패키지 기선(본 SPEC 기여분은 테스트 전용, 신규 프로덕션 스테이트먼트 0) |
| CLI 표면 | `./bin/moai worktree --help`, `./bin/moai clean --help` | `hoist <tree-path>` 동사 + `--reports-archive`/`--reports-archive-days` 플래그 관측 |
| 불변 | reports 관련 `git status --porcelain` 잔용 | 0건 (§D.4 조기 신호 클린) |

process 정정 기록: M3 에서 템플릿 미러(worktree-integration.md) 편집 커밋에 `make build` 를 즉행 실행하지 않았다 — M5 의 `make build` 에서 워킹 diff 0 으로 확인(해시 대상이 아닌 파일이었음)했으나, 절차상 미러 편집 커밋마다 build 를 run 하는 규율 위반이므로 여기에 기록한다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-29
run_commit_sha: "292cf70f8 (M-final)"   # backfilled with the M-final commit SHA in the follow-up commit (D3 exemption)
run_status: complete
ac_pass_count: 12
ac_fail_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "n/a (card worktree, no push — develop push is the leader's batch act)"
l44_post_push_fetch: "n/a (same)"
new_warnings_or_lints_introduced: 0   # 4 errcheck findings on new hoist.go code fixed within the run; final lint 0 issues
cross_platform_build.darwin: "pass (native)"
cross_platform_build.windows: "pass (GOOS=windows GOARCH=amd64 go build, changed packages)"
total_run_phase_files: 14   # 511 renamed + guard test + parity test + hoist.go/test + done.go + root.go/test + 2 SKILL.md + catalog.yaml + 2 worktree-integration.md + clean.go + clean_reports_archive.go/test + defaults.go + spec/progress
m1_to_m5_commit_strategy: "one commit per milestone (M1 9fa86e6bb, M2 476cb929d, M3 4aff3bbc7, M4 ac49df61d, M-final this commit)"
milestones: [M1, M2, M3, M4, M5]
evidence_dir: ".moai/reports/t1320/ (gitignored — m1-migration.txt, m2-render-observation.txt, m3-red-evidence.txt, m4-red-evidence.txt)"
gaps:
  - "internal/cli 패키지 전체 -cover 재측정 미실행(전 스위트 1485.088s 를 cover 플래그로 재실행하지 않음) — worktree 86.5%, template 83.8%(기선) 만 측정. 전 패키지 판정은 CI 몫"
  - "27개 기존 카드 트리의 증거 인출은 spec §E 배제 항목(run phase 산출물 아님)"
sync_should_verify:
  - "AC 전수 판정 명령·출력이 §E.2 에 전사돼 있는지 (§D.5 클로저 게이트 1)"
  - "run_commit_sha 백필 (이 커밋 직후 chore 커밋)"
  - "spec.md:89/acceptance.md:10 의 N1 cosmetic 잔여 (KNOWN sync-phase leftover — 본 러닝 미수리)"
  - "미러 변경분 중립성 재판정 (CI template-neutrality-check 안전망)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-29
sync_commit_sha: "pending-backfill-sync"   # replaced with the real sync commit SHA in the follow-up chore commit (D3 exemption)
sync_status: complete
b12_self_test_a: "pre-emission grep SPEC-REPORTS-LIFECYCLE-001=0, t1320=0 in CHANGELOG.md — cleared for emission"
b12_self_test_b: "AC count match — acceptance.md live identifiers AC-RLC-001..012 = 12 (reserved-token rows excluded by the live-identifier rule); CHANGELOG entry cites 12 AC"
b12_self_test_c: "file path verification — hoist.go, done.go, clean.go, clean_reports_archive.go(+test), worktree/hoist_test.go, defaults.go, both SKILL.md mirrors, 4 README locales: all confirmed via ls/grep"
changelog_entry_position: "CHANGELOG.md [Unreleased] > Added (first entry)"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed (merged into the single sync commit)"
  plan_md: "n/a (this SPEC carries no separate plan.md frontmatter transition; plan.md untouched)"
  acceptance_md: "untouched — body frozen per ownership policy"
  progress_md: "no status field; §E.4 authored in this sync"
canary_compliance_check:
  spec_body_untouched: true   # N1 already repaired in 7aaa5bb03 before sync; no body edits in sync phase
  stage_by_pathspec: true
  no_push_no_merge: true      # lane defers to the leader's batch push + integration window
sync_should_verify_dispositions:
  - item: "AC 전수 판정 명령·출력 §E.2 전사 (§D.5 클로저 게이트 1)"
    result: "confirmed — §E.2 carries command+output transcriptions for all 12 AC"
  - item: "run_commit_sha 백필"
    result: "already landed pre-sync — commit 503ec2e8f backfilled §E.3 with 292cf70f8 (M-final)"
  - item: "N1 cosmetic 잔여 (spec.md:89/acceptance.md:10)"
    result: "already repaired pre-sync — commit 7aaa5bb03 reworded AC-RLC-002 cells to sorted blob-SHA list identity"
  - item: "미러 변경분 중립성 재판정"
    result: "SKILL.md mirrors carry no SPEC IDs, dates, SHAs, or internal paths (verified by read); CI template-neutrality-check remains the safety net"
readme_docs_site_decision: "README 4-locale verb-table rows updated (worktree row gains hoist, clean row gains --reports-archive) — existing rows documenting the new verbs, no invented sections; docs-site deferred (worktree guide/faq + moai-clean pages need authored 4-locale sections + hugo build — follow-up card)"
```
