# REQ-CSN-003 문면 정정 — 예산 근거 재측정 (manager-spec 배치, 2026-09-01)

측정 트리: `.claude/worktrees/t196` (branch `WT-codex-skill-neutral`, HEAD `c284ea0fb`).
모든 값은 이 트리에서 이 배치 안에서 잰 값이다. 리드 프롬프트가 준 수치(822/391 바이트)는
운반하지 않았다 — 아래는 전부 재측정값이다.

## 1. 가드 위치와 공식 (리드 지목 경로 정정)

리드가 `internal/hook/token_budget_guard.go` 로 지목했으나 그 파일은 이 트리에 없다.
실제 위치는 **`internal/config/token_budget_guard.go`** 다 (spec.md §F 가 이미 이 경로를
인용하고 있었다):

- `AlwaysLoadedTokenBudget = 76000` — `token_budget_guard.go:32`
- `estimateTokens` = `len(b) / 4` — `token_budget_guard.go:105-107`
- 고정 표면 슬롯 4개(CLAUDE.md · **AGENTS.md** · moai.md · MEMORY.md) — `token_budget_guard.go:195-200`
  — AGENTS.md 는 always-loaded 측정 표면이므로 결속표 추가는 이 예산을 소비한다.
- `CodexContractByteCeiling = 24576` — `token_budget_guard.go:41` (바이트 상한 — 별개 축)

## 2. 현재 여유 — 가드 본인의 측정

```
$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v -count=1
=== RUN   TestAlwaysLoadedTokenBudget
    token_budget_guard_test.go:69: always-loaded surface = 75799 tokens (budget 76000, headroom 201, 18 entries)
--- PASS: TestAlwaysLoadedTokenBudget (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.452s
```

**여유 201 tokens × 4 bytes/token = 804 bytes.** AGENTS.md 에 804 bytes 를 넘겨 넣으면
가드가 트립한다 (75,799 + 추가/4 > 76,000).

## 3. 후보 표 두 형태 — 3열 마크다운, 이 트리에서 재구성해 측정

후보 형태: 헤더 2줄 + 행들. 행 = `| 능력(중립 이름) | Claude 구현 | 이 하네스에 없을 때 행동 |`.
11행 형태는 `tool_classes` 값 10개 + 신설 `question-channel`; 4행 형태는 코덱스에 없는
능력만(question-channel · task-list · design-sync · cross-session-messaging).
표 텍스트 전문: `csn003-table-11row.txt` · `csn003-table-11row-honest.txt` · `csn003-table-4row.txt`.

```
$ wc -c .moai/reports/t196/csn003-table-11row.txt .moai/reports/t196/csn003-table-11row-honest.txt .moai/reports/t196/csn003-table-4row.txt
     797 .moai/reports/t196/csn003-table-11row.txt
     814 .moai/reports/t196/csn003-table-11row-honest.txt
     373 .moai/reports/t196/csn003-table-4row.txt
```

가드 산술 (75,799 + bytes/4 vs 76,000):

| 형태 | bytes | tokens (len/4) | 합계 | 판정 |
|---|---|---|---|---|
| 11행 — 최소 구성(부재 행동 칸을 가장 짧게) | 797 | 199 | 75,998 | 여유 **2 tokens** — 트립 안 한다 |
| 11행 — 정직 구성(보유 능력 칸에 "none needed — present here") | 814 | 203 | 76,002 | **가드 트립** (2 tokens 초과) |
| 4행 — 부재 능력만 | 373 | 93 | 75,892 | 여유 **108 tokens** |

## 4. 판정

- **인벤토리 형태(11행)는 문구와 무관하게 여유의 끝에 붙는다** — 같은 11행이 17바이트
  문구 차이로 2-under(797 B)와 2-over(814 B)를 갈라놓는다. 리드가 측정했다던 822 B 는
  이 배치의 구성에서 재현되지 않았다(문구 의존값 — 특정 셀 문구의 길이지 형태의 속성이
  아니다). 결정은 바이트 수에 기대지 않는다: **201 tokens 여유가 먼저고, 인벤토리 형태는
  그 여유 전부를 문구 운에 맡겨 소비하며, 부재 형태만이 설계로 여유를 남긴다.**
- 4행 = 373 B = 93 tokens. 리드 수치(391 B, ~104 tokens 남음)와 같은 방향, 다른 소수점 —
  행 구성과 셀 문구가 다르기 때문이다. 본 SPEC 에 적힌 값은 이 재측정값이다.

## 5. 셀렉터 재측정 (AC-CSN-009 정정 근거)

```
$ go test ./internal/template/ -run 'TestSkillDirToken' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.477s [no tests to run]
(exit 0 — 0매치 셀렉터가 PASS 를 냈다)

$ go test ./internal/template/ -run 'TestSkillTreeHasNoClaudeSkillDirToken' -v -count=1
=== RUN   TestSkillTreeHasNoClaudeSkillDirToken
    skill_dir_token_guard_test.go:94: census: 0 lines carrying CLAUDE_SKILL_DIR across 255 files walked under .claude/skills
--- PASS: TestSkillTreeHasNoClaudeSkillDirToken (0.01s)
PASS
(1매치)
```

함수 위치: `internal/template/skill_dir_token_guard_test.go:41`.

## 6. tool_classes 값 집합 — §A.5 정정 근거

```
$ awk '/^tool_classes:/{f=1;next} /^[^ ]/{f=0} f&&NF==2{gsub(":","",$2);print $2}' internal/template/agentemit/agents-codex.yaml | sort -u
cross-session-messaging
design-sync
file-read
file-write
moai-mcp
shell
skill-loader
subagent-spawn
task-list
web
(정렬 없이: sort -u | wc -l → 10)
```

**10개.** spec.md §A.5 종전 목록(9개)은 `cross-session-messaging` 을 빠뜨린 누락이다.

## 7. 패리티 등록 확인 (plan M2/M3 메모 근거)

`internal/template/agentemit/agents-codex.yaml`:
- `workflowOptMirroredPaths` (`rule_template_mirror_test.go:42-98`) — **미등재**
- `lateBranchMirroredPaths` (`:109-127`) — **미등재**
- yaml 헤더 자체 선언: "Build input only: this file lives in the emitter package, NOT
  under templates/, and is never distributed to user projects."
- → **단일 트리 편집.** 같은 커밋에 cp 할 미러가 없다.

대조(등재 파일): `.claude/rules/moai/workflow/worktree-integration.md` — `:57` 등재.
`skill-authoring.md` 는 어느 목록에도 없고 두 사본은 실측상 DIFFER (sanitized pair 관리 밖).

## 8. 매니페스트 일관성 강제 (11번째 클래스 편집 형태)

`internal/template/agentemit/manifest.go:121-124` — 매핑 값이 처분(`classes:`) 행 없이는
로드 실패. `:112-118` — 처분 행은 유효한 disposition + 비어 있지 않은 rationale 를 요구한다.
즉 11번째 클래스는 **매핑 행 + 처분 행(rationale 이 클래스의 커버 범위를 서술)** 이
한 편집에 함께 들어가야 한다.
