# acceptance — SPEC-CODEX-SKILLS-CANONICAL-001

Tier M 상한: 요구사항 16 / 판정 기준 16. 본 문서는 **16개**를 담는다(상한 도달).

모든 항목은 Given-When-Then 이며 기계적으로 참·거짓이 갈려야 한다. 명령을 실행하고 출력을 인용하지 않은 항목은 통과로 셈하지 않는다.

## §D. AC 매트릭스

| AC | 검증 대상 REQ | 형태 |
|---|---|---|
| AC-CSC-001 | REQ-CSC-002 | Go 테스트 (링크 가시 수집 + 트리 전체 카운트) |
| AC-CSC-002 | REQ-CSC-001 | Go 테스트 (전량 배포) |
| AC-CSC-003 | REQ-CSC-001, REQ-CSC-006 | Go 테스트 (슬림 배포 + 전량 대비 관계) |
| AC-CSC-004 | REQ-CSC-003 | Go 테스트 |
| AC-CSC-005 | REQ-CSC-004 | Go 테스트 (폴백 주입) |
| AC-CSC-006 | REQ-CSC-005 | Go 테스트 (반환 결과 관측, 양방향) |
| AC-CSC-007 | REQ-CSC-008 | Go 테스트 (구분자 중립 + 슬라이스 순서) |
| AC-CSC-008 | REQ-CSC-008, REQ-CSC-009 | Go 테스트 (**미러 4형태 단일 테스트** — dangling 포함) |
| AC-CSC-009 | REQ-CSC-009 | Go 테스트 (사용자 소유 네임스페이스 전반 생존) |
| AC-CSC-010 | REQ-CSC-007 | Go 테스트 (**동일 프로세스 seam 토글 불변식**) |
| AC-CSC-011 | REQ-CSC-012, REQ-CSC-013, REQ-CSC-014 | Go 테스트 (대상 선점 3-상태) |
| AC-CSC-012 | REQ-CSC-010 | Go 테스트 (manifest 부재 + 재배포분 백업 부재 + 은퇴분 백업 보존) |
| AC-CSC-013 | REQ-CSC-011 | Go 테스트 (fail-open, AC-010 불변식 참조) |
| AC-CSC-014 | REQ-CSC-006 | Go 테스트 (합성 FS) |
| AC-CSC-015 | REQ-CSC-016 + 비기능 | Go 테스트 (`.gitignore`) + 기존 게이트 재실행 |
| AC-CSC-016 | REQ-CSC-015 | Go 테스트 (접두 불변식) |

## §D.1 판정 기준

### AC-CSC-001 — 임베드 심볼릭 링크 소실 가드 (양팔)

**Given** 템플릿 소스 트리 `internal/template/templates/` 와 임베드 FS `template.EmbeddedTemplates()` 가 있고,
**When** 아래 두 수집을 각각 수행하면,
**Then** 두 단언이 모두 참이어야 한다.

1. **트리 전체 링크 카운트 == 0.** `internal/template/templates/` **전체**를 `filepath.WalkDir` 로 걷되 각 항목을 `d.Type()&fs.ModeSymlink != 0` 로 판정해 심볼릭 링크를 센다. 결과가 0 이어야 한다. 범위는 `.claude/skills/` 가 아니라 **트리 전체**다 — REQ-CSC-002 의 범위가 그렇고, `.claude/agents/` 나 `.claude/rules/` 에 들어간 링크도 같은 무음 소실을 일으킨다.
2. **집합 동치.** 파일시스템 쪽 `.claude/skills/` 1단계 항목을 수집할 때 **`d.IsDir()` 를 쓰지 않는다.** `fs.DirEntry` 는 `Lstat` 기반이라 디렉터리를 가리키는 링크의 `IsDir()` 는 `false` 이고, 그러면 링크가 파일시스템 집합에서도 빠져 임베드 집합과의 등식이 유지된다 — **테스트가 막으려는 바로 그 상태에서 통과한다.** 수집은 `os.Stat`(링크를 따라가 디렉터리로 인식) 또는 "디렉터리 OR 심볼릭 링크" 판정으로 하고, 그렇게 모은 집합이 임베드 집합과 **정확히 같아야** 한다.

[HARD] 1번만으로 충분해 보인다는 이유로 2번의 수집 방식을 `d.IsDir()` 로 되돌리지 않는다. 두 단언은 서로의 사각을 덮는다.

이 항목은 개수 상수를 쓰지 않는다. 34 는 작성 시점 관측값일 뿐이고, 지키는 것은 "파일시스템에 있는 것이 임베드에도 있다"는 불변식이다.

### AC-CSC-002 — 전량 배포 시 양쪽 경로 도달 가능

**Given** `t.TempDir()` 프로젝트 루트와 전량(non-slim) 임베드 FS 로 구성한 배포기가 있고,
**When** `Deploy` 를 1회 실행하면,
**Then** 배포된 모든 스킬 `<name>` 에 대해 `.claude/skills/<name>/SKILL.md` 와 `.agents/skills/<name>/SKILL.md` 가 **둘 다 읽히고 내용이 동일**해야 한다.

단서: 여기서 "모든"은 REQ-CSC-001 의 예외 두 가지 — REQ-CSC-014 의 대상 선점, REQ-CSC-011 의 미러 생성 실패 — 가 발생하지 않은 스킬에 한한다. 이 AC 의 fixture 는 빈 프로젝트에 1회 배포하므로 두 예외 모두 발생하지 않으며, 예외 상태의 판정은 AC-CSC-011·013 이 담당한다.

### AC-CSC-003 — 슬림 배포 시 집합 동치, 그리고 tier 필터가 실제로 관통했음

**Given** 동일 구성에 slim FS(core tier 만) 를 물린 배포기와, 전량 FS 를 물린 배포기를 각각 별도 `t.TempDir()` 에 배포하고,
**When** 두 결과의 `.agents/skills/` · `.claude/skills/` 항목 집합을 수집하면,
**Then** 세 단언이 모두 참이어야 한다.

1. 슬림 쪽 `.agents/skills/` 집합 == 슬림 쪽 `.claude/skills/` 집합.
2. 슬림 쪽에 없는 non-core 스킬 이름이 슬림 쪽 `.agents/skills/` 에 **하나도 없다**(깨진 링크 0).
3. **슬림 집합의 크기 < 전량 집합의 크기.** 1번만으로는 "배포 후 대상 FS 를 다시 읽어 미러를 만드는" 구현에서 **정의상 항상 참**이 되어 tier 필터를 통과했는지 구분하지 못한다. 3번이 그 구분을 만든다.

### AC-CSC-004 — 링크는 상대 경로이며 정본을 가리킨다

**Given** AC-CSC-002 의 배포 결과가 있고 (심볼릭 링크 모드가 성공한 플랫폼에서),
**When** 임의의 `.agents/skills/<name>` 에 `os.Readlink` 를 호출하면,
**Then** 반환값이 `../../.claude/skills/<name>` 이어야 하며(절대 경로면 FAIL), 그 경로를 따라간 결과가 프로젝트 루트 내부의 정본 디렉터리여야 한다.

### AC-CSC-005 — 폴백 복사본이 읽힌다

**Given** 심볼릭 링크 생성이 실패하도록 주입된 배포기가 있고,
**When** `Deploy` 를 실행하면,
**Then** `.agents/skills/<name>/SKILL.md` 가 **실제 파일로** 존재하고 정본과 내용이 같아야 하며, `Deploy` 는 오류를 반환하지 않아야 한다.

### AC-CSC-006 — 폴백이 관측된다 (양방향, 반환 결과 기준)

**Given** AC-CSC-005 와 동일한 주입 조건에서,
**When** `Deploy` 의 **반환 결과에 담긴 모드·경고 정보**를 읽으면,
**Then** 복사 모드가 사용됐음이 그 결과에 나타나야 한다. 링크 모드가 성공한 실행에서는 **나타나지 않아야** 한다(양방향).

[HARD] "`Deploy` 의 출력 writer 를 버퍼로 캡처한다"는 형태로 쓰지 않는다. 실측하면 `internal/template` 에는 `io.Writer` 가 **없다** — `Deploy` 의 시그니처는 `(ctx, projectRoot, m manifest.Manager, tmplCtx *TemplateContext) error` 이고 패키지 전체에 printer 계층이 없다. 출력 표면은 호출자인 `internal/cli/` 계층에만 있다. 따라서 판정은 **배포기가 반환하는 값**을 대상으로 하며, 사용자 표시 여부는 호출부의 책임이다(REQ-CSC-005, §B.D6). 이 seam 은 run-phase 가 새로 만들어야 하는 것이며 plan M2 의 산출물이다.

### AC-CSC-007 — 청소 목록에 등록되어 있다 (경로 구분자 중립)

**Given** 임의의 `projectRoot` 가 있고,
**When** `deploy.ManagedCleanTargets(projectRoot)` 를 호출하면,
**Then** 두 단언이 모두 참이어야 한다.

1. `filepath.ToSlash(t.DisplayPath) == ".agents/skills/moai*"` 이고 `IsGlob == true` 인 항목이 **정확히 1개** 있다.
2. **그 항목의 인덱스가 `.claude/skills/moai*` 항목의 인덱스보다 작다** — 청소가 정본을 지우기 전에 미러를 처리하도록 순서를 고정한다(REQ-CSC-008, §B.D6 의 이중 방어 (b)). 이 단언이 없으면 순서는 나중에 조용히 바뀐다.

[HARD] 슬래시 리터럴 동치로 단언하지 않는다. 기존 항목은 전부 `filepath.Join` 으로 만들어지므로 Windows 에서 `DisplayPath` 는 백슬래시를 쓴다 — 리터럴 비교는 Windows CI 에서만 실패하고, §D.5 가 전량 판정을 CI 에 맡기므로 그 실패는 반드시 드러난다.

### AC-CSC-008 — 미러의 **네 형태**를 한 테스트에 심고 함께 단언한다

**Given** `t.TempDir()` 프로젝트 하나에 아래 네 항목을 **모두** 심어 두고,

| 항목 | 형태 | 비고 |
|---|---|---|
| `.agents/skills/moai-live` | 심볼릭 링크, 정본 `.claude/skills/moai-live/SKILL.md` **존재** | 정상 배포 산출물 |
| `.agents/skills/moai-gone` | 심볼릭 링크, 정본이 **이미 삭제됨(dangling)** | 실제 실행 순서가 만드는 상태 |
| `.agents/skills/moai-copied/SKILL.md` | **실 디렉터리** | 복사 모드 산출물 |
| `.agents/skills/hns-user-owned/SKILL.md` | **실 디렉터리** | 사용자 소유 |

**When** `CleanMoaiManagedPaths` 를 1회 실행하면,
**Then** 다음 네 단언이 **같은 테스트 안에서 모두** 참이어야 한다.

1. `moai-live` 가 제거됐다.
2. **`moai-gone` 이 제거됐다.**
3. `moai-copied` 가 제거됐다.
4. `hns-user-owned/SKILL.md` 가 그대로 읽히고 내용이 변하지 않았다.

추가로, **정본 `.claude/skills/moai-live/SKILL.md` 가 링크를 통해 삭제되지 않았다** — 청소가 링크를 따라가 정본을 지우는 형태를 함께 배제한다.

[HARD] **2번이 이 AC 의 존재 이유다.** iter-3 까지 이 AC 는 `.agents/skills/<name>/SKILL.md` 를 **실 파일로만** 심었다. 그런데 이 SPEC 이 실제로 배포하는 산출물은 심볼릭 링크이고, 실 디렉터리는 `os.Stat` 이 성공해 청소가 정상 동작하므로 — **§A.10 의 결함이 살아 있는 상태에서도 통과했다.** 놓친 축은 제거/생존이 아니라 **실 항목 / 링크 항목**이었다. 네 형태를 한 테스트에 두는 이유는 그 축을 fixture 안에 강제로 넣기 위해서다.

[HARD] 네 단언을 별개 테스트로 쪼개지 않는다. 제거 쪽만 단언하는 테스트는 청소가 `.agents/skills/` 전체를 삼키고 있어도 통과하고, 실 항목만 심는 테스트는 dangling 결함을 통과시킨다.

### AC-CSC-009 — 사용자 소유 네임스페이스 전반이 생존한다

**Given** 같은 프로젝트에 `.agents/skills/harness-legacy/`, `.agents/skills/my-harness-legacy/`, `.agents/skills/my-own/` 을 각각 `SKILL.md` 와 함께 심어 두고,
**When** `CleanMoaiManagedPaths` 를 실행하면,
**Then** 세 경로가 **모두 그대로 남아** 있어야 한다.

AC-CSC-008 이 계약의 대표 사례(`hns-*`)를 고정한다면, 이 AC 는 `moai*` 접두 밖 전체로 범위를 넓힌다 — legacy 두 세대와 계약에 이름조차 없는 임의 이름까지.

판정력의 한계를 적어 둔다: `filepath.Glob(".agents/skills/moai*")` 는 이 세 이름을 **원리적으로** 매치하지 않으므로, 올바르게 구현된 글롭에 대해 이 AC 는 반증되기 어렵다. 그럼에도 유지하는 이유는 AP-6 형태(`.agents/skills/*` 또는 `.agents/` 전체)를 잡기 때문이며, 그것이 이 AC 가 지키는 유일한 축이다. 두 감사자가 모두 "얇다"고 지적했고 그 판단은 옳다 — 얇다는 것을 알고 두는 것과 모르고 두는 것은 다르다.

### AC-CSC-010 — Claude Code 경로 무회귀 (동일 프로세스 seam 토글 불변식)

**Given** 미러 생성을 **끄고** 배포하는 seam 과 **켜고** 배포하는 seam 이 있고(같은 테스트 프로세스 안에서 둘 다 실행 가능해야 한다 — 이 실행 가능성 자체가 run-phase 의 설계 제약이다),
**When** 서로 다른 두 `t.TempDir()` 에 각각 배포한 뒤 양쪽의 `.claude/skills/` 산출물을 (상대경로, SHA-256, 퍼미션) 목록으로 수집하면,
**Then** 두 목록이 **완전히 동일**해야 한다. 파일 1개의 추가·삭제·바이트 변화·퍼미션 변화도 FAIL.

[HARD] "변경 전 커밋에서 얻은 기준선"과 대조하는 형태로 쓰지 않는다. Go 테스트는 이전 커밋을 체크아웃해 그 시점 코드로 배포할 수 없다 — 그렇게 쓰면 실행 가능한 테스트가 아니라 일회성 수동 절차가 되고, 변경이 착지한 뒤에는 재실행조차 불가능해 회귀 가드로 남지 않는다. 여기서 요구하는 것은 **같은 프로세스 안에서 양쪽을 실행해 얻는 불변식**이며, 이것은 착지 이후에도 매번 재실행된다.

커밋 SHA 를 명시한 일회성 수동 대조는 §D.4(간접 검증)로 옮겼다 — 유용하지만 AC 가 아니다.

### AC-CSC-011 — 대상 경로가 이미 점유된 세 상태

**Given** 1회차 배포가 끝난 `t.TempDir()` 프로젝트에서, 세 스킬의 미러 경로를 각각 다음 상태로 만들어 두고 — (i) 그대로(올바른 링크), (ii) 엉뚱한 곳을 가리키는 링크로 교체, (iii) 링크를 지우고 사용자 파일 `USER.md` 를 담은 **실 디렉터리**를 생성 —
**When** 같은 프로젝트 루트에 `Deploy` 를 한 번 더 실행하면,
**Then** 세 단언이 모두 참이어야 한다.

1. (i) 은 변경되지 않는다 — 링크 대상 동일, (복사 모드라면) 파일 해시 동일. `Deploy` 는 오류를 반환하지 않는다.
2. (ii) 는 올바른 정본을 가리키도록 **교체**되어 있다.
3. (iii) 의 `USER.md` 가 **그대로 읽힌다.** 그리고 건너뛰었다는 경고가 `Deploy` 의 **반환 결과**에 나타난다(AC-CSC-006 과 같은 seam — 출력 writer 가 아니다).

[HARD] 3번이 이 AC 의 핵심이다. `os.Symlink` 가 `EEXIST` 를 낼 때 가장 자연스러운 대응은 "지우고 다시 만든다"인데, 대상이 사용자의 실 디렉터리면 그 순간 **데이터 손실**이다. 이름이 `moai*` 로 한정된다는 사실은 이를 완화할 뿐 제거하지 않는다 — 사용자도 `moai-` 이름을 쓸 수 있다.

이 AC 는 clean 없는 재배포 경로(`moai init` 재실행 등)를 본다. 실제 `moai update` 는 clean → Deploy 순서라 (ii)·(iii) 상태를 만나지 않지만, 만나는 경로가 따로 존재하는 이상 규정되어야 한다.

### AC-CSC-012 — 미러는 기록되지도 백업되지도 않는다 (양팔)

**Given** AC-CSC-002 의 배포에 사용한 manifest manager 와, 그 뒤 `CleanMoaiManagedPaths` 를 실행한 결과가 있고,
**When** manifest 항목과 `.moai-backups/` 트리를 각각 조회하면,
**Then** 두 단언이 모두 참이어야 한다.

1. **manifest 부재.** manifest 에 `.agents/` 로 시작하는 키가 **0개**이고, `.claude/skills/` 항목의 기존 기록은 변하지 않았다.
2. **재배포될 미러는 백업되지 않는다.** 템플릿이 같은 이름의 스킬을 가진 미러 항목에 대해 `.moai-backups/**/pre-clean/.agents/` 아래 파일 수가 **0**이다. 링크 모드뿐 아니라 **복사 모드에서도** 0이어야 한다 — 이 단언의 존재 이유가 복사 모드다.
3. **재배포되지 않을 실 항목은 백업된다.** 템플릿에 대응 스킬이 없는 `.agents/skills/moai-retired/` 실 디렉터리를 심어 두면, 청소 후 그 내용이 `.moai-backups/**/pre-clean/.agents/skills/moai-retired/` 아래에 **보존되어 있어야** 한다.

세 단언의 관계가 이 AC 의 요점이다. 1·2번은 "미러는 정본의 파생물"이라는 원칙에서 나오고(spec §A.6·§A.7), **3번은 그 원칙을 절대 형태로 쓰면 열리는 손실 경로를 막는다**(spec §A.11). 청소 글롭은 `moai*` 이고 `moai` 접두는 사용자도 쓸 수 있으므로, 백업을 통째로 금지하면 사용자 항목이 경고도 백업도 없이 사라진다. 판별자는 "이번 실행이 다시 만들 것인가" — `backupThenRemove` 가 template-managed 파일을 백업하지 않는 기존 사유와 같은 기준이다.

2번만 있고 3번이 없으면 구현자는 `.agents/` 를 백업에서 통째로 제외해 AC 를 통과시키고, 그 순간 3번이 막으려는 손실이 발생한다.

### AC-CSC-013 — 미러 실패는 배포를 죽이지 않는다

**Given** 심볼릭 링크와 복사가 **둘 다** 실패하도록 주입된 배포기가 있고,
**When** `Deploy` 를 실행하면,
**Then** 세 단언이 모두 참이어야 한다.

1. `Deploy` 가 오류를 반환하지 않는다.
2. `.claude/skills/` 산출물이 **AC-CSC-010 의 불변식 형태**로 미러-비활성 seam 의 산출물과 (상대경로, SHA-256, 퍼미션) 동일하다 — 같은 프로세스 안에서 얻는 비교 대상이며, 커밋 기준선이 아니다.
3. 경고가 `Deploy` 의 **반환 결과**에 나타난다(AC-CSC-006 과 같은 seam).

### AC-CSC-014 — 미러 집합은 상수에서 파생되지 않는다

**Given** 스킬 **2개**만 담은 합성 `fstest.MapFS`(예: `moai-alpha`, `moai-beta`)로 구성한 배포기가 있고,
**When** `Deploy` 를 실행하면,
**Then** `.agents/skills/` 항목이 **정확히 2개**(`moai-alpha`, `moai-beta`)여야 한다. 3개 이상이거나 실제 배포 카탈로그 이름이 섞여 나오면 코드 어딘가가 상수·실제 카탈로그를 참조하고 있다는 뜻이므로 FAIL.

### AC-CSC-015 — `.gitignore` + Template-First / 중립성 게이트

**Given** 배포되는 템플릿 `.gitignore`(`internal/template/templates/.gitignore`) 와 임베드 FS 가 있고,
**When** 그 내용을 읽고 기존 게이트를 실행하면,
**Then** 세 단언이 모두 참이어야 한다.

1. `.gitignore` 가 **`.agents/skills/moai*`** 를 무시하는 항목을 담고 있다.
2. **`.agents/` 전체를 무시하는 항목은 담고 있지 않다** — 좁은-범위 원칙(§B.D7). 사용자의 `.agents/skills/hns-*` 와 후속 마일스톤의 `.agents/` 소스 파일이 추적에서 빠지면 안 된다.
3. 1번 항목이 임베드 FS 쪽에서도 읽힌다(`make build` 반영 확인 — 템플릿 소스만 고치고 빌드를 빼먹는 형태를 잡는다).
4. `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` 통과.

### AC-CSC-016 — 배포 스킬 이름 접두 불변식

**Given** 임베드 템플릿 FS 의 `.claude/skills/` 1단계 디렉터리 이름 전부를 수집하고,
**When** 각 이름이 **`moai`(하이픈 없음)** 로 시작하는지 검사하면,
**Then** 접두를 갖지 않는 이름이 **0개**여야 한다.

[HARD] `template.EmbeddedMoaiSkillNames()` 로 구현하지 않는다. 그 함수의 상수는 `moai-`(하이픈 포함)이고, 카탈로그에는 이름이 **정확히 `moai`** 인 통합 스킬이 있어 함수는 33개만 돌려준다 — 그것으로 검사하면 이 AC 는 대상 하나를 통째로 보지 못하고, 같은 함수를 미러 집합 산출에 재사용하면 `moai` 스킬이 조용히 빠져 REQ-CSC-006 이 깨진다. 수집은 디렉터리 목록에서 직접 한다(spec §A.9).

이 AC 가 지키는 것은 미러 집합(REQ-CSC-006, 이름 제약 없음)과 청소 집합(REQ-CSC-008, `moai*` 한정)의 일치다. 현재 그 일치는 우연이며(비-`moai` 이름 0개), `catalog.yaml` 의 `harness_generated.skills` 는 비어 있는 슬롯이다. 여기 비-`moai` 이름이 들어오면 미러는 만들어지고 청소는 못 지운다 — spec §A.9. 이 테스트는 그 순간 실패해 결정을 사람 앞으로 가져온다.

## §D.2 심각도

- **MUST (하나라도 FAIL 이면 run-phase 미종료)**: AC-CSC-001, 002, 003, 005, 006, 007, 008, 009, 010, 011, 012, 013, 014, 015, 016
- **SHOULD**: AC-CSC-004

AC-CSC-004 가 SHOULD 인 이유는 폴백 플랫폼에서 원리적으로 해당 없음이 되기 때문이다 — 링크 모드가 성공한 플랫폼에서는 MUST 로 취급한다.

iter-4 승격 두 건. **AC-CSC-006 → MUST**: 반환 결과 seam 이 REQ-CSC-005 에 확정됐으므로 더 이상 조건부가 아니고, 같은 seam 에 AC-011(3)·013(3) 이 걸려 있어 이것이 통과하지 못하면 그 둘도 판정 불가다. **AC-CSC-015 → MUST**: `.gitignore` 변경이 REQ-CSC-016 으로 확정돼 "변경이 있으면"이라는 조건이 사라졌다.

[HARD] REQ-CSC-007 의 "본 변경 전과 동일"에 대한 **착지 시점의 유일한 근거는 §D.4 의 1회 커밋 대조**다. AC-CSC-010 이 증명하는 것은 "미러 기능을 켜고 끈 것 사이가 같다"이며, 미러-비활성 seam 자체가 이전 동작과 달라지면 AC 는 통과하고 REQ 는 깨진다. 그래서 REQ-CSC-007 의 문구를 AC 가 실제로 증명하는 범위로 좁혔고(v0.4.0), 좁히면서 잃은 축을 §D.4 가 1회 덮는다 — 그 대조는 AC 가 아니므로 회귀 가드가 아니라는 점을 여기 명시한다.

## §D.3 추적성

REQ 16개 전부가 최소 1개 AC 에 대응한다: 001→AC-002/003, 002→AC-001, 003→AC-004, 004→AC-005, 005→AC-006, 006→AC-003/014, 007→AC-010, 008→AC-007/008, 009→AC-008(생존 팔)/AC-009, 010→AC-012, 011→AC-013, 012→AC-011(1), 013→AC-011(2), 014→AC-011(3), 015→AC-016, 016→AC-015.

REQ-CSC-008 은 세 절(글롭 등록 / `Lstat` 판정 + dangling 제거 / 슬라이스 순서)을 담으며 각각 다른 팔에서 판정된다 — 글롭 등록은 AC-CSC-007(1), 순서는 AC-CSC-007(2), dangling 제거는 AC-CSC-008(2). 절과 판정이 1:1 로 대응하므로 "어느 절이 실패했는가"가 실행 시점에 구분된다.

REQ-CSC-009 가 AC 두 개에 걸리는 것은 의도다 — AC-CSC-008 은 제거 단언과 **같은 테스트 안에** 묶어 "제거는 되는데 사용자 파일도 함께 사라지는" 상태를 잡고, AC-CSC-009 는 접두 밖 이름 전반으로 넓힌다. REQ-CSC-012·013·014 가 AC-CSC-011 하나에 모이는 것도 의도다 — 세 상태는 **같은 재배포 실행의 세 대상**이라 한 테스트에서 함께 관측해야 상호 간섭(예: (ii) 교체 로직이 (iii) 까지 지우는 형태)이 드러난다.

## §D.4 간접 검증 (AC 아님 — 실행하되 판정에는 넣지 않는다)

- **커밋 기준선 대조 (일회성 수동)**: 변경 전 커밋에서 `t.TempDir()` 에 배포해 얻은 `.claude/skills/` 산출물 목록과 착지 후 목록을 1회 대조하고, 커밋 SHA 와 함께 `progress.md` §E.2 에 기록한다. AC-CSC-010 의 불변식이 회귀 가드를 담당하고, 이 절차는 착지 시점의 1회 확인만 담당한다.
- **Windows 실동작**: 로컬에서 확인할 수 없다. `GOOS=windows go vet ./internal/...` 로 컴파일 가능성만 게이트하고, 실동작 판정은 CI 매트릭스에 맡긴다.
- **Codex 실제 노출**: CI 러너에 `codex` 바이너리가 없고 버전 의존성이 커 **게이트로 삼지 않는다**. 운영자 수동 확인 항목으로 `progress.md` 에 남기며, 확인 방법은 t91 §4 의 격리 방식을 따른다 — `CODEX_HOME=<scratch>` 로 사용자 홈을 분리하고 `codex debug prompt-input`(모델 호출 0회)으로 노출 여부를 관측한다. (iter-3 까지 제외 사유를 "사용자 홈 상태에 의존하므로 기계 판정 불가"로 적었는데, t91 이 바로 그 변수를 격리로 제거하고 기계적으로 판정했으므로 **틀린 사유**였다. 제외 결정 자체는 유지하되 근거를 실제 이유로 바꾼다.) 확인할 때 **상대 디렉터리 링크 형태**로 노출되는지를 명시적으로 본다: M0 실측(`t91/README.md` §4)은 `.agents/skills/t91-link-src -> .claude/skills/t91-link-src` 가 노출된다는 사실만 기록했고 그 링크가 상대였는지 절대였는지, 파일 링크였는지 디렉터리 링크였는지는 **기록하지 않았다**. REQ-CSC-003 이 강제하는 것은 상대 디렉터리 링크이므로, "링크가 동작한다"는 관측이 "이 형태의 링크가 동작한다"를 함의하지 않는다 — spec §A.2 가 세운 규율(런타임 사실과 빌드타임 사실을 섞지 말라)을 같은 문서에 적용한 결과다.

## §D.5 종료 게이트

- 대상 패키지 테스트 통과: `go test ./internal/template/... ./internal/cli/update/...`
- `go vet ./internal/...` 및 `GOOS=windows go vet ./internal/...`
- 전량 스위트 판정은 CI (로컬 `go test ./...` 금지)

## §D.6 Definition of Done

- MUST AC 15개 전부 실행 출력과 함께 PASS
- SHOULD AC 1개는 PASS 또는 "해당 없음" 사유 명시
- 기존 7개 청소 뿌리에 대한 `os.Lstat` 전환 회귀 확인 결과 기록(spec §B.D6 의 폭발 반경)
- `progress.md` §E.2 에 §D.4 의 세 항목(커밋 SHA 대조 결과, Windows vet, Codex 수동 확인)과 착수 시점 스킬 개수 재측정값, 실행한 명령과 출력이 기록됨
- spec §D 의 범위 밖 항목이 하나도 실행되지 않았음이 diff 로 확인됨 (특히 `~/.codex/` 무변경)

## §D.7 미해결 / 전방 확인

- Claude Code 자체가 `.claude/skills/` 아래 심볼릭 링크를 따라가는지는 미실측이다. D1 방향에서는 이 사실이 필요 없지만, 훗날 방향을 뒤집자는 제안이 나오면 그때는 선결 조건이다.
- pre-clean 백업이 심볼릭 링크의 **대상을 따라가는지**는 더 이상 미해결이 아니다 — 따라가지 않는다(spec §A.7 실측). 그 자리에 남은 실제 위험은 반대 방향인 복사 모드이며, REQ-CSC-010 과 AC-CSC-012(2번 팔)가 규정한다.
- AC-CSC-010 이 요구하는 "미러 생성을 끄고 켤 수 있는 seam" 은 run-phase 설계에 제약을 건다. M1 에서 이 토글 가능성을 먼저 확보하지 않으면 AC-CSC-010 과 AC-CSC-013(2번 단언)이 함께 막힌다.
- **같은 성격의 제약이 하나 더 있다 — 출력 seam.** `internal/template` 에는 `io.Writer` 가 없고 `Deploy` 는 `error` 만 돌려준다(실측). 모드·경고를 반환 결과로 올리는 seam 을 M2 에서 만들지 않으면 **AC-CSC-006 · 011(3) · 013(3) 세 개가 함께 막힌다.** 그중 둘은 MUST 다.
- `os.Lstat` 전환의 폭발 반경(spec §B.D6)은 이 SPEC 의 AC 가 덮지 않는다. 기존 7개 청소 뿌리에서 dangling 링크가 이제 제거되므로, run-phase 는 그 뿌리들에 대한 회귀를 별도로 확인해야 한다 — 본 SPEC 은 그 확인을 요구할 뿐 판정 기준을 세우지 않는다.
