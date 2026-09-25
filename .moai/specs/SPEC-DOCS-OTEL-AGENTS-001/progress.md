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

2026-09-25 manager-develop 실행. 트리: 브랜치 `WT-docs-otel-env`, 카드 커밋 `353521256d7b05085d101ec35787bcf0af44ed60`(부모 = plan 커밋 `2b2e14c63`, base `a520187f1`). 커밋 직전 재확인: `git rev-parse --short HEAD` → `2b2e14c63`, `git branch --show-current` → `WT-docs-otel-env`.

적용 방식: plan.md M1 의 AFTER 10블록을 §A `after` 헬퍼와 같은 규칙(표지 → ```` ```text ```` → 한 줄)으로 기계 추출해, 대상 행을 내용 기준(행 번호 아님)으로 찾아 행 전체를 치환했다. 텔레메트리 행은 `| \`CLAUDE_CODE_ENABLE_TELEMETRY\` |` 로 시작하는 행 하나만 지웠다. 파일마다 적중 행이 정확히 1개임을 단언한 뒤 썼다. plan.md 문구와의 편차 없음.

### Hugo 빌드 (AC-DOA-009)

명령: `hugo --source docs-site --minify --gc --destination "$SCRATCH/hugo-after" > "$SCRATCH/hugo-after.log" 2>&1; echo "exit=$?"` (`$SCRATCH` = 세션 스크래치 디렉터리, 트리 밖). 커밋 전 1회, 커밋 후 1회 실행했고 둘 다 같은 결과였다.

```text
exit=0
grep -c WARN hugo-after.log → 0
grep -c ERROR hugo-after.log → 0
Pages KO 188 / EN 186 / JA 186 / ZH 186, Total in 4218 ms (커밋 전 실행 tail)
sitemap.xml 존재
```

### AC 배치 출력 원문 (이 실행, HEAD `353521256`)

모든 명령은 acceptance.md §D 의 형태 그대로이며, 워크트리 루트에서 한 스크립트로 실행했다(`B=a520187f1`, `$C` 는 acceptance.md §A 정의).

```text
HEAD=353521256d7b05085d101ec35787bcf0af44ed60 C=353521256d7b05085d101ec35787bcf0af44ed60
### AC-DOA-001
[control] git grep -c CLAUDE_CODE_ENABLE_TELEMETRY a520187f1 -- docs-site/content:
a520187f1:docs-site/content/en/advanced/settings-json.md:1
a520187f1:docs-site/content/ja/advanced/settings-json.md:1
a520187f1:docs-site/content/zh/advanced/settings-json.md:1
[then] grep -rn 'CLAUDE_CODE_ENABLE_TELEMETRY\|OTEL_' docs-site/content:
exit=1
[then] numstat:
0	1	docs-site/content/en/advanced/settings-json.md
0	1	docs-site/content/ja/advanced/settings-json.md
0	1	docs-site/content/zh/advanced/settings-json.md
### AC-DOA-002
ko control_old=1 base281=0 basehelp=0 | row281=1 row198=1 rowhelp=1 old_now=0
en control_old=1 base281=0 basehelp=0 | row281=1 row198=1 rowhelp=1 old_now=0
ja control_old=1 base281=0 basehelp=0 | row281=1 row198=1 rowhelp=1 old_now=0
zh control_old=1 base281=0 basehelp=0 | row281=1 row198=1 rowhelp=1 old_now=0
### AC-DOA-003
docs-site/content/en/claude-code/foundations/commands.md base=1 now=0
docs-site/content/ja/claude-code/foundations/commands.md base=1 now=0
docs-site/content/zh/claude-code/foundations/commands.md base=1 now=0
docs-site/content/ja/claude-code/agentic/sub-agents.md base=1 now=0
docs-site/content/zh/claude-code/agentic/sub-agents.md base=1 now=0
### AC-DOA-004
ko oldhead base=1 now=0
ko oldpara base=1 now=0
en oldhead base=1 now=0
en oldpara base=1 now=0
docs-site/content/ko/claude-code/foundations/commands.md -A2 v2.1.281 base=0 now=1
docs-site/content/en/claude-code/foundations/commands.md -A2 v2.1.281 base=0 now=1
### AC-DOA-005
ko 폴더에 마크다운 파일을 직접 만들기=1
en Create a markdown file directly under=1
docs-site/content/ko/claude-code/foundations/commands.md SPAWN_DEPTH=1 description_agents=1
docs-site/content/en/claude-code/foundations/commands.md SPAWN_DEPTH=1 description_agents=1
### AC-DOA-006
ja 対話的に生成 base=1 now=0
zh 交互式生成 base=1 now=0
docs-site/content/ja/claude-code/agentic/sub-agents.md /agents+v2.1.198 base=0 now=1
docs-site/content/zh/claude-code/agentic/sub-agents.md /agents+v2.1.198 base=0 now=1
### AC-DOA-007
ko-row lines=1 base=0 now=1
en-row lines=1 base=0 now=1
ja-row lines=1 base=0 now=1
zh-row lines=1 base=0 now=1
ko-head lines=1 base=0 now=1
ko-para lines=1 base=0 now=1
en-head lines=1 base=0 now=1
en-para lines=1 base=0 now=1
ja-sub lines=1 base=0 now=1
zh-sub lines=1 base=0 now=1
### AC-DOA-008
count=1
0	1	docs-site/content/en/advanced/settings-json.md
3	3	docs-site/content/en/claude-code/foundations/commands.md
0	1	docs-site/content/ja/advanced/settings-json.md
1	1	docs-site/content/ja/claude-code/agentic/sub-agents.md
1	1	docs-site/content/ja/claude-code/foundations/commands.md
3	3	docs-site/content/ko/claude-code/foundations/commands.md
0	1	docs-site/content/zh/advanced/settings-json.md
1	1	docs-site/content/zh/claude-code/agentic/sub-agents.md
1	1	docs-site/content/zh/claude-code/foundations/commands.md
msg_t1181=2
### AC-DOA-009
WARN=0
status_docs_site=[]
ls_remote=[]
```

보강: `git ls-remote --exit-code --heads origin WT-docs-otel-env; echo "exit=$?"` → `exit=2`(원격에 도달했고 일치 ref 없음 — 네트워크 실패가 아니라 push 부재).

### AC 판정표

| AC | 판정 | 근거(위 원문) |
|---|---|---|
| AC-DOA-001 | PASS | 대조 3파일 `:1`; 작업 트리 grep exit 1; numstat 세 파일 `0 1` |
| AC-DOA-002 | PASS | 네 로케일 옛 서술자 base 1 → 0; 표 행 `v2.1.281`·`v2.1.198`·`/help` 각 1 (base 281·help 0) |
| AC-DOA-003 | PASS | 다섯 파일 base 1 → 0 |
| AC-DOA-004 | PASS | 네 grep base 1 → 0; `-A2` `v2.1.281` base 0 → 1 (ko/en) |
| AC-DOA-005 | PASS | 보존 대상 4종 각 1 |
| AC-DOA-006 | PASS | 옛 문구 base 1 → 0; `/agents` 행 `v2.1.198` base 0 → 1 (ja/zh) |
| AC-DOA-007 | PASS | 10키 모두 추출 1행, base 0 → 1 |
| AC-DOA-008 | PASS | 카드 docs 커밋 1개; numstat 아홉 행이 기대 표와 일치; 메시지 `t1181` 2회 |
| AC-DOA-009 | PASS | exit 0, WARN 0, `git status --short docs-site` 빈 출력, ls-remote 빈 출력(exit 2) |

Gaps: Hugo `WARN` 검출기의 양성 대조 없음. 렌더된 HTML 을 눈으로 확인하지 않았다. 배포 부재는 push 부재로 대리 판정했다 — push 가 없으면 Vercel 빌드 트리거도 없다고 간주한다. `.moai/reports/t1181/verdict.md` 는 이 실행이 쓰지 않았다(리드 판정 영역).

Residual-risk: 표 행 버전 열 `v2.1.139+` 는 이 카드 범위 밖이라 검증·수정하지 않았다(plan.md §B). 커밋 후 SPEC lint(소유권 검사)는 이 progress 커밋 뒤에 다시 재야 한다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-25
run_commit_sha: 353521256d7b05085d101ec35787bcf0af44ed60   # docs-site card commit; this progress/status commit follows it
run_status: audit-ready
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 0   # no file under docs-site/ outside the nine targets changed
l44_pre_commit_fetch: not-run     # lane rule: no push; lead batch-pushes develop
l44_post_push_fetch: not-applicable
new_warnings_or_lints_introduced: 0   # hugo WARN 0
cross_platform_build: not-applicable  # docs-only, no Go change
total_run_phase_files: 11   # 9 docs-site files + spec.md status + progress.md
m1_to_mN_commit_strategy: "2 commits — docs-site card commit, then progress/status commit touching no docs-site file"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
