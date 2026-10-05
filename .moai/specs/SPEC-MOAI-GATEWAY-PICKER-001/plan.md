# SPEC-MOAI-GATEWAY-PICKER-001 — 구현 계획

## A. 경계와 선행 실측

Tier M으로 시작 선택과 picker 표면 한 경계를 다룬다. registry·정확한 모델 ID 라우팅·GLM tier 슬롯 정리/후주입·
선택 검증 인식기는 `SPEC-MOAI-GATEWAY-001` 소관이다. GPT 로그인은 `SPEC-MOAI-GPT-AUTH-001` 소관이다.
Default 행의 실제 의미, 명시 모델과 세션 설정 우선순위, continue/resume의 모델 전달은 실제 Claude Code TUI로 먼저
측정한다. 문서·mock만으로 클라이언트의 동작을 확정하지 않는다. 행 표시·격리 설정으로 아래 계약을 달성할 수 없으면
기존 사용자 설정을 바꾸는 우회 대신 증거와 blocker를 제출한다.

### A.1 공식 설정 후보와 적용 한계 (2026-09-11)

세션의 `--settings` overlay에 `modelPicker.options`를 두는 방식을 우선 시험한다. 각 행은 `model`·`label`·`description`으로
구성하고, `model`에는 canonical GPT 네 ID와 `claude-opus-5`·`claude-sonnet-5`를 그대로 넣는다. 필요한 GLM 항목은 코어
catalog 값에서 가져온다. GPT에 가짜 Anthropic 이름을 붙이지 않는다. `replaceBuiltInOptions: true`를 후보로 사용하되
Default와 현재 모델 행은 여전히 남는다는 전제로 둘을 각각 확인한다. 이 기능은 v2.1.242 이상이며 프로젝트/local 설정에서는
무시된다. managed 우선순위·allowlist·계정 접근 조건으로 행이 빠지거나 비활성화될 수 있다. 이는 공식 문서 판독이며 실제
계정의 행 표시 성공이 아니다. [설정 참조](https://code.claude.com/docs/en/settings-reference#modelpicker).

Gateway discovery는 GPT picker 등록 수단으로 의존하지 않는다. 공식 프로토콜의 `GET /v1/models?limit=1000` 결과는 ID에
`claude` 또는 `anthropic`이 있는 항목만 남기며, custom header만으로 인증하면 discovery 자체를 생략한다. 따라서 bare GPT
catalog 응답만으로 네 행이 나타난다는 전제는 채택하지 않는다. 이 제약을 우회하려고 코어 catalog ID나 인증 헤더 계약을
바꾸지 않는다. [Gateway discovery](https://code.claude.com/docs/en/llm-gateway-protocol#model-discovery).

Default 행은 `ANTHROPIC_DEFAULT_MODEL=<초기 canonical 모델>`을 후보로 시험한다. 명시 `--model` 전달을 대체하지 않는다.
조직 기본값·allowlist·계정 제약에 따라 이 변수가 무시될 수 있으므로 표시와 실제 요청을 함께 확인한다.
`/model`은 사용자 settings에 기본값을 저장한다는 문서가 있으므로 `--settings` overlay의 읽기 범위만으로 저장 격리를
충족했다고 판단하지 않는다. 설정 저장 경로 격리와 인증 유지의 양립은 아직 미결 실측이다. 이를 해결하려고 사용자
credential을 임시 디렉터리로 복사하지 않는다. [모델 기본값과 저장](https://code.claude.com/docs/en/model-config#set-a-default-model-for-new-sessions).

### A.2 사전 측정 표

| 조건 | 기대 관측과 기록 | 연결 |
|---|---|---|
| `--settings` modelPicker, 접근 허용 catalog | 네 GPT 행과 두 Claude 행 표시·선택·실제 요청 ID; Default·현재 행도 캡처 | AC-GP-004·006 |
| 같은 modelPicker를 프로젝트/local에만 배치 | 사용자 행이 그 설정으로 생기지 않음; 위 양성 대조군과 같은 버전 사용 | AC-GP-004 |
| discovery만 활성화하고 bare GPT·Claude ID 반환 | GPT는 discovery로 추가되지 않고 Claude 대조 항목의 수신·표시 여부를 기록 | AC-GP-004 |
| custom header 인증만 제공 | discovery 수신 0과 debug 생략 표시; modelPicker 선택은 별도로 판정 | AC-GP-004 |
| allowlist가 한 후보를 제외하거나 managed lineup이 우선 | 제외·비활성·우선순위 화면과 요청 0; 전체 성공으로 세지 않음 | AC-GP-004·008 |
| 초기 모델별 Default 후보와 조직 제한 변형 | 표시 provider·model과 실제 turn 일치, 무시되는 조건은 blocker | AC-GP-006 |
| `/model` 저장·`s` 선택·동시 세션·재개 | 전역/프로젝트 전후 해시·실제 쓰기 대상·요청 ID; 인증 보존 여부 별도 | AC-GP-003·005 |

사전 측정은 실제 TUI와 로컬 mock으로 형태·설정 동작을 확인한다. 실제 모델 추론·계정 가용성은 공통 실계정 조건에 따라
별도로 확인한다. 이 표를 작성한 것만으로 어느 행도 PASS가 아니다.

## B. 실행 순서

1. High — 세션 설정 경계·Default 표시·새 실행/continue/resume/restart의 클라이언트 행위를 실제 TUI로 측정한다.
2. High — 초기 선택 우선순위와 원본 설정 보존·동시 세션 격리 시험을 먼저 실패시킨 뒤 최소 overlay를 구현한다.
3. High — 정확한 네 GPT 선택 경로와 두 모델 변경 표면을 코어 catalog에 연결한다. gpt-5.6-sol을 빈 GPT 기본값의
   기준으로 사용한다. cc의 기본은 프로젝트의 정상 Claude 기본값, glm의 기본은 코어의 설정된 GLM 초기값이다.
   기본값이 catalog에 없거나 사용할 수 없으면 명시 오류로 종료하며 다른 provider로 대체하지 않는다.
4. High — 실제 TUI에서 /model·s·Default·새 실행·continue·resume·restart를 검증하고 네 GPT 실제 모델 왕복을 수행한다.
5. Medium — 코어·AUTH와 통합 감사 후 문서를 동기화한다. 원격 출시·병합은 기존 별도 지시 경계를 유지한다.

## C. 검증과 증거

TDD 설정은 `.moai/config/sections/quality.yaml`을 따른다. 전역 설정·프로젝트 설정의 전후 해시는 격리된 임시 fixture로
재현하고 실제 사용자 파일은 변경하지 않는다. 경로·argv·모델·upstream 계수·화면 증거를 함께 남기며 비밀을 제외한다.
/model의 저장 기본값이 세션 전용 설정에만 닿는지를 파일·화면 양쪽에서 확인한다. 대화 선택과 기록은 보존하되 재개 시
모델은 이번 launcher 선택을 우선한다는 정책을 사용자 표시에도 반영한다.

## D. 현재 준비 상태

이 문서는 초안이며 클라이언트 사전 측정·계획 감사·별도 카드 dispatch가 필요하다. 이번 문서 생성은 카드 생성이나
구현 착수를 뜻하지 않는다. GPT-AUTH의 실제 OAuth 등록 근거와 계정 접근이 해결되기 전 전체 실제 모델 AC는 미완료다.

## 공통 실계정 시험 조건

Claude 실서비스 시험은 사용자 지시에 따라 **2026-09-11 19:00 Asia/Seoul 이후** 실행한다. Claude 시험 입력은
`claude-opus-5`와 `claude-sonnet-5`이며 계정의 실제 사용 가능 여부는 별도 관측한다. Sonnet 4.5 과거 캡처는 대체 근거가 아니다.
GPT 시험의 upstream ID는 `gpt-6-astra`, `gpt-5.6-sol`, `gpt-5.6-terra`, `gpt-5.6-luna`다. 임의 `gpt-6` 별칭을 만들지 않는다.
공식 `gpt-5.6` 별칭을 UI에 추가하더라도 공개 근거와 canonical 매핑 시험이 먼저 필요하며 본 초안의 필수 범위에는 넣지 않는다.
429는 실패한 인증 방식의 증명도 성공도 아니다. Claude OAuth의 코어 T09가 INCONCLUSIVE이면 그 상태를 유지한다.

## 근거와 완료 판단

사용자 범위·공식 모델 문서 위치는 `../../reports/SPEC-MOAI-GATEWAY-001/goal-execution-20260911.md`, 코어 경계는
`../SPEC-MOAI-GATEWAY-001/spec.md` 및 `design.md`가 기준이다. 코어의 기존 lint·mock PASS는 이 SPEC의 완료 근거가 아니다.
각 AC별 실행 명령·그 출력·HEAD·사용한 모델·mock/실계정 구분·Gap·잔여 위험을 보고한다. 관련 구현·통합 시험·독립 감사와
문서 동기화 전 완료로 표시하지 않는다. 이번 계획 작성은 numeric card 생성이나 기존 worktree의 새 카드 착수가 아니다.

## D. 19시 이후 관측의 현재 범위 (0.2.0)

실제 Claude Code 2.1.268 안내는 `Enter to set as default · s to use this session only`다.
`picker-reasoning-runtime-observation.md`의 Sol/Luna 요청과 `picker-followup-runtime-observation.md`의 Astra/Terra 요청은
정확한 네 ID가 로컬 mock에 도착한 근거다. 계정 권한·실제 추론의 PASS는 아니다. Default의 표시·실제 요청은 Sol이고
그 선택 직후 settings snapshot은 model 없는 빈 객체였다. 저장 안내만으로 새 세션 영속성을 추론하지 않는다.

`picker-resume-runtime-observation.md`의 두 프로세스는 같은 격리 HOME에서 정확한 UUID --resume으로 이어졌고
합성 opaque/tool pair/metadata.user_id 내부 session_id를 보존했다. 두 프로세스 exit 0이다. 이는 exact resume의 양성군이며
fresh session·continue picker·fork-session·원본 사용자 설정 격리·실계정 인증 보존은 별도다. 대화 UUID 일치만으로
권한을 증명하지 않는다. 코어의 대화별 hash receipt는 코어 design §4.3 소관이며 launcher의 명시 재개 대상과 연결한다.
이 보고서들은 모두 `../../reports/SPEC-MOAI-GATEWAY-001/` 아래 있다. 기존 A.2 음성군과 출시 결합을 유지한다.

## E. Receipt와 연결하는 private profile 수명 (0.2.0 내 감사 보정)

코어 design §4.3의 RPA-1 lifecycle 표를 공통 계약으로 따른다. launcher가 new의 UUID를 만들고, exact/continue/선택
resume은 MoAI 소유 대화 인덱스에서 UUID와 family를 확정하여 native exact resume로 인계한다. fork도 parent와 새 UUID를
송신 전에 확정한다. native 요청 metadata는 그 결정의 일치 검사이며 root 선택권이 아니다. 원본 사용자 세션을 자동 복사하지 않는다.

private native config/transcript와 UUID별 hash receipt는 retained family에 보존한다. 정상 종료·취소에는 임시 overlay만
기존대로 정리하고 sole resume transcript를 지우지 않는다. 명시 대화 폐기만 소유 상태를 제거하며 살아 있는 sibling/fork의
공유 family를 삭제하지 않는다. family 실행 lease는 같은 profile에 대한 두 lead의 /model 쓰기 경합을 명시 busy로 거절하고
다른 family는 독립 실행한다. 한 lead의 병렬 gateway 요청에는 별도 후보 수용 표를 적용한다.

원래 secure-storage namespace를 private config 변경 전 확보하여 CLAUDE_SECURESTORAGE_CONFIG_DIR에 보존한다.
2.1.268의 source/Opus 5 양성은 picker-settings-source-observation.md의 제한된 근거다. credential 복사·전역/프로젝트
설정 쓰기는 금지한다. 다른 버전·플랫폼·미검증 인수 조합은 preflight로 확인하기 전 egress를 열지 않는다. 이 내부 경로를
공개 지원 약속으로 표시하지 않는다. 실제 continue·선택 resume·fork 및 retained profile/OAuth 통합 AC는 계속 필수다.
