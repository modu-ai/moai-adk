# t107 — 보고서 마일스톤→카드 대조 게이트: 증거

card: t107 · worktree: `.claude/worktrees/t107` (branch WT-t107, base = origin/release/v3.1.1 @ 6b44bdd2e) · date: 2026-08-17

## Claim (주장)

1. **(3) 기계화 — writer**: t99 edges writer에 `report-milestone`(보고서→마일스톤)·`milestone-card`(마일스톤→카드) 엣지 종별 2개 추가. `.moai/reports/*.md`의 `## Card Cross-Check` 절(최초 마크다운 표)을 파싱 — card 칼럼은 **헤더 이름**(card/카드)으로 식별하고 tNN 토큰은 그 칼럼에서만 읽으므로, 본문에 다른 카드 번호가 언급돼도 오염되지 않음. 레이어 fail-open(보고서 없음 = 0엣지).
2. **(3) 기계화 — reader/CLI**: t100 reader 인터페이스를 그대로 소비. `MilestoneClaims` 순수함수(graph 패키지) + `moai graph query --milestones-no-card` 셀렉터. 카드 미주장 또는 주장 카드가 live 큐(queued/picked; dropped은 미달)에 없는 마일스톤 검출. 큐 루트는 `moai todo`와 동일한 primary-checkout 해석(`resolveTodoQueueRoot`) 재사용 — 워크트리에서 실행해도 진짜 큐를 본다. 큐 파일 부패 시 비교를 건너뛰고 그 사실을 출력(조용한 통과 아님). `BlastRadius`는 무수정으로 새 종별 소모(역방향: 카드→마일스톤→보고서).
3. **(1) 소급 대조표 작성**: unified-board-design-20260817.md용 대조표를 **실측으로** 작성 완료 — paste-ready 파일 `unified-board-card-crosscheck.md` 참조. primary 체크아웃 적용은 lane 워크트리 격리 가드가 직접 편집을 차단하여 **리드 적용으로 위임**(1회 붙여넣기).
4. **(2) 카드 요청 전 전수 대조 규율**: **적용 완료**(2026-08-17 후반, t114 머지로 예산 봉쇄 해소 후). `kanban-dispatch.md`(always-loaded stub) "Entry into the board" 절 아래 `### Report milestones ↔ queue cards` [HARD] 삽입. 동일 파일 상충 절감 5건 동반(t104 카드가 명문화한 "추가분만큼 동량 이상 절감" 규율) — 전부 t114의 stub=규범/detail=근거 분리 원칙 준수: ① Class A↔CodeRabbit 유비 괄호문(detail 중복) ② SendMessage 전체 형식(detail 존재) ③ verdict "structural, not ceremonial" 근거 → detail 신규 절 `## The verdict's home`로 이관 ④ "Branch protection is not the lever"(detail 중복) ⑤ recipe-hazard executor-blame 근거 꼬리. detail 컴패니언은 `paths:` 제한 lazy라 always-loaded 비용 0. 템플릿 미러 2본 byte-identical 동반.

## Evidence (증거)

코드 — 전부 Go, 템플릿 표면 무수정(make build·mirror-parity 불필요):

- `internal/graph/report.go` (신규, ~210줄): 종별 상수 + 파서 + 레이어
- `internal/graph/graph.go`: Build() 4번째 레이어 배선 + 패키지 문서
- `internal/graph/query.go`: `MilestoneClaim`/`MilestoneClaims`
- `internal/cli/graph.go`: `--milestones-no-card` 셀렉터 + `liveQueueCards()`(큐 seam: `todoQueueRootFn`) + 주의문구 2종
- 테스트: `internal/graph/report_test.go`(신규 6건), `query_test.go`(+2건), `internal/cli/graph_cmd_test.go`(+2건, 큐 fixture 밀봉 — 진짜 primary 큐 미참조)

검증 명령·출력 (worktree `.claude/worktrees/t107`에서):

```
$ go test ./internal/graph/ -count=1
ok  github.com/modu-ai/moai-adk/internal/graph  0.213s

$ go test ./internal/cli/ -count=1 -timeout 560s
ok  github.com/modu-ai/moai-adk/internal/cli  291.391s
(300s 타임아웃 1회 실패는 수트 벽시간 한계 — 재측정 통과)

$ go vet ./internal/graph/ ./internal/cli/   → rc=0
$ golangci-lint run internal/graph/... internal/cli/   → 0 issues
```

실물 데모 (워크트리 바이너리 `/tmp/t107-moai`, 데모 루트 = 실제 unified-board 보고서 사본 + 신규 절, 큐 = primary 진짜 큐 자동 해석):

```
$ /tmp/t107-moai graph build --root /tmp/t107demo
OK: wrote 16 edges to /tmp/t107demo/.moai/project/graph/edges.jsonl
  report-milestone: 8
  milestone-card: 8

$ /tmp/t107-moai graph query --root /tmp/t107demo --milestones-no-card
unified-board-design-20260817#S0  claimed t109 — not in live queue: t109 (done or never issued)
unified-board-design-20260817#S3  claimed t56 — not in live queue: t56 (done or never issued)
unified-board-design-20260817#S4  claimed t55 — not in live queue: t55 (done or never issued)
unified-board-design-20260817#S5  claimed t85 — not in live queue: t85 (done or never issued)
unified-board-design-20260817#S6  no card claimed ([new card needed])
unified-board-design-20260817#S7  claimed t58,t59 — not in live queue: t58 (done or never issued)
milestones without a live card: 6 of 8
NOTE: 'not in live queue' covers completed AND never-issued cards — done removes queue rows.
Resolve each flag with: git log --oneline --grep 'merge: tNN' (완결이면 통과, 미발급이면 새 카드).

$ /tmp/t107-moai graph query --root /tmp/t107demo --blast t59
blast radius of t59: 2
.moai/reports/unified-board-design-20260817.md
unified-board-design-20260817#S7
```

대조표 근거 실측:

- 큐(`moai todo`, 2026-08-17 15:1x·16:0x 재확인): live = t108·t113·t59 (queued). t55·t56·t58·t84·t85·t109 큐 부재.
- git(`git log --oneline --grep`): t109=2c70e7aed, t56=400dde787, t55=1ea829c76, t85=162f74d99, t58=b8a25b62f 머지 확인. **t84 = 0건 (미발급)** — 카드가 지적한 갭의 잔존분.
- 원 카드 근거(S0·S1 미발행, S3·S4·S7 큐 부재)는 작성 시점엔 맞았으나 이후 t108·t113 발급으로 S1·S2는 해소 — 대조표가 현재 상태를 반영.

## Baseline-attribution (baseline 귀속)

- 모든 go test·vet·lint 출력은 본 워크트리 HEAD(아래 커밋)에서 이번 실행으로 관측한 것.
- 큐/git 실측은 대조표 작성 직전 시점(2026-08-17) 관측. 큐는 가변 상태 — 적용 시점에 재측정 권장.

### 후반 이어서 진행 (release tip 동기화 + 규율 반영, 2026-08-17)

전제 변화: t114 머지(`40b0e3f27`)로 always-loaded 예산 봉쇄 해소 → 보류했던 (2) 반영 실행.

- release tip 병합: `origin/release/v3.1.1`(d169c4aec, 11커밋: t114·t59·t60·#1577·#1578 등) → merge `411158492`, 충돌 0건.
- 예산 실측(일회용 프로브 테스트, 측정 후 삭제): 반영 전 surface **75,794** tokens(여유 206 — t114 직후 703에서 t59·t60 머지로 497 감소) → 규율 삽입+절감 5건 후 **75,793**(여유 207, 순 −1토큰). `TestAlwaysLoadedTokenBudget` PASS.
- 템플릿: 미러 2본 byte-identical(복제 전 diff로 동일성 확인) + `make build`(embed 재생성, catalog.yaml 무변경) + `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/` → ok (44.932s).
- 최종 트리 재검증(병합+규율 반영 상태): `go vet ./internal/graph/ ./internal/cli/` rc=0 · `golangci-lint run internal/graph/... internal/cli/` 0 issues · `go test ./internal/graph/ -count=1` ok 0.341s · `go test ./internal/cli/ -count=1 -timeout 560s` ok **277.525s**. Go 패키지는 커밋 head와 동일(잔여 diff는 .md만).

## Gaps (미검증)

- **primary 보고서 편집 미적용** — 워크트리 격리 가드 차단. 리드가 `unified-board-card-crosscheck.md`의 블록을 삽입해야 (1)이 완결.
- darwin/windows 크로스컴파일 게이트는 CI 몫(로컬 미실행 — 경로 처리에 filepath.Join/FromSlash만 사용).
- rule-authoring.md (b) 고시 의무: stub 순증가 없음(절감이 상쇄, bytes 기준 순 감소)이라 미발동.

## Residual-risk (잔여 위험)

- 대조표 없는 보고서는 이 검출기에 **보이지 않음**(쿼리가 "no Card Cross-Check sections found"로 알림). 규율 (1)의 보급률에 좌우.
- "not in live queue"는 완결/미발급을 구분 못 함(backlog.json이 done 행을 삭제) — 설계상 인간이 git으로 판정. 주의문구가 절차를 안내.
- 다른 보고서가 같은 마일스톤 id(S3 등)를 쓰면 stem 한정자(`보고서명#S3`)로 충돌 없음 — 테스트로 확인.
- 다른 세션의 동시 큐 변경은 읽는 순간 스냅샷과 달라질 수 있음(게이트는 어디까지나 요청 시점 점검).
- stub 절감 5건은 전부 근거 이관/중복 제거지만, 리뷰어가 절감 문맥(특히 verdict 근거의 detail 이관)을 다르게 볼 수 있음 — 절감 내역은 커밋 메시지에 명시.
