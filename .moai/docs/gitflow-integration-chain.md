# GitFlow 통합 체인 — 운영 절차와 실측 근거

> CLAUDE.local.md §4.1 의 운영 절차 이하에서 이관했다(card t750, 2026-09-14). **로컬 전용 문서** — 템플릿에 미러하지 않는다(내부 카드 id·SPEC id·내부 날짜 포함). 이 문서의 정본은 develop 트리의 이 사본이다(CLAUDE.local.md §0.1 판별식 준용).
> 체인 다이어그램·[HARD] 규율 5항·레인 의무의 교리는 CLAUDE.local.md §4.1 에 남아 있다 — 이 문서는 그 실행 절차와 실측 근거만 운반한다. 레인의 develop 갱신 판정식은 `.claude/rules/local/gitflow-lane-protocol.md` §11 이 소유한다.

---
**운영 절차**

```bash
# 통합 워크트리 진입 (raw `git worktree add` 금지 — 런처 경유)
moai cc -w develop                # 재진입
moai cc -w develop --branch develop  # 최초 provisioning (기존 develop 브랜치 체크아웃)

# 창 안에서
moai integration acquire --name <lane> --card <card-id>
# 흡수 전에 로컬 develop 을 먼저 최신화한다. 판정식(ref 비교)과 갱신 경로는
# `.claude/rules/local/gitflow-lane-protocol.md` §11 이 소유한다 — 여기 복사하지 않는다(두 벌이 되면 갈라진다).
git -C <카드워크트리> merge develop            # 흡수 — 대상은 로컬 develop
# 어긋나는 방향은 둘이고, 둘 다 같은 결함을 낸다.
#   앞설 때: 다른 레인이 로컬 병합을 마쳤고 리드가 아직 push 하지 않은 구간 — 원격을 흡수하면 그 착지분이 빠진 베이스에서 재측정한다.
#   뒤처질 때: 다른 레인의 병합이 이미 원격에 올라간 뒤 — 최신화 없이 로컬을 흡수하면 낡은 베이스에서 재측정한다.
# 거울상이므로 한쪽만 막으면 다른 쪽으로 새어 나간다.
# 병합 트리에서 재측정 후
git merge --no-ff <카드브랜치>                  # develop 워크트리 안에서
moai integration release
# push는 창 밖 — 리드가 레인 병합 SHA를 모아 일괄로 한다 (아래 절차)
```

**[HARD] 리드 develop 일괄 push (2026-09-02)** — 레인은 `develop`을 push하지 않는다; develop push는 리드의 단일 소관이다. 근거(실측): 3카드(t336·t372·t413)의 22커밋을 push 한 번(`09bf452c0..ad272be20`)으로 흡수 — Vercel 빌드 3→1.

1. **수집 근거** — 레인 완료 보고가 카드 id와 로컬 병합 SHA를 운반한다(위 완료 보고 필드 목록). 리드는 보고를 읽어 수집한다.
2. **push 시점 판단 — 임계값 트리거 (SPEC-MAIN-COMMIT-BAN-001)** — 리드는 `git rev-list --count origin/develop..develop`을 읽고, 계수가 `git_strategy.manual.lead_push_threshold` 이상이면 배치를 닫아 한 번 push한다. 초기값 20은 운영자 지정값이며(기록된 값 — 유도된 값이 아니다), **현재값의 원천은 설정 파일이고 이 문서는 키와 계수 명령을 명명할 뿐이다**. 임계값은 배치 닫기 **판단**만 트리거한다 — push는 창 밖 행위 그대로고(위 [HARD]), 열려 있는 통합 창을 임계값이 끊는 일은 없다. 임계값이 레인에게 push를 승인하지도 않는다(§4 — develop push는 리드의 단일 소관).
3. **원격 착지 검증** — push 뒤 `git fetch origin develop` + `git rev-parse origin/develop`으로 원격이 움직였는지 확인하고, 그 뒤에야 카드 done과 워크트리 폐기 승인을 낸다.

```bash
# 리드 — 통합 워크트리(.claude/worktrees/develop)에서, 창 밖에서
git fetch origin develop
git push origin develop
git rev-parse origin/develop   # 원격 착지 확인 — 카드 done/폐기 승인의 전제
```

**로컬 CI를 두지 않는 이유** — 검토 후 기각(2026-08-15):

- **self-hosted runner**: job을 보내는 주체가 GitHub이라 **원격에 없는 ref에는 애초에 job이 오지 않는다** — 로컬 develop 검증에 무용하다. 게다가 `modu-ai/moai-adk`는 **공개 저장소**라, 러너를 붙이면 누구나 포크 PR로 이 머신에서 임의 코드를 실행할 수 있다. 공개 저장소에 권장되지 않는 구성이다.
- **`act`**: 리눅스 컨테이너 한정이라 이 리포의 darwin×2 / windows 빌드와 macOS·Windows 통합 테스트를 재현하지 못한다. 실제 CI와 어긋나면 진단이 틀어진다.
- **비용 근거 없음**: GitHub 공식 문서 — *"GitHub Actions usage is free for self-hosted runners and for public repositories that use standard GitHub-hosted runners."* 공개 저장소는 **분 수 제한 없이 무료**다. CI를 아낄 이유가 없다.

**알려진 마찰** — BranchGuard가 읽기 전용 `git branch --list` / `git branch -vv`까지 막는다(패턴 `\bgit\s+branch\b`가 조회를 구분하지 않음). §18 독트린은 `git branch -vv`를 허용 조회로 명시하므로 과다 매칭이다(`git merge-base`는 제외 처리돼 있는 것과 대비). 우회: `git show-ref --verify refs/heads/<name>`.

**docs-site Vercel 바인딩 주의 (2026-08-27 §4.1.3에서 인계)** — docs-site의 Vercel 프로덕션 프로젝트는 여전히 특정 브랜치에 묶여 있다. `develop`에 docs-site 변경이 들어갈 때 프리뷰/프로덕션 배포가 어떻게 반응하는지는 **미검증**이다 — docs-site를 만지는 카드는 이 점을 별도로 확인한다. (2026-08-27 전환 기록과 CI 트리거 6본 `develop` 확장 내역은 git 이력의 구 §4.1.3에 남긴다)

---

## 로컬 main 잔여물 처분 절차 — 운영자 터미널 전용 (SPEC-MAIN-COMMIT-BAN-001, 카드 t1337)

> 로컬 `main`은 commit-dead다(CLAUDE.local.md §4.1 규율 6, `.claude/rules/local/gitflow-lane-protocol.md` §1) — 어느 세션도 그 안에서 커밋하지 않는다. 그런데 primary 체크아웃은 `main`에 체크아웃된 채 develop 상태 파일들을 미커밋 수정으로 들고 있는 **잔여물**이 남아 있다. 유래: `c8f245c2c` (card t1317) — `chore(policy): retire primary CLAUDE.local.md` 커밋이 main에 올라가면서 primary가 "main 체크아웃 + develop 워킹 사본" 상태가 됐고, `git rev-list --count origin/main..main` = 1(정확히 그 커밋)로 main이 origin보다 앞선 거짓 상태를 만들었다. 이 절차는 그 잔여물을 되돌린다.
>
> **실행 채널은 운영자의 자기 터미널이다 — 예외가 아니라 설계다.** 아래 2-3단계(브랜치 전환과 `git branch -f`)는 가드가 에이전트 세션에 거부하도록 설계된 바로 그 명령이다. 세션 안에서 이 절차가 `BRANCH_GUARD_VIOLATION`을 맞는 것은 올바른 동작이지 버그가 아니며, 레인은 이 단계를 실행하지 않는다(워크트리 규율 + 단일 작성자 규율).

**전제 — 보존 사본 확인(참조 전용).** primary의 `.moai/state/retired/CLAUDE.local.md` (sha256 `1db8d302…`) 존재를 확인한다. 이 절차는 새 보존 물건을 만들지 않는다 — develop 측에서 그 파일은 비추적+gitignore가 정상 상태다.

**1단계 — 재고 + 내용 안전 전제(primary에서, 운영자).** 두 검사가 모두 통과해야 진행한다(리허설 실측: `.moai/reports/t1337/rehearsal-main-residue.md`): (a) `git diff --stat`(unstaged, HEAD 대비)이 이름 붙이는 수정 파일들이 모두 `git diff --name-only c8f245c2c origin/develop`의 변경 집합 안에 있고, (b) `git status --porcelain`의 `??` 행들이 develop-추가 경로(`git diff --name-only --diff-filter=A c8f245c2c origin/develop`)와 대응한다. 그래야 워킹 사본의 모든 내용이 develop 것이라 내용 보존적 전환이 된다. 벗어나는 항목이 하나라도 있으면 중단하고 운영자가 항목마다 판별한다. (주의: `git diff develop --stat`이 비어 있어야 한다는 형태는 리허설에서 반증됐다 — develop-추가 파일이 비추적이면 그 diff가 '삭제'로 셈하므로 이 상태 등급에서는 절대 비지 않는다.)

**전환이 실제로 건너는 경계(리허설 관측).** `main`은 `CLAUDE.local.md`를 추적하지 않는다(`c8f245c2c`가 추적에서 삭제; `git cat-file -e main:CLAUDE.local.md` → 부재) — 그 파일은 전환 양쪽에서 모두 비추적+gitignore라 조용히 살아남는다. 실제 경계는 primary의 develop 상태 워킹 사본이다: (a) main이 여전히 추적하는 파일들에 대한 미커밋 수정 — **modified-tracked-set 거부**("Your local changes ... would be overwritten by checkout", 워킹 내용이 develop 것과 같아도 git은 거부한다 — 검사는 HEAD 대비 수정 여부이지 목표와의 동일성이 아니다), (b) 수정 집합을 정리한 뒤에야 드러나는 — **untracked-overwrite 거부**("The following untracked working tree files would be overwritten").

**2단계 — primary를 `develop`으로 먼저 전환** (`git switch develop`). `develop`이 다른 워크트리(통합 워크트리)에 체크아웃돼 있으면 먼저 그쪽을 정리한다 — develop은 이중 체크아웃될 수 없다. git의 거부 문구는 관측된 대로 그대로 기록한다(아래 리허설).

**3단계 — `main`이 어디에도 체크아웃돼 있지 않은 상태에서 참조 재정렬.**

```bash
git fetch origin
git branch -f main origin/main
```

대상은 `origin/main`이다 — `develop`이 아니다. main의 유일한 정당한 역할은 릴리스 PR 착지면의 동기화된 거울이고, `branch -f main develop`은 "main이 origin보다 앞선다"는 거짓 상태 — 정확히 `c8f245c2c`의 결함 모양 — 를 만들며, `origin/main` 기준 ↓N 신호를 거짓으로 읽게 한다. `origin/main`이 해석되지 않으면 절차를 중단한다. 이 단계는 워킹 트리에 영향이 없다.

**금지(그대로 유효).** `git restore CLAUDE.local.md`는 금지다(CLAUDE.local.md §0.4) — 워킹 사본을 폐기된 제3 모델로 되돌리는 회귀다. 이 절차는 그 파일을 만지지 않는다.

**리허설 필수 — 실제 상태 등급을 재현할 것.** origin/main에서의 깨끗한 클론(그 파일이 tracked+clean)은 전환 시 거부가 아니라 **조용한 삭제**를 관측하므로 리허설 대상이 아니다. 레시피: 클론 → `git switch --detach c8f245c2c`(bare SHA 전환은 관측상 거부된다 — git이 `--detach`를 제안; 리허설 Observed 0) → develop 상태 내용을 main 추적집합 위에 unstaged 수정으로 재현(`git restore --source=develop -- <develop-tracked 경로들>`)하고 develop이 추적하는 비추적 파일을 만든다 → 절차 단계를 실행 → 실제 거부 집합(modified-tracked-set / untracked-overwrite 형태)을 기록한다. git이 무엇을 하든 그것이 절차가 문서화하는 것이다. 실측 로그: `.moai/reports/t1337/`.

**후속 — 운영자 문서 동기(REQ-4.6).** 절차 후 primary 전용 `CLAUDE.local.md` §4.1 워킹 사본에 위 commit-dead 규율과 `lead_push_threshold` 키 인용을 운영자가 직접 반영한다(비추적 파일 — primary에서만 도달 가능; 레인이 쓰지 않는다).

**후행 조건 — 재무장된 전환 경계(결함이 아니라 문서화된 상태).** `main = origin/main`이 되면 main은 `CLAUDE.local.md`를 **다시 추적**한다(`c8f245c2c`는 origin/main 계보에 없다) — 반면 develop 측 primary는 그 파일을 비추적으로 든다. 따라서 나중에 primary를 `main`으로 전환하려면 "untracked working copy would be overwritten" 거부를 맞는다. 이것은 재무장된 경계로, 절차가 문서화하는 상태이지 고칠 결함이 아니다 — 그때는 그 사본을 임시로 밖에 옮겼다 돌려놓는 운영자 판단이 필요하다.

