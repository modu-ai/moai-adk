# t748 sync-audit — SPEC-CODEMAPS-FOLD-GUARD-001

측정 대상: worktree `.claude/worktrees/t748` · branch `WT-codemaps-fold-ac` · HEAD `509d048c1` · sync-auditor(opus/high) 수행, 이번 실행 재측정. 평가 프로필: built-in default (flat weighted) — must-pass = Functionality + Security.

## 판정: **PASS** — 92/100 (harmonic)

must-pass 방화벽(Functionality 95 + Security 95) 양 차원 독립 통과. 차단 findings 0.

## 차원 점수

| 차원 | 점수 | 판정 | 핵심 증거(이번 실행 재측정) |
|---|---|---|---|
| Functionality 40% | 95 | PASS | `go test ./internal/graph/ -count=1` ok 42.35s · 조건 스코프 9/9 서브테스트 PASS · AC-CFG-001..005 전부 테스트 코드 재유도 |
| Security 25% | 95 | PASS | 테스트 전용·입력 저장소 국소·fixture 쓰기 t.TempDir 한정·secret 0; 스탬프 독립 grep — provenance/graph.check 코드 참조 0건(주석만) |
| Craft 20% | 88 | PASS | lint 0 issues 재실행·모든 실패 경로 단언(전수 판독)·양방향 단언·RED 3건 verbatim byte 일치 — 감점 F1/F2/F5계 |
| Consistency 15% | 92 | PASS | 동결 표면 zero-diff(codemaps·t475·t747·check.go·check_citations.go)·브랜치 diff 6파일 중 internal/은 테스트 1파일뿐(test-only 검증)·커밋 규약 준수 |

## AC 재판정 (5/5 PASS)

- AC-CFG-001 PASS — 가드 존재+실트리 녹색(RED-now ls EXIT=1 → 본 실행 PASS 전환).
- AC-CFG-002 PASS — 변조→FAIL(단위·문서 명명), 복원→PASS; RED 근거 byte 일치(`prlink_landedref.go appears in modules.md:329`).
- AC-CFG-003 PASS — **이중 잠금 실재**: `runFoldGuard`가 문서 스캔 전(:136-143) 파싱 단위와 `floorFoldUnits`(:51-57) 정확 일치 대조 → floor 행 삭제·변형·주석 삽입 모두 FAIL(RED probe B verbatim).
- AC-CFG-004 PASS — 스탬프 독립 구조적 증명(/tmp fixture에 provenance.json 없이 PASS)+checker 무변경(check.go/check_citations.go `fold` 참조 0 — 감사관 소스 grep 보강).
- AC-CFG-005 PASS — 모든 실패 경로 단언, floor 단언 선행으로 공허 녹색 부재, 기본 스위트 편입 재확인.
- **D1 차단 조항** — 문서 부재·판독 불가 실제 FAIL(:153-155 `generator document missing or unreadable`), fixture가 data-flow.md 명명 FAIL 고정. **단일 판독 경로 실재** — 실트리+전 fixture 스캔이 같은 `runFoldGuard`(:124) 경유.
- **위반 술어 공격**: 백틱·코드블록 인용 무관 적중 / `core/git` 단축형 적법 유지(fixture 고정) / PAIR SOURCE `fieldsets_codex.templ`은 `.templ`≠`_templ` 부분열 비충돌로 미탐지(실트리 PASS 교차 확인). **파일명 단독 언급은 회피됨 → F1.**

## 결함 (전원 optional — 판정 무영향)

- **F1 [Medium][선택]** `codemaps_fold_guard_test.go:85-90` — 파일명 단독 토큰 회피: `foldToken`이 `.go`를 떼면 끝이라 토큰 `internal/kanban/prlink_landedref` 형태인데 생성기 문서가 디렉터리 없이 파일명만 대면 미탐지. plan §F 규약("기록의 경로·파일명 그대로")과 사고 서명(전체 경로 6곳)은 충족, AC 어느 것도 이 형태를 요구하지 않아 blocking 아님. 수리: stem 토큰 2번째 추가+bare stem 0건 대조군 — 별도 소형 카드 또는 SPEC amendment(급하지 않음: 사고 서명은 전부 덮임).
- F2 [Low][선택] :81-84 — `foldToken` 주석이 "stem 노출·§A.2 미러"라 하나 실제는 `.go`만 절삭(§A.2 grep은 bare stem으로 더 넓음). 주석 정정.
- F3 [Low][선택] :337-340 — 디렉터리 토큰 단언의 복합 조건이 현재값에서 준-공허. `foldToken("internal/core/git") == "internal/core/git"` 직접 동등 단언으로 교체.

## 권장

F2/F3는 수 줄짜리 정정으로 묶어 경제적. F1 stem 토큰은 별도 소형 카드 또는 amendment. AC-CFG-004의 간접 판정(체커 소스 무변경)은 감사가 소스 grep으로 보강 — 다음 체커 계약 카드에서 `moai graph check` 레이어 출력 직접 대조 한 줄 추가 권고.

## Gaps (감사 직접 미관측)

`moai graph check` CLI 실행·레이어 출력 diff(소스 zero-diff는 직접 관측) · GOOS=windows 빌드 재관측(레인 E2 주장 귀속, 테스트 전용 저위험) · 커버리지 %(프로덕션 statement 0추가로 N/A).

## 잔여 위험

토큰을 전혀 대지 않는 산문 재기술은 가드 회피(REQ-CFG-002 명시 수용) · 파일명 단독 언급 회피(F1, 수용 잔여보다 좁은 새 틈) · 가드 발화는 CI 기본 스위트 편입 유지에 의존.

— 본 파일은 sync-auditor 전달 보고를 lane이 보존한 것(감사자 미작성 규약).
