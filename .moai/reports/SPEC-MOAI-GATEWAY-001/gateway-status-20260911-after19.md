# 최신 진행 보정 — 2026-09-11 20:54 KST

이 보정은 아래 20:02 스냅샷보다 우선한다. 전체 제품 인수는 아직 미완료다.

- Windows API 저장 기준과 GPT 구독 서버 출력 정책은 사용자 승인 후 SPEC에 반영했다.
- Windows 필수 시험 목록은 기존 9개를 보존한 23개다. 실제 Windows CI는 미실행이다.
- Windows AUTH 구현 감사에서 후보 파일 identity의 지연 조회 결함이 확인되었다. 구현 담당이 handle 기반 identity로 수리하고 macOS AUTH race 및 Windows 교차 컴파일을 통과했다고 보고했다. 변경분 독립 재감사는 남아 있다.
- Factory/SSE의 두 결함은 수정 후 독립 재감사를 통과했다. 새 모델 정책·구독 출력 정책은 구현 후 독립 감사 중이다. 감사에서 native SSE의 이벤트별 CRLF 바이트 계수 문제가 추가 재현되어 수리가 필요하다.
- Receipt core의 마지막 guard 취소 경계를 수리하고 독립 재감사 중이다. Reasoning envelope와 도구 ID codec 구현을 시작했다. 전체 대화 귀속·launcher 연결은 남아 있다.
- 실제 GPT 네 모델의 제목 JSON 응답과 native fork의 로컬 보존 관측을 추가했다. 각각 직접 provider 및 native 모의 경로의 증거이며 제품 통합 완료를 뜻하지 않는다.

근거: `windows-store-functional-review.md`, `windows-store-identity-repair.md`, `factory-sse-functional-review-iter2.md`, `native-policy-verification.md`, `receipt-final-guard-repair.md`, `title-policy-runtime-observation.md`, `native-fork-runtime-observation.md`.

기준: WT-unified-gateway / 81c1d58f9의 미커밋 트리. source_session_id는 developer 제공 01a08e7b-6aa0-7361-ab7e-ea8da1f02228이며 현재 WT CLI는 environment-fallback이다. 각 시험의 명령과 원문 출력은 해당 근거 보고서에 귀속한다. Windows 실환경·전체 인수·원격 게시 완료를 주장하지 않는다.

---

# MoAI Gateway 진행 보고

2026-09-11 20:02 KST · status · basic (MoAI-Easy)

## Claim

GPT 네 모델의 실제 구독 응답과 직접 함수 왕복, Claude OAuth 갱신은 확인했다. 다만 실제 moai gpt 명령부터 Claude Code 대화·도구·재개까지 연결한 제품 경로는 아직 완성·검증 전이다. 60개는 활성 인수 조건 수이며 완료 개수가 아니다.

| 작업 | 활성 AC | 상태 | 확인 범위 | 남은 조건 |
|---|---:|---|---|---|
| 코어 Gateway | 24 | OAuth 구성품 감사 통과·제품 조립 중 | 실제 Claude OAuth 갱신·요청별 자격 격리·라우팅과 일반 변환 | 실제 명령 연결·새 factory/SSE 독립 감사·reasoning 유실 방어 |
| GPT AUTH | 10 | 새 로그인·네 모델 직접 왕복 확인 | GPT 4종 실제 text·함수 왕복, Luna 암호화 item의 다음 턴 보존 | 승인된 출력 정책 구현·Claude 통합·Windows Store와 CI |
| 모델 선택 PICKER | 9 | native 선택·재개·인증 분리 관측 | 네 모델 선택 요청·Default Sol·합성 carrier 재개·원본 설정 보존 | 지속되는 전용 대화 설정·전체 launcher 경로·동시 세션 |
| CG 폐기 | 8 | 로컬 기능·문서 검증 | 폐기 진단·명시적 이전·4개 언어 문서·템플릿 정리 | gateway 연결 뒤 정상 경로 재확인·혼합 팀 이전 |
| TEAMMATE | 9 | native 호출 지점 관측·연결 전 | named Agent가 지정 helper를 실행하는 실제 로컬 관측 | 일회용 부트스트랩·세션별 인증·native 자식과 수명 |

## 운영자 결정

- MoAI 전용 OAuth 등록 정보 없음. 공식 Codex의 새 MoAI 전용 로그인은 실제 성공했다.
- Windows 검증은 GitHub CI에서 수행한다. 필수 native 시험은 9개이며 실제 CI 실행은 아직 대기다.
- 19:00 KST 이후에만 실행한다는 제한을 지켜 실계정 관측을 수행했다.
- 시험 대상은 Claude Opus 5·Sonnet 5와 GPT의 Astra·Sol·Terra·Luna다.

## Evidence

이 작업 세션에서 작성한 범위별 보고서의 기록을 연결했다. 상태 보고서 작성 중 시험을 다시 실행한 것은 아니다. 명령과 원문 출력은 아래 근거 문서에 있다.

- [실제 Claude OAuth 정상 응답](m0-after19-observation.md)
- [직접 OAuth 갱신 관측](m0-refresh-transport-observation.md)
- [요청별 OAuth 독립 감사](oauth-passthrough-functional-review.md)
- [GPT 4종 직접 응답·함수 왕복](auth-responses-runtime-observation.md)
- [모델 선택과 합성 reasoning](picker-reasoning-runtime-observation.md)
- [Astra·Terra·Default 요청](picker-followup-runtime-observation.md)
- [native 정확한 세션 재개](picker-resume-runtime-observation.md)
- [설정 격리와 기존 OAuth 공존](picker-settings-source-observation.md)
- [Windows 잠금·저장 계약](windows-store-contract-gap.md)
- [Windows CI 증거 검사](windows-ci-preparation.md)
- [실제 teammate helper 호출](teammate-runtime-observation.md)
- [구체적인 계약 수정안](after19-contract-proposal.md)
- [reasoning 기록 설계 감사](receipt-plan-audit-iter1.md)
- [CG 실행 파일 진단](cg-diagnostic-fix-parent.md)
- [4개 언어 문서 빌드](cg-docs-verification.md)

## Baseline-attribution

보고서 생성 시 git에서 직접 읽음: `WT-unified-gateway` · `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`.

Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`. 미커밋 변경·신규 파일이 있으므로 HEAD만으로 전체 구현을 재현할 수 없다. 각 시험의 소스·명령·해시는 해당 보고서에 귀속한다.

source_session_id: `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`

## Gaps

실제 명령·gateway 조립, factory와 구독 SSE 감사 결함 두 건의 수정·재감사, 대화별 reasoning 유실 방어, PICKER의 지속 설정과 전체 진입 경로, TEAMMATE 자식 인증·수명, Windows Store와 실제 CI가 남아 있다. 사용자가 GPT 구독 서버 출력 정책과 Windows API 저장 기준을 모두 승인했다. 두 계약의 구현·검증이 남아 있다. reasoning 기록 설계의 변경분 재감사는 PASS이며, receipt core 구현을 시작했다.

## Residual-risk

직접 API 요청, native 모의 TUI, 제품 구성품 시험은 서로 다른 검증이다. 이를 전체 사용자 경로 통과로 합산하지 않았다. 설정·보안 저장소 분리 변수는 설치 Claude 2.1.268에서 관측한 내부 경로이며 공개 지원 약속은 확인하지 못했다. 전체 AC PASS·Windows native 실행·출시·push·PR·병합은 아직 완료하지 않았다.

## 다음 순서

1. 승인된 계약 구현: GPT 구독 출력 정책·Windows API 저장 기준을 반영하고, 감사 결함 두 건과 receipt core를 구현·검증한다.
2. 실제 명령과 세션 연결: gateway 조립과 설정·원래 인증의 분리, 재개 가능한 대화 상태 및 teammate 연결을 구현한다.
3. Claude Code 안에서 GPT 4종 확인: 같은 실제 세션의 도구·스트림·대화 기록·resume 및 잘못된 입력의 송신 차단을 검증한다.
4. Windows CI와 최종 인수 판정: 준비된 필수 시험을 GitHub에서 실제 실행하고 각 인수 조건을 판정한다.

## 확정된 두 운영자 결정

GPT 구독은 서버 출력 정책을 사용하고 MoAI의 byte·취소 제한은 별도로 유지한다. Windows는 파일 flush·같은 디스크 원자 교체·권한/내용 재확인과 GitHub CI의 프로세스 종료 후 복구 검증을 적용한다. 두 안 모두 사용자가 승인했다.

[정확한 수정 문안](after19-contract-proposal.md)
