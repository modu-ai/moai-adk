# progress.md — SPEC-LEARN-CHANNEL-SCOPE-001

> §E 하위 4 섹션은 era 분류 파서가 인식하는 lifecycle 구조다 (섹션명·순서 고정). §E.5(Mx-phase)는 폐지 — 새 SPEC에 발행하지 않는다.

## §E.1 Plan-phase Audit-Ready Signal

_plan-phase artifacts authored 2026-09-01 (manager-spec, v0.1.0 → v0.1.1 iter-1 repair: D1-D7 — writer 2패밀리 재프레임·패밀리-집합 판정식·스윕 목록 명시·RED-now 셀 실관측 → v0.1.2 iter-2 repair: N1 영어 마커 재키잉·N2 pointer-only 정렬) — plan_status: pending kickoff (iter-2 PASS-WITH-DEBT 0.86, N1 해소 후 기계 대조 예정)._

## §E.2 Run-phase Evidence

_run-phase executed 2026-09-02 (manager-develop, cycle_type=tdd — docs-only SPEC으로 RED-GREEN-REFACTOR 부적용, 측정·grep 관측값이 검증 층; worktree `WT-learn-channel-gap`, 시작 HEAD `d7ce6c6bd`)._

**측정 신선도 (인용값 재사용 금지 원칙)** — run-phase 시작 시 §C 명령 전수 재측정: 인박스 **5,958행 전부 `tool_failure`** (22버킷, `{tool_failure, test_fail}` 집합 밖 **0**), `grep -c 'test_fail:'` → `0`/exit 1, 인간 채널 `feedback_*.md` **165개** (08-25 이후 **146**). plan-phase baseline(5,942/164/145) 대비 append-only 성장, 구성 결론 불변 — 5번째 연속 일치 관측. run 종료 시 재측정 5,958행(외부 append 없던 창), `test_fail` 0 유지.

**AC 매트릭스 (E1 — GREEN 전부 본 라운드 관측, RED 셀은 spec.md §I 인용):**

| AC | 판정 | 명령 | 원문 출력 (본 라운드, 이 트리) |
|---|---|---|---|
| AC-LCS-001 | **PASS** (M1 flip; RED `test -f …` → 출력 없음/exit 1 @ `d7ce6c6bd`) | `test -f .moai/docs/learning-channel-scope.md` | exit 0. tally 재실행: `jq -r '.event_key // "NO_EVENT_KEY"' … \| cut -d: -f1-2 \| sed 's/:$//' \| sort -u` → 22버킷 전부 `tool_failure:*`; 집합 밖 버킷 수 `grep -cv '^tool_failure'` → `0`/exit 1 — 재검증 가능성 성립 (baseline: 2026-09-02, `d7ce6c6bd8dcc5f48a9ab46555f52d14e68540d9`, anchor doc § Dated baseline) |
| AC-LCS-002 | **PASS** (M2 flip; RED `grep -c 'human-mediated loop' …detail.md` → `0`/exit 1 @ `d7ce6c6bd`) | `grep -c 'human-mediated loop' .claude/rules/moai/core/moai-constitution-detail.md` (미러 동일) | 로컬 `1`/exit 0 · 미러 `1`/exit 0 — 양쪽 경계 주장+채널 명명. 미러 neutrality: SPEC id `0`, `t260` `0`, `/Users/` `0`, `2026-` `0`, 행수 토큰 `0`, `624` `0` (전부 exit 1 = 부재) |
| AC-LCS-003 | **PASS** (M2 flip; RED SKILL.md 동 마커 → `0`/exit 1 @ `d7ce6c6bd`) | `grep -c 'human-mediated loop' .claude/skills/hns-lsel-curator/SKILL.md` | 마커 `1`/exit 0; 스테일 카운트 `grep -c '624 stubs'` → `0`/exit 1 (제거됨); 앵커 포인터 `grep -c 'learning-channel-scope'` → `1`/exit 0 — constraint-7 pointer-only 형태 |
| AC-LCS-004 | **PASS** (M3 flip; RED `grep -c 'learning-channel-scope' SKILL.md` → `0`/exit 1 @ `d7ce6c6bd`) | `grep -rln 'lessons-inbox' . --exclude-dir={.git,specs,reports}` → 26파일 → §A.5 배제표 분류 | claim 표면 4종(+anchor 신규 1) 전부 경계 주장 또는 anchor 포인터 보유; 발산 claim 0. 배제 분류: navigator 3종+미러 3종(읽기-집합 배제 선언), navigator 테스트 4건·nonoverlap 2건(픽스처), `apply_test.sh`(픽스처), `frozen-allowlist.json`·`.gitignore:223`(설정), `CHANGELOG.md`(이력), `failure_observer.go`·`lessons_inbox_test.go`(코드), curator 스크립트 4종(코드), `lsel-drain-loop.js`(코드) |
| AC-LCS-005 | **PASS** (M3; regression-guard) | `git diff --name-status d7ce6c6bd..HEAD` | 8파일 전부 markdown(+미러): M `.claude/rules/.../moai-constitution-detail.md` · M `.claude/skills/hns-lsel-curator/SKILL.md` · A `.moai/docs/learning-channel-scope.md` · A `.moai/specs/.../{plan,progress,spec}.md` · M `CLAUDE.local.md` · M `internal/template/templates/.../moai-constitution-detail.md`. 보호 경로 diff (`git diff d7ce6c6bd..HEAD -- internal/hook/failure_observer.go …/drain.sh …/session_drain.sh …/backlog_check.sh internal/graph/`) → `0`행 — 바이트 무변경 |
| AC-LCS-006 | **PASS** (M3; regression-guard) | `git diff d7ce6c6bd..HEAD --name-only \| grep -c '^\.moai/config/sections/'` | `0` — 신규 config 섹션 없음; name-status상 신규 훅·스크립트·포맷 0; `failure_observer.go` 바이트 무변경 → 배선 패밀리 2종 불변; 기록 대상 `feedback_*.md`+`MEMORY.md` 유지 |
| AC-LCS-007 | **PASS** (M2 flip; RED `grep -c '인간 매개 루프' CLAUDE.local.md` → `0`/exit 1 @ `d7ce6c6bd`) | `grep -c '인간 매개 루프' CLAUDE.local.md` | `1`/exit 0; 앵커 포인터 `grep -c 'learning-channel-scope'` → `1`/exit 0 — §28 신설 절 "인박스 유용성 범위 (경계 선언)" |

**(baseline 관측)** `grep -c 'test_fail:' <primary>/.moai/lessons-inbox.jsonl` → run 시작·종료 모두 `0`/exit 1 — dated baseline 유효.

**E2 — 문서 전용 diff 증명**: 위 AC-LCS-005 행 (8파일, 보호 경로 0행).

**E3 — 미러 패리티 + neutrality**: `diff` 로컬 vs 미러 → 유일 차이 `54d53` (로컬 전용 서브불릿 "Scope anchor (maintainer repository): `.moai/docs/learning-channel-scope.md` …"); 공유 본문 바이트 동일. neutrality grep 전 항목 0 (위 AC-LCS-002 행). 가드: `go test ./internal/template/ -run 'TestTemplateNeutralityAudit\|TestTemplateNoInternalContentLeak' -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 1.071s`.

**E4 — 임베드 재생성**: `make build` → rc 0 (`catalog.yaml updated successfully (12899 bytes)` 후 `go build … -o bin/moai` — catalog 재생성 결과가 커밋 버전과 바이트 동일하여 캐스케이드 diff 없음).

**E5 — 일관성 스윕**: 위 AC-LCS-004 행 — §A.5 절차 재실행, 목록 1-4 전부 경계 주장/anchor 포인터, 배제 목록 전부 비-claim 유지 (신규 히트 `.gitignore:223`·nonoverlap 테스트 2건은 배제표 성질과 동류: 설정·픽스처).

**E6 — 인박스 무손상**: 측정은 읽기 전용 jq/grep/find뿐. run 시작 5,958행 → 종료 5,958행 (창 내 외부 append 없음; append-only 성장은 baseline 날짜로 설명 — anchor doc § Dated baseline), `test_fail` 0 유지.

**kickoff 조건부 제안 (failure_observer.go:80 주석 1행 정정) — 미채택**: run 위임 지시가 docs-only 범위(REQ-LCS-005)로 런타임 코드 변경을 금지했으므로 채택하지 않았다. §B.1 관측 기록(:80 헤더 주석이 :109-111 인박스 스텁 누락)만 남기고 종결 — 제약 1의 단일 예외 미발동, Go 소스 0줄 유지.

**MX 스캔**: 터치 파일 8종 전부 markdown — MX 코드 주석 대상 아님. 추가/갱신/제거 0.

**커밋 (E1 귀속)**: M1 `d1044d9d2` (SPEC 산출물 + draft→in-progress 전이 + anchor doc) → M2 `efe39e914` (표면 전파) → M3 (본 커밋, 증거+스윕). 카드 id t260 전 커밋 메시지 포함.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-02
run_commit_sha: "0ea123429"   # M3 커밋 (2026-09-02 후속 커밋에서 backfill — D3 면제 패턴)
run_status: complete
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 8   # plan.md §A.4 PRESERVE 불릿 전수 — 위반 0 (보호 경로 diff 0행으로 관측)
l44_pre_commit_fetch: "n/a — worktree-local branch (WT-learn-channel-gap), push 없음; 통합은 리드 창 소관"
l44_post_push_fetch: "n/a — 본 SPEC run-phase push 없음 (동일 사유)"
new_warnings_or_lints_introduced: 0   # Go 소스 0줄 — lint 표면 불변; 가드 테스트 2종 격리 초록
cross_platform_build:
  applicable: false
  reason: "Go 소스 0줄 변경 (docs-only); make build darwin rc=0 만 관측"
total_run_phase_files: 8
m1_to_mN_commit_strategy: "M1 d1044d9d2 → M2 efe39e914 → M3(본 커밋) — 마일스톤당 1커밋, 결정 우선 순서"
```

## §E.4 Sync-phase Audit-Ready Signal

_sync-phase executed 2026-09-02 (manager-docs, worktree `WT-learn-channel-gap`, 시작 HEAD `3a6db9f16`)._

```yaml
sync_complete_at: 2026-09-02
sync_commit_sha: "pending-backfill-sync"   # D3 SHA-placeholder exemption (spec-frontmatter-schema.md) — 커밋은 자신의 해시를 인용할 수 없어 후속 chore 커밋에서 backfill
sync_status: complete
frontmatter_status_transitions:
  spec.md: "in-progress -> completed (단일 sync 커밋의 3-phase close; implemented는 별도 단계로 존재한 적 없음 — merged-close 관례대로 정직 기록). status + updated만 변경, 본문 무변경. updated: 2026-09-02는 run 종료 시각에 이미 현재일이라 값 불변"
  plan.md: "n/a — status 전이 없음, 본문 무변경"
  progress.md: "프론트매터 없음; §E.4 이번 커밋에서 작성"
changelog_entry_position: "CHANGELOG.md [Unreleased] > ### Added (최신 우선 관례에 따라 첫 불릿으로 삽입, card t196 항목 위)"
ac_count_verified: 7   # grep -o 'AC-LCS-00[0-9]' spec.md | sort -u | wc -l -> 7 (Tier S — acceptance는 spec.md §I 인라인, acceptance.md 없음); §E.2 매트릭스 7 PASS와 일치
duplicate_guard: 0   # grep -c 'SPEC-LEARN-CHANNEL-SCOPE-001' CHANGELOG.md (발행 전) -> 0
b12_self_test_a: "grep -c 'SPEC-LEARN-CHANNEL-SCOPE-001' CHANGELOG.md (pre-emission) -> 0/exit 1 -> PASS, 중복 없음"
b12_self_test_b: "grep -o 'AC-LCS-00[0-9]' spec.md | sort -u | wc -l -> 7 == CHANGELOG 엔트리의 7/7 ACs PASS -> PASS (0은 적색 신호 규정대로 검사 — 실측 7)"
b12_self_test_c: "ls .moai/docs/learning-channel-scope.md .claude/rules/moai/core/moai-constitution-detail.md .claude/skills/hns-lsel-curator/SKILL.md CLAUDE.local.md internal/template/templates/.claude/rules/moai/core/moai-constitution-detail.md -> 5종 전부 존재 -> PASS"
canary_compliance_check:
  documentation_locale: "ko — 단, CHANGELOG.md [Unreleased] 기존 엔트리 전체가 영어 산문이므로 하우스 스타일 우선으로 영어로 발행 (§E.4에 그 근거 기록)"
  git_commit_messages: "en — 커밋 제목·본문 영어"
mx_scan: "run 터치 8파일 + sync 터치 5파일(CHANGELOG, spec.md, progress.md, plan-audit 보고서 2종) 전부 markdown — MX 코드 주석 대상 아님. 추가 0 / 갱신 0 / 제거 0 (확인)"
sync_phase_observations:
  - "run-phase가 docs-only 변경에 feat(...) 커밋 제목을 채택 (리드 제안은 docs(...)) — internal/spec/lint_ownership.go M1 허용 집합 {feat|fix|refactor|perf|test}을 근거로, docs(는 ownerManagerDocs로 분류된다는 이유 제시. 이유 있는 소유권-린트 부합 선택으로 기록하며 수리하지 않음"
  - "이 카드에는 오케스트레이터 증거 로그가 없다 — run 검증은 grep 기반이며 §E.2 매트릭스 + plan-audit 보고서(.moai/reports/t260/{plan-audit-iter1,plan-audit-iter2}.md, 이번 sync 커밋에서 추적 전환)에 기록돼 있다"
sync_session_gaps: "오케스트레이터 증거 로그 부재(위 관측 2행과 동일 사실); CI 판정 부재 — 브랜치 미푸시(위임 제약, 통합은 리드 창 소관)"
readme_review: "해당 없음 — 본 SPEC은 유지자 저장소 운영 문서(.moai/docs 앵커, rules 트리, dev-only 스킬, CLAUDE.local.md)만 건드린다. 인박스 포착 범위를 주장하는 README/docs-site 표면은 없음 (AC-LCS-004 스윕에서 확인 — claim 표면 4종+앵커 외 발산 claim 0)"
known_residual_docs_drift: []
```
