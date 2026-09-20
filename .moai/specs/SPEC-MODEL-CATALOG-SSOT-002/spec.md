---
id: SPEC-MODEL-CATALOG-SSOT-002
title: "model_catalog.yaml 단일 원천화 — 현행 develop 기준 축소 재작성 (alias·컨텍스트 윈도우·시드 단일화)"
version: "0.1.0"
status: draft
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P2
phase: "v3.3.0 target"
module: "internal/config, internal/template, internal/statusline"
lifecycle: spec-anchored
tags: "model-catalog, ssot, alias, context-window, model-policy, t1030"
era: V3R6
tier: M
related_specs: [SPEC-MODEL-CATALOG-SSOT-001, SPEC-MODEL-PROFILE-MATRIX-002, SPEC-GLM-EFFORT-MAX-001]
---

# SPEC-MODEL-CATALOG-SSOT-002 — model_catalog.yaml 단일 원천화 (축소 재작성)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-20 | manager-spec | 초판. `SPEC-MODEL-CATALOG-SSOT-001`(카드 t842, `status: draft`, 미푸시 브랜치 `WT-model-catalog-ssot`)의 **축소 재작성**. 수렴 4축 중 2축(gateway catalog · translate effort allowlist)이 트리에서 소멸했고, 1축(web matrix)은 인접 in-progress SPEC과 겹쳐 제외됐다. 카드 t1030, 워크트리 `.claude/worktrees/t1030`, 기준 트리 HEAD `9bd329c17`. |

## §A Problem Statement

### A.1 무엇이 문제인가

모델 id 리터럴과 컨텍스트 윈도우 값이 Go 코드 여러 곳에 흩어져 있다. 값이 바뀔 때(모델 추가, 윈도우 확인, alias 승격) 여러 파일을 손봐야 하고, 하나라도 빠뜨리면 리더 간 불일치가 조용히 생긴다. 이 카드는 그 값들의 단일 원천(SSOT)을 `.moai/config/sections/model_catalog.yaml`에 두고, **현재 트리에 실재하는** 리더만 그 파일 위로 수렴시킨다.

### A.2 전임 SPEC과 무엇이 다른가 — 재작성의 근거

전임 `SPEC-MODEL-CATALOG-SSOT-001`은 수렴 축을 넷으로 잡았다. 그중 **둘이 트리에서 사라졌다.** 카드 t857-t865가 moai gpt gateway를 철회(`2d25a88eb feat(cli)!: withdraw the moai gpt gateway`, 병합 `f67d2193f`)하면서 `internal/gateway` 트리 전체와 `internal/cli/gateway_*.go`가 삭제됐다.

본 트리 HEAD `9bd329c17`에서 카드 t1030이 직접 측정했다 — **부재 주장마다 같은 명령 형태의 양성 대조를 짝지었다**:

```bash
$ git ls-tree -r --name-only HEAD -- internal/gateway internal/cli/gateway_prepare.go \
      internal/cli/gateway_product_binding.go internal/cli/gateway_session.go
# → 0행

$ git ls-tree -r --name-only HEAD -- internal/cli/glm.go internal/config/defaults.go \
      internal/config/profile.go internal/profile/preferences.go internal/statusline/memory.go \
      internal/statusline/usage.go internal/template/model_policy.go internal/web/agentfm.go
# → 8행 전부 존재                                    ← 양성 대조
```

심볼 축도 같은 방식으로 확인했다:

```bash
$ git grep -c -E 'NewSessionCatalog|gptEffortAllowlist|gatewayNativeModels|gatewayGPTModels|gptModelEffortAllowed|PolicyGPTNative|PolicyAnthropicNative' -- internal
# → 0행

$ git grep -c -E 'ModelAliasCanonicalID|glmContextWindows|agentFMModelValues|validOverrideModels' -- internal
# → 20행 (internal/config/profile.go:3 · internal/statusline/memory.go:8 ·
#          internal/template/model_policy.go:2 · internal/web/agentfm.go:4 외)   ← 양성 대조
```

두 축이 서로 독립이므로, 결론이 뒤집히려면 경로 축과 심볼 축이 **동시에** 틀려야 한다.

전임 SPEC은 `status: draft`인 채 미푸시 브랜치 `WT-model-catalog-ssot`에만 존재하며, 그 15개 AC 중 **완전 공허 5건 + 부분 공허 4건**(카드 t1024 측정)이다. 종이 위에서 남은 결함(NF-1)을 옳게 고쳐도 **대상이 없는 계획**이 남는다. 그래서 반복(iterate)이 아니라 재작성이다.

### A.3 살아남은 범위 — 층 A

원 census 정규식을 `internal/` 한정 · `_test.go` 제외로 겨눈 결과는 **네 번의 독립 측정에서 불변**이다:

| 값 | 09-18 `2ede711ca` | 09-20 `a580945e9` | 09-20 `17633e1fb` | **본 트리 `9bd329c17`** |
|---|---:|---:|---:|---:|
| 전체행(주석 포함) | 90 | 90 | 90 | **90** |
| 비주석 코드행 | 30 | 30 | 30 | **30** |
| 비주석 리터럴 보유 파일 | 6 | 6 | 6 | **6** |

앞 세 값은 카드 t842·t1024의 측정이고, 네 번째는 카드 t1030이 본 트리에서 직접 잰 것이다. 그 사이 develop은 122+ 커밋 움직였다 — 「며칠 지난 카드이니 전제가 낡았을 것」은 이 카드에서 실측으로 기각된다.

**비주석 리터럴을 보유한 6파일**: `internal/cli/glm.go` · `internal/config/defaults.go` · `internal/profile/preferences.go` · `internal/statusline/memory.go` · `internal/statusline/usage.go` · `internal/template/model_policy.go`.

### A.4 census는 두 층이다 — 가드 설계를 가르는 구분

「90 / 30 / 6」은 **두 층을 섞은 표기**다. 카드 t1030이 본 트리에서 층을 분해했다:

| 층 | 행 | 파일 |
|---|---:|---:|
| raw (주석 포함) | 90 | **17** |
| 비주석 코드행 | 30 | **6** |

즉 **「6파일」은 비주석 리터럴을 보유한 파일 수**다. 나머지 11파일은 모델 id를 **주석 안에만** 담는다: `internal/cli/launcher.go` · `internal/cli/model.go` · `internal/cli/profile_setup.go` · `internal/config/closed_sets.go` · `internal/config/types.go` · `internal/hook/session_start.go` · `internal/settings/schema_sections.go` · `internal/statusline/metrics.go` · `internal/statusline/types.go` · `internal/template/glm_effort_overlay.go` · `internal/web/agentfm.go`.

이 구분이 REQ-MCS2-007(리터럴 재유입 금지 가드)의 설계를 지배한다 — 자세한 것은 그 요구사항 본문에.

`internal/web/agentfm.go`가 이 **주석 전용** 11파일 군에 있다는 점을 따로 적어 둔다. 그 파일의 census 출현은 수렴 대상이라는 뜻이 아니며, 본 SPEC은 그 파일을 §Out of Scope에서 별도 근거로 제외한다.

### A.5 수렴 후 기대 상태 (산술)

| 파일 | 비주석 리터럴 행 | 처분 |
|---|---:|---|
| `internal/statusline/memory.go` (:34-42) | 9 | **수렴** → 카탈로그 |
| `internal/template/model_policy.go` (:51,:57,:79-81,:95-96,:98) | 8 | **수렴** → 카탈로그 |
| `internal/config/defaults.go` (:199-200,:215-221) | 9 | **잔류 — 유일 컴파일 시드** (REQ-MCS2-010) |
| `internal/cli/glm.go` (:110,:310) | 2 | **잔류 — 표시 문자열** (Out of Scope) |
| `internal/statusline/usage.go` (:32) | 1 | **잔류 — 선언된 가드 예외** (REQ-MCS2-007) |
| `internal/profile/preferences.go` (:26) | 1 | **잔류 — 계수 인공물** (아래) |
| 합계 | **30** | 수렴 17 / 잔류 13 |

`internal/profile/preferences.go:26`은 `Model string \`yaml:"model,omitempty"\` // e.g. "claude-opus-4-6", "claude-opus-4-7"` 이다 — 모델 리터럴은 **행 끝 주석 안**에 있고 코드는 어떤 모델 값도 운반하지 않는다. 행 선두가 `//`가 아니어서 비주석 필터를 통과할 뿐이다. 따라서 **「30 비주석 코드행」은 실제 값 운반 행을 정확히 1행 과다계상한다.** 이 사실을 여기 적어 두지 않으면 다음 사람이 29를 보고 결함으로 읽는다.

## §B Requirements (GEARS)

### REQ-MCS2-001 — 카탈로그 파일과 축소된 스키마 (Ubiquitous)

`model_catalog` 설정 섹션은 `.moai/config/sections/model_catalog.yaml`에, 템플릿 미러는 `internal/template/templates/.moai/config/sections/model_catalog.yaml`에 존재해야 하며(shall exist), 각 모델 엔트리마다 다음을 담아야 한다: 모델 id, provider, 컨텍스트 윈도우(토큰), 근거 표기를 동반한 단가, 사용자 대면 alias와 폐기 id 정규화, 표시 레이블.

스키마는 전임 SPEC 대비 **축소된다**. `auth` · `family` · `compat_alternate` · `gateway_default` · `native_policy` · `effort_allowlist` 필드는 담지 않는다 — 전부 소멸한 gateway/translate 축의 소비자 전용이며, 소비자 없는 필드를 스키마에 두는 것은 데이터가 존재하지 않는 리더를 흉내 내는 것이다.

### REQ-MCS2-002 — 섹션 로더 관례 준수 (Ubiquitous)

카탈로그 로더는 `internal/config/CLAUDE.md`의 섹션 파일 관례를 따라야 한다(shall follow): 타입이 지정된 `ModelCatalogConfig` 구조체, 섹션 매니저에 배선된 `Loader` 로드 경로, 배포 템플릿과 값이 일치하는 `Defaults()` 팩토리, 비엄격 YAML 언마셜(미지 키 무시, 타입 불일치만 보고), 그리고 언마셜 직후 호출되는 `Validate()` 훅.

추가로, 새 섹션은 **파리티 감사 레지스트리에 등록되어야 한다** — `internal/config/audit_registry.go`의 `yamlToStructRegistry`에 `"model_catalog": "ModelCatalogConfig"` 행을 더하고 `internal/config/types.go:1125` `sectionNames`에 섹션 이름을 더한다. 등록 없이 새 yaml 파일을 추가하면 기존 `TestAuditParity`가 실패한다.

### REQ-MCS2-003 — 행위 보존 (Ubiquitous)

배포 기본 카탈로그가 자리에 있는 동안(While), 수렴된 모든 리더는 수렴 이전 하드코딩 동작과 동일한 출력을 내야 하며(shall produce identical output), 그 동일성은 각 리더를 재배선하기 **전에** 커밋된 characterization 테스트로 검증되어야 한다. 현행 값이 벤더 문서와 어긋나더라도 현행 값을 그대로 인코딩하고, 어긋남은 in-file 노트로 표기만 한다 — 조용히 고치지 않는다.

### REQ-MCS2-004 — model policy alias 수렴 (Event-driven)

모델 alias가 어느 방향으로든 해석될 때(When) — `ModelAliasTable`, `ModelAliasCanonicalID`, `ModelAliasFromCanonicalID`, `ModelDeprecatedCanonicalIDs` — alias 매핑은 `internal/template/model_policy.go:79-81,:95-98`의 map 리터럴 대신 카탈로그 파일의 alias 데이터에서 공급되어야 하며(shall be sourced from the catalog), `opusplan` 자기매핑과 폐기 id 정규화(`claude-opus-4-6`/`4-7`/`4-8`→`opus`, `claude-sonnet-4-6`→`sonnet`)를 보존해야 한다.

`ModelAliasTable`은 `@MX:ANCHOR`(fan_in ≥ 3 — `launcher.go` `expandModelString`, `profile_setup.go` `normalizeModel`, `settings/schema.go` `modelOptions`)를 달고 있다. ANCHOR가 새 홈으로 이동하면 MX 규약에 따라 갱신하고 보고해야 한다.

### REQ-MCS2-005 — statusline 컨텍스트 윈도우 수렴 (Event-driven)

statusline이 모델의 컨텍스트 윈도우를 해석할 때(When), 빌트인 테이블은 `internal/statusline/memory.go:33-42`의 `glmContextWindows` map 리터럴 대신 카탈로그 파일에서 읽어야 한다(shall read from the catalog file). 현행 우선순위를 그대로 보존한다: `llm.yaml`의 `glm.context_windows`가 비어 있지 않으면 빌트인 테이블을 **전면 대체**하고(병합 아님 — `len>0` 조건), 그 다음 빌트인, 그리고 longest-substring 매칭이다.

`memory.go:28-31`이 명시한 등록 시점 규약도 데이터로 보존되어야 한다: 미등록 `glm-5.3-*` 변종은 `glm-5.3`의 부분문자열 매칭으로 1M을 상속하며, 명시적 `glm-5.3-flash` 엔트리가 바로 그 규칙의 발산 가드다.

전임 SPEC REQ-007의 나머지 절반(gateway capability 윈도우, `gateway_product_binding.go:26,:91`)은 대상이 부재하므로 본 요구사항에 없다.

### REQ-MCS2-006 — 신뢰 경계 검증 (Ubiquitous)

카탈로그 로더는 설정 신뢰 경계에서 검증해야 한다(shall validate). 다음 각각은 위반 엔트리와 필드명을 지목하는 검증 오류를 내야 한다: 미지 provider 값, 중복 모델 id, 빈 id, 0 이하 컨텍스트 윈도우, `{authoritative, approximate, unconfirmed}` 밖의 단가 근거(provenance), 여러 엔트리에 중복 등장하는 alias, 그리고 빈 `models:` 리스트.

`auth` enum 검증은 담지 않는다 — REQ-MCS2-001이 그 필드를 스키마에서 제거했다.

### REQ-MCS2-007 — 리터럴 재유입 금지 가드 (Ubiquitous)

가드 테스트는 수렴된 리더 파일에 모델 id 리터럴이 재유입되면 실패해야 한다(shall fail). 가드는 **축과 열거 정책을 명시적으로 선언해야 한다**:

- **축 — 비주석 코드행 한정.** 가드는 행 선두가 `//` 또는 `#`인 행을 제외한 뒤에만 판정한다. §A.4가 측정한 대로 전체 행을 훑는 가드는 **17파일**을 보며, 모델 id를 주석으로만 언급하는 11파일 전부를 위반으로 신고한다 — 아무도 수렴할 의도가 없는 주석에 대해 실패하는 과잉 가드다.
- **열거 정책 — 고정 목록이 아니라 트리 전수 스캔 + 선언된 예외 집합.** 가드를 6개 파일명에만 겨누면 반대 결함이 생긴다: 7번째 파일에 재유입된 리터럴을 **원리상 보지 못한다**. 그래서 가드는 `internal/`(`_test.go` 제외)을 전수 스캔해 비주석 리터럴 보유 파일 집합을 구하고, 그 집합이 아래 **선언된 예외 집합과 정확히 같은지**를 단정한다. 새 파일이 나타나면 집합 불일치로 실패한다.

수렴 완료 후 선언된 예외 집합(§A.5 산술과 동일, 4파일 / 13행):

| 파일 | 행 | 예외 근거 |
|---|---:|---|
| `internal/config/defaults.go` | 9 | 유일 컴파일 시드 (REQ-MCS2-010) |
| `internal/cli/glm.go` | 2 | 사용자 대면 help 문자열 (Out of Scope) |
| `internal/statusline/usage.go` | 1 | 날짜 스냅샷 id `claude-haiku-4-5-20251001` — 소비자 1인 축 (Out of Scope) |
| `internal/profile/preferences.go` | 1 | 행 끝 주석 계수 인공물 (§A.5) |

`_test.go` 파일의 리터럴은 자유롭다 — characterization fixture가 리터럴 없이는 성립하지 않는다.

### REQ-MCS2-008 — 파일 부재 시 폴백 (Event-driven)

카탈로그 파일이 부재하거나 파싱에 실패할 때(When), 로더는 배포 템플릿과 동일한 컴파일 내장 기본값으로 폴백해야 하며(shall fall back), statusline 윈도우 해석은 현행과 정확히 동일하게 degrade해야 한다(fail-open, statusline 크래시 없음). statusline은 매 렌더 실행되는 핫 패스이므로, 현행 `readLLMYAMLContextWindows`와 같은 등급의 파일 읽기 전략을 유지한다.

### REQ-MCS2-009 — 근거 표기 (Ubiquitous)

카탈로그 파일은 모든 근사값의 데이터 품질을 in-file로 표기해야 한다(shall mark): 차트 유래 단가는 `provenance: approximate`와 미확인 단위 표기를 달고, 벤더 문서와 어긋나면서 현행 코드 동작과 일치하는 윈도우 값은 SPEC을 읽지 않고도 어긋남이 보이도록 in-file 노트를 단다. 템플릿 사본은 중립을 유지한다 — 내부 SPEC id, 카드 id, 내부 날짜, 기계 종속 값 금지.

### REQ-MCS2-010 — 시드 단일성 (Ubiquitous)

`internal/config/defaults.go:199-200,:215-221`의 GLM `Default*` 상수 9개는 **유일한 Go 컴파일 타임 시드로 잔류해야 한다.** 카탈로그의 `Defaults()` 팩토리는 그 상수들을 참조해야 하며(shall reference), 모델 id 문자열을 다시 적어서는 안 된다(shall not restate) — `Defaults()`가 자기 리터럴을 들면 시드가 둘이 되고, 그것은 이 SPEC이 없애려는 결함을 이 SPEC의 산출물이 다시 싣는 것이다.

## §C Constraints

- **행위 보존이 최상위 제약이다.** REQ-MCS2-003이 다른 모든 요구사항을 지배한다. 카탈로그 기본값이 현행과 다르면 그것은 결함이다.
- **Template-First** (`CLAUDE.local.md` §2 [HARD]): 카탈로그 파일은 템플릿 미러와 함께 작성하고 `make build`로 임베드한다. 템플릿 중립성은 `.moai/docs/template-internal-isolation-doctrine.md` §25.1 C-class 기준을 따른다.
- **섹션 관례**: `internal/config/CLAUDE.md`의 Loader/Defaults/Validate/비엄격 YAML 패턴 + 파리티 감사 레지스트리 등록(REQ-MCS2-002). 새 env var는 필요하지 않을 것으로 보이며, 생기면 `internal/config/envkeys.go` 상수로만 등록한다.
- **범위 고정**: 이 카드는 본 트리에 **실재하는** 리더만 수렴한다. 트리에 없는 표면에 대한 요구사항은 쓰지 않는다 — 그것이 전임 SPEC이 폐기된 이유다.
- **TRUST 5**: Tested(새 로더 85%+ 커버리지), Secured(REQ-MCS2-006 신뢰 경계 검증), Trackable(Conventional Commit, 카드 id 포함).

## §D Success Criteria

- 배포 기본 카탈로그 로드 시 `model_policy` alias 해석과 statusline 윈도우 해석의 출력이 수렴 전과 동일함을 characterization 테스트로 증명
- 층 A census 재측정에서 비주석 코드행이 **30 → 13**, 비주석 리터럴 보유 파일이 **6 → 4**로 줄고, 그 4파일이 REQ-MCS2-007의 선언된 예외 집합과 **정확히 일치**
- raw 층(주석 포함)은 변하지 않아도 무방하다 — 주석은 수렴 대상이 아니므로 17파일/90행 근처 유지가 정상이며, raw 층 감소를 성공 지표로 읽지 않는다
- 새 로더 커버리지 85% 이상, `go vet` / `golangci-lint run` 0 오류
- 기존 `TestAuditParity` 계열 감사 테스트가 새 섹션 등록 후 녹색

## §E Out of Scope

### Out of Scope — 소멸한 gateway catalog 축 (전임 REQ-004)

- `NewSessionCatalog` · `gatewayNativeModels` · `gatewayGPTModels` rows 생성, provider별 기본 모델, native-policy 적격성 게이트, receipt family 귀속과 호환 alternate
- **근거**: 대상 전부가 본 트리에 부재하다. `git ls-tree -r --name-only HEAD -- internal/gateway internal/cli/gateway_prepare.go internal/cli/gateway_product_binding.go internal/cli/gateway_session.go` → 0행, 같은 명령 형태의 양성 대조(생존 8파일) → 8행. 심볼 축 `git grep -c -E 'NewSessionCatalog|gatewayNativeModels|gatewayGPTModels|PolicyGPTNative|PolicyAnthropicNative' -- internal` → 0행, 양성 대조 → 20행. 원인은 카드 t857-t865의 gpt gateway 철회(`2d25a88eb`, 병합 `f67d2193f`). 측정 귀속: 카드 t1030, 본 트리 HEAD `9bd329c17`.

### Out of Scope — 소멸한 translate effort allowlist 축 (전임 REQ-005)

- `gptEffortAllowlist` / `gptModelEffortAllowed` 수렴, astra `none` 거부 예외의 데이터화
- **근거**: `git grep -c -E 'gptEffortAllowlist|gptModelEffortAllowed' -- internal` → 0행 (위와 같은 양성 대조 짝). `internal/gateway/translate/` 디렉터리 자체가 부재하다.
- 따라서 전임 SPEC의 t848 잔여(astra `max` live 재측정)도 본 SPEC의 관심사가 아니다 — 고칠 허용집합이 트리에 없다.

### Out of Scope — web matrix 표시 축 (전임 REQ-008)

- `internal/web/agentfm.go` `agentFMModelValues()`, `agentModelCostRank`, 그리고 전임 REQ-008이 함께 묶었던 `internal/config/profile.go:120` `validOverrideModels`의 공개 접근자 재배선
- **근거 — 인접 SPEC과의 겹침 측정**: 카드 t1030이 `SPEC-MODEL-PROFILE-MATRIX-002`(`status: in-progress`, Tier L)의 6개 산출물 전부를 파일 수준으로 훑었다.
  - 층 A 생존 6파일은 MATRIX-002 산출물 어디에도 **0회** 등장한다 (`grep -c -E 'statusline/memory\.go|profile/preferences\.go|statusline/usage\.go|cli/glm\.go|config/defaults\.go|template/model_policy\.go'` → 6파일 전부 0).
  - 유일한 접촉점이 `internal/web/agentfm.go`다 — `plan.md` 6건 · `research.md` 7건 · `acceptance.md` 4건 · `spec.md` 3건 · `design.md` 2건.
  - 그중 결정적인 두 행: `plan.md:153`(**M5 — Web console cleanup**) "Remove `haiku` from the agentfm model selector option set" 은 전임 REQ-008이 수렴하려던 바로 그 집합을 **변경**하는 작업이고, `plan.md:128`(**M3 — Effort actualization**)은 `agentfm.go`를 네 seam 중 하나로 배선한다. `acceptance.md:231,236`이 그 파일을 AC 검증 대상으로 든다.
  - `progress.md` 기준 **M3·M5 둘 다 미착수**다("Not started"). 즉 아직 오지 않은 변경과 정면으로 겹친다.
- 두 카드가 같은 함수를 동시에 만지면 한쪽이 다른 쪽을 덮는다. 이 축은 **증거에 의해 제외된 것이지 누락이 아니다** — 제외 사유가 소멸하려면 MATRIX-002 M3·M5가 먼저 닫혀야 한다.
- `internal/web/agentfm.go`가 층 A census의 **주석 전용 11파일** 군에 있다는 사실(§A.4)은 이 제외와 정합한다 — 그 파일에는 수렴할 비주석 리터럴 행이 애초에 없다.

### Out of Scope — 표시용 문자열

- `internal/cli/glm.go:110`(help 본문 "glm-5.2 1M, glm-4.7 ~202K")과 `:310`(`Fprintln` "Main session context window: 1M (glm-5.3, glm-5.3-flash)")
- **근거**: 두 행 모두 사람이 읽는 문장이고 기계가 판정에 쓰는 값이 아니다. 같은 파일의 나머지 5건(`:5`, `:332`, `:336`, `:430`, `:930`)은 주석이다.
- **잔여 위험으로 기록한다**: `:310`은 1M 윈도우 집합 `{glm-5.3, glm-5.3-flash}`를 문장으로 재진술하므로, 카탈로그에서 그 집합이 바뀌면 이 문장이 조용히 거짓이 된다. 본 SPEC은 고치지 않고 드러내기만 한다 — 후속 카드 후보.

### Out of Scope — usage probe의 날짜 스냅샷 id

- `internal/statusline/usage.go:32` `haikuProbeModel = "claude-haiku-4-5-20251001"`
- **근거**: 이 값은 rate-limit 헤더 프로브 전용의 **날짜 접미 id**이며, 트리의 어떤 다른 표면도 이 형태를 담지 않는다(`ModelAliasTable`의 haiku는 `claude-haiku-4-5`로 접미 없음). 카탈로그에 날짜 스냅샷 축을 신설하면 소비자가 정확히 1인인 필드가 생긴다 — YAGNI다.
- 다만 **침묵하지는 않는다**: REQ-MCS2-007 가드의 **선언된 예외**로 등재되므로, 가드는 이 행을 보고도 통과하되 그 통과가 명시적 선언에 근거한다.

### Out of Scope — `moai update` 배포 의미론

- `moai update`의 섹션 파일 배포·보존 정책 변경. 이 카드는 템플릿 미러를 배포 가능한 상태로 두는 것까지만 책임진다. (`CLAUDE.local.md` §2.3이 관리 대상 뿌리 전량 삭제를 기술한다 — `.moai/config`가 그 뿌리에 있으므로 사용자 수정 보존은 별도 축이다.)

### Out of Scope — 단가를 소비하는 UI

- 카탈로그 `price` 블록은 자리와 검증만 만든다. 이 카드에 그 값을 소비하는 리더는 없다 — $/task 셀, 프리셋, availability 표시는 model-matrix redesign 후속 단계 소관이다.

## §F Cross-references

- `acceptance.md` — AC-MCS2-001..011 Given-When-Then
- `plan.md` — M1..M4 마일스톤
- `SPEC-MODEL-CATALOG-SSOT-001` — 전임 SPEC (`status: draft`, 미푸시 브랜치 `WT-model-catalog-ssot`). 본 SPEC이 축소 재작성한다
- `SPEC-MODEL-PROFILE-MATRIX-002` — 인접 in-progress Tier L. §E web matrix 제외의 근거
- `internal/config/CLAUDE.md` — 섹션 로더 관례
- `.moai/reports/t1030/` — census 두 층 분해(`census-README.md`, `census-raw.txt`, `census-code.txt`, `census-files.txt`, `census-files-code.txt`), MATRIX-002 겹침(`overlap-matrix002.txt`), REQ 생존표(`req-survival.txt`)
