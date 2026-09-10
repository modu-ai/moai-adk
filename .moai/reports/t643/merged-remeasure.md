# t643 — 흡수 후 재측정 (lane-8)

- 창: `moai integration acquire --name lane-8` → `release-integration window acquired by 7e33c4bd-a91e-4a67-bf0c-15ad255da456 on WT-toolchain-doc-fixes`
- 흡수 전: 로컬 develop `8203040b8`, 카드 HEAD `db0016e90`, `git status --short` 출력 없음
- 흡수: `git merge --no-edit --no-stat develop` → `Merge made by the 'ort' strategy.` (충돌 없음)
- 병합 트리: `git rev-parse HEAD 'HEAD^{tree}'` → `920f6cf38918afe34683eef580d500ac5a9d8d63` / `de52aa237b2e88b5aa580115907adb3c82f47407`
- `git merge-base develop HEAD` → `8203040b857b377975a6a32156deb978fa4232bd`

## 카드 변경 범위

```text
$ git diff --stat develop...HEAD
 .moai/project/codemaps/modules.md     |  2 ++
 .moai/project/codemaps/overview.md    |  4 ++-
 .moai/reports/t643/repro-siblings.txt |  8 ++++++
 .moai/reports/t643/repro-traces.txt   | 17 ++++++++++++
 .moai/reports/t643/repro.md           | 51 +++++++++++++++++++++++++++++++++++
 .moai/reports/t643/verdict.md         | 51 +++++++++++++++++++++++++++++++++++
 CHANGELOG.md                          |  2 +-
 7 files changed, 133 insertions(+), 2 deletions(-)
```

흡수 전 `git diff --stat 296ba7aa5 HEAD`의 7개 파일과 같다. 흡수한 develop은 이 파일들을 건드리지 않았다.

## 문서 3개 재측정

```text
old_phrase=0          # grep -c -F 'reachable from the `go.mod`' CHANGELOG.md          (정정 전 1)
new_phrase=1          # grep -c -F "each with a call path from this module's code to the affected standard-library symbol" CHANGELOG.md
overview_41f445fa5=1  # grep -c -F '41f445fa5' .moai/project/codemaps/overview.md
modules_41f445fa5=1   # grep -c -F '41f445fa5' .moai/project/codemaps/modules.md
```

리드가 제시한 기대값(옛 문구 0, 새 문구 1, 인용 문서별 1)과 일치한다. 코드 변경이 0이라 빌드와 테스트는 돌리지 않았다.
