---
title: 프롬프트 캐싱 — 개념과 Claude Code에서의 동작
weight: 30
draft: false
---

프롬프트 캐싱은 요청의 **앞부분(접두사)이 직전 요청과 동일할 때, 그 부분을
다시 처리하지 않고 재사용**하는 API 기능입니다. 캐시에서 읽은 토큰은 기본
입력 단가의 **0.1배**로 청구되므로, 반복되는 컨텍스트(시스템 프롬프트,
프로젝트 지침, 대화 이력)가 클수록 절감 효과가 커집니다. MoAI-ADK
토크노믹스의 "컨텍스트 다이어트" 축이 항시 로드 컨텍스트를 줄이는 쪽이라면,
프롬프트 캐싱은 남은 컨텍스트를 싸게 재사용하는 쪽입니다.

{{< callout type="info" >}}
**쉬운 비유** — 매 턴마다 모델은 대화 전체를 처음부터 다시 읽습니다.
캐싱은 "앞부분은 아까 읽은 그대로네요" 하고 건너뛰는 책갈피입니다.
앞부분이 한 글자라도 바뀌면 책갈피가 무효가 되어 그 지점부터 다시 읽습니다.
{{< /callout >}}

## 핵심 개념 (API 공통)

- **접두사 매칭**: 캐시 적중에는 중단점까지의 내용(도구 정의·시스템
  프롬프트·메시지 이력 포함)이 **100% 동일**해야 합니다. 공백 하나만 달라도
  그 지점 이후는 전부 재계산됩니다.
- **가격 배수** (기본 입력 단가 대비): 캐시 기록 5분 TTL **1.25배** ·
  1시간 TTL **2배** · 캐시 읽기 **0.1배**.
- **TTL(수명)**: 기본 5분, 옵션 1시간. TTL 안에서 재사용될 때마다 수명이
  무료로 연장됩니다.
- **모델별 최소 캐시 토큰**: 이보다 짧은 접두사는 캐시되지 않습니다
  (오류 없이 일반 처리). 예: Fable 5 = 512, Opus 4.8·Sonnet 5 = 1,024,
  Opus 4.7 = 2,048, Haiku 4.5 = 4,096 토큰.

## Claude Code 사용자라면 — 캐싱은 자동입니다

Claude Code는 프롬프트 캐싱을 **자동으로 관리**합니다. `cache_control`을
직접 설정할 필요도, 설정할 방법도 없습니다. 공식 문서 기준 동작은 다음과
같습니다:

- **TTL 자동 선택**: 구독 플랜(Pro/Max/Team/Enterprise)에서는 1시간 TTL을
  자동 요청합니다(플랜 요금에 포함되므로 추가 비용 없음). API 키·클라우드
  제공자 경유 시에는 기본 5분이며, `ENABLE_PROMPT_CACHING_1H=1`로 1시간을
  옵트인할 수 있습니다.
- **환경변수 제어**: `FORCE_PROMPT_CACHING_5M=1`(5분 강제),
  `DISABLE_PROMPT_CACHING=1`(전체 비활성 — 디버깅 용도 외 비권장),
  모델별 `DISABLE_PROMPT_CACHING_OPUS` 등.
- **요청 구성 최적화**: Claude Code는 잘 안 바뀌는 내용(시스템 프롬프트 →
  프로젝트 컨텍스트 → 대화)이 앞에 오도록 요청을 배열해 접두사 적중률을
  높입니다.

### 캐시를 무효화하는 행동 (한 턴 느려지고 비싸짐)

- 모델 전환(`/model`) · effort 변경(`/effort`) — 모델·effort별로 캐시가
  분리되어 있습니다
- MCP 서버 연결/해제(도구 정의가 접두사에 로드된 경우)
- 도구 전체 거부(deny) 규칙 추가/제거
- `/compact`(대화 이력이 요약으로 교체됨) · Claude Code 업그레이드 후 첫 턴

### 캐시를 유지하는 행동

- 저장소 파일 편집(읽을 때만 대화에 추가됨) · 스킬/커맨드 호출 ·
  권한 모드 전환 · `/rewind`(이미 캐시된 접두사로 되돌아감) ·
  CLAUDE.md 중간 수정(캐시는 유지되지만 **변경도 적용되지 않음** — 다음
  `/clear`·재시작에 반영)

### 적중률 확인

- 응답의 `cache_creation_input_tokens`(기록) / `cache_read_input_tokens`(읽기)
  비율이 지표입니다 — 읽기가 높을수록 잘 작동 중.
- MoAI statusline의 cache_hit 세그먼트로 세션 중 실시간 확인이 가능합니다.
- 기록이 턴마다 계속 높다면 위의 "무효화하는 행동" 중 무언가가 접두사를
  바꾸고 있다는 신호입니다.

## API 직접 호출 시 (참고)

아래는 **Anthropic API를 직접 호출하는 개발자**에게만 해당하는 사용
예시입니다. Claude Code 사용자는 해당하지 않습니다.

```python
# API 직접 호출: 안정적인 시스템 프롬프트에 캐시 중단점 배치
response = client.messages.create(
    model="claude-opus-4-8",
    max_tokens=1024,
    system=[{
        "type": "text",
        "text": "안정적인 시스템 프롬프트...",
        "cache_control": {"type": "ephemeral", "ttl": "1h"}
    }],
    messages=[{"role": "user", "content": user_query}]
)
```

원칙은 하나입니다: **중단점은 매 요청 바뀌는 데이터(질문, 타임스탬프) 앞의
마지막 안정 블록에** 둡니다. 손익분기는 요청 2개 — 첫 요청의 기록 프리미엄은
TTL 내 두 번째 요청의 0.1배 읽기로 회수됩니다.

## MoAI cache.yaml의 적용 범위

`.moai/config/sections/cache.yaml`(`enabled`, `session_ttl`)은 **MoAI가 자체
SDK 래퍼 경로로 Anthropic API를 직접 호출할 때의 cache_control 주입**에만
적용됩니다. **Claude Code 세션의 캐싱과는 무관합니다** — Claude Code의
캐싱은 위 섹션대로 런타임이 자동 관리하며 MoAI가 개입할 수 없습니다.

> **GLM 백엔드**: z.ai(GLM)는 콘텐츠 유사도 기반 **암묵적 캐싱**을 사용하므로
> MoAI는 GLM 경로에 `cache_control`을 주입하지 않습니다.

## 요약

- **Claude Code 사용자**: 아무것도 설정할 필요 없음. 모델/effort 전환과
  `/compact`를 작업 사이 자연스러운 경계에서만 수행하면 적중률이 유지됩니다.
- **API 직접 호출자**: 안정 블록에 중단점, 요청 2개 이상일 때만 1시간 TTL.
- **모니터링**: statusline cache_hit 세그먼트 + `cache_read/creation` 토큰 비율.

**출처 (공식 문서):**

- [How Claude Code uses prompt caching](https://code.claude.com/docs/en/prompt-caching) — 자동 관리, TTL 자동 선택, 무효화/유지 행동, 환경변수
- [Prompt caching (API)](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) — cache_control, 가격 배수, 모델별 최소 토큰
- [Manage costs effectively](https://code.claude.com/docs/en/costs) — Claude Code의 자동 비용 최적화
