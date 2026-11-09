# [DR][05] 하위 7편 수정안 — 설치·Host·릴리스·PLAN·RUN·소유권·GitHub 작업 구조

## 배경
종합 보고서 §3 (`review-group3-install-host-plan-run-release.md` §1~§9). 대상 7편: [83f] 홈 설치, [dff] Host adapter, [25f] 릴리스·Desktop, [e08] PLAN·GEARS, [3d9] RUN·TDD/DDD, [487] 파일 소유권·worktree, [66a] RC1 GitHub 작업 구조. 모두 ChatGPT Space 페이지이므로 **적용 가능한 수정안 문서**를 만든다.

## 범위
- `docs/design-review/fixes/group3-<pageid>.md` 7개(또는 1개 파일에 7절): 각 검토 절의 "개선 제안(영향 순)" 전부를 (위치, 현재, 교체, 근거) 표로. 우선: [66a] L1270 권위 방향(S1), [25f] §5·§10 결정됨/구현 상태(S2), [83f] §1 경로·F ID(S3·S8), [dff] §9 HA별 RC1 상태(S5), 상태 어휘 대응표(S4 — DR-07 부록 참조)
- X11 "세 문서에 붙여 넣은 같은 문단"은 한 교체문을 세 위치에 동일 적용하도록 표시
- 근거 행은 현재 tip에서 재확인(`src/claim.rs`·`handoff.rs`·`worktree.rs`는 검토 후 변경됨)
- 입력: `references/chatgpt-space/page_{83f162ef…,dff088ea…,25fa9b23…,e0874134…,3d917728…,487fa473…,66acadc2…}.md` 읽기 전용

## 수용 기준
- AC-1: 7편 모두 ≥ 4항목, 각 (위치, 현재, 교체, 근거) 채움, tip SHA 기록
- AC-2: 시험 ID 행마다 RC1 상태(passed receipt / not_run / not_applicable)를 `docs/support-matrix.md`·`docs/e2e-receipts/receipts.jsonl`에서 확인해 기입
- AC-3: 가설 근거 단독 사용 0, 개인 절대 경로·비밀 0건
- AC-4: check set 실측 exit

## 사용자 결정
- 2026-11-09 00:1x 보고서 컨펌
