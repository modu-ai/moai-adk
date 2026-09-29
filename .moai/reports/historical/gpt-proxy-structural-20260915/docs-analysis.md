# 문서 사이트·README·CHANGELOG 전수 읽기 보충 보고

## Claim

`b642a6ee1a5fff8ebc9bedfc4b438ddfa3ae8f8e` 체크아웃에서 요청된 Markdown, README, CHANGELOG, 문서 사이트 mjs/ts/astro/css/public JS를 inventory하고 EOF까지 읽었다. **27개 일반 파일 2,810줄**이며, CHANGELOG를 가리키는 심볼릭 링크를 별도 경로로 세면 **28경로 3,354줄**이다. 심볼릭 링크는 별도 원문이 아니므로 고유 소스 줄 수에 중복 합산하지 않는다.

문서가 설명하는 핵심은 Claude Code 클라이언트 환경을 유지하면서 Anthropic 트래픽을 변환하는 로컬 proxy다. 문서 자체도 완전한 프로토콜 동등성을 보장하지 않고, 필드 생략과 제공자별 제한을 명시한다. 따라서 이 프로젝트를 참고하는 것과 MoAI에서 모든 Claude Code 기능을 검증했다는 주장은 구분해야 한다.

## Evidence / Baseline-attribution

체크아웃: `/tmp/moai-proxy-structural.IArJmM/source`.

```text
git ls-files '*.md' '*.mjs' '*.ts' '*.astro' '*.css' 'docs/public/*.js' 'README*' 'CHANGELOG*'
[아래 장부의 28경로]

wc -l README.md CHANGELOG.md docs/astro.config.mjs docs/public/image-zoom.js docs/src/content.config.ts docs/src/components/HeaderLinks.astro docs/src/styles/site.css docs/src/content/docs/*.md docs/src/content/docs/*/*.md
    3354 total

ls -l docs/src/content/docs/reference/changelog.md
lrwxr-xr-x@ 1 goos  wheel  27 Sep 15 01:55 docs/src/content/docs/reference/changelog.md -> ../../../../../CHANGELOG.md
```

각 일반 파일은 `cat`으로 전체를 읽었다. 최초 묶음 출력에서 잘린 `providers/codex.md`는 전체 재독하여 공백을 없앴다. symlink는 대상 CHANGELOG 1–544줄을 전부 읽었다. 이미지·lock·binary는 제외했다. Rust 구현 전체 읽기 장부는 다른 보고서에 있으며 여기서 중복 계산하지 않는다.

## 문서에서 직접 확인한 운영 계약

### 세션과 모델을 나눠 설명한다

`configure-claude-code.md`, `switching-models-and-backends.md`는 base URL/auth가 프로세스 시작 시 결정되고, 같은 proxy 세션 내 모델 변경은 요청별 routing이라는 점을 구분한다. 이는 `moai gpt`를 하나의 주 대화 세션으로 유지하려는 요구와 맞는다. 문서의 small-fast model은 제목·백그라운드 요청용이지 plan/run/routine 등 제품 역할별 메인 세션을 강제로 만드는 설계가 아니다.

`CLAUDE_CODE_DISABLE_NONSTREAMING_FALLBACK=1`의 이유도 명시한다. 이미 일부 스트림·도구 실행이 진행된 요청을 non-stream으로 다시 요청하면 도구 중복이 생길 수 있다는 운영 계약이다. MoAI에서는 설정 존재뿐 아니라 부분 출력 후 실패 시 실제 재실행 여부를 시험해야 한다.

### 초기 응답·사용량 측정과 가짜 속도를 구분한다

monitor 문서는 matched upstream timing과 cumulative usage sample로 throughput을 계산한다고 설명한다. CHANGELOG v0.1.17은 완전한 usage/timing이 없는 요청을 속도 계산에서 제외한다고 기록한다. 이 원칙은 MoAI의 TPS 보고에도 유효하다. 도구 실행·대기·요약을 포함한 전체 elapsed로 output token을 나눈 값을 순수 decode TPS로 발표하지 않는다.

`count_tokens`는 compaction 판단용이며 billing count가 아니라고 HTTP API와 limitations 문서가 명시한다. `[1m]`은 Claude Code의 local hint이지 제공자 context 확장 기능이 아니다. 문서의 구독 context 수치는 외부 게시물을 인용한 값이고 이번 조사에서 현재 서비스 값으로 독립 검증하지 않았다.

### 인증·약관·노출 범위는 별개다

README와 limitations는 unauthenticated local listener, 기본 loopback, non-loopback 보호 필요, 비공식 클라이언트 계정 위험을 명시한다. Codex 문서는 다른 harness 사용에 관한 OpenAI 인사 게시물을 인용하지만, limitations는 그 공개 발언이 향후 정책·계정 처리를 보장하지 않는다고 제한한다.

따라서 이 문서가 ToS 허용 확인서인 것은 아니다. 이번 보충 작업은 로컬 문서 원문 분석이며 링크된 소셜 게시물, 공급자 약관, 실제 entitlement를 재검증하지 않았다. 법률·공식지원 최종 판정에 이 로컬 문서만 사용하면 안 된다.

### 이미지 기능의 실제 문서 제한

Codex/HTTP API/limitations는 opt-in Images API를 **gpt-image-2**로 제한한다고 일관되게 설명한다. generation과 JSON/multipart edit는 있으나 mask, URL source/output, variation은 지원하지 않는다. 내부 ChatGPT 인터페이스라 공용 Platform API의 호환성 보장이 없다고도 적혀 있다.

이는 사용자 지정 `gpt-image-2.5`를 그대로 지원한다는 증거가 아니다. MoAI 이미지 모델은 사용자 요구를 바꾸거나 proxy 이름을 조용히 치환하지 말고, 실제 AppServer/공식 기능 노출과 계약을 확인한 뒤 구현해야 한다. 이번 조사에서는 이미지 호출을 수행하지 않았다.

## 문서와 코드 대조 시 주의할 부분

아래는 코드와 문서의 정적 대조 결과이며 실제 서비스 실패 재현 판정은 아니다.

| 문서 설명 | 이번에 읽은 코드와의 관계 | 보고 시 처리 |
|---|---|---|
| Cursor는 설치 JS bundle에서 protobuf와 dynamic catalog를 읽고 effort variant를 선택 | 현재 Cursor provider는 수동 protobuf/prost 정의와 prefix model 해석을 사용한다. provider 내 bundle getter 호출이나 effort 입력 처리를 확인하지 못했다 | 문서 기능을 현재 Rust 실행 기능으로 승격하지 않는다 |
| Cursor metadata로 conversation resume/mode 지정 | 현재 client는 run마다 새 conversation UUID를 만들고 mode는 model prefix에서 고른다 | metadata-resume 동작은 추가 기계 검증 전 미확정 |
| Kimi 문서는 한 wire model과 K2.6 alias 중심 | CHANGELOG v0.1.25와 읽은 model/request 코드는 K3 분기·별도 effort를 포함 | 문서의 모델 표를 현재 catalog의 완전 목록으로 쓰지 않는다 |
| 모든 제공자가 stream responses | Kimi/Cursor는 provider 반환이 BufferedSse이며 전체 upstream 수집 뒤 변환 | SSE 지원과 실시간 downstream 전달을 구분 |
| OpenCode Go는 auth status 없음 | 읽은 `OpenCodeCli::status` 구현은 key source를 출력한다. top-level CLI 노출 여부는 부모 담당 범위 | 구현 함수 존재만으로 CLI 접근 가능 여부를 단정하지 않는다 |

Cursor 확인용 명령은 다음과 같다. 이는 전수 읽기를 대신한 것이 아니라 대조한 심볼 위치 확인이다.

```text
rg -n 'agent_bundle|agentBundle|cursor_chat_id|cursorChatId|cursor_resume|cursorResume|cursor_mode|effort|catalog' src/providers/cursor src/config.rs src/registry.rs
src/config.rs:87:    #[serde(rename = "agentBundle")]
src/config.rs:88:    pub agent_bundle: Option<String>,
src/config.rs:949:pub fn cursor_agent_bundle() -> Option<String> {
[Cursor provider에서는 catalog 주석·model 해석과 관련 테스트만 표시]
```

## CHANGELOG에서 얻을 회귀 테스트 우선순위

과거 release note는 성공 실측이 아니라 테스트 후보를 찾는 자료로 사용한다.

- v0.1.38: Artifact tool schema / Claude Code 2.1.265+ 변화 — 필드 변환·optional 보존 계약.
- v0.1.36, v0.1.31: 도구 뒤 continuation, 직접 자식 agent별 독립 owner, 겹치는 compaction — MoAI의 상태 경계와 가장 관련 있다.
- v0.1.33, v0.1.31, v0.1.23: 조용한 스트림, setup 실패, 이전 socket, bounded error body — transport 회귀 후보.
- v0.1.16: 취소·대체된 prompt의 오래된 continuation이 후속 turn에 영향을 주지 않아야 한다.
- v0.1.15: optional tool parameter 보존과 강제된 인자 생성 방지.
- v0.1.39: completion 뒤 harmless keepalive 허용과 malformed tail/late failure 거부를 함께 확인해야 한다.

이 목록의 issue/PR 링크는 변경 기록의 출처 포인터이며 이번 보충 작업에서는 링크된 토론과 CI를 열지 않았다.

## 문서 UI 코드 분석

Astro/Starlight 구성은 sidebar를 명시적으로 관리하고 llms.txt plugin 및 LLM 링크를 제공한다. 이미지 확대는 native dialog와 button, aria-label을 사용한다. CSS는 밝은/어두운 테마, 반응형 rail, focus-visible을 정의한다. 이 코드는 문서 표시 기능일 뿐 gateway 실행에 관여하지 않는다. 브라우저 렌더·접근성·빌드 테스트는 수행하지 않았다.

## EOF 읽기 장부

각 행의 읽기 범위는 1–마지막 줄 전체다. 미독 구간은 없다.

| 경로 | 마지막 줄 |
|---|---:|
| README.md | 133 |
| CHANGELOG.md | 544 |
| docs/astro.config.mjs | 79 |
| docs/public/image-zoom.js | 72 |
| docs/src/content.config.ts | 7 |
| docs/src/components/HeaderLinks.astro | 24 |
| docs/src/styles/site.css | 275 |
| docs/src/content/docs/getting-started.md | 63 |
| docs/src/content/docs/how-it-works.md | 48 |
| docs/src/content/docs/index.md | 40 |
| docs/src/content/docs/providers/choosing-a-provider.md | 32 |
| docs/src/content/docs/providers/codex.md | 133 |
| docs/src/content/docs/providers/cursor-agent.md | 75 |
| docs/src/content/docs/providers/grok.md | 98 |
| docs/src/content/docs/providers/kimi.md | 52 |
| docs/src/content/docs/providers/opencode-go.md | 61 |
| docs/src/content/docs/reference/changelog.md | 544, CHANGELOG.md symlink |
| docs/src/content/docs/reference/command-reference.md | 90 |
| docs/src/content/docs/reference/compatibility-and-limitations.md | 93 |
| docs/src/content/docs/reference/configuration.md | 160 |
| docs/src/content/docs/reference/files-and-storage.md | 77 |
| docs/src/content/docs/reference/http-api.md | 170 |
| docs/src/content/docs/using/configure-claude-code.md | 68 |
| docs/src/content/docs/using/for-coding-agents.md | 79 |
| docs/src/content/docs/using/models-and-routing.md | 84 |
| docs/src/content/docs/using/monitor-tui.md | 63 |
| docs/src/content/docs/using/switching-models-and-backends.md | 79 |
| docs/src/content/docs/using/troubleshooting.md | 111 |

## Gaps / Residual-risk

문서의 링크 대상, 공식 약관, 실제 계정 권한, 이미지 생성, 사이트 빌드, 브라우저 동작, 공급자 runtime 기능은 이번 보충 범위에서 미검증이다. 문서의 기능 설명을 실제 호환성 PASS로 사용하지 않는다. 제품 변경·인증·배포 없이 보고서만 작성했다.
