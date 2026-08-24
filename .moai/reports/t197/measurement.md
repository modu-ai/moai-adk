# t197 — Codex 전용 런처: plan-phase 실측 기록

**증거의 원본은 이 문서가 아니라 `probe-output.txt` 다.** 이전 판들은 명령·rc·출력을 손으로 옮겨 적었고, 실제로 여러 건이 틀어졌다 — 줄번호 4건, 축약한 grep 1건, 에이전트 TOML 수(12→11), codex 커맨드 수(24→28). 게다가 문서에 적힌 `${PIPESTATUS[…]}` 형태는 **실행한 명령이 아니라 문서용 재구성** 이었고 rc 는 다른 형태의 명령에서 얻은 것이었다. 그래서 방식을 바꿨다: 스크립트가 스스로 찍고, 이 문서는 그 전사본의 줄 범위를 가리킨다.

| 파일 | 역할 |
|---|---|
| `probe.sh` | 측정 스크립트. **자기완결** — 측정 대상 바이너리를 스스로 빌드하고 doctor JSON 도 스스로 만든다. 각 명령의 **자기 rc** 를 검사해 예상과 다르면 마지막에 `PROBE FAILED` + rc 1 로 끝난다 (실패가 초록 전사본 뒤에 숨지 못한다) |
| `probe-output.txt` | 그 스크립트를 1회 실행해 받은 **무편집 전사본** |
| `authshape.py` | `auth.json` 의 **형태** 와 비밀 아닌 값만 찍는 보조 스크립트 |
| 이 문서 | 전사본을 **해석** 한다. 사실 주장마다 줄 범위를 가리킨다 |

재현: `bash .moai/reports/t197/probe.sh`

**읽기 전용이 아니다.** 두 단계가 저장소 밖을 건드린다 — 빌드가 임시 디렉터리에 바이너리를 쓰고(종료 시 삭제), `codex` 호출은 자기 홈에 PATH 별칭을 만들려 시도한다(이 머신에서는 실패하고 경고를 낸다). 저장소 작업 트리는 무변경이다. 이전 판의 "읽기 전용" 표기는 과장이었고 철회한다.

측정 환경: `bash 3.2.57` 로 실행했다 (전사본 L2-4). `${PIPESTATUS[…]}` 는 bash 전용이므로 실행 셸을 고정했다 — 대화형 도구 셸(zsh)에서 같은 형태를 쓰면 unset 이 된다.

| 항목 | 값 | 전사본 |
|---|---|---|
| 측정 대상 트리 | `1ed61e4ac` — 프로브가 스스로 찍은 값 | L24-26 |
| 작업 트리 | 스크립트가 자기 산출물만 수정 | L28-31 |
| t88(M4) 포함 | `git merge-base --is-ancestor 7b217da7c HEAD` rc 0 | L33-34 |

---

## M-1. `moai codex` 부재 (카드 전제) — L36-54

도움말 **전문** 대상 대소문자 무시 검색이 0건, `grep -c` rc 1(매치 없음). LAUNCH COMMANDS 는 `cc` / `glm` / `cg` 셋뿐(L41-50). 소스 쪽 유일한 히트는 `internal/cli/hook.go:221` 의 `codex-review-gate` — `moai hook` 하위 커맨드다(L51-53).

## M-2. auth 상태가 항상 `unknown` — 원인 확정 — L55-116

- `codex login status` → `Logged in using ChatGPT`, rc 0 (L56-58).
- 그 문구는 **stdout 0바이트 / stderr 24바이트** — 전량 stderr (L60-66).
- `realCodexRunner.run` (L68-81) 이 `cmd.Stderr = &bytes.Buffer{}` 로 stderr 를 버리고 stdout 만 반환.
- `classifyCodexAuth` (L83-102) 가 빈 문자열을 받아 `codexAuthUnknown` 으로 fail-open.

같은 블록이 두 번째 결함도 보여준다: 그 `switch` 는 **부분 일치** 라, 스트림을 고쳐 오류 문구가 들어오면 인증 성공으로 읽는다. 이것이 §C.2 에서 산문 파싱을 1순위에서 내리고 남긴 2단마저 전체 행 문법으로 고정한 이유다.

프로브 공유 표면 셋 (L104-115): `codex_setup` MCP 도구, `moai web` 콘솔 카드(`codex_state.go` / `schemaform.go` / `fieldsets_templ.go`), 신설될 런처.

**MCP 프로브 재현은 셸 명령이 아니다** — 전사본에 없다. `mcp__moai__codex_setup` 도구를 인자 없이 호출해 받은 결과 전문이며 셸에서 재실행할 수 없으므로 rc 가 없다:

```json
{"allow_write":false,"auth_provider":"unknown","binary":"/Users/goos/.local/bin/codex","enable_review_gate":false,"installed":true,"node_bridge":false,"version":"codex-cli 0.149.0"}
```

`installed`·`version` 은 맞고 `auth_provider` 만 틀렸다는 것이 위 기전과 일치한다.

## M-2b. 더 나은 auth 원천 — `auth_mode` 필드 — L117-167

- `codex doctor` 는 auth 를 구조화해 안다: `stored auth mode  chatgpt` (L118-125).
- 기계 판독 형태 — `checks["auth.credentials"].details` (L127-144).
- **런처가 부를 수 없다**: 커밋된 이 전사본에서 `codex doctor --json` 이 **67초** 걸렸다 (L19-21). 이것이 이 문서가 인용하는 유일한 소요 시간이다 — 실행마다 값이 갈리므로 커밋된 전사본에 있는 값만 쓴다. 판단(대화형 리드아웃이 매번 이만큼 기다릴 수는 없다)은 이 한 값으로 서고, 초 단위 정확도에 의존하지 않는다.
- 무거운 이유는 doctor 자신의 진단이 말한다 (L146-148): `state.rollout_db_parity: rollout files and state DB thread inventory differ`.
- doctor 가 읽는 원본 파일은 즉시 읽힌다 (L150-166): `auth_mode = 'chatgpt'`, `OPENAI_API_KEY populated: False`, `tokens present: True`, `tokens non-empty values: 4`.

`authshape.py` 는 값이 아니라 **형태** 를 찍고, 값으로는 비밀이 아닌 것(모드 문자열, 채워짐 여부 bool, 비지 않은 항목 수)만 낸다.

**이 머신에서 관측한 조합은 하나뿐이다** — `chatgpt` + 토큰 4종. `apikey` / `provider` 모드의 실제 파일 형태는 미관측이라, 설계는 알려지지 않은 값을 추측하지 않고 명령 프로브로 하강한다.

## M-3. codex CLI 가 이미 제공하는 표면 — L168-209

`Commands:` 절 전문이 전사본에 있고(L169-204), 항목 수는 **28** 이다(L206-208). 이전 판은 24로 적었다 — 눈으로 센 값이었고 틀렸다.

본 SPEC 이 참조하는 것은 셋: `app`(데스크톱 앱 기동, 없으면 설치 관리자), `doctor`(설치·설정·auth·런타임 진단), `login`(로그인 관리).

## M-4. moai 쪽 Codex 배선 표면 (t88 M4) — L210-249

- `.codex/hooks.json` · `.codex/config.toml` (L211-214)
- 에이전트 TOML **11종** (L216-232). 이전 판의 12는 `ls | wc -l` 이 별칭 긴 형식의 `total` 행을 함께 센 결과였다.
- 생성기 `moai init --agent claude|codex|both` (L234-236), 진단 `checkCodexWiring` (L238-240), 런타임 `--harness codex` (L242-244)
- 이 저장소 자체는 `.codex/` 없음, rc 1 (L246-248). 미배선이 기본 상태다.

## M-5. CODEX_HOME 을 읽는 런타임 코드는 0건 — L250-253

출력 없음 + rc 1 = 매치 0건. `| wc -l` 로 세지 않았다 — 파이프 앞이 죽어도 0이 나와 부재의 증거가 되지 못한다.

## M-6. 기존 런처 3종의 공통 경로는 재사용 불가 — L254-269

`cc` / `cg` / `glm` **셋 다** `unifiedLaunch` 로 수렴한다 (L260-264). 본체(`unifiedLaunchDefault`)는 `.claude/settings.local.json` 변형 + Claude 프로필 해석 + `claude` exec 이므로, 다른 바이너리인 codex 는 얹을 수 없다. 공유 가능한 것은 `spawnLaunch` (L266-268).

## M-7. `codexCommandRunner` 구현체는 3개 — L270-276

프로덕션 1 (`realCodexRunner`) + 시험 2 (`stubCodexRunner`, `fakeCodexRunner`).

최초 기록은 "프로덕션 1 + 시험 1" 이었고 틀렸다 — Go 인터페이스는 구조적이라 이름 grep 으로는 암묵 구현을 찾을 수 없다. 이 정정이 인터페이스를 넓히지 않는 설계(§C.2)의 계기다.

## 미관측 항목 (명시)

추정으로 채우지 않고 남긴다:

- `auth_mode` 의 `apikey` / `provider` 모드 실제 파일 형태 (로그아웃·재로그인 필요)
- `codex login status` 의 로그아웃 상태 출력 문구와 rc
- 데스크톱 앱 기동 (`codex app`) 의 실제 동작
- 런처 구현이 아직 없으므로 인자 전달·tmux 인용·크로스 플랫폼 동작 일체
- `moai init --agent codex` 를 실제 미배선 프로젝트에 돌린 결과 — 초기화는 `SPEC-CODEX-INIT-001` 로 분리됐고 그쪽 run-phase 에서 관측한다
