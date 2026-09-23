# SPEC-APPJS-FIRE-GUARD-001 — 구현 계획

> 절 순서는 **되돌리기 비용이 큰 결정부터**다. 근거가 되는 실측은 spec.md §B 다.

---

## §A00 개정 결정 2(card t1108) — 스왑 뒤 지표는 실제 boost 스왑과 `htmx:afterSettle` 이벤트 뒤에만 (개정 2 에서 가장 되돌리기 비싼 결정)

> 이 절은 card t1108 개정 2 다. §A0(t1106)과 §A~§D(원판)는 한 글자도 바뀌지 않았다. 근거가 되는 실측은 spec.md §B.7 이다.

**문제:** REQ-AFG-007 은 「스왑 뒤 재바인딩」을 재라고 한다. 그러나 매니페스트가 행사한 것은 boost 조상이 없는 `/todo` 링크의 **전체 이동**이었고, 탐침은 URL 만 기다린 채 새 문서의 초기화보다 먼저 클릭할 수 있었다. 이 결정이 뒤집히면 탐침 5·6단계, 매니페스트 두 항목, 판정 규칙이 함께 뒤집힌다.

**데이터 모델에 미치는 변화(가장 먼저 검토돼야 할 것):**

| 항목 | 개정 전 | 개정 후 |
|---|---|---|
| 스왑 항목 id | `swap_todo_nav` | `swap_boosted_tab` (이름 변경 — 파급은 M8.6) |
| 스왑 항목 `page` | `"/"` — 실제 행사는 `/settings` 위에서 일어났다(spec.md §B.7.1) | `"/settings"` |
| 스왑 항목 `selector` | `a[href="/todo"]` (boost 조상 없음) | `/settings` 설정 폼 안의 boost 링크. 측정된 후보는 `/settings?tab=audit`(spec.md §B.7.4). 문자열은 run-phase 소관 |
| 스왑 항목 판정 | URL 이 `/todo` 가 됐는가 | REQ-AFG-016 자기확인 네 다리가 모두 참인가 |
| 스왑 뒤 항목 id | `popover_after_swap` | 유지 |
| 스왑 뒤 항목 `page` | `"/todo"` | `"/settings"` |
| 스왑 뒤 대기 | `location.pathname` 폴링 | `htmx:afterSettle` 이벤트(리스너는 클릭 전에 등록) |
| `line_group` | 스왑 `None`, 스왑 뒤 `73` | 불변 — `INVENTORY_TOTAL = 13` 불변(spec.md §B.7.6) |
| 효과 종별 | 스왑 `swap`, 스왑 뒤 `visibility` | 불변 — 탭 클릭은 GET 조회라 비영속 계열에 머문다(spec.md §B.7.5) |
| 보고서 키 | `p5_*`·`p6_*` | 불변 — 드라이버 JSON 태그(`appjs_fire_guard_test.go:70`·`:73`)가 고정한다 |

**채택:** 리드가 정한 방향 (a) 다. REQ-AFG-007 의 의도를 유지하고, 스왑 자기확인(REQ-AFG-016)을 더해 전제가 조용히 재발하지 않게 한다. 채택·기각의 근거 전문은 spec.md §C 「채택(개정 2)」이다.

**대기 설계의 요지.** 리스너는 클릭 **전에** 같은 문서에 붙인다. `app.js` 는 로드 시점에 자기 `htmx:afterSettle` 리스너를 등록하고, 탐침은 그보다 늦게 등록한다. 같은 이벤트의 리스너는 등록 순서대로 실행되므로 `initConsole` 의 재결합이 탐침의 해제보다 먼저 돈다. 이 순서는 계약이 아니라 관찰에 기댄 설계이므로 §F 에 위험으로 적는다. 대기 상한은 부재를 이름 붙은 적색으로 바꾸는 장치다. 상한 만료를 통과로 읽는 경로는 없다.

**기각:** (b) 옛 링크를 두고 `DOMContentLoaded` 뒤를 기다린다 — 스왑 뒤 재바인딩을 어떤 단계도 재지 않게 된다 / 제품에 boost 된 `/todo` 링크를 만든다 — 가드를 위해 제품을 바꾸는 역순이다 / 고정 시간 대기 — 부하에 다시 깨진다 / 자기확인 없이 링크만 바꾼다 — 재발 경로가 남는다.

## §A0 개정 결정(card t1106) — 제출을 행사하는 유일한 조건: 사본 + 무쓰기 단언 (이 개정에서 가장 되돌리기 비싼 결정)

> 이 절은 card t1106 개정분이다. §A~§D(원판 결정)는 한 글자도 바뀌지 않았다.

**문제:** 검증 거부 배너의 **화면 도달**을 재려면 폼을 제출해야 하고, 제출은 REQ-AFG-012 가 정면으로 다루는 영속화 계열이다. 이 결정이 뒤집히면 탐침의 매니페스트 스키마와 드라이버의 서버 배선이 함께 뒤집힌다.

**채택(spec.md §C 개정 결정과 동일):** 두 절반이 함께다 — (1) 그 항목만 **일회용 프로젝트 사본** 위에서 서빙하고, (2) 탐침이 제출 전후 사본의 **바이트 불변**을 스스로 단언한다. (1)이 노출을 구조적으로 없애고, (2)가 「거부는 아무것도 쓰지 않는다」는 명제를 계속 측정한다. 어느 한쪽만으로는 부족한 이유는 spec.md §C 개정 결정 표에 있다.

**REQ-AFG-012 는 약화되지 않는다.** 그 조문이 이미 이름한 두 경로 중 둘째(일회용 사본)를 처음 사용하는 것이다. 드라이버 소스 상단 주석이 같은 것을 지시문으로 적어 뒀고(`ProjectRoot` 가 그 배선점이라고 명시), 이 개정은 그 지시문을 실행한다.

**데이터 모델에 미치는 변화(가장 먼저 검토돼야 할 것):** 매니페스트 항목 스키마가 두 필드를 얻는다 — 사본 서빙 요구 표식과 무쓰기 단언 요구 표식. 효과 종별 `validation-reject` 는 두 필드 없이는 유효하지 않으며(불가분성), 그 판정은 `--lint-manifest` 와 그것을 강제하는 Go 테스트가 함께 운반한다. 필드 이름·술식 문자열은 run-phase 소관이되, **불가분성 자체는 계약**이다(REQ-AFG-014).

**인벤토리는 건드리지 않는다.** 신설 항목은 app.js 등록 지점이 아니라 htmx boost + 서버 렌더 표면이므로 기존 `swap_todo_nav` 와 같은 `line_group: None` 계열이다. `INVENTORY_TOTAL = 13` 과 lint 커버리지 산식은 불변이다(spec.md §B.6.4). 이 사실을 여기 적어 두는 이유는, 모르고 `INVENTORY_TOTAL` 을 올리면 `TestAppJsFireManifestInventoryCount` 가 즉시 적색이 되기 때문이다.

**폭발 반경 — 안 (a) 채택, 두 번째 서버 인스턴스(이 개정에서 가장 되돌리기 비싼 둘째 결정):** `startFireGuardServer` 는 서버를 **하나** 띄우고 매니페스트 전 항목이 그 하나를 쓴다. `ProjectRoot` 를 사본으로 바꾸는 것은 신설 항목에 국한되지 않고 **기존 8항목 전부의 서빙 루트를 바꾼다** — `glm_reveal` 은 시드된 자격증명 상태에, `swap_todo_nav`·`popover_after_swap` 은 `/todo` 에, `copy_button` 은 `/specs` 에 의존한다. 채택안은 **(a) 사본은 신설 제출 항목만 서빙한다**: 두 번째 서버 인스턴스를 띄우고, REQ-AFG-014 (1) 이 이미 요구하는 **사본 서빙 표식을 라우팅 키로** 쓴다(새 개념 없음). 기존 항목의 서빙 루트는 불변이고 폭발 반경은 0 이다. (b)(전 항목 사본 서빙)는 기각했다 — 착지 전에 기존 8항목의 사본 위 무회귀를 **측정**해야 하고, 더 무겁게는 원판이 의도적으로 실저장소를 서빙하던 fidelity 를 합성 사본으로 바꿔 「틀린 트리를 잰 초록」 위험을 한 항목 때문에 전면 도입한다. 기각 근거 전문은 spec.md §C 개정 결정 표.

**정리는 프레임워크가 보증한다:** 사본 수명은 `t.TempDir()` 에 묶는다 — panic·조기 실패를 포함한 모든 종료 경로에서 제거된다. `defer` 나 함수 말미의 제거문은 **채택 불가**다(도달하지 못할 수 있는 줄). 이 저장소의 「정리는 프레임워크 등록이어야 한다」 규율과 같은 근거이고, 원판 드라이버가 서버·Chrome 수명을 `t.Cleanup` 에 등록한 것(REQ-AFG-006)과 같은 형태다.

**무쓰기 단언의 비교 창(시간 경계):** 스냅샷은 **제출 직전**과 **거부 렌더가 안정된 직후**에 찍는다. 페이지 로드 일반 구간까지 감싸면 서버의 정상적인 읽기-경로 부수효과가 잡음으로 들어와 단언이 간헐 적색이 된다. 비교에서 빼는 경로가 생기면 **각각 사유를 소스에 적는다** — 사유 없는 제외는 단언을 조용히 공허하게 만든다(REQ-AFG-014).

**기각:** 무쓰기 단언만(사후 탐지 — 실저장소가 이미 망가진 뒤에 잡는다) / 사본 서빙만(명제가 가정으로 되돌아간다) / 신규 러너(카드 명시 금지, 계측기 자신의 버그가 제품 결함으로 위장한다).

## §A 결정 1 — 가드의 아키텍처: 네 조각, 두 서면 (가장 되돌리기 비싼 결정)

**문제:** 런타임 발화 가드는 어디에 살고, 언제 발화하고, 누가 red 를 만드는가. 이 결정이 뒤집히면 탐침·드라이버·job 전부를 다시 쓴다.

**채택:** spec.md §C 의 네 조각 — 커밋된 탐침(`internal/web/testdata/appjs_fire_probe.py`) / 환경 게이트된 드라이버 테스트(`internal/web/appjs_fire_guard_test.go`) / 돌연변이기(`internal/web/testdata/appjs_fire_mutation.py`) / 신설 `test-browser` job.

**왜 두 서면(로컬 드라이버 = in-process, CI 레드 = 실바이너리)으로 갈라지는가:**

- 로컬 드라이버는 `go test` 안에서 스스로 수명을 정리할 수 있어야 한다 — 테스트가 바이너리를 빌드하면 빌드 캐시·쓰기 권한·속도가 테스트 전제에 합류한다. in-process `NewServer(cfg)` + `Handler()`(server.go:122/162)가 그 답이다.
- 돌연변이 방향은 원리상 재빌드를 요구한다 — app.js 는 `//go:embed` 로 실리므로(spec.md §B.2 절차), 드라이버가 아니라 **job 스텝**이 돌연변이→재빌드→탐침→복원을 주관한다. 임베드된 사본을 메모리 변형으로 대체하는 방법은 없다.
- 두 서면이 같은 표면을 보는지는 **가정하지 않는다** — M1 이 같은 탐침을 양쪽에 돌려 지표 등가를 측정한다.

**기각:** 단일 서면 강제(테스트 안 재빌드 또는 job 이 in-process 를 흉내) — 어느 쪽이든 위의 원리 제약을 부수고 부담을 옮길 뿐이다.

## §B 결정 2 — 돌연변이 방향의 실현 장소와 전제 단언

역방향 AC 의 실현 장소는 **CI job 의 스텝 시퀀스**다:

```
build(원본) → probe → exit 0 기대
mutate → build(변이) → probe → exit 1 기대
restore → cmp byte-동일 → git status 청결 → build(원본) → probe → exit 0 기대
```

**전제 단언이 먼저다:** 돌연변이기가 727행 등록을 못 찾으면 변이가 만들어지지 않은 채 「exit 0」이 나오고 레드 단계는 조용히 공허해진다. 그래서 돌연변이기는 대상 부재를 `exit != 0` 으로 내고, job 은 돌연변이기의 exit 를 **mutate 단계 자체의 성패로** 읽는다(spec.md REQ-AFG-008). 돌연변이 좌표를 상수로 박지 않는다 — 줄 번호가 아니라 내용 패턴으로 찾는다(형제 SPEC plan §B 와 같은 이유).

**복원은 성패와 무관하게 실행된다** — 스크립트가 어떤 경로로 끝나도 restore→cmp 가 돈다(REQ-AFG-009). 프로토타입이 이 순서 그대로 `RESTORED_BYTE_IDENTICAL` 을 이미 실측했다. 이 시퀀스의 판정 형태(구체 명령·판정점·RED-now 셀 E5)는 acceptance.md AC-AFG-002 가 고정한다.

## §C 결정 3 — 탐침의 판정 계약: 3값 exit + 살아있는 매니페스트

**채택:**

- exit **0** = 전 지표 발화 + load/swap ReferenceError 0건 + 매니페스트 전 셀렉터 생존
- exit **1** = 지표 붕괴 또는 셀렉터 미달 — 실패 보고에 **무엇이** 뒤집혔는지 이름을 담는다
- exit **2** = 기계결함(CDP 미도달, 서버 기동 실패) — 결함이 가드의 잘못인지 제품의 잘못인지를 갈라놓는다

**매니페스트 규율:** (페이지, 셀렉터, 관측 효과) 삼중의 명시 목록. B.3 인벤토리 13줄 각 그룹이 항목 또는 **사유 붙은 제외**다. 0행 매니페스트는 실패다(형제 REQ-AIG-004 의 `checked > 0` 정신의 계승). 셀렉터가 아무것도 매치하지 않으면 red 다 — 매니페스트의 스테일이 조용해지지 않게 하는 조항이고, 이것이 verification-completeness §1.3(continued firing)에 대한 이 가드의 답이다.

**영속화 사고 방지(안전 절충 불가):** 탐침이 행사하는 버튼은 되돌아가는 효과(가시성·라벨·탭)만 건드린다. 저장 계열 제어는 제외하거나, 행사할 경우 드라이버가 일회용 프로젝트 사본 위에서 서빙한다(REQ-AFG-012). 기본값은 제외다 — 행사가 필요해지면 그때 사본 서빙을 켠다.

## §D 결정 4 — 파일 배치: 신규 파일 넷 + workflow 1행 블록

| 파일 | 성격 |
|---|---|
| `internal/web/testdata/appjs_fire_probe.py` | 탐침 (매니페스트 + exit 계약) |
| `internal/web/appjs_fire_guard_test.go` | 게이트된 드라이버 테스트 |
| `internal/web/testdata/appjs_fire_mutation.py` | 돌연변이기 (전제 단언 포함) |
| `.github/workflows/ci.yml` | `test-browser` job **추가** — 기존 행 1바이트 불변 |

`testdata/` 에 둔 이유: Go 도구가 testdata 를 빌드에서 무시하므로 Python 파일이 패키지 빌드에 섞이지 않는다. 기존 테스트 파일(`appjs_iife_scope_test.go`, `appjs_reinit_test.go`)은 한 줄도 건드리지 않는다 — 정적 형제와의 파일 경계가 곧 책임 경계다.

## §E 마일스톤

### M1 — 탐침 저작 + 등가 측정 (Priority High)

1. `appjs_fire_probe.py` — 매니페스트(B.3 13줄 조사 포함), 3값 exit 계약, 셀렉터 생존 검사
2. 매니페스트 효과 종별 허용목록 {`visibility`, `label`, `clipboard`, `tab`, `swap`} + `--lint-manifest` 자기검증 모드(AC-AFG-009 판정면)
3. 실바이너리 기준선 재측정 → exit 0 (spec.md §B.1 재현)
4. in-process 서면 등가 측정 — 같은 탐침, 같은 지표 (§A 의 측정 의무)
5. 영속화 제외 목록 확정 (REQ-AFG-012)
6. 탐침 상단 영어 주석 헤더 — AC-008 앵커 토큰(`orthogonal`, `SPEC-APPJS-IIFE-GUARD-001`, `manifest`, `scenario`, `Chrome`) 포함, t1041 파생 출처 명기(AC-AFG-005)

### M2 — 드라이버 테스트 (Priority High)

1. `appjs_fire_guard_test.go` — 게이트·이름 붙은 skip(영어, `error_messages: en` 정책)·in-process 서버·`t.Cleanup` 정리
2. 테스트 이름은 AC 판정 명령의 고정 앵커다: 전체 사이클 `TestAppJsHandlersFireRuntime`, 셀렉터 미달 자기검증 서브테스트 `TestAppJsHandlersFireSelectorMiss`
3. 드라이버 상단 영어 주석 — `orthogonal`·`static-scope` 앵커 포함(AC-AFG-008)
4. AC-AFG-001(정방향)/AC-AFG-003(스킵 공시) 충족

### M3 — 돌연변이기 + 양방향 로컬 재측정 (Priority High)

1. `appjs_fire_mutation.py` — 내용 패턴 매칭 + 전제 단언
2. §B 스텝 시퀀스를 로컬에서 전부 실행 — exit 0 → 1 → 복원 → 0, 증거로 기록

### M4 — CI job (Priority Medium)

1. `test-browser` job — Chrome-for-Testing 고정본 다운로드(버전+해시 고정), `websockets==<pinned>`, 게이트 설정
2. 레드 단계 스텝 시퀀스 + 복원 검증
3. 기존 job diff 0 확인 (AC-AFG-006)

### M5 — 문서화 + 전체 재측정 (Priority Medium)

1. 드라이버·탐침 상단에 §F 한계 + 직교성 주석 (REQ-AFG-013)
2. `go test ./internal/web/ -count=1` 전체 통과, 정적 형제 무손상 (AC-AFG-005)
3. progress.md §E.2/§E.3 에 출력 그대로 기록

### M6 — 탐침 확장: `validation-reject` 종별 + paint 술식 + 무쓰기 단언 (card t1106, Priority High)

1. 매니페스트 항목 스키마에 사본 서빙 요구·무쓰기 단언 요구 표식 추가, 효과 종별 `validation-reject` 를 허용목록에 **두 표식과 불가분으로** 편입
2. `--lint-manifest` 확장 — 조건이 빠진 항목을 exit 1 로 거부하고 **어느 조건이 빠졌는지** 이름으로 보고 (AC-AFG-012 (b))
3. 신설 항목(`line_group: None`, page `/settings`, 검증 실패 값 제출) — `INVENTORY_TOTAL` 은 건드리지 않는다
4. paint 술식 — 레이아웃 박스 + 조상 사슬 미은닉 + 비어 있지 않은 텍스트. 「존재만」은 금지 (AC-AFG-010)
5. 사본 루트 인자 수용 + 제출 전후 바이트 비교, 차이 시 exit 1 + 달라진 경로 이름 보고 (AC-AFG-011 (b)(c))
6. 탐침 상단 영어 주석에 개정분 한계 2축(§F 5·6) 추가 — 앵커 토큰 `paint`(§F 5)·`sandbox-root`(§F 6) 를 그대로 실는다(AC-AFG-008). §F 7(두 서면 등가 미주장)은 드라이버 쪽 사실이므로 M7.7 이 진다

### M7 — 드라이버 확장: 두 번째 서버 인스턴스 + 양방향 관측 (card t1106, Priority High)

1. **두 번째 서버 인스턴스** — 사본을 `ProjectRoot` 로 삼는 서버를 별도로 띄우고, 사본 서빙 표식을 가진 항목만 그쪽으로 라우팅한다. 기존 `startFireGuardServer` 의 `ProjectRoot: findRepoRoot(t)` 는 **건드리지 않는다**(안 (a), AC-AFG-013)
2. 사본 provisioning 은 `t.TempDir()` 파생 경로 + 서빙에 필요한 최소 구성만 복사, 수명은 프레임워크 등록으로만 보증(`defer`/말미 제거문 금지, AC-AFG-011 (d))
3. `TestAppJsFireSandboxRouting` — 라우팅 **양방향** 판정: 표식 없는 항목이 사본으로 가도 적색, 표식 있는 항목이 실루트로 가도 적색 (AC-AFG-013)
4. `TestAppJsFireValidationRejectPaints` — 게이트된 전체 사이클, 사본 서버 위, AC-AFG-010 판정
5. `TestAppJsFireValidationRejectNoWrites` — 무쓰기 단언 + **합성 쓰기 1바이트에서 exit 1 을 관측**하는 반대 방향 (AC-AFG-011 (b)(c))
6. `TestAppJsFireSandboxPairing` — 게이트 없이 항상 도는 불가분성 강제 테스트(조건 빠진 합성 항목 거부 관측 포함)
7. 드라이버 상단 주석의 REQ-AFG-012 「wiring point」 문단을 **실행된 지시문**으로 갱신 — 무엇이 사본 서버 위에서 돌고 무엇이 실루트 서버 위에서 도는지 명시. 같은 주석에 **§F 7 한계 축(두 서버 서면의 등가를 주장하지 않는다)과 앵커 토큰 `two-surfaces`** 를 싣는다(AC-AFG-008; card t1106 iter-2, D3 수리 — 종전에는 §F 7 이 어느 마일스톤의 주석 작업에도 배정되지 않았다)
8. `MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsHandlersFireRuntime'` 재측정 — 판정 집합은 **실루트 계열 8항목**이고 서빙 루트는 `findRepoRoot` 로 불변이다. 제출 항목은 이 사이클에 **들어오지 않는다** — 그 판정은 M7.4(`TestAppJsFireValidationRejectPaints`, AC-AFG-010)가 별도 실행으로 진다. 두 계열을 가르는 기제는 REQ-AFG-014 (1) 의 **적극적 축소 선언**이며, 이 실행은 그 선언을 달고 돈다(사본 base 의 부재로 대신하지 않는다 — 선언 없는 부재는 exit 2). 이 실행의 보고서는 운전 8항목과 **선언으로 제외된 표식 항목의 이름**을 함께 담아야 한다. 판정: AC-AFG-001 (a) + AC-AFG-013 (c) + `go test ./internal/web/ -count=1` 무회귀 (card t1106 iter-2, D1 수리 — 종전 문언은 AC-AFG-001 (a) 와 같은 집합을 이름하지 않았다)

### M8 — 탐침: 실제 boost 스왑 + afterSettle 대기 + 스왑 자기확인 (card t1108, Priority High)

1. 5단계를 `/settings` 위 boost 링크 클릭으로 바꾼다. 매니페스트 스왑 항목의 `page`·`selector`·`check` 를 새 정의로 바꾼다(§A00 표)
2. 클릭 **전에** 같은 문서에 `htmx:afterSwap`·`htmx:afterSettle` 리스너와 문서 표지를 심고, 옛 트리거 노드에 표지를 단다(REQ-AFG-016 (b)(d))
3. 스왑 뒤 대기를 `htmx:afterSettle` 이벤트로 바꾼다. `location.pathname` 폴링과 고정 대기 증량을 없앤다. 상한 만료는 스왑 뒤 항목을 이름으로 지목하는 실패다(REQ-AFG-007 (3))
4. 자기확인 네 다리를 보고서에 다리별로 싣는다. 한 다리라도 거짓이면 exit 1 로 실패하고, 스왑 항목과 거짓 다리를 사유로 보고하며, 스왑 뒤 지표를 발화로 판정하지 않는다(REQ-AFG-016)
5. 6단계의 `page` 를 `/settings` 로 바꾸고, 주석(`:543-544`)을 실제 대기와 맞춘다
6. **id 변경 `swap_todo_nav` → `swap_boosted_tab` 의 파급 — 전수 열거**(`git grep -n -e 'swap_todo_nav' -e 'popover_after_swap' -e 'p5_swap' -- . ':!.moai/reports'` 로 HEAD `52a486635` 에서 셌다):
   - 바꾸는 곳 — `internal/web/testdata/appjs_fire_probe.py`: `:172`(매니페스트 id), `:198`(`validation_reject_banner` 주석이 옛 id 를 비유로 쓴다), `:531`(5단계 주석), `:768`(`by_id` 키 — 판정식도 `== "/todo"` 에서 자기확인으로 바뀐다), `:788`(`selector_found` 키)
   - 바뀌지 않는 곳, 참조 0건 — 드라이버 `internal/web/appjs_fire_guard_test.go`(참조하는 것은 보고서 키 `p5_swap_referenceerrors`·`p6_popover_after_swap_fired` 뿐, `:70`·`:73`·`:474`), `.github/workflows/ci.yml`(레드 단계는 `stampRefreshed` 만 grep 한다), `--lint-manifest` 규칙(`post_swap` 필드와 효과 종별로 판정하고 id 를 보지 않는다, `:341-342`), SelectorMiss 픽스처(`[data-copy]` → `copy_button` 만 변조, `appjs_fire_guard_test.go:493-512`)
   - 이력 문서, 원문 유지 — spec.md §B.6.4·§C(t1106) 개정 결정, plan.md §A0, acceptance.md AC-AFG-013 근거 문단(주석을 달았다), progress.md §E.2(run-phase 증거 — manager-spec 수정 금지), `.moai/reports/**`
7. `popover_after_swap` 의 「selector matched nothing」 판정(`:789`, 패널 기준)은 **바꾸지 않는다** — spec.md §E 관찰

### M9 — 드라이버: 스로틀 반복 + 돌연변이 판 + 자기확인 양방향 (card t1108, Priority High)

1. `TestAppJsFirePostSwapSettleWait` — 게이트된 in-process 실루트 표면에서 탭 한정 CPU 스로틀 12배를 걸고, 정상 판과 돌연변이 판(대기를 URL 폴링으로 되돌리거나 제거한 일회용 사본)을 각각 10회 돌린다. 스로틀만으로 돌연변이 판이 10/10 적색이 아니면, 두 판에 똑같이 settle 지연을 넓히고 그 효과(afterSwap→afterSettle 간격)를 관측해 보고서에 싣는다(AC-AFG-014)
2. `TestAppJsFireSwapPremise` — 커밋된 탐침의 자기확인 네 다리 참(exit 0)과, 스왑 선택자를 `a[href="/todo"]` 로 바꾼 일회용 사본의 exit 1 + 지목된 다리를 둘 다 관측한다. 사본 방식은 `TestAppJsHandlersFireSelectorMiss` 와 같다(AC-AFG-015)
3. `TestAppJsHandlersFireRuntime` 의 (c) 단언에 자기확인 네 다리를 묶는다(AC-AFG-001 (c) 개정 문언)
4. 스로틀·돌연변이 사본·Chrome 탭의 수명은 `t.Cleanup`·`t.TempDir()` 에만 맡긴다(REQ-AFG-006)
5. 두 테스트 이름은 `AppJs.*Fire` 에 걸리게 짓는다 — M10 의 CI 선택자가 이 이름으로 고른다
6. AC-AFG-002 의 0→1→0 레드 사이클을 개정 2 매니페스트로 **다시 측정**하고, 돌연변이 아래 어느 지표가 무너지는지를 옛 관측과 대조해 기록한다

### M10 — CI 선택자 확장 + 병합 트리 측정 (card t1108, Priority Medium)

1. `.github/workflows/ci.yml:672` 의 `-run 'AppJsHandlersFire'` 를 `-run 'AppJs.*Fire'` 로 바꾼다. **이 한 곳만** 바꾼다. `--primary-entries-only` 3곳(734/745/767)과 `-timeout 10m` 은 그대로 둔다
2. 병합 트리에서 AC-AFG-016 의 세 명령을 실행한다. 선택된 테스트 전부 `--- PASS`, `--- SKIP` 0건
3. AC-AFG-006 재측정 — base `3e35fbacf` 대비 삭제 0, 기존 job 8개 목록·순서 불변
4. 리드의 일괄 push 뒤 `test-browser` 로그에서 같은 테스트 이름을 읽어 기록한다(레인은 push 하지 않는다)

## §F 위험

| 위험 | 성질 | 완화 |
|---|---|---|
| 탐침 타이밍 취약성(고정 drain 대기) | 조용한 간헐 red | 효과 대기는 sleep 이 아니라 조건 폴링으로 전환 가능한 곳부터; 재시도 상한과 플레이크 기록 규율을 job 에 명시 |
| Chrome-for-Testing 버전 드리프트/다운로드 실패 | 시끄러움 — job red | 버전+해시 고정; 다운로드 실패는 exit 2(기계결함)로 분류되어 제품 결함과 갈린다 |
| 돌연변이 앵커가 app.js 진화로 미매치 | 조용한 공허 | REQ-AFG-008 전제 단언 — 즉시 실패 |
| in-process↔실바이너리 등가 붕괴 | 조용할 수 있음 | M1 이 **측정**으로 등가를 확정한 뒤 M2 로 진행 — 붕괴 시 드라이버도 실바이너리로 통일 |
| 포트 충돌(18441/9333 등) | 시끄러움 | 드라이버는 임시 포트, Chrome CDP 포트도 가용 포트 탐색; job 은 러너 단독 환경 |
| 탐침이 저장 계열 버튼을 행사해 설정을 쓰는 사고 | **데이터 손실 경로** | REQ-AFG-012 — 매니페스트 기본 제외 + 필요 시 일회용 사본 서빙 |
| 사본 서빙 배선을 빠뜨린 채 `validation-reject` 항목이 도는 것 | **데이터 손실 경로** | REQ-AFG-014 불가분성 — `--lint-manifest` + `TestAppJsFireSandboxPairing` 이 기계로 거부 (AC-AFG-012) |
| 사본이 신설 항목을 넘어 기존 항목까지 서빙하게 되는 것 | **조용함 — 틀린 트리를 잰 초록** | 안 (a) 는 계약이다(§A0). `TestAppJsFireSandboxRouting` 이 라우팅을 양방향으로 판정하고, AC-AFG-013 (c) 가 기존 전체 사이클 무회귀를 함께 건다 |
| 사본 수명을 `defer`/말미 제거문에 맡기는 것 | 조용한 누수 | `t.TempDir()` 파생만 허용(REQ-AFG-014 (3)); AC-AFG-011 (d) 가 경로 파생과 제거문 부재를 함께 확인 |
| paint 술식이 「노드 존재」로 약화되는 것 | 조용한 공허 | AC-AFG-010 의 돌연변이 프로브 — 배너를 `hidden` 으로 렌더하는 합성 변이가 통과하면 채택 불가 |
| 무쓰기 비교 창이 너무 넓어 읽기-경로 부수효과를 잡는 것 | 간헐 적색 | 스냅샷을 제출 직전·거부 렌더 직후로 한정(§A0); 제외 경로는 사유와 함께 열거 |
| 사본 구성이 부족해 `/settings` 가 거부 경로에 도달하지 못하는 것 | 시끄러움 — red | M7 이 최소 구성 집합을 측정으로 확정; 도달 실패는 exit 2(기계결함)로 분류돼 제품 결함과 갈린다 |
| 실제 스왑 경로에서 URL 폴링판이 스로틀만으로는 실패하지 않는 것 (card t1108) | 조용함 — 대기의 필요성을 보이지 못한 초록 | AC-AFG-014 가 settle 지연 확대를 두 판에 똑같이 허용하고, 그 효과를 관측하게 한다. 그래도 적색이 아니면 blocker |
| 스왑이 프로필 트리거를 교체하지 않아 옛 핸들러가 살아남는 것 (card t1108) | 조용함 — 재바인딩이 아니라 옛 결합을 잰 초록 | REQ-AFG-016 (d). 정상 판에서 (d) 가 거짓이면 다리를 빼지 않고 blocker 로 보고한다 |
| `htmx:afterSettle` 리스너 실행 순서 — 탐침의 해제가 `initConsole` 의 재결합보다 먼저 도는 것 (card t1108) | 간헐 적색 | 탐침 리스너는 `app.js` 보다 늦게 등록되므로 등록 순서상 뒤에 돈다(§A00). 이 순서가 깨지면 AC-AFG-014 정상 판이 적색으로 드러낸다 |
| `p5_swap_referenceerrors` 창의 의미 변화 (card t1108) | 조용함 — 창 이름은 그대로인데 재는 것이 달라진다 | 개정 전 이 창은 전체 이동 중, 즉 새 문서가 `app.js` 를 다시 실행하며 던진 **로드 시점** 예외를 모았다. 개정 후에는 실제 스왑 중에 난 예외를 모은다. boost 스왑은 `defer` 로 로드된 `app.js` 를 다시 실행하지 않으므로, 로드 시점 예외는 이 창에 다시 찍히지 않는다(추론 — 미관측). 로드 시점 예외는 여전히 `p1_load_referenceerrors`·`p7_load_referenceerrors` 가 잡는다. 레드 단계의 `stampRefreshed` 가 어느 창에 찍히는지는 M9.6 의 재측정으로 관측한다 |
| 선택 집합이 커져 CI `-timeout 10m` 을 넘는 것 (card t1108) | 시끄러움 — job red | AC-AFG-016 이 로컬 병합 트리에서 같은 상한으로 재서 미리 드러낸다. 상한 조정은 개정 2 범위 밖이므로 넘으면 리드에게 보고한다 |
| 스왑 링크 후보(`?tab=audit`)가 탭 구성 변경으로 사라지는 것 (card t1108) | 시끄러움 — red | 셀렉터 생존 검사(REQ-AFG-004)와 자기확인 (a) 가 이름 붙은 적색으로 드러낸다 |

## §G 안티패턴 — 하지 않을 것

- 게이트 없이 테스트를 기존 suite 에 얹는 것 — 러너 이미지에 Chrome 이 있으면 기존 job 거동이 러너에 기생한다
- skip 을 사유 없이 출력하는 것 — REQ-AFG-001
- 탐침이 보고만 하고 job 이 JSON 을 육안 해석하는 것 — REQ-AFG-005 (B.5 정직 기록의 재발 방지)
- 돌연변이 좌표를 줄 번호 상수로 박는 것 — §B
- 복원 검증 없이 job 을 끝내는 것 — REQ-AFG-009
- 예시 탐침 원본(다른 워크트리)을 수정하거나 그 트리에 무언가 쓰는 것 — 읽기 전용 참조
- Playwright 등 프레임워크 도입 — spec.md §E 제외
- 제출 항목을 실저장소 루트(`findRepoRoot`) 위에서 행사하는 것 — REQ-AFG-014 (1)
- 무쓰기 단언 없이 사본만 쓰는 것 / 사본 없이 무쓰기 단언만 하는 것 / 수명을 프레임워크에 묶지 않는 것 — 세 조건은 분리 불가(§A0, REQ-AFG-014)
- `startFireGuardServer` 의 `ProjectRoot` 를 사본으로 바꿔 전 항목을 옮기는 것 — 안 (b) 는 기각됐다(§A0)
- 사본 수명을 `defer os.RemoveAll(...)` 이나 함수 말미 제거문에 맡기는 것 — 도달하지 못할 수 있는 줄이다
- 배너 노드의 **존재**만 확인하고 통과시키는 것 — REQ-AFG-015
- `INVENTORY_TOTAL` 을 올려 신설 항목을 맞추려는 것 — 신설 항목은 app.js 등록 지점이 아니다(spec.md §B.6.4)
- 신규 러너·별도 브라우저 하네스를 만드는 것 — card t1106 명시 금지
- 무쓰기 비교에서 경로를 사유 없이 빼는 것 — 단언이 조용히 공허해진다
- 스왑 뒤 대기를 URL 변경 폴링이나 고정 시간으로 두는 것 — REQ-AFG-007 (2) (card t1108)
- `htmx:afterSettle` 리스너를 클릭 **뒤에** 붙이는 것 — 이벤트를 놓치고, 그 누락이 시간 초과로 위장한다
- 스왑 자기확인이 적색일 때 스왑 뒤 지표를 발화로 세는 것 — REQ-AFG-016
- 정상 판에서 자기확인 (d) 가 거짓으로 나왔다고 다리를 빼는 것 — 재바인딩 전제가 이 항목으로는 성립하지 않는다는 신호이므로 blocker 다
- CI 를 넓히면서 `--primary-entries-only` 를 지우는 것 — t1106 F1 이 되살아난다
- 보고서 키 `p5_*`·`p6_*` 이름을 바꾸는 것 — 드라이버 JSON 태그가 깨진다
- 가드를 통과시키려고 제품에 boost 된 `/todo` 링크를 만드는 것 — 순서가 뒤집혔다

## §H 교차 참조

- `spec.md` §B — 이 계획이 딛고 선 실측 전부 (B.1 정방향 / B.2 역방향 / B.3 인벤토리 / B.4 기반시설 / B.5 탐침 출처)
- `spec.md` §C — 채택·기각 결정 본문
- `acceptance.md` — AC 정본
- `.claude/rules/moai/development/verification-completeness.md` — §1.3 continued firing(셀렉터 스테일=red), §2 two-cell 규율
- `spec.md` §B.6 — 개정분(card t1106)이 딛고 선 실측(t1105 수리의 검증면, 오늘의 허용목록·드라이버 배선, 인벤토리 불변, 세 빈 스윕)
- `spec.md` §C 개정 결정 — 두 절반이 함께여야 하는 이유와 기각안
- `internal/web/handlers.go` — `handleSave` 검증 거부 블록(card t1105 수리 지점)
- `spec.md` §B.7 — 개정 2(card t1108)가 딛고 선 실측(전체 이동, 초기화 전 클릭, boost 링크 인벤토리, 드라이버 표면의 실제 스왑, CI 선택자)
- `spec.md` §C 「채택(개정 2)」 — 방향 (a) 와 자기확인 채택의 근거
- `.moai/reports/t1108/verdict.md` — 원인 확정 판정서(좌표는 t1106 병합 전 트리 기준)
