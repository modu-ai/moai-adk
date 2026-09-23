---
id: SPEC-DUAL-HARNESS-RECOVERY-001
document: acceptance
created: 2026-09-23
updated: 2026-09-23
author: manager-spec
card: t1100
---

# Acceptance — SPEC-DUAL-HARNESS-RECOVERY-001

## §A 판정 규칙

- 모든 AC는 아래 명령의 마지막 출력이 정확히 `true`일 때만 PASS다. 그 밖의 출력, 명령 실패, 증거 파일 부재는 FAIL이다.
- `SKIP`과 `NOT_RUN`은 PASS가 아니다. SKIP이 나온 AC는 "증명되지 않음"이며, 판정표에는 `UNPROVEN`으로 적는다. 패키지 단위 `ok` 줄은 증거로 쓰지 않는다(REQ-DHR-024). 실제로 이번 plan 단계에서 `TestFactoryLive*` 4개가 모두 `SKIP`인 채 `ok ... internal/cli`가 출력되는 것을 관측했다(`research.md` §D).
- 테스트 이름은 run 단계에서 만들 이름이다. 이름을 바꾸면 이 파일의 명령도 함께 고친다. 이름이 없으면 pass 수가 0이 되어 FAIL로 판정된다.
- 증거는 `.moai/reports/t1100/`에 남긴다.
- `internal/cli` 명령은 kanban 환경 변수를 같은 호출 안에서 지운다(`unset ... && go test ...`).
- LIVE AC(AC-DHR-012, AC-DHR-018)는 결정적 AC와 따로 집계한다. LIVE 미실행은 `NOT_RUN`으로 적고 해당 설계 기준을 PASS로 올리지 않는다.

공통 판정식(각 명령에 그대로 들어 있다): 지정 이름의 `pass` 이벤트 수가 기대값과 같고, `fail`·`skip` 이벤트가 0이며, 출력에 `NOT_RUN`이 없다.

### AC ↔ 요구사항 매핑

| AC | Requirement | 설계 기준 |
|---|---|---|
| AC-DHR-001 | REQ-DHR-002, REQ-DHR-003 | AC-MIG-01 |
| AC-DHR-002 | REQ-DHR-004 | AC-MIG-01 |
| AC-DHR-003 | REQ-DHR-001, REQ-DHR-005 | AC-MIG-01 |
| AC-DHR-004 | REQ-DHR-006, REQ-DHR-001 | AC-MIG-01 |
| AC-DHR-005 | REQ-DHR-007 | AC-MIG-01 |
| AC-DHR-006 | REQ-DHR-008, REQ-DHR-011 | AC-WT-01 |
| AC-DHR-007 | REQ-DHR-009 | AC-WT-01 |
| AC-DHR-008 | REQ-DHR-010 | AC-WT-01 |
| AC-DHR-009 | REQ-DHR-012 | AC-WT-01 |
| AC-DHR-010 | REQ-DHR-013 | AC-AGENT-01 |
| AC-DHR-011 | REQ-DHR-014 | AC-AGENT-01 |
| AC-DHR-012 | REQ-DHR-014 | AC-AGENT-01 (LIVE) |
| AC-DHR-013 | REQ-DHR-015 | AC-AGENT-01 |
| AC-DHR-014 | REQ-DHR-016, REQ-DHR-017, REQ-DHR-018, REQ-DHR-019 | AC-MSG-01 |
| AC-DHR-015 | REQ-DHR-018, REQ-DHR-020 | AC-MSG-01 |
| AC-DHR-016 | REQ-DHR-021 | AC-MSG-01 |
| AC-DHR-017 | REQ-DHR-022 | AC-FACT-01 |
| AC-DHR-018 | REQ-DHR-023 | AC-FACT-01 (LIVE) |
| AC-DHR-019 | REQ-DHR-024 | AC-FACT-01 |

## §B 인수 기준

### AC-DHR-001 — 저널 기반 쓰기, 대조, 동시 수정 거부 (REQ-DHR-002, REQ-DHR-003)

**Given** 사용자 handler가 든 `.codex/hooks.json`과 사용자 테이블이 든 `.codex/config.toml`이 있는 임시 프로젝트,
**When** 배선을 실행하되 (a) 방해 없이, (b) 해시 재확인 직전에 테스트 seam으로 대상 파일을 바꿔서 실행하면,
**Then** (a)는 저널 항목이 `complete`이고 기록 후 해시가 의도한 해시와 같으며, (b)는 대상이 바이트 그대로이고 저널 항목이 `conflict`이며 두 해시가 보고된다.

음성·변이: rename 직전 재확인을 지운 변이, 기록 후 대조를 지운 변이는 각각 이 테스트에서 실패해야 한다(테스트 안의 변이 표로 확인).

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/codexwiring -run '^(TestCodexWiringJournaledWriteReadback|TestCodexWiringRefusesConcurrentModification)$' -count=1 -timeout=120s > .moai/reports/t1100/ac001.jsonl; jq -se '([.[]|select(.Action=="pass" and ((.Test//"")|test("^(TestCodexWiringJournaledWriteReadback|TestCodexWiringRefusesConcurrentModification)$")))]|length)==2 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac001.jsonl
```

### AC-DHR-002 — 중단 설치 복구 (REQ-DHR-004)

**Given** failpoint 네 곳(임시 파일 작성 후, 저널 추가 후, rename 후 대조 전, 대조 후 완료 표시 전)에서 멈춘 배선 실행 네 건,
**When** 각 건에 대해 복구를 실행하면,
**Then** 대상 해시에 따라 `completed` / `not-applied` / `diverged`로 분류되고, `diverged` 대상은 바이트 그대로 남으며, 어떤 경우에도 임시 파일이 대상 자리에 남지 않는다.

음성: 복구 전에 대상을 사용자가 바꾼 경우(`diverged`)에 복구가 덮어쓰면 FAIL.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/codexwiring -run '^TestCodexWiringInterruptedInstallRecovery$' -count=1 -timeout=120s > .moai/reports/t1100/ac002.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexWiringInterruptedInstallRecovery")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac002.jsonl
```

### AC-DHR-003 — MoAI 소유 부분만 제거 (REQ-DHR-001, REQ-DHR-005)

**Given** MoAI가 새로 만든 `.codex/hooks.json`(MoAI handler만), 사용자 handler와 MoAI handler가 섞인 `hooks.json`, 사용자 테이블 + MoAI가 추가한 `[mcp_servers.moai]`·`[tui].status_line`이 든 `config.toml`, 설치 전부터 `[mcp_servers.moai]`가 있던 `config.toml`,
**When** unwire를 실행하면,
**Then** MoAI만 만든 파일은 삭제되고, 섞인 파일은 MoAI 부분만 빠진 채 나머지 바이트가 그대로이며, 설치 전부터 있던 `[mcp_servers.moai]`는 남고 `user-owned`로 보고된다.

변이: 사전 존재 테이블을 MoAI 소유로 기록하는 변이는 FAIL해야 한다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/codexwiring -run '^TestCodexUnwireRemovesOnlyOwnedParts$' -count=1 -timeout=120s > .moai/reports/t1100/ac003.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexUnwireRemovesOnlyOwnedParts")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac003.jsonl
```

### AC-DHR-004 — 해시만으로 지우지 않음 (REQ-DHR-006, REQ-DHR-001)

**Given** provenance 기록이 없지만 내용 해시가 정본과 같은 파일, 기록은 MoAI 소유지만 사용자가 수정한 파일, 프로젝트 밖을 가리키는 symlink인 `.codex/hooks.json`,
**When** unwire를 실행하면,
**Then** 세 대상 모두 바이트 그대로(symlink 대상 포함) 남고, 각각 `no-provenance`, `modified`, `symlink-boundary`로 보고된다.

변이: provenance 확인을 지우고 해시만 비교하는 변이는 첫 대상을 지우므로 FAIL해야 한다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/codexwiring -run '^TestCodexUnwireRefusesUnownedModifiedSymlink$' -count=1 -timeout=120s > .moai/reports/t1100/ac004.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexUnwireRefusesUnownedModifiedSymlink")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac004.jsonl
```

### AC-DHR-005 — 하네스 프로필 전환 행렬 (REQ-DHR-007)

**Given** 전환 여섯 행(`claude → both`, `gpt → both`, `both → claude`, `both → gpt`, `gpt → claude`, 구버전 배포 → 신버전)마다 사용자 수정 파일 1개, 손상된 `hooks.json` 1개, 수정되지 않은 템플릿 관리 파일이 있는 임시 프로젝트,
**When** 각 전환을 적용하면,
**Then** 모든 행에서 사용자 소유 파일·부분이 바이트 그대로이고, 대상 프로필이 더 이상 배포하지 않는 템플릿 관리 파일은 pristine일 때만 백업 후 제거되며, 수정·미확인 파일과 고아 배선 파일은 남은 채 보고된다. 손상된 `hooks.json`은 건드리지 않는다.

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestHarnessProfileTransitionPreservesUserData$' -count=1 -timeout=600s > .moai/reports/t1100/ac005.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestHarnessProfileTransitionPreservesUserData")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac005.jsonl
```

### AC-DHR-006 — Codex 트리 anchor와 생성 base (REQ-DHR-008, REQ-DHR-011)

**Given** 임시 git 저장소와 설정된 base 브랜치,
**When** `moai codex -w new-tree`(새 트리)와 `moai codex -w existing-tree`(기존 트리)를 launch seam으로 실행하면,
**Then** 두 트리 모두 codex 프로세스가 될 pid를 담은 git worktree lock을 갖고, 기존 anchor 판정이 두 트리를 anchored로 본다. 새 트리의 HEAD는 해석된 base 커밋과 같다. base 해석 결과와 생성된 HEAD가 다르도록 seam을 바꾸면 launch가 거부되고 트리는 삭제되지 않는다. 죽은 pid의 lock은 교체되고, 살아 있는 pid의 lock은 교체되지 않는다.

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestCodexWorktreeAnchorLockAndBase$' -count=1 -timeout=600s > .moai/reports/t1100/ac006.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexWorktreeAnchorLockAndBase")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac006.jsonl
```

### AC-DHR-007 — 동시 writer 거부 (REQ-DHR-009)

**Given** 살아 있는 다른 프로세스 pid로 lock된 트리, 판독 불가 lock 사유를 가진 트리, 레지스트리에만 살아 있는 세션이 기록된 트리,
**When** 각 트리에 `moai codex -w`와 `moai cc -w`를 실행하면,
**Then** 여섯 호출이 모두 0이 아닌 종료 코드와 anchor 출처·보유자를 적은 진단을 내고, 트리의 lock·브랜치·작업 파일은 바뀌지 않는다.

음성: 죽은 pid의 lock만 있는 트리에는 launch가 허용되어야 한다(과잉 거부 방지).

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestWorktreeLaunchRejectsConcurrentWriter$' -count=1 -timeout=600s > .moai/reports/t1100/ac007.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestWorktreeLaunchRejectsConcurrentWriter")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac007.jsonl
```

### AC-DHR-008 — 미통합 트리 삭제 거부 (REQ-DHR-010)

**Given** `moai codex -w`로 만든 트리 세 개(base에 없는 커밋 보유, 미커밋 변경 보유, 살아 있는 codex anchor 보유)와 병합이 끝난 트리 하나,
**When** `moai worktree clean --stale --apply`, `moai worktree done`, 세션 종료 정리를 각각 적용하면,
**Then** 앞의 세 트리는 세 경로 모두에서 남고 이유가 보고되며, 병합이 끝나고 anchor가 없는 트리만 제거 대상이 된다.

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli/worktree ./internal/cli -run '^TestWorktreeDisposalRefusesUnintegratedCodexTree$' -count=1 -timeout=600s > .moai/reports/t1100/ac008.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestWorktreeDisposalRefusesUnintegratedCodexTree")]|length)>=1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac008.jsonl
```

### AC-DHR-009 — Codex kanban 진입 동등성 (REQ-DHR-012)

**Given** `moai cc -k`의 진입 형태 표(`plan.md`의 결정 [NEEDS CLARIFICATION: codex-kanban-roles]가 정한 역할 범위),
**When** 같은 인자로 `moai codex`와 `moai cc`를 launch seam으로 실행하면,
**Then** 두 경로가 같은 kanban launch facts(backend만 `codex`/`claude`로 다름)와 같은 세션 이름 규칙을 내고, codex 자식 인자에 `-k`/`--kanban`/역할 토큰이 없다. 지원하지 않는 형태는 사용법 진단과 0이 아닌 종료로 끝나며 일반 launch로 떨어지지 않는다.

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestCodexKanbanEntryParity$' -count=1 -timeout=600s > .moai/reports/t1100/ac009.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexKanbanEntryParity")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac009.jsonl
```

### AC-DHR-010 — 역할 권한 계약과 조용한 누락 금지 (REQ-DHR-013)

**Given** 12개 Codex 역할의 권한 계약,
**When** agent 생성기를 실행하면,
**Then** 모든 역할의 모든 축(쓰기, shell, MCP, 하위 에이전트, 웹)이 `enforced` 또는 `UNSUPPORTED` 중 하나로 매핑되고, 생성된 `sandbox_mode`는 계약이 요구하는 값보다 넓지 않다.

변이: 한 역할의 한 축에서 매핑을 지운 변이, `read-only` 역할을 `workspace-write`로 바꾼 변이는 생성기가 거부해야 한다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/template/agentemit -run '^TestRolePermissionContractNoSilentDrop$' -count=1 -timeout=120s > .moai/reports/t1100/ac010.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestRolePermissionContractNoSilentDrop")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac010.jsonl
```

### AC-DHR-011 — UNSUPPORTED는 PASS가 아님 (REQ-DHR-014)

**Given** 계약에서 `UNSUPPORTED`로 선언된 축 목록(예상 목록은 `design.md` §C),
**When** 역할 검증 보고를 만들면,
**Then** 보고의 `UNSUPPORTED` 집합이 예상 목록과 정확히 같고, 그 축 어느 것도 PASS로 집계되지 않는다.

변이: `UNSUPPORTED`를 PASS로 세는 변이는 FAIL해야 한다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/template/agentemit -run '^TestRolePermissionUnsupportedNeverPass$' -count=1 -timeout=120s > .moai/reports/t1100/ac011.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestRolePermissionUnsupportedNeverPass")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac011.jsonl
```

### AC-DHR-012 — [LIVE] 12개 역할 실제 로드와 read-only 강제 (REQ-DHR-014)

**Given** 격리된 임시 저장소와 임시 `CODEX_HOME`, 설치된 codex 바이너리, 호출 예산(`plan.md` 결정 [NEEDS CLARIFICATION: live-budget]),
**When** 12개 역할 TOML을 로드시키고 `read-only` 역할 두 개에 파일 쓰기를 시도시키면,
**Then** codex가 12개 역할을 모두 오류 없이 로드하고(`Ignoring malformed agent role definition` 없음), `read-only` 역할의 쓰기는 파일을 만들지 못하며, 증거에 codex 버전이 적힌다. MCP 경유 부작용은 이 AC가 증명하지 않으며 `UNSUPPORTED`로 남는다.

SKIP 의미: `MOAI_CODEX_ROLE_LIVE=1`이 없으면 이 테스트는 SKIP하고, 그 경우 AC-DHR-012는 `NOT_RUN`이다. AC-AGENT-01의 런타임 부분은 증명되지 않은 것으로 보고한다. CI에서 이 게이트를 켜는 워크플로는 이번 plan 시점에 관측되지 않았다.

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && MOAI_CODEX_ROLE_LIVE=1 go test -json ./internal/cli -run '^TestCodexRoleLiveLoadAndReadOnly$' -count=1 -timeout=900s > .moai/reports/t1100/ac012-live.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexRoleLiveLoadAndReadOnly")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0 and ([.[]|select((.Output//"")|test("CODEX_VERSION=[0-9]"))]|length)>=1' .moai/reports/t1100/ac012-live.jsonl
```

### AC-DHR-013 — 감사 역할의 쓰기 범위 경계 (REQ-DHR-015)

**Given** snapshot을 잡은 임시 저장소와 허용 목록(감사 보고·판정 경로),
**When** (a) 허용 경로만 바꾼 뒤 검증하고, (b) 허용 목록 밖 소스 파일 하나를 바꾼 뒤 검증하면,
**Then** (a)는 종료 코드 0, (b)는 0이 아닌 종료 코드와 바뀐 경로 목록을 낸다. (b)에서 감사 판정은 기각 상태로 기록된다.

변이: 허용 목록 비교를 지운 변이는 (b)를 통과시키므로 FAIL해야 한다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/cli/worktree -run '^TestAuditorWriteScopeBoundary$' -count=1 -timeout=300s > .moai/reports/t1100/ac013.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestAuditorWriteScopeBoundary")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac013.jsonl
```

### AC-DHR-014 — 중복·ack 유실·재시작에도 결과 한 번 (REQ-DHR-016 ~ REQ-DHR-019)

**Given** `started` 상태의 dispatch 하나,
**When** 같은 결과가 (a) 두 번 전달되고, (b) 결과 기록 후 receipt 전에 lease가 만료되어 재전달되고, (c) 송신 worker가 새 세션 UUID로 재시작해 같은 멱등 키로 다시 보내고, (d) 브로커 프로세스가 결과 기록 직후 재시작하면,
**Then** 네 경우 모두 dispatch는 `result_recorded`로 한 번만 바뀌고 result 참조는 처음 것과 같으며, 두 번째 적용은 `duplicate`를 반환한다. 같은 범위에 다른 본문을 보내면 거부되고 저장 상태는 바뀌지 않는다.

변이: 멱등 범위를 송신자 세션 UUID로 되돌린 변이는 (c)에서 두 번 적용하므로 FAIL해야 한다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/factorymsg -run '^TestDispatchResultExactlyOnce$' -count=1 -timeout=120s > .moai/reports/t1100/ac014.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestDispatchResultExactlyOnce")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac014.jsonl
```

### AC-DHR-015 — 낡은 fencing token과 superseded 메시지 (REQ-DHR-018, REQ-DHR-020)

**Given** generation 1로 할당된 dispatch와, 수신 lane이 generation 2로 재등록된 뒤에도 generation 1 앞으로 남은 메시지,
**When** generation 1 주체가 결과를 적용하고 메시지를 claim·read·dispose·ack하려 하면,
**Then** 모든 시도가 stale 결과로 거부되고 dispatch·메시지 행은 바뀌지 않으며, 브로커 status는 그 메시지를 `pending`이 아니라 `superseded`로 센다. generation 2의 정상 작업은 성공한다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/factorymsg -run '^TestDispatchStaleFencingAndSuperseded$' -count=1 -timeout=120s > .moai/reports/t1100/ac015.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestDispatchStaleFencingAndSuperseded")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac015.jsonl
```

### AC-DHR-016 — 재할당과 늦은 응답 (REQ-DHR-021)

**Given** attempt 1이 `started`인 dispatch,
**When** (a) attempt 1 소유자가 살아 있는 상태에서 재할당하고, (b) 소유자 종료가 확인된 뒤 재할당하고, (c) attempt 2가 결과를 기록한 뒤 attempt 1의 늦은 결과가 도착하면,
**Then** (a)는 명시적 철회 없이 거부되고, (b)는 attempt 2와 새 assignee generation을 기록하며, (c)의 늦은 결과는 stale로 거부되어 적용된 결과는 attempt 2의 것 하나뿐이다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/factorymsg -run '^TestDispatchReassignmentFencesLateResult$' -count=1 -timeout=120s > .moai/reports/t1100/ac016.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestDispatchReassignmentFencesLateResult")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac016.jsonl
```

### AC-DHR-017 — 결정적 4조합 카드 흐름 (REQ-DHR-022)

**Given** backend 표지가 붙은 lead·worker peer 네 쌍(Claude→Claude, Codex→Codex, Claude→Codex, Codex→Claude)과 실제 브로커 저장소,
**When** 각 쌍이 dispatch를 `assigned → delivered → started`로 진행한 뒤 worker 중단, 새 worker로 재할당, 새 attempt의 결과 기록, 옛 attempt의 늦은 결과, lead의 `integrated` 기록을 거치면,
**Then** 네 하위 테스트 모두 dispatch가 `integrated`로 끝나고 적용된 결과는 하나이며, 메시지 도착만으로는 `delivered`를 넘지 않는다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/factorymsg -run '^TestFactoryCardFlowFourCombinations$' -count=1 -timeout=180s > .moai/reports/t1100/ac017.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryCardFlowFourCombinations")]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestFactoryCardFlowFourCombinations/(claude-claude|codex-codex|claude-codex|codex-claude)$")))]|length)==4 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac017.jsonl
```

### AC-DHR-018 — [LIVE] 실제 CLI 4조합 카드 흐름 (REQ-DHR-023)

**Given** 격리된 임시 저장소·`MOAI_HOME`·`CODEX_HOME`, 조합별 모델 호출 예산과 시간 제한, 첫 프로세스 생성 전에 등록한 정리 함수,
**When** 네 조합을 한 번에 하나씩 실행하면(기존 `requireFactoryLive`는 `MOAI_FACTORY_LIVE_CASE`가 다르면 `Fatal`하므로 조합마다 별도 호출),
**Then** 각 조합이 AC-DHR-017의 흐름을 실제 별도 CLI·모델 문맥에서 마치고, 조합별 증거 파일에 nonce, 적용된 결과 1건, 거부된 늦은 결과, 정리된 프로세스 목록이 남는다.

SKIP 의미: `MOAI_FACTORY_LIVE=1`이 없으면 네 테스트는 SKIP하며 AC-DHR-018은 `NOT_RUN`, AC-FACT-01은 증명되지 않음이다. 이번 plan 시점에 `.github/workflows/`에서 `MOAI_FACTORY_LIVE`를 쓰는 워크플로는 관측되지 않았으므로 CI 실행을 주장하지 않는다.

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && for c in claude-claude codex-codex claude-codex codex-claude; do n=$(printf '%s' "$c" | awk -F- '{print toupper(substr($1,1,1)) substr($1,2) toupper(substr($2,1,1)) substr($2,2)}'); MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE="cardflow-$c" go test -json ./internal/cli -run "^TestFactoryLiveCardFlow${n}\$" -count=1 -timeout=900s > ".moai/reports/t1100/ac018-live-$c.jsonl"; done; jq -se '([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestFactoryLiveCardFlow(ClaudeClaude|CodexCodex|ClaudeCodex|CodexClaude)$")))]|length)==4 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1100/ac018-live-*.jsonl
```

### AC-DHR-019 — 판정식이 SKIP·NOT_RUN을 거부함 (REQ-DHR-024)

**Given** 이 파일이 쓰는 공통 판정식,
**When** 다섯 가지 합성 `go test -json` 흐름(정상 pass, 전부 skip, 이름 불일치, fail 포함, 출력에 `NOT_RUN`)을 넣으면,
**Then** 정상 흐름만 `true`이고 나머지 넷은 `false`다.

```bash
P='([.[]|select(.Action=="pass" and .Test=="TestX")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0'; a=$(printf '{"Action":"pass","Test":"TestX"}\n{"Action":"pass"}\n' | jq -se "$P"); b=$(printf '{"Action":"skip","Test":"TestX"}\n{"Action":"pass"}\n' | jq -se "$P"); c=$(printf '{"Action":"pass","Test":"TestY"}\n' | jq -se "$P"); d=$(printf '{"Action":"pass","Test":"TestX"}\n{"Action":"fail","Test":"TestX/sub"}\n' | jq -se "$P"); e=$(printf '{"Action":"output","Test":"TestX","Output":"NOT_RUN\\n"}\n{"Action":"pass","Test":"TestX"}\n' | jq -se "$P"); [ "$a$b$c$d$e" = "truefalsefalsefalsefalse" ] && echo true || echo false
```

## §C 설계 기준별 판정 집계

| 설계 기준 | 결정적 AC | LIVE AC | PASS 조건 |
|---|---|---|---|
| AC-MIG-01 | AC-DHR-001 ~ 005 | 없음 | 다섯 AC 모두 `true` |
| AC-WT-01 | AC-DHR-006 ~ 009 | 없음 | 네 AC 모두 `true` |
| AC-AGENT-01 | AC-DHR-010, 011, 013 | AC-DHR-012 | 결정적 셋 `true` + LIVE `true`. LIVE `NOT_RUN`이면 `PARTIAL`이며 PASS 아님. `UNSUPPORTED` 축은 PASS 집계에서 뺀 채 따로 나열 |
| AC-MSG-01 | AC-DHR-014 ~ 016 | 없음 | 세 AC 모두 `true` |
| AC-FACT-01 | AC-DHR-017, 019 | AC-DHR-018 | 결정적 둘 `true` + LIVE `true`. LIVE `NOT_RUN`이면 `PARTIAL` |

## §D 품질 게이트와 완료 정의

- 변경 패키지 단위 테스트만 로컬에서 실행한다(`internal/codexwiring`, `internal/factorymsg`, `internal/template/agentemit`, `internal/cli/worktree`, `internal/cli`의 지정 이름). 전체 스위트 판정은 CI 몫이다.
- `go vet`과 `golangci-lint run`을 변경 패키지에 대해 실행해 0건이어야 한다.
- `internal/template/templates/.claude/agents/moai/*.md`를 고치면 `make agents-emit`을 실행하고, `.codex/agents/moai/*.toml`을 손으로 고치지 않는다.
- factorymsg 스키마 변경은 기존 DB 파일을 여는 마이그레이션 테스트를 포함한다(기존 행 보존, `SchemaVersion` 증가).
- 완료 정의: §C 표의 다섯 기준 판정과 그 근거 파일 경로가 `progress.md` §E.2에 기록되고, LIVE 항목은 실행했으면 증거 경로, 안 했으면 `NOT_RUN`으로 적힌다. `PARTIAL`을 PASS로 적지 않는다.
