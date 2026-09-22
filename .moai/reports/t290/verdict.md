# 카드 t290 판정 — GLM 설정 섹션 개명 + flash effort 잠금 잔여분 복구

> Evidence-bearing report (AGENTS.md §1). 작성: factory lane-2, 2026-08-27.
> 브랜치 `WT-glm-flash-residue` (워크트리 `.claude/worktrees/t290-apply`) · develop 기점 `8ef14f5ae` · 병합 전 커밋 완료 상태.

## Claim (주장)

primary 체크아웃에 미커밋으로 남아 있던 t289 후속 잔여분(rescue 패치 11파일, 25.8KB)이
develop 기점 브랜치에 두 번째 사본으로 확보됐다. PR #1671과의 중복 반입은 A안(패치 진행·PR 폐쇄,
리드 확정)으로 정리됐고, 3파일 충돌은 t289의 배송된 의미(flash 기본값 서열)를 유지하는 방향으로
해소했다. 이 커밋을 확인하면 리드가 primary의 원본 11파일을 정리할 수 있다.

## Evidence (증거)

적용: `git apply -3 rescue/t290-glm-flash-residue.patch`
- 8/11 클린 적용(glm.go·app.js·i18n.js·console_ux_fix_test.go·glmkey.go·glmkey_test.go·handlers.go·schemaform.go)
- 3/11 충돌(closed_sets.go·defaults.go·glm_tier_test.go) — 파일당 1훙크씩

충돌 해소(ours 채택 근거는 progress.md 참조): 세 훙크 모두 `<<<<<<< ours`=t289 의미(`DefaultGLM53Flash`
선행·전티어 flash 바인딩·want 리스트 flash-first), `theirs`=이전 포지션(glm-5.3 기본). ours 유지 +
theirs가 갖던 "flash는 reasoning_effort max 한정" 설명 주석만 defaults.go와 glm_tier_test.go에 흡수
(grep "sparse-attention" = 1회 — 중복 없음 실측).

독립 검증 배치(본 워크트리, 동일 상태에서 관측):
| # | 항목 | 명령 | 관측 |
|---|---|---|---|
| V1 | 형식 | `gofmt -l <7개 Go 파일>` | 빈 목록(clean) |
| V2 | 마커 잔존 | `grep '^<<<<<<<\|…' 3파일` | 매치 0건 |
| V3 | config 테스트 | `go test ./internal/config/ -count=1` | ok 3.097s |
| V4 | web 테스트 | `go test ./internal/web/ -count=1` | ok 3.283s |
| V5 | cli 테스트 | `go test ./internal/cli/ -count=1` | ok 197.937s, exit 0 |
| V6 | 윈도 빌드 | `GOOS=windows GOARCH=amd64 go build …config/web/cli` | OK |

## Baseline-attribution (baseline 귀속)

모든 관측은 **본 실행에서, `.claude/worktrees/t290-apply`, HEAD(커밋 직전 검증 시점)** 에 대해 수행.
패치 대조근거: PR #1671 파일목록(gh pr view) 및 FETCH_HEAD diff --stat — patch-only 3파일 영흔 실측.

## Gaps (미검증)

- 해소 결과물과 primary 작업 사본의 바이트 동등성은 비교하지 않았다(격리 가드상 외부 트리 순회 금지).
  공유 8파일은 클린 적용이라 semantically 동일하나 문맥 차(develop base 진행)로 행번호 차이 가능.
- `app.js`/`i18n.js` JS 자산의 전용 테스트는 존재하지 않음 — web 패키지 통과(handlers 계약)로 우회 검증.
- i18n.js 4개 로케일 문자열 일관성은 diff 내용상 확인했으나 번역품 심사는 범위 밖.

## Residual-risk (잔여 위험)

- 리드가 primary를 정리한 뒤 로컬 main 동기화 전까지, primary의 무결성 검증(`git status | grep ' D'`)
  은 리드 절차대로 수행되어야 한다 — 본 사본이 브랜치에 있어 손실은 없음.
- PR #1671 폐쇄는 리드 소관(superseded-by-t290). 폐쇄 전 로컬 `WT-glm-settings-rename` 트리 보존.
