# SPEC-VERSION-STAMP-PREDICATE-001 — 수락 기준

[HARD] **문서 수준 트리 핀: `9a3e2dabe`** (워크트리 `.claude/worktrees/t392`). 아래에서 자기
핀을 따로 달지 않은 모든 측정은 이 SHA에 귀속된다. 브랜치 이름으로 핀하지 않는다 — 브랜치는
움직이고 등가는 조용히 썩는다.

[HARD] **두 종류의 셀을 섞지 않는다.**

- **측정됨(plan)** — 이 세션에서 실제로 실행하고 출력을 관측했다. 명령과 출력이 함께 적혀 있다.
- **RED 대기(run)** — 검사가 아직 없으므로 관측하지 못했다. 실행할 명령과 기대 실패 문자열만
  **미리** 못박아 둔다(`plan.md` §D.2). 이 셀을 통과로 읽지 않는다.

---

## §D 수락 기준 (AC 15)

### AC-VSP-001 — 술어는 권위 토큰이며 출처가 하나다

**Given** 작업 트리의 `pkg/version/version.go` 가 `Version = "<token>"` 을 담고 있고
**When** 검사가 스윕 술어를 정할 때
**Then** 그 술어는 그 파일에서 읽은 `<token>` 하나이며, 일반 `vX.Y.Z` 형태가 아니다.

- RED 대기(run): 술어를 정규식 `vX.Y.Z` 로 바꾼 뮤턴트가 34가 아니라 595 파일을 스윕하고
  개수 단언이 운다.
- 측정됨(plan) — 술어의 출처가 유일함:

  ```
  $ grep -n 'Version = ' pkg/version/version.go
  8:	Version = "v3.1.3"
  exit 0
  ```

### AC-VSP-002 — 등록부는 정확 경로 28항목이고 글로브를 담지 않으며 스윕의 상위집합이다

**Given** 등록부 리터럴
**When** 항목을 읽을 때
**Then** 모든 항목이 정확 경로이고, `*` · `?` · `[` · 끝이 `/` 인 항목이 하나도 없으며,
항목 수가 28이고, 각 항목이 `stamp` 또는 `prose` 로 분류되어 있으며, 스윕 결과의 모든 경로가
등록부에 있다(상위집합). 등록부에만 있고 스윕에 없는 항목은 **위반이 아니다** — 인용이 낡은
`prose` 는 옳은 항목이다.

- 양방향(run): 등록부에만 있는 `prose` 항목 하나를 합성 입력에 두고 finding 이 나오지
  **않는지** 같은 실행에서 확인한다. 이 방향이 무너지면 §3의 4행이 결함으로 판정된다.

- 판정 명령(run): `grep -nE '"[^"]*[*?\[]' <검사 파일의 등록부 블록>` 이 0건.
- 측정됨(plan) — 28의 유도:

  ```
  $ git grep -lF v3.1.3 -- . ':!.moai/reports' ':!.moai/specs' ':!.moai/release-notes' \
      ':!CHANGELOG.md' ':!*_test.go' ':!docs-site/content/*/changelog*' | wc -l
  34
  exit 0
  $ git grep -lF v3.1.3 -- . <같은 거부 목록> | grep -c testdata
  6
  exit 0
  ```
  34 − 6 = 28.

### AC-VSP-003 — 미등록 + 현재 를 잡고 경로를 이름으로 부른다 [완결성]

**Given** 권위 토큰을 담았으나 등록부에 없는 파일이 트리에 있고
**When** 검사가 실행될 때
**Then** 검사는 실패하고 그 경로를 실패 출력에 이름으로 적는다.

- **RED 대기(run) — [HARD] pin 이전 트리에서만 관측된다(`plan.md` §D).**
  명령: `go test ./internal/cli/ -run TestVersionStampRegistry -count=1` (M3 커밋, M4 이전).
  기대 실패: `unregistered file carries the authoritative token: internal/cli/testdata/…` ×6.
- 측정됨(plan) — RED의 전제(golden 6개가 지금 권위 토큰을 담고 있음):

  ```
  $ grep -n 'v3\.1\.3' internal/cli/testdata/status-nocolor.golden internal/cli/testdata/doctor-nocolor.golden
  internal/cli/testdata/status-nocolor.golden:6:- **ADK**: moai-adk v3.1.3
  internal/cli/testdata/doctor-nocolor.golden:17:│    ok      MoAI Version           moai-adk v3.1.3   …
  exit 0
  ```
- green 경로: M4의 pin + 재생성이 여섯을 술어에서 빼면 스윕이 28이 되고 미등록 0.
- 뮤턴트 확인: finding 을 인쇄하고 exit 0 으로 끝나는 검사는 이 AC를 만족하지 못한다 —
  증거는 인쇄된 목록이 아니라 **비영 종료**다.

### AC-VSP-004 — 등록 + 낡음 을 잡고, 서술은 판정하지 않는다 [신선도]

**Given** `stamp` 로 분류된 등록부 항목이 권위 토큰을 담지 않고
**When** 검사가 실행될 때
**Then** 검사는 실패하고 그 경로를 이름으로 적는다. 같은 상태의 `prose` 항목은 실패를
일으키지 않는다.

- **RED 대기(run) — 합성 입력.** 순수 판정 코어에 (`stamp` 항목 하나의 내용 = `v0.0.0`)
  을 먹인다. 기대 실패:
  `registered stamp does not carry the authoritative token: <path>`.
  같은 표에서 `prose` 항목을 같은 상태로 두면 finding 이 나오지 않아야 한다(양방향).
- **실트리 RED이 없는 이유를 적는다.** 오늘 등록된 스탬프 7개는 전부 권위 토큰을 담고 있다
  (34 목록에 7개 모두 포함, AC-VSP-002 측정). 따라서 실트리에서는 이 단언이 처음부터
  초록이며, 합성 입력이 유일하게 가능한 RED이다.
- 모델링 대상 사고: `eba919e44`(v3.1.3 범프가 `hugo.toml` 을 빠뜨려 그 트리에서
  `version.go` = v3.1.3, `hugo.toml` = v3.1.2). 합성 픽스처는 이 모양을 흉내 내되 VCS를
  조회하지 않는다(REQ-VSP-010).

### AC-VSP-005 — 빈 스윕은 실패다, 그리고 상수는 등록부의 성질이다 [비공허성]

**Given** 등록부 항목 수 · `stamp` 분류 수 · 스윕 결과
**When** 검사가 실행될 때
**Then** 항목 수가 28이 아니거나 `stamp` 수가 7이 아니면 실패하고, 스윕 결과가 `stamp` 로
분류된 경로 하나라도 담고 있지 않으면 실패하며 **빠진 경로 전부를 이름으로** 적는다.
검사는 스윕 결과의 **개수**에 대한 기대값을 들고 있지 않다.

- **RED 대기(run) — 합성 입력 셋.**
  - 스윕 결과에 빈 슬라이스 → `registered stamp missing from sweep: <path>` ×7.
  - 등록부에서 항목 하나 제거 → `registry entries=27 expected=28`.
  - `stamp` 하나를 `prose` 로 바꾼 입력 → `stamp entries=6 expected=7`.
- [HARD] 이 RED은 **M3의 빨강과 구별된다.** M3에서 우는 것은 완결성 단언(미등록 6)이며,
  개수 단언은 M3에서 울지 않는다 — 상수가 등록부의 성질이고 등록부는 M1 이후 28로 고정이다.
  005의 RED은 실트리에 존재하지 않는 상태들이므로 합성 입력이 유일한 경로다.
- [HARD] **스윕 개수 기대값을 되살리지 않는다.** 그것은 범프 직후 28 → 7 로 움직이며,
  움직이는 원인이 §3 4행의 **정상** 상태다. 뮤턴트 확인: 「스윕 개수 = 28」을 다시 넣은
  뮤턴트는 범프 커밋 하나(`eba919e44` 모양)를 흉내 낸 합성 입력에서 실패해야 한다.
- 자기 참조 금지: 두 상수 어느 쪽도 스윕이나 등록부 파싱에서 유도하지 않는다
  (`plan.md` §D.3).
- 측정됨(plan) — 상수 둘이 범프에 불변인 근거:

  ```
  $ git show --numstat --format='%h %s' eba919e44
  eba919e44 chore: bump version to v3.1.3
  2	2	.moai/config/sections/system.yaml
  1	1	README.ja.md
  1	1	README.ko.md
  1	1	README.md
  1	1	README.zh.md
  1	1	pkg/version/version.go
  exit 0
  ```
  범프가 건드린 6파일은 **전부 스탬프**이고 서술은 하나도 없다. 파일이 늘거나 줄지 않으므로
  등록부 28도 `stamp` 7도 움직이지 않는다.

### AC-VSP-006 — 이름이 아니라 내용으로 판정한다

**Given** 파일 이름에 버전 토큰이 있고 내용에는 없는 경로가 있고
**When** 스윕이 그 파일을 판정할 때
**Then** 그 파일은 스윕 결과에 포함되지 않으며 어떤 개수에도 기여하지 않는다.

- **RED 대기(run) — 합성 입력 + 뮤턴트.** 판정을 경로 문자열로 바꾼 뮤턴트가
  `RELEASE-NOTES-v2.17.0.md` 류를 결과에 넣으면 이 AC가 실패해야 한다.
- 측정됨(plan) — 함정의 크기와, 이 술어에서 아직 물지 않는다는 사실:

  ```
  $ git grep -lE 'v[0-9]+\.[0-9]+\.[0-9]+' -- . <거부 목록> | grep -cE 'v[0-9]+\.[0-9]+\.[0-9]+'
  8
  exit 0
  $ git grep -lF v3.1.3 -- .moai/release/ | wc -l
  0
  exit 0
  ```
  이름-토큰 파일은 여덟, 그중 여섯이 `.moai/release/`(제외된 `.moai/release-notes/` 와 다른
  디렉터리)에 있으나, 권위 토큰 기준으로는 0을 기여한다.

### AC-VSP-007 — 거부 목록은 그룹별 사유와 실측 수로 열거된다

**Given** `.moai/docs/version-management.md` 의 거부 목록 절
**When** 사람이 읽을 때
**Then** 그룹마다 자기 사유와 감추는 파일 수가 적혀 있고, 그룹을 뭉뚱그린 단일 면제 조항이
없다.

- 판정 명령(run): 문서 표의 행 수 = 검사의 리터럴 제외 그룹 수 = **6**. `.git` 행은 없다 —
  모집단이 `git ls-files` 이므로 `.git/` 은 후보에 들어오지 않는다(iter-2, `spec.md` §2.3).
- [HARD] 표의 수치는 트리 SHA에 못박혀 있어야 한다. 판정 명령: 표 근처에서 `9a3e2dabe`
  (또는 run-phase 병합 트리의 SHA)를 grep 해 1건 이상.
- 측정됨(plan) — 귀속이 맞는지:

  ```
  reports 57 · specs 58 · release-notes 1 · CHANGELOG.md 1 · *_test.go 4 · changelog-pages 0
  합 121;  전체 트리 155;  155 - 121 = 34   ← 거부 목록 적용 값과 일치
  ```
  각 값은 `git grep -lF v3.1.3 -- <그룹> | wc -l` 을 그룹마다 따로 실행해 얻었다(exit 0).

### AC-VSP-008 — golden 은 pin으로 술어에서 빠진다 (등록이 아니다)

**Given** `status_golden_test.go` 와 `doctor_golden_test.go`
**When** 테스트가 실행될 때
**Then** 두 파일 모두 **빌드 시점 값과 무관한 고정 버전 토큰**으로 렌더하고 테스트가 끝날 때
원래 값을 되돌리며, 재생성된 6개 픽스처에 버전 토큰이 남아 있지 않고, 등록부에 그 여섯
경로가 없다. (구현 식별자는 요구가 아니라 `plan.md` M4가 소유한다 — 선례
`internal/cli/version_test.go:180-186`.)

- 측정됨(plan) — pin 부재(RED 상태):

  ```
  $ grep -c 'version\.Version' internal/cli/status_golden_test.go internal/cli/doctor_golden_test.go
  internal/cli/status_golden_test.go:0
  internal/cli/doctor_golden_test.go:0
  exit 1        # grep: 매치 0
  ```
  대조 선례는 `internal/cli/version_test.go:180-186`(`version.Version = "v0.0.0-test"` + defer).
- green 경로(run): M4 이후
  `grep -l 'v3\.1\.3' internal/cli/testdata/*.golden` 이 0건.
- [HARD] 여섯 경로가 등록부에 **없어야** 한다 — 등록으로 해결하면 이 AC는 실패다.

### AC-VSP-009 — 등록부의 stamp 집합과 문서 목록이 같다

**Given** 등록부의 `stamp` 항목 집합과 `parseVersionStampEntries` 가 문서에서 뽑은 경로 집합
**When** 검사가 두 집합을 대조할 때
**Then** 두 집합이 같아야 하며, 다르면 검사가 실패하고 **양쪽 차분 경로 전부**를 이름으로 적는다.

- **RED 대기(run) — 합성 입력.** 한쪽에서 경로 하나를 뺀다. 기대 실패:
  `stamp set differs from documentation list: <path>`.
- 측정됨(plan) — 오늘 두 집합의 크기가 같음:

  ```
  문서 「Version Stamps:」 항목  7   (t388 상수 expectedVersionStampEntries = 7)
  등록부 stamp 분류            7   (spec.md §2.2 분류표)
  ```
- 이 AC가 없으면 스탬프 목록이 두 곳에 생기고 조용히 갈라진다.

### AC-VSP-010 — 검사는 체크아웃된 색인과 작업 트리만 읽는다

**Given** 검사 파일 전체
**When** 코드를 훑을 때
**Then** (a) 판정 코어에 `exec.` 호출이 하나도 없고, (b) 파일 전체에서 `git` argv 리터럴이
**정확히 하나**이며 그것이 **정확히** `[]string{"git", "ls-files"}` 다.

- 판정 명령(run) — 두 방향:
  ```
  # (a) 코어에 외부 프로세스 없음 — 0건이어야 한다
  grep -nE '\bexec\.|os/exec' <검사 파일의 순수 코어 블록>
  # (b) 허용 목록 — git 리터럴이 하나뿐이고
  grep -c '"git"' internal/cli/version_stamp_registry_test.go            # → 1
  # (b) 그 하나가 정확히 이 argv 다
  grep -cF '[]string{"git", "ls-files"}' internal/cli/version_stamp_registry_test.go   # → 1
  ```
- [HARD] **(b)는 허용 목록이지 거부 목록이 아니다.** iter-2판은 금지 서브커맨드 열 개를
  열거했고, 감사가 `ls-tree` · `show-ref` · `for-each-ref` · `describe` · `worktree` ·
  `status` 여섯이 빠졌음을 지적했다 — `git ls-tree -r HEAD --name-only` 를 쓴 드라이버는
  이력을 읽으면서 그 grep을 **깨끗이 통과**한다. 거부 목록은 **빠뜨린 것에서 조용히**
  깨지고, 허용 목록은 열거하지 않은 것에서만 깨진다. REQ-VSP-010이 허용 목록으로 서술돼
  있으므로 판정도 같은 형태여야 한다. 열거할 금지어가 없으므로 빠뜨릴 것도 없다.
  이 판정이 성립하려면 argv가 이름 붙인 리터럴 하나여야 한다 — `plan.md` M3의 [HARD]가
  그 형태를 요구하고 AP-12가 흩는 것을 금지한다.
- **뮤턴트 확인(양방향) — 넷.**
  1. `[]string{"git", "ls-tree", "-r", "HEAD"}` 로 바꾼 뮤턴트 → 둘째 grep 0건, (b) 실패.
     (iter-2판 거부 목록은 이것을 **통과시켰다** — 이 뮤턴트가 N5 수리의 판별식이다.)
  2. `exec.Command("git", "show", …)` 를 한 줄 더한 뮤턴트 → 첫째 grep 2건, (b) 실패.
  3. `ls-files` 호출을 판정 코어로 옮긴 뮤턴트 → (a)가 잡는다.
  4. 손대지 않은 원본 → (a) 0건 · (b) 1건 · 1건으로 **통과해야 한다**. 통과 방향을 함께 보지
     않으면 grep이 파일을 못 찾는 경우와 구별되지 않는다.
- 근거: t388 §5의 [HARD]가 지키려던 두 성질 — 도달 불가 브랜치 비의존, 합성 입력 RED
  가능성 — 은 그대로다. 좁힌 것은 「git을 전혀 부르지 않는다」는 **수단**이지 그 두 성질이
  아니다(`spec.md` §2.0).

### AC-VSP-011 — 부분 보장 서술이 정확히 그만큼만 좁혀진다

**Given** `.moai/docs/version-management.md` 의 부분 보장 문단
**When** M5 이후 그것을 읽을 때
**Then** (a) 두 검사가 각각 무엇을 잡는지 적혀 있고, (b) 잡지 못하는 것들이 **열린 열거**로
적혀 있으며 그 열거가 적어도 여섯 — 미등록 + 낡음 · `prose` 오분류 · 제외 그룹 안 인라인 ·
날짜형 스탬프 · 렌더되는 버전 · 미추적 파일 — 을 이름으로 담고, (c) 「목록이 더는 썩지
않는다」에 해당하는 서술이 문서에도 이 SPEC에도 없으며, (d) 그 열거가 **닫힌 수**로 제시되지
않는다.

[HARD] **판정은 영문으로 한다. 대상 문서가 영문 전용이기 때문이다.** 측정(plan, 트리
`9a3e2dabe`): `grep -cP '[가-힣]' .moai/docs/version-management.md` → **0**(104줄). iter-2판은
`미등록` · `추적` 같은 한국어 낱말로 판정했고, 그 grep들은 문서가 담아야 할 어떤 문장에도
걸릴 수 없었다. 아울러 `_test.go` · `system.yaml.tmpl` 은 문서의 **다른 절에 이미 있으므로**
(현 L83 · L90) 판정어로 쓰면 M5가 아무것도 안 써도 통과한다. 그래서 여섯 항목을 §E 문안에만
존재하는 **구절**로 못박는다.

- (b) 판정 명령(run) — 여섯 전부 **1건 이상**:
  ```
  grep -cF 'aged-out token' .moai/docs/version-management.md                              # 1항
  grep -cF 'registered as `prose`' .moai/docs/version-management.md                       # 2항
  grep -cF 'inlined inside a file the exclusion set hides' .moai/docs/version-management.md # 3항
  grep -cF 'not a version token at all' .moai/docs/version-management.md                  # 4항
  grep -cF 'renders the version rather than carrying it' .moai/docs/version-management.md # 5항
  grep -cF 'the repository does not track' .moai/docs/version-management.md               # 6항
  ```
- **(d) 열린-집합 표지의 존재를 양성으로 단언한다.** 닫힌 수의 부재만 보면 파일이 없거나
  문단이 통째로 사라진 경우와 구별되지 않는다.
  ```
  grep -cF 'this list is not exhaustive' .moai/docs/version-management.md   # → 1건 이상
  ```
- **(d) 닫힌 수 서술의 부재** — 구절이 아니라 **구조**로 판정한다:
  ```
  grep -cEi '(does not|cannot|fails to) (catch|detect)[^.]*\b(two|three|four|five|six|seven|eight|nine|ten|[2-9])\b|\b(two|three|four|five|six|seven|eight|nine|ten|[2-9]) +(cases?|items?|things?|kinds?|directions?)[^.]*(remain|left|uncaught|not caught|catch)|(uncaught|not caught)[^.]*\b(two|three|four|five|six|seven|eight|nine|ten|[2-9])\b|(list|enumeration)[^.]*\bis exhaustive\b|잡지 못하는 것[이은][^.]*(둘|셋|넷|다섯|여섯)' .moai/docs/version-management.md
  ```
  → **0건**. 앞 넷은 영문 형태(부정+수사 / 수사+분류어+잔존 / 미포착+수사 / 「is exhaustive」),
  마지막은 한국어 구형을 남긴 것이다.
- [HARD] **뮤턴트 집합을 미리 선언한다 — 여덟.** iter-2판의 정규식은 **자신이 반박하려던
  단 하나의 문안**에 맞춰 조정돼 있어서, 감사가 먹인 일곱 중 여섯이 빠져나갔다(영문 둘 포함).
  같은 실수를 되풀이하지 않도록, 통과해야 할 뮤턴트를 하나가 아니라 집합으로 못박는다.
  아래 여덟을 **각각 따로** 먹여 전부 1건 이상이어야 한다:

  | # | 뮤턴트 | 형태 |
  |---|---|---|
  | 1 | `It does not catch three cases.` | 부정 + 수사 |
  | 2 | `Three cases remain uncaught.` | 수사 + 분류어 + 잔존 |
  | 3 | `The following three cases remain.` | 같은 형태, 앞머리 변형 |
  | 4 | `There are three things it still does not catch.` | 삽입절 |
  | 5 | `Still uncaught: three.` | 콜론형 |
  | 6 | `The list below is exhaustive.` | 수 없이 닫는 형태 |
  | 7 | `It fails to detect 3 kinds of site.` | 아라비아 숫자 |
  | 8 | `여전히 잡지 못하는 것이 셋이다.` | 한국어 구형(회귀 가드) |

- **양방향 — 실측(plan).** 위 정규식을 §E의 실제 교체 문안과 여덟 뮤턴트에 각각 돌렸다.
  트리 `9a3e2dabe`, 명령은 `grep -cEif <정규식파일> <대상>`:

  ```
  교체 문안(§E 영문, 8줄)  → 0
  뮤턴트 1..8              → 1 1 1 1 1 1 1 1   (여덟 전부 적중)
  현재 문서 전체            → rc=1 (0건, 오탐 없음)
  ```
  0건과 8건을 **함께** 봤으므로 「정규식이 아무것도 안 본다」와 구별된다.
- (b)/(d) 양성 판정의 **미리 못박은 RED** — 위 일곱 구절(여섯 + 열린-집합 표지)이 현재 문서에
  **하나도 없다**:

  ```
  $ grep -c -Ff <일곱 구절 목록> .moai/docs/version-management.md
  0
  exit 1
  ```
  M5 이전에는 (b)·(d) 양성이 전부 RED이고, M5가 §E 문안을 붙여 넣는 것으로만 GREEN이 된다.
- 교체 문안은 `plan.md` §E에 **영문으로** 미리 못박혀 있다. run-phase가 문안을 새로 짓지
  않는다.
- [HARD] **판정 구절은 한 줄 안에 있어야 한다** — `grep` 이 줄 단위이므로 구절 가운데 개행이
  들어가면 문안이 정확히 들어갔는데도 0건이 나온다(거짓 빨강). 실측(plan): 이 AC와
  AC-VSP-014가 쓰는 구절 열하나가 `plan.md` §E 안에서 각각 **1줄**에 있다
  (`grep -cF` 를 구절마다 따로 실행, 전부 1 이상).
- 측정됨(plan) — 좁힘 대상 문장이 현존함:

  ```
  $ grep -n 'does not detect' .moai/docs/version-management.md
  (「Files Requiring Version Sync」 절의 부분 보장 문단 안 1건)
  exit 0
  ```

### AC-VSP-012 — 스윕의 모집단은 추적 파일 집합이다

**Given** 모집단 드라이버
**When** 스윕이 후보 경로 목록을 얻을 때
**Then** 그 목록은 `git ls-files` 의 출력이며, `filepath.WalkDir` 도 `os.ReadDir` 재귀도
쓰지 않는다.

- 판정 명령(run):
  `grep -nE 'filepath\.Walk|WalkDir|ReadDir' internal/cli/version_stamp_registry_test.go`
  → 0건.
- **양방향(run) — 합성 모집단 둘.** 순수 코어에 모집단 둘을 먹인다: 하나는 문제의 경로를
  **담지 않고**(추적되지 않은 상태를 모형화), 다른 하나는 **담는다**(`git add` 이후를
  모형화). 두 모집단의 파일 내용은 같고 권위 토큰을 담는다. 앞은 finding 0건으로
  **통과해야** 하고, 뒤는 REQ-VSP-003이 그 경로를 이름으로 부르며 **실패해야** 한다. 한
  방향만 보면 「모집단이 좁다」와 「검사가 아무것도 안 본다」를 구별하지 못한다.
- [HARD] **실제 색인을 건드리지 않는다**(iter-3, N6 수리). iter-2판은 실파일을 만들어
  `git add` 하도록 적었다. `spec.md` §6의 [HARD] 순수 코어 분리가 **정확히 그것을 피하려고**
  존재하고(합성 입력으로 RED을 내기 위해), 이 저장소의 독트린은 공유 체크아웃의 색인을
  만지는 것을 따로 금지한다 — 정리 계약도 필요해진다. 합성 모집단 둘은 같은 두 방향을
  색인 변경 없이 낸다. `plan.md` AP-13이 되돌아가는 것을 금지한다.
- 측정됨(plan) — 두 모집단이 실제로 다르고, 파일시스템 walk가 판정 트리를 갈라놓음:

  ```
  $ grep -rlF --exclude-dir=.git v3.1.3 . | wc -l
  162
  exit 0
  $ git grep -lF v3.1.3 -- . | wc -l
  155
  exit 0
  $ comm -23 <파일시스템 목록> <추적 목록>
  .moai/reports/t392/baseline.md
  .moai/reports/t392/plan-audit.md
  .moai/reports/t392/red-observation.md
  .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/acceptance.md
  .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/plan.md
  .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/progress.md
  .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/spec.md
  exit 0
  $ find /Users/goos/MoAI/moai-adk-go/.claude/worktrees -maxdepth 1 -mindepth 1 -type d | wc -l
  183
  exit 0
  $ grep -rlF v3.1.3 /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t379 | wc -l
  144
  exit 0
  ```
  차이 7은 전부 이 카드 자신의 미추적 산출물이다. primary 체크아웃에서 walk는 형제 워크트리
  183개로 내려가고, **한 개만으로 144파일**이다.

### AC-VSP-013 — 등록부의 모든 항목이 실재하는 파일을 가리킨다 [유령 가드]

**Given** 등록부 28항목
**When** 검사가 각 항목을 확인할 때
**Then** 항목마다 작업 트리에 파일이 있어야 하며, 없으면 검사가 실패하고 **그 경로를
이름으로** 적는다.

- **RED 대기(run) — 합성 입력.** 실재하지 않는 경로 하나를 등록부에 넣는다. 기대 실패:
  `registry entry does not resolve to a file: <path>`.
- **왜 필요한가.** 이 가드가 없으면 삭제·이동된 서술 경로는 아무 단언도 어기지 않는다 —
  등록부는 스윕의 상위집합이므로(AC-VSP-002) 스윕에서 빠지는 것이 정상이기 때문이다.
  t388이 세운 검사에는 이 성질이 있고(`version_sync_list_test.go` 가 목록의 각 경로 실재를
  단언한다), 그것을 물려받지 않으면 **선행 카드에 대한 회귀**다.
- 양방향(run): 실재하는 경로만 담은 등록부에서 finding 이 0건이어야 한다.

### AC-VSP-014 — 문서가 등록부의 유지 계약을 적는다

**Given** `.moai/docs/version-management.md`
**When** 범프하는 사람이 읽을 때
**Then** (a) 등록부가 어느 파일에 있는지, (b) 누가 언제 고치는지, (c) 고치기 전까지 검사가
무엇을 말하는지, (d) **범프는 등록부를 건드리지 않는다**는 것이 적혀 있다. 서술 21경로는
문서에 옮겨 적히지 **않는다**.

- 판정 명령(run) — **영문**(대상 문서가 영문 전용, AC-VSP-011의 실측). (b)·(c)는 iter-2판에
  판정 명령이 아예 없었다 — 감사가 지적한 대로, 명령 없는 AC 항목은 판정되지 않는다:
  ```
  grep -cF 'version_stamp_registry_test.go' .moai/docs/version-management.md            # (a) → 1+
  grep -cF 'edits the registry in that same commit' .moai/docs/version-management.md    # (b) 누가·언제 → 1+
  grep -cF 'the check fails naming the path' .moai/docs/version-management.md           # (c) 그 사이 → 1+
  grep -cF 'A version bump does not touch the registry' .moai/docs/version-management.md # (d) → 1+
  grep -c 'docs-site/content' .moai/docs/version-management.md                          # 서술 경로 유출 → 0
  ```
  앞 넷은 1건 이상, 마지막은 0.
- **미리 못박은 RED(plan 실측).** 네 구절 모두 현재 문서에 **0건**이다(AC-VSP-011의
  `grep -c -Ff` 실측에 포함). M5 전에는 네 판정이 전부 빨강이다.
- **양방향(run).** 다섯째 명령은 문서에 서술 21경로가 새어 들어오면 0이 아니게 되어야 한다 —
  `docs-site/content/...` 한 줄을 넣은 뮤턴트로 확인한다. 0건만 보면 문서를 못 읽은 경우와
  구별되지 않는다.
- 측정됨(plan) — (d)가 참인 근거는 AC-VSP-005의 numstat 인용과 같다. 범프 커밋 둘
  (`eba919e44` 6파일 · `61921f1ba` 7파일)이 건드린 것은 전부 스탬프다.
- 이 AC가 없으면 §8 R-4가 **틀린 방향의 위험**을 미측정으로 적어 두는 초판 상태로 되돌아간다.

### AC-VSP-015 — 스윕이 자기가 받은 모집단을 실제로 훑었다 [도달 범위]

**Given** 드라이버가 넘긴 경로 집합과 등록부 28경로
**When** 검사가 실행될 때
**Then** (가) 등록부 경로 중 모집단에 없는 것이 있으면 실패하고 **그 경로 전부를 이름으로**
적으며, (나) 코어가 훑은 경로 수가 드라이버가 넘긴 수와 다르면 실패하고 **두 수를 함께** 적는다.
검사는 모집단의 **크기**에 대한 기대값을 들고 있지 않다.

- **RED 대기(run) — 합성 입력 둘.**
  - 등록부 경로 하나를 뺀 모집단 → `registry path missing from population: <path>`.
  - 코어가 마지막 하나를 건너뛰도록 만든 상태 → `judged=27 handed=28`.
- **양방향(run).** 온전한 모집단에서 (가)·(나) 둘 다 finding 0건이어야 한다. 빨강만 보면
  단언이 항상 우는 경우와 구별되지 않는다.
- **뮤턴트 — 이 AC가 없으면 통과하는 것들.** 각각 (가)가 잡아야 한다:
  1. 드라이버를 `internal/cli` 에서 실행해 `git ls-files` 가 그 하위로 좁혀진 경우.
  2. 제외 그룹 리터럴에 `docs-site/` 한 줄을 더한 경우 — 서술 20여 경로가 조용히 빠진다.
  3. 드라이버가 앞 N개만 넘기도록 잘린 경우.
  [HARD] 셋 다 iter-2 설계에서는 **네 단언이 전부 초록**이었다. 특히 범프 직후에는 정상
  스윕 결과가 스탬프 7과 정확히 같아지므로(`spec.md` §6.2의 실측), 그 순간 「전부 훑고 7」과
  「거의 못 훑고 7」이 구별되지 않았다.
- [HARD] **크기 리터럴을 넣지 않는다.** `git ls-files | wc -l`(측정 시점 10,048)은 범프에는
  불변이지만 **파일을 더하는 평범한 커밋마다** 움직인다. 리터럴로 두면 무관한 커밋에서 맨
  숫자로 실패하고 값을 깎는 값싼 수리 경로가 생긴다 — AC-VSP-005가 없앤 결함 부류의 재현이다
  (`spec.md` §6.2, `plan.md` AP-11). (가)의 좌변은 등록부(경로를 내는 SSOT)이고 (나)는
  양변이 그 실행에서 나온다.
- **닫지 않는 것.** 등록부 경로만 남기고 나머지를 버린 모집단은 (가)·(나)를 모두 통과한다.
  `spec.md` §8 R-9에 이름 붙여 열어 두며, 이 AC는 그것을 잡는다고 주장하지 않는다.
- 측정됨(plan) — (가)의 전제(오늘 등록부 28이 전부 모집단 안)는 **구성상** 참이다: 28은
  §2.3의 여섯 제외 그룹을 적용한 `git grep -lF <토큰> -- . <거부 목록>` 결과 34에서 golden
  6을 뺀 것이므로, 정의상 추적 파일이고 제외 그룹 밖이다. 별도 측정이 아니라 귀속이며,
  run-phase는 병합 트리에서 34를 다시 재면서 이 전제도 함께 확인한다(§8 R-7).

---

## §D.1 심각도 분류

| AC | 분류 | 사유 |
|---|---|---|
| AC-VSP-003 · 004 · 005 | **릴리스 차단** | 검사의 세 단언 자체. 하나라도 공허하면 이 SPEC이 아무것도 세우지 않는다 |
| AC-VSP-008 | **릴리스 차단** | pin이 없으면 003의 green이 오지 않는다 |
| AC-VSP-012 | **릴리스 차단** | 모집단이 틀리면 003·005의 판정 대상 자체가 다른 집합이 된다 |
| AC-VSP-015 | **릴리스 차단** | 모집단이 **줄어든 것**을 아무도 못 잡으면 003·005가 좁아진 집합 위에서 공허하게 초록이 된다. 012가 모집단의 **출처**를 지키고 015가 그 **도달 범위**를 지킨다 |
| AC-VSP-001 · 002 · 006 · 009 · 010 · 013 | 회귀 가드 | 검사의 형태를 지킨다. 합성 RED으로 관측된다 |
| AC-VSP-007 · 011 · 014 | 문서 | 실행으로 판정하되 코드 동작을 바꾸지 않는다 |

[HARD] 릴리스 차단으로 분류된 **여섯**은 `verification-completeness.md` §2.1의 네 요소(명령 ·
verbatim stdout · exit code · 트리 SHA)를 **run-phase에서** 갖춘다. plan-phase의 「RED 대기」
셀은 그 요소를 갖추지 않았으므로 통과로 세지 않는다.

## §D.2 완료 정의 (Definition of Done)

- 15 AC 전부 green, 각 릴리스-차단 AC의 RED이 **자기 입력으로** 한 번씩 관측되고 verbatim이
  `progress.md §E.2` 에 기록되었다.
- [HARD] 양방향 확인이 요구된 AC(002 · 004 · 005 · 010 · 011 · 012 · 013 · 014 · 015)는 잡는
  방향과 통과해야 하는 방향을 **같은 실행에서** 관측하고 둘 다 기록했다.
- [HARD] 뮤턴트 집합이 선언된 AC는 **집합 전부**를 돌렸다: AC-VSP-010 넷 · AC-VSP-011 여덟 ·
  AC-VSP-015 셋. 하나만 걸리는 것으로는 통과로 세지 않는다 — iter-2의 N2가 정확히 「보여준
  뮤턴트 하나에만 맞춰진 정규식」이었다.
- `go test ./internal/cli/... -count=1` green (병합 트리에서 재측정).
- `grep -c 'v3\.1\.3' internal/cli/testdata/*.golden` 0건.
- M0 재측정(34 / 등록부 28 / `stamp` 7 / 121)이 병합 트리에서 확인되었거나, 다르면 상수와
  문서 표가 그 값으로 고쳐졌다. **이 카드의 산출물 파일 하나마다** 제외 그룹 수와 합·전체
  트리 값이 1씩 느는 것은 **미리 선언된 규칙**이며 드리프트가 아니다(`plan.md` §B,
  `spec.md` §2.3). 정수 예측은 두지 않는다 — 감사 산출물이 계속 늘어 세 번 낡았다.
- `.moai/docs/version-management.md` 가 M5 이후에도 **영문 전용**이다
  (`grep -cP '[가-힣]' .moai/docs/version-management.md` → 0).
- CI head가 조용하고(취소 없이 완주) 초록이다.
