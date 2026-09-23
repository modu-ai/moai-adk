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
- `SKIP`, `NOT_RUN`, `ABORTED`는 PASS가 아니다. SKIP이 나온 AC는 "증명되지 않음"이며 판정표에 `UNPROVEN`으로 적는다. 패키지 단위 `ok` 줄은 증거로 쓰지 않는다(REQ-DHR-024). plan 단계에서 `TestFactoryLive*` 4개가 모두 `SKIP`인 채 `ok ... internal/cli`가 출력되는 것을 관측했다(`research.md` §D).
- 테스트 이름은 run 단계에서 만들 이름이다. 이름을 바꾸면 이 파일의 명령도 함께 고친다. 이름이 없으면 pass 수가 0이 되어 FAIL로 판정된다.
- 증거는 `.moai/reports/t1100/`에 남긴다.
- `internal/cli` 명령은 kanban 환경 변수를 같은 호출 안에서 지운다(`unset ... && go test ...`).
- 명령에는 실행 중에 계산한 값을 git·go 명령으로 넘기는 형태를 쓰지 않는다(워크트리 세션 가드가 거부한다). 조합마다 리터럴 명령을 쓴다.
- LIVE AC(AC-DHR-012, AC-DHR-018, AC-DHR-023)는 결정적 AC와 따로 집계한다. LIVE 미실행은 `NOT_RUN`, 예산 초과 중단은 `ABORTED`로 적고 해당 설계 기준을 PASS로 올리지 않는다. LIVE 판정은 테스트 pass만이 아니라 증거 파일의 양성 증거(시도 증거, 양성 대조, 호출 수)를 요구한다.
- 증거 채널(AC-DHR-012, 018, 020, 023): 증거 JSON은 표준 출력 줄이 아니라 파일로 남긴다. `go test -json`은 1024바이트를 넘는 출력 줄을 여러 `output` 이벤트로 쪼개므로(`$(go env GOROOT)/src/cmd/internal/test2json/test2json.go` `outBuffer = 1024`), 판정식이 한 이벤트에 긴 줄 하나가 담겨 있기를 기대하면 올바른 구현도 `true`를 낼 수 없다(plan-audit iter-2 ND3 실측: 긴 증거 줄이 4조각, `jq` exit 5).
  - 테스트는 환경 변수 `MOAI_T1100_EVIDENCE_DIR`이 가리키는 디렉터리에 AC별 증거 파일을 쓴다. 명령은 이 값을 리터럴 `../../.moai/reports/t1100`으로 준다. `go test`는 테스트 바이너리를 패키지 디렉터리에서 실행하고, 대상 패키지 `internal/cli`와 `internal/factorymsg`는 모두 모듈 뿌리에서 두 단계 아래이므로 이 경로가 저장소의 `.moai/reports/t1100/`이 된다. 값이 비어 있으면 테스트는 `NOT_RUN`을 찍고 실패한다.
  - 파일을 쓴 뒤 테스트는 표준 출력에 짧은 고정 형식 줄 `<TAG>_SHA256 [<case> ]<64자 hex>`을 한 줄 찍는다(`fmt.Println`, 약 100바이트). 이 줄은 1 KiB보다 훨씬 짧아 쪼개지지 않는다.
  - 실행 명령은 먼저 해당 증거 파일을 `rm -f`로 지운다. 판정 명령은 `shasum -a 256`으로 파일 해시를 다시 재고, 태그 줄의 해시와 같을 때만 파일 내용을 판정한다. 파일이 없으면 `shasum`이 실패해 판정식에 닿지 않으므로 `true`가 출력되지 않는다(FAIL).
  - 이 채널 자체의 양·음 사례(1 KiB를 넘는 증거 파일, 해시 불일치, 파일 부재)는 AC-DHR-019 두 번째 명령이 판정한다.
- 호출 수의 단위는 `codex exec`(또는 `claude -p`) 프로세스 하나다. 그 프로세스 안의 하위 에이전트는 따로 세지 않는다.

공통 판정식(각 명령에 그대로 들어 있다): 지정 이름의 `pass` 이벤트 수가 기대값과 같고, `fail`·`skip` 이벤트가 0이며, 출력에 `NOT_RUN`·`ABORTED`가 없다.

### AC ↔ 요구사항 매핑

| AC | Requirement | 설계 기준 |
|---|---|---|
| AC-DHR-001 | REQ-DHR-002, REQ-DHR-003, REQ-DHR-006 | AC-MIG-01 |
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
| AC-DHR-015 | REQ-DHR-017, REQ-DHR-018, REQ-DHR-020 | AC-MSG-01 |
| AC-DHR-016 | REQ-DHR-018, REQ-DHR-021 | AC-MSG-01 |
| AC-DHR-017 | REQ-DHR-022 | AC-FACT-01 |
| AC-DHR-018 | REQ-DHR-023 | AC-FACT-01 (LIVE) |
| AC-DHR-019 | REQ-DHR-024 | AC-FACT-01 |
| AC-DHR-020 | REQ-DHR-025 | AC-MSG-01 |
| AC-DHR-021 | REQ-DHR-004 | AC-MIG-01 |
| AC-DHR-022 | REQ-DHR-002 | AC-MIG-01 |
| AC-DHR-023 | REQ-DHR-015 | AC-AGENT-01 (LIVE) |

## §B 인수 기준

### AC-DHR-001 — 저널 기반 쓰기, 대조, 동시 수정 거부, 설치 쪽 symlink 경계 (REQ-DHR-002, REQ-DHR-003, REQ-DHR-006)

**Given** 사용자 handler가 든 `.codex/hooks.json`과 사용자 테이블이 든 `.codex/config.toml`이 있는 임시 프로젝트,
**When** 배선을 실행하되 (a) 방해 없이, (b) 해시 재확인 직전에 테스트 seam으로 대상 파일을 바꿔서, (c) `.codex/hooks.json`이 프로젝트 안 파일을 가리키는 symlink인 상태에서, (d) `.codex` 디렉터리가 프로젝트 밖을 가리키는 symlink인 상태에서 실행하면,
**Then** (a)는 저널 항목이 `complete`이고 기록 후 해시가 의도한 해시와 같으며 manifest에 부분 기록이 있다. (b)는 대상이 바이트 그대로이고 임시 파일이 남지 않으며 저널 항목이 `conflict`이고 두 해시가 보고된다. (c)와 (d)는 링크와 링크 대상이 모두 바이트 그대로이고, 프로젝트 밖에 파일이 생기지 않으며, `symlink-boundary`가 보고된다.

음성·변이: rename 직전 재확인을 지운 변이, 기록 후 대조를 지운 변이, Lstat 검사를 지운 변이는 각각 이 테스트에서 실패해야 한다(테스트 안의 변이 표로 확인).

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/codexwiring -run '^(TestCodexWiringJournaledWriteReadback|TestCodexWiringRefusesConcurrentModification|TestCodexWiringInstallRefusesSymlinkBoundary)$' -count=1 -timeout=120s > .moai/reports/t1100/ac001.jsonl; jq -se '([.[]|select(.Action=="pass" and ((.Test//"")|test("^(TestCodexWiringJournaledWriteReadback|TestCodexWiringRefusesConcurrentModification|TestCodexWiringInstallRefusesSymlinkBoundary)$")))]|length)==3 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac001.jsonl
```

### AC-DHR-002 — 중단 지점 전부의 복구 (REQ-DHR-004)

**Given** `design.md` §A.4의 중단 지점 아홉 곳(P0 저널 없는 임시 파일, P1 저널 추가 후, P2 임시 파일 작성 후, P3 재확인 후 rename 전, P4 rename 후 대조 전, P5 대조 후 manifest 반영 전, P6 manifest 반영 후 완료 표시 전, D1 unwire 삭제 저널 후 삭제 전, D2 삭제 후 manifest 반영 전)에서 멈춘 실행 아홉 건과, P3에서 멈춘 뒤 사용자가 대상을 바꾼 한 건,
**When** 각 건에 대해 복구를 실행하면,
**Then** 각 건은 설계 표의 분류(`completed` / `not-applied` / `diverged`)를 받고, `completed` 건은 manifest에 저널의 provenance가 반영되며, `diverged` 대상은 바이트 그대로 남되 그 항목이 참조하던 임시 파일은 지워지며(REQ-DHR-004), 어떤 경우에도 `.codexwiring-*` 임시 파일이 남지 않는다.

음성: `diverged` 건을 복구가 덮어쓰면 FAIL. `P3_user_modified` 건에서 임시 파일이 남으면 FAIL. P6 건에서 provenance가 두 번 적용되어 부분 기록이 중복되면 FAIL.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/codexwiring -run '^TestCodexWiringInterruptedChangeRecovery$' -count=1 -timeout=120s > .moai/reports/t1100/ac002.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexWiringInterruptedChangeRecovery")]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestCodexWiringInterruptedChangeRecovery/(P[0-6]|D[12]|P3_user_modified)$")))]|length)==10 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac002.jsonl
```

### AC-DHR-003 — MoAI 소유 부분만 제거, 형식별 보존 기준 (REQ-DHR-001, REQ-DHR-005)

**Given** 다섯 픽스처 — (1) MoAI가 새로 만든 `.codex/hooks.json`(MoAI handler와 MoAI가 추가한 `description`), (2) 사용자 최상위 키·사용자 handler·MoAI handler가 한 entry에 섞인 `hooks.json`, (3) 사용자 테이블 + MoAI가 덧붙인 `[mcp_servers.moai]` + MoAI가 통째로 덧붙인 `[tui]`가 든 `config.toml`, (4) 사용자 `[tui]`에 MoAI가 `status_line` 한 줄만 넣은 `config.toml`, (5) 설치 전부터 `[mcp_servers.moai]`가 있던 `config.toml`,
**When** `moai tool disable codex`의 unwire 단계를 실행하면,
**Then** (1)은 파일이 삭제된다. (2)는 unwire 직전 파일을 파싱한 구조와 비교해 MoAI 외 최상위 키 값, 사용자 entry의 matcher, 사용자 handler가 JSON 값으로 같고 순서도 같으며 MoAI handler와 MoAI가 추가한 `description`만 빠진다. (3)과 (4)는 결과 바이트가 unwire 직전 바이트에서 기록된 영역(MoAI가 넣은 구분 빈 줄 포함)만 잘라 낸 것과 `bytes.Equal`이다. (5)의 `[mcp_servers.moai]`는 남고 `user-owned`로 보고된다.

변이: 사전 존재 테이블을 `created`로 기록하는 변이, 구분 빈 줄을 영역에서 빼는 변이는 FAIL해야 한다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/codexwiring -run '^TestCodexUnwireRemovesOnlyOwnedParts$' -count=1 -timeout=120s > .moai/reports/t1100/ac003.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexUnwireRemovesOnlyOwnedParts")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac003.jsonl
```

### AC-DHR-004 — 해시만으로 지우지 않음 (REQ-DHR-006, REQ-DHR-001)

**Given** provenance 기록이 없지만 내용 해시가 정본과 같은 파일, 사이드카는 있으나 부분 기록이 없는 구버전 설치의 `config.toml`, 기록은 MoAI 소유지만 사용자가 수정한 파일, 프로젝트 밖을 가리키는 symlink인 `.codex/hooks.json`,
**When** unwire를 실행하면,
**Then** 네 대상 모두 바이트 그대로(symlink 대상 포함) 남고, 각각 `no-provenance`, `unknown-origin`, `modified`, `symlink-boundary`로 보고된다. 구버전 설치를 한 번 다시 배선한 뒤에도 그 `config.toml` 부분은 `unknown-origin`으로 남는다.

변이: provenance 확인을 지우고 해시만 비교하는 변이는 첫 대상을 지우므로 FAIL해야 한다. `unknown`을 재배선 때 `created`로 올리는 변이는 FAIL해야 한다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/codexwiring -run '^TestCodexUnwireRefusesUnownedModifiedSymlink$' -count=1 -timeout=120s > .moai/reports/t1100/ac004.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexUnwireRefusesUnownedModifiedSymlink")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac004.jsonl
```

### AC-DHR-005 — 하네스 프로필 전환 행렬 (REQ-DHR-007)

**Given** 전환 여섯 행(`claude → both`, `gpt → both`, `both → claude`, `both → gpt`, `gpt → claude`, 구버전 배포 → 신버전)마다 `.codex/` 아래 사용자 파일 1개(예: 사용자 역할 TOML), 사용자 테이블이 든 `.codex/config.toml`, 손상된 `.codex/hooks.json` 1개, 수정되지 않은 `.codex/` 템플릿 관리 파일이 있는 임시 프로젝트,
**When** 각 전환을 `moai update` 경로로 적용하면,
**Then** 모든 행에서 `.codex/` 아래 사용자 소유 파일·부분이 바이트 그대로이고, 고아가 된 배선 파일과 더 이상 배포되지 않는 `.codex/` 템플릿 관리 파일은 바이트 그대로 남은 채 보고되며, 배선 파일 보고에는 `moai tool disable codex`가 적힌다. 손상된 `hooks.json`은 건드리지 않는다.

이 AC는 `.claude/` 관리 뿌리의 상태를 판정하지 않는다. update의 관리 경로 정리 단계는 프로필과 무관하게 그 뿌리를 지우며(`design.md` §A.6, 코드 판독·미측정), 이 SPEC의 범위 밖이다(`spec.md` §F). 테스트는 `.claude/` 아래 파일의 존재나 부재를 단언하지 않는다.

변이: update가 고아 배선 파일을 지우거나 고치는 변이는 FAIL해야 한다.

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestHarnessProfileTransitionPreservesUserData$' -count=1 -timeout=600s > .moai/reports/t1100/ac005.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestHarnessProfileTransitionPreservesUserData")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac005.jsonl
```

### AC-DHR-006 — Codex 트리 anchor, 교체 경합, 생성 base (REQ-DHR-008, REQ-DHR-011)

**Given** 임시 git 저장소와 설정된 base 브랜치,
**When** `moai codex -w new-tree`(새 트리)와 `moai codex -w existing-tree`(기존 트리)를 launch seam으로 실행하고, 죽은 pid lock이 걸린 트리에 두 launcher가 교체를 동시에 시도하게 하면,
**Then** 두 트리 모두 codex 프로세스가 될 pid를 담은 git worktree lock을 갖고, 기존 anchor 판정이 두 트리를 anchored로 본다. 새 트리의 HEAD는 해석된 base 커밋과 같다. base 해석 결과와 생성된 HEAD가 다르도록 seam을 바꾸면 launch가 거부되고 트리는 삭제되지 않는다. 죽은 pid의 lock은 교체되고, 살아 있는 pid의 lock은 교체되지 않는다. 동시 교체에서는 정확히 한 launcher만 진행하고 다른 하나는 0이 아닌 종료로 거부되며, 최종 lock 사유의 pid는 진행한 launcher의 것이다.

Windows 분기(대기하는 부모 pid로 lock)는 `//go:build windows` 코드이므로 darwin 로컬 실행으로는 검증되지 않는다. CI의 Windows 잡에서 이 테스트의 pass가 관측될 때만 PASS로 인용하고, 그 전에는 Windows 분기를 `NOT_RUN`으로 적는다.

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^(TestCodexWorktreeAnchorLockAndBase|TestCodexWorktreeAnchorLockReplacementRace)$' -count=1 -timeout=600s > .moai/reports/t1100/ac006.jsonl; jq -se '([.[]|select(.Action=="pass" and ((.Test//"")|test("^(TestCodexWorktreeAnchorLockAndBase|TestCodexWorktreeAnchorLockReplacementRace)$")))]|length)==2 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac006.jsonl
```

### AC-DHR-007 — 동시 writer 거부 (REQ-DHR-009)

**Given** 살아 있는 다른 프로세스 pid로 lock된 트리, 판독 불가 lock 사유를 가진 트리, 레지스트리에만 살아 있는 세션이 기록된 트리,
**When** 각 트리에 `moai codex -w`와 `moai cc -w`를 실행하면,
**Then** 여섯 호출이 모두 0이 아닌 종료 코드와 anchor 출처·보유자를 적은 진단을 내고, 트리의 lock·브랜치·작업 파일은 바뀌지 않는다.

음성: 죽은 pid의 lock만 있는 트리에는 launch가 허용되어야 한다(과잉 거부 방지).

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestWorktreeLaunchRejectsConcurrentWriter$' -count=1 -timeout=600s > .moai/reports/t1100/ac007.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestWorktreeLaunchRejectsConcurrentWriter")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac007.jsonl
```

### AC-DHR-008 — L1 Codex 트리를 지울 수 있는 경로의 보호와 `done`의 L1 거부 유지 (REQ-DHR-010)

**Given** `moai codex -w <name>`으로 `<project root>/.claude/worktrees/` 아래 만든 L1 트리 네 개 — (1) base에 없는 커밋 보유, (2) 미커밋 변경 보유, (3) 살아 있는 pid의 codex lock 보유, (4) base에 병합이 끝나고 lock이 없으며 깨끗함 — 와 같은 네 상태의 Claude 생성 L1 트리 네 개,
**When** `moai worktree done`(`--force` 없이, 그리고 `--force`로), `moai worktree clean --stale --yes`, `moai worktree remove`(`--force` 없이), PR-merge 정리(`auto_cleanup` 켬)를 각 트리에 적용하면,
**Then**
- `done`: 여덟 트리 모두 `--force` 유무와 무관하게 `L1_SESSION_WORKTREE`로 거부되고 트리가 남는다(SPEC-WORKTREE-DONE-TIER-001의 기존 계약. 이 SPEC은 바꾸지 않는다).
- `clean --stale --yes`: (1)(2)(3)은 남고 각각 미병합·미커밋·anchor(출처 lock)가 사유로 보고되며, (4)만 제거된다.
- `remove`: (3)은 anchor 출처(lock)와 보유 pid를 적은 거부로 끝나고 트리가 남는다. (2)는 거부되고 트리가 남는다.
- PR-merge 정리: (2)(3)과 unpushed 커밋이 있는 트리는 남고 사유(`dirty`, `anchored-by-lock`, `unpushed-commits` 중 해당)가 보고된다.
- 모든 경로에서 Codex 트리의 결과(남음·제거·사유 종류)가 같은 상태의 Claude 트리 결과와 같다.

변이: `remove`의 anchor 판정을 레지스트리 기반으로 되돌린 변이는 (3)이 git 오류로만 거부되거나 anchor 출처가 보고되지 않아 FAIL해야 한다. `done`에서 L1 tier 거부를 `--force`로 넘기게 한 변이는 FAIL해야 한다.

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli/worktree ./internal/cli -run '^(TestWorktreeDisposalRefusesUnintegratedCodexTree|TestPRMergeCleanupRefusesAnchoredCodexTree)$' -count=1 -timeout=600s > .moai/reports/t1100/ac008.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestWorktreeDisposalRefusesUnintegratedCodexTree" and ((.Package//"")|endswith("/internal/cli/worktree")))]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestPRMergeCleanupRefusesAnchoredCodexTree" and ((.Package//"")|endswith("/internal/cli")))]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestWorktreeDisposalRefusesUnintegratedCodexTree/(done|done_force|clean|remove)$")))]|length)==4 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac008.jsonl
```

### AC-DHR-009 — Codex kanban 진입 동등성 (REQ-DHR-012)

**Given** `moai cc -k`의 lead·companion 진입 형태 표,
**When** 같은 인자로 `moai codex`와 `moai cc`를 launch seam으로 실행하면,
**Then** lead와 companion 각각에서 두 경로가 같은 kanban launch facts(backend만 `codex`/`claude`로 다름)와 같은 세션 이름 규칙을 내고, codex 자식 인자에 `-k`/`--kanban`/역할 토큰이 없다. 지원하지 않는 형태는 사용법 진단과 0이 아닌 종료로 끝나며 일반 launch로 떨어지지 않는다.

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestCodexKanbanEntryParity$' -count=1 -timeout=600s > .moai/reports/t1100/ac009.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexKanbanEntryParity")]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestCodexKanbanEntryParity/(lead|companion|unsupported)$")))]|length)==3 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac009.jsonl
```

### AC-DHR-010 — 역할 권한 계약과 조용한 누락 금지 (REQ-DHR-013)

**Given** 12개 Codex 역할의 권한 계약(축 `sandbox`, `write-path-scope`, `shell`, `mcp-server`, `mcp-tool`, `subagent`, `web`),
**When** agent 생성기를 실행하면,
**Then** 모든 역할에서 계약이 제한을 요구하는 모든 축이 `enforced` 또는 `UNSUPPORTED` 중 하나로 매핑되고 근거(`measured` / `documented` / `unmeasured`)를 가지며, `unmeasured` 근거의 `enforced`는 없고, 생성된 `sandbox_mode`는 계약에 적힌 값과 같다.

변이: 한 역할의 한 축에서 매핑을 지운 변이, `read-only` 계약 역할을 `workspace-write`로 내보내는 변이, `unmeasured` 축을 `enforced`로 바꾼 변이는 생성기가 거부해야 한다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/template/agentemit -run '^TestRolePermissionContractNoSilentDrop$' -count=1 -timeout=120s > .moai/reports/t1100/ac010.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestRolePermissionContractNoSilentDrop")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac010.jsonl
```

### AC-DHR-011 — UNSUPPORTED는 PASS가 아님 (REQ-DHR-014)

**Given** 12개 역할의 계약과 `design.md` §C.1의 축별 매핑,
**When** 역할 검증 보고를 만들면,
**Then** 보고의 `UNSUPPORTED` 집합이 "계약이 제한을 요구하고 그 축이 `UNSUPPORTED`로 매핑된 (역할, 축)" 집합과 정확히 같고, 그 항목 어느 것도 PASS로 집계되지 않는다.

변이: `UNSUPPORTED`를 PASS로 세는 변이는 FAIL해야 한다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/template/agentemit -run '^TestRolePermissionUnsupportedNeverPass$' -count=1 -timeout=120s > .moai/reports/t1100/ac011.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestRolePermissionUnsupportedNeverPass")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac011.jsonl
```

### AC-DHR-012 — [LIVE] 12개 역할 실제 로드와 read-only 강제 (REQ-DHR-014)

**Given** 격리된 임시 저장소와 임시 `CODEX_HOME`, 설치된 codex 바이너리, 호출 예산 14회(역할 로드 12 + 감사 역할 쓰기 시도 2; 리드 결정 4),
**When** (i) 12개 역할마다 한 번씩 역할별로 다른 nonce를 되돌려 달라는 작업을 주고, 그중 `workspace-write` 역할 하나에는 탐침 파일 쓰기를 함께 시키며(양성 대조), (ii) 부모 lane 오케스트레이터 세션을 두 번 띄워 각각 `plan-auditor`와 `sync-auditor` 하위 에이전트가 탐침 파일 쓰기를 시도하게 하면,
**Then** 증거 파일 `ac012-evidence.json`(§A 증거 채널)에 codex 버전, 12개 역할 이름과 각 역할이 받은 nonce·되돌린 nonce, 양성 대조 탐침 파일의 존재와 sha256, 두 감사 역할 각각의 역할 이름·쓰기 시도 출력(sandbox 거부 문구를 포함한 명령 출력)·거부 여부·탐침 파일 부재, 호출 수가 기록된다. 판정은 다음을 모두 요구한다.
- 역할 이름 집합이 생성된 `internal/template/templates/.codex/agents/moai/*.toml` 파일 이름 집합과 정확히 같다(판정 명령이 그 목록을 직접 만든다. 기대 개수 12).
- 모든 역할의 nonce가 비어 있지 않고, 되돌린 값이 보낸 값과 같으며, 12개 nonce가 서로 다르다.
- 양성 대조 탐침 파일이 있고 그 sha256이 64자 hex다.
- 쓰기 시도가 `plan-auditor`와 `sync-auditor` 각각 한 번씩이고, 둘 다 출력이 비어 있지 않으며 거부되었고 탐침 파일이 없다.
- 호출 수가 정확히 14다(예산 14, 필요한 호출도 14).
- 출력에 `Ignoring malformed agent role definition`이 없다.

MCP 경유 부작용은 이 AC가 증명하지 않으며 `UNSUPPORTED`로 남는다. `mission-governor`와 `super-advisor`의 read-only 강제는 이 AC가 측정하지 않는다.

음성·변이(판정식이 `false`여야 하는 입력, plan-audit iter-3에서 합성 입력으로 확인): 같은 역할 이름 12개와 빈 nonce(iter-2 변이), 같은 역할 이름 12개와 유효한 nonce, nonce 하나가 빈 값, 12개 역할이 같은 nonce, 생성 목록 밖의 역할 이름, 양성 대조 해시가 빈 값, 호출 수 0, 호출 수 15, 같은 감사 역할 두 번, 태그 줄 해시와 파일 해시 불일치, 테스트 skip. 증거 파일이 없으면 판정 명령이 판정식에 닿지 않아 `true`가 나오지 않는다.

SKIP 의미: `MOAI_CODEX_ROLE_LIVE=1`이 없으면 이 테스트는 SKIP하고 AC-DHR-012는 `NOT_RUN`이다. 15번째 호출이 필요해지면(호출 수가 14를 넘게 되면) 테스트는 그 호출을 시작하지 않고 남은 단계를 멈추며 `ABORTED`를 찍은 뒤 실패한다. 호출 수가 정확히 14로 끝난 실행은 `ABORTED`가 아니다. CI에서 이 게이트를 켜는 워크플로는 plan 시점에 관측되지 않았다.

실행:

```bash
mkdir -p .moai/reports/t1100 && rm -f .moai/reports/t1100/ac012-evidence.json .moai/reports/t1100/ac023-evidence.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && MOAI_CODEX_ROLE_LIVE=1 MOAI_T1100_EVIDENCE_DIR=../../.moai/reports/t1100 go test -json ./internal/cli -run '^TestCodexRoleLiveLoadAndReadOnly$' -count=1 -timeout=1800s > .moai/reports/t1100/ac012-live.jsonl
```

판정:

```bash
find internal/template/templates/.codex/agents/moai -maxdepth 1 -name '*.toml' | sed 's|.*/||; s/\.toml$//' > .moai/reports/t1100/ac012-roles.txt && shasum -a 256 .moai/reports/t1100/ac012-evidence.json > .moai/reports/t1100/ac012-evidence.sha && jq -se --rawfile roles .moai/reports/t1100/ac012-roles.txt --rawfile sha .moai/reports/t1100/ac012-evidence.sha --slurpfile ev .moai/reports/t1100/ac012-evidence.json '($sha|.[0:64]) as $h | ($roles|split("\n")|map(select(length>0))|sort) as $want | ($ev[0]) as $e | ([.[]|select(.Action=="pass" and .Test=="TestCodexRoleLiveLoadAndReadOnly")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED|Ignoring malformed agent role definition"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^AC012_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^AC012_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and ($want|length)==12 and (($e.codex_version|type)=="string" and ($e.codex_version|test("^[0-9]"))) and $e.invocations==14 and ([$e.roles[].name]|sort)==$want and ([$e.roles[]|select((.nonce_sent|type)=="string" and (.nonce_sent|length)>0 and .nonce_returned==.nonce_sent)]|length)==12 and ([$e.roles[].nonce_sent]|unique|length)==12 and $e.positive_control.exists==true and (($e.positive_control.sha256//"")|test("^[0-9a-f]{64}$")) and ([$e.write_attempts[].role]|sort)==["plan-auditor","sync-auditor"] and ([$e.write_attempts[]|select((.attempt_output|type)=="string" and (.attempt_output|length)>0 and .denied==true and .probe_exists==false)]|length)==2' .moai/reports/t1100/ac012-live.jsonl
```

### AC-DHR-013 — Codex 감사 역할 read-only 예외의 범위 (REQ-DHR-015)

**Given** 생성기 입력(`agents-codex.yaml`), 생성기의 neutral 원본인 C2 `internal/template/templates/.claude/agents/moai/plan-auditor.md`·`sync-auditor.md`, 로컬 사본인 C1 `.claude/agents/moai/plan-auditor.md`·`sync-auditor.md`,
**When** agent 생성기를 실행하면,
**Then** 생성된 `plan-auditor.toml`과 `sync-auditor.toml`의 `sandbox_mode`가 `read-only`이고, 두 파일에만 Codex 전용 반환문 지시가 있으며, C1·C2의 네 파일에는 그 지시가 없고, Codex가 읽는 부모 지시 표면에 "반환문 그대로 판정 파일을 기록한다"는 항목이 있다.

변이: `role_values`에서 감사 역할을 뺀 변이, 반환문 지시를 neutral 원본에 넣는 변이, 부모 지시 항목을 지운 변이는 FAIL해야 한다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/template/agentemit -run '^TestCodexAuditRolesReadOnlyScopedException$' -count=1 -timeout=120s > .moai/reports/t1100/ac013.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditRolesReadOnlyScopedException")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac013.jsonl
```

### AC-DHR-014 — 중복·ack 유실·재시작에도 결과 한 번 (REQ-DHR-016 ~ REQ-DHR-019)

**Given** `started` 상태의 dispatch 하나와 AC-DHR-020이 기록한 멱등 범위 측정 결과,
**When** 같은 결과가 (a) 두 번 전달되고, (b) 수신자가 결과 적용 후 receipt 전에 멈춰 lease 만료로 재전달되고, (c) 송신 worker가 결과를 보낸 뒤 새 세션 UUID로 재시작해 같은 멱등 키로 다시 보내고, (d) 브로커 프로세스가 결과 적용 직후 재시작한 뒤 재전달되고, (e) (c) 이후 이전 generation 메시지가 다시 전달되고, (f) 같은 attempt에 다른 본문의 결과가 오면(구성은 아래),
**Then** (a)~(d)에서 dispatch는 `result_recorded`로 한 번만 바뀌고 result 참조는 처음 것과 같으며 두 번째 판정은 `duplicate`다. (e)는 `stale`, (f)는 `collision`이고 레코드는 바뀌지 않는다. 결과 적용이 receipt보다 먼저 기록된다.

(f)의 구성: 같은 송신 세션이 같은 멱등 키로 다른 본문을 보내면 메시지 층이 먼저 `idempotency key collision with different request`로 거부하므로(`store.go:624-626`) 그 경로로는 dispatch 층의 `collision`에 닿지 않는다. 테스트는 그 메시지 층 거부를 먼저 확인하고(메시지 행 수 불변), 이어서 메시지 층을 거치지 않고 결과 적용 연산을 직접 호출해 같은 attempt에 다른 digest의 결과를 넣어 `collision`을 확인한다. 분기 A(범위 이관)에서도 같은 순서다. 재시작한 송신자로 보내는 방법은 분기 A에서 메시지 층이 다시 거부하므로 쓰지 않는다.

(c)의 메시지 층 결과는 AC-DHR-020의 측정 결과에 따라 둘 중 하나만 적용한다. 측정 결과와 다른 분기의 명령은 `N/A (branch)`로 적고 PASS로 세지 않는다.

- 분기 A(`reproduced`, 범위 이관): 재시작 뒤 같은 키의 전송은 원래 메시지를 돌려주고 메시지 행은 1개다. 기존 DB 행은 이관 후 모두 남는다.
- 분기 B(`not-reproduced`, 현행 유지): 스키마가 바뀌지 않았고(`UNIQUE(sender_session,idem_key)`), 결과는 한 번만 적용된다.

분기는 AC-DHR-020의 증거 파일(`ac020-evidence.json`)의 `outcome`으로 고른다. AC-DHR-020이 `true`가 아니면(측정 실패 또는 `NOT_RUN`) 어느 분기도 판정하지 않고 AC-DHR-014 전체를 `UNPROVEN`으로 적는다.

변이: 판정 순서 5를 지운 변이는 (a)·(c)의 두 번째 판정이 `duplicate`가 아니라 (a)에서는 (7) `invalid-state`, (c)에서는 (6) `stale`이 되어 FAIL해야 한다(결과가 두 번 적용되지는 않는다. 상태가 이미 `result_recorded`이고 두 번째 적용은 뒤 단계에서 거부된다). 4와 5의 순서를 바꾼 변이는 (e)가 `duplicate`가 되어 FAIL해야 한다. 분기 A에서 범위를 송신자 세션 UUID로 되돌린 변이는 (c)에서 행이 2개가 되어 FAIL해야 한다.

공통 명령:

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/factorymsg -run '^TestDispatchResultExactlyOnce$' -count=1 -timeout=120s > .moai/reports/t1100/ac014.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestDispatchResultExactlyOnce")]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestDispatchResultExactlyOnce/(duplicate_delivery|lost_receipt_redelivery|sender_restart_after_result|broker_restart|old_generation_redelivery|collision)$")))]|length)==6 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac014.jsonl
```

분기 A 명령(AC-DHR-020이 `reproduced`일 때만):

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/factorymsg -run '^(TestDispatchResultExactlyOnce|TestIdemScopeLaneMigration)$' -count=1 -timeout=120s > .moai/reports/t1100/ac014-branch.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestDispatchResultExactlyOnce/sender_restart_lane_scope")]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestIdemScopeLaneMigration")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac014-branch.jsonl
```

분기 B 명령(AC-DHR-020이 `not-reproduced`일 때만):

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/factorymsg -run '^TestDispatchResultExactlyOnce$' -count=1 -timeout=120s > .moai/reports/t1100/ac014-branch.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestDispatchResultExactlyOnce/sender_restart_session_scope")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac014-branch.jsonl
```

### AC-DHR-015 — 낡은 fencing token, superseded 메시지, 같은 attempt 재부여 (REQ-DHR-017, REQ-DHR-018, REQ-DHR-020)

**Given** generation 1로 할당된 dispatch와, 수신 lane이 generation 2로 재등록된 뒤에도 generation 1 앞으로 남은 메시지,
**When** generation 1 주체가 결과를 적용하고 메시지를 claim·read·dispose·ack하려 하고, 이어서 같은 attempt 재부여 연산으로 assignee generation을 2로 바꾼 뒤 generation 2가 결과를 적용하면,
**Then** generation 1의 모든 시도는 `stale`로 거부되고 dispatch·메시지 행은 바뀌지 않으며, 브로커 status는 그 메시지를 `pending`이 아니라 `superseded`로 센다. 재부여 뒤 generation 2의 결과는 `accepted`이고 attempt는 그대로다. 브로커는 superseded 메시지 body를 generation 2에게 스스로 넘기지 않는다. 재부여 뒤 lead가 같은 멱등 키로 할당 메시지를 generation 2에게 다시 보내면 메시지 층이 수신 generation 불일치로 거부하고(REQ-DHR-017의 "different recipient") 메시지 행 수는 그대로다. 리드 조정 결정(2026-09-23, `spec.md` §E)에 따라 그 재전송은 새 키를 쓰며, 새 키로 보낸 할당은 저장된다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/factorymsg -run '^TestDispatchStaleFencingAndSuperseded$' -count=1 -timeout=120s > .moai/reports/t1100/ac015.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestDispatchStaleFencingAndSuperseded")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac015.jsonl
```

### AC-DHR-016 — 재할당과 늦은 응답 (REQ-DHR-018, REQ-DHR-021)

**Given** attempt 1이 `started`인 dispatch,
**When** (a) attempt 1 소유자가 살아 있는 상태에서 재할당하고, (b) 소유자 종료가 확인된 뒤 다른 lane으로 재할당하고, (c) attempt 2가 결과를 기록한 뒤 attempt 1의 늦은 결과가 도착하면,
**Then** (a)는 명시적 철회 없이 거부되고, (b)는 attempt 2와 새 assignee lane·generation을 기록하며 attempt 2의 할당 메시지가 멱등 키 충돌 없이 저장되고, (c)의 늦은 결과는 `stale`로 거부되어 적용된 결과는 attempt 2의 것 하나뿐이다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/factorymsg -run '^TestDispatchReassignmentFencesLateResult$' -count=1 -timeout=120s > .moai/reports/t1100/ac016.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestDispatchReassignmentFencesLateResult")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac016.jsonl
```

### AC-DHR-017 — 결정적 4조합 카드 흐름 (REQ-DHR-022)

**Given** backend 표지가 붙은 lead·worker peer 네 쌍(Claude→Claude, Codex→Codex, Claude→Codex, Codex→Claude)과 실제 브로커 저장소,
**When** 각 쌍이 dispatch를 `assigned → delivered → started`로 진행한 뒤 worker 중단, 새 worker로 재할당, 새 attempt의 결과 기록, 옛 attempt의 늦은 결과, lead의 `integrated` 기록을 거치면,
**Then** 네 하위 테스트 모두 dispatch가 `integrated`로 끝나고 적용된 결과는 하나이며, 메시지 도착만으로는 `delivered`를 넘지 않는다.

```bash
mkdir -p .moai/reports/t1100 && go test -json ./internal/factorymsg -run '^TestFactoryCardFlowFourCombinations$' -count=1 -timeout=180s > .moai/reports/t1100/ac017.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryCardFlowFourCombinations")]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestFactoryCardFlowFourCombinations/(claude-claude|codex-codex|claude-codex|codex-claude)$")))]|length)==4 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac017.jsonl
```

### AC-DHR-018 — [LIVE] 실제 CLI 4조합 카드 흐름 (REQ-DHR-023)

**Given** 격리된 임시 저장소·`MOAI_HOME`·`CODEX_HOME`, 조합당 모델 호출 8회·900초 예산(리드 결정 4), 첫 프로세스 생성 전에 등록한 정리 함수,
**When** 네 조합을 한 번에 하나씩 실행하면(기존 `requireFactoryLive`는 `MOAI_FACTORY_LIVE_CASE`가 다르면 `Fatal`하므로 조합마다 별도 호출),
**Then** 각 조합의 증거 파일 `ac018-evidence-<case>.json`(§A 증거 채널, 태그 줄은 `AC018_EVIDENCE_SHA256 <case> <hex>`)에 조합 이름, nonce, 적용된 결과 수 1, 늦은 결과의 판정 `stale`, 호출 수(1 이상 8 이하), 경과 초(900 이하), 정리한 프로세스 pid 목록(비어 있지 않음)이 기록되고, 각 파일의 조합 이름이 파일 이름의 조합과 같으며, 네 nonce가 서로 다르고, 테스트가 pass한다.

음성·변이(판정식이 `false`, 합성 입력으로 확인): 한 조합 호출 9, 한 조합 정리 pid 없음, 네 조합 같은 nonce, 파일 이름과 다른 조합 이름, 실행 뒤 수정된 증거 파일(태그 해시 불일치). 증거 파일 하나가 없으면 판정 명령이 판정식에 닿지 않는다.

SKIP 의미: `MOAI_FACTORY_LIVE=1`이 없으면 네 테스트는 SKIP하며 AC-DHR-018은 `NOT_RUN`, AC-FACT-01은 증명되지 않음이다. 예산을 넘게 된 조합(9번째 모델 호출이 필요해지거나 경과 시간이 900초를 넘음)은 그 호출 전에 멈추고 `ABORTED`를 찍은 뒤 실패한다. 호출 8회 이하·900초 이하로 끝난 조합은 `ABORTED`가 아니다. plan 시점에 `.github/workflows/`에서 `MOAI_FACTORY_LIVE`를 쓰는 워크플로는 관측되지 않았으므로 CI 실행을 주장하지 않는다.

```bash
mkdir -p .moai/reports/t1100 && rm -f .moai/reports/t1100/ac018-evidence-claude-claude.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=cardflow-claude-claude MOAI_T1100_EVIDENCE_DIR=../../.moai/reports/t1100 go test -json ./internal/cli -run '^TestFactoryLiveCardFlowClaudeClaude$' -count=1 -timeout=1000s > .moai/reports/t1100/ac018-live-claude-claude.jsonl
```

```bash
rm -f .moai/reports/t1100/ac018-evidence-codex-codex.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=cardflow-codex-codex MOAI_T1100_EVIDENCE_DIR=../../.moai/reports/t1100 go test -json ./internal/cli -run '^TestFactoryLiveCardFlowCodexCodex$' -count=1 -timeout=1000s > .moai/reports/t1100/ac018-live-codex-codex.jsonl
```

```bash
rm -f .moai/reports/t1100/ac018-evidence-claude-codex.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=cardflow-claude-codex MOAI_T1100_EVIDENCE_DIR=../../.moai/reports/t1100 go test -json ./internal/cli -run '^TestFactoryLiveCardFlowClaudeCodex$' -count=1 -timeout=1000s > .moai/reports/t1100/ac018-live-claude-codex.jsonl
```

```bash
rm -f .moai/reports/t1100/ac018-evidence-codex-claude.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=cardflow-codex-claude MOAI_T1100_EVIDENCE_DIR=../../.moai/reports/t1100 go test -json ./internal/cli -run '^TestFactoryLiveCardFlowCodexClaude$' -count=1 -timeout=1000s > .moai/reports/t1100/ac018-live-codex-claude.jsonl
```

판정(네 조합을 한 번에):

```bash
shasum -a 256 .moai/reports/t1100/ac018-evidence-claude-claude.json .moai/reports/t1100/ac018-evidence-codex-codex.json .moai/reports/t1100/ac018-evidence-claude-codex.json .moai/reports/t1100/ac018-evidence-codex-claude.json > .moai/reports/t1100/ac018-evidence.sha && jq -s . .moai/reports/t1100/ac018-evidence-claude-claude.json .moai/reports/t1100/ac018-evidence-codex-codex.json .moai/reports/t1100/ac018-evidence-claude-codex.json .moai/reports/t1100/ac018-evidence-codex-claude.json > .moai/reports/t1100/ac018-evidence-all.json && jq -se --rawfile sha .moai/reports/t1100/ac018-evidence.sha --slurpfile ev .moai/reports/t1100/ac018-evidence-all.json '($sha|split("\n")|map(select(length>0)|capture("^(?<h>[0-9a-f]{64})  .*ac018-evidence-(?<c>[a-z]+-[a-z]+)\\.json$")|{(.c):.h})|add) as $fh | ($ev[0]) as $e | [.[]|select((.Output//"")|test("^AC018_EVIDENCE_SHA256 [a-z]+-[a-z]+ [0-9a-f]{64}\n?$"))|.Output|capture("^AC018_EVIDENCE_SHA256 (?<c>[a-z]+-[a-z]+) (?<h>[0-9a-f]{64})")|{(.c):.h}] as $tl | ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestFactoryLiveCardFlow(ClaudeClaude|CodexCodex|ClaudeCodex|CodexClaude)$")))]|length)==4 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($fh|keys)==["claude-claude","claude-codex","codex-claude","codex-codex"] and ($tl|length)==4 and ($tl|add)==$fh and ($e|length)==4 and [$e[].case]==["claude-claude","codex-codex","claude-codex","codex-claude"] and ([$e[]|select((.nonce|type)=="string" and (.nonce|length)>0 and .applied_results==1 and .late_result_outcome=="stale" and .invocations>=1 and .invocations<=8 and .elapsed_seconds<=900 and (.cleaned_pids|length)>0)]|length)==4 and ([$e[].nonce]|unique|length)==4' .moai/reports/t1100/ac018-live-claude-claude.jsonl .moai/reports/t1100/ac018-live-codex-codex.jsonl .moai/reports/t1100/ac018-live-claude-codex.jsonl .moai/reports/t1100/ac018-live-codex-claude.jsonl
```

### AC-DHR-019 — 판정식이 SKIP·NOT_RUN·ABORTED를 거부함 (REQ-DHR-024)

**Given** 이 파일이 쓰는 공통 판정식,
**When** 여섯 가지 합성 `go test -json` 흐름(정상 pass, 전부 skip, 이름 불일치, fail 포함, 출력에 `NOT_RUN`, 출력에 `ABORTED`)을 넣으면,
**Then** 정상 흐름만 `true`이고 나머지 다섯은 `false`다. 증거 채널(§A)도 같은 방식으로 시험한다: 4096바이트 증거 파일과 해시가 맞는 태그 줄은 `true`, 태그 줄 해시가 다르면 `false`, 증거 파일이 없으면 판정식에 닿지 않아 출력이 없다. 두 명령 모두 마지막 출력이 `true`여야 PASS다.

```bash
P='([.[]|select(.Action=="pass" and .Test=="TestX")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'; a=$(printf '{"Action":"pass","Test":"TestX"}\n{"Action":"pass"}\n' | jq -se "$P"); b=$(printf '{"Action":"skip","Test":"TestX"}\n{"Action":"pass"}\n' | jq -se "$P"); c=$(printf '{"Action":"pass","Test":"TestY"}\n' | jq -se "$P"); d=$(printf '{"Action":"pass","Test":"TestX"}\n{"Action":"fail","Test":"TestX/sub"}\n' | jq -se "$P"); e=$(printf '{"Action":"output","Test":"TestX","Output":"NOT_RUN\\n"}\n{"Action":"pass","Test":"TestX"}\n' | jq -se "$P"); f=$(printf '{"Action":"output","Test":"TestX","Output":"ABORTED budget\\n"}\n{"Action":"pass","Test":"TestX"}\n' | jq -se "$P"); [ "$a$b$c$d$e$f" = "truefalsefalsefalsefalsefalse" ] && echo true || echo false
```

증거 채널 시험:

```bash
d=$(mktemp -d) && head -c 4096 /dev/zero | tr '\0' 'x' | jq -Rsc '{pad:.}' > "$d/ev.json" && shasum -a 256 "$d/ev.json" > "$d/ev.sha" && printf '{"Action":"output","Test":"TestX","Output":"EV_SHA256 %s\\n"}\n{"Action":"pass","Test":"TestX"}\n' "$(cut -c1-64 "$d/ev.sha")" > "$d/ok.jsonl" && printf '{"Action":"output","Test":"TestX","Output":"EV_SHA256 %s\\n"}\n{"Action":"pass","Test":"TestX"}\n' "0000000000000000000000000000000000000000000000000000000000000000" > "$d/bad.jsonl" && Q='([.[]|select(.Action=="pass" and .Test=="TestX")]|length)==1 and ([.[]|select((.Output//"")|test("^EV_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^EV_SHA256 (?<h>[0-9a-f]{64})").h]==[($sha|.[0:64])]) and ($ev|length)==1 and (($ev[0].pad|length)>1024)' && a=$(jq -se --rawfile sha "$d/ev.sha" --slurpfile ev "$d/ev.json" "$Q" "$d/ok.jsonl") && b=$(jq -se --rawfile sha "$d/ev.sha" --slurpfile ev "$d/ev.json" "$Q" "$d/bad.jsonl"); c=$(shasum -a 256 "$d/absent.json" > "$d/absent.sha" 2>/dev/null && jq -se --rawfile sha "$d/absent.sha" --slurpfile ev "$d/absent.json" "$Q" "$d/ok.jsonl"); rm -rf "$d"; [ "$a|$b|$c" = "true|false|" ] && echo true || echo false
```

### AC-DHR-020 — 멱등 범위 재현 측정 (REQ-DHR-025)

run의 첫 마일스톤에서, factorymsg 스키마를 바꾸기 전 트리에서 한 번 실행하고 그 jsonl과 증거 파일을 보존한다. 스키마를 바꾼 뒤 다시 실행한 결과로 이 측정을 대신하지 않는다.

**Given** 바뀌지 않은 브로커 저장소, 송신 lane slot 하나, 수신 lane 하나,
**When** 송신자가 멱등 키 K로 보내고, 같은 slot을 새 세션 UUID로 재등록해 generation이 오른 뒤 같은 요청을 K로 다시 보내고, 새 키로 대조 메시지 하나를 보낸 뒤, 수신자가 claim하면,
**Then** 증거 파일 `ac020-evidence.json`(§A 증거 채널, 태그 줄 `IDEM_SCOPE_REPRO_SHA256 <hex>`)에 결과(`reproduced` 또는 `not-reproduced`), 저장소가 연 unique 제약 문자열, 송신자 세션 UUID 두 값과 generation 두 값, 첫 전송의 메시지 ID, 두 번째 전송의 결과(`new-id` / `same-id` / `error`), K를 가진 메시지 행 수, 수신자가 claim한 K 메시지의 서로 다른 ID 수, 대조 메시지 claim 수가 기록된다. 판정은 측정이 실제로 경로를 지났음을 먼저 요구한다: 두 세션 UUID가 비어 있지 않고 서로 다르며, generation이 올랐고, 첫 전송 ID가 있고, 대조 메시지가 1건 claim되었고, 행 수와 claim ID 수가 각각 1 이상이다. 그 위에서 결과가 수치와 맞아야 한다: `reproduced` ⇔ 행 수 ≥ 2, claim ID 수 ≥ 2, 두 번째 전송 결과 `new-id`. `not-reproduced` ⇔ 행 수 = 1, claim ID 수 = 1, 두 번째 전송 결과 `same-id`. 측정 시점의 HEAD는 `.moai/reports/t1100/ac020-head.txt`에 따로 남긴다.

측정이 경로를 지나지 않은 실행(첫 전송이 행을 남기지 않음, 재등록이 다른 세션·더 높은 generation을 만들지 않음, 두 번째 전송이 오류, 대조 메시지 claim 실패)에서 테스트는 `NOT_RUN`을 찍고 실패하며, 이 AC는 `NOT_RUN`이고 REQ-DHR-017의 어느 분기도 고르지 않는다(REQ-DHR-025).

음성·변이(판정식이 `false`, 합성 입력으로 확인): `not-reproduced`에 행 0·claim 0(iter-2 변이), 두 번째 전송 오류, 두 전송이 같은 세션, 대조 메시지 claim 0, `reproduced`인데 행 1, generation 미증가, 출력에 `NOT_RUN`. 정상 `reproduced`(2/2/`new-id`)와 정상 `not-reproduced`(1/1/`same-id`)만 `true`다.

실행:

```bash
mkdir -p .moai/reports/t1100 && git rev-parse --short HEAD > .moai/reports/t1100/ac020-head.txt
```

```bash
rm -f .moai/reports/t1100/ac020-evidence.json && MOAI_T1100_EVIDENCE_DIR=../../.moai/reports/t1100 go test -json ./internal/factorymsg -run '^TestIdemScopeRestartReproduction$' -count=1 -timeout=120s > .moai/reports/t1100/ac020.jsonl
```

판정:

```bash
shasum -a 256 .moai/reports/t1100/ac020-evidence.json > .moai/reports/t1100/ac020-evidence.sha && jq -se --rawfile sha .moai/reports/t1100/ac020-evidence.sha --slurpfile ev .moai/reports/t1100/ac020-evidence.json '($sha|.[0:64]) as $h | ($ev[0]) as $e | ([.[]|select(.Action=="pass" and .Test=="TestIdemScopeRestartReproduction")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^IDEM_SCOPE_REPRO_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^IDEM_SCOPE_REPRO_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and (($e.unique_constraint|type)=="string" and ($e.unique_constraint|length)>0) and (($e.sender_sessions|length)==2 and ($e.sender_sessions[0]|length)>0 and ($e.sender_sessions[1]|length)>0 and $e.sender_sessions[0]!=$e.sender_sessions[1]) and ($e.sender_generations[1] > $e.sender_generations[0]) and (($e.first_send_id|type)=="string" and ($e.first_send_id|length)>0) and $e.control_claimed==1 and $e.rows>=1 and $e.claimed_ids>=1 and ((($e.outcome=="reproduced") and $e.rows>=2 and $e.claimed_ids>=2 and $e.second_send_result=="new-id") or (($e.outcome=="not-reproduced") and $e.rows==1 and $e.claimed_ids==1 and $e.second_send_result=="same-id"))' .moai/reports/t1100/ac020.jsonl
```

### AC-DHR-021 — 복구 진입점과 doctor 읽기 전용 (REQ-DHR-004)

**Given** 미완료 저널 항목 하나와 참조 없는 `.codexwiring-*` 임시 파일 하나가 있는 임시 프로젝트 네 벌,
**When** 각각 `moai tool enable codex`, `moai tool disable codex`, update 경로의 배선 갱신, `moai doctor`를 실행하면,
**Then** 앞의 세 명령은 새 쓰기 전에 복구를 실행해 저널이 완료되고 임시 파일이 없어진다. `moai doctor`는 미완료 항목과 임시 파일, 복구 명령을 보고하고, 실행 전후 프로젝트 트리의 모든 파일 해시가 같다.

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestCodexWiringRecoveryEntryPoints$' -count=1 -timeout=600s > .moai/reports/t1100/ac021.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexWiringRecoveryEntryPoints")]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestCodexWiringRecoveryEntryPoints/(enable|disable|update|doctor_readonly)$")))]|length)==4 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac021.jsonl
```

### AC-DHR-022 — 배선 잠금: update 안과 단독 실행 (REQ-DHR-002)

**Given** 배선된 임시 프로젝트,
**When** (a) `moai update` 경로가 update 잠금을 쥔 채 배선 갱신을 호출하고, (b) `moai tool enable codex`를 단독으로 실행하고, (c) 살아 있는 다른 프로세스가 배선 잠금을 쥔 상태에서 update 경로를 실행하고, (d) 죽은 pid의 배선 잠금이 남은 상태에서 실행하면,
**Then** (a)와 (b)는 배선 파일이 갱신되고 저널이 완료된다. (c)는 배선 파일이 바이트 그대로이고 update 출력에 배선이 갱신되지 않았다는 경고가 있다. (d)는 죽은 잠금이 정리되고 갱신된다.

변이: 배선 잠금 대신 update 잠금을 다시 잡는 변이는 (a)에서 갱신이 일어나지 않으므로 FAIL해야 한다.

```bash
mkdir -p .moai/reports/t1100 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestCodexWiringLockUnderUpdateLock$' -count=1 -timeout=600s > .moai/reports/t1100/ac022.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexWiringLockUnderUpdateLock")]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestCodexWiringLockUnderUpdateLock/(update_holds_update_lock|standalone_enable|lock_held_by_live_owner|dead_owner_lock)$")))]|length)==4 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1100/ac022.jsonl
```

### AC-DHR-023 — [LIVE] 감사 역할 쓰기 차단과 부모가 쓴 판정 파일의 원문 일치 (REQ-DHR-015)

AC-DHR-012와 같은 실행(같은 jsonl, 같은 호출 예산 14회)의 (ii) 두 호출에서 나온 증거를 판정한다. 추가 모델 호출은 없다.

**Given** AC-DHR-012 (ii)의 부모 lane 오케스트레이터 세션 두 개와 각 세션의 Codex 세션 기록,
**When** `plan-auditor`와 `sync-auditor` 하위 에이전트가 판정문을 반환하고 부모가 판정 파일을 기록하면,
**Then** 같은 테스트가 쓴 증거 파일 `ac023-evidence.json`(§A 증거 채널, 태그 줄 `AC023_EVIDENCE_SHA256 <hex>`)의 두 항목마다 역할이 `plan-auditor`와 `sync-auditor` 각각 하나이고, 감사 역할의 쓰기 시도가 거부되었으며(AC-DHR-012의 쓰기 시도와 같은 실행), 판정 파일이 있고, 세션 기록에서 꺼낸 하위 에이전트 반환문의 sha256(64자 hex)과 판정 파일의 sha256이 같으며, 반환문에 그 실행의 nonce가 들어 있다.

음성·변이(판정식이 `false`, 합성 입력으로 확인): 반환문 해시와 판정 파일 해시 불일치, 같은 역할 두 번, 빈 해시끼리 같음, 실행 뒤 수정된 증거 파일(태그 해시 불일치).

세션 기록에 하위 에이전트 반환문이 남지 않으면 테스트는 출력에 `NOT_RUN`을 찍지 않고 `ac023-evidence.json`에 `"not_run": true`만 기록하며(같은 jsonl을 읽는 AC-DHR-012 판정식이 이 사유로 `false`가 되지 않게 한다), 이 AC는 `NOT_RUN`이다(반환문을 얻을 수 없으면 원문 일치를 판정하지 않는다). 실행 명령은 AC-DHR-012의 실행 명령이다(그 명령이 `ac023-evidence.json`도 먼저 지운다).

판정:

```bash
shasum -a 256 .moai/reports/t1100/ac023-evidence.json > .moai/reports/t1100/ac023-evidence.sha && jq -se --rawfile sha .moai/reports/t1100/ac023-evidence.sha --slurpfile ev .moai/reports/t1100/ac023-evidence.json '($sha|.[0:64]) as $h | ($ev[0]) as $e | ([.[]|select(.Action=="pass" and .Test=="TestCodexRoleLiveLoadAndReadOnly")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^AC023_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^AC023_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and (($e.audits|length)==2) and ([$e.audits[].role]|sort)==["plan-auditor","sync-auditor"] and ([$e.audits[]|select(.write_denied==true and .verdict_file_exists==true and ((.returned_sha256//"")|test("^[0-9a-f]{64}$")) and .returned_sha256==.verdict_file_sha256 and .returned_contains_nonce==true)]|length)==2' .moai/reports/t1100/ac012-live.jsonl
```

## §C 설계 기준별 판정 집계

| 설계 기준 | 결정적 AC | LIVE AC | PASS 조건 |
|---|---|---|---|
| AC-MIG-01 | AC-DHR-001 ~ 005, 021, 022 | 없음 | 일곱 AC 모두 `true` |
| AC-WT-01 | AC-DHR-006 ~ 009 | 없음 | 네 AC 모두 `true`. AC-DHR-006의 Windows 분기는 CI 관측 전까지 `NOT_RUN`으로 따로 적음 |
| AC-AGENT-01 | AC-DHR-010, 011, 013 | AC-DHR-012, 023 | 결정적 셋 `true` + LIVE 둘 `true`. LIVE가 `NOT_RUN`·`ABORTED`이면 `PARTIAL`이며 PASS 아님. `UNSUPPORTED` 항목은 PASS 집계에서 뺀 채 따로 나열 |
| AC-MSG-01 | AC-DHR-014 ~ 016, 020 | 없음 | 네 AC 모두 `true`. AC-DHR-014는 공통 명령과 측정 결과에 맞는 분기 명령 둘 다 `true` |
| AC-FACT-01 | AC-DHR-017, 019 | AC-DHR-018 | 결정적 둘 `true` + LIVE `true`. LIVE가 `NOT_RUN`·`ABORTED`이면 `PARTIAL` |

## §D 품질 게이트와 완료 정의

- 변경 패키지 단위 테스트만 로컬에서 실행한다(`internal/codexwiring`, `internal/factorymsg`, `internal/template/agentemit`, `internal/cli/worktree`, `internal/cli`의 지정 이름). 전체 스위트 판정은 CI 몫이다.
- `go vet`과 `golangci-lint run`을 변경 패키지에 대해 실행해 0건이어야 한다.
- `agents-codex.yaml`이나 `internal/template/templates/.claude/agents/moai/*.md`를 고치면 `make agents-emit`을 실행하고, `.codex/agents/moai/*.toml`을 손으로 고치지 않는다.
- factorymsg 스키마를 바꾸는 경우(분기 A)만 기존 DB 파일을 여는 마이그레이션 테스트를 포함한다(기존 행 보존, `SchemaVersion` 증가). 분기 B에서는 스키마가 그대로임을 AC-DHR-014 분기 B 명령이 확인한다.
- 완료 정의: §C 표의 다섯 기준 판정과 그 근거 파일 경로가 `progress.md` §E.2에 기록되고, LIVE 항목은 실행했으면 증거 경로, 안 했으면 `NOT_RUN`, 예산으로 멈췄으면 `ABORTED`로 적힌다. `PARTIAL`을 PASS로 적지 않는다.
- plan 단계에서 정한 기록 의무(plan-audit iter-3 N3): 증거 디렉터리 `.moai/reports/t1100/`는 gitignore 대상이라 워크트리와 함께 사라지므로, AC-DHR-020의 측정 결과(`outcome`), 측정한 커밋 SHA(`ac020-head.txt`의 값), 증거 파일의 sha256을 경로가 아니라 값으로 `progress.md` §E.2에 직접 적는다.
