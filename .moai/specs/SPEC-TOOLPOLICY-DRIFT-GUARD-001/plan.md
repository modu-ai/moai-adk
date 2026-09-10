# plan.md — SPEC-TOOLPOLICY-DRIFT-GUARD-001

## §A 맥락

- 카드: **t619** (Class C 설계 변경, Tier M). 브랜치 이름(`WT-toolpolicy-drift`)에는 카드 id 가 없으므로 이 문서와 커밋 메시지가 카드 id 를 운반한다.
- 워크트리: `.claude/worktrees/t619`. plan 작성 기준 HEAD `c7b8d110b`, 1회차 감사 HEAD `b2cbfd207`, 2회차 감사 HEAD `dc220b8fe`. `git merge-base HEAD develop` → `d1b61005d20967fdbd970ec7ec734c6d14f29dc3` (2026-09-10 측정, 로컬 develop 은 `296ba7aa5`).
- 근거 판정서: `.moai/reports/t619/verdict.md` — 20개 항목 목록, 519b848fb / 8df71b18d / 0de8517e5 이력, 공식 문서 인용, 기준선 귀속.
- 감사 이력: 1회차 FAIL 0.71, 2회차 FAIL 0.79. Tier M 상한(2회) 도달 뒤 운영자가 수정 후 3회차 감사 1회를 승인했다(2026-09-10).

### 운영자 결정 (2026-09-10, 레인 세션에서 운영자에게 직접 받음)

리드가 전한 확인을 근거로 삼지 않고, 이 레인 세션의 질문 채널로 운영자에게 직접 받았다.

1. **수리 방향**: YAML 을 커밋된 settings.json 에 맞춘다. allow 7개를 YAML 에 추가하고, `MultiEdit` allow 1개와 `Glob` / `Grep` / `Write` 경로 deny 12개를 YAML 에서 뺀다(Grep 포함). 적용 권한은 바뀌지 않는다.
2. **검사 위치**: `make build` 선행 + CI. 읽기 전용이며 재생성하지 않는다(`agents-emit-check` 선례).
3. **주장 정정**: YAML 머리말의 "구조적으로 막는다" 진술을 실제로 막는 장치(검사)에 맞춘다.

### plan 제안 검토 뒤 추가 결정 (2026-09-10, 운영자)

4. **비교 방식 = 집합 비교.** settings.json 은 바이트 그대로 둔다. 목록 내 중복과 allow/deny 겹침은 별도 단정으로 잡는다.
5. **주장 정정 범위 = 전수.** 1회차 감사로 열거 밖 두 곳이 더 드러나 대상은 다섯 곳이다(spec.md REQ-TDG-005).
6. **판독 대상 = 워킹 트리의 `.claude/settings.json`.** 권한 블록만 비교한다. 오케스트레이터 측정: primary 체크아웃의 미커밋 settings.json 수정은 권한 블록이 커밋본과 같다(`jq -S '.permissions'` 비교 → `PERM-SAME`, 권한 밖 키만 다름). 따라서 §F 위험 1 은 현재 잠복 상태다.

## §B 조사 결과 (plan 단계 실측)

| # | 질문 | 관측 | 귀결 |
|---|---|---|---|
| a | YAML 머리말을 생성기가 쓰는가 | 생성기는 넘겨받은 대상 파일만 쓴다(`internal/config/toolpolicy/codegen.go:207-253`, 쓰기 `:247`). build 대상은 로컬 settings.json(`internal/cli/tool_policy.go:116-127`)과 템플릿(`:129-164`) 둘뿐이다. YAML 을 쓰는 코드 경로는 없다 | 머리말은 손으로 고쳐도 덮이지 않는다. "머리말도 생성된다"는 진술(`tool-policy.yaml:5-9`, `types.go:1-5`)은 거짓 |
| b | 템플릿에 YAML 사본이 있는가 | `internal/template/templates/.moai/config/sections/` 에 없음. 마지막 커밋 `7d9bc4147` 이 제거 | Template-First 는 YAML 편집에 적용되지 않는다 |
| c | 템플릿 대상은 어떻게 처리되나 | 템플릿 권한 블록에 GitMode 조건문(`settings.json.tmpl:456-462`) → `PermissionsRegionHasDirective`(`codegen.go:175-182`) 참 → build 가 템플릿을 건너뜀(`tool_policy.go:139-149`) | 템플릿 대조는 범위 밖. `generated_into` 의 템플릿 항목은 사실과 다르다 |
| d | 커밋된 YAML 을 읽는 기존 테스트 | `internal/cli/tool_policy_test.go:47,108,143` — 부분 문자열·JSON 모양만 단정. `wantCount`(`:55`)는 선언만 있고 읽히지 않음 | 개수 단정 없음. 다시 돌려 확인만 한다 |
| e | CI 가 이 테스트를 도는가 | `test` 작업이 `go test ... ./...` 실행(`.github/workflows/ci.yml:208`). 경로 필터 `go_code`(`ci.yml:78-92`)에 `.claude/settings.json` 도 `.claude/**` 도 없음 → settings.json 만 바꾼 변경은 대체 작업(`ci.yml:317-347`)이 초록 보고 | 필터에 한 줄 추가. 별도 CI 단계는 필요 없음 |
| f | 파일이 없을 때 | 테스트는 moai-adk 모듈 안에서만 컴파일된다 | 이 저장소에서 부재는 곧 실패. 건너뛰기 금지 |
| g | 바이트 비교 vs 집합 비교 | 스크래치 build 뒤 `diff` 종료 1, 4개 헝크, 커밋본 160줄 vs 생성본 166줄(실제 차이 20개). 커밋본 목록은 C 정렬이 아님. 생성기는 정렬(`loader.go:87-94`)·자체 들여쓰기(`settings_region.go:237-250`) | 집합 비교(결정 4). `defaultMode` 는 매개변수로 들어오고(`codegen.go:54`) build 가 기존 값을 보존(`codegen.go:217-224`)하므로 제외 |
| h | env_gate 항목 | YAML 에 env_gate 항목 5개: `Write` deny(`args_pattern: ""`), `Write` allow(`.moai/specs/SPEC-*/{spec,plan,acceptance}.md`), `WebSearch` deny, `WebFetch` deny, `Read` deny(이미지 확장자). 정렬 JSON sha256 `5e0cba521c5c81a2d7bb82fbba52c59b027d2329b2c0bf8b912ed6b2150e6d27` | 삭제 대상 12개와 구별하고, 내용 보존을 해시로 확인한다(2회차 D15) |
| i | 주장 진술 전수 스윕 | `grep -n -i -E 'structurally\|prevent\|audit surface\|drift\|are generated\|both generated'` 를 `internal/config/toolpolicy/{codegen,loader,settings_region,tier_render,types}.go`, `internal/cli/tool_policy.go`, YAML 1-48행에 실행 → 방지·생성 주장은 `types.go:5`, `types.go:97`, `tool-policy.yaml:6-9`, `:13-14` 뿐. 2회차 감사의 저장소 전체 스윕이 추가로 `codegen_test.go:209-213` 을 확인(§5 범위 밖) | REQ-TDG-005 대상 다섯 곳 확정, 제외 목록에 `codegen_test.go:209-213` 명시 |
| j | 워크트리 세션 가드가 허용하는 셸 형태 | 거부: `trap` 포함 명령, heredoc 으로 python·bash 에 스크립트를 넘기는 명령, 변수로 계산된 스크립트 경로를 python 에 넘기는 명령. 허용: 리터럴 절대 경로의 python 스크립트 호출, 리터럴 경로의 `shasum … \| cmp -s - <sha> && cp <bak> <target> && echo … \|\| echo …` 사슬, `git status --porcelain --untracked-files=all > <file>` (모두 2026-09-10 실측) | AC-TDG-004 조건부 복원과 AC-TDG-003 상태 비교를 허용 형태로 쓴다 |

추가 관측: YAML 로더는 `metadata` 필드를 검증하지 않는다(`loader.go` 에서 `metadata`/`generated` 검증 코드 0건). `generated_into` 항목 삭제가 로딩을 깨지 않는다.

### Reference

- Reference: `Makefile:34` — `build:` 선행 목록 (`agents-emit-check commands-emit-check templ-generate`)
- Reference: `Makefile:41-49` — `agents-emit-check` 주석과 타깃. 레시피가 `@` 로 시작하고, `-count=1` 로 돌리며, 실패 시 재생성 동사를 이름으로 밝히는 stderr 문구
- Reference: `Makefile:54-60` — `commands-emit-check` (같은 모양)
- Reference: `internal/config/toolpolicy/codegen.go:54-99` — `BuildPermissions` (env_gate 건너뜀, 정렬, 결정별 중복 제거). 검사의 기대 집합은 이 함수로 유도한다
- Reference: `internal/config/toolpolicy/settings_region.go:42-87` — `locatePermissionsRegion` (권한 블록 부재 오류), `:138-165` — `extractPermissions` ("permissions object parse" 오류)
- Reference: `internal/config/toolpolicy/loader.go` — `Load` (YAML 파싱·검증 오류 경로)
- Reference: `internal/cli/tool_policy.go:139-149` — 템플릿 권한 블록에 조건문이 있으면 건너뜀
- Reference: `internal/config/toolpolicy/types.go:1-5`, `:94-98` — 정정할 문서 주석 두 곳
- Reference: `.moai/config/sections/tool-policy.yaml:5-9`, `:11-16`, `:54` — 정정할 머리말 두 문단과 `generated_into` 항목
- Reference: `.github/workflows/ci.yml:78-92` — `go_code` 경로 필터, `:53-72` — 필터 근거 주석 블록

## §C 기술 접근

- 검사는 `internal/config/toolpolicy` 패키지의 새 테스트 파일이다. 새 프로덕션 API 는 만들지 않는다. 같은 패키지이므로 `Load`, `BuildPermissions`, `extractPermissions` 를 그대로 쓴다.
  - 기대 집합: 워킹 트리 YAML 을 로드해 `BuildPermissions` 로 유도(생성기와 같은 규칙).
  - 실제 집합: 워킹 트리 `.claude/settings.json` 에서 `extractPermissions` 로 판독.
  - 두 원천이 서로 다르므로 비교가 항진명제가 되지 않는다.
- 비교 함수는 경로 두 개를 받아 차이 목록 또는 오류를 돌려준다.
- **실패 원인 구별 계약(고정, 2회차 D11).** 테스트 파일은 원인별 센티널 오류 네 개를 정의하고, 비교 함수는 모든 실패를 `fmt.Errorf("…: %w", <센티널>)` 로 감싸 돌려준다. 한 오류는 센티널을 정확히 하나만 감싼다.

  | 센티널 | 원인 | 판정 순서 |
  |---|---|---|
  | `errDriftInputMissing` | YAML 또는 settings 파일이 없음 | 1 |
  | `errDriftInputParse` | YAML 해석·검증 실패, 또는 권한 블록 JSON 해석 실패 | 2 |
  | `errDriftNoPermissionsRegion` | settings 에 `"permissions"` 키가 없음 | 3 |
  | `errDriftEmptySet` | YAML allow·YAML deny·settings allow·settings deny 중 하나라도 빔 | 4 |

  판정은 표의 순서로 하고, 앞 단계에서 실패하면 뒤 단계로 진행하지 않는다. 따라서 해석에 실패한 입력은 빈 집합 판정에 도달하지 않는다.
- 차이 보고 형식(고정): 한 줄에 하나, `<decision> only-in-yaml: <specifier>` / `<decision> only-in-settings: <specifier>`. `<decision>` 은 `allow` / `ask` / `deny` 중 하나다.
- 테스트 이름(수용 기준 명령이 선택자로 사용, 고정):
  - `TestToolPolicyDrift_CommittedSettingsMatchYAML` — 커밋 트리 집합 비교
  - `TestToolPolicyDrift_NoDuplicatesOrOverlap` — 커밋본 allow·ask·deny 목록 내 중복과 allow/deny 겹침
  - `TestToolPolicyDrift_Mutation` — 하위 테스트 7개: `settings_side_missing`, `yaml_side_flip`, `duplicate_settings_allow`, `duplicate_settings_ask`, `duplicate_settings_deny`, `allow_deny_overlap`, `unmutated`
  - `TestToolPolicyDrift_FailClosed` — 하위 테스트 9개: `missing_yaml`, `missing_settings`, `malformed_yaml`, `malformed_settings_json`, `no_permissions_region`, `empty_yaml_allow`, `empty_yaml_deny`, `empty_settings_allow`, `empty_settings_deny`. 각 하위 테스트는 `errors.Is(err, <자기 원인 센티널>)` 가 참이고 나머지 세 센티널에 대한 `errors.Is` 가 모두 거짓임을 단정한다. `missing_*` 는 `errDriftInputMissing`, `malformed_*` 는 `errDriftInputParse`, `no_permissions_region` 은 `errDriftNoPermissionsRegion`, `empty_*` 는 `errDriftEmptySet` 이 자기 원인이다.
- **Makefile 타깃(고정, 2회차 D18).** 실제 Makefile 에 들어갈 줄은 아래와 같다. 레시피 줄은 탭으로 시작하고, 레시피 본문은 `@` 로 시작한다(선례 `Makefile:48` 과 같다 — 성공 실행에서 레시피 줄 자체가 문구를 출력하지 않도록).

  ```makefile
  tool-policy-drift-check: ## Verify tool-policy.yaml and the .claude/settings.json permissions block declare the same sets (read-only; never regenerates)
  	@go test ./internal/config/toolpolicy/... -run 'TestToolPolicyDrift_(CommittedSettingsMatchYAML|NoDuplicatesOrOverlap)$$' -count=1 \
  		|| { printf 'tool-policy drift: .claude/settings.json permissions differ from .moai/config/sections/tool-policy.yaml — reconcile tool-policy.yaml, then regenerate with `moai tool-policy build --local-only`\n' >&2; exit 1; }
  ```

  이 테스트에는 쓰기 경로가 없으므로 비울 갱신용 환경 변수가 없다. 수용 기준은 차이 줄(`only-in-…`)과 안내 문구의 두 조각(`reconcile tool-policy.yaml`, `moai tool-policy build --local-only`)을 각각 센다.
- YAML 신규 7개 항목의 필드 제안: `risk_tier` 는 `CronList` / `EnterPlanMode` / `ExitPlanMode` 가 `read`, `CronCreate` / `CronDelete` / `EnterWorktree` / `ExitWorktree` 가 `write`. `owner_agent: orchestrator`, `source: ".claude/settings.json#permissions.allow"`. 검사는 명세자와 결정만 비교하므로 이 값은 판단이다(§F 위험 2).
- 삭제 13개는 `args_pattern` 이 없는 `MultiEdit` allow 1개와, env_gate 가 없고 `args_pattern` 이 네 경로 중 하나인 `Glob`/`Grep`/`Write` deny 12개로 한정한다. env_gate 항목 5개는 한 글자도 바꾸지 않는다.

## §D 마일스톤 (결정이 바뀔 가능성이 큰 것부터, 의존 순서)

### M1 — 비교 의미와 실패 폐쇄 규칙 확정, 검사 테스트 작성 [Priority High]

- 집합 비교, 제외 필드, `ask` 부재 = 빈 목록, 중복·겹침 별도 단정, §C 실패 원인 구별 계약을 테스트로 고정한다.
- 현재 트리에서 `TestToolPolicyDrift_CommittedSettingsMatchYAML` 은 **붉은색이어야 하며**, 판정서의 20개 명세자를 서로 다른 20줄로 나열해야 한다(AC-TDG-001 재현 증거 명령). 출력은 progress.md §E.2 에 기록한다.
- M2 전에 AC-TDG-010 의 env_gate 해시 기준선을 다시 재서 §E.2 에 기록한다.
- 선행: 없음.

### M2 — YAML 정합 복구 [Priority High]

- allow 7개 추가, `MultiEdit` allow 와 env_gate 없는 `Glob` / `Grep` / `Write` 경로 deny 12개 삭제.
- AC-TDG-001 초록, AC-TDG-002(스크래치 build 집합 일치), AC-TDG-003(settings.json 바이트·워킹 트리 상태 불변), AC-TDG-010 확인.
- M2 결과는 AC-TDG-004 전에 커밋한다(AC-TDG-004 의 마지막 `git diff` 가 커밋된 YAML 을 기준으로 삼는다).
- 선행: M1.

### M3 — 뮤테이션 대조 [Priority Medium]

- 자동: `TestToolPolicyDrift_Mutation`, `TestToolPolicyDrift_FailClosed` (AC-TDG-005, AC-TDG-006).
- 수동: 카드 워크트리 YAML 에서 `ExitWorktree` 항목을 지워 붉은색과 안내 문구 → 조건부 복원, sha256 일치 → 초록 (AC-TDG-004). 복원 규칙은 §G.
- 뮤턴트 점검 세 가지(AC-TDG-005·006 DoD): 비교 함수가 항상 빈 차이를 돌려주는 뮤턴트, YAML 해석 오류를 빈 정책으로 삼키는 뮤턴트, 권한 블록 JSON 해석 오류를 빈 목록으로 삼키는 뮤턴트. 각 뮤턴트에서 붉어져야 하는 하위 테스트 이름은 acceptance.md 가 고정한다.
- 선행: M2.

### M4 — 연결: Makefile 과 CI [Priority Medium]

- §C 의 타깃을 그대로 추가, `.PHONY` 와 `build:` 선행 목록에 등록.
- `ci.yml` `go_code` 필터 블록에 `- '.claude/settings.json'` 추가, `:53-72` 근거 주석에 이 테스트가 읽는다는 한 줄 추가.
- AC-TDG-008.
- 선행: M1.

### M5 — 주장 정정과 무회귀 확인 [Priority Low]

- REQ-TDG-005 의 다섯 곳 정정: `tool-policy.yaml:5-9` 머리말 문단, `tool-policy.yaml:11-16` "ANALOGOUS drift class" 문단, `types.go:1-5` 패키지 주석, `types.go:94-98` `Metadata` 주석, `tool-policy.yaml:54` `generated_into` 템플릿 항목 (AC-TDG-007). 정정 전에 AC-TDG-007 기준선을 한 번 재서 §E.2 에 기록한다.
- 영향 테스트 재실행 (AC-TDG-009).
- 선행: M2, M4 (정정 문구가 타깃 이름을 가리키므로).

## §E 변경 파일

| 표지 | 경로 | 이유 |
|---|---|---|
| [MODIFY] | `.moai/config/sections/tool-policy.yaml` | 20개 항목 정합(M2), 머리말 두 문단·`generated_into` 템플릿 항목 정정(M5) |
| [NEW] | `internal/config/toolpolicy/drift_check_test.go` | 검사 본체, 센티널 오류, 뮤테이션·실패 폐쇄 하위 테스트(M1, M3) |
| [MODIFY] | `Makefile` | `tool-policy-drift-check` 타깃, `.PHONY`, `build:` 선행(M4) |
| [MODIFY] | `.github/workflows/ci.yml` | `go_code` 필터에 `.claude/settings.json` 추가와 근거 주석(M4) |
| [MODIFY] | `internal/config/toolpolicy/types.go` | `:1-5` 패키지 주석, `:94-98` `Metadata` 주석 정정(M5) |
| [MODIFY] | `.moai/specs/SPEC-TOOLPOLICY-DRIFT-GUARD-001/progress.md` | §E.2/§E.3 는 run 단계 소유자가 채움 |

변경하지 않는 파일: `.claude/settings.json`, `internal/template/templates/.claude/settings.json.tmpl`, `internal/cli/tool_policy.go`, `internal/config/toolpolicy/codegen.go`, `internal/config/toolpolicy/codegen_test.go`, 종결된 SPEC-V3R6-TOOL-POLICY-SSOT-001 산출물. 변경 파일 중 템플릿 트리 아래에 있는 것이 없으므로 Template-First 미러 대상이 없다.

## §F 위험

1. **워킹 트리 판독의 부작용.** 런타임이 tracked settings.json 권한 블록을 미커밋으로 바꾸면 로컬 `make build` 가 붉어진다. 오케스트레이터 측정상 현재 primary 체크아웃의 미커밋 수정은 권한 블록 밖에만 있다(`PERM-SAME`) — 잠복 위험이다.
2. **새 항목의 `risk_tier` 판단.** 검사는 명세자와 결정만 본다. 7개 항목의 위험 등급과 감사 문구는 검증되지 않는다.
3. **`moai update` 뒤 붉은색.** update 는 settings.json 을 템플릿에서 다시 렌더한다. GitMode 에 따라 `git commit` / `git tag` / `git push` allow 가 빠지면 검사가 붉어진다. 올바른 신호이지만 놀랄 수 있다. CLAUDE.local.md §2.3 의 update 후 복원 절차가 이미 있다.
4. **Grep 경로 deny 의 문서 공백.** 공식 문서의 무시 목록에 Grep 은 이름이 없다(판정서 §4 갭 1). 적용 중인 settings.json 에는 이미 없으므로 적용 권한은 어느 쪽이든 바뀌지 않는다.
5. **집합 비교가 가리는 것.** 순서와 서식은 의도적으로 보지 않는다. 중복·겹침은 별도 단정으로 막는다.
6. **수동 뮤테이션 중단.** AC-TDG-004 는 추적 파일을 제자리에서 바꾼다. 세션이 중간에 끊기면 변이된 YAML 이 남을 수 있다. 완화는 §G 복원 규칙.

## §G 금지 사항과 복원 규칙

- 실제 트리 대상 `moai tool-policy build` 실행 금지(스크래치 사본만).
- 실행 중 세션이 읽는 워킹 트리 `.claude/settings.json` 에 대한 뮤테이션 금지(`t.TempDir()` 격리 사본은 허용).
- 검사 안의 `t.Skip` 금지.
- env_gate 항목 삭제·수정 금지.
- 산출물에 역슬래시-u 이스케이프 표기 금지.
- **AC-TDG-004 복원 규칙(조건부, 2회차 D13).**
  1. 변이 전에 백업 경로, 원본 sha256, 변이 뒤 sha256 을 progress.md §E.2 에 적는다.
  2. 복원은 "현재 파일 sha256 이 기록한 변이 뒤 sha256 과 같을 때만 백업을 복사한다" 는 한 줄 사슬로만 한다(acceptance.md AC-TDG-004 의 리터럴 경로 명령). 가드 호환은 스크래치 사본에서 실측했다(일치 시 복원, 불일치 시 복사하지 않고 끼어든 수정 보존).
  3. 사슬이 `restore=STOPPED` 를 출력하면(현재 파일이 기록한 변이 상태가 아니면) 복원하지 않고 멈춘다. 현재 sha256 과 백업 경로를 progress.md §E.2 에 적고 리드에게 보고한다. 이후 처분은 사람이 정한다.
  4. 이어받은 세션은 다른 어떤 명령보다 먼저 2-3 을 실행한다.
  5. `git restore` / `git checkout --` 로 복원하지 않는다(M2 이후 미커밋 작업을 버릴 수 있다).

## §H plan-audit 결함 처분

### 1회차 (`.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-1.md`)

| 결함 | 등급 | 처분 | 반영 위치 |
|---|---|---|---|
| D1 AC-TDG-010 env_gate | blocking | 반영. 질의를 env_gate 없음 + 네 경로로 좁혀 기대값 0 유지, env-gated `Write` deny 가 1로 남는 양성 대조 추가. 실측: 좁힌 질의 `12`, env-gated Write deny `1` | acceptance.md AC-TDG-010, spec.md REQ-TDG-001·§4, plan.md §B h·§C |
| D2 REQ-TDG-005 열거 누락 | blocking | 반영. `tool-policy.yaml:11-16`, `types.go:1-5` 추가. 부재 grep 과 기준선 실측: `ANALOGOUS drift class` `1`, `comment (audit surface) are generated` `1` | spec.md REQ-TDG-005, acceptance.md AC-TDG-007, plan.md M5·§E |
| D3 해석 실패 미검증 | blocking | 반영. `malformed_yaml`, `malformed_settings_json` 추가. 2회차 D11 로 오라클 보강 | acceptance.md AC-TDG-006, plan.md §C |
| D4 붉은색 증거 명령 | optional | 반영. 붉은색 실행 명령 명시, `sort -u` 뒤 개수와 결정·쪽별 분할 개수로 판정 | acceptance.md AC-TDG-001 |
| D5 레시피 에코 | optional | 반영. `@` 접두 명시, 판별 문자열을 `only-in-` 차이 줄로 교체. 2회차 D14 로 안내 문구 검증 복원 | plan.md §C, acceptance.md AC-TDG-004 |
| D6 고정 기준 SHA | optional | 반영. 기준을 `git merge-base HEAD develop` 로 바꿈 | acceptance.md AC-TDG-003 |
| D7 문자열 존재 추정 | optional | 반영. `make -n build` 검사 본체 토큰, `go_code:` 블록 한정 grep. 2회차 D17 로 선행 목록 직접 확인 추가 | acceptance.md AC-TDG-008 |
| D8 제자리 변이 중단 | optional | 부분 반영 → 2회차 D13 으로 조건부 복원 완성. `trap` 단일 호출은 가드가 거부(§B j) | acceptance.md AC-TDG-004, plan.md §F 6·§G |
| D9 GEARS 라벨·서술어 | optional | 반영 | spec.md §2 |
| D10 비어 있음 대상 쪽 | optional | 반영. 네 집합 명시, 하위 테스트 쪽별 분리(FailClosed 9개) | spec.md REQ-TDG-004, acceptance.md AC-TDG-006·009, plan.md §C |

### 2회차 (`.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-2.md`)

| 결함 | 등급 | 처분 | 반영 위치 |
|---|---|---|---|
| D11 뮤턴트가 빈 집합 규칙에 가려짐 | blocking | 반영. 실패 원인 구별을 요구로 올리고(REQ-TDG-004), 센티널 네 개와 판정 순서를 고정(§C). 각 FailClosed 하위 테스트는 자기 원인 `errors.Is` 참 + 나머지 세 원인 거짓을 단정. DoD 뮤턴트는 해석 오류를 빈 정책으로 삼키면 `errDriftEmptySet` 이 나오므로 `malformed_yaml`(또는 `malformed_settings_json`)의 `errors.Is(err, errDriftInputParse)` 가 거짓이 되어 붉어진다 | spec.md REQ-TDG-004, plan.md §C·M3, acceptance.md AC-TDG-006, spec-compact.md |
| D12 §4 뮤테이션 범위 | optional | 반영. 제자리 수동 뮤테이션은 YAML 만, `t.TempDir()` 격리 사본은 자동 대조에서 허용 | spec.md §4, plan.md §G |
| D13 복원 전 확인이 복원을 막지 않음 | blocking | 반영. 리터럴 경로 `shasum … \| cmp -s - <변이 sha> && cp <백업> <대상> && echo restore=DONE \|\| echo restore=STOPPED` 한 줄로 복원을 조건화. 가드 호환·동작 실측(§B j): 일치 시 `restored=yes`, 끼어든 수정이 있으면 `restored=NO_stop` 이고 그 수정이 남음(`grep -c` `1`). 멈춤 시 절차는 §G | acceptance.md AC-TDG-004, plan.md §B j·§G |
| D14 조정 안내 미검증 | blocking | 반영. 뮤테이션 로그에서 `reconcile tool-policy.yaml` 과 `moai tool-policy build --local-only` 각각 `1` 이상, 복원 로그에서 각각 `0` | acceptance.md AC-TDG-004, plan.md §C |
| D15 env_gate 개수만 판정 | optional | 반영. 정렬 JSON sha256 행 추가. 실측 `5e0cba521c5c81a2d7bb82fbba52c59b027d2329b2c0bf8b912ed6b2150e6d27` | acceptance.md AC-TDG-010, plan.md §B h·M1 |
| D16 쓰기 표면 | optional | 반영. 검사 전후 `git status --porcelain --untracked-files=all` 출력 비교(가드 허용 실측, §B j) | acceptance.md AC-TDG-003 |
| D17 선행 순서·세 목록 중복 | optional | 반영. (a) `grep -c -E '^build:.*tool-policy-drift-check' Makefile` 행 추가(현재 `0`, 대조 `^build:.*agents-emit-check` `1`). (b) 중복 하위 테스트를 allow·ask·deny 셋으로 나눔(Mutation 7개) | acceptance.md AC-TDG-005·008, plan.md §C |
| D18 레시피 조각 | optional | 반영. 레시피를 펜스 코드 블록의 실제 Makefile 줄로 교체 | plan.md §C |
| D19 제외 목록 | optional | 반영. `codegen_test.go:209-213` 을 REQ-TDG-005 제외 문단과 §5 에 명시 | spec.md REQ-TDG-005·§5, plan.md §B i·§E |
| D20 DoD Cf 범위 | optional | 반영. 대상을 §E 변경 파일 다섯 개로 한정하고 판정 명령·대조군 명시. 현재 기준선(기존 네 파일) 실측 모두 `(0, 0)`, 대조 `(1, 1)` | acceptance.md §D.3 |
