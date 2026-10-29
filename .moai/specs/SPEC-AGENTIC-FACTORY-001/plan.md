# SPEC-AGENTIC-FACTORY-001 — 구현 계획 (0.6.1)

순서는 바뀔 가능성이 큰 결정부터 둔다(데이터 모델·권위, effect 경계, 생명주기). 기계적인 단계(CLI·template·행렬)는 뒤에 둔다. 시간 예측은 쓰지 않는다.

## §A 맥락

- 규범 원천: Issue #1948 본문과 댓글 11개(가장 새 revision 우선). spec.md §A.1.
- 작성 트리: worktree `af-contract-spec`, 브랜치 `WT-agentic-factory-contract`, HEAD `3275706cc9`.
- 산출물: Tier L 다섯 개(spec.md, plan.md, acceptance.md, design.md, research.md) + progress.md.
- 하위 SPEC: SPEC-AGENTIC-FACTORY-IDENTITY-001(역할 부여·세션 신원·요청자 계보). 이 SPEC의 M7은 하위 SPEC 완료에 묶인다.
- 레인 소유: design §10(lane-17 설정, lane-18 core, lane-19 template, lane-20 effect, lane-21 브라우저, lane-22 시험).

## §B 진행 원칙 — 완료까지 지속 개선 (정본 [9] S52, DEC-AF-10)

이 원칙은 REQ가 아니다(REQ로 두면 26이 되어 상한을 넘는다). 모든 마일스톤에 적용한다.

1. 개선 루프: 실제 관찰 → 오류·병목 재현과 Issue 기록 → 근본 원인·영향 범위·불변식 분석 → 최소 범위 근본 개선 → 회귀 시험 축적 → 같은 변경의 증거 확인과 전후 비교 누적 → 설계·Issue·구현 정합 갱신.
2. 금지: 증상만 가리는 부분 패치, 근거 없는 상한 증가, 무한 수리·감사 반복.
3. 새 방식이 낫다는 결론을 측정 전에 내리지 않는다. 측정 기록은 design §4.7 어휘를 쓴다(REQ-AF-025, AC-AF-024).
4. 임계값·표본 설계는 `proposed`와 `approved`를 구분하며 승인은 운영자 몫이다(OD-1).
5. 기록 행위 자체의 비용(장부 갱신, 댓글, checkpoint)도 측정 범위에 넣는다.

## §C 알려진 문제

- control ref는 genesis뿐이다. 운영 쓰기 연결(CAS, fencing)은 lane-18 구현을 기다린다.
- 설정 키 → `core.Collaboration` 연결이 없다(lane-18, lane-17 읽기 코드 사용).
- 접수 상태 7행은 lane-18이 구현했다고 보고했으나(미검증), lane-20 표시 이름 6개는 정렬 중이다.
- 보호 약화 거절 경로가 코드에 없다(lane-22 보고). AC-AF-008의 거절 시험은 NOT_RUN.
- 잠금 상수·귀속 원천이 코드에 없다(lane-22 보고). AC-AF-011은 NOT_RUN.
- 충돌 PR: #1937·#1944(설정), #1943·#1916·#1921(factory·launcher). 착지 순서 후보는 design §9.5.
- 미결 운영자 결정 OD-1~OD-6(spec.md §G.1).

## §D 사전 점검 (run 진입 전)

1. `git ls-remote origin refs/heads/moai-control`으로 control ref가 genesis 또는 그 뒤 커밋인지 확인(acceptance EV-CTRL-REF).
2. `go test ./internal/agenticfactory/... -count=1`로 대상 부재 참고 관측 재측정(EV-REF-0, RED 아님).
3. AC마다 시험을 먼저 쓰고 의도한 단언 때문의 `EXPECTED_RED`를 기록(acceptance §A 두 칸 채택).
4. 관련 레인의 브랜치 사실(lane-17 `5c278d8b51`, lane-18 `da7f3a4290`·`f9f44e5ecc`, lane-22 `3c7db03d7b`)을 읽어 §B.2 보고 사실을 확인. 확인 전에는 미검증으로 둔다.
5. `pr_queue.wip_limit` 미설정 동작(OD-7)과 checkpoint 상한 미설정 동작(OD-8)의 운영자 결정. 결정 전에는 두 동작을 구현하지 않는다.

## §E 제약

- 강제 종료, 터미널 종료, 작업 트리 삭제, 미커밋 변경 폐기 금지.
- 기존 PR 닫기, 브랜치 덮어쓰기, 카드 일괄 done 금지.
- GLM managed 금지. Codex는 `workflow.agentic_factory.codex.enabled: true`일 때만.
- 검사 삭제·보호 약화·관리자 우회 금지.
- 남의 Issue 변경 금지(협업 옵션 기록은 승인이 아님).
- checkpoint 경로의 금지 사항(design §4.6) 준수.
- 숫자 설정값(OQ-22)을 지어내지 않는다.

## §F 마일스톤 (바뀔 가능성 큰 순)

### M1 — 권위·데이터 모델·claim [High, lane-18]

- control ref git 구현(CAS 쓰기, readback, 재구축), 로컬 캐시는 재구축 전용.
- 상태 모형(READY·BUILDING·REVIEWING·MERGING·DONE·CANCELLED + PARKED), E01~E27 전이표, 닫힌 사유 집합(design §2.1a).
- claim·attempt·epoch fencing, 활성 판정(`homestate.ClassifyOwnerWith`).
- record(design §4.8): DELTA1 다섯 종.
- 초록 대상: AC-AF-001, 004, 016, 017. AC-AF-002의 권위 사례(A1~A4).

### M2 — effect 경계·조정·보호·예산·checkpoint [High, lane-20 + lane-18]

- effect intent → 전송 → receipt, 결과 불명 재조회, 전달 상태 네 단계.
- 응답 분류(design §4.5)와 예산 순서, 공급자 기본값은 설정 스키마로.
- 보호 약화 일곱 연산 거절(design §11.3, lane-20 집행, lane-18 사유 타입).
- Issue 쓰기 판정 연결: lane-17 설정 읽기 → `core.Collaboration` 주입 → lane-20 effect 경계.
- checkpoint 네 단계와 상태(design §4.6). 상태·git·manifest는 lane-18, Issue 링크는 lane-20, 충돌 회귀·측정은 lane-22.
- 초록 대상: AC-AF-002(checkpoint 사례 포함), 003, 005, 008, 014.

### M3 — 접수·인계·DONE [High, lane-18 + lane-20]

- 접수·준비 7행(design §4.2), 준비 receipt를 control ref에.
- PR 인계 E03, 수리 E06, E10 → E06, PR 적체 기록.
- DONE 순서(E12 → Issue 종료 → E13 DONE, E14 PARKED `pending-effect`), 로컬 동기화 digest.
- 초록 대상: AC-AF-006, 009, 012.

### M4 — 병합 게이트·독립성·잠금 [High, lane-18 + lane-20]

- 병합 게이트(design §11)와 단일 사유 표(design §2.1b). 게이트 단위 시험에서만 귀속 대역 허용.
- 역할 쓰기 경계와 Issue 단위 독립성.
- **잠금 도입**: 컴파일 시점 상수 하나, 설정·환경·파일 입력 없음, 시나리오 하네스 사전 점검(design §11.2).
- `pr_queue.wip_limit` 검증과 배차 정지.
- 초록 대상: AC-AF-007, 010, 011(잠긴 형태).

### M5 — 장부·정제·관찰·효율 [Medium, lane-20 + lane-22]

- Issue 맥락 장부, 접근성 분류, STALE 계기, 증거 단계(`.moai/evidence/issue-<N>/`).
- 정제 C1~C5와 결정적 마지막 관문, `RESTRICTED`.
- 관찰 보기(준비 상태·PR 적체 구분, 읽기 전용).
- 효율 기록 형식과 checkpoint 비용 측정.
- 초록 대상: AC-AF-013, 018, 019, 024.

### M6 — 설정·브라우저·CLI·template·행렬·공급자 [Medium, 기계적]

- **첫 커밋은 호환 golden 파일만**(design §9.6). S1 등록은 그 다음 커밋부터.
- 설정 검증(design §6), 브라우저 3선택.
- CLI(design §9.0), OD-2 결정을 반영. 결정 전에는 버전 선언 없이 새 문법을 받지 않는다.
- template 참조·`not supported` 보고·댓글 묶음.
- 환경 행렬.
- Codex 관문과 이관 단계 기록(§F.3, design §9.7).
- 초록 대상: AC-AF-015, 020, 021, 022, 023.

### M7 — 잠금 해제와 실제 Issue 시나리오 [High, 운영자 승인 필요]

- 선행: 하위 SPEC `completed`와 그 AC 전부 PASS, 하위 OQ-AFI-1 운영자 위험 수용 기록(OD-4), 시나리오 keep-set 승인(OD-6).
- 잠금 상수 한 곳을 바꾸는 코드 변경 하나(lane-18). 같은 착지에서 AC-AF-011 L1·L2를 해제 뒤 형태로 다시 쓴다(manager-spec 재위임).
- 시나리오 묶음 단일 실행(acceptance §A.1).
- 초록 대상: AC-AF-025, AC-AF-011 해제 뒤 형태.

### F.3 기존 Codex 기본값의 전환 순서 (지금은 끄지 않음)

1. 관문 도입: `workflow.agentic_factory.codex.enabled` 관문. 기존 `workflow.codex.*`, `audit.model: codex`, `audit.gates.codex`는 그대로.
2. 미설정 0건 검증: AC-AF-015 PASS 기록.
3. 기본값 변경 제안: 그 뒤에만 별도 작업으로. 기존 기능은 대체 검증 뒤에만 제거.

각 단계는 별도 커밋이며 `progress.md` §J에 CX 행으로 적는다(design §9.7).

### F.4 Phase 2 경계

남은 카드의 Issue 이관은 이 SPEC의 마일스톤이 아니다. 이관 자체는 구현 완료로 세지 않는다.

## §G 위험

| 위험 | 대응 |
|---|---|
| control ref 쓰기가 늦어 모든 전이가 막힘 | `store-unavailable`로 멈추고 로컬 캐시로 진행하지 않음. 이것은 의도한 동작 |
| 잠금 때문에 M7 전 병합이 하나도 안 됨 | 의도한 동작(DEC-AF-02). M7 전 실행은 진단 실행으로만 기록 |
| 요청별 운영자 인증이 Issue마다 사람 승인이 됨 | 정본 [0]§2와의 잔여 충돌. OD-3에서 운영자 판단 |
| 중계 경로 위험(하위 OQ-AFI-1) | 잔여 위험으로 남고 OD-4 수용 기록 없이는 M7 불가 |
| 수치 기본값과 미설정 동작이 정해지지 않아 기능이 멈추거나 정의되지 않음 | 숫자는 OQ-22, 미설정 동작은 OD-7·OD-8의 운영자 결정. 숫자를 지어내지 않음 |
| GitHub 저장소 이름 변경·이전(301) | redirect를 허용으로 보지 않음(`repository-not-allowed`) |
| 문서 확인이 WebFetch 요약뿐(uncollated) | 설정 스키마에 `collation: uncollated` 표시. 구현 때 원문 대조 |
| 시나리오 재실행마다 실제 Issue 한 쌍과 실제 병합이 듦 | 단일 실행 원칙(acceptance §A) |

## §H 참조

- `spec.md`, `design.md`, `acceptance.md`, `research.md`
- 하위 SPEC-AGENTIC-FACTORY-IDENTITY-001
- `.claude/rules/moai/workflow/factory-dispatch*.md`, `.claude/rules/local/repo-local-pr-policy.md`
