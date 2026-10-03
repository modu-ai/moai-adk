# SPEC-TODO-CARD-ISSUANCE-001 — 계획

Tier L. 카드 t1454, 계획 시작 HEAD `2de0a2cb6`(전체 SHA `2de0a2cb613b04765a1554f86685a3b48e0be806`; 이터레이션 2 는 그 위에 SPEC 첫 커밋 하나를 얹은 `1894984c3` 에서, 이터레이션 3 은 plan 커밋 둘을 얹은 `ad02a5677`(전체 SHA `ad02a56779473afc2ef72e9ecfe16d517a297a7b`)에서 다시 쟀고, 이터레이션 4 는 그 위에 SPEC 디렉터리 편집 커밋 하나를 얹은 `8fff427eb`(제품 경로 변화 없음 — `acceptance.md` G26)에서 모든 행을 다시 돌렸다 — 문서 수준 핀은 `acceptance.md` 머리), 워크트리 `.claude/worktrees/t1454`, 브랜치 `WT-card-issuance-overlap-graph`. 개발 방식은 TDD(`.moai/config/sections/quality.yaml` 의 `constitution.development_mode: tdd`)다 — 마일스톤마다 RED 테스트를 먼저 커밋하고 GREEN 으로 넘긴다. 시간 추정은 쓰지 않는다. 우선순위 라벨과 순서("A 를 마치고 B 를 시작")만 쓴다.

문서 구성: 결정이 바뀔 가능성이 큰 순서(§F.0)가 먼저 나오고, 마일스톤은 의존 순서(실행 순서)로 이어진다. 기계적인 작업(M0 의 측정 복사, M5 의 미러 정합)은 뒤에 있다.

## §A 배경

- **Epic 참조.** 운영자가 선언한 Epic 은 없다. 이 SPEC 은 단독 SPEC 이다(`.claude/rules/moai/development/sprint-round-naming.md` 의 단일 SPEC 허용 형태). 같은 주제(카드 큐의 품질)의 인접 SPEC 은 SPEC-TODO-AUTO-PICK-001(completed, 카드 t1448)과 SPEC-GITHUB-FLOW-DEFAULT-001(in-progress, 카드 t1453, **이 트리에는 없다** — `WT-github-flow-default` 브랜치에만 있다).
- **용어.** 마일스톤은 M0~M6 이다(`Milestone`). M5 만 게이트가 있고 나머지는 의존 순서로 실행한다.
- **무엇을 하는가.** `spec.md` §A.2 와 §C.
- **측정 근거.** `research.md` §3. 카드가 인용한 git 쪽 수치는 이 트리에서 전부 재현됐고, 큐 쪽 수치는 스냅숏 위에서 다시 쟀으며, 발행 시점 제시의 설계를 바꾼 새 측정 두 가지가 있다 — 어휘 겹침이 약한 변별자라는 것(SB01~SB05), 예상 파일 입력이 거의 비어 있다는 것(SB06~SB07).

## §B 알려진 문제

1. **낡은 전제 여덟 가지**는 `spec.md` §A.3 에서 바로잡았다(near-duplicate 기록 위치, t1349 완료, `assign --after` 의 정체, 병합 금지 문면의 위치, t1448/t1453 착지 상태, 추적되지 않는 `edges.jsonl`, 카드가 재사용하라던 두 도우미를 이 SPEC 이 재사용하지 않는 것, 기준선·초안 경로가 무시된다는 것).
2. **`assign --after` 의 구조적 한계.** 이전 카드의 병합 가드는 할당 간선(`homestate/card_transition.go` 의 `guardAssign`)에서 돈다. 이전 카드가 병합되기 전에는 다음 카드를 레인에 할당할 수 없어서, "같은 레인에 다음 카드를 미리 걸어 둔다"는 지금의 도구로 표현되지 않는다(`research.md` §5.4).
3. **선택 호의 오류 처리(읽기 추정, 미확인).** 선행 미병합 `after` 후보를 만난 임대 선택 호가 건너뛰지 않고 오류로 끝날 수 있다 — M4 가 첫 RED 테스트로 확인한다(`research.md` §10 항목 9).
4. **규칙 파일의 예산과 분기.** `gtd.md` 40,037자(37자 초과), `kanban-dispatch-detail.md` 43,138자(3,138자 초과), `kanban-dispatch.md` 는 상시 로드 28,092자. 로컬과 템플릿의 `kanban-dispatch.md`·`sync-auditor.md` 사본은 이미 갈라져 있고 어느 가드에도 등록돼 있지 않다.
5. **같은 파일을 만지는 다른 카드**: t1453(절체 시점에 같은 규칙 파일을 고친다, 미착지), t1450·t1452·t1359(대기) — `research.md` §8.
6. **레인 가드.** 레인 세션은 `add --dry-run` 도 거절된다(`todoRefuseLaneMutation` 의 읽기 전용 동사 목록에 `add` 가 없다).
7. **`.moai/reports/` 는 추적되지 않는다.** `.gitignore:235`(`.moai/reports/*`)가 그 아래 모든 경로를 막으므로(`git check-ignore -v .moai/reports/t1454/baseline/x.txt` 가 종료 0), 이터레이션 1 이 "추적되는 형태"라 부른 기준선은 어느 클론·CI 도 볼 수 없었고 "기준선 커밋이 변경 커밋보다 앞선다"는 순서는 커밋 그래프가 증언할 수 없었다. 감사 산출물 규약이 보고서를 트리에 강제로 넣거나 무시 규칙을 넓히는 것을 금하므로, 기준선과 M5 초안은 추적되는 SPEC 디렉터리로 옮겼다(`design.md` §12).
8. **`--first-parent` 게이트는 흡수 뒤 열리지 못한다.** 카드 브랜치가 `git merge develop` 으로 develop 을 흡수하면 develop 의 병합 커밋은 흡수 병합의 두 번째 부모 쪽에만 있다. 이터레이션 1 의 게이트가 그것을 세지 못함을 고정 SHA 로 보였다(`acceptance.md` C17~C20; `design.md` §13).

## §C 사전 점검 (M0 시작 전, 결과는 `progress.md` §E.2 에 기록)

1. 핀과 청결: `git rev-parse HEAD` 가 계획 시작 SHA 와 같은지(다르면 재측정·재고정), `git status --short` 가 비었는지.
2. 판정 도구를 트리에서 빌드해 **경로로** 호출한다: `go build -o <scratch>/moai ./cmd/moai`. 설치된 `moai` 는 쓰지 않는다 — 측정 인용에는 트리 HEAD 와 판정 빌드의 커밋을 함께 적는다(`verification-claim-integrity.md` §2.2).
3. 병렬 세션 점검(`gitflow-lane-protocol.md` §11 의 비교): `git fetch origin develop` 가 끝난 뒤 `git rev-list --count --left-right origin/develop...develop`. 카드 브랜치는 develop 을 병합으로 흡수한 뒤(`git merge develop`, 카드 트리 안에서) 측정을 다시 고정한다.
4. M5 게이트 판독을 한 번 읽어 둔다(§F.8, 다섯 판독). 지금 읽으면 게이트는 닫혀 있다(t1453 병합 수 0, t1448·t1344 대조군 1·1, 두 번째 부모 대조 333 > 214).
5. 큐 스냅숏을 scratchpad 로 복사하고(원본 DB 를 읽기 전용 플래그로 여는 것은 샌드박스에서 실패했다) 기준선 스크립트를 복사한다.
6. 레인 환경: 이 트리의 세션은 레인 표지가 있어 라이브 `moai todo add` 는 거절된다. add 경로 검증은 `go test` 로만 한다(이 SPEC 의 새 시험은 `sdClearLaneEnv` 형으로 레인 변수를 스스로 지운다; 지우지 않는 기존 시험은 §D 의 SCRUB 접두사가 필요하다).
7. 계획 단계에서 이미 한 일: `TestACCounterFullCorpusMatchesBaseline` 이 새 SPEC 을 "absent-from-snapshot"으로 보고하고 실패하지 않음을 확인한다(`progress.md` §G).
8. 경로 점검: M0 가 쓸 경로가 무시되지 않는지 `git check-ignore -v .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md` 가 빈 출력·종료 1 인지 읽는다(옛 경로 `.moai/reports/t1454/baseline/x.txt` 는 종료 0 이어야 대조가 선다).

## §D 제약

- **허브 파일 직렬.** 이 카드 자신의 실행에서도 `internal/cli/todo.go`(M1·M2·M3·M4 가 모두 만진다)와 `internal/template/catalog.yaml` 은 허브다. 마일스톤을 순서대로 한 작성자가 진행하므로 겹치지 않는다. `catalog.yaml` 은 템플릿 아티팩트를 고친 마일스톤마다 같은 커밋에서 생성기로 다시 만든다.
- **순서 귀속.** 기준선 산출물(M0)은 **자기 커밋**으로 착지하고 그것이 측정한 변경 커밋보다 앞서야 한다(`verification-claim-integrity.md` §2.3). 커밋 그래프만이 순서의 증인이고, 증인 명령은 `acceptance.md` AC-TCI-001 의 순서 증인 1~5 와 완료 정의 #2 가 이름으로 대며 병합 전에 읽는다 — 증인 3 은 두 질의의 **SHA 목록**이 같고 비어 있지 않은지를 읽고(`--full-history`; 개수의 일치는 집합의 일치가 아니다 — 이터레이션 3 감사 D26), 증인 5(`B^` 에서 닿는 이력의 제품 커밋 수 0)가 B 보다 앞선 제품 커밋을 보는 유일한 증인이다. 기준선이 `.moai/reports/` 에 있으면 이 증인이 존재할 수 없다.
- **검증은 마일스톤이 만지는 패키지로 한정**한다. `go test ./...` 를 돌리지 않는다(`gitflow-lane-protocol.md` §8). 시험 명령은 레인 변수를 한 번의 복합 호출에서 지운다: `SCRUB = unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND &&`. 앞의 다섯은 kanban 레인의 형태(`kanban-dispatch.md` 의 env-isolated 검증 형태)이고 뒤의 넷은 팩토리 레인이 레인 가드(`factoryLaneRefusal`: `MOAI_FACTORY_ROLE`·`MOAI_FACTORY_WORKER`·`MOAI_KANBAN_BACKEND`)에 쓰는 변수다 — 이터레이션 3 이 팩토리 레인 세션에서 앞의 다섯만 지운 호출은 `TestRelateAndUnrelateRefusals` 를 `lane boundary` 거절로 붉게 만들고(자기 환경을 지우지 않는 기존 시험), 뒤의 넷까지 지운 호출은 통과시키는 것을 관측했다(`acceptance.md` G24·G25). 레인 변수를 스스로 지우는 새 시험(`sdClearLaneEnv` 형)은 이 접두사에 기대지 않는다. 모든 시험 이름은 `-run` 에서 하나씩 `^TestName$` 로 앵커하고(여러 개는 `^TestA$|^TestB$` 로 가지마다 쓴다), 같은 패턴에 `go test <패키지> -list '<패턴>'` 를 짝지어 출력된 이름 수가 지정한 시험 수와 같은지 본다 — 0개를 고르고도 `ok` 가 나오는 선택자를 증거로 쓰지 않는다(`verification-completeness.md` §1.1). 이터레이션 2 가 이 다중 가지 형태를 기존 시험에서 관측했다(`acceptance.md` G3).
- **예산.** 상시 로드 `kanban-dispatch.md` 는 순증가 0, 건드린 모든 규칙 파일은 40,000자 이하이거나 이전보다 크지 않아야 한다(D15). 1,000바이트를 넘는 상시 로드 증가는 `rule-authoring.md` 의 크기·비용 진술을 커밋 본문에 쓴다.
- **템플릿 가드.** 새 `t####` 를 템플릿 사본에 넣지 않는다(`internal/template` 의 `internal_content_leak_test.go` 가 `card_id_leak_test.go` 의 (파일, 카드 id) 기준선으로 훑는다 — `TestTemplateNoInternalContentLeak`), 템플릿 사본에는 SPEC·REQ id·날짜·9자 이상 16진 단어가 없어야 한다(`internal/cli` 의 `todo_skill_doc_parity_test.go` 의 중립성 정규식 — `internal/template` 이 아니다 — 바이트 동일 쌍 전체에 사실상 적용), `kanban-dispatch-detail.md` 의 `paths:` 는 바꾸지 않는다(`internal/template` 의 `workflow_rule_paths_pinned_test.go` — `TestWorkflowRulePathsPinned`). 새로 더하는 배포 규칙 사본은 추가로 `.moai/` 경로와 측정된 수치를 담지 않는다(`TestCardIssuanceTemplateIsMechanismOnly`) — 수치는 로컬 전용 규칙에만 쓴다. 이 시험의 수치 검사는 새 `card-issuance.md` 사본 하나에만 걸린다: 편집되는 다섯 사본은 카드 id·중립성 가드까지만 기계가 보고 숫자 임계값은 M5 커밋의 `git diff` 리뷰가 읽는다(수용 잔여, `acceptance.md` MU-121). **카드 id 기준선 가드**: `internal/template/card_id_leak_test.go` 의 `TestCardIDBaselineHasNoStaleEntries` 는 기준선에 오른 (파일, 카드 id) 짝의 id 가 그 파일에서 사라지면 붉다(`:69`; 이터레이션 4 가 읽고 현재 트리에서 통과함을 실행해 확인했다). M5 가 만지는 파일에 걸린 짝은 넷이다 — `gtd.md t696`(`:244`), `kanban-dispatch.md t1330`(`:94`), `kanban-dispatch-detail.md t133`·`t224`. **M5 의 압축은 그 짝의 문장을 지울 수 있다**(특히 `kanban-dispatch.md:94` 의 "measured precedent … t1330" 문장과 `gtd.md:244` 의 `(t696)` 절은 증거 기록이라 압축 후보다). 지우면 같은 커밋에서 `cardIDBaseline` 의 해당 항목을 지워야 하고(제거 전용 기준선이라 삭제는 허용되는 방향이다), `gtd.md t696` 짝이면 `TestCardIDBaselineIsPerFileAndPerLiteral` 이 그 짝을 양성 대조 상수(`:98`)로 쓰므로 다른 짝으로 고쳐야 한다 — 그래서 `card_id_leak_test.go` 가 M5 의 조건부 파일이다(`spec.md` §E). 문장을 지우지 않으면 이 파일은 건드리지 않는다. 새 `t####` 는 어느 경우에도 더하지 않는다.
- **출하 코드는 SPEC·reports 경로를 읽지 않는다.** 사용자 프로젝트에는 그 경로가 없다. 허브 목록처럼 기준선이 만드는 출하 입력은 임베드되는 데이터 파일이고, 시험이 기준선의 추적되는 사본·측정 명령과 대조한다(`design.md` §12).
- **레인 가드.** 새 읽기 동사 `trace` 는 `todoLaneReadOnlyVerbs` 에 넣고, `merge`·`bundle` 은 넣지 않는다.
- **분석 불변식.** 분석기·`analyze`·`relate` 는 카드를 접지 못한다 — 코드 모양과 테스트로 계속 강제한다(`TestTodoMergeNeverInvokedByAnalysis`).

## §E 자기 검증

- `moai spec lint SPEC-TODO-CARD-ISSUANCE-001`, `--strict` — 명령과 종료 코드는 `progress.md` §G 에 기록한다.
- 요구·기준 개수: 요구 24 ≤ 25, 기준 24 ≤ 25, 모듈 5 ≤ 5(Tier L 상한). 모든 요구가 기준 하나 이상에서 인용된다. 기준 분류: 출시 차단 20, 조건부 1(AC-TCI-021), 회귀 가드 3(AC-TCI-007·009·023).
- RED-now: 출시 차단 20개 기준은 각각 원장 행을 가진다(`acceptance.md`); 조건부 기준도 원장 행(L24, L36)을 가진다.
- SPEC ID 점검은 Bash 로 실행해 `PASS` 를 출력했다.

## §F 마일스톤

### F.0 결정이 바뀔 가능성 순서 (가장 바뀔 것부터)

| 순위 | 결정 | 바뀌면 다시 해야 하는 것 | 마일스톤 |
|---|---|---|---|
| 1 | D5 MCP 전달·engage 합류, D8 유사도 척도·소음, D13 플래그 입력 방식(사용자가 직접 보는 흐름) | 제시 렌더러·플래그 파싱·테스트 기대값 | M1, M2 |
| 2 | D4 스키마 형태, D9 closed-at, D10 부모 원천, D11 처분 효과(데이터 모형) | 컬럼·읽기/쓰기 목록·동결 튜플 문자열·JSON 투영 | M2 |
| 3 | D2 저장소 통합, D3 간선 운반체, D14 묶음 메커니즘(새 형식·인터페이스) | 해석기·간선 층·팩토리 `cards` 컬럼과 선택 호 | M3, M4 |
| 4 | D1 M5 순서, D6 병합 독트린 범위, D12 부채 대장 운반체(독트린) | 규칙 문면·개정 행 | M4, M5 |
| 5 | D7 수치 임계값(값은 기준선에서, 값의 자리는 로컬 전용 규칙) | 값 하나가 바뀌면 기준 리터럴과 로컬 규칙 한 곳만 | M0 에서 확정 |
| — | D15(이미 정해짐) | — | M5 |

### F.1 실행 순서 (의존)

M0 → M1 → M2 → M3 → M4 → M6 → M5(게이트). 한 마일스톤을 마치고 다음을 시작한다.

- M1 은 기존 저장소만으로 제시를 만든다. 진행 중 레인 겹침 입력은 예상 파일 필드가 생기면(M2) 측정값이 되고, 그 전에는 `unmeasured` 로 나간다(REQ-TAU-013 의 열린 입력 집합과 같은 모양).
- M2 가 속성과 처분을 추가한다. M3 은 M2 의 부모·출처 속성을 읽어 `parent-of`/`follow-up-of` 를 투영한다. M4 는 M2 의 예상 파일과 M3 의 `merged-into` 를 쓴다. M6 은 M3 의 해석기를 읽는다.
- M5 는 t1453 착지에 걸려 있어 마지막에 둔다(D1). M6 은 M5 와 독립이다.

### F.2 M0 — 기준선 반입과 재측정 (Priority High, 기계적)

- **목적**: 카드의 측정과 이 계획의 새 측정을 **추적되는** 위치에 옮기고 실행 단계에서 다시 잰다. 이후 모든 임계값은 이 기록에서 읽는다(REQ-TCI-001). 출하되는 허브 파일 목록도 같은 커밋이 만든다(REQ-TCI-020).
- **선행**: §C 사전 점검(특히 항목 8 의 무시 규칙 점검).
- **파일(모두 신규, 모두 `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/` 아래)**: `baseline.md`(figure 행·개수 행·임계값 표), `scripts/`(카드의 `lib.py`·`01`~`06`, 이 계획의 부록 A 다섯 개), `README.md`(재현 절차), `hub-files.txt`(허브 경로 목록). 큰 원자료 TSV(`fp.tsv`·`allc.tsv` 등)는 추적하지 않는다 — 큐 본문이 섞인 스냅숏이고 크기가 크므로 figure 행과 스크립트만 추적하며 README 의 절차로 다시 만든다.
- **작업**: (1) 카드 스크립트를 복사하고 원본이 사라졌다면 `research.md` §3.1~§3.4 의 방법 서술로 재구성한다. (2) `fp.tsv`·`allc.tsv` 를 실행 시점 트리에서 다시 만들고 스크립트를 돌려 `research.md` §3 의 figure 35개(GB 16 + QB 10 + SB 9)를 다시 잰다. 각 행은 `figure:`, `command:`, `output:`, `tree:` 줄을 가진다. git 이력 기반 figure(GB)는 카드 브랜치 HEAD 가 아니라 카드 브랜치가 흡수한 **develop 팁 SHA** 에서 `--first-parent` 로 읽고 그 SHA 를 `tree:` 로 기록한다 — 통합 브랜치 자체의 first-parent 는 맞지만 흡수한 카드 브랜치의 first-parent 는 흡수된 develop 병합을 세지 못한다(§13). (3) 핵심 git 쪽 헤드라인이 `research.md` §3.1 과 다르면 `drift:` 줄에 이유를 적는다. (4) **개수 figure 마다 그 개수를 낸 명령을 바로 옆 `command:` 줄에 적고 읽은 시각을 붙인다** — 최소한 `git branch --list 'WT-*' | wc -l`, `git worktree list | wc -l`, `ls .moai/specs | wc -l`, `git grep -l '^card:' -- '.moai/specs/*/spec.md' | wc -l`. 이 값들은 저장소 전역 ref 이거나 작업 트리의 디렉터리 수라 트리의 성질이 아니고 읽을 때마다 움직인다(353·74 → 367·89, SPEC 디렉터리 1,029 → 1,034 → 1,031 — `spec.md` §G) — 기준선은 "측정 시각의 값"으로 적고 `tree:` 줄에 이 값들을 귀속하지 않는다. 기준선 머리에 `card-head:` 줄(측정을 시작한 시점의 카드 브랜치 끝 `git rev-parse HEAD`)을 한 줄 둔다 — 순서 증인 4 가 읽는 값이고, 줄 이름에 `figure:`·`command:`·`tree:` 를 부분 문자열로 담지 않는다. (5) `baseline.md` 끝에 "임계값" 표를 둔다 — §F.11 의 작업 기본값을 재측정한 분포에서 다시 도출하고 각 값에 측정 명령을 붙인다. (6) 허브 파일 목록(만진 서로 다른 카드 수 순위)과 허브 임계값을 정하고 `hub-files.txt` 를 만든다: 머리 주석에 develop 팁 SHA(`tree:`)·기간·경로별 측정 명령의 형태(`git log --first-parent --since=<S> --until=<U> --format=%H <develop 팁 SHA> -- <path>` 의 출력 줄 수)를 적고, 각 줄은 `<경로>` 와 기록 개수를 탭으로 나눈다.
- **RED/GREEN**: RED 는 `git ls-files --error-unmatch .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md` 가 종료 1(원장 L1). GREEN 은 같은 명령이 경로를 출력하고 종료 0 이며 `figure:`·`command:`·`tree:` 줄 수가 같다.
- **검증 명령**: `git ls-files --error-unmatch .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md`, `git grep -c -F "figure:" -- .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md`, 같은 형태로 `command:`, `tree:`; 세 수가 같고 35 이상이다. 순서 증인 1~5(AC-TCI-001; 증인 3 은 `--full-history` 두 질의의 SHA 목록이 같고 비어 있지 않음, 증인 5 는 `git rev-list --count --grep=t1454 <B>^ -- internal cmd pkg scripts .claude .codex` 가 `0`).
- **종료 증거**: AC-TCI-001. **되돌리기**: 커밋 되돌리기(코드 변경 없음). **커밋**: 이 마일스톤의 커밋은 위 파일들**만** 담는 단독 커밋이고 다른 어떤 실행 커밋보다 앞선다 — 같은 커밋에 코드나 시험을 섞으면 순서 증인 2 가 깨지고(MU-79), 카드 `t1454` 의 제품 커밋을 이 커밋 앞에 착지시키면 증인 5 가 깨진다(MU-94; 제품 커밋은 카드 id 를 담고 `internal`·`cmd`·`pkg`·`scripts`·`.claude`·`.codex` 를 바꾸는 커밋이며 SPEC 디렉터리만 바꾸는 plan 커밋은 세지 않는다).

### F.3 M1 — 발행 시점 제시 (Priority High, 사용자가 보는 흐름이라 바뀔 가능성이 가장 크다)

- **목적**: `add` 가 막지 않고 알린다(REQ-TCI-002~006).
- **파일**: `internal/kanban/backlog_issuance.go`(신규, 읽기 전용 이웃·구성요소·완료 SPEC 조회와 표시 하한·상한 상수), `internal/cli/todo_issuance.go`(신규, 제시 조립과 렌더, 탐침 시간 상한), `internal/cli/todo.go`(`scanTodoAddArgs`·`todoAddScan` 에 `--dry-run`, 잠금 밖에서 제시를 계산·출력), `internal/cli/mcp_todo.go`(결과 텍스트에 제시 덧붙임), `internal/cli/gtd.go`(engage 가 제시만 출력), 각 `_test.go`.
- **RED 테스트(먼저 쓴다)**: kanban — `TestIssuanceNeighborsIncludeArchivedAndDropped`, `TestIssuanceNeighborsStripDropPrefix`, `TestIssuanceNeighborsLimitAndFloor`, `TestIssuanceSameComponentOpenCards`, `TestIssuanceCompletedSpecCoverage`, `TestIssuanceInFlightOverlapUnmeasured`, `TestIssuanceInFlightOverlapReportsSharedPath`, `TestIssuanceInFlightOverlapNoneIsMeasured`. cli — `TestTodoAddPresentationStderrOnly`, `TestTodoAddPresentationShowsInFlightOverlap`, `TestTodoAddPresentationNeverBlocks`, `TestTodoAddPresentationProbeOutsideLock`, `TestTodoAddPresentationTimeBound`, `TestTodoAddPresentationDryRunWritesNothing`, `TestTodoAddDryRunFlagParsed`, `TestTodoAddMCPCarriesPresentation`, `TestGTDEngagePresentationRecordsNothing`.
- **구현 요점**: 이웃 조회는 `LoadPure` 스냅숏을 입력으로 받는 순수 함수이고 `ClassifyCardText` 를 쓰지 않는다(분류기는 dropped 를 건너뛰는 독트린을 지킨다). dropped 본문은 `[DROPPED — ` 접두사를 벗기고 사유는 따로 보인다. 보관 카드는 `rec.Archived` 에서 읽는다(추가 질의 없음). 출력은 stderr 에 상위 3개를 점수·척도·상태와 함께 적고, 표시 하한 미만의 어휘 이웃은 `--dry-run` 에서만 보인다(D8 작업 기본값). 진행 중 레인 겹침은 `rec.Runtime.Assignments` 의 레인 카드와 예상 파일을 비교한다 — 예상 파일은 `spec.md` Module A 의 용어(본문이 이름으로 대는 경로 ∪ M2 이후 `files` 속성, 진행 중 카드는 시간 상한 안의 레인 브랜치 변경 파일을 더한다)이고, 후보에 입력이 없거나 비교할 진행 중 카드에 입력이 하나도 없으면 `unmeasured`, 입력이 있는 비교에서 공유 경로마다 `measure=file-overlap` 항목, 겹침이 없으면 `none`(줄을 쓰지 않는다)이다. 판정은 진행 중 카드 집합 단위(입력이 없는 카드는 비교에서 빠지고 보고하지 않는다)이고 경로 일치는 정규화한 경로 전체의 일치다 — 파일 이름이나 문자열 접두사가 아니다(`acceptance.md` AC-TCI-003 (g)(h)). 본문 경로 추출은 읽기 전용이고 저장하지 않으며(D13) 허브 체인(M4)은 이것을 읽지 않는다. git 탐침은 잠금 밖에서, 시간 상한 안에서만 돌고 상한을 넘으면 `unmeasured (time bound)` 로 출력한다. 시퀀스는 (스냅숏 읽기 → 제시 계산 → `Mutate` 로 admit → 제시 출력 → stdout 한 줄)이다.
- **검증 명령(패키지 한정)**: `SCRUB go test <패키지> -count=1 -v` 에 `-run` 플래그로 AC-TCI-003 의 앵커된 패턴을 주는 명령과 위 이름 각각의 -list 짝(가지마다 `^이름$` 로 앵커한 정확한 명령과 통과 출력의 모양은 `acceptance.md` AC-TCI-002~006). 회귀: `^TestClassifyCardText$`, `^TestNormalizeCardText$`, `^TestTokenSetJaccard$`(kanban), `^TestTodoAddRefusesExactDuplicate$`, `^TestTodoAddNearDuplicateRecordsOnly$`, `^TestTodoAnalysisNeverReordersQueue$`, `^TestTodoAdd_PrintsIDAndPosition$`, `^TestSD_AC014_MCPMatchesCLIWithProjectRoot$`(cli). 정적: `go vet ./internal/kanban/... ./internal/cli/...`, CI 가 쓰는 버전의 `golangci-lint run ./internal/kanban/... ./internal/cli/...`.
- **종료 증거**: AC-TCI-002~007. **되돌리기**: 커밋 되돌리기(스키마 변경 없음).

### F.4 M2 — 카드 스키마 (Priority High, 되돌리기 비용이 가장 크다)

- **목적**: 발행 속성과 finding 처분을 가산적으로 추가한다(REQ-TCI-007~011), `add` 의 발행 플래그와 그 검증(REQ-TCI-013 의 `add` 절).
- **파일**: `internal/kanban/backlog_store.go`(타입: `BacklogItem`·`BacklogFinding`·`BacklogArchivedFinding`), `backlog_sqlite.go`(`ensureColumn` 가산 경로를 `ensureSchema` 의 **마지막** retrofit 로 — 카드 표 둘에 `issuance`, finding 표 둘에 `disposition`; `backlogItemsTableColumns` 에는 넣지 않는다), `backlog_migrate.go`(INSERT 네 곳 — 카드 표 `:586`·`:469` 와 finding 표 `:595`·`:477`, `readSnapshot` 의 live finding 읽기 `:200`·`readArchive` 의 보관 finding 읽기 `:379` 가 `columnExpr` 로 새 컬럼을 읽도록, `assertBacklogParity`), `backlog_schema_freeze_test.go`(카드 표 둘의 동결 튜플 문자열에 같은 커밋에서 새 튜플을 추가하고, **finding 표 둘은 오늘 열 튜플이 동결돼 있지 않으므로 새 고정을 더한다**), `todo_queue_merge.go`(`rewriteArchivedFindings`)·`todo_merge_procedure.go`(항목·소견 복사 경로 — 카드와 finding 이 새 컬럼을 나르는지), `internal/cli/todo.go`(`add` 의 명시 플래그와 검증), `todo_drop.go`(사유 속성 저장, 접두사 유지), `todo_claim.go`(`todoJSONProjection` 또는 `omitempty` 로 골든 무변), `todo_export.go`(새 필드 처리 확인).
- **RED 테스트**: kanban — `TestBacklogIssuanceColumnRetrofit`, `TestBacklogIssuancePureReaderNoDDL`, `TestBacklogIssuanceArchiveRestoreRoundTrip`, `TestBacklogParityCoversIssuance`, `TestBacklogDispositionColumnRetrofit`, `TestBacklogDispositionPureReaderNoDDL`, `TestBacklogDispositionArchiveRestoreRoundTrip`, `TestBacklogParityCoversDisposition`, `TestQueueMergeCarriesIssuanceAndDisposition`, `TestBacklogIssuanceStoredAsNullWhenAbsent`, `TestBacklogDispositionStoredAsNullWhenAbsent`(원시 열 값이 SQL NULL — `{}`·빈 문자열이 아니다), 동결 테스트(기존 `TestTodoHistoryAddsNoSchemaChange`·`TestSchemaFreezeRecordsTransitionStamps` 를 새 튜플 기대로 고친다), `TestCardClosedAtAccessor`, `TestFindingDispositionRecordOnly`, `TestFindingDispositionAbsentForLegacy`. cli — `TestTodoDropStoresReason`, `TestTodoDropKeepsTextPrefix`, `TestTodoRelateDispositionVerb`, `TestTodoAddIssuanceFlagRefusals`.
- **구현 요점(D4 작업 기본값)**: `items`·`archived_items` 에 nullable TEXT 컬럼 하나(`issuance`, JSON: 스폰한 카드, 출처, 크기 추정 줄 수, 예상 파일, drop 사유), `findings`·`archived_findings` 에 nullable TEXT 컬럼 하나(`disposition`). 없음은 SQL NULL·nil 이고 `omitempty` 로 직렬화한다. closed-at 은 저장하지 않고 접근자 하나가 `archived_at`·`dropped_at` 에서 읽는다(D9). 과거 카드는 소급하지 않는다. `add` 의 새 플래그는 폴스루 경로가 플래그를 받지 않으므로 `add` 서브커맨드에만 있고, 닫힌 집합 검증(`--origin`)·존재 검증(`--parent`; 보관 카드도 존재로 본다)·플래그 중복 검증은 쓰기 전에 한다(D13).
- **검증 명령**: `SCRUB go test <패키지> -count=1 -v` 에 `-run` 플래그로 AC-TCI-008 의 앵커된 패턴을 주는 명령과 -list 짝(`acceptance.md` AC-TCI-008, -010, -011, -013). 회귀: `^TestTodoListJSON_GoldenByteIdentity$`(cli), 기존 동결·retrofit 테스트(`internal/kanban`), 큐 병합·이주 테스트. 정적 검사 M1 과 같다.
- **종료 증거**: AC-TCI-008, -010, -011, -013 의 `add` 절과 회귀 가드 AC-TCI-009. **되돌리기**: 커밋 되돌리기. 컬럼은 DB 에 남아도 읽는 쪽이 모르면 무해하다(가산 컬럼 선례). 단, 되돌리기 전에 쓴 데이터는 읽히지 않는다.

### F.5 M3 — 관계 모델과 그래프 (Priority High)

- **목적**: 일곱 종류 어휘, 종류별 제약, 해석기, 추이 질의, card→file 간선(REQ-TCI-012~017).
- **파일**: `internal/kanban/backlog_relation.go`(신규: 어휘 정규화·매핑·제약·해석기), `backlog_store.go`(순환 가드 도우미 일반화 — `WaitsOnOf` 는 그대로 두고 호출한다), `internal/cli/todo_relate.go`(새 쓰기 종류 `duplicates`·`supersedes`·`relates-to` 수용, 처분 동사, 제약 거절), `internal/cli/todo_trace.go`(신규), `internal/cli/todo.go`(서브커맨드 등록과 `todoLaneReadOnlyVerbs` 에 `trace`), `internal/graph/card_file.go`(신규: 카드 귀속 병합 커밋의 변경 파일 간선), `graph.go`(층 호출과 정렬), `meta.go`(지문 출처 추가), `internal/cli/graph.go`(간선 종류 집계 루프).
- **RED 테스트**: kanban — `TestRelationOntologyMapsLegacyKinds`, `TestRelationConstraintRefusesCycle`, `TestRelationConstraintSymmetricNormalization`, `TestCardGTDResolverBothDirections`, `TestCardGTDResolverAddsNoGTDWritePath`. cli — `TestRelationLegacyReadersByteIdentical`, `TestTodoRelateConstraintRefusalByteIdentical`, `TestTodoRelateRefusesProjectionKinds`, `TestTodoWhyShowsGTDRelations`, `TestTodoTraceTransitiveDeterministic`, `TestTodoTraceBreaksCycles`, `TestTodoTraceLaneReadOnlyAllowed`, `TestTodoTraceWritesNothing`. graph — `TestGraphCardFileEdgesDeterministic`, `TestGraphCardFileEdgesCarryNoQueueState`, `TestGraphCheckNoticesCardFileSource`, `TestGraphCardFileEdgesSeeAbsorbedMerge`, `TestGraphCardFileEdgesIgnoreAbsorbMerge`.
- **구현 요점**: 저장된 소견 행은 다시 쓰지 않는다 — 읽기 시 매핑(design §5)과 새 쓰기 종류는 기존 `findings` 어휘 확장으로 기록한다. `relate` 가 쓸 수 있는 새 종류는 `duplicates`·`supersedes`·`relates-to` 이고(쓰기 가능 집합은 오늘의 여섯 이름 + 이 셋), `merged-into` 는 `todo merge`(M4)만 쓴다. `parent-of`·`follow-up-of`·`merged-into` 를 `relate` 에 주면 오늘처럼 `--relation must be one of …` 로 거절돼야 한다 — 일곱 종류 어휘를 쓰기 가능 집합에 통째로 넣는 확장 실수를 `TestTodoRelateRefusesProjectionKinds` 가 막는다. `parent-of`/`follow-up-of` 는 M2 속성에서 읽기 투영(D10)이라 관계 동사가 쓰지 않으므로 그 종류의 순환·둘째 부모 검사는 `add` 플래그 검증(M2)이 맡는다. 해석기는 `gtd_items.card_id` 로 양방향이고 `gtd_relations` 에는 쓰지 않는다. 간선 층은 `moai graph build` 안에서 카드 귀속 병합과 그 병합이 가져온 파일에서만 만들고(큐를 읽지 않는다) 출처 지문에 병합 목록을 넣는다. **순회는 HEAD 에서 어느 부모 경로로든 닿는 병합을 본다**(첫 부모 경로만 걸으면 develop 을 흡수한 브랜치에서 흡수된 카드 병합을 놓친다 — `design.md` §6·§13). 귀속은 새 정규식이 아니라 기존 단일 귀속 지점(`subjectAttribution`)을 재사용해 흡수 방향 병합이 귀속하지 않도록 하고, 흡수 병합의 두 번째 부모 쪽에서만 닿는 카드 병합과 흡수 방향 제목을 한 고정 저장소에 둔 RED 테스트가 순회 선택을 정한다.
- **검증 명령**: kanban·cli·graph 패키지 각각 위 이름과 -list 짝(`acceptance.md` AC-TCI-012~016). 회귀: 기존 `todo why`·`list`·`export`·픽업 필터·`TestAutoRank*` 계열(cli), `internal/graph` 의 기존 `TestBuild*`·`TestCheck*`, GTD 관련 기존 테스트(`internal/kanban` gtd 계열).
- **종료 증거**: AC-TCI-012~016 과 AC-TCI-013 의 `relate` 절. **되돌리기**: 커밋 되돌리기(저장 행 불변이므로 데이터 되돌리기 없음).

### F.6 M4 — 묶음·병합·허브 직렬 (Priority High, 임대 경로를 건드린다)

- **목적**: 묶음 경로, 운영자 호출 병합 동사, 허브 파일 체인(REQ-TCI-018~020).
- **파일**: `internal/homestate/factory.go`(`cards` 에 묶음 컬럼을 `ALTER TABLE ... ADD COLUMN` 목록 방식으로 추가 — D14 기본값), `card_record.go`·`card_picked.go`·`card_transition.go`, `internal/homestate/hub_files.go`(신규: 임베드된 목록 적재)와 `hub_files.txt`(신규: 임베드 데이터, M0 의 `hub-files.txt` 의 경로 열), `internal/cli/factory_card.go`(적재 동사와 선택 호의 묶음 예약·선행 미병합 건너뛰기), `internal/cli/todo_merge.go`(신규), `internal/cli/todo.go`(서브커맨드 등록).
- **RED 테스트**: cli — `TestTodoMergeRecordsAndDrops`, `TestTodoMergeRefusesPickedAndCycles`, `TestTodoMergeNeverInvokedByAnalysis`, `TestTodoMergeRefusedInLane`, `TestFactoryNextBundleSerialLane`, `TestFactoryBundleKeepsSerialSlot`, `TestFactoryAssignBundleOrderGuard`, `TestFactoryAssignBundleHubChain`, `TestFactoryKeepSetReadsNoFileOverlap`. homestate — `TestHubFileListFromBaseline`, `TestHubFileListMatchesMeasuringCommand`, `TestHubFileLoaderReadsNoProjectPath`, `TestHubFileLoaderIgnoresProjectTree`, `TestHubFileMeasurementSkipPolicy`. 첫 테스트는 선행 미병합 `after` 후보가 선택 호 맨 앞에 설 때의 현재 동작을 관측으로 기록한다(§B.3).
- **구현 요점**: 병합은 `<into>` 본문에 `<from>` 본문을 명시 절로 덧붙이고, `merged-into` 관계를 기록하고, `<from>` 을 drop 사유와 함께 닫는다. picked·이미 병합·닫힌 대상·순환은 거절한다. 레인과 분석 경로에서는 호출할 수 없다. 묶음은 팩토리 레코드의 묶음 식별·순번 컬럼과 적재 동사로 만들고, 선택 호는 (a) 묶음 선행이 미병합인 멤버를 건너뛰고 (b) 묶음 레인이 다음 멤버를 우선 받고 (c) 다른 레인에는 묶음 멤버를 주지 않는다. 직렬 슬롯 의미와 keep-set 은 그대로다. 허브 체인은 **임베드된 허브 목록**과 두 열린 카드의 `files` 속성(명시 입력 — 본문 경로 추출은 읽지 않는다) 교차로 생성한다 — 목록은 기준선의 추적되는 `hub-files.txt` 와 경로 집합이 같고(시험이 대조), 적재 코드는 SPEC 디렉터리나 reports 경로를 읽지 않는다(정적 검사 + 작업 디렉터리를 바꿔 보는 동작 검사). 측정 시험은 SHA 가 없을 때 `CI` 환경 변수가 있으면 건너뛰지 않고 실패한다(`.github/workflows/ci.yml:131` 의 test 작업은 `fetch-depth: 0`). 측정 시험의 본문은 보고 이음매를 받는 함수 하나이고 `TestHubFileMeasurementSkipPolicy` 가 그 함수를 세 입력으로 직접 구동해 건너뜀·실패·실행을 관측한다 — 순수 결정 함수만 읽지 않는다(D32).
- **검증 명령**: cli·homestate 패키지 각각 위 이름과 -list 짝(`acceptance.md` AC-TCI-017~019). 회귀: 기존 `TestFactoryNext*`·`TestSD_AC014_...`, 직렬 슬롯·keep-set 테스트.
- **종료 증거**: AC-TCI-017~019. **되돌리기**: 커밋 되돌리기. 컬럼은 가산이라 남아도 무해하다.

### F.7 M6 — `moai web` 관계 그래프 보기 (Priority Medium)

- **목적**: `/todo?view=graph`(REQ-TCI-023~024).
- **파일**: `internal/web/todo_queue_read.go`(`readTodoQueue` 옆에 노드·간선을 읽는 이음매 — 보관 카드·관계·`gtd_relations` 를 `LoadPure`/읽기 전용 소스로), `todo_view.go`(VM 확장), `screens.templ`(그래프 패널, 생성물 `screens_templ.go`), `assets.go`(임베드 목록에 새 자산 이름), `assets/`(새 CSS·SVG 필요 시), `assets/i18n.js`(키), 테스트.
- **RED 테스트**: `TestTodoGraphViewRendersRelations`, `TestTodoGraphViewReadOnly`(POST 405·큐 파일 바이트 불변·웹 패키지 비시험 소스의 락 토큰 어휘 검사), `TestTodoGraphViewBounded`, `TestTodoGraphAssetsEmbedded`, `TestTodoGraphViewDoesNotWaitOnQueueLock`(저장소 자신의 `BacklogStore.Mutate` 안에서 락을 쥔 채 `GET /todo?view=graph` 가 2초 안에 200 과 모든 고정 카드 노드를 받는다; 같은 락 아래 `GET /todo` 통제를 함께 읽는다), 회귀 `TestTodoPageUnchangedWithoutViewParam`(M6 시작 시 변경 전 렌더러에 대해 먼저 쓴다 — 도착 시 초록인 가드).
- **구현 요점**: 서버 렌더 Templ, GET 전용, 쓰기·락 없음(락 파일의 바이트·mtime 은 락 획득의 증거가 아니다 — advisory flock 은 흔적을 남기지 않으므로 락을 쥔 채 요청하는 시험이 증언한다), 외부 요청 없음, 노드 수 상한(값은 M0 측정에서 — 관계에 이름이 오른 카드 수와 열린 카드 수가 하한 근거), 결정적 레이아웃(같은 입력 같은 출력). 새 문자열은 모든 로케일에 둔다. 새 JS 는 만들지 않는다 — `TestTodoGraphAssetsEmbedded` 가 임베드 자산 목록의 `.js` 항목이 M6 시작 때와 같음을 읽고, 응답 본문에서 `http:`·`https:`·`//` 로 시작하는 `src`·`href`·`action`·`data-src`·`srcset` 값과 `url(`·`@import` 인자가 없음을 읽는다(AC-TCI-022 (d)).
- **검증 명령**: `SCRUB go test ./internal/web -count=1 -v` 에 `-run` 플래그로 AC-TCI-022 의 앵커된 패턴을 주는 명령(`acceptance.md` AC-TCI-022), i18n 거버넌스 테스트, 저장소의 templ 생성 대상(`make templ-generate`, Makefile 의 `templ-generate`)으로 생성물을 다시 만든 뒤 생성물이 커밋에 포함됐는지 확인.
- **종료 증거**: AC-TCI-022~023. **되돌리기**: 커밋 되돌리기(자산·키 포함).

### F.8 M5 — 규칙 개정 (Priority High, 게이트됨)

**게이트 판독.** M5 는 카드 브랜치가 develop 을 흡수한 뒤, 아래 다섯 판독을 **한 번씩 따로**(각각 단일 호출, 파이프 없음) 읽고 출력된 수를 읽는다 — 모두 종료 코드 0 을 내므로 종료 코드는 판정 근거가 아니다(수를 읽지 않고 종료 코드만 보는 것이 이 게이트의 함정이다).

이 판독들은 움직이는 ref `HEAD` 를 의도적으로 쓴다 — 질문이 "지금 이 작업 트리에서 t1453 의 착지에 닿는가"라서 답이 뒤집히는 것이 신호이고, 핀하면 질문이 사라진다(`verification-completeness.md` §4 의 판별 질문). 핀한 값은 `acceptance.md` 의 L21·C8·C9·C21·C22·C28~C36 이 SHA 로 기록한다.

1. 대상: (a) `T` = `git rev-list --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9]' HEAD` — HEAD 에서 **어느 부모 경로로든** 닿는 t1453 제목 커밋 수(`--first-parent` 를 쓰지 않는다; id 뒤 `[^0-9]` 는 `t14530` 류를 거른다). (b) `A` = `git rev-list --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9].*absorb' HEAD` — `--grep` 하나(`--all-match` 없음)로 카드 id 와 `absorb` 를 같은 패턴에 적는다. git 은 `--grep` 정규식을 메시지의 줄 단위로 맞추므로 `A` 는 **병합 제목 줄**에 `absorb` 가 든 흡수 방향 병합 수이고(미착지 브랜치가 이미 다섯 개를 갖고 있다), `absorb` 가 본문에만 있는 착지는 세지 않는다(이터레이션 3 의 메시지 전체 형태는 그런 착지 t1439 를 흡수 방향으로 세어 게이트를 닫힌 채 읽었다 — `acceptance.md` C33~C36). **착지 후보 = T − A**, 1 이상이면 후보. **`T ≥ 1` 인데 `T − A = 0` 이면 닫힘으로 읽기 전에 판독 2 의 목록을 한 번 읽어** 제목이 착지로 읽히는 줄이 있는지 본다(제목 줄에 `absorb` 가 든 착지도 빠진다 — 안전한 방향이지만 신호 없이 닫히지 않게 한다). 술어를 말로 하면 "t1453 의 착지 병합(병합 제목 줄이 develop 을 그 브랜치에 흡수한다고 하지 않는 병합)이 HEAD 의 조상이다"이다.
2. 고정: 후보가 있으면 `git log --format='%h %s' -E -i --grep='^merge[( :]+(card )?t1453[^0-9]' HEAD` 의 목록에서 제목에 `absorb` 가 없는 줄의 SHA `S` 를 `progress.md` §E.2 에 적고, `git rev-list --no-walk --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9].*absorb' <S>` 가 `0`(판독 1(b) 와 같은 같은 줄 형태 — 메시지 전체 `--grep=absorb` 는 본문에 그 낱말이 든 착지에서 1 이라 쓰지 않는다: C36), `git merge-base --is-ancestor <S> HEAD` 가 종료 0 인지 읽는다. 이후 재독은 이 SHA 술어를 쓴다(핀한 주소).
3. 양성 대조: 같은 형태 1(a) 를 t1448·t1344 에 걸어 각각 1 이상(핀에서 1·1), t1448 에 형태 1(b) 가 0(핀에서 0 — 착지 병합의 제목 줄은 `absorb` 를 담지 않는다), 본문에만 `absorb` 가 든 착지 t1439 에 1(a) 가 1·1(b) 가 0(핀에서 1·0 — 후보 1; `acceptance.md` C33·C34).
4. 두 번째 부모 대조(눈먼 선택자 방지): `git rev-list --count -E -i --grep='^merge[( :]+(card )?t[0-9]+' HEAD` 가 같은 명령에 `--first-parent` 를 붙인 값보다 엄격히 크다(핀에서 333 대 214). 같거나 작으면 선택자가 첫 부모 경로 밖의 커밋을 본다는 증거가 없으므로 게이트는 "미측정 = 미충족"이고 보고서에 그렇게 적는다. 카드 브랜치가 아직 develop 을 흡수하지 않았다면 흡수한 뒤 다시 읽는다.
5. 형태 점검: t1453 이 `merge(t1453)` 이 아닌 제목으로, 또는 빨리감기·스쿼시로 착지하면 판독 1 이 0 으로 남아 게이트는 닫힌 채 읽힌다(안전한 방향). 그때는 `moai gtd pr t1453` 가 낸 착지 SHA 를 `git merge-base --is-ancestor <그 SHA> HEAD` 로 직접 읽고 그 출력을 `progress.md` §E.2 에 적은 뒤 리더가 모드를 판정한다. 착지가 되돌려진(revert) 경우는 원래 병합이 조상으로 남아 열린 채로 읽힌다 — 이 이력에는 되돌림 사례가 없어 측정하지 못한 추정이고, 판독 2 의 SHA 를 리더가 읽는다.

핀에서의 값: 판독 1 은 T=0·A=0, 판독 3 은 1·1(그리고 형태 1(b) 가 0; t1439 는 1·0), 판독 4 는 333 > 214. 흡수 병합 `b05c3be90…` 에서는 T=5·A=5 로 착지 후보가 0 이다(C28·C29). 이터레이션 1 의 `--first-parent` 형태는 카드 브랜치가 develop 을 흡수한 뒤 t1453 이 착지해도 영영 0 을 읽었다 — 흡수된 develop 병합은 흡수 병합의 두 번째 부모 쪽에만 있다. 그 맹점은 고정 SHA 로 보였다: 두 부모 병합 `b05c3be90…` 에서 첫 부모로는 닿지 않는 `merge(t1407)` 커밋 `46be0b8c8…` 을 `--first-parent` 선택자는 0 으로, 모든 부모를 걷는 선택자는 1 로 읽는다(`acceptance.md` C17~C20). 게이트는 M5 시작 직전과 M5 첫 커밋 직전에 다시 읽는다(탐침은 썩는다).

**Mode A (게이트 열림, REQ-TCI-021).**

1. 시작 측정: 카드 브랜치 기준 `git merge-base develop HEAD` 를 읽는 시점에 다시 구하고(핀하지 않는다), 대상 파일 여섯과 템플릿 사본의 크기(문자·바이트)를 다시 잰다. 그 사이 t1450·t1452·t1359·t1453 이 바꾼 것이 있는지 `git diff --name-only <그 SHA> -- <여섯 파일>` 로 읽는다.
2. 배치: 새 규칙은 두 파일이다 — (i) `.claude/rules/moai/workflow/card-issuance.md`(경로 한정, 상시 로드 아님, 템플릿 미러 있음)가 **메커니즘만** 싣는다: 여섯 섹션(카드 크기, 후속 지적 규칙, 파생 깊이, 동시 진행 한도, 발행 체크리스트, 부채 대장)이 각 한도를 "프로젝트가 측정해 정한 값"이라 부르고 값은 쓰지 않는다. (ii) `.claude/rules/local/card-issuance-thresholds.md`(로컬 전용, 미러 없음, 추적됨)가 **수치만** 싣는다 — 각 값은 기준선 기록의 값과 같다(`TestCardIssuanceRuleValuesMatchBaseline`). 이 한 가지 읽기를 `spec.md` REQ-TCI-021, `design.md` §8·§12 와 같은 문장으로 쓴다 — 배포 사본에는 카드 id·`.moai/` 경로·측정 수치가 들어가지 않는다(`TestCardIssuanceTemplateIsMechanismOnly`). `paths:` 는 `**/.claude/skills/moai/workflows/gtd.md,**/.claude/agents/moai/manager-todo.md` 로 한정하고 `kanban-dispatch*` 글롭과 자기 매칭하는 이름을 피한다(이름 함정은 SPEC-INSTRUCTION-BUDGET-SCOPE-001 이 기록했다). 로컬 전용 규칙의 `paths:` 는 같은 두 패턴이어서 메커니즘과 값이 함께 적재된다.
3. 도달성(발행 세션): 경로 한정 규칙은 `gtd.md` 나 `manager-todo.md` 를 열지 않는 세션에는 닿지 않는다. 그래서 상시 로드 `kanban-dispatch.md` 는 `card-issuance.md` 를 이름으로 가리키는 한 문장을 기존 문장을 압축해 얻은 자리로 받고(순증가 0자·0바이트), 이 문장이 도달을 맡는다. `card-issuance.md` 의 `paths:` 는 `internal/template/workflow_rule_paths_pinned_test.go` 의 고정 목록에 등록해 경로가 바뀌면 시험이 붉어지게 한다(조용히 멈추는 점화를 막는다, AC-TCI-021 (h)). `gtd.md` 는 병합 독트린 두 문장(`:63` 행과 `:114` 문장)을 개정하고 40,000자 이하로 맞춘다 — 그 압축이 내용의 이동이면 도달성(스킬 본문이 경로 트리거 없이 적재되는 경우)을 확인한다(`rule-loading-budget.md`: 도달성 확인 없는 바이트 절감은 성능 결과가 아니다). `sync-auditor.md` 에는 후속 지적 규칙과 부채 대장 문장을, `manager-todo.md` 에는 후속 지적·파생 깊이 문장을 더한다.
4. 미러: 바이트 동일 쌍(`gtd.md`, `kanban-dispatch-detail.md`, `kanban-dispatch-mechanics.md`, `manager-todo.md`)은 같은 편집으로 동일하게 유지한다. **분기된 두 쌍**(`kanban-dispatch.md`: 로컬에만 있는 `moai worktree sweep` 한 문장, `sync-auditor.md`: 18줄 차이)은 분기를 흡수하지 않고 **보존**한다 — 편집은 두 사본에 같은 텍스트로 하고, 분기 hunk 수가 편집 전과 같음을 `diff` 로 보인다(AUTO-PICK 의 선례). 두 쌍을 `declaredForkedPairs` 에 등록할지는 이 SPEC 이 정하지 않는다(범위 밖; 후속 항목으로 보고).
5. 생성물: `sync-auditor.md`·`manager-todo.md` 를 고쳤으면 `make agents-emit` 으로 `.codex/agents/moai/*.toml`(저장소 루트와 템플릿 하위)을 다시 만들고, 템플릿 아티팩트를 고친 뒤 마지막에 `go run ./internal/template/scripts/gen-catalog-hashes.go --all` 로 `catalog.yaml` 해시를 다시 만들어 같은 커밋에 스테이징한다.
6. 가드(패키지별 한 번씩 — 시험 이름이 있는 패키지를 지정한다; `acceptance.md` G21·G22 가 이 귀속을 읽었다): `go test ./internal/template` 에 `-run` 으로 `^TestManifestHashFormat$`, `^TestCatalogHashCoversSkillSubfiles$`, `^TestDeclaredRuleMirrorForks$`, `^TestTemplateNoInternalContentLeak$`, `^TestWorkflowRulePathsPinned$`, `^TestJevAutoExceptionLinkage$`, `^TestJevAutoExceptionWording$`, `^TestCardIDBaselineHasNoStaleEntries$`, `^TestCardIDBaselineIsPerFileAndPerLiteral$` 를 가지마다 앵커해 준다. `go test ./internal/template/agentemit` 에 `^TestGoldenCommittedArtifactsMatchEmission$`. `go test ./internal/cli` 에 `^TestTodoSkillDocumentsJevSource$`(`todo_skill_doc_parity_test.go`), `^TestTodoSkillDocumentsClassification$` 와 `TestAutoRank*` 계열(`todo_auto_rank_test.go`·`todo_auto_doc_test.go`). `go test ./internal/spec` 에 `^TestACCounterFullCorpusMatchesBaseline$`. 각각 `-list` 로 선택된 수를 확인한다.
7. 커밋 본문: 상시 로드 파일의 측정 전후 바이트·문자와, 이 규칙이 필요 없는 세션이 치르는 비용(`rule-authoring.md` (c))을 적는다.

**Mode B (게이트 닫힘, REQ-TCI-022).** 여섯 규칙 파일과 템플릿 사본은 편집하지 않는다. 추적되는 `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/m5-draft/` 에 세 파일을 쓴다 — `anchors.md`(기존 다섯 파일에 더할 각 문장과 **삽입 위치 앵커**: 파일, 앵커 문장, 앞/뒤), `card-issuance.md`(배포 규칙 초안, 메커니즘만), `card-issuance-thresholds.md`(로컬 전용 규칙 초안, 수치만) — 그리고 `progress.md` §E.2 에 게이트 다섯 판독의 출력과 리더가 발행할 후속 카드 문안을 적는다. 이 경우 완료 보고는 "요구는 충족했으나 규칙은 아직 효력이 없다"를 명시한다. D1 이 SPEC 을 in-progress 로 붙들라고 정하면 sync 는 M5 적용 뒤에 한다.

### F.9 위험과 완화

| # | 위험 | 마일스톤 | 완화 |
|---|---|---|---|
| R1 | 어휘 이웃이 소음이거나 재현율이 낮아 알림이 무시된다 | M1 | 점수·척도를 항상 표기, 하한 아래는 `--dry-run` 전용(D8), 확실한 일치를 먼저 보임; 하한은 M0 재측정으로 확정 |
| R2 | 제시가 `add` 를 느리게 한다(git 탐침 0.25초/브랜치) | M1 | 락 밖·시간 상한·`unmeasured (time bound)`; 탐침 수 상한은 M0 가 잰 진행 중 레인 수에서 |
| R3 | 새 컬럼이 동결 튜플·JSON 골든·보관/복원·큐 병합을 깬다 | M2 | JSON 컬럼 하나, `omitempty`, 동결 테스트를 같은 커밋에서 갱신, 큐 병합 복사 경로 RED 테스트 |
| R4 | retrofit 순서 오류로 v1→v2 재구성이 새 컬럼을 떨어뜨린다 | M2 | `ensureSchema` 의 마지막 가산 패스, `backlogItemsTableColumns` 에 넣지 않음(카드 t1310 순서) |
| R5 | 저장소 통합이 보관·복원 의미를 흔든다 | M3 | 저장 행 불변, 읽기 해석기만(D2 기본값), 읽기 시 매핑 |
| R6 | git 증거 간선이 `graph build` 를 느리게 하거나 비결정적이 된다 | M3 | 병합 SHA 키 증분 캐시 후보, 정렬·시각 없음, 결정성 테스트, 비용은 M3 첫 측정 |
| R7 | 묶음이 임대 경로를 깨거나 직렬 슬롯 의미를 바꾼다 | M4 | 선택 호 변경을 묶음 예약·선행 미병합 건너뛰기로 한정, 비묶음 카드 골든 테스트, A3 개정 행 |
| R8 | 병합 동사가 "분석은 접지 않는다" 불변식을 우회로로 연다 | M4 | 레인·분석·`relate` 에서 호출 불가를 테스트로 강제, 운영자 전용 |
| R9 | M5 압축이 내용을 경로 한정 파일로 옮겨 도달 불가를 만든다 | M5 | 도달성 확인, 상시 로드 순증 0, 스텁이 도달을 맡고 경로 고정 시험이 끊김을 막음 |
| R10 | t1453 이 같은 파일을 고쳐 충돌한다 | M5 | 게이트가 선행, 시작 직전 재측정과 diff 확인, Mode B |
| R11 | 웹 그래프가 큰 큐에서 무겁다 | M6 | 노드 상한(M0 측정 근거), 결정적 레이아웃, 서버 렌더 |
| R12 | 템플릿 가드(카드 id·중립성·카탈로그 해시)가 늦게 터진다 | M5 | 편집마다 가드 실행, 카탈로그 재생성을 같은 커밋에 |
| R13 | 분기된 미러 쌍이 흡수돼 의도치 않은 로컬 문장이 배포된다 | M5 | 분기 보존, hunk 수 증명 |
| R14 | 기준선이 코드와 한 커밋에 섞이거나, 제품 커밋 뒤에 착지하거나, 무시되는 경로에 놓여 순서를 증언할 수 없다 | M0 | 기준선 단독 커밋, 추적되는 SPEC 경로, 순서 증인 1~5(증인 5 가 B 앞의 제품 커밋을 본다), 사전 점검 8 |
| R15 | 임베드 허브 목록이 기준선 사본과 어긋나거나, 사용자 프로젝트의 같은 경로에 이 저장소의 측정에서 나온 직렬화를 적용한다 | M4 | 두 사본 경로 집합 대조 시험과 측정 명령 재실행 시험(CI 에서 SHA 없음은 실패). 사용자 프로젝트와의 경로 겹침은 **검증하지 못한 수용 잔여**다(`internal/config/defaults.go` 같은 흔한 경로가 후보) — 영향은 두 카드를 한 줄로 세우는 순서 지정이고 거절이 아니며 입력이 명시 `files` 속성뿐이다(`spec.md` §G) |
| R16 | 게이트가 눈먼 선택자로 읽혀 열림/닫힘이 거짓이 된다 | M5 | 모든 부모 경로 술어, 고정 SHA 대조와 런타임 부등식 대조, "미측정 = 미충족" |

### F.10 @MX 태그 계획 (`mx_plan`)

`mx_plan`
- **ANCHOR**(fan_in ≥ 3 예상): (1) 관계 어휘 정규화·제약 검증 함수(`relate`·`merge`·`trace`·웹 읽기 이음매·그래프 층이 부른다), (2) 읽기 전용 이웃 조회 함수(`add`·`--dry-run`·MCP·engage 가 부른다), (3) `WaitsOnOf`(기존 앵커 유지, 새 호출자를 앵커의 `@MX:REASON` fan_in 수에 반영). 앵커마다 `@MX:REASON` 필수.
- **WARN**: (1) git 탐침을 동시에 돌리는 경우(고루틴이 있다면) — 시간 상한과 취소 맥락 필수, (2) 선택 호의 묶음 예약 분기(복잡도 증가 시), (3) 그래프 층의 git 증거 추출(순회 복잡도).
- **NOTE**: 표시 하한·이웃 상한·구성요소 깊이·허브 임계값·탐침 시간 상한 상수(각 상수에 근거 측정 id 를 적는다), 어휘 매핑 표, drop 접두사 규약(`[DROPPED — ` 형태), 임베드 허브 목록의 출처(기준선 `tree:` SHA 와 갱신 절차).
- **TODO**: 처분이 `list` 표시에 반영되는지(D11 이 정하면), 성능 측정이 끝나지 않은 탐침 병렬화.
- **DEBT**: `closed_at` 접근자의 "unknown" 폴백(`@MX:CEILING` 소급 불가, `@MX:UPGRADE` D9 가 컬럼을 고르면). 태그 문구 언어는 `language.yaml` 의 `code_comments`(en)를 따른다.

### F.11 작업 기본값 요약 (계획이 진행하는 값, 판정이 아님)

모든 수치는 `research.md` §3 의 측정에서 왔다. **M0 재측정이 이 값을 다시 도출하며, 기준과 규칙은 이 표가 아니라 추적되는 기준선 기록을 가리킨다. 값을 싣는 곳은 로컬 전용 규칙 하나뿐이고 배포 사본은 메커니즘만 쓴다.**

| 항목 | 작업 값 | 근거 측정 |
|---|---|---|
| 유사 카드 상한 | 3 | 카드 본문이 정한 값(고정) |
| 표시 하한(token-set Jaccard) | 0.30 — 추가의 8.4%(최근 4.9%)에 알림, 기록된 관계 재현율 6% | SB03, SB05 |
| 구성요소 깊이 | 경로 앞 두 마디 — 깊이 3 은 공유 쌍 0 | SB07 |
| 카드 크기 하한 트리거 | 예상 제품 줄 50 미만이면서 제품 파일 3개 이하(착지 카드의 29%)는 같은 주 파일의 열린 카드와 묶거나 이유를 적는다 | GB09, GB13 |
| 카드 크기 상한 | 제품 줄 1,604 초과 또는 제품 파일 21 초과(착지 카드의 상위 10%) | GB03, GB04 |
| 파생 깊이 상한 | 2 — depth ≥ 3 은 71장(6.0%) | QB06 |
| 동시 진행 한도 | 16 — 한 3시간 구간에 커밋한 서로 다른 카드의 P90 대리 지표(중앙 5, 최대 41) | QB10(스냅숏 시점 picked 16) |
| 허브 임계값 | 72시간 창 최대 겹침 순위의 상위(`catalog.yaml` 20, `defaults.go` 12, `todo.go` 10, `kanban-dispatch.md` 10) — 정확한 컷과 목록은 M0 가 `hub-files.txt` 로 낸다 | GB15 |
| git 탐침 시간 상한 | M1 이 진행 중 레인 수와 탐침 지연 분포를 먼저 재서 정한다 | SB08 |
| 웹 그래프 노드 상한 | 관계에 이름이 오른 카드 176장과 열린 카드 37장의 합집합 규모에서 M6 가 정한다 | QB04, QB02 |

결정 열다섯 개의 작업 기본값은 `spec.md` §B 표, 선택지와 근거는 `design.md` §11, 열린 질문은 `decision-index.md`.

### F.12 커밋 구조

1. M0 기준선(코드 변경 없음, `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/` 아래 파일만) — **자기 커밋**이고 가장 먼저. 병합 전에 순서 증인 1~5 를 읽어 `progress.md` §E.2 에 기록한다(증인 3 은 두 질의의 SHA 목록 — 개수가 아니라 목록 — 이 같고 비어 있지 않은지를 읽고, 증인 5 는 `B` 만의 함수라 병합 뒤에도 같은 값이다).
2. 마일스톤마다 RED 시험 커밋, GREEN 커밋(필요하면 리팩터링 커밋). 모든 커밋 메시지에 카드 id(`t1454`)를 넣는다(브랜치 이름이 카드를 식별하지 않는다).
3. M5 Mode A 는 규칙·템플릿 사본·로컬 전용 규칙·생성 Codex TOML·`catalog.yaml` 을 한 커밋으로(생성기 출력 상한). 압축이 카드 id 기준선 짝(§D 템플릿 가드)의 문장을 지웠다면 `internal/template/card_id_leak_test.go` 의 기준선 항목 삭제도 같은 커밋에 담는다. Mode B 는 `m5-draft/` 와 `progress.md` 기록을 한 커밋으로.
4. sync 단계: 대상 SPEC 개정 행(A1~A6 중 적용분)은 각 대상 SPEC 별 커밋으로(manager-spec 재위임, 종결 커밋 제목은 전체 SPEC id 하나).

## §G 안티패턴

- 어휘 이웃이 의미 중복 판정을 대체한다고 약속하는 것. 표시는 "참고용"과 "확실함"을 구분한다.
- 진행 중 레인 겹침 입력이 비었는데 `none` 을 찍는 것. 읽지 못한 입력은 항상 `unmeasured` 다(REQ-TAU-013 의 규율).
- 제시 계산을 `Mutate` 안에서 하거나 제시 실패를 admit 실패로 바꾸는 것.
- 게이트 명령의 종료 코드만 읽고 출력된 수를 읽지 않는 것(0 을 내도 "열림"이 아니다).
- 게이트 선택자에서 `absorb` 를 가리지 않는 것 — 미착지 브랜치의 `merge(t1453): absorb … into WT-…` 다섯 개가 착지 없이 게이트를 연다(`acceptance.md` C28·C29). 반대로 `absorb` 를 메시지 **전체**에서 찾는 것도 틀리다 — 본문에만 그 낱말이 든 진짜 착지(t1439)가 흡수 방향으로 세어져 게이트가 닫힌 채 읽힌다(C33~C36); 카드 id 와 `absorb` 를 같은 `--grep` 의 같은 줄에 적는다.
- 순서 증인에서 두 질의의 **개수**가 같다는 것으로 "모든 제품 커밋이 B 의 후손"을 증언하는 것 — 곁가지 제품 커밋을 카드 id 가 적힌 병합으로 들이면 개수는 같고 집합이 다르다. SHA 목록을 `--full-history` 로 비교한다(`acceptance.md` AC-TCI-001 증인 3, MU-114·115).
- 락 파일의 바이트·mtime 불변으로 "락을 잡지 않았다"를 증언하는 것 — advisory flock 은 락 파일에 흔적을 남기지 않는다. 락을 쥔 채 요청해 제때 200 이 오는지로 읽는다(AC-TCI-022 (b2)).
- 기준선 순서를 `B^..HEAD` 같은 뒤쪽 범위만으로 증언하는 것 — 앞선 제품 커밋은 그 범위 밖이다(`acceptance.md` C25·C26; 증인 5 가 앞쪽을 본다).
- 게이트나 증거 선택자를 `--first-parent` 로 읽는 것 — develop 을 흡수한 카드 브랜치에서는 흡수된 병합을 영영 못 센다. 통합 브랜치 자체를 읽는 경우(G10)만 first-parent 가 맞다.
- "추적되는 형태"를 말하면서 `.moai/reports/` 아래에 쓰는 것, 또는 `ls` 종료 0 을 추적의 증거로 읽는 것 — 디스크에 있다는 것과 추적된다는 것은 다르다(`git ls-files --error-unmatch` 로 묻는다).
- 출하 코드가 SPEC 디렉터리나 reports 경로의 파일을 런타임에 읽는 것.
- 갈라진 미러 쌍을 "고치는 김에" 바이트 동일로 합치는 것(AUTO-PICK 의 보존 선례).
- 규칙 파일에 수치를 SPEC 에서 복사해 붙이는 것, 또는 배포 사본에 측정 수치를 쓰는 것 — 값은 기준선 기록에서 읽고 로컬 전용 규칙에만 둔다.
- `-run` 에 다중 선택자를 `^(A|B)$` 로 쓰는 것(lint 가 권하는 형태가 선택을 비운다) — 가지마다 `^TestName$` 를 쓴다.

## §H 상호참조

`spec.md`(요구·개정 행), `acceptance.md`(기준·RED 원장·변이 탐침), `design.md`(구조·표·선택지·기준선 운반체 §12·게이트 술어 §13), `research.md`(측정·재독 인용·공백), `decision-index.md`, `progress.md`. 규칙: `verification-completeness.md`, `verification-claim-integrity.md`, `spec-frontmatter-schema.md`, `rule-authoring.md`, `rule-loading-budget.md`, `sprint-round-naming.md`, `kanban-dispatch.md`, `gitflow-lane-protocol.md`, `.moai/docs/audit-artifact-convention.md`.
