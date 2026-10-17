# decision-index.md — SPEC-TRUNK-BACKUP-SYNC-001

> 상태축 없음(stateless). 이 색인은 계획 단계에서 운영자가 정하지 않은 결정 4건을 기록한다. 네 항목 모두 미결이며 `Operator verdict:` 줄은 비워 둔다. 이 파일은 어떤 옵션도 추천하거나 선택하지 않는다. 카드 t1832의 지시문("예: sync/main")은 커밋된 문서가 아니므로 권한 앵커로 쓰지 않는다 — 권한 등록부는 커밋된 산출물(`.moai/project/*.md`, 이전 완료 SPEC의 HISTORY/Amendments, `.moai/config/sections/*.yaml` 운영자 설정, 프로젝트 헌장)만 인정한다. REQ-TBS-001은 판결이 내려질 때까지 잠정값(`moai integration backup`, `sync/main`, 상수 refspec `--force`, 문서 전용 룬북)을 쓰며, 운영자가 다른 값을 정하면 상수·테스트·룬북이 한 커밋으로 함께 수정된다(spec.md §E). 어떤 수용 기준도 실제 origin push를 요구하지 않으므로 미결 판결은 진행을 막지 않는다.

### Q1: 백업 동사의 표면 — 기존 `moai integration` 그룹의 하위동사로 둘 것인가, 새 최상위 동사로 둘 것인가?

Label: FOUNDER
Class: product-level (배포되는 CLI의 사용자에게 보이는 명령 표면을 정하는 결정이므로 fail-closed 규칙에 따라 product-level이다.)
Authority anchor: (none — 커밋된 트리에 새 동사의 배치를 정하는 조항이 없다. `internal/cli/integration.go:280`의 그룹 등록부와 `internal/cli/integration_candidate.go`의 선행 형태는 존재 사실의 증거이지 결정 권한이 아니다.)
Why unresolved:
- `moai integration` 그룹은 9개 하위동사(status, acquire, release, policy, remeasure, merge, preflight, resync, candidate)를 가지며 `backup`은 없다.
- 사이드 ref push의 유일한 선행(REQ-CCI-016, 운영자 승인 2026-10-03)은 이 그룹의 `candidate` 동사다. 같은 그룹에 두는 것이 선행과 나란히 놓이는 배치이지만, integration 그룹의 의미(착지 창·후보·착지)에 백업이 속하는지는 정해진 바 없다.
- 어떤 설정 키도, 어떤 이전 SPEC의 HISTORY/Amendments 행도 이 질문을 정하지 않는다.

Options:
- (a) `moai integration backup` 하위동사. 귀결: 등록은 `integration.go:280`의 AddCommand 목록에 한 줄로 추가되고 `moai --help` 최상위 표면은 그대로다. 사이드 ref push 선행(candidate)과 같은 그룹에 놓인다.
- (b) 새 최상위 동사(예: `moai trunk-sync`). 귀결: 최상위 명령 표면이 하나 늘고, 등록 지점도 새로 생긴다. 백업이 integration 창 규율과 독립임이 표면에 드러난다.

Default: (a) (rule: smaller user-visible surface — 현행 유지·단일 revert 두 규칙은 양쪽이 같고, 최상위 동사를 늘리지 않는 쪽이 사용자에게 보이는 표면이 더 작다.)
Operator verdict: (비어 있음)

### Q2: 백업 ref 이름 — `sync/main`을 쓸 것인가, 다른 이름을 쓸 것인가?

Label: FOUNDER
Class: product-level (동사의 출력과 원격 저장소의 ref 이름으로 사용자에게 보이는 기본 동작을 정하므로 fail-closed 규칙에 따라 product-level이다.)
Authority anchor: (none — 카드 지시문은 `sync/main`을 예시("e.g.")로만 적었고, 카드는 권한 등록부에 없다. 커밋된 설정·헌법·이전 SPEC 어디에도 이 ref 이름을 정하는 조항이 없다.)
Why unresolved:
- 카드 t1832의 지시문은 "a remote ref no workflow reacts to (e.g. `sync/main`)"로 예시 형태다. 예시는 결정이 아니다.
- 현재 13개 워크플로 push 트리거 어느 것도 `sync/**`, `backup/**`, `trunk-backup/**` 어떤 이름도 허용하지 않는다(plan.md §A.3, `93f4fc75c`에서 측정) — 어느 후보도 배제 불변식을 만족한다.
- ref 이름은 상수로 박히므로(REQ-TBS-002) 이름 변경은 상수·테스트·룬북을 함께 고치는 한 커밋 수정이다.

Options:
- (a) `sync/main`. 귀결: 카드 예시와 같은 이름. 원격에 `refs/heads/sync/main`이 생긴다.
- (b) `backup/main`. 귀결: 이름이 목적(백업)을 직접 말한다. 원격에 `refs/heads/backup/main`이 생긴다.
- (c) `trunk-backup/main` (또는 그 외 다른 접두어). 귀결: SPEC 접두어와 일치하는 이름. 원격 ref 이름이 길어진다.

Default: not ranked. 발행 규칙(현행 유지 → 단일 revert → 더 작은 사용자 표면)은 이름 후보 사이에서 순위를 매기지 못한다. 잠정값 (a)은 카드 예시를 따른 것일 뿐 규칙이 선택한 기본값이 아니다.
Operator verdict: (비어 있음)

### Q3: ref 갱신의 force 방식 — 상수 refspec에 일반 `--force`를 쓸 것인가, `--force-with-lease`를 쓸 것인가?

Label: FOUNDER
Class: product-level (동사의 사용자에게 보이는 실패 양상을 정하므로 fail-closed 규칙에 따라 product-level이다.)
Authority anchor: (none — 카드 조건 "force-push is allowed on the sync ref ONLY"는 허용 범위를 정할 뿐 방식을 정하지 않고, 커밋된 트리에 이 질문을 정하는 조항이 없다.)
Why unresolved:
- `sync/main`은 어느 워크플로도 반응하지 않는 폐기 가능한 사본 ref다. 로컬 main이 되감긴 뒤(복구 시나리오) 사본이 조상이 아니면 push는 비-FF가 되고, 일반 push는 거부된다.
- `--force-with-lease`는 원격 추적 ref가 로컬에 있어야 하므로 사전 fetch 단계가 필요하고, 첫 실행(추적 ref 없음)에서 거부하는 실패 모드가 생긴다.
- 선행 `integration candidate`는 일반 `--force`를 명시적 refspec과 함께 쓴다(`internal/cli/integration_candidate.go:333`). 이는 형태의 증거이지 이 질문의 결정이 아니다.

Options:
- (a) 상수 refspec에 한정된 일반 `--force`. 귀결: 사전 fetch 없이 항상 성공 갱신. 되감림 포함 모든 상태에서 사본이 로컬 main을 그대로 따른다.
- (b) `--force-with-lease`. 귀결: 원격이 예상과 다르게 움직였을 때 거부한다. 사전 fetch와 첫 실행 거부 처리가 필요하다.

Default: (a) (rule: smaller user-visible surface — lease는 사전 fetch 단계와 첫 실행 거부라는 추가 실패 모드를 사용자 표면에 더한다. 두 옵션 모두 이 카드의 커밋 revert 하나로 되돌린다.)
Operator verdict: (비어 있음)

### Q4: 리더 라운드 통합 표면 — 문서 전용 룬북으로 둘 것인가, 리더 코드 훅을 넣을 것인가?

Label: FOUNDER
Class: product-level (리더 라운드의 자동 동작 여부는 운영자에게 보이는 팩토리 동작을 정하므로 fail-closed 규칙에 따라 product-level이다.)
Authority anchor: (none — 카드가 "mechanism + tests + instructions"를 요구한다고 적었지만 카드는 권한 등록부에 없고, 커밋된 트리에 리더 라운드의 백업 호출을 정하는 조항이 없다.)
Why unresolved:
- 리더 배치 push는 리더 세션이 완료 보고의 병합 SHA를 모아 수행하는 행동 절차다(`gitflow-lane-protocol.md` §4·§7). 백업 동사를 그 절차의 어느 지점에서 누가 부르는지는 현재 어디에도 정해져 있지 않다.
- 자동 호출(팩토리 리더 루프·훅·cron)을 넣으면 리더 코드가 늘고, 그 훅의 실패 처리·권한 검토가 추가된다.
- 카드 지시문은 이 통합을 문서 수준 통합 지점으로 유지하라고 하나(레인 지시), 그것은 결정의 기록이 아니다.

Options:
- (a) 문서 전용 룬북(`.moai/docs/trunk-backup-runbook.md`). 귀결: 리더(또는 운영자)가 라운드 절차에서 동사를 직접 실행한다. 리더 코드는 바뀌지 않는다.
- (b) 리더 라운드 코드 훅. 귀결: 병합 착지 감지 시 자동 실행이 코드로 들어간다. 실패 처리와 권한 검토가 이 카드의 범위를 넘어 늘어난다.

Default: (a) (rule: preserves current behavior — 리더 코드는 현재 아무 백업 호출도 하지 않으며, 문서 추가는 현행 동작을 바꾸지 않는다.)
Operator verdict: (비어 있음)
