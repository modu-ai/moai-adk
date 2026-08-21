# plan — SPEC-CODEX-SKILLS-CANONICAL-001

> 마일스톤은 **되돌리기 어려운 결정부터** 배치했다. M1·M2 가 뒤집히면 그 아래가 전부 다시 작성된다. M3·M5 는 iter-3 에서 실측 결과에 따라 내용과 우선순위가 바뀌었으므로 iter-2 판본의 기억으로 읽지 않는다.

## §A. 맥락

Codex CLI 는 `<repo>/.agents/skills/`(CWD→상향 병합)와 `$CODEX_HOME/skills/` 만 스캔하며 `.claude/skills/` 를 보지 않는다(M0 실측). moai 가 배포하는 스킬을 Codex 에서도 보이게 하려면 배포 시점에 `.agents/skills/` 경로를 추가로 만들어야 한다.

핵심 제약은 `//go:embed` 가 심볼릭 링크를 무음으로 버린다는 사실이다(spec §A.2). 링크는 템플릿이 아니라 배포기가 만든다.

## §B. 알려진 이슈 / 위험

| # | 위험 | 등급 | 완화 |
|---|---|---|---|
| R1 | 상수 기반 미러 집합이 슬림 프로젝트에서 깨진 링크를 만든다 | 높음 | REQ-CSC-006 + AC-CSC-014 (합성 FS 2-스킬 테스트) |
| R2 | Windows 에서 `os.Symlink` 실패 → 스킬 미노출 | 높음 | 복사 폴백(REQ-CSC-004) + `GOOS=windows go vet` 게이트 |
| R3 | 폴백이 조용해 "링크인 줄 알았는데 복사본" 상태가 잠복 | 중간 | REQ-CSC-005 관측 가능성 + AC-CSC-006 |
| R4 | `.agents/` 미등록으로 은퇴 스킬 영구 잔존 | 높음 | REQ-CSC-008 + AC-CSC-007/008 |
| R5 | 청소 글롭이 사용자 소유 항목까지 삼킨다 | **높음** | REQ-CSC-008/009 접두 한정 + AC-CSC-008(양팔) + AC-CSC-009 |
| R6 | 미러 실패가 배포 전체를 실패시켜 Claude Code 경로까지 잃는다 | 중간 | fail-open(REQ-CSC-011) + AC-CSC-013 |
| R7 | 템플릿 트리에 누군가 심볼릭 링크를 다시 넣어 스킬이 무음 소실 | 중간 | REQ-CSC-002 + AC-CSC-001 **양팔**(트리 전체 카운트 + 링크 가시 수집) |
| R8 | 복사 모드에서 `moai update` 가 매번 스킬 전량을 백업 트리에 복제 | **높음** | REQ-CSC-010 + AC-CSC-012(2번 팔). 근거 spec §A.7 |
| R9 | `.agents/` 가 사용자 저장소에 커밋됨 — Windows 체크아웃에서 링크가 텍스트 파일로 실체화 | **높음** | REQ-CSC-016 + AC-CSC-015. 근거 spec §A.8 |
| R10 | 미러 대상에 사용자 실 디렉터리가 있는데 `EEXIST` 를 "지우고 재생성"으로 처리 → 데이터 손실 | **높음** | REQ-CSC-014 + AC-CSC-011(3번 단언) |
| R11 | 비-`moai` 이름 스킬이 카탈로그에 들어와 미러는 만들고 청소는 못 지움 | 중간 | REQ-CSC-015 + AC-CSC-016. 근거 spec §A.9 |
| R12 | `manifest.Track` 이 디렉터리 링크에서 EISDIR → `Deploy` 실패, fail-open 과 충돌 | 중간 | REQ-CSC-010(기록 금지) + AC-CSC-012(1번 팔). 근거 spec §A.6 |
| R13 | 실행 순서가 만든 dangling 링크를 청소가 찾아내고도 지우지 않아 미러가 영구 잔존 | **높음** | REQ-CSC-008 (a)`Lstat`+dangling 제거 · (b)순서 + AC-CSC-007(2) + AC-CSC-008(2). 근거 spec §A.10 |
| R14 | 백업 금지를 절대 형태로 써서 사용자의 `moai` 접두 실 항목이 무백업 손실 | **높음** | REQ-CSC-010 판별자 한정 + AC-CSC-012(3번 팔). 근거 spec §A.11 |
| R15 | 출력 seam 부재로 모드·경고가 관측 불가 → MUST AC 3개 판정 불능 | 중간 | REQ-CSC-005(반환 결과) + M2 산출물. 근거 spec §B.D3 + §A.9b |
| R16 | `os.Lstat` 전환이 기존 7개 청소 뿌리 동작을 함께 바꾼다 | 중간 | 받아들인 결정(spec §B.D6 폭발 반경) + M4 회귀 확인 |
| R17 | `moai-`(하이픈) 접두 seam 재사용으로 `moai` 스킬이 미러·가드·정리에서 누락 | 중간 | REQ-CSC-015 철자 고정 + AC-CSC-016 [HARD] + M6·§H. 근거 spec §A.9 |

## §C. 사전 점검 (착수 전)

- `git rev-parse --show-toplevel` 이 이 worktree 인지 확인한다.
- `find internal/template/templates -type l | wc -l` == 0 (기준선, **트리 전체** — `.claude/skills/` 하위가 아니다).
- `grep -n 'agents' internal/template/templates/.gitignore` — 착수 시점에 항목이 없음을 확인한다(M5 의 전제).
- `ManagedCleanTargets` 의 현재 항목 수와 `.claude/skills/moai*` 의 인덱스를 기록한다(M4 의 순서 배치가 무엇 앞으로 가는지 고정하기 위해).
- `find internal/template/templates/.claude/skills -mindepth 1 -maxdepth 1 -type d -exec basename {} \; | grep -cv '^moai-'` — **1** 이 나와야 한다(이름이 정확히 `moai` 인 스킬). 0 이 나오면 카탈로그가 바뀐 것이므로 spec §A.9 를 다시 읽는다.
- `find internal/template/templates/.claude/skills -mindepth 1 -maxdepth 1 -type d | wc -l` 를 착수 시점에 **다시 측정**한다. 34 는 작성 시점 값이며, 테스트는 상수가 아니라 파일시스템과 임베드 FS 를 **서로 대조**해야 한다.

## §D. 제약

- 로컬 전체 스위트(`go test ./...`) 금지. 대상 패키지만 돌리고 전량 판정은 CI 에 맡긴다.
- `.claude/` · `.moai/` 신규·변경 파일은 템플릿 미러 + `make build` 필수.
- 템플릿 내용 중립성(SPEC ID · 내부 날짜 · 커밋 SHA · macOS 편향 경로 금지).
- 시간 추정 금지. 우선순위 라벨만 사용한다.

## §E. 자가 검증

각 마일스톤 종료 시 해당 AC 를 실제로 실행하고 출력을 인용한다. "통과했을 것"은 근거가 아니다.

## §F. 마일스톤

### M1 — 미러 집합 파생 방식 확정 (Priority High, 되돌리기 가장 어려움)

배포기가 "이번 실행에서 실제로 배포한 스킬 집합"을 어떻게 알아내는지 결정하고 구현한다. 후보 두 가지 — (a) Deploy 의 walk 도중 `.claude/skills/<name>/` 경로를 관측해 누적, (b) 배포 완료 후 대상 FS 를 다시 읽어 파생. (a) 는 tier 필터가 이미 적용된 FS 를 그대로 따르므로 §A.3 문제를 구성상 회피한다.

- 대상: `internal/template/` (신규 파일 1 + `deployer.go` 소폭)
- 닫힘 조건: AC-CSC-006, AC-CSC-014
- 이 결정이 바뀌면 M2~M4 전부 재작성이다.

### M2 — 링크 생성 + 폴백 (Priority High)

`.agents/skills/<name>` 을 `../../.claude/skills/<name>` 상대 심볼릭 링크로 만들고, 실패 시 복사 폴백 + 모드 보고. 실패-실패 시 fail-open.

- 대상: M1 이 만든 파일 + **출력 seam 신설**
- **[HARD] 출력 seam 은 이 마일스톤의 산출물이다.** iter-3 까지 여기 "출력 경로는 기존 printer 계층 재사용"이라고 적혀 있었는데 **사실이 아니다** — 실측하면 `internal/template` 에 `io.Writer` 가 없고 `Deploy` 는 `error` 만 돌려주며 패키지 전체에 printer 계층이 없다. 출력 표면은 호출자인 `internal/cli/` 에만 있다. 따라서 모드·경고를 **반환 결과로 올리는 seam** 을 여기서 만든다(REQ-CSC-005 / spec §B.D3 + §A.9b — 그 실측은 §A 에 절로 승격돼 있다). 이것을 빠뜨리면 AC-CSC-006·011(3)·013(3) 세 개가 함께 막히고 그중 둘은 MUST 다.
- 닫힘 조건: AC-CSC-002, AC-CSC-003, AC-CSC-004, AC-CSC-005, AC-CSC-006, AC-CSC-011, AC-CSC-013
- 상대 링크를 쓰는 이유: 프로젝트 디렉터리를 통째로 옮기거나 복사해도 링크가 살아남는다. 절대 경로는 `moai init` 시점 경로로 굳는다(CLAUDE.local.md §14 와 같은 실패 형태).

### M3 — 미러를 기록·백업 대상에서 제외 (Priority High)

**iter-3 에서 방향이 뒤집혔다.** iter-2 의 M3 는 "미러를 manifest 에 기록한다"였는데, 실측 결과 그 설계는 기존 seam 에서 **구현 불가능**하다(spec §A.6 — `Track` → `HashFile` → `io.Copy` 가 디렉터리 링크에서 EISDIR). 게다가 기록을 시도하면 error 가 올라와 REQ-CSC-011 의 fail-open 과 충돌한다.

따라서 이 마일스톤이 하는 일은 **두 가지를 하지 않도록 만드는 것**이다.

1. 미러 항목에 대해 `manifest.Track` 을 호출하지 않는다.
2. pre-clean 백업이 미러를 보존하지 않는다 — 복사 모드에서 특히(spec §A.7).

2번은 `backupThenRemove` 쪽 처리가 필요하다. 템플릿에 `.agents/` 가 없어 `templateManagedPaths` 가 항상 공집합이 되는 것이 원인이므로, **이번 실행이 다시 만들 미러인지**(템플릿이 같은 이름의 스킬을 가졌는지)를 판별하는 분기를 새로 작성한다 — REQ-CSC-010 의 판별자 그대로다.

[HARD] **"미러 뿌리를 백업 대상에서 제외하는" 절대 형태로 구현하지 않는다.** 그 형태는 AP-16 이 금지하는 것이며, `moai` 접두를 쓴 사용자 실 항목이 경고도 백업도 없이 사라진다(spec §A.11). 판별은 뿌리 단위가 아니라 **이름 단위**다.

- 닫힘 조건: AC-CSC-012 (3팔 — manifest 부재 / 재배포분 백업 부재 / 은퇴분 백업 보존)

### M4 — 청소 경로 등록 (Priority High, 독립적)

`ManagedCleanTargets` 에 `.agents/skills/moai*` 글롭 추가. 기존 `.claude/skills/moai*` 항목과 같은 형태(`IsGlob: true`)를 따른다.

[HARD] 글롭은 **접두 `moai*` 한정**이다(spec §B.D5). `.agents/` 나 `.agents/skills/` 전체를 잡는 형태는 네임스페이스 분리 계약 위반이며, 이것이 `moai update` 의 삭제 범위를 넓히는 **동작 변경**이라는 점은 spec §A.5 에 근거와 함께 기록돼 있다. 등록 전에 §A.5 를 읽는다.

**[HARD] 이 마일스톤의 진짜 작업은 글롭 등록이 아니라 dangling 링크 제거다.** 등록만 하면 청소는 미러를 **한 개도 지우지 못한다** — 실행 순서가 clean→deploy 이고 청소 안에서 정본 글롭이 먼저 돌아 모든 미러 링크가 dangling 이 되는데, `backupThenRemove` 는 `os.Stat` 으로 판정해 `IsNotExist` 에서 조용히 `return 0, nil` 한다(spec §A.10, 재현 출력 포함). 따라서 두 가지를 함께 넣는다.

- **(a) `os.Lstat` 판정 + dangling 제거** — 결함을 실제로 닫는 절.
- **(b) 슬라이스 순서** — `.agents/skills/moai*` 를 `.claude/skills/moai*` **앞에** 배치. 이중 방어이며, 중복이라며 지우지 않는다(spec §B.D6).

**(a) 는 공유 코드 변경이다.** `backupThenRemove` 는 모든 관리 대상 청소 항목이 지나가므로 기존 7개 뿌리에서도 dangling 링크가 제거되기 시작한다. spec §B.D6 이 이를 폭발 반경과 함께 **기록된 결정**으로 받아들였으므로, 이 마일스톤은 기존 뿌리에 대한 회귀 확인을 함께 수행한다.

- 대상: `internal/cli/update/deploy/deploy.go`, `internal/defs/dirs.go`(경로 상수)
- 닫힘 조건: AC-CSC-007(양팔 — 글롭 + 순서), AC-CSC-008(4형태 — dangling 팔 포함), AC-CSC-009
- M1~M3 과 독립이므로 병행 가능하다.

### M5 — `.gitignore` 에 `.agents/skills/moai*` 등록 + 템플릿 미러 (Priority **High**)

**iter-3 에서 Priority Low → High 로 올렸다.** iter-2 는 이 항목을 "`.agents/` 취급이 필요한지 **확인한다**"로 두어 판단을 run-phase 로 미뤘는데, 실측해 보니 확인이 아니라 **결정된 작업**이다: 템플릿 `.gitignore` 에 `.agents/` 항목이 없고(spec §A.8), 이대로 착지하면 링크 모드에서는 Windows 체크아웃이 링크를 텍스트 파일로 실체화하고 복사 모드에서는 스킬 전량이 사용자 저장소에 중복 커밋된다. "기계 작업" 분류도 결과와 맞지 않았다.

- `internal/template/templates/.gitignore` 에 **`.agents/skills/moai*`** 항목 추가 → `make build` → 임베드 반영 확인(AC-CSC-015 의 3번 단언이 빌드 누락을 잡는다).
- [HARD] `.agents/` **전체**를 무시하는 형태로 쓰지 않는다(spec §B.D7). 생성물은 `.agents/skills/moai*` 뿐이고, 전체를 무시하면 사용자의 `.agents/skills/hns-*` 와 후속 마일스톤(M2 AGENTS.md 정본화)의 소스 파일까지 조용히 추적에서 빠진다. AC-CSC-015 의 2번 단언이 이 형태를 금지한다.
- `.claude/` · `.moai/` 아래 다른 신규·변경 파일이 있으면 함께 미러한다.
- 닫힘 조건: AC-CSC-015, `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...`

### M6 — 접두 불변식 가드 (Priority Medium, 독립적)

배포 스킬 이름이 전부 **`moai` 접두(하이픈 없음)** 를 갖는다는 불변식을 테스트로 고정한다(REQ-CSC-015).

[HARD] **이 불변식은 코드에 없다.** iter-3 까지 여기 "코드에 암묵적으로 존재한다(`moaiSkillPrefix` 필터)"고 적혀 있었는데 틀렸다 — 그 상수는 `moai-`(하이픈 포함)이고, 카탈로그에는 이름이 정확히 `moai` 인 스킬이 있어 필터는 33개만 통과시킨다(spec §A.9). 따라서 `template.EmbeddedMoaiSkillNames()` 를 이 가드에도, 미러 집합 산출에도 **쓰지 않는다**. 재사용하면 `moai` 스킬 하나가 조용히 빠져 REQ-CSC-006 이 깨진다.

- 닫힘 조건: AC-CSC-016
- M1~M5 와 독립이므로 병행 가능하다.

## §G. 안티패턴

- **AP-1** — 템플릿 트리에 심볼릭 링크를 넣는다. 빌드 오류 없이 스킬이 사라진다(spec §A.2).
- **AP-2** — 미러 대상 수를 상수(32 / 34 / 36)로 박는다. 슬림 프로젝트에서 깨진 링크를 만든다(spec §A.3).
- **AP-3** — 폴백을 조용히 처리한다. R3.
- **AP-4** — 미러 실패 시 배포 전체를 실패시킨다. Claude Code 경로까지 잃는 과잉 반응이다(REQ-CSC-011).
- **AP-5** — 절대 경로 심볼릭 링크. 프로젝트 이동 시 끊긴다.
- **AP-6** — `.agents/` 또는 `.agents/skills/` 전체를 청소 대상으로 등록한다. 사용자 소유 항목을 삼킨다(REQ-CSC-009, spec §B.D5).
- **AP-7** — `ls | wc -l` 로 개수를 센다. 셸 별칭 때문에 +2 되는 형태가 이미 두 번 발생했다(spec §A.1). SPEC 자신도 이 형태로 §A.3 부분합을 틀렸다.
- **AP-8** — `fs.DirEntry.IsDir()` 로 스킬 디렉터리를 수집한다. `Lstat` 기반이라 디렉터리 링크가 **빠지고**, 그러면 소실 검사의 등식이 유지되어 테스트가 통과한다(AC-CSC-001 2번 단언).
- **AP-9** — 테스트가 "변경 전 커밋 기준선"을 참조한다. Go 테스트는 이전 커밋 코드로 배포할 수 없다. 불변식은 **같은 프로세스 안에서** 얻는다(AC-CSC-010).
- **AP-10** — `os.Symlink` 의 `EEXIST` 를 "지우고 다시 만든다"로 처리한다. 대상이 사용자 실 디렉터리면 데이터 손실이다(REQ-CSC-014).
- **AP-11** — 미러 항목을 `manifest.Track` 에 넘긴다. 디렉터리 링크에서 EISDIR 로 실패하고 fail-open 과 충돌한다(spec §A.6).
- **AP-12** — 청소 테스트에서 **제거만** 단언한다. `.agents/skills/` 전체를 지우고 있어도 통과하므로, 사용자 소유 스킬이 사라지는 상태를 검출하지 못한다. 제거와 생존은 같은 테스트 안에 있어야 한다(AC-CSC-008).
- **AP-13** — 청소 대상의 존재를 `os.Stat` 으로 판정한다. dangling 링크에서 `IsNotExist` 가 참이 되어 **찾아내고도 지우지 않는다**(spec §A.10).
- **AP-14** — 청소 테스트 fixture 를 **실 디렉터리로만** 심는다. 실제 산출물은 링크이고, 실 디렉터리는 `os.Stat` 이 성공하므로 dangling 결함이 살아 있어도 통과한다(AC-CSC-008).
- **AP-15** — `EmbeddedMoaiSkillNames()`(접두 `moai-`)를 미러 집합·가드·정리 판별에 재사용한다. 이름이 정확히 `moai` 인 스킬이 조용히 빠진다(spec §A.9).
- **AP-16** — 백업 금지를 `.agents/` 전체에 절대 형태로 적용한다. 사용자의 `moai` 접두 실 항목이 경고도 백업도 없이 사라진다(spec §A.11).

## §H. 부록 — `~/.codex/skills/` 정리 방안 (문서화만, 실행은 별도 승인)

**본 SPEC 의 run-phase 는 이 절을 실행하지 않는다.** spec §D 참조.

- **오염 내용**: 사용자 홈 `~/.codex/skills/` 에 2026-06-07 자 구 moai 스킬이 잔존하며, `moai-lang-*` · `moai-platform-*` 등 현재 카탈로그에 없는 이름이 다수 포함된다(M0 실측).
- **판별 기준**: `~/.codex/skills/` 에 존재하지만 **`internal/template/templates/.claude/skills/*/` 의 디렉터리 이름 집합**에 없는 `moai*`(하이픈 없는 접두) 항목만 대상. 이름이 그 집합에 있으면 건드리지 않는다.
- **[HARD] 판별자는 위 하나뿐이다.** `template.EmbeddedMoaiSkillNames()` 를 쓰지 않는다 — 그 함수의 접두 상수는 `moai-`(하이픈 포함)라 이름이 정확히 `moai` 인 현역 통합 스킬을 집합에서 빼고, 그러면 사용자 홈의 `~/.codex/skills/moai` 가 **삭제 후보로 분류된다**. 아래 절차 1단계(목록 출력)가 오염된 목록을 사람 앞에 내놓게 되어, 승인 근거 자체가 무너진다. 두 판별자를 "또는"으로 병기했던 iter-3 문구는 그래서 잘못이었다(spec §A.9).
- **절대 제외**: `~/.codex/skills/.system` (Codex 소유).
- **삭제 전 필수 절차**: (1) 대상 목록을 먼저 출력해 사람이 읽는다, (2) 백업을 뜬다, (3) 운영자 승인을 받는다, (4) 그 다음에 지운다. 목록 출력 없이 지우는 형태는 금지한다.
- **후속 카드 후보**: 이 정리를 `moai doctor` 의 능동 점검 항목으로 승격할지 여부는 별도 판단이다.
