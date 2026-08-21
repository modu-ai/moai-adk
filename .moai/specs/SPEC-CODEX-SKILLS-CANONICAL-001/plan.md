# plan — SPEC-CODEX-SKILLS-CANONICAL-001

> 마일스톤은 **되돌리기 어려운 결정부터** 배치했다. M1·M2 가 뒤집히면 그 아래가 전부 다시 작성된다.
>
> **v0.6.0 범위 축소**: 청소 계열(구 M4 + M3 의 백업 절)은 승계 SPEC 으로 전출했다. 마일스톤 번호는 옮기지 않아 **M4 자리에 구멍이 있다** — 감사 보고서가 인용하는 번호를 보존하기 위해서다(spec HISTORY iter-6). 위험표와 안티패턴 목록에도 같은 이유로 구멍이 있다(R4·R5·R8·R13·R14·R16, AP-6·AP-12·AP-13·AP-14·AP-16).

## §A. 맥락

Codex CLI 는 `<repo>/.agents/skills/`(CWD→상향 병합)와 `$CODEX_HOME/skills/` 만 스캔하며 `.claude/skills/` 를 보지 않는다(M0 실측). moai 가 배포하는 스킬을 Codex 에서도 보이게 하려면 배포 시점에 `.agents/skills/` 경로를 추가로 만들어야 한다.

핵심 제약은 `//go:embed` 가 심볼릭 링크를 무음으로 버린다는 사실이다(spec §A.2). 링크는 템플릿이 아니라 배포기가 만든다.

## §B. 알려진 이슈 / 위험

| # | 위험 | 등급 | 완화 |
|---|---|---|---|
| R1 | 상수 기반 미러 집합이 슬림 프로젝트에서 깨진 링크를 만든다 | 높음 | REQ-CSC-006 + AC-CSC-014 (합성 FS 2-스킬 테스트) |
| R2 | Windows 에서 `os.Symlink` 실패 → 스킬 미노출 | 높음 | 복사 폴백(REQ-CSC-004) + `GOOS=windows go vet` 게이트 |
| R3 | 폴백이 조용해 "링크인 줄 알았는데 복사본" 상태가 잠복 | 중간 | REQ-CSC-005 관측 가능성 + AC-CSC-006 |
| R6 | 미러 실패가 배포 전체를 실패시켜 Claude Code 경로까지 잃는다 | 중간 | fail-open(REQ-CSC-011) + AC-CSC-013 |
| R7 | 템플릿 트리에 누군가 심볼릭 링크를 다시 넣어 스킬이 무음 소실 | 중간 | REQ-CSC-002 + AC-CSC-001 **양팔**(트리 전체 카운트 + 링크 가시 수집) |
| R9 | `.agents/` 가 사용자 저장소에 커밋됨 — Windows 체크아웃에서 링크가 텍스트 파일로 실체화 | **높음** | REQ-CSC-016 + AC-CSC-015. 근거 spec §A.8 |
| R10 | 미러 대상에 사용자 실 디렉터리가 있는데 `EEXIST` 를 "지우고 재생성"으로 처리 → 데이터 손실 | **높음** | REQ-CSC-014 + AC-CSC-011(3번 단언) |
| R11 | 비-`moai` 이름 스킬이 카탈로그에 들어와, 여기서 만든 미러에 승계 SPEC 의 청소가 닿지 못함 | 중간 | REQ-CSC-015 + AC-CSC-016. 근거 spec §A.9 |
| R12 | `manifest.Track` 이 디렉터리 링크에서 EISDIR → `Deploy` 실패, fail-open 과 충돌 | 중간 | REQ-CSC-010(기록 금지) + AC-CSC-012(1번 팔). 근거 spec §A.6 |
| R15 | 출력 seam 부재로 모드·경고가 관측 불가 → MUST AC 3개 판정 불능 | 중간 | REQ-CSC-005(반환 결과) + M2 산출물. 근거 spec §B.D3 + §A.9b |
| R17 | `moai-`(하이픈) 접두 seam 재사용으로 `moai` 스킬이 미러·가드·정리에서 누락 | 중간 | REQ-CSC-015 철자 고정 + AC-CSC-016 [HARD] + M6·§H. 근거 spec §A.9 |

## §C. 사전 점검 (착수 전)

- `git rev-parse --show-toplevel` 이 이 worktree 인지 확인한다.
- `find internal/template/templates -type l | wc -l` == 0 (기준선, **트리 전체** — `.claude/skills/` 하위가 아니다).
- `grep -n 'agents' internal/template/templates/.gitignore` — 착수 시점에 항목이 없음을 확인한다(M5 의 전제).
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
- 이 결정이 바뀌면 M2·M3·M5·M6 전부 재작성이다.

### M2 — 링크 생성 + 폴백 (Priority High)

`.agents/skills/<name>` 을 `../../.claude/skills/<name>` 상대 심볼릭 링크로 만들고, 실패 시 복사 폴백 + 모드 보고. 실패-실패 시 fail-open.

- 대상: M1 이 만든 파일 + **출력 seam 신설**
- **[HARD] 출력 seam 은 이 마일스톤의 산출물이다.** iter-3 까지 여기 "출력 경로는 기존 printer 계층 재사용"이라고 적혀 있었는데 **사실이 아니다** — 실측하면 `internal/template` 에 `io.Writer` 가 없고 `Deploy` 는 `error` 만 돌려주며 패키지 전체에 printer 계층이 없다. 출력 표면은 호출자인 `internal/cli/` 에만 있다. 따라서 모드·경고를 **반환 결과로 올리는 seam** 을 여기서 만든다(REQ-CSC-005 / spec §B.D3 + §A.9b — 그 실측은 §A 에 절로 승격돼 있다). 이것을 빠뜨리면 AC-CSC-006·011(3)·013(3) 세 개가 함께 막히고 그중 둘은 MUST 다.
- 닫힘 조건: AC-CSC-002, AC-CSC-003, AC-CSC-004, AC-CSC-005, AC-CSC-006, AC-CSC-011, AC-CSC-013
- 상대 링크를 쓰는 이유: 프로젝트 디렉터리를 통째로 옮기거나 복사해도 링크가 살아남는다. 절대 경로는 `moai init` 시점 경로로 굳는다(CLAUDE.local.md §14 와 같은 실패 형태).

### M3 — 미러를 manifest 기록 대상에서 제외 (Priority High)

**방향이 두 번 바뀐 자리다.** iter-2 의 M3 는 "미러를 manifest 에 기록한다"였는데 실측 결과 그 설계는 기존 seam 에서 **구현 불가능**하다(spec §A.6 — `Track` → `HashFile` → `io.Copy` 가 디렉터리 링크에서 EISDIR). 기록을 시도하면 error 가 올라와 REQ-CSC-011 의 fail-open 과 충돌한다. 그래서 iter-3 이 "기록하지 않는다"로 뒤집었고, v0.6.0 범위 축소에서 **백업 관련 작업은 승계 SPEC 으로 전출**해 이 마일스톤에는 manifest 한 가지만 남았다.

하는 일: 미러 항목에 대해 `manifest.Track` 을 **호출하지 않는다.** 배포기가 정본 파일을 기록하는 기존 경로는 그대로 두고, 미러 생성 경로만 기록 호출을 갖지 않게 한다.

- 닫힘 조건: AC-CSC-012 (manifest 부재)
- pre-clean 백업 정책(`backupThenRemove` 판별 분기)은 **이 SPEC 의 작업이 아니다** — 승계 SPEC 소관(spec HISTORY iter-6 전출 목록).

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
- **AP-7** — `ls | wc -l` 로 개수를 센다. 셸 별칭 때문에 +2 되는 형태가 이미 두 번 발생했다(spec §A.1). SPEC 자신도 이 형태로 §A.3 부분합을 틀렸다.
- **AP-8** — `fs.DirEntry.IsDir()` 로 스킬 디렉터리를 수집한다. `Lstat` 기반이라 디렉터리 링크가 **빠지고**, 그러면 소실 검사의 등식이 유지되어 테스트가 통과한다(AC-CSC-001 2번 단언).
- **AP-9** — 테스트가 "변경 전 커밋 기준선"을 참조한다. Go 테스트는 이전 커밋 코드로 배포할 수 없다. 불변식은 **같은 프로세스 안에서** 얻는다(AC-CSC-010).
- **AP-10** — `os.Symlink` 의 `EEXIST` 를 "지우고 다시 만든다"로 처리한다. 대상이 사용자 실 디렉터리면 데이터 손실이다(REQ-CSC-014).
- **AP-11** — 미러 항목을 `manifest.Track` 에 넘긴다. 디렉터리 링크에서 EISDIR 로 실패하고 fail-open 과 충돌한다(spec §A.6).
- **AP-15** — `EmbeddedMoaiSkillNames()`(접두 `moai-`)를 미러 집합·가드·정리 판별에 재사용한다. 이름이 정확히 `moai` 인 스킬이 조용히 빠진다(spec §A.9).

## §H. 부록 — `~/.codex/skills/` 정리 방안 (문서화만, 실행은 별도 승인)

**본 SPEC 의 run-phase 는 이 절을 실행하지 않는다.** spec §D 참조.

- **오염 내용**: 사용자 홈 `~/.codex/skills/` 에 2026-06-07 자 구 moai 스킬이 잔존하며, `moai-lang-*` · `moai-platform-*` 등 현재 카탈로그에 없는 이름이 다수 포함된다(M0 실측).
- **판별 기준**: `~/.codex/skills/` 에 존재하지만 **`internal/template/templates/.claude/skills/*/` 의 디렉터리 이름 집합**에 없는 `moai*`(하이픈 없는 접두) 항목만 대상. 이름이 그 집합에 있으면 건드리지 않는다.
- **[HARD] 판별자는 위 하나뿐이다.** `template.EmbeddedMoaiSkillNames()` 를 쓰지 않는다 — 그 함수의 접두 상수는 `moai-`(하이픈 포함)라 이름이 정확히 `moai` 인 현역 통합 스킬을 집합에서 빼고, 그러면 사용자 홈의 `~/.codex/skills/moai` 가 **삭제 후보로 분류된다**. 아래 절차 1단계(목록 출력)가 오염된 목록을 사람 앞에 내놓게 되어, 승인 근거 자체가 무너진다. 두 판별자를 "또는"으로 병기했던 iter-3 문구는 그래서 잘못이었다(spec §A.9).
- **절대 제외**: `~/.codex/skills/.system` (Codex 소유).
- **삭제 전 필수 절차**: (1) 대상 목록을 먼저 출력해 사람이 읽는다, (2) 백업을 뜬다, (3) 운영자 승인을 받는다, (4) 그 다음에 지운다. 목록 출력 없이 지우는 형태는 금지한다.
- **후속 카드 후보**: 이 정리를 `moai doctor` 의 능동 점검 항목으로 승격할지 여부는 별도 판단이다.
