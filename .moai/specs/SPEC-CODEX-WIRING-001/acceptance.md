# Acceptance — SPEC-CODEX-WIRING-001

> 모든 기준은 이진 판정 가능: 명령 exit code, Go 테스트, 또는 기대 횟수를 명시한 grep.
> Given-When-Then으로 서술. 검증하는 GEARS 요구는 `spec.md` §D.

## Severity legend

- **MUST** — Definition of Done 차단.
- **SHOULD** — 비차단. 미달 시 사유와 함께 부채로 기록.

## §A. AC Matrix

> grep 계열 토큰의 base-0 실측(2026-08-23, 본 트리 origin/main @ 76b2c4ece):
> `default_tools_approval_mode` 0 · `Flags().String("agent"` 0 · `checkCodexWiring` 0 ·
> `package codexwiring` 0 · `"Codex Wiring"` 0 · `/hooks to re-trust` 0 · `codex /hooks` 0 ·
> `mcp_servers`(Go 코드) 0 — 전부 채택. `/hooks` 단독은 기존 문구 충돌로 채택 제외(히트수는
> 범위 의존 — 귀속 명령 `grep -rn 're-approve\|/hooks' internal/cli/*.go | grep -v _test | wc -l`
> → 10, 글롭형·비재귀; 재귀 범위를 넓히면 수치가 달라진다).
> **v0.3.0 추가 실측(2026-08-24, rebase 후 트리 @ 915c310de)**: `statusLineAllowlist` 0 ·
> `17827`(README.md·README.ko.md·README.en.md·README.ja.md·README.zh.md·docs-site/content) 0 —
> 채택. `status_line` 단독은 internal/ Go 46hit(Claude statusline 코드 충돌)으로 채택 제외 —
> 저장소 코드 토큰은 `statusLineAllowlist`를 쓰고, `status_line` grep은 스크래치 산출물
> 파일 대상으로만 사용한다.
> 스크래치 e2e는 전부 `$(mktemp -d)` 하부 + 필요시 `CODEX_HOME=$(mktemp -d)` 격리.

### M1/M2 — 플래그·생성

**AC-CW-001** (MUST, REQ-CW-001) — *Given* 빌드된 moai 바이너리, *When*
`moai init --help`를 실행하면, *Then* `--agent` 플래그가 3값(claude, codex, both)을 문서한다
(`moai init --help | grep -c -- '--agent'` ≥ 1). *And Given* 무효값, *When*
`moai init /tmp/cw-invalid --agent gemini --non-interactive`를 실행하면, *Then* exit code ≠ 0
이고 stderr가 유효값 집합을 나열한다.

**AC-CW-002** (MUST, REQ-CW-002 · REQ-CW-003) — *Given* 스크래치 디렉터리, *When*
`moai init proj --agent codex --non-interactive`를 실행하면, *Then* `proj/.codex/hooks.json`이
존재하고: (a) 이벤트 키 집합이 정확히 `{PreToolUse, PostToolUse, SessionStart, SessionEnd, Stop,
UserPromptSubmit}` (jq: `jq -r '.hooks | keys_unsorted | sort | join(",")'` 비교);
(b) 모든 handler command가 `moai hook `로 시작하고 `--harness codex`를 포함 —
`grep -c -- '--harness codex' proj/.codex/hooks.json` 값이 handler 총수
`jq '[.hooks[][].hooks[]] | length' proj/.codex/hooks.json` 값과 같다(예: 6); (c) 최상위 키가
`{description, hooks}`의 부분집합(`grep -c '"version"' proj/.codex/hooks.json` = 0);
(d) SessionEnd handler의 timeout ≤ 3.

**AC-CW-003** (MUST, REQ-CW-004) — *Given* AC-CW-002의 프로젝트, *When*
`proj/.codex/config.toml`을 읽으면, *Then* `[mcp_servers.moai]` 테이블이 `command = "moai"`와
`default_tools_approval_mode = "writes"`를 포함하고(`grep -c 'default_tools_approval_mode'
proj/.codex/config.toml` ≥ 1), 도구명 열거가 없다(`grep -cE 'enabled_tools|disabled_tools'
proj/.codex/config.toml` = 0), 파일이 TOML로 파싱된다
(`python3 -c "import tomllib;tomllib.load(open('proj/.codex/config.toml','rb'))"` exit 0).

**AC-CW-004** (MUST, REQ-CW-001) — *Given* 스크래치 디렉터리, *When* 플래그 없이
`moai init proj2 --non-interactive`를 실행하면, *Then* `proj2/.codex/hooks.json`와
`proj2/.codex/config.toml`이 모두 부재하고(`test ! -f …` 2건 통과) `.mcp.json`에 moai 엔트리가
존재한다(오늘 동작 보존; `jq -r '.mcpServers.moai.command' proj2/.mcp.json` = `moai`).

**AC-CW-005** (MUST, REQ-CW-001) — *Given* 스크래치 디렉터리, *When*
`moai init proj3 --agent both --non-interactive`를 실행하면, *Then* `.mcp.json` moai 엔트리와
`.codex/hooks.json`·`.codex/config.toml`이 모두 존재한다.

### M1 — 병합·멱등·보존

**AC-CW-006** (MUST, REQ-CW-006) — *Given* AC-CW-002에서 배선된 프로젝트, *When* 생성 경로를
재실행하면(init 재실행 또는 `moai update`), *Then* hooks.json·config.toml 모두 sha256 무변경
(재실행 전후 `shasum -a 256` 비교 — 빈 diff).

**AC-CW-007** (MUST, REQ-CW-005) — *Given* 사용자 소유 hooks 엔트리(임의 matcher와
`"command": "my-own-hook"` handler)와 사용자 config.toml 테이블(`[mcp_servers.other]`)을 사전
삽입한 프로젝트, *When* `--agent codex` 배선을 실행하면, *Then* 사용자 엔트리·테이블이 그대로
존재하고(`grep -c 'my-own-hook' …/hooks.json` = 1; `grep -c 'mcp_servers.other'
…/config.toml` = 1) MoAI handler들이 추가된다. *And Given* 이미 존재하는 사용자 소유
`[mcp_servers.moai]` 테이블(내용 변형), *When* 배선을 실행하면, *Then* 그 테이블은 바이트 불변
(생성기 미수정 — 불일치는 doctor가 보고).

### M2 — 안내·update·가드

**AC-CW-008** (MUST, REQ-CW-008) — *Given* 신규 프로젝트, *When* `--agent codex` init를
실행하면, *Then* stdout/stderr에 Codex 신뢰 흐름 안내가 출력된다(실행 명령:
`moai init <dir> --agent codex --non-interactive 2>&1 | grep -c -- 'codex /hooks'` ≥ 1 —
토큰 base-0 실측이라 관측은 구현분에 귀속). *And Given* 배선 후 사용자가 MoAI handler 하나를
제거한 hooks.json, *When* 갱신 경로가 재생성하여 내용을 복원하면, *Then* 재신뢰 안내가 출력된다
(`… 2>&1 | grep -c -- '/hooks to re-trust'` ≥ 1). *And Given* 무변경
재생성, *Then* 재신뢰 안내는 출력되지 않는다(같은 grep = 0).

**AC-CW-009** (MUST, REQ-CW-009) — *Given* AC-CW-004의 claude 초기화 프로젝트, *When*
`moai update`를 실행하면, *Then* `.codex/hooks.json`·`.codex/config.toml`이 여전히 부재하다
(update는 배선을 만들지 않는다). *And Given* codex 배선 프로젝트, *When* `moai update`를
실행하면, *Then* hooks.json이 무변경이다(AC-CW-006과 동일 판정).

**AC-CW-010** (MUST, REQ-CW-011) — *Given* run 베이스 트리, *When*
`go test ./internal/cli/ -run TestMoaiMCPServer_AnnotationsMatchCatalog`를 실행하면, *Then*
테스트가 통과한다 — 비교 기준선은 **유효 read-only 값**(선언된 annotation 값, 선언 부재 시
기본 false)과 catalog `WriteCapable`의 도구별 동치이지 선언의 존재가 아니다(spec.md REQ-CW-011).
이 기준선에서 base 트리는 4도구(audit_cache·codex_audit·glm_audit·audit_multi — catalog READ,
유효 false)에서 실패하며, 이 4개 등록에 `WithReadOnlyHintAnnotation(true)`를 추가하는 것이
가드를 녹색으로 만드는 최소 델타다(plan M2가 수행, PRESERVE 예외로 허용된 유일한
mcp_server.go 편집). 음성 주장 포함: catalog-read 도구의 annotation 누락(유효 false)과
catalog-write 도구의 read-only 선언이 모두 가드에서 실패한다(가드의 실제 이빨).
선언-부재-허용 주장: catalog-write 도구의 선언 부재(유효 false, 예 — goal_arm·verify_snapshot)는
선언을 강제하지 않는 한 통과해야 한다.

### M3 — 런타임 어댑터

**AC-CW-011** (MUST, REQ-CW-007) — *Given* t83 골든 형식의 canned payload들, *When*
`go test ./internal/cli/ -run 'HarnessCodex'`를 실행하면, *Then* (a) `continue:false`(+stopReason)
를 emit하는 훅의 출력이 `decision:"block"` + 비지 않은 reason으로 재작성되고; (b) UserPromptSubmit
아닌 이벤트의 `systemMessage`가 폐기되며 폐기 기록이 `.moai/logs/codex-adapter.jsonl`에
event·key·content_length와 함께 append되고; (c) payload `hook_event_name`이 부속명령과 불일치하면
0이 아닌 exit + 진단; (d) 대상 훅의 exit code·stderr는 변경 없이 통과한다.

### M4 — doctor

**AC-CW-012** (MUST, REQ-CW-010 · REQ-CW-012) — *Given* codex 배선 프로젝트, *When*
`moai doctor`를 실행하면, *Then* "Codex Wiring" 진단이 표시된다
(`moai doctor 2>&1 | grep -c 'Codex Wiring'` ≥ 1). *And Given* hooks.json을 무단 수정한
상태(사이드카 해시와 불일치), *When* doctor를 실행하면, *Then* divergence가 보고되고
`/hooks to re-trust` 안내가 따른다. *And Given* claude 초기화 프로젝트, *Then* 동일 진단이
정보성 스킵 상태로 표시되고 doctor는 계속된다(advisory·fail-open). *And When*
`git diff -- internal/template/templates/.codex/`를 실행하면, *Then* 변경이 없다(M5 무변경).

### M1 — statusline (v0.3.0)

**AC-CW-013** (MUST, REQ-CW-013) — *Given* 스크래치 디렉터리, *When*
`moai init proj --agent codex --non-interactive`를 실행하면, *Then* `proj/.codex/config.toml`이
TOML로 파싱되고 `tui.status_line` 값이 정확히 기본 구성 5종이다(판정:
`python3 -c "import tomllib;c=tomllib.load(open('proj/.codex/config.toml','rb'));print(c['tui']['status_line']==['model-with-reasoning','context-remaining','git-branch','current-dir','thread-id'])"`
→ True). *And Given* 사용자가 `[tui]`에 `status_line = ["model"]`을 사전 설정한 config.toml,
*When* `--agent codex` 배선을 실행하면, *Then* 그 라인은 바이트 불변이다(사용자 소유 키 우선 —
`grep -c 'status_line = \["model"\]' …/config.toml` = 1 유지). *And Given* 회귀 가드, *When*
`go test ./internal/codexwiring/ -run StatusLine`을 실행하면, *Then* (a) 기본 구성이 5개
정식 토큰과 정확히 일치하고, (b) 기본 구성 ⊆ `statusLineAllowlist`(발행 원천 상수), (c) 기본
구성에 파싱 별칭 6종(session-id·context-usage·model-name·project·project-root·status·approval)이
0건이다(정식 토큰만 발행).

**AC-CW-014** (SHOULD, REQ-CW-014) — *Given* sync-phase 문서 갱신, *When*
`grep -rn '17827' README.md README.ko.md README.en.md README.ja.md README.zh.md docs-site/content/ko/`
를 실행하면, *Then* openai/codex#17827 미해소로 MoAI 고유 statusline 항목(goal·todo·SPEC) 노출이
불가함을 언급하는 히트 ≥ 1(토큰 base-0: 2026-08-24 실측 0hit). 미달 시 사유와 함께 부채 기록.

## §B. Edge cases (negative tests, MUST)

- **파손된 사용자 hooks.json**(JSON 파싱 불가) → 생성기는 파일을 수정하지 않고 진단 경고 후
  계행(init는 실패하지 않음). 파일 바이트 불변 주장.
- **파손된 config.toml**(테이블 경계 탐지 불가) → 미수정·경고·계행.
- **읽기 전용 파일시스템/권한** → 경고 후 init 계행(`.mcp.json` provisioning과 동일 원칙).
  단 REQ-CW-003 선검증 실패는 어떤 경로에서도 파일을 쓰지 않는다(미기록).
- **`.codex/` 디렉터리 부재** → 생성(파일 쓰기의 자연 전제).
- **e2e 테스트의 Codex 격리** — codex-cli를 호출하는 테스트(있다면)는 scratch `CODEX_HOME` 필수,
  실제 `~/.codex/` mtime 무변경 전후 확인(t83 위생). 산출물 단위 테스트은 codex-cli 불필요.
- **크로스플랫폼 컴파일** — `GOOS=windows GOARCH=amd64 go vet ./internal/...` exit 0
  (Windows 실행 세부는 문서화된 수동 확인 항목 — spec §H).

## §C. Definition of Done

- AC-CW-001..014 전부 판정 완료(MUST 13 통과 + SHOULD 1 판정) — §E.2에 명령 원문·출력과 함께
  기록. AC-CW-014(SHOULD)는 sync-phase 산출물이므로 run 종료 시점에는 위임 상태로 기록 가능.
- `go test ./internal/codexwiring/... -cover` ≥ 85% · `golangci-lint run` 해당 패키지 0 error.
- PRESERVE 목록(plan §D) 위반 0 — 특히 `internal/codexadapter`·`.codex/agents/**`·
  기존 init 플래그 무변경.
- 템플릿 무변경(`git diff --stat internal/template/templates/` 공백) — template 반입이
  생겼다면 Template-First + `make build` + neutrality 테스트가 대신 통과해야 함.
- CI(원격 전수) 초록 — 로컬 전체 스위트 금지 규율 준수.

## §D. Traceability

| AC | REQ | 검증 축 |
|---|---|---|
| AC-CW-001 | REQ-CW-001 | CLI 표면 + 폐쇄집합 |
| AC-CW-002 | REQ-CW-002 · REQ-CW-003 | 산출물 형태 + 화이트리스트 |
| AC-CW-003 | REQ-CW-004 | config 테이블 + 열거 부재 + TOML 파싱 |
| AC-CW-004 | REQ-CW-001 | 기본값 무변경(backward compat) |
| AC-CW-005 | REQ-CW-001 | both 의미론 |
| AC-CW-006 | REQ-CW-006 | 멱등성 |
| AC-CW-007 | REQ-CW-005 | 사용자 보존·무분쇄 |
| AC-CW-008 | REQ-CW-008 | 신뢰 안내 토큰(base-0 실측) |
| AC-CW-009 | REQ-CW-009 | update 규칙 |
| AC-CW-010 | REQ-CW-011 | annotation 가드 |
| AC-CW-011 | REQ-CW-007 | 런타임 매핑·기록·거부·패스스루 |
| AC-CW-012 | REQ-CW-010 · REQ-CW-012 | doctor 진단 + M5 무변경 |
| AC-CW-013 | REQ-CW-013 | status_line 산출 + create-if-absent + 정식 토큰 가드 |
| AC-CW-014 | REQ-CW-014 | 한계 문서화(base-0 토큰, SHOULD) |
