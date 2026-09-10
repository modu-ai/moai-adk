# SPEC-INIT-TUX-I18N-001 — 구현 계획

> 카드 t586 · Tier M · 개발 방식 TDD(`quality.yaml` `development_mode: tdd`)

## §A 맥락

v1 프로필 위저드를 v2 위저드에 흡수하고, 프로필 없음 확인창을 없애고, 위저드 레이아웃과 남은 영어 표면을 고친다. 요구는 `spec.md` §B, 판정 방식은 `acceptance.md` 에 있다.

**Tier 판단 — M.** 제품 코드 파일은 대략 12개다: `profile_setup.go`, `profile_setup_translations.go`, `init.go`, `update.go`, `update_version.go`, `huh_theme.go`(삭제), `launcher.go`(주석), `wizard/wizard.go`, 새 프로필 질문 파일, `wizard/translations.go`, 새 확인창 헬퍼 파일, `go.mod`/`go.sum`. 여기에 테스트 파일(기존 `profile_setup_*_test.go` 8개 안팎의 정비, `huh_theme_test.go` 삭제, pty 하네스와 골든 새 파일)이 더해지면 20개를 넘는다. Tier 표의 파일 수 기준은 테스트를 셈에 넣으면 L 경계에 닿지만, 제품 코드 규모와 설계 판단의 폭(결정 D1~D6 이 이미 내려짐)을 보면 M 이 맞다. plan-auditor 가 L 로 올려야 한다고 보면 `design.md`·`research.md` 를 보태는 쪽으로 대응한다.

## §B 알려진 사실과 코드에서 드러난 불일치

- **`LangSelectTitle`/`LangSelectDesc` 는 코드에서 참조된다.** `schema_bridge.go:34` 의 `f.conversation_lang` 브리지가 읽는다. 다만 그 브리지를 부르는 `uikit.SchemaKeyToTUIField`/`FieldDefTUILabel` 의 비테스트 호출자는 0건이라 실행 경로에서는 쓰이지 않는다. 비테스트 코드 어디에서도 읽히지 않는 키는 `LangGroupTitle` 뿐이다. REQ-ITI-011 의 "읽히지 않는 키" 판정은 참조 여부로 한다(실행 도달 여부가 아니다).
- **`profileSetupText` 는 흡수 뒤에도 남는다.** `printProfileSummary`(`profile_setup.go:537`)와 스키마 브리지 세 개가 읽는다. 결정 D2-c 의 "키 이관"은 **폼 문자열**(질문 제목·설명·그룹 제목·1단계 언어 문구)을 위저드 번역 테이블로 옮기는 것으로 한정한다. 요약·브리지용 필드는 구조체에 남긴다.
- **위저드 `HelpSelect`/`HelpInput` 은 문장형이다**(예: "방향키로 이동, Enter로 선택, Esc로 취소"). huh 도움말 줄은 키 바인딩별 짧은 라벨(`toggle`, `next`, `submit`, `up`, `down`, `filter`, `select`)로 조립되므로, 이 두 문자열을 그대로 바인딩 라벨로 쓸 수 없다. 바인딩별 라벨을 새로 두고 두 문장 키는 지우는 쪽을 기본안으로 한다(§H 미결 2).
- **import 방향 제약.** `wizard` 패키지는 `cli` 를 import 할 수 없는데, 스키마 옵션 라벨은 `cli` 의 `schemaOptionBridge` 에서 풀린다. 프로필 질문의 옵션 목록은 `cli` 쪽에서 만들어 위저드에 넘기거나, 라벨 해석을 위저드가 받아 쓸 수 있는 형태로 넘겨야 한다.
- **소스 스캔 테스트.** `profile_setup_removed_questions_test.go:61,87,116` 이 `profile_setup.go` 파일 본문을 읽어 제거된 질문의 재등장을 막는다. 폼 코드가 옮겨 가면 스캔 대상 파일도 따라 옮겨야 가드가 공허해지지 않는다.
- **huh v2 공개 API 확인(모듈 캐시 `charm.land/huh/v2@v2.0.3`).** `(*Confirm).WithButtonAlignment` `field_confirm.go:361`, 기본값 `lipgloss.Center` `:53`; `(*Select[T]).Height` `field_select.go:256`; `(*Form).WithKeyMap` `form.go:284`; `NewDefaultKeyMap` `keymap.go:107`; `FieldSeparator` `theme.go:27`, 기본값 `"\n\n"` `:104`. 확인창 안쪽 `"\n"` 두 번은 `field_confirm.go:261-264` 에 박혀 있어 공개 API 로 뺄 수 없다(D3 유지 근거).
- **폭 계산 도구.** `go.mod` 직접 의존성에 `github.com/mattn/go-runewidth v0.0.29` 와 `charm.land/lipgloss/v2 v2.0.6` 이 있다. D5 의 표시 폭 계산에 새 의존성은 필요 없다.

## §C 사전 점검 (run 단계 첫 커밋 전)

1. `BASELINE_SHA=$(git rev-parse HEAD)` 를 첫 run 커밋 전에 잡아 progress 기록에 남긴다(흡수 전 발판 마일스톤용). t583 흡수 뒤에는 흡수 커밋에서 한 번 더 잡는다.
2. `go test ./internal/cli/wizard/... ./internal/cli/ -run 'Profile|Wizard|HuhTheme|UpdateVersion' -count=1` 기준선. 선택된 테스트 수가 0 이 아닌지 먼저 센다(빈 선택 초록 금지).
3. v1 import 기준선: `grep -rln '"github.com/charmbracelet/huh"' --include=*.go internal cmd pkg` (plan 단계 측정: 5개 파일, 트리 `e7a7d4bb3`, 참고값).
4. 판정 문서의 pty 캡처가 쓴 조건(tmux 3.6a, `TERM=xterm-256color`, 80×30) 이 로컬에 재현되는지 확인. 안 되면 pty AC 는 Gap 으로 보고한다.
5. t583 병합 여부: `git merge-base --is-ancestor <t583 병합 SHA> HEAD` — 리드 보고로 받은 SHA 를 쓴다.

## §D 제약 (위반 금지)

- `spec.md` §C 제약 전부.
- t583 겹침 파일(`wizard/wizard.go`, `wizard/questions.go`, `wizard/translations.go`, `init.go`) 은 흡수 게이트(§F) 전까지 수정하지 않는다.
- 렌더 결함 수리를 소스 문자열 검사로 판정하지 않는다. `WithButtonAlignment` 가 코드에 있다는 grep 결과는 버튼이 왼쪽에 그려졌다는 증거가 아니다.
- 커밋 메시지마다 카드 id `t586` 을 넣는다. 레인은 push 하지 않는다.

## §E 자기 검증 항목 (run 단계 §E 가 인용)

- E1 `acceptance.md` AC 16개 PASS/FAIL 표 — 판정 방식별 명령과 원문 출력
- E2 `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`
- E3 `go test -cover ./internal/cli/wizard/...` 와 `./internal/cli/` 변경 경로 커버리지
- E4 RED 출력(수리 전 골든·단위 테스트 실패 원문)
- E5 `go vet`, `golangci-lint run` (변경 패키지)
- E6 pty 캡처 파일 경로(`.moai/reports/t586/` 로 반출한 사본)

## §F 마일스톤

순서 원칙: 흡수 게이트 앞에서는 t583 파일과 겹치지 않는 작업만 한다. 게이트 앞뒤 각 묶음 안에서는 바뀔 가능성이 큰 결정(데이터·타입·사용자 흐름)을 앞에, 기계적 정리를 뒤에 둔다.

### 흡수 게이트 앞 — t583 과 겹치지 않음

#### M1 — 프로필 질문 세트와 번역 테이블 정의 (새 파일만)

- 프로필 위저드가 쓸 질문 세트를 `wizard` 패키지의 **새 파일**로 정의한다(아직 배선하지 않음). 필드 목록은 REQ-ITI-005 그대로.
- 폼 문자열(질문 제목·설명·그룹 제목, 1단계 언어 문구)을 4개 로케일로 새 번역 파일에 둔다. `profileSetupText` 쪽 원본은 아직 지우지 않는다.
- 옵션 목록을 `cli` 에서 받아 쓰는 형태를 정한다(§B import 방향 제약).
- 단위 테스트: 질문 id 집합이 현재 v1 폼 필드 집합과 같은지, 4개 로케일 키가 동등한지.

#### M2 — 다운그레이드 확인창 v2 이관 (`update_version.go`)

- `update_version.go:327-334` 의 v1 확인창을 v2 확인창으로 옮긴다. v2 확인창 생성은 `wizard` 패키지의 **새 파일** 헬퍼로 둔다.
- 문구 4개 로케일과 언어 해석은 새 파일에 둔다(`translations.go` 는 게이트 뒤에 합친다).
- REQ-ITI-009 골든(ko·en 폴백) + 기존 `update_version_test.go` 초록 유지.
- 이 시점에도 `huh_theme.go` 는 init/update/profile_setup 이 쓰므로 남는다.

#### M3 — 판정 발판 (pty 하네스 · 골든 헬퍼)

- 환경 변수로 게이트되는 pty 캡처 테스트 도구(테스트 바이너리 + `tmux capture-pane`, `t.Cleanup` 으로 세션 종료, 자식 프로세스 HOME 은 임시 디렉터리)를 새 테스트 파일로 만든다.
- ANSI 제거·표시 폭 기준 열 계산 헬퍼, View() 골든 비교 헬퍼를 새 테스트 파일로 만든다.
- 수리 전 RED 캡처를 떠서 `.moai/reports/t586/` 로 반출한다(순서 증거는 별도 커밋 — 기준선이 수리보다 먼저 커밋돼야 한다).

### 흡수 게이트 — t583 develop 병합 확인 → 이 워크트리에서 develop 흡수 → 흡수 트리에서 §C 재측정

#### M4 — 흡수 배선 (`profile_setup.go`, `wizard.go`)

- `runProfileSetup` 이 v2 위저드를 실행하게 바꾼다. 세 진입 경로(REQ-ITI-004) 모두 이 함수를 지나므로 호출 지점은 그대로다.
- REQ-ITI-006 보존 동작 다섯 가지를 표 테스트로 먼저 RED 로 세운 뒤 옮긴다: 모델 id 정규화 뒤 사전 선택, acceptEdits 정규화와 확인 줄, `quality.yaml` 에서 읽는 개발 방식, 세그먼트 맵 보존, 저장값 사전 선택. `profile_setup.go:40-41` 의 `@MX:NOTE`/`@MX:REASON` 은 새 바인딩 지점으로 옮긴다.
- REQ-ITI-007 명령 계약(이름 인자, 세션 워크트리, 취소 시 0, 요약) 기존 테스트 초록 유지.
- 소스 스캔 가드(`profile_setup_removed_questions_test.go`)의 스캔 대상을 폼이 옮겨 간 파일로 바꾼다.

#### M5 — 프로필 없음 확인창 제거 (`init.go`, `update.go`) 와 v1 퇴역

- `init.go:647-663`, `update.go:174-192` 의 확인창을 지우고 곧바로 `runProfileSetup` 을 부른다(REQ-ITI-001). 취소·실패 시 계속 진행(REQ-ITI-002), 비대화형 조건 유지(REQ-ITI-003). 호출을 테스트 가능한 이음새로 빼서 단위 테스트한다.
- `huh_theme.go`, `huh_theme_test.go` 삭제. v1 import 0 확인 뒤 `go mod tidy`.
- `launcher.go:1104-1106` 주석이 가리키는 함수 설명이 새 구조에서도 참인지 다시 읽고, 거짓이 된 부분만 고친다.

#### M6 — 레이아웃과 도움말 줄 (`wizard.go`, `translations.go`)

- 확인창 버튼 왼쪽 정렬(REQ-ITI-012), 필드 구분자 축소와 선택 필드 높이(REQ-ITI-013), 옵션 설명 열 정렬(REQ-ITI-014, D5 — `wizard.go:290` 단순 연결 대체), 키맵 도움말 라벨 현지화(REQ-ITI-010).
- M2 에서 새 파일에 둔 다운그레이드 문구와 M1 번역을 `translations.go` 체계로 합친다.
- 각 항목은 골든 RED → 수리 → 골든 GREEN, 이어서 pty 캡처로 수리 판정.

#### M7 — 키 정리와 판정 마감 (기계적)

- `profileSetupText` 에서 폼 전용이던 필드와 `LangGroupTitle`, 위저드 `HelpSelect`/`HelpInput` 등 읽히지 않게 된 키를 지운다(REQ-ITI-011). 키 참조 스윕 테스트와 로케일 동등성 테스트를 둔다.
- 골든 확정, pty 판정 캡처 반출, §E 증거 정리.

## §G 연기 항목

- **D4 단계 표시 줄 소실.** t583 흡수 뒤 M3 하네스로 80×30 에서 다시 캡처해, 증상이 남는지와 조건만 progress 기록에 적는다. 원인 규명·수리는 이 SPEC 의 AC 가 아니다.

## §H 미결 사항

1. [NEEDS CLARIFICATION: 프로필이 없는 상태의 `moai init` 에서 D1 을 따르면 프로필 위저드의 대화 언어 질문 뒤에 init 위저드의 대화 언어 질문이 다시 나온다(`wizard/questions.go:63` `conversation_language`). 두 번 묻는 것을 받아들일지, 프로필 답을 init 위저드 기본값으로만 넘길지(현재도 `RunWithDefaults` 가 프로필 값을 기본값으로 채운다), 한쪽 질문을 건너뛸지. t583 의 4문항 구성과 맞물린다.]
2. [NEEDS CLARIFICATION: 도움말 줄 현지화에서 기존 문장형 `HelpSelect`/`HelpInput` 을 지우고 키 바인딩별 짧은 라벨을 새로 둘지, 문장형을 어딘가에 계속 노출할지.]
3. [NEEDS CLARIFICATION: 다운그레이드 확인창의 언어 해석 순서 — 프로젝트 `language.yaml`(`wizard.ReadLocaleFromProject`) 우선인지, 활성 프로필 `ConversationLang` 우선인지. `moai update --version` 은 프로젝트 밖에서도 실행될 수 있다.]
4. [NEEDS CLARIFICATION: 흡수된 프로필 위저드도 init 위저드처럼 단계 표시 줄을 보여 줄지. 현재 v1 프로필 위저드에는 단계 표시가 없다.]

## §I 위험

| 위험 | 영향 | 대응 |
|---|---|---|
| t583 병합이 늦어짐 | M4~M7 착수 지연 | M1~M3 을 먼저 끝내 두고, 게이트에서 흡수 트리 재측정 |
| t583 이 `buildSelectField`·`buildConfirmField` 구조를 바꿈 | M6 수리 지점 이동 | 게이트 뒤 수리 지점을 다시 찾고 줄 번호를 다시 인용 |
| 옵션 라벨 브리지를 넘기는 방식이 import 순환을 만듦 | 빌드 실패 | M1 에서 옵션 목록 전달 형태를 먼저 확정 |
| 선택 필드 높이를 옵션 수에 고정하면 긴 목록이 잘림 | 옵션 가려짐 | 높이는 보이는 옵션 수와 터미널 높이를 함께 고려, 골든에 긴 목록 사례 포함 |
| tmux pty 폭 처리와 실제 터미널 차이(한글·한자) | pty 판정이 실제와 어긋남 | 판정 문서 Residual-risk 승계, 표시 폭 계산을 골든으로도 검증 |
| 소스 스캔 가드가 파일 이동으로 공허해짐 | 제거된 질문 재등장 미검출 | M4 에서 스캔 대상 변경 + 뮤턴트(제거된 질문 문자열 재삽입)로 RED 확인 |
| 파일 수가 Tier L 경계 | plan-audit 에서 Tier 상향 요구 | §A 판단 근거 제시, 필요 시 산출물 보강 |
