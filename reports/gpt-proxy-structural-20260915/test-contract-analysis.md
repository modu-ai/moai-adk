# claude-code-proxy 테스트 전체 독해와 MoAI 회귀 계약

작성일: 2026-09-15. 범위: 확보한 소스의 `tests/*.rs` 11개 파일, 9,883줄. 이 문서는 테스트 소스를 끝까지 읽어 확인한 **검증 의도와 assertion의 강도**를 다룬다. Cargo 실행 결과나 실제 구독 계정 운영 성공을 대신하지 않는다.

## Claim — 핵심 판단

가장 가치 있는 부분은 Rust 언어 자체가 아니라 **대화 소유자별 연속성, 재시도 가능 시점, 스트리밍 완료 판정**을 제품 HTTP 경로와 계측 가능한 가짜 upstream으로 검증하는 방식이다. 특히 `codex_agent_continuation.rs`는 메인·자식·중첩 자식·형제·동일 이름의 다른 세션을 구분하고, 오래된 완료가 최신 상태를 덮지 않는지 확인한다. MoAI가 우선 흡수할 대상은 이러한 상태 전이 계약이다.

반대로 이 테스트들이 통과한다고 해서 Claude Code의 모든 기능, App Server의 미완료 도구 호출, GPT 구독 약관, 실제 생성 속도, 이미지 모델 접근 권한이 입증되지는 않는다. 일부 테스트는 이름보다 assertion이 약하다. 따라서 전체 테스트 개수만으로 안정성을 판단하면 안 된다.

## Baseline-attribution — 조사 기준과 EOF 원장

소스 기준 경로: `/tmp/moai-proxy-structural.IArJmM/source`. MoAI 보고서 작업 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/gpt-session-repair`, 이번 확인의 HEAD는 `b45c81349`다. 소스 저장소의 커밋 식별은 상위 조사 보고서의 확보 기록에 따른다. 여기서는 확인하지 않은 upstream 커밋을 임의로 기재하지 않는다.

`cat` 또는 연속된 `sed -n` 범위로 아래 모든 줄을 읽었다. 파일명 검색만으로 독해를 대체하지 않았다. `foundation.rs`가 사용하는 `tests/fixtures/anthropic-message.json`과 `tests/fixtures/sse-basic.txt`도 끝까지 읽었다.

| 파일 | EOF 독해 범위 | 실제 assertion이 다루는 계약 | 해석의 한계 |
|---|---:|---|---|
| `cli.rs` | 1–119 | 버전·도움말·모델 목록·잘못된 명령 종료 코드·일부 인증 상태 CLI | 모델 목록이 실제 모델 호출 성공을 뜻하지 않음 |
| `codex_auth.rs` | 1–122 | 임시 저장소 인증 상태, 만료 시간, accountId 호환, 누락 상태 | 실제 로그인·토큰 갱신·구독 이용 허가 검증 아님 |
| `codex_compact_effort.rs` | 1–187 | compact effort 기본 상한 low, none/off/global 조합의 직렬화 | 품질·속도·비용 비교 실험 아님 |
| `foundation.rs` | 1–265 | 요청 fixture, SSE 프레이밍, OS 경로, 비밀정보 가림, 로그 상한, backoff, 메모리 인증 저장소 | 실제 계정이나 운영 로그의 모든 민감정보 패턴 검증 아님 |
| `public_codex_api_compat.rs` | 1–94 | 공개 Rust API의 기존 호출 형태·타입 호환 | 실행하지 않는 함수 참조도 포함; 네트워크 성공 근거 아님 |
| `codex_websocket.rs` | 1–369 | 직접 WebSocket 연결·프레임·시간 제한 실험 | 제품 클라이언트가 아닌 직접 연결이며 일부 assertion이 무효 또는 약함 |
| `codex_websocket_proxy.rs` | 1–478 | 제품 경로의 프록시 환경변수·CONNECT·인증 헤더·거부·fallback 경로 | 로컬 mock으로 네트워크 정책을 확인; 실제 공급자 가용성 검증 아님 |
| `cursor_native.rs` | 1–1549 | protobuf/Connect 요청·프레임·텍스트/usage·도구 pause/resume·허용 이름 | 실제 Cursor 이용 허가 및 App Server 도구 계약과 별개 |
| `server.rs` | 1–1576 | ingress identity, 오류 형식, 크기 제한, request-id, 모델 라우팅, 선택 기능 게이트, monitor | 일부는 FakeProvider 또는 거부 응답만 확인 |
| `codex_agent_continuation.rs` | 1–2112 | 실제 listener 경유 owner 격리·delta·소켓 재사용·stale 완료·재시도 후 연속성 | Responses 완료 기반 계약; 실행 중 App Server RPC는 별도 |
| `smoke_cutover.rs` | 1–3012 | HTTP/WS 제품 경로, compaction, 초기 응답, 재시도 경계, 취소, 명시적 실패, 로그 | mock 지연과 재시도 0초 override가 실제 처리량을 뜻하지 않음 |

## Evidence — 조사 명령과 관측 출력

전체 읽기는 `cat <파일>`과 `sed -n '<시작>,<끝>p' <파일>`로 수행했다. 큰 파일은 연속 범위를 나누어 읽었으며 `smoke_cutover.rs`는 `1–760`, `761–1510`, `1511–2250`, `2251–3012`까지 확인했다. 아래는 독해 후 다시 실행한 파일 원장 명령의 출력이다.

```text
$ wc -l /tmp/moai-proxy-structural.IArJmM/source/tests/*.rs
     119 /tmp/moai-proxy-structural.IArJmM/source/tests/cli.rs
    2112 /tmp/moai-proxy-structural.IArJmM/source/tests/codex_agent_continuation.rs
     122 /tmp/moai-proxy-structural.IArJmM/source/tests/codex_auth.rs
     187 /tmp/moai-proxy-structural.IArJmM/source/tests/codex_compact_effort.rs
     369 /tmp/moai-proxy-structural.IArJmM/source/tests/codex_websocket.rs
     478 /tmp/moai-proxy-structural.IArJmM/source/tests/codex_websocket_proxy.rs
    1549 /tmp/moai-proxy-structural.IArJmM/source/tests/cursor_native.rs
     265 /tmp/moai-proxy-structural.IArJmM/source/tests/foundation.rs
      94 /tmp/moai-proxy-structural.IArJmM/source/tests/public_codex_api_compat.rs
    1576 /tmp/moai-proxy-structural.IArJmM/source/tests/server.rs
    3012 /tmp/moai-proxy-structural.IArJmM/source/tests/smoke_cutover.rs
    9883 total
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/gpt-session-repair rev-parse --short HEAD
b45c81349
```

Cargo 테스트는 병렬 조사 담당자가 실행 중이므로 이 분석에서는 중복 실행하지 않았다. 아래의 “검증한다”는 문구는 **테스트 코드가 그 조건을 assertion으로 요구한다**는 뜻이며 이번 문서의 실행 PASS 선언이 아니다.

## 흡수할 강한 계약

### 1. Owner와 연속성의 분리

`codex_agent_continuation.rs:970`은 메인·자식·중첩 자식의 HTTP 요청을 실제 서버에 보내 각 소켓과 `previous_response_id`, 입력 delta를 확인한다. `:1088`은 같은 prefix를 가진 형제가 다른 owner의 continuation을 가져가지 못하도록 한다. `:1238`은 같은 agent ID라도 세션이 다르면 독립이어야 함을 요구한다.

`:1750`은 같은 owner의 이전 요청이 늦게 끝나더라도 최신 continuation을 덮지 못하도록 한다. `:1868`은 서로 다른 owner의 완료 순서를 교차시킨다. 이처럼 **요청 성공뿐 아니라 후속 요청의 실제 전송 내용까지 검사**하는 방식이 중요하다.

단, 잘못된 identity의 처리는 MoAI에 그대로 옮기면 안 된다. `server.rs:284` 및 continuation `:1321`은 identity가 잘못되거나 없을 때 **stateless 전체 문맥 요청**으로 처리한다. 이는 다른 owner의 캐시를 사용하지 않는다는 계약이지, MoAI의 private receipt나 App Server thread 권한을 생략해도 된다는 뜻이 아니다.

### 2. 자동 보조 요청은 메인 연속성을 오염시키지 않음

continuation `:1436`은 agent 헤더가 붙은 자동 보안 검토도 stateless로 처리하며 기존 agent의 후속 continuation이 유지되는지 확인한다. smoke `:1303`은 자동 검토 기본 모델과 명시적 override를 upstream 요청에서 확인한다.

MoAI에서는 자동 요약과 감사 요청의 실제 타입·이력·소유권을 각각 판정해야 한다. upstream의 자동 검토 판별을 가져와 모든 summary를 동일하게 취급하거나, 사용자가 선택한 메인 모델을 무조건 낮추는 것은 이 테스트가 정당화하지 않는다.

### 3. 재시도는 사용자에게 의미 있는 출력이 전달되기 전으로 제한

smoke `:1770–1838`, `:2029–2035`는 control event, 구조만 열린 메시지/도구, 잘못된 JSON/UTF-8 및 초기 EOF에서 회복을 검사한다. 반면 `:2041`은 텍스트가 전달된 뒤 overload가 나면 upstream 시도를 **1회로 유지**하도록 한다. `:2197`은 이미 보낸 텍스트와 읽기 오류를 보존하고, `:2247`은 도구 블록이 닫혔더라도 정상 terminal 없이 끝나면 성공으로 만들지 않는다.

MoAI에 필요한 구분은 더 세밀하다. **권한이 확인된 도구의 잘못된 인자에 대한 실패 응답 후 모델의 수정**은 같은 turn의 도구 피드백이다. 이미 실행한 요청 전체를 다시 전송하는 transport retry와 다르다. 최근 Agent 필수 인자 누락 문제를 처리할 때 이 둘을 섞으면 중복 실행 또는 502 오분류가 재발할 수 있다.

### 4. 재시도 예산·취소·실패 이유가 관측 가능

smoke `:1984`, `:2097`은 특정 상태/overload 재시도를 총 4회로 제한한다. `:1257`, `:2827`의 빈 완료 경로는 총 11회 후 503을 요구한다. `:2147`은 요청 취소 후 backoff가 끝나도 두 번째 upstream 시도가 없어야 함을 확인한다. `:1898`은 영구 정책 거부를 1회·400으로 유지한다.

“bounded”만으로 빠르다고 판단하면 안 된다. 11회 정책은 호출 수가 제한됨을 뜻할 뿐 체감 지연이 적다는 뜻이 아니다. 테스트의 `ZeroRetryDelayGuard`를 운영 지연으로 해석하지 않는다. MoAI는 오류 종류별 예산과 취소를 검증하고, 실행·권한 오류를 transport 재시도 대상으로 감추지 않아야 한다.

### 5. 첫 응답과 최종 완료를 따로 검증

smoke `:1679`는 upstream 완료를 막은 상태에서 HTTP 헤더와 초기 `message_start`·ping·텍스트가 먼저 도착함을 확인한다. `:2442`는 WebSocket terminal 이전에 텍스트 delta가 도착하고 `message_stop`은 아직 없어야 함을 요구한다. `:1852`는 텍스트 뒤 `response.incomplete`를 정상 `message_stop`이나 `max_tokens`로 꾸미지 않는다.

이 테스트 패턴은 MoAI의 스트리밍 버퍼링 회귀를 잡는 데 적합하다. 다만 이 로컬 mock의 시간 제한을 실제 GPT의 정상 TTFT 또는 초당 토큰 성능 기준으로 사용할 수는 없다.

### 6. 인증·프록시·로그의 안전한 관측

`codex_websocket_proxy.rs`는 HTTP_PROXY/ALL_PROXY/NO_PROXY와 CGI 환경 예외, CONNECT의 host/port, proxy authorization을 확인한다. 프록시 거부 시 직접 origin으로 우회하지 않아야 하며 오류 응답에 프록시 자격증명이 없어야 한다.

`foundation.rs:99`, `:118`은 로그 비밀정보 가림·크기 상한을 다룬다. smoke `:1585`, `:1633`, `:2288`, `:2974`는 upstream/downstream 및 reducer 오류 기록을 확인한다. 모니터링은 단순 문자열 검색보다 request-id, owner, turn, terminal 여부, 재시도 횟수, 첫 출력 시점을 연결할 수 있어야 한다.

## 이름보다 약한 테스트 — 검증 범위 주의

| 위치 | 관측한 assertion | 판단 및 보완 |
|---|---|---|
| `codex_websocket.rs:248` | `assert!(timeout.is_err() || timeout.is_ok());` | 어떤 결과든 참. idle timeout 회귀를 검출하지 못하므로 timeout 결과와 실제 제품 취소 상태를 명시해야 함 |
| `codex_websocket.rs:83`, `:324` | 연결 성공 분기에서만 후속 검증 | 실패 연결이 분기 밖에서 누락됨. 동시성/soak 성공 건수 및 모든 task 결과를 요구해야 함 |
| `codex_websocket.rs:107` 일대 | 직접 tungstenite 및 detached mock handler | 제품의 pool·retry·invalidation 보장을 이 테스트 이름만으로 주장할 수 없음. 강한 continuation 테스트와 구별해야 함 |
| `cursor_native.rs:241` | `CursorHttpClient::new()`가 panic하지 않음 | 이름은 URL 검증이지만 실제 URL equality assertion 없음 |
| `cursor_native.rs:890` 일대 | `status != 401 && status != 400` | 500도 조건을 만족함. HTTP200과 Anthropic JSON 본문을 확인해야 함 |
| `server.rs:716`, `:741` | NOT_IMPLEMENTED가 아닌지 확인 | 공급자 경로 연결용 smoke이며 성공적인 모델 응답 증거 아님 |
| `public_codex_api_compat.rs:1–94` | 함수 포인터·타입 API 호환 | source compatibility 보장과 runtime generation 보장을 구분해야 함 |

이 표는 **테스트 assertion의 확인된 한계**다. 이 한계만으로 제품 기능 자체가 실패한다고 단정하지 않는다. 예를 들어 Cursor에는 별도로 실제 HTTP 경로와 mock upstream을 연결하는 더 강한 `cursor_native.rs:905` 테스트가 있다.

## MoAI 필수 회귀 매트릭스

아래는 채택할 acceptance 계약이다. 이 표 전체를 MoAI에서 이미 실행 PASS 했다는 뜻이 아니다. 현 작업의 Agent 인자 회복 및 `/clear` fixture 실행 증거는 해당 수정 보고서에서 별도로 추적한다.

| 우선순위 | 시나리오 | 반드시 관측할 결과 | proxy 참고 |
|---|---|---|---|
| High | 텍스트 전송 후 Agent 필수 인자 누락 → 수정 호출 | schema 실패 피드백 1회, 잘못된 도구 0회 실행, 수정 도구 1회 실행, 같은 thread 완료, 400/502와 failed barrier 없음 | smoke `:2041`의 재전송 금지와 구분해 App Server 전용 검증 |
| High | 모르는 도구·다른 owner·잘못된 binding | 모델의 수정 가능한 인자 오류와 구분; 권한 거부 유지, 실행 0회 | Cursor 허용 이름 `:1333`보다 강한 authority 검사 필요 |
| High | 잘못된 인자 연속 반복 | 정해진 피드백 예산 이후 명시적 종료; 무한 루프·전체 turn 재생 없음 | smoke `:2097`의 bounded retry 패턴 |
| High | `/clear` 후 UUID 변경 | native clear 증명 있는 요청만 새 thread; 기존 thread 분리; 증명·CWD·header 불일치는 RPC 이전 거부 | continuation `:1238`, `:1321`; clear 증명 자체는 MoAI 전용 |
| High | 자동 요약·감사와 메인/Agent 교차 | 보조 요청이 메인 pending tools·이력·barrier를 변경하지 않음 | continuation `:1436` |
| High | 같은 owner 겹친 요청의 완료 역전 | 늦은 이전 완료가 최신 상태를 덮지 않음 | continuation `:1750` |
| High | 같은 prefix의 형제·다른 session의 동일 agent ID | 각 thread/도구 결과/이력 독립; owner 사이 재사용 금지 | continuation `:1088`, `:1238`, `:1868` |
| High | transport EOF: 출력 전/텍스트 후/도구 실행 후 | 부작용 없는 구간만 제한적 재시도; 이후 원인 보존·중복 실행 0 | smoke `:1813`, `:2041`, `:2247` |
| High | 도구 대기 중 cancel·연결 단절 | RPC와 pending 상태 일관성, 다음 요청에 조용한 상태 오염 없음 | App Server 전용; proxy request abort `:2147`만으로 대체 불가 |
| High | 정상 terminal 없는 텍스트/닫힌 도구 | 성공으로 마감하지 않음; 오류 이유와 기존 출력 보존 | smoke `:1852`, `:2197`, `:2247` |
| Medium | 긴 응답 스트리밍 | 첫 의미 있는 출력과 최종 완료 시각 분리, heartbeat만 TTFT로 집계하지 않음 | smoke `:1679`, `:2442` |
| Medium | cache/delta 회복 | 검증된 동일 owner만 delta 사용; 안전한 full-context 회복이 권한 우회를 만들지 않음 | continuation `:1510`, `:2022` |
| Medium | 모델·effort 변경과 자동 작업 | 사용자 설정의 의도 보존; 모델 ID와 effective effort를 별도 기록 | compact effort 및 smoke `:1303` |
| Medium | 실제 성능 계측 | queue, upstream TTFT, 생성 구간 tokens/s, 도구 실행 시간, 전체 응답 시간 분리 | 로컬 mock 시간 제한은 운영 성능 증거 아님 |
| Medium | 이미지 생성·편집 | 지정 image 모델 실제 dispatch·결과 파일·사용량·취소·권한 검증 | server `:806–977`은 게이트/오류 검증 중심 |

## 추가 독해 — 실행·빌드·배포 코드

추가 지시로 다음 파일도 모두 EOF까지 읽었다. 실행하지 않고 내용을 분석했다. 아래 11개 파일은 1,041줄이며, 실제 check 명령을 해석하기 위해 `checkle.toml`도 전부 읽었다.

| 파일 | EOF 범위 | 확인한 실행 계약 |
|---|---:|---|
| `scripts/cua-monitor-demo` | 1–115 | HEAD archive에 dirty/untracked 변경을 덧씌운 임시 소스, Docker ARM64 debug build, 기존 CuaBot 컨테이너로 복사 후 Kitty demo 실행 |
| `scripts/debug-proxy` | 1–147 | 기존 인증 저장소 사용, 임시 state, random localhost port와 health 확인, trap으로 자식 프로세스 정리, prompt/tool 원문 capture 보관 경고 |
| `scripts/install-git-hook-shims` | 1–38 | Git common hooks 경로에 shim 설치, 실행 시 현재 worktree의 tracked hook 선택 |
| `scripts/install.sh` | 1–269 | OS/arch 감지, release archive 및 checksum 다운로드, 임시 파일→rename 설치, macOS ad-hoc signing, 설치 후 `--version` 실행 |
| `hooks/pre-commit` | 1–2 | `checkle pre-commit` 위임 |
| `justfile` | 1–82 | checkle 그룹 실행, CI에서 tracked dirty diff 거부, install/dev/docs/release 보조 명령 |
| `flake.nix` | 1–66 | 4개 Unix 시스템 패키지 정의, Cargo.lock 사용, `doCheck = false` |
| `.github/workflows/ci.yml` | 1–24 | Ubuntu stable Rust의 `just check-ci` |
| `.github/workflows/nix.yml` | 1–27 | Ubuntu Nix build, 실패 시 최대 5회 재시도 |
| `.github/workflows/docs.yml` | 1–53 | Bun frozen lockfile로 docs build 후 Pages artifact/deploy |
| `.github/workflows/release.yml` | 1–218 | Ubuntu `cargo test --locked` 선행, 6개 OS/arch release build와 버전 확인, archive/checksum 업로드, release 및 Homebrew tap 갱신 |

`checkle.toml`의 실제 all 그룹은 `cargo fmt --all -- --check`, `cargo clippy --all-targets ... -D warnings`, `cargo build --all`, `cargo test --all`이다. 따라서 CI “Full checks”는 해당 그룹을 뜻한다. CI 파일이 존재한다는 사실은 특정 커밋에서 실행·통과했다는 증거가 아니다.

주의할 점은 다음과 같다.

- Nix build는 `doCheck = false`이므로 테스트 통과 판정으로 재사용하면 안 된다. Release의 전체 테스트는 Ubuntu에서 수행하며 각 플랫폼의 바이너리 검증은 주로 `--version`이다. 이를 모든 플랫폼의 실제 대화 기능 검증으로 확대할 수 없다.
- Installer는 checksum 도구가 둘 다 없으면 검증을 건너뛰는 warning 이후에도 `Checksum verified`를 출력한다. 소스로 확인한 메시지 정확성 문제다. MoAI 배포 검증에서는 검증 도구가 없을 때 성공으로 표시하면 안 된다.
- Installer의 unsupported-platform 안내에 “build from source with Bun”이 남아 있으나 실제 justfile/release는 Cargo 빌드다. 이 안내는 실제 빌드 경로와 불일치한다.
- `debug-proxy`는 인증을 새로 격리하지 않고 기존 공급자 인증을 사용한다. state가 임시 경로라는 것과 계정·요금·데이터가 분리된다는 것은 다르다. 원문 traffic capture의 민감성 경고는 MoAI에서도 유지해야 한다.
- `cua-monitor-demo`는 dirty와 untracked 파일까지 컨테이너로 전달한다. 편의 기능이지 재현 가능한 clean release 증거가 아니며, 사용자 작업물을 외부 환경에 전송할 권한도 별도로 확인해야 한다.
- `install-dev`의 전역 symlink, hook shim 설치, release/tap push는 상태 변경이다. 본 분석에서는 실행하지 않았다.

### MoAI와 직접 대조한 계약 두 가지

1. `internal/codexbridge/engine.go:523` 이후는 operation 시작 후 실패를 불확실한 결과로 간주하고 barrier를 저장하며 **App Server RPC 또는 tool response를 자동 재시도하지 않는다**. 반면 proxy smoke `:1813`, `:2029`는 의미 있는 출력 이전의 EOF/잘못된 JSON에서 재시도한다. 이는 확인한 설계 차이다. MoAI의 회복 가능성 개선 후보이지만, App Server가 요청을 받았는지 모르는 상태에서 proxy 방식으로 그대로 재시도하는 것은 안전하지 않다. 부작용 이전임을 입증하는 계약이 없는 현재 상태를 “재시도 누락 버그”라고 단정하지 않는다.
2. MoAI `Engine.StepStream`은 `phase == waiting`에서 입력이 없고 pending 개수·ID·내용 타입이 정확해야 한다고 요구한다. proxy continuation `:2022`는 완료된 Responses 함수 호출 후 canonical JSON이 다른 tool result delta를 검사한다. **미완료 도구 barrier의 result count 오류 회복과 완료된 response의 delta 최적화는 서로 다른 계약**이다. upstream 테스트 PASS를 MoAI pending barrier 검증 PASS로 대신할 수 없다는 점을 코드로 대조했다.

## Gaps — 이번 분석이 입증하지 않은 것

- 본 담당자는 Cargo 실행을 하지 않았다. 별도 실행 담당자의 로그를 확인하지 않은 채 PASS 수를 넣지 않았다.
- 범위는 `tests/*.rs` 11개와 필요한 2개 fixture다. `src` 내부 unit test 전체의 부재나 충족 여부를 이 문서만으로 판단하지 않는다.
- 실제 GPT/Claude/Cursor 구독 계정 로그인·토큰 갱신·이미지 생성·과금·ToS 적합성은 이 테스트로 입증되지 않는다.
- Claude Code 실제 바이너리의 permission UI, hooks, MCP, Agent 비동기 알림, factory lane 전 과정, 장시간 부하와 운영 복구는 이 mock 기반 계약과 별도다.
- 이번 독해는 upstream 코드 수정이나 MoAI 제품 변경을 포함하지 않는다. 저장된 보고서만 추가한다.
- MoAI 저장소 전체 테스트 판정은 통합 브랜치 CI 실행이 담당하며 이 보고 시점에는 PENDING이다.

## Residual-risk — 적용 시 남는 위험

Responses의 `previous_response_id` 캐시를 잃은 상황과 App Server가 실제 도구 응답을 기다리는 상황은 다르다. 전자의 full-context retry를 후자에 그대로 적용하면 이미 실행된 도구가 다시 실행되거나 소유권 경계가 흐려질 수 있다. 가장 안전한 적용 순서는 **오류 분류 → owner/turn 상태 전이 → 출력 및 실행 경계별 재시도 → 계측 → 실제 운영 검증**이다.

마지막으로, 네트워크 fixture의 성공과 완성도 높은 assertion은 필요한 안전망이지만 충분조건은 아니다. 배포된 바이너리 식별, 실제 세션의 request/turn 연결, 정상 terminal, 사용자 작업 결과까지 확인해야 운영 완료라고 말할 수 있다.
