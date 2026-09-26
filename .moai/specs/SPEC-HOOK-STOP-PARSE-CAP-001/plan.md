# Plan — SPEC-HOOK-STOP-PARSE-CAP-001

- 카드: t1272 (Class C). 브랜치 `WT-stop-parse-cap`, 기준 트리 `e464fd5d0`.
- Tier: **M**. 판단 근거 — 바뀌는 파일이 대략 6~8개(`internal/cli/hook_stdin_failclosed.go`, `internal/cli/hook.go`, 셈 기록 구현 파일 1개, 그 테스트, `internal/codexadapter/stop_cap.go` 주석, 운영자 문서 템플릿 원본과 로컬 사본)이고, 예상 변경량은 테스트 포함 300~600줄 수준이다. 설계 결정(열쇠, 만료, 실패 방향)이 여럿이라 AC 를 별도 파일로 두는 편이 감사에 유리하다. 5개 산출물이 필요한 헌법급 변경은 아니다.

---

## §A 맥락

선행 SPEC-HOOK-STDIN-FAILCLOSED-001 은 Claude Stop 의 파싱 실패 루프를 호스트 상한이 끊는다고 가정했다. t1230 과 t1272 의 측정(spec.md §A.1)이 그 가정이 성립하지 않는 두 조건을 보였다 — 200 주입 세션, 그리고 도구 사용이 끼어드는 루프. 리드 판정 D1(spec.md §A.3)에 따라 moai 가 호스트 프로세스별로 셈을 두고 N=8 을 넘으면 차단을 멈춘다. 같은 카드에서 사유 문구를 바꿔 모델이 훅·설정 파일을 편집하지 말고 사람에게 알리도록 한다.

현재 코드의 사실(이 트리에서 판독):

- `answerStdinParseFailure`(`internal/cli/hook_stdin_failclosed.go`)는 상태가 없다. 같은 입력이면 몇 번째 호출이든 같은 출력을 낸다. 이 트리에서 빌드된 바이너리로 파싱 실패 Stop 한 번을 실행한 결과는 acceptance.md §E 증거 원장 L1 에 있다.
- 사유 문구는 `stdinParseFailClosedReason()` 한 함수가 두 하네스에 공통으로 준다. REQ-SPC-009·010 을 함께 만족하려면 사유가 하네스별로 갈라져야 한다 — Codex 는 기존 문자열, Claude 는 새 문자열.
- 파싱 성공 경로는 `runHookEvent`(`internal/cli/hook.go`)의 `ReadInput` 뒤에 있다. REQ-SPC-005 의 초기화는 이 지점에서 Claude·Stop 일 때만 일어난다.

---

## §B 질문

### B.1 Kickoff 전에 판정이 필요한 질문

- [NEEDS CLARIFICATION: 만료 시간의 값] REQ-SPC-007 의 만료 시간. 제안값 **60분**. 근거: 파싱 실패 Stop 사이의 간격은 한 턴의 작업 시간이다. t1272 측정에서 23회 Stop 이 138초 안에 났지만, 긴 빌드·테스트를 도는 턴은 수십 분이 걸릴 수 있다. 만료가 턴 간격보다 짧으면 매 Stop 마다 셈이 1 로 돌아가 상한이 걸리지 않는다(fail-closed 쪽으로 기울며 오늘의 동작과 같다). 길수록 부모 프로세스 id 재사용 창이 넓어진다(spec.md §F 첫 행). 선택지: (a) 60분(권장), (b) 10분 — 재사용 창은 좁지만 긴 턴에서 상한이 풀리지 않는다, (c) 24시간 — 긴 턴에도 확실히 걸리지만 재사용 창이 넓다. 사용자에게 보이는 동작(언제 차단이 다시 시작되는가)을 바꾸는 값이므로 리드 판정을 받는다.

### B.2 판정이 필요 없는 설계 선택 — 근거와 함께 기록

- **부모 프로세스를 식별할 수 없을 때의 방향: 차단 유지(fail-closed).** REQ-SPC-008. 근거 셋. (1) 오늘의 동작과 같으므로 리드 판정 밖의 사용자 가시 변화가 없다. (2) 식별 실패를 해제로 바꾸면 「식별 정보가 없다」는 조건이 곧 거부를 통과로 바꾸는 길이 된다 — 선행 SPEC REQ-HSF-001 의 금지 절(거부를 통과로 바꾸는 스위치 금지)과 같은 위험이다. (3) 식별 실패는 드문 경로다: `os.Getppid()` 는 유닉스에서 항상 값을 주며, 1 이하는 부모가 이미 죽어 init/launchd 로 재부모된 경우다. 대가: 그 드문 경우 루프는 호스트 상한만으로 묶인다(선행 SPEC 의 상태). 이 방향은 리드 판정의 범위 안이라 표식을 달지 않는다.
- **셈 기록 손상 시 방향: 없는 것으로 보고 1 부터.** 역시 fail-closed 쪽이다. 손상된 기록을 큰 값으로 해석하는 선택지는 두지 않는다.
- **상한 해제 뒤 지속 해제.** REQ-SPC-003. 파싱할 수 없는 Stop 에서 턴 경계를 읽을 수 없으므로(`stop_hook_active` 판독 불가) 「턴마다 N 번」은 구현할 수 없다. D1 의 「연속 파싱 실패 Stop」을 문자 그대로 적용한 결과이며, 되돌리는 것은 파싱 성공과 만료뿐이다.
- **N=8 의 근거.** 리드 판정 값이며 호스트 기본 상한(8)과 같다. 호스트 기본 상한이 이미 걸리는 무도구 루프에서는 두 상한이 거의 같은 시점에 걸린다 — t1230·t1272 B팔에서 훅이 9번째 호출까지 차단했으므로, moai 의 9번째 상한 해제와 호스트의 9번째 종료가 겹친다(해석, 순서는 미측정).

---

## §C 착수 전 점검 (Pre-flight)

run-phase 착수 시 다음을 다시 재고 progress.md §E.2 에 남긴다.

1. `git rev-parse HEAD` 와 `git merge-base develop HEAD` — 기준 트리가 `e464fd5d0` 에서 움직였으면 아래 판독을 다시 한다.
2. `internal/cli/hook_stdin_failclosed.go` 의 `answerStdinParseFailure` 서명·분기 순서가 plan 시점과 같은지. 다르면 D-NEW-1 경로로 spec 을 먼저 고친다.
3. `.claude/settings.json` 의 `Stop` 항목과 `.claude/hooks/moai/handle-stop.sh` 가 여전히 `exec` 사슬인지(spec.md §A.4). 템플릿 원본(`internal/template/templates/.claude/settings.json.tmpl`, `internal/template/templates/.claude/hooks/moai/handle-stop.sh.tmpl`)도 같은지 확인한다. `exec` 가 아니면 부모 프로세스 열쇠가 무의미해지므로 run 을 멈추고 blocker 로 보고한다.
4. 셈 기록 위치 제안: `<프로젝트 루트>/.moai/state/stop-parse-cap/<부모 pid>.json`. 프로젝트 루트는 기존 영속 기록과 같은 `resolveHookProjectRoot()` 로 정한다. `.moai/state/` 는 런타임 관리 영역이며 저장소에 커밋되지 않는다 — plan 시점 확인: `git check-ignore -v .moai/state/stop-parse-cap/123.json` → `.gitignore:398:.moai/state/`. 착수 시 다시 확인한다.
5. 열쇠 보강 선택지(요구 아님): `homestate.ProbeProcessIdentity(pid)` 가 돌려주는 프로세스 지문을 셈 기록에 함께 적고, 지문이 다르면 기록을 없는 것으로 본다. 재사용 위험(spec.md §F)을 줄이지만 매 파싱 실패 Stop 마다 프로세스 조회 비용이 든다. 채택 여부는 run-phase 가 비용을 재고 정하며, 채택해도 REQ 문구는 바뀌지 않는다.

---

## §D 제약

- `internal/hook` 패키지를 수정하지 않는다(선행 SPEC 과 같은 원칙 — 이음매는 CLI 계층).
- Codex 쪽 출력 바이트와 사유 문구를 바꾸지 않는다(REQ-SPC-009, 카드 t1233 범위).
- 테스트는 셈 기록 위치를 `t.TempDir()` 아래로 돌린다. 부모 pid 는 테스트가 주입할 수 있어야 한다 — 실제 `os.Getppid()` 는 테스트 프로세스의 부모라 여러 테스트가 공유하므로, 병렬 테스트가 같은 셈 기록을 건드리지 않게 한다.
- 전체 스위트를 로컬에서 돌리지 않는다. 영향 패키지(`./internal/cli/...`, `./internal/codexadapter/...`)만 돌리고 전체 판정은 CI 에 맡긴다. `internal/cli` 전체 스위트는 무거운 실행이므로 `moai slot acquire` 로 자원 임대를 먼저 잡는다.
- 운영자 문서는 템플릿 원본을 먼저 고치고 `make build` 로 임베드한다(Template-First).

---

## §E 자기 검증

run-phase 완료 보고는 acceptance.md 의 AC 별로 명령·출력 원문·트리 SHA 를 싣는다. TDD 이므로 AC-SPC-002·004·006 의 RED 는 구현 전 실패 출력 원문으로 남긴다(현재 코드에는 셈이 없어 9번째 호출도 거부한다 — 그 실패가 RED 다).

---

## §F 마일스톤 (결정의 번복 가능성 순)

### M1 — 셈 기록의 모양과 열쇠 (Priority High)

- 셈 기록의 필드(연속 횟수, 마지막 갱신 시각, 선택 시 프로세스 지문), 위치(§C 4), 쓰기 방식(임시 파일 + 이름 바꾸기로 반쯤 쓰인 내용 방지 — REQ-SPC-008), 만료 판정(B.1 의 값), 만료 정리(REQ-SPC-007).
- 부모 pid 공급원을 주입 가능한 한 곳에 둔다. 1 이하이면 「식별 불가」.
- 이 마일스톤이 먼저인 이유: 열쇠와 만료는 사용자 가시 동작(언제 해제되고 언제 다시 차단되는가)을 정하며, 뒤집히면 모든 테스트가 바뀐다.

### M2 — 사유 문구 하네스 분리 (Priority High)

- Claude 하네스 사유를 REQ-SPC-010 의 고정 문자열로, Codex 하네스 사유는 기존 문자열 그대로 둔다.
- 기존 테스트 `internal/cli/hook_stdin_failclosed_test.go` 의 기대 사유(현재 `:815`)를 하네스별로 갈라 갱신한다. 선행 SPEC AC-HSF-001(d)·(e4), AC-HSF-013 의 Claude 모드 행이 이 변경의 영향을 받는다.
- 사용자 가시 문구이므로 M3 보다 앞에 둔다.

### M3 — 상한 분기와 초기화 배선 (Priority High)

- `answerStdinParseFailure` 의 Claude·Stop 거부 분기 앞에 셈 증가와 N 판정을 둔다. N 이하는 기존 거부, 초과는 상한 해제 응답 + 전용 키 기록 + stderr 한 줄.
- `runHookEvent` 의 파싱 성공 지점에서 Claude·Stop 일 때 셈 기록을 지운다(REQ-SPC-005). 다른 이벤트는 셈에 손대지 않는다(REQ-SPC-006).
- 판정에 환경 변수를 읽지 않는다(REQ-SPC-004) — AC-SPC-005 의 정적 검사가 잡는다.

### M4 — 주석·운영자 문서 정정 (Priority Medium)

- `internal/codexadapter/stop_cap.go` `HostLacksStopBlockCap` 주석의 「Claude Code is not covered here … has not been measured」 문장을 t1230·t1272 측정과 이 SPEC 의 moai 자체 상한으로 고친다(REQ-SPC-013).
- `internal/cli/hook_stdin_failclosed.go` 의 Stop 루프 `@MX:WARN`·`@MX:REASON` 을 「호스트 상한 + moai 자체 상한 N=8」로 고친다.
- 운영자 문서 템플릿 원본과 로컬 사본에 REQ-SPC-012 의 내용을 더한다(sync-phase 로 넘겨도 된다 — 그 경우 sync 의 manager-docs 가 소유).

### M5 — LIVE 재측정 (Priority Low, 운영자 예산)

- 착수 전 상한 선언: 실행 2회(A: `CAP=200`, B: 상한 변수 unset), 실행당 `--max-turns 30`, 벽시계 300초. 재실행 없음.
- 훅 래퍼는 **`exec` 형**으로 쓴다: `#!/bin/sh` 다음 줄에서 `exec <빌드한 바이너리> hook stop < <파손 stdin 파일>`. t1272 의 `hook.sh` 는 파이프로 moai 를 불러 부모가 셸이었으므로 그대로 쓰면 셈이 매번 1 이다(spec.md §F).
- 기대: 두 팔 모두 파싱 실패 Stop 이 9번째에서 `{}` 를 받고 턴이 끝난다. 관측 항목(차단 AC 아님): 새 사유 문구를 받은 모델의 훅·설정 편집 시도 횟수.
- 증거는 `.moai/reports/t1272/` 아래(gitignored). 모델 사용량에 달려 있으므로 이 마일스톤이 막혀도 M1~M4 의 완료를 막지 않는다.

---

## §G 금지 패턴

- 호스트 상한 값을 읽어 판정하는 구현(D2 부활).
- 셈 기록을 세션 id 로 열쇠 삼는 구현 — 파싱 실패에서는 세션 id 를 읽을 수 없다.
- 손상·판독 불가를 해제 쪽으로 해석하는 구현.
- 사유 문구에 복구 절차(`disableAllHooks`, moai 갱신 방법)를 넣는 것.
- Codex 사유 문구·번역 틀을 함께 바꾸는 것.

---

## §H 교차 참조

- `.moai/specs/SPEC-HOOK-STDIN-FAILCLOSED-001/spec.md` — REQ-HSF-001·009·010·011·012·013, §F.2 (0.4.3 정정본)
- `.moai/reports/t1272/verdict.md`, `.moai/reports/t1230/verdict.md` — 로컬 증거
- `internal/cli/launcher_blockcap_infinite.go` — 200 주입 조건
- `.claude/rules/moai/workflow/goal-directive.md` — 호스트 상한의 기존 서술
