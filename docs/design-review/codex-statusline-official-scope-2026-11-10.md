# Codex CLI·Desktop 상태 표시 공식 지원 범위와 Rust 어댑터 선택지 (#2357)

- 조사: Codex t3j, `session=codex-t3j`, epoch 1, `Moai-Claim: 2357#6755712814`.
- epoch 3 독립 재검증: agent-10-t6b, `Moai-Claim: 2357#6759371745`, 2026-11-10 — 11개 소스 전건 재확보·해시 대조와 주장별 원문 확인. 상세 기록: `codex-statusline-verification-2026-11-10.md`. epoch 3은 판정을 바꾸지 않았고 사실 서술 2곳을 공식 원문 실측 명칭으로 정정했다.
- 조사일: 2026-11-10 UTC. 출처별 정확한 시각·응답 SHA-256은 마지막 표에 있다.
- 저장소 기준 ref: `6bf79ecd85b9d40653c798f89266a7ffe2f28cbf`(착수 시 원격 integration tip 조회, 이 문서가 커밋된 커밋의 부모). 인계 경위: Codex t3j의 `/tmp` 인계 초안을 agent-8 [860b69]가 4파일 SHA-256 검증(254952ea…·a6e8b621…·3c311f56…·179d9c5b…)과 출처 스냅샷 11건·claude-contract-snapshot hash 대조, 원격 URL 2건 재확인(200·bytes 일치) 뒤 원문 그대로 `docs/design-review/`로 커밋했다. 조사 본문과 판정은 t3j 원문이며 이 커밋이 내용을 고르지 않았다.
- 비교 계약: #2331의 병합본 `01e123e6c374386f952277103673bee6a05a3ad0`, `docs/design-review/statusline-rust-stage1-2026-11-10.md` §1·2·6을 실제 읽었다. 아래 Claude 열은 이 고정 문서의 계약이며, 현재 설치된 Claude의 실세션을 관측한 결과가 아니다.
- 범위: 공식 문서·릴리스 기록 조사와 설계 질문. 코드·실제 Codex 설정·Go·Orca wrapper·hooks·rules 변경, 인증 추출, 실제 호스트 기동·유료 호출은 하지 않았다.

## 1. 결론과 판정 기준

Codex CLI의 내장 상태줄과 터미널 제목 선택은 공식 지원된다. 그러나 내장 항목을 고르는 설정을, 외부 Rust 명령에 세션 JSON을 보내 그 stdout을 상시 상태줄에 그리는 계약으로 해석하면 안 된다. 후자의 CLI·Desktop 확장점은 읽은 공식 자료만으로 확인하지 못했으므로 `not_verified`다. 문서에 없다는 이유만으로 제품 전체가 절대로 지원하지 않는다고 단정하지 않는다.

공통 Rust 세그먼트 계산은 입력이 실제로 공급된 항목에 한해 설계할 수 있다. 호스트별 데이터 수집과 표시 위치는 별도 문제이며, 이번 카드가 구현 방식·설정 위치를 선택하거나 설치하지 않는다.

판정은 `supported`(공식 계약 확인), `unsupported`(공식적인 제거·비지원 명시), `not_verified`(자료가 부족하거나 설치본/권한/동작을 확인하지 않음)로 나눈다. 아래 모든 판정은 문서 기준이고 라이브 검증은 `not_run`이다. S 번호는 출처 표의 URL·시각·전체 응답 해시와 연결된다.

## 2. 호스트별 지원 표 (AC-1)

| 호스트·수단 | 판정 | 공식으로 확인한 설정·입력·갱신 | 제약·미확인 | 근거 |
|---|---|---|---|---|
| CLI 내장 footer | supported | `/statusline`, `tui.status_line`. 모델·reasoning·context·rate limit·git·토큰·세션·디렉터리·버전 등의 내장 항목 선택·순서 조정. picker 변경은 즉시 반영, 세션 한정 인자도 설명됨 | 외부 명령 stdin JSON/사용자 정의 항목 등록 계약은 아님. 데이터별 정확한 갱신 주기는 not_verified | S01·S02·S03 |
| CLI 터미널 제목 | supported | `/title`, `tui.terminal_title`. project·spinner·status·thread·git branch·model·task-progress 등 내장 ID 선택 | OS 터미널 제목이며 Desktop 본문/전역 footer 확장과 다름. 임의 Rust stdout 연결은 not_verified | S01·S02·S03 |
| CLI 외부 Rust 상시 상태줄 | not_verified | 읽은 문서에는 위 내장 ID 설정만 확인 | Claude `statusLine.command`와 같은 실행 명령·stdin JSON·refreshInterval 계약을 확인하지 못함 | S01·S02·S03 |
| CLI 터미널 알림 | supported | `tui.notifications`, `tui.notification_method`의 auto/osc9/bel, `tui.notification_condition`. 완료·승인 요청 등 이벤트에 반응 | 알림 기능이며 연속 표시·모든 세그먼트 공급을 보장하지 않음 | S04 |
| 외부 notify 프로그램 | supported | `notify`가 현재 `agent-turn-complete`에서 프로그램을 실행. JSON은 stdin이 아니라 단일 인자로 전달됨. `type`, `thread-id`, `turn-id`, `cwd`, 입력 메시지·마지막 답변 필드 설명 | 완료 이벤트 중심. 토큰·context·과금·사용량의 정기 공급 계약은 확인되지 않음. 메시지 원문 수집은 이번 설계 입력에서 제외해야 함 | S04 |
| CLI·Desktop hook 상태 메시지 | supported | command hook stdin JSON의 `session_id`, `transcript_path`, `cwd`, `hook_event_name`, `model`; turn-scoped 이벤트는 `turn_id` 등. `statusMessage`는 hook 실행 중 표시, `systemMessage`는 경고 UI/event stream | 이벤트별·일시적 표시. raw stdout을 상시 footer로 쓰는 계약 아님. 비관리 hook은 검토·신뢰가 선행됨. transcript 파일 형식은 안정 인터페이스가 아니라고 명시 | S05 |
| Desktop 완료 알림·기본 UI | supported | 공식 Desktop Settings의 완료 알림·외관·활동 관련 기본 기능, 명령/단축키의 Codex 전용 terminal·side chat·review tab 등(epoch 3 정정: S07 원문에 "Codex view"라는 명칭은 없고, 실측 명칭은 "Toggle terminal (Codex only)", "Open side chat (Codex only)", "Open review tab (Codex only)"이다) | 공식 문서가 현재 ChatGPT desktop app의 Codex 화면을 함께 설명함. 앱 버전별 실제 표시와 현재 사용자의 설치본은 not_verified | S06·S07 |
| Desktop 사용자 정의 고정 footer/Rust command slot | not_verified | Desktop 설정·명령·hooks 자료를 읽었으나 해당 계약을 확인하지 못함 | CLI `tui.*` 키나 Desktop의 내장 화면이 곧 사용자 정의 확장 슬롯이라는 추정은 하지 않음 | S02·S05·S06·S07 |
| app-server 기반 별도 클라이언트 | supported(문서상 통합 인터페이스 존재) | 자체 JSON-RPC; thread/turn·plan·token usage·rate limit 이벤트/조회가 문서화됨. 기존 Desktop에 주입하는 API가 아니라 별도 클라이언트 설계 경로 | 운영 안정성은 문서 간 주의 문구가 다름. S11은 app-server 전체를 experimental/production 비지원으로 설명하고 S08은 특히 WebSocket을 experimental/unsupported로 명시. 최소 버전·선택 transport·인가·스키마 검증 전 제품 의존 채택 불가 | S08·S11 |
| CLI·Desktop의 외부 MCP 서버 | supported | host 측 MCP 연결, STDIO/Streamable HTTP, 도구 요청과 설정 공유가 공식 설명됨 | 도구 결과·외부 문맥 경로이지, 모델 없이 지속 갱신하는 native footer 또는 CLI 내부 사용량 구독권을 보장하지 않음. 상세 UI 슬롯·표시 주기는 not_verified | S10 |
| 제거된 `codex mcp-server`로 telemetry 제공 | unsupported | 공식 제거 문서가 명령·standalone 서버 제거와 이전 예제 미지원을 명시 | 외부 MCP 서버를 Codex가 이용하는 기능의 제거와 혼동 금지. 대체 app-server도 즉시 production 허용으로 보지 않음 | S11 |
| 릴리스 노트의 실험 UI renderer | not_verified(공개 확장 계약) | 공식 changelog에 상태줄의 experimental UI tree renderer 관련 항목을 확인 | 릴리스 항목만으로 외부 등록 schema·프로그램 입력·Desktop 표시 지원이 확정되지 않음. 내부 renderer와 공용 확장 API를 동일시하지 않음 | S09 |
| 일부 workspace의 내장 credits/cost 표시 | supported(출시 기록의 자격 범위 한정) | changelog는 eligible workspace의 /status·상태줄·터미널 제목에 추정 credits/cost 표시를 언급 | 모든 계정 지원·정확한 과금·외부 Rust 입력 schema는 not_verified. 현재 계정의 사용량을 조회하지 않음 | S09 |

설정의 비활성화 표기도 자료 차이가 있다. reference는 `null`을 비활성으로 설명하고, sample은 status_line 비활성 표기 없이 주석 예시(항목 목록)만 두고 있다(epoch 3 정정: S03 원문에 status_line의 빈 배열 `[]` 표기는 없고, `[]`는 sample의 다른 키들에 있다). 이 카드에서는 어느 값을 실제 설정에 쓰거나 설치본에서 동작을 단정하지 않는다(S02·S03).

## 3. 입력 데이터와 갱신 경계

| 입력 경로 | 확인한 데이터·시점 | 공통 계산에 쓰기 전 조건 |
|---|---|---|
| CLI 내장 TUI | 표시 가능한 항목 종류와 선택 직후 UI 반영 | UI가 값을 표시한다는 사실과 외부 Rust 프로세스에 값을 내보낸다는 계약은 별개 |
| notify | 완료된 turn의 식별자·cwd 등, 단일 JSON 인자 | 연속 telemetry로 가정하지 않음. 대화 원문을 상태 계산에 요구하지 않음 |
| hooks | lifecycle/tool 이벤트의 공통 JSON·이벤트별 확장 필드 | `session_id`와 `thread-id`를 무조건 같은 식별자라고 보지 않음. 부모/자식 세션 범위도 보존 |
| app-server | `thread/status/changed`, `turn/started`, `turn/completed`, `turn/plan/updated`, `thread/tokenUsage/updated`; `account/rateLimits/read`·updated | token usage 상세 필드·context 상한은 이번에 읽은 이벤트 설명만으로 완결 schema를 확인하지 못함. 구현 전 공식 schema와 버전을 고정해야 함 |
| rate limit 응답 | primary/secondary 등 bucket, `usedPercent`, `windowDurationMins`, `resetsAt`; 일부 응답의 plan·credits | 창을 무조건 5H/7D로 고정하거나 credits를 USD로 변환하지 않음. 현재 계정 권한·가용성은 not_verified. 실제 호출은 하지 않음 |
| Desktop/MCP 결과 | 기본 UI·도구 요청과 응답 | refresh 주기·확장 panel·항목 export를 새로 가정하지 않음 |

S08의 이벤트를 읽기 위해 이번 세션에서 서버를 기동하거나 account·rate-limit·auth API를 호출하지 않았다. `transcript_path`가 있다는 이유로 실제 transcript·SQLite·캐시를 읽지 않았으며 이를 안정된 telemetry API로 채택하지 않는다(S05).

## 4. 공통/호스트별 세그먼트 비교 (AC-2)

아래 '공통 가능'은 순수 계산·정규화의 설계 가능성이지 Codex 화면 착륙을 검증한 뜻이 아니다. 미제공 값은 unknown/미제공으로 남기고 0·정상·fresh로 만들지 않는다.

| 세그먼트 | Claude #2331 고정 계약 | Codex 공식 근거·차이 | 분류 |
|---|---|---|---|
| 모델·작업 디렉터리 | stdin model·workspace/cwd | hook model·cwd와 CLI 내장 항목(S01·S05) | 입력이 있을 때 공통 가능 |
| reasoning/effort | effort·thinking | CLI model+reasoning 내장 표시 확인. 외부 hook의 effort 공급은 미확인 | 계산 함수 공유 가능, Codex 수집은 별도 확인 |
| context 비율·토큰 | context_window·current_usage | CLI 내장 context/token 표시, app-server usage 이벤트 존재(S01·S08). 외부 전달 필드·분모 상세는 미확인 | 조건부 공통; 단위·상한·cache 포함 정의 필요 |
| 사용 한도·리셋 | five_hour·seven_day 등 | app-server 가변 길이 bucket(S08). Claude 명칭을 그대로 대입할 수 없음 | window 길이를 가진 공통 모델 가능 |
| 비용·credits | total_cost_usd 등 | 일부 workspace의 내장 추정 표시 출시 기록은 있음(S09). 일반 알림/hook의 USD 입력은 미확인이고 rate-limit credits는 USD와 동의어가 아님 | 공통 금액 표시 채택 보류 |
| cache-hit·prompt cache | current_usage·prompt_cache | 동등한 공개 hook 입력·계산 의미 미확인 | 현재는 Claude 계약 전용; Codex not_verified |
| 세션/turn·진행 상태 | session_id, session_name 등 | hook·notify 식별자와 app-server thread/turn/status/plan(S04·S05·S08) | 출처와 namespace를 보존하는 공통 상태 모델 가능 |
| git branch·변경 수·저장소 | stdin repo 일부와 로컬 git | CLI 내장 branch, 외부 통계 수집은 별도(S01) | 승인된 읽기 collector와 순수 렌더 분리 가능 |
| PR/Issue·backlog·MoAI task | stdin pr 및 로컬 MoAI 기록 | Codex title task-progress가 MoAI claim/Issue 상태라는 근거 없음 | 기존 승인 ledger/cache 별도 입력; native 항목과 혼동 금지 |
| host 버전·MoAI 버전 | version, 별도 MoAI 버전 | CLI 내장 Codex version은 확인. 외부 export·실제 설치 버전 미확인 | 제공된 버전 라벨만 공통 가능 |
| vim·output style·Claude cache 세부 | #2331 표의 전용 필드 | Codex 동등 입력 계약 미확인 | Claude 전용/추후 확인 |
| Codex title·thread runtime·quota bucket | Claude statusLine stdin과 별개 | 제목의 spinner/status/task-progress, thread runtime 상태·가변 quota(S01·S08) | Codex 어댑터별 선택 항목 |

가능한 구조는 `호스트별 collector → 출처·시각·단위·가용성 있는 normalized input → 순수 Rust segment 계산 → 호스트별 presenter`다. 이는 논의용 대안이며 이번에 구현/선택하지 않았다. Claude의 renderer stdout 계약을 Codex presenter의 계약으로 복사하지 않는다. #2350의 Go 동등성 golden 30개 통과가 Codex 표시·지원 증거도 되지 않는다.

## 5. 리더가 사용자에게 확인할 선택지 (AC-3)

| 결정 질문 | 선택지·장점 | 대가·남은 확인 |
|---|---|---|
| Codex에서는 어디에 표시할까? | A 내장 footer/title만 사용: 공식 범위가 명확함. B 완료 알림·필요 시 상태 요약: 별도 상시 UI 불필요. C 별도 app-server client/panel: 세부 상태 설계 가능 | A 임의 MoAI 세그먼트 주입 미확인. B 연속 telemetry 아님. C 안정성·스키마·인가·보안·개발 비용 별도 결정. Desktop native footer 확장은 미확인 상태로 남음 |
| Rust 공통 계층의 입력은 어디까지인가? | A 호출자가 넘긴 최소 데이터만 순수 계산. B 승인된 읽기 collector·캐시 추가. C app-server 이벤트 추가 | A 누락값이 많음. B 경로·소유권·freshness 계약 필요. C 실험 API·인증·버전 경계 필요. transcript 내부 파일 직접 해석은 안정 API가 아님 |
| 결과가 없는 세그먼트는 어떻게 할까? | A 숨김. B unknown·출처/시각 표시 | A 사용자가 미지원과 0을 혼동할 수 있음. B 표시가 길어짐. 리더가 선택하도록 남김 |
| 설정 위치·적용 주체는? | A 제품의 별도 계획만 생성. B 검토된 host별 plan/apply로 선택 항목만 적용 | B는 별도 승인·receipt·rollback·소유권이 필요. 이번 카드에서 ~/.codex나 wrapper를 바꾸지 않음 |
| Desktop와 CLI의 목표를 같게 할까? | A 데이터 계산만 공유하고 각 presenter가 가능한 정보만 표시. B 기능별 동등성을 별도 AC로 정의 | 표시 위치·갱신 시점·입력 가용성이 달라 4줄 UI 완전 동등성은 아직 근거 없음 |

비용·보안·설정·게시 HOLD를 이 선택지 목록으로 해제하지 않는다. 릴리스 renderer 항목을 후속 조사하려면 별도 공식 설정/schema와 설치 버전 검증을 AC로 삼아야 한다.

## 6. 확인 한계와 인계 (AC-4)

- actual runtime: Codex CLI/Desktop·hook·notify·MCP·app-server를 실행하거나 연결하지 않았다. 사용자 설정·대화·인증 정보도 읽지 않았다.
- 로컬 metadata: PATH는 standalone Codex를 가리키고, 별도 npm 패키지 README/package.json은 0.130.0이었다. 이 둘을 같은 설치본으로 보거나 현재 활성 버전으로 적지 않는다. 공식 자료는 현재 문서의 지원 범위일 뿐 이 컴퓨터의 동작 보증이 아니다.
- Desktop 문서는 현재 ChatGPT desktop app 안의 Codex 기능도 다룬다. 이것을 Orca wrapper의 기능으로 확대하지 않는다.
- source acquisition: Markdown 404와 잘린 첫 changelog 응답은 미확인/부분 응답으로 보존했다. 이후 정상 URL의 완전 응답만 공식 근거로 사용한다. 실패를 지원 부재의 증거로 삼지 않는다.
- native worktree: list_worktrees가 로컬 MCP transport 연결 오류로 실패했다. worktree 생성·대체 git worktree add·repo 쓰기·commit/push는 하지 않았다. HERDR_ENV 조건도 미충족이며 우회하지 않는다.
- `cargo test`, `cargo clippy`, `cargo fmt`, Python 전체 unittest, Git 작업 tree diff 검사, fresh clone 검사는 모두 not_run이다. 문서 PR·독립 감사·병합도 미실행이다. 출처·문서 검사와 전체 check set을 혼동하지 않는다.
- 디스크: 착수 때 약 5.9GiB, 후속 조회 약 9.1GiB. 이 세션에서 정리·빌드·설치를 하지 않았다.
- 역할 정정 `adk-role-20261110-0957`을 접수했다. #2357은 실제 착수 카드이므로 산출·소유권·검사 한계를 갖춘 리더의 명시 안전 인계 전까지 보존한다. 그 뒤 신규 PR 전담 역할로 전환한다. 이 초안이 있다고 AC-4·이슈 종료를 완료 처리하지 않는다.

## 7. 출처 원장

SHA-256은 HTTP body 원문(수신한 decoded bytes) 기준이다. 시각은 UTC이다. 전체 응답 해시와 부분 응답 해시를 구분한다. 공식 문서는 문서 지원의 근거, #2331 사본은 비교 계약의 근거이며 구체적 runtime 관측의 근거가 아니다.

<!-- sources-table:begin -->
| ID | 공식 URL | 확인 시각(UTC) | HTTP·bytes | 응답 SHA-256 |
|---|---|---|---|---|
| S01 | [commands](https://learn.chatgpt.com/docs/developer-commands.md) | 2026-11-10T00:55:13.246053+00:00 | 200 · 64320 | `d65c47d1e7095474508d4d1f3ac83e67c23985bcc442d36223bb77e2c5458181` |
| S02 | [config-reference](https://learn.chatgpt.com/docs/config-file/config-reference.md) | 2026-11-10T00:55:13.247149+00:00 | 200 · 153298 | `78b5f4af406fa8ab57dce0267a14c62954cdfeb4543d4b18f938ad76a70c7e96` |
| S03 | [config-sample](https://learn.chatgpt.com/docs/config-file/config-sample.md) | 2026-11-10T00:56:04.904695+00:00 | 200 · 36506 | `f5200b38f5d7ff485c3fe9f54995f1c1876d2a1ac8e38424919a1ade6dab8664` |
| S04 | [config-advanced](https://learn.chatgpt.com/docs/config-file/config-advanced.md) | 2026-11-10T00:55:13.248020+00:00 | 200 · 56830 | `ce85d4b5fd1604b8d9da8c82bb2ddba4acd5e9ebca0fc9c5f7fd4ce6a1b5bcaf` |
| S05 | [hooks](https://learn.chatgpt.com/docs/hooks.md) | 2026-11-10T00:55:13.321424+00:00 | 200 · 48400 | `a50c60b697d8df62e2f907adbb18b2f5a1a25cd3360e7b843cde25f18ab6d782` |
| S06 | [desktop-settings](https://learn.chatgpt.com/docs/reference/settings.md) | 2026-11-10T00:56:04.900595+00:00 | 200 · 5007 | `b66dc62c6718eb2e48aaaaf76f98b1e3ebd2ef5f8bb5979e2f8a2f6fb4e973e0` |
| S07 | [desktop-commands](https://learn.chatgpt.com/docs/app/commands.md) | 2026-11-10T00:56:04.903955+00:00 | 200 · 26950 | `c5e9b64bb8424a75df6f1e0437d9a2f068a8f39dcdaa28f3cdbdac1130267233` |
| S08 | [app-server](https://learn.chatgpt.com/docs/app-server.md) | 2026-11-10T00:55:13.342589+00:00 | 200 · 106105 | `b9e75507755ba0be2ebc615ca1f4399a54a386125f3935f615405ea93efe77f2` |
| S09 | [changelog-full](https://learn.chatgpt.com/docs/changelog) | 2026-11-10T00:58:58.802768+00:00 | 200 · 4771489 | `5ca7955b2fb4b5fc80be87488ccc4a83fc262367a3fb92a4f9d1a5c85aa2e13b` |
| S10 | [mcp-current](https://learn.chatgpt.com/docs/extend/mcp.md) | 2026-11-10T00:57:42.911943+00:00 | 200 · 31613 | `faf9407740483f7ee27ac4d0db08a9d15510b3d241251edc9d00cf7b35c29e3b` |
| S11 | [mcp-server-removal](https://learn.chatgpt.com/docs/mcp-server.md) | 2026-11-10T00:57:42.912329+00:00 | 200 · 1239 | `9fe223993523b4e6d991fd229888ecb1a770a3df3aa7bea946ae6674f8af32da` |
<!-- sources-table:end -->

비교 사본: [#2331 고정 문서](https://github.com/modu-ai/moai-adk/blob/01e123e6c374386f952277103673bee6a05a3ad0/docs/design-review/statusline-rust-stage1-2026-11-10.md), 이번에 받은 1–160행 사본 SHA-256 `849376c3d4b79582c27bb2787ab3707be5e5a8321e9f1ab738f4202492737976`. 전체 원본 파일 해시라고 주장하지 않는다.

미확인 수집 이력: changelog.md와 이전 추정 mcp.md는 HTTP 404; 첫 changelog HTML은 3,000,000 bytes 상한의 부분 응답으로 제외했다. 실제 공식 현재 MCP 경로와 changelog 전체 응답은 위 S10·S09다. 이 실패 기록도 sources.json에 유지한다.
