# t371 — verdict (진행 중)

카드: t371 · SPEC-SPECLINT-GITBLIND-001 · 워크트리 `.claude/worktrees/t371` · 브랜치 `WT-lint-shallow-clone`

| 단계 | 상태 | 근거 |
|---|---|---|
| 보존 커밋 | 완료 | `3b637a290` — SPEC 4종 + 증거 6종, 명시 pathspec |
| develop 흡수 | 완료 | `git merge origin/develop`(`9328a5242`) rc=0 충돌 0 → HEAD `35bc0715f` |
| 인용 재측정 | 완료 | `.moai/reports/t371/remeasure-35bc0715f.md` |
| 린트 기준값 재산출 | 완료 | `0 error / 1096 warning` — `.moai/reports/t371/lint-local-35bc0715f.txt` |
| SPEC 인용 갱신 (v0.4.0) | 진행 | manager-spec 위임 |
| v0.4.0 plan-audit (Tier M 문턱 0.80) | 미착수 | — |
| kickoff 승인 | 미착수 | — |

---

## 관측 보존 — cwd 오해로 생긴 경로 (삭제하지 않음)

리드 지시로 **삭제하지 않고 증거로 남긴다**. 경로 자체가 증거이므로 지우면 사라진다.

```
.moai/reports/t371/.moai/state/config-cache.json                                       8612 bytes, 0 newlines
.moai/reports/t371/.moai/state/context-usage/387e58f8-17c6-4ee2-916a-43c271573db7.json  327 bytes
```

둘 다 mtime `2026-08-31 01:55`. `.gitignore:216` 의 `**/.moai/state/` 에 걸려 미추적 목록에도 뜨지 않으므로, 이 항목이 유일한 기록이다.

**정정 하나** — 최초 보고에서 `config-cache.json` 을 "0 bytes" 라 적었으나 틀렸다. `wc -l` 이 낸 0 은 **개행 수**였고(단일 행 JSON, 끝 개행 없음), 실제 크기는 **8612 bytes** 다. 빈 파일이 아니라 온전한 설정 캐시 한 벌이다.

### 무엇이 썼는가

`context-usage` 파일 내용:

```json
{ "schema_version": 2,
  "session_id": "387e58f8-17c6-4ee2-916a-43c271573db7",
  "writer_pid": 33339,
  "captured_at": "2026-08-31T01:55:02.656137+09:00",
  "context_window_size": 1000000, "tokens_used": 220000, "raw_pct": 22,
  "stage": "none", "band": "large", "model": "Opus 5 (1M context)", "effort": "medium" }
```

- **다른 세션이다.** `387e58f8-…` 는 현재 세션(`b2e58561-…`)이 아니다. `/clear` 이전의 t371 세션이 남긴 것이다.
- **statusline 쓰기 경로다.** context-usage 스냅샷은 statusline 이 렌더마다 `<projectDir>/.moai/state/context-usage/<session-id>.json` 에 쓴다. 그것이 `.moai/reports/t371/` 아래에 생겼다는 것은, 그 시점 프로젝트 루트가 **`.moai/reports/t371` 로 해소됐다**는 뜻이다.

### 같은 방향을 가리키는 두 번째 신호

`config-cache.json` 이 담은 값이 이 프로젝트의 설정이 아니다:

```
"User":{"Name":""}   "Language":{"ConversationLanguage":"en","ConversationLanguageName":"English", …}
```

이 프로젝트는 `.moai/config/sections/user.yaml` 이 `GOOS 오라버니~`, `language.yaml` 이 `ko` 다. 캐시에 박힌 것은 **컴파일 기본값**이다 — 그 루트에 `.moai/config/` 가 없어 로더가 설정을 못 찾고 기본값으로 채운 뒤 그 자리에 캐시를 썼다.

두 신호가 같은 결론을 가리킨다: **어떤 명령이 cwd 를 프로젝트 루트로 삼았고, 그 cwd 가 `.moai/reports/t371` 이었다.** 워크트리 루트를 거슬러 올라가 찾는 대신 서 있는 자리를 루트로 읽었다.

### 아직 확인되지 않은 것

- 어느 명령이었는지 특정하지 못했다. 이 디렉터리의 증거 생성 스크립트 `walker-input.sh` 는 `.moai/reports/t371/statusgit-18-ids.txt` 를 **상대 경로로** 읽으므로 워크트리 루트에서 실행되도록 쓰였고 자체적으로 `cd` 하지 않는다 — 이 스크립트 자신은 용의자가 아니다.
- statusline 이 그 cwd 를 어디서 받았는지(훅 입력의 `cwd` 필드인지, `os.Getwd()` 폴백인지)는 미확인. `input.CWD` 빈 값 → `os.Getwd()` 폴백은 알려진 형태이나, **이 사건이 그 형태라고 단정할 근거는 아직 없다.** 가설이지 관측이 아니다.

후속 카드 재료로 남긴다.
