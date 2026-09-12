# SPEC-MOAI-GPT-AUTH-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

0.1.0 초안. Tier M 문서 3종과 진행 기록을 생성했다. 계획 감사는 아직 수행하지 않았다. 공식 Codex 인증 broker·private scratch·MoAI 원자 publish/세대 경계를 계획에 반영했다.
custom client 등록 질문은 broker 경로의 blocker로 유지하지 않는다. 설치본 RPC·endpoint·새 로그인·refresh 검증은 미실행이다.

## §E.2 Run-phase Evidence

로컬 AUTH wrapper를 구현하고 변경 범위 시험을 실행했다. 전체 AUTH 상태는 미완료다.

| 대상 | Actual Output | Status |
| --- | --- | --- |
| AC-GA-002·003 wrapper ID/취소/종료 | `go test ./internal/gateway/auth` 패키지 PASS에 포함 | 로컬 부분 검증; broker PKCE 실제 시험 미실행 |
| AC-GA-004 private 원자 저장·손상·cleanup | `coverage: 87.1% of statements`, cleanup 실패 시 기존 세대 보존 | macOS 부분 검증; Windows 저장 미지원 |
| AC-GA-005 세대/CAS·다중 프로세스 | 두 process refresh 교환 1회, crash lock 회수, logout 뒤 late refresh 거절 | 로컬 mock 검증 |
| AC-GA-006 최초 header write/logout | 실제 net.Pipe Write 차단 시 logout deadline; SSE는 logout을 막지 않음 | HTTP/1.1 로컬 검증; core 연결 미검증 |
| AC-GA-007·008 endpoint/참조 | 구독·API key endpoint 혼용 거절 | 로컬 부분 검증; 실 endpoint preflight 미실행 |
| AC-GA-009 broker 격리 | fake process 환경·symlink 경계 시험 PASS | 실제 설치 broker 정책 미검증 |
| AC-GA-001·010 실제 계정·네 모델 | 실행하지 않음 | 미완료 |
| race | `ok  github.com/modu-ai/moai-adk/internal/gateway/auth 7.984s` | PASS |
| vet·LSP | exit 0, 빈 출력 | 최종 변경 범위 PASS |
| Windows amd64 | test binary compile exit 0, 빈 출력 | compile만 검증; runtime 미실행 |

상세 RED/GREEN·명령·기준선·Gap은 `../../reports/SPEC-MOAI-GATEWAY-001/auth-implementation-verification.md`에 기록했다.
실제 broker 로그인/refresh/provider verifier, remote revoke, core adapter 연결 및 Windows ACL/내구성 구현은 남아 있다.

## §E.3 Run-phase Audit-Ready Signal

미완료. 제한된 로컬 wrapper는 독립 코드 감사 대상이다. 전체 AC PASS나 run_complete_at을 기록하지 않는다.
commit·push·PR을 수행하지 않았으며 통합 브랜치 CI의 저장소 전체 verdict는 PENDING이다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
