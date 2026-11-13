# Pi v1.5.0 pin spike (#2199)

상위: #2158 (EP-F). 근거: 설계 §14, decisions 00:5x(11-09) "§14 Pi — 고정 버전 spike 1장(conformance·보안·취소·인증) 뒤 채택 결정". 측정일 2026-11-09(UTC 02:4x). 채택 여부는 사용자가 정한다. 이 문서는 측정과 권고만 담는다.

## 측정 대상과 환경 (REQ-F1-01)

| 항목 | 값 |
| --- | --- |
| 패키지 | `@earendil-works/pi-coding-agent@1.5.0` (npm, 2026-11-02T11:05:09Z 게시, license 필드 MIT, repository `github.com/earendil-works/pi`) |
| tarball 무결성 | `npm pack` 결과 integrity `sha512-xNDZ5kqowO6Z03FS2gwAkPEGk1ssxVcYKQ6qZId84mFEvwACvvyc2l0vAMOS+c5zvxLRMLBE9r4y44IGYkMJaQ==`가 registry `dist.integrity`와 같다. tarball sha256 `5bfb66848a4c4b81c637ce8d2136be01945b70aac8e8556b091cb02241b385df` |
| 설치 | `npm install --ignore-scripts --save-exact @earendil-works/pi-coding-agent@1.5.0` (git이 무시하는 `target/` 아래 별도 디렉터리). 123 패키지 |
| 런타임 | Node v24.20.0 / npm 11.19.0 (macOS arm64). `engines`는 `node >=22.19.0` |
| 격리 | 허용 목록 환경만 전달: `PATH`(node bin + `/usr/bin:/bin`), `HOME`·`PI_CODING_AGENT_DIR`(작업 디렉터리 안 임시), `PI_OFFLINE=1`, `PI_SKIP_VERSION_CHECK=1`, `PI_TELEMETRY=0`, `LANG`, 상속 확인용 `SPIKE_CANARY`. 공급자 키 변수는 하나도 전달하지 않았다 |
| 모델 | `127.0.0.1` 임시 포트의 OpenAI 호환 mock(`api: openai-completions`, `models.json` custom provider). 호스팅 모델 호출 0, 키 발급 0 |
| 문서 근거 | 고정 tarball에 포함된 문서(아래 sha256). `main` 브랜치 문서가 아니라 1.5.0과 함께 배포된 사본이다 |

문서 sha256 (tarball `package/docs/`):

```text
37a03d180115460a0a275022f4aeabef1e5345458df6a9b6e8a6de18474de720 rpc.md
4deeee367852f4633ef594bc079f9e9fb14640a4e3ba8a69b7c89b6f54740838 rpc-commands.md
f103191816fdcf3627dfc98b89acad8a28eee737697130ddaeb2581131213543 json.md
3b1e40a8d28b17c9eb7e496a7ba88fbe04665e3e6ae02873a58b943e325ad504 security.md
bc8798e4cc3209e455e3dc50c6652ae73c057d40122a4a741d8b5b65a041c2f1 providers.md
1eb60b0f1faad6f9fffd62ea257b545dbc983dbe4c5decc4803b1a241d604155 models-configuration.md
1a59b88928be611f04758e8607c96beea91327de803ff696c672c8216996dfc9 environment-variables.md
```

재현 스크립트 `measure_pi.py` 는 2026-11-13 #2682 에서 삭제했다 — pi 가 지원 대상에서 빠졌다(#2592 6841920695). 다시 재려면 마지막 판을 꺼낸다: `git show 3495ff9a63:docs/spikes/pi-v1.5.0/measure_pi.py > measure_pi.py`. 당시 재현 명령(저장소 루트에서):

```bash
python3 -I measure_pi.py --node-bin <node>=22.19 bin 디렉터리> \
  --install <pi 1.5.0이 설치된 디렉터리> --work <임시 작업 디렉터리>
```

스크립트는 `<work>/results.json`을 쓰고 출력한다. 아래 값은 두 번째 실행(run2) 결과다. 첫 실행(run1)은 "자격 증명 없는 prompt"를 mock 모델이 설정된 agent dir에서 재서 측정이 성립하지 않았고, 그래서 빈 agent dir로 바꿔 다시 쟀다. 나머지 항목은 run1과 판정이 같았다(시간 값만 다름).

## 4항 측정 결과 (AC-1)

| 항목 | 측정한 것 | 결과 | 판정 |
| --- | --- | --- | --- |
| **1. Conformance** (설치본이 고정 문서대로 동작하는가) | `--version`, `--help` 플래그 7개(`--mode` `--no-session` `--session-dir` `--provider` `--model` `--approve` `--no-approve`), RPC `get_state` id 대응, 잘못된 JSON, 모르는 명령, 문서의 RPC 명령 33개 중 세션을 바꾸지 않는 23개 호출, stdin 닫기 종료 | `1.5.0`, 빠진 플래그 0. `get_state` 성공·id 반사. 잘못된 JSON은 id 없는 `command:"parse", success:false`(문서와 같음). 모르는 명령은 `success:false, error:"Unknown command: no_such_command"`. 23개 모두 인식(미인식 0). stdin 닫으면 exit 0 | passed (23/33 호출. 나머지 10개는 세션 변경·작업 실행 명령이라 이름 인식만 문서로 확인, not_run) |
| **2. 취소** | 대기 중 `abort`, mock이 0.1초마다 토큰을 보내는 중 `abort`, 취소 뒤 다음 prompt, RPC `bash`(`sleep`) 실행 중 `abort_bash` | 대기 중 abort 성공. 스트리밍 중 abort 응답 162 ms(run1 112 ms), `agent_settled.aborted=true`, mock 쪽 연결 끊김 200 ms(run1 140 ms), 이후 `isStreaming=false`. 다음 prompt 정상 완료(`aborted=false`). `abort_bash` 응답 13 ms, bash 결과 `cancelled=true`, 남은 `sleep` 프로세스 0 | passed |
| **3. 보안** | 의존성 설치 스크립트, `npm audit`, RPC 모드 project trust(작업 폴더 `.pi/extensions/probe.ts`가 로드되면 표식 파일 생성), 모델의 도구 호출에 승인 단계가 있는가, 도구 자식 프로세스 환경 | 설치 스크립트 있는 의존성 2개(`@google/genai`, `protobufjs`) — `--ignore-scripts`로 설치해도 위 측정 전부 동작. `npm audit` 취약점 0(info·low·moderate·high·critical 모두 0). trust: 재정의 없음 → 확장 실행 안 됨, `--no-approve` → 안 됨, `--approve` → 실행됨. mock이 `bash` 도구 호출을 돌려주자 승인 요청 없이 실행되어 표식 파일이 생겼다. 도구의 자식 프로세스는 `AI_AGENT=pi`, `PI_CODING_AGENT=true`와 Pi 프로세스의 환경 변수(canary)를 그대로 받는다 | 측정 완료. 문서(security.md)가 밝힌 대로 **도구 호출 승인과 sandbox가 내장돼 있지 않다**. 경계는 호출하는 쪽이 OS·컨테이너로 만들어야 한다 |
| **4. 인증** | 자격 증명 없는 빈 agent dir에서 모델 목록·기본 모델·내장 모델 지정·prompt, `models.json`의 `apiKey: "!command"` 해석, mock이 받은 헤더, 생성 파일 권한 | 자격 증명이 없으면 사용 가능한 모델 0, 기본 모델 `unknown`, `set_model anthropic/...`은 `Model not found`, prompt는 `success:false, "No API key found for the selected model. Use /login ..."`이고 mock 요청 0(네트워크 전 거절). `!command` 키는 `Authorization: Bearer`로 전달됨(값 비교만, 출력 0). `auth.json`·`models-store.json`은 실행만으로 생성되고 모드 `0600`(`auth.json`은 빈 객체). `models.json`·`settings.json`은 `0644`. mock에 들어온 GET(카탈로그·텔레메트리 등) 0 | 측정 완료. OAuth/`/login`과 실제 공급자 인증은 not_run |

추가 관찰:

- Node v22.14.0(기본 PATH의 node)으로 `--version`을 실행하면 exit 1과 번들된 HTTP 클라이언트 코드의 스택이 출력된다. 원인은 분리하지 않았다. `engines >=22.19.0`을 지켜야 한다.
- 이 세션의 첫 `pi --version` 1회는 격리 환경 없이 실행됐다(실수). 실제 `~/.pi`의 수정 시각은 이전 그대로(2026-09-10)였고 새 파일은 없었다. 이후 모든 측정은 하네스의 격리 환경에서만 했다.

## not_run

| 항목 | 이유 |
| --- | --- |
| 호스팅 공급자 실제 호출, OAuth `/login`, API 키 발급 | 리더 지시: 유료 외부 호출·키 발급 없이 측정 |
| 문서 RPC 명령 10개(`prompt` `steer` `follow_up` `bash` `compact` `export_html` `switch_session` `fork` `clone` `new_session`)의 빈 본문 호출 | 세션 변경·작업 실행 명령. `prompt`와 `bash`는 다른 시나리오에서 실제 사용 |
| SDK(in-process) API | RPC만 측정. 설계가 고려한 subprocess 경계가 RPC다 |
| session 보존·재개, tool interception(확장 훅) | 설계 §14가 pin 대상으로 적었지만 이 카드의 4항에 들지 않아 측정하지 않았다 |
| `PI_OFFLINE=1`일 때 mock 밖으로 나가는 네트워크 | mock에 들어온 요청만 셌다. 프로세스 전체 egress는 관측하지 않았다 |
| Linux·Windows | macOS arm64에서만 실행 |
| 설치 스크립트 2개를 실행했을 때의 동작 | 일부러 `--ignore-scripts`로 막았다 |

## 권고 (AC-2)

**권고: 조건부 채택 — RPC subprocess로만, 호출하는 쪽이 격리 경계를 책임지는 형태.** 결정은 사용자 몫이다.

근거:

- 고정 버전의 RPC 계약은 측정 범위 안에서 문서와 일치했다(플래그, id 대응, 오류 형태, 명령 인식). 취소는 스트림·도구 모두 1초 안에 끝났고, 취소 뒤 세션이 다시 쓰일 수 있었다.
- 인증은 자격 증명이 없을 때 네트워크 전에 거절하고, 저장 파일은 `0600`이며, `!command` 키로 디스크에 비밀을 쓰지 않을 수 있다.
- 반면 Pi는 도구 호출에 승인 단계도 sandbox도 없고, 도구 자식 프로세스가 Pi의 환경 변수를 전부 받는다. 이는 결함이 아니라 문서에 적힌 설계다. 따라서 MoAI가 쓴다면 경계를 MoAI 쪽에서 만들어야 한다.

채택한다면 붙일 조건:

1. 버전과 integrity를 고정하고 `--ignore-scripts`로 설치한다(Node `>=22.19.0`).
2. RPC 자식 프로세스로만 띄우고 환경 변수는 허용 목록으로 넘긴다(키는 필요한 하나만, 가능하면 `!command`).
3. 자동 실행에는 `--no-approve`를 명시해 project trust를 고정한다.
4. 도구 실행은 컨테이너·sandbox 안에서만 한다(security.md 권고와 같다).
5. Desktop UI 카드 전에 not_run 항목 중 OAuth·실제 공급자·session 보존을 사용자가 준 시험용 자격 증명으로 따로 측정한다.

선택지:

| 선택지 | 결과 |
| --- | --- |
| A. 위 조건부 채택 | 후속 카드에서 RPC 어댑터와 격리 실행을 설계. not_run 항목 측정이 선행 |
| B. 비채택 | Desktop UI 카드는 다른 런타임을 고를 때까지 blocked 유지 |
| C. 보류 후 추가 spike | OAuth·session 보존·Linux 측정을 한 장 더 연 뒤 결정 |
