# SPEC-CODEX-REVIEW-ASYNC-001 — progress

## Decision Log

모든 판정은 Jev `jev-1.13.0`(운영자 위임, 표시 전용 신호)·리더 판정이며 권위 인용이 아니다. 같은 카드 t1422 의 형제 SPEC `SPEC-CODEX-REVIEW-OWNERSHIP-001` 의 Decision Log 와 별개로 이 SPEC 의 결정만 적는다.

| # | Question | Resolution | Confidence | Disposition |
|---|----------|-----------|------------|-------------|
| A-1 | 중복 실행 방지 방식 | 트리별 락, 보유 중이면 건너뜀, 죽은 보유자의 락 인수 | 0.83 | REQ-CRA-004. gate-run 락 원시 기능(flock) 일반화 재사용(plan.md §B.3) |
| A-2 | 오래된 결과 처리 | `label_stale` — 표지를 붙여 전달, 조용히 버리지도 표지 없이 전달하지도 않음 | 0.99 | REQ-CRA-006 (오래됨의 정의를 HEAD+digest 로 넓힌 것은 O-C = A-7 에서 리더 수용) |
| A-3 | 구조 | 형제 SPEC 분리 (asyncRewake 전환은 이 SPEC) | 0.93 | 형제 SPEC 이 먼저 착지(`depends_on`) |
| A-4 | 신호 계약 | 실패 = 종료 코드 2 + stderr 요약(카드·reviewed HEAD·findings), 통과/오류/inconclusive/리뷰어 부재 = 조용한 종료 코드 0 (운영자 지시, 리더 경유) | — | REQ-CRA-001~003. SPEC-MOAI-MCP-SERVER-001 REQ-MCP-012 fail-open 유지 |
| A-5 | O-A 연속 실패 상한 | Jev: `cap3_silent`(0.21, 약함) → **오케스트레이터가 `cap3_notify_once_at_cap` 으로 덮어씀**(사유: 상한 뒤 완전 침묵은 "통과"로 오독). 트리당 연속 실패 깨움 3회, 세 번째 깨움이 마지막 알림(`이후 알림 없음, 상태는 미해결` 고정 문구 + 영어 FINAL NOTICE 줄), 이후 **모든** 실패는 pass 까지 침묵(plan-audit 1회차 D1 이후 "같은 지문" 조건은 폐기 — A-15) | 0.21 → override | PROVISIONAL. 리더 수용(2026-10-02). REQ-CRA-014·015, AC-014·015. 읽기 확인 요청: plan.md §G "확인이 필요한 읽기", decision-index Q19 |
| A-6 | O-B 타임아웃 마진 | OPEN — E-2 프로브 전까지 미정, 60초는 제안값 | — | REQ-CRA-008, AC-008(값 확정 전 통과 기록 불가) |
| A-7 | O-C 오래됨 정의 | `full_state_key` — HEAD + 작업 트리 digest 전체, 표지는 둘 다 명시 | 0.56 | 리더 수용. REQ-CRA-006 (운영자 문구 "HEAD" 의 확장 수용) |
| A-8 | O-D 크기 | `accept_size` | 0.98 | plan.md §H |
| A-9 | O-E 영수증 공유 | `do_not_share` — 재전달 기록을 Codex 검증 영수증과 공유하지 않음 | 0.02 (약함) | PROVISIONAL, 리더 수용. REQ-CRA-009, AC-009 |
| A-10 | O-F 프로브 실행 주체 | 레인이 헤드리스 프로브를 먼저, 대화형만 가능한 프로브는 리더가 리더 세션에서 | 0.60 | 절차 `.moai/reports/t1422/async-probes.md`, 아래 프로브 기록 표 |
| A-11 | O-G 카드 칸 | 워크트리 디렉터리 이름, 카드 워크트리가 아니면 빈 칸 | 1.00 | REQ-CRA-001, AC-001 |
| A-12 | O-H 알려진 상태 건너뜀 | `skip_known_state_for_pass` — pass 에도 적용(현행: 같은 상태도 Stop 마다 재리뷰 → 변경) | 0.41 | PROVISIONAL, 리더 수용. 동작 변경은 **SPEC 본문에만** 기술하며 CHANGELOG 항목이 아니다. 상태가 바뀐 Stop 이 리뷰를 유지하는 보존 AC 가 AC-005 에 있다 |
| A-13 | O-I 스코프 로그 행 | stderr 유지, E-4(`E4-d2-exit0stderr`) 결과 대기 | 0.97 | AC-002 괄호 |
| A-14 | O-J Tier | M 유지 — 코드+테스트 ≈14, 미러·문서 ≈5 | 0.70 | plan.md §A |
| A-15 | plan-audit 1회차 D1 — 상한 계수 규칙 (decision-index Q19) | 지문 폐기: 트리의 연속 실패 깨움 횟수만 세고 pass 하나로 해제. REQ-CRA-005 를 셋으로 가름(005 알려진 상태·014 상한·015 마지막 알림) | 작성자 선택 — **확인 필요** | 오케스트레이터 문구를 좁혀 읽음(plan §G) |
| A-16 | D2 — 전달 확인 (Q20) | `session_id`+`delivered`(stderr 쓰기 성공 뒤)+전달 창 30분 | 작성자 선택 — **확인 필요** | REQ-CRA-005, AC-005 (e)(f) |
| A-17 | D3·D4 — 래퍼 판정 (Q22) | 종료 코드 2 **그리고** sentinel 줄 `codex review gate: FAIL`; 래퍼가 stderr 를 캡처해 sentinel 줄부터만 전달, 비판정 2 는 exit 0+로그 행 | 작성자 선택 — **확인 필요** | REQ-CRA-001·003, AC-001·003 |
| A-18 | D5·D6 — 상태 키 다섯 필드, 파일은 git 디렉터리 (Q21) | `Head`·`TreeDigest`·`ConfigDigest`·`Command`·`ToolVersion`; 락·기록·로그를 `<git 디렉터리>/moai/` 에 둠 | 작성자 선택 — **확인 필요** | REQ-CRA-004·005·009·013, AC-004·009 |
| A-19 | D7·D11 — 보유 시간·크기 숫자 (Q23) | 상태 계산 예산 60s×2(보유 상한 900+120s), findings 10·줄 300자·summary 1500자·전체 8000바이트, 전달 창 30분 | 근거 없는 시작값 | REQ-CRA-001·004, AC-001·004. 훅 출력 크기 상한은 미확인 Gap |
| A-20 | D12 — 프로브 밀폐 | 키트 `run-clean.sh`(MOAI_* 등 환경 제거)+`claude --setting-sources project` | — | async-probes.md §1, plan §D |

### 확인된 사실과 미관측의 구분 (이 레인의 기록)

- **확인(공식 훅 레퍼런스, 이 레인이 직접 읽음):** `async`/`asyncRewake`/`timeout` 필드 행(spec.md §A.2), `Stop` 종료 코드 2 = "Prevents Claude from stopping, continues the conversation".
- **미관측 → EVIDENCE-NEEDED:** E-1(idle 세션 깨움), E-2(`asyncRewake` 에서 `timeout`), E-3(동시성·새 턴·세션 종료), E-4(JSON `decision: block` 무시 여부, 종료 코드 0 stderr 가시성), E-5(실행 중 설정 편집 반영). "Run hooks in the background" 절은 이 레인이 받은 내용에서 잘려 있었다. spec.md §A.3.
- **미관측 → Gap:** Claude Code 설치 버전 2.1.287(코디네이터 제공, 직접 미확인), `asyncRewake` 를 모르는 버전의 동작, 슬롯 임대의 훅 경로 적합성(검토 안 함), 락 N 중 실행의 실제 발생(코드 판독만), t1425 수치(리더 제공).

## 프로브 기록 (M1 에서 채움 — AC-011)

절차: `.moai/reports/t1422/async-probes.md`(로컬 gitignored). 키트: `.moai/reports/t1422/async-probe/kit/`. 기록 파일: `.moai/reports/t1422/async-probe/<프로브 ID>.txt`. 모든 칸은 비어 있다 — 아직 실행된 프로브가 없다.

| 프로브 ID | E 항목 | 실행 주체 | 결과 토큰 | 기록 파일 sha256 |
|---|---|---|---|---|
| `E1-headless` | E-1 (보조) | 레인 헤드리스, 먼저 | _<pending M1>_ | _<pending>_ |
| `E1-interactive` | E-1 (결정) | **리더만** | _<pending M1>_ | _<pending>_ |
| `E2-timeout` | E-2 | 레인 헤드리스, 먼저 | _<pending M1>_ | _<pending>_ |
| `E3-c1-concurrent` | E-3 c1 | 레인 스트림 입력 헤드리스, 먼저 | _<pending M1>_ | _<pending>_ |
| `E3-c2-newturn` | E-3 c2 | **리더만** | _<pending M1>_ | _<pending>_ |
| `E3-c3-sessionexit` | E-3 c3 | **리더만** | _<pending M1>_ | _<pending>_ |
| `E4-d1-json` | E-4 d1 | 레인 헤드리스, 먼저 | _<pending M1>_ | _<pending>_ |
| `E4-d2-exit0stderr` | E-4 d2 | 레인 헤드리스, 먼저 | _<pending M1>_ | _<pending>_ |
| `E5-reload` | E-5 | **리더만** | _<pending M1>_ | _<pending>_ |

## plan-audit 1회차 반응 (2026-10-02)

보고서: `.moai/reports/t1422/async-plan-audit.md`(로컬, FAIL 0.74, 감사 트리 `a0d801409`, 필수 수정 D1-D4·권고 D5-D12·노트 D13-D21, 26개 변이 중 11개 생존). 이 개정이 반영한 것: D1(지문 폐기·계수 규칙 통일: REQ-CRA-014·015) · D2(`session_id`+`delivered`: REQ-CRA-005, AC-005 (e)(f)) · D3(sentinel 줄: REQ-CRA-003, AC-003 (b)) · D4(운영 stderr 스트림: REQ-CRA-001, AC-001 (b)) · D5(다섯 필드 키) · D6(파일을 git 디렉터리에: REQ-CRA-004·009·013, AC-004 (h)·AC-009 (c)) · D7(보유 시간 상한, AC-004 (g)) · D8(AC 별 GREEN 경로 열, L6·L8·L9·L11·L12 를 뒤집히는 파일 집합 스윕·`return emitHookOutput(` 개수로 재조준) · D9(생존 변이 전부) · D10(REQ-CRA-005 분리, 15 REQ) · D11(크기 숫자) · D12(밀폐 프로브) · D13(unlisted 파일 두 개를 plan §F.2 에 추가) · D14(depends_on 사전 점검이 M1 도 막는다는 사실을 plan §A 에 기록) · D15(래퍼가 읽는 `workflow.yaml` 서술을 시각 표기된 관측으로 고침, plan §E) · D16(오케스트레이터 제공 문장을 spec §A.2 에 추가) · D17(정렬·doctor 문구를 AC-007 에, `moai update` 전달은 Gap) · D18(리뷰어 자식 생존 Gap, plan §G 위험 11) · D19(plan §D 와 acceptance §E 의 Gap 수용 예외 문장 일치) · D20(비용 공시, plan §G 위험 4) · D21(명시 실행 경로 락 없음 — 이미 §E 범위 밖, 위험 5).

**검증 Gap(감사 보고서 인용).** 감사 도구 환경에서 형식 검증 verb 세 개(REQ↔AC 추적 awk, CN-4 순서 awk, D8 awk)가 worktree 격리 가드에 거부되어 grep 기반 손 추적으로 대체됐다("awk with -f / a program … cannot be shown not to be git") — 감사의 `COLLECTED:`/`CONFLICT:` 줄이 없다. 이 레인은 같은 verb 를 실행하지 않았고 헤딩 개수(`grep -c '^### REQ-'`·`'^### AC-'`)와 `moai spec lint` 로만 점검했다.

## 변이 점검 결과 (이 개정의 자체 스윕 — 작성 시점 판독)

acceptance.md §C 표가 15개 AC 모두에 대해 요구를 어기면서 기준을 만족하는 가장 값싼 변이와 그것을 죽이는 단정 칸을 적는다. 요약: AC-001 8 · AC-002 4 · AC-003 7 · AC-004 9 · AC-005 7 · AC-006 5 · AC-007 5 · AC-008 2 · AC-009 3 · AC-010 3 · AC-011 3 · AC-012 2 · AC-013 4 · AC-014 7 · AC-015 5 = **74 개 변이 행**(감사 보고서의 26개에서 확장). **죽임 71 / 열림 3**: 열림은 AC-008 의 "마진 0"(O-B 가 정해질 때까지 통과로 기록할 수 없음), AC-011 순서 항("해시 기록 커밋이 등록 커밋보다 뒤")·AC-012 순서 항("등록 뒤 run-phase 커밋") — 커밋 순서는 구현 뒤에만 관측되는 사후 조건이라 회귀 가드로 분류한다. 감사 보고서의 생존 변이 11개(#4,#5,#9,#10,#11,#12,#14,#15,#18,#20,#21,#23)는 각각 AC-004 (b)(d), AC-014 (b)(c)(d), AC-005 (e), AC-005 (d), AC-006 (d), AC-006 (f), AC-003 (b), AC-003 등록 명령 실행·(f), AC-001 (b), AC-007 (a) 로 죽는다. "죽임"은 구현이 없는 plan 단계의 판독이며 실행 관측이 아니다 — M1 선택 항목이 실행 관측으로 격상한다(acceptance §E).

## 인계 항목

1. **프로브 실행 주체(O-F 확정).** 레인이 헤드리스 변형(`E1-headless`·`E2-timeout`·`E3-c1-concurrent`·`E4-d1-json`·`E4-d2-exit0stderr`)을 먼저 돌리고, 대화형만 가능한 `E1-interactive`·`E3-c2-newturn`·`E3-c3-sessionexit`·`E5-reload` 는 리더가 리더 세션에서 `.moai/reports/t1422/async-probes.md` 절차대로 돌린다. 헤드리스 결과가 대화형과 같다는 보장은 없으므로 `E1-headless` 는 E-1 을 결정하지 못한다. 스트림 입력 헤드리스 구동 스크립트(`stream-session.sh`)의 플래그·입력 형식은 아직 실행해 보지 않은 틀이다.
2. **형제 SPEC 착지.** 이 SPEC 의 M2 이후는 형제 SPEC 이 착지한 트리에서 시작한다(`depends_on`, 핸들러 삽입 위치 확정을 위해 소스를 다시 읽는다).
3. **결정 현황.** O-A·O-C~O-J 는 확정(A-5, A-7~A-14), O-B 는 E-2 결과 전까지 OPEN, O-I 는 E-4 결과 대기. 상한 동작(O-A)의 읽기 확인 요청 한 건이 plan.md §G 에 있다.

## §E.1 Plan-phase Audit-Ready Signal

- 산출: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` · `decision-index.md` (Tier M 산출 집합 + progress + decision gate `interview.decision_gate: on`).
- Tier: **M**. 개수 규칙: 요구 = `### REQ-` 제목 수 15, 수락 기준 = `### AC-` 제목 수 15(하위 ID 없음) — 상한 16 이하(여유 1). 파일: 코드+테스트 ≈14(하한), 미러·문서 ≈5, 합 ≈19(Tier L 대역에 걸림 → O-J 에서 M 유지로 확정, 신뢰도 0.70). 규모는 초안 ≈220 LOC 에서 Go ≈260 + 셸 ≈35 로 늘었다(plan §H) — 리더에게 재통지.
- plan_status: **audit-ready (개정 2판)** — plan_complete_at: 2026-10-02. plan-audit 1회차 = FAIL 0.74(보고서 `.moai/reports/t1422/async-plan-audit.md`, 감사 트리 `a0d801409`). 이 개정이 D1-D21 을 반영했고 **2회차 감사는 아직 수행되지 않았다**(작성 레인의 준비 신호이지 감사 판정이 아니다).
- 측정 원천: 본 트리 HEAD `a0d8014096f894c3a45577482b0eb92892716e04` 의 코드 좌표·시험 실행(acceptance.md 증거 장부; 인용 코드는 최초 핀 `3ae43ed8e` 과 동일), 공식 훅 레퍼런스. 리더 제공(미재현): t1425 수치.
- 충돌 사전 검사: 형제 SPEC 이 같은 `HandleCodexReviewGate` 를 쓴다(plan.md §G 위험 2). t1399(`MOAI_KANBAN*` 삭제)는 `reviewGateEnvContext` 에 닿는다. Codex 경로·multi 게이트는 건드리지 않는다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
