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

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-22
sync_commit_sha: pending-backfill-sync   # 커밋은 자기 해시를 인용할 수 없다 — 다음 커밋에서 backfill
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
