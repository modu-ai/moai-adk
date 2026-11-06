# t2141 진행 기록 — pre-commit 훅 gofmt /dev/stdin Windows 운영 결함 + 형제 시험 4건

Card t2141 · run to2bbk · lane-31 · Class B (재현 먼저, 동결 해제 판단 포함) · 2026-11-07

## 배차·임대·기준

- 배차: 리더 — "pre-commit 훅 gofmt /dev/stdin Windows 운영 결함과 형제 시험 4건, 동결 해제 판단 포함, class B. 재현 먼저. Source: lane-31 t2134 report" — 운영자 승인 d-20261106T163347Z-7adc (제안된 후속 카드)
- Codex: 감사 단계당 최대 2회(305c)
- 임대: `moai factory next --card t2141 --run to2bbk` → worktree=t2141 (branch `WT-the-frozen-precommithook`, base 04d8c27a7e)
- 전환 경로: t2134 merge-ready 확정(영수증 rcpt-0d109a464b62cf621e5d71f1 포함, RECOMMEND 갱신) 후 본 카드 착수

## 레인 코드 조사 (t2134 런에서 확보한 재현·귀속 기록 이관)

- **재현 기록(Windows, CI 근거)**: run 42147372729 (release-pr-multi-os.yml, Release Verify windows-latest cli-3) artifact 스트림 — 유효 PATH+실제 gofmt 환경에서 `TestPreCommitRelocation_GoodCommitPasses`·`TestPreCommitHook_GoVetBlocks`가 **깨끗하고 파싱 가능한 입력**에서 `[pre-commit] FAILED: gofmt could not check the staged bytes of: good.go / bad.go`로 실패. 원인: 동결된 preCommitHookContent(hook_install_precommit.go:44, `#!/bin/sh`)의 `git show ":$f" | gofmt -l /dev/stdin` — 네이티브 gofmt.exe는 /dev/stdin 특수 파일이 없어 staged 바이트 검사 불가 → 훅이 gofmt 단계 실패로 처리 → 실제 Windows 사용자 Go 커밋 차단
- **형제 4건**: precommit_vet_output_e2e_test.go ×2(기관 A — t2134와 동일 POSIX PATH 문제, hookE2EPath 미적용 파일), hook_install_precommit_test.go + precommit_relocation_test.go(기관 C 실 gofmt 경로)
- **동결 해제 판단(카드가 판단 포함 지시)**: t2134에서 프로덕션 동결을 존중해 미수리했으나, 본 카드가 운영자 승인으로 동결 해제 판단을 소관으로 함. 휴대형 대안: staged 바이트를 임시 파일에 쓰고 `gofmt -l <tmpfile>`로 검사(/dev/stdin 특수 파일 의존 제거) — POSIX sh와 Windows sh(Git for Windows) 모두에서 동작하는 표준 패턴. 임시 파일 정리(cleanup trap) 동반 필요

## 위임

- `manager-develop` 재개(동일 구현자 — t2134 fixture·귀속 맥락 보유), 지시: (1) 훅 본문 직독+휴대형 재작성 설계, (2) 형제 4건 수리(A 2건은 hookE2EPath 재사용, C 2건은 휴대형 훅으로 GREEN화), (3) 재현·RED는 CI 소유 명시, (4) 단일 커밋 전방
- 상태: 백그라운드 실행 중 — 완료 통지 대기 (결과를 예측·보고하지 않음)

## run 단계 결과 (스페셜리스트 반환 + 레인 재검증)

### 커밋

- `8d6d4ef59f923bac78a9cecfcf61ea4341dd08b4` — `[t2141] fix(cli): make the pre-commit hook's gofmt check portable; cover the remaining Windows fixtures (card t2141)` (5파일 +105/−6, 트리 깨끗함 — 직독)

### 동결 해제 근거 (카드 지시 판단)

- 결함은 훅 본문 외 어디서도 고칠 수 없음: gofmt.exe가 /dev/stdin을 열지 못함(Windows 특수 파일 부재) → 실제 Windows 사용자 Go 커밋 전부 차단(CI run 42147372729에서 깨끗한 입력 실패 2건 = 재현 기록, 카드 지시대로 t2134 보고서에서 인용). darwin RED 없음(darwin은 파이프 형태 정상 실행) — 로컬 RED 대신 CI 재현+기계적 원인이 근거
- 운영자 승인 d-20261106T163347Z-7adc가 동결 해제 판단을 카드 소관으로 위임 — 해제 결정: 휴대형 재작성(아래)

### 휴대형 설계

- staged 바이트(`git show ":$f"`)를 pid 키 임시파일(`${TMPDIR:-/tmp}/moai-precommit-gofmt.$$`)에 기록 → `gofmt -l <tmpfile>` — /dev/stdin 제거, Git-for-Windows sh 범위의 bashism만 사용(백틱 없음 — raw-string 제약 준수)
- 정리: 파일당 검사 뒤 항상 rm + 스캔 뒤 1회(항상-실행 옵션), pid 키로 동시 커밋 충돌 방지
- POSIX 판정 바이트 동등: 인덱스 읽기 실패=빈 파일=gofmt 통과(구 파이프와 동일), 파싱 불가=err 분기(차단), 미포맷=fmt 분기(차단), 메시지 불변 — darwin 수트가 회귀 증명
- 템플릿 미러(.git_hooks/pre-commit) 동일 편집 — TestPreCommitTemplateMatchesConstant가 바이트 동일 강제(미러는 같은 표면의 배포 형태, lockstep 편집 필수 — 범위 원장에 표기)

### 형제 4건 처분 + 검증 중 발견 실결함

- TestPrecommitVetFailureShowsDiagnostics·TestPrecommitVetPassStaysQuiet(기관 A): hookE2EPath 포팅(신설 공유 precommit_e2e_fixtures_test.go)
- TestPreCommitHook_GoVetBlocks·TestPreCommitRelocation_GoodCommitPasses(기관 C): 본문 변경 없이 휴대형 훅으로 Windows GREEN 예정
- **검증 중 발견·수리**: moduleShimPath가 rm을 심링크하지 않아 모듈워크 4시험이 `rm: command not found`(exit 127) — 휴대형 훅의 정당한 rm 의존(sed·sort·tr과 동류)에 대한 fixture 결함으로 수정·재검증(환경 실패 아님)

### 레인 재검증 (HEAD 8d6d4ef59f 직접 실행, 격리 환경 귀속)

- `GOTOOLCHAIN=go1.27.1 GOCACHE=/tmp/t2117-gocache go test ./internal/cli/ -run 'TestPre[Cc]ommit|TestPreCommit' -count=1` → ok 30.960s (49시험)
- GOOS=windows vet/compile — 배경 실행 중(완료 통지 대기), 스페셜리스트 실행분: WIN-VET-OK·WIN-TESTC-OK(같은 트리 키 재사용 예정)
- gofmt 무출력(스페셜리스트, 같은 트리 키)

### 미검증 (Gaps)

- Windows 실행 판정 → CI(release-pr-multi-os). 로컬은 vet/compile만
- 전체 수트·타 패키지 미실행(레인 로컬 규율)
- staged_read·gate_scope 2파일은 이 트리에서 여전히 pre-t2134 fixture(t2134 수정분이 미착지) — 그 카드 착지로 해소, 본 카드 범위 밖

### Gaps
