# progress — SPEC-CODEMAPS-ACCURACY-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC-ID: SPEC-CODEMAPS-ACCURACY-001 · status: draft · Tier M · 2026-09-02
- 산출물: spec.md / plan.md / acceptance.md / progress.md (Tier M 4종 — spec.md §5 Tier 분류 근거)
- 사전검사: SPEC-ID 정규식 `PASS` (실행 출력 인용, plan §E); 카탈로그 충돌 0건 (`SPEC-CODEMAPS-ACCURACY-*` 부재 — 인접 ID SPEC-DWF-CODEMAPS-PILOT-001·SPEC-V3R6-DOCS-CODEMAPS-V3-001는 별개)
- 조사 완료: §1.1 부재 8개 분류표 완결 (P1–P8) · §1.2 ListActive 3지점 실API 대조 (registry.go 직독) · §1.3 생성기 입력 재발 경로 특정 (스킬 Phase 2 병합 입력 + Phase 4 비기계 검증) · 설계 결정 D1–D4 확정
- 병합 순서 제약 선언: run-phase는 M0(origin/develop 흡수 + 재측정, REQ-CMA-008)부터 — t432 병합 전 codemaps 편집 금지
- plan-phase는 `.moai/project/codemaps/`·`internal/` 무편집 · 커밋 없음 (레인 소관)

## §E.2 Run-phase Evidence

### M0 — M2 게이트 + post-absorb 재측정 (REQ-CMA-008)

- M2 게이트 (2026-09-02, 병합 흡수 전, 트리 `061985ec8`):
  - `git fetch origin develop -q` → exit 0
  - `git merge-base --is-ancestor 7e1c4d94f origin/develop && echo MERGED || echo NOT_MERGED` → **`MERGED`** (t432 재생성 착지 — 디스패치 작성 시점의 NOT_MERGED 평가는 무효화됨)
  - 흡수: `git merge origin/develop` → merge commit `d79bdd2b3` (충돌 0)
- §1.1 재측정 (post-absorb 트리 `d79bdd2b3`): 전수 추출 `grep -ohE '\b(internal|pkg|cmd)/[A-Za-z0-9_/.-]*' .moai/project/codemaps/*.md | sed 's/[.,;)$]//g' | sed 's:/$::' | sort -u` → **104 유니크 토큰** (수정 후 기준). 정리 규칙(후행 문장부호/슬래시 제거·`cmd/moai/main`→`cmd/moai/main.go`·`.go` 복원) 적용 존재 검사 결과:
  - **P1 `internal/design`** — modules.md blockquote 노트 유지 (양성 0) → 조치 없음 ✔
  - **P2 `internal/migrate`** — modules.md:163 blockquote ✔ · **P3 `internal/state`** — modules.md:248 blockquote ✔ · **P4 `internal/research`** — modules.md:272 blockquote ✔ · **P5 `internal/evaluator`** — modules.md:319 blockquote ✔
  - **P6 `internal/factory`** — modules.md:169 `### internal/factory` 양성 절 잔존 (t432 재생성 후에도) → M2에서 `internal/kanban` 절로 병합 재작성 (REQ-CMA-003) — 유일한 양성 유령이었음
  - **P7 `internal/bodp`** — t432가 dependencies.md:185 blockquote 각주로 전환한 상태로 착지 → M2에서 5개 노트와 같은 굵은 토큰 형식으로 정렬 (REQ-CMA-004/D3)
  - **P8 `cmd/moai/main`** — overview.md 호출-연쇄 표기, 정규화로 실존 판정 ✔
- §1.2 재측정: `ListActive` data-flow.md 3개 지점 **197행(mermaid 노드)·214행(흐름 단계)·357행(인터페이스 블록)** — 전부 잔존, M2에서 실API로 수정 (아래 AC-CMA-003).
- 좌표 정정 참고: spec.md §1의 develop 기준 행번호를 본 progress.md 재측정 좌표가 대체한다 (manager-develop는 spec.md 본문 편집 금지 — plan §C 3항의 "spec.md §1 좌표 정정 커밋"은 소유 규칙과 충돌하여 progress.md 기록으로 충족).

### M1 — citations 검사 축 (Go) [REQ-CMA-002, D1/D2]

- RED (E8 원본: `.moai/reports/t304/m1-red-evidence.md`): `go test ./internal/graph/ -run TestCheckCitations -count=1` → exit 1, `undefined: checkCitations / LayerCitations / MetricPositiveCitedPathAbsence / normalizeCitedPath` (구현 전 테스트).
- GREEN: `internal/graph/check_citations.go` (checkCitations·positiveCitedPaths·normalizeCitedPath), `check.go` (LayerCitations·MetricPositiveCitedPathAbsence 상수 + CheckFreshness 4행 구성), `internal/cli/graph_check.go` (help text). exit-code 계약 0/1/2 불변 — 신규 행은 기존 LayerReport/Failed()/OffendingLayers() 소비 경로 그대로 통과.
- 테스트 4종+CLI 1종: 유령 red+뮤턴트 왕복 / blockquote 면제 / 정리 규칙 표 / 부재-디렉터리 unjudgeable / CheckFreshness 4-레이어 / CLI exit-1 citations 행 지목.

### M2 — codemaps 사실 수정 [REQ-CMA-001/003/004/005/006]

- modules.md: `internal/factory` 유령 절의 내용을 기존 `### internal/kanban (33 non-test 파일)` 절에 병합 (중복 제목 제거, 서술 보존 — REQ-CMA-003). 개명 계기 언급은 양성 행의 부재 경로 재인용이 되어 삭제 (REQ-CMA-001 허용 집합은 P1–P5+P7로 열거 완결).
- data-flow.md: ListActive 3지점 → `QueryActiveWork` (mermaid·흐름 단계), 인터페이스 블록 → 실제 리시버 메서드 4종(`Register(sessionID, specID, phase string) error`·`Heartbeat(sessionID string) error`·`Deregister(sessionID string) error`·`Query(optSpecID string) ([]Entry, error)`) + 패키지 함수 서명 (REQ-CMA-005, registry.go 직독 대조).
- dependencies.md: bodp 각주를 5개 경고 노트와 같은 `> **`internal/bodp`** —` 형식으로 정렬 (REQ-CMA-004/D3).
- 수정 후 전수: 부재 집합 = {P1–P5, P7} 7토큰, 전원 blockquote 행 (양성 부재 0) — AC-CMA-001.
- 뮤턴트 왕복 (실문서, `go run ./cmd/moai graph check --json`): `internal/zzz-phantom` 제목 주입 → citations value=1 verdict=stale driving_paths=[internal/zzz-phantom] → 원복 → value=0 verdict=fresh (AC-CMA-005).

### M3 — 스킬 재발 방지 (Template-First) [REQ-CMA-007]

- 템플릿 정본 편집 → 로컬 미러 동일 내용 (diff 확인 IDENTICAL) → `make build` 성공. Phase 2: 기존 codemaps 콘텐츠를 패키지 존재의 권위에서 배제; Phase 4: `moai graph check` citations 행 실행 명령 + 부정 인용 blockquote 규약 ("negative citations MUST use blockquote form" — 형식 규정이지 blockquote 독점 주장 아님) + 신선도 초록≠정확성 명시.

### M4 — F1 정정 + t432 무결 [REQ-CMA-009]

- `.moai/reports/t304/f1-record.md`: t432 보고서 §3.1 제목 "26항목" vs 27행 표 (251행 Gaps 라인 "27항목"과 상충) 정정 기록 + t432 트리 무결 관측 (HEAD ref `WT-codemaps-refresh`, 본 카드 쓰기 0; 교차 `git -C`는 worktree-session 가드가 거부 — ref 파일 직독으로 대체).

### AC 판정 행렬

| AC | 상태 | 판정 명령 | 관측 출력 (요지) |
|----|------|-----------|------------------|
| AC-CMA-001 | PASS | 전수 grep+정리규칙 존재검사 (위 M0/M2 명령) | 부재 집합 = 7 blockquote 토큰 정확히 일치, 비-blockquote 부재 0 |
| AC-CMA-002 | PASS | `grep -n '### internal/factory' modules.md` / `grep -c '### internal/kanban'` / 양성 bodp grep | 0행 / 1 / 0건 + blockquote 노트 1 |
| AC-CMA-003 | PASS | `grep -c 'ListActive' data-flow.md` / registry.go 서명 대조 | 0 / 인터페이스 블록 = 리시버 4종+Entry+QueryActiveWork, `Session` 인용 0 |
| AC-CMA-004 | PASS | known-5 토큰 출현 검색 | 각 1건 이상 blockquote, 양성 0 |
| AC-CMA-005 | PASS | 테스트 3방향 + 실 gate `--json` + 뮤턴트 왕복 | red/green/면제 성립; 실문서 citations verdict=fresh value=0; 뮤턴트 stale→fresh 왕복 관측 |
| AC-CMA-006 | PASS | 스킬 양본 Phase 2/4 검사 | 실행 명령+규약 문구 양본 동일 존재, `make build` 성공 |
| AC-CMA-007 | PASS | 본 §E.2 M0 블록 | 게이트 명령+출력+재측정 좌표 기록됨 |
| AC-CMA-008 | PASS | `.moai/reports/t304/f1-record.md` 존재+내용 | F1 정정 1건 + t432 무결 관측 기록 |
| AC-CMA-009 | PASS | 본 행렬 | 전 AC 관측 출력 제시; 신선도 green은 정확성 근거로 미인용 (citations는 정확성 축으로 판정) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-02
run_commit_sha: "fa7b06bf0"   # sync-phase backfill (manager-docs): M4 최종 run-phase head — pending-backfill-97bff985c 플레이스홀더 대체
run_status: complete
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 5   # known-5 경고 노트 (P1-P5) 전량 잔존 관측
l44_pre_commit_fetch: "git fetch origin develop -q → exit 0"
l44_post_push_fetch: "not pushed — lead batch (레인 push 금지, gitflow-lane-protocol §4)"
new_warnings_or_lints_introduced: 0   # golangci-lint ./internal/graph/... ./internal/cli/... = 0 issues (baseline 0과 동일)
cross_platform_build:
  darwin: "exit 0"
  windows_amd64: "exit 0"
total_run_phase_files: 11   # go 5 (check.go, check_citations.go, graph_check.go, + test 3 포함 별도) · codemaps 3 · 스킬 2 · progress 1
m1_to_mN_commit_strategy: per-milestone conventional commits, no push, no amend, no --no-verify
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-02
sync_commit_sha: "pending-backfill-<parent-of-sync-commit>"   # 커밋은 자신의 SHA를 모른다 — 리드/lane 후속 백필 (D3 SHA placeholder exemption)
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-CODEMAPS-ACCURACY-001' CHANGELOG.md → 0 (중복 없음, emission 진행)"
b12_self_test_b: "grep -oE 'AC-CMA-[0-9]+' acceptance.md | sort -u | wc -l → 9; CHANGELOG entry cites 9 ACs (9 PASS)"
b12_self_test_c: "모든 인용 경로 ls 검증 — internal/graph/check_citations.go · internal/graph/check.go · internal/cli/graph_check.go · codemaps.md 양 미러 존재 확인"
changelog_entry_position: "CHANGELOG.md [Unreleased] → Added 선두 (SPEC-CODEMAPS-ACCURACY-001)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (본 sync 커밋에서 3-phase close 병합 — 별도 Mx 커밋 없음)"
  updated_bumped: "2026-09-02 (동일 — 당일 close)"
canary_compliance_check:
  spec_body_edits: 0   # spec.md/plan.md/acceptance.md 본문 무편집 (frontmatter status/updated만)
  codemaps_regeneration: "not executed — accuracy는 run-phase 수리 + citations 축이 담당; 재생성은 범위 밖 (follow-up)"
  docs_site_4locale: "not expanded — citations 축·스킬 규약은 user-facing이나 본 크기 문서화는 scope 밖으로 판정 (blocker/scope-estimate 불요 — 스킬 양미러가 사용자 면 커버)"
mx_tag_changes: "added 0 / removed 0 / updated 0 — checkCitations에 @MX:NOTE [AUTO] 이미 존재(run-phase); LayerCitations·MetricPositiveCitedPathAbsence 상수는 fan-in 2(check.go·check_citations.go)로 ANCHOR 기준 미달, NOTE 추가 불요"
```

## §F Phase 4 Mode Selection

**Input parameters**: tier M · scope ~9 files (codemaps 2·3문서 + internal/graph + internal/cli + 테스트 + 스킬 로컬/템플릿 미러 2 + SPEC 산출물) · domain count 4 (Go 코드, 문서, 스킬 템플릿, 증거) · file language mix markdown+Go · concurrency benefit LOW (coding-heavy, 마일스톤 간 의존 M0→M2) · Agent Teams prereqs 미충족(명시 요청 없음)

**Mode evaluation**:

| Mode | 선택 | 근거 |
|------|------|------|
| direct | not selected | 다중 파일·Go 신규 코드 — trivial 아님 |
| serial | **selected** | coding-heavy (Anthropic coding-task caveat) + 마일스톤 순서 의존 + 문서 단일 작성자 |
| fanout | not selected | research-heavy 아님 — 병렬 이득 낮음 |
| sweep | not selected | ~30파일 기계변환 아님, 새 코드 포함 |

**Decision**: serial

**Justification**: 구현이 Go 신규 코드(M1)와 문서 수리(M2)·스킬 편집(M3)의 순서 의존 체인이라 병렬화 이득이 없고, Anthropic의 coding-task parallelism caveat에 따라 serial이 기본이다. 마일스톤당 manager-develop 1회 spawn, 병합 순서 제약(M0 게이트)이 상태를 공유하므로 단일 작성자가 옳다.

**Autonomy**: Implementation Kickoff Approval 통과(운영자가 본 레인 터미널에서 AskUserQuestion으로 직접 승인, 2026-09-02) · `ac_converge` goal 무장 후 **해제**(2026-09-02, 리드 지시 + t436 결함 — `moai goal` 산문 조건 오분류 계열. 직접 관측: 본 조건은 `conditions[0].type = model`로 정상 분류돼 t436 트리거(후행 `exits <N>` → mechanical)가 발화하지 않았고 turns 0 — 그러나 결함 축이 오늘 활발히 관리 중이므로 리드의 전 레인 무장 해제 방침을 따름) · **반자율 전환**: 마일스톤 경계마다 완료 보고 후 진행, 운영자 결정은 blocker 보고로 lead-1 경유
