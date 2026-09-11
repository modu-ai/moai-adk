# t633 통합 창 기록

창 보유: lane-8. 리드 지명(lane-3 반납 뒤) → status free 확인 → `moai integration acquire --name lane-8` exit 0.
흡수 전 사전 예측에서 `internal/template/catalog.yaml` 충돌이 보여 창을 잡기 전에 멈추고 보고했고, 리드가 해결안을 승인했다.

## 흡수

- 흡수 대상: 로컬 develop `8a44505a1d6153c9591e95b54238a0a30c350adb` (= origin/develop, 리드 확인). 흡수 전 카드 HEAD `7595338b9`.
- 사전 예측: `git merge-tree --write-tree --name-only 8a44505a1 HEAD` → `CONFLICT (content): Merge conflict in internal/template/catalog.yaml` 1건. `sync/quality-gates-context.md` 는 로컬·템플릿 모두 자동 병합.
- 실행: `git merge --no-ff develop` → exit 1, 같은 1건(`absorb-merge.log`). 충돌 표지 9/11/13행(`catalog-conflict-markers.txt`, `catalog-conflict.diff`) — t633 과 t652 가 같은 `moai` 해시 줄을 바꿨다.

## catalog.yaml 해결 — 생성물이라 재생성

1. 충돌 표지만 없애도록 develop 쪽(stage 3)을 가져왔다(`catalog-stage3-develop.yaml`; 카드 쪽 stage 2 는 `catalog-stage2-card.yaml`). 표지 0.
2. 병합된 템플릿 트리에서 `go run ./internal/template/scripts/gen-catalog-hashes.go --all --dry-run` → exit 0, 45개 엔트리 계산값(`catalog-dryrun-before.txt`).
3. `--all` 로 재생성 → exit 0(`catalog-regen.log`). develop 쪽 catalog 대비 바뀐 해시는 3개(`catalog-vs-develop.diff`):
   - `moai`: `55e4d1df…` → `c4cf9e70…`
   - `moai-workflow-loop`: `1268c2c6…` → `de4377e1…`
   - `moai-domain-html-report`: `7e7e427f…` → `60ca28ac…`
   셋 다 t633 이 템플릿 본문을 바꾼 스킬이다. **두 카드(t652·t633)가 건드린 엔트리 밖에서 바뀐 스테일 해시는 없다.** t652 의 엔트리(`manager-git`·`moai-workflow-worktree`)는 develop 에 이미 옳은 값이 있어 변화가 없었다.
4. `--all` 을 한 번 더 돌려 결과가 바이트 동일(`cmp` exit 0) — 재생성은 멱등이다(`catalog-regen2.log`).

## 흡수 트리 재측정

| 검사 | 파일 | 결과 |
|---|---|---|
| 링크 검사(28파일, `linkcheck.pl`) | `linkcheck-after-absorb.tsv` | OK 167(ROOT 128 · REL 34 · ROOT+TMPL 5), SKIP 4, BROKEN 1 — 흡수 전과 같다. 남은 1건은 알려진 잔여(html-report:137 의 런타임 파일 `settings.local.json` 서술) |
| 앵커 검사(16파일, `anchorcheck.pl`) | `anchorcheck-after-absorb.tsv` | `ANCHOR=MISSING` 0, `ANCHOR=OK` 26 |
| `go test -count=1 -v ./internal/template/...` | `go-test-template.txt`, `.exit` | exit 0 · `--- FAIL` 0 · `--- PASS: TestCatalogHashCoversSkillSubfiles` · `--- PASS: TestManifestHashFormat` · 템플릿·agentemit·commandemit 모두 `ok` |

테스트 직전 `pgrep` 로 다른 go test/build/vet·cli.test 0 확인.

## 자동 병합된 quality-gates-context.md

- 흡수 뒤 로컬↔템플릿 `cmp` → 다르다(9행부터).
- 대조: 기준점 `ac6c42c2d` 에서도, develop `8a44505a1` 에서도 두 사본은 이미 달랐다. 차이는 로컬 사본이 템플릿보다 앞선 문구(TRACE PROBE 주석 3줄, sync 모드 범위 설명 2줄)를 갖고 있는 기존 분기이며, 흡수가 만든 것이 아니다. 이 카드는 두 사본의 같은 링크·앵커만 고쳤다.

## 흡수 병합 커밋

- `f640e543f8c1351c704dfc56511278e3d8024946`, 부모 `7595338b9` · `8a44505a1`, 트리 `7b9c152d3a9410eea9eea249ea15d9cf17fb9070`.

## 미검증

- quality-gates-context.md 두 사본의 기존 분기를 어느 쪽으로 맞출지는 이 카드 범위 밖이다(후속 후보).
- `make build`·CI(darwin·windows 매트릭스)는 보지 않았다.
