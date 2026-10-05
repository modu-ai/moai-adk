# MoAI-ADK Claude Code / Codex CLI 작업 핸드오프

> 2026-09-23 · 원 세션 `01a0c84c-0085-7242-8f4e-14e3b43b2906` · **전체 완료 아님**

## 다음 세션에 전달할 지시문

`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/dual-harness-parity-rebuild`에서 MoAI-ADK의 Claude Code/Codex CLI 동등 지원 작업을 이어가라. 먼저 이 문서와 `reports/moai-dual-harness-full-design-20260922.md`, `reports/moai-dual-harness-implementation-status-20260923.md`를 읽고, 실제 `git status --short`, branch, HEAD, `origin/develop`과의 차이를 다시 측정하라. 설계의 13개 인수 기준에서 미완·미인증 항목을 구현하고 양쪽 CLI 실사용 테스트까지 완료하라. 테스트의 `SKIP` 또는 훅 설정 파일 존재를 통과로 세지 마라. 공유 기본 체크아웃과 다른 세션의 dirty `develop` worktree를 건드리지 마라. 원격 push는 이전 자동 승인 심사가 거부했으므로 승인 없이 재시도하거나 우회하지 마라.

## 작업 위치와 Git 상태

- 격리 작업 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/dual-harness-parity-rebuild`
- 브랜치: `WT-dual-harness-parity-rebuild`
- 확인한 HEAD: `ad18a641ca066f268ec96439168b2884671b4994`
- 커밋: `8925682d2` 공통 배포·Codex worktree 생성, `ad18a641c` 최신 `origin/develop` 31개 커밋 병합
- 현재 측정: `git rev-list --count --left-right origin/develop...HEAD` → `0 2`. 다음 세션에서 재측정 필수.
- 기본 체크아웃은 공유 `main`이고 dirty, 별도 `develop` worktree도 다른 작업과 충돌 상태였다. 두 트리의 파일을 정리·reset·stash·병합하지 마라.
- 원격 새 브랜치 push는 **자동 승인 심사에서 거부**되었다. 사유: 명시 승인 없는 `origin` 새 브랜치로 저장소 코드 전송. 같은 push를 재시도하거나 간접 수단으로 우회하지 마라. 로컬 작업을 마친 뒤 사용자에게 정확한 원격·브랜치·전송 범위 승인을 요청해야 한다.

## 미커밋 파일

`internal/template/agentemit/`의 `agentemit_edge_test.go`, `agents-codex.yaml`, `golden_test.go`, `manifest.go`, `writer.go` 및 생성된 `internal/template/templates/.codex/agents/moai/{mission-governor,super-advisor}.toml`에 역할별 sandbox 변경이 있다. `mission-governor`와 `super-advisor`는 `read-only`; 허용하지 않은 sandbox 값과 읽기 전용 역할의 `Write`·`Edit` 도구를 거부한다. 이것은 로컬 테스트만 통과했고 Codex 런타임 로딩은 실측하지 않았다. 이 파일들을 외부 세션의 변경으로 오인해 덮어쓰지 마라.

다음 문서도 untracked이다.

- `reports/moai-dual-harness-full-design-20260922.md`
- `reports/moai-dual-harness-full-design-20260922.html`
- `reports/moai-dual-harness-implementation-status-20260923.md`
- 이 핸드오프 문서

디자인 HTML/MD는 원래 공유 기본 체크아웃에서 작성된 것을 격리 트리로 복사했다. 제안 설계이며 완료 증거가 아니다.

## 검증 증거와 한계

이 작업 트리에서 관찰한 최신 명령/출력:

```text
GOCACHE=/tmp/moai-dual-go-cache go test ./internal/template/agentemit -count=1
ok  github.com/modu-ai/moai-adk/internal/template/agentemit  0.343s

GOCACHE=/tmp/moai-dual-go-cache go test ./internal/template -run 'TestHarnessProfilesResolveSharedReferences|TestRuleTemplateMirrorDrift|TestTemplateNoInternalContentLeak|TestRuleDateProvenance|TestRuleProvenanceAudit' -count=1
ok  github.com/modu-ai/moai-adk/internal/template  1.957s

GOCACHE=/tmp/moai-dual-go-cache go vet ./internal/template/agentemit
(exit 0; 출력 없음)

git diff --check
(exit 0; 출력 없음)
```

앞선 단계에서 전체 `internal/template` 및 범위를 좁힌 `internal/cli` 검사도 통과했으나 그 뒤 `origin/develop`을 병합했다. 병합 뒤 다시 실행한 Codex worktree/init/update 관련 선별 테스트는 PASS였고, `internal/cli` 전체 테스트는 장시간 대기 후 중단했다. 전체 패키지 PASS로 기록하지 마라. Codex live 테스트는 기본 `~/.codex` 상태 저장소 쓰기가 샌드박스에서 거부되었고, 임시 `CODEX_HOME`에서는 app-server 초기 handshake만 확인했다. 팩토리 네 조합 테스트 명령은 exit 0이었으나 네 개 모두 `SKIP`이었다.

## 설계 기준 대비 다음 작업

1. **M2 훅·승인·목표:** Claude 필수 Stop 체인과 Codex 효과 동등성, `PreCompact`·`PostCompact`·`PermissionRequest`·`Interrupt` 실제 발화/결정 보존을 구현·검증한다. `internal/codexadapter/events.go`는 12행 중 8개만 adapted이고, Codex `internal/hook/stop.go`는 advisory allow 경로다. Codex 훅의 timeout 안에서 장시간 검사를 돌리지 말고 설계의 receipt 방식을 검토한다.
2. **M3 협업·권한:** Codex `-k` 진입, 역할 전체의 런타임 권한, 새/기존 worktree 및 동시 writer 보호를 검증한다. 역할 TOML은 shell/MCP 개별 도구 제한을 표현하지 못한다. `sync-auditor`는 판정 파일을 써야 해 `workspace-write`가 남아 있다.
3. **M4 상태·복구:** `internal/codexwiring/wire.go`의 설치뿐인 경로에 사용자 파일 보존형 rollback/unwire 및 중단 복구를 설계·구현한다. 파일 해시만으로 사용자 소유 설정을 삭제하면 안 된다. 메시지 중복/재시작/fencing과 혼합 팩토리 네 조합의 카드 전체 흐름을 테스트한다.
4. **M5 인증·진단:** 17개 명령 × 두 CLI의 성공/대표 실패 fixture, MCP handshake·호출·승인·취소, 진단의 출처 귀속을 실행한다. macOS 이외 환경과 Desktop/Web은 별도 프로파일이며 미검증을 완료로 표시하지 않는다.

모든 required 기준의 실제 효과가 확인되기 전에는 “완벽 지원” 또는 “전체 완료”로 보고하지 마라. 보고서는 Claim, 실행 명령과 원문 출력, 측정 HEAD/트리, 미관찰 항목, 잔여 위험을 분리한다.

## 재진입·변경 시 주의

작업 트리는 `moai cc -w dual-harness-parity-rebuild` 또는 `EnterWorktree(<path>)`로 재진입한다. `git -C <path>`로 명령을 실행하고 기본 체크아웃의 브랜치를 바꾸지 마라. 설치된 `moai` 바이너리는 어느 기준 브랜치와 비교해야 현재인지 아직 결론이 없다. `moai session current`로 원 세션 ID를 확인할 수 있으며, 불가하면 `moai session current --show-fallback`을 사용한다. 새 세션의 검증은 반드시 실제 소스/바이너리 버전에 귀속시켜라.
