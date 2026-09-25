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
| 초기 plan 프로브는 `codex-cli 0.156.1`에서 측정했다. 이 결과는 과거 버전의 근거다 | 당시 `codex --version`, `verdict.md` §Plan |
| 이번 LIVE 대상 CLI는 `codex-cli 0.157.0`이다 | 이 worktree의 `codex --version` → `codex-cli 0.157.0` (2026-09-25); 운영자 버전 선택 |
| 베이스 `0356e8117`에 t1143 보안 수리 `51d3e5be6`(F1–F3, F5)와 병합 `a0b78213d`가 조상으로 들어 있다 | `git merge-base --is-ancestor` 두 번 모두 성공 |
| 생성 설정 writer는 `internal/codexwiring/configtoml.go` `EnsureMCPTable`(표가 없을 때만 만든다, 있으면 바이트 그대로). 표에는 `default_tools_approval_mode = "writes"`(`:99`)가 들어가고, 의사 점검용 정규형 목록도 같은 값을 적는다(`:174`). 호출처는 `internal/codexwiring/ownership.go:125` 하나, 의사 점검은 `internal/cli/doctor_codex.go:297` | 해당 줄 판독, `grep -rn 'EnsureMCPTable('` |
| `codex_role_audit`는 `mcp.WithReadOnlyHintAnnotation(false)`로 선언된 쓰기 도구다 | `internal/cli/codex_audit_mcp.go:204` |
| 부모용 지시면은 `internal/template/templates/AGENTS.md.tmpl`의 `audit-verdict-file` 행 하나다. 네 read-only 역할 TOML addenda(`internal/template/agentemit/agents-codex.yaml` `codex_role_addenda`)는 launcher가 띄운 **자식** 역할이 읽는 글이며, 거부는 부모 쪽에서 일어나므로 자식은 거부를 보지 못한다 | 두 파일 판독 |
| 이 저장소의 루트 `AGENTS.md`에는 `audit-verdict-file` 행이 없다. 이 행의 로컬 미러는 없다 | `grep -n 'audit-verdict-file' AGENTS.md` 출력 없음 |
| `codex mcp get moai --json`은 `enabled_tools`·`disabled_tools`·timeout 필드만 보이고 승인 관련 필드는 보이지 않는다 | 임시 프로젝트에서 실행, 출력 전문은 verdict §Plan |
| 틀린 enum 값(`"bogus"`)은 로더가 거부하고, 틀린 필드 이름(`approval_modex`)은 `codex mcp get`이 조용히 무시한다 | t1143 verdict 156–168행 |
| `codex exec --strict-config`는 틀린 필드 이름을 적재 단계에서 거부한다: `unknown configuration field \`mcp_servers.moai.tools.codex_role_audit.approval_modex\``, exit 1, 세션 없음. `-c` 재정의의 모르는 키도 같은 방식으로 거부한다(`stream_max_retries`) | `plan-checks/typo-strict-exec.err` |
| 대상 `codex-cli 0.157.0` 재측정: 정상 설정은 non-git 시작 게이트(`Not inside a trusted directory…`)에 도달한다. 오타는 `Error: <임시 경로>/.codex/config.toml:<행>:<열>: unknown configuration field \`mcp_servers.moai.tools.codex_role_audit.approval_modex\``, 잘못된 enum은 같은 접두사에 `unknown variant \`bogus\``를 낸다. 세 경우 모두 exit 1이다 | 2026-09-25, 로그인 없는 임시 `CODEX_HOME`, 접속 불가 제공자, 이 worktree의 `codex-cli 0.157.0`; 로더 3경우 직접 실행 |
| `codex --strict-config mcp get`, `codex --strict-config debug prompt-input`은 "not supported" | 실행 출력 |
| git 저장소가 아닌 디렉터리에 `CODEX_HOME` 신뢰 항목을 두면 `codex debug prompt-input`은 기본 sandbox를 `workspace-write`로 렌더한다(신뢰 항목이 없으면 `read-only`). 같은 디렉터리에서 `codex exec`는 `Not inside a trusted directory and --skip-git-repo-check was not specified.`로 시작을 거부한다 | `plan-checks/prompt-input-{trusted,untrusted}.txt`, `plan-checks/ok-strict-exec.err`, `plan-checks/codexhome-config.toml` |
| git 저장소 루트에서는 신뢰 항목 없이 시작 검사를 통과한다. 접속 불가 제공자로 돌리면 `thread.started`·`turn.started` 뒤에 `Reconnecting... waiting for network` 오류만 나오고 `item.*` 이벤트 0개, `timeout 90`으로 끝났다(exit 124). 끝난 뒤 남은 프로세스 없음 | `plan-checks/repo-exec.out`, `ps` 필터 출력 없음 |

**이 표가 확립하지 않는 것.** (1) 비git 디렉터리 신뢰 항목이 `exec` 게이트에서 통하지 않는 이유(경로 정규화인지, git 루트만 인정하는지)는 가르지 않았다. 픽스처를 git 저장소로 만들면 이 질문은 필요 없다. (2) 접속 불가 제공자 아래서 moai MCP 서버가 세션 시작 때 뜨는지는 위 표의 실행으로는 재지 않았다(신뢰되지 않은 트리였다). plan-audit iter 1이 신뢰 항목이 있는 git 픽스처에서 이것을 쟀다: 기본 모델로 moai 기동 1회, `item.` 0회, stderr 0바이트(`.moai/reports/t1172/plan-audit-iter1-probes/startup-default-model.jsonl`, `moai-launches.log`). 같은 측정에서 `model = "gpt-5"`는 비모델 `item.completed`(`type:"error"`)를 1회 냈고(`startup-model-gpt5.jsonl`), 서버는 루트에 `.moai/state/config-cache.json`을 썼다. M1-a는 이것을 실제 픽스처에서 다시 확인할 뿐이다. 기동 0이 나오면 그 픽스처는 측정 결함이므로 LIVE를 시작하지 않고 리드에게 보고한다(대체 기준은 두지 않는다). (3) `approval_mode = "approve"`가 `approval_policy = "never"`를 이기는지는 여전히 측정되지 않았다. 이것이 M1-b의 질문이다.

## §B 결정 (바뀔 가능성이 큰 것부터)

### D1. 채택 형태는 운영자 결정이다 (REQ-CPP-005)

세 갈래를 모두 적어 두고 측정 뒤에 하나를 고른다. 이 SPEC은 미리 고르지 않는다.

| 갈래 | 조건 | moai가 만드는 설정 | 지시면 |
|---|---|---|---|
| **A1 템플릿 기본값** | 처치 `RAN`, 대조 `REFUSED`, 운영자가 A1 선택 | 새로 만드는 `[mcp_servers.moai]` 표 뒤에 `[mcp_servers.moai.tools.codex_role_audit]` `approval_mode = "approve"`를 붙인다 | 거부 시 blocker 지시(D4)를 둔다. 기존 표를 가진 프로젝트는 여전히 거부될 수 있기 때문이다 |
| **A2 opt-in** | 처치 `RAN`, 대조 `REFUSED`, 운영자가 A2 선택 | 기본은 표를 넣지 않는다. 운영자가 켠 경우에만 넣는다. 켜는 방법은 D1-sub에서 정한다 | 거부 시 blocker 지시에 더해, 사전 승인을 켜는 방법 한 줄 |
| **N 채택하지 않음** | 처치 `REFUSED` 또는 `NOT MEASURED`, 또는 대조 `RAN`(판별 불가) | 아무것도 바꾸지 않는다 | 거부 시 blocker 지시만 |

대조가 `RAN`이면(쓰기 승인 모드에서도 호출이 통과) 사전 승인은 필요 없다는 뜻이다. 판별 결과로 쓰지 않고 `NOT DISCRIMINATING`으로 적은 뒤 리드에게 보고한다. #37과 다른 결과이므로 무엇이 달라졌는지부터 가려야 한다.

D1(A1 대 A2) 자체는 여기서 정하지 않는다. 판별 결과가 `RAN`이고 대조가 `REFUSED`일 때만 운영자에게 묻는다. 운영자 결정은 리드가 받아 `verdict.md`에 `OPERATOR-DECISION adoption=<값> card=t1172 recorded_by=lead at=<UTC RFC 3339>` 한 줄로 적는다(AC-CPP-006). 이 줄은 리드의 기록이며, 운영자 결정의 증거는 그 줄이 가리키는 대화다.

**D1-sub (A2일 때만) — 운영자 Kickoff 확정: `config-section-boolean`.** 켜는 수단은 `.moai/config/sections/` 아래 YAML 파일의 불리언 키 하나이고, Codex 설정 생성기(`internal/codexwiring`)가 그 키를 읽는다. `moai update`가 섹션 파일을 보존하므로 선택이 업데이트 뒤에도 남는다. 키 이름과 파일은 run M2에서 기존 섹션 구조에 맞춰 정하고 progress.md에 적는다.

### D2. 기존 사용자 소유 표 (REQ-CPP-006)

writer는 이미 있는 `[mcp_servers.moai]` 표를 바이트 그대로 둔다(기존 규칙 유지). 그래서 A1을 골라도 이미 초기화된 프로젝트는 사전 승인을 받지 못한다.

**운영자 Kickoff 확정: `doctor-warning`.** A1 또는 A2가 채택되고(A2는 켠 경우만) 프로젝트의 `[mcp_servers.moai]` 표에 도구별 표가 없으면, `moai doctor`는 실패가 아닌 WARNING으로 보고하고, 더할 표(`[mcp_servers.moai.tools.codex_role_audit]` `approval_mode = "approve"`)를 수동 추가 안내로 보여 준다. N 갈래에서는 보고하지 않는다.

운영자가 확정한 D1-sub·D2와 다른 구현을 하려면, run 전에 이 SPEC을 고치고 plan-audit를 다시 받는다.

**확정 기록 (D-N9, D-N15).** 운영자는 첫 LIVE 호출 전에 D1-sub·D2를 `config-section-boolean`·`doctor-warning`으로 확정했다. 리드는 첫 LIVE 전에 `verdict.md`에 `OPERATOR-CONFIRM d1sub=config-section-boolean d2=doctor-warning card=t1172 recorded_by=lead at=<UTC RFC 3339>` 한 줄을 적는다. AC-CPP-006은 모든 갈래에서 이 줄이 정확히 하나이고 그 시각이 장부의 첫 LIVE 시작보다 앞서기를 요구한다. M1-b 뒤의 운영자 결정 지점에서 정하는 것은 D1(A1/A2)뿐이며, 그 결정은 `OPERATOR-DECISION` 줄(판별 마지막 시도가 끝난 뒤의 시각)로 남는다. 이 기록은 운영자 답변을 반출한 것으로 취급하며, 판별 결과나 LIVE 성공의 증거로 취급하지 않는다.

### D3. 판별 프로브 설계 (REQ-CPP-001–004)

- **두 팔을 모두 새로 잰다.** #37은 재구성할 수 없으므로 대조 결과로 다시 쓰지 않는다. 대조 팔은 "이 픽스처에서 거부가 재현된다"는 양성 대조 역할도 한다.
- **같은 경로, 순차 실행, 루트 재사용.** 두 팔은 같은 루트 경로, 같은 `CODEX_HOME` 경로를 쓴다. 팔마다 `CODEX_HOME`을 같은 씨앗(로그인 사본 + 같은 `config.toml`)에서 새로 만든다. 그래야 인자 벡터의 `-C <root>`와 `CODEX_HOME` 경로가 두 팔에서 같다. 루트는 새로 만들지 않고 재사용한다(새로 만들면 커밋 시각이 달라져 HEAD가 바뀐다).
- **루트 상태 초기화와 반출 (D8).** moai MCP 서버는 기동 때 루트에 `.moai/state/config-cache.json`을 쓴다(감사 측정). 그래서 시작 검사와 LIVE 호출 **각각의 직전에** 하네스가 루트에서 `git clean -ffdx`를 돌리고, 그 팔의 `.codex/config.toml`을 반출본 바이트로 다시 쓴다. 이 시점의 루트 상태(`git status --porcelain --ignored` 출력과 `.git` 밖 파일 목록)를 `root-state.txt`로 반출하고, 두 팔의 `root-state.txt`는 같아야 한다(팔 diff에 들어가므로). 하네스는 매 호출 직전에 그 순간의 루트 상태 해시를 장부 행 `inputs_sha256.root_state`에 적는다. 반출본 해시와 다르면 호출하지 않고 `stop` 행을 남긴다.
- **초기화 대상 보호 (D-N6).** `git clean -ffdx`는 반드시 `git -C <root> clean -ffdx` 형태로만 부르고, 그 전에 다음을 모두 단언한다: `<root>`가 비어 있지 않은 절대 경로다, `t.TempDir()`(따라서 `os.TempDir()`) 아래에 있다, `git -C <root> rev-parse --show-toplevel`이 `<root>`와 같다, 이 저장소의 어떤 워크트리 경로(`git worktree list --porcelain`의 `worktree` 줄)와도 같거나 그 아래가 아니다, 픽스처를 만들 때 쓴 표지 파일 `<root>/.fixture-sentinel`(내용: 하네스가 만든 무작위 토큰)이 있고 내용이 같다. 하나라도 어긋나면 초기화를 거부하고 `stop` 행을 남긴다. 모든 시작 검사·LIVE 행에 `fixture_root`(그 절대 경로)와 `fixture_sentinel: true`(표지 확인 결과)를 적는다(AC-CPP-014). 거부 경로는 가짜 codex 테스트의 하위 테스트 `clean_refuses_non_fixture_root`가 빈 경로·상대 경로·임시 디렉터리 밖 경로·표지 없는 경로 넷을 주고 초기화가 일어나지 않음을 단언한다(AC-CPP-012).
- **루트는 git 저장소.** 커밋 하나(README 한 파일)를 가진 저장소를 만들고 `.codex/`는 추적하지 않는다. 그래서 두 팔의 HEAD가 같다. 루트의 신뢰 항목도 `CODEX_HOME` `config.toml`에 넣는다(프로젝트 층 설정을 적재시키기 위함). 픽스처 저장소 생성은 테스트 코드 안에서 한다.
- **모델 고정 (D11, D-N3).** `CODEX_HOME` `config.toml`과 프로젝트 `.codex/config.toml` 어디에도 `model` 키를 두지 않고, 인자에도 모델 지정을 주지 않는다 — `-m`/`--model`의 떨어진 형태와 붙은 형태(`--model=…`, `-m…`), `-c model=…`(두 인자이든 한 인자이든), `--config=model=…` 모두(D-N14). 시작 검사와 LIVE가 같은 기본 모델을 쓴다. `model = "gpt-5"` 같은 설정은 비모델 `error` 항목을 만든다(감사 측정).
- **장부와 해시 연결 (D6, D7).** 모든 시작 검사·LIVE 호출·정지는 `.moai/reports/t1172/ledger.json` 한 파일에 행으로 적는다. 행 필드: `kind`(`startup`·`live`·`stop`·`invalidate`), `fixture`(`disc-control`·`disc-treatment`·`car010`·`car011`; `invalidate` 행은 묶음 이름 `disc`·`car010`·`car011`), `label`(이월 LIVE 행은 호출 라벨 — D6), LIVE 행만 `attempt`(1부터 시작하는 정수 시도 번호, 재실행마다 1 증가), `temp_root`·`fixture_root`·`fixture_sentinel`(시작 검사·LIVE 행; `temp_root`는 하네스가 받은 `t.TempDir()` 절대 경로이고 `fixture_root`는 그 아래다, D-N13), `started_ns`·`ended_ns`(정수 epoch 나노초 — 문자열 시각 비교를 쓰지 않는다), `argv`(배열), `exit`, `bound_by`(`test` 또는 `launcher`), `inputs_sha256`(`codex_home_config`·`project_config`·`argv`·`prompt`·`root_state`·`moai_binary`), 시작 검사 행만 `provider_base_url`·`auth_file_present`·`env_auth_present`(인증 환경 변수가 시작 검사 프로세스 환경에 있었는지), LIVE 행만 `auth_sha256_before`·`auth_sha256_after`, `stop` 행만 `reason`, `invalidate` 행만 `attempt`·`recorded_by: "lead"`·`recorded_ns`·`reason`. `inputs_sha256`은 **그 호출이 실제로 쓴 바이트**를 해시한 값이고, 판정식은 이것을 반출본의 해시와 비교한다. 반출 매니페스트는 픽스처 묶음마다 하나(`discriminator/export-manifest.json`, `car010/export-manifest.json`, `car011/export-manifest.json`)이고 `exported_ns`와 반출 파일별 sha256을 담는다. 다시 반출하면 매니페스트를 새로 쓴다(`exported_ns`가 바뀐다). 팔 diff 메타(`arm-diff.meta.json`)는 `written_ns`를 담는다. 판별 증거 `evidence.json`은 그 판정이 읽은 시도 번호 `attempt`와 `recorded_pids`·`killed_pids`(하네스가 기록한 pid와 실제로 종료한 pid)를 담고, 팔마다 `user_turns`(그 팔 `--json` 출력의 `turn.started` 수, 호출당 사용자 턴 1개 상한의 증거, D-N16)를 담는다.
- **재실행 (D-N1).** 리드가 한 시도를 `INVALID`로 판정하면, 다음 시도 전에 리드가 장부에 `invalidate` 행(`fixture`는 묶음 이름, `attempt`는 무효가 된 번호, `recorded_ns`는 그 시도의 마지막 끝과 다음 시도의 첫 시작 사이)을 적는다. 다음 시도는 번호를 1 올린다. 묶음마다 재실행은 한 번(시도 번호 최대 2). 판별 결과 파일(`evidence.json`·`outcome.txt`·`live.jsonl`)과 시작 검사 원자료는 마지막 시도의 것이고, 이전 시도의 것은 `attempt-<n>/` 아래로 옮겨 보존한다. 장부 행은 지우지 않는다. 판정식은 묶음별로 마지막 시도만 보고, 무효 시도가 기록 없이 남아 있으면 거짓이다.
- **재반출과 시작 검사 순서 (D-N2).** 픽스처 입력이 바뀌면 다시 반출하고(매니페스트 `exported_ns` 갱신), 그 뒤 그 픽스처의 시작 검사를 다시 돌린 다음에야 LIVE를 시작한다. AC-CPP-004는 픽스처마다 **마지막** 시작 검사 행만 반출본과 대조하고, 그 행이 마지막 반출 뒤에 시작해 마지막 시도의 첫 LIVE 전에 끝났기를 요구한다. 이전 시작 검사 행(예: M1-a의 car010)은 기록으로 남고 대조 대상이 아니다.
- **시작 검사의 안전 (D9, D-N3).** 시작 검사용 `CODEX_HOME`에는 반출된 `config.toml`만 두고 로그인 파일을 두지 않으며, 프로세스 환경에서 인증 변수(`OPENAI_API_KEY` 등 `*_API_KEY`)를 지운다(`env_auth_present: false`로 기록). 제공자는 `-c 'model_providers.dead={name="dead",base_url="http://127.0.0.1:9/v1",wire_api="responses",stream_max_retries=0,request_max_retries=0}' -c 'model_provider="dead"'`로 준다. 호출당 `timeout 20`.
- **MCP 서버는 moai 하나.** decoy는 두지 않는다. 서버 명령은 기동을 기록한 뒤 이 워크트리에서 빌드한 moai로 넘기는 래퍼다. 빌드 커밋과 바이너리 sha256을 반출한다.
- **자식 프로세스가 생기지 않게.** 요청의 `out`을 `AGENTS.md`로 준다. 호출이 통과하면 launcher는 목적지 검증(`.moai/reports/` 아래가 아님)에서 거부하고, 프로세스를 띄우지 않으며 launch record도 쓰지 않는다(REQ-CAR-007). 그래서 팔 하나는 `codex exec` 한 개다. 통과했다는 증거는 도구 결과의 launcher 거부 문구다.
- **프롬프트.** `exec` custom tool을 통해 `codex_role_audit`를 정확히 한 번 부르라고 한다. shell 명령, 다른 도구, 파일 변경은 금지한다. #38의 결함(exec 통로까지 금지)을 되풀이하지 않는다. 두 팔의 프롬프트는 바이트 같다.
- **반출 → 비교 → 시작 검사 → LIVE** 순서를 코드로 강제한다. 앞 단계가 실패하면 뒤 단계는 시작하지 않고, 장부에 `stop` 행(사유 포함)을 남긴다. `stop` 행 뒤에는 LIVE 행이 없어야 한다(AC-CPP-012). 이 거부 경로는 가짜 codex로 결정적 테스트를 한다(`TestCodexPreApprovalGateRefusals`: 입력 하나 삭제, `argv.txt` 한 줄 추가, 시작 검사 표준 오류에 게이트 문구 — 각 경우 가짜 codex의 LIVE 기동 0회와 `stop` 행을 단언).
- **팔 diff 생성 (D5).** 하네스는 `exec.Command("diff", "-r", "inputs/control", "inputs/treatment")`를 작업 디렉터리 `discriminator/`에서 돌려 `arm-diff.txt`를 만든다(셸을 거치지 않으므로 셸 함수 `diff --color`의 영향이 없다). 판정식은 `command diff -r`로 같은 출력을 다시 만든다.
- **결과 분류.** `REFUSED` = moai/`codex_role_audit` `mcp_tool_call` 항목의 오류가 Codex의 승인 거부 문구. `RAN` = 같은 항목의 결과 문구가 moai 서버에서 온 것 — moai의 `toolErr`가 붙이는 접두사 `codex_role_audit: `(`internal/cli/mcp_server.go:975`, `Text: tool + ": " + err.Error()`)로 시작하고, 이어서 launcher의 거부 문구 `codex audit sync-auditor: destination rejected: ` 또는 `codex audit sync-auditor: working root rejected: `가 온다(`internal/cli/codex_audit_launch.go:165`, `:171`, `:175`). 요청의 역할은 `sync-auditor`로 고정한다. Codex 쪽 timeout 같은 다른 오류는 `RAN`이 아니라 `INVALID` 후보다. `NOT MEASURED` = 그런 항목 0개. `INVALID` = 리드가 측정 결함으로 기록. `ABORTED` = 상한 때문에 멈춤.
- **판별 결과의 단일 출처**는 `.moai/reports/t1172/discriminator/outcome.txt`(`REFUSED`·`RAN`·`NOT MEASURED`·`NOT DISCRIMINATING` 중 하나). D1 분기와 AC-CPP-006이 이 파일을 읽는다.

### D4. 거부 시 지시 (REQ-CPP-008, F4)

`AGENTS.md.tmpl` `audit-verdict-file` 행에 한 문장을 더한다. 뜻: `codex_role_audit`가 승인 요구로 거부되면(오류 문구 `MCP tool call requires approval`) 감사를 건너뛰지 말고, `spawn_agent`로 대신하지 말고, 거부 문구를 인용한 blocker를 돌려준다. 이 문장은 세 갈래 모두에서 넣는다. 템플릿에는 SPEC ID, 카드 ID, 날짜를 넣지 않는다. 네 역할 addenda는 바꾸지 않는다(자식은 거부를 보지 못함, §A). 바꾸게 되면 `make agents-emit`으로 방출본을 다시 만든다. 루트 `AGENTS.md`에는 이 행이 없으므로 로컬 미러 수정은 없다.

### D5. 키 이름 로더 검증 (REQ-CPP-007)

- 경로: `codex exec --strict-config`. `codex mcp get`은 승인 필드를 보이지 않으므로 되읽기 수단이 못 된다(§A).
- 픽스처: git 저장소가 **아닌** 임시 디렉터리 + 그 경로의 `CODEX_HOME` 신뢰 항목 + 그 안 `.codex/config.toml`. 이 형태에서는 설정 적재가 시작 게이트보다 먼저 일어나고(§A: 오타는 적재 오류, 정상 키는 게이트 문구), 게이트가 세션을 막으므로 네트워크 요청도 없다. 제공자도 접속 불가 주소로 준다(이중 안전).
- 입력: 테스트가 손으로 쓴 문자열이 아니라 제품 함수가 만든 설정 바이트(채택 갈래의 writer 출력). A2에서는 opt-in을 **켠** writer 출력을 쓴다(D13). A1·A2에서는 로더에 넣는 바이트에 도구별 표가 정확히 한 번 있음을 테스트가 먼저 단언한다(하위 테스트 `generated_config_has_table`). N 갈래에서도 writer 출력의 기존 키(`command`, `args`, `env_vars`, `default_tools_approval_mode`)를 같은 방식으로 검증한다.
- 환경 변수 이름(`MOAI_CODEX_BIN`, `MOAI_CODEX_PREAPPROVAL_LIVE`, `MOAI_T1172_EVIDENCE_DIR`)은 테스트 전용이므로 기존 `envT1143EvidenceDir`처럼 해당 패키지 테스트 파일의 이름 있는 상수로 둔다(D26). 제품 코드가 읽는 환경 변수는 새로 만들지 않는다.
- 양성 대조: 같은 픽스처에서 `approval_modex`로 바꾼 설정이 `unknown configuration field` + 그 점 경로로 보고되어야 한다. 보고되지 않으면 테스트는 실패한다(skip 아님). 프로젝트 층이 아예 적재되지 않는 픽스처였다면 여기서 드러난다.
- 판정 기준: 정상 설정 → `Not inside a trusted directory` 게이트에 닿고, 로더 오류 접두사와 `unknown configuration field`·`unknown variant`는 없다. 오타 설정은 두 형식만 허용한다: (1) 0.157.0의 한 행 `Error: … unknown configuration field \`mcp_servers.moai.tools.codex_role_audit.approval_modex\``; (2) 과거 원자료의 첫 행 `Error loading config.toml:` 바로 다음 행에 같은 진단. 잘못된 enum도 같은 형식 안에서 `unknown variant \`bogus\``를 요구한다. 세 경우 모두 exit 1이므로 exit만 판정하지 않는다. 접두사만 있는 일반 오류나 떨어진 위치의 진단을 붙인 무관한 로그는 실패시킨다.
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
| AC-CAR-010·011의 공통 입력 5종(`codex-home-config.toml`, `project-config.toml`, `repo.txt`, `moai-build.txt`, `root-state.txt`)을 `.moai/reports/t1172/car010/inputs/`·`car011/inputs/`에 반출하고, 인자·프롬프트는 호출 라벨마다 `inputs/<라벨>/argv.txt`·`prompt.txt`로 반출한다(car011: `mission-governor`·`super-advisor`, car010: `direct`(a)·`parent`(b)). 매니페스트 하나가 공통 5종과 모든 라벨 파일을 담고, 그 반출은 그 묶음의 시작 검사와 첫 호출보다 앞선다. AC-CAR-010 부모의 `--json` 출력 전체는 `car010/parent.jsonl`로 남긴다 | 이월 테스트는 역할·경로마다 인자(`developer_instructions`)와 과제가 다른 호출을 여러 번 띄운다(D-N11). 한 벌만 반출하면 실제 호출과 대조할 수 없다. AC-CPP-013이 각 LIVE 행을 자기 라벨의 파일과만 비교하고, decoy 표의 `enabled_tools`, 프로젝트 층 도구별 표, 부모의 `codex_role_audit` 호출 서버가 `moai`뿐임을 확인한다(이월 판정식은 건드리지 않음) |
| 이월 호출마다 그 호출 직전에 픽스처 루트를 초기화한다(D3의 초기화 대상 보호를 거쳐 `git clean -ffdx` 뒤 반출된 설정 재기록) | 앞 호출이 남긴 파일(예: `audit-probe-<역할>.txt`, `.moai/state/config-cache.json`)이 다음 호출의 입력이 되지 않게 한다. 초기화 뒤 루트 상태는 반출본 `root-state.txt`와 같다 |
| AC-CAR-010·011의 시작 검사·LIVE 행을 공용 `ledger.json`에 `fixture` `car010`·`car011`로 적는다. 하네스가 띄운 호출은 `bound_by: test`, launcher가 띄운 자식은 `bound_by: launcher` | AC-CPP-004의 순서 검사와 AC-CPP-014의 상한 검사가 네 픽스처 모두에 걸린다(D12, D16) |

증거 디렉터리 환경 변수 이름 `MOAI_T1143_EVIDENCE_DIR`는 기존 테스트의 상수이므로 이름은 그대로 두고 값만 `../../.moai/reports/t1172`로 준다(치환 결과와 같음).

**N 갈래에서 AC-CAR-010.** (b) 경로가 승인 거부로 막히므로 실행하지 않고 `NOT_RUN — pre-approval not adopted`로 적는다. AC-CAR-011은 launcher를 직접 부르므로 승인 경로와 무관하게 돌린다.

## §C Pre-flight (run 착수 전)

1. `git -C <worktree> rev-parse --show-toplevel`이 이 워크트리, 브랜치 `WT-codex-preapproval-probe`.
2. AC-CPP-001 `true`(보안 수리가 조상).
3. `codex --version` = `codex-cli 0.157.0`. 다르면 멈추고 보고한다. §A의 0.156.1 측정은 과거 기준이므로 M1-a의 비모델 시작 검사와 로더 게이트를 0.157.0으로 다시 통과시킨 뒤에만 LIVE를 시작한다.
4. 로그인 파일 sha256을 적는다(LIVE 전후 비교용).
5. `make build`로 moai를 빌드하고 커밋·sha256을 적는다.

## §D LIVE 호출 상한 (REQ-CPP-004, REQ-CPP-010)

첫 LIVE 호출 전에 이 표를 `.moai/reports/t1172/verdict.md`에 옮겨 적는다. 단위는 `codex exec` 프로세스 하나다.

| 항목 | 시도당 호출 수 | 재실행(리드가 `invalidate` 행을 적은 뒤 1회, 시도 2) | 호출당 timeout | 시도당 벽시계 창 | 모델 턴 |
|---|---|---|---|---|---|
| 판별 프로브(대조 1 + 처치 1) | 2 | +2 | 330 s | 900 s | 호출당 사용자 턴 1 |
| AC-CAR-010 | 3 | +3 | 330 s(테스트가 띄운 호출), 자식은 launcher 상한 `config.DefaultCodexAuditTimeout` | 1100 s(`-timeout=1200s` 안) | 호출당 1 |
| AC-CAR-011 | 2 | +2 | 330 s | 800 s(`-timeout=900s` 안) | 호출당 1 |
| **절대 상한(모든 시도 합계)** | **14** | | | | |

- 시작 검사(REQ-CPP-003)는 LIVE가 아니다. `ledger.json`에 `kind: startup`으로 적고 상한 8회(픽스처 넷 × 1, 재실행 대비 × 2), 호출당 `timeout 20`(AC-CPP-014는 정리 여유 5 s를 더해 25 s까지 허용). 모델 끝점에 닿지 않는 것이 조건이다. 비모델 `error` 항목 밖의 `item.*` 이벤트가 한 번이라도 보이면 즉시 멈추고 `ABORTED`로 적는다.
- 상한은 장부의 정수 시각으로 AC-CPP-014가 판정한다. 호출 수와 벽시계 창은 **시도마다 따로** 잰다(판별 시도당 대조 1 + 처치 1·900 s, car010 시도당 테스트 호출 2 + launcher 자식 1·1100 s, car011 시도당 2·800 s). 묶음마다 시도는 최대 2(재실행 1회). 호출당 330 s는 테스트가 띄운 모든 호출에 모든 시도에서 적용된다. 절대 상한 14는 무효 시도를 포함한 전체 LIVE 행 수다.
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

- M1-a (LIVE 없음): 픽스처 생성·반출·팔 비교·시작 검사를 하는 하네스를 `internal/cli`의 LIVE 게이트 테스트로 만든다(환경 변수 `MOAI_CODEX_PREAPPROVAL_LIVE=1`, `MOAI_T1172_EVIDENCE_DIR`가 없으면 `NOT_RUN`으로 skip). 반출·비교 로직은 LIVE 없이 결정적으로 시험한다. 시작 검사를 넷 픽스처(판별 대조, 판별 처치, AC-CAR-010 부모, AC-CAR-011)에 돌려 moai·decoy 기동 수를 적는다. 감사 측정으로 기동 1이 확인되어 있으므로, 여기서 0이 나오면 픽스처 결함으로 보고 LIVE 없이 멈추고 리드에게 보고한다(대체 기준 없음). 게이트 거부 경로의 가짜 codex 테스트(`TestCodexPreApprovalGateRefusals`, AC-CPP-012)도 이 마일스톤에서 만든다. 하위 테스트는 다섯이다: `missing_input`, `arm_diff_exceeds`, `startup_check_fails`, `clean_refuses_non_fixture_root`(D-N6), `kills_only_recorded_pids`(정리 경로가 하네스가 기록한 pid만 종료함을 가짜 프로세스로 단언, D-N8).
- M1-b (LIVE 2): 판별 프로브 두 호출. `outcome.txt` 작성.
- 선행 조건: `OPERATOR-CONFIRM` 줄(D1-sub·D2 확정)은 Kickoff에서 이미 기록되어 있다(§B 확정 기록, D-N15). M1-b의 첫 LIVE는 그 뒤에 시작한다.
- 운영자 결정 지점: `outcome.txt`가 `RAN`이고 대조가 `REFUSED`이면 D1(A1/A2)을 운영자가 정하고 리드가 `OPERATOR-DECISION` 줄을 적는다. 아니면 N. D1-sub·D2는 여기서 다시 정하지 않는다(Kickoff 확정 값이 바뀌어야 하면 SPEC 개정 + 재감사).

### M2 — 조건부 방출과 로더 검증 (Priority High)

- 채택 갈래에 맞춰 `internal/codexwiring`의 writer와 정규형 목록을 바꾼다(N이면 코드 변경 없음). 기존 표 바이트 불변 테스트 유지.
- 키 이름 로더 테스트(D5)를 `internal/codexwiring`에 둔다. N 갈래에서도 둔다(기존 키 검증 + 양성 대조).
- A1·A2 갈래: `moai doctor`의 WARNING과 수동 추가 안내(D2 확정값), 그 테스트 `TestDoctorCodexPreApprovalWarning`(`internal/cli`). A2: 설정 섹션 불리언 키(D1-sub 확정값).
- 방출 테스트는 writer 출력 바이트에서 `[mcp_servers.moai.tools.*]` 헤더 집합을 직접 단언한다(하위 테스트 `tool_tables_exact`, D22).
- 재측정 범위: `./internal/codexwiring/...`, `./internal/cli -run 'Doctor|Codex'`.

### M3 — 거부 시 지시 (Priority Medium)

- `AGENTS.md.tmpl` 행 수정(D4). `internal/template/agentemit` 지시면 테스트 갱신. 중립성 판정.
- 재측정 범위: `./internal/template/agentemit/...`, `./internal/template/...` 중 AGENTS 렌더 테스트, `./internal/config -run 'Budget'`(`TestCodexNestedTemplateDiscoveryBudget`가 `AGENTS.md.tmpl` 크기를 잰다, D21).
- 행에 더할 문장은 REQ-CPP-008이 고정한 그대로다: "When the call is refused with `MCP tool call requires approval, but approval policy is never`, do not skip the audit and do not fall back to `spawn_agent`; return a blocker that quotes the refusal text". 문장은 행 끝 ` |` 바로 앞에 `. `와 함께 붙이고, 그 밖의 글자는 행에 더하지 않는다. AC-CPP-010은 행 전체가 "베이스 커밋 `0356e8117`의 행 + `. ` + 고정 문장 + ` |`"와 바이트 같은지를 본다(D-N12 — 앞뒤에 부정·예외 구절을 붙인 변형은 모두 거짓).

### M4 — 이월 LIVE (Priority Medium)

- AC-CAR-011 (LIVE 2): 모든 갈래. 두 역할 호출의 인자·프롬프트를 라벨별로 반출하고, 호출마다 루트를 초기화한다(D6, D-N11).
- AC-CAR-010 (LIVE 3): A1/A2 갈래만. N이면 `NOT_RUN`. 순서(D-N2): M2 writer 변경이 끝난 뒤 car010 입력을 채택 갈래의 writer 출력으로 **다시 반출**하고(`car010/export-manifest.json` 갱신), car010 시작 검사를 **다시** 돌린 다음 LIVE를 시작한다. M1-a의 car010 시작 검사는 decoy 기동을 일찍 재는 용도로 남고, AC-CPP-004는 마지막 시작 검사만 대조한다.
- 픽스처 수정은 D6.

### 마지막 — 기계적 정리 (Priority Low)

- progress.md §E 기록은 run·sync 담당 에이전트가 한다. 이 plan은 자리만 만든다.

## §G 위험

| 위험 | 대응 |
|---|---|
| 실제 픽스처에서 시작 검사가 moai 기동 0을 보인다(감사 측정은 1) | 픽스처 결함으로 보고 LIVE 없이 멈추고 리드에게 보고 |
| 처치 결과가 moai 문구도 거부 문구도 아니다(Codex timeout 등) | `RAN`으로 세지 않는다. 리드가 `INVALID` 여부를 기록 |
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
