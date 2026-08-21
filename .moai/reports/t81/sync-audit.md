# Sync-phase 감사 — SPEC-CODEX-SKILLS-CANONICAL-001 (카드 t81)

**감사 시점 고정**: `d0feb5b40`, `git status --short` 무출력(clean), 브랜치 `WT-skills-canonical`.
감사 시작·종료 시점 모두 재확인했고 감사 중 트리는 움직이지 않았다.

```
$ git log --oneline -1
d0feb5b40 docs(SPEC-CODEX-SKILLS-CANONICAL-001): record run-phase evidence and open the SPEC
$ git status --short
(무출력)
```

---

## Overall Verdict: **PASS** — 0.90 (Tier M 임계 0.80)

| 차원 | 점수 | 판정 | 근거 (기계 검증 출력) |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | `go test -count=1 ./internal/template/...` → `ok … 21.059s`; 위증 검사 7종 전부 붉어짐(아래 §2); 구현자 주장 3건 전부 독립 재현 |
| Security (25%) | 92/100 | PASS | 파괴 분기 위증 확인(USER.md 소멸 관측); 링크 대상은 임베드 파생 이름만(외부 입력 0); `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/` → `ok … 21.627s`; `go.mod`/`go.sum` 무변경 |
| Craft (20%) | 85/100 | PASS | `go test -cover ./internal/template/` → `coverage: 86.1% of statements`; `golangci-lint run ./internal/template/...` → `0 issues.`; 감점 사유는 F3·F7 |
| Consistency (15%) | 92/100 | PASS | `go vet ./internal/...` exit 0, `GOOS=windows go vet ./internal/...` exit 0; 변경 8파일 전부 `internal/template/` 안 |

**Must-pass 방화벽**: Functionality 92 ≥ 80, Security 92 ≥ 80 — 둘 다 독립 통과.
가중 평균 90.6, 조화 평균 90.1. 어느 쪽으로 계산해도 0.90.

**blocking 결함 0건.** 아래 findings 는 전부 optional 이며, 그중 둘(F1·F4)은 SPEC 이 §D 에 이미 잔존 고지로 적어 둔 항목을 감사가 실측으로 확인한 것이다.

---

## 1. 독립 재도출한 AC vs 보고를 받아들인 AC

「재도출」은 판정이 지키는 대상을 직접 부수어 붉어짐을 관측했거나, 판정 밖에서 같은 사실을 다른 수단으로 측정했다는 뜻이다.

| AC | 감사 처리 | 방식 |
|---|---|---|
| AC-CSC-001 | **재도출** | 템플릿 트리에 디렉터리 심볼릭 링크 주입 → 양팔 동시 red. 이어 2번 팔 수집을 `d.IsDir()` 로 되돌리자 **2번 팔이 침묵**(1번 팔만 보고) — acceptance 의 [HARD] 가 경계한 사각이 실재함을 감사가 독립 확인 |
| AC-CSC-002 | 실행 + 부분 재도출 | 테스트 통과. 별도로 전량 배포 후 미러 34개·정본 34개 일치와 `moai` 스킬 포함을 프로브로 직접 측정 |
| AC-CSC-003 | **재도출** | 미러 집합을 `EmbeddedMoaiSkillNames()` 상수로 바꾸자 3단언 전부 red, 특히 `slim mirror count 33 must be < full mirror count 33` 이 독립적으로 발화 |
| AC-CSC-004 | **재도출** | 프로브로 `readlink .agents/skills/moai` → `../../.claude/skills/moai` 직접 관측(상대·정본 지시) |
| AC-CSC-005 | **재도출** | 결과 seam 을 끊자 `mirror mode for "moai-alpha" = "", want "copy"` red |
| AC-CSC-006 | **재도출** | 같은 seam 절단으로 `CopyFallbackUsed() = false … the fallback is silent` red |
| AC-CSC-010 | 실행 + **다른 축 재도출** | 프로세스 내 on/off 불변식은 테스트 통과. 감사는 그와 **별개로** §D.4 커밋 대조를 자체 재현(§3-①) — AC 가 덮지 못하는 축을 감사가 직접 덮었다 |
| AC-CSC-011 | **재도출** | 비-심볼릭 선점 분기를 `os.RemoveAll` 로 바꾸자 `USER.md was destroyed by the re-deploy` red |
| AC-CSC-012 | **재도출** | 미러 파일 경로를 `manifest.Track` 에 전달하자 `manifest tracks mirror path ".agents/skills/moai-beta/SKILL.md"` red. 근거 사실도 직접 측정: `manifest.Track(dirsymlink) err=manifest track hash: hash file: read …: is a directory` |
| AC-CSC-013 | **재도출** | seam 절단으로 `mirror mode = "", want "failed"` red |
| AC-CSC-014 | **재도출** | 상수 치환으로 `mirror set = [moai-domain-backend … ], want exactly [moai-alpha moai-beta]` red |
| AC-CSC-015 | **부분 재도출** | `.gitignore` 항목 삭제 → 1번·3번 팔 red. 단 3번 팔은 **독립적으로 붉어질 수 없다** — F3 참조 |
| AC-CSC-016 | 실행 + 사실 재측정 | 템플릿 스킬 34개, 전부 `moai` 접두, 이름이 정확히 `moai` 인 스킬 존재, `moaiSkillPrefix = "moai-"` 이므로 헬퍼는 33개만 반환 — AC 의 [HARD] 근거를 소스에서 직접 확인 |

**보고를 그대로 받아들인 것은 없다.** 리드가 미검증으로 넘긴 커버리지·린트도 재실행했다(각각 86.1%, 0 issues).

---

## 2. 위증 검사(falsification) 실행 기록

전부 편집 → 단일 테스트 실행 → 즉시 복원 → `git status --short` 무출력 확인의 순서로 수행했다. 커밋·푸시 없음.

| 부순 것 | 붉어진 판정 | 관측 출력(발췌) |
|---|---|---|
| 템플릿 트리에 디렉터리 링크 주입 | AC-001 양팔 | `template source tree contains 1 symlink(s): [templates/.claude/skills/zz-probe-link]` + `skill set on disk (35) differs from the embedded set (34)` |
| 위 상태에서 2번 팔을 `d.IsDir()` 로 | (침묵 확인) | 1번 팔만 발화 — 2번 팔의 수집 방식이 load-bearing 임을 확인 |
| 미러 집합을 상수(`EmbeddedMoaiSkillNames`)로 | AC-014 · AC-003 | `want exactly [moai-alpha moai-beta]` / `slim mirror count 33 must be < full mirror count 33` + dangling 13건 |
| 선점 실 디렉터리를 지우고 재생성 | AC-011(3) | `USER.md was destroyed by the re-deploy: … no such file or directory` |
| `result.SkillMirrors` 미기록(seam 절단) | AC-004·005·006·013 | `no mirror entry recorded for "moai-alpha"` / `CopyFallbackUsed() = false … the fallback is silent` / `failure not reported in the result: []` |
| 미러 파일을 `manifest.Track` 에 전달 | AC-012 | `manifest tracks mirror path ".agents/skills/moai-alpha/SKILL.md"` |
| `.gitignore` 항목 삭제 | AC-015(1)(3) | `template .gitignore has no ".agents/skills/moai*" entry` |

되돌린 뒤 전량 재실행: `go test -count=1 ./internal/template/...` → `ok … 21.059s`.

---

## 3. 구현자가 제시한 세 측정 — 감사 자체 재현

### ① 커밋 기준선 대조 (재현 성공)

`git archive a338eab1b` 로 변경 전 트리를 스크래치에 펼치고, **양쪽 트리에 동일한 덤프 테스트를 넣어** 전량 임베드 FS 를 `t.TempDir()` 에 배포한 뒤 `.claude/skills/` 전체를 (상대경로, SHA-256, 퍼미션) 으로 덤프해 `diff` 했다.

```
baseline(a338eab1b)  ZZDUMP entries=262
current(d0feb5b40)   ZZDUMP entries=262
$ diff base.txt cur.txt ; echo diff_exit=$?
diff_exit=0
```

- **내용까지 비교했다** — 이름 목록이 아니라 파일별 SHA-256 + 퍼미션이다.
- **트리 전체다** — `.claude/skills/` 아래 262개 파일 전부(부분집합 아님). 비교 범위가 `.claude/skills/` 인 것은 REQ-CSC-007 의 범위와 일치한다.
- **미러는 켜져 있었다** — `NewDeployer(full)` 는 옵션 없는 기본 생성자이고 `skillMirrorDisabled` 의 zero value 가 「미러 ON」이다(`WithSkillMirror(false)` 를 주지 않았다). 즉 미러 기능이 동작하는 상태에서 262/262 일치했다.

구현자가 보고한 262 / diff exit 0 / 미러 활성은 **셋 다 사실**이다.

### ② Codex 노출 (재현 성공)

`codex --version` → `codex-cli 0.147.0` (설치 확인). `CODEX_HOME` 을 스크래치로 격리한 프로젝트에 정본 `.claude/skills/moai-probeskill/SKILL.md` 와 **상대 디렉터리 심볼릭 링크** `moai-probeskill -> ../../.claude/skills/moai-probeskill` 를 만들고 `codex debug prompt-input`(모델 호출 0회) 실행:

```
lrwxr-xr-x  moai-probeskill -> ../../.claude/skills/moai-probeskill
…
- moai-probeskill: audit probe skill for codex exposure verification
  (file: <scratch>/cxprobe/proj/.agents/skills/moai-probeskill/SKILL.md)
```

acceptance §D.4 가 지적한 공백(M0 관측이 링크의 **형태**를 기록하지 않음)은 이것으로 닫힌다 — REQ-CSC-003 이 강제하는 **상대·디렉터리** 링크가 실제로 따라가진다. 부수 관측 하나: `CODEX_HOME` 은 codex 홈만 격리하고 `$HOME/.agents/skills/` 는 여전히 스캔된다(출력에 사용자 홈 스킬 다수 동반). 이 SPEC 의 판정에는 영향이 없다.

### ③ `manifest.Track` EISDIR (재현 성공)

```
PROBE manifest.Track(dirsymlink) err=manifest track hash: hash file: read …/.agents/skills/x: is a directory
```

spec §A.6 의 주장 그대로다.

---

## 4. 자진 신고한 세 이탈에 대한 판단

### 이탈 1 — M2·M3 이 M1 과 같은 커밋 계열에 착지(사전 RED 없음)

**항목별 판단: 대부분 등가, 두 곳은 등가가 아니다.**

사전 RED 와 사후 위증 검사가 증명하는 것은 다르다. 사전 RED 는 「판정이 요구사항에서 나왔다」를, 위증 검사는 「판정이 지금 구현을 실제로 구속한다」를 증명한다. 후자가 놓치는 것은 *판정이 구현의 모양에 맞춰 재단됐을 가능성* 이다.

| 판정 | 등가? | 사유 |
|---|---|---|
| AC-011(3) 파괴 방지 | **등가** | 부수면 반드시 붉어지고(관측), 「지우고 다시 만든다」가 가장 자연스러운 오구현이라 판정이 구현 모양에 재단될 여지가 없다 |
| AC-006·013(3) 결과 seam | **등가** | seam 절단이 4개 판정을 동시에 붉힌다. seam 자체가 acceptance 의 [HARD] 로 **사전 규정**돼 있었고(§D.1 AC-006), 구현이 그 규정을 따랐지 판정이 구현을 따라간 것이 아니다 |
| AC-014·003 파생 불변식 | **등가** | 상수 치환으로 붉어지며, 합성 2-스킬 FS 라는 fixture 형태가 구현과 무관하다 |
| AC-012 manifest 부재 | **부분 등가** | 「`.agents/` 키 0개」는 현재 구현에서 **아무도 Track 을 부르지 않으므로 공허하게 참**이다. 미래 회귀(미러 파일 경로를 track)는 잡지만, 방향이 사후적이다. 사전 RED 였다면 「무엇을 기록하지 않는가」가 먼저 고정됐을 것이다. 다만 그 방향의 오구현을 실제로 잡는다는 것은 감사가 실측했으므로 손실은 제한적이다 |
| AC-015(3) 임베드 팔 | **등가 아님** | F3 참조 — 이 팔은 어떤 상태에서도 독립적으로 붉어지지 못한다. 사전 RED 였다면 「1번 팔이 초록인데 3번 팔만 붉은 상태」를 만들려다 그 불가능성이 드러났을 가능성이 높다 |

**결론**: 이탈 1 은 판정 능력에 실질적 손실을 만들지 않았다. 유일하게 순서가 실제로 중요했을 자리(AC-015 3번 팔)의 결과가 F3 이며, 그것은 low 다. 신고 자체가 정확하고 표까지 남겼다.

### 이탈 2 — `REQ-CSC-005` 호출부 표시 미배선

**scope 해석은 방어 가능하다. 다만 「REQ-CSC-005 충족」은 배포기 계층에 한해서만 정직하다.**

- REQ-CSC-005 의 `shall` 은 **배포기**에 걸려 있다: 「모드와 경고를 자신의 반환 결과에 담아 호출부로 올려야 하며(shall) … 배포기 내부에서 직접 출력해서는 안 된다(shall not)」. 「사용자 표시는 호출부가 수행한다」는 역할 분담의 **서술**이지 배포기에 건 의무가 아니다. 이 문장 구조 아래에서 배포기 계층의 의무는 완전히 충족됐고, 대응 AC(AC-CSC-006)도 [HARD] 로 「반환 결과를 대상으로 판정한다」고 못박아 두었다. 판정은 정직하다.
- 그러나 감사 실측: **`ResultDeployer` / `DeployWithResult` 의 프로덕션 소비자는 0건**이다.
  ```
  $ grep -rn "ResultDeployer\|DeployWithResult" --include='*.go' internal/ cmd/ | grep -v _test.go
  → internal/template/deployer.go 내부 정의·주석·자기호출만. internal/cli 소비자 없음.
  ```
  결과적으로 오늘 폴백 플랫폼에서 **경고는 누구에게도 도달하지 않는다.** REQ-CSC-001 이 예외 축을 정당화하며 쓴 문장 — 「그 두 경우에는 **경고가 접근 경로의 자리를 대신한다**」 — 은 시스템 수준에서 아직 성립하지 않는다. 자리를 대신할 경고가 아무도 읽지 않는 struct 안에 있다.
- **판정**: SPEC 범위 내 위반 아님(blocking 아님). 다만 「REQ-CSC-005 완료」로 상위 보고하면 시스템 수준에서 과장이 된다. F2 로 등록하고 승계 카드 지목을 권고한다.

### 이탈 3 — `draft → in-progress` 전환이 마감 커밋에 실림

확인함. `git log -S"status: in-progress"` → `d0feb5b40` 단 1건. M1(`9c94c6b7a`)부터 M6(`42c0c2167`)까지 다섯 커밋이 `status: draft` 인 채로 착지했다. 기능 영향 0, 라이프사이클 기록의 정확성 문제다. 강제 수단이 없었다는 점도 사실이다 — `status-transition-ownership.sh` 는 advisory(항상 exit 0). **자진 신고 정확, 심각도 info.**

---

## 5. Findings (구조화 결함 목록)

- **F1 [Medium][optional]** `internal/template/skill_mirror.go:190` — 폴백(복사) 모드 미러는 2회차 배포부터 **영구 고착**된다. 감사 실측:
  ```
  PROBE first  deploy mode=copy
  PROBE second deploy mode=skipped mirrorContent="alpha"   (정본은 이미 "alpha-v2")
  ```
  배포기는 *자기가 지난 실행에 만든 복사본* 과 *사용자 디렉터리* 를 구분하는 판별자가 없어 전자도 보존한다. **SPEC §D 가 「폴백 플랫폼 미러 고착」으로 이미 고지한 잔존 결과**이며 승계 카드 소관이다(축소가 만든 결과 — 전출 전에는 청소가 낡은 복사본을 지우고 다시 만들었다). 감사는 그 고지가 문서상 추정이 아니라 **실제 동작**임을 확인했다. Required fix (승계 카드): 미러 항목에 판별자(예: 미러 루트 안의 sentinel 파일 또는 `.moai/manifest` 밖의 별도 기록)를 두어 자기 산출물만 갱신 대상으로 삼을 것.
  Confidence: high. Windows/권한 없는 환경 전체가 영향 범위.

- **F2 [Medium][optional]** `internal/template/deployer.go:64` (`ResultDeployer`) — 반환 seam 은 만들어졌으나 **프로덕션 소비자가 0건**이라 모드·경고가 사용자에게 도달하지 않는다(§4 이탈 2). REQ-CSC-005 의 배포기 측 `shall` 은 충족, 시스템 수준 목적은 미달. Required fix: `internal/cli` 의 init/update 경로에서 `dep.(template.ResultDeployer)` 로 승격해 `Warnings()` 를 출력하는 후속 카드 등록. AC 가 없으므로 이 SPEC 의 판정에는 넣지 않는다.
  Confidence: high (grep 실측).

- **F3 [Low][optional]** `internal/template/gitignore_agents_mirror_test.go:80-93` — AC-CSC-015 3번 팔(임베드 FS 확인)은 **독립적으로 실패할 수 없다**. `//go:embed` 는 테스트 바이너리 컴파일 시점에 같은 on-disk 트리에서 임베드하므로, `go test` 문맥에서 「소스만 고치고 `make build` 를 빼먹은 상태」는 재현 불가능하다. 실측: 항목을 지우면 1번·3번 팔이 **항상 함께** 붉어지고, 1번이 초록인데 3번만 붉은 상태는 존재하지 않는다. AC 전체는 1·2번 팔로 정직하게 통과하지만, 3번 팔이 표방한 목적(빌드 누락 검출)은 달성되지 않는다. Required fix: 목적을 살리려면 설치된 바이너리(`~/go/bin/moai`) 를 대상으로 하는 별도 검사여야 하고, 그렇지 않다면 주석에서 「빌드 누락을 잡는다」는 주장을 내리는 편이 정직하다.
  Confidence: high.

- **F4 [Low][optional]** `internal/template/skill_mirror.go:196` — F1 상황에서 나가는 경고 문구 `a non-symlink entry already exists at … — left untouched` 가 **우리 산출물을 사용자 항목으로 오귀속**한다. 폴백 플랫폼의 2회차 이후 모든 실행에서 스킬 수만큼(현재 34건) 이 경고가 나간다. spec §D(라인 225)가 이 오귀속을 이미 적어 두었다. Required fix: F1 의 판별자가 들어오면 자연히 해소된다.
  Confidence: high.

- **F5 [Low][optional]** 저장소 루트 `.gitignore:127` 이 `.agents/` **전체**를 무시한다 — 이 SPEC 이 템플릿에 [HARD] 로 금지한 형태(§B.D7 좁은-범위 원칙)와 정반대다. AC-CSC-015 는 템플릿 파일만 보므로 잡히지 않는다. 배포 사용자에게는 영향 없고, 이 저장소가 훗날 `.agents/` 에 소스를 두면 조용히 추적에서 빠진다. Required fix: 후속 카드에서 루트 `.gitignore` 를 `.agents/skills/moai*` 로 좁힐지 판단(범위 밖이므로 이 카드에서 고치지 말 것).
  Confidence: high. Severity 낮음(개발 저장소 한정).

- **F6 [Info][optional]** `internal/template/skill_mirror.go:96` — `func (r *DeployResult) MirrorMode(skill string) (MirrorMode, bool)` 이 타입 `MirrorMode` 와 이름이 겹친다. 합법이나 가독성 손해. Required fix: 필요 시 `ModeFor` 등으로 개명(회귀 위험 0이지만 이 카드 범위 밖).

- **F7 [Info]** 이탈 3(status 전환 커밋 위치). 위 §4 참조.

**blocking 결함 없음.** F1·F4 는 SPEC 이 명시적으로 승계 카드에 넘긴 잔존이고, F2 는 대응 AC 가 없는 시스템 수준 갭이며, F3·F5·F6 은 품질 노트다.

---

## 6. 리드가 지목한 결함 클래스별 결과

| 지목된 클래스 | 결과 |
|---|---|
| 임베드 심볼릭 링크 소실 가드 | **실재 확인.** 링크 주입 → 양팔 red, 임베드에서 실제로 조용히 누락(35 vs 34) |
| `manifest.Track` EISDIR | **실재 확인.** `is a directory` 직접 관측 + 오구현 주입으로 판정 red |
| 출력 seam | **실재 확인.** 절단 시 4개 판정 동시 red |
| `moai-` vs `moai` 접두 | **실재 확인.** `moaiSkillPrefix = "moai-"`, 카탈로그에 이름이 정확히 `moai` 인 스킬 존재 → 헬퍼는 33/34. 미러 집합은 walk 파생이라 `moai` 포함(`readlink` 로 확인). 상수 치환 위증 검사에서 `moai` 가 미러에서 빠지는 것을 실제로 관측 |
| tier 의존 스킬 개수 | **실재 확인.** slim 21 / full 34. 상수 치환 시 `33 must be < 33` 으로 3번 단언 독립 발화 |
| 복사 폴백이 capability 기반인가 | **예.** `runtime.GOOS` 분기 없음(`grep` 실측 — `skill_mirror.go`/`deployer.go` 0건). `os.Symlink` 실패라는 **관측된 능력 부재**로만 선택된다 |
| 폴백 복사본이 읽히는가 | **예.** `TestSkillMirror_CopyFallbackIsReadable` 이 실 파일 여부(`Lstat().Mode().IsRegular()`)와 내용 동일성, 중첩 파일까지 본다 |
| 폴백이 stale/unreadable 을 내는 것을 테스트가 잡는가 | **아니오 — F1.** unreadable 은 잡지만 **stale 은 어떤 테스트도 잡지 않는다** |
| 선점 3-상태, 특히 (iii) 실 디렉터리 | **파괴 불가 확인.** 단언이 아니라 위증으로 증명 — 가드를 제거하니 USER.md 가 실제로 사라졌다 |
| 멱등성 | **확인.** 링크 모드 재배포는 대상 무변경(AC-011(1)), dangling 생성 없음, `os.Remove`/`os.Readlink` 는 링크를 따라가지 않으므로 정본이 링크를 통해 삭제되는 경로가 없다 |
| 템플릿 중립성 / Template-First | **통과.** `.gitignore` 추가분에 SPEC ID·날짜·SHA·macOS 경로 없음; `MOAI_TEMPLATE_LEAK_STRICT=1` 통과. 새 `.claude/`·`.moai/` 파일 없음. `catalog.yaml` 해시는 스킬 디렉터리만 덮으므로 `.gitignore` 변경에 재생성 불필요 — 즉 `make build` 산출물 커밋 누락 없음 |
| scope discipline (청소 경로 침범) | **위반 없음.** `git diff --stat a338eab1b..d0feb5b40 -- internal/cli/` 무출력. 변경 8파일 전부 `internal/template/`, 그 외 SPEC 문서 2건. `go.mod`/`go.sum` 무변경 |

---

## 7. 어떤 주장과도 어긋난 실측 (contradiction)

**구현 보고서·SPEC·앞선 감사의 주장 중 감사 실측과 어긋난 것은 없었다.** 확인한 항목:

- 13/13 AC PASS — 재현됨(그중 11건은 위증 검사까지).
- coverage 86.1% — 재현됨(`coverage: 86.1% of statements`).
- `golangci-lint` 0 issues — 재현됨(`0 issues.`).
- `go vet` / `GOOS=windows go vet` exit 0 — 재현됨.
- `MOAI_TEMPLATE_LEAK_STRICT=1` — 재현됨.
- 262 항목 / diff exit 0 / 미러 활성 — 재현됨(감사가 독립 덤프로).
- codex 노출(상대 디렉터리 링크) — 재현됨, `codex-cli 0.147.0`.
- 「폴백 플랫폼 미러 고착」 — SPEC §D 의 고지가 **문서상 추정이 아니라 실제 동작**임을 감사가 처음으로 실행 관측했다(구현 보고서는 이 항목을 실행 관측으로 남기지 않았다). 어긋남이 아니라 **보강**이다.

새로 추가된 사실 하나: **AC-CSC-015 3번 팔의 무력성(F3)** 은 SPEC·구현 보고서·앞선 plan 감사 어디에도 기록되지 않았다. 「스스로 막으려는 실패에서 통과하는 판정」이라는, 이 카드가 다섯 번 반복해 잡아 온 결함 클래스의 잔여 1건이다.

---

## 8. 감사 위생

- 읽기 전용 조사 + 본 보고서만 산출. 구현·테스트·SPEC 아티팩트 **무수정**.
- 위증 검사는 편집 → 단일 테스트 → 즉시 복원 → `git status --short` 무출력 확인의 반복으로 수행했고, 최종 상태는 `d0feb5b40` clean.
- `go test ./...` (전량) **미실행**. 패키지 단위로 `-timeout` 을 걸어 1건씩만 돌렸다.
- 백그라운드 부하 생성 0. 감사가 만든 스크래치(`t81base`, `cxprobe`, 덤프 2건)는 전부 삭제했다. 스크래치 디렉터리에 남은 `baseline/`·`codexproj/` 등은 run-phase 가 남긴 것이며 감사 산출물이 아니라 손대지 않았다.

---

**Auditor**: sync-auditor (독립 판정)
**감사 대상 커밋**: `d0feb5b40` (`a338eab1b` 대비 6커밋)
