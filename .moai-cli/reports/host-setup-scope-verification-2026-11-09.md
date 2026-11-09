# host-setup-scope 보고 독립 재검증 (2026-11-09, agent-1 [0dd806])

- 검증 대상: `.moai-cli/reports/host-setup-scope-2026-11-09.md` (리더 작성, 72줄)
- 검증자: agent-1 [0dd806], 세션 `113e4ddd-566b-4bd7-b27d-f6cc8c19dcf1`
- 측정 시각: `date -u` → `2026-11-09T02:57:50Z`
- 기준 트리: `git fetch origin v2026.11.0`(exit 0) 후 `git rev-parse origin/v2026.11.0` = `d76a5b7cedd235e0c621ffef8c6768aab4baf2d0` (FETCH_HEAD와 동일). `git archive d76a5b7ce src docs adapters skills/moai`로 세션 scratch에 풀어 읽음. 주 체크아웃·원격에는 아무것도 쓰지 않음.
- 방법: 소스 읽기와 grep/awk 계수만. 바이너리 빌드·실행은 하지 않음(아래 Gaps).

## 0. 기준 트리 차이

리더 보고는 기준을 `6ee819cae`로 적었고 현재 tip은 `d76a5b7ce`다(9커밋 앞, 아래 확인 결과 인용 파일 변화 없음).

```
$ git merge-base --is-ancestor 6ee819cae d76a5b7ce && echo ancestor   → ancestor
$ git rev-list --count 6ee819cae..d76a5b7ce                            → 9
$ git diff --stat 6ee819cae d76a5b7ce -- src/ docs/ adapters/ skills/moai/SKILL.md appendix/
 docs/protocol-v1.md | 13 +++++++++++++
 src/claim.rs        | 26 ++++++++++++++++++++++++++
 src/queue.rs        |  4 ++++
```

인용된 파일(claude.rs, hosts/codex.rs, update.rs, agents.rs, sources.rs, generation.rs, session.rs, main.rs, docs 계약·지원표)은 두 커밋 사이에 바뀌지 않았으므로 줄 번호 대조는 그대로 유효하다. 주 체크아웃 차이: `git rev-list --count HEAD..6ee819cae` = **491**(리더 수치와 일치), 지금 tip 기준 `HEAD..d76a5b7ce` = 500(= 491 + 9). 불일치가 아니라 기준 시점 이동이다.

## 1. 행별 판정

판정: **확인됨** / **불일치(실제 값)** / **확인 불가**

### 1.1 중점 항목

| # | 주장 | 판정 | 근거 (tip d76a5b7ce) |
|---|---|---|---|
| F1 | Claude 훅 0개: settings 키 `hooks` 비관리 (claude.rs:266-269) | 확인됨 | 266-268 주석 "The RC1 hook set is empty (P04 = A), so `hooks` is not managed", 269 `MANAGED_SETTINGS: [&str; 2] = ["outputStyle", "statusLine"]` |
| F2 | skill frontmatter `hooks` 거부 (claude.rs:577-578) | 확인됨 | 577-578 "Fields that grant tools or register hooks (`allowed-tools`, `hooks`, ...) are not offered", 579-589 `CLAUDE_FIELDS` 9개에 hooks·allowed-tools 없음 |
| F3 | Codex 훅 0개 (codex.rs:191) | 확인됨 | 191 `"hooks": {"generated": [], "reason": "RC1 interim decision: no Codex hooks are generated"}`; 16-18 모듈 주석 "no hooks are generated"; docs/codex-adapter-contract.md:10 "0개. 아무것도 생성하지 않는다" |
| F4 | settings.json 쓰기는 라이브러리만, CLI 명령 없음 | 확인됨 | claude.rs `pub fn settings_plan`(431)·`settings_apply`(507) 존재. `grep -rn 'settings_plan\|settings_apply\|plan_settings\|claude::projection\|claude::doctor' src`(claude.rs 제외) 결과는 `src/main.rs:949 ... claude::doctor(&project)` 1건뿐. OPERATIONS 49쌍 중 claude 관련은 `("claude-doctor", "claude.doctor")` 하나. docs/claude-adapter-contract.md:61 "no CLI operation carries it yet" |
| F5 | config.toml 미구현 (TOML 파서 없음, HOLD) | 확인됨 | codex.rs:16-18, 31-32 `TOML_HOLD` "not read: no TOML parser in RC1", 175-183 요청 시 `"status": "hold" ... "writes": []`; main.rs:318,321 "config.toml is not read"; `grep -rnE 'hooks\.json|config\.toml|\.mcp\.json|AGENTS\.override' src`에 쓰기 경로 없음(읽기·존재 확인·주석만) |
| F6 | AGENTS.override.md 감지만, 생성 안 함 (codex.rs:274-276,346) | 확인됨 | 274-276 global 탐색 이름 `["AGENTS.override.md","AGENTS.md"]`, 289 프로젝트 체인도 동일, 346 `"agents_override_created": false`; agents.rs:365-373은 존재 시 "present; preserved" 경고만 |
| F7 | update.rs 홈 skill 복사 경로 (update.rs:3-5,51-61) | 확인됨 | 3-5 "Claude `~/.claude/skills/moai/`, Codex `~/.agents/skills/moai/` ... project-level projections are a P04 HOLD", 51-57 `skills_dir`: claude→`.claude/skills`, codex→`.agents/skills`, 59-61 `home.join(skills_dir(host)).join("moai")`; 43 `HOSTS = ["claude","codex"]`. 단순 복사이며 심링크는 blocked(19-22) |

### 1.2 Claude Code 표

| 행 | 판정 | 근거 |
|---|---|---|
| AGENTS.md managed 구간 (agents.rs:16,161) | 확인됨 | 16 `AGENTS_SECTION = "moai-workflow"`, 161 `fn render_agents`·162 `markers(AGENTS_SECTION, "html")` |
| v2.1.277+·CLAUDE.md 변형 조건, doctor 예측 (claude.rs:136-264) | 확인됨 | 29 `MIN_AGENTS_VERSION = (2, 1, 277)`, 136 `pub fn doctor`, 169 `"minimum_version": "2.1.277"`, 177-178 `"basis": "predicted"`, `"host_loaded": "not_run"` |
| CLAUDE.md `@AGENTS.md` 제안만 (claude.rs:247-250) | 확인됨 | 246 "Proposals only (REQ-02B-05): nothing is applied", 248 "add an `@AGENTS.md` import to the CLAUDE.md you keep" |
| 기존 CLAUDE.md 미수정 (agents.rs:365-373) | 확인됨 | 366-369 CLAUDE.md·.claude/CLAUDE.md·CLAUDE.local.md·AGENTS.override.md 존재 시 "preserved" |
| `moai` skill 홈 복사 (update.rs:3-5,51-61) | 확인됨 | F7 |
| Claude 전용 frontmatter: 라이브러리만 (claude.rs:576-589,632) | 확인됨 | 632 `pub fn projection(...)`, 631 "Never writes."; OPERATIONS에 claude projection 항목 없음 |
| 외부 skill (sources.rs:506-515) | 확인됨 | 506 `host_skill_dir`, 507-510 codex→`.agents/skills` 그 외→`.claude/skills`, 512 project는 `root`, 513 그 외는 `home` |
| 생성 skill 투영 (generation.rs:575-578) | 확인됨 | 577 `.agents/skills/{id}/SKILL.md`, 578 `.claude/skills/{id}/SKILL.md` |
| role 파일 (generation.rs:342-354,516-518) | 확인됨 | 349 `name`/`description`/선택 `tools`, 517 `.claude/agents/{role_id}.md`; `request.native_roles`에 있을 때만 생성(514) |
| settings.json outputStyle·statusLine, 영수증 (claude.rs:266-269,429-570) | 확인됨 | 269 키 2개, 320 `SETTINGS_RECEIPT = "manifests/claude-settings.json"`, 429-431 "Never writes"(plan), 571-572 `host_loaded`/`execution_verified` not_run |
| output-styles·statusline: Python만, Rust 미호출 | 확인됨 | `adapters/claude/install.py` 존재; `grep -rn 'install\.py\|adapters/claude' src` 결과 0건; docs/claude-adapter-contract.md:62 "not called by the Rust product" |
| hooks·permissions·MCP 미구현, "런처가 `--dangerously-skip-permissions`만 전달" | **불일치(부분)** | 미구현 부분은 확인됨(F1·F5 grep). 런처 서술은 실제와 다름: main.rs:615 `skip_permissions: role.is_some() && host != "codex" && !o.permissions`(main.rs:75 `permissions: bool`), session.rs:151-154는 `r.skip_permissions`일 때만 추가하고, 17-19 "Without a role ... `--dangerously-skip-permissions` is never added", docs/rust-product-path.md:446 "`--permissions` drops the permission skip". 실제 값 = **역할 세션(-l/-a)에서 기본으로 추가, `--permissions`로 끔, 역할 없는 일반 세션에는 추가 안 함**. 또 같은 argv에 `--name --model --effort --advisor`도 붙으므로(session.rs:6, rust-product-path.md:446) "만"은 권한 관련 플래그에 한정해 읽어야 맞다 |

### 1.3 Codex 표

| 행 | 판정 | 근거 |
|---|---|---|
| AGENTS.md 32 KiB 경고 (agents.rs:405) | 확인됨 | 403-405 `projected > 32 * 1024` → "may exceed the 32 KiB default instruction budget of Codex"; codex.rs:29 `DEFAULT_MAX_BYTES = 32 * 1024` |
| AGENTS.override.md 설계만 | 확인됨 | F6; docs/design-review/personal-coding-core-2026-11-09.md §3.1(29행)·§3.2(34행) 제목 존재(본문 대조는 하지 않음) |
| `moai` skill → ~/.agents/skills/moai (update.rs:54) | 확인됨 | 54 `"codex" => ".agents/skills"` |
| agents/openai.yaml 계획만 (codex.rs:56-67,128-201) | 확인됨 | 56 `pub fn openai_yaml`(문자열 렌더), 128 `pub fn projection_plan`, 194 "plan only: nothing is written; installing into a user home is EP-05" |
| 외부·생성 skill (sources.rs:508, generation.rs:577) | 확인됨 | 위 1.2 |
| role 파일 .toml (generation.rs:356-363,520-522) | 확인됨 | 358 `name`/`description`/`developer_instructions`, 521 `.codex/agents/{role_id}.toml` |
| config.toml 미구현 (codex.rs:16-18,31-33,175-183) | 확인됨 | F5 |
| hooks·MCP 미구현, doctor는 존재만 확인 (codex.rs:191,350-356) | 확인됨 | 351 `config_toml: ...is_file()`, 352 `hooks_json: ...is_file()`, 353-354 not_evaluated |
| 실행 플래그 `-s workspace-write -a on-request` | 확인됨 | session.rs:13, 158 `fn codex_argv`, 163·165, 176 "codex runs with -s workspace-write -a on-request" |
| 실세션 검증 not_run (support-matrix.md:5,26-29) | 확인됨 | 5 상태 한계, 20 정적 `claude plugin validate` 1행, 26-29 AC-08C-3/6/4/8(T2) not_run |

### 1.4 §3–§5 서술

| 주장 | 판정 | 근거 |
|---|---|---|
| host-support-matrix.md "MoAI RC1 hook은 0개" | 확인됨 | docs/host-support-matrix.md:72 "MoAI RC1 hook은 0개(P04 = A)" |
| codex-adapter-contract.md:10 | 확인됨 | 위 F3 |
| rust-product-path.md:446 "later Issue" | 확인됨 | 446 원문 끝: "The skip is the operator's interim default (2026-11-08) and lets every tool call run without a prompt; a settings key is a later Issue." |
| decisions-2026-11-08.md:115-118 EP-H 근거 | 확인됨(행 존재) | 115 PERSONAL_CODING_CORE 수신·`AGENTS.override.md`, 117 override 학습 루프 승인 수준. H1/H2/H5 명칭은 personal-coding-core §8(76행 "영향 카드와 신규 카드(EP-H, 미발행)")에 있음 |
| 설계 textonly 행 번호(154-160, 166, 293-318, 438-442, 555-575, 742) | **확인 불가** | 해당 파일을 tip 트리와 `<home>/MoAI/moai-rc1-handoff-ref`(로컬 인계 참조 디렉터리)(depth 4, `*text*`)에서 찾지 못함 |
| rc.2 카드 번호·제목(#2160, #2163, #2164, #2170, #2176)과 "카드가 없는 갭" 8개 | **확인 불가** | Issue 원장(GitHub) 조회는 이번 과제 범위 밖이라 not_run |

## 2. 추가 측정: OPERATIONS 개수

### 2.1 `src/main.rs` (tip d76a5b7ce)

`const OPERATIONS: &[(&str, &str)] = &[`는 638행. 두 가지 방법으로 셈:

```
$ awk 'NR>=638 && /^\];/{exit} NR>638' src/main.rs | grep -cE '^\s*\("[a-z-]+", "[a-z_.]+"\),$'
49
$ awk 'NR>638{ if ($0 ~ /^\];/) exit; print }' src/main.rs | grep -o '("' | wc -l
      49
$ awk 'NR>638{ if ($0 ~ /^\];/) exit; n++ } END{print n}' src/main.rs     # 블록 전체 줄 수
49
```

**main.rs OPERATIONS = 49쌍.** 마지막 항목은 `("parity-check", "parity.check")`.

### 2.2 `docs/design-review/appendix/interfaces.md` (tip, 마지막 변경 07602b74e 2026-11-09T10:39:52+09:00)

문서 자체 표기: 3행·20행 "tip `604792666` 47개 + rc.2 A2 `queue-next` = **48개**". 목록(21행부터 `(추출` 줄 전까지)의 CLI 단어를 셈:

```
$ awk 'NR>=21 && /^\(추출/{exit} NR>=21' interfaces.md | grep -oE '`[a-z][a-z-]*`' | tr -d '`' | sort > doc_words.txt
$ wc -l < doc_words.txt           → 48
$ sort -u doc_words.txt | wc -l   → 48
$ comm -23 main_words.txt doc_words.txt   # main.rs에만 있음
parity-check
$ comm -13 main_words.txt doc_words.txt   # 문서에만 있음
(없음)
```

**interfaces.md 목록 = 48개.** 차이 1건은 `parity-check`로, `git log -S'"parity-check"' -- src/main.rs` → `6588aa2c5 feat(parity): semantic parity check op - template vs generated vs doc (D6 #2192)`에서 들어왔다. `6588aa2c5`(2026-11-09T10:38:21+09:00)는 interfaces.md 마지막 변경 `07602b74e`(10:39:52+09:00)의 조상이 아니다(`git merge-base --is-ancestor` 실패) — 두 변경이 서로 다른 브랜치에서 진행돼 문서가 `parity-check` 병합 전 상태를 기준으로 쓰였다. 문서 갱신 후속 카드의 근거로 쓸 수 있다.

## 3. 요약

- 대조한 주장: 중점 7건 + Claude 표 12행 + Codex 표 10행 + §3–§5 서술 6건
- **불일치 1건 (부분, 경미)**: 런처 서술 "`--dangerously-skip-permissions`만 전달" → 실제는 역할 세션(-l/-a)·Claude 계열 호스트에서만 기본 추가, `--permissions`로 끔, 일반 세션은 추가 안 함(main.rs:615, session.rs:17-19,151-154). 훅·permissions·MCP 미구현이라는 결론은 그대로 성립.
- 기준 트리 이동(6ee819cae → d76a5b7ce, 9커밋)은 불일치가 아니다: 인용 파일 변화 없음, 491은 리더 기준에서 정확.
- **확인 불가 2건**: 설계 textonly 행 번호(파일 미발견), rc.2 카드 목록·갭 8개(Issue 조회 not_run).
- 나머지는 모두 확인됨. 핵심 결론(훅 0개, settings CLI 없음, config.toml 미구현, AGENTS.override.md 미생성, 홈 skill 경로)은 tip에서 그대로 성립한다.
- 측정값: main.rs OPERATIONS **49**, interfaces.md **48** (누락 `parity-check`).

## Gaps (미검증)

- 바이너리를 빌드하거나 실행하지 않았다. 모든 판정은 소스·문서 읽기와 grep/awk 계수에 기반한다(§2.2 도구 출처 대상 아님 — 프로젝트 도구로 측정한 값 없음). 런타임 동작(예: update-apply가 실제로 홈에 쓰는지)은 not_run.
- 설계 textonly 문서, GitHub Issue 원장, personal-coding-core §3·§8 본문은 대조하지 않았다.
- 거부된 도구 호출 없음.

## Residual-risk

- grep은 문자열 기반이므로, 매크로나 간접 경로로 settings/hooks 쓰기를 호출하는 코드가 있다면 놓칠 수 있다(호출자 grep은 함수명 기준).
- interfaces.md 계수는 백틱 안의 `[a-z-]` 단어 기준이다. 목록 형식이 바뀌면 다시 세어야 한다. 독립 측정(중복 제거 수·main.rs와의 `comm` 차집합)이 서로 맞아 계수 누락 가능성은 낮다.
- tip은 계속 움직인다. 이 보고는 `d76a5b7ce` 시점의 값이다.
