# MoAI GPT 구조 비교 및 개선 보고

## Claim — 결론과 완료 범위

claude-code-proxy 원본 Rust 126파일 77,078행을 담당별로 EOF까지 읽었다. 원본을 통째로 복제하는 대신 **Codex App Server를 유지하고, 오류 분류·대화 소유권·스트리밍 완료·도구 왕복의 계약을 강화**한다. 원본에도 버퍼링 경로, 임의 인자 보정, 약한 assertion, 문서와 구현의 차이가 있으므로 완벽한 기준 구현으로 취급하지 않는다.

재현한 MoAI 결함 하나는 구조적으로 수정했다. 권한이 확인된 도구의 schema 오류를 세션 권한/전송 오류와 구분하고, GPT에 실패 피드백을 전달하여 같은 turn에서 수정하도록 했다. **유효한 호출만 Claude Code에 전달하며, 외부 세션·미등록 도구는 계속 거부한다.** 소스·로컬 회귀 검증 완료이며 설치본에는 아직 반영하지 않았다.

이 문서는 전체 MoAI 저장소 EOF 감사, 모든 Claude Code 기능 호환, 과거 모든 400/502 해결, 현재 구독 약관 허용을 선언하지 않는다.

## Baseline-attribution — 조사 기준

| 대상 | 실제 기준 |
|---|---|
| Proxy | `b642a6ee1a5fff8ebc9bedfc4b438ddfa3ae8f8e`, Cargo 0.1.39, clean clone `/tmp/moai-proxy-structural.IArJmM/source` |
| MoAI | `WT-gpt-session-repair`, HEAD `b45c81349` + 기존 변경을 보존한 dirty tree |
| 설치본 | rc.11 / `v3.2.0-rc.11-b45c81349-summary-isolation-dirty`, built 2026-09-14T16:05:26Z |
| 설치본 SHA256 | `f40ba4cb307e62fb1e4b48031a4e8536e78cd79e0081986a43555e36116184d1` |
| 변경 제한 | 설치본 교체·사용자 프로세스 종료·commit/push/integration 미실행 |

Rust EOF 장부: 공통/서버/UI 26파일 15,921행, Codex 35파일 30,465행, 기타 제공자+공통 54파일 20,809행, 통합 테스트 11파일 9,883행. 문서·사이트 소스는 고유 27파일 2,810행을 별도로 읽었다. 의존성 lock·이미지·바이너리는 EOF 코드 독해 대상에서 제외한다. 세부 범위는 아래 장부를 따른다.

- [주 담당 장부](parent-read-ledger.md)
- [Codex 전체 분석](codex-analysis.md)
- [기타 제공자 전체 분석](other-providers-analysis.md)
- [통합 테스트 계약 분석](test-contract-analysis.md)
- [문서·사이트 소스 분석](docs-analysis.md)

## 구조 비교 — 흡수할 것과 유지할 것

| 경계 | 원본에서 읽은 방식 | MoAI 결정 |
|---|---|---|
| 모델 연결 | Anthropic 요청을 제공자별 HTTP/WS/protobuf로 변환 | 인증·native thread는 App Server에 유지. Rust로 재개발하지 않음 |
| 도구 실행 | Codex는 function call 변환, Cursor는 별도 XML 도구 bridge | Claude Code가 도구·권한·Agent 실행의 주체. Cursor bridge로 대체하지 않음 |
| 잘못된 도구 인자 | Read에 whitespace/offset 등의 개별 보정 | 공통 schema 실패 피드백. 인자를 추측하거나 필드를 임의 생성하지 않음 |
| 세션 연속성 | owner+generation+socket 검증, stale 완료 격리 | 현재 receipt/owner 검증 유지. main·agent·보조 요청별 회귀 계약 강화 |
| 보조 요청 | 자동 보안 검토는 stateless이며 기존 affinity를 갱신하지 않음 | agent_summary/away_summary/review를 실제 요청 증거로 구별. 일괄 우회 금지 |
| 재시도 | 의미 출력 전만 retry, 부분 출력 이후 재생 금지 | 기존 공개 출력 이후 재생 금지 유지. 도구 인자 수정은 transport retry와 별도 |
| 스트림 | Codex live와 buffered의 상태 처리 분리, Kimi/Cursor 전체 버퍼 경로 | 현재 증분 출력 유지. 공통 상태 전이 계약과 live/non-stream 동일 fixture 검증을 우선 |
| 성능 | 생성 구간·token sample 측정, 값이 없으면 속도 미상 | TTFB/첫 텍스트/완료/도구 시간/생성 TPS를 구분. 추정 TPS로 정상 판정 금지 |
| 모델 정책 | 제공자별 대체 모델과 capability 변형 존재 | 지정한 GPT 4종만 유지. 사용자 몰래 모델·effort 변경 금지 |
| 이미지 | 원본은 `gpt-image-2` | 요청한 `gpt-image-2.5`와 다름. 지원을 증명하지 않으며 별도 검증 필요 |

제품 모델은 **메인 세션 모델 하나와 서브에이전트에 지정된 모델**이다. plan/run/routine별 메인 세션 분할이나 예시 이름의 새 agent preset은 도입하지 않는다. Factory는 lead+lane 세션 구성이고 이 gateway 계약을 각 세션에 적용한다. Kanban 제품 기능을 되살리지 않는다.

## 재현된 장애와 이번 수정

### Agent schema 오류가 세션 중단으로 확대된 문제 — High

운영 family `04c54aae-7cd6-4753-a933-a584d40f6553`의 luna/max 호출에서 Agent의 필수 `description`이 누락됐다. 해당 도구 schema는 `description,prompt`를 요구했다. 검증 실패가 기존 일반 오류 경로를 따라 turn 중단과 failed barrier로 이어지고, 후속 요청이 recovery_required로 거부됐다.

변경 지점:

- `internal/codextools/registry.go:332`: owner/tool/schema binding 검증 후 schema 실패만 `ErrArguments`로 분류.
- `internal/codexbridge/engine.go:702`: 최대 3회의 schema 실패 피드백. 유효한 호출 전에는 public tool ID나 실제 도구 실행을 만들지 않는다.
- 첫 tool과 병렬 tool burst 양쪽에 적용. 교정 예산 초과, 다른 thread, 미등록 도구는 차단한다.
- `tool_argument_repair_test.go`와 `gpt_appserver_argument_repair_integration_test.go`: 잘못된 인자 → negative feedback → 유효한 Agent 한 번 → HTTP200 완료를 검증.

기존 실패 barrier를 임의로 활성화하지 않는다. 이 수정이 실제 GPT의 모든 잘못된 JSON을 복구한다는 뜻도 아니다. malformed JSON, 호출 ID 중복, schema 변경, 권한 위반 등은 별도 경계다.

### 앞선 수집 건과의 관계

취소 후 이미 수락한 tool result의 중복 전달, 검증된 native clear의 새 thread 분리, 디렉터리를 hook 실행 파일로 오인하는 문제는 앞선 [장애 수정 보고서](../gpt-incident-repair-20260915/report.md)에 근거와 검증을 기록했다. 이번 schema 오류와 같은 원인으로 합치지 않는다.

과거 실제 `agent_summary` scope mismatch의 정확한 내부 거부 지점은 아직 재현되지 않았다. synthetic summary 성공과 실제 오류 해결을 구분한다. 새 코드의 비민감 reason 진단은 원인을 좁히기 위한 것이며 우회 처리가 아니다.

## Evidence — 이번 최종 검사 명령과 출력

환경 정리 후 실행한 핵심 패키지 전체 race 검사:

```text
unset CLAUDECODE CLAUDE_CONFIG_DIR ANTHROPIC_BASE_URL ANTHROPIC_AUTH_TOKEN ANTHROPIC_API_KEY MOAI_GPT_LIVE MOAI_GPT_LONG_WAIT_LIVE && go test -race ./internal/codextools ./internal/codexbridge ./internal/codexapp ./internal/gateway/... -count=1
ok  github.com/modu-ai/moai-adk/internal/codextools 1.314s
ok  github.com/modu-ai/moai-adk/internal/codexbridge 18.318s
ok  github.com/modu-ai/moai-adk/internal/codexapp 4.258s
ok  github.com/modu-ai/moai-adk/internal/gateway 15.370s
ok  github.com/modu-ai/moai-adk/internal/gateway/auth 7.320s
ok  github.com/modu-ai/moai-adk/internal/gateway/conversation 4.944s
ok  github.com/modu-ai/moai-adk/internal/gateway/opaque 3.268s
ok  github.com/modu-ai/moai-adk/internal/gateway/receipt 5.540s
ok  github.com/modu-ai/moai-adk/internal/gateway/translate 6.784s
```

출력 원문: `/tmp/moai-gpt-structural-final-core.log`.

```text
go test -race ./internal/cli -run 'TestProductionGatewayCorrectsInvalidAgentAfterStreamingText|TestProductionGatewayClearStartsDistinctAttestedThread|TestManagedGPT|TestGPTAppServer' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli 3.204s

go vet ./internal/codextools ./internal/codexbridge ./internal/codexapp ./internal/gateway/... ./internal/cli
[출력 없음, exit 0]
```

CLI도 위와 같은 compound unset 환경에서 실행했다. 출력: `/tmp/moai-gpt-structural-scoped-cli.log`, `/tmp/moai-gpt-structural-vet.log`.

넓힌 CLI 검사에서는 다음 실패를 실제 관측했다. 성공한 필터 결과로 이를 숨기지 않는다.

```text
go test -race ./internal/cli -run 'TestProductionGateway|TestManagedGPT|TestGPTAppServer' -count=1
--- FAIL: TestProductionGatewayFactoryRejectsUnauthenticatedRequest (8.01s)
    gateway_product_binding_test.go:98: gateway private configuration or verified dependencies unavailable
FAIL
```

단독 실행도 같은 초기화 오류를 재현했다. 원인은 shared socket이 없는 fixture에서 `os.Executable()`이 MoAI CLI가 아닌 Go test binary를 실행한 것이었다. 기존 shared protocol fixture를 재사용하고 잘못된 session Bearer로 로컬401을 검사하도록 테스트 파일만 수정했다. 제품 인증 우회는 하지 않았다. 동일한 넓은 필터를 다시 실행했다.

```text
go test -race ./internal/cli -run 'TestProductionGateway|TestManagedGPT|TestGPTAppServer' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli 3.351s
```

앞의 최종 CLI 출력 파일은 이 재검증 결과로 갱신됐다. 최초 실패 출력은 본문에 보존했다.

로컬 후보 빌드도 완료했다. `go build -ldflags`로 version/commit/date/BuildID를 명시하고 `./cmd/moai`를 빌드했다. 설치 경로에는 복사하지 않았다.

```text
/tmp/moai-rc11-structural.A9FWOD/moai version
[v3.2.0-rc.11] [v3.2.0-rc.11-b45c81349-argument-repair-dirty] [built 2026-09-14T17:13:32Z]
shasum -a 256 /tmp/moai-rc11-structural.A9FWOD/moai
8c627028d8321a84bab142ce3f2b52929328a6bfbc3dfa38a41d11404b3978f1  /tmp/moai-rc11-structural.A9FWOD/moai
```

Proxy 자체 검사(우리 제품 검증과 별개):

```text
cargo test --offline --locked --lib providers::codex --quiet -- --test-threads=2
running 394 tests
test result: ok. 394 passed; 0 failed; 0 ignored; 0 measured; 490 filtered out; finished in 55.16s
```

## 운영 관측과 성능

| family | 관측 | 해석 |
|---|---|---|
| `04c54aae…` | 첫 byte 16,230ms 후 Agent schema 실패와 recovery_required | 실제 장애. 수정 소스는 미배포 |
| `d17bf130…` 첫 턴 | 첫 visible text 6.926초, 40 tool results 중 오류 0, 해당 턴 API400/502 0 | 해당 턴 정상. 전체 기능 보장 아님 |
| `d17bf130…` 보조 요청 | 17:05:17.610Z away_summary, first byte 2549ms, 17:05:21.240Z finished | 과거 agent_summary와 다른 유형 |
| `eefd29ac…` | agent_summary scope 오류5건(luna2, sol2, terra1), 17:08:25.875Z SessionEnd | private rejection5건과 대응. reason 없음. 실행 당시 gateway inode는 종료 후 확인 불가 |

새 세션 main은 sol/medium, thread `01a0a0e1-f86e-78c0-aeda-376e1f7a8a59`. native exec의 `setTimeout(resolve,30000)` 3회 실제 대기 합85.950초(30.009+30.027+25.914, 마지막 취소), App Server turn145.006초의59.3%였다. first-byte38.893초 구간에는 첫30.009초 sleep이 포함된다. native 로그 SQL의 해당4thread WARN/ERROR 조회는0행이지만 gateway 로컬 거부5건은 별개로 존재한다. GPT 오류가 없다는 뜻이 아니다.

private rejection: `/Users/goos/.moai/gpt-appserver/moai-rejections-ab92dd4aa4d1b7e8b27705de74c5b4a2777b7701dfe396b9422b0328c12f87fd.jsonl`. 긴 설정 allow-list는 원문 출력 대신 축약 처리했으므로 운영 native 로그의 전 바이트 독해라고 주장하지 않는다. 실행 타임라인과 이력·상태를 연결해 분석했다.

추가 이력 대조: sol summary의 17:07:14.035→17:07:32.281(18.246초), luna summary의17:07:43.961→17:07:59.418(15.457초)은 자식 도구 대기와 겹치고 거부 시각이 다음 Bash tool_use 생성 시각과 맞물린다. 다만 실제 HTTP body/headers와 실패 summary 입력이 저장되지 않아 classifier 실패·owner 충돌·pending-input 검사 중 어디인지 확정할 수 없다. transcript에서 요청을 추측 복원하여 동일 재현이라고 하지 않는다. 최신 후보의 비민감 private reason이 다음 재현에서 필요한 최소 추가 근거다.

이 수치로 “정상 TPS”를 판정하지 않는다. 첫 byte는 SSE 시작 이벤트일 수 있다. 빠른 연속 tool segment도 새로운 모델 추론 속도를 뜻하지 않는다. 도구 실행·사용자 대기·reasoning을 포함한 전체 시간으로 decode TPS를 계산하지 않는다.

## 다음 구조 개선의 수락 기준 — 구현 전 확인 대상

1. **오류 계약:** 모델 입력 오류, 권한/이력 오류, 일시적 전송 오류, 취소, 자원 한도를 구분한다. 새 실패 종류는 재현 fixture부터 추가한다.
2. **소유권 계약:** main/child/sibling/summary/clear별 owner와 generation을 교차시켜, 오래된 완료가 새 상태를 바꾸지 않는지 검사한다. 기존 receipt 검증을 단순화 명목으로 제거하지 않는다.
3. **스트림 계약:** 초기 delta가 종료 전에 도착하고, 오류 후 성공 terminal이 없으며, 부분 출력 이후 upstream 재실행과 tool 중복이 0인지 검사한다. 이미 있는 테스트를 재사용하고 누락 fixture만 보강한다.
4. **성능 계약:** 요청 수신, upstream 시작, 첫 의미 출력, tool 대기, 완료, usage를 같은 요청에 연결한다. 계측 없는 구간은 미상으로 기록한다. 원본의 전체 버퍼링이나 11회 빈 응답 재시도를 그대로 도입하지 않는다.
5. **운영 수락:** 수정 후보 빌드로 실제 Claude Code→App Server→GPT의 main/agent/summary/cancel/clear를 검증한 뒤 설치본 교체를 판단한다. 테스트 성공 전에 전체 호환을 선언하지 않는다.

대규모 엔진 재작성은 아직 결정하지 않는다. 현재 Engine/Registry/Store/stream helper를 재사용하는 최소 변경부터 한다. 새 구조 변경에 앞서 실제 누락 계약과 재현 근거를 제시하고 확인받는다.

## Gaps

- 실제 실패한 agent_summary의 완전 재현과 내부 거부 지점 확인 미완료.
- 최신 인자 교정 경로의 실제 GPT 자기 교정 운영 수락 미실행.
- 넓힌 CLI 검사 초기화 실패는 fixture 수정 후 동일필터로 통과했다. 전체 lint PASS가 아니며 이전 lint 60건도 별도 기록되어 있다.
- 전체 MoAI 코드, 전체 공급자 실서버, UI 전 기능, 이미지 2.5, 구독 ToS 판정은 이 보고서의 검증 범위 밖이다.
- 보고서의 EOF 독해와 실제 테스트 실행은 서로 다른 증거다. 상류 테스트 assertion이 약한 경우를 별도 문서에 표시했다.

## Residual-risk

부분 스트림·취소·프로세스 재시작·공급자 프로토콜 변경·동시 세션 조합은 추가 실패를 만들 수 있다. 설치된 기존 바이너리에는 새 코드가 없으므로 새 창을 열었다는 사실만으로 수정본을 검증한 것이 되지 않는다. App Server 사용 자체도 Claude Code의 모든 기능 지원이나 약관상 허용을 자동 보증하지 않는다.
