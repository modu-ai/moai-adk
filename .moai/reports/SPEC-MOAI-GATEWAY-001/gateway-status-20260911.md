# MoAI Gateway 진행 보고

2026-09-11 18:30 KST · status · basic (MoAI-Easy)

## Claim

핵심 부품의 구현과 로컬 검증은 진행됐으나, 실제 moai gpt에서 Claude Code를 통해 GPT를 사용하는 전체 경로는 아직 연결·검증 전이다. 관련 5개 SPEC의 활성 인수 조건은 60개이며 완료 개수가 아니다.

| 작업 | 활성 AC | 상태 | 확인 범위 | 남은 조건 |
|---|---:|---|---|---|
| 코어 Gateway | 24 | 부품 구현·로컬 검증 | 라우팅·요청 검증·변환·스트림·자식 환경 정리 | 실제 launcher와 gateway 조립 연결, 모델별 한도, 실제 대화·도구 왕복 |
| GPT AUTH | 10 | 부품 구현·실계정 대기 | 격리 저장소·공식 broker 연계·갱신 resolver·종료 경합 수리 | 새 로그인·실제 구독 요청·운영 갱신 연결·Windows CI |
| 모델 선택 PICKER | 9 | 준비 구현·실측 대기 | 초기 모델 인자·모델 목록·관측 스크립트 | 실제 /model 선택·저장 격리·Default·fallback 제어 |
| CG 폐기 | 8 | 로컬 기능·문서 검증 | 폐기 진단·명시적 이전·4개 언어 문서·템플릿 정리 | gateway 연결 뒤 정상 경로 재확인·혼합 팀 이전 |
| TEAMMATE | 9 | 진입점 후보·구현 대기 | 설치 바이너리 진입점 조사·관측 스크립트 | 실제 named teammate 관측·일회용 부트스트랩·세션별 인증과 수명 |

## 운영자 결정

- MoAI 전용 OAuth 등록 정보 없음. 승인된 공식 Codex 로그인 연계 기준을 유지한다. 실제 로그인 성공은 아직 미관측이다.
- Windows 권한·잠금·내구성 시험은 GitHub CI에서 수행한다. 로컬 cross-compile은 실행 증거가 아니다.
- Claude 및 실제 계정 시험은 2026-09-11 19:00 KST 이후에 진행한다.
- 시험 대상: Claude Opus 5·Sonnet 5, gpt-6-astra·gpt-5.6-sol·gpt-5.6-terra·gpt-5.6-luna.

## Evidence

이 작업 세션에서 작성한 범위별 보고서의 기록을 연결했다. 상태 보고서 작성 중 시험을 다시 실행한 것은 아니다. 명령과 원문 출력은 아래 근거 문서에 있다.

- [60개 인수 조건 지도](pre-live-ac-map.md)
- [인증 갱신 resolver](auth-resolve-verification.md)
- [인증·수명 변경분 감사](auth-supervisor-code-audit-iter2.md)
- [스트림 변경분 독립 감사](protocol-functional-review-iter2.md)
- [입력 추정·잘림 정책](context-policy-local-verification.md)
- [CG 실행 파일 진단](cg-diagnostic-fix-parent.md)
- [CG 기능 일관성 감사](cg-functional-consistency-review.md)
- [4개 언어 문서 빌드](cg-docs-verification.md)
- [문서 브라우저 검증](cg-docs-browser-verification.md)
- [템플릿 검증](cg-template-verification.md)
- [Windows 후보 구현](windows-auth-verification.md)
- [CLI 연결 부품](cli-integration-verification.md)

## Baseline-attribution

보고서 생성 시 git에서 직접 읽음: `WT-unified-gateway` · `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`.

Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`. 미커밋 변경·신규 파일이 있으므로 HEAD만으로 전체 구현을 재현할 수 없다. 각 시험의 소스·명령·해시는 해당 보고서에 귀속한다.

source_session_id: `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`

## Gaps

실제 launcher factory/binding 연결, 실제 로그인과 갱신, GPT 4종의 도구·스트림·대화 기억·resume, PICKER 저장 격리, TEAMMATE 인증과 수명, Windows CI 실행 결과가 남아 있다. 기존 인수 조건 지도 작성 이후 ResolveFresh와 입력 크기 추정 함수가 추가됐으나 생산 경로 연결은 남아 있다.

## Residual-risk

로컬 모의 응답과 단위 시험은 실제 공급자·Claude Code 동작을 보장하지 않는다. 입력 크기는 추정치다. 전체 보안 감사·전체 인수 조건 PASS·출시 준비 완료 판정은 없다. 이 작업의 push·PR·병합·배포는 수행하지 않았다.

## 다음 순서

1. 인증과 실제 요청 관측: 19:00 KST 이후 Claude 인증 전달, 공식 Codex 새 로그인과 GPT 구독 요청을 관측한다.
2. 실제 명령과 모델 선택 연결: 관측 결과로 launcher 연결, 모델 선택과 설정 저장, 대화 상태 처리 및 teammate 구현을 마친다.
3. 같은 세션에서 GPT 4종 확인: 실제 도구·스트림·대화 회상·resume 및 실패 경로를 확인한다.
4. Windows CI와 최종 인수 판정: GitHub CI에서 Windows 실행 증거를 확보하고 각 인수 조건을 판정한다.
