# SPEC-TODO-CARD-ISSUANCE-001 — 수용 기준

Tier L. 출시 차단으로 분류한 모든 기준은 `verification-completeness.md` §2 와 §2.1 에 따라 **RED-now 셀**(읽기 전용 단일 호출 명령, 그 명령의 verbatim stdout, 종료 코드, 고정된 트리)과 **green-path 셀**(어느 마일스톤이 뒤집는지와 통과했을 때의 출력)을 짝으로 가진다. 도착 시점에 붉을 수 없는 기준은 **회귀 가드**로 표시하고, 모드에 따라 적용 여부가 갈리는 기준은 **조건부**로 표시한다 — 둘 다 출시 차단으로 세지 않는다. Given-When-Then 은 검증 층의 형식이고 요구는 `spec.md` 의 GEARS 다. 모델 행위라서 기계 검사가 닿지 않는 조항은 *(doctrine-only)* 로 적는다.

**분류 개수.** 24개 = 출시 차단 20(-001~-006, -008, -010~-020, -022, -024) + 조건부 1(-021, Mode A 에서만 적용) + 회귀 가드 3(-007, -009, -023).

**문서 수준 핀.** 아래 원장의 모든 행은 트리 `68a4d813787c24f8dccdd0573b7d57a00653dfca`(`68a4d8137`)에서 이터레이션 5 가 다시 측정했다. 이 트리는 이터레이션 4 가 감사한 SPEC 커밋 `9e4307782aa8e66073ec8d5a04cd42229fc86114` 위에 develop `30ce3a02df92a70eac7307a9539f2000c1132de9` 를 흡수한 병합 커밋 하나를 얹은 것이다(브랜치 `WT-card-issuance-overlap-graph`). **이 핀 이동은 제품 경로를 바꾼다** — `git diff --shortstat` 가 이전 핀 `ad02a5677…` 과 이 핀 사이 `internal cmd pkg scripts .claude .codex` 의 912개 파일 변경을 읽는다(G30). 카드 t1399 가 `internal/kanban` 을 `internal/factory` 로(M8, `8a7d60ba2`), 디스패치 규칙 셋을 `kanban-dispatch*.md` 에서 `factory-dispatch*.md` 로(M10, `859477bc7`) 개명한 것이 그 일부다. 그래서 이터레이션 4 까지의 핀 사슬(`2de0a2cb6`→`1894984c3`→`ad02a5677`→`8fff427eb`)이 기대던 "제품 경로가 같으니 행의 SHA 를 그대로 둔다"는 근거(G20, G26)는 흡수에서 끊겼고, 이터레이션 5 는 어떤 행의 값도 이전 핀에서 가져오지 않았다. 명령에 SHA 가 박힌 행은 그 커밋의 이력을 읽고(그 SHA 는 모두 이 핀의 조상이라 이력은 흡수 뒤에도 같다), SHA 가 없는 `git grep`·`wc`·`cmp`·`go test -list` 행은 이 핀의 추적된 작업 트리를 읽는다. 측정 시점의 워킹 트리 변경은 이 SPEC 디렉터리 안의 미커밋 편집뿐이었고(이터레이션 5 시작 시 `git status --short` 는 빈 출력이었다), 모든 행이 추적된 경로나 커밋 객체만 읽으며 `git grep` 행은 SPEC 디렉터리 밖 경로로 한정돼 있다. 증거 행은 `HEAD`·`develop` 같은 움직이는 ref 대신 SHA 를 쓴다. 움직이는 ref 를 그대로 둔 곳은 G19 와 AC-TCI-020 의 게이트 판독, AC-TCI-020 Mode B·AC-TCI-021 (b) 의 `git merge-base develop HEAD` 뿐이다 — 이 질문들은 "지금 이 작업 트리에서 t1453 병합에 닿는가", "지금 흡수한 기준이 무엇인가"를 묻고 ref 가 움직여 답이 뒤집히는 것이 대상에 대한 참 신호이므로 핀하면 질문이 사라진다(`verification-completeness.md` §4 의 판별 질문). 그 이유는 해당 자리에서 다시 적는다. 핀이 따로 적히지 않은 기준은 이 핀에 묶인다. 판정 도구(`git`, `ls`, `wc`, `cmp`, `go test -list`)는 트리가 아니라 설치된 일반 명령이고, `moai` 바이너리는 RED 행에 쓰지 않았다(린트에만 `go build -o /tmp/moai-t1454 ./cmd/moai` 로 이 트리 `68a4d8137` 에서 만든 빌드를 **경로로** 호출했다 — 판정 빌드의 소스와 측정 트리가 같다; 이 세션은 레인 표지가 있어 `moai todo add` 가 거절된다).

**읽는 법 두 가지.**

- `ls` 의 stdout 모양은 셸과 별칭마다 다르다 — 긴 목록과 `total` 행·점 항목을 내는 셸이 있고 파일 이름만 내는 셸이 있다(이터레이션 2 저작자의 셸은 앞쪽을, 이터레이션 2 감사와 이터레이션 3 의 셸은 뒤쪽을 냈다). 그래서 `ls` 행은 stdout 모양이 아니라 **종료 코드**로 판정한다. 그리고 "디스크에 있다"와 "추적된다"는 다른 사실이다 — 이터레이션 1 의 L1(`ls .moai/reports/t1454`)은 감사 보고서가 그 디렉터리에 쓰이자마자 종료 0 이 되어 붉지 않게 됐고(G12), 같은 파일에 `git ls-files --error-unmatch` 는 여전히 종료 1 이다(G13). 파일이 **추적되는 형태로** 생기는지가 기준이면 `git ls-files --error-unmatch` 로 묻는다.
- green-path 셀은 통과했을 때의 출력이 **어떤 모양이 될 것인가**만 적는다. 실행 단계에서 관측한 값이 아니며, 시험 이름이 아직 없으므로 지금 실행하면 `[no tests to run]` 이다. 모양의 기준은 기존 시험에서 이 반복이 관측한 출력이다: `go test ./internal/factory -run '^TestClassifyCardText$|^TestNormalizeCardText$|^TestTokenSetJaccard$' -count=1 -v` 가 시험마다 들여쓰지 않은 `--- PASS: <이름> (…s)` 한 줄(하위 시험은 들여쓴 `--- PASS`)과 끝의 `PASS`, `ok  github.com/modu-ai/moai-adk/internal/factory  <시간>` 을 냈고, 같은 패턴의 `go test ./internal/factory -list '…'` 가 이름 3개를 냈다. 실행 단계의 시험 명령은 `plan.md` §D 의 SCRUB 복합 호출(레인 변수를 한 번의 호출에서 지움) 앞에 붙고, `-run` 패턴은 가지마다 `^이름$` 로 앵커하며 `^(A|B)$` 래퍼는 쓰지 않는다.

**상시 점화 규칙(모든 기준이 인용).** `verification-completeness.md` §1.2·§1.3 에 맞춰 기준마다 네 가지를 적는다. 공통 항은 여기서 한 번만 정의한다.

- **W(언제 돌아야 하나)**: 소유 마일스톤의 종료 시점과 sync-audit 시점. 마일스톤 이전에 돌리면 구조적으로 항상 초록이거나 항상 빨강이다.
- **V(붉음을 누가 보나)**: 마일스톤을 소유한 레인이 마일스톤 게이트에서 실패한 `go test` 종료 코드 1 로 본다. CI 는 같은 테스트를 푸시마다 실패한 패키지로 본다. `git grep -c -F` 형 기준은 빈 출력과 종료 코드 1, `git ls-files --error-unmatch` 형 기준은 stderr 의 `did not match any file(s) known to git` 과 종료 코드 1 이다.
- **S(멈춘 것을 독자가 어떻게 아나)**: 시험 이름 기준은 `go test <패키지> -list '<앵커된 패턴>'` 가 출력한 이름 수가 지정한 수와 같은지를 짝으로 본다. 이름이 바뀌거나 지워지면 짝의 `git grep -c -F` 가 비어 종료 코드 1 이고 `-list` 수가 모자란다. 이 짝은 마일스톤 종료 때 `progress.md` §E.2 에 기록되고 sync-audit 에서 다시 돈다.

각 기준 끝의 "점화" 줄은 W·V·S 를 가리키고, 붉어지는 **입력**(R)만 기준별로 적는다.

## 증거 원장 (RED-now 관측)

열: Id, 명령(단일 호출, 파이프·리디렉션·`&&`·`;`·서브셸 없음), 그 명령의 verbatim stdout, 종료 코드(별도 필드), 붉은 이유. stdout 이 비면 `(empty)`; `git ls-files --error-unmatch` 와 `ls` 가 실패하면 오류 문장은 stderr 로 나가므로 따로 적는다. 행 id 는 `L<n>`(RED-now), `C<n>`(대조군), `G<n>`(맥락·가드)다. `L1`, `L21`, `L22`, `L23`, `C8`, `C9`, `C11` 은 이터레이션 1 과 **명령이 다르다**(§ 이터레이션 2 변경 행); `L21`, `C8`, `C9`, `C19`~`C22`, `G9`, `G10`, `G12` 는 이터레이션 2 와 명령이나 기록이 다르다(§ 이터레이션 3 변경 행); 나머지 행은 같은 명령이 같은 출력을 냈다. `L37`~`L44`, `C23`~`C32`, `G19`~`G22` 는 이터레이션 3 이 새로 만든 행이다. `L45`, `C33`~`C40`, `G26`~`G29` 는 이터레이션 4 가 새로 만든 행이고, `C23`~`C26`, `C29`~`C32`, `G20` 은 이터레이션 4 에서 명령이 다르다(§ 이터레이션 4 변경 행). `L46`, `C41`~`C44`, `G30` 은 이터레이션 5 가 새로 만든 행이고, 이터레이션 5 가 명령이나 값을 바꾼 행(`L21`, `L24`, `C8`, `C9`, `C16`, `C21`, `C23`, `C24`, `C30`, `C33`~`C35`, `G1`, `G4`~`G7`, `G9`, `G10`, `G19`, `G24`, `G25`, `G27`, `G28`)은 `이터레이션 5 변경 행` 표에 있다. 이터레이션 4 까지 `internal/kanban` 을 읽던 행의 경로는 `internal/factory` 로 바뀌었다 — 개명 전 경로를 읽는 행은 파일이 없어 `git grep` 이 **빈 출력·종료 1 로 거짓 초록**이 되므로(이터레이션 4 의 원장을 개명 없이 이 핀에서 돌려 관측했다: 대조군 `C2`·`C4` 는 불일치했고 `L3`·`L8`·`L9` 등 같은 경로의 RED 행은 그대로 일치했다 — 그 일치가 곧 거짓 초록이었다) 각 RED 행은 대조군(`C2`·`C4`·`C5`·`C6`·`C15`)이 같은 경로에서 적중해야 읽힌다.

| Id | 명령 | verbatim stdout | 종료 | 붉은 이유 |
|---|---|---|---|---|
| L1 | `git ls-files --error-unmatch .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md` | (empty); stderr `error: pathspec '.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md' did not match any file(s) known to git` 다음 줄 `Did you forget to 'git add'?` | 1 | 추적되는 기준선 기록이 아직 없다 — M0 가 만든다. 감사가 쓰는 `.moai/reports/` 는 이 경로를 만들 수 없다 |
| L2 | `git grep -c -F "TestTodoAddPresentationStderrOnly" -- internal/cli` | (empty) | 1 | 제시 stderr 시험이 아직 없다 |
| L3 | `git grep -c -F "TestIssuanceNeighborsIncludeArchivedAndDropped" -- internal/factory` | (empty) | 1 | 보관·dropped 를 포함하는 이웃 조회와 시험이 아직 없다 |
| L4 | `git grep -c -F "TestTodoAddPresentationProbeOutsideLock" -- internal/cli` | (empty) | 1 | 락 밖 탐침 시험이 없다 |
| L5 | `git grep -c -F "TestTodoAddPresentationDryRunWritesNothing" -- internal/cli` | (empty) | 1 | `--dry-run` 시험이 없다 |
| L6 | `git grep -c -F "dryRun" -- internal/cli/todo.go` | (empty) | 1 | `add` 인자 스캔에 `--dry-run` 이 없다 |
| L7 | `git grep -c -F "TestTodoAddMCPCarriesPresentation" -- internal/cli` | (empty) | 1 | MCP 가 제시를 싣는 시험이 없다 |
| L8 | `git grep -c -F "issuance:TEXT:0:NULL" -- internal/factory/backlog_schema_freeze_test.go` | (empty) | 1 | 동결 튜플 시험이 새 컬럼을 아직 고정하지 않는다 |
| L9 | `git grep -c -F "TestBacklogIssuanceArchiveRestoreRoundTrip" -- internal/factory` | (empty) | 1 | 보관·복원 왕복 시험이 없다 |
| L10 | `git grep -c -F "TestTodoDropStoresReason" -- internal/cli` | (empty) | 1 | drop 사유 속성 저장 시험이 없다 |
| L11 | `git grep -c -F "TestFindingDispositionRecordOnly" -- internal/factory` | (empty) | 1 | finding 처분과 그 시험이 없다 |
| L12 | `git grep -c -F "TestRelationOntologyMapsLegacyKinds" -- internal/factory` | (empty) | 1 | 일곱 종류 어휘와 매핑 시험이 없다 |
| L13 | `git grep -c -F "TestTodoRelateConstraintRefusalByteIdentical" -- internal/cli` | (empty) | 1 | `relate` 의 종류별 제약 거절 시험이 없다 |
| L14 | `git grep -c -F "TestRelationConstraintRefusesCycle" -- internal/factory` | (empty) | 1 | 순환 제약 시험이 없다 |
| L15 | `git grep -c -F "TestCardGTDResolverBothDirections" -- internal/factory` | (empty) | 1 | 카드↔GTD 해석기가 없다 |
| L16 | `git grep -c -F "TestTodoTraceTransitiveDeterministic" -- internal/cli` | (empty) | 1 | `trace` 동사와 시험이 없다 |
| L17 | `git grep -c -F "TestGraphCardFileEdgesDeterministic" -- internal/graph` | (empty) | 1 | card→file 간선 층과 시험이 없다 |
| L18 | `git grep -c -F "TestFactoryNextBundleSerialLane" -- internal/cli` | (empty) | 1 | 묶음 직렬 레인 시험이 없다 |
| L19 | `git grep -c -F "TestTodoMergeRecordsAndDrops" -- internal/cli` | (empty) | 1 | 병합 동사와 시험이 없다 |
| L20 | `git grep -c -F "TestFactoryAssignBundleHubChain" -- internal/cli` | (empty) | 1 | 허브 체인 시험이 없다 |
| L21 | `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9]' 68a4d813787c24f8dccdd0573b7d57a00653dfca` | `0` | 0 | t1453 제목의 병합이 이 SHA 에서 어느 부모 경로로도 닿지 않는다 — **M5 게이트가 닫혀 있다**(종료 코드는 0 이므로 출력의 수를 읽는다; id 뒤 `[^0-9]` 가 `t14530` 류를 거른다; 흡수 방향 병합을 가리는 둘째 판독은 AC-TCI-020 판독 1). 이 행은 핀한 SHA 에서 한 측정이고, 지금의 작업 트리를 묻는 게이트 판독은 AC-TCI-020 이 `HEAD` 로 읽는다 |
| L22 | `git ls-files --error-unmatch .claude/rules/moai/workflow/card-issuance.md` | (empty); stderr `error: pathspec '.claude/rules/moai/workflow/card-issuance.md' did not match any file(s) known to git` 다음 줄 `Did you forget to 'git add'?` | 1 | 발행 규칙 파일이 없다 |
| L23 | `git ls-files --error-unmatch .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/m5-draft/anchors.md` | (empty); stderr `error: pathspec '.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/m5-draft/anchors.md' did not match any file(s) known to git` 다음 줄 `Did you forget to 'git add'?` | 1 | M5 초안과 삽입 위치 앵커 파일이 없다 — Mode B 가 만든다 |
| L24 | `wc -m .claude/skills/moai/workflows/gtd.md` | `   41160 .claude/skills/moai/workflows/gtd.md` | 0 | `gtd.md` 가 40,000자 한도를 1,160자 넘는다(이터레이션 4 까지는 37자 — 흡수가 문서를 키웠다) |
| L25 | `git grep -c -F "TestTodoGraphViewRendersRelations" -- internal/web` | (empty) | 1 | 웹 그래프 보기와 시험이 없다 |
| L26 | `git grep -c -F "amendment_of" -- .moai/specs/SPEC-TODO-ANALYSIS-001/spec.md` | (empty) | 1 | 대상 SPEC 이 아직 개정 선언을 갖지 않는다 |
| L27 | `git grep -c -F "TestTodoMergeNeverInvokedByAnalysis" -- internal/cli` | (empty) | 1 | 분석 경로가 병합을 못 부른다는 시험이 없다 |
| L28 | `git grep -c -F "TestBacklogDispositionArchiveRestoreRoundTrip" -- internal/factory` | (empty) | 1 | finding 처분의 보관·복원 왕복 시험이 없다 |
| L29 | `git grep -c -F "disposition:TEXT:0:NULL" -- internal/factory/backlog_schema_freeze_test.go` | (empty) | 1 | finding 표 두 곳의 동결 튜플이 오늘 고정돼 있지 않다 |
| L30 | `git grep -c -F "TestHubFileListFromBaseline" -- internal/homestate` | (empty) | 1 | 출하 허브 목록과 기준선 사본의 대조 시험이 없다 |
| L31 | `git ls-files --error-unmatch internal/homestate/hub_files.txt` | (empty); stderr `error: pathspec 'internal/homestate/hub_files.txt' did not match any file(s) known to git` 다음 줄 `Did you forget to 'git add'?` | 1 | 출하 허브 목록 데이터 파일이 없다 |
| L32 | `git grep -c -F "TestTodoAddIssuanceFlagRefusals" -- internal/cli` | (empty) | 1 | `add` 발행 플래그 거절 시험이 없다 |
| L33 | `git grep -c -F "TestGraphCardFileEdgesSeeAbsorbedMerge" -- internal/graph` | (empty) | 1 | 두 번째 부모로만 닿는 카드 병합을 보는 시험이 없다 |
| L34 | `git ls-files --error-unmatch .claude/rules/local/card-issuance-thresholds.md` | (empty); stderr `error: pathspec '.claude/rules/local/card-issuance-thresholds.md' did not match any file(s) known to git` 다음 줄 `Did you forget to 'git add'?` | 1 | 수치를 싣는 로컬 전용 규칙이 없다(Mode A 가 만든다) |
| L35 | `git grep -c -F "TestQueueMergeCarriesIssuanceAndDisposition" -- internal/factory` | (empty) | 1 | 큐 병합 복사 경로의 새 컬럼 시험이 없다 |
| L36 | `git grep -c -F "card-issuance" -- .claude/rules/moai/workflow/factory-dispatch.md` | (empty) | 1 | 상시 로드 스텁이 새 규칙을 이름으로 가리키지 않는다(Mode A 가 더한다) |
| L37 | `git grep -c -F "TestIssuanceInFlightOverlapReportsSharedPath" -- internal/factory` | (empty) | 1 | 진행 중 레인 카드와 공유하는 경로를 항목으로 내는 양성 겹침 시험이 없다 |
| L38 | `git grep -c -F "TestIssuanceInFlightOverlapNoneIsMeasured" -- internal/factory` | (empty) | 1 | 입력이 있는 비교에서 겹침이 없을 때 `none`(측정됨)을 내는 시험이 없다 |
| L39 | `git grep -c -F "TestTodoAddPresentationShowsInFlightOverlap" -- internal/cli` | (empty) | 1 | `add` 의 stderr 가 겹침 줄을 싣는 시험이 없다 |
| L40 | `git grep -c -F "TestTodoRelateRefusesProjectionKinds" -- internal/cli` | (empty) | 1 | `relate` 가 `parent-of`·`follow-up-of`·`merged-into` 를 거절한다는 시험이 없다 |
| L41 | `git grep -c -F "TestBacklogIssuanceStoredAsNullWhenAbsent" -- internal/factory` | (empty) | 1 | 속성 없는 카드가 SQL NULL 로 저장된다는 시험이 없다 |
| L42 | `git grep -c -F "TestBacklogDispositionStoredAsNullWhenAbsent" -- internal/factory` | (empty) | 1 | 처분 없는 소견이 SQL NULL 로 저장된다는 시험이 없다 |
| L43 | `git grep -c -F "TestHubFileLoaderIgnoresProjectTree" -- internal/homestate` | (empty) | 1 | 적재의 동작 시험(프로젝트 트리에 독립)이 없다 |
| L44 | `git grep -c -F "TestHubFileMeasurementSkipPolicy" -- internal/homestate` | (empty) | 1 | 측정 시험의 건너뜀 정책 시험이 없다 |
| L45 | `git grep -c -F "TestTodoGraphViewDoesNotWaitOnQueueLock" -- internal/web` | (empty) | 1 | 큐 락을 쥔 채 그래프 보기가 제때 200 을 낸다는 시험이 없다 |
| L46 | `git log --diff-filter=A --format=%H -- .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md` | (empty) | 0 | 순서 증인 1 이 오늘 줄을 하나도 읽지 못한다(증인 1 은 정확히 한 줄 `B` 를 요구한다) — 기준선 커밋이 아직 없다. 종료 코드는 0 이므로 출력의 줄 수를 읽는다; 대조군 C41 |

**대조군**(같은 경로·같은 탐침이 존재하는 것을 맞히므로 위 빈 행이 측정된 부재임을 보인다).

| Id | 명령 | verbatim stdout | 종료 | 대조하는 행 |
|---|---|---|---|---|
| C1 | `git grep -c -F "TestTodoAddRefusesExactDuplicate" -- internal/cli` | `internal/cli/todo_analysis_add_test.go:2` | 0 | L2, L4~L5, L7, L10, L13, L16, L18~L20, L27, L32, L39, L40 |
| C2 | `git grep -c -F "TestClassifyCardText" -- internal/factory` | `internal/factory/backlog_analysis_test.go:2` | 0 | L3, L9, L11, L12, L14, L15, L28, L35, L37, L38, L41, L42 |
| C3 | `git grep -c -F "scan.pick" -- internal/cli/todo.go` | `internal/cli/todo.go:2` | 0 | L6 |
| C4 | `git grep -c -F "lease_expires_at:TEXT:0:NULL" -- internal/factory/backlog_schema_freeze_test.go` | `internal/factory/backlog_schema_freeze_test.go:2` | 0 | L8, L29 |
| C5 | `git grep -c -F "func TestBuild" -- internal/graph/graph_test.go` | `internal/graph/graph_test.go:3` | 0 | L17, L33 |
| C6 | `git grep -c -F "func Test" -- internal/web/todo_route_test.go` | `internal/web/todo_route_test.go:6` | 0 | L25, L45 |
| C7 | `git grep -c -F "TestSD_AC014_MCPMatchesCLIWithProjectRoot" -- internal/cli` | `internal/cli/factory_m3_test.go:1` | 0 | L7(MCP 패리티 시험은 이미 있다) |
| C8 | `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1448[^0-9]' 68a4d813787c24f8dccdd0573b7d57a00653dfca` | `1` | 0 | L21(선택자가 눈먼 것이 아니다) |
| C9 | `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1344[^0-9]' 68a4d813787c24f8dccdd0573b7d57a00653dfca` | `1` | 0 | L21(두 번째 양성 대조) |
| C10 | `git grep -c -F "amendment_of" -- .moai/specs/SPEC-DRIFT-CLOSE-BODY-001/spec.md` | `.moai/specs/SPEC-DRIFT-CLOSE-BODY-001/spec.md:1` | 0 | L26 |
| C11 | `git ls-files --error-unmatch .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/research.md` | `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/research.md` | 0 | L1, L23(같은 SPEC 디렉터리의 추적된 파일을 맞힌다) |
| C12 | `git ls-files --error-unmatch .claude/rules/moai/workflow/factory-dispatch-detail.md` | `.claude/rules/moai/workflow/factory-dispatch-detail.md` | 0 | L22 |
| C13 | `git ls-files --error-unmatch internal/homestate/factory.go` | `internal/homestate/factory.go` | 0 | L31 |
| C14 | `git ls-files --error-unmatch .claude/rules/local/gitflow-lane-protocol.md` | `.claude/rules/local/gitflow-lane-protocol.md` | 0 | L34 |
| C15 | `git grep -c -F "func Test" -- internal/homestate/fr_schema_test.go` | `internal/homestate/fr_schema_test.go:1` | 0 | L30, L43, L44 |
| C16 | `git grep -c -F "factory-dispatch-detail" -- .claude/rules/moai/workflow/factory-dispatch.md` | `.claude/rules/moai/workflow/factory-dispatch.md:14` | 0 | L36 |

**게이트 선택자의 눈먼 곳과 그 대조군**(AC-TCI-020 의 근거; 이터레이션 2 가 새로 만든 행). 이터레이션 1 의 게이트는 `--first-parent` 로 읽었다. 카드 브랜치가 `git merge develop` 으로 develop 을 흡수하면 develop 의 병합 커밋은 흡수 병합의 **두 번째 부모** 쪽에만 있어서 `--first-parent` 는 그것을 세지 못한다. 아래 고정 SHA 는 그것을 보인다 — `b05c3be90…` 은 브랜치 `WT-github-flow-default` 가 develop 을 흡수한 두 부모 병합이고(G11), 첫 부모 `5e31abbcb…` 에서는 닿지 않는 `merge(t1407)` 커밋 `46be0b8c8…` 이 두 번째 부모의 끝이다(C17·C18). 같은 선택자가 첫 부모 경로만 걸으면 0, 모든 부모를 걸으면 1 이다(C19·C20). 이 SHA 들은 미푸시 브랜치의 객체이므로 그 객체가 없는 클론에서는 C17~C20 과 흡수 방향 대조 행(C28·C29·C31)이 "미측정"이다 — 이식 가능한 런타임 대조는 C21·C22 의 부등식이다. C19~C22 의 선택자는 이터레이션 3 이 게이트와 같은 형태(id 뒤 `[^0-9]`)로 맞췄다.

| Id | 명령 | verbatim stdout | 종료 | 보이는 것 |
|---|---|---|---|---|
| C17 | `git merge-base --is-ancestor 46be0b8c87c76591cff7ec46b007571dafd74ca5 5e31abbcb3d09a03fc782bcfc842d8221b508dba` | (empty) | 1 | `merge(t1407)` 커밋은 흡수 병합의 첫 부모에서 닿지 않는다 |
| C18 | `git merge-base --is-ancestor 46be0b8c87c76591cff7ec46b007571dafd74ca5 b05c3be9049822851b4cc2088cd3fc011fe39899` | (empty) | 0 | 그러나 흡수 병합에서는 닿는다(두 번째 부모를 통해) |
| C19 | `git rev-list --merges --first-parent --count -E -i --grep='^merge[( :]+(card )?t1407[^0-9]' b05c3be9049822851b4cc2088cd3fc011fe39899` | `0` | 0 | `--first-parent` 선택자는 흡수된 병합을 못 본다 |
| C20 | `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1407[^0-9]' b05c3be9049822851b4cc2088cd3fc011fe39899` | `1` | 0 | 모든 부모를 걷는 선택자는 본다 |
| C21 | `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t[0-9]+' 68a4d813787c24f8dccdd0573b7d57a00653dfca` | `382` | 0 | 런타임 대조(이식 가능): 어느 부모 경로든 센 병합 수. 값은 읽는 SHA 와 함께 움직이고(게이트 판독 4 는 `HEAD` 로 읽는다) 기준은 C22 와의 **엄격한 부등식**이다 |
| C22 | `git rev-list --merges --first-parent --count -E -i --grep='^merge[( :]+(card )?t[0-9]+' 68a4d813787c24f8dccdd0573b7d57a00653dfca` | `214` | 0 | 같은 명령에 `--first-parent` 만 더한 값. 382 > 214 이므로 선택자가 첫 부모 경로 밖의 커밋을 본다 |

**게이트 선택자의 흡수 방향 대조군**(AC-TCI-020 판독 1·2 의 근거; 이터레이션 3 이 새로 만들고 이터레이션 4 가 둘째 질의 `A` 를 바꿨다). 이터레이션 2 의 선택자 `^merge[( :]+(card )?t1453` 는 제목 grep 이라 착지가 아닌 흡수 병합도 센다 — 미푸시 브랜치 `WT-github-flow-default` 가 이미 `merge(t1453): absorb … into WT-github-flow-default` 제목의 커밋 다섯 개를 갖고 있다(G23 이 흡수 병합 `b05c3be90…` 에서 닿는 다섯 제목을 낸다). 카드 브랜치가 그 미착지 브랜치를 병합하면(미착지 코드에 기대는 카드는 그 브랜치를 새 워크트리 안에서 병합한다 — `factory-dispatch.md` 의 새 카드 조항) 이 다섯이 HEAD 에서 닿으므로 `absorb` 를 가리지 않는 선택자는 착지 없이 5 를 읽는다(C28).

이터레이션 3 의 둘째 질의는 `--all-match --grep=absorb` 라서 **메시지 전체**에서 `absorb` 를 찾았다 — 병합 제목이 아니라 본문에 그 낱말이 있는 진짜 착지도 흡수 방향으로 세어 후보에서 뺐다. 이터레이션 3 감사가 이 트리에서 보였고(`merge(t1439)` 착지 `d53e6ca59…` 는 T=1, 옛 `A`=1, 후보 0 — 본문의 "absorbed tree" 한 구절 때문이다; C33·C35) 이터레이션 4 가 핀 `ad02a5677` 에서 쟀고(커밋 12,444개; `merge` 로 시작하고 카드 id 를 담은 333개 가운데 제목 줄에 `absorb` 가 든 것 76개, 본문에만 든 착지 28개, 옛 `A` 104 = 76 + 28, 제목 줄에 `absorb` 가 없는 257개 중 28개 = 10.9%) 이터레이션 5 가 같은 읽기를 두 핀에서 다시 돌려 이터레이션 4 의 값을 그대로 재현했고 현재 핀 `68a4d8137` 의 값을 얻었다: 커밋 12,998개; 카드 id 를 담은 병합 제목 382개 가운데 제목 줄에 `absorb` 가 든 것이 83개, 본문에만 든 착지가 31개이고(`python3` 로 `git log --format=%B` 를 줄 단위로 읽은 값 — 스크래치 스크립트이고 원장 행이 아니다), 옛 `A` 는 114(= 83 + 31)로 제목 줄에 `absorb` 가 없는 299개 중 31개(10.4%)를 흡수 방향으로 세어 뺐다(그 299개를 전부 착지라고 단정하지는 않는다 — 제목 모양만 읽은 수다). 새 `A` 는 `--grep` **하나**에 카드 id 와 `absorb` 를 같은 패턴으로 적는다 — git 은 `--grep` 정규식을 메시지의 줄 단위로 맞추므로 `.*` 가 줄바꿈을 건너지 못해 둘이 **같은 줄**에 있어야 한다: `^merge…t1453[^0-9].*absorb`. 병합 제목은 첫 줄이므로 본문에만 있는 `absorb` 는 세지 않는다(C34 는 0, 옛 형태 C35 는 1 — 같은 커밋 이력에서 두 형태가 갈린다). 아래 행은 같은 두 질의를 흡수 병합 `b05c3be90…` 에 걸어 첫 질의가 5 를 내고 새 둘째 질의도 5 를 내서 착지 후보(= 첫째 − 둘째)가 0 임을 보이고, 실제 착지 병합(t1448)에서는 둘째 질의가 0 이라 후보가 1 임을 보이며, 본문에 `absorb` 가 든 착지(t1439)에서도 후보가 1 임을 보인다.

| Id | 명령 | verbatim stdout | 종료 | 보이는 것 |
|---|---|---|---|---|
| C28 | `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9]' b05c3be9049822851b4cc2088cd3fc011fe39899` | `5` | 0 | 판독 1 의 첫 질의: 흡수 병합에서 닿는 t1453 제목 커밋 — 전부 흡수 방향이다 |
| C29 | `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9].*absorb' b05c3be9049822851b4cc2088cd3fc011fe39899` | `5` | 0 | 판독 1 의 둘째 질의(`--grep` 하나, 같은 줄): 그 가운데 병합 제목 줄에 `absorb` 를 담은 것 — 다섯 모두. 착지 후보 = C28 − C29 = 0 |
| C30 | `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1448[^0-9].*absorb' 68a4d813787c24f8dccdd0573b7d57a00653dfca` | `0` | 0 | 양성 대조: 실제 착지 병합(t1448)의 제목 줄은 `absorb` 를 담지 않는다 — 착지 후보 = C8 − C30 = 1 |
| C31 | `git rev-list --merges --no-walk --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9].*absorb' b05c3be9049822851b4cc2088cd3fc011fe39899` | `1` | 0 | 판독 2 의 점검(같은 줄 형태): 핀한 SHA 가 흡수 방향이면 1(그 SHA 는 착지가 아니다) |
| C32 | `git rev-list --merges --no-walk --count -E -i --grep='^merge[( :]+(card )?t1448[^0-9].*absorb' 4315f0d0e9742b4e3741eb20efc1614f7f395238` | `0` | 0 | 착지 병합(t1448 의 `4315f0d0e…`)은 0 |
| C33 | `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1439[^0-9]' 68a4d813787c24f8dccdd0573b7d57a00653dfca` | `1` | 0 | 본문에 `absorb` 가 든 착지의 양성 대조(t1439, 병합 `d53e6ca59…`, 제목 `merge(t1439): WT-aside-browser-cli into develop - …`): 판독 1(a) `T` = 1 |
| C34 | `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1439[^0-9].*absorb' 68a4d813787c24f8dccdd0573b7d57a00653dfca` | `0` | 0 | 같은 착지에 새 판독 1(b) 를 걸면 `A` = 0 — 제목 줄에 `absorb` 가 없다. 착지 후보 = C33 − C34 = 1 이라 게이트가 열린다 |
| C35 | `git rev-list --count -E -i --all-match --grep='^merge[( :]+(card )?t1439[^0-9]' --grep=absorb 68a4d813787c24f8dccdd0573b7d57a00653dfca` | `1` | 0 | 이터레이션 3 의 `A` 형태(메시지 전체)를 같은 착지에 걸면 1 — 후보 = C33 − C35 = 0 이라 진짜 착지에서 게이트가 닫힌 채로 읽힌다(D27 의 결함. 변이 MU-116) |
| C36 | `git rev-list --no-walk --count -E -i --grep=absorb d53e6ca592cba7584a25c5d10aa7e1b66cd011d9` | `1` | 0 | 이터레이션 3 의 판독 2 점검(`--grep=absorb`, 메시지 전체)을 같은 착지 커밋에 걸면 1 — 착지인데 "흡수 방향"으로 읽힌다. 본문의 "absorbed tree" 구절 때문이다 |

**순서 증인의 눈먼 곳과 그 대조군**(AC-TCI-001 순서 증인 3·5 의 근거; 이터레이션 3 이 새로 만들고 이터레이션 4 가 증인 3 의 형태와 제품 경로 필터를 바꿨다). 증인 3 의 두 질의는 범위 `B^..HEAD`·`B..HEAD` 만 보므로 B 보다 앞선 변경 커밋은 두 목록 어디에도 들어가지 않는다 — 이터레이션 2 감사가 카드 t1460(제품 커밋 세 개가 `1894984c3` 보다 앞선다)을 대역으로 같은 이력에서 보인 관측을 고정 SHA 로 다시 적는다. C25·C26 은 증인 3 형태(SHA 목록, `--full-history`)가 이 이력에서 빈 목록 둘을 내는 것(눈먼 곳), C24 는 증인 5 형태가 같은 이력에서 `3` 을 내는 것(앞선 제품 커밋을 본다), C23 은 이 카드의 plan 커밋 앞에 `t1454` 제품 커밋이 없다는 오늘의 값(증인 5 의 초록 모양), C27 은 경로 필터가 제품 커밋 정의의 일부임을 보인다. 제품 경로 필터는 이터레이션 3 의 `internal cmd .claude` 에서 `internal cmd pkg scripts .claude .codex` 로 넓혔다(D33 — 이 저장소가 `pkg/`·`scripts/` 를 추적하고, 이 SPEC 의 M5 가 루트 `.codex/` 를 고친다); 넓힌 필터에서도 C23·C24·G20 의 값은 같다.

**증인 3 의 옛 형태가 틀리는 곳(이터레이션 3 감사 D26)과 새 형태의 대조군.** 옛 증인 3 은 두 질의의 **개수**가 같으면 통과시켰다. 개수의 일치는 목록의 일치가 아니다 — 기준선 B 가 main 에, 제품 커밋 P 가 B 와 무관한 곁가지에 있고 곁가지를 B 뒤에 병합하면서 병합 메시지에 카드 id 를 적으면(이 저장소의 병합 메시지 규약이 그렇다), 첫 질의는 P 를, 둘째 질의(`--ancestry-path`)는 P 대신 병합 커밋을 세어 두 개수가 같은데 집합이 다르다. 반대 방향의 거짓 실패도 있다 — 정당한 이력(모든 커밋이 B 의 후손)인데 카드 id 를 적은 병합이 있으면 기본 이력 단순화가 그 병합을 첫 질의에서만 가려 개수가 갈린다. 아래 두 대조군은 **실제 이력**에서 그 거짓 실패를 보인다: 카드 t1460 의 이력(병합 `merge(t1460)`·`merge: absorb local develop (card t1460)` 가 카드 id 를 적는다)에서 B 대역 `e892b8c79…`(t1460 첫 제품 커밋 `b2c6af74c…` 의 부모)로 옛 형태는 `3` 대 `4`(C37·C38 — 정당한 이력인데 개수가 달라 옛 증인이 실패한다), 새 형태는 같은 네 SHA 두 목록(C39·C40)이다. 곁가지 반례는 스크래치 저장소에서 만들어 옛 형태와 새 형태를 함께 돌렸다(아래 "이터레이션 4 가 스크래치 저장소에서 실행한 순서 증인 탐침" 표).

| Id | 명령 | verbatim stdout | 종료 | 보이는 것 |
|---|---|---|---|---|
| C23 | `git rev-list --full-history --simplify-merges --count --grep=t1454 1894984c3254d62f2d57ced961fef8af63eb59c8^ -- internal cmd pkg scripts .claude .codex` | `0` | 0 | 증인 5 의 초록 모양: 이 카드의 첫 plan 커밋 앞에는 `t1454` 제품 커밋이 없다 |
| C24 | `git rev-list --full-history --simplify-merges --count --grep=t1460 1894984c3254d62f2d57ced961fef8af63eb59c8 -- internal cmd pkg scripts .claude .codex` | `3` | 0 | 증인 5 형태가 앞선 제품 커밋을 센다 — 변이 MU-94(기준선이 제품 커밋 뒤에 착지)가 내는 값의 모양(대역 카드 t1460; 이 질의는 0 이어야 하는데 3 이다) |
| C25 | `git rev-list --full-history --grep=t1460 1894984c3254d62f2d57ced961fef8af63eb59c8^..ad02a56779473afc2ef72e9ecfe16d517a297a7b -- internal cmd pkg scripts .claude .codex` | (empty) | 0 | 같은 이력을 증인 3 의 첫 질의 범위로 보면 SHA 목록이 비었다 — 앞선 제품 커밋 셋이 보이지 않는다(눈먼 곳). 종료 0 이고 빈 출력이다 |
| C26 | `git rev-list --full-history --ancestry-path --grep=t1460 1894984c3254d62f2d57ced961fef8af63eb59c8..ad02a56779473afc2ef72e9ecfe16d517a297a7b -- internal cmd pkg scripts .claude .codex` | (empty) | 0 | 증인 3 의 둘째 질의도 빈 목록 — 두 목록이 같아서(둘 다 빈 것) 목록 비교만으로는 증인 3 이 B 앞을 보지 못한다. 증인 3 은 "목록이 비어 있지 않다"를 따로 요구하므로 이 대역 이력에서는 그 요구가 먼저 실패한다(MU-94 에서는 B 뒤 커밋이 있어 비어 있지 않다) |
| C27 | `git rev-list --count --grep=t1454 1894984c3254d62f2d57ced961fef8af63eb59c8^` | `2` | 0 | 경로 필터가 없으면 `t1454` 를 메시지에 적은 다른 카드(SPEC-TODO-AUTO-PICK-001 의 plan 커밋 둘, `.moai/` 만 바꾼다)가 센다 — 제품 커밋은 카드 id 와 제품 경로 둘을 모두 만족하는 커밋이다 |

**맥락·가드 행**(RED 셀이 아니다 — 도착 시 초록이거나 정보용).

| Id | 명령 | verbatim stdout | 종료 | 쓰임 |
|---|---|---|---|---|
| G1 | `git diff --name-only 68a4d813787c24f8dccdd0573b7d57a00653dfca -- internal/factory/backlog_analysis.go` | (empty) | 0 | AC-TCI-007 의 도착 시 상태: 분류기 파일이 핀에서 변하지 않았다(파일은 t1399 M8 이 `internal/kanban/` 에서 `internal/factory/` 로 옮겼고 내용은 같다 — 개명 전 핀 `2de0a2cb6` 에는 이 경로가 없어 옛 형태의 비교는 쓸 수 없다) |
| G2 | `go test ./internal/cli -list '^TestTodoListJSON_GoldenByteIdentity$'` | `TestTodoListJSON_GoldenByteIdentity` · `ok  	github.com/modu-ai/moai-adk/internal/cli	1.468s` | 0 | AC-TCI-009 의 골든 시험이 존재하고 선택된다(선택 수 1; 이터레이션 1 은 같은 출력에 시간 `1.528s`, 이터레이션 2 는 `1.893s` 를 기록했다 — 시간은 실행마다 다르고 이름과 `ok` 가 판정이다) |
| G3 | `go test ./internal/factory -list '^TestClassifyCardText$'` | `TestClassifyCardText` · `ok  	github.com/modu-ai/moai-adk/internal/factory	0.502s` | 0 | AC-TCI-007 의 고정 시험이 선택된다(이터레이션 1 은 시간 `0.409s`, 이터레이션 2 는 `0.513s` 를 기록했다 — 이름과 `ok` 가 판정이다). 이터레이션 2 가 다중 가지 형태를 다시 관측했다 — factory `-list` 에 `^TestClassifyCardText$`·`^TestNormalizeCardText$`·`^TestTokenSetJaccard$` 세 가지 앵커를 수직선으로 이어 한 번에 주어 이름 3개와 `ok`(종료 0), cli `-list` 에 `^TestTodoAddRefusesExactDuplicate$`·`^TestTodoAddNearDuplicateRecordsOnly$`·`^TestSD_AC014_MCPMatchesCLIWithProjectRoot$`·`^TestTodoAdd_PrintsIDAndPosition$`·`^TestTodoListJSON_GoldenByteIdentity$` 다섯 가지를 한 번에 주어 이름 5개와 `ok`(종료 0) |
| G4 | `git cat-file -s 68a4d813787c24f8dccdd0573b7d57a00653dfca:.claude/rules/moai/workflow/factory-dispatch.md` | `27948` | 0 | AC-TCI-021 상시 로드 바이트 기준(핀; 이터레이션 4 까지는 개명 전 `kanban-dispatch.md` 28,301 바이트) |
| G5 | `wc -m .claude/rules/moai/workflow/factory-dispatch.md` | `   27744 .claude/rules/moai/workflow/factory-dispatch.md` | 0 | 같은 파일의 문자 기준(핀; 템플릿 사본은 27,423자) |
| G6 | `wc -m .claude/rules/moai/workflow/factory-dispatch-detail.md` | `   40659 .claude/rules/moai/workflow/factory-dispatch-detail.md` | 0 | 아직 40,000자 초과(659자) — 증가 금지 기준(핀; 이터레이션 4 까지는 43,138자) |
| G7 | `cmp .claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md` | `.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md differ: char 24809, line 181` | 1 | 분기된 쌍의 기존 분기(보존 대상 — 갈라진 줄은 로컬에만 `moai worktree sweep` 문장이 더 있는 한 줄이다) |
| G8 | `cmp .claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai/workflows/gtd.md` | (empty) | 0 | 바이트 동일 쌍(동일 유지 대상) |
| G9 | `git merge-base 30ce3a02df92a70eac7307a9539f2000c1132de9 68a4d813787c24f8dccdd0573b7d57a00653dfca` | `30ce3a02df92a70eac7307a9539f2000c1132de9` | 0 | 측정 시점의 `CARD_BASE` 를 핀한 SHA 쌍(`30ce3a02d…` 은 이 핀이 흡수한 develop 팁)에서 잰 값. 카드 브랜치가 develop 을 흡수했으므로 기준은 이제 흡수한 팁 자신이다(이터레이션 4 까지는 흡수 전이라 `2de0a2cb6`) |
| G10 | `git rev-list -n 1 --first-parent --grep='^merge(t1448)' 68a4d813787c24f8dccdd0573b7d57a00653dfca` | `4315f0d0e9742b4e3741eb20efc1614f7f395238` | 0 | t1448 병합(develop 의 first-parent 줄 위에 있다 — 통합 브랜치 쪽 이력을 읽는 이 행에는 `--first-parent` 가 맞다) |
| G11 | `git rev-list --parents -n 1 b05c3be9049822851b4cc2088cd3fc011fe39899` | `b05c3be9049822851b4cc2088cd3fc011fe39899 5e31abbcb3d09a03fc782bcfc842d8221b508dba 46be0b8c87c76591cff7ec46b007571dafd74ca5` | 0 | 고정 SHA `b05c3be90…` 이 두 부모 병합이다 |
| G12 | `ls .moai/reports/t1454` | (셸 의존 모양 — 이터레이션 3 의 셸은 `plan-audit-iter1.md` 와 `plan-audit-iter2.md` 두 파일 이름을, 이터레이션 4 의 셸은 `plan-audit-iter1.md`·`plan-audit-iter2.md`·`plan-audit-iter3.md`·`verdict.md` 네 이름을 냈다 — 디렉터리가 채워졌다) | 0 | 이터레이션 1 의 L1 이 더 이상 붉지 않다(감사가 이 디렉터리에 보고서를 썼다) — D3 의 실패 모습. 종료 코드로 판정한다 |
| G13 | `git ls-files --error-unmatch .moai/reports/t1454/plan-audit-iter1.md` | (empty); stderr `error: pathspec '.moai/reports/t1454/plan-audit-iter1.md' did not match any file(s) known to git` | 1 | 같은 파일은 디스크에 있지만 추적되지 않는다 |
| G14 | `git check-ignore -v .moai/reports/t1454/baseline/x.txt` | `.gitignore:235:.moai/reports/*	.moai/reports/t1454/baseline/x.txt` | 0 | 옛 기준선 경로는 무시 규칙에 걸린다 — 커밋 그래프가 증언할 수 없는 경로 |
| G15 | `git check-ignore -v .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md` | (empty) | 1 | 새 기준선 경로는 무시되지 않는다 |
| G16 | `git check-ignore -v internal/homestate/hub_files.txt` | (empty) | 1 | 출하 허브 목록 경로는 무시되지 않는다 |
| G17 | `git check-ignore -v .claude/rules/local/card-issuance-thresholds.md` | (empty) | 1 | 로컬 전용 규칙 경로는 무시되지 않는다(`.claude/rules/local/` 은 추적된다) |
| G18 | `git check-ignore -v .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/m5-draft/anchors.md` | (empty) | 1 | 새 초안 경로는 무시되지 않는다 |
| G19 | `git merge-base develop HEAD` | `30ce3a02df92a70eac7307a9539f2000c1132de9` | 0 | **움직이는 ref 를 그대로 둔 행**(이유: 이 질의는 "지금 로컬 develop 과 이 작업 트리가 마지막으로 만난 곳"을 묻고, 카드 브랜치가 develop 을 흡수하거나 develop 이 앞서가 답이 바뀌는 것이 대상에 대한 참 신호다 — 핀하면 `CARD_BASE` 라는 질문이 사라진다). 범위 판정식(AC-TCI-020 Mode B, AC-TCI-021 (b))은 이 값을 읽는 시점에 다시 구하며 이 행의 값은 이터레이션 5 의 관측이다(로컬 develop 은 그 뒤로 더 나아갔어도 병합 기준은 흡수한 팁에 머문다 — G9 와 같은 값) |
| G20 | `git diff --name-only 1894984c3254d62f2d57ced961fef8af63eb59c8 ad02a56779473afc2ef72e9ecfe16d517a297a7b -- internal cmd pkg scripts .claude .codex` | (empty) | 0 | 이터레이션 2 가 `1894984c3` 에서 읽은 코드 위치(`design.md` §4.2 의 줄 번호 등)가 이 핀 `ad02a5677` 에서도 그대로였다 — 두 커밋 사이에 제품 경로 변화가 없다. **역사 행**: 지금도 재현되지만 현재 핀 `68a4d8137` 의 근거는 아니다 — 흡수가 이 사슬을 끊었다(G30) |
| G23 | `git log --merges --format='%h %s' -E -i --grep='^merge[( :]+(card )?t1453[^0-9]' b05c3be9049822851b4cc2088cd3fc011fe39899` | 다섯 줄 — `b05c3be90 merge(t1453): absorb local develop 46be0b8c8 (t1407, t1448) into WT-github-flow-default`, `41b3e019f merge(t1453): absorb origin/develop e892b8c79 into WT-github-flow-default`, `eac566a61 merge(t1453): absorb origin/develop 1e2151a38 into WT-github-flow-default before M6`, `eda614195 merge(t1453): absorb develop bfcf2b8d3 into WT-github-flow-default before M2`, `0028671ed merge(t1453): absorb develop 7109e0900 into WT-github-flow-default before the run phase` | 0 | 미착지 브랜치의 t1453 제목 커밋 다섯이 모두 흡수 방향이다(C28·C29 의 다섯과 같다) |
| G26 | `git diff --name-only ad02a56779473afc2ef72e9ecfe16d517a297a7b 8fff427eb98a24e864cf41f1aeb2835aa0eb94ca -- internal cmd pkg scripts .claude .codex` | (empty) | 0 | 문서 수준 핀 `ad02a5677` 과 이터레이션 4 가 모든 행을 다시 돌린 트리 `8fff427eb` 사이에 제품 경로 변화가 없었다 — 이터레이션 4 가 기존 행의 SHA 를 그대로 둔 근거. **역사 행**: 이터레이션 5 의 핀 이동에는 이 근거가 없어(G30) 행의 SHA 를 갈았다 |
| G27 | `git grep -c -E -e 'Mutate\(' -e LockPath -e Flock -e acquireLock -e StateLock -- internal/web ':!internal/web/*_test.go'` | (empty) | 1 | AC-TCI-022 (b3) 의 어휘 검사가 오늘 거는 질의(이터레이션 4 의 토큰 `BoardLock` 은 t1399 M8·t1458 이후 없다 — 락 도우미가 `board_*.go` 에서 `state_lock*.go` 로 개명·재배치돼 `StateLock` 이 됐다): 웹 패키지의 시험이 아닌 소스에 락을 잡는 호출 토큰이 없다 — 이 기준의 한 항목은 도착 시 초록이다(RED 는 AC 전체를 두는 L25·L45 가 읽는다). 빈 출력은 G28·G29 와 짝으로만 읽는다 |
| G28 | `git grep -c -E -e 'Mutate\(' -e LockPath -e Flock -e acquireLock -e StateLock -- internal/factory/backlog_store.go` | `internal/factory/backlog_store.go:21` | 0 | G27 의 양성 대조: 같은 패턴이 락을 잡는 코드에서는 21행에 적중한다 — 패턴이 눈먼 것이 아니다 |
| G29 | `git ls-files --error-unmatch internal/web/app.go` | `internal/web/app.go` | 0 | G27 이 훑는 경로 집합이 비어 있지 않다(시험 파일은 C6 이 읽는다) — 빈 집합에서의 0 적중은 아무것도 말하지 않는다 |
| G30 | `git diff --shortstat ad02a56779473afc2ef72e9ecfe16d517a297a7b 68a4d813787c24f8dccdd0573b7d57a00653dfca -- internal cmd pkg scripts .claude .codex` | ` 912 files changed, 48231 insertions(+), 17299 deletions(-)` | 0 | 이터레이션 4 의 핀과 이 핀 사이에 제품 경로가 바뀌었다 — 이전 핀의 SHA 를 행에 그대로 둘 근거가 없다(G20·G26 은 흡수 앞쪽 사슬에서만 성립했다). 종료 코드는 0 이므로 출력의 수를 읽는다 |

**펜스 원장**(명령에 수직선이 들어 있어 표 셀에 쓰면 열이 깨지므로 verbatim 으로 펜스에 적는다).

```
G21  command: go test ./internal/template -list '^TestManifestHashFormat$|^TestCatalogHashCoversSkillSubfiles$|^TestGoldenCommittedArtifactsMatchEmission$|^TestDeclaredRuleMirrorForks$'
     stdout:  TestManifestHashFormat
              TestCatalogHashCoversSkillSubfiles
              TestDeclaredRuleMirrorForks
              ok  	github.com/modu-ai/moai-adk/internal/template	0.422s
     exit:    0   tree: 68a4d813787c24f8dccdd0573b7d57a00653dfca
     reads:   이름 3개 — 이터레이션 2 가 AC-TCI-021 (f) 에 적은 네 이름 가운데 `TestGoldenCommittedArtifactsMatchEmission` 은 이 패키지에 없다(D25 의 같은 원인)

G22  command: go test ./internal/template/agentemit -list '^TestGoldenCommittedArtifactsMatchEmission$'
     stdout:  TestGoldenCommittedArtifactsMatchEmission
              ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.401s
     exit:    0   tree: 68a4d813787c24f8dccdd0573b7d57a00653dfca
     reads:   이름 1개 — 그 시험은 `agentemit` 패키지에 있다

G31  command: go test ./internal/template/pluginemit -list '^TestGoldenCommittedArtifactsMatchEmission$|^TestCommittedVersionMatchesSSOT$'
     stdout:  TestGoldenCommittedArtifactsMatchEmission
              TestCommittedVersionMatchesSSOT
              ok  	github.com/modu-ai/moai-adk/internal/template/pluginemit	0.306s
     exit:    0   tree: 68a4d813787c24f8dccdd0573b7d57a00653dfca
     reads:   이름 2개 — 흡수가 생성 플러그인 페이로드를 더했다: `plugins/moai/skills/moai/workflows/gtd.md` 는 템플릿 `gtd.md` 의 바이트 사본이다(`cmp` 종료 0). M5 가 `gtd.md` 를 고치면 `make plugin-emit` 으로 이 사본도 같은 커밋에서 다시 만들어야 이 두 시험이 초록이다

G32  command: go test ./internal/cli -list '^TestAutoPickMirrorParity$'
     stdout:  TestAutoPickMirrorParity
              ok  	github.com/modu-ai/moai-adk/internal/cli	1.375s
     exit:    0   tree: 68a4d813787c24f8dccdd0573b7d57a00653dfca
     reads:   이름 1개 — 갈라진 쌍 `factory-dispatch.md` 가 "정확히 한 줄(로컬에만 `moai worktree sweep` 이 든 줄)만 다르다"는 것을 고정하는 시험이다(`todo_auto_pick_doc_test.go`): M5 의 두 사본 편집이 줄 수와 그 한 줄의 앞부분 접두사 관계를 깨면 붉다

G33  command: git grep -c -F "Analysis never folds one card into another" -- internal/cli/todo_auto_doc_test.go .claude/skills/moai/workflows/gtd.md
     stdout:  .claude/skills/moai/workflows/gtd.md:1
              internal/cli/todo_auto_doc_test.go:1
     exit:    0   tree: 68a4d813787c24f8dccdd0573b7d57a00653dfca
     reads:   M5 가 개정하려는 `gtd.md` 의 "Analysis never folds one card into another…" 문장(현재 `gtd.md:115`)은 `autoDocGTDProhibitions` 가 문장 전체를 고정한다 — 개정은 그 문장을 지우거나 고치지 않고 뒤에 예외 문장을 덧붙이는 형태여야 하거나 같은 커밋에서 그 시험의 리터럴을 함께 고쳐야 한다

G34  command: git grep -l -E 'gtd\.md' -- '*_test.go'
     stdout:  internal/cli/doc_json_shape_test.go
              internal/cli/todo_auto_doc_test.go
              internal/cli/todo_auto_pick_doc_test.go
              internal/cli/todo_classify_doc_parity_test.go
              internal/cli/todo_hold_doc_test.go
              internal/cli/todo_landed_doc_test.go
              internal/cli/todo_skill_doc_parity_test.go
              internal/cli/todo_skill_doc_test.go
              internal/cli/todo_test.go
              internal/template/backlog_json_disclosure_mirror_test.go
              internal/template/card_id_leak_test.go
              internal/template/gtd_canonical_surface_test.go
              internal/template/jev_auto_exception_test.go
              internal/template/workflow_rule_paths_pinned_test.go
     exit:    0   tree: 68a4d813787c24f8dccdd0573b7d57a00653dfca
     reads:   이 파일을 이름으로 대는 시험 파일 14개 — M5 의 `gtd.md` 압축·개정이 닿는 가드의 전체 집합(plan §F.8 단계 6 이 패키지별로 돌린다)

G35  command: git grep -l -E 'manager-todo\.md' -- '*_test.go'
     stdout:  internal/cli/todo_auto_doc_test.go
              internal/cli/todo_auto_pick_doc_test.go
              internal/mission/governor_test.go
              internal/settings/agentfm/agentfm_test.go
              internal/template/jev_auto_exception_test.go
              internal/web/agent_overrides_test.go
     exit:    0   tree: 68a4d813787c24f8dccdd0573b7d57a00653dfca
     reads:   6개

G36  command: git grep -l -E 'factory-dispatch(-detail|-mechanics)?\.md' -- '*_test.go'
     stdout:  internal/cli/init_headroom_export_test.go
              internal/cli/todo_auto_doc_test.go
              internal/cli/todo_auto_pick_doc_test.go
              internal/template/card_id_leak_test.go
              internal/template/codex_review_ownership_m4_test.go
              internal/template/contract_mode_guided_test.go
              internal/template/docs_delegation_lane_flow_test.go
              internal/template/jev_auto_exception_test.go
              internal/template/lane_recheck_doctrine_test.go
              internal/template/workflow_rule_paths_pinned_test.go
     exit:    0   tree: 68a4d813787c24f8dccdd0573b7d57a00653dfca
     reads:   10개 — 개명된 규칙 셋을 이름으로 대는 시험(`internal/template` 의 `contract_mode_guided_test.go` 의 Kickoff 분류 시험은 환경 변수 `MOAI_GR_BASE` 가 가리키는 기준 ref 의 문서를 읽고 그 변수가 없으면 건너뛴다 — 작업 트리의 새 `card-issuance.md` 를 훑지 않는다)

G37  command: git grep -l -E 'sync-auditor\.md' -- '*_test.go'
     stdout:  internal/cli/agentlint/agent_lint_test.go
              internal/cli/agentlint/linter_stale_test.go
              internal/cli/update_security_m2_test.go
              internal/homestate/fr_producer_test.go
              internal/template/agent_frontmatter_audit_test.go
              internal/template/audit_plan_doc_surface_test.go
              internal/template/cg_retirement_test.go
              internal/template/claude_audit_surface_test.go
              internal/template/closure_ac25_test.go
              internal/template/cross_model_audit_wiring_test.go
     exit:    0   tree: 68a4d813787c24f8dccdd0573b7d57a00653dfca
     reads:   10개 — 갈라진 쌍 `sync-auditor.md` 를 이름으로 대는 시험
```

**실제 이력의 증인 3 대조군 C37~C40**(AC-TCI-001 증인 3 의 옛 형태와 새 형태를 같은 이력에 건다; 결과가 여러 줄이라 펜스에 적는다). 대역 이력은 카드 t1460 이고, B 대역은 그 첫 제품 커밋 `b2c6af74c…` 의 부모 `e892b8c79b35ac2f30bcd11298dfe2fdca42f1b4`, HEAD 대역은 t1460 착지 병합 `2de0a2cb613b04765a1554f86685a3b48e0be806` 이다. 그 이력에는 카드 id 를 적은 병합 둘(`merge: absorb local develop (card t1460)`, `merge(t1460): …`)이 있고 모든 커밋이 B 대역의 후손이다 — 정당한 모양이다.

```
C37  command: git rev-list --count --grep=t1460 e892b8c79b35ac2f30bcd11298dfe2fdca42f1b4^..2de0a2cb613b04765a1554f86685a3b48e0be806 -- internal cmd pkg scripts .claude .codex
     stdout:  3
     exit:    0   tree: 2de0a2cb613b04765a1554f86685a3b48e0be806
     reads:   옛 증인 3 의 첫 질의 — 기본 이력 단순화가 카드 id 를 적은 병합 하나를 가린다

C38  command: git rev-list --count --ancestry-path --grep=t1460 e892b8c79b35ac2f30bcd11298dfe2fdca42f1b4..2de0a2cb613b04765a1554f86685a3b48e0be806 -- internal cmd pkg scripts .claude .codex
     stdout:  4
     exit:    0   tree: 2de0a2cb613b04765a1554f86685a3b48e0be806
     reads:   옛 증인 3 의 둘째 질의 — 3 대 4 라 옛 증인("같은 수")은 정당한 이력에서 거짓으로 실패한다

C39  command: git rev-list --full-history --grep=t1460 e892b8c79b35ac2f30bcd11298dfe2fdca42f1b4^..2de0a2cb613b04765a1554f86685a3b48e0be806 -- internal cmd pkg scripts .claude .codex
     stdout:  2de0a2cb613b04765a1554f86685a3b48e0be806
              a165d164d814158d213bc3cde9d2e1d56528f64f
              6ac7b0aae77718d08776418b1737c78ae791100d
              b2c6af74cac1bbf571d5e30fef983baebb51f85e
     exit:    0   tree: 2de0a2cb613b04765a1554f86685a3b48e0be806
     reads:   새 증인 3 의 첫 질의 — SHA 목록 넷(병합 둘과 제품 커밋 둘)

C40  command: git rev-list --full-history --ancestry-path --grep=t1460 e892b8c79b35ac2f30bcd11298dfe2fdca42f1b4..2de0a2cb613b04765a1554f86685a3b48e0be806 -- internal cmd pkg scripts .claude .codex
     stdout:  2de0a2cb613b04765a1554f86685a3b48e0be806
              a165d164d814158d213bc3cde9d2e1d56528f64f
              6ac7b0aae77718d08776418b1737c78ae791100d
              b2c6af74cac1bbf571d5e30fef983baebb51f85e
     exit:    0   tree: 2de0a2cb613b04765a1554f86685a3b48e0be806
     reads:   새 증인 3 의 둘째 질의 — 첫 질의와 같은 네 SHA, 같은 순서. 목록이 같고 비어 있지 않아 새 증인은 통과한다
```

**증인 5 의 형태를 이 카드의 실제 이력에 건 대조군 C41~C44**(이터레이션 5, 이터레이션 4 감사 D34 대응; C41 은 L46 의 대조군이다). 증인 5 에 `--full-history` 만 붙이면 이 카드의 실제 이력에서 **정당한 M0 가 실패한다** — 카드 브랜치가 M0 앞에서 develop 을 흡수한 병합 커밋 `68a4d8137` 이 메시지에 카드 id 를 담고 첫 부모 대비 제품 경로를 바꾸기 때문이다. 채택한 형태(`--full-history --simplify-merges`)는 같은 이력에서 `0` 이다. 기본형(옛 형태)도 `0` 이지만 스크래치 AUD1~AUD3 을 놓친다(아래 "이터레이션 5 가 스크래치 저장소에서 실행한 증인 5 탐침" 표).

| Id | 명령 | verbatim stdout | 종료 | 보이는 것 |
|---|---|---|---|---|
| C41 | `git log --diff-filter=A --format=%H -- .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/research.md` | `1894984c3254d62f2d57ced961fef8af63eb59c8` | 0 | L46 의 대조군: 증인 1 의 형태가 한 번 추가된 추적 파일에서는 정확히 한 줄을 읽는다 |
| C42 | `git rev-list --full-history --count --grep=t1454 68a4d813787c24f8dccdd0573b7d57a00653dfca -- internal cmd pkg scripts .claude .codex` | `1` | 0 | 증인 5 에 `--full-history` 만 붙인 형태는 이 이력에서 `1` 을 낸다 — 카드 id 를 담은 흡수 병합 `68a4d8137` 하나. 정당한 M0 가 이 형태에서 실패한다 |
| C43 | `git rev-list --full-history --simplify-merges --count --grep=t1454 68a4d813787c24f8dccdd0573b7d57a00653dfca -- internal cmd pkg scripts .claude .codex` | `0` | 0 | 채택한 증인 5 는 같은 이력에서 `0` — 기여한 선택된 커밋이 없는 흡수 병합을 덜어낸다 |
| C44 | `git rev-list --count --grep=t1454 68a4d813787c24f8dccdd0573b7d57a00653dfca -- internal cmd pkg scripts .claude .codex` | `0` | 0 | 옛(기본 단순화) 형태도 같은 이력에서 `0` — 거짓 실패는 없지만 곁가지에서 제품 커밋을 넣었다 되돌린 변이(AUD1~AUD3)를 놓친다 |

**게이트 선택자의 `--merges` 대조군 C45~C48**(이터레이션 6, 이터레이션 5 감사 D40 대응). `T`·`A` 에 `--merges` 를 더한 현재 형태를 실제 이력의 착지(t1448)·`HEAD`·`--merges` 없는 옛 형태에 다시 건다. 병합이 아닌 커밋의 본문 줄·제목 위조와 병합 본문 줄 인용은 실제 이력에 없으므로 스크래치 저장소 탐침(아래 "이터레이션 6 이 스크래치 저장소에서 실행한 게이트 선택자 탐침" 표)이 맡는다. t1439·t1344·t1407·흡수 병합 다섯과 `T`·`A` 쌍은 C8·C9·C20·C28~C34 가 같은 형태로 이미 쥔다.

| Id | 명령 | verbatim stdout | 종료 | 보이는 것 |
|---|---|---|---|---|
| C45 | `git rev-list --merges --no-walk --count -E -i --grep='^merge[( :]+(card )?t1448[^0-9]' 4315f0d0e9742b4e3741eb20efc1614f7f395238` | `1` | 0 | 실제 착지 병합(t1448, 제목 `merge(t1448): …`)에서 `T` = 1 — `--merges` 는 진짜 두 부모 병합을 떨어뜨리지 않는다(양성 대조; 같은 SHA 의 `A` 는 C32 = 0) |
| C46 | `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9]' c459477b5fd19c53525e82ef090ea2e2c0be5559` | `0` | 0 | 이 카드 브랜치의 현재 끝에서 `T` = 0 — 게이트가 닫혀 읽힌다(종료 코드는 0 이므로 출력의 수를 읽는다) |
| C47 | `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9].*absorb' c459477b5fd19c53525e82ef090ea2e2c0be5559` | `0` | 0 | 같은 끝에서 `A` = 0 — 착지 후보 = C46 − C47 = 0 |
| C48 | `git rev-list --count -E -i --grep='^merge[( :]+(card )?t[0-9]+' 68a4d813787c24f8dccdd0573b7d57a00653dfca` | `382` | 0 | `--merges` 없는 옛 형태가 C21 과 같은 382 — 이 이력에서 카드 id 를 담은 병합 제목은 전부 병합이라 `--merges` 가 떨어뜨리는 착지가 없다(비-병합 위조가 있었다면 이 값이 더 컸을 것이다) |

**환경 지우기 대조**(맥락 행 — 환경 변수를 지우는 것은 본래 복합 호출이라 단일 호출 형태 밖이고 RED 셀이 아니다; 형태는 `factory-dispatch.md` 의 env-isolated 검증 형태를 따른다). 팩토리 레인 세션(`MOAI_FACTORY_ROLE`·`MOAI_FACTORY_WORKER` 가 설정된 셸)에서 자기 환경을 지우지 않는 기존 시험 `TestRelateAndUnrelateRefusals` 를 두 방식으로 돌렸다.

```
G24  command: unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -run '^TestRelateAndUnrelateRefusals$' -count=1 -v
     stdout:  todo_relate_test.go:348: seed relate: moai relate: refused — lane boundary: a lane session cannot mutate the queue (read-only here: bare todo, list, history, show, why, pr, triage); a lane takes its next card through moai factory next
              --- FAIL: TestRelateAndUnrelateRefusals (1.16s)
              FAIL	github.com/modu-ai/moai-adk/internal/cli	2.930s
     exit:    1   tree: 68a4d813787c24f8dccdd0573b7d57a00653dfca
     reads:   `factory-dispatch.md` 의 env-isolated 형태(변수 셋)는 팩토리 레인에서 부족하다 — 붉은 이유가 시험 내용이 아니라 레인 가드다(wrong-reason red). 이 호출이 지운 셋 가운데 이 셸에 설정돼 있던 것은 `MOAI_KANBAN_ID`·`MOAI_KANBAN_SETTINGS_INJECTED` 둘뿐이었고, 레인 가드(`factoryLaneRefusal`, `factory_card.go:98-104`)가 읽는 `MOAI_FACTORY_ROLE`·`MOAI_FACTORY_WORKER`·`MOAI_KANBAN_BACKEND` 가 남아 거절했다

G25  command: unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -run '^TestRelateAndUnrelateRefusals$' -count=1 -v
     stdout:  --- PASS: TestRelateAndUnrelateRefusals (6.00s)  (하위 시험 여덟 개 모두 PASS)
              PASS
              ok  	github.com/modu-ai/moai-adk/internal/cli	7.385s
     exit:    0   tree: 68a4d813787c24f8dccdd0573b7d57a00653dfca
     reads:   plan §D 의 일곱 변수 SCRUB 와 같은 호출이다(이터레이션 4 까지의 아홉 변수 가운데 개명·삭제로 현재 환경 키에 없는 `MOAI_KANBAN`·`MOAI_KANBAN_LABEL` 은 뺐다 — `internal/config/envkeys.go`). 팩토리 레인 변수를 함께 지우면 같은 시험이 통과한다 — 오늘의 `relate` 가 쓰기 가능 집합 밖의 값(`near-duplicate`)·부재 카드·자기 관계·중복 기록을 거절하고 큐 다이제스트를 바꾸지 않는다는 기존 시험의 관측이며, AC-TCI-013 (f) 의 "오늘도 거절된다" 문단이 기대는 사실(세 종류 `parent-of`·`follow-up-of`·`merged-into` 자체는 이 시험이 걸지 않는다)
```

### 이터레이션 6 변경 행

| 행 | 이터레이션 5 | 이터레이션 6 에서 관측한 것 | 이유 |
|---|---|---|---|
| L21, C8, C9, C19~C22, C28~C34, G23 | `git rev-list`·`git log` 에 `--merges` 없음 | 명령에 `--merges` 를 더했다(C19·C22 는 `--merges --first-parent`, C31·C32 는 `--merges --no-walk`). 값과 종료 코드는 그대로다: L21 `0`, C8 `1`, C9 `1`, C19 `0`, C20 `1`, C21 `382`, C22 `214`, C28 `5`, C29 `5`, C30 `0`, C31 `1`, C32 `0`, C33 `1`, C34 `0`, G23 다섯 줄(전부 종료 0) | D40 — 병합이 아닌 커밋의 본문 줄·제목 위조를 거른다. 이 이력에서 카드 id 를 담은 병합 제목 382개는 전부 병합이라 떨어지는 착지가 없다(C48) |
| C35, C36 | 옛 `A` 형태(메시지 전체 `--grep`) | 변경 없음 — 옛 형태를 보이는 역사 대조군이라 옛 명령 그대로이고 값도 같다(`1`·`1`) | MU-116 의 형태를 보인다 |
| C45~C48 | 없음 | 새로 관측(`1` / `0` / `0` / `382`) | D40 — t1448 착지 병합의 `T`, 이 브랜치 끝의 `T`·`A`, `--merges` 없는 형태의 같은 값 |
| 증인 5 탐침의 AUD1 | 한 행 `FAIL(3)` | 두 행: AUD1(main 이 제품 경로로 갈라짐) `FAIL(3)`, AUD1b(제품 경로 밖) `FAIL(2)` | D41 — 두 구성을 다시 만들어 둘 다 재현했다 |

이터레이션 6 의 전수 실행: 표·펜스 원장에서 명령을 뽑아 `python3` 로 돌렸다 — 142건 가운데 120건이 문서의 stdout·종료 코드와 일치했고, 나머지는 변경 행 표의 비-원장 셀(G1·G9·L23 의 옛 행), 셸 의존 G12(이번 셸은 여섯 이름 `plan-audit-iter1.md`~`iter5.md`·`verdict.md` 를 냈다 — 종료 0 으로 판정), 직접 읽어 일치를 확인한 G23(다섯 줄 그대로), 건너뛴 17건(`go test`·`unset`·파이프 펜스·명령이 아닌 셀)이다. 재현되지 않아 회귀 가드로 내린 행: **없다**.

### 이터레이션 5 변경 행

| 행 | 이터레이션 4 | 이터레이션 5 에서 관측한 것 | 이유 |
|---|---|---|---|
| `internal/kanban` 경로를 읽는 모든 행(L3, L8, L9, L11, L12, L14, L15, L28, L35, L37, L38, L41, L42, C2, C4, G3, G28 등) | `internal/kanban/…` | `internal/factory/…` — 대조군 C2·C4·G3·G28 은 새 경로에서 적중한다 | t1399 M8(`8a7d60ba2`)의 패키지 개명. 개명 전 경로를 그대로 두면 RED 행이 거짓 초록이다 |
| C12, C16, G4~G7 | `kanban-dispatch*.md` | `factory-dispatch*.md`; C16 `…:15` → `…:14`, G4 `28301` → `27948`(바이트), G5 `28092` → `27744`, G6 `43138` → `40659`, G7 `char 25113` → `char 24809`(줄 181 은 같다) | t1399 M10(`859477bc7`)의 규칙 개명과 그 뒤의 내용 변화 |
| L21, C8, C9, C21, C22, C30, C33~C35, G10 | 핀 `ad02a5677` | 핀 `68a4d8137`: `0` / `1` / `1` / `382` / `214` / `0` / `1` / `0` / `1` / `4315f0d0e…` | 문서 수준 핀 이동 — 값은 C21(`333` → `382`)만 다르다 |
| L24 | `40037` | `41160` | 흡수가 `gtd.md` 를 키웠다(한도 초과 37자 → 1,160자) |
| C23, C24 | 기본 이력 단순화형 증인 5: `0` / `3` | `--full-history --simplify-merges` 형: `0` / `3` | D34 — 값은 같다 |
| G1 | `git diff --name-only 2de0a2cb6 -- internal/kanban/backlog_analysis.go` | `git diff --name-only 68a4d8137… -- internal/factory/backlog_analysis.go`: (empty) | 개명 — 개명 전 핀에는 새 경로가 없다 |
| G9, G19 | `2de0a2cb6…` | `30ce3a02d…` | 흡수 — 병합 기준이 흡수한 develop 팁이 됐다 |
| G20, G26 | 핀 사슬의 근거 행 | 여전히 재현되는 역사 행(빈 출력) — 현재 핀의 근거가 아니다 | 흡수가 사슬을 끊었다(G30) |
| G24, G25 | 변수 다섯 / 여섯 | 변수 셋 / 일곱 — 같은 결과(FAIL `lane boundary` / PASS) | `MOAI_KANBAN`·`MOAI_KANBAN_LABEL` 이 환경 키에서 사라졌고 디스패치 규칙의 env-isolated 형태가 셋으로 줄었다 |
| G27, G28 | 토큰 `BoardLock`, `…:15` | 토큰 `StateLock`, `…:21` | 락 도우미가 `board_*.go` 에서 `state_lock*.go` 로 개명·재배치됐다 |
| L46, C41~C44, G30~G33 | 없음 | 새로 관측 | D34(C42~C44), D35(L46, C41), D37(G32·G33 은 M5 가드), 흡수 대응(G30, G31) |

재현되지 않아 회귀 가드로 내린 행: **없다** — 개명 뒤 모든 행이 새 경로에서 문서의 값으로 재현된다.

### 이터레이션 4 변경 행

| 행 | 이터레이션 3 | 이터레이션 4 에서 관측한 것 | 이유 |
|---|---|---|---|
| C23, C24 | `-- internal cmd .claude`: `0` / `3` | `-- internal cmd pkg scripts .claude .codex`: `0` / `3` | 제품 경로 필터를 넓혔다(D33) — 값은 같다 |
| G20 | `-- internal cmd .claude`: (empty) | 넓힌 필터: (empty) | 같은 이유 |
| G12 | 파일 이름 둘, 종료 0 | 파일 이름 넷(`plan-audit-iter1.md`·`-iter2.md`·`-iter3.md`·`verdict.md`), 종료 0 | 감사 보고서와 판정서가 더 쓰였다 — 셸 의존 행이라 종료 코드로 판정한다(D3 의 같은 현상) |
| C25, C26 | `--count` 한 수: `0` / `0` | `--full-history` SHA 목록: (empty) / (empty), 종료 0 | 증인 3 이 개수 비교에서 SHA 목록 비교로 바뀌었다(D26) |
| C29 | `--all-match --grep=<카드> --grep=absorb`: `5` | 같은 줄 `--grep` 하나: `5` | 둘째 질의 `A` 를 메시지 전체가 아니라 병합 제목 줄에 맞췄다(D27) — 값은 같다 |
| C30 | 같은 옛 형태, t1448: `0` | 새 형태: `0` | 같은 이유 — 값은 같다 |
| C31, C32 | `--grep=absorb` 메시지 전체: `1` / `0` | 같은 줄 형태: `1` / `0` | 판독 2 의 점검을 같은 줄 형태로 맞췄다 — 값은 같다 |
| L45, C33~C40, G26~G29 | 없음 | 새로 관측 | D26(C37~C40), D27(C33~C36), D28(L45, G27~G29), D33(G26) 대응 |

옛 형태와 새 형태가 **실제 이력에서 갈리는 곳**: C35(옛 `A` 형태, t1439)는 `1`, C34(새 `A` 형태)는 `0` — 같은 착지에서 후보가 0 대 1; C37·C38(옛 증인 3)은 `3` 대 `4`, C39·C40(새 증인 3)은 같은 네 SHA.

### 이터레이션 3 변경 행

| 행 | 이터레이션 2 | 이터레이션 3 에서 관측한 것 | 이유 |
|---|---|---|---|
| L21, C8, C9 | `HEAD` 위에서, id 뒤 앵커 없음: `0` / `1` / `1` | 핀한 SHA `ad02a5677…` 위에서, id 뒤 `[^0-9]`: `0` / `1` / `1` | 움직이는 ref 제거와 `t14530` 류 차단(D20) |
| C19, C20 | id 뒤 앵커 없음: `0` / `1` | id 뒤 `[^0-9]`: `0` / `1` | 게이트와 같은 선택자 형태 |
| C21, C22 | `HEAD` 위에서 `333` / `214` | 핀한 SHA 위에서 `333` / `214` | 움직이는 ref 제거(값은 같다) |
| G9 | `git merge-base develop HEAD` → `2de0a2cb6…` | 핀한 SHA 쌍 → `2de0a2cb6…`; 움직이는 형태는 G19 로 이유와 함께 남겼다 | 증거 행의 `develop` 제거(§4 판별 질문) |
| G10 | `HEAD` 위에서 `4315f0d0e…` | 핀한 SHA 위에서 `4315f0d0e…` | 움직이는 ref 제거 |
| G12 | `total 48` 다음에 점 항목(이터레이션 2 저작자의 셸) | 셸 의존, 이터레이션 3 의 셸은 파일 이름 둘, 종료 0 | D24: 셸 의존 문장 정정 |
| L37~L44, C23~C32, G19~G23 | 없음 | 새로 관측 | D15·D16·D17·D18·D20·D23·D25 대응 |

### 이터레이션 2 변경 행

| 행 | 이터레이션 1 | 이터레이션 2 에서 관측한 것 | 이유 |
|---|---|---|---|
| L1 | `ls .moai/reports/t1454` → stderr `No such file or directory`, 종료 1 | **같은 명령이 종료 0 이 됐다**(G12). 새 L1 은 `git ls-files --error-unmatch …/baseline/baseline.md` → 종료 1 | 감사 export 가 그 디렉터리를 채운다(D3). 새 질의는 추적 상태를 묻고 어떤 감사 export 도 만들 수 없는 경로다 |
| L21, C8, C9 | `--first-parent` 형태: `0` / `1` / `1` | 어느 부모든 형태: `0` / `1` / `1` | `--first-parent` 는 카드 브랜치가 develop 을 흡수한 뒤 t1453 병합을 영영 못 센다(D2) — 통합 브랜치 자체를 읽는 G10 만 first-parent 로 남는다 |
| L22, C11(옛) | `ls …` | `git ls-files --error-unmatch …` 로 교체 | `ls` 출력 모양이 셸마다 다르고 추적 여부를 말하지 않는다 |
| L23 | `ls .moai/reports/t1454/m5-draft` | `git ls-files --error-unmatch …/m5-draft/anchors.md` | 초안도 무시되는 `.moai/reports/` 에서 추적되는 SPEC 디렉터리로 옮겼다 |
| L28~L36, C11~C22, G11~G18 | 없음 | 새로 관측 | D1·D2·D5·D6·D14 대응 |

기준선 재측정 요약은 `research.md` §3(figure 35개, git 쪽 헤드라인 전부 재현). 위 행은 SPEC 파일이 생기기 **전**의 코드·규칙 상태를 읽는다 — 이 SPEC 의 미커밋 편집은 `git grep` 이 닿는 경로와 `git ls-files` 가 묻는 경로를 바꾸지 않는다.

## AC-TCI-001 — 기준선이 추적되는 형태로 반입되고 재현 가능하며 자기 커밋이 앞선다 (REQ-TCI-001)

**Covers**: maps REQ-TCI-001

Release-blocking.

**읽는 시점은 둘이고 섞지 않는다**(이터레이션 5, 이터레이션 4 감사 D35 대응 — plan §F.2·§F.12, design §12.4, spec-compact, 완료 정의 #2 가 같은 문장으로 적는다). M0 종료 증거는 AC-TCI-001 (a)~(e) 다 — 기준선 파일과 그 안의 줄만 읽으므로 제품 커밋이 없는 트리에서 판정된다. 순서 증인 1~5 는 병합 전에 한 번만 읽는다 — 마지막 마일스톤이 끝난 뒤, 카드 브랜치가 통합 브랜치에 병합되기 전(sync-audit 이 증거를 읽는 시점). M0 종료에서는 읽지 않는다: 증인 3 이 B 뒤의 제품 커밋 목록이 비어 있지 않을 것을 요구하는데 M0 종료에는 제품 커밋이 하나도 없어 두 목록이 비기 때문이다(스크래치 S11). 병합 전 시점은 카드 브랜치가 아직 로컬에만 있는 때라, B 가 자기 커밋이 아니거나 제품 커밋 뒤에 착지했다는 위반이 이때 드러나도 커밋 순서를 다시 쓸 수 있다.

**Given** (a)~(e): M0 가 끝난 트리, `research.md` §3 의 figure 35개(GB 16, QB 10, SB 9). (f): 마지막 마일스톤이 끝난 병합 전 트리
**When** (a)~(e): `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/` 를 읽는다. (f): 아래 순서 증인 다섯을 읽는다
**Then** (a) `baseline.md` 가 **추적된다**(`git ls-files --error-unmatch` 가 경로를 출력하고 종료 0), (b) `figure:`·`command:`·`tree:` 줄의 수가 같고 35 이상이다(세 번의 `git grep -c -F` 가 같은 수를 낸다), (c) git 쪽 헤드라인 figure(GB02, GB05, GB06, GB10, GB13, GB14)를 같은 명령으로 다시 돌린 값이 기록과 같거나 각 차이가 `drift:` 줄에 설명돼 있다, (d) 끝의 "임계값" 표의 각 값 옆에 측정 명령이 있고 개수 figure(WT- 브랜치, 워크트리, SPEC 디렉터리, `card:` 를 가진 SPEC)마다 그 개수를 낸 명령이 바로 옆 `command:` 줄에 있으며 값은 읽은 시각을 달고 "측정 시각의 값"으로 적혀 현재 값처럼 적히지 않는다, (e) `hub-files.txt` 가 추적되고 머리 주석에 측정 명령과 `tree:` 가 있으며 각 경로 줄에 그 경로의 단일 호출 측정 명령이 기록돼 있다, (f) 기준선 커밋이 **자기 커밋**이고 어느 제품 변경 커밋보다 **앞선다** — 증인은 커밋 그래프이고 아래 명령 다섯이 **병합 전에 한 번만** 읽는다(완료 정의 #2).

순서 증인(완료 정의 #2 도 같은 명령을 인용한다). 기준선 커밋을 `B` 라 하고, `baseline.md` 머리의 `card-head:` 줄 — 측정을 시작한 시점의 카드 브랜치 끝 `git rev-parse HEAD` — 의 SHA 를 `X` 라 한다. `card-head:` 는 줄 이름에 `figure:`·`command:`·`tree:` 를 부분 문자열로 담지 않도록 정한 이름이라 (b) 의 세 개수에 끼지 않는다. figure 마다의 `tree:` 줄은 서로 다를 수 있다(git 이력 figure 는 흡수한 develop 팁, 나머지는 카드 브랜치 끝). **제품 커밋**은 "메시지에 카드 id `t1454` 를 담고 `internal`·`cmd`·`pkg`·`scripts`·`.claude`·`.codex` 아래 경로를 하나라도 바꾸는 커밋"이고(이 저장소는 `pkg/` 와 `scripts/` 를 추적한다 — D33. 루트 `.codex/` 는 `.gitignore:137` 로 추적 제외이고 추적 파일은 `.codex/config.toml` 하나뿐이다; M5 가 고치는 추적된 Codex TOML 은 `internal/template/templates/.codex/agents/moai/` 아래라 `internal` 이 이미 덮는다 — `.codex` 항목은 `config.toml` 을 위한 여유 경로이고 이터레이션 4 까지 "M5 가 루트 `.codex/` 를 고친다"고 적은 근거는 틀렸다), `.moai/` 아래만 바꾸는 SPEC·기준선 커밋은 제품 커밋이 아니라서 세지 않는다. 이 필터 밖의 추적 경로(`.agents`, `mods`, `Makefile`, `go.mod`)는 이 SPEC 의 계획이 고치지 않는 경로라 수용된 범위 밖이다(MU-110 과 같은 종류의 잔여).

1. `git log --diff-filter=A --format=%H -- .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md` 가 정확히 한 줄 `B` 를 낸다.
2. 자기 커밋: `git show --name-only --format= <B> -- . ':!.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline'` 이 빈 출력(종료 0)이다 — B 가 기준선 경로 밖의 경로를 건드리지 않는다. 양성 대조: 같은 형태를 기준선 경로 밖의 경로를 가진 커밋(`git show --name-only --format= 1894984c3 -- . ':!.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline'`, 이터레이션 3 이 이 형태를 같은 커밋에 걸어 SPEC 디렉터리 파일 여덟 경로를 관측했다, 종료 0)에 걸면 경로가 나열된다.
3. 뒤에서 본다(**SHA 목록 비교** — 개수 비교가 아니다): `git rev-list --full-history --grep=t1454 <B>^..HEAD -- internal cmd pkg scripts .claude .codex` 와 `git rev-list --full-history --ancestry-path --grep=t1454 <B>..HEAD -- internal cmd pkg scripts .claude .codex` 가 **같은 SHA 목록**(정렬한 뒤 줄 단위로 같다; 순서는 주장이 아니다)을 내고 그 목록이 **비어 있지 않다**. 다르면 B 의 후손이 아닌 곁가지에 제품 커밋이 있다 — 둘째 질의는 B 의 후손만 나열하므로 첫 질의에만 있는 SHA 가 곧 B 의 후손이 아닌 카드 id 제품 커밋이다. 양성 대조는 목록이 비어 있지 않다는 것 자체이고(빈 목록 둘은 같지만 질의가 아무것도 세지 않은 것이다), 실제 이력의 대조는 C37~C40 이다. **`--full-history` 가 필요한 이유**: 기본 이력 단순화는 병합이 한쪽 부모와 트리가 같으면 다른 쪽 부모를 따라가지 않아 곁가지의 제품 커밋을 첫 질의에서 통째로 가릴 수 있다(스크래치 S2d: 곁가지에서 제품 커밋을 넣었다 되돌리면 기본 단순화는 두 질의 모두 `M1` 하나만 낸다). **이 증인은 B 보다 앞선 제품 커밋을 보지 못한다** — `B^..HEAD` 는 `B^` 에서 닿는 모든 커밋을 제외하므로 B 앞의 제품 커밋은 두 목록에 들어가지 않고 두 목록은 같게 나온다(대조 C25·C26 이 같은 이력에서 빈 목록 둘을 낸다).
4. 측정 대상 트리: `git diff --name-only <X> <B>^ -- internal cmd pkg scripts .claude .codex` 가 빈 출력이다 — 기록이 측정한 트리가 B 의 부모와 제품 경로에서 같다. 한계: `X` 는 저작자가 적는 값이라 `X = B^` 로 적으면 이 증인은 통과한다 — 앞선 제품 커밋은 증인 5 가 잡는다.
5. 앞에서 본다: `git rev-list --full-history --simplify-merges --count --grep=t1454 <B>^ -- internal cmd pkg scripts .claude .codex` 가 `0` 이다. `B^` 에서 어느 부모 경로로든 닿는 커밋 중 제품 커밋을 세므로 0 이면 B 보다 앞선 제품 커밋이 없다(B 가 제품 변경 앞에 착지했다). **`--full-history` 가 필요한 이유(이터레이션 4 감사 D34)**: 기본 이력 단순화는 병합이 한쪽 부모와 트리가 같으면 다른 쪽 부모를 따라가지 않아, 곁가지에서 제품 커밋을 넣었다 되돌린 뒤 B **앞에서** 병합하면 그 커밋 둘을 통째로 가린다 — 증인 1~4 와 기본형 증인 5 가 모두 통과하는데 B 앞에 카드 id 제품 커밋이 둘 있다(스크래치 AUD1~AUD3: 기본형 `0`, 아래 표; 증인 3 이 같은 이유로 `--full-history` 를 쓰는 것의 앞쪽 짝이다). **`--simplify-merges` 가 필요한 이유(이터레이션 5)**: `--full-history` 만 쓰면 이 카드의 실제 이력에서 정당한 M0 가 실패한다 — 카드 브랜치가 M0 앞에서 develop 을 흡수한 병합 커밋 `68a4d8137` 이 메시지에 카드 id(`… (card t1454)`)를 담고 첫 부모 대비 제품 경로를 바꾸므로, `--full-history` 단독형은 그 병합 하나를 `1` 로 센다(C42). `--simplify-merges` 는 기여한 선택된 커밋이 없는 병합을 이력에서 덜어내 같은 이력에서 `0` 을 낸다(C43; 기본형은 C44 에서 `0`). 곁가지의 제품 커밋은 선택된 채 남으므로 AUD1~AUD3 은 `--simplify-merges` 아래에서도 센다. `--no-merges` 도 같은 거짓 실패를 피하지만 병합 커밋의 충돌 해결 안에 제품 변경을 담은 변이(스크래치 EVIL)를 `0` 으로 통과시켜 쓰지 않는다. 양성 대조: (i) 같은 질의 형태가 앞선 제품 커밋이 있는 이력에서 0 이 아님 — 카드 t1460 을 대역으로 `1894984c3` 에서 `3`(C24), (ii) 증인 3 의 "목록이 비어 있지 않다"가 같은 grep 과 경로 필터가 이 카드의 이력에서 제품 커밋을 센다는 것을 읽는 시점에 보인다, (iii) 경로 필터가 정의의 일부임: 필터 없이는 `t1454` 를 메시지에 적은 다른 카드의 SPEC plan 커밋 둘이 세어진다(C27).

- **RED-now:** (a)~(e): L1(종료 1; 대조군 C11). 경로 판정은 `ls` 가 아니라 `git ls-files --error-unmatch` 다 — 감사 export 는 `.moai/reports/` 에만 쓰이고 그 경로(G14)는 이 질의의 대상이 아니다. (f): L46(출력 없음 — 증인 1 이 오늘 줄을 하나도 읽지 못한다, 증인 1 은 정확히 한 줄 `B` 를 요구한다; 대조군 C41 은 같은 형태가 한 번 추가된 추적 파일에서 정확히 한 줄을 읽는다). 증인 2~5 는 B 가 생기기 전에는 명령 자체를 실행할 수 없어 RED 셀이 아니라, 형태의 눈먼 곳과 초록 모양을 대역·스크래치 이력에서 읽는다: 순서 증인 5 의 오늘 값은 대역 이력에서 0 이 아닌 쪽(C24)과 이 카드 이력에서 0 인 쪽(C23, C43)을 둘 다 관측했고, 같은 형태에서 `--full-history` 단독형이 이 카드 이력에서 `1` 을 내는 것(C42)이 `--simplify-merges` 를 붙인 이유이며, 순서 증인 3 의 새 형태는 실제 이력(카드 t1460 대역)에서 같은 네 SHA 두 목록(C39·C40)을, 옛 형태는 `3` 대 `4`(C37·C38)를 냈다.
- **Green path:** M0 종료 — `git ls-files --error-unmatch .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md` 가 그 경로를 한 줄 출력하고 종료 0; `git grep -c -F "figure:" -- .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md`, 같은 형태로 `command:`, `tree:` 가 `<경로>:N`(같은 N ≥ 35, 종료 0)을 출력한다((a)~(e); 순서 증인은 이 시점에 읽지 않는다). 병합 전 — 순서 증인 1~5 가 위 결과를 낸다(증인 3 은 같은 SHA 목록 둘이고 비어 있지 않으며, 증인 5 는 `0`). 통과한 모양일 뿐 관측값이 아니다.
- **Mutant probe:** (MU-1) 디렉터리만 만들고 비어 있음 → (a)(b) 실패. (MU-2) 이 문서의 숫자를 베끼되 명령이 없음 → `command:` 수가 0 이라 (b) 실패. (MU-3) 명령은 있으나 출력이 재현되지 않는 값 → (c) 가 git 쪽 값에서 잡는다. (MU-78) 기준선을 `.moai/reports/t1454/baseline/` 에만 쓴다 → 그 경로는 무시되므로(G14) (a) 의 `git ls-files --error-unmatch` 가 종료 1. (MU-79) 기준선 파일을 첫 M1 코드 변경과 같은 커밋에 넣는다 → 순서 증인 2 가 기준선 밖의 경로를 나열한다. (MU-94) 카드 `t1454` 의 RED 시험이나 코드 커밋을 먼저 착지시키고, 그 뒤에 `card-head:` 를 `B^` 로 적은 기준선 커밋 B 를 얹고, 이후 커밋을 더한다(증인 3 의 목록이 비어 있지 않다) → 증인 1~4 는 모두 통과한다(B 는 기준선만 담고, 앞선 커밋은 `B^..HEAD` 밖이며, `X = B^` 라 증인 4 가 비고, 두 목록이 같다) — 증인 5 만 `1` 이상을 내서 잡는다. 실행 관측: 대역 카드로 같은 형태가 `3` 을 낸다(C24)는 것과 증인 3 형태가 같은 이력에서 빈 목록 둘을 낸다는 것(C25·C26)을 고정 SHA 위에서 관측했다. (MU-114) 제품 커밋 P 를 B 와 무관한 곁가지에 두고(B 는 main 에) 곁가지를 B 뒤에 `merge(t1454): …` 로 병합한다 — 이터레이션 3 감사가 지적한 변이(D26) → **옛 증인 3(개수 비교)은 `2`·`2` 로 통과한다**(병합 커밋이 둘째 질의에, P 가 첫 질의에 셈해져 개수만 같다; 스크래치 S2b·S3b), **새 증인 3 은 첫 질의에만 있는 SHA 가 있어 실패한다**(`--full-history` 개수는 `3` 대 `2`). 병합 메시지에 카드 id 가 없으면 옛 형태도 `2` 대 `1` 로 잡는다(S2·S3) — 이 저장소는 병합 메시지에 카드 id 를 적는 것이 규약이라 변이가 정상 모양으로 나온다. (MU-115) 곁가지에서 제품 커밋을 넣었다 되돌리고(순효과 0) 병합한다 — 기본 이력 단순화 아래에서는 두 질의가 모두 `M1` 하나로 같아 옛 형태와 `--full-history` 없는 SHA 목록 비교가 모두 통과한다(S2d) → `--full-history` 형태만 `3` 대 `1` 로 잡는다. 둘 다 스크래치 저장소에서 실행해 관측했다(아래 표). (MU-123) 이터레이션 4 감사 D34 의 변이 — 증인 5 의 앞쪽 눈먼 곳: 곁가지에서 제품 커밋 P(`test: red t1454`, 제품 경로)와 그것을 지우는 둘째 커밋(`revert red t1454`)을 만들고 그 곁가지를 B **앞에서** `main` 에 병합한 뒤 B 를 얹는다 → 증인 1~4 는 통과하고(B 는 기준선만 담고, 두 목록은 B 뒤 커밋만 본다, `X = B^`), 기본 이력 단순화 증인 5 도 `0` 이다(병합이 한쪽 부모와 트리가 같아 곁가지를 따라가지 않는다) — **옛 형태는 통과한다**. `--full-history --simplify-merges` 형태는 곁가지의 제품 커밋 둘(과 병합)을 세어 `2`~`3` 으로 실패한다(스크래치 AUD1: 병합 메시지에 카드 id 가 있고 `main` 이 갈라진 경우 `3`, AUD2: id 가 없는 경우 `2`, AUD3: `main` 이 갈라지지 않은 경우 `2`; 이 카드 실제 이력 위의 클론에서 AUD1 형태는 기본형 `0`, 채택한 형태 `3`). 스크래치 저장소에서 실행해 관측했다(아래 표). 잔여(수용): 증인 3·5 는 메시지에 카드 id `t1454` 를 담은 제품 커밋만 센다 — 카드 id 가 없는 제품 변경 커밋은 이 두 증인이 잡지 못하고 증인 4 도 `X` 를 저작자가 적으므로 확실히 잡지 못한다(MU-110; 막는 것은 `plan.md` §F.12 의 "모든 커밋 메시지에 카드 id" 규율과 `gitflow-lane-protocol.md` §1 이고 기계 검사가 아니다). 다른 카드의 커밋이 메시지에 `t1454` 를 적고 제품 경로를 바꾸면 증인 5 가 거짓으로 실패한다(안전한 방향 — 저작자가 그 커밋을 읽고 판정한다). 잔여: 큐 쪽 값은 비공개 DB 위의 스냅숏이라 제3자가 재실행할 수 없다 — `queue-snapshot:` 줄로 귀속만 하고 재현 주장을 하지 않는다(수용된 잔여).
- **점화:** W 는 (a)~(e) 가 M0 종료, 순서 증인 1~5 가 병합 전 한 번(sync-audit 이 증거를 읽는 시점) — 같은 증인을 두 시점에 읽지 않는다, R 은 디렉터리 삭제·`command:` 없는 figure·무시되는 경로로의 이동·기준선과 코드가 한 커밋에 섞이는 변경, V·S 는 공통(+ `git ls-files` 종료 1).

## AC-TCI-002 — 제시는 stderr 로만 나가고 stdout 은 기계 줄 그대로다 (REQ-TCI-002)

**Covers**: maps REQ-TCI-002

Release-blocking.

**Given** 큐에 live 카드 L, `[DROPPED — <reason>] ` 접두사를 단 dropped 카드 D, 보관 카드 A 가 있고 셋 모두 새 본문 N 과 표시 하한 이상으로 겹친다. 그리고 진행 중 레인 카드 P(picked, 레인 `lane-1` 에 배정, 본문이 `internal/cli/todo.go` 를 이름으로 댄다)가 있고 N 도 같은 경로를 이름으로 댄다. 두 변형 입력: N2 는 어느 진행 중 카드도 공유하지 않는 경로만 대고, N3 는 경로를 하나도 대지 않는다
**When** `add "<N>"`, `add "<N2>"`, `add "<N3>"` 이 각각 실행된다(`go test` 도구, 레인 변수 제거)
**Then** (a) stdout 은 정확히 `t<id> <pos>\n`(바이트 비교), (b) stderr 에 이웃 줄이 3개 이하로 나오고 각 줄이 id·상태(`live|dropped|archived`)·점수·`measure=`·본문 앞 60자를 담으며, (c) D 의 줄은 접두사를 뺀 본문을 보이고 사유를 따로 밝히며, (d) L·D·A 의 id 가 모두 나타난다, (e) N 의 stderr 에 겹침 줄이 P 의 id·레인·공유 경로 `internal/cli/todo.go` 와 `measure=file-overlap` 을 담고, N2 에서는 겹침 줄도 `unmeasured` 줄도 없으며(입력이 있는 비교에서 겹침이 없음 — 줄을 쓰지 않는다), N3 에서는 이유를 담은 `unmeasured` 겹침 줄이 나온다.

필수 시험 2개: `TestTodoAddPresentationStderrOnly`, `TestTodoAddPresentationShowsInFlightOverlap`.

- **RED-now:** L2, L39(대조군 C1). 붉은 이유: 제시 경로와 그 시험이 아직 없다.
- **Green path:** M1 — `go test ./internal/cli -run '^TestTodoAddPresentationStderrOnly$|^TestTodoAddPresentationShowsInFlightOverlap$' -count=1 -v` 의 통과 출력은 들여쓰지 않은 `--- PASS:` 줄 2개와 `ok  github.com/modu-ai/moai-adk/internal/cli`; 짝 `go test ./internal/cli -list '<같은 패턴>'` 는 이름 2개.
- **Mutant probe:** (MU-4) 제시를 stdout 에 출력 → (a) 실패. (MU-5) live 만 조회 → (d) 에서 D·A 부재로 실패. (MU-6) 접두사를 벗기지 않음 → (c) 실패. (MU-7) 상위 4개를 출력 → 줄 수 단언 실패. (MU-95) 조회는 겹침 항목을 내지만 렌더러가 겹침 줄을 쓰지 않음 → (e) 의 N 단언이 실패한다(AC-TCI-003 만으로는 렌더 단계의 누락을 잡지 못한다).
- **점화:** W 는 M1 종료, R 은 stdout 에 한 줄이라도 더 쓰는 변경과 겹침 줄 렌더 변경, V·S 공통.

## AC-TCI-003 — 제시의 출처 네 가지와 읽기 전용 조회 (REQ-TCI-002, -006)

**Covers**: maps REQ-TCI-002, REQ-TCI-006

Release-blocking.

**Given** 출처별 고정 입력(보관·dropped 이웃, 경로를 담은 열린 카드, `status: completed` 인 SPEC 디렉터리 고정 입력과 `status: draft` 인 하나, 예상 파일이 없는 신규 카드)과 겹침 입력: 진행 중 레인 카드 셋 — L1(본문이 `internal/cli/todo.go` 를 이름으로 댄다), L2(본문에 경로가 없고, 그 레인 브랜치의 변경 파일을 돌려주는 주입된 시험 이음매가 `internal/web/todo_view.go` 를 돌려준다), L3(본문에 경로가 없고 이음매도 변경 파일을 돌려주지 않는다 — 입력이 없는 카드) — 와 후보 본문 넷: N1(L1·L2 의 두 경로를 모두 이름으로 댄다), N2(어느 진행 중 카드도 공유하지 않는 `internal/graph/graph.go` 만 댄다), N3(경로를 하나도 대지 않는다), N4(경로 비교 방식을 가리는 음성 입력 — `internal/web/todo.go` 는 L1 의 `internal/cli/todo.go` 와 파일 이름 `todo.go` 만 같고 디렉터리가 다르며, `internal/cli/todo.go.orig` 는 L1 의 경로 `internal/cli/todo.go` 를 문자열 접두사로 갖지만 같은 경로가 아니다)
**When** 조회가 실행된다
**Then** (a) 이웃은 live·dropped·archived 를 모두 후보로 삼고 dropped 접두사를 벗기며 상한 3·표시 하한을 지킨다, (b) 같은 구성요소의 열린 카드는 공유 구성요소 키(경로 앞 두 마디)가 있을 때만 나온다, (c) 완료 SPEC 조회는 `completed` 만 포함하고 `draft` 는 제외하며 `heuristic` 표지를 단다, (d) 예상 파일이 없는 신규 카드(N3)의 겹침 항목은 `unmeasured` 이고 `none` 이 아니다 — 비교할 진행 중 카드의 입력이 하나도 없는 경우도 같다, (e) 모든 항목이 `source` 와 `measure` 를 가진다, (f) **양성 겹침**: N1 은 겹침 항목 둘을 낸다 — 항목마다 진행 중 카드 id·레인·공유 경로를 담고 `measure=file-overlap` 이다(L1 과 `internal/cli/todo.go`, L2 와 `internal/web/todo_view.go`), (g) **측정된 없음**: N2 는 `none`(`measure=file-overlap`)이고 `unmeasured` 가 아니다 — 입력이 있는 진행 중 카드(L1·L2)를 실제로 비교했기 때문이다. 입력이 없는 L3 이 섞여 있어도 마찬가지다: 판정은 집합 단위라 입력이 없는 진행 중 카드는 비교에서 빠지고 `none` 으로도 `unmeasured` 줄로도 보고되지 않으며, N1 의 겹침 항목에도 L3 은 나오지 않는다. 비교할 입력이 하나도 없을 때만 `unmeasured` 다((d)), (h) **음성 겹침**: N4 는 L1·L2 어느 것과도 겹침 항목을 내지 않고 `none`(`measure=file-overlap`)이다 — 겹침은 정규화한 저장소 경로 **전체**의 일치이고 파일 이름(basename)이나 문자열 접두사의 일치가 아니다.

필수 시험 8개(green path 가 전부 포함, `-list` 수 8; (g)·(h) 의 L3·N4 입력은 새 시험 이름이 아니라 기존 `TestIssuanceInFlightOverlapNoneIsMeasured` 의 입력이다): `TestIssuanceNeighborsIncludeArchivedAndDropped`, `TestIssuanceNeighborsStripDropPrefix`, `TestIssuanceNeighborsLimitAndFloor`, `TestIssuanceSameComponentOpenCards`, `TestIssuanceCompletedSpecCoverage`, `TestIssuanceInFlightOverlapUnmeasured`, `TestIssuanceInFlightOverlapReportsSharedPath`, `TestIssuanceInFlightOverlapNoneIsMeasured`.

- **RED-now:** L3, L37, L38(대조군 C2). 붉은 이유: 이웃·겹침 조회와 그 시험이 아직 없다 — 특히 L37·L38 은 양성 겹침(f)과 측정된 없음(g)을 붙드는 시험이 없음을 읽는다(`unmeasured` 만 내는 구현이 (d)(e)를 통과해도 (f)에서 잡히도록 한 쌍이다).
- **Green path:** M1 — `go test ./internal/factory -run '^TestIssuanceNeighborsIncludeArchivedAndDropped$|^TestIssuanceNeighborsStripDropPrefix$|^TestIssuanceNeighborsLimitAndFloor$|^TestIssuanceSameComponentOpenCards$|^TestIssuanceCompletedSpecCoverage$|^TestIssuanceInFlightOverlapUnmeasured$|^TestIssuanceInFlightOverlapReportsSharedPath$|^TestIssuanceInFlightOverlapNoneIsMeasured$' -count=1 -v` 의 통과 출력은 들여쓰지 않은 `--- PASS:` 줄 8개(위 이름마다 하나)와 `ok  github.com/modu-ai/moai-adk/internal/factory`, `--- FAIL`·`--- SKIP` 없음; 짝 `go test ./internal/factory -list '<같은 패턴>'` 는 이름 8개.
- **Mutant probe:** (MU-8) 이웃을 `rec.Items` 에서만 뽑음 → (a) 에서 보관 후보 부재. (MU-9) 하한 무시 → 소음 고정 입력이 나와 실패. (MU-10) 완료 SPEC 조회가 draft 포함 → (c) 실패. (MU-11) 겹침을 비었을 때 `none` 으로 표기 → (d) 실패. (MU-12) 이웃 조회가 `ClassifyCardText` 호출 → AC-TCI-007 의 파일 불변 가드가 잡는다. (MU-96) 겹침 항목을 항상 `unmeasured` 로 내는 구현(입력이 있어도) → (d)(e) 와 AC-TCI-002 (a)~(d) 는 통과하지만 (f) 가 공유 경로 항목을 요구해 실패한다 — 이것이 감사가 지적한 변이다. (MU-97) 입력이 있으면 진행 중 카드 모두를 겹친다고 보고함 → (g) 의 N2 가 `none` 을 내지 않아 실패한다. (MU-98) 겹침 항목에 공유 경로나 레인이 빠짐 → (f) 의 이름 단언이 실패한다. (MU-119) 겹침을 파일 이름(basename)이나 경로 문자열 접두사로 비교함 → (f)(g) 의 양성·`none` 입력은 모두 통과하지만 (h) 의 N4 가 L1 과 겹침 항목을 내서 실패한다(D30; 코드가 없어 기준 문면에 대한 읽기 판정). 읽기 판정이 아니라 변이 코드를 만들어 돌려야 확정되는 변이들이며 코드가 아직 없으므로 이 반복은 기준 문면에 대해 판정했다(아래 "이터레이션 3 에서 실제로 실행한 변이 탐침" 참조).
- **점화:** W 는 M1 종료, R 은 후보 집합이 줄어드는 변경, V·S 공통.

## AC-TCI-004 — 제시는 admit 을 바꾸지 않고 락 밖에서 돈다 (REQ-TCI-003)

**Covers**: maps REQ-TCI-003

Release-blocking; **연언**이라 공허하지 않다.

**Given** 같은 큐와 같은 본문, 제시 출처 하나가 실패하거나 시간 상한을 넘는 주입(시험 이음매)
**When** `add` 가 실행된다
**Then** (a) 종료 코드와 큐 파일이 제시를 끈 실행과 같고(새 카드 행과 분석기가 원래 기록하던 소견만 다르다), (b) 정확 중복 거절은 여전히 유일한 거절이고 그때 큐 파일은 바이트 동일이며, (c) 제시 탐침이 대기하는 동안 다른 프로세스의 `Mutate` 가 끝난다(락을 쥐지 않는다), (d) 시간 상한을 넘은 항목은 `unmeasured (time bound)` 로 출력되고 admit 은 성공한다.

필수 시험 3개: `TestTodoAddPresentationProbeOutsideLock`, `TestTodoAddPresentationNeverBlocks`, `TestTodoAddPresentationTimeBound`.

- **RED-now:** L4(대조군 C1).
- **Green path:** M1 — `go test ./internal/cli -run '^TestTodoAddPresentationProbeOutsideLock$|^TestTodoAddPresentationNeverBlocks$|^TestTodoAddPresentationTimeBound$' -count=1 -v` 의 통과 출력은 들여쓰지 않은 `--- PASS:` 줄 3개와 `ok  github.com/modu-ai/moai-adk/internal/cli`(시간 상한 시험이 제한 시간 안에 끝난다); 짝 `go test ./internal/cli -list '<같은 패턴>'` 는 이름 3개.
- **Mutant probe:** (MU-13) 탐침을 `Mutate` 안으로 옮김 → (c) 에서 두 번째 쓰기가 막혀 실패. (MU-14) 탐침 오류를 add 오류로 전파 → (a) 실패. (MU-15) 상한 없음 → (d) 의 막힘 이음매에서 시험 시간 초과.
- **점화:** W 는 M1 종료, R 은 탐침이 락 안으로 들어가는 변경, V 는 시험 실패(시간 초과 포함), S 공통.

## AC-TCI-005 — `--dry-run` 은 아무것도 쓰지 않는다 (REQ-TCI-004)

**Covers**: maps REQ-TCI-004

Release-blocking.

**Given** 큐 파일 F 와 `meta.last_seq` 값
**When** `add --dry-run "<본문>"` 이 실행된다(정확 중복 본문 한 번 포함)
**Then** (a) 종료 0, (b) F 가 바이트 동일, (c) `last_seq` 불변(id 소비 없음), (d) stdout 에 `<id> <pos>` 줄이 없고 stderr 에 제시와 `dry-run: nothing was written` 줄이 있으며, (e) 정확 중복이면 `a real add would refuse: t<N> already holds this card` 를 출력하고 여전히 종료 0, (f) 알 수 없는 플래그는 기존처럼 `unknown flag` 로 거절된다(회귀).

필수 시험 2개: `TestTodoAddPresentationDryRunWritesNothing`, `TestTodoAddDryRunFlagParsed`(`scanTodoAddArgs` 가 `--dry-run` 을 아는 것과 폴스루가 플래그를 받지 않는 것 둘 다).

- **RED-now:** L5, L6(대조군 C1, C3).
- **Green path:** M1 — `go test ./internal/cli -run '^TestTodoAddPresentationDryRunWritesNothing$|^TestTodoAddDryRunFlagParsed$' -count=1 -v` 의 통과 출력은 들여쓰지 않은 `--- PASS:` 줄 2개와 `ok  github.com/modu-ai/moai-adk/internal/cli`; 짝 `go test ./internal/cli -list '<같은 패턴>'` 는 이름 2개; 짝 `git grep -c -F "dryRun" -- internal/cli/todo.go` 는 `internal/cli/todo.go:N`(N ≥ 1, 종료 0).
- **Mutant probe:** (MU-16) id 를 소비한 뒤 되돌림 → (c) 실패. (MU-17) dry-run 이 소견을 기록 → (b) 실패. (MU-18) 중복 본문에서 비제로 종료 → (e) 실패.
- **점화:** W 는 M1 종료, R 은 dry-run 경로에서 `Mutate` 를 부르는 변경, V·S 공통.

## AC-TCI-006 — MCP 와 engage 도 제시를 싣고 기존 계약은 그대로다 (REQ-TCI-005)

**Covers**: maps REQ-TCI-005

Release-blocking.

**Given** 이웃이 있는 큐, 그리고 이웃이 없는 빈 큐
**When** MCP `todo_add` 와 `gtd engage` 가 카드를 admit 한다
**Then** (a) MCP 결과 텍스트의 첫 줄은 `<id> <pos>` 이고 그 뒤 빈 줄과 제시가 이어지며, (b) 이웃이 없을 때 MCP 결과 텍스트는 CLI stdout 과 같고(기존 `TestSD_AC014_MCPMatchesCLIWithProjectRoot` 가 그대로 통과), (c) engage 는 제시를 stderr 로 내고 소견을 기록하지 않으며 거절하지 않는다.

필수 시험 3개(앞의 둘은 새 시험, 셋째는 이미 있는 패리티 시험): `TestTodoAddMCPCarriesPresentation`, `TestGTDEngagePresentationRecordsNothing`, `TestSD_AC014_MCPMatchesCLIWithProjectRoot`.

- **RED-now:** L7(대조군 C7: 패리티 시험은 이미 존재).
- **Green path:** M1 — `go test ./internal/cli -run '^TestTodoAddMCPCarriesPresentation$|^TestGTDEngagePresentationRecordsNothing$|^TestSD_AC014_MCPMatchesCLIWithProjectRoot$' -count=1 -v` 의 통과 출력은 들여쓰지 않은 `--- PASS:` 줄 3개와 `ok  github.com/modu-ai/moai-adk/internal/cli`; 짝 `go test ./internal/cli -list '<같은 패턴>'` 는 이름 3개(셋째는 오늘도 선택된다 — G3 의 다중 가지 관측과 같은 형태).
- **Mutant probe:** (MU-19) 제시를 첫 줄 앞에 둠 → (a) 실패. (MU-20) engage 가 `near-duplicate` 소견을 기록 → (c) 실패. (MU-21) MCP 가 제시가 비었을 때도 빈 줄을 덧붙임 → (b) 실패.
- **점화:** W 는 M1 종료, R 은 MCP 텍스트 형태 변경, V·S 공통.

## AC-TCI-007 — 분류기와 분석 동작은 변하지 않는다 (REQ-TCI-006)

**Covers**: maps REQ-TCI-006

**회귀 가드**(도착 시 초록이라 출시 차단이 아니다; 도착 시 상태는 G1).

**Given** 핀 `68a4d813787c24f8dccdd0573b7d57a00653dfca`(분류기 파일은 t1399 M8 이 `internal/kanban/` 에서 `internal/factory/` 로 옮겼고 내용은 같다)
**When** 작업이 끝난 뒤 `git diff --name-only 68a4d813787c24f8dccdd0573b7d57a00653dfca -- internal/factory/backlog_analysis.go` 를 읽고 고정 시험을 돌린다
**Then** (a) diff 출력이 비어 있다(파일 불변, 이 SPEC 은 같은 파일에 함수를 더하지 않는다), (b) `TestClassifyCardText`, `TestNormalizeCardText`, `TestTokenSetJaccard`, `TestTodoAddRefusesExactDuplicate`, `TestTodoAddNearDuplicateRecordsOnly`, `TestTodoAnalysisNeverReordersQueue` 가 각각 `-list` 로 선택되고 통과한다.

- **RED-now:** 해당 없음 — 도착 시 초록(G1, G3). 회귀 가드다.
- **Mutant probe:** (MU-22) `ClassifyCardText` 가 보관 카드를 포함하도록 바꿈 → (a) 의 diff 가 비지 않아 잡힌다. (MU-23) 같은 파일에 새 함수를 더함 → (a) 가 잡는다.
- **점화:** W 는 M1 종료와 sync-audit, R 은 그 파일의 어떤 수정, V 는 diff 출력, S 는 `-list` 짝. 핀은 재베이스 때 다시 잰다.

## AC-TCI-008 — 스키마는 가산적이고 카드 표와 finding 표 모두 이주·parity·보관 왕복이 지켜진다 (REQ-TCI-007, -008, -009)

**Covers**: maps REQ-TCI-007, REQ-TCI-008, REQ-TCI-009

Release-blocking.

**Given** 옛 컬럼 집합으로 만든 DB 고정 입력, 발행 속성을 가진 카드와 가지지 않은 카드, 처분을 가진 소견과 가지지 않은 소견
**When** 새 엔진이 열고, 순수 읽기가 읽고, 보관·복원이 오가고, 두 큐가 병합된다
**Then** (a) `items`·`archived_items` 가 `issuance` 를, `findings`·`archived_findings` 가 `disposition` 을 갖고 기존 행의 해시가 같으며, (b) 새 컬럼이 없는 DB 를 순수 읽기가 읽어도 네 표 모두에서 값은 없음이고 DB 파일 바이트가 같으며(DDL 없음), (c) 동결 튜플 시험이 `items`·`archived_items` 의 마지막 튜플로 `issuance:TEXT:0:NULL` 을, `findings`·`archived_findings` 의 마지막 튜플로 `disposition:TEXT:0:NULL` 을 고정하고(finding 표 두 곳은 오늘 열 튜플이 동결돼 있지 않고 표 이름만 고정돼 있으므로 이 변경이 새 고정을 더한다) 같은 커밋에서 시험 문자열을 갱신하며, (d) 보관 → 복원 뒤 카드의 `issuance` 와 각 소견의 `disposition` 이 같고, (e) parity 단언이 두 컬럼을 네 표 모두에서 포함하며, (f) 두 큐의 병합이 항목 복사 경로와 소견 복사·재매핑 경로(보관 소견 포함)에서 두 컬럼을 나른다, (g) **없음의 저장 형태**: 속성이 하나도 없는 새 카드는 `items` 에서 `issuance` 가 SQL NULL 로 읽히고(빈 JSON 객체 `{}` 나 빈 문자열이 아니다) 보관 뒤 `archived_items` 에서도 NULL 이며, 처분이 없는 소견은 `findings`·`archived_findings` 에서 `disposition` 이 NULL 이다 — 시험은 `list --json` 투영이 아니라 원시 열 값(`sql.NullString` 의 `Valid`)을 읽는다.

필수 시험 13개(새 시험 11 + 갱신되는 기존 동결 시험 2): `TestBacklogIssuanceColumnRetrofit`, `TestBacklogIssuancePureReaderNoDDL`, `TestBacklogIssuanceArchiveRestoreRoundTrip`, `TestBacklogParityCoversIssuance`, `TestBacklogDispositionColumnRetrofit`, `TestBacklogDispositionPureReaderNoDDL`, `TestBacklogDispositionArchiveRestoreRoundTrip`, `TestBacklogParityCoversDisposition`, `TestQueueMergeCarriesIssuanceAndDisposition`, `TestBacklogIssuanceStoredAsNullWhenAbsent`, `TestBacklogDispositionStoredAsNullWhenAbsent`, 그리고 기존 `TestTodoHistoryAddsNoSchemaChange`, `TestSchemaFreezeRecordsTransitionStamps`(새 튜플 기대로 고친다).

- **RED-now:** L8, L9, L28, L29, L35, L41, L42(대조군 C4, C2). 동결 튜플 리터럴은 시험이 컬럼을 고정하지 않으면 도착한 트리에서 빈다. L41·L42 는 "없음이 NULL 로 저장된다"를 붙드는 시험이 없음을 읽는다.
- **Green path:** M2 — `go test ./internal/factory -run '^TestBacklogIssuanceColumnRetrofit$|^TestBacklogIssuancePureReaderNoDDL$|^TestBacklogIssuanceArchiveRestoreRoundTrip$|^TestBacklogParityCoversIssuance$|^TestBacklogDispositionColumnRetrofit$|^TestBacklogDispositionPureReaderNoDDL$|^TestBacklogDispositionArchiveRestoreRoundTrip$|^TestBacklogParityCoversDisposition$|^TestQueueMergeCarriesIssuanceAndDisposition$|^TestBacklogIssuanceStoredAsNullWhenAbsent$|^TestBacklogDispositionStoredAsNullWhenAbsent$|^TestTodoHistoryAddsNoSchemaChange$|^TestSchemaFreezeRecordsTransitionStamps$' -count=1 -v` 의 통과 출력은 들여쓰지 않은 `--- PASS:` 줄 13개(위 이름마다 하나)와 `ok  github.com/modu-ai/moai-adk/internal/factory`; 짝 `go test ./internal/factory -list '<같은 패턴>'` 는 이름 13개; 짝 `git grep -c -F "issuance:TEXT:0:NULL" -- internal/factory/backlog_schema_freeze_test.go` 와 `git grep -c -F "disposition:TEXT:0:NULL" -- …` 는 각각 `<경로>:N`(N ≥ 2, 종료 0 — 두 표의 문자열).
- **Mutant probe:** (MU-24) 컬럼을 `items` 에만 추가 → 보관 왕복(d)과 (a) 가 `archived_items` 에서 실패. (MU-25) 컬럼을 v1→v2 재구성 목록에 넣음 → v1 DB 재구성 시험이 컬럼을 잃는다(기존 카드 t1310 순서 시험). (MU-26) retrofit 을 버전 조정보다 앞에서 실행 → 같은 시험이 잡는다. (MU-27) parity 가 새 컬럼을 무시 → (e) 실패. (MU-80) `disposition` 을 `findings` 에만 추가하고 `archived_findings` 에는 빠뜨림 → (d) 에서 보관 뒤 처분이 사라지고 (a) 가 실패한다. (MU-81) 순수 읽기가 `columnExpr` 없이 `disposition` 을 직접 SELECT → 옛 DB 에서 오류나 DDL 이 나 (b) 실패. (MU-82) 큐 병합 복사 경로가 두 컬럼을 떨어뜨림 → (f) 실패. (MU-83) 동결 시험의 finding 표 튜플을 고정하지 않음 → (c) 의 문자열 `git grep` 이 비어 실패. (MU-103) 속성이 없는 새 카드에 빈 속성 구조체를 직렬화해 `issuance` 에 `{}` 를 씀(Go 의 `omitempty` 는 구조체 값을 생략하지 않는다 — 이터레이션 3 이 독립 Go 프로그램으로 관측했다: 값 형 필드는 `{"id":"t1","issuance":{}}` 를, 포인터 형 필드는 `{"id":"t1"}` 를 냈다) → AC-TCI-009 의 `list --json` 골든은 투영이 `{}` 를 숨기면 통과하지만 (g) 의 원시 열 읽기가 NULL 이 아님을 보고 실패한다. (MU-104) 처분이 없는 소견에 빈 문자열을 `disposition` 에 씀 → (g) 실패.
- **점화:** W 는 M2 종료, R 은 컬럼 목록 변경, V·S 공통.

## AC-TCI-009 — `list --json` 골든은 바이트 동일하다 (REQ-TCI-007)

**Covers**: maps REQ-TCI-007

**회귀 가드**(G2 로 도착 시 초록: 시험이 존재하고 선택된다).

**Given** 새 속성을 하나도 갖지 않은 카드들
**When** `list --json` 이 렌더된다
**Then** 출력이 `internal/cli/testdata/ac_tst_012_golden.json` 과 바이트 동일하다(`TestTodoListJSON_GoldenByteIdentity`).

- **RED-now:** 해당 없음(도착 시 초록).
- **Mutant probe:** (MU-28) 새 필드를 `omitempty` 없이 직렬화 → 골든이 `"issuance":null` 로 달라져 실패.
- **점화:** W 는 M2 종료, R 은 직렬화 형태 변경, V·S 공통.

## AC-TCI-010 — drop 사유가 속성에 저장되고 닫힘 시각은 접근자로 읽힌다 (REQ-TCI-010)

**Covers**: maps REQ-TCI-010

Release-blocking.

**Given** 카드 둘, 하나는 drop 하고 하나는 `done` 한다. 스탬프가 없는 과거 카드 하나
**When** `moai todo drop <id> "<reason>"` 과 접근자가 실행된다
**Then** (a) 사유가 속성에 저장되고, (b) 텍스트 접두사 `[DROPPED — <reason>] ` 도 그대로 쓰이며(기존 독자가 불변), (c) 접근자가 보관 카드는 `archived_at`, dropped 카드는 `dropped_at` 에서 시각을 내고, (d) 스탬프가 없으면 "unknown" 을 낸다.

필수 시험 3개: cli 의 `TestTodoDropStoresReason`, `TestTodoDropKeepsTextPrefix` 와 factory 의 `TestCardClosedAtAccessor`(접근자는 순수 모형 층에 둔다).

- **RED-now:** L10(대조군 C1).
- **Green path:** M2 — `go test ./internal/cli -run '^TestTodoDropStoresReason$|^TestTodoDropKeepsTextPrefix$' -count=1 -v` 의 통과 출력은 들여쓰지 않은 `--- PASS:` 줄 2개와 `ok  github.com/modu-ai/moai-adk/internal/cli`(`-list` 이름 2개), 그리고 `go test ./internal/factory -run '^TestCardClosedAtAccessor$' -count=1 -v` 의 통과 출력은 `--- PASS: TestCardClosedAtAccessor ` 한 줄과 `ok  github.com/modu-ai/moai-adk/internal/factory`(`-list` 이름 1개).
- **Mutant probe:** (MU-29) 사유를 접두사에만 둠 → (a) 실패. (MU-30) 접두사를 제거 → (b) 실패. (MU-31) 스탬프 없는 카드에 0 시각을 냄 → (d) 실패.
- **점화:** W 는 M2 종료, R 은 drop 경로 변경, V·S 공통.

## AC-TCI-011 — finding 처분은 기록만 한다 (REQ-TCI-011)

**Covers**: maps REQ-TCI-011

Release-blocking.

**Given** 소견 하나와 카드 둘, 처분 동사
**When** 처분(accept·merge·reject)을 기록한다
**Then** (a) 처분이 소견에만 저장되고 카드 필드·소견의 관계·큐 순서·픽업 필터가 바이트 동일하며, (b) 기존 소견은 처분이 없고 표시도 그대로이며, (c) 알 수 없는 처분 값은 아무것도 쓰지 않고 거절된다.

필수 시험 3개: factory 의 `TestFindingDispositionRecordOnly`, `TestFindingDispositionAbsentForLegacy` 와 cli 의 `TestTodoRelateDispositionVerb`.

- **RED-now:** L11(대조군 C2).
- **Green path:** M2 — `go test ./internal/factory -run '^TestFindingDispositionRecordOnly$|^TestFindingDispositionAbsentForLegacy$' -count=1 -v` 의 통과 출력은 `--- PASS:` 줄 2개와 `ok  github.com/modu-ai/moai-adk/internal/factory`(`-list` 이름 2개), 그리고 `go test ./internal/cli -run '^TestTodoRelateDispositionVerb$' -count=1 -v` 의 통과 출력은 `--- PASS: TestTodoRelateDispositionVerb ` 한 줄과 `ok  github.com/modu-ai/moai-adk/internal/cli`(`-list` 이름 1개).
- **Mutant probe:** (MU-32) `merge` 처분이 카드를 drop → (a) 실패. (MU-33) `reject` 가 소견을 삭제 → (a) 실패. (MU-34) 기존 소견에 기본 처분을 채움 → (b) 실패.
- **점화:** W 는 M2 종료, R 은 처분 효과가 카드 필드를 건드리는 변경, V·S 공통.

## AC-TCI-012 — 일곱 종류 어휘와 읽기 시 매핑, 옛 독자는 불변이다 (REQ-TCI-012, -017)

**Covers**: maps REQ-TCI-012, REQ-TCI-017

Release-blocking.

**Given** 옛 소견 여덟 종류와 `gtd_relations` 아홉 종류를 담은 고정 입력
**When** 해석기가 읽고 옛 독자(`list`, `why`, `export`, 픽업 필터, 자동 순위의 near-duplicate 경로)가 렌더한다
**Then** (a) 읽기 결과는 일곱 종류만 쓰고 매핑 표(depends·blocks→blocks, near-duplicate·duplicate-forced→duplicates, replaces→supersedes, contains·absorbs·conflicts→relates-to 와 한정어)를 지키며, (b) 저장된 소견 행의 바이트가 읽기 전후 같고, (c) 옛 독자의 출력이 변경 전 골든과 같다.

필수 시험 2개: factory 의 `TestRelationOntologyMapsLegacyKinds` 와 cli 의 `TestRelationLegacyReadersByteIdentical`.

- **RED-now:** L12(대조군 C2).
- **Green path:** M3 — `go test ./internal/factory -run '^TestRelationOntologyMapsLegacyKinds$' -count=1 -v` 의 통과 출력은 `--- PASS: TestRelationOntologyMapsLegacyKinds ` 와 `ok  github.com/modu-ai/moai-adk/internal/factory`(`-list` 이름 1개), 그리고 `go test ./internal/cli -run '^TestRelationLegacyReadersByteIdentical$' -count=1 -v` 의 통과 출력은 `--- PASS: TestRelationLegacyReadersByteIdentical ` 와 `ok  github.com/modu-ai/moai-adk/internal/cli`(`-list` 이름 1개).
- **Mutant probe:** (MU-35) 매핑이 저장 행을 다시 씀(`replaces`→`supersedes`) → (b) 실패. (MU-36) `conflicts` 를 매핑에서 빠뜨림 → (a) 실패. (MU-37) 새 어휘가 옛 독자의 표시 문자열을 바꿈 → (c) 실패.
- **점화:** W 는 M3 종료, R 은 소견 읽기·렌더 변경, V·S 공통.

## AC-TCI-013 — 실제 쓰기 동사가 위반할 수 있는 제약은 거절하고 파일을 건드리지 않는다 (REQ-TCI-013)

**Covers**: maps REQ-TCI-013

Release-blocking.

이 기준이 다루는 위반은 **설계된 쓰기 동사가 실제로 만들 수 있는 것**뿐이다. `parent-of`·`follow-up-of` 는 `add` 시점 속성의 투영이고 `merged-into` 는 `todo merge` 만 쓰므로 관계 동사에는 그 종류의 순환이나 두 번째 부모를 만들 길이 없다(새 카드는 후손이 없고 부모 id 는 이미 존재하는 더 작은 id 라서 순환은 구조적으로 도달할 수 없다 — 시험하지 않고 이유를 `design.md` §5.3 에 적었다). 이 "길이 없다"는 `relate` 가 그 종류를 쓰지 않아야 성립하므로 (f) 가 그 전제를 붙든다. `merged-into` 의 거절은 AC-TCI-018 이 소유한다.

**Given** 큐 파일과 위반 쓰기들: `relate` 의 자기 간선, `blocks` 순환(기존 가드 유지), `supersedes` 순환, 같은 `duplicates` 쌍의 역순 재기록, 그리고 `relate` 가 쓰면 안 되는 종류 셋 — `relate tA parent-of tB`, `relate tA follow-up-of tB`, `relate tA merged-into tB`; `add` 의 `--parent` 두 번, `--origin` 두 번, 큐에 없는 `--parent` id, 닫힌 집합 밖의 `--origin`; 그리고 합법적 입력(보관 카드를 `--parent` 로 지목)
**When** 쓰기 동사가 실행된다
**Then** (a) `relate` 의 자기 간선·`blocks` 순환·`supersedes` 순환이 종류와 쌍을 이름으로 대는 메시지로 거절되고 큐 파일이 바이트 동일이며, (b) `duplicates`·`relates-to` 쌍은 정규화돼 역순 재기록이 두 번째 레코드를 만들지 않고, (c) `relate` 의 합법적 쓰기는 통과하며, (d) `add` 의 둘째 `--parent`·둘째 `--origin` 은 플래그 이름을 대는 메시지로 거절되고 큐 파일이 바이트 동일이며 id 를 소비하지 않고, (e) 큐에 없는 `--parent` id 와 닫힌 집합 밖의 `--origin` 은 거절되고 아무것도 쓰지 않지만 **보관된 카드를 `--parent` 로 지목하는 것은 통과한다**(닫힌 부모의 후속은 정당한 흐름이다), (f) `relate` 에 준 `parent-of`·`follow-up-of`·`merged-into` 는 각각 거절되고 메시지가 받은 종류를 이름으로 대며 큐 파일이 바이트 동일이고 소견이 하나도 늘지 않는다 — 그 종류는 `add` 속성의 투영이거나 `todo merge` 만 쓴다(REQ-TCI-013 마지막 절). 오늘의 `relate` 는 이 값들을 `--relation must be one of … (got "<값>")` 로 이미 거절한다(`internal/cli/todo_relate.go:63-65`; 쓰기 가능 집합은 `BacklogSemanticRelations` 의 여섯 이름) — (f) 는 M3 가 어휘를 `duplicates`·`supersedes`·`relates-to` 로 넓히면서 그 거절을 잃지 않는다는 것을 붙든다. 변이는 쓰기 가능 집합에 일곱 종류 어휘를 통째로 넣는 자연스러운 확장 실수다.

필수 시험 5개: factory 의 `TestRelationConstraintRefusesCycle`, `TestRelationConstraintSymmetricNormalization` 과 cli 의 `TestTodoRelateConstraintRefusalByteIdentical`, `TestTodoAddIssuanceFlagRefusals`, `TestTodoRelateRefusesProjectionKinds`(세 종류를 표로 돈다).

- **RED-now:** L13, L14, L32, L40(대조군 C1, C2). L40 은 세 종류 거절을 붙드는 시험이 없음을 읽는다(위 문단: 거절 자체는 오늘도 있으므로 붉은 이유는 시험의 부재다).
- **Green path:** (d)(e) 는 M2 가 `add` 플래그와 함께 뒤집고 (a)~(c)(f) 는 M3 가 뒤집는다 — 기준은 M3 가 끝나야 모두 초록이다. `go test ./internal/factory -run '^TestRelationConstraintRefusesCycle$|^TestRelationConstraintSymmetricNormalization$' -count=1 -v` 의 통과 출력은 `--- PASS:` 줄 2개와 `ok  github.com/modu-ai/moai-adk/internal/factory`(`-list` 이름 2개), 그리고 `go test ./internal/cli -run '^TestTodoRelateConstraintRefusalByteIdentical$|^TestTodoAddIssuanceFlagRefusals$|^TestTodoRelateRefusesProjectionKinds$' -count=1 -v` 의 통과 출력은 `--- PASS:` 줄 3개와 `ok  github.com/modu-ai/moai-adk/internal/cli`(`-list` 이름 3개).
- **Mutant probe:** (MU-38) 자기 간선 검사를 `blocks` 분기에만 둠 → `relate t5 relates-to t5` 가 통과해 (a) 실패. (MU-39) `add` 가 둘째 `--parent` 를 조용히 덮어씀(마지막이 이김) → (d) 실패. (MU-40) 거절 전에 일부를 씀 → 바이트 동일 단언 실패. (MU-84) `--parent` 존재 검사가 live 카드만 봄 → 보관 카드 지목이 거절돼 (e) 의 합법 입력이 실패. (MU-85) `supersedes` 순환을 검사하지 않음(`blocks` 만) → (a) 실패. (MU-99) `relate` 의 쓰기 가능 집합이 `parent-of` 를 받아 소견으로 기록함 → 감사가 지적한 변이: (a)~(e) 는 모두 통과하지만 (f) 가 `relate tA parent-of tB` 의 거절을 요구해 실패한다. (MU-100) 같은 방식으로 `follow-up-of` 를 받음 → (f) 실패. (MU-101) 같은 방식으로 `merged-into` 를 받음(접는 기록의 작성자가 `todo merge` 하나라는 불변식이 사라진다) → (f) 실패. (MU-102) 세 종류를 거절하되 메시지가 받은 종류를 대지 않거나 거절 전에 일부를 씀 → (f) 의 이름·바이트 동일 단언이 실패한다.
- **점화:** W 는 M2(add 플래그)와 M3(relate) 종료, R 은 쓰기 동사 검증 순서 변경, V·S 공통.

## AC-TCI-014 — 카드↔GTD 해석기 (REQ-TCI-014)

**Covers**: maps REQ-TCI-014

Release-blocking.

**Given** engage 된 GTD 항목(관계 보유)과 engage 되지 않은 카드(고정 입력, 라이브 `gtd_relations` 는 0행이라 고정 입력이 필요하다)
**When** 해석기가 양방향으로 읽는다
**Then** (a) `tN → gtd 항목 id → tN` 이 왕복하고, (b) engage 된 적 없는 카드는 GTD 쪽이 비며, (c) `todo why <tN>` 이 GTD 관계 줄을 `source` 표지와 함께 보이고, (d) `gtd_relations` 에 쓰는 새 경로가 없고 `moai gtd organize` 동작이 바이트 동일이다.

필수 시험 3개: factory 의 `TestCardGTDResolverBothDirections`, `TestCardGTDResolverAddsNoGTDWritePath` 와 cli 의 `TestTodoWhyShowsGTDRelations`.

- **RED-now:** L15(대조군 C2).
- **Green path:** M3 — `go test ./internal/factory -run '^TestCardGTDResolverBothDirections$|^TestCardGTDResolverAddsNoGTDWritePath$' -count=1 -v` 의 통과 출력은 `--- PASS:` 줄 2개와 `ok  github.com/modu-ai/moai-adk/internal/factory`(`-list` 이름 2개), 그리고 `go test ./internal/cli -run '^TestTodoWhyShowsGTDRelations$' -count=1 -v` 의 통과 출력은 `--- PASS: TestTodoWhyShowsGTDRelations ` 와 `ok  github.com/modu-ai/moai-adk/internal/cli`(`-list` 이름 1개).
- **Mutant probe:** (MU-41) 해석기가 `gtd_relations` 에 쓴다 → (d) 실패. (MU-42) 한 방향만 구현 → (a) 실패.
- **점화:** W 는 M3 종료, R 은 `gtd_items.card_id` 연결 변경, V·S 공통.

## AC-TCI-015 — `trace` 는 결정적이고 순환에서 끝나며 쓰지 않는다 (REQ-TCI-015)

**Covers**: maps REQ-TCI-015

Release-blocking.

**Given** `A parent-of B parent-of C` 와 `A blocks C`, 그리고 레거시 순환 데이터(`X blocks Y`, `Y blocks X`)
**When** `trace A [--kind parent-of] [--depth 1]` 와 `trace X` 가 실행된다
**Then** (a) 도달한 노드가 (깊이, 종류, id) 순으로 같은 출력을 두 번 내고, (b) `--depth 1` 은 B 까지만, `--kind parent-of` 는 `blocks` 간선을 따르지 않으며, (c) 순환 입력에서 끝난다, (d) 큐 파일이 바이트 동일, (e) 레인 세션에서 거절되지 않는다.

필수 시험 4개: `TestTodoTraceTransitiveDeterministic`, `TestTodoTraceBreaksCycles`, `TestTodoTraceLaneReadOnlyAllowed`, `TestTodoTraceWritesNothing`.

- **RED-now:** L16(대조군 C1).
- **Green path:** M3 — `go test ./internal/cli -run '^TestTodoTraceTransitiveDeterministic$|^TestTodoTraceBreaksCycles$|^TestTodoTraceLaneReadOnlyAllowed$|^TestTodoTraceWritesNothing$' -count=1 -v` 의 통과 출력은 들여쓰지 않은 `--- PASS:` 줄 4개와 `ok  github.com/modu-ai/moai-adk/internal/cli`; 짝 `go test ./internal/cli -list '<같은 패턴>'` 는 이름 4개.
- **Mutant probe:** (MU-43) 깊이 무시 → (b) 실패. (MU-44) 맵 순회 순서 의존 → (a) 의 이중 실행이 달라 실패. (MU-45) 방문 집합 없음 → (c) 시험 시간 초과.
- **점화:** W 는 M3 종료, R 은 탐색 순서·방문 처리 변경, V·S 공통.

## AC-TCI-016 — 그래프 간선은 커밋된 증거에서만 나오고 결정적이며 두 번째 부모로 닿는 병합도 본다 (REQ-TCI-016)

**Covers**: maps REQ-TCI-016

Release-blocking.

**Given** `merge(t5): …` 와 `Merge card t6 …` 제목의 병합이 있고 각각 파일을 바꾼 고정 git 저장소, 그 저장소에 develop 을 **흡수하는 병합**(두 번째 부모 쪽에만 `merge(t7): …` 가 닿는다)과 흡수 방향 병합 제목(`merge: absorb local develop (card t8)` 과 다른 통합 대상을 이름으로 대는 `Merge … into <other>` 형태), 그리고 큐 DB 가 있는/없는 두 환경
**When** `graph build` 가 두 번 돌고 `graph check` 가 읽는다
**Then** (a) `card-file` 간선(`Source: t5`, `Target: <경로>`)이 `(kind, source, target, line)` 순으로 나오고, (b) 같은 트리와 같은 도달 가능한 이력에서 두 빌드의 출력이 바이트 동일, (c) 큐 DB 의 유무가 출력을 바꾸지 않으며 미착지 카드 id·예상 파일·소견이 출력에 없고, (d) 새 카드 귀속 병합이 착지하면 `graph check` 가 stale 로 읽고, (e) 흡수 병합의 두 번째 부모로만 닿는 `merge(t7)` 의 간선이 나온다(첫 부모 경로만 걸으면 빠진다), (f) 흡수 방향 병합은 간선을 하나도 내지 않는다(귀속은 기존의 단일 귀속 규칙을 쓰며 `t8` 의 간선이 없다).

필수 시험 5개: `TestGraphCardFileEdgesDeterministic`, `TestGraphCardFileEdgesCarryNoQueueState`, `TestGraphCheckNoticesCardFileSource`, `TestGraphCardFileEdgesSeeAbsorbedMerge`, `TestGraphCardFileEdgesIgnoreAbsorbMerge`.

- **RED-now:** L17, L33(대조군 C5).
- **Green path:** M3 — `go test ./internal/graph -run '^TestGraphCardFileEdgesDeterministic$|^TestGraphCardFileEdgesCarryNoQueueState$|^TestGraphCheckNoticesCardFileSource$|^TestGraphCardFileEdgesSeeAbsorbedMerge$|^TestGraphCardFileEdgesIgnoreAbsorbMerge$' -count=1 -v` 의 통과 출력은 들여쓰지 않은 `--- PASS:` 줄 5개와 `ok  github.com/modu-ai/moai-adk/internal/graph`; 짝 `go test ./internal/graph -list '<같은 패턴>'` 는 이름 5개.
- **Mutant probe:** (MU-46) 큐에서 예상 파일을 읽어 간선에 넣음 → (c) 실패. (MU-47) 시각·정렬 불안정 → (b) 실패. (MU-48) 출처 지문 미등록 → (d) 실패. (MU-86) 첫 부모 경로만 걸음 → 흡수 병합 쪽 `merge(t7)` 의 간선이 빠져 (e) 실패. (MU-87) 귀속 규칙을 거치지 않는 자체 정규식으로 흡수 방향 병합을 카드 귀속으로 셈 → (f) 실패. 잔여(수용): PR 의 CI 가 쓰는 합성 병합 참조에서 같은 입력이 로컬과 같은 출력을 내는지는 측정하지 못했다(`spec.md` §G).
- **점화:** W 는 M3 종료와 `graph-freshness` CI, R 은 간선 층 입력·순회 변경, V·S 공통.

## AC-TCI-017 — 묶음은 한 레인에 직렬로 실린다 (REQ-TCI-018)

**Covers**: maps REQ-TCI-018

Release-blocking.

**Given** 팩토리 런과 레인 둘, 묶음 B 의 세 카드 `b1 b2 b3`(레인 `lane-1`)와 묶음이 아닌 대기 카드 `u1`
**When** `factory next` 가 반복된다
**Then** (a) `lane-1` 이 `b1` 을 임대하고, `b1` 이 병합되기 전에는 `b2` 를 받지 못하며(선행 미병합이 오류가 아니라 건너뜀), (b) `lane-2` 는 `b2`·`b3` 를 받지 못하고, (c) `b1` 이 로컬 병합된 뒤 `lane-1` 의 다음 임대가 `u1` 보다 `b2` 를 우선하며, (d) 묶음이 아닌 카드의 직렬 슬롯 의미와 선택 결과는 변경 전 골든과 같다.

필수 시험 3개: `TestFactoryNextBundleSerialLane`, `TestFactoryBundleKeepsSerialSlot`, `TestFactoryAssignBundleOrderGuard`.

- **RED-now:** L18(대조군 C1).
- **Green path:** M4 — `go test ./internal/cli -run '^TestFactoryNextBundleSerialLane$|^TestFactoryBundleKeepsSerialSlot$|^TestFactoryAssignBundleOrderGuard$' -count=1 -v` 의 통과 출력은 들여쓰지 않은 `--- PASS:` 줄 3개와 `ok  github.com/modu-ai/moai-adk/internal/cli`; 짝 `go test ./internal/cli -list '<같은 패턴>'` 는 이름 3개.
- **Mutant probe:** (MU-49) 레인 친화를 무시해 다른 레인이 받음 → (b) 실패. (MU-50) 묶음 멤버를 병렬 임대 → (a) 실패. (MU-51) 직렬 슬롯 의미 변경 → (d) 실패.
- **점화:** W 는 M4 종료, R 은 선택 호 변경, V·S 공통.

## AC-TCI-018 — 병합 동사는 운영자 호출 전용이고 기록하고 닫는다 (REQ-TCI-019)

**Covers**: maps REQ-TCI-019

Release-blocking.

**Given** queued 카드 `a`·`b`, picked 카드 `c`, 그리고 분석 경로
**When** `merge a b`, `merge a c`, `merge b a`(순환), 이미 병합된 `b` 를 다시 `merge d b`, 레인 세션의 `merge`, 그리고 `add`·`analyze`·`relate` 실행
**Then** (a) `merge a b` 뒤 `a` 의 본문에 `[merged from b] …` 절이 붙고 `b` 가 사유 `merged into a` 로 drop 되며 `merged-into` 소견(`b → a`)이 기록되고, (b) picked 카드·이미 병합된 카드(둘째 `merged-into` 대상)·닫히거나 이미 병합된 카드로의 병합(순환 포함)은 거절되고 큐 파일이 바이트 동일이며, (c) 레인 세션에서는 `todoRefuseLaneMutation` 이 거절하고, (d) `add`·`analyze`·`relate`·분석기가 병합 함수를 호출하지 않는다는 것이 호출 그래프 시험으로 강제된다.

필수 시험 4개: `TestTodoMergeRecordsAndDrops`, `TestTodoMergeRefusesPickedAndCycles`, `TestTodoMergeNeverInvokedByAnalysis`, `TestTodoMergeRefusedInLane`.

- **RED-now:** L19, L27(대조군 C1).
- **Green path:** M4 — `go test ./internal/cli -run '^TestTodoMergeRecordsAndDrops$|^TestTodoMergeRefusesPickedAndCycles$|^TestTodoMergeNeverInvokedByAnalysis$|^TestTodoMergeRefusedInLane$' -count=1 -v` 의 통과 출력은 들여쓰지 않은 `--- PASS:` 줄 4개와 `ok  github.com/modu-ai/moai-adk/internal/cli`; 짝 `go test ./internal/cli -list '<같은 패턴>'` 는 이름 4개.
- **Mutant probe:** (MU-52) 병합이 큐 순서도 바꿈 → (a) 의 순서 단언 실패. (MU-53) `relate` 가 병합을 부름 → (d) 실패. (MU-54) 병합이 관계 기록 없이 drop 만 함 → (a) 실패.
- **점화:** W 는 M4 종료와 sync-audit, R 은 병합 함수의 호출자 추가, V·S 공통.

## AC-TCI-019 — 허브 파일을 담은 카드는 체인으로 직렬화되고 출하 목록은 추적되는 입력이다 (REQ-TCI-020)

**Covers**: maps REQ-TCI-020

Release-blocking.

**Given** M0 가 만든 허브 목록(고정 입력으로 `internal/template/catalog.yaml` 포함)을 임베드한 데이터 파일과, `files` 속성에 그 경로를 명시로 담은 열린 카드 둘(본문 경로 추출만으로는 체인을 만들지 않는다 — `spec.md` Module A 용어)
**When** 둘째 카드의 팩토리 레코드가 만들어지고 선택이 돌고, 목록 시험이 실행된다
**Then** (a) 둘째의 `after` 가 첫째로 채워지고 첫째가 로컬 병합되기 전에는 임대되지 않으며, (b) keep-set 판정이 파일 겹침을 읽지 않는다는 것이 시험으로 고정되고(`TestFactoryKeepSetReadsNoFileOverlap`), (c) 허브가 아닌 경로만 겹치는 카드는 체인이 되지 않으며, (d) 출하 코드의 목록 적재가 SPEC 디렉터리·`.moai/reports/` 경로를 읽지 않고 임베드된 데이터 파일만 읽는다 — 정적 검사와 동작 검사 둘이다: `TestHubFileLoaderReadsNoProjectPath` 는 적재 파일에 `.moai` 문자열 리터럴이 없음을 읽고(경로를 조립하는 호출은 이 검사를 피한다), `TestHubFileLoaderIgnoresProjectTree` 는 빈 임시 디렉터리와, `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/hub-files.txt`·`.moai/reports/` 아래 파일에 표식 경로를 심은 임시 디렉터리 둘을 각각 작업 디렉터리로 삼아 적재해 두 결과가 같고 임베드 목록과 같으며 표식 경로가 없음을 읽는다(경로 조립으로 프로젝트 트리를 읽는 적재를 잡는다), (e) 임베드 목록의 경로 집합이 기준선의 추적되는 `hub-files.txt` 의 경로 집합과 같고 기준선 파일이 없으면 **실패**한다(건너뛰지 않는다; `TestHubFileListFromBaseline`), (f) 기준선이 적은 각 경로의 단일 호출 측정 명령(통합 브랜치 자체를 읽으므로 `--first-parent` 형태)을 기록된 develop 팁 SHA 에서 다시 돌려 기록된 개수와 비교한다(`TestHubFileListMatchesMeasuringCommand`). 그 SHA 를 클론이 갖지 않으면 이 시험은 사유를 출력하고 **건너뛰며 그 건너뜀은 통과가 아니라 공백이다**. 건너뜀은 `ok` 로 출력되어 읽는 사람이 묻지 않으면 보이지 않으므로(`verification-completeness.md` §1.3), 멈춤 신호를 시험이 스스로 낸다: 환경 변수 `CI` 가 설정된 실행에서는 SHA 가 없을 때 건너뛰지 않고 **실패**한다 — CI 의 test 작업은 `fetch-depth: 0` 이라(`.github/workflows/ci.yml:131`, 이터레이션 3 이 `grep -n fetch-depth` 로 읽었다) SHA 가 있어야 하고, 없다는 것은 통합 브랜치에 push 되지 않은 SHA 를 기록했거나 체크아웃이 얕아졌다는 뜻이다. 건너뛸지 실패할지의 결정은 순수 함수이고 `TestHubFileMeasurementSkipPolicy` 가 표로 읽는다(SHA 없음 + `CI` 미설정 → 건너뜀, SHA 없음 + `CI` 설정 → 실패, SHA 있음 → 실행). 이 시험은 함수만 읽지 않고 **측정 시험의 본문이 그 함수를 거친다는 것**도 읽는다 — 측정 시험의 본문은 건너뜀·실패·실행을 기록하는 작은 보고 이음매 하나를 받는 함수 하나로 구현하고, 정책 시험이 그 함수를 같은 세 입력으로 직접 구동해 건너뜀·실패·실행을 각각 관측한다(D32). 그리고 **래퍼가 그 함수의 보고를 거친다는 것**도 읽는다(이터레이션 4 감사 D37): 정책 시험은 측정 시험 파일을 `go/parser` 로 읽어 래퍼 `TestHubFileListMatchesMeasuringCommand` 의 본문이 (i) 측정 함수를 정확히 한 번 부르고 그 인자로 `t` 를 보고 이음매에 감싸 넘기며, (ii) 직접 부르는 `t.Skip*`·`t.Fatal*`·`t.Error*` 가 하나도 없음을(건너뜀과 실패가 모두 보고 이음매를 지난다) 단언한다. 측정 함수를 부르지 않고 SHA 가 없으면 곧바로 건너뛰는 래퍼는 (i) 과 (ii) 둘 다에서 걸린다. 레인의 전체 이력 클론에서는 통과해야 한다.

필수 시험 7개: cli 의 `TestFactoryAssignBundleHubChain`, `TestFactoryKeepSetReadsNoFileOverlap` 과 목록을 소유한 패키지(작업 이름 `internal/homestate`)의 `TestHubFileListFromBaseline`, `TestHubFileListMatchesMeasuringCommand`, `TestHubFileLoaderReadsNoProjectPath`, `TestHubFileLoaderIgnoresProjectTree`, `TestHubFileMeasurementSkipPolicy`.

- **RED-now:** L20, L30, L31, L43, L44(대조군 C1, C15, C13). 붉은 이유: 허브 체인·임베드 목록·적재와 그 시험이 아직 없다.
- **Green path:** M4 — `go test ./internal/cli -run '^TestFactoryAssignBundleHubChain$|^TestFactoryKeepSetReadsNoFileOverlap$' -count=1 -v` 의 통과 출력은 `--- PASS:` 줄 2개와 `ok  github.com/modu-ai/moai-adk/internal/cli`(`-list` 이름 2개), 그리고 `go test ./internal/homestate -run '^TestHubFileListFromBaseline$|^TestHubFileListMatchesMeasuringCommand$|^TestHubFileLoaderReadsNoProjectPath$|^TestHubFileLoaderIgnoresProjectTree$|^TestHubFileMeasurementSkipPolicy$' -count=1 -v` 의 통과 출력은 `--- PASS:` 줄 5개와 `ok  github.com/modu-ai/moai-adk/internal/homestate`(`-list` 이름 5개; 둘째가 `--- SKIP` 이면 통과로 세지 않는다). 짝 `git ls-files --error-unmatch internal/homestate/hub_files.txt` 는 그 경로를 출력하고 종료 0.
- **Mutant probe:** (MU-55) 겹침을 keep-set 안에서 읽음 → (b) 실패. (MU-56) 비허브 겹침도 체인 → (c) 실패. (MU-88) 적재가 `.moai/specs`·`.moai/reports` 문자열 리터럴로 프로젝트 파일을 읽음 → (d) 의 정적 검사 실패. (MU-105) 적재가 경로를 조각이나 설정 값에서 조립해(리터럴 `.moai` 를 한 번에 쓰지 않고) 프로젝트 파일을 읽음 → 정적 검사는 피하지만 (d) 의 동작 검사에서 표식 경로가 결과에 나타나 실패한다. (MU-89) 임베드 사본이 기준선 사본과 한 경로 어긋남 → (e) 실패. (MU-90) 기록된 개수가 틀린 경로 줄 → (f) 실패. (MU-106) 측정 시험이 `CI` 가 설정돼도 SHA 없음에서 조용히 건너뜀 → 순수 함수를 고친 변이는 `TestHubFileMeasurementSkipPolicy` 의 표가 잡고, 함수를 거치지 않고 SHA 가 없으면 무조건 건너뛰는 변이(이터레이션 3 감사 D32)는 정책 시험이 측정 본문을 세 입력으로 직접 구동해 (SHA 없음 + `CI` 설정)에서 실패 대신 건너뜀을 관측하고 실패하며, 함수를 부르지 않고 SHA 가 없으면 무조건 건너뛰는 래퍼는 (f) 의 소스 읽기 (i)(ii) 가 호출 부재와 직접 건너뜀 호출로 잡는다(코드가 없어 기준 문면에 대한 읽기 판정). 한계(수용): 래퍼가 측정 함수를 부르되 `skip := t.Skip` 같은 별칭으로 건너뛰는 변이는 호출 표현만 보는 이 읽기가 못 본다 — MU-111·MU-122 와 같은 계열이다. 한계(수용, 이터레이션 5 감사 D42): 소스 읽기는 래퍼 본문만 읽고 보고 이음매의 **어댑터**(`Skip`·`Fail` 이 `t` 로 전달되는지)는 읽지 않는다 — 보고를 삼키는 어댑터는 (i)(ii) 를 통과하고 `CI` 에서 SHA 가 없어도 조용히 지나간다. 어댑터가 `t` 로 전달함을 단언하는 한 줄이 닫는 길이지만 정책 시험의 읽기 범위를 넓히는 설계 변경이라 보류한다(이터레이션 6 은 설계를 바꾸지 않는다). 잔여(수용): (f) 는 목록에 **오른** 경로가 기록 개수를 만족하는지(건전성)만 다시 잰다 — 목록에 **빠진** 허브(완전성)는 M0 스크립트 실행 기록에만 기댄다. 잔여(수용, MU-111): 정적·동작 검사는 작업 디렉터리와 문자열 리터럴만 본다 — 실행 파일 위치나 `$HOME` 에서 경로를 만들어 프로젝트 파일을 읽는 적재는 둘 다 통과한다. 적재가 `go:embed` 한 함수라는 설계(`design.md` §7.4)가 그 틈을 좁히고 리뷰가 닫는다.
- **점화:** W 는 M4 종료, R 은 허브 목록·레코드 생성 시점·적재 경로 변경, V·S 공통.

## AC-TCI-020 — M5 모드 선택은 게이트 명령이 하고 정확히 한 모드가 성립한다 (REQ-TCI-021, -022)

**Covers**: maps REQ-TCI-021, REQ-TCI-022

Release-blocking. 이 SPEC 이 항상 초록으로 만들 수 있도록 두 모드의 합집합이 아니라 **게이트가 가리킨 모드**가 성립하는 것이 기준이다.

**Given** M5 시작 직전의 게이트 읽기. 다섯 판독을 한 번씩 따로(각각 단일 호출, 파이프 없음) 읽고 출력된 **수**를 읽는다 — 판독 2 의 목록 출력 외에는 모두 종료 코드 0 을 내므로 종료 코드는 판정 근거가 아니다. 이 판독들은 움직이는 ref `HEAD` 를 **의도적으로** 쓴다: 질문이 "지금 이 작업 트리에서 t1453 의 착지에 닿는가"이고, t1453 이 착지하거나 카드 브랜치가 develop 을 흡수해 답이 뒤집히는 것이 대상에 대한 참 신호이기 때문이다 — 핀하면 질문이 사라진다(`verification-completeness.md` §4 의 판별 질문). 핀한 값은 L21·C8·C9·C21·C22·C28~C36·C45~C48 이 SHA 로 기록한다.

1. 대상: (a) `T` = `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9]' HEAD` 와 (b) `A` = `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9].*absorb' HEAD` — `--grep` **하나**이고 `--all-match` 가 없다. `T` 는 어느 부모 경로로든 닿는 t1453 제목 **병합** 수(`--merges`; `--first-parent` 를 쓰지 않는다 — L21)이고 `A` 는 그 가운데 **병합 제목 줄**(`merge` 로 시작해 카드 id 를 담은 줄)에 `absorb` 가 같이 든 흡수 방향 병합 수다. git 은 `--grep` 정규식을 메시지의 줄 단위로 맞추므로 `.*` 가 줄바꿈을 건너지 못해 `absorb` 가 본문에만 있는 착지는 `A` 에 세어지지 않는다(D27; C34 는 0, 옛 형태 C35 는 1). **착지 후보 = T − A ≥ 1** 이면 후보(필요 조건 — 후보의 제목 위치 확인은 판독 2 의 목록이 한다). `A` 를 빼는 이유: 미착지 브랜치가 이미 다섯 개의 흡수 병합을 갖고 있어(G23·C28·C29) 그것을 세는 게이트는 착지 없이 열린다. **`T ≥ 1` 인데 `T − A = 0` 이면 닫힘으로 읽기 전에 판독 2 의 `git log` 목록을 한 번 읽어** 제목이 착지로 읽히는 줄이 있는지 본다 — 제목 줄에 `absorb` 가 든 착지도 `A` 가 흡수 방향으로 세어 뺀다(안전한 방향이지만 신호 없이 닫히지 않게 한다). **그 목록의 처분(이터레이션 4 감사 D36)**: 제목이 착지로 읽히는 줄이 있으면 카드 레인은 게이트를 스스로 열지 않는다 — 그 줄의 SHA 를 `progress.md` §E.2 에 적어 리더에게 보고하고, 리더가 그 SHA `S` 에 판독 2 의 두 점검(같은 줄 형태 `.*absorb` 질의, `git merge-base --is-ancestor <S> HEAD`)과 제목을 읽은 결과를 적은 판정을 `progress.md` 에 남기면 그 판정이 `T − A ≥ 1` 과 같은 효력으로 판독 1 의 후보를 대신한다. 읽을 줄이 없거나 리더 판정이 없으면 닫힘(Mode B)이다. 핀 `68a4d8137` 에서 카드 id 를 담은 병합 제목 382개 가운데 제목 줄에 `absorb` 가 든 것이 83개이고(이터레이션 4 핀 `ad02a5677` 에서는 333개 가운데 76개), 그 가운데 제목이 `into develop` 으로의 착지로 읽히는 것이 적어도 셋 있다(`584cfc1a5`, `615d18c1f`, `4c3b1653c` — 제목 모양만 읽은 추정이다). **반대 방향의 거짓 열림도 같은 앵커가 만든다** — 흡수 방향 병합인데 제목에 낱말 `absorb` 가 없으면 `A` 가 세지 못해 착지 후보로 읽힌다. 고정한 통합 팁의 첫 부모 사슬을 기준으로 첫 부모가 사슬 밖이고 둘째 부모가 사슬 위인 병합을 흡수 방향, 반대를 착지로 갈라 SHA 만으로 쟀다: 핀 `68a4d8137`(사슬 끝 `30ce3a02d`)에서 흡수 방향 74개 가운데 낱말이 없는 것이 2개(`c2703f698`, `35e2d7157`; 2.7%)이고 착지 260개 가운데 제목에 낱말이 든 것이 7개(거짓 닫힘), 이터레이션 4 핀 `ad02a5677`(사슬 끝 `42d8474de`)에서는 흡수 방향 68개 가운데 같은 2개와 착지 217개 가운데 6개다 — 감사가 움직이는 ref 로 읽은 75개 가운데 2개는 SHA 로 재현되지 않는다(`spec.md` §G). 또 `^` 는 메시지의 **어느 줄 머리에나** 맞는다(이터레이션 5 감사 D40) — git 의 `--grep` 은 메시지 전체를 줄 단위로 훑고 줄마다 따로 맞추므로, 그 줄이 제목인지 본문인지 알 수 있는 정규식은 하나의 git 호출 안에 없다(스크래치로 `-P` 의 `\A`·뒤돌아보기도 시도했고 전부 같은 값이었다). 그래서 두 층으로 막는다. **(i) `--merges`** 가 비-병합 커밋을 모두 뺀다 — 본문 줄만 `merge(t1453):` 로 시작하는 단일 부모 커밋(F1)과 제목이 `merge(t1453): x` 인 비-병합 커밋(F2)은 `T` 에 세어지지 않는다(스크래치 탐침: 옛 형태 각각 1, 새 형태 각각 0 — 아래 "이터레이션 6 이 스크래치 저장소에서 실행한 게이트 선택자 탐침" 표; MU-124). **(ii) 후보의 제목 위치는 판독 2 의 목록이 확인한다** — 목록은 `%s` 로 각 병합의 **제목**을 찍으므로 제목이 다른 카드이고 본문 줄이 그 줄을 인용한 병합(M2)은 `T` 에 1 로 세어져도 목록의 제목이 `merge…t1453` 로 시작하지 않아 `S` 가 되지 못한다 — 이 두 번째 층은 읽는 단계이고 기계가 세는 수가 아니다(수용 잔여 MU-125). 핀 `68a4d8137` 의 12,998개 커밋에서 카드 id 를 가리지 않는 같은 패턴(`t[0-9]+`)을 줄 단위로 읽은 값: 제목이 맞는 커밋 382개는 모두 병합이고 본문 줄만 맞는 커밋은 병합·비병합 모두 0개다(`python3` 스크래치 스크립트이고 원장 행이 아니다; 핀 `b05c3be90` 의 12,490개 커밋에서도 333·0·0). `T`·`A` 의 `--merges` 는 이 이력에서 아무 착지도 떨어뜨리지 않는다(C21 382 와 C48 382 가 같다).
2. 고정: 후보가 있으면 `git log --merges --format='%h %s' -E -i --grep='^merge[( :]+(card )?t1453[^0-9]' HEAD` 의 목록에서 **제목 자체**(SHA 뒤 글자)가 `merge[( :]+(card )?t1453[^0-9]` 로 시작하고 `absorb` 가 없는 줄의 SHA `S` 를 `progress.md` §E.2 에 적고, `git rev-list --merges --no-walk --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9].*absorb' <S>` 가 `0`(판독 1(b) 와 같은 같은 줄 형태 — C31·C32 가 두 방향의 값을 보이고, 본문에 `absorb` 가 든 착지 t1439 도 이 형태에서 0 이다: C34; 이터레이션 3 의 `--grep=absorb`(메시지 전체)는 그 착지에서 1 이라 쓰지 않는다: C36), `git merge-base --is-ancestor <S> HEAD` 가 종료 0 이다. 이후 재독은 이 SHA 술어로 한다(핀한 주소). 목록에 제목이 그렇게 시작하고 `absorb` 가 없는 줄이 없으면(병합 본문 줄만 맞은 병합뿐인 경우 포함) 후보가 아니다. (판독 1 의 `T ≥ 1` 인데 `T − A = 0` 인 경우의 목록은 판독 1 이 적은 대로 처분한다 — 레인은 열지 않고 SHA 를 적어 리더에게 보고한다.)
3. 양성 대조: t1448·t1344 에 판독 1(a) 형태를 걸어 각각 ≥ 1(C8, C9), t1448 에 판독 1(b) 형태가 0(C30 — 착지 병합의 제목 줄은 `absorb` 를 담지 않는다), 그리고 **본문에만 `absorb` 가 든 착지**(t1439)에 판독 1(a) 가 1·(b) 가 0(C33·C34 — 착지 후보 1; 이터레이션 3 의 `A` 형태는 같은 착지에서 1 이라 후보가 0 이었다: C35).
4. 두 번째 부모 대조: `git rev-list --merges --count -E -i --grep='^merge[( :]+(card )?t[0-9]+' HEAD` 가 같은 명령에 `--first-parent` 를 붙인 값보다 **엄격히 크다**(C21 > C22). 같거나 작으면 선택자가 첫 부모 경로 밖의 커밋을 본다는 증거가 없으므로 "미측정 = 미충족".
5. 형태 점검: t1453 이 `merge(t1453)` 형태가 아닌 제목으로, 또는 빨리감기·스쿼시로 착지하면 판독 1 이 0 으로 남아 게이트는 닫힌 채 읽힌다(안전한 방향). 그때는 `moai gtd pr t1453` 가 낸 착지 SHA 를 `git merge-base --is-ancestor <그 SHA> HEAD` 로 직접 읽고 그 출력을 `progress.md` §E.2 에 적은 뒤 리더가 모드를 판정한다. 착지가 되돌려진(revert) 경우는 원래 병합이 조상으로 남아 열린 채로 읽힌다 — 판독 2 의 SHA 와 `git log` 로 리더가 읽는다(이 이력에는 되돌림 사례가 없어 측정하지 못했고 추정이다).

게이트가 열림은 판독 1(착지 후보 ≥ 1, 또는 리더의 판정이 후보를 대신한 경우)·2·3·4 가 모두 성립하는 것이다. 이터레이션 1 의 `--first-parent` 형태는 흡수 뒤 영영 열리지 못하는 눈먼 선택자였다 — 고정 SHA 대조(C17~C20)가 이를 보인다. 이터레이션 2 의 `absorb` 를 가리지 않는 선택자는 반대로 착지 없이 열리는 선택자였다 — 고정 SHA 대조(C28·C29)가 이를 보인다. 이터레이션 3 의 메시지 전체 `A` 는 진짜 착지에서 닫히는 선택자였다 — 실제 이력 대조(C33~C35)가 이를 보인다. 이터레이션 5 의 `--merges` 없는 `T`·`A` 는 반대로 병합이 아닌 커밋의 본문 줄·제목을 착지로 세어 열리는 선택자였다 — 스크래치 탐침(F1·F2, MU-124)이 이를 보이고 C45~C48 이 현재 선택자를 실제 이력의 착지·흡수·HEAD 에 다시 건다.

**When** M5 가 끝난다
**Then** 정확히 한 모드가 성립한다. **Mode A**(게이트 열림): `.claude/rules/moai/workflow/card-issuance.md` 와 템플릿 사본이 있고 여섯 섹션(카드 크기, 후속 지적 규칙, 파생 깊이, 동시 진행 한도, 발행 체크리스트, 부채 대장)을 **메커니즘만** 가지며, 수치는 **로컬 전용 규칙 `.claude/rules/local/card-issuance-thresholds.md` 에만** 있고 각 수치가 기준선 기록의 값과 같다(`TestCardIssuanceRuleValuesMatchBaseline`; 두 파일 모두 추적되므로 CI 가 읽는다), 배포 사본은 `t####` 카드 id·`.moai/` 경로·측정된 수치를 담지 않는다(`TestCardIssuanceTemplateIsMechanismOnly`). 이 시험의 측정 수치 검사는 새 `card-issuance.md` 사본 하나에만 건다 — 편집되는 다섯 사본(`gtd.md`, `factory-dispatch*.md`, `sync-auditor.md`, `manager-todo.md`)은 기존 가드(카드 id 누출, 중립성 정규식의 SPEC·REQ id·날짜·16진 단어)가 닿는 범위까지만 기계가 보고, 숫자 임계값이 그 다섯 사본에 들어가는지는 M5 커밋의 `git diff` 를 리뷰가 읽는다(수용 잔여, MU-121; D31). **Mode B**(게이트 닫힘): 여섯 규칙 파일과 템플릿 사본이 카드 브랜치의 병합 기준(읽는 시점에 `git merge-base develop HEAD` — 움직이는 ref 를 의도적으로 쓴다: "지금 흡수한 기준 이후 무엇이 바뀌었나"가 질문이라 기준이 움직이는 것이 신호이고, 핀한 리터럴은 흡수 뒤 다른 카드의 변경을 이 카드 몫으로 센다)에 대해 변하지 않았고(`git diff --name-only <그 SHA> -- <여섯 파일과 사본>` 이 비어 있다), `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/m5-draft/` 가 **추적되는** 초안(`anchors.md`, `card-issuance.md`, `card-issuance-thresholds.md`)과 삽입 위치 앵커를 갖고, `progress.md` §E.2 에 게이트 다섯 판독의 출력과 후속 카드 문안이 있다. 두 모드의 산출물이 함께 있으면(편집과 초안) 실패다.

- **RED-now:** L21(게이트 0 — 정보용으로 Mode B 선택을 가리킨다), L22(규칙 파일 부재), L23(초안 부재). 두 모드의 산출물이 모두 도착 시 없으므로 붉다.
- **Green path:** M5 — 게이트가 열려 있으면 Mode A, 닫혀 있으면 Mode B. 핀의 게이트 값은 닫힘이므로 지금 읽으면 Mode B 가 성립해야 한다. Mode B 의 통과 출력: `git ls-files --error-unmatch .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/m5-draft/anchors.md` 가 그 경로를 출력하고 종료 0, 그리고 `git diff --name-only <merge-base> -- <여섯 파일과 사본>` 이 빈 출력(종료 0; 양성 대조는 같은 명령에 이 SPEC 이 바꾼 경로를 걸어 비지 않음을 보이는 것). Mode A 의 통과 출력: `go test ./internal/template -run '^TestCardIssuanceRuleValuesMatchBaseline$|^TestCardIssuanceTemplateIsMechanismOnly$' -count=1 -v` 가 `--- PASS:` 줄 2개와 `ok  github.com/modu-ai/moai-adk/internal/template`(`-list` 이름 2개), 그리고 `git ls-files --error-unmatch .claude/rules/moai/workflow/card-issuance.md` 가 경로를 출력하고 종료 0.
- **Mutant probe:** (MU-57) 게이트가 닫혔는데 규칙 파일을 편집(Mode A 산출물) → Mode B 의 diff 가 비지 않아 실패. (MU-58) 게이트를 `--first-parent` 로 읽음(눈먼 선택자) → 같은 선택자가 고정 SHA 위에서 0 을 내고(C19) 모든 부모를 걷는 형태가 1 을 내는(C20) 대조가 이 읽기를 무효로 만든다. (MU-59) 초안에 앵커가 없음 → 앵커 검사 실패. (MU-60) 수치를 기준선 기록 대신 SPEC 에서 복사 → 값 일치 시험 실패. (MU-91) 배포 사본이 측정 수치·`.moai/` 경로·카드 id 를 담음 → `TestCardIssuanceTemplateIsMechanismOnly` 와 카드 id 누출 시험이 실패. (MU-107) 게이트가 `absorb` 를 가리지 않고 t1453 제목 커밋을 센다(판독 1 의 (a) 만 읽음) → 미착지 브랜치의 흡수 병합 다섯 개로 착지 없이 열린다: 같은 선택자가 흡수 병합 `b05c3be90…` 에서 5 를 내는데(C28) 흡수 병합을 가리는 판독 1(b) 도 5 를 내서(C29) 착지 후보가 0 이라는 대조가 이 읽기를 무효로 만든다. (MU-116) 게이트의 `A` 가 `--all-match --grep=absorb` 로 메시지 **전체**를 봄(이터레이션 3 의 형태, D27) → 본문에만 `absorb` 가 든 진짜 착지에서 후보가 0 이 되어 게이트가 닫힌 채 읽힌다: 같은 착지(t1439)에서 옛 `A` 가 1(C35), 새 `A` 가 0(C34)이라 후보가 0 대 1 이라는 대조가 이 읽기를 무효로 만든다(실제 이력에서 실행해 관측). (MU-124) `T`·`A` 의 `--merges` 를 뺌(이터레이션 5 의 형태, D40) → 병합이 아닌 커밋의 본문 줄이나 제목이 `merge(t1453)` 로 시작하면 착지로 세어 게이트가 착지 없이 열린다: 스크래치의 F1·F2 에서 옛 형태 1, 새 형태 0 이라는 대조가 이 읽기를 무효로 만든다(스크래치에서 실행해 관측; C45~C48 이 실제 이력에서 새 형태의 양성·음성 값을 쥔다). 잔여(수용, MU-125): 병합 커밋의 본문 줄이 그 줄을 인용하면 `T` 는 센다 — 후보의 제목 위치는 판독 2 의 목록을 읽어 확인한다. 잔여(수용): 고정 SHA 대조 행(C17~C20, C28·C29·C31)은 미푸시 브랜치의 객체가 있는 클론에서만 재현된다 — 이식 가능한 런타임 대조는 판독 4 의 부등식이다. 잔여(수용, MU-112): 착지가 되돌려진 경우는 게이트가 열린 채로 읽힌다(추정 — 되돌림 사례가 이 이력에 없어 측정하지 못했다; 판독 2 가 SHA 를 적어 리더가 읽는다).
- **점화:** W 는 M5 시작 직전과 첫 커밋 직전, 그리고 sync-audit, R 은 t1453 착지로 게이트 값이 바뀌는 사건, V 는 게이트 출력, S 는 게이트 읽기를 `progress.md` §E.2 에 매번 기록하는 것(읽지 않으면 모드 선택의 근거가 없다).

## AC-TCI-021 — Mode A 의 예산·미러·생성물 정합과 발행 세션 도달성 (REQ-TCI-021)

**Covers**: maps REQ-TCI-021

**조건부** — Mode A 에서만 적용된다. Mode B 에서는 "해당 없음(Mode B)"으로 기록하고 **통과로 세지 않으며**, AC-TCI-020 의 Mode B 가 같은 파일을 건드리지 않는다는 것을 보증한다. 출시 차단으로 세지 않는다(게이트가 오늘 닫혀 있어 도착 시 Mode B 가 성립한다).

**Given** Mode A 가 성립한 M5 커밋
**When** 크기와 가드를 읽는다
**Then** (a) `wc -m .claude/skills/moai/workflows/gtd.md` ≤ 40000(도착 시 41160), (b) 상시 로드 `factory-dispatch.md` 의 문자 수와 바이트 수가 병합 기준 blob 보다 크지 않다(기준은 읽는 시점에 `git merge-base develop HEAD` 로 구하고 핀하지 않는다 — 움직이는 ref 를 의도적으로 쓰는 자리이고 이유는 AC-TCI-020 Mode B 와 같다, 핀 값은 G4·G5), (c) `factory-dispatch-detail.md` 가 40,659자(G6)보다 크지 않다, (d) 바이트 동일 쌍(`gtd.md`, `factory-dispatch-detail.md`, `factory-dispatch-mechanics.md`, `manager-todo.md`)은 `cmp` 가 같고, 분기된 쌍(`factory-dispatch.md`, `sync-auditor.md`)은 분기 hunk 수가 편집 전과 같다(G7 에서 `factory-dispatch.md` 는 1 hunk), (e) `card-issuance.md` 가 40,000자 이하이고 `paths:` 패턴 어느 것도 `factory-dispatch*` 글롭과 자기 매칭하지 않으며, (f) `make agents-emit` 으로 `internal/template/templates/.codex/agents/moai/*.toml` 이(템플릿 `.md` 에서 생성된다 — 루트 `.codex/` 는 `.gitignore:137` 로 추적되지 않는다) 갱신되고, `make plugin-emit` 으로 생성 플러그인 페이로드 `plugins/moai/skills/moai/workflows/gtd.md` 가 템플릿 `gtd.md` 와 바이트 동일하게 다시 만들어지고(흡수가 더한 미러 — G31), `catalog.yaml` 해시가 생성기로 다시 만들어져 `internal/template` 의 `TestManifestHashFormat`·`TestCatalogHashCoversSkillSubfiles` 와 `internal/template/agentemit` 의 `TestGoldenCommittedArtifactsMatchEmission` 과 `internal/template/pluginemit` 의 `TestGoldenCommittedArtifactsMatchEmission`·`TestCommittedVersionMatchesSSOT` 가 통과하며(첫 둘과 agentemit 시험의 패키지 귀속은 G21·G22, pluginemit 은 G31), (g) 템플릿 가드(`internal/template` 의 카드 id 누출 시험 `TestTemplateNoInternalContentLeak`, 카드 id 기준선 가드 `TestCardIDBaselineHasNoStaleEntries`·`TestCardIDBaselineIsPerFileAndPerLiteral`, 경로 고정 `TestWorkflowRulePathsPinned`, `TestDeclaredRuleMirrorForks`; `internal/cli` 의 중립성·문서 정합 `TestTodoSkillDocumentsJevSource`·`TestTodoSkillDocumentsClassification`, 갈라진 쌍 `factory-dispatch.md` 의 분기 보존 `TestAutoPickMirrorParity`(G32), `gtd.md` 금지 문장 고정 `todo_auto_doc_test.go`(G33); 편집 파일을 이름으로 대는 시험 전부는 plan §F.8 단계 6 이 모은다)가 통과하고 — M5 가 만지는 파일에 걸린 카드 id 기준선 짝은 네 개(`gtd.md t696`, `factory-dispatch.md t1330`, `factory-dispatch-detail.md t133`·`t224`; 템플릿 사본의 줄은 `gtd.md:249`, `factory-dispatch.md:94`)라서 압축이 그 짝의 문장을 지우면 같은 커밋에서 `internal/template/card_id_leak_test.go` 의 `cardIDBaseline` 항목을 지워야 하고(남기면 `TestCardIDBaselineHasNoStaleEntries` 가 `stale baseline entry` 로 붉다), `gtd.md t696` 짝을 지우면 `TestCardIDBaselineIsPerFileAndPerLiteral` 이 그 짝을 양성 대조 상수로 쓰므로 다른 짝으로 고쳐야 한다, (h) **발행 세션 도달성**: 상시 로드 `factory-dispatch.md` 가 `card-issuance.md` 를 이름으로 가리키는 문장을 가지며(`git grep -c -F "card-issuance" -- .claude/rules/moai/workflow/factory-dispatch.md` 가 1 이상), `card-issuance.md` 의 `paths:` 가 `workflows/gtd.md` 를 포함하고 `internal/template/workflow_rule_paths_pinned_test.go` 의 고정 목록에 등록돼 있어 경로가 바뀌면 그 시험이 붉어진다(점화 §1.3 의 멈춤 신호 — 경로 한정 규칙만으로는 `gtd.md` 나 `manager-todo.md` 를 열지 않는 세션에 닿지 않으므로 스텁이 도달을 맡고 고정 시험이 그 도달이 조용히 끊기는 것을 막는다).

- **RED-now:** L24(`gtd.md` 41,160자 — 한도 초과, 붉은 이유는 이 작업이 `gtd.md` 를 고치며 한도 안으로 되돌려야 하기 때문), L36(스텁이 아직 새 규칙을 가리키지 않음; 대조군 C16). Mode B 에서는 L24 가 영영 붉으므로 이 기준은 Mode B 에서 **적용하지 않는다**(wrong-reason red 를 통과로 기록하지 않는다).
- **Green path:** M5 Mode A — `wc -m .claude/skills/moai/workflows/gtd.md` 가 `40000 이하` 의 수와 경로를 출력하고(종료 0), `git grep -c -F "card-issuance" -- .claude/rules/moai/workflow/factory-dispatch.md` 가 `<경로>:N`(N ≥ 1, 종료 0)을 출력하고, `go test ./internal/template -run '^TestManifestHashFormat$|^TestCatalogHashCoversSkillSubfiles$|^TestDeclaredRuleMirrorForks$' -count=1 -v` 가 `--- PASS:` 줄 3개와 `ok  github.com/modu-ai/moai-adk/internal/template`(`-list` 이름 3개)를, `go test ./internal/template/agentemit -run '^TestGoldenCommittedArtifactsMatchEmission$' -count=1 -v` 가 `--- PASS: TestGoldenCommittedArtifactsMatchEmission ` 한 줄과 `ok  github.com/modu-ai/moai-adk/internal/template/agentemit`(`-list` 이름 1개)을 출력한다.
- **Mutant probe:** (MU-61) `gtd.md` 를 한도에 맞추려 findings-source 열거 문장을 삭제 → 문서 정합 시험 실패. (MU-62) 로컬만 편집 → (d) 의 `cmp` 실패. (MU-63) 상시 로드 스텁에 내용을 이동해 예산 충족 → (b) 실패. (MU-64) 템플릿 사본에 카드 id 추가 → 누출 시험 실패. (MU-65) 카탈로그 미재생성 → 해시 시험 두 개 실패(AUTO-PICK 계획이 관측한 같은 모양). (MU-108) 패키지 귀속이 틀린 선택자(`TestGoldenCommittedArtifactsMatchEmission` 을 `internal/template` 에서 찾음) → 그 패키지의 `-list` 가 이름 3개를 내서 지정한 4개와 달라지므로 짝이 잡는다 — 이터레이션 2 문서가 실제로 이 오류를 갖고 있었다(G21·G22). (MU-92) 스텁 문장 없음 → (h) 의 `git grep` 이 비어 실패. (MU-93) 새 규칙의 `paths:` 를 고정 목록에 등록하지 않음 → 경로가 바뀌어도 시험이 붉어지지 않으므로 (h) 의 등록 확인이 실패. (MU-120) 압축이 기준선 짝(예: `factory-dispatch.md` 의 t1330 문장)을 지우되 `cardIDBaseline` 항목은 남김 → `TestCardIDBaselineHasNoStaleEntries` 가 실패한다(`card_id_leak_test.go:69`; 이터레이션 3 감사의 추정 위험 — 이터레이션 4 가 그 시험을 읽고 현재 트리에서 통과함을 실행해 확인했다, 변이 자체는 만들지 않았다).
- **점화:** W 는 M5 종료, R 은 규칙 파일 증가·스텁 삭제·경로 변경, V·S 공통(+ `wc -m` 출력).

## AC-TCI-022 — 웹 관계 그래프 보기 (REQ-TCI-023)

**Covers**: maps REQ-TCI-023

Release-blocking.

**Given** live·dropped·보관 카드와 각 종류의 관계를 담은 고정 큐, 그리고 노드 상한을 넘는 큰 고정 큐
**When** `GET /todo?view=graph` 와 `POST /todo?view=graph` 가 요청된다
**Then** (a) 200 이고 모든 고정 카드(보관·dropped 포함)의 노드와 각 관계 종류의 간선이 서버 렌더 HTML/SVG 로 나온다, (b) POST 는 405 이고 큐 파일의 바이트가 바뀌지 않으며, 그래프 보기는 **큐 락을 기다리지도 잡지도 않는다** — (b1) 락 파일의 바이트·mtime 이 안 바뀐다는 점검은 락을 잡지 않았다는 증거로 쓰지 않는다: advisory flock 은 락 파일에 아무것도 남기지 않아(스크래치 확인: 잡았다 놓은 뒤 바이트·mtime·크기가 그대로다) 락 파일을 열어 flock 만 하는 보기는 그 점검을 통과한다. 이 저장소의 락 도우미(`acquireStateLockImpl`, `state_lock_unix.go` 의 공용 opener)를 거치면 소유자 기록이 락 파일에 쓰여 그 점검이 그 모양만은 잡지만 raw flock 은 못 잡는다(이터레이션 3 의 (b) 가 이 점검만 읽었다 — D28). (b2) **락을 쥔 채 요청**: 시험 `TestTodoGraphViewDoesNotWaitOnQueueLock` 이 저장소 자신의 도우미 `BacklogStore.Mutate` 안에서 락을 쥔 채(콜백이 신호를 기다리며 막혀 있다) `GET /todo?view=graph` 를 요청하고 **2초 안에 200** 과 모든 고정 카드의 노드를 받는다. 같은 시험이 같은 락 아래 기본 보기 `GET /todo` 도 2초 안에 200 이라는 통제를 함께 읽는다 — 락을 쥔 하네스 자체가 읽기를 막지 않음을 보여, 그래프 보기만 막히는 경우를 가려낸다(wrong-reason red 방지; 큐 읽기는 락이 없다는 `backlog_store.go` 머리 주석을 읽은 것이고 이 반복은 이 통제를 실행하지 못했다). 2초는 락 도우미의 대기 예산(`stateLockWaitBudget` = `stateLockSupportedWriters`(10) × `stateLockCIMutationCost`(33ms) × `stateLockHeadroom`(10) = 3.3초; `state_lock_wait.go` — 이터레이션 4 까지는 `board_store.go` 의 `boardLock*` 이름이었고 값은 같다)보다 짧아서, 도우미를 거쳐 기다리다 예산을 다 쓰고 오류를 삼킨 채 그리는 보기도 3.3초 뒤에야 돌아오므로 잡히고, 블로킹 flock 보기는 풀릴 때까지 돌아오지 못한다. 건강한 보기가 작은 고정 큐를 `-race` 아래에서 2초 안에 그린다는 것은 구현 단계가 관측해 확인할 몫이다(이 반복은 관측하지 못했다). (b3) **어휘 검사**: 웹 패키지의 시험이 아닌 소스에 락을 잡는 호출 토큰(`Mutate(`, `LockPath`, `Flock`, `acquireLock`, `StateLock`)이 하나도 없다(`TestTodoGraphViewReadOnly` 가 소스를 읽는다; 오늘 값과 양성 대조는 G27~G29) — (b2) 는 락을 **시도만 하고** 계속하는 보기(`LOCK_NB` 가 실패해도 진행)를 통과시키므로 (b3) 가 그 틈을 닫는다. 한계(수용, MU-122): (b3) 은 토큰만 보므로 토큰을 피해 이름을 바꿔 락을 시도만 하는 보기는 못 잡는다 — `design.md` §9 의 "락·쓰기 없음"과 리뷰가 닫는다, (c) 큰 큐는 상한을 넘으면 "N개 생략" 문구와 함께 상한 개수의 노드만 낸다, (d) 응답 본문에 외부 참조가 없다 — `src`·`href`·`action`·`data-src`·`srcset` 속성값과 `url(…)`·`@import` 인자가 `http:`·`https:`·`//`(프로토콜 상대) 로 시작하는 것이 하나도 없고, 그래프 보기가 참조하는 자산 경로는 모두 임베드 파일 시스템에서 해석되며, 그래프 보기가 새 `.js` 자산을 더하지 않는다(임베드 자산 목록의 `.js` 항목이 M6 시작 때와 같다). 한계(수용, MU-113): 서버가 렌더한 마크업과 CSS 를 훑는 정적 검사이므로 기존 JS 가 런타임에 만드는 요청은 보지 못한다 — `design.md` §9 의 "새 JS 를 만들지 않는다"와 위 목록 불변이 그 틈을 좁힌다, (e) 새 문자열은 모든 로케일에 키가 있다(i18n 거버넌스 시험 통과), (f) 같은 입력은 같은 출력이다.

필수 시험 5개: `TestTodoGraphViewRendersRelations`, `TestTodoGraphViewReadOnly`, `TestTodoGraphViewBounded`, `TestTodoGraphAssetsEmbedded`, `TestTodoGraphViewDoesNotWaitOnQueueLock`.

- **RED-now:** L25, L45(대조군 C6). 붉은 이유: 그래프 보기와 그 시험, 락을 쥔 채 요청하는 시험이 아직 없다.
- **Green path:** M6 — `go test ./internal/web -run '^TestTodoGraphViewRendersRelations$|^TestTodoGraphViewReadOnly$|^TestTodoGraphViewBounded$|^TestTodoGraphAssetsEmbedded$|^TestTodoGraphViewDoesNotWaitOnQueueLock$' -count=1 -v` 의 통과 출력은 들여쓰지 않은 `--- PASS:` 줄 5개와 `ok  github.com/modu-ai/moai-adk/internal/web`(마지막 시험의 통제 하위 시험이 `--- SKIP` 이면 통과로 세지 않는다); 짝 `go test ./internal/web -list '<같은 패턴>'` 는 이름 5개. 같은 마일스톤이 i18n 거버넌스 시험을 통과시킨다.
- **Mutant probe:** (MU-66) 보기 렌더가 락을 잡음(락 도우미를 거쳐) → (b2) 가 2초 안에 200 을 받지 못해 실패한다(도우미를 거치면 락 파일의 소유자 기록도 바뀌어 (b1) 도 그 모양을 본다). (MU-117) 보기가 락 파일을 직접 열어 **블로킹 flock** 을 함 → 락 파일의 바이트·mtime 은 그대로라 이터레이션 3 의 (b) 는 통과했지만(D28; 스크래치에서 flock 이 락 파일에 아무것도 남기지 않고, 락을 쥔 동안 두 번째 설명자의 블로킹 flock 이 1.0초 안에 돌아오지 못함을 관측했다) (b2) 는 락을 쥔 채 2초 안에 200 을 받지 못해 실패한다. (MU-118) 보기가 락을 **시도만 하고**(`LOCK_NB`) 실패해도 계속함 → (b2) 는 통과하지만(스크래치: `LOCK_NB` 시도는 락을 쥔 동안 즉시 실패하고 돌아온다) (b3) 의 어휘 검사가 `Flock` 토큰으로 잡는다. (MU-67) CDN 링크(`https://…`) → (d) 실패. (MU-109) 프로토콜 상대 `src="//cdn.example.com/x.js"`, CSS 의 `@import url(//…)`, 외부 `srcset`·`data-src` → 이터레이션 2 의 "`http` 로 시작하는 `src`/`href`" 문구는 이 변이를 통과시켰고 넓힌 (d) 는 `//`·`url(`·`@import`·`srcset` 까지 읽어 실패한다. (MU-68) 노드 무제한 → (c) 실패. (MU-69) 로케일 키 누락 → (e) 실패. 잔여(수용): MU-113 — 기존 JS 가 런타임에 요청을 만드는 변이는 정적 검사가 보지 못한다(위 한계). 잔여(수용): MU-122 — 어휘 검사 (b3) 을 피해 이름을 바꾸어 락을 시도만 하는 보기는 못 잡는다(위 한계).
- **점화:** W 는 M6 종료, R 은 보기 렌더의 쓰기·네트워크 추가, V·S 공통.

## AC-TCI-023 — 기존 `/todo` 는 그대로다 (REQ-TCI-024)

**Covers**: maps REQ-TCI-024

**회귀 가드**(M6 시작 시 변경 전 렌더러에 대해 먼저 쓰므로 도착 시 초록).

**Given** `view` 매개변수 없는 `GET /todo` 와 같은 큐
**When** M6 이후 렌더한다
**Then** 응답 바이트가 M6 시작 시 기록한 기준과 같고 기존 `todo_route_test.go`·정렬·상세·live 시험이 통과한다(`TestTodoPageUnchangedWithoutViewParam`).

- **RED-now:** 해당 없음(도착 시 초록).
- **Mutant probe:** (MU-70) 기본 보기에 그래프 패널을 끼움 → 기준 바이트가 달라 실패.
- **점화:** W 는 M6 종료와 sync-audit, R 은 기본 렌더 변경, V·S 공통.

## AC-TCI-024 — 대상 SPEC 개정이 sync 에서 선언된다 (Amendments A1~A6)

**Covers**: maps REQ-TCI-012, REQ-TCI-018, REQ-TCI-019, REQ-TCI-023 (개정 행의 근거 요구)

Release-blocking; sync 단계에서 확인한다.

**Given** sync 단계, 결정 D2·D5·D6·D14 의 판정에 따라 적용으로 남은 개정 행(`spec.md` `## Amendments`; 기본 작업값에서는 A2·A3·A4·A6 이 적용, A1·A5 는 조건부로 미적용)
**When** manager-spec 재위임으로 대상 SPEC 의 개정 커밋이 만들어진다
**Then** (a) 적용된 대상 SPEC 의 `spec.md` 가 `amendment_of:`(자기 참조)와 `## Amendments` 절을 가지며 절이 `prior_completed_sha` 로 이 SPEC 의 행과 같은 값을 적고(ANALYSIS `b6716a748`, AUTO-PICK `4293b2d73`, RELATION-PICKUP-FILTER `dcf744fdc`, GTD-AUTONOMY `5ec516165ef5c5e31cdf372b28e6e538f8347719`, WEB-TODO-QUEUE `f1e71db4b`), (b) 상태가 `completed → in-progress → completed`(sync 커밋)로 전이하고 `moai spec audit` 가 그 SPEC 에 `SyncStatusDrift` 를 내지 않으며, (c) 적용하지 않기로 한 조건부 행(A1·A5 등)은 "미적용"으로 기록돼 대상 SPEC 이 편집되지 않는다.

- **RED-now:** L26(대조군 C10: `amendment_of` 를 가진 SPEC 이 존재).
- **Green path:** sync 단계의 개정 커밋 — 통과 출력은 적용된 각 대상에 대해 `git grep -c -F "amendment_of" -- .moai/specs/<대상 SPEC>/spec.md` 가 `<경로>:N`(N ≥ 1, 종료 0)을 내는 것(SPEC-TODO-ANALYSIS-001, SPEC-TODO-AUTO-PICK-001, SPEC-RELATION-PICKUP-FILTER-001, SPEC-WEB-TODO-QUEUE-001), `git grep -c -F "<prior_completed_sha>" -- .moai/specs/<대상 SPEC>/spec.md` 가 같은 모양으로 N ≥ 1 인 것, 미적용 대상 SPEC-GTD-AUTONOMY-001 에서 같은 `amendment_of` 질의가 빈 출력 종료 1 인 것(편집되지 않았다는 증거), 그리고 이 트리에서 만든 빌드를 경로로 호출한 `/tmp/moai-t1454 spec audit --filter-spec <대상 SPEC> --json` 의 `drift_findings` 에 `SyncStatusDrift` 항목이 없는 것이다.
- **Mutant probe:** (MU-71) 대상 SPEC 본문을 `amendment_of` 없이 편집 → `moai spec audit` 가 잡는다. (MU-72) `prior_completed_sha` 가 progress.md 의 값과 다름 → (a) 실패. (MU-73) 미적용이어야 할 조건부 행을 적용 → (c) 실패.
- **점화:** W 는 sync 단계, R 은 대상 SPEC 의 무선언 편집, V 는 `moai spec audit` 와 manager-spec 의 개정 커밋 검토, S 는 대상 SPEC 이 `completed` 인데 이 SPEC 의 개정 행이 적용 기록 없이 남는 상태를 sync-audit 가 읽는 것.

## 변이 탐침 요약

| 변이 | 위반 | 잡는 기준 |
|---|---|---|
| MU-1~3 빈 디렉터리, 명령 없는 숫자, 재현되지 않는 값 | REQ-TCI-001 | AC-TCI-001 |
| MU-78~79 무시되는 경로에만 기준선, 기준선이 코드와 한 커밋 | REQ-TCI-001 | AC-TCI-001 |
| MU-4~7 stdout 출력, live 만 조회, 접두사 미제거, 상위 4개 | REQ-TCI-002 | AC-TCI-002 |
| MU-8~12 live 만 이웃, 하한 무시, draft SPEC 포함, 빈 겹침을 none, 분류기 호출 | REQ-TCI-002, -006 | AC-TCI-003, -007 |
| MU-13~15 락 안 탐침, 탐침 오류 전파, 상한 없음 | REQ-TCI-003 | AC-TCI-004 |
| MU-16~18 id 소비, 소견 기록, 중복에서 비제로 종료 | REQ-TCI-004 | AC-TCI-005 |
| MU-19~21 제시를 첫 줄 앞에, engage 소견 기록, 빈 제시에 빈 줄 | REQ-TCI-005 | AC-TCI-006 |
| MU-22~23 분류기 변경, 같은 파일에 함수 추가 | REQ-TCI-006 | AC-TCI-007 |
| MU-24~27 한 표에만 컬럼, 재구성 목록에 포함, retrofit 순서, parity 무시 | REQ-TCI-007~009 | AC-TCI-008 |
| MU-80~83 `archived_findings` 만 누락, 순수 읽기 DDL, 큐 병합 누락, finding 튜플 미고정 | REQ-TCI-008, -009 | AC-TCI-008 |
| MU-28 `omitempty` 없음 | REQ-TCI-007 | AC-TCI-009 |
| MU-29~31 접두사에만 사유, 접두사 제거, 0 시각 | REQ-TCI-010 | AC-TCI-010 |
| MU-32~34 merge 처분이 drop, reject 가 삭제, 기본 처분 | REQ-TCI-011 | AC-TCI-011 |
| MU-35~37 저장 행 재작성, 매핑 누락, 표시 변경 | REQ-TCI-012, -017 | AC-TCI-012 |
| MU-38~40 자기 간선 검사 한 분기, 둘째 `--parent` 덮어쓰기, 거절 전 일부 쓰기 | REQ-TCI-013 | AC-TCI-013 |
| MU-84~85 live 만 보는 부모 검사, `supersedes` 순환 미검사 | REQ-TCI-013 | AC-TCI-013 |
| MU-41~42 해석기가 씀, 한 방향 | REQ-TCI-014 | AC-TCI-014 |
| MU-43~45 깊이 무시, 비결정 순회, 방문 집합 없음 | REQ-TCI-015 | AC-TCI-015 |
| MU-46~48 큐에서 간선 생성, 비결정 출력, 지문 미등록 | REQ-TCI-016 | AC-TCI-016 |
| MU-86~87 첫 부모 경로만 순회, 자체 정규식으로 흡수 병합 귀속 | REQ-TCI-016 | AC-TCI-016 |
| MU-49~51 레인 친화 무시, 병렬 임대, 슬롯 의미 변경 | REQ-TCI-018 | AC-TCI-017 |
| MU-52~54 순서 변경, `relate` 가 병합 호출, 관계 없이 drop | REQ-TCI-019 | AC-TCI-018 |
| MU-55~56 keep-set 이 겹침 읽음, 비허브도 체인 | REQ-TCI-020 | AC-TCI-019 |
| MU-88~90 런타임에 SPEC·reports 경로를 읽음, 임베드와 기준선 사본 불일치, 틀린 기록 개수 | REQ-TCI-020 | AC-TCI-019 |
| MU-57~60 게이트 닫힘에 편집, 눈먼 선택자로 판정, 앵커 없는 초안, 수치 복사 | REQ-TCI-021, -022 | AC-TCI-020 |
| MU-91 배포 사본에 수치·경로·카드 id | REQ-TCI-021 | AC-TCI-020 |
| MU-61~65 열거 문장 삭제, 로컬만 편집, 스텁 이동, 카드 id 추가, 카탈로그 미재생성 | REQ-TCI-021 | AC-TCI-021 |
| MU-92~93 스텁 문장 없음, 새 규칙 경로 미고정 | REQ-TCI-021 | AC-TCI-021 |
| MU-66~69 락 획득, CDN, 무제한, 로케일 누락 | REQ-TCI-023 | AC-TCI-022 |
| MU-70 기본 보기에 패널 | REQ-TCI-024 | AC-TCI-023 |
| MU-71~73 무선언 편집, SHA 불일치, 미적용 행 적용 | Amendments | AC-TCI-024 |
| MU-94 기준선이 제품 커밋 뒤에 착지(증인 1~4 는 통과, `card-head:` 를 `B^` 로 적어 증인 4 도 통과) | REQ-TCI-001 | AC-TCI-001 (증인 5) |
| MU-95 조회는 겹침을 내지만 렌더러가 겹침 줄을 쓰지 않음 | REQ-TCI-002 | AC-TCI-002 (e) |
| MU-96~98 겹침을 항상 `unmeasured`, 모든 진행 중 카드를 겹침으로, 겹침 항목에 경로·레인 누락 | REQ-TCI-002 | AC-TCI-003 (f)(g) |
| MU-99~102 `relate` 가 `parent-of`·`follow-up-of`·`merged-into` 를 받음, 거절하되 종류 미명시·부분 쓰기 | REQ-TCI-013 | AC-TCI-013 (f) |
| MU-103~104 속성 없는 카드가 `{}` 로, 처분 없는 소견이 빈 문자열로 저장 | REQ-TCI-007, -011 | AC-TCI-008 (g) |
| MU-105~106 경로를 조립해 프로젝트 파일을 읽는 적재, CI 에서 조용히 건너뛰는 측정 시험 | REQ-TCI-020 | AC-TCI-019 (d)(f) |
| MU-107 `absorb` 를 가리지 않는 게이트 선택자 | REQ-TCI-021 | AC-TCI-020 (판독 1) |
| MU-108 패키지 귀속이 틀린 시험 선택자 | REQ-TCI-021 | AC-TCI-021 (f) 의 `-list` 짝 |
| MU-109 프로토콜 상대·`@import`·`srcset` 외부 참조 | REQ-TCI-023 | AC-TCI-022 (d) |
| MU-114~115 B 와 무관한 곁가지의 제품 커밋을 카드 id 를 적은 병합으로 들임, 순효과 0 곁가지 | REQ-TCI-001 | AC-TCI-001 (증인 3 의 `--full-history` SHA 목록 비교) |
| MU-116 메시지 전체를 보는 게이트 `A` | REQ-TCI-021 | AC-TCI-020 (판독 1(b)·3) |
| MU-123 B 앞의 곁가지에서 제품 커밋을 넣었다 되돌려 병합(증인 5 의 앞쪽 눈먼 곳) | REQ-TCI-001 | AC-TCI-001 (증인 5 의 `--full-history --simplify-merges`) |
| MU-117~118 락 파일을 직접 flock 하는 보기, 락을 시도만 하는 보기 | REQ-TCI-023 | AC-TCI-022 (b2)(b3) |
| MU-119 파일 이름·접두사로 겹침 비교 | REQ-TCI-002 | AC-TCI-003 (h) |
| MU-120 압축이 카드 id 기준선 짝 문장을 지우되 항목은 남김 | REQ-TCI-021 | AC-TCI-021 (g) |
| MU-124 `--merges` 없는 게이트 선택자(병합이 아닌 커밋의 본문 줄·제목을 착지로 셈) | REQ-TCI-021 | AC-TCI-020 (판독 1(i), C45~C48, 스크래치 F1·F2) |

**이터레이션 6 이 실제로 실행한 변이 탐침(D40).** 명령으로 실행해 관측한 것: **MU-124** — 스크래치 저장소에서 옛 형태와 새 형태를 같은 커밋에 걸었다(아래 표): 옛 형태는 본문 줄만 위조한 비-병합 F1 과 제목만 위조한 비-병합 F2 에서 각각 1 을 내고 새 형태는 각각 0 을 낸다. **MU-125**(병합의 본문 줄이 인용한 `merge(t1453)` 를 `T` 가 셈)도 같은 저장소에서 실행했다 — 새 형태도 M2 에서 1 을 내므로 기계가 세는 수로는 막지 못하고, 목록(`git log --merges --format='%h %s' …`)이 제목 `merge(t1999): other` 를 찍어 `S` 가 되지 못함을 보인다. 이 층은 읽는 단계라 수용 잔여다. **읽기로 판정한 것**: 없다.

**이터레이션 6 이 스크래치 저장소에서 실행한 게이트 선택자 탐침(D40).** 워크트리 밖 scratchpad 에서 `git`(2.54.0, Apple Git-157; 작성자·날짜 고정, 전역 설정 끔 — 같은 스크립트가 같은 SHA 를 낸다)으로 저장소를 만들고 커밋 넷을 세웠다: **F1** 비-병합, 제목 `fix: unrelated`, 본문 줄 `merge(t1453): forged body line`(`9d34c363d`); **F2** 비-병합, 제목 `merge(t1453): x`(`45a99837a`); **M1** 두 부모 병합, 제목 `merge(t1453): WT-x into develop - real landing`(`f0aa47134`); **M2** 두 부모 병합, 제목 `merge(t1999): other`, 본문 줄 `merge(t1453): quoted body line in a real merge`(`f47bd6ff2`). 각 커밋에 `--no-walk` 로 `T` 를 걸었다 — 옛 형태는 `git rev-list --no-walk --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9]' <SHA>`, 새 형태는 같은 명령에 `--merges`. 구성 스크립트는 `progress.md` 이터레이션 6 기록에 있다.

| 커밋 | 부모 | 옛 `T` | 새 `T` | 기대 | 읽는 법 |
|---|---|---|---|---|---|
| F1 비-병합, 본문 줄 위조 | 1 | 1 | **0** | 0 | 새 형태가 본문 줄 위조를 거른다 |
| F2 비-병합, 제목 위조 | 1 | 1 | **0** | 0 | 새 형태가 비-병합 제목 위조를 거른다 |
| M1 진짜 두 부모 병합, 제목 `merge(t1453)` | 2 | 1 | **1** | 1 | 양성 대조 — 진짜 착지 형태는 센다 |
| M2 병합, 제목은 다른 카드·본문이 그 줄을 인용 | 2 | 1 | 1 | 센다(수용 한계) | 수로는 못 거른다. 목록이 제목 `merge(t1999): other` 를 찍어 `S` 가 되지 못한다(MU-125) |

저장소 전체(`HEAD`)에서 옛 `T` = 4, 새 `T` = 2, 새 `A` = 0 이고, 새 형태의 목록(`git log --merges --format='%h %s' -E -i --grep='^merge[( :]+(card )?t1453[^0-9]' HEAD`)은 두 줄을 낸다 — `f47bd6f merge(t1999): other`, `f0aa471 merge(t1453): WT-x into develop - real landing`. 두 번째 줄만 제목이 패턴으로 시작하므로 `S` 후보는 `f0aa471` 하나다. 한계: 이 값은 이 머신의 `git` 2.54.0 동작이다.

**이터레이션 5 가 실제로 실행한 변이 탐침.** 명령으로 실행해 관측한 것: (i) **MU-123**(B 앞의 곁가지에서 제품 커밋을 넣었다 되돌려 병합) — 스크래치 저장소와 이 카드의 실제 이력 위의 클론에서(아래 표): 옛 증인 5 는 통과하고 채택한 형태는 실패한다. (ii) 증인 5 의 후보 형태 셋(`--full-history` 단독, `--full-history --no-merges`, 채택한 `--full-history --simplify-merges`)을 같은 모양들에 걸어 비교했다 — 단독형은 이 카드의 실제 이력에서 정당한 M0 를 거짓으로 실패시키고(C42), `--no-merges` 는 병합 커밋의 충돌 해결 안에 제품 변경을 담은 변이(EVIL)를 통과시켜 새 쓰기 가능 변이를 만든다 — 둘 다 쓰지 않았다. **읽기로 판정한 것**(코드가 아직 없다): MU-106 의 소스 읽기 보강((f) (i)(ii))과 D36 의 목록 처분 문장.

**이터레이션 5 가 스크래치 저장소에서 실행한 증인 5 탐침(D34).** 워크트리 밖 scratchpad 에서 `git`(2.54.0, Apple Git-157)으로 저장소를 만들고 증인 5 의 세 형태 — 옛 기본 이력 단순화형(`옛`), `--full-history` 단독(`단독`), 채택한 `--full-history --simplify-merges`(`채택`) — 를 같은 `B^` 에 걸었다. 각 칸은 증인 1~3 이 통과한 뒤의 증인 5 판정과 세어진 수다(통과면 `PASS(0)`). 위반 모양은 B 앞에 카드 id 제품 커밋이 있는 것이고 "이 카드 실제 이력" 모양은 `68a4d8137` 의 클론(`git clone`)에 B 를 얹은 것이다. 모든 저장소는 `root` 커밋(`internal/seed.go`) 위에서 시작하고 B 는 `.moai/specs/S/baseline/baseline.md` 만 바꾼다.

| 모양 | 옛 | 단독 | 채택 | 진실 | 읽는 법 |
|---|---|---|---|---|---|
| S0 정상: B, M1, M2 가 한 줄 | PASS(0) | PASS(0) | PASS(0) | 정상 | 세 형태 모두 통과 |
| AUD1: 곁가지(`test: red t1454` 로 `internal/p.go` 추가, `revert red t1454` 로 삭제)를 B 앞에서 `merge(t1454): side` 로 병합, main 은 **제품 경로 파일**(`internal/other.go`, 카드 id 없음)로 갈라짐 | **PASS(0)** | FAIL(3) | FAIL(3) | 위반 | 옛 형태가 놓친다(D34). 3 = P + 되돌림 + 병합(병합이 곁가지 쪽 부모와 제품 경로에서 달라 목록에 든다) |
| AUD1b: AUD1 과 같되 main 이 **제품 경로 밖 파일**(`other.txt`)로 갈라짐(이터레이션 5 감사가 만든 모양) | **PASS(0)** | FAIL(2) | FAIL(2) | 위반 | 이터레이션 6 이 두 모양을 다시 만들어 `3`·`2` 둘 다 재현했다(D41) — 병합이 두 부모와 제품 경로에서 같아 목록에서 빠진다. 방향(옛 0, 채택형 실패)은 같다 |
| AUD2: AUD1 과 같되 병합 메시지에 카드 id 없음 | **PASS(0)** | FAIL(2) | FAIL(2) | 위반 | 같다 |
| AUD3: AUD1 과 같되 main 이 갈라지지 않음 | **PASS(0)** | FAIL(2) | FAIL(2) | 위반 | 같다 |
| 곁가지 P(순효과 있음)를 B 앞에서 병합 | FAIL(2) | FAIL(2) | FAIL(2) | 위반 | 옛 형태도 잡는다 |
| S1: P 가 B 보다 앞선 한 줄(MU-94) | FAIL(1) | FAIL(1) | FAIL(1) | 위반 | 변화 없음 |
| 정당한 곁가지가 B 뒤, 병합 | PASS(0) | PASS(0) | PASS(0) | 정상 | 변화 없음 |
| develop 흡수가 B 뒤(id 메시지) | PASS(0) | PASS(0) | PASS(0) | 정상 | 변화 없음 |
| 순서만 바꿔 B 가 먼저(그래프상 정상) | PASS(0) | PASS(0) | PASS(0) | 정상 | 변화 없음 |
| MU-110: 카드 id 없는 P 가 B 앞 | PASS(0) | PASS(0) | PASS(0) | 수용 잔여 | 세 형태 모두 놓친다(MU-110) |
| develop 흡수가 B 앞, main 에 제품 변경이 있어 병합이 양쪽 부모와 다름 | FAIL(1) | FAIL(1) | FAIL(1) | 정상 | 옛 형태부터 있던 안전한 방향의 거짓 실패 — 변화 없음 |
| **이 카드 모양**: 제품 변경 없는 카드 브랜치가 B 앞에서 develop 을 `merge: absorb … (card t1454)` 로 흡수 | PASS(0) | **FAIL(1)** | PASS(0) | 정상 | 단독형만 거짓으로 실패한다 |
| EVIL: 충돌 해결 안에 제품 변경을 담은 병합(메시지에 카드 id)이 B 앞 | FAIL(1) | FAIL(1) | FAIL(1) | 위반 | `--no-merges` 형은 `PASS(0)` 으로 통과시킨다 |
| **이 카드 실제 이력 클론**: `68a4d8137` 위에 B, M1 | PASS(0) | **FAIL(1)** | PASS(0) | 정상 | 단독형은 M0 를 실패시킨다(C42) |
| 실제 이력 클론: P 가 B 앞 직선 | FAIL(1) | FAIL(2) | FAIL(1) | 위반 | 변화 없음 |
| 실제 이력 클론: AUD1 형태 | **PASS(0)** | FAIL(4) | FAIL(3) | 위반 | 옛 형태가 놓친다 |
| 실제 이력 클론: B 뒤에 develop 을 다시 흡수 | PASS(0) | FAIL(1) | PASS(0) | 정상 | 단독형의 `1` 은 `B^` 에 든 `68a4d8137` 흡수 병합이다 |

증인 3 이 먼저 잡는 모양(B 뒤 곁가지 S2b·S2d, main 을 곁가지로 병합)은 증인 5 와 무관해 뺐다. 한계: 이 값은 `git` 의 동작이고 `--simplify-merges` 의 병합 선택은 이 머신의 `git` 2.54.0 에서만 관측했다.

**이터레이션 4 가 실제로 실행한 변이 탐침.** 명령으로 실행해 관측한 것: (i) **MU-114·MU-115**(순서 증인 3)을 스크래치 저장소에서 만들어 옛 형태와 새 형태를 함께 돌렸다(아래 표); (ii) **MU-116**(게이트 `A` 가 메시지 전체를 봄)의 형태를 실제 이력에서 — 본문에 `absorb` 가 든 착지(t1439)에 옛 `A` 형태가 `1`, 새 `A` 형태가 `0`(C33~C36); (iii) **MU-117·MU-118**(그래프 보기가 큐 락을 기다림·시도만 함)의 전제인 advisory flock 의 동작 — `python3` 의 `fcntl.flock` 으로 확인했다: 락을 쥔 동안 두 번째 열기 설명자의 블로킹 flock 은 1.0초 안에 돌아오지 못하고(`False`), `LOCK_NB` 시도는 즉시 실패하며, 락을 쥔 채 일반 읽기는 즉시 돌아오고, 락을 비어 있는 파일에서 잡았다 놓아도 락 파일의 바이트·mtime·크기는 그대로다(`mtime 1000000000 → 1000000000`, 크기 4 → 4). 이것은 Go 시험이 아니라 flock 의미를 확인한 것이다. **읽기로 판정한 것**: MU-119(경로 비교 방식)는 기준 문면에서, MU-120(기준선 짝 문장을 지움)은 `internal/template/card_id_leak_test.go` 의 `TestCardIDBaselineHasNoStaleEntries`(`:69`)를 읽고 현재 트리에서 그 시험이 통과함을 실행해 확인했다(변이 자체는 만들지 않았다), MU-121 은 수용 잔여다.

**이터레이션 4 가 스크래치 저장소에서 실행한 순서 증인 탐침(D26).** 워크트리 밖 scratchpad 에서 `git`(2.54.0, Apple Git-157)으로 저장소를 만들고 증인 3 의 옛 형태(개수, `B^..HEAD` 와 `--ancestry-path B..HEAD`)와 새 형태(`--full-history` SHA 목록 비교)를 `-- internal cmd pkg scripts .claude .codex` 와 함께 돌렸다. 모든 저장소는 `root` 커밋(`internal/seed.go`) 위에서 시작하고, 기준선 커밋 B 는 `.moai/specs/S/baseline/baseline.md` 만 바꾸고, 제품 커밋은 메시지에 `t1454` 를 담고 `internal/` 아래 파일을 바꾼다. "곁가지" 병합은 `git merge --no-ff` 이다. 아래 표의 `a/b` 는 두 질의의 개수(옛 형태) 또는 두 SHA 목록의 길이(새 형태)이고, 판정은 증인 3 이 통과(PASS)인지 실패(FAIL)인지다.

| 저장소(모양) | 옛 증인 3(개수 `a/b`) | 새 증인 3(`--full-history` 목록 `a/b`) | 기대 | 읽는 법 |
|---|---|---|---|---|
| S0 정상: B, M1, M2 가 한 줄 | `2/2` PASS | `2/2` PASS | PASS | 정상 이력은 둘 다 통과 |
| S8b 정상: B 뒤 곁가지 Q, main 의 M1, 곁가지를 `merge(t1454): …` 로 병합, M2 | `2/3` **FAIL** | `3/3` PASS | PASS | 옛 형태의 거짓 실패 — 기본 단순화가 병합 하나를 첫 질의에서 가린다 |
| S8c 정상: 곁가지 둘을 각각 `merge(t1454): …` 로 병합 | `4/5` **FAIL** | `5/5` PASS | PASS | 같은 거짓 실패 |
| S9 정상: develop 을 `merge: absorb local develop (card t1454)` 로 흡수(develop 커밋은 카드 id 없음) | `3/3` PASS | `3/3` PASS | PASS | 흡수 병합은 두 형태에서 같다 |
| S2 위반: 곁가지 P → B(main) → `merge side`(카드 id 없음) | `2/1` FAIL | `2/1` FAIL | FAIL | 병합 메시지에 카드 id 가 없으면 옛 형태도 잡는다 |
| S2b 위반(MU-114): S2 와 같되 병합 메시지가 `merge(t1454): side` | **`2/2` PASS** | `3/2` FAIL | FAIL | 옛 형태는 이 변이를 놓친다 — 개수만 같다 |
| S3b 위반(MU-114): B 가 곁가지에, P 가 main 에, 병합 메시지가 `merge(t1454): side` | **`2/2` PASS** | `3/2` FAIL | FAIL | 대칭 변이도 같다 |
| S2d 위반(MU-115): 곁가지에서 P 를 넣었다 되돌린 뒤 `merge(t1454): side` | **`1/1` PASS** | `3/1` FAIL | FAIL | 기본 단순화는 곁가지를 통째로 가려 `--full-history` 없는 목록 비교(`1/1`)도 통과한다 |
| S1 위반: P 가 B 보다 앞선 한 줄(MU-94) | `1/1` PASS | `1/1` PASS | (증인 5) | 증인 3 은 B 앞을 보지 못한다 — 증인 5 가 `1` 로 잡는다 |
| S10: develop 의 커밋이 메시지에 `t1454` 를 적고 B 뒤에 흡수됨 | `3/2` FAIL | `3/2` FAIL | FAIL(안전한 방향) | 다른 카드의 커밋이 카드 id 를 적으면 거짓 실패한다 — 저자가 그 커밋을 읽고 판정한다 |
| S11: B 뒤에 제품 커밋이 하나도 없음 | `0/0` FAIL | `0/0` FAIL | FAIL | "비어 있지 않다" 요구가 잡는다 |

재구성 방법의 핵심(S2b): `git init -b main`, `root` 커밋, `git checkout -b side`, 제품 커밋 P(`test: red t1454`, `internal/side.go`), `git checkout main`, 기준선 커밋 B(`docs: baseline t1454`), `git merge --no-ff -m "merge(t1454): side" side`, 제품 커밋 M1(`feat: M1 t1454`). 그 저장소에서 옛 형태는 `git rev-list --count --grep=t1454 B^..HEAD -- internal cmd pkg scripts .claude .codex`(`2`)와 `git rev-list --count --ancestry-path --grep=t1454 B..HEAD -- …`(`2`), 새 형태는 같은 두 질의에 `--full-history` 를 더하고 `--count` 를 뺀 목록(길이 `3` 과 `2`; 첫째는 M1·병합·P, 둘째는 M1·병합)이다. 한계: 이 값은 `git` 의 동작이라 저장소가 달라도 같지만 다른 `git` 버전에서의 `--full-history` 의 병합 선택은 이 머신 밖에서 재현하지 않았다.

**이터레이션 3 에서 실제로 실행한 변이 탐침.** 명령으로 실행해 관측한 것: (i) **MU-94 의 형태**: 대역 카드 t1460 으로 증인 5 형태가 앞선 제품 커밋 셋을 `3` 으로 세는 것(C24), 증인 3 형태가 같은 이력에서 `0`·`0` 을 내는 것(C25·C26), 이 카드 이력에서 증인 5 형태가 `0` 인 것(C23), 경로 필터가 정의의 일부인 것(C27); (ii) **MU-107 의 형태**: 흡수 병합 `b05c3be90…` 에서 `absorb` 를 가리지 않는 질의가 5(C28), 가리는 질의가 5 라 착지 후보 0(C29), 실제 착지 병합에서 후보 1(C30); (iii) **MU-108 의 형태**: 틀린 패키지에서 `-list` 가 이름 3개만 내고 올바른 패키지가 1개를 냄(G21·G22); (iv) **MU-103 의 메커니즘**: 값 형 구조체 필드의 `omitempty` 가 `{}` 를 쓰고 포인터 형은 생략한다는 것을 독립 Go 프로그램으로 관측했다(AC-TCI-008 의 변이 문단). **읽기로 판정한 것**(코드가 아직 없어 변이를 만들어 돌리지 못했다): MU-95~102·104~106·109 — 변이가 기준 문면을 만족하면서 요구를 위반할 수 있는지를 문면에서 판정했다. 이 판정은 실행 단계의 RED-now 관측과 변이 시험이 확인할 몫이다.

**이터레이션 2 에서 실제로 실행한 변이 탐침(이전 반복의 기록).** 코드가 아직 없으므로 대부분의 변이는 "변이가 기준을 만족하면서 요구를 위반할 수 있는가"를 읽어 판정하는 저작 시점의 탐침이다. 명령으로 실행해 확인한 것은 둘이다 — (i) **MU-58 의 변이(`--first-parent` 게이트)**: C19·C20 이 같은 선택자가 같은 고정 커밋에서 0 과 1 로 갈리는 것을 관측했고, (ii) **MU-78 의 변이(무시되는 경로)**: G14 가 옛 경로를 무시 규칙에 걸리게 하고 G13 이 디스크에 있는 같은 위치의 파일이 `git ls-files --error-unmatch` 로 종료 1 임을 관측했다. 나머지 변이의 "잡는 기준"은 실행 단계의 RED-now 관측이 확인할 몫이다.

**도착 시 통과하는데 수용하는 변이**(잡히지 않는 것과 이유):

| 변이 | 위반 | 잡히지 않는 이유와 수용 |
|---|---|---|
| MU-74 큐 쪽 baseline figure 의 값을 바꿔 적음 | REQ-TCI-001 | 비공개 DB 스냅숏이라 제3자가 재실행할 수 없다 — `queue-snapshot:` 귀속만 하고 재현은 주장하지 않는다 |
| MU-75 `--dry-run` 이 레인 세션에서 쓰이지 않음 | REQ-TCI-004 | `add` 는 레인 가드가 거절한다(읽기 전용 목록에 없다) — 관찰했고 정책 의도 여부는 미확인(공백 13) |
| MU-76 선행 미병합 후보가 오류가 아니라 건너뛰는 현재 동작이 이미 옳은 경우 | REQ-TCI-018 | M4 의 첫 시험이 현재 동작을 관측해 기록한다 — 읽기 추정이라 RED 일 수도 초록일 수도 있다 |
| MU-77 웹 그래프가 큰 큐에서 느림 | REQ-TCI-023 | 노드 상한과 결정성만 시험하고 브라우저 렌더 시간은 재지 않았다(공백 6) |
| MU-110 카드 id 가 없는 제품 변경 커밋이 기준선 앞에 착지 | REQ-TCI-001 | 증인 3·5 는 메시지의 카드 id 로 센다 — 막는 것은 커밋 메시지 규율이다(AC-TCI-001 잔여) |
| MU-111 실행 파일 위치나 `$HOME` 에서 경로를 만들어 프로젝트 파일을 읽는 허브 목록 적재 | REQ-TCI-020 | 정적·동작 검사는 작업 디렉터리와 리터럴만 본다 — 적재가 `go:embed` 한 함수라는 설계와 리뷰가 닫는다(AC-TCI-019 잔여) |
| MU-112 되돌려진 t1453 착지를 게이트가 열림으로 읽음 | REQ-TCI-021 | 이 이력에 되돌림 사례가 없어 측정하지 못했다(추정) — 판독 2 의 SHA 를 리더가 읽는다(AC-TCI-020 잔여) |
| MU-113 기존 JS 가 런타임에 외부 요청을 만듦 | REQ-TCI-023 | 정적 마크업·CSS 검사는 보지 못한다 — 새 JS 없음과 JS 목록 불변이 틈을 좁힌다(AC-TCI-022 한계) |
| MU-121 편집되는 다섯 템플릿 사본 중 하나에 숫자 임계값을 적음 | REQ-TCI-021 | 측정 수치 검사는 새 `card-issuance.md` 사본에만 건다 — 다섯 사본은 카드 id·중립성 가드까지만 기계가 보고 숫자는 M5 커밋의 diff 리뷰가 읽는다(AC-TCI-020 Mode A 문단, D31) |
| MU-122 어휘 검사를 피해 이름을 바꾸어 락을 시도만 하는 보기 | REQ-TCI-023 | (b3) 은 토큰만 본다 — `design.md` §9 의 "락·쓰기 없음"과 리뷰가 닫는다(AC-TCI-022 한계) |
| MU-125 병합 커밋의 본문 줄이 인용한 `merge(t1453): …` 를 게이트 `T` 가 셈 | REQ-TCI-021 | 한 번의 git 호출 안에는 줄이 제목인지 가리는 정규식이 없다 — `--merges` 가 비-병합 위조를 거르고(MU-124) 후보의 제목 위치는 판독 2 의 목록을 읽어 확인한다. 읽는 단계라 기계 판정이 아니다(AC-TCI-020 판독 1·2, 스크래치 M2); 핀 `68a4d8137` 의 실제 이력에는 그런 병합이 0개다 |

개수: 변이 125개 명명 — 114개는 기준이 잡고 11개(MU-74~77, MU-110~113, MU-121·122·125)는 이유와 함께 수용한다. 이터레이션 6 이 새로 이름 붙인 변이는 2개(MU-124·125)이고 MU-124 는 기준이 잡는다. 이터레이션 5 가 새로 이름 붙인 변이는 1개(MU-123)이고 기준이 잡는다. 이터레이션 3 이 새로 이름 붙인 변이는 20개(MU-94~113)이고 그 가운데 16개(MU-94~109)는 기준이 잡는다. 이터레이션 4 가 새로 이름 붙인 변이는 9개(MU-114~122)이고 그 가운데 7개(MU-114~120)는 기준이 잡는다. 이터레이션 3 의 MU-66 은 문구가 바뀌었다(기준이 잡는 쪽이 (b) 에서 (b2) 로).

## 경계 사례

- 한국어 본문처럼 공백 없이 이어진 토큰은 토큰 집합이 작아져 Jaccard 가 흔들린다 — 제시는 점수와 척도를 항상 함께 적는다.
- dropped 카드 본문의 접두사가 대시 변형(`—` 가 아닌 `-`)이면 `stripTodoDropMarker` 가 벗기지 못할 수 있다 — 벗기지 못한 본문은 접두사를 단 채 비교되고 점수가 낮게 나온다(보수적).
- 정규화 본문이 같은 카드가 live 와 보관 양쪽에 있으면 둘 다 `exact` 로 보이되 상위 3개 상한은 지킨다.
- git 이 없거나 탐침이 실패하면 해당 항목이 `unmeasured` 이고 admit 은 성공한다.
- `--dry-run` 과 `--pick`/`--force` 를 함께 주면 dry-run 이 우선이고 쓰기는 없다(둘 중 하나가 무시됐음을 stderr 에 한 줄 적는다).
- 폴스루(`moai todo <단어…>`)는 플래그를 받지 않으므로 `--dry-run` 과 발행 플래그는 `add` 서브커맨드에서만 쓴다.
- 워크트리 세션에서도 큐 루트는 primary 체크아웃으로 해석되므로 제시는 같은 큐를 읽는다.
- 시험은 라이브 `moai todo add` 가 아니라 `go test` 도구로만 한다(레인 가드가 라이브 호출을 거절한다). 이 SPEC 의 새 시험은 레인 변수를 스스로 지우고(`sdClearLaneEnv` 형), 레인 변수를 지우지 않는 기존 시험을 팩토리 레인 세션에서 돌릴 때는 plan §D 의 SCRUB 접두사가 필요하다(G24·G25).
- 새 카드의 `--parent` 가 보관 카드를 가리키는 것은 정당하다(후속은 닫힌 카드에서 나온다) — 거절 대상은 큐에 존재하지 않는 id 뿐이다.
- 얕은 클론에서는 허브 목록의 측정 명령 시험이 고정 SHA 를 찾지 못해 건너뛴다 — 건너뜀은 공백으로 기록하고 통과로 세지 않는다. `CI` 환경 변수가 설정된 실행(CI 의 test 작업은 `fetch-depth: 0`)에서는 건너뛰지 않고 실패한다(AC-TCI-019 (f)).
- 겹침 제시: 후보가 경로를 대고 비교할 진행 중 카드에 입력이 있는데 겹침이 없으면 줄을 쓰지 않는다(`none`, 측정됨). 후보에 경로가 없거나 비교할 진행 중 카드에 입력이 하나도 없으면 `unmeasured` 줄을 쓴다 — 진행 중 레인 카드가 아예 없는 경우도 후자다. 입력이 있는 진행 중 카드와 없는 카드가 섞여 있으면 판정은 집합 단위다 — 입력이 있는 카드만 비교하고 입력이 없는 카드는 `none` 으로도 `unmeasured` 줄로도 보고하지 않는다(AC-TCI-003 (g)). 경로 일치는 정규화한 경로 전체의 일치이고 파일 이름이나 접두사의 일치가 아니다(AC-TCI-003 (h)).

## 품질 게이트

- TRUST 5: 시험을 먼저 쓰고 RED 를 관측한 뒤 GREEN(회귀 가드 AC-TCI-007·009·023 만 예외), 변경 함수 커버리지 ≥ 85%, CI 가 쓰는 버전의 `golangci-lint` 청결, `GOOS=windows GOARCH=amd64 go build` 를 변경 패키지에 한해 통과.
- 문서 게이트: `moai spec lint SPEC-TODO-CARD-ISSUANCE-001` 과 `--strict` 청결(plan 단계), M5 Mode A 의 템플릿 가드 전부 통과.
- MX: 새 고팬인 함수(관계 정규화·제약 검증, 읽기 전용 이웃 조회)에 `@MX:ANCHOR`(`@MX:REASON` 필수), 탐침 병렬 코드가 있으면 `@MX:WARN`, 상수에 `@MX:NOTE`.

## 완료 정의

1. AC-TCI-001~006, -008, -010~020, -022, -024(출시 차단 20개)가 통과하고 RED-now 셀이 관측되고 green-path 출력이 `progress.md` §E.2 에 기록된다. AC-TCI-021(조건부)은 Mode A 에서 통과하거나 Mode B 에서 "해당 없음(Mode B)"으로 기록되며 통과로 세지 않는다(AC-TCI-020 이 모드를 보증한다). AC-TCI-007·009·023 은 회귀 가드로 유지된다.
2. **M0 기준선 커밋이 어느 제품 변경 커밋보다 앞서 커밋 그래프에 있다**(`verification-claim-integrity.md` §2.3). 증인은 커밋 그래프이고 읽는 명령은 AC-TCI-001 의 순서 증인 1~5 다 — 기준선 커밋 `B` 를 `git log --diff-filter=A --format=%H -- .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md` 로 찾고, `git show --name-only --format= <B> -- . ':!.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline'` 이 비어 B 가 자기 커밋임을 보이고, `git rev-list --full-history --grep=t1454 <B>^..HEAD -- internal cmd pkg scripts .claude .codex` 와 `git rev-list --full-history --ancestry-path --grep=t1454 <B>..HEAD -- internal cmd pkg scripts .claude .codex` 가 **같은 SHA 목록**이고 비어 있지 않음을 보이고(개수가 아니라 목록이다 — 두 개수가 같아도 집합은 다를 수 있다, MU-114; B 뒤의 제품 커밋이 모두 B 의 후손), `git diff --name-only <기준선 머리의 card-head: SHA> <B>^ -- internal cmd pkg scripts .claude .codex` 가 비어 측정한 트리가 B 의 부모와 같음을 보이고, **`git rev-list --full-history --simplify-merges --count --grep=t1454 <B>^ -- internal cmd pkg scripts .claude .codex` 가 `0` 이어서 B 보다 앞선 제품 커밋이 없음을 보인다**(앞의 두 목록은 B 앞을 보지 못한다 — C25·C26; 제품 커밋은 카드 id 를 담고 `internal`·`cmd`·`pkg`·`scripts`·`.claude`·`.codex` 를 바꾸는 커밋이고 `.moai/` 만 바꾸는 SPEC·기준선 커밋은 세지 않는다). 기준선이 gitignored 경로(`.moai/reports/`)에 있으면 이 증인이 존재할 수 없으므로 그 배치는 거절한다(MU-78). 순서 증인 1~5 는 병합 전에 한 번만 읽는다 — 마지막 마일스톤이 끝난 뒤, 카드 브랜치가 통합 브랜치에 병합되기 전(sync-audit 이 증거를 읽는 시점). M0 종료에서는 읽지 않는다: 증인 3 이 B 뒤의 제품 커밋 목록이 비어 있지 않을 것을 요구하는데 M0 종료에는 제품 커밋이 하나도 없어 두 목록이 비기 때문이다(스크래치 S11). M0 종료 증거는 AC-TCI-001 (a)~(e) 다 — 기준선 파일과 그 안의 줄만 읽으므로 제품 커밋이 없는 트리에서 판정된다.
3. 각 마일스톤의 시험 선택 수(`-list`)와 `git grep -c -F` 짝이 `progress.md` §E.2 에 있다.
4. M5 커밋 본문(Mode A)에 상시 로드 전후 바이트·문자와 이 규칙이 필요 없는 세션의 비용 문장이 있다.
5. 완료 보고가 보존된 분기(`factory-dispatch.md` 한 문장, `sync-auditor.md` 18줄), 게이트 모드(A/B)와 Mode B 의 경우 "요구는 충족했으나 규칙은 아직 효력이 없다"는 사실, 그리고 열린 잔여 위험(spec §G)을 명시한다.
6. `TestACCounterFullCorpusMatchesBaseline` 이 통과한다(이 SPEC 의 acceptance.md 는 스냅숏에 없어 보고만 되고 실패하지 않는다; 기준선 재생성이 필요하면 그것은 plan 커밋의 몫이다).
