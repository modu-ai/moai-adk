# Todo 감사 개선 — 커밋·카드 완료 기록

## Claim

22개 개선의 구현·근거를 로컬 브랜치 `WT-todo-audit`에 커밋했고, 이번 감사 카드 `t647`를 완료 보관한 뒤 다시 조회했다. 원격 반영과 로컬 구현 완료는 구별한다.

## Evidence

커밋 직전 브랜치·HEAD 확인:

```text
ee99507fb
WT-todo-audit
```

`git fetch origin main --quiet` 이후 `git rev-list --count --left-right origin/main...HEAD`는 `0 2993`이었다. 작업 트리의 예상 HEAD·브랜치 일치를 조건으로 확인하고 55개 파일을 명시적으로 스테이징했다. 기존 다른 작업의 `t647/verdict.md`는 포함하지 않았다.

커밋 명령의 출력:

```text
[WT-todo-audit 3e03b1dc6] fix(todo): harden queue integrity and refresh performance (card t647)
 55 files changed, 3737 insertions(+), 603 deletions(-)
```

`git log -1 --format='%H%n%s'` 재조회:

```text
3e03b1dc68f7a5b23bfc83f315aff24c2d2ebc7a
fix(todo): harden queue integrity and refresh performance (card t647)
```

구현 커밋 직후 `git status --short`는 출력이 없었다. 이후 변경은 이 완료 기록과 보고서의 커밋·상태 표기다.

카드 종결:

```sh
moai todo done t647 --expect '[TODO 전수 감사 2026-09-11'
```

```text
done t647 landing=unknown
```

`moai todo history t647` 전체 원문으로 이번 감사 카드임을 확인한 후 첫 세 필드를 다시 확인했다.

```sh
moai todo history t647 | cut -f1-3
```

```text
t647	archived	picked
```

`picked`는 보관 직전의 상태이며 현재 위치는 `archived`다. 원문과 이력은 보존되므로 `moai todo undone t647`로 복원할 수 있다. 이 복원 명령을 실사용 카드에 실행하지는 않았다.

HTML 생성 후 구조 검사:

```text
HTML_REPORT_VALID: ko, 22 finding rows, result/evidence/risk sections, linked evidence files present; bytes=29713
```

이는 완료 상태 문장을 추가하기 전 HTML의 측정이다. 한국어 문서 설정·22개 finding 행·결과/근거/위험 섹션·근거 파일 존재를 검사했고, `open reports/todo-audit-20260911.html`이 exit 0이었다. 최종 상태를 추가한 뒤 다시 생성·구조 검사·열기를 실행하여 다음 출력을 얻었다.

```text
HTML_FINAL_VALID: ko, 22 findings, implementation commit, archived state, 5 evidence files; bytes=30130
```

실제 브라우저 화면의 시각 품질을 자동 검증했다는 뜻은 아니다. 구현 커밋 이후의 tracked diff는 보고서 Markdown·HTML과 판정 문서뿐이고, 새 파일은 이 완료 기록이다. 생산 코드와 테스트는 변경하지 않았다.

## Baseline-attribution

- 기능 검증 기준: `ee99507fbe3b4a22c6a0a74815723d222dfdc04d` + 이번 구현 변경.
- 구현 커밋: `3e03b1dc68f7a5b23bfc83f315aff24c2d2ebc7a`.
- source_session_id: `01a08e64-c9d3-76f3-a589-5d5d893e1b62`.
- 최종 시험 명령·출력: [verdict.md](verdict.md).

## Gaps

원격 push/PR/merge, 전체 저장소 CI, 설치 바이너리 교체, 실사용 홈 데이터 이관은 수행하지 않았다. `landing=unknown`을 원격 착지 성공으로 해석하면 안 된다. Git 설정은 manual/local 및 auto_push/auto_pr=false다.

## Residual-risk

이 구현이 들어 있는 worktree는 원격 통합 전까지 보존해야 한다. 기존의 같은 번호를 사용하는 과거 보고서와는 날짜별 경로 및 카드 원문으로 구별한다. 기술적 잔여 위험은 최종 판정 문서를 따른다.
