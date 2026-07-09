---
description: "Illustrative examples and full 4-locale localization table for the session-handoff doctrine"
paths: "**/session-handoff.md"
---

# Session Handoff — Examples and Full Localization Table

> This is a path-scoped reference file for `session-handoff.md`. It holds illustrative Example sections and the full 4-locale Localization Table extracted from the always-loaded doctrine file to reduce context weight. The core doctrine (6-block skeleton, cut-line markers, Field-by-Field Spec, Pre-emit self-check, Auto-Memory Integration, Post-Paste /goal Follow-up Block, Diet Constraints) remains in `session-handoff.md`.

## Localization Table (Full 4-Locale)

The cut-line marker text AND the 6-block skeleton verbs/headers translate per `conversation_language`. This table is the SSOT for the locale renderings (the canonical skeleton uses the `<entering verb>` / `<header>` placeholders; concrete locale renderings live here). Cross-verified for consistency with `.claude/output-styles/moai/moai.md §8` (the canonical render surface).

| Element | English | Korean | Japanese | Chinese |
|---------|---------|--------|----------|---------|
| Cut-line top text | `Copy from here` | `여기부터 복사` | `ここからコピー` | `从这里复制` |
| Cut-line bottom text | `Copy to here` | `여기까지 복사` | `ここまでコピー` | `到这里复制` |
| Block 1 entering verb | `entering` | `진입` | `開始` | `进入` |
| Block 3 Preconditions header | `Preconditions:` | `전제 검증:` | `前提条件:` | `前提条件:` |
| Block 5 Run header | `Run:` | `실행:` | `実行:` | `执行:` |
| Block 6 After-merge header (PR workflow) | `After merge:` | `머지 후:` | `マージ後:` | `合并后:` |
| Block 6 Follow-up header (trunk no-PR) | `Follow-up:` | `후속:` | `後続:` | `后续:` |
| Memory heading | `## Next Session Entry Point` | `## 다음 세션 시작점` | `## 次セッション開始点` | `## 下一会话起点` |
| Post-paste /goal instruction line | Send the `/goal` line below as its own standalone message AFTER Implementation Kickoff Approval — slash commands parse only at input start, and setting a goal starts a turn immediately. | 아래 `/goal` 라인을 구현 착수 승인 후 **별도 메시지로 단독 전송** — 슬래시 커맨드는 입력 시작에서만 인식되며, goal 설정 즉시 턴이 시작됨. | 下記の `/goal` 行を実装着手承認後に**単独メッセージとして送信** — スラッシュコマンドは入力の先頭でのみ認識され、goal 設定と同時にターンが開始される。 | 在实现启动批准后，将下方 `/goal` 行**作为独立消息单独发送** — 斜杠命令仅在输入开头被识别，设定 goal 会立即开始一个回合。 |

Read `conversation_language` from `.moai/config/sections/language.yaml` at render time; substitute the localized text between the `✂────` decorators (cut-line markers) while keeping `✂` and `─` characters verbatim, and substitute the locale rendering for each Block 1/3/5/6 placeholder when emitting the paste-ready message.

**Fallback rule for locales not in the table.** The table above lists concrete renderings for en / ko / ja / zh only. When `conversation_language` is an ISO-639 code whose language column is NOT in this table (e.g. `fr`, `de`, `es`, `pt`, `vi`), English is the canonical fallback skeleton and each label translates to that locale using the naturalization principle (idiomatic phrasing a native reader expects, never literal word-by-word transliteration). In other words: locales not in the table fall back to the English column for the structural skeleton, with the label text rendered in the configured ISO-639 language — ISO-639 not in the table ⇒ English-skeleton fallback, not English-output.

## Example (Illustrative; substitute project-specific values when adapting)

```
✂──── 여기부터 복사 ────✂

ultrathink. SPEC-MYPROJ-001 implementation 진입.
applied lessons: <lesson-id-1>, <lesson-id-2>.
source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>

전제 검증:
1) git log --oneline -1 → <commit-sha> 확인
2) ls .moai/specs/SPEC-MYPROJ-001/ → N files

실행: /moai run SPEC-MYPROJ-001

머지 후: SPEC-MYPROJ-002 → SPEC-MYPROJ-003

✂──── 여기까지 복사 ────✂

아래 /goal 라인을 구현 착수 승인 후 별도 메시지로 단독 전송 — 슬래시 커맨드는 입력 시작에서만 인식됨 (run-phase + machine-verifiable end-state일 때만 방출; 아니면 생략):

✂──── 여기부터 복사 ────✂

/goal the SPEC's test suite passes AND lint is clean, or stop after 20 turns

✂──── 여기까지 복사 ────✂
```

## Example with Block 0 (Illustrative)

```
✂──── 여기부터 복사 ────✂

[New Terminal — START IN WORKTREE]
$ cd ~/.moai/worktrees/<project>/SPEC-MYPROJ-001
$ moai cc        # 또는 moai glm | claude (3가지 launcher 중 선택; 본 예시는 moai cc)

ultrathink. SPEC-MYPROJ-001 Epic N 진입.
applied lessons: <lesson-id-1>, <lesson-id-2>.

전제 검증:
0) git rev-parse --show-toplevel → ~/.moai/worktrees/<project>/SPEC-MYPROJ-001 (★ critical)
1) gh pr view <PR-number> → MERGED

실행: /moai run SPEC-MYPROJ-001 --team

후속: Milestone M<N+1> (single-SPEC next step) 또는 Epic N+1 (multi-SPEC next grouping)

✂──── 여기까지 복사 ────✂
```
