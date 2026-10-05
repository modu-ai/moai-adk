# spec-compact.md — SPEC-WIN-SMARTPATH-001

## Requirements

- REQ-CWSP-001: GOOS 주입 생성기(buildSmartPATHFor) + windows 분기(홈 2항목 + SystemRoot\System32 + 실존 프로브 Git Bash 후보, POSIX 배제, PathListSeparator 유지), darwin/linux 바이트 동일
- REQ-CWSP-002: windows SmartPATH 의 실제 템플릿 렌더가 `;`-조인 Windows 전용 값(jsonEscape exactly-once 유지)
- REQ-CWSP-003: 프로브 false 후보 미채택, 후보 전무 시에도 홈 항목 유지(빈 값·POSIX 금지)
- REQ-CWSP-004: 기존 불변 스위트(WSL2/#467/EssentialDirs/requiredKeys/exec-form) 무수정 green + :244 주석 annotated correction
- REQ-CWSP-005: 수리 판정은 GOOS 주입 테이블·렌더 테스트 — 크로스빌드는 smoke 한정
- REQ-CWSP-006: Windows 주장 기록의 재현성 분할(미검증이지 반증이 아님)

## Acceptance criteria

AC-CWSP-001(windows 분기 존재, RED grep 0) · AC-CWSP-002(windows 정확-문자열 테이블, 뮤턴트 A/B 실패) · AC-CWSP-003(darwin/linux 바이트 안정성+픽스처) · AC-CWSP-004(windows 렌더 테스트) · AC-CWSP-005(트립와이어 전수) · AC-CWSP-006(크로스빌드 smoke+패키지) · AC-CWSP-007(프로브 semantics) · AC-CWSP-008(재현성 기록 규율, count 0)

## Files to modify

- internal/template/settings.go (유일 소스 수정 — 리팩터링+windows 분기)
- internal/template/settings_test.go (테이블·렌더 테스트·:244 주석 정정)

## Exclusions (What NOT to Build)

- update.go 전역 PATH 부활 정책 변경(#598 핀 — 값이 올바르면 무해)
- 템플릿 훅 형태 변경(exec form 유지) · 템플릿 파일 자체 무변경
- darwin/linux 출력 변경(바이트 동일 금지) · WSL2 동작 변경
- CI windows 러너 확인 · CLAUDE_ENV_FILE 대안 검토
