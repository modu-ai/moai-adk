# t696 Verdict — CLI --json 출력 모양 문서 보강 (todo list / model profile)

Date: 2026-09-13 · Lane: lane-1 · Branch: `WT-json-shape-docs` (develop `f3d94ddd0` 기점) · Class B (run→sync)

## Claim

`moai todo list --json`과 `moai model profile --json`의 최상위는 **객체**인데 소비 에이전트가 배열로 짐작해 jq exit 5로 실패하던 사례(운영자/에이전트 실측)에 대해, ① 세 문서 표면에 실제 출력 예시와 올바른 필터를 보강하고 ② 문서 예시의 최상위 키와 실제 출력을 대조하는 가드 테스트를 붙여 재발을 잡는다. Template-First(템플릿 사본 먼저 편집, 로컬 사본 동일 내용 미러).

## Changed files

| 파일(템플릿 + 로컬 쌍) | 변경 |
|---|---|
| `internal/template/templates/.claude/rules/moai/development/model-policy.md` (+`.claude/rules/...` 로컬) | `model profile --json` 출력 예시(`{profile, backend, agents}`) + 올바른/잘못된 jq 필터 쌍 추가 |
| `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` (+로컬) | Per-Spawn Model Injection 불릿에 "단일 JSON 객체, `agents` 배열 하의 `{model, effort}`" 명시 + model-policy 포인터 |
| `internal/template/templates/.claude/skills/moai/workflows/todo.md` (+로컬) | 예시에 `project_uuid`·`archived` 보강(실제 출력과 정합) + "최상위는 객체, 배열은 `.items`" 사용 지침 + 필터 쌍 |
| `internal/cli/doc_json_shape_test.go` (신규) | 문서↔출력 형태 대조 가드 2건 |

## Evidence

| # | Command / 관측 | Output |
|---|---|---|
| E1 | 실제 출력 최상위 키 관측 | `todo list --json` → `dict` `['project_uuid','version','last_seq','items','findings','archived']` · `model profile --json` → `dict` `['profile','backend','agents']` — 둘 다 **객체**, 배열은 named key 아래 |
| E2 | jq 오류 실측: `moai model profile --json \| jq '.[0]'` | `jq: error: Cannot index object with number` · **exit 5** — 이슈 증상 그대로 재현 |
| E3 | 신규 가드 2건 실행 | `TestTodoListJSONShapeMatchesDoc` · `TestModelProfileJSONShapeMatchesDoc` PASS — 문서 예시 키 ⊆ 실제 출력 키 (템플릿+로컬 양 사본) |
| E4 | 수리 전 트리에서의 가드 상태 | 구형 `model-policy.md`의 유일한 ```json 펜스는 객체가 아님(수정 전 테스트 실행에서 "invalid character ':' after top-level value"로 관측) → 구형 트리에서 본 가드는 "no ```json object example"으로 **RED** 성립. `git show HEAD:…model-policy.md \| grep -c '```json'` → 1(비객체 펜스) |
| E5 | `go test ./internal/cli/ -run 'TestTodoListJSONShapeMatchesDoc\|TestModelProfileJSONShapeMatchesDoc\|TestResolveModelProfileReport' -count=1` | ok · gofmt/vet 클린 |

## Gaps

- E5는 선택자 범위 — `internal/cli` 패키지 전량은 CI 몫.
- 가드는 "문서 예시 키 ⊆ 실제 출력 키"의 단방향 대조다 — 출력이 문서에 없는 키를 **추가**하는 것은 잡지 못한다(상향 호환이므로 소비자 실패를 만들지 않는 축으로 판정).
- 문서 3곳 외에 `todo.md`의 `--json` 언급 2곳(§228 등)은 형태 서술이 없어 보강 대상에서 제외했다(출력 예시가 있는 §110만 보강).
- `model profile --json`의 GLM 백엔드 분기(`glm_model` 등 추가 필드)는 문서 예시에 미기재 — Claude 백엔드 기준 예시이며 GLM 필드 추가는 상향 호환.

## Residual risk

- 가드가 문서의 "첫 번째 객체 펜스"만 읽는다 — 미래에 예시가 여러 개가 되고 두 번째 예시가 드리프트하면 잡히지 않는다.
- 본 수리는 문서+가드이므로, 이미 굳어진 에이전트의 잘못된 짐작은 문서를 다시 읽어야 교정된다 — 배포(릴리스)가 사용자 프로젝트에 전달하는 행위다.

🗿 MoAI
