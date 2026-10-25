# t1959 progress — SPEC-less merge-ready 경로 수리 (GH #1913)

- Class: B-유사 직접 수리(parallelizable, 리더 배정 10-25) · run tn8rpj · worktree t1930 바인딩 확인(`factory next --card t1959 --run tn8rpj` → `t1959 stage=- worktree=t1930`)
- 브랜치: WT-merge-ready-specless @ origin/main 53fbc2d206에서 생성

## 결함 (GH #1913 재현)

`moai factory merge ready`는 SPEC-less(Class B) 카드를 CLI 게이트에서 전부 거부:
- `internal/cli/factory_merge.go:77-79` — 빈 `--spec`에 하드 에러 (`--spec is required`)
- `:279` — `MarkFlagRequired("spec")`로 플래그 생략 자체가 cobra 거부
반면 엔진은 SPEC-less 경로를 이미 보유(`internal/factorylane/merge.go:126,182` `SpecDir==""` → `merge_specless.go`가 `.moai/reports/<card>/verdict.md`로 조건 (a) 판정) — 동사에서 도달 불가능했을 뿐.

## 수리 (커밋 c1c9c26d98, 2 files +90/−23)

1. `--spec` 생략 또는 카드 id를 `--spec`으로 전달 → SPEC-less 경로 라우팅(`specID=""` → SpecDir 빈 문자열)
2. 빈 specID의 `filepath.Join` 함정 명시 차단(빈 문자열 유지 — specs 디렉터리 아님)
3. SPEC 보유 카드는 progress.md 게이트 유지 — SPEC 경로·타 게이트(conflict-free·tree-identity·re-measure·window) 전부 불변
4. `MarkFlagRequired("spec")` 제거 + 플래그 헬프에 SPEC-less 계약 기술
5. 테스트 신계약: 구 git-flow `--spec is required` 핀 → 성공 회귀 2건(spec 생략·card id, cleared+window 획득) + 거부 유지(판정 파일 부재 → 조건 (a) 거부·창 미획득)

## 검증 (레인 직접, 이 트리·이 커밋)

- `gofmt -l` 두 파일: 클린 · `go build ./internal/cli/`: exit 0
- `go test ./internal/cli/ -run 'Specless|MergeReady' -count=1` → `ok ... 21.569s`
- `go test ./internal/cli/ -run 'TestFactoryMergeReadySpecLess' -count=1 -v` → `spec_omitted` PASS · `spec_is_card_id` PASS · `TestFactoryMergeReadySpecLessWithoutVerdictRefused` PASS (ok 4.291s)
- `go vet ./internal/cli/`: exit 0 · `golangci-lint run ./internal/cli/...`: `0 issues`

## 착지

- push: `WT-merge-ready-specless` → origin(신규 브랜치, non-force)
- PR: **#1915 draft**(base main, 제목 `[t1959] fix(factory): reach merge ready for SPEC-less Class B cards`)
- card-review: `.moai/reports/t1959/card-review.md` — codex CLI 폴백(MCP 도구 부재, fallback 사유 기재) **verdict pass** @ c1c9c26d9891845fb1c11fb1a3df1a26ccb6b4ce, tree_digest 261ab9e326f0f6f2, codex-cli 0.163.0, findings 없음
- 독립 감사: sync-auditor 스폰(head 결속 1회) — 판정 `.moai/reports/t1959/verdict.md` 대기 중. N1 하네스 결함(수령 가드 재앵커) 재발 시 1회 mint 후 Gaps 기재·판정 진행 지시 완료.
- 병합·이슈 close: 리더(이슈 한글 코멘트와 함께 CI+감사 PASS 뒤)

decision record: decided_by=lane-7 (t1959 repaired, verified, landed as draft PR #1915) evidence_refs=git show c1c9c26d98, ok 21.569s, card-review.md pass, PR 1915 ladder_path=leader directive — awaiting head-bound audit verdict, then leader reply
- [감사 완료 + 영수증 착지 10-25T08:2xKST] 독립 감사 **PASS @ c1c9c26d98**(판정 파일 verdict.md, 4차원 90-95, blocking 0) — 영수증 **rcpt-26567a378837d950e5081ccf** 1회 발급 성공(타이밍 거부 없음, N1 불요). **codex 필수 게이트 원판정 fail**(P1 2건, head에서 재현) → 감사자 독립 재평가로 F1/F3 [Medium][optional]로 접어 PASS 유지(all-optional 규칙) — verdict.md §Cross-model·iteration 2에 원판정 그대로 기록. F1: --spec 생략 시 카드 레코드 SpecID 미조회 라우팅(후속 카드: SpecID 기반 라우팅) · F3: SPEC-less audited_sha가 worktree HEAD 바인딩, --branch 후보와 불일치 가능(후속 카드: 후보 팁 바인딩 또는 불일치 거부). 수렴 check 관측: `moai verify audit-plan --result-file` → `convergence_check.ok: true, unmet: []` — 출력이 판정 극성을 에코하지 않고 필수 백엔드 실행 영수증을 소비하는 것으로 관측(관측 사실로 기재). 리더 지시에 따라 `.moai/reports/t1959/` 3파일 `add -f` 커밋(`.moai/reports/*` gitignore:235 우회 — t1910·t1894 등 확립 전례) → push, 새 head **2d2494f6a9**(증거 전용, 감사 코드 head는 c1c9c26d98 불변). decision record: decided_by=lane-7 (audit PASS + receipt landed, codex fail disclosed to leader) evidence_refs=verdict.md AUDIT-VERDICT line, rcpt-26567a378837d950e5081ccf.json 249B, convergence ok ladder_path=leader directive — PR #1915 merge = leader decision on codex-fail disclosure
- [F1 수리 라운드 (리더 판정 이행) 10-25T08:3xKST] 리더 판정 수납: r1 optional 접기는 병합 근거로 수용 안 됨 → **F1 이-PR 최소 수리**: `factoryRecordSpecID` 헬퍼(레코드 SpecID 조회 — NO_ACTIVE_FACTORY·미기록 카드는 기존 SPEC-less 서비스, 모호 run은 refuse), `--spec` 생략/카드 id 시 레코드 준거 라우팅. RED-first: `TestFactoryMergeReadyRecordSpecIDKeepsSpecPathOnOmittedFlag`를 현재 코드에서 **FAIL 관측**(판정 파일만으로 cleared) → 수리 후 GREEN. 기존 3테스트 포함 scoped suite `ok 34.611s` · vet 0 · lint 0 issues. forward 커밋 `309b2670b6` push. F3는 리더 판정대로 미수리(후속 카드 후보). card-review r2: **fail** @ 309b2670b6 — finding 1건은 F3 동일 클래스(재확인), 리더 선행 처분대로 기록.
- [재감사 완료 — PASS-WITH-DEBT 10-25T08:4xKST] 감사자 라운드 2: 판정 파일 verdict.md 갱신 — **PASS-WITH-DEBT @ 309b2670b6**, F1 수리 독해·기계 검증 통과(ok 18.960s, 회귀 PASS 1.55s). **codex 원판정 fail 그대로 수록(재분류 없음)**: 신규 P2 → **F4/부채 D1** — run id에 `NO_ACTIVE_FACTORY` 문자열 포함 시 `strings.Contains`(factory_merge.go:321)가 AMBIGUOUS_FACTORY를 실행-부재로 오분류, codex가 이 head에서 재현(오염 id cleared+창 획득). 처방: 전용 센티넬 + errors.Is + 회귀. 감사자 평가: 저장소 전반 Contains 관례의 선천적 약점 + 인위적 전제 필요 + 증거 닫힘 유지 → 부채 인정 PASS-WITH-DEBT(접기 아님, fail 라인 표면화). 영수증 **rcpt-49a23ab7055d4d6033b70b86** 1회 발급. F3 미등장·원문 유지. 증거 커밋·push 후 리더 회신 — 병합 여부는 리더. decision record: decided_by=lane-7 (F1 repaired RED-first, re-audit PASS-WITH-DEBT with codex fail verbatim) evidence_refs=verdict.md iteration 3, rcpt-49a23ab7055d4d6033b70b86, ok 34.611s/18.960s ladder_path=leader directive — leader merge decision on disclosed debt
