# Acceptance — SPEC-REPORTS-LIFECYCLE-001

측정 기준 트리: `.moai/worktrees/t1320` @ `6879cfa5e`. Block 등급 RED 값은 **본 러닝에서 명령을 실행해 관측한 전사**(증거: `.moai/reports/t1320/red-baseline-delta.txt`, blob 목록: `red-blobs-before.txt`)로 채웠다. 이 워크트리에서 재측정 불가한 것(27 트리 오염량, primary `.moai/reports/` 564MB)은 **위임자 측정 인용**으로 표기한다 — 그 둘의 구분이 이 문서의 판정 신뢰성이다.

## §D AC 매트릭스

| AC | REQ | RED (6879cfa5e, 실측) | GREEN (목표) |
|---|---|---|---|
| AC-RLC-001 | REQ-RLC-001 | `git ls-files reports/` = **511** / historical = 0 | reports/ = 0 / historical = 이동 전 집합과 동일 |
| AC-RLC-002 | REQ-RLC-001 | — (이동 전 상태) | blob SHA 총합 불변 + `--follow` 이력 도달 |
| AC-RLC-003 | REQ-RLC-002 | historical 부재 | 신규 파일 무시 + tracked 유지 + D 0건 |
| AC-RLC-004 | REQ-RLC-008 | reports 무시행 가드 테스트 **부재** | 가드 통과 + 변이에서 붉음 |
| AC-RLC-005 | REQ-RLC-003, 009 | `<cwd>/reports/` 양 미러 **각 2건** (65·74행) | 0건 + `.moai/reports/` 존재 + build 드리프트 클린 |
| AC-RLC-006 | REQ-RLC-003, 004 | 렌더가 `<cwd>/reports/` 에 기록 | `.moai/reports/` 기록 + 디렉터리 자동 생성 |
| AC-RLC-007 | REQ-RLC-005 | done.go hoist **0건** (grep exit 1 실측) | hoist 동사 인출 + 건수/바이트 출력; done(L2) 은 같은 루틴 호출 |
| AC-RLC-008 | REQ-RLC-006 | 충돌 경로 자체 부재 | 미덮어쓰기 + 미해결 경로 보고 |
| AC-RLC-009 | REQ-RLC-007 | reports-archive 액션 **부재** | 술어 적합 항목 move + 건수/바이트 출력, 비적합 무영향 |
| AC-RLC-010 | REQ-RLC-007 | 보호 제외 로직 부재 | 보호 대상(historical·plan-audit·worktrees·archive·tracked 포함) 무영향 |
| AC-RLC-011 | REQ-RLC-009 | — (변경 전) | 미러 변경분 카드 유래물 0건 |
| AC-RLC-012 | REQ-RLC-005 | 폐기 플로우 hoist 의무 **0건** (grep 실측) | 양 사본에 의무 조항 존재 |

## §D.1 AC 상세

### AC-RLC-001 — 이동 완전성 (tracked 집합 보존)

- **Given** M1 착수 시점에 재측정한 `git ls-files reports/` 집합 R 이 있고,
- **When** 이동 커밋이 착지하면,
- **Then** `git ls-files reports/` 가 0줄이고, `git ls-files .moai/reports/historical/` 의 각 경로에서 `.moai/reports/historical/` 접두를 벗긴 집합이 R 과 정확히 일치한다(diff 0).

판정: `git ls-files reports/ | wc -l` · `git ls-files .moai/reports/historical/ | sed 's|^\.moai/reports/historical/||' | sort` vs R 정렬집합 diff.
**RED(본 러닝 전사, exit 0)**: `$ git ls-files reports/ | wc -l` → `511`. (증거: `.moai/reports/t1320/red-baseline-delta.txt` §R1)

### AC-RLC-002 — 이력·원격 보존

- **Given** 이동이 `git mv` 로 수행돼 인덱스에 스테이지됐고,
- **When** (a) `git ls-files -s reports/` 에서 추출한 blob SHA **정렬 목록**(이동 전, `red-blobs-before.txt` — 511행, digest `113a9ea0037d4df5141ebb4b3c90ef8b6f7159d9`)과 `git ls-files -s .moai/reports/historical/` 의 같은 목록을 비교하고, (b) `git diff --cached --stat` 의 모든 행을 검사하면,
- **Then** (a) 두 정렬 목록이 바이트 동일(diff 0 — blob 집합 불변 = 내용 무손상)이고, (b) 모든 행이 rename(`R100`)이며 내용 수정(`M`) 행이 0개다. 부가로 표본 파일(최상위 각 디렉터리 1개 이상, ≥10개)에 `git log --follow --oneline` 이 이동 이전 커밋에 도달한다.

판정 명령: `git ls-files -s <path> | awk '{print $2}' | sort | shasum` 이동 전후 비교 + `git diff --cached --stat | grep -cv '=>'` = 0.
**iter1 판정 명령 기각 기록**: `awk '{s+=$2}'` 숫자 합산은 16진 SHA 를 수치 강제해 0으로 붕괴한다(감사 실측: `abc123def` → 0) — 오염된 이동도 통과할 수 있어 폐기. 필드 추출(`{print $2}`)은 강제가 없으므로 안전하다.

**RED(본 러닝 전사, exit 0)**: 이동 전 상태 — `git ls-files -s reports/ | awk '{print $2}' | sort | wc -l` → `511`, digest `113a9ea0037d4df5141ebb4b3c90ef8b6f7159d9`. (증거: `.moai/reports/t1320/red-blobs-before.txt`)

### AC-RLC-003 — 무시 규칙 말단 상태

- **Given** 이동이 착지한 트리가 있고,
- **When** `git check-ignore -v` 매트릭스를 돌리면,
- **Then** `.moai/reports/newfile.md` → 무시(규칙 `.moai/reports/*` 또는 `:399`), `.moai/reports/historical/newfile.md` → 무시, tracked 이행 파일은 `git status --porcelain` 에 나타나지 않으며 `git status --porcelain | grep -c '^ D'` = 0.

### AC-RLC-004 — 무시 규칙 회귀 가드

- **Given** reports/worktrees 무시 매트릭스 가드 테스트가 소속 패키지에 존재하고,
- **When** 테스트를 돌리면 통과하고,
- **When** 규칙 한 줄(예: `.moai/reports/*`)을 삭제한 사본 트리에 대해 같은 가드를 돌리면,
- **Then** 실패한다(변이 주입 관측 — "붉어지는 것"이 증거로 남는다).

### AC-RLC-005 — skill 기본 경로 양 미러 패리티

- **Given** M2 커밋이 착지했고,
- **When** `grep -c '<cwd>/reports/'` 를 양 SKILL.md 에 대해 돌리면,
- **Then** 각각 0이고, `.moai/reports/` 가 65행 표와 74행 Output 절 양쪽에 존재하며, `make build` 후 emit-drift 검사가 클린하다.
**RED(본 러닝 전사, exit 0)**: `$ grep -c '<cwd>/reports/' <local SKILL.md> <mirror SKILL.md>` → `…:2` / `…:2` (양 사본 각 2건, 65·74행). (증거: `.moai/reports/t1320/red-baseline-delta.txt` §R3)

### AC-RLC-006 — skill 동작 관측

- **Given** output_path 미지정 렌더 요청과 존재하지 않는 `.moai/reports/` 를 갖는 임시 프로젝트가 있고,
- **When** 렌더 플로우를 실행하면,
- **Then** `.moai/reports/<slug>-<YYYYMMDD>.html` 과 `.md` 가 생성된다(디렉터리 자동 생성 포함).

### AC-RLC-007 — hoist 동사 정상 경로 + done(L2) 배선

- **Given** `.moai/reports/evidence.md` 를 포함하는 트리 디렉터리(임시 프로젝트 내)가 있고,
- **When** `moai worktree hoist <tree-path>` 를 실행하면,
- **Then** 루트 `.moai/reports/worktrees/<tree-name>/evidence.md` 가 존재하고, 출력에 인출 건수와 바이트가 포함된다.
- **And** `moai worktree done` 의 L2 트리 제거 경로가 같은 hoist 루틴을 제거 **전에** 호출한다(t.TempDir 테스트에서 L2 트리 제거 시 인출 파일 관측).
**RED(본 러닝 전사)**: `$ grep -c hoist internal/cli/worktree/done.go` → `0`, exit `1`. 부가 실측 — done 은 L1 트리를 두 경로에서 거부(`refuseL1SessionWorktree(targetPath)` 적중 `:80`·`:284`, SPEC-WORKTREE-DONE-TIER-001): 즉 hoist 의 실행 메커니즘은 독립 동사여야 한다(plan §F.0 D2). (증거: `.moai/reports/t1320/red-baseline-delta.txt` §R4·§R5)

### AC-RLC-008 — hoist 충돌 보호

- **Given** hoist 목적지에 같은 상대경로·다른 내용의 파일이 이미 있고,
- **When** hoist 가 실행되면,
- **Then** 기존 파일이 변경되지 않고(바이트 동일), 미해결 경로 각각이 출력에 보고된다.

### AC-RLC-009 — 아카이브 move 동작 (default-deny 술어)

- **Given** 임시 프로젝트의 `.moai/reports/` 아래에 술어 적합 후보(mtime 90일 전인 `t-old/`)와 비적합 항목(어제의 `t-new/`, 이름 패턴 불일치 `random-notes/`)이 있고,
- **When** reports-archive 액션을 실행하면,
- **Then** `t-old` 만 `.moai/reports/archive/<YYYY-MM>/t-old/` 로 **이동**(원본 부재)되고 출력에 건수/바이트가 있으며, `t-new`·`random-notes` 는 제자리다.

### AC-RLC-010 — 아카이브 보호 대상 (명시 보호 + tracked 자동 보호)

- **Given** 같은 임시 프로젝트에 `.moai/reports/historical/`, `.moai/reports/plan-audit/.gitkeep`, `.moai/reports/worktrees/`(hoist 산출), `.moai/reports/archive/` 기존분, 그리고 tracked 파일을 포함하는 항목(fixture `t338`형)이 있고,
- **When** reports-archive 액션을 실행하면,
- **Then** 셋 모두 제자리·내용 불변이다(전수 비교). tracked 포함 항목은 이름이 `^t[0-9]+$` 로 술어 1·2조건을 적중해도 `git ls-files` 비어있음 조건에서 배제된다.

### AC-RLC-011 — 템플릿 중립성

- **Given** 본 SPEC 의 템플릿 미러 변경분이 있고,
- **When** 변경 라인에서 `t1320`, `SPEC-REPORTS-LIFECYCLE`, 본 카드 생성일, 커밋 SHA 패턴을 grep 하면,
- **Then** 0건이다.

### AC-RLC-012 — 폐기 플로우 문서 반영 (실행 메커니즘과 결합)

- **Given** `worktree-integration.md` 로컬·템플릿 미러 양 사본이 있고,
- **When** hoist-before-dispose 의무 조항을 grep 하면,
- **Then** 양쪽에서 적중하며(사본 간 결함 대칭성), L1 세션 종료 폐기 경로가 **`moai worktree hoist` 호출을 절차로 명시**한다 — 조항 존재만으로 GREEN 하지 않고, AC-RLC-007 이 검증한 동사를 참조해야 실행 메커니즘이 성립한다(done 은 L1 을 거부하므로 동사가 유일한 실행 경로다).
**RED(본 러닝 전사, exit 1)**: `$ grep -c hoist <양 사본>` → `…:0` / `…:0`. (증거: `.moai/reports/t1320/red-baseline-delta.txt` §R6)

## §D.2 심각도

- **Block**: AC-RLC-001, AC-RLC-002 (데이터 소실 직결), AC-RLC-005 (양 미러 패리티 — Template-First 위반형)
- **Major**: AC-RLC-003, 004, 006, 007, 008, 009, 010, 011
- **Minor**: AC-RLC-012 (문서이나 REQ 본체의 후반부라 Minor 아님 없음 — 단, 미반영 시 sync 재작업)

## §D.3 추적성

| REQ | AC |
|---|---|
| REQ-RLC-001 | AC-RLC-001, AC-RLC-002 |
| REQ-RLC-002 | AC-RLC-003 |
| REQ-RLC-003 | AC-RLC-005, AC-RLC-006 |
| REQ-RLC-004 | AC-RLC-006 |
| REQ-RLC-005 | AC-RLC-007, AC-RLC-012 |
| REQ-RLC-006 | AC-RLC-008 |
| REQ-RLC-007 | AC-RLC-009, AC-RLC-010 |
| REQ-RLC-008 | AC-RLC-004 |
| REQ-RLC-009 | AC-RLC-005, AC-RLC-011 |

전 REQ 가 ≥1 AC 를 갖는다(고아 REQ 없음). 전 AC 가 ≥1 REQ 에 귀속된다(고아 AC 없음).

## §D.4 간접 검증 항목

- `moai doctor` 가 관련 체크에서 Fail 0 — 템플릿 재배포 축이 깨지지 않았는지의 간접 신호.
- `make embed-check` — 임베드 축(바이너리 vs 커밋본) 일치.
- `git status --porcelain` 이 M1 이후 reports 관련 항목을 영구히 나르지 않는다 — 무시 규칙 상호작용의 조기 신호.

## §D.5 클로저 게이트

1. 전 AC 판정 명령의 출력이 `progress.md` §E.2 에 전사돼 있다 — 요약 아님.
2. Block 등급 AC 전부 GREEN.
3. iter1 의 미해결 요구 확인 마커 2건(D1 재포함 여부, D4 파라미터)이 progress §E.1 의 레인 채택 기록으로 해소돼 있다(채택 근거·기각 기록: plan §F.0).
4. 미러 변경 커밋마다 `make build` 흔적이 커밋 이력에 존재한다.

## §D.6 선언된 미측정 (Gaps — plan-phase 기준)

- 27 트리 × 8.9~12MB 워크트리 오염량: **위임자 측정 인용**, 본 러닝 미재측정(대상이 primary 의 `.moai/worktrees/`).
- primary `.moai/reports/` 2,565항목/564MB: 동일하게 인용 — REQ 설계에 쓰이지 않았으므로 영향 없음.
- M1 착수 시점의 tracked 집합 R: 6879cfa5e 시점 511 — 타 카드 병합으로 변동 가능, M1 pre-flight 에서 재측정이 AC-RLC-001의 전제다.
