# SPEC-V3R3-CLI-TUI-001 Acceptance Criteria

> Acceptance criteria는 모든 EARS 요구사항(REQ-CLI-TUI-001~019)에 대해 관찰 가능하고 자동 검증 가능한 증거를 정의한다. Each AC includes a Given-When-Then scenario plus an automated check command.

---

## AC-CLI-TUI-001: `internal/tui/` 패키지 존재 + 디자인 토큰 1:1 매칭

**관련 REQ**: REQ-CLI-TUI-001, REQ-CLI-TUI-002

### Given-When-Then

- **Given** 본 SPEC의 M1 milestone이 머지된 상태이고
- **When** 개발자가 `find internal/tui -name '*.go' -type f` 명령을 실행하면
- **Then** `theme.go`, `box.go`, `pill.go`, `status.go`, `form.go`, `table.go`, `prompt.go`, `term.go`, `help.go` 9개 파일이 모두 존재해야 한다.
- **And When** `go test ./internal/tui/... -run "TestLightTokens|TestDarkTokens" -v` 실행 시
- **Then** 테스트는 GREEN이며, 28개 토큰 키(chrome/chromeBorder/titleBar/bg/panel/fg/body/dim/faint/rule/ruleSoft/accent/accentDeep/accentSoft/accentSofter/success/successSoft/warning/warningSoft/danger/dangerSoft/info/infoSoft/cursor/selection/promptArrow/promptPath/shadow)별 라이트/다크 두 값이 모두 디자인 소스 `source/project/tui.jsx:9-69`와 정확히 일치해야 한다.

### 검증 명령

```bash
cd "$(git rev-parse --show-toplevel)"
ls internal/tui/{theme,box,pill,status,form,table,prompt,term,help}.go
go test ./internal/tui/... -run "TestLightTokens|TestDarkTokens" -v
```

### Edge Cases

- 토큰 추가 (예: `linkColor`)는 디자인 소스 변경 + 본 SPEC update 동반 필수
- 라이트/다크 한쪽만 정의 금지 (양쪽 모두 mandatory)

---

## AC-CLI-TUI-002: `moai banner` 출력이 design v2와 일치 (golden snapshot)

**관련 REQ**: REQ-CLI-TUI-005, REQ-CLI-TUI-013

### Given-When-Then

- **Given** M3가 머지된 상태이고 `MOAI_THEME=dark`, `NO_COLOR` 미설정 환경에서
- **When** 사용자가 `moai banner`를 실행하면
- **Then** 출력에 ANSI escape sequence `\x1b[38;2;62;179;164m`(다크 accent `#3eb3a4`) 또는 `\x1b[38;2;20;74;70m`(라이트 accent `#144a46`)이 포함되어야 하고, `\x1b[38;2;196;90;60m`(terra cotta `#C45A3C`) 또는 `\x1b[38;2;218;119;86m`(`#DA7756`)는 0건이어야 한다.
- **And When** `moai banner > /tmp/banner-actual.txt` 후 `diff testdata/banner-dark.golden /tmp/banner-actual.txt`
- **Then** diff 0줄이어야 한다.

### 검증 명령

```bash
# 1) terra cotta 색이 제거되었는지 확인
grep -E '#(C45A3C|DA7756)' internal/cli/banner.go && echo "FAIL: terra cotta remains" || echo "PASS"

# 2) 라이트 + 다크 + no-color 3 케이스 golden test
go test ./internal/cli/... -run "TestBanner_(Light|Dark|NoColor)" -v
```

### Edge Cases

- 버전 문자열이 비어있을 때 (`pkg/version` build flag 미주입) "unknown" 표시 유지
- `moai` (서브커맨드 없음)도 `banner` 호출하므로 동일 출력 검증

---

## AC-CLI-TUI-003: `moai doctor` 19 항목 + CheckLine 컴포넌트 통일

**관련 REQ**: REQ-CLI-TUI-006

### Given-When-Then

- **Given** M4 step 4b가 머지된 상태이고
- **When** 사용자가 `moai doctor`를 실행하면
- **Then** 출력은 두 섹션 헤더(`tui.Section`)로 시작해야 하며 "시스템" 그룹과 "MoAI-ADK" 그룹이 모두 표시되어야 한다.
- **And** 검사 항목은 모두 `tui.CheckLine`의 row 형식(`<status icon> <label> <value> · <hint>`)을 따라야 한다.
- **And** 출력 마지막에 `tui.Box("요약", accent=true)` 안에 `tui.Pill` 3개(통과/주의/실패 카운트)가 표시되어야 한다.
- **And** 검사 항목 총 개수는 19개 이상이어야 한다.

### 검증 명령

```bash
# 1) 19개 이상 항목 검증
moai doctor 2>&1 | grep -cE '^\s*[✓!✗·●○]' | awk '{ exit ($1 < 19) }' && echo "PASS"

# 2) CheckLine 통일 검증 (golden snapshot)
go test ./internal/cli/... -run "TestDoctor_CheckLine" -v

# 3) 실제 CLI 출력 시각 검증
moai doctor --no-color  # NO_COLOR 모드에서도 항목 row 그대로 유지
```

### Edge Cases

- 일부 검사가 fail / warn 인 환경 (Glamour 캐시 미존재 등)에서도 row 포맷 유지
- `--json` 플래그 출력은 본 SPEC scope 외 (PRESERVE)

---

## AC-CLI-TUI-004: `moai init` huh wizard Stepper + 라디오 prefix + 이모지 금지

**관련 REQ**: REQ-CLI-TUI-007, REQ-CLI-TUI-014

### Given-When-Then

- **Given** M5가 머지된 상태이고 비어있는 디렉토리(`/tmp/moai-init-test`)에서
- **When** 사용자가 `moai init my-app`을 실행하고 1단계에서 종료(Ctrl+C)하면
- **Then** 화면에 `tui.Stepper(current=1, total=6)` 출력이 포함되어야 하고, 라디오 옵션은 `◆` (선택) 또는 `◇` (미선택) prefix 두 종류만 사용해야 한다.
- **And** 이모지(U+1F300~U+1FAFF) 출력 0건이어야 한다.
- **And** 위저드 단계 진행 시 stepper의 `current` 값이 1→2→3...→6 증가해야 한다.

### 검증 명령

```bash
# 1) 이모지 0건 검증 (script로 init 자동화)
expect_script_run_init_and_capture > /tmp/init-output.txt
python3 -c "
import re; data=open('/tmp/init-output.txt').read()
emoji = re.findall(r'[\U0001F300-\U0001FAFF]', data)
print(f'Emoji count: {len(emoji)}')
exit(1 if len(emoji) > 0 else 0)
"

# 2) ◆/◇ prefix 검증
grep -E '^[[:space:]]*[◆◇]' /tmp/init-output.txt | wc -l  # >= 5 (5 라디오 옵션)

# 3) Stepper 6단계 검증
grep -cE '\b[1-6]/6\b' /tmp/init-output.txt
```

### Edge Cases

- `--non-interactive` 플래그 사용 시 wizard 우회 — golden snapshot 미적용
- huh 0.8 API 한계로 prefix 커스터마이징 불가 시 wrapper로 직접 그리기 (plan.md §12 OQ2)

---

## AC-CLI-TUI-005: NO_COLOR 환경변수 시 ANSI 색 0건

**관련 REQ**: REQ-CLI-TUI-009

### Given-When-Then

- **Given** M7가 머지된 상태이고
- **When** 사용자가 `NO_COLOR=1 moai doctor 2>&1 | cat` 명령을 실행하면
- **Then** 출력에 ANSI escape sequence(`\x1b\[`)가 0건 포함되어야 한다.
- **And** 텍스트 내용 자체는 색상 없는 상태에서도 의미상 동일해야 한다 (라벨/값/힌트 모두 표시).

### 검증 명령

```bash
NO_COLOR=1 moai doctor 2>&1 | cat | python3 -c "
import sys, re
data = sys.stdin.buffer.read()
ansi = re.findall(rb'\x1b\[[0-9;]*[a-zA-Z]', data)
print(f'ANSI sequences: {len(ansi)}')
sys.exit(1 if len(ansi) > 0 else 0)
"

# 14개 명령어 모두 검증 (배치)
for cmd in "moai" "moai doctor" "moai status" "moai version" "moai cc --dry-run" \
           "moai update --dry-run" "moai loop start --dry-run" "moai help"; do
  NO_COLOR=1 $cmd 2>&1 | grep -cE $'\x1b\\[' && echo "FAIL: $cmd has ANSI" || echo "PASS: $cmd"
done
```

### Edge Cases

- `NO_COLOR=` (빈 문자열)도 NO_COLOR로 인식 (NO_COLOR 표준 spec 준수)
- `MOAI_THEME=dark NO_COLOR=1`: NO_COLOR가 우선 (REQ-CLI-TUI-009)

---

## AC-CLI-TUI-006: Windows cmd.exe legacy 환경 fallback

**관련 REQ**: REQ-CLI-TUI-011, REQ-CLI-TUI-018

### Given-When-Then

- **Given** Windows runner (`windows-latest`) 환경이고 `cmd.exe`(not pwsh)에서 실행되는 상황
- **When** 사용자가 `moai doctor` 실행 시
- **Then** ANSI escape 시퀀스가 cmd.exe legacy 환경에서 그대로 출력되어 깨진 문자(`←[38;2;...m`)로 보이지 않아야 한다.
- **And** colorprofile.Detect()가 `colorprofile.NoTTY` 또는 `colorprofile.Ascii`를 반환하면 자동으로 NO_COLOR fallback이 적용되어야 한다.

### 검증 명령

```yaml
# .github/workflows/test.yml (추가)
- name: Windows cmd.exe TUI fallback
  if: runner.os == 'Windows'
  shell: cmd
  run: |
    moai.exe doctor > out.txt 2>&1
    findstr /C:"←[" out.txt && (echo FAIL: ANSI leaked & exit /b 1) || (echo PASS)
```

```bash
# 로컬 (macOS/Linux) — Wine 사용한 cross-test 또는 unit test에서 colorprofile mock
go test ./internal/tui/... -run "TestProfile_Ascii_Fallback" -v
```

### Edge Cases

- pwsh / Windows Terminal: truecolor 정상 출력
- WSL (Linux): TERM 환경변수 정상 → 정상 색 출력

---

## AC-CLI-TUI-007: 한글 + 영문 혼용 정렬 깨짐 0건

**관련 REQ**: REQ-CLI-TUI-001 (R-01 mitigation)

### Given-When-Then

- **Given** M2가 머지된 상태이고 18개 한글+영문 mixed 케이스 golden snapshot이 commit된 상태에서
- **When** `go test ./internal/tui/... -run "TestRenderMixedKoEn" -v` 실행
- **Then** 18 케이스 모두 GREEN이어야 하고, 각 케이스의 box border는 좌우 대칭이어야 한다.

### 검증 케이스 (18 mixed cases)

```
케이스 1: "환경 감지" (전부 한글)
케이스 2: "Layer 1" (전부 영문)
케이스 3: "환경 감지 Layer 1" (한글 + 영문 + 한글 패턴)
케이스 4: "환경 감지 · 환경 변수 우선" (가운뎃점 포함)
케이스 5: "moai init my-app" (영문 + 코드)
케이스 6: "v3.2.4 · 2026-04-28" (코드 + 한글기호 + 숫자)
케이스 7: "통과 9 · 주의 1 · 실패 0" (한글 + 숫자 mixed)
케이스 8: "feat/SPEC-AUTH-007" (영문/슬래시/하이픈)
케이스 9: "8 ahead · 0 behind origin/main" (영한 mixed + 슬래시)
케이스 10: "/usr/local/bin/moai" (전부 영문 코드)
케이스 11: "TS · React · Next" (3 token 한가운뎃점)
케이스 12: "한 줄에 한자漢字 섞음" (한글 + 한자)
케이스 13: "Pretendard · JetBrains Mono" (영영 mixed font names)
케이스 14: "MoAI-ADK · 모두의AI" (조합)
케이스 15: "↑↓ 이동 · enter 선택" (특수 + 한 + 영)
케이스 16: "yuna@air ~/work/my-app" (호스트 prompt)
케이스 17: "✓ 통과 · ! 주의 · ✗ 실패" (기호 + 한글)
케이스 18: "긴 한글 텍스트가 박스에 들어갈 때 줄바꿈" (한 줄 70+ chars)
```

### 검증 명령

```bash
go test ./internal/tui/... -run "TestRenderMixedKoEn" -v -count=1
ls internal/tui/testdata/mixed-*.golden | wc -l  # >= 18 (light) + 18 (dark) = 36
```

### Edge Cases

- runewidth 라이브러리 버전이 변경되면 snapshot 재생성 (rare; CI에서 자동 감지)
- emoji가 포함된 케이스 (현재 0건이지만 사용자 입력으로 들어올 가능성) → AC-CLI-TUI-004로 위임

---

## AC-CLI-TUI-008: 추가 dependency 0건 (go.mod 변경 없음)

**관련 REQ**: REQ-CLI-TUI-008

### Given-When-Then

- **Given** 본 SPEC의 모든 milestone이 머지된 상태이고
- **When** `git diff origin/main..HEAD -- go.mod | grep '^[+-][^+-]'` 명령 실행
- **Then** `+ require` 또는 `+ <new package>` 라인이 0건이어야 한다.
- **And** `bubbles` indirect → direct 승격(주석 ` // indirect` 제거)은 허용된다.

### 검증 명령

```bash
cd "$(git rev-parse --show-toplevel)"

# 1) 신규 require 0건 검증
git diff origin/main..HEAD -- go.mod | \
  grep '^+' | grep -v '^+++' | \
  grep -E 'require|github\.com|gopkg\.in|golang\.org' | \
  grep -v 'bubbles' | \
  awk 'END { exit (NR > 0) }' && echo "PASS: no new deps"

# 2) bubbles 승격만 허용 검증
git diff origin/main..HEAD -- go.mod | grep -E '^\-.*bubbles.*indirect'  # OLD: indirect
git diff origin/main..HEAD -- go.mod | grep -E '^\+.*bubbles' | grep -v indirect  # NEW: direct (no indirect comment)

# 3) go.sum 추가 0건 (bubbles 승격 외)
git diff origin/main..HEAD -- go.sum | wc -l  # = 0 또는 minimal
```

### Edge Cases

- `go mod tidy` 자동 실행으로 빈 줄 정리는 허용 (substantive change 아님)
- transitive dependency 자동 추가는 0건 (모두 기존 indirect chain 활용)

---

## AC-CLI-TUI-009: 인간 언어 i18n 메시지 카탈로그 골격 (ko + en)

**관련 REQ**: REQ-CLI-TUI-019, REQ-CLI-TUI-003

> **Taxonomy 분리 (D3 정정)**: 본 AC는 **인간 언어 i18n 카탈로그**(ISO 639-1 코드 기준)를 다룬다. CLAUDE.local.md §15 "16-language neutrality"는 **프로그래밍 언어** 중립성(go/python/typescript/.../swift)이며 별개 개념. 본 SPEC scope는 `ko` + `en` 두 카탈로그를 채우고, 나머지 인간 언어는 별도 SPEC에서 확장한다.

### Given-When-Then

- **Given** M7가 머지된 상태이고
- **When** `ls internal/tui/messages/` 명령 실행
- **Then** 최소 `ko.yaml` + `en.yaml` 두 개가 존재해야 한다 (본 SPEC scope).
- **And** ko.yaml과 en.yaml은 **동일한 키 셋**을 가져야 한다 (M7에서 양쪽 모두 채움).
- **And** `tui.Translate(key, lang)` 함수가 존재하며, 부재 키는 영어로 fallback해야 한다 (REQ-CLI-TUI-019).

### 검증 명령

```bash
# 1) 본 SPEC 필수 카탈로그 존재 (ko + en)
test -f internal/tui/messages/ko.yaml && \
  test -f internal/tui/messages/en.yaml && \
  echo "PASS: ko + en 카탈로그 존재" || echo "FAIL"

# 2) ko + en 키 셋 동일성 (yq 필요)
diff <(yq eval 'keys' internal/tui/messages/ko.yaml | sort) \
     <(yq eval 'keys' internal/tui/messages/en.yaml | sort) && \
  echo "PASS: 키 셋 일치" || echo "FAIL: 키 불일치"

# 3) tui.Translate 함수 + fallback 단위 테스트
go test ./internal/tui/... -run "TestI18n_FallbackToEn" -v -count=1
```

### Edge Cases

- 누락된 언어 키 lookup은 영어로 fallback (REQ-CLI-TUI-019)
- `conversation_language` 설정값이 카탈로그 부재 코드일 시 영어 fallback
- 향후 zh/ja/기타 인간 언어 카탈로그는 별도 SPEC (예: `SPEC-V3R3-I18N-EXPAND-001`)에서 추가
- 프로그래밍 언어 중립성(go..swift 16종)은 본 AC 범위 외 — 별도 검증 (예: `SPEC-V3R2-WF-002` 위임 정책)

---

## AC-CLI-TUI-010: 기존 단위 테스트 GREEN 유지 (회귀 0건)

**관련 REQ**: 전체 (회귀 방지 horizontal AC)

### Given-When-Then

- **Given** 본 SPEC의 모든 milestone이 머지된 상태이고
- **When** `go test ./... -count=1 -race` 실행
- **Then** 모든 테스트가 GREEN이어야 하고, race condition 0건이어야 한다.
- **And** `make build` 성공 후 generated 산출물의 diff가 0이어야 한다 (`internal/template/embedded.go` 변경 없음).

### 검증 명령

```bash
cd "$(git rev-parse --show-toplevel)"

# 1) 전체 테스트 + race
go test ./... -count=1 -race -timeout 5m

# 2) golangci-lint clean
golangci-lint run ./...

# 3) make build 동작 검증 (template 변경 없음 확인)
make build
git diff --stat HEAD~1 internal/template/embedded.go  # 0 lines

# 4) 기존 cli command tests 회귀 검증
go test ./internal/cli/... -count=1 -run "TestInit|TestDoctor|TestStatus|TestVersion|TestUpdate"
```

### Edge Cases

- Windows CI flaky test (TestSupervisor_NonZeroExit ETXTBSY) 재발 시 retry로 처리 (CLAUDE.local.md §18.11 lesson #4 참조)
- snapshot 변경은 의도적인 경우 PR description에 명시 + reviewer 확인

---

## AC-CLI-TUI-011: 텍스트 박스 문자 신규 작성 금지 (lipgloss `Border()` only)

**관련 REQ**: REQ-CLI-TUI-004

### Given-When-Then

- **Given** 본 SPEC의 모든 milestone이 머지된 상태이고
- **When** 마이그레이션된 Go 소스 (production 코드, `_test.go` 제외) 전체에서 unicode 박스 그리기 문자(`╭ ─ ╮ │ └ ┘ ━ ┃ ┏ ┓ ┗ ┛ ┌ ┐ ├ ┤ ┬ ┴ ┼`)를 grep
- **Then** `internal/tui/` 외부에서 0건 매치되어야 한다 (`internal/cli/`, `internal/statusline/`, `pkg/`, `cmd/` 모두 클린).
- **And** 모든 박스/카드/패널 시각 표현은 lipgloss `Border()` API 호출 경유여야 한다.

### 검증 명령

```bash
cd "$(git rev-parse --show-toplevel)"

# 1) 텍스트 박스 문자 sweep (production 코드만, _test.go 제외)
grep -RnE '╭|─|╮|│|└|┘|━|┃|┏|┓|┗|┛|┌|┐|├|┤|┬|┴|┼' \
  internal/cli/ internal/statusline/ pkg/ cmd/ \
  --include='*.go' | grep -v '_test.go' | \
  awk 'END { exit (NR > 0) }' && echo "PASS: no hand-drawn boxes" || echo "FAIL"

# 2) lipgloss.Border() 호출 횟수 검증 (positive evidence)
grep -RnE 'lipgloss\.(RoundedBorder|NormalBorder|ThickBorder|DoubleBorder|HiddenBorder)\(\)' \
  internal/tui/ --include='*.go' | wc -l  # >= 5 (Box, ThickBox, etc.)
```

### Edge Cases

- `_test.go` 파일은 golden snapshot 비교 시 박스 문자 substring 검증 목적으로 허용 (substring literal은 검증 데이터)
- `internal/tui/` 내부에서는 lipgloss API가 결과적으로 박스 문자를 출력하므로 직접 sweep 대상 아님
- 한국어/일본어 본문 자체에 들어가는 횡선 글자(`─` ASCII em-dash와 다름) 사용 시 false-positive 가능성 → 정확한 unicode codepoint 매칭만 grep

---

## AC-CLI-TUI-012: 다크 자동감지 + MOAI_THEME override 우선순위

**관련 REQ**: REQ-CLI-TUI-010, REQ-CLI-TUI-012

### Given-When-Then

- **Given** M7가 머지된 상태이고 `internal/tui/theme_test.go`에 테마 환경 매트릭스 테이블(`TestThemeResolve`)이 존재
- **When** `go test -run TestThemeResolve ./internal/tui/...` 실행
- **Then** 8개 케이스가 모두 PASS이어야 한다 (라이트/다크 백그라운드 × `MOAI_THEME` 4값 (light/dark/auto/unset) × `NO_COLOR` 2값).
- **And** 우선순위는 다음과 같이 검증되어야 한다: `NO_COLOR` > `MOAI_THEME` (명시값 light/dark만, auto/unset은 다음 우선순위로 위임) > `lipgloss.HasDarkBackground()` > 기본값 (dark, REQ-CLI-TUI-010 + spec.md §3.3).

### 매트릭스 테이블 (8 케이스 최소)

| 케이스 | NO_COLOR | MOAI_THEME | HasDarkBackground | 기대 결과            |
|--------|----------|------------|-------------------|----------------------|
| 1      | unset    | unset       | false (light)      | LightTheme           |
| 2      | unset    | unset       | true  (dark)       | DarkTheme            |
| 3      | unset    | "light"     | true  (dark)       | LightTheme (override) |
| 4      | unset    | "dark"      | false (light)      | DarkTheme  (override) |
| 5      | unset    | "auto"      | false (light)      | LightTheme           |
| 6      | unset    | "auto"      | true  (dark)       | DarkTheme            |
| 7      | "1"      | "dark"      | true  (dark)       | MonochromeTheme (NO_COLOR 우선) |
| 8      | "1"      | unset       | false (light)      | MonochromeTheme       |

### 검증 명령

```bash
cd "$(git rev-parse --show-toplevel)"
go test -run TestThemeResolve ./internal/tui/... -v -count=1
```

### Edge Cases

- `MOAI_THEME=invalid` → 자동 감지로 fallback (default-dark)
- 빈 문자열 `MOAI_THEME=""` → unset과 동일 처리
- `NO_COLOR=0` (string "0")도 NO_COLOR 표준 spec 준수 → set으로 인식 (NO_COLOR is "any non-empty value")

---

## AC-CLI-TUI-013: 금지 hex 색상(마젠타·오렌지·핑크) sweep

**관련 REQ**: REQ-CLI-TUI-015

### Given-When-Then

- **Given** 본 SPEC의 모든 milestone이 머지된 상태이고
- **When** 마이그레이션된 Go 소스 + 템플릿 전체에서 금지된 hex 색상 코드(마젠타/오렌지/핑크 계열 10종)를 grep
- **Then** 0건 매치되어야 한다.

### 검증 명령

```bash
cd "$(git rev-parse --show-toplevel)"

# 마젠타·오렌지·핑크 계열 금지 hex 10종 sweep
grep -RniE '#(FF00FF|E91E63|FF9800|FF5722|FF1493|FF6347|EE82EE|DA70D6|C71585|9932CC)' \
  internal/ cmd/ pkg/ \
  --include='*.go' --include='*.tmpl' --include='*.yaml' | \
  awk 'END { exit (NR > 0) }' && echo "PASS: no forbidden gradient hex" || echo "FAIL"

# 추가: 그라디언트 키워드 자체도 검증 (linear-gradient with magenta/orange)
grep -RniE 'gradient.*(magenta|orange|pink)' \
  internal/ cmd/ pkg/ \
  --include='*.go' --include='*.tmpl' | \
  awk 'END { exit (NR > 0) }' && echo "PASS: no forbidden gradient names"
```

### Edge Cases

- 색상이 주석으로만 (코드에서 비활성) 등장해도 reject (의도 표현 금지)
- 사용자 SPEC 작성 가이드 문서(`.moai/docs/`)는 grep target 외 (예시 색 표기 허용)
- 디자인 소스 핸드오프 번들(`.moai/design/SPEC-V3R3-CLI-TUI-001/source/`)은 grep target 외 (참조 자산, 실행 코드 아님)

---

## AC-CLI-TUI-014: 순백 #FFFFFF / 순흑 #000000 배경 금지

**관련 REQ**: REQ-CLI-TUI-016

### Given-When-Then

- **Given** M1이 머지된 상태이고 `internal/tui/theme.go` 토큰 정의가 존재
- **When** `internal/tui/` 패키지 전체에서 hex `#FFFFFF` (대소문자 무관) 또는 `#000000` 출현 grep + token-usage 단위 테스트 실행
- **Then** 토큰 정의 파일에서 `#FFFFFF` / `#000000` 매치 0건이어야 하고, `TestNoSurfaceUsesPureColors` 테스트가 GREEN이어야 한다.

### 검증 명령

```bash
cd "$(git rev-parse --show-toplevel)"

# 1) hex 직접 sweep (대소문자 변형 포함)
grep -RnE '#(FFFFFF|FFFFff|ffffff|Ffffff|FfffFf|000000)' \
  internal/tui/ --include='*.go' | \
  awk 'END { exit (NR > 0) }' && echo "PASS: no pure white/black tokens" || echo "FAIL"

# 2) 토큰 usage 검증 — 모든 surface(bg, panel, chrome)가 정의된 토큰 경유
go test -run TestNoSurfaceUsesPureColors ./internal/tui/... -v -count=1

# 3) 라이트 ivory / 다크 ink positive evidence
grep -RnE '#fbfaf6|#0a110f' internal/tui/theme.go | wc -l  # >= 2 (라이트 bg + 다크 bg)
```

### Edge Cases

- ANSI 표준에서 `lipgloss.Color("0")` (8 컬러 검정 인덱스) vs hex `#000000`은 별개 — 본 AC는 hex 표기만 검증
- terminal default fg/bg(ANSI reset)는 hex 코드가 아니므로 grep 대상 아님
- 사용자 NO_COLOR 모드에서 출력되는 무색은 `lipgloss.NoColor{}`로 처리, 순백/순흑 hex 미사용

---

## AC-CLI-TUI-015: prefers-reduced-motion 정적 폴백

**관련 REQ**: REQ-CLI-TUI-017

### Given-When-Then

- **Given** M2가 머지된 상태이고 `MOAI_REDUCED_MOTION=1` 환경에서
- **When** `MOAI_REDUCED_MOTION=1 go test -run TestSpinnerStatic ./internal/tui/...` 실행
- **Then** Spinner는 정적 dot(`●`) 단일 출력만 생성해야 하고 (애니메이션 frame rotation 0건), Progress는 단일 채워진 바(`██████████`)를 width-equivalent으로 출력해야 하며 transition CSS-equivalent 0건이어야 한다.

### 검증 명령

```bash
cd "$(git rev-parse --show-toplevel)"

# 1) 환경변수 설정 후 Spinner/Progress 정적 출력 검증
MOAI_REDUCED_MOTION=1 go test -run "TestSpinnerStatic|TestProgressStatic" ./internal/tui/... -v -count=1

# 2) 기본 동작 (motion 허용) 시 frame rotation 검증 (negative control)
go test -run "TestSpinnerAnimated" ./internal/tui/... -v -count=1

# 3) golden snapshot 차이 확인
diff internal/tui/testdata/spinner-reduced.golden internal/tui/testdata/spinner-animated.golden
# 두 파일은 달라야 함 (reduced ≠ animated)
```

### Edge Cases

- `MOAI_REDUCED_MOTION=0` (string "0")은 unset과 동일 처리 (boolean 환경변수 표준)
- 빈 문자열 `MOAI_REDUCED_MOTION=""` → unset 처리
- 자동 시스템 감지(예: macOS `defaults read com.apple.universalaccess reduceMotion`)는 본 SPEC scope 외 (별도 SPEC) — 환경변수 명시 override만 검증

---

## AC-CLI-TUI-016: 글로벌 hex 색상 sweep (REQ-CLI-TUI-013)

**관련 REQ**: REQ-CLI-TUI-013

> **D4 정정**: AC-CLI-TUI-002는 banner.go scope만 검증. 본 AC는 14개 마이그레이션 surface 전체에서 hex 색상 코드가 `internal/tui/` 외부에 0건 존재함을 검증한다.

### Given-When-Then

- **Given** 본 SPEC의 모든 milestone (M1~M7)이 머지된 상태이고
- **When** production Go 코드 (`internal/cli/`, `internal/statusline/`, `pkg/`, `cmd/`) 전체에서 hex 색상 패턴(`#[0-9a-fA-F]{3,8}`)을 grep
- **Then** `_test.go`/`testdata/`/주석 내 substring을 제외한 hex 코드는 **0건**이어야 한다.
- **And** `internal/tui/theme.go` 내부의 hex 토큰 정의는 양수 evidence (≥28 unique tokens) 으로 존재해야 한다.

### 검증 명령

```bash
cd "$(git rev-parse --show-toplevel)"

# 1) 글로벌 hex sweep — production Go 코드 (테스트/스냅샷 제외)
grep -RnE '#[0-9a-fA-F]{6}' \
  internal/cli/ internal/statusline/ pkg/ cmd/ \
  --include='*.go' \
  | grep -v '_test.go' \
  | grep -vE '//.*#[0-9a-fA-F]{6}' \
  | awk 'END { exit (NR > 0) }' && echo "PASS: no hex outside tui/" || echo "FAIL"

# 2) internal/tui/theme.go positive evidence (>= 28 unique tokens)
grep -oE '#[0-9a-fA-F]{6}' internal/tui/theme.go | sort -u | wc -l \
  | awk '$1 >= 28 { print "PASS: theme tokens 28+"; exit 0 } { print "FAIL: insufficient tokens"; exit 1 }'

# 3) (선택) 3-digit hex 단축 표기(#FFF) 도 sweep — 표현 일관성 강제
grep -RnE '#[0-9a-fA-F]{3}\b' \
  internal/cli/ internal/statusline/ pkg/ cmd/ \
  --include='*.go' \
  | grep -v '_test.go' | wc -l  # = 0 권장 (informational)
```

### Edge Cases

- `_test.go` golden snapshot 비교 substring literal은 검증 데이터로 허용
- `testdata/*.golden` 자체에 ANSI escape 안 hex가 들어가는 것은 정상 (output capture)
- 주석 내 hex 표기 (예: `// terra cotta #C45A3C 제거`)는 grep filter `grep -vE '//.*#'`로 제외 가능 — 또는 제거 권장
- 디자인 소스 핸드오프 번들 (`.moai/design/SPEC-V3R3-CLI-TUI-001/source/`)은 grep target 외 (참조 자산)

---

## AC-CLI-TUI-017: 글로벌 emoji 출력 sweep (REQ-CLI-TUI-014)

**관련 REQ**: REQ-CLI-TUI-014

> **D5 정정**: AC-CLI-TUI-004는 `moai init` scope만 검증. 본 AC는 14개 마이그레이션 surface 전체 production 코드에서 emoji codepoint(U+1F300~U+1FAFF + 일반 변형)가 0건 출현함을 검증한다.

### Given-When-Then

- **Given** 본 SPEC의 모든 milestone (M1~M7)이 머지된 상태이고
- **When** production Go 코드 + 템플릿 + 메시지 카탈로그 (`internal/cli/`, `internal/statusline/`, `pkg/`, `cmd/`, `internal/tui/messages/`) 전체에서 emoji codepoint 매칭
- **Then** `_test.go`/`testdata/`/주석을 제외한 emoji 출현은 **0건**이어야 한다.
- **And** 시각 기호로는 `tui.StatusIcon()` 의 단색 unicode (`✓`, `✗`, `!`, `·`, `●`, `○` 등)만 허용된다 — 이들은 emoji codepoint 범위 외.

### 검증 명령

```bash
cd "$(git rev-parse --show-toplevel)"

# 1) Python codepoint sweep — production code 전역
python3 - <<'PY'
import os, re, sys

EMOJI_RANGES = [
    (0x1F300, 0x1FAFF),   # Misc symbols & pictographs + supplemental
    (0x2600, 0x26FF),     # Misc symbols (☀ ☂ etc.)
    (0x2700, 0x27BF),     # Dingbats (✅ ❌ ❎ etc.)
    (0x1F000, 0x1F2FF),   # Mahjong + Domino + Playing cards
    (0x1F600, 0x1F64F),   # Emoticons
]

scan_paths = ["internal/cli", "internal/statusline", "pkg", "cmd", "internal/tui/messages"]
allowed_unicode = set("✓✗!·●○◆◇┌┐└┘─│┃━┏┓┗┛├┤┬┴┼╭╮╯╰")

violations = []
for base in scan_paths:
    if not os.path.isdir(base):
        continue
    for root, _, files in os.walk(base):
        for f in files:
            if f.endswith("_test.go") or f.endswith(".golden"):
                continue
            if not (f.endswith(".go") or f.endswith(".yaml") or f.endswith(".yml") or f.endswith(".tmpl")):
                continue
            path = os.path.join(root, f)
            with open(path, encoding="utf-8", errors="ignore") as fh:
                for lineno, line in enumerate(fh, 1):
                    if line.lstrip().startswith("//") or line.lstrip().startswith("#"):
                        continue
                    for ch in line:
                        cp = ord(ch)
                        for lo, hi in EMOJI_RANGES:
                            if lo <= cp <= hi and ch not in allowed_unicode:
                                violations.append(f"{path}:{lineno} U+{cp:04X} '{ch}'")
                                break

if violations:
    print("FAIL: emoji codepoints found:")
    for v in violations[:20]:
        print(f"  {v}")
    print(f"  ({len(violations)} total)" if len(violations) > 20 else "")
    sys.exit(1)
else:
    print("PASS: zero emoji codepoints in production code")
PY

# 2) Go 단위 테스트 (선택, 공통 헬퍼 활용)
go test ./internal/tui/... -run "TestNoEmojiInProductionCatalog" -v -count=1
```

### Edge Cases

- `tui.StatusIcon()` 의 `✓`(U+2713), `✗`(U+2717), `!`(U+0021), `·`(U+00B7), `●`(U+25CF), `○`(U+25CB), `◆`(U+25C6), `◇`(U+25C7) 는 단색 시각 기호로 허용 (emoji 범위 외)
- 박스 그리기 문자(`┌─┐│└┘╭╮╯╰` 등)는 별도 AC-CLI-TUI-011 scope (이중 검증)
- `_test.go` golden snapshot 안 emoji 비교 substring은 검증 데이터로 허용
- `.moai/design/SPEC-V3R3-CLI-TUI-001/source/` 핸드오프 번들은 sweep target 외 (참조 자산)
- 템플릿 메시지 카탈로그 (`internal/tui/messages/*.yaml`)는 사용자 한글 카피 안 emoji 0건 강제

---

## Quality Gate (TRUST 5 Mapping)

본 SPEC의 acceptance 통과는 곧 TRUST 5 quality gate 통과를 의미한다.

| TRUST 5 차원   | 본 SPEC AC                                                     |
|----------------|----------------------------------------------------------------|
| **Tested**     | AC-CLI-TUI-001 (theme), 002 (banner), 003 (doctor), 004 (init), 005 (no-color), 007 (mixed), 010 (regression), 012 (theme matrix), 015 (reduced motion) |
| **Readable**   | AC-CLI-TUI-007 (한글 정렬) + AC-CLI-TUI-011 (no hand-drawn boxes) — 가독성 핵심 |
| **Unified**    | AC-CLI-TUI-001 (single source of truth) + AC-CLI-TUI-013 (no forbidden gradient hex) + AC-CLI-TUI-014 (no #FFFFFF/#000000) + REQ-CLI-TUI-013 (no hardcoded hex outside `tui/`) |
| **Secured**    | (본 SPEC은 시각 layer만이라 보안 영향 0; 단 NO_COLOR 환경변수 부주의 처리는 escape leak 가능성 — AC-CLI-TUI-005에서 차단) |
| **Trackable**  | AC-CLI-TUI-008 (no go.mod drift) + 모든 commit이 milestone ID와 매칭 |

---

## Definition of Done (전체 SPEC)

- [ ] AC-CLI-TUI-001 ~ 017 모두 검증 명령 GREEN (D4/D5 정정 후 17개)
- [ ] EARS REQ-CLI-TUI-001 ~ 019 모두 코드/테스트로 evidence 존재 (AC 매핑 완전 커버)
- [ ] 글로벌 hex sweep (AC-016) + 글로벌 emoji sweep (AC-017) PASS
- [ ] golden snapshot 100+ files commit
- [ ] `internal/tui/` 패키지 godoc 100% (모든 export 함수에 //)
- [ ] `internal/tui/CHANGELOG.md` (또는 commit log) 에 디자인 소스 출처(`source/project/tui.jsx`, `source/project/screens.jsx`) 명시
- [ ] PR description에 라이트/다크 양쪽 스크린샷 attach (실제 ANSI 캡처)
- [ ] M7 완료 시 `acceptance.md` 본 파일을 update — Self-Audit checklist 모두 ✓
- [ ] CLAUDE.local.md §17 (docs-site 4개국어) 영향 검토 — 본 SPEC은 CLI 변경이라 docs-site reference 별도 SPEC 또는 next /moai sync에서 처리

---

## Self-Audit (작성자 검증)

- [x] 모든 EARS REQ (001-019)에 대응하는 AC 또는 검증 명령 존재 또는 in TRUST 5 mapping
- [x] AC당 Given-When-Then + 자동화 검증 명령 동시 제공
- [x] 시간 예측 0건 (priority 기반)
- [x] Edge case 각 AC별 명시
- [x] Definition of Done 단일 source-of-truth
- [x] Self-Audit checklist 본 파일 끝부분 위치
