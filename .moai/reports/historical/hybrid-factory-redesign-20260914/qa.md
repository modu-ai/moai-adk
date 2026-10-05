# 렌더링 검증

Claim: HTML 생성, 로컬 링크 정합성, 두 viewport의 가로 넘침과 목차 anchor를 확인했다. 모바일 첫 화면 스크린샷을 직접 보았다.

Evidence: node render.cjs 출력:

```text
{"bytes":35435,"sections":13,"brokenLocalLinks":[]}
```

agent-browser --session hybrid-redesign-01a09dfb로 파일을 연 뒤 set viewport 및 eval로 측정했다.

```text
{"width":1440,"scrollWidth":1440,"sections":13,"brokenAnchors":0}
{"width":390,"scrollWidth":390}
```

Baseline-attribution: 2026-09-14 이 디렉터리 report.html. [모바일 화면](mobile.png).

Gaps: 전체 스크롤의 시각 검사, 실제 LLM 호출, 제품 테스트는 수행하지 않았다. 최초 렌더링에서 로컬 링크 2개 오류를 발견해 상대 경로를 수정한 뒤 위 검사를 재실행했다.

Residual-risk: 외부 폰트와 선택적 Mermaid 표시는 네트워크에 의존한다. 기본 흐름도와 본문은 파일 안에 포함된다. 화면 확인은 설계 구현이나 실제 절감 효과의 검증이 아니다.
