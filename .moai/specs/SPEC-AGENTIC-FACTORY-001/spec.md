---
id: SPEC-AGENTIC-FACTORY-001
title: "AgenticFactory 실행 계약 — GitHub Issue 원장 기반 소유권·전이·안전 게이트"
version: "0.6.1"
status: draft
created: 2026-10-29
updated: 2026-10-29
author: GOOS
priority: P0
phase: "v3.3.0 target"
module: "internal/agenticfactory"
lifecycle: spec-anchored
tags: "agentic-factory, github-issues, control-ref, ownership, lease, policy, intake, context-ledger, rate-limit, checkpoint, contract"
tier: L
issue_number: 1948
related_specs: [SPEC-ASIDE-BROWSER-001, SPEC-RESOURCE-SLOT-LEASE-001, SPEC-FACTORY-LEASE-RENEW-001, SPEC-AGENTIC-FACTORY-IDENTITY-001]
---

# SPEC-AGENTIC-FACTORY-001 — AgenticFactory 실행 계약

## HISTORY

| 버전 | 날짜 | 작성 | 내용 |
|---|---|---|---|
| 0.1.0 | 2026-10-29 | manager-spec | 최초 작성 (draft). 규범 원천은 GitHub Issue #1948 하나뿐이다. 조정자 추가 요구(단일 Issue 시나리오 7종·6개 결정 사항, 세션 구조·명칭·CLI 재정의, 명칭 미확정 표시·lane-22 대응표 정렬), 입력 보고서 세 건(lane-22 갭 매트릭스·이관 대응표, lane-21 브라우저 계약), 상위 설계 권고(Agentic Factory / ControlService / Coordinator / Worker, `moai factory start·status·attach`, TTL 대 epoch 교차 점검)를 같은 판에 PROPOSAL로 반영했다. |
| 0.2.0 | 2026-10-29 | manager-spec | plan-audit iteration-1 FAIL 0.69(`.moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter1.md`)의 결함 D1~D18을 앞으로 고치는 판. **REQ ID 재사용(상한 25 유지)**: 옛 REQ-AF-002(차단·보류)는 REQ-AF-001에 합치고 REQ-AF-002는 `done` 수용 조건(원격 반영 + 로컬 동기화 + 배포 증거 또는 기록된 N/A)으로, 옛 REQ-AF-005(비밀 가림)는 REQ-AF-004에 합치고 REQ-AF-005는 관찰 모니터 표시 항목으로 다시 썼다. REQ-AF-015는 환경별 단계 실제 실행, REQ-AF-016은 3선택 선택 모델(선택·저장·다시 읽기), REQ-AF-014는 거울 쓰기 실패 규칙 하나로 정리(D2), REQ-AF-010은 활성 판정 정본을 `homestate.ClassifyOwnerWith`로 정함(D12, §F.2), REQ-AF-025는 단계별 호환 기준(D10). AC는 25개 그대로이며 시나리오 하네스 계약(D5), 증거 원장의 "대상 부재 참고 관측" 전환(D1), AC-AF-019·024 기준점(D1·D9), AC-AF-018 여섯 증거 제거 사례(D6), AC-AF-002 미전달 dispatch 사례(D7), AC-AF-020 운영 기록 점검(D8), 마일스톤 재배치(D11)를 반영했다. `factory.go:284-286`·`approve` `Use:` 문자열 사실 정정(D17), progress §E.1 기록(D18). 명칭·명령은 계속 미확정이다. |
| 0.3.0 | 2026-10-29 | manager-spec | plan-audit iteration-2 FAIL 0.69(`.moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter2.md`)의 차단 결함 R-D5·N1~N8과 선택 O1~O3·O6·O7을 앞으로 고치는 판. REQ·AC 수는 25/25 그대로이며 새 ID를 만들지 않고 기존 문장에 합쳤다. N1: 결정 D2를 고쳐 Codex가 아닌 **공급자 중립 리뷰 경로**(독립 리뷰 세션 또는 운영자 기록 리뷰)를 정의하고 M4에 둔다(§F.7, design §11.1). N2: 증거 생산자 역할 결속(REQ-AF-018·020, design §7.2·§11). N3: 로컬 동기화를 경로별 내용 digest로(REQ-AF-002 (b), design §4). N4~N7: AC-AF-018·003·007·020 사례 보강(N7은 커밋 그래프 확인). N8: §A.3과 acceptance §E에 3선택 UI 항목 미완료 명시. R-D5: acceptance §A.1에 05→06 증거 생산 단계, 분기 사례의 격리 실행, 재실행 규칙, 두 번째 Issue의 사후 상태 추가. O6: `cc -l` 리더 검증 위치(`internal/discovery/factory_discovery.go` `verifyCandidate`) 기록. ID 재사용 안내(O8): 1회차 판정 파일의 REQ-AF-002·005는 0.1.0의 뜻(차단 기록·비밀 가림)이며 지금은 REQ-AF-001·004 안에 있다. |
| 0.4.0 | 2026-10-29 | manager-spec | plan-audit iteration-3 FAIL 0.69(`.moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter3.md`) 뒤의 델타 수리 판. 운영자 결정(2026-10-29, 리더 경유 전달)을 **결정 DEC-AF-01**(§C.5.1, design §11.2)로 기록하고 REQ-AF-018·AC-AF-018에 합쳤다: 증거를 내는 역할은 ControlService만 운영자 인증 아래 부여하고, Worker는 자기에게도 남에게도 역할을 줄 수 없으며, 한 세션은 run 안에서 역할 하나만 갖고, 세션 신원은 ControlService가 관측한 pid·시작 시각·session id·epoch에 묶는다(D1). D2: 격리 사본을 독립 gitdir·인덱스의 별도 클론으로 하고 실제 트리 HEAD·인덱스·경로별 digest 불변을 단언. D3: AC-AF-018에 코드 리뷰·보안 리뷰·QA 각각의 Worker 생산 사례와 이전 attempt Worker의 자기 코드 리뷰 사례 추가. D4: AC-AF-020 넷째 검사를 커밋마다 값 비교로 바꾸고 `llm.yaml` 두 벌 추가. D5: AC-AF-007에 인덱스만·모드만 바뀐 변형 추가. D6: 보류 사유 `operator-hold`를 design §2.1a에 넣고 REQ-AF-001·§A.1을 맞춤. D7: AC-AF-003 ④가 `received`를 세우지 않게 고침. 선택 O1~O5·O7 반영(O6 S4 AC는 남김, OQ-20). 전 AC 변형 점검표(acceptance §D) 추가와 그에 따른 AC 보강. REQ·AC 수는 25/25 그대로이며 ID 집합은 바뀌지 않았다. |
| 0.5.0 | 2026-10-29 | manager-spec | plan-audit iteration-4 FAIL 0.76(`.moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter4.md`)의 분할 제안과 운영자 재결정(2026-10-29, 리더 경유)에 따른 델타 판. 역할 부여·세션 신원·운영자 인증(결정 DEC-AF-01, REQ-AF-018의 해당 조항, design §11.2와 §7.2·§11·§11.1의 해당 내용, AC-AF-018 R1~R10, OQ-19)을 하위 SPEC-AGENTIC-FACTORY-IDENTITY-001로 옮기고, 그 자리에 하위 SPEC의 귀속 조회 의존과 **잠금 결정 DEC-AF-02**(§C.5.1: 하위 SPEC 완료 전에는 리뷰·감사·QA·coordinator 증거가 모두 귀속 불가로 막힘)를 넣었다. N3: AC-AF-008을 M1에서 닿는 상태 엔진 조회 수준으로 다시 썼고, 관찰 출력·종료 코드 단언은 M6의 AC-AF-022 (e)로 옮겼다(0.6.0 정정: 이 행은 원래 "AC-AF-024"로 잘못 적혀 있었다 — iteration-5 D6). 변형 점검표(acceptance §D)의 잘못 귀속된 두 행(005 하네스 행, 018 R1/R2 행)과 이관 항목 행을 고쳤다. REQ·AC 수는 25/25 그대로다. |
| 0.6.0 | 2026-10-29 | manager-spec | 정본 정합 판. 규범 원천을 Issue 본문만에서 **Issue #1948 본문 + 댓글 11개(가장 새 revision 우선)**로 넓혔다(§A.1). plan-audit iteration-5 FAIL(`.moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter5.md`)의 D1~D6을 원인별로 고쳤다: (a) 잠금을 코드 수준 상수로 정하고 설정·환경 변수로는 풀 수 없게 하며 시나리오 하네스까지 덮음(DEC-AF-04, REQ-AF-011, AC-AF-011·025, plan M7) — D1~D3; (b) 귀속 결과 하나 → 게이트 사유 하나의 단일 표(design §2.1b) — D4; (c) writer 인계의 세션 변경·epoch 변경 분리(AC-AF-009·011) — D5; (d) 위 0.5.0 행 정정과 두 가지 셈 규칙(acceptance §A) — D6. 구조는 리더 결정 K.6 (마)를 따른다: 판단이 날카로운 항목(상태·소유·병합)은 묶음 REQ, 전이표·record·template·시험 manifest는 design 표(DEC-AF-10). **REQ·AC ID의 뜻이 바뀌었다**: REQ-AF-001~015 = 묶음 G01~G15, REQ-AF-016~025 = 표 T01~T10, AC-AF-0NN = 리더 배정표 K.6b의 AC NN. 옛 ID 대응은 research.md R12. 새로 넣은 것: 정본 상태 모형(READY·BUILDING·REVIEWING·MERGING·DONE·CANCELLED + PARKED, E01~E27), 권위 = 온라인 control ref + Issue(로컬 claim 저장소는 재구축 가능한 캐시), PR 인계·수리 단일 소유·PR 적체 상한, 백그라운드 접수·준비 7행, Issue별 맥락 장부·접근성 분류·정제 범주 C1~C5·증거 위치 단계, GitHub/GitLab 호출 효율·429/403 판별, Issue 쓰기 권한 정책과 협업 옵션, checkpoint 복구, 효율 측정 어휘, 명칭(Coordinator 내부 / leader 표시 / agent 실행자). REQ 25 / AC 25(상한과 같음). 이 판은 감사받지 않았다(새 감사 승인 없음). |
| 0.6.1 | 2026-10-29 | manager-spec | 조정자 후속 지시에 따른 정정 판(REQ·AC 수와 ID는 그대로). (1) 0.6.0이 §C.7에 예약으로 넣었던 운영자 결정 행 하나와 §H·progress의 해당 언급을 지웠다. 그 지시는 잘못 전달된 것이었다. 다른 DEC 번호는 다시 매기지 않았다(깨지는 참조 없음). (2) `pr_queue.wip_limit` 미설정 시 배차 정지와 checkpoint 상한 미설정 시 거절은 새 정책 선택이므로 채택하지 않고 §G.1 OD-7·OD-8(운영자 결정 대기)로 옮겼다. 옛 OQ-24는 OD-7로 옮겨 G.2에서 지웠고, OQ-22는 숫자 값 질문만 남겼다. design §4.6·§6, acceptance AC-AF-007, plan §D·§G를 맞췄다. (3) 닫힌 사유 문자열 8개를 lane-18 트리의 Go 상수와 바이트 단위로 대조해 design §2.1a에 결과를 적었다(8개 모두 일치, 문자열 수정 없음). (4) lane-18·lane-20 보고 사실과 lane-18이 밝힌 공백을 §B.2·§H에 넣었다. |

## §A 범위와 규범 원천

### A.1 규범 원천

이 SPEC의 규범 원천은 GitHub Issue #1948 "[P0] AgenticFactory 우선 완성"의 **본문과 댓글 11개**다. 이 실행에서 `gh issue view 1948 --comments --json comments`(읽기 전용)로 받은 사본은 댓글 11개였고, 앞서 저장한 사본(v6)과 같았다. 문장이 서로 부딪히면 **가장 새 revision이 이긴다**.

| 번호 | revision | 다루는 것 |
|---|---|---|
| [0]~[2] | S28 기본 설계 1/3~3/3 | 상태 모형 E01~E27, PARKED overlay, 권위·effect FSM, 시험 manifest M01~M56 |
| [3] | DELTA1 | 수집기·delta·record 5종(CollectorCheckpoint·SourceRevision·ConsumerDelivery·ConsumerACK·SourceTombstone), 공유 API 예산, SYNC01~12 |
| [4] | DELTA2 | Issue·PR template, 댓글 묶음, WT 시험 |
| [5] | DELTA3 | 계층·분류·설정, GH 시험 |
| [6] | DELTA4 | 레인별 소유(lane-17~22) |
| [7] | S46 `AF-AGENT-HANDOFF-S46` | PR 인계(E03), 수리 재배차(E06), PRQueue, HANDOFF01~08, CLI19~23 |
| [8] | S51 `AF-INTAKE-CONTEXT-S51` | 백그라운드 접수·준비 7행, Issue 맥락 장부, PREP01~10, CONTEXT01~06 |
| [9] | S52 `AF-EFFICIENCY-S52` | 완료까지 지속 개선, 효율 측정, EFF01~06 |
| [10] | S53 `AF-RATE-BUDGET-S53` | webhook 우선, ETag 범위, 429/403 판별, 안전 쓰기 예산, SYNC13~18 |

정본에 없는 내용은 이 SPEC의 결정(§C.7 DEC 표)과 리더·레인이 보고한 사실(§B.2)로만 채웠고, 각각 출처를 붙였다. 확인하지 못한 것은 §G·§H로 넘겼다.

### A.2 이 SPEC이 다루는 것

이 SPEC은 **계약 + 정책 + 설정 초안**이다. 코드를 쓰지 않는다. 다루는 것: 정본 상태 모형과 전이 guard, 권위·캐시·reconcile 방향, 소유권·단일 writer·fencing, effect 불명 처리, DONE 순서, 병합 게이트와 PR 적체 상한, 보호 규칙 비약화, PR 인계와 수리, 독립성과 역할 잠금, 접수·준비, 비밀·공개 검토, 인증·rate limit·예산, 공급자 관문(GLM 제외·Codex 명시 설정), Issue 쓰기 권한과 협업 옵션, checkpoint 복구, 맥락 장부와 증거 위치, 관찰 보기, 설정·브라우저, CLI, template 참조, 환경 행렬, 효율 측정 어휘.

### A.3 "구현 완료"의 정의 경계

Issue #1948은 이 문서들이 "상위 착수 원장이지 구현 완료나 실행 시작의 증명이 아니다"라고 적는다. 이 SPEC이 PASS를 받아도 AgenticFactory가 구현된 것이 아니고, Phase 2(남은 카드의 Issue 이관)를 한 것도 아니다. 정본 S51은 "병목을 줄이기 위한 실행 계약이지 **병목 0을 보장하는 주장이 아니다**"라고 적으며, 이 SPEC도 그렇다. 정본 S52에 따라 문서 작성만으로 성능 검증이나 전체 완료를 선언하지 않는다.

사람이 보는 설정 화면(브라우저 3선택 UI)의 시험은 이 SPEC 완료로 닫히지 않는다(REQ-AF-021, §I).

### A.4 명칭 (리더 결정 DEC-AF-08)

| 개념 | 내부 역할 이름 | 사용자·CLI 표시 | 비고 |
|---|---|---|---|
| LLM 조율 | `Coordinator` | `leader` | 기존 `moai cc -f` 진입 |
| 실행자 | `agent` | `agent` | "worker"는 폐기. 레거시 이름 `worker-<n>` 거절은 유지 |
| 결정적 제어·복구(LLM 없음) | `ControlService` | (상주 서비스) | 이름은 미확정(provisional) |
| 실행 단위·시도 | `run`, `attempt` | — | 미확정 |

0.5.0까지의 `Worker`는 이 판에서 모두 `agent`로 읽는다.

## §B 배경 — 사실

### B.1 이 트리에서 읽거나 실행한 것

트리: worktree `af-contract-spec`, 브랜치 `WT-agentic-factory-contract`, HEAD `3275706cc9`(이 실행 시작 때 깨끗함). 상세 명령은 research.md.

- `git ls-remote origin refs/heads/moai-control` → `ae3a4edcaa529bc927357b345051b96cd04605a1	refs/heads/moai-control`(이 실행, exit 0). control ref가 원격에 genesis 커밋으로 있다. 그 안의 `control/schema.json` sha256 `1fc23184522e104b985bc6a8429a347836aaaab4714526939d5bc9363b290640`은 **리더 보고**(원격 readback 일치)이며 이 실행은 다시 재지 않았다.
- `internal/agenticfactory/`는 이 트리에 없다(0.5.0 EV-REF-0, 트리 `5c3f0ebf27`; 이 트리에서 다시 실행하지 않음).
- 활성 판정: `homestate.ClassifyOwnerWith`(`internal/homestate/factory_run_retire.go:152-168`)는 `ProbeProcessIdentity`(`internal/homestate/profile_lease.go:210-220`)가 돌려준 fingerprint와 기록값을 **문자열 동등 비교**만 한다(`fingerprint == processStart` → live, 다르면 dead). fingerprint 원천: Linux 등 `ps -o lstart= -p <pid>` 출력 문자열(`process_fingerprint_unix.go:11-17`, 초 단위, 순서 비교 정의 없음), darwin `"%d.%06d"` 초.마이크로초(`process_fingerprint_darwin.go`), Windows 생성 시각(`process_fingerprint_windows.go`). 모두 이 실행에서 읽음(darwin·Windows는 0.5.0 이전 판에서 읽은 함수 위치를 다시 쓰지 않고 인용).
- `FactoryProcessAlive`(`internal/factory/factory_alive_unix.go:22-28`)는 PID만 본다(0.2.0 측정).
- `main` 보호 설정(0.2.0 `gh api` 측정, acceptance EV-AC008-BASE): 필수 검사 3개, `enforce_admins: true`, 강제 푸시·삭제 금지, `required_linear_history: false`.
- 이 프로젝트의 로컬 `workflow.yaml`은 Codex 기능을 켜 두었고(`audit.model: codex` 등), 코드 기본값과 배포 템플릿은 꺼 둔다(0.1.0 측정).
- (0.6.1, 트리 HEAD `89f4e812a3`) 닫힌 사유 문자열 8개를 lane-18 트리(worktree `af-core-claim-runner`, 체크아웃된 브랜치 `WT-agentic-factory-core-claim-runner`의 ref 파일 값 `fa781ef0e9`)의 `core/state.go`·`core/authz.go`와 바이트 단위로 대조했다. 그 트리에는 git을 실행하지 않았고 읽기만 했다. 같은 두 파일을 이 트리에서 `git show f9f44e5ecc:<경로>`로도 읽었다. `git merge-base --is-ancestor f9f44e5ecc fa781ef0e9`는 종료 0이었고, `git diff --stat f9f44e5ecc fa781ef0e9 -- <두 파일>`의 출력은 비어 있었다. 결과는 design §2.1a.
- (0.6.1) `git cat-file -t`로 `da7f3a4290`·`3c7db03d7b`·`f9f44e5ecc`·`e6499b731a`·`51edce26a8`·`fa781ef0e9`가 모두 `commit`임을 이 트리에서 관측했다. 앞의 세 커밋은 리더가 확인한 사실과 같다.
- lane-17 설정 읽기 코드(worktree `af-settings-keys`, `internal/config/agentic_factory.go`)를 이 실행에서 grep으로 읽었다: 키 `workflow.agentic_factory.browser.selected`·`codex.enabled`·`pr_queue.wip_limit`·`collaboration.enabled`·`collaboration.allowed_repositories`·`collaboration.permissions.{edit_body,close,labels,assignees}`, 함수 `NormalizeRepositoryPath(provider, s)`, `ResolveAgenticFactoryBrowser`, `CodexOptIn`.

### B.2 리더·레인이 보고한 사실 (이 실행이 검증하지 않음)

| 사실 | 출처 | 상태 |
|---|---|---|
| 접수 상태 7행을 core에 구현 | lane-18 커밋 `da7f3a4290`, 리더 보고 | 미검증 |
| 닫힌 사유 문자열 8개(design §2.1a의 `unclassified-403`과 표 끝 7행) | lane-18 커밋 `f9f44e5ecc`(브랜치 `WT-agentic-factory-core-claim-runner`), **lane-18 보고·리더 확인, `core/state.go`의 임시 상수** | 0.6.1에서 바이트 대조함: 8개 모두 일치(design §2.1a) |
| 커밋 `da7f3a4290`, `3c7db03d7b`, `f9f44e5ecc`가 저장소에 있다 | 'reported by lane-18 / lane-20, relayed by the leader; commit existence confirmed by the leader via cat-file' | 리더 확인. 이 실행도 `cat-file -t` → `commit` 관측 |
| `f9f44e5ecc`에서 `op-not-defined`는 `core/authz.go`에, `foreign-closing-keyword`는 `core/state.go`에 있다 | 'reported by lane-18 / lane-20, relayed by the leader; commit existence confirmed by the leader via cat-file' | 리더 확인. 이 실행의 바이트 대조: `"foreign-closing-keyword"`는 `state.go:83`에 있다. 따옴표 붙은 `"op-not-defined"`는 `state.go:86`에만 있고 `authz.go`에는 0개다. `authz.go`는 상수 이름 `ReasonOpNotDefined`(208행)로 그 값을 쓴다(design §2.1a) |
| claim 문서 종류 `collector`(Safety 등급) | 커밋 `fa781ef0e9`, 'reported by lane-18 / lane-20, relayed by the leader; commit existence confirmed by the leader via cat-file' | 내용은 미검증. 이 실행도 `cat-file -t fa781ef0e9` → `commit` 관측(lane-18 브랜치 ref 파일 값과 같음) |
| checkpoint 상태 모형 구현: `local`·`pending`·`remote-durable`·`recovered` | 커밋 `e6499b731a`, 'reported by lane-18 / lane-20, relayed by the leader; commit existence confirmed by the leader via cat-file' | 내용은 미검증. design §4.6의 일곱 상태 가운데 넷(§H) |
| 약 190 KB를 넘는 collector 스냅숏은 저장되지 않는다(lane-20 한계) | 커밋 `51edce26a8`, 'reported by lane-18 / lane-20, relayed by the leader; commit existence confirmed by the leader via cat-file' | 내용은 미검증(§H) |
| 협업 판정 규칙은 `core.Collaboration.Authorize` 한 곳. 설정 키 → `core.Collaboration{Provider, Enabled, AllowedRepositories, EditBody, Close, Labels, Assignees}` 연결은 아직 없음(주입 필드만) | lane-18 보고, 리더 전달 | 미검증, 열린 항목(§G.1) |
| 자동 종결 키워드 목록 = GitHub 문서의 아홉 낱말(close/closes/closed, fix/fixes/fixed, resolve/resolves/resolved). 대소문자 혼합·URL 형태는 문서에 없지만 안전 쪽으로 포함. 문서 확인은 WebFetch 요약이며 원문 대조 안 함(uncollated) | lane-18 보고, 리더 전달 | 미검증 |
| 설정 읽기 코드 `5c278d8b51`(브랜치 `WT-agentic-factory-settings-keys`): 리더가 독립 재검증(2파일, 깨끗한 트리, 판정 도우미 0, `internal/config` 시험 ok) | 리더 보고 | 이 실행은 grep만 함 |
| `--agent` 인자가 Claude로 그대로 넘어감(argv 수준 특성화 시험) | lane-22 커밋 `3c7db03d7b`, 리더 보고. Claude 쪽 효과는 단언하지 않음 | 미검증 |
| `TestProtectionGuard_*` e2e 시험 3개 작성됨(레이블 쓰기 endpoint 고정, 쓰기 기본 꺼짐 시 전송 0, 제품 소스의 보호 설정 용어 0). "약화 연산 일곱 개가 각각 사유 하나로 거절"은 거절 경로가 없어 NOT_RUN | lane-22 보고(리더 전달) | 미검증 |
| 종결 순서(Issue 종료 → DONE)는 lane-18 코드의 기존 시험으로 증명됨 | lane-22 보고 | 미검증, 시험 이름 미전달 |
| GitHub REST: "The name is not case sensitive."(https://docs.github.com/en/rest/repos/repos, 2026-10-29 확인, WebFetch 요약·uncollated). 저장소 이름 변경·이전 때 301 응답(uncollated) | 리더 전달 | 원문 대조 안 함 |

## §C 요구사항 (GEARS)

요구사항은 묶음 15개(G01~G15)와 표 10개(T01~T10)로 25개다(Tier L 상한 25). 표 REQ는 "design.md의 표를 지킨다"는 한 줄 계약이며 표 본문이 세부다(DEC-AF-10). 각 REQ의 소유 레인은 design §10.

### C.1 묶음 요구사항 — 상태·소유·병합

| ID | 묶음 | 요구사항 |
|---|---|---|
| REQ-AF-001 | G01 활성 claim·fencing | The claim authority SHALL hold at most one active claim and exactly one accountable writer per Issue: of concurrent claims for the same Issue exactly one SHALL succeed, every reissue, attempt, and its history SHALL stay under the same Issue key, and every state write SHALL carry the attempt epoch so that a write carrying an epoch older than the current one is refused. When a dispatch names an Issue that already has an active claim, a live attempt, or an undelivered dispatch, the dispatcher SHALL refuse it and SHALL record the claim it collided with. |
| REQ-AF-002 | G02 권위·재구축·checkpoint 복구 | (a) The online control ref (`refs/heads/moai-control`) SHALL be the authority for technical state (claim, attempt, epoch, effect intent and receipt, checkpoint manifest) and the Issue SHALL be the authority for business fields (`design.md` §4); the local claim store SHALL be a cache rebuildable from the control ref and approved immutable artifacts and SHALL NOT by itself confer a claim, a grant, or a completion. (b) When the control ref cannot be read, or a compare-and-set write against it fails, the ControlService SHALL keep the dependent transition from proceeding and SHALL NOT proceed on the local cache alone. (c) A checkpoint SHALL pass four independent steps — local commit, remote acknowledgement, control-ref manifest verification, Issue summary link — each with its own evidence and failure state (`design.md` §4.6); a checkpoint SHALL be `remote-durable` only after the remote acknowledgement and the manifest verification are both observed, a failed Issue link SHALL NOT revoke `remote-durable`, and before any push the checkpoint writer SHALL detect a ref and SHA already acknowledged and SHALL NOT push them again. (d) When a checkpoint is reused or recovered on the same or another host, the ControlService SHALL first establish that the previous writer is `dead` or recorded a safe stop and SHALL advance the epoch, and SHALL NOT start a second writer. (e) The checkpoint path SHALL NOT stage or commit the user's uncommitted changes automatically, push unfinished work to `main`, force-push or rewrite history, or carry any value of redaction categories C1–C5, and checkpoint storage SHALL stay within the configured cap and retention parameters. |
| REQ-AF-003 | G03 소유·Issue 쓰기 권한 확인 | When ownership verification fails or cannot be completed (claim absent, holder mismatch, authority unreadable), the writer SHALL refuse the write and SHALL NOT fall back to an unverified write. When AgenticFactory is about to write to an Issue, the single Issue-write decision (`design.md` §4.1) SHALL allow a body edit or a close only on an Issue whose author's immutable numeric id equals the authenticated user's numeric id, never by comparing login names; on another user's Issue it SHALL allow only related comments, plus a claim only where the collaboration option is enabled and the repository is on its allow-list, and a body edit, close, label change, or assignee change only where additionally that operation's own permission key is true; it SHALL refuse label, assignee, title, milestone, and reopen operations on the user's own Issue whatever any permission key holds; it SHALL refuse with exactly one reason of `design.md` §2.1a when an identity cannot be established, the repository is not allowed or answers with a different repository or any redirect, the operation is not permitted or not defined, a foreign owner's handoff is unconfirmed, or a PR body or commit carries a closing keyword aimed at an Issue it may not close. |
| REQ-AF-004 | G04 활성 판정·인계 | While an attempt's lease is expired and the liveness classification of its holder (`homestate.ClassifyOwnerWith` over the pid and process-start fingerprint recorded at claim time) is `live` or `indeterminate`, the ControlService SHALL refuse a takeover and SHALL mark the attempt `lapsed`; a new attempt SHALL be issued only when the classification is `dead` or the holder recorded a safe stop, with the epoch incremented and the prior attempt's worktree, branch, SHA, and dirty state preserved and referenced. When a session disconnects, a lease expires, the ControlService restarts, a stale HEAD result arrives, or a step fails partway, the recovery path SHALL first query the actual execution state and SHALL then resume, reissue, or park; it SHALL NOT force-kill a process and SHALL NOT start a duplicate attempt. |
| REQ-AF-005 | G05 effect 불명 재조회·전달 상태 | The delivery tracker SHALL record `sent`, `received`, `executed`, and `completed` as four separate states per command, each set only from its own evidence source (`design.md` §3); a send result or an idle notice SHALL NOT set a later state. When the outcome of an external effect (Issue creation, comment, branch push, merge request, close) is unknown, the effect executor SHALL have recorded the effect intent before sending, SHALL re-query the remote fact before any re-issue, SHALL NOT re-issue blindly, and SHALL keep the item `EFFECT_UNKNOWN` or `MERGING` with `PARKED unknown-effect` until the fact is established. |
| REQ-AF-006 | G06 DONE 순서 | When the remote integration of a merge is confirmed (E12), the state engine SHALL record `DONE` only after every completion obligation of E13 is met — acceptance evidence, merge commit, an approved local sync whose per-path content digest, index entry, and mode are unchanged for every path that was modified, staged, or untracked before the sync, a deploy record or a recorded `deploy: N/A` with its reason, and the Issue close — so that the Issue close precedes `DONE`; when the sync or the close fails it SHALL keep `MERGING` with `PARKED pending-effect` and SHALL NOT revert the merge (E14); the next dispatch for the freed capacity SHALL be admitted only after `DONE`; and migration of remaining cards SHALL NOT be recorded as `DONE` or as implementation completion. |
| REQ-AF-007 | G07 병합 게이트·PR 적체 상한 | The merge gate SHALL admit E08 only when, for the latest HEAD, every required check actually ran and passed, a same-HEAD code review and security review each carry `PASS`, the required independent audit is a debt-free `PASS`, the official QA skill verdict is `PASS`, and a valid PR handoff receipt exists (REQ-AF-009); every review, audit, QA, writer-handoff, and leader-acceptance item SHALL count only when the attribution source (REQ-AF-011) returns `attributed` with the role that evidence kind requires, and every other result SHALL be recorded through the single reason table of `design.md` §2.1b; a skipped, missing, failed, cancelled, stale, `FAIL`, `BLOCKED`, unparseable, or `PASS-WITH-DEBT` item or an exhausted audit budget SHALL be recorded as a block and SHALL NOT count as a pass. While open PR handoffs plus the repair backlog of a repository are at or above `workflow.agentic_factory.pr_queue.wip_limit`, the dispatcher SHALL dispatch no new Issue while closure and repair continue; the key SHALL be an integer of at least 1, and `0`, a negative, a non-integer, or an out-of-range value SHALL be refused, never read as "off". |
| REQ-AF-008 | G08 보호 비약화 | AgenticFactory SHALL NOT delete a required check, weaken branch protection, change the repository's linear-history setting, force-push `main`, delete a protected branch, merge through an administrator bypass, or change a protection rule through any label or settings endpoint; the effect boundary SHALL refuse each such operation (`design.md` §11.3) with exactly one recorded reason before any request is sent. |
| REQ-AF-009 | G09 PR 인계·수리 단일 소유 | When an agent publishes a PR, the ControlService SHALL move the Issue to `REVIEWING` (E03) only on a PR handoff receipt carrying the remote PR URL, remote HEAD and base, adopted SPEC, task results, check state, the original writer's provenance, the preserved-change and quiescence evidence, and the leader's acceptance, and SHALL then return the agent's claim and slot while preserving its worktree, branch, and changes; PR publication alone SHALL NOT be recorded as done. When a repair is decided (E06) or a merging Issue's HEAD or protection changes (E10 → E06), the repair SHALL be dispatched as a new attempt under an exclusive claim on the original Issue and branch, keeping Issue, branch, worktree, and epoch continuity, to an available agent, and SHALL NOT overwrite that agent's other Issue. |
| REQ-AF-010 | G10 독립성·역할 쓰기 경계 | The ControlService SHALL be the only role that writes claim, lease, liveness, reconciliation, and resume-decision state; the Coordinator SHALL act on that state only through ControlService commands; an agent SHALL write only within its own attempt. No review, security-review, audit, or QA evidence SHALL count when its producer was the writer of any attempt on the same Issue, or when the Coordinator implemented the change it would judge. |
| REQ-AF-011 | G11 역할 귀속·잠금 | The merge gate SHALL judge every evidence item through the read-only attribution query of SPEC-AGENTIC-FACTORY-IDENTITY-001 (REQ-AFI-007) and never through a value the item states about itself. Until the lock is lifted, the production attribution source SHALL be a fixed implementation whose behaviour is selected by a single compile-time constant and by no configuration key, environment variable, file, or flag, and SHALL return `unattributable` with reason `identity-spec-incomplete` for every evidence kind, so that the gate records `producer-not-allowed`; the lock SHALL cover every caller including the scenario harness, attribution stubs SHALL be confined to gate unit tests, and the lock SHALL be lifted only by a single code change that lands together with SPEC-AGENTIC-FACTORY-IDENTITY-001 `completed`, its acceptance criteria green, and the operator's recorded risk acceptance of its relay-path residual risk. |
| REQ-AF-012 | G12 접수·준비 | When a work request is accepted, the intake service SHALL record a durable request with its idempotency key, canonical payload hash, and approval scope on the control ref before acknowledging, and SHALL keep acknowledgement, Issue publication, branch registration, agent reservation, worktree preparation, READY acknowledgement, and start as separate states (`design.md` §4.2); it SHALL run preparations in a bounded-concurrency queue with backpressure, ordering, stage deadlines, and restart reconciliation of recorded requests only; it SHALL prepare a worktree only after a slot and an exclusive reservation exist, SHALL let an agent write only after the preparer ended its writes and handed over READY, and SHALL NOT create a new ledger store, unbounded subtasks, or an always-on lane. |
| REQ-AF-013 | G13 비밀·공개 검토 | Before any ledger entry, comment, checkpoint manifest, Issue summary, or refined evidence file is published, the refiner SHALL remove every value of redaction categories C1–C5 (`design.md` §4.4) with a deterministic non-LLM check as the last gate, SHALL refuse the whole publication when any category value remains, and SHALL record material that cannot be published as `RESTRICTED` with only its state, a short summary, and its SHA-256. The ledger writer SHALL refuse a write to a field by a role that is not that field's declared owner, and observation reads SHALL NOT confer write authority. |
| REQ-AF-014 | G14 인증·rate limit·예산 | When a GitHub or GitLab response is 401, 403, 404, or 429, the API client SHALL classify it by the table of `design.md` §4.5 before any retry decision, recording a 403 that matches neither a rate-limit row nor a permission row as `unclassified-403` with zero retries; it SHALL NOT send a request against a rate-limited resource before the time the response names, SHALL take every limit, remaining count, and reset time only from response headers and record `UNKNOWN` when they are absent, SHALL keep no numeric limit as a code constant, SHALL NOT switch key, token, account, app installation, or endpoint to obtain the same data or effect, and SHALL send state-safety writes before safety notices and safety notices before progress or report comments while the shared budget is below its reserve; an operation that would add a server, credential, scope, permission, or cost SHALL wait for a separate operator approval. |
| REQ-AF-015 | G15 공급자 관문 | AgenticFactory implementation work and local-environment verification SHALL NOT invoke a GLM-managed backend. Where `workflow.agentic_factory.codex.enabled` is not explicitly `true`, AgenticFactory SHALL make zero Codex calls. The Codex default migration SHALL follow the order of `plan.md` §F.3 — introduce the opt-in gate, verify zero calls while unset, then change defaults — and no existing Codex feature SHALL be removed before its replacement is verified. |

### C.2 표 요구사항 — design.md 표가 본문

| ID | 표 | 요구사항 |
|---|---|---|
| REQ-AF-016 | T01 전이표 | The state engine SHALL apply a transition only when it appears in the transition table E01–E27 of `design.md` §2.2 with its guard evidence present and the requester is that edge's writer, SHALL refuse every other request with the prior state unchanged, and SHALL record each `PARKED` overlay with exactly one reason of the closed set of `design.md` §2.1a, its evidence, and its resume condition. |
| REQ-AF-017 | T02 record | The ControlService SHALL keep the records of `design.md` §4 (CollectorCheckpoint, SourceRevision, ConsumerDelivery, ConsumerACK, SourceTombstone, PRQueue, PRHandoff, effect intent and receipt) with their declared fields, run exactly one active collector per repository, and give no agent a polling path of its own. |
| REQ-AF-018 | T03 준비 상태표 | The intake service SHALL move a request only along the seven-row preparation table of `design.md` §4.2, and E01 SHALL require that request's preparation receipt and the reserved agent's READY acknowledgement. |
| REQ-AF-019 | T04 맥락 장부·증거 위치 | The context ledger SHALL keep the entry fields, stale rule, and accessibility classes of `design.md` §4.3, SHALL link evidence only by full-commit-SHA URLs, SHALL mark an invalidated entry `STALE` without deleting it, and SHALL publish refined evidence only under `.moai/evidence/issue-<N>/` through the stages `PREPARED → MASKED → COMMITTED → PUSHED → REMOTE_VERIFIED`, reporting `AVAILABLE` only after a remote lookup. |
| REQ-AF-020 | T05 관찰 보기 | The observation view SHALL show, per project, Issue, and owner, the items of `design.md` §8.1, each with its source and read time, SHALL show the preparation states and the PR backlog distinctly from started work, and running it SHALL change no state. |
| REQ-AF-021 | T06 설정·브라우저 | The configuration loader SHALL refuse any `workflow.agentic_factory.*` value outside the rules of `design.md` §6, naming the key, the allowed range, and the file, without clamping; a read error SHALL NOT be treated as an unset key. The browser selection SHALL offer exactly `egolite`, `aside`, and `chrome`, SHALL use `chrome` only while nothing is selected, SHALL invoke only an explicitly selected browser, and SHALL block without falling back when it is not installed or cannot connect. |
| REQ-AF-022 | T07 CLI | The launcher SHALL follow the command table of `design.md` §9.0 and §8: existing entry commands SHALL keep their behaviour through the deprecation stages of §9.4, a new syntax SHALL require an explicit version declaration and refuse an undeclared mix with a compatibility error, and no argument SHALL silently change a session's role. |
| REQ-AF-023 | T08 template·분류·설정 | Issue and PR templates, taxonomy, and setup SHALL follow the reference table of `design.md` §16, SHALL report a provider capability the provider lacks as `not supported`, and SHALL batch non-urgent progress comments per its window rule without delaying security failures, approvals, or actual blocks. |
| REQ-AF-024 | T09 환경 행렬 | The environment matrix SHALL record for each environment of `design.md` §5.1 the actually executed command, exit code, and output excerpt per step, SHALL record `PASS` only when every step ran and succeeded, and SHALL record a not-run step as `not-run`, never as success. |
| REQ-AF-025 | T10 효율 측정 어휘 | Every efficiency comparison record SHALL carry the fields of `design.md` §4.7 and keep measured, estimated, `UNKNOWN`, and `NOT_RUN` values separate; the API, token, and time cost of checkpoints SHALL be recorded only as measured values or as a recorded Gap; and no comparison SHALL report an efficiency gain that hides a quality, safety, or correctness regression. |

### C.3 Issue 쓰기 권한과 협업 옵션 (REQ-AF-003의 근거)

- **Issue 쓰기 권한 정책**(DEC-AF-17, 출처 '리더 세션의 사용자 채널로 받음, 출처 표기 없음'): 본문 편집은 사용자 **자기** Issue에서만 한다. 판정은 인증된 사용자의 불변 **숫자** id와 Issue 작성자의 숫자 id가 같을 때이며, 로그인 이름은 비교하지 않는다. 남의 Issue에는 관련 댓글만 쓴다. 신원이 불분명하면 쓰기를 막는다. 따로 명시된 허락 없이 남의 본문 편집·인수·레이블/담당자 변경·종료를 하지 않는다. PR의 자동 종결 효과는 미리 점검한다. 남의 작업을 가로채는 자기 Issue 중복 발행은 금지다. 이 정책은 코드 수리를 막지 않는다. 자기 Issue 종료는 허용한다(E13/E14).
- **협업 옵션**(DEC-AF-18, 출처 '부모 대화에서 사용자가 확인, 리더 세션 경유'): 제품 기본값은 자기 Issue만 다룬다. 남의 Issue는 명시적으로 켠 협업 옵션 **그리고** 명시적 저장소 허용 목록 안에서만 claim할 수 있다. 남의 Issue의 본문 편집·종료·레이블·담당자 권한은 각각 따로이며, 옵션만으로는 아무것도 허용되지 않는다. claim 전에 인증 id, 작성자 id, 기존 owner의 인계를 확인한다. 신원 불분명이면 막는다. PR 자동 종결을 점검한다. 중복 claim을 막는다. 이 기록은 지금 남의 Issue를 바꾸라는 승인이 아니며, 새 writer·권한 확대·추가 감사를 승인하지 않는다.
- **자기 Issue의 정의되지 않은 동작**(DEC-AF-19, 리더 결정(리더 세션 경유)): 자기 Issue의 제목 편집·마일스톤 변경·다시 열기는 정본 근거가 없어 `op-not-defined`로 거절한다. 자기 Issue의 레이블·담당자 변경은 `permission-not-granted`로 거절하며, 어떤 `permissions.*` 키가 `true`여도 거절이다. `collaboration.permissions.*` 키는 **남의 Issue에 대한 쓰기에만** 적용되고 자기 Issue의 어떤 동작도 열지 않는다. 이것들은 정책 확장이 아니라 공백 기록이며(§H), 결정이 필요해지면 리더에게 보고한다.
- 판정 규칙은 한 곳(lane-18 `core.Collaboration.Authorize`)에만 있다. 설정 읽기 코드(lane-17)는 값을 읽고 검증만 한다(design §4.1 두 열 표).

### C.4 잠금 (REQ-AF-011의 근거, DEC-AF-02·DEC-AF-04)

- **무엇이 잠기나**: 귀속이 필요한 모든 증거 종류(코드 리뷰, 보안 리뷰, 독립 감사, QA, writer 인계, leader 수락). 잠긴 동안 제품 귀속 원천은 언제나 `unattributable{identity-spec-incomplete}`를 돌려주고, 게이트는 design §2.1b 표에 따라 `producer-not-allowed`를 기록한다.
- **어디에 있나**: 제품 코드의 **컴파일 시점 상수 하나**. 게이트는 그 상수만 읽는다. 설정 키, 환경 변수, 파일, 플래그 어느 것도 이 상수를 대신하거나 덮지 못한다(AC-AF-011 L3).
- **누구를 덮나**: 시나리오 하네스를 포함한 모든 호출자. 귀속 대역(stub)은 게이트 단위 시험 안에서만 허용되며, 시나리오 묶음(`af_e2e`)은 제품 원천을 쓰고 시작 때 원천 종류를 기록해 대역이면 FAIL한다(AC-AF-011 L4). 그래서 잠긴 트리에서 시나리오의 병합 단계는 초록이 될 수 없다.
- **어떻게 풀리나**: 그 상수 한 곳을 바꾸는 코드 변경 하나로만. 그 변경은 하위 SPEC `completed`, 하위 AC 전부 PASS, 하위 OQ-AFI-1 중계 경로 잔여 위험에 대한 운영자 위험 수용 기록과 **같은 착지**여야 한다(plan M7). 풀 때 AC-AF-011의 L1·L2를 "제품 원천이 등록된 증거를 귀속한다"로 다시 쓴다(acceptance §E).
- **사유 문자열** `identity-spec-incomplete`의 소유: 이 SPEC(상위)이 값을 정하고, 하위 SPEC은 같은 값을 귀속 이유 집합에 둔다.

### C.5 checkpoint 복구 (REQ-AF-002 (c)~(e)의 근거, DEC-AF-20)

- 출처: '상위 대화에서 사용자가 승인, 리더 세션 경유'. 이 SPEC은 이미 정해진 정책을 계약 문장으로 옮겼을 뿐이다.
- 단계·상태·부분 실패·manifest 필드·저장 상한은 design §4.6 표다. 소유: 상태·git·manifest = lane-18, Issue 링크 = lane-20, 충돌 회귀·측정 = lane-22.
- 묶음 G02에 넣은 이유: checkpoint는 권위(control ref)에 기술 상태를 남기고 재구축하는 일이라 G02와 같은 판단 축이다. 새 묶음을 만들면 REQ가 26이 되어 상한을 넘는다.

### C.6 정본과 부딪힌 문장의 처리

| 대상 | 0.5.0 문장 | 정본 문장(인용) | 0.6.0 처리 |
|---|---|---|---|
| 권위 저장소 | §F.3 "로컬 claim 저장소가 권위, Issue는 거울" | [3] DELTA1 "local cache는 control ref와 승인된 immutable 원본 artifact에서 재구축 가능해야 한다"; SYNC12 "local cache 소실·비LLM 서비스 재시작 → online control ref와 원본 artifact로 복구"; [8] S51 "receipt와 재개 checkpoint는 기존 온라인 control ref 권위에 두고" | REQ-AF-002 (a): control ref가 권위, 로컬은 캐시. §F.3 폐기 |
| 상태 모형 | design §2.1 `queued`…`done`, `blocked`/`on-hold` | [0]§8 READY/BUILDING/REVIEWING/MERGING/DONE/CANCELLED + PARKED, E01~E27 | design §2 전면 교체. 옛 이름 대응은 design §2.3 |
| 종결 순서 | REQ-AF-021 "`done` 기록 뒤 Issue 종료" | [0]§8 E13 "수용증거·M·승인된 local sync·close·정리 → DONE", E14 "sync/close 실패 → MERGING + PARKED pending-effect" | REQ-AF-006: 종료가 DONE보다 먼저 |
| 수리 경로 | 0.5.0 `verifying → implementing` | E06 REVIEWING → BUILDING/FIX(S46 개정), E10 MERGING → REVIEWING | REQ-AF-009: E10 → E06. 0.5.0에 없던 간선 `landing → verifying`은 E10이 메운다 |
| PR 발행 | 0.5.0에 PR 인계 개념 없음 | [7] S46 E03 "PR 원격 H/base·계약·SPEC·task 결과·검사 상태·원 writer·leader 수락 receipt·quiescence·보존 확인 → REVIEWING; … agent claim/slot 반환" | REQ-AF-009 |
| 하위 Issue 발행 | §I "하위 이슈를 만들거나 기존 이슈를 대량 발행·종료하지 않는다" | [8] S51 `ISSUE_CONFIRMED`(접수 요청의 Issue 발행) | §I 문장을 "Phase 2 이관 문맥의 대량 발행"으로 좁힘 |
| 429 판별 | REQ-AF-014 "authentication failure, permission denial, or HTTP 429" | [10] S53 SYNC13 "GitHub primary/secondary 403·429와 권한 403 → 오류 분류·정상 재개 조건 분리" | REQ-AF-014: 403도 rate limit일 수 있음, 판별표 |
| 증거 위치 | design §11 "파일 배치는 현행 관례(`.moai/reports/<id>/`)" | [8] S51 "`.moai/report`와 `.moai/reports` 중 하나라고 추정하거나 존재하지 않는 경로를 기록하지 않는다", "접근 가능한 위치가 확인되기 전에는 LINK_BLOCKED" | REQ-AF-019, DEC-AF-13: 정제본만 `.moai/evidence/issue-<N>/` |
| 접수 상태 수 | (없음) | [8] S51 준비 상태표 7행 | REQ-AF-018. lane-20 표시 이름 6개는 정렬 중인 표시 이름일 뿐이며 코드 단계 수로 쓰지 않음 |
| 역할 부여 운영자 인증 | DEC-AF-01 ② 요청마다 운영자 인증 | [0]§2 "매 카드 사람 승인을 새로 추가하거나 기존 required review를 삭제하지 않음", [0]§1 "사람은 시작 시 승인 범위를 정하고 마지막에 검수한다" | **좁게 기록**: 정본이 현재 문언보다 넓게 허용하는 부분은 적용하지 않는다. 문언 유지. 잔여 충돌은 운영자 열린 항목(§G.1) |

### C.7 결정 표

| ID | 내용 | 출처 | 반영 |
|---|---|---|---|
| DEC-AF-01 | ① 리뷰·감사 역할은 ControlService만 부여 ② 등록에 운영자 인증 필요 ③ agent는 자기에게도 남에게도 역할을 줄 수 없음 ④ 한 세션이 agent이면서 감사자일 수 없음. 정본과의 충돌 1은 좁게 기록(§C.6 마지막 행) | 운영자 결정 (2026-10-29, 리더 경유. 0.4.0 기록) | 하위 SPEC §F.1 |
| DEC-AF-02 | 하위 SPEC 완료 전 역할 증거 효력 없음(잠금) | 운영자 결정 (0.5.0 HISTORY의 운영자 재결정) | REQ-AF-011 |
| DEC-AF-04 | 잠금은 코드 수준 상수 하나, 설정·환경 변수로 풀 수 없음, 시나리오 하네스 포함, 단일 코드 변경으로만 해제, 해제 때 L1·L2 재작성 | 리더 결정 (리더 세션 경유) | REQ-AF-011, §C.4 |
| DEC-AF-05 | SPEC은 둘뿐(이 SPEC + IDENTITY-001). IDENTITY-001은 넓히지 않는다 | 리더 결정 (리더 세션 경유) | §I |
| DEC-AF-06 | 권위 = 온라인 control ref(기술) + Issue(업무). 로컬 claim 저장소는 재구축 가능한 캐시·시험 구현 | 리더 결정 (리더 세션 경유) | REQ-AF-002 |
| DEC-AF-07 | Issue 종료가 DONE보다 먼저(E13/E14, 시험 M34). 수리 E10 → E06. Issue당 활성 claim 최대 1, 이력은 같은 키. PR 적체 상한 키 `workflow.agentic_factory.pr_queue.wip_limit`(1 이상, 범위 밖 거절, `0`=끔 금지) | 리더 결정 (리더 세션 경유) | REQ-AF-001·006·007·009 |
| DEC-AF-08 | 명칭: 내부 Coordinator / 표시 leader / 실행자 agent | 리더 결정 (리더 세션 경유) | §A.4 |
| DEC-AF-09 | CLI: `-f/--factory`, 인자형 `--leader <name>`, `-w/--worktree` 유지, agent 진입 `-a/--agent`, 이전 `-l/--leader` 제안 폐기. 남은 4건은 운영자 판단 대기(OD-2) | 리더 결정 (리더 세션 경유) | design §9.0 |
| DEC-AF-10 | K.6 (마): 판단 항목은 묶음 REQ, 전이표·record·template·manifest는 design 표, S52 루프는 plan 진행 원칙 | 리더 결정 (리더 세션 경유) | §C 구조, plan §B |
| DEC-AF-11 | 시험마다 주 AC 하나. 보조 근거는 별도 열이며 충족에 세지 않음 | 리더 결정 (리더 세션 경유) | design §15 |
| DEC-AF-12 | 정제 기본선 C1~C5, 게시 불가 자료는 상태·요약·hash만 | 리더 결정(기본값) (리더 세션 경유) | REQ-AF-013 |
| DEC-AF-13 | 원본은 `.moai/reports`(ignore 규칙 그대로), 공개 안전 정제본만 추적 디렉터리 `.moai/evidence/issue-<N>/`, 커밋 고정 링크, 단계 PREPARED→MASKED→COMMITTED→PUSHED→REMOTE_VERIFIED, AVAILABLE은 원격 조회 뒤 | 리더 결정 (리더 세션 경유) | REQ-AF-019 |
| DEC-AF-14 | rate limit 문서값은 provider 설정 스키마의 기본값으로만, 출처 URL·조회 날짜·"uncollated" 표시. 없으면 UNKNOWN, 보수적 처리. 코드에 숫자 금지 | 리더 결정 (리더 세션 경유) | REQ-AF-014, design §6 |
| DEC-AF-15 | 접수 상태는 정본 S51의 7행을 따름 | 리더 결정 (리더 세션 경유) | REQ-AF-018 |
| DEC-AF-16 | 닫힌 사유 8개: `unclassified-403`, `not-issue-author`, `identity-unknown`, `repository-not-allowed`, `permission-not-granted`, `owner-unconfirmed`, `foreign-closing-keyword`, `op-not-defined` | 구현 사실 — lane-18 보고·리더 확인, `core/state.go`의 임시 상수(이 실행은 검증 안 함) | design §2.1a |
| DEC-AF-17 | Issue 쓰기 권한 정책(§C.3 첫 항목) | '리더 세션의 사용자 채널로 받음, 출처 표기 없음' | REQ-AF-003 |
| DEC-AF-18 | 협업 옵션(§C.3 둘째 항목) | '부모 대화에서 사용자가 확인, 리더 세션 경유' | REQ-AF-003, design §4.1 |
| DEC-AF-19 | 자기 Issue의 제목·마일스톤·다시 열기 = `op-not-defined`, 레이블·담당자 = `permission-not-granted`, `permissions.*` 키는 남의 Issue에만 적용 | 리더 결정 (리더 세션 경유). 공백 기록 | REQ-AF-003, §H |
| DEC-AF-20 | checkpoint 복구 설계(§C.5) | '상위 대화에서 사용자가 승인, 리더 세션 경유' | REQ-AF-002 |
| DEC-AF-21 | 하위 SPEC의 정본 충돌 2·3·4를 정본 쪽으로 고침(Issue 단위 writer 배제, Coordinator를 부여 대상에서 뺌, 부여 기록은 control ref에서 재구축) | 리더 결정 (리더 세션 경유) | 하위 SPEC 0.2.0 |
| DEC-AF-22 | 약화 거절(REQ-AF-008, 0.5.0의 REQ-AF-019)은 lane-20이 effect 경계에서 집행, 사유 타입은 lane-18, 시험은 lane-22 | 리더 결정 (리더 세션 경유) | REQ-AF-008, design §10 |

## §D 수용 기준 요약

`acceptance.md`에 25개(AC-AF-001~025)가 있다. AC-AF-0NN의 주 REQ는 REQ-AF-0NN(NN = 1~15), AC-AF-016~024의 주 REQ는 각각 REQ-AF-016·017·019·020·021·022·023·024·025, AC-AF-025는 실제 Issue 시나리오이며 주 REQ는 REQ-AF-006이다. REQ-AF-018(준비 상태표)은 AC-AF-012가 보조로 덮는다. 모든 AC는 `EXPECTED_RED`가 기록되기 전까지 채택 전이다.

## §E 추적표 (REQ → AC)

| REQ | 주 AC | 보조 AC |
|---|---|---|
| REQ-AF-001 | AC-AF-001 | AC-AF-025 |
| REQ-AF-002 | AC-AF-002 | — |
| REQ-AF-003 | AC-AF-003 | — |
| REQ-AF-004 | AC-AF-004 | AC-AF-025 |
| REQ-AF-005 | AC-AF-005 | AC-AF-025 |
| REQ-AF-006 | AC-AF-006, AC-AF-025 | — |
| REQ-AF-007 | AC-AF-007 | AC-AF-009 |
| REQ-AF-008 | AC-AF-008 | — |
| REQ-AF-009 | AC-AF-009 | — |
| REQ-AF-010 | AC-AF-010 | — |
| REQ-AF-011 | AC-AF-011 | AC-AF-025 |
| REQ-AF-012 | AC-AF-012 | — |
| REQ-AF-013 | AC-AF-013 | AC-AF-018 |
| REQ-AF-014 | AC-AF-014 | — |
| REQ-AF-015 | AC-AF-015 | AC-AF-023 |
| REQ-AF-016 | AC-AF-016 | — |
| REQ-AF-017 | AC-AF-017 | — |
| REQ-AF-018 | — | AC-AF-012 |
| REQ-AF-019 | AC-AF-018 | — |
| REQ-AF-020 | AC-AF-019 | — |
| REQ-AF-021 | AC-AF-020 | — |
| REQ-AF-022 | AC-AF-021 | — |
| REQ-AF-023 | AC-AF-022 | — |
| REQ-AF-024 | AC-AF-023 | — |
| REQ-AF-025 | AC-AF-024 | — |

## §F 유지되는 결정 (0.5.0에서 이어짐)

### F.1 claim은 slot lease가 아니다

Issue claim은 `internal/factory/slot_lease.go`의 자원 slot lease가 아니다. slot lease는 선언 기한이 지나면 소유 프로세스가 살아 있어도 `--force` 없이 인계하고(REQ-RSL-006), 키가 자원 이름이라 Issue·attempt 식별자가 없다. claim은 Issue 하나에 키 하나다. #1938(죽은 run에 남은 카드)과 #1813(같은 이름의 다른 트리)의 대체 동작은 design §12.

### F.2 활성 판정 정본과 그 한계

- 정본: `homestate.ClassifyOwnerWith`. `live`와 `indeterminate`는 인계를 막고 `lapsed`로 표시, `dead` 또는 안전 정지 기록만 인계(REQ-AF-004).
- **한계(정직한 기록)**: 비교는 문자열 동등뿐이다. Linux 원천은 초 단위 `ps lstart` 문자열이라 **같은 초 안의 PID 재사용은 구분하지 못한다**. 그 위험은 epoch fencing(REQ-AF-001)과 하위 SPEC의 계보 검사가 보완한다. 더 정밀한 원천을 요구할지는 OQ-17이다.
- `FactoryProcessAlive`(PID만)는 claim 판정에 쓰지 않으며 기존 호출처도 고치지 않는다.

### F.3 (폐기) 로컬 claim 저장소 권위

0.5.0의 "로컬 claim 저장소가 권위, Issue 댓글은 거울"은 정본 DELTA1·S51과 충돌해 폐기했다(§C.6). 대체는 REQ-AF-002.

### F.4 Codex 명시 설정 키

`workflow.agentic_factory.codex.enabled`(bool, 미설정 허용). 미설정·`false`는 호출 0건, `true`만 켜짐. 기존 `workflow.codex.*`·`audit.model: codex`는 이 SPEC 단계에서 바꾸지 않고 plan §F.3 순서를 따른다. 읽기는 lane-17 `CodexOptIn`(§B.1).

### F.5 "공식 QA 스킬"

`/moai gate`(같은 HEAD의 lint·format·type·test)로 정하고, 웹·브라우저 표면을 건드릴 때만 `/moai e2e`를 붙인다. 최소 계약: 판정 HEAD 전체 SHA 기록, 결과 `PASS`/`FAIL`/`BLOCKED`, 빈 실행은 PASS 아님, 검사별 명령·종료 코드·분류 기록, 비밀 미출력. 운영자 의도 확인은 OQ-5.

### F.6 선형 이력

측정한 `main` 보호 설정을 바꾸지 않고 `main` 이력을 다시 쓰지 않는다. `required_linear_history`는 꺼져 있으며 AgenticFactory가 켜지도 기대지도 않는다. 켤지는 OQ-6.

### F.7 병합 게이트 D1~D5 (0.3.0 결정 유지)

D1 보호 규칙 불변. D2 Codex 미설정 프로젝트의 리뷰는 공급자 중립 경로(design §11.1). D3 무부채 PASS는 AgenticFactory 경로의 Issue에만, 착지한 카드를 다시 열지 않음. D4 SHA 결합 증거는 코드 후보 브랜치 밖(0.6.0: 정제본 위치는 DEC-AF-13, 원본 위치는 `.moai/reports`). D5 필수 검사 = 보호 규칙의 3개. 근거와 출처는 0.5.0 본문(커밋 `3275706cc9`의 이 파일 §F.7)과 research.md.

### F.8 카드 TTL/renew 대 지속 배정 + epoch

TTL·renew는 활성 힌트, epoch가 권위(design §14). 기한이 지난 뒤 살아 있는 소유자의 renew 처리는 OQ-11이며 그전까지 현행 거절 유지.

## §G 미결 질문

### G.1 운영자 결정 대기 (열린 채로 둠)

| ID | 결정 | 막는 것 |
|---|---|---|
| OD-1 | S52 임계값·표본 설계의 승인(정본이 "제안과 승인 상태를 구분"하라고 요구) | REQ-AF-025 비교 기준 |
| OD-2 | CLI 4건: 버전 선언 방식, 새 문법의 대상 leader 플래그 이름, `-l/--lane` 장기 처분, 기본 문법 전환 시점. **충돌 미해결**. 어느 답이든 명시적 버전 선언, 호환 오류, 역할 뒤바뀜 0·조용한 흡수 0 회귀 시험, `--` 경계가 필요하다 | REQ-AF-022 |
| OD-3 | DEC-AF-01 ② 문언과 정본 [0]§2의 잔여 충돌: 실제 운용에서는 리뷰·감사가 Issue마다 새 세션이라 요청별 운영자 인증이 카드마다 사람 승인이 된다. 원문 뜻이 "요청마다"인지 "run마다"인지와 이 충돌의 수용 여부 | 하위 REQ-AFI-002 |
| OD-4 | 하위 SPEC 중계 경로 잔여 위험(하위 OQ-AFI-1)의 운영자 위험 수용. 수용 기록 없이는 잠금이 풀리지 않는다 | REQ-AF-011 해제, plan M7 |
| OD-5 | 추가 감사 라운드(새 감사는 지금 승인되지 않았다) | 이 판의 채택 |
| OD-6 | 실제 Issue 시나리오(AC-AF-025)에 쓸 Issue 한 쌍과 보호된 `main` 실제 병합·Issue 종료 승인(keep-set) | AC-AF-025 |
| OD-7 | **대상**: `workflow.agentic_factory.pr_queue.wip_limit` 키가 설정되지 않았을 때 새 Issue 배차를 어떻게 할지. **선택지**: (가) 새 배차를 `not-configured`로 막는다. (나) 설정 단계에서 명시 값을 요구하고, 값이 없으면 factory 시작 자체를 거절한다. (다) 문서에 적은 보수적 기본값을 쓰되, 그 값이 운영자가 정하지 않은 기본값임을 설정 스키마와 관찰 보기에 표시한다. **근거**: 정본은 상한 키와 범위 검사(1 이상, `0`=끔 금지)만 정하고 미설정 동작은 정하지 않는다. 어느 쪽이든 새 정책이며, 이 SPEC은 숫자를 지어내지 않는다. (가)·(나)는 배차가 멈출 수 있고 (다)는 근거 없는 숫자가 들어간다. 결정 전에는 미설정 사례를 채택 근거로 쓰지 않는다 | REQ-AF-007, AC-AF-007 적체 사례 |
| OD-8 | **대상**: `workflow.agentic_factory.checkpoint.{max_per_issue,retention,max_manifest_bytes}`가 설정되지 않았을 때 새 checkpoint를 어떻게 할지. **선택지**: (가) 새 checkpoint를 `not-configured`로 거절한다. (나) 설정 단계에서 명시 값을 요구한다. (다) 문서에 적은 보수적 기본값을 쓰되, 운영자가 정하지 않은 기본값임을 표시한다. **근거**: 정본은 checkpoint의 단계·상태·금지 사항을 정하지만 저장 상한 값과 미설정 동작은 정하지 않는다(숫자 자체는 OQ-22). (가)·(나)는 복구 지점이 생기지 않을 수 있고, (다)는 근거 없는 숫자가 들어간다. 결정 전에는 미설정 사례를 채택 근거로 쓰지 않는다 | REQ-AF-002 (c)~(e), design §4.6 |

### G.2 그 밖의 미결 질문

| ID | 질문 | 막는 것 |
|---|---|---|
| OQ-2 | `run`의 범위와 실행 자리 `slot`이 자원 slot lease와 이름이 겹치는 문제 | REQ-AF-010·022 |
| OQ-5 | "공식 QA 스킬" 결정(§F.5)이 운영자 의도와 맞는가 | REQ-AF-007 |
| OQ-6 | `required_linear_history`를 켤 의도가 있는가, `git-strategy.yaml` live 값과 로컬 규칙의 차이 | REQ-AF-008 문구 |
| OQ-7 | egolite·aside의 연결 실패 문구·종료 코드, aside 로그인 감지 명령 미측정 | REQ-AF-021 |
| OQ-11 | 기한이 지난 뒤 살아 있는 소유자의 renew 처리 | REQ-AF-004 세부 |
| OQ-12 | 제안 `moai factory status`와 기존 명령 이름 충돌 | REQ-AF-022 |
| OQ-13 | `CI=true`에서 aside 처리, aside 권한 기본값 | REQ-AF-021 |
| OQ-15 | 필수 검사 집합 확장 여부 | REQ-AF-007 |
| OQ-17 | 같은 초 PID 재사용을 구분하는 더 정밀한 시작 시각 원천을 요구할지 | REQ-AF-004 |
| OQ-18 | 브라우저 3선택 화면 표시 | REQ-AF-021 표시 범위 |
| OQ-20 | 명령 제거 단계(S4) 판정 형식 | REQ-AF-022 S4 |
| OQ-21 | 큐 포화 시 새 접수를 받고 대기시키는지 거절하는지(정본은 "포화를 수락으로 숨기지 않는다"까지만 정함) | REQ-AF-012 |
| OQ-22 | 수치 설정값: `intake.max_concurrent`, `intake.stage_timeout`, polling 간격, backoff 배수·횟수, 예산 reserve, 댓글 묶음 window, checkpoint 상한·보존 기간. 모두 **OPEN**이며 이 SPEC은 숫자를 정하지 않는다 | REQ-AF-002·012·014·023 |
| OQ-23 | GitLab을 이 SPEC 범위에 넣을지, 정본 [1]§19 순서 6("승인된 adapter·GitLab·부하를 확대")으로 미룰지 | REQ-AF-014 GitLab 행 |
| OQ-25 | 자기 Issue 레이블·담당자 변경을 `op-not-defined`가 아닌 `permission-not-granted`로 거절하는 리더 교정은 받았다. 이 동작들을 언젠가 허용할지 | REQ-AF-003 |

해소된 0.5.0 질문: OQ-1(시나리오 Issue → OD-6로 옮김), OQ-3(lane-22 대응표 — 정본 DELTA4와 리더 배정으로 대체), OQ-4(여섯 범위 → 레인 소유 표 design §10), OQ-8(단일 owner → DEC-AF-07), OQ-9(PR #1937 → plan §F.3 순서 안에서 판정), OQ-10(환경 목록 → design §5.1), OQ-14(증거 ref → DEC-AF-13), OQ-16(`slots` 상한 → OQ-22에 합침), OQ-19(하위 OQ-AFI-1). 0.6.1에서 OQ-24(`wip_limit` 미설정 동작)는 OD-7로 옮겼다.

## §H 확인 공백

- lane-18이 밝힌 공백(리더 전달): 실제 GitHub 원격 조회와 댓글 멱등성은 가짜(fake)로만 시험했다. CONTEXT06 "새 대상의 증거가 필요하다"는 **호출자 규칙**이며 모형이 집행하지 않는다.
- lane-18 코드에 있지만 SPEC 닫힌 사유 집합에 없는 문자열: `core/authz.go:96`의 판정 값 `"unknown-operation"`(`DenyUnknownOp`). 이것은 `BlockReason`이 아니라 `AuthzVerdict` 값이며 `Reason()`이 `op-not-defined`로 바꾼다(authz.go:207-208). 판정 값 `"allow"`(authz.go:90)도 사유가 아니다. `BlockReason` 상수 26개는 모두 SPEC 집합 안에 있다. 반대로 SPEC의 PARKED 전용 사유 6개(`dependency`·`stop-unknown`·`unknown-effect`·`pending-effect`·`contract-change`·`start-unknown`)는 lane-18 코드에 없다.
- lane-18 코드의 상태 상수는 0.5.0 모형(`queued`…`on-hold`, state.go:9-17)이다. 0.6.0의 READY…PARKED 모형(design §2.1)과 다르다. 이 실행은 그 차이를 고치지 않고 기록만 한다.
- lane-18 트리에 git을 실행하지 않았으므로 그 트리의 커밋되지 않은 변경 여부는 모른다. 대조한 줄 번호는 체크아웃된 파일과 `f9f44e5ecc`의 파일에서 같았다.
- checkpoint 상태: lane-18 보고 구현(`e6499b731a`)은 넷(`local`·`pending`·`remote-durable`·`recovered`)이고 design §4.6은 일곱이다(`committed-unacked`·`acked-manifest-unverified`·`durable-link-pending`이 보고에 없음). 이 실행은 그 커밋의 내용을 읽지 않았다.
- collector 스냅숏 약 190 KB 초과 미저장(lane-20 한계, `51edce26a8`)은 설계에 대응 규칙이 없다. 그 크기에서 수집 상태가 어떻게 복구되는지 확인되지 않았다.
- 자기 Issue의 레이블·담당자·제목·마일스톤·다시 열기 동작에는 정본 근거가 없다(DEC-AF-19). 결정이 필요해지면 리더에게 보고한다.
- 설정 키 → `core.Collaboration` 연결은 아직 없다(lane-18 보고). 연결 담당은 lane-18, 읽기는 lane-17.
- 저장소 이름 정규화: GitHub 경로는 ASCII 소문자로 바꿔 비교한다(GitHub 문서 "The name is not case sensitive.", WebFetch 요약·uncollated). GitLab과 알 수 없는 공급자는 정확히 일치로 비교한다. GitLab의 대소문자 규칙은 2026-10-29에 공식 문장을 찾지 못했고, 실제 인스턴스 관측은 하지 않는다.
- 잔여 위험: GitHub은 저장소 이름 변경·이전에 301로 답한다(uncollated). 완화: 어떤 redirect도 허용으로 보지 않고 `repository-not-allowed`로 막는다(AC-AF-003 IA9·IA10).
- 자동 종결 키워드 목록의 문서 확인은 WebFetch 요약이며 원문 대조가 아니다(§B.2).
- `control/schema.json`의 sha256은 리더 보고이며 이 실행이 다시 재지 않았다.
- 레인 보고 사실(§B.2)은 하나도 이 실행이 검증하지 않았다.
- 효율 측정의 토큰 receipt 경로, 구 Factory baseline 원천은 측정하지 않았다(정본 S52 → NOT_RUN 또는 비교 불가).
- GitLab 조건부 요청(ETag)·429 헤더 누락 규칙·403 사용 범위는 공식 문서에서 확인하지 못했다(research R10).
- 이 실행 중 worktree 가드가 한 줄 반복문(SPEC ID 두 개를 한 번에 검사)을 "너무 복잡"하다고 거절했다. 각 ID를 따로 검사해 둘 다 `PASS`를 얻었다.
- 같은 가드가 `git show … | awk …` 형태의 명령(0.5.0 REQ 의미를 뽑으려던 것)을 "명령을 실행할 수 있는 awk 프로그램"이라며 거절했다. awk 없이 `git show 3275706cc9:…spec.md | grep -o -E`로 대신 읽었다(research R12).
- lint 판정 빌드(`moai-tree5`)의 커밋 표기는 빌드 때 손으로 넣은 `8dc835756`이고 트리 HEAD는 `3275706cc9`다. 표기가 판정한 트리를 가리키지 않는다(progress §E.1).
- 0.5.0 §H의 나머지 공백(PR 변경 파일 목록 미열람, `moai codex -l` 구현 미열람, Aside 공식 문서 미열람, #1926 둘째 댓글 일부만 읽음, `auto-merge.yml` 일부만 읽음)은 그대로다.

## §I 범위 밖

### Out of Scope — 별도 SPEC 분할

- 이 프로그램의 SPEC은 둘뿐이다(DEC-AF-05). 나머지는 이 SPEC의 절이다.
- 하위 SPEC-AGENTIC-FACTORY-IDENTITY-001은 넓히지 않는다. D2 완화 구현, 기준점 여러 개(D8), 응답자 인증은 범위 밖이다.

### Out of Scope — 남은 카드의 Issue 이관 (Phase 2)

- 남은 카드·열린 PR·이슈의 대사, 중복 확인, 재우선순위, 기존 owner 인계는 이 SPEC 뒤의 단계다.
- 이관 자체는 구현 완료로 세지 않는다.
- 이관 문맥에서 기존 Issue를 대량 발행·종료하지 않는다. 접수 요청의 Issue 발행(REQ-AF-012)은 이 금지에 들지 않는다.

### Out of Scope — 기존 설정·코드 수정

- `.moai/config/**`, `internal/config`, `internal/settings`, 템플릿 트리는 이 SPEC이 고치지 않는다. 설정 읽기는 lane-17의 별도 작업이다.
- `FactoryProcessAlive`와 그 호출처, SPEC-ASIDE-BROWSER-001은 고치지 않는다.

### Out of Scope — 남의 Issue 변경 승인

- 협업 옵션 기록(DEC-AF-18)은 지금 남의 Issue를 바꾸라는 승인이 아니다. 새 writer, 권한 확대, 추가 감사를 승인하지 않는다.

### Out of Scope — 기존 결함의 종료와 기존 작업의 처분

- #1938 / #1926 / #1924 / #1847 / #1813을 자동으로 닫지 않는다. 기존 PR을 닫거나 브랜치를 덮어쓰거나 카드를 일괄 done 처리하지 않는다.
- 강제 프로세스 종료, 작업 트리 삭제, 미커밋 변경 폐기는 하지 않는다.

### Out of Scope — provider template 작성과 그림

- Issue·PR template의 작성은 다른 worktree의 별도 작업이다. 이 SPEC은 참조만 한다(design §16).
- Pele SVG는 생성하지 않았다(기록된 공백).

### Out of Scope — 설정 화면(웹 UI) 표시

- 브라우저 3선택을 사람이 보는 화면에 그리는 일은 다루지 않는다(OQ-18).

### Out of Scope — 공급자·비용

- GLM managed 경로, Codex 자동 활성화, 새 키·유료 용량·추가 지출, Cloud 개발 재개, 앱 설치, webhook 수신 서버 운영은 별도 승인 대상이며 이 SPEC이 승인하지 않는다.
