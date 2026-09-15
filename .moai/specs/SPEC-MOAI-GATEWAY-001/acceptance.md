# SPEC-MOAI-GATEWAY-001 — 수용 기준

설계 보고서 §11의 T01~T20 중 이 SPEC에 배정된 13개(T01·T02·T03·T04·T11·T12·T13·T14·
T15·T16·T17·T18·T19)를 검증 가능한 Given/When/Then 계약으로 옮겼다. 그 가운데 T02는 0.6.0에서 형제
`SPEC-MOAI-GATEWAY-PICKER-001`(제안)로 이관했고, 그 자리의 `AC-MG-002`는 폐기 묘비로 남긴다. 배정되지 않은
T05~T10·T20은 형제 SPEC 소관이다. T02의 제공자별 목록·Default 판정은 t649에서 AC-MG-001/003 보강으로 회수했다
(`spec.md` §G, 이 문서 §E).

각 AC 헤더의 괄호는 추적 대상 요구사항을 모두 적는다. 각 AC에는 **검증 가능성** 줄을 둔다.
외부 유료 호출, 실제 계정 로그인, Windows 실행처럼 현재 환경에서 수행할 수 없는 항목은
PASS가 아니라 **Gap**이다. AC가 존재한다는 사실은 그 AC의 Gap 부분이 PASS한다는 뜻이 아니다.

명칭 대응: 설계 보고서와 iter1 감사는 `AC-MP-nnn`, 이 문서는 `AC-MG-nnn`을 쓴다. 번호는
같다(`spec.md` §HISTORY 0.2.0 대응표).

## A. 배정 시험 유래 (T-매핑)

**AC-MG-001** (T01, REQ-MG-019, REQ-MG-002) — Given 세 launcher가 각각 기동한 세션이 있고, 격리
`CLAUDE_CONFIG_DIR`의 `settings.json`에 저장된 `model` 값이 없으며, `moai cc`의 설정된 Claude 기본 모델이 비어 있지
않을 때, When 사용자 `--model` 없이 기동한 각 세션의 최초 turn 요청이 gateway에 도달하면, Then 요청 본문 `model`이
`moai cc`는 설정된 Claude 기본 모델, `moai gpt`는 `gpt-5.6-sol`, `moai glm`은 설정된 GLM 기본 모델과 일치한다. 같은
픽스처에서 사용자가 `--model <X>`를 주면 최초 turn 요청의 `model`은 `<X>`다(`REQ-MG-002`).
`moai gpt`의 기본 `gpt-5.6-sol`은 `--model`이 없을 때만 적용된다. 같은 launch env의 네 별칭 슬롯은
`FABLE=gpt-6-astra`, `OPUS=gpt-5.6-sol`, `SONNET=gpt-5.6-terra`, `HAIKU=gpt-5.6-luna`로
서로 다른 직접 매핑을 유지한다.

"최초 turn 요청"은 제목 생성 요청과 구분한다. 프로브에서 제목 생성 요청은 turn 요청보다 먼저 도착했고 그 시점에 선택된
모델을 실었다(`research.md` §15 F8). 두 판정(launcher 초기 모델, 사용자 `--model`)은 서로 독립적으로 PASS 또는 Gap을
받는다. 한 판정의 PASS가 다른 판정의 PASS를 뜻하지 않는다.

**t649 보강 (REQ-MG-019).** Given 다른 제공자의 저장 기본값, 빈 Claude 기본값, 명시 모델 및 같은 제공자
재개 기록을 각각 준비했을 때, When 각 MoAI launcher를 실제로 시작하면, Then REQ-MG-019의 우선순위와
제공자 경계가 최초 요청에서 유지된다. `--model=X`와 `-m=X`도 각각 판정한다. 다른 제공자 ID는 명시 오류이며
세 upstream 요청 계수는 모두 0이다. 같은 제공자의 허용 ID는 실제 최초 요청에 그대로 실린다.
검증 가능성: 실제 MoAI launcher PTY와 요청 캡처가 필요하며 직접 Claude 프로브만으로 PASS하지 않는다.

**AC-MG-002** — [RETIRED] 0.6.0에서 형제 `SPEC-MOAI-GATEWAY-PICKER-001`(제안)로 이관되어 폐기되었다(배정 시험 T02,
picker 항목 집합과 "Default" 행 판정). 번호는 재사용하지 않는다. 이 줄은 결번 표시일 뿐 수용 기준이 아니며, 추적
요구사항을 갖지 않고 수용 기준 예산 계수에서 제외한다.

**AC-MG-003** (T03, REQ-MG-011, REQ-MG-015, REQ-MG-019, REQ-MG-023) — Given 각 launcher가
자기 제공자 모델 둘 이상을 가진 세션을 열었을 때, When `/model <id>` 및 picker `s`로 같은 제공자 모델을
전환하고 요청하면, Then 선택한 ID만 해당 제공자로 전달되고 다른 두 upstream의 요청 계수는 0이다.
Given 다른 제공자 ID를 직접 입력하거나 요청 본문에 넣었을 때, When gateway에 도달하면,
Then HTTP 404 `not_found_error`이며 모든 upstream 요청 계수는 0이다. 선택 목록과 요청 강제를 독립 판정한다.

**선택 시점 검증 요청 판정 (`plan.md` 결정 9).** 모든 upstream mock이 요청 계수기를 갖춘다. 검증 요청 픽스처는
`plan.md` M1 진입 게이트가 실제 Claude Code TUI에서 억제 플래그 없이 캡처한 **원본 요청 본문**에서만 고정한다. 인식 기준은
가정 A-VAL-2.1.268(`design.md` §4.1)의 네 조건이다 — `stream` 키가 없거나 값이 JSON `false`, `max_tokens`가 정수 `1`,
`messages`가 길이 1이고 그 항목의 `role`이 `user`, `tools`가 없거나 빈 배열. 원본 캡처의 형태가 이 기준과 다르면 픽스처를
고정하지 않고 SPEC을 먼저 고친다. 아래 (c)의 각 변형은 픽스처에서 한 가지씩만 벗어나며, 기대 결과가 미리 정해져 있다.

- (a) Given 검증 요청 형태의 요청이 세션 catalog에 없는 모델 ID를 담을 때, When gateway에 도달하면, Then 응답은
  HTTP 404 `not_found_error`이고 모든 upstream mock의 요청 계수는 `0`이다.
- (b) Given 검증 요청 형태의 요청이 catalog에 있으나 route의 직접 credential 또는 App Server 관리 세션 권한을 얻을 수 없는 모델 ID를 담을 때, When
  gateway에 도달하면, Then 응답은 HTTP 401 `authentication_error`이고 모든 upstream mock의 요청 계수는 `0`이며,
  클라이언트는 전환을 거부하고 다음 turn을 이전 모델로 보낸다.
- (c) Given catalog에 있고 credential도 얻을 수 있는 모델 ID를 담은 요청이 검증 요청 픽스처에서 한 가지씩만 벗어났을
  때, When gateway에 도달하면, Then 그 요청은 로컬로 답해지지 않고 대상 provider mock에 정확히 한 번 전달되며, 나머지
  upstream mock의 요청 계수는 `0`이다. 변형은 각각 따로 보낸다.
  - `max_tokens` 2
  - `max_tokens` 0 — `max_tokens <= 1`로 인식하는 구현은 이 변형에서 적색이다
  - `stream: true`
  - `stream: "false"` — 문자열은 JSON 불리언 `false`와 구분하여 인식하지 않는다
  - `stream: null` — 참거짓 판정(`!stream`)으로 인식하는 구현은 이 변형에서 적색이다
  - 메시지 둘
  - 메시지 하나이되 `role`이 `assistant`
  - 메시지 하나이되 `role`이 `system`
  - `tools`에 도구 정의가 하나 이상

  아홉 변형 가운데 `max_tokens` 0과 `role`이 `system`인 단일 메시지는, 본문 형태 검사(`design.md` §4)가 두 값을 거절
  조건으로 두지 않았다는 전제에서 같은 기대 결과를 갖는다. run 단계가 형태 검사나 `role: "system"` 항목 처리 규칙
  (`design.md` §4.2)에 두 형태를 막는 조건을 더하면, 코드보다 먼저 이 AC에서 그 변형의 기대 결과를 고친다. 어느 경우든
  로컬 성공 응답은 적색이다.

  양성 대조군은 두 개다. 원본처럼 `stream` 키가 없는 픽스처와, 같은 픽스처에 JSON `false`만 추가한 요청을 각각 보낸다.
  두 요청 모두 로컬 성공 응답을 받고 모든 upstream mock의 요청 계수가 `0`이다. 어느 대조군이든 적색이면 변형들의 결과는
  판정 근거가 되지 않는다. M1에서 캡처한 실제 첫 turn·이후 turn·제목 요청은 음성 대조군이며 로컬 검증 성공 응답을
  받지 않고 대상 provider mock에 각각 한 번 전달된다. 대조군의 원본 해시와 비식별 처리 이력은 픽스처 README에 기록한다.
- (d) Given picker `s`로 credential을 얻을 수 없는 모델로 전환한 세션일 때, When 그 모델로 첫 turn 요청
  (`stream: true`)이 도착하면, Then gateway는 명시 오류를 클라이언트에 드러내고, 모든 upstream mock의 요청 계수는
  `0`이며, 다른 provider로 넘기지 않는다.
- (e) Given 네 GPT 별칭 각각과 effort `max`·`high`·`medium`·`low`·`ultra` 및 effort 누락의
  직교 조합일 때, When launcher가 모델과 effort를 조립하고 App Server turn을 시작하면, Then alias는 항상
  지정된 GPT ID로만 해석되고 effort는 명시값 또는 누락 상태를 보존한다. 모델 선택에 따라 effort가 바뀌거나
  effort에 따라 모델이 바뀌는 조합은 0건이다. 설치 App Server가 특정 effort를 지원하지 않으면 다른 값으로
  대체하지 않고 해당 조합을 명시 오류로 거절한다.

검증 가능성: 부분.

- 클라이언트 동작은 실측되었다 — 같은 세션의 연속 전환, `/model <id>`의 검증 요청과 401·404에서의 전환 거부,
  picker `s` 경로가 검증 요청을 보내지 않는다는 사실이 Claude Code 2.1.267에서 관측되었다(프로브 1
  `.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe/README.md`, 프로브 2
  `.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe2/README.md`). 두 프로브는 Python mock을 상대로 한
  측정이므로 이 AC의 PASS 근거가 아니다.
- 두 프로브는 네 억제 플래그(`CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `DISABLE_TELEMETRY`, `DISABLE_ERROR_REPORTING`,
  `DISABLE_AUTOUPDATER`)를 켠 채 실행했고, 요청을 원본 본문이 아니라 mock이 파생한 요약으로 기록했다. `stream`은 Python
  `bool` 변환값이라 키 없음과 `false`를 가를 수 없다(`research.md` §15.6). 그래서 인식 기준의 형태 일치는 M1 진입 게이트의
  원본 캡처에 달렸다. 0.8.0은 2.1.268의 원문 다섯 건으로 조건을 보정했다(`research.md` §17).
  Sonnet 4.5를 쓴 이 과거 캡처는 향후 Opus 5·Sonnet 5 시험을 대신하지 않는다.
- gateway 수준 판정은 `plan.md` M1의 loopback 재현에 묶인다. 라우팅 기록과 (a)~(d)의 응답·계수는 upstream mock으로
  결정적으로 판정된다. (b)의 클라이언트 전환 거부는 실제 TUI를 gateway에 붙인 M1 재현에서만 관측되며, 그 전까지
  **Gap**이다.
- 실제 세 provider의 inference 왕복은 계정 권한이 필요하므로 **Gap**이다.
- 인식 기준은 2.1.268의 캡처로 보정한 가정이다. 다른 클라이언트 버전은 M1 캡처 게이트를 다시 통과해야 한다.

**AC-MG-004** (T04, REQ-MG-013) — Given 세 provider 각각으로 라우팅되는 요청이 있을 때,
When Read, 승인된 임시 Write, Bash 도구 호출이 발생하면, Then tool_use 블록이 대상
provider 형식으로 변환되어 나가고 tool_result가 원래 도구 이름과 ID로 역매핑되어
대화에 재입력된다.
GPT 변형에서는 App Server가 서로 다른 RPC ID·thread·turn·call ID를 가진 두 dynamic tool call을 보낼 때
각 호출이 Claude Code 승인·hook·실행을 정확히 한 번 거쳐 동일 RPC ID 응답으로 돌아가야 한다. 결과 순서를
뒤집어도 올바른 호출에 상관되어야 하며, 중복·미지 RPC ID·다른 대화 결과는 App Server 전달 0건으로 명시
거절된다. Codex native shell·file·MCP·agent·hook 실행 계수는 0이다.
검증 가능성: 부분 — 변환·역매핑 로직은 golden 시험으로 검증 가능. 실 provider 왕복은
**Gap**.

**AC-MG-005** (T11, REQ-MG-006) — Given 실행 직전 `~/.claude/settings.json`의
sha256, 프로젝트 `.claude/settings.json`의 sha256, 부모 shell의 `ANTHROPIC_BASE_URL` 값을
기록했을 때, When 세 launcher 중 하나로 세션을 열고 모델 전환 없이 종료하면, Then 세 값이
모두 변하지 않으며 시스템의 쓰기 대상 목록에 `~/.claude/settings.json`이 없다.
t649에서는 picker `s`·Enter·`/model <id>` 전환 뒤에도 같은 공유 설정 해시를 판정한다.
Claude의 선택 저장은 해당 제공자 사설 프로필에 한정되며 다른 launcher나 일반 Claude의 기본값을 바꾸지 않는다.
검증 가능성: 검증 가능 — 해시·env 비교와 쓰기 대상 목록 단언은 기계적이다.

**AC-MG-006** (T12, REQ-MG-005, REQ-MG-008, REQ-MG-009) — Given gateway child와 Claude
child가 살아 있을 때, When Claude가 종료 코드 `N`으로 끝나면, Then launcher 프로세스가 같은
`N`을 반환하고, gateway child가 정리 후 종료하며 listener 포트가 해제된다. 같은 방식으로 Ctrl-C(SIGINT), Ctrl-Z 후 `fg`, 부모 강제 종료 각각에 대해 좀비 gateway가
남지 않는다.
lead 종료 뒤의 요청(`REQ-MG-008`)은 다음으로 판정한다. Given gateway child가 lead 세션 종료를 감지해 종료했을 때,
When 옛 loopback 주소로 요청을 보내면, Then 요청은 연결 오류로 실패하고, Z.AI mock을 포함한 모든 upstream mock의 요청
계수는 lead 종료 시점 이후 `0`이다.
검증 가능성: 부분 — POSIX(darwin/linux)는 로컬에서 검증 가능. Windows 절반(`REQ-MG-009`)의 판정
지점은 **release PR 게이트**다. supervisor Windows 시험은 소유 패키지(`internal/gateway` 또는
`internal/cli`)의 `integration` 빌드 태그 없는 평범한 Go 시험이며 Windows에서 `t.Skip` 하지 않는다.
판정은 `release/*` → `main` PR에서 `.github/workflows/release-pr-multi-os.yml`의 windows-latest 레그가
올린 이벤트 스트림 아티팩트 `test-stream-release-verify-windows-latest`(파일 `test-stream.json.gz`)를
읽어, 이름을 정한 supervisor 시험 각각의 종료 이벤트가 `"Action":"pass"`임을 단언한다. 이름을 정한
시험 중 하나라도 스트림에 없거나 `pass`가 아닌 동작(`skip` 포함)으로 끝나거나 아티팩트 자체가 없으면
PASS가 아니다 — 업로드 단계가 `if-no-files-found: warn`이라 아티팩트가 없어도 레그는 실패하지 않는다.
**미리 선언한 Gap**: 대화형 TTY와 job control(Ctrl-Z 후 `fg`)의 Windows 동작은 CI runner에서 재현되지
않고, Windows에 대해 로컬에서 잰 것은 없다. 카드·develop CI(`ci.yml`)는 이 시험을 Windows에서 실행하지
않으므로 Windows 회귀는 release PR 시점에야 드러난다(잔여 위험).
**판정 시점과 기록.** 카드 run/sync가 닫히는 시점에는 Windows 판정이 아직 없으므로, 카드 완료 보고는 이 AC의
Windows 절반을 PASS가 아니라 Gap(판정 대기)으로 기록한다. release 배치의 리드(release PR을 여는 세션)가
windows-latest 레그가 끝난 뒤 아티팩트 보존 기간 7일(`release-pr-multi-os.yml:219`의 `retention-days: 7`)
안에 아티팩트를 읽고, 이름을 정한 시험별 종료 동작과 실행 URL을
`.moai/reports/SPEC-MOAI-GATEWAY-001/windows-release-verdict.md`에 기록한다. 기간 안에 읽지 못했으면
`workflow_dispatch`로 다시 실행해 읽으며, 만료로 사라진 아티팩트는 PASS의 근거가 되지 않는다.
in-process teammate가 lead의 gateway 주소를 물려받는지, lead 프로세스가 끝난 뒤 남는 teammate가 있는지는 이 AC가
판정하지 않는다. 둘 다 `plan.md` M7의 측정 과제이며 측정 전까지 **Gap**이다. tmux pane teammate의 수명 계약은 0.6.0에서
형제 `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안)로 이관했다.

**AC-MG-007** (T13, REQ-MG-013) — Given 하나의 응답 안에서 두 개 이상의 tool call이
교차로 스트리밍될 때, When gateway가 이를 중계하면, Then 각 call의 index·ID·인수가 서로
섞이지 않고 이름 역매핑이 요청 경계를 넘어 오염되지 않는다.
검증 가능성: 검증 가능 — 합성 SSE 픽스처로 결정적으로 재현된다.

**AC-MG-008** (T14, REQ-MG-012) — Given SSE 응답을 중계하는 중일 때, When upstream이
중간 EOF로 끊기거나 429를 반환하거나 클라이언트가 취소하면, Then gateway는 `message_stop`
같은 성공 terminal event를 만들어 붙이지 않고 오류를 그대로 드러내며, 이미 bytes를 보낸
뒤에는 재시도하지 않는다.
검증 가능성: 검증 가능 — 고의로 끊는 mock upstream으로 세 경로 모두 재현된다.

**AC-MG-009** (T15, REQ-MG-015, REQ-MG-017) — Given GPT 대화가 App Server의 소유 thread에 연결되어 있을 때,
When 답변·Claude 도구 왕복·새 프로세스 resume·model 변경·fork·compaction을 수행하면, Then 공개 맥락과
도구 결과가 정확히 한 번 반영되며 reasoning은 Codex의 이력으로 보존된다. MoAI가 raw opaque를 추출하거나
재구성하여 성공시키지 않는다. 다른 provider 요청·다른 대화의 결과·불명확한 pending 호출은 실행 전에 거절한다.

**400/history replay 이진 행렬.** Given 요청 history, receipt/manifest, App Server RPC, 외부 생성,
HTTP 응답과 구조화 로그를 각각 기록하는 격리 fixture가 있고, prompt·tool_result·reasoning·credential에 서로
다른 canary 비밀을 넣었을 때, When 다음 변형을 각각 독립 실행하면, Then 아래 결과와 정확히 일치해야 한다.

| 변형 | 외부 생성 횟수 | App Server `turn/start` | HTTP | 필수 로그 원인 |
|---|---:|---:|---|---|
| 동일 완료 prefix와 동일 `Idempotency-Key: history-contract-idempotency`의 transport 재시도 두 번 | 전체 시도 합계 2 | 전체 시도 합계 2 | 200 / 200 | 정상 귀속, upstream HTTP 합계 2 |
| 완료 prefix 뒤 새 사용자 입력 1건 | 1 | 1 | 성공 | 정상 귀속 |
| 이전 공개 message를 변경·삽입·삭제한 history | 0 | 0 | 400 `invalid_request_error` | `history_changed` |
| receipt에 귀속되지 않은 `agent_summary` 삽입 | 0 | 0 | 400 `invalid_request_error` | `agent_summary_untrusted` |
| request prefix와 receipt/manifest chain 불일치 | 0 | 0 | 400 `invalid_request_error` | `receipt_manifest_mismatch` |
| resume 직후 이미 반영된 사용자 입력 재제출 | 0 | 0 | 400 `invalid_request_error` | `resume_duplicate_input` |

`agent_summary` 변형은 구조적으로
`{"type":"agent_summary","summary":"CANARY-SUMMARY-92","source":"untrusted"}`를 포함해야 한다.
`ServerConfig.RejectionLogger`에 주입되는 recorder는 `RecordGatewayRejection(fields map[string]string)`를
구현하며 정확히 `cause`, `route`, 길이 64의 `digest` 세 필드만 받아야 한다. prompt·summary·tool·reasoning·
token·credential·canary key/value는 0건이어야 한다. 다섯 canary의 원문과 부분 문자열이 HTTP 오류·stdout·
stderr·구조화 로그 어디에도 0건이어야 한다. 양성 두 행이 실제 외부 생성/RPC 계수를
증가시키지 않으면 음성 행의 0회 판정은 유효하지 않다. 과거 t672·t851 보고나 receipt fixture의 존재는 이
현재 트리 행렬의 PASS를 대신하지 않는다.
검증 가능성: 아래 0.11.0의 카드별 시나리오 전체로 판정한다. 과거 direct carrier 프로브는 현재 경로의 PASS가 아니다.

**AC-MG-010** (T16, REQ-MG-011, REQ-MG-023) — Given 현재 대화의 토큰 길이가 대상 모델의
context 한도를 넘거나 대화에 미지원 이미지/PDF 입력이 있을 때, When 그 모델로 요청이
가면, Then gateway는 명시 오류를 반환하고 앞부분을 조용히 잘라내지 않는다.
검증 가능성: 검증 가능 — capability 표와 합성 입력으로 판정된다.

**AC-MG-011** (T17, REQ-MG-011) — Given 병렬 subagent 요청과 보조 slot 요청이 동시에
들어올 때, When gateway가 라우팅하면, Then 각 요청은 자기 `model` 값으로 독립 라우팅되고
직전 요청의 provider로 덮어써지지 않으며, 고정 slot 표시와 실제 route가 일치한다.
검증 가능성: 부분 — 동시 요청 라우팅 독립성은 검증 가능. 실제 Claude Code의 background
slot 동작 관측은 대화형 세션이 필요하므로 **Gap**.

**AC-MG-012** (T18, REQ-MG-002) — Given hooks와 MCP 서버가 활성인 프로젝트일 때,
When gateway 경유 세션에서 `/compact`를 포함한 통상 작업을 수행하면, Then hook 실행,
MCP 도구 호출, 승인 경계가 gateway 도입 이전과 동일하게 동작한다.
검증 가능성: **Gap** — 실제 대화형 세션과 승인 상호작용이 필요하다.

**AC-MG-013** (T19, REQ-MG-010, REQ-MG-024) — Given gateway가 기동해 있을 때, When 등록되지
않은 모델 ID, 등록되지 않은 path, 인증 없는 요청, redirect를 동반한 upstream 응답, 비정상
압축 본문이 각각 도착하면, Then 모두 외부 송신 이전에 거절되고, 로그 어디에도 API key·
OAuth token·요청 본문이 남지 않으며, listener는 loopback 외 주소를 열지 않는다.
검증 가능성: 검증 가능 — 각 케이스가 결정적 단위 시험으로 재현된다.

## B. 구조 계약 (요구사항 직접 유래)

**AC-MG-014** (REQ-MG-001, REQ-MG-002, REQ-MG-026) — Given 이 트리에서 빌드한 `moai`
바이너리가 있을 때, When 다음 판정을 각각 실행하면, Then 모두 성립한다.

- (a) `moai --help` → Launchers 그룹에 `moai cc`·`moai gpt`·`moai glm` 세 행이 보인다.
- (b) `moai gpt --help` → exit 0으로 사용법을 출력한다.
- (c) `moai gpt --spawn …`의 재발행 명령 조립 → 조립된 명령이 `moai gpt`로 시작한다
  (`cc`나 `glm` 리터럴로 새지 않는다).
- (d) `moai gpt -f`의 Factory lead와 Factory worker 분기에서 Dispatch 기록의 초기 provider가
  `gpt`이고, 같은 Factory 실행의 모델 별칭·effort가 REQ-MG-019 계약대로 유지된다.
- (e) `moai gpt -k`와 `moai gpt --kanban`은 각각 비성공으로 끝나고 `-f` 안내를 출력하며, Factory
  실행·Todo 변경·Tasks 생성·Dispatch 기록은 모두 0건이다.
- (f) 활성 CLI 도움말·새 사용자 출력·일반 런타임 식별자와 import 경로는 Factory/Todo/Tasks/Dispatch/
  Orchestration 명칭을 사용한다. 옛 명칭은 과거 SPEC·commit·report ID 또는 이름을 표시한
  migration/legacy 경계에서만 발견된다. `orchestration.BackendGPT`의 저장 값은 `"gpt"`다.
  판정은 `research.md` §22의 기계 allowlist와 inventory 명령을 그대로 사용한다.
  - allowlist를 제외한 production Go 파일의 `Kanban|kanban|-k|--kanban|MOAI_KANBAN_*` hit는 0이다.
    `internal/kanban` 디렉터리·`package kanban`·옛 import도 0이다.
  - legacy fixture의 구 저장 레코드와 `MOAI_KANBAN_*` 입력은 새 reader가 의미·ID·backend를 보존해
    읽는다. 같은 실행의 새 저장 write, env write, 사용자 출력에는 `Kanban`·`kanban`·`MOAI_KANBAN_*`
    hit가 0이고 Dispatch/Orchestration 이름만 있다.
  - `moai gpt -k`와 `moai gpt --kanban` 각각에서 종료 코드는 비성공이고 stderr에 `-f` 안내가 있으며,
    전후 Todo DB bytes/hash·Tasks snapshot·Dispatch event count·Factory process count·관련 env 투영이
    같다. 각 부작용 대조군은 `moai gpt -f`가 해당 계수나 상태를 실제로 바꾸는 양성 fixture다.
  - 현재 inventory의 47 production package 파일, 85 package test 파일, 36 active production importer와
    별도 historical utility 1개가 전부 새 경계로 분류됐는지 파일 집합 equality로 판정한다. 누락·추가
    파일이 있으면 allowlist를 조용히 넓히지 않고 inventory를 재감사한다.
  - 추적 manifest는 `.moai/specs/SPEC-MOAI-GATEWAY-001/naming-migration-manifest.json`이며 schema와
    현재 source set은 `research.md` §22.6이 고정한다. 미래 단일 checker 명령은
    `go test ./internal/orchestration -run '^TestNamingMigrationManifestContract$' -count=1 -v`이다.
    equality 성공은 test 1개·case 5개 PASS·exit 0을 출력해야 한다. manifest에만 있는 파일은
    `missing_from_tree=<sorted paths>`, tree에만 있는 파일은 `extra_in_tree=<sorted paths>`, schema
    불일치는 `schema_error=<field>:<reason>`을 해당 subtest의 실패 출력에 남기고 전체 exit를
    non-zero로 만들어야 한다. 누락·추가·schema 오류를 빈 목록 성공으로 처리하면 실패다.

같은 판정에서 UI의 `gpt` 표시가 과금 방식을 단정하지 않는지 확인한다(`REQ-MG-026`).
검증 가능성: 부분 — (a)~(f)의 구조·부작용 0건은 단위·통합 시험으로 판정한다. 실제 Factory 세션의
Claude Code Agent 실행 전 구간은 대화형 실증이 필요하므로 그 실증 전에는 **Gap**이다.

**AC-MG-015** (REQ-MG-004) — Given 이 트리에서 빌드한 `moai` 바이너리가 있을 때, When
(a) `moai --help` 출력을 캡처하고 (b) cobra 루트 명령의 하위 명령 트리를 전수 열거하면,
Then (a) 출력의 Launchers 행에 `moai gg`가 없고 같은 출력에 `moai cc` 행이 있으며(대조군),
(b) 열거 결과에 이름이 정확히 `gg`인 명령이 없고, 같은 열거에 `cc` 명령이 있으며
`helpGroupFrequency`의 `launch` 항목이 존재한다(대조군).

판정 계기는 help 출력과 command tree다. 저장소 전체 문자열 스윕은 이 AC의 계기가 아니다 —
이 SPEC 디렉터리(`.moai/specs/SPEC-MOAI-GATEWAY-001/`)와 감사 보고서(`.moai/reports/`)가 그
토큰을 담고 있어 스윕이 자기 자신을 세게 된다. 보조로 스윕을 돌린다면 **두 경로를 모두
제외**하고, 그 제외를 결과와 함께 기록한다.
검증 가능성: 검증 가능.

**AC-MG-016** (REQ-MG-005) — Given POSIX 빌드일 때, When launch 경로를 정적으로 읽으면,
Then Claude child를 띄우는 호출이 여전히 `syscall.Exec`이고, 그 환경 인수가
`withSessionPID(env, os.Getpid())` 형태 — 치환될 프로세스 자신의 PID를 `MOAI_SESSION_PID`로
각인하는 형태 — 이며, gateway는 그 호출 이전에 별도 프로세스로 분기되어 있다.
검증 가능성: 검증 가능 — 소스 가드 시험(호출 형태와 인수 형태 단언)으로 회귀를 막는다.

**AC-MG-017** (REQ-MG-010) — Given 요청 포트가 이미 점유되어 있을 때, When gateway가
기동하면, Then 임의 여유 포트를 얻거나 명시 오류로 끝나며, 어떤 경우에도 `0.0.0.0`이나
비-loopback 인터페이스에 bind하지 않는다.
검증 가능성: 검증 가능.

**AC-MG-018** (REQ-MG-018, REQ-MG-021, REQ-MG-022) — 이 AC의 판정 대상은 파일 전체가 아니라 **GLM 정리
키 집합**(`design.md` §6.1)이다. 층 (a) 라우팅 키 `ANTHROPIC_AUTH_TOKEN`, `MOAI_BACKUP_AUTH_TOKEN`,
`ANTHROPIC_BASE_URL`, `ANTHROPIC_DEFAULT_HAIKU_MODEL`, `ANTHROPIC_DEFAULT_SONNET_MODEL`,
`ANTHROPIC_DEFAULT_OPUS_MODEL`, `ANTHROPIC_DEFAULT_FABLE_MODEL`, 층 (b) 동작 영향 키
`CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, `API_TIMEOUT_MS`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW`,
`CLAUDE_CODE_MAX_CONTEXT_TOKENS`, 비회귀 키 `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`,
`CLAUDE_CODE_TEAMMATE_DISPLAY`, `MOAI_STATUSLINE_CONTEXT_SIZE`의 열네 키다. 이 목록은 기존 정리 함수
하나의 목록이 아니다 — `removeGLMEnv`의 `env` 키 13개에 `CLAUDE_CODE_MAX_CONTEXT_TOKENS`를 더한 것과
같다.

**제외 대상.** 최상위 `teammateMode` 키는 GLM 정리 키 집합의 투영 판정에서 명시적으로 뺀다. 이 키는 (a)의 teammate 표시 판정이 따로 본다. SessionStart 체인은 매 세션
`ensureTeammateMode`(`internal/hook/session_start.go:995`, 호출 `:631`)로 tmux 여부에 따라 이 키를
다시 쓰고(값이 이미 맞으면 `:1026-1027`에서 쓰기 전에 돌아간다), `removeGLMEnv`는 매 launch 이 키를
지운다(`internal/cli/launcher.go:402`). 둘 다 GLM과 무관한 정당한 쓰기이며, 이 키를 포함한 해시는 올바른
구현에서도 적색이 된다. `ensureTeammateMode`는 이 키를 다시 쓸 때 `env`의 legacy
`CLAUDE_CODE_TEAMMATE_DISPLAY`도 지운다(`:1033-1047`). 그 키는 열네 키의 구성원이지만 이 함수는 지우기만
하고 쓰지 않으며, launch 정리가 이미 지운 키이므로 (b)의 투영을 바꾸지 않는다. 집합 밖의 `env` 키(예:
Windows에서 쓰이는 `CLAUDE_ENV_FILE`)도 판정에서 뺀다.

- (a) **launch 단계 정리 — 남은 키 경로의 결정적 픽스처.** Given 격리 디렉터리의
  `settings.local.json` `env`에 열네 키가 모두 들어 있을 때(`ANTHROPIC_BASE_URL`은 Z.AI 주소, 네 모델
  슬롯은 GLM 모델 ID, `MOAI_BACKUP_AUTH_TOKEN`은 비어 있지 않은 표지 값, 나머지 키는 비어 있지 않은 임의
  값), When 비-GLM gateway launch(`moai cc`, `moai gpt` 각각)가 exec 직전에 이르면, Then 그 시점 파일의
  `env`에는 `ANTHROPIC_AUTH_TOKEN`을 뺀 열세 키가 하나도 없고, `ANTHROPIC_AUTH_TOKEN`은 픽스처의 백업
  표지 값과 같다. `MOAI_BACKUP_AUTH_TOKEN` 하나만 뺀 같은 픽스처로는 exec 직전 `env`에 열네 키가 하나도
  없다. `MOAI_BACKUP_AUTH_TOKEN`을 빈 문자열로 둔 같은 픽스처로도 exec 직전 `env`에 열네 키가 하나도 없다 —
  빈 백업 키를 남기는 구현(`removeGLMEnv`를 그대로 쓴 경우, `internal/cli/launcher.go:406-411`)은 이 변형에서
  적색이다. `moai glm` gateway launch도 같은 결과를 내야 한다. 세 launcher 모두 launch가 이 파일에
  `ANTHROPIC_BASE_URL`이나 GLM credential 값을 새로 쓰지 않았고, 자식 env의 `ANTHROPIC_BASE_URL`은
  loopback gateway 주소다. 기존 `removeGLMEnv`만 호출하는 구현은 `CLAUDE_CODE_MAX_CONTEXT_TOKENS`가 남아
  이 판정에서 적색이다.
  **Claude child env의 상속 키 정리 (iter5 G5-B1, 0.8.0 G6-A1).** Given launcher 프로세스 env에
  `design.md` §6.1의 GLM 정리 집합 14키 전부와 별도 `Z_AI_API_KEY`가 미리 들어 있을 때 —
  `ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`, `MOAI_BACKUP_AUTH_TOKEN`,
  `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL`, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`,
  `API_TIMEOUT_MS`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS`,
  `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `CLAUDE_CODE_TEAMMATE_DISPLAY`, `MOAI_STATUSLINE_CONTEXT_SIZE`,
  `Z_AI_API_KEY`. BASE_URL은 Z.AI 주소이고 나머지는 비어 있지 않은 상속 표지다. 표지 값은 서로 다르고 어떤 launcher도 스스로 설정하지 않는 문자열로
  둔다. `moai glm` 변형에서는 GLM credential 저장소가 상속 표지와 다른 값을 돌려주도록 `glmcred.Load`의 시험 seam
  `MOAI_TEST_GLM_KEY`(`internal/glmcred/glmcred.go:32`, `Load` `:100-103`)로 고정하고, 격리 프로젝트 설정의
  `llm.glm.models` 픽스처는 `high`·`medium`·`low`·`fable`에 서로 다르고 어느 상속 표지와도 다른 GLM 모델 ID를 둔다. When `moai cc`, `moai gpt`,
  `moai glm` 각각이 exec 직전에 이르러 exec에 넘길 env를 확정하면, Then 그 env에 대해 다음이 성립한다.
  - `moai cc`·`moai gpt` — 상속 표지 값이 어느 항목의 값에도 나타나지 않고, `MOAI_BACKUP_AUTH_TOKEN`과 `Z_AI_API_KEY` 키가
    없다. `moai gpt`에는 네 모델 슬롯이 모두 있고 값은 정확히
    `FABLE=gpt-6-astra`, `OPUS=gpt-5.6-sol`, `SONNET=gpt-5.6-terra`,
    `HAIKU=gpt-5.6-luna`다. 네 값을 현재 기본 모델 하나로 채우거나 GLM의 high·medium·low·fable
    설정을 재사용한 구현은 적색이다.
  - `moai glm` — `Z_AI_API_KEY`의 상속 표지를 뺀 상속 표지 값이 어느 항목의 값에도 나타나지 않고, `MOAI_BACKUP_AUTH_TOKEN`
    키가 없으며, `Z_AI_API_KEY`는 있고 그 값은 credential 저장소 픽스처 값과 같으며 상속 표지와 다르다. 네 모델 슬롯 키
    `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL`이 모두 있고, 각 값은 tier 픽스처의 `high`·`medium`·`low`·`fable`
    값과 같다(`spec.md` `REQ-MG-021`).
  - 세 launcher 모두 `ANTHROPIC_BASE_URL`은 loopback gateway 주소이고 `MOAI_LAUNCH_PROVIDER`는 launcher에 맞는 값이다.

  상속 표지 값은 세 launcher 모두에서 없어야 한다. cc·gpt도 모델 슬롯 키 넷이 모두 있고 값은 자기 제공자의
  세션 catalog에 있어야 한다. glm은 설정 tier 값과 비교한다. 세션 인증 운반 키는 확정한 별도 헤더 계약을 따른다.

  정리 집합 14키 각각을 하나씩 남기는 뮤턴트를 별도로 실행하면 해당 상속 표지 부재 판정이 적색이어야 한다.
  그 밖에 다음 조립 순서·후주입 뮤턴트도 판정한다. 오늘처럼 `os.Environ()`을 그대로 넘기는 조립기(`internal/cli/launcher.go:837`,
  `buildEnvForLaunch` `:1180-1200`)는 세 launcher 모두에서 적색이다. 정리를 launcher의 더하기 뒤에 두는 조립기는 loopback
  `ANTHROPIC_BASE_URL`까지 지워 적색이다. `moai glm`에서 상속된 `Z_AI_API_KEY`를 남기는 조립기는 값 비교에서 적색이다. `moai glm`에서 모델 슬롯 키를 더하지
  않거나 상속된 슬롯 값을 옮겨 쓰는 조립기는 슬롯 키의 유무·값 비교에서 적색이며, tier 대응을 뒤바꾼 조립기도 tier 픽스처
  값이 서로 다르므로 같은 비교에서 적색이다.

  **tmux 표면과 teammate 표시 (`plan.md` 결정 12).** Given tmux 세션 env 쓰기·지우기를 기록하는 가짜
  `tmux.SessionManager`(`injectTmuxSessionEnvVia`처럼 세션 관리자를 인수로 받는 경계)와 `tmux` 명령 대역, `TMUX` 설정,
  격리 프로젝트의 `settings.local.json`에 `teammateMode: "tmux"`가 미리 들어 있는 픽스처일 때, When tmux 세션 안의 gateway
  launch(`moai cc`, `moai gpt`, `moai glm` 각각)가 exec 직전에 이르고 이어 그 세션의 SessionStart 체인이 끝나면, Then 다음
  셋이 성립한다.
  - 기록된 tmux 세션 env 쓰기·지우기 호출이 launch와 SessionStart를 통틀어 `0`건이다.
  - SessionStart 체인이 끝난 시점의 `settings.local.json` 최상위 `teammateMode`는 `in-process`다. `tmux`도 `auto`도
    아니다 — Claude Code 문서는 `auto`가 tmux 세션 안에서 split pane을 연다고 적는다(`research.md` §16).
  - `moai glm` 변형에서는 시험이 공급한 GLM 키 표지 값과 Z.AI 주소가 어떤 tmux 호출의 인수에도 나타나지 않는다. 오늘
    `applyGLMMode`의 tmux 주입을 그대로 둔 구현은 이 변형에서 적색이다.

  같은 판정을 `TMUX`를 설정하지 않은 채 `teammateMode: "auto"`에서 시작하는 픽스처로도 한다. 세 launcher 각각에서
  SessionStart 체인이 끝난 시점의 `teammateMode`는 `in-process`다 — `TMUX`가 있을 때만 `in-process`를 강제하는 구현은 이
  변형에서 적색이다(iter5 G5-A1).

  대조군으로, `teammateMode: "in-process"`를 미리 넣은 같은 픽스처에서 gateway launch가 아닌 세션(`MOAI_LAUNCH_PROVIDER`
  없음)의 SessionStart 체인이 끝나면 `teammateMode`는 오늘처럼 `tmux`로 바뀐다. 값을 실제로 바꾸는 이 대조군이 픽스처가
  `ensureTeammateMode`의 tmux 분기(`internal/hook/session_start.go:1021-1023`)에 닿는다는 증거다. 본 판정의 픽스처는
  반대로 `tmux`에서 시작하므로, 값을 건드리지 않는 구현은 본 판정에서 적색이다. 클라이언트가 `in-process` 설정을 실제로
  따르는지는 이 AC가 판정하지 않는다(`plan.md` M7 측정 과제). tmux 밖 변형의 대조군으로, `teammateMode: "in-process"`에서
  시작한 gateway launch가 아닌 세션은 `TMUX` 없이 SessionStart 체인이 끝나면 오늘처럼 `auto`로 바뀐다
  (`internal/hook/session_start.go:1021`).
- (b) **세션 구간 불변** — Given gateway launch로 시작해 `MOAI_LAUNCH_PROVIDER`가 설정된
  세션일 때, When SessionStart 체인과 SessionEnd 정리가 실행되면, Then launch 정리 직후(exec
  직전)와 SessionEnd 직후의 GLM 정리 키 집합 투영(키 이름과 값의 정렬 목록)이 같다. 없던
  키는 계속 없다. `moai cc`와 `moai gpt`로 시작한 비-GLM 세션에서 반드시 확인하며, `moai glm`
  세션에서도 결과가 같아야 한다.
- (c) **판정 이관과 훅 억제** — Given 같은 세션에서, When `hookProcessEnvHasGLM`의 대체 판정이 실행되면,
  Then 그 판정은 `MOAI_LAUNCH_PROVIDER`가 `glm`일 때만 참이다.
  `cleanupGLMSettingsLocal`의 억제는 SessionEnd 시점의 `settings.local.json` `env`에 `ANTHROPIC_BASE_URL`
  (Z.AI 주소)과 대표 GLM 키 `ANTHROPIC_DEFAULT_OPUS_MODEL`(GLM 모델 ID)이 **아직 남아 있는** 픽스처로
  판정한다. 동시에 도는 다른 세션이 파일을 다시 쓴 상황을 본뜬 것이다. launch 정리가 base URL을 이미 지운
  파일로는 이 함수가 억제 여부와 무관하게 입구의 키 존재 판정에서 돌아가므로
  (`internal/hook/session_end.go:733-736`) 판정이 공허해진다. `MOAI_LAUNCH_PROVIDER`가 있으면 호출 뒤에도 두
  키가 값 그대로 남는다. 대조군으로, 같은 픽스처에서 `MOAI_LAUNCH_PROVIDER`가 없으면 두 키가 모두 지워진다 —
  이 대조군이 픽스처가 정리 분기에 실제로 닿는다는 증거다. 같은 판정에서 `ensureGLMCredentials`를 GLM 모델
  슬롯이 남은 픽스처로 **직접** 호출한다. launch 정리가 모델 슬롯을 지운 뒤에는 이 함수가 입구의 슬롯
  판정에서 돌아가 두 분기에 닿지 않으므로, (b)만으로는 억제 누락을 볼 수 없기 때문이다. 두 픽스처 모두
  `ANTHROPIC_DEFAULT_OPUS_MODEL`에 context 창 해석기가 아는 GLM 모델 ID를 담고, CG 모드 판정이 거짓인
  격리 프로젝트 디렉터리를 쓴다.
  - 토큰 있음 픽스처: 비어 있지 않은 `ANTHROPIC_AUTH_TOKEN`이 있고
    `CLAUDE_CODE_AUTO_COMPACT_WINDOW`·`CLAUDE_CODE_MAX_CONTEXT_TOKENS`는 없다.
  - 토큰 없음 픽스처: `ANTHROPIC_AUTH_TOKEN`이 없고, `MOAI_HOME`을 **절대 경로**인 격리 디렉터리(예:
    절대 경로를 돌려주는 `t.TempDir()`)로 두어 그 아래 `.env.glm`에 키가 있다. `paths.MoaiHome()`은
    비어 있지 않으면서 `filepath.IsAbs`가 참인 `MOAI_HOME`만 따르고, 상대 경로는 말없이 무시한 채 실제 홈
    아래 `.moai`로 돌아간다(`internal/paths/paths.go:69`). 그래서 시험은 훅을 실행하기 **전에**
    `paths.GlmEnvFile()`이 그 격리 디렉터리 아래 경로로 풀리는지 단언한다. 실제 홈으로 새면 시험은
    통과하지 않고 그 자리에서 실패한다. 훅의 키 읽기는 `glmcred.Load`를 거치지 않고 `paths.GlmEnvFile()`을
    직접 읽으므로 `MOAI_TEST_GLM_KEY` 시험 seam은 이 분기에 닿지 않는다.
    픽스처의 `.env.glm`에는 `loadGLMKeyFromEnvFile`의 스캐너가 받아들이는 형태 그대로 `GLM_API_KEY=<값>`
    한 줄을 쓴다(`internal/hook/session_start.go:1355-1370`). 스캐너는 줄 앞뒤 공백을 걷어 내고 빈 줄과
    `#`로 시작하는 줄을 건너뛴다(`:1357-1358`). 첫 `=`에서 나눈 왼쪽을 공백 제거 후 `GLM_API_KEY`와
    정확히 비교하므로(`:1361`, `:1365`, `:1369`) `export GLM_API_KEY=…`처럼 접두가 붙은 줄은 버려진다.
    오른쪽은 공백 제거 뒤 양 끝의 `"`·`'` 문자만 걷어 내고 이스케이프는 풀지 않으므로(`:1366-1367`), 값은
    `$`·`"`·`'`·`\`·공백이 없는 영숫자 문자열로 한다. 값이 비어 있지 않은 첫 일치 줄에서 반환한다
    (`:1369-1370`). 형태가 틀린 줄이면 훅은 키를 읽지 못해 파일을 쓰지 않고 돌아가므로
    (`session_start.go:872-878`), 아래 대조군의 토큰 없음 픽스처에 `ANTHROPIC_AUTH_TOKEN`이 생기지 않아
    대조군이 적색으로 이 실수를 드러낸다. 다만 이 노출은 앞의 절대 경로 단언이 실제 홈으로의 폴백을 막을
    때만 성립한다 — 실제 홈의 `.env.glm`에 키가 있으면 대조군도 녹색이 된다.

  `MOAI_LAUNCH_PROVIDER`가 설정되어 있으면 두 픽스처 모두 호출 전후의 GLM 정리 키 집합 투영이 같다.
  대조군으로, `MOAI_LAUNCH_PROVIDER`가 없으면 토큰 있음 픽스처에는 `CLAUDE_CODE_MAX_CONTEXT_TOKENS`가 생기고
  토큰 없음 픽스처에는 `ANTHROPIC_AUTH_TOKEN`이 생긴다 — 이 대조군이 픽스처가 실제로 두 분기에 닿는다는
  증거다. credential 주입 분기만 억제한 구현은 토큰 있음 픽스처에서 적색이다. 또한
  `MOAI_LAUNCH_PROVIDER`가 없는 세션은 `hookProcessEnvHasGLM`·`cleanupGLMSettingsLocal`의 기존 판정(부분
  문자열 검사, 키 존재 검사)과 기존 정리 동작을 그대로 유지한다.

  **훅의 tmux 동작 억제.** SessionEnd의 tmux 세션 env 정리와 SessionStart의 `ensureTmuxGLMEnv`를 tmux 호출을
  기록하는 대역으로 실행한다. 두 함수 모두 오늘은 대역을 받는 경계가 없다 — `ensureTmuxGLMEnv`는 함수 안에서
  세션 관리자를 만들어 쓰고(`internal/hook/glm_tmux.go:132`, `:143`), 훅의 tmux 정리는 `tmux` 명령을 직접
  실행한다(`session_end.go:658`). 그 경계는 run 단계가 만든다. 픽스처는 `TMUX`를 설정하고,
  `settings.local.json`에 `teammateMode: "tmux"`와 비어 있지 않은 `ANTHROPIC_AUTH_TOKEN`(백업에서 복원된 사용자
  토큰을 본뜬 값)을 둔다. `MOAI_LAUNCH_PROVIDER`가 있으면 두 동작 모두 tmux 세션 env에 대한 쓰기·지우기
  호출이 0건이다. 대조군으로 signal이 없으면 SessionEnd 정리는 지우기 호출을, `ensureTmuxGLMEnv`는 민감 값
  채널 주입 호출을 각각 한 건 이상 남긴다.

검증 가능성: 검증 가능 — (a)~(c)는 격리 디렉터리의 `settings.local.json` 픽스처와 키 투영
비교로, (a)의 Claude child env 판정은 exec에 넘길 env의 기록으로 기계적으로 판정된다. 올바른 구현이 이 AC를 통과할 수 있는 근거: 현재 프로덕션 launch 경로 중
이 파일에 GLM 키를 쓰는 것은 없고(`research.md` §1.5), 이 키 집합을 세션 구간에 쓰는 SessionStart
경로는 `ensureGLMCredentials`(그 안에서 호출되는 context 창 보조 함수 둘 포함) 하나이며(`ensureTeammateMode`는
그 집합의 `CLAUDE_CODE_TEAMMATE_DISPLAY`를 지울 수는 있어도 쓰지 않는다), gateway
launch에서는 두 분기 모두 억제된다. 이 근거는 non-test Go 정적 검색으로 확인한 것이며 런타임에서
관측한 것은 아니다(`research.md` §6.3, §6.5). settings `env`가 프로세스 env보다 우선하는지는 이 AC가
판정하지 않는다 — 측정하지 않은 Claude Code 동작이다.
(a)~(c)와 각 하위 판정의 tmux 표면·Claude child env 판정은 서로 독립적으로 PASS 또는 Gap을 받는다. 한 하위 판정의 PASS가 다른 하위
판정의 PASS를 뜻하지 않는다. tmux 표면의 판정은 기록형 대역의 호출 기록과 파일 판독으로 결정적이다. tmux pane
teammate가 tmux 세션 env를 물려받는지는 0.6.0에서 형제 `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안)로 이관했다.

**AC-MG-019** (REQ-MG-024) — Given `MOAI_LAUNCH_PROVIDER`를 포함한 새 gateway 환경변수 이름을
도입했을 때, When 변경 패키지 테스트 `go test ./internal/config/...`를 실행하면, Then
`TestNoBareAnthropicEnvVarLiteralsInProduction`과 `TestAnthropicBannedSetCoversAllNames`가
통과한다. 이 가드는 `go build`로는 잡히지 않으므로 이 테스트 실행 자체가 판정 계기다.
검증 가능성: 검증 가능.

**AC-MG-020** (REQ-MG-025, REQ-MG-017) — Given 일반 Codex와 선택한 MoAI App Server 인증 프로필이 있을 때,
When managed 로그인·생성·갱신·로그아웃과 API 선택을 수행하면, Then MoAI의 credential 파일 읽기·쓰기·삭제는 0이고
Codex의 승인된 인증 변경은 선택 프로필에만 적용된다. 다른 프로필의 hash·mtime은 유지된다.
검증 가능성: 프로세스별 파일 접근과 auth 모드 전환을 실제로 기록한다. Codex 자체 갱신까지 파일 불변으로 요구하지 않는다.

## C. 무연결 요구사항 해소 (iter1 B1)

**AC-MG-021** (REQ-MG-017, REQ-MG-016) — Given gateway의 세션 catalog가 조립되었을 때,
When (a) registry의 GPT provider 항목을 열거하고, (b) 등록되지 않은 GPT 계열 ID
`gpt-5.3`과 `gpt-6-astro`로 각각 요청을 보내고, (c) T09 측정이 양성으로 기록되지 않은
상태에서 Claude 모델 요청을 보내면, Then (a) 열거 결과가 정확히 `gpt-5.6-sol`·
`gpt-5.6-terra`·`gpt-5.6-luna`·`gpt-6-astra` 네 개이고, (b) 두 요청 모두 upstream 송신
0건으로 명시 거절되며 다른 GPT ID로 대체되지 않고, (c) Anthropic passthrough 경로가 활성화되지
않아 OAuth 헤더가 upstream으로 전달되지 않는다.

(c)의 게이트 상태는 `design.md` §2.2가 정한 표현으로 판정한다. 게이트는 설정 키나 런타임
플래그가 아니라 **코드와 catalog 항목의 부재**로 표현된다 — 구독 OAuth passthrough용
`CredentialRef` 구체 타입이 없고, catalog에 그 인증 방식을 쓰는 항목이 없는 상태다. 따라서
(c)는 catalog 열거와 upstream mock의 OAuth 헤더 수신 0건으로 판정한다.

이 AC는 서로 다른 두 요구사항(GPT ID 정확성, passthrough 게이트)을 한 판정 표면에 묶는다.
수용 기준 예산이 25/25로 차 있어 묶었고, 두 판정은 서로 독립적으로 PASS 또는 Gap을 받는다.
검증 가능성: 부분 — (a)(b)와 (c)의 게이트가 닫힌 쪽 동작은 mock upstream 요청 계수로
결정적으로 판정된다. **미리 선언한 Gap** 두 가지: 네 GPT ID 각각의 실제 upstream 접근 가능
여부(`REQ-MG-017`), T09 공존이 양성일 때의 passthrough 동작(`REQ-MG-016`). 둘 다 계정 권한이
필요하다. 이 AC가 존재한다는 사실은 두 Gap 부분의 PASS를 뜻하지 않는다.

**AC-MG-022** (REQ-MG-022) — Given 세 provider 각각의 mock upstream과, 선택되지 않은 유료 API
endpoint mock이 모두 요청 계수기를 갖추고 있을 때, When 선택된 provider의 upstream이 5xx, 429,
인증 실패, 연결 거부를 각각 반환하면, Then gateway는 그 오류를 클라이언트에 드러내고, 다른
provider mock과 선택되지 않은 유료 API mock의 요청 계수는 모두 `0`이다.
검증 가능성: 검증 가능 — mock 계수 비교로 결정적이다. 설계 보고서 T08의 실계정 과금 판정은
형제 SPEC 소관이며, 이 AC는 T08 가운데 이 SPEC에 남은 fallback 부재 의무만 검증한다. 이 계수는
**gateway를 거친 요청만** 센다. `settings.local.json`에 남은 키 때문에 요청이 gateway 자체를
우회하는 경로는 원리상 이 AC가 보지 못하며, 그 경로는 `AC-MG-018` (a)가 launch 단계에서
판정한다. gateway launch가 tmux pane teammate 표시를 고르지 않고 tmux 세션 env에 쓰지 않는다는 조건은 `AC-MG-018`
(a)의 tmux 표면이 판정한다. tmux pane teammate가 tmux 세션 env에 남은 키로 gateway를 우회하는 경로는 형제
`SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안) 소관이다.

**AC-MG-023** (REQ-MG-003, REQ-MG-017) — Given 현재 바이너리가 있을 때, When `gpt login`·`logout`·
`status`와 미등록 동사를 실행하면, Then 앞의 셋은 각각의 처리기에 연결되고 로그인은 브라우저 구독·device-code 구독·
API 키를 명시 선택할 수 있다. 취소·실패는 명시 비성공이며 미등록 동사는 오류다. `--model`은 launch 인수로 처리된다.
검증 가능성: 실제 App Server managed/API 상태 및 오류 종료 코드를 대조한다.

**AC-MG-024** (REQ-MG-014) — Given gateway가 기동해 있고 세션 catalog가 조립되었을 때,
When 아래 세 요청을 각각 보내면, Then 세 의무가 서로 독립적으로 판정된다.

- (a) **count_tokens 정확도 표기** — `/v1/messages/count_tokens`를 provider마다 한 번씩
  호출하면, 응답 또는 문서화된 응답 필드가 그 값이 tokenizer 정확값인지 추정치인지를
  provider별로 표시하고, 추정치를 정확값으로 표시한 provider가 `0`개다.
- (b) **`/v1/models` 범위** — `/v1/models`의 모델 ID 집합이 세션 catalog의 route ID 집합과
  정확히 같다(누락 0, 추가 0).
- (c) **미등록 path 정책** — path 정책 표에 없는 path(`/v1/complete`, `/v1/unknown-path`)로
  요청하면 명시 거절되고, 모든 mock upstream의 요청 계수가 `0`이다.

검증 가능성: 검증 가능 — 세 판정 모두 mock 계수와 응답 본문으로 결정적이다.

**AC-MG-025** (REQ-MG-018, REQ-MG-019) — Given `llm.glm.models` 픽스처(`high`·`medium`·`low`·`fable`에 서로 다른 GLM 모델
ID)로 `moai glm` gateway 세션을 열고 Z.AI mock upstream이 수신 요청을 기록할 때, When GLM 모델 ID로 요청을 보내면, Then
기록된 요청에 `anthropic-beta` 헤더가 없고, 인증 값은 기존 GLM credential 저장소에서 해석된다. tier별 모델 매핑은 두
판정으로 본다 — 네 tier ID가 각각 세션 registry에서 Z.AI mock route로 해석되고 그 ID로 보낸 요청이 Z.AI mock에 기록되며,
`llm.glm.models`에 없는 Claude 모델 ID는 Z.AI route로 해석되지 않는다. launcher가 슬롯 키에 싣는 값은 `AC-MG-018` (a)가
판정한다. 같은 판정에서
`moai glm setup`·`moai glm status`·`moai glm tools` 하위 명령이 gateway 도입 이전과 같은
종료 코드와 출력 형태를 낸다.
검증 가능성: 부분 — 헤더·인증 해석·매핑·하위 명령은 mock과 시험 seam(`MOAI_TEST_GLM_KEY`)으로
판정된다. 실제 Z.AI inference 왕복은 계정 권한이 필요하므로 **Gap**.

**AC-MG-026** (REQ-MG-027, REQ-MG-019) — launcher 생산 통합과 로컬 배포 게이트(0.13.0, t654).
네 하위 판정은 서로 독립적으로 PASS 또는 Gap을 받는다.

- (a) **launcher 생산 통합.** Given AS-014~AS-022의 검증 게이트가 통과한 트리일 때, When 이 트리에서
  빌드한 `moai gpt`를 launch 인수와 함께 실행하면, Then launch가 "GPT gateway launch is awaiting
  transport verification" 대기 오류로 끝나지 않고 실제 launch 경로로 진행하며, 세 launcher 모두
  provider 전용 catalog·인증 방식 표시·App Server transport를 같은 launch 조립에서 넘긴다. 검증
  게이트 통과 전 트리에서는 같은 실행이 대기 오류로 끝나는 대조군이 된다. 기계 판정: 게이트 통과
  트리에서 비-테스트 `internal/cli`에 그 대기 오류 리터럴이 0건 —
  `grep -rn 'awaiting transport verification' internal/cli --include='*.go' | grep -v _test` → 0,
  launch 조립 시험 `go test ./internal/cli/ -run 'TestGatewayLaunch'` PASS(RED→GREEN 로그 보존).
- (b) **compaction epoch 생산 복원.** Given 세션 영구 상태에 마지막 적용 epoch가 기록된 대화 scope가
  있을 때, When 프로세스를 다시 띄워 compaction 원장을 복원하면, Then 주입된 applied epoch가 기록값과
  같고 뒤이은 첫 rebase가 그 값 기준으로 판정된다. 영구 상태가 없으면 0으로 시작하고, 판독 불가·훼손
  상태는 명시 오류다. 기계 판정: `go test ./internal/gateway/... -run 'TestRebaseLedgerRestoration'`
  RED→GREEN — 고정값·호출자 추측 주입(현행, 비-테스트 호출자 0) 구현은 복원 시험에서 적색이다.
- (c) **fork inherited prefix 원장 대조.** Given 원본 family의 완료 원장과 `--fork-session` 자식의
  inherited prefix가 있을 때, When gateway가 자식 배리어를 수용하면, Then prefix가 `Manifest.ChainTo`
  경계와 일치할 때만 수용되고, 변조·불일치·미지 원본은 자식 상태 생성 전에 명시 거절된다. engine의
  caller-asserted 값을 대조 없이 수용하는 구현은 변조 변형에서 적색이다. 기계 판정:
  `go test ./internal/gateway/... -run 'TestForkPrefixCrossCheck'` RED→GREEN.
- (d) **rc 로컬 배포 게이트(종결).** Given (a)·(b)·(c)가 PASS이고 AS-017·AS-019·AS-021이 PASS이며
  AS-014·AS-018·AS-020·AS-022가 PASS 또는 근거를 갖춘 Gap일 때 — 즉 강제 전제 집합은
  {(a), (b), (c), AS-017, AS-019, AS-021}의 전수 PASS다(이 여섯에 Gap은 허용되지 않는다) — When
  배포 절차를 실행하면, Then 다음 세 증거가 각각의 실제
  명령과 출력과 함께 `.moai/reports/t654/as5-deploy-verdict.md`에 남는다.
  1. `make build VERSION=v<다음 미사용 rc>` → exit 0. 버전 번호는 `.moai/docs/version-management.md`
     Local RC Numbering의 다음 미사용 번호다 — 카드 문구의 rc.8은 2026-09-12 발행 시점 표기이며,
     발행 시점에 이미 소비됐으면 다음 번호를 쓰고 그 사실을 보고서에 명시한다.
  2. `rm -f ~/go/bin/moai && cp bin/moai ~/go/bin/moai` → clean 재설치(inode 갱신; 맨 cp 덮어쓰기는
     exit 137 전례가 있어 clean 재설치가 계약이다).
  3. `sh scripts/verify-local-install.sh` → `bin/moai`와 `~/go/bin/moai`가 byte 단위로 같고 설치본의
     `version` 명령이 exit 0이며 측정 시점 HEAD의 short SHA를 출력한다. macOS `strings`나 Xcode
     라이선스 상태에 의존하는 판정은 허용하지 않는다.
  보고서는 CHANGELOG 발행 검토 결과(사용자 가시 표면 기준, 사전-발행 grep
  `grep -c 'SPEC-MOAI-GATEWAY-001' CHANGELOG.md` 포함)를 함께 담는다.
  push·PR·병합·워크트리 제거는 없으며, 배포 절차 후 `git status --short`가 증거 파일 외 로컬 변경
  없음을 보이는 것까지 판정에 포함한다.
검증 가능성: (a)~(c)는 기계 시험으로, (d)는 명령 출력과 보고서로 판정된다. (d)의 전제인 실제 계정
실증(AS-017·AS-019·AS-021)이 세션 환경에서 불가하면 (d)는 창 대기 상태로 남고 PASS로 세지 않는다.

## D. Definition of Done

- `AC-MG-001` ~ `AC-MG-026` 가운데 폐기 묘비 `AC-MG-002`를 뺀 25개 각각이 판정되어야 한다.
  아래 RED-now 행은 현재 제품 결함을 재현하는 selector이며 모두 GREEN이어야 한다. 이미 GREEN인 회귀
  가드와 아직 실행하지 않은 live/Windows 사건은 RED 자격을 주장하지 않지만 별도 완료 의존성으로 남는다.
- M14-R0.2는 제품을 바꾸지 않은 기준선
  `4056f69e1c20d942d4f9fc7363d3d79bffde899a`에서 실행됐다. 다음 fenced ledger가 각 실제 RED selector의
  명령·원문 stdout 실패 일부·exit·raw path/digest를 결합하는 canonical MP-8 carrier다.

`EV-R0-ALIAS-ENV-RED`
```text
command: go test ./internal/cli -run '^TestGPTAliasEffortMatrix$' -count=1 -v
baseline: 4056f69e1c20d942d4f9fc7363d3d79bffde899a
verbatim:     gpt_alias_effort_contract_test.go:48: alias fable: got gpt-5.6-sol, want gpt-6-astra (effort max)
exit: 1
log: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/01-alias-effort.log sha256=28f2d0aef1d104363db0106bcfb0d2917d3a145153c75fd8cbd434fbf88d6e10
exit-carrier: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/01-alias-effort.exit sha256=4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865
```

`EV-R0-EFFORT-RPC-RED`
```text
command: go test ./internal/cli -run '^TestGPTAliasEffortRPCMatrix$' -count=1 -v
baseline: 4056f69e1c20d942d4f9fc7363d3d79bffde899a
verbatim:     gpt_alias_effort_contract_test.go:124: turn/start.effort="" want="max" alias=fable model=gpt-6-astra
exit: 1
log: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/01-effort-rpc.log sha256=97806a722b7921fe57bf108efc790837df7471601493d7eb423dd3989cbbf275
exit-carrier: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/01-effort-rpc.exit sha256=4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865
```

`EV-R0-APP-SERVER-RPC-RED`
```text
command: go test ./internal/gateway -run '^TestGPTProductionAppServerToolRoundTrip$' -count=1 -v
baseline: 4056f69e1c20d942d4f9fc7363d3d79bffde899a
verbatim:     gpt_production_appserver_roundtrip_contract_test.go:202: App Server claude-tool-result-continuation got=502 want=200
exit: 1
log: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/02-appserver-roundtrip.log sha256=a9a68cb73b7cac772d0da398ced18faa07150e9afde2ccf823bab17d3dc3d6c3
exit-carrier: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/02-appserver-roundtrip.exit sha256=4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865
```

`EV-R0-PRODUCTION-WIRING-RED`
```text
command: go test ./internal/cli -run '^TestGPTProductionAppServerWiring$' -count=1 -v
baseline: 4056f69e1c20d942d4f9fc7363d3d79bffde899a
verbatim:     gpt_appserver_factory_contract_test.go:111: managed App Server assembly touched absent/poisoned legacy auth store: handler=<nil> err=gateway private configuration or verified dependencies unavailable, want success with zero legacy open/read/refresh
exit: 1
log: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/02-production-wiring.log sha256=c6a107951d539950363c3e150544bc7a80bb73951cf0f093582ee09bf10a1905
exit-carrier: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/02-production-wiring.exit sha256=4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865
```

`EV-R0-FACTORY-RED`
```text
command: go test ./internal/cli -run '^TestGPTFactoryAndRetiredKanbanContract$' -count=1 -v
baseline: 4056f69e1c20d942d4f9fc7363d3d79bffde899a
verbatim:     gpt_appserver_factory_contract_test.go:277: gpt Factory backend env dispatch="" legacy-kanban="gpt", want gpt/empty
exit: 1
log: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/03-factory-retired.log sha256=6758afc04b04f1bcdd1a6e8215fbf76e2533a1f863723ed50ba48cbbd185b319
exit-carrier: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/03-factory-retired.exit sha256=4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865
```

`EV-R0-HISTORY-RED`
```text
command: go test ./internal/gateway ./internal/codexbridge -run '^TestGPTAppServerLifecycleAndHistoryAttribution$' -count=1 -v
baseline: 4056f69e1c20d942d4f9fc7363d3d79bffde899a
verbatim:     gpt_lifecycle_history_contract_test.go:196: production ServerConfig RejectionLogger seam available=false settable=false type=<nil>, want injectable structured rejection recorder
exit: 1
log: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/04-lifecycle-history.log sha256=51fd46a96c1a835fe2c389be01e3953e56dde47b1a3863b6b85e422b7162c0a6
exit-carrier: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/04-lifecycle-history.exit sha256=4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865
```

`EV-R0-NAMING-RED`
```text
command: go test ./internal/orchestration -run '^TestNamingMigrationManifestContract$' -count=1 -v
baseline: 4056f69e1c20d942d4f9fc7363d3d79bffde899a
verbatim:     naming_manifest_contract_test.go:71: schema_error=manifest: open naming-migration-manifest.json: open /Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop/.moai/specs/SPEC-MOAI-GATEWAY-001/naming-migration-manifest.json: no such file or directory
exit: 1
log: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/05-naming-manifest.log sha256=844a5d5e886635cc4576925764730f0ad039fc0425290eea246fd1e3b61e2c0a
exit-carrier: .moai/reports/SPEC-MOAI-GATEWAY-001/m14-r0/05-naming-manifest.exit sha256=4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865
```

| RED-now ID | 추적 / owner | fenced ledger | 현재 결과와 GREEN 종결 조건 |
|---|---|---|---|
| RB-GPT-ALIAS-ENV | AC-MG-001·003·011·018 / M14-R1 | `EV-R0-ALIAS-ENV-RED` | 21=6 PASS/15 RED. 21/21 PASS와 live AS-018이 필요하다. |
| RB-GPT-EFFORT-RPC | AC-MG-003·018 / M14-R1 | `EV-R0-EFFORT-RPC-RED` | 21=1 PASS/20 RED. 실제 `turn/start.model`과 `turn/start.effort` 21/21 PASS가 필요하다. |
| RB-APP-SERVER-RPC | AC-MG-004·007·009 / M14-R2 | `EV-R0-APP-SERVER-RPC-RED` | 8=5 PASS/3 RED. 한 batch 두 call, 역순 continuation 뒤 JSON-RPC `rpc-a`·`rpc-b` 각각 정확히 1회 응답해야 한다. |
| RB-APP-SERVER-WIRING | AC-MG-004·007·009 / M14-R2 | `EV-R0-PRODUCTION-WIRING-RED` | 6 RED. 네 managed route, concrete AppServerAdapter, absent/poisoned legacy auth store open/read/refresh 0, 두 번의 deterministic close가 모두 PASS해야 한다. |
| RB-GPT-FACTORY | AC-MG-014·AS-017·018 / M14-R3 | `EV-R0-FACTORY-RED` | 7 RED. stdout/stderr, 실행 전후 env, launch/Todo/Tasks/Dispatch/Factory 불변과 active prompt/env 전달을 전부 PASS한다. |
| RB-HISTORY-400-LOGGER | AC-MG-009·AS-010·012 / M14-R2 | `EV-R0-HISTORY-RED` | 11=6 PASS/5 RED. 네 400 CauseCode와 주입 가능한 structured rejection recorder를 모두 PASS한다. |
| RB-NAMING-MANIFEST | AC-MG-014(f) / M14-R3 | `EV-R0-NAMING-RED` | 5=3 PASS/2 RED. exact schema/current equality와 모든 mutation guard를 5/5 PASS한다. |

| 이미 GREEN인 회귀 가드 | 현재 증거 | 보존 조건 |
|---|---|---|
| App Server 정상/격리 5건 | [`02-appserver-roundtrip.log`](../../reports/SPEC-MOAI-GATEWAY-001/m14-r0/02-appserver-roundtrip.log) `a9a68cb73b7cac772d0da398ced18faa07150e9afde2ccf823bab17d3dc3d6c3` | initialize/thread/turn 및 duplicate/foreign/cross-thread 거절 5 PASS를 유지한다. |
| lifecycle/history 양성 6건 | [`04-lifecycle-history.log`](../../reports/SPEC-MOAI-GATEWAY-001/m14-r0/04-lifecycle-history.log) `51fd46a96c1a835fe2c389be01e3953e56dde47b1a3863b6b85e422b7162c0a6` | resume/model/compact/fork, 같은 Idempotency-Key 두 요청, 완료 prefix+새 사용자 입력 6 PASS를 유지한다. |
| portable process 6건 | [`06-portable-process.log`](../../reports/SPEC-MOAI-GATEWAY-001/m14-r0/06-portable-process.log) `c8bed0206b2e036149fcc104047cbb13d8a12af112171f6d7324250deb848608` | readiness/HTTP/overlay/0600/cancel/wait/exit 23의 6 PASS를 유지한다. |
| wiring 반복 안정성 | [`02-production-wiring-count20.log`](../../reports/SPEC-MOAI-GATEWAY-001/m14-r0/02-production-wiring-count20.log) `da1d39a798254c98f6a0f042ee05d9a11fad36c8d33e3dd5235b4cdc94f6f442` | 같은 제품 RED만 20회 재현하며 fixture race가 없어야 한다. GREEN 뒤 동일 반복은 exit 0이어야 한다. |

| 완료 의존성 — 현재 RED 주장 아님 | owner / 증거 경로 | 완료 조건 |
|---|---|---|
| 실제 Factory Agent/tool/approval/hooks/Tasks/Dispatch | M14-R5 / AS-017·018 | 실제 `moai gpt -f` lead·worker에서 전수 양성 증거와 digest를 남긴다. |
| GPT 실계정·PTY·history | M14-R5 / AS-001~015·019~021, t844 AS-010~012 | 각 AS owner가 원문·digest를 제출하고 t844의 live 양성을 소비한다. local contract PASS로 대체하지 않는다. |
| `CD-WINDOWS-BUILD` / `CD-WINDOWS-RUNTIME` | M14-R4·R5 / AS-016·022 | `windows-latest` build와 실제 종료·재개·flush·권한 사건을 각각 PASS한다. portable 6 PASS로 대체하지 않는다. |
| rc 종결 | M14-R5 / AS-022 | 필수 live gate 뒤 clean local rc 설치·version·binary SHA·상태 readback을 PASS한다. |

- Gap으로 남을 수 있는 비-release-blocking 항목은 완료 보고의 Gaps 구획에 이유와 함께 열거한다.
- 변경 패키지의 `go vet`, `golangci-lint run`, `go test`가 통과했다.
- 새 코드 경로 커버리지 85% 이상.
- 어떤 AC도 실행하지 않은 명령의 출력을 근거로 인용하지 않았다.

## E. 이관된 시험과 게이트 측정

| 시험 | 추적 | 처리 |
|---|---|---|
| T02 3사 picker 항목 집합 | — | 0.6.0에서 `SPEC-MOAI-GATEWAY-PICKER-001` (제안)으로 이관. `AC-MG-002`는 폐기 묘비 |
| T05 PKCE state/취소/timeout | — | `SPEC-MOAI-GPT-AUTH-001` (제안)으로 이관 |
| T06 GPT logout 소유권 경계 | — | `SPEC-MOAI-GPT-AUTH-001` (제안)으로 이관 |
| T07 GPT 모델별 접근 권한 | — | `SPEC-MOAI-GPT-AUTH-001` (제안)으로 이관. 이 SPEC은 등록 ID 정확성만 `AC-MG-021`로 검증 |
| T08 구독/API 과금 경계 | (REQ-MG-022 → AC-MG-022, AC-MG-018) | 실계정 과금 판정은 이관. fallback 부재와 settings 우회 차단 의무만 이 SPEC에 남음 |
| T09 Claude 구독 OAuth × 로컬 gateway 인증 공존 | (REQ-MG-016 → AC-MG-021) | **이관하지 않음.** 선행 측정으로 수행하고(`plan.md` M0) `REQ-MG-016`을 게이트한다 |
| T10 외부 auth 보존 | (REQ-MG-025 → AC-MG-020) | `SPEC-MOAI-GPT-AUTH-001` (제안)으로 이관. Codex 쪽 불변식만 이 SPEC에 보존 |
| T20 `cg`/`gg` 제거와 기존 회귀 | (REQ-MG-004 → AC-MG-015) | `SPEC-MOAI-CG-RETIRE-001` (제안)으로 이관. `gg` 부재만 이 SPEC에 보존 |
| T21 AS-010·011·012 실세션 양성 실증 | (REQ-MG-015, REQ-MG-017 → AC-MG-009/AS-010, AC-MG-003/AS-011, AC-MG-009/AS-012) | 실세션 양성 실증(실세션 회상·실제 turn model 일치·실제 Claude 압축 수집)은 카드 t844가 실행 owner다. 자동 거절 변형군과 기계적 검증은 카드 t653 run 커밋 `14dba89c5`·`e45f50a8d`에서 착지했지만, M14-R5가 t844 원문 evidence와 digest를 이 SPEC에 소비하기 전까지 전체 기능 통과는 보류다. AS-013은 t844 이관 대상이 아니며 M14-R5가 설치본 native fork를 실제 실행한다. 0.13.0의 NOT-RUN 기록은 역사적 상태이고 0.16.0 완료 기준으로 사용할 수 없다 |
| T22 AS-019와 T21의 경계 정합 (0.13.0, t654) | (REQ-MG-019, REQ-MG-015 → AC-MG-001/AS-019) | 카드 t654 문구의 "실제 Claude PTY … 재개·모델전환"은 launcher 측(Claude/GLM 재개·전환과 provider 경계)으로 판정한다. AS-010·011·012의 GPT thread 실세션 양성은 T21대로 t844에 유지되며 AS-019가 흡수·대체하지 않는다 — 두 표면이 같은 "재개·모델전환" 어휘를 쓰지만 판정 대상이 다르다. t844 이관 내용이 바뀌지 않는 한 T21 행은 이 판에서 수정하지 않는다 |
| tmux pane teammate 표면 (옛 `AC-MG-006` teammate 수명 판정, 옛 `AC-MG-018` (a) tmux 주입 판정) | — | 0.6.0에서 `SPEC-MOAI-GATEWAY-TEAMMATE-001` (제안)으로 이관. tmux 세션 env 무기록과 in-process 표시 판정만 이 SPEC에 남음 |

## 0.9.0 기존 AC의 형식·인증 대조군 보강

AC-MG-007·008: Given 동일한 완료 text/tool SSE에 media type이 있는 응답과 없는 구독 응답이 있을 때, When 중계하면,
Then 공개 출력·index·ID·종료는 같아야 한다. 완료 item이 있는 빈 completed.output과 일치하는 비어 있지 않은 snapshot은
양성이다. 헤더 없음+HTML/임의 JSON, 충돌 media type, 중복 JSON field·done·terminal, ID/index 충돌, 모순 snapshot,
미완료 item·failed·다른 incomplete·truncated EOF·취소는 성공 terminal 0이다. 기존 max-output 제한 대조군을 삭제하지 않는다.

AC-MG-021: Given research §19의 양성 M0 근거와 제품 gateway가 있을 때, When Opus 5·Sonnet 5 요청과 인증 실패 뒤
재요청을 실행하면, Then 별도 세션 헤더의 로컬 검증·upstream 수신 0과 허용 OAuth/beta 보존을 각각 확인한다.
잘못되거나 중복인 세션 헤더는 외부 송신 0이다. 직접 refresh 발급 출처는 POST 응답과 같은 본문 재전송의 Bearer hash
상관관계로 판정하며 changed-Bearer-only 관측으로 대체하지 않는다. 선행 refresh와 제품 통합의 시험 귀속을 분리한다.

## 0.9.0 receipt 실험의 현재 적용 경계

직접 Responses 경로의 receipt 승인·실험은 과거 기록으로 보존한다. 0.11.0 GPT 구독 경로는 Codex thread 이력을
사용하므로 raw carrier 게시·복원 시험을 제품 활성화 기준으로 적용하지 않는다. 현재 귀속·유실·중복·복구 판정은
AC-MG-009와 아래 t650~t654 시나리오다.

## 승인된 출력·native 정책의 App Server 적용

Given 구독/API 모드와 Claude 출력 상한이 있을 때, When 각각 App Server로 요청하면, Then 두 모드 모두 서버 출력
정책을 사용함을 표시하고 Claude 생성 토큰 상한을 동일하게 적용했다고 표시하지 않는다. MoAI 바이트·취소 제한은
별도로 동작하며 생성 토큰 수 제한과 구별된다. API 모드는 명시 API 과금을 유지하고 구독 실패로 자동 선택되지 않는다.
입력 max_tokens의 누락·0·음수·소수·문자열·null·중복·범위 초과는 기존 검증대로 거절한다. 바이트·취소 한도는
생성 토큰 한도와 같다고 표시하지 않는다. 해당 실제 경로의 검증은 t650·t654 AC다.

Given Claude/GPT의 명시 medium/high·제목 schema·일반 main·서브에이전트 요청이 있을 때, When adapter를 거치면,
Then 지원된 의미가 해당 경로에 보존된다. 노력 수준의 공급자 간 계산량 동등성을 주장하지 않는다.
미지원 policy·타입·중복·schema 의미 손실은 명시 거절하며 정책이나 도구를 빼고 성공시키지 않는다.
Claude native signature 보존은 Claude adapter에서 유지하고 GPT reasoning은 App Server thread 소유권으로 판정한다.
과거 직접 Responses body·receipt 형식은 새 경로의 성공 판정이 아니다. 실제 title/main/subagent 성공은 각각 분리한다.
metadata context, 경로 유효 한도와 실제 큰 입력 수용을 구분하고 Windows 실행 게이트를 유지한다.

## t649 제공자별 실제 제품 판정 — 기존 AC 보강

**AC-MG-001 / AC-MG-003 (REQ-MG-019).** Given 격리된 세 제공자 프로필과 실행 전 공유 사용자·프로젝트
설정의 해시가 있을 때, When 실제 `moai cc`·`moai glm`·`moai gpt` PTY에서 `/model`을 열고 Default·각 행·
직접 입력·Enter 저장·`s` 선택을 수행하면, Then 표시되고 선택 가능한 모델은 해당 제공자뿐이며 Default와
현재 행도 같은 제공자로 해석된다. GPT 목록의 GPT-6 Astra 선택은 실제 요청 `gpt-6-astra`로 확인한다.
새 세션·다른 launcher·일반 Claude를 다시 열어도 다른 제공자의 저장 선택이 전파되지 않고 공유 설정 해시는 같다.
managed 충돌·빈 목록은 명시 실패로 확인하며 내장 타 제공자 목록으로 묵시 복귀하면 실패다.

**AC-MG-018 / AC-MG-025 (REQ-MG-019, REQ-MG-021).** Given 다른 제공자의 상속 env·저장 기본값이
있는 세 launcher에서, When 기본 요청·네 tier 서브에이전트·fallback을 각각 실행하면, Then 전송 ID는 해당
세션 catalog에 속하며 타 upstream 계수는 0이다. 임의 접미사나 미등록 ID는 자동 확장 없이 거절한다.
모델 설정을 위해 공유 설정에 값을 쓰거나 삭제하면 실패다. 기존 GLM 오염 정리와 인증 불변식은 유지한다.

검증 가능성: 제품 PTY·실제 계정·tool·resume·음성 시험은 각각 독립 판정한다. 아직 증거가 없는 항목은 Pending/Gap이며
직접 Claude 임시 UI 프로브, 단위 시험, 로그인 상태로 대체하지 않는다. Windows 실행은 GitHub CI 결과를 별도로 기록한다.

## 0.11.0 App Server 현재 실행 게이트 — t650~t654

다음은 기존 AC의 하위 시나리오이며 새 기본 AC 번호가 아니다. 각 카드의 모든 양성·음성 판정을 충족해야 한다.
범위가 미검증이면 오류로 막는 것은 구현 중 안전 경계일 뿐 기능 완료가 아니다.

### t650 기반과 인증

**AC-MG-001 / AS-001 (REQ-MG-015, REQ-MG-017)** Given 설치 Codex의 schema와 기능 협상이 있을 때, When App Server를 시작하고 initialize·thread/start를
수행하면, Then 동적 도구 capability를 확인하고 unsupported 버전은 명시 오류로 끝난다. 자동 설치나 direct backend 우회는 없다.

**AC-MG-020 / AS-002 (REQ-MG-015, REQ-MG-017)** Given 구독 managed 계정과 API 키 전용 프로필을 각각 준비했을 때, When 로그인·상태·생성·갱신·로그아웃을
수행하면, Then 선택한 auth 모드로만 처리된다. 구독에서 MoAI의 토큰 파일 읽기·refresh·direct backend 요청은 0이고,
구독 만료/한도 오류가 API 과금 경로로 넘어가지 않는다. 같은 프로필을 경쟁 로그인으로 덮어쓰지 않는다.

**AC-MG-022 / AS-003 (REQ-MG-015, REQ-MG-017)** Given Codex native 도구가 가능한 기본 환경과 MoAI 제한 환경이 있을 때, When shell·file·MCP·agent 실행을
각각 유도하면, Then 제한 환경의 native 실행 계수는 0이고 실제 Claude 도구만 호출된다. 모델 도구 inventory도 대조한다.
approval never·readonly 또는 단순 설정 파일 존재로 이 AC를 통과시키지 않는다.

### t651 도구 실행권과 ToolSearch

**AC-MG-004 / AS-004 (REQ-MG-015, REQ-MG-017)** Given 초기 완전 정의 도구와 ToolSearch 뒤 처음 발견하는
도구가 있는 실제 Claude 세션에서, When hybrid를 실행하면, Then 초기 도구는 native schema로, 후발 도구만
dispatcher로 연결되고 ToolSearch는 유지된다. 두 경로 모두 실제 Claude에서 정확히 한 번 실행된다.
기본 최초 전체 schema 전제가 반증된 기록과 비활성 비교는 과거 판단 근거이며 현재 채택은 hybrid다.

AS-004의 여섯 음성군은 각각 양성 대조군과 실행 계수로 판정한다.
1. Given native·dispatcher·예약 namespace의 이름 충돌 또는 한 요청 안의 중복 정의, When 등록하면, Then 실행 0과 명시 오류다.
2. Given tool_reference만 있고 전체 schema가 없거나 미등록 이름, When 호출하면, Then 실행 0이다.
3. Given 기존 이름 재정의나 발견 후/pending 중 schema 변경, When 결과를 소비하면, Then 새 schema를 옛 호출에
   적용하지 않고 명시 거절하며 도구를 재실행하지 않는다.
4. Given required·타입·enum·중첩·additionalProperties 위반 또는 미지원 dialect/$ref, When 호출하면, Then 실행 0이다.
5. Given 다른 대화 참조·중복 결과·thread/turn/call 교환, When 처리하면, Then 다른 호출을 소비하지 않고 실행 0이다.
6. Given 늦게 발견한 유효 도구, When 정상 왕복하면, Then 기존 thread·turn이 유지되고 등록 때문에 추가된
   thread/start·fork·resume 호출은 0이다. 초기 native 도구의 dispatcher 우회와 dispatcher 재귀도 실행 0이다.

**AC-MG-007 / AS-005 (REQ-MG-015, REQ-MG-017)** Given App Server dynamic call이 발생했을 때, When Claude가 tool_use를 받아 승인·실행하고 다음 HTTP로
tool_result를 보내면, Then App Server의 같은 thread/turn/call에 결과가 한 번만 전달되고 최종 답변이 Claude 화면에 보인다.
텍스트·로컬 이미지·도구 실패를 각각 판정하며 unsupported 결과는 명시 오류다.

**AC-MG-013 / AS-006 (REQ-MG-015, REQ-MG-017)** Given ToolSearch가 `tool_reference`와 설명을 반환하는 실제 Claude 요청이 있을 때, When 검색 결과를
처리하고 발견된 도구를 호출하면, Then 400 없이 타입·이름·등록 schema가 검증되고 도구가 Claude에서 정확히 한 번 실행된다.
잘못된 tool_reference·미등록 이름·한 요청 안의 중복 참조·schema 변경·다른 대화의 참조는 송신 전 거절한다.
Given 앞선 요청에서 이미 등록한 도구와 같은 schema가 있을 때, When 이후 요청의 정상 ToolSearch가 그 도구를
다시 발견하면, Then 재발견은 멱등 처리되고 발견 epoch·schema digest·기존 native/dispatcher route가 변하지 않는다.
반복 검색 자체를 오류로 거절하거나 새 등록·추가 도구 실행으로 세면 실패다.
기존 textContent/receipt projection 거절 보고는 회귀 픽스처의 유래이며, 최신 partial fix 상태는 실행으로 재확인한다.

### t652 수명·권한·복구

**AC-MG-009 / AS-007 (REQ-MG-015, REQ-MG-017)** Given 도구 대기 때문에 첫 HTTP 응답이 끝난 turn이 있을 때, When 다음 HTTP 결과 요청이 오면,
Then App Server process와 pending RPC가 살아 있고 원 turn을 잇는다. 동시 대화 A/B의 결과를 바꾸면 둘 다 서로의 call을
소비하지 않는다. 같은 결과 재전송은 도구 재실행 없이 일관된 처리 또는 명시 중복 오류다.

**AC-MG-009 / AS-008 (REQ-MG-015, REQ-MG-017)** Given pending call의 실행 전·실행 후·결과 저장 후 각 지점에서 프로세스를 강제 종료했을 때,
When 재개하면, Then 실행 여부가 불명확한 도구를 자동 재실행하지 않는다. 복구 불가능한 active RPC ID는 명시 실패와
복구 안내를 제공한다. 성공 상태를 합성하거나 새 process에 옛 RPC 응답을 주입하지 않는다.

**AC-MG-008 / AS-009 (REQ-MG-015, REQ-MG-017)** Given 활동 중 turn 또는 병렬 tool call이 있을 때, When 사용자가 취소하거나 연결이 끊기면,
Then turn/interrupt와 대기 호출 정리가 해당 turn에만 적용되고 뒤늦은 결과가 다른 turn을 진행시키지 않는다.
EOF·RPC 오류·출력 상한 초과 때 성공 Messages terminal이 없어야 한다.

### t653 이력·모델·압축·fork

**AC-MG-009 / AS-010 (REQ-MG-015, REQ-MG-017)** Given GPT 대화를 종료하고 새 MoAI process로 시작했을 때, When 소유 thread ID로 resume하면,
Then 합성 사실·도구 결과가 유지된 정상 답변이 나온다. 다른 계정·다른 대화·변조 mapping은 거절한다.
thread/resume.history 또는 raw reasoning을 직접 가져와 재구축한 결과로 통과할 수 없다.

**AC-MG-003 / AS-011 (REQ-MG-015, REQ-MG-017)** Given 같은 GPT thread가 idle이고 허용 모델이 있을 때, When `/model`로 GPT-6 Astra와 5.6을 각각
선택하면, Then 실제 turn model이 선택 ID와 일치하며 App Server 경고·실패가 명시된다. 다른 제공자와 bare gpt-6는
거절된다. active tool 대기 중 전환은 잘못된 turn에 적용되지 않는다. family 호환 성공은 실제 계정별 별도 증거가 필요하다.

**AC-MG-009 / AS-012 (REQ-MG-015, REQ-MG-017)** Given 실제 Claude 압축 요청과 그 응답·후행 hook을 수집할 때,
When 수동 print·대화형·자동·자식 압축을 각각 수행하면, Then 정상 App Server turn의 실제 공개 요약이 Claude에
표시되고 PostCompact의 exact summary digest 및 인증 scope/epoch와 일치해야 한다. PreCompact로 다음 HTTP를
분류하지 않으며 MoAI의 추가 thread/compact/start 호출은 0회다. 이미 반영한 입력과 완료 요청은 중복 생성하지 않는다.
검증된 후속 wrapper에서 새 입력만 반영하고 압축 전 합성 사실·도구 결과를 실제 답변에서 회상하며 새 프로세스
resume 뒤에도 회상한다. 압축 뒤 ToolSearch로 후발 도구를 사용할 수 있어야 한다.
같은 scope의 일반 요청·재시도·중복/지연/유실 PostCompact·HTTP 응답 유실·재시작을 주입하면 한 번만 rebase하고
모호한 상태에서는 명시 오류이며 외부 생성은 0회다. foreign/변조/오래된 epoch·summary substring만 맞춘 일반 입력은
거절한다. 인증된 통지와 완료 응답이 없는 상태에서 요약 wrapper를 임의 채택하지 않는다. 각 경로 증거를 구분하고
단순 구조 probe나 안전한 거절만으로 전체 기능을 통과 처리하지 않는다.

**AC-MG-009 / AS-013 (REQ-MG-015, REQ-MG-017)** Given 실제 Claude non-fork Agent 둘과 중첩 일반 자식이 있을 때,
When 자식별 첫 context와 후속 ToolSearch·도구 결과·취소를 처리하면, Then 각 agent-id의 독립 thread에 native child
context만 최초 한 번 반영하고 부모·형제의 후속 입력이나 취소를 섞지 않는다. 도구 실행은 Claude가 담당한다.
Given 원본 family의 완료 prefix와 completedTurnID 원장이 있을 때, When launcher의 --fork-session으로 분기하면,
Then 공식 lastTurnId로 그 완료 위치까지 분기하고 새 family/thread에 원본의 분기 전 사실을 보존한다. 부모가 뒤의
turn을 완료한 경우에도 지정 경계 이후 사실이 자식에 유입되지 않아야 한다. 새 프로세스로 자식을 resume해도 유지된다.
다른 계정·미완료 경계·변조 prefix/원장·알 수 없는 원본은 외부 분기 전에 거절한다. 일반 자식 성공은 명시 분기 성공을
대체하지 않는다.
Given 설치본의 native Agent(fork)/subtask 가용성을 확인한 실제 Claude 경로가 있을 때, When 부모 이력에서
자식을 분기하고 병렬·중첩·부모의 후속 진행·자식 resume를 수행하면, Then 분기 전 사실이 보존되고 분기 이후의
부모·형제 입력과 결과·취소가 섞이지 않아야 한다. 다른 계정·변조된 귀속·불명확한 분기 위치는 실행 전에 거절한다.
설치본에서 native fork를 찾지 못하면 전제 실패 NOT-RUN으로 기록하고 AS4 및 전체 지원 완료를 보류한다.
이 양성 의무는 유지되며 일반 non-fork 또는 --fork-session 성공으로 대체하지 않는다. 가용성 조사 및 native fork
증거 확보와 독립적으로 다른 AS4 경로의 구현·검증은 진행한다.

### t654 제품 검증

**AC-MG-001 / AS-014 (REQ-MG-015, REQ-MG-017)** Given 세 MoAI launcher와 오염된 전역 모델 기본값이 있을 때, When 실제 PTY `/model`의 Default·현재 행·
목록·직접 입력·Enter·s·재개를 수행하면, Then cc는 Claude, glm은 GLM, gpt는 GPT만 선택·요청하며 공유 설정이 변하지 않는다.
구독과 API 인증 표시는 실제 선택과 일치해야 한다. Codex TUI를 열어 성공한 것은 이 AC의 증거가 아니다.

**AC-MG-010 / AS-015 (REQ-MG-015, REQ-MG-017)** Given 설치 App Server model metadata 및 토큰 사용량이 있을 때, When 한도 경계 안팎의 실제 입력을
보내면, Then 현재 경로 한도와 오류를 정확히 표시하고 direct endpoint 921k나 미검증 1M을 수용 보장으로 쓰지 않는다.
872k metadata도 실제 수용 시험과 별개로 표기한다. 구독 출력은 서버 정책이며 local 취소·출력 바이트 한도는 동작한다.

**AC-MG-006 / AS-016 (REQ-MG-015, REQ-MG-017)** Given 기존 review RPC 소비자와 Windows GitHub CI가 있을 때, When 공통 RPC 추출의 회귀 및
Windows process 종료·재개·파일 flush/원자교체·권한 시험을 실행하면, Then 기존 review 의미와 승인된 Windows API 계약이 유지된다.
Windows cross compile만으로 runtime AC를 통과시키지 않는다.

**AC-MG-004 / AS-017 (REQ-MG-015, REQ-MG-017)** Given 실제 launcher PTY 세션과 초기 완전 정의 도구, ToolSearch 뒤 처음 발견하는
도구가 있을 때, When 실제 turn에서 초기 도구와 후발 도구를 각각 실행하면, Then 초기 도구는 native schema로, 후발 도구는 dispatcher로
정확히 한 번 실행되고 등록 때문에 추가된 thread/start·fork·resume 호출은 0이다. t651의 fake 기반 AS-004를 대체하지 않고 제품 판정을
더한다. 증거: `.moai/reports/t654/as5-pty-toolsearch.log` + gateway 요청 기록. 검증 가능성: 실제 계정 권한 필요 — 불가하면 Gap.

**AC-MG-011 / AS-018 (REQ-MG-019, REQ-MG-015)** Given 세 launcher PTY 세션과 서브에이전트의 네 슬롯 별칭 요청을 유도하는 작업이 있을
때, When 서브에이전트가 별칭·모델로 요청을 보내면, Then 요청은 해당 제공자 세션 catalog 안에서 해석되고 타 upstream 요청 계수는 0이며,
미등록 ID는 자동 확장 없이 거절된다. GPT 세션에서는 `fable`·`opus`·`sonnet`·`haiku`가 각각
`gpt-6-astra`·`gpt-5.6-sol`·`gpt-5.6-terra`·`gpt-5.6-luna` 요청으로 관측되고, 각 요청의
effort는 별칭과 독립적으로 지정값 또는 누락을 보존한다. 증거: `.moai/reports/t654/as5-pty-subagent.log`.
검증 가능성: 실제 PTY 필요.

**AC-MG-001 / AS-019 (REQ-MG-019, REQ-MG-015)** Given 같은 제공자의 세션 기록이 있을 때, When launcher 재개 플로우와 `/model` 전환을
실제 PTY에서 수행하면, Then 재개는 같은 제공자·대화 소유권이 확인된 기록에 한해 수용되고, 전환 뒤 실제 turn 요청의 `model`이 선택 ID와
일치하며 타 provider 요청 계수는 0이다. **t844 경계**: AS-010(GPT thread 실세션 회상)·AS-011(GPT 실제 turn model 일치)·AS-012(실제
Claude 압축 수집)의 실세션 양성은 T21대로 t844 소관이며, 이 AS는 그것을 흡수·대체하지 않는다 — 이 AS는 launcher 측(Claude/GLM 재개·
전환과 provider 경계)과 GPT 측 picker·인증 표시(AS-014)까지만 판정한다. 증거: `.moai/reports/t654/as5-pty-resume-model.log`. 검증
가능성: 실제 PTY 필요.

**AC-MG-010 / AS-020 (REQ-MG-011, REQ-MG-014)** Given 설치 App Server model metadata와 provider별 capability 선언이 있을 때, When
GLM 세션에서 이미지 입력과 Claude 세션에서 이미지 입력을 각각 시도하면, Then GLM은 text-only로 명시 거절하고 Claude는 선언 capability대로
수용하며, 표시는 모델 명목 창·현재 경로 유효 한도·누적 사용량을 구분한다. `gateway_product_binding.go`의 공통
`Capabilities{ContextTokens: 1000000, Images: true}`는 provider별 양성·음성 시험으로 재판정한다 — 그 선언 자체는 수용 보장이 아니다.
증거: `.moai/reports/t654/as5-context-paths.md`. 검증 가능성: 부분 — capability 음성·양성은 mock/합성 입력으로 결정적, 실제 대형 입력
수용은 실계정 권한이 필요하므로 Gap 가능.

**AC-MG-020 / AS-021 (REQ-MG-017, REQ-MG-025)** Given 구독 managed 계정과 API 키 전용 프로필이 있을 때, When 두 모드를 실제 PTY에서
각각 선택·실행하면, Then 표시된 인증 방식이 실제 선택과 일치하고, MoAI의 구독 토큰 파일 접근은 0이며, 구독 실패·만료 유도 시 API 과금
경로의 요청 계수는 0이다(자동전환 금지). 두 모드 모두 공식 App Server 출력 정책을 사용함을 표시하고 Claude 생성 토큰 상한과 같다고
표시하지 않는다. 증거: `.moai/reports/t654/as5-auth-modes.md`. 검증 가능성: 실제 계정 권한 필요.

**AC-MG-006 / AS-022 (REQ-MG-009, REQ-MG-015)** Given supervisor·launcher의 이름을 정한 Windows 시험과 `release-pr-multi-os.yml`의
`workflow_dispatch` 트리거가 있을 때, When GitHub CI windows-latest 레그를 실행하면, Then `test-stream-release-verify-windows-latest`
아티팩트에서 이름을 정한 시험 각각의 종료 이벤트가 `"Action":"pass"`임을 확인하고 결과를
`.moai/reports/t654/as5-windows-ci-verdict.md`에 기록한다. 시험 부재·skip·아티팩트 부재는 PASS가 아니다. Windows cross compile exit 0만으로
이 AS를 PASS로 세지 않는다. 검증 가능성: GitHub CI 실행 필요(로컬 darwin에서 불가 — 원격 실행 또는 창 대기).

### AS-001~AS-022 실행·증거 책임표 (0.16.0)

아래 표는 기존 Given-When-Then을 바꾸지 않고 누가 어떤 실행으로 닫는지를 고정한다. 모든 경로는 실행 명령,
원문 출력, 기준 HEAD, exit 또는 event, artifact SHA-256을 함께 기록한다. `future` 경로는 M14-R5에서 실제
파일이 생겨야 하며, 파일명만 예약된 상태는 PASS가 아니다.

| AS | 구현/기계 gate | live 실행과 evidence | 의존·종결 규칙 |
|---|---|---|---|
| AS-001 | M14-R2 `TestGPTProductionAppServerWiring` 6/6 GREEN의 managed adapter·initialize·legacy store 0-use·두 번 close | 설치본 schema/initialize/thread-start 원문 → `.moai/reports/t654/as5-appserver-capability.log` | unsupported 음성 포함; direct fallback 0 |
| AS-002 | M14-R2 authority/profile 단위·통합 시험과 `TestGPTProductionAppServerWiring` | managed 구독/API-key login·status·generation·logout → `.moai/reports/t654/as5-auth-modes.md` | profile별 계수; token read/refresh/direct/API fallback 0 |
| AS-003 | M14-R2 App Server 시작 config의 native 도구 차단 회귀 | shell/file/MCP/agent 유도와 Claude tool 양성 recorder → `.moai/reports/t654/as5-native-tool-negative.log` | native 실행 0과 Claude 실행 >0을 함께 요구 |
| AS-004 | 기존 t651 hybrid selector 전수 + M14-R2 multi-call 회귀 | 실제 PTY 초기 native schema/후발 dispatcher → `.moai/reports/t654/as5-pty-toolsearch.log` | 여섯 음성군 각각 양성 대조군 필요 |
| AS-005 | `TestGPTProductionAppServerToolRoundTrip` 8/8 GREEN | approval/hook 뒤 text·image·실패 tool_result와 final 화면 → `as5-pty-tool-roundtrip.log` | 첫 batch 두 call·역순·exact `rpc-a/rpc-b` 포함 |
| AS-006 | 기존 t651 ToolSearch schema/epoch selector | 실제 tool_reference 발견·재발견 → `as5-pty-toolsearch.log` | 400 0, 실행 exact-once, 추가 thread RPC 0 |
| AS-007 | M14-R2 pending batch/duplicate/foreign/cross-thread 통합 시험 | HTTP 사이 process 생존과 동시 A/B → `as5-pending-lifetime.log` | `TestGPTProductionAppServerToolRoundTrip` GREEN 선행 |
| AS-008 | 기존 t652 crash barrier/receipt 복구 selector | 실행 전·후·저장 후 강제종료 → `as5-crash-recovery.log` | 불명확 자동 재실행 0, 옛 RPC 주입 0 |
| AS-009 | cancel/EOF/RPC/output-limit 통합 시험 + portable cancel guard | 실제 PTY cancel/disconnect → `as5-cancel-isolation.log` | 다른 turn 진행 0, 성공 terminal 0 |
| AS-010 | `TestGPTAppServerLifecycleAndHistoryAttribution` 11/11의 resume·두 retry·새 입력·네 400/logger GREEN | **t844** 실세션 회상 evidence를 `.moai/reports/t844/`에서 읽어 `.moai/reports/t654/as5-t844-consumption.md`에 digest/index | t844 실제 회상 PASS 없이는 완료 차단 |
| AS-011 | 같은 selector의 `model` + alias env/RPC 각 21/21 GREEN | **t844** 실제 turn model 및 PTY picker → `as5-t844-consumption.md`, `as5-pty-subagent.log` | 선택 ID와 App Server turn model exact equality |
| AS-012 | 같은 selector의 `compact`(추가 compact RPC 0) + rebase 음성 | **t844** 수동/대화형/자동/자식 압축·회상 → `as5-t844-consumption.md` | t844 실제 summary/PostCompact digest 없이는 완료 차단 |
| AS-013 | 같은 selector의 durable `fork` 및 ownership 음성 | 설치본 native Agent(fork)/subtask 병렬·중첩·resume → `.moai/reports/t654/as5-native-fork.log` | 일반 child/`--fork-session`으로 대체 금지; NOT-RUN은 완료 차단 |
| AS-014 | provider catalog/env/settings 회귀 + alias matrix | 세 launcher 실제 `/model` 전수 → `.moai/reports/t654/as5-pty-model-picker.log` | 공유 설정 hash 전후 동일, provider 격리 |
| AS-015 | model metadata/limit/error 통합 시험 | 한도 안팎 실제 입력 → `.moai/reports/t654/as5-context-limits.md` | 명목 metadata와 실제 수용을 분리 |
| AS-016 | portable 6/6 회귀 + review RPC 회귀 | Windows named runtime tests → `.moai/reports/t654/as5-windows-ci-verdict.md` | build-only 금지; runtime pass events 필요 |
| AS-017 | M14-R2 RPC 8/8 + M14-R3 Factory prompt/env GREEN | 실제 Factory Agent가 초기/후발 tool을 실행하고 approval/hook/Dispatch recorder 일치 → `as5-factory-agent-tool.log` | 단위 prompt-forwarding으로 대체 금지; Tasks/Dispatch 양성 필요 |
| AS-018 | `TestGPTAliasEffortMatrix`와 `TestGPTAliasEffortRPCMatrix` 각각 21/21 + Factory `alias-effort` GREEN | 실제 서브에이전트 네 별칭×effort/누락 → `as5-pty-subagent.log` | provider/model/`turn/start.effort` exact equality |
| AS-019 | lifecycle model/resume regression | 세 launcher PTY resume/model → `as5-pty-resume-model.log` | launcher 경계만 소유; GPT 실세션 양성은 AS-010·011/t844를 별도 소비 |
| AS-020 | provider capability mock 양성·음성 | Claude image 양성/GLM text-only 거절/표시 → `as5-context-paths.md` | 공통 1M 선언만으로 PASS 금지 |
| AS-021 | managed authority/profile/fallback 음성 | 구독/API-key 실제 PTY → `as5-auth-modes.md` | direct token read/refresh/backend/API 자동전환 모두 0 |
| AS-022 | portable 6/6과 Windows compile 사전 gate | GitHub `windows-latest` named build+runtime artifact → `as5-windows-ci-verdict.md` | skip/부재/compile-only는 실패 |

`t844`는 별도 카드라는 이유로 이 SPEC의 완료 판정에서 빠지지 않는다. M14-R5는 t844의 AS-010·011·012
원문 carrier를 실제로 읽고 SHA-256과 해당 PASS assertion을 `as5-t844-consumption.md`에 색인해야 한다.
증거가 없거나 현재 구현과 기준 SHA가 호환되지 않으면 그 AS를 다시 실행하며, `Gap`으로 완료하지 않는다.


완료 판정: t650→t651→t652→t653→t654 의존 순서로 실제 증거를 기록하고 독립 감사를 통과한다.
단일 App Server moai_echo 프로브, account/read 또는 compile 성공으로 실제 Claude PTY·tool·resume·fork·compact
전체를 완료하지 않는다. Windows는 GitHub CI의 실행 증거를 요구한다. t654의 종결 판정은 AS-014~AS-022와
`AC-MG-026` (d)의 rc 로컬 배포 게이트다 — 배포 게이트는 검증 전수 PASS를 전제로 하며, push·PR·병합·
워크트리 제거는 포함하지 않는다.

### AS-002·AS-007의 관리 세션 권한 대조군 (REQ-MG-023)

Given App Server profile·account·auth mode·generation·allowed model이 일치하는 로컬 권한 snapshot이 있을 때,
When 선택 검증 요청을 처리하면, Then 최소 성공 응답이며 생성·refresh·외부 요청은 0이다. 직접 credential Apply와
토큰 파일 읽기도 0이다. 각 필드를 하나씩 바꾼 음성군은 401, 세션 catalog 밖 ID는 404이고 실행 0이다.
Given 선택 검증 뒤 logout·계정 변경으로 generation이 달라졌을 때, When 실제 turn을 시작하면, Then 이전 권한으로
진행하지 않는다. 직접 API·GLM credential 부재/세대 변경의 기존 음성군은 계속 통과해야 한다.
