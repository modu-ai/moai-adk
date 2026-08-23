# SPEC 감사 보고서: SPEC-CODEX-SKILLS-CANONICAL-001

Iteration: 1/2 (Tier M 상한 2 — `.moai/reports/plan-audit/` 에 선행 감사 보고서 없음. spec HISTORY 의 iter-1/iter-2 는 작성자 자체 개정이며 감사 반복이 아니다)
Verdict: **FAIL**
Overall Score: **0.775** (Tier M PASS 임계 0.80 — `spec-workflow.md` § SPEC Complexity Tier)

감사 대상: `.moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/{spec.md,plan.md,acceptance.md}` (Tier M 입력 계약).
Reasoning context ignored per M1 Context Isolation — 리드 지시문에 담긴 전제·유도 문구는 판단 근거에서 제외했고, 아래 모든 수치는 이 worktree(`WT-skills-canonical`)에서 직접 재실행한 명령의 출력이다.

---

## Must-Pass 결과

- **[PASS] MP-1 REQ 번호 일관성** — `grep -o 'REQ-CSC-[0-9]*' spec.md | sort -u` → `REQ-CSC-001 … REQ-CSC-012` 연속 12개, 결번·중복 0, 3자리 zero-pad 일관. 정의 라인 수 `grep -c '^- \*\*REQ-CSC-'` = 12 로 일치.
- **[PASS] MP-2 GEARS 형식 준수 (요구사항 계층 판정)** — REQ-CSC-001~012 전부 GEARS 5패턴에 대응한다: ubiquitous(001·007·010), unwanted `shall not`(002), event-driven `When`(003·005·008·011·012), where-gate `Where`(004), state-driven `While`(009), 파생 조건 + `shall not`(006). 비형식 서술("should" / "가능하면")은 0건. `acceptance.md` 의 Given-When-Then 항목은 **검증 계층(AC-XXX)** 이므로 이 판정 대상이 아니며 Group 4 에서 별도 채점했다.
- **[PASS] MP-3 YAML frontmatter 유효성** — canonical 12필드 전부 존재·타입 적합(`id` 정규식 적합, `version: "0.2.0"` quoted semver, `status: draft` ∈ 8-value enum, `created`/`updated` ISO, `priority: P2`, `phase: "v3.2.0 target"` — 금지 lifecycle 토큰 아님, `module: internal/template`, `lifecycle: spec-anchored`, `tags` CSV 문자열). snake_case alias(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. 추가 필드 `tier: M` 은 허용 범위.
- **[N/A] MP-4 §22 언어 중립성** — 본 SPEC 은 moai-adk 자체의 Go 배포기·템플릿 트리를 다루며 16개 지원 언어의 도구 체인을 열거·차등하지 않는다. §E 가 템플릿 중립성 doctrine 을 인용할 뿐 언어별 도구명을 담지 않으므로 해당 없음.
- **[PASS] MP-5 D7 교차-SPEC 정합** — 본문에서 추출된 외부 SPEC 참조는 `SPEC-CODEX-PHASE2-001` 1건(그것도 `progress.md` 의 중복 검사 기록). `.moai/specs/SPEC-CODEX-PHASE2-001/spec.md` 존재, `status: completed` — retired/superseded/archived 아님. BLOCKING 없음.
- **[PASS] MP-6 D8 크로스 플랫폼 규율** — `grep -c 'syscall'` = spec 0 / plan 0 / acceptance 0. 자동 PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/` → rc=1(매치 0). 미해결 마커 없음. (`research.md` 는 Tier M 이라 부재 — 정상)

Must-pass 는 전부 통과다. **FAIL 은 must-pass 위반이 아니라 집계 점수(0.775 < 0.80)에서 나왔고**, 그 점수를 끌어내린 것은 아래 blocking 4건이다.

---

## 범주별 점수 (rubric 앵커)

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.80 | 0.75 | REQ 12개가 단일 해석을 갖는다. 감점: REQ-CSC-001 "읽을 수 있는 접근 경로를 제공"은 링크/복사 어느 쪽도 허용하는 의도적 추상이지만 "읽을 수 있는"의 판정 주체가 REQ 안에 없다(AC-002 가 뒤에서 고정). REQ-CSC-006·011 이 한 항목에 `shall` + `shall not` 두 modality 를 담아 원자성이 깨진다. |
| Completeness | 0.70 | 0.75 밴드 미달 | 절 구성·frontmatter·Out of Scope(H3 4개, 각 `-` 불릿 보유)는 완비. 그러나 **부하를 지는 항목 4개가 REQ·AC 어디에도 없다**: `.gitignore` 의 `.agents/` 취급(D5), pre-clean 백업과 복사 모드의 상호작용(D4), 미러 집합 ⊆ `moai*` 불변식(D3), 대상 경로에 사용자의 실 디렉터리가 이미 있는 경우(D6). plan §F M5 가 `.gitignore` 를 "확인한다"로만 두고 Priority Low 에 배치한 것이 대표적이다. |
| Testability | 0.65 | 0.50–0.75 | MUST AC 10개 중 3개가 기술된 형태로는 기계 판정이 서지 않는다(D1·D2·D7). 특히 AC-CSC-001 은 **자기가 막으려는 실패 형태에서 통과할 수 있다**. |
| Traceability | 0.95 | 1.0 근접 | §D 매트릭스 15행 + §D.3 역방향 매핑이 REQ 12개를 모두 덮고, 고아 AC 0. AC-CSC-015 만 REQ 가 아닌 §E 비기능에 걸리는데 이는 명시돼 있어 흠이 아니다. |

집계 = (0.80 + 0.70 + 0.65 + 0.95) / 4 = **0.775**.

---

## Defects Found

### D1 — AC-CSC-001 은 자신이 막으려는 실패에서 통과한다 (blocking)

`acceptance.md:L34-40` — "두 트리의 `.claude/skills/` 하위 **1단계 디렉터리 집합**을 각각 수집해 비교" — Severity: **critical** — Class: **blocking**

Go 의 `fs.DirEntry` 는 `Lstat` 기반이므로 **디렉터리를 가리키는 심볼릭 링크의 `IsDir()` 는 false** 다. 이 worktree 에서 직접 측정:

```
os.Stat(link).IsDir()=true   os.Lstat(link).IsDir()=false
```

즉 누군가 `internal/template/templates/.claude/skills/moai-new -> ...` 를 넣으면, 임베드 FS 에서는 사라지고(§A.2 그대로) **파일시스템 쪽 수집에서도 `IsDir()` 가 false 라 빠진다**. 두 집합은 여전히 같고 테스트는 **통과한다** — 스킬이 무음으로 소실된 바로 그 상태에서.

AC 본문의 "파일시스템 쪽 심볼릭 링크 수가 0이고" 절이 유일한 방어인데, 그 카운트의 **범위가 명시돼 있지 않다**. 더 나쁜 것은 REQ-CSC-002 의 범위(`internal/template/templates/` **전체**)가 AC-CSC-001 의 검증 범위(`.claude/skills/` 하위)보다 넓다는 점이다 — `.claude/agents/` 나 `.claude/rules/` 에 링크가 들어가면 REQ 위반인데 이 AC 는 보지 못한다.

Required fix: AC-CSC-001 을 두 개의 독립 단언으로 다시 쓴다. (1) `find internal/template/templates -type l` 에 상당하는 **트리 전체** 링크 카운트 == 0 (REQ-CSC-002 의 실제 범위), (2) 집합 비교는 `os.Stat` 기반(또는 `d.Type()&fs.ModeSymlink != 0` 를 별도로 세는) 수집으로 바꿔 링크 항목이 FS 집합에 **포함되게** 한 뒤 임베드 집합과 대조. 지금처럼 `d.IsDir()` 로만 모으면 등식이 깨지지 않는다.

### D2 — AC-CSC-010 / AC-CSC-013 이 참조하는 "변경 전 기준선"은 테스트 시점에 존재하지 않는다 (blocking)

`acceptance.md:L14`(매트릭스 "Go 테스트 (기준선 대조)") + `L96-101` + `L119-123` — Severity: **critical** — Class: **blocking**

AC-CSC-010 은 "**변경 전 커밋에서** `t.TempDir()` 에 배포해 얻은 목록"을 기준선으로 삼는다. Go 테스트는 이전 커밋을 체크아웃해 그 시점 코드로 배포할 수 없다. 이것은 실행 가능한 테스트가 아니라 **일회성 수동 절차**이며, 그런데도 §D 매트릭스는 "Go 테스트"로, §D.2 는 **MUST** 로 분류한다. 게다가 변경이 착지한 뒤에는 재실행 자체가 불가능해 회귀 가드로 남지 않는다.

파급이 더 큰 쪽은 AC-CSC-013 이다. 링크·복사 **양쪽 실패를 주입한 단위 테스트** 안에서 "`.claude/skills/` 아래 산출물이 **AC-CSC-010 기준선과 동일**"을 단언하라고 요구하는데, 그 기준선은 그 프로세스 안에 없다. 구현자는 이 절을 만족시킬 방법이 없어 조용히 약한 단언(예: "파일이 존재한다")으로 대체하게 되고, 그러면 AC-CSC-013 의 MUST 성격이 형해화된다.

Required fix: AC-CSC-010 을 (a) 기계 판정 가능한 **불변식 테스트**(예: `.agents/` 미러 로직을 비활성화한 seam 과 활성화한 seam 이 `.claude/skills/` 산출물의 (상대경로, SHA-256, 퍼미션) 목록을 동일하게 낸다 — 같은 프로세스 안에서 양쪽을 실행)와 (b) 커밋 SHA 를 명시한 **일회성 수동 대조 절차**로 분리하고, 매트릭스의 "Go 테스트" 라벨을 실제 형태에 맞춘다. AC-CSC-013 의 두 번째 단언은 (a) 의 불변식 형태를 참조하도록 다시 쓴다.

### D3 — 미러 집합과 청소 집합의 비대칭이 남는다 — §A.5 가 스스로 지목한 실패 형태 (blocking)

`spec.md:§A.5` + `§B.D5` + `REQ-CSC-006` / `REQ-CSC-008` — Severity: **major** — Class: **blocking**

§A.5 는 뿌리를 정확히 짚었다: "**청소 규칙과 배포 규칙이 같은 집합을 가리키지 않으면 조용히 어긋난다.**" 그런데 SPEC 이 만드는 두 집합은 정의가 서로 다르다.

- 미러 집합(REQ-CSC-006) = **이번 실행에서 실제로 배포된 스킬 전부** — 이름 제약 없음.
- 청소 집합(REQ-CSC-008) = `.agents/skills/moai*` — **`moai` 접두만**.

현재 이 둘이 우연히 같다. 측정: 템플릿 스킬 34개 중 `moai` 로 시작하지 않는 이름은 **0개**(`find … -exec basename {} \; | grep -cv '^moai'` → 0). 그러나 이 등식이 성립한다는 사실이 **SPEC 어디에도 요구사항으로 적혀 있지 않다**. `catalog.yaml` 에는 이미 `harness_generated` tier 슬롯이 있고(현재 `skills: []`), 여기 스킬이 하나라도 들어오는 순간 미러는 `.agents/skills/<비-moai>` 를 만들고 청소 글롭은 그것을 **영원히 지우지 못한다** — §A.4 가 `~/.codex/skills/` 오염으로 관측한 그 형태가 `.agents/` 에서 재현된다.

이 불변식은 코드에도 이미 암묵적으로 존재한다(`internal/template/skills_manifest.go:42` 가 `moaiSkillPrefix` 로 필터). 암묵을 명시로 올리지 않으면 나중에 조용히 깨진다.

Required fix: REQ 를 하나 추가한다 — "배포되는 스킬 이름은 모두 `moai` 접두를 가져야 한다(shall)", 그리고 이를 임베드 카탈로그에 대해 기계 검증하는 AC 를 붙인다. 또는 REQ-CSC-008 의 글롭을 미러 집합에서 파생하도록 다시 쓴다(단 §B.D5 의 접두 한정 계약은 유지). 어느 쪽이든 **두 집합의 관계가 단언되어야** 한다.

### D4 — 복사 폴백 모드에서 `moai update` 가 매 실행마다 전 스킬을 백업 트리에 복제한다 (blocking)

`internal/cli/update/deploy/deploy.go:371-399` 상호작용 + `spec.md:§D.7` (틀린 방향을 보고 있음) — Severity: **major** — Class: **blocking**

§D.7 은 "`backupThenRemove` 의 pre-clean 백업이 심볼릭 링크를 만나면 링크를 복사하는지 **대상을 복사하는지**" 를 M4 확인 사항으로 남겼다. 직접 측정한 결과 **그 방향은 문제가 아니다**:

```
visit .../link  isdir=false regular=false
regular files walked from symlink root: 0
after RemoveAll(link): link gone=true  canonical SKILL.md alive=true
```

`filepath.WalkDir` 는 루트를 `Lstat` 하므로 링크 루트에서 정규 파일을 0개 걷고, `os.RemoveAll` 은 링크만 지운다. 정본은 안전하다.

**실제 위험은 반대 방향이고 SPEC 이 전혀 다루지 않는다.** 복사 폴백(Windows)에서 `.agents/skills/moai-*` 는 **실 디렉터리**다. `backupThenRemove` 는 `templateManagedPaths(tmplFS, ".agents/skills/moai-*")` 를 계산하는데, REQ-CSC-002 가 템플릿 트리에 링크를 금지하고 `.agents/` 자체가 템플릿에 없으므로 이 집합은 **항상 공집합**이다(`deploy.go:410-413` — 템플릿이 그 뿌리 아래 아무것도 안 가지면 전부 unmanaged). 결과: `moai update` 를 돌릴 때마다 **34개 스킬의 모든 파일이** `.moai-backups/<timestamp>/pre-clean/.agents/skills/...` 로 복사된다. 매 업데이트마다.

Required fix: (a) §D.7 의 미해결 항목을 측정 결과로 대체하고(링크 방향은 안전), (b) 복사 모드에서의 pre-clean 백업 동작을 REQ 로 규정한다 — 미러는 정본의 파생물이므로 백업 대상이 아니어야 한다. (c) 이를 검증하는 AC 를 추가한다: 복사 모드 배포 후 `CleanMoaiManagedPaths` 를 실행했을 때 `.moai-backups/**/pre-clean/.agents/` 아래 파일 수가 0.

### D5 — `.gitignore` 의 `.agents/` 취급이 부하를 지는데 REQ·AC 가 없다 (blocking 아님 / major)

`plan.md:§F M5` ("`.gitignore` 에 `.agents/` 취급이 필요한지 확인한다", Priority **Low**, "기계 작업") — Severity: **major** — Class: **blocking**

측정: `internal/template/templates/.gitignore` 는 `.agents/` 를 무시하지 않는다(`.claude/` 관련 항목만 존재 — L174-178). 따라서 이 변경이 착지하면 모든 사용자 프로젝트가 커밋 대상 후보로 34개 항목을 새로 얻는다.

- 링크 모드: git 은 심볼릭 링크를 심볼릭 링크로 저장하지만, **심볼릭 링크를 지원하지 않는 Windows 체크아웃에서는 경로 문자열을 담은 일반 텍스트 파일로 실체화된다** — 커밋되는 순간 크로스 플랫폼 결함이 된다.
- 복사 모드: 스킬 34개 전체가 사용자 저장소에 **중복 커밋**된다.

"필요한지 확인한다"는 판단을 run-phase 로 미루는 형태이고, Priority Low·"기계 작업" 분류는 이 결과와 맞지 않는다. `.agents/` 는 생성물이지 소스가 아니라는 것을 plan 본문이 이미 알고 있으면서 요구사항으로 승격하지 않았다.

Required fix: "미러 산출물은 사용자 저장소의 버전 관리 대상이 아니어야 한다(shall not)" 취지의 REQ 를 추가하고, 템플릿 `.gitignore` 에 `.agents/` 항목이 존재하는지 검사하는 AC 를 붙인다. plan 의 우선순위도 Low 에서 올린다.

### D6 — `moai update` 는 `CleanMoaiManagedPaths` → `Deploy` 순서로 도는데, 멱등 AC 는 clean 없는 재배포만 본다 (major)

`acceptance.md:AC-CSC-011` + `internal/cli/update_template_sync.go:275-330` — Severity: **major** — Class: **blocking**

AC-CSC-011 은 리드 질문대로 **1회차 위에 2회차 `Deploy` 를 실제로 태운다** — 그 점은 확인했다. 그러나 이것이 검증하는 것은 실제 `moai update` 경로가 아니다. 실제 경로는 clean(미러 제거) → Deploy(미러 재생성)이라 링크가 이미 존재하는 상태를 만나지 않는다. AC-CSC-011 이 잡는 것은 `moai init` 재실행 같은 clean 없는 경로다 — 유용하지만, REQ-CSC-012 의 "이미 올바른 정본을 가리키고 있을 때"는 **잘못된 것을 가리키고 있을 때**와 **사용자가 손으로 만든 실 디렉터리가 이미 있을 때**를 규정하지 않는다.

후자가 위험하다. 링크 모드 구현이 `os.Symlink` 의 `EEXIST` 를 만나면 자연스러운 대응은 "지우고 다시 만든다"인데, 대상이 사용자의 실 디렉터리라면 그 순간 **데이터 손실**이다. 이름이 `moai*` 로 한정된다는 사실이 완화하지만 제거하지는 않는다(사용자가 `moai-` 이름을 쓸 수 있다). SPEC §D 는 "사용자가 스스로 만든 `.agents/skills/` 항목의 생성·관리"를 범위 밖으로 두면서 REQ-CSC-009(청소 비대상)만 보장하는데, **배포기가 덮어쓰는 경우**는 그 보장 밖이다.

Required fix: REQ-CSC-012 를 세 상태로 확장한다 — (i) 올바른 링크 → 무변경, (ii) 잘못된 링크 → 교체, (iii) 링크가 아닌 실 항목 → 정의된 동작(덮어쓰기라면 명시적으로, 아니면 건너뛰고 경고). (iii) 에 대한 AC 를 추가한다.

### D7 — REQ-CSC-010 / AC-CSC-012 는 기존 manifest seam 으로 링크 모드를 만족시킬 수 없다 (blocking)

`spec.md:REQ-CSC-010` + `acceptance.md:AC-CSC-012` + `internal/manifest/manifest.go:130-149`, `internal/manifest/hasher.go:16-34` — Severity: **major** — Class: **blocking**

`manifest.Manager.Track(path, provenance, hash)` 는 내부에서 `HashFile(absPath)` 를 호출하고 **실패하면 error 를 반환한다**. `HashFile` 은 `os.Open` + `io.Copy` 다. `.agents/skills/moai-x` 가 **디렉터리를 가리키는 심볼릭 링크**이면 `io.Copy` 가 EISDIR 로 실패 → `Track` 이 error → 호출부가 그대로 올리면 `Deploy` 가 실패한다. 이는 REQ-CSC-011(fail-open)과도 정면 충돌한다.

즉 REQ-CSC-010 은 현재 seam 에서 **링크 모드에 대해 구현 불가능**하다. 가능한 길은 (a) 링크가 가리키는 파일들을 개별 항목으로 기록 — 그러면 정본 항목과 중복되고 `DetectChanges()` 가 매번 두 번 센다, (b) manifest API 를 바꾼다 — M1 범위를 넘는다, (c) REQ 를 복사 모드에만 걸고 링크 모드는 다른 형태로 기록한다.

plan 은 이 결정을 M3 "Priority Medium" 으로 두고 "기록 형태가 달라야 하는지 여기서 정한다"고만 적었다. 실행 가능성 자체가 미확인이라는 사실이 드러나 있지 않다.

Required fix: REQ-CSC-010 을 seam 실측에 맞춰 다시 쓴다(어떤 경로를, 어떤 provenance 로, 링크 모드에서 무엇을 hash 하는지). 불가능하다고 판단되면 REQ 를 범위 밖으로 내리고 AC-CSC-012 를 함께 제거한다 — 지금처럼 남겨두면 run-phase 가 M3 에서 막힌다.

### D8 — AC-CSC-007 의 `DisplayPath` 리터럴이 Windows 에서 깨진다 (major)

`acceptance.md:AC-CSC-007` + `internal/cli/update/deploy/deploy.go:52-80` — Severity: **minor** — Class: **blocking**

기존 항목은 전부 `filepath.Join(defs.ClaudeDir, defs.SkillsSubdir, "moai*")` 로 만들어지므로 Windows 에서 `DisplayPath` 는 `.claude\skills\moai*` 다. AC-CSC-007 은 "`DisplayPath` 가 `.agents/skills/moai*`"라는 **슬래시 리터럴 동치**를 요구한다. §D.5 종료 게이트가 전량 판정을 CI(윈도우 매트릭스 포함)에 맡기므로, 이 AC 를 문자 그대로 구현하면 Windows CI 에서 실패한다.

Required fix: AC 문구를 `filepath.Join(".agents", "skills", "moai*")` 와의 동치 또는 `filepath.ToSlash` 정규화 후 비교로 바꾼다.

### D9 — §A.3 의 tier 분해가 틀렸다 (minor)

`spec.md:§A.3` — "스킬 카탈로그 항목 34개 중 core 21개 / non-core 13개(`optional-pack:*` **12** + `harness-generated` **1**)" — Severity: **minor** — Class: **optional**

실측:

```
21 core
 3 optional-pack:backend
 1 optional-pack:design
 5 optional-pack:devops
 4 optional-pack:frontend
```

`optional-pack:*` 는 **13**이고 `harness-generated` 스킬은 **0**이다. `catalog.yaml:261-262` 의 `harness_generated:` 블록은 `skills: []` 이며, 그 tier 를 가진 유일한 항목은 **에이전트** `builder-harness`(L263-268)다.

합계(21/13/34)와 §A.3 의 결론(슬림 21 / 전량 34)은 영향받지 않는다. 그러나 이 SPEC 의 §A 전체가 "남이 잘못 센 숫자를 정정한다"는 자세로 서 있고 AP-7 을 세워 계수 방법을 규율한다. 그 문서 안의 부분합이 틀린 것은 §A 의 권위를 깎는다.

Required fix: "`optional-pack:*` 13" 으로 정정하고 `harness-generated` 언급을 뺀다.

### D10 — Codex 의 **상대** 심볼릭 링크 해석은 미실측이다 (minor)

`spec.md:§A.2`/`REQ-CSC-003` ↔ `.moai/reports/t91/README.md` §4 — Severity: **minor** — Class: **optional**

t91 이 실측한 것은 `.agents/skills/t91-link-src -> .claude/skills/t91-link-src` 이며, 보고서는 그 링크가 **상대 경로였는지 절대 경로였는지, 파일 링크였는지 디렉터리 링크였는지 기록하지 않는다**. REQ-CSC-003 은 `../../.claude/skills/<name>` 이라는 **상대 디렉터리 링크**를 강제한다.

OS 레벨 해석이 동일하므로 실패 확률은 낮지만, 이 SPEC 의 §A.2 자체가 "런타임 사실과 빌드타임 사실을 섞지 말라"는 교훈 위에 서 있다. 같은 규율을 적용하면 "링크가 동작한다"는 관측이 "이 형태의 링크가 동작한다"를 함의하지 않는다.

Required fix: §D.7(미해결/전방 확인)에 한 줄로 남기거나, `progress.md` 의 운영자 수동 확인 항목(§D.4 가 이미 마련한 자리)에 "상대 링크 형태로 Codex `/skills` 노출 확인"을 명시한다.

### D11 — §B.D5 의 doctrine 인용이 기계적 보호 범위를 과장한다 (minor)

`spec.md:§B.D5` — "같은 계약이 `.claude/agents/harness/` · `.claude/commands/harness/` · `.moai/harness/` 에도 걸린다" — Severity: **minor** — Class: **optional**

`internal/cli/update/plan/plan.go:152-177` 의 `IsUserOwnedNamespace` 는 전부 **`.claude/` 경로 접두**로 판정한다(`.claude/skills/hns-`, `.claude/skills/harness-`, `.claude/skills/my-harness-`, `.claude/agents/harness`). `.agents/` 아래 항목은 이 술어의 시야에 **없다**. 따라서 `.agents/skills/hns-*` 의 생존은 doctrine 의 기계적 보호가 아니라 **글롭이 좁다는 사실 하나**에 전적으로 의존한다 — 백업·preserve·doctor 어느 계층도 그것을 보지 않는다.

REQ-CSC-009 가 필요한 보장을 직접 제공하므로 결론은 옳다. 다만 근거 서술이 실제보다 두꺼운 방어층을 암시한다.

Required fix: §B.D5 에 한 문장 — `.agents/` 아래에서 이 계약을 지키는 것은 글롭 한정뿐이며 `IsUserOwnedNamespace` 는 관여하지 않는다 — 를 덧붙인다. 그래야 나중에 글롭을 넓히자는 제안이 왔을 때 무엇이 유일한 방어인지 읽힌다.

### D12 — REQ-CSC-006 / REQ-CSC-011 의 modality 가 복합적이다 (minor)

`spec.md:REQ-CSC-006`, `REQ-CSC-011` — Severity: **minor** — Class: **optional**

각각 한 항목 안에 `shall` 과 `shall not` 을 함께 담는다. 두 절 모두 형식 GEARS 이므로 MP-2 위반은 아니지만, 원자성이 깨져 추적 시 "어느 절이 실패했는가"가 AC 매트릭스에서 구분되지 않는다. Tier M REQ 예산(16)에 여유 4가 있으므로 분리 여력은 있다.

Required fix: 선택 사항. 분리하려면 REQ-CSC-006 을 "집합 일치"와 "상수 파생 금지"로, REQ-CSC-011 을 "경고 후 계속"과 "배포 취소 금지"로 나눈다.

### D13 — AC-CSC-003 은 파생 방식과 무관하게 참이 될 수 있다 (minor)

`acceptance.md:AC-CSC-003` — Severity: **minor** — Class: **optional**

"`.agents/skills/` 항목 집합이 `.claude/skills/` 항목 집합과 정확히 같아야" — plan §F M1 후보 (b)("배포 완료 후 대상 FS 를 다시 읽어 파생")를 택하면 이 단언은 **정의상 항상 참**이라 슬림/전량을 구분하지 못한다. AC-CSC-014(합성 2-스킬 FS)가 상수 하드코딩은 잡으므로 REQ-CSC-006 이 무방비인 것은 아니다.

Required fix: 선택 사항. AC-CSC-003 에 "슬림 배포 항목 수 < 전량 배포 항목 수"라는 관계 단언을 덧붙이면 tier 필터가 실제로 관통했음을 판정할 수 있다.

---

## 내 측정이 반박하는 전제

리드가 별도로 요구한 항목이다. **SPEC 본문에서 반박된 것은 D9 하나뿐이고, 나머지 셋은 선행 실측 문서 쪽이다.**

| # | 출처 | 문서의 주장 | 내 측정 |
|---|---|---|---|
| 1 | `spec.md:§A.3` | non-core 13 = `optional-pack:*` **12** + `harness-generated` **1** | `optional-pack:*` **13**, `harness-generated` 스킬 **0**. `catalog.yaml` 의 `harness_generated.skills` 는 `[]` 이고 해당 tier 의 유일 항목은 에이전트 `builder-harness` |
| 2 | `m1-preflight-measurements.md:§D` | `ManagedCleanTargets` 관리 대상에 `.moai/config` 가 포함된 8개 | `ManagedCleanTargets` 는 **7개**를 반환하고 전부 `.claude/` 하위. `.moai/config` 는 그 목록이 아니라 `CleanMoaiManagedPaths` **함수 본문의 별도 블록**(`deploy.go:164-182`)에서 지워진다. **spec.md §A.4 는 이 점을 바르게 적었다** — SPEC 이 preflight 를 정정한 사례 |
| 3 | `m1-preflight-measurements.md:§D/§E`, `t91/README.md:§4` | `~/.codex/skills/` 에 구 moai 스킬 **46개** | `find ~/.codex/skills -mindepth 1 -maxdepth 1` → **45** 항목, 그중 `moai*` 디렉터리 **44**. 46 은 SPEC 자신이 AP-7 로 규율한 그 계수 형태(+1~+2)로 보인다. spec.md 는 "다수"라고만 적어 이 오류를 물려받지 않았다 |
| 4 | `spec.md:§D.7` | pre-clean 백업이 심볼릭 링크의 **대상을 따라가 정본을 중복 저장**할 수 있다(M4 확인 필요) | 따라가지 **않는다**. `filepath.WalkDir` 가 루트를 `Lstat` 하므로 링크 루트에서 정규 파일 0개, `RemoveAll` 은 링크만 제거, 정본 생존 — 직접 실행 확인. 실제 위험은 반대 방향(복사 모드에서 매 update 마다 전량 백업 복제)이며 SPEC 이 다루지 않는다 → D4 |

### 반박되지 않고 **확인된** 전제 (독립 재현)

- **`//go:embed` 심볼릭 링크 무음 소실 — 확인.** go1.26.4 darwin/arm64 에서 최소 재현을 새로 작성: `templates/` 안에 디렉터리 링크(`linkdir -> real`), 내부 파일 링크(`linkfile -> real/a.txt`), **임베드 루트 밖을 가리키는 링크**(`outlink -> ../other/b.txt`) 셋을 두고 `//go:embed all:templates` 와 `//go:embed templates` 두 형태를 모두 빌드. 결과는 양쪽 동일:
  ```
  templates / templates/real / templates/real/a.txt
  ```
  세 링크 전부 소실, **빌드 오류 0 · 경고 0**. `all:` 접두 유무와 무관하고, 링크 대상이 임베드 루트 안이든 밖이든 무관하다. §A.2 는 성립하며 설계를 구속한다는 결론도 성립한다.
- **템플릿 배포 대상 34 — 확인.** `find … -type d | wc -l` = 34, `find … -name SKILL.md | wc -l` = 34, 로컬 44, `hns-*` 10 (44 = 34 + 10). 템플릿 트리 심볼릭 링크 0. `.agents/` 는 로컬·템플릿 양쪽 부재.
- **배포 스킬 수의 tier 의존 — 확인.** `internal/cli/init.go:653-664` 가 `shouldDistributeAll(cmd)` 로 갈라 기본은 `NewSlimDeployerWithRenderer`(core 21) 를, `--all`/`MOAI_DISTRIBUTE_ALL` 에서만 전량(34) 을 쓴다. `moai update` 는 `template.EmbeddedTemplates()`(전량) 를 쓴다. 재설계 문서의 성공 지표 "Codex `/skills` 에 32개"는 **기본 슬림 init 에서 원리적으로 도달 불가**라는 §A.3 결론은 옳다. 다만 AC-CSC-003 이 개수가 아니라 **집합 동치**로 판정하므로 이 사실이 AC-CSC-003 이 단언할 수 있는 범위를 좁히지는 않는다 — 판정 형태 선택은 적절하다.
- **`ManagedCleanTargets` 에 `.agents/` 없음 — 확인.** 7개 전부 `.claude/` 하위(`deploy.go:52-80`).
- **런타임/빌드타임 혼동 없음 — 확인.** §A.2 는 t91 의 심볼릭 링크 관측을 "배포된 사용자 프로젝트의 런타임 사실"로 명시 분리하고 임베드 사실과 구분한다. 리드가 지목한 혼동은 SPEC 에서 발생하지 않았다.
- **AC-CSC-008 의 양팔 재작성 — 확인.** 제거 단언과 생존 단언이 한 테스트 안에 [HARD] 로 묶여 있고, 분리 금지 사유(제거만 단언하는 테스트는 전체 삭제 상태와 정상 상태를 구분하지 못함)가 본문에 적혀 있다. **AC-CSC-009 는 재진술이 아니라 확장이다** — 008 은 `hns-*` 하나를 대표 사례로 고정하고, 009 는 `harness-*` · `my-harness-*` · 계약에 이름조차 없는 임의 이름(`my-own`)까지 접두 밖 전체로 넓힌다. 두 legacy 세대 + 무명 항목은 008 에 없다.
- **글롭 과다 매치 없음 — 확인.** 제안된 형태는 `.agents/skills/moai*` 단 하나이며 `AP-6` 이 `.agents/` · `.agents/skills/` 전체 형태를 명시 금지한다. `moai*` 는 `hns-*` · `harness-*` · `my-harness-*` 어느 것과도 매치하지 않는다.
- **템플릿 중립성 — 위반 없음(현 시점).** 이 SPEC 이 템플릿 트리에 넣기로 한 파일은 아직 없다(§E·M5 가 조건부). 다만 D5 가 요구하는 `.gitignore` 변경은 템플릿 파일 변경이므로 착지 시 `MOAI_TEMPLATE_LEAK_STRICT=1` 게이트 대상이 된다 — AC-CSC-015 가 이를 덮는다.

---

## Recommendation

FAIL. Tier M 임계 0.80 에 0.775 로 미달했고, 원인은 blocking 7건 중 특히 판정 계층의 세 구멍이다. 아래 순서로 고치면 임계를 넘을 것으로 본다. **must-pass 는 전부 통과했으므로 SPEC 을 다시 쓸 필요는 없다 — 판정(acceptance.md) 보강이 작업량의 대부분이다.**

1. **D1 — AC-CSC-001 재작성** (최우선). 지금 형태는 §A.2 라는 이 SPEC 의 유일한 설계 구속 사실을 지키지 못한다. 링크 카운트를 REQ-CSC-002 의 실제 범위(템플릿 트리 전체)로 올리고, 집합 수집을 `Lstat` 기반(`d.IsDir()`)에서 링크를 **보이게** 하는 형태로 바꾼다.
2. **D2 — AC-CSC-010 / AC-CSC-013 을 실행 가능한 형태로 분리**. "Go 테스트" 라벨과 MUST 등급을 실제 형태에 맞춘다. AC-CSC-013 의 기준선 참조를 같은 프로세스에서 얻을 수 있는 불변식으로 교체한다.
3. **D7 — REQ-CSC-010 실행 가능성 확정**. `manifest.Track` 이 링크에 대해 error 를 낸다는 사실을 §A 에 실측으로 기록하고, REQ 를 그 사실 위에 다시 세우거나 범위 밖으로 내린다. 지금은 M3 에서 막힌다.
4. **D3 — 미러 집합 ⊆ `moai*` 불변식을 REQ 로 승격**. §A.5 가 스스로 지목한 실패 형태를 SPEC 이 다시 만들고 있다. REQ 예산 여유 4 안에서 처리된다.
5. **D4 — 복사 모드 pre-clean 백업 규정 + AC**. §D.7 의 미해결 항목을 측정 결과로 대체하고(링크 방향 안전), 반대 방향에 REQ·AC 를 붙인다.
6. **D5 — `.gitignore` 를 REQ 로 승격**하고 plan §F M5 의 우선순위를 올린다.
7. **D6 — REQ-CSC-012 를 3-상태로 확장**(올바른 링크 / 잘못된 링크 / 실 항목), (iii) 에 AC 추가.
8. **D8 — AC-CSC-007 을 경로 구분자 중립 비교로**.
9. **D9 — §A.3 부분합 정정**(13 + 0).
10. D10~D13 은 optional. 오케스트레이터 재량이며, 이것들만으로 재감사를 돌릴 이유는 없다.

재감사(iteration 2)는 위 D1~D9 열거 델타에 한정한 회귀 확인으로 족하다. Tier M 상한이 2 이므로 iteration 2 에서 FAIL 이면 사용자 개입(범위 축소 / PASS-with-debt / 상한 연장)을 선택해야 한다.
