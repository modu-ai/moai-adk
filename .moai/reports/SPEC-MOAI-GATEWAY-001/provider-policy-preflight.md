# 공급자 정책 사전 확인

2026-09-11 부모의 후속 측정 메모다. 구현 또는 실제 공급자 호환성 PASS가 아니다.

## 관측

`internal/gateway/testdata/validation-capture/claude-code-2.1.268/request-*.json`을 Python json으로 읽고 정책 필드만 출력했다(exit 0). 이 파일은 과거 캡처의 비식별 픽스처이며 오늘 요구된 Opus 5·Sonnet 5 실계정 시험을 대체하지 않는다.

| 파일 | 관측한 정책 필드 |
|---|---|
| request-001 | `thinking.type=disabled`, `temperature=1`, 제목용 `output_config.format.type=json_schema` |
| request-002 | `thinking.type=enabled`, `budget_tokens=31999`, `clear_thinking_20251015`의 `keep=all` |
| request-004 | `output_config.effort=high`, 제목용 JSON schema |
| request-005 | `thinking.type=adaptive`, `output_config.effort=high`, `clear_thinking_20251015`의 `keep=all` |

공식 모델 문서를 이번 실행에서 열고 `effort supports`를 검색했다. [Astra](https://developers.openai.com/api/docs/models/gpt-6-astra)는 low/medium/high/xhigh/max를 명시한다. [Sol](https://developers.openai.com/api/docs/models/gpt-5.6-sol), [Terra](https://developers.openai.com/api/docs/models/gpt-5.6-terra), [Luna](https://developers.openai.com/api/docs/models/gpt-5.6-luna)는 none도 명시한다. 이 차이는 disabled를 모든 모델에서 같은 값으로 바꿀 수 있다는 근거가 없음을 뜻한다. 실제 구독 endpoint의 수용 여부는 아직 시험하지 않았다.

[Claude context editing](https://platform.claude.com/docs/en/build-with-claude/context-editing)의 thinking 설정 표와 예제에서 `keep=all`은 모든 thinking block 보존을 의미한다. 이를 지원하려면 gateway의 opaque reasoning 왕복도 보존되어야 한다. 단순 필드 삭제로 해당 계약을 충족했다고 판정하지 않는다.

공개 Codex 소스 `/tmp/openai-codex-audit.zbvNXB`의 기준은 `5a9eb14`다. `codex-rs/codex-api/src/common.rs:259`의 ResponsesApiRequest 필드와 `codex-rs/core/src/client.rs:828` 이후 구성 코드를 `sed`로 읽었다(exit 0). 후자는 `store:false`, `stream:true`, `include:["reasoning.encrypted_content"]`를 구성한다. 읽은 request 타입에는 `max_output_tokens`가 없다. 이것은 endpoint가 그 필드를 거절한다는 증거가 아니며 설치본과 동일한 코드라는 증거도 아니다.

## 후속 실측 조건

2026-09-11 오후 6시대에 위 OpenAI 공식 모델 페이지 네 개를 다시 열고 `context window`를 검색했다. Astra 페이지 882~884행, Sol 878~880행, Terra 878~880행, Luna 878~880행은 각각 문맥 1,050,000과 최대 출력 128,000을 명시했다. 이는 공개 API 모델 문서 값이다. Codex 구독 endpoint의 계정별 한도나 실제 수용을 측정한 값은 아니며, 이 문서 관측만으로 생산 catalog capability를 활성화하지 않았다. 네 페이지의 reasoning effort 목록도 기존 기록과 같았다.

M0 후속 관측기에는 메시지·도구 내용 없이 thinking 유형, budget, output effort, JSON schema 요청 여부, max_tokens, stream, 도구 개수만 남기는 `policy_metadata`를 추가했다. 합성 내용 표지와 JSON schema 내부 표지가 결과에 포함되지 않는 단위 확인은 `POLICY_METADATA_SYNTHETIC_OK`였다. 실제 Opus 5·Sonnet 5 요청의 필드값은 19시 이후에 관측한다.

- 오후 7시 이후 Opus 5·Sonnet 5 실제 요청에서 정책 필드를 다시 수집한다.
- 구독 endpoint에서 `max_output_tokens`, nonstream, JSON schema의 실제 수용을 각각 판정한다. 오류를 없애려고 출력 상한을 조용히 제거하지 않는다.
- Astra의 disabled 요청에는 등가 매핑이 확인되지 않았다. 실제 제목 요청과 본 대화 요청을 관측하고 지원 정책을 정한다. 다른 모델로 몰래 보내지 않는다.
- ordinary projection fixture의 통과는 실제 Claude 요청 전체의 통과가 아니다. reasoning·format·context 정책을 포함한 원본 형식의 수용과 응답을 추가 검증한다.
- 내려받아 둔 공식 settings-reference의 `fallbackModel` 절(932~951행)을 직접 읽었다. 모델 배열을 설정하면 실패 시 다른 모델을 시도하며, 가장 높은 설정 파일의 배열 전체가 쓰이고 `--fallback-model`은 이 값보다 우선한다고 설명한다. gateway 자체의 재시도 금지만으로 Claude 클라이언트의 공급자 변경까지 막았다고 판정하지 않는다. 기존 fallback 설정이 있는 격리 fixture에서 503 응답 후 실제 요청 ID를 수집하고, 빈 fallback 배열 overlay와 명시 CLI 처리의 효과를 시험한다. 아직 실제 우회가 관측된 것은 아니다.

## 송신과 로그아웃 경계

설치된 Go의 `/opt/homebrew/opt/go/libexec/src/net/http/request.go:732`에서 `WroteHeaders()` 호출 뒤에 `bw.Flush()`가 나오는 순서를 직접 읽었다. `:584`의 WroteRequest는 함수 반환 시 defer로 호출된다. HTTP/2 경로도 `h2_bundle.go:8800` 이후 헤더와 body 쓰기 뒤 WroteRequest를 호출한다. 아직 이것만으로 구현의 동시성 보장을 주장하지 않는다. AUTH 작성자에게 실제 net.Conn 쓰기를 막은 RED와 logout 경합 시험을 요청했으며, 유한한 request 전체 쓰기 완료까지 잠그고 response SSE를 기다리기 전에 놓는 경계를 검증하기로 했다.

## 미검증과 잔여 위험

실제 Claude·OpenAI 요청은 이 메모 작성 중 실행하지 않았다. 각 공급자의 현재 계정별 모델 허용, 구독 전용 request 제약, opaque reasoning의 Claude Code 재전송, write callback과 취소의 실제 경합이 남아 있다. 이 메모는 SPEC의 정책 선택이나 출시 게이트를 대신하지 않는다.
