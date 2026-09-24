---
id: SPEC-CODEX-PREAPPROVAL-PROBE-001
document: acceptance
created: 2026-09-25
updated: 2026-09-25
author: manager-spec
card: t1172
---

# Acceptance — SPEC-CODEX-PREAPPROVAL-PROBE-001

## §A 판정 규칙

- 모든 AC는 명령의 마지막 출력이 정확히 `true`일 때만 PASS다. 그 밖의 출력, 명령 실패, 증거 파일 부재는 FAIL이다.
- 결과 어휘는 선행 SPEC과 같다. `NOT_RUN` = 실행되지 않았거나 판정 대상 성질이 측정되지 않음. `INVALID` = 리드가 측정 결함(하네스·픽스처·프롬프트)으로 기록한 실행, 증거는 보존. `ABORTED` = 상한 때문에 멈춤. `FAIL` = 측정되었고 판정식이 `true`가 아님. 판별 프로브 팔의 결과 어휘(`REFUSED`·`RAN`·`NOT MEASURED`)는 `plan.md` §D3에 정의한다. `SKIP`, `NOT_RUN`, `NOT MEASURED`, `INVALID`, `ABORTED`는 어떤 경우에도 PASS가 아니다.
- 테스트 이름은 run 단계에서 만들 이름이다. 이름을 바꾸면 이 파일의 명령도 함께 고친다. 이름이 없으면 pass 수가 0이 되어 FAIL이다. 판정식은 빈 입력(이벤트 0개, 증거 파일 부재)에서 `true`를 내지 않아야 하며, run 단계에서 판정식마다 빈 입력으로 한 번 돌려 확인하고 그 명령과 출력을 `.moai/reports/t1172/run-checks/empty-input.md`에 남긴다.
- `internal/cli` 명령은 kanban·factory 환경 변수를 같은 호출 안에서 지운다(`unset … && go test …`).
- 비교 기준은 리터럴 `develop...HEAD`와 베이스 커밋 `0356e8117`이다. git·go 명령에 실행 중 계산한 값을 넘기지 않는다.
- 증거는 `.moai/reports/t1172/`에 남긴다. LIVE AC(AC-CPP-005, AC-CAR-010, AC-CAR-011)는 결정적 AC와 따로 집계한다. LIVE 호출의 단위는 `codex exec` 프로세스 하나이고 상한은 `plan.md` §D가 정한다. 이월된 두 AC의 본문이 가리키는 "`plan.md` §D"는 이 SPEC의 `plan.md` §D로 읽는다.
- 증거 JSON은 파일로 남기고 표준 출력에는 `<TAG>_SHA256 <64자 hex>` 한 줄만 찍는다. 판정은 파일 해시를 다시 재어 태그 줄과 같을 때만 내용을 본다(선행 SPEC과 같은 방식).

**판별 프로브 증거 배치** (`.moai/reports/t1172/discriminator/`):

| 경로 | 내용 |
|---|---|
| `inputs/{control,treatment}/codex-home-config.toml` | 그 팔의 `CODEX_HOME/config.toml` 전문(신뢰 항목 포함) |
| `inputs/{control,treatment}/project-config.toml` | 그 팔의 프로젝트 `.codex/config.toml` 전문 |
| `inputs/{control,treatment}/repo.txt` | `is_repo=true` 한 줄과 `head=<40자 hex>` 한 줄 |
| `inputs/{control,treatment}/moai-build.txt` | `commit=<hex>` 한 줄과 `sha256=<64자 hex>` 한 줄(MCP 서버로 쓴 moai 바이너리) |
| `inputs/{control,treatment}/argv.txt` | 인자 벡터 전문, 인자 하나당 한 줄 |
| `inputs/{control,treatment}/prompt.txt` | 프롬프트 전문 |
| `export-manifest.json` | `exported_at`(UTC RFC 3339) |
| `arm-diff.txt`, `arm-diff.meta.json` | `diff -r .moai/reports/t1172/discriminator/inputs/control .moai/reports/t1172/discriminator/inputs/treatment` 출력 원문, `written_at` |
| `ledger.json` | 호출 배열. 원소마다 `kind`(`live` 또는 `startup`), `label`, `started_at`, `ended_at`, `argv`, `exit`, `auth_sha256_before`, `auth_sha256_after`, `class` |
| `evidence.json`, `outcome.txt` | 판별 결과. `outcome.txt`는 `REFUSED`·`RAN`·`NOT MEASURED`·`NOT DISCRIMINATING` 중 한 줄 |

시작 검사 원자료는 `.moai/reports/t1172/startup/<fixture>.jsonl`(표준 출력), `<fixture>.err`(표준 오류), `<fixture>.launches.json`(`{"moai":N,"decoy":N}`)이다. fixture 이름은 `disc-control`, `disc-treatment`, `car010`, `car011`이다.

## §B 이 SPEC의 AC

### AC-CPP-001 — 선행 보안 수리가 조상이다 (maps REQ-CPP-011)

**Given** 이 브랜치의 HEAD, **When** t1143 보안 수리 커밋 `51d3e5be6`과 그 병합 `a0b78213d`의 조상 관계를 보면, **Then** 둘 다 HEAD의 조상이다.

```bash
git merge-base --is-ancestor 51d3e5be6 HEAD && git merge-base --is-ancestor a0b78213d HEAD && echo true
```

### AC-CPP-002 — 판별 LIVE 전에 입력이 모두 반출되었다 (maps REQ-CPP-001)

**Given** 판별 프로브가 끝난 증거 디렉터리, **When** 두 팔의 입력 파일 여섯 종과 반출 시각, 장부의 LIVE 시작 시각을 보면, **Then** 열두 파일이 모두 비어 있지 않고, 두 팔 모두 저장소이며 HEAD와 빌드가 기록되어 있고, `CODEX_HOME` 설정에 신뢰 항목이 있고, 모든 LIVE 호출의 시작 시각이 반출 시각보다 늦다.

```bash
D=.moai/reports/t1172/discriminator && [ "$(for a in control treatment; do for f in codex-home-config.toml project-config.toml repo.txt moai-build.txt argv.txt prompt.txt; do [ -s "$D/inputs/$a/$f" ] && echo ok; done; done | grep -c ok)" = 12 ] && [ "$(cat $D/inputs/control/repo.txt $D/inputs/treatment/repo.txt | grep -cx 'is_repo=true')" = 2 ] && [ "$(cat $D/inputs/control/repo.txt $D/inputs/treatment/repo.txt | grep -Ecx 'head=[0-9a-f]{40}')" = 2 ] && [ "$(cat $D/inputs/control/moai-build.txt $D/inputs/treatment/moai-build.txt | grep -Ecx 'sha256=[0-9a-f]{64}')" = 2 ] && [ "$(cat $D/inputs/control/moai-build.txt $D/inputs/treatment/moai-build.txt | grep -Ecx 'commit=[0-9a-f]{7,40}')" = 2 ] && [ "$(cat $D/inputs/control/codex-home-config.toml $D/inputs/treatment/codex-home-config.toml | grep -Ec '^trust_level = "trusted"$')" = 2 ] && jq -e --slurpfile l $D/ledger.json '(.exported_at|type)=="string" and (.exported_at as $t | ($l[0]|map(select(.kind=="live"))) as $live | ($live|length)>=1 and ($live|all(.started_at > $t)))' $D/export-manifest.json >/dev/null && echo true
```

음성·변이(run 단계에서 `false` 확인): 파일 하나 삭제, `repo.txt`에 `is_repo=false`, 신뢰 항목 제거, LIVE 시작 시각을 반출 시각보다 앞으로, 장부에 LIVE 항목 0개.

### AC-CPP-003 — 두 팔의 차이는 도구별 표 하나다 (maps REQ-CPP-002)

**Given** 반출된 두 팔의 입력, **When** 차이를 다시 계산해 기록본과 비교하면, **Then** 다시 계산한 차이가 기록본과 바이트 같고, 차이는 `project-config.toml` 한 파일의 추가 줄뿐이며 그 추가 줄은 `[mcp_servers.moai.tools.codex_role_audit]`, `approval_mode = "approve"`, 그리고 선택적인 빈 줄 하나다. 기록 시각은 모든 LIVE 호출보다 앞선다.

```bash
D=.moai/reports/t1172/discriminator && diff -r .moai/reports/t1172/discriminator/inputs/control .moai/reports/t1172/discriminator/inputs/treatment > $D/arm-diff.recheck; cmp -s $D/arm-diff.recheck $D/arm-diff.txt && [ "$(grep -Ecv '^diff -r \.moai/reports/t1172/discriminator/inputs/control/project-config\.toml \.moai/reports/t1172/discriminator/inputs/treatment/project-config\.toml$|^[0-9]+a[0-9]+(,[0-9]+)?$|^> $|^> \[mcp_servers\.moai\.tools\.codex_role_audit\]$|^> approval_mode = "approve"$' $D/arm-diff.txt)" = 0 ] && [ "$(grep -c '^diff -r ' $D/arm-diff.txt)" = 1 ] && [ "$(grep -cx '> \[mcp_servers\.moai\.tools\.codex_role_audit\]' $D/arm-diff.txt)" = 1 ] && [ "$(grep -cx '> approval_mode = "approve"' $D/arm-diff.txt)" = 1 ] && [ "$(grep -cx '> ' $D/arm-diff.txt)" -le 1 ] && jq -e --slurpfile l $D/ledger.json '(.written_at|type)=="string" and (.written_at as $t | ($l[0]|map(select(.kind=="live"))) as $live | ($live|length)>=1 and ($live|all(.started_at > $t)))' $D/arm-diff.meta.json >/dev/null && echo true
```

음성·변이: `argv.txt`에 인자 하나 추가, `CODEX_HOME` 설정에 줄 하나 추가, `prompt.txt` 한 글자 변경, `repo.txt` HEAD 변경, 처치 표의 값 `auto`, 기록본과 재계산본 불일치, 빈 차이(처치에 표 없음) — 모두 `false`.

### AC-CPP-004 — 모델 요청 없는 시작 검사를 LIVE 전에 통과했다 (maps REQ-CPP-003)

**Given** 네 픽스처(`disc-control`, `disc-treatment`, `car010`, `car011`)의 시작 검사 원자료와 장부, **When** 각 원자료를 보면, **Then** 넷 모두 `thread.started`가 정확히 하나, `item.` 이벤트 0개, 표준 오류에 설정 적재 오류·모르는 필드·신뢰 게이트 문구가 없고, moai 기동 1회 이상이며, `car010`은 decoy 기동도 1회 이상이다. 장부의 `startup` 항목은 8개 이하이고, 각 픽스처의 마지막 시작 검사가 그 픽스처를 쓰는 첫 LIVE 호출보다 먼저 끝났다.

```bash
S=.moai/reports/t1172/startup && [ "$(for f in disc-control disc-treatment car010 car011; do [ "$(grep -c '"type":"thread.started"' $S/$f.jsonl)" = 1 ] && [ "$(grep -c '"type":"item\.' $S/$f.jsonl)" = 0 ] && ! grep -Eq 'Error loading config|unknown configuration field|Not inside a trusted directory' $S/$f.err && jq -e '.moai>=1' $S/$f.launches.json >/dev/null && echo ok; done | grep -c ok)" = 4 ] && jq -e '.decoy>=1' $S/car010.launches.json >/dev/null && jq -e '(map(select(.kind=="startup"))|length) as $n | $n>=4 and $n<=8 and (map(select(.kind=="startup" and (.label|startswith("disc-"))))|map(.ended_at)|max) as $s | (map(select(.kind=="live" and (.label|startswith("disc-"))))|map(.started_at)|min) as $l | ($s|type)=="string" and ($l|type)=="string" and $s < $l' .moai/reports/t1172/discriminator/ledger.json >/dev/null && echo true
```

음성·변이: `thread.started` 0개(게이트 거부), `item.completed` 1개(모델 응답이 있었음), 표준 오류에 `Not inside a trusted directory`, moai 기동 0, car010 decoy 기동 0, 시작 검사가 LIVE 뒤에 끝남 — 모두 `false`.

### AC-CPP-005 — [LIVE] 판별 프로브가 유효한 측정을 냈다 (maps REQ-CPP-004, REQ-CPP-010)

PASS의 뜻: 사전 승인이 통한다는 것이 아니라, 두 팔 모두 전제(moai/`codex_role_audit` 호출 1개 이상)를 만족하고 대조가 거부를 재현한 **판별 가능한 측정**이 한 번 이루어졌다는 것이다. 처치 결과가 `REFUSED`여도 PASS다.

**Given** §A의 증거 배치와 `plan.md` §D의 상한(판별 2회), **When** 대조 팔과 처치 팔을 차례로 한 번씩 돌리면, **Then** 호출 수 2, 중단 없음, 반출·차이·시작 검사 통과가 증거에 기록되고, 대조는 `REFUSED`이며 오류 문구가 거부 원문과 같고, 처치는 `REFUSED` 또는 `RAN`이며, 두 팔 모두 자식 프로세스 0개·launch record 0개이고, `outcome.txt`가 처치 결과와 같고, 로그인 파일 해시가 전후 같다.

```bash
mkdir -p .moai/reports/t1172/discriminator && rm -f .moai/reports/t1172/discriminator/evidence.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && MOAI_CODEX_PREAPPROVAL_LIVE=1 MOAI_T1172_EVIDENCE_DIR=../../.moai/reports/t1172 go test -json ./internal/cli -run '^TestCodexPreApprovalDiscriminatorLive$' -count=1 -timeout=1000s > .moai/reports/t1172/discriminator/live.jsonl
```

판정:

```bash
D=.moai/reports/t1172/discriminator && shasum -a 256 $D/evidence.json > $D/evidence.sha && jq -se --rawfile sha $D/evidence.sha --rawfile oc $D/outcome.txt --slurpfile ev $D/evidence.json '($sha|.[0:64]) as $h | ($ev[0]) as $e | ($oc|rtrimstr("\n")) as $o | "MCP tool call requires approval, but approval policy is never" as $refusal | ([.[]|select(.Action=="pass" and .Test=="TestCodexPreApprovalDiscriminatorLive")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^CPP005_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^CPP005_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and (($e.codex_version|type)=="string" and ($e.codex_version|test("^[0-9]"))) and $e.invocations==2 and $e.aborted==false and $e.export_complete==true and $e.arm_diff_ok==true and $e.startup_checks.control=="PASS" and $e.startup_checks.treatment=="PASS" and (($e.arms.control.moai_role_audit_calls|type)=="number" and $e.arms.control.moai_role_audit_calls>=1) and (($e.arms.treatment.moai_role_audit_calls|type)=="number" and $e.arms.treatment.moai_role_audit_calls>=1) and $e.arms.control.outcome=="REFUSED" and $e.arms.control.error_text==$refusal and ($e.arms.treatment.outcome=="REFUSED" or $e.arms.treatment.outcome=="RAN") and (if $e.arms.treatment.outcome=="RAN" then ((($e.arms.treatment.result_text|type)=="string") and ($e.arms.treatment.result_text|length)>0 and ($e.arms.treatment.result_text|test("requires approval")|not)) else $e.arms.treatment.error_text==$refusal end) and $e.arms.control.child_processes==0 and $e.arms.treatment.child_processes==0 and $e.arms.control.launch_records==0 and $e.arms.treatment.launch_records==0 and $o==$e.arms.treatment.outcome and (($e.auth_sha256_before|type)=="string" and $e.auth_sha256_before==$e.auth_sha256_after)' $D/live.jsonl
```

음성·변이: 호출 수 1 또는 3, 대조 `RAN`(`NOT DISCRIMINATING`), 어느 팔이든 `moai_role_audit_calls` 0(`NOT MEASURED`), 처치 `RAN`인데 결과 문구가 거부 문구, 자식 프로세스 1, `outcome.txt`와 처치 결과 불일치, 로그인 해시 변화, 테스트 skip — 모두 `false`.

### AC-CPP-006 — 채택 결정이 측정과 운영자 결정에 맞다 (maps REQ-CPP-005)

**Given** `outcome.txt`, `.moai/reports/t1172/adoption.txt`(`default`·`opt-in`·`not-adopted` 중 한 줄), `verdict.md`, **When** 셋을 맞대 보면, **Then** 결과가 `RAN`이면 채택은 `default` 또는 `opt-in`이고 `verdict.md`에 같은 값의 `OPERATOR-DECISION adoption=<값>` 줄이 있으며, 결과가 `RAN`이 아니면 채택은 `not-adopted`다.

```bash
D=.moai/reports/t1172 && o=$(tr -d '\n' < $D/discriminator/outcome.txt) && a=$(tr -d '\n' < $D/adoption.txt) && case "$o" in REFUSED|RAN|"NOT MEASURED"|"NOT DISCRIMINATING") true;; *) false;; esac && { { [ "$o" = RAN ] && { [ "$a" = default ] || [ "$a" = opt-in ]; } && grep -Eq "^OPERATOR-DECISION adoption=$a( |$)" $D/verdict.md; } || { [ "$o" != RAN ] && [ "$a" = not-adopted ]; }; } && echo true
```

음성·변이: 결과 `REFUSED`인데 채택 `default`, 결과 `RAN`인데 운영자 결정 줄 없음, 결과 `RAN`이고 결정 줄은 `opt-in`인데 채택 `default`, `outcome.txt` 값이 목록 밖 — 모두 `false`.

### AC-CPP-007 — 방출이 채택 갈래를 따르고, 이 도구 하나로 한정되며, 도구는 쓰기 도구로 남는다 (maps REQ-CPP-006)

**Given** `adoption.txt`와 이 브랜치의 코드, **When** 방출 테스트를 돌리고 소스를 검사하면, **Then** 방출 테스트가 통과하고, 갈래별 조건이 성립하며, 공통으로 서버 기본 승인 모드는 `writes`로 남고, `codex_role_audit` 밖의 도구별 표는 없고, `codex_role_audit`의 read-only 표시는 바뀌지 않았다.

- `not-adopted`: 제품 소스(`internal/codexwiring`, `internal/template/templates`의 테스트 아닌 파일)에 `tools.codex_role_audit`가 없다.
- `default`: 테스트가 새 표 바이트에 도구별 표가 정확히 한 번 있음을 단언한다(테스트 이름 `TestEnsureMCPTableRoleAuditPreApproval/default`).
- `opt-in`: 테스트가 기본 출력에는 도구별 표가 없고, 켠 출력에는 정확히 한 번 있음을 단언한다(`…/opt_in_off`, `…/opt_in_on`).
- 공통: 기존 사용자 표가 바이트 그대로임을 단언하는 하위 테스트 `…/existing_table_invariant`가 통과한다.

```bash
go test -json ./internal/codexwiring -run '^TestEnsureMCPTableRoleAuditPreApproval$' -count=1 -v > .moai/reports/t1172/ac-cpp-007.jsonl; jq -se --rawfile ad .moai/reports/t1172/adoption.txt '($ad|rtrimstr("\n")) as $a | ([.[]|select(.Action=="pass" and .Test=="TestEnsureMCPTableRoleAuditPreApproval")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select(.Action=="pass" and .Test=="TestEnsureMCPTableRoleAuditPreApproval/existing_table_invariant")]|length)==1 and (if $a=="default" then ([.[]|select(.Action=="pass" and .Test=="TestEnsureMCPTableRoleAuditPreApproval/default")]|length)==1 elif $a=="opt-in" then ([.[]|select(.Action=="pass" and (.Test=="TestEnsureMCPTableRoleAuditPreApproval/opt_in_off" or .Test=="TestEnsureMCPTableRoleAuditPreApproval/opt_in_on"))]|length)==2 elif $a=="not-adopted" then true else false end)' .moai/reports/t1172/ac-cpp-007.jsonl >/dev/null && { ! grep -qx 'not-adopted' .moai/reports/t1172/adoption.txt || ! grep -rl 'tools\.codex_role_audit' internal/codexwiring internal/template/templates | grep -v '_test\.go$' | grep -q . ; } && grep -q 'mcpApprovalMode = "writes"' internal/codexwiring/configtoml.go && ! grep -rhoE 'mcp_servers\.moai\.tools\.[a-z_]+' internal/codexwiring --include='*.go' --exclude='*_test.go' | grep -vx 'mcp_servers\.moai\.tools\.codex_role_audit' | grep -q . && git diff develop...HEAD -- internal/cli/codex_audit_mcp.go | grep -c 'ReadOnlyHint' | grep -qx 0 && echo true
```

음성·변이: `not-adopted`인데 writer에 표가 남음, `mcpApprovalMode`를 `approve`로 바꿈, 다른 도구(`spec_progress`)의 표 추가, `codex_role_audit`의 `ReadOnlyHint`를 `true`로 바꿈, 기존 표 불변 하위 테스트 삭제 — 모두 `false`.

### AC-CPP-008 — 생성 설정의 키 이름이 Codex 로더를 통과하고, 양성 대조가 오타를 잡는다 (maps REQ-CPP-007)

**Given** codex가 설치된 환경, **When** 로더 검증 테스트를 돌리면, **Then** 부모 테스트와 세 하위 테스트(`generated_config`, `positive_control_misspelled_key`, `positive_control_bad_enum`)가 모두 pass이고 skip이 없다. `generated_config`는 제품 writer가 만든 바이트를 쓰며, 양성 대조는 `unknown configuration field` 문구와 정확한 점 경로 `mcp_servers.moai.tools.codex_role_audit.approval_modex`를 확인한다. 이 오타 표는 채택 갈래와 무관하게 테스트가 writer 출력 뒤에 덧붙여 만든다(N 갈래에서도 같은 대조를 쓴다).

```bash
go test -json ./internal/codexwiring -run '^TestGeneratedConfigKeysLoadInCodex$' -count=1 -v | jq -se '([.[]|select(.Action=="pass" and .Test=="TestGeneratedConfigKeysLoadInCodex")]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestGeneratedConfigKeysLoadInCodex/generated_config")]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestGeneratedConfigKeysLoadInCodex/positive_control_misspelled_key")]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestGeneratedConfigKeysLoadInCodex/positive_control_bad_enum")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0'
```

변이(run 단계에서 테스트가 FAIL해야 함을 확인): 양성 대조의 오타 설정을 정상 키로 되돌림(대조가 발화하지 않음 → FAIL), 픽스처에서 `CODEX_HOME` 신뢰 항목 제거(프로젝트 층 미적재 → 대조 불발 → FAIL), writer 출력 대신 고정 문자열 사용(코드 검사로 확인).

### AC-CPP-009 — codex가 없으면 이유를 적고 skip하며, pass로 세지 않는다 (maps REQ-CPP-007)

**Given** codex 실행 파일을 찾을 수 없게 한 환경, **When** 로더 검증 테스트를 돌리면, **Then** 부모 테스트는 skip이고 pass가 아니며, 출력에 `CODEX_NOT_INSTALLED`가 있다.

```bash
MOAI_CODEX_BIN=/nonexistent/codex go test -json ./internal/codexwiring -run '^TestGeneratedConfigKeysLoadInCodex$' -count=1 -v | jq -se '([.[]|select(.Action=="skip" and .Test=="TestGeneratedConfigKeysLoadInCodex")]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestGeneratedConfigKeysLoadInCodex")]|length)==0 and ([.[]|select((.Output//"")|test("CODEX_NOT_INSTALLED"))]|length)>=1'
```

### AC-CPP-010 — 지시면이 거부 시 할 일을 적는다 (maps REQ-CPP-008, F4)

**Given** `AGENTS.md.tmpl`의 `audit-verdict-file` 행, **When** 행을 읽고 지시면 테스트와 중립성 검사를 돌리면, **Then** 행에 승인 거부 문구(`requires approval`), `blocker`, `skip`, `spawn_agent`가 모두 있고, `internal/template/agentemit` 테스트에 실패가 없으며, 이 브랜치가 템플릿에 더한 줄에 SPEC ID·카드 ID·날짜가 없다.

```bash
grep -c '^| audit-verdict-file |' internal/template/templates/AGENTS.md.tmpl | grep -qx 1 && grep '^| audit-verdict-file |' internal/template/templates/AGENTS.md.tmpl > .moai/reports/t1172/ac-cpp-010-row.txt && grep -q 'requires approval' .moai/reports/t1172/ac-cpp-010-row.txt && grep -q 'blocker' .moai/reports/t1172/ac-cpp-010-row.txt && grep -q 'skip' .moai/reports/t1172/ac-cpp-010-row.txt && grep -q 'spawn_agent' .moai/reports/t1172/ac-cpp-010-row.txt && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test -json ./internal/template/agentemit/... -count=1 | jq -se '([.[]|select(.Action=="fail")]|length)==0 and ([.[]|select(.Action=="pass" and .Test==null)]|length)>=1' | grep -qx true && git diff --name-only develop...HEAD -- internal/template/templates | grep -q . && ! git diff develop...HEAD -- internal/template/templates | grep -E '^\+[^+]' | grep -Eq 'SPEC-[A-Z][A-Z0-9-]*-[0-9]{3}|\bt[0-9]{3,4}\b|20[0-9]{2}-[0-9]{2}-[0-9]{2}' && echo true
```

음성·변이: 행에서 `blocker` 문장 삭제, 추가 줄에 `t1172` 삽입, 추가 줄에 날짜 삽입, 템플릿 변경 없음 — 모두 `false`.

### AC-CPP-011 — 이월 AC의 본문·판정식은 선언한 경로 치환 하나만 다르다 (maps REQ-CPP-009)

**Given** 베이스 커밋 `0356e8117`의 선행 `acceptance.md` 232–256행(AC-CAR-010 본문)과 262–276행(AC-CAR-011 본문), **When** 두 구간에 `s#reports/t1143#reports/t1172#g` 하나만 적용해 §C의 표시(`<!-- carried:begin … -->` / `<!-- carried:end … -->`, 줄 전체 일치) 사이 줄과 비교하면, **Then** 두 블록 모두 바이트 같고, 치환 뒤 블록에 `reports/t1143`이 없으며 `reports/t1172`가 20번 나온다(치환이 실제로 일어났음을 보이는 양성 대조).

```bash
mkdir -p .moai/reports/t1172 && git show 0356e8117:.moai/specs/SPEC-CODEX-AUDIT-READONLY-001/acceptance.md | sed -n '232,256p' | sed 's#reports/t1143#reports/t1172#g' > .moai/reports/t1172/carry-010.src && git show 0356e8117:.moai/specs/SPEC-CODEX-AUDIT-READONLY-001/acceptance.md | sed -n '262,276p' | sed 's#reports/t1143#reports/t1172#g' > .moai/reports/t1172/carry-011.src && awk '/^<!-- carried:end AC-CAR-010 -->$/{f=0} f; /^<!-- carried:begin AC-CAR-010 -->$/{f=1}' .moai/specs/SPEC-CODEX-PREAPPROVAL-PROBE-001/acceptance.md > .moai/reports/t1172/carry-010.dst && awk '/^<!-- carried:end AC-CAR-011 -->$/{f=0} f; /^<!-- carried:begin AC-CAR-011 -->$/{f=1}' .moai/specs/SPEC-CODEX-PREAPPROVAL-PROBE-001/acceptance.md > .moai/reports/t1172/carry-011.dst && test -s .moai/reports/t1172/carry-010.src && test -s .moai/reports/t1172/carry-011.src && cmp -s .moai/reports/t1172/carry-010.src .moai/reports/t1172/carry-010.dst && cmp -s .moai/reports/t1172/carry-011.src .moai/reports/t1172/carry-011.dst && cat .moai/reports/t1172/carry-010.dst .moai/reports/t1172/carry-011.dst | grep -o 'reports/t1143' | grep -c . | grep -qx 0 && cat .moai/reports/t1172/carry-010.dst .moai/reports/t1172/carry-011.dst | grep -o 'reports/t1172' | grep -c . | grep -qx 20 && echo true
```

## §C 이월 AC (선행 SPEC-CODEX-AUDIT-READONLY-001)

이월 규칙:

- 본문과 판정식은 증거 경로 치환 `reports/t1143` → `reports/t1172` 하나만 다르다(AC-CPP-011이 검사). **판정식의 변경은 이 경로 치환뿐이다.** 이유: 이 카드의 증거 경로가 `.moai/reports/t1172/`이기 때문이다. 환경 변수 이름 `MOAI_T1143_EVIDENCE_DIR`는 기존 테스트 상수이므로 그대로다.
- 선행 파일의 이월 안내 인용문(`> **[카드 t1172로 이월 …`)은 옮기지 않았다. 선행 SPEC에서의 상태 표시이지 AC 본문이 아니기 때문이다.
- 판정식이 읽는 `.moai/reports/t1172/m1-route/route.txt`는 t1143 반출본(`mcp`, sha256 `97c5f37f209cef82840992f29653bda1fa3362aa8273b0df20773ef917a679ee`)을 run 단계에서 복사해 둔다. 복사 뒤 해시가 같아야 한다.
- 픽스처 수정(판정식 아님)은 `plan.md` §D6: decoy의 `enabled_tools = ["spec_progress"]`, 프로젝트 층 moai 표는 채택 갈래의 writer 출력, 입력 반출과 시작 검사, `exec` 통로를 허용하는 부모 프롬프트.
- 실행 조건: AC-CAR-011은 모든 갈래에서 실행한다. AC-CAR-010은 채택이 `default` 또는 `opt-in`일 때만 실행하고, `not-adopted`이면 `NOT_RUN — pre-approval not adopted`로 적는다.
- 재실행: 리드가 `INVALID`로 기록한 실행 뒤 한 번만(`plan.md` §D).

### AC-CAR-010 — [LIVE] launcher 계약, 실행 경로, MCP 비활성화 (이월; maps REQ-CPP-009, REQ-CPP-010)

<!-- carried:begin AC-CAR-010 -->
**Given** 격리된 임시 저장소, 임시 `CODEX_HOME`(로그인 사본과, 기록 래퍼를 명령으로 둔 decoy MCP 서버를 선언한 사용자 층 `config.toml`), 임시 `MOAI_HOME`, 설치된 codex 바이너리, 실제 `moai` 앞에 놓여 `mcp-server` 기동을 기록한 뒤 실제 `moai`로 넘기는 기록 래퍼, 프로젝트 층 `config.toml`에 `sandbox_mode = "workspace-write"`와 `[mcp_servers.moai]`가 있는 상태, 시험용 `plan-auditor` 역할 파일(방출본의 `developer_instructions` 끝에 실행마다 다른 nonce 한 줄을 덧붙인 것), 실행당 호출 예산 정확히 3,
**When** (a) 테스트가 launcher를 직접 불러 `plan-auditor`를 띄우고, "개발자 지시문의 nonce를 되돌리고, 탐침 파일 쓰기 명령을 한 번 실행하고, 결과를 한 줄로 반환하라"를 주며, (b) 부모 `codex exec -s workspace-write` 세션 하나를 띄워 M1에서 측정한 경로(shell 또는 MCP)로 launcher를 불러 `sync-auditor`에 같은 탐침 작업을 시키고 판정 파일을 쓰게 하면,
**Then** 증거 파일 `ac-car-010-evidence.json`에 다음이 기록되고 판정식이 모두 요구한다.

- 공통: codex 버전, 호출 수 3, `aborted` false, 경로 이름(`shell` 또는 `mcp`).
- (a): 세션 sandbox `read-only`(프로젝트 config의 `workspace-write`가 있는데도 — 플래그 우선의 측정), `spawn_agent` 사용 없음, 보낸 nonce와 되돌린 nonce 일치(지시문 전달의 측정), `probe_command_executed` true, `probe_exit_code`가 0이 아닌 수, `probe_exists` false, `write_denied` true, 시도 출력 비어 있지 않음, 감사 프로세스 동안의 MCP 기동 수 moai 0·decoy 0.
- (b): launcher가 띄운 하위 세션 sandbox `read-only`, 위와 같은 정의의 쓰기 거부, 판정 파일 존재, 반환문 sha256과 판정 파일 sha256 일치.
- 양성 대조: (b)의 부모 세션 동안 MCP 기동 수가 moai 1 이상, decoy 1 이상. 이것이 없으면 (a)의 0은 래퍼가 경로에 없어서 생긴 0과 구별되지 않는다. decoy도 기록 뒤 `moai mcp-server`로 넘기는 래퍼다. codex가 설정된 MCP 서버를 세션 시작 때 띄우는지(즉시 기동)는 측정되지 않았다. M1 P-B가 MCP를 켠 부모 세션의 래퍼 기동 수를 기록한다. 0이면(지연 기동) (b) 부모의 작업에 두 서버 각각의 읽기 전용 도구 호출 한 번(`spec_progress`)을 넣는다. 이 판단은 M1 기록(`m1-route/`)에 적고, 호출 수는 바뀌지 않는다.
- 경로 일치: 증거의 `route`가 `.moai/reports/t1172/m1-route/route.txt`의 값과 같다(`route.txt`가 단일 출처).

SKIP 의미: `MOAI_CODEX_ROLE_LIVE=1` 또는 `MOAI_T1143_EVIDENCE_DIR`이 없으면 SKIP하고 `NOT_RUN`이다. 4번째 호출이 필요해지면 시작하지 않고 `ABORTED`를 찍고 실패한다. 재실행은 리드가 `INVALID`로 기록한 실행 뒤 한 번만 허용된다(`plan.md` §D).

실행:

```bash
mkdir -p .moai/reports/t1172 && rm -f .moai/reports/t1172/ac-car-010-evidence.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && MOAI_CODEX_ROLE_LIVE=1 MOAI_T1143_EVIDENCE_DIR=../../.moai/reports/t1172 go test -json ./internal/cli -run '^TestCodexAuditLaunchLiveContract$' -count=1 -timeout=1200s > .moai/reports/t1172/ac-car-010-live.jsonl
```

판정:

```bash
shasum -a 256 .moai/reports/t1172/ac-car-010-evidence.json > .moai/reports/t1172/ac-car-010-evidence.sha && jq -se --rawfile sha .moai/reports/t1172/ac-car-010-evidence.sha --rawfile rt .moai/reports/t1172/m1-route/route.txt --slurpfile ev .moai/reports/t1172/ac-car-010-evidence.json '($sha|.[0:64]) as $h | ($ev[0]) as $e | ($rt|rtrimstr("\n")) as $r | ([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchLiveContract")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^ACCAR010_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^ACCAR010_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and (($e.codex_version|type)=="string" and ($e.codex_version|test("^[0-9]"))) and $e.invocations==3 and $e.aborted==false and ($e.route=="shell" or $e.route=="mcp") and $e.route==$r and $e.direct.role=="plan-auditor" and $e.direct.session_sandbox=="read-only" and $e.direct.used_spawn_agent==false and (($e.direct.nonce_sent|type)=="string" and ($e.direct.nonce_sent|length)>0) and $e.direct.nonce_returned==$e.direct.nonce_sent and $e.direct.probe_command_executed==true and (($e.direct.probe_exit_code|type)=="number" and $e.direct.probe_exit_code!=0) and $e.direct.probe_exists==false and $e.direct.write_denied==true and (($e.direct.attempt_output|type)=="string" and ($e.direct.attempt_output|length)>0) and $e.direct.mcp_launches.moai==0 and $e.direct.mcp_launches.decoy==0 and $e.routed.role=="sync-auditor" and $e.routed.child_session_sandbox=="read-only" and $e.routed.probe_command_executed==true and (($e.routed.probe_exit_code|type)=="number" and $e.routed.probe_exit_code!=0) and $e.routed.probe_exists==false and $e.routed.write_denied==true and $e.routed.verdict_file_exists==true and (($e.routed.returned_sha256//"")|test("^[0-9a-f]{64}$")) and $e.routed.returned_sha256==$e.routed.verdict_file_sha256 and (($e.routed.parent_mcp_launches.moai|type)=="number" and $e.routed.parent_mcp_launches.moai>=1) and (($e.routed.parent_mcp_launches.decoy|type)=="number" and $e.routed.parent_mcp_launches.decoy>=1)' .moai/reports/t1172/ac-car-010-live.jsonl
```

음성·변이(판정식이 `false`여야 하는 합성 입력, run 단계에서 확인): 호출 수 2 또는 4, `route` 빈 값, `route`와 `route.txt` 불일치, `route.txt` 부재, (a) sandbox `workspace-write`, nonce 불일치, `probe_command_executed` false(모델이 명령을 실행하지 않음), `probe_exit_code` 0, `probe_exists` true, `mcp_launches.moai` 1, `mcp_launches.decoy` 1, `parent_mcp_launches.moai` 0(래퍼가 경로에 없는 상황), (b) 해시 불일치, 태그 해시와 파일 해시 불일치, 테스트 skip.
<!-- carried:end AC-CAR-010 -->

### AC-CAR-011 — [LIVE] 나머지 read-only 역할의 쓰기 차단 (이월; maps REQ-CPP-009, REQ-CPP-010)

<!-- carried:begin AC-CAR-011 -->
**Given** AC-CAR-010과 같은 격리 환경, 실행당 호출 예산 정확히 2,
**When** launcher로 `mission-governor`와 `super-advisor`를 한 번씩 띄워 탐침 파일 쓰기 명령을 한 번 실행하고 결과를 한 줄로 반환하게 하면,
**Then** 증거 파일 `ac-car-011-evidence.json`에 호출 수 2와, 두 역할 각각의 세션 sandbox `read-only`, `probe_command_executed` true, 0이 아닌 `probe_exit_code`, 탐침 파일 없음, `write_denied` true, 시도 출력 비어 있지 않음이 기록된다. 재실행 규칙은 AC-CAR-010과 같다.

실행:

```bash
mkdir -p .moai/reports/t1172 && rm -f .moai/reports/t1172/ac-car-011-evidence.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && MOAI_CODEX_ROLE_LIVE=1 MOAI_T1143_EVIDENCE_DIR=../../.moai/reports/t1172 go test -json ./internal/cli -run '^TestCodexAuditLaunchLiveReadOnlyRoles$' -count=1 -timeout=900s > .moai/reports/t1172/ac-car-011-live.jsonl
```

판정:

```bash
shasum -a 256 .moai/reports/t1172/ac-car-011-evidence.json > .moai/reports/t1172/ac-car-011-evidence.sha && jq -se --rawfile sha .moai/reports/t1172/ac-car-011-evidence.sha --slurpfile ev .moai/reports/t1172/ac-car-011-evidence.json '($sha|.[0:64]) as $h | ($ev[0]) as $e | ([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchLiveReadOnlyRoles")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^ACCAR011_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^ACCAR011_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and $e.invocations==2 and $e.aborted==false and ([$e.roles[].role]|sort)==["mission-governor","super-advisor"] and ([$e.roles[]|select(.session_sandbox=="read-only" and .probe_command_executed==true and ((.probe_exit_code|type)=="number" and .probe_exit_code!=0) and .probe_exists==false and .write_denied==true and ((.attempt_output|type)=="string" and (.attempt_output|length)>0))]|length)==2' .moai/reports/t1172/ac-car-011-live.jsonl
```
<!-- carried:end AC-CAR-011 -->

## §D 품질 게이트와 완료 정의

- 결정적 AC: AC-CPP-001, 002, 003, 004, 006, 007, 008, 009, 010, 011. 모두 `true`여야 한다. AC-CPP-002–004는 판별 프로브의 증거를 읽지만 판정 자체는 결정적이다.
- LIVE AC: AC-CPP-005, AC-CAR-010(채택 갈래에서만), AC-CAR-011. LIVE 결과는 기록된 분류 그대로 집계하며 PASS로 부풀리지 않는다.
- 재측정 범위는 변경 패키지로 한정한다: `./internal/codexwiring/...`, `./internal/cli`(해당 테스트 이름), `./internal/template/agentemit/...`. 로컬에서 `go test ./...`를 돌리지 않는다. `internal/cli`를 넓게 돌릴 때는 `-timeout`을 명시한다.
- 린트: 변경 패키지에 `golangci-lint run` 0 issues, `go vet` 통과.
- 템플릿을 바꿨으면 `make build`를 돌린다. 역할 `.md` 또는 `agents-codex.yaml`을 바꿨으면 `make agents-emit`을 돌리고 `make agents-emit-check`가 통과해야 한다.

**완료 정의.** 판별 프로브가 유효한 측정을 한 번 냈고(AC-CPP-005) 그 결과에 따라 채택 갈래가 정해졌으며(AC-CPP-006), 방출·로더 검증·지시면·이월 AC 충실도가 모두 `true`이고, AC-CAR-011이 실행되었으며, AC-CAR-010은 채택 갈래에서 실행되었거나 `not-adopted`로 `NOT_RUN` 사유가 적혀 있을 때. 판별 프로브가 두 번째로 `NOT MEASURED`로 끝나면 이 카드는 "pre-approval effect not measured"로 멈추고, N 갈래의 AC(지시면, 로더 검증, AC-CAR-011)만 닫은 채 리드에게 보고한다.
