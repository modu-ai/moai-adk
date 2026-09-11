# t619 판정서 — tool-policy.yaml 과 settings.json 권한 블록 드리프트

- card: t619
- 브랜치: WT-toolpolicy-drift (base: 로컬 develop d1b61005d)
- 측정 트리: .claude/worktrees/t619 @ d1b61005d
- 작성: 2026-09-10, plan 단계 판정 (권한 표면 파일은 손대지 않았다)

## 1. 주장

1. 서로 다른 드리프트 항목은 **20개**다. 카드의 "18"은 순증감(allow -6, deny +12)을 더한 값이고, allow 쪽은 실제로 7개 소실 + 1개 신규다.
2. **의도된 상태는 커밋된 settings.json 이다. YAML 이 뒤처져 있다.** 20개 항목 모두에서 settings.json 과 배포 템플릿(settings.json.tmpl)이 같은 쪽에 서 있고, YAML 만 반대편이다.
3. 이 드리프트는 이미 한 번 사고로 터졌다. 519b848fb(2026-08-05)의 settings.json diff 가 지금 build 가 만들 변화와 **같은 20개 항목**을 담고 있고, 그 뒤 8df71b18d(2026-08-05)와 0de8517e5(2026-08-08)가 손으로 되돌렸다. YAML 은 그때 고쳐지지 않았다.
4. 따라서 수리 방향은 "YAML 을 settings.json 에 맞춘다"이고, 그렇게 하면 **실제 적용되는 권한(settings.json)은 바뀌지 않는다.**

## 2. 증거

### 2.1 생성기 재현 (스크래치 루트, 실제 트리 미변경)

```
moai tool-policy build --repo-root <scratchpad>/t619/root --policy .claude/worktrees/t619/.moai/config/sections/tool-policy.yaml --local-only
  <scratchpad>/t619/root/.claude/settings.json [json]: allow=108 ask=0 deny=60 env_gated_skipped=5
tool-policy build: regenerated 1 target(s)
```

커밋본 `jq '.permissions' .claude/settings.json` 집계: allow 114, ask 키 없음, deny 48.

### 2.2 차이 목록 (`comm -3 committed generated` — 들여쓰기 없음=커밋본에만, 탭=생성본에만)

allow:
```
CronCreate
CronDelete
CronList
EnterPlanMode
EnterWorktree
ExitPlanMode
ExitWorktree
	MultiEdit
```

deny:
```
	Glob(./secrets/**)
	Glob(~/.aws/**)
	Glob(~/.config/gcloud/**)
	Glob(~/.ssh/**)
	Grep(./secrets/**)
	Grep(~/.aws/**)
	Grep(~/.config/gcloud/**)
	Grep(~/.ssh/**)
	Write(./secrets/**)
	Write(~/.aws/**)
	Write(~/.config/gcloud/**)
	Write(~/.ssh/**)
```

### 2.3 템플릿 대조

settings.json.tmpl 권한 영역의 항목 집합을 커밋본 allow+deny 와 `comm -3` 으로 비교한 결과, 차이는 JSON 이스케이프 표기 차이 5쌍(Windows 드라이브 경로의 콜론 이스케이프)과 템플릿 조건문 단어(`manual`, `team`)뿐이었다. 템플릿에는 7개 도구 allow 가 있고(440-446행), MultiEdit 와 Glob/Grep/Write deny 12개는 없다. 템플릿은 settings.json 과 같은 편이다.

### 2.4 항목별 이력과 판정

| # | 항목 | settings.json 이력 | YAML | 판정 |
|---|---|---|---|---|
| 1-7 | allow `CronCreate` `CronDelete` `CronList` `EnterPlanMode` `ExitPlanMode` `EnterWorktree` `ExitWorktree` | 0fb0c2829(2026-03-30) 도입(`-S CronCreate` 기준) → d85d80d7a(07-09) 재정렬 → **519b848fb(08-05) 7개 모두 삭제** → **0de8517e5(08-08) 7개 모두 복원** | 시드 73336e296 과 현재 HEAD 모두 항목 없음 (grep 0행) | settings.json 이 옳다. 7개 도구 모두 이 세션의 현행 도구 목록에 있다. **YAML 에 allow 7개 추가** |
| 8 | allow `MultiEdit` | 4aadc2061(06-03) 템플릿에서 퇴역 도구로 제거 → d85d80d7a(07-09) 로컬에서 제거 → **519b848fb 재삽입** → **0de8517e5 재제거** | 시드 73336e296 에 포함, 현재 521행 | 퇴역 도구. 공식 문서가 "the legacy `MultiEdit` tool"로 부르고, 저장소 CI 가드(`internal/template/tool_catalog_audit_test.go`)도 MultiEdit 선언을 금지한다. **YAML 에서 삭제** |
| 9-12 | deny `Write(./secrets/**)` `Write(~/.ssh/**)` `Write(~/.aws/**)` `Write(~/.config/gcloud/**)` | f30f0ac0f(07-17) 제거 → **519b848fb 재삽입** → **8df71b18d 재제거** (8df71b18d 커밋 메시지가 519b848fb 의 재삽입을 명시) | 389-413행 | 공식 문서상 Write 경로 규칙은 받아들이되 조회하지 않고 시작 시 경고한다. 같은 경로 `Edit(...)` deny 가 커밋본에 있다. **YAML 에서 삭제** |
| 13-16 | deny `Glob(...)` 4개 | 위와 같음 | 473-497행 | 공식 문서가 같은 문장에서 Glob 경로 규칙도 무시·경고 대상으로 명시. 같은 경로 `Read(...)` deny 존재. **YAML 에서 삭제** |
| 17-20 | deny `Grep(...)` 4개 | 위와 같음 | 445-469행 | 문서의 무시 목록에 Grep 은 **이름이 없다**(갭 1). 다만 문서가 "Grep and Glob search the directory the `path` argument resolves to. Claude Code applies `Read` deny rules to that directory." 라고 적고 있고, 같은 경로 `Read(...)` deny 가 커밋본에 있다. 적용 중인 settings.json 에는 이미 없으므로 YAML 삭제는 적용 권한을 바꾸지 않는다. **YAML 에서 삭제** |

### 2.5 519b848fb 가 같은 항목을 건드렸다는 증거

`git show --format= 519b848fb -- .claude/settings.json` 에서 8개 도구 이름으로 거른 출력:

```
-      "CronCreate",
-      "CronList",
-      "CronDelete",
-      "EnterWorktree",
-      "ExitWorktree",
-      "EnterPlanMode",
-      "ExitPlanMode",
+          "MultiEdit",
```

같은 커밋의 deny 쪽 diff 에 Glob/Grep/Write 경로 규칙 12줄이 `+` 로 들어 있다. 0de8517e5 는 allow 8개를 정확히 반대 방향으로, 8df71b18d 는 deny 12개를 반대 방향으로 되돌렸다.

### 2.6 공식 문서 (https://code.claude.com/docs/en/permissions, 2026-09-10 조회)

> Claude Code checks file permissions against `Edit(path)` and `Read(path)` rules only. If you write a path rule for `Write`, `NotebookEdit`, `Glob`, or the legacy `MultiEdit` tool instead, Claude Code accepts the rule but never consults it, and warns at startup, except for a `Glob` rule passed in `--allowedTools`. ... Requires Claude Code v2.1.210 or later.

> `Edit` rules apply to all built-in tools that edit files. Claude makes a best-effort attempt to apply `Read` rules to all built-in tools that read files like Grep and Glob, ...

## 3. 기준선 귀속

- 모든 목록과 수치는 이 세션에서 `.claude/worktrees/t619` @ `d1b61005d` 의 파일을 대상으로 실행한 명령의 출력이다.
- 생성기: 설치본 `moai`(MCP 서버 표기 v3.2.0-rc.5, 84fa4ece4). `internal/config/toolpolicy/codegen.go` 의 `BuildPermissions` 는 env_gate 항목(5개)을 건너뛰고 decision 별로 정렬·중복 제거한다. 드리프트 20개 중 env_gate 항목은 없다.
- 레인-7(t609 창)의 108/60 측정값이 이번 측정에서 다시 나왔다. 이월이 아니라 재측정이다.

## 4. 갭 (관측하지 않은 것)

1. **Grep 경로 규칙**: 공식 문서의 "never consults" 목록에 Grep 이 없다. Grep(path) deny 가 무시되는지는 문서로 확인하지 못했다. 판정은 "적용 중인 settings.json 에 이미 없다"는 사실과 같은 경로 Read deny 의 존재에 기댄다.
2. 7개 도구의 현존 근거는 이 세션의 도구 목록 관측이다. tools-reference 문서는 조회하지 않았다.
3. YAML 을 고친 뒤 build 출력이 커밋본 settings.json 과 **바이트 단위로** 같아지는지는 아직 재지 않았다(이번 비교는 집합 단위). run 단계 수용 기준으로 둔다.
4. 템플릿 대상 build(`--template-only`) 동작은 이번에 재지 않았다.
5. settings.local.json 과 사용자 전역 설정의 권한은 범위 밖이다.

## 5. 잔여 위험

- Grep deny 4개가 과거 누군가의 의도적 이중 방어였을 가능성. 다만 f30f0ac0f 와 8df71b18d 두 커밋 모두 이를 중복·무효로 보고 제거했고, 되돌린 기록은 519b848fb 의 생성기 덮어쓰기뿐이다.
- 드리프트 검사를 넣지 않으면 519b848fb 유형의 사고가 다시 조용히 난다. 판정만으로는 재발을 막지 못한다.

## 6. 산출물 방향 (plan 입력)

- (a) YAML 수정: allow 7개 추가, MultiEdit allow 1개와 Glob/Grep/Write deny 12개 삭제. 수정 후 build 출력이 커밋 settings.json 과 같아야 한다(적용 권한 변화 0).
- (b) 읽기 전용 드리프트 검사: 재생성하지 않고 차이만 보고한다. 대조군으로 한 항목을 일부러 어긋나게 해 검사가 실패하는지 보인다.
- (c) tool-policy.yaml 머리말의 "structurally preventing YAML↔settings.json drift" 진술을 검사 존재 사실에 맞춘다.

## 7. 운영자 결정 (레인 세션에서 직접 받음, 2026-09-10)

리드가 전한 확인을 근거로 삼지 않고, 이 레인 세션의 AskUserQuestion 으로 운영자에게 직접 물었다.

- 수리 방향: **YAML 을 settings.json 에 맞춤** (Grep deny 4개 포함 삭제. 적용 권한 변화 0)
- 드리프트 검사 위치: **make build 선행 + CI** (agents-emit-check 선례, 읽기 전용)

plan 제안 검토 뒤 받은 추가 결정: 집합 비교, 거짓 진술 전수 정정, 워킹 트리 settings.json 판독.

## 8. plan 감사 경과와 오케스트레이터 직접 확인 (2026-09-10)

SPEC-TOOLPOLICY-DRIFT-GUARD-001 plan 감사는 1회차 FAIL 0.71, 2회차 FAIL 0.79(Tier M 상한 2회 도달, 운영자가 3회차 1회 승인), 3회차 FAIL 0.89(blocking D21 1건)였다. 운영자는 "지금 고치고 진행"을 골랐고, v0.1.3 의 D21-D23 수정은 독립 재감사 없이 오케스트레이터가 아래를 직접 확인했다.

- lint: `moai spec lint SPEC-TOOLPOLICY-DRIFT-GUARD-001` → `0 error(s), 5 warning(s)` (경고는 한국어 요구 문장의 ModalityUnjudged).
- 이스케이프·Cf 스캔: 다섯 파일 모두 `(0, 0)`, 대조군 `(1, 1)`.
- D21 fixture 네 쌍(acceptance.md AC-TDG-005): 양쪽 고유 집합이 같고 어긋나는 규칙은 하나뿐임을 쌍마다 대조했다. allow 중복(Read×2), ask 중복(WebFetch×2, YAML 에도 ask WebFetch), deny 중복(Bash(rm)×2), 겹침(Read 가 allow 와 deny 양쪽, YAML 에도 같은 선언). 두 뮤턴트가 각자 자기 쪽 하위 테스트만 붉게 만든다는 서술도 이 구성과 논리적으로 맞다.
- D22 고정 fixture `{"permissions":{"allow":["Read"],"deny":["Bash(rm -rf /:*)"],}}`: 괄호 2:2·대괄호 2:2, Python `json.loads` 는 `Illegal trailing comma` 로 거부. 스크래치 루트에서 `moai tool-policy build --local-only` 출력은 `permissions object parse: invalid character '}' looking for beginning of object key string` — 권한 블록 위치는 찾고 해석 단계에서 실패한다. 대조군(쉼표만 뺀 같은 JSON)은 `allow=108 ask=0 deny=60 env_gated_skipped=5` 로 정상 생성.

갭: 뮤턴트와 새 비교 함수는 아직 코드가 없어 실행하지 않았다. 위 확인은 fixture 구성과 생성기 경로에 대한 것이며, plan-auditor 판정을 대신하지 않는다.

## 9. run 단계 결과와 AC-TDG-009 처분 (2026-09-10~11)

run 은 manager-develop 이 M1~M5 로 진행했다(커밋 `dfcf7c519`..`2ed2475cc`, push 없음). 오케스트레이터가 다시 잰 항목:

- 적용 권한 불변: `git --no-optional-locks diff --exit-code d1b61005d HEAD -- .claude/settings.json internal/template/templates/.claude/settings.json.tmpl` → `settings_diff_rc=0`.
- 드리프트 검사: `make tool-policy-drift-check` → `make_rc=0`. `go test ./internal/config/toolpolicy/ -run 'TestToolPolicyDrift' -count=1 -v` → `drift_rc=0`, Mutation 하위 PASS `7`, FailClosed 하위 PASS `10`, FAIL·SKIP `0`.
- 배선: `.github/workflows/ci.yml:92` `- '.claude/settings.json'`, `Makefile:34` `build:` 선행에 `tool-policy-drift-check`.
- 린트: `golangci-lint run --timeout=2m ./internal/config/toolpolicy/...` → `0 issues.`
- 이스케이프·Cf: 바뀐 파일 7개 모두 `(0, 0)`, 대조군 `(1, 1)`.

**AC-TDG-009 는 FAIL 이며 t609 선재 결함에 귀속한다.**

- 관측: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -timeout 600s ./internal/cli/ -run 'TestToolPolicyList_QueryFilters' -v` → `cli_rc=1`, `tool_policy_test.go:97: output missing substring "ask";`, `--- FAIL: TestToolPolicyList_QueryFilters/filter_ask`. 나머지 하위 테스트 5개 PASS.
- 귀속 근거: 이 하위 테스트는 커밋된 YAML 에 `decision: ask` 항목이 있어야 통과한다. `git show 1ac333952^:.moai/config/sections/tool-policy.yaml | grep -c 'decision: ask'` → `6`, `git show d1b61005d:… | grep -c 'decision: ask'` → `0`. `d1b61005d..HEAD` 사이 `internal/cli/` diff 없음, YAML 의 ask 줄 변경 `0`. 테스트 입력이 base 와 같다.
- develop: 로컬 develop `647ad0157` 에도 같은 하위 테스트(`tool_policy_test.go:61`)가 있고 YAML ask 는 `0` 이다. develop 트리에서 직접 실행하지는 않았다(갭).
- 처분(운영자 결정, 2026-09-11): 별도 수리 카드로 분리한다. 리드가 수리 카드 **t660** 을 발행했다(ask 0 정책 유지, 테스트를 커밋 YAML 이 아닌 ask 포함 fixture 로 검증, develop push 전 필수). 병합 판정에서 이 AC 는 t660 소관이다.

## 10. Factory 실행 종료 관문 — 보안 deep scan (2026-09-11)

- 호출: `/moai review --security --deep --branch develop`. Factory 계약 문구는 `--repo` 이나, 운영자가 이 카드 변경분만 같은 깊이로 검사하도록 결정했다(추적 파일 12,946개 대 변경 코드·설정 파일 5개). `--patch` 없음(운영자 결정).
- 실행: PRIMARY 경로(Dynamic Workflow, Claude Code 2.1.267), 에이전트 7명·오류 0. 검사 커밋 `1977040b7`, 공통 조상 `d1b61005d`.
- 결과: 사냥 후보 2건(정책·빌드·CI 영역 0, Go 영역 2), 두 건 모두 3인 패널 0/3 으로 탈락. **확인된 결함 0건 → S2, sync 진행.**
- 결과 폴더: `.moai/reports/security-deepscan-20260910T181938Z/` (자체 `.gitignore` 로 커밋되지 않음 — 이 절이 커밋되는 기록이다).
- 탈락 후보 C1 에 대한 오케스트레이터 판단: 원문 스캐너가 첫 번째 `"permissions"` 키를 읽고 표준 JSON 파서는 마지막 키를 쓰는 차이는 실재한다. 보안 결함으로는 기각에 동의한다(입력 주체가 YAML 도 함께 바꿀 수 있는 PR 작성자이고, 검사는 권한 관문이 아니라 일치 검사다). 다만 중복 키가 드리프트를 가릴 수 있는 검사 견고성 빈틈이므로 후속 카드 후보로 리드에게 전달했다. 해당 스캐너는 이 카드가 바꾸지 않은 기존 코드다: `git --no-optional-locks diff --stat d1b61005d HEAD -- internal/config/toolpolicy/settings_region.go` → 출력 없음. t619 는 `internal/cli` 테스트를 건드리지 않는다. 리드에게 사실과 결정을 전달했고, 리드는 이 적색이 develop push 전에 닫혀야 한다고 판정했다.
