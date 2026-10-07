# card t1576 — card-review (codex_review scope=card)

- 실행: `mcp__moai__codex_review` scope=card, project_root=카드 트리, base=87da0636746331c1a4b6dc53e243fac9fd75fd82 (origin/main), HEAD=296cba78b, backend=codex
- 판정: **fail — 1건 (advisory=true)** — 본 단계는 자문형으로 병합을 가로지르지 않는다(공장 규율: advisory only, never a replacement for the leader's evidence read).

## 지적과 처분

| # | 등급 | 지적 | 처분 |
|---|---|---|---|
| CR-1 | P2 | `shellSegments`·`shellFields`가 POSIX `\'` 임베드(shellJoinArgs가 생성하는 `'…'\''…'`)와 이중인용 내 `\"` 이스케이프에서 따옴표 상태를 어긋나게 읽어, 인자 데이터 안의 `; go test -json`을 유령 세그먼트로 인식 | **수리(본인 결함)** — 양 스캐너에 이스케이프 가드: 외부의 `\X`는 다음 바이트와 함께 소비(따옴표 상태 불변), 이중인용 내 `\"`는 닫지 않음, 단일인용은 무이스케이프 유지(POSIX). 회귀 테스트 `TestClassifyEscapedQuoteStaysOneCommand`(이중인용 이스케이프 형태 — 구코드 실패 입증형). 판정: 전 분류기 계열+vet+gofmt+lint(b660ffjnq→후속 실행) |

- 커밋 296cba78b 이후 수리이므로 후속 커밋으로 적재(동일 브랜치).
- 이 판정문의 전문은 kgasezam8 태스크 결과로 보존(요약 상단 인용).
