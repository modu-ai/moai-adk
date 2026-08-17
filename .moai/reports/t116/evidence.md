# t116 — 배치 PR 선결 CI 붉음 2건 해소: 증거

card: t116 · worktree: `.claude/worktrees/t116` (branch WT-t116 = origin/main 931e4138a + origin/release/v3.1.1 @ 091a42e16 merge) · date: 2026-08-17 · head: 265bcba0a

## Claim (주장)

1. **① TestHomeJoinSiteCountIsPinned 5≠4 해소**: count pin을 **명시적 allowlist 집합 동등성**으로 전환(`homeJoinSiteAllowlist`, 5사이트 — tokens.go 포함). 동기: 값 count는 "뭔가 움직였다"만 알리고 목록이 코드를 따라갔는지는 못 보므로, release에서 tokens.go가 추가된 뒤에도 pin=4인 채 확정 붉음으로 방치됐다(t97 run-evidence가 별도 카드 권장 후 미발령 상태였음). 집합 동등성은 **양방향 동기화 누수**에 붉어진다 — 코드에 사이트 추가/삭제/이동 시 allowlist 미갱신 → 붉음, allowlist 잔류 행 → 붉음. 경로는 `filepath.ToSlash` 정규화(CI darwin/windows 매트릭스 대응).
2. **② TestConsumerOnly_M0AndMxByteUnchanged 배치 PR 확정 붉음 해소 — 비교 기준 조정 채택**: diff baseline을 "HEAD의 가장 가까운 **리뷰 경계**"로 선택. `origin/main`이 기본(3-dot, verbatim AC 형태 그대로)이고, **push된 `origin/release/*` tip이 HEAD의 조상이며 main merge-base보다 엄격히 가까울 때만** 그 tip이 baseline이 된다. 동점은 origin/main 유지(최대 강도 선택). 근거: push된 release tip은 리뷰 경계다 — release/*에 도착하는 모든 통합은 review-PASS를 동반한 레인 머지다. 핀의 관할은 경계 **사이**의 변경분이지, 이미 경계를 넘은 리뷰 누적분이 아니다. 이 조정 없으면 배치 PR(release→main)은 물론 release tip을 머지한 **모든 레인 브랜치**에서 확정 붉음(557877c49 internal/mx/spec_loader — t99 그래프 writer, `git branch -r --contains`로 origin/release/v3.1.1 도착 확인).
3. **일반 PR 핀 강도 보존 — 실측 입증**: pure 정책 함수 `chooseConsumerOnlyBaseline` 추출 + 5케이스 테이블 테스트(무 release ref / 비조상 무시 / 조상+근접 채택 / 동점 main 유지 / 다수 중 최근접). 정책은 release 조상이 없는 평범한 피처 브랜치에서 **원본과 byte-identical** 경로(origin/main 3-dot)만 탄다.

## Evidence (증거)

재현 (수정 전, 본 워크트리):

- ① `--- FAIL: TestHomeJoinSiteCountIsPinned — found 5 file(s) ... want 4: [internal/cli/memory.go internal/cli/migrate_profiles.go internal/cli/preference/cmd.go internal/cli/tokens.go internal/hook/session_end.go]`
- ② `--- FAIL: TestConsumerOnly_M0AndMxByteUnchanged — AC-NS2-005a FAIL: M1 diff touches mx producer path "internal/mx/spec_loader.go" / "internal/mx/spec_loader_test.go"` · 원인 커밋 557877c49(feat(graph): edges.jsonl writer)은 origin/release/v3.1.1 소속.

수정 후 (head 265bcba0a):

```
$ go test ./internal/hook/ -run 'TestConsumerOnly_M0AndMxByteUnchanged|TestHomeJoinSiteCountIsPinned|TestChooseConsumerOnlyBaseline' -count=1
ok  github.com/modu-ai/moai-adk/internal/hook  2.559s   (전건 PASS, 정책 5 서브케이스 포함)

$ go test ./internal/hook/ -count=1 -timeout 300s
ok  github.com/modu-ai/moai-adk/internal/hook  22.393s

$ go vet ./internal/hook/ → rc=0 · $ golangci-lint run internal/hook/... → 0 issues
```

**배치 PR 조건 시뮬레이션** (리드 지시 재현): WT-t116 = origin/main + release tip merge 형태로 **배치 PR merge-ref(main×release 병합)와 동일 구조**.

- 구가드 관점 확정 붉음 입력 실측: `git diff --name-only origin/main...HEAD | grep -E '^internal/(navigator/sync|mx)/'` → `internal/mx/spec_loader.go` · `internal/mx/spec_loader_test.go` (rev-list `origin/main...HEAD` = 171커밋).
- 신가드: baseline = `refs/remotes/origin/release/v3.1.1`(091a42e16) 선택, `git diff --name-only 091a42e16..HEAD` = 정확히 본 카드의 2 테스트 파일 → 가드 PASS.

**핀 강도 보존 프로브** (경계 위 신규 mx 변경은 여전히 잡히는지): `internal/mx/spec_loader.go`에 주석 1행 추가 커밋(임시) 후 가드 실행 →

```
--- FAIL: TestConsumerOnly_M0AndMxByteUnchanged
    AC-NS2-005a FAIL: diff vs refs/remotes/origin/release/v3.1.1 touches mx producer
    path "internal/mx/spec_loader.go" — the Detect layer must consume internal/mx/ read-only
```

→ 리뷰 경계(release tip) 위의 레인 수준 mx 변경은 **여전히 붉게 잡힘** 입증. 프로브는 `git reset --soft HEAD~1` + `git restore`(rm --hard는 권한 거부로 비파괴 경로 사용)로 폐기 — 최종 트리에 잔존 없음(`git status` clean, HEAD=265bcba0a).

## Baseline-attribution (baseline 귀속)

- 모든 테스트·vet·lint·git 출력은 본 워크트리 HEAD 265bcba0a에서 이번 실행으로 관측.
- ①의 5사이트 목록은 가드 실패 메시지가 실측 나열한 것(추측 아님).

## Gaps (미검증)

- 배치 PR의 실제 CI 런(main향 PR 생성 후)은 본 카드 범위 밖 — 배치 PR이 최종 확인.
- "평범한 피처 브랜치에서 원본과 byte-identical"은 정책 단위테스트(비조상 → main)로 입증했으나, release 미포함 실제 브랜치에서의 라이브 실행은 별도 워크트리를 파지 않아 미실시(코드 경로 동일 — baseline 후보가 전부 distance=-1이면 main 3-dot).
- 리뷰 경계의 전제("release/* 도착분은 전부 리뷰 동반")는 칸반 규율의 운영 사실이지 기계 검증이 아님 — 직접 커밋이 release에 들어오면 이 가드의 관할 밖(배치 PR 시점엔 그대로 diff에 잡힘: baseline이 tip이므로 tip 이후 커밋은 전부 잡힘).

## Residual-risk (잔여 위험)

- `origin/release/*` ref가 아주 오래된 tip을 가리키면(레인이 로컬에서 release에 머지했으나 미푸시) baseline=마지막 푸시 tip → 미푸시 머지분 전체가 diff에 잡힘 — 의도된 보수적 거동(미푸시분은 아직 '리뷰 경계 넘음'으로 간주하지 않음).
- allowlist는 파일 경로 기준 — 같은 파일 안에서 site 개수가 늘어나는 경우(1파일 2 join)는 감지 못함(기존 count pin도 동일 한계, needle이 파일 단위).
- ①의 `wantSites`라는 변수명은 제거했으나 테스트 함수명 `TestHomeJoinSiteCountIsPinned`는 유지(2개 파일 주석 참조 + t97 보고서 근거 — 개명은 참조 3곳 동반 필요로 스코프 초과).
