# HTML 보고서 검증 기록

## Claim
최종 HTML을 브라우저로 열어 1440/768/390px, 목차 이동, 증거 펼치기, 온라인 Mermaid, 오프라인 fallback, Letter PDF의 주요 텍스트를 검사했다. 상세 결과는 아래 실행 명령과 출력으로 한정한다.

## Baseline-attribution
/Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/report.html

SHA-256: bc6bb25d646f802d31583c5bcd30081a47420acbb8008bf6950f3f680d0852ff

## Evidence
```text
"agent-browser" "--session" "workflow-audit-20260911" "set" "offline" "off"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "network" "unroute"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "open" "file:///Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/report.html"
exit=0
stdout:
✓ MoAI 워크플로우 심층 감사 · 2026-09-11
  file:///Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/report.html

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "wait" "--fn" "!!document.querySelector(\".mermaid svg\")"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "eval" "document.fonts.ready.then(()=>document.fonts.status)"
exit=0
stdout:
"loaded"

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "set" "viewport" "1440" "1000"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "eval" "window.scrollTo(0,0)"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "eval" "JSON.stringify({width:innerWidth,scrollWidth:document.documentElement.scrollWidth,findings:document.querySelectorAll(\".finding\").length,ruleRows:document.querySelectorAll(\"#rules tbody tr\").length,agents:document.querySelectorAll(\"#agents tbody tr\").length,duplicateIDs:[...document.querySelectorAll(\"[id]\")].map(e=>e.id).filter((x,i,a)=>a.indexOf(x)!==i),brokenAnchors:[...document.querySelectorAll(\"a[href^=\\\"#\\\"]\")].filter(a=>!document.getElementById(a.hash.slice(1))).map(a=>a.hash),mermaidSVG:!!document.querySelector(\".mermaid svg\")})"
exit=0
stdout:
"{\"width\":1440,\"scrollWidth\":1440,\"findings\":40,\"ruleRows\":84,\"agents\":11,\"duplicateIDs\":[],\"brokenAnchors\":[],\"mermaidSVG\":true}"

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "screenshot" "/Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/desktop-1440.png"
exit=0
stdout:
✓ Screenshot saved to /Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/desktop-1440.png

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "set" "viewport" "768" "1000"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "eval" "window.scrollTo(0,0)"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "eval" "JSON.stringify({width:innerWidth,scrollWidth:document.documentElement.scrollWidth,findings:document.querySelectorAll(\".finding\").length,ruleRows:document.querySelectorAll(\"#rules tbody tr\").length,agents:document.querySelectorAll(\"#agents tbody tr\").length,duplicateIDs:[...document.querySelectorAll(\"[id]\")].map(e=>e.id).filter((x,i,a)=>a.indexOf(x)!==i),brokenAnchors:[...document.querySelectorAll(\"a[href^=\\\"#\\\"]\")].filter(a=>!document.getElementById(a.hash.slice(1))).map(a=>a.hash),mermaidSVG:!!document.querySelector(\".mermaid svg\")})"
exit=0
stdout:
"{\"width\":768,\"scrollWidth\":768,\"findings\":40,\"ruleRows\":84,\"agents\":11,\"duplicateIDs\":[],\"brokenAnchors\":[],\"mermaidSVG\":true}"

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "screenshot" "/Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/tablet-768.png"
exit=0
stdout:
✓ Screenshot saved to /Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/tablet-768.png

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "set" "viewport" "390" "844"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "eval" "window.scrollTo(0,0)"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "eval" "JSON.stringify({width:innerWidth,scrollWidth:document.documentElement.scrollWidth,findings:document.querySelectorAll(\".finding\").length,ruleRows:document.querySelectorAll(\"#rules tbody tr\").length,agents:document.querySelectorAll(\"#agents tbody tr\").length,duplicateIDs:[...document.querySelectorAll(\"[id]\")].map(e=>e.id).filter((x,i,a)=>a.indexOf(x)!==i),brokenAnchors:[...document.querySelectorAll(\"a[href^=\\\"#\\\"]\")].filter(a=>!document.getElementById(a.hash.slice(1))).map(a=>a.hash),mermaidSVG:!!document.querySelector(\".mermaid svg\")})"
exit=0
stdout:
"{\"width\":390,\"scrollWidth\":390,\"findings\":40,\"ruleRows\":84,\"agents\":11,\"duplicateIDs\":[],\"brokenAnchors\":[],\"mermaidSVG\":true}"

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "screenshot" "/Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/mobile-390.png"
exit=0
stdout:
✓ Screenshot saved to /Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/mobile-390.png

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "click" "nav a[href=\"#findings\"]"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "get" "url"
exit=0
stdout:
file:///Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/report.html#findings

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "click" "#F02 summary"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "eval" "JSON.stringify({opened:document.querySelector(\"#F02 details\").open,sourceEvidenceVisible:document.querySelector(\"#F02 pre\").getBoundingClientRect().height>0,width:innerWidth,scrollWidth:document.documentElement.scrollWidth})"
exit=0
stdout:
"{\"opened\":true,\"sourceEvidenceVisible\":true,\"width\":390,\"scrollWidth\":390}"

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "screenshot" "/Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/mobile-evidence.png"
exit=0
stdout:
✓ Screenshot saved to /Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/mobile-evidence.png

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "errors"
exit=0
stdout:

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "set" "offline" "on"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "network" "route" "**cdn.jsdelivr.net/**" "--abort"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "reload"
exit=0
stdout:
file:///Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/report.html#findings

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "wait" "--fn" "document.querySelector(\".mermaid\").getAttribute(\"data-fallback\")===\"true\""
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "eval" "JSON.stringify({offline:!navigator.onLine,mermaidSVG:!!document.querySelector(\".mermaid svg\"),fallback:document.querySelector(\".mermaid\").getAttribute(\"data-fallback\"),fallbackSteps:document.querySelectorAll(\".flow-fallback div\").length,findings:document.querySelectorAll(\".finding\").length,width:innerWidth,scrollWidth:document.documentElement.scrollWidth})"
exit=0
stdout:
"{\"offline\":true,\"mermaidSVG\":false,\"fallback\":\"true\",\"fallbackSteps\":4,\"findings\":40,\"width\":390,\"scrollWidth\":390}"

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "click" "nav a[href=\"#flow\"]"
exit=0
stdout:
✓ Done

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "screenshot" "/Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/offline-flow.png"
exit=0
stdout:
✓ Screenshot saved to /Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/offline-flow.png

stderr:

```

```text
"agent-browser" "--session" "workflow-audit-20260911" "pdf" "/Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/print-check.pdf"
exit=0
stdout:
✓ PDF saved to /Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/print-check.pdf

stderr:

```

```text
"pdfinfo" "/Users/goos/MoAI/moai-adk-go/reports/workflow-performance-audit-20260911/print-check.pdf"
exit=0
stdout:
Title:           MoAI 워크플로우 심층 감사 · 2026-09-11
Creator:         Chromium
Producer:        Skia/PDF m145
CreationDate:    Fri Sep 11 12:45:09 2026 KST
ModDate:         Fri Sep 11 12:45:09 2026 KST
Custom Metadata: no
Metadata Stream: no
Tagged:          no
UserProperties:  no
Suspects:        no
Form:            none
JavaScript:      no
Pages:           54
Encrypted:       no
Page size:       612 x 792 pts (letter)
Page rot:        0
File size:       1417180 bytes
Optimized:       no
PDF version:     1.4

stderr:

```

```text
pdftotext print-check.pdf -; programmatic content assertions
exit=0
stdout:
{"sourceQuote":true,"lastFinding":true,"residualRisk":true,"A4":false}
stderr:

```

```text
Node: report bytes, local href existence, noscript content, SHA-256
exit=0
stdout:
{"html_bytes":118192,"under_120000":true,"missingLocalLinks":[],"nonemptyNoscript":true,"sha256":"bc6bb25d646f802d31583c5bcd30081a47420acbb8008bf6950f3f680d0852ff"}
stderr:

```

## Gaps
전체 PDF 페이지의 시각 검수, 모든 브라우저/스크린리더, JavaScript 자체를 비활성화한 브라우저는 검사하지 않았다. noscript는 정적 존재 검사이며 offline은 실제 네트워크 차단 검사다. 표는 모바일에서 컨테이너 안의 가로 스크롤을 허용한다.

## Residual-risk
폰트와 Mermaid CDN은 외부 자원이며 offline에서는 시스템 폰트와 텍스트/단계 fallback을 쓴다. PDF는 인쇄 확인용 산출물이며 접근성 태깅이 된 배포용 PDF는 아니다. 이전 Letter 시안의 rasterizer는 Type 3 glyph bounding-box 경고를 냈고 첫 페이지 육안 검수에는 누락이 보이지 않았다. 최종 Letter의 전체 글리프 정확성을 보증하지 않는다.

PDF CLI가 CSS의 A4 지정을 따르지 않아 실제 출력은 Letter로 관측됐다. A4 크기 출력은 미검증이다.
