# card-review — t1587 (lane-7, 2026-10-08)

- backend: codex (mcp codex_review scope=card — 120s 초과로 배경 전환 후 완료)
- base: `48a96cbb1` · tree: `.moai/worktrees/t1587` @ `e588f9882` (브랜치 WT-init-retry-filters, 카드 커밋 2건)
- verdict: **fail (advisory)** — 3건 전부 P2. **카드 diff(`db6d875f6`+`e588f9882`: internal/cli/init.go·init_user_asset_retry_test.go·.github/test-input-filters.yml·test_input_filter_parity_test.go)에 대한 발견 0건.**
- 성격: 3건 = 카드가 건드리지 않은 파일의 기존 원장 항목(아래 처분표).

## 발견별 처분

| # | 발견 | 좌표 | 처분 |
|---|---|---|---|
| 1 | [P2] Windows 획득 가드 비정상 종료 복구 부재(하루 지난 marker가 획득 차단 재현) | internal/userassets/lock_guard_windows.go:23 | **carried-over — t1591(1) 보유**(리더 P1 승격 기록 확인됨). 본 카드에서 처분 안 함 |
| 2 | [P2] 미완료 설치 journal 복구 없이 bundle 제거가 journal을 삭제 → 미추적 파일 잔존(collision 오판 후속) | internal/cli/bundle.go:145 | **신규 기존 항목 — 리더 원장 적립 요청**(다음 접촉 시 일괄 전달). t1509 저널 계열(게이트 4차 install.go:239/255와 같은 crash-recovery 가족) |
| 3 | [P2] Deploy는 공통 자산을 제외하는데 ListTemplates는 포함 → managedRedeployCount·AnalyzeMergeChanges 과보고(SKILL.md 재현) | internal/template/deployer.go:192 | **carried-over — 리더 원장 기보유**(본 세션 게이트 2차 릴레이 3번 항목과 동일 파일·계열) |

## 판정

- 카드 diff 자체에는 codex 교차에서 **반례 0건** — 결함 (1) 수리(순서 재배치)와 결함 (2) 수리(필터 3경로+핀) 모두 지적 없음.
- advisory 결과로 카드 PASS/FAIL은 리더의 증거 판독에 귀속(스테이지 규정). 레인 자체 검증 배치(RED→GREEN 전문·사전/사후 패밀리 재측정·vet/gofmt/golangci 클린)는 진행 기록 §Run-phase Evidence 참조.
- ceiling: 수리 대상 발견 0건 — 재검토 불요.

decision record: decided_by=lane-7 evidence_refs=codex_review 응답 전문(본 파일 인용)+t1591 보유 기록+리더 원장 접수 회람 ladder_path=card-review 스테이지 advisory 처분
