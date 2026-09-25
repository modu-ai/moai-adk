# t1172 M3 독립 감사

SPEC: `SPEC-CODEX-PREAPPROVAL-PROBE-001`  
대상: `8ab0fa13b`의 M3 변경  
판정: **PASS (99/100)** — M3 범위에 한함

## 차원별 판정

| 차원 | 점수 | 판정 | 근거 |
|---|---:|---|---|
| Functionality (40%) | 100 | PASS | 고정 문구 1회, 기준 행과 바이트 동일, 관련 테스트 통과 |
| Security (25%) | 100 | PASS | 승인 거부 시 감사 생략·`spawn_agent` 대체를 금지하는 문구 확인; 변경 파일은 템플릿과 테스트뿐 |
| Craft (20%) | 95 | PASS | `agentemit` 전체 패키지 커버리지 88.5%, `gofmt`와 diff 검사 통과 |
| Consistency (15%) | 100 | PASS | 템플릿 중립성·Codex 예산·방출 검사 통과 |

## Claim

M3는 `AGENTS.md.tmpl`의 `audit-verdict-file` 행 끝에 REQ-CPP-008의 고정 문장 하나만 추가했다. 기준 커밋 `0356e8117`에 비해 그 밖의 템플릿 바이트 변화는 없다. `internal/template/agentemit/audit_launcher_surface_test.go:131-136`은 문장 위치와 1회 출현을 검사한다. `internal/template/templates/AGENTS.md.tmpl:33`은 승인 거부를 blocker로 돌리도록 지시한다. 이 판정은 M3의 정적 지시·생성 검사에 한정한다.

## Evidence

측정 명령과 출력(모두 아래 Baseline-attribution의 트리에서 실행):

```text
$ python3 -c '<git show 0356e8117의 audit-verdict-file 행에 고정 문장을 붙여 현재 템플릿과 바이트 비교>'
base_rows 1 current_rows 1 fixed_count 1 skip_count 1
exact_row True
only_template_delta True

$ python3 -c '<템플릿 diff의 추가 줄에 SPEC/카드/날짜/로컬 경로/긴 16진 문자열 검사>'
added_lines 1 neutrality_bad 0

$ go test -cover ./internal/template/agentemit -count=1
ok  github.com/modu-ai/moai-adk/internal/template/agentemit  0.197s  coverage: 88.5% of statements

$ go test -cover ./internal/template/agentemit/... -run 'TestAuditRoleLauncherInstructionSurface|TestCodexAuditRolesReadOnlyScopedException' -count=1
ok  github.com/modu-ai/moai-adk/internal/template/agentemit  0.321s  coverage: 74.1% of statements

$ go test ./internal/template -run 'TestTemplateNeutralityAudit$|TestTemplateNoInternalContentLeak$|TestDeployerSingleRender$|TestTemplateNeutralityAuditC8Preserve$' -count=1
ok  github.com/modu-ai/moai-adk/internal/template  0.934s

$ go test -v ./internal/template -run 'TestCodexOnlyForceUpdateVariant$|TestTemplateNeutralityAudit$|TestTemplateNoInternalContentLeak$' -count=1
--- PASS: TestCodexOnlyForceUpdateVariant (0.16s)
--- PASS: TestTemplateNoInternalContentLeak (0.57s)
--- PASS: TestTemplateNeutralityAudit (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/template  1.015s

$ go test -v ./internal/config -run 'TestCodexContractByteCeiling$|TestCodexNestedTemplateDiscoveryBudget$' -count=1
contract document AGENTS.md = 16441 bytes (ceiling 24576, headroom 8135)
contract document internal/template/templates/AGENTS.md.tmpl = 19380 bytes (ceiling 24576, headroom 5196)
--- PASS: TestCodexContractByteCeiling (0.01s)
--- PASS: TestCodexNestedTemplateDiscoveryBudget (0.07s)
PASS
ok  github.com/modu-ai/moai-adk/internal/config  0.185s

$ make agents-emit-check
ok  github.com/modu-ai/moai-adk/internal/template/agentemit  0.138s

$ gofmt -l internal/template/agentemit/audit_launcher_surface_test.go
<출력 없음>
$ git diff --check 0356e8117..HEAD -- internal/template/templates/AGENTS.md.tmpl internal/template/agentemit/audit_launcher_surface_test.go
<출력 없음>
```

`go test -v` 중립성 검사에는 다른 템플릿의 기존 `TEMPLATE_NEUTRALITY_WARN` 권고 로그가 있었으나 테스트는 PASS였다. M3가 더한 행은 경고 대상으로 출력되지 않았다.

## Baseline-attribution

작업 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1172`, HEAD `8ab0fa13b`; 비교 기준은 SPEC AC-CPP-010에 고정된 `0356e8117`. 변경 범위는 `git show --stat --oneline 8ab0fa13b` 출력상 `internal/template/agentemit/audit_launcher_surface_test.go` 7줄 추가와 `internal/template/templates/AGENTS.md.tmpl` 한 행 변경이었다. 이 감사에서 실행한 결과만 위에 기록했다.

## Findings

발견된 결함 없음. M3 요구와 관련된 차단 이슈 없음.

## Gaps

- AC-CPP-010의 원문 셸 명령은 기존 증거 파일 두 개를 덮어쓰므로 실행하지 않았다. 같은 기준 행을 `git show`로 읽고 메모리에서 바이트 비교하는 읽기 전용 판정으로 핵심 조건을 확인했다.
- 실제 설치 프로젝트에 방출한 `AGENTS.md`의 행 바이트를 별도로 수집하지 않았다. `TestCodexOnlyForceUpdateVariant`는 `AGENTS.md` 파일 생성을 확인하지만 그 행의 내용까지 비교하지 않는다. 템플릿의 고정 행에는 동적 치환이 없고 정적 행은 직접 비교했다.
- Codex 모델 호출이나 MCP 승인 거부 상황은 실행하지 않았다. M3의 정적 지시 검증이며 LIVE 판별과 운영 채택 판정은 다른 마일스톤의 범위다.

## Residual-risk

향후 렌더러가 Markdown 본문을 변형한다면 현재 배포 테스트는 고정 문장의 바이트 보존을 직접 잡지 못할 수 있다. 현재 M3 변경에는 렌더러 수정이 없고, 실제 템플릿 바이트 비교와 생성 파일 존재 검사는 통과했다.

## 반복 이력

이번 M3 독립 감사 1회. 이전 M1-a/M1-b 감사와 판정 범위가 다르다.
