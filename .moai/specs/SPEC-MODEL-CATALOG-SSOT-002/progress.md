# SPEC-MODEL-CATALOG-SSOT-002 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-20
tier: M
artifacts: spec.md, plan.md, acceptance.md, progress.md
card: t1030
worktree: .claude/worktrees/t1030 (branch WT-model-matrix-rewrite)
base_tree: 9bd329c17 (= local develop at authoring time)
spec_id_regex_check: PASS (verbatim, Bash-run at plan phase — `[[ "SPEC-MODEL-CATALOG-SSOT-002" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → PASS)
id_uniqueness: no existing `.moai/specs/SPEC-MODEL-CATALOG-SSOT-002` directory at authoring time (`ls -d` → No such file or directory; positive control: `ls -d .moai/specs/SPEC-MODEL-*` → 4 sibling directories)
plan_audit: not yet run

### 무엇을 저작했는가

`SPEC-MODEL-CATALOG-SSOT-001`(카드 t842, `status: draft`, 미푸시 브랜치 `WT-model-catalog-ssot`)의 **축소 재작성**. 반복이 아니라 재작성인 이유는 수렴 4축 중 2축의 대상이 트리에서 소멸했기 때문이다 — 종이 위에서 남은 결함을 옳게 고쳐도 대상 없는 계획이 남는다.

- 요구사항 **10건** (REQ-MCS2-001..010). 전임 12건 중 2건 폐기, 1건(REQ-008) 겹침 제외, 8건 승계·축소, 1건(REQ-MCS2-010 시드 단일성) 신설.
- 인수 기준 **11건** (AC-MCS2-001..011). 전임 15건에서 **하나도 상속하지 않았다** — 전임 AC 9건이 공허(완전 5 + 부분 4)했고, 승계 경로를 끊는 것이 의도다.
- 마일스톤 **4개** (M1 스키마·로더 → M2 alias → M3 윈도우 → M4 가드), 결정 가역성 순.
- `design.md` / `research.md`는 저작하지 않았다 — Tier M이며, 스키마가 전임 설계의 축소판이라 별도 설계 탐색 없이 결정이 선다.

### 이 트리에서 직접 측정한 것 (카드 t1030, HEAD `9bd329c17`)

모든 부재 주장은 같은 명령 형태의 양성 대조와 짝지었다.

| 측정 | 결과 | 양성 대조 |
|---|---|---|
| 소멸 경로 (`internal/gateway` 외 3) | `git ls-tree` 0행 | 생존 8파일 → 8행 |
| 소멸 심볼 (`NewSessionCatalog` 외 6) | `git grep -c` 0행 | 생존 4심볼 → 20행 |
| 층 A census — raw | 90행 / **17파일** | — |
| 층 A census — 비주석 코드 | 30행 / **6파일** | — |
| MATRIX-002 겹침 — 층 A 6파일 | 6개 산출물 전부 **0회** 언급 | `agentfm.go` → 22회 (6산출물) |

census 두 층 분해는 카드 t1030의 4번째 독립 측정이며, 앞 세 측정(09-18 `2ede711ca`, 09-20 `a580945e9`, 09-20 `17633e1fb`)의 90/30/6과 파일 이름까지 일치한다. **「6파일」은 비주석 리터럴 보유 파일 수**이고, 주석에만 모델 id를 담는 11파일이 따로 있다 — 이 구분이 REQ-MCS2-007 가드의 축과 열거 정책을 지배한다.

증거 파일: `.moai/reports/t1030/` — `census-README.md` · `census-raw.txt`(90) · `census-code.txt`(30) · `census-files.txt`(17) · `census-files-code.txt`(6) · `overlap-matrix002.txt` · `req-survival.txt`.

### 승계 수치의 정정

인계된 「전임 AC 15건 중 6건 영향」은 **과소계상**이다. 카드 t1024의 개별 심볼 확인 결과는 완전 공허 5 + 부분 공허 4 = 9건이며, 선행 판정서의 awk 패턴에 `gptModelEffortAllowed`(AC-006)와 `receipt_history.go`(AC-014)가 빠져 있던 것이 원인이다. 본 SPEC은 6도 9도 상속하지 않고 AC를 새로 썼으므로 이 정정은 기록 목적이다.

### 검증하지 못한 것 (Gaps)

- **`go test` / `golangci-lint` 미실행.** plan-phase이고 코드 변경이 0이다. `plan.md` §C Pre-flight의 베이스라인 녹색 확인은 run-phase 착수 시 실행한다.
- **`ModelAliasTable` 소비자 3곳의 기존 테스트 존재를 재지 않았다.** fan_in ≥ 3(`launcher.go` `expandModelString` · `profile_setup.go` `normalizeModel` · `settings/schema.go` `modelOptions`)은 `model_policy.go:74-75`의 `@MX:ANCHOR`/`@MX:REASON` 주석에서 읽은 것이지 호출 그래프를 내가 추적한 것이 아니다. Pre-flight 항목 2.
- **패키지 의존 방향을 확정하지 않았다.** REQ-MCS2-010은 `internal/config`의 `Defaults()`가 시드 상수를 참조하라고 요구하는데, Claude 계열 id 상수(`ModelIDOpus5` / `ModelIDOpus48`)는 `internal/template`에 있다. `internal/config` → `internal/template` 참조가 허용되는지(순환 여부)는 재지 않았다. M1이 방향을 **한 쪽으로만** 정하도록 `plan.md` M1-2에 분기를 적어 뒀으나, 어느 분기가 성립하는지는 미측정이다.
- **새 섹션 추가가 건드리는 감사 테스트의 전량을 열거하지 않았다.** `audit_registry.go`와 `types.go:1125`는 읽어서 확인했으나, `audit_loader_wiring_test.go` · `audit_struct_yaml_symmetry_test.go` · `audit_loader_completeness_test.go`가 각각 무엇을 추가로 요구하는지는 읽지 않았다. Pre-flight 항목 4.
- **단가 값의 출처를 재확인하지 않았다.** 트리 내 선행 근거는 주석 1행(`model_policy.go:47-48`, opus-5 $5/$25)뿐이고 나머지는 전임 SPEC의 dispatch 전달값이다. 본 카드에 단가 소비자가 없으므로 어떤 판정에도 영향이 없으나, 값 자체는 미검증이다.
- **`internal/template/templates/` 하위의 모델 id 리터럴을 재지 않았다.** 층 A census는 `internal/` 전체를 `--include='*.go'`로 겨눴으므로 템플릿 YAML·MD는 범위 밖이다. 전임 research가 템플릿 트리 29행(비주석 20)을 기록했으나 본 트리에서 재측정하지 않았다.
- **MATRIX-002 겹침 측정은 산출물 텍스트 grep이다.** 그 SPEC이 실제로 `agentfm.go`를 어떻게 바꿀지는 코드가 아직 없어 알 수 없고, M3·M5 "Not started"는 그 SPEC의 `progress.md` 자기 보고를 읽은 것이지 내가 커밋 그래프로 확인한 것이 아니다.

### 잔여 위험

- `internal/cli/glm.go:310`이 1M 윈도우 집합 `{glm-5.3, glm-5.3-flash}`를 문장으로 재진술한다. 카탈로그에서 그 집합이 바뀌면 이 문장이 조용히 거짓이 된다. 본 SPEC은 Out of Scope로 두고 드러내기만 한다.
- REQ-MCS2-007 가드의 전수 스캔 + 정확 일치 정책은 **새 파일이 정당하게 모델 id를 들 때도 실패한다.** 그것이 의도다(선언 없이는 들 수 없다). 다만 가드가 마찰로 인식되어 예외 집합에 파일이 무심코 추가되면 가드는 서서히 공허해진다 — 예외 추가는 근거 1행을 동반해야 한다는 규율이 테스트 주석으로 고정되어야 한다.
- 「전제가 122커밋 뒤에도 불변」은 원 census 정규식이 담은 모델명 집합 안에서만 성립한다. develop에 새 모델 id 체계가 들어오면 이 측정은 보지 못한다.

## §E.2 Run-phase Evidence

_<pending run-phase — owned by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
