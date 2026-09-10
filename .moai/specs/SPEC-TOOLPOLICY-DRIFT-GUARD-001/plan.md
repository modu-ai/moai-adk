# plan.md — SPEC-TOOLPOLICY-DRIFT-GUARD-001

## §A 맥락

- 카드: **t619** (Class C 설계 변경, Tier M). 브랜치 이름(`WT-toolpolicy-drift`)에는 카드 id 가 없으므로 이 문서와 커밋 메시지가 카드 id 를 운반한다.
- 워크트리: `.claude/worktrees/t619`, plan 기준 HEAD `c7b8d110b` (base: 로컬 develop `d1b61005d`).
- 근거 판정서: `.moai/reports/t619/verdict.md` — 20개 항목 목록, 519b848fb / 8df71b18d / 0de8517e5 이력, 공식 문서 인용, 기준선 귀속.

### 운영자 결정 (2026-09-10, 레인 세션에서 운영자에게 직접 받음)

리드가 전한 확인을 근거로 삼지 않고, 이 레인 세션의 질문 채널로 운영자에게 직접 받았다.

1. **수리 방향**: YAML 을 커밋된 settings.json 에 맞춘다. allow 7개를 YAML 에 추가하고, `MultiEdit` allow 1개와 `Glob` / `Grep` / `Write` deny 12개를 YAML 에서 뺀다(Grep 포함). 적용 권한은 바뀌지 않는다.
2. **검사 위치**: `make build` 선행 + CI. 읽기 전용이며 재생성하지 않는다(`agents-emit-check` 선례).
3. **주장 정정**: YAML 머리말의 "구조적으로 막는다" 진술을 실제로 막는 장치(검사)에 맞춘다.

### plan 제안 검토 뒤 추가 결정 (2026-09-10, 운영자)

4. **비교 방식 = 집합 비교.** settings.json 은 바이트 그대로 둔다. 목록 내 중복과 allow/deny 겹침은 별도 단정으로 잡는다.
5. **주장 정정 범위 = 전수.** YAML 머리말의 두 진술(드리프트 방지, 머리말 생성), `types.go:94-98` 문서 주석, `metadata.generated_into` 의 템플릿 항목.
6. **판독 대상 = 워킹 트리의 `.claude/settings.json`.** 권한 블록만 비교한다. 오케스트레이터 측정: primary 체크아웃의 미커밋 settings.json 수정은 권한 블록이 커밋본과 같다(`jq -S '.permissions'` 비교 → `PERM-SAME`, 권한 밖 키만 다름). 따라서 §E 위험 1 은 현재 잠복 상태다.

## §B 조사 결과 (plan 단계 실측)

| # | 질문 | 관측 | 귀결 |
|---|---|---|---|
| a | YAML 머리말을 생성기가 쓰는가 | 생성기는 넘겨받은 대상 파일만 쓴다(`internal/config/toolpolicy/codegen.go:207-253`, 쓰기 `:247`). build 대상은 로컬 settings.json(`internal/cli/tool_policy.go:116-127`)과 템플릿(`:129-164`) 둘뿐이다. YAML 을 쓰는 코드 경로는 없다 | 머리말은 손으로 고쳐도 덮이지 않는다. "머리말도 생성된다"는 진술(`tool-policy.yaml:5-9`) 자체가 거짓 |
| b | 템플릿에 YAML 사본이 있는가 | `internal/template/templates/.moai/config/sections/` 에 없음. 마지막 커밋 `7d9bc4147` 이 제거 | Template-First 는 YAML 편집에 적용되지 않는다 |
| c | 템플릿 대상은 어떻게 처리되나 | 템플릿 권한 블록에 GitMode 조건문(`settings.json.tmpl:456-462`) → `PermissionsRegionHasDirective`(`codegen.go:175-182`) 참 → build 가 템플릿을 건너뜀(`tool_policy.go:139-149`) | 템플릿 대조는 범위 밖. `generated_into` 의 템플릿 항목은 사실과 다르다 |
| d | 커밋된 YAML 을 읽는 기존 테스트 | `internal/cli/tool_policy_test.go:47,108,143` — 부분 문자열·JSON 모양만 단정. `wantCount`(`:55`)는 선언만 있고 읽히지 않음. 나머지는 임시 픽스처(`internal/core/project/autonomy_bundle_test.go:188-206`) 또는 섹션 이름만 참조 | 개수 단정 없음. 다시 돌려 확인만 한다 |
| e | CI 가 이 테스트를 도는가 | `test` 작업이 `go test ... ./...` 실행(`.github/workflows/ci.yml:208`). 그러나 경로 필터 `go_code`(`ci.yml:78-92`)에 `.claude/settings.json` 도 `.claude/**` 도 없음 → settings.json 만 바꾼 변경은 대체 작업(`ci.yml:317-347`)이 초록 보고 | 필터에 한 줄 추가가 필요. 별도 CI 단계는 필요 없음 |
| f | 파일이 없을 때 | 테스트는 moai-adk 모듈 안에서만 컴파일된다. 소비자 프로젝트는 이 테스트를 돌리지 않는다 | 이 저장소에서 부재는 곧 실패. 건너뛰기 금지 |
| g | 바이트 비교 vs 집합 비교 | 스크래치 build 뒤 `diff` 종료 1, 4개 헝크, 커밋본 160줄 vs 생성본 166줄(실제 차이 20개). 커밋본 목록은 C 정렬이 아님(`allow-NOT-C-sorted`, `deny-NOT-C-sorted`). 생성기는 정렬(`loader.go:87-94`)·자체 들여쓰기(`settings_region.go:237-250`) | 집합 비교(결정 4). `defaultMode` 는 매개변수로 들어오고(`codegen.go:54`) build 가 기존 값을 보존(`codegen.go:217-224`)하므로 제외 |

추가 관측: YAML 로더는 `metadata` 필드를 검증하지 않는다(`internal/config/toolpolicy/loader.go` 에서 `metadata`/`generated` 검증 코드 0건, 주석 1건). `generated_into` 항목 삭제가 로딩을 깨지 않는다. run 단계에서 AC-TDG-010 의 `list` 성공으로 다시 확인한다.

### Reference

- Reference: `Makefile:34` — `build:` 선행 목록 (`agents-emit-check commands-emit-check templ-generate`)
- Reference: `Makefile:41-49` — `agents-emit-check` 주석과 타깃. `-count=1` 로 돌리고, 실패 시 재생성 동사를 이름으로 밝히는 stderr 문구
- Reference: `Makefile:54-60` — `commands-emit-check` (같은 모양)
- Reference: `internal/config/toolpolicy/codegen.go:54-99` — `BuildPermissions` (env_gate 건너뜀, 정렬, 결정별 중복 제거). 검사의 기대 집합은 이 함수로 유도한다
- Reference: `internal/config/toolpolicy/settings_region.go:138-165` — `extractPermissions` (권한 블록 판독, 파싱 실패 시 오류)
- Reference: `internal/cli/tool_policy.go:139-149` — 템플릿 권한 블록에 조건문이 있으면 건너뜀
- Reference: `internal/config/toolpolicy/types.go:94-98` — 정정할 문서 주석
- Reference: `.moai/config/sections/tool-policy.yaml:5-9` (머리말 진술), `:52-54` (`generated_into`)
- Reference: `.github/workflows/ci.yml:78-92` — `go_code` 경로 필터, `:53-72` — 필터 근거 주석 블록

## §C 기술 접근

- 검사는 `internal/config/toolpolicy` 패키지의 새 테스트 파일이다. 새 프로덕션 API 는 만들지 않는다. 같은 패키지이므로 `BuildPermissions` 와 `extractPermissions` 를 그대로 쓴다.
  - 기대 집합: 워킹 트리 YAML 을 로드해 `BuildPermissions` 로 유도 (생성기와 같은 규칙).
  - 실제 집합: 워킹 트리 `.claude/settings.json` 에서 `extractPermissions` 로 판독.
  - 두 원천이 서로 다르므로 비교가 항진명제가 되지 않는다.
- 비교 함수는 경로 두 개를 받아 차이 목록을 돌려준다. 커밋 트리 테스트는 저장소 루트 상대 경로(`../../../`)로 부르고, 뮤테이션·실패 폐쇄 테스트는 `t.TempDir()` 사본으로 부른다. 같은 함수를 부르므로 뮤테이션 대조가 실제 검사를 지킨다.
- 테스트 이름(수용 기준 명령이 선택자로 사용):
  - `TestToolPolicyDrift_CommittedSettingsMatchYAML` — 커밋 트리 집합 비교
  - `TestToolPolicyDrift_NoDuplicatesOrOverlap` — 커밋본 목록 내 중복·allow/deny 겹침
  - `TestToolPolicyDrift_Mutation` — 하위 테스트 `settings_side_missing`, `yaml_side_flip`, `duplicate_in_settings`, `allow_deny_overlap`, `unmutated`
  - `TestToolPolicyDrift_FailClosed` — 하위 테스트 `missing_yaml`, `missing_settings`, `no_permissions_region`, `empty_allow`, `empty_deny`
- 차이 보고 형식: 한 줄에 하나, `<decision> only-in-yaml: <specifier>` / `<decision> only-in-settings: <specifier>`.
- Makefile 타깃 `tool-policy-drift-check`: `go test ./internal/config/toolpolicy/... -run 'TestToolPolicyDrift_(CommittedSettingsMatchYAML|NoDuplicatesOrOverlap)$$' -count=1` 실패 시 `tool-policy drift: .claude/settings.json permissions differ from .moai/config/sections/tool-policy.yaml — reconcile tool-policy.yaml, then regenerate with ` + "`moai tool-policy build --local-only`" 문구를 stderr 로. 이 테스트에는 쓰기 경로가 없으므로 비울 갱신용 환경 변수가 없다.
- YAML 신규 7개 항목의 필드 제안: `risk_tier` 는 `CronList` / `EnterPlanMode` / `ExitPlanMode` 가 `read`, `CronCreate` / `CronDelete` / `EnterWorktree` / `ExitWorktree` 가 `write`. `owner_agent: orchestrator`, `source: ".claude/settings.json#permissions.allow"`. 검사는 명세자와 결정만 비교하므로 이 값은 판단이다(§E 위험 2).

## §D 마일스톤 (결정이 바뀔 가능성이 큰 것부터, 의존 순서)

### M1 — 비교 의미와 실패 폐쇄 규칙 확정, 검사 테스트 작성 [Priority High]

- 집합 비교, 제외 필드(`defaultMode`·기타 권한 키), `ask` 부재 = 빈 목록, 중복·겹침 별도 단정, 부재·빈 입력 실패를 테스트로 고정한다.
- 현재 트리에서 `TestToolPolicyDrift_CommittedSettingsMatchYAML` 은 **붉은색이어야 하며**, 판정서의 20개 항목을 정확히 나열해야 한다. 이것이 재현 증거다. 출력은 `.moai/specs/SPEC-TOOLPOLICY-DRIFT-GUARD-001/progress.md` §E.2 에 기록한다.
- 선행: 없음.

### M2 — YAML 정합 복구 [Priority High]

- allow 7개 추가, `MultiEdit` allow 와 `Glob` / `Grep` / `Write` deny 12개 삭제.
- AC-TDG-001 초록, AC-TDG-002(스크래치 build 집합 일치), AC-TDG-003(settings.json 바이트 불변) 확인.
- 선행: M1 (붉은색을 먼저 보았어야 초록이 의미를 가진다).

### M3 — 뮤테이션 대조 [Priority Medium]

- 자동: `TestToolPolicyDrift_Mutation`, `TestToolPolicyDrift_FailClosed` (AC-TDG-005, AC-TDG-006).
- 수동: 카드 워크트리 YAML 에서 한 항목을 지워 붉은색 → 저장해 둔 사본으로 복원, sha256 일치 → 초록 (AC-TDG-004).
- 뮤턴트 점검: 비교 함수가 항상 "차이 없음"을 돌려주도록 바꾸면 뮤테이션 하위 테스트가 실패하는지 확인한다.
- 선행: M2 (`ExitWorktree` 항목이 YAML 에 있어야 수동 뮤테이션 대상이 된다).

### M4 — 연결: Makefile 과 CI [Priority Medium]

- `tool-policy-drift-check` 타깃 추가, `.PHONY` 와 `build:` 선행 목록에 등록.
- `ci.yml` `go_code` 필터에 `- '.claude/settings.json'` 추가, `:53-72` 근거 주석에 이 테스트가 읽는다는 한 줄 추가.
- AC-TDG-008.
- 선행: M1 (타깃이 부를 테스트가 있어야 한다).

### M5 — 주장 정정과 무회귀 확인 [Priority Low]

- `tool-policy.yaml` 머리말 두 진술, `types.go:94-98` 주석, `metadata.generated_into` 템플릿 항목 정정 (AC-TDG-007).
- 영향 테스트 재실행 (AC-TDG-009), 항목 단위 결과 확인 (AC-TDG-010).
- 선행: M2, M4 (정정 문구가 타깃 이름을 가리키므로).

## §E 변경 파일

| 표지 | 경로 | 이유 |
|---|---|---|
| [MODIFY] | `.moai/config/sections/tool-policy.yaml` | 20개 항목 정합(M2), 머리말 두 진술·`generated_into` 템플릿 항목 정정(M5) |
| [NEW] | `internal/config/toolpolicy/drift_check_test.go` | 검사 본체, 뮤테이션·실패 폐쇄 하위 테스트(M1, M3) |
| [MODIFY] | `Makefile` | `tool-policy-drift-check` 타깃, `.PHONY`, `build:` 선행(M4) |
| [MODIFY] | `.github/workflows/ci.yml` | `go_code` 필터에 `.claude/settings.json` 추가와 근거 주석(M4) |
| [MODIFY] | `internal/config/toolpolicy/types.go` | `:94-98` 문서 주석 정정(M5) |
| [MODIFY] | `.moai/specs/SPEC-TOOLPOLICY-DRIFT-GUARD-001/progress.md` | §E.2/§E.3 는 run 단계 소유자가 채움 |

변경하지 않는 파일: `.claude/settings.json`, `internal/template/templates/.claude/settings.json.tmpl`, `internal/cli/tool_policy.go`, `internal/config/toolpolicy/codegen.go`, 종결된 SPEC-V3R6-TOOL-POLICY-SSOT-001 산출물. 변경 파일 중 템플릿 트리 아래에 있는 것이 없으므로 Template-First 미러 대상이 없다.

## §F 위험

1. **워킹 트리 판독의 부작용.** 런타임이 tracked settings.json 권한 블록을 미커밋으로 바꾸면 로컬 `make build` 가 붉어진다. 오케스트레이터 측정상 현재 primary 체크아웃의 미커밋 수정은 권한 블록 밖에만 있다(`PERM-SAME`) — 잠복 위험이다.
2. **새 항목의 `risk_tier` 판단.** 검사는 명세자와 결정만 본다. 7개 항목의 위험 등급과 감사 문구는 검증되지 않는다.
3. **`moai update` 뒤 붉은색.** update 는 settings.json 을 템플릿에서 다시 렌더한다. GitMode 에 따라 `git commit` / `git tag` / `git push` allow 가 빠지면 검사가 붉어진다. 올바른 신호이지만 놀랄 수 있다. CLAUDE.local.md §2.3 의 update 후 복원 절차가 이미 있다.
4. **Grep 경로 deny 의 문서 공백.** 공식 문서의 무시 목록에 Grep 은 이름이 없다(판정서 §4 갭 1). 적용 중인 settings.json 에는 이미 없으므로 적용 권한은 어느 쪽이든 바뀌지 않는다.
5. **집합 비교가 가리는 것.** 순서와 서식은 의도적으로 보지 않는다. 중복·겹침은 별도 단정으로 막는다.

## §G 금지 사항

- 실제 트리 대상 `moai tool-policy build` 실행 금지(스크래치 사본만).
- 실행 중 세션이 읽는 `.claude/settings.json` 에 대한 뮤테이션 금지.
- 검사 안의 `t.Skip` 금지.
- 산출물에 역슬래시-u 이스케이프 표기 금지.
