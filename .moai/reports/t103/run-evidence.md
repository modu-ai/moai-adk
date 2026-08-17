# t103 — statusline 세그먼트 표기 변경 + 줄 순서 변경 — 실행 증거

Worktree: `.claude/worktrees/t103` · Branch: `WT-t103` · Base: `0ede5db6a` (origin/release/v3.1.1, 작성 시점 tip — 디스패치 지정 `c2a520cc9`보다 3커밋 앞선 t106 머지 포함 판. 사이 커밋(t106)은 statusline 표면과 무교집합)

## 1. Claim (주장)

디스패치 5건의 처치:

1. **`🔄 TODO: 4/20` 표기 (이행)** — `renderSessionLine`의 백로그 세그먼트가 `🔄 %d / ⤵️ %d`에서 `🔄 TODO: %d/%d`로 변경됐다. **두 숫자 의미 확정: 첫 숫자 = picked(진행 중), 둘째 = queued(대기 큐)** — 기존 `data.Backlog.Picked, data.Backlog.Queued` 순서 그대로. 구 설계 근거 주석("Symbols rather than words because this binary ships to users of every supported language")은 운영자 지시(2026-08-17)로 단어가 들어오며 폐기됨을 명시하는 근거로 교체 — TODO는 16개 지원 언어에서 동일하게 읽히는 유일한 라벨이라 중립성 훼손이 최소라는 근거 동반.
2. **`🔀 %d / %d` 단일 아이콘화 (이행)** — GitHub 세그먼트가 `⚠️ %d / 🔀 %d`에서 `🔀 %d / %d`로. 순서 유지: 첫 숫자 = open issues, 둘째 = open PRs. 주석에 운영자 요청 근거와 함께 기록.
3. **repo 🔀→📡 잔여 확인 (확인 완료, 변경 불필요)** — `caf435ec4`("swap segment icons — issues ⚠️, PRs 🔀, repo 📡")에서 이미 전환됨. statusline 비테스트 코드의 🔀는 전부 PR 아이콘 문맥(renderer.go:179-183, 501 — "distinct from the repo indicator, which renders 📡"). 전체 스윕 `grep -rn "🔀.*owner|owner.*🔀|🔀 %s/%s"` → 0건.
4. **구분자 │→| 변경 의도 확인 (확인 완료, 그런 변경 없음)** — `joinSegments`의 separator는 `Renderer` 초기화의 `" │ "`(U+2502, renderer.go:30 "v3 separator: U+2502 box drawing vertical line")가 현행이며 `caf435ec4`는 separator를 건드리지 않음(해당 커밋 renderer.go diff에 separator 0회). **│→| 변경은 존재하지 않으므로 확인할 의도가 없음** — 보고로 종결. 구분자는 │(U+2502) 유지.
5. **세션줄 마지막 줄로 이동 (이행, 운영자 지시 2026-08-17)** — `renderDefaultV3`에서 세션줄(🏷️ 신원 │ 🔄 백로그 │ GitHub)을 L0(첫 줄)에서 **조건부 마지막 줄**로 이동. 줄 배치: L1 정보 → L2 바 → L3 프로젝트(📁/🅱️) → 마지막 세션줄. 신원·백로그 없는 세션은 종래대로 3줄 레이아웃 유지. 배치를 고정하는 회귀 테스트 추가(`TestRenderDefaultV3_SessionLineClosesTheStatusline` — 마지막 줄에 세션 신원, 이전 줄에 🏷️/🔄 TODO: 부재 검증).

## 2. Evidence (증거 — 명령 + 출력)

```
$ go vet ./internal/statusline/
(출력 없음 — 통과)

$ go test ./internal/statusline/
ok  github.com/modu-ai/moai-adk/internal/statusline  15.041s
  # 기존 전부 + 신규 TestRenderDefaultV3_SessionLineClosesTheStatusline 통과

$ grep -rn "🔀.*owner|owner.*🔀|🔀 %s/%s" internal/ (비테스트)
0건 — repo 아이콘 오용 잔여 없음

$ golangci-lint run ./internal/statusline/...
0 issues.
```

테스트 기대값 갱신: `⚠️ 7 / 🔀 3` → `🔀 7 / 3` (github_test 2곳), `🔄 12 / ⤵️ 26` → `🔄 TODO: 12/26` (github_test 1곳 + session_identity_test 1곳). 에러 메시지의 아이콘 서술도 갱신.

## 3. Baseline-attribution (baseline 귀속)

모든 측정은 본 워크트리(`.claude/worktrees/t103`, branch `WT-t103`)에서, base `0ede5db6a`에 본 카드 diff 3파일(renderer.go, github_test.go, session_identity_test.go)를 적용한 트리에 대해 이번 실행으로 관측했다.

## 4. Gaps (미검증)

- **렌더링 실물 화면 확인 미실행** — statusline은 Claude Code가 표시하므로 터미널 실물 스크린샷 검증은 운영자 화면에서만 가능. 단위 테스트가 세그먼트 문자열·줄 순서를 핀.
- **전체 스위트 미실행** — 로컬 부하 규율(레인 타깃 패키지만). statusline 패키지 외 의존처 없음(순수 표기 변경)이지만 CI가 전수 판정.

## 5. Residual-risk (잔여 위험)

- **표기 변경은 사용자 가시적** — `🔄 TODO:` / `🔀` 단일 아이콘은 기존 사용자에게 익숙한 표기 교체. 운영자 지시 사항이므로 의도된 변경.
- **세그먼트 비활성 검사의 ⚠️ 잔여** — github_test의 disabled 케이스가 여전히 `⚠️` 부재도 검사(이제 렌더에 ⚠️가 없어 항상 참인 조건). 오류는 아니나 다음 정리 때 🔀 단일 조건으로 좁힐 여지.
