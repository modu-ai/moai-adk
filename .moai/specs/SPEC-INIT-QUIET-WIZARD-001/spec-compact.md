# spec-compact.md — SPEC-INIT-QUIET-WIZARD-001 (v0.1.4)

## Requirements

- REQ-IQW-001: init 위저드 = 정확히 4문항(`conversation_language`, `user_name`, `agent_wiring`, `autonomy_tier`), 이 순서
- REQ-IQW-002: 제거 14문항은 init 에 제시하지 않음
- REQ-IQW-003: 묻지 않은 제거 키는 오늘 기본값 수락 사용자와 같은 값으로 해석 — 단, 파일시스템 루트 init 의 `project_name` 은 제외(미측정)
- REQ-IQW-004: reconfigure 12문항 ID·순서 불변(D1)
- REQ-IQW-005: 대화형은 MCP 항목 보장 호출 기본 실행(codex 건너뜀·both 실행 규칙 유지, D2)
- REQ-IQW-006: 비대화형 현행 유지(보장 호출 생략, 템플릿 `.mcp.json` 에 moai, 무플래그 workflow.yaml 바이트 동일)
- REQ-IQW-007: `--project-mode`·`--worktree-auto-create`·`--autonomy-tier`·`--llm` 플래그 기록 유지
- REQ-IQW-008: 워크트리 init 기록 규칙 서술 하나로 정리(F1 모순 문서 3곳)
- REQ-IQW-009: "미설정=기본값" 은 실제 init 실행 결과(디스크·로더 해석값) 관측으로 입증, 목록 검사만으로는 불충분
- REQ-IQW-010: 관측기가 기본값 아닌 값을 잡음 — 테스트 안 음성 대조군 + 기록된 코드 뮤턴트
- REQ-IQW-011: init 셸 설정 단계는 교체 가능한 테스트 시접을 거침(운영 기본 = 현재 기록 동작), 테스트는 실제 쓰기 없이 호출 관측, 실행 테스트보다 먼저 착지
- REQ-IQW-012: 새로 쓰거나 본문을 다시 쓴 init 실행 테스트는 대조 8항목(settings.json sha256, hooks/moai 존재, rc 6개 mtime·sha256) 불변 — 8항목 밖 홈 경로는 미측정 Gap. 필드 참조만 지우는 기존 테스트는 기존 HOME 헬퍼 유지
- REQ-IQW-013: 다른 호출자가 없는 제거 질문 흔적(번역·분기·필드·매핑·기록기) 삭제, 호출자 있으면 유지
- REQ-IQW-014: 기존·신규 init 실행 테스트를 돌리는 모든 run 슬롯은 선언 후 실행 전후 실제 홈 지문을 같은 명령·stderr 분리로 채집해 §E.2 에 기록
- REQ-IQW-015: 대화형 4문항 init 의 섹션 파일 5개(workflow·project·report·feedback·llm)는 같은 이름 디렉터리의 무플래그 비대화형 init 과 바이트 동일
- REQ-IQW-016: 지문 차이나 stderr 비어 있지 않음 → 작업 멈춤·리드 보고, 실제 홈 파일은 되돌리지 않음

## Acceptance criteria

AC-IQW-001(init 4문항·순서·그룹, 스윕 확인) · AC-IQW-002(제거 부재·번역/분기 고아 0, 공유 3문항 reconfigure 잔존) · AC-IQW-003(reconfigure 테스트 본문 무변경·update_wizard.go 를 카드가 바꾸지 않음 — 흡수한 develop 과의 merge-base 기준, 병합 전 전용) · AC-IQW-004(셸 설정 단계 스파이 실행 관측 1회·게이트 true 0회/false 1회·기본값 직접 읽기·`ConfigOptions` 본문 보존 추출 비교·배선 제거 뮤턴트 A/B·필드 누락 뮤턴트 E RED 기록) · AC-IQW-005(카드가 추가한 init 실행 테스트 스윕 ∪ 계획한 3파일, 하한 3·도달 대조군, HOME Setenv·Parallel·옛 HOME 헬퍼 부재, 네 번째 파일 뮤턴트 F, 가드 음성 사례, 전후 대조 8항목) · AC-IQW-006(실행 테스트 해석값 표) · AC-IQW-007a(관측기 음성 대조군) · AC-IQW-007b(코드 뮤턴트 A/B FAIL 기록) · AC-IQW-008(대화형·비대화형 섹션 파일 5개 바이트 동일, `workflow.audit` 아래 model·gates·`todo` 미삽입) · AC-IQW-009(MCP 기본값 claude/codex/both) · AC-IQW-010(비대화형 고정 테스트 본문 무변경) · AC-IQW-011(플래그 기록) · AC-IQW-012(워크트리 서술 정리) · AC-IQW-013(죽은 기록기 0·유지 항목 존재) · AC-IQW-014(영향 패키지 테스트·vet·lint, 빈 스윕 없음) · AC-IQW-015(슬롯 선언 수 = 지문 파일 수 = `home-diff-exit=0` 기록 수, 앞뒤 동일 명령·stderr 분리) · AC-IQW-016(`ConfigureShellEnvFn` 대입 테스트·헬퍼 전부를 검증 시점 스윕으로 뽑음(빈 스윕 실패) — `t.Parallel` 부재는 텍스트+`t.Setenv` 충돌 panic 실행, 원복은 대입 줄 수·`t.Cleanup` 텍스트+`-count=2` 재진입 실행, 뮤턴트 C1·C2·D 기록)

## Files to modify

- internal/cli/wizard/questions.go (InitQuestions·Page3Questions)
- internal/cli/wizard/wizard.go (saveAnswer·saveBoolAnswer 분기만)
- internal/cli/wizard/types.go (WizardResult 필드 11개)
- internal/cli/wizard/translations.go (11 ID × ko/ja/zh)
- internal/cli/init.go (매핑 삭제·대화형 MCP 기본값·주석)
- internal/core/project/initializer.go (셸 설정 단계 시접, 죽은 기록기 호출부·필드·주석), initializer_expansion.go, initializer_audit.go (죽은 기록기)
- 새 테스트: internal/cli/init_quiet_wizard_test.go, internal/cli/init_home_guard_test.go, internal/cli/init_shell_seam_test.go, internal/core/project/initializer_shell_seam_test.go
- 갱신·삭제 테스트: plan.md §H 목록 (필드 참조만 지우는 실행 테스트는 기존 HOME 헬퍼 유지)

## Exclusions (What NOT to Build)

- `--no-mcp` 플래그(운영자 결정)
- reconfigure 가 버리는 `report_format`·`project_name` 답(t588)
- 자율 등급 재정의(t584) · 하네스 3-way 배포(t585) · 위저드 렌더링·i18n·huh v1 통합·그룹 라벨 재구성(t586)
- web 콘솔에 project.name·project.mode·`.mcp.json` 프로비저닝 표면 추가
- 다른 SPEC 본문·HISTORY 수정(sync 단계)
- `--force` 재초기화·기존 `.mcp.json` 동작(미측정 Gap)
- `runInitForAutonomyAtHomeCapturingOut` 와 호출 테스트의 시접 기반 헬퍼 이관(후속 후보, 카드는 리드가 발행)
- `moai update` 의 셸 설정 경로(`internal/cli/update.go:824`)
