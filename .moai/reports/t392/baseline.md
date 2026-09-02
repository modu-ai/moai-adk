# t392 baseline — 버전 스탬프 판별식

측정 트리: worktree `.claude/worktrees/t392`, HEAD `9a3e2dabe` (= `origin/develop`, t388 착지본)
측정 시각: 2026-09-01

## B.1 카드 술어 재측정 (카드 base `9328a5242` → 현재 base)

    git grep -nE 'v[0-9]+\.[0-9]+\.[0-9]+' -- . ':!.moai/reports' ':!.moai/specs' \
      ':!.moai/release-notes' ':!CHANGELOG.md' ':!*_test.go' ':!docs-site/content/*/changelog*'

| 값 | 카드 기재 (`9328a5242`) | 재측정 (`9a3e2dabe`) |
|---|---|---|
| 줄 | 2225 | 2229 |
| 파일 | 592 | 595 |

카드 수치는 근거로 쓰지 않는다. 아래 판단은 전부 `9a3e2dabe` 측정이다.

## B.2 술어를 「권위 버전 토큰」으로 좁혔을 때의 면적

권위 버전: `pkg/version/version.go:8` → `v3.1.3` (트리 값. 최고 태그는 `v3.1.2` — 결번 상태)

    git grep -lF v3.1.3 -- . <같은 거부 목록>   →  34 파일

전체 트리 기준으로는 155 파일이며, 거부 목록 적용 후 34.
595 → 34. 카드가 Tier M 부담의 근거로 삼은 면적은 **모든 `vX.Y.Z` 토큰**을 술어로 잡았을 때의
값이고, 누락 탐지에 필요한 술어는 그것이 아니다.

### 34 파일의 분류

| 부류 | 수 | 내용 |
|---|---|---|
| 등록된 스탬프 | 7 | README×4 · `.moai/config/sections/system.yaml` · `docs-site/hugo.toml` · `pkg/version/version.go` |
| **미등록 스탬프** | **6** | `internal/cli/testdata/{status,doctor}-{dark,light,nocolor}.golden` |
| docs-site 서술 | 20 | 5문서 × 4로케일 (`faq` · `statusline` · `claude-cloud` · `codex-dual-harness` · `moai-feedback`) — 예시 출력 인용 |
| 목록 문서 자신 | 1 | `.moai/docs/version-management.md:90` (t388이 남긴 「닫지 못한다」 서술) |

## B.3 [핵심] 미등록 스탬프 사이트가 지금 살아 있다

golden 픽스처 6개는 `version.Version` 을 고정하지 않고 실행 시점 값을 그대로 찍는다
(`internal/cli/status_golden_test.go` · `doctor_golden_test.go` 에 pin 없음 — 대조:
`version_test.go:180` 은 `v0.0.0-test` 로 pin 한다). 따라서 범프가 이 파일을 다시 찍지 않으면
`go test ./internal/cli/...` 가 깨진다. **즉 스탬프 사이트다.** 그런데 t388 목록에 없다.

이력이 재발을 보인다:

    b37e86b64  test(cli): stamp doctor/status golden fixtures at v3.1.3   ← 사후 수리
    44b4c3c1e  chore: bump version to v3.1.1                              ← 그 전 갱신
    61921f1ba  chore: bump version to v3.1.4  (7파일 9줄, golden 없음)

**현재 `origin/release/v3.1.4` 에서 실제로 어긋나 있다:**

    git show origin/release/v3.1.4:pkg/version/version.go                  → Version = "v3.1.4"
    git show origin/release/v3.1.4:internal/cli/testdata/status-nocolor.golden → moai-adk v3.1.3

t388 §1 의 `hugo.toml`(v3.1.3) 사고와 같은 모양이 **다른 사이트에서 반복 중**이다.
이 카드가 닫으려는 방향이 가설이 아니라 관측이라는 뜻이다.

Gap: 「그 트리에서 golden 테스트가 실제로 FAIL 한다」는 아직 실행으로 관측하지 않았다.
코드 판독(pin 부재 + 리터럴)과 값 어긋남까지만 관측했다.
→ **닫힘**: `red-observation.md` §R.2-R.3 (6/6 PASS → 버전 주입 시 6/6 FAIL).

[정정 2026-09-01] 이 절의 초판은 `origin/release/v3.1.4` 를 「도달 가능한 RED 후보」라고
적었다. **틀렸다.** 두 가지가 겹쳐 읽혔다.

- 그 트리의 권위 토큰은 `v3.1.4` 이고 낡은 golden 은 `v3.1.3` 이다. 권위 토큰 술어는 그
  golden 을 **스치지 않는다** — 드리프트는 있는데 새 검사는 조용하다. RED 이 아니다.
  (이것이 SPEC 이 2×2 의 네 번째 사분면 「미등록 AND 낡음」을 어느 단언도 닫지 못한다고
  이름 붙인 자리다.)
- 거기서 관측된 CI 적색은 **golden 테스트 자신의 실패**이지 새 검사의 신호가 아니다.
- 도달 불가 문제도 실제로는 재발한다: `git merge-base --is-ancestor 26898312e HEAD` → rc=1.
  이 카드가 그 트리를 쓰지 않기 때문에 물리지 않을 뿐이다.

새 검사의 RED 은 **이 트리, pin 이전**에 있다 — 여기서는 golden 이 권위 토큰(`v3.1.3`)을
그대로 담고 있어 술어에 걸리고, 등록부에 없으므로 실패한다.

## B.4 두 방향은 서로 다른 검사를 요구한다

| 방향 | 사고 사례 | 술어 |
|---|---|---|
| 등록된 스탬프가 낡음 | `hugo.toml` v3.1.3 | 목록의 7경로가 **전부 권위 토큰을 담는가** |
| 미등록 사이트 존재 | golden 6개 (진행 중) | 권위 토큰을 담은 파일이 **전부 분류되어 있는가** |

첫째는 목록만 읽으면 되고 값이 싸다. 둘째가 (가)/(나) 갈림길이다.

## B.5 등록부 유지비 추정

    v3.1.2 를 담은 파일 (같은 거부 목록) → 8, 그 8개 전부가 v3.1.3 집합에도 포함
    v3.1.2 에만 있는 파일 → 0

직전 버전 토큰의 잔존 면적이 작다. 릴리스마다 등록부가 통째로 흔들리지는 않는다.

## Gaps (관측하지 않은 것)

- golden 테스트의 실제 FAIL 실행 (B.3)
- docs-site 20파일이 「범프가 건드리지 않는다」는 점의 커밋 근거 — 성격상 서술로 판단했을 뿐
  numstat 으로 확인하지 않았다
- `.moai/release/` 디렉터리(t388 §2 가 경고한 한 글자 차이)는 권위 토큰 술어에서는 34 안에
  나타나지 않았다 — 과거 버전만 담기 때문. 술어를 계보 전체로 넓히면 다시 걸린다
