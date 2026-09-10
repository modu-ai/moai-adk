# t264 — Stage A 실행 결과 (운영자 승인 2026-09-02, Stage A 한정)

- 실행 세션: `WT-stale-branch-sweep` (`.claude/worktrees/t264`) · 기준선 `origin/develop` = `ad272be20` (실행 전후 동일 확인)

## 1. Claim (주장)

승인 범위인 Stage A를 전수 시행했다: 병합 확인 워크트리 142건 중 100건 폐기, 대상 브랜치 214건 중 210건 삭제. 보호 브랜치(main·develop·release/*)와 HARD 11트리에는 **어떤 쓰기도 발생하지 않았다**. 거부된 항목은 전부 git 네이티브 안전망이 자체 거부한 것이며 `--force`는 1건도 사용하지 않았다.

## 2. Evidence (증거)

| 항목 | 값 | 근거 |
|---|---|---|
| 실행 전 브랜치 | 455 (병합 263 + 미병합 192) | `data/r2-merged.txt`·`data/r2-unmerged.txt` |
| 실행 후 브랜치 | **245** (병합 56 + 미병합 189) | `git for-each-ref` 최종 재측정 |
| 브랜치 삭제 | **210** (시도 214 − 거부 3) | `data/del-b4.log`~`del-b11.log` |
| 실행 전 트리 등록 | 187 | `data/r2-worktrees-porcelain.txt` |
| 실행 후 트리 등록 | **84** | `data/r4-porcelain.txt` |
| 트리 폐기 | **100** (manifest 142 − 거부 42) | 배치 1~12 각 `git worktree list \| wc -l` |
| pilot 가드 판정 | 워크트리 세션에서 `git worktree remove <타경로>` 허용 | t139 pilot (배치 1 전) |

정합: 455 − 210 = 245 ✓ · 187 − 100 − 3(병렬, 아래) = 84 ✓

## 3. 거부 항목 (전부 skip + 보고, `--force` 미사용)

**트리 42건** — `contains modified or untracked files` 거부: t349·t428·agent-ac385069535544f99·t303·t347·t281·t140·t362·agent-a8b0420c635295c5a·t355·t253·t182·t215·t404·t373·agent-a62468d0d1a7040cf·card-relnotes·t62·t379·t166·agent-ad55b5fbe632611a7·agent-af39f39d9c430af36·t46-anchor-guard·t74·agent-a20f1e50d08875998·t351·t233·t279·t426·t289·t312·t80·t143·t155·t148·t334·t335·agent-a77cdfb3c00bf2557·t340·t350·agent-a2e81b36e7e5646c7·t150. 이 트리들의 브랜치도 삭제에서 제외됨(최종 목록 214 = 256 − 42).

**브랜치 3건** — `git branch -d` 자체 거부: `WT-version-stamp-predicate`(HEAD에는 병합됐으나 upstream `origin/WT-version-stamp-predicate`와 tip 불일치)·`WT-integration-lock`·`WT-web-sigterm-toctou`(미병합 판정). 전부 안전망의 판정을 존중해 보존.

## 4. 병렬 행위자 관측

manifest 밖에서 **t430·t431·t433** 3트리가 실행 도중 사라졌다 — 전부 HARD 목록 트리로, 본 세션은 건드리지 않았다(리드 병합 창 쪽 폐기로 추정). lsof 재측정에서 t430 점유가 사라진 것과 시의성이 일치. 브랜치 집계와 정합(위 정합식).

## 5. Gaps (미관측)

- 거부 42트리의 파일이 **untracked인지 modified인지 미분류** — 워크트리 세션 가드가 타 트리 `git -C`를 거부. untracked-only(증거 파일 등, primary에 이미 반출된 것)면 `--force` 후보지만 본 단계에서 판정하지 않음.
- 거부 3브랜치의 upstream divergence 정체(원격에만 있는 커밋인지)는 미판독.
- lsof 점유 프로세스(t336 고아 디렉터리 등)의 신원은 Stage D 소관으로 미처리.

## 6. Residual-risk

- 거부 42트리 안의 미측정 파일은 유일본일 수 있다 — 트리를 지우려면 untracked-only 입증이 먼저다.
- 3거부 브랜치는 tip이 움직였거나 원격 쌍둥이와 갈린 상태 — 원인 불명으로 남겨 둠(측정 시점 병합 판정과 현재 상태의 불일치).
- Stage A 이후 미병합 189건이 남아 있다 — Stage B(축 2 저작경로 검증)는 아직 미승인·미실행.

## 7. B·C 판별식에 대한 A 실행의 영향 (리드 요청)

- **Stage B**: 판별식 자체는 불변(축 2 저작경로 공집합). 다만 A가 두 가지 실측을 추가했다: (a) 거부 42트리의 브랜치가 B 후보군에 추가되며, B 실행 시에도 동일한 더티-거부가 트리 단계에서 재발할 것 — B는 트리 처리 전략(untracked-only 입증 절차)을 포함해야 한다. (b) `-d` upstream 거부 3건은 "로컬-원격 쌍둥이 tip 불일치" 클래스의 존재를 증명 — B 축 2 판정에 **원격 쌍둥이 존재·상태 확인**을 한 축 추가할 것을 권고.
- **Stage C**: 원격 11 후보 중 8건(`WT-card-landing-state`·`WT-codex-launcher`·`WT-freshness-sync`·`WT-gate-three-axes`·`WT-ossdocs-v311`·`WT-t69`·`WT-wscfg-graph`·`WT-wscfg-worktree`)의 로컬 쌍둥이가 A에서 삭제됐다 — 내용은 전부 develop 병합 확인분이라 무손실이지만, 원격 삭제 전 로컬 백업은 이제 없다. `WT-version-stamp-predicate`는 로컬 생존(거부)이며 원격 tip이 develop 병합 상태라 원격 삭제는 무손실. stale tracking refs 66건은 그대로(`fetch --prune`은 C 소관).
