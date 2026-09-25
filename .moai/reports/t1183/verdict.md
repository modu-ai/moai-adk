# t1183 — Claude Code 2.1.277–2.1.282 규칙 문서 갱신

## Claim

규칙 문서 3종과 배포 템플릿 사본에 공식 릴리스 노트의 `rm` 보호, 훅 제약, 신규 환경변수와 `maxProseWidth`, 프로젝트 설정의 OTel 제한을 반영했다. t1184가 별도로 수정 중인 `effortLevel` 문장은 건드리지 않았다.

## Evidence

- `claude --version` → `2.1.282 (Claude Code)`.
- 공식 근거: [v2.1.277](https://github.com/anthropics/claude-code/releases/tag/v2.1.277) (`${VAR:?}` 안내), [v2.1.280](https://github.com/anthropics/claude-code/releases/tag/v2.1.280) (PermissionRequest agent 훅, MCP 설명 길이), [v2.1.281](https://github.com/anthropics/claude-code/releases/tag/v2.1.281) (dangerous-rm 대기·치환 대상 프롬프트·mcp_tool 연결 대기·auto-mode 변수), [v2.1.282](https://github.com/anthropics/claude-code/releases/tag/v2.1.282) (`maxProseWidth`, 텔레메트리 진단), [Monitoring](https://code.claude.com/docs/en/monitoring-usage) (프로젝트·로컬 OTel 설정 제한).
- `make build > /tmp/t1183-make-build.log 2>&1` → exit 0; 출력: `catalog.yaml updated successfully (13408 bytes)`.
- `go test ./internal/template -run '^Test(RuleTemplateMirrorDrift|MCPConfigurationDoctrine|SettingsTemplateValidJSON)$' -count=1 -timeout=90s` → exit 0; 패키지 결과 `ok` (`internal/template`).
- `git diff --check` → exit 0.

## Baseline-attribution

전용 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1183`, 브랜치 `WT-claude-277-282-rules`, 수정 전 HEAD `8a0ce2d14`. 공식 문서는 이번 작업 중 조회했다.

## Gaps

- 실제 위험한 `rm` 명령이나 승인 대기를 실행하지 않았다. 동작 설명은 공식 릴리스 노트에 근거한다.
- OTel 설정 경고는 이 카드에서 새로 실세션 관측하지 않았다. t1184의 별도 측정을 이 카드의 측정으로 주장하지 않는다.
- Windows 및 Bedrock·Vertex·Foundry 환경에서 직접 확인하지 않았다.

## Residual-risk

Claude Code 버전·호스트별 동작은 달라질 수 있다. 이 문서는 2.1.282 설치 버전과 위 공식 자료를 기준으로 한다. t1184의 `effortLevel` PR이 병합되면 같은 문서의 변경을 충돌 없이 통합해야 한다.
