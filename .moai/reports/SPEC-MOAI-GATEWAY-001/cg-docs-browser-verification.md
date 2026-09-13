# CG 이전 안내 브라우저 확인

## Claim

부모가 생성 사이트의 한국어 페이지를 1280px 데스크톱과 390px 모바일에서 열고, 영어·일본어·중국어 페이지를 390px 모바일에서 직접 열었다. 제목·이전 명령은 렌더되었으며 해당 화면의 document scrollWidth는 viewport와 같았다. 아래 다섯 화면을 이미지 도구로 읽어 본문과 미리보기 명령의 표시를 확인했다. 전체 사이트의 접근성·레이아웃 무결성 판정은 아니다.

## Evidence

`timeout 600 python3 -m http.server 38761 --bind 127.0.0.1 --directory /tmp/cg-docs-render-20260911`로 제한된 로컬 서버를 열었다. sandbox bind 거절 뒤 승인된 재실행을 사용했다. 첫 서버는 브라우저 연결 전에 timeout 124로 끝나 connection refused가 있었으며, 새 서버의 다음 성공 관측과 구분한다.

각 경로에서 `agent-browser --session gateway-cg-docs-20260911 open http://127.0.0.1:38761/<lang>/multi-llm/cg-mode/`, `eval`의 DOM 측정과 `screenshot`을 실행했다. 한국어 초기 snapshot은 출력 한도를 넘겨 일부 잘렸으므로 전체 AX 검증을 주장하지 않는다. 별도 DOM 측정은 제목·코드블록·폭만 반환했다.

실제 서버 출력:

```text
127.0.0.1 - - [11/Sep/2026 17:49:40] "GET /ko/multi-llm/cg-mode/ HTTP/1.1" 200 -
127.0.0.1 - - [11/Sep/2026 17:50:41] "GET /en/multi-llm/cg-mode/ HTTP/1.1" 200 -
127.0.0.1 - - [11/Sep/2026 17:50:43] "GET /ja/multi-llm/cg-mode/ HTTP/1.1" 200 -
127.0.0.1 - - [11/Sep/2026 17:50:45] "GET /zh/multi-llm/cg-mode/ HTTP/1.1" 200 -
```

DOM 출력의 JSON 값(도구의 문자열 포장만 제거):

```json
{"lang":"en","width":390,"scrollWidth":390,"h1":"CG retirement and migration","code":["moai migrate cg\nmoai migrate cg --target claude-only","moai migrate cg --target claude-only --apply --accept-role-change","moai migrate cg --target claude-glm"]}
{"lang":"ja","width":390,"scrollWidth":390,"h1":"CG の廃止と設定の移行","code":["moai migrate cg\nmoai migrate cg --target claude-only","moai migrate cg --target claude-only --apply --accept-role-change","moai migrate cg --target claude-glm"]}
{"lang":"zh","width":390,"scrollWidth":390,"h1":"CG 停用与配置迁移","code":["moai migrate cg\nmoai migrate cg --target claude-only","moai migrate cg --target claude-only --apply --accept-role-change","moai migrate cg --target claude-glm"]}
```

한국어 desktop DOM은 `width:1280, scrollWidth:1280`, mobile DOM은 `width:390, scrollWidth:390`이었다. 모바일 pre 세 개 모두 `client:350, scroll:350, overflow:"auto"`였다. 한국어 페이지에서 `errors`는 exit 0, 빈 출력이었다.

- `/tmp/cg-docs-desktop-20260911.png`
- `/tmp/cg-docs-mobile-20260911.png`
- `/tmp/cg-docs-en-mobile-20260911.png`
- `/tmp/cg-docs-ja-mobile-20260911.png`
- `/tmp/cg-docs-zh-mobile-20260911.png`

검증 뒤 `close`는 `✓ Browser closed`를 반환했다. 서버에 Ctrl-C를 보내 `Keyboard interrupt received, exiting.`과 exit 0을 확인했다.

## Baseline-attribution

2026-09-11 한국 시각 17:49~17:51, WT `moai-proxy-unified`, HEAD `81c1d58f9`, `WT-unified-gateway`. 입력은 `cg-docs-verification.md`의 `/tmp/cg-docs-render-20260911` 생성물이다. browser 관측에서 새 Hugo 빌드를 했다고 주장하지 않는다. 제품 코드·원본 문서는 이 확인에서 수정하지 않았다.

## Gaps

모바일 screenshot의 공통 상단 GitHub 버튼 오른쪽 일부가 화면 끝에서 잘려 보인다. 해당 레이아웃의 이전 기준 화면은 측정하지 않았으므로 이번 문서 변경의 회귀라고 판정하지 않았다. 문서 본문 측정과 별개의 관측이다. 모든 페이지·모든 viewport·독립 원어민·키보드 접근성·외부 URL 가용성 검사는 하지 않았다. 화면의 업데이트 날짜는 2026-08-27로 남아 있다. `hugo.toml`은 `enableGitInfo = true`, single template는 `.Lastmod`를 표시하며 이 작업은 아직 commit하지 않았다. 이 정보만으로 날짜 계산 전체를 검증했다고 주장하지 않는다.

## Residual-risk

현재 snapshot과 DOM 폭은 보이는 본문과 지정 요소의 제한된 관측이다. 숨김 요소·전체 페이지 스크롤·다른 locale의 브라우저 오류까지 없다고 보장하지 않는다. CG 이전 명령의 실행 검증은 별도 `cg-functional-consistency-review.md`의 범위이며, 실제 gateway·TEAMMATE 완료나 사이트 배포를 뜻하지 않는다.
