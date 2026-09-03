# t297 sync-audit 판정 (card t297, Class B — no SPEC)

- Auditor: sync-auditor (독립 컨텍스트, lane-15 디스팟)
- Tree: `WT-launch-ledger-write` @ `55aec1b45`
- Date: 2026-09-03
- 이 문서는 감사자 판정을 lane이 기록한 것

## Verdict

**PASS — 0.93/1.00** (harmonic: Functionality 95 / Security 92 / Craft 93 / Consistency 94;
must-pass 독립 통과). 결함 3건 전부 MINOR·optional, 차단 0. **카드 종결 승인 권고.**

## 재현 판정 — 유효

red-prefix-run.log 전문 직독: `TestRecordFromSubtreeFoldsIntoRegisteredRoot` FAIL이 실함수
`RecordLastUsedProfileForProject`(profile.go:553) 경로의 실행 기반. mutant1(fold 제거)이 동일
실패 서명(4행, want 1) 재현 — RED가 fold 결함임을 교차 확인. 성장 수치(5회→6행) 관측 문장 존재.

## 불변식 fold == resolve (감사자 코드 직독 표, 7사례 전부 ✓)

exact/alias 자리 갱신 ✓ · 중첩 독립 프로젝트 exact 우선 ✓(mutant3 teeth) · 등록 생존 조상 fold ✓ ·
죽은 프로필 조상 fold 후 덮어씀 ✓(profile.go:629-634 문서화된 의도적 비대칭) · disposed 조상 skip ✓ ·
cold-start ✓ · 레거시 중복 자기 행 갱신 ✓. 쓰기 경로는 어떤 것도 delete하지 않음(코드 확인).

## 다른 독자 — 없음 (명시적 답)

grep 전수(internal/ + cmd/, 테스트 제외): 원장 읽기는 `ResolveLaunchProfileForProject` 단일 표면
(launcher.go:150 호출 유일). statusline·tui 원장 접근 0건. `last_profile`은 write-only.
append-only/per-worktree 가정 리더 없음 — 구조적 최대 리스크 부재 확인.

## 회수 감사 — 통과

6지점 코드 확인(배선 실재). 면제 검증: `--json` 분기(clean.go:69), preview 조기 반환(264-271) —
prune(299) 미도달. best-effort(실패=stderr 경고, 폐기 실패 안 시킴). 술어 오탐 심각도 Low
(편의 캐시 손실, 다음 -p 실행 재등록 자가 치유).

## 감사자 직접 실행

- `go test ./internal/profile -run 'TestRegression_|TestRecordFromSubtree|TestRecordFromRegisteredNested|TestPruneStaleProjectEntries' -count=1 -v` → **14/14 PASS**
- `go test ./internal/cli/worktree -run '…ledger…'` → 배선 13건 포함 전부 PASS
- `go test ./internal/web -count=1` → ok 6.515s
- `go test ./internal/profile -count=1 -cover` → 86.0% (run 주장과 일치)
- golangci-lint 0 issues · vet OK · gofmt clean (감사자 실행)

## teeth — 진짜

4개 변이 로그 FAIL 줄이 테스트 소스 단언문과 정확히 일치
(write_normalization_test.go:83/109/168/171/198/201/205, ledger_prune_test.go:51/81/238/278,
launch_ledger_test.go:125/192/210/213).

## docs·close 위생 — 통과

profile.md 4-locale now-false 문장 제거+신규 동작 문장 코드 일치 · worktree.md 4-locale 회수 기술
코드 일치 · CHANGELOG 6 call sites/면제/멱등/출력 전부 코드·테스트와 일치 · 스코프
304bc8158..55aec1b45 = 26파일 이탈 0 · sync 커밋 internal/ 미접촉 · backfill 확인.

## 결함 (전부 MINOR·optional)

- **F1** prune이 임시 rename/unmount 프로젝트의 기억 프로필 행 삭제 — 자가 치유, 수리 불요
- **F2** PruneStaleProjectEntries 주석 "lookupProjectKey requires os.Stat" — exact-match 분기
  (Stat 없이 반환, profile.go:301-302)에 부정확. 결론은 성립, 근거 서술 1분기 오류. 기록 자문 잔존
  (lane 처분: 코드 코멘트만, 감사 종결 후 수리 가치 없음)
- **F3** prune이 선존재 무잠금 read-modify-write race(internal/cli/profile.go:79 문서화)에 작성자 추가 —
  최악 편의 행 갱신 유실. 후속 카드 불요, 관측만.

## 리드 잔여 위험

launch.yaml 동시 쓰기 race는 선존재·문서화 한계 — 이 변경으로 교차 빈도 미약 상승.
증상 최악은 편의 행 1개 유실, 유계 결함 아님.
