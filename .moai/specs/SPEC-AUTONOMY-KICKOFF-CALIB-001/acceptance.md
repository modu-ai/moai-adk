# acceptance.md — SPEC-AUTONOMY-KICKOFF-CALIB-001 (v0.1.0)

모든 AC 는 Given-When-Then 이며 명령과 기대 출력으로 판정한다. 명령은 워크트리 세션 가드를 통과하도록 **한 줄짜리 단순 명령**만 쓴다 — `git` 을 `$( )`·`<( )`·heredoc 안에 두지 않는다. `git` 이 필요한 검사는 기준 ref 를 환경 변수로 넘긴다(`MOAI_CALIB_BASE` = progress.md §E.2 의 `BASE` SHA).

**빈 선택은 통과가 아니다.** grep 계열 AC 는 매칭 건수를 함께 출력해 0 이 아님을 보이고, 0 이 나오면 FAIL 이다.

**AC 의 두 층.** 이 SPEC 은 측정 설계 카드라 AC 가 두 층으로 갈린다 — **[PLAN]** 플랜 페이즈에서 지금 판정 가능한 검사(설계 산출물의 존재·내용)와 **[RUN]** run-phase 측정 실행이 처음으로 대상이 되는 규율 검사(판정 명령은 run 기록 위에서 돈다). [RUN] AC 는 run 진입 전까지 `--- PENDING-RUN` 상태이며, 이는 FAIL 이 아니다.

## §A. 수용 기준

### AC-CALIB-001 — 추출 술어 실측 근거의 귀속 [PLAN] (REQ-CALIB-001)

- **Given** 본 워크트리와 프루브 산출물
- **When** spec.md REQ-CALIB-001 이 프루브 명령·실행일·카운트(148건·148/148·87파일·오류 0)를 기록하는지 읽고, 프루브 출력 파일이 존재하는지 확인한다
- **Then** spec.md 에 실행일 2026-09-26과 네 카운트가 모두 있고, `.moai/reports/t1244/probe/probe_output_20260926.txt` 가 존재하며 그 안에 `blocks mentioning kickoff` 행이 있다

```bash
grep -c "148" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; ls .moai/reports/t1244/probe/probe_output_20260926.txt; grep "blocks mentioning kickoff" .moai/reports/t1244/probe/probe_output_20260926.txt
```

기대: 첫 숫자 1 이상, `ls` 가 경로를 출력, `blocks mentioning kickoff (header/question/label): 148` 행. 이 AC 는 프루브 **재현**이 아니라 귀속 기록의 존재를 판정한다 — 재현 명령은 research.md §3 에 있다.

### AC-CALIB-002 — 라벨 매핑의 사전 고정 [PLAN] (REQ-CALIB-002)

- **Given** spec.md
- **When** 라벨 공간 4값·매핑 규칙·복합 응답 처리·매핑 불가 재결 규칙의 존재를 읽는다
- **Then** `approve`·`hold`·`modify`·`other` 네 값과 `compound` 표지, 「판사 실행 전에 규칙을 추가로 정해 일괄 소급 적용」 문장이 모두 있다

```bash
grep -c "compound" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "일괄 소급 적용" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 두 카운트 모두 1 이상.

### AC-CALIB-003 — 번역 금지 조항 [PLAN] (REQ-CALIB-003)

- **Given** spec.md
- **When** 번역 금지 조항의 존재와 그 근거 문장(§29 영어 표본·§30 대조군)을 읽는다
- **Then** 「번역하지 않는다」와 근거 문장이 있다

```bash
grep -c "번역하지 않는다" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 1 이상.

### AC-CALIB-004 — 파이프라인 양성 대조 명세 [PLAN] (REQ-CALIB-004)

- **Given** spec.md
- **When** 대조 부분집합 크기(20)·합격선(20/20)·실패 시 무효 규정의 존재를 읽는다
- **Then** 세 요소가 모두 있다

```bash
grep -c "20/20" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 1 이상.

### AC-CALIB-005 — 판사 양성 대조와 양팔 무효 규칙 [PLAN] (REQ-CALIB-005)

- **Given** spec.md
- **When** 응답-포함 대조 5건·판사 팔 실패 시 무효·양팔 무효 규칙의 존재를 읽는다
- **Then** 세 요소가 모두 있다

```bash
grep -c "어느 하나라도 실패하면 실행 전체가 무효" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 1 이상.

### AC-CALIB-006 — 상수 기준선 동시 보고 의무 [PLAN] (REQ-CALIB-006)

- **Given** spec.md
- **When** always-approve·always-hold 정의와 「단독으로 보고하지 않는다」 조항을 읽는다
- **Then** 둘 다 있다

```bash
grep -c "always-approve" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "단독으로 보고하지 않는다" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 두 카운트 모두 1 이상.

### AC-CALIB-007 — 판정 기준 수치의 사전 등록 [PLAN] (REQ-CALIB-007)

- **Given** spec.md
- **When** 밴드 자격 조건 세 수치(n≥20, +10%p, ≤10%)·최저 밴드 채택·무자격 시 기각 권고·측정 후 기준 변경 금지의 존재를 읽는다
- **Then** 다섯 요소가 모두 있다

```bash
grep -c "n ≥ 20" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "+ 10%p" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "측정 후 기준" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 세 카운트 모두 1 이상.

### AC-CALIB-008 — LIVE 상한의 사전 선언 명세 [PLAN] (REQ-CALIB-008)

- **Given** spec.md
- **When** 항목당 호출 상한(2회)·총 상한 산식·벽시계 상한(8시간)·run-record 경로·판사 경로의 미추적 명기·fail-open 처리의 존재를 읽는다
- **Then** 여섯 요소가 모두 있다

```bash
grep -c "벽시계 상한 = 배치당 8시간" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "git 미추적 로컬 파일" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 두 카운트 모두 1 이상.

### AC-CALIB-009 — run 진입 전제의 검증 가능성 [PLAN] (REQ-CALIB-009)

- **Given** 본 워크트리
- **When** A1 병합 확인 명령을 지금 실행한다
- **Then** exit code 가 0 이다 — 전제가 이미 성립함을 본 카드가 관측했다는 뜻이며, run 진입 시 같은 명령을 다시 읽는다

```bash
git merge-base --is-ancestor e4ea8eb05 HEAD; echo $?
```

기대: `0`. 명령이 1을 내면 run 전제 미충족으로 progress.md §E.2 에 기록한다 — 그때도 이 AC 는 「명령이 전제 상태를 정확히 보고한다」는 점에서 자체로는 유효하다.

### AC-CALIB-010 — t943 비재실행·인용 규율·범위 무결 [PLAN] (REQ-CALIB-010)

- **Given** spec.md와 본 카드의 커밋 집합
- **When** 비재실행 문장·§30 상수 기준선 동반 인용 규칙·적용처 좁힘(REQ-GR-013 한 곳) 문장의 존재를 읽고, 본 카드가 `internal/`·`internal/template/` 아래를 바꾸지 않았는지 확인한다
- **Then** 세 문장이 있고 변경 파일 목록에 그 경로들이 없다

```bash
grep -c "재실행하지 않는다" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md; grep -c "상시 상수 기준선" .moai/specs/SPEC-AUTONOMY-KICKOFF-CALIB-001/spec.md
```

기대: 두 카운트 모두 1 이상. 파일 범위는 `MOAI_CALIB_BASE=<BASE>` 로:

```bash
MOAI_CALIB_BASE=<BASE> git diff --name-only "$MOAI_CALIB_BASE" HEAD -- internal/ internal/template/ | wc -l
```

기대: `0`.

## §B. run-phase 판정으로 미루어지는 검사

REQ-CALIB-001(스냅샷 기록)·004(대조 실제 통과)·005(판사 대조 실제 통과)·006(기준선 실제 산출)·007(판정의 실제 적용)·008(상한의 실제 선언)·009(b. 운영자 Kickoff 승인)의 **이행**은 run 기록과 판정서를 대상으로 run-phase 에 판정한다. 플랜 페이즈 산출물은 그 규율이 사전에 문서로 고정돼 있다는 것까지를 담보한다.
