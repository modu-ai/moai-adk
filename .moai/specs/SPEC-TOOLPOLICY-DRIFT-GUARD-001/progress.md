# progress.md — SPEC-TOOLPOLICY-DRIFT-GUARD-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-10
artifacts: spec.md + plan.md + acceptance.md + progress.md (Tier M) + spec-compact.md — v0.1.3
card: t619
baseline: worktree `.claude/worktrees/t619`, branch `WT-toolpolicy-drift`, plan 작성 HEAD `c7b8d110b`, 1회차 감사 HEAD `b2cbfd207`, 2회차 `dc220b8fe`, 3회차 `7c96ed2e1`, `git merge-base HEAD develop` → `d1b61005d20967fdbd970ec7ec734c6d14f29dc3`
evidence_base: `.moai/reports/t619/verdict.md`
spec_id_check: `[[ "SPEC-TOOLPOLICY-DRIFT-GUARD-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS` (실행 출력), `ls .moai/specs | grep -c TOOLPOLICY-DRIFT-GUARD` → `0` (작성 전)
open_clarifications: 0 — 운영자 결정 6건(2026-09-10) 반영. 수리 방향·검사 위치·주장 정정(레인 세션 직접 수령), 집합 비교·주장 정정 전수·워킹 트리 판독(plan 제안 검토 후)
plan_measurements: 스크래치 build `allow=108 ask=0 deny=60 env_gated_skipped=5`, `diff` 종료 1 / 4 헝크 / 160 vs 166 줄, 커밋본 목록 `allow-NOT-C-sorted` `deny-NOT-C-sorted`, 커밋본 중복 0 / allow·deny 겹침 0
plan_audit: 1회차 FAIL 0.71 → 2회차 FAIL 0.79(Tier M 상한 도달, 운영자가 3회차 1회 승인) → 3회차 FAIL 0.89(blocking D21 1건) → 운영자 결정 "지금 고치고 진행", v0.1.3 수정
gap: **v0.1.3 의 D21-D23 수정은 독립 재감사를 받지 않았다.** 3회차가 최종 감사였고, 운영자 결정에 따라 수정 줄은 오케스트레이터가 직접 확인한다. 이 수정분에 대한 plan-auditor 판정은 존재하지 않는다.

### plan-audit 1회차 (2026-09-10) — FAIL 0.71 대응

- 보고서: `.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-1.md`
- blocking D1·D2·D3 반영, optional D4-D10 은 반영 또는 부분 반영(D8). 결함별 처분 표는 plan.md §H.
- 수리 회차 실측(HEAD `b2cbfd207`): env_gate 없음 + 네 경로 deny 질의 `12`, env-gated Write deny `1`, env_gate 항목 `5`, `ANALOGOUS drift class` `1`, `comment (audit surface) are generated` `1`, `tool-policy-drift-check` 두 파일 모두 `0`, `.claude/settings.json#permissions` `171`, `make -n build` 의 `TestToolPolicyDrift_` `0` / `TestGoldenCommittedArtifactsMatchEmission` `2`, `go_code:` 블록 14줄·`.moai/**` `1`·`.claude/settings.json` `0`, merge-base 기준 settings 두 파일 diff 출력 없음.
- D8 관련 가드 실측: `trap` 포함 명령, heredoc 으로 python/bash 에 스크립트를 넘기는 명령, 변수로 계산된 스크립트 경로를 python 에 넘기는 명령이 모두 워크트리 세션 가드에 거부됨. 리터럴 절대 경로의 스크래치 스크립트 호출은 통과.

### plan-audit 2회차 (2026-09-10) — FAIL 0.79 대응, 운영자 승인 3회차

- 보고서: `.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-2.md`
- Tier M 상한(2회) 도달. 운영자가 수정 후 3회차 감사 1회를 승인했다.
- blocking D11·D13·D14, optional D12·D15-D20 모두 반영. 처분 표는 plan.md §H 2회차.
- 수리 회차 실측(대상 파일 무수정 상태): env_gate 항목 정렬 JSON sha256 `5e0cba521c5c81a2d7bb82fbba52c59b027d2329b2c0bf8b912ed6b2150e6d27`, `^build:.*tool-policy-drift-check` `0` / `^build:.*agents-emit-check` `1`, 기존 네 파일 역슬래시-u / Cf 모두 `(0, 0)`, 조건부 복원 사슬 일치 시 `restored=yes`·불일치 시 `restored=NO_stop`, `git status --porcelain --untracked-files=all > <파일>` 가드 통과.

### plan-audit 3회차 (2026-09-10) — FAIL 0.89, blocking D21 1건, 운영자 결정 "지금 고치고 진행"

- 보고서: `.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-3.md`
- D21(blocking)·D22·D23(optional) 모두 반영. 처분 표는 plan.md §H 3회차.
- 수리 회차 실측(대상 파일 무수정 상태, 스크래치):
  - 고정 fixture 네 쌍을 `moai tool-policy build --local-only` 로 재생성: 생성 고유 집합이 fixture 고유 집합과 모두 같음. `duplicate_settings_allow`·`_deny` `{"allow":["Read"],"ask":[],"deny":["Bash(rm -rf /:*)"]}`, `_ask` ask `["WebFetch"]`, `allow_deny_overlap` deny `["Bash(rm -rf /:*)","Read"]`. 중복 fixture 길이·고유 길이 `[2,1]`, 겹침 fixture allow∩deny `["Read"]`. 세 YAML fixture 로딩 성공.
  - `malformed_settings_json` 바이트 → `permissions object parse: invalid character '}' looking for beginning of object key string`, 종료 `1`.
  - `wrong_type_settings_list` 바이트(`"ask": 1`) → 생성기 build 종료 `0`, ask 사라짐(`settings_region.go:173` 오류 폐기).
  - `malformed_yaml` 모양 → `tool-policy parse … yaml: line 5: mapping values are not allowed in this context`, 종료 `1`.
  - 조건부 변이 사슬(백업에서 계산한 원본 sha, 리터럴 경로): 원본 상태에서 `mutate=DONE`·항목 173→172, 한 줄 덧붙인 불일치 상태에서 `mutate=STOPPED`·항목 173 유지. 가드 통과.
  - 가드 추가 실측: 셸 반복문 안에서 변수 인자로 `moai` 를 부르는 명령, 여러 heredoc 을 묶은 명령은 거부됨. 파일마다 `printf '%s\n' … > <리터럴 경로>` 로 나누면 통과.

## Run Phase 1 — Plan Audit Gate

- audit_verdict: BYPASSED
- audit_report: .moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-3.md (최종 감사, FAIL 0.89)
- audit_at: 2026-09-10
- bypass_user: GOOS 오라버니~ (`.moai/config/sections/user.yaml`)
- bypass_reason: Tier M 감사 상한(2회)을 운영자 승인으로 1회 연장해 3회차까지 돌렸다. 3회차에 남은 blocking D21 을 두고 운영자가 "지금 고치고 진행"을 골랐고, 구현 시작 승인에서 "D21~D23 수정분이 독립 재감사 없이 진행되는 것"을 전제로 run 진입을 승인했다. 따라서 4회차 감사는 돌리지 않는다. 수정분의 오케스트레이터 직접 확인은 `.moai/reports/t619/verdict.md` §8.
- run_trigger: bypassed (operator decision, lane session AskUserQuestion)

## Phase 4 Mode Selection

- Input: tier M · 수정 파일 5개(`tool-policy.yaml`, `internal/config/toolpolicy/drift_check_test.go` 신규, `Makefile`, `.github/workflows/ci.yml`, `internal/config/toolpolicy/types.go`) · 영역 1개(Go 설정 패키지와 빌드 배선) · 언어 Go + YAML + Makefile · 병렬 이득 없음(M1→M5 가 서로 의존)
- 평가: serial — 적합(단일 작성자, 순차 의존). fanout — 부적합(조사 단계가 이미 끝남). agent-team — 부적합(파일 5개, 조정 비용이 이득보다 큼). sweep — 부적합(균일한 기계적 변환이 아니고 30파일 미만)
- Decision: Scale-based mode: serial / Focused (files: 5, domains: 1)
- Justification: 드리프트 검사가 먼저 붉게 재현돼야 YAML 수정의 초록이 의미를 가지므로 M1→M2→M3 이 순서 의존이다. 쓰기 가능한 에이전트는 manager-develop 하나만 둔다.
- progression: autonomous (운영자 선택, 2026-09-10). `/moai goal` 은 워크트리 세션 키 불일치 위험(기존 교훈) 때문에 무장하지 않고, 오케스트레이터가 단계마다 증거를 읽어 진행한다.

## §E.2 Run-phase Evidence

### 기준선 귀속

- 워크트리 `.claude/worktrees/t619`, 브랜치 `WT-toolpolicy-drift`. 착수 HEAD `b75d4fdcd`, 마지막 코드 커밋 `e2fc880b8`, `git merge-base HEAD develop` → `d1b61005d20967fdbd970ec7ec734c6d14f29dc3`.
- 아래 수치는 모두 이번 run 에서 이 트리를 대상으로 실행한 명령의 출력이다. 로그는 세션 스크래치(`t619/`)에 두었고 휘발성이므로, 판정에 쓰인 출력은 이 절에 그대로 옮긴다.
- develop 흡수는 하지 않았다(지시). push·PR 없음.

### Pre-flight (코드 변경 전, HEAD `b75d4fdcd`)

- `go build ./...` → `build_rc=0`, `GOOS=windows GOARCH=amd64 go build ./...` → `winbuild_rc=0`
- `golangci-lint run --timeout=2m ./internal/config/toolpolicy/...` → `0 issues.`
- `go test -count=1 ./internal/config/toolpolicy/...` → `ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.447s`
- acceptance.md 의 "현재" 기준선 재측정 — 전부 일치, 차이 없음:
  - AC-TDG-007 여덟 grep: `1`(두 파일) / `1` / `1` / `1` / `1` / `0`(두 파일) / `171` / `1`
  - AC-TDG-008: `0` / `1` / `0` / `2` / 블록 `14` 줄 / `0` / `1`
  - AC-TDG-010: `rc=0`, `MultiEdit` `1`, env_gate 없는 네 경로 deny `12`, env-gated Write deny `1`, env_gate `5`, 정렬 JSON sha256 `5e0cba521c5c81a2d7bb82fbba52c59b027d2329b2c0bf8b912ed6b2150e6d27`, 일곱 도구 allow `0`, Read deny `5`, 항목 수 `173`
  - AC-TDG-003: merge-base 기준 `git diff --exit-code` 출력 없음, `rc=0`

### M1 RED 재현 (E8, M2 전 — 커밋 `dfcf7c519` 직전 트리)

```
$ go test ./internal/config/toolpolicy/ -run 'TestToolPolicyDrift_CommittedSettingsMatchYAML$' -count=1 -v
=== RUN   TestToolPolicyDrift_CommittedSettingsMatchYAML
    drift_check_test.go:255: tool-policy.yaml and .claude/settings.json permissions declare different sets (20 specifiers):
        allow only-in-yaml: MultiEdit
        allow only-in-settings: CronCreate
        allow only-in-settings: CronDelete
        allow only-in-settings: CronList
        allow only-in-settings: EnterPlanMode
        allow only-in-settings: EnterWorktree
        allow only-in-settings: ExitPlanMode
        allow only-in-settings: ExitWorktree
        deny only-in-yaml: Glob(./secrets/**)
        deny only-in-yaml: Glob(~/.aws/**)
        deny only-in-yaml: Glob(~/.config/gcloud/**)
        deny only-in-yaml: Glob(~/.ssh/**)
        deny only-in-yaml: Grep(./secrets/**)
        deny only-in-yaml: Grep(~/.aws/**)
        deny only-in-yaml: Grep(~/.config/gcloud/**)
        deny only-in-yaml: Grep(~/.ssh/**)
        deny only-in-yaml: Write(./secrets/**)
        deny only-in-yaml: Write(~/.aws/**)
        deny only-in-yaml: Write(~/.config/gcloud/**)
        deny only-in-yaml: Write(~/.ssh/**)
        reconcile tool-policy.yaml to the intended state, then regenerate with `moai tool-policy build --local-only`
--- FAIL: TestToolPolicyDrift_CommittedSettingsMatchYAML (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.370s
FAIL
```

집계(AC-TDG-001 붉은색 명령 그대로): `rc=1`, 서로 다른 차이 줄 `20`, 분할 `7` / `1` / `12` / `0`, `--- FAIL:` `1`. 판정서 §2.2 목록과 항목이 같다.
M2 전 env_gate 해시 재측정: `5e0cba521c5c81a2d7bb82fbba52c59b027d2329b2c0bf8b912ed6b2150e6d27` (기준선과 같음).

### AC 매트릭스

| AC | 상태 | 명령 | 관측 출력 |
|---|---|---|---|
| AC-TDG-001 | PASS | 붉은색: 위 E8. 초록(M2 뒤): `go test … -run 'TestToolPolicyDrift_(CommittedSettingsMatchYAML\|NoDuplicatesOrOverlap)$' -count=1 -v` 와 `make tool-policy-drift-check` | 붉은색 `rc=1`/`20`/`7 1 12 0`/`1`. 초록 `rc=0`, `--- PASS:` `1`·`1`, `only-in-` `0`. make: `ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.189s` 후 `rc=0` |
| AC-TDG-002 | PASS | `moai tool-policy build --repo-root <스크래치>/root --policy .moai/config/sections/tool-policy.yaml --local-only` 후 정렬 목록 `diff` | `<스크래치>/root/.claude/settings.json [json]: allow=114 ask=0 deny=48 env_gated_skipped=5`, `allow diff_rc=0`, `deny diff_rc=0`, ask 길이 `0` `0` |
| AC-TDG-003 | PASS | HEAD `e2fc880b8` 에서 merge-base `git diff --exit-code`, 검사 전후 `shasum -a 256` 과 `git status --porcelain --untracked-files=all` 비교 | `rc=0`(출력 없음), `check_rc=0`, `cmp_rc=0`, `status_cmp_rc=0`. 입력 sha: settings `25e19e906f639044c38371fadec796f080cab2630b0d295478a62bdc47e73c39`, YAML `edff7e725e02738f78a77d8d9959a2f933b2e6fe705ce8d044e251048c544cdf` |
| AC-TDG-004 | PASS | acceptance.md 조건부 사슬 그대로(M5 커밋 `e2fc880b8` 뒤) | 아래 절 참조. `mutate=DONE` → 붉은색 → `restore=DONE` → 초록, `diff_rc=0` |
| AC-TDG-005 | PASS | `go test … -run 'TestToolPolicyDrift_Mutation$' -count=1 -v` | `rc=0`, `--- PASS: …/` `7`, `--- FAIL:` `0`, `errDriftDuplicate` `7`, `errDriftOverlap` `7` |
| AC-TDG-006 | PASS | `go test … -run 'TestToolPolicyDrift_FailClosed$' -count=1 -v` | `rc=0`, `--- PASS: …/` `10`, `--- SKIP:` `0`, `t.Skip` `0`, `errors.Is(` `4`, `TestToolPolicyDrift` `8` |
| AC-TDG-007 | PASS | 여덟 grep (M5 뒤) | `0`(두 파일) / `0` / `0` / `0` / `0` / YAML `1`·types.go `2` / `165` / `1`. 7번 `171`→`165` 는 M2 가 `source: ".claude/settings.json#permissions.*"` 를 가진 항목 13개를 빼고 7개를 더한 결과(−6)이며 기대값 "1 이상"을 만족한다 |
| AC-TDG-008 | PASS | `^build:` grep, `make -n build`, `go_code` 블록 추출 | `1` / `1` / `1` / `2` / `15` / `1` / `1` |
| AC-TDG-009 | **FAIL (선재 결함, 이 변경 귀속 아님)** | 패키지·cli·config 세 실행 | 아래 절 참조. toolpolicy `rc=0`·RoundTrip PASS `1`·FAIL `0`, config `rc=0`·PASS `1`, **cli `rc=1`** |
| AC-TDG-010 | PASS | `moai tool-policy list … --format json` 질의 | `rc=0`, `0`, `0`, `1`, `5`, `5e0cba521c5c81a2d7bb82fbba52c59b027d2329b2c0bf8b912ed6b2150e6d27`, 일곱 도구 각 `1`, Read deny `5`, 항목 수 `167`(= 173 − 13 + 7) |

### AC-TDG-004 수동 뮤테이션 기록

- 백업 경로: `<세션 스크래치>/t619/tool-policy.yaml.bak`
- 원본 sha256(백업에서 계산): `edff7e725e02738f78a77d8d9959a2f933b2e6fe705ce8d044e251048c544cdf`
- 조건부 변이 사슬 → `mutate=DONE`
- 변이 sha256: `5837f0bcee9d982b0c9bb4aa1e8094439ddceecaf044217ff99d16bc4703746e`, `mutation_applied_cmp_rc=1`
- 변이 뒤 `make tool-policy-drift-check` → `rc=2`, `allow only-in-settings: ExitWorktree` `1`, 서로 다른 차이 줄 `1`, `reconcile tool-policy.yaml` `2`, `moai tool-policy build --local-only` `2`. 출력:
  ```
  --- FAIL: TestToolPolicyDrift_CommittedSettingsMatchYAML (0.00s)
      drift_check_test.go:257: tool-policy.yaml and .claude/settings.json permissions declare different sets (1 specifiers):
          allow only-in-settings: ExitWorktree
          reconcile tool-policy.yaml to the intended state, then regenerate with `moai tool-policy build --local-only`
  FAIL
  FAIL	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.340s
  FAIL
  tool-policy drift: .claude/settings.json permissions differ from .moai/config/sections/tool-policy.yaml — reconcile tool-policy.yaml, then regenerate with `moai tool-policy build --local-only`
  make: *** [tool-policy-drift-check] Error 1
  ```
- 조건부 복원 사슬 → `restore=DONE`
- 복원 뒤: `sha_rc=0`, make `rc=0`, `only-in-` `0`, 안내 문구 `0`·`0`, `git diff --exit-code -- .moai/config/sections/tool-policy.yaml` → `diff_rc=0`
- `STOPPED` 는 한 번도 나오지 않았다. `git restore`/`git checkout --` 는 쓰지 않았다.

### AC-TDG-009 실패 — 원인 귀속

관측(HEAD `e2fc880b8`, 뮤턴트가 `internal/config/toolpolicy` 테스트 파일에 있던 시점이나 `internal/cli` 는 그 파일을 컴파일하지 않는다):

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -timeout 600s ./internal/cli/ -run 'TestToolPolicy' -v
rc=1
    tool_policy_test.go:97: output missing substring "ask";
--- FAIL: TestToolPolicyList_QueryFilters (0.00s)
    --- FAIL: TestToolPolicyList_QueryFilters/filter_ask (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_deny (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/all_entries (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_tool_Bash (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_irreversible (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_allow (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.001s
```

`filter_ask` 는 커밋된 YAML 에 `decision: ask` 항목이 하나 이상 있어야 통과한다. 이 run 이전부터 그런 항목이 없다는 증거:

- 착수 트리의 YAML(`git show b75d4fdcd:.moai/config/sections/tool-policy.yaml`)에서 `decision: ask` 줄 `0`, 현재 YAML 도 `0`. M2 는 allow·deny 항목만 바꿨다.
- 현재 트리의 list 코드로 착수 트리 YAML 을 조회: `go run ./cmd/moai tool-policy list --policy <착수 YAML 사본> --decision ask` → 머리글 `TOOL  ARGS  RISK  DECISION  OWNER  AUDIT` 한 줄만, `base_rc=0`. 소문자 `ask` 가 없으므로 착수 트리에서도 같은 단정이 실패한다.
- ask 항목을 없앤 커밋: `1ac333952 fix(t609): drop the six permissions.ask rules from template, local settings, and the policy SSOT`. `git merge-base --is-ancestor 1ac333952 d1b61005d…` → `rc=0`(이 카드의 base 에 이미 포함). `decision: ask` 줄 수: `1ac333952^` 에서 `6`, `1ac333952` 에서 `0`. 그 커밋의 변경 파일은 `.claude/settings.json`, `tool-policy.yaml`, `.moai/reports/t609/verdict.md`, `settings.json.tmpl` 이며 `internal/cli/tool_policy_test.go` 는 없다.

갭: `internal/cli` 테스트 자체를 착수 트리에서 실행하지는 않았다(워크트리 추가 금지). 귀속은 같은 list 코드에 착수 트리 YAML 을 넣은 대조 실행과 git 객체 판독에 기댄다. 수리는 `internal/cli/tool_policy_test.go` 수정이 필요해 이 SPEC 의 파일 범위 밖이므로 하지 않았고, 차단 보고로 올린다.

### DoD 뮤턴트 6종

각 뮤턴트는 테스트 파일에 임시로 넣고, 지정 명령을 로그 경로만 바꿔 실행한 뒤 되돌렸다. 되돌린 직후마다 `git diff --exit-code -- internal/config/toolpolicy/drift_check_test.go` → `rc=0`(커밋본과 같음)을 확인했다.

| 뮤턴트 | 변이 | 관측 | 기대 |
|---|---|---|---|
| M-always-empty-diff | `driftSetDiff` 가 `return nil, nil` | `rc=1`, `--- FAIL: …/settings_side_missing` `1` (`removing "AskUserQuestion" from settings allow: diff = [], want exactly ["allow only-in-yaml: AskUserQuestion"]`). `yaml_side_flip` 도 붉음, 나머지 다섯 초록 | `1` |
| M-no-duplicate-check | 중복 판정 `if false && n > 1` | `rc=1`, `duplicate_settings_` FAIL `3`, `allow_deny_overlap` FAIL `0` | `3` / `0` |
| M-no-overlap-check | 겹침 판정 `if false && allow[spec]` | `rc=1`, `allow_deny_overlap` FAIL `1`, `duplicate_settings_` FAIL `0` | `1` / `0` |
| M-parse-yaml | `Load` 오류 무시, 빈 `PolicyDocument` 로 계속 | `rc=1`, `malformed_yaml` FAIL `1` (`driftSetDiff error yaml allow set is empty: … does not wrap its cause … input file cannot be parsed`) | `1` |
| M-parse-json | `extractPermissions` 오류 무시, 빈 `PermissionsBlock` 으로 계속 | `rc=1`, `malformed_settings_json` FAIL `1` (`settings allow set is empty … does not wrap its cause … input file cannot be parsed`) | `1` |
| M-list-type | 엄격 디코드 제거, `block.Allow/Ask/Deny` 사용 | `rc=1`, `wrong_type_settings_list` FAIL `1` (`driftSetDiff returned no error (diff []); want a failure wrapping … input file cannot be parsed`) | `1` |

되돌린 트리에서 다시: Mutation `rc=0`·PASS `7`, FailClosed `rc=0`·PASS `10`, 두 로그 FAIL `0`.

### 품질 게이트와 경계

- `go vet ./internal/config/toolpolicy/` → `vet_rc=0`
- `golangci-lint run --timeout=2m ./internal/config/toolpolicy/...` → `0 issues.` (기준선 `0 issues.`, 신규 0)
- `go test -count=1 -cover ./internal/config/toolpolicy/...` → `coverage: 89.1% of statements`
- `gofmt -l internal/config/toolpolicy/` → 출력 없음
- `grep -rn 'AskUserQuestion' internal/config/toolpolicy | grep -v _test.go | grep -v '// '` → 출력 없음
- 크로스 빌드(HEAD `e2fc880b8`): `build_rc=0`, `winbuild_rc=0`
- 역슬래시-u / Cf 스캔(acceptance.md §D.3 명령, 대조 `CONTROL (1, 1)`): `tool-policy.yaml`, `Makefile`, `ci.yml`, `types.go`, `drift_check_test.go`, `spec.md` 모두 `(0, 0)`

### 커밋

| SHA | 제목 | `Authored-By-Agent` |
|---|---|---|
| `dfcf7c519` | test(t619): add tool-policy drift check against settings.json permissions | `manager-develop` |
| `6ca23b0bd` | fix(t619): reconcile tool-policy.yaml with settings.json permissions | `manager-develop` |
| `3229f52dd` | test(t619): add mutation and fail-closed controls for the drift check | `manager-develop` |
| `508c6c5e8` | build(t619): wire tool-policy drift check into make build and CI filter | `manager-develop` |
| `e2fc880b8` | docs(t619): correct drift-prevention claims in tool-policy.yaml and types.go | `manager-develop` |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-10
run_commit_sha: e2fc880b8   # 마지막 코드 커밋. 이 progress.md 기록 커밋은 그 뒤에 온다
run_status: complete-with-blocker
ac_pass_count: 9
ac_fail_count: 1
ac_fail_detail: "AC-TDG-009 — internal/cli TestToolPolicyList_QueryFilters/filter_ask, t609(1ac333952) 이후 선재 실패. 이 SPEC 파일 범위 밖"
dod_mutants_red_observed: 6/6
preserve_list_post_run_count: 2   # .claude/settings.json, settings.json.tmpl — merge-base 대비 바이트 불변
l44_pre_commit_fetch: not-run     # 레인 지시: 커밋만, develop 흡수는 통합 창에서
l44_post_push_fetch: n/a          # push 없음
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: 0
  windows_amd64: 0
coverage_toolpolicy: "89.1%"
total_run_phase_files: 7   # tool-policy.yaml, drift_check_test.go, Makefile, ci.yml, types.go, spec.md, progress.md
m1_to_mN_commit_strategy: "마일스톤마다 커밋 1개(M1, M2, M3 자동 대조, M4, M5) + progress 기록 커밋"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-11
sync_commit_sha: pending-backfill-sync-commit   # backfilled in the follow-up commit (spec-frontmatter-schema.md D3)
sync_status: complete-with-attributed-preexisting-failure
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed (merged into this sync commit)"
  plan_md: "n/a — stateless on the status axis (spec-frontmatter-schema.md Artifact Statelessness), no frontmatter status field"
  acceptance_md: "n/a — stateless on the status axis, no frontmatter status field"
  spec_compact_md: "n/a — stateless on the status axis, no frontmatter status field"
  progress_md: "n/a — records phase progress in body sections, not frontmatter"
changelog_entry_position: "CHANGELOG.md line 12, first bullet under '## [Unreleased]' / '### Fixed'"
b12_self_test_a: "grep -c 'SPEC-TOOLPOLICY-DRIFT-GUARD-001' CHANGELOG.md -> 1 (post-insertion; pre-insertion was 0) — no duplicate entry"
b12_self_test_b: "grep -oE 'AC-TDG-[0-9]{3}' .moai/specs/SPEC-TOOLPOLICY-DRIFT-GUARD-001/acceptance.md | sort -u | wc -l -> 10, matches CHANGELOG '10 acceptance criteria' claim"
b12_self_test_c: "ls -la internal/config/toolpolicy/drift_check_test.go .moai/config/sections/tool-policy.yaml .claude/settings.json Makefile .github/workflows/ci.yml internal/config/toolpolicy/types.go -> all six exist"
pre_sync_gate:
  head: 1ff407dfe
  go_version: go1.26.4 (local toolchain; gap — see gaps below)
  go_vet: "go vet ./internal/config/toolpolicy/... -> exit 0"
  gofmt: "gofmt -l internal/config/toolpolicy/ -> empty output"
  golangci_lint: "golangci-lint run --timeout=3m ./internal/config/toolpolicy/... -> 0 issues."
  go_test_cover: "go test -count=1 -cover ./internal/config/toolpolicy/... -> ok, coverage: 89.1% of statements"
  drift_check: "make tool-policy-drift-check -> exit 0"
sync_audit:
  verdict: PASS
  overall_score: 90.7
  dimensions:
    functionality: 88
    security: 95
    craft: 90
    consistency: 92
  report: .moai/reports/t619/sync-audit.md
  ac_breakdown: "9 PASS / 1 FAIL (AC-TDG-009 attributed out-of-scope pre-existing failure; AC-TDG-004 PASS relied on the run-phase recorded mutation chain, not re-executed by the auditor)"
security_review:
  phase: "Phase 8 Step 0.55.1 — per-spawn read-only reviewer"
  exported_to_file: false
  note: "orchestrator-relayed summary, not an independently-citable verdict basis"
  critical: 0
  high: 0
  medium: 0
  low: 2
  low_1: "duplicate 'permissions' JSON key first-vs-last parse ambiguity in pre-existing settings_region.go:92-120 (out of this SPEC's scope)"
  low_2: "YAML reconciliation removed declared-but-unenforced Write/Grep/Glob secret-path denies; enforcement itself is unchanged, but whether Read/Edit denies already cover the same paths is unverified"
  dedup_predicate: "FALSE — run-exit deep scan scope was branch-level not repo-level, scanned_commit 1977040b7 != HEAD 1ff407dfe; only verdict.md changed since scan, so Step 0.55.1 ran fresh"
mx_tag_validation:
  tags_added: 0
  p1_p2_findings: none
  note: "types.go change is comment-only; drift_check_test.go is a test file with unexported helpers, no goroutines — no new @MX obligation"
ac_tdg_009_disposition:
  status: FAIL
  attribution: pre-existing
  cause_commit: "1ac333952 (t609) removed the six permissions.ask rules from template, local settings, and the policy SSOT"
  scope_check: "internal/cli/ unchanged between base d1b61005d and HEAD 1ff407dfe"
  repair_card: t660
  gate: "must land before the develop push per operator decision D1"
follow_up_candidates_forwarded_to_lead:
  - "secret-path Write/Grep/Glob deny gap in .claude/settings.json (security_review low_2)"
  - "Makefile:68 failure message always reads 'permissions differ' regardless of which check failed"
  - "duplicate-permissions-key parse robustness in settings_region.go (previously forwarded, still open)"
delivery:
  route: "git-flow (per CLAUDE.local.md §4.1)"
  branch: WT-toolpolicy-drift
  push_state: "not pushed — awaiting lead integration window (local develop merge)"
  integration_window_remeasure: "pending, to be run on go1.26.8 per lane discipline"
gaps:
  - "local toolchain measured at go1.26.4, not the go1.26.8 lane standard; re-measure at integration window"
  - "full test suite not run locally per CLAUDE.local.md §6 — CI owns the full-suite verdict"
  - "README / docs-site / .moai/project docs intentionally untouched per operator decision D2 (internal developer check, no user-facing change)"
```
