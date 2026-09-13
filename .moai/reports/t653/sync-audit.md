# t653 sync-audit 판정 (2026-09-14)

> 독립 싱크 감사 — 카드 t653 (SPEC-MOAI-GATEWAY-001, AS-4 마일스톤). 감사 대상 트리:
> worktree `.claude/worktrees/t653`, branch `WT-gateway-as4-resume`, HEAD `c71545213` + 본 싱크 커밋.
> 판정 형식: 4차원 점수(기하조화평균) + must-pass 판정.

## 종합 판정

**PASS** — 총점 **0.88** (Tier L 기준선 0.85 이상). must-pass 실패 없음. 단, 기록 의무 조항 2건(커버리지 목표 미달 고지, attribution gap)이 Gaps로 남는다.

## 감사 항목별 검증

### 1. 자동화 조항 vs 구현 (AS-010..AS-013)

- **AS-010 자동화 부분**: resume 경로·거절군 — `go test ./internal/codexbridge/ -run 'TestResume'` 계열 PASS 관측(이번 재측정, 셀렉터 그룹 10건 PASS에 포함). legacy 레코드 거절(`ErrRecovery`) 테스트로 고정.
- **AS-011 자동화 부분**: idle 전용 전환 + gateway 거절군 — `TestIdleModelChange|TestWaitingAndNewPhase` PASS + gateway 셀렉터 `TestModel|TestGPT|TestManaged|TestUnknown` 5건 PASS(이번 재측정).
- **AS-012 자동화 부분**: PostCompact exact-digest·1회 rebase·거절군 — receipt 패키지 `ok` 88.9%(재측정) + `TestPostCompact|TestCompact|TestRebase` 셀렉터 PASS.
- **AS-013 자동화 부분**: 경계 fork — `TestForkAt|TestChainTo` 6건 PASS(receipt+conversation, 재측정). 실분기·native fork는 T21/NOT-RUN 처분(아래 2항).
- RED-before-GREEN 증거가 마일스톤마다 존재(m3/m4/m5 red·green 로그 쌍) — verification-completeness §2 두 칸 규율과 일치.
- 실증 항목을 PASS로 세지 않은 점 — acceptance.md AS-012의 "단순 구조 probe나 안전한 거절만으로 전체 기능을 통과 처리하지 않는다" 조항이 준수됐다.

### 2. T21 이관 정합성

acceptance.md:469(T21) ↔ spec.md HISTORY 0.12.0 ↔ progress.md §E.3 세 표면을 직접 대조 — 이관 대상(AS-010·011·012 → t844), 비이관(AS-013), 전체 통과 보류 문구가 삼자 일치. **정합.**

### 3. §E.3/§E.4 신호 무결성

- §E.3: `run_complete_at: 2026-09-14`, `run_commit_sha: 9f3dc41e0` — 존재 확인, 커밋 `git log`와 일치.
- §E.4: `sync_commit_sha: "pending-backfill"` — schema doctrine D3 승인 플레이스홀더(spec-frontmatter-schema.md § SHA placeholder backfill exemption). 후속 커밋에서 실측 SHA 백필 대상.
- 문법 검사: `sync_status: audit-ready` + "SPEC 미종결" 명시 — 다중 카드 시리즈에서 카드 단위 싱크와 SPEC 종결을 구분한 올바른 표기.

### 4. SPEC 상태 규율

- spec.md frontmatter: `status: implemented` — **completed 아님** 확인. 본 감사가 상태 전이 자체를 검증했다: in-progress → implemented 전이만 발행됐고 umbrella 시리즈(t654+)가 남아 있어 implemented 유지가 옳다.
- 소유 규율: 전이 수행자 manager-docs, 본 싱크 커밋에 `Authored-By-Agent: manager-docs` 트레일러 동반(OwnershipTransitionRule의 WHO 신호).
- 금지 수정면: spec.md/plan.md/acceptance.md 본문 무변경 — 본 감사가 `git diff`로 확인한다(커밋 전 검증, 아래 Evidence).

### 5. run 증거 경로 해소

`.moai/reports/t653/` 15개 파일(m1~m6 로그·문서) 전원 `ls` 존재 확인 — **전 해소.**

## 4차원 점수

| 차원 | 점수 | 근거 |
|---|---|---|
| Functionality | 0.90 | 자동화 조항 전원이 재측정에서 GREEN; 실증 항목의 과대 주장 없음; T21/NOT-RUN 처분이 정확 |
| Security | 0.90 | 거절군 조밀(foreign/stale/duplicate PostCompact, 미지·영·사이클 fork 경계, 모델 권한 거절 ErrManagedAuthority/ErrUnknownModel); 신규 보안 면 없음 |
| Craft | 0.85 | lint 0·gofmt 정결·RED-GREEN 증거 완비. 단, codexbridge 커버리지 82.5~83.1%가 패키지 목표 85% 미만(환경 차단 4건의 기여분 미측정 — Gaps) |
| Consistency | 0.88 | T21 삼자 정합, §E.3/§E.4 무결, 상태 규율 준수, CHANGELOG 보류 결정 근거 기록. 단, run-phase 커버리지 skip 표현식 미기록 attribution gap |

**기하조화평균**: 4 / (1/0.90 + 1/0.90 + 1/0.85 + 1/0.88) ≈ **0.88** → PASS (Tier L 0.85 이상).

## Gaps (미검증)

- codexbridge 커버리지 85% 목표 미달(82.5~82.7% 재측정 / 83.1% run 기록) — 환경 차단 4건의 커버리지 기여분이 이 샌드박스에서 측정 불가. CI 전체 스위트가 유일한 전량 판정면.
- run-phase 83.1%의 정확한 재현 불가 — skip 표현식 미기록.
- 실증 NOT-RUN 4항목(AS-010/011/012 → t844, AS-013 전제 실패) — verdict Gaps와 동일.
- 사전 존재 환경 실패 5건의 원인 규명 — 카드 범위 밖.

## Residual-risk (잔여 위험)

- 본 감사는 카드 t653 스코프의 싱크 무결성을 판정한다. umbrella SPEC의 최종 completed 전이는 t654+ 착지 후 후속 싱크의 몫이며, 그때 본 §E.4의 `pending-backfill`이 백필됐는지 재확인이 필요하다.
- t654가 RebaseLedger `appliedEpoch`·inherited prefix의 생산 배선을 설계하지 않고 넘어가면 M4/M5의 검증된 동작이 런처에서 실현되지 않는다.
