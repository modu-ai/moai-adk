# HTML 보고서 확인 기록

## Claim

생성된 report.html을 실제 브라우저에서 데스크톱·모바일 폭으로 열었다. 두 폭에서 문서 전체 가로 넘침이 없었으며, 첫 화면 스크린샷을 직접 확인했다. 제품의 실계정 실행 검증은 아니다.

## Evidence

브라우저 세션: factory-research-01a09dfb. 실행: agent-browser set viewport, eval, screenshot. 각 명령은 해당 세션을 명시했다.

데스크톱 1440 × 1100 DOM 측정 출력:

```text
{"title":"Claude Code 다중 LLM Factory — 구독·MCP·Gateway 최종 비교","width":1440,"scrollWidth":1440,"sections":17,"tables":11,"bodyText":20338}
```

모바일 390 × 844에서 다음 식을 eval로 실행했다.

```js
JSON.stringify({width:innerWidth,scrollWidth:document.documentElement.scrollWidth,overflowing:[...document.querySelectorAll("section,header,main,nav")].filter(e=>e.getBoundingClientRect().right>innerWidth+1).map(e=>e.tagName)})
```

출력:

```text
{"width":390,"scrollWidth":390,"overflowing":[]}
```

정적 파일 검사 출력:

```text
{"htmlBytes":53536,"sections":17,"externalSourceLinks":27,"brokenLocalLinks":[],"brokenAnchors":[]}
```

화면 기록: [데스크톱](desktop.png), [모바일](mobile.png).

## Baseline-attribution

2026-09-14, 이 디렉터리의 report.html. 세션 01a09dfb-5908-7e50-ac3a-bde14d9af79f. 제품 소스·테스트 기준은 report.md의 14·15절에 별도 기록했다.

## Gaps

전체 본문의 모든 스크롤 위치를 스크린샷으로 검사하지 않았다. 모든 외부 링크의 장기 가용성, 보조 기술 접근성, 모든 브라우저 조합은 검증하지 않았다. 초기 샌드박스 실행은 브라우저 소켓 디렉터리 권한으로 실패했고, 승인된 권한으로 다시 실행했다.

## Residual-risk

외부 폰트·선택적 다이어그램 라이브러리는 네트워크 상태에 영향을 받는다. 본문과 기본 구조도는 HTML에 포함된다. 문서 화면 검증을 실제 다중 LLM Factory 운영 성공으로 해석할 수 없다.
