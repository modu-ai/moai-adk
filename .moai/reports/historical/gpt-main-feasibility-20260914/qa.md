# 보고서 파일 확인

Claim: HTML을 생성하고 브라우저로 열었다. 첫 화면을 모바일에서 직접 확인했다.

Evidence: node render.cjs의 최종 출력:

```text
{"bytes":32423,"sections":11,"brokenLocalLinks":[]}
```

agent-browser --session gpt-main-report-01a09dfb의 set viewport와 eval 결과:

```text
{"width":1440,"scrollWidth":1440,"title":"moai gpt — GPT 메인 연결의 허용성과 성능","sections":11,"brokenAnchors":1}
```

발견한 anchor 오류를 수정한 뒤 파일을 다시 열고 확인:

```text
{"width":390,"scrollWidth":390,"brokenAnchors":0}
```

Baseline-attribution: 이 디렉터리 report.html, 2026-09-14. [모바일 화면](mobile.png). 초기 렌더러의 Buffer context 오류도 수정 후 재실행했다.

Gaps: 전체 스크롤 위치의 시각 검증과 모든 외부 링크 가용성은 확인하지 않았다. 제품 검증 범위는 report.md에 기록했다.

Residual-risk: 웹폰트·선택적 Mermaid는 네트워크에 의존한다. 기본 구조와 본문은 HTML 내부에 포함된다. 문서 검증은 실제 GPT main 운영 검증이 아니다.
