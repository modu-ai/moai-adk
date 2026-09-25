# Progress — SPEC-DOCS-OTEL-AGENTS-001

card: t1181

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase 산출물(spec.md / plan.md / acceptance.md / progress.md)을 2026-09-25 manager-spec 이 브랜치 `WT-docs-otel-env`, base `a520187f1` 위에서 작성했다. 미커밋.
- 전제 실측(이 실행, 이 트리): `CLAUDE_CODE_ENABLE_TELEMETRY`·`OTEL_` 3행, `/agents` 행·문장의 `2026-07` 단서 5건, ja/zh `commands.md` 에 `### /agents` 절 없음, `commands#agents` 앵커 링크 0건 — spec.md §A.4.
- base Hugo 기준선: `hugo --source docs-site --minify --gc --destination <scratch>` → exit 0, `WARN` 0건, 이후 `git status --short` 빈 출력.
- Spec lint: 아래 §E.1.1.

### §E.1.1 Spec lint result

명령: `moai spec lint SPEC-DOCS-OTEL-AGENTS-001` (2026-09-25, 커밋 전, 트리 `a520187f1` + 미커밋 SPEC 디렉터리). 출력 원문: `✓ No findings — all SPEC documents are valid`, exit 0. 설치본 빌드와 이 트리의 조상 관계는 이번 실행에서 확인하지 않았다(Gap). 소유권 검사는 커밋 전에는 판정하지 못하므로 커밋 뒤 한 번 더 잰다.

양성 대조 실측(이 실행, base `a520187f1`): `git grep -c CLAUDE_CODE_ENABLE_TELEMETRY` → en/ja/zh 각 `:1`; `/agents`+`2026-07` → en/ja/zh commands.md·ja/zh sub-agents.md 각 `1`; `살펴보는 명령`·`a command to inspect`·`対話的に生成`·`交互式生成` 각 `1`; `/agents` 행의 `v2.1.281` → ko `0`. 보존 대상: ko/en description 의 `/agents` 각 `1`, `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH=1` 각 `1`.

### §E.1.2 iter-1 FAIL(0.74) 반영 — 0.2.0 개정

감사 보고서 `.moai/reports/plan-audit/SPEC-DOCS-OTEL-AGENTS-001-review-1.md` 의 D1~D11 을 반영했다(2026-09-25, manager-spec, 같은 트리 `a520187f1` + 미커밋 SPEC 디렉터리). Tier S→M 승격(대상 9개 파일, `harness.yaml:76` Tier S 는 5개 미만) — plan-audit 반복 상한 2, 통과 기준 0.80. AC 9개(Tier M 상한 16), REQ 7개.

재측정 방법: base 아홉 파일을 `git archive a520187f1 <9 paths>` 로 스크래치에 두 벌 풀고, 한 벌에 plan.md M1 의 AFTER 를 적용해(내용 기준 행 치환) 개정 AC 명령을 두 벌에 모두 돌렸다. docs-site 작업 트리는 건드리지 않았다.

| 검사 | base | AFTER 적용본 |
|---|---|---|
| D1: `grep '/agents' <f> \| grep -c 'v2.1.198'` ja / zh `sub-agents.md` | `0` / `0` | `1` / `1` |
| (참고) ja `sub-agents.md` 파일 전체 `grep -c 'v2.1.198'` | `4` | `5` |
| D2: 파일별 추가·삭제 행 수(`diff` 로 계산) en/ja/zh settings-json | `—` | 각 `0 1` |
| 같은 방법 ko/en commands · ja/zh commands · ja/zh sub-agents | `—` | `3 3` · `1 1` · `1 1` |
| D3(a) 옛 제목 ko / en | `1` / `1` | `0` / `0` |
| D3(b) 옛 서술자 ko / en / ja / zh | 각 `1` | 각 `0` |
| D3(c) AFTER 10블록 행 전체 일치(`grep -cxF -e "$(after <key>)"`) | 10개 모두 `0` | 10개 모두 `1`, 블록마다 추출 행 수 `1` |
| 표 행 `v2.1.281` / `v2.1.198` / `/help` (네 로케일) | `0` / `1` / `0` | `1` / `1` / `1` |
| 인접 행 `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` / `CLAUDE_CODE_DISABLE_BACKGROUND_TASKS` (en/ja/zh) | `2` / `1` | `2` / `1` |

카드 커밋 특정 명령 형식은 앞 카드로 검증했다: `git log --no-merges --format=%H --grep=t1178 bfd902049..a520187f1 -- docs-site` → `112d5a3e5f3ea5ec8cf9d16bf61848b92e213f12` 한 줄(메시지에 `t1178` 을 담은 병합 커밋 `a520187f1` 은 배제됨). `git show --numstat --format= 112d5a3e5` 출력에 앞머리 빈 줄이 없음도 확인했다.

Spec lint(개정 뒤): `moai spec lint SPEC-DOCS-OTEL-AGENTS-001` → `✓ No findings — all SPEC documents are valid`, `exit=0`. 설치본 `moai-adk v3.2.0-rc.15`(`moai_cp/20260910_130400-3003-g372c1bb0b`, built 2026-09-24T08:57:47Z); 이 트리와의 조상 관계는 확인하지 않았다(Gap).

Gaps: AFTER 적용본은 run-phase 편집의 시뮬레이션이지 실제 편집이 아니다 — 카드 커밋 단위 numstat(AC-DOA-008)은 커밋이 생겨야 판정된다. Hugo 빌드는 이 개정에서 다시 돌리지 않았다(docs-site 를 건드리지 않았으므로 §E.1 의 base 기준선이 그대로다).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
