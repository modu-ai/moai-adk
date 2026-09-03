# t315 — 판정서 (재개 회차)

측정 트리: `.claude/worktrees/t315`, 브랜치 `WT-release-notes-gitflow`,
HEAD `504d5ab5a`(로컬 develop `bac2cf15b` 흡수 후). 측정일 2026-09-03.

이 문서는 `d7a-behavior-delta.md`를 **대체하지 않는다.** 그 문서는 삽입을 결정하기까지의
논증이고, 이 문서는 그 뒤에 남은 이관 축의 판정이다. 원문은 그대로 둔다.

---

## 0. 배차 전제 정정 — §3·§4는 미완 표시가 아니다

배차문은 「§3·§4의 제목이 "아니다"라고 적혀 있으므로 저작이 끝났으나 자리가 맞는지
미판정」이라고 읽었다. 실측은 다르다.

| 측정 | 명령 | 관측 |
|---|---|---|
| 문안 삽입 여부 | `grep -c 'card t315, D7a release-note duty' CHANGELOG.md` | `1` (254행) |
| 삽입 커밋 | `git show --stat bc6ae180e` | `CHANGELOG.md \| 1 +` |
| §5 초안 ↔ 삽입 블록 ↔ 실제 삽입 | 세 형태 대조 | 일치 |

§4의 "아니다"는 **삽입 이전의 CHANGELOG 상태**에 대한 판정이다("있는 항목이 다른 질문에
답하고 있다"). 그 판정이 삽입을 낳았고, 삽입은 같은 커밋에서 이루어졌다. 즉 §3·§4는
미완 표시가 아니라 **작업의 근거**다.

리드가 제목만 읽고 본문을 읽지 않았다고 밝힌 그대로이며, 그 유보 덕에 이 정정이
가능했다. 남은 진짜 질문은 §3이 연 **이관 축** 하나다.

## 1. 이관 판정 — 필요하다. 다만 이 레인 소관이 아니다

### 1.1 측정

| 측정 | 명령 | 관측 |
|---|---|---|
| develop 쪽 D7a 위치 | `grep -n 'card t315, D7a release-note duty' CHANGELOG.md` | 254행 |
| develop 쪽 구간 경계 | `grep -n '^## \[' CHANGELOG.md` | `[Unreleased]` 8행, 다음 구간 `[3.1.3]` 512행 → **254행은 `[Unreleased]` 안** |
| develop에 `[3.1.4]` 구간 존재? | 위와 동일 | **없음** |
| release 쪽 D7a 존재? | `git show origin/release/v3.1.4:CHANGELOG.md \| grep -c 'card t315, D7a release-note duty'` | **`0`** |
| release 쪽 t303 항목 존재? | `git show origin/release/v3.1.4:CHANGELOG.md \| grep -c 'SPEC-SYNC-STRATEGY-KEY-001'` | `1` |
| release 쪽 구간 경계 | `git show origin/release/v3.1.4:CHANGELOG.md \| grep -n '^## \['` | `[Unreleased]` 8행, `[3.1.4] - 2026-08-31` 10행 |
| 두 선의 조상 관계 | `git merge-base --is-ancestor origin/release/v3.1.4 develop` | **rc=1 — 조상 아님(갈라져 있음)** |
| t303 착지 위치 | `git merge-base --is-ancestor 7ed6edb3e origin/release/v3.1.4` | rc=0 |
| 릴리스 PR 상태 | `gh pr view 1685 --json state,headRefOid` | `OPEN`, head `10948d057` |

### 1.2 판정

**이관이 필요하다.** 근거는 조상 관계 한 줄이다 — `release/v3.1.4`가 develop의 조상이
아니므로, develop의 CHANGELOG에 있는 D7a 문안은 **#1685를 통해 main에 도달하지 않는다.**
#1685가 지금 상태로 머지되면 v3.1.2 → v3.1.4로 올린 사용자가 배포 기본값 변화를 예고
없이 만난다. D7a가 존재하는 이유가 바로 그 상황을 막는 것이다.

develop 쪽 배치 자체는 **옳다.** develop의 t303 항목이 `[Unreleased]`에 있으므로, 그
항목의 하위 불릿으로 붙은 D7a도 같은 구간에 있는 것이 정합적이다. 고칠 것이 없다.
문제는 develop 쪽이 틀린 것이 아니라 **release 쪽에 없는 것**이다.

### 1.3 왜 이 레인이 이관하지 않는가

세 가지가 겹친다.

1. **릴리스 브랜치 편집은 릴리스 소관.** `CLAUDE.local.md` §4.1 — main으로는 release
   브랜치의 PR만 올라가고, 카드는 develop으로만 병합한다. 레인이 `release/v3.1.4`를
   편집하면 그 체인을 우회한다.
2. **배포 판단이 딸려 온다.** 설정을 안 고친 사용자의 배포 동작이 바뀌는 것은 patch가
   아니라 minor 급이다(`d7a-behavior-delta.md` §3-2). v3.1.4로 낼지 번호를 올릴지는
   릴리스 판정이며, 문안이 들어갈 **구간 제목이 그 판정에 딸려 있다** — 판정 없이
   이관하면 틀린 구간에 넣게 된다.
3. **문안은 이미 준비돼 있다.** 이관에 필요한 것은 저작이 아니라 배치 결정뿐이다.
   붙여 넣을 형태는 `changelog-insert-block.md` 1행(그대로), 근거는 `§5`.

따라서 이 레인의 산출은 **판정 + 준비된 문안 + 리드 보고**이고, 편집은 하지 않는다.

## 2. D6 — 소진하지 않음 (유지)

카드가 못박은 정상 경로를 그대로 유지한다. 흡수 후 트리에서 좌표가 여전히 유효한지만
재측정했다.

| 측정 | 명령 | 관측 | 종전 |
|---|---|---|---|
| 배포 템플릿 센티넬 | `grep -c spec_git_workflow internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` | `1` | `1` — 불변 |
| 로컬 미러 센티넬 | `grep -c spec_git_workflow .claude/skills/moai/workflows/sync/delivery.md` | `1` | `1` — 불변 |
| 템플릿 설정의 사어 키 | `grep -c spec_git_workflow internal/template/templates/.moai/config/sections/system.yaml.tmpl` | `0` (rc=1) | `0` — 불변 |

## 3. D7a 문안의 전제 재측정 (흡수 373커밋 이후)

문안은 `delivery.md`가 지시하는 (a)/(b)/(c) 세 경로에 근거한다. 흡수로 그 경로가
바뀌었다면 문안이 거짓이 되므로 재측정했다.

| 측정 | 명령 | 관측 |
|---|---|---|
| 경로 구조 | `sed -n '230,270p' …/sync/delivery.md \| grep -n 'Strategy:\|branch state'` | `Strategy: github-flow` / `Any other branch state → stop and report` / `Strategy: git-flow` |

세 경로가 그대로다 — (a) main 무변화, (b) 비-main PR 생성, (c) 미매칭 중단. **문안은
흡수 후 트리에서도 참이다.**

## 4. 재측정 범위 — 0 패키지 (측정으로 확정)

| 측정 | 명령 | 관측 |
|---|---|---|
| 델타 파일 | `git diff --name-only develop...HEAD` | `CHANGELOG.md` + `.moai/reports/t315/` 3건 |
| 델타 중 Go 파일 | 위 \| `grep -c '\.go$'` | `0` |
| 델타 중 임베드 템플릿 | 위 \| `grep -c '^internal/template/templates/'` | `0` |
| CHANGELOG 내용을 읽는 Go 소비자 | `grep -rln CHANGELOG --include='*.go' internal/ pkg/ cmd/` → 9건 전수 판독 | **0건** — 전부 경로 문자열/제외목록/주석 |
| 누출 가드 스캔 루트 | `internal_content_leak_test.go` 판독 | `internal/template/templates/` 한정 — 루트 CHANGELOG.md는 스캔 밖 |

파일 델타가 0이라는 사실만으로 0 판정을 내리지 않았다 — `CHANGELOG`를 언급하는 Go
참조 9건을 전수 판독해 **내용을 읽는 소비자가 없음**을 확인했고, 유일하게 의심스러웠던
누출 가드는 스캔 루트가 템플릿 하위 한정임을 확인했다. 따라서 역의존 폐포도 공집합이다.

**결론: Go 재측정 대상 0 패키지.** 이 카드는 문서 축이며, 실행할 테스트가 없다는 것이
관측 결과다.

## 5. 흡수 기록

| 측정 | 관측 |
|---|---|
| 흡수 전 | `bc6ae180e`, `develop...HEAD` = `373 1` |
| 흡수 대상 | 로컬 develop `bac2cf15b` |
| 결과 | `504d5ab5a`, `MERGE_HEAD` rc=1(병합 완료), `git status --short` 0행 |
| CHANGELOG 충돌 | **없음** — 배차문이 예상한 both-add는 발생하지 않았고, D7a 문안 1건 그대로 생존 |

## Gaps

- **G1** `release/v3.1.4`를 편집하지 않았고 push도 하지 않았다(§1.3). 이관 실행은 릴리스
  소관이며 이 판정서는 그 근거만 제공한다.
- **G2** `d7a-behavior-delta.md` §G2가 그대로 열려 있다 — docs-site 4로케일에 이 동작
  변화를 실을지 재지 않았다. 이번 회차에서도 재지 않았다.
- **G3** SemVer 축(v3.1.4 patch vs 번호 상향)을 판정하지 않았다. 릴리스 판정이다.
- **G4** 사용자 프로젝트에서 upgrade를 돌려 (a)/(b)/(c)를 end-to-end로 재현하지 않았다.
  §3의 재측정은 두 tree의 `delivery.md` 원문 대조다.

## Residual-risk

- **R1** #1685가 이 판정이 리드에 닿기 전에 머지되면 이관 창이 닫힌다. 그 경우 고지는
  다음 릴리스로 밀리고, v3.1.4 사용자는 예고 없이 변화를 만난다.
- **R2** `delivery.md`는 에이전트가 읽는 지시문이지 기계 코드가 아니다(원문 R1 유지).
  §3의 경로 재측정도 같은 한계를 갖는다.
- **R3** 이 판정서는 로컬 develop `bac2cf15b` 기준이다. 창을 받을 때 develop이 더
  나아가 있으면 §5의 흡수를 다시 하고 §3·§4를 재측정해야 한다.
