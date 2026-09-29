# HTML 브라우저 검사

## Claim

로컬 HTML을 Chromium 기반 agent-browser에서 열고 21개 section과 내부 앵커, 화면 폭을 검사했다. 제품 기능을 검사한 것은 아니다.

## Evidence

실행: agent-browser --session moai-raine-report --allow-file-access open file:///Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop/reports/raine-proxy-gateway-design-20260914-01a09dfb/report.html

출력: ✓ MoAI Gateway — 다중 모델 통합 설계 및 ToS 검토

실행: agent-browser --session moai-raine-report eval 'JSON.stringify({title:document.title,sections:document.querySelectorAll("main section").length,width:innerWidth,scroll:document.documentElement.scrollWidth,links:document.querySelectorAll("a").length,missing:[...document.querySelectorAll("a[href^=\"#\"]")].filter(a=>!document.getElementById(a.hash.slice(1))).length})'

```json
{"title":"MoAI Gateway — 다중 모델 통합 설계 및 ToS 검토","sections":21,"width":1280,"scroll":1280,"links":79,"missing":0}
```

실행: agent-browser --session moai-raine-report set viewport 390 844

출력: ✓ Done

실행: agent-browser --session moai-raine-report eval 'JSON.stringify({width:innerWidth,scroll:document.documentElement.scrollWidth,sections:document.querySelectorAll("main section").length})'

```json
{"width":390,"scroll":390,"sections":21}
```

## Baseline-attribution

최초 렌더 59,349 bytes 기준. 이후 호환성 표·설정 흐름·본 검사 링크를 추가했으므로 최종 렌더 재검사를 별도로 기록한다. [데스크톱](desktop.png), [모바일](mobile.png).

최종 렌더 명령: node /Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop/reports/raine-proxy-gateway-design-20260914-01a09dfb/render.cjs

```json
{"bytes":62272,"sections":21,"missing_local_links":[],"broken_anchors":[]}
```

최종 HTML을 다시 열어 앞의 모바일 DOM 검사와 내부 앵커 검사를 실행했다.

```json
{"sections":21,"width":390,"scroll":390,"missing":0}
```

## Gaps

전체 접근성 감사, 인쇄 결과, 오프라인 폰트, 모든 외부 링크의 응답, 실계정 모델 생성은 검사하지 않았다.

## Residual-risk

브라우저·화면 크기·폰트 상태에 따라 줄바꿈은 달라질 수 있다. 보고서 렌더링 성공은 설계의 런타임 구현 성공을 뜻하지 않는다.
