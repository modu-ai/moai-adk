# Hooks 전수 목록·검증 및 개선 보고서

기준: 2213871afb7d655f411c46a34fd38bb8153278fc / main / 2026-09-11. 현재 수정된 설정을 포함한 파일 복제본에서 측정. 원본 코드 수정 없음.

개선 항목 11건을 재현했다. P1 5건은 경로 격리·자료 보존·차단 결과 보존에 관한 우선 개선 사항이며, P2 6건은 검사 누락과 실행·검증 정확성에 관한 사항이다.

## Claim

개선 항목 11건을 재현했다. P1 5건은 경로 격리·자료 보존·차단 결과 보존에 관한 우선 개선 사항이며, P2 6건은 검사 누락과 실행·검증 정확성에 관한 사항이다.

## Baseline-attribution

원본: /Users/goos/MoAI/moai-adk-go
시험 복제본: /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-hooks-audit-zmytrok0
601개 파일 · 148,058줄 · Go 463개 · 셸 96개 · 시험/fixture 305개 · 비시험 296개. SHA-256 비교 결과 변경 0개.

## 개선 항목

### H01 · P1 · 새 파일은 심볼릭 링크를 따라 프로젝트 밖에 기록할 수 있음
- 위치: internal/hook/pre_tool.go:989
- 확인: 프로젝트 내부 linked 디렉터리를 외부 디렉터리로 연결한 뒤, 아직 존재하지 않는 linked/new.txt를 검사했다. 경로 검사는 거부하지 않았고, 임시 파일은 외부 대상에 생성되었다.
- 원인: EvalSymlinks가 최종 파일 부재로 실패하면 전체 경로를 원래 문자열로 되돌린다. 이미 존재하는 부모의 링크까지 해석하지 않으므로 프로젝트 경계 검사가 우회된다.
- 근거: `TestAuditNewFileSymlinkBoundary`
```text
new_symlink_file decision="" reason=""
outside_write=true
```
- 권고: 가장 가까운 기존 부모를 실제 경로로 해석한 뒤 나머지 새 경로를 결합하고, 프로젝트 및 허용된 외부 디렉터리 경계를 다시 검사한다.
- 완료 조건: 일반 신규 파일은 허용하고, 외부로 연결된 부모 아래의 신규 파일은 거부하는 대조 시험을 추가한다.
- 신뢰도: 높음 — 재현 조건 안에서 확인. 병합 판단: 관련 경로 수정·출시 전 해결 권고.
- 한계: 도구의 OS 샌드박스까지 우회한다는 주장은 아니다. 이 finding은 MoAI 자체 경로 검사에 관한 것이다.

### H02 · P1 · 최근 갱신한 다른 팀의 자료가 오래된 디렉터리 시각 때문에 삭제됨
- 위치: internal/hook/session_end.go:446
- 확인: 팀 디렉터리 시각을 과거로 설정하고 config.json을 새로 기록했다. 다른 session ID와 in_progress 작업을 둔 상태에서도 GC가 팀과 작업 디렉터리를 모두 삭제했다.
- 원인: 부모 디렉터리의 수정 시각은 기존 파일 내용의 갱신 시각과 다르다. 현재 세션의 종료가 홈 디렉터리 전체 팀을 대상으로 이 시각만 보고 정리한다.
- 근거: `TestAuditStaleActiveTeam`
```text
recently_updated_team_deleted=true task_deleted=true
```
- 권고: 팀 소유 세션의 종료가 확인된 경우에만 정리하고, 세션의 생존 상태나 명시적인 임대 만료 기록을 사용한다. 확인 불가 항목은 보존한다.
- 완료 조건: 최근 갱신한 타 세션 팀과 오래된 종료 팀을 함께 만들고 전자는 보존되는지 검증한다.
- 신뢰도: 높음 — 재현 조건 안에서 확인. 병합 판단: 관련 경로 수정·출시 전 해결 권고.
- 한계: 실제 사용자의 활성 세션을 종료하거나 자료를 삭제하지 않았다. 재현은 전용 임시 홈 디렉터리에서 수행했다.

### H03 · P1 · 팀 디렉터리가 없다는 이유만으로 다른 작업 목록을 즉시 삭제함
- 위치: internal/hook/session_end.go:498
- 확인: 별도 세션 이름의 tasks 디렉터리에 방금 생성한 in_progress 작업을 두고 대응하는 teams 디렉터리는 만들지 않았다. GC는 이 작업을 즉시 삭제했다.
- 원인: 팀 디렉터리 부재를 작업 소유권이나 종료의 증거로 간주한다. 작업의 생성 시각, 상태, 세션 소유자를 확인하지 않는다.
- 근거: `TestAuditStandaloneTaskGC`
```text
fresh_standalone_task_deleted=true
```
- 권고: 현재 세션이 생성·소유한 작업만 정리하거나, 별도 소유권 기록으로 삭제 권한과 종료 상태를 확인한다. Stat 오류와 실제 부재도 구분한다.
- 완료 조건: 새 독립 작업 목록과 타 세션 작업 목록은 보존하고, 종료가 입증된 소유 작업만 삭제하는 시험을 추가한다.
- 신뢰도: 높음 — 재현 조건 안에서 확인. 병합 판단: 관련 경로 수정·출시 전 해결 권고.
- 한계: 정상 런타임에서 이러한 목록이 얼마나 자주 생기는지는 측정하지 않았다. 생성된 자료를 삭제하는 조건 자체를 재현했다.

### H04 · P2 · 이름이 {}인 디렉터리 안의 임의 자료를 일괄 삭제함
- 위치: internal/hook/session_end.go:793
- 확인: 프로젝트의 {}/user-data.txt에 임의 자료를 기록한 뒤 정리 함수를 호출했다. 파일의 생성 주체와 관계없이 전체 디렉터리가 삭제되었다.
- 원인: 과거 템플릿 치환 오류의 잔재인지 확인하지 않고 디렉터리 이름만으로 RemoveAll을 실행한다. 링크 자체는 제외하지만 일반 사용자 디렉터리는 보호하지 못한다.
- 근거: `TestAuditBogusDirectoryData`
```text
unattributed_user_file_deleted=true
```
- 권고: MoAI 생성 표식과 예상 하위 구조가 확인된 잔재만 정리한다. 근거가 없으면 경고를 남기고 보존한다.
- 완료 조건: 같은 이름의 사용자 디렉터리는 보존하고, 명시적으로 표식된 잔재만 정리되는지 검사한다.
- 신뢰도: 높음 — 재현 조건 안에서 확인. 병합 판단: 개선 작업에 포함 권고.
- 한계: 특이한 디렉터리 이름을 전제로 하므로 영향 가능 범위는 제한적이다.

### H05 · P1 · 작업 공간 이름 ../..가 기본 체크아웃 재사용으로 연결됨
- 위치: internal/hook/worktree_create.go:94
- 확인: 빈 임시 Git 저장소에 WorktreeName="../.."를 전달했다. 핸들러가 반환한 실제 경로는 해당 저장소의 기본 체크아웃과 같았다.
- 원인: 이름을 경로에 바로 결합하고, 그 위치에 디렉터리만 있으면 재사용한다. 경로 경계, 정상 worktree 등록 여부, branch 일치를 확인하지 않는다.
- 근거: `TestAuditWorktreeTraversalReuse`
```text
worktree_name=../.. returned_primary_checkout=true
```
- 권고: 이름의 경로 구분자·상위 이동을 거부하고 실제 경로가 지정된 부모 아래인지 검사한다. 재사용 전 git worktree list와 소속·branch를 확인한다.
- 완료 조건: ../.., 경로 구분자, 외부 링크, 일반 디렉터리는 거부하고 등록된 동일 worktree만 재사용하도록 검사한다.
- 신뢰도: 높음 — 재현 조건 안에서 확인. 병합 판단: 관련 경로 수정·출시 전 해결 권고.
- 한계: 정상 Claude가 악성 이름을 생성한다는 증거는 없다. 입력 검증과 격리 계약의 실패를 함수 경계에서 재현했다.

### H06 · P1 · 동일 커밋의 두 번째 품질 검사에서 앞선 차단 결과가 사라짐
- 위치: .claude/hooks/moai/sync-phase-quality-gate.sh:179
- 확인: go vet와 go build가 모두 7로 종료하도록 한 동일 커밋에서 훅을 두 번 실행했다. 첫 응답에는 차단 결과가 있으나 두 번째 응답은 비어 있었다.
- 원인: 검사 전에 HEAD만 저장하고 다음 호출은 즉시 반환한다. 성공·실패·검사 중단을 구분하지 않으며 실패 결과도 재전달하지 않는다.
- 근거: `sync_sentinel`
```text
first_exit=0; first_stdout contains decision:block
second_exit=0; second_stdout=""
```
- 권고: 검사 결과와 상태를 함께 원자적으로 저장한다. 완료된 실패는 같은 입력에서 재전달하고 중단된 검사는 다시 수행한다. 검사 대상이 작업 트리라면 HEAD 외 대상 내용 식별자도 필요하다.
- 완료 조건: 실패→동일 입력 재호출, 중단→재호출, 수정→재호출을 각각 검사한다.
- 신뢰도: 높음 — 재현 조건 안에서 확인. 병합 판단: 관련 경로 수정·출시 전 해결 권고.
- 한계: 실제 Claude의 이 JSON 수용 여부는 별도 검증하지 않았다. 이 항목은 훅이 생성하는 응답의 차단 결과 소실에 관한 것이다.

### H07 · P2 · Java·Ruby 검사 실패가 성공 종료 코드로 집계됨
- 위치: .claude/hooks/moai/sync-phase-quality-gate.sh:243
- 확인: 컴파일러/검사기 대역이 호출된 사실을 기록하고 7로 종료하도록 했다. 두 언어 모두 게이트 로그에는 0과 allow가 기록되었다.
- 원인: Java는 뒤의 head 성공이 파이프라인 상태가 되고, Ruby의 find -exec … \;는 각 검사 실패를 전체 명령의 실패로 전달하지 않는다.
- 근거: `sync_java / sync_ruby`
```text
java: compiler_invoked=true; decision=allow; javac compile check=0
ruby: compiler_invoked=true; decision=allow; ruby syntax=0
```
- 권고: 실제 검사 명령의 종료 코드를 먼저 보존하고 출력만 나중에 제한한다. 파일별 실행은 하나라도 실패하면 실패를 반환하도록 합산한다.
- 완료 조건: 실패 대역과 성공 대역을 각각 실행해 실제 코드가 최종 판정에 반영되는지 검사한다. PHP·Scala 등 유사 분기도 같은 방식으로 검증한다.
- 신뢰도: 높음 — 재현 조건 안에서 확인. 병합 판단: 개선 작업에 포함 권고.
- 한계: Java·Ruby에서 직접 확인했다. 유사한 다른 언어 분기까지 실제 실행했다고 주장하지 않는다.

### H08 · P2 · .cpp 파일에서 C++ 검사기가 호출되지 않음
- 위치: .claude/hooks/moai/sync-phase-quality-gate.sh:267
- 확인: CMakeLists.txt와 잘못된 내용의 bad.cpp를 둔 커밋에서 실행했다. g++ 대역 호출 기록이 없는데 게이트는 검사 코드 0으로 허용했다.
- 원인: find의 -o 결합 우선순위 때문에 .cpp는 왼쪽 조건만으로 일치하며 오른쪽의 -exec를 실행하지 않는다.
- 근거: `sync_cpp`
```text
compiler_invoked=false; decision=allow; g++ syntax check=0
```
- 권고: 확장자 조건 전체를 괄호로 묶고 공통 -exec를 연결한다. 검사 대상 0건은 성공 검사와 구분해 기록한다.
- 완료 조건: .cpp와 .cc를 각각 포함한 대조 시험에서 호출 파일 목록과 실패 전달을 모두 확인한다.
- 신뢰도: 높음 — 재현 조건 안에서 확인. 병합 판단: 개선 작업에 포함 권고.
- 한계: 이번 재현은 .cpp 단독 입력이다. 모든 C++ 빌드 시스템의 컴파일 결과를 측정한 것은 아니다.

### H09 · P2 · Kotlin 프로젝트가 Java로 감지되어 검사 없이 종료됨
- 위치: .claude/hooks/moai/sync-phase-quality-gate.sh:62
- 확인: build.gradle.kts와 Main.kt만 있는 sync 커밋에서 실행했다. kotlinc가 호출되지 않았고 게이트 로그도 생성되지 않았다.
- 원인: Gradle 마커를 모두 Java로 분류한 뒤 변경 파일은 .java만 찾는다. .kt 변경은 코드 변경 0건으로 처리되어 Kotlin 분기에 도달하지 않는다.
- 근거: `sync_kotlin`
```text
compiler_invoked=false; log_exists=false; exit=0
```
- 권고: 빌드 마커와 실제 소스 언어를 함께 판정하거나 Gradle 프로젝트의 Java·Kotlin 변경을 모두 다룬다. 다중 언어 프로젝트도 명시적으로 처리한다.
- 완료 조건: Kotlin 단독, Java 단독, 혼합 프로젝트에서 변경 파일별 검사 경로를 확인한다.
- 신뢰도: 높음 — 재현 조건 안에서 확인. 병합 판단: 개선 작업에 포함 권고.
- 한계: Kotlin/Gradle 실행 도구의 실제 설치나 빌드 성공 여부와 무관하게 감지 경로를 재현했다.

### H10 · P2 · 단발성 ConfigChange 프로세스가 비동기 검증 결과를 남기지 못함
- 위치: internal/hook/config_change.go:75
- 확인: 같은 빌드의 CLI를 별도 프로세스로 다섯 번 실행하고 잘못된 YAML을 전달했다. 다섯 번 모두 거부 경고가 없었다. 핸들러의 WaitGroup 완료를 기다리는 대조 시험에서는 거부 경고가 나타났다.
- 원인: 핸들러는 고루틴을 만든 직후 반환하며 검증은 20ms 지연 후 시작한다. CLI의 종료 시 trace만 기다리고 핸들러의 비동기 작업은 기다리지 않는다.
- 근거: `probes-async.json / TestAuditConfigAsyncControl`
```text
CLI: exit=0, stdout={}, rejection_observed=false (5/5)
joined control: WARN config reload rejected (async)
```
- 권고: 단발성 훅에서는 짧은 검증을 동기 수행하거나 종료 전에 제한된 범위에서 완료를 기다린다. 긴 작업은 지속 프로세스에 전달하고 접수·완료를 기록한다.
- 완료 조건: 핸들러 단위 WaitGroup 시험뿐 아니라 실제 CLI 프로세스가 종료된 뒤 기대한 결과가 남는지 검사한다.
- 신뢰도: 높음 — 재현 조건 안에서 확인. 병합 판단: 개선 작업에 포함 권고.
- 한계: 다섯 번의 실행을 관측한 결과다. 모든 비동기 핸들러가 항상 실패한다고 일반화하지 않는다.

### H11 · P2 · pre-commit 형식 검사가 스테이징 내용 대신 작업 파일을 읽음
- 위치: internal/cli/hook_install_precommit.go:61
- 확인: 형식이 맞지 않는 main.go를 스테이징하고 작업 파일만 gofmt로 수정했다. 인덱스의 내용은 여전히 형식이 맞지 않지만 훅의 gofmt 단계는 허용했다.
- 원인: git diff --cached로 파일 이름을 구한 뒤 gofmt -l은 작업 트리 파일을 읽는다. 부분 스테이징 상태에서는 커밋할 내용과 검사한 내용이 다르다.
- 근거: `probes-index.json`
```text
staged_needs_formatting=true; hook_exit=0
```
- 권고: git show :path로 인덱스 바이트를 검사한다. 패키지 단위 검사도 인덱스 스냅샷을 대상으로 하거나 검사 범위를 명확히 표시한다.
- 완료 조건: 스테이징 내용과 작업 파일을 서로 반대로 만든 양방향 대조 시험을 추가한다.
- 신뢰도: 높음 — 재현 조건 안에서 확인. 병합 판단: 개선 작업에 포함 권고.
- 한계: gofmt 단계만 분리하기 위해 go와 moai는 성공 대역으로 두었다. 전체 heavy gate가 이 커밋을 반드시 허용한다는 주장은 아니다.

## Evidence

### 핵심 패키지 시험
```text
cwd=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-hooks-audit-zmytrok0
unset CLAUDE_PROJECT_DIR CLAUDE_SESSION_ID CODEX_THREAD_ID MOAI_SESSION_ID MOAI_PROJECT_DIR && go test -p 2 -count=1 -timeout 180s ./internal/hook/... ./internal/codexadapter/... ./internal/codexwiring/...
ok  	github.com/modu-ai/moai-adk/internal/hook	61.499s
ok  	github.com/modu-ai/moai-adk/internal/hook/handoff	1.146s
ok  	github.com/modu-ai/moai-adk/internal/hook/memo	0.347s
ok  	github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy	0.380s
ok  	github.com/modu-ai/moai-adk/internal/hook/mx	18.705s
ok  	github.com/modu-ai/moai-adk/internal/hook/mx/complexity	0.474s
ok  	github.com/modu-ai/moai-adk/internal/hook/perf	17.323s
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	13.180s
ok  	github.com/modu-ai/moai-adk/internal/hook/security	7.034s
ok  	github.com/modu-ai/moai-adk/internal/hook/testutil	0.465s
ok  	github.com/modu-ai/moai-adk/internal/hook/trace	0.538s
ok  	github.com/modu-ai/moai-adk/internal/codexadapter	0.846s
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	1.356s
```
원본 기록: tests-core.log

### Go 구문 분석
```text
cwd=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-hooks-audit-zmytrok0
go run /tmp/moai-hook-ast-01a08e35.go /Users/goos/MoAI/moai-adk-go/reports/hooks-audit-20260911-01a08e35/inventory.json /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-hooks-audit-zmytrok0
go_ast_files=463 parse_errors=0 functions=3917
```
원본 기록: ast.log

### 셸 문법 검사
```text
cwd=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-hooks-audit-zmytrok0
python3로 inventory의 셸 파일마다 subprocess.run(["bash", "-n", snapshot_path]) 실행
bash_syntax_files=96 failures=0
```
원본 기록: shell-syntax.json

### 정적 분석
```text
cwd=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-hooks-audit-zmytrok0
unset CLAUDE_PROJECT_DIR CLAUDE_SESSION_ID CODEX_THREAD_ID MOAI_SESSION_ID MOAI_PROJECT_DIR && go vet -p 2 ./internal/hook/... ./internal/codexadapter/... ./internal/codexwiring/...
vet_exit=0
(vet stdout/stderr: empty)
```
원본 기록: vet.log

### CLI 집중 시험
```text
cwd=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-hooks-audit-zmytrok0
unset CLAUDE_PROJECT_DIR CLAUDE_SESSION_ID CODEX_THREAD_ID MOAI_SESSION_ID MOAI_PROJECT_DIR && go test -p 2 -count=1 -timeout 120s ./internal/cli -run '^Test(HookCmd|HookHarness|HookProtocol|PrePushHookConventionBlockPlacement|MultiReviewGate|CodexHarness|CodexReviewGate)'
ok  	github.com/modu-ai/moai-adk/internal/cli	1.014s
```
원본 기록: tests-cli-focused.log

### 넓은 통합 시험 — 미통과
```text
cwd=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-hooks-audit-zmytrok0
unset CLAUDE_PROJECT_DIR CLAUDE_SESSION_ID CODEX_THREAD_ID MOAI_SESSION_ID MOAI_PROJECT_DIR && go test -p 2 -count=1 -timeout 180s ./internal/cli ./internal/template -run 'Hook|PreCommit|PrePush|ReviewGate|StopGoal|Codex'
--- FAIL: TestPreCommitLegacyNoRecord (0.52s)
    --- FAIL: TestPreCommitLegacyNoRecord/c_no_record_pinned_released_body (0.04s)
        hook_install_precommit_attribution_test.go:259: cannot read the pinned released hook body "v3.1.2:internal/template/templates/.git_hooks/pre-commit" from git in /private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-hooks-audit-zmytrok0: exit status 128
            AC-PCP-005 sub-case (c) is red until it runs — a skip here would report an unmeasured population as a pass. Fetch tags (fetch-depth: 0) and re-run.
panic: test timed out after 3m0s
	running tests:
		TestPreCommitNoSilentReplacement (0s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	181.353s
ok  	github.com/modu-ai/moai-adk/internal/template	2.304s
FAIL
```
원본 기록: tests-integration.log

## Gaps

601개 파일을 목록화하고 전체 내용 기반 참조·위험 API 검색을 수행했다. Go 463개는 AST 구문 분석, 셸 96개는 문법 검사 대상으로 삼았다. 의미·행동 심층 검증은 아래 11개 후보 및 실행한 시험 경로에 집중했다. 601개 파일의 모든 행·분기·호출 조합을 수동 검토하거나 실행했다는 뜻은 아니다. 저장소 전체 시험, 전체 race/coverage, Windows·Linux 실제 실행, 현재 Claude·Codex 앱의 이벤트 수용과 실제 사용자 자료 손실은 검증하지 않았다.

넓은 CLI 시험은 v3.1.2 태그가 없는 복제 환경에서 일부 실패하고 180초 제한에 도달했다. 제품 결함으로 귀속하지 않았다. 이후 집중 CLI 시험 통과는 이 전체 묶음 통과를 대신하지 않는다.

## Residual-risk

합성 입력과 명시된 함수/프로세스 경계에서 재현했다. 실제 장애 발생 횟수·사용자 영향 규모·수정 효과는 측정하지 않았다. 기존 시험 통과는 이번 경계 조건 결함의 부재를 증명하지 않는다.

## 우선순위

High: H01·H02·H03·H05·H06. Medium: H04·H07·H08·H09·H10·H11. 운영 코드는 이번 요청에 따라 조사만 했으며 수정하지 않았다.

## 검증 자료

inventory.json, registration.json, findings.json, source-integrity.json, probes-go-confirmed.log, probes-shell.json, probes-async.json, probes-async-control.log, probes-index.json. 첫 Go 탐침의 worktree 비교는 /var와 /private/var 차이로 실패해 EvalSymlinks 비교로 고친 뒤 재검증했다. 초기 기록도 probes-go.log에 남겼다.

## HTML 검증

1440/768/390px와 390px 오프라인에서 viewport == documentWidth. 중복 ID 0, 깨진 앵커 0, page errors 없음. browser-checks.json에 원본 측정값 보존. PDF는 Letter 18쪽으로 출력했으며 11개 항목과 Residual-risk 본문이 포함됨을 텍스트 추출로 확인했다. 인쇄 4쪽을 이미지로 확인했다.
