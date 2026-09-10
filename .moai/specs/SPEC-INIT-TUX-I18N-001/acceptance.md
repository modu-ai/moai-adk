# SPEC-INIT-TUX-I18N-001 — 인수 기준

> 카드 t586. 이 문서는 검증 층이다. 요구(GEARS)는 `spec.md` §B 에 있고, 여기 AC 는 모두 Given-When-Then 형식의 이진 판정이다. 각 AC 머리의 `maps REQ-…` 는 요구 매핑 선언이다(spec lint 의 커버리지 수집 형식).

## §A 판정 방식 세 가지

| 방식 | 무엇으로 판정하나 | 용도 |
|---|---|---|
| **pty** | 환경 변수로 게이트된 테스트 바이너리를 tmux pty(80×30, `TERM=xterm-256color`)에서 실행하고 `tmux capture-pane -p` 로 뜬 화면 | **수리 판정.** 렌더 결함이 실제로 고쳐졌는지. CI 기본값에서는 건너뛴다 |
| **golden** | 폼을 `formDriver` 식 메시지 구동으로 그린 `View()` 문자열(ANSI 제거)을 저장된 골든과 비교 | **회귀 가드.** CI 에서 매번 돈다 |
| **mech** | 단위 테스트(이음새 주입) 또는 단일 grep/`go list` 명령 | 흐름·저장·import·키 참조 같은 비렌더 성질 |

[HARD] 렌더 성질(버튼 위치, 빈 줄, 빈 카드 줄, 열 정렬, 현지화된 문구가 화면에 나오는지)은 **그려진 문자열이나 캡처 화면**으로만 판정한다. 폼 구성 코드에 특정 호출(`WithButtonAlignment`, `Height`, `WithKeyMap` 등)이 있다는 소스 검사는 렌더 증거로 인정하지 않는다.

[HARD] 골든·단위 테스트 실행 결과를 읽기 전에 선택된 테스트 수가 0 이 아닌지 먼저 확인한다. `[no tests to run]` 이 섞인 초록은 판정이 아니다.

## §D AC 매트릭스

### 프로필 없음 진입 (REQ-ITI-001~003)

- **AC-ITI-001** (maps REQ-ITI-001; mech) **Given** 대화형 TTY 이음새가 참이고 활성 프로필이 설정되지 않은 상태, **When** `moai init` 과 `moai update` 의 프로필 진입 지점을 각각 실행하면, **Then** 프로필 위저드 실행 이음새가 정확히 1회 호출되고, 그 전에 확인형 필드가 하나도 만들어지지 않으며, 비테스트 Go 소스에서 `No profile found` 문자열 검색 결과가 0건이다.
- **AC-ITI-002** (maps REQ-ITI-001; pty) **Given** 임시 HOME(프로필 없음)을 가진 자식 프로세스, **When** pty 에서 프로필 진입 경로를 실행하고 첫 화면을 캡처하면, **Then** 첫 화면에 대화 언어 질문 제목이 보이고 `Yes`/`No`/`예`/`아니오` 버튼 줄이 없다.
- **AC-ITI-003** (maps REQ-ITI-002; mech) **Given** 프로필 위저드 이음새가 (a) 사용자 취소를 돌려주는 경우와 (b) 오류를 돌려주는 경우, **When** 각 경우로 init 과 update 의 진입 지점을 실행하면, **Then** (a) 는 프로필 파일을 쓰지 않고 이후 흐름으로 넘어가며 종료 코드에 영향이 없고, (b) 는 경고 한 줄을 남기고 이후 흐름으로 넘어간다.
- **AC-ITI-004** (maps REQ-ITI-003; mech) **Given** (a) stdin 비 TTY, (b) `moai init --non-interactive`, (c) `moai update --yes` 세 조건, **When** 프로필이 없는 상태로 진입 지점을 실행하면, **Then** 세 조건 모두에서 프로필 위저드 이음새 호출 횟수가 0 이다.

### v1 흡수 (REQ-ITI-004~008)

- **AC-ITI-005** (maps REQ-ITI-004; mech) **Given** 수리된 트리, **When** `grep -rln '"github.com/charmbracelet/huh"' --include=*.go internal cmd pkg` 를 실행하면, **Then** 출력이 0줄이다. **And** `moaiHuhTheme`·`moaiHuhStyles` 식별자의 비테스트 소스 검색 결과가 0건이다. **And** `moai profile setup`, `moai profile --setup`, init/update 진입 세 경로가 모두 같은 v2 위저드 실행 이음새에 도달함을 단위 테스트가 보인다.
- **AC-ITI-006** (maps REQ-ITI-005; mech) **Given** `t.TempDir()` 프로젝트와 임시 프로필 저장 경로, 그리고 모든 필드에 답을 주입한 위저드 결과, **When** 저장 단계를 실행하면, **Then** `preferences.yaml` 에 이름·네 언어·모델·모델 정책·추론 강도·권한 모드가 주입값대로 기록되고, `quality.yaml` 의 `development_mode` 가 주입값으로 바뀌며, 각 선택 필드의 옵션 값 집합이 `settings.FieldOptionDefs` 가 돌려주는 값 집합과 같다(`model_policy` 는 `template.ValidModelPolicies()` 와 빈 값).
- **AC-ITI-007** (maps REQ-ITI-006; mech) **Given** 다음 다섯 사례를 담은 표 테스트, **When** 흡수된 위저드의 바인딩과 저장을 실행하면, **Then** 다섯 사례 모두 기대대로다.
  1. 저장값이 퇴역한 모델 id 이면 정규화된 별칭이 사전 선택된다(값이 빈 값으로 떨어지지 않는다).
  2. `acceptEdits` 를 고르면 저장되는 권한 모드가 빈 문자열이고, 출력에 `acceptEditsConfirmationLine` 이 정확히 한 번 나온다.
  3. 개발 방식 선택의 초기값이 프로젝트 `quality.yaml` 에서 온다(`preferences.yaml` 에 다른 값이 있어도).
  4. 저장 전 `statusline_segments` 맵이 저장 뒤 `preferences.yaml` 에 바이트 단위로 같게 남는다.
  5. 나머지 각 선택 필드는 저장된 현재 값이 사전 선택된다.
- **AC-ITI-008** (maps REQ-ITI-007; mech) **Given** 기존 명령 계약 테스트(`profile_worktree_test.go`, `profile_setup_summary_test.go`, 취소 경로 테스트), **When** 흡수 뒤 `internal/cli` 의 해당 테스트를 실행하면, **Then** 선택된 테스트 수가 0 이 아니고 모두 PASS 이며, 이름 인자 생략 시 `default` 프로필에 저장되고, 사용자 취소 시 종료 상태가 0 이다.
- **AC-ITI-009** (maps REQ-ITI-008; golden) **Given** 흡수된 프로필 위저드 폼, **When** 첫 질문에서 `ko`·`ja`·`zh` 를 각각 고르고 다음 그룹들로 이동하며 `View()` 를 그리면, **Then** 각 로케일 골든과 일치하고, 그려진 화면에 해당 필드의 영어 제목·설명·버튼·도움말 문자열이 하나도 남아 있지 않다(옵션 값 토큰과 고유명사는 예외 목록으로 명시).

### 잔여 영어 표면 (REQ-ITI-009~011)

- **AC-ITI-010** (maps REQ-ITI-009; golden) **Given** 다운그레이드 확인창, **When** (a) 해석된 언어가 `ko` 인 경우와 (b) 해석 가능한 언어가 없는 경우로 `View()` 를 그리면, **Then** (a) 는 제목·설명·두 버튼·도움말이 모두 한국어로 그려지고 골든과 일치하며, (b) 는 영어로 그려지고 골든과 일치한다.
- **AC-ITI-011** (maps REQ-ITI-010; golden) **Given** 선택형·입력형·확인형 필드가 있는 위저드 폼, **When** `en`·`ko`·`ja`·`zh` 각각으로 `View()` 를 그리면, **Then** 도움말 줄의 동작 라벨이 각 로케일 문자열로 그려지고 12개(필드 3종 × 로케일 4개) 골든과 모두 일치한다. `moai update -c` 재구성 위저드 폼도 같은 방식으로 `ko` 골든 1개를 포함한다.
- **AC-ITI-012** (maps REQ-ITI-011; mech) **Given** 수리된 트리, **When** 키 참조 스윕 테스트와 로케일 동등성 테스트를 실행하면, **Then** 위저드 번역 테이블과 `profileSetupText` 의 모든 키가 비테스트 코드에서 1회 이상 참조되고, `ko`·`ja`·`zh` 의 키 집합이 `en` 과 같다. **And** 스윕 테스트는 참조되지 않는 키를 하나 심은 뮤턴트에서 실패함이 관측돼 있다.

### 레이아웃 (REQ-ITI-012~014)

- **AC-ITI-013** (maps REQ-ITI-012; pty + golden) **Given** 확인형 질문이 있는 위저드 그룹, **When** pty 에서 캡처하고 같은 폼의 `View()` 를 그리면, **Then** 두 결과 모두에서 버튼 줄의 첫 버튼 시작 열이 같은 카드의 설명 줄 시작 열과 같다(판정 문서의 수리 전 값: v2 들여쓰기 7칸). 골든이 이를 고정한다.
- **AC-ITI-014** (maps REQ-ITI-013; pty + golden) **Given** 첫 페이지 그룹과 필드가 둘 이상인 그룹, **When** pty 캡처와 `View()` 를 뜨면, **Then** 연속한 두 필드 사이의 빈 줄이 0 이고, 선택 필드 마지막 옵션 줄 바로 아래에 내용 없는 카드 줄(`┃` 뒤 공백뿐인 줄)이 0 이다(수리 전 값: 필드 사이 1줄, 첫 페이지 빈 카드 줄 4줄).
- **AC-ITI-015** (maps REQ-ITI-014; pty + golden) **Given** 라벨에 한글·한자·가나가 섞인 옵션과 영문 라벨 옵션을 함께 가진 선택 필드, **When** pty 캡처와 `View()` 를 뜨고 각 옵션 줄에서 설명 시작 위치를 **표시 폭**(동아시아 전각 문자 2칸)으로 계산하면, **Then** 모든 옵션 줄의 설명 시작 표시 열이 같다. **And** 바이트 수나 룬 수로 열을 맞춘 뮤턴트에서는 이 검사가 실패함이 관측돼 있다.

### 판정 하네스 (REQ-ITI-015)

- **AC-ITI-016** (maps REQ-ITI-015; mech) **Given** pty 캡처 테스트, **When** (a) 환경 변수 없이 실행하고 (b) 환경 변수를 켜고 실행하면, **Then** (a) 는 캡처 테스트가 SKIP 으로 보고되고 tmux 세션을 하나도 만들지 않으며, (b) 는 캡처 파일을 만들고 종료 뒤 하네스가 연 tmux 세션이 0개 남는다. **And** 두 경우 모두 실행 전후 실제 HOME 아래 파일 목록(하네스가 쓸 수 있는 경로 한정)이 같다.

## §D.1 판정 방식과 중요도

| AC | 방식 | 중요도 | 비고 |
|---|---|---|---|
| AC-ITI-001, 003, 004 | mech | MUST | D1 흐름과 비대화형 보존 |
| AC-ITI-002 | pty | MUST | D1 수리 판정 |
| AC-ITI-005, 006, 007 | mech | MUST | D2 흡수와 저장값 보존 |
| AC-ITI-008 | mech | MUST | 명령 계약 회귀 가드 |
| AC-ITI-009, 010, 011 | golden | MUST | 현지화 회귀 가드 |
| AC-ITI-012 | mech | SHOULD | 키 정리(런타임 동작 불변) |
| AC-ITI-013, 014, 015 | pty + golden | MUST | pty 로 수리 판정, 골든으로 회귀 가드 |
| AC-ITI-016 | mech | MUST | 하네스 안전성(HOME 무기록, 세션 정리) |

pty 로 판정하는 AC: 002, 013, 014, 015 (4개). 골든으로 판정·가드하는 AC: 009, 010, 011, 013, 014, 015 (6개, 그중 3개는 pty 와 겹침). 기계 검사 AC: 001, 003, 004, 005, 006, 007, 008, 012, 016 (9개).

## §D.2 추적성

REQ-ITI-001→AC-ITI-001/002 · REQ-ITI-002→AC-ITI-003 · REQ-ITI-003→AC-ITI-004 · REQ-ITI-004→AC-ITI-005 · REQ-ITI-005→AC-ITI-006 · REQ-ITI-006→AC-ITI-007 · REQ-ITI-007→AC-ITI-008 · REQ-ITI-008→AC-ITI-009 · REQ-ITI-009→AC-ITI-010 · REQ-ITI-010→AC-ITI-011 · REQ-ITI-011→AC-ITI-012 · REQ-ITI-012→AC-ITI-013 · REQ-ITI-013→AC-ITI-014 · REQ-ITI-014→AC-ITI-015 · REQ-ITI-015→AC-ITI-016. 모든 REQ 에 AC 가 1개 이상 있고, 매핑 없는 AC 는 없다.

## §D.3 RED 기준선

plan 단계에서 실제로 실행한 명령만 적는다. 나머지 AC 의 RED 는 run 단계 사전 점검에서 `BASELINE_SHA` 를 잡은 뒤 뜬다(`plan.md` §C). 렌더 AC 의 RED 는 판정 문서의 캡처(트리 `110320ed274d3a18319f9fbd16d8001afcc214b8`)를 참고값으로 삼되, t583 흡수 뒤 흡수 트리에서 다시 뜬 캡처를 공식 RED 로 쓴다.

| 원장 | 명령 | 출력(원문) | 종료 코드 | 트리 | 연결 AC |
|---|---|---|---|---|---|
| L1 | `grep -rln '"github.com/charmbracelet/huh"' --include=*.go internal cmd pkg` | `internal/cli/huh_theme.go` / `internal/cli/update.go` / `internal/cli/update_version.go` / `internal/cli/profile_setup.go` / `internal/cli/init.go` | 0 | `e7a7d4bb3` (참고값, run 단계에서 재측정) | AC-ITI-005 |

RED 이유: L1 은 v1 import 가 5개 파일에 남아 있어 빨갛다. 흡수(M4)와 퇴역(M5)이 이를 0줄로 바꾼다.

## §D.4 경계 사례

- 프로필 위저드 도중 대화 언어를 바꾼 뒤 이전 그룹으로 돌아가면 이전 그룹도 새 언어로 다시 그려진다. 버튼 라벨은 huh v2 에 동적 설정자가 없어 폼 구성 시점 언어로 고정된다(`wizard.go:491-498` 주석). 이 한계는 REQ-ITI-008 판정에서 "첫 질문 이후 새로 그려지는 그룹" 기준으로 본다.
- 모델 저장값이 알 수 없는 문자열이면 정규화가 빈 값으로 떨어뜨리는 현재 동작을 유지한다(`normalizeModelLegacy1M`).
- 프로젝트 밖(`.moai` 없음)에서 `moai profile setup` 을 실행하면 개발 방식 저장은 건너뛰고 요약에 "프로젝트 동기화 없음"이 나온다(현재 동작 유지).
- 옵션 라벨이 터미널 폭보다 길면 정렬 열 계산이 폭을 넘지 않도록 줄인다. 골든에 긴 라벨 사례를 하나 둔다.
- Windows 에서는 pty 캡처 테스트가 SKIP 이며, 그 SKIP 은 PASS 로 세지 않는다.

## §D.5 품질 게이트

- 변경 패키지 테스트: `go test ./internal/cli/wizard/... -count=1`, `go test ./internal/cli/ -run '<변경 영역 선택자>' -count=1` — 선택 수 0 금지.
- 크로스 빌드: `GOOS=windows GOARCH=amd64 go build ./...`.
- 정적 검사: `go vet ./internal/cli/...`, `golangci-lint run` 신규 지적 0.
- 전체 스위트 판정은 develop push 뒤 CI.

## §D.6 완료 정의

- MUST AC 15개가 모두 PASS 이고, 각 판정에 명령과 원문 출력이 붙어 있다.
- pty 판정 캡처(수리 전·후)가 `.moai/reports/t586/` 에 반출돼 있고, 수리 전 캡처 커밋이 수리 커밋보다 앞선다.
- 골든 파일이 커밋돼 있고 CI 에서 돈다.
- `spec.md` §D 의 D3 한계 목록과 D4 연기 기록이 progress 기록에 남아 있다.
- 템플릿(`internal/template/templates/**`) diff 가 0 이다.
