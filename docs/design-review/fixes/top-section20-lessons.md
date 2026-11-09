# 상위 §20 "이전 설계에서 남기는 교훈" 표 재작성안 (DR-07 #2143)

갱신(G3 #2208): 기준 tip `86762b8e1cb4fa7a3638e859331d41d542767f14`(`origin/v2026.11.0`), 확인일 2026-11-09. 이 안은 저장소 PR로 먼저 갱신하고, Space 상위 페이지 적용은 리더가 페이지마다 사용자 확인을 받은 뒤 한다(decisions 01:2x(11-09) "rc.2 발행 확정" 행). 설계와 RC1 결정이 부딪히면 RC1 결정이 우선이다(decisions 00:4x(11-09)).

출처: review-group4(문서 줄 인용은 그대로 승계) + synthesis S7. 구조적 결함(group4 §7.3): ① 비계승 열이 문서 1·4·5가 명시적으로 막은 결함을 문서 귀속으로 적음 ② 계승·비계승 항목의 교차 귀속(5↔3 cancel·배포 SHA, 6↔1 Pele, 2↔1 고정 반복 간격, 3↔4 Claude 우선) ③ 비계승으로 보낸 lane·lease·epoch·agent-N·Issue별 PR은 RC1이 재채택 ④ 두 문서 이상이 공유하는 교훈 10건이 상위 본문에 없음.

## 행별 교체안 (현재 문구 → 교체 문구 + RC1 재채택 열)

### 행 1 — MoAI Factory 생산성 흐름(문서1 `85f9`)

| 열 | 현재(상위:2266·2270) | 교체안 |
| --- | --- | --- |
| 남길 것 | 독립 작업·통합 writer·고정 candidate, 보고만으로 CI 재시작하지 않기 | 동일 + "진행 기록 commit으로 HEAD를 움직이지 않는다"(문서1:20·1057 — 표에만 있고 본문 0건이므로 §15 stale 행에 실제 문장으로 삽입) |
| 비계승 | 작업별 PR 또는 release 단위 PR을 모든 프로젝트에 강제 | **로컬 선통합 release-branch 후보 + 단일 integration writer 모델(문서1:161) — RC1은 Issue별 PR + 에이전틱 게이트로 대체**(문서1:18의 "모든 프로젝트에 강제"라는 문장은 없음) |
| RC1 재채택 | — | ExitWorktree(keep)→main→/clear 종료 순서(decisions:90, AGENTS.local.md:33) · 완료는 증거로 읽는다(문서1:534) · 멱등 키·stable marker·늦은 epoch 거절(문서1:792 → outbox.rs) · **역행**: Issue별 PR 강제는 문서1이 뺀 모델(decisions:39, pr-contract.md) |

### 행 2 — TODO 제어 루프(문서2 `08f6`)

| 열 | 현재(상위:2280·2284) | 교체안 |
| --- | --- | --- |
| 남길 것 | 실행 용량·차단 이유·owner·증거 준비도 구분 | 동일 + **CI trigger 도달성 사전 확인(문서2:142)** + **delivered ≠ done(문서2:1311)** |
| 비계승 | 과거 SQLite/lane/cron 구조와 고정 반복 간격 | **당시 Go Factory의 SQLite 운영 DB와 card_transition.go 수리 범위(문서2:1690)로 한정** — "고정 반복 간격"은 문서1(5분 루프)의 것이며 문서2는 오히려 polling cron 증식을 막았다(문서2:1127) |
| RC1 재채택 | — | lane 구조·owner 보존+만료 자동 탈취 금지(문서2:1135 → claim.rs:22) · durable intent→dispatched→confirmed(문서2:1255 → outbox.rs) · 조회 무변경(문서2:532) |

### 행 3 — 3.2 독립 코어(문서3 `0914`)

| 열 | 현재(상위:2294·2298) | 교체안 |
| --- | --- | --- |
| 남길 것 | adapter 호환·멱등 복구·단일 owner·사용자 환경 보호 | 동일 유지 |
| 비계승 | Go 기본·Claude 우선 Codex 후속·과거 릴리스 선후 | **Go 유지 결론과 AF→rc39→3.2 선후(문서3:7·688) — 단 '측정 뒤 언어 결정' 방법(문서3:784)은 §16에 계승**; "Claude 우선 Codex 후속"은 행 4(문서4:145)로 이동 — 문서3:285는 Codex skills-only를 정반대로 비난 |
| RC1 재채택 | — | 사용자 수정 충돌 diff 반환(문서3:374) · 같은 digest no-op(문서3:373 → HA-02) · known-good rollback(문서3:415 → recovery.rs) · effect receipt(문서3:267 → outbox.rs) · **역행**: Rust 전면 채택은 문서3이 승인하지 않은 결정 |

### 행 4 — Project Plan Run Sync(문서4 `5a44`)

| 열 | 현재(상위:2308·2312) | 교체안 |
| --- | --- | --- |
| 남길 것 | producer/consumer schema·digest·상태 의미 parity | 동일 + **판정 6종 분리표(문서4:390) — 작성자≠감사자≠통합 리더** |
| 비계승 | 정적 조사 후보를 전부 실기 재현 결함으로 표현 | **당시 Go 지침 본문의 결함 목록(문서4:185)과 "AF rc.39 선행 → Claude 중심 → Codex 보류" 순서(문서4:145)** — 문서4:20·175·187·790은 정적 조사임을 스스로 세 번 밝혔다. 현재 문구는 삭제 |
| RC1 재채택 | — | not_run→passed 승격 금지 기계 검사(status-crosswalk forbidden_promotions) · 정렬 canonical digest(spec.rs) · 감사 base/head 불일치 거부(audit.rs Stale) · 검증은 argv(shell 없음, verify.rs) |

### 행 5 — provider·browser 계약(문서5 `bb3c`)

| 열 | 현재(상위:2322·2326) | 교체안 |
| --- | --- | --- |
| 남길 것 | DOM/AX/screenshot/trace/cancel별 capability, 배포 SHA 결합 | **host별 plugin·skill·hook·MCP 계약 차이표(문서5:235) + documented/runtime_verified/unverified/unsupported 4분류(문서5:523) + browser capability 행렬(문서5:552)** — cancel·배포 SHA는 행 3(문서3:929·968)으로 이동 |
| 비계승 | 문서 지원을 실기 성공으로 표시, 특정 ADE를 필수 제품화 | **Codex alpha CLI 문법과 2026-11-01 시점 URL을 현행 계약으로 재사용하지 않기(문서5:223·476)** — 문서5는 "공식 확인 ≠ 실행 성공"(512)과 "두 제품 모두 검증 어댑터 아님"(523)으로 스스로 그 결함을 막았다 |
| RC1 재채택 | — | 4분류(documented/not_run 분리 — support-matrix 머리글·e2e-receipt schema) · Codex `.agents/skills`(문서5:281) · **훅: RC1은 0개 → rc.2 I1에서 허용 목록 생성으로 전환**(RC1 0개는 decisions 19:1x의 02-B·02-C 리더 임시 결정 두 행 "hook 0개". rc.2 I1 #2238, PR #2260 병합 `a6f29649d5`가 고정된 MoAI 훅 집합을 허용 목록 + 영수증으로 생성 — `docs/hooks-generation.md`. P04는 04:54(11-09)에 첫 RC 인증 조합만 정했으므로 훅의 근거가 아니다. 문서5:468 방향과의 관계는 적용 시 원문으로 다시 확인) · `$moai` 변환(문서5:302 → user-entry-draft.md) |

### 행 6 — 팀 개발 워크플로(문서6 `0078`)

| 열 | 현재(상위:2336·2340) | 교체안 |
| --- | --- | --- |
| 남길 것 | 과거 상세 계약·Pele 원문과 검토 기록 보존 | **Issues 정본·FSM E01–E27·위임 상태 PREPARED→APPLIED·recover-release(문서6:1568–2313)** — "Pele"는 문서6:483 한 번뿐(도식 원문은 문서1:1429 소유) |
| 비계승 | 당시 담당자·lane·DB·제공자 제한을 새 제품 전체에 적용 | **GLM 제외·@codex 조건부(문서6:202)와 "담당 agent가 PR 후 종결까지"(문서6:1400) — RC1은 레인 이동 + 리더 통합 큐로 결정(decisions:77)**. "DB 제한"은 비계승이 아니라 계승(문서6:198 = SQLite 미도입)이며 "담당자·lane"은 RC1 agent-N으로 재채택 |
| RC1 재채택 | — | agent-N·`-a` 명명(문서6:1406 → session.rs) · Issues 정본(문서6:198) · PR ≠ DONE(문서6:208 → pr-contract.md) · 증거 없는 force 해제 불가(문서6:2313 → claim.rs) |

## 교훈 10건 — 본문 삽입 위치와 완성문 (synthesis §5 / group4 §7.2)

| # | 완성문 | 삽입 위치 |
| --- | --- | --- |
| 1 | 보고서·진행 기록 전용 커밋이 head를 바꿔 검증을 stale로 만드는 순환 — 증거는 freeze 전에 포함하거나 HEAD 밖에 둔다 | §15 검증·감사 절(stale 행 직후) |
| 2 | 종료 순서는 ExitWorktree(keep) → main 확인 → /clear → 다음 카드이며 dirty·미push·활성 검사를 보존한다 | §6 역할·handoff 점검 절 |
| 3 | PR 이후 책임 귀속 — 레인 이동 + 리더 통합 큐(또는 담당자 종결까지) 중 하나를 명시한다 | §6 단일 owner 절 |
| 4 | 리더 제어 루프는 긴 작업을 안에서 기다리지 않고 대기 이유·다음 trigger를 기록한다 | §15 차단 절 앞 새 문단 |
| 5 | 필수 검사·CI trigger는 저장소 보호 설정과 현 head·event에서 읽고 경로가 없으면 CI_TRIGGER_UNREACHABLE로 표시한다 | §15 검증 절 |
| 6 | hooks·skill은 보안 경계가 아니며 host hook은 fail-open일 수 있다 | §14 host adapter 절 |
| 7 | 성능·생산성 수치는 같은 조건 baseline 뒤에만 비교하고 측정 전 절감률을 발표하지 않는다 | §16–17 지표 절 |
| 8 | 자동 설치·quarantine 제거·trust 우회를 bootstrap 기본값에 넣지 않는다 | §3 설치 절 |
| 9 | 개인 절대 경로·비밀은 보고서·증거 작성 시점에 제외한다 | §15 전송 행 앞(작성 시점으로 확장) |
| 10 | 외부 SHA·버전·시각은 한 표로 고정한다 | §22 출처 절(Go 4fa2aa9 행 외에 확장) |

## 적용 순서

1. 위 행별 교체안을 상위 §20 표에 적용하고 "RC1 재채택" 열을 추가한다. 2. 교훈 10건을 지정 절에 완성문 그대로 넣는다. 3. 경로·어휘·인터페이스·시험 ID는 부록 1~4를 참조만 하게 한다(§4-1 교차 규칙).

## 부록 1~4의 위치 — 저장소 링크로 둔다

부록 4종은 Space 상위 페이지에 본문을 복사하지 않고 **저장소 경로 링크**로 둔다. 상위 페이지 부록 절에는 아래 네 줄과 확인 기준(tip `86762b8e1cb4fa7a3638e859331d41d542767f14`, 2026-11-09)만 적는다.

| 부록 | 저장소 경로 |
| --- | --- |
| 1. 상태 어휘 대응표 | `docs/design-review/appendix/state-vocabulary-crosswalk.md` |
| 2. 경로 표 | `docs/design-review/appendix/paths.md` |
| 3. 인터페이스/명령 대응표 | `docs/design-review/appendix/interfaces.md` |
| 4. 시험 ID 대응표 | `docs/design-review/appendix/test-ids.md` |

이유: 부록 3의 OPERATIONS 목록은 `src/main.rs`의 시험 `operations_doc_matches_code`가 코드와 대조하는 살아 있는 표라서 작업이 병합될 때마다 바뀐다. 부록 1·2는 RC1 열을 tip `604792666`에서 실측했다고 적고 있다. 그래서 Space에 옮긴 복사본은 곧 저장소와 어긋나고, 어긋남을 검사할 장치도 없다. 링크는 `v2026.11.0` 브랜치 경로를 가리키고, 적용 시점의 tip SHA와 확인일을 함께 적는다.
