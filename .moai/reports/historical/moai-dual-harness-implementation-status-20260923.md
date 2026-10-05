# MoAI-ADK 이중 하네스 구현 현황

> 상태 보고 · 2026-09-23 · 기준: `WT-dual-harness-parity-rebuild`의 `ad18a641c`와 아래 미커밋 역할 권한 변경. **전체 완료 판정: FAIL.**

## Claim

Claude Code와 Codex CLI를 위한 공통 리소스 배포, Codex worktree 생성, 일부 역할 권한 축소는 로컬에 구현했다. 설계 문서의 전체 동등 지원 계약은 아직 충족하지 못했다. `develop` 반영과 두 CLI의 실사용 인증도 완료되지 않았다.

## 구현된 범위

| 영역 | 로컬 구현 | 현재 판정 |
|---|---|---|
| 템플릿 배포 | `claude`·`gpt`·`both`별 배포기와 `.moai/policies`·`.moai/workflows` 참조 해석, 사용자 파일 보존 | 관련 단위 검사 통과; 실사용 미인증 |
| 실행기 | `moai codex -w [name]`이 기존 L1 트리를 재진입하거나 설정된 기준에서 새 트리 생성 | 관련 단위 검사 통과; 여러 실행 환경 미인증 |
| 에이전트 | `mission-governor`·`super-advisor`를 Codex `read-only`로 생성하고 쓰기 도구 충돌 거부 | 생성기 검사 통과; 호스트 적용 실측 없음 |
| 설계 | 전체 목표·13개 인수 기준을 Markdown/HTML로 기록 | 설계 산출물이며 구현 증거 아님 |

## Evidence

이 실행에서 관찰한 출력이다. Go 캐시 기본 경로는 샌드박스에서 거부되어 `GOCACHE=/tmp/moai-dual-go-cache`로 재실행했다.

```text
$ git branch --show-current
WT-dual-harness-parity-rebuild
$ git rev-parse --short HEAD
ad18a641c
$ git rev-list --count --left-right origin/develop...HEAD
0    2
$ GOCACHE=/tmp/moai-dual-go-cache go test ./internal/template/agentemit -count=1
ok  github.com/modu-ai/moai-adk/internal/template/agentemit  0.343s
$ GOCACHE=/tmp/moai-dual-go-cache go test ./internal/template -run 'TestHarnessProfilesResolveSharedReferences|TestRuleTemplateMirrorDrift|TestTemplateNoInternalContentLeak|TestRuleDateProvenance|TestRuleProvenanceAudit' -count=1
ok  github.com/modu-ai/moai-adk/internal/template  1.957s
$ GOCACHE=/tmp/moai-dual-go-cache go vet ./internal/template/agentemit
(exit 0; 출력 없음)
$ git diff --check
(exit 0; 출력 없음)
```

앞선 같은 작업 트리 측정에서 `go test ./internal/template -count=1` 및 범위를 좁힌 `internal/cli` 테스트가 통과했다. 그 뒤 `origin/develop` 병합이 있었으므로 이 보고서의 현재 기준 PASS에는 위 재실행 결과만 포함한다.

## Baseline-attribution

위 증거의 기준은 이 작업 트리의 `ad18a641c`와 역할 권한 변경 파일이다. 배포된 `moai` 바이너리, 공유 기본 체크아웃, 사용자 홈의 실제 Claude/Codex 설정에 대한 증거로 해석하지 않는다.

## Gaps

| 인수 기준 | 판정 | 빠진 증거 또는 구현 |
|---|---|---|
| AC-POL-01 | 미인증 | 모든 필수 의무의 양쪽 적용·검사 연결 |
| AC-TPL-01/02 | 부분 | 참조 검사 통과; 모든 새 설치·갱신·사용자 수정·복구 조합 미실행 |
| AC-HOOK-01/02 | 미완 | Claude 필수 Stop 체인과 Codex의 효과 동등성, compact·permission·interrupt 실제 발화와 거부 보존 |
| AC-GOAL-01 | 미완 | Codex의 지속·취소·예산 종료 실측 |
| AC-AGENT-01 | 부분 | 두 역할 생성 검사만 통과; 12개 역할의 실제 로드·도구 권한·산출물 검증 없음 |
| AC-WT-01 | 부분 | 새/기존 Codex 트리 검사; 동시 writer·미통합 삭제 방지 전체 미검증 |
| AC-MSG-01/FACT-01 | 미인증 | 중복·재시작·늦은 응답 및 혼합 팩토리 네 조합의 카드 전체 흐름 |
| AC-MCP-01 | 미인증 | 두 호스트의 연결·도구 호출·승인·취소 전체 흐름 |
| AC-MIG-01 | 미완 | 안전한 uninstall/rollback과 중단 복구 |
| AC-WF-01 | 미인증 | 17개 명령 × 두 CLI의 필수 계약과 실패 사례 |
| AC-OBS-01 | 미완 | 실행 환경·코드·로그에 귀속된 지원 상태 진단 |

`internal/cli` 전체 패키지 테스트는 이 작업 중 장시간 출력 없이 대기해 중단했으므로 PASS가 아니다. Codex live 테스트는 기본 `~/.codex` 상태 저장소 생성이 샌드박스에서 거부됐다. 임시 `CODEX_HOME`의 app-server 초기 handshake만 확인했으며 실제 모델 턴·훅 효과는 확인하지 않았다. 팩토리 테스트 네 개는 실행 명령의 종료 코드가 0이었지만 모두 `SKIP`이므로 PASS가 아니다.

## Residual-risk

Codex TOML의 역할별 sandbox는 파일 쓰기를 제한하지만 역할별 shell/MCP 도구 호출 제한까지 표현하지 못한다. `sync-auditor`는 판정 파일을 써야 하므로 현재 `workspace-write`를 유지한다. 호스트 기능의 표현 한계와 미실행 실사용 검증 때문에 이번 변경을 전체 동등 지원으로 홍보하면 승인·종료·복구 동작을 과장하게 된다.

로컬 변경을 `origin`의 새 브랜치로 보내는 push는 자동 승인 심사에서 거부됐다. 거부 이유는 원격 저장소로 코드가 전송되며 사용자에게 해당 원격·브랜치 push 승인이 명시되지 않았다는 것이다. 따라서 원격 PR·`develop` 병합도 진행하지 않았다.
