# SPEC-APPJS-FIRE-GUARD-001 — 구현 계획

> 절 순서는 **되돌리기 비용이 큰 결정부터**다. 근거가 되는 실측은 spec.md §B 다.

---

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

## §F 위험

| 위험 | 성질 | 완화 |
|---|---|---|
| 탐침 타이밍 취약성(고정 drain 대기) | 조용한 간헐 red | 효과 대기는 sleep 이 아니라 조건 폴링으로 전환 가능한 곳부터; 재시도 상한과 플레이크 기록 규율을 job 에 명시 |
| Chrome-for-Testing 버전 드리프트/다운로드 실패 | 시끄러움 — job red | 버전+해시 고정; 다운로드 실패는 exit 2(기계결함)로 분류되어 제품 결함과 갈린다 |
| 돌연변이 앵커가 app.js 진화로 미매치 | 조용한 공허 | REQ-AFG-008 전제 단언 — 즉시 실패 |
| in-process↔실바이너리 등가 붕괴 | 조용할 수 있음 | M1 이 **측정**으로 등가를 확정한 뒤 M2 로 진행 — 붕괴 시 드라이버도 실바이너리로 통일 |
| 포트 충돌(18441/9333 등) | 시끄러움 | 드라이버는 임시 포트, Chrome CDP 포트도 가용 포트 탐색; job 은 러너 단독 환경 |
| 탐침이 저장 계열 버튼을 행사해 설정을 쓰는 사고 | **데이터 손실 경로** | REQ-AFG-012 — 매니페스트 기본 제외 + 필요 시 일회용 사본 서빙 |

## §G 안티패턴 — 하지 않을 것

- 게이트 없이 테스트를 기존 suite 에 얹는 것 — 러너 이미지에 Chrome 이 있으면 기존 job 거동이 러너에 기생한다
- skip 을 사유 없이 출력하는 것 — REQ-AFG-001
- 탐침이 보고만 하고 job 이 JSON 을 육안 해석하는 것 — REQ-AFG-005 (B.5 정직 기록의 재발 방지)
- 돌연변이 좌표를 줄 번호 상수로 박는 것 — §B
- 복원 검증 없이 job 을 끝내는 것 — REQ-AFG-009
- 예시 탐침 원본(다른 워크트리)을 수정하거나 그 트리에 무언가 쓰는 것 — 읽기 전용 참조
- Playwright 등 프레임워크 도입 — spec.md §E 제외

## §H 교차 참조

- `spec.md` §B — 이 계획이 딛고 선 실측 전부 (B.1 정방향 / B.2 역방향 / B.3 인벤토리 / B.4 기반시설 / B.5 탐침 출처)
- `spec.md` §C — 채택·기각 결정 본문
- `acceptance.md` — AC 정본
- `.claude/rules/moai/development/verification-completeness.md` — §1.3 continued firing(셀렉터 스테일=red), §2 two-cell 규율
