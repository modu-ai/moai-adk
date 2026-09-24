---
id: SPEC-CODEX-PREAPPROVAL-PROBE-001
document: plan
created: 2026-09-25
updated: 2026-09-25
author: manager-spec
card: t1172
---

# Plan — SPEC-CODEX-PREAPPROVAL-PROBE-001

## §A 맥락과 측정 근거

plan 단계에서 확인한 사실이다. 모델에 닿은 호출은 0회다. `codex exec`는 접속할 수 없는 제공자(`http://127.0.0.1:9/v1`)로만 돌렸고, 임시 `CODEX_HOME`(로그인 정보 없음)을 썼다. 명령·출력 전문은 `.moai/reports/t1172/verdict.md` §Plan, 원자료는 `.moai/reports/t1172/plan-checks/`.

| 사실 | 근거 |
|---|---|
| 설치된 codex는 `codex-cli 0.156.1` | `codex --version` |
| 베이스 `0356e8117`에 t1143 보안 수리 `51d3e5be6`(F1–F3, F5)와 병합 `a0b78213d`가 조상으로 들어 있다 | `git merge-base --is-ancestor` 두 번 모두 성공 |
| 생성 설정 writer는 `internal/codexwiring/configtoml.go` `EnsureMCPTable`(표가 없을 때만 만든다, 있으면 바이트 그대로). 표에는 `default_tools_approval_mode = "writes"`(`:99`)가 들어가고, 의사 점검용 정규형 목록도 같은 값을 적는다(`:174`). 호출처는 `internal/codexwiring/ownership.go:125` 하나, 의사 점검은 `internal/cli/doctor_codex.go:297` | 해당 줄 판독, `grep -rn 'EnsureMCPTable('` |
| `codex_role_audit`는 `mcp.WithReadOnlyHintAnnotation(false)`로 선언된 쓰기 도구다 | `internal/cli/codex_audit_mcp.go:204` |
| 부모용 지시면은 `internal/template/templates/AGENTS.md.tmpl`의 `audit-verdict-file` 행 하나다. 네 read-only 역할 TOML addenda(`internal/template/agentemit/agents-codex.yaml` `codex_role_addenda`)는 launcher가 띄운 **자식** 역할이 읽는 글이며, 거부는 부모 쪽에서 일어나므로 자식은 거부를 보지 못한다 | 두 파일 판독 |
| 이 저장소의 루트 `AGENTS.md`에는 `audit-verdict-file` 행이 없다. 이 행의 로컬 미러는 없다 | `grep -n 'audit-verdict-file' AGENTS.md` 출력 없음 |
| `codex mcp get moai --json`은 `enabled_tools`·`disabled_tools`·timeout 필드만 보이고 승인 관련 필드는 보이지 않는다 | 임시 프로젝트에서 실행, 출력 전문은 verdict §Plan |
| 틀린 enum 값(`"bogus"`)은 로더가 거부하고, 틀린 필드 이름(`approval_modex`)은 `codex mcp get`이 조용히 무시한다 | t1143 verdict 156–168행 |
| `codex exec --strict-config`는 틀린 필드 이름을 적재 단계에서 거부한다: `unknown configuration field \`mcp_servers.moai.tools.codex_role_audit.approval_modex\``, exit 1, 세션 없음. `-c` 재정의의 모르는 키도 같은 방식으로 거부한다(`stream_max_retries`) | `plan-checks/typo-strict-exec.err` |
| `codex --strict-config mcp get`, `codex --strict-config debug prompt-input`은 "not supported" | 실행 출력 |
| git 저장소가 아닌 디렉터리에 `CODEX_HOME` 신뢰 항목을 두면 `codex debug prompt-input`은 기본 sandbox를 `workspace-write`로 렌더한다(신뢰 항목이 없으면 `read-only`). 같은 디렉터리에서 `codex exec`는 `Not inside a trusted directory and --skip-git-repo-check was not specified.`로 시작을 거부한다 | `plan-checks/prompt-input-{trusted,untrusted}.txt`, `plan-checks/ok-strict-exec.err`, `plan-checks/codexhome-config.toml` |
| git 저장소 루트에서는 신뢰 항목 없이 시작 검사를 통과한다. 접속 불가 제공자로 돌리면 `thread.started`·`turn.started` 뒤에 `Reconnecting... waiting for network` 오류만 나오고 `item.*` 이벤트 0개, `timeout 90`으로 끝났다(exit 124). 끝난 뒤 남은 프로세스 없음 | `plan-checks/repo-exec.out`, `ps` 필터 출력 없음 |

**이 표가 확립하지 않는 것.** (1) 비git 디렉터리 신뢰 항목이 `exec` 게이트에서 통하지 않는 이유(경로 정규화인지, git 루트만 인정하는지)는 가르지 않았다. 픽스처를 git 저장소로 만들면 이 질문은 필요 없다. (2) 접속 불가 제공자 아래서 moai MCP 서버가 세션 시작 때 뜨는지는 재지 않았다(시작 검사에 쓴 트리는 신뢰되지 않아 프로젝트 층 MCP가 적재되지 않았을 수 있다). M1-a가 첫 측정으로 잰다. (3) `approval_mode = "approve"`가 `approval_policy = "never"`를 이기는지는 여전히 측정되지 않았다. 이것이 M1-b의 질문이다.

## §B 결정 (바뀔 가능성이 큰 것부터)

### D1. 채택 형태는 운영자 결정이다 (REQ-CPP-005)

세 갈래를 모두 적어 두고 측정 뒤에 하나를 고른다. 이 SPEC은 미리 고르지 않는다.

| 갈래 | 조건 | moai가 만드는 설정 | 지시면 |
|---|---|---|---|
| **A1 템플릿 기본값** | 처치 `RAN`, 대조 `REFUSED`, 운영자가 A1 선택 | 새로 만드는 `[mcp_servers.moai]` 표 뒤에 `[mcp_servers.moai.tools.codex_role_audit]` `approval_mode = "approve"`를 붙인다 | 거부 시 blocker 지시(D4)를 둔다. 기존 표를 가진 프로젝트는 여전히 거부될 수 있기 때문이다 |
| **A2 opt-in** | 처치 `RAN`, 대조 `REFUSED`, 운영자가 A2 선택 | 기본은 표를 넣지 않는다. 운영자가 켠 경우에만 넣는다. 켜는 방법은 D1-sub에서 정한다 | 거부 시 blocker 지시에 더해, 사전 승인을 켜는 방법 한 줄 |
| **N 채택하지 않음** | 처치 `REFUSED` 또는 `NOT MEASURED`, 또는 대조 `RAN`(판별 불가) | 아무것도 바꾸지 않는다 | 거부 시 blocker 지시만 |

대조가 `RAN`이면(쓰기 승인 모드에서도 호출이 통과) 사전 승인은 필요 없다는 뜻이다. 판별 결과로 쓰지 않고 `NOT DISCRIMINATING`으로 적은 뒤 리드에게 보고한다. #37과 다른 결과이므로 무엇이 달라졌는지부터 가려야 한다.

**D1-sub (A2일 때만).** [NEEDS CLARIFICATION: A2를 고르면 켜는 수단 — moai 설정 섹션의 키 하나인가, `moai codex` 계열 동사의 플래그인가, 사용자가 직접 표를 적는 문서 안내인가]

### D2. 기존 사용자 소유 표 (REQ-CPP-006)

writer는 이미 있는 `[mcp_servers.moai]` 표를 바이트 그대로 둔다(기존 규칙 유지). 그래서 A1을 골라도 이미 초기화된 프로젝트는 사전 승인을 받지 못한다. 의사 점검(`InspectMCPTable`)이 도구별 표의 부재를 보고할지는 운영자가 정한다.

[NEEDS CLARIFICATION: A1 또는 A2 채택 시, 기존 표에 도구별 표가 없는 프로젝트를 `moai doctor`가 보고해야 하는가(정규형 판정에 포함) 아니면 조용히 두는가]

### D3. 판별 프로브 설계 (REQ-CPP-001–004)

- **두 팔을 모두 새로 잰다.** #37은 재구성할 수 없으므로 대조 결과로 다시 쓰지 않는다. 대조 팔은 "이 픽스처에서 거부가 재현된다"는 양성 대조 역할도 한다.
- **같은 경로, 순차 실행.** 두 팔은 같은 루트 경로, 같은 `CODEX_HOME` 경로를 쓴다. 팔마다 `CODEX_HOME`을 같은 씨앗(로그인 사본 + 같은 `config.toml`)에서 새로 만든다. 그래야 인자 벡터의 `-C <root>`와 `CODEX_HOME` 경로가 두 팔에서 같다.
- **루트는 git 저장소.** 커밋 하나(README 한 파일)를 가진 저장소를 만들고 `.codex/`는 추적하지 않는다. 그래서 두 팔의 HEAD가 같다. 루트의 신뢰 항목도 `CODEX_HOME` `config.toml`에 넣는다(프로젝트 층 설정을 적재시키기 위함). 픽스처 저장소 생성은 테스트 코드 안에서 한다.
- **MCP 서버는 moai 하나.** decoy는 두지 않는다. 서버 명령은 기동을 기록한 뒤 이 워크트리에서 빌드한 moai로 넘기는 래퍼다. 빌드 커밋과 바이너리 sha256을 반출한다.
- **자식 프로세스가 생기지 않게.** 요청의 `out`을 `AGENTS.md`로 준다. 호출이 통과하면 launcher는 목적지 검증(`.moai/reports/` 아래가 아님)에서 거부하고, 프로세스를 띄우지 않으며 launch record도 쓰지 않는다(REQ-CAR-007). 그래서 팔 하나는 `codex exec` 한 개다. 통과했다는 증거는 도구 결과의 launcher 거부 문구다.
- **프롬프트.** `exec` custom tool을 통해 `codex_role_audit`를 정확히 한 번 부르라고 한다. shell 명령, 다른 도구, 파일 변경은 금지한다. #38의 결함(exec 통로까지 금지)을 되풀이하지 않는다. 두 팔의 프롬프트는 바이트 같다.
- **반출 → 비교 → 시작 검사 → LIVE** 순서를 코드로 강제한다. 앞 단계가 실패하면 뒤 단계는 시작하지 않는다.
- **결과 분류.** `REFUSED` = moai/`codex_role_audit` `mcp_tool_call` 항목의 오류가 Codex의 승인 거부 문구. `RAN` = 같은 항목의 결과·오류가 moai 서버의 문구(launcher 거부). `NOT MEASURED` = 그런 항목 0개. `INVALID` = 리드가 측정 결함으로 기록. `ABORTED` = 상한 때문에 멈춤.
- **판별 결과의 단일 출처**는 `.moai/reports/t1172/discriminator/outcome.txt`(`REFUSED`·`RAN`·`NOT MEASURED`·`NOT DISCRIMINATING` 중 하나). D1 분기와 AC-CPP-006이 이 파일을 읽는다.

### D4. 거부 시 지시 (REQ-CPP-008, F4)

`AGENTS.md.tmpl` `audit-verdict-file` 행에 한 문장을 더한다. 뜻: `codex_role_audit`가 승인 요구로 거부되면(오류 문구 `MCP tool call requires approval`) 감사를 건너뛰지 말고, `spawn_agent`로 대신하지 말고, 거부 문구를 인용한 blocker를 돌려준다. 이 문장은 세 갈래 모두에서 넣는다. 템플릿에는 SPEC ID, 카드 ID, 날짜를 넣지 않는다. 네 역할 addenda는 바꾸지 않는다(자식은 거부를 보지 못함, §A). 바꾸게 되면 `make agents-emit`으로 방출본을 다시 만든다. 루트 `AGENTS.md`에는 이 행이 없으므로 로컬 미러 수정은 없다.

### D5. 키 이름 로더 검증 (REQ-CPP-007)

- 경로: `codex exec --strict-config`. `codex mcp get`은 승인 필드를 보이지 않으므로 되읽기 수단이 못 된다(§A).
- 픽스처: git 저장소가 **아닌** 임시 디렉터리 + 그 경로의 `CODEX_HOME` 신뢰 항목 + 그 안 `.codex/config.toml`. 이 형태에서는 설정 적재가 시작 게이트보다 먼저 일어나고(§A: 오타는 적재 오류, 정상 키는 게이트 문구), 게이트가 세션을 막으므로 네트워크 요청도 없다. 제공자도 접속 불가 주소로 준다(이중 안전).
- 입력: 테스트가 손으로 쓴 문자열이 아니라 제품 함수가 만든 설정 바이트(채택 갈래의 writer 출력). N 갈래에서도 writer 출력의 기존 키(`command`, `args`, `env_vars`, `default_tools_approval_mode`)를 같은 방식으로 검증한다.
- 양성 대조: 같은 픽스처에서 `approval_modex`로 바꾼 설정이 `unknown configuration field` + 그 점 경로로 보고되어야 한다. 보고되지 않으면 테스트는 실패한다(skip 아님). 프로젝트 층이 아예 적재되지 않는 픽스처였다면 여기서 드러난다.
- 판정 기준: 정상 설정 → 표준 오류에 `Error loading config`도 `unknown configuration field`도 없다. 오타 설정 → 둘 다 있고 점 경로가 정확히 일치한다.
- codex가 없을 때: `exec.LookPath`(또는 테스트용 재정의 환경 변수 `MOAI_CODEX_BIN`) 실패 시 `CODEX_NOT_INSTALLED: <이유>`로 skip한다. skip은 AC 판정에서 PASS가 아니다.
- 프로세스는 `timeout` 성격의 context로 묶고, 기록한 pid(프로세스 그룹)만 정리한다.

### D6. 이월 AC의 픽스처 수정 (REQ-CPP-009)

AC-CAR-010/011의 본문·판정식은 경로 치환 하나만 적용해 그대로 둔다. 바꾸는 것은 테스트 픽스처다(판정식이 아니다).

| 수정 | 이유 |
|---|---|
| 사용자 층 decoy 표에 `enabled_tools = ["spec_progress"]`를 둔다 | #36이 decoy 쪽 `codex_role_audit`로 라우팅되었다. decoy가 이 도구를 노출하지 않으면 부모는 moai 쪽을 부른다. decoy의 기동 자체(양성 대조 `parent_mcp_launches.decoy >= 1`)는 유지된다 — M1-a의 시작 검사가 이것을 LIVE 전에 잰다 |
| 프로젝트 층 moai 표는 채택 갈래의 writer 출력을 그대로 쓴다 | 제품이 실제로 만드는 설정으로 (b) 경로를 재기 위함 |
| 픽스처 입력 반출(REQ-CPP-001)과 시작 검사(REQ-CPP-003)를 AC-CAR-010/011에도 적용 | #39·B''의 재구성 불가 재발 방지 |
| 부모 프롬프트가 `exec` custom tool 통로를 허용 | #38 교훈 |
| `route.txt`를 `.moai/reports/t1172/m1-route/route.txt`에 둔다(값 `mcp`, t1143 반출본과 sha256 `97c5f37f…a679ee` 같음) | 치환된 판정식이 이 경로를 읽는다 |

증거 디렉터리 환경 변수 이름 `MOAI_T1143_EVIDENCE_DIR`는 기존 테스트의 상수이므로 이름은 그대로 두고 값만 `../../.moai/reports/t1172`로 준다(치환 결과와 같음).

**N 갈래에서 AC-CAR-010.** (b) 경로가 승인 거부로 막히므로 실행하지 않고 `NOT_RUN — pre-approval not adopted`로 적는다. AC-CAR-011은 launcher를 직접 부르므로 승인 경로와 무관하게 돌린다.

## §C Pre-flight (run 착수 전)

1. `git -C <worktree> rev-parse --show-toplevel`이 이 워크트리, 브랜치 `WT-codex-preapproval-probe`.
2. AC-CPP-001 `true`(보안 수리가 조상).
3. `codex --version` = `codex-cli 0.156.1`. 다르면 멈추고 보고한다(§A의 측정이 이 버전 기준).
4. 로그인 파일 sha256을 적는다(LIVE 전후 비교용).
5. `make build`로 moai를 빌드하고 커밋·sha256을 적는다.

## §D LIVE 호출 상한 (REQ-CPP-004, REQ-CPP-010)

첫 LIVE 호출 전에 이 표를 `.moai/reports/t1172/verdict.md`에 옮겨 적는다. 단위는 `codex exec` 프로세스 하나다.

| 항목 | 기본 호출 수 | 재실행(리드가 `INVALID` 기록 뒤 1회만) | 호출당 timeout | 총 벽시계 | 모델 턴 |
|---|---|---|---|---|---|
| 판별 프로브(대조 1 + 처치 1) | 2 | +2 | 330 s | 900 s | 호출당 사용자 턴 1 |
| AC-CAR-010 | 3 | +3 | 330 s(테스트가 띄운 호출), 자식은 launcher 상한 `config.DefaultCodexAuditTimeout` | 1100 s(`-timeout=1200s` 안) | 호출당 1 |
| AC-CAR-011 | 2 | +2 | 330 s | 800 s(`-timeout=900s` 안) | 호출당 1 |
| **절대 상한** | **14** | | | | |

- 시작 검사(REQ-CPP-003)는 LIVE가 아니다. 별도 장부 G에 적고 상한 8회(픽스처 넷 × 1, 재실행 대비 × 2), 호출당 `timeout 20`. 모델 요청 0이 조건이다. 한 번이라도 모델 응답(`item.*` 이벤트)이 보이면 즉시 멈추고 `ABORTED`로 적는다.
- 키 이름 로더 테스트(REQ-CPP-007)의 `codex exec`는 게이트에서 끝나 모델 요청이 없다. LIVE가 아니며 장부 G에도 넣지 않는다(결정적 테스트).
- 정지 규칙: 상한에 닿으면 다음 호출을 시작하지 않고 `ABORTED`. 전제(`mcp_tool_call` moai/`codex_role_audit` 1개 이상)가 없으면 `NOT MEASURED`. 판별 프로브가 두 번째로도 `NOT MEASURED`이면 LIVE를 더 하지 않고 "pre-approval effect not measured"로 멈춘 뒤 N 갈래로 간다.
- 프로세스 정리: 테스트가 기록한 pid(프로세스 그룹)만 죽인다. 이름으로 찾아 죽이지 않는다.
- 로그인 파일 sha256을 모든 LIVE 호출 전후에 기록한다.

## §E 자가 검증 (plan 단계)

- SPEC ID 정규식 검사: `PASS`(실행 출력).
- 12개 필수 frontmatter 필드: spec.md에 모두 있음.
- ID 중복: `.moai/specs/`에 `PREAPPROVAL` 포함 디렉터리 0개.
- 이월 AC 치환 검사(AC-CPP-011)는 plan 커밋에서 실행해 `true`를 확인한다(verdict §Plan에 출력).

## §F 마일스톤 (우선순위 순)

### M1 — 판별 프로브 (Priority High, 결정 변경 가능성 가장 큼)

- M1-a (LIVE 없음): 픽스처 생성·반출·팔 비교·시작 검사를 하는 하네스를 `internal/cli`의 LIVE 게이트 테스트로 만든다(환경 변수 `MOAI_CODEX_PREAPPROVAL_LIVE=1`, `MOAI_T1172_EVIDENCE_DIR`가 없으면 `NOT_RUN`으로 skip). 반출·비교 로직은 LIVE 없이 결정적으로 시험한다. 시작 검사를 넷 픽스처(판별 대조, 판별 처치, AC-CAR-010 부모, AC-CAR-011)에 돌려 moai·decoy 기동 수를 적는다. 접속 불가 제공자 아래에서 MCP 서버가 뜨지 않으면(측정 결과 0) 멈추고 리드에게 보고한다. 대체 기준은 리드가 정한다.
- M1-b (LIVE 2): 판별 프로브 두 호출. `outcome.txt` 작성.
- 운영자 결정 지점: `outcome.txt`가 `RAN`이고 대조가 `REFUSED`이면 D1(A1/A2)과 D2를 운영자가 정한다. 아니면 N.

### M2 — 조건부 방출과 로더 검증 (Priority High)

- 채택 갈래에 맞춰 `internal/codexwiring`의 writer와 정규형 목록을 바꾼다(N이면 코드 변경 없음). 기존 표 바이트 불변 테스트 유지.
- 키 이름 로더 테스트(D5)를 `internal/codexwiring`에 둔다. N 갈래에서도 둔다(기존 키 검증 + 양성 대조).
- 재측정 범위: `./internal/codexwiring/...`, `./internal/cli -run 'Doctor|Codex'`.

### M3 — 거부 시 지시 (Priority Medium)

- `AGENTS.md.tmpl` 행 수정(D4). `internal/template/agentemit` 지시면 테스트 갱신. 중립성 판정.
- 재측정 범위: `./internal/template/agentemit/...`, `./internal/template/...` 중 AGENTS 렌더 테스트.

### M4 — 이월 LIVE (Priority Medium)

- AC-CAR-011 (LIVE 2): 모든 갈래.
- AC-CAR-010 (LIVE 3): A1/A2 갈래만. N이면 `NOT_RUN`.
- 픽스처 수정은 D6.

### 마지막 — 기계적 정리 (Priority Low)

- progress.md §E 기록은 run·sync 담당 에이전트가 한다. 이 plan은 자리만 만든다.

## §G 위험

| 위험 | 대응 |
|---|---|
| 접속 불가 제공자 아래에서 MCP 서버가 늦게 떠서 시작 검사가 기동 0을 보인다 | M1-a 첫 측정. 0이면 리드 보고 후 결정 |
| 처치 호출에서 모델이 도구를 부르지 않는다(`NOT MEASURED`) | 프롬프트가 `exec` 통로를 허용. 두 번째 `NOT MEASURED`면 LIVE 종료 |
| 픽스처 git 저장소 생성이 워크트리 세션 가드에 막힌다 | 저장소 생성은 테스트 프로세스 안(`t.TempDir()`)에서 한다. 셸 명령으로 만들지 않는다 |
| codex 버전이 올라가 `--strict-config` 동작이나 거부 문구가 바뀐다 | Pre-flight 3에서 버전 고정 확인. 로더 테스트는 양성 대조가 먼저 실패하므로 조용히 통과하지 않는다 |
| A1을 골라도 기존 프로젝트는 거부된다 | D4 지시가 세 갈래 모두에 들어간다. D2는 운영자 결정 |
| 사전 승인이 통과하면 쓰기 도구 하나가 무인으로 호출된다 | 선행 수리(F1–F3)로 목적지는 `.moai/reports/` 아래 디렉터리 핸들로 묶여 있다(AC-CPP-001). 사전 승인은 이 도구 하나로 한정(REQ-CPP-006) |

## §H 교차 참조

- 선행 SPEC: `.moai/specs/SPEC-CODEX-AUDIT-READONLY-001/` (REQ-CAR-005, 007, 010, 011; AC-CAR-010, 011).
- 선행 증거(primary checkout 반출, 커밋되지 않음): `.moai/reports/t1143/verdict.md`(sha256 `00eea641…18ae`), `.moai/reports/t1143/sync-audit.md`(sha256 `93f8bedb…6cca3`).
- writer 규칙: SPEC-CODEX-WIRING-001(표가 없을 때만 만들고, 기존 표는 의사 점검이 보고).
- 이 카드의 증거: `.moai/reports/t1172/`.
