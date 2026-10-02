# decision-index.md — SPEC-CODEX-REVIEW-ASYNC-001

`interview.decision_gate: on` — 카드 t1422(형제 SPEC `SPEC-CODEX-REVIEW-OWNERSHIP-001` 과 같은 카드) 조립 중 표면화된 결정 중 운영자가 인터뷰에서 확정하지 않은 것. 라벨 어휘: DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER. Q1-Q3 과 Q9-Q18 은 2026-10-02 운영자 위임(Jev `jev-1.13.0`)·리더 판정으로 verdict 가 채워졌고(Q10 은 OPEN, Q17 은 프로브 결과 대기, Q9 는 Jev 선택을 오케스트레이터가 덮어씀), Q4-Q8(EVIDENCE-NEEDED 프로브 항목)은 프로브 결과 전까지 비어 있다. Q19-Q23 은 plan-audit 1회차(`async-plan-audit.md`) 반응에서 새로 표면화된 결정이라 verdict 가 비어 있다 — 운영자·리더 확인 대기. 각 행은 무엇이 미결인지와 왜인지만 적으며 권고를 싣지 않는다(권고는 plan.md §G; `O-*` 번호는 그 표의 ID).

**권위 등록부 점검.** 권위 인용은 커밋된 산출물만 쓴다. Jev 판정은 표시 전용 신호라 권위 등록부 밖이고, 공식 훅 레퍼런스(외부 문서)도 등록부 밖이다. 어느 행의 후보 권위도 커밋 트리에서 확인하지 못했으므로 DECIDED·POLICY-COVERED 로 라우팅한 행은 없다.

### Q1: 같은 트리에서 Stop 이 겹칠 때 중복 리뷰를 어떻게 막는가?

- Label: FOUNDER
- Authority anchor: 없음
- Why unresolved: 기다림·건너뜀·세션별 락·트리별 락이 서로 다른 비용(훅 프로세스 적체, 리뷰 누락, 같은 트리 다중 세션 중복)을 낳는다. 어느 쪽을 택할지는 선호 판단이다.
- Operator verdict: Jev (operator-delegated), per_tree_lock_skip_when_held_take_over_dead_holder, confidence 0.83, 2026-10-02

### Q2: 리뷰 도중 상태가 바뀐 뒤 끝난 결과를 어떻게 다루는가?

- Label: FOUNDER
- Authority anchor: 없음
- Why unresolved: 버림·표지 없이 전달·표지를 붙여 전달 중 선택이며, 버림은 실패 신호를 잃고 무표지 전달은 현재 상태에 대한 판정으로 오독된다.
- Operator verdict: Jev (operator-delegated), label_stale, confidence 0.99, 2026-10-02

### Q3: asyncRewake 전환을 형제 SPEC 에 넣는가, 별도 SPEC 으로 분리하는가?

- Label: FOUNDER
- Authority anchor: 없음
- Why unresolved: 두 변경이 같은 `HandleCodexReviewGate` 를 만진다. 한 SPEC 에 묶으면 Tier 상한과 검증 표면이 커지고, 분리하면 착지 순서 의존이 생긴다.
- Operator verdict: Jev (operator-delegated), sibling_spec, confidence 0.93, 2026-10-02

### Q4: [E-1] 끝난 비동기 Stop 훅이 idle 세션에서 새 턴을 시작하는가?

- Label: EVIDENCE-NEEDED
- Authority anchor: 없음 — 공식 훅 레퍼런스는 "wakes Claude on exit code 2" 라고 적지만 세션이 idle 일 때 새 턴이 시작되는지는 이 레인이 받은 페이지에 없다("Run hooks in the background" 절이 잘림).
- Why unresolved: 관측이 없다. 이 SPEC 의 존재 이유(턴 종료가 기다리지 않고 실패 때만 깨움)가 걸려 있다.
- Operator verdict:

### Q5: [E-2] `asyncRewake` 훅에 `timeout` 은 어떻게 적용되는가?

- Label: EVIDENCE-NEEDED
- Authority anchor: 없음 — 레퍼런스의 `timeout` 행은 `async: true` 에 대해서만 "enforce 하지 않는다"고 적는다.
- Why unresolved: 취소·출력 폐기 여부가 관측되지 않았다. 타임아웃 마진 값(Q10)이 이 결과에 달려 있다.
- Operator verdict:

### Q6: [E-3] 동시성·중복 제거, 훅 실행 중 새 턴 시작, 훅 실행 중 세션 종료 시 동작은?

- Label: EVIDENCE-NEEDED
- Authority anchor: 없음
- Why unresolved: 문서에 없다. 락 보유자가 세션 종료와 함께 죽는지 고아가 되는지가 락 설계의 보유자 사망 시나리오와 직결된다.
- Operator verdict:

### Q7: [E-4] 비동기 훅에서 JSON `decision: block` 출력이 무시되는가, 종료 코드 0 훅의 stderr 가 Claude 에 보이는가?

- Label: EVIDENCE-NEEDED
- Authority anchor: 없음 — 레퍼런스는 동기 Stop 의 종료 코드 2 에서도 stdout JSON 을 읽는다고 적지만 비동기 경우는 적지 않는다.
- Why unresolved: 옛 신호를 병행할 수 있는지와 기존 스코프 로그 행(stderr)의 위치(Q17)가 이 결과에 달려 있다.
- Operator verdict:

### Q8: [E-5] 실행 중 세션에서 `.claude/settings.json` 훅 편집은 언제 반영되는가?

- Label: EVIDENCE-NEEDED
- Authority anchor: 없음
- Why unresolved: 문서에서 찾지 못했다. 레지스트리 변경이 같은 트리의 진행 중 세션에 미치는 영향을 단정할 수 없다.
- Operator verdict:

### Q9: [O-A] 트리당 연속 실패 전달 상한 값과 도달 시 동작은?

- Label: FOUNDER
- Authority anchor: 없음
- Why unresolved: 값(후보 3)은 근거 없는 제안이다. 도달 뒤 침묵이 "통과"로 오독될 위험과 무한 깨움 루프 위험 사이의 균형이다.
- Operator verdict: Jev (operator-delegated), cap3_silent, confidence 0.21 (약함), 2026-10-02 — **오케스트레이터가 cap3_notify_once_at_cap 으로 덮어씀**(사유: 상한 뒤 완전 침묵은 "통과"로 오독되므로 상한에서 마지막 알림 한 번이 필요하다). 상한 = 트리당 연속 실패 깨움 3회, 세 번째 깨움이 `이후 알림 없음, 상태는 미해결` 을 담은 마지막 알림이고 이후 실패는 침묵(plan-audit 1회차 D1 뒤 "같은 지문" 조건을 버리고 pass 까지 모든 실패가 침묵 — 계수 규칙은 Q19). PROVISIONAL, 리더 수용(2026-10-02)

### Q10: [O-B] 등록 타임아웃 마진 값은?

- Label: EVIDENCE-NEEDED
- Authority anchor: 없음
- Why unresolved: Q5(E-2) 결과 전에는 정할 수 없다. 현재 등록 900 과 Go 쪽 상한 900 은 동점이다.
- Operator verdict: OPEN — 프로브 E-2(`E2-timeout`) 결과 전까지 열려 있다. 60초는 제안값일 뿐 확정이 아니다 (2026-10-02)

### Q11: [O-C] "오래된 결과"의 정의를 HEAD 만이 아니라 상태 키 전체(HEAD+작업 트리 digest)로 넓히는가?

- Label: FOUNDER
- Authority anchor: 없음 — 운영자 문구는 "HEAD"만 말한다.
- Why unresolved: 넓히면 미커밋 편집 중의 낡은 결과도 표지를 받지만 운영자 요구 문구의 확장이다.
- Operator verdict: Jev (operator-delegated), full_state_key, confidence 0.56, 2026-10-02 — 오래됨 = 검토한 상태 키(HEAD + 작업 트리 digest) 전체의 차이, 표지는 둘 다 명시. 리더 수용(2026-10-02)

### Q12: [O-D] 가장 단순한 기준선(≈30 LOC) 대비 ≈7배의 크기를 수용하는가?

- Label: FOUNDER
- Authority anchor: 없음
- Why unresolved: 그 기준선은 운영자 요구 2·3 을 충족하지 못한다. 요구 2·3 의 최소 구현 대비로는 ≈1.0배다. 어느 기준선이 유효한지는 판단이다.
- Operator verdict: Jev (operator-delegated), accept_size, confidence 0.98, 2026-10-02

### Q13: [O-E] Claude 훅의 재전달 기록을 Codex 경로의 검증 영수증과 공유하는가?

- Label: FOUNDER
- Authority anchor: 없음
- Why unresolved: 공유하면 Claude 훅의 fail 이 Codex 멤버 6 의 차단 근거가 된다(Codex 경로 동작 변경).
- Operator verdict: Jev (operator-delegated), do_not_share, confidence 0.02 (약함), 2026-10-02 — PROVISIONAL, 리더 수용(2026-10-02)

### Q14: [O-F] 프로브를 누가 실행하는가?

- Label: FOUNDER
- Authority anchor: 없음
- Why unresolved: E-1·E-3(c2,c3)·E-5 는 대화형 세션이 필요할 수 있고 레인은 대화형 세션을 열 수 없다. 레인의 헤드리스 변형이 대화형과 같은 동작인지도 검증되지 않았다.
- Operator verdict: Jev (operator-delegated), lane_headless_first_leader_interactive_only, confidence 0.60, 2026-10-02 — 레인이 헤드리스 프로브를 먼저 돌리고(E1-headless, E2-timeout, E3-c1-concurrent, E4-d1-json, E4-d2-exit0stderr), 대화형만 가능한 프로브(E1-interactive, E3-c2-newturn, E3-c3-sessionexit, E5-reload)는 리더가 리더 세션에서 돌린다. 절차: `.moai/reports/t1422/async-probes.md`

### Q15: [O-G] 알림의 "카드" 칸을 트리 디렉터리 이름으로 대신하는가?

- Label: FOUNDER
- Authority anchor: `.claude/rules/local/gitflow-lane-protocol.md` §1 (워크트리 디렉터리가 카드 id 를 유지한다는 레인 규약 — 코드가 강제하지는 않음)
- Why unresolved: 코드에는 카드 id 의 출처가 없다. 규약 기반 대용이 허용되는지는 판단이다.
- Operator verdict: Jev (operator-delegated), worktree_dir_name_blank_when_not_card_worktree, confidence 1.00, 2026-10-02 — 카드 칸은 `WT-` 카드 워크트리의 디렉터리 이름, 카드 워크트리가 아니면 빈 칸

### Q16: [O-H] "알려진 상태 건너뜀"을 pass 에도 적용해 현행(상태가 같아도 Stop 마다 재리뷰)을 바꾸는가?

- Label: FOUNDER
- Authority anchor: 없음
- Why unresolved: 같은 상태 재리뷰는 비용만 들지만 비결정적 리뷰어에서는 결과가 달라질 수 있다.
- Operator verdict: Jev (operator-delegated), skip_known_state_for_pass, confidence 0.41, 2026-10-02 — PROVISIONAL, 리더 수용(2026-10-02). 동작 변경은 SPEC 본문에만 기술하고 CHANGELOG 항목으로 올리지 않는다

### Q17: [O-I] 스코프 로그 행(stderr)을 E-4 결과에 따라 유지하는가, 로그 파일로 옮기는가?

- Label: EVIDENCE-NEEDED
- Authority anchor: 없음
- Why unresolved: 종료 코드 0 비동기 훅의 stderr 가시성(Q7)에 달려 있다. 옮기면 SPEC-CODEX-GATE-SCOPE-001 REQ-CGS-010 의 sink 가 바뀌므로 별도 개정이 필요하다.
- Operator verdict: Jev (operator-delegated), keep_scope_log_line_on_stderr_pending_probe, confidence 0.97, 2026-10-02 — 프로브 E-4(`E4-d2-exit0stderr`) 결과 전까지 stderr 유지

### Q18: [O-J] Tier — 합계 파일 ≈19(>15)인데 M 유지인가, L 재분류인가?

- Label: FOUNDER
- Authority anchor: 없음
- Why unresolved: 코드+테스트 ≈14·LOC 는 M 대역이고 합계만 L 대역이다. 상한을 조용히 풀지 않고 올린다.
- Operator verdict: Jev (operator-delegated), keep_tier_m, confidence 0.70, 2026-10-02 — 코드+테스트 ≈14 파일과 미러·문서 ≈5 파일로 나누어 적는다(합 ≈19)

### Q19: [O-A 계수 규칙] 연속 실패 상한은 무엇을 세고 무엇이 해제하는가?

- Label: FOUNDER
- Authority anchor: 없음 — 오케스트레이터 재정 문구("동일한 실패는 reviewed state 가 바뀔 때까지 억제")는 커밋된 산출물이 아니다.
- Why unresolved: 문구를 문자 그대로 읽으면 같은 상태 재리뷰가 이미 REQ-CRA-005 로 막혀 있어 상한이 아무것도 제한하지 못하고, 마지막 알림의 고정 문구("이후 알림 없음")와도 모순된다. "동일"을 findings 지문으로 정의하는 읽기는 LLM 산문 제목에 의존해 재서술로 회피되고 findings 0건 fail 에서 충돌한다(감사 D1). 이 SPEC 은 연속 실패 깨움 횟수만 세고 pass 하나로 해제하는 읽기를 적었다.
- Operator verdict:

### Q20: [전달 확인] fail 기록은 언제 "전달됨"이 되고, 전달되지 않은 fail 은 어떻게 다시 전달되는가?

- Label: FOUNDER
- Authority anchor: 없음
- Why unresolved: 기록을 전달 전에 쓰면 전달이 유실된 fail 이 같은 상태에서 영구히 침묵한다(감사 D2). 세션 id+`delivered` 표지+전달 창(30분)은 설계 하나이고, fail 에는 건너뜀을 적용하지 않고 연속 실패 상한만으로 묶는 안이 대안이다. 표지 구간의 한 번 더 전달 가능성은 잔여 위험이다.
- Operator verdict:

### Q21: [파일 위치] 락·재전달 기록·억제 로그를 어디에 두는가?

- Label: FOUNDER
- Authority anchor: `.gitignore:397-398`(`.moai/logs/`·`.moai/state/` 무시 줄 — 이 저장소 한정, 다른 저장소에는 없을 수 있음)
- Why unresolved: 작업 트리 아래(`.moai/state`·`.moai/logs`)는 `.gitignore` 가 숨기는 저장소에서만 상태 키가 안정하고 그렇지 않으면 알려진 상태 건너뜀이 발화하지 않으며 Codex 경로 키가 바뀐다(감사 D6). git 디렉터리 아래는 `.gitignore` 와 무관하지만 로그 위치가 `.moai/logs/` 관례와 달라진다. `verify.Key` 에서 세 경로를 제외하는 안은 Codex 경로의 키 계산을 건드린다.
- Operator verdict:

### Q22: [래퍼 판정] 래퍼는 무엇을 보고 세션을 깨우는가?

- Label: FOUNDER
- Authority anchor: `internal/cli/CLAUDE.md:13`(종료 코드 2 = 시스템 오류 문서화)
- Why unresolved: 종료 코드 2 는 판정 실패뿐 아니라 Go 패닉·시스템 오류도 낸다(감사 D3). 종료 코드 2 와 sentinel 줄(`codex review gate: FAIL`)을 함께 요구하고 래퍼가 stderr 를 캡처해 sentinel 줄부터만 전달하는 설계는 래퍼를 복잡하게 하고 `mktemp` 의존을 더한다.
- Operator verdict:

### Q23: [숫자] findings 10건·줄 300자·summary 1500자·전체 8000바이트·전달 창 30분·상태 계산 예산 60s×2 를 받아들이는가?

- Label: EVIDENCE-NEEDED
- Authority anchor: 없음
- Why unresolved: 훅 출력 크기 상한과 상태 계산의 실제 비용 분포가 관측되지 않았다. 모두 보수적 시작값이며 `internal/config` 단일 원천에서 바꿀 수 있다.
- Operator verdict:
