# 트렁크 백업 sync — `moai integration backup` 운영.runbook

> SPEC-TRUNK-BACKUP-SYNC-001 (card t1832)가 반출한 운영 문서. **로컬 전용 문서** — 분산 템플릿에 미러하지 않는다(REQ-TBS-009). 리더·운영자가 배치 사이 간격에 로컬 `main`의 두 번째 사본을 origin에 확보하는 절차와 그 경계를 규정한다.

## 1. 무엇을 하는가

레인의 카드는 로컬 `main`에 착지하고(`.claude/rules/local/gitflow-lane-protocol.md` §2), `origin/main`은 리더의 릴리스 배치 PR(`release/main-batch-YYYYMMDD`)로만 앞선다(§4). 배치 사이에는 최대 20개 머지 커밋(`git_strategy.manual.lead_push_threshold`)이 로컬 머신에만 존재한다. 이 verb는 그 단일 사본 창을 닫는다:

```bash
moai integration backup
```

로컬 `refs/heads/main`을 origin의 `refs/heads/sync/main`에 밀어넣는다 — `.github/workflows/`의 그 어떤 workflow도 반응하지 않는 ref(동일성은 Go 회귀 테스트 `internal/cli/workflow_sync_ref_guard_test.go`가 지킨다). push argv의 플래그·원격·대상 ref는 소스 상수고, refspec의 소스측은 실행 시작에 한 번 해석해 고정한 SHA다:

```
push --force --no-follow-tags --no-recurse-submodules origin <sha>:refs/heads/sync/main
```

- refspec의 `<sha>`는 verb가 실행 시작에 한 번 해석한 로컬 `main`의 전체 SHA다(고정 refspec — sync-audit round 2 R2-F1/R2-F2). 전송된 커밋과 성공 줄이 출력하는 SHA는 항상 동일하다: `remote.origin.pushurl`이 fetch url과 달라도 출력이 바뀌지 않고, push 뒤에 경쟁 writer가 `sync/main`을 덮어써도 이번 실행의 보고는 바뀌지 않는다. push 뒤의 원격 조회는 존재하지 않는다.
- `--force`는 이 refspec에만, 구성으로 스코프된다(REQ-TBS-002). 다른 어떤 remote ref도 갱신하지 않는다.
- `--no-follow-tags`는 호출자의 `push.followTags=true` 설정에도 항상 유지된다 — 어노테이션 태그가 push에 실려 `release.yml`을 켜는 일이 없다.
- `--no-recurse-submodules`는 호출자의 `push.recurseSubmodules` 설정에도 항상 유지된다(sync-audit round 1 S1) — `only` 설정에서 이 flag가 없으면 push가 종료 코드 0으로 조용한 no-op가 되고 remote에는 아무 ref도 생기지 않으며, `on-demand`에서는 서브모듈 remote라는 두 번째 remote 면이 열린다.
- `refs/heads/main`이 해석되지 않으면 push 시도 이전에 거부하고(비영구 종료 + 안내), push 실패 시 양쪽 스트림을 출력하고 추가 git 명령을 내리지 않는다(REQ-TBS-003/004).
- verb는 인수·플래그를 받지 않는다. 호출한 작업 디렉터리 기준으로 저장소를 해석하므로 primary 체크아웃이든 연결 worktree든 같은 공유 ref를 읽는다.

## 2. 리더 라운드 단계 (REQ-TBS-007 i)

카드 하나의 머지가 로컬 `main`에 착지했을 때, 리더(또는 운영자)는 이 verb를 실행해 착지 작업이 배치 사이 구간에도 origin의 두 번째 ref로 존재하게 둘 수 있다. 카드 1회 착지마다 실행할 수도 있고 라운드 마무리에 실행할 수도 있다 — `lead_push_threshold` 트리거(SPEC-LEAD-AUTOPUSH-001)와 독립이며 그 임계값을 읽거나 옮기지 않는다. 자동화는 없다: 스케줄러·크론·훅·팩토리-리더 코드가 이 verb를 호출하지 않는다(SPEC §F, 결정 Q4).

## 3. origin push의 승인 경계 (REQ-TBS-007 ii)

이 ref의 **모든** origin push는 운영자 승인 배치 범위다. 카드 t1832의 경계 조건 그대로: 이 카드는 메커니즘 + 테스트 + 안내서만 반출하며 실행 시점의 origin push는 리더의 승인된 배치 행위다. 레인은 이 verb를 실행하지 않는다(레인의 push 금지, `gitflow-lane-protocol.md` §4 — 본 runbook은 그 규율의 예외가 아니다).

## 4. `sync/main`은 redundancy copy (REQ-TBS-007 iii)

`sync/main`은 중복 사본일 뿐이다:

- **착지하지 않는다** — 어떤 브랜치로도 머지하지 않는다.
- `release/main-batch-YYYYMMDD` pull-request 경로를 **대체하지 않는다**(SPEC-LOCAL-MAIN-FLOW-001 REQ-LMF-011).
- 무엇도 `sync/main`을 다시 읽지 않는다. 복구가 필요해지면 그 복구는 별도의 운영자 행위이며 이 SPEC 밖이다(§G non-goal).

## 5. 불변의 것들 (REQ-TBS-007 iv)

배치 릴리스 규율과 `main` 보호는 이 ref의 존재로 바뀌지 않는다:

- `origin/main`은 여전히 `release/main-batch-YYYYMMDD` PR로만 앞선다(필수 PR, `enforce_admins: true` — `.claude/rules/local/repo-local-pr-policy.md`).
- 브랜치 보호 규칙·가드·훅 설정·분산 템플릿(`internal/template/templates/**`)은 이 SPEC이 아무것도 건드리지 않았다(REQ-TBS-009).
- `sync/main` 이외의 어떤 ref에도 force-push가 허용되지 않는다. 에이전트 세션의 `git push --force` 확인 규율은 그대로다 — 이 verb의 고정 refspec(`<sha>:refs/heads/sync/main`)이 카드가 승인한 유일한 예외다.

## 6. guard 유지 보수

`.github/workflows/`에 workflow를 추가·수정할 때 `on.push` 브랜치 필터가 `sync/main`을 제외하는지 이 테스트가 검증한다(빈 `on: push`는 전체 CI 매트릭스를 켠다 — 감지된다). workflow 파일만 바뀐 diff도 Go suite를 돌리도록 `.github/test-input-filters.yml`의 `go_code`가 `.github/workflows/**`를 매핑한다(AC-TBS-010). 매처가 지원하지 않는 패턴 문법(`?`, brace 등)을 만나면 테스트는 조용히 통과하는 대신 실패한다 — guard를 확장하라는 신호다.
