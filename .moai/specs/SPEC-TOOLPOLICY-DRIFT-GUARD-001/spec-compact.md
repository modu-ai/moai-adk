# spec-compact.md — SPEC-TOOLPOLICY-DRIFT-GUARD-001

## Requirements

- REQ-TDG-001: YAML 이 유도하는 allow/ask/deny 집합(env_gate 제외·중복 제거)이 커밋된 settings.json 권한 블록 세 목록과 집합으로 같아야 함. allow 7개(CronCreate·CronDelete·CronList·EnterPlanMode·ExitPlanMode·EnterWorktree·ExitWorktree) 선언, MultiEdit allow 와 env_gate 없는 Glob/Grep/Write 경로 deny 12개 미선언, env_gate 항목은 삭제 대상 아님. defaultMode·기타 권한 키는 비교 제외, ask 부재 = 빈 목록
- REQ-TDG-002: settings.json 과 settings.json.tmpl 바이트 불변, 실제 트리 대상 build 실행 금지(스크래치만)
- REQ-TDG-003: `make build` 선행 검사, 워킹 트리의 settings.json 권한 블록과 YAML 판독, 차이는 명세자·결정·소속 쪽과 조정 방법으로 보고, 쓰기·재생성 금지, settings.json 또는 YAML 변경 시 CI Go 테스트 작업에서 실행
- REQ-TDG-004: 파일 부재·해석 실패(YAML 문법, 권한 블록 JSON)·권한 블록 부재는 실패(건너뛰기 없음), YAML allow·YAML deny·settings allow·settings deny 중 하나라도 비면 실패(ask 빈 것은 제외), 목록 내 중복·allow/deny 겹침은 별도 실패, 한 항목 어긋남은 붉은색·복원은 초록
- REQ-TDG-005: 방지·머리말 생성 주장 다섯 곳 정정 — tool-policy.yaml:5-9, tool-policy.yaml:11-16 (ANALOGOUS drift class), types.go:1-5 패키지 주석, types.go:94-98 Metadata 주석, tool-policy.yaml:54 generated_into 템플릿 항목

## Acceptance criteria

AC-TDG-001(정정 전 붉은색 서로 다른 20줄·7/1/12/0 분할, 정정 뒤 PASS 줄 1·1, only-in 0) · AC-TDG-002(스크래치 build allow=114 ask=0 deny=48 env_gated_skipped=5, 정렬 diff 0) · AC-TDG-003(merge-base 기준 git diff 0, 검사 전후 sha256 동일) · AC-TDG-004(가드 호환 평범한 명령, 백업·원본 sha 선기록, ExitWorktree 삭제 → rc≠0·차이 줄 정확히 1, 복원 전 변이 sha 확인, 복원 → sha 일치·rc=0·only-in 0) · AC-TDG-005(Mutation 하위 5개 PASS) · AC-TDG-006(FailClosed 하위 9개 PASS: missing×2, malformed×2, no_permissions_region, empty×4, t.Skip 0) · AC-TDG-007(다섯 대상 부재 grep 0, 검사 이름 양성, 파일별 grep 대조) · AC-TDG-008(make -n build 에 TestToolPolicyDrift_, go_code 블록 안 .claude/settings.json 1) · AC-TDG-009(toolpolicy·cli·config 영향 테스트 PASS 개수) · AC-TDG-010(list JSON 항목 단위 개수, env_gate Write deny 1·env_gate 5 보존)

## Files to modify

- [MODIFY] .moai/config/sections/tool-policy.yaml (20개 항목 정합, 머리말 두 문단·generated_into 정정)
- [NEW] internal/config/toolpolicy/drift_check_test.go (검사·뮤테이션·실패 폐쇄)
- [MODIFY] Makefile (tool-policy-drift-check 타깃 `@` 접두, .PHONY, build 선행)
- [MODIFY] .github/workflows/ci.yml (go_code 필터에 .claude/settings.json, 근거 주석)
- [MODIFY] internal/config/toolpolicy/types.go (:1-5 패키지 주석, :94-98 Metadata 주석)

## Exclusions (What NOT to Build)

- settings.json.tmpl 과 YAML 대조(GitMode 조건문, 다른 파일 쌍)
- settings.json 권한 항목 변경, 생성기 바이트 형태로 재작성
- 생성기 정렬·서식, env_gate 훅 방출, YAML 부재 시 CLI 무동작 변경, 자동 재생성
- env_gate 항목 삭제
- defaultMode·기타 권한 키·settings.local.json·사용자 전역 설정 비교
- 종결된 SPEC-V3R6-TOOL-POLICY-SSOT-001 AC-TPS-005 소급 수정
