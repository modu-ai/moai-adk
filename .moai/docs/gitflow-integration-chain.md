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
2. **push 시점 판단** — 리드가 배치를 닫을 시점을 정해 한 번 push한다.
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

