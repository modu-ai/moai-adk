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
| AC-DPF-007 | REQ-DPF-003 | M2 | release-blocking (RED-now E-05) |

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
docs-site/content/ko/advanced/factory-mode.md:77:| `--clear-policy <값>` | 카드를 끝낸 뒤 레인이 컨텍스트를 어떻게 비울지 정합니다. `clear-each`(기본값, 카드마다 `/clear`를 요청), `clear-when-full`(컨텍스트 사용량이 모델별 인수인계 임계치에 닿을 때만 요청), `relaunch`(세션 종료를 요청하고 감독 런처가 다음 카드용 새 세션을 띄움) |
docs-site/content/zh/advanced/factory-mode.md:77:| `--clear-policy <值>` | 决定 lane 做完一张卡片后如何清理上下文。`clear-each`（默认，每张卡片后请求 `/clear`）、`clear-when-full`（上下文用量达到按模型而定的交接阈值时才请求）、`relaunch`（请求结束会话，由监督启动器为下一张卡片拉起全新会话） |
docs-site/content/en/advanced/factory-mode.md:77:| `--clear-policy <value>` | Decides how the lane clears its context after finishing a card. `clear-each` (the default, asks for `/clear` after every card), `clear-when-full` (asks only when context usage reaches the model-specific handoff threshold), `relaunch` (asks you to end the session, and the supervising launcher starts a fresh session for the next card) |
docs-site/content/ja/advanced/factory-mode.md:77:| `--clear-policy <値>` | カードを終えたあと、レーンがコンテキストをどう空けるかを決めます。`clear-each` (既定値。カードごとに `/clear` を求める)、`clear-when-full` (コンテキスト使用量がモデルごとの引き継ぎ閾値に届いたときだけ求める)、`relaunch` (セッションの終了を求め、監督ランチャーが次のカード用に新しいセッションを起動する) |
exit: 0
tree: 13cfafa454c50effb777b693483b30287e2c7b83
```

재기록 사유(iter-1 D4): 초판 장부가 임의 순서(ja·en·ko·zh)를 기록했다. 이 트리에서 같은 명령을 재실행해 실측 순서(ko·zh·en·ja)를 원문 그대로 기록했다. 실행 바이너리: 이 셸의 `grep`은 ugrep 7.8.4 래퍼(zsh 함수, `ARGV0=ugrep`)로 다중 파일 출력 순서가 인수 순서가 아니며 실행마다 달라진다(연속 2회 실측: ko·zh·en·ja → en·ja·ko·zh). 네 줄의 내용은 실행과 트리에 걸쳐 바이트 동일하다 — 74ad197380 핀 대비 docs 바이트 무변경을 `git diff --name-only 74ad197380..HEAD`가 5개 SPEC 아티팩트 파일만 가리키는 것으로 확인했고, iter-1 감사(MP-8)도 74ad197380 재실행에서 네 줄이 내용 동일함을 확인했다. AC-DPF-002의 판정(출력 없음, exit 1)은 순서와 무관하다.

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

### E-05 — STALE-3 레인 `/clear` 안내 12줄의 현재 존재 (RED-now; AC-DPF-007, iter-2 D7 신설)

```
command: grep -nE '카드마다 `/clear`를 요청|asks for `/clear` after every card|カードごとに `/clear` を求める|每张卡片后请求 `/clear`|그 레인을 `/clear`합니다|the lane is cleared before it takes the next one|そのレーンを `/clear` します|先对该 lane 执行 `/clear`|레인이 `/clear`를 요청한다|the lane asks for a `/clear`|レーンが `/clear` を依頼する|泳道会请你 `/clear`' docs-site/content/ko/advanced/factory-mode.md docs-site/content/en/advanced/factory-mode.md docs-site/content/ja/advanced/factory-mode.md docs-site/content/zh/advanced/factory-mode.md README.md README.ko.md README.ja.md README.zh.md
stdout:
docs-site/content/zh/advanced/factory-mode.md:77:| `--clear-policy <值>` | 决定 lane 做完一张卡片后如何清理上下文。`clear-each`（默认，每张卡片后请求 `/clear`）、`clear-when-full`（上下文用量达到按模型而定的交接阈值时才请求）、`relaunch`（请求结束会话，由监督启动器为下一张卡片拉起全新会话） |
docs-site/content/zh/advanced/factory-mode.md:134:- **`/clear` 的边界在卡片与卡片之间。** 卡片到达 `done` 后，先对该 lane 执行 `/clear`，再接下一张卡片（可用 `--clear-policy` 调整）。
docs-site/content/ko/advanced/factory-mode.md:77:| `--clear-policy <값>` | 카드를 끝낸 뒤 레인이 컨텍스트를 어떻게 비울지 정합니다. `clear-each`(기본값, 카드마다 `/clear`를 요청), `clear-when-full`(컨텍스트 사용량이 모델별 인수인계 임계치에 닿을 때만 요청), `relaunch`(세션 종료를 요청하고 감독 런처가 다음 카드용 새 세션을 띄움) |
docs-site/content/ko/advanced/factory-mode.md:134:- **`/clear` 경계는 카드 사이입니다.** 카드가 `done`에 닿으면 다음 카드를 받기 전에 그 레인을 `/clear`합니다(`--clear-policy`로 바꿀 수 있습니다).
docs-site/content/ja/advanced/factory-mode.md:77:| `--clear-policy <値>` | カードを終えたあと、レーンがコンテキストをどう空けるかを決めます。`clear-each` (既定値。カードごとに `/clear` を求める)、`clear-when-full` (コンテキスト使用量がモデルごとの引き継ぎ閾値に届いたときだけ求める)、`relaunch` (セッションの終了を求め、監督ランチャーが次のカード用に新しいセッションを起動する) |
docs-site/content/ja/advanced/factory-mode.md:134:- **`/clear` の境目はカードとカードの間です。** カードが `done` に届いたら、次のカードを受ける前にそのレーンを `/clear` します (`--clear-policy` で変えられます)。
docs-site/content/en/advanced/factory-mode.md:77:| `--clear-policy <value>` | Decides how the lane clears its context after finishing a card. `clear-each` (the default, asks for `/clear` after every card), `clear-when-full` (asks only when context usage reaches the model-specific handoff threshold), `relaunch` (asks you to end the session, and the supervising launcher starts a fresh session for the next card) |
docs-site/content/en/advanced/factory-mode.md:134:- **The `/clear` boundary falls between cards.** Once a card reaches `done`, the lane is cleared before it takes the next one (changeable with `--clear-policy`).
README.zh.md:87:有两条规则让工厂保持诚实。主导**只凭自己从卡片 `progress.md` 里读到的证据**推进卡片 —— 不凭泳道的回复，因为回复是主张而不是观测，而且跨会话投递并不保证送达。另外，一张卡片做完后，泳道会请你 `/clear`。`/clear` 是用户亲手敲的命令，无法当作指令发送（可用 `--clear-policy` 改变这一行为）。
README.md:87:Two rules keep the factory honest. The leader advances a card **only on evidence it read** from the card's `progress.md` — never on a lane's reply, because a reply is a claim and inter-session delivery is not guaranteed. And when a card finishes, the lane asks for a `/clear`, since `/clear` is user-typed and cannot be sent as an instruction (`--clear-policy` changes this).
README.ko.md:87:팩토리를 정직하게 유지하는 규칙이 둘 있다. 리더는 카드의 `progress.md`에서 **직접 읽은 증거로만** 카드를 넘긴다 — 레인의 답장으로는 넘기지 않는다. 답장은 관측이 아니라 주장이고, 세션 간 전달은 보장되지도 않기 때문이다. 그리고 카드가 끝나면 레인이 `/clear`를 요청한다. `/clear`는 사람이 직접 치는 명령이라 지시로 보낼 수 없다(`--clear-policy`로 바꿀 수 있다).
README.ja.md:87:ファクトリーを正直に保つルールが 2 つある。リーダーはカードの `progress.md` を**自分で読んだ証拠だけで**カードを進める。レーンの返信では進めない。返信は観測ではなく主張であり、セッション間の配信は保証されないからだ。そしてカードが終わると、レーンが `/clear` を依頼する。`/clear` は人が直接打つコマンドなので、指示としては送れない（`--clear-policy` で変えられる）。
exit: 0
tree: b484ecacce
```

측정 비고(iter-2 D7): 이 셀은 이 실행에서 현재 HEAD b484ecacce에서 직접 측정했다. `git diff --name-only 74ad197380..HEAD`가 SPEC 아티팩트 5개 파일만 가리키므로 문서 바이트는 무변경이고 74ad197380 핀에서도 같은 12줄이다. 빨간 이유: 12줄(로캘별 factory-mode 77행 `clear-each` 셀·134행, README 87행)이 대체된 레인 `/clear` 안내(STALE-3, iter-1 D2 재분류)를 현재 의미로 아직 서술하기 때문 — 구현 전 실측이다. 실행 바이너리는 E-02와 같은 ugrep 7.8.4 래퍼라 다중 파일 출력 순서가 실행마다 다르다(이 실행: zh·ko·ja·en ×(77·134) 후 README zh·en·ko·ja; 직전 예비 실행은 공통 줄의 내용이 바이트 동일한 다른 순열) — AC-DPF-007의 판정(green = 출력 없음, exit 1)은 순서와 무관하다. 이 정규식은 범위 밖 줄을 치지 않는다: factory-mode 100행의 mermaid 에지 라벨(범위표 밖)과 README 42·164·178행 등의 일반 세션 연속성 서술은 어느 교차 문구와도 적중하지 않는다 — 이 실행의 출력이 정확히 범위표의 STALE-3 12줄이라는 것이 그 증거다.

## C. Acceptance criteria

### AC-DPF-001 — README 4종의 relaunch 루프 인용이 정리됐다 (REQ-DPF-001·003) — release-blocking

- **RED-now** (E-01, 트리 74ad197380): 네 README 각각 1줄이 "카드를 순회하는 `--clear-policy relaunch` 루프"를 현재 동작으로 서술한다(카드 t1554로 대체된 소비). 빨간 이유: 대체된 소비 경로의 인용이 살아 있는 상태 — 구현 전 트리의 실측.
- **Green path** (M2가 뒤집는다): `grep -c -- "--clear-policy relaunch" README.md README.ko.md README.ja.md README.zh.md` → 각 파일 `0`, exit 1 (짝 없음). 치환 문장은 Q13에 따라 현재 메커니즘(단일 레인 세션이 `moai todo --auto` 큐를 소비)을 이름 붙인다(문장 대조는 M3).
- 뮤턴트 검사: 인용 줄째로 삭제만 해도 이 AC는 통과하지만 Q13 위반이다 — 처분 충실은 green path 문장 요건과 M3 대조가 담당한다.

### AC-DPF-002 — factory-mode 4 로캘의 relaunch 셀 서술이 정리됐다 (REQ-DPF-002·003) — release-blocking

- **RED-now** (E-02, 재측정 트리 13cfafa454 — 74ad197380 핀 대비 docs 바이트 무변경 확인): 네 로캘 77행 `relaunch` 셀이 "감독 런처가 카드마다 다음 카드용 새 세션을 띄운다"를 현재 의미로 서술한다(대체된 카드별 순환). 빨간 이유: 동일 — 구현 전 실측.
- **Green path** (M2): E-02의 명령 → 출력 없음, exit 1. `--clear-policy <값>` 행 자체와 `clear-when-full` 셀 서술은 유지되고, `relaunch` 셀과 `clear-each` 셀의 `/clear` 요청 서술은 치환된다(Q13) — `relaunch`는 현재 소비 서술로, `clear-each`는 세션 안 카드 전환(`ExitWorktree(keep)` → `moai factory next` → `EnterWorktree`) 서술로(plan.md F.2 M2, spec.md 범위표 77행).
- 로캘별 문구가 달라도 검사는 의미 조각(각 로캘의 "카드마다 새 세션" 절)을 때린다 — E-02의 교차 로캘 정규식이 그 역할을 한다.

### AC-DPF-003 — `--factory-run` 인용 보존 (REQ-DPF-005) — regression-guard

- 현재 동작(선택자·합류·이어받기)을 서술하는 인용이므로 구현 전 트리에서 빨간색이 될 수 없어 회귀 가드로 분류한다(§2.1 vacuous 방향 — 단, green path와 짝을 이뤄 "전부 삭제" 뮤턴트를 막는다).
- **Green path** (M3): E-03 명령 재실행 → 파일별 카운트가 baseline과 동일(factory-mode 각 2, launchers 각 1, README 각 2), exit 0.
- 릴리스 블로킹 아님 — 이 기준 위반은 AC-DPF-001·002의 green 상태에서도 남는 과잉 삭제를 잡는 가드다.

### AC-DPF-004 — `--clear-policy` 선택 인용 보존 (REQ-DPF-005) — regression-guard

- **Green path** (M3): E-04 명령 재실행 → factory-mode 각 2 이상, launchers 각 1, README 각 1 이상. exit 0. 카운트는 치환 문장이 살아 있는 `--clear-policy`를 계속 이름 붙이므로 유지된다(README는 74행 치환으로 2→1이 될 수 있다).
- launchers 37행의 값 나열과 `clear-when-full`의 임계치 조건 서술은 LIVE — 무변경. factory-mode 77행의 `clear-each` 셀·134행과 README 87행의 레인 `/clear` 안내는 STALE-3(iter-1 D2 재분류)으로 REQ-DPF-003 치환 대상이며, 이 AC가 보존하는 것은 그 줄의 플래그 인용 존재다.

### AC-DPF-005 — 4 로캘 패리티, 단일 변경 세트 (REQ-DPF-004, Q17) — regression-guard

- 구현 전 트리에서 이미 대칭이라 RED-now가 불가능하다 — Q17 운영자 결정(d-20261019T015908Z-220a)의 절차 위반(분산 편집)을 잡는 절차 가드로 분류한다.
- **Green path** (M2·M3): `grep -c -- "--clear-policy"`와 `grep -c -- "--factory-run"`을 문서군별로 비교 → factory-mode ×4 동일, launchers ×4 동일, README ×4 동일. 런 페이즈 결과물은 하나의 커밋(또는 같은 변경 세트의 커밋들)에 담긴다.

### AC-DPF-006 — internal/·cmd/ 무변경 (REQ-DPF-006) — regression-guard

- 음(-) 검증: 변경이 일어나기 전에는 공집합이라 RED-now가 불가능하다 — help-text 추가(Out of Scope)가 코드로 새는 것을 막는 범위 가드다.
- **Green path** (M3, iter-1 D3 강화 — 네 상태 전부 열거): 검증자가 판정하는 트리 상태를 두 단일 명령으로 검사한다.
  1. `git diff --name-only <base>..HEAD -- internal/ cmd/` → 출력 없음, exit 0 — 커밋된 변경 없음.
  2. `git status --porcelain -- internal/ cmd/` → 출력 없음, exit 0 — 스테이지된 변경·작업 트리 수정·추적되지 않는 새 파일 전부 없음(커밋 전 트리에서도 판정 유효 — 스테이지만 한 변경도 이 명령에 잡힌다).
- **재핀 규칙**(verification-completeness §4): `<base>`는 판정 시점의 감사 대상 커밋 — 런 페이즈 변경 세트 이전의 HEAD SHA — 이며 이동 브랜치명이 아니다. 플랜 페이즈 측정 핀은 74ad197380이고, 런 페이즈는 실제 판정 기준 커밋으로 재핀해 그 SHA를 증거 행에 기록한다.
- **허용 목록 대조**(iter-2 D8 확대): E4의 두 명령, 전체 `git status --porcelain`, 그리고 경로스펙 없는 `git diff --name-only <base>..HEAD`(커밋된 변경 경로 전체 — 명령 1의 pathspec 한정이 internal/·cmd/ 밖 커밋을 못 보는 구멍과, 커밋 후 전체 porcelain이 비게 되는 구멍을 함께 닫는다)의 합집합(변경 경로 전체)은 plan.md §D의 12개 허용 파일 안에 있어야 한다 — 목록 밖 경로 하나라도 있으면 이 AC는 FAIL이다.
- 뮤턴트 폐쇄(iter-1 D3): 초판의 단일 커밋 범위 diff는 (i) 스테이지만 한 금지 변경과 (ii) 언트래킹 신규 파일을 놓쳤다 — 명령 2가 둘 다 닫고, 허용 목록 대조가 범위 밖 커밋 경로를 닫는다.

### AC-DPF-007 — STALE-3 레인 `/clear` 안내가 12줄 전부 사라진다 (REQ-DPF-003) — release-blocking (iter-2 D7 신설)

- **RED-now** (E-05, 트리 b484ecacce — 이 실행): 12줄(factory-mode 77행 `clear-each` 셀 ×4·134행 ×4, README 87행 ×4)이 대체된 레인 `/clear` 안내(STALE-3)를 현재 의미로 아직 서술한다. 빨간 이유: 그 12줄이 구현 전 트리에 존재하기 때문 — E-05의 교차 로캘 문구 12개가 전부 적중한다(출력 12줄 전부 범위표의 STALE-3 줄).
- **Green path** (M2가 뒤집는다): E-05 명령 → 출력 없음, exit 1 — 12줄 전부 제거. 치환은 삭제가 아니다: 같은 줄의 치환 서술은 plan.md F.2 M2의 현재 메커니즘(세션 안 카드 전환)을 이름 붙이고 `--clear-policy` 인용을 유지하며, 그 존재는 E-04 카운트(AC-DPF-004 — factory-mode 각 2, launchers 각 1, README 각 1 이상)가, 치환 문장의 정합은 M3 대조가 확인한다.
- 뮤턴트 검사(iter-2 D7): STALE-3 치환을 전부 스킵해도 AC-DPF-001~006의 green path는 전부 통과하던 vacuous-green 구멍을 이 AC가 닫는다. 반대 방향 뮤턴트 — 12줄째로 삭제 — 는 E-05를 통과하지만 AC-DPF-004(E-04 카운트)가 과잉 삭제로 잡는다.

## D. Quality gates / Definition of Done

- AC-DPF-001·002·007 PASS(각각 RED-now → green 플립 관측됨) + AC-DPF-003~006 PASS.
- 모든 AC 보고는 명령 + 원문 출력 + 트리 귀속을 동반한다(verification-claim-integrity §1·§2; manager-develop §E 형식).
- 뮤턴트 점검 기록: "전부 삭제" 뮤턴트는 AC-DPF-003·004가, "한 로캘만 편집" 뮤턴트는 AC-DPF-005가, "help 추가 우회" 뮤턴트는 AC-DPF-006이, "STALE-3 치환 전면 스킵" 뮤턴트는 AC-DPF-007이 막는다. 넷 중 하나라도 뮤턴트를 통과시키는 구현은 이 SPEC의 DoD가 아니다.
- 선택 항목: Hugo 툴체인이 있으면 경고 없는 문서 빌드 확인 — 없으면 보고서 Gaps에 명시(침묵 생략 금지).
