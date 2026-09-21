# t158 — ANTHROPIC_* SSOT 미열거 정리

베이스: `main` @ `4b2f203fe` · 워크트리 `.claude/worktrees/t158` · 브랜치 `WT-ssot-3names` · Tier S
선행: t157(`f80c5d3d3`) 범위 밖 발견. 성격은 t157과 같다 — **결함 수정이 아니라 SSOT 정렬**.

---

## 1. 카드 전제 정정 — 5개가 아니라 18개 (3차 재발 방지 기록)

카드는 미열거 이름을 **5개**로 셌다: `ANTHROPIC_MODEL` · `ANTHROPIC_CUSTOM_MODEL_OPTION`(+`_NAME`/`_DESCRIPTION`) · `ANTHROPIC_SMALL_FAST_MODEL`.
공식 문서를 이 세션에서 다시 받아 이름별로 세어 보니 **18개**였다.

| 출처 | 이름 | 수 |
|---|---|---|
| 카드가 센 것 | `ANTHROPIC_MODEL`(문서 14건) · `ANTHROPIC_CUSTOM_MODEL_OPTION`(4) + `_NAME`(2) + `_DESCRIPTION`(2) · `ANTHROPIC_SMALL_FAST_MODEL`(1) | 5 |
| 카드가 못 센 것 | tier 4종(`OPUS`/`SONNET`/`HAIKU`/`FABLE`) × `_NAME`/`_DESCRIPTION`/`_SUPPORTED_CAPABILITIES` | 12 |
| 카드가 못 센 것 | `ANTHROPIC_CUSTOM_MODEL_OPTION_SUPPORTED_CAPABILITIES` | 1 |

누락의 원인은 한 문장이다 — 문서 §"Customize pinned model display and capabilities" 말미:

> The same `_NAME`, `_DESCRIPTION`, and `_SUPPORTED_CAPABILITIES` suffixes are available for
> `ANTHROPIC_DEFAULT_SONNET_MODEL`, `ANTHROPIC_DEFAULT_HAIKU_MODEL`, `ANTHROPIC_DEFAULT_FABLE_MODEL`,
> and `ANTHROPIC_CUSTOM_MODEL_OPTION`.

**표만 세면 12개를 놓친다.** 문서의 접미사 표는 Opus 행 3개만 싣고 나머지 4개 base 는 산문 한 줄로 위임한다.
t157(→t158) 에 이어 세 번째 카드가 열리는 것을 막으려면, 열거 대상은 표가 아니라 **본문 산문까지** 훑어야 한다.
이번 조치는 18개 전량이므로 남은 미열거는 0이다(§6 검증).

리드 판정: 18개 전량 + F2/F3 교정 승인(부분 열거는 계약 위반 잔존, 편집 지점이 5개일 때와 동일해 증분 비용 0).

## 2. [HARD] 선확인 2건 — main 재실측

카드의 두 [HARD] 조항을 t157 결과 인용이 아니라 main 에서 다시 쟀다.

| 조항 | 확인 | 결과 |
|---|---|---|
| 런타임이 읽지 않음(=결함 아님) | 18개 각각 `grep -rn --include='*.go' … internal/ pkg/ cmd/` | 전부 0건. 유일 예외 `ANTHROPIC_MODEL` 2건은 `envkeys.go:356,366` **주석**이라 코드 참조 0 |
| prefix 매칭으로 딸려 들어가지 않음 | `statusline/memory.go:147-151` / `metrics.go:210-216` / `EnvAnthropicPrefix` 소비자 | memory 는 열거 슬라이스 3키, metrics 는 3분기 switch, prefix 는 프로덕션 소비자 0(정의부 2 + 테스트 2) |

**결론**: glob/prefix 스윕 경로가 프로덕션에 없으므로 새 이름 18개가 export 돼도 슬롯 해석에 끼어들 수 없다.
카드 전제대로 고칠 런타임 결함은 없었다. (카드의 "셋 다 0건"은 **코드 기준으로만** 정확 — 주석 2건은 별개.)

## 3. 조치

### 3.1 `internal/config/envkeys.go` — 18개 추가 (10 → 28)

블록 머리에 계약을 명시했다: 이 블록은 `ANTHROPIC_*` 네임스페이스를 **완전 열거**하며, 대부분은 MoAI 소비자가
없고 침묵의 공백을 없애기 위해 선언된다는 것, 그리고 출처가 `code.claude.com/docs/en/model-config` 라는 것.

배치는 축별로 묶었다 — 세션 모델(`ANTHROPIC_MODEL` → `ANTHROPIC_DEFAULT_MODEL`) / tier alias 4종 /
표시·능력 companion 12종 / custom picker 4종 / deprecated 1종. `_SUPPORTED_CAPABILITIES` 12·4종은
개별 주석으로 무엇을 덮는지만 적고, 적용 범위(3rd-party provider, gateway 시 `_NAME`/`_DESCRIPTION` 만,
`api.anthropic.com` 에서는 무효)는 묶음 머리 주석에 한 번만 적었다.

`ANTHROPIC_SMALL_FAST_MODEL` 은 **DEPRECATED**(→`ANTHROPIC_DEFAULT_HAIKU_MODEL`)임을 주석에 명시하고,
"열거는 하되 쓰지 말 것"을 적었다.

`EnvAnthropicDefaultModel` 주석은 §4 에 따라 다시 썼다(changelog-only → 문서 확정 + 우선순위 명문화).

### 3.2 `internal/config/anthropic_env_ssot_test.go` — 파생 집합 3곳 동시 갱신

t157 이 남긴 함정 그대로 **세 곳을 함께** 고쳤다. 한 곳만 고치면 크기 검사와 존재 검사가 어긋난다:

1. `bannedAnthropicEnvNames` — 18개 추가(`EnvAnthropicPrefix` 선두 유지, 나머지 상수명 알파벳 순)
2. `wantLen` 10 → 28, 그리고 헤더 주석의 "exactly the 10" 문구도 28 로
3. 런타임 커버리지 표 — 18행 추가(28행 전량)

### 3.3 `model-policy.md` — F2/F3 교정 + companion 절 신설

카드의 셋째 항목("세 축 표에 편입 여부 판단")에 해당한다. 편입했다.

- **세 축 표 정정**: `ANTHROPIC_MODEL` 행의 lifetime 을 정밀화(설정에 저장된 `/model` 선택이 이를 넘지 못하고
  다음 실행이 이 변수 값으로 돌아온다), `ANTHROPIC_DEFAULT_MODEL` 행에 v2.1.236+ 요구와 Default 행 라벨
  (`Set by ANTHROPIC_DEFAULT_MODEL`) 추가
- **F3 — 우선순위 명문화**: "the documented precedence" 문단 신설. 적용 조건(`--model`·`ANTHROPIC_MODEL`·
  settings `model`·조직 기본값 중 아무것도 모델을 고르지 않았을 때만), 무시 조건(`default`/`inherit`/
  `opusplan`/`haiku` 값, `enforceAvailableModels` on, allowlist 배제/계정 미보유), resume 시 동작
- **F2 — 낡은 서술 교체**: "문서 지연 / 우선순위 미관측 / working assumption" 두 문단을 제거하고,
  관측 범위 문단으로 대체 — 무엇이 뒤집혔는지(그때는 0건, 지금은 전용 절)를 남겨 이력이 사라지지 않게 했다
- **companion 절 신설**: `_NAME`/`_DESCRIPTION`/`_SUPPORTED_CAPABILITIES` 3종의 적용 범위, custom picker
  항목, `ANTHROPIC_SMALL_FAST_MODEL` deprecated, 그리고 "MoAI 는 하나도 읽지/쓰지 않으며 완전 열거가
  계약이라 선언한다"는 근거

### 3.4 Template-First

편집 **전** `diff -q` 로 로컬↔템플릿 byte-identical 확인 후에만 `cp`(중립화 차이가 있는 파일에 cp 하면
그 차이가 지워진다 — 여기서는 차이 0). 이후 `make build`. `catalog.yaml` 변동 없음(스킬 추가 아님 — 정상).

**중립성 가드에 한 번 걸렸고 고쳤다**: 초안이 관측 시점을 `2026-08-22` 로 적었는데
`TestTemplateNoInternalContentLeak` 가 `class=S1-internal-date` 로 잡았다. 날짜를 빼고 "Claude Code
v2.1.236 문서 기준"이라는 버전 기준 표현으로 바꿨다(날짜는 템플릿이 아닌 이 verdict 에 남긴다 — 관측 일자 2026-08-22).

## 4. 관측 — 무엇을 보고 무엇을 안 봤는가

**직접 관측**: `WebFetch https://code.claude.com/docs/en/model-config` 전문에서 `ANTHROPIC_[A-Z0-9_]*` 를
뽑아 세었다(이 세션, 2026-08-22).

```
15 ANTHROPIC_DEFAULT_OPUS_MODEL   14 ANTHROPIC_MODEL    8 ANTHROPIC_DEFAULT_SONNET_MODEL
 7 ANTHROPIC_DEFAULT_MODEL         7 ANTHROPIC_DEFAULT_FABLE_MODEL   6 ANTHROPIC_DEFAULT_HAIKU_MODEL
 5 ANTHROPIC_BASE_URL              4 ANTHROPIC_CUSTOM_MODEL_OPTION
 2 ANTHROPIC_DEFAULT_OPUS_MODEL_{SUPPORTED_CAPABILITIES,NAME,DESCRIPTION}
 2 ANTHROPIC_CUSTOM_MODEL_OPTION_{NAME,DESCRIPTION}   1 ANTHROPIC_SMALL_FAST_MODEL
 1 ANTHROPIC_DEFAULT_   ← 산문 속 글롭 조각, 이름 아님
```

**t157 대비 뒤집힌 값 (F2 의 근거)**: t157 이 같은 페이지에서 잰 값은 `ANTHROPIC_MODEL` 12건 ·
`ANTHROPIC_DEFAULT_MODEL` **0건**이었다. 지금은 각각 14건 · **7건**이고, §"Set a default model for new
sessions" 전용 절이 생겼다. 상류 문서가 따라잡았고, 따라서 t157 이 "문서 지연"이라 적은 문단은 지금 거짓이다.
t157 이 그것을 틀리게 적은 것이 아니라 — 그때는 사실이었고 지금 낡은 것이다.

**관측하지 않은 것 (Gap)**:
- **13개는 문서에 인쇄된 것이 아니라 규칙에서 파생했다.** 문서에 리터럴로 등장하는 이름은 15개뿐이고
  (`ANTHROPIC_DEFAULT_OPUS_MODEL_{NAME,DESCRIPTION,SUPPORTED_CAPABILITIES}` 는 표에 실렸다),
  Sonnet/Haiku/Fable 의 3접미사 9개 + `ANTHROPIC_CUSTOM_MODEL_OPTION_SUPPORTED_CAPABILITIES` 1개,
  합 10개는 §1 의 산문 한 줄("The same … suffixes are available for …")에 **적용 규칙만** 있고
  이름 자체는 인쇄돼 있지 않다. 나는 그 규칙을 적용해 이름을 **구성**했다. 규칙이 정확하다면 맞지만,
  이는 인용이 아니라 파생이다 — 상류가 어느 base 에 어느 접미사를 실제로 구현했는지는 재보지 않았다.
  집합 검증(§6)은 이 파생분을 "문서 리터럴에 없음"으로 정직하게 분류한다.
- **동작 시험 0.** 18개 중 어느 것도 export 해 보지 않았다. 우선순위·무시 조건 서술은 전부 **문서 인용**이며
  이 세션의 측정이 아니다. 문서가 틀렸다면 서술도 틀린다.
- `_SUPPORTED_CAPABILITIES` 의 값 문법(capability 토큰 집합)은 문서 표를 그대로 옮겼을 뿐 검증하지 않았다.
- `ANTHROPIC_SMALL_FAST_MODEL` 의 deprecated 이후 **실제 동작**(무시되는지 계속 먹히는지)은 문서가 말하지
  않고 재보지도 않았다. 주석은 "deprecated, 쓰지 말 것"까지만 주장한다.
- 전체 스위트 미실행(로컬 부하 규율, CLAUDE.local.md §4). 전 패키지 판정은 CI.

## 5. 범위 밖 발견 (조치하지 않음)

- 같은 문서가 `CLAUDE_CODE_SUBAGENT_MODEL` 을 모델 해석 표에 함께 싣는다. `CLAUDE_CODE_*` 네임스페이스는
  이 카드의 대상이 아니고, envkeys.go 에도 없다. `ANTHROPIC_*` 와 달리 그쪽은 "완전 열거" 계약을 선언한 적이
  없으므로 결함은 아니다 — 다만 같은 종류의 판단(계약을 걸 것인가)이 필요한 지점이라 기록만 남긴다.
- t157 verdict §5 가 후보로 남긴 3종은 이 카드가 흡수했다. 그 항목은 종결로 봐도 된다.

## 6. 검증

| 검사 | 명령 | 결과 |
|---|---|---|
| 남은 미열거 0 | `comm -23 <문서 리터럴 이름> <envkeys.go 이름>` | 차집합 `ANTHROPIC_DEFAULT_` 1건뿐 — 산문 속 글롭 조각이지 이름 아님 → **실질 0** |
| SSOT 이름 수 | `grep -o '"ANTHROPIC_[A-Z0-9_]*"' envkeys.go \| sort -u \| wc -l` | `28` (10 + 18) |
| 역방향 차집합 | `comm -13` (SSOT 에 있고 문서 리터럴에 없음) | 14건 — 파생 10건(§4) + 이 페이지 소관 밖 4건(`ANTHROPIC_`/`API_KEY`/`AUTH_TOKEN`/`REASONING_EFFORT`). 미확인 이름 0 |
| SSOT 가드 2개 | `go test ./internal/config/ -run 'TestAnthropic\|TestNoBareAnthropic' -v` | `--- PASS` 2/2 |
| config 패키지 | `go test ./internal/config/... -timeout 300s` | `config`/`atomicfile`/`toolpolicy` 3개 `ok` |
| 소비 패키지 회귀 | `go test ./internal/statusline/... -timeout 300s` | `ok` 13.956s |
| 정적 검사 | `go vet ./internal/config/... ./internal/statusline/... ./internal/cli/...` | rc=0 |
| 포맷 | `gofmt -l` (변경 2파일) | 출력 없음 |
| 전 패키지 컴파일 | `go build ./...` | `BUILD_ALL_OK` |
| 미러 동일성 | `diff -q` 로컬↔템플릿 (cp 전/후) | 양쪽 `IDENTICAL` |
| 템플릿 중립성·parity | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -run 'Leak\|Neutral\|Parity\|Sanitiz'` | `ok`(초안 1건 날짜 유출 → 수정 후 통과) |
| 임베드 재생성 | `make build` | 성공, `catalog.yaml` 변동 없음 |

## 7. 잔여 위험

- **§4 의 동작 시험 0 이 유일한 실질 리스크다.** 우선순위 서술은 전부 문서 인용이고, 문서가 다시 바뀌면
  이 문단도 t157 의 문단처럼 낡는다. 그래서 model-policy.md 에 관측 출처(페이지 + 절 이름 + 버전 기준)를
  본문에 박아 두었다 — 다음 사람이 무엇을 다시 재야 하는지 알 수 있게.
- **상수 28개 중 프로덕션 소비자는 여전히 소수다.** 죽은 상수로 보일 수 있으나 이 패키지의 계약이
  "쓰이는 이름만"이 아니라 "완전 열거"이므로 계약대로다. 블록 머리 주석이 그 이유를 적고 있다.
- `.claude/rules/moai/` 아래를 건드리므로 미러가 같은 커밋에 없으면 다음 `moai update` 에 덮어써진다.
  미러는 같은 커밋에 포함했다.
- 이 verdict 는 리드 지시대로 **primary 경로**에 썼고 브랜치에는 커밋하지 않았다 — primary 에서 untracked 다.
  PR 에 증거를 싣고 싶으면 알려주면 브랜치에 함께 커밋한다.
