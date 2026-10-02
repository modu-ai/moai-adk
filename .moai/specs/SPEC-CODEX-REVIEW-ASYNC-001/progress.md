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

## plan-audit 이력 (2026-10-02)

| 회차 | 판정 | 점수 | 보고서 (로컬 증거) | 감사 커밋 |
|---|---|---|---|---|
| 1 | FAIL | 0.74 | `.moai/reports/t1422/async-plan-audit.md` | `a0d8014096f894c3a45577482b0eb92892716e04` |
| 2 | **PASS** | **0.84** (Tier M 임계 0.80) | `.moai/reports/t1422/async-plan-audit-iter2.md` | `0eb437eec67d03db2e077511341481bf7f6ae346` |

2회차: 필수 통과 기준 아홉 개 모두 통과, 필수 수정 0, 권고 7(S1-S7), 노트 12(N1-N12), 1회차 항목 21개 모두 해결 또는 공시 수용(미해결 0). 점수 여유는 0.04 로 얇다. 감사 도구 환경에서 형식 verb(REQ↔AC 추적 awk, CN-4, D8)는 격리 가드에 거부돼 grep 손 추적으로 대체됐다(감사 Gaps). `moai spec lint` 도 설치된 빌드의 결과다.

## 해시 기준선 — Kickoff "산출물 불변" 점검용 (HEAD `0eb437eec67d03db2e077511341481bf7f6ae346`, 이 레인이 `shasum -a 256` 으로 계산)

```text
496f26779d92395a6c73639c8f4919bbee14f5258590164270ea5b81edae82c5  .moai/specs/SPEC-CODEX-REVIEW-ASYNC-001/spec.md
8fd7317e823d7c98b936707acfa8c904e789240093d143b3cdde0ba020bcf40a  .moai/specs/SPEC-CODEX-REVIEW-ASYNC-001/plan.md
ecc196ff4054c123762a2bf7c70e5c4c6df530067321d60e2f3f552b1558b1bd  .moai/specs/SPEC-CODEX-REVIEW-ASYNC-001/acceptance.md
```

감사 2회차가 읽은 값과 같다. **이 PASS 는 이 해시에 걸려 있다** — 세 파일(`spec.md`·`plan.md`·`acceptance.md`)을 고치면 PASS 가 무효가 된다. `progress.md`·`decision-index.md`·프로브 키트는 해시 대상이 아니다.

## AMENDMENT REQUIRED BEFORE ASYNC KICKOFF (감사 2회차 S1-S5 + 프로브 결과)

> **이 SPEC 은 감사를 통과했지만 아래 다섯 결함은 PASS 가 닫지 않은 채 남아 있다. ASYNC 의 run(Kickoff)을 시작하기 전에 스펙 개정으로 반드시 닫는다.** 지금은 spec/plan/acceptance 를 고치지 않는다(해시 보존). 아래 항목과 프로브 결과(E-1~E-5, O-B 타임아웃 마진)를 **하나의 스펙 개정**으로 합쳐 반영하고, 그 뒤 **마지막 재감사(3회차)** 를 거쳐야 ASYNC run 을 시작한다. ASYNC 의 run 은 `SPEC-CODEX-REVIEW-OWNERSHIP-001` 이 `completed` 가 된 뒤에만 시작하므로(`depends_on`; 사전 점검이 M1 도 막는다) 그 사이에 프로브를 먼저 돌려 결과를 개정에 담을 수 있다.

| ID | 결함 (감사 2회차) | 증거 | 개정이 정해야 할 것 |
|---|---|---|---|
| **S1** | 락·상태를 쓸 수 없을 때 **실패 verdict 가 무엇을 전달하는지 정의돼 있지 않다.** plan 단계 2·4 는 "락·상태 없이 진행/상태 기반 단계 건너뜀"이라 하는데 요약 구성·상한 확인·기록 쓰기가 모두 단계 9 안에 있다. 전달하면 상한·중복 제거가 없는 경로가 되어 무한 깨움 루프가 되고(변이), 종전 `Reason`("codex review gate: "+Summary)을 전달하면 sentinel 줄이 없어 래퍼가 exit 0 으로 접어 실제 실패가 사라진다. | 보존 시험 R1 `TestReviewGate_CodexFailBlocks` 가 `HandleCodexReviewGate(gateInput(false), true, "/proj")` 로 **존재하지 않는 비 git 디렉터리**에서 `Decision == block`·비어 있지 않은 `Reason` 을 요구한다(`codex_review_gate_test.go:125-140`) — 이 경로가 바로 정의 없는 분기다. REQ-CRA-013 은 "락/상태 불가"를 억제 사건으로 나열하는데 §F 는 리뷰가 진행된다고 한다. | 기록을 남길 수 없을 때 실패 리뷰를 **[sentinel 요약으로 한 번 전달하고 기록 | 전달하지 않고 기록]** 중 무엇으로 하는지 REQ-CRA-001/013 에 한 문장, plan 에 한 단계, AC(락 불가·상태 불가 변형 × fail verdict)에 종료 코드·sentinel 유무·다음 Stop 의 이중 전달 없음. R1 의 비 git 디렉터리 사례를 이 정의와 맞출 것 |
| **S2** | **연속 실패 카운터가 전달보다 먼저 증가한다.** 단계 9 가 `consecutive_fail = 기록+1` 을 전달 전에 쓰므로 AC-005(e) 시나리오(전달 쓰기 실패)에서 같은 상태 재시도가 한 번 더 증가한다 — 카운터는 깨움이 아니라 **시도** 횟수가 되고, 마지막 알림 문구 "this is the third consecutive failure notification" 이 실제 전달 횟수보다 많은 수를 주장할 수 있다. | plan 단계 9, spec REQ-CRA-014("failure wake-ups"). AC-014 는 전달 실패와 상한을 결합하지 않고 AC-005(e) 는 카운터를 읽지 않는다. | 카운터를 `markReviewDelivered` 에서 올리거나 미전달 기록의 같은 상태·세션 재처리에서는 올리지 않는다; 마지막 알림 문구와 AC(실패 → 전달 쓰기 실패 → 재시도 → 카운터 == 전달 수) |
| **S3** | **REQ-CRA-005 와 REQ-CRA-014 가 상한 뒤 "다른 세션·같은 상태"에서 충돌한다.** 005 는 다른 세션의 fail 을 미전달로 보고 "다시 리뷰·전달"한다고 하고, 014 는 상한 뒤 모든 실패가 침묵이라 한다. plan 단계 5 절 나(상한 뒤 같은 상태는 리뷰조차 건너뜀)는 어느 REQ 문장에도 어느 AC 에도 없다. | AC-014 의 상한 뒤 단계 D·E·G 는 모두 새 상태·한 세션이라 "상한 뒤 같은 상태에 다른 세션" 변이(재전달)가 산다(감사 변이 8). | REQ-CRA-005 에 "REQ-CRA-014 가 적용되는 경우를 제외하고"를 넣고, plan 단계 5 절 나에 REQ 문장을 주고, AC-014 에 단계 하나(상한 뒤·같은 상태·다른 세션 ⇒ 침묵, 리뷰어 호출 0) |
| **S4** | **래퍼의 `non-verdict-exit-2` 로그 행이 어느 트리의 것인지 정해져 있지 않다.** Go 쪽은 훅 입력 cwd 로 트리를 정하는데(`reviewScopeSessionDir`, `treeroot.go:15-19` 는 `CLAUDE_PROJECT_DIR` 를 일부러 보지 않는다) 래퍼가 가진 것은 `${CLAUDE_PROJECT_DIR:-$PWD}`(래퍼 `:26-27`)뿐이다. 연결 워크트리 세션에서 `CLAUDE_PROJECT_DIR` 는 primary 를 가리키므로 래퍼가 그것으로 git 디렉터리를 찾으면 행이 primary 의 `.git/moai/` 에 쓰이고 Go 쪽 행은 워크트리 것에 쓰여 로그가 갈라진다. | plan §B.5, spec REQ-CRA-003·013; AC-003(b)·AC-013 은 "git 저장소 픽스처"만 쓴다(감사 변이 17). | 래퍼가 트리를 훅 입력 `cwd`(없으면 `$PWD`)에서 구한다고 REQ-CRA-003/plan §B.5 에 명시, AC-003(b)에 cwd 와 `CLAUDE_PROJECT_DIR` 가 서로 다른 트리를 가리키는 두 트리 픽스처 |
| **S5** | **AC-011 의 "결과 토큰이 있다"가 빈 기록 템플릿으로 충족된다.** 템플릿의 verdict 칸이 세 토큰을 모두 범례로 담았으므로 `cp record-template.txt …/E1-interactive.txt` 만으로 L10 의 `ls` 가 성공하고 토큰 검사가 통과한다 — 등록(M5)을 게이트하는 파일이다. | acceptance L10, AC-011(a), 템플릿. | "토큰"을 `verdict :` 뒤의 **단일 값**으로 정의하고 정확히 하나일 것을 요구, L10 의 RED/GREEN 이 그 값을 관측. **키트 쪽은 이미 고쳤다**(아래 — 템플릿에서 범례 제거); AC 문구와 L10 셀은 개정 몫 |

**이 개정에 함께 접어 넣을 선택 항목(감사 노트, 싸게 닫히는 것만):** N1/N2 `reviewDeliveryTicket` 전역 변수 대신 내부 평가 함수가 ticket 을 반환하게 하고 실행기가 락 안에서 `markReviewDelivered` 를 부르면 표지 구간·읽고-고치는 쓰기 경합이 사라진다; N4 래퍼가 종료 코드와 무관하게 sentinel 만으로 깨우는 변이(AC-003(v)의 가짜 stderr 에 sentinel 문구 요구)·`Summary` 에 맨 sentinel 줄이 들어가는 경우의 정화·`mktemp` 부재/잔여 파일 처리·로그 행; N7 잔여 위험 한 줄(락 때문에 건너뛴 Stop 의 상태가 후속 Stop 이 없으면 리뷰되지 않는다); N9 대용량 stderr 변형을 `E4-d2-exit0stderr` 에 더해 8000바이트 시작값을 관측으로; N10 `AcquireReviewLock` 이 윈도우에서 "held" 를 돌려주기 전에 `ClearStaleGateLock` 의 경로형을 부르는지 명시; N11 파일 수 하한 표기(코드+테스트 ≈15-16, 합 ≈20); N12 컴파일된 `moai` 로 끝에서 끝까지 도는 AC 한 건.

**프로브 키트 수정(로컬, 해시 밖) — 이 개정이 아니라 지금 했다:**

- **S6 (감사)** 프로브 훅이 매번 종료 코드 2 를 내 E-1 이 양성이면 자기 유지 깨움 루프가 시작되고 E3-c1 이 오염되던 문제: `probe-hook.sh` 가 요청된 종료 코드 2 를 이 일회용 프로젝트에서 처음 `<예산>` 번에만 낸다 — 예산은 `probe.wake-budget` 파일의 정수(없거나 잘못되면 **기본 1**), 호출 순번은 `mkdir` 로 원자적으로 배정. E3-c1 은 예산 0. 절차 문서(§1 깨움 예산, §3·§4·§6)와 기대 관측을 갱신했다.
- **S7 (감사)** `stream-session.sh` 의 `CLAUDE-EXIT status=$?` 가 항상 0 이던 문제: `wait` 바로 다음 줄에서 `claude_rc=$?` 로 잡는다.
- **S5 (키트 쪽)** `record-template.txt` 의 verdict 줄에서 세 토큰 범례를 지웠다(빈 칸 + 채움 규칙 주석).
- 자체 시험(일회용 mktemp 프로젝트, 모델 호출 없음, 시험 스크립트는 스크래치패드): 기본 예산 `2 0 0` · 예산 0 `0` · 예산 2 `2 2 0` · 잘못된 예산 파일 `2 0` · 동시 5 호출에서 종료 코드 2 정확히 1, 0 이 4 · 종료 코드 0 요청 불변 · 잘못된 인자 `64` · 가짜 `claude`(종료 3)에서 `status=3`. `bash -n` 여섯 스크립트 + `run-clean.sh`·`rm-scratch.sh` 모두 통과. `new-scratch.sh` 는 `TMPDIR` 를 저장소 안으로 돌리면 `ABORT: scratch directory is inside the repository` 로 거부하고 자기 디렉터리를 지우며 `rm-scratch.sh` 는 `t1422-probe-*` 가 아닌 디렉터리를 거부한다(둘 다 관측).

## 인계 항목

0. **ASYNC run 시작 조건.** (a) `SPEC-CODEX-REVIEW-OWNERSHIP-001` 이 `completed`, (b) 위「AMENDMENT REQUIRED BEFORE ASYNC KICKOFF」의 스펙 개정 1회와 마지막 재감사 PASS, (c) 프로브 결과 E-1~E-5 와 O-B 마진 값(개정에 접어 넣음). 위 해시 기준선의 PASS 로 자율 Kickoff 를 하지 않는다.
1. **프로브 실행 주체(O-F 확정).** 레인이 헤드리스 변형(`E1-headless`·`E2-timeout`·`E3-c1-concurrent`·`E4-d1-json`·`E4-d2-exit0stderr`)을 먼저 돌리고, 대화형만 가능한 `E1-interactive`·`E3-c2-newturn`·`E3-c3-sessionexit`·`E5-reload` 는 리더가 리더 세션에서 `.moai/reports/t1422/async-probes.md` 절차대로 돌린다. 헤드리스 결과가 대화형과 같다는 보장은 없으므로 `E1-headless` 는 E-1 을 결정하지 못한다. 스트림 입력 헤드리스 구동 스크립트(`stream-session.sh`)의 플래그·입력 형식은 아직 실행해 보지 않은 틀이다.
2. **형제 SPEC 착지.** 이 SPEC 의 M2 이후는 형제 SPEC 이 착지한 트리에서 시작한다(`depends_on`, 핸들러 삽입 위치 확정을 위해 소스를 다시 읽는다).
3. **결정 현황.** O-A·O-C~O-J 는 확정(A-5, A-7~A-14), O-B 는 E-2 결과 전까지 OPEN, O-I 는 E-4 결과 대기. 상한 동작(O-A)의 읽기 확인 요청 한 건이 plan.md §G 에 있다.

## §E.1 Plan-phase Audit-Ready Signal

- 산출: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` · `decision-index.md` (Tier M 산출 집합 + progress + decision gate `interview.decision_gate: on`).
- Tier: **M**. 개수 규칙: 요구 = `### REQ-` 제목 수 15, 수락 기준 = `### AC-` 제목 수 15(하위 ID 없음) — 상한 16 이하(여유 1). 파일: 코드+테스트 ≈14(하한), 미러·문서 ≈5, 합 ≈19(Tier L 대역에 걸림 → O-J 에서 M 유지로 확정, 신뢰도 0.70). 규모는 초안 ≈220 LOC 에서 Go ≈260 + 셸 ≈35 로 늘었다(plan §H) — 리더에게 재통지.
- plan_status: **audit-ready** — plan_complete_at: 2026-10-02. plan-audit 1회차 = FAIL 0.74, **2회차 = PASS 0.84**(보고서 `.moai/reports/t1422/async-plan-audit-iter2.md`, 감사 커밋 `0eb437eec67d03db2e077511341481bf7f6ae346` — 위「plan-audit 이력」). **단, ASYNC run 전에 스펙 개정이 필요하다**(위「AMENDMENT REQUIRED BEFORE ASYNC KICKOFF」 S1-S5 + 프로브 결과 → 개정 1회 → 마지막 재감사). 해시 기준선은 위 절의 세 값이다.
- 측정 원천: 본 트리 HEAD `a0d8014096f894c3a45577482b0eb92892716e04` 의 코드 좌표·시험 실행(acceptance.md 증거 장부; 인용 코드는 최초 핀 `3ae43ed8e` 과 동일), 공식 훅 레퍼런스. 리더 제공(미재현): t1425 수치.
- 충돌 사전 검사: 형제 SPEC 이 같은 `HandleCodexReviewGate` 를 쓴다(plan.md §G 위험 2). t1399(`MOAI_KANBAN*` 삭제)는 `reviewGateEnvContext` 에 닿는다. Codex 경로·multi 게이트는 건드리지 않는다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
