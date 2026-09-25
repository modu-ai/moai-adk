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

- 모든 AC는 판정 명령의 마지막 출력이 정확히 `true`일 때만 PASS다. 그 밖의 출력, 명령 실패, 증거 파일 부재는 FAIL이다. `실행:` 블록이 따로 있는 AC는 실행 블록을 먼저 돌리고 판정 블록을 돌린다.
- 결과 어휘는 선행 SPEC과 같다. `NOT_RUN` = 실행되지 않았거나 판정 대상 성질이 측정되지 않음. `INVALID` = 리드가 측정 결함(하네스·픽스처·프롬프트)으로 기록한 실행, 증거는 보존. `ABORTED` = 상한 때문에 멈춤. `FAIL` = 측정되었고 판정식이 `true`가 아님. 판별 프로브 팔의 결과 어휘(`REFUSED`·`RAN`·`NOT MEASURED`)는 `plan.md` §D3에 정의한다. `SKIP`, `NOT_RUN`, `NOT MEASURED`, `INVALID`, `ABORTED`는 어떤 경우에도 PASS가 아니다.
- 테스트 이름은 run 단계에서 만들 이름이다. 이름을 바꾸면 이 파일의 명령도 함께 고친다. 이름이 없으면 pass 수가 0이 되어 FAIL이다. 판정식은 빈 입력(이벤트 0개, 증거 파일 부재)에서 `true`를 내지 않는다. plan 단계에서 각 판정식을 합성 입력의 참·거짓 경우로 돌린 결과는 `.moai/reports/t1172/verdict.md` §Plan에 있다.
- `internal/cli` 명령은 kanban·factory 환경 변수를 같은 호출 안에서 지운다(`unset … && go test …`).
- **판정 명령의 형태 (워크트리 세션 가드).** 경로는 모두 리터럴로 쓴다. `git`이 들어간 판정 명령에는 `$(...)`, `{ … }` 묶음, 반복문을 쓰지 않는다(가드가 거부함을 이 세션에서 확인). `git`이 없는 판정 명령은 그 제약을 받지 않는다. 비교 기준은 리터럴 `develop...HEAD`와 베이스 커밋 `0356e8117`이다. 셸 함수 `diff`(`--color`)를 피하려고 판정식은 `command diff`를 쓴다.
- 증거는 `.moai/reports/t1172/`에 남긴다. LIVE AC(AC-CPP-005, AC-CAR-010, AC-CAR-011)는 결정적 AC와 따로 집계한다. LIVE 호출의 단위는 `codex exec` 프로세스 하나이고 상한은 `plan.md` §D가 정한다. 이월된 두 AC의 본문이 가리키는 "`plan.md` §D"는 이 SPEC의 `plan.md` §D로 읽는다.
- 시각은 모두 정수 epoch 나노초(`*_ns`)로 적고 숫자로 비교한다.
- 증거 JSON은 파일로 남기고 표준 출력에는 `<TAG>_SHA256 <64자 hex>` 한 줄만 찍는다. 판정은 파일 해시를 다시 재어 태그 줄과 같을 때만 내용을 본다.

**증거 배치** (`.moai/reports/t1172/`):

| 경로 | 내용 |
|---|---|
| `discriminator/inputs/{control,treatment}/codex-home-config.toml` | 그 팔의 `CODEX_HOME/config.toml` 전문(신뢰 항목 포함, `model` 키 없음) |
| `discriminator/inputs/{control,treatment}/project-config.toml` | 그 팔의 프로젝트 `.codex/config.toml` 전문 |
| `discriminator/inputs/{control,treatment}/repo.txt` | `is_repo=true` 한 줄과 `head=<40자 hex>` 한 줄 |
| `discriminator/inputs/{control,treatment}/moai-build.txt` | `commit=<hex>` 한 줄과 `sha256=<64자 hex>` 한 줄(MCP 서버로 쓴 moai 바이너리) |
| `discriminator/inputs/{control,treatment}/argv.txt` | 인자 벡터 전문, 인자 하나당 한 줄, 끝에 개행 |
| `discriminator/inputs/{control,treatment}/prompt.txt` | 프롬프트 전문 |
| `discriminator/inputs/{control,treatment}/root-state.txt` | `git clean -ffdx`와 설정 재기록 직후 루트의 `git status --porcelain --ignored` 출력과 `.git` 밖 파일 목록(`plan.md` §D3) |
| `discriminator/export-manifest.json` | `exported_ns`(정수), `files`(`"<팔>/<파일>"` → sha256, 14개) |
| `discriminator/arm-diff.txt`, `arm-diff.meta.json` | 작업 디렉터리 `discriminator/`에서 `diff -r inputs/control inputs/treatment`를 돌린 출력 원문, `written_ns`(정수) |
| `discriminator/evidence.json`, `outcome.txt`, `live.jsonl` | 판별 결과. `outcome.txt`는 `REFUSED`·`RAN`·`NOT MEASURED`·`NOT DISCRIMINATING` 중 한 줄 |
| `car010/inputs/`, `car011/inputs/` | 위 7종과 같은 이름의 반출본. car010의 `codex-home-config.toml`에는 사용자 층 decoy 표가 있다 |
| `car010/parent.jsonl` | AC-CAR-010 부모 세션의 `--json` 출력 전체 |
| `startup/<fixture>.jsonl`, `.err`, `.launches.json` | 시작 검사의 표준 출력, 표준 오류, 기동 수(`{"moai":N,"decoy":N}`). fixture: `disc-control`, `disc-treatment`, `car010`, `car011` |
| `ledger.json` | 시작 검사·LIVE·정지 전부를 담는 한 배열. 행 필드는 `plan.md` §D3 "장부와 해시 연결" |
| `adoption.txt` | `default`·`opt-in`·`not-adopted` 중 한 줄 |
| `m1-route/route.txt` | 이월 판정식이 읽는 경로 값(`mcp`). `plan-checks/t1143-route.txt`의 복사본 |

## §B 이 SPEC의 AC

### AC-CPP-001 — 선행 보안 수리가 조상이다 (maps REQ-CPP-011)

**Given** 이 브랜치의 HEAD, **When** t1143 보안 수리 커밋 `51d3e5be6`과 그 병합 `a0b78213d`의 조상 관계를 보면, **Then** 둘 다 HEAD의 조상이다.

```bash
git merge-base --is-ancestor 51d3e5be6 HEAD && git merge-base --is-ancestor a0b78213d HEAD && echo true
```

### AC-CPP-002 — 판별 LIVE 전에 입력이 모두 반출되었고, LIVE는 반출본을 그대로 썼다 (maps REQ-CPP-001)

**Given** 판별 프로브가 끝난 증거 디렉터리, **When** 두 팔의 입력 7종, 반출 매니페스트, 장부의 판별 LIVE 행을 맞대 보면, **Then** 열네 파일이 모두 비어 있지 않고, 두 팔 모두 저장소이며 HEAD와 빌드가 기록되어 있고, `CODEX_HOME` 설정에 신뢰 항목이 있고, 매니페스트의 파일 해시가 실제 파일 해시와 같고, 판별 LIVE 행이 팔마다 정확히 하나씩이며 모두 반출 뒤에 시작했고, 각 행이 실제로 쓴 인자 벡터가 `argv.txt`와 같고 설정·프롬프트·루트 상태·moai 바이너리 해시가 반출본과 같다.

```bash
[ "$(for a in control treatment; do for f in codex-home-config.toml project-config.toml repo.txt moai-build.txt argv.txt prompt.txt root-state.txt; do [ -s ".moai/reports/t1172/discriminator/inputs/$a/$f" ] && echo ok; done; done | grep -c ok)" = 14 ] && cat .moai/reports/t1172/discriminator/inputs/control/repo.txt .moai/reports/t1172/discriminator/inputs/treatment/repo.txt | grep -cx 'is_repo=true' | grep -qx 2 && cat .moai/reports/t1172/discriminator/inputs/control/repo.txt .moai/reports/t1172/discriminator/inputs/treatment/repo.txt | grep -Ecx 'head=[0-9a-f]{40}' | grep -qx 2 && cat .moai/reports/t1172/discriminator/inputs/control/moai-build.txt .moai/reports/t1172/discriminator/inputs/treatment/moai-build.txt | grep -Ecx 'sha256=[0-9a-f]{64}' | grep -qx 2 && cat .moai/reports/t1172/discriminator/inputs/control/moai-build.txt .moai/reports/t1172/discriminator/inputs/treatment/moai-build.txt | grep -Ecx 'commit=[0-9a-f]{7,40}' | grep -qx 2 && cat .moai/reports/t1172/discriminator/inputs/control/codex-home-config.toml .moai/reports/t1172/discriminator/inputs/treatment/codex-home-config.toml | grep -Ec '^trust_level = "trusted"$' | grep -qx 2 && shasum -a 256 .moai/reports/t1172/discriminator/inputs/control/* .moai/reports/t1172/discriminator/inputs/treatment/* > .moai/reports/t1172/discriminator/inputs.sha && jq -e --slurpfile l .moai/reports/t1172/ledger.json --rawfile s .moai/reports/t1172/discriminator/inputs.sha --rawfile ac .moai/reports/t1172/discriminator/inputs/control/argv.txt --rawfile at .moai/reports/t1172/discriminator/inputs/treatment/argv.txt --rawfile mc .moai/reports/t1172/discriminator/inputs/control/moai-build.txt --rawfile mt .moai/reports/t1172/discriminator/inputs/treatment/moai-build.txt '.exported_ns as $e | ($s|split("\n")|map(select(length>0)|capture("^(?<h>[0-9a-f]{64})  .*/inputs/(?<p>.+)$"))|map({(.p):.h})|add) as $fs | {control:$ac, treatment:$at} as $argv | {control:($mc|capture("sha256=(?<x>[0-9a-f]{64})").x), treatment:($mt|capture("sha256=(?<x>[0-9a-f]{64})").x)} as $bin | ($l[0]|map(select(.kind=="live" and (.fixture|startswith("disc-"))))) as $live | ($e|type)=="number" and .files==$fs and ($fs|length)==14 and ($live|map(.fixture)|sort)==["disc-control","disc-treatment"] and ($live|all(.started_ns > $e)) and ($live|all((.fixture|ltrimstr("disc-")) as $a | ((.argv|join("\n"))+"\n")==$argv[$a] and .inputs_sha256.codex_home_config==$fs[$a+"/codex-home-config.toml"] and .inputs_sha256.project_config==$fs[$a+"/project-config.toml"] and .inputs_sha256.argv==$fs[$a+"/argv.txt"] and .inputs_sha256.prompt==$fs[$a+"/prompt.txt"] and .inputs_sha256.root_state==$fs[$a+"/root-state.txt"] and .inputs_sha256.moai_binary==$bin[$a]))' .moai/reports/t1172/discriminator/export-manifest.json >/dev/null && echo true
```

음성·변이(plan 단계에서 `false` 확인, verdict §Plan): 입력 하나 삭제, 처치 LIVE가 쓴 인자에 `--extra` 추가(반출본과 불일치), LIVE 시작이 반출보다 앞섬.

### AC-CPP-003 — 두 팔의 차이는 도구별 표 하나다 (maps REQ-CPP-002)

**Given** 반출된 두 팔의 입력, **When** 차이를 다시 계산해 기록본과 비교하면, **Then** 다시 계산한 차이가 기록본과 바이트 같고, 차이는 `project-config.toml` 한 파일의 추가 줄뿐이며 그 추가 줄은 `[mcp_servers.moai.tools.codex_role_audit]`, `approval_mode = "approve"`, 그리고 선택적인 빈 줄 하나다. 기록 시각은 모든 판별 LIVE 호출보다 앞선다.

```bash
(cd .moai/reports/t1172/discriminator && command diff -r inputs/control inputs/treatment) > .moai/reports/t1172/discriminator/arm-diff.recheck; cmp -s .moai/reports/t1172/discriminator/arm-diff.recheck .moai/reports/t1172/discriminator/arm-diff.txt && grep -Ecv '^diff -r inputs/control/project-config\.toml inputs/treatment/project-config\.toml$|^[0-9]+a[0-9]+(,[0-9]+)?$|^> $|^> \[mcp_servers\.moai\.tools\.codex_role_audit\]$|^> approval_mode = "approve"$' .moai/reports/t1172/discriminator/arm-diff.txt | grep -qx 0 && grep -c '^diff -r ' .moai/reports/t1172/discriminator/arm-diff.txt | grep -qx 1 && grep -cx '> \[mcp_servers\.moai\.tools\.codex_role_audit\]' .moai/reports/t1172/discriminator/arm-diff.txt | grep -qx 1 && grep -cx '> approval_mode = "approve"' .moai/reports/t1172/discriminator/arm-diff.txt | grep -qx 1 && grep -cx '> ' .moai/reports/t1172/discriminator/arm-diff.txt | grep -Eqx '0|1' && jq -e --slurpfile l .moai/reports/t1172/ledger.json '.written_ns as $t | ($l[0]|map(select(.kind=="live" and (.fixture|startswith("disc-"))))) as $live | ($t|type)=="number" and ($live|length)>=1 and ($live|all(.started_ns > $t))' .moai/reports/t1172/discriminator/arm-diff.meta.json >/dev/null && echo true
```

음성·변이(plan 단계에서 `false` 확인): 처치 `argv.txt`에 인자 하나 추가, 처치 `CODEX_HOME` 설정에 줄 하나 추가, 처치 표의 값 `auto`, 기록본 헤더가 `diff --color -r`, 기록 시각이 LIVE 뒤, 입력 파일 하나 없음.

### AC-CPP-004 — 모델 끝점에 닿지 않는 시작 검사를 픽스처마다 LIVE 전에 통과했다 (maps REQ-CPP-003)

**Given** 네 픽스처(`disc-control`, `disc-treatment`, `car010`, `car011`)의 시작 검사 원자료, 반출된 설정, 장부, **When** 각각을 보면, **Then** 넷 모두 표준 오류 파일이 있고, `thread.started`가 정확히 하나, 비모델 `error` 항목을 뺀 `item.` 이벤트가 0개, 표준 오류에 설정 적재 오류·모르는 필드·신뢰 게이트 문구가 없고, moai 기동이 1회 이상이며, `car010`은 decoy 기동도 1회 이상이다. 네 픽스처의 `CODEX_HOME` 설정에 `model` 키가 없고 어느 장부 행에도 `-m`/`--model`이 없다. 장부의 `startup` 행은 4개 이상 8개 이하이고, 모두 제공자가 `http://127.0.0.1:9/v1`이며 로그인 파일이 없었고 `--strict-config`로 돌았으며, 적재한 설정의 해시가 그 픽스처의 반출본과 같다. 각 픽스처의 마지막 시작 검사는 그 픽스처의 첫 LIVE 호출보다 먼저 끝났다(그 픽스처에 LIVE가 없으면 순서 조건은 적용하지 않는다).

```bash
[ "$(for f in disc-control disc-treatment car010 car011; do test -f ".moai/reports/t1172/startup/$f.err" && grep -c '"type":"thread.started"' ".moai/reports/t1172/startup/$f.jsonl" | grep -qx 1 && jq -se '[.[]|select((.type|startswith("item.")) and (.item.type!="error"))]|length==0' ".moai/reports/t1172/startup/$f.jsonl" | grep -qx true && ! grep -Eq 'Error loading config|unknown configuration field|Not inside a trusted directory' ".moai/reports/t1172/startup/$f.err" && jq -e '.moai>=1' ".moai/reports/t1172/startup/$f.launches.json" >/dev/null && echo ok; done | grep -c ok)" = 4 ] && jq -e '.decoy>=1' .moai/reports/t1172/startup/car010.launches.json >/dev/null && shasum -a 256 .moai/reports/t1172/discriminator/inputs/control/codex-home-config.toml .moai/reports/t1172/discriminator/inputs/control/project-config.toml .moai/reports/t1172/discriminator/inputs/treatment/codex-home-config.toml .moai/reports/t1172/discriminator/inputs/treatment/project-config.toml .moai/reports/t1172/car010/inputs/codex-home-config.toml .moai/reports/t1172/car010/inputs/project-config.toml .moai/reports/t1172/car011/inputs/codex-home-config.toml .moai/reports/t1172/car011/inputs/project-config.toml > .moai/reports/t1172/startup/inputs.sha && ! grep -Eq '^model *=' .moai/reports/t1172/discriminator/inputs/control/codex-home-config.toml .moai/reports/t1172/discriminator/inputs/treatment/codex-home-config.toml .moai/reports/t1172/car010/inputs/codex-home-config.toml .moai/reports/t1172/car011/inputs/codex-home-config.toml && jq -e --rawfile s .moai/reports/t1172/startup/inputs.sha '. as $all | ($s|split("\n")|map(select(length>0)|capture("^(?<h>[0-9a-f]{64})  (?<p>.+)$"))|map({(.p):.h})|add) as $fs | {"disc-control":".moai/reports/t1172/discriminator/inputs/control","disc-treatment":".moai/reports/t1172/discriminator/inputs/treatment","car010":".moai/reports/t1172/car010/inputs","car011":".moai/reports/t1172/car011/inputs"} as $dir | ($all|map(select(.kind=="startup"))) as $st | ($st|length)>=4 and ($st|length)<=8 and ($all|all((.argv|any(.=="-m" or .=="--model"))|not)) and ($st|all(.provider_base_url=="http://127.0.0.1:9/v1" and .auth_file_present==false and (.argv|any(.=="--strict-config")) and (.argv|join(" ")|contains("base_url=\"http://127.0.0.1:9/v1\"")) and .inputs_sha256.codex_home_config==$fs[$dir[.fixture]+"/codex-home-config.toml"] and .inputs_sha256.project_config==$fs[$dir[.fixture]+"/project-config.toml"])) and (["disc-control","disc-treatment","car010","car011"]|all(. as $f | ($st|map(select(.fixture==$f))) as $sf | ($all|map(select(.kind=="live" and .fixture==$f))) as $lf | ($sf|length)>=1 and (($lf|length)==0 or (($sf|map(.ended_ns)|max) < ($lf|map(.started_ns)|min)))))' .moai/reports/t1172/ledger.json >/dev/null && echo true
```

음성·변이(plan 단계에서 확인): `item.completed`(`agent_message`) 1개 → `false`, car011 `.err` 없음 → `false`, car010 시작 검사 행 `auth_file_present: true` → `false`, `CODEX_HOME` 설정에 `model = "gpt-5"` → `false`, car011 시작 검사가 그 LIVE 뒤에 끝남 → `false`. 비모델 `error` 항목(`Model metadata … not found`) 1개만 있는 경우 → `true`(의도된 허용).

### AC-CPP-005 — [LIVE] 판별 프로브가 유효한 측정을 냈다 (maps REQ-CPP-004, REQ-CPP-010)

PASS의 뜻: 사전 승인이 통한다는 것이 아니라, 두 팔 모두 전제(moai/`codex_role_audit` 호출 1개 이상)를 만족하고 대조가 거부를 재현한 **판별 가능한 측정**이 한 번 이루어졌다는 것이다. 처치 결과가 `REFUSED`여도 PASS다.

**Given** §A의 증거 배치와 `plan.md` §D의 상한(판별 2회), **When** 대조 팔과 처치 팔을 차례로 한 번씩 돌리면, **Then** 호출 수 2, 중단 없음, 반출·차이·시작 검사 통과가 증거에 기록되고, 대조는 `REFUSED`이며 오류 문구가 거부 원문과 같고, 처치는 `REFUSED`(거부 원문) 또는 `RAN`이며, `RAN`이면 결과 문구가 moai 접두사 `codex_role_audit: codex audit sync-auditor: `로 시작하고 이어 `destination rejected: ` 또는 `working root rejected: `가 온다. 두 팔 모두 자식 프로세스 0개·launch record 0개이고, `outcome.txt`가 처치 결과와 같고, 로그인 파일 해시가 전후 같다.

실행:

```bash
mkdir -p .moai/reports/t1172/discriminator && rm -f .moai/reports/t1172/discriminator/evidence.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && MOAI_CODEX_PREAPPROVAL_LIVE=1 MOAI_T1172_EVIDENCE_DIR=../../.moai/reports/t1172 go test -json ./internal/cli -run '^TestCodexPreApprovalDiscriminatorLive$' -count=1 -timeout=1000s > .moai/reports/t1172/discriminator/live.jsonl
```

판정:

```bash
shasum -a 256 .moai/reports/t1172/discriminator/evidence.json > .moai/reports/t1172/discriminator/evidence.sha && jq -se --rawfile sha .moai/reports/t1172/discriminator/evidence.sha --rawfile oc .moai/reports/t1172/discriminator/outcome.txt --slurpfile ev .moai/reports/t1172/discriminator/evidence.json '($sha|.[0:64]) as $h | ($ev[0]) as $e | ($oc|rtrimstr("\n")) as $o | "MCP tool call requires approval, but approval policy is never" as $refusal | ([.[]|select(.Action=="pass" and .Test=="TestCodexPreApprovalDiscriminatorLive")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^CPP005_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^CPP005_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and (($e.codex_version|type)=="string" and ($e.codex_version|test("^[0-9]"))) and $e.invocations==2 and $e.aborted==false and $e.export_complete==true and $e.arm_diff_ok==true and $e.startup_checks.control=="PASS" and $e.startup_checks.treatment=="PASS" and (($e.arms.control.moai_role_audit_calls|type)=="number" and $e.arms.control.moai_role_audit_calls>=1) and (($e.arms.treatment.moai_role_audit_calls|type)=="number" and $e.arms.treatment.moai_role_audit_calls>=1) and $e.arms.control.outcome=="REFUSED" and $e.arms.control.error_text==$refusal and ($e.arms.treatment.outcome=="REFUSED" or $e.arms.treatment.outcome=="RAN") and (if $e.arms.treatment.outcome=="RAN" then (($e.arms.treatment.result_text|type)=="string" and ($e.arms.treatment.result_text|test("^codex_role_audit: codex audit sync-auditor: (destination rejected|working root rejected): ")) and ($e.arms.treatment.result_text|test("requires approval")|not)) else $e.arms.treatment.error_text==$refusal end) and $e.arms.control.child_processes==0 and $e.arms.treatment.child_processes==0 and $e.arms.control.launch_records==0 and $e.arms.treatment.launch_records==0 and $o==$e.arms.treatment.outcome and (($e.auth_sha256_before|type)=="string" and $e.auth_sha256_before==$e.auth_sha256_after)' .moai/reports/t1172/discriminator/live.jsonl
```

음성·변이(plan 단계에서 확인): 처치 결과 문구 `tool call timed out after 60s`(plan-audit iter 1의 합성 증거 그대로) → `false`(이전 판정식은 `true`였다). 그 밖에 호출 수 1 또는 3, 대조 `RAN`, 어느 팔이든 `moai_role_audit_calls` 0, 자식 프로세스 1, `outcome.txt` 불일치, 로그인 해시 변화, 테스트 skip — 모두 `false`.

### AC-CPP-006 — 채택 결정이 측정과 운영자 결정에 맞다 (maps REQ-CPP-005)

**Given** `outcome.txt`, `adoption.txt`, `verdict.md`, **When** 셋을 맞대 보면, **Then** 두 파일 값이 허용 목록 안이고, 결과가 `RAN`이면 채택은 `default` 또는 `opt-in`이며 `verdict.md`에 같은 값의 `OPERATOR-DECISION adoption=<값> card=t1172 recorded_by=lead at=<UTC RFC 3339>` 줄이 있고, 결과가 `RAN`이 아니면 채택은 `not-adopted`다. 결정 줄은 리드가 운영자 답을 받아 적는 기록이다(운영자 결정 자체의 증거는 그 줄이 가리키는 대화다).

```bash
grep -Ecx 'REFUSED|RAN|NOT MEASURED|NOT DISCRIMINATING' .moai/reports/t1172/discriminator/outcome.txt | grep -qx 1 && grep -Ecx 'default|opt-in|not-adopted' .moai/reports/t1172/adoption.txt | grep -qx 1 && { { grep -qx 'RAN' .moai/reports/t1172/discriminator/outcome.txt && { { grep -qx 'default' .moai/reports/t1172/adoption.txt && grep -Eq '^OPERATOR-DECISION adoption=default card=t1172 recorded_by=lead at=[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$' .moai/reports/t1172/verdict.md; } || { grep -qx 'opt-in' .moai/reports/t1172/adoption.txt && grep -Eq '^OPERATOR-DECISION adoption=opt-in card=t1172 recorded_by=lead at=[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$' .moai/reports/t1172/verdict.md; }; }; } || { ! grep -qx 'RAN' .moai/reports/t1172/discriminator/outcome.txt && grep -qx 'not-adopted' .moai/reports/t1172/adoption.txt; }; } && echo true
```

음성·변이(plan 단계에서 확인): 결과 `RAN`인데 채택 `not-adopted` → `false`, 결정 줄에 `card=`·`recorded_by=`·`at=` 없음 → `false`. 결과 `REFUSED` + 채택 `not-adopted` → `true`.

### AC-CPP-007 — 방출이 채택 갈래를 따르고, 이 도구 하나로 한정되며, 도구는 쓰기 도구로 남는다 (maps REQ-CPP-006)

**Given** `adoption.txt`와 이 브랜치의 코드, **When** 방출 테스트(와 채택 갈래에서는 doctor 경고 테스트)를 돌리고 소스를 검사하면, **Then** 방출 테스트와 하위 테스트 `existing_table_invariant`·`tool_tables_exact`가 통과하고, 갈래별 하위 테스트가 통과하며(`default`: `…/default`; `opt-in`: `…/opt_in_off`와 `…/opt_in_on`), 채택 갈래에서는 `TestDoctorCodexPreApprovalWarning`이 통과하고, `not-adopted`에서는 제품 소스에 `tools.codex_role_audit`가 없다. 공통으로 서버 기본 승인 모드는 `writes`, `codex_role_audit` 밖의 도구별 표 문자열은 없고, `codex_role_audit` 도구 선언의 `mcp.WithReadOnlyHintAnnotation(false)`가 그대로다.

`tool_tables_exact`는 writer 출력 바이트에서 `[mcp_servers.moai.tools.*]` 헤더 집합을 직접 단언한다(`not-adopted`: 빈 집합, 채택 갈래: `{codex_role_audit}`). 소스 grep은 그 보조다.

실행:

```bash
go test -json ./internal/codexwiring -run '^TestEnsureMCPTableRoleAuditPreApproval$' -count=1 -v > .moai/reports/t1172/ac-cpp-007.jsonl; unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestDoctorCodexPreApprovalWarning$' -count=1 -v > .moai/reports/t1172/ac-cpp-007-doctor.jsonl
```

판정:

```bash
jq -se --rawfile ad .moai/reports/t1172/adoption.txt '($ad|rtrimstr("\n")) as $a | ([.[]|select(.Action=="pass" and .Test=="TestEnsureMCPTableRoleAuditPreApproval")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select(.Action=="pass" and .Test=="TestEnsureMCPTableRoleAuditPreApproval/existing_table_invariant")]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestEnsureMCPTableRoleAuditPreApproval/tool_tables_exact")]|length)==1 and (if $a=="default" then ([.[]|select(.Action=="pass" and .Test=="TestEnsureMCPTableRoleAuditPreApproval/default")]|length)==1 elif $a=="opt-in" then ([.[]|select(.Action=="pass" and (.Test=="TestEnsureMCPTableRoleAuditPreApproval/opt_in_off" or .Test=="TestEnsureMCPTableRoleAuditPreApproval/opt_in_on"))]|length)==2 elif $a=="not-adopted" then true else false end)' .moai/reports/t1172/ac-cpp-007.jsonl | grep -qx true && { grep -qx 'not-adopted' .moai/reports/t1172/adoption.txt || jq -se '([.[]|select(.Action=="pass" and .Test=="TestDoctorCodexPreApprovalWarning")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0' .moai/reports/t1172/ac-cpp-007-doctor.jsonl | grep -qx true; } && { ! grep -qx 'not-adopted' .moai/reports/t1172/adoption.txt || ! grep -rl 'tools\.codex_role_audit' internal/codexwiring internal/template/templates | grep -v '_test\.go$' | grep -q . ; } && grep -q 'mcpApprovalMode = "writes"' internal/codexwiring/configtoml.go && ! grep -rhoE 'mcp_servers\.moai\.tools\.[a-z_]+' internal/codexwiring --include='*.go' --exclude='*_test.go' | grep -vx 'mcp_servers\.moai\.tools\.codex_role_audit' | grep -q . && awk '/mcp\.NewTool\(codexRoleAuditToolName,/{f=1} f; f&&/handleCodexRoleAudit\}/{f=0}' internal/cli/codex_audit_mcp.go | grep -c 'mcp.WithReadOnlyHintAnnotation(false)' | grep -qx 1 && echo true
```

음성·변이(plan 단계에서 확인): `…/default` 하위 테스트 없음 → `false`, 도구 선언을 `WithReadOnlyHintAnnotation(true)`로 바꿈 → `false`, writer 소스에 `mcp_servers.moai.tools.spec_progress` 추가 → `false`, `mcpApprovalMode`를 `approve`로 바꿈 → `false`.

### AC-CPP-008 — 생성 설정의 키 이름이 Codex 로더를 통과하고, 양성 대조가 오타를 잡는다 (maps REQ-CPP-007)

**Given** codex가 설치된 환경, **When** 로더 검증 테스트를 돌리면, **Then** 부모 테스트와 하위 테스트 `generated_config`, `positive_control_misspelled_key`, `positive_control_bad_enum`이 pass이고 skip이 없으며, 채택 갈래에서는 `generated_config_has_table`(로더에 넣는 바이트에 도구별 표가 정확히 한 번 있음, A2는 켠 출력)도 pass다. `generated_config`는 제품 writer가 만든 바이트를 쓰며, 양성 대조는 `unknown configuration field` 문구와 정확한 점 경로 `mcp_servers.moai.tools.codex_role_audit.approval_modex`를 확인한다. 이 오타 표는 채택 갈래와 무관하게 테스트가 writer 출력 뒤에 덧붙여 만든다(N 갈래에서도 같은 대조를 쓴다).

실행:

```bash
go test -json ./internal/codexwiring -run '^TestGeneratedConfigKeysLoadInCodex$' -count=1 -v > .moai/reports/t1172/ac-cpp-008.jsonl
```

판정:

```bash
jq -se --rawfile ad .moai/reports/t1172/adoption.txt '($ad|rtrimstr("\n")) as $a | ([.[]|select(.Action=="pass" and .Test=="TestGeneratedConfigKeysLoadInCodex")]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestGeneratedConfigKeysLoadInCodex/generated_config")]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestGeneratedConfigKeysLoadInCodex/positive_control_misspelled_key")]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestGeneratedConfigKeysLoadInCodex/positive_control_bad_enum")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ($a=="not-adopted" or ([.[]|select(.Action=="pass" and .Test=="TestGeneratedConfigKeysLoadInCodex/generated_config_has_table")]|length)==1)' .moai/reports/t1172/ac-cpp-008.jsonl | grep -qx true && echo true
```

음성·변이(plan 단계에서 합성 스트림으로 확인): 양성 대조 하위 테스트 skip → `false`. run 단계에서 테스트가 FAIL해야 하는 변이: 오타 설정을 정상 키로 되돌림(대조 불발), 픽스처에서 `CODEX_HOME` 신뢰 항목 제거(프로젝트 층 미적재 → 대조 불발), writer 출력 대신 고정 문자열 사용(코드 검사).

### AC-CPP-009 — codex가 없으면 이유를 적고 skip하며, pass로 세지 않는다 (maps REQ-CPP-007)

**Given** codex 실행 파일을 찾을 수 없게 한 환경, **When** 로더 검증 테스트를 돌리면, **Then** 부모 테스트는 skip이고 pass가 아니며, 출력에 `CODEX_NOT_INSTALLED`가 있다.

실행:

```bash
MOAI_CODEX_BIN=/nonexistent/codex go test -json ./internal/codexwiring -run '^TestGeneratedConfigKeysLoadInCodex$' -count=1 -v > .moai/reports/t1172/ac-cpp-009.jsonl
```

판정:

```bash
jq -se '([.[]|select(.Action=="skip" and .Test=="TestGeneratedConfigKeysLoadInCodex")]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestGeneratedConfigKeysLoadInCodex")]|length)==0 and ([.[]|select((.Output//"")|test("CODEX_NOT_INSTALLED"))]|length)>=1' .moai/reports/t1172/ac-cpp-009.jsonl | grep -qx true && echo true
```

음성·변이(plan 단계에서 확인): 부모 테스트가 pass로 끝남(사유 없음) → `false`.

### AC-CPP-010 — 지시면이 거부 시 할 일을 적는다 (maps REQ-CPP-008, F4)

**Given** `AGENTS.md.tmpl`의 `audit-verdict-file` 행, **When** 행을 마침표로 나눠 거부 문구(`requires approval`)가 든 문장만 뽑고, 지시면 테스트와 중립성 검사를 돌리면, **Then** 그 문장에 `blocker`, `skip`, `spawn_agent`가 모두 있고(현재 행의 기존 `spawn_agent` 문장으로는 충족되지 않는다), `internal/template/agentemit` 테스트에 실패가 없으며, 이 브랜치가 템플릿에 더한 줄에 SPEC ID·카드 ID·날짜·로컬 절대 경로(`/Users/`, `/private/var/`)·9자 이상 16진 문자열(커밋 SHA 모양)이 없다. 그 밖의 중립성 항목은 CI 가드(`template-neutrality-check`)가 맡는다.

```bash
grep -c '^| audit-verdict-file |' internal/template/templates/AGENTS.md.tmpl | grep -qx 1 && grep '^| audit-verdict-file |' internal/template/templates/AGENTS.md.tmpl | tr '.' '\n' | grep 'requires approval' > .moai/reports/t1172/ac-cpp-010-sentence.txt && grep -q 'blocker' .moai/reports/t1172/ac-cpp-010-sentence.txt && grep -q 'skip' .moai/reports/t1172/ac-cpp-010-sentence.txt && grep -q 'spawn_agent' .moai/reports/t1172/ac-cpp-010-sentence.txt && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test -json ./internal/template/agentemit/... -count=1 | jq -se '([.[]|select(.Action=="fail")]|length)==0 and ([.[]|select(.Action=="pass" and .Test==null)]|length)>=1' | grep -qx true && git diff --name-only develop...HEAD -- internal/template/templates | grep -q . && ! git diff develop...HEAD -- internal/template/templates | grep -E '^\+[^+]' | grep -Eq 'SPEC-[A-Z][A-Z0-9-]*-[0-9]{3}|\bt[0-9]{3,4}\b|20[0-9]{2}-[0-9]{2}-[0-9]{2}|/Users/|/private/var/|\b[0-9a-f]{9,40}\b' && echo true
```

plan 단계 결과: 현재 트리에 적용 → `false`(행에 거부 문장이 없고 템플릿 변경도 없다 — run 전 RED). 문장 추출 부분을 합성 행(새 문장 한 개를 덧붙인 현재 행)에 적용 → 참. 현재 행의 `spawn_agent` 출현 수는 1이며 그 문장은 `requires approval`을 담지 않는다.

### AC-CPP-011 — 이월 AC의 본문·판정식은 선언한 경로 치환 하나만 다르다 (maps REQ-CPP-009)

**Given** 베이스 커밋 `0356e8117`의 선행 `acceptance.md` 232–256행(AC-CAR-010 본문)과 262–276행(AC-CAR-011 본문), **When** 두 구간에 `s#reports/t1143#reports/t1172#g` 하나만 적용해 §C의 표시(`<!-- carried:begin … -->` / `<!-- carried:end … -->`, 줄 전체 일치) 사이 줄과 비교하면, **Then** 두 블록 모두 바이트 같고, 치환 뒤 블록에 `reports/t1143`이 없으며 `reports/t1172`가 20번 나온다(치환이 실제로 일어났음을 보이는 양성 대조).

```bash
mkdir -p .moai/reports/t1172 && git show 0356e8117:.moai/specs/SPEC-CODEX-AUDIT-READONLY-001/acceptance.md | sed -n '232,256p' | sed 's#reports/t1143#reports/t1172#g' > .moai/reports/t1172/carry-010.src && git show 0356e8117:.moai/specs/SPEC-CODEX-AUDIT-READONLY-001/acceptance.md | sed -n '262,276p' | sed 's#reports/t1143#reports/t1172#g' > .moai/reports/t1172/carry-011.src && awk '/^<!-- carried:end AC-CAR-010 -->$/{f=0} f; /^<!-- carried:begin AC-CAR-010 -->$/{f=1}' .moai/specs/SPEC-CODEX-PREAPPROVAL-PROBE-001/acceptance.md > .moai/reports/t1172/carry-010.dst && awk '/^<!-- carried:end AC-CAR-011 -->$/{f=0} f; /^<!-- carried:begin AC-CAR-011 -->$/{f=1}' .moai/specs/SPEC-CODEX-PREAPPROVAL-PROBE-001/acceptance.md > .moai/reports/t1172/carry-011.dst && test -s .moai/reports/t1172/carry-010.src && test -s .moai/reports/t1172/carry-011.src && cmp -s .moai/reports/t1172/carry-010.src .moai/reports/t1172/carry-010.dst && cmp -s .moai/reports/t1172/carry-011.src .moai/reports/t1172/carry-011.dst && cat .moai/reports/t1172/carry-010.dst .moai/reports/t1172/carry-011.dst | grep -o 'reports/t1143' | grep -c . | grep -qx 0 && cat .moai/reports/t1172/carry-010.dst .moai/reports/t1172/carry-011.dst | grep -o 'reports/t1172' | grep -c . | grep -qx 20 && echo true
```

### AC-CPP-012 — 입력 누락·팔 diff 초과·시작 검사 실패면 LIVE가 시작되지 않는다 (maps REQ-CPP-001, REQ-CPP-002, REQ-CPP-003, REQ-CPP-004)

**Given** 가짜 codex(기동마다 인자를 파일에 적고, `--strict-config`와 접속 불가 제공자가 붙은 시작 검사 기동과 그 밖의 LIVE 기동을 구별해 센다)와 합성 픽스처, **When** 하네스를 세 음성 경우 — (a) 반출 입력 하나 삭제, (b) 처치 `argv.txt`에 한 줄 추가, (c) 시작 검사의 표준 오류에 `Not inside a trusted directory` — 로 돌리면, **Then** 세 경우 모두 가짜 codex의 LIVE 기동이 0회이고, 장부에 사유를 담은 `stop` 행이 있으며, `stop` 행 뒤에 `live` 행이 없다. 이 AC는 LIVE 호출을 쓰지 않는다. 실제 LIVE 실행의 장부에 대해서도 같은 순서 조건(`stop` 뒤 `live` 없음)을 AC-CPP-014가 본다.

실행:

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexPreApprovalGateRefusals$' -count=1 -v > .moai/reports/t1172/ac-cpp-012.jsonl
```

판정:

```bash
jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexPreApprovalGateRefusals")]|length)==1 and ([.[]|select(.Action=="pass" and (.Test=="TestCodexPreApprovalGateRefusals/missing_input" or .Test=="TestCodexPreApprovalGateRefusals/arm_diff_exceeds" or .Test=="TestCodexPreApprovalGateRefusals/startup_check_fails"))]|length)==3 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0' .moai/reports/t1172/ac-cpp-012.jsonl | grep -qx true && echo true
```

음성·변이(plan 단계에서 합성 스트림으로 확인): 하위 테스트 셋 중 하나만 pass → `false`.

### AC-CPP-013 — 이월 AC의 픽스처 조건이 지켜졌다 (maps REQ-CPP-009)

**Given** 반출된 car010·car011 입력, car010 부모 세션 출력, `adoption.txt`, **When** 이것들을 보면, **Then** 채택 갈래(`default`·`opt-in`)에서는 car010 사용자 층 decoy 표 안에 `enabled_tools = ["spec_progress"]`가 정확히 한 번 있고, car010 프로젝트 설정에 도구별 표가 정확히 한 번 있으며, 부모 세션의 `codex_role_audit` 호출이 1회 이상이고 그 서버가 모두 `moai`다(`not-adopted`에서는 AC-CAR-010을 돌리지 않으므로 이 부분을 보지 않는다). 모든 갈래에서 car011 입력 7종이 비어 있지 않고, `m1-route/route.txt`가 t1143 반출본 사본과 바이트 같으며 값이 `mcp`다.

```bash
{ grep -qx 'not-adopted' .moai/reports/t1172/adoption.txt || { awk '/^\[/{s=($0=="[mcp_servers.decoy]")} s' .moai/reports/t1172/car010/inputs/codex-home-config.toml | grep -cx 'enabled_tools = \["spec_progress"\]' | grep -qx 1 && grep -cx '\[mcp_servers\.moai\.tools\.codex_role_audit\]' .moai/reports/t1172/car010/inputs/project-config.toml | grep -qx 1 && jq -se '[.[]|select(.type=="item.completed" and .item.type=="mcp_tool_call" and .item.tool=="codex_role_audit")] as $c | ($c|map(select(.item.server=="moai"))|length)>=1 and ($c|map(select(.item.server!="moai"))|length)==0' .moai/reports/t1172/car010/parent.jsonl | grep -qx true; }; } && [ "$(for f in codex-home-config.toml project-config.toml repo.txt moai-build.txt argv.txt prompt.txt root-state.txt; do [ -s ".moai/reports/t1172/car011/inputs/$f" ] && echo ok; done | grep -c ok)" = 7 ] && cmp -s .moai/reports/t1172/m1-route/route.txt .moai/reports/t1172/plan-checks/t1143-route.txt && grep -qx 'mcp' .moai/reports/t1172/m1-route/route.txt && echo true
```

음성·변이(plan 단계에서 확인): decoy 표에 `enabled_tools` 없음 → `false`, 부모의 `codex_role_audit` 호출 서버가 `decoy` → `false`, `route.txt` 없음 → `false`. `not-adopted` 갈래 → car010 부분을 건너뛰고 `true`.

### AC-CPP-014 — 장부의 LIVE 호출이 선언한 상한 안에 있다 (maps REQ-CPP-004, REQ-CPP-010)

**Given** `ledger.json`, **When** 정수 시각으로 호출 시간과 창을 계산하면, **Then** LIVE 행이 1개 이상 14개 이하이고, 모든 LIVE 행의 시각이 정수이며 끝이 시작보다 늦거나 같고, 하네스가 띄운(`bound_by: test`) 호출은 각각 330 s 이하, 판별 LIVE는 4회 이하이며 첫 시작에서 마지막 끝까지 900 s 이하, car010은 6회 이하·1100 s 이하, car011은 4회 이하·800 s 이하이고, 시작 검사는 각각 25 s 이하이며, `stop` 행이 있으면 모든 LIVE 행은 첫 `stop`보다 먼저 시작했다. launcher가 띄운 자식(`bound_by: launcher`)의 개별 상한은 launcher 자신의 `config.DefaultCodexAuditTimeout`이며 이 AC는 창 합계로만 본다.

```bash
jq -e '. as $all | ($all|map(select(.kind=="live"))) as $live | (($all|map(select(.kind=="stop"))|map(.started_ns)|min)) as $stop | ($live|length)>=1 and ($live|length)<=14 and ($live|all((.started_ns|type)=="number" and (.ended_ns|type)=="number" and .ended_ns>=.started_ns)) and ($live|map(select(.bound_by=="test"))|all(.ended_ns-.started_ns <= 330000000000)) and ([["disc-",900,4],["car010",1100,6],["car011",800,4]]|all(. as [$p,$cap,$n] | ($live|map(select(.fixture|startswith($p)))) as $r | ($r|length)<=$n and (($r|length)==0 or (($r|map(.ended_ns)|max)-($r|map(.started_ns)|min) <= $cap*1000000000)))) and ($all|map(select(.kind=="startup"))|all((.ended_ns|type)=="number" and .ended_ns-.started_ns <= 25000000000)) and ($stop==null or ($live|all(.started_ns < $stop)))' .moai/reports/t1172/ledger.json >/dev/null && echo true
```

음성·변이(plan 단계에서 확인): 처치 LIVE 400 s → `false`, 첫 LIVE 앞에 `stop` 행 → `false`.

## §C 이월 AC (선행 SPEC-CODEX-AUDIT-READONLY-001)

이월 규칙:

- 본문과 판정식은 증거 경로 치환 `reports/t1143` → `reports/t1172` 하나만 다르다(AC-CPP-011이 검사). **판정식의 변경은 이 경로 치환뿐이다.** 이유: 이 카드의 증거 경로가 `.moai/reports/t1172/`이기 때문이다. 환경 변수 이름 `MOAI_T1143_EVIDENCE_DIR`는 기존 테스트 상수이므로 그대로다.
- 선행 파일의 이월 안내 인용문(`> **[카드 t1172로 이월 …`)은 옮기지 않았다. 선행 SPEC에서의 상태 표시이지 AC 본문이 아니기 때문이다.
- 이름 대응: 이월 본문의 "M1"은 선행 SPEC의 M1을 가리킨다. "M1에서 측정한 경로"와 `route.txt`는 이 SPEC의 `.moai/reports/t1172/m1-route/route.txt`(값 `mcp`, AC-CPP-013이 사본임을 확인)이고, "M1 P-B가 기록하는 MCP를 켠 부모 세션의 래퍼 기동 수"는 이 SPEC AC-CPP-004의 `car010` 시작 검사 기동 수(`startup/car010.launches.json`)다. 이 SPEC의 M1(판별 프로브)과 섞어 읽지 않는다.
- 판정식이 읽는 `.moai/reports/t1172/m1-route/route.txt`는 t1143 반출본(`mcp`, sha256 `97c5f37f209cef82840992f29653bda1fa3362aa8273b0df20773ef917a679ee`)을 run 단계에서 복사해 둔다.
- 픽스처 수정(판정식 아님)은 `plan.md` §D6이고, 그 조건은 AC-CPP-013이 확인한다.
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

### AC-CAR-011 — [LIVE] 나머지 read-only 역할의 쓰기 차단 (이월; REQ-CPP-009, REQ-CPP-010)

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

- 결정적 AC: AC-CPP-001, 002, 003, 004, 006, 007, 008, 009, 010, 011, 012, 013, 014. 모두 `true`여야 한다. AC-CPP-002–004·013·014는 LIVE 실행이 남긴 증거를 읽지만 판정 자체는 결정적이다.
- LIVE AC: AC-CPP-005, AC-CAR-010(채택 갈래에서만), AC-CAR-011. LIVE 결과는 기록된 분류 그대로 집계하며 PASS로 부풀리지 않는다.
- 재측정 범위는 변경 패키지로 한정한다: `./internal/codexwiring/...`, `./internal/cli`(해당 테스트 이름), `./internal/template/agentemit/...`, `./internal/config -run 'Budget'`. 로컬에서 `go test ./...`를 돌리지 않는다. `internal/cli`를 넓게 돌릴 때는 `-timeout`을 명시한다.
- 린트: 변경 패키지에 `golangci-lint run` 0 issues, `go vet` 통과.
- 템플릿을 바꿨으면 `make build`를 돌린다. 역할 `.md` 또는 `agents-codex.yaml`을 바꿨으면 `make agents-emit`을 돌리고 `make agents-emit-check`가 통과해야 한다.
- **push 금지(리드 HARD 조건 6).** 카드 브랜치가 원격에 없어야 한다. 아래 명령이 `true`를 내야 완료로 본다(같은 파이프라인을 원격에 있는 `develop`에 걸면 `true`가 나오지 않음을 plan 단계에서 확인).

```bash
git ls-remote --heads origin WT-codex-preapproval-probe | grep -c . | grep -qx 0 && echo true
```

**완료 정의.** 판별 프로브가 유효한 측정을 한 번 냈고(AC-CPP-005) 그 결과에 따라 채택 갈래가 정해졌으며(AC-CPP-006), 방출·로더 검증·지시면·이월 AC 충실도·게이트 거부·이월 픽스처 조건·상한이 모두 `true`이고, AC-CAR-011이 실행되었으며, AC-CAR-010은 채택 갈래에서 실행되었거나 `not-adopted`로 `NOT_RUN` 사유가 적혀 있고, 위 push 금지 검사가 `true`일 때. 판별 프로브가 두 번째로 `NOT MEASURED`로 끝나면 이 카드는 "pre-approval effect not measured"로 멈추고, N 갈래의 AC(지시면, 로더 검증, AC-CAR-011)만 닫은 채 리드에게 보고한다.
