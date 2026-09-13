# t560 두 번째 통합 창 기록 — 병합

창 보유: lane-8. 리드 지명 뒤 `moai integration status` → `free` 확인, `moai integration acquire --name lane-8` → exit 0, 이어진 status → `held`, holder lane-8.
워크트리 `.claude/worktrees/t560`, 브랜치 `WT-hook-env-scrub`. 창 직전 HEAD `35d4c8c89`, 추적 수정 0, MERGE_HEAD 없음.

## 흡수

- 흡수 대상: 로컬 `develop` `526249cf682af1105d06bd8b42a5d9942533e6e9`(리드가 알린 tip 과 `git rev-parse develop` 일치).
- 지난 흡수 기준 `84e5666d9` 가 `526249cf6` 의 조상: `git merge-base --is-ancestor` exit 0 (`develop-ancestry.txt`).
- 사전 예측: `git merge-tree --write-tree --name-only 526249cf6 HEAD` → 트리 `2dff1237…` 만 출력, 충돌 줄 없음, exit 0 (`merge-tree-526.txt`).
- 실행: `git merge --no-ff -m "…(card t560)" develop` → exit 0, `Merge made by the 'ort' strategy.`, 76 files changed, 15530 insertions(+), 삭제 0 (`absorb-merge.log`).
- 흡수 병합 커밋 `7a76beccacd8c735cc574bf038d19f4e550e1be9`, 부모 `35d4c8c89` · `526249cf6`, 트리 `2dff123743722ee577fccf9184403201f2c606a4` — 사전 예측 트리와 같다.

## 델타가 t560 패키지에 닿는가

리드 지시: 닿지 않으면 테스트 재측정을 생략하고 앞선 창 결과와 리드 판정을 인용, 닿으면 멈추고 보고.

- `git diff --name-only 84e5666d9 526249cf6` → 76파일 (`develop-delta-files.txt`, 커밋 목록 `develop-delta-log.txt`: t661·t553).
- `^internal/(gitenv|hook)/` → 0. 대조 `^internal/cli/` → 3.
- 보고서도 `_test.go` 도 아닌 파일 → 출력 없음(grep exit 1). 코드 변경은 `internal/cli` 의 테스트 파일 3개뿐이다(`main_test.go`, `update_clean_install_test.go`, `update_skip_sync_test.go`).
- 판독: 닿지 않는다. `_test.go` 는 다른 패키지의 컴파일에 전이되지 않으므로 internal/gitenv·internal/hook·hook/quality·hook/security 와 훅 사본 테스트의 입력은 첫 창과 같다. 기록: `delta-check.txt`.
- 흡수 뒤 `git diff --name-only develop HEAD` → 31파일, 전부 t560 커밋 파일(보고서 16, CHANGELOG, `internal/gitenv` 2, `internal/hook` 12).

## 테스트 — 재측정 생략, 인용

- 첫 창(`../window/summary.md`): gitenv·hook/quality·hook/security ok, 훅 사본 동일성 RUN 2 / PASS 2, internal/hook 은 `TestSessionStart_DeferredScanDoesNotBlockReturn` 1건 FAIL(559ms > 500ms, 부하 18~21).
- 격리 재실행(`../rerun/summary.md`): 그 테스트는 실패 목록에 없음. GLM 테스트 2건 실패는 격리 형태의 부산물.
- 리드 판정 (ii): 추가 재실행 없이 병합. 해당 핸들러·테스트 blob 이 develop tip 과 같아 t560 에 귀속되지 않는다(`../rerun/summary.md` 판정 절).

## 증거 인용 정정

`../window/summary.md` 가 인용한 파일 이름 6개가 실제 커밋된 이름과 달랐다(실제 이름에는 `t560-` 접두사가 붙어 있음). 인용을 실제 이름으로 고쳤다: `t560-merge-tree-84e.txt`, `t560-window-merge.log`, `t560-changelog-vs-develop.diff`, `t560-added-now.txt`, `t560-added-orig.txt`, `t560-ss-handler-tests.txt`, `t560-home-isolating-tests.txt`. 내용과 수치는 바꾸지 않았다.

## 미검증

- 흡수 트리 `2dff1237…` 위에서 테스트를 다시 돌리지 않았다(리드 지시로 생략). 근거는 위 델타 판독뿐이다.
- `TestSessionStart_DeferredScanDoesNotBlockReturn` 의 559ms 원인(부하 대 동기 admission·lease 작업)은 여전히 가르지 못했다.
- CI(darwin·windows 매트릭스)는 보지 않았다.
