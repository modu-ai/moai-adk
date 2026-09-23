# SPEC-APPJS-FIRE-GUARD-001 — 진행 기록

- SPEC: SPEC-APPJS-FIRE-GUARD-001
- card: t1060
- worktree: `.claude/worktrees/t1060` (branch `WT-appjs-handler-guard`)
- plan-phase base: `0314801c2` (local develop)

---

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-22

**Claim** — plan-phase 산출물 4종(spec.md / plan.md / acceptance.md / progress.md)이 작성됐고, 이 SPEC 이 딛고 선 실측값 — 특히 런타임 발화 가드의 **양방향 프로토타입** — 은 이 트리에서 이번 실행으로 측정됐다. iter-1 감사(FAIL 0.75, `.moai/reports/t1060/plan-audit.md`)의 결함 D1~D9 를 수리했고, RED-now 4요소 셀의 재측정은 HEAD `d726ac709` 위에서 이번 iter-2 실행으로 이루어졌다.

### plan-audit 이력

| iter | 판정 | 내용 |
|---|---|---|
| 1 | FAIL 0.75 (역치 0.80) | D1 REQ 3건 무매핑 · D2/D3 two-cell 4요소 부재 · D4 판정 명령 부재 · D5 자리표시자 · D6 완료 신호 토큰 부재 · D7 동일 측정 상이 인용 · D8 육안 단계 · D9 skip 문구 — 수리는 acceptance.md §A/§B/§B2/§C/§D 와 아래 iter-2 증거 |
| 2 | (재심사 대기) | 본 커밋 |

**Evidence**

- 배차문 전제 재검증(모두 적중):
  - `git grep -o 'addEventListener' -- internal/web/assets/app.js | wc -l` → `20`
  - `git merge-base --is-ancestor 4db730b72 HEAD; echo $?` → `exit=0`
  - `git merge-base --is-ancestor f4466cb67 HEAD; echo $?` → `exit=0`
  - `git log --all --oneline -S 'addEventListener' -- internal/web/assets/app.js | grep b1ab60454` → `b1ab60454 feat(SPEC-WEB-CONSOLE-001): M1 브라우저 기반 설정 CRUD 콘솔 구현 (cycle_type=tdd)` (양성 대조 발화; 전체 16행)
  - `sed -n '125p;210p' .github/workflows/ci.yml` → `os: [ubuntu-latest]` / `go test -json -coverprofile=coverage.out -covermode=atomic ./... > test-stream.json || rc=$?`
  - `git grep -n '9333' -- internal/web/` → 출력 없음, `grep_exit=1`
- 트리 빌드: `(cd $WT && go build -o /tmp/t1060-probe/moai ./cmd/moai)` → `BUILD_OK`, `moai version` = v3.1.3 `VERSION_EXIT=0`
- 예시 탐침 판독: `.claude/worktrees/t1041/.moai/reports/t1041/browser-probe.py` (157행, 읽기 전용 — 원본 미수정, `/tmp` 사본으로 실행) + t1041 판정서 E3(양군 대조) 판독
- **iter-2 재측정 (모두 HEAD `d726ac709`, 2026-09-22)** — acceptance.md §B2 증거 장부의 원문:
  - `go test ./internal/web/ -run 'AppJsHandlersFireRuntime' -v -count=1` → `testing: warning: no tests to run` / `PASS` / `ok … [no tests to run]`, exit 0 (E1)
  - `go test ./internal/web/ -run 'AppJsHandlersFireSelectorMiss' -v -count=1` → 동일 형태, exit 0 (E2)
  - `ls internal/web/testdata/appjs_fire_mutation.py` → `No such file or directory`, exit 1 (E3); `ls internal/web/testdata/appjs_fire_probe.py` → 동일, exit 1 (E4)
  - 돌연변이 사이클 재실행 → `MUTATED removed@727 inserted@552`, 탐침 전체 붕괴 JSON + `PROBE_EXIT=0`(E5 — AC-AFG-002 의 올바른-이유 적색), `RESTORED_BYTE_IDENTICAL`, assets porcelain 빈 출력
- **정방향 프로토타입**(실바이너리 `moai web --port 18441 --no-open --no-reuse` + 로컬 Chrome `--headless=new --remote-debugging-port=9333`): 5 지표 전부 발화, ReferenceError 0건 — **이하 두 JSON 이 이 측정의 유일한 정본 원문이다.** spec.md §B.1/§B.2 와 acceptance.md §B2 E5 는 이것과 바이트 동일하게 인용한다(iter-2 D7 수리). 출력 그대로:

```json
{
  "label": "t1060-baseline",
  "port": "18441",
  "p1_load_referenceerrors": [],
  "p1_has_glm_btn": true,
  "p2_revealed_hidden_before": true,
  "p2_revealed_hidden_after": false,
  "p2_glm_handler_fired": true,
  "p3_swap_clicked": true,
  "p3_swap_referenceerrors": [],
  "p3_url_after_swap": "/todo",
  "p4_load_referenceerrors": [],
  "p3_has_copy_btn": true,
  "p4_label_after_click": "✓",
  "p4_label_before_click": "Copy",
  "p4_copy_handler_fired": true
}
```

- **역방향 프로토타입**(727행 등록을 합성 좌표 552행으로 이동하는 돌연변이 → 재빌드 → 동일 탐침): 발화 지표 전부 붕괴 + 3Phase ReferenceError — 출력 그대로:

```json
{
  "label": "t1060-mutation",
  "p1_load_referenceerrors": [
    "Uncaught ReferenceError: stampRefreshed is not defined\n    at http://127.0.0.1:18442/static/app.js:552:49\n    at http://127.0.0.1:18442/static/app.js:661:3"
  ],
  "p1_has_glm_btn": true,
  "p2_revealed_hidden_before": true,
  "p2_revealed_hidden_after": true,
  "p2_glm_handler_fired": false,
  "p3_swap_clicked": true,
  "p3_swap_referenceerrors": ["(동일 ReferenceError, app.js:552)"],
  "p3_url_after_swap": "/todo",
  "p4_load_referenceerrors": ["(동일 ReferenceError, app.js:552)"],
  "p3_has_copy_btn": true,
  "p4_label_after_click": "Copy",
  "p4_label_before_click": "Copy",
  "p4_copy_handler_fired": false
}
```

  돌연변이 합성: `MUTATED removed@727 inserted@552` — 복원: `RESTORED_BYTE_IDENTICAL` (`cmp`), 종료 후 assets 변경 0건.
- 상호작용 인벤토리: `grep -n -E "addEventListener\((['\"])(click|submit|change|input)" internal/web/assets/app.js` → 13줄 (73, 83, 109, 158, 327, 352, 406, 414, 473, 513, 530, 603, 648)
- 기반시설: Chrome 존재(`/Applications/Google Chrome.app`), `python3 -c "import websockets"` → 15.0.1, `golang.org/x/net v0.58.0` 직접 의존(go.mod 27행, indirect 마커 없음), `internal/web/server.go` `NewServer`/`Handler()` 테스트면, `moai web [--port] [--no-open] [--no-reuse]` 플래표면(internal/cli/web.go)

**Baseline-attribution** — 전부 worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1060`, branch `WT-appjs-handler-guard`, base `0314801c2` 위에서 2026-09-22 이번 plan-phase 실행 중에 측정. 바이너리는 이 트리에서 빌드해 경로로 호출했다(§2.2 도구 출처).

**Gaps**

- `go test ./internal/web/` 를 **돌리지 않았다** — plan-phase 는 저작만 하며 기존 패키지 테스트의 현재 상태는 이 실행에서 측정되지 않았다.
- in-process 서버면(`Handler()` + httptest)은 **재지 않았다** — 프로토타입은 실바이너리 표면만 쟀다. 두 서면의 등가는 plan.md §E M1 이 측정으로 확정한다.
- 예시 탐침의 exit 계약(3값)은 미구현 상태 그대로다 — 프로토타입 탐침은 양방향 모두 `PROBE_EXIT=0` 을 출력했다(보고 전용). AC-AFG-002 의 exit 1 은 run-phase 산출물이다.
- 셀렉터 미달 방향(AC-AFG-004)과 skip 사유 방향(AC-AFG-003)의 **green 통과 형태**는 아직 관측되지 않았다 — 그 적색(빈 스윕)은 iter-2 에서 E1/E2 로 관측됐다. iter-1 이 이 둘을 「관측 불가」로 기록한 것은 틀린 주장이었고(빈 스윕은 관측 가능한 상태다 — iter-1 감사 D2), acceptance.md §B2 셀과 §D 선언으로 교체했다.
- ubuntu-latest 러너의 Chrome 사전 설치 여부는 검증하지 않았다 — REQ-AFG-010 이 그 전제를 쓰지 않도록(고정 다운로드) 설계에 반영했다.
- 러너에서의 Chrome-for-Testing 다운로드·pip 설치 실측은 없다 — M4 의 몫이다.
- glm 버튼이 돌연변이에서 죽은 **기전**(등록이 552 이후 최상위 효과에서 일어나는가)은 가설로만 표시했다 — 관측(붕괴)은 실측이지만 기전은 미확인이다.

**Residual-risk**

- 탐침의 고정 drain 대기(4.0s 등)는 느린 환경에서 간헐 red 의 씨앗이다 — plan.md §F 의 완화(조건 폴링 전환·플레이크 기록)로 관리한다.
- Chrome/CDP 프로토콜·websockets 라이브러리의 미래 버전이 탐침을 깎을 수 있다 — 버전 고정(REQ-AFG-010)이 완화다.
- 돌연변이 프로토타입은 단일 결함 계열만 쟀다 — 이 가드가 그 계열 이외의 발화 결함을 잡는다는 주장은 매니페스트가 그 경로를 행사할 때만 성립한다(acceptance.md §D).

---

### 개정 plan-phase (card t1106, 2026-09-23)

plan_status: audit-ready
plan_complete_at: 2026-09-23

**Claim** — `completed` 였던 이 SPEC 이 in-place amendment 로 plan-phase 에 재진입했다(`status: completed → in-progress`, `amendment_of: SPEC-APPJS-FIRE-GUARD-001`, HISTORY `### Amendments`). 개정 범위는 REQ-AFG-014·015 와 AC-AFG-010·011·012·013 신설이며, 기존 REQ-AFG-001~013 / AC-AFG-001~009 의 번호·문언·분류는 불변이다. REQ-AFG-012 는 약화되지 않았다 — 그 조문이 이미 이름한 두 경로 중 둘째(일회용 프로젝트 사본)를 처음 사용한다.

**Evidence** — 전부 이 트리(`.claude/worktrees/t1106`, branch `WT-fireguard-reject-submit`, HEAD `176d8b658`), 2026-09-23 측정. 원문 출력·exit·트리 SHA 4요소는 acceptance.md §B2.1 장부 E6~E9:

- `go test ./internal/web/ -run 'AppJsFireValidationRejectPaints' -v -count=1` → `testing: warning: no tests to run` / `PASS` / `ok … 0.424s [no tests to run]`, exit `0` (E6)
- `go test ./internal/web/ -run 'AppJsFireValidationRejectNoWrites' -v -count=1` → 동일 형태 `0.429s`, exit `0` (E7)
- `go test ./internal/web/ -run 'AppJsFireSandboxPairing' -v -count=1` → 동일 형태 `0.428s`, exit `0` (E8)
- `python3 internal/web/testdata/appjs_fire_probe.py --lint-manifest` → `LINT OK: 8 entries + 7 exclusions cover 13 inventory groups; all effects within ['clipboard', 'label', 'swap', 'tab', 'visibility']; post-swap entry present`, exit `0` (E9 — 보조, 적색 아님)
- `grep -c 'validation-reject' internal/web/testdata/appjs_fire_probe.py` → `0`, exit `1`
- `grep -n 'ProjectRoot:' internal/web/appjs_fire_guard_test.go` → `204:		ProjectRoot:    findRepoRoot(t),`, exit `0` — 오늘의 드라이버가 실저장소 루트를 서빙함의 근거
- `grep -n -E "addEventListener\((['\"])(click|submit|change|input)" internal/web/assets/app.js` → 13행(73/83/109/158/327/352/406/414/473/513/530/603/648) — `INVENTORY_TOTAL = 13` 불변 확인
- `go test ./internal/web/ -run 'AppJsFireSandboxRouting' -v -count=1` → `testing: warning: no tests to run` / `PASS` / `ok … 0.646s [no tests to run]`, exit `0` (E10 — AC-AFG-013 의 RED-now)

**Baseline-attribution** — 위 명령 전부 이 실행, 이 트리(HEAD `176d8b658`)에서 측정. E1~E5 는 다른 트리(`d726ac709`)의 측정이므로 섞어 인용하지 않는다.

**결정 기록** — 사본 서빙 범위는 **안 (a)**(사본은 신설 제출 항목만 서빙, 두 번째 서버 인스턴스, 사본 서빙 표식이 라우팅 키)를 채택했다. 안 (b)(전 항목 사본 서빙)는 기각 — 기존 8항목의 사본 위 무회귀를 먼저 측정해야 하고, 원판이 의도적으로 실저장소를 서빙하던 fidelity 를 합성 사본으로 바꾼다. 사본 수명은 `t.TempDir()` 파생으로만 보증한다(`defer`/말미 제거문 금지).

**Gaps** — (1) 개정분 plan-audit 은 iter-1(FAIL 0.86) 이 실행됐고 iter-2 수리가 적용됐다 — **iter-2 재심사는 아직 실행되지 않았다**(아래 「개정분 plan-audit 이력」). (2) AC-010 의 green 통과 형태(배너 paint 관측), AC-011 의 (c) 방향(합성 쓰기에서 exit 1), AC-012 의 (b) 방향(조건 빠진 합성 항목 거부)은 모두 아직 관측되지 않았다 — 뒤집는 것은 M6·M7 의 몫이다. (3) 배너를 칠하는 구체 술식과 검증 실패 값의 선택은 run-phase 소관으로 남겼다. (4) 사본에 무엇을 복사해야 `/settings` 가 거부 경로까지 도달하는지는 측정되지 않았다(M7 이 측정으로 확정). (5) 안 (a) 를 택했으므로 기존 8항목의 **사본 위** 거동은 측정하지 않았고, 이 개정은 그것을 주장하지 않는다 — 기존 항목은 실루트 서버에 그대로 남는다.

**Residual-risk** — 사본 구성이 부족하면 제출이 거부 경로가 아니라 다른 오류로 끝날 수 있다(그 경우 기계결함 exit 2 로 갈린다). 무쓰기 비교 창이 넓으면 읽기-경로 부수효과가 간헐 적색을 만든다(plan §A0 의 시간 경계로 완화).

### 개정분 plan-audit 이력 (card t1106)

| iter | 판정 | 내용 |
|---|---|---|
| 1 | FAIL 0.86 (Tier M 역치 0.80) | 차단 결함 4건. `.moai/reports/t1106/plan-audit.md`. 역치를 넘겼으나 must-pass 미실패 상태에서 **blocking AC 안의 모순·계약 구멍**을 이유로 FAIL — 합계가 두 blocking AC 의 불일치를 흡수하지 않는다 |
| 2 | PASS-WITH-DEBT 0.90 (Tier M 역치 0.80) | 아래 4건 수리 적용. REQ·AC 신설 0건 — 전부 기존 조문의 절 편집. D1~D4·N3 전부 CLOSED, 차단급 신규 결함 2건(N-1 major / N-2 minor). `.moai/reports/t1106/plan-audit-iter2.md` |
| (인라인 수리) | — | Tier M iter 상한 소진 상태에서 리드 판정으로 N-1·N-2 를 절 편집으로 인라인 수리(아래 블록). REQ·AC 신설 0건 |

**iter-2 처분 (4건 전부 수리, 범위 밖 재저작 없음)** — 2026-09-23, HEAD `176d8b658`:

- **D1(major, blocking) — 사본 base 부재 시 계약 미정의 + AC-001 의 8/9 모호.** 깊은 절반은 **결정으로** 처리했다(리드 지침): **두 계열을 AC 층에서 가른다** — AC-AFG-001 은 상속된 blocking 기준이므로 그 「전 항목」을 아홉째로 넓히지 않고 **실루트 계열 8항목**으로 명시 확정하고, 제출 사이클은 AC-AFG-010(+011·013)이 별도 실행으로 진다. 아홉 항목 중 판정 없이 남는 것은 없다.
  - `spec.md` REQ-AFG-014 (1) 에 두 불릿 신설. ① 표식 항목이 있는데 `--sandbox-base-url` 이 없으면 **exit 2 + 항목 이름 보고**, 침묵 스킵과 primary base 대체 운전을 **이름으로 금지**(exit 1 이 아닌 2 인 이유는 호출자 배선 결함이기 때문 — REQ-AFG-005 의 기계결함 값). ② 일부만 운전하려면 **적극적 축소 선언**이 있어야 한다 — 부재가 축소를 함의하지 않고(shall not), 보고서가 운전 수와 제외 항목 **이름**을 담으며, 제외 집합은 정확히 표식 계열이고 운전 집합이 비면 exit 1(「아무것도 운전하지 않는 선언」으로 초록 불가).
  - `acceptance.md` AC-AFG-001 (a) 를 실루트 계열 8항목으로 확정 + 가름의 근거·비공허성·제출 계열의 귀속처를 본문에 서술. AC-AFG-010 에 「이 실행은 AC-001 의 실행과 별개이고, 제출 계열의 전제는 여기서 전제로 서술된다」를 명시.
  - `plan.md` M7.8 을 같은 집합(8항목 + 적극적 축소 선언 + 제외 항목 이름 보고)으로 정렬하고, 제출 판정이 M7.4 소관임을 명시.
- **D2(major, blocking) — 무쓰기 제외 목록의 범위 무제한.** `spec.md` REQ-AFG-014 (2) 에 범위 금지 추가: 제외는 **행사 요청의 쓰기 이음매가 닿는 경로를 덮을 수 없다**(`/save` 의 경우 사본 루트 안 `.moai/config/sections/**` — `handleSave` → `SyncToProjectConfig`·`writeProjectConfig`). `acceptance.md` AC-AFG-011 (c) 의 합성 쓰기를 「사본 안 임의 파일」 → **이음매 도달 가능 경로**로 고정(§C DoD 의 같은 문장도 함께 정렬).
- **D3(minor, blocking) — AC-008 한계 축 스테일.** `네 한계 축` → **`일곱 한계 축`**, 앵커 토큰 3개 신설(`paint`=§F 5, `sandbox-root`=§F 6, `two-surfaces`=§F 7) + grep 판정 명령 3행 추가. `plan.md` M6.6 이 `paint`·`sandbox-root` 를, **M7.7 이 §F 7 과 `two-surfaces` 를** 지도록 배정(종전에는 축 7 이 어느 마일스톤에도 배정되지 않았다).
- **D4(major, blocking) + N3 — 조건 표식 수 모순(AC-009 「세 조건」 vs AC-012 「두 조건」).** `acceptance.md` AC-AFG-009 를 **(1)·(2) 두 조건 표식**으로 고치고, 조건 (3)(`t.TempDir()` 파생 수명)은 매니페스트가 운반할 수 없는 Go 쪽 속성이므로 **AC-AFG-011 (d)** 가 진다고 명시. `spec.md` §G (7) 을 같은 분담으로 정렬(N3).

**iter-2 가 건드리지 않은 것** — 감사가 닫혔다고 확인한 두 결정(안 (a) 사본은 신설 항목만 서빙 / 사본 서빙과 무쓰기 단언의 불가분 쌍)은 재개방하지 않았다. REQ-AFG-012 본문 무수정. 조건부 계열은 여전히 **닫힌 열거**(오늘 원소 1개, `validation-reject`)이며 「표식만 갖추면 통과」 규칙으로 넓어지지 않았다. `moai spec audit` 의 `SyncStatusDrift` MUST-FIX 는 **오탐 — 별도 카드 소관**이다(리드 처분). 수리하지 않았고 우회도 넣지 않았으며, 어떤 요구사항 문언도 이것을 이유로 바뀌지 않았다.

**N2(REQ-AFG-010 의 「전용 job 이 이 가드를 돌린다」 절반이 개정분 테스트에 대해 AC 미커버)** — 리드 확인대로 **상속된 부채**이며 이 카드에서 닫지 않는다. AC 를 더하지 않았고 커버리지를 그대로 뒀다 — 카드 판정서에 상속 부채로 기록된다.

**iter-2 판정 뒤 인라인 수리 (N-1·N-2, 2건, card t1106)** — 2026-09-23, HEAD `176d8b658`. 리드 판정으로 두 건만 절 편집한다. REQ·AC 신설 0건, 결정 재개방 0건.

- **N-1(major, blocking) — 운전 집합이 호출자 선언으로 축소 가능해진 뒤에도 AC-AFG-001 (a) 의 하한이 `> 0` 에 머물러 있었다.** iter-2 의 **적극적 축소 선언**(REQ-AFG-014 (1))이 운전 집합을 호출자가 줄일 수 있게 만들었는데, (a) 는 여전히 「수가 0보다 크며 그 전부가 발화」로만 쟀다 — 실루트 8항목 중 일곱을 선언으로 제외하면 하나를 운전·발화시키고 **blocking AC 를 통과**한다. REQ-AFG-014 (1)(iii)(제외 집합은 정확히 표식 계열, 운전 집합이 비면 exit 1)과 (ii)(운전 수·제외 이름 보고)가 `shall` 로 금지하지만 **그것을 재는 AC 가 없었다**. 수리: AC-AFG-001 (a) 의 판정을 **개수가 아니라 술어**로 바꿨다 — 운전 집합이 「사본 서빙 표식이 없는 항목 전부」와 **정확히 일치**해야 하고(표식 없는 항목이 하나라도 제외되면 실패), 비공허성(`> 0` + 전부 발화)은 유지하며, 제외 집합이 정확히 표식 계열임을 확인할 수 있도록 **운전 항목 수 + 선언 제외 항목 이름**의 보고서 기재를 Then 절에 추가했다. 이 문안은 같은 AC 의 「가름의 형태」 블록과 REQ-AFG-014 (1)(i)(ii)(iii) 이 이미 쓴 성질을 판정 절로 승격한 것이고, 새 기제를 도입하지 않는다. 「8」이라는 스냅숏 수치는 술어의 괄호 주석으로 내렸다(run-phase 가 표식 없는 항목을 더해도 스테일해지지 않는다).
- **N-2(minor, blocking) — AC 매트릭스 서사가 상속 AC 문언 불변을 거짓 주장.** `acceptance.md` §A 의 `AC-001~008 의 번호·문언·분류는 불변이다` 는 이 개정분에서 거짓이다 — AC-AFG-001 의 Then(판정 집합)과 AC-AFG-008 의 Then(한계 축 4→7 + 앵커 토큰 3개)이 재저작됐다. 수리: 참인 부분(**번호·분류·대응 REQ 불변**)은 보존하고, 문언 개정 **세 곳**(AC-AFG-001 판정 집합의 술어화 + 보고서 기재, AC-AFG-008 축 4→7 및 `paint`·`sandbox-root`·`two-surfaces`, AC-AFG-009 효과 종별 계열)을 이름으로 적었다. N-1 수리 뒤의 **최종 상태**를 서술한다.

**인라인 수리가 건드리지 않은 것** — 8항목 분리 결정, 축소 선언의 세 조건, `spec.md` §D 의 `세 조건` 문언(정확하다 — 조건은 셋, 매니페스트 표식으로 표현되는 것이 둘), AC-AFG-009 조건부 계열의 닫힌 열거, REQ-AFG-012 본문. N-3·N-4·N-5(optional)는 닫지 않았다 — §C DoD 확장(N-5)은 감사자의 N-1 권고에 포함돼 있으나 이 위임의 범위 밖이라 미수행 부채로 남긴다.

**감사 창 안의 외부 쓰기 (프로세스 결함, card t1106)** — 2026-09-23 **12:49:11**. plan-audit iteration 2 가 진행 중인 트리에 SPEC 4파일이 통째로 다시 쓰였고, 변경은 표면이 아니라 실질이었다 — D1 처분이 9항목에서 8항목 + `적극적 축소 선언` 으로 반전됐다.

- **원인.** 레인이 iteration-2 수리 에이전트를 **띄운 뒤에** 리드의 D1 방향(AC 층에서 사이클 분리)을 후속 메시지로 전달했다. 그 에이전트는 이미 9항목 안을 만들어 완료 보고를 마친 상태였고, 뒤늦게 깨어나 분리안을 적용하면서 같은 트리의 **두 번째 작성자**가 됐다. 지시는 spawn 프롬프트에 실렸어야 했다. 원인의 일부는 리드 쪽에도 있다(리드 자인) — 그 메시지가 iteration 2 가 이미 떠 있는 시점에 도착했다.
- **탐지.** 감사자가 이미 읽은 파일을 다시 읽어 내용이 달라진 것을 발견하고, `stat` 으로 확인한 뒤 **조용히 진행하지 않고 보고**했다. 그리고 판정을 명명된 스냅샷에 고정해 보고서가 어느 텍스트를 판정한 것인지 말하게 했다.
- **조치.** 보고 즉시 `TaskStop` → 4파일 mtime `12:49:11` 고정 확인 → `md5 -q acceptance.md` = `c4bace1f157a0bdafa244acd0d78882b` 가 감사자 핀과 바이트 동일함을 대조 → 감사자에게 "핀이 곧 디스크 상태이니 확정하라, 원인이 레인 쪽이라고 판정을 무르게 하지 말라" 회신.
- **폐기된 것.** 레인이 9항목 판본에서 수행한 D1~D4 확인은 폐기 텍스트에 대한 것이므로 이월하지 않았고, 감사자에게 고정 스냅샷에서 전부 재도출하도록 지시했다. 감사자는 그 확인 내용을 애초에 본 적이 없다고 판정서에 적었다.
- **손실.** 없음. 비용은 레인의 중간 검증 한 회차가 다시 필요해진 것뿐이다.
- **재발 방지.** 서브에이전트에 지시를 추가할 때는 ① spawn 프롬프트에 싣거나 ② 보내기 전에 종료하고 새로 띄우거나, 둘 중 하나만 한다. 그리고 후속 전송 전에 "이 에이전트가 대기 중인가"만이 아니라 **"지금 이 트리를 읽고 있는 다른 주체가 있는가"** 를 함께 묻는다 — 감사·리뷰·검증 패스가 열려 있으면 안전한 후속은 없다. spawn 이후 도착한 지시는 **다음 spawn 의 프롬프트**로 간다. 이후 이 카드의 모든 에이전트(`manager-spec` ×3, `plan-auditor` ×2)는 보고 직후 종료하고 그 이름으로 메시지를 보내지 않았다.
- 당대 기록 정본: `.moai/reports/t1106/incident-audit-window-write.md`. 교훈은 레인 메모리의 기존 항목에 **3번째 발생**으로 갱신했다(신규 파일 아님) — 그 항목의 종전 "how to apply" 가 이 경우엔 틀린 답을 준다는 정정을 함께 실었다.

## §E.2 Run-phase Evidence

run-phase 측정 전체는 worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1060`, branch `WT-appjs-handler-guard` 위에서 2026-09-22 이번 run 실행 중 수행됐다(각 측정 시점 HEAD는 항목별로 병기). 바이너리는 이 트리에서 빌드해 경로로 호출했다(`go build -o /tmp/t1060-run/moai ./cmd/moai` — §2.2 도구 출처).

### M1 — 탐침 저작 + 매니페스트 + 3값 exit 계약 (HEAD `3e35fbacf` + M1 커밋)

| # | 측정 | 명령 | 관측 결과 | exit |
|---|---|---|---|---|
| M1-1 | 매니페스트 자기검증 (AC-AFG-009 green) | `python3 internal/web/testdata/appjs_fire_probe.py --lint-manifest` | `LINT OK: 8 entries + 7 exclusions cover 13 inventory groups; all effects within ['clipboard', 'label', 'swap', 'tab', 'visibility']; post-swap entry present` | 0 |
| M1-2 | 정방향 baseline 재측정 — 실바이너리 표면 (spec §B.1 확장: 신규 지표 popover×3·tabs·post-swap 포함, 전부 발화) | `python3 internal/web/testdata/appjs_fire_probe.py --cdp-port <cdp> 18511 t1060-m1-healthy` | 아래 JSON 전문 — 전 지표 true, `failures: []`, `missing_selectors: []` | 0 |
| M1-3 | 셀렉터 미달 RED — **계약 이전 계측기**(t1041 예시 탐침, report-only; `[data-copy]`→`[data-copy-nonexistent]` 변조, CDP 포트만 /tmp 사본에서 재지정) | `python3 /tmp/t1060-run/exemplar-tampered-final.py 18512 t1060-m1-tampered-exemplar` | `p3_has_copy_btn: false`, `p4_copy_handler_fired: false` — 지표 붕괴에도 **exit 0** (acceptance §B2 E5 와 같은 모양의, 올바른-이유 적색 — exit 계약이 없어서 통과 코드를 낸다) | 0 (적색) |
| M1-4 | 셀렉터 미달 GREEN — 커밋된 탐침, 동일 변조 | `python3 /tmp/t1060-run/committed-tampered.py --cd-port <cdp> 18512 t1060-m1-tampered-committed` | 아래 실패 블록 — `copy_button` / `[data-copy-nonexistent]` 를 **이름으로** 지목 | 1 |

M1-2 정방향 baseline 전문 (실바이너리 `moai web --port 18511 --no-open --no-reuse`, 로컬 Chrome headless, CDP `--remote-debugging-port=0` → `DevToolsActivePort` 자동 탐색):

```json
{
  "label": "t1060-m1-healthy",
  "port": "18511",
  "base_url": "http://127.0.0.1:18511",
  "cdp_port": 65110,
  "p1_load_referenceerrors": [],
  "p1_has_glm_btn": true,
  "p2_revealed_hidden_before": true,
  "p2_revealed_hidden_after": false,
  "p2_glm_handler_fired": true,
  "p3_panel_hidden_before": true,
  "p3_panel_hidden_after_open": false,
  "p3_popover_open_fired": true,
  "p3_has_close_btn": true,
  "p3_panel_hidden_after_close": true,
  "p3_popover_close_btn_fired": true,
  "p3_panel_hidden_after_outside": true,
  "p3_popover_outside_close_fired": true,
  "p4_tab_count": 14,
  "p4_tab_selected_before": 0,
  "p4_tab_clicked": 1,
  "p4_tab_selected_after": 1,
  "p4_settings_tabs_fired": true,
  "p5_swap_clicked": true,
  "p5_url_after_swap": "/todo",
  "p5_swap_referenceerrors": [],
  "p6_panel_hidden_before": true,
  "p6_panel_hidden_after": false,
  "p6_popover_after_swap_fired": true,
  "p7_load_referenceerrors": [],
  "p7_has_copy_btn": true,
  "p7_label_before_click": "Copy",
  "p7_label_after_click": "✓",
  "p7_copy_handler_fired": true,
  "failures": [],
  "missing_selectors": [],
  "exit": 0
}
```

M1-4 실패 블록 (판정점 부분 — 전문은 셀렉터 미달 시 나머지 지표가 모두 정상 발화함도 함께 보여준다):

```json
  "failures": [
    {
      "entry": "copy_button",
      "reason": "selector matched nothing",
      "selector": "[data-copy-nonexistent]"
    }
  ],
  "missing_selectors": [
    {
      "entry": "copy_button",
      "selector": "[data-copy-nonexistent]"
    }
  ],
  "exit": 1
```

**run 중 발견·수리한 결함 (M1):** 탐침 초판의 judge 가 미달 셀렉터를 `missing_selectors` 에 기록만 하고 실패로 합산하지 않아, 변조 실행에서 exit 0 을 냈다(M1-3 세션 1차 실행 — 관측 원문: `missing_selectors` 채워지고 `exit: 0`). REQ-AFG-004(스테일 셀렉터=red) 위반이며, 수리는 judge 가 미달 셀렉터를 `failures` 에 합산하도록 1개 블록 추가. 수리 뒤 동일 변조가 exit 1 로 뒤집힘(M1-4). — 계약의 red 가 실제로 관측된 후 green 이 뒤집은 순서로 기록한다.

**서면 등가 측정 일정 주석 (plan §E M1.4):** in-process 서면(`NewServer`+`Handler()`)에 대한 같은 탐침 측정은 M2 의 드라이버가 수행한다 — 드라이버 자체가 in-process 하네스라서 M1 시점에는 운반체가 없다. M1 에서는 실바이너리 표면을 재측정했고, 등가 판정은 M2 녹색 출력에서 같은 지표 집합의 발화로 확정한다(plan §F 위험 표의 「측정으로 확정」 요구는 이 순서로 충족된다).

### M2 — 드라이버 테스트 (HEAD = M1 커밋 위 M2 커밋)

| # | 측정 | 명령 | 관측 결과 | exit |
|---|---|---|---|---|
| M2-1 | 게이트 없는 실행의 이름 붙은 skip (AC-AFG-003 green) | `go test ./internal/web/ -run 'AppJsHandlersFire\|AppJsFireManifest' -v -count=1` | `--- SKIP: TestAppJsHandlersFireRuntime` + 사유행 `MOAI_BROWSER_GUARD is not set to 1 — …` (게이트 변수명 영어 노출), `TestAppJsHandlersFireSelectorMiss` 동일, 그리고 무게이트 정적 검증 `--- PASS: TestAppJsFireManifestInventoryCount` (살아있는 app.js 의 13 그룹 수 ↔ 매니페스트 `INVENTORY_TOTAL=13` 교차검증 — 매니페스트 스테일은 무게이트 영역에서도 적색) | 0 |
| M2-2 | 게이트 켠 전 사이클 — in-process 서면 (AC-AFG-001 green + 등가 측정) | `MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsHandlersFire' -v -count=1 -timeout 10m` | `--- PASS: TestAppJsHandlersFireRuntime (11.68s)` + `--- PASS: TestAppJsHandlersFireSelectorMiss (16.75s)` + `ok github.com/modu-ai/moai-adk/internal/web 28.767s` | 0 |
| M2-3 | in-process 탐침 보고 전문 (등가 증거 — M1-2 실바이너리 보고와 지표별 동일) | (위 실행의 `probe t1060-driver-runtime stdout` 로그) | 아래 JSON 전문 — 8 지표 전부 true, `failures: []`, `missing_selectors: []`, `exit: 0` | 0 |

M2-3 in-process 보고 전문 (랜덤 포트 `50691` 서버, CDP `50692` — `DevToolsActivePort` 자동 탐색):

```json
{
  "label": "t1060-driver-runtime",
  "port": "50691",
  "base_url": "http://127.0.0.1:50691",
  "cdp_port": "50692",
  "p1_load_referenceerrors": [],
  "p1_has_glm_btn": true,
  "p2_revealed_hidden_before": true,
  "p2_revealed_hidden_after": false,
  "p2_glm_handler_fired": true,
  "p3_panel_hidden_before": true,
  "p3_panel_hidden_after_open": false,
  "p3_popover_open_fired": true,
  "p3_has_close_btn": true,
  "p3_panel_hidden_after_close": true,
  "p3_popover_close_btn_fired": true,
  "p3_panel_hidden_after_outside": true,
  "p3_popover_outside_close_fired": true,
  "p4_tab_count": 14,
  "p4_tab_selected_before": 0,
  "p4_tab_clicked": 1,
  "p4_tab_selected_after": 1,
  "p4_settings_tabs_fired": true,
  "p5_swap_clicked": true,
  "p5_url_after_swap": "/todo",
  "p5_swap_referenceerrors": [],
  "p6_panel_hidden_before": true,
  "p6_panel_hidden_after": false,
  "p6_popover_after_swap_fired": true,
  "p7_load_referenceerrors": [],
  "p7_has_copy_btn": true,
  "p7_label_before_click": "Copy",
  "p7_label_after_click": "✓",
  "p7_copy_handler_fired": true,
  "failures": [],
  "missing_selectors": [],
  "exit": 0
}
```

**등가 판정 (plan §E M1.4 / §F 위험 표):** M1-2(실바이너리)와 M2-3(in-process `NewServer`+`Handler()`)의 보고가 같은 탐침으로 **지표별 동일**(8/8 발화, ReferenceError 0건, exit 0) — 두 서면이 같은 app.js 표면을 서빙함을 측정으로 확정. 이 판정이 M3 이후의 서면 혼용(in-process 녹색/실바이너리 적색)을 정당화한다.

드라이버 구조 비고: `appjs_fire_guard_test.go` — 게이트·이름 붙은 영어 skip(`MOAI_BROWSER_GUARD`/`chrome`/`python3`/`websockets` 각각 명명)·`t.Cleanup` 정리(서버·Chrome·탐침 전부)·`--lint-manifest` 게이트 실행 동반·무게이트 매니페스트-인벤토리 교차검증 테스트. `gofmt -l` 출력 없음, `go vet ./internal/web/` 통과, `GOOS=windows GOARCH=amd64 go build ./internal/web/` 통과(실행 전 재확인은 E2).

### M3 — 돌연변이기 + 양방향 재측정 (HEAD = M2 커밋 위 M3 커밋)

AC-AFG-002 의 0→1→0 시퀀스를 커밋된 계측기(탐침 + 돌연변이기)로 실바이너리 표면에서 전 구간 실행. 서버·Chrome 기동 래퍼는 plan §B 스텝 시퀀스와 동일(포트 18513, CDP 51062). acceptance.md §B2 판정 명령 블록의 `/tmp/t1060-probe/probe.py` 자리는 커밋된 `internal/web/testdata/appjs_fire_probe.py` 로 대체해 실행했다 — 프로토타입이 아니라 납품 계측기로의 재측정이 더 강한 근거다.

| # | 측정 | 명령 | 관측 결과 | exit |
|---|---|---|---|---|
| M3-1 | 전제 단언 음성 시험 (REQ-AFG-008 — 패턴 부재 대상) | `grep -v 'stampRefreshed' app.js > /tmp 복사본 && python3 internal/web/testdata/appjs_fire_mutation.py --target <복사본>` | `MUTATION FAILED: registration pattern not found …` + 프리스틴 미생성 확인 | 1 (기대) |
| M3-2 | 변이 전 탐침 | `python3 internal/web/testdata/appjs_fire_probe.py --cdp-port 51062 18513 m3-pre` | 전 지표 발화 | 0 |
| M3-3 | 돌연변이 합성 | `python3 internal/web/testdata/appjs_fire_mutation.py --pristine /tmp/t1060-run/appjs.pristine` | `MUTATED removed@727 inserted@552 pristine=/tmp/t1060-run/appjs.pristine` — plan-phase 프로토타입과 동일 좌표 | 0 |
| M3-4 | 변이 재빌드 뒤 탐침 — **AC-AFG-002 판정점** | (재빌드 + 서버 재기동 뒤) `python3 internal/web/testdata/appjs_fire_probe.py --cdp-port 51062 18513 m3-post-mutate` | **exit 1** — 아래 요지 발췌: `stampRefreshed is not defined` ReferenceError(load·swap 창) + 미발화 지표 `glm_reveal`·`copy_button` 이 **이름으로** 보고됨 | 1 |
| M3-5 | 복원 + byte 동일 검증 (REQ-AFG-009) | `cp pristine app.js && cmp app.js pristine && git diff --quiet -- app.js && git status --porcelain -- internal/web/assets/` | `RESTORED_BYTE_IDENTICAL (cmp exit 0)` · `GIT_DIFF_VS_HEAD_EMPTY` · `ASSETS_PORCELAIN_LINES= 0` | 0 |
| M3-6 | 복원 재빌드 뒤 탐침 | `python3 internal/web/testdata/appjs_fire_probe.py --cdp-port 51062 18513 m3-post-restore` | 전 지표 발화, `exit: 0` | 0 |

M3-4 판정점 발췌 (`/tmp/t1060-run/probe-m3-post-mutate.json` 요지 — 전문은 실행 산출물):

```text
exit: 1
p1_load_referenceerrors: ["Uncaught ReferenceError: stampRefreshed is not defined\n    at http://127.0.0.1:18513/static/app.js:552:49\n    at http://127.0.0.1:18513/static/app.js:661:3"]
p2_glm_handler_fired: False
p7_copy_handler_fired: False
failure entries: ['glm_reveal', 'copy_button', None, None]   (None = ReferenceError 창 실패 2건)
```

정직한 관측 하나: 변이 상태에서도 `p6_popover_after_swap_fired` 는 true 였다 — popover 등록 계열(line 73)은 삽입점(합성 552행)보다 앞서 살아남기 때문이다. 이는 spec §B.2 가 이미 표시해 둔 「삽입점 이후 최상위 효과 전멸」 기전 가설과 정합하는 관측이며, AC-AFG-002 가 요구하는 것은 최소 1개 지표의 미발화 또는 ReferenceError — 충족된다. 가드의 red 가 모든 지표의 동시 붕괴를 함의하지 않는다는 것을 이 실행이 재확인했다.

**종료 후 잔여 검증 (세션 종료 시점 재측정):** `git diff --stat 3e35fbacf -- internal/web/assets/app.js` 출력 없음 — base 대비 byte 동일 유지. assets porcelain 0행.

### M4 — test-browser CI job (HEAD = M3 커밋 위 M4 커밋)

| # | 측정 | 명령 | 관측 결과 |
|---|---|---|---|
| M4-1 | 기존 job·step 무변경 (AC-AFG-006a) | `git diff --numstat -- .github/workflows/ci.yml` (커밋 후 base..HEAD 동형) | `157	0	.github/workflows/ci.yml` — **삭제 0**, 추가만 존재 |
| M4-2 | 삭제 행 0 (AC-AFG-006b) | `git diff -- .github/workflows/ci.yml \| grep -c '^-[^-]'` | `0` (grep exit 1 = 매치 없음) |
| M4-3 | 기존 job 키 목록·순서 보존 (AC-AFG-006c) | `grep -n -E '^  [a-z0-9-]+:' .github/workflows/ci.yml` | detect(44)/test(116)/test-race(256)/test-skip-marker(321)/test-integration(373)/lint(427)/build(474)/constitution-check(544) — base 측정치와 동일·동순서, 신설 `test-browser:` 는 605행 EOF 부록뿐 |
| M4-4 | YAML 문법 | `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"` | `YAML_OK` |
| M4-5 | Chrome-for-Testing 고정본 실측 (REQ-AFG-010 — 러너 이미지의 Chrome 을 신뢰하지 않음) | `curl -fsSL https://storage.googleapis.com/chrome-for-testing-public/153.0.8010.52/linux64/chrome-linux64.zip` + `shasum -a 256` | 195,708,470 bytes, `sha256 e66f66d4802a46d4a022667e668aa950e277cadbfbed4b3777915b47413a0ef9` — job 은 이 URL+해시를 그대로 고정하고 `sha256sum --check --strict` 로 검증 |
| M4-6 | websockets 고정 (REQ-AFG-010) | job step: `python3 -m pip install --user "websockets==15.0.1"` | 15.0.1 — 로컬 실측 버전(C4)과 동일 고정. go.mod/go.sum 무관(탐침 유일 서드파티 의존은 Python 쪽) |

job 설계 비고: `test-browser` 는 EOF 에 덧붙는 유일한 판정면이다 — 게이트된 드라이버 테스트는 다른 어디서도 skip 되므로(AC-AFG-003), 런타임 발화 축의 CI 판정은 이 job 만이 운반한다. Green 단계(게이트 켠 드라이버, in-process 서면)와 Red 단계(돌연변이 → 재빌드 → 탐침 exit 1 기대 + `stampRefreshed` 재검증 → 복원 → cmp/git diff byte 동일 → 재빌드 → 탐침 exit 0 기대)를 모두 운반한다. Chrome 은 고정 CfT 다운로드(`MOAI_BROWSER_GUARD_CHROME` 로 드라이버·레드 단계 양쪽에 주입), `--no-sandbox` 는 루프백 전용 시험 브라우저에 한해. **이 job 의 러너 실측(다운로드·pip·전 사이클)은 아직 없다 — 그것은 push 뒤 origin/develop CI 의 몫이며(§4.1 규율), 여기서 로컬 실행하지 않는다(배차문 D 제약).**

### M5 — 문서화 + 전체 재측정 (HEAD = M4 커밋 위 M5 커밋)

| # | 측정 | 명령 | 관측 결과 |
|---|---|---|---|
| M5-1 | 한계·직교성 주석 존재 (AC-AFG-008, REQ-AFG-013) | `grep -c` 앵커 토큰 6종 (acceptance §B AC-008 명령 그대로) | probe: `orthogonal`=1 `manifest`=18 `scenario`=3 `Chrome`=4 / driver: `orthogonal`=1 `SPEC-APPJS-IIFE-GUARD-001`=1 `static-scope`=2 — 전부 ≥1. **측정 경위**: 초판 헤더가 ORTHOGONAL 대문자로 쓰여 AC 의 소문자 토큰 0건 — 주석을 소문자 토큰으로 수리 뒤 재측정 (red 관측 → 수리 → green) |
| M5-2 | 기존 테스트 전체 통과 + 정적 형제 무손상 (AC-AFG-005a) | `go test ./internal/web/ -count=1 -timeout 30m` | `ok github.com/modu-ai/moai-adk/internal/web 25.282s`, exit 0. `git diff --stat 3e35fbacf -- internal/web/appjs_iife_scope_test.go internal/web/appjs_reinit_test.go` 출력 없음 |
| M5-3 | 탐침 존재 + 파생 출처 (AC-AFG-005b) | `ls internal/web/testdata/appjs_fire_probe.py && grep -c 't1041' <동일>` | 파일 존재, `t1041` 4건 |
| M5-4 | go.mod/go.sum 무변경 (AC-AFG-005c) | `git diff --stat 3e35fbacf..HEAD -- go.mod go.sum` + working-tree porcelain | 출력 없음 양쪽 |
| M5-5 | 최종 HEAD 게이트 실행 재측정 | `MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsHandlersFire' -v -count=1 -timeout 10m` | `--- PASS: TestAppJsHandlersFireRuntime (11.75s)` / `--- PASS: TestAppJsHandlersFireSelectorMiss (16.53s)` / `ok … 28.677s` — HEAD `1f9f2d321` (주석 수리 후 최종 원본 바인딩) |
| M5-6 | gofmt·vet·lint (공통 DoD + E5) | `gofmt -l internal/web/appjs_fire_guard_test.go` / `go vet ./internal/web/` / `golangci-lint run --timeout=2m ./internal/web/...` | gofmt 출력 없음 · vet 통과 · `0 issues.` (golangci-lint 2.10.1, homebrew 설치본 — 서드파티 도구로서 트리 래그 개념 외부, 버전 명기) |
| M5-7 | 자산·모듈 보존 최종 (AC-AFG-007/E6) | `git diff --stat 3e35fbacf -- internal/web/assets/app.js` + `git status --porcelain -- internal/web/assets/ go.mod go.sum` | 출력 없음 — base 대비 byte 동일, working tree 청결 |

### M6·M7 — 개정분: `validation-reject` 편입 (card t1106, HEAD `0fbc75afc` → `19f210fe5`)

개정 run 은 worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1106`, branch `WT-fireguard-reject-submit` 에서 2026-09-23 에 수행됐다. 원판 M1~M5 기록(위)은 한 줄도 고치지 않았다 — 나중 사실에 맞춰 관측을 손보지 않는다.

| # | 측정 | 명령 | 관측 결과 | exit |
|---|---|---|---|---|
| M6-1 | AC-AFG-012 RED — 불가분성을 판정할 규칙 자체가 없음 | `go test ./internal/web/ -run 'AppJsFireSandboxPairing' -count=1` | `want exit 1, got exit status 2` / `appjs_fire_probe.py: error: no such option: --extra-entry` (세 합성 항목 전부) | 1 |
| M6-2 | AC-AFG-012 GREEN — 정방향 + 역방향 3건 | 같은 명령 | `--- PASS: TestAppJsFireSandboxPairing` + 하위 3건(`missing_sandbox-serving_marker`·`missing_no-write-assertion_marker`·`missing_both_markers`) 전부 PASS — 각 거부가 빠진 조건을 이름으로 보고 | 0 |
| M6-3 | AC-AFG-009 정방향 — 커밋된 매니페스트가 두 닫힌 계열을 만족 | `python3 internal/web/testdata/appjs_fire_probe.py --lint-manifest` | `LINT OK: 9 entries + 7 exclusions cover 13 inventory groups; effects within unconditional ['clipboard', 'label', 'swap', 'tab', 'visibility'] or conditional ['validation-reject'] (conditional entries carry ['requires_sandbox_serving', 'requires_no_write_assertion']); post-swap entry present` — `INVENTORY_TOTAL` 불변(13) | 0 |
| M7-1 | AC-AFG-013 RED — 라우팅 판정 주체 부재 | `go test ./internal/web/ -run 'AppJsFireSandboxRouting' -count=1` | `undefined: startFireGuardServerAt` / `undefined: startFireGuardSandboxServer` (빌드 실패) | 1 |
| M7-2 | AC-AFG-013 GREEN (a)(b)(c-배선) | 같은 명령 | `--- PASS: TestAppJsFireSandboxRouting` — primary root == `findRepoRoot`, sandbox root ≠ 실루트, 표식 8/1 양방향 일치 | 0 |
| M7-3 | AC-AFG-013 돌연변이 ① — 전 항목을 사본으로 (안 (b) 를 몰래 취한 상태) | `route_for_entry` 의 `return primary_base` → `return sandbox_base or primary_base` 후 같은 명령 | `(b) unmarked entry "popover_open" routes to "http://127.0.0.1:60604", want the primary base "http://127.0.0.1:60603"` 외 7건 | 1 (적색 관측) |
| M7-4 | AC-AFG-013 돌연변이 ② — 표식 항목을 실루트로 | `return sandbox_base or None` → `return primary_base` 후 같은 명령 | `(b) marked entry "validation_reject_banner" routes to "…:60695", want the sandbox base "…:60696"` | 1 (적색 관측) |
| M7-5 | AC-AFG-010/011 RED — 판정 주체는 생겼으나 탐침에 사본 시나리오 없음 | `MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsFireValidationReject' -count=1 -v` | `failures: [{"entry": "validation_reject_banner", "reason": "selector matched nothing", "selector": "#settings-form"}]`, `"exit": 1` — 같은 실행에서 기존 8지표는 전부 발화 | 1 |
| M7-6 | AC-AFG-010 GREEN — 거부 배너가 칠해짐 | `MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsFireValidationRejectPaints' -count=1 -v` | `s_banner_before_submit: false` → `s_invalid_set: "bogus"` → `s_banner_text: "Validation failed — no changes were saved."`, paint 술식 통과, `s_window_referenceerrors: []` / `--- PASS` (38.72s) | 0 |
| M7-7 | AC-AFG-010 돌연변이 — 배너를 `hidden` 으로 렌더 | 같은 실행 안 2회차, `--inject-banner-hidden` | `s_injected_banner_hidden: true` → 탐침 exit 1. 「노드 존재」 술식이었다면 통과했을 변이가 적색 — 술식 채택 가능 | 1 (적색 관측) |
| M7-8 | AC-AFG-011 (b) 정방향 — 사본 바이트 불변 | `MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsFireValidationRejectNoWrites' -count=1 -v` | `s_snapshot_files: 35`, `s_changed_paths: []`, `"exit": 0` | 0 |
| M7-9 | AC-AFG-011 (c) 역방향 — 쓰기 이음매가 닿는 자리에 1바이트 | 같은 실행 안 2회차, `--inject-sandbox-write .moai/config/sections/quality.yaml` | `s_changed_paths: [{"path": ".moai/config/sections/quality.yaml", "change": "modified"}]`, `"exit": 1` — 달라진 경로를 이름으로 보고 | 1 (적색 관측) |
| M7-10 | AC-AFG-011 (d) — 수명 판정의 비공허성 | 드라이버 말미에 `defer os.RemoveAll(sandboxRoot)` 주석 1행 주입 후 같은 명령 | `(d) driver carries "defer os.Remove" …` 2건 적색 → 복원 후 녹색. 첫 판은 자기 픽스처 리터럴에 걸려 적색이었고(자기 참조), 바늘을 문자열 연결로 조립해 수리 | 1 → 0 |
| M7-11 | AC-AFG-001 (a) — 축소 선언과 운전/제외 회계 | `MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsHandlersFireRuntime' -count=1 -v` | `reduction_declared: true`, `driven_count: 8`, `excluded_entries` = 표식 계열, `"exit": 0` / `--- PASS` (12.79s) — 운전 집합이 「표식 없는 전 항목」 술어와 정확히 일치 | 0 |
| M7-12 | REQ-AFG-014 (1)(i) — 부재는 선언이 아니다 | `go test ./internal/web/ -run 'AppJsFireReductionDeclaration' -count=1` | exit 2 + `--sandbox-base-url` 미배선 · `validation_reject_banner` 이름 보고 → PASS | 0 |
| M7-13 | 같은 판정의 비공허성 확인 | `if driven_marked and not sandbox_base:` → `if False and …` 후 같은 명령 | **첫 판은 이 변이를 통과시켰다**(모든 기계결함이 exit 2 이므로 exit 코드만으로는 갈리지 않는다) — 판정에 「명명된 원인」을 더한 뒤 재측정: `the fault does not name the missing wiring (--sandbox-base-url)` 적색 | 1 (적색 관측) |
| M7-14 | AC-AFG-008 — 앵커 토큰 7종 + 직교성 | acceptance §B AC-008 의 `grep -c` 9행 그대로 | probe: `orthogonal`=1 `manifest`=26 `scenario`=6 `Chrome`=4 `paint`=12 `sandbox-root`=5 / driver: `orthogonal`=1 `SPEC-APPJS-IIFE-GUARD-001`=1 `static-scope`=2 `two-surfaces`=1 — 전부 ≥1 | 0 |
| M7-15 | AC-AFG-013 (c) — 전 게이트 사이클 무회귀 | `MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJs' -count=1` | `ok github.com/modu-ai/moai-adk/internal/web 96.737s` | 0 |
| M7-16 | 패키지 전체 + 공통 DoD | `go test ./internal/web/ -count=1` / `go vet ./internal/web/` / `gofmt -l internal/web/` / `golangci-lint run ./internal/web/...` / `go build ./...` | `ok … 23.840s` · vet 통과 · gofmt 출력 없음 · `0 issues.` · build 통과 | 0 |

**사본 구성 실측:** 사본에 복사하는 것은 `.moai/config` 뿐이고(35파일 — `s_snapshot_files`), 그것이 `handleSave` 의 쓰기 이음매(`SyncToProjectConfig`·`writeProjectConfig`)가 겨누는 자리다. 무쓰기 비교의 제외 목록은 **비어 있고**, 그 비어 있음은 가정이 아니라 측정이다 — 정방향 35파일 전수 비교에서 변경 0건. `lint_manifest` 가 `.moai/config/sections` 를 덮는 제외를 기계로 거부하므로, 이 단언이 나중에 「구성상 초록」으로 약해질 수 없다.

**두 서면을 섞지 않는 방식:** 제출 계열(AC-010·011)과 실루트 계열(AC-001)은 **서로 다른 실행**이고, 어느 쪽 초록도 상대의 집합을 대신 주장하지 않는다. 실루트 실행은 축소를 적극 선언하고 운전 8건·제외 1건을 보고서에 남긴다(M7-11). 선언 없는 부재는 exit 2 다(M7-12).

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-22
run_commit_sha: eb624ae89
run_base_sha: 3e35fbacf
ac_pass_count: 6/6 blocking (AC-AFG-001·002·003·004·007·009 전부 관측 출력으로 PASS) + regression-class 3건(AC-AFG-005·006·008) 전부 기록 완료 — §D 분류 선언대로 기록 산출물이며 차단 요건 아님
preserve_list: `internal/web/assets/app.js` base `3e35fbacf` 대비 byte 동일 (`git diff --stat 3e35fbacf -- <경로>` 출력 없음; M3 사이클 복원 구간에서 `cmp` exit 0 + `RESTORED_BYTE_IDENTICAL` 관측) · `go.mod`/`go.sum` 무변경 (websockets 는 Python 의존 — Go 모듈 추가 없음) · 기존 Go 코드 수정 0건(신규는 드라이버 테스트 파일 1개 + testdata 2개) · `appjs_iife_scope_test.go`/`appjs_reinit_test.go` 무손상 · 기존 CI job 8개 무변경(ci.yml numstat 157/0)
cross_platform_build: darwin `go build ./internal/web/` exit 0 · `GOOS=windows GOARCH=amd64 go build ./internal/web/` exit 0 (실행 전 pre-flight와 M5 최종 HEAD 양쪽에서 측정) · 드라이버는 syscall/build-tag 없이 exec/os/net 만 사용
m1_to_m5_commit_strategy: 마일스톤당 1커밋 — M1 탐침(70f37a688, draft→in-progress 전환 포함) → M2 드라이버(8375c25f8) → M3 돌연변이기(5973ac3c9) → M4 CI job(1f9f2d321) → M5 문서화+전체 재측정(pending-backfill-M5). 전 커밋 본문에 card t1060 명기 + `Authored-By-Agent: manager-develop` 트레일러. push·PR 없음(레인 규율 — 리드 일괄)
gate_env: `MOAI_BROWSER_GUARD=1` — 게이트 없는 실행은 게이트 변수명을 영어로 명명하는 skip(AC-AFG-003 관측 M2-1), CI 판정면은 `test-browser` job 이 유일하게 운반
evidence_paths: M1/M2/M3 세션 원문은 `/tmp/t1060-run/` (probe-*.json, m2-runtime.log, m5-fullpkg.log) — 휘발성 스크래치이므로 판정 근거가 되는 명령·출력 요지는 전부 이 §E.2 에 전사했다. §E.2 표가 이 SPEC 판정의 로컬 정본 기록이다.


### 개정분 Run-phase Audit-Ready Signal (card t1106)

run_complete_at: 2026-09-23
run_commit_sha: 19f210fe5
run_base_sha: 0fbc75afc
run_status: PASS
ac_pass_count: 4/4 개정 blocking (AC-AFG-010·011·012·013 전부 양방향 관측 출력으로 PASS) + 상속 AC-AFG-001 개정 문언(운전 집합 술어 + 회계) PASS + AC-AFG-009 개정 문언 PASS; regression-class AC-AFG-008 앵커 7종 기록 완료
ac_fail_count: 0
preserve_list_post_run_count: `internal/web/assets/app.js` 무변경 · `go.mod`/`go.sum` 무변경(사본 스냅샷은 Python stdlib `hashlib`/`os` 뿐) · 제품 소스(`internal/web/*.go` 비테스트) 무변경 · `startFireGuardServer` 의 실루트 배선 의미 불변 · 정적 형제 2파일 무손상 · `.github/workflows/ci.yml` 무변경 — run-phase 변경 파일은 정확히 2개(`internal/web/appjs_fire_guard_test.go`, `internal/web/testdata/appjs_fire_probe.py`)
l44_pre_commit_fetch: 미실행 — 레인 규율상 이 워크트리는 push 하지 않으며, 커밋 직전 `git rev-parse --short HEAD` + `git branch --show-current` 재독만 수행했다(매 커밋 전 3회). 부재를 「청결」로 읽지 않기 위해 기록한다
l44_post_push_fetch: 해당 없음 — push 없음(리드 일괄)
new_warnings_or_lints_introduced: 0 (`golangci-lint run ./internal/web/...` → `0 issues.`, `go vet` 통과, `gofmt -l` 무출력)
cross_platform_build:
  darwin: `go build ./...` exit 0
  windows: 미측정 — 개정분은 Go 테스트 파일 1개와 Python testdata 1개뿐이고 syscall·build tag 를 쓰지 않으나, 재지 않은 것은 재지 않았다고 적는다. 매트릭스 판정은 push 뒤 CI 몫
total_run_phase_files: 3 — `internal/web/appjs_fire_guard_test.go`, `internal/web/testdata/appjs_fire_probe.py`, `.moai/specs/SPEC-APPJS-FIRE-GUARD-001/progress.md`(§E.2/§E.3 개정분)
run_phase_correction_after_f1 (card t1106, sync-audit-delta F2 — 추기이지 정정 덮어쓰기가 아니다):
  위 `preserve_list_post_run_count` 와 `total_run_phase_files` 두 줄은 **쓰인 시점(`19f210fe5`)에는 참이었고**, 그 뒤 sync-audit 이 낸 차단 결함 F1 을 수리하면서 거짓이 됐다. 관측 기록을 나중 사실에 맞춰 고쳐 쓰면 「무엇을 언제 봤는가」가 사라지므로, 원문을 남기고 여기에 덧붙인다.
  - `.github/workflows/ci.yml` 은 **무변경이 아니다.** 커밋 `9722f76f1` 이 `test-browser` 잡의 맨손 탐침 호출 3곳(734·745·767)에 `--primary-entries-only` 를 달았다. 사유: 이 카드가 넣은 `validation_reject_banner`(`requires_sandbox_serving: True`) 때문에 선언 없는 호출이 REQ-AFG-014 (1) 에 따라 exit 2 로 거부되어 CI 잡이 죽었다. 즉 **이 카드가 만든 파손의 수리**이며, 탐침의 거부는 설계대로 올바른 동작이라 손대지 않았다.
  - 따라서 run-phase 변경 파일은 2개가 아니라 **3개**(`appjs_fire_guard_test.go`, `appjs_fire_probe.py`, `ci.yml`)이고, `total_run_phase_files` 는 3 이 아니라 **4**다(위 3개 + 이 `progress.md`). 실측: `git diff --name-only 0fbc75afc..HEAD` → 4행.
  - 이 수정이 CI 가 보는 범위를 줄이지는 않는다 — 기계 측정: base `0fbc75afc` 의 매니페스트는 8항목 전부 무표식이고, 현재 9항목 중 표식은 이 카드 신설분 하나뿐이다. `--primary-entries-only` 가 구동하는 8항목은 **카드 이전 CI 가 보던 바로 그 집합**이다. 잃은 커버리지 0.
  - 독트린 확인: `ci-autofix-protocol.md` § CI Infrastructure Preservation 의 워크플로 수정 금지는 자기 마지막 줄이 `applies to every cycle_type=autofix invocation` 으로 범위를 선언한다. 이 수리는 리드 지시의 run-phase 수리이므로 저촉되지 않는다.

m1_to_mN_commit_strategy: 슬라이스당 1커밋 — 35a1015f9(표식·라우팅: M6.1-6.3 + M7.1-7.3) → 1302f1f76(사본 시나리오·paint·무쓰기: M6.4-6.5 + M7.4-7.5) → 19f210fe5(축소 선언·한계 주석·재측정: M6.6 + M7.6-7.8) → 본 progress 기록. 전 커밋 본문에 card t1106 명기. push·PR 없음
status_transition: 없음 — `status: in-progress` 는 plan-phase 에서 이미 설정돼 있었다(spec/plan/acceptance/progress 4파일 모두). `draft → in-progress` 는 이 run 에서 수행할 것이 남아 있지 않았고, `implemented`/`completed` 로의 전진은 manager-docs 소관이라 건드리지 않았다
unmeasured:
  - `test-browser` CI job 의 러너 실측(개정분 포함) — push 뒤 CI 몫이며 여기서 돌리지 않았다
  - windows/linux 크로스 빌드 — 위 cross_platform_build 참조
  - `moai spec audit` 의 `SyncStatusDrift` MUST-FIX — 별도 카드 소유의 알려진 오탐이며 그 처방(`--backfill-only`)은 개정을 되돌리므로 적용하지 않았다
  - 전체 스위트(`go test ./...`) — 로컬 금지(§4.1), 판정은 CI 몫
carried_debt:
  - `§C` DoD 한 줄(N-5) 미작성 — plan-audit 이 알고 넘긴 부채
  - `exit 2` 를 Then 절로 재는 AC 없음 — M7-12/13 이 보고서 출력 층에서 교차 확인했으나 AC 층의 공백은 그대로다
  - D-Δ1(§A 서사의 AC-AFG-009 REQ 매핑 미공시), N2(REQ-AFG-010 job-coverage 절반) 미해소

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-22
sync_commit_sha: 0e2377323   # sync 커밋 본 SHA — backfill 커밋으로 기입 (커밋은 자기 해시를 인용할 수 없다)
sync_status: complete
frontmatter_status_transitions:
  in-progress: 2026-09-22    # M1 커밋 70f37a688 (draft → in-progress, manager-develop)
  implemented: 2026-09-22    # 본 sync 커밋 (merged transition — 별도 Mx chore 커밋 없음)
  completed: 2026-09-22      # 본 sync 커밋 (merged transition)
changelog_entry_added: no    # 카드 sync 커밋 CHANGELOG 미작성 규율(verified-batch 선례) — grep -c 'SPEC-APPJS-FIRE-GUARD-001' CHANGELOG.md → 0 적중 전후 동일
ac_count_check: acceptance.md 고유 AC 식별자 9건(AC-AFG-001..009, grep -oE 'AC-AFG-[0-9]+' | sort -u | wc -l) = §E.3 ac_pass_count 6 blocking + 3 regression-class 합 9건
total_sync_phase_files: 2    # progress.md(§E.4) + spec.md(frontmatter status+updated) — CHANGELOG.md 미작성
canary_compliance_check: not-applicable  # 이 SPEC 은 장래 정책을 정의하지 않는다 — 브라우저 발화 가드 인프라(탐침·드라이버·돌연변이기·CI job) 납품이 전부
b12_self_test_a_pre_emission_grep: 0 hits (no-emission 경로 — 중복 방지 grep 통과, 엔트리 미작성으로 유지)
b12_self_test_b_ac_count_match: 9 == 9 (pass)
b12_self_test_c_file_path_verification: not-applicable (CHANGELOG 엔트리가 없어 경로 인용 대상도 없다)
mx_validation:
  status: no-op
  reason: run-phase 신규 파일은 Go 테스트 파일 1개 + Python testdata 스크립트 2개 + ci.yml EOF 덧붙임 — 신규 exported 함수·고 fan_in·위험 패턴 해당 0건 (@MX 스캔 3개 신규 파일 0적중; 제품 소스·assets 무변경은 §E.3 preserve_list)
sync_phase_scope_note: writable set honored exactly — progress.md §E.4 + spec.md frontmatter(status+updated)만 수정; §E.1–§E.3, spec/plan/acceptance 본문, CHANGELOG.md, internal/web/** 소스 전부 미수정
```

## §F Phase 4 Mode Selection

- Input: tier M · scope ≈6 files (probe py, driver test.go, mutation py, ci.yml job, SPEC artifacts) · domains 4 (Go test, Python, CI YAML, SPEC artifacts) · concurrency benefit LOW (coding-heavy) · agent-team prereqs: not requested
- Evaluation: direct=not selected (multi-file, semantic) · fanout=not selected (coding-heavy per Anthropic caveat) · sweep=not selected (not mechanical-uniform, no Workflow scale) · **serial=selected**
- Decision: `serial`
- Justification: coding-heavy implementation with inter-file dependencies (probe ↔ driver ↔ CI job) — sequential single-agent milestones are the safe default; the lane reports to the lead at phase boundaries, so a goal-armed autonomous loop adds coordination cost without removing waits. Kickoff approved by the operator 2026-09-22 (세션 내 직렬 선택); this log precedes the first run-phase Agent() spawn per orchestration-mode-selection.md §D.
