# t197 — Codex 전용 런처: plan-phase 실측 기록

**증거의 원본은 이 문서가 아니라 `probe-output.txt` 다.** 이전 판들은 명령·rc·출력을 손으로 옮겨 적었고, 그 과정에서 실제로 6건이 틀어졌다 — 줄번호 4건, 축약한 grep 1건, 그리고 에이전트 TOML 수. 게다가 문서에 적힌 `${PIPESTATUS[…]}` 형태는 **내가 실제로 실행한 명령이 아니라** 문서용으로 재구성한 형태였고, rc 값은 다른 형태의 명령에서 얻은 것이었다. 그래서 방식을 바꿨다:

| 파일 | 역할 |
|---|---|
| `probe.sh` | 측정 스크립트. 각 단계의 명령·출력·rc 를 스스로 찍는다. 읽기 전용 |
| `probe-output.txt` | 그 스크립트를 1회 실행해 통째로 받은 **전사본**. 손을 대지 않았다 |
| `authshape.py` | `auth.json` 의 **형태** 와 비밀 아닌 두 값만 찍는 보조 스크립트 |
| 이 문서 | 전사본을 **해석** 한다. 사실 주장마다 전사본의 줄 범위를 가리킨다 |

재현:

```
bash .moai/reports/t197/probe.sh
```

측정 환경: 스크립트를 `bash 3.2.57` 로 실행했다 (`probe-output.txt` L2-4). `${PIPESTATUS[…]}` 는 bash 전용이므로 실행 셸을 이렇게 고정했다 — 대화형 도구 셸에서 같은 형태를 쓰면 unset 이 된다.

| 항목 | 값 | 전사본 |
|---|---|---|
| 트리 | `.claude/worktrees/t197` (worktree) | — |
| 브랜치 | `WT-codex-launcher` | — |
| 측정 기준 커밋 | `6bfb076bc` | L7-9 |
| 작업 트리 | 스크립트 자신의 산출물 3개만 untracked | L11-15 |
| t88(M4) 포함 | `git merge-base --is-ancestor 7b217da7c HEAD` rc 0 | L17-18 |

---

## M-1. `moai codex` 부재 (카드 전제) — 전사본 L20-27

도움말 **전문** 대상 대소문자 무시 검색이 0건이고 `grep -c` 의 rc 는 1(매치 없음)이다. 파이프 앞이 죽어 0이 나온 경우와 구분된다.

소스 쪽 유일한 히트는 `internal/cli/hook.go:221` 의 `codex-review-gate` — `moai hook` 하위 커맨드다. 최상위 `codex` 커맨드는 없다.

## M-2. auth 상태가 항상 `unknown` — 원인 확정 — 전사본 L29-89

- `codex login status` 는 `Logged in using ChatGPT` 를 rc 0 으로 낸다 (L30-32).
- 그런데 그 문구는 **stdout 0바이트 / stderr 24바이트** 로, 전량 stderr 로 나간다 (L34-40).
- `realCodexRunner.run` (L42-55) 이 `cmd.Stderr = &bytes.Buffer{}` 로 stderr 를 버리고 stdout 만 반환한다.
- `classifyCodexAuth` (L57-76) 는 그 빈 문자열을 받아 `codexAuthUnknown` 으로 fail-open 한다.

같은 블록이 두 번째 결함도 보여준다: 그 `switch` 는 **부분 일치** 라, 스트림을 고쳐 오류 문구가 들어오면 그 문구를 인증 성공으로 읽는다. 이것이 §C.2 설계에서 산문 파싱을 1순위에서 내리고, 남긴 2단마저 전체 행 문법으로 고정한 이유다.

프로브를 공유하는 표면은 셋이다 (L78-89): `codex_setup` MCP 도구, `moai web` 콘솔 카드(`internal/web/codex_state.go`, `schemaform.go`, `fieldsets_templ.go`), 그리고 신설될 런처.

**MCP 프로브 재현은 셸 명령이 아니다.** `mcp__moai__codex_setup` 도구를 인자 없이 호출해 받은 결과 전문이며, 셸에서 재실행할 수 없으므로 rc 가 없다:

```json
{"allow_write":false,"auth_provider":"unknown","binary":"/Users/goos/.local/bin/codex","enable_review_gate":false,"installed":true,"node_bridge":false,"version":"codex-cli 0.149.0"}
```

`installed`·`version` 은 맞고 `auth_provider` 만 틀렸다는 것이 위 기전과 일치한다.

## M-2b. 더 나은 auth 원천 — `auth_mode` 필드 — 전사본 L91-136

- `codex doctor` 는 auth 를 구조화해 안다: `stored auth mode  chatgpt` (L92-99).
- 기계 판독 형태도 있다 — `checks["auth.credentials"].details` (L101-118).
- **그러나 런처가 부를 수 없다**: `codex doctor --json` 이 **46.357초** 걸린다. 이 측정만 `time` 의 출력 형식 때문에 rc 를 함께 찍지 못했고, 따라서 전사본이 아니라 아래에 그대로 인용한다:

```
$ time codex doctor --json > /dev/null 2>&1
codex doctor --json > /dev/null 2>&1  11.76s user 9.82s system 46% cpu 46.357 total
```

느린 이유는 doctor 자신이 말한다 — `⚠ rollouts  31,525 active files · 1.44 GB on disk`.

- doctor 가 읽는 원본 파일은 즉시 읽힌다 (L120-136). `auth_mode = 'chatgpt'`, `OPENAI_API_KEY populated: False`, `tokens present: True`, `tokens non-empty values: 4`.

`authshape.py` 는 값이 아니라 **형태** 를 찍고, 값으로는 비밀이 아닌 둘(모드 문자열, 채워짐 여부 bool)만 낸다. 토큰 값은 어디에도 출력되지 않는다.

**이 머신에서 관측한 조합은 하나뿐이다** — `chatgpt` + 토큰 4종. `apikey` / `provider` 모드의 실제 파일 형태는 미관측이라, 설계는 알려지지 않은 값을 추측하지 않고 명령 프로브로 하강한다.

## M-3. codex CLI 가 이미 제공하는 표면

`codex --help` 의 `Commands:` 절은 24개 항목이다. 본 SPEC 이 참조하는 것은 셋 — `app`(데스크톱 앱 기동, 없으면 설치 관리자), `doctor`(설치·설정·auth·런타임 진단), `login`(로그인 관리). 이 절은 `probe.sh` 에 넣지 않았다(길이 대비 가치가 낮다); 필요하면 `codex --help` 로 직접 확인한다. 재실행 가능하지만 이 전사본에는 포함돼 있지 않다는 뜻이다.

## M-4. moai 쪽 Codex 배선 표면 (t88 M4) — 전사본 L138-176

- `.codex/hooks.json` · `.codex/config.toml` (L139-142)
- 에이전트 TOML **11종** (L144-160). **정정**: 이전 판은 12로 적었다. `ls | wc -l` 로 셌는데 이 셸의 `ls` 가 긴 형식 별칭이라 `total` 행이 함께 세어졌다. `find -type f` 로 세면 별칭에 영향받지 않는다.
- 생성기 `moai init --agent claude|codex|both` (L162-164), 진단 `checkCodexWiring` (L166-168), 런타임 `--harness codex` (L170-172)
- 이 저장소 자체는 `.codex/` 없음 — rc 1 (L174-176). 미배선이 기본 상태다.

## M-5. CODEX_HOME 을 읽는 런타임 코드는 0건 — 전사본 L178-180

출력 없음 + rc 1 = 매치 0건. `| wc -l` 로 세지 않았다 — 파이프 앞이 죽어도 0이 나와 부재의 증거가 되지 못한다.

## M-6. 기존 런처 3종의 공통 경로는 재사용 불가 — 전사본 L182-196

`cc` / `cg` / `glm` **셋 다** `unifiedLaunch` 로 수렴한다 (L188-192). 그 본체(`unifiedLaunchDefault`)는 `.claude/settings.local.json` 변형 + Claude 프로필 해석 + `claude` exec 이므로, 다른 바이너리인 codex 는 이 경로에 얹을 수 없다. 공유 가능한 것은 `spawnLaunch` 다 (L194-196).

## M-7. `codexCommandRunner` 구현체는 3개 — 전사본 L198-205

프로덕션 1 (`realCodexRunner`) + 시험 2 (`stubCodexRunner`, `fakeCodexRunner`).

최초 기록은 "프로덕션 1 + 시험 1" 이었고 **틀렸다** — Go 인터페이스는 구조적이라 이름 grep 으로는 암묵 구현을 찾을 수 없다. 메서드 시그니처로 세야 한다. 이 정정이 인터페이스를 넓히지 않는 설계(§C.2)의 계기다.

## 미관측 항목 (명시)

추정으로 채우지 않고 남긴다:

- `auth_mode` 의 `apikey` / `provider` 모드 실제 파일 형태 (로그아웃·재로그인 필요)
- `codex login status` 의 로그아웃 상태 출력 문구와 rc
- 데스크톱 앱 기동 (`codex app`) 의 실제 동작
- 런처 구현이 아직 없으므로 인자 전달·tmux 인용·크로스 플랫폼 동작 일체
- `moai init --agent codex` 를 실제 미배선 프로젝트에 돌린 결과 (REQ-CL-015/016 은 run-phase 에서 관측한다)
