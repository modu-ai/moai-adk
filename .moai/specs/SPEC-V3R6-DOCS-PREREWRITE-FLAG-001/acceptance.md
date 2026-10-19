# Acceptance Criteria — SPEC-V3R6-DOCS-PREREWRITE-FLAG-001

## A. Pin and conventions

- RED-now 계측 핀: 트리 `74ad197380` (2026-10-19, 이 실행). HEAD가 움직이면 런 페이즈 Pre-flight에서 재측정·재핀(verification-completeness §4).
- 두 칸 채택 규율: 릴리스 블로킹 AC는 RED-now 셀 + green path 셀을 쌍으로 가진다. RED-now 셀은 네 요소(단일 호출 명령, 원문 stdout, 별도 exit 코드 필드, 트리 SHA)를 증거 장부 항목으로 가진다(§2.1; 운반체는 이 장부의 fenced 항목).
- 회귀 가드 분류: 구현 전 트리에서 빨간색이 될 수 없는(공집합·보존·음(-) 성격의) 기준은 regression-guard로 분류하고 릴리스 블로킹에서 제외한다 — RED-now를 강하게 써 둔다고 관문이 되지 않는다는 규칙의 처분(§2.1 undecidable disposition)을 따른다.
- 인용 원문은 원시 파일 바이트 기준이다(렌더링 뷰 아님).

### A.1 Traceability matrix (AC → REQ → Milestone)

| AC | REQ | Milestone | 분류 |
|---|---|---|---|
| AC-DPF-001 | REQ-DPF-001, REQ-DPF-003 | M2 → M3 | release-blocking (RED-now E-01) |
| AC-DPF-002 | REQ-DPF-002, REQ-DPF-003 | M2 → M3 | release-blocking (RED-now E-02) |
| AC-DPF-003 | REQ-DPF-005 | M3 | regression-guard |
| AC-DPF-004 | REQ-DPF-005 | M3 | regression-guard |
| AC-DPF-005 | REQ-DPF-004 | M2·M3 | regression-guard (Q17) |
| AC-DPF-006 | REQ-DPF-006 | M3 | regression-guard (음 검증) |

## B. Evidence ledger (원문)

### E-01 — README relaunch 루프 인용 (RED-now; AC-DPF-001)

```
command: grep -c -- "--clear-policy relaunch" README.md README.ko.md README.ja.md README.zh.md
stdout:
README.md:1
README.ko.md:1
README.ja.md:1
README.zh.md:1
exit: 0
tree: 74ad197380
```

### E-02 — factory-mode relaunch 셀의 대체 소비 서술 (RED-now; AC-DPF-002)

```
command: grep -nE "다음 카드용 새 세션|fresh session for the next card|次のカード用に新しいセッション|为下一张卡片拉起全新会话" docs-site/content/ko/advanced/factory-mode.md docs-site/content/en/advanced/factory-mode.md docs-site/content/ja/advanced/factory-mode.md docs-site/content/zh/advanced/factory-mode.md
stdout:
docs-site/content/ja/advanced/factory-mode.md:77:| `--clear-policy <値>` | カードを終えたあと、レーンがコンテキストをどう空けるかを決めます。`clear-each` (既定値。カードごとに `/clear` を求める)、`clear-when-full` (コンテキスト使用量がモデルごとの引き継ぎ閾値に届いたときだけ求める)、`relaunch` (セッションの終了を求め、監督ランチャーが次のカード用に新しいセッションを起動する) |
docs-site/content/en/advanced/factory-mode.md:77:| `--clear-policy <value>` | Decides how the lane clears its context after finishing a card. `clear-each` (the default, asks for `/clear` after every card), `clear-when-full` (asks only when context usage reaches the model-specific handoff threshold), `relaunch` (asks you to end the session, and the supervising launcher starts a fresh session for the next card) |
docs-site/content/ko/advanced/factory-mode.md:77:| `--clear-policy <값>` | 카드를 끝낸 뒤 레인이 컨텍스트를 어떻게 비울지 정합니다. `clear-each`(기본값, 카드마다 `/clear`를 요청), `clear-when-full`(컨텍스트 사용량이 모델별 인수인계 임계치에 닿을 때만 요청), `relaunch`(세션 종료를 요청하고 감독 런처가 다음 카드용 새 세션을 띄움) |
docs-site/content/zh/advanced/factory-mode.md:77:| `--clear-policy <值>` | 决定 lane 做完一张卡片后如何清理上下文。`clear-each`（默认，每张卡片后请求 `/clear`）、`clear-when-full`（上下文用量达到按模型而定的交接阈值时才请求）、`relaunch`（请求结束会话，由监督启动器为下一张卡片拉起全新会话） |
exit: 0
tree: 74ad197380
```

### E-03 — `--factory-run` 보존 baseline (verification; RED 셀 아님)

```
command: grep -c -- "--factory-run" docs-site/content/ko/advanced/factory-mode.md docs-site/content/en/advanced/factory-mode.md docs-site/content/ja/advanced/factory-mode.md docs-site/content/zh/advanced/factory-mode.md docs-site/content/ko/cli-reference/launchers.md docs-site/content/en/cli-reference/launchers.md docs-site/content/ja/cli-reference/launchers.md docs-site/content/zh/cli-reference/launchers.md README.md README.ko.md README.ja.md README.zh.md
stdout (파일별 라인 카운트): factory-mode.md ×4 = 각 2, launchers.md ×4 = 각 1, README ×4 = 각 2
exit: 0
tree: 74ad197380
```

### E-04 — `--clear-policy` 보존 baseline (verification; RED 셀 아님)

```
command: grep -c -- "--clear-policy" docs-site/content/ko/advanced/factory-mode.md docs-site/content/en/advanced/factory-mode.md docs-site/content/ja/advanced/factory-mode.md docs-site/content/zh/advanced/factory-mode.md docs-site/content/ko/cli-reference/launchers.md docs-site/content/en/cli-reference/launchers.md docs-site/content/ja/cli-reference/launchers.md docs-site/content/zh/cli-reference/launchers.md README.md README.ko.md README.ja.md README.zh.md
stdout (파일별 라인 카운트): factory-mode.md ×4 = 각 2 (77·134행), launchers.md ×4 = 각 1 (37행), README ×4 = 각 2 (74·87행)
exit: 0
tree: 74ad197380
```

## C. Acceptance criteria

### AC-DPF-001 — README 4종의 relaunch 루프 인용이 정리됐다 (REQ-DPF-001·003) — release-blocking

- **RED-now** (E-01, 트리 74ad197380): 네 README 각각 1줄이 "카드를 순회하는 `--clear-policy relaunch` 루프"를 현재 동작으로 서술한다(카드 t1554로 대체된 소비). 빨간 이유: 대체된 소비 경로의 인용이 살아 있는 상태 — 구현 전 트리의 실측.
- **Green path** (M2가 뒤집는다): `grep -c -- "--clear-policy relaunch" README.md README.ko.md README.ja.md README.zh.md` → 각 파일 `0`, exit 1 (짝 없음). 치환 문장은 Q13에 따라 현재 메커니즘(단일 레인 세션이 `moai todo --auto` 큐를 소비)을 이름 붙인다(문장 대조는 M3).
- 뮤턴트 검사: 인용 줄째로 삭제만 해도 이 AC는 통과하지만 Q13 위반이다 — 처분 충실은 green path 문장 요건과 M3 대조가 담당한다.

### AC-DPF-002 — factory-mode 4 로캘의 relaunch 셀 서술이 정리됐다 (REQ-DPF-002·003) — release-blocking

- **RED-now** (E-02, 트리 74ad197380): 네 로캘 77행 `relaunch` 셀이 "감독 런처가 카드마다 다음 카드용 새 세션을 띄운다"를 현재 의미로 서술한다(대체된 카드별 순환). 빨간 이유: 동일 — 구현 전 실측.
- **Green path** (M2): E-02의 명령 → 출력 없음, exit 1. `--clear-policy <값>` 행과 clear-each·clear-when-full 서술은 유지되고, `relaunch` 셀은 현재 소비 서술로 치환된다(Q13).
- 로캘별 문구가 달라도 검사는 의미 조각(각 로캘의 "카드마다 새 세션" 절)을 때린다 — E-02의 교차 로캘 정규식이 그 역할을 한다.

### AC-DPF-003 — `--factory-run` 인용 보존 (REQ-DPF-005) — regression-guard

- 현재 동작(선택자·합류·이어받기)을 서술하는 인용이므로 구현 전 트리에서 빨간색이 될 수 없어 회귀 가드로 분류한다(§2.1 vacuous 방향 — 단, green path와 짝을 이뤄 "전부 삭제" 뮤턴트를 막는다).
- **Green path** (M3): E-03 명령 재실행 → 파일별 카운트가 baseline과 동일(factory-mode 각 2, launchers 각 1, README 각 2), exit 0.
- 릴리스 블로킹 아님 — 이 기준 위반은 AC-DPF-001·002의 green 상태에서도 남는 과잉 삭제를 잡는 가드다.

### AC-DPF-004 — `--clear-policy` 선택 인용 보존 (REQ-DPF-005) — regression-guard

- **Green path** (M3): E-04 명령 재실행 → factory-mode 각 2 이상(77·134행 유지), launchers 각 1(37행 유지), README 각 1 이상(87행 유지; 74행 치환으로 카운트가 2→1이 될 수 있다). exit 0.
- launchers 37행의 값 나열과 factory-mode 134행·README 87행의 정책 변경 서술은 LIVE이므로 무변경이 원칙이다.

### AC-DPF-005 — 4 로캘 패리티, 단일 변경 세트 (REQ-DPF-004, Q17) — regression-guard

- 구현 전 트리에서 이미 대칭이라 RED-now가 불가능하다 — Q17 운영자 결정(d-20261019T015908Z-220a)의 절차 위반(분산 편집)을 잡는 절차 가드로 분류한다.
- **Green path** (M2·M3): `grep -c -- "--clear-policy"`와 `grep -c -- "--factory-run"`을 문서군별로 비교 → factory-mode ×4 동일, launchers ×4 동일, README ×4 동일. 런 페이즈 결과물은 하나의 커밋(또는 같은 변경 세트의 커밋들)에 담긴다.

### AC-DPF-006 — internal/·cmd/ 무변경 (REQ-DPF-006) — regression-guard

- 음(-) 검증: 변경이 일어나기 전에는 공집합이라 RED-now가 불가능하다 — help-text 추가(Out of Scope)가 코드로 새는 것을 막는 범위 가드다.
- **Green path** (M3): `git diff --name-only 74ad197380..HEAD -- internal/ cmd/` → 출력 없음, exit 0. (플랜 페이즈 아티팩트 커밋이 범위에 끼더라도 경로스펙이 internal/·cmd/만 검사하므로 판정은 유효하다.)

## D. Quality gates / Definition of Done

- AC-DPF-001·002 PASS(각각 RED-now → green 플립 관측됨) + AC-DPF-003~006 PASS.
- 모든 AC 보고는 명령 + 원문 출력 + 트리 귀속을 동반한다(verification-claim-integrity §1·§2; manager-develop §E 형식).
- 뮤턴트 점검 기록: "전부 삭제" 뮤턴트는 AC-DPF-003·004가, "한 로캘만 편집" 뮤턴트는 AC-DPF-005가, "help 추가 우회" 뮤턴트는 AC-DPF-006이 막는다. 셋 중 하나라도 뮤턴트를 통과시키는 구현은 이 SPEC의 DoD가 아니다.
- 선택 항목: Hugo 툴체인이 있으면 경고 없는 문서 빌드 확인 — 없으면 보고서 Gaps에 명시(침묵 생략 금지).
