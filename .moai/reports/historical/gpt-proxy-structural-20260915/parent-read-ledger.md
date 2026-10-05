# Proxy 구조 비교 — 주 담당 독해 기록

상태: 주 담당 Rust 26파일 15,921행 EOF 독해 완료. 전체 동작 검증 또는 전체 MoAI 저장소 감사 완료라는 뜻은 아니다.

원본: `/tmp/moai-proxy-structural.IArJmM/source`
원격: https://github.com/raine/claude-code-proxy
커밋: `b642a6ee1a5fff8ebc9bedfc4b438ddfa3ae8f8e`, Cargo 버전 `0.1.39`.
처음 확인한 이전 `/tmp/moai-raine-probe.Y1Qb5T/source`는 dirty였으므로 비교 원본으로 사용하지 않았다.

## 주 담당이 실제 출력으로 읽은 범위

| 파일 | 읽은 범위 |
|---|---|
| Cargo.toml, LICENSE | EOF |
| src/provider.rs | 1–126 EOF |
| src/request_identity.rs | 1–338 EOF |
| src/session.rs | 1–184 EOF |
| src/retry.rs | 1–82 EOF |
| src/anthropic/error.rs | 1–38 EOF |
| src/anthropic/mod.rs | 1–11 EOF |
| src/anthropic/schema.rs | 1–27 EOF |
| src/anthropic/sse.rs | 1–130 EOF |
| src/lib.rs | 1–22 EOF |
| src/server.rs | 1–2372 EOF |
| src/registry.rs | 1–450 EOF |
| src/main.rs | 1–321 EOF |
| src/paths.rs | 1–144 EOF |
| src/project.rs | 1–173 EOF |
| src/config.rs | 1–1346 EOF |
| src/logging.rs | 1–274 EOF |
| src/auth.rs | 1–625 EOF |
| src/traffic.rs | 1–845 EOF |
| src/openai_compat/mod.rs | 1–135 EOF |
| src/openai_compat/request.rs | 1–1370 EOF |
| src/openai_compat/response.rs | 1–712 EOF |
| src/openai_compat/stream.rs | 1–873 EOF |
| src/monitor.rs | 1–1629 EOF |
| src/monitor/mock.rs | 1–939 EOF |
| src/tui.rs | 1–2659 EOF |
| src/tui/layout.rs | 1–96 EOF |

다른 제공자·Codex·tests의 실제 EOF 범위는 각각 독립 담당 보고서에 기록한다. Rust 외 스크립트·문서는 별도 담당 장부에 기록한다. 생성된 의존성 lock과 바이너리·이미지의 본문은 코드 독해 범위에서 제외한다.

## 현재 구조 관측 — 런타임 정상 보장이 아님

- ProviderError는 authentication/permission/rate-limit/invalid-request/api를 분리한다. 우리 오류 분류와 비교할 계약이다.
- RequestMonitorGuard는 본문 완료, 실패, drop을 구별한다. 헤더를 반환한 시각의 request_completed 로그와 실제 body 완료는 다르므로 성능 비교 시 구별해야 한다.
- request-id 헤더를 응답에 연결하고 원본 ID가 있으면 보존한다. 우리 쪽 동등 동작 유무는 아직 확인하지 않았다.
- auto-review 요청은 별도 모델 라우팅 및 conversation identity 없음으로 전달한다. 일반 세션 affinity도 변경하지 않는다. 자동 요약의 직접 해결책이라고 단정할 수 없다.
- identity parser는 모호한 header를 stateless None으로 처리한다. MoAI의 private receipt 권한 검증을 이와 같이 완화하면 안 된다.
- `retry.rs`는 재시도할 HTTP status와 횟수/대기를 제한한다. 이미 공개된 출력 이후 replay 허용 여부는 Codex client/stream 분석 담당 결과와 연결해야 한다.
- proxy 원본에도 도구별 인자 추정 복구와 제공자별 정보 생략 정책이 있으므로 전체 복제 대상이 아니다.

## 새 운영 장애

family `04c54aae-7cd6-4753-a933-a584d40f6553`: 설치된 기존 summary-isolation 빌드의 gateway PID 65850. 첫 luna/max 요청 첫 byte 16,230ms, Agent의 필수 description 누락 후 stream transport 오류와 recovery_required 반복. 실제 스키마 required=[description,prompt] 및 호출 누락을 독립 담당이 확인하고 RED 재현했다. 모델 인자 오류를 세션 권한 오류와 분리하여 negative tool feedback으로 제한된 복구 기회를 주는 구현·통합 검사가 통과했다. 임의 description 생성은 하지 않는다. 설치본은 변경하지 않았다.

family `d17bf130-6c5f-4c4d-9a7c-e19a9d7fbf9b`: 기존 빌드 gateway PID 66491. 첫 턴은 정상 종료. 실제 main terra/low, 40개 tool result 중 오류 0, API 400/502 관측 0. 자동 요약 요청이 없었으므로 요약 호환성 증거는 아니다. 첫 visible text 6.926초; 13개 todo pr 조회가 순차 실행된 구간 약 20.5초. 전체 작업 시간을 순수 모델 생성 지연으로 볼 수 없다.

추가 관측: 같은 family의 17:05:17.610Z `away_summary` 요청은 first byte 2549ms 후 17:05:21.240Z finished로 기록되었다. 이것은 첫 턴 이후의 별도 관측이며 과거 `agent_summary` 실패가 해결됐다는 증거가 아니다.
