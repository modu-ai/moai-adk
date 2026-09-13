# t586 — t583 흡수 게이트 증거 (2026-09-12)

리드가 연 흡수 게이트의 판독 근거. 병합 커밋 `10ea337fe`
(부모 `4d985fe48` = stage B tip, `03a48b0df` = 로컬 develop == origin/develop).

| 파일 | 내용 |
|---|---|
| `merge-commit.txt` | 병합 커밋 SHA + 양쪽 부모 SHA |
| `file-overlap.txt` | 카드 전체 변경 파일 ∩ develop 변경 파일 — **빈 파일(0줄)** |
| `post-build.txt` | 흡수 후 `go build ./...` (exit 0) |
| `post-vet.txt` | 흡수 후 `go vet ./internal/cli/...` (exit 0, 무출력) |
| `post-wizard.txt` | 흡수 후 `go test ./internal/cli/wizard/` (exit 0) |

## 판독

겹치는 파일이 없다. 이 판정식의 왼쪽 끝은 **흡수한 ref 와의 merge-base** 로 구했다
(`git merge-base develop HEAD` → `03a48b0df`), 리터럴 base SHA 가 아니다 —
`gitflow-lane-protocol.md` §8 [HARD]. 리터럴 핀(`cd7dc491c`, 직전 흡수점)으로 재면
stage B 기여분 56개만 보이고 Phase A 산출물(`internal/cli/ptycaptest/**`,
`profile_questions.go`, `profile_translations.go` 등)이 범위 밖으로 빠진다.

- 카드 전체 기여: 138 파일 (보고서 제외 33 파일) — 대조군 1 이상, 측정 가능
- develop 델타: `eb50af5a8..03a48b0df` 2322 파일
- 교집합: **0**

카드가 만진 wizard 파일은 `downgrade_confirm.go` / `downgrade_confirm_test.go` /
`help_keymap.go` / `profile_questions.go` / `profile_translations.go` 이고, t583 이
만진 것은 `questions.go` / `translations.go` / `types.go` / `wizard.go` 다. 같은
패키지지만 서로 다른 파일이라 텍스트 충돌이 날 자리가 없었고, 실제로 병합은 충돌
없이 커밋됐다.

t583 이 삭제한 `WizardResult` 필드를 stage B 코드가 참조했다면 빌드가 깨진다.
깨지지 않았으므로 **컴파일 층의 의미 충돌은 0건**이다.

## 아직 세우지 않은 것

빌드 초록은 "컴파일된다"만 세우고 "golden 이 맞는다"는 세우지 않는다.
`internal/cli/testdata/downgrade-confirm/*.golden` 4개를 포함한 테스트 층 판정은
`internal/cli` 슬롯에서 수행한다 — 선택자는 stage B 와 같은 범위를 쓴다:

```
go test ./internal/cli/ -count=1 -list 'Profile\|Wizard\|HuhTheme\|UpdateVersion\|TUI' -timeout 600s
go test ./internal/cli/ -run  'Profile\|Wizard\|HuhTheme\|UpdateVersion\|TUI' -count=1 -v -timeout 600s
```

대조 기준은 stage B 직후 값(`after-baseline-cli-*.txt`: 선택 132, RUN 224, PASS 224,
FAIL 0, SKIP 0)이다. 그 앞의 `baseline2-cli-*.txt`(선택 129, RUN 217)는 stage B 이전
지점이므로 대조군이 아니다.

선택자 이름 수와 top-level PASS 수를 반드시 대조한다 — `-run` 선택자는 존재하지 않는
이름을 조용히 버리므로, 수를 맞춰 보지 않으면 줄어든 실행을 초록으로 읽는다. t583 이
위저드 질문을 16→4 로 줄이며 테스트를 지웠으니 **132 에서 줄어드는 것은 정상일 수 있다**;
줄어든 만큼이 t583 의 삭제분과 일치하는지 귀속시켜야 하고, 일치하지 않으면 그것이 결함이다.

## 흡수 전 값 (같은 트리, 병합 직전 `4d985fe48`)

`go build ./...` exit 0 · `go test ./internal/cli/wizard/` ok 3.399s.
이 두 값은 병합 전에 이 세션에서 관측했고 파일로 남기지 못했다 — `4d985fe48` 을
체크아웃하면 누구나 재유도할 수 있다. 흡수 후 값과 짝을 이루어 "흡수가 아무것도 깨지
않았다"는 귀속을 만든다.

🗿 MoAI
