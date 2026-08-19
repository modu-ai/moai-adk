# t143 — session-handoff ↔ moai.md §8 센티널 상호 모순 해소

판정: **PASS**

## Claim

`.claude/output-styles/moai/moai.md`(렌더 표면)의 드리프트 센티널이 SSOT의
Localization Table을 4열(en/ko/ja/zh)로 전제하고 있었으나, 실제 SSOT는 인라인
en/ko 2열 + ja/zh는 동반 파일로 이전된 구조다. 센티널 문장을 실제 구조에 맞게
정정했다.

## Evidence

### 1. SSOT 실측 (worktree WT-t143, base 4100d8767)

`.claude/rules/moai/workflow/session-handoff.md:65-79` — Localization Table 헤더:

```
| Element | English | Korean |
```

본문 서술(:67): "This table carries the en / ko columns inline (the inline
locales); the full 4-locale table (en / ko / ja / zh) lives in
`session-handoff-examples.md` § Localization Table (Full 4-Locale)."

SSOT 자신의 센티널(:164)도 동일하게 서술: "this file carries the en / ko subset
inline with the ja / zh columns relocated to `session-handoff-examples.md`".

템플릿 미러도 동일:

```
$ awk '/^### Localization Table/{f=1} f&&/^\| Element/{print; exit}' \
    internal/template/templates/.claude/rules/moai/workflow/session-handoff.md
| Element | English | Korean |
```

→ 렌더 표면 센티널만 낡았음이 확인됨(모순의 방향 확정).

### 2. 수정 (양 미러 동일 1줄)

`moai.md:689` 패리티 조건절:

- before: `the SSOT Localization Table must carry the same locale column count
  (en / ko / ja / zh — 4 columns) as this block's translation tables,`
- after: `the SSOT Localization Table must carry its inline en / ko subset with
  the ja / zh columns relocated to session-handoff-examples.md § Localization
  Table (Full 4-Locale) — that split is intended, so parity is checked against it
  rather than against a 4-column count in the SSOT itself, while this block's own
  translation tables stay full 4-locale (en / ko / ja / zh),`

반대 방향(SSOT에 4열 복원)은 카드 금지 조항대로 수행하지 않음.

### 3. 검증 명령 + 관측

```
$ git diff --stat
 .claude/output-styles/moai/moai.md                             | 2 +-
 internal/template/templates/.claude/output-styles/moai/moai.md | 2 +-
 2 files changed, 2 insertions(+), 2 deletions(-)

$ make build            # catalog.yaml updated successfully (12652 bytes) → 내용 변화 없음(git status 미표시)

$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...
ok   github.com/modu-ai/moai-adk/internal/template  28.638s

$ go test ./internal/config/... -run 'Budget|Token'
ok   github.com/modu-ai/moai-adk/internal/config  1.306s
```

## Baseline-attribution

- 트리: worktree `.claude/worktrees/t143`, 브랜치 `WT-t143`, base `4100d8767`
  (origin/main). 위 관측은 전부 이 트리·이 실행에서 나온 것.
- 미러 패리티: 수정 전 두 파일의 해당 문장이 바이트 동일(grep 동시 확인),
  수정 후에도 동일 치환 1건씩.

## Gaps (미검증)

- 전체 스위트 미실행(카드 금지 조항). `internal/template`, `internal/config`
  두 패키지만 실행. 전 패키지 판정은 CI 몫.
- 기계 패리티 검사기(4열을 문자 그대로 세는 스크립트)는 이 리포에 존재하지
  않음 — 센티널은 사람/에이전트가 읽는 서술 규율이며, 실패하던 것은 "문자 그대로
  적용했을 때"의 논리적 검사였다. 자동 검사기 신설은 이 카드 범위 밖.
- `session-handoff-examples.md`의 Full 4-Locale 표 내용 자체는 검증 대상이
  아니어서 열람만 하고 정합성 감사는 하지 않음.

## Residual-risk

- 향후 ja/zh 열이 SSOT로 되돌아오면 이 센티널이 다시 낡는다. 되돌림은 t142
  다이어트 반전이라 현행 금지 상태.
- release/v3.1.1 기준으로는 같은 문장이 `:657`에 있다(base가 main이라 행번호만
  다름). 통합 시 문자열 일치 기반 병합이므로 충돌 위험은 낮으나, 리드가 통합할 때
  moai.md 동시 편집 카드(t131)와의 순서를 확인할 것.
