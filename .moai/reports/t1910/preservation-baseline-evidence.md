# t1910 보존 측정 evidence 색인 (manager-spec, 2026-10-20)

측정 목적: t1909(제거) 이전 시점에 codex_role_audit 의존 경로가 살아 있음을 기계적으로 기록. 재실행 시 이 파일과 같은 형식으로 receipt 비교.

## 실행 receipt

| 항목 | 값 |
|---|---|
| 실행 경로 | MCP `codex_role_audit` (role=super-advisor, route=mcp) |
| worktree_root | /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1910 |
| job id | `job-20261020T091536Z-997d55c4` — invocation-transient (launch record에 job-id 필드가 없고 제3의 디스크 레코드도 없어 이 표가 유일한 기록이다; 향후 launcher가 launch record에 job id를 기록하면 해소) [t1910-D1 주기 2026-10-21] |
| launch record | `.moai/reports/codex-audit/super-advisor-20261020T091538Z-35f8900d.launch.json` |
| sandbox / mcp_servers | read-only / disabled (record `argv`: `exec -s read-only -c approval_policy="never" ... -C <t1910>`) |
| 시작/종료 | 2026-10-20T09:15:38Z / 2026-10-20T09:15:53Z (15초) |
| exit_code | 0 |
| verdict 파일 | `.moai/reports/t1910/preservation-baseline.md` |
| record verdict_sha256 | `08e6580ca95e0e434ecec476c5b355f6c63944143082d7ba9f2fbada4991ace8` |
| shasum -a 256 재계산 | `08e6580ca95e0e434ecec476c5b355f6c63944143082d7ba9f2fbada4991ace8` — 일치 |

## verdict 원문 (2행 전체)

```
go version go1.26.8 darwin/arm64
// Package decision implements the decision board: one append-only,
```

## 교차 검증 (별도 관측, 같은 트리)

- `sed -n '1p' internal/decision/board.go` → `// Package decision implements the decision board: one append-only,` — 역할 응답과 일치
- `go version` → `go version go1.26.8 darwin/arm64` — 일치

## 도구 출처 (verification-claim-integrity §2.2 이중 좌표)

- 실행 주체: 상주 MCP 서버 pid 15402, build `moai_cp/20261007_195927` commit `7705b365e2` (2026-10-18T10:45:08Z 빌드)
- 측정 대상 트리: HEAD `8249bff8c3` — `git merge-base --is-ancestor 7705b365e2 HEAD` → 조상 맞음(지연 존재, 아래에 명시)
- 지연 실횡성: `git diff --stat 7705b365e2..HEAD -- internal/cli/codex_audit_launch.go` = env helper 이름 변경 1건(`codexChildEnv` → `codexAuditChildEnv`, :618/:705)뿐 — 인용한 역할 적재 구간(:497-542)은 build과 HEAD가 동일
- 설치 CLI 바이너리: `moai version` → `v3.2.0-rc.37  moai_cp/20261007_195927-2046-g8249bff8c3  built 2026-10-19T12:42:24Z` — 트리 HEAD 빌드

## read base (같은 턴 실측)

- `git rev-parse main` = `8249bff8c3933d647b4f99e59048595f9808e83a` (트리 HEAD = local main 선단)
- `git rev-parse origin/main` = `6dde1c1c1bdc84bf82c1b8266c071c44756fe7c0`
- `git rev-list --count --left-right origin/main...main` → `0	559`
- batch branch: `16d03deea3` = "fix(cli): repair ST1008 and QF1001 flagged on the #1830 batch head (#1855)", local main 미병합; `git rev-list --count main..16d03deea3` = 5, 네 감사 파일에 대한 `git diff --stat main...16d03deea3` = 공집합

## Gaps

- `codex_role_audit_status` / `_result` MCP 도구는 호출하지 않았다 — 결과는 launcher가 기록한 verdict 파일+launch record로 판독(실행 산출물 자체가 증거).
- go test / go build / golangci-lint 미실행 — 파견 지시가 조사 전용(감사 1회 실행만 실행 항목)이므로.
- 미사용 workspace-write 역할 TOML 8개의 테스트면(bundle/golden 12건 고정 여부)은 측정하지 않았다 — 보강 문서에서 위험 표기만.
- kill 규율: 패턴 kill(pkill/killall/kill $(pgrep)) 미사용 — 프로세스 종료 자체가 없었다. 추적 작업은 TaskStop, 가드 거부는 재시도 없음(이번 턴 가드 거부 1건: 명령 치환 sed 범위 형식 → 분할 명령으로 우회 아닌 분리 실행).

## Residual-risk [t1910-D4 보강 2026-10-21]

- build-vs-tree 지연: 실행 주체는 상주 MCP 서버 build `7705b365e2`, 측정 대상은 트리 HEAD `8249bff8c3` — 인용 구간은 지연 실효성 실측으로 동일 확인됐으나 샌드박스 실행 스트림 자체는 재관측 불가하다(경계 문서 L130 동일 측정).
- 행 번호 핀 지평: 이 색인의 인용은 트리 `8249bff8c3` 고정 — local main이 이후 흡수돼 감사 파일에 변화가 생기면 재판독이 필요하다. 상세: `adk-codex-cleanup-20261020/boundary-survey-reinforcement-t1910.md` L130.
