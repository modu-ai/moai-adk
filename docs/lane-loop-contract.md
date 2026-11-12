# 레인 반복 계약과 리더 점검 루프 (제안, 2026-11-08)

> 현행 상태(#2167): 이 문서는 2026-11-08 제안으로 남겨 둔 기록이다. 현행 규칙은 `AGENTS.local.md` § Agent loop (self-claim)·§ Leader supervision·§ Gate와 `docs/design-review/autonomy-modes-spec.md` §3(self-claim)·§5(리더 감독)에 있다. 아래 HOLD 중 결정된 것: P08 claim 방식과 lease 4h·자동 takeover 없음(#2166, D-5), 통합 owner = 리더(통합 게이트), 레인 수령 = 리더 배차 대신 self-claim(#2161). 독립 감사는 PR 개설 뒤에 한다(아래 표 4·5단계 순서와 다르다).

이 문서는 RC1 Issue 상세 설계가 끝난 뒤 각 레인(Claude Code·Codex·Z Code·OpenCode 세션)이 Issue 하나씩 처리하는 반복과, 리더가 PR·Issue를 점검하고 수리를 배차하는 루프의 계약 초안이다. 설계 정본 §8(한 Issue 한 세션의 claim과 병렬 처리)과 Issue #2005·#2006·#2007(06-A/B/C)의 상세 설계를 따르며, P08(claim primitive·lease·handoff)이 HOLD인 부분은 선택지로 둔다. 배차 시작 조건과 HOLD 해소 전에는 적용하지 않는다.

## 1. 전제와 경계

- GitHub Issues/PR이 영구 작업 원장이다. 레인 간 실시간 메시지(통신 허브)는 전달 계층일 뿐, 메시지만으로 소유·승인·완료가 확정되지 않는다.
- 한 Issue에 활성 owner는 한 세션이다. GitHub assignee·label은 표시이며 lock이 아니다(설계 00:591·607·636, 11:72).
- 통합 대상 브랜치는 `v2026.11.0`. 원격 `main`·`2026.11.0-rc.1`은 건드리지 않는다. Go `.moai/`·`~/.moai/`는 보존하고 Rust 상태는 `<project>/.moai-cli/`·`~/.moai-cli/`만 쓴다.
- 미결 기술 선택(P04~P15, Issue 본문의 HOLD)은 레인이 임의로 확정하지 않는다. 막히면 Issue 소유 댓글에 blocked로 기록하고 리더에게 알린다.

## 2. 레인 반복 (lane loop)

| 단계 | 레인이 하는 일 | 증거·기록 |
| --- | --- | --- |
| 0 수령 | 리더가 조정 record에 claim(issue·owner session·attempt·epoch·base SHA·write scope)을 기록한 뒤 Issue 번호를 전달. 레인은 Issue 본문(설계·HOLD·의존성)과 선행 Issue 상태를 읽고 blocked_by가 열려 있으면 시작하지 않는다 | 소유 댓글 1건: `상태: assigned`, claim epoch |
| 1 EnterWorktree | Issue 전용 worktree·브랜치(`issue/<번호>-<키>`, base = 최신 `origin/v2026.11.0`)를 만든다. 공유 checkout에는 쓰지 않는다 | worktree 경로·base SHA를 소유 댓글에 기록 |
| 2 구현 | Issue의 파일 범위 안에서만 변경. 공유 interface(agents.rs·schema·lockfile·.github/workflows 등)는 설계가 지정한 소유자 Issue만 바꾼다 | 커밋 메시지에 `Refs #번호` |
| 3 로컬 build/test | 저장소 검사 묶음을 실제 실행하고 exit code·개수를 기록: `cargo test --offline --locked`, `cargo clippy --offline --locked --all-targets -- -D warnings`, `cargo fmt --check`, `python3 -m unittest discover -s tests -p '*_test.py'`, `git diff --check` + Issue의 AC 검증 명령. 빌드 폴더는 worktree 안 `target/` 하나(`CARGO_TARGET_DIR` 지정 안 함, #2380 WT-2) | 소유 댓글에 명령·exit·수치. 실행 안 한 항목은 not_run |
| 4 독립 review | 작성 세션과 다른 세션(감사 레인)이 SPEC·exact diff·시험 증거를 대조 | review verdict 링크 |
| 5 PR | `v2026.11.0` 대상 PR 개설, 본문에 Issue·SPEC digest·base/candidate SHA·시험 증거·독립 review·남은 gap. `Refs #번호` 기본, `Closes`는 승인 범위·완료조건 충족 때만 | 상태 `pr_ready → in_review` |
| 6 통합 | 통합 owner(리더 또는 지정 세션)가 최신 base 재검증·CI·보호 규칙·HOLD 확인 뒤 merge. 레인은 merge하지 않는다(HOLD-06C: 통합 owner 선택지) | 상태 `integrated`, 통합 SHA |
| 7 ExitWorktree | 미커밋 변경·검증 artifact 보존 확인 후 worktree 종료, claim 해제 요청. worktree는 레인이 지우지 않는다. 병합 뒤 리더가 `worktree-gc-plan` 목록으로 cargo 출력을 먼저 지우고, worktree는 사용자 보고 뒤 지운다(결정 2026-11-10 Q4 B, #2380 WT-2) | 소유 댓글 최종 갱신, 리더 record 해제 |
| 8 다음 Issue | 리더에게 다음 Issue 요청. 동시에 두 Issue의 활성 claim을 갖지 않는다 | — |

중단·인계: 세션이 끊기거나 멈추면 리더는 journal·PR·worktree 상태를 먼저 읽고(결과 불명은 uncertain), 명시 handoff로만 새 attempt를 연다. 무응답만으로 자동 takeover하지 않는다(P08 HOLD).

## 3. 리더 점검 루프 (leader loop)

1. 주기적으로 열린 PR·Issue를 읽는다(`gh pr list`, `gh issue list --milestone 2026.11.0-rc.1`, 각 Issue 소유 댓글의 `상태:`).
2. 오류·차단 후보(CI 실패, review FAIL, 소유 댓글 blocked, HOLD 미해소, 의존성 미충족 PR)를 모아 **fan-out 리뷰**: 읽기 전용 검토 세션 여러 개에 PR/Issue를 나눠 맡기고 finding을 수집한다.
3. 수리가 필요하면 원래 owner 세션에 지시한다. owner가 없거나 응답하지 않으면 명시 handoff 후 다른 세션에 수리 Issue(또는 같은 Issue의 새 attempt)를 배차한다.
4. 통합 owner 역할로 최신 base·CI·보호 규칙·HOLD를 확인한 뒤 허용된 PR만 merge하고, 결과를 Issue 소유 댓글과 조정 record에 재조회로 기록한다.
5. 이전 세대 label/assignee를 상태 권위로 쓰지 않는다. 상태는 댓글의 `상태:` 줄과 record다.

## 4. 각 레인에 보낼 /goal 지시문 (초안)

> 당신은 Issue #<번호> 하나의 활성 owner입니다. 이 Issue가 `integrated`로 끝나거나 리더가 명시 handoff를 선언할 때까지 반복합니다: EnterWorktree(`issue/<번호>-<키>`, base `origin/v2026.11.0`) → Issue 본문의 파일 범위 안에서 구현 → 저장소 검사 묶음과 AC 검증 명령을 실제 실행해 exit code 기록 → 독립 review 요청 → PR(`v2026.11.0` 대상, `Refs #번호`) → 리더 통합 확인 → ExitWorktree → 리더에게 다음 Issue 요청. HOLD 항목은 확정하지 말고 blocked로 보고합니다. Go `.moai/`·`~/.moai/`·원격 `main`은 건드리지 않습니다. 실행하지 않은 검증은 not_run입니다.

## 5. HOLD (사용자 결정 전)

- P08: claim primitive(로컬 journal+OS lock / 원격 ref create-if-absent / CI concurrency / Go lease 의미 재사용), lease 만료·heartbeat 수치, 자동 인계 정책.
- 통합 owner: 리더 단일 vs owner merge+dispatcher 직렬화(HOLD-06C).
- 통신 허브(세션 간 실시간 전달)의 전송·수명·ACK 설계: 별도 조사 보고 뒤 결정.
- 브랜치 이름 규칙 `issue/<번호>-<키>`와 PR 템플릿 정본 Issue(00-C/06-C/08-A/08-B 중 하나).
