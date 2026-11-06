# t2134 진행 기록 — Windows pre-commit e2e 실패 12건 (재현·원인 판정 먼저)

Card t2134 · run to2bbk · lane-31 · Class B (재현·귀속 먼저, Windows 훅 불가 시에만 사유 출력+skip) · 2026-11-06

## 배차·임대·기준

- 배차: 리더 — "pre-commit e2e: 12 Windows test failures. Reproduce and attribute first, then fix or skip with reason only where the hook cannot run on Windows. Classification table: t2123 tree .moai/reports/t2123/progress.md + sync-audit.md (lane-20). Parent t2123"
- 운영자 결정(직독 확인): d-20261106T133813Z-3350 — Windows 시험 정책(휴대 부분만 단언, POSIX 모드비트 동등성 탈락, 남는 것이 없으면 사유 출력 skip), t2130~t2135 발행, **이 카드들은 Codex 미승인 — Claude 감사만**
- 임대: `moai factory next --card t2134 --run to2bbk` → worktree=t2134; 만료(13:55:51Z) 후 owner 재임대 단일 호출로 복구(stage/worktree 보존 확인)
- 워크트리: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2134` (branch `WT-pre-commit-e2e`, base 04d8c27a7e = local main/rc.2)

## 레인 코드 조사 (직독)

- CI 분석(t2077 트리 pr1969-ci-analysis.md:22): "Windows: pre-commit e2e(12건) | cli | run cli-3 127703364510" — 실패 문구 2종: `precommit_gate_scope_e2e_test.go: commit blocked under the default posture (AC-002 violated)`, `precommit_staged_read_e2e_test.go: block did not come from the gofmt step`
- 대상 시험 파일: `internal/cli/precommit_gate_scope_e2e_test.go`(4시험), `internal/cli/precommit_staged_read_e2e_test.go` — 실제 훅 본문(preCommitHookContent, `#!/bin/sh` — hook_install_precommit.go:44)을 실제 git 저장소에서 구동하는 e2e
- Windows 비호환 전제(코드 직독):
  1. `gateE2EFakeMoai`/`stagedReadFailingGofmt` — `#!/bin/sh` 스크립트를 PATH의 실행 파일("moai"/"gofmt" 확장자 없음)로 설치 → Windows에서 exec 불가(확장자·셔뱅 모두 불일치)
  2. `hookPath = "/usr/bin:/bin:/usr/sbin:/sbin"` — 콜론 구분 POSIX PATH + 존재하지 않는 Windows 디렉터리
  3. 훅 본문 자체는 #!/bin/sh — Git for Windows가 자체 sh로 구동 가능하므로 훅 실행 자체는 가능성 있음(실패는 fixture 측 전제)
- 근거 체인: t2123(조각 ①, /bin/sh 19건) PASS 0.92 @47963571a1 — ④ pre-commit e2e 12건은 t2123에서 미착수로 명시

## 위임

- `manager-develop` 스폰 (depth-1 리프, effort=medium — 클래스 B 매핑), 지시:
  1. 귀속 먼저: gh run 127703364510 로그(read-only)로 12건 전수 열거 + 코드 직독 대조로 원인별 분류
  2. 수리 우선: fixture 교차 플랫폼화(Go 빌드 스텁+적확 확장자, os.PathListSeparator, POSIX 디렉터리 제거) — skip은 훅이 Windows에서 실제 불가한 경우만, 사유 출력과 함께
  3. 로컬 증거: darwin 표적 시험 유지 통과 + GOOS=windows vet/compile; Windows 실행 판정은 CI
- 상태: 백그라운드 실행 중 — 완료 통지 대기 (결과를 예측·보고하지 않음)

## run 단계 결과 (스페셜리스트 반환 + 레인 재검증)

### 커밋

- `cb57add07379b8f6d33964274e7f7ad2e73eee3a` — `[t2134] test(cli): make the pre-commit e2e fixtures cross-platform (card t2134)` (3파일 +183/−31, 트리 깨끗함 — 직독)

### 12건 귀속 (구현원 표 — CI 근거)

- 카드 명시 런 127703364510은 404(소멸) → 카드 폴백 절차대로 **run 42147372729**(release-pr-multi-os.yml, Release Verify windows-latest cli-3, selector ^Test[J-R])의 artifact 스트림에서 원문 확보
- **기관 A — 9건 전부**: `hookPath = "/usr/bin:/bin:/usr/sbin:/sbin"`(콜론 PATH+존재하지 않는 디렉터리) → 커밋 서브프로세스 PATH 전체 붕괴 → git.exe가 번들 sh 스폰 불가: `error: cannot spawn .git/hooks/pre-commit: No such file or directory`(gate_scope 4건 + staged_read 5건, 시험×줄 원문은 구현원 보고·판정 파일 경유)
- **기관 B — 잠재**: `#!/bin/sh` 스텁("moai", gofmt 스텁) 확장자 없음 → Windows exec 조회 불가(A 수리 후에 바일 것)
- **기관 C — 훅 측, 미수리(레인 판단: 프로덕션 동결 존중)**: `preCommitHookContent`의 `git show ":$f" | gofmt -l /dev/stdin`은 네이티브 gofmt.exe로 불가(/dev/stdin 무특수 취급) — **실제 Windows 사용자의 Go 커밋 차단 프로덕션 결함**. 근거: 범위 밖 2시험(TestPreCommitRelocation_GoodCommitPasses·TestPreCommitHook_GoVetBlocks)이 유효 PATH+실제 gofmt에서 깨끗한 입력에 실패. **후속 카드 후보 + 동결 해제 판단은 리더 소관으로 보고**
- 범위 밖 형제 실패 4건(precommit_vet_output_e2e ×2=기관 A, hook_install_precommit_test+precommit_relocation_test=기관 C) — 보고만

### 수리 요지

- 기관 A: 신설 `hookE2EPath` — os.PathListSeparator + exec.LookPath로 실재 디렉터리(git·sh·grep·sed·sort·tr·dirname)만 조립, POSIX 디렉터리는 존재 시에만. 실제 moai 비도달(AC-007 격리 유지)
- 기관 B: 신설 `buildE2EBinary` — `go build`로 실제 바이너리 스텁(Windows moai.exe/gofmt.exe, 타 OS 무확장자). 계약 보존: fake moai=awk 파싱 충실 포트(MOAI_PRECOMMIT=1+pre_commit.enabled 미true→0, 그 외 1, 파일 부재→0), 실패 gofmt=exit 1 무stdout
- 기관 C: 4개 실제-gofmt staged_read 사례는 Windows에서 **사유 출력과 함께 skip**(d-20261106T133813Z-3350의 "nothing portable remains" 절 적용 — 래퍼 스텁으로 C를 가리면 실제 프로덕션 결함이 은폐됨), gofmt-스텁 사례는 전 플랫폼 실행

### 레인 재검증 (HEAD cb57add073 직접 실행)

- `go test ./internal/cli/ -run 'TestPrecommitE2E|TestPrecommitStaged' -count=1` → ok 6.107s
- `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` → exit 0
- `GOOS=windows go test -c -o /dev/null ./internal/cli/` → exit 0
- 스페셜리스트 실행분(같은 트리 키) 재사용: darwin -v 9건 PASS, 광역 TestPre[Cc]ommit ok 19.728s, gofmt 무출력

### 미검증 (Gaps)

- Windows 실행 판정 전부 → CI. 잔여 불확실 2가지: 유효 PATH에서 git.exe가 #!/bin/sh 훅을 스폰하는지(로그에 sh-직접호출 선례만 존재), MSYS `command -v`의 .exe 해석 — cli-3 재실행 판정
- 형제 4건 + 기관 C — 범위 밖, 후속 카드/리더 판단 기록

## sync 감사 1라운드 결과 (2026-11-06 ~23:45 local)

- 판정: **PASS-WITH-DEBT** — 가중 0.93 (기능성 92 / 보안성 96 / 장인정신 90 / 일관성 94). 판정 파일 `.moai/reports/t2134/sync-audit-iter1.md` (verdict:/audited_sha:/receipt: 줄 직독 확인, audited=cb57add07379b8f6d33964274e7f7ad2e73eee3a)
- 부채(판정 명시): 기관 C(훅 측 gofmt -l /dev/stdin Windows 블로커 — 동결 존중 미수리) + 형제 2건(precommit_vet_output_e2e_test.go:50 기관 A)
- 영수증 상태: **none — 이 패밀리 Codex 미승인(d-20261106T133813Z-3350, 결정판 직독 근거)**. 트리 일반 설정 gates.codex: required의 Stop-hook가 PASS-계열 판정에 영수증을 요구했으나 감사관이 호출 대신 운영자 비승인 우선을 적용, 거부 이벤트·처분을 판정 파일 Iteration history에 기록(침묵 대체 없음, 게이트 압력으로 토큰 변경 없음). 해소(패밀리 gates 면제 또는 명시적 Codex 승인)는 리더/운영자 소관 — 리더 전달
- 감사관 자체 검증: 표적 9/9 PASS, 패밀리 광역 ok 20.649s, windows vet/compile exit 0, gofmt 무출력, 귀속 표 코드 대조
- 결정 기록: decided_by=lane-31 sync-audit-r1 evidence_refs=.moai/reports/t2134/sync-audit-iter1.md ladder_path=Claude sync-auditor(af-flow, Codex 미승인 패밀리) — PASS-WITH-DEBT 수용

### Gaps
