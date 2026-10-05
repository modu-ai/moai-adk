# acceptance.md — SPEC-AUTONOMY-KICKOFF-CALIB-001 (v0.2.0)

모든 AC 는 Given-When-Then 이며 명령과 기대 출력으로 판정한다. 명령은 워크트리 세션 가드를 통과하도록 **한 줄짜리 단순 명령**만 쓴다 — `git` 을 `$( )`·`<( )`·heredoc 안에 두지 않는다. `git` 이 필요한 검사는 기준 ref 를 환경 변수로 넘긴다(`MOAI_CALIB_BASE` — 본 카드의 분기점 `38148d891`; develop 흡수 후 재생성 시에는 그 시점 분기점으로 갱신).

**빈 선택은 통과가 아니다.** grep 계열 AC 는 매칭 건수를 함께 출력해 0 이 아님을 보이고, 0 이 나오면 FAIL 이다.

**AC 의 두 층.** 이 SPEC 은 측정 설계 카드라 AC 가 두 층으로 갈린다 — **[PLAN]** 플랜 페이즈에서 지금 판정 가능한 검사(설계 산출물의 존재·내용)와 **[RUN]** run-phase 측정 실행이 처음으로 대상이 되는 규율 검사(판정 명령은 run 기록 위에서 돈다). [RUN] AC 는 run 진입 전까지 `--- PENDING-RUN` 상태이며, 이는 FAIL 이 아니다.

## §A. 수용 기준

### AC-CALIB-001 — 추출 술어 실측 근거의 귀속 [PLAN] — maps REQ-CALIB-001

- **Given** 본 워크트리와 프루브 산출물
- **When** spec.md REQ-CALIB-001 이 프루브 명령·실행일·카운트(148건·148/148·87파일·오류 0)를 기록하는지 읽고, 프루브 출력 파일이 존재하는지 확인한다
- **Then** spec.md 에 실행일 2026-09-26과 네 카운트가 모두 있고, `.moai/reports/t1244/probe/probe_output_20260926.txt` 가 존재하며 그 안에 `blocks mentioning kickoff` 행이 있다

```bash
grep -c "148" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; ls .moai/reports/t1244/probe/probe_output_20260926.txt; grep "blocks mentioning kickoff" .moai/reports/t1244/probe/probe_output_20260926.txt
```

기대: 첫 숫자 1 이상, `ls` 가 경로를 출력, `blocks mentioning kickoff (header/question/label): 148` 행. 이 AC 는 프루브 **재현**이 아니라 귀속 기록의 존재를 판정한다 — 재현 명령은 research.md §3 에 있다.

### AC-CALIB-002 — 라벨 매핑의 사전 고정 [PLAN] — maps REQ-CALIB-002

- **Given** spec.md
- **When** 라벨 공간 4값·매핑 규칙·복합 응답 처리·매핑 불가 재결 규칙·동시 적중 선결 규칙의 존재를 읽는다
- **Then** `approve`·`hold`·`modify`·`other` 네 값과 `compound` 표지, 「판사 실행 전에 규칙을 추가로 정해 일괄 소급 적용」·「텍스트 순서상 첫 동사」 문장이 모두 있다

```bash
grep -c "compound" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "일괄 소급 적용" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "텍스트 순서상 첫 동사" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 세 카운트 모두 1 이상.

### AC-CALIB-003 — 번역 금지 조항 [PLAN] — maps REQ-CALIB-003

- **Given** spec.md
- **When** 번역 금지 조항의 존재와 그 근거 문장(§29 영어 표본·§30 대조군)을 읽는다
- **Then** 「번역하지 않는다」와 근거 문장이 있다

```bash
grep -c "번역하지 않는다" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 1 이상.

### AC-CALIB-004 — 파이프라인 양성 대조 명세 [PLAN] — maps REQ-CALIB-004

- **Given** spec.md
- **When** 대조 부분집합 크기(20)·합격선(20/20)·실패 시 무효 규정의 존재를 읽는다
- **Then** 세 요소가 모두 있다

```bash
grep -c "20/20" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 1 이상.

### AC-CALIB-005 — 판사 양성 대조와 양팔 무효 규칙 [PLAN] — maps REQ-CALIB-005

- **Given** spec.md
- **When** 응답-포함 대조 5건·판사 팔 실패 시 무효·양팔 무효 규칙의 존재를 읽는다
- **Then** 세 요소가 모두 있다

```bash
grep -c "어느 하나라도 실패하면 실행 전체가 무효" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 1 이상.

### AC-CALIB-006 — 상수 기준선 동시 보고 의무 [PLAN] — maps REQ-CALIB-006

- **Given** spec.md
- **When** always-approve·always-hold 정의와 「단독으로 보고하지 않는다」 조항을 읽는다
- **Then** 둘 다 있다

```bash
grep -c "always-approve" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "단독으로 보고하지 않는다" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 두 카운트 모두 1 이상.

### AC-CALIB-007 — 판정 기준 수치의 사전 등록 [PLAN] — maps REQ-CALIB-007

- **Given** spec.md
- **When** 밴드 자격 조건 세 수치(n≥20, +10%p, ≤10%)·최저 밴드 채택·무자격 시 기각 권고·측정 후 기준 변경 금지의 존재를 읽는다
- **Then** 다섯 요소가 모두 있다

```bash
grep -c "n ≥ 20" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "+ 10%p" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "측정 후 기준" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 세 카운트 모두 1 이상.

### AC-CALIB-008 — LIVE 상한의 사전 선언 명세 [PLAN] — maps REQ-CALIB-008

- **Given** spec.md
- **When** 항목당 호출 상한(2회)·총 상한 산식·벽시계 상한(8시간)·run-record 경로·판사 경로의 미추적 명기·fail-open 처리의 존재를 읽는다
- **Then** 여섯 요소가 모두 있다

```bash
grep -c "벽시계 상한 = 배치당 8시간" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "git 미추적 로컬 파일" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 두 카운트 모두 1 이상.

### AC-CALIB-009 — run 진입 전제의 검증 가능성과 게이트 처분 [PLAN] — maps REQ-CALIB-009

- **Given** 본 워크트리와 spec.md
- **When** A1 병합 확인 명령을 지금 실행하고, 게이트 처분 문언(운영자 결정 2026-09-26의 기록과 종전 전제의 대체)을 읽는다
- **Then** exit code 가 0 이고, spec.md 가 자율 Kickoff 결정을 REQ-CALIB-012 와 연결해 기록한다

```bash
git merge-base --is-ancestor e4ea8eb05 HEAD; echo $?
grep -c "모든 킥오프 승인은 자율적으로 진행한다" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: `0`과 1 이상. 병합 확인이 1을 내면 run 전제 미충족으로 progress.md §E.2 에 기록한다 — 그때도 이 AC 는 「명령이 전제 상태를 정확히 보고한다」는 점에서 자체로는 유효하다. (b) 의 대체 전(0.1.x) 문언 「운영자가 레인 창에서 직접 승인한다」는 HISTORY 0.2.0 에서 대체 경위가 남아 있다.

### AC-CALIB-010 — t943 비재실행·인용 규율·범위 무결 [PLAN] — maps REQ-CALIB-010

- **Given** spec.md와 본 카드의 커밋 집합
- **When** 비재실행 문장·§30 상수 기준선 동반 인용 규칙·적용처 좁힘(REQ-GR-013 한 곳) 문장의 존재를 읽고, 본 카드가 `internal/`·`internal/template/` 아래를 바꾸지 않았는지 확인한다
- **Then** 세 문장이 있고 변경 파일 목록에 그 경로들이 없다

```bash
grep -c "재실행하지 않는다" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "상시 상수 기준선" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 두 카운트 모두 1 이상. 파일 범위는 분기점 기준으로(감사 D11 — 구체값 지정; 감사가 `38148d891` 로 0파일 독립 검증함):

```bash
MOAI_CALIB_BASE=38148d891 git diff --name-only "$MOAI_CALIB_BASE" HEAD -- internal/ internal/template/ | wc -l
```

기대: `0`.

### AC-CALIB-011 — 아웃바운드 스크럽 명세의 존재 [PLAN] — maps REQ-CALIB-011

- **Given** spec.md
- **When** 스크럽 대상 네 부류(키 형태 문자열·settings.local 계열 값·절대 경로·고객 데이터), 데이터 최소화(전사본 문맥 통째·세션 덤프 금지), 전송 전 기계 deny-스캔, 더미 비밀 양성 대조, 적중 시 fail-closed 차단의 존재를 읽는다
- **Then** 다섯 요소가 모두 있다

```bash
grep -c "AKIA" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "세션 덤프는 결코 보내지 않는다" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "발화한 적 없는 스캐너" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "스크럽 없는 전송은 없다" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "하한 패턴 집합" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 다섯 카운트 모두 1 이상.

### AC-CALIB-012 — 스크럽 이행: 양성 대조와 페이로드별 스캔 [RUN] — maps REQ-CALIB-011

- **Given** run 기록의 스크럽 기록(더미 비밀 양성 대조 결과, 페이로드별 `payload_id`+`scan: clean|hit` 행)
- **When** 양성 대조(더미 비밀 심기 → 스캐너 발화 확인) 기록이 첫 코퍼스 페이로드 전송보다 앞서는지, 페이로드별 스캔 행이 전부 존재하는지, 적중(`hit`) 상태로 전송된 페이로드가 0건인지 센다
- **Then** 세 가지가 모두 성립한다 — 하나라도 빠지면 [RUN] FAIL 이고 어떤 판사 수치도 판정 근거가 되지 못한다

```bash
grep -c "positive-control" .moai/reports/t1244/run-record.md; grep -c "scan: hit" .moai/reports/t1244/run-record.md
```

기대: `positive-control` 1 이상(첫 전송 이전 기록)이고, run 기록의 `scan: hit` 행마다 `blocked` 표지가 따르며 `blocked 없이 sent` 인 행은 0건이다. 이 AC 는 명세가 아니라 **이행**을 판정하므로 run 기록이 생기기 전에는 `--- PENDING-RUN` 이다.

### AC-CALIB-013 — 사전 등록 커밋이 첫 판사 호출보다 앞선다 [RUN] — maps REQ-CALIB-012

- **Given** run-record.md 의 `criteria_commit:` 필드(카드 브랜치 커밋 해시)와 `runs/` 로그의 첫 판사 호출 행(타임스탬프 포함)
- **When** 그 커밋이 카드 브랜치에 존재하고 커밋 시각이 첫 판사 호출 타임스탬프보다 앞서는지 본다
- **Then** 커밋 존재(exit 0)이고 커밋 epoch 시각 < 첫 호출 타임스탬프다

```bash
git cat-file -e "<criteria_commit>"^{commit}; echo $?
git log -1 --format=%ct "<criteria_commit>"
```

`<criteria_commit>` 은 run-record.md 에서 읽은 해시로 치환해 실행한다. 첫 판사 호출 타임스탬프는 `runs/` 첫 행의 `ts` 필드(epoch 초)로 읽어 수 비교한다. 해시가 없거나 비교가 성립하지 않으면 [RUN] FAIL — 사후 기준 변경과 구별할 수 없는 실행이기 때문이다.

### AC-CALIB-014 — 세 상한이 첫 호출 전에 선언됐고 안에서 집행됐다 [RUN] — maps REQ-CALIB-012

- **Given** run-record.md 의 `turn_cap:`·`call_cap:`·`wall_clock_cap:` 세 필드와 각각의 `declared_at`, `runs/` 첫 판사 호출 타임스탬프, 실행 종료 시점의 실제 호출 수와 종료 시각
- **When** 세 `declared_at` 이 각각 첫 호출 타임스탬프보다 앞서는지, 실제 호출 수 ≤ `call_cap`, 소요 시간 ≤ `wall_clock_cap`, 항목별 호출 ≤ `turn_cap` 을 센다
- **Then** 모두 성립한다 — 하나라도 넘었으면 넘은 지점까지의 수치만 미측정으로 기록하고 판정서에 그 사실을 적는다

```bash
grep -c "declared_at" .moai/reports/t1244/run-record.md
```

기대: `declared_at` 3 이상(세 상한 각각). 이 AC 도 **이행**을 판정하므로 run 기록이 생기기 전에는 `--- PENDING-RUN` 이다.

## §B. run-phase 판정으로 미루어지는 검사

REQ-CALIB-001(스냅샷 기록)·004(대조 실제 통과)·005(판사 대조 실제 통과)·006(기준선 실제 산출)·007(판정의 실제 적용)·008(상한의 실제 선언)·011(스크럽 이행 — AC-CALIB-012)·012(자율 Kickoff 세 조건 이행 — AC-CALIB-013·014)의 **이행**은 run 기록과 판정서를 대상으로 run-phase 에 판정한다. 플랜 페이즈 산출물은 그 규율이 사전에 문서로 고정돼 있다는 것까지를 담보한다.
