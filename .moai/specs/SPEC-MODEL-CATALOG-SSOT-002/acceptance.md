# SPEC-MODEL-CATALOG-SSOT-002 acceptance.md — 인수 기준

작성: 2026-09-20 (manager-spec, 카드 t1030).

## 이 AC 집합의 작성 규율

전임 `SPEC-MODEL-CATALOG-SSOT-001`의 AC 15건 중 9건이 공허하다 — 완전 공허 5(AC-005, AC-006, AC-013, AC-014, AC-015) + 부분 공허 4(AC-004, AC-008, AC-009, AC-012), 카드 t1024 측정. 인계된 「6/15」는 과소계상이었고, 과소계상은 피해를 작게 보이게 해 「조금만 고치면 된다」를 그럴듯하게 만든다.

그래서 **전임 AC를 하나도 상속하지 않았다.** 아래 11건은 전부 새로 썼고, 각 AC가 이름 댄 심볼·경로는 **본 트리 HEAD `9bd329c17`에서 존재를 확인한 것만** 쓴다. 각 AC는 이 트리에서 실행 가능한 명령을 하나 이상 든다.

## §D AC Matrix

| AC | 요구 | 마일스톤 | 검증 축 |
|---|---|---|---|
| AC-MCS2-001 | REQ-MCS2-001 | M1 | 파일·미러 존재 + 파싱 |
| AC-MCS2-002 | REQ-MCS2-002 | M1 | 로더 관례 + 감사 레지스트리 등록 |
| AC-MCS2-003 | REQ-MCS2-006 | M1 | 신뢰 경계 검증 fixture 7종 |
| AC-MCS2-004 | REQ-MCS2-008 | M1 | 파일 부재 폴백 + fail-open |
| AC-MCS2-005 | REQ-MCS2-009 | M1 | provenance 표기 + 템플릿 중립성 |
| AC-MCS2-006 | REQ-MCS2-010 | M1 | 시드 단일성 |
| AC-MCS2-007 | REQ-MCS2-004 | M2 | alias 쌍방향 정규화 |
| AC-MCS2-008 | REQ-MCS2-005 | M3 | statusline 윈도우 우선순위 |
| AC-MCS2-009 | REQ-MCS2-003 | M2·M3 | 행위 보존 characterization |
| AC-MCS2-010 | REQ-MCS2-007 | M4 | 리터럴 가드 — 축 + 열거 정책 |
| AC-MCS2-011 | §D Success Criteria | M4 | census 두 층 재측정 |

## 개별 시나리오

### AC-MCS2-001 — 카탈로그 파일과 미러 존재·로드

- **Given** `.moai/config/sections/model_catalog.yaml`와 `internal/template/templates/.moai/config/sections/model_catalog.yaml`가 존재하고
- **When** `config.NewLoader()`로 프로젝트를 로드하면
- **Then** `ModelCatalogConfig`가 현행 모델 집합(glm 9종 — `glm-5.3-flash`/`glm-5.3`/`glm-5.2`/`glm-5.1`/`glm-5`/`glm-4.7`/`glm-4.6`/`glm-4.5`/`glm-4.5-air`, claude alias 대상 4종 + 폐기 id 4종)과 함께 반환되고, 미러와 `Defaults()`의 모델 데이터가 동일하다. 스키마에 `auth` · `family` · `compat_alternate` · `gateway_default` · `native_policy` · `effort_allowlist` 필드가 **없다**.
- **검증**:
  ```bash
  test -f .moai/config/sections/model_catalog.yaml
  test -f internal/template/templates/.moai/config/sections/model_catalog.yaml
  go test ./internal/config/ -run 'TestModelCatalog' -v -count=1
  # 제거된 필드의 부재 — 양성 대조를 같은 명령 형태로 짝짓는다
  grep -cE '^\s*(auth|family|compat_alternate|gateway_default|native_policy|effort_allowlist):' \
    internal/template/templates/.moai/config/sections/model_catalog.yaml   # → 0
  grep -cE '^\s*(id|provider|context_window):' \
    internal/template/templates/.moai/config/sections/model_catalog.yaml   # → 0 초과 (양성 대조)
  ```

### AC-MCS2-002 — 로더 관례 정합 + 감사 레지스트리 등록

- **Given** `internal/config/CLAUDE.md`의 섹션 관례 기준표를 적용하고
- **When** 구현을 점검하면
- **Then** (a) 타입이 지정된 구조체, (b) `Defaults()` 팩토리, (c) unmarshal 직후 `Validate()` 호출, (d) 비엄격 YAML(미지 키 무시, 타입 불일치만 보고) 네 항목이 코드로 확인되고, (e) `internal/config/audit_registry.go`의 `yamlToStructRegistry`에 `"model_catalog"` 행이, (f) `internal/config/types.go` `sectionNames`에 섹션 이름이 등재되어 있다.
- **검증**:
  ```bash
  grep -n '"model_catalog"' internal/config/audit_registry.go internal/config/types.go
  go test ./internal/config/ -run 'TestAudit' -v -count=1   # 기존 파리티 감사 전부 녹색
  ```

### AC-MCS2-003 — 신뢰 경계 검증 (fixture 7종)

- **Given** 카탈로그 파일에 결함을 하나씩만 넣은 7개 fixture — (1) 미지 provider, (2) 중복 모델 id, (3) 빈 id, (4) 0 이하 context window, (5) 폐쇄집합 밖 provenance, (6) 두 엔트리에 중복 등장하는 alias, (7) 빈 `models:` 리스트
- **When** 각각을 로드하면
- **Then** 7건 모두 로드가 실패하고, 각 오류 메시지가 해당 entry id와 필드명을 포함한다. 결함 클래스가 fixture별로 격리되어 하나의 검증이 다른 것을 가리지 않는다.
- **검증**: `go test ./internal/config/ -run 'TestModelCatalogValidate' -v -count=1` — 7개 하위 테스트가 개별 이름으로 나타난다.

### AC-MCS2-004 — 파일 부재 폴백 + statusline fail-open

- **Given** `.moai/config/sections/model_catalog.yaml`가 없는 프로젝트, 그리고 파일은 있으나 파싱 불가한 프로젝트 두 fixture를 두고
- **When** 로드와 `statusline.ResolveGLMContextWindow`를 실행하면
- **Then** 두 경우 모두 컴파일 내장 기본값이 적용되고, 어떤 경로에서도 패닉·크래시가 없으며, statusline은 수렴 전과 동일한 윈도우 값을 반환한다.
- **검증**: `go test ./internal/config/ ./internal/statusline/ -run 'Fallback|FailOpen' -v -count=1`

### AC-MCS2-005 — 근거 표기 + 템플릿 중립성

- **Given** 템플릿 카탈로그 사본을 읽고
- **When** 단가를 가진 모든 엔트리를 점검하면
- **Then** `provenance` 필드가 존재하고 근사값은 `approximate`, 단위 미확인은 `unit: unconfirmed`로 표기되며, 사본 어디에도 내부 SPEC id · 카드 id · 내부 날짜 · 기계 종속 경로가 없다.
- **검증**:
  ```bash
  # 중립성 — 0이어야 한다
  grep -cE 'SPEC-[A-Z]|REQ-[A-Z]|\bt[0-9]{2,4}\b|2026-[0-9]{2}-[0-9]{2}|/Users/' \
    internal/template/templates/.moai/config/sections/model_catalog.yaml     # → 0
  # 양성 대조 — 같은 파일에서 비어 있지 않은 것을 확인해 프로브 생존을 보인다
  grep -c 'provenance:' internal/template/templates/.moai/config/sections/model_catalog.yaml  # → 0 초과
  ```

### AC-MCS2-006 — 시드 단일성

- **Given** `internal/config/defaults.go:199-200,:215-221`의 GLM `Default*` 상수 9개가 잔류하고
- **When** 카탈로그 `Defaults()` 팩토리의 소스를 읽으면
- **Then** 팩토리는 그 상수들을 **참조**하며 모델 id 문자열 리터럴을 하나도 담지 않는다. 즉 Go 컴파일 타임 시드는 여전히 정확히 한 곳이다.
- **검증**:
  ```bash
  # Defaults() 팩토리 파일의 모델 id 리터럴 — 0이어야 한다
  grep -nE '"(glm-|claude-)[a-z0-9.-]+"' internal/config/<defaults 팩토리 파일>.go   # → 0행
  # 양성 대조: 시드 상수는 그대로 있다
  grep -cE '"(glm-)[a-z0-9.-]+"' internal/config/defaults.go                          # → 9
  # 참조 관계가 실재함
  grep -n 'DefaultGLM' internal/config/<defaults 팩토리 파일>.go                       # → 비어 있지 않음
  ```
  (`<defaults 팩토리 파일>`의 실제 이름은 run-phase에서 확정되며, 판정 명령에 그 이름을 박아 보고한다.)

### AC-MCS2-007 — alias 쌍방향 정규화 (수렴 전후 동일)

- **Given** 기본 카탈로그가 로드된 상태에서
- **When** `ModelAliasCanonicalID` / `ModelAliasFromCanonicalID` / 폐기 id 정규화를 실행하면
- **Then** 다음이 수렴 전과 동일하다:

  | 입력 | 기대 출력 |
  |---|---|
  | `ModelAliasCanonicalID("opus")` | `claude-opus-5` |
  | `ModelAliasCanonicalID("sonnet")` | `claude-sonnet-5` |
  | `ModelAliasCanonicalID("fable")` | `claude-fable-5` |
  | `ModelAliasCanonicalID("haiku")` | `claude-haiku-4-5` |
  | `ModelAliasCanonicalID("opusplan")` | `opusplan` (자기매핑) |
  | `ModelAliasCanonicalID("<미등록>")` | 입력 그대로 반환 |
  | `ModelAliasFromCanonicalID("claude-opus-5")` | `opus` |
  | 폐기 정규화 `claude-opus-4-6` / `4-7` / `4-8` | `opus` |
  | 폐기 정규화 `claude-sonnet-4-6` | `sonnet` |

- **검증**: `go test ./internal/template/ -run 'TestModelAlias' -v -count=1` (M2 RED로 선행 커밋된 테스트를 **무편집으로** 재실행)

### AC-MCS2-008 — statusline 윈도우 우선순위 보존

- **Given** (a) `llm.yaml`에 `glm.context_windows`가 부재/비어 있는 fixture, (b) 1건 이상 존재하는 fixture를 두고
- **When** `statusline.ResolveGLMContextWindow`를 실행하면
- **Then** (a)는 카탈로그 빌트인 값을 반환한다 — `glm-5.3-flash`/`glm-5.3`/`glm-5.2` = 1,000,000 · `glm-5.1` = 200,000 · `glm-5`/`glm-4.7`/`glm-4.6`/`glm-4.5`/`glm-4.5-air` = 128,000 — 그리고 (b)는 사용자 값으로 **전면 대체**한다(병합이 아니다). 추가로 미등록 `glm-5.3-<변종>`은 1,000,000을 부분문자열로 상속하고, `glm-4.5`가 `glm-4.5-air`를 마스킹하지 않으며, 카탈로그 손상 시 수렴 전과 동일하게 fail-open한다.
- **검증**: `go test ./internal/statusline/ -run 'TestResolveGLMContextWindow|TestContextWindow' -v -count=1` — 기존 `memory_test.go`의 테스트가 무편집으로 통과하는 것이 필수 조건이다.

### AC-MCS2-009 — 행위 보존 (characterization)

- **Given** M2·M3 착수 **전에** 커밋된 characterization 테스트가 있고 (RED 커밋이 GREEN 커밋보다 git 이력에서 **앞선다**)
- **When** 수렴 커밋 이후 동일 테스트를 실행하면
- **Then** 모든 characterization 테스트가 **무편집으로** 통과한다. 테스트 파일이 GREEN 단계에서 수정되었다면 이 AC는 FAIL이다.
- **검증**:
  ```bash
  # 순서는 커밋 그래프가 증인이다 — 커밋 메시지나 세션 기록이 아니다
  git log --format='%h %s' -- internal/template/model_policy_alias_characterization_test.go
  git merge-base --is-ancestor <RED 커밋> <GREEN 커밋>; echo $?      # → 0
  # 그리고 같은 명령이 RED==GREEN 인 경우 exit 0 을 내므로, 두 SHA 가 다름을 함께 단정한다
  test "<RED 커밋>" != "<GREEN 커밋>"; echo $?                        # → 0
  git diff <RED 커밋>..<GREEN 커밋> -- '*characterization_test.go'    # → 빈 출력
  ```
  (`--is-ancestor X X`는 exit 0을 내므로 동일 SHA 비교가 이 기준을 공허하게 통과한다. 그래서 「다름」을 별도 행으로 단정한다.)

### AC-MCS2-010 — 리터럴 가드: 축과 열거 정책

- **Given** REQ-MCS2-007의 가드 테스트가 구현되어 있고
- **When** 테스트를 실행하면
- **Then** 세 성질이 모두 단정된다:
  1. **축** — 가드가 비주석 필터를 적용한다. 필터를 끈 대조 실행이 **17파일**을, 켠 실행이 **4파일**을 보고하며, 두 수가 가드 자신의 fixture로 고정된다.
  2. **열거** — 전수 스캔으로 얻은 비주석 리터럴 보유 파일 집합이 선언된 예외 집합 `{internal/config/defaults.go, internal/cli/glm.go, internal/statusline/usage.go, internal/profile/preferences.go}`와 **정확히 같다**(부분집합이 아니다).
  3. **재유입 검출** — 가드 fixture에 7번째 파일을 모사해 넣으면 가드가 실패한다(변이 테스트). 이 항목이 없으면 가드가 무엇이든 통과시키는지 알 수 없다.
- **검증**:
  ```bash
  go test ./internal/... -run 'TestModelLiteralGuard' -v -count=1
  # 수렴 대상 2파일의 리터럴 0행 — 각각 양성 대조를 짝짓는다
  grep -cE '"(glm-|claude-)[a-z0-9.-]+"' internal/statusline/memory.go      # → 0
  grep -cE '"(glm-|claude-)[a-z0-9.-]+"' internal/template/model_policy.go  # → 0
  grep -cE '"(glm-)[a-z0-9.-]+"' internal/config/defaults.go                # → 9  (양성 대조)
  ```

### AC-MCS2-011 — census 두 층 재측정

- **Given** `plan.md` §E의 두 층 census 명령과, 착수 전에 `.moai/reports/t1030/`에 보존된 베이스라인(raw 90행/17파일 · 코드 30행/6파일 — 카드 t1030이 본 트리 HEAD `9bd329c17`에서 측정)을 두고
- **When** 수렴 완료 후 같은 정규식으로 재측정하면
- **Then** **코드 층**이 30행 → **13행**, 6파일 → **4파일**이고 그 4파일이 AC-MCS2-010의 선언된 예외 집합과 일치한다. **raw 층**은 감소 의무가 없으며, raw 층 감소를 성공 지표로 인용하지 않는다.
- **Then (잔류 13행의 산술 내역)**: `defaults.go` 9 + `glm.go` 2 + `usage.go` 1 + `preferences.go` 1 = 13. 실측이 13과 다르면 차이가 어느 파일에서 왔는지를 보고한다 — 합계만 맞고 분포가 다르면 그것도 결함이다.
- **검증**: `plan.md` §E의 명령 2쌍을 verbatim 실행하고 출력 전문을 인용한다.

## 엣지 케이스

- 카탈로그에 미지 모델 id를 추가 → 로드는 성공하고, 그 id를 아는 리더가 없으므로 동작 변화 0
- `llm.yaml` 오버라이드에 카탈로그에 없는 모델 → 현행과 동일하게 오버라이드 테이블에만 존재하는 것으로 처리
- 카탈로그 YAML에 미지 최상위 키 → 비엄격 언마셜이 무시 (관례상 실패하지 않음)
- 카탈로그 YAML의 타입 불일치(윈도우 자리에 문자열) → `ConfigTypeError` 보고
- `glm-5.3-<신규변종>` 등록 없이 출현 → `glm-5.3` 부분문자열로 1M 상속 (`memory.go:28-31`의 등록 시점 규약 보존)

## 품질 게이트

- 새/변경 패키지 `go test` 통과 + 새 로더 커버리지 **85% 이상**
- `go vet ./internal/...` 0 오류, `golangci-lint run` 변경 패키지 0 이슈
- `make build` 성공 (템플릿 임베드 갱신)
- 기존 `TestAuditParity` 계열 감사 테스트 녹색

## Definition of Done

- AC-MCS2-001..011 전부 PASS 증거 확보 — 각 AC마다 실행한 명령과 **출력 전문**을 인용한다
- 부재를 주장하는 명령마다 같은 명령 형태의 양성 대조가 짝지어져 있다. 짝 없는 무출력은 Gap으로 보고한다
- census 두 층 재측정 보고 (코드 층 30 → 13 / 6파일 → 4파일, raw 층은 참고 수치로 병기)
- 잔여 위험 문서화: (a) `glm.go:310`의 1M 집합 재진술, (b) 차트 유래 단가의 단위 미확인, (c) `SPEC-MODEL-PROFILE-MATRIX-002` M3·M5 착수 시 `agentfm.go` 축 재검토 필요성
