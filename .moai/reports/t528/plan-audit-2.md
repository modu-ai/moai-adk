# SPEC Review Report: SPEC-AC-COLLECTOR-ANCHOR-001 (카드 t528) — iteration 2

- Iteration: 2/2 (Tier M 상한 도달)
- **Verdict: PASS-WITH-DEBT** — must-pass 7건 전부 통과, 총점 0.84 ≥ Tier M 임계 0.80. 다만 **critical blocking 2건**은 M1(파서 편집) 착수 **전에** 반드시 해소돼야 하는 부채로 남긴다(§ Debt gate).
- Overall Score: **0.84** (iteration 1: 0.80 → 회귀 없음, STOP 신호 미발동)
- 감사 트리: `.claude/worktrees/t528`, 브랜치 `WT-ac-collector-anchor`, HEAD `52f863f36`
- **감사 대상 핀 (SHA-256, 2026-09-08 02:51 KST 재판독)**
  - `spec.md` `86f81c260276cc9f3a8e3a97d269db4e1333ee7f220e75a05234573118e07b3a`
  - `plan.md` `328c7081db7719743d236f6fac14aa6ed826bcb2516fa68319a2a6cffcfec262`
  - `acceptance.md` `406815040a6feca94cbe79086d46260363aaddcb0eb396bb662547eb903a70c0`
- Reasoning context ignored per M1 Context Isolation. 배차문의 서술은 감사 입력이 아니라 **검증할 주장**으로만 소비했다.
- 이것은 재감사다. 판정 근거는 「개정이 그렇게 하겠다고 적었는가」가 아니라 **「트리가 실제로 그렇게 되어 있는가」**다.

---

## P0 — 감사 창 중 대상이 두 번 이동했다 (프로세스 결함)

**Claim**: 감사 대상 3파일이 감사 진행 중 **두 번** 재작성됐다.

**Evidence**:
```
02:44:35 acceptance.md / 02:42:35 plan.md / 02:46:37 spec.md   ← 최초 판독 시점
02:48:13 acceptance.md / 02:47:23 plan.md / 02:46:37 spec.md   ← 1차 이동(plan.md 201→248줄)
02:48:39 spec.md · 02:48:54 plan.md                            ← 2차 이동(스냅샷 이후)
```
**Baseline-attribution**: `stat -f '%Sm %z %N'` + `shasum -a 256` 3회, 같은 트리, HEAD 불변 `52f863f36`.

`agent-common-protocol.md` § Background Agent Execution: **「감사 중인 워크트리는 정확히 한 명의 기록자를 갖는다」**. 이 창에는 둘이 있었다.

- 감사 절차는 스냅샷을 떠서 안정화했고(`scratchpad/t528-snap/`), **판정은 위 최종 핀에 대해 내린다.**
- 2차 이동분을 diff로 대조한 결과 **아래 D1'·D2' 두 결함에 닿는 편집은 없었다** — 교차참조 보강, Gaps 분해, 반패턴 4건 추가뿐이다. 즉 판정은 이동에 영향받지 않았다.
- 그럼에도 이것은 조용히 지나갈 사항이 아니다: 이동이 결함 자리에 닿았다면 이 보고서 전체가 존재하지 않는 트리에 대한 것이 됐다. 리드에게 보고한다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `REQ-ACA-001-001` ~ `-016` 16건, `grep -o 'REQ-ACA-001-[0-9]*' | sort -u` 결과 결번·중복 0, 제로패딩 일관.
- **[PASS] MP-2 GEARS 형식 (requirement layer)** — 신규 5건 전수 확인: `-012`/`-014`/`-015`/`-016` ubiquitous(`…해야 한다(shall)`), `-013` **Where** + `shall`/`shall not` 복합. 기존 11건은 iteration 1에서 통과, 문면 변경 없음. `acceptance.md`의 Given-When-Then은 검증층이므로 여기서 벌하지 않았다(M3 § Scope) — 판정은 `spec.md` §2 요구층에 대해서만 내렸다.
- **[PASS] MP-3 프론트매터** — `spec.md:1-15` canonical 12필드 전부 present, snake_case alias 0, `version: "0.1.0"` quoted, `tier: M` optional 적법.
- **[N/A] MP-4 언어 중립성** — `internal/spec` Go 단일 패키지 한정. 16개 프로그래밍 언어를 다루지 않는다.
- **[PASS] MP-5 D7 교차-SPEC** — 실행 결과: `SPEC-ARTIFACT-STATELESS-001` / `SPEC-CLIFIX-CONCURRENCY-001` / `SPEC-COVERAGE-RULE-SCOPE-001` 셋 다 존재 · `status: completed`. retired/superseded/archived 0건.
- **[PASS] MP-6 D8 크로스플랫폼** — `grep -c 'syscall'` = `0 / 0 / 0`. 자동 통과.
- **[PASS] MP-7 clarification gate** — 실행:
  ```
  $ grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-AC-COLLECTOR-ANCHOR-001/
  (무출력)  rc=1
  ```
  **마커 삭제만으로 통과시키지 않았다.** 결정이 실제로 닫혔는지 별도로 판정했다: `plan.md` §B.1 「결정(확정)」, `spec.md` §1.4 「이 결정은 이제 열린 상태가 아니다」, `acceptance.md:19` 「이 문서와 `plan.md`는 같은 결정을 말한다」 — 세 문서가 같은 결정을 같은 상태로 말한다. iteration 1의 D2(열린 결정 vs 확정 취급) 모순은 **해소됐다.**

  **다만 그 결정의 근거 절반이 거짓이다 — 아래 D1'.** 마커 삭제는 형식 통과이고, 근거의 진위는 별개 축이라 defect로 분리했다. 배차문이 물은 「설명 산문에서 마커 문자열을 걷어내 grep을 통과시킨 것이 정당한가」에 대한 답: **정당하다.** `acceptance.md`의 판정 문장(`Definition of Done`)이 그 이유를 스스로 적는다 — 판정문이 마커 문자열을 담으면 자기 자신에게 걸려 영원히 FAIL이다. 자기충족적 검사가 아니라 자기참조 회피다.

---

## Category Scores (rubric-anchored)

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 대부분 단일 해석. 그러나 `acceptance.md` §D.0의 판별식 라벨이 보존된 산출물과 정면으로 어긋나, 구현자가 같은 명령을 두 가지로 읽는다(D2') |
| Completeness | 0.85 | 0.75~1.0 | 전 필수 절 present, `### Out of Scope — <topic>` H3 **8개** + 불릿, Gaps 10항 · Residual-risk 6항. 감점: 헤드라인 수치 944 / 100이 어떤 보존 산출물로도 재유도되지 않고, 보존 프로브가 내는 951 / 101과의 대조가 기록에 없다(D2') |
| Testability | 0.75 | 0.75 | 큰 폭 개선 — 비공허성 3요소(D1 해소), 양성 대조군(D3 해소), 과수용 AC(D5 해소), 순서 게이트(D2 해소). 잔여: AC-001의 비교 지시가 판별식이 다른 두 수를 나란히 놓게 만든다(D2') |
| Traceability | 1.00 | 1.0 | REQ 16건 전부 ≥1 AC 대응(매트릭스 전수 대조), 고아 AC 0, 존재하지 않는 REQ 인용 0, 형제 문서의 AC 인용 전부 실재 AC로 해소 |

산술 평균 **0.8375 → 0.84**. Tier M 임계 0.80 초과. iteration 1(0.80) 대비 상승 — 점수 회귀 없음이므로 LEAN STOP 신호는 발동하지 않는다.

---

## 종전 결함의 해소 판정 (Regression Check — iteration 1 defect delta)

| ID | 판정 | 근거 |
|---|---|---|
| **MP-7** | **해소** | `grep` 무출력 + 세 문서의 결정 상태 일치 |
| **D1** (공허한 판정 명령) | **부분 해소** | 프로브 소스 3본 + `EXIT=0` 출력 3본이 트리에 있고, AC-001에 3요소 비공허성 가드가 들어갔다. **그러나 판별식이 어긋난 채 보존됐다 → D2'로 승계** |
| **D2** (M1→M2 순서 모순) | **해소** | `plan.md` §E가 M0(대조군) → M1(구현) → M2(재측정) → M3(비교)로 재배열. 종전 순서를 지시하는 잔여 문장 0건(전수 확인). AC-002의 before-image가 `52f863f36`(파서 편집 이전) 산출물에 핀됨 |
| **D3** (신호 없는 경로 계측) | **해소** | AC-010 셋째 절이 lint 경로에서 `ParseAcceptanceCriteria` 두 번째 반환값 직독으로 재지정. AC-015가 CLI 경로를 별도로 계측. `internal/cli/spec_view.go` 바이트 동일성이 인수 기준(`acceptance.md:356`, `plan.md` §D, `spec.md` §3)으로 들어감. halt-and-report 4곳 명시. **양성 대조군 존재 확인**(`spec.md:246`, `plan.md` M7, `acceptance.md` §D.15 등급 주석) |
| **D4** (「가드를 약화시키지 않는다」) | **해소** | `spec.md` §4에 「종전 주장의 철회」 절이 신설되고, 감사가 만든 반례 `- AC-OGR-003 (RETIRED) — …`가 verbatim으로 인용됨. 무효 결과(엠대시 19파일 전수 열람에서 오탐 0)가 **미측정의 표시로** 명시되고 944줄 미열람 한계가 함께 적힘 |
| **D5** (과수용 무측정) | **해소** | REQ-012 + AC-014 신설. 표본 ≥30 + **코퍼스에서 뽑은** 비-AC 불릿 픽스처(발명 금지 명시). `plan.md` §C.1이 「뮤턴트는 이 축을 그릴 수 없다」를 [HARD]로 못박음 |
| **D6** (핀 SHA가 집합을 고정 못 함) | **해소** | `probe/filelist.txt`(807줄)가 분모를 산출물로 고정. own-card 효과를 실측으로 분리(파일 +1, 선언 +0) |
| **D7** (뮤턴트 6 미행사) | **해소** | 뮤턴트 표가 6/7행으로 분리됨(`plan.md:90-91`, `acceptance.md:254-255`) |
| **D8** (소비자 지도 누락) | **해소** | `lint.go:636` 호출 지점이 지도에 들어감. `CheckDanglingReferences` 프로덕션 호출자 0 기록. `buildTree:136-141` 중복 `continue` 폐기가 구체적 실패 모양과 함께 3곳에 기록 |
| **D9** (등급 미분류) | **해소** | 16건 전수 분류 + 「must-pass 표기 ≠ 진척」 [HARD] 주석 |
| **D10** (`.1` 형 미논의) | **해소** | §1.4가 점 하위번호 10건을 명시 분해하고 별도 축으로 처분 |

---

## Defects Found

### D1' — §1.4가 「선언이 아니다」라고 적은 7건은 **전부 진짜 AC 선언이다**
- 위치: `spec.md` §1.4 단어 꼬리 행 + 그 아래 첫 근거 불릿 · `spec.md` §4 「Out of Scope — 미포함 39건」 · `plan.md` §B.1 항목 1
- Severity: **critical** · Class: **blocking**

**Claim**: 개정은 미포함 39건을 「규칙이 작동한 결과 7」 + 「진짜 선언일 개연성 32」로 가르고, 7건에 대해 **「다른 SPEC 본문이 AC 절 안에서 개념 라벨로 쓴 토큰이지 선언 ID가 아니다」**, **「영구 배제」**라고 적는다. 코퍼스를 열어 보면 **7건 전부가 Given-When-Then을 갖춘 실제 AC 선언 불릿**이다.

**Evidence** (실행: `/usr/bin/grep -rn -E '^\s*[-*+]\s+\*{0,2}AC-(GREEN|MUTANT|EVIDENCE|SCOPE|ORDERING|MRR-GREEN|HFC-GATE)\b' .moai/specs --include=spec.md`):
```
SPEC-SPECLINT-ARTIFACT-STATUS-001/spec.md:48:- **AC-GREEN**: Given the repaired tree, When `go run ./cmd/moai spec lint --strict` runs …
SPEC-SPECLINT-ARTIFACT-STATUS-001/spec.md:49:- **AC-MUTANT**: Given the two `status:` lines restored via `git checkout 615d18c1f -- …`, When …
SPEC-SPECLINT-ARTIFACT-STATUS-001/spec.md:50:- **AC-ORDERING**: Given the branch history, When `git log --oneline` is read, Then …
SPEC-SPECLINT-ARTIFACT-STATUS-001/spec.md:51:- **AC-SCOPE**: Given the branch diff for the two files, When … Then it shows exactly two deleted lines …
SPEC-SPECLINT-ARTIFACT-STATUS-001/spec.md:52:- **AC-EVIDENCE** (parent: REQ-007): Given `.moai/reports/t490/`, When the evidence files are read, Then …
SPEC-V3R6-MAIN-RED-REMEDIATION-001/spec.md:124:- AC-MRR-GREEN: `go test ./internal/template/...` 0 fail + cross-platform build exit 0 + lint baseline 유지.
SPEC-HOOK-FAILURE-CLASSIFY-001/spec.md:97:- **AC-HFC-GATE** (quality gate): **Given** the full package, **When** `go test ./internal/hook/… ` runs, **Then** …
```
**Baseline-attribution**: 트리 `52f863f36`, 명령 위 1건, `/usr/bin/grep`(셸 `grep`은 ugrep 래퍼라 조용히 건너뛴다), 대상 `.moai/specs/**/spec.md`. 7개 형태 전부에 대해 각각 1건 이상 실재 확인 — 프로브가 센 건수(각 1)와 일치한다.

**왜 blocking인가.** 이것은 문구 다툼이 아니라 **MP-7이 닫았다고 판정한 그 결정의 근거**다(VCI §1.1 surface 4 — 권고의 전제). 개정은 이 7건의 배제를 「가드가 일한 증거」로 제시하고, 그 위에서 숫자 꼬리 고정을 확정했다. 근거가 반대로 뒤집힌다: 7건은 **32건과 같은 부류** — 진짜 선언인데 이 카드의 축으로는 회수하지 못하는 것 — 이며, 「영구 배제」 처분은 근거를 잃는다.

**결정 자체는 살아남는다**(중요): 숫자 꼬리 후보는 여전히 1128/1167을 덮고, 단어 꼬리를 받으려면 `AC-` 접두 임의 토큰을 전부 받게 되므로 별도 축이 필요하다. 무너지는 것은 **분류와 처분**이지 결정이 아니다. 그래서 재설계가 아니라 정정으로 갚을 수 있는 부채다.

**부수 관측(Gap)**: 「숫자 꼬리를 놓으면 `AC-`로 시작하는 임의 대문자 토큰이 전부 들어온다」는 **실측 근거가 0이다.** 관측된 16개 미포함 형태 중 산문 라벨은 **하나도 없다.** 위험은 개연적이지만 코퍼스에서 확인된 바 없으며, 그 사실이 기록에 없다.

**아프게 남길 것**: 이 결함은 개정 자신이 `plan.md` §D.1에 [HARD]로 승격한 규율 — **「코퍼스 수치는 그 뒤의 표본 줄을 읽기 전까지 채택하지 않는다」** — 을 그 규율을 쓴 문서가 스스로 지키지 않아 생겼다. 39건 분해는 표본 줄 열람 없이 채택됐다. 판별식 오형성 **네 번째** 사례로 HISTORY 표에 들어가야 한다.

**필요한 수정**:
1. §1.4 단어 꼬리 행의 처분을 「의도적 배제 — 규칙이 작동한 결과」 → **「진짜 선언, 다른 축, 이 카드 밖」**으로 정정하고, 위 7줄을 verbatim 인용한다.
2. §4 「단어 꼬리 7건은 **영구 배제**다」를 철회한다 — 후속 카드 후보로 32건과 합쳐 **39건**으로 적는다.
3. `plan.md` §B.1 항목 1을 같은 방향으로 정정한다(「이 규칙이 존재하는 이유」 → 「이 카드의 축이 닿지 않는 형태」).
4. 「임의 토큰이 전부 들어온다」를 **미측정 가설**로 명시하거나, 코퍼스에서 실제 산문 라벨을 찾아 근거를 세운다.
5. HISTORY 「이것이 이 카드에서 세 번째다」 표에 4행을 추가한다.

### D2' — 보존된 프로브는 **판별식 B**인데 §D.0이 **판별식 A**라고 적는다 (D1의 4번째 재발)
- 위치: `acceptance.md` §D.0 「판별식」 행 · 같은 절 「verbatim 출력」 행 · `acceptance.md` §D.1 · `spec.md` §1.1 표의 `216` / `944` / `100` 행(산출물 열 `—`)
- Severity: **critical** · Class: **blocking**

**Claim**: `acceptance.md` §D.0은 「판별식: **A**(기준선) — `probe/ac_anchor_probe_test.go`가 판별식 그 자체다」라고 적는다. 그 프로브는 **판별식 B**다.

**Evidence** (소스 대조 — `spec.md` §1.1이 정의한 두 판별식):
```
spec.md §1.1 정의
  판별식 A: ^\s*[-*+]\s+\*{0,2}(AC-[A-Za-z0-9-]*[0-9])\*{0,2}\s*(.*)$      ← 점 불가, 숫자 종료
  판별식 B: 위에서 점(.)을 허용하고 ID가 문자로 끝나도 받는다

probe/ac_anchor_probe_test.go:30 (보존된 프로브, AC-001의 판정 명령이 부르는 것)
  var declRe = regexp.MustCompile(`^\s*[-*+]\s+\*{0,2}(AC-[A-Za-z0-9.-]*[A-Za-z0-9])\*{0,2}\s*(.*)$`)
                                                              ^^^ 점 허용    ^^^^^^^^^^ 문자 종료  → 판별식 B

probe/drift_probe_test.go:18-19 (판별식 A는 여기에만 있다)
  var declRe_A = ...(AC-[A-Za-z0-9-]*[0-9])...     var declRe_B = ...(AC-[A-Za-z0-9.-]*[A-Za-z0-9])...
```
출력이 그 사실을 확증한다:
```
probe/run-20260908.txt   (TestT528Anchor, 보존 프로브)
  IN-SECTION declarations = 1167 · accepted = 216 · rejected = 951 · FULLY-BLIND = 101
probe/drift-20260908.txt (TestT528DiscriminatorDrift)
  discriminator A ... decls=1160     discriminator B ... decls=1167
measurement-20260908.md:24-27 (삭제된 zz_t528_probe_test.go, 판별식 A)
  in-section=1160 · accepted-by-current-parser=216 · rejected=944 · fully blind = 100
```
**Baseline-attribution**: 트리 `52f863f36`, 위 4개 파일 직독(`cat -n`), 재유도 없음.

**결과 셋을 못박는다.**

1. **§D.0의 RED-now 셀이 두 실행을 하나로 붙여 놓았다.** 「명령: `-run TestT528Anchor`」 + 「verbatim 출력: `run-20260908.txt` 전문」 + 「기준선 세 줄은 `measurement-20260908.md`」 — 그런데 그 명령의 출력에는 **1160도 944도 100도 없다.** `verification-completeness.md` §2.1이 요구하는 것은 「그 명령이 실제로 낸 출력」이며, 여기서는 명령과 출력의 소유자가 다르다. release-blocking AC 2건이 이 셀 위에 서 있다.
2. **944와 100은 어떤 보존 산출물로도 재유도되지 않는다.** 1160만 `drift-20260908.txt`가 되살린다. `grep -rn '951\|101'`은 SPEC 3파일에서 **무출력** — 보존 프로브가 내는 값과 문서가 인용하는 값의 대조가 기록 어디에도 없다. D1의 「판별식이 사라지면 수치는 Claim이 되고 Evidence가 아니게 된다」가 **두 수에 대해 그대로 살아 있다.**
3. **AC-001이 금지된 비교를 지시한다.** 「`-run TestT528Anchor`를 다시 태워」 나온 수(B 기준)를 「기준선 216/1160(판별식 A)과 **나란히**」 적으라고 한다. `plan.md` §H가 반패턴으로, REQ-016이 [HARD]로 금지한 **판별식이 다른 두 수의 나란한 배치**를 AC 본문이 지시하는 셈이다. `acceptance.md:98`이 구분자 82/42에 대해서는 정확히 이 함정을 경고하면서, 같은 AC의 주 판정 축에서는 그 경고를 적용하지 않았다.
4. **AC-012의 「blind 100개 중 하나」도 같은 문제를 승계한다.** 보존 프로브는 101을 낸다.

**참고(정정된 안전 지점)**: `accepted=216`과 positive-needle 18은 두 판별식에서 동일하다 — A가 잡는 줄은 B도 잡고, 수용 판정은 `currentRe`가 줄 전체에 대해 내리므로 캡처 폭과 무관하다. 즉 **18개 대조군은 유효하고, 216도 유효하다.** 무너지는 것은 944 · 100 · 1160↔1167 대조 축이다. 이 구분을 적지 않으면 정정이 필요 이상으로 커진다.

**필요한 수정**:
1. §D.0 「판별식」 행을 **B**로 정정하고, 「A의 정본은 `drift_probe_test.go`」를 함께 적는다.
2. §D.0을 **두 셀로 분리한다** — (a) 기준선 셀: 명령 `-run TestT528Probe`(삭제됨) · 출력 `measurement-20260908.md:24-27` · 판별식 A · 종료 코드 없음, (b) 재유도 셀: 명령 `-run TestT528Anchor` · 출력 `run-20260908.txt` · 판별식 B · `EXIT=0`. 두 셀의 수가 다른 이유를 한 줄로 적는다.
3. **944 / 100의 처분을 정한다.** 둘 중 하나 — (a) `drift_probe_test.go`를 M0에서 함께 커밋된 테스트로 승격해 A 기준 `accepted`/`rejected`/`blind`까지 내게 한다(권장: 재유도 가능성이 완성된다), 또는 (b) 기준선을 **B 기준 1167/216/951/101**로 재선언하고 A 계열 수치는 「삭제된 프로브의 기록, 재유도 불가」로 강등한다.
4. AC-001의 비교 지시를 **같은 판별식끼리** 비교하도록 고친다. `spec.md` §1.1 표의 `216`/`944`/`100` 행 산출물 열(`—`)도 함께 채운다.
5. HISTORY의 「세 번째다」 표에 이 건을 추가한다 — **판별식 오형성이 이 카드에서 다섯 번째다**(D1'과 합쳐).

### D3' — `plan.md` §I 교차참조가 산출물 수를 낡은 값으로 적었다 (2차 이동에서 자체 수정됨)
- 위치: `plan.md` §I
- Severity: minor · Class: **optional** · **상태: 이미 해소**
- 스냅샷 시점 「프로브 소스 2본 + 실행 출력 2본」(실제 3+3). 02:48:54 편집에서 「3본 + 3본 + `pkg-baseline`」으로 정정됨. 기록으로만 남긴다 — 감사 창 중 대상이 움직였다는 P0의 증거이기도 하다.

### D4' — 코드 인용 줄번호 2건이 ±2 어긋난다
- 위치: `plan.md` §I · `spec.md` §7 — `internal/spec/parser.go:83-85` / `:136-141`
- Severity: minor · Class: **optional**
- 실측: `##` break는 `parser.go:85-87`, 중복 ID `continue`는 `parser.go:137-143`. 나머지 인용은 **전부 정확**하다 — `lint.go:636`(`criteria, _ :=`) · `lint.go:915`(`collectAllREQIDs`) · `spec_view.go:72`/`:80-86`(`default: return fmt.Errorf("parse error: %w", err)`) · `parser.go:218`(현행 앵커) · `CheckDanglingReferences` 프로덕션 호출자 **0**(정의 `parser.go:297`, 호출 `parser_test.go:313` 단 1건) 전수 확인.
- 함수 이름이 함께 적혀 있어 실무 영향은 없다. 다음 편집 때 함께 고치면 충분하다.

### D5' — Tier M 상한 동시 도달이 「알려진 제약」으로 기록되지 않았다
- 위치: `acceptance.md:53-58`
- Severity: minor · Class: **optional**
- REQ 16 / AC 16 — 두 축 모두 Tier M 상한과 **같다.** 접기(fold) 2건은 검사했고 **실질 손실 없음**으로 판정한다: 대조군 선행(REQ-015)은 AC-002의 RED-now 절이 원래 요구하던 순서이고 `acceptance.md:124`에 **독립 판정 절**로 살아 있다(「커밋 순서로 보인다」 + 「뒤집혔으면 다시 만든다」); 재측정 3요소(REQ-016)는 AC-016에 **표 형태로 3행 전부** 보존됐다. 접힌 것은 항 번호이지 판정이 아니다.
- 남는 것은 다음 요구가 생기면 **접을 자리가 없다**는 사실이다. `acceptance.md:58`이 그 답(Tier L 승격 또는 분할)을 적었으나, **지금 두 축이 동시에 상한에 걸려 있다**는 현재 상태를 「알려진 제약」으로 명시하지는 않았다. 한 줄이면 된다.
- **Tier M 유지 자체는 재론하지 않는다**(리드 판정). 파일 하나·패키지 하나·함수 하나 변경에 `design.md` + `research.md`를 붙이는 것은 위험을 줄이지 않는다.

---

## Debt gate — M1 착수 전 해소 조건 (PASS-WITH-DEBT의 조건절)

이 판정은 무조건 PASS가 아니다. **D1'와 D2'는 `internal/spec/parser.go`의 첫 편집 이전에 해소된다.** 둘 다 문서 정정이며 새 측정을 요구하지 않는다(필요한 실측은 이 보고서의 Evidence 절에 이미 있다).

해소 판정은 이진이다.

1. `spec.md` §1.4 단어 꼬리 7건의 처분이 「진짜 선언, 다른 축」으로 바뀌고, 7줄이 verbatim으로 인용돼 있다. §4의 「영구 배제」가 철회돼 있다.
2. `acceptance.md` §D.0의 판별식 라벨이 **B**이고, 기준선 셀과 재유도 셀이 분리돼 있으며, 944 / 100의 처분(승격 또는 강등)이 명시돼 있다.
3. AC-001의 비교 지시가 같은 판별식끼리 비교하도록 고쳐져 있다.
4. HISTORY의 판별식 오형성 표가 5행이다.

넷이 성립하면 이 SPEC은 무조건 PASS다. 성립 전에 M1이 시작되면 **AC-001의 재측정이 기준선과 비교 불가능한 수를 낳고**, 그것은 이 카드가 고치겠다고 나선 결함과 같은 부류의 재생산이다.

---

## Gaps — 이 감사가 관측하지 않은 것

- **넓힌 앵커의 실제 정규식을 보지 못했다.** §B는 네 축의 방향만 정한다. D1'·D2'는 문서와 보존 산출물에 대한 판정이며, 구현이 §B와 다르게 넓히면 크기가 달라진다.
- **944(또는 951)개 거절 줄을 표본 열람하지 않았다.** 과수용 실제 비율은 이 감사도 모른다. AC-014가 그것을 재는 자리이며, 이 감사는 그 AC의 존재와 형태만 판정했다.
- **32건(알파벳 접미·점 하위번호·범위 표기·문자 접두)의 실재성을 개별 확인하지 않았다.** 7건은 전수 확인했고 전부 진짜 선언이었다. 32건은 표본 없이 「진짜 선언일 개연성」으로 남는다 — 방향이 같으므로 결론은 바뀌지 않지만, 확인하지 않았다는 사실을 적는다.
- **`moai spec view`를 코퍼스에 돌리지 않았다.** CLI 치명 경로는 코드 판독 근거이고, 중복 재료 0은 프로브 실측이다. 실제 CLI 실행 건수는 M7 몫이다.
- **`CoverageRule` finding 수를 세지 않았다.** run 단계 몫이며, 이 감사는 AC-010의 계측 지점이 옮겨졌는지만 판정했다.
- **뮤턴트를 주입해 보지 않았다.** 뮤턴트 표의 사전 선언 여부만 문서 판독으로 확인했다.

## Residual-risk

- **감사 창 중 대상이 두 번 이동했다**(P0). 2차 이동분은 diff로 대조해 결함 자리에 닿지 않음을 확인했으나, 이 보고서가 손을 뗀 뒤 3차 이동이 일어나면 위 핀 SHA-256 세 개가 그 사실을 드러낸다. **핀이 다르면 이 판정은 그 트리에 대한 것이 아니다.**
- **D1'가 이 카드의 축 구성을 흔들 여지**: 7건이 진짜 선언으로 확정되면 미포함 39건 전부가 「회수 대상 후보」가 된다. 3.3%라는 크기는 그대로이므로 이 카드의 범위 판정은 유지되지만, 후속 카드의 크기 예측은 32가 아니라 39에서 출발해야 한다.
- **판별식 오형성이 이 카드에서 다섯 번 재현됐다**(HISTORY 3 + D1' + D2'). 개정이 그 부류를 스스로 이름 붙이고 규율로 승격했는데도 두 번 더 일어났다. 이는 문서 품질의 문제가 아니라 **이 작업의 성질**이다 — run 단계에서도 같은 빈도로 일어날 것으로 보아야 하며, `plan.md` §D.1의 표본 열람 규율과 AC-015의 양성 대조군이 그 방어선이다. 방어선이 세워진 것은 이 개정의 실질 성과다.
- 이 감사는 **문서와 코드와 산출물**을 읽었고 **구현을 보지 않았다.**

---

## Recommendation

**PASS-WITH-DEBT.** iteration 2는 Tier M 상한이므로 iteration 3은 없다. 아래를 리드에게 권고한다.

1. **Debt gate 4건을 M0 안에서, M1 착수 전에 해소한다.** 전부 문서 정정이고 새 측정을 요구하지 않는다 — 필요한 실측은 이 보고서 D1'·D2'의 Evidence 절이 그대로 쓸 수 있는 형태로 담고 있다.
2. **재감사는 요구하지 않는다.** 해소 판정이 이진이므로 M0 산출물에서 4항을 확인하는 것으로 충분하다. 다시 감사자를 부르는 비용이 이득을 넘는다.
3. optional 3건(D3'는 이미 해소, D4' 줄번호, D5' 상한 기록)은 같은 편집에서 함께 처리하면 싸다. **이것들이 판정을 좌우하지 않는다** — 이 verdict는 D1'·D2' 두 건이 만든다.

**근거와 함께 칭찬을 남긴다.** iteration 1의 blocking 6건 + MP-7이 전부 실질적으로 닫혔고, 닫는 방식이 옳았다 — 주장으로 반박하지 않고 **프로브를 트리에 남기고, 판별식을 분리해 태우고, 거짓 수치를 스스로 정정해 기록했다.** 특히 「4 파일 / 24 줄」 정정은 이 프로젝트 증거 규율의 모범이다: 간극을 **넓히는** 방향의 정정이었는데도 숨기지 않고 HISTORY에 표로 남겼고, 그 사건에서 「코퍼스 수치는 표본 줄을 읽기 전까지 채택하지 않는다」는 규율을 스스로 뽑아냈다. AC-015의 양성 대조군 요구 — 「0은 재료 없음과 탐지기 미작동 둘 다로 나타난다」 — 도 감사가 요구하기 전에 스스로 세운 것이다.

이 SPEC의 남은 결함은 정직성의 결함이 아니다. **자기가 만든 규율을 자기 문서의 마지막 두 자리에 적용하지 않은 것**이며, 그 두 자리가 공교롭게도 이 카드가 고치려는 결함과 정확히 같은 모양이다.
