---
id: SPEC-WEB-BROWSER-OBS-001
title: "plan — REQ-A real-browser (CDP) observation"
version: "0.1.0"
created: 2026-09-22
author: manager-spec
tier: M
---

## §A Context

- **카드**: t1081 (Factory lane, Class C — 설계 판정 포함 측정 SPEC)
- **트리**: 이 워크트리, 브랜치 `WT-cdp-observation` @ cd99336bf (배차 base = 로컬 develop cd99336bf; t1051 의 REQ-A 구현 착지 트리)
- **SPEC 산출물**: `.moai/specs/SPEC-WEB-BROWSER-OBS-001/{spec,plan,acceptance,progress}.md`
- **판정 기록 거처**: `.moai/reports/t1081/verdict.md` (run 산출물), 프로브·캡처도 같은 디렉터(gitignored)
- **PRESERVE (변경 금지)**: `internal/web/**` 전체, `internal/template/**`, `.moai/specs/SPEC-WEB-CONSOLE-017/**`, `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1080/**`(읽기 전용), `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1041/**`(읽기 전용)
- **EXTEND**: 없음 — 측정 전용 SPEC. run 단계의 쓰기는 `.moai/reports/t1081/`(무추적)와 SPEC 디렉터의 상태 전이뿐이다.
- **본보기**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1041/.moai/reports/t1041/browser-probe.py` — python3 + websockets CDP 클라이언트(탭 개폐, `Runtime`/`Log`/`Page` 이벤트 캡처). 기구 모양을 재사용하고 표적을 REQ-A 슬롯 관측으로 바꾼다.

### §A.1 형제 카드 경계 (재확인)

t1080 (SPEC-WEB-TRANSPORT-001) 의 **종결 시점과 병합 상태는 본 SPEC 의 전제가 아니다**(REQ-BO-005 — 분할(t1080 §1.2)은 어느 쪽이 먼저 착수해도 상대의 전제를 삼키지 않게 설계됐다). 서버 측 400 특성화·임베드 htmx 계약 추출은 t1080 소관, 실브라우저 REQ-A 갭 관측은 본 SPEC 의 단독 소관이다. run 중 t1080 의 상태가 어느 값이든, 그 판정을 인용해 본 기록을 고치지 않는다.

### §A.2 사전 확인된 측정 전제 (본 레인, cd99336bf)

`save__msg save__msg--error` 슬롯(`internal/web/shell.templ`, `role="alert"`), `hx-boost="true"` 폼(`internal/web/root.templ`), `renderErrorPage` 정의(`internal/web/handlers.go`)의 존재를 grep 으로 관측했다. 콘솔 에러 수·본문 크기는 측정하지 않았다 — REQ-BO-003 의 run 대상이다.

## §B Known Issues (관련 축만)

- **B3 (서브에이전트 경계)** — run 실행자(manager-develop)는 사용자 질의 없이 blocker 보고로 닫는다.
- **B7 (CWD 경로)** — CDP 프로브의 서버 URL·디버깅 포트는 명시 인자로 전달하고, 작업 디렉터 추정으로 서버 주소를 만들지 않는다(t1041 의 `CDP_HOST` 상수 패턴 계승).
- **B8 (워킹 트리 위생)** — `.moai/reports/t1081/` 은 gitignored 로컬 증거 디렉터다. 프로브 스크립트·캡처를 추적 대상에 두지 않는다(HARD-1, REQ-BO-008).
- **B10 (범위 규율)** — `internal/web` 수정 금지. 브라우저 관측이 모순을 내도 수리 카드 라우팅이 정답이다(REQ-BO-007).
- **B11 (질의 금지)** — 실패 유도 방법 선택은 run 실행자의 설계 결정이며 사용자 질의를 요구하지 않는다(spec.md §5).

## §C Pre-flight

```bash
git -C <이 워크트리> rev-parse --short HEAD        # cd99336bf 확인
git -C <이 워크트리> status --porcelain             # clean 확인
grep -n "save__msg--error" internal/web/shell.templ # 슬롯 존재 (전제 재확인)
grep -n "hx-boost" internal/web/root.templ          # 부스트 폼 존재
grep -c "responseHandling" internal/web/assets/htmx.min.js   # 정보성 — 분석은 t1080 소관이므로 값만 참고
```

### §C.1 브라우저 기동 절차와 의존성 사전 확인 (plan-audit iter-1 D3)

t1041 의 `browser-probe.py` 는 **이미 떠 있는 Chrome 에 연결한 선례**다 — 기동 절차의 선례가 아니다. 본 run 이 브라우저 기동의 소유자다:

```bash
# 1. Chrome 바이너리 발견 — 환경변수 우선, 미설정 시 표준 경로 후보를 순서대로 존재 확인
#    (macOS: /Applications/Google Chrome.app/Contents/MacOS/Google Chrome,
#     그 밖: `command -v google-chrome` 계열). 발견 실패 시 run 실행자가 절대 경로를 명시 인자로 제공한다.
CHROME_BIN="${CHROME_PATH:-<발견 절차의 결과>}"

# 2. 기동 — 전용 디버깅 포트 + 임시 user-data-dir (다른 레인 세션과 절대 공유하지 않는다)
timeout <상한> "$CHROME_BIN" --headless=new --remote-debugging-port=<전용 포트> \
  --user-data-dir="$(mktemp -d /tmp/t1081-chrome.XXXXXX)" &

# 3. 기동 준비 프루브 — CDP HTTP 단이 준비됐는지 유계 재시도로 확인
#    curl -sf http://127.0.0.1:<전용 포트>/json/version  (성공까지 재시도, 상한 초과 시 기동 실패로 분류)

# 4. python 의존성 프루브 + 폴백
python3 -c 'import websockets' \
  || { python3 -m venv /tmp/t1081-venv && /tmp/t1081-venv/bin/pip install websockets; } \
  || echo "BLOCKER: websockets 확보 불가 — run 실행자가 blocker 보고로 처리"
```

프로브는 t1041 과 같은 `PUT /json/new` 탭 개폐 경로를 쓰며, 연결 대상은 위 2단이 기동한 세션이다.

## §D Constraints

- HARD-1..7 (spec.md §4) 전체. 핵심: `internal/web` 0변경, 브라우저 전용 판별식, 전제값과 측정값의 구분 표기, 포트 충돌 금지, 좌표는 심볌+SHA 앵커, MCP `project_root` 명시 전달.
- 포트 할당: 웹 서버·원격 디버깅 포트는 run 시작 시 `lsof -iTCP:<후보> -sTCP:LISTEN` 비었음을 확인한 뒤 택한다. t1041 이 쓴 9333 은 기본 후보에서 뺀다(HARD-5).
- 프로세스 정리: Chrome 과 `moai web` 모두 `timeout` 래퍼로 기동하고, 정리 경로(트랩 또는 명시 kill 시퀀스)가 스크립트 후반에 반드시 실행된다. 종료 후 `pgrep` 프로브로 고아 0을 확인한다(REQ-BO-008, AC-BO-006). **브라우저 기동·해체의 소유자는 본 run 이다**(§C.1) — Chrome 임시 `user-data-dir` 도 같은 정리 경로가 함께 지운다.

## §E Self-Verification

run 완료 시 manager-develop §E 양식(VCI 5단 보고)으로 보고한다. 핵심 항목:

- **E1** AC-BO-001..006 PASS/FAIL 행렬 — 각 행에 관측 명령·CDP 이벤트 출력 전문.
- **E5** 영향 패키지 테스트 — `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/web/...` exit 0 (변경 0이므로 회귀 가드 성격).
- **E6** diff 범위 — `git diff --name-only $(git merge-base develop HEAD)..HEAD` 가 `.moai/specs/SPEC-WEB-BROWSER-OBS-001/**` + 상태 전이만을 담는 것(`internal/` 0행 — §8 gitflow 규율의 merge-base 판별식 준수).
- **E8** — TDD 해당 없음(측정 SPEC). 대신 **RED-now 성격의 관측**: 판별식 프로브가 세션 준비 단계에서 의도적 오입력으로 1회 실패 관측을 남긴다(프로브가 무조건 초록을 내는 기구가 아님을 확인 — `verification-completeness.md` §1.1).

## §F Milestones (가역성 순서 — 바뀔 가능성이 큰 결정이 앞에 온다)

### M1 — 판별식 기구 설계·구현 (가장 바뀔 결정)

1. 이 워크트리에서 `moai` 바이너리 빌드(cd99336bf). 별도 설치하지 않고 빌드 산출물을 직접 실행한다.
2. 실패 유도 방법 확정 — 9 seam 중 하나를 브라우저 세션에서 실패시키는 방법(설정 파일 경로 비가록 등)과 그때 기대되는 안정 문구(t1051 §5.1 표)를 기록한다.
3. CDP 프로브 `.moai/reports/t1081/browser-probe-t1081.py` 작성 — t1041 기구 재사용 + §5 인벤토리 7항목. **스왑 vs 내비게이션 삼원 분류기(REQ-BO-002 — 메인 프레임 내비게이션 마커 + 슬롯 DOM 변이, href 는 기록 전용)가 이 마일스톤의 핵심 산출물이다.**
4. 서버·브라우저 기동/해체 절차(Chrome 바이너리 발견·기동·준비 프루브는 plan §C.1, 포트 선점 확인, `timeout`, 정리 경로, 고아 프로브) 포함.
5. 프로브 무결성 자기확인 — 의도적 오입력 1회 실패 관측(§E8).

### M2 — 세션 실행: 실패·성공 제출 관측

1. 실패 제출 세션 — §5 항목 1-6 관측, 삼원 분류 수행, 원본 캡처 JSON 저장.
2. 성공 제출 세션 — 콘솔 에러 수 관측(기대 0이지만 **기대가 아니라 측정값으로** 기록).
3. 실패 응답 본문 크기 측정(`Network.getResponseBody`).

### M3 — 부수 관측 (갭 2) + 전제 승격 기록

1. 같은 세션에서 검증 거부(빈 필드) 제출 1회 — htmx 4xx 처리 부수 기록 또는 관측 불가 기록(REQ-BO-004).
2. t1051 전제값과 측정값의 대조표 작성 — 구분 표기(REQ-BO-003).

### M4 — 판정 기록 + 결함 라우팅 (기계적 마무리)

1. `.moai/reports/t1081/verdict.md` 작성 — 삼원 분류, 판별 신호(내비게이션 마커·슬롯 변이)와 기록 전용 href의 출력 전문, 측정 대조표, 신뢰도 등급 `browser-observed`, 분할 경계 확인 한 줄, (모순 시) 결함 발견 + 후속 수리 카드 제안.
2. 고아 프로세스 0 확인, SPEC 상태 전이, 완료 보고.

## §G Anti-Patterns (반복 금지)

- **본문 문자열 탐침 판정** — `curl http://127.0.0.1:<port>/save` 본문을 grep 해 REQ-A 를 판정하는 것. 폐기된 판별식이다(HARD-2).
- **전제값 재인용** — 「실패 시 에러 2건」을 t1051 인용으로 재진술하는 것. 측정 전이라면 그것은 전제다(HARD-3).
- **렌더 존재 = 가시성** — DOM 노드 존재만으로 REQ-A 충족 판정. `role="alert"` 슬롯이 보이는지까지 관측한다(REQ-BO-001).
- **t1080 판정 대기** — 형제가 먼저 끝나기를 기다리는 것. 분할은 어느 쪽이 먼저여도 되게 설계됐다(REQ-BO-005).
- **고아 방치** — 세션 종료 후 Chrome/서버 프로세스 잔존. 정리 경로와 고아 프로브가 이를 잡는다(REQ-BO-008).

## §H Cross-References

- `SPEC-WEB-CONSOLE-017` (t1051, 감사 종료) — REQ-A 정의, 9 seam 문구 표(§5.1), 갭 선언(§6), §F 배제 계승.
- `SPEC-WEB-TRANSPORT-001` (t1080, 미병합 — 본 트리에 없음) — §1.2 분할표, 신뢰도 어휘(`contract-level` vs `browser-observed`), REQ-TR400-003 deferred 이연의 소유자.
- `.moai/reports/t1041/` — browser-probe.py 본보기 + verdict.md (관측 기록 양식 참조).
- `.claude/rules/moai/development/verification-completeness.md` §1.1 — 프로브 무결성(의도적 실패 관측) 근거.
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — 전제/측정 귀속, 5단 보고 양식.
