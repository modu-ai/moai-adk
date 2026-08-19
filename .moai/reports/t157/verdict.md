# t157 — ANTHROPIC_DEFAULT_MODEL SSOT 정렬

베이스: 로컬 `release/v3.1.1` @ `18e0de51f` · 브랜치 `WT-anthropic-ssot` · Class B (plan 생략)
출처: `.moai/research/cc-update-2.1.235-to-2.1.236.md` GD-2 (gitignore 대상, 읽기만)

**성격: 결함 수정이 아니라 SSOT 를 상류에 맞추는 작업.** 런타임 결함은 없었고, 실패하던 테스트도 없었다.
아래 "비결함 재확인"은 그 사실을 제 손으로 다시 측정한 기록이다.

---

## 1. 전제 재확인 (리드 제시 4건 — 전부 실측 일치)

카드가 준 전제를 그대로 받지 않고 다시 쟀다. 네 건 모두 일치했다.

| 전제 | 확인 방법 | 결과 |
|---|---|---|
| `ANTHROPIC_DEFAULT_MODEL` 이 Go 소스에 0건 | `grep -rn --include='*.go' 'ANTHROPIC_DEFAULT_MODEL' internal/ pkg/ cmd/ \| wc -l` | `0` |
| `statusline/memory.go` 가 glob 이 아닌 열거 슬라이스 | 140-155행 직접 읽음 | `envSlots := []string{Opus, Sonnet, Haiku}` — 3키 열거 확인 |
| `statusline/metrics.go` 가 tier 별 switch | 205-220행 직접 읽음 | opus/sonnet/haiku 3분기 switch, glob 없음 |
| `EnvAnthropicPrefix` 에 프로덕션 소비자 없음 | `grep -rn 'EnvAnthropicPrefix' internal/ pkg/ cmd/` | 4건 전부 정의부(`envkeys.go` 2) + 테스트(`anthropic_env_ssot_test.go` 2) |

**비결함 재확인 결론**: 새 키가 tier 슬롯 변수로 오인될 경로가 없다. prefix 스윕이 프로덕션에 존재하지
않으므로 `ANTHROPIC_DEFAULT_MODEL` 이 export 돼도 슬롯 해석에 끼어들 수 없다. 고칠 결함이 없었다는
리드의 판단이 맞다.

## 2. 왜 그런데도 작업인가 (Tier 2 근거)

`anthropic_env_ssot_test.go` 의 파일 헤더가 스스로 계약을 선언한다 —
*"adding a constant without adding it here is a visible omission"*.
즉 이 패키지는 **ANTHROPIC_ 네임스페이스를 완전 열거한다**를 계약으로 삼고 있고, 상류에 새 이름이
생긴 순간 그 계약이 조용히 깨진 상태였다. 결함은 아니지만 SSOT 의 완전성 주장이 사실과 어긋난다.

## 3. 조치

### 3.1 `internal/config/envkeys.go` — 상수 추가

`EnvAnthropicDefaultModel = "ANTHROPIC_DEFAULT_MODEL"` 을 tier 4형제 **바로 앞**에 배치했다.
같은 블록 안이되 4형제와 붙여 두지 않은 이유는 축이 다르기 때문이다 — 주석에 그 구분을 적었다:
tier 변수는 alias→ID 해석이고, 이 변수는 세션 시작 모델이다.

주석에 관측 범위를 명시했다: MoAI 는 이 변수를 읽지도 쓰지도 않으며, 상수는 네임스페이스 완전 열거와
bare literal 방지를 위해 존재한다. 근거는 v2.1.236 changelog 이고 공식 문서에는 없다는 점도 적었다.

### 3.2 `internal/config/anthropic_env_ssot_test.go` — 파생 집합 갱신

- `bannedAnthropicEnvNames` 에 항목 추가 (알파벳 순 위치 유지)
- `TestAnthropicBannedSetCoversAllNames` 의 `wantLen` 9 → 10, 헤더 주석 문구도 9 → 10
- 런타임 커버리지 표에 `{"EnvAnthropicDefaultModel", EnvAnthropicDefaultModel}` 행 추가

세 곳을 모두 고쳐야 한다: `wantLen` 만 올리고 표를 빠뜨리면 크기 검사는 통과하는데 개별 존재 검사가
비게 되고, 표만 늘리면 크기 검사가 깨진다.

### 3.3 `.claude/rules/moai/development/model-policy.md` — 상호작용 서술

`## Default-Model Cost Lever` 절 끝(GLM-mode reconciliation 다음)에
`### Three model-selection env axes (they are not interchangeable)` 를 추가했다. 담은 내용:

- 세 축을 표로 분리 — `ANTHROPIC_MODEL`(세션 한정, 비영속) / `ANTHROPIC_DEFAULT_MODEL`(신규 세션 시작
  모델, `/model` 이 덮고 그 선택은 영속) / `ANTHROPIC_DEFAULT_<TIER>_MODEL`(alias→ID 해석)
- MoAI 가 쓰는 것은 **세 번째 축뿐**이라는 명시 (`setGLMEnv` in `glm.go`)
- 겹침 서술: 전역 export 된 `ANTHROPIC_DEFAULT_MODEL` 은 GLM 세션 포함 모든 세션의 시작 모델을 정하므로,
  슬롯 매핑과 별개로 둘이 동시에 작용한다. GLM 세션이 예상 밖 모델로 시작하면 슬롯 매핑을 의심하기 전에
  전역 export 를 먼저 확인하라는 실무 지침
- **관측 범위 + 추정 표시** (아래 §4)

### 3.4 Template-First

`model-policy.md` 는 템플릿 미러를 가진다. 편집 **전** `diff -q` 로 두 파일이 byte-identical 임을 확인한
뒤에만 `cp` 로 반영했다(중립화 차이가 있는 파일에 cp 를 쓰면 그 차이가 지워진다 — 여기서는 차이가 0이었다).
이후 `make build` 로 임베드 재생성. `catalog.yaml` 은 변동 없음(스킬 추가가 아니므로 정상).

## 4. 동작 단정 회피 — 무엇을 관측했고 무엇을 안 했는가

**직접 관측한 것** (보고서 인용이 아니라 이 세션에서 `WebFetch` 로 재확인):
`https://code.claude.com/docs/en/model-config` 전문에서 `ANTHROPIC_[A-Z_]*` 를 뽑아 세었다.

```
15 ANTHROPIC_DEFAULT_OPUS_MODEL     12 ANTHROPIC_MODEL
 8 ANTHROPIC_DEFAULT_SONNET_MODEL    7 ANTHROPIC_DEFAULT_FABLE_MODEL
 6 ANTHROPIC_DEFAULT_HAIKU_MODEL     5 ANTHROPIC_BASE_URL
 4 ANTHROPIC_CUSTOM_MODEL_OPTION     1 ANTHROPIC_SMALL_FAST_MODEL
 1 ANTHROPIC_DEFAULT_   ← 산문 속 `ANTHROPIC_DEFAULT_*_MODEL` 글롭 표기 조각(781행), 대상 이름 아님
```

→ **`ANTHROPIC_DEFAULT_MODEL` 은 공식 문서에 0건**. 유일한 근사 매치는 글롭 표기였다.
문서 지연이 사실임을 제 관측으로 확인했고, changelog 가 인용 가능한 유일 출처다.
같은 페이지 111행이 `ANTHROPIC_MODEL` 을 "launch 한 세션에만 적용"이라 적어, changelog 가 말한
대비(영속 여부)와도 어긋나지 않는다.

**관측하지 않은 것 (Gap — 단정하지 않았다)**:
- `ANTHROPIC_DEFAULT_MODEL` 과 tier 변수의 **우선순위를 측정하지 않았다.** 두 변수를 동시에 export 한
  세션을 실행하지 않았고, 그 순서를 서술한 문서도 없다.
- 따라서 `model-policy.md` 의 "두 축은 직교한다"는 문장은 **두 변수의 선언된 목적으로부터의 추론**이며,
  본문에 그렇게 표시했다 — *"an inference from the two variables' stated purposes … not a measurement.
  Treat it as a working assumption."*
- changelog 문구 자체의 정확한 의미론(예: `/model` 선택의 영속 범위가 프로젝트 단위인지 전역인지)도
  검증하지 않았다. 문서가 따라잡기 전까지는 changelog 문구 이상을 주장하지 않는다.

## 5. 범위 밖 발견 (조치하지 않음, 기록만)

상류 문서 열거 과정에서 드러난 사실: **공식 문서에 있는데 MoAI SSOT 에 없는 ANTHROPIC_ 이름이
`ANTHROPIC_DEFAULT_MODEL` 하나가 아니다.**

| 이름 | 공식 문서 | envkeys.go |
|---|---|---|
| `ANTHROPIC_MODEL` | 12건 | 없음 |
| `ANTHROPIC_CUSTOM_MODEL_OPTION` (+`_NAME`/`_DESCRIPTION`) | 4건 | 없음 |
| `ANTHROPIC_SMALL_FAST_MODEL` | 1건 | 없음 |

확인: `grep -rn --include='*.go' -e '"ANTHROPIC_MODEL"' -e 'ANTHROPIC_CUSTOM_MODEL_OPTION'
-e 'ANTHROPIC_SMALL_FAST_MODEL' internal/ pkg/ cmd/` → 0건.

이 카드의 범위는 `ANTHROPIC_DEFAULT_MODEL` 이므로 손대지 않았다. 다만 "네임스페이스 완전 열거"를
계약으로 내건 이상, 이 셋도 같은 종류의 미열거다 — **별도 카드 후보**로 남긴다. 한 축씩 정리하는 편이
낫다고 본다: 이 셋은 changelog-only 가 아니라 정식 문서화된 이름이라 서술 근거의 성격이 다르다.

## 6. 검증

| 검사 | 명령 | 결과 |
|---|---|---|
| 미러 사전 동일성 | `diff -q <local> <template>` (편집 전) | `IDENTICAL` — cp 안전 확인 |
| 미러 사후 동일성 | `diff -q <local> <template>` (cp 후) | `IDENTICAL` |
| 템플릿 중립성·parity | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -run 'Leak\|Neutral\|Parity\|Sanitiz'` | `ok` |
| 임베드 재생성 | `make build` | 성공, `catalog.yaml` 변동 없음 |
| 정적 검사 | `go vet ./internal/config/...` | 통과 |
| 패키지 테스트 | `go test ./internal/config/... -timeout 300s` | `config` / `atomicfile` / `toolpolicy` 3개 `ok` |
| SSOT 가드 명시 실행 | `go test ./internal/config/ -run 'TestAnthropic\|TestNoBareAnthropic' -v` | `--- PASS` 2/2 |

전체 스위트 미실행(카드 지시 + 로컬 부하 규율). 전 패키지 판정은 CI.
`make build` 가 `go build ./...` 를 포함하므로 전 패키지 컴파일은 확인됐다.

## 7. 잔여 위험

- §4 의 미측정 우선순위가 이 카드의 유일한 실질 리스크다. 문서가 따라잡은 뒤 직교성 추론이 틀린 것으로
  드러나면 `model-policy.md` 의 해당 문단을 고쳐야 한다 — 그래서 추론임을 본문에 남겼다.
- 상수는 프로덕션 소비자가 없다(의도된 상태). 죽은 상수로 보일 수 있으나, 이 패키지의 계약은
  "쓰이는 이름만 선언"이 아니라 "네임스페이스를 완전히 열거"이므로 계약대로다. 주석이 그 이유를 적고 있다.
- 이 변경은 `.claude/rules/moai/` 아래 파일을 건드리므로, 미러가 함께 커밋되지 않으면 다음
  `moai update` 에 템플릿판으로 덮어써진다. 미러는 같은 커밋에 포함했다.
