# 동기화 단계 감사 보고서 — SPEC-REVIEW-SECRET-SCAN-REFS-001

- 카드: t629 · 감사자: sync-auditor(독립, 읽기 전용) · 날짜: 2026-09-11
- 대상 트리: 워크트리 `.claude/worktrees/t629`, 브랜치 `WT-secret-scan-refs`, HEAD `3ac00a284`(동기화 커밋)
- 평가 프로필: `.moai/config/evaluator-profiles/default.md`(SPEC에 `evaluator_profile` 없음 → `harness.yaml` `default_profile: "default"`), 평면 가중 채점
- 판정 대상 변경: 카드 기준점 `feeecc980` → 편집 커밋 R = `cc4092513` → 종결 앵커 K = `c10626ab7` → M5 `9a9cedc3d` → 동기화 `3ac00a284`

## 종합 판정: PASS

| 차원 | 점수 | 판정 | 근거 요약 |
|---|---|---|---|
| Functionality (40%) | 85/100 | PASS | 재측정 가능한 AC(002·003·004·005·006·007·010 순서·011·015·016)를 이 트리에서 다시 재 모두 통과. 픽스처 AC(001·008·009·013·014)와 셀 ③의 기록은 판정식·종료 코드·대조군이 갖춰져 있어 타당. 분리 HEAD 재보고(F2)·저장소 선갱신(F3)은 요구사항 위반이 아닌 가장자리 사례 |
| Security (25%) | 82/100 | PASS | Critical·High 없음. 파괴적 명령은 고정 경로 `rm -r` 하나. 억제는 정확한 값 비교만. 평문 자격 증명 사본이 `.moai/state/`에 남는 점(F4)은 Medium 이하로 문서화 대상 |
| Craft (20%) | 78/100 | PASS | 두 사본 바이트 동일, 카탈로그 해시 재생성과 테스트 통과, 엄격 누출 검사 통과. 13단계 억제 절차의 복잡도(운영자 결정으로 유지)와 작업 트리 단계의 명령 부재(F5) |
| Consistency (15%) | 80/100 | PASS | 고정 절차 문장 4줄이 섹션에 그대로 존재, 내부 ID·비용 수치·언어 편향·7-8자 16진 없음. CHANGELOG의 "유일한 미포함 부류" 표현이 문서 본문보다 넓게 주장(F1) |

- 가중 조화평균: 0.4/85 + 0.25/82 + 0.2/78 + 0.15/80 = 0.012085 → **82.7**
- 필수 통과 방화벽(Functionality·Security): 둘 다 통과.
- 차단(blocking) 결함: **0건**. 아래 결함은 모두 선택(optional)으로 분류했다 — 판단 근거는 각 항목에 적었다. F1은 리드가 차단으로 올릴 여지가 있는 경계 사례라 맨 앞에 두었다.

## 증거 — 이 트리에서 재측정한 것

모든 출력은 세션 스크래치패드 `SP/sa/`(저장소 밖)에 두었고 종료 코드는 파이프 없이 읽었다.

| 확인 | 명령 | 관측 출력 |
|---|---|---|
| 트리 상태 | `git rev-parse --show-toplevel`; `git rev-parse --short HEAD`; `git status --porcelain \| wc -l` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t629`; `3ac00a284`; `0` |
| AC-002 | `/usr/bin/grep -c` 정확한 동등성 문구, `'no finding class is dropped'`, `-ci 'same coverage'`, 각각 LOC·TPL | 여섯 값 모두 `0`, 마지막 exit `1` |
| AC-003 (a) | `cmp LOC TPL`; `diff -q LOC TPL` | exit `0`, 출력 없음 |
| AC-003 (b) | `git diff --stat feeecc980 c10626ab7 -- LOC TPL` | 두 파일 모두 `53 +++…---`; `2 files changed, 94 insertions(+), 12 deletions(-)` |
| K 이후 사본 불변 | `git diff --stat c10626ab7 HEAD -- LOC TPL` → 파일 | `0` 바이트 |
| AC-004 (HEAD까지 확장) | `git diff feeecc980 HEAD --output=SP/sa/card-diff-head.txt`; `wc -l`; 추가 줄 REGEX grep | `4123` 줄; count `0`, exit `1`. 대조: `+`로 시작하는 PEM형 헤더 한 줄(조각 조립) → `1` |
| AC-005 grep | `SPEC-`, `t629`, 비용 수치(`12,157\|12157\|76\.084\|0\.207\|12490\|60\.01\|0\.229`), 16개 언어명 `-ciwE`, `R language`, 7-8자 16진, 날짜 형태 | 전부 `0`. 7-8자 16진 대조 `see abcdef1 here` → `1` |
| AC-005 엄격 누출 | `MOAI_TEMPLATE_LEAK_STRICT=1 go test -v ./internal/template/ -run '^TestTemplateNoInternalContentLeak$' -count=1 > SP/sa/leak.txt 2>&1` | exit `0`; `--- PASS: TestTemplateNoInternalContentLeak (1.33s)` 개수 `1`; `--- FAIL` 개수 `0` |
| AC-006 | `SECTION`(61줄)에서 `log -p` 줄 → `SP/sa/cmds.txt` | `2`줄; `-vc --all` `0`; `-c --stdin` `1`; `-cF '^%(objectname)'` `1`; `-ci 'every ref'` `2`; HEAD-SHA 문구 `0` |
| AC-015 | `/usr/bin/grep -cE -- 'REGEX' LOC TPL` | `0`, `0`, exit `1`. 대조(조각 조립 PEM 헤더) → `1` |
| AC-016 (b) | `git show f167a9cd8:PROG`, `git show cc4092513~1:PROG` → `sed -nE` 고정 블록 추출 → `cmp` | `28`줄, `28`줄, `cmp_exit=0` |
| AC-016 (a) + 미포함 문장 | 고정 블록의 `Tip recording:`·`Scan command:`·`Missing-tip handling:`·`Uncovered commits:` 줄(접두사 제거, 각 1줄·비어 있지 않음)을 `grep -cF -f`로 `SECTION`에 대조 | `1`, `1`, `1`, `1` |
| AC-007 | `git merge-base --is-ancestor af7eb142b cc4092513` | `ac007_exit=0` |
| AC-010 순서 | `git merge-base --is-ancestor 7e06766c8 023a25e7a` | `ac010_order_exit=0` |
| AC-011 | `git merge-base --is-ancestor 6e56840d5 cc4092513`; 같은 형태로 `f167a9cd8` | `0`; `0` |
| 카탈로그 | `go test -v ./internal/template/ -run '^(TestCatalogHashCoversSkillSubfiles\|TestManifestHashFormat)$' -count=1` | exit `0`; `--- PASS: TestCatalogHashCoversSkillSubfiles (0.07s)`, `--- PASS: TestManifestHashFormat (0.08s)`, `ok … internal/template 0.455s` |
| AC 개수 | `acceptance.md`의 `AC-0NN` 고유값 | `16` |
| SPEC 파일·CHANGELOG의 REGEX | `/usr/bin/grep -cE -- 'REGEX'` 5개 파일 | 전부 `0` |
| 수명 주기 | `grep -c '^status: completed$' spec.md`; 동기화 커밋 제목 | `1`; `docs(SPEC-REVIEW-SECRET-SCAN-REFS-001): sync-phase artifacts, 3-phase close (card t629)` |
| lint | `moai spec lint SPEC-REVIEW-SECRET-SCAN-REFS-001` | exit `0`, `✓ No findings — all SPEC documents are valid` — 판정 빌드는 설치된 `v3.2.0-rc.7`(이 트리에서 만든 빌드가 아님, §E.4 항목과 동일) |
| 옛 체크포인트 이름 잔존 | `grep -rln 'secrets-scan-checkpoint' internal/ .claude/ docs-site/content` | docs-site 4개 로케일 `utility-commands/moai-review.md`만 적중, 각 77행. 템플릿·로컬 트리 0건 |

### 스크래치패드 픽스처 확인 (마커는 조각으로 조립한 PEM 레이블만, 개수만 기록)

`SP/sa/fx`, `git version 2.50.1 (Apple Git-155)`.

| 확인 | 절차 | 관측 |
|---|---|---|
| 분리 HEAD | `main`에 깨끗한 커밋 → `switch --detach` → `DETACHCELL` 마커 커밋 → `for-each-ref` 저장소 기록(1줄) → 스캔 명령 → 다시 기록(`cmp` exit `0`) → 다시 스캔 | 두 스캔 모두 exit `0`, `DETACHCELL` `1`과 `1`; 두 번째 스캔 범위 `git log --format=%H --all --stdin` `1`줄 |
| 빈 저장소 파일 | 0바이트 저장소를 `--stdin`으로 공급 | exit `0`, `DETACHCELL` `1` — 전체 ref 범위로 동작 |
| 빈 입력의 `split` | 빈 파일에 `split -a 6 -l 1` | exit `0`, `/usr/bin/find -type f` 결과 `0`줄(파일 생성 안 함). 참고: 처음 `ls \| wc -l`은 `3`을 찍었는데 프로필 alias 오염이라 버리고 `find`로 다시 쟀다 |
| 누락 팁 오류 문구의 로캘 | 존재하지 않는 40자 16진을 캐럿 저장소로 공급, `LC_ALL=ko_KR.UTF-8` | exit `128`, `fatal: bad object 1111111111111111111111111111111111111111` — 이 git 빌드에서는 영문 그대로 |

## 기록된 증거의 타당성 판단 (픽스처 기반 AC)

- **AC-008 / 셀 ①**(`progress.md` § Gate cell 1): 기준 스캔 0바이트, 모든 스캔 exit 0, `is-ancestor` exit 1을 집계 직전에 재확인, R2에서 `HEADCELL 1`·`SIDECELL 1`. 저장소 공급 유무에 따른 범위 2 대 3(D22)도 기록. 판정식과 수치가 일치한다.
- **AC-009 / 셀 ②**: 구성 확인 `cat-file -e` exit 1과 `--batch-check`의 ` missing` 이중 판독, R4 exit 128과 `fatal: bad object` 1줄, 캐럿 제거 대조(캐럿 저장소 `0` 대 제거본 `1`)로 탐지 판별식이 공허하지 않음을 보였다. 대체 스캔 exit 0, `GONECELL 1`. 타당.
- **AC-010 / 셀 ③**: `B − A + L = 12490 − 12488 + 0 = 2`, 증분 범위 `2`, `B2 = B`로 스캔 중 ref 이동 없음. 첫 실행 범위를 스캔 **후**에 잰 점은 기록된 갭 그대로다. 측정 도구 차이(`real` 대 셸 `time`의 `total`)를 승인한 두 번째 리드 메시지는 증거 커밋 안에만 인용돼 있고 별도 선행 커밋이 없다 — §E.4 (a)의 리드 소관 항목과 같은 사안이며 새 결함으로 세지 않는다.
- **AC-001**(M3 `fx1`): B·H·S·T·N 다섯 칸 모두 요구값 충족, R3 `SIDECELL 0`은 REQ-002와 정합. 13단계 억제 명령까지 매 회차 전부 실행한 기록이다.
- **AC-013·AC-014**(M3 `fx2`): 한 커밋·한 헝크(`^@@` `1`)에 LIST·NEAR·OTHER·MIX가 나란히 있어 경로·파일·헝크·줄 단위 억제 뮤턴트를 모두 가른다. 추출본 기준 `LISTCELL 0`, 나머지 전부 `1`. `xargs cat`으로 `find`·`sort` 순서가 `split` 순서와 같음을 `cmp`로 보인 보조 확인도 있다. 타당.
- **다이제스트 값**: 등재 값을 조립하지 않는다는 제약 때문에 직접 재계산하지 않았다. M2 기록의 `git hash-object`와 `blob 21\0` 헤더를 붙인 `shasum -a 1` 독립 재계산 일치, 그리고 AC-013에서 해당 행이 실제로 억제된 결과로 판단했다(갭 G3).

## 결함 목록 (구조화)

- **F1** [Medium] [optional — 리드 판단으로 차단 승격 가능] `CHANGELOG.md:12` — "The one class still outside every step is now stated in a sentence of its own: commits no ref and no HEAD reaches" 는 미포함 부류가 하나뿐이라고 읽힌다. 그런데 문서 본문(`review.md:122`)은 병합 커밋 자체의 변경으로만 들어온 줄(충돌 해결 등)을 **어느 스캔도 보고하지 않는다**고 따로 적고, M3의 `MERGECELL` 대조(두 스캔 `0`, `--diff-merges=first-parent` 대조 `1`)가 이를 실측했다. CHANGELOG는 이 한계를 빼고 "유일한 부류"라고 적어, 이 SPEC이 고치려던 것과 같은 모양(범위 과대 주장)을 릴리스 노트에 남긴다. 문장이 "커밋" 단위로 한정돼 있어 문자 그대로는 방어 가능하고, SPEC 요구사항(REQ-003·005)은 워크플로 문서에 걸려 있어 차단으로 분류하지 않았다. **수정 지시**: 해당 문장을 "Two limits stay stated: commits no ref and no HEAD reaches (such as reflog-only commits) are outside every step, and a line that only a merge commit's own changes introduce is reported by neither scan." 식으로 바꾼다. 신뢰도 높음.
- **F2** [Low] [optional] `review.md:105` — 스캔 명령 설명의 줄표 구절 "the commits that became reachable since the last completed scan"은 분리 HEAD에서만 도달 가능한 커밋에 대해 사실과 다르다. `for-each-ref`는 HEAD를 기록하지 않으므로 그런 커밋은 매 리뷰마다 다시 스캔·재보고된다(스크래치 픽스처: 저장소 `cmp` exit 0인 두 회차 모두 `DETACHCELL 1`). 방향은 과다 보고라 커버리지 손실은 아니며, 주절("every commit reachable from any ref or from HEAD that no recorded tip reaches")은 정확하다. **수정 지시**: 줄표 구절에 "(a commit reachable only from a detached HEAD is scanned again at every review, because the store records refs, not HEAD)"를 붙이거나 팁 기록에 HEAD를 추가한다(후자는 절차 변경이라 새 게이트 회차 대상 — REQ-013). 신뢰도 높음(실측).
- **F3** [Low] [optional] `review.md:97` + `review.md:130-145` — 저장소 교체는 스캔 exit 0에만 묶여 있고 억제·발견 단계 완료에는 묶여 있지 않다. 억제 파이프라인이 중간에 실패한 뒤 실행자가 멈추면, 해당 커밋들은 스캔된 것으로 기록되지만 발견은 보고되지 않은 채 다음 리뷰가 `secrets-scan-output.txt`를 덮어쓴다. 고정 절차(REQ-013)가 이 순서를 정했으므로 이번 카드에서 바꿀 수는 없다. **수정 지시**: 후속 카드에서 "저장소 교체는 발견 파일이 만들어진 뒤"로 순서를 옮기고 게이트를 다시 거친다. 신뢰도 중간(추론, 미실측).
- **F4** [Medium] [optional] `review.md:102,108,134-144` — 절차가 이력의 자격 증명 값을 평문으로 담은 파일(`secrets-scan-output.txt`, `-matches`, `-distinct`, `-table`, `-kept`, `-unsuppressed`, `-findings`)을 `.moai/state/`에 만들고 지우지 않는다(`rm -r`은 분할 디렉터리만). 이전 절차는 표준 출력으로만 냈으므로 저장 상태의 사본은 이번 변경으로 새로 생긴다. 추적 파일로 새지 않는 근거는 배포 `.gitignore`의 `.moai/state/` 항목(`internal/template/templates/.gitignore:240`; 이 저장소는 `.gitignore:267` `**/.moai/state/`)과 update의 gitignore 병합뿐이며, 섹션은 이 의존을 말하지 않는다. **수정 지시**: 섹션에 "these files hold matched values; `.moai/state/` must stay ignored, and remove the files once findings are reported" 한 문장을 추가한다. 신뢰도 높음(경로 확인), 실제 유출 사례는 미관측.
- **F5** [Low] [optional] `review.md:126` — "the example-value rule below applies to its matches as well"이라고 하지만 억제 파이프라인은 `secrets-scan-output.txt`(히스토리 스캔 출력)만 읽고, 작업 트리 단계에는 명령이 없다(M3 갭에도 기록). 게다가 `.gitignore`를 따르지 않는 작업 트리 검색은 F4의 평문 사본을 다시 찾아낸다. 작업 트리 명령 부재 자체는 변경 전부터 있던 상태. **수정 지시**: 후속에서 작업 트리 명령(예: 추적 파일만 보는 `git grep`)과 억제 파이프라인 연결을 명시한다. 신뢰도 높음.
- **F6** [Low] [optional] `review.md:111` — 누락 팁 판별이 git 오류 문구의 영문 리터럴 `bad object`에 의존한다. 이 머신의 Apple Git은 `ko_KR` 로캘에서도 영문으로 냈지만, 번역 카탈로그를 싣는 다른 git 빌드는 재지 않았다. 번역되면 "그 밖의 비0 종료 = 스캔 실패, 저장소 유지" 분기로 떨어져 **실패를 보고하며 멈춘다**(조용한 누락이 아니라 fail-closed 정체). **수정 지시**: 후속에서 `LC_ALL=C` 고정 또는 판별을 팁 이름 일치만으로 하는 방안을 검토한다. 신뢰도 중간(한 빌드만 실측).
- **F7** [Info] [optional] 억제는 "정규식이 일치시킨 텍스트"를 비교하지 토큰 전체를 비교하지 않는다. 등재 값 뒤에 영숫자가 더 붙은 긴 토큰은 `grep -o`가 앞 20자만 잘라내므로 억제된다. REQ-011의 문언과 정확히 일치하는 동작이고, 실제 AWS 액세스 키 ID는 정확히 20자라 실질 위험은 낮다. SHA-1 블롭 다이제스트의 한계: 등재되지 않은 값이 억제되려면 SHA-1 제2원상이 필요하며(알려진 실용 공격 없음, 알려진 공격은 두 입력을 모두 고르는 충돌), SHA-256 객체 형식 저장소에서는 아무것도 억제되지 않는 과다 보고 방향이다(문서에 명시). 추론, 미실측.
- **F8** [Info] [optional, 기존 결함] `review.md:99` "an explicit full-scan flag is passed" — `/moai review`에 그런 플래그가 정의돼 있지 않다(`--full`·`full-scan` 검색 결과 이 줄뿐). 기준점 `feeecc980`에도 같은 문구가 있었으므로 이번 변경이 만든 것은 아니다.

## 리드 소관 항목 확인 (§E.4, 감사 실패 사유로 세지 않음)

- (a) AC-010 문구 `real` 대 승인된 셸 `time`의 `total`: 기록과 일치. 도구 변경 승인이 별도 선행 커밋 없이 증거 커밋 안에만 인용된 점도 함께 판단 대상이다.
- (b) AC-011의 가장 오래된 G가 폐기된 1회차 `6e56840d5`로 풀림: 확인(3줄 중 첫 줄). `f167a9cd8`의 선행도 재측정으로 확인했다.
- (c) §E.3 `run_commit_sha: pending-backfill`: 실제 M5 커밋은 `9a9cedc3d`.
- (d) docs-site 4개 로케일 77행 스테일: 확인.
- (e) 13단계 억제 절차의 복잡도: F3·F5와 함께 후속 후보로 남긴다.
- (f) lint 판정 빌드 귀속: 이번 재측정도 설치된 `v3.2.0-rc.7`로 했으므로 같은 갭이다.

모두 기록이 사실과 맞으며 잘못된 항목은 찾지 못했다.

## 기준 귀속 (Baseline-attribution)

- 모든 재측정은 이 실행에서, 트리 `3ac00a284`(K 이후 두 사본 불변, 위 `diff --stat` 0바이트)에 대해 수행했다.
- `go test` 두 건은 이 트리에서 컴파일됐다. `moai spec lint`만 설치 빌드(`v3.2.0-rc.7`)로 판정됐다.
- 픽스처 AC와 셀 ①~③의 수치는 `progress.md` §E.2의 기록을 판단한 것이며 재실행하지 않았다(셀 ③ 전체 이력 스캔은 금지 조건).

## 미검증 (Gaps)

- G1: 픽스처 AC(001·008·009·013·014)와 셀 ③은 재실행하지 않고 기록으로 판단했다. 원본 스캔 출력은 레인의 스크래치패드에만 있고 반출되지 않았다(기록된 알려진 손실).
- G2: 교차 모델 감사(`mcp__moai__audit_multi`·codex·GLM)는 돌리지 않았다. 설정에 `audit_model` 키가 없고, 트리가 깨끗해 미커밋 diff가 없으며, `baseBranch` diff는 이 카드와 무관한 이력을 대량 포함한다.
- G3: 등재 다이제스트가 AWS 문서 예시 값의 블롭 이름과 같은지는 값 조립 금지 제약 때문에 직접 재계산하지 않았다.
- G4: Linux·Windows(GNU `split`·`sort`·`grep`, Git for Windows 셸), 토큰(`ghp_`) 대안, 태그·원격 추적 ref·stash, 여러 개의 누락 팁, "그 밖의 비0 종료" 분기, 스캔 도중 착지한 커밋은 이 감사에서도 재지 않았다. 빈 저장소 파일과 빈 입력의 `split`은 macOS에서만 확인했다.
- G5: `make build`·임베드 검사는 범위 밖(리드 소관).

## 잔여 위험 (Residual-risk)

- 절차는 문서로만 존재한다. 실제 리뷰에서 실행자가 13단계를 빠짐없이, 올바른 순서로 따르는지는 관측 대상이 아니다(SPEC §5 범위 밖). 단계 하나를 건너뛰면 억제가 과다 또는 과소로 기울 수 있다.
- F3의 순서 때문에 억제 단계 실패 후 방치되면 발견이 조용히 사라지는 경로가 남는다.
- F4의 평문 사본은 `.moai/state/`가 무시되지 않는 프로젝트에서 커밋될 수 있다.
- 셀 ③의 시간 수치는 경합 중인 공유 머신 한 쌍의 측정이다. 판정식은 기록의 완전성·일관성만 본다.

## 권고

- F1은 릴리스 PR에 이 항목이 실리기 전에 한 문장으로 고치기를 권한다(비용이 가장 작고, 사용자에게 보이는 범위 주장이다).
- F2·F4는 문서 한 문장씩으로 해결되지만 섹션 변경이므로 REQ-013·게이트 규칙에 비추어 이 카드 안에서 할지 후속으로 돌릴지 리드가 정한다. F4는 절차가 아닌 설명 문장 추가라 게이트 재측정 대상이 아닐 가능성이 높다(판단, 미확정).
- F3·F5·F6은 13단계 단순화 후속 카드에 묶는다.
