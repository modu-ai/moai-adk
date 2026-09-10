# t613 통합 창 — develop 흡수 후 재측정

- 카드: t613 · 브랜치: `WT-web-host-read`
- 창: 리드 지명 후 `moai integration acquire --name t613` → 출력 `release-integration window acquired by ffea0165-c412-489c-9f67-33b62a0cb783 on WT-web-host-read`(설정 드리프트 경고 없음)
- 적용 규칙: `verification-claim-integrity.md` §2(측정 트리 귀속), CLAUDE.local.md §4.1 규율 5(병합 전 검증을 병합 후 근거로 재사용하지 않는다)

## 1. Claim

로컬 develop `304d031e5` 를 흡수한 병합 트리에서, go1.26.8 로 `./internal/web/...` 를 다시 재면 통과한다.

## 2. Evidence

흡수 전후:

```
$ git rev-parse --short develop      → 304d031e5
$ git rev-parse --short HEAD         → 971f1c70c   (흡수 전 카드 HEAD)
$ git merge --no-edit develop        → exit 0, 충돌 없음
$ git rev-parse --short HEAD         → 9afbc7aa0   (흡수 병합 커밋)
$ git status --short                 → 출력 없음
```

델타 판정:

```
$ git diff --shortstat 971f1c70c 9afbc7aa0
 406 files changed, 31273 insertions(+), 265 deletions(-)
$ git diff --name-only 971f1c70c 9afbc7aa0 -- go.mod go.sum internal/web .github
go.mod
$ git diff 971f1c70c 9afbc7aa0 -- go.mod
-go 1.26.4
+go 1.26.8
```

흡수분은 `internal/web` 을 건드리지 않는다. 다만 `go` 지시어가 바뀌어 새 도구체인으로 다시 컴파일되므로 재측정 대상이다.

도구체인:

```
$ go version
go version go1.26.8 darwin/arm64
```

재측정(직렬 1건, 원문 `.moai/reports/t613/window/web-test.txt`):

```
$ GOMAXPROCS=2 go test -p 1 -count=1 -timeout 600s ./internal/web/...
exit=0
ok  	github.com/modu-ai/moai-adk/internal/web	16.368s
```

## 3. Baseline-attribution

모든 측정은 이번 창에서, 카드 워크트리의 병합 트리 `9afbc7aa0` 에서 했다. 창 밖에서 잰 `971f1c70c` · go1.26.4 결과(`ok … 17.138s`)는 병합 근거로 쓰지 않는다.

이 문서를 싣는 증거 커밋은 `internal/web` 을 바꾸지 않는다. 따라서 위 측정은 증거 커밋 트리의 `internal/web` 에도 그대로 해당한다. develop 에 병합한 뒤에는 병합 커밋 트리와 증거 커밋 트리가 같은지 `git rev-parse <sha>^{tree}` 로 대조해 보고한다.

## 4. develop 에 들어가는 docs-site 경로

```
$ git diff --name-only develop 9afbc7aa0 -- docs-site
docs-site/content/en/advanced/moai-web-console.md
docs-site/content/ja/advanced/moai-web-console.md
docs-site/content/ko/advanced/moai-web-console.md
docs-site/content/zh/advanced/moai-web-console.md
```

## 5. Gaps

- 이번 창에서는 `./internal/web/...` 만 다시 쟀다. 흡수분 406개 파일이 속한 다른 패키지는 이 카드 범위가 아니며, 판정은 develop push 뒤 CI 몫이다.
- 창 안에서 `golangci-lint` 와 변이본은 다시 돌리지 않았다. 둘 다 창 밖 트리(go1.26.4)의 값이다.
- 실제 sshd 경유 터널과 실제 브라우저 DNS rebinding 은 여전히 재지 않았다(`run.md` §5).

## 6. Residual-risk

- 흡수분은 `internal/web` 을 바꾸지 않지만, 공유 패키지의 동작 변화가 `internal/web` 테스트에 드러나지 않을 가능성은 남는다. 이 위험은 CI 전체 실행이 판정한다.
- docs-site 네 파일이 develop 에 들어갈 때 Vercel 이 어떻게 반응하는지는 확인하지 않았다(리드가 push 시점에 확인).
