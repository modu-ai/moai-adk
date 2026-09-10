# t310 — git-flow 문서 정합 감사 판정

- 카드: t310
- 브랜치: `WT-gitflow-doc-align`
- 워크트리: `.claude/worktrees/t310`
- 측정 트리 base: `6310dbf28` (= `origin/develop` at audit time)
- 판정: **(b) 잔존 모순 있음 → 좁혀서 수리 완료**

---

## Claim

카드 t310이 지목한 세 문서 중 **문서 3은 정합 완료**, **문서 1·2에는 잔존 모순 11건**이 남아 있었다.
리드가 의심한 지점(`git-workflow-doctrine.md:120·:134` tier routing 표)은 **이미 `:113-114` 고지가
"아래 tier routing 표와 본 섹션 나머지"라고 명시적으로 범위 한정**하고 있어 형태 (c)에 해당했고,
실제 미수리는 **구조 계층**(브랜치 다이어그램 / 명명 표 / 머지 전략)과 **릴리스 분기점**이었다.

## Evidence

모든 줄 번호는 **감사 시점 워크트리(base `6310dbf28`)** 판독값이다. primary 체크아웃 판과 다를 수 있다.

### 문서 3 — `.claude/rules/local/repo-local-pr-policy.md` (20줄): 정합 완료, 수정 0건

- `:2` description이 git-flow를 명시. `:10-14`가 현행 카드 워크플로(develop 분기 / 카드 PR 없음 /
  origin/develop 판정 / 릴리스 전용 PR ceremony)를 4항으로 서술. `:16`의 종전 all-tier PR 조항은
  취소선 + `[RETIRED 2026-08-27]`. 잔존 모순 없음.

### 문서 1 — `.moai/docs/git-workflow-doctrine.md` (431줄): 잔존 9건

| # | 줄 | 잔존 내용 | 형태 | 심각도 |
|---|---|---|---|---|
| D1-1 | `:306` | `git checkout -b release/v2.15.0 main` — 릴리스 브랜치를 `main`에서 분기 | 무표시 라이브 명령 | **최상** |
| D1-2 | `:64-76` | §18.1 브랜치 다이어그램에 `develop`이 없고 feat/fix/docs/chore가 `main`에서 분기 | 무표시 라이브 구조 | 상 |
| D1-3 | `:82-92` | §18.2 [HARD] 명명 표 11종에 카드 브랜치 `WT-<slug>` 부재 | 무표시 라이브 [HARD] | 상 |
| D1-4 | `:100` | §18.3 [HARD] "feature/fix/chore/docs → main, squash" | 무표시 라이브 [HARD] | 상 |
| D1-5 | `:147` | §18.4 패치 행 "`fix/*` PR (self-merge) 여러 개 → main" | 무표시 라이브 | 중 |
| D1-6 | `:319-324` | §18.8 Patch Release "fix 브랜치에서 수정 + PR + squash merge" | 무표시 라이브 | 중 |
| D1-7 | `:1` | 제목 "Enhanced GitHub Flow" — `:19`는 현재 모델을 git-flow로 선언 | 자기모순 제목 | 중 |
| D1-8 | `:329` | "Enhanced GitHub Flow 공식 인프라" 라벨 | 명명 잔재 | 하 |
| D1-9 | `:361` | 금지사항이 `feat/SPEC-*`를 정상 관례로 지목 (같은 목록 `:358`은 이미 개정됨) | 목록 내부 불일치 | 하 |

**D1-1이 가장 무거운 이유**: 같은 문서 `:8` 헤더 고지가 "릴리스 브랜치가 `main` 대신 `develop`에서
분기한다는 점만 바뀐다"고 **직접 적어놓고** 본문 명령은 `main`을 그대로 둔다. 문서가 자기 자신과
어긋나며, 이건 은퇴한 카드 경로가 아니라 **여전히 구속력 있다고 그 고지가 선언한 릴리스 경로**다.
정본 `gitflow-lane-protocol.md:113`: "`release/vX.Y.Z`를 **`develop`에서** 분기".

### 문서 2 — `.moai/docs/git-local-workflow-doctrine.md` (214줄): 잔존 2건

| # | 줄 | 잔존 내용 | 형태 |
|---|---|---|---|
| D2-1 | `:211` | "§23.9 PR-mandatory routing (모든 tier PR 경유, 2026-07-20)은 single/multi-session 공통 기본값" | 무표시 라이브 |
| D2-2 | `:174` | "Owner 명시 … 모든 tier에서 PR 생성은 manager-git의 책임" — `:161` 고지가 "표와 routing 결정 흐름"만 이름을 대서 이 단락은 미포함 | 형태 (c) |

## Baseline-attribution

- 트리: `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t310`
- base: `git rev-parse --short HEAD` → `6310dbf28`; `git merge-base --is-ancestor HEAD origin/develop` → rc 0;
  `git rev-list --count HEAD..origin/develop` → `0` (base가 origin/develop과 동일)
- 판독: `cat -n` 전문 판독 — 문서 1 431줄 / 문서 2 214줄 / 문서 3 20줄 / 정본 `gitflow-lane-protocol.md` 127줄.
  발췌 grep이 아니라 전수 판독이다.
- 대조 정본: `.claude/rules/local/gitflow-lane-protocol.md` §1(분기점) · §2(통합 면) · §4(push) · §10(릴리스)

## 수리 내역 (13개 지점)

문서 1 9건 · 문서 2 4건. 전부 문자열 단일-매치 치환으로 적용했고, 매치 수가 1이 아니면 스크립트가
중단하도록 해 오치환 가능성을 배제했다.

- D1-1 → `release/v2.15.0 develop` + 개정 주석
- D1-2 → 다이어그램을 `main` ← `release/*` ← `develop` ← `WT-<slug>` 2단 구조로 교체 + [HARD] 개정 고지
- D1-3 → 표 최상단에 `WT-<slug>` 행 추가 + "카드 경로에서 더 이상 쓰이지 않음" [HARD] 고지
- D1-4 → 표 위 [HARD] 고지("카드 통합은 이 표의 대상이 아니다") + 해당 행 취소선 `[RETIRED 2026-08-27]`
- D1-5 / D1-6 → 현행 git-flow 경로로 본문 교체
- D1-7 → 제목을 `git-flow (2026-08-27 전환; 종전 모델: Enhanced GitHub Flow)`로
- D1-8 / D1-9 → 라벨·금지사항 문구 교정
- D2-1 → 현행 카드 routing 서술로 교체 + 종전 조항 취소선. 완화책 서술도 워크트리 격리로 정정
- D2-2 + `:178` 라우팅 흐름 heading → 인라인 `[RETIRED 2026-08-27]` 마커 추가(grep 독자 대응)
- D2 `:198` → pre-spawn baseline `origin/main`이 배포 SSOT 문구임을 밝히고, 로컬 카드 작업에서는
  `origin/develop`으로 읽으라는 주의를 삽입(SSOT 자체는 건드리지 않음)

## Gaps — 명시적으로 관측하지 않은 것

- **`CLAUDE.local.md`**: 카드 지시대로 손대지 않았다(t308 소관). 읽지도 않았다.
- **정본 2종**(`gitflow-lane-protocol.md`, `CLAUDE.local.md` §4.1): 대조 기준으로만 읽었고 바꾸지 않았다.
- **세 문서 밖의 다른 문서**: 스캔하지 않았다. 이 감사의 범위는 카드가 지목한 3개 파일이다.
- **빌드/테스트**: 문서 전용 변경이라 돌리지 않았다. Go 코드 변경 0건.
- **템플릿 미러**: 두 문서 모두 `internal/template/templates/.moai/docs/` 에 대응 파일이 **없음**을
  `ls`로 확인했다(둘 다 `No such file or directory`). Template-First 의무와 중립성 가드 대상이 아니다.

## Residual-risk — 수리하지 않고 보고하는 것 (범위 밖)

1. **`git-workflow-doctrine.md` §18.12 BODP — [자기정정 2026-08-27] 최초 판정이 미검증이었다.**

   최초 verdict는 "§18.12의 `→ main @ origin/main` 기본값은 `internal/bodp/relatedness.go`의 실제
   코드 동작을 정확히 기술한 것이므로 코드가 전환을 못 따라온 것"이라고 적었다. **이 주장은 코드를
   읽지 않고 문서의 자기 인용을 그대로 받아 적은 것이었다.** 리드 요청으로 실측한 결과 **거짓**이다.

   실측:

   | 명령 | 출력 | 뜻 |
   |---|---|---|
   | `ls internal/bodp/` | `No such file or directory` | 디렉터리 부재 |
   | `git ls-files -- 'internal/bodp' \| wc -l` | `0` | 추적 파일 0건 |
   | `grep -rn 'applyMatrix\|relatedness\.go' --include='*.go' .` | 출력 없음 | Go 참조 0건 |
   | `git ls-files \| grep bodp` | `.moai/bodp/plan/SPEC-V3R4-HARNESS-003-2026-05-15.json`, `.moai/reports/expert-debug/bodp-audit-trail-leak-2026-05-09.md` | 남은 것은 계획 JSON과 옛 리포트뿐 |

   **`internal/bodp/relatedness.go` 는 이 트리에 존재하지 않는다.** §18.12가 `Check()` / `applyMatrix()`
   함수를 파일 경로까지 대며 인용하는 것은 **죽은 코드 경로 인용**이다.

   현재 살아 있는 표면 2곳(실측):
   - `.claude/rules/moai/development/branch-origin-protocol.md:26` — [ZONE:Evolvable] [HARD]
     "When no signal fires, the recommendation is `origin/main`". 같은 파일 `:42`가 동일한 8행 행렬을
     보유. 즉 **기본값의 실제 소유자는 Go 코드가 아니라 이 배포 룰 파일**이다.
   - `internal/cli/doctor.go:894-906` `checkBODPConfig` — `.moai/branches/` 디렉터리 존재 여부만
     `os.Stat` 한다. **base 선택 로직이 전혀 없다.**

   **따라서 결함의 종류가 바뀐다.** "코드가 전환을 못 따라옴"(코드 카드)이 아니라 **① 문서가 없는
   패키지를 인용함(스테일 인용) + ② 기본값의 실제 SSOT는 배포되는 룰 파일**이다. ②는 GitHub Flow를
   쓰는 다운스트림 사용자에게는 `origin/main`이 옳으므로 이 리포 사정으로 바꿀 수 없고, 로컬
   git-flow와의 간극을 어디서 흡수할지가 별도 설계 판단이다.

   **수리하지 않은 이유는 그대로 유효하나 근거가 다르다** — 코드와 어긋나서가 아니라, ①은 git-flow
   정합이 아닌 스테일-인용 결함이고 ②는 배포 SSOT 소관이라 이 카드의 3파일 범위 밖이기 때문이다.
   후속 카드는 **코드 변경 카드가 아니라 문서 인용 정정 + 배포 룰 SSOT 판단 카드**로 잡아야 한다.
2. **`git-local-workflow-doctrine.md` 제목/`:12` 헤딩의 "PR-mandatory 1-person OSS".**
   릴리스 경로에는 여전히 참이고, `:5` 전환 고지가 제목 바로 아래 4줄 거리에 있어 grep 독자도
   즉시 만난다. 오도 위험이 낮다고 판단해 남겼다 — 이견이 있으면 리드 판단.
3. **`git-workflow-doctrine.md` 머지 표의 `plan/* → main` · `dependabot/* → main` 행.**
   `plan/*`는 카드 경로에서 사실상 사어이나 표 위 개정 고지가 범위를 한정한다. `dependabot/*`의
   실제 대상 브랜치는 GitHub 기본 브랜치 설정에 달려 있고 **측정하지 않았다** — 미검증이라 손대지 않았다.
4. **`§18.8` 3단계의 `git checkout main && git pull`.** 공유 체크아웃 브랜치 가드와 부딪히는 서술이나
   전환 이전부터 있던 별개 사안이고 릴리스 하네스 소관이다.
