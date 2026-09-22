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

(비워 둔다 — manager-develop 소유. run 이 시작될 때 채워진다.)

## §E.3 Run-phase Audit-Ready Signal

(비워 둔다 — manager-develop 소유.)

## §E.4 Sync-phase Audit-Ready Signal

(비워 둔다 — manager-docs 소유. sync 커밋이 `sync_commit_sha` 를 채운다.)
