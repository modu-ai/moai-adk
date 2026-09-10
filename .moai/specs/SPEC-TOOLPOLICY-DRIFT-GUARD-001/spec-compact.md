# spec-compact.md — SPEC-TOOLPOLICY-DRIFT-GUARD-001

## Requirements

- REQ-TDG-001: YAML 이 유도하는 allow/ask/deny 집합(env_gate 제외·중복 제거)이 커밋된 settings.json 권한 블록 세 목록과 집합으로 같음. allow 7개(CronCreate·CronDelete·CronList·EnterPlanMode·ExitPlanMode·EnterWorktree·ExitWorktree) 선언, MultiEdit allow 와 Glob/Grep/Write 경로 deny 미선언. defaultMode·기타 권한 키는 비교 제외, ask 부재 = 빈 목록
- REQ-TDG-002: settings.json 과 settings.json.tmpl 바이트 불변, 실제 트리 대상 build 실행 금지(스크래치만)
- REQ-TDG-003: `make build` 선행 검사, 워킹 트리의 settings.json 권한 블록과 YAML 판독, 차이는 명세자·결정·소속 쪽과 조정 방법으로 보고, 쓰기·재생성 없음, settings.json 또는 YAML 변경 시 CI Go 테스트 작업에서 실행
- REQ-TDG-004: 파일 부재·해석 실패·권한 블록 부재·빈 allow/deny 는 실패(건너뛰기 없음), 목록 내 중복·allow/deny 겹침은 별도 실패, 한 항목 어긋남은 붉은색·복원은 초록
- REQ-TDG-005: 구조적 방지 주장 전수 정정 — YAML 머리말 두 진술(드리프트 방지, 머리말 생성), types.go:94-98 주석, metadata.generated_into 템플릿 항목

## Acceptance criteria

AC-TDG-001(커밋 트리 통과, PASS 줄 개수 1·1) · AC-TDG-002(스크래치 build allow=114 ask=0 deny=48 env_gated_skipped=5, 정렬 diff 0) · AC-TDG-003(git diff d1b61005d..HEAD 0, 검사 전후 sha256 동일) · AC-TDG-004(YAML ExitWorktree 삭제 → rc≠0·문구 포착, 저장 사본 복원 → sha256 일치·rc=0) · AC-TDG-005(Mutation 하위 5개 PASS) · AC-TDG-006(FailClosed 하위 5개 PASS, t.Skip 0, 양성 대조) · AC-TDG-007(주장 grep 0, 검사 이름 양성 대조) · AC-TDG-008(make -n build 에 검사 문구, ci.yml 필터 한 줄) · AC-TDG-009(toolpolicy·cli·config 영향 테스트 PASS) · AC-TDG-010(list JSON 항목 단위 개수)

## Files to modify

- [MODIFY] .moai/config/sections/tool-policy.yaml (20개 항목 정합, 머리말·generated_into 정정)
- [NEW] internal/config/toolpolicy/drift_check_test.go (검사·뮤테이션·실패 폐쇄)
- [MODIFY] Makefile (tool-policy-drift-check 타깃, .PHONY, build 선행)
- [MODIFY] .github/workflows/ci.yml (go_code 필터에 .claude/settings.json, 근거 주석)
- [MODIFY] internal/config/toolpolicy/types.go (:94-98 주석)

## Exclusions (What NOT to Build)

- settings.json.tmpl 과 YAML 대조(GitMode 조건문, 다른 파일 쌍)
- settings.json 권한 항목 변경, 생성기 바이트 형태로 재작성
- 생성기 정렬·서식, env_gate 훅 방출, YAML 부재 시 CLI 무동작 변경, 자동 재생성
- defaultMode·기타 권한 키·settings.local.json·사용자 전역 설정 비교
- 종결된 SPEC-V3R6-TOOL-POLICY-SSOT-001 AC-TPS-005 소급 수정
