# t656 verdict — SPEC-GITSTRAT-WORKFLOW-READER-001

- 카드: git flow 선택을 설정에서 고르게 하고 실제로 읽는다 (lane-4 원래 전제 "reader 0"는 폐기 — t449/t637 이 reader 를 이미 착지; 운영자 재범위로 **기존 reader 확장**으로 진행)
- 브랜치: `WT-git-flow-reader` (base `b1bd81b23` = 로컬 develop) · 최종 HEAD `4628dcf64` · **미푸시**
- 클래스 C · Tier M · cycle_type tdd · Phase 4 모드 serial

## Claim

카드의 재범위 목표 — `git_strategy.<mode>.workflow` 허용값 4개 검증(무효값 명시적 처분), flow 별 상설 브랜치·통합 대상 해석(github-flow→main, git-flow→develop_branch 현행, gitlab-flow→environment, release-flow→release_branch_prefix), 기존 git-flow 동작 무변경(특성화 먼저), 템플릿 기본값 github-flow 유지 — 이 plan→run→sync 전 주기를 독립 감사 통과와 함께 완수됐고 SPEC 은 completed 로 닫혔다.

## Evidence

- **plan**: 산출물 5종 `cd48ec891` → iter1 FAIL 0.78(MUST-FIX 5건: 거짓 소비자 주장·REQ 토큰·AC 누락·실행불가 증명·서사 모순) → 수리 `242acfaf3` → iter2 **PASS 0.96**(조화평균 0.90/0.95/1.0/1.0, 반복 상한 2/2). 보고서: `plan-audit-SPEC-GITSTRAT-WORKFLOW-READER-001-iter{1,2}.md`.
- **run**: 8커밋(M1 특성화 `08298ae28` → M2 처분 `ea51ce891` → M3 doctor+경고 `a317e5896` → M4 인벤토리 `c8833fbe6` 등). 오케스트레이터 독립 검증 배치 재현: `go build ./...` exit 0, `GOOS=windows go build` exit 0, `go test -cover ./internal/config/ -count=1` → `ok ... 82.2%`, doctor 스코프 6/6 서브테스트 PASS, `golangci-lint run internal/config/... internal/cli/...` → `0 issues.`, AC-GWS-010 diff `08298ae28..HEAD -- loader_integration_branch_test.go` = **0바이트**. RED 증거(M2/M3 build-fail)는 progress.md §E.2.
- **sync-audit**: **PASS 9.0/10**(Functionality/Security/Craft/Consistency 각 9, Critical 0) — 배포 4차원 워크플로 스크립트(`sync-audit-4dim.js`)가 파스 에러(템플릿 리터럴 내 비이스케이프 백틱, 148행)로 기동 불가 → 교리상 폴백 판정 주체인 cold sync-auditor 가 구속 판정. 보고서: `sync-audit-SPEC-GITSTRAT-WORKFLOW-READER-001.md`.
- **close**: 단일 sync 커밋 `d0fe0cc2a`(CHANGELOG [Unreleased], spec.md implemented→completed, §E.4 — `Authored-By-Agent: manager-docs` 트레일러 확인) + 백필 `4628dcf64`(`sync_commit_sha: "d0fe0cc2a"`). MX 수리 `e6c15e056`(@MX:ANCHOR, sync-audit F2).
- **스코프 감사**: `moai spec audit`(filter=본 SPEC) → `modern_era_clean: 1`, drift 0(`EraAutoDetected` INFO 1건뿐).

## Baseline-attribution

위 모든 수치는 이 워크트리(`.claude/worktrees/t656`)에서 이번 실행에 직접 측정한 것 — 각 행의 명령+출력은 progress.md §E.2와 감사 보고서에 verbatim 기록돼 있다. 구현자 수치는 오케스트레이터·감사자가 같은 명령으로 재측정해 교차 확인했다(커버리지 82.2% 양측 일치, 베이스라인은 감사자가 /tmp git-init 추출로 82.0% 재측정).

## Gaps

- `internal/cli` 패키지 수준 커버리지 미관측 — `-cover` 계측 600s 초과(머신 부하 병목), 스코프 `-cover` 실행으로 대체.
- 공유 진단 스냅샷(`moai verify check --key-current`)에 현재 키 기록 없음 — 간격으로 기록, 전 항목 재실행으로 대체.
- 라이브 `moai doctor` 실행은 worktree git-가드 거부 — 테이블 테스트+골든 렌더의 §D.4 간접 경로로 충족.
- docs-site 4로케일 미반영 — 카드 규정상 리드 판단 사항.
- init/update 마법사 workflow 선택지 — D3 연기(마법사가 workflow 값을 쓰지 않음을 실측; 후속 카드 후보).

## Residual-risk

- **AC-GWS-012 PASS-WITH-DEBT**(감사자 구속 판정): 패키지 82.2% < 85.0 리터럴 — 고정 베이스 82.0% 로 사전 결함이며 신규/확장 함수는 100%. 리프트는 범위 밖 레거시 테스트 요구 → **레거시 internal/config 커버리지 후속 카드** 권고(progress.md §E.3 명기).
- Minor 2건 잔존: F3 `workflowStandingBranches` 방어 반환 미테스트, F4 doctor 가 unreadable/absent 를 모두 OK 로 병합 — 후속 카드 후보.
- 후속 카드 후보: acquire/automerge 의 D2 해석 테이블 채택(REQ-GWS-008 동결 유지), 마법사 선택지.
- **배포 스크립트 결함(피드백 대상)**: `sync-audit-4dim.js` 파스 에러 + `node --check` 이 이 파일 클래스에서 거짓-초록(F5) — 워크플로 게이트는 ESM 컴파일/실행 검증이 필요.
