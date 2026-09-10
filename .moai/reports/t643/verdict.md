# t643 판정 기록 — t610 sync 감사 보류 2건(F1·F7) 정정

- 카드: t643 (Class B, 문서 정정, plan 생략). 근거: `.moai/reports/t610/sync-audit.md`의 F1(should-fix)·F7(advisory)
- 워크트리: `.claude/worktrees/t643`, 브랜치 `WT-toolchain-doc-fixes`
- 기반: `origin/develop` `d060e0d13`에서 생성한 뒤 `git merge --no-ff 296ba7aa5`로 정렬했다(병합 커밋 `8addcf742`, `HEAD^2` = `296ba7aa5682da20ef906dde6b2764431442b348`).
- 작성: lane-8, 2026-09-10. 병합 판정은 리드의 몫이다.

## 주장

1. **F1:** codemaps 두 문서(`overview.md`, `modules.md`)의 머리말이 **Go** 버전 한 값을 측정 트리(t592 `e7bd89ee3`)에서 잰 값처럼 보이게 했다. 이제 그 값의 실제 출처(t610 커밋 `41f445fa5`의 `go.mod:3`)를 밝힌다. 다른 수치와 측정 트리 줄은 바꾸지 않았다.
2. **F7:** `CHANGELOG.md:12`의 "reachable from the `go.mod` build directive" 문구를 govulncheck가 실제로 판정하는 기준인 코드 호출 경로로 바꿨다. 새 문구는 "each with a call path from this module's code to the affected standard-library symbol"이다.

## 재현 (정정 전, 트리 `8addcf742`)

증거: `repro.md`, `repro-traces.txt`, `repro-siblings.txt`

| 항목 | 명령 | 관측 출력 | 판정 |
|---|---|---|---|
| F1 측정 트리 값 | `git show e7bd89ee3:go.mod` | 3행 `go 1.26.4` (문서에는 `**Go**: 1.26.8`) | 재현됨 |
| F1 머리말 | `overview.md` 3-4행 읽기 | "모든 수치는 아래 트리에서 직접 잰 것이고, 다른 트리·다른 시점에서 옮겨온 값은 없습니다." | 재현됨 |
| F7 문구 | `grep -n -o 'reachable from the[^;]*;' CHANGELOG.md` | `12:reachable from the \`go.mod\` build directive at go1.26.4 — ...` | 재현됨 |
| F7 판정 기준 | `grep -c 'Example traces found' .moai/reports/t610/baseline/govulncheck-auto.log` | `8` (예: `internal/update/checker.go:76:26: update.checker.CheckLatest calls http.Client.Do, which eventually calls url.URL.Parse`) | 도달성은 호출 경로로 판정됨 |
| F7 형제 문구 | `grep -rn -l -F 'reachable from the \`go.mod\`' --include='*.md' --exclude-dir=node_modules --exclude-dir=reports .` | `CHANGELOG.md` 하나 | 살아 있는 사본은 1곳 |

## 정정 뒤 측정 (작업 트리, 커밋 전)

| 항목 | 명령 | 관측 출력 | 기대 |
|---|---|---|---|
| F7 옛 문구 | `grep -c -F 'reachable from the \`go.mod\`' CHANGELOG.md` | `0` (정정 전 `1`) | 0 |
| F7 새 문구 | `grep -c -F "each with a call path from this module's code to the affected standard-library symbol" CHANGELOG.md` | `1` | 1 |
| F1 출처 인용 | `grep -c -F '41f445fa5'` on `overview.md` / `modules.md` | `1` / `1` | 각 1 |
| 인용 커밋이 조상인지 | `git merge-base --is-ancestor 41f445fa5 HEAD` | 오류 없음(exit 0) | 조상 |
| 인용 커밋 내용 | `git show --format='%h %s' 41f445fa5 -- go.mod` | `41f445fa5 fix(SPEC-GO-TOOLCHAIN-SEC-002): M1 raise go directive to 1.26.8 (t610)`, 헌크 `-go 1.26.4` / `+go 1.26.8` | go.mod:3 변경 커밋 |
| 인코딩 | `\u` 이스케이프 grep / 보이지 않는 문자 스캔 | 0건(grep exit 1) / 0건 | 0 |

바뀐 파일은 `CHANGELOG.md` 1줄, `overview.md`(1줄 수정, 2줄 추가), `modules.md`(2줄 추가)다. 최종 문안은 커밋 diff에서 볼 수 있다.

## 기준 귀속

모든 측정은 2026-09-10에 레인이 이 워크트리에서 직접 했다. 정정 전 기준은 기반 정렬 병합 커밋 `8addcf742`의 트리다.

## 확인하지 못한 것

- 코드·테스트는 바꾸지 않았다. 문서 정정이라 빌드와 테스트는 돌리지 않았다.
- codemaps를 새 트리에서 다시 생성하는 방안(감사가 제시한 대안)은 택하지 않았다. 다른 수치까지 다시 재야 해서 이 카드 범위를 넘는다.
- `.moai/reports/` 아래 과거 증거 파일에 남은 옛 문구는 기록물이라 고치지 않았다(`repro-siblings.txt` 범위 설명 참조).

## 잔여 위험

- **codemaps 재생성 시:** 다음에 codemaps를 다시 생성하면 머리말이 새 측정 트리 기준으로 바뀐다. 그때 이번 예외 문장은 필요 없어지므로 함께 지워야 한다.
- **F7 문구의 한계:** 새 문구는 govulncheck의 기준선 로그에 호출 경로 블록이 8개 있다는 사실에 기댄다. 각 경로가 실제로 실행되는지는 공격 재현이 아니므로 확인하지 않았다(t610과 같은 한계).
