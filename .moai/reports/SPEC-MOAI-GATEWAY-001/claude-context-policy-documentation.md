# Claude 5 공식 문맥·thinking 정책 근거

## Claim

2026-09-11 web 도구로 공식 모델 문서를 직접 열었다. Opus5와 Sonnet5의 문맥·출력 수치 및 native thinking 문법을 catalog/profile 근거로 사용할 수 있으나, 현재 OAuth 계정의 대형 입력 수용을 측정한 것은 아니다.

## Evidence

web open:
- https://platform.claude.com/docs/en/models/opus-5/whats-new-opus-5
- https://platform.claude.com/docs/en/models/sonnet-5/whats-new-sonnet-5

Opus 문서60행은 context 기본/최대1M, 출력128k로 설명한다. 81–84행은 adaptive 기본값, 기본 display에서 빈 thinking 문자열, 도구 결과와 함께 완전한 thinking block 보존, max_tokens가 thinking+text를 포함하는 hard 출력 제한임을 명시한다. 112–115행에서 disabled는 effort high 이하에 허용하고 xhigh/max와의 조합은 오류다. [Opus5 공식 문서](https://platform.claude.com/docs/en/models/opus-5/whats-new-opus-5)

Sonnet 문서는 기본/최대1M context와128k 출력, 기본 adaptive, manual extended thinking 및 비기본 sampling parameter의400을 설명한다. 새 tokenizer 때문에 이전 모델의 계량을 동일하게 볼 수 없다. [Sonnet5 공식 문서](https://platform.claude.com/docs/en/models/sonnet-5/whats-new-sonnet-5)

## Baseline-attribution

이 실행에서 공식 웹페이지를 열어 읽은 문서 선언이다. 제품 WT81c1d58f9의 코드나 구독 계정 runtime 수치가 아니다. 정책 구현 담당에게 native disabled/high와 GPT의 검증되지 않은 비추론 대응을 혼동하지 않도록 인계했다.

## Gaps

대형 입력·최대 출력 실측, 계정별 정책, native의 모든 beta/event 변형, GLM 문맥은 검증하지 않았다. 현재 실제M0의 작은Opus/Sonnet 요청200은 별도 제한된 근거다.

## Residual-risk

문서의1M/128k를 실제계정에서측정한수치로표시하지않는다. GPT서버출력정책승인을Anthropic의max_tokens생략으로확대하지않는다. 기존최소subset밖의샘플링/effort문법은지원표와시험근거없이일괄허용하지않는다.
