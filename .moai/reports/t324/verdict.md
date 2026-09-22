# t324 판정 보고서 — develop 브랜치 보호 상태검사 필수 승격 설계

- 날짜: 2026-09-02 (KST) / 2026-09-01T19:33Z
- 레인 세션: card t324 · worktree `.claude/worktrees/t324` · branch `WT-devprot-required`
- 분기 기준: `origin/develop` @ `fa8ff89ba`

## 산출물

- `.moai/specs/SPEC-DEVPROT-REQUIRED-001/spec.md` (v0.2.0) · `plan.md` · `acceptance.md` · `progress.md` · `research.md`
- Tier M · GEARS 13 REQ / 14 AC (릴리스 차단 13 + 회귀 가드 1)

## 감사

- plan-auditor iter-1: **FAIL** 0.875 (MP-7 clarification gate — NEEDS CLARIFICATION 3건; D2 AC 3건 누락; D3 research.md 반증 전제 잔존)
- plan-auditor iter-2: **PASS** 0.99 (단조 0.875→0.99, Tier M 상한 2/2 내). 잔여 D5/D6/D7 전부 MINOR·선택.
- 근거: `.moai/reports/t324/plan-audit-verdict.md`

## 설계 요약 (3축)

- (a) 필수 검사: 1단계 = `Test (ubuntu-latest)`, `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)` — push에서 무조건 보고되는 컨텍스트만. `Release PR Multi-OS Gate`(구조적 부재)·test-install(paths 필터→Pending 위험) 배제.
- (b) 레인 push 경로: B1 병합커밋 사전검증 — 병합 SHA를 `verify/<card-id>` ref에 push → CI 초록 → 동일 SHA를 develop에 push (statuses는 SHA 귀속이므로 통과). 창 대기는 `gh run watch --exit-status` 자동화, 커밋 상태 API 폴백. B2(enforce_admins=false 관리자 우회)는 전이 옵션으로 병기.
- (c) enforce_admins: true는 B1 경로가 살아난 뒤에만 가능 — 롤아웃 순서(워크플로 변경 → 창 절차 → 보호 적용 마지막)를 어기면 통합 창이 벽돌된다. 적용 명령(runbook, apply+rollback)은 운영자 게이트.

## 자동 해결된 결정 3건 (운영자 기각 가능 — 카드 마감 검토에서)

근거와 함께 plan.md §A.1에 기록:

1. `Analyze (Go) (go)` 1단계 포함 — main이 이미 필수로 요구(라이브 API 실측), codeql analyze 잡은 push 이벤트에서 무조건 실행(codeql.yml:59-63), develop/main 정합성
2. verify ref = 카드별 `verify/<card-id>` — 통합 락 `--force` 이탈 경쟁 상태 내성 + 카드별 실행 귀속 + 착지 후 ref 정리
3. B1 CI 대기 = `gh run watch --exit-status` 자동화 — 창 실행 주체가 에이전트 레인, `scripts/ci-watch`가 저장소 기존 관용구

## 리드에게 전하는 운영 사실

- 브랜치/HEAD: `WT-devprot-required` @ (커밋 후 갱신 — 본 문서와 함께 커밋됨)
- 미푸시 커밋: 1 (이 커밋 — **push하지 않음**: WT 브랜치 push 금지, develop 병합 창이 유일한 공개 경로)
- 증거 경로: `.moai/specs/SPEC-DEVPROT-REQUIRED-001/` + `.moai/reports/t324/` — 모두 이 브랜치에 커밋됨 (primary 반출은 리드가 병합하면 자동 해소)
- 재측정 범위: 감사자가 전 주장을 기계적으로 재검증(저자 말 믿은 것 0건) — moai spec lint exit 0, 마커 grep 0건, 라이브 gh api develop 404/main 5컨텍스트, codeql.yml 조건 직접 판독, `fa8ff89ba` check-runs 실측
- **참고**: primary 체크아웃에 세션 시작 전부터 untracked `.moai/reports/t324/`가 존재했음(세션 시작 git status 스냅샷). 이 브랜치가 develop에 병합되면 해당 경로가 tracked가 되므로, primary의 untracked 사본과 충돌할 수 있음 — 병합 창 전에 리드가 정리(내용 대조 후 제거) 권장
- 이 SPEC은 plan 완료 상태(`draft` + audit-ready)로 합류 — run-phase(워크플로 YAML 편집 + 창 절차 문서 + runbook)와 보호 적용은 별도 운영자 결정
